package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KauanCarvalho/rinha-de-backend-2025-go/internal/crons"
	"github.com/KauanCarvalho/rinha-de-backend-2025-go/internal/queue"
	"github.com/KauanCarvalho/rinha-de-backend-2025-go/internal/redis"
)

func main() {
	redis.MustConnect()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				crons.RunHealthcheckManager()
			}
		}
	}()

	queue.StartBroker(ctx)

	<-ctx.Done()
}
