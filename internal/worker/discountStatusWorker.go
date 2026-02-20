package worker

import (
	"context"
	"log"
	"provider_management/internal/service"
	"time"
)

type DiscountStatusWorker struct {
	svc *service.DiscountService
}

func NewDiscountStatusWorker(svc *service.DiscountService) *DiscountStatusWorker {
	return &DiscountStatusWorker{svc: svc}
}

func (w *DiscountStatusWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
log.Println("Starting DiscountStatusWorker to sync discount statuses every 1 minute")
	go func() {
		for {
			select {
			case <-ticker.C:
				syncCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				if err := w.svc.SyncDiscountStatuses(syncCtx); err != nil {
					log.Println("discount status sync error:", err)
				}
				cancel()
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}