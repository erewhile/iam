package cmd

import (
	"context"
	"log"

	"github.com/erewhile/iam/config"
	"github.com/erewhile/iam/internal/cache/rds"
	"github.com/erewhile/iam/internal/database"
	"github.com/erewhile/iam/internal/logger"
	"github.com/erewhile/iam/pkg/aes"
)

func setup() {
	config.Init()
	aes.Init([]byte(config.Get().Aes.Key))
	if err := database.Init(config.Get().Database); err != nil {
		log.Fatalf("failed to init database: %v", err)
	}
	if err := database.Migrate(context.Background()); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}
	if err := rds.Init(config.Get().Redis); err != nil {
		log.Fatalf("failed to init redis: %v", err)
	}
	if err := logger.Init(config.Get().Logger); err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
}

func release() {
	database.Close()
	rds.Close()
	logger.Close()
}
