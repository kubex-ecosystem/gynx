package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/kubex-ecosystem/gnyx/internal/config"
	"github.com/kubex-ecosystem/gnyx/internal/features/providers/registry"
	runtimeMW "github.com/kubex-ecosystem/gnyx/internal/runtime/middlewares"
	"github.com/kubex-ecosystem/gnyx/internal/types"
	kbxTypes "github.com/kubex-ecosystem/kbx/types"
	gl "github.com/kubex-ecosystem/logz"
)

type runtimeAIController struct {
	cfg  *config.MainConfig
	reg  *registry.Registry
	prod *runtimeMW.ProductionMiddleware
}

type dependencyStatus struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type unifiedRequest struct {
	Lang        string             `json:"lang"`
	Purpose     string             `json:"purpose"`
	PurposeType string             `json:"purpose_type"`
	Ideas       []string           `json:"ideas"`
	Prompt      string             `json:"prompt"`
	MaxTokens   int                `json:"max_tokens"`
	Model       string             `json:"model"`
	Provider    string             `json:"provider"`
	APIKey      string             `json:"api_key"`
	Messages    []kbxTypes.Message `json:"messages"`
	Meta        map[string]any     `json:"meta"`
}

type providerAttempt struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Class    string `json:"class"`
	Message  string `json:"message"`
}

type upstreamErrorInfo struct {
	Class       string
	Summary     string
	Retryable   bool
	ModelScoped bool
}

func registerRuntimeAIRoutes(
	r *gin.RouterGroup,
	container types.IContainer,
	reg *registry.Registry,
	prod *runtimeMW.ProductionMiddleware,
) {
	cfg, _ := container.Config().(*config.MainConfig)
	ctl := &runtimeAIController{
		cfg:  cfg,
		reg:  reg,
		prod: prod,
	}

	r.GET("/health", ctl.health)
	r.GET("/healthz", ctl.health)
	r.GET("/providers", ctl.providers)
	r.GET("/config", ctl.runtimeConfig)
	r.GET("/test", ctl.testProvider)
	r.POST("/unified", ctl.unified)
	r.POST("/unified/stream", ctl.unifiedStream)
}

func (ctl *runtimeAIController) health(c *gin.Context) {
	deps := map[string]dependencyStatus{}
	for _, providerName := range ctl.reg.ListProviders() {
		status := ctl.providerDependencyStatus(providerName)
		deps[providerName] = status
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "gnyx-gateway",
		"timestamp": time.Now().Unix(),
		"version":   "v1.0.0",
		"dependencies": gin.H{
			"providers": deps,
			"gobe_proxy": gin.H{
				"status":  "not_configured",
				"message": "not configured in active gateway runtime",
			},
		},
	})
}

func (ctl *runtimeAIController) providers(c *gin.Context) {
	availableNames := ctl.reg.ListProviders()
	availableSet := make(map[string]struct{}, len(availableNames))
	data := make([]gin.H, 0, len(availableNames))
	for _, name := range availableNames {
		availableSet[name] = struct{}{}
		cfg := ctl.reg.GetProviderConfig(name)
		dependencyStatus := ctl.providerDependencyStatus(name)
		available := dependencyStatus.Status == "healthy" || dependencyStatus.Status == "degraded" || dependencyStatus.Status == "unknown"
		candidateModels := ctl.resolveCandidateModels(name, unifiedRequest{})
		defaultModel := ""
		if len(candidateModels) > 0 {
			defaultModel = candidateModels[0]
		} else {
			defaultModel = providerDefaultModel(cfg)
		}
		item := gin.H{
			"name":         name,
			"type":         providerType(name, cfg),
			"available":    available,
			"defaultModel": defaultModel,
			"health":       dependencyStatus,
		}
		if dependencyStatus.Error != "" {
			item["error"] = dependencyStatus.Error
		}
		data = append(data, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"object":    "list",
		"data":      data,
		"hasMore":   false,
		"timestamp": time.Now().Unix(),
		"providers": ctl.providerCatalog(availableSet),
	})
}

