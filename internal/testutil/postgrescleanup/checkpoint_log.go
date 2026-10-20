package postgrescleanup

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// CheckpointCompletion is the only retained projection of a PostgreSQL log line.
// Its clock is relative to the proved owned service start, not a drop start.
type CheckpointCompletion struct {
	ServiceOrdinal int   `json:"service_ordinal"`
	ElapsedMS      int64 `json:"elapsed_ms"`
	WriteMS        int64 `json:"write_ms"`
	SyncMS         int64 `json:"sync_ms"`
	Files          int64 `json:"files"`
	LongestSyncMS  int64 `json:"longest_sync_ms"`
	AverageSyncMS  int64 `json:"average_sync_ms"`
}

// Pinned PostgreSQL 18 log grammar, including its default prefix and Docker's
// requested RFC3339 timestamp. Unknown or extended text is never retained.
var checkpointCompletionPattern = regexp.MustCompile(`^(\S+) \d{4}-\d\d-\d\d \d\d:\d\d:\d\d\.\d{3} UTC \[\d+\] LOG:  checkpoint complete: wrote \d+ buffers \(\d+\.\d%\), wrote \d+ SLRU buffers; \d+ WAL file\(s\) added, \d+ removed, \d+ recycled; write=(\d+)\.(\d{3}) s, sync=(\d+)\.(\d{3}) s, total=\d+\.\d{3} s; sync files=(\d+), longest=(\d+)\.(\d{3}) s, average=(\d+)\.(\d{3}) s; distance=\d+ kB, estimate=\d+ kB; lsn=[0-9A-F]+/[0-9A-F]+, redo lsn=[0-9A-F]+/[0-9A-F]+$`)

// ParseCheckpointCompletion discards all raw input, including process identity
// and LSNs. The boolean distinguishes admitted numeric data from unknown format.
func ParseCheckpointCompletion(line string, serviceStart time.Time) (CheckpointCompletion, bool) {
	if len(line) > 4096 || serviceStart.IsZero() {
		return CheckpointCompletion{}, false
	}
	match := checkpointCompletionPattern.FindStringSubmatch(strings.TrimSuffix(line, "\r"))
	if match == nil {
		return CheckpointCompletion{}, false
	}
	stamp, err := time.Parse(time.RFC3339Nano, match[1])
	if err != nil || stamp.Before(serviceStart) || stamp.Sub(serviceStart) > 24*time.Hour {
		return CheckpointCompletion{}, false
	}
	values := make([]int64, 9)
	for i := range values {
		values[i], err = strconv.ParseInt(match[i+2], 10, 64)
		if err != nil || values[i] < 0 || values[i] > 86400000 {
			return CheckpointCompletion{}, false
		}
	}
	milliseconds := func(seconds, fraction int64) int64 { return seconds*1000 + fraction }
	value := CheckpointCompletion{ElapsedMS: stamp.Sub(serviceStart).Milliseconds(), WriteMS: milliseconds(values[0], values[1]), SyncMS: milliseconds(values[2], values[3]), Files: values[4], LongestSyncMS: milliseconds(values[5], values[6]), AverageSyncMS: milliseconds(values[7], values[8])}
	if value.WriteMS > 86400000 || value.SyncMS > 86400000 || value.LongestSyncMS > 86400000 || value.AverageSyncMS > 86400000 {
		return CheckpointCompletion{}, false
	}
	return value, true
}
