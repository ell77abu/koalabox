package main

import (
	"fmt"
	"log"

	"task-backend/internal/bootstrap"
	"task-backend/internal/config"
	"task-backend/internal/router"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	db, err := bootstrap.InitMySQL(cfg.Database.DSN, cfg.Database.MaxIdleConns, cfg.Database.MaxOpenConns, cfg.Database.ConnMaxLifetimeMinutes)
	if err != nil {
		log.Fatalf("init mysql failed: %v", err)
	}

	r := router.NewRouter(db)
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("server starting on %s", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("server stopped with error: %v", err)
	}
}
