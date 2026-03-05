package config

import "os"

type Config struct {
	DatabaseURL       string
	TemporalHost      string
	TemporalNamespace string
	HTTPAddr          string
}

func Load() Config {
	return Config{
		DatabaseURL:       getEnv("PGURL","postgres://postgres:pass123@192.168.215.3:5432/ttdb"),
		TemporalHost:      getEnv("TEMPORAL_HOST", "localhost:7233"),
		TemporalNamespace: getEnv("TEMPORAL_NAMESPACE", "default"),
		HTTPAddr:          getEnv("HTTP_ADDR", ":8080"),
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("required env var not set: " + key)
	}
	return v
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
