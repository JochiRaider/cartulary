import {
  rowHistoryActionTestId,
  rowHistoryDeleteButtonTestId,
  rowHistoryDestructiveCancelButtonTestId,
  rowHistoryDestructiveConfirmButtonTestId,
  rowHistoryDestructiveConfirmPanelTestId,
  rowHistoryItemTestId,
  rowHistoryRestoreButtonTestId,
  rowHistoryRollbackCancelButtonTestId,
  rowHistoryRollbackConfirmButtonTestId,
  rowHistoryRollbackPreviewTestId,
} from "@cartulary/ui-contracts";
import type { CSSProperties } from "react";
import { useHistoryActionPermission } from "../history/WorkbookHistoryContext";
import { historyOperationLabel } from "../history/workbookHistoryOperation";
import type { InspectorRecordHistoryAction } from "./inspectorCapabilityResolver";
import {
  WorkbookHistoryEvent,
  WorkbookHistoryList,
} from "./presentation/WorkbookHistoryPresentation";
import { WorkbookInspectorActionButton } from "./presentation/WorkbookInspectorActions";
import { WorkbookInspectorConfirmation } from "./presentation/WorkbookInspectorFeedback";
import {
  historyActionIdentity,
  historyRollbackActionIdentity,
} from "./useWorkbookRecordHistoryFocus";
import {
  workbookHistoryEventPresentation,
  workbookHistoryPendingTechnicalFields,
  workbookHistoryRollbackLabel,
} from "./workbookHistoryPresentationModel";
import type { WorkbookInspectorSubject } from "./workbookInspectorSubject";
import type {
  RecordHistoryData,
  RecordHistoryItem,
  RecordHistoryRollbackAction,
  WorkbookRecordHistoryPendingAction,
} from "./workbookRecordHistoryModel";

type HistoryFocusBindings = {
  readonly capture: (
    identity: string,
    kind: "delete" | "restore" | "rollback",
    element: HTMLButtonElement,
  ) => void;
  readonly register: (
    identity: string,
    element: HTMLButtonElement | null,
  ) => void;
};

export function WorkbookRecordHistoryLoadedPresentation({
  actions,
  busy,
  canMutate,
  data,
  destructiveSubject,
  focus,
  pendingAction,
  subject,
  onCancelPendingAction,
  requestedChangeSetId,
  onConfirmPendingAction,
  onPreviewDeleteRestore,
  onPreviewRollback,
}: {
  readonly actions: ReadonlySet<InspectorRecordHistoryAction>;
  readonly busy: boolean;
  readonly canMutate: boolean;
  readonly data: RecordHistoryData;
  readonly destructiveSubject: string;
  readonly focus: HistoryFocusBindings;
  readonly pendingAction: WorkbookRecordHistoryPendingAction | null;
  readonly subject: WorkbookInspectorSubject;
  readonly requestedChangeSetId?: string | undefined;
  readonly onCancelPendingAction: () => void;
  readonly onConfirmPendingAction: () => void;
  readonly onPreviewDeleteRestore: (operation: "delete" | "restore") => void;
  readonly onPreviewRollback: (
    item: RecordHistoryItem,
    action: RecordHistoryRollbackAction,
  ) => void;
}) {
  return (
    <>
      <WorkbookRecordHistoryDestructiveActions
        actions={actions}
        busy={busy}
        canMutate={canMutate}
        data={data}
        focus={focus}
        subject={subject}
        onPreviewDeleteRestore={onPreviewDeleteRestore}
      />
      <WorkbookRecordHistoryConfirmation
        destructiveSubject={destructiveSubject}
        pendingAction={pendingAction}
        onCancel={onCancelPendingAction}
        onConfirm={onConfirmPendingAction}
      />
      <WorkbookRecordHistoryEvents
        requestedChangeSetId={requestedChangeSetId}
        actions={actions}
        busy={busy}
        canMutate={canMutate}
        data={data}
        focus={focus}
        subject={subject}
        onPreviewRollback={onPreviewRollback}
      />
    </>
  );
}

function WorkbookRecordHistoryDestructiveActions({
  actions,
  busy,
  canMutate,
  data,
  focus,
  subject,
  onPreviewDeleteRestore,
}: {
  readonly actions: ReadonlySet<InspectorRecordHistoryAction>;
  readonly busy: boolean;
  readonly canMutate: boolean;
  readonly data: RecordHistoryData;
  readonly focus: HistoryFocusBindings;
  readonly subject: WorkbookInspectorSubject;
  readonly onPreviewDeleteRestore: (operation: "delete" | "restore") => void;
}) {
  const mayDelete = useHistoryActionPermission("delete");
  const mayRestore = useHistoryActionPermission("restore");
  const operation = availableDestructiveOperation(actions, subject, data);
  if (operation === null) return null;
  const identity = historyActionIdentity(subject, operation);
  return (
    <div style={actionsStyle}>
      <WorkbookInspectorActionButton
        data-testid={
          operation === "delete"
            ? rowHistoryDeleteButtonTestId()
            : rowHistoryRestoreButtonTestId()
        }
        disabled={
          !canMutate ||
          busy ||
          !(operation === "delete" ? mayDelete : mayRestore)
        }
        ref={(element) => focus.register(identity, element)}
        tone={operation === "delete" ? "destructive" : "ordinary"}
        onClick={(event) => {
          focus.capture(identity, operation, event.currentTarget);
          onPreviewDeleteRestore(operation);
        }}
      >
        {operation === "delete" ? "Soft-delete row" : "Restore deleted row"}
      </WorkbookInspectorActionButton>
    </div>
  );
}

