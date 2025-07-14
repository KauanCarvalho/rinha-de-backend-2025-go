package handler

import (
	"encoding/json"

	"github.com/KauanCarvalho/rinha-de-backend-2025-go/internal/model"
	"github.com/KauanCarvalho/rinha-de-backend-2025-go/internal/queue"
	"github.com/valyala/fasthttp"
)

func HandlePaymentCreate(ctx *fasthttp.RequestCtx) {
	if !ctx.IsPost() {
		ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
		return
	}

	var input model.PaymentInput
	if err := json.Unmarshal(ctx.PostBody(), &input); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString("invalid JSON payload")
		return
	}

	if err := input.Validate(); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString(err.Error())
		return
	}

	if err := queue.EnqueuePayment(input); err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString("failed to enqueue payment")
		return
	}

	ctx.SetStatusCode(fasthttp.StatusAccepted)
}

func HandlePaymentPurge(ctx *fasthttp.RequestCtx) {
	if !ctx.IsPost() {
		ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
		return
	}

	if err := queue.PurgePayments(); err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString("failed to purge Redis queue")
		return
	}

	ctx.SetContentType("application/json")
	json.NewEncoder(ctx).Encode(map[string]string{
		"message": "Redis queue purged successfully",
	})
}
