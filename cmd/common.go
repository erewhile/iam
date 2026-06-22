package cmd

import (
	"context"
	"log"

	"github.com/erewhile/iam/config"
	"github.com/erewhile/iam/internal/database"
)

func setup() {
	config.Init()
	if err := database.Init(config.Get().Database); err != nil {
		log.Fatalf("failed to init database: %v", err)
	}
	if err := database.Migrate(context.Background()); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}
}

func release() {
	database.Close()
}