func (ctl *runtimeAIController) runtimeConfig(c *gin.Context) {
	availableNames := ctl.reg.ListProviders()
	availableSet := make(map[string]struct{}, len(availableNames))
	for _, name := range availableNames {
		availableSet[name] = struct{}{}
	}

	serverName := "GNyx"
	serverVersion := "v1.0.0"
	serverPort := "5000"
	serverStatus := "ready"
	if ctl.cfg != nil && ctl.cfg.ServerConfig != nil {
		if appName := strings.TrimSpace(ctl.cfg.ServerConfig.Basic.AppName); appName != "" {
			serverName = appName
		}
		if port := strings.TrimSpace(ctl.cfg.ServerConfig.Runtime.Port); port != "" {
			serverPort = port
		}
	}

	providers := ctl.providerCatalog(availableSet)
	availableProviders := make([]string, 0, len(availableNames))
	for _, name := range availableNames {
		if _, ok := providers[name]; ok {
			availableProviders = append(availableProviders, name)
		}
	}
	sort.Strings(availableProviders)

	defaultProvider := ""
	if len(availableProviders) > 0 {
		defaultProvider = availableProviders[0]
	}

	c.JSON(http.StatusOK, gin.H{
		"server": gin.H{
			"name":    serverName,
			"version": serverVersion,
			"port":    serverPort,
			"status":  serverStatus,
		},
		"providers":           providers,
		"available_providers": availableProviders,
		"default_provider":    defaultProvider,
		"environment": gin.H{
			"demo_mode": false,
		},
		"openai_available":   providerReady(providers, "openai"),
		"deepseek_available": providerReady(providers, "deepseek"),
		"ollama_available":   providerReady(providers, "ollama"),
		"claude_available":   providerReady(providers, "claude") || providerReady(providers, "anthropic"),
		"gemini_available":   providerReady(providers, "gemini"),
		"chatgpt_available":  providerReady(providers, "chatgpt") || providerReady(providers, "openai"),
	})
}

func (ctl *runtimeAIController) testProvider(c *gin.Context) {
	providerName := strings.TrimSpace(c.Query("provider"))
	if providerName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"available": false,
			"message":   "provider query parameter is required",
		})
		return
	}

	if ctl.reg.ResolveProvider(providerName) == nil {
		c.JSON(http.StatusOK, gin.H{
			"available": false,
			"message":   fmt.Sprintf("provider '%s' is not available in the active registry", providerName),
		})
		return
	}

	if err := ctl.reg.ResolveProvider(providerName).Available(); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"available": false,
			"message":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"available": true,
		"message":   fmt.Sprintf("provider '%s' is ready", providerName),
	})
}

func (ctl *runtimeAIController) unified(c *gin.Context) {
	req, providerName, modelCandidates, err := ctl.decodeUnifiedRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": err.Error()})
		return
	}

	content, modelName, usage, attempts, err := ctl.executeUnified(c.Request.Context(), providerName, modelCandidates, req)
	if err != nil {
		info := classifyUpstreamError(err)
		c.JSON(http.StatusBadGateway, gin.H{
			"error":       err.Error(),
			"message":     err.Error(),
			"error_class": info.Class,
			"provider":    providerName,
			"attempts":    attempts,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"response": strings.TrimSpace(content),
		"provider": providerName,
		"model":    modelName,
		"mode":     "server",
		"usage":    toUnifiedUsage(usage),
		"attempts": attempts,
	})
}

