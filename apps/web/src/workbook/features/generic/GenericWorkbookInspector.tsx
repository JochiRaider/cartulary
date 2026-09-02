import type {
  InspectorDisabledCondition,
  InspectorFeatureGroup,
  ViewContract,
} from "@cartulary/view-contracts";
import type { CSSProperties, ReactNode } from "react";
import type { WorkbookIncidentRole } from "../../../shared/workbookShellContracts";
import { InspectorCreateRelatedWorkflow } from "../../inspector/InspectorCreateRelatedWorkflow";
import type { InspectorRelatedRecordWorkflowState } from "../../inspector/inspectorRelatedRecordModel";
import {
  WorkbookInspectorFeedbackView,
  WorkbookInspectorPublicError,
} from "../../inspector/presentation/WorkbookInspectorFeedback";
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
import type {
  WorkbookInspectorErrorPresentation,
  WorkbookInspectorFeedback,
} from "../../inspector/workbookInspectorErrorModel";
import type { WorkbookRecordHistorySubject } from "../../inspector/workbookRecordHistoryModel";
import type { GenericReferenceOptions } from "../../models/workbookReferenceOptions";
import type {
  IndicatorWorkflowPort,
  RecordRouteCommandPort,
} from "../../mutations/workbookMutationCommandPorts";
import { IndicatorInspectorWorkflow } from "../indicators/IndicatorInspectorWorkflow";
import {
  type IndicatorInspectorHandler,
  resolveIndicatorInspectorHandler,
} from "../indicators/indicatorInspectorHandlers";

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

export function GenericWorkbookInspector({
  children,
  config,
  currentIncidentRole,
  disabledTokens,
  evidenceContent,
  history,
  indicator,
  mutationError,
  onClose,
  referenceLoadError,
  referenceLoadErrorTestId,
  related,
  relatedFeedback,
  subject,
  surfaceTitle,
  viewSchemaId,
}: {
  readonly children: ReactNode;
  readonly config: ViewContract["inspectorConfig"];
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly disabledTokens: ReadonlySet<InspectorDisabledCondition>;
  readonly evidenceContent: ReactNode;
  readonly history: {
    readonly actions: ReadonlySet<"delete" | "restore" | "rollback">;
    readonly canMutate: boolean;
    readonly commands: RecordRouteCommandPort;
    readonly effects: RecordHistoryEffects;
    readonly subject: WorkbookRecordHistorySubject | null;
  };
  readonly indicator: {
    readonly handler: IndicatorInspectorHandler | null;
    readonly onMutationCommitted: () => Promise<void> | void;
    readonly port: IndicatorWorkflowPort;
    readonly recordId: string;
    readonly rowVersion: number;
    readonly select: (handler: IndicatorInspectorHandler | null) => void;
  } | null;
  readonly mutationError: WorkbookInspectorErrorPresentation | null;
  readonly onClose: () => void;
  readonly referenceLoadError: WorkbookInspectorErrorPresentation | null;
  readonly referenceLoadErrorTestId?: string | undefined;
  readonly related: {
    readonly begin: (featureGroup: InspectorFeatureGroup) => boolean;
    readonly referenceOptions: GenericReferenceOptions;
    readonly state: InspectorRelatedRecordWorkflowState | null;
    readonly cancel: () => void;
    readonly submit: () => Promise<void>;
    readonly updateDraft: (fieldKey: string, value: string) => void;
  };
  readonly relatedFeedback: WorkbookInspectorFeedback | null;
  readonly subject: WorkbookInspectorSubjectPresentation | null;
  readonly surfaceTitle: string;
  readonly viewSchemaId: string;
}) {
  const liveSubject = history.subject?.kind === "live" ? subject : null;
  function dispatchContextualAction(
    capability: InspectorContextualCapability,
  ): void {
    switch (capability.kind) {
      case "indicator":
        indicator?.select(
          resolveIndicatorInspectorHandler(
            viewSchemaId,
            capability.featureGroup,
          ),
        );
        return;
      case "create_related":
        indicator?.select(null);
        related.begin(capability.featureGroup);
        return;
      case "record_history":
        indicator?.select(null);
        return;
    }
  }

  return (
    <WorkbookInspectorShell
      accessibleLabel={`${surfaceTitle} inspector`}
      noRowHeading={`${surfaceTitle} inspector`}
      subject={subject}
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
            {indicator?.handler?.panelId === panel.panelId ? (
              <IndicatorInspectorWorkflow
                action={indicator.handler.action}
                indicatorRecordId={indicator.recordId}
                port={indicator.port}
                rowVersion={indicator.rowVersion}
                onMutationCommitted={indicator.onMutationCommitted}
              />
            ) : null}
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
            {panel.panelId === "evidence" ? evidenceContent : null}
          </WorkbookInspectorPanelSection>
        ))}
      {history.subject?.kind === "deleted" ? null : children}
      <WorkbookInspectorFeedbackView
        feedback={relatedFeedback}
        neutralStyle={feedbackStyle}
      />
      {referenceLoadError === null ? null : (
        <WorkbookInspectorPublicError
          error={referenceLoadError}
          testId={referenceLoadErrorTestId}
        />
      )}
      {mutationError === null ? null : (
        <WorkbookInspectorPublicError error={mutationError} />
      )}
    </WorkbookInspectorShell>
  );
}

const feedbackStyle = {
  color: "var(--ct-colors-ink-muted)",
  margin: 0,
} satisfies CSSProperties;
