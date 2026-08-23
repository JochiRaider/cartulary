package indicators

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/records"
)

func TestIndicatorTargetEnvelopeRoleClassification(t *testing.T) {
	t.Parallel()
	incidentID := uuid.New()
	otherIncidentID := uuid.New()
	sourceID := uuid.New()
	indicatorID := uuid.New()
	deletedAt := time.Date(2026, 8, 3, 11, 0, 0, 0, time.UTC)
	envelopes := map[uuid.UUID]records.Envelope{
		sourceID:    {RecordID: sourceID, IncidentID: incidentID, RecordType: "timeline_event", RowVersion: 7},
		indicatorID: {RecordID: indicatorID, IncidentID: incidentID, RecordType: "indicator", RowVersion: 3},
	}

	if err := validateObservationSourceEnvelope(envelopes, incidentID, sourceID, 7); err != nil {
		t.Fatalf("valid source envelope: %v", err)
	}
	if err := validateResolvedIndicatorEnvelope(envelopes, incidentID, indicatorID); err != nil {
		t.Fatalf("valid resolved Indicator envelope: %v", err)
	}
	if err := validateAddressedIndicatorEnvelope(envelopes, incidentID, indicatorID); err != nil {
		t.Fatalf("valid addressed Indicator envelope: %v", err)
	}
	if err := validateLifecycleSupportEnvelopes(envelopes, incidentID, []uuid.UUID{sourceID, indicatorID}); err != nil {
		t.Fatalf("valid support envelopes: %v", err)
	}

	tests := []struct {
		name      string
		envelope  records.Envelope
		validate  func(map[uuid.UUID]records.Envelope) error
		wantError error
	}{
		{name: "missing source", validate: func(values map[uuid.UUID]records.Envelope) error {
			return validateObservationSourceEnvelope(values, incidentID, uuid.New(), 1)
		}, wantError: ErrIndicatorSourceNotFound},
		{name: "foreign source", envelope: records.Envelope{RecordID: sourceID, IncidentID: otherIncidentID, RecordType: "timeline_event", RowVersion: 7}, validate: func(values map[uuid.UUID]records.Envelope) error {
			return validateObservationSourceEnvelope(values, incidentID, sourceID, 7)
		}, wantError: ErrIndicatorSourceNotFound},
		{name: "deleted source", envelope: records.Envelope{RecordID: sourceID, IncidentID: incidentID, RecordType: "timeline_event", RowVersion: 7, DeletedAt: &deletedAt}, validate: func(values map[uuid.UUID]records.Envelope) error {
			return validateObservationSourceEnvelope(values, incidentID, sourceID, 7)
		}, wantError: ErrIndicatorSourceNotFound},
		{name: "stale source", envelope: envelopes[sourceID], validate: func(values map[uuid.UUID]records.Envelope) error {
			return validateObservationSourceEnvelope(values, incidentID, sourceID, 6)
		}, wantError: ErrRowVersionConflict},
		{name: "wrong resolved target type", envelope: records.Envelope{RecordID: indicatorID, IncidentID: incidentID, RecordType: "timeline_event"}, validate: func(values map[uuid.UUID]records.Envelope) error {
			return validateResolvedIndicatorEnvelope(values, incidentID, indicatorID)
		}, wantError: ErrResolvedIndicatorNotFound},
		{name: "foreign resolved target", envelope: records.Envelope{RecordID: indicatorID, IncidentID: otherIncidentID, RecordType: "indicator"}, validate: func(values map[uuid.UUID]records.Envelope) error {
			return validateResolvedIndicatorEnvelope(values, incidentID, indicatorID)
		}, wantError: ErrResolvedIndicatorNotFound},
		{name: "wrong addressed target type", envelope: records.Envelope{RecordID: indicatorID, IncidentID: incidentID, RecordType: "timeline_event"}, validate: func(values map[uuid.UUID]records.Envelope) error {
			return validateAddressedIndicatorEnvelope(values, incidentID, indicatorID)
		}, wantError: ErrIndicatorNotFound},
		{name: "foreign addressed target", envelope: records.Envelope{RecordID: indicatorID, IncidentID: otherIncidentID, RecordType: "indicator"}, validate: func(values map[uuid.UUID]records.Envelope) error {
			return validateAddressedIndicatorEnvelope(values, incidentID, indicatorID)
		}, wantError: ErrIndicatorNotFound},
		{name: "foreign support", envelope: records.Envelope{RecordID: sourceID, IncidentID: otherIncidentID, RecordType: "timeline_event"}, validate: func(values map[uuid.UUID]records.Envelope) error {
			return validateLifecycleSupportEnvelopes(values, incidentID, []uuid.UUID{sourceID})
		}, wantError: ErrInvalidCreateRequest},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			values := map[uuid.UUID]records.Envelope{}
			if test.envelope.RecordID != uuid.Nil {
				values[test.envelope.RecordID] = test.envelope
			}
			if err := test.validate(values); !errors.Is(err, test.wantError) {
				t.Fatalf("role classification error = %v, want %v", err, test.wantError)
			}
		})
	}

	prior := IndicatorObservationRecord{IncidentID: incidentID, SourceRecordID: sourceID, ResolvedIndicatorRecordID: &indicatorID}
	for name, values := range map[string]map[uuid.UUID]records.Envelope{
		"missing source":    {indicatorID: envelopes[indicatorID]},
		"missing Indicator": {sourceID: envelopes[sourceID]},
		"wrong Indicator": {
			sourceID:    envelopes[sourceID],
			indicatorID: {RecordID: indicatorID, IncidentID: incidentID, RecordType: "timeline_event"},
		},
	} {
		if err := validatePriorObservationDependencies(values, prior); !errors.Is(err, ErrIndicatorObservationNotFound) {
			t.Fatalf("%s prior dependency error = %v, want ErrIndicatorObservationNotFound", name, err)
		}
	}
}

