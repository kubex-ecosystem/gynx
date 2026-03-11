package bi

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var codeFencePattern = regexp.MustCompile("(?s)^```(?:json)?\\s*(.*?)\\s*```$")
var sqlAliasPattern = regexp.MustCompile(`(?i)\bAS\s+([A-Za-z_][A-Za-z0-9_]*)`)

const initialSystemPrompt = `You are the GNyx BI Board Planner.

Your job is to generate a grounded board plan for Sankhya ERP analytics.
You do not generate markdown.
You do not generate prose outside the JSON contract.
You must work only with the metadata provided in the request context.
Never invent tables, columns, joins, filters, enum values, business meanings, or widget types that are not supported by the provided metadata.

Your output is NOT the final renderer contract.
Your output is an intermediate board-plan JSON that the backend will validate and compile into the final SKW Dynamic UI DashboardSchema.

Rules:
1. Output valid JSON only.
2. Use only the allowed widget types: kpi, chart_bar, chart_donut, data_table.
3. Use only tables and columns present in the provided metadata context.
4. Prefer joins backed by physical foreign-key evidence when available.
5. Prefer one main business domain per request.
6. Prefer one main time column per request.
7. Keep the first board small, useful, and explainable.
8. Every widget must include grounding information.
9. Every uncertainty must go to assumptions or warnings.
10. SQL must be read-only, single-statement, and compatible with the target ERP SQL execution context.
11. If the request is under-specified, still produce the safest possible board plan using the strongest metadata evidence.
12. If a requested metric cannot be grounded safely, do not fake it. Emit a warning and choose a safer alternative if possible.
13. The required top-level fields are: board_title, board_description, business_goal, domain, time_reference, filters, joins, widgets, assumptions, warnings, grounding.
14. Every widget must contain: id, type, title, subtitle, metric_goal, size, data_source, display, grounding.
15. data_source must contain: main_table, sql, expected_granularity.
16. Never return a simplified legacy query object like {"query": {...}} instead of data_source.
17. If you are unsure about some optional field, keep it simple, but never omit required fields.

Return JSON only.`

func InitialSystemPrompt() string {
	return initialSystemPrompt
}

