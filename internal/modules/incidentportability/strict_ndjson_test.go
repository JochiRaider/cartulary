package incidentportability

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync/atomic"
	"testing"
)

func TestNDJSONStreamRejectsAmbiguousOrLossyRecords_Unit(t *testing.T) {
	for _, tc := range []struct {
		name, input, kind string
		ordinal           int
	}{
		{"blank record", "{}\n\n", "blank_line", 2},
		{"duplicate field", "{\"x\":1,\"x\":2}", "duplicate_member", 1},
		{"nested duplicate", "{\"x\":[{\"k\":1,\"k\":2}]}", "duplicate_member", 1},
		{"escaped duplicate", "{\"x\":1,\"\\u0078\":2}", "duplicate_member", 1},
		{"trailing value", "{} {}", "multiple_values", 1},
		{"array record", "[]", "row_not_object", 1},
		{"null record", "null", "row_not_object", 1},
		{"invalid UTF-8", "{\"x\":\"\xff\"}", "malformed_json", 1},
		{"unpaired high surrogate", "{\"x\":\"\\ud800\"}", "malformed_json", 1},
		{"unpaired low surrogate", "{\"x\":\"\\udc00\"}", "malformed_json", 1},
		{"line bound", "{\"x\":\"" + strings.Repeat("a", 128) + "\"}", "line_too_long", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			count := 0
			err := WalkNDJSON(context.Background(), strings.NewReader(tc.input), "data/source.ndjson", 128, func(map[string]any) error { count++; return nil })
			var failure *StrictNDJSONDecodeError
			if !errors.As(err, &failure) || failure.Kind != tc.kind || failure.LogicalPath != "data/source.ndjson" {
				t.Fatal(err, failure)
			}
			if failure.RowOrdinal != tc.ordinal || count != tc.ordinal-1 {
				t.Fatalf("row ordinal=%d visits=%d want ordinal=%d visits=%d", failure.RowOrdinal, count, tc.ordinal, tc.ordinal-1)
			}
		})
	}
}

