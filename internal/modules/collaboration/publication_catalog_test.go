package collaboration

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	collabprotocol "github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
)

func TestPublicationCatalogRejectsIncompleteAndAmbiguousContributions(t *testing.T) {
	validView := ViewPublicationContribution{
		ViewSchemaID: "cartulary.view.notes.v1", PublicFieldKeys: []string{"note.body"}, PatchFieldKeys: []string{"note.body"},
	}
	valid := PublicationContribution{
		ContributionID: "artifacts.publication.v1", SourceOwnerID: "module.artifacts", RecordTypes: []string{"note"}, AffectedViews: []ViewPublicationContribution{validView},
	}
	tests := []struct {
		name          string
		contributions []PublicationContribution
		want          string
	}{
		{name: "empty", want: "catalog is empty"},
		{name: "missing owner", contributions: []PublicationContribution{{ContributionID: valid.ContributionID, RecordTypes: valid.RecordTypes, AffectedViews: valid.AffectedViews}}, want: "is incomplete"},
		{name: "duplicate contribution", contributions: []PublicationContribution{valid, valid}, want: "is duplicated"},
		{name: "duplicate record type", contributions: []PublicationContribution{{ContributionID: valid.ContributionID, SourceOwnerID: valid.SourceOwnerID, RecordTypes: []string{"note", "note"}, AffectedViews: valid.AffectedViews}}, want: "record type"},
		{name: "duplicate view", contributions: []PublicationContribution{
			valid,
			{ContributionID: "other", SourceOwnerID: "other", RecordTypes: []string{"other"}, AffectedViews: []ViewPublicationContribution{validView}},
		}, want: "view \"cartulary.view.notes.v1\" is duplicated"},
		{name: "private patch field", contributions: []PublicationContribution{{
			ContributionID: valid.ContributionID, SourceOwnerID: valid.SourceOwnerID, RecordTypes: valid.RecordTypes,
			AffectedViews: []ViewPublicationContribution{{ViewSchemaID: validView.ViewSchemaID, PublicFieldKeys: validView.PublicFieldKeys, PatchFieldKeys: []string{"note.private"}}},
		}}, want: "is not public"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			catalog, err := NewPublicationCatalog(test.contributions)
			if err == nil || catalog != nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("catalog = %#v, error = %v; want error containing %q", catalog, err, test.want)
			}
		})
	}
}

func TestPublicationCatalogCopiesItsOwnerDeclarations(t *testing.T) {
	recordTypes := []string{"note"}
	publicKeys := []string{"note.body"}
	patchKeys := []string{"note.body"}
	catalog, err := NewPublicationCatalog([]PublicationContribution{{
		ContributionID: "artifacts.publication.v1", SourceOwnerID: "module.artifacts", RecordTypes: recordTypes,
		AffectedViews: []ViewPublicationContribution{{ViewSchemaID: "cartulary.view.notes.v1", PublicFieldKeys: publicKeys, PatchFieldKeys: patchKeys}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	recordTypes[0] = "private"
	publicKeys[0] = "note.private"
	patchKeys[0] = "note.private"
	policy := catalog.views["cartulary.view.notes.v1"]
	if _, ok := policy.recordTypes["note"]; !ok {
		t.Fatalf("catalog record types changed with caller input: %#v", policy.recordTypes)
	}
	if _, ok := policy.publicKeys["note.body"]; !ok {
		t.Fatalf("catalog public keys changed with caller input: %#v", policy.publicKeys)
	}
	if _, ok := policy.patchKeys["note.body"]; !ok {
		t.Fatalf("catalog patch keys changed with caller input: %#v", policy.patchKeys)
	}
}

func TestRecordChangedPublicationRejectsUndeclaredEffects(t *testing.T) {
	catalog, err := NewPublicationCatalog([]PublicationContribution{{
		ContributionID: "artifacts.publication.v1", SourceOwnerID: "module.artifacts", RecordTypes: []string{"note"},
		AffectedViews: []ViewPublicationContribution{{ViewSchemaID: "cartulary.view.notes.v1", PublicFieldKeys: []string{"note.body", "note.title"}, PatchFieldKeys: []string{"note.body"}}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	appender := &publicationAppender{catalog: catalog}
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
