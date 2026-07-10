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

const testClockControlSchemaID = "cartulary.test.clock_control.v1"

const (
	testClockModeWall   = "wall"
	testClockModeOffset = "offset"
	testClockModeFixed  = "fixed"
)

type TestClockSnapshot struct {
	Mode          string
	Now           time.Time
	OffsetSeconds int64
	FixedNow      *time.Time
}

type testClockControlResult struct {
	SchemaID      string `json:"schema_id"`
	Mode          string `json:"mode"`
	Now           string `json:"now"`
	OffsetSeconds int64  `json:"offset_seconds"`
	FixedNow      string `json:"fixed_now,omitempty"`
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

func (c *TestClock) Reset() time.Time {
	c.mu.Lock()
	c.offset = 0
	c.fixed = time.Time{}
	c.fixedSet = false
	now := time.Now().UTC()
	c.mu.Unlock()
	return now
}

func (c *TestClock) Snapshot() TestClockSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.fixedSet {
		fixed := c.fixed.UTC()
		return TestClockSnapshot{
			Mode:          testClockModeFixed,
			Now:           fixed,
			OffsetSeconds: 0,
			FixedNow:      &fixed,
		}
	}
	now := time.Now().UTC().Add(c.offset)
	mode := testClockModeWall
	if c.offset != 0 {
		mode = testClockModeOffset
	}
	return TestClockSnapshot{
		Mode:          mode,
		Now:           now,
		OffsetSeconds: int64(c.offset / time.Second),
	}
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

			request, err := decodeTestClockSetRequest(r)
			if err != nil {
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

			snapshot := clock.Snapshot()
			snapshot.Now = next
			_ = WriteSuccess(w, r, http.StatusOK, testClockControlResponse(snapshot))
		}))
		mux.HandleFunc("/api/v1/test/clock/reset", guard.Protect(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			if err := validateTestClockResetBody(r); err != nil {
				writeInvalidClockSetPayload(w, r, "clock")
				return
			}
			next := clock.Reset()
			_ = WriteSuccess(w, r, http.StatusOK, testClockControlResponse(TestClockSnapshot{
				Mode:          testClockModeWall,
				Now:           next,
				OffsetSeconds: 0,
			}))
		}))
		mux.HandleFunc("/api/v1/test/clock/state", guard.Protect(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			_ = WriteSuccess(w, r, http.StatusOK, testClockControlResponse(clock.Snapshot()))
		}))
		return nil
	}
}

func decodeTestClockSetRequest(r *http.Request) (struct {
	OffsetSeconds *int64  `json:"offset_seconds"`
	FixedNow      *string `json:"fixed_now"`
}, error) {
	var request struct {
		OffsetSeconds *int64  `json:"offset_seconds"`
		FixedNow      *string `json:"fixed_now"`
	}
	if r.Body == nil {
		return request, errors.New("body is required")
	}
	defer r.Body.Close()
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return request, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return request, errors.New("body must contain a single JSON object")
	}
	return request, nil
}

func validateTestClockResetBody(r *http.Request) error {
	if r.Body == nil || r.Body == http.NoBody {
		return nil
	}
	defer r.Body.Close()
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	var body struct{}
	if err := decoder.Decode(&body); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("body must contain a single JSON object")
	}
	return nil
}

func testClockControlResponse(snapshot TestClockSnapshot) testClockControlResult {
	result := testClockControlResult{
		SchemaID:      testClockControlSchemaID,
		Mode:          snapshot.Mode,
		Now:           snapshot.Now.UTC().Format(time.RFC3339Nano),
		OffsetSeconds: snapshot.OffsetSeconds,
	}
	if snapshot.FixedNow != nil {
		result.FixedNow = snapshot.FixedNow.UTC().Format(time.RFC3339Nano)
	}
	return result
}

func writeInvalidClockSetPayload(w http.ResponseWriter, r *http.Request, field string) {
	_ = WriteError(w, r, http.StatusBadRequest, "invalid_mutation_payload", "invalid clock set payload", map[string]any{
		"field": field,
	})
}
