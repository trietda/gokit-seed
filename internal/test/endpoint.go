package test

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

func makeTestEnpoint(sv TestService) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		return sv.Test(request), nil
	}
}
