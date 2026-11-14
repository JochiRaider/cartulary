package revisions_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
)

type revisionsPreparationEnvelopeProbe struct{ calls int }

type revisionsPreparationFailingReader struct {
	portabilitysupport.MapBundle
	path string
	err  error
}

func (r revisionsPreparationFailingReader) OpenMember(ctx context.Context, path string) (io.ReadCloser, incidentportability.Member, error) {
	if path == r.path {
		return nil, incidentportability.Member{}, r.err
	}
	return r.MapBundle.OpenMember(ctx, path)
}

func (p *revisionsPreparationEnvelopeProbe) VisitIncidentRecordIDsTx(context.Context, pgx.Tx, uuid.UUID, func(uuid.UUID) error) error {
	p.calls++
	return errors.New("preparation must not enumerate database envelopes")
}

func (p *revisionsPreparationEnvelopeProbe) RecordTypeTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) (string, error) {
	p.calls++
	return "", errors.New("preparation must not load database envelopes")
}

func newRevisionsPreparationHarness(t testing.TB) (revisionsPortabilityHarness, *revisionsPreparationEnvelopeProbe) {
	t.Helper()
	contributions := mustRevisionProviderContributions(t)
	snapshots, err := revisions.NewRecordSnapshotCaptureCatalog(contributions, records.ValidateSnapshotEnvelope)
	if err != nil {
		t.Fatal(err)
	}
	targets, err := revisions.NewTargetSemanticsCatalog(contributions)
	if err != nil {
		t.Fatal(err)
	}
	probe := &revisionsPreparationEnvelopeProbe{}
	port, err := revisions.NewIncidentBundleSourcePort(probe, snapshots, targets)
	if err != nil {
		t.Fatal(err)
	}
	h := revisionsPortabilityHarness{
		actor: authn.UserRecord{ID: uuid.New()}, incidentID: uuid.New(), recordID: uuid.New(),
		changeSet: uuid.MustParse("00000000-0000-4000-8000-000000000001"),
		createdAt: time.Date(2026, time.August, 1, 12, 0, 0, 123456000, time.UTC),
	}
	actors, err := sourceport.NewActorCatalog([]sourceport.ActorDescriptor{{SourceActorID: h.actor.ID.String()}})
	if err != nil {
		t.Fatal(err)
	}
	input := sourceport.ImportContext{IncidentID: h.incidentID, ActorUserID: h.actor.ID, Actors: actors, BundleVersion: 8, OperationID: uuid.NewString()}
	h.port, h.context = port, input
	return h, probe
}

