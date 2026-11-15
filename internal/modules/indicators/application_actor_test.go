package indicators

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
)

func TestIndicatorApplicationRejectsNilActorBeforeDependencies(t *testing.T) {
	t.Parallel()

	application := &Application{}
	incidentID := uuid.MustParse("00000000-0000-4000-8000-000000000711")
	recordID := uuid.MustParse("00000000-0000-4000-8000-000000000712")
	observationID := uuid.MustParse("00000000-0000-4000-8000-000000000713")
	ctx := context.Background()

	// Every input below is otherwise valid, so only the actor guard can return
	// before the zero-value Application reaches a dependency.
	assertRejected := func(name string, typed bool, call func() error) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			err := call()
			if typed {
				var validation *IndicatorCreateValidationError
				if !errors.As(err, &validation) || validation.Field != "actor_user_id" || validation.ReasonCode != "missing_required_field" {
					t.Fatalf("nil actor error = %#v, want actor_user_id missing_required_field", err)
				}
				return
			}
			if !errors.Is(err, ErrInvalidCreateRequest) {
				t.Fatalf("nil actor error = %v, want ErrInvalidCreateRequest", err)
			}
		})
	}

	assertRejected("create row", true, func() error {
		_, err := application.CreateIndicatorRow(ctx, uuid.Nil, incidentID, CreateCommand{
			ClientTxnID:   "txn-nil-actor-create",
			IndicatorType: "ipv4_addr",
			ValueKind:     "atomic",
			DisplayValue:  "192.0.2.1",
		}, "request")
		return err
	})
	assertRejected("create observation", false, func() error {
		_, err := application.CreateIndicatorObservation(ctx, uuid.Nil, IndicatorObservationCreateParams{
			IncidentID:     incidentID,
			SourceRecordID: recordID,
			BaseRowVersion: 1,
			SourceFieldKey: "timeline.source_text",
			RequestID:      "request",
			ClientTxnID:    "txn-nil-actor-observation",
		})
		return err
	})
	assertRejected("resolve observation", false, func() error {
		_, err := application.ResolveIndicatorObservation(ctx, uuid.Nil, IndicatorObservationResolveParams{
			ObservationID:             observationID,
			ResolvedIndicatorRecordID: recordID,
			BaseRowVersion:            1,
			RequestID:                 "request",
			ClientTxnID:               "txn-nil-actor-resolve",
		})
		return err
	})
	assertRejected("dismiss observation", false, func() error {
		_, err := application.DismissIndicatorObservation(ctx, uuid.Nil, IndicatorObservationActionParams{
			ObservationID:  observationID,
			BaseRowVersion: 1,
			RequestID:      "request",
			ClientTxnID:    "txn-nil-actor-dismiss",
		})
		return err
	})
	assertRejected("restore observation", false, func() error {
		_, err := application.RestoreIndicatorObservation(ctx, uuid.Nil, IndicatorObservationActionParams{
			ObservationID:  observationID,
			BaseRowVersion: 1,
			RequestID:      "request",
			ClientTxnID:    "txn-nil-actor-restore",
		})
		return err
	})
	assertRejected("append lifecycle", false, func() error {
		_, err := application.AppendIndicatorLifecycleInterval(ctx, uuid.Nil, IndicatorLifecycleAppendParams{
			IncidentID:        incidentID,
			IndicatorRecordID: recordID,
			BaseRowVersion:    1,
			LifecycleState:    "active",
			ValidFrom:         time.Unix(1, 0),
			RequestID:         "request",
			ClientTxnID:       "txn-nil-actor-lifecycle",
		})
		return err
	})
	assertRejected("transaction participant", true, func() error {
		_, err := application.FindOrCreateIndicatorParticipantTx(ctx, nil, IndicatorFindOrCreateParticipantCommand{
			IncidentID:        incidentID,
			IndicatorType:     "ipv4_addr",
			ValueKind:         "atomic",
			DisplayValue:      "192.0.2.1",
			OperationContext:  "networkflow.indicator_link",
			OperationOccurred: time.Unix(1, 0),
		})
		return err
	})
	assertRejected("import contribution", true, func() error {
		_, err := application.createImportRowTx(ctx, nil, ownerfacade.ImportOwnerCreateCommand{
			Request: ownerfacade.ImportOwnerCreateRequest{
				IncidentID:         incidentID,
				TargetViewSchemaID: ViewSchemaID,
			},
		})
		return err
	})
}
