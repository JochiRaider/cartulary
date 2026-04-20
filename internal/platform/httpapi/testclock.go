package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"
)

type TestClock struct {
	mu     sync.RWMutex
	offset time.Duration
}

func NewTestClock() *TestClock {
	return &TestClock{}
}

func (c *TestClock) Now() time.Time {
	offset := c.currentOffset()
	return time.Now().UTC().Add(offset)
}

func (c *TestClock) SetOffset(duration time.Duration) time.Time {
	c.mu.Lock()
	c.offset = duration
	now := time.Now().UTC().Add(c.offset)
	c.mu.Unlock()
	return now
}

func (c *TestClock) currentOffset() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.offset
}

func (c *TestClock) OffsetSeconds() int64 {
	return int64(c.currentOffset() / time.Second)
}

func RegisterTestClockRoutes(clock *TestClock) RouteRegistrar {
	return func(mux *http.ServeMux, deps DependencySet) error {
		_ = deps
		if clock == nil {
			return errors.New("test clock is required")
		}

		mux.HandleFunc("/api/v1/test/clock/set", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			var request struct {
				OffsetSeconds int64 `json:"offset_seconds"`
			}
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				_ = WriteError(w, r, http.StatusBadRequest, "invalid_mutation_payload", "invalid clock set payload", map[string]any{
					"field": "offset_seconds",
				})
				return
			}

			next := clock.SetOffset(time.Duration(request.OffsetSeconds) * time.Second)
			_ = WriteSuccess(w, r, http.StatusOK, map[string]any{
				"offset_seconds": clock.OffsetSeconds(),
				"now":            next.Format(time.RFC3339Nano),
			})
		})
		return nil
	}
}
