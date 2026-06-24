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

type Token struct {
	Kid                   string        `json:"kid"`
	Aad                   string        `json:"aad"`
	AccessTokenTTL        time.Duration `json:"access_token_ttl"`
	AccessTokenCookieKey  string        `json:"access_token_cookie_key"`
	RefreshTokenTTL       time.Duration `json:"refresh_token_ttl"`
	RefreshTokenCookieKey string        `json:"refresh_token_cookie_key"`
}

type Session struct {
	CookieKey string        `json:"cookie_key"`
	CookieTTL time.Duration `json:"cookie_ttl"`
}

type Aes struct {
	Key string `json:"key"`
}

type Redis struct {
	Addr         string        `json:"addr"`
	Password     string        `json:"password"`
	DB           int           `json:"db"`
	Prefix       string        `json:"prefix"`
	PoolSize     int           `json:"pool_size"`
	MinIdleConns int           `json:"min_idle_conns"`
	MaxRetries   int           `json:"max_retries"`
	DialTimeout  time.Duration `json:"dial_timeout"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	PoolTimeout  time.Duration `json:"pool_timeout"`
}

type Logger struct {
	LogsDir    string `json:"logs_dir"`
	MaxSize    int    `json:"max_size"` // MB
	MaxBackups int    `json:"max_backups"`
	MaxAge     int    `json:"max_age"`
}

type CORS struct {
	AllowOrigins     []string      `json:"allow_origins"`
	AllowMethods     []string      `json:"allow_methods"`
	AllowHeaders     []string      `json:"allow_headers"`
	AllowCredentials bool          `json:"allow_credentials"`
	MaxAge           time.Duration `json:"max_age"`
}

type Config struct {
	Scheme   Scheme   `json:"scheme"`
	Database Database `json:"database"`
	Token    Token    `json:"token"`
	Session  Session  `json:"session"`
	Aes      Aes      `json:"aes"`
	Redis    Redis    `json:"redis"`
	Logger   Logger   `json:"logger"`
	CORS     CORS     `json:"cors"`
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
	logsDir := filepath.Join(flags.Data, "logs")

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
		Token: Token{
			Kid:                   "erewhile-iam-public-key",
			Aad:                   "30bAOV+0Upo+D3T7c9DPl/hah5ChhXy0",
			AccessTokenTTL:        5 * time.Minute,
			AccessTokenCookieKey:  "atck",
			RefreshTokenTTL:       24 * time.Hour,
			RefreshTokenCookieKey: "rtck",
		},
		Session: Session{
			CookieKey: "iam_sid",
			CookieTTL: 12 * time.Hour,
		},
		Aes: Aes{
			Key: "co1FsGScYJirTXZ+ymVm/mbZ+4Lhrep2",
		},
		Redis: Redis{
			Addr:         "127.0.0.1:6379",
			Password:     "",
			DB:           0,
			Prefix:       "iam",
			PoolSize:     100,
			MinIdleConns: 10,
			MaxRetries:   3,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
			PoolTimeout:  4 * time.Second,
		},
		Logger: Logger{
			LogsDir:    logsDir,
			MaxSize:    50,
			MaxBackups: 10,
			MaxAge:     24,
		},
		CORS: CORS{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
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
