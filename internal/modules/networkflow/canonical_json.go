package networkflow

import (
	"bytes"
	"encoding"
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

var errCanonicalValue = errors.New("invalid Network Flow canonical value")

type canonicalMember struct {
	name  string
	value any
}

type canonicalObject []canonicalMember

// canonicalJSON retains the distinction between closed owner types and dynamic
// JSON maps. It never uses a lossy JSON round trip to determine object order.
func canonicalJSON(value any) ([]byte, error) {
	var output bytes.Buffer
	if err := appendCanonical(&output, reflect.ValueOf(value), map[canonicalVisit]bool{}); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func writeCanonical(out *bytes.Buffer, value any) error {
	data, err := canonicalJSON(value)
	if err != nil {
		return err
	}
	out.Write(data)
	return nil
}

// Invalid encodings never establish equality at an admission boundary.
func canonicalEqual(left, right any) bool {
	a, err := canonicalJSON(left)
	if err != nil {
		return false
	}
	b, err := canonicalJSON(right)
	return err == nil && bytes.Equal(a, b)
}

type canonicalVisit struct {
	typeID reflect.Type
	ptr    uintptr
}

func appendCanonical(out *bytes.Buffer, value reflect.Value, active map[canonicalVisit]bool) error {
	if !value.IsValid() {
		out.WriteString("null")
		return nil
	}
	if value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer {
		if value.IsNil() {
			out.WriteString("null")
			return nil
		}
		if value.Kind() == reflect.Interface {
			return appendCanonical(out, value.Elem(), active)
		}
	}
	if value.Kind() == reflect.Pointer || value.Kind() == reflect.Map || value.Kind() == reflect.Slice {
		visit := canonicalVisit{value.Type(), uintptr(value.UnsafePointer())}
		if active[visit] {
			return errCanonicalValue
		}
		active[visit] = true
		defer delete(active, visit)
	}
	if value.Kind() == reflect.Pointer {
		return appendCanonical(out, value.Elem(), active)
	}
	write := func(v any) error { return appendCanonical(out, reflect.ValueOf(v), active) }
	if value.CanInterface() {
		switch v := value.Interface().(type) {
		case json.Number:
			text := string(v)
			if !canonicalInteger(text) {
				return errCanonicalValue
			}
			out.WriteString(text)
			return nil
		case json.RawMessage:
			decoded, err := canonicaljson.DecodeStrict(v)
			if err != nil {
				return errCanonicalValue
			}
			return write(decoded)
		case retainedRawValues:
			object, err := v.canonicalObject()
			if err != nil {
				return err
			}
			return write(object)
		case timestampProfile:
			return write(v.canonicalObject())
		case fieldMapping:
			switch v.MappingKind {
			case mappingKindSourceColumn:
				return write(canonicalObject{{"mapping_kind", v.MappingKind}, {"field_key", v.FieldKey}, {"source_column_ordinal", v.SourceColumnOrdinal}, {"transform_id", v.TransformID}, {"empty_value_policy", v.EmptyValuePolicy}, {"combinability", v.Combinability}})
			case mappingKindIgnoredSourceColumn:
				return write(canonicalObject{{"mapping_kind", v.MappingKind}, {"source_column_ordinal", v.SourceColumnOrdinal}, {"ignore_reason", v.IgnoreReason}})
			case mappingKindSystemDerivation:
				return write(canonicalObject{{"mapping_kind", v.MappingKind}, {"field_key", v.FieldKey}, {"derivation_id", v.DerivationID}, {"combinability", v.Combinability}})
			default:
				return errCanonicalValue
			}
		case canonicalObject:
			out.WriteByte('{')
			seen := map[string]bool{}
			for i, member := range v {
				if seen[member.name] {
					return errCanonicalValue
				}
				seen[member.name] = true
				if i > 0 {
					out.WriteByte(',')
				}
				if err := appendCanonicalString(out, member.name); err != nil {
					return err
				}
				out.WriteByte(':')
				if err := write(member.value); err != nil {
					return err
				}
			}
			out.WriteByte('}')
			return nil
		case json.Marshaler:
			// Scalar owner types (timestamps and UUIDs) retain their JSON scalar
			// representation. Closed custom objects require an explicit order.
			data, err := v.MarshalJSON()
			if err != nil {
				return err
			}
			decoded, err := canonicaljson.DecodeStrict(data)
			if err != nil {
				return errCanonicalValue
			}
			if _, object := decoded.(map[string]any); object {
				return errCanonicalValue
			}
			return write(decoded)
		case encoding.TextMarshaler:
			data, err := v.MarshalText()
			if err != nil {
				return err
			}
			return appendCanonicalString(out, string(data))
		}
	}
	switch value.Kind() {
	case reflect.String:
		return appendCanonicalString(out, value.String())
	case reflect.Bool:
		out.WriteString(strconv.FormatBool(value.Bool()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		out.WriteString(strconv.FormatInt(value.Int(), 10))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		out.WriteString(strconv.FormatUint(value.Uint(), 10))
	case reflect.Map:
		if value.Type().Key().Kind() != reflect.String {
			return errCanonicalValue
		}
		if value.IsNil() {
			out.WriteString("null")
			return nil
		}
		keys := value.MapKeys()
		sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })
		out.WriteByte('{')
		for i, key := range keys {
			if i > 0 {
				out.WriteByte(',')
			}
			if err := appendCanonicalString(out, key.String()); err != nil {
				return err
			}
			out.WriteByte(':')
			if err := appendCanonical(out, value.MapIndex(key), active); err != nil {
				return err
			}
		}
		out.WriteByte('}')
	case reflect.Array, reflect.Slice:
		if value.Kind() == reflect.Slice && value.IsNil() {
			out.WriteString("null")
			return nil
		}
		out.WriteByte('[')
		for i := 0; i < value.Len(); i++ {
			if i > 0 {
				out.WriteByte(',')
			}
			if err := appendCanonical(out, value.Index(i), active); err != nil {
				return err
			}
		}
		out.WriteByte(']')
	case reflect.Struct:
		members := canonicalObject{}
		for i := 0; i < value.NumField(); i++ {
			field := value.Type().Field(i)
			if !field.IsExported() {
				continue
			}
			name, options, _ := strings.Cut(field.Tag.Get("json"), ",")
			if name == "-" {
				continue
			}
			if name == "" || field.Anonymous || options != "" && options != "omitempty" {
				return errCanonicalValue
			}
			member := value.Field(i)
			if options == "omitempty" && canonicalEmpty(member) {
				continue
			}
			members = append(members, canonicalMember{name, member.Interface()})
		}
		return write(members)
	default:
		return errCanonicalValue
	}
	return nil
}

func canonicalEmpty(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return value.Len() == 0
	case reflect.Struct:
		return false
	default:
		return value.IsZero()
	}
}

func canonicalInteger(text string) bool {
	if strings.HasPrefix(text, "-") {
		text = text[1:]
		if text == "0" {
			return false
		}
	}
	if text == "" || len(text) > 1 && text[0] == '0' {
		return false
	}
	for _, digit := range text {
		if digit < '0' || digit > '9' {
			return false
		}
	}
	return true
}

func appendCanonicalString(out *bytes.Buffer, value string) error {
	if !utf8.ValidString(value) {
		return errCanonicalValue
	}
	out.WriteByte('"')
	for _, r := range value {
		switch r {
		case '"', '\\':
			out.WriteByte('\\')
			out.WriteRune(r)
		case '\b':
			out.WriteString(`\b`)
		case '\t':
			out.WriteString(`\t`)
		case '\n':
			out.WriteString(`\n`)
		case '\f':
			out.WriteString(`\f`)
		case '\r':
			out.WriteString(`\r`)
		default:
			if r < 0x20 {
				out.WriteString(`\u00`)
				out.WriteByte("0123456789abcdef"[r>>4])
				out.WriteByte("0123456789abcdef"[r&15])
			} else {
				out.WriteRune(r)
			}
		}
	}
	out.WriteByte('"')
	return nil
}

func (profile timestampProfile) canonicalObject() canonicalObject {
	members := canonicalObject{{"schema_id", profile.SchemaID}, {"mode", profile.Mode}, {"precision", profile.Precision}}
	switch profile.Mode {
	case "rfc3339":
		members = append(members, canonicalMember{"timezone", profile.Timezone}, canonicalMember{"timezone_ruleset_id", profile.TimezoneRulesetID}, canonicalMember{"ambiguous_local_time_policy", profile.AmbiguousLocalTimePolicy}, canonicalMember{"local_time_gap_policy", profile.LocalTimeGapPolicy})
	case "netflow_sys_uptime_milliseconds":
		members = append(members, canonicalMember{"netflow_export_time_column_ordinal", profile.NetFlowExportTimeColumnOrdinal}, canonicalMember{"netflow_export_time_mode", profile.NetFlowExportTimeMode}, canonicalMember{"netflow_exporter_uptime_at_export_column_ordinal", profile.NetFlowExporterUptimeAtExportColumnOrdinal})
	}
	return members
}
