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
