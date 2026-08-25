package protocol

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

// RecordChangedEvent is the decoded semantic form of a sequenced public
// record_changed message. It is independent of Collaboration publication
// inputs and source-owner revision facts.
type RecordChangedEvent struct {
	IncidentID       uuid.UUID
	RecordID         uuid.UUID
	RowVersion       int64
	ChangeSetID      uuid.UUID
	ClientTxnID      string
	ActorUserID      uuid.UUID
	ChangedFieldKeys []string
	AffectedViews    []RecordChangedView
	StreamSeq        int64
	EventID          uuid.UUID
	EmittedAt        time.Time
}

type RecordChangedView struct {
	ViewSchemaID string
	ChangeKind   string
	PatchCells   map[string]any
}

func RecordChangeFromSequencedMessage(message Message) (RecordChangedEvent, error) {
	var payload struct {
		RecordID         string           `json:"record_id"`
		RowVersion       int64            `json:"row_version"`
		ChangeSetID      string           `json:"change_set_id"`
		ClientTxnID      string           `json:"client_txn_id"`
		ActorUserID      string           `json:"actor_user_id"`
		ChangedFieldKeys []string         `json:"changed_field_keys"`
		AffectedViews    []map[string]any `json:"affected_views"`
	}
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		return RecordChangedEvent{}, fmt.Errorf("decode sequenced record_changed payload: %w", err)
	}
	incidentID, incidentErr := uuid.Parse(message.IncidentID)
	recordID, recordErr := uuid.Parse(payload.RecordID)
	changeSetID, changeSetErr := uuid.Parse(payload.ChangeSetID)
	actorUserID, actorErr := uuid.Parse(payload.ActorUserID)
	eventID, eventErr := uuid.Parse(message.EventID)
	emittedAt, emittedErr := time.Parse(time.RFC3339Nano, message.EmittedAt)
	if incidentErr != nil || recordErr != nil || changeSetErr != nil || actorErr != nil || eventErr != nil ||
		emittedErr != nil || message.StreamSeq == nil || *message.StreamSeq < 1 || payload.RowVersion < 1 || len(payload.AffectedViews) == 0 {
		return RecordChangedEvent{}, errors.New("invalid sequenced record_changed identity")
	}
	affectedViews := make([]RecordChangedView, 0, len(payload.AffectedViews))
	for _, rawView := range payload.AffectedViews {
		viewSchemaID, _ := rawView["view_schema_id"].(string)
		changeKind, _ := rawView["change_kind"].(string)
		var patchCells map[string]any
		if value, ok := rawView["patch_cells"].(map[string]any); ok {
			patchCells = value
		}
		if viewSchemaID == "" || !validRecordChangeKind(changeKind) || (changeKind == "patch") != (patchCells != nil) {
			return RecordChangedEvent{}, errors.New("invalid sequenced record_changed affected view")
		}
		affectedViews = append(affectedViews, RecordChangedView{ViewSchemaID: viewSchemaID, ChangeKind: changeKind, PatchCells: patchCells})
	}
	if !slices.IsSortedFunc(affectedViews, func(left RecordChangedView, right RecordChangedView) int {
		return strings.Compare(left.ViewSchemaID, right.ViewSchemaID)
	}) {
		return RecordChangedEvent{}, errors.New("invalid sequenced record_changed affected view order")
	}
	for index := 1; index < len(affectedViews); index++ {
		if affectedViews[index-1].ViewSchemaID == affectedViews[index].ViewSchemaID {
			return RecordChangedEvent{}, errors.New("invalid sequenced record_changed duplicate affected view")
		}
	}
	return RecordChangedEvent{
		IncidentID: incidentID, RecordID: recordID, RowVersion: payload.RowVersion, ChangeSetID: changeSetID,
		ClientTxnID: payload.ClientTxnID, ActorUserID: actorUserID, ChangedFieldKeys: payload.ChangedFieldKeys,
		AffectedViews: affectedViews, StreamSeq: *message.StreamSeq, EventID: eventID, EmittedAt: emittedAt,
	}, nil
}

func RecordChangePayload(change RecordChangedEvent) map[string]any {
	changedKeys := append([]string(nil), change.ChangedFieldKeys...)
	slices.Sort(changedKeys)
	changedKeys = slices.Compact(changedKeys)
	views := append([]RecordChangedView(nil), change.AffectedViews...)
	slices.SortFunc(views, func(left RecordChangedView, right RecordChangedView) int {
		return strings.Compare(left.ViewSchemaID, right.ViewSchemaID)
	})
	affectedViews := make([]map[string]any, 0, len(views))
	for _, affected := range views {
		view := map[string]any{"view_schema_id": affected.ViewSchemaID, "change_kind": affected.ChangeKind}
		if affected.PatchCells != nil {
			view["change_kind"] = "patch"
			view["patch_cells"] = affected.PatchCells
		}
		affectedViews = append(affectedViews, view)
	}
	return map[string]any{
		"record_id": change.RecordID.String(), "row_version": change.RowVersion,
		"change_set_id": change.ChangeSetID.String(), "client_txn_id": change.ClientTxnID,
		"actor_user_id": change.ActorUserID.String(), "changed_field_keys": changedKeys,
		"affected_views": affectedViews,
	}
}

func BuildViewRowPatch(row map[string]any, changedFieldKeys []string) map[string]any {
	if row == nil {
		return nil
	}
	recordID, recordOK := row["record_id"]
	rowVersion, versionOK := row["row_version"]
	cells, cellsOK := row["cells"].(map[string]any)
	if !recordOK || !versionOK || !cellsOK {
		return nil
	}
	changed := make(map[string]struct{}, len(changedFieldKeys))
	for _, fieldKey := range changedFieldKeys {
		changed[fieldKey] = struct{}{}
	}
	patchCells := make(map[string]any, len(changed))
	for fieldKey := range changed {
		if cell, ok := cells[fieldKey]; ok {
			patchCells[fieldKey] = cell
		}
	}
	if len(patchCells) == 0 {
		return nil
	}
	patch := map[string]any{"record_id": recordID, "row_version": rowVersion, "cells": patchCells}
	if groupValues, ok := row["group_values"].(map[string]any); ok {
		patchGroupValues := make(map[string]any)
		for fieldKey := range changed {
			if value, ok := groupValues[fieldKey]; ok {
				patchGroupValues[fieldKey] = value
			}
		}
		if len(patchGroupValues) > 0 {
			patch["group_values"] = patchGroupValues
		}
	}
	return patch
}

func validRecordChangeKind(changeKind string) bool {
	switch changeKind {
	case "invalidate", "patch", "remove":
		return true
	default:
		return false
	}
}
