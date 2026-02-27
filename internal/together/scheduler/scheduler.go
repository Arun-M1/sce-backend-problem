package scheduler

import (
	"log"
	"stock-monitor/internal/together/finnhub"
	"stock-monitor/internal/together/store"
	"sync"
	"time"
)

type job struct {
	ticker *time.Ticker
	stop   chan struct{}
}

type Scheduler struct {
	mu     sync.Mutex
	jobs   map[string]*job
	store  *store.Store
	client *finnhub.Client
}

func New(s *store.Store, c *finnhub.Client) *Scheduler {
	return &Scheduler{
		jobs:   make(map[string]*job),
		store:  s,
		client: c,
	}
}

// monitor for symbol at interval, if job is running already stop it
func (sc *Scheduler) Start(symbol string, interval time.Duration) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	existing, ok := sc.jobs[symbol]
	if ok {
		existing.ticker.Stop()
		close(existing.stop)
	}

	//create ticker and stop channel
	ticker := time.NewTicker(interval)
	stopCh := make(chan struct{})

	sc.jobs[symbol] = &job{
		ticker: ticker,
		stop:   stopCh,
	}

	//goroutine
	go func() {
		log.Printf("[scheduler] started monitoring %s every %s", symbol, interval)
		for {
			select {
			case <-ticker.C:
				sc.fetchAndStore(symbol)
			case <-stopCh:
				log.Printf("[scheduler] stopped monitoring %s", symbol)
				return
			}
		}
	}()
}

// cancels monitoring job for symbol
func (sc *Scheduler) Stop(symbol string) bool {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	j, ok := sc.jobs[symbol]

	if !ok {
		return false
	}

	j.ticker.Stop()
	close(j.stop)
	delete(sc.jobs, symbol) //remove job from scheduler map
	return true
}

func (sc *Scheduler) fetchAndStore(symbol string) {
	quote, err := sc.client.FetchQuote(symbol)

	if err != nil {
		log.Printf("[scheduler] error fetching %s: %v", symbol, err)
		return
	}

	sc.store.Append(quote)
	log.Printf("[scheduler] fetched %s @ %.2f", symbol, quote.CurrentPrice)
}
