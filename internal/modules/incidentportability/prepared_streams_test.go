package incidentportability

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"sort"
	"strings"
	"testing"
)

// memoryMembers is a bounded in-memory member store. sealVerified marks it as a
// reader whose streams verify the seal themselves.
type memoryMembers struct {
	files        map[string][]byte
	opens        int
	sealVerified bool
}

func newMemoryMembers() *memoryMembers { return &memoryMembers{files: map[string][]byte{}} }

func (m *memoryMembers) describe(path string) Member {
	digest := sha256.Sum256(m.files[path])
	return Member{Path: path, ByteSize: int64(len(m.files[path])), SHA256: hex.EncodeToString(digest[:])}
}
func (m *memoryMembers) Members() []Member {
	paths := make([]string, 0, len(m.files))
	for path := range m.files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	members := make([]Member, len(paths))
	for index, path := range paths {
		members[index] = m.describe(path)
	}
	return members
}
func (m *memoryMembers) OpenMember(_ context.Context, path string) (io.ReadCloser, Member, error) {
	raw, ok := m.files[path]
	if !ok {
		return nil, Member{}, &VerificationFailure{ReasonCode: "missing_required_file"}
	}
	m.opens++
	return io.NopCloser(bytes.NewReader(raw)), m.describe(path), nil
}
func (m *memoryMembers) WriteMember(ctx context.Context, path string, source io.Reader) (Member, error) {
	if _, exists := m.files[path]; exists {
		return Member{}, errors.New("duplicate member")
	}
	raw, err := io.ReadAll(source)
	if err != nil {
		return Member{}, err
	}
	m.files[path] = raw
	return m.describe(path), ctx.Err()
}
func (m *memoryMembers) VerifiesSealedContent() bool { return m.sealVerified }

type typedTestRow struct {
	ID    string          `json:"id"`
	Note  *string         `json:"note"`
	Inner json.RawMessage `json:"inner"`
}

// checkTypedWalk: one strict scan and one typed decode per line, with the
// WalkNDJSON framing and error kinds and a closed top-level shape.
func checkTypedWalk(t *testing.T) {
	shape := RecordShape{Required: []string{"id", "note", "inner"}, Nullable: []string{"note"}}
	run := func(input string) ([]string, error) {
		members := newMemoryMembers()
		members.files["data/rows.ndjson"] = []byte(input)
		var ids []string
		err := WalkTyped(t.Context(), members, "data/rows.ndjson", shape, func(row *typedTestRow, line []byte) error {
			if !json.Valid(line) {
				return errors.New("visitor received a non-JSON line")
			}
			ids = append(ids, row.ID)
			return nil
		})
		return ids, err
	}
	ids, err := run("{\"id\":\"a\",\"note\":null,\"inner\":{}}\n{\"inner\":[1],\"note\":\"x\",\"id\":\"b\"}\n")
	if err != nil || strings.Join(ids, ",") != "a,b" {
		t.Fatalf("typed rows = %v %v", ids, err)
	}
	for _, tc := range []struct{ name, input, kind string }{
		{"duplicate nested member", "{\"id\":\"a\",\"note\":null,\"inner\":{\"k\":1,\"k\":2}}\n", "duplicate_member"},
		{"blank line", "{\"id\":\"a\",\"note\":null,\"inner\":{}}\n\n", "blank_line"},
		{"trailing value", "{\"id\":\"a\",\"note\":null,\"inner\":{}} {}\n", "multiple_values"},
		{"invalid UTF-8", "{\"id\":\"\xff\",\"note\":null,\"inner\":{}}\n", "malformed_json"},
	} {
		_, err := run(tc.input)
		var failure *StrictNDJSONDecodeError
		if !errors.As(err, &failure) || failure.Kind != tc.kind {
			t.Fatalf("%s: %v", tc.name, err)
		}
	}
	for name, input := range map[string]string{
		"missing member":    "{\"id\":\"a\",\"note\":null}\n",
		"unknown member":    "{\"id\":\"a\",\"note\":null,\"inner\":{},\"x\":1}\n",
		"forbidden null":    "{\"id\":null,\"note\":null,\"inner\":{}}\n",
		"wrong member type": "{\"id\":7,\"note\":null,\"inner\":{}}\n",
	} {
		_, err := run(input)
		var shapeErr *RecordShapeError
		if !errors.As(err, &shapeErr) || shapeErr.RowOrdinal != 1 {
			t.Fatalf("%s: %v", name, err)
		}
	}
}

