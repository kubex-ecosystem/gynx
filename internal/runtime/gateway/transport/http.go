// Package transport sets up HTTP routes and handlers for the GNyx Gateway,
// including merged Repository Intelligence endpoints.
package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	inviteapi "github.com/kubex-ecosystem/gnyx/internal/api/invite"
	"github.com/kubex-ecosystem/gnyx/internal/api/routes"
	"github.com/kubex-ecosystem/gnyx/internal/app"
	"github.com/kubex-ecosystem/gnyx/internal/config"
	"github.com/kubex-ecosystem/gnyx/internal/features/health"
	"github.com/kubex-ecosystem/gnyx/internal/features/lookatni"
	"github.com/kubex-ecosystem/gnyx/internal/features/providers/registry"
	"github.com/kubex-ecosystem/gnyx/internal/features/ui"
	"github.com/kubex-ecosystem/gnyx/internal/runtime/middlewares"

	// providers "github.com/kubex-ecosystem/gnyx/internal/types"
	gl "github.com/kubex-ecosystem/logz"

	"github.com/kubex-ecosystem/gnyx/internal/web"
	"github.com/kubex-ecosystem/gnyx/internal/webhook"
)

var w http.ResponseWriter

// httpHandlers holds the HTTP route handlers
type httpHandlers struct {
	registry             *registry.Registry
	productionMiddleware *middlewares.ProductionMiddleware
	lookAtniHandler      *lookatni.Handler    // LookAtni integration
	webhookHandler       *webhook.HTTPHandler // Meta-recursive webhook handler
	healthEngine         *health.Engine       // AI Provider health monitoring
	healthRegistry       *health.ProberRegistry
	healthScheduler      *health.Scheduler // Background health checks
	inviteService        inviteapi.Service
}

