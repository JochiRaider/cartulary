import { requireViewContract } from "@cartulary/view-contracts";
import { act, renderHook, waitFor } from "@testing-library/react";
import { StrictMode } from "react";
import { describe, expect, it, vi } from "vitest";
import { fullWorkbookViewRow } from "../../../testing/timelineWorkbookTestSupport";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import {
  createDraftRow,
  normalizeTimelineFullRow,
  rowFromApi,
  type WorkbookRow,
} from "../models/timelineRowModel";
import { useTimelineBulkTagController } from "./useTimelineBulkTagController";

const timelineContract = requireViewContract(timelineViewSchemaId);

function actionContext(authorized: boolean) {
  return {
    authorized,
    capabilityAvailable: true,
  };
}

function committedRow(recordId: string, rowVersion: number): WorkbookRow {
  return rowFromApi(
    normalizeTimelineFullRow(
      fullWorkbookViewRow(timelineContract, recordId, rowVersion, {
        "timeline.activity_synopsis_text": recordId,
      }),
      "bulk tag controller fixture",
    ),
  );
}

describe("useTimelineBulkTagController", () => {
  it("keeps selection record-keyed, page-scoped, and pruned by accepted rows", async () => {
    const first = committedRow("11111111-1111-4111-8111-111111111111", 3);
    const second = committedRow("22222222-2222-4222-8222-222222222222", 4);
    const pending = {
      ...committedRow("33333333-3333-4333-8333-333333333333", 5),
      pendingSignature: "pending",
    };
    const rowsRef = { current: [first, second, pending, createDraftRow(1)] };
    const port = { assignTag: vi.fn(() => "batch") };
    const { result, rerender } = renderHook(
      ({ canAssign, rows }) =>
        useTimelineBulkTagController({
          context: actionContext(canAssign),
          port,
          precedingSaves: async () => undefined,
          rows,
          rowsRef,
        }),
      { initialProps: { canAssign: true, rows: rowsRef.current } },
    );

    act(() => {
      result.current.snapshot.gridSelection.onSelectedRecordIdsChange(
        new Set([first.recordId ?? "", second.recordId ?? ""]),
      );
    });
    expect([...result.current.snapshot.selectedRecordIds]).toEqual([
      first.recordId,
      second.recordId,
    ]);
    expect(
      result.current.snapshot.gridSelection.isRecordSelectable?.({
        data: pending,
        kind: "data",
        mutationIdentity: {
          baseRowVersion: 5,
          kind: "core_row_version",
        },
        rowIdentity: {
          kind: "core_record",
          recordId: pending.recordId ?? "",
        },
      }),
    ).toBe(false);

    rowsRef.current = [second];
    rerender({ canAssign: true, rows: rowsRef.current });
    await waitFor(() =>
      expect([...result.current.snapshot.selectedRecordIds]).toEqual([
        second.recordId,
      ]),
    );

    rerender({ canAssign: false, rows: rowsRef.current });
    await waitFor(() =>
      expect(result.current.snapshot.selectedRecordIds.size).toBe(0),
    );
  });

  it("captures one versioned action per gesture and retains selection and draft", async () => {
    const first = committedRow("11111111-1111-4111-8111-111111111111", 3);
    const rowsRef = { current: [first] };
    const assignTag = vi.fn(() => "batch");
    const ready = Promise.resolve();
    const { result, rerender, unmount } = renderHook(
      ({ authorized }) =>
        useTimelineBulkTagController({
          context: actionContext(authorized),
          port: { assignTag },
          precedingSaves: () => ready,
          rows: rowsRef.current,
          rowsRef,
        }),
      { wrapper: StrictMode, initialProps: { authorized: true } },
    );
    act(() => {
      result.current.commands.changeSelectedRecordIds(
        new Set([first.recordId ?? ""]),
      );
      result.current.commands.changeTagName("  triaged  ");
    });
    act(() => {
      void result.current.commands.assignTag();
      void result.current.commands.assignTag();
    });
    expect(assignTag).toHaveBeenCalledOnce();
    expect(assignTag).toHaveBeenCalledWith(
      {
        tagName: "triaged",
        targets: [{ recordId: first.recordId, baseRowVersion: 3 }],
      },
      { delivery: expect.any(Object), ready },
    );
    await act(async () => {
      await Promise.resolve();
    });
    expect(result.current.snapshot.tagName).toBe("  triaged  ");
    expect(result.current.snapshot.selectedRecordIds.size).toBe(1);
    act(() => {
      void result.current.commands.assignTag();
    });
    expect(assignTag).toHaveBeenCalledTimes(2);
    expect(assignTag.mock.calls[0]).not.toBe(assignTag.mock.calls[1]);
    await act(async () => {
      await Promise.resolve();
    });
    rerender({ authorized: false });
    act(() => {
      void result.current.commands.assignTag();
    });
    expect(assignTag).toHaveBeenCalledTimes(2);
    unmount();
  });
});
