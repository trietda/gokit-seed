package test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	kitzap "github.com/go-kit/kit/log/zap"
	"github.com/go-kit/kit/sd/lb"
	"github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"
	"go.uber.org/zap"
)

type Handler struct {
	*http.ServeMux
}

func (h Handler) Pattern() string {
	return "/test/"
}

func MakeHandler(logger *zap.Logger, sv TestService) Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorHandler(
			transport.NewLogErrorHandler(
				kitzap.NewZapSugarLogger(logger, zap.DebugLevel),
			),
		),
		kithttp.ServerErrorEncoder(decodeHelloError),
	}

	handler := Handler{}
	router := http.NewServeMux()

	reverseHandler := kithttp.NewServer(
		makeReverseEndpoint(sv),
		decodeReverseRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)
	router.Handle("POST /foo", reverseHandler)

	helloHandler := kithttp.NewServer(
		makeHelloEndpoint(sv),
		kithttp.NopRequestDecoder,
		kithttp.EncodeJSONResponse,
		opts...,
	)
	router.Handle("GET /bar", helloHandler)

	handler.ServeMux = router
	return handler
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
