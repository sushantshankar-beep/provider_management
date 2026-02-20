package worker

import (
	"context"
	"log"
	"provider_management/internal/service"
	"time"
)

type PromoStatusWorker struct {
	svc *service.PromoCodeService
}

func NewPromoStatusWorker(svc *service.PromoCodeService) *PromoStatusWorker {
	return &PromoStatusWorker{svc: svc}
}

func (w *PromoStatusWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				syncCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				if err := w.svc.SyncPromoStatuses(syncCtx); err != nil {
					log.Println("promo status sync error:", err)
				}
				cancel()
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}