// WireHTTP sets up HTTP routes
func WireHTTP(mux *gin.Engine, reg *registry.Registry, prodMiddleware *middlewares.ProductionMiddleware, container *app.Container) {
	// Initialize LookAtni handler
	workDir := "./lookatni_workspace" // TODO: Make configurable
	lookAtniHandler := lookatni.NewHandler(workDir)

	// Initialize webhook handler (mock for now - TODO: implement real actors)
	webhookHandler := webhook.NewHTTPHandler(nil) // TODO: Initialize with real handler

	// Initialize AI Provider Health Monitoring
	healthStore := health.NewStore()
	healthRegistry := health.NewProberRegistry()

	// Registra probers no registry local
	healthRegistry.Register(health.NewGroqProber())
	healthRegistry.Register(health.NewGeminiProber())

	// Cria engine com probers do registry
	groqProber := health.NewGroqProber()
	geminiProber := health.NewGeminiProber()
	healthEngine := health.NewEngine(healthStore, groqProber, geminiProber)

	// Initialize Background Health Scheduler - ARQUITETURA QUE NÃO SE SABOTA! 🔥
	schedulerConfig := health.DefaultSchedulerConfig()
	schedulerConfig.LogVerbose = false // Production mode
	healthScheduler := health.NewScheduler(healthEngine, healthRegistry, schedulerConfig)

	// Start scheduler in background
	if err := healthScheduler.Start(); err != nil {
		gl.Log("error", "Failed to start health scheduler: %v", err)
	}

	var inviteSvc inviteapi.Service
	if container == nil {
		ctx := context.Background()
		ctnr, err := app.NewContainer(ctx, config.LoadConfig())
		if err != nil {
			gl.Log("error", "Failed to create app container: %v", err)
			return
		}
		if err = ctnr.Bootstrap(ctx); err != nil {
			gl.Log("error", "Failed to bootstrap app container: %v", err)
			return
		}
		container = ctnr
	}
	if container == nil {
		gl.Log("error", "Failed to create app container")
		return
	}

	inviteSvc = container.InviteService().(inviteapi.Service)

	h := &httpHandlers{
		registry:             reg,
		productionMiddleware: prodMiddleware,
		// engine:               nil, // TODO: Initialize scorecard engine with real clients
		lookAtniHandler: lookAtniHandler,
		webhookHandler:  webhookHandler,
		healthEngine:    healthEngine,
		healthRegistry:  healthRegistry,
		healthScheduler: healthScheduler,
		inviteService:   inviteSvc,
	}

	// API endpoints (higher priority routes)
	mux.GET("/healthz", h.healthCheckGin)
	mux.GET("/v1/providers", h.listProvidersGin)

	// Repository Intelligence endpoints - MERGE POINT! 🚀
	mux.Any("/api/v1/scorecard", gin.WrapF(h.handleRepositoryScorecard))
	mux.Any("/api/v1/scorecard/advice", gin.WrapF(h.handleScorecardAdvice))
	mux.Any("/api/v1/metrics/ai", gin.WrapF(h.handleAIMetrics))
	mux.Any("/api/v1/health", gin.WrapF(h.handleRepositoryHealth))

	// AI Provider Health Monitoring - ARQUITETURA QUE NÃO SE SABOTA! 🔥
	health.RegisterRoutes(mux, h.healthEngine, h.healthRegistry)
	mux.Any("/health/scheduler/stats", gin.WrapF(h.handleSchedulerStats))
	mux.Any("/health/scheduler/force", gin.WrapF(h.handleSchedulerForce))

	// LookAtni Integration endpoints - CODE NAVIGATION! 🔍
	mux.Any("/api/v1/lookatni/extract", gin.WrapF(h.lookAtniHandler.HandleExtractProject))
	mux.Any("/api/v1/lookatni/archive", gin.WrapF(h.lookAtniHandler.HandleCreateArchive))
	mux.Any("/api/v1/lookatni/download/*filepath", gin.WrapF(h.lookAtniHandler.HandleDownloadArchive))
	mux.Any("/api/v1/lookatni/projects", gin.WrapF(h.lookAtniHandler.HandleListExtractedProjects))
	mux.Any("/api/v1/lookatni/projects/*filepath", gin.WrapF(h.lookAtniHandler.HandleProjectFragments))

	// Meta-Recursive Webhook endpoints - INSANIDADE RACIONAL! 🔄
	mux.Any("/v1/webhooks", gin.WrapF(h.webhookHandler.HandleWebhook))
	mux.Any("/v1/webhooks/health", gin.WrapF(h.webhookHandler.HealthCheck))

	// OAuth callbacks (root) -> API handlers
	mux.GET("/auth/v1/callback", func(c *gin.Context) {
		target := "/api/v1/auth/v1/callback"
		if c.Request.URL.RawQuery != "" {
			target = target + "?" + c.Request.URL.RawQuery
		}
		c.Redirect(http.StatusFound, target)
	})
	mux.GET("/auth/google/callback", func(c *gin.Context) {
		target := "/api/v1/auth/google/callback"
		if c.Request.URL.RawQuery != "" {
			target = target + "?" + c.Request.URL.RawQuery
		}
		c.Redirect(http.StatusFound, target)
	})

	if h.inviteService != nil {
		h.registerInviteRoutes(mux)
	}

	initData := container.GetConfig().ServerConfig
	gl.Debugf("HTTP server initialized on %s:%s (bind=%s) [env=%s, debug=%v, release=%v, confidential=%v]",
		initData.Runtime.Host, initData.Runtime.Port, initData.Runtime.Bind, initData.Files.EnvFile, initData.Basic.Debug, initData.Basic.ReleaseMode, initData.Basic.IsConfidential)

	routes.RegisterRoutes(mux.Group("/api/v1"), container)

	// Web Interface - Frontend embarcado! (registrado por último para não capturar /invite) - Se habilitado
	uiSvc, ok := container.UIService().(*ui.UIService)
	if !ok || uiSvc == nil {
		gl.Log("warn", "UIService is not available. Web interface will be disabled.")
		return
	}

	webHandler, err := web.NewHandler(uiSvc.GetWebFS())
	if err != nil {
		gl.Log("warn", "Failed to initialize web interface: %v", err)
	} else {
		// Register web interface on /app/* and root using Gin adapters
		mux.Any("/app/*filepath", gin.WrapH(http.StripPrefix("/app/", webHandler)))
		// Root path serves the frontend (mas depois de /invite e APIs)
		mux.Any("/", gin.WrapH(webHandler))
		gl.Log("info", "Web interface enabled at /app/ and /")
	}

	gl.Info("AI Provider Health Monitoring enabled")
}

// healthCheck provides a simple health endpoint
func (h *httpHandlers) healthCheckGin(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "gnyx-gw",
	})
}

