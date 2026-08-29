package collaboration

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	collabprotocol "github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
)

func TestPublicationCatalogRejectsIncompleteAndAmbiguousContributions(t *testing.T) {
	validView := ViewPublicationContribution{
		ViewSchemaID: "cartulary.view.notes.v1", RecordTypes: []string{"artifact"},
		PublicFieldKeys: []string{"note.body"}, PatchFieldKeys: []string{"note.body"},
	}
	valid := PublicationContribution{
		ContributionID: "artifacts.publication.v1", SourceOwnerID: "artifacts", AffectedViews: []ViewPublicationContribution{validView},
	}
	canonical := []CanonicalPublicationView{{
		ViewSchemaID: validView.ViewSchemaID, SourceOwnerID: valid.SourceOwnerID, RecordTypes: []string{"artifact"},
	}}
	tests := []struct {
		name          string
		contributions []PublicationContribution
		canonical     []CanonicalPublicationView
		want          string
	}{
		{name: "empty contribution set", canonical: canonical, want: "catalog is empty"},
		{name: "empty canonical set", contributions: []PublicationContribution{valid}, want: "canonical publication views are empty"},
		{name: "missing owner", contributions: []PublicationContribution{{ContributionID: valid.ContributionID, AffectedViews: valid.AffectedViews}}, canonical: canonical, want: "is incomplete"},
		{name: "duplicate contribution", contributions: []PublicationContribution{valid, valid}, canonical: canonical, want: "contribution \"artifacts.publication.v1\" is duplicated"},
		{name: "duplicate owner", contributions: []PublicationContribution{
			valid,
			{ContributionID: "other", SourceOwnerID: valid.SourceOwnerID, AffectedViews: []ViewPublicationContribution{{ViewSchemaID: "cartulary.view.other.v1", RecordTypes: []string{"other"}, PublicFieldKeys: []string{"other.value"}, PatchFieldKeys: []string{"other.value"}}}},
		}, canonical: append(append([]CanonicalPublicationView(nil), canonical...), CanonicalPublicationView{ViewSchemaID: "cartulary.view.other.v1", SourceOwnerID: valid.SourceOwnerID, RecordTypes: []string{"other"}}), want: "source owner \"artifacts\" is duplicated"},
		{name: "duplicate record type", contributions: []PublicationContribution{{
			ContributionID: valid.ContributionID, SourceOwnerID: valid.SourceOwnerID,
			AffectedViews: []ViewPublicationContribution{{ViewSchemaID: validView.ViewSchemaID, RecordTypes: []string{"artifact", "artifact"}, PublicFieldKeys: validView.PublicFieldKeys, PatchFieldKeys: validView.PatchFieldKeys}},
		}}, canonical: canonical, want: "record type"},
		{name: "duplicate view", contributions: []PublicationContribution{
			valid,
			{ContributionID: "other", SourceOwnerID: "other", AffectedViews: []ViewPublicationContribution{validView}},
		}, canonical: canonical, want: "view \"cartulary.view.notes.v1\" is duplicated"},
		{name: "unknown view", contributions: []PublicationContribution{{
			ContributionID: valid.ContributionID, SourceOwnerID: valid.SourceOwnerID,
			AffectedViews: []ViewPublicationContribution{{ViewSchemaID: "cartulary.view.unknown.v1", RecordTypes: validView.RecordTypes, PublicFieldKeys: validView.PublicFieldKeys, PatchFieldKeys: validView.PatchFieldKeys}},
		}}, canonical: canonical, want: "view \"cartulary.view.unknown.v1\" is unknown"},
		{name: "cross owner", contributions: []PublicationContribution{{
			ContributionID: valid.ContributionID, SourceOwnerID: "parties", AffectedViews: valid.AffectedViews,
		}}, canonical: canonical, want: "belongs to source owner \"artifacts\", not \"parties\""},
		{name: "record view mismatch", contributions: []PublicationContribution{{
			ContributionID: valid.ContributionID, SourceOwnerID: valid.SourceOwnerID,
			AffectedViews: []ViewPublicationContribution{{ViewSchemaID: validView.ViewSchemaID, RecordTypes: []string{"note"}, PublicFieldKeys: validView.PublicFieldKeys, PatchFieldKeys: validView.PatchFieldKeys}},
		}}, canonical: canonical, want: "record types do not match"},
		{name: "missing canonical relationship", contributions: []PublicationContribution{valid}, canonical: append(append([]CanonicalPublicationView(nil), canonical...), CanonicalPublicationView{
			ViewSchemaID: "cartulary.view.comm_log.v1", SourceOwnerID: valid.SourceOwnerID, RecordTypes: []string{"artifact"},
		}), want: "view \"cartulary.view.comm_log.v1\" is missing"},
		{name: "duplicate canonical view", contributions: []PublicationContribution{valid}, canonical: append(append([]CanonicalPublicationView(nil), canonical...), canonical[0]), want: "canonical publication view \"cartulary.view.notes.v1\" is duplicated"},
		{name: "incomplete canonical view", contributions: []PublicationContribution{valid}, canonical: []CanonicalPublicationView{{ViewSchemaID: validView.ViewSchemaID, RecordTypes: validView.RecordTypes}}, want: "is incomplete"},
		{name: "duplicate canonical record type", contributions: []PublicationContribution{valid}, canonical: []CanonicalPublicationView{{ViewSchemaID: validView.ViewSchemaID, SourceOwnerID: valid.SourceOwnerID, RecordTypes: []string{"artifact", "artifact"}}}, want: "record type"},
		{name: "private patch field", contributions: []PublicationContribution{{
			ContributionID: valid.ContributionID, SourceOwnerID: valid.SourceOwnerID,
			AffectedViews: []ViewPublicationContribution{{ViewSchemaID: validView.ViewSchemaID, RecordTypes: validView.RecordTypes, PublicFieldKeys: validView.PublicFieldKeys, PatchFieldKeys: []string{"note.private"}}},
		}}, canonical: canonical, want: "is not public"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			catalog, err := NewPublicationCatalog(test.contributions, test.canonical)
			if err == nil || catalog != nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("catalog = %#v, error = %v; want error containing %q", catalog, err, test.want)
			}
		})
	}
}

