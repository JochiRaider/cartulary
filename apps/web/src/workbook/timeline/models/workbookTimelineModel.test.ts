import { describe, expect, it } from "vitest";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { PendingReplayUnitState } from "../../utils/workbookPendingQueue";
import {
  applyViewRowPatch,
  buildAttachedEvidenceCreatePayload,
  buildAttachedEvidencePatchPayload,
  buildCollectionPatchIntent,
  buildExpandedTimelineColumnWidths,
  buildScalarPatchIntent,
  createDraftRow,
  createDraftRowForKey,
  inputFocusKey,
  materializePendingReplayPayload,
  normalizeTimelineFullRow,
  normalizeTimelinePatchCells,
  readTimelineCellValue,
  readTimelineTagItems,
  rowFromApi,
  timelineColumnWidth,
  timelineFieldBinding,
  timelineFocusFieldForFieldKey,
  timelineGroupLabel,
  timelineInspectorBindings,
  timelineRelationshipLabel,
  timelineScalarBindings,
  timelineVisibleBindings,
  validateTimelineViewSchemaId,
} from "./workbookTimelineModel";

const timelineWidthFieldKeys = [
  "timeline.occurred_at",
  "timeline.summary",
  "timeline.host_refs",
  "timeline.identity_refs",
  "timeline.evidence_count",
  "timeline.tags",
  "timeline.edited_at",
] as const;
const timelineWidthFixedChrome = {
  actionsColumnWidth: 0,
  rowGutterWidth: 58,
};
const timelineBaseDataWidth = timelineWidthFieldKeys.reduce(
  (sum, fieldKey) => sum + timelineColumnWidth(fieldKey),
  0,
);
const timelineBaseShellWidth =
  timelineBaseDataWidth +
  timelineWidthFixedChrome.actionsColumnWidth +
  timelineWidthFixedChrome.rowGutterWidth;

function timelineRow(cells: Record<string, { value: unknown }> = {}) {
  return {
    view_schema_id: timelineViewSchemaId,
    record_id: "timeline-1",
    row_version: 4,
    cells: {
      "timeline.occurred_at": { value: "2026-04-24T12:00:00Z" },
      "timeline.summary": { value: "Initial access" },
      "timeline.details": { value: "Observed by EDR" },
      "timeline.source_text": { value: "EDR alert" },
      "timeline.capture_state": { value: "rough" },
      "timeline.host_refs": { value: { items: [] } },
      "timeline.identity_refs": { value: { items: [] } },
      "timeline.evidence_count": { value: 0 },
      "timeline.tags": {
        value: {
          items: [
            { raw_text: "credential_access", item_ref: "tag-1" },
            { display_text: "lateral_movement" },
            { raw_text: "" },
          ],
        },
      },
      "timeline.attached_evidence_ids": { value: { items: [] } },
      "timeline.edited_at": { value: "" },
      "timeline.recorded_at": { value: "" },
      "timeline.sort_ts": { value: "2026-04-24T12:00:00Z" },
      "timeline.replacement_record_id": { value: null },
      "timeline.occurred_day": { value: "2026-04-24" },
      "timeline.recorded_day": { value: "" },
      "timeline.has_evidence": { value: false },
      "timeline.has_unresolved_mentions": { value: false },
      ...cells,
    },
    group_values: { "timeline.capture_state": "rough" },
  };
}

