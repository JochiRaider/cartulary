package revisions

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
)

const (
	referenceProbeType = "reference_probe"
	otherProbeType     = "other_probe"
)

// TestIncidentBundleReferencePassReadsByPage_Unit: the reference pass reads
// owners and envelopes once per page of referenced records, never once per
// record. It still judges records in order: verdicts before the first failing
// record are kept, and that record's references failure is returned.
func TestIncidentBundleReferencePassReadsByPage_Unit(t *testing.T) {
	incidentID := uuid.New()
	t.Run("reads are bounded by pages", func(t *testing.T) {
		probe := newReferenceProbe(incidentID)
		var keys []string
		for range 300 {
			recordID := probe.add(referenceProbeType)
			keys = append(keys, referenceKey(recordID, "r", ""), referenceKey(recordID, "t", referenceProbeType))
		}
		for range 20 {
			keys = append(keys, referenceKey(probe.add(referenceProbeType), "r", ""))
		}
		outcome, err := probe.check(t, keys)
		if err != nil || outcome != (referenceOutcome{}) {
			t.Fatalf("valid references = %+v, %v", outcome, err)
		}
		const pages = 3 // 320 referenced records, read 128 at a time
		if probe.single != 0 || probe.ownerReads > 2*pages || probe.envelopeReads > pages {
			t.Fatalf("320 references used %d single reads, %d owner reads and %d envelope reads; want 0, at most %d and at most %d",
				probe.single, probe.ownerReads, probe.envelopeReads, 2*pages, pages)
		}
	})
	t.Run("records are judged in order", func(t *testing.T) {
		probe := newReferenceProbe(incidentID)
		var ordered []uuid.UUID
		for range 200 {
			ordered = append(ordered, probe.add(referenceProbeType))
		}
		slices.SortFunc(ordered, func(left, right uuid.UUID) int { return slices.Compare(left[:], right[:]) })
		// The earliest record is a declared-type mismatch, a later-rank verdict;
		// a middle record is missing, the references failure.
		probe.envelopes[ordered[0]] = otherProbeType
		delete(probe.rows, ordered[150])
		delete(probe.envelopes, ordered[150])
		var keys []string
		for _, recordID := range ordered {
			keys = append(keys, referenceKey(recordID, "r", ""), referenceKey(recordID, "t", referenceProbeType))
		}
		outcome, err := probe.check(t, keys)
		var failure *sourceport.Failure
		if !errors.As(err, &failure) || failure.InvariantID() != revisionsReferencesInvariant {
			t.Fatalf("missing record error = %v, want the references failure", err)
		}
		if !outcome.typeMismatch {
			t.Fatal("type mismatch judged before the failure was lost")
		}
	})
	t.Run("declared type mismatch resolves through the envelope", func(t *testing.T) {
		probe := newReferenceProbe(incidentID)
		recordID := probe.add(referenceProbeType)
		outcome, err := probe.check(t, []string{referenceKey(recordID, "r", ""), referenceKey(recordID, "t", otherProbeType)})
		if err != nil || !outcome.typeMismatch || outcome.terminalMismatch {
			t.Fatalf("declared mismatch = %+v, %v", outcome, err)
		}
	})
}

type referenceProbe struct {
	incidentID    uuid.UUID
	rows          map[uuid.UUID]map[string]any
	envelopes     map[uuid.UUID]string
	single        int
	ownerReads    int
	envelopeReads int
}

func newReferenceProbe(incidentID uuid.UUID) *referenceProbe {
	return &referenceProbe{incidentID: incidentID, rows: map[uuid.UUID]map[string]any{}, envelopes: map[uuid.UUID]string{}}
}

func (p *referenceProbe) add(recordType string) uuid.UUID {
	recordID := uuid.New()
	p.rows[recordID] = map[string]any{
		"record": map[string]any{"record_id": recordID.String(), "incident_id": p.incidentID.String(), "record_type": recordType},
		"source": map[string]any{"record_id": recordID.String()},
	}
	p.envelopes[recordID] = recordType
	return recordID
}

func (p *referenceProbe) check(t *testing.T, keys []string) (referenceOutcome, error) {
	t.Helper()
	store := portabilitysupport.NewPreparationStore()
	stager := incidentportability.NewSortedKeyStager(t.Context(), store, "prepared/references/", 0)
	for _, key := range keys {
		if err := stager.Add(key); err != nil {
			t.Fatal(err)
		}
	}
	runs, err := stager.Finish()
	if err != nil {
		t.Fatal(err)
	}
	validation := &incidentBundleValidationCatalog{
		envelopes: p,
		snapshots: &RecordSnapshotCaptureCatalog{byRecordType: map[string]recordSnapshotCapture{
			referenceProbeType: {recordType: referenceProbeType, source: referenceProbeSource{probe: p, recordType: referenceProbeType}},
			otherProbeType:     {recordType: otherProbeType, source: referenceProbeSource{probe: p, recordType: otherProbeType}},
		}},
		targets: &TargetSemanticsCatalog{},
	}
	prepared := preparedRevisionsImport{incidentID: p.incidentID, references: runs}
	return prepared.checkReferencesTx(t.Context(), nil, store, validation, nil)
}

func (p *referenceProbe) RecordTypesTx(_ context.Context, _ pgx.Tx, incidentID uuid.UUID, recordIDs []uuid.UUID) (map[uuid.UUID]string, error) {
	p.envelopeReads++
	types := map[uuid.UUID]string{}
	for _, recordID := range recordIDs {
		if recordType, found := p.envelopes[recordID]; found && incidentID == p.incidentID {
			types[recordID] = recordType
		}
	}
	return types, nil
}

func (p *referenceProbe) VisitIncidentRecordIDsTx(context.Context, pgx.Tx, uuid.UUID, func(uuid.UUID) error) error {
	return errors.New("the reference pass must not enumerate the incident")
}

// referenceProbeSource owns the probe records whose envelope has its type.
type referenceProbeSource struct {
	probe      *referenceProbe
	recordType string
}

func (s referenceProbeSource) SnapshotTx(_ context.Context, _ pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	s.probe.single++
	if row, found := s.probe.rows[recordID]; found && s.probe.envelopes[recordID] == s.recordType {
		return row, nil
	}
	return nil, pgx.ErrNoRows
}

func (s referenceProbeSource) SnapshotsTx(_ context.Context, _ pgx.Tx, recordIDs []uuid.UUID) (map[uuid.UUID]map[string]any, error) {
	s.probe.ownerReads++
	rows := map[uuid.UUID]map[string]any{}
	for _, recordID := range recordIDs {
		if row, found := s.probe.rows[recordID]; found && s.probe.envelopes[recordID] == s.recordType {
			rows[recordID] = row
		}
	}
	return rows, nil
}

func (referenceProbeSource) UpdateSourceDeleteStateTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time, bool) error {
	return nil
}

func (referenceProbeSource) ViewSchemaID(context.Context, pgx.Tx, uuid.UUID) (string, error) {
	return "cartulary.view.reference_probe.v1", nil
}

func (referenceProbeSource) PrepareStateTransitionTx(context.Context, pgx.Tx, deleterestorecontract.StateTransitionRequest) (deleterestorecontract.StateTransitionPreparation, error) {
	return deleterestorecontract.StateTransitionPreparation{}, nil
}

func (referenceProbeSource) ValidateStateTransitionTx(context.Context, pgx.Tx, deleterestorecontract.StateTransitionRequest) (deleterestorecontract.StateTransitionPreparation, error) {
	return deleterestorecontract.StateTransitionPreparation{}, nil
}
