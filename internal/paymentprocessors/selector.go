package paymentprocessors

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/KauanCarvalho/rinha-de-backend-2025-go/internal/clients/processor"
	"github.com/KauanCarvalho/rinha-de-backend-2025-go/internal/lock"
	"github.com/KauanCarvalho/rinha-de-backend-2025-go/internal/redis"
)

const (
	lockKey                  = "lock:choose_processor"
	lockTTL                  = 3 * time.Second
	cacheKey                 = "selected_payment_processor"
	cacheTTL                 = 10 * time.Second
	DefaultProcessor         = "default"
	FallbackProcessor        = "fallback"
	maxLatencyInMs           = 50
	maxLatencyDifferenceInMs = 50
)

type cachedProcessor struct {
	CurrentProcessor string          `json:"current_processor"`
	Def              json.RawMessage `json:"def"`
	Fbk              json.RawMessage `json:"fbk"`
	Overwritten      bool            `json:"overwritten"`
	TS               time.Time       `json:"ts"`
}

func ChooseAndCacheProcessor(ctx context.Context) error {
	def, _ := processor.DefaultHealthcheck()
	fbk, _ := processor.FallbackHealthcheck()

	selected := selectProcessor(def, fbk)

	return setProcessorCache(ctx, selected, def, fbk, false)
}

func RecalculateProcessor(ctx context.Context, processorFailing string) error {
	return lock.WithRedisLock(ctx, lockKey, lockTTL, func() {
		doRecalculateProcessor(ctx, processorFailing)
	})
}

func doRecalculateProcessor(ctx context.Context, processorFailing string) error {
	cache, err := getCachedProcessor(ctx)
	if err != nil || cache == nil || cache.Overwritten || processorFailing != cache.CurrentProcessor {
		return nil
	}

	def := &processor.HealthResponse{}
	fbk := &processor.HealthResponse{}

	if err := json.Unmarshal(cache.Def, def); err != nil {
		return err
	}
	if err := json.Unmarshal(cache.Fbk, fbk); err != nil {
		return err
	}

	if processorFailing == DefaultProcessor {
		def.Failing = true
	} else {
		fbk.Failing = true
	}

	selected := selectProcessor(def, fbk)

	return setProcessorCache(ctx, selected, def, fbk, true)
}

func CurrentProcessor(ctx context.Context) (string, error) {
	cache, err := getCachedProcessor(ctx)
	if err != nil || cache == nil {
		return DefaultProcessor, nil
	}

	switch cache.CurrentProcessor {
	case DefaultProcessor, FallbackProcessor:
		return cache.CurrentProcessor, nil
	default:
		return DefaultProcessor, nil
	}
}

func getCachedProcessor(ctx context.Context) (*cachedProcessor, error) {
	val, err := redis.Client.Get(ctx, cacheKey).Result()
	if err != nil || val == "" {
		return nil, err
	}

	var parsed cachedProcessor
	if err := json.Unmarshal([]byte(val), &parsed); err != nil {
		return nil, err
	}

	return &parsed, nil
}

func setProcessorCache(ctx context.Context, selected string, def, fbk *processor.HealthResponse, overwritten bool) error {
	defJSON, err := json.Marshal(def)
	if err != nil {
		return fmt.Errorf("failed to marshal def: %w", err)
	}
	fbkJSON, err := json.Marshal(fbk)
	if err != nil {
		return fmt.Errorf("failed to marshal fbk: %w", err)
	}

	payload := cachedProcessor{
		CurrentProcessor: selected,
		Def:              defJSON,
		Fbk:              fbkJSON,
		Overwritten:      overwritten,
		TS:               time.Now().UTC(),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	return redis.Client.Set(ctx, cacheKey, data, cacheTTL).Err()
}

func selectProcessor(def, fbk *processor.HealthResponse) string {
	switch {
	case def != nil && !def.Failing && (fbk == nil || fbk.Failing):
		return DefaultProcessor

	case fbk != nil && !fbk.Failing && (def == nil || def.Failing):
		return FallbackProcessor

	case def == nil && fbk == nil:
		return DefaultProcessor

	case def != nil && fbk != nil && !def.Failing && !fbk.Failing:
		if isDefaultPreferred(def, fbk) {
			return DefaultProcessor
		}
		return FallbackProcessor
	}

	return DefaultProcessor
}

func isDefaultPreferred(def, fbk *processor.HealthResponse) bool {
	if def.MinResponseTime <= maxLatencyInMs {
		return true
	}
	if fbk.MinResponseTime <= maxLatencyInMs {
		return false
	}
	return def.MinResponseTime < fbk.MinResponseTime+maxLatencyDifferenceInMs
}
