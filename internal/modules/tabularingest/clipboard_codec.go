package tabularingest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
)

const MaxClipboardBytes = 8_388_608

var ErrInvalidClipboard = errors.New("invalid clipboard payload")

// Clipboard lexical validation uses the same incremental decoder as Imports;
// its rectangular geometry and limits remain specific to this adapter.
func decodeDelimited(text, format string, maxColumns, maxRows int, rectangular bool) ([][]string, error) {
	if text == "" {
		return nil, fmt.Errorf("empty_table")
	}
	decoder, err := NewDelimitedDecoder(context.Background(), strings.NewReader(text), format, maxColumns)
	if err != nil {
		return nil, err
	}
	rows := make([][]string, 0)
	for {
		record, err := decoder.Next(true)
		if err == io.EOF {
			if len(rows) == 0 {
				return nil, fmt.Errorf("empty_table")
			}
			return rows, nil
		}
		if err != nil {
			return nil, err
		}
		if rectangular && len(rows) > 0 && record.FieldCount != len(rows[0]) {
			return nil, fmt.Errorf("ragged_rows")
		}
		rows = append(rows, record.Fields)
		if maxRows > 0 && len(rows) > maxRows {
			return nil, fmt.Errorf("too_many_rows")
		}
	}
}
