package extensioncodec

import (
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

// PostgresRestoreAlgorithmID selects the Core-owned structural binding validator.
const PostgresRestoreAlgorithmID = "extension_validate_postgres_binding_restore_v1"

type Limits struct {
	MaxItems        int64
	MaxEntryBytes   int64
	MaxBindingBytes int64
}

func (limits Limits) valid() bool {
	return limits.MaxItems > 0 && limits.MaxItems <= 1048576 && limits.MaxEntryBytes > 0 && limits.MaxBindingBytes >= limits.MaxEntryBytes
}

type Binding struct {
	BindingID                  string
	LogicalFamilyID            string
	CodecContractRef           string
	LogicalIdentityAlgorithmID string
	ContentEncodingAlgorithmID string
	Limits                     Limits
}

// Postgres is an immutable, admitted logical codec and physical mapping.
// Copying its pointer is safe; every accessor that exposes a slice copies it.
type Postgres struct {
	binding Binding
	family  Family
	mapping Mapping
}

func NewPostgres(definition Definition, mapping Mapping, binding Binding) (*Postgres, error) {
	if definition.validate() != nil || mapping.validate() != nil || !binding.Limits.valid() ||
		binding.CodecContractRef != definition.SchemaID || mapping.CodecContractRef != definition.SchemaID ||
		mapping.BindingID != binding.BindingID || !strings.HasPrefix(binding.BindingID, definition.ProfileID+".") ||
		binding.LogicalIdentityAlgorithmID != binding.LogicalFamilyID+"_identity_v1" ||
		binding.ContentEncodingAlgorithmID != "extension_registry_canonical_json_v1" {
		return nil, ErrInvalidContract
	}
	for _, family := range definition.Families {
		if family.LogicalFamilyID != binding.LogicalFamilyID {
			continue
		}
		if len(mapping.Columns) != len(family.Fields) {
			return nil, ErrInvalidContract
		}
		for i, field := range family.Fields {
			if mapping.Columns[i].FieldID != field.FieldID {
				return nil, ErrInvalidContract
			}
		}
		family.Fields = append([]Field(nil), family.Fields...)
		family.IdentityFieldIDs = append([]string(nil), family.IdentityFieldIDs...)
		mapping.Columns = append([]Column(nil), mapping.Columns...)
		return &Postgres{binding: binding, family: family, mapping: mapping}, nil
	}
	return nil, ErrInvalidContract
}

func (codec *Postgres) Binding() Binding { return codec.binding }

func (codec *Postgres) Family() Family {
	family := codec.family
	family.Fields = append([]Field(nil), family.Fields...)
	family.IdentityFieldIDs = append([]string(nil), family.IdentityFieldIDs...)
	return family
}

func (codec *Postgres) Mapping() Mapping {
	mapping := codec.mapping
	mapping.Columns = append([]Column(nil), mapping.Columns...)
	return mapping
}

type Entry struct {
	Identity []byte
	Content  []byte
}

// EncodeRow receives values in the mapping's logical-field order. JSON cells
// must be original JSON bytes/text, avoiding database-driver float conversion.
// Physical names and database-native whole-row serialization are never inputs.
func (codec *Postgres) EncodeRow(values []any) (Entry, error) {
	if codec == nil || len(values) != len(codec.family.Fields) {
		return Entry{}, ErrInvalidEntry
	}
	object := make(map[string]any, len(values))
	for i, field := range codec.family.Fields {
		value, err := logicalValue(field, values[i])
		if err != nil {
			return Entry{}, err
		}
		object[field.FieldID] = value
	}
	identityValues := make([]any, len(codec.family.IdentityFieldIDs))
	for i, fieldID := range codec.family.IdentityFieldIDs {
		identityValues[i] = object[fieldID]
	}
	identity, err := canonicaljson.MarshalExtension(identityValues)
	if err != nil {
		return Entry{}, ErrInvalidEntry
	}
	if len(identity) == 0 || len(identity) > 4096 {
		return Entry{}, ErrLimitExceeded
	}
	content, err := canonicaljson.MarshalExtension(object)
	if err != nil {
		return Entry{}, ErrInvalidEntry
	}
	if int64(len(content)) > codec.binding.Limits.MaxEntryBytes {
		return Entry{}, ErrLimitExceeded
	}
	return Entry{Identity: identity, Content: content}, nil
}

func logicalValue(field Field, value any) (any, error) {
	if value == nil {
		if field.Nullable {
			return nil, nil
		}
		return nil, ErrInvalidEntry
	}
	switch field.ValueKind {
	case "text":
		if text, ok := value.(string); ok && utf8.ValidString(text) {
			return text, nil
		}
	case "uuid":
		var encoded string
		switch typed := value.(type) {
		case [16]byte:
			encoded = hex.EncodeToString(typed[:])
		case string:
			if len(typed) != 36 || typed[8] != '-' || typed[13] != '-' || typed[18] != '-' || typed[23] != '-' {
				return nil, ErrInvalidEntry
			}
			encoded = strings.ReplaceAll(typed, "-", "")
			decoded, err := hex.DecodeString(encoded)
			if err != nil || len(decoded) != 16 {
				return nil, ErrInvalidEntry
			}
			encoded = hex.EncodeToString(decoded)
		default:
			return nil, ErrInvalidEntry
		}
		return encoded[:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:], nil
	case "int64":
		switch typed := value.(type) {
		case int:
			return int64(typed), nil
		case int16:
			return int64(typed), nil
		case int32:
			return int64(typed), nil
		case int64:
			return typed, nil
		}
	case "boolean":
		if boolean, ok := value.(bool); ok {
			return boolean, nil
		}
	case "timestamp":
		if timestamp, ok := value.(time.Time); ok {
			timestamp = timestamp.UTC()
			if timestamp.Year() < 1 || timestamp.Year() > 9999 || timestamp.Nanosecond()%1000 != 0 {
				return nil, ErrInvalidEntry
			}
			return timestamp.Format("2006-01-02T15:04:05.000000Z"), nil
		}
	case "text_array":
		var values []any
		switch typed := value.(type) {
		case []string:
			values = make([]any, len(typed))
			for i, text := range typed {
				values[i] = text
			}
		case []any:
			values = append(make([]any, 0, len(typed)), typed...)
		default:
			return nil, ErrInvalidEntry
		}
		for _, value := range values {
			if text, ok := value.(string); !ok || !utf8.ValidString(text) {
				return nil, ErrInvalidEntry
			}
		}
		return values, nil
	case "json":
		var raw []byte
		switch typed := value.(type) {
		case json.RawMessage:
			raw = typed
		case []byte:
			raw = typed
		case string:
			raw = []byte(typed)
		default:
			return nil, ErrInvalidEntry
		}
		decoded, err := canonicaljson.DecodeStrict(raw)
		if err != nil {
			return nil, ErrInvalidEntry
		}
		return decoded, nil
	}
	return nil, ErrInvalidEntry
}
