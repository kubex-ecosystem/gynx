package cli

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/joho/godotenv"
	getlextr "github.com/kubex-ecosystem/getl/extr"
	getlsql "github.com/kubex-ecosystem/getl/sql"
	getlutils "github.com/kubex-ecosystem/getl/utils"
	gcfg "github.com/kubex-ecosystem/gnyx/internal/config"
	dsclient "github.com/kubex-ecosystem/gnyx/internal/dsclient"
	dsstore "github.com/kubex-ecosystem/gnyx/internal/dsclient/datastore"
	kbxMod "github.com/kubex-ecosystem/gnyx/internal/module/kbx"
	gl "github.com/kubex-ecosystem/logz"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
)

const (
	defaultSankhyaCatalogDir     = "/ALL/KUBEX/JOBS/skw-dynamic-ui/docs/references/data/catalogo_bi"
	defaultSankhyaConfigDir      = "./config/getl/sankhya_catalog"
	defaultSankhyaSyncManifest   = "./config/getl/sankhya_catalog/sync.manifest.json"
	defaultSankhyaCatalogSchema  = "sankhya_catalog"
	defaultSankhyaCatalogDomain  = "bi_catalog"
	defaultSankhyaCatalogRefresh = "full_refresh"
)

var sqlIdentifierPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

type sankhyaSyncOptions struct {
	envFile          string
	catalogDir       string
	configDir        string
	syncManifestPath string
	pgDSN            string
	schema           string
	refreshMode      string
	datasets         []string
	debug            bool
}

type sankhyaSyncManifest struct {
	Version        string                       `json:"version"`
	SourceSystem   string                       `json:"source_system"`
	Domain         string                       `json:"domain"`
	Schema         string                       `json:"schema"`
	SourceManifest string                       `json:"source_manifest"`
	Datasets       []sankhyaSyncManifestDataset `json:"datasets"`
}

type sankhyaSyncManifestDataset struct {
	Name   string `json:"name"`
	Config string `json:"config"`
	Domain string `json:"domain,omitempty"`
}

type sankhyaCatalogManifest struct {
	ExtractedAtUTC string                          `json:"extracted_at_utc"`
	Server         string                          `json:"server"`
	Database       string                          `json:"database"`
	Datasets       []sankhyaCatalogManifestDataset `json:"datasets"`
}

type sankhyaCatalogManifestDataset struct {
	Dataset string `json:"dataset"`
	Rows    int64  `json:"rows"`
	File    string `json:"file"`
}

type sankhyaSyncResult struct {
	Dataset    string
	TableName  string
	RowCount   int64
	Status     string
	ConfigPath string
	Error      error
}

func MetadataCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "metadata",
		Short: "Metadata and catalog operations",
		Long:  "Metadata commands for loading and managing external analytical catalogs used by GNyx.",
	}

	rootCmd.AddCommand(sankhyaMetadataCommand())
	return rootCmd
}

func sankhyaMetadataCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sankhya",
		Short: "Sankhya BI catalog operations",
		Long:  "Commands for loading and registering Sankhya BI metadata catalogs into the Domus PostgreSQL substrate.",
	}

	cmd.AddCommand(syncSankhyaCatalogCommand())
	return cmd
}

