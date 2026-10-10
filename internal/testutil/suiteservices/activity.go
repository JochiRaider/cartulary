package suiteservices

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/rootedfs"
)

// MeasureMigration measures only the supplied initialization/apply operation.
// The private beginning survives fatal exits; diagnostic loss never changes err.
func MeasureMigration(ctx context.Context, operation func() error) (time.Duration, error) {
	return measureMigration(ctx, operation, time.Now, migrationObservation)
}

func measureMigration(ctx context.Context, operation func() error, now func() time.Time, observe func() func(time.Duration, string)) (time.Duration, error) {
	finish := observe()
	started := now()
	err := operation()
	duration := now().Sub(started)
	outcome := "passed"
	if err != nil {
		outcome = "failed"
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
			outcome = "cancelled"
		}
	}
	finish(duration, outcome)
	return duration, err
}

func activityID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes[:])
}

var activityClockID = activityID()

func migrationObservation() func(time.Duration, string) {
	noop := func(time.Duration, string) {}
	if LookupEnvValue(nil, "CARTULARY_HARNESS_EXECUTION_ENABLED") != "1" {
		return noop
	}
	runtimeRoot, err := resolveSuiteRuntimeRoot(nil)
	if err != nil {
		return noop
	}
	root, err := rootedfs.Open(filepath.Join(runtimeRoot, "execution-observations"))
	if err != nil {
		return noop
	}
	defer func() { _ = root.Close() }()
	read := func(name string, limit int64) ([]byte, error) {
		value, metadata, err := root.ReadRegular(rootedfs.MustParseReference(name), limit)
		if err == nil && metadata.Mode.Perm() != 0o600 {
			return nil, errors.New("unsafe private observation file")
		}
		return value, err
	}
	isOpen := func() bool { value, err := read("open", 8); return err == nil && string(value) == "1\n" }
	if !isOpen() {
		return noop
	}
	var limits struct {
		Records int `json:"maximum_execution_records"`
		Bytes   int `json:"maximum_record_bytes"`
	}
	value, err := read("limits.json", 65536)
	if err != nil || json.Unmarshal(value, &limits) != nil || limits.Records < 1 || limits.Records > 4096 || limits.Bytes != 65536 {
		return noop
	}
	var launch struct {
		Run        string  `json:"run_id"`
		Invocation *string `json:"invocation_id"`
		Unit       *string `json:"unit_id"`
	}
	if json.Unmarshal([]byte(LookupEnvValue(nil, "CARTULARY_HARNESS_LAUNCH_CONTEXT")), &launch) != nil || launch.Run != LookupEnvValue(nil, testRunIDEnv) {
		return noop
	}
	id := activityID()
	if id == "" {
		return noop
	}
	var allocation any
	if ref := LookupEnvValue(nil, "CARTULARY_HARNESS_ALLOCATION_REF"); ref != "" {
		allocation = ref
	}
	record := map[string]any{"kind": "activity", "ref": "activity:" + id, "activity": "migration", "unit_id": launch.Unit,
		"invocation_id": launch.Invocation, "allocation_ref": allocation, "outcome": "incomplete", "clock": "unavailable",
		"clock_identity": nil, "resolution_ms": nil, "duration_ms": nil, "availability": "not_observed"}
	write := func(writer io.Writer) error { return json.NewEncoder(writer).Encode(record) }
	var slot rootedfs.Reference
	for index := 0; index < limits.Records; index++ {
		candidate := rootedfs.MustParseReference(fmt.Sprintf("slot-%d", index))
		if err := root.CreateExclusive(context.Background(), candidate, write); err == nil {
			slot = candidate
			break
		} else if !errors.Is(err, os.ErrExist) {
			return noop
		}
	}
	if slot.String() == "" {
		_ = root.CreateExclusive(context.Background(), rootedfs.MustParseReference("truncated"), func(w io.Writer) error { _, err := io.WriteString(w, "1\n"); return err })
		return noop
	}
	var once sync.Once
	return func(duration time.Duration, outcome string) {
		once.Do(func() {
			root, err = rootedfs.Open(filepath.Join(runtimeRoot, "execution-observations"))
			if err != nil {
				return
			}
			defer root.Close()
			if !isOpen() {
				return
			}
			previous, err := read(slot.String(), 65536)
			var initial map[string]any
			if err != nil || json.Unmarshal(previous, &initial) != nil || initial["ref"] != record["ref"] || initial["outcome"] != "incomplete" {
				return
			}
			record["outcome"] = outcome
			if duration >= 0 && activityClockID != "" {
				record["clock"], record["clock_identity"], record["resolution_ms"] = "go_monotonic", "clock:"+activityClockID, 0.000001
				record["duration_ms"], record["availability"] = float64(duration)/float64(time.Millisecond), "available"
			} else {
				record["availability"] = "unavailable"
			}
			_ = root.AtomicReplace(context.Background(), slot, write)
		})
	}
}
