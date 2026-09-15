package tabularingest

import (
	"errors"
	"fmt"
	"strings"
)

const MaxClipboardBytes = 8_388_608

var ErrInvalidClipboard = errors.New("invalid clipboard payload")

// DecodeDelimited completes lexical decoding before either lifecycle's mapping
// adapter runs. Imports owns its missing-cell geometry and source byte limits.
func decodeDelimited(text, format string, maxColumns, maxRows int, rectangular bool) ([][]string, error) {
	if text == "" {
		return nil, fmt.Errorf("empty_table")
	}
	normalized, err := normalizedSourceFormat(text, format)
	if err != nil {
		return nil, err
	}
	delimiter := byte('\t')
	if normalized == SourceFormatCSV {
		delimiter = ','
	}
	rows := make([][]string, 0)
	row := make([]string, 0)
	index := 0
	for {
		var value string
		if index < len(text) && text[index] == '"' {
			index++
			start := index
			var cell strings.Builder
			for {
				if index == len(text) {
					return nil, fmt.Errorf("malformed_quotes")
				}
				if text[index] != '"' {
					index++
					continue
				}
				cell.WriteString(text[start:index])
				index++
				if index == len(text) || text[index] != '"' {
					break
				}
				cell.WriteByte('"')
				index++
				start = index
			}
			value = cell.String()
		} else {
			start := index
			for index < len(text) && text[index] != delimiter && text[index] != '\r' && text[index] != '\n' {
				if text[index] == '"' {
					return nil, fmt.Errorf("malformed_quotes")
				}
				index++
			}
			value = text[start:index]
		}
		row = append(row, value)
		if maxColumns > 0 && len(row) > maxColumns {
			return nil, fmt.Errorf("too_many_columns: tabular column count exceeded")
		}
		if index < len(text) && text[index] == delimiter {
			index++
			continue
		}
		if index < len(text) && text[index] != '\r' && text[index] != '\n' {
			return nil, fmt.Errorf("malformed_quotes")
		}
		if rectangular && len(rows) > 0 && len(row) != len(rows[0]) {
			return nil, fmt.Errorf("ragged_rows")
		}
		rows = append(rows, row)
		if maxRows > 0 && len(rows) > maxRows {
			return nil, fmt.Errorf("too_many_rows")
		}
		row = make([]string, 0)
		if index == len(text) {
			break
		}
		if text[index] == '\r' && index+1 < len(text) && text[index+1] == '\n' {
			index++
		}
		index++
		if index == len(text) {
			break
		}
	}
	return rows, nil
}
