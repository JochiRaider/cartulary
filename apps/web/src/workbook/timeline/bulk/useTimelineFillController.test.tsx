import type {
  GridCellTarget,
  GridFillIntent,
  GridInteractionMode,
} from "@cartulary/grid-adapter";
import { requireViewContract } from "@cartulary/view-contracts";
import { act, renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { fullWorkbookViewRow } from "../../../testing/timelineWorkbookTestSupport";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { TimelineFillMutationPort } from "../../mutations/workbookMutationCommandPorts";
import {
  normalizeTimelineFullRow,
  rowFromApi,
  type WorkbookRow,
} from "../models/timelineRowModel";
import {
  planTimelineFill,
  timelineFillRejectedMessage,
  useTimelineFillController,
} from "./useTimelineFillController";

const timelineContract = requireViewContract(timelineViewSchemaId);
const sourceId = "11111111-1111-4111-8111-111111111111";
const firstTargetId = "22222222-2222-4222-8222-222222222222";
const secondTargetId = "33333333-3333-4333-8333-333333333333";
const summaryFieldKey = "timeline.activity_synopsis_text";

function committedRow(
  recordId: string,
  rowVersion: number,
  summary = recordId,
): WorkbookRow {
  return rowFromApi(
    normalizeTimelineFullRow(
      fullWorkbookViewRow(timelineContract, recordId, rowVersion, {
        [summaryFieldKey]: summary,
      }),
      "fill controller fixture",
    ),
  );
}

function cellTarget(
  recordId: string,
  rowVersion: number,
  fieldKey = summaryFieldKey,
): GridCellTarget {
  return {
    fieldKey,
    mutationIdentity: {
      kind: "core_row_version",
      baseRowVersion: rowVersion,
    },
    rowIdentity: { kind: "core_record", recordId },
    surface: { kind: "view_schema", viewSchemaId: timelineViewSchemaId },
  };
}

function fillIntent(
  targets: readonly GridCellTarget[] = [
    cellTarget(firstTargetId, 4),
    cellTarget(secondTargetId, 5),
  ],
  source = cellTarget(sourceId, 3),
): GridFillIntent {
  return {
    range: { start: source, end: targets.at(-1) ?? source },
    source,
    target: targets.at(-1) ?? source,
    targets,
  };
}

function planInput(
  overrides: {
    readonly groupBy?: string | null;
    readonly interactionMode?: GridInteractionMode;
    readonly intent?: GridFillIntent;
    readonly rows?: readonly WorkbookRow[];
    readonly visibleFieldKeys?: ReadonlySet<string>;
  } = {},
) {
  return {
    contract: timelineContract,
    groupBy: null,
    interactionMode: { kind: "editable" } as const,
    intent: fillIntent(),
    rows: [
      committedRow(sourceId, 3, "  leading source value"),
      committedRow(firstTargetId, 4),
      committedRow(secondTargetId, 5),
    ],
    visibleFieldKeys: new Set([summaryFieldKey]),
    ...overrides,
  };
}

describe("useTimelineFillController", () => {
  it("plans one ordered versioned command from a committed visible scalar source", () => {
    expect(planTimelineFill(planInput())).toEqual({
      kind: "accepted",
      command: {
        fieldKey: summaryFieldKey,
        sourceAnchor: cellTarget(sourceId, 3),
        targets: [
          { recordId: firstTargetId, baseRowVersion: 4 },
          { recordId: secondTargetId, baseRowVersion: 5 },
        ],
        value: "  leading source value",
      },
    });
  });

  it("rejects every malformed, hidden, grouped, stale, pending, or non-editable fill atomically", () => {
    const source = committedRow(sourceId, 3);
    const first = committedRow(firstTargetId, 4);
    const second = committedRow(secondTargetId, 5);
    const pendingSource = { ...source, pendingSignature: "source-pending" };
    const pendingTarget = { ...first, pendingSignature: "target-pending" };
    const wrongSurfaceTarget: GridCellTarget = {
      ...cellTarget(firstTargetId, 4),
      surface: { kind: "view_schema", viewSchemaId: "cartulary.view.notes.v1" },
    };
    const presentationTarget: GridCellTarget = {
      ...cellTarget(firstTargetId, 4),
      rowIdentity: {
        kind: "extension_resource",
        extensionProfileId: "presentation",
        resourceKind: "row",
        resourceId: firstTargetId,
      },
      surface: {
        kind: "extension_grid",
        extensionProfileId: "presentation",
        gridSchemaId: "presentation-grid",
        workspaceKey: "presentation-workspace",
      },
    };
    const rejectedInputs = [
      planInput({
        interactionMode: { kind: "read_only", label: "Read only" },
      }),
      planInput({ groupBy: "timeline.analyst_text" }),
      planInput({ visibleFieldKeys: new Set() }),
      planInput({
        intent: fillIntent([cellTarget(firstTargetId, 4)], {
          ...cellTarget(sourceId, 3),
          surface: {
            kind: "view_schema",
            viewSchemaId: "cartulary.view.notes.v1",
          },
        }),
      }),
      planInput({ rows: [first, second] }),
      planInput({ intent: fillIntent(undefined, cellTarget(sourceId, 2)) }),
      planInput({ rows: [pendingSource, first, second] }),
      planInput({ intent: fillIntent([]) }),
      planInput({
        intent: fillIntent(
          [cellTarget(firstTargetId, 4, "timeline.host_refs")],
          cellTarget(sourceId, 3, "timeline.host_refs"),
        ),
        visibleFieldKeys: new Set(["timeline.host_refs"]),
      }),
      planInput({
        intent: fillIntent(
          [cellTarget(firstTargetId, 4, "timeline.capture_state")],
          cellTarget(sourceId, 3, "timeline.capture_state"),
        ),
        visibleFieldKeys: new Set(["timeline.capture_state"]),
      }),
      planInput({ intent: fillIntent([wrongSurfaceTarget]) }),
      planInput({ intent: fillIntent([presentationTarget]) }),
      planInput({
        intent: fillIntent([
          cellTarget(firstTargetId, 4, "timeline.raw_activity_text"),
        ]),
      }),
      planInput({ rows: [source, second] }),
      planInput({ intent: fillIntent([cellTarget(firstTargetId, 3)]) }),
      planInput({ rows: [source, pendingTarget, second] }),
      planInput({
        intent: fillIntent([
          cellTarget(firstTargetId, 4),
          cellTarget(firstTargetId, 4),
        ]),
      }),
      planInput({
        intent: fillIntent([
          cellTarget(firstTargetId, 4),
          cellTarget(secondTargetId, 4),
        ]),
      }),
    ];

    for (const input of rejectedInputs) {
      expect(planTimelineFill(input)).toEqual({
        kind: "rejected",
        message: timelineFillRejectedMessage,
      });
    }
  });

  it("captures semantic fill and preceding saves without local completion effects", () => {
    const fillDown = vi.fn<TimelineFillMutationPort["fillDown"]>(() => "batch");
    const ready = Promise.resolve();
    const rowsRef = { current: planInput().rows };
    const setError = vi.fn();
    const { result } = renderHook(() =>
      useTimelineFillController({
        contract: timelineContract,
        precedingSaves: () => ready,
        getVisibleFieldKeys: () => new Set([summaryFieldKey]),
        groupBy: null,
        interactionMode: { kind: "editable" },
        port: { fillDown },
        rowsRef,
        setError,
      }),
    );
    const intent = fillIntent();
    act(() => result.current.commands.onFillCells(intent));
    expect(fillDown).toHaveBeenCalledWith(
      {
        fieldKey: summaryFieldKey,
        targets: [
          { recordId: firstTargetId, baseRowVersion: 4 },
          { recordId: secondTargetId, baseRowVersion: 5 },
        ],
        value: "  leading source value",
      },
      { delivery: intent, ready },
    );
    act(() =>
      result.current.commands.onFillCells(
        fillIntent([
          cellTarget(firstTargetId, 4),
          cellTarget(secondTargetId, 4),
        ]),
      ),
    );
    expect(setError).toHaveBeenLastCalledWith(timelineFillRejectedMessage);
    expect(fillDown).toHaveBeenCalledOnce();
  });
});