func (ctl *runtimeAIController) unifiedStream(c *gin.Context) {
	req, providerName, modelCandidates, err := ctl.decodeUnifiedRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": err.Error()})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming unsupported", "message": "streaming unsupported"})
		return
	}

	selectedModel, usage, attempts, err := ctl.streamUnified(c, providerName, modelCandidates, req)
	if err != nil {
		info := classifyUpstreamError(err)
		writeSSE(c.Writer, gin.H{
			"error":       err.Error(),
			"error_class": info.Class,
			"provider":    providerName,
			"attempts":    attempts,
			"done":        true,
		})
		writeSSE(c.Writer, "[DONE]")
		flusher.Flush()
		return
	}
	writeSSE(c.Writer, gin.H{
		"done":     true,
		"provider": providerName,
		"model":    selectedModel,
		"usage":    toUnifiedUsage(usage),
		"attempts": attempts,
	})
	writeSSE(c.Writer, "[DONE]")
	flusher.Flush()
}

func (ctl *runtimeAIController) decodeUnifiedRequest(c *gin.Context) (unifiedRequest, string, []string, error) {
	var req unifiedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return req, "", nil, err
	}

	providerName := strings.TrimSpace(req.Provider)
	if providerName == "" {
		available := ctl.reg.ListProviders()
		if len(available) == 0 {
			return req, "", nil, gl.Errorf("no providers are available in the active registry")
		}
		providerName = available[0]
	}

	provider := ctl.reg.ResolveProvider(providerName)
	if provider == nil {
		return req, "", nil, gl.Errorf("provider '%s' is not available in the active registry", providerName)
	}
	if err := provider.Available(); err != nil {
		return req, "", nil, gl.Errorf("provider '%s' is unavailable: %v", providerName, err)
	}

	modelCandidates := ctl.resolveCandidateModels(providerName, req)
	if len(modelCandidates) == 0 {
		return req, "", nil, gl.Errorf("provider '%s' has no model candidates configured", providerName)
	}

	return req, providerName, modelCandidates, nil
}

func (ctl *runtimeAIController) providerCatalog(availableSet map[string]struct{}) map[string]gin.H {
	out := map[string]gin.H{}
	cfg := ctl.reg.Config()
	for name, providerCfg := range cfg.Providers {
		_, loaded := availableSet[name]
		health := ctl.providerDependencyStatus(name)
		available := health.Status == "healthy" || health.Status == "degraded" || health.Status == "unknown"
		models := []string{}
		models = ctl.resolveCandidateModels(name, unifiedRequest{})
		defaultModel := ""
		if len(models) > 0 {
			defaultModel = models[0]
		} else {
			defaultModel = providerDefaultModel(providerCfg)
		}

		status := "needs_api_key"
		mode := "byok"
		if loaded {
			status = "loaded"
			mode = "server"
		}
		if available {
			status = "ready"
			mode = "server"
		} else if health.Status == "unhealthy" {
			status = "offline"
			mode = "server"
		} else if providerCfg != nil && providerCfg.KeyEnv != "" && os.Getenv(providerCfg.KeyEnv) != "" {
			status = "offline"
			mode = "offline"
		}

		out[name] = gin.H{
			"name":          name,
			"display_name":  providerDisplayName(name),
			"available":     available,
			"configured":    available || hasConfiguredKey(providerCfg),
			"models":        models,
			"endpoint":      providerEndpoint(providerCfg),
			"default_model": defaultModel,
			"status":        status,
			"mode":          mode,
			"supports_byok": providerSupportsBYOK(name),
		}
	}
	return out
}

