package otel

func NameRpc(serviceName string, methodName string) string {
	return serviceName + "." + methodName
}
