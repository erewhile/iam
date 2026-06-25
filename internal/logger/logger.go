package logger

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/erewhile/iam/cmd/flags"
	"github.com/erewhile/iam/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	globalLogger atomic.Value
	level        zap.AtomicLevel
	once         sync.Once
	initErr      error
)

type nopSyncWriteSyncer struct {
	zapcore.WriteSyncer
}

func (w *nopSyncWriteSyncer) Sync() error {
	err := w.WriteSyncer.Sync()
	if err == nil {
		return nil
	}
	if isUnsupportedSyncErr(err) {
		return nil
	}
	return err
}

func isUnsupportedSyncErr(err error) bool {
	if errors.Is(err, os.ErrInvalid) {
		return true
	}

	errStr := err.Error()
	return strings.Contains(errStr, "invalid argument") ||
		strings.Contains(errStr, "not supported") ||
		strings.Contains(errStr, "handle is invalid")
}

func Init(cfg config.Logger) error {
	once.Do(func() {
		initErr = doInit(cfg)
	})
	return initErr
}

func doInit(cfg config.Logger) error {
	if err := os.MkdirAll(cfg.LogsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	level = zap.NewAtomicLevelAt(zap.InfoLevel)
	if flags.Debug {
		level.SetLevel(zap.DebugLevel)
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderCfg.EncodeCaller = zapcore.ShortCallerEncoder

	jsonEncoder := zapcore.NewJSONEncoder(encoderCfg)

	newWriter := func(filename string) zapcore.WriteSyncer {
		ws := zapcore.AddSync(&lumberjack.Logger{
			Filename:   filepath.Join(cfg.LogsDir, filename),
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   true,
			LocalTime:  true,
		})

		return &zapcore.BufferedWriteSyncer{
			WS:            ws,
			Size:          256 * 1024,
			FlushInterval: 5 * time.Second,
		}
	}

	var cores []zapcore.Core

	cores = append(cores, zapcore.NewCore(
		jsonEncoder,
		newWriter("info.log"),
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return level.Enabled(lvl) && lvl < zapcore.WarnLevel
		}),
	))

	cores = append(cores, zapcore.NewCore(
		jsonEncoder,
		newWriter("error.log"),
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return level.Enabled(lvl) && lvl >= zapcore.WarnLevel
		}),
	))

	if flags.Debug {
		consoleEnc := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		stdoutSyncer := &nopSyncWriteSyncer{WriteSyncer: zapcore.AddSync(os.Stdout)}
		cores = append(cores, zapcore.NewCore(
			consoleEnc,
			stdoutSyncer,
			level,
		))
	}

	core := zapcore.NewTee(cores...)

	opts := []zap.Option{
		zap.AddCaller(),
		zap.AddCallerSkip(2),
		zap.AddStacktrace(zapcore.ErrorLevel),
	}
	if flags.Debug {
		opts = append(opts, zap.Development())
	}

	globalLogger.Store(zap.New(core, opts...))
	return nil
}

func SetLevel(lvl zapcore.Level) {
	level.SetLevel(lvl)
}

func L() *zap.Logger {
	if v := globalLogger.Load(); v != nil {
		return v.(*zap.Logger)
	}
	return zap.NewNop()
}

func Close() {
	if v := globalLogger.Load(); v != nil {
		if err := v.(*zap.Logger).Sync(); err != nil {
			fmt.Fprintf(os.Stderr, "logger: sync failed: %v\n", err)
		}
	}
}

func toFields(args ...any) []zap.Field {
	fields := make([]zap.Field, 0, len(args))
	extraIdx := 0

	for _, arg := range args {
		switch v := arg.(type) {
		case zap.Field:
			fields = append(fields, v)

		case error:
			if v != nil {
				fields = append(fields, zap.Error(v))
			}

		default:
			fields = append(fields, zap.Any(fmt.Sprintf("extra_%d", extraIdx), v))
			extraIdx++
		}
	}

	return fields
}

func write(lvl zapcore.Level, msg string, args ...any) {
	L().Check(lvl, msg).Write(toFields(args...)...)
}

func Debug(msg string, args ...any) { write(zapcore.DebugLevel, msg, args...) }
func Info(msg string, args ...any)  { write(zapcore.InfoLevel, msg, args...) }
func Warn(msg string, args ...any)  { write(zapcore.WarnLevel, msg, args...) }
func Error(msg string, args ...any) { write(zapcore.ErrorLevel, msg, args...) }

func Fatal(msg string, args ...any) {
	L().Check(zapcore.FatalLevel, msg).Write(toFields(args...)...)
	Close()
	os.Exit(1)
}
