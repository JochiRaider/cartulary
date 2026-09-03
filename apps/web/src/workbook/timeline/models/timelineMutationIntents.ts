import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import {
  type CollectionFieldKey,
  timelineCollectionBindings,
  timelineScalarBindings,
} from "./timelineFieldRegistry";
import type { WorkbookRow } from "./timelineRowModel";

function buildCollectionActions(
  fieldKey: CollectionFieldKey,
  rawInput: string,
) {
  const actions = rawInput
    .split(/\r?\n/u)
    .filter((segment) => segment.trim() !== "")
    .map((rawText) =>
      fieldKey === "timeline.tags"
        ? { op: "add_tag", tag_name: rawText }
        : { op: "add_token", raw_text: rawText },
    );
  return actions.length < 1 ? null : { kind: "collection_actions_v1", actions };
}

export function buildScalarPatchIntent(row: WorkbookRow, clientTxnId: string) {
  const changes = timelineScalarBindings
    .flatMap((field) => {
      const current = row.values[field.key];
      return current === row.committedValues[field.key]
        ? []
        : [{ field_key: field.fieldKey, value: current }];
    })
    .sort((left, right) => left.field_key.localeCompare(right.field_key));
  return changes.length < 1
    ? null
    : {
        view_schema_id: timelineViewSchemaId,
        client_txn_id: clientTxnId,
        changes,
      };
}

export function buildCollectionPatchIntent(
  fieldKey: CollectionFieldKey,
  draftValue: string,
  clientTxnId: string,
) {
  const actionPayload = buildCollectionActions(fieldKey, draftValue);
  return actionPayload === null
    ? null
    : {
        view_schema_id: timelineViewSchemaId,
        client_txn_id: clientTxnId,
        changes: [{ field_key: fieldKey, action_payload: actionPayload }],
      };
}

export type BuildCreatePayloadOptions = {
  readonly allowZeroFieldCreate?: boolean;
};

export function buildCreatePayload(
  row: WorkbookRow,
  clientTxnId: string,
  options: BuildCreatePayloadOptions = {},
) {
  const payload: Record<string, unknown> = { client_txn_id: clientTxnId };
  for (const field of timelineScalarBindings) {
    const normalized = row.values[field.key];
    if (normalized !== "") payload[field.fieldKey] = normalized;
  }
  for (const field of timelineCollectionBindings) {
    const actions = buildCollectionActions(
      field.fieldKey,
      row.collectionDrafts[field.draftKey],
    );
    if (actions !== null) payload[field.fieldKey] = actions;
  }
  return Object.keys(payload).length < 2 && !options.allowZeroFieldCreate
    ? null
    : payload;
}
