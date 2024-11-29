package test

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

func makeReverseEndpoint(sv TestService) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(ReverseRequest)
		result := sv.Reverse(req.Value)
		return ReverseResponse{result}, nil
	}
}

func makeHelloEndpoint(sv TestService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		result, resultErr := sv.Hello(ctx)
		return HelloResponse{result}, resultErr
	}
}
