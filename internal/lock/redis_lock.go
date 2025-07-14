package lock

import (
	"context"
	"errors"
	"time"

	"github.com/KauanCarvalho/rinha-de-backend-2025-go/internal/redis"
)

func WithRedisLock(ctx context.Context, key string, ttl time.Duration, fn func()) error {
	value := time.Now().UnixNano()

	ok, err := redis.Client.SetNX(ctx, key, value, ttl).Result()
	if err != nil || !ok {
		return errors.New("lock not acquired")
	}
	defer releaseLock(ctx, key, value)

	fn()
	return nil
}

func releaseLock(ctx context.Context, key string, value int64) {
	lua := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`
	redis.Client.Eval(ctx, lua, []string{key}, value)
}