func (ctl *runtimeAIController) executeUnified(
	ctx context.Context,
	providerName string,
	modelCandidates []string,
	req unifiedRequest,
) (string, string, *kbxTypes.Usage, []providerAttempt, error) {
	var attempts []providerAttempt

	for idx, modelName := range modelCandidates {
		start := time.Now()
		chatReq := buildChatRequest(ctx, providerName, modelName, req, false)
		ch, err := ctl.reg.Chat(ctx, chatReq)
		if err != nil {
			ctl.recordProviderCheck(providerName, start, err)
			attempts = append(attempts, ctl.describeAttempt(providerName, modelName, err))
			if !shouldRetryWithNextModel(err) || idx == len(modelCandidates)-1 {
				return "", modelName, nil, attempts, err
			}
			ctl.logModelFallback(providerName, modelName, err, idx+1 < len(modelCandidates), modelCandidates)
			continue
		}

		var builder strings.Builder
		var usage *kbxTypes.Usage
		var streamErr error
		for chunk := range ch {
			if chunk.Error != "" {
				streamErr = gl.Errorf("provider=%s model=%s error=%s", providerName, modelName, chunk.Error)
				break
			}
			if chunk.Content != "" {
				builder.WriteString(chunk.Content)
			}
			if chunk.Done && chunk.Usage != nil {
				usage = chunk.Usage
			}
		}

		if streamErr != nil {
			ctl.recordProviderCheck(providerName, start, streamErr)
			attempts = append(attempts, ctl.describeAttempt(providerName, modelName, streamErr))
			if !shouldRetryWithNextModel(streamErr) || idx == len(modelCandidates)-1 {
				return "", modelName, nil, attempts, streamErr
			}
			ctl.logModelFallback(providerName, modelName, streamErr, idx+1 < len(modelCandidates), modelCandidates)
			continue
		}

		ctl.recordProviderCheck(providerName, start, nil)
		if idx > 0 {
			gl.Infof("provider=%s model=%s request recovered after fallback", providerName, modelName)
		}
		return builder.String(), modelName, usage, attempts, nil
	}

	return "", "", nil, attempts, gl.Errorf("all model candidates failed for provider '%s'", providerName)
}

func (ctl *runtimeAIController) streamUnified(
	c *gin.Context,
	providerName string,
	modelCandidates []string,
	req unifiedRequest,
) (string, *kbxTypes.Usage, []providerAttempt, error) {
	flusher := c.Writer.(http.Flusher)
	var attempts []providerAttempt

	for idx, modelName := range modelCandidates {
		start := time.Now()
		chatReq := buildChatRequest(c.Request.Context(), providerName, modelName, req, true)
		ch, err := ctl.reg.Chat(c.Request.Context(), chatReq)
		if err != nil {
			ctl.recordProviderCheck(providerName, start, err)
			attempts = append(attempts, ctl.describeAttempt(providerName, modelName, err))
			if !shouldRetryWithNextModel(err) || idx == len(modelCandidates)-1 {
				return modelName, nil, attempts, err
			}
			ctl.logModelFallback(providerName, modelName, err, idx+1 < len(modelCandidates), modelCandidates)
			continue
		}

		var usage *kbxTypes.Usage
		var streamErr error
		hasEmittedContent := false
		for chunk := range ch {
			if chunk.Error != "" {
				streamErr = gl.Errorf("provider=%s model=%s error=%s", providerName, modelName, chunk.Error)
				break
			}
			if chunk.Content != "" {
				hasEmittedContent = true
				writeSSE(c.Writer, gin.H{"chunk": chunk.Content, "content": chunk.Content, "done": false, "provider": providerName, "model": modelName})
				flusher.Flush()
			}
			if chunk.Done && chunk.Usage != nil {
				usage = chunk.Usage
			}
		}

		if streamErr != nil {
			ctl.recordProviderCheck(providerName, start, streamErr)
			attempts = append(attempts, ctl.describeAttempt(providerName, modelName, streamErr))
			if hasEmittedContent || !shouldRetryWithNextModel(streamErr) || idx == len(modelCandidates)-1 {
				return modelName, nil, attempts, streamErr
			}
			ctl.logModelFallback(providerName, modelName, streamErr, idx+1 < len(modelCandidates), modelCandidates)
			continue
		}

		ctl.recordProviderCheck(providerName, start, nil)
		return modelName, usage, attempts, nil
	}

	return "", nil, attempts, gl.Errorf("all model candidates failed for provider '%s'", providerName)
}

