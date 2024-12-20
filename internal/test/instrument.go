package test

import (
	"context"
	"gokit-seed/internal/common"
	"time"
)

type instrumentmw struct {
	TestService
	reverseMetrics common.Metrics
	hellowMetrics  common.Metrics
}

func MakeInstrumentMiddleware(reverseMetrics common.Metrics, hellowMetrics common.Metrics) ServiceMiddleware {
	return func(ts TestService) TestService {
		return instrumentmw{
			ts,
			reverseMetrics,
			hellowMetrics,
		}
	}
}

func (mw instrumentmw) Reverse(s string) string {
	defer mw.reverseMetrics.Collect(time.Now(), "method", "reverse")
	return mw.TestService.Reverse(s)
}

func (mw instrumentmw) Hello(ctx context.Context) (string, error) {
	defer mw.hellowMetrics.Collect(time.Now(), "method", "hello")
	return mw.TestService.Hello(ctx)
}
