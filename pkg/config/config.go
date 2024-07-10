package config

import (
	"os"
)

type Config struct {
	LokiURL       string
	ProxyPort     string
	ExporterPort  string
	TenantFile    string
	TrackfilePath string
}

func NewConfig() Config {
	return Config{
		LokiURL:       getenv("FLOKI_LOKI_URL", "http://localhost:3100"),
		ProxyPort:     getenv("FLOKI_PROXY_PORT", "8080"),
		ExporterPort:  getenv("FLOKI_EXPORTER_PORT", "3100"),
		TenantFile:    getenv("FLOKI_TENANTFILE_PATH", "/opt/floki/tenants.yaml"),
		TrackfilePath: getenv("FLOKI_TRACKFILE_PATH", "/opt/floki/track"),
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}
