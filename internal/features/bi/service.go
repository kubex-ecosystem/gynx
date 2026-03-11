package bi

import (
	"context"
	"sort"
	"strings"

	"github.com/kubex-ecosystem/gnyx/internal/dsclient"
	"github.com/kubex-ecosystem/gnyx/internal/dsclient/datastore"
)

const DefaultCatalogSchema = "sankhya_catalog"

type Service struct {
	catalogSchema string
}

func NewService(catalogSchema string) *Service {
	catalogSchema = strings.TrimSpace(catalogSchema)
	if catalogSchema == "" {
		catalogSchema = DefaultCatalogSchema
	}
	return &Service{catalogSchema: catalogSchema}
}

func (s *Service) BuildSalesCommercialContext(ctx context.Context) (*GroundingContext, error) {
	conn, err := datastore.Connection(ctx)
	if err != nil {
		return nil, err
	}
	pgExec, err := dsclient.GetPGExecutor(ctx, conn)
	if err != nil {
		return nil, err
	}

	tables, err := s.loadSalesTables(ctx, pgExec)
	if err != nil {
		return nil, err
	}
	joins, err := s.loadSalesJoins(ctx, pgExec)
	if err != nil {
		return nil, err
	}
	filters, err := s.loadSalesFilters(ctx, pgExec)
	if err != nil {
		return nil, err
	}
	options, err := s.loadSalesDomainOptions(ctx, pgExec)
	if err != nil {
		return nil, err
	}

	return &GroundingContext{
		Domain:             "sales",
		CatalogSchema:      s.catalogSchema,
		PrimaryTables:      tables,
		Joins:              joins,
		FilterCandidates:   filters,
		DomainOptions:      options,
		RecommendedWidgets: []WidgetType{WidgetTypeKPI, WidgetTypeChartBar, WidgetTypeChartDonut, WidgetTypeDataTable},
		Warnings: []string{
			"Prefer TGFCAB as the main commercial header table for the first slice.",
			"Prefer DTNEG as the primary time reference unless the user request clearly requires another date column.",
			"Prefer TIPMOV and STATUSNOTA only when they are grounded by the provided domain options.",
		},
	}, nil
}

