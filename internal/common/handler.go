package common

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"go.uber.org/zap"
	"go.uber.org/zap/zapio"
)

type key int

var loggerKey key

func LoggerFromContext(ctx context.Context) *zap.Logger {
	return ctx.Value(loggerKey).(*zap.Logger)
}

func ContextWithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

type HandleChain func(http.Handler) http.Handler

func BaseHandler(logger *zap.Logger, h http.Handler, chains ...HandleChain) http.Handler {
	handler := WithLogging(h)

	for _, chain := range chains {
		handler = chain(handler)
	}

	return WithLogger(logger, WithRequestId(handler))
}

type withLogging struct {
	next http.Handler
}

func WithLogging(next http.Handler) http.Handler {
	return &withLogging{next}
}

func (h *withLogging) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := LoggerFromContext(r.Context())
	writer := &zapio.Writer{
		Log:   logger,
		Level: zap.InfoLevel,
	}
	handlers.LoggingHandler(writer, h.next).ServeHTTP(w, r)
}

type withLogger struct {
	logger *zap.Logger
	next   http.Handler
}

func WithLogger(logger *zap.Logger, next http.Handler) http.Handler {
	return &withLogger{logger: logger, next: next}
}

func (h *withLogger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), loggerKey, h.logger)
	h.next.ServeHTTP(w, r.WithContext(ctx))
}

type withRequestId struct {
	next http.Handler
}

func WithRequestId(next http.Handler) http.Handler {
	return &withRequestId{next}
}

func (h *withRequestId) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestId := r.Header.Get("X-Request-Id")

	if requestId == "" {
		requestId = uuid.NewString()
	}

	logger := LoggerFromContext(r.Context())
	logger = logger.With(zap.String("request_id", requestId))

	w.Header().Set("X-Request-Id", requestId)
	ctx := context.WithValue(r.Context(), loggerKey, logger)
	h.next.ServeHTTP(w, r.WithContext(ctx))
}
