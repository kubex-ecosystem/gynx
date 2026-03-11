package bi

import "testing"

func TestRecoverBoardPlanFromLegacyDraft(t *testing.T) {
	grounding := &GroundingContext{
		Domain: "sales",
		PrimaryTables: []GroundingTable{
			{Table: "TGFCAB", RelevantFields: []GroundingField{{Column: "VLRNOTA"}, {Column: "NUNOTA"}, {Column: "DTNEG"}, {Column: "CODEMP"}, {Column: "CODPARC"}, {Column: "TIPMOV"}}, DateColumns: []string{"DTNEG"}},
			{Table: "TGFPAR", RelevantFields: []GroundingField{{Column: "CODPARC"}, {Column: "NOMEPARC"}}},
			{Table: "TGFITE", RelevantFields: []GroundingField{{Column: "NUNOTA"}, {Column: "CODPROD"}}},
			{Table: "TGFPRO", RelevantFields: []GroundingField{{Column: "CODPROD"}}},
		},
		Joins: []GroundingJoin{{LeftTable: "TGFCAB", LeftColumn: "CODPARC", RightTable: "TGFPAR", RightColumn: "CODPARC", JoinType: "INNER", Basis: "curated_catalog"}},
	}

	raw := "```json\n{\n  \"widgets\": [\n    {\n      \"type\": \"kpi\",\n      \"title\": \"Total Sales\",\n      \"description\": \"Total sales amount\",\n      \"query\": {\n        \"table\": \"TGFCAB\",\n        \"columns\": [\"VLRNOTA\"],\n        \"aggregation\": \"SUM\",\n        \"joins\": [],\n        \"filters\": []\n      }\n    },\n    {\n      \"type\": \"kpi\",\n      \"title\": \"Order Count\",\n      \"description\": \"Number of orders\",\n      \"query\": {\n        \"table\": \"TGFCAB\",\n        \"columns\": [\"NUNOTA\"],\n        \"aggregation\": \"COUNT\",\n        \"joins\": [],\n        \"filters\": []\n      }\n    },\n    {\n      \"type\": \"chart_bar\",\n      \"title\": \"Sales by Month\",\n      \"description\": \"Sales amount by month\",\n      \"query\": {\n        \"table\": \"TGFCAB\",\n        \"columns\": [\"DTNEG\", \"VLRNOTA\"],\n        \"aggregation\": \"SUM\",\n        \"group_by\": [\"EXTRACT(MONTH FROM DTNEG)\"],\n        \"joins\": [],\n        \"filters\": []\n      }\n    },\n    {\n      \"type\": \"data_table\",\n      \"title\": \"Top Customers\",\n      \"description\": \"Top customers by sales amount\",\n      \"query\": {\n        \"table\": \"TGFCAB\",\n        \"columns\": [\"CODPARC\", \"VLRNOTA\"],\n        \"aggregation\": \"SUM\",\n        \"group_by\": [\"CODPARC\"],\n        \"order_by\": [\"VLRNOTA DESC\"],\n        \"limit\": 10,\n        \"joins\": [\n          {\n            \"left_table\": \"TGFCAB\",\n            \"left_column\": \"CODPARC\",\n            \"right_table\": \"TGFPAR\",\n            \"right_column\": \"CODPARC\",\n            \"join_type\": \"INNER\"\n          }\n        ],\n        \"filters\": []\n      }\n    }\n  ],\n  \"filters\": [\n    {\n      \"table\": \"TGFCAB\",\n      \"column\": \"DTNEG\",\n      \"label\": \"Date\",\n      \"filter_group\": \"periodo_data\"\n    }\n  ]\n}\n```"

	plan, err := RecoverBoardPlan(raw, "Create a compact sales overview dashboard.", "sales", grounding)
	if err != nil {
		t.Fatalf("RecoverBoardPlan returned error: %v", err)
	}
	if len(plan.Widgets) != 4 {
		t.Fatalf("expected 4 widgets, got %d", len(plan.Widgets))
	}
	if plan.Domain != "sales" {
		t.Fatalf("expected domain sales, got %q", plan.Domain)
	}
	if err := ValidateBoardPlan(plan, grounding); err != nil {
		t.Fatalf("ValidateBoardPlan returned error: %v", err)
	}
}
