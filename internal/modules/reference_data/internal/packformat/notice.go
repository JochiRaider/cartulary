package packformat

import (
	"bytes"
	"context"
	"errors"
	"io"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// ValidateNoticeStream checks exact producer bytes. It never normalizes signed
// content, and retains at most one fixed-size chunk plus an incomplete NFC
// segment. I/O and cancellation errors remain operational failures.
func ValidateNoticeStream(ctx context.Context, source io.Reader) error {
	buffer := make([]byte, 65536)
	used := 0
	total := int64(0)
	var last, previous byte
	first := true
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		n, readErr := source.Read(buffer[used:])
		used += n
		eof := errors.Is(readErr, io.EOF)
		if readErr != nil && !eof {
			return readErr
		}
		if first && used < 3 && !eof {
			continue
		}
		if first && (used >= 3 || eof) {
			if bytes.HasPrefix(buffer[:used], []byte{0xef, 0xbb, 0xbf}) {
				return fail("content_semantic_invalid")
			}
			first = false
		}
		span, err := norm.NFC.Span(buffer[:used], eof)
		if err != nil && !errors.Is(err, transform.ErrShortSrc) {
			return fail("content_semantic_invalid")
		}
		accepted := buffer[:span]
		if !utf8.Valid(accepted) {
			return fail("content_semantic_invalid")
		}
		for len(accepted) > 0 {
			r, width := utf8.DecodeRune(accepted)
			if unicode.IsControl(r) && r != '\n' && r != '\t' {
				return fail("content_semantic_invalid")
			}
			for _, b := range accepted[:width] {
				previous, last = last, b
				total++
			}
			accepted = accepted[width:]
		}
		used -= span
		copy(buffer, buffer[span:span+used])
		if eof {
			if used != 0 || total == 0 || last != '\n' || (total > 1 && previous == '\n') {
				return fail("content_semantic_invalid")
			}
			return nil
		}
		if used == len(buffer) {
			return fail("content_semantic_invalid")
		}
	}
}