func TestIndicatorLockedEnvelopeStorageFailurePropagates(t *testing.T) {
	t.Parallel()
	want := errors.New("injected record-envelope storage failure")
	store := &Store{recordStore: failingIndicatorRecordStore{err: want}}
	if _, err := store.lockAffectedRecordsTx(context.Background(), nil, []uuid.UUID{uuid.New()}); !errors.Is(err, want) {
		t.Fatalf("locked envelope failure = %v, want injected storage failure", err)
	}
}

type failingIndicatorRecordStore struct {
	err error
}

func (store failingIndicatorRecordStore) InsertTx(context.Context, pgx.Tx, records.InsertParams) (uuid.UUID, error) {
	panic("unexpected InsertTx")
}

func (store failingIndicatorRecordStore) LoadEnvelopesTx(context.Context, pgx.Tx, []uuid.UUID, bool) (map[uuid.UUID]records.Envelope, error) {
	return nil, store.err
}

func (store failingIndicatorRecordStore) AdvanceVersionTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time) (int64, error) {
	panic("unexpected AdvanceVersionTx")
}

func TestIndicatorLifecycleVocabulary(t *testing.T) {
	t.Parallel()
	for _, state := range []string{"active", "benign", "false_positive", "retired"} {
		state := state
		t.Run("accept/"+state, func(t *testing.T) {
			t.Parallel()
			params := validLifecycleContractParams(state)
			if err := normalizeLifecycleAppendParams(&params); err != nil {
				t.Fatalf("normalize %q: %v", state, err)
			}
		})
	}
	for _, state := range []string{"", " active", "active ", "ACTIVE", "false-positive", "inactive", "dismissed"} {
		state := state
		t.Run("reject/"+state, func(t *testing.T) {
			t.Parallel()
			params := validLifecycleContractParams(state)
			if err := normalizeLifecycleAppendParams(&params); !errors.Is(err, ErrInvalidCreateRequest) {
				t.Fatalf("normalize %q = %v, want ErrInvalidCreateRequest", state, err)
			}
		})
	}
}

func TestIndicatorSourceSpanAdmission(t *testing.T) {
	t.Parallel()
	text := "α.example"
	selected, err := sourceSpan(text, 0, len("α"))
	if err != nil || selected != "α" {
		t.Fatalf("valid UTF-8 span = %q, %v", selected, err)
	}
	for _, span := range []struct {
		name       string
		text       string
		start, end int
	}{
		{name: "negative start", text: text, start: -1, end: 2},
		{name: "empty", text: text, start: 2, end: 2},
		{name: "past end", text: text, start: 0, end: len(text) + 1},
		{name: "split start rune", text: text, start: 1, end: 2},
		{name: "split end rune", text: text, start: 0, end: 1},
		{name: "invalid source UTF-8", text: string([]byte{0xff}), start: 0, end: 1},
		{name: "NUL selected", text: "a\x00b", start: 0, end: 3},
	} {
		span := span
		t.Run(span.name, func(t *testing.T) {
			t.Parallel()
			if _, err := sourceSpan(span.text, span.start, span.end); !errors.Is(err, ErrInvalidCreateRequest) {
				t.Fatalf("sourceSpan(%q, %d, %d) = %v, want ErrInvalidCreateRequest", span.text, span.start, span.end, err)
			}
		})
	}
}

