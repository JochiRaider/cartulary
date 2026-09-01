import {
  rowHistoryActionTestId,
  rowHistoryDeleteButtonTestId,
  rowHistoryDestructiveCancelButtonTestId,
  rowHistoryDestructiveConfirmButtonTestId,
  rowHistoryDestructiveConfirmPanelTestId,
  rowHistoryItemTestId,
  rowHistoryLoadingTestId,
  rowHistoryMessageTestId,
  rowHistoryOpenInspectorButtonTestId,
  rowHistoryOpenSelectedButtonTestId,
  rowHistoryPanelTestId,
  rowHistoryRestoreButtonTestId,
  rowHistoryRollbackCancelButtonTestId,
  rowHistoryRollbackConfirmButtonTestId,
  rowHistoryRollbackPreviewTestId,
  timelineInspectorSectionTestId,
} from "@cartulary/ui-contracts";
import type { CSSProperties } from "react";
import {
  WorkbookHistoryEvent,
  WorkbookHistoryList,
} from "../../inspector/presentation/WorkbookHistoryPresentation";
import { WorkbookInspectorActionButton } from "../../inspector/presentation/WorkbookInspectorActions";
import {
  WorkbookInspectorConfirmation,
  WorkbookInspectorPublicError,
  WorkbookInspectorTechnicalDetails,
} from "../../inspector/presentation/WorkbookInspectorFeedback";
import {
  workbookHistoryEventPresentation,
  workbookHistoryPendingTechnicalFields,
  workbookHistoryRollbackLabel,
} from "../../inspector/workbookHistoryPresentationModel";
import type {
  RecordHistoryItem,
  RecordHistoryRollbackAction,
  RecordHistoryState,
  RowHistoryPendingAction,
} from "../models/timelineHistoryModel";
import {
  bodyStyle,
  inlineButtonRowStyle,
  inspectorSectionStyle,
} from "./TimelineWorkbookStyles";

type TimelineHistoryPanelProps = {
  readonly currentRecordId: string | null | undefined;
  readonly history: RecordHistoryState;
  readonly pendingAction: RowHistoryPendingAction | null;
  readonly selectedActiveRowRecordId: string | null;
  readonly onCancelPendingAction: () => void;
  readonly onConfirmPendingAction: () => void;
  readonly onOpenHistory: (recordId: string) => void;
  readonly onPreviewDeleteRestore: (operation: "delete" | "restore") => void;
  readonly onPreviewRollback: (
    item: RecordHistoryItem,
    action: RecordHistoryRollbackAction,
  ) => void;
};