func TestRevisionsIncidentBundlePreparationPrecedence_Unit(t *testing.T) {
	h, probe := newRevisionsPreparationHarness(t)
	exerciseRevisionsSequenceScalarAdmission(t, h, probe)
	exerciseRevisionActorContextAdmission(t, h, probe)
	exerciseRevisionContextExportCapacity(t, h, probe)
	exerciseHistoryOrderAdmission(t, h, probe)
	port, input := h.port, h.context
	check := func(t *testing.T, bundle portabilitysupport.MapBundle, invariant string) {
		t.Helper()
		before := clonePortableBundle(bundle)
		_, err := port.PrepareImport(t.Context(), bundle, input)
		if invariant == "" {
			if err != nil {
				t.Fatalf("valid preparation rejected: %v", err)
			}
		} else {
			requireRevisionsInvariant(t, err, invariant)
		}
		if !reflect.DeepEqual(bundle, before) {
			t.Fatal("preparation changed member bytes")
		}
		if probe.calls != 0 {
			t.Fatalf("preparation made %d envelope calls", probe.calls)
		}
	}
	row := func(path string, index int) map[string]any {
		r := decodePortableRows(t, h.validBundle(t)[path])[0]
		switch path {
		case revisionsTestChangeSetsPath:
			r["change_set_id"] = fmt.Sprintf("00000000-0000-4000-8000-%012d", index+1)
		case revisionsTestMutationsPath:
			r["sequence_no"] = index + 1
		case revisionsTestHistoryPath:
			r["revision_id"], r["row_version"] = index+1, index+1
		}
		return r
	}
	type defect struct {
		name, invariant string
		rank            int
		apply           func(map[string]any)
	}
	reference := defect{"reference", "revisions.references_complete", 1, func(r map[string]any) { r["change_set_id"] = "invalid-uuid" }}
	clock := defect{"clock", "revisions.history_reconstruction", 5, func(r map[string]any) { r["created_at"] = "invalid-clock" }}
	unknownActor := "ffffffff-ffff-4fff-8fff-ffffffffffff"
	families := []struct {
		path    string
		defects []defect
	}{
		{revisionsTestChangeSetsPath, []defect{reference, {"actor", "revisions.actor_references_complete", 2, func(r map[string]any) { r["actor_user_id"] = unknownActor }}, clock}},
		{revisionsTestMutationsPath, []defect{reference, {"actor", "revisions.actor_references_complete", 2, func(r map[string]any) { r["after_value"] = map[string]any{"owner_user_id": unknownActor} }}, {"sequence", "revisions.mutation_sequence_contiguous", 3, func(r map[string]any) { r["sequence_no"] = 0 }}}},
		{revisionsTestHistoryPath, []defect{reference, {"actor", "revisions.actor_references_complete", 2, func(r map[string]any) { r["after_json"] = map[string]any{"owner_user_id": unknownActor} }}, {"version", "revisions.record_version_unique", 4, func(r map[string]any) { r["revision_id"] = 0 }}, clock}},
	}
	// Null snapshots are legal preparation input. The integration row separately
	// qualifies full snapshots, provider binding and transaction validation.
	check(t, h.validBundle(t), "")
	for _, first := range families {
		for _, second := range families {
			for _, left := range first.defects {
				for _, right := range second.defects {
					if left.rank == right.rank {
						continue
					}
					t.Run(fmt.Sprintf("%s_%s_then_%s_%s", first.path, left.name, second.path, right.name), func(t *testing.T) {
						bundle := h.validBundle(t)
						a, b := row(first.path, 0), row(second.path, 1)
						left.apply(a)
						right.apply(b)
						if first.path == second.path {
							bundle[first.path] = encodePortableRows(t, []map[string]any{a, b})
						} else {
							bundle[first.path] = encodePortableRows(t, []map[string]any{a})
							bundle[second.path] = encodePortableRows(t, []map[string]any{b})
						}
						want := left.invariant
						if right.rank < left.rank {
							want = right.invariant
						}
						check(t, bundle, want)
					})
				}
			}
		}
		for _, high := range first.defects {
			for _, low := range first.defects {
				if high.rank >= low.rank {
					continue
				}
				t.Run(first.path+"_same_row_"+high.name+"_"+low.name, func(t *testing.T) {
					bundle := h.validBundle(t)
					r := row(first.path, 0)
					high.apply(r)
					low.apply(r)
					bundle[first.path] = encodePortableRows(t, []map[string]any{r})
					check(t, bundle, high.invariant)
				})
			}
		}
		t.Run(first.path+"_late_malformed", func(t *testing.T) {
			bundle := h.validBundle(t)
			r := row(revisionsTestChangeSetsPath, 0)
			r["actor_user_id"] = unknownActor
			bundle[revisionsTestChangeSetsPath] = encodePortableRows(t, []map[string]any{r})
			bundle[first.path] = append(bundle[first.path], []byte("{broken\n")...)
			check(t, bundle, "revisions.references_complete")
		})
	}
	for _, duplicate := range []bool{false, true} {
		t.Run(fmt.Sprintf("identity_after_rejected_row_duplicate=%t", duplicate), func(t *testing.T) {
			bundle := h.validBundle(t)
			a, b := row(revisionsTestChangeSetsPath, 1), row(revisionsTestChangeSetsPath, 0)
			a["actor_user_id"] = unknownActor
			if duplicate {
				b["change_set_id"] = a["change_set_id"]
			}
			bundle[revisionsTestChangeSetsPath] = encodePortableRows(t, []map[string]any{a, b})
			check(t, bundle, "revisions.references_complete")
		})
	}
	t.Run("preparation_precedes_database_reference", func(t *testing.T) {
		bundle := h.validBundle(t)
		changes := row(revisionsTestChangeSetsPath, 0)
		changes["created_at"] = "invalid-clock"
		history := row(revisionsTestHistoryPath, 0)
		history["record_id"] = uuid.NewString()
		bundle[revisionsTestChangeSetsPath] = encodePortableRows(t, []map[string]any{changes})
		bundle[revisionsTestHistoryPath] = encodePortableRows(t, []map[string]any{history})
		check(t, bundle, "revisions.history_reconstruction")
	})
	for _, test := range []struct {
		path, field string
		value       any
	}{
		{revisionsTestChangeSetsPath, "incident_id", unknownActor},
		{revisionsTestChangeSetsPath, "source", ""},
		{revisionsTestChangeSetsPath, "reason", 7},
		{revisionsTestChangeSetsPath, "client_txn_id", 7},
		{revisionsTestChangeSetsPath, "request_id", 7},
		{revisionsTestMutationsPath, "target_kind", ""},
		{revisionsTestMutationsPath, "target_id", ""},
		{revisionsTestMutationsPath, "operation_kind", ""},
		{revisionsTestMutationsPath, "before_version_id", ""},
		{revisionsTestMutationsPath, "after_version_id", ""},
		{revisionsTestHistoryPath, "record_id", "invalid-uuid"},
	} {
		t.Run(test.path+"_reference_scalar_"+test.field, func(t *testing.T) {
			bundle := h.validBundle(t)
			r := row(test.path, 0)
			switch test.path {
			case revisionsTestChangeSetsPath:
				r["actor_user_id"] = "invalid-actor"
			case revisionsTestMutationsPath:
				r["sequence_no"] = 0
			case revisionsTestHistoryPath:
				r["revision_id"] = 0
			}
			r[test.field] = test.value
			bundle[test.path] = encodePortableRows(t, []map[string]any{r})
			check(t, bundle, "revisions.references_complete")
		})
	}
	t.Run("later_reader_failure_propagates", func(t *testing.T) {
		bundle := h.validBundle(t)
		changes := row(revisionsTestChangeSetsPath, 0)
		changes["actor_user_id"] = unknownActor
		bundle[revisionsTestChangeSetsPath] = encodePortableRows(t, []map[string]any{changes})
		failure := errors.New("member reader failed")
		_, err := port.PrepareImport(t.Context(), revisionsPreparationFailingReader{MapBundle: bundle, path: revisionsTestHistoryPath, err: failure}, input)
		if !errors.Is(err, failure) {
			t.Fatalf("reader error = %v, want original failure", err)
		}
		if probe.calls != 0 {
			t.Fatalf("preparation made %d envelope calls", probe.calls)
		}
	})
}

