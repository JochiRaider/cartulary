import {
  rowHistoryOpenInspectorButtonTestId,
  rowHistoryOpenSelectedButtonTestId,
  timelineInspectorSectionTestId,
} from "@cartulary/ui-contracts";
import { requireViewContract } from "@cartulary/view-contracts";
import { inspectorRecordHistoryActions } from "../../inspector/inspectorCapabilityResolver";
import { WorkbookRecordHistoryPanel } from "../../inspector/WorkbookInspectorRecordHistory";
import type {
  RecordHistoryItem,
  RecordHistoryRollbackAction,
  WorkbookRecordHistoryState,
} from "../../inspector/workbookRecordHistoryModel";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { inspectorSectionStyle } from "./TimelineWorkbookStyles";

const timelineHistoryActions = inspectorRecordHistoryActions(
  requireViewContract(timelineViewSchemaId).inspectorConfig,
);

export function TimelineHistoryPanel({
  canMutate,
  history,
  selectedActiveRowRecordId,
  onCancelPendingAction,
  onConfirmPendingAction,
  onOpenHistory,
  onPreviewDeleteRestore,
  onPreviewRollback,
}: {
  readonly canMutate: boolean;
  readonly history: WorkbookRecordHistoryState;
  readonly selectedActiveRowRecordId: string | null;
  readonly onCancelPendingAction: () => void;
  readonly onConfirmPendingAction: () => void;
  readonly onOpenHistory: (recordId: string) => void;
  readonly onPreviewDeleteRestore: (operation: "delete" | "restore") => void;
  readonly onPreviewRollback: (
    item: RecordHistoryItem,
    action: RecordHistoryRollbackAction,
  ) => void;
}) {
  return (
    <section
      data-testid={timelineInspectorSectionTestId("history")}
      style={inspectorSectionStyle}
      tabIndex={-1}
    >
      <WorkbookRecordHistoryPanel
        actions={timelineHistoryActions}
        canMutate={canMutate}
        destructiveSubject="this timeline row"
        idleRecordId={selectedActiveRowRecordId ?? undefined}
        openTestId={rowHistoryOpenSelectedButtonTestId()}
        refreshControl={
          selectedActiveRowRecordId === null
            ? undefined
            : {
                label: "Refresh history",
                onRefresh: () => onOpenHistory(selectedActiveRowRecordId),
                testId: rowHistoryOpenInspectorButtonTestId(
                  selectedActiveRowRecordId,
                ),
              }
        }
        state={history}
        onCancelPendingAction={onCancelPendingAction}
        onConfirmPendingAction={onConfirmPendingAction}
        onOpenHistory={() => {
          const recordId =
            history.subject?.recordId ?? selectedActiveRowRecordId ?? undefined;
          if (recordId !== undefined) onOpenHistory(recordId);
        }}
        onPreviewDeleteRestore={onPreviewDeleteRestore}
        onPreviewRollback={onPreviewRollback}
      />
    </section>
  );
}