func syncSankhyaCatalogCommand() *cobra.Command {
	opts := &sankhyaSyncOptions{}

	cmd := &cobra.Command{
		Use:     "sync",
		Aliases: []string{"sync-catalog", "load-catalog"},
		Short:   "Load Sankhya catalog CSVs into PostgreSQL and register them in Domus",
		Example: ConcatenateExamples([]string{
			"gnyx metadata sankhya sync --pg-dsn 'postgres://kubex_adm:admin123@localhost:5432/postgres?sslmode=disable'",
			"gnyx metadata sankhya sync --datasets tdd_tabelas,tdd_campos,tdd_ligacoes --schema sankhya_catalog",
		}),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSankhyaCatalogSync(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.debug, "debug", "D", false, "Enable debug logging")
	cmd.Flags().StringVarP(&opts.envFile, "env-file", "e", os.ExpandEnv(kbxMod.DefaultEnvPath), "Path to the env file used by GNyx runtime")
	cmd.Flags().StringVar(&opts.catalogDir, "catalog-dir", defaultSankhyaCatalogDir, "Directory containing Sankhya catalog CSV files")
	cmd.Flags().StringVar(&opts.configDir, "config-dir", defaultSankhyaConfigDir, "Directory containing GETL config files for Sankhya catalog datasets")
	cmd.Flags().StringVar(&opts.syncManifestPath, "sync-manifest", defaultSankhyaSyncManifest, "Path to the dataset sync manifest file")
	cmd.Flags().StringVar(&opts.pgDSN, "pg-dsn", os.Getenv("KUBEX_SANKHYA_CATALOG_DSN"), "PostgreSQL DSN used by GETL destination")
	cmd.Flags().StringVar(&opts.schema, "schema", defaultSankhyaCatalogSchema, "Destination schema for Sankhya metadata tables")
	cmd.Flags().StringVar(&opts.refreshMode, "refresh-mode", defaultSankhyaCatalogRefresh, "Refresh mode for dataset loading (currently full_refresh)")
	cmd.Flags().StringSliceVar(&opts.datasets, "datasets", nil, "Optional subset of dataset names to sync")

	return cmd
}

func runSankhyaCatalogSync(ctx context.Context, opts *sankhyaSyncOptions) error {
	gl.SetDebugMode(opts.debug)

	if err := loadOptionalEnv(opts.envFile); err != nil {
		return err
	}
	if opts.catalogDir == "" {
		opts.catalogDir = defaultSankhyaCatalogDir
	}
	if opts.configDir == "" {
		opts.configDir = defaultSankhyaConfigDir
	}
	if opts.syncManifestPath == "" {
		opts.syncManifestPath = defaultSankhyaSyncManifest
	}
	if opts.schema == "" {
		opts.schema = defaultSankhyaCatalogSchema
	}
	if opts.refreshMode == "" {
		opts.refreshMode = defaultSankhyaCatalogRefresh
	}
	if strings.TrimSpace(opts.pgDSN) == "" {
		return fmt.Errorf("pg-dsn is required")
	}
	if opts.refreshMode != defaultSankhyaCatalogRefresh {
		return fmt.Errorf("unsupported refresh-mode %q", opts.refreshMode)
	}
	if err := validateSQLIdentifier(opts.schema); err != nil {
		return fmt.Errorf("invalid schema: %w", err)
	}

	catalogDir, err := filepath.Abs(opts.catalogDir)
	if err != nil {
		return fmt.Errorf("failed to resolve catalog-dir: %w", err)
	}
	configDir, err := filepath.Abs(opts.configDir)
	if err != nil {
		return fmt.Errorf("failed to resolve config-dir: %w", err)
	}
	syncManifestPath, err := filepath.Abs(opts.syncManifestPath)
	if err != nil {
		return fmt.Errorf("failed to resolve sync-manifest path: %w", err)
	}

	_ = os.Setenv("KUBEX_SANKHYA_CATALOG_SOURCE_DIR", catalogDir)
	_ = os.Setenv("KUBEX_SANKHYA_CATALOG_DSN", opts.pgDSN)
	_ = os.Setenv("KUBEX_SANKHYA_CATALOG_SCHEMA", opts.schema)

	syncManifest, err := loadSankhyaSyncManifest(syncManifestPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(syncManifest.Schema) != "" {
		syncManifest.Schema = strings.TrimSpace(syncManifest.Schema)
	}
	if strings.TrimSpace(syncManifest.SourceSystem) == "" {
		syncManifest.SourceSystem = "sankhya"
	}
	if strings.TrimSpace(syncManifest.Domain) == "" {
		syncManifest.Domain = defaultSankhyaCatalogDomain
	}
	catalogManifestPath := filepath.Join(catalogDir, "_manifest.json")
	if strings.TrimSpace(syncManifest.SourceManifest) != "" {
		catalogManifestPath = os.ExpandEnv(syncManifest.SourceManifest)
	}
	catalogManifest, err := loadSankhyaCatalogManifest(catalogManifestPath)
	if err != nil {
		return err
	}

	cfg := gcfg.LoadConfig()
	if _, err := dsstore.Init(ctx, cfg); err != nil {
		return fmt.Errorf("failed to initialize DS runtime: %w", err)
	}
	registryStore, err := dsstore.ExternalMetadataStore(ctx)
	if err != nil {
		return fmt.Errorf("failed to resolve external metadata store: %w", err)
	}

	pgDB, err := sql.Open("postgres", opts.pgDSN)
	if err != nil {
		return fmt.Errorf("failed to open postgres connection: %w", err)
	}
	defer pgDB.Close()

	if err := pgDB.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping postgres destination: %w", err)
	}
	if err := ensureSchema(ctx, pgDB, opts.schema); err != nil {
		return err
	}

	selected, err := selectDatasets(syncManifest, opts.datasets)
	if err != nil {
		return err
	}

	results := make([]sankhyaSyncResult, 0, len(selected))
	for _, dataset := range selected {
		result := runSingleSankhyaDatasetSync(ctx, pgDB, registryStore, syncManifest, catalogManifest, configDir, catalogDir, opts.schema, opts.refreshMode, dataset)
		results = append(results, result)
		if result.Error != nil {
			gl.Errorf("dataset %s failed: %v", result.Dataset, result.Error)
		} else {
			gl.Successf("dataset %s synced (%d rows)", result.Dataset, result.RowCount)
		}
	}

	failed := 0
	for _, result := range results {
		if result.Error != nil {
			failed++
		}
	}

	gl.Infof("sankhya catalog sync finished: %d total, %d success, %d failed", len(results), len(results)-failed, failed)
	for _, result := range results {
		status := result.Status
		if status == "" {
			status = "unknown"
		}
		if result.Error != nil {
			gl.Infof(" - %s -> %s (%s): %v", result.Dataset, result.TableName, status, result.Error)
			continue
		}
		gl.Infof(" - %s -> %s (%s, %d rows)", result.Dataset, result.TableName, status, result.RowCount)
	}

	if failed > 0 {
		return fmt.Errorf("%d dataset(s) failed during sankhya catalog sync", failed)
	}
	return nil
}

func runSingleSankhyaDatasetSync(
	ctx context.Context,
	pgDB *sql.DB,
	registryStore dsclient.ExternalMetadataStore,
	syncManifest *sankhyaSyncManifest,
	catalogManifest *sankhyaCatalogManifest,
	configDir string,
	catalogDir string,
	schema string,
	refreshMode string,
	dataset sankhyaSyncManifestDataset,
) sankhyaSyncResult {
	result := sankhyaSyncResult{
		Dataset:   dataset.Name,
		TableName: qualifiedName(schema, dataset.Name),
	}

	configPath := dataset.Config
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(configDir, dataset.Config)
	}
	result.ConfigPath = configPath

	if err := validateSQLIdentifier(dataset.Name); err != nil {
		result.Error = fmt.Errorf("invalid dataset/table identifier: %w", err)
		_ = upsertRegistryFailure(ctx, registryStore, syncManifest, dataset, schema, refreshMode, nil, result.Error)
		return result
	}

	cfg, err := getlutils.LoadConfigFile(configPath)
	if err != nil {
		result.Error = fmt.Errorf("failed to load GETL config: %w", err)
		_ = upsertRegistryFailure(ctx, registryStore, syncManifest, dataset, schema, refreshMode, nil, result.Error)
		return result
	}

	cfg.SourceType = "csv"
	cfg.DestinationType = "postgres"
	cfg.SourceConnectionString = filepath.Join(catalogDir, dataset.Name+".csv")
	cfg.DestinationConnectionString = os.ExpandEnv(cfg.DestinationConnectionString)
	cfg.DestinationTable = result.TableName
	rowCount := result.RowCount

	if entry, ok := catalogManifestDatasetByName(catalogManifest, dataset.Name); ok && entry.Rows == 0 {
		if err := ensureEmptyCSVTable(ctx, pgDB, schema, dataset.Name, cfg.SourceConnectionString); err != nil {
			result.Error = fmt.Errorf("failed to materialize empty dataset: %w", err)
			_ = upsertRegistryFailure(ctx, registryStore, syncManifest, dataset, schema, refreshMode, catalogManifest, result.Error)
			return result
		}
		rowCount = 0
		result.RowCount = rowCount
		result.Status = "empty"
		goto registerResult
	}

	if refreshMode == defaultSankhyaCatalogRefresh {
		if err := dropQualifiedTable(ctx, pgDB, schema, dataset.Name); err != nil {
			result.Error = fmt.Errorf("failed to prepare destination table: %w", err)
			_ = upsertRegistryFailure(ctx, registryStore, syncManifest, dataset, schema, refreshMode, catalogManifest, result.Error)
			return result
		}
	}

	if err := getlsql.LoadData(nil, cfg); err != nil {
		if isEmptyDatasetError(err) {
			if err := ensureEmptyCSVTable(ctx, pgDB, schema, dataset.Name, cfg.SourceConnectionString); err != nil {
				result.Error = fmt.Errorf("failed to materialize empty dataset: %w", err)
				_ = upsertRegistryFailure(ctx, registryStore, syncManifest, dataset, schema, refreshMode, catalogManifest, result.Error)
				return result
			}
			result.RowCount = 0
			result.Status = "empty"
		} else {
			result.Error = fmt.Errorf("getl load failed: %w", err)
			_ = upsertRegistryFailure(ctx, registryStore, syncManifest, dataset, schema, refreshMode, catalogManifest, result.Error)
			return result
		}
	}

	if result.Status != "empty" {
		rowCount, err = countRows(ctx, pgDB, schema, dataset.Name)
		if err != nil {
			result.Error = fmt.Errorf("failed to count destination rows: %w", err)
			_ = upsertRegistryFailure(ctx, registryStore, syncManifest, dataset, schema, refreshMode, catalogManifest, result.Error)
			return result
		}
		result.RowCount = rowCount
		result.Status = "ready"
	}

registerResult:
	manifestPayload, err := buildRegistryManifestPayload(syncManifest, catalogManifest, dataset, cfg.SourceConnectionString, result.TableName, refreshMode, rowCount)
	if err != nil {
		result.Error = fmt.Errorf("failed to build registry manifest payload: %w", err)
		_ = upsertRegistryFailure(ctx, registryStore, syncManifest, dataset, schema, refreshMode, catalogManifest, result.Error)
		return result
	}

	loadedAt := time.Now().UTC()
	status := result.Status
	_, err = registryStore.Upsert(ctx, &dsclient.UpsertExternalMetadataInput{
		SourceSystem: syncManifest.SourceSystem,
		Domain:       datasetDomain(syncManifest, dataset),
		SchemaName:   schema,
		DatasetName:  dataset.Name,
		TableName:    result.TableName,
		Manifest:     manifestPayload,
		RowCount:     &rowCount,
		LastLoadedAt: &loadedAt,
		LoadMode:     stringPtr(refreshMode),
		Status:       &status,
	})
	if err != nil {
		result.Error = fmt.Errorf("failed to upsert registry record: %w", err)
		return result
	}

	return result
}

