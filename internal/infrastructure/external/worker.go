package external

import (
	"context"
	"log"
	"time"
)

// StartContestsRefresher calls ListForUser("") every interval until ctx is cancelled.
// This warms the cache so user requests are snappy.
func StartContestsRefresher(ctx context.Context, cc *CachedContests, interval time.Duration) {
	go func() {
		// warm immediately on boot
		if _, err := cc.ListForUser(""); err != nil {
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
				if _, err := cc.ListForUser(""); err != nil {
					log.Printf("contests refresher: refresh failed: %v", err)
				}
			}
		}
	}()
}
