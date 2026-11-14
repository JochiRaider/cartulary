package incidentbundles

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/importfinalizerport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestImportSequenceRepairFollowsAllOwnerValidation_Unit(t *testing.T) {
	stages := []string{
		"sequence.begin", "incident.lookup", "incident.apply", "references.apply", "artifacts.apply",
		"revisions.apply", "saved_views.apply", "incident.validate", "revisions.validate",
		"saved_views.validate", "sequence.finish", "projections", "finalize",
	}
	for _, failAt := range append([]string{""}, stages...) {
		name := failAt
		if name == "" {
			name = "success"
		}
		t.Run(name, func(t *testing.T) {
			trace := &sequenceOrderTrace{
				t: t, ctx: t.Context(), failAt: failAt, failure: errors.New("injected coordinator stage failure"),
				incident:  uuid.MustParse("40000000-0000-4000-8000-000000000001"),
				actor:     uuid.MustParse("40000000-0000-4000-8000-000000000002"),
				published: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
			}
			tx := &sequenceOrderTx{trace: trace}
			trace.tx = tx
			prepared := &preparedImport{
				IncidentID:    trace.incident,
				members:       portabilitysupport.MapBundle{"data/actors.ndjson": {}},
				importContext: sourceport.ImportContext{IncidentID: trace.incident, OperationID: "sequence-order"},
				artifacts:     trace,
			}
			for _, family := range []string{"incident", "revisions", "saved_views"} {
				prepared.sourcePreparations = append(prepared.sourcePreparations, preparedSource{
					port: sequenceOrderPort{family: family, trace: trace},
				})
			}
			coordinator := importer{
				referencePacks: sequenceOrderReferences{trace: trace}, projectionRebuild: trace, finalizer: trace,
			}
			got, err := coordinator.applyPreparedImportTx(trace.ctx, tx, prepared, importParams{
				ActorUserID: trace.actor, PublishedAt: trace.published,
			})
			wantStages := stages
			if failAt == "" {
				if err != nil || got != trace.incident {
					t.Fatalf("completed import = %s, %v; want expected incident", got, err)
				}
			} else {
				wantStages = stages[:slices.Index(stages, failAt)+1]
				if got != uuid.Nil {
					t.Fatal("failed import returned an incident identity")
				}
				if failAt == "sequence.begin" || failAt == "sequence.finish" {
					var failure *verificationError
					if !errors.As(err, &failure) || failure.ReasonCode != "source_family_invalid" || failure.SourceFamilyID != "revisions" || failure.InvariantID != "revisions.sequence_repair_after_validation" {
						t.Fatalf("sequence failure = %#v, %v", failure, err)
					}
				} else if !errors.Is(err, trace.failure) {
					t.Fatalf("stage failure was lost: %v", err)
				}
			}
			if !slices.Equal(trace.events, wantStages) {
				t.Fatalf("coordinator stages = %v, want %v", trace.events, wantStages)
			}
		})
	}
}

type sequenceOrderTrace struct {
	t               *testing.T
	ctx             context.Context
	tx              pgx.Tx
	failAt          string
	failure         error
	events          []string
	incident, actor uuid.UUID
	published       time.Time
}

func (s *sequenceOrderTrace) step(ctx context.Context, tx pgx.Tx, stage string) error {
	s.t.Helper()
	if ctx != s.ctx || tx != s.tx {
		s.t.Fatalf("%s escaped the publication context or transaction", stage)
	}
	s.events = append(s.events, stage)
	if stage == s.failAt {
		return s.failure
	}
	return nil
}

func (s *sequenceOrderTrace) ApplyTx(ctx context.Context, tx pgx.Tx) error {
	return s.step(ctx, tx, "artifacts.apply")
}
func (s *sequenceOrderTrace) RebuildImportedIncidentTx(ctx context.Context, tx pgx.Tx, incident uuid.UUID) error {
	if incident != s.incident {
		s.t.Fatal("projection rebuild changed incident identity")
	}
	return s.step(ctx, tx, "projections")
}
func (s *sequenceOrderTrace) FinalizeIncidentBundleImportTx(ctx context.Context, tx pgx.Tx, params importfinalizerport.Params) error {
	if params.IncidentID != s.incident || params.SubmittedByUserID != s.actor || !params.PublishedAt.Equal(s.published) {
		s.t.Fatal("finalizer changed publication identity, actor or time")
	}
	return s.step(ctx, tx, "finalize")
}

type sequenceOrderReferences struct {
	reference_data.IncidentReferences
	trace *sequenceOrderTrace
}

func (s sequenceOrderReferences) ApplyImportTx(ctx context.Context, tx pgx.Tx, _ *reference_data.PreparedReferenceImport) error {
	return s.trace.step(ctx, tx, "references.apply")
}

type sequenceOrderPort struct {
	sourceport.Port
	family string
	trace  *sequenceOrderTrace
}

func (s sequenceOrderPort) Descriptor() sourceport.Descriptor {
	return sourceport.Descriptor{FamilyID: s.family}
}
func (s sequenceOrderPort) ApplyImportTx(ctx context.Context, tx pgx.Tx, _ sourceport.Prepared, input sourceport.ImportContext) error {
	if input.IncidentID != s.trace.incident || input.OperationID != "sequence-order" {
		s.trace.t.Fatal("source application changed prepared identity")
	}
	return s.trace.step(ctx, tx, s.family+".apply")
}
func (s sequenceOrderPort) ValidateImportTx(ctx context.Context, tx pgx.Tx, _ sourceport.Prepared, input sourceport.ImportContext) error {
	if input.IncidentID != s.trace.incident || input.OperationID != "sequence-order" {
		s.trace.t.Fatal("source validation changed prepared identity")
	}
	return s.trace.step(ctx, tx, s.family+".validate")
}

type sequenceOrderTx struct {
	pgx.Tx
	trace *sequenceOrderTrace
}

func (tx *sequenceOrderTx) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	return sequenceOrderRow(func(dest ...any) error {
		if len(dest) != 1 {
			tx.trace.t.Fatal("unexpected coordinator query destination")
		}
		var stage string
		switch strings.TrimSpace(query) {
		case "SELECT public.revisions_incident_bundle_sequence_begin_v1()":
			stage = "sequence.begin"
			if len(args) != 0 {
				tx.trace.t.Fatal("unexpected sequence acquisition arguments")
			}
			*dest[0].(*int64) = 37
		case "SELECT count(*) FROM incidents WHERE id = $1":
			stage = "incident.lookup"
			if len(args) != 1 || args[0] != tx.trace.incident {
				tx.trace.t.Fatal("incident lookup changed identity")
			}
			*dest[0].(*int) = 0
		case "SELECT public.revisions_incident_bundle_sequence_finish_v1($1)":
			stage = "sequence.finish"
			if len(args) != 1 || args[0] != int64(37) {
				tx.trace.t.Fatal("sequence repair lost the acquired original next value")
			}
			*dest[0].(*int64) = 37
		default:
			tx.trace.t.Fatal("unexpected coordinator SQL operation")
		}
		return tx.trace.step(ctx, tx, stage)
	})
}

type sequenceOrderRow func(...any) error

func (r sequenceOrderRow) Scan(dest ...any) error { return r(dest...) }
