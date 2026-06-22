package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/erewhile/iam/cmd/flags"
	"github.com/erewhile/iam/config"
	"github.com/erewhile/iam/internal/ent/db"
	"github.com/erewhile/iam/internal/ent/db/migrate"
	_ "github.com/erewhile/iam/internal/ent/db/runtime" // fix db: uninitialized user.DefaultCreatedAt (forgotten import db/runtime?)
	_ "github.com/go-sql-driver/mysql"
)

const (
	pingTimeout  = 5 * time.Second
	closeTimeout = 5 * time.Second
)

var (
	mu     sync.RWMutex
	client *db.Client
	sqlDB  *sql.DB
)

func Init(cfg config.Database) error {
	dsn := cfg.DSN()
	if dsn == "" {
		return fmt.Errorf("database dsn is empty")
	}

	drv, err := entsql.Open(dialect.MySQL, dsn)
	if err != nil {
		return fmt.Errorf("failed opening connection to mysql: %w", err)
	}

	dbConn := drv.DB()
	dbConn.SetMaxIdleConns(cfg.MaxIdleConns)
	dbConn.SetMaxOpenConns(cfg.MaxOpenConns)
	dbConn.SetConnMaxLifetime(cfg.MaxLifetime * time.Second)
	dbConn.SetConnMaxIdleTime(cfg.MaxLifetime * time.Second / 2)

	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()
	if err := dbConn.PingContext(ctx); err != nil {
		_ = drv.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	newClient := db.NewClient(db.Driver(drv))
	if flags.Debug {
		newClient = newClient.Debug()
	}

	mu.Lock()
	client = newClient
	sqlDB = dbConn
	mu.Unlock()

	return nil
}

func GetDB() *db.Client {
	mu.RLock()
	defer mu.RUnlock()
	if client == nil {
		log.Fatal("database client is not initialized, call Init() first")
	}
	return client
}

func Migrate(ctx context.Context) error {
	if err := GetDB().Schema.Create(
		ctx,
		migrate.WithForeignKeys(false),
	); err != nil {
		return fmt.Errorf("failed creating schema resources: %w", err)
	}
	return nil
}

func Healthy(ctx context.Context) error {
	mu.RLock()
	conn := sqlDB
	mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("database is not initialized")
	}
	return conn.PingContext(ctx)
}

func Close() error {
	mu.Lock()
	c := client
	client = nil
	sqlDB = nil
	mu.Unlock()

	if c == nil {
		return nil
	}

	done := make(chan error, 1)
	go func() { done <- c.Close() }()

	select {
	case err := <-done:
		if err != nil {
			log.Printf("error failed to close database: %v\n", err)
			return err
		}
		return nil
	case <-time.After(closeTimeout):
		err := fmt.Errorf("timeout closing database after %s", closeTimeout)
		log.Println(err)
		return err
	}
}
