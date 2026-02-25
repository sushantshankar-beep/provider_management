package worker

import (
	"context"
	"log"
	"time"

	"provider_management/internal/config"
	"provider_management/internal/service"
)

func StartPayoutWorker(ctx context.Context, payoutService *service.PayoutService) {
	hours := config.GetPayoutIntervalHours()
	interval := time.Duration(hours) * time.Hour

	ticker := time.NewTicker(interval)

	go func() {

		for {
			select {
			case <-ticker.C:
				if err := payoutService.CreatePayoutLast6Hours(ctx, hours); err != nil {
					log.Println("Payout worker error:", err)
				}
			case <-ctx.Done():
				log.Println("Payout worker stopped")
				ticker.Stop()
				return
			}
		}
	}()
}
