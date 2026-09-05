import { renderHook } from "@testing-library/react";
import { expect, it, vi } from "vitest";
import { createWorkbookPendingQueueRuntime } from "../runtime/workbookPendingReplayRuntime";
import { useTimelineCommittedRecordIdle } from "./hooks/useTimelineCommittedRecordIdle";
import type { TimelinePendingSavesRefs } from "./models/timelinePendingSaves";

function pendingSavesRefs(): TimelinePendingSavesRefs {
  return {
    collectionKeyboardCommitRef: { current: new Map() },
    pendingQueueRef: {
      current: createWorkbookPendingQueueRuntime({
        clientInstanceId: "client-1",
        incidentId: "incident-1",
      }),
    },
    pendingReplayOrderRef: { current: 1 },
    replayContextByUnitId: new Map(),
    pendingSignaturesRef: { current: new Map() },
    saveQueueRef: { current: Promise.resolve() },
  };
}

it("useTimelineCommittedRecordIdle refreshes a missing committed version at most once", async () => {
  let rowVersion: number | null = null;
  const loadRows = vi.fn(async () => {
    rowVersion = 7;
  });
  const refs = pendingSavesRefs();
  const conflictQueueRef = { current: {} };
  const { result } = renderHook(() =>
    useTimelineCommittedRecordIdle({
      conflictQueueRef,
      latestCommittedRowVersion: () => rowVersion,
      latestCommittedTimelineRow: () => null,
      loadRows,
      pendingSavesRefs: refs,
    }),
  );

  await expect(result.current("record-1")).resolves.toEqual({
    row: null,
    rowVersion: 7,
  });
  expect(loadRows).toHaveBeenCalledTimes(1);

  rowVersion = null;
  await expect(
    result.current("record-1", { refreshIfMissing: false }),
  ).resolves.toBeNull();
  expect(loadRows).toHaveBeenCalledTimes(1);

  conflictQueueRef.current = { conflict: true };
  await expect(result.current("record-1")).resolves.toBeNull();
  expect(loadRows).toHaveBeenCalledTimes(1);
});