// sealVerifiedTestRead returns the admitted descriptor with substituted bytes.
// sealVerified declares that its streams verify the seal themselves.
type sealVerifiedTestRead struct {
	testMemberRead
	sealVerified bool
}

func (s *sealVerifiedTestRead) VerifiesSealedContent() bool { return s.sealVerified }

// checkSealVerifiedReaders: a reader that already verifies the immutable seal
// is not hashed a second time; any other reader still is.
func checkSealVerifiedReaders(t *testing.T) {
	raw := "{}\n"
	digest := sha256.Sum256([]byte(raw))
	expected := Member{Path: "data/rows.ndjson", ByteSize: int64(len(raw)), SHA256: hex.EncodeToString(digest[:])}
	for _, verified := range []bool{false, true} {
		members := &sealVerifiedTestRead{testMemberRead: testMemberRead{member: expected, bytes: "[]\n"}, sealVerified: verified}
		reader, err := OpenVerifiedMember(t.Context(), members, expected)
		if err != nil {
			t.Fatal(err)
		}
		_, err = io.ReadAll(reader)
		var failure *VerificationFailure
		hashed := errors.As(err, &failure) && failure.ReasonCode == "checksum_mismatch"
		if hashed == verified {
			t.Fatalf("seal-verified=%v rehashed=%v (%v)", verified, hashed, err)
		}
		bound, err := BindMembers(members, []string{"data/rows.ndjson"})
		if err != nil {
			t.Fatal(err)
		}
		if bound.VerifiesSealedContent() != verified {
			t.Fatal("bound members changed seal verification")
		}
	}
}

type stagedTestRow struct {
	ID    string
	Count int64
}
type stagedTestCodec struct{}

func (stagedTestCodec) Encode(r stagedTestRow) ([]byte, error) {
	return json.Marshal([]any{r.ID, r.Count})
}
func (stagedTestCodec) Decode(line []byte) (stagedTestRow, error) {
	var raw []json.RawMessage
	var r stagedTestRow
	if err := json.Unmarshal(line, &raw); err != nil || len(raw) != 2 {
		return r, errors.New("bad staged row")
	}
	if json.Unmarshal(raw[0], &r.ID) != nil || json.Unmarshal(raw[1], &r.Count) != nil {
		return r, errors.New("bad staged row")
	}
	return r, nil
}
func (stagedTestCodec) Columns() []string                     { return []string{"id", "count"} }
func (stagedTestCodec) Values(r stagedTestRow) ([]any, error) { return []any{r.ID, r.Count}, nil }

// copyRows drives an opened stream exactly as a COPY statement does and
// reports the acknowledged count to Finish.
func copyRows(t *testing.T, members MemberReader, stream PreparedRowStream) ([][]any, error) {
	t.Helper()
	source, err := OpenPreparedCopy(t.Context(), members, stream, stagedTestCodec{})
	if err != nil {
		return nil, err
	}
	var rows [][]any
	for source.Next() {
		values, err := source.Values()
		if err != nil {
			return nil, err
		}
		rows = append(rows, values)
	}
	return rows, source.Finish(int64(len(rows)), source.Err())
}

