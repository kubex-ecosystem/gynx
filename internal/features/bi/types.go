package bi

type WidgetType string

const (
	WidgetTypeKPI        WidgetType = "kpi"
	WidgetTypeChartBar   WidgetType = "chart_bar"
	WidgetTypeChartDonut WidgetType = "chart_donut"
	WidgetTypeDataTable  WidgetType = "data_table"
)

type GroundingContext struct {
	Domain             string            `json:"domain"`
	CatalogSchema      string            `json:"catalog_schema"`
	PrimaryTables      []GroundingTable  `json:"primary_tables"`
	Joins              []GroundingJoin   `json:"joins"`
	FilterCandidates   []GroundingFilter `json:"filter_candidates"`
	DomainOptions      []GroundingOption `json:"domain_options"`
	RecommendedWidgets []WidgetType      `json:"recommended_widgets"`
	Warnings           []string          `json:"warnings,omitempty"`
}

type GroundingTable struct {
	Table          string           `json:"table"`
	Description    string           `json:"description"`
	PhysicalSchema string           `json:"physical_schema,omitempty"`
	RowCount       int64            `json:"row_count"`
	RelevantFields []GroundingField `json:"relevant_fields"`
	DateColumns    []string         `json:"date_columns,omitempty"`
}

type GroundingField struct {
	Column      string `json:"column"`
	Description string `json:"description"`
	FieldType   string `json:"field_type,omitempty"`
	OptionCount int    `json:"option_count,omitempty"`
}

type GroundingJoin struct {
	LeftTable   string `json:"left_table"`
	LeftColumn  string `json:"left_column"`
	RightTable  string `json:"right_table"`
	RightColumn string `json:"right_column"`
	JoinType    string `json:"join_type"`
	Basis       string `json:"basis"`
}

type GroundingFilter struct {
	Table       string `json:"table"`
	Column      string `json:"column"`
	Label       string `json:"label"`
	FieldType   string `json:"field_type,omitempty"`
	FilterGroup string `json:"filter_group,omitempty"`
}

type GroundingOption struct {
	Table       string `json:"table"`
	Column      string `json:"column"`
	Label       string `json:"label,omitempty"`
	Value       string `json:"value"`
	OptionLabel string `json:"option_label"`
	IsDefault   bool   `json:"is_default,omitempty"`
}

type BoardPlan struct {
	BoardTitle       string                  `json:"board_title"`
	BoardDescription string                  `json:"board_description"`
	BusinessGoal     string                  `json:"business_goal"`
	Domain           string                  `json:"domain"`
	TimeReference    BoardPlanTimeReference  `json:"time_reference"`
	Filters          []BoardPlanFilter       `json:"filters"`
	Joins            []GroundingJoin         `json:"joins"`
	Widgets          []BoardPlanWidget       `json:"widgets"`
	Assumptions      []string                `json:"assumptions"`
	Warnings         []string                `json:"warnings"`
	Grounding        []BoardPlanGroundingRef `json:"grounding"`
}

type BoardPlanTimeReference struct {
	Table  string `json:"table"`
	Column string `json:"column"`
	Reason string `json:"reason"`
}

type BoardPlanFilter struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Table       string `json:"table"`
	Column      string `json:"column"`
	FilterGroup string `json:"filter_group"`
	Required    bool   `json:"required"`
}

type BoardPlanWidget struct {
	ID          string                   `json:"id"`
	Type        WidgetType               `json:"type"`
	Title       string                   `json:"title"`
	Subtitle    string                   `json:"subtitle"`
	MetricGoal  string                   `json:"metric_goal"`
	Size        BoardPlanWidgetSize      `json:"size"`
	DataSource  BoardPlanDataSource      `json:"data_source"`
	Display     BoardPlanWidgetDisplay   `json:"display"`
	Grounding   BoardPlanWidgetGrounding `json:"grounding"`
	Assumptions []string                 `json:"assumptions"`
	Warnings    []string                 `json:"warnings"`
}

type BoardPlanWidgetSize struct {
	Cols int `json:"cols"`
	Rows int `json:"rows"`
}

type BoardPlanDataSource struct {
	MainTable           string `json:"main_table"`
	SQL                 string `json:"sql"`
	ExpectedGranularity string `json:"expected_granularity"`
}

type BoardPlanWidgetDisplay struct {
	Color   string            `json:"color,omitempty"`
	Unit    string            `json:"unit,omitempty"`
	Columns []DashboardColumn `json:"columns,omitempty"`
}

type BoardPlanWidgetGrounding struct {
	TablesUsed       []string `json:"tables_used"`
	FieldsUsed       []string `json:"fields_used"`
	FilterColumns    []string `json:"filter_columns"`
	ReasoningSummary string   `json:"reasoning_summary"`
}

type BoardPlanGroundingRef struct {
	Table          string   `json:"table"`
	Source         string   `json:"source"`
	Description    string   `json:"description"`
	RowCount       int64    `json:"row_count"`
	RelevantFields []string `json:"relevant_fields"`
}

type DashboardSchema struct {
	DashboardID string            `json:"dashboardId"`
	Title       string            `json:"title"`
	Description string            `json:"description,omitempty"`
	Widgets     []DashboardWidget `json:"widgets"`
}

type DashboardWidget struct {
	ID     string                `json:"id"`
	Type   WidgetType            `json:"type"`
	Size   DashboardWidgetSize   `json:"size"`
	Config DashboardWidgetConfig `json:"config"`
}

type DashboardWidgetSize struct {
	Cols int `json:"cols"`
	Rows int `json:"rows"`
}

type DashboardWidgetConfig struct {
	Title           string            `json:"title,omitempty"`
	Subtitle        string            `json:"subtitle,omitempty"`
	Color           string            `json:"color,omitempty"`
	SQLQuery        string            `json:"sqlQuery,omitempty"`
	Unit            string            `json:"unit,omitempty"`
	Compact         bool              `json:"compact,omitempty"`
	Orientation     string            `json:"orientation,omitempty"`
	BarColor        string            `json:"barColor,omitempty"`
	ShowValues      bool              `json:"showValues,omitempty"`
	ShowGrid        bool              `json:"showGrid,omitempty"`
	Columns         []DashboardColumn `json:"columns,omitempty"`
	PageSize        int               `json:"pageSize,omitempty"`
	Striped         bool              `json:"striped,omitempty"`
	RefreshInterval int               `json:"refreshInterval,omitempty"`
}

type DashboardColumn struct {
	Key      string `json:"key"`
	Header   string `json:"header"`
	Align    string `json:"align,omitempty"`
	Sortable bool   `json:"sortable,omitempty"`
}
