package test

import (
	"context"
	"encoding/json"
	"errors"
	"gokit-seed/internal/common"
	"net/http"

	kitzap "github.com/go-kit/kit/log/zap"
	"github.com/go-kit/kit/sd/lb"
	"github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	// "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

func MakeHandler(logger *zap.Logger, sv TestService) (router *common.RouteGroup) {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorHandler(
			transport.NewLogErrorHandler(
				kitzap.NewZapSugarLogger(logger, zap.DebugLevel),
			),
		),
		kithttp.ServerErrorEncoder(decodeHelloError),
	}

	router = common.NewRouteGroup("/strings")

	var reverseHandler http.Handler
	reverseHandler = kithttp.NewServer(
		makeReverseEndpoint(sv),
		decodeReverseRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)
	reversePath := "/reversions"
	reverseHandler = otelhttp.WithRouteTag(reversePath, reverseHandler)
	router.Handler("POST", reversePath, reverseHandler)

	var helloHandler http.Handler
	helloHandler = kithttp.NewServer(
		makeHelloEndpoint(sv),
		kithttp.NopRequestDecoder,
		kithttp.EncodeJSONResponse,
		opts...,
	)
	helloPath := "/greetings"
	helloHandler = otelhttp.WithRouteTag(helloPath, helloHandler)
	router.Handler("GET", helloPath, helloHandler)

	return
}

type ReverseRequest struct {
	Value string `json:"value"`
}

func decodeReverseRequest(_ context.Context, r *http.Request) (interface{}, error) {
	body := ReverseRequest{}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}

	return body, nil
}

type ReverseResponse struct {
	Result string `json:"result"`
}

type HelloResponse struct {
	Result string `json:"result"`
}

func decodeHelloResponse(_ context.Context, r *http.Response) (interface{}, error) {
	body := HelloResponse{}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}

	return body, nil
}

func decodeHelloError(ctx context.Context, err error, w http.ResponseWriter) {
	if errors.As(err, &lb.RetryError{}) {
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}

	kithttp.DefaultErrorEncoder(ctx, err, w)
}
