import { describe, expect, it } from "vitest";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { PendingReplayUnitState } from "../../utils/workbookPendingQueue";
import {
  buildTimelineDeleteRestorePayload,
  buildTimelineRecordActionPayload,
  buildTimelineRollbackPayload,
} from "../services/timelineMutationRequests";
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
  "timeline.date_entered_text",
  "timeline.analyst_text",
  "timeline.mitre_stage_text",
  "timeline.device_object_text",
  "timeline.ip_address_text",
  "timeline.activity_utc_text",
  "timeline.activity_local_text",
  "timeline.raw_activity_text",
  "timeline.activity_synopsis_text",
  "timeline.data_source_text",
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
      "timeline.date_entered_text": { value: "2026-04-24" },
      "timeline.analyst_text": { value: "Analyst A" },
      "timeline.mitre_stage_text": { value: "TA0001" },
      "timeline.device_object_text": { value: "HOST-1" },
      "timeline.ip_address_text": { value: "192.0.2.10" },
      "timeline.activity_utc_text": { value: "2026-04-24T12:00:00Z" },
      "timeline.activity_local_text": { value: "2026-04-24T08:00:00-04:00" },
      "timeline.raw_activity_text": { value: "EDR alert" },
      "timeline.activity_synopsis_text": { value: "Initial access" },
      "timeline.data_source_text": { value: "EDR" },
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
      "timeline.activity_sort_ts": { value: "2026-04-24T12:00:00Z" },
      "timeline.date_entered_sort_day": { value: "2026-04-24" },
      "timeline.activity_time_pair_state": { value: "paired_user_preserved" },
      "timeline.replacement_record_id": { value: null },
      "timeline.has_evidence": { value: false },
      "timeline.has_unresolved_mentions": { value: false },
      ...cells,
    },
    group_values: {
      "timeline.date_entered_sort_day": "2026-04-24",
      "timeline.activity_time_pair_state": "paired_user_preserved",
      "timeline.capture_state": "rough",
      "timeline.has_evidence": false,
      "timeline.has_unresolved_mentions": false,
    },
  };
}

describe("workbookTimelineModel", () => {
  it("exposes Timeline bindings and focus keys without component-local field maps", () => {
    expect(
      timelineFieldBinding("timeline.activity_synopsis_text"),
    ).toMatchObject({
      kind: "scalar",
      key: "activitySynopsisText",
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
      "timeline.date_entered_text",
      "timeline.analyst_text",
      "timeline.mitre_stage_text",
      "timeline.device_object_text",
      "timeline.ip_address_text",
      "timeline.activity_utc_text",
      "timeline.activity_local_text",
      "timeline.raw_activity_text",
      "timeline.activity_synopsis_text",
      "timeline.data_source_text",
    ]);
    expect(timelineInspectorBindings.map((binding) => binding.key)).toEqual([
      "rawActivityText",
      "activitySynopsisText",
    ]);
    expect(timelineVisibleBindings.length).toBeGreaterThan(0);
    expect(timelineColumnWidth("timeline.ip_address_text")).toBe(160);
    expect(timelineColumnWidth("timeline.raw_activity_text")).toBe(320);
    expect(
      inputFocusKey("timeline-1", "activitySynopsisText", "inspector"),
    ).toBe("timeline-1:activitySynopsisText:inspector");
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
    expect(expandedWidths["timeline.activity_utc_text"]).toBe(
      timelineColumnWidth("timeline.activity_utc_text"),
    );
    expect(expandedWidths["timeline.ip_address_text"]).toBe(
      timelineColumnWidth("timeline.ip_address_text"),
    );
    expect(expandedWidths["timeline.raw_activity_text"]).toBe(
      timelineColumnWidth("timeline.raw_activity_text") + 42,
    );
    expect(expandedWidths["timeline.activity_synopsis_text"]).toBe(
      timelineColumnWidth("timeline.activity_synopsis_text") + 28,
    );
    expect(expandedWidths["timeline.device_object_text"]).toBe(
      timelineColumnWidth("timeline.device_object_text") + 14,
    );
    expect(expandedWidths["timeline.data_source_text"]).toBe(
      timelineColumnWidth("timeline.data_source_text") + 16,
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
        activityUTCText: "2026-04-24T12:00:00Z",
        activitySynopsisText: "Initial access",
      },
    });
    expect(
      readTimelineTagItems(normalized).map((item) => item.displayText),
    ).toEqual(["credential_access", "lateral_movement"]);
    expect(
      readTimelineCellValue(normalized, "timeline.activity_synopsis_text"),
    ).toBe("Initial access");
    expect(timelineGroupLabel(row, "timeline.missing")).toBe("Unassigned");

    const patch = normalizeTimelinePatchCells(
      {
        record_id: "timeline-1",
        row_version: 5,
        cells: {
          "timeline.activity_synopsis_text": { value: "Updated summary" },
        },
        group_values: { "timeline.capture_state": "refined" },
      },
      "timeline model patch",
    );
    expect(applyViewRowPatch(normalized, patch)).toMatchObject({
      row_version: 5,
      cells: {
        "timeline.activity_synopsis_text": { value: "Updated summary" },
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
        dateEnteredText: "2026-04-24",
        analystText: "Analyst A",
        mitreStageText: "TA0001",
        deviceObjectText: "HOST-1",
        ipAddressText: "192.0.2.10",
        activityUTCText: "2026-04-24T12:00:00Z",
        activityLocalText: "2026-04-24T08:00:00-04:00",
        rawActivityText: "",
        activitySynopsisText: "Changed summary",
        dataSourceText: "EDR",
      },
    };

    expect(buildScalarPatchIntent(row, "txn-1")).toEqual({
      view_schema_id: timelineViewSchemaId,
      client_txn_id: "txn-1",
      changes: [
        {
          field_key: "timeline.activity_synopsis_text",
          value: "Changed summary",
        },
        { field_key: "timeline.raw_activity_text", value: "" },
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
    expect(
      buildTimelineRecordActionPayload({
        action: "supersede",
        baseRowVersion: 4,
        clientTxnId: "txn-action",
        replacementRecordId: "record-new",
      }),
    ).toEqual({
      base_row_version: 4,
      client_txn_id: "txn-action",
      reason: "Superseded from workbook",
      replacement_record_id: "record-new",
    });
    expect(
      buildTimelineDeleteRestorePayload({
        baseRowVersion: 4,
        clientTxnId: "txn-delete",
        operation: "delete",
      }),
    ).toEqual({
      base_row_version: 4,
      client_txn_id: "txn-delete",
      reason: "Deleted from workbook history",
    });
    expect(
      buildTimelineRollbackPayload({
        baseRowVersion: 4,
        clientTxnId: "txn-rollback",
        target: { kind: "change_set", change_set_id: "change-1" },
      }),
    ).toEqual({
      base_row_version: 4,
      client_txn_id: "txn-rollback",
      reason: "Rollback from workbook history",
      target: { kind: "change_set", change_set_id: "change-1" },
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
