package redis

import (
	"context"
	"os"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func MustConnect() {
	Client = redis.NewClient(&redis.Options{
		Addr:     getEnv("REDIS_URL", "redis:6379"),
		Password: "",
		DB:       0,
	})

	if err := Client.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
