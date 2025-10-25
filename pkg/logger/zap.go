package logger

import "go.uber.org/zap"

type ZapLogger struct {
	*zap.Logger
}

func NewLogger(logLevel string) (*ZapLogger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	return &ZapLogger{logger}, nil
}
