package indicators

import (
	"context"
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

	assertRejected := func(name string, call func() error) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			if err := call(); err == nil {
				t.Fatal("nil actor was accepted")
			}
		})
	}

	assertRejected("create row", func() error {
		_, err := application.CreateIndicatorRow(ctx, uuid.Nil, incidentID, CreateCommand{}, "request")
		return err
	})
	assertRejected("create observation", func() error {
		_, err := application.CreateIndicatorObservation(ctx, uuid.Nil, IndicatorObservationCreateParams{})
		return err
	})
	assertRejected("resolve observation", func() error {
		_, err := application.ResolveIndicatorObservation(ctx, uuid.Nil, IndicatorObservationResolveParams{ObservationID: observationID, ResolvedIndicatorRecordID: recordID})
		return err
	})
	assertRejected("dismiss observation", func() error {
		_, err := application.DismissIndicatorObservation(ctx, uuid.Nil, IndicatorObservationActionParams{ObservationID: observationID})
		return err
	})
	assertRejected("restore observation", func() error {
		_, err := application.RestoreIndicatorObservation(ctx, uuid.Nil, IndicatorObservationActionParams{ObservationID: observationID})
		return err
	})
	assertRejected("append lifecycle", func() error {
		_, err := application.AppendIndicatorLifecycleInterval(ctx, uuid.Nil, IndicatorLifecycleAppendParams{})
		return err
	})
	assertRejected("transaction participant", func() error {
		_, err := application.FindOrCreateIndicatorParticipantTx(ctx, nil, IndicatorFindOrCreateParticipantCommand{
			IncidentID:        incidentID,
			OperationContext:  "networkflow.indicator_link",
			OperationOccurred: time.Unix(1, 0),
		})
		return err
	})
	assertRejected("import contribution", func() error {
		_, err := application.createImportRowTx(ctx, nil, ownerfacade.ImportOwnerCreateCommand{})
		return err
	})
}
