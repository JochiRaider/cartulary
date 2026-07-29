package incidentportability

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
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

func DecodeStrictNDJSONObjects(payload []byte, logicalPath string) ([]map[string]any, error) {
	if len(payload) == 0 {
		return []map[string]any{}, nil
	}
	lines := bytes.Split(payload, []byte{'\n'})
	rows := make([]map[string]any, 0, len(lines))
	for index, line := range lines {
		if index == len(lines)-1 && len(line) == 0 {
			continue
		}
		ordinal := index + 1
		if len(bytes.TrimSpace(line)) == 0 {
			return nil, strictNDJSONError(logicalPath, ordinal, "blank_line")
		}
		if len(line) > MaxNDJSONLogicalLineBytes {
			return nil, strictNDJSONError(logicalPath, ordinal, "line_too_long")
		}
		decoder := json.NewDecoder(bytes.NewReader(line))
		decoder.UseNumber()
		value, err := decodeStrictJSONValue(decoder)
		if err != nil {
			var strictErr *StrictNDJSONDecodeError
			if errors.As(err, &strictErr) {
				strictErr.LogicalPath = logicalPath
				strictErr.RowOrdinal = ordinal
				return nil, strictErr
			}
			return nil, strictNDJSONError(logicalPath, ordinal, "malformed_json")
		}
		if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
			return nil, strictNDJSONError(logicalPath, ordinal, "multiple_values")
		}
		row, ok := value.(map[string]any)
		if !ok {
			return nil, strictNDJSONError(logicalPath, ordinal, "row_not_object")
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func decodeStrictJSONValue(decoder *json.Decoder) (any, error) {
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	delim, isDelim := token.(json.Delim)
	if !isDelim {
		return token, nil
	}
	switch delim {
	case '{':
		value := map[string]any{}
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return nil, err
			}
			key, ok := keyToken.(string)
			if !ok {
				return nil, strictNDJSONError("", 0, "malformed_object")
			}
			if _, duplicate := value[key]; duplicate {
				return nil, strictNDJSONError("", 0, "duplicate_member")
			}
			member, err := decodeStrictJSONValue(decoder)
			if err != nil {
				return nil, err
			}
			value[key] = member
		}
		closing, err := decoder.Token()
		if err != nil || closing != json.Delim('}') {
			return nil, strictNDJSONError("", 0, "malformed_object")
		}
		return value, nil
	case '[':
		value := []any{}
		for decoder.More() {
			item, err := decodeStrictJSONValue(decoder)
			if err != nil {
				return nil, err
			}
			value = append(value, item)
		}
		closing, err := decoder.Token()
		if err != nil || closing != json.Delim(']') {
			return nil, strictNDJSONError("", 0, "malformed_array")
		}
		return value, nil
	default:
		return nil, strictNDJSONError("", 0, "unexpected_delimiter")
	}
}

func strictNDJSONError(logicalPath string, rowOrdinal int, kind string) *StrictNDJSONDecodeError {
	return &StrictNDJSONDecodeError{
		LogicalPath: logicalPath,
		RowOrdinal:  rowOrdinal,
		Kind:        kind,
	}
}
