package zapx

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
	*zap.Logger
}

func NewLogger(logLevel string) (*ZapLogger, error) {
	logLevelMap := map[string]zapcore.Level{
		"debug": zapcore.DebugLevel,
		"info":  zapcore.InfoLevel,
		"warn":  zapcore.WarnLevel,
		"error": zapcore.ErrorLevel,
		"fatal": zapcore.FatalLevel,
		"panic": zapcore.PanicLevel,
	}

	level, exists := logLevelMap[logLevel]
	if !exists {
		return &ZapLogger{zap.NewNop()}, nil
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(level)
	cfg.InitialFields = map[string]any{"pid": os.Getpid()}

	zl, err := cfg.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	)
	if err != nil {
		return nil, err
	}

	return &ZapLogger{zl}, nil
}

func (zl *ZapLogger) Close() error {
	return zl.Logger.Sync()
}
