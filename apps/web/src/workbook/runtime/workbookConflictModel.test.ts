import { describe, expect, it } from "vitest";
import {
  buildWorkbookConflictResolutionPayload,
  collectionActionsAgainstSaved,
  parseSameFieldConflict,
  parseSameFieldConflictFields,
  workbookConflictEntry,
  workbookConflictResolutionClass,
} from "./workbookConflictModel";

const validConflict = {
  base_row_version: 3,
  base_value: "base",
  client_value: "local",
  conflict_resolution_class: "text_compare_merge",
  conflict_token: "token-1",
  current_row_version: 4,
  field_key: "timeline.activity_synopsis_text",
  record_id: "row-1",
  server_updated_at: "2026-04-24T12:00:00Z",
  server_updated_by: "analyst@example.com",
  server_value: "saved",
  suggested_merged_value: "merged",
};

describe("workbookConflictModel", () => {
  it("parses same-field conflict envelopes without optional metadata loss", () => {
    expect(
      parseSameFieldConflict({
        error: {
          code: "same_field_conflict",
          conflict: validConflict,
        },
      }),
    ).toEqual(validConflict);
  });

  it("rejects missing, malformed, and incomplete conflict envelopes", () => {
    expect(parseSameFieldConflict(null)).toBeNull();
    expect(
      parseSameFieldConflict({ error: { code: "validation_failed" } }),
    ).toBeNull();
    for (const key of [
      "conflict_token",
      "record_id",
      "field_key",
      "conflict_resolution_class",
      "base_row_version",
      "current_row_version",
    ] as const) {
      const conflict = { ...validConflict };
      delete conflict[key];
      expect(
        parseSameFieldConflict({
          error: { code: "same_field_conflict", conflict },
        }),
      ).toBeNull();
    }
  });

  it("rejects empty string identity fields for pending queue admission", () => {
    expect(
      parseSameFieldConflictFields({
        ...validConflict,
        record_id: "",
      }),
    ).toBeNull();
    expect(parseSameFieldConflictFields(validConflict)).toMatchObject({
      conflict_token: "token-1",
      record_id: "row-1",
    });
  });

  it("falls unknown classes back to atomic replacement", () => {
    expect(workbookConflictResolutionClass("future_class")).toBe(
      "atomic_replace",
    );
  });

  it("starts text merge from the saved value and never the suggestion", () => {
    const entry = workbookConflictEntry({
      conflict: validConflict,
      focusKey: "row-1:timeline.activity_synopsis_text",
      rowLabel: "Row 1",
      surfaceLabel: "Timeline",
      viewSchemaId: "cartulary.view.timeline.v2",
    });
    expect(entry.mergedDraft).toBe("saved");
    expect(
      buildWorkbookConflictResolutionPayload({
        clientTxnId: "txn-1",
        entry,
        resolutionKind: "merged_value",
      }),
    ).toMatchObject({
      conflict_token: "token-1",
      resolution_kind: "merged_value",
      resolved_value: "saved",
    });
  });

  it("preserves typed collection items and derives actions against saved", () => {
    const saved = {
      kind: "collection_value_v1" as const,
      ordered: false,
      items: [
        {
          item_kind: "tag",
          item_ref: "tag-old",
          display_text: "Old",
        },
      ],
    };
    const local = {
      kind: "collection_value_v1" as const,
      ordered: false,
      items: [
        {
          item_kind: "tag",
          item_ref: "tag-new",
          display_text: "New",
        },
      ],
    };
    expect(collectionActionsAgainstSaved(saved, local)).toEqual({
      kind: "collection_actions_v1",
      actions: [
        { op: "remove_tag", item_ref: "tag-old" },
        { op: "add_tag", tag_name: "New" },
      ],
    });
    const entry = workbookConflictEntry({
      conflict: {
        ...validConflict,
        conflict_resolution_class: "collection_review",
        server_value: saved,
        client_value: local,
      },
      rowLabel: "Row 1",
      surfaceLabel: "Timeline",
      viewSchemaId: "cartulary.view.timeline.v2",
    });
    expect(
      buildWorkbookConflictResolutionPayload({
        clientTxnId: "txn-collection",
        entry,
        resolutionKind: "merged_value",
      }),
    ).toMatchObject({
      resolved_value: {
        kind: "collection_actions_v1",
        actions: [
          { op: "remove_tag", item_ref: "tag-old" },
          { op: "add_tag", tag_name: "New" },
        ],
      },
    });
  });
});
