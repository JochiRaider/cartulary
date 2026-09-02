import {
  rowHistoryLoadingTestId,
  rowHistoryMessageTestId,
  rowHistoryPanelTestId,
} from "@cartulary/ui-contracts";
import type { CSSProperties } from "react";
import type { RecordRouteCommandPort } from "../mutations/workbookMutationCommandPorts";
import type { InspectorRecordHistoryAction } from "./inspectorCapabilityResolver";
import { WorkbookInspectorActionButton } from "./presentation/WorkbookInspectorActions";
import {
  WorkbookInspectorFeedbackView,
  WorkbookInspectorPublicError,
  WorkbookInspectorTechnicalDetails,
} from "./presentation/WorkbookInspectorFeedback";
import { useWorkbookRecordHistoryController } from "./useWorkbookRecordHistoryController";
import { useWorkbookRecordHistoryFocus } from "./useWorkbookRecordHistoryFocus";
import { WorkbookRecordHistoryLoadedPresentation } from "./WorkbookRecordHistoryPresentation";
import type { WorkbookInspectorSubject } from "./workbookInspectorSubject";
import {
  type RecordHistoryItem,
  type RecordHistoryRollbackAction,
  type WorkbookRecordHistoryState,
  workbookRecordHistoryFeedback,
  workbookRecordHistoryLoadError,
  workbookRecordHistoryLoadedData,
  workbookRecordHistoryPendingAction,
} from "./workbookRecordHistoryModel";
import type { WorkbookRecordHistoryOwnerEffects } from "./workbookRecordHistoryOwnerEffects";

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
  readonly subject: WorkbookInspectorSubject | null;
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
  const data = workbookRecordHistoryLoadedData(state);
  const error = workbookRecordHistoryLoadError(state);
  const feedback = workbookRecordHistoryFeedback(state);
  const pendingAction = workbookRecordHistoryPendingAction(state);
  const focus = useWorkbookRecordHistoryFocus({
    canMutate,
    state,
    onCancelPendingAction,
    onConfirmPendingAction,
  });
  const presentedRecordId = state.subject?.recordId ?? idleRecordId ?? null;
  if (presentedRecordId === null) return null;
  const busy = state.phase === "loading" || state.phase === "submitting";
  return (
    <section
      data-testid={rowHistoryPanelTestId()}
      ref={focus.panelRef}
      style={panelStyle}
      tabIndex={-1}
    >
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
      {state.phase === "idle" ? (
        <WorkbookInspectorActionButton
          data-testid={openTestId}
          onClick={onOpenHistory}
        >
          {refreshLabel}
        </WorkbookInspectorActionButton>
      ) : null}
      {state.phase === "loading" ? (
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
      {data === null || state.subject === null ? null : (
        <>
          <WorkbookInspectorTechnicalDetails
            fields={[
              { label: "Current row version", value: String(data.row_version) },
              { label: "Deleted", value: data.deleted ? "yes" : "no" },
            ]}
          />
          <WorkbookRecordHistoryLoadedPresentation
            actions={actions}
            busy={busy}
            canMutate={canMutate}
            data={data}
            destructiveSubject={destructiveSubject}
            focus={{
              capture: focus.captureFocusRequest,
              register: focus.registerActionElement,
            }}
            pendingAction={pendingAction}
            subject={state.subject}
            onCancelPendingAction={focus.cancelPendingAction}
            onConfirmPendingAction={focus.confirmPendingAction}
            onPreviewDeleteRestore={onPreviewDeleteRestore}
            onPreviewRollback={onPreviewRollback}
          />
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
