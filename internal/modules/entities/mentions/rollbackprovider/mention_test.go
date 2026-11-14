package rollbackprovider

import (
	"context"
	"errors"
	"maps"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/entities/mentions/historyvalue"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

func TestParseMentionTargetRejectsMalformedIdentity(t *testing.T) {
	t.Parallel()
	mentionID := uuid.New()
	before := completeMentionHistoryValue(mentionID, uuid.New())
	after := maps.Clone(before)
	after["resolution_status"], after["row_version"] = "dismissed", int64(2)
	before["entity_mention_id"] = uuid.NewString()
	target := rollbackcontract.NonRowTarget{
		TargetKind: "entity_mention", TargetID: mentionID.String(), OperationKind: "patch",
		BeforeValue: before, AfterValue: after,
	}
	if _, err := parseMentionTarget(target); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
		t.Fatalf("parseMentionTarget error = %v", err)
	}
}

func TestParseMentionTargetAcceptsRetainedMention(t *testing.T) {
	t.Parallel()
	mentionID, sourceID := uuid.New(), uuid.New()
	before := completeMentionHistoryValue(mentionID, sourceID)
	after := maps.Clone(before)
	after["resolution_status"], after["row_version"] = "dismissed", int64(2)
	target := rollbackcontract.NonRowTarget{
		TargetKind: "entity_mention", TargetID: mentionID.String(), OperationKind: "patch",
		BeforeValue: before, AfterValue: after,
	}
	identity, err := parseMentionTarget(target)
	if err != nil || identity.sourceRecordID != sourceID {
		t.Fatalf("parseMentionTarget = %#v, %v", identity, err)
	}
	for _, status := range []string{"unresolved", "dismissed"} {
		t.Run(status, func(t *testing.T) {
			identity.status = status
			if err := validateResolvedTargetTx(t.Context(), nil, uuid.New(), identity, target.BeforeValue); err != nil {
				t.Fatalf("empty resolution metadata: %v", err)
			}
		})
	}
	for _, test := range []struct {
		name   string
		method any
		valid  bool
	}{
		{"valid method", "explicit_resolve_route", true},
		{"missing method", nil, false},
		{"empty method", "", false},
		{"blank method", " \t", false},
		{"wrong type", 42, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			value := map[string]any{
				"resolved_record_id": uuid.NewString(), "resolved_by_user_id": uuid.NewString(),
				"resolved_at": "2026-11-11T12:00:00Z",
			}
			if test.method != nil {
				value["resolution_method"] = test.method
			}
			identity.status = "resolved"
			tx := &mentionTargetProbe{}
			err := validateResolvedTargetTx(t.Context(), tx, uuid.New(), identity, value)
			if test.valid {
				if err != nil || tx.queries != 1 {
					t.Fatalf("valid resolution: queries=%d error=%v", tx.queries, err)
				}
			} else if !errors.Is(err, rollbackcontract.ErrTargetNotReversible) || tx.queries != 0 {
				t.Fatalf("malformed resolution: queries=%d error=%v", tx.queries, err)
			}
		})
	}
}

func completeMentionHistoryValue(mentionID, sourceID uuid.UUID) map[string]any {
	return (historyvalue.Value{MentionID: mentionID, SourceRecordID: sourceID,
		SourceFieldKey: "timeline.host_refs", EntityType: "host", OriginKind: "manual_entry",
		OriginLocator: "cell", RawText: "HOST-1", NormalizedText: "host-1",
		ResolutionStatus: "unresolved", RowVersion: 1, Ordinal: 1,
		CreatedByUserID: uuid.New(), CreatedAt: time.Date(2026, 11, 11, 12, 0, 0, 0, time.UTC)}).Map()
}

type mentionTargetProbe struct {
	pgx.Tx
	queries int
}

func (tx *mentionTargetProbe) QueryRow(context.Context, string, ...any) pgx.Row {
	tx.queries++
	return mentionTargetPresent{}
}

type mentionTargetPresent struct{}

func (mentionTargetPresent) Scan(dest ...any) error {
	*dest[0].(*bool) = true
	return nil
}
