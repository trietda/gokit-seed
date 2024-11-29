package common

import "os"

func MustGetEnv(key string) string {
	value := os.Getenv(key)

	if value == "" {
		panic(key + " is not set")
	}

	return value
}

func GetEnv(key string) *string {
	value := os.Getenv(key)

	if value == "" {
		return nil
	}

	return &value
}

func DefaultGetEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)

	if value == "" {
		return defaultValue
	}

	return value == "true"
}
