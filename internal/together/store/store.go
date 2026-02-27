package store

import (
	"stock-monitor/internal/together/models"
	"sync"
)

type Store struct {
	mu      sync.RWMutex
	history map[string][]models.StockQuote
}

// constructor go equivalent
func New() *Store {
	return &Store{
		history: make(map[string][]models.StockQuote),
	}
}

// add to actual Store object history
func (s *Store) Append(q models.StockQuote) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.history[q.Symbol] = append(s.history[q.Symbol], q)
}

func (s *Store) GetHistory(symbol string) []models.StockQuote {
	s.mu.RLock()
	defer s.mu.RUnlock()

	quotes := s.history[symbol]

	if quotes == nil {
		return []models.StockQuote{}
	}

	return quotes
}

func (s *Store) Clear(symbol string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.history[symbol]

	if !ok {
		return false
	}

	delete(s.history, symbol)
	return true

}
