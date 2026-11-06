package incidentportability

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"

	"github.com/JochiRaider/cartulary/internal/platform/strictjson"
)

const MaxNDJSONLogicalLineBytes = 16 * 1024 * 1024

type StrictNDJSONDecodeError struct {
	LogicalPath string
	RowOrdinal  int
	Kind        string
}

func (e *StrictNDJSONDecodeError) Error() string {
	return "strict NDJSON decode failed at admitted path and row"
}

// WalkNDJSON bounds allocation by one owner's admitted record, independently of
// member size. It rejects invalid Unicode, duplicate fields at every depth,
// blank records and trailing JSON values before calling the owner. The visitor
// must not retain rows when processing bulk members.
func WalkNDJSON(ctx context.Context, source io.Reader, logicalPath string, maximumRecordBytes int, visit func(map[string]any) error) error {
	if source == nil || visit == nil || maximumRecordBytes < 1 || maximumRecordBytes > MaxNDJSONLogicalLineBytes {
		return strictNDJSONError(logicalPath, 0, "invalid_record_limit")
	}
	reader := bufio.NewReaderSize(source, 32<<10)
	for ordinal := 1; ; ordinal++ {
		var line []byte
		for {
			if err := ctx.Err(); err != nil {
				return err
			}
			fragment, err := reader.ReadSlice('\n')
			final := len(fragment) > 0 && fragment[len(fragment)-1] == '\n'
			if final {
				fragment = fragment[:len(fragment)-1]
			}
			if len(fragment) > maximumRecordBytes-len(line) {
				return strictNDJSONError(logicalPath, ordinal, "line_too_long")
			}
			line = append(line, fragment...)
			if errors.Is(err, bufio.ErrBufferFull) {
				continue
			}
			if err != nil && !errors.Is(err, io.EOF) {
				return err
			}
			if errors.Is(err, io.EOF) && len(line) == 0 {
				return nil
			}
			break
		}
		if len(bytes.TrimSpace(line)) == 0 {
			return strictNDJSONError(logicalPath, ordinal, "blank_line")
		}
		_, err := strictjson.DecodeBoundedUnicodeObject(bytes.NewReader(line), int64(maximumRecordBytes))
		if err != nil {
			kind := "malformed_json"
			switch {
			case errors.Is(err, strictjson.ErrDuplicateMember):
				kind = "duplicate_member"
			case errors.Is(err, strictjson.ErrNotObject):
				kind = "row_not_object"
			case errors.Is(err, strictjson.ErrTrailingData):
				kind = "multiple_values"
			}
			return strictNDJSONError(logicalPath, ordinal, kind)
		}
		var row map[string]any
		decoder := json.NewDecoder(bytes.NewReader(line))
		decoder.UseNumber()
		if err := decoder.Decode(&row); err != nil {
			return strictNDJSONError(logicalPath, ordinal, "malformed_json")
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := visit(row); err != nil {
			return err
		}
	}
}

// DecodeStrictNDJSONObjects is for owners whose admitted input is already a
// bounded in-memory document. Bulk member consumers use WalkNDJSON directly.
func DecodeStrictNDJSONObjects(payload []byte, logicalPath string) ([]map[string]any, error) {
	rows := []map[string]any{}
	err := WalkNDJSON(context.Background(), bytes.NewReader(payload), logicalPath, MaxNDJSONLogicalLineBytes, func(row map[string]any) error {
		rows = append(rows, row)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func strictNDJSONError(logicalPath string, rowOrdinal int, kind string) *StrictNDJSONDecodeError {
	return &StrictNDJSONDecodeError{LogicalPath: logicalPath, RowOrdinal: rowOrdinal, Kind: kind}
}
