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
import {
  WorkbookInspectorConfirmation,
  WorkbookInspectorPublicError,
  WorkbookInspectorTechnicalDetails,
} from "../../inspector/presentation/WorkbookInspectorPresentation";
import type { WorkbookHistoryEventPresentation } from "../../inspector/presentation/workbookInspectorPresentationModel";
import type {
  RecordHistoryItem,
  RecordHistoryRollbackAction,
  RecordHistoryState,
  RowHistoryPendingAction,
} from "../models/timelineHistoryModel";
import {
  actionButtonStyle,
  bodyStyle,
  inlineButtonRowStyle,
  inspectorSectionStyle,
  secondaryActionButtonStyle,
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
  const historyMessage = historyMatchesActiveRecord ? history.message : null;
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
          <button
            data-testid={rowHistoryOpenInspectorButtonTestId(
              selectedActiveRowRecordId,
            )}
            style={secondaryActionButtonStyle}
            type="button"
            onClick={() => onOpenHistory(selectedActiveRowRecordId)}
          >
            Refresh history
          </button>
        ) : null}
        {recordId === null ? null : (
          <WorkbookInspectorTechnicalDetails
            fields={[{ label: "Record ID", value: recordId }]}
          />
        )}
        {historyStatus === "idle" && selectedActiveRowRecordId !== null ? (
          <button
            data-testid={rowHistoryOpenSelectedButtonTestId()}
            style={actionButtonStyle}
            type="button"
            onClick={() => onOpenHistory(selectedActiveRowRecordId)}
          >
            Open history
          </button>
        ) : null}
        {historyStatus === "loading" ? (
          <p data-testid={rowHistoryLoadingTestId()} style={bodyStyle}>
            Loading history...
          </p>
        ) : null}
        {historyMessage ? (
          <WorkbookInspectorPublicError
            message={historyMessage}
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
                <button
                  data-testid={rowHistoryDeleteButtonTestId()}
                  style={destructiveActionButtonStyle}
                  type="button"
                  onClick={() => onPreviewDeleteRestore("delete")}
                >
                  Soft-delete row
                </button>
              ) : null}
              {historyData.deleted ? (
                <button
                  data-testid={rowHistoryRestoreButtonTestId()}
                  style={actionButtonStyle}
                  type="button"
                  onClick={() => onPreviewDeleteRestore("restore")}
                >
                  Restore row
                </button>
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
                technicalFields={pendingTechnicalFields(pendingHistoryAction)}
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
                operation={`${rollbackLabel(pendingHistoryAction.action)} rollback`}
                subject="this history state"
                technicalFields={[
                  ...pendingTechnicalFields(pendingHistoryAction),
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
                const event = historyEventPresentation(item);
                return (
                  <WorkbookHistoryEvent
                    actions={
                      item.available_rollback_actions.length === 0 ? (
                        <p style={emptyStateStyle}>No rollback action</p>
                      ) : (
                        <div style={inlineButtonRowStyle}>
                          {item.available_rollback_actions.map((action) => (
                            <button
                              data-testid={rowHistoryActionTestId({
                                action,
                                historyItemRef: item.history_item_ref,
                              })}
                              key={action}
                              style={
                                action === "row_restore"
                                  ? actionButtonStyle
                                  : secondaryActionButtonStyle
                              }
                              type="button"
                              onClick={() => onPreviewRollback(item, action)}
                            >
                              {rollbackLabel(action)}
                            </button>
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

function pendingTechnicalFields(action: RowHistoryPendingAction) {
  return [
    { label: "Record ID", value: action.recordId },
    {
      label: "Row version",
      value: action.rowVersion === null ? "unknown" : String(action.rowVersion),
    },
  ];
}

function historyEventPresentation(
  item: RecordHistoryItem,
): WorkbookHistoryEventPresentation {
  return {
    committedAt: item.committed_at,
    key: item.history_item_ref,
    operation: item.operation,
    summary: item.diff_summary.summary,
    technicalFields: [
      { label: "Actor ID", value: item.actor_user_id },
      { label: "History reference", value: item.history_item_ref },
      { label: "Change set ID", value: item.change_set_id },
      ...(item.history_entry_ref === undefined
        ? []
        : [{ label: "History entry", value: item.history_entry_ref }]),
      ...(item.revision_no === undefined
        ? []
        : [{ label: "Revision", value: String(item.revision_no) }]),
    ],
  };
}

function rollbackLabel(action: RecordHistoryRollbackAction): string {
  if (action === "history_entry") return "Rollback entry";
  if (action === "change_set") return "Rollback change set";
  return "Restore row fields";
}

const historyPanelStyle = {
  display: "grid",
  gap: "var(--ct-spacing-sm)",
} satisfies CSSProperties;

const destructiveActionButtonStyle = {
  ...actionButtonStyle,
  borderColor: "var(--ct-colors-semantic-destructive)",
  background: "transparent",
  color: "var(--ct-colors-semantic-destructive)",
} satisfies CSSProperties;

const emptyStateStyle = {
  margin: 0,
  color: "var(--ct-colors-ink-tertiary)",
} satisfies CSSProperties;
