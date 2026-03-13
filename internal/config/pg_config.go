// Package config fornece a configuração e utilitários para conexão com o banco de dados.
package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kubex-ecosystem/gnyx/internal/module/kbx"
)

// PGConfig define parâmetros de conexão com o Postgres exposto pelo Kubex DS.
type PGConfig struct {
	DSN             string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

const (
	envDSN             = "KUBEX_DOMUS_DSN"
	envFallbackDSN     = "postgres_URL"
	envMaxIdle         = "postgres_MAX_IDLE_CONNS"
	envMaxOpen         = "postgres_MAX_OPEN_CONNS"
	envConnLifetime    = "postgres_CONN_MAX_LIFETIME"
	envConnIdleTimeout = "postgres_CONN_MAX_IDLE_TIME"
)

// ConfigFromEnv constrói a configuração a partir das variáveis de ambiente.
func ConfigFromEnv() *PGConfig {
	cfg := &PGConfig{
		DSN:             kbx.GetEnvOrDefault(envDSN, firstNonEmpty(os.Getenv(envFallbackDSN), defaultDSN())),
		MaxIdleConns:    intFromEnv(envMaxIdle, 4),
		MaxOpenConns:    intFromEnv(envMaxOpen, 10),
		ConnMaxLifetime: durationFromEnv(envConnLifetime, 30*time.Minute),
		ConnMaxIdleTime: durationFromEnv(envConnIdleTimeout, 5*time.Minute),
	}
	return cfg
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func intFromEnv(key string, fallback int) int {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if num, err := strconv.Atoi(v); err == nil {
			return num
		}
	}
	return fallback
}

func durationFromEnv(key string, fallback time.Duration) time.Duration {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

func defaultDSN() string {
	return "postgres://kubex_adm:DOMUS_DB_PASSWORD@localhost:5432/postgres?sslmode=disable"
}
