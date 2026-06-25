package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/erewhile/iam/config"
	"github.com/erewhile/iam/internal/cache/rds"
	"github.com/erewhile/iam/internal/database"
	"github.com/erewhile/iam/internal/logger"
	"github.com/erewhile/iam/internal/repository"
	"github.com/erewhile/iam/pkg/aes"
)

var (
	cancelCleanup context.CancelFunc
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

	repo := repository.NewTokenRepository(database.GetDB())

	var ctx context.Context
	ctx, cancelCleanup = context.WithCancel(context.Background())

	go startTokenCleanupTimer(ctx, repo)
}

func release() {
	if cancelCleanup != nil {
		cancelCleanup()
	}
	database.Close()
	rds.Close()
	logger.Close()
}

func startTokenCleanupTimer(ctx context.Context, repo repository.TokenRepository) {
	executeCleanup(ctx, repo)

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			executeCleanup(ctx, repo)
		case <-ctx.Done():
			logger.Info("token cleanup timer stopped safely.")
			return
		}
	}
}

func executeCleanup(ctx context.Context, repo repository.TokenRepository) {
	totalDeleted := 0
	batchCount := 0

	for {
		select {
		case <-ctx.Done():
			logger.Info("cleanup job interrupted during batch execution due to application shutdown.")
			return
		default:
		}

		execCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		affected, err := repo.ClearExpiredOrRevoked(execCtx)
		cancel()

		if err != nil {
			logger.Error("failed to clear token batch", err)
			return
		}

		if affected == 0 {
			if totalDeleted > 0 {
				logger.Info(fmt.Sprintf("batch cleanup finished. Total deleted: %d rows across %d batches.", totalDeleted, batchCount))
			}
			return
		}

		totalDeleted += affected
		batchCount++

		select {
		case <-time.After(100 * time.Millisecond):
		case <-ctx.Done():
			logger.Info("cleanup job interrupted during sleep due to application shutdown.")
			return
		}
	}
}
