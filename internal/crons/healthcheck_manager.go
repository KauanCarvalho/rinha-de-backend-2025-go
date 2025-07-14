package crons

import (
	"context"
	"time"

	"github.com/KauanCarvalho/rinha-de-backend-2025-go/internal/lock"
	"github.com/KauanCarvalho/rinha-de-backend-2025-go/internal/paymentprocessors"
	"github.com/KauanCarvalho/rinha-de-backend-2025-go/internal/redis"
)

const (
	throttleKey = "healthcheck:throttle"
	lockKey     = "healthcheck:lock"
	throttleTTL = 5 * time.Second
	lockTTL     = 5 * time.Second
)

func RunHealthcheckManager() {
	ctx := context.Background()

	ok, err := redis.Client.SetNX(ctx, throttleKey, "1", throttleTTL).Result()
	if err != nil || !ok {
		return
	}

	err = lock.WithRedisLock(ctx, lockKey, lockTTL, func() {
		_ = paymentprocessors.ChooseAndCacheProcessor(ctx)
	})
	if err != nil {
		return
	}
}
