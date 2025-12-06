package admin

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/bantuaku/backend/services/storage"
)

// liveBroadcaster provides in-memory fan-out for live usage payloads with backpressure.
type liveBroadcaster struct {
	db    *storage.Postgres
	mu    sync.RWMutex
	subs  map[chan string]struct{}
	tick  time.Duration
	stop  chan struct{}
	start sync.Once
}

var (
	liveHub     *liveBroadcaster
	liveHubOnce sync.Once
)

func ensureLiveHub(db *storage.Postgres) *liveBroadcaster {
	liveHubOnce.Do(func() {
		liveHub = &liveBroadcaster{
			db:   db,
			subs: make(map[chan string]struct{}),
			tick: 5 * time.Second,
			stop: make(chan struct{}),
		}
		liveHub.run()
	})
	return liveHub
}

func (b *liveBroadcaster) run() {
	b.start.Do(func() {
		go func() {
			ticker := time.NewTicker(b.tick)
			defer ticker.Stop()
			for {
				select {
				case <-b.stop:
					return
				case <-ticker.C:
					payload, err := b.fetchAndMarshal()
					if err != nil {
						continue
					}
					b.broadcast(payload)
				}
			}
		}()
	})
}

func (b *liveBroadcaster) fetchAndMarshal() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	payload, err := (&AdminHandler{db: b.db}).fetchLiveUsage(ctx)
	if err != nil {
		return "", err
	}
	bytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (b *liveBroadcaster) broadcast(msg string) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.subs {
		select {
		case ch <- msg:
		default:
			// Backpressure: drop if subscriber is slow
		}
	}
}

func (b *liveBroadcaster) subscribe() (chan string, func()) {
	ch := make(chan string, 8) // small buffer to smooth bursts
	b.mu.Lock()
	b.subs[ch] = struct{}{}
	b.mu.Unlock()

	cancel := func() {
		b.mu.Lock()
		delete(b.subs, ch)
		close(ch)
		b.mu.Unlock()
	}
	return ch, cancel
}