func TestPublicationCatalogCopiesItsOwnerDeclarations(t *testing.T) {
	recordTypes := []string{"artifact"}
	canonicalRecordTypes := []string{"artifact"}
	publicKeys := []string{"note.body"}
	patchKeys := []string{"note.body"}
	catalog, err := NewPublicationCatalog([]PublicationContribution{{
		ContributionID: "artifacts.publication.v1", SourceOwnerID: "artifacts",
		AffectedViews: []ViewPublicationContribution{{ViewSchemaID: "cartulary.view.notes.v1", RecordTypes: recordTypes, PublicFieldKeys: publicKeys, PatchFieldKeys: patchKeys}},
	}}, []CanonicalPublicationView{{ViewSchemaID: "cartulary.view.notes.v1", SourceOwnerID: "artifacts", RecordTypes: canonicalRecordTypes}})
	if err != nil {
		t.Fatal(err)
	}
	recordTypes[0] = "private"
	canonicalRecordTypes[0] = "private"
	publicKeys[0] = "note.private"
	patchKeys[0] = "note.private"
	policy := catalog.views["cartulary.view.notes.v1"]
	if _, ok := policy.publicKeys["note.body"]; !ok {
		t.Fatalf("catalog public keys changed with caller input: %#v", policy.publicKeys)
	}
	if _, ok := policy.patchKeys["note.body"]; !ok {
		t.Fatalf("catalog patch keys changed with caller input: %#v", policy.patchKeys)
	}
}

