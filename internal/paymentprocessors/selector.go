package paymentprocessors

import (
	"context"
	"encoding/json"
	"time"

	"github.com/KauanCarvalho/rinha-de-backend-2025-go/internal/clients/processor"
	"github.com/KauanCarvalho/rinha-de-backend-2025-go/internal/redis"
)

const (
	cacheKey                 = "selected_payment_processor"
	ttl                      = 5 * time.Second
	defaultProcessor         = "default"
	fallbackProcessor        = "fallback"
	maxLatencyDifferenceInMs = 50
)

func ChooseAndCacheProcessor(ctx context.Context) error {
	def, _ := processor.DefaultHealthcheck()
	fbk, _ := processor.FallbackHealthcheck()

	selected := selectProcessor(def, fbk)

	payload := map[string]interface{}{
		"current_processor": selected,
		"ts":                time.Now().UTC(),
	}
	data, _ := json.Marshal(payload)

	return redis.Client.Set(ctx, cacheKey, data, ttl).Err()
}

func selectProcessor(def, fbk *processor.HealthResponse) string {
	if def != nil && !def.Failing && (fbk == nil || fbk.Failing) {
		return defaultProcessor
	}
	if fbk != nil && !fbk.Failing && (def == nil || def.Failing) {
		return fallbackProcessor
	}
	if def != nil && fbk != nil && !def.Failing && !fbk.Failing {
		if def.MinResponseTime < fbk.MinResponseTime+maxLatencyDifferenceInMs {
			return defaultProcessor
		}
		return fallbackProcessor
	}
	return defaultProcessor
}

func CurrentProcessor(ctx context.Context) (string, error) {
	val, err := redis.Client.Get(ctx, cacheKey).Result()
	if err != nil || val == "" {
		return defaultProcessor, nil
	}

	var parsed struct {
		CurrentProcessor string `json:"current_processor"`
	}

	if err := json.Unmarshal([]byte(val), &parsed); err != nil {
		return defaultProcessor, nil
	}

	if parsed.CurrentProcessor != defaultProcessor && parsed.CurrentProcessor != fallbackProcessor {
		return defaultProcessor, nil
	}

	return parsed.CurrentProcessor, nil
}
