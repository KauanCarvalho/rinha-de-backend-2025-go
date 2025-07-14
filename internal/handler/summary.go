package handler

import (
	"encoding/json"

	"github.com/KauanCarvalho/rinha-de-backend-2025-go/internal/payments"
	"github.com/valyala/fasthttp"
)

func HandlePaymentsSummary(ctx *fasthttp.RequestCtx) {
	from := string(ctx.QueryArgs().Peek("from"))
	to := string(ctx.QueryArgs().Peek("to"))

	summary, err := payments.GetSummary(from, to)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString("failed to summarize payments")
		return
	}

	ctx.SetContentType("application/json")
	json.NewEncoder(ctx).Encode(summary)
}
