package incidentportability

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"

	"github.com/JochiRaider/cartulary/internal/platform/strictjson"
)

// RecordShape closes the top-level members of one owner's NDJSON record.
// Nested values remain the owner's typed contract.
type RecordShape struct {
	Required []string
	Optional []string
	Nullable []string
}

func (s RecordShape) admits(object map[string]json.RawMessage) bool {
	for _, key := range s.Required {
		if _, ok := object[key]; !ok {
			return false
		}
	}
	for key, value := range object {
		if !contains(s.Required, key) && !contains(s.Optional, key) {
			return false
		}
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) && !contains(s.Nullable, key) {
			return false
		}
	}
	return true
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

// RecordShapeError reports a record whose top-level members or typed decode
// violate its owner's closed record contract. Owners map it to an invariant.
type RecordShapeError struct {
	LogicalPath string
	RowOrdinal  int
}

func (e *RecordShapeError) Error() string {
	return "incident portability record violates its closed owner shape"
}

// WalkTyped reads one admitted member in a single pass. Its bytes are hashed
// once by the seal-verifying reader. Each line is framed with the WalkNDJSON
// rules, scanned once for strict admission (valid Unicode, no duplicate member
// at any depth, one object, no trailing value), checked against the closed
// top-level shape and decoded once into T. visit receives the original line
// for owner-specific checks and must not retain it.
func WalkTyped[T any](ctx context.Context, members MemberReader, path string, shape RecordShape, visit func(row *T, line []byte) error) error {
	if visit == nil {
		return strictNDJSONError(path, 0, "invalid_record_limit")
	}
	bound, err := BindMembers(members, []string{path})
	if err != nil {
		return err
	}
	stream, _, err := bound.OpenMember(ctx, path)
	if err != nil {
		return err
	}
	err = walkLines(ctx, stream, path, MaxNDJSONLogicalLineBytes, func(ordinal int, line []byte) error {
		object, err := strictjson.DecodeBoundedUnicodeObject(bytes.NewReader(line), int64(MaxNDJSONLogicalLineBytes))
		if err != nil {
			return strictNDJSONError(path, ordinal, strictDecodeKind(err))
		}
		if !shape.admits(object) {
			return &RecordShapeError{LogicalPath: path, RowOrdinal: ordinal}
		}
		var row T
		if err := json.Unmarshal(line, &row); err != nil {
			return &RecordShapeError{LogicalPath: path, RowOrdinal: ordinal}
		}
		return visit(&row, line)
	})
	return errors.Join(err, stream.Close())
}
