package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"
	"time"
)

type TestClock struct {
	mu       sync.RWMutex
	offset   time.Duration
	fixed    time.Time
	fixedSet bool
}

func NewTestClock() *TestClock {
	return &TestClock{}
}

func (c *TestClock) Now() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.fixedSet {
		return c.fixed
	}
	return time.Now().UTC().Add(c.offset)
}

func (c *TestClock) SetOffset(duration time.Duration) time.Time {
	c.mu.Lock()
	c.offset = duration
	c.fixed = time.Time{}
	c.fixedSet = false
	now := time.Now().UTC().Add(c.offset)
	c.mu.Unlock()
	return now
}

func (c *TestClock) SetFixed(now time.Time) time.Time {
	c.mu.Lock()
	c.offset = 0
	c.fixed = now.UTC()
	c.fixedSet = true
	fixed := c.fixed
	c.mu.Unlock()
	return fixed
}

func (c *TestClock) Advance(duration time.Duration) time.Time {
	c.mu.Lock()
	if c.fixedSet {
		c.fixed = c.fixed.Add(duration).UTC()
		advanced := c.fixed
		c.mu.Unlock()
		return advanced
	}
	c.offset += duration
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
		if !TestRoutesEnabled(deps.Env) {
			return nil
		}
		if clock == nil {
			return errors.New("test clock is required")
		}
		guard, err := NewTestRouteGuard(deps.Env)
		if err != nil {
			return err
		}

		mux.HandleFunc("/api/v1/test/clock/set", guard.Protect(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			var request struct {
				OffsetSeconds *int64  `json:"offset_seconds"`
				FixedNow      *string `json:"fixed_now"`
			}
			decoder := json.NewDecoder(r.Body)
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&request); err != nil {
				writeInvalidClockSetPayload(w, r, "clock")
				return
			}
			if err := decoder.Decode(&struct{}{}); err != io.EOF {
				writeInvalidClockSetPayload(w, r, "clock")
				return
			}

			commandCount := 0
			if request.OffsetSeconds != nil {
				commandCount++
			}
			if request.FixedNow != nil {
				commandCount++
			}
			if commandCount != 1 {
				writeInvalidClockSetPayload(w, r, "clock")
				return
			}

			var next time.Time
			if request.OffsetSeconds != nil {
				next = clock.SetOffset(time.Duration(*request.OffsetSeconds) * time.Second)
			} else {
				fixedNow, err := time.Parse(time.RFC3339Nano, *request.FixedNow)
				if err != nil {
					writeInvalidClockSetPayload(w, r, "fixed_now")
					return
				}
				next = clock.SetFixed(fixedNow)
			}

			_ = WriteSuccess(w, r, http.StatusOK, map[string]any{
				"offset_seconds": clock.OffsetSeconds(),
				"now":            next.Format(time.RFC3339Nano),
			})
		}))
		return nil
	}
}

func writeInvalidClockSetPayload(w http.ResponseWriter, r *http.Request, field string) {
	_ = WriteError(w, r, http.StatusBadRequest, "invalid_mutation_payload", "invalid clock set payload", map[string]any{
		"field": field,
	})
}
