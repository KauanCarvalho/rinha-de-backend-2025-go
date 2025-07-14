package payments

import (
	"context"
	"encoding/json"
	"time"

	"github.com/KauanCarvalho/rinha-de-backend-2025-go/internal/redis"
)

type PaymentEntry struct {
	Amount      float64 `json:"amount"`
	Processor   string  `json:"processor"`
	CreatedAt   string  `json:"createdAt"`
}

type Summary struct {
	TotalRequests int     `json:"totalRequests"`
	TotalAmount   float64 `json:"totalAmount"`
}

func GetSummary(fromStr, toStr string) (map[string]Summary, error) {
	ctx := context.Background()

	data, err := redis.Client.HGetAll(ctx, "payments").Result()
	if err != nil {
		return nil, err
	}

	from, _ := parseTime(fromStr)
	to, _ := parseTime(toStr)

	summary := map[string]Summary{
		"default":  {0, 0},
		"fallback": {0, 0},
	}

	for _, raw := range data {
		var entry PaymentEntry
		if err := json.Unmarshal([]byte(raw), &entry); err != nil {
			continue
		}

		createdAt, err := time.Parse(time.RFC3339, entry.CreatedAt)
		if err != nil {
			continue
		}

		if !inRange(createdAt, from, to) {
			continue
		}

		if val, ok := summary[entry.Processor]; ok {
			val.TotalRequests++
			val.TotalAmount += entry.Amount
			summary[entry.Processor] = val
		}
	}

	roundSummary(summary)
	return summary, nil
}

func parseTime(s string) (time.Time, bool) {
	if s == "" {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339, s)
	return t, err == nil
}

func inRange(t time.Time, from time.Time, to time.Time) bool {
	if !from.IsZero() && t.Before(from) {
		return false
	}
	if !to.IsZero() && t.After(to) {
		return false
	}
	return true
}

func roundSummary(s map[string]Summary) {
	for k, v := range s {
		s[k] = Summary{
			TotalRequests: v.TotalRequests,
			TotalAmount:   float64(int(v.TotalAmount*100)) / 100.0,
		}
	}
}
