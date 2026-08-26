package rollback

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

func sourceForRollbackValue(value map[string]any) (map[string]any, bool) {
	if source, ok := objectMap(value, "source"); ok {
		return source, len(source) > 0
	}
	return nil, false
}

func policyText(value any, valid func(string) bool) bool {
	text, ok := value.(string)
	return ok && valid(text)
}

func nullableSQLString(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *value, Valid: true}
}

func nullableSQLTime(value *time.Time) sql.NullTime {
	if value == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: value.UTC(), Valid: true}
}

func nullableSQLUUID(value *uuid.UUID) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: value.String(), Valid: true}
}

type fieldKind int

const (
	fieldText fieldKind = iota
	fieldUUID
	fieldTime
)

type fieldSpec struct {
	key  string
	kind fieldKind
}

func typedValues(source map[string]any, fields []fieldSpec) ([]any, error) {
	values := make([]any, 0, len(fields)*2)
	for _, field := range fields {
		raw, present := source[field.key]
		var err error
		switch field.kind {
		case fieldUUID:
			raw, _, err = nullableUUID(source, field.key)
		case fieldTime:
			raw, _, err = nullableTime(source, field.key)
		case fieldText:
			if raw != nil {
				_, ok := raw.(string)
				if !ok {
					err = errors.New("invalid text")
				}
			}
		}
		if err != nil {
			return nil, err
		}
		values = append(values, present, raw)
	}
	return values, nil
}

func validTypedFields(source map[string]any, fields []fieldSpec) bool {
	_, err := typedValues(source, fields)
	return err == nil
}

func nullableUUID(value map[string]any, key string) (*uuid.UUID, bool, error) {
	raw, present := value[key]
	if !present || raw == nil {
		return nil, present, nil
	}
	text, ok := raw.(string)
	if !ok || strings.TrimSpace(text) == "" {
		return nil, true, errors.New("invalid uuid")
	}
	parsed, err := uuid.Parse(text)
	if err != nil {
		return nil, true, err
	}
	return &parsed, true, nil
}

func nullableTime(value map[string]any, key string) (*time.Time, bool, error) {
	raw, present := value[key]
	if !present || raw == nil {
		return nil, present, nil
	}
	if timestamp, ok := raw.(time.Time); ok {
		utc := timestamp.UTC()
		return &utc, true, nil
	}
	text, ok := raw.(string)
	if !ok || strings.TrimSpace(text) == "" {
		return nil, true, errors.New("invalid timestamp")
	}
	parsed, err := time.Parse(time.RFC3339Nano, text)
	if err != nil {
		return nil, true, err
	}
	utc := parsed.UTC()
	return &utc, true, nil
}

func nullableText(value map[string]any, key string) (*string, bool, error) {
	raw, present := value[key]
	if !present || raw == nil {
		return nil, present, nil
	}
	text, ok := raw.(string)
	if !ok {
		return nil, true, errors.New("invalid text")
	}
	if strings.TrimSpace(text) == "" {
		return nil, true, nil
	}
	return &text, true, nil
}

func objectMap(value map[string]any, key string) (map[string]any, bool) {
	raw, ok := value[key]
	if !ok || raw == nil {
		return nil, false
	}
	typed, ok := raw.(map[string]any)
	return typed, ok
}

func nonEmptyText(value any) bool {
	text, ok := value.(string)
	return ok && strings.TrimSpace(text) != ""
}
