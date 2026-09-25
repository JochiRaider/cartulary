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
    ): Promise<TimelineCommittedRecordIdleResult | null> => {
      if (options.signal.aborted || mutationRuntime.retired)
        return Promise.resolve(null);
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
          options.signal.removeEventListener("abort", cancel);
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
        options.signal.addEventListener("abort", cancel, { once: true });
        if (options.signal.aborted) {
          cancel();
          return;
        }
        const read =
          async (): Promise<TimelineCommittedRecordIdleResult | null> => {
            let attemptedRefresh = false;
            for (;;) {
              const observedPublication = publication;
              const readiness = await mutationRuntime.waitForPendingRecordIdle({
                recordId,
                viewSchemaId: timelineViewSchemaId,
                signal: controller.signal,
              });
              if (readiness !== "idle" || controller.signal.aborted)
                return null;
              // A publication can occur after the coordinator resolves and
              // before this continuation runs. Recheck before accepting evidence.
              if (publication !== observedPublication) continue;
              const row = latestCommittedTimelineRow(recordId);
              const rowVersion =
                latestCommittedRowVersion(recordId) ??
                options.fallbackRowVersion;
              if (typeof rowVersion === "number") return { row, rowVersion };
              if (options.refreshIfMissing === false || attemptedRefresh)
                return null;
              attemptedRefresh = true;
              await loadRows({ showLoading: false });
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
    },
    [
      mutationRuntime,
      latestCommittedRowVersion,
      latestCommittedTimelineRow,
      loadRows,
    ],
  );
}