// checkPreparedRowStreams: staged write-ready rows round-trip through one
// sealed member into one COPY, and any count or byte mismatch aborts it.
func checkPreparedRowStreams(t *testing.T) {
	members := newMemoryMembers()
	stream, err := StagePreparedRows(t.Context(), members, "prepared/test/rows.ndjson", stagedTestCodec{}, func(emit func(stagedTestRow) error) error {
		for index := range 300 {
			if err := emit(stagedTestRow{ID: "row-" + strings.Repeat("x", index%7), Count: int64(index)}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil || stream.Rows != 300 || stream.Member != members.describe("prepared/test/rows.ndjson") {
		t.Fatalf("staged stream = %+v %v", stream, err)
	}
	rows, err := copyRows(t, members, stream)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 300 || rows[299][1] != int64(299) {
		t.Fatalf("copied %d rows", len(rows))
	}
	if rows, err := copyRows(t, members, PreparedRowStream{}); err != nil || len(rows) != 0 {
		t.Fatal("empty stream", rows, err)
	}
	overstated := stream
	overstated.Rows++
	if _, err := copyRows(t, members, overstated); !errors.Is(err, ErrPreparedStream) {
		t.Fatal("count mismatch was not detected", err)
	}
	changed := stream
	changed.Member.SHA256 = strings.Repeat("e", 64)
	if _, err := copyRows(t, members, changed); !errors.Is(err, ErrPreparedStream) {
		t.Fatal("descriptor mismatch was not detected", err)
	}
	read := 0
	if err := ReadPreparedRows(t.Context(), members, stream, stagedTestCodec{}, func(stagedTestRow) error { read++; return nil }); err != nil || read != 300 {
		t.Fatal("read back", read, err)
	}
}

// checkAttributionStreams: attribution is staged once, grouped, admitted at
// staging and expanded in order with its per-table aggregates.
func checkAttributionStreams(t *testing.T) {
	members := newMemoryMembers()
	admitted := func(actor string) bool { return actor == "actor-a" || actor == "actor-b" }
	var want []AttributionRow
	for index := range 600 {
		actor := "actor-a"
		if index >= 400 {
			actor = "actor-b"
		}
		table := "owner_rows"
		if index < 3 {
			table = "owner_tables"
		}
		want = append(want, AttributionRow{SourceTable: table, SourceRowID: "row-" + strings.Repeat("y", index%5) + string(rune('a'+index%26)), SourceColumn: "created_by_user_id", SourceActorID: actor})
	}
	stream, err := StageAttributions(t.Context(), members, "prepared/test/attribution.ndjson", admitted, func(emit func(AttributionRow) error) error {
		for _, row := range want {
			if err := emit(row); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil || stream.Rows != 600 || stream.Tables["owner_tables"] != 3 || stream.Tables["owner_rows"] != 597 {
		t.Fatalf("attribution stream = %+v %v", stream, err)
	}
	if lines := bytes.Count(members.files["prepared/test/attribution.ndjson"], []byte("\n")); lines >= 600 || lines < 4 {
		t.Fatalf("attribution lines = %d, want grouped lines", lines)
	}
	reader := OpenAttributions(t.Context(), members, stream)
	var got []AttributionRow
	for reader.Next() {
		got = append(got, reader.Row())
	}
	if err := errors.Join(reader.Err(), reader.Close()); err != nil || reader.Count() != 600 || reader.Expected() != 600 {
		t.Fatal("expanded attribution", reader.Count(), err)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("attribution row %d = %+v want %+v", index, got[index], want[index])
		}
	}
	if tables := AttributionTables(stream); strings.Join(tables, ",") != "owner_rows,owner_tables" {
		t.Fatal("attribution tables", tables)
	}
	_, err = StageAttributions(t.Context(), newMemoryMembers(), "prepared/test/rejected.ndjson", admitted, func(emit func(AttributionRow) error) error {
		return emit(AttributionRow{SourceTable: "owner_rows", SourceRowID: "row", SourceColumn: "created_by_user_id", SourceActorID: "unknown"})
	})
	if !errors.Is(err, ErrAttributionStream) {
		t.Fatal("unadmitted actor was staged", err)
	}
	understated := stream
	understated.Rows--
	reader = OpenAttributions(t.Context(), members, understated)
	for reader.Next() {
	}
	if !errors.Is(errors.Join(reader.Err(), reader.Close()), ErrPreparedStream) {
		t.Fatal("attribution count mismatch was not detected")
	}
}

// checkPreparedVerdicts: a deferred bundle-only verdict is reported at its
// rank, independent of the order in which preparation observed failures.
func checkPreparedVerdicts(t *testing.T) {
	var verdicts PreparedVerdicts
	first, second := errors.New("b"), errors.New("a")
	verdicts.Record(2, "row-2", first)
	verdicts.Record(2, "row-1", second)
	verdicts.Record(2, "row-3", errors.New("c"))
	if verdicts.At(2) != second || verdicts.At(1) != nil || verdicts.Empty() {
		t.Fatal("deferred verdict selection depends on observation order")
	}
}

// checkSortedKeyRuns: keys staged across bounded runs merge into their distinct
// union in byte order, and a tampered run is rejected by its seal.
func checkSortedKeyRuns(t *testing.T) {
	members := newMemoryMembers()
	stager := NewSortedKeyStager(t.Context(), members, "prepared/keys/", 3)
	for _, key := range []string{"d", "b", "b", "a", "e", "c", "a", "f", "d"} {
		if err := stager.Add(key); err != nil {
			t.Fatal(err)
		}
	}
	runs, err := stager.Finish()
	if err != nil || len(runs.Runs) != 3 || runs.Keys() != 8 {
		t.Fatalf("runs=%+v err=%v", runs, err)
	}
	var merged []string
	if err := MergeSortedKeys(t.Context(), members, runs, func(key string) error {
		merged = append(merged, key)
		return nil
	}); err != nil || strings.Join(merged, "") != "abcdef" {
		t.Fatalf("merged=%v err=%v", merged, err)
	}
	if err := stager.Add("x\ny"); err == nil {
		t.Fatal("multi-line key was staged")
	}
	if runs, err := NewSortedKeyStager(t.Context(), members, "prepared/empty/", 3).Finish(); err != nil || len(runs.Runs) != 0 {
		t.Fatalf("empty runs=%+v err=%v", runs, err)
	}
	path := runs.Runs[0].Member.Path
	members.files[path] = []byte("z\n")
	if err := MergeSortedKeys(t.Context(), members, runs, func(string) error { return nil }); !errors.Is(err, ErrPreparedStream) {
		t.Fatalf("tampered run err=%v", err)
	}
}

// checkEmptyStreamsStageNothing: a stream with no rows creates no staged
// object (Core 01 REQ-01-641). Its descriptor is the zero member with zero
// rows and per-table counts, and readers and COPYs of it open nothing.
func checkEmptyStreamsStageNothing(t *testing.T) {
	members := newMemoryMembers()
	rows, rowsErr := StagePreparedRows(t.Context(), members, "prepared/test/empty.rows", stagedTestCodec{}, func(func(stagedTestRow) error) error { return nil })
	attribution, attributionErr := StageAttributions(t.Context(), members, "prepared/test/empty.attribution", func(string) bool { return true }, func(func(AttributionRow) error) error { return nil })
	runs, runsErr := NewSortedKeyStager(t.Context(), members, "prepared/test/keys/", 4).Finish()
	if err := errors.Join(rowsErr, attributionErr, runsErr); err != nil {
		t.Fatal(err)
	}
	if len(members.files) != 0 {
		t.Fatalf("empty streams staged %d members", len(members.files))
	}
	if rows != (PreparedRowStream{}) || attribution.PreparedRowStream != (PreparedRowStream{}) || len(attribution.Tables) != 0 || runs.Keys() != 0 {
		t.Fatalf("empty descriptors = %+v %+v %+v", rows, attribution, runs)
	}
	read := 0
	if err := ReadPreparedRows(t.Context(), members, rows, stagedTestCodec{}, func(stagedTestRow) error { read++; return nil }); err != nil || read != 0 {
		t.Fatal("empty read", read, err)
	}
	if copied, err := copyRows(t, members, rows); err != nil || len(copied) != 0 {
		t.Fatal("empty copy", copied, err)
	}
	reader := OpenAttributions(t.Context(), members, attribution)
	if reader.Next() || reader.Err() != nil || reader.Expected() != 0 || reader.Close() != nil {
		t.Fatal("empty attribution", reader.Err())
	}
	if err := MergeSortedKeys(t.Context(), members, runs, func(string) error { return errors.New("key from no runs") }); err != nil {
		t.Fatal(err)
	}
	if members.opens != 0 {
		t.Fatalf("empty streams opened %d members", members.opens)
	}
	// A descriptor naming a member is not empty: zero rows cannot bind bytes.
	named := PreparedRowStream{Member: Member{Path: "prepared/test/none.rows", SHA256: strings.Repeat("0", 64)}}
	if err := ReadPreparedRows(t.Context(), members, named, stagedTestCodec{}, func(stagedTestRow) error { return nil }); !errors.Is(err, ErrPreparedStream) {
		t.Fatal("a zero-row stream naming a member was admitted", err)
	}
	if _, err := copyRows(t, members, named); !errors.Is(err, ErrPreparedStream) {
		t.Fatal("a zero-row COPY naming a member was admitted", err)
	}
}
