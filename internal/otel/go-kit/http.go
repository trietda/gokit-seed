package gokit

import (
	"context"
	"encoding/json"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func DefaultJsonEncoder(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(w.Header()))
	return json.NewEncoder(w).Encode(response)
}
