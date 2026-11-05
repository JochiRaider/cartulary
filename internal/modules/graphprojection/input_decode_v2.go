package graphprojection

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Components refer to decoded member names or numeric array indices. A path
// is materialized only for the winning diagnostic, within the message limit.
type inputPathV2 []any

func (p inputPathV2) render(root string) string {
	for _, part := range p {
		var next string
		switch v := part.(type) {
		case string:
			next = canonicalInputMemberPath(root, v)
		case int:
			next = root + "[" + strconv.Itoa(v) + "]"
		}
		if next == root || len(next) > graphProjectionLimits.MaxValidationMessageBytes {
			break
		}
		root = next
	}
	return root
}

func decodeInputV2(ctx context.Context, data []byte) (any, error) {
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	v, err := readInputValueV2(ctx, d, data, nil)
	if err != nil {
		return nil, err
	}
	if _, err := d.Token(); err != io.EOF {
		return nil, invalidProjectionErrorV2("invalid_json_syntax", "")
	}
	return v, nil
}

func readInputValueV2(ctx context.Context, d *json.Decoder, data []byte, path inputPathV2) (any, error) {
	if err := ctx.Err(); err != nil {
		return nil, cancelledProjectionErrorV2(err)
	}
	token, err := unicodeTokenV2(d, data)
	if err != nil {
		return nil, invalidProjectionErrorV2("invalid_json_syntax", "")
	}
	delim, container := token.(json.Delim)
	if !container {
		return token, nil
	}
	limit := graphProjectionLimits.MaxInputJSONDepth
	if len(path) >= limit {
		return nil, resourceProjectionErrorV2("maximum_input_json_depth", limit, limit+1)
	}
	var value any
	switch delim {
	case '{':
		object := make(map[string]any)
		for d.More() {
			k, err := unicodeTokenV2(d, data)
			key, ok := k.(string)
			if err != nil || !ok {
				return nil, invalidProjectionErrorV2("invalid_json_syntax", "")
			}
			child := append(path, key)
			if _, duplicate := object[key]; duplicate {
				return nil, invalidProjectionErrorV2("duplicate_object_member", child.render("$"))
			}
			object[key], err = readInputValueV2(ctx, d, data, child)
			if err != nil {
				return nil, err
			}
		}
		value = object
	case '[':
		items := []any{}
		for d.More() {
			v, err := readInputValueV2(ctx, d, data, append(path, len(items)))
			if err != nil {
				return nil, err
			}
			items = append(items, v)
		}
		value = items
	default:
		return nil, invalidProjectionErrorV2("invalid_json_syntax", "")
	}
	if _, err := d.Token(); err != nil {
		return nil, invalidProjectionErrorV2("invalid_json_syntax", "")
	}
	return value, nil
}

// encoding/json reports malformed surrogate escapes as U+FFFD. Inspect the
// original token only when that scalar appears in the decoded string.
func unicodeTokenV2(d *json.Decoder, data []byte) (json.Token, error) {
	start := d.InputOffset()
	token, err := d.Token()
	value, isString := token.(string)
	if err != nil || !isString || !strings.ContainsRune(value, utf8.RuneError) {
		return token, err
	}
	raw := data[start:d.InputOffset()]
	for i := 0; i < len(raw); i++ {
		if raw[i] != '\\' {
			continue
		}
		i++
		if raw[i] != 'u' {
			continue
		}
		scalar, _ := strconv.ParseUint(string(raw[i+1:i+5]), 16, 16)
		i += 4
		if scalar < 0xd800 || scalar > 0xdfff {
			continue
		}
		if scalar > 0xdbff || len(raw)-i <= 6 || raw[i+1] != '\\' || raw[i+2] != 'u' {
			return nil, invalidProjectionErrorV2("invalid_json_syntax", "")
		}
		low, err := strconv.ParseUint(string(raw[i+3:i+7]), 16, 16)
		if err != nil || low < 0xdc00 || low > 0xdfff {
			return nil, invalidProjectionErrorV2("invalid_json_syntax", "")
		}
		i += 6
	}
	return token, nil
}
