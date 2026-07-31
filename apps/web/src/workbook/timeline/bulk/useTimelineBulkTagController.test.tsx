import { requireViewContract } from "@cartulary/view-contracts";
import { act, renderHook, waitFor } from "@testing-library/react";
import { StrictMode } from "react";
import { describe, expect, it, vi } from "vitest";
import { deferred } from "../../../testing/fetchMockTestSupport";
import { fullWorkbookViewRow } from "../../../testing/timelineWorkbookTestSupport";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { TimelineBulkMutationPort } from "../../mutations/workbookMutationCommandPorts";
import {
  createDraftRow,
  normalizeTimelineFullRow,
  rowFromApi,
  type WorkbookRow,
} from "../models/workbookTimelineModel";
import { useTimelineBulkTagController } from "./useTimelineBulkTagController";

const timelineContract = requireViewContract(timelineViewSchemaId);

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

function acceptedBulkResult(affectedRowCount: number, conflictCount = 0) {
  return {
    kind: "accepted" as const,
    value: {
      affectedRowCount,
      changeSetId: "30000000-0000-4000-8000-000000000001",
      conflictCount,
    },
  };
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
    const port = { assignTag: vi.fn(async () => acceptedBulkResult(2)) };
    const { result, rerender } = renderHook(
      ({ canAssign, rows }) =>
        useTimelineBulkTagController({
          canAssign,
          port,
          refreshRows: async () => undefined,
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

  it("submits one versioned command, refreshes, and retains valid selection and draft", async () => {
    const first = committedRow("11111111-1111-4111-8111-111111111111", 3);
    const second = committedRow("22222222-2222-4222-8222-222222222222", 4);
    const rowsRef = { current: [first, second] };
    const pending = deferred<ReturnType<typeof acceptedBulkResult>>();
    const assignTag = vi.fn(() => pending.promise);
    const refreshRows = vi.fn(async () => undefined);
    const port: Pick<TimelineBulkMutationPort, "assignTag"> = { assignTag };
    const { result } = renderHook(
      () =>
        useTimelineBulkTagController({
          canAssign: true,
          port,
          refreshRows,
          rows: rowsRef.current,
          rowsRef,
        }),
      { wrapper: StrictMode },
    );
    act(() => {
      result.current.snapshot.gridSelection.onSelectedRecordIdsChange(
        new Set([first.recordId ?? "", second.recordId ?? ""]),
      );
      result.current.commands.changeTagName("  bulk-tag  ");
    });

    let firstSubmission: Promise<void> | undefined;
    act(() => {
      firstSubmission = result.current.commands.assignTag();
      void result.current.commands.assignTag();
    });
    expect(assignTag).toHaveBeenCalledOnce();
    expect(assignTag).toHaveBeenCalledWith({
      tagName: "bulk-tag",
      targets: [
        {
          recordId: "11111111-1111-4111-8111-111111111111",
          baseRowVersion: 3,
        },
        {
          recordId: "22222222-2222-4222-8222-222222222222",
          baseRowVersion: 4,
        },
      ],
    });
    await act(async () => {
      pending.resolve(acceptedBulkResult(2));
      await firstSubmission;
    });
    expect(refreshRows).toHaveBeenCalledOnce();
    expect(result.current.snapshot.message).toEqual({
      kind: "success",
      message: "Assigned tag to 2 selected records.",
    });
    expect(result.current.snapshot.tagName).toBe("  bulk-tag  ");
    expect(result.current.snapshot.selectedRecordIds.size).toBe(2);
  });

  it("retains recoverable state for rejection and reports only bounded conflict counts", async () => {
    const row = committedRow("11111111-1111-4111-8111-111111111111", 3);
    const rowsRef = { current: [row] };
    const assignTag = vi
      .fn<Pick<TimelineBulkMutationPort, "assignTag">["assignTag"]>()
      .mockResolvedValueOnce({
        kind: "rejected",
        failure: {
          kind: "authorization_lost",
          message: "Your access changed.",
        },
      })
      .mockResolvedValueOnce(acceptedBulkResult(0, 1));
    const refreshRows = vi.fn(async () => undefined);
    const { result } = renderHook(() =>
      useTimelineBulkTagController({
        canAssign: true,
        port: { assignTag },
        refreshRows,
        rows: rowsRef.current,
        rowsRef,
      }),
    );
    act(() => {
      result.current.commands.changeSelectedRecordIds(
        new Set([row.recordId ?? ""]),
      );
      result.current.commands.changeTagName("review-tag");
    });

    await act(async () => result.current.commands.assignTag());
    expect(refreshRows).not.toHaveBeenCalled();
    expect(result.current.snapshot.message).toEqual({
      kind: "error",
      message: "Your access changed.",
    });
    expect(result.current.snapshot.tagName).toBe("review-tag");
    expect(result.current.snapshot.selectedRecordIds.size).toBe(1);

    await act(async () => result.current.commands.assignTag());
    expect(refreshRows).toHaveBeenCalledOnce();
    expect(result.current.snapshot.message).toEqual({
      kind: "error",
      message:
        "Assigned tag to 0 selected records; 1 record changed and needs review.",
    });
  });

  it("rejects stale selection and ignores late completion after authorization loss", async () => {
    const row = committedRow("11111111-1111-4111-8111-111111111111", 3);
    const rowsRef = { current: [row] };
    const pending = deferred<ReturnType<typeof acceptedBulkResult>>();
    const assignTag = vi.fn(() => pending.promise);
    const refreshRows = vi.fn(async () => undefined);
    const { result, rerender } = renderHook(
      ({ canAssign, rows }) =>
        useTimelineBulkTagController({
          canAssign,
          port: { assignTag },
          refreshRows,
          rows,
          rowsRef,
        }),
      { initialProps: { canAssign: true, rows: rowsRef.current } },
    );
    act(() => {
      result.current.commands.changeSelectedRecordIds(
        new Set([row.recordId ?? ""]),
      );
      result.current.commands.changeTagName("tag");
    });

    rowsRef.current = [];
    await act(async () => result.current.commands.assignTag());
    expect(assignTag).not.toHaveBeenCalled();
    expect(result.current.snapshot.message?.message).toContain(
      "Selection changed",
    );

    rowsRef.current = [row];
    rerender({ canAssign: true, rows: rowsRef.current });
    act(() => {
      result.current.commands.changeSelectedRecordIds(
        new Set([row.recordId ?? ""]),
      );
    });
    let submission: Promise<void> | undefined;
    act(() => {
      submission = result.current.commands.assignTag();
    });
    rerender({ canAssign: false, rows: rowsRef.current });
    await act(async () => {
      pending.resolve(acceptedBulkResult(1));
      await submission;
    });
    expect(refreshRows).not.toHaveBeenCalled();
    expect(result.current.snapshot.message).toBeNull();
    expect(result.current.snapshot.tagName).toBe("tag");
    expect(result.current.snapshot.submitting).toBe(false);
  });
});
