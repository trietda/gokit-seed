package otel

import (
	"gokit-seed/internal/common"
	"net/http"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type withTraceIdLog struct {
	next http.Handler
}

func WithTraceIdLog(next http.Handler) http.Handler {
	return &withTraceIdLog{next}
}

func (h *withTraceIdLog) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := common.LoggerFromContext(r.Context())
	spanContext := trace.SpanContextFromContext(r.Context())
	logger = logger.With(zap.String("trace_id", spanContext.TraceID().String()))

	ctx := common.ContextWithLogger(r.Context(), logger)
	h.next.ServeHTTP(w, r.WithContext(ctx))
}