function WorkbookRecordHistoryConfirmation({
  destructiveSubject,
  pendingAction,
  onCancel,
  onConfirm,
}: {
  readonly destructiveSubject: string;
  readonly pendingAction: WorkbookRecordHistoryPendingAction | null;
  readonly onCancel: () => void;
  readonly onConfirm: () => void;
}) {
  if (pendingAction === null) return null;
  if (pendingAction.kind === "destructive") {
    const operationLabel =
      pendingAction.operation === "delete" ? "Soft-delete" : "Restore";
    return (
      <WorkbookInspectorConfirmation
        cancelTestId={rowHistoryDestructiveCancelButtonTestId({
          operation: pendingAction.operation,
        })}
        confirmLabel={`Confirm ${operationLabel.toLowerCase()}`}
        confirmTestId={rowHistoryDestructiveConfirmButtonTestId({
          operation: pendingAction.operation,
        })}
        destructive={pendingAction.operation === "delete"}
        operation={operationLabel}
        subject={destructiveSubject}
        technicalFields={workbookHistoryPendingTechnicalFields(pendingAction)}
        testId={rowHistoryDestructiveConfirmPanelTestId({
          operation: pendingAction.operation,
        })}
        onCancel={onCancel}
        onConfirm={onConfirm}
      />
    );
  }
  return (
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
      operation={historyOperationLabel({ pending: pendingAction })}
      subject={
        pendingAction.target.kind === "row_restore"
          ? "for this row; independent relationships and evidence remain unchanged"
          : pendingAction.target.kind === "change_set"
            ? "for all reversible changes in this change set, including affected records beyond the displayed row or returned record list"
            : "for this history entry"
      }
      technicalFields={[
        ...workbookHistoryPendingTechnicalFields(pendingAction),
        { label: "History item", value: pendingAction.historyItemRef },
        { label: "Target kind", value: pendingAction.target.kind },
      ]}
      testId={rowHistoryRollbackPreviewTestId({
        action: pendingAction.action,
        historyItemRef: pendingAction.historyItemRef,
      })}
      onCancel={onCancel}
      onConfirm={onConfirm}
    />
  );
}

function WorkbookRecordHistoryEvents({
  requestedChangeSetId,
  actions,
  busy,
  canMutate,
  data,
  focus,
  subject,
  onPreviewRollback,
}: {
  readonly requestedChangeSetId?: string | undefined;
  readonly actions: ReadonlySet<InspectorRecordHistoryAction>;
  readonly busy: boolean;
  readonly canMutate: boolean;
  readonly data: RecordHistoryData;
  readonly focus: HistoryFocusBindings;
  readonly subject: WorkbookInspectorSubject;
  readonly onPreviewRollback: (
    item: RecordHistoryItem,
    action: RecordHistoryRollbackAction,
  ) => void;
}) {
  return (
    <WorkbookHistoryList>
      {data.items.map((item) => {
        const event = workbookHistoryEventPresentation(item);
        return (
          <WorkbookHistoryEvent
            highlighted={item.change_set_id === requestedChangeSetId}
            actions={
              actions.has("rollback") ? (
                <WorkbookRecordHistoryRollbackActions
                  busy={busy}
                  canMutate={canMutate}
                  focus={focus}
                  item={item}
                  subject={subject}
                  onPreviewRollback={onPreviewRollback}
                />
              ) : null
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
  );
}

function WorkbookRecordHistoryRollbackActions({
  busy,
  canMutate,
  focus,
  item,
  subject,
  onPreviewRollback,
}: {
  readonly busy: boolean;
  readonly canMutate: boolean;
  readonly focus: HistoryFocusBindings;
  readonly item: RecordHistoryItem;
  readonly subject: WorkbookInspectorSubject;
  readonly onPreviewRollback: (
    item: RecordHistoryItem,
    action: RecordHistoryRollbackAction,
  ) => void;
}) {
  const permitted = useHistoryActionPermission("rollback");
  if (item.available_rollback_actions.length === 0) {
    return <p style={emptyStateStyle}>No rollback action</p>;
  }
  return (
    <div style={actionsStyle}>
      {item.available_rollback_actions.map((action) => {
        const identity = historyRollbackActionIdentity(
          subject,
          item.history_item_ref,
          action,
        );
        return (
          <WorkbookInspectorActionButton
            data-testid={rowHistoryActionTestId({
              action,
              historyItemRef: item.history_item_ref,
            })}
            disabled={!canMutate || busy || !permitted}
            key={action}
            ref={(element) => focus.register(identity, element)}
            tone={action === "row_restore" ? "ordinary" : "secondary"}
            onClick={(event) => {
              focus.capture(identity, "rollback", event.currentTarget);
              onPreviewRollback(item, action);
            }}
          >
            {workbookHistoryRollbackLabel(action)}
          </WorkbookInspectorActionButton>
        );
      })}
    </div>
  );
}

function availableDestructiveOperation(
  actions: ReadonlySet<InspectorRecordHistoryAction>,
  subject: WorkbookInspectorSubject,
  data: RecordHistoryData,
): "delete" | "restore" | null {
  if (actions.has("delete") && subject.kind === "live" && !data.deleted) {
    return "delete";
  }
  if (actions.has("restore") && subject.kind === "deleted" && data.deleted) {
    return "restore";
  }
  return null;
}

const actionsStyle = {
  display: "flex",
  flexWrap: "wrap",
  gap: "var(--ct-spacing-sm)",
} satisfies CSSProperties;

const emptyStateStyle = {
  color: "var(--ct-colors-ink-tertiary)",
  margin: 0,
} satisfies CSSProperties;
