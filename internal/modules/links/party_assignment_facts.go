package links

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/links/partyassignment"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var partyAssignmentFields = []string{
	"host.subject_party_ids", "host.responsible_party_ids", "host.observer_party_ids",
	"identity.subject_party_ids", "identity.responsible_party_ids", "identity.observer_party_ids",
}

// IsPartyAssignmentField identifies only the explicitly owned collections.
// Assignment membership is not an evidence-support relationship.
func IsPartyAssignmentField(field string) bool {
	_, _, ok := partyAssignmentField(field)
	return ok
}

func partyAssignmentField(field string) (string, string, bool) {
	family, member, ok := strings.Cut(field, ".")
	if !ok || (family != "host" && family != "identity") {
		return "", "", false
	}
	switch member {
	case "subject_party_ids":
		return family, "subject_party", true
	case "responsible_party_ids":
		return family, "responsible_party", true
	case "observer_party_ids":
		return family, "observer_party", true
	default:
		return "", "", false
	}
}

// LoadPartyAssignmentsTx reads the caller's source snapshot. The active-link
// relation excludes deleted links and endpoints. Endpoint types remain in the
// projection for validation; a malformed assignment must not silently vanish.
func (FactReader) LoadPartyAssignmentsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) ([]partyassignment.Fact, error) {
	return loadPartyAssignments(ctx, tx, incidentID, nil)
}

// LoadPartyAssignmentsForRecords restricts source hydration to an admitted row window.
func (FactReader) LoadPartyAssignmentsForRecords(ctx context.Context, query interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}, incidentID uuid.UUID, recordIDs []uuid.UUID) ([]partyassignment.Fact, error) {
	if len(recordIDs) == 0 {
		return []partyassignment.Fact{}, nil
	}
	return loadPartyAssignments(ctx, query, incidentID, recordIDs)
}

func loadPartyAssignments(ctx context.Context, query interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}, incidentID uuid.UUID, recordIDs []uuid.UUID) ([]partyassignment.Fact, error) {
	sql := `SELECT record_link_id, incident_id, src_record_id, src_record_type,
dst_record_id, dst_record_type, link_type, field_key, provenance, created_by_user_id, created_at
FROM active_record_links_v1 WHERE incident_id=$1 AND field_key=ANY($2::text[])`
	args := []any{incidentID, partyAssignmentFields}
	if recordIDs != nil {
		sql += ` AND src_record_id=ANY($3::uuid[])`
		args = append(args, recordIDs)
	}
	sql += ` ORDER BY record_link_id`
	rows, err := query.Query(ctx, sql, args...)
	if err != nil {
		return nil, &FactReadError{err: err}
	}
	defer rows.Close()
	result := []partyassignment.Fact{}
	seen := map[[3]string]bool{}
	for rows.Next() {
		var state partyassignment.State
		var destinationType, linkType, fieldKey string
		var createdAt time.Time
		if err := rows.Scan(&state.AssignmentID, &state.IncidentID, &state.SubjectRecordID, &state.SubjectRecordType,
			&state.PartyID, &destinationType, &linkType, &fieldKey, &state.Provenance, &state.CreatedByUserID, &createdAt); err != nil {
			return nil, &FactReadError{err: err}
		}
		subjectType, kind, valid := partyAssignmentField(fieldKey)
		if !valid || state.IncidentID != incidentID || state.SubjectRecordType != subjectType || destinationType != "party" || linkType != LinkTypeReferencesRecord.String() || createdAt.Year() < 1 || createdAt.Year() > 9999 || createdAt.Nanosecond()%1000 != 0 {
			return nil, &FactReadError{err: errors.New("invalid party assignment source")}
		}
		state.AssignmentKind, state.AssignmentState = kind, "active"
		state.CreatedAt = createdAt.UTC().Format("2006-01-02T15:04:05.000000Z")
		fact, err := partyassignment.NewFact(state)
		if err != nil {
			return nil, &FactReadError{err: err}
		}
		key := [3]string{state.SubjectRecordID.String(), state.PartyID.String(), kind}
		if seen[key] {
			return nil, &FactReadError{err: errors.New("duplicate active party assignment")}
		}
		seen[key] = true
		result = append(result, fact)
	}
	if err := rows.Err(); err != nil {
		return nil, &FactReadError{err: err}
	}
	return result, nil
}