// listProviders returns available providers with health status
func (h *httpHandlers) listProvidersGin(c *gin.Context) {
	providerNames := h.registry.ListProviders()
	config := h.registry.Config()

	healthStatuses := make(map[string]interface{})
	if h.productionMiddleware != nil {
		healthMonitor := h.productionMiddleware.GetHealthMonitor()
		if healthMonitor != nil {
			allHealth := healthMonitor.GetAllHealth()
			for providerName, health := range allHealth {
				healthStatuses[providerName] = map[string]interface{}{
					"status":        health.Status.String(),
					"last_check":    health.LastCheck,
					"response_time": health.ResponseTime.String(),
					"uptime":        health.Uptime,
					"error":         health.ErrorMsg,
				}
			}
		}
	}

	enrichedProviders := make([]map[string]interface{}, 0, len(providerNames))
	for _, providerName := range providerNames {
		enrichedProvider := map[string]interface{}{
			"name":      providerName,
			"type":      providerName,
			"available": true,
		}
		if healthStatus, exists := healthStatuses[providerName]; exists {
			enrichedProvider["health"] = healthStatus
			if healthMap, ok := healthStatus.(map[string]interface{}); ok {
				if status, ok := healthMap["status"].(string); ok {
					enrichedProvider["available"] = (status == "healthy" || status == "degraded")
				}
			}
		} else {
			enrichedProvider["health"] = map[string]interface{}{
				"status":     "unknown",
				"last_check": nil,
				"uptime":     100.0,
				"error":      "",
			}
		}
		enrichedProviders = append(enrichedProviders, enrichedProvider)
	}

	response := map[string]interface{}{
		"providers": enrichedProviders,
		"config":    config.Providers,
		"timestamp": "2024-01-01T00:00:00Z",
		"service":   "github.com/kubex-ecosystem/gnyx-gateway",
		"version":   "v1.0.0",
	}
	c.JSON(http.StatusOK, response)
}

// chatSSE handles chat completion with Server-Sent Events
// func (h *httpHandlers) chatSSE(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	var req providers.ChatRequest
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
// 		return
// 	}

// 	// Validate required fields
// 	if req.Provider == "" {
// 		http.Error(w, "Provider is required", http.StatusBadRequest)
// 		return
// 	}

// 	provider := h.registry.ResolveProvider(req.Provider)
// 	if provider == nil {
// 		http.Error(w, fmt.Sprintf("Provider '%s' not found", req.Provider), http.StatusBadRequest)
// 		return
// 	}

// 	// Check if provider is available
// 	if err := provider.Available(); err != nil {
// 		http.Error(w, fmt.Sprintf("Provider unavailable: %v", err), http.StatusServiceUnavailable)
// 		return
// 	}

// 	// Handle BYOK (Bring Your Own Key)
// 	if externalKey := r.Header.Get("x-external-api-key"); externalKey != "" {
// 		// TODO: Implement secure BYOK handling
// 		// For now, we'll pass it through meta
// 		if req.Meta == nil {
// 			req.Meta = make(map[string]interface{})
// 		}
// 		req.Meta["external_api_key"] = externalKey
// 	}

// 	// Set default temperature if not provided
// 	if req.Temp == 0 {
// 		req.Temp = 0.7
// 	}

// 	// Force streaming for SSE
// 	req.Stream = true

// 	// Start chat completion
// 	ch, err := provider.Chat(r.Context(), req)
// 	if err != nil {
// 		http.Error(w, fmt.Sprintf("Chat request failed: %v", err), http.StatusInternalServerError)
// 		return
// 	}

// 	// Set SSE headers
// 	w.Header().Set("Content-Type", "text/event-stream")
// 	w.Header().Set("Cache-Control", "no-cache")
// 	w.Header().Set("Connection", "keep-alive")
// 	w.Header().Set("Access-Control-Allow-Origin", "*")

// 	flusher, ok := w.(http.Flusher)
// 	if !ok {
// 		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
// 		return
// 	}

// 	// Create SSE coalescer to improve streaming UX
// 	coalescer := NewSSECoalescer(func(content string) {
// 		data, _ := json.Marshal(map[string]interface{}{
// 			"content": content,
// 			"done":    false,
// 		})
// 		gl.Info(w, "data: %s\n\n", data)
// 		flusher.Flush()
// 	})
// 	defer coalescer.Close()

