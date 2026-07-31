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
import type {
  RecordHistoryItem,
  RecordHistoryRollbackAction,
  RecordHistoryState,
  RowHistoryPendingAction,
} from "../models/timelineHistoryModel";

export type TimelineHistoryPanelProps = {
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
    >
      <div data-testid={rowHistoryPanelTestId()}>
        <div style={historySectionHeaderStyle}>
          <h3 style={sectionTitleStyle}>Row history</h3>
          {selectedActiveRowRecordId !== null ? (
            <button
              data-testid={rowHistoryOpenInspectorButtonTestId(
                selectedActiveRowRecordId,
              )}
              style={secondaryActionButtonStyle}
              type="button"
              onClick={() => {
                onOpenHistory(selectedActiveRowRecordId);
              }}
            >
              Refresh history
            </button>
          ) : null}
        </div>
        {recordId !== null ? (
          <p style={historyMetaStyle}>Record {recordId}</p>
        ) : null}
        {historyStatus === "idle" && selectedActiveRowRecordId !== null ? (
          <button
            data-testid={rowHistoryOpenSelectedButtonTestId()}
            style={actionButtonStyle}
            type="button"
            onClick={() => {
              onOpenHistory(selectedActiveRowRecordId);
            }}
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
          <p
            aria-live="assertive"
            data-testid={rowHistoryMessageTestId()}
            role="alert"
            style={genericErrorTextStyle}
          >
            {historyMessage}
          </p>
        ) : null}
        {historyData !== null ? (
          <>
            <dl style={historyMetaGridStyle}>
              <div>
                <dt style={detailTermStyle}>Current row version</dt>
                <dd style={detailValueStyle}>{historyData.row_version}</dd>
              </div>
              <div>
                <dt style={detailTermStyle}>Deleted</dt>
                <dd style={detailValueStyle}>
                  {historyData.deleted ? "yes" : "no"}
                </dd>
              </div>
            </dl>
            <div style={inlineButtonRowStyle}>
              {selectedActiveRowRecordId !== null && !historyData.deleted ? (
                <button
                  data-testid={rowHistoryDeleteButtonTestId()}
                  style={destructiveActionButtonStyle}
                  type="button"
                  onClick={() => {
                    onPreviewDeleteRestore("delete");
                  }}
                >
                  Soft-delete row
                </button>
              ) : null}
              {historyData.deleted ? (
                <button
                  data-testid={rowHistoryRestoreButtonTestId()}
                  style={actionButtonStyle}
                  type="button"
                  onClick={() => {
                    onPreviewDeleteRestore("restore");
                  }}
                >
                  Restore row
                </button>
              ) : null}
            </div>
            {pendingHistoryAction?.kind === "destructive" ? (
              <div
                aria-label={`${pendingHistoryAction.operation} row confirmation`}
                aria-modal="true"
                data-testid={rowHistoryDestructiveConfirmPanelTestId({
                  operation: pendingHistoryAction.operation,
                })}
                role="alertdialog"
                style={historyConfirmPanelStyle}
              >
                <p style={bodyStyle}>
                  Confirm {pendingHistoryAction.operation} for record{" "}
                  {pendingHistoryAction.recordId} at row version{" "}
                  {pendingHistoryAction.rowVersion ?? "unknown"}.
                </p>
                <div style={inlineButtonRowStyle}>
                  <button
                    data-testid={rowHistoryDestructiveConfirmButtonTestId({
                      operation: pendingHistoryAction.operation,
                    })}
                    style={
                      pendingHistoryAction.operation === "delete"
                        ? destructiveActionButtonStyle
                        : actionButtonStyle
                    }
                    type="button"
                    onClick={onConfirmPendingAction}
                  >
                    Confirm{" "}
                    {pendingHistoryAction.operation === "delete"
                      ? "soft-delete"
                      : "restore"}
                  </button>
                  <button
                    data-testid={rowHistoryDestructiveCancelButtonTestId({
                      operation: pendingHistoryAction.operation,
                    })}
                    style={secondaryActionButtonStyle}
                    type="button"
                    onClick={onCancelPendingAction}
                  >
                    Cancel
                  </button>
                </div>
              </div>
            ) : null}
            {pendingHistoryAction?.kind === "rollback" ? (
              <div
                aria-label="Rollback preview"
                aria-modal="true"
                data-testid={rowHistoryRollbackPreviewTestId({
                  action: pendingHistoryAction.action,
                  historyItemRef: pendingHistoryAction.historyItemRef,
                })}
                role="dialog"
                style={historyConfirmPanelStyle}
              >
                <p style={bodyStyle}>
                  Preview rollback {pendingHistoryAction.action} for history
                  item {pendingHistoryAction.historyItemRef} on record{" "}
                  {pendingHistoryAction.recordId} at row version{" "}
                  {pendingHistoryAction.rowVersion ?? "unknown"}.
                </p>
                <p style={historyMetaStyle}>
                  Target {String(pendingHistoryAction.target.kind ?? "")}
                </p>
                <div style={inlineButtonRowStyle}>
                  <button
                    data-testid={rowHistoryRollbackConfirmButtonTestId({
                      action: pendingHistoryAction.action,
                      historyItemRef: pendingHistoryAction.historyItemRef,
                    })}
                    style={actionButtonStyle}
                    type="button"
                    onClick={onConfirmPendingAction}
                  >
                    Confirm rollback
                  </button>
                  <button
                    data-testid={rowHistoryRollbackCancelButtonTestId({
                      action: pendingHistoryAction.action,
                      historyItemRef: pendingHistoryAction.historyItemRef,
                    })}
                    style={secondaryActionButtonStyle}
                    type="button"
                    onClick={onCancelPendingAction}
                  >
                    Cancel
                  </button>
                </div>
              </div>
            ) : null}
            <ol style={historyListStyle}>
              {historyData.items.map((item) => {
                const itemAnchor = {
                  historyItemRef: item.history_item_ref,
                };
                const actionButtons = item.available_rollback_actions.map(
                  (action) => {
                    const label =
                      action === "history_entry"
                        ? "Rollback entry"
                        : action === "change_set"
                          ? "Rollback change set"
                          : "Restore row fields";
                    return (
                      <button
                        data-testid={rowHistoryActionTestId({
                          ...itemAnchor,
                          action,
                        })}
                        key={action}
                        style={
                          action === "row_restore"
                            ? actionButtonStyle
                            : secondaryActionButtonStyle
                        }
                        type="button"
                        onClick={() => {
                          onPreviewRollback(item, action);
                        }}
                      >
                        {label}
                      </button>
                    );
                  },
                );
                return (
                  <li
                    data-testid={rowHistoryItemTestId(itemAnchor)}
                    key={item.history_item_ref}
                    style={historyItemStyle}
                  >
                    <div style={historyItemHeaderStyle}>
                      <strong>{item.operation}</strong>
                      <time dateTime={item.committed_at}>
                        {formatHistoryTimestamp(item.committed_at)}
                      </time>
                    </div>
                    <dl style={detailListStyle}>
                      <div>
                        <dt style={detailTermStyle}>Actor</dt>
                        <dd style={detailValueStyle}>{item.actor_user_id}</dd>
                      </div>
                      <div>
                        <dt style={detailTermStyle}>Diff</dt>
                        <dd style={detailValueStyle}>
                          {item.diff_summary.summary}
                        </dd>
                      </div>
                      <div>
                        <dt style={detailTermStyle}>Change set</dt>
                        <dd style={detailValueStyle}>{item.change_set_id}</dd>
                      </div>
                    </dl>
                    {actionButtons.length > 0 ? (
                      <div style={inlineButtonRowStyle}>{actionButtons}</div>
                    ) : (
                      <p style={emptyRelationshipStyle}>No rollback action</p>
                    )}
                  </li>
                );
              })}
            </ol>
          </>
        ) : null}
      </div>
    </section>
  );
}