func (s *Service) loadSalesTables(ctx context.Context, pgExec dsclient.PGExecutor) ([]GroundingTable, error) {
	coreOrder := []string{"TGFCAB", "TGFITE", "TGFPAR", "TGFPRO"}
	relevantFields := map[string][]string{
		"TGFCAB": {"NUNOTA", "CODPARC", "DTNEG", "VLRNOTA", "CODVEND", "CODEMP", "TIPMOV", "STATUSNOTA"},
		"TGFITE": {"NUNOTA", "SEQUENCIA", "CODPROD", "VLRUNIT", "QTDNEG"},
		"TGFPAR": {"CODPARC", "NOMEPARC", "CLIENTE", "FORNECEDOR", "CODVEND"},
		"TGFPRO": {"CODPROD", "DESCRPROD", "CODGRUPOPROD", "TIPO"},
	}

	const tablesQuery = `
		SELECT table_name, table_description, COALESCE(physical_schema, ''), COALESCE(NULLIF(row_count::text, '')::bigint, 0)
		FROM sankhya_catalog.tdd_tabelas
		WHERE table_name = ANY($1)
	`
	rows, err := pgExec.Query(ctx, strings.ReplaceAll(tablesQuery, "sankhya_catalog", s.catalogSchema), coreOrder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tablesByName := map[string]GroundingTable{}
	for rows.Next() {
		var table GroundingTable
		if err := rows.Scan(&table.Table, &table.Description, &table.PhysicalSchema, &table.RowCount); err != nil {
			return nil, err
		}
		tablesByName[table.Table] = table
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	const fieldsQuery = `
		SELECT table_name, column_name, COALESCE(field_description, ''), COALESCE(field_type, ''), COALESCE(NULLIF(option_count::text, '')::integer, 0)
		FROM sankhya_catalog.tdd_campos
		WHERE table_name = ANY($1)
		ORDER BY table_name, column_name
	`
	fieldRows, err := pgExec.Query(ctx, strings.ReplaceAll(fieldsQuery, "sankhya_catalog", s.catalogSchema), coreOrder)
	if err != nil {
		return nil, err
	}
	defer fieldRows.Close()

	for fieldRows.Next() {
		var tableName string
		var field GroundingField
		if err := fieldRows.Scan(&tableName, &field.Column, &field.Description, &field.FieldType, &field.OptionCount); err != nil {
			return nil, err
		}
		if !containsString(relevantFields[tableName], field.Column) {
			continue
		}
		table := tablesByName[tableName]
		table.RelevantFields = append(table.RelevantFields, field)
		tablesByName[tableName] = table
	}
	if err := fieldRows.Err(); err != nil {
		return nil, err
	}

	const datesQuery = `
		SELECT table_name, column_name
		FROM sankhya_catalog.colunas_data
		WHERE table_name = ANY($1)
		ORDER BY table_name, column_name
	`
	dateRows, err := pgExec.Query(ctx, strings.ReplaceAll(datesQuery, "sankhya_catalog", s.catalogSchema), coreOrder)
	if err != nil {
		return nil, err
	}
	defer dateRows.Close()

	for dateRows.Next() {
		var tableName, columnName string
		if err := dateRows.Scan(&tableName, &columnName); err != nil {
			return nil, err
		}
		table := tablesByName[tableName]
		table.DateColumns = append(table.DateColumns, columnName)
		tablesByName[tableName] = table
	}
	if err := dateRows.Err(); err != nil {
		return nil, err
	}

	result := make([]GroundingTable, 0, len(coreOrder))
	for _, tableName := range coreOrder {
		table, ok := tablesByName[tableName]
		if !ok {
			continue
		}
		sort.Slice(table.RelevantFields, func(i, j int) bool { return table.RelevantFields[i].Column < table.RelevantFields[j].Column })
		sort.Strings(table.DateColumns)
		result = append(result, table)
	}
	return result, nil
}

func (s *Service) loadSalesJoins(ctx context.Context, pgExec dsclient.PGExecutor) ([]GroundingJoin, error) {
	coreTables := []string{"TGFCAB", "TGFITE", "TGFPAR", "TGFPRO"}
	const physicalQuery = `
		SELECT parent_table, parent_columns, referenced_table, referenced_columns
		FROM sankhya_catalog.relacoes_fisicas_fk
		WHERE parent_table = ANY($1) AND referenced_table = ANY($1)
	`
	rows, err := pgExec.Query(ctx, strings.ReplaceAll(physicalQuery, "sankhya_catalog", s.catalogSchema), coreTables)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	joins := map[string]GroundingJoin{}
	for rows.Next() {
		var parentTable, parentColumns, referencedTable, referencedColumns string
		if err := rows.Scan(&parentTable, &parentColumns, &referencedTable, &referencedColumns); err != nil {
			return nil, err
		}
		leftColumn := firstCSVToken(parentColumns)
		rightColumn := firstCSVToken(referencedColumns)
		if leftColumn == "" || rightColumn == "" {
			continue
		}
		join := GroundingJoin{
			LeftTable:   parentTable,
			LeftColumn:  leftColumn,
			RightTable:  referencedTable,
			RightColumn: rightColumn,
			JoinType:    "INNER",
			Basis:       "physical_fk",
		}
		joins[joinKey(join)] = join
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, fallback := range []GroundingJoin{
		{LeftTable: "TGFCAB", LeftColumn: "CODPARC", RightTable: "TGFPAR", RightColumn: "CODPARC", JoinType: "INNER", Basis: "curated_catalog"},
		{LeftTable: "TGFITE", LeftColumn: "NUNOTA", RightTable: "TGFCAB", RightColumn: "NUNOTA", JoinType: "INNER", Basis: "curated_catalog"},
		{LeftTable: "TGFITE", LeftColumn: "CODPROD", RightTable: "TGFPRO", RightColumn: "CODPROD", JoinType: "INNER", Basis: "curated_catalog"},
	} {
		if _, ok := joins[joinKey(fallback)]; !ok {
			joins[joinKey(fallback)] = fallback
		}
	}

	result := make([]GroundingJoin, 0, len(joins))
	for _, join := range joins {
		result = append(result, join)
	}
	sort.Slice(result, func(i, j int) bool { return joinKey(result[i]) < joinKey(result[j]) })
	return result, nil
}

func (s *Service) loadSalesFilters(ctx context.Context, pgExec dsclient.PGExecutor) ([]GroundingFilter, error) {
	coreTables := []string{"TGFCAB", "TGFITE", "TGFPAR", "TGFPRO"}
	const q = `
		SELECT table_name, column_name, COALESCE(field_description, ''), COALESCE(field_type, ''), COALESCE(filter_group, '')
		FROM sankhya_catalog.campos_filtro_candidatos
		WHERE table_name = ANY($1)
		ORDER BY table_name, filter_group, column_name
	`
	rows, err := pgExec.Query(ctx, strings.ReplaceAll(q, "sankhya_catalog", s.catalogSchema), coreTables)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []GroundingFilter
	for rows.Next() {
		var item GroundingFilter
		if err := rows.Scan(&item.Table, &item.Column, &item.Label, &item.FieldType, &item.FilterGroup); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) loadSalesDomainOptions(ctx context.Context, pgExec dsclient.PGExecutor) ([]GroundingOption, error) {
	const q = `
		SELECT table_name, column_name, COALESCE(field_description, ''), option_value, option_label, COALESCE(is_default, '')
		FROM sankhya_catalog.dominios_filtro_opcoes
		WHERE table_name = 'TGFCAB' AND column_name IN ('TIPMOV', 'STATUSNOTA')
		ORDER BY table_name, column_name, option_order
	`
	rows, err := pgExec.Query(ctx, strings.ReplaceAll(q, "sankhya_catalog", s.catalogSchema))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []GroundingOption
	for rows.Next() {
		var item GroundingOption
		var isDefault string
		if err := rows.Scan(&item.Table, &item.Column, &item.Label, &item.Value, &item.OptionLabel, &isDefault); err != nil {
			return nil, err
		}
		item.IsDefault = strings.EqualFold(isDefault, "S")
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func firstCSVToken(value string) string {
	parts := strings.Split(value, ",")
	if len(parts) == 0 {
		return strings.TrimSpace(value)
	}
	return strings.TrimSpace(parts[0])
}

func joinKey(join GroundingJoin) string {
	return strings.Join([]string{join.LeftTable, join.LeftColumn, join.RightTable, join.RightColumn}, ":")
}
