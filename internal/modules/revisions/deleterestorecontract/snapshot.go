package deleterestorecontract

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// SourceFieldGroup describes scalar fields selected explicitly by a source owner.
// Every named member is required, including members whose value may be null.
type SourceFieldGroup struct {
	Names    string
	Kind     SourceScalarKind
	Nullable bool
}

type SourceScalarKind uint8

const (
	SourceString SourceScalarKind = iota + 1
	SourceUUID
	SourceTimestamp
	SourceInteger
	SourceBoolean
)

// ValidateSourceObject checks only closed scalar shape and record/incident
// identity. Domain vocabulary, normalization and lifecycle rules belong to owners.
// Validation never changes either object or normalizes retained strings.
func ValidateSourceObject(record, source map[string]any, groups ...SourceFieldGroup) error {
	fields := make(map[string]struct{})
	for _, group := range groups {
		for _, name := range strings.Fields(group.Names) {
			if _, duplicate := fields[name]; duplicate {
				return fmt.Errorf("duplicate source schema field %q", name)
			}
			fields[name] = struct{}{}
			value, present := source[name]
			if !present || (value == nil && !group.Nullable) {
				return fmt.Errorf("missing or null source field %q", name)
			}
			if value != nil && !validSourceScalar(value, group.Kind) {
				return fmt.Errorf("invalid source field %q", name)
			}
		}
	}
	if len(source) != len(fields) {
		return fmt.Errorf("source object contains unexpected members")
	}
	for _, name := range []string{"record_id", "incident_id"} {
		if value, ok := source[name].(string); !ok || value != record[name] {
			return fmt.Errorf("source %s does not match record envelope", name)
		}
	}
	return nil
}

func validSourceScalar(value any, kind SourceScalarKind) bool {
	switch kind {
	case SourceString:
		_, ok := value.(string)
		return ok
	case SourceUUID:
		text, ok := value.(string)
		id, err := uuid.Parse(text)
		return ok && err == nil && id != uuid.Nil && text == id.String()
	case SourceTimestamp:
		text, ok := value.(string)
		parsed, err := time.Parse(time.RFC3339Nano, text)
		if !ok || err != nil {
			return false
		}
		_, offset := parsed.Zone()
		return offset == 0 && parsed.Year() >= 1 && parsed.Year() <= 9999 && parsed.Nanosecond()%1000 == 0 &&
			(text == parsed.UTC().Format(time.RFC3339Nano) ||
				text == strings.TrimSuffix(parsed.UTC().Format(time.RFC3339Nano), "Z")+"+00:00")
	case SourceInteger:
		number, ok := value.(json.Number)
		if !ok {
			return false
		}
		_, err := number.Int64()
		return err == nil
	case SourceBoolean:
		_, ok := value.(bool)
		return ok
	default:
		return false
	}
}
