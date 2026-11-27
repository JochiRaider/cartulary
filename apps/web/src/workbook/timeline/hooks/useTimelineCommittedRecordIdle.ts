import { useCallback } from "react";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import type { TimelineCommittedRecordIdleResult } from "../models/timelineControllerPorts";
import type { WorkbookRow } from "../models/timelineRowModel";

type TimelineCommittedRecordIdleOptions = {
  readonly signal: AbortSignal;
  readonly fallbackRowVersion?: number | null | undefined;
  readonly refreshIfMissing?: boolean;
};

/**
 * Waits until the record has no pending write, then reads committed evidence
 * in that same publication. `evidence` returns null when none is known;
 * `refresh`, when given, runs once before the wait gives up.
 */
export function waitForTimelineRecordEvidence<T>(
  mutationRuntime: WorkbookMutationRuntime,
  recordId: string,
  signal: AbortSignal,
  evidence: () => T | null,
  refresh?: () => Promise<void>,
): Promise<T | null> {
  if (signal.aborted || mutationRuntime.retired) return Promise.resolve(null);
  const authorityEpoch = mutationRuntime.authorizationEpoch;
  // Only this wait is cancelled. A shared, authority-fenced query may finish
  // for its other consumers after the requesting action has gone away.
  return new Promise((resolve, reject) => {
    const controller = new AbortController();
    let finished = false;
    let publication = 0;
    let unsubscribe = () => {};
    const cleanup = () => {
      if (finished) return false;
      finished = true;
      unsubscribe();
      signal.removeEventListener("abort", cancel);
      controller.abort();
      return true;
    };
    const cancel = () => {
      if (cleanup()) resolve(null);
    };
    unsubscribe = mutationRuntime.subscribe(() => {
      publication++;
      if (
        mutationRuntime.retired ||
        mutationRuntime.authorizationEpoch !== authorityEpoch
      )
        cancel();
    });
    signal.addEventListener("abort", cancel, { once: true });
    if (signal.aborted) {
      cancel();
      return;
    }
    const read = async (): Promise<T | null> => {
      let attemptedRefresh = false;
      for (;;) {
        const observedPublication = publication;
        const readiness = await mutationRuntime.waitForPendingRecordIdle({
          recordId,
          viewSchemaId: timelineViewSchemaId,
          signal: controller.signal,
        });
        if (readiness !== "idle" || controller.signal.aborted) return null;
        // A publication can occur after the coordinator resolves and
        // before this continuation runs. Recheck before accepting evidence.
        if (publication !== observedPublication) continue;
        const value = evidence();
        if (value !== null) return value;
        if (!refresh || attemptedRefresh) return null;
        attemptedRefresh = true;
        await refresh();
        if (controller.signal.aborted) return null;
      }
    };
    void read().then(
      (value) => {
        if (cleanup()) resolve(value);
      },
      (error: unknown) => {
        if (cleanup()) reject(error);
      },
    );
  });
}

export function useTimelineCommittedRecordIdle({
  mutationRuntime,
  latestCommittedRowVersion,
  latestCommittedTimelineRow,
  loadRows,
}: {
  readonly mutationRuntime: WorkbookMutationRuntime;
  readonly latestCommittedRowVersion: (
    recordId: string,
  ) => number | null | undefined;
  readonly latestCommittedTimelineRow: (recordId: string) => WorkbookRow | null;
  readonly loadRows: (options: {
    readonly showLoading: boolean;
  }) => Promise<void>;
}) {
  return useCallback(
    (
      recordId: string,
      options: TimelineCommittedRecordIdleOptions,
    ): Promise<TimelineCommittedRecordIdleResult | null> =>
      waitForTimelineRecordEvidence(
        mutationRuntime,
        recordId,
        options.signal,
        () => {
          const row = latestCommittedTimelineRow(recordId);
          const rowVersion =
            latestCommittedRowVersion(recordId) ?? options.fallbackRowVersion;
          return typeof rowVersion === "number" ? { row, rowVersion } : null;
        },
        options.refreshIfMissing === false
          ? undefined
          : () => loadRows({ showLoading: false }),
      ),
    [
      mutationRuntime,
      latestCommittedRowVersion,
      latestCommittedTimelineRow,
      loadRows,
    ],
  );
}
