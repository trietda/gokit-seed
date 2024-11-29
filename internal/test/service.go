package test

import (
	"context"
	"fmt"
	"os"
)

type TestService interface {
	Reverse(s string) string
	Hello(ctx context.Context) (string, error)
}

type ServiceMiddleware func(TestService) TestService

type testService struct{}

func NewTestService() TestService {
	return &testService{}
}

// Reverse string, use body.Value as input
func (sv testService) Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func (sv testService) Hello(ctx context.Context) (string, error) {
	return fmt.Sprintf("Hello world! %s", os.Getenv("NAME")), nil
}
