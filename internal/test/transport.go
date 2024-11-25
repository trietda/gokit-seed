package test

import (
	"context"
	"encoding/json"
	"net/http"

	kitzap "github.com/go-kit/kit/log/zap"
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
		kithttp.ServerErrorEncoder(kithttp.DefaultErrorEncoder),
	}

	handler := Handler{}
	router := http.NewServeMux()

	testHandler := kithttp.NewServer(
		makeTestEnpoint(sv),
		decodeTestRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)
	router.Handle("POST /foo", testHandler)

	handler.ServeMux = router
	return handler
}

func decodeTestRequest(_ context.Context, r *http.Request) (interface{}, error) {
	body := map[string]interface{}{}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}

	return body, nil
}
