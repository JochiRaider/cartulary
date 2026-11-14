package incidentbundles_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
)

type sourceBindingTx struct{ pgx.Tx }
type sourceBindingAttributions struct{}

func (sourceBindingAttributions) RecordImportedAttribution(string, string, string, string) error {
	return nil
}
func (sourceBindingAttributions) VisitImportedAttributions(string, string, func(incidentportability.ImportedAttribution) error) error {
	return nil
}

func TestSourcePortPreparedBinding_Unit(t *testing.T) {
	actorRows := []sourceport.ActorDescriptor{
		{SourceActorID: uuid.NewString(), DisplayName: "Portable first", EmailHint: "first@example.test", ProviderSubjectHint: "subject-first"},
		{SourceActorID: uuid.NewString(), DisplayName: "Portable second"},
	}
	actors := func(t *testing.T, rows []sourceport.ActorDescriptor) sourceport.ActorCatalog {
		t.Helper()
		got, err := sourceport.NewActorCatalog(rows)
		if err != nil {
			t.Fatal(err)
		}
		return got
	}
	base := sourceport.ImportContext{IncidentID: uuid.New(), ActorUserID: uuid.New(), OperationID: "bound-operation", BundleVersion: 8, Actors: actors(t, actorRows)}
	descriptor := sourceport.Descriptor{OwnerID: "module.binding", FamilyID: "binding", ContractMajor: sourceport.ContractMajor, Paths: []sourceport.Path{{LogicalPath: "data/binding.ndjson", Versions: []int{8, 9}}}}
	for _, phase := range []string{"apply", "validate"} {
		for _, defect := range []string{"none", "operation", "incident", "actor", "version", "actor_missing", "actor_added", "actor_display_name", "actor_email_hint", "actor_provider_subject", "other_instance", "copied_instance", "replaced_instance", "other_family", "zero_prepared"} {
			t.Run(phase+"/"+defect, func(t *testing.T) {
				prepares, callbacks := 0, 0
				payload := &struct{ marker string }{marker: "private owner value"}
				tx := &sourceBindingTx{}
				callback := func(_ context.Context, gotTx pgx.Tx, value any, input sourceport.ImportContext) error {
					callbacks++
					if gotTx != tx || value != payload || input.OperationID != base.OperationID || input.IncidentID != base.IncidentID || input.ActorUserID != base.ActorUserID || input.BundleVersion != base.BundleVersion {
						t.Fatal("bound callback input changed")
					}
					return nil
				}
				options := sourceport.AdapterOptions{Descriptor: descriptor, Prepare: func(context.Context, sourceport.Bundle, sourceport.ImportContext) (any, error) {
					prepares++
					return payload, nil
				}, Apply: callback, Validate: callback}
				port := sourceport.NewAdapter(options)
				prepared, err := port.PrepareImport(t.Context(), portabilitysupport.MapBundle{}, base)
				if err != nil || prepares != 1 {
					t.Fatalf("prepare calls=%d error=%v", prepares, err)
				}
				input := base
				switch defect {
				case "operation":
					input.OperationID = "other-operation"
				case "incident":
					input.IncidentID = uuid.New()
				case "actor":
					input.ActorUserID = uuid.New()
				case "version":
					input.BundleVersion = 9
				case "actor_missing":
					input.Actors = actors(t, actorRows[:1])
				case "actor_added":
					input.Actors = actors(t, append(append([]sourceport.ActorDescriptor(nil), actorRows...), sourceport.ActorDescriptor{SourceActorID: uuid.NewString()}))
				case "actor_display_name", "actor_email_hint", "actor_provider_subject":
					rows := append([]sourceport.ActorDescriptor(nil), actorRows...)
					switch defect {
					case "actor_display_name":
						rows[0].DisplayName = "Changed portable name"
					case "actor_email_hint":
						rows[0].EmailHint = "changed@example.test"
					case "actor_provider_subject":
						rows[0].ProviderSubjectHint = "changed-subject"
					}
					input.Actors = actors(t, rows)
				case "other_instance":
					port = sourceport.NewAdapter(options)
				case "copied_instance":
					copy := *port
					port = &copy
				case "replaced_instance":
					*port = *sourceport.NewAdapter(options)
				case "other_family":
					options.Descriptor.FamilyID = "other"
					port = sourceport.NewAdapter(options)
				case "zero_prepared":
					prepared = sourceport.Prepared{}
				}
				if phase == "apply" {
					err = port.ApplyImportTx(t.Context(), tx, prepared, input)
				} else {
					err = port.ValidateImportTx(t.Context(), tx, prepared, input)
				}
				if defect == "none" {
					if err != nil || callbacks != 1 {
						t.Fatalf("matching binding calls=%d error=%v", callbacks, err)
					}
				} else if !errors.Is(err, sourceport.ErrPreparedBinding) || callbacks != 0 {
					t.Fatalf("mismatched binding calls=%d error=%v", callbacks, err)
				}
			})
		}
	}
	for _, invalid := range []string{"operation", "incident", "actor", "version", "major_zero", "major_old", "major_future"} {
		t.Run("prepare_scope/"+invalid, func(t *testing.T) {
			input, desc := base, descriptor
			switch invalid {
			case "operation":
				input.OperationID = ""
			case "incident":
				input.IncidentID = uuid.Nil
			case "actor":
				input.ActorUserID = uuid.Nil
			case "version":
				input.BundleVersion = 7
			case "major_zero":
				desc.ContractMajor = 0
			case "major_old":
				desc.ContractMajor = 2
			case "major_future":
				desc.ContractMajor = 4
			}
			calls := 0
			port := sourceport.NewAdapter(sourceport.AdapterOptions{Descriptor: desc, Prepare: func(context.Context, sourceport.Bundle, sourceport.ImportContext) (any, error) {
				calls++
				return nil, nil
			}})
			if _, err := port.PrepareImport(t.Context(), portabilitysupport.MapBundle{}, input); !errors.Is(err, sourceport.ErrInvalidCatalog) || calls != 0 {
				t.Fatalf("invalid scope reached owner: calls=%d error=%v", calls, err)
			}
		})
	}
	t.Run("equal_catalog_and_phase_capabilities", func(t *testing.T) {
		calls := 0
		var seenAttributions incidentportability.AttributionRecorder
		var seenRewritten incidentportability.MemberReader
		callback := func(_ context.Context, _ pgx.Tx, _ any, input sourceport.ImportContext) error {
			calls++
			seenAttributions = input.Attributions
			seenRewritten = input.RewrittenObjectBlobs
			return nil
		}
		port := sourceport.NewAdapter(sourceport.AdapterOptions{Descriptor: descriptor, Prepare: func(context.Context, sourceport.Bundle, sourceport.ImportContext) (any, error) { return "private", nil }, Apply: callback, Validate: callback})
		prepared, err := port.PrepareImport(t.Context(), portabilitysupport.MapBundle{}, base)
		if err != nil {
			t.Fatal(err)
		}
		input := base
		input.Actors = actors(t, []sourceport.ActorDescriptor{actorRows[1], actorRows[0]})
		attributions := &sourceBindingAttributions{}
		rewritten := &portabilitysupport.MapBundle{}
		input.Attributions, input.RewrittenObjectBlobs = attributions, rewritten
		if err := port.ApplyImportTx(t.Context(), nil, prepared, input); err != nil {
			t.Fatal(err)
		}
		if err := port.ValidateImportTx(t.Context(), nil, prepared, input); err != nil {
			t.Fatal(err)
		}
		if calls != 2 || seenAttributions != attributions || seenRewritten != rewritten {
			t.Fatal("equal catalog or phase capabilities were changed")
		}
	})
	t.Run("owner_errors_preserve_cause", func(t *testing.T) {
		failure := errors.New("owner capability failure")
		failPrepare := true
		callback := func(context.Context, pgx.Tx, any, sourceport.ImportContext) error { return failure }
		port := sourceport.NewAdapter(sourceport.AdapterOptions{Descriptor: descriptor, Prepare: func(context.Context, sourceport.Bundle, sourceport.ImportContext) (any, error) {
			if failPrepare {
				return nil, failure
			}
			return nil, nil
		}, Apply: callback, Validate: callback})
		if _, err := port.PrepareImport(t.Context(), portabilitysupport.MapBundle{}, base); !errors.Is(err, failure) {
			t.Fatalf("prepare cause lost: %v", err)
		}
		failPrepare = false
		prepared, err := port.PrepareImport(t.Context(), portabilitysupport.MapBundle{}, base)
		if err != nil {
			t.Fatal(err)
		}
		if err := port.ApplyImportTx(t.Context(), nil, prepared, base); !errors.Is(err, failure) {
			t.Fatalf("apply cause lost: %v", err)
		}
		if err := port.ValidateImportTx(t.Context(), nil, prepared, base); !errors.Is(err, failure) {
			t.Fatalf("validate cause lost: %v", err)
		}
	})
}
