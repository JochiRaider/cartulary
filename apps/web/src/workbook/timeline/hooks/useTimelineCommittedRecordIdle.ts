import { useCallback } from "react";
import {
  refreshBlocksWorkbookPendingRecord,
  type WorkbookPendingSavesRefs,
} from "../../runtime/workbookPendingReplayRuntime";
import type {
  PendingReplayRuntimeMeta,
  TimelineMutableRef,
} from "../models/timelineControllerPorts";
import type { WorkbookRow } from "../models/workbookTimelineModel";

type TimelineCommittedRecordIdleOptions = {
  readonly fallbackRowVersion?: number | null | undefined;
  readonly refreshIfMissing?: boolean;
};

export type TimelineCommittedRecordIdleResult = {
  readonly row: WorkbookRow | null;
  readonly rowVersion: number;
};

export function useTimelineCommittedRecordIdle({
  conflictQueueRef,
  latestCommittedRowVersion,
  latestCommittedTimelineRow,
  loadRows,
  pendingSavesRefs,
}: {
  readonly conflictQueueRef: TimelineMutableRef<Record<string, unknown>>;
  readonly latestCommittedRowVersion: (
    recordId: string,
  ) => number | null | undefined;
  readonly latestCommittedTimelineRow: (recordId: string) => WorkbookRow | null;
  readonly loadRows: (options: {
    readonly showLoading: boolean;
  }) => Promise<void>;
  readonly pendingSavesRefs: WorkbookPendingSavesRefs<PendingReplayRuntimeMeta>;
}) {
  return useCallback(
    async (
      recordId: string,
      options: TimelineCommittedRecordIdleOptions = {},
    ): Promise<TimelineCommittedRecordIdleResult | null> => {
      let attemptedRefresh = false;
      for (;;) {
        const pending = pendingSavesRefs.pendingQueueRef.current;
        const snapshot = pending.model.snapshot();
        const hasPendingRecordWork = snapshot.units.some(
          (unit) => unit.recordId === recordId,
        );
        if (
          snapshot.authPaused ||
          snapshot.halted !== null ||
          snapshot.overflow !== null ||
          snapshot.sameFieldConflicts.length > 0 ||
          Object.keys(conflictQueueRef.current).length > 0
        ) {
          return null;
        }
        if (
          !hasPendingRecordWork &&
          !refreshBlocksWorkbookPendingRecord(pending, recordId)
        ) {
          const row = latestCommittedTimelineRow(recordId);
          const rowVersion =
            latestCommittedRowVersion(recordId) ?? options.fallbackRowVersion;
          if (typeof rowVersion === "number") return { row, rowVersion };
          if (
            options.refreshIfMissing !== false &&
            attemptedRefresh === false
          ) {
            attemptedRefresh = true;
            await loadRows({ showLoading: false });
            continue;
          }
          return null;
        }
        await new Promise((resolve) => window.setTimeout(resolve, 16));
      }
    },
    [
      conflictQueueRef,
      latestCommittedRowVersion,
      latestCommittedTimelineRow,
      loadRows,
      pendingSavesRefs,
    ],
  );
}
