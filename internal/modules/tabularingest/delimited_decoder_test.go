package tabularingest_test

import (
	"context"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/tabularingest"
)

type oneByteReader struct{ io.Reader }

func (r oneByteReader) Read(p []byte) (int, error) { return r.Reader.Read(p[:min(len(p), 1)]) }

func TestDelimitedDecoderIncrementalContract(t *testing.T) {
	for _, tc := range []struct {
		text string
		rows [][]string
	}{
		{"a,b\r1\r\n2,\n\n", [][]string{{"a", "b"}, {"1"}, {"2", ""}, {""}}},
		{"\"a,\"\"b\r\nc\",d\nlast,", [][]string{{"a,\"b\r\nc", "d"}, {"last", ""}}},
		{"\xff,\"\"", [][]string{{"\xff", ""}}},
	} {
		for _, retain := range []bool{true, false} {
			decoder, err := tabularingest.NewDelimitedDecoder(t.Context(), oneByteReader{strings.NewReader(tc.text)}, "csv", 0)
			if err != nil {
				t.Fatal(err)
			}
			for _, expected := range tc.rows {
				record, err := decoder.Next(retain)
				if err != nil || record.FieldCount != len(expected) {
					t.Fatalf("record %#v: %v", record, err)
				}
				if retain && !reflect.DeepEqual(record.Fields, expected) {
					t.Fatalf("got %#v want %#v", record.Fields, expected)
				}
				if !retain && record.Fields != nil {
					t.Fatal("discarded record retained fields")
				}
			}
			if _, err := decoder.Next(retain); err != io.EOF {
				t.Fatalf("EOF: %v", err)
			}
		}
	}
	for _, source := range []string{"a\"b", "\"unterminated", "\"closed\"suffix"} {
		decoder, err := tabularingest.NewDelimitedDecoder(t.Context(), oneByteReader{strings.NewReader(source)}, "csv", 0)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = decoder.Next(false); !errors.Is(err, tabularingest.ErrInvalidDelimited) {
			t.Fatalf("%q: %v", source, err)
		}
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	decoder, err := tabularingest.NewDelimitedDecoder(ctx, strings.NewReader("a"), "csv", 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = decoder.Next(true); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation: %v", err)
	}
}
