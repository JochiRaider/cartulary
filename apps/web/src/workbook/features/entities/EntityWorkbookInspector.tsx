import type {
  InspectorDisabledCondition,
  InspectorFeatureGroup,
  ViewContract,
} from "@cartulary/view-contracts";
import type { CSSProperties, ReactNode } from "react";
import type { WorkbookIncidentRole } from "../../../shared/workbookShellContracts";
import { InspectorCreateRelatedWorkflow } from "../../inspector/InspectorCreateRelatedWorkflow";
import type { InspectorRelatedRecordWorkflowState } from "../../inspector/inspectorRelatedRecordModel";
import { WorkbookInspectorFeedbackView } from "../../inspector/presentation/WorkbookInspectorFeedback";
import {
  WorkbookInspectorPanelSection,
  WorkbookInspectorShell,
} from "../../inspector/presentation/WorkbookInspectorShell";
import type { WorkbookInspectorSubjectPresentation } from "../../inspector/presentation/workbookInspectorPresentationModel";
import {
  type InspectorContextualCapability,
  inspectorContextualCapabilities,
} from "../../inspector/semanticInspectorDispatcher";
import { WorkbookInspectorContextualActions } from "../../inspector/WorkbookInspectorContextualActions";
import { WorkbookInspectorRecordHistory } from "../../inspector/WorkbookInspectorRecordHistory";
import type { WorkbookInspectorFeedback } from "../../inspector/workbookInspectorErrorModel";
import type { WorkbookRecordHistorySubject } from "../../inspector/workbookRecordHistoryModel";
import type { GenericReferenceOptions } from "../../models/workbookReferenceOptions";
import type { RecordRouteCommandPort } from "../../mutations/workbookMutationCommandPorts";

type RecordHistoryEffects = {
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

export function EntityWorkbookInspector({
  actionFeedback,
  children,
  config,
  currentIncidentRole,
  disabledTokens,
  feedbackTestId,
  history,
  mergeFeedback,
  mergePreconditionDetails,
  onClose,
  related,
  subject,
  surfaceTitle,
  testId,
  viewSchemaId,
}: {
  readonly actionFeedback: WorkbookInspectorFeedback | null;
  readonly children: ReactNode;
  readonly config: ViewContract["inspectorConfig"];
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly disabledTokens: ReadonlySet<InspectorDisabledCondition>;
  readonly feedbackTestId: string;
  readonly history: {
    readonly actions: ReadonlySet<"delete" | "restore" | "rollback">;
    readonly canMutate: boolean;
    readonly commands: RecordRouteCommandPort;
    readonly effects: RecordHistoryEffects;
    readonly subject: WorkbookRecordHistorySubject | null;
  };
  readonly mergeFeedback: WorkbookInspectorFeedback | null;
  readonly mergePreconditionDetails: ReactNode;
  readonly onClose: () => void;
  readonly related: {
    readonly begin: (featureGroup: InspectorFeatureGroup) => boolean;
    readonly referenceOptions: GenericReferenceOptions;
    readonly state: InspectorRelatedRecordWorkflowState | null;
    readonly cancel: () => void;
    readonly submit: () => Promise<void>;
    readonly updateDraft: (fieldKey: string, value: string) => void;
  };
  readonly subject: WorkbookInspectorSubjectPresentation | null;
  readonly surfaceTitle: string;
  readonly testId?: string | undefined;
  readonly viewSchemaId: string;
}) {
  const liveSubject = history.subject?.kind === "live" ? subject : null;
  function dispatchContextualAction(
    capability: InspectorContextualCapability,
  ): void {
    if (capability.kind === "create_related") {
      related.begin(capability.featureGroup);
    }
  }

  return (
    <WorkbookInspectorShell
      accessibleLabel={`${surfaceTitle} inspector`}
      noRowHeading={`${surfaceTitle} inspector`}
      subject={subject}
      testId={testId}
      viewSchemaId={viewSchemaId}
      onClose={onClose}
    >
      {config.panels
        .filter(
          (panel) =>
            history.subject?.kind !== "deleted" || panel.panelId === "history",
        )
        .map((panel) => (
          <WorkbookInspectorPanelSection
            config={config}
            key={panel.panelId}
            panelId={panel.panelId}
          >
            {liveSubject === null ? null : (
              <WorkbookInspectorContextualActions
                config={config}
                currentIncidentRole={currentIncidentRole}
                disabledTokens={disabledTokens}
                capabilities={inspectorContextualCapabilities({
                  config,
                  panelId: panel.panelId,
                })}
                subject={liveSubject}
                onAction={dispatchContextualAction}
              />
            )}
            {related.state?.featureGroup.panelId === panel.panelId ? (
              <InspectorCreateRelatedWorkflow
                referenceOptions={related.referenceOptions}
                state={related.state}
                onCancel={related.cancel}
                onSubmit={() => void related.submit()}
                onUpdateDraft={related.updateDraft}
              />
            ) : null}
            {panel.panelId === "history" ? (
              <WorkbookInspectorRecordHistory
                actions={history.actions}
                canMutate={history.canMutate}
                commands={history.commands}
                ownerEffects={history.effects}
                subject={history.subject}
              />
            ) : null}
          </WorkbookInspectorPanelSection>
        ))}
      {history.subject?.kind === "deleted" ? null : children}
      {mergeFeedback === null ? null : (
        <div style={feedbackBlockStyle}>
          <WorkbookInspectorFeedbackView
            feedback={mergeFeedback}
            neutralStyle={feedbackStyle}
            testId={feedbackTestId}
          />
          {mergePreconditionDetails}
        </div>
      )}
      <WorkbookInspectorFeedbackView
        feedback={actionFeedback}
        neutralStyle={feedbackStyle}
        testId={feedbackTestId}
      />
    </WorkbookInspectorShell>
  );
}

const feedbackBlockStyle = {
  display: "grid",
  gap: "0.4rem",
} satisfies CSSProperties;

const feedbackStyle = {
  color: "var(--ct-colors-ink-muted)",
  lineHeight: 1.5,
  margin: 0,
} satisfies CSSProperties;
