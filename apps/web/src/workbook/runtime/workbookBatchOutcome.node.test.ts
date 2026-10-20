import { describe, expect, it } from "vitest";
import type {
  WorkbookBatchEntry,
  WorkbookBatchPlan,
  WorkbookBatchReceipt,
} from "./workbookBatchOperation";
import { workbookBatchOutcome } from "./workbookBatchOutcome";
import { workbookBatchRecoveryItems } from "./workbookBatchRecoveryItems";

const view = "cartulary.view.timeline.v2";
const targets: [
  { record_id: string; base_row_version: number },
  { record_id: string; base_row_version: number },
] = [
  { record_id: "first", base_row_version: 1 },
  { record_id: "second", base_row_version: 1 },
];
const plans: WorkbookBatchPlan[] = [
  {
    operation: "pasteWorkbookClipboard",
    recordIds: ["first", "second"],
    request: {
      view_schema_id: view,
      clipboard_text: "a\nb\nc",
      start_field_key: "timeline.activity_synopsis_text",
      columns: ["timeline.activity_synopsis_text"],
      targets: [
        { kind: "record", ...targets[0] },
        { kind: "record", ...targets[1] },
        { kind: "create" },
      ],
    },
  },
  ...["clear_cells_v1", "fill_down_v1", "multi_row_tag_assignment_v1"].map(
    (kind) =>
      ({
        operation: "applyWorkbookBulkMutation",
        recordIds: ["first", "second"],
        request: {
          view_schema_id: view,
          kind,
          targets,
          ...(kind === "clear_cells_v1"
            ? {
                field_keys: [
                  "timeline.activity_synopsis_text",
                  "timeline.raw_activity_text",
                ],
              }
            : kind === "fill_down_v1"
              ? { field_key: "timeline.activity_synopsis_text", value: "fill" }
              : { tag_name: "triage" }),
        },
      }) as WorkbookBatchPlan,
  ),
];
const conflict = {
  record_id: "second",
  field_key: "timeline.activity_synopsis_text",
  base_row_version: 1,
  current_row_version: 2,
  conflict_token: "token",
  conflict_resolution_class: "text_compare_merge" as const,
  base_value: "old",
  client_value: "local",
  server_value: "saved",
};
const receipt: WorkbookBatchReceipt = {
  viewSchemaId: view,
  changeSetId: "change",
  rows: [{ record_id: "first", row_version: 2, cells: {} }],
  conflicts: [],
};
function entry(
  plan: WorkbookBatchPlan,
  changes: Partial<WorkbookBatchEntry> = {},
): WorkbookBatchEntry {
  return {
    id: "batch",
    plan,
    phase: "acknowledged",
    receipt,
    attempt: null,
    failure: null,
    transportPending: false,
    reconciliation: "complete",
    ...changes,
  };
}

describe("Batch outcome facts", () => {
  it("separates all outcomes from refresh and current conflicts for every Timeline operation", () => {
    for (const plan of plans) {
      for (const [phase, kind] of [
        ["waiting", "waiting"],
        ["preparing", "applying"],
        ["submitting", "applying"],
        ["uncertain", "uncertain"],
        ["rejected", "rejected"],
      ] as const) {
        expect(
          workbookBatchOutcome(entry(plan, { phase, receipt: null }), 0),
        ).toMatchObject({ kind, canReview: false });
      }
      expect(workbookBatchOutcome(entry(plan), 0)).toMatchObject({
        kind: "saved",
        canReview: true,
        attention: "completed",
        returnedRecordCount: 1,
      });
      const partial = entry(plan, {
        receipt: { ...receipt, conflicts: [conflict] },
      });
      expect(workbookBatchOutcome(partial, 1)).toMatchObject({
        kind: "saved_with_conflicts",
        originalConflictCount: 1,
        unresolvedConflictCount: 1,
        canReview: true,
        attention: "attention",
      });
      expect(workbookBatchOutcome(partial, 0)).toMatchObject({
        kind: "saved",
        originalConflictCount: 1,
        unresolvedConflictCount: 0,
        attention: "completed",
      });
      const only = entry(plan, {
        receipt: {
          ...receipt,
          changeSetId: null,
          rows: [],
          conflicts: [conflict],
        },
      });
      for (const count of [0, 1])
        expect(workbookBatchOutcome(only, count)).toMatchObject({
          kind: "conflicts_only",
          canReview: false,
          originalConflictCount: 1,
          unresolvedConflictCount: count,
        });
      expect(
        workbookBatchOutcome(
          entry(plan, { receipt: { ...receipt, changeSetId: null, rows: [] } }),
          0,
        ),
      ).toMatchObject({ kind: "no_op", canReview: false });
      for (const reconciliation of [
        "pending",
        "refreshing",
        "required",
      ] as const) {
        expect(
          workbookBatchOutcome(entry(plan, { reconciliation }), 0),
        ).toMatchObject({
          kind: "saved",
          refresh: reconciliation,
          attention: "attention",
          canReview: true,
        });
      }
    }
  });
  it("labels requested targets including creation slots without inventing changed cells or complete scope", () => {
    for (const plan of plans) {
      const outcome = workbookBatchOutcome(entry(plan), 0);
      expect(outcome.requestedScope).toContain(
        `Requested: ${plan.request.targets.length} target rows`,
      );
      expect(outcome.detail).toContain("1 record returned");
      expect(outcome.detail).not.toMatch(/cells|all|total/);
      if (plan.operation === "pasteWorkbookClipboard")
        expect(outcome.requestedScope).toContain("(1 new)");
      expect(outcome.originalInput).not.toContain("timeline.");
      expect(outcome.inputPreview).not.toContain("timeline.");
      expect(Array.from(outcome.inputPreview).length).toBeLessThanOrEqual(73);
    }
  });
  it("keeps Entity reuse out of Timeline review and preserves grouping and concealment", () => {
    const plan = plans[0];
    if (!plan) throw new Error("Missing plan");
    const entity = entry(plan, {
      receipt: {
        ...receipt,
        viewSchemaId: "cartulary.view.hosts.v1",
        rows: [...receipt.rows, ...receipt.rows],
      },
    });
    expect(workbookBatchOutcome(entity, 0)).toMatchObject({
      canReview: false,
      returnedRecordCount: 1,
    });
    const state = {
      authority: {
        actorId: "actor",
        incidentId: "incident",
        role: "editor" as const,
        sessionIdentity: "session",
        closed: false,
      },
      entries: [entry(plan)],
      admissionError: null,
    };
    expect(workbookBatchRecoveryItems(state, new Map())[0]?.attention).toBe(
      "completed",
    );
    expect(
      workbookBatchRecoveryItems(state, new Map([["batch", 1]]))[0]?.attention,
    ).toBe("attention");
    expect(
      workbookBatchRecoveryItems({ ...state, authority: null }, new Map()),
    ).toEqual([]);
  });
});