func buildChatRequest(ctx context.Context, providerName, modelName string, req unifiedRequest, stream bool) kbxTypes.ChatRequest {
	temp := float32(0.7)
	headers := map[string]string{}
	_ = ctx
	if strings.EqualFold(strings.TrimSpace(req.PurposeType), "dashboard_schema") {
		temp = 0.2
	}
	if req.Meta != nil {
		if value, ok := req.Meta["purpose_type"].(string); ok && strings.EqualFold(strings.TrimSpace(value), "dashboard_schema") {
			temp = 0.2
		}
	}
	messages := req.Messages
	if len(messages) == 0 {
		content := strings.TrimSpace(req.Prompt)
		if content == "" {
			content = buildStructuredPrompt(req)
		}
		messages = []kbxTypes.Message{{Role: "user", Content: content}}
	}

	if req.Meta == nil {
		req.Meta = map[string]any{}
	}
	if strings.TrimSpace(req.Lang) != "" {
		req.Meta["lang"] = req.Lang
	}
	if strings.TrimSpace(req.Purpose) != "" {
		req.Meta["purpose"] = req.Purpose
	}
	if strings.TrimSpace(req.PurposeType) != "" {
		req.Meta["purpose_type"] = req.PurposeType
	}
	if req.MaxTokens > 0 {
		req.Meta["max_tokens"] = req.MaxTokens
	}

	return kbxTypes.ChatRequest{
		Provider: providerName,
		Model:    modelName,
		Messages: messages,
		Temp:     temp,
		Stream:   stream,
		Meta:     req.Meta,
		Headers:  headers,
	}
}

func (ctl *runtimeAIController) resolveCandidateModels(providerName string, req unifiedRequest) []string {
	candidates := []string{}
	if model := strings.TrimSpace(req.Model); model != "" {
		candidates = append(candidates, model)
	}
	candidates = append(candidates, extractCandidateModels(req.Meta)...)
	candidates = append(candidates, parseModelCandidatesEnv(providerName)...)
	if fallback := defaultModelForProvider(ctl.reg, providerName); fallback != "" {
		candidates = append(candidates, fallback)
	}
	return uniqueStrings(candidates)
}

func buildStructuredPrompt(req unifiedRequest) string {
	lang := strings.TrimSpace(req.Lang)
	if lang == "" {
		lang = "portuguese"
	}
	purpose := strings.TrimSpace(req.Purpose)
	if purpose == "" {
		purpose = strings.TrimSpace(req.PurposeType)
	}
	if purpose == "" {
		purpose = "general"
	}

	ideas := make([]string, 0, len(req.Ideas))
	for _, idea := range req.Ideas {
		idea = strings.TrimSpace(idea)
		if idea != "" {
			ideas = append(ideas, "- "+idea)
		}
	}
	if len(ideas) == 0 {
		return "Provide a concise helpful answer."
	}

	return strings.Join([]string{
		fmt.Sprintf("You are an expert prompt engineer. Respond in %s.", lang),
		fmt.Sprintf("Craft a high-quality structured prompt for the following purpose: %s.", purpose),
		"Use clear sections, practical constraints, and produce a prompt ready to paste into another AI tool.",
		"",
		"Input ideas:",
		strings.Join(ideas, "\n"),
	}, "\n")
}

func writeSSE(w http.ResponseWriter, payload any) {
	switch value := payload.(type) {
	case string:
		_, _ = fmt.Fprintf(w, "data: %s\n\n", value)
	default:
		data, _ := json.Marshal(value)
		_, _ = fmt.Fprintf(w, "data: %s\n\n", data)
	}
}

func (ctl *runtimeAIController) describeAttempt(providerName, modelName string, err error) providerAttempt {
	info := classifyUpstreamError(err)
	return providerAttempt{
		Provider: providerName,
		Model:    modelName,
		Class:    info.Class,
		Message:  info.Summary,
	}
}

