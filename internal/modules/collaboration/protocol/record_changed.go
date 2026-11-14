package protocol

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"

	"github.com/google/uuid"
)

type recordChangedPayload struct {
	RecordID         string                     `json:"record_id"`
	RowVersion       int64                      `json:"row_version"`
	ChangeSetID      string                     `json:"change_set_id"`
	ClientTxnID      string                     `json:"client_txn_id"`
	ActorUserID      string                     `json:"actor_user_id"`
	ChangedFieldKeys []string                   `json:"changed_field_keys"`
	AffectedViews    []recordChangedViewPayload `json:"affected_views"`
}

type recordChangedViewPayload struct {
	ViewSchemaID string          `json:"view_schema_id"`
	ChangeKind   string          `json:"change_kind"`
	PatchCells   json.RawMessage `json:"patch_cells"`
}

func validateRecordChangedPayload(data json.RawMessage) error {
	var payload recordChangedPayload
	var views []json.RawMessage
	if err := decodeKnownPayloadFields(data, payloadField{"record_id", &payload.RecordID}, payloadField{"row_version", &payload.RowVersion}, payloadField{"change_set_id", &payload.ChangeSetID}, payloadField{"client_txn_id", &payload.ClientTxnID}, payloadField{"actor_user_id", &payload.ActorUserID}, payloadField{"changed_field_keys", &payload.ChangedFieldKeys}, payloadField{"affected_views", &views}); err != nil {
		return errors.New("record_changed payload is invalid")
	}
	for _, raw := range views {
		var view recordChangedViewPayload
		if err := decodeKnownPayloadFields(raw, payloadField{"view_schema_id", &view.ViewSchemaID}, payloadField{"change_kind", &view.ChangeKind}, payloadField{"patch_cells", &view.PatchCells}); err != nil {
			return err
		}
		payload.AffectedViews = append(payload.AffectedViews, view)
	}
	recordID, recordErr := uuid.Parse(payload.RecordID)
	changeSetID, changeSetErr := uuid.Parse(payload.ChangeSetID)
	actorUserID, actorErr := uuid.Parse(payload.ActorUserID)
	if recordErr != nil || recordID == uuid.Nil || changeSetErr != nil || changeSetID == uuid.Nil ||
		actorErr != nil || actorUserID == uuid.Nil || payload.RowVersion < 1 ||
		strings.TrimSpace(payload.ClientTxnID) == "" || len(payload.AffectedViews) == 0 {
		return errors.New("record_changed identity is invalid")
	}
	if !slices.IsSorted(payload.ChangedFieldKeys) {
		return errors.New("record_changed changed_field_keys must be sorted")
	}
	for index, fieldKey := range payload.ChangedFieldKeys {
		if strings.TrimSpace(fieldKey) == "" || fieldKey != strings.TrimSpace(fieldKey) ||
			(index > 0 && payload.ChangedFieldKeys[index-1] == fieldKey) {
			return errors.New("record_changed changed_field_keys are invalid")
		}
	}
	viewSchemaIDs := make([]string, 0, len(payload.AffectedViews))
	for _, view := range payload.AffectedViews {
		if strings.TrimSpace(view.ViewSchemaID) == "" || view.ViewSchemaID != strings.TrimSpace(view.ViewSchemaID) ||
			!validRecordChangeKind(view.ChangeKind) {
			return errors.New("record_changed affected view is invalid")
		}
		if view.ChangeKind == "patch" {
			if err := validateRecordPatch(view.PatchCells, payload); err != nil {
				return err
			}
		} else if len(view.PatchCells) != 0 {
			return errors.New("record_changed affected view patch is not admitted")
		}
		viewSchemaIDs = append(viewSchemaIDs, view.ViewSchemaID)
	}
	if !slices.IsSorted(viewSchemaIDs) {
		return errors.New("record_changed affected views must be sorted")
	}
	for index := 1; index < len(viewSchemaIDs); index++ {
		if viewSchemaIDs[index-1] == viewSchemaIDs[index] {
			return errors.New("record_changed affected views must be unique")
		}
	}
	return nil
}

func validateRecordPatch(raw json.RawMessage, payload recordChangedPayload) error {
	var recordID string
	var version int64
	var cells, groups map[string]json.RawMessage
	if err := decodeKnownPayloadFields(raw, payloadField{"record_id", &recordID}, payloadField{"row_version", &version}, payloadField{"cells", &cells}, payloadField{"group_values", &groups}); err != nil {
		return err
	}
	if recordID != payload.RecordID || version != payload.RowVersion {
		return errors.New("record_changed patch identity differs from its event")
	}
	for _, fields := range []map[string]json.RawMessage{cells, groups} {
		for key := range fields {
			if _, found := slices.BinarySearch(payload.ChangedFieldKeys, key); !found {
				return errors.New("record_changed patch contains an unchanged field")
			}
		}
	}
	for _, raw := range groups {
		var scalar any
		if err := json.Unmarshal(raw, &scalar); err != nil {
			return errors.New("record_changed grouping scalar is invalid")
		}
		switch scalar.(type) {
		case nil, bool, string, float64:
		default:
			return errors.New("record_changed grouping value must be scalar")
		}
	}
	return nil
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
