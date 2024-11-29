package otel

import (
	"fmt"
	"net/http"
)

func NameHttpRequest(route string) func(string, *http.Request) string {
	return func(_ string, _ *http.Request) string {
		return fmt.Sprintf("HTTP %s", route)
	}
}

func NameRpc(serviceName string, methodName string) string {
	return serviceName + "." + methodName
}