func toUnifiedUsage(usage *kbxTypes.Usage) gin.H {
	if usage == nil {
		return nil
	}
	return gin.H{
		"prompt_tokens":     usage.Prompt,
		"completion_tokens": usage.Completion,
		"total_tokens":      usage.Tokens,
		"estimated_cost":    usage.CostUSD,
		"latency_ms":        usage.Ms,
	}
}

func providerAvailabilityError(reg *registry.Registry, name string) string {
	provider := reg.ResolveProvider(name)
	if provider == nil {
		return "provider not loaded"
	}
	if err := provider.Available(); err != nil {
		return err.Error()
	}
	return ""
}

func (ctl *runtimeAIController) providerDependencyStatus(providerName string) dependencyStatus {
	if ctl.prod != nil {
		if monitor := ctl.prod.GetHealthMonitor(); monitor != nil {
			if check, ok := monitor.GetHealth(providerName); ok {
				status := strings.ToLower(check.Status.String())
				if status == "unknown" {
					return dependencyStatus{
						Status: "unknown",
					}
				}
				return dependencyStatus{
					Status: status,
					Error:  strings.TrimSpace(check.ErrorMsg),
				}
			}
		}
	}

	status := dependencyStatus{Status: "unknown"}
	if provider := ctl.reg.ResolveProvider(providerName); provider == nil {
		status.Status = "unavailable"
	} else if err := provider.Available(); err != nil {
		status.Status = "unhealthy"
		status.Error = err.Error()
	}
	return status
}

func (ctl *runtimeAIController) recordProviderCheck(providerName string, start time.Time, err error) {
	if ctl.prod == nil {
		return
	}
	monitor := ctl.prod.GetHealthMonitor()
	if monitor == nil {
		return
	}
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	monitor.RecordCheck(providerName, err == nil, time.Since(start), errMsg)
}

func defaultModelForProvider(reg *registry.Registry, name string) string {
	if cfg := reg.GetProviderConfig(name); cfg != nil {
		if model := strings.TrimSpace(cfg.DefaultModel); model != "" {
			return model
		}
	}
	if provider := reg.ResolveProvider(name); provider != nil {
		if models, err := provider.ListModels(context.Background()); err == nil && len(models) > 0 {
			return strings.TrimSpace(models[0])
		}
	}
	return ""
}

func providerReady(providers map[string]gin.H, name string) bool {
	if item, ok := providers[name]; ok {
		if value, ok := item["available"].(bool); ok {
			return value
		}
	}
	return false
}

func providerType(name string, cfg *kbxTypes.LLMProviderConfig) string {
	if cfg != nil {
		if typ := strings.TrimSpace(cfg.Type()); typ != "" {
			return typ
		}
	}
	return strings.ToLower(strings.TrimSpace(name))
}

func providerDefaultModel(cfg *kbxTypes.LLMProviderConfig) string {
	if cfg == nil {
		return ""
	}
	return strings.TrimSpace(cfg.DefaultModel)
}

func providerEndpoint(cfg *kbxTypes.LLMProviderConfig) string {
	if cfg == nil {
		return ""
	}
	return strings.TrimSpace(cfg.URLBase())
}

func hasConfiguredKey(cfg *kbxTypes.LLMProviderConfig) bool {
	return cfg != nil && cfg.KeyEnv != "" && os.Getenv(cfg.KeyEnv) != ""
}

func providerSupportsBYOK(name string) bool {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "openai", "gemini", "groq", "anthropic", "claude", "chatgpt", "deepseek":
		return true
	default:
		return false
	}
}

func providerDisplayName(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "openai", "chatgpt":
		return "OpenAI"
	case "gemini":
		return "Google Gemini"
	case "groq":
		return "Groq"
	case "anthropic", "claude":
		return "Anthropic Claude"
	case "deepseek":
		return "DeepSeek"
	case "ollama":
		return "Ollama"
	default:
		return strings.Title(strings.ReplaceAll(name, "_", " "))
	}
}

func shouldRetryWithNextModel(err error) bool {
	info := classifyUpstreamError(err)
	return info.Retryable && info.ModelScoped
}