// 	// Stream the response with coalescence
// 	for chunk := range ch {
// 		if chunk.Error != "" {
// 			// Flush any pending content before error
// 			coalescer.Close()

// 			// Send error event
// 			data, _ := json.Marshal(map[string]interface{}{
// 				"error": chunk.Error,
// 				"done":  true,
// 			})
// 			gl.Info(w, "data: %s\n\n", data)
// 			flusher.Flush()
// 			return
// 		}

// 		if chunk.Content != "" {
// 			// Add to coalescer instead of immediate flush
// 			coalescer.AddChunk(chunk.Content)
// 		}

// 		if chunk.Done {
// 			// Flush any remaining content before final chunk
// 			coalescer.Close()

// 			// Send final chunk with usage info
// 			data, _ := json.Marshal(map[string]interface{}{
// 				"done":  true,
// 				"usage": chunk.Usage,
// 			})
// 			gl.Info(w, "data: %s\n\n", data)
// 			flusher.Flush()

// 			// Log usage for monitoring
// 			if chunk.Usage != nil {
// 				gl.Log("info", "Usage: provider=%s model=%s tokens=%d latency=%dms cost=$%.6f",
// 					chunk.Usage.Provider, chunk.Usage.Model, chunk.Usage.Tokens,
// 					chunk.Usage.Ms, chunk.Usage.CostUSD)
// 			}
// 			break
// 		}
// 	}
// }

// productionStatus returns comprehensive status including middleware metrics
func (h *httpHandlers) productionStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := map[string]interface{}{
		"service":   "gnyx-gw",
		"status":    "healthy",
		"providers": h.registry.ListProviders(),
	}

	// Add production middleware status if available
	if h.productionMiddleware != nil {
		status["production_features"] = h.productionMiddleware.GetStatus()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// handleAdvise handles POST /v1/advise for AI-powered analysis advice
func (h *httpHandlers) handleAdvise(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check BF1_MODE restrictions
	bf1Config := config.GetBF1Config()
	if bf1Config.Enabled {
		// In BF1 mode, add guard-rails and limitations
		w.Header().Set("X-BF1-Mode", "true")
		w.Header().Set("X-BF1-WIP-Cap", fmt.Sprintf("%d", bf1Config.WIPCap))
	}

	// Parse request body
	var req struct {
		Mode     string                 `json:"mode"`     // "exec" or "code"
		Provider string                 `json:"provider"` // optional, defaults to first available
		Context  map[string]interface{} `json:"context"`  // repository, hotspots, scorecard
		Options  map[string]interface{} `json:"options"`  // timeout_sec, temperature
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}

	// Validate mode
	if req.Mode != "exec" && req.Mode != "code" {
		http.Error(w, "Mode must be 'exec' or 'code'", http.StatusBadRequest)
		return
	}

	// Set SSE headers for streaming response
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Generate mock response based on mode
	if req.Mode == "exec" {
		// Simulate exec mode response
		execResponse := map[string]interface{}{
			"summary": map[string]interface{}{
				"grade":               "B+",
				"chi":                 72.5,
				"lead_time_p95_hours": 24.0,
				"deploys_per_week":    3.2,
			},
			"top_focus": []map[string]interface{}{
				{
					"title":      "Reduce Lead Time",
					"why":        "Long deployment cycles impacting delivery velocity",
					"kpi":        "lead_time_p95_hours",
					"target":     "< 12 hours",
					"confidence": 0.85,
				},
			},
			"quick_wins": []map[string]interface{}{
				{
					"action":        "Implement automated testing",
					"effort":        "M",
					"expected_gain": "20% faster deployments",
				},
			},
			"risks": []map[string]interface{}{
				{
					"risk":       "Technical debt accumulation",
					"mitigation": "Schedule dedicated refactoring sprints",
				},
			},
			"call_to_action": "Focus on deployment automation and testing coverage",
		}

		// Send as SSE
		gl.Info(w, "data: %s\n\n", mustMarshal(execResponse))

	} else { // code mode
		// Simulate code mode response
		codeResponse := map[string]interface{}{
			"chi_now": 72.5,
			"drivers": []map[string]interface{}{
				{
					"metric": "cyclomatic_complexity",
					"value":  15.2,
					"impact": "high",
				},
			},
			"refactor_plan": []map[string]interface{}{
				{
					"step":    1,
					"theme":   "Simplify complex functions",
					"actions": []string{"Break down large functions", "Extract common utilities"},
					"kpi":     "cyclomatic_complexity",
					"target":  "< 10",
				},
			},
			"guardrails": []string{"Maintain test coverage > 80%", "No functions > 50 lines"},
			"milestones": []map[string]interface{}{
				{
					"in_days": 14,
					"goal":    "Reduce complexity by 30%",
				},
			},
		}

		// Send as SSE
		gl.Info(w, "data: %s\n\n", mustMarshal(codeResponse))
	}

	// Send completion event
	gl.Info(w, "data: {\"done\": true}\n\n")

	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

// Helper function to marshal JSON (panic on error for simplicity)
func mustMarshal(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(data)
}

// Repository Intelligence Handlers - MERGED! 🚀

// handleRepositoryScorecard handles GET /api/v1/scorecard
func (h *httpHandlers) handleRepositoryScorecard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Implement with real scorecard engine
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Schema-Version", "scorecard@1.0.0")
	w.Header().Set("X-Server-Version", "gnyx-v1.0.0")

	// Placeholder response
	placeholder := map[string]interface{}{
		"status":  "not_implemented",
		"message": "Repository Intelligence API under development",
		"endpoints": []string{
			"/api/v1/scorecard",
			"/api/v1/scorecard/advice",
			"/api/v1/metrics/ai",
		},
	}
	json.NewEncoder(w).Encode(placeholder)
}

// handleScorecardAdvice handles POST /api/v1/scorecard/advice
func (h *httpHandlers) handleScorecardAdvice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Implement advice generation using existing advise system
	w.Header().Set("Content-Type", "application/json")
	placeholder := map[string]interface{}{
		"status":  "not_implemented",
		"message": "Will integrate with existing /v1/advise system",
	}
	json.NewEncoder(w).Encode(placeholder)
}

