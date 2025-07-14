package handler

import (
	"context"
	"encoding/json"

	"github.com/KauanCarvalho/rinha-de-backend-2025-go/internal/redis"
	"github.com/valyala/fasthttp"
)

func HandleHealthcheck(ctx *fasthttp.RequestCtx) {
	if !ctx.IsGet() {
		ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
		return
	}

	res, err := redis.Client.Ping(context.Background()).Result()
	if err != nil || res != "PONG" {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		json.NewEncoder(ctx).Encode(map[string]string{"status": "error"})
		return
	}

	ctx.SetContentType("application/json")
	json.NewEncoder(ctx).Encode(map[string]string{"status": "ok"})
}
