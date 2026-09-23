import { useEffect, useRef } from "react";
import { useWorkbookHistorySurfaceRefresh } from "../../history/WorkbookHistoryContext";
import { useWorkbookRecordHistoryController } from "../../inspector/useWorkbookRecordHistoryController";
import type {
  WorkbookRecordHistoryEvent,
  WorkbookRecordHistoryState,
} from "../../inspector/workbookRecordHistoryModel";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { WorkbookRecordSubject } from "../../ports/WorkbookRecordSubject";
import type { TimelineCommittedRecordIdleResult } from "../models/timelineControllerPorts";

type TimelineHistoryLoadRowsOptions = {
  showLoading: boolean;
  requireAcceptance?: boolean;
};

export function useTimelineHistoryActions({
  acceptTimelineRecordVersion,
  activeHistorySubject,
  dispatchRowHistory,
  rowHistory,
  loadRows,
  setIsInspectorOpen,
  setSelectedRowId,
  waitForCommittedRecordIdle,
  enqueueOrderedRead,
  presentationActive = true,
}: {
  readonly presentationActive?: boolean;
  readonly acceptTimelineRecordVersion: (
    recordId: string,
    rowVersion: number,
  ) => void;
  readonly activeHistorySubject: WorkbookRecordSubject | null;
  readonly enqueueOrderedRead: (work: () => Promise<void>) => void;
  readonly loadRows: (options: TimelineHistoryLoadRowsOptions) => Promise<void>;
  readonly dispatchRowHistory: (
    event: WorkbookRecordHistoryEvent,
  ) => WorkbookRecordHistoryState;
  readonly rowHistory: WorkbookRecordHistoryState;
  readonly setIsInspectorOpen: (isOpen: boolean) => void;
  readonly setSelectedRowId: (recordId: string | null) => void;
  readonly waitForCommittedRecordIdle: (
    recordId: string,
    options?: {
      readonly fallbackRowVersion?: number | null | undefined;
      readonly refreshIfMissing?: boolean;
      readonly signal?: AbortSignal;
    },
  ) => Promise<TimelineCommittedRecordIdleResult | null>;
}) {
  const requestedRecord = useRef<string | null>(null);
  useWorkbookHistorySurfaceRefresh(timelineViewSchemaId, () =>
    loadRows({ showLoading: false, requireAcceptance: true }),
  );
  const controller = useWorkbookRecordHistoryController({
    canMutate: presentationActive,
    subject: activeHistorySubject,
    presentationActive,
    presentation: { snapshot: rowHistory, dispatch: dispatchRowHistory },
    coordinate: (recordId, signal) =>
      new Promise<number | null>((resolve) => {
        signal.addEventListener("abort", () => resolve(null), { once: true });
        enqueueOrderedRead(async () => {
          if (signal.aborted) {
            resolve(null);
            return;
          }
          try {
            const idle = await waitForCommittedRecordIdle(recordId, {
              signal,
              fallbackRowVersion: activeHistorySubject?.rowVersion,
              refreshIfMissing: activeHistorySubject?.kind !== "deleted",
            });
            resolve(signal.aborted ? null : (idle?.rowVersion ?? null));
          } catch {
            resolve(null);
          }
        });
      }),
    ownerEffects: {
      deleteAccepted: (receipt) =>
        acceptTimelineRecordVersion(receipt.recordId, receipt.rowVersion),
      restoreAccepted: (receipt) => {
        acceptTimelineRecordVersion(receipt.recordId, receipt.rowVersion);
        setSelectedRowId(receipt.recordId);
      },
      rollbackAccepted: (receipt) =>
        acceptTimelineRecordVersion(receipt.recordId, receipt.rowVersion),
      refresh: () => loadRows({ showLoading: false, requireAcceptance: true }),
    },
  });
  const open = controller.commands.open;
  useEffect(() => {
    if (
      requestedRecord.current &&
      requestedRecord.current === activeHistorySubject?.recordId
    ) {
      requestedRecord.current = null;
      open();
    }
  }, [activeHistorySubject?.recordId, open]);
  return {
    cancelRowHistoryPendingAction: controller.commands.cancel,
    confirmRowHistoryPendingAction: () => void controller.commands.confirm(),
    openRowHistory: (recordId: string) => {
      setSelectedRowId(recordId);
      setIsInspectorOpen(true);
      if (activeHistorySubject?.recordId === recordId) open();
      else requestedRecord.current = recordId;
    },
    previewRowHistoryDeleteRestore: controller.commands.previewDeleteRestore,
    previewRowHistoryRollback: controller.commands.previewRollback,
    historyBrowsingControls: controller.commands,
  };
}
