package indicators

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

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