// handleAIMetrics handles GET /api/v1/metrics/ai
func (h *httpHandlers) handleAIMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Implement AI metrics calculation
	w.Header().Set("Content-Type", "application/json")
	placeholder := map[string]interface{}{
		"status":  "not_implemented",
		"message": "AI Metrics (HIR/AAC/TPH) calculation under development",
	}
	json.NewEncoder(w).Encode(placeholder)
}

// handleRepositoryHealth handles GET /api/v1/health
func (h *httpHandlers) handleRepositoryHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	health := map[string]interface{}{
		"status":  "healthy",
		"service": "repository-intelligence",
		"components": map[string]string{
			"scorecard_engine": "not_initialized",
			"dora_calculator":  "not_initialized",
			"chi_calculator":   "not_initialized",
			"ai_metrics":       "not_initialized",
		},
		"version": "gnyx-v1.0.0",
	}
	json.NewEncoder(w).Encode(health)
}

// handleSchedulerStats returns health scheduler statistics
func (h *httpHandlers) handleSchedulerStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	stats := h.healthScheduler.GetStats()
	json.NewEncoder(w).Encode(stats)
}

// handleSchedulerForce forces immediate health checks for all providers
func (h *httpHandlers) handleSchedulerForce(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.healthScheduler.ForceCheck()

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":  "triggered",
		"message": "Force health check initiated for all providers",
		"time":    time.Now(),
	}
	json.NewEncoder(w).Encode(response)
}

// ----------------------------------------------------------------------
// Invitation, Meta Config e Visit log routes

func (h *httpHandlers) registerInviteRoutes(mux *gin.Engine) {
	// Legacy invite endpoints (email links). The canonical API continues em /api/v1/invites via gin controllers.
	mux.POST("/invite", gin.WrapF(h.handleInviteCollection))
	mux.Any("/invite/*filepath", gin.WrapF(h.handleInviteToken))
	mux.Any("/meta/config", gin.WrapF(h.handleMetaConfig))
	mux.POST("/log/visit", gin.WrapF(h.handleVisitLog))

	// Gin-native aliases to evitar 404 em tokens com caracteres especiais
	mux.GET("/invite/:token", func(c *gin.Context) {
		token := sanitizeInviteToken(c.Param("token"))
		dto, err := h.inviteService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, dto)
	})
	mux.POST("/invite/:token/accept", func(c *gin.Context) {
		token := sanitizeInviteToken(c.Param("token"))
		var req inviteapi.AcceptInviteReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid payload"})
			return
		}
		res, err := h.inviteService.AcceptInvite(c.Request.Context(), token, req)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "error", "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})
}