func upsertRegistryFailure(
	ctx context.Context,
	registryStore dsclient.ExternalMetadataStore,
	syncManifest *sankhyaSyncManifest,
	dataset sankhyaSyncManifestDataset,
	schema string,
	refreshMode string,
	catalogManifest *sankhyaCatalogManifest,
	failure error,
) error {
	manifestPayload, err := buildRegistryManifestPayload(syncManifest, catalogManifest, dataset, filepath.Join(os.Getenv("KUBEX_SANKHYA_CATALOG_SOURCE_DIR"), dataset.Name+".csv"), qualifiedName(schema, dataset.Name), refreshMode, 0)
	if err != nil {
		return err
	}
	status := "failed"
	notes := failure.Error()
	loadedAt := time.Now().UTC()
	_, err = registryStore.Upsert(ctx, &dsclient.UpsertExternalMetadataInput{
		SourceSystem: syncManifest.SourceSystem,
		Domain:       datasetDomain(syncManifest, dataset),
		SchemaName:   schema,
		DatasetName:  dataset.Name,
		TableName:    qualifiedName(schema, dataset.Name),
		Manifest:     manifestPayload,
		LastLoadedAt: &loadedAt,
		LoadMode:     stringPtr(refreshMode),
		Status:       &status,
		Notes:        &notes,
	})
	return err
}

