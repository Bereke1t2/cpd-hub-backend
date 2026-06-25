//go:build ignore
// Template for Phase 6 — copy to: internal/infrastructure/external/worker.go
//
// Optional background worker that warms the contests cache so user requests are
// always cache hits. Start it from main.go with the server's cancelable context.
package external

import (
	"context"
	"log"
	"time"
)

// StartContestsRefresher calls List() every interval until ctx is cancelled.
// Keep interval modest (5–10m) to respect upstream rate limits.
func StartContestsRefresher(ctx context.Context, cc *CachedContests, interval time.Duration) {
	go func() {
		// warm immediately on boot
		if _, err := cc.List(); err != nil {
			log.Printf("contests refresher: initial warm failed: %v", err)
		}
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Println("contests refresher: stopping")
				return
			case <-t.C:
				if _, err := cc.List(); err != nil {
					log.Printf("contests refresher: refresh failed: %v", err)
				}
			}
		}
	}()
}

// Wiring in cmd/server/main.go (after building cc):
//   ctx, cancel := context.WithCancel(context.Background())
//   defer cancel()
//   external.StartContestsRefresher(ctx, cc, 7*time.Minute)
