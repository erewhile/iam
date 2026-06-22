package config

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/erewhile/iam/cmd/flags"
	"github.com/erewhile/iam/pkg/utils"
)

type Scheme struct {
	Port string `json:"port"`
}

type Database struct {
	Host         string        `json:"host"`
	Port         string        `json:"port"`
	User         string        `json:"user"`
	Password     string        `json:"password"`
	DBName       string        `json:"db_name"`
	Timezone     string        `json:"timezone"`
	MaxIdleConns int           `json:"max_idle_conns"`
	MaxOpenConns int           `json:"max_open_conns"`
	MaxLifetime  time.Duration `json:"max_lifetime"`
}

func (d *Database) DSN() string {
	tz := d.Timezone
	if tz == "" {
		tz = "Asia/Shanghai"
	}

	params := url.Values{}
	params.Set("charset", "utf8mb4")
	params.Set("parseTime", "True")
	params.Set("loc", tz)

	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?%s",
		d.User,
		d.Password,
		d.Host,
		d.Port,
		d.DBName,
		params.Encode(),
	)
}

type Config struct {
	Scheme   Scheme   `json:"scheme"`
	Database Database `json:"database"`
}

var (
	cfgPtr   atomic.Pointer[Config]
	fullPath string
	once     sync.Once
)

func Get() *Config {
	return cfgPtr.Load()
}

func defaultConfig() *Config {
	return &Config{
		Scheme: Scheme{
			Port: ":26621",
		},
		Database: Database{
			Host:         "127.0.0.1",
			Port:         "3306",
			User:         "root",
			Password:     "root",
			DBName:       "iam",
			Timezone:     "Asia/Shanghai",
			MaxIdleConns: 10,
			MaxOpenConns: 100,
			MaxLifetime:  time.Hour,
		},
	}
}

func load() error {
	if fullPath == "" {
		return errors.New("config path not initialized, call Init() first")
	}

	var newConfig Config
	if err := utils.ReadJSON(fullPath, &newConfig); err != nil {
		return fmt.Errorf("failed to load config file: %w", err)
	}

	cfgPtr.Store(&newConfig)

	return nil
}

func Init() {
	once.Do(func() {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("failed to get working directory: %v\n", err)
		}

		fullPath = filepath.Join(wd, flags.Data, "config.json")

		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			def := defaultConfig()
			if err := utils.WriteJSON(fullPath, &def); err != nil {
				log.Fatalf("failed to initialize config file: %v", err)
			}
			cfgPtr.Store(def)
		} else {
			if err := load(); err != nil {
				log.Fatalf("failed to load config file: %v", err)
			}
		}
	})
}