func loadOptionalEnv(envFile string) error {
	if strings.TrimSpace(envFile) == "" {
		return nil
	}
	envFile = os.ExpandEnv(envFile)
	if _, err := os.Stat(envFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			gl.Warnf("env file not found at %s, continuing with current environment", envFile)
			return nil
		}
		return fmt.Errorf("failed to inspect env file %s: %w", envFile, err)
	}
	if err := godotenv.Overload(envFile); err != nil {
		return fmt.Errorf("failed to load env file %s: %w", envFile, err)
	}
	return nil
}

func loadSankhyaSyncManifest(path string) (*sankhyaSyncManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read sync manifest %s: %w", path, err)
	}
	var manifest sankhyaSyncManifest
	if err := json.Unmarshal(stripUTF8BOMBytes(data), &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse sync manifest %s: %w", path, err)
	}
	if len(manifest.Datasets) == 0 {
		return nil, fmt.Errorf("sync manifest %s does not declare any datasets", path)
	}
	return &manifest, nil
}

func loadSankhyaCatalogManifest(path string) (*sankhyaCatalogManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read source catalog manifest %s: %w", path, err)
	}
	var manifest sankhyaCatalogManifest
	if err := json.Unmarshal(stripUTF8BOMBytes(data), &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse source catalog manifest %s: %w", path, err)
	}
	return &manifest, nil
}

