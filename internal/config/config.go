package config

import (
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {
	Addr              string
	DBPath            string
	RitaBaseURL       string
	RitaOrigin        string
	RitaVisitorSecret string
	RitaModelTypeID   int
	RitaModelID       int
	CookieName        string
}

func Load() Config {
	return Config{
		Addr:              getEnv("RITA_ADDR", ":8080"),
		DBPath:            getEnv("RITA_DB_PATH", filepath.Join("data", "rita.db")),
		RitaBaseURL:       getEnv("RITA_BASE_URL", "https://api_v2.rita.ai"),
		RitaOrigin:        getEnv("RITA_ORIGIN", "https://www.rita.ai"),
		RitaVisitorSecret: getEnv("RITA_VISITOR_SECRET", "e3438fe855d6c27b6aa3c50357d3d7115fd57917d86a7dc3c66d3086c2d8479a"),
		RitaModelTypeID:   getEnvInt("RITA_MODEL_TYPE_ID", 1032),
		RitaModelID:       getEnvInt("RITA_MODEL_ID", 1121),
		CookieName:        getEnv("RITA_COOKIE_NAME", "rita_session"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return fallback
}
