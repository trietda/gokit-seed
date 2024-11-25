package test

import (
	"context"
	"gokit-seed/internal/common"

	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
)

type proxymw struct {
	ctx   context.Context
	next  TestService
	test2 endpoint.Endpoint
}

func MakeProxyTestService(proxyUrl *string) ServiceMiddleware {
	return func(ts TestService) TestService {
		if proxyUrl == nil {
			return ts
		}

		return proxymw{
			context.Background(),
			ts,
			makeHelloProxy(context.Background(), *proxyUrl),
		}
	}
}

func (p proxymw) Reverse(s string) string {
	return p.next.Reverse(s)
}

func (p proxymw) Hello() (string, error) {
	response, responseErr := p.test2(p.ctx, nil)

	if responseErr != nil {
		return "", responseErr
	}

	responseData := response.(HelloResponse)

	return responseData.Result, nil
}

func makeHelloProxy(ctx context.Context, url string) endpoint.Endpoint {
	return kithttp.NewClient(
		"GET",
		common.MustParseUrl(url),
		kithttp.EncodeJSONRequest,
		decodeHelloResponse,
	).Endpoint()
}