func exerciseRevisionsSequenceScalarAdmission(t *testing.T, h revisionsPortabilityHarness, probe *revisionsPreparationEnvelopeProbe) {
	t.Helper()
	for _, test := range []struct {
		value    string
		admitted bool
	}{
		{"1", true}, {"2", true}, {"2147483646", true}, {"2147483647", true},
		{"2147483648", false}, {"9007199254740993", false},
		{"9223372036854775807", false}, {"9223372036854775808", false},
	} {
		for _, higher := range []string{"none", "reference", "actor"} {
			t.Run("sequence_scalar_capacity/"+test.value+"/"+higher, func(t *testing.T) {
				bundle := h.validBundle(t)
				rows := decodePortableRows(t, bundle[revisionsTestMutationsPath])
				rows[0]["sequence_no"] = json.Number(test.value)
				want := ""
				if !test.admitted {
					want = "revisions.mutation_sequence_contiguous"
				}
				switch higher {
				case "reference":
					rows[0]["target_id"] = ""
					want = "revisions.references_complete"
				case "actor":
					rows[0]["after_value"] = map[string]any{"owner_user_id": uuid.NewString()}
					want = "revisions.actor_references_complete"
				}
				bundle[revisionsTestMutationsPath] = encodePortableRows(t, rows)
				before := clonePortableBundle(bundle)
				_, err := h.port.PrepareImport(t.Context(), bundle, h.context)
				if want == "" {
					if err != nil {
						t.Fatalf("in-range scalar rejected during preparation: %v", err)
					}
				} else {
					requireRevisionsInvariant(t, err, want)
				}
				if !reflect.DeepEqual(bundle, before) {
					t.Fatal("sequence admission changed member bytes")
				}
				if probe.calls != 0 {
					t.Fatalf("sequence preparation made %d envelope calls", probe.calls)
				}
			})
		}
	}
}
