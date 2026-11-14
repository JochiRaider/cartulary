package provenancevalue

import (
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Row struct {
	RecordID       uuid.UUID
	SourceIdentity [32]byte
	RowOrdinal     int64
	ColumnOrdinal  int64
	Values         Values
	CreatedAt      time.Time
}

func Identity(raw map[string]any) (uuid.UUID, int64, int64, string, bool) {
	id, ok := raw["record_id"].(string)
	if !ok || !CanonicalUUID(id) {
		return uuid.Nil, 0, 0, "", false
	}
	digest, ok := raw["source_identity_sha256"].(string)
	if !ok || !LowerDigest(digest) {
		return uuid.Nil, 0, 0, "", false
	}
	row, ok := ordinal(raw["source_row_ordinal"])
	if !ok {
		return uuid.Nil, 0, 0, "", false
	}
	column, ok := ordinal(raw["source_column_ordinal"])
	if !ok {
		return uuid.Nil, 0, 0, "", false
	}
	return uuid.MustParse(id), row, column, digest, true
}

func Parse(raw map[string]any) (Row, bool) {
	var row Row
	fields := []string{"record_id", "source_identity_sha256", "source_row_ordinal", "source_column_ordinal", "source_kind", "source_metadata", "source_header", "raw_value", "cell_kind", "created_at"}
	if len(raw) != len(fields) {
		return row, false
	}
	for _, name := range fields {
		if _, ok := raw[name]; !ok {
			return row, false
		}
	}
	id, r, c, digest, ok := Identity(raw)
	if !ok {
		return row, false
	}
	values, ok := ParseValues(raw["source_kind"], raw["source_metadata"], raw["source_header"], raw["raw_value"], raw["cell_kind"])
	if !ok {
		return row, false
	}
	stamp, ok := raw["created_at"].(string)
	if !ok {
		return row, false
	}
	created, err := time.Parse(time.RFC3339Nano, stamp)
	_, offset := created.Zone()
	canonical := created.UTC().Format(time.RFC3339Nano)
	if err != nil || offset != 0 || created.Year() < 1 || created.Year() > 9999 || created.Nanosecond()%1000 != 0 || (stamp != canonical && stamp != strings.TrimSuffix(canonical, "Z")+"+00:00") {
		return row, false
	}
	decoded, _ := hex.DecodeString(digest)
	row.RecordID, row.RowOrdinal, row.ColumnOrdinal, row.Values, row.CreatedAt = id, r, c, values, created.UTC()
	copy(row.SourceIdentity[:], decoded)
	return row, true
}

func (r Row) Coherent() bool { return r.Values.Coherent() && r.SourceIdentity == r.Values.Digest }

func ordinal(raw any) (int64, bool) {
	n, ok := raw.(json.Number)
	if !ok {
		return 0, false
	}
	value, err := n.Int64()
	return value, err == nil && value >= 1 && value <= 2147483647 && strconv.FormatInt(value, 10) == n.String()
}
