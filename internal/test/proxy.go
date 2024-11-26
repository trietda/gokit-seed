package test

import (
	"context"
	"gokit-seed/internal/common"
	"strings"
	"time"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	kitratelimit "github.com/go-kit/kit/ratelimit"
	kitsd "github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/lb"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
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

		instanceUrls := strings.Split((*proxyUrl), ",")

		if len(instanceUrls) == 0 {
			return ts
		}

		var endpointers = kitsd.FixedEndpointer{}

		for _, instanceUrl := range instanceUrls {
			e := makeHelloProxy(context.Background(), instanceUrl)
			e = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(e)
			e = kitratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(5*time.Second), 1))(e)

			endpointers = append(endpointers, e)
		}

		loadbalancer := lb.NewRoundRobin(endpointers)
		retry := lb.Retry(1, time.Duration(5*time.Second), loadbalancer)

		return proxymw{
			context.Background(),
			ts,
			retry,
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
