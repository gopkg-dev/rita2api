package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"rita.ai/internal/app"
	"rita.ai/internal/config"
	"rita.ai/internal/httpapi"
	"rita.ai/internal/store"
	"rita.ai/internal/upstream"
)

func main() {
	cfg := config.Load()
	if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0o755); err != nil {
		log.Fatalf("create data dir: %v", err)
	}

	repo, err := store.OpenSQLite(cfg.DBPath)
	if err != nil {
		log.Fatalf("open sqlite: %v", err)
	}
	defer repo.Close()

	client := upstream.NewClient(cfg.RitaBaseURL, cfg.RitaOrigin, cfg.RitaModelTypeID, cfg.RitaModelID)
	service := app.NewService(repo, client, app.Config{
		VisitorSecret: cfg.RitaVisitorSecret,
		DefaultRatio:  "1:1",
		DefaultRes:    "1K",
	})

	if err := service.RecoverRunningTasks(context.Background()); err != nil {
		log.Printf("recover tasks: %v", err)
	}

	server := httpapi.NewServer(service, httpapi.ServerConfig{
		CookieName: cfg.CookieName,
	})

	log.Printf("rita ai server listening on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, server.Handler()); err != nil {
		log.Fatalf("listen and serve: %v", err)
	}
}
