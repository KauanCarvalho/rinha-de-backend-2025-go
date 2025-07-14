package model

import (
	"errors"
)

type PaymentInput struct {
	CorrelationID string  `json:"correlationId"`
	Amount        float64 `json:"amount"`
}

func (p *PaymentInput) Validate() error {
	if p.CorrelationID == "" {
		return errors.New("correlationId is required")
	}
	if p.Amount <= 0 {
		return errors.New("amount must be greater than zero")
	}
	return nil
}
