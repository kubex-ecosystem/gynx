// Package datastore fornece uma interface simplificada para interagir com o DataStore (DS)
package datastore

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"sync"

	cf "github.com/kubex-ecosystem/gnyx/internal/config"
	ds "github.com/kubex-ecosystem/gnyx/internal/dsclient"
	gl "github.com/kubex-ecosystem/logz"
)

var (
	clientOnce   sync.Once
	client       ds.DSClient
	clientErr    error
	cfgOnce      sync.Once
	cfg          *cf.Config
	cfgDBService *DBServiceConfig
	dbKey        string
)

// PGExecutor é exposto para reuso interno no BE.
type PGExecutor = ds.PGExecutor

// DBServiceConfig reduz a dependência às informações necessárias do DS.
type DBServiceConfig = cf.DataServiceConfig

// Init configura o DSClient global usando a configuração do app.
func Init(ctx context.Context, argCfg *cf.Config) (ds.DSClient, error) {
	if argCfg == nil {
		argCfg = cf.LoadConfig()
	}
	cfgOnce.Do(func() {
		cfg = argCfg
		cfgDBService = argCfg.DataService
		dbKey = argCfg.DataService.DBName
	})

	clientOnce.Do(func() {
		logger := gl.GetLoggerZ("github.com/kubex-ecosystem/gnyx_datastore")
		client = ds.NewDSClient(ctx, cfg, logger)
		if key, err := resolveDBKey(cfgDBService); err == nil && strings.TrimSpace(key) != "" {
			dbKey = key
		}
		clientErr = client.Init(ctx)
	})

	return client, clientErr
}

// Client retorna o DSClient inicializado.
func Client(ctx context.Context) (ds.DSClient, error) {
	if client != nil && clientErr == nil {
		return client, nil
	}
	return Init(ctx, cfg)
}

// Connection retorna a conexão padrão do BE.
func Connection(ctx context.Context) (*ds.BackendConnection, error) {
	c, err := Client(ctx)
	if err != nil {
		return nil, err
	}
	candidates := []string{dbKey, cfg.DataService.DBName, "domus"}
	seen := map[string]bool{}
	var conn *ds.BackendConnection
	for _, name := range candidates {
		n := strings.TrimSpace(name)
		if n == "" || seen[n] {
			continue
		}
		seen[n] = true
		conn, err = c.GetConn(ctx, n)
		if err == nil {
			dbKey = n
			break
		}
	}
	if err != nil {
		// Fallback: tenta resolver novamente o ID e reabrir dinamicamente.
		if key, resErr := resolveDBKey(cfgDBService); resErr == nil && strings.TrimSpace(key) != "" && !seen[key] {
			dbKey = key
			conn, err = c.GetConn(ctx, dbKey)
		}
	}
	if err != nil {
		return nil, gl.Errorf("failed to get DS connection: %v", err)
	}
	return conn, nil
}

// UserStore retorna o store de usuários do DS.
func UserStore(ctx context.Context) (ds.UserStore, error) {
	c, err := Client(ctx)
	if err != nil {
		return nil, err
	}
	conn, err := c.GetConn(ctx, dbKey)
	if err != nil {
		return nil, gl.Errorf("failed to get user store: %v", err)
	}
	store, err := ds.NewUserStore(ctx, conn)
	if err != nil {
		return nil, gl.Errorf("failed to get user store: %v", err)
	}
	return store, nil
}

// PendingAccessStore retorna o store de solicitações de acesso pendentes do DS.
func PendingAccessStore(ctx context.Context) (ds.PendingAccessStore, error) {
	c, err := Client(ctx)
	if err != nil {
		return nil, err
	}
	conn, err := c.GetConn(ctx, dbKey)
	if err != nil {
		return nil, gl.Errorf("failed to get pending access store: %v", err)
	}
	store, err := ds.NewPendingAccessStore(ctx, conn)
	if err != nil {
		return nil, gl.Errorf("failed to get pending access store: %v", err)
	}
	return store, nil
}

// GetInviteStore retorna o store de convites do DS.
func GetInviteStore(ctx context.Context) (ds.InviteStore, error) {
	c, err := Client(ctx)
	if err != nil {
		return nil, err
	}
	conn, err := c.GetConn(ctx, dbKey)
	if err != nil {
		return nil, gl.Errorf("failed to get invite store: %v", err)
	}
	store, err := ds.NewInviteStore(ctx, conn)
	if err != nil {
		return nil, gl.Errorf("failed to get invite store: %v", err)
	}
	return store, nil
}

// GetPGExecutor retorna o executor PG associado à conexão padrão.
func GetPGExecutor(ctx context.Context, conn *ds.BackendConnection) (PGExecutor, error) {
	return ds.GetPGExecutor(ctx, conn)
}

// resolveDBKey busca o ID habilitado correspondente ao nome/config informado.
func resolveDBKey(c *DBServiceConfig) (string, error) {
	path := strings.TrimSpace(c.ConfigPath)
	if path == "" {
		return "", gl.Errorf("config path is required")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var root struct {
		Databases []struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Enabled bool   `json:"enabled"`
		} `json:"databases"`
	}
	if err := json.Unmarshal(data, &root); err != nil {
		return "", err
	}
	target := strings.TrimSpace(c.DBName)
	for _, db := range root.Databases {
		if !db.Enabled {
			continue
		}
		if db.Name == target || db.ID == target {
			return db.ID, nil
		}
	}
	for _, db := range root.Databases {
		if db.Enabled {
			return db.ID, nil
		}
	}
	return "", gl.Errorf("no enabled database found")
}
