package main

import (
	"log"
	"time"

	"github.com/KauanCarvalho/rinha-de-backend-2025-go/internal/handler"
	"github.com/KauanCarvalho/rinha-de-backend-2025-go/internal/redis"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

func main() {
	redis.MustConnect()

	r := router.New()
	r.GET("/healthcheck", handler.HandleHealthcheck)
	r.POST("/payments", handler.HandlePaymentCreate)
	r.GET("/payments-summary", handler.HandlePaymentsSummary)
	r.POST("/purge-payments", handler.HandlePaymentPurge)

	server := &fasthttp.Server{
		Handler:      r.Handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(":8080"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