func selectDatasets(manifest *sankhyaSyncManifest, requested []string) ([]sankhyaSyncManifestDataset, error) {
	if len(requested) == 0 {
		selected := append([]sankhyaSyncManifestDataset(nil), manifest.Datasets...)
		sort.Slice(selected, func(i, j int) bool { return selected[i].Name < selected[j].Name })
		return selected, nil
	}

	lookup := make(map[string]sankhyaSyncManifestDataset, len(manifest.Datasets))
	for _, dataset := range manifest.Datasets {
		lookup[dataset.Name] = dataset
	}

	selected := make([]sankhyaSyncManifestDataset, 0, len(requested))
	for _, name := range requested {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		dataset, ok := lookup[name]
		if !ok {
			return nil, fmt.Errorf("dataset %q not found in sync manifest", name)
		}
		selected = append(selected, dataset)
	}
	sort.Slice(selected, func(i, j int) bool { return selected[i].Name < selected[j].Name })
	return selected, nil
}

func ensureSchema(ctx context.Context, db *sql.DB, schema string) error {
	quotedSchema, err := quoteIdentifier(schema)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", quotedSchema))
	if err != nil {
		return fmt.Errorf("failed to ensure schema %s: %w", schema, err)
	}
	return nil
}

func dropQualifiedTable(ctx context.Context, db *sql.DB, schema string, table string) error {
	qualified, err := quoteQualifiedName(schema, table)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", qualified))
	if err != nil {
		return fmt.Errorf("failed to drop table %s.%s: %w", schema, table, err)
	}
	return nil
}

func countRows(ctx context.Context, db *sql.DB, schema string, table string) (int64, error) {
	qualified, err := quoteQualifiedName(schema, table)
	if err != nil {
		return 0, err
	}
	var count int64
	if err := db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", qualified)).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func ensureEmptyCSVTable(ctx context.Context, db *sql.DB, schema string, table string, sourceFile string) error {
	csvTable := getlextr.NewCSVDataTable(nil, sourceFile)
	if err := csvTable.LoadFile(); err != nil {
		return err
	}

	headers := csvTable.Headers()
	if len(headers) == 0 {
		return fmt.Errorf("csv %s does not contain headers", sourceFile)
	}

	columns := make([]string, 0, len(headers))
	for _, header := range headers {
		if err := validateSQLIdentifier(header); err != nil {
			return fmt.Errorf("invalid csv header %q: %w", header, err)
		}
		quoted, err := quoteIdentifier(header)
		if err != nil {
			return err
		}
		columns = append(columns, fmt.Sprintf("%s TEXT", quoted))
	}

	qualified, err := quoteQualifiedName(schema, table)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", qualified, strings.Join(columns, ", ")))
	if err != nil {
		return fmt.Errorf("failed to create empty table %s: %w", qualified, err)
	}
	return nil
}

