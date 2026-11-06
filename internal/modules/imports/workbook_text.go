package imports

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"github.com/JochiRaider/cartulary/internal/modules/tabularingest"
)

// UTF-8 admission belongs to Workbook interpretation, not raw source upload or
// the target-neutral lexer. Validation runs even when preview discards fields.
type workbookTextReader struct {
	source           *bufio.Reader
	remaining        int
	minimum, maximum byte
}

func newWorkbookTextReader(source io.Reader) io.Reader {
	r := bufio.NewReaderSize(source, 64*1024)
	if prefix, _ := r.Peek(3); bytes.Equal(prefix, []byte{0xef, 0xbb, 0xbf}) {
		_, _ = r.Discard(3)
	}
	return &workbookTextReader{source: r}
}

func (r *workbookTextReader) Read(p []byte) (int, error) {
	n, err := r.source.Read(p)
	for _, b := range p[:n] {
		if r.remaining > 0 {
			if b < r.minimum || b > r.maximum {
				return 0, invalidWorkbookUTF8()
			}
			r.remaining--
			r.minimum, r.maximum = 0x80, 0xbf
			continue
		}
		r.minimum, r.maximum = 0x80, 0xbf
		switch {
		case b < 0x80:
		case b >= 0xc2 && b <= 0xdf:
			r.remaining = 1
		case b >= 0xe0 && b <= 0xef:
			r.remaining = 2
			if b == 0xe0 {
				r.minimum = 0xa0
			}
			if b == 0xed {
				r.maximum = 0x9f
			}
		case b >= 0xf0 && b <= 0xf4:
			r.remaining = 3
			if b == 0xf0 {
				r.minimum = 0x90
			}
			if b == 0xf4 {
				r.maximum = 0x8f
			}
		default:
			return 0, invalidWorkbookUTF8()
		}
	}
	if err == io.EOF && r.remaining != 0 {
		return 0, invalidWorkbookUTF8()
	}
	return n, err
}

func invalidWorkbookUTF8() error {
	return fmt.Errorf("%w: invalid_utf8", tabularingest.ErrInvalidDelimited)
}
