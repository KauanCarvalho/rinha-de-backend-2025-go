package queue

import (
	"context"
	"encoding/json"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/KauanCarvalho/rinha-de-backend-2025-go/internal/clients/processor"
	"github.com/KauanCarvalho/rinha-de-backend-2025-go/internal/paymentprocessors"
	"github.com/KauanCarvalho/rinha-de-backend-2025-go/internal/redis"
)

const (
	mainQueue      = "payments_created"
	resultHash     = "payments"
	defaultWorkers = 200
	enqueuers      = 1
)

type Payload struct {
	CorrelationID string  `json:"correlationId"`
	Amount        float64 `json:"amount"`
}

var paymentQueue = make(chan []byte, 20000)

func StartBroker(ctx context.Context) {
	workerCount := defaultWorkers

	for i := range enqueuers {
		go startRedisEnqueuer(ctx, i)
	}

	for i := range workerCount {
		go startPaymentProcessor(ctx, i)
	}
}

func startRedisEnqueuer(ctx context.Context, id int) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			result, err := redis.Client.BLPop(ctx, 1*time.Second, mainQueue).Result()
			if err == goredis.Nil || len(result) < 2 {
				continue
			} else if err != nil {
				continue
			}

			msg := result[1]

			select {
			case paymentQueue <- []byte(msg):
			default:
				_ = redis.Client.LPush(ctx, mainQueue, msg).Err()
			}
		}
	}
}

func startPaymentProcessor(ctx context.Context, workerNum int) {
	for {
		select {
		case <-ctx.Done():
			return
		case raw := <-paymentQueue:
			var payload Payload
			if err := json.Unmarshal(raw, &payload); err != nil {
				continue
			}
			processPayload(ctx, payload)
		}
	}
}

func processPayload(ctx context.Context, payload Payload) {
	processorName, err := paymentprocessors.CurrentProcessor(ctx)
	if err != nil {
		requeue(ctx, payload)
		return
	}

	params := processor.PaymentRequest{
		CorrelationID: payload.CorrelationID,
		Amount:        payload.Amount,
		RequestedAt:   time.Now().UTC().Format(time.RFC3339Nano),
	}

	var sendErr error
	switch processorName {
	case "default":
		sendErr = processor.DefaultCreatePayment(params)
	case "fallback":
		sendErr = processor.FallbackCreatePayment(params)
	default:
		sendErr = processor.DefaultCreatePayment(params)
	}

	if sendErr != nil {
		paymentprocessors.RecalculateProcessor(ctx, processorName)
		requeue(ctx, payload)
		return
	}

	data := map[string]interface{}{
		"correlationId": payload.CorrelationID,
		"amount":        payload.Amount,
		"processor":     processorName,
		"requestedAt":   params.RequestedAt,
	}
	resultJSON, _ := json.Marshal(data)
	_ = redis.Client.HSet(ctx, resultHash, payload.CorrelationID, resultJSON).Err()
}

func requeue(ctx context.Context, payload Payload) {
	msg, _ := json.Marshal(payload)
	_ = redis.Client.LPush(ctx, mainQueue, msg).Err()
}
