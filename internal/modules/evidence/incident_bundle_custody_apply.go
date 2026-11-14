package evidence

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/jackc/pgx/v5"
)

func (p preparedEvidenceImport) custodyReferences(ctx context.Context, tx pgx.Tx, d sourceport.Descriptor) error {
	return p.custody(ctx, d, func(row portableCustodyRow) error {
		var valid bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM evidence e JOIN records r ON r.record_id=e.record_id WHERE e.record_id=$1 AND e.incident_id=$2 AND r.incident_id=$2 AND r.record_type='evidence')`, row.EvidenceRecordID, p.incidentID).Scan(&valid); err != nil {
			return err
		}
		if !valid {
			return d.DeclaredFailure("evidence.custody_same_incident")
		}
		return nil
	})
}
func (p preparedEvidenceImport) insertCustody(ctx context.Context, tx pgx.Tx, input sourceport.ImportContext, d sourceport.Descriptor) error {
	if err := p.custodyReferences(ctx, tx, d); err != nil {
		return err
	}
	if err := p.custody(ctx, d, func(row portableCustodyRow) error {
		var exists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM evidence_custody_events WHERE custody_event_id=$1)`, row.CustodyEventID).Scan(&exists); err != nil {
			return err
		}
		if exists {
			return d.DeclaredFailure("evidence.source_identity_admitted")
		}
		return nil
	}); err != nil {
		return err
	}
	return p.custody(ctx, d, func(row portableCustodyRow) error {
		clock, ok := portableSourceClock(row.OccurredAt)
		if !ok {
			return d.DeclaredFailure("evidence.custody_ordered")
		}
		actor := row.ActorUserID
		if actor != nil {
			actor = &p.actorID
		}
		tag, err := tx.Exec(ctx, `INSERT INTO evidence_custody_events(custody_event_id,incident_id,evidence_record_id,custody_event_type,actor_user_id,occurred_at,location_text,note,metadata) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9::jsonb)`, row.CustodyEventID, row.IncidentID, row.EvidenceRecordID, row.CustodyEventType, actor, clock, row.LocationText, row.Note, []byte(row.Metadata))
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return d.DeclaredFailure("evidence.source_identity_admitted")
		}
		if row.ActorUserID != nil {
			return input.Attributions.RecordImportedAttribution("evidence_custody_events", row.CustodyEventID.String(), "actor_user_id", row.ActorUserID.String())
		}
		return nil
	})
}
func (p preparedEvidenceImport) compareCustody(ctx context.Context, tx pgx.Tx, d sourceport.Descriptor) error {
	if err := p.custodyReferences(ctx, tx, d); err != nil {
		return err
	}
	rows, err := tx.Query(ctx, `SELECT custody_event_id,incident_id,evidence_record_id,custody_event_type,actor_user_id,occurred_at,location_text,note,metadata FROM evidence_custody_events WHERE incident_id=$1 ORDER BY custody_event_id`, p.incidentID)
	if err != nil {
		return err
	}
	defer rows.Close()
	err = p.custody(ctx, d, func(expected portableCustodyRow) error {
		if !rows.Next() {
			if err := rows.Err(); err != nil {
				return err
			}
			return d.DeclaredFailure("evidence.source_state_exact")
		}
		var actual portableCustodyRow
		var clock time.Time
		var raw []byte
		if err := rows.Scan(&actual.CustodyEventID, &actual.IncidentID, &actual.EvidenceRecordID, &actual.CustodyEventType, &actual.ActorUserID, &clock, &actual.LocationText, &actual.Note, &raw); err != nil {
			return err
		}
		var metadata any
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.UseNumber()
		if err := decoder.Decode(&metadata); err != nil {
			return err
		}
		normalized, ok := custodyJSON(metadata)
		if !ok {
			return d.DeclaredFailure("evidence.source_state_exact")
		}
		actual.Metadata = normalized
		actual.OccurredAt = clock.UTC().Format(time.RFC3339Nano)
		expectedClock, ok := portableSourceClock(expected.OccurredAt)
		if !ok {
			return d.DeclaredFailure("evidence.custody_ordered")
		}
		expected.OccurredAt = expectedClock.Format(time.RFC3339Nano)
		if expected.ActorUserID != nil {
			expected.ActorUserID = &p.actorID
		}
		if !reflect.DeepEqual(actual, expected) {
			return d.DeclaredFailure("evidence.source_state_exact")
		}
		return nil
	})
	if err != nil {
		return err
	}
	if rows.Next() {
		return d.DeclaredFailure("evidence.source_state_exact")
	}
	return rows.Err()
}
func (p preparedEvidenceImport) validateCustodyAttributions(ctx context.Context, recorder incidentportability.AttributionRecorder, d sourceport.Descriptor) error {
	var expectedCount, actualCount uint64
	if err := p.custody(ctx, d, func(row portableCustodyRow) error {
		present := row.ActorUserID != nil
		if present {
			expectedCount++
		}
		if err := recorder.VisitImportedAttributions("evidence_custody_events", row.CustodyEventID.String(), func(a incidentportability.ImportedAttribution) error {
			if !present || a.SourceTable != "evidence_custody_events" || a.SourceRowID != row.CustodyEventID.String() || a.SourceColumn != "actor_user_id" || a.SourceActorID != row.ActorUserID.String() || a.LocalUserID != p.actorID {
				return d.DeclaredFailure("evidence.attribution_exact")
			}
			present = false
			return ctx.Err()
		}); err != nil {
			return err
		}
		if present {
			return d.DeclaredFailure("evidence.attribution_exact")
		}
		return nil
	}); err != nil {
		return err
	}
	if err := recorder.VisitImportedAttributions("evidence_custody_events", "", func(incidentportability.ImportedAttribution) error { actualCount++; return ctx.Err() }); err != nil {
		return err
	}
	if actualCount != expectedCount {
		return d.DeclaredFailure("evidence.attribution_exact")
	}
	return nil
}