func TestPublicationAppendersCannotBeWidenedAcrossFamilies(t *testing.T) {
	catalog, err := NewPublicationCatalog([]PublicationContribution{{
		ContributionID: "artifacts.publication.v1", SourceOwnerID: "artifacts",
		AffectedViews: []ViewPublicationContribution{{
			ViewSchemaID: "cartulary.view.notes.v1", RecordTypes: []string{"artifact"},
			PublicFieldKeys: []string{"note.body"}, PatchFieldKeys: []string{"note.body"},
		}},
	}}, []CanonicalPublicationView{{
		ViewSchemaID: "cartulary.view.notes.v1", SourceOwnerID: "artifacts", RecordTypes: []string{"artifact"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	recordChanges, err := NewRecordChangedAppender(catalog)
	if err != nil {
		t.Fatal(err)
	}
	jobProgress := NewJobProgressAppender()
	extensionChanges := NewExtensionResourceChangedAppender()

	if _, widened := recordChanges.(JobProgressAppender); widened {
		t.Fatal("record-change capability widened to job progress")
	}
	if _, widened := recordChanges.(ExtensionResourceChangedAppender); widened {
		t.Fatal("record-change capability widened to extension resource changes")
	}
	if _, widened := jobProgress.(RecordChangedAppender); widened {
		t.Fatal("job-progress capability widened to record changes")
	}
	if _, widened := jobProgress.(ExtensionResourceChangedAppender); widened {
		t.Fatal("job-progress capability widened to extension resource changes")
	}
	if _, widened := extensionChanges.(RecordChangedAppender); widened {
		t.Fatal("extension-resource-change capability widened to record changes")
	}
	if _, widened := extensionChanges.(JobProgressAppender); widened {
		t.Fatal("extension-resource-change capability widened to job progress")
	}
}

func TestRecordChangedPublicationRejectsUndeclaredEffects(t *testing.T) {
	catalog, err := NewPublicationCatalog([]PublicationContribution{{
		ContributionID: "artifacts.publication.v1", SourceOwnerID: "artifacts",
		AffectedViews: []ViewPublicationContribution{{ViewSchemaID: "cartulary.view.notes.v1", RecordTypes: []string{"artifact"}, PublicFieldKeys: []string{"note.body", "note.title"}, PatchFieldKeys: []string{"note.body"}}},
	}}, []CanonicalPublicationView{{ViewSchemaID: "cartulary.view.notes.v1", SourceOwnerID: "artifacts", RecordTypes: []string{"artifact"}}})
	if err != nil {
		t.Fatal(err)
	}
	appender := &recordChangedAppender{catalog: catalog}
	recordID := uuid.New()
	base := RecordChangeIntentInput{
		IncidentID: uuid.New(), RecordID: recordID, ChangeSetID: uuid.New(), ActorUserID: uuid.New(), RowVersion: 2,
		ClientTxnID: "txn", PublicFieldKeys: []string{"note.body"},
		AffectedViews: []AffectedViewChange{{ViewSchemaID: "cartulary.view.notes.v1", RecordID: recordID, RowVersion: 2, ChangeKind: "invalidate"}},
	}
	validPatch := collabprotocol.BuildViewRowPatch(map[string]any{
		"record_id": recordID.String(), "row_version": int64(2),
		"cells": map[string]any{"note.body": map[string]any{"value": "body"}},
	}, []string{"note.body"})
	tests := []struct {
		name string
		edit func(*RecordChangeIntentInput)
		want string
	}{
		{name: "unknown view", edit: func(input *RecordChangeIntentInput) {
			input.AffectedViews[0].ViewSchemaID = "cartulary.view.private.v1"
		}, want: "affected view"},
		{name: "record mismatch", edit: func(input *RecordChangeIntentInput) { input.AffectedViews[0].RecordID = uuid.New() }, want: "affected view"},
		{name: "version mismatch", edit: func(input *RecordChangeIntentInput) { input.AffectedViews[0].RowVersion++ }, want: "affected view"},
		{name: "invalid change kind", edit: func(input *RecordChangeIntentInput) { input.AffectedViews[0].ChangeKind = "replace" }, want: "affected view"},
		{name: "patch required", edit: func(input *RecordChangeIntentInput) { input.AffectedViews[0].ChangeKind = "patch" }, want: "affected view"},
		{name: "patch forbidden", edit: func(input *RecordChangeIntentInput) { input.AffectedViews[0].PatchCells = validPatch }, want: "affected view"},
		{name: "private changed field", edit: func(input *RecordChangeIntentInput) { input.PublicFieldKeys = []string{"note.private"} }, want: "is not public"},
		{name: "private patch field", edit: func(input *RecordChangeIntentInput) {
			input.PublicFieldKeys = []string{"note.title"}
			input.AffectedViews[0].ChangeKind = "patch"
			input.AffectedViews[0].PatchCells = collabprotocol.BuildViewRowPatch(map[string]any{"record_id": recordID.String(), "row_version": int64(2), "cells": map[string]any{"note.title": map[string]any{"value": "title"}}}, []string{"note.title"})
		}, want: "patch field \"note.title\" is not admitted"},
		{name: "patch field not declared changed", edit: func(input *RecordChangeIntentInput) {
			input.PublicFieldKeys = []string{"note.title"}
			input.AffectedViews[0].ChangeKind = "patch"
			input.AffectedViews[0].PatchCells = validPatch
		}, want: "patch field \"note.body\" was not declared changed"},
		{name: "undeclared patch field", edit: func(input *RecordChangeIntentInput) {
			input.AffectedViews[0].ChangeKind = "patch"
			input.AffectedViews[0].PatchCells = collabprotocol.BuildViewRowPatch(map[string]any{"record_id": recordID.String(), "row_version": int64(2), "cells": map[string]any{"note.body": map[string]any{"value": "body"}, "note.title": map[string]any{"value": "title"}}}, []string{"note.body", "note.title"})
		}, want: "patch field \"note.title\" is not admitted"},
		{name: "unknown patch member", edit: func(input *RecordChangeIntentInput) {
			input.AffectedViews[0].ChangeKind = "patch"
			input.AffectedViews[0].PatchCells = map[string]any{"record_id": recordID.String(), "row_version": int64(2), "cells": map[string]any{"note.body": map[string]any{"value": "body"}}, "private": true}
		}, want: "patch member \"private\" is not admitted"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := base
			input.PublicFieldKeys = append([]string(nil), base.PublicFieldKeys...)
			input.AffectedViews = append([]AffectedViewChange(nil), base.AffectedViews...)
			test.edit(&input)
			if _, err := appender.recordChangedIntent(input); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v; want error containing %q", err, test.want)
			}
		})
	}
}
