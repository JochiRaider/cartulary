package incidentportability

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Queryer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type ImportedAttribution struct {
	SourceTable   string
	SourceRowID   string
	SourceColumn  string
	SourceActorID string
	LocalUserID   uuid.UUID
}

// AttributionRecorder is the final transaction's historical-attribution
// capability (Core 01 REQ-01-641 step 9). Source ports stage attribution while
// preparing and require only the capability's presence; the attribution owner
// flushes every staged stream once, after ValidateImportTx.
type AttributionRecorder interface {
	FlushImportedAttributions(ctx context.Context, source MemberReader, streams ...AttributionStream) error
}

type UnexpectedColumnsError struct{}

func (*UnexpectedColumnsError) Error() string {
	return "incident portability row contains undeclared columns"
}

type MalformedPayloadError struct {
	Err error
}

func (e *MalformedPayloadError) Error() string {
	return "incident portability payload malformed: " + e.Err.Error()
}

func (e *MalformedPayloadError) Unwrap() error {
	return e.Err
}

type VerificationFailure struct {
	ReasonCode string
}

func (e *VerificationFailure) Error() string {
	return "incident portability verification failed: " + e.ReasonCode
}

type fixedImportFailure struct {
	logicalPath string
}

func (e *fixedImportFailure) Error() string {
	return "incident portability fixed import failed"
}

func FixedImportFailure(logicalPath string) error {
	return &fixedImportFailure{logicalPath: logicalPath}
}

func FixedImportFailurePath(err error) (string, bool) {
	var failure *fixedImportFailure
	if !errors.As(err, &failure) || strings.TrimSpace(failure.logicalPath) == "" {
		return "", false
	}
	return failure.logicalPath, true
}

func ValidateAllowedColumns(row map[string]any, allowed []string) error {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, column := range allowed {
		allowedSet[column] = struct{}{}
	}
	for column := range row {
		if _, ok := allowedSet[column]; !ok {
			return &UnexpectedColumnsError{}
		}
	}
	return nil
}

func DecodeNDJSON(payload []byte) ([]map[string]any, error) {
	rows, err := DecodeStrictNDJSONObjects(payload, "")
	if err != nil {
		return nil, &MalformedPayloadError{Err: err}
	}
	return rows, nil
}

func ValidateRequiredColumns(row map[string]any, required []string, identity []string) error {
	for _, column := range required {
		if _, ok := row[column]; !ok {
			return &VerificationFailure{ReasonCode: "malformed_manifest"}
		}
	}
	if SourceRowID(row, identity) == "" {
		return &VerificationFailure{ReasonCode: "malformed_manifest"}
	}
	return nil
}

func SourceRowID(row map[string]any, identity []string) string {
	parts := make([]string, 0, len(identity))
	for _, column := range identity {
		value := StringFromAny(row[column])
		if strings.TrimSpace(value) == "" {
			return ""
		}
		parts = append(parts, value)
	}
	return strings.Join(parts, ":")
}

func CanonicalRawJSON(raw []byte) ([]byte, error) {
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	return CanonicalJSONString(value)
}

func CanonicalJSONString(value any) ([]byte, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return append(payload, '\n'), nil
}

func StringFromAny(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case json.Number:
		return typed.String()
	default:
		return fmt.Sprint(typed)
	}
}
