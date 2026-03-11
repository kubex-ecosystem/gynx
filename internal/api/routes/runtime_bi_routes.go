package routes

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kubex-ecosystem/gnyx/internal/config"
	"github.com/kubex-ecosystem/gnyx/internal/features/bi"
	"github.com/kubex-ecosystem/gnyx/internal/features/providers/registry"
	runtimeMW "github.com/kubex-ecosystem/gnyx/internal/runtime/middlewares"
	"github.com/kubex-ecosystem/gnyx/internal/types"
	kbxTypes "github.com/kubex-ecosystem/kbx/types"
)

type runtimeBIController struct {
	cfg       *config.Config
	ai        *runtimeAIController
	grounding *bi.Service
}

type boardGenerateRequest struct {
	Prompt     string `json:"prompt"`
	Domain     string `json:"domain"`
	Provider   string `json:"provider"`
	Model      string `json:"model"`
	MaxWidgets int    `json:"max_widgets"`
}

func registerRuntimeBIRoutes(
	r *gin.RouterGroup,
	container types.IContainer,
	reg *registry.Registry,
	prod *runtimeMW.ProductionMiddleware,
) {
	cfg, _ := container.Config().(*config.Config)
	ctl := &runtimeBIController{
		cfg:       cfg,
		ai:        &runtimeAIController{cfg: cfg, reg: reg, prod: prod},
		grounding: bi.NewService(""),
	}

	r.POST("/bi/boards/generate", ctl.generateBoard)
}

func (ctl *runtimeBIController) generateBoard(c *gin.Context) {
	var req boardGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload", "message": err.Error()})
		return
	}
	if strings.TrimSpace(req.Prompt) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prompt is required"})
		return
	}
	if strings.TrimSpace(req.Domain) == "" {
		req.Domain = "sales"
	}
	if req.Domain != "sales" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only domain 'sales' is supported in this slice"})
		return
	}
	if req.MaxWidgets <= 0 || req.MaxWidgets > 6 {
		req.MaxWidgets = 4
	}

	grounding, err := ctl.grounding.BuildSalesCommercialContext(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load grounding context", "message": err.Error()})
		return
	}
	prompt, err := bi.BuildPlanningPrompt(req.Prompt, req.MaxWidgets, grounding)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build planning prompt", "message": err.Error()})
		return
	}

	unifiedReq := unifiedRequest{
		Provider: req.Provider,
		Model:    req.Model,
		Messages: []kbxTypes.Message{
			{Role: "system", Content: bi.InitialSystemPrompt()},
			{Role: "user", Content: prompt},
		},
		Meta: map[string]any{
			"purpose":      "bi_board_generation",
			"purpose_type": "dashboard_schema",
			"domain":       req.Domain,
		},
	}

	providerName := strings.TrimSpace(req.Provider)
	if providerName == "" {
		providerName = firstProvider(ctl.ai.reg.ListProviders())
	}
	if providerName == "" {
		c.JSON(http.StatusBadGateway, gin.H{"error": "no providers are available in the active registry"})
		return
	}
	modelCandidates := ctl.ai.resolveCandidateModels(providerName, unifiedReq)
	if len(modelCandidates) == 0 {
		c.JSON(http.StatusBadGateway, gin.H{"error": "provider has no model candidates configured", "provider": providerName})
		return
	}

	content, modelName, usage, attempts, err := ctl.ai.executeUnified(c.Request.Context(), providerName, modelCandidates, unifiedReq)
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

	plan, parseErr := bi.ParseBoardPlan(content)
	generationMode := "llm"
	fallbackReason := ""
	if parseErr != nil {
		generationMode = "fallback_template"
		fallbackReason = parseErr.Error()
		plan = bi.FallbackSalesOverviewPlan(req.Prompt, grounding, fallbackReason)
	} else if validateErr := bi.ValidateBoardPlan(plan, grounding); validateErr != nil {
		generationMode = "fallback_template"
		fallbackReason = validateErr.Error()
		plan = bi.FallbackSalesOverviewPlan(req.Prompt, grounding, fallbackReason)
	}
	compiled, err := bi.CompileDashboardSchema(plan)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to compile dashboard schema", "message": err.Error(), "plan": plan, "provider": providerName, "model": modelName})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"domain":            req.Domain,
		"provider":          providerName,
		"model":             modelName,
		"usage":             toUnifiedUsage(usage),
		"attempts":          attempts,
		"generation_mode":   generationMode,
		"fallback_reason":   fallbackReason,
		"grounding_context": grounding,
		"plan":              plan,
		"dashboard_schema":  compiled,
		"raw_provider_json": content,
	})
}

func firstProvider(items []string) string {
	if len(items) == 0 {
		return ""
	}
	return items[0]
}