func buildRegistryManifestPayload(syncManifest *sankhyaSyncManifest, catalogManifest *sankhyaCatalogManifest, dataset sankhyaSyncManifestDataset, sourceFile string, destinationTable string, refreshMode string, rowCount int64) ([]byte, error) {
	datasetEntry := map[string]any{}
	catalogSummary := map[string]any{}
	if catalogManifest != nil {
		catalogSummary["extracted_at_utc"] = catalogManifest.ExtractedAtUTC
		catalogSummary["server"] = catalogManifest.Server
		catalogSummary["database"] = catalogManifest.Database
		if entry, ok := catalogManifestDatasetByName(catalogManifest, dataset.Name); ok {
			datasetEntry = map[string]any{
				"dataset": entry.Dataset,
				"rows":    entry.Rows,
				"file":    entry.File,
			}
		}
	}

	payload := map[string]any{
		"source_system": syncManifest.SourceSystem,
		"domain":        datasetDomain(syncManifest, dataset),
		"dataset":       dataset.Name,
		"source_file":   sourceFile,
		"destination": map[string]any{
			"table":        destinationTable,
			"refresh_mode": refreshMode,
			"row_count":    rowCount,
		},
		"catalog_manifest": mergeMap(catalogSummary, map[string]any{"dataset": datasetEntry}),
	}
	return json.Marshal(payload)
}

func datasetDomain(syncManifest *sankhyaSyncManifest, dataset sankhyaSyncManifestDataset) string {
	if strings.TrimSpace(dataset.Domain) != "" {
		return strings.TrimSpace(dataset.Domain)
	}
	if strings.TrimSpace(syncManifest.Domain) != "" {
		return strings.TrimSpace(syncManifest.Domain)
	}
	return defaultSankhyaCatalogDomain
}

func quoteIdentifier(identifier string) (string, error) {
	if err := validateSQLIdentifier(identifier); err != nil {
		return "", err
	}
	return `"` + identifier + `"`, nil
}

func quoteQualifiedName(schema string, table string) (string, error) {
	quotedSchema, err := quoteIdentifier(schema)
	if err != nil {
		return "", err
	}
	quotedTable, err := quoteIdentifier(table)
	if err != nil {
		return "", err
	}
	return quotedSchema + "." + quotedTable, nil
}

func qualifiedName(schema string, table string) string {
	return fmt.Sprintf("%s.%s", schema, table)
}

func validateSQLIdentifier(identifier string) error {
	if !sqlIdentifierPattern.MatchString(identifier) {
		return fmt.Errorf("identifier %q must match %s", identifier, sqlIdentifierPattern.String())
	}
	return nil
}

func stringPtr(value string) *string {
	return &value
}

func mergeMap(base map[string]any, extra map[string]any) map[string]any {
	merged := make(map[string]any, len(base)+len(extra))
	for key, value := range base {
		merged[key] = value
	}
	for key, value := range extra {
		merged[key] = value
	}
	return merged
}

func stripUTF8BOMBytes(data []byte) []byte {
	return []byte(strings.TrimPrefix(string(data), "\uFEFF"))
}

func isEmptyDatasetError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "no data to extract")
}

func catalogManifestDatasetByName(manifest *sankhyaCatalogManifest, datasetName string) (sankhyaCatalogManifestDataset, bool) {
	if manifest == nil {
		return sankhyaCatalogManifestDataset{}, false
	}
	for _, entry := range manifest.Datasets {
		if entry.Dataset == datasetName {
			return entry, true
		}
	}
	return sankhyaCatalogManifestDataset{}, false
}
