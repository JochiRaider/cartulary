package queryengine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	entityports "github.com/JochiRaider/cartulary/internal/modules/entities/projectionports"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type IdentityReader struct {
	db postgres.DB
}

func NewIdentityReader(db postgres.DB) *IdentityReader {
	return &IdentityReader{db: db}
}

func (reader *IdentityReader) SelectIdentityQueryProjections(
	ctx context.Context,
	incidentID uuid.UUID,
	query viewschema.QueryMeta,
	window querypage.Window,
) ([]entityports.IdentityQueryProjection, error) {
	if reader == nil || reader.db == nil {
		return nil, fmt.Errorf("query identity projections: database is required")
	}
	sqlText, args, err := buildIdentityQueryPageSQL(incidentID, query, window)
	if err != nil {
		return nil, err
	}
	rows, err := reader.db.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, fmt.Errorf("query identity projections: %w", err)
	}
	defer rows.Close()

	result := make([]entityports.IdentityQueryProjection, 0)
	for rows.Next() {
		row, scanErr := scanIdentityQueryProjection(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate identity projections: %w", err)
	}
	return result, nil
}

func (reader *IdentityReader) CollectIdentityDerivedFactsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
) ([]entityports.DerivedFact, error) {
	rows, err := tx.Query(ctx, `
SELECT i.record_id, to_jsonb(i) - 'incident_id'
  FROM identity_grid_projection i
  JOIN records r
    ON r.incident_id = i.incident_id
   AND r.record_id = i.record_id
   AND r.deleted_at IS NULL
 WHERE i.incident_id = $1
 ORDER BY i.record_id
`, incidentID)
	if err != nil {
		return nil, fmt.Errorf("collect identity projection facts: %w", err)
	}
	defer rows.Close()

	facts := make([]entityports.DerivedFact, 0)
	for rows.Next() {
		var (
			recordID uuid.UUID
			raw      []byte
		)
		if err := rows.Scan(&recordID, &raw); err != nil {
			return nil, fmt.Errorf("scan identity projection fact: %w", err)
		}
		value := map[string]any{}
		if err := json.Unmarshal(raw, &value); err != nil {
			return nil, fmt.Errorf("decode identity projection fact: %w", err)
		}
		facts = append(facts, entityports.DerivedFact{
			RecordID:     recordID,
			RecordType:   "identity",
			ContentClass: "derived_analytic",
			Value:        value,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate identity projection facts: %w", err)
	}
	return facts, nil
}

var identitySortExpressions = map[string]string{
	"record_id":                   "p.record_id",
	"identity.display_name":       "p.display_name",
	"identity.upn":                "p.upn",
	"identity.email":              "p.email",
	"identity.sam_account_name":   "p.sam_account_name",
	"identity.identity_state":     "p.identity_state",
	"identity.linked_event_count": "p.linked_event_count",
	"identity.evidence_count":     "p.evidence_count",
	"identity.privilege_level":    "p.privilege_level",
	"identity.mfa_state":          "p.mfa_state",
	"identity.reset_status":       "p.reset_status",
	"identity.edited_at":          "p.edited_at",
}

func buildIdentityQueryPageSQL(
	incidentID uuid.UUID,
	query viewschema.QueryMeta,
	window querypage.Window,
) (string, []any, error) {
	var builder strings.Builder
	builder.WriteString(`
SELECT
    p.record_id,
    p.identity_state,
    p.linked_event_count,
    p.evidence_count,
    p.privilege_level,
    p.mfa_state,
    p.reset_status
  FROM identity_grid_projection p
  JOIN records r
    ON r.record_id = p.record_id
   AND r.incident_id = p.incident_id
 WHERE p.incident_id = $1
   AND r.deleted_at IS NULL
   AND p.identity_state IN ('stub', 'canonical')`)
	args := []any{incidentID}

	for _, filter := range query.Filters {
		var expression string
		switch filter.FieldKey {
		case "identity.identity_state":
			expression = "p.identity_state"
		case "identity.privilege_level":
			expression = "p.privilege_level"
		case "identity.mfa_state":
			expression = "p.mfa_state"
		case "identity.reset_status":
			expression = "p.reset_status"
		default:
			return "", nil, fmt.Errorf("identity query filter field %q not mapped", filter.FieldKey)
		}
		if err := appendEntityQueryTextClause(&builder, &args, expression, filter); err != nil {
			return "", nil, err
		}
	}

	if err := querypage.AppendKeyset(
		&builder,
		&args,
		query.Sort,
		entityPageFields(identitySortExpressions),
		window.Position,
	); err != nil {
		return "", nil, err
	}
	if err := appendEntityOrderBy(&builder, query.Sort, identitySortExpressions); err != nil {
		return "", nil, err
	}
	if err := querypage.AppendLimit(&builder, &args, window.Limit); err != nil {
		return "", nil, err
	}
	return builder.String(), args, nil
}

func scanIdentityQueryProjection(scanner interface{ Scan(...any) error }) (entityports.IdentityQueryProjection, error) {
	var (
		row              entityports.IdentityQueryProjection
		privilegeLevel   pgtype.Text
		mfaState         pgtype.Text
		resetStatus      pgtype.Text
		linkedEventCount int32
		evidenceCount    int32
	)
	if err := scanner.Scan(
		&row.RecordID,
		&row.IdentityState,
		&linkedEventCount,
		&evidenceCount,
		&privilegeLevel,
		&mfaState,
		&resetStatus,
	); err != nil {
		return entityports.IdentityQueryProjection{}, fmt.Errorf("scan identity projection query row: %w", err)
	}
	row.LinkedEventCount = int(linkedEventCount)
	row.EvidenceCount = int(evidenceCount)
	row.PrivilegeLevel = queryTextPointer(privilegeLevel)
	row.MFAState = queryTextPointer(mfaState)
	row.ResetStatus = queryTextPointer(resetStatus)
	return row, nil
}
