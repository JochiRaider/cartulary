package httpapi

import (
	"bytes"
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

type testClockSetRequest struct {
	OffsetSeconds *int64
	FixedNow      *string
}

func decodeTestClockSetRequest(r *http.Request) (testClockSetRequest, error) {
	var request testClockSetRequest
	body, err := decodeTestClockObject(r, false)
	if err != nil {
		return request, err
	}
	if len(body) != 1 {
		return request, errors.New("exactly one clock command is required")
	}
	for key, value := range body {
		switch key {
		case "offset_seconds":
			if err := json.Unmarshal(value, &request.OffsetSeconds); err != nil {
				return request, err
			}
			if request.OffsetSeconds == nil {
				return request, errors.New("offset_seconds must be an integer")
			}
		case "fixed_now":
			if err := json.Unmarshal(value, &request.FixedNow); err != nil {
				return request, err
			}
			if request.FixedNow == nil {
				return request, errors.New("fixed_now must be a timestamp")
			}
		default:
			return request, errors.New("unknown clock command")
		}
	}
	return request, nil
}

func validateTestClockResetBody(r *http.Request) error {
	body, err := decodeTestClockObject(r, true)
	if err != nil {
		return err
	}
	if len(body) != 0 {
		return errors.New("reset body must be an empty object")
	}
	return nil
}

func decodeTestClockObject(r *http.Request, allowEmpty bool) (map[string]json.RawMessage, error) {
	if r.Body == nil || r.Body == http.NoBody {
		if allowEmpty {
			return nil, nil
		}
		return nil, errors.New("body is required")
	}
	defer r.Body.Close()
	const maxBodyBytes = 1 << 20
	data, err := io.ReadAll(io.LimitReader(r.Body, maxBodyBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxBodyBytes {
		return nil, errors.New("clock body exceeds its size limit")
	}
	if len(data) == 0 && allowEmpty {
		return nil, nil
	}
	return DecodeStrictJSONObject(bytes.NewReader(data))
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
