package reporting

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

var errCanonicalReportingJSON = errors.New("reporting: invalid canonical JSON")

// Reporting owns UTF-8 byte ordering and integer-only number tokens. This is
// deliberately separate from the Reference Pack RFC 8785 canonicalizer, whose
// UTF-16 ordering and number admission have different contracts.
func canonicalReportingBytes(raw []byte) ([]byte, error) {
	value, err := canonicaljson.DecodeStrict(raw)
	if err != nil {
		return nil, errCanonicalReportingJSON
	}
	return appendReportingJSON(nil, value)
}

func marshalReportingJSON(value any) ([]byte, error) {
	if err := validateReportingValue(reflect.ValueOf(value), 0); err != nil {
		return nil, err
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, errCanonicalReportingJSON
	}
	return canonicalReportingBytes(raw)
}

// Check producer values before encoding/json can repair malformed strings or
// encode a binary slice as an ordinary string. RawMessage is JSON admission,
// never permission to skip duplicate, Unicode or number checks.
func validateReportingValue(v reflect.Value, depth int) error {
	if !v.IsValid() {
		return nil
	}
	if depth > 10000 {
		return errCanonicalReportingJSON
	}
	if v.Type() == reflect.TypeFor[json.RawMessage]() {
		if v.IsNil() {
			return nil
		}
		_, err := canonicalReportingBytes(v.Bytes())
		return err
	}
	if v.Type() == reflect.TypeFor[time.Time]() {
		return nil
	}
	switch v.Kind() {
	case reflect.Interface, reflect.Pointer:
		if v.IsNil() {
			return nil
		}
		return validateReportingValue(v.Elem(), depth+1)
	case reflect.String:
		if !utf8.ValidString(v.String()) {
			return errCanonicalReportingJSON
		}
	case reflect.Float32, reflect.Float64:
		n := v.Float()
		if math.IsNaN(n) || math.IsInf(n, 0) || math.Trunc(n) != n || math.Abs(n) > 9007199254740991 || (n == 0 && math.Signbit(n)) {
			return errCanonicalReportingJSON
		}
	case reflect.Slice, reflect.Array:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return errCanonicalReportingJSON
		}
		for i := 0; i < v.Len(); i++ {
			if err := validateReportingValue(v.Index(i), depth+1); err != nil {
				return err
			}
		}
	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String {
			return errCanonicalReportingJSON
		}
		iter := v.MapRange()
		for iter.Next() {
			if !utf8.ValidString(iter.Key().String()) {
				return errCanonicalReportingJSON
			}
			if err := validateReportingValue(iter.Value(), depth+1); err != nil {
				return err
			}
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)
			if !field.IsExported() || strings.Split(field.Tag.Get("json"), ",")[0] == "-" {
				continue
			}
			if err := validateReportingValue(v.Field(i), depth+1); err != nil {
				return err
			}
		}
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
	default:
		return errCanonicalReportingJSON
	}
	return nil
}

func appendReportingJSON(out []byte, value any) ([]byte, error) {
	switch v := value.(type) {
	case nil:
		return append(out, "null"...), nil
	case bool:
		return strconv.AppendBool(out, v), nil
	case string:
		return appendReportingString(out, v), nil
	case json.Number:
		s := v.String()
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil || n < -9007199254740991 || n > 9007199254740991 || strconv.FormatInt(n, 10) != s {
			return nil, errCanonicalReportingJSON
		}
		return append(out, s...), nil
	case []any:
		out = append(out, '[')
		for i, item := range v {
			if i > 0 {
				out = append(out, ',')
			}
			var err error
			out, err = appendReportingJSON(out, item)
			if err != nil {
				return nil, err
			}
		}
		return append(out, ']'), nil
	case map[string]any:
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		slices.Sort(keys)
		out = append(out, '{')
		for i, key := range keys {
			if i > 0 {
				out = append(out, ',')
			}
			out = appendReportingString(out, key)
			out = append(out, ':')
			var err error
			out, err = appendReportingJSON(out, v[key])
			if err != nil {
				return nil, err
			}
		}
		return append(out, '}'), nil
	default:
		return nil, errCanonicalReportingJSON
	}
}

func appendReportingString(out []byte, value string) []byte {
	out = append(out, '"')
	const hex = "0123456789abcdef"
	for _, r := range value {
		switch r {
		case '"', '\\':
			out = append(out, '\\', byte(r))
		case '\b':
			out = append(out, '\\', 'b')
		case '\t':
			out = append(out, '\\', 't')
		case '\n':
			out = append(out, '\\', 'n')
		case '\f':
			out = append(out, '\\', 'f')
		case '\r':
			out = append(out, '\\', 'r')
		default:
			if r < 32 {
				out = append(out, '\\', 'u', '0', '0', hex[r>>4], hex[r&15])
			} else {
				out = utf8.AppendRune(out, r)
			}
		}
	}
	return append(out, '"')
}

// Object identities include the owner's schema domain. Exact file checksums
// and Extensions output checksums continue to use hashHex on raw file bytes.
func reportingObjectDigest(schemaID string, canonical []byte) string {
	var input bytes.Buffer
	input.WriteString(schemaID)
	input.WriteByte('\n')
	input.Write(canonical)
	return hashHex(input.Bytes())
}
