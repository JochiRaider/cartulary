import {
  rowHistoryActionTestId,
  rowHistoryDeleteButtonTestId,
  rowHistoryDestructiveCancelButtonTestId,
  rowHistoryDestructiveConfirmButtonTestId,
  rowHistoryDestructiveConfirmPanelTestId,
  rowHistoryItemTestId,
  rowHistoryLoadingTestId,
  rowHistoryMessageTestId,
  rowHistoryPanelTestId,
  rowHistoryRestoreButtonTestId,
  rowHistoryRollbackCancelButtonTestId,
  rowHistoryRollbackConfirmButtonTestId,
  rowHistoryRollbackPreviewTestId,
} from "@cartulary/ui-contracts";
import { type CSSProperties, useEffect, useRef } from "react";
import type { RecordRouteCommandPort } from "../mutations/workbookMutationCommandPorts";
import {
  WorkbookHistoryEvent,
  WorkbookHistoryList,
} from "./presentation/WorkbookHistoryPresentation";
import { WorkbookInspectorActionButton } from "./presentation/WorkbookInspectorActions";
import {
  WorkbookInspectorConfirmation,
  WorkbookInspectorFeedbackView,
  WorkbookInspectorPublicError,
  WorkbookInspectorTechnicalDetails,
} from "./presentation/WorkbookInspectorFeedback";
import type { InspectorRecordHistoryAction } from "./semanticInspectorDispatcher";
import { useWorkbookRecordHistoryController } from "./useWorkbookRecordHistoryController";
import {
  workbookHistoryEventPresentation,
  workbookHistoryPendingTechnicalFields,
  workbookHistoryRollbackLabel,
} from "./workbookHistoryPresentationModel";
import type {
  RecordHistoryItem,
  RecordHistoryRollbackAction,
  WorkbookRecordHistoryState,
  WorkbookRecordHistorySubject,
} from "./workbookRecordHistoryModel";

type WorkbookRecordHistoryOwnerEffects = {
  readonly deleteAccepted: (accepted: {
    readonly recordId: string;
    readonly rowVersion: number;
  }) => Promise<void> | void;
  readonly restoreAccepted: (accepted: {
    readonly recordId: string;
    readonly rowVersion: number;
  }) => Promise<void> | void;
  readonly rollbackAccepted: (accepted: {
    readonly recordId: string;
    readonly rowVersion: number;
  }) => Promise<void> | void;
};

export function WorkbookInspectorRecordHistory({
  actions,
  canMutate,
  commands,
  ownerEffects,
  subject,
}: {
  readonly actions: ReadonlySet<InspectorRecordHistoryAction>;
  readonly canMutate: boolean;
  readonly commands: RecordRouteCommandPort;
  readonly ownerEffects: WorkbookRecordHistoryOwnerEffects;
  readonly subject: WorkbookRecordHistorySubject | null;
}) {
  const controller = useWorkbookRecordHistoryController({
    canMutate,
    commands,
    ownerEffects,
    subject,
  });
  return (
    <WorkbookRecordHistoryPanel
      actions={actions}
      canMutate={canMutate}
      idleRecordId={subject?.recordId}
      state={controller.snapshot}
      onCancelPendingAction={controller.commands.cancel}
      onConfirmPendingAction={() => void controller.commands.confirm()}
      onOpenHistory={controller.commands.open}
      onPreviewDeleteRestore={controller.commands.previewDeleteRestore}
      onPreviewRollback={controller.commands.previewRollback}
    />
  );
}