describe("workbookTimelineModel", () => {
  it("exposes Timeline bindings and focus keys without component-local field maps", () => {
    expect(timelineFieldBinding("timeline.summary")).toMatchObject({
      kind: "scalar",
      key: "summary",
    });
    expect(timelineFieldBinding("timeline.host_refs")).toMatchObject({
      kind: "collection",
      draftKey: "hostRefs",
    });
    expect(timelineFieldBinding("timeline.row_version")).toMatchObject({
      kind: "readonly",
    });
    expect(timelineFocusFieldForFieldKey("timeline.identity_refs")).toBe(
      "identityRefs",
    );
    expect(timelineScalarBindings.map((binding) => binding.fieldKey)).toEqual([
      "timeline.occurred_at",
      "timeline.summary",
      "timeline.details",
      "timeline.source_text",
    ]);
    expect(timelineInspectorBindings.map((binding) => binding.key)).toEqual([
      "details",
      "sourceText",
    ]);
    expect(timelineVisibleBindings.length).toBeGreaterThan(0);
    expect(timelineColumnWidth("timeline.evidence_count")).toBe(112);
    expect(timelineColumnWidth("timeline.edited_at")).toBe(248);
    expect(inputFocusKey("timeline-1", "summary", "inspector")).toBe(
      "timeline-1:summary:inspector",
    );
    expect(timelineRelationshipLabel("timeline.identity_refs")).toBe(
      "Identities",
    );
  });

  it("expands Timeline columns only when the shell has surplus width", () => {
    const baseWidths = Object.fromEntries(
      timelineWidthFieldKeys.map((fieldKey) => [
        fieldKey,
        timelineColumnWidth(fieldKey),
      ]),
    );

    expect(
      buildExpandedTimelineColumnWidths({
        ...timelineWidthFixedChrome,
        fieldKeys: timelineWidthFieldKeys,
        gridShellWidth: timelineBaseShellWidth - 24,
      }),
    ).toEqual(baseWidths);

    const expandedWidths = buildExpandedTimelineColumnWidths({
      ...timelineWidthFixedChrome,
      fieldKeys: timelineWidthFieldKeys,
      gridShellWidth: timelineBaseShellWidth + 100,
    });
    expect(expandedWidths["timeline.occurred_at"]).toBe(
      timelineColumnWidth("timeline.occurred_at"),
    );
    expect(expandedWidths["timeline.evidence_count"]).toBe(
      timelineColumnWidth("timeline.evidence_count"),
    );
    expect(expandedWidths["timeline.summary"]).toBe(
      timelineColumnWidth("timeline.summary") + 30,
    );
    expect(expandedWidths["timeline.host_refs"]).toBe(
      timelineColumnWidth("timeline.host_refs") + 20,
    );
    expect(expandedWidths["timeline.identity_refs"]).toBe(
      timelineColumnWidth("timeline.identity_refs") + 20,
    );
    expect(expandedWidths["timeline.tags"]).toBe(
      timelineColumnWidth("timeline.tags") + 10,
    );
    expect(expandedWidths["timeline.edited_at"]).toBe(
      timelineColumnWidth("timeline.edited_at") + 20,
    );
  });

  it("uses the full measured Timeline grid width when distributing surplus pixels", () => {
    const shellWidth = timelineBaseShellWidth + 137;
    const widths = buildExpandedTimelineColumnWidths({
      ...timelineWidthFixedChrome,
      fieldKeys: timelineWidthFieldKeys,
      gridShellWidth: shellWidth,
    });
    const totalWidth =
      timelineWidthFieldKeys.reduce(
        (sum, fieldKey) => sum + (widths[fieldKey] ?? 0),
        0,
      ) +
      timelineWidthFixedChrome.actionsColumnWidth +
      timelineWidthFixedChrome.rowGutterWidth;

    expect(totalWidth).toBe(shellWidth);
  });

  it("normalizes rows, tag collections, grouping, and sparse live patches", () => {
    const normalized = normalizeTimelineFullRow(
      timelineRow(),
      "timeline model test row",
    );
    const row = rowFromApi(normalized);
    expect(row).toMatchObject({
      recordId: "timeline-1",
      rowVersion: 4,
      captureState: "rough",
      values: {
        occurredAt: "2026-04-24T12:00:00Z",
        summary: "Initial access",
      },
    });
    expect(
      readTimelineTagItems(normalized).map((item) => item.displayText),
    ).toEqual(["credential_access", "lateral_movement"]);
    expect(readTimelineCellValue(normalized, "timeline.summary")).toBe(
      "Initial access",
    );
    expect(timelineGroupLabel(row, "timeline.missing")).toBe("Unassigned");

    const patch = normalizeTimelinePatchCells(
      {
        record_id: "timeline-1",
        row_version: 5,
        cells: {
          "timeline.summary": { value: "Updated summary" },
        },
        group_values: { "timeline.capture_state": "refined" },
      },
      "timeline model patch",
    );
    expect(applyViewRowPatch(normalized, patch)).toMatchObject({
      row_version: 5,
      cells: {
        "timeline.summary": { value: "Updated summary" },
      },
      group_values: { "timeline.capture_state": "refined" },
    });
    expect(() =>
      validateTimelineViewSchemaId("cartulary.view.notes.v1", "bad row"),
    ).toThrow(/view_schema_id/u);
  });

  it("builds scalar, collection, evidence, and replay payloads", () => {
    const row = {
      ...rowFromApi(normalizeTimelineFullRow(timelineRow(), "payload row")),
      values: {
        occurredAt: "2026-04-24T12:00:00Z",
        summary: "Changed summary",
        details: "",
        sourceText: "EDR alert",
      },
    };

    expect(buildScalarPatchIntent(row, "txn-1")).toEqual({
      view_schema_id: timelineViewSchemaId,
      client_txn_id: "txn-1",
      changes: [
        { field_key: "timeline.details", value: null },
        { field_key: "timeline.summary", value: "Changed summary" },
      ],
    });
    expect(
      buildCollectionPatchIntent("timeline.tags", " alpha\nbeta ", "txn-2"),
    ).toEqual({
      view_schema_id: timelineViewSchemaId,
      client_txn_id: "txn-2",
      changes: [
        {
          field_key: "timeline.tags",
          action_payload: {
            kind: "collection_actions_v1",
            actions: [
              { op: "add_tag", tag_name: " alpha" },
              { op: "add_tag", tag_name: "beta " },
            ],
          },
        },
      ],
    });
    expect(buildAttachedEvidenceCreatePayload("evidence-1", "txn-3")).toEqual({
      client_txn_id: "txn-3",
      "timeline.attached_evidence_ids": {
        kind: "collection_actions_v1",
        actions: [{ op: "add_record_ref", linked_record_id: "evidence-1" }],
      },
    });
    expect(
      buildAttachedEvidencePatchPayload(row, "evidence-2", "txn-4"),
    ).toMatchObject({
      view_schema_id: timelineViewSchemaId,
      base_row_version: 4,
      client_txn_id: "txn-4",
    });

    const patchUnit = {
      kind: "patch",
      payloadIntent: { changes: [] },
    } as unknown as PendingReplayUnitState;
    expect(materializePendingReplayPayload(patchUnit, row)).toEqual({
      changes: [],
      base_row_version: 4,
    });
    expect(
      materializePendingReplayPayload(patchUnit, createDraftRow(1)),
    ).toBeNull();
    expect(createDraftRowForKey("draft-22")).toMatchObject({ key: "draft-22" });
    expect(createDraftRowForKey("timeline-1")).toBeNull();
  });
});
