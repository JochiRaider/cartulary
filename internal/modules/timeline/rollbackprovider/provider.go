package rollbackprovider

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

type TimelineProvider struct{}

var _ rollbackcontract.RowSourceProvider = TimelineProvider{}

func NewTimelineProvider() TimelineProvider {
	return TimelineProvider{}
}

func (TimelineProvider) ValidateRollbackValue(value map[string]any) error {
	_, ok, err := sourceForRollbackValue(value)
	if err != nil {
		return err
	}
	if !ok {
		return rollbackcontract.ErrTargetNotReversible
	}
	return nil
}

func (TimelineProvider) RestoreTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.RestoreRequest) error {
	source, ok, err := sourceForRollbackValue(request.RetainedValue)
	if err != nil {
		return err
	}
	if !ok {
		return rollbackcontract.ErrTargetNotReversible
	}
	return updateSourceTx(ctx, tx, request.RecordID, request.ActorUserID, request.Now, request.NextRowVersion, source)
}

func sourceForRollbackValue(value map[string]any) (map[string]any, bool, error) {
	if source, ok := objectMap(value, "source"); ok {
		return source, len(source) > 0, nil
	}
	return nil, false, nil
}

func updateSourceTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, actorUserID uuid.UUID, now time.Time, rowVersion int64, source map[string]any) error {
	captureState, ok := stringFromMap(source, "capture_state")
	if !ok || captureState == "" {
		return rollbackcontract.ErrTargetNotReversible
	}
	pairState, ok := stringFromMap(source, "activity_time_pair_state")
	if !ok || pairState == "" {
		pairState = "disabled"
	}
	_, err := tx.Exec(ctx, `
UPDATE timeline_events
   SET date_entered_text = $2,
       analyst_text = $3,
       mitre_stage_text = $4,
       device_object_text = $5,
       ip_address_text = $6,
       activity_utc_text = $7,
       activity_local_text = $8,
       raw_activity_text = $9,
       activity_synopsis_text = $10,
       data_source_text = $11,
       activity_utc_generated = false,
       activity_local_generated = false,
       activity_time_pair_state = $12,
       capture_state = $13,
       row_version = $14,
       edited_at = $15,
       updated_by_user_id = $16,
       reviewed_by_user_id = $17,
       reviewed_at = $18,
       superseded_by_user_id = $19,
       superseded_at = $20
 WHERE record_id = $1
`, recordID,
		nullableStringAny(source, "date_entered_text"),
		nullableStringAny(source, "analyst_text"),
		nullableStringAny(source, "mitre_stage_text"),
		nullableStringAny(source, "device_object_text"),
		nullableStringAny(source, "ip_address_text"),
		nullableStringAny(source, "activity_utc_text"),
		nullableStringAny(source, "activity_local_text"),
		nullableStringAny(source, "raw_activity_text"),
		nullableStringAny(source, "activity_synopsis_text"),
		nullableStringAny(source, "data_source_text"),
		pairState,
		captureState,
		rowVersion,
		now.UTC(),
		actorUserID,
		nullableUUIDAny(source, "reviewed_by_user_id"),
		nullableAny(source, "reviewed_at"),
		nullableUUIDAny(source, "superseded_by_user_id"),
		nullableAny(source, "superseded_at"))
	return err
}

func objectMap(value map[string]any, key string) (map[string]any, bool) {
	raw, ok := value[key]
	if !ok || raw == nil {
		return nil, false
	}
	typed, ok := raw.(map[string]any)
	return typed, ok
}

func stringFromMap(value map[string]any, key string) (string, bool) {
	raw, ok := value[key]
	if !ok || raw == nil {
		return "", false
	}
	text, ok := raw.(string)
	return text, ok
}

func nullableStringAny(value map[string]any, key string) any {
	if raw, ok := value[key]; ok {
		return raw
	}
	return nil
}

func nullableAny(value map[string]any, key string) any {
	if raw, ok := value[key]; ok {
		return raw
	}
	return nil
}

func nullableUUIDAny(value map[string]any, key string) any {
	raw, ok := value[key]
	if !ok || raw == nil {
		return nil
	}
	text, ok := raw.(string)
	if !ok || strings.TrimSpace(text) == "" {
		return nil
	}
	parsed, err := uuid.Parse(text)
	if err != nil {
		return nil
	}
	return parsed
}