export function WorkbookRecordHistoryPanel({
  actions,
  canMutate,
  destructiveSubject = "this record",
  idleRecordId,
  openTestId,
  refreshLabel = "Open history",
  refreshControl,
  state,
  onCancelPendingAction,
  onConfirmPendingAction,
  onOpenHistory,
  onPreviewDeleteRestore,
  onPreviewRollback,
}: {
  readonly actions: ReadonlySet<InspectorRecordHistoryAction>;
  readonly canMutate: boolean;
  readonly destructiveSubject?: string;
  readonly idleRecordId?: string | undefined;
  readonly openTestId?: string;
  readonly refreshLabel?: string;
  readonly refreshControl?:
    | {
        readonly label: string;
        readonly testId: string;
        readonly onRefresh: () => void;
      }
    | undefined;
  readonly state: WorkbookRecordHistoryState;
  readonly onCancelPendingAction: () => void;
  readonly onConfirmPendingAction: () => void;
  readonly onOpenHistory: () => void;
  readonly onPreviewDeleteRestore: (operation: "delete" | "restore") => void;
  readonly onPreviewRollback: (
    item: RecordHistoryItem,
    action: RecordHistoryRollbackAction,
  ) => void;
}) {
  const { data, error, feedback, pendingAction, phase, subject } = state;
  const submittedReturnFocusTestIdRef = useRef<string | null>(null);
  useEffect(() => {
    if (pendingAction !== null) {
      submittedReturnFocusTestIdRef.current =
        pendingAction.kind === "destructive"
          ? pendingAction.operation === "delete"
            ? rowHistoryDeleteButtonTestId()
            : rowHistoryRestoreButtonTestId()
          : rowHistoryActionTestId({
              action: pendingAction.action,
              historyItemRef: pendingAction.historyItemRef,
            });
      return;
    }
    if (phase === "ready" && submittedReturnFocusTestIdRef.current !== null) {
      const testId = submittedReturnFocusTestIdRef.current;
      submittedReturnFocusTestIdRef.current = null;
      queueMicrotask(() => {
        const returnControl = Array.from(
          document.querySelectorAll<HTMLElement>("[data-testid]"),
        ).find((element) => element.dataset.testid === testId);
        returnControl?.focus({ preventScroll: true });
      });
      return;
    }
    if (phase === "idle") submittedReturnFocusTestIdRef.current = null;
  }, [pendingAction, phase]);
  const presentedRecordId = subject?.recordId ?? idleRecordId ?? null;
  if (presentedRecordId === null) return null;
  const busy = phase === "loading" || phase === "submitting";
  const pendingConfirmation =
    pendingAction?.kind === "destructive" ? (
      <WorkbookInspectorConfirmation
        cancelTestId={rowHistoryDestructiveCancelButtonTestId({
          operation: pendingAction.operation,
        })}
        confirmLabel={`Confirm ${
          pendingAction.operation === "delete" ? "soft-delete" : "restore"
        }`}
        confirmTestId={rowHistoryDestructiveConfirmButtonTestId({
          operation: pendingAction.operation,
        })}
        destructive={pendingAction.operation === "delete"}
        operation={
          pendingAction.operation === "delete" ? "Soft-delete" : "Restore"
        }
        returnFocusTestId={
          pendingAction.operation === "delete"
            ? rowHistoryDeleteButtonTestId()
            : rowHistoryRestoreButtonTestId()
        }
        subject={destructiveSubject}
        technicalFields={workbookHistoryPendingTechnicalFields(pendingAction)}
        testId={rowHistoryDestructiveConfirmPanelTestId({
          operation: pendingAction.operation,
        })}
        onCancel={onCancelPendingAction}
        onConfirm={onConfirmPendingAction}
      />
    ) : pendingAction?.kind === "rollback" ? (
      <WorkbookInspectorConfirmation
        cancelTestId={rowHistoryRollbackCancelButtonTestId({
          action: pendingAction.action,
          historyItemRef: pendingAction.historyItemRef,
        })}
        confirmLabel="Confirm rollback"
        confirmTestId={rowHistoryRollbackConfirmButtonTestId({
          action: pendingAction.action,
          historyItemRef: pendingAction.historyItemRef,
        })}
        operation={`${workbookHistoryRollbackLabel(pendingAction.action)} rollback`}
        returnFocusTestId={rowHistoryActionTestId({
          action: pendingAction.action,
          historyItemRef: pendingAction.historyItemRef,
        })}
        subject="this history state"
        technicalFields={[
          ...workbookHistoryPendingTechnicalFields(pendingAction),
          { label: "History item", value: pendingAction.historyItemRef },
          { label: "Target kind", value: pendingAction.target.kind },
        ]}
        testId={rowHistoryRollbackPreviewTestId({
          action: pendingAction.action,
          historyItemRef: pendingAction.historyItemRef,
        })}
        onCancel={onCancelPendingAction}
        onConfirm={onConfirmPendingAction}
      />
    ) : null;
  return (
    <section data-testid={rowHistoryPanelTestId()} style={panelStyle}>
      {refreshControl === undefined ? null : (
        <WorkbookInspectorActionButton
          data-testid={refreshControl.testId}
          tone="secondary"
          onClick={refreshControl.onRefresh}
        >
          {refreshControl.label}
        </WorkbookInspectorActionButton>
      )}
      <WorkbookInspectorTechnicalDetails
        fields={[{ label: "Record ID", value: presentedRecordId }]}
      />
      {phase === "idle" ? (
        <WorkbookInspectorActionButton
          data-testid={openTestId}
          onClick={onOpenHistory}
        >
          {refreshLabel}
        </WorkbookInspectorActionButton>
      ) : null}
      {phase === "loading" ? (
        <p data-testid={rowHistoryLoadingTestId()} style={metadataStyle}>
          Loading history...
        </p>
      ) : null}
      {error === null ? null : (
        <WorkbookInspectorPublicError
          error={error}
          testId={rowHistoryMessageTestId()}
        />
      )}
      <WorkbookInspectorFeedbackView
        feedback={feedback}
        neutralStyle={metadataStyle}
        testId={rowHistoryMessageTestId()}
      />
      {data === null ? null : (
        <>
          <WorkbookInspectorTechnicalDetails
            fields={[
              { label: "Current row version", value: String(data.row_version) },
              { label: "Deleted", value: data.deleted ? "yes" : "no" },
            ]}
          />
          <div style={actionsStyle}>
            {actions.has("delete") &&
            subject?.kind === "live" &&
            !data.deleted ? (
              <WorkbookInspectorActionButton
                data-testid={rowHistoryDeleteButtonTestId()}
                disabled={!canMutate || busy}
                tone="destructive"
                onClick={() => onPreviewDeleteRestore("delete")}
              >
                Soft-delete row
              </WorkbookInspectorActionButton>
            ) : null}
            {actions.has("restore") &&
            subject?.kind === "deleted" &&
            data.deleted ? (
              <WorkbookInspectorActionButton
                data-testid={rowHistoryRestoreButtonTestId()}
                disabled={!canMutate || busy}
                onClick={() => onPreviewDeleteRestore("restore")}
              >
                Restore row
              </WorkbookInspectorActionButton>
            ) : null}
          </div>
          {pendingConfirmation}
          <WorkbookHistoryList>
            {data.items.map((item) => {
              const event = workbookHistoryEventPresentation(item);
              return (
                <WorkbookHistoryEvent
                  actions={
                    !actions.has("rollback") ? null : item
                        .available_rollback_actions.length === 0 ? (
                      <p style={emptyStateStyle}>No rollback action</p>
                    ) : (
                      <div style={actionsStyle}>
                        {item.available_rollback_actions.map((action) => (
                          <WorkbookInspectorActionButton
                            data-testid={rowHistoryActionTestId({
                              action,
                              historyItemRef: item.history_item_ref,
                            })}
                            disabled={!canMutate || busy}
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
    </section>
  );
}

const panelStyle = {
  display: "grid",
  gap: "var(--ct-spacing-sm)",
} satisfies CSSProperties;

const metadataStyle = {
  color: "var(--ct-colors-ink-muted)",
  margin: 0,
  overflowWrap: "anywhere",
} satisfies CSSProperties;

const actionsStyle = {
  display: "flex",
  flexWrap: "wrap",
  gap: "0.5rem",
} satisfies CSSProperties;

const emptyStateStyle = {
  color: "var(--ct-colors-ink-tertiary)",
  margin: 0,
} satisfies CSSProperties;
