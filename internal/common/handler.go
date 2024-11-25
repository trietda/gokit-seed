package common

import (
	"net/http"

	"github.com/gorilla/handlers"
	"go.uber.org/zap"
	"go.uber.org/zap/zapio"
)

func LoggingHandler(logger *zap.Logger, h http.Handler) http.Handler {
	writer := &zapio.Writer{
		Log:   logger,
		Level: zap.InfoLevel,
	}
	return handlers.LoggingHandler(writer, h)
}
