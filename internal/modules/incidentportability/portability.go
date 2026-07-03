package incidentportability

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Queryer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type File struct {
	Path    string
	Payload []byte
}

type AttributionRecorder interface {
	RecordImportedAttribution(table string, row map[string]any, column string, sourceActorID string)
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

func ExportNDJSON(ctx context.Context, q Queryer, incidentID uuid.UUID, path string, query string) (File, error) {
	payload, err := ExportNDJSONPayload(ctx, q, incidentID, query)
	if err != nil {
		return File{}, err
	}
	return File{Path: path, Payload: payload}, nil
}

func ExportNDJSONPayload(ctx context.Context, q Queryer, incidentID uuid.UUID, query string) ([]byte, error) {
	rows, err := q.Query(ctx, query, incidentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return EncodeRows(rows)
}

func EncodeRows(rows pgx.Rows) ([]byte, error) {
	var buf bytes.Buffer
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		canonical, err := CanonicalRawJSON(raw)
		if err != nil {
			return nil, err
		}
		buf.Write(canonical)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func ImportBundleFileNDJSON(ctx context.Context, tx pgx.Tx, target ImportTargetDescriptor, files map[string][]byte, actorUserID uuid.UUID, attributions AttributionRecorder) error {
	payload, ok := files[target.LogicalBundlePath]
	if !ok {
		return &VerificationFailure{ReasonCode: "missing_required_file"}
	}
	return ImportNDJSON(ctx, tx, target, payload, actorUserID, attributions)
}

func ImportNDJSON(ctx context.Context, tx pgx.Tx, target ImportTargetDescriptor, payload []byte, actorUserID uuid.UUID, attributions AttributionRecorder) error {
	if err := validateRegisteredImportTarget(target); err != nil {
		return err
	}
	rows, err := DecodeNDJSON(payload)
	if err != nil {
		return err
	}
	return ImportRows(ctx, tx, target, rows, actorUserID, attributions)
}

func ImportRows(ctx context.Context, tx pgx.Tx, target ImportTargetDescriptor, rows []map[string]any, actorUserID uuid.UUID, attributions AttributionRecorder) error {
	if err := validateRegisteredImportTarget(target); err != nil {
		return err
	}
	for _, row := range rows {
		if err := ValidateRequiredColumns(target, row); err != nil {
			return err
		}
		RemapTopLevelUserFields(row, target.TargetRelation, actorUserID, attributions)
		raw, err := json.Marshal(row)
		if err != nil {
			return err
		}
		query := fmt.Sprintf("INSERT INTO %s SELECT * FROM jsonb_populate_record(NULL::%s, $1::jsonb) ON CONFLICT DO NOTHING", target.TargetRelation, target.TargetRelation)
		if _, err := tx.Exec(ctx, query, raw); err != nil {
			return err
		}
	}
	return nil
}

func DecodeNDJSON(payload []byte) ([]map[string]any, error) {
	scanner := bufio.NewScanner(bytes.NewReader(payload))
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	var rows []map[string]any
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var row map[string]any
		decoder := json.NewDecoder(bytes.NewReader(line))
		decoder.UseNumber()
		if err := decoder.Decode(&row); err != nil {
			return nil, &MalformedPayloadError{Err: err}
		}
		rows = append(rows, row)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return rows, nil
}

func RemapTopLevelUserFields(row map[string]any, table string, actorUserID uuid.UUID, attributions AttributionRecorder) {
	for key, value := range row {
		if !strings.HasSuffix(key, "_user_id") || value == nil {
			continue
		}
		sourceActorID := StringFromAny(value)
		if strings.TrimSpace(sourceActorID) == "" {
			continue
		}
		if attributions != nil {
			attributions.RecordImportedAttribution(table, row, key, sourceActorID)
		}
		row[key] = actorUserID.String()
	}
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
