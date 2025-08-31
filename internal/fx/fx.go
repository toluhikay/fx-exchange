package fx

import (
	"context"
	"database/sql"
	"fmt"
	"maps"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
)

type FXProvider interface {
	GetRate(ctx context.Context, from, to string) (float64, error)
	SubscribeRates() chan map[string]map[string]float64
	StartRateUpdates(ctx context.Context, db *sql.DB, interval time.Duration)
}

type MockFXProvider struct {
	rates     map[string]map[string]float64
	dynamic   bool
	mu        sync.RWMutex
	rateChans []chan map[string]map[string]float64
	chansMu   sync.Mutex
}

func NewMockFXProvider(dynamic bool) *MockFXProvider {
	baseRates := map[string]map[string]float64{
		"cNGN": {"USDx": 0.0006, "EURx": 0.0005, "cXAF": 0.36},
		"cXAF": {"USDx": 0.0017, "EURx": 0.0015, "cNGN": 2.78},
		"USDx": {"cNGN": 1666.67, "cXAF": 588.24, "EURx": 0.88},
		"EURx": {"cNGN": 2000.0, "cXAF": 666.67, "USDx": 1.14},
	}
	return &MockFXProvider{
		rates:     baseRates,
		dynamic:   dynamic,
		rateChans: []chan map[string]map[string]float64{},
	}
}

func (m *MockFXProvider) GetRate(ctx context.Context, from, to string) (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if from == to {
		return 1.0, nil
	}
	rateMap, ok := m.rates[from]
	if !ok {
		return 0, fmt.Errorf("unsupported currency: %s", from)
	}
	rate, ok := rateMap[to]
	if !ok {
		return 0, fmt.Errorf("unsupported currency pair: %s/%s", from, to)
	}
	return rate, nil
}

func (m *MockFXProvider) SubscribeRates() chan map[string]map[string]float64 {
	m.chansMu.Lock()
	defer m.chansMu.Unlock()
	ch := make(chan map[string]map[string]float64, 1)
	m.rateChans = append(m.rateChans, ch)
	return ch
}

func (m *MockFXProvider) StartRateUpdates(ctx context.Context, db *sql.DB, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			m.chansMu.Lock()
			for _, ch := range m.rateChans {
				close(ch)
			}
			m.rateChans = nil
			m.chansMu.Unlock()
			return
		case <-ticker.C:
			m.mu.Lock()
			if m.dynamic {
				for from, rateMap := range m.rates {
					for to, rate := range rateMap {
						fluctuation := 1 + (rand.Float64()*0.01 - 0.005) // ±0.5%
						m.rates[from][to] = rate * fluctuation
						query := `INSERT INTO fx_rates (id, from_currency, to_currency, rate, timestamp) 
                                 VALUES ($1, $2, $3, $4, $5)`
						_, err := db.ExecContext(ctx, query, uuid.New().String(), from, to, m.rates[from][to], time.Now())
						if err != nil {
							fmt.Printf("Failed to log FX rate: %v\n", err)
						}
					}
				}
			}
			rateCopy := make(map[string]map[string]float64)
			for from, rateMap := range m.rates {
				rateCopy[from] = make(map[string]float64)
				// rateCopy[from][to] = rate
				maps.Copy(rateCopy[from], rateMap)
			}
			m.mu.Unlock()
			m.chansMu.Lock()
			activeChans := m.rateChans[:0]
			for _, ch := range m.rateChans {
				select {
				case ch <- rateCopy:
					activeChans = append(activeChans, ch)
				default:
					close(ch)
				}
			}
			m.rateChans = activeChans
			m.chansMu.Unlock()
		}
	}
}

// Placeholder for live FX client
type FXClient struct{}

func NewFXClient() *FXClient {
	return &FXClient{}
}

func (c *FXClient) GetRate(ctx context.Context, from, to string) (float64, error) {
	return 0, fmt.Errorf("live FX client not implemented")
}

func (c *FXClient) SubscribeRates() chan map[string]map[string]float64 {
	return make(chan map[string]map[string]float64)
}

func (c *FXClient) StartRateUpdates(ctx context.Context, db *sql.DB, interval time.Duration) {
	// Implement live rate updates when needed
}