func (h *httpHandlers) handleInviteCollection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req inviteapi.CreateInviteReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	if strings.TrimSpace(req.InvitedBy) == "" {
		req.InvitedBy = firstNonEmpty(r.Header.Get("X-User-ID"), r.Header.Get("X-Invited-By"))
	}
	if strings.TrimSpace(req.InvitedBy) == "" {
		respondError(w, http.StatusBadRequest, "missing invited_by (use X-User-ID header)")
		return
	}

	dto, err := h.inviteService.CreateInvite(r.Context(), req)
	if err != nil {
		respondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, dto)
}

func (h *httpHandlers) handleInviteToken(w http.ResponseWriter, r *http.Request) {
	token, action := parseInvitePath(r.URL.Path)
	if token == "" {
		respondError(w, http.StatusBadRequest, "missing invite token")
		return
	}
	token = sanitizeInviteToken(token)

	switch {
	case r.Method == http.MethodGet && action == "":
		dto, err := h.inviteService.ValidateToken(r.Context(), token)
		if err != nil {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondJSON(w, http.StatusOK, dto)
	case r.Method == http.MethodPost && action == "accept":
		var req inviteapi.AcceptInviteReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid payload")
			return
		}
		res, err := h.inviteService.AcceptInvite(r.Context(), token, req)
		if err != nil {
			respondError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		respondJSON(w, http.StatusOK, res)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

var clientMetaConfig = map[string]map[string]interface{}{
	"default": {
		"brand": map[string]string{
			"name":          "Kubex",
			"primary_color": "#6A0DAD",
			"accent_color":  "#008CFF",
			"logo":          "https://cdn.kubex.world/assets/logo-light.svg",
		},
		"features": []string{"invites", "analytics", "academy"},
	},
	"kortex": {
		"brand": map[string]string{
			"name":          "Kortex",
			"primary_color": "#1F2937",
			"accent_color":  "#F97316",
			"logo":          "https://cdn.kubex.world/assets/clients/kortex.svg",
		},
		"features": []string{"invites", "academy"},
	},
	"pulse": {
		"brand": map[string]string{
			"name":          "Pulse",
			"primary_color": "#111827",
			"accent_color":  "#10B981",
			"logo":          "https://cdn.kubex.world/assets/clients/pulse.svg",
		},
		"features": []string{"invites", "analytics"},
	},
}

func (h *httpHandlers) handleMetaConfig(w http.ResponseWriter, r *http.Request) {
	client := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("client")))
	if client == "" {
		client = hostToClient(r.Host)
	}
	cfg, ok := clientMetaConfig[client]
	if !ok {
		cfg = clientMetaConfig["default"]
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{"status": "ok", "data": cfg})
}

func (h *httpHandlers) handleVisitLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	gl.Log("info", fmt.Sprintf("visit_log: %v", payload))
	respondJSON(w, http.StatusCreated, map[string]any{"status": "ok"})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]interface{}{"status": "error", "message": message})
}

func parseInvitePath(path string) (token string, action string) {
	var trimmed string
	if strings.HasPrefix(path, "/invite") {
		trimmed = strings.TrimPrefix(path, "/invite")
	} else {
		return "", ""
	}

	segments := strings.Split(strings.Trim(trimmed, "/"), "/")
	if len(segments) == 0 || segments[0] == "" {
		return "", ""
	}
	token = segments[0]
	if len(segments) > 1 {
		action = segments[1]
	}
	return token, action
}

// sanitizeInviteToken remove artefatos comuns (ex: "=" de quebra de linha em e-mails)
// sem alterar a compatibilidade com tokens atuais (hex).
func sanitizeInviteToken(token string) string {
	return strings.ReplaceAll(strings.TrimSpace(token), "=", "")
}

func hostToClient(host string) string {
	host = strings.Split(host, ":")[0]
	switch {
	case strings.Contains(host, "kortex"):
		return "kortex"
	case strings.Contains(host, "pulse"):
		return "pulse"
	default:
		return "default"
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
