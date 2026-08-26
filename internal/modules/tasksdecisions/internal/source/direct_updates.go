// Package source owns transaction-bound Task/Decision persistence mechanics.
package source

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/policy"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/sourcecatalog"
)

type Error struct {
	Operation string
	Err       error
}

func (e *Error) Error() string { return "tasksdecisions source: " + e.Operation }
func (e *Error) Unwrap() error { return e.Err }

func ApplyTaskDirectChangeTx(
	ctx context.Context,
	tx pgx.Tx,
	catalog *sourcecatalog.Catalog,
	recordID uuid.UUID,
	fieldKey string,
	value policy.FieldValue,
	now time.Time,
) (bool, error) {
	query, ok := directUpdateStatement(catalog, "cartulary.view.task_requests.v1", fieldKey)
	if !ok {
		return false, &policy.ValidationError{Field: fieldKey, ReasonCode: "unsupported_field_key"}
	}
	tag, err := tx.Exec(ctx, query, recordID, DirectDBValue(value), now.UTC())
	if err != nil {
		return false, &Error{Operation: "apply task direct change", Err: err}
	}
	return tag.RowsAffected() > 0, nil
}

func ApplyDecisionDirectChangeTx(
	ctx context.Context,
	tx pgx.Tx,
	catalog *sourcecatalog.Catalog,
	recordID uuid.UUID,
	fieldKey string,
	value policy.FieldValue,
	now time.Time,
) (bool, error) {
	query, ok := directUpdateStatement(catalog, "cartulary.view.decisions.v1", fieldKey)
	if !ok {
		return false, &policy.ValidationError{Field: fieldKey, ReasonCode: "unsupported_field_key"}
	}
	tag, err := tx.Exec(ctx, query, recordID, DirectDBValue(value), now.UTC())
	if err != nil {
		return false, &Error{Operation: "apply decision direct change", Err: err}
	}
	return tag.RowsAffected() > 0, nil
}

func directUpdateStatement(catalog *sourcecatalog.Catalog, viewSchemaID string, fieldKey string) (string, bool) {
	field, ok := catalog.Field(fieldKey)
	if !ok || field.ViewSchemaID != viewSchemaID || field.Kind != sourcecatalog.FieldKindDirect {
		return "", false
	}
	table := pgx.Identifier{field.Storage.Table}.Sanitize()
	column := pgx.Identifier{field.Storage.Column}.Sanitize()
	return fmt.Sprintf(
		"UPDATE %s SET %s = $2, updated_at = $3 WHERE record_id = $1 AND %s IS DISTINCT FROM $2",
		table, column, column,
	), true
}

func DirectDBValue(value policy.FieldValue) any {
	switch {
	case value.Text != nil:
		return *value.Text
	case value.Timestamp != nil:
		return value.Timestamp.UTC()
	case value.UUID != nil:
		return *value.UUID
	case value.Number != nil:
		return *value.Number
	case value.Bool != nil:
		return *value.Bool
	default:
		return nil
	}
}

func IsUniqueViolation(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) && postgresError.Code == "23505"
}

func ClassifyUniqueConflict(err error, operation string) error {
	if IsUniqueViolation(err) {
		return &Error{Operation: operation, Err: err}
	}
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", operation, err)
}
