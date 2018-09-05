package lagerx

import (
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/logx"
)

type Logger struct {
	logger lager.Logger
}

func NewLogger(logger lager.Logger) *Logger {
	return &Logger{
		logger: logger,
	}
}

func (l *Logger) WithName(name string) logx.Logger {
	return &Logger{
		logger: l.logger.Session(name),
	}
}

func (l *Logger) WithData(data ...logx.Data) logx.Logger {
	return &Logger{
		logger: l.logger.WithData(convertData(data...)),
	}
}

func (l *Logger) Debug(msg string, data ...logx.Data) {
	l.logger.Debug(msg, convertData(data...))
}

func (l *Logger) Info(msg string, data ...logx.Data) {
	l.logger.Info(msg, convertData(data...))
}

func (l *Logger) Error(msg string, err error, data ...logx.Data) {
	l.logger.Error(msg, err, convertData(data...))
}

func convertData(data ...logx.Data) lager.Data {
	lData := make(map[string]interface{})

	for _, d := range data {
		lData[d.Key] = d.Value
	}

	return lData
}