func classifyUpstreamError(err error) upstreamErrorInfo {
	if err == nil {
		return upstreamErrorInfo{Class: "ok", Summary: "ok"}
	}

	raw := strings.ToLower(strings.TrimSpace(err.Error()))
	info := upstreamErrorInfo{
		Class:   "provider_error",
		Summary: strings.TrimSpace(err.Error()),
	}

	switch {
	case strings.Contains(raw, "invalid api key"), strings.Contains(raw, "invalid_api_key"):
		info.Class = "invalid_api_key"
		info.Summary = "provider API key is invalid"
	case strings.Contains(raw, "reported as leaked"):
		info.Class = "compromised_api_key"
		info.Summary = "provider API key was flagged as leaked"
	case strings.Contains(raw, "resource_exhausted"), strings.Contains(raw, "quota exceeded"), strings.Contains(raw, "rate limit"):
		info.Class = "quota_exhausted"
		info.Summary = "provider quota or rate limit exhausted"
		info.Retryable = true
		info.ModelScoped = true
	case strings.Contains(raw, "model_permission_blocked_project"), strings.Contains(raw, "blocked at the project level"):
		info.Class = "model_blocked"
		info.Summary = "model is blocked for the current provider project"
		info.Retryable = true
		info.ModelScoped = true
	case strings.Contains(raw, "permission_denied"), strings.Contains(raw, "permissions_error"), strings.Contains(raw, "api key sem permissão"):
		info.Class = "permission_denied"
		info.Summary = "provider credentials do not have permission for this operation"
	case strings.Contains(raw, "model not found"), strings.Contains(raw, "unknown model"), strings.Contains(raw, "does not exist"), strings.Contains(raw, "is not found"), strings.Contains(raw, "not supported for generatecontent"):
		info.Class = "model_not_found"
		info.Summary = "requested model is not available"
		info.Retryable = true
		info.ModelScoped = true
	case strings.Contains(raw, "context deadline exceeded"), strings.Contains(raw, "timeout"), strings.Contains(raw, "connection reset"), strings.Contains(raw, "temporary unavailable"):
		info.Class = "transient_upstream_error"
		info.Summary = "provider request failed due to transient upstream condition"
		info.Retryable = true
	}

	return info
}

func (ctl *runtimeAIController) logModelFallback(providerName, modelName string, err error, hasNext bool, candidates []string) {
	info := classifyUpstreamError(err)
	gl.Warnf(
		"provider request failed provider=%s model=%s class=%s retryable=%v model_scoped=%v fallback_remaining=%v summary=%s raw=%s candidates=%v",
		providerName,
		modelName,
		info.Class,
		info.Retryable,
		info.ModelScoped,
		hasNext,
		info.Summary,
		strings.TrimSpace(err.Error()),
		candidates,
	)
}

func extractCandidateModels(meta map[string]any) []string {
	if len(meta) == 0 {
		return nil
	}
	raw, ok := meta["candidate_models"]
	if !ok {
		raw, ok = meta["models"]
		if !ok {
			return nil
		}
	}

	switch value := raw.(type) {
	case string:
		return splitCommaSeparated(value)
	case []string:
		return uniqueStrings(value)
	case []any:
		out := make([]string, 0, len(value))
		for _, item := range value {
			if str, ok := item.(string); ok {
				out = append(out, str)
			}
		}
		return uniqueStrings(out)
	default:
		return nil
	}
}

func parseModelCandidatesEnv(providerName string) []string {
	envName := "KUBEX_GNYX_PROVIDER_MODELS_" + sanitizeEnvKey(providerName)
	return splitCommaSeparated(os.Getenv(envName))
}

func splitCommaSeparated(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return uniqueStrings(out)
}

func sanitizeEnvKey(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	var b strings.Builder
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			continue
		}
		b.WriteRune('_')
	}
	return b.String()
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
