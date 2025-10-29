package zapx

import "go.uber.org/zap"

type Logger struct {
	*zap.Logger
}

func NewLogger(logLevel string) (*Logger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	return &Logger{logger}, nil
}

func (l *Logger) Sync() error {
	return l.Logger.Sync()
}
