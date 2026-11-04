package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultSunatBetaURL = "https://e-beta.sunat.gob.pe/ol-ti-itcpfegem-beta/billService"
	DefaultSunatProdURL = "https://e-factura.sunat.gob.pe/ol-ti-itcpfegem/billService"
)

type Config struct {
	HTTPAddr       string
	DatabaseURL    string
	RedisURL       string
	SessionTTL     time.Duration
	MigrateOnStart bool
	CORSOrigins    []string

	WorkerPollInterval time.Duration
	WorkerBatchSize    int
	WorkerMaxIntentos  int
	WorkerBackoffBase  time.Duration
	SunatFakeMode      string
	BlobDir            string
	FrontendPublicURL  string

	SunatRUC             string
	SunatSOLUser         string
	SunatSOLPassword     string
	SunatCertPath        string
	SunatCertPassword    string
	SunatWSURL           string
	SunatEnv             string
	SunatAllowProduction bool
	SunatMinInterval     time.Duration
}

func Load() (Config, error) {
	ttlStr := getenv("SESSION_TTL", "8h")
	ttl, err := time.ParseDuration(ttlStr)
	if err != nil {
		return Config{}, fmt.Errorf("SESSION_TTL: %w", err)
	}
	migrateOnStart, err := strconv.ParseBool(getenv("MIGRATE_ON_START", "true"))
	if err != nil {
		return Config{}, fmt.Errorf("MIGRATE_ON_START: %w", err)
	}

	pollStr := getenv("WORKER_POLL_INTERVAL", "1s")
	pollInterval, err := time.ParseDuration(pollStr)
	if err != nil {
		return Config{}, fmt.Errorf("WORKER_POLL_INTERVAL: %w", err)
	}
	batchSize, err := strconv.Atoi(getenv("WORKER_BATCH_SIZE", "10"))
	if err != nil {
		return Config{}, fmt.Errorf("WORKER_BATCH_SIZE: %w", err)
	}
	maxIntentos, err := strconv.Atoi(getenv("WORKER_MAX_INTENTOS", "5"))
	if err != nil {
		return Config{}, fmt.Errorf("WORKER_MAX_INTENTOS: %w", err)
	}
	backoffStr := getenv("WORKER_BACKOFF_BASE", "1s")
	backoffBase, err := time.ParseDuration(backoffStr)
	if err != nil {
		return Config{}, fmt.Errorf("WORKER_BACKOFF_BASE: %w", err)
	}

	minIntervalStr := getenv("SUNAT_MIN_INTERVAL", "30s")
	minInterval, err := time.ParseDuration(minIntervalStr)
	if err != nil {
		return Config{}, fmt.Errorf("SUNAT_MIN_INTERVAL: %w", err)
	}

	allowProd, err := strconv.ParseBool(getenv("SUNAT_ALLOW_PRODUCTION", "false"))
	if err != nil {
		return Config{}, fmt.Errorf("SUNAT_ALLOW_PRODUCTION: %w", err)
	}

	cfg := Config{
		HTTPAddr:             getenv("HTTP_ADDR", ":8080"),
		DatabaseURL:          os.Getenv("DATABASE_URL"),
		RedisURL:             os.Getenv("REDIS_URL"),
		SessionTTL:           ttl,
		MigrateOnStart:       migrateOnStart,
		CORSOrigins:          parseCORSOrigins(getenv("CORS_ORIGINS", "http://localhost:3000")),
		WorkerPollInterval:   pollInterval,
		WorkerBatchSize:      batchSize,
		WorkerMaxIntentos:    maxIntentos,
		WorkerBackoffBase:    backoffBase,
		SunatFakeMode:        getenv("SUNAT_FAKE_MODE", "accept"),
		BlobDir:              getenv("BLOB_DIR", "./data/blobs"),
		FrontendPublicURL:    getenv("FRONTEND_PUBLIC_URL", "http://localhost:3000"),
		SunatRUC:             strings.TrimSpace(os.Getenv("SUNAT_RUC")),
		SunatSOLUser:         strings.TrimSpace(os.Getenv("SUNAT_SOL_USER")),
		SunatSOLPassword:     os.Getenv("SUNAT_SOL_PASSWORD"),
		SunatCertPath:        strings.TrimSpace(os.Getenv("SUNAT_CERT_PATH")),
		SunatCertPassword:    os.Getenv("SUNAT_CERT_PASSWORD"),
		SunatWSURL:           strings.TrimSpace(os.Getenv("SUNAT_WS_URL")),
		SunatEnv:             strings.ToLower(strings.TrimSpace(getenv("SUNAT_ENV", "beta"))),
		SunatAllowProduction: allowProd,
		SunatMinInterval:     minInterval,
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	return cfg, nil
}

// HasSunatCreds indica si el set mínimo para cliente real + firma está completo.
func (c Config) HasSunatCreds() bool {
	return c.SunatRUC != "" &&
		c.SunatSOLUser != "" &&
		c.SunatSOLPassword != "" &&
		c.SunatCertPath != "" &&
		c.SunatCertPassword != ""
}

// ResolveSunatEndpoint elige fake | beta | production.
// Producción exige SUNAT_ALLOW_PRODUCTION=true; si no, error (no silenciar).
func (c Config) ResolveSunatEndpoint() (mode, url string, err error) {
	if !c.HasSunatCreds() {
		return "fake", "", nil
	}
	env := c.SunatEnv
	if env == "" {
		env = "beta"
	}
	switch env {
	case "production", "prod":
		if !c.SunatAllowProduction {
			return "", "", fmt.Errorf("SUNAT_ENV=production requiere SUNAT_ALLOW_PRODUCTION=true")
		}
		url = DefaultSunatProdURL
		mode = "production"
	case "beta", "test", "homologacion", "homologación":
		url = DefaultSunatBetaURL
		mode = "beta"
	default:
		return "", "", fmt.Errorf("SUNAT_ENV=%q inválido (use beta o production)", c.SunatEnv)
	}
	if c.SunatWSURL != "" {
		url = strings.TrimSuffix(c.SunatWSURL, "?wsdl")
		url = strings.TrimSuffix(url, "?WSDL")
	}
	return mode, url, nil
}

// ValidateAPI exige Redis; el worker no lo usa.
func (c Config) ValidateAPI() error {
	if c.RedisURL == "" {
		return fmt.Errorf("REDIS_URL is required")
	}
	return nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseCORSOrigins(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{"http://localhost:3000"}
	}
	return out
}