func TestNDJSONStreamBoundsReadsAndPreservesValues_Unit(t *testing.T) {
	line := `{"text":"` + strings.Repeat("x", 1024) + `\t\r\n雪\ud83d\ude80","number":18446744073709551615}` + "\n"
	const records = 70000
	if int64(len(line))*records <= 64<<20 {
		t.Fatal("bulk fixture must exceed the control ceiling")
	}
	wantText := strings.Repeat("x", 1024) + "\t\r\n雪🚀"
	source := &repeatedRecordReader{line: line, remaining: records}
	count := 0
	err := WalkNDJSON(context.Background(), source, "data/source.ndjson", 2048, func(row map[string]any) error {
		count++
		if len(row) != 2 || row["text"] != wantText {
			return errors.New("complete source text or key set changed")
		}
		if row["number"] != json.Number("18446744073709551615") {
			return errors.New("source values changed")
		}
		return nil
	})
	if err != nil || count != records || source.maximumRead > 32<<10 {
		t.Fatal(count, source.maximumRead, err)
	}
	t.Run("cancellation stops before another owner call", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		count := 0
		err := WalkNDJSON(ctx, strings.NewReader("{}\n{}\n"), "data/source.ndjson", 128, func(map[string]any) error { count++; cancel(); return nil })
		if !errors.Is(err, context.Canceled) || count != 1 {
			t.Fatal(count, err)
		}
	})
	t.Run("reader failure is not EOF", func(t *testing.T) {
		count := 0
		err := WalkNDJSON(context.Background(), io.MultiReader(strings.NewReader("{}"), brokenRecordReader{}), "data/source.ndjson", 128, func(map[string]any) error { count++; return nil })
		if !errors.Is(err, io.ErrUnexpectedEOF) || count != 0 {
			t.Fatal(count, err)
		}
	})
	t.Run("exact record ceiling admits final newline or EOF", func(t *testing.T) {
		for _, input := range []string{"{}", "{}\n"} {
			calls := 0
			if err := WalkNDJSON(context.Background(), strings.NewReader(input), "data/source.ndjson", 2, func(row map[string]any) error {
				calls++
				if len(row) != 0 {
					return errors.New("empty source object changed")
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			if calls != 1 {
				t.Fatalf("input=%q visits=%d want1", input, calls)
			}
		}
	})
}

type repeatedRecordReader struct {
	line                           string
	remaining, offset, maximumRead int
}

func (r *repeatedRecordReader) Read(p []byte) (int, error) {
	r.maximumRead = max(r.maximumRead, len(p))
	if r.remaining == 0 {
		return 0, io.EOF
	}
	n := copy(p, r.line[r.offset:])
	r.offset += n
	if r.offset == len(r.line) {
		r.offset = 0
		r.remaining--
	}
	return n, nil
}

type brokenRecordReader struct{}

func (brokenRecordReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

func TestMemberProducerIsIncrementalAndJoinedOnFailure_Unit(t *testing.T) {
	const total = 65 << 20
	emitted := int64(0)
	sink := memberWriterFunc(func(ctx context.Context, path string, source io.Reader) (Member, error) {
		var first [1]byte
		if _, err := io.ReadFull(source, first[:]); err != nil {
			return Member{}, err
		}
		if atomic.LoadInt64(&emitted) >= total {
			return Member{}, errors.New("producer buffered the complete member")
		}
		n, err := io.CopyBuffer(io.Discard, source, make([]byte, 32<<10))
		if n+1 != total {
			return Member{}, fmt.Errorf("member byte count = %d", n+1)
		}
		return Member{Path: path, ByteSize: n + 1}, err
	})
	err := WriteMemberFrom(context.Background(), sink, "data/large.ndjson", func(ctx context.Context, output io.Writer) error {
		block := make([]byte, 32<<10)
		for sent := 0; sent < total; sent += len(block) {
			if _, err := output.Write(block); err != nil {
				return err
			}
			atomic.AddInt64(&emitted, int64(len(block)))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	failure := errors.New("storage stopped reading")
	joined := atomic.Bool{}
	sink = memberWriterFunc(func(context.Context, string, io.Reader) (Member, error) { return Member{}, failure })
	err = WriteMemberFrom(context.Background(), sink, "data/failure.ndjson", func(ctx context.Context, output io.Writer) error {
		defer joined.Store(true)
		<-ctx.Done()
		return ctx.Err()
	})
	if !errors.Is(err, failure) || !joined.Load() {
		t.Fatalf("storage failure=%v producer joined=%v", err, joined.Load())
	}
	sourceFailure := errors.New("source failed after a partial member")
	sink = memberWriterFunc(func(ctx context.Context, path string, source io.Reader) (Member, error) {
		_, err := io.Copy(io.Discard, source)
		return Member{}, err
	})
	err = WriteMemberFrom(context.Background(), sink, "data/partial.ndjson", func(ctx context.Context, output io.Writer) error {
		if _, err := io.WriteString(output, "{}\n"); err != nil {
			return err
		}
		return sourceFailure
	})
	if !errors.Is(err, sourceFailure) {
		t.Fatalf("source failure = %v", err)
	}
}

func TestNDJSONRecordWriterEnforcesLimitsAndShortWrites_Unit(t *testing.T) {
	row := map[string]any{"number": json.Number("9007199254740993"), "text": "<literal>\t\nλ"}
	var output strings.Builder
	if err := WriteNDJSONRecord(&output, row); err != nil {
		t.Fatal(err)
	}
	want := "{\"number\":9007199254740993,\"text\":\"\\u003cliteral\\u003e\\t\\nλ\"}\n"
	if output.String() != want {
		t.Fatalf("record bytes=%q want=%q", output.String(), want)
	}
	if err := WriteNDJSONRecord(shortRecordWriter{}, row); !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("short write = %v", err)
	}
	output.Reset()
	err := WriteNDJSONRecord(&output, map[string]any{"value": strings.Repeat("a", MaxNDJSONLogicalLineBytes)})
	var failure *StrictNDJSONDecodeError
	if !errors.As(err, &failure) || failure.Kind != "line_too_long" || output.Len() != 0 {
		t.Fatalf("oversized row: %v bytes=%d", err, output.Len())
	}
}

type memberWriterFunc func(context.Context, string, io.Reader) (Member, error)

func (f memberWriterFunc) WriteMember(ctx context.Context, path string, source io.Reader) (Member, error) {
	return f(ctx, path, source)
}

type shortRecordWriter struct{}

func (shortRecordWriter) Write(p []byte) (int, error) { return len(p) - 1, nil }
