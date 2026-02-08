package helpers

import (
	"os"
	"strconv"
	"time"
)

func GetEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func GetDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	sec, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return time.Duration(sec) * time.Second
}
