package routes

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

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

type boardExportRequest struct {
	Prompt          string               `json:"prompt"`
	Domain          string               `json:"domain"`
	Provider        string               `json:"provider"`
	Model           string               `json:"model"`
	GenerationMode  string               `json:"generation_mode"`
	FallbackReason  string               `json:"fallback_reason"`
	Usage           map[string]any       `json:"usage"`
	Grounding       *bi.GroundingContext `json:"grounding_context"`
	Plan            *bi.BoardPlan        `json:"plan"`
	DashboardSchema *bi.DashboardSchema  `json:"dashboard_schema"`
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
	r.GET("/bi/catalog/status", ctl.catalogStatus)
	r.POST("/bi/boards/export", ctl.exportBoardBundle)
}

func (ctl *runtimeBIController) catalogStatus(c *gin.Context) {
	status, err := ctl.grounding.GetCatalogStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load catalog status", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, status)
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
		recovered, recoverErr := bi.RecoverBoardPlan(content, req.Prompt, req.Domain, grounding)
		if recoverErr == nil {
			generationMode = "llm_recovered"
			fallbackReason = parseErr.Error()
			plan = recovered
		} else {
			generationMode = "fallback_template"
			fallbackReason = parseErr.Error()
			plan = bi.FallbackSalesOverviewPlan(req.Prompt, grounding, fallbackReason)
		}
	} else if validateErr := bi.ValidateBoardPlan(plan, grounding); validateErr != nil {
		recovered, recoverErr := bi.RecoverBoardPlan(content, req.Prompt, req.Domain, grounding)
		if recoverErr == nil {
			generationMode = "llm_recovered"
			fallbackReason = validateErr.Error()
			plan = recovered
		} else {
			generationMode = "fallback_template"
			fallbackReason = validateErr.Error()
			plan = bi.FallbackSalesOverviewPlan(req.Prompt, grounding, fallbackReason)
		}
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

func (ctl *runtimeBIController) exportBoardBundle(c *gin.Context) {
	var req boardExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload", "message": err.Error()})
		return
	}
	if req.Plan == nil || req.DashboardSchema == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "plan and dashboard_schema are required"})
		return
	}

	filename := buildBoardBundleFilename(req.Plan.BoardTitle)
	archiveData, err := buildBoardBundleArchive(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build board bundle", "message": err.Error()})
		return
	}

	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Data(http.StatusOK, "application/zip", archiveData)
}

func buildBoardBundleFilename(title string) string {
	base := strings.TrimSpace(title)
	if base == "" {
		base = "generated-board"
	}
	return fmt.Sprintf("%s-%s.zip", slugifyBundleName(base), time.Now().UTC().Format("20060102150405"))
}

func slugifyBundleName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer(" ", "-", "_", "-", "/", "-", "\\", "-", ":", "-", ".", "-")
	value = replacer.Replace(value)
	value = strings.Trim(value, "-")
	if value == "" {
		return "generated-board"
	}
	return value
}

func buildBoardBundleArchive(req boardExportRequest) ([]byte, error) {
	var buffer bytes.Buffer
	zw := zip.NewWriter(&buffer)

	metadata := map[string]any{
		"prompt":          req.Prompt,
		"domain":          req.Domain,
		"provider":        req.Provider,
		"model":           req.Model,
		"generation_mode": req.GenerationMode,
		"fallback_reason": req.FallbackReason,
		"usage":           req.Usage,
		"exported_at":     time.Now().UTC().Format(time.RFC3339),
	}

	if err := writeZipJSON(zw, "dashboard-schema.json", req.DashboardSchema); err != nil {
		return nil, err
	}
	if err := writeZipJSON(zw, "board-plan.json", req.Plan); err != nil {
		return nil, err
	}
	if req.Grounding != nil {
		if err := writeZipJSON(zw, "grounding-context.json", req.Grounding); err != nil {
			return nil, err
		}
	}
	if err := writeZipJSON(zw, "generation-metadata.json", metadata); err != nil {
		return nil, err
	}
	if err := writeZipText(zw, "README.md", buildBoardBundleREADME(req)); err != nil {
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func writeZipJSON(zw *zip.Writer, name string, payload any) error {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return writeZipText(zw, name, string(data)+"\n")
}

func writeZipText(zw *zip.Writer, name string, content string) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(content))
	return err
}

func buildBoardBundleREADME(req boardExportRequest) string {
	title := "Generated BI Board Bundle"
	if req.Plan != nil && strings.TrimSpace(req.Plan.BoardTitle) != "" {
		title = req.Plan.BoardTitle
	}
	return strings.Join([]string{
		"# " + title,
		"",
		"This archive was exported by GNyx UI as a portable BI board proof-of-concept bundle.",
		"",
		"## Contents",
		"- `dashboard-schema.json`: final SKW Dynamic UI-compatible renderer contract",
		"- `board-plan.json`: intermediate grounded board plan",
		"- `grounding-context.json`: metadata context used for grounding",
		"- `generation-metadata.json`: provider/model/runtime generation details",
		"",
		"## Generation Summary",
		"- Domain: " + strings.TrimSpace(req.Domain),
		"- Provider: " + strings.TrimSpace(req.Provider),
		"- Model: " + strings.TrimSpace(req.Model),
		"- Generation mode: " + strings.TrimSpace(req.GenerationMode),
		"- Fallback reason: " + strings.TrimSpace(req.FallbackReason),
		"",
		"## Prompt",
		strings.TrimSpace(req.Prompt),
		"",
		"## Notes",
		"- This bundle is intended for demonstration and proof-of-concept flows.",
		"- It does not require a live Sankhya runtime to be inspected or shared.",
	}, "\n") + "\n"
}

func firstProvider(items []string) string {
	if len(items) == 0 {
		return ""
	}
	return items[0]
}