func TestIndicatorObservationTransitionStateMachine(t *testing.T) {
	t.Parallel()
	target := uuid.New()
	otherTarget := uuid.New()
	actor := uuid.New()
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	resolved := func(id uuid.UUID) IndicatorObservationRecord {
		return IndicatorObservationRecord{ResolutionStatus: "resolved", ResolvedIndicatorRecordID: &id, RowVersion: 4}
	}
	tests := []struct {
		name       string
		current    IndicatorObservationRecord
		transition observationTransition
		target     uuid.UUID
		wantStatus string
		wantError  bool
	}{
		{name: "resolve unresolved", current: IndicatorObservationRecord{ResolutionStatus: "unresolved", RowVersion: 1}, transition: observationTransitionResolve, target: target, wantStatus: "resolved"},
		{name: "resolve retargets", current: resolved(otherTarget), transition: observationTransitionResolve, target: target, wantStatus: "resolved"},
		{name: "resolve same target", current: resolved(target), transition: observationTransitionResolve, target: target, wantError: true},
		{name: "resolve dismissed", current: IndicatorObservationRecord{ResolutionStatus: "dismissed"}, transition: observationTransitionResolve, target: target, wantError: true},
		{name: "dismiss unresolved", current: IndicatorObservationRecord{ResolutionStatus: "unresolved", RowVersion: 1}, transition: observationTransitionDismiss, wantStatus: "dismissed"},
		{name: "dismiss resolved", current: resolved(target), transition: observationTransitionDismiss, wantStatus: "dismissed"},
		{name: "dismiss dismissed", current: IndicatorObservationRecord{ResolutionStatus: "dismissed"}, transition: observationTransitionDismiss, wantError: true},
		{name: "restore dismissed", current: IndicatorObservationRecord{ResolutionStatus: "dismissed", RowVersion: 2}, transition: observationTransitionRestore, wantStatus: "unresolved"},
		{name: "restore unresolved", current: IndicatorObservationRecord{ResolutionStatus: "unresolved"}, transition: observationTransitionRestore, wantError: true},
		{name: "restore resolved", current: resolved(target), transition: observationTransitionRestore, wantError: true},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			next, err := nextObservationTransition(test.current, test.transition, test.target, actor, now)
			if test.wantError {
				if !errors.Is(err, ErrIllegalTransition) {
					t.Fatalf("transition error = %v, want ErrIllegalTransition", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("transition: %v", err)
			}
			if next.ResolutionStatus != test.wantStatus || next.RowVersion != test.current.RowVersion+1 {
				t.Fatalf("transition = %#v, want status %q and version %d", next, test.wantStatus, test.current.RowVersion+1)
			}
		})
	}
}

func TestCreateCommandProjectsIdentityValidation(t *testing.T) {
	t.Parallel()
	_, err := indicatorInputFromCreateCommand(CreateCommand{
		IndicatorType: "domain",
		ValueKind:     "atomic",
		DisplayValue:  "example.test",
	})
	var validation *IndicatorCreateValidationError
	if !errors.As(err, &validation) || validation.Field != "indicator.indicator_type" || validation.ReasonCode != "invalid_value" {
		t.Fatalf("validation = %#v, want indicator.indicator_type/invalid_value", err)
	}

	_, err = indicatorInputFromCreateCommand(CreateCommand{
		IndicatorType: "sha256",
		ValueKind:     "atomic",
		DisplayValue:  strings.Repeat("a", 64),
		HashAlgorithm: stringPointer("sha256"),
	})
	if !errors.As(err, &validation) || validation.Field != "indicator.hash_value" || validation.ReasonCode != "invalid_value" {
		t.Fatalf("hash validation = %#v, want indicator.hash_value/invalid_value", err)
	}
}

func validLifecycleContractParams(state string) IndicatorLifecycleAppendParams {
	return IndicatorLifecycleAppendParams{
		IncidentID:        uuid.New(),
		IndicatorRecordID: uuid.New(),
		LifecycleState:    state,
		ValidFrom:         time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC),
		SupportRefs:       []uuid.UUID{},
	}
}
