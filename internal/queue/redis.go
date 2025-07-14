package queue

import (
	"context"
	"encoding/json"

	"github.com/KauanCarvalho/rinha-de-backend-2025-go/internal/model"
	"github.com/KauanCarvalho/rinha-de-backend-2025-go/internal/redis"
)

func EnqueuePayment(p model.PaymentInput) error {
	payload, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return redis.Client.RPush(context.Background(), mainQueue, payload).Err()
}

func PurgePayments() error {
	return redis.Client.Del(context.Background(), mainQueue).Err()
}