function formatHistoryTimestamp(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toISOString();
}

const bodyStyle = {
  margin: 0,
  lineHeight: 1.5,
  color: "var(--ct-colors-ink-muted)",
};

const actionButtonStyle = {
  borderRadius: "var(--ct-component-button-secondary-rounded)",
  border: "var(--ct-component-button-secondary-border)",
  background: "var(--ct-component-button-secondary-backgroundColor)",
  color: "var(--ct-component-button-secondary-textColor)",
  padding: "0.55rem 0.9rem",
  font: "inherit",
  cursor: "pointer",
};

const secondaryActionButtonStyle = {
  ...actionButtonStyle,
  background: "var(--ct-colors-surface-3)",
};

const destructiveActionButtonStyle = {
  ...actionButtonStyle,
  borderColor: "var(--ct-colors-semantic-destructive)",
  background: "transparent",
  color: "var(--ct-colors-semantic-destructive)",
};

const genericErrorTextStyle = {
  margin: 0,
  color: "var(--ct-colors-semantic-conflict)",
  fontWeight: 700,
};

const inspectorSectionStyle = {
  display: "grid",
  gap: "0.75rem",
  marginBottom: "1rem",
};

const historySectionHeaderStyle = {
  display: "flex",
  alignItems: "center",
  justifyContent: "space-between",
  gap: "0.75rem",
  flexWrap: "wrap" as const,
};

const historyMetaStyle = {
  margin: 0,
  color: "var(--ct-colors-ink-muted)",
  fontSize: "0.85rem",
  overflowWrap: "anywhere" as const,
};

const historyMetaGridStyle = {
  display: "grid",
  gridTemplateColumns: "repeat(auto-fit, minmax(8rem, 1fr))",
  gap: "0.75rem",
  margin: 0,
};

const historyListStyle = {
  display: "grid",
  gap: "0.75rem",
  margin: 0,
  paddingInlineStart: "1.25rem",
};

const historyItemStyle = {
  display: "grid",
  gap: "0.65rem",
  padding: "0.75rem",
  borderRadius: "0.5rem",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
};

const historyItemHeaderStyle = {
  display: "flex",
  alignItems: "baseline",
  justifyContent: "space-between",
  gap: "0.75rem",
  flexWrap: "wrap" as const,
};

const historyConfirmPanelStyle = {
  display: "grid",
  gap: "0.55rem",
  padding: "0.75rem",
  borderRadius: "0.5rem",
  border: "1px solid var(--ct-colors-semantic-caution)",
  background: "var(--ct-colors-surface-2)",
};

const sectionTitleStyle = {
  margin: 0,
  fontSize: "1rem",
};

const emptyRelationshipStyle = {
  color: "var(--ct-colors-ink-tertiary)",
  fontSize: "0.9rem",
};

const detailListStyle = {
  display: "grid",
  gap: "0.75rem",
  margin: 0,
};

const detailTermStyle = {
  fontSize: "0.75rem",
  letterSpacing: "0.08em",
  textTransform: "uppercase" as const,
  color: "var(--ct-colors-ink-muted)",
};

const detailValueStyle = {
  margin: "0.2rem 0 0",
};

const inlineButtonRowStyle = {
  display: "flex",
  gap: "0.5rem",
  flexWrap: "wrap" as const,
};