export function TimelineHistoryPanel({
  currentRecordId,
  history,
  pendingAction,
  selectedActiveRowRecordId,
  onCancelPendingAction,
  onConfirmPendingAction,
  onOpenHistory,
  onPreviewDeleteRestore,
  onPreviewRollback,
}: TimelineHistoryPanelProps) {
  const recordId = currentRecordId ?? null;
  const historyMatchesActiveRecord =
    recordId !== null && history.recordId === recordId;
  const historyData =
    recordId !== null && history.data?.record_id === recordId
      ? history.data
      : null;
  const historyStatus =
    recordId !== null &&
    history.status !== "idle" &&
    !historyMatchesActiveRecord
      ? "loading"
      : history.status;
  const historyError = historyMatchesActiveRecord ? history.error : null;
  const pendingHistoryAction =
    recordId !== null && pendingAction?.recordId === recordId
      ? pendingAction
      : null;

  return (
    <section
      data-testid={timelineInspectorSectionTestId("history")}
      style={inspectorSectionStyle}
      tabIndex={-1}
    >
      <div data-testid={rowHistoryPanelTestId()} style={historyPanelStyle}>
        {selectedActiveRowRecordId !== null ? (
          <WorkbookInspectorActionButton
            data-testid={rowHistoryOpenInspectorButtonTestId(
              selectedActiveRowRecordId,
            )}
            tone="secondary"
            onClick={() => onOpenHistory(selectedActiveRowRecordId)}
          >
            Refresh history
          </WorkbookInspectorActionButton>
        ) : null}
        {recordId === null ? null : (
          <WorkbookInspectorTechnicalDetails
            fields={[{ label: "Record ID", value: recordId }]}
          />
        )}
        {historyStatus === "idle" && selectedActiveRowRecordId !== null ? (
          <WorkbookInspectorActionButton
            data-testid={rowHistoryOpenSelectedButtonTestId()}
            onClick={() => onOpenHistory(selectedActiveRowRecordId)}
          >
            Open history
          </WorkbookInspectorActionButton>
        ) : null}
        {historyStatus === "loading" ? (
          <p data-testid={rowHistoryLoadingTestId()} style={bodyStyle}>
            Loading history...
          </p>
        ) : null}
        {historyError ? (
          <WorkbookInspectorPublicError
            error={historyError}
            testId={rowHistoryMessageTestId()}
          />
        ) : null}
        {historyData === null ? null : (
          <>
            <WorkbookInspectorTechnicalDetails
              fields={[
                {
                  label: "Current row version",
                  value: String(historyData.row_version),
                },
                {
                  label: "Deleted",
                  value: historyData.deleted ? "yes" : "no",
                },
              ]}
            />
            <div style={inlineButtonRowStyle}>
              {selectedActiveRowRecordId !== null && !historyData.deleted ? (
                <WorkbookInspectorActionButton
                  data-testid={rowHistoryDeleteButtonTestId()}
                  tone="destructive"
                  onClick={() => onPreviewDeleteRestore("delete")}
                >
                  Soft-delete row
                </WorkbookInspectorActionButton>
              ) : null}
              {historyData.deleted ? (
                <WorkbookInspectorActionButton
                  data-testid={rowHistoryRestoreButtonTestId()}
                  onClick={() => onPreviewDeleteRestore("restore")}
                >
                  Restore row
                </WorkbookInspectorActionButton>
              ) : null}
            </div>
            {pendingHistoryAction?.kind === "destructive" ? (
              <WorkbookInspectorConfirmation
                cancelTestId={rowHistoryDestructiveCancelButtonTestId({
                  operation: pendingHistoryAction.operation,
                })}
                confirmLabel={`Confirm ${
                  pendingHistoryAction.operation === "delete"
                    ? "soft-delete"
                    : "restore"
                }`}
                confirmTestId={rowHistoryDestructiveConfirmButtonTestId({
                  operation: pendingHistoryAction.operation,
                })}
                destructive={pendingHistoryAction.operation === "delete"}
                operation={
                  pendingHistoryAction.operation === "delete"
                    ? "Soft-delete"
                    : "Restore"
                }
                subject="this timeline row"
                technicalFields={workbookHistoryPendingTechnicalFields(
                  pendingHistoryAction,
                )}
                testId={rowHistoryDestructiveConfirmPanelTestId({
                  operation: pendingHistoryAction.operation,
                })}
                onCancel={onCancelPendingAction}
                onConfirm={onConfirmPendingAction}
              />
            ) : null}
            {pendingHistoryAction?.kind === "rollback" ? (
              <WorkbookInspectorConfirmation
                cancelTestId={rowHistoryRollbackCancelButtonTestId({
                  action: pendingHistoryAction.action,
                  historyItemRef: pendingHistoryAction.historyItemRef,
                })}
                confirmLabel="Confirm rollback"
                confirmTestId={rowHistoryRollbackConfirmButtonTestId({
                  action: pendingHistoryAction.action,
                  historyItemRef: pendingHistoryAction.historyItemRef,
                })}
                operation={`${workbookHistoryRollbackLabel(pendingHistoryAction.action)} rollback`}
                subject="this history state"
                technicalFields={[
                  ...workbookHistoryPendingTechnicalFields(
                    pendingHistoryAction,
                  ),
                  {
                    label: "History item",
                    value: pendingHistoryAction.historyItemRef,
                  },
                  {
                    label: "Target kind",
                    value: String(pendingHistoryAction.target.kind ?? ""),
                  },
                ]}
                testId={rowHistoryRollbackPreviewTestId({
                  action: pendingHistoryAction.action,
                  historyItemRef: pendingHistoryAction.historyItemRef,
                })}
                onCancel={onCancelPendingAction}
                onConfirm={onConfirmPendingAction}
              />
            ) : null}
            <WorkbookHistoryList>
              {historyData.items.map((item) => {
                const event = workbookHistoryEventPresentation(item);
                return (
                  <WorkbookHistoryEvent
                    actions={
                      item.available_rollback_actions.length === 0 ? (
                        <p style={emptyStateStyle}>No rollback action</p>
                      ) : (
                        <div style={inlineButtonRowStyle}>
                          {item.available_rollback_actions.map((action) => (
                            <WorkbookInspectorActionButton
                              data-testid={rowHistoryActionTestId({
                                action,
                                historyItemRef: item.history_item_ref,
                              })}
                              key={action}
                              tone={
                                action === "row_restore"
                                  ? "ordinary"
                                  : "secondary"
                              }
                              onClick={() => onPreviewRollback(item, action)}
                            >
                              {workbookHistoryRollbackLabel(action)}
                            </WorkbookInspectorActionButton>
                          ))}
                        </div>
                      )
                    }
                    event={event}
                    key={event.key}
                    testId={rowHistoryItemTestId({
                      historyItemRef: item.history_item_ref,
                    })}
                  />
                );
              })}
            </WorkbookHistoryList>
          </>
        )}
      </div>
    </section>
  );
}

const historyPanelStyle = {
  display: "grid",
  gap: "var(--ct-spacing-sm)",
} satisfies CSSProperties;

const emptyStateStyle = {
  margin: 0,
  color: "var(--ct-colors-ink-tertiary)",
} satisfies CSSProperties;
