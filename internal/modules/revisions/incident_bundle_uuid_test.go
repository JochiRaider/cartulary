package revisions_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
)

func TestRevisionsIncidentBundleUUIDAdmission_Unit(t *testing.T) {
	h, probe := newRevisionsPreparationHarness(t)
	canonical := uuid.MustParse("abcdef01-2345-4678-9abc-def012345678").String()
	for _, field := range []struct {
		path, name string
		actor      bool
		snapshot   bool
	}{
		{revisionsTestChangeSetsPath, "change_set_id", false, false},
		{revisionsTestChangeSetsPath, "incident_id", false, false},
		{revisionsTestChangeSetsPath, "actor_user_id", true, false},
		{revisionsTestMutationsPath, "change_set_id", false, false},
		{revisionsTestHistoryPath, "change_set_id", false, false},
		{revisionsTestHistoryPath, "record_id", false, false},
		{revisionsTestMutationsPath, "before_value", true, true},
		{revisionsTestMutationsPath, "after_value", true, true},
		{revisionsTestHistoryPath, "before_json", true, true},
		{revisionsTestHistoryPath, "after_json", true, true},
	} {
		for index, value := range []any{
			canonical, nil, uuid.Nil.String(), "", "bad-uuid", true, 1,
			strings.ToUpper(canonical), " " + canonical + " ",
			strings.ReplaceAll(canonical, "-", ""), "{" + canonical + "}", "urn:uuid:" + canonical,
		} {
			t.Run(fmt.Sprintf("%s_%s_case_%d", field.path, field.name, index), func(t *testing.T) {
				input := h.context
				if field.actor {
					descriptors := []sourceport.ActorDescriptor{{SourceActorID: h.actor.ID.String()}}
					if text, ok := value.(string); ok {
						if _, err := uuid.Parse(strings.TrimSpace(text)); err == nil {
							// Use the real catalog, including its currently permissive
							// descriptor forms: membership cannot admit invalid actors.
							descriptors = append(descriptors, sourceport.ActorDescriptor{SourceActorID: text})
						}
					}
					var err error
					input.Actors, err = sourceport.NewActorCatalog(descriptors)
					if err != nil {
						t.Fatalf("build actor admission fixture: %v", err)
					}
				}
				if field.name == "incident_id" {
					if value == canonical {
						input.IncidentID = uuid.MustParse(canonical)
					} else if value == uuid.Nil.String() {
						input.IncidentID = uuid.Nil
					}
				}
				bundle := h.validBundle(t)
				rows := decodePortableRows(t, bundle[field.path])
				if field.snapshot {
					rows[0][field.name] = map[string]any{"nested": []any{map[string]any{"owner_user_id": value}}}
				} else {
					rows[0][field.name] = value
				}
				bundle[field.path] = encodePortableRows(t, rows)
				before := clonePortableBundle(bundle)
				_, err := h.port.PrepareImport(t.Context(), bundle, input)
				if value == canonical || (field.snapshot && value == nil) {
					if err != nil {
						t.Fatalf("valid UUID preparation rejected: %v", err)
					}
				} else {
					invariant := "revisions.references_complete"
					if field.actor {
						invariant = "revisions.actor_references_complete"
					}
					requireRevisionsInvariant(t, err, invariant)
				}
				if !reflect.DeepEqual(bundle, before) {
					t.Fatal("preparation changed UUID source bytes")
				}
				if probe.calls != 0 {
					t.Fatalf("preparation made %d envelope calls", probe.calls)
				}
			})
		}
		if field.actor {
			t.Run(field.path+"_"+field.name+"_unlisted_actor", func(t *testing.T) {
				bundle := h.validBundle(t)
				rows := decodePortableRows(t, bundle[field.path])
				if field.snapshot {
					rows[0][field.name] = map[string]any{"owner_user_id": canonical}
				} else {
					rows[0][field.name] = canonical
				}
				bundle[field.path] = encodePortableRows(t, rows)
				_, err := h.port.PrepareImport(t.Context(), bundle, h.context)
				requireRevisionsInvariant(t, err, "revisions.actor_references_complete")
			})
		}
		if !field.snapshot {
			t.Run(field.path+"_"+field.name+"_missing_member", func(t *testing.T) {
				bundle := h.validBundle(t)
				rows := decodePortableRows(t, bundle[field.path])
				delete(rows[0], field.name)
				bundle[field.path] = encodePortableRows(t, rows)
				_, err := h.port.PrepareImport(t.Context(), bundle, h.context)
				requireRevisionsInvariant(t, err, "revisions.references_complete")
			})
		}
	}
}