func BuildPlanningPrompt(userRequest string, maxWidgets int, grounding *GroundingContext) (string, error) {
	if maxWidgets <= 0 || maxWidgets > 6 {
		maxWidgets = 4
	}
	payload := map[string]any{
		"user_request":         userRequest,
		"target_domain":        grounding.Domain,
		"max_widgets":          maxWidgets,
		"allowed_widget_types": []WidgetType{WidgetTypeKPI, WidgetTypeChartBar, WidgetTypeChartDonut, WidgetTypeDataTable},
		"required_contract":    boardPlanTemplate(),
		"legacy_shapes_not_allowed": []string{
			"widgets[].query",
			"widgets[].description without widgets[].data_source",
		},
		"grounding_context": grounding,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func boardPlanTemplate() map[string]any {
	return map[string]any{
		"board_title":       "Sales Overview",
		"board_description": "Compact grounded commercial overview for the selected period.",
		"business_goal":     "Describe the business goal in one sentence.",
		"domain":            "sales",
		"time_reference": map[string]any{
			"table":  "TGFCAB",
			"column": "DTNEG",
			"reason": "Primary commercial time axis.",
		},
		"filters": []map[string]any{{
			"key":          "period",
			"label":        "Negotiation Period",
			"table":        "TGFCAB",
			"column":       "DTNEG",
			"filter_group": "periodo_data",
			"required":     true,
		}},
		"joins": []map[string]any{{
			"left_table":   "TGFCAB",
			"left_column":  "CODPARC",
			"right_table":  "TGFPAR",
			"right_column": "CODPARC",
			"join_type":    "INNER",
			"basis":        "physical_fk",
		}},
		"widgets": []map[string]any{{
			"id":          "kpi-total-sales",
			"type":        "kpi",
			"title":       "Total Sales",
			"subtitle":    "Selected period",
			"metric_goal": "Sum invoice value for sales movements",
			"size": map[string]any{
				"cols": 3,
				"rows": 1,
			},
			"data_source": map[string]any{
				"main_table":           "TGFCAB",
				"sql":                  "SELECT COALESCE(SUM(VLRNOTA), 0) AS value FROM TGFCAB WHERE DTNEG BETWEEN :P_PERIODO.INI AND :P_PERIODO.FIN AND CODEMP IN :P_CODEMP AND TIPMOV = '4'",
				"expected_granularity": "single_row_metric",
			},
			"display": map[string]any{
				"color": "green",
				"unit":  "R$",
			},
			"grounding": map[string]any{
				"tables_used":       []string{"TGFCAB"},
				"fields_used":       []string{"VLRNOTA", "DTNEG", "CODEMP", "TIPMOV"},
				"filter_columns":    []string{"DTNEG", "CODEMP"},
				"reasoning_summary": "Grounded on TGFCAB commercial totals.",
			},
		}},
		"assumptions": []string{"Use the safest grounded slice when the request is under-specified."},
		"warnings":    []string{},
		"grounding": []map[string]any{{
			"table":           "TGFCAB",
			"source":          "sankhya_catalog",
			"description":     "Commercial header table",
			"row_count":       0,
			"relevant_fields": []string{"DTNEG", "CODEMP", "CODPARC", "VLRNOTA", "TIPMOV"},
		}},
	}
}

func ParseBoardPlan(raw string) (*BoardPlan, error) {
	cleaned := sanitizeJSON(raw)
	var plan BoardPlan
	if err := json.Unmarshal([]byte(cleaned), &plan); err != nil {
		return nil, fmt.Errorf("failed to parse board plan JSON: %w", err)
	}
	return &plan, nil
}

type legacyBoardDraft struct {
	BoardTitle       string              `json:"board_title"`
	BoardDescription string              `json:"board_description"`
	BusinessGoal     string              `json:"business_goal"`
	Domain           string              `json:"domain"`
	Widgets          []legacyBoardWidget `json:"widgets"`
	Filters          []legacyBoardFilter `json:"filters"`
}

type legacyBoardWidget struct {
	Type        WidgetType          `json:"type"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Query       legacyBoardQuery    `json:"query"`
	Filters     []legacyBoardFilter `json:"filters"`
}

type legacyBoardQuery struct {
	Table       string              `json:"table"`
	Columns     []string            `json:"columns"`
	Aggregation string              `json:"aggregation"`
	GroupBy     []string            `json:"group_by"`
	OrderBy     []string            `json:"order_by"`
	Limit       int                 `json:"limit"`
	Joins       []GroundingJoin     `json:"joins"`
	Filters     []legacyBoardFilter `json:"filters"`
}

type legacyBoardFilter struct {
	Table       string `json:"table"`
	Column      string `json:"column"`
	Label       string `json:"label"`
	FilterGroup string `json:"filter_group"`
}

func RecoverBoardPlan(raw, userRequest, domain string, grounding *GroundingContext) (*BoardPlan, error) {
	cleaned := sanitizeJSON(raw)
	var draft legacyBoardDraft
	if err := json.Unmarshal([]byte(cleaned), &draft); err != nil {
		return nil, fmt.Errorf("failed to parse legacy board draft JSON: %w", err)
	}
	if len(draft.Widgets) == 0 {
		return nil, fmt.Errorf("legacy board draft has no widgets")
	}
	if strings.TrimSpace(domain) == "" {
		domain = strings.TrimSpace(draft.Domain)
	}
	if strings.TrimSpace(domain) == "" {
		domain = "sales"
	}
	if domain != "sales" {
		return nil, fmt.Errorf("legacy recovery currently supports only sales")
	}

	base := FallbackSalesOverviewPlan(userRequest, grounding, "recovered from legacy provider draft")
	canonical := map[string]BoardPlanWidget{}
	for _, widget := range base.Widgets {
		canonical[widget.ID] = widget
	}

	recovered := make([]BoardPlanWidget, 0, len(draft.Widgets))
	seen := map[string]struct{}{}
	for idx, widget := range draft.Widgets {
		candidate, ok := recoverSalesWidget(widget, idx, canonical)
		if !ok {
			base.Warnings = append(base.Warnings, fmt.Sprintf("Widget '%s' could not be recovered from the simplified provider draft and was skipped.", strings.TrimSpace(widget.Title)))
			continue
		}
		if _, exists := seen[candidate.ID]; exists {
			continue
		}
		seen[candidate.ID] = struct{}{}
		recovered = append(recovered, candidate)
	}
	if len(recovered) == 0 {
		return nil, fmt.Errorf("legacy board draft did not contain any recoverable widgets")
	}

	base.Widgets = recovered
	base.Domain = domain
	if value := strings.TrimSpace(draft.BoardTitle); value != "" {
		base.BoardTitle = value
	}
	if value := strings.TrimSpace(draft.BoardDescription); value != "" {
		base.BoardDescription = value
	}
	if value := strings.TrimSpace(draft.BusinessGoal); value != "" {
		base.BusinessGoal = value
	}
	base.Assumptions = append(base.Assumptions,
		"The provider returned a simplified draft using widgets[].query; the backend normalized it into the canonical board-plan contract.",
	)
	base.Grounding = buildGroundingRefs(grounding)
	return base, nil
}

func recoverSalesWidget(widget legacyBoardWidget, index int, canonical map[string]BoardPlanWidget) (BoardPlanWidget, bool) {
	title := strings.ToLower(strings.TrimSpace(widget.Title))
	agg := strings.ToUpper(strings.TrimSpace(widget.Query.Aggregation))
	hasColumn := func(target string) bool {
		for _, column := range widget.Query.Columns {
			if strings.EqualFold(strings.TrimSpace(column), target) {
				return true
			}
		}
		return false
	}
	hasJoinTable := func(target string) bool {
		for _, join := range widget.Query.Joins {
			if strings.EqualFold(join.LeftTable, target) || strings.EqualFold(join.RightTable, target) {
				return true
			}
		}
		return false
	}
	clone := func(id string) (BoardPlanWidget, bool) {
		item, ok := canonical[id]
		if !ok {
			return BoardPlanWidget{}, false
		}
		if value := strings.TrimSpace(widget.Title); value != "" {
			item.Title = value
		}
		if value := strings.TrimSpace(widget.Description); value != "" {
			item.MetricGoal = value
		}
		return item, true
	}

	switch {
	case widget.Type == WidgetTypeKPI && agg == "SUM" && hasColumn("VLRNOTA"):
		return clone("kpi-total-sales")
	case widget.Type == WidgetTypeKPI && (agg == "COUNT" || hasColumn("NUNOTA")):
		return clone("kpi-order-count")
	case widget.Type == WidgetTypeChartBar && hasColumn("DTNEG") && hasColumn("VLRNOTA"):
		return clone("chart-sales-by-month")
	case widget.Type == WidgetTypeDataTable && (hasColumn("CODPARC") || hasJoinTable("TGFPAR") || strings.Contains(title, "customer")):
		return clone("table-top-customers")
	default:
		_ = index
		return BoardPlanWidget{}, false
	}
}

func FallbackSalesOverviewPlan(userRequest string, grounding *GroundingContext, fallbackReason string) *BoardPlan {
	assumptions := []string{
		"DTNEG is used as the primary commercial time reference.",
		"TIPMOV = '4' is used as the default sales movement slice for this first template.",
		"ERP filter placeholders :P_PERIODO and :P_CODEMP are expected to be resolved by the target runtime.",
	}
	if strings.TrimSpace(fallbackReason) != "" {
		assumptions = append(assumptions, "Fallback activated because the provider response was not valid enough to compile directly: "+fallbackReason)
	}

	return &BoardPlan{
		BoardTitle:       "Sales Overview",
		BoardDescription: "Compact commercial overview for the selected period.",
		BusinessGoal:     strings.TrimSpace(userRequest),
		Domain:           "sales",
		TimeReference: BoardPlanTimeReference{
			Table:  "TGFCAB",
			Column: "DTNEG",
			Reason: "DTNEG is the strongest commercial timeline in the curated catalog slice.",
		},
		Filters: []BoardPlanFilter{
			{
				Key:         "period",
				Label:       "Negotiation Period",
				Table:       "TGFCAB",
				Column:      "DTNEG",
				FilterGroup: "periodo_data",
				Required:    true,
			},
			{
				Key:         "company",
				Label:       "Company",
				Table:       "TGFCAB",
				Column:      "CODEMP",
				FilterGroup: "empresa",
				Required:    true,
			},
		},
		Joins: []GroundingJoin{
			{LeftTable: "TGFCAB", LeftColumn: "CODPARC", RightTable: "TGFPAR", RightColumn: "CODPARC", JoinType: "INNER", Basis: "curated_catalog"},
			{LeftTable: "TGFITE", LeftColumn: "NUNOTA", RightTable: "TGFCAB", RightColumn: "NUNOTA", JoinType: "INNER", Basis: "curated_catalog"},
			{LeftTable: "TGFITE", LeftColumn: "CODPROD", RightTable: "TGFPRO", RightColumn: "CODPROD", JoinType: "INNER", Basis: "curated_catalog"},
		},
		Widgets: []BoardPlanWidget{
			{
				ID:         "kpi-total-sales",
				Type:       WidgetTypeKPI,
				Title:      "Total Sales",
				Subtitle:   "Selected period",
				MetricGoal: "Sum invoice value for commercial sales movements",
				Size:       BoardPlanWidgetSize{Cols: 3, Rows: 1},
				DataSource: BoardPlanDataSource{
					MainTable:           "TGFCAB",
					SQL:                 "SELECT COALESCE(SUM(VLRNOTA), 0) AS value FROM TGFCAB WHERE DTNEG BETWEEN :P_PERIODO.INI AND :P_PERIODO.FIN AND CODEMP IN :P_CODEMP AND TIPMOV = '4'",
					ExpectedGranularity: "single_row_metric",
				},
				Display: BoardPlanWidgetDisplay{Color: "green", Unit: "R$"},
				Grounding: BoardPlanWidgetGrounding{
					TablesUsed:       []string{"TGFCAB"},
					FieldsUsed:       []string{"VLRNOTA", "DTNEG", "CODEMP", "TIPMOV"},
					FilterColumns:    []string{"DTNEG", "CODEMP"},
					ReasoningSummary: "TGFCAB carries commercial header values and date scope for total sales.",
				},
			},
			{
				ID:         "kpi-order-count",
				Type:       WidgetTypeKPI,
				Title:      "Order Count",
				Subtitle:   "Selected period",
				MetricGoal: "Count sales notes in the selected period",
				Size:       BoardPlanWidgetSize{Cols: 3, Rows: 1},
				DataSource: BoardPlanDataSource{
					MainTable:           "TGFCAB",
					SQL:                 "SELECT COUNT(*) AS value FROM TGFCAB WHERE DTNEG BETWEEN :P_PERIODO.INI AND :P_PERIODO.FIN AND CODEMP IN :P_CODEMP AND TIPMOV = '4'",
					ExpectedGranularity: "single_row_metric",
				},
				Display: BoardPlanWidgetDisplay{Color: "blue", Unit: "un"},
				Grounding: BoardPlanWidgetGrounding{
					TablesUsed:       []string{"TGFCAB"},
					FieldsUsed:       []string{"NUNOTA", "DTNEG", "CODEMP", "TIPMOV"},
					FilterColumns:    []string{"DTNEG", "CODEMP"},
					ReasoningSummary: "TGFCAB is the safest source for counting commercial documents in this first slice.",
				},
			},
			{
				ID:         "chart-sales-by-month",
				Type:       WidgetTypeChartBar,
				Title:      "Sales by Month",
				Subtitle:   "Selected period",
				MetricGoal: "Show sales evolution by month",
				Size:       BoardPlanWidgetSize{Cols: 6, Rows: 2},
				DataSource: BoardPlanDataSource{
					MainTable:           "TGFCAB",
					SQL:                 "SELECT FORMAT(DTNEG, 'MMM/yy') AS label, SUM(VLRNOTA) AS value FROM TGFCAB WHERE DTNEG BETWEEN :P_PERIODO.INI AND :P_PERIODO.FIN AND CODEMP IN :P_CODEMP AND TIPMOV = '4' GROUP BY FORMAT(DTNEG, 'MMM/yy') ORDER BY MIN(DTNEG)",
					ExpectedGranularity: "time_series_month",
				},
				Display: BoardPlanWidgetDisplay{Color: "purple", Unit: "R$"},
				Grounding: BoardPlanWidgetGrounding{
					TablesUsed:       []string{"TGFCAB"},
					FieldsUsed:       []string{"DTNEG", "VLRNOTA", "CODEMP", "TIPMOV"},
					FilterColumns:    []string{"DTNEG", "CODEMP"},
					ReasoningSummary: "TGFCAB exposes commercial date and value fields needed for a monthly sales trend.",
				},
			},
			{
				ID:         "table-top-customers",
				Type:       WidgetTypeDataTable,
				Title:      "Top Customers",
				Subtitle:   "Selected period",
				MetricGoal: "Show customers with the highest commercial value",
				Size:       BoardPlanWidgetSize{Cols: 6, Rows: 2},
				DataSource: BoardPlanDataSource{
					MainTable:           "TGFCAB",
					SQL:                 "SELECT TOP 10 PAR.NOMEPARC AS customer_name, SUM(CAB.VLRNOTA) AS total_sales, COUNT(*) AS order_count FROM TGFCAB CAB INNER JOIN TGFPAR PAR ON CAB.CODPARC = PAR.CODPARC WHERE CAB.DTNEG BETWEEN :P_PERIODO.INI AND :P_PERIODO.FIN AND CAB.CODEMP IN :P_CODEMP AND CAB.TIPMOV = '4' GROUP BY PAR.NOMEPARC ORDER BY total_sales DESC",
					ExpectedGranularity: "customer_ranking",
				},
				Display: BoardPlanWidgetDisplay{
					Columns: []DashboardColumn{
						{Key: "customer_name", Header: "Customer", Sortable: true},
						{Key: "total_sales", Header: "Total Sales", Align: "right", Sortable: true},
						{Key: "order_count", Header: "Orders", Align: "center", Sortable: true},
					},
				},
				Grounding: BoardPlanWidgetGrounding{
					TablesUsed:       []string{"TGFCAB", "TGFPAR"},
					FieldsUsed:       []string{"CODPARC", "NOMEPARC", "VLRNOTA", "DTNEG", "CODEMP", "TIPMOV"},
					FilterColumns:    []string{"DTNEG", "CODEMP"},
					ReasoningSummary: "TGFCAB joined with TGFPAR is the safest first ranking slice for customer performance.",
				},
			},
		},
		Assumptions: assumptions,
		Warnings: []string{
			"This first compiled board is constrained to the curated sales slice and does not yet generalize across all Sankhya domains.",
		},
		Grounding: buildGroundingRefs(grounding),
	}
}

func sanitizeJSON(raw string) string {
	raw = strings.TrimSpace(raw)
	if matches := codeFencePattern.FindStringSubmatch(raw); len(matches) == 2 {
		raw = matches[1]
	}
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start >= 0 && end > start {
		raw = raw[start : end+1]
	}
	return strings.TrimSpace(raw)
}

func buildGroundingRefs(grounding *GroundingContext) []BoardPlanGroundingRef {
	if grounding == nil {
		return nil
	}
	result := make([]BoardPlanGroundingRef, 0, len(grounding.PrimaryTables))
	for _, table := range grounding.PrimaryTables {
		fields := make([]string, 0, len(table.RelevantFields))
		for _, field := range table.RelevantFields {
			fields = append(fields, field.Column)
		}
		result = append(result, BoardPlanGroundingRef{
			Table:          table.Table,
			Source:         "sankhya_catalog",
			Description:    table.Description,
			RowCount:       table.RowCount,
			RelevantFields: fields,
		})
	}
	return result
}

func ValidateBoardPlan(plan *BoardPlan, grounding *GroundingContext) error {
	if plan == nil {
		return fmt.Errorf("board plan is required")
	}
	if strings.TrimSpace(plan.Domain) == "" {
		return fmt.Errorf("domain is required")
	}
	if len(plan.Widgets) == 0 {
		return fmt.Errorf("at least one widget is required")
	}
	if len(plan.Widgets) > 6 {
		return fmt.Errorf("too many widgets: %d", len(plan.Widgets))
	}

	allowedTables := map[string]map[string]struct{}{}
	for _, table := range grounding.PrimaryTables {
		fields := map[string]struct{}{}
		for _, field := range table.RelevantFields {
			fields[field.Column] = struct{}{}
		}
		for _, column := range table.DateColumns {
			fields[column] = struct{}{}
		}
		allowedTables[table.Table] = fields
	}

	seenIDs := map[string]struct{}{}
	for _, widget := range plan.Widgets {
		if _, ok := seenIDs[widget.ID]; ok {
			return fmt.Errorf("duplicated widget id: %s", widget.ID)
		}
		seenIDs[widget.ID] = struct{}{}
		if !isAllowedWidgetType(widget.Type) {
			return fmt.Errorf("unsupported widget type: %s", widget.Type)
		}
		if err := validateGroundedFields(widget.Grounding.TablesUsed, widget.Grounding.FieldsUsed, allowedTables); err != nil {
			return fmt.Errorf("widget %s grounding is invalid: %w", widget.ID, err)
		}
		if err := validateSQL(widget.DataSource.SQL); err != nil {
			return fmt.Errorf("widget %s sql is invalid: %w", widget.ID, err)
		}
	}
	return nil
}

func CompileDashboardSchema(plan *BoardPlan) (*DashboardSchema, error) {
	if plan == nil {
		return nil, fmt.Errorf("board plan is required")
	}
	widgets := make([]DashboardWidget, 0, len(plan.Widgets))
	for _, widget := range plan.Widgets {
		config := DashboardWidgetConfig{
			Title:    widget.Title,
			Subtitle: widget.Subtitle,
			SQLQuery: strings.TrimSpace(widget.DataSource.SQL),
			Unit:     widget.Display.Unit,
		}
		switch widget.Type {
		case WidgetTypeKPI:
			config.Color = normalizedColor(widget.Display.Color)
			config.Compact = widget.Size.Cols <= 2
		case WidgetTypeChartBar:
			config.Orientation = "vertical"
			config.BarColor = chartColor(widget.Display.Color)
			config.ShowValues = true
			config.ShowGrid = true
		case WidgetTypeChartDonut:
			// no extra config required for first slice
		case WidgetTypeDataTable:
			config.PageSize = 10
			config.Striped = true
			config.Columns = widget.Display.Columns
			if len(config.Columns) == 0 {
				config.Columns = inferColumnsFromSQL(widget.DataSource.SQL)
			}
		}
		widgets = append(widgets, DashboardWidget{
			ID:     widget.ID,
			Type:   widget.Type,
			Size:   DashboardWidgetSize{Cols: clamp(widget.Size.Cols, 3, 12), Rows: clamp(widget.Size.Rows, 1, 4)},
			Config: config,
		})
	}
	return &DashboardSchema{
		DashboardID: slugify(plan.BoardTitle) + "-" + time.Now().UTC().Format("20060102150405"),
		Title:       plan.BoardTitle,
		Description: plan.BoardDescription,
		Widgets:     widgets,
	}, nil
}

func normalizedColor(color string) string {
	switch strings.ToLower(strings.TrimSpace(color)) {
	case "green", "blue", "red", "yellow", "purple", "neutral":
		return strings.ToLower(strings.TrimSpace(color))
	default:
		return "blue"
	}
}

func chartColor(color string) string {
	switch normalizedColor(color) {
	case "green":
		return "#22c55e"
	case "red":
		return "#ef4444"
	case "yellow":
		return "#f59e0b"
	case "purple":
		return "#a855f7"
	case "neutral":
		return "#64748b"
	default:
		return "#2563eb"
	}
}

func inferColumnsFromSQL(sql string) []DashboardColumn {
	matches := sqlAliasPattern.FindAllStringSubmatch(sql, -1)
	columns := make([]DashboardColumn, 0, len(matches))
	seen := map[string]struct{}{}
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		alias := match[1]
		if _, ok := seen[alias]; ok {
			continue
		}
		seen[alias] = struct{}{}
		columns = append(columns, DashboardColumn{Key: alias, Header: alias, Sortable: true})
	}
	return columns
}

func isAllowedWidgetType(widgetType WidgetType) bool {
	switch widgetType {
	case WidgetTypeKPI, WidgetTypeChartBar, WidgetTypeChartDonut, WidgetTypeDataTable:
		return true
	default:
		return false
	}
}

func validateGroundedFields(tables []string, fields []string, allowedTables map[string]map[string]struct{}) error {
	for _, table := range tables {
		if _, ok := allowedTables[table]; !ok {
			return fmt.Errorf("table %s is outside grounded context", table)
		}
	}
	for _, field := range fields {
		found := false
		for _, tableFields := range allowedTables {
			if _, ok := tableFields[field]; ok {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("field %s is outside grounded context", field)
		}
	}
	return nil
}

func validateSQL(sql string) error {
	trimmed := strings.TrimSpace(sql)
	if trimmed == "" {
		return fmt.Errorf("sql is required")
	}
	lower := strings.ToLower(trimmed)
	for _, forbidden := range []string{" insert ", " update ", " delete ", " drop ", " alter ", " truncate ", " create "} {
		if strings.Contains(" "+lower+" ", forbidden) {
			return fmt.Errorf("sql must be read-only")
		}
	}
	if strings.Count(trimmed, ";") > 1 || (strings.Contains(trimmed, ";") && !strings.HasSuffix(trimmed, ";")) {
		return fmt.Errorf("sql must be single-statement")
	}
	return nil
}

func clamp(value int, min int, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer(" ", "-", "_", "-", "/", "-", "\\", "-", ":", "-", ".", "-")
	value = replacer.Replace(value)
	value = strings.Trim(value, "-")
	if value == "" {
		return "board"
	}
	return value
}
