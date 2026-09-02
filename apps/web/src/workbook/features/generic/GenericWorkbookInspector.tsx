import type {
  InspectorDisabledCondition,
  InspectorFeatureGroup,
  InspectorPanelId,
  ViewContract,
} from "@cartulary/view-contracts";
import type { CSSProperties, ReactNode } from "react";
import type { WorkbookIncidentRole } from "../../../shared/workbookShellContracts";
import { InspectorCreateRelatedWorkflow } from "../../inspector/InspectorCreateRelatedWorkflow";
import type { InspectorContextualCapability } from "../../inspector/inspectorCapabilityResolver";
import type { InspectorRelatedRecordWorkflowState } from "../../inspector/inspectorRelatedRecordModel";
import {
  WorkbookInspectorFeedbackView,
  WorkbookInspectorPublicError,
} from "../../inspector/presentation/WorkbookInspectorFeedback";
import { WorkbookInspectorShell } from "../../inspector/presentation/WorkbookInspectorShell";
import { WorkbookInspectorDeclaredPanelList } from "../../inspector/WorkbookInspectorDeclaredPanelList";
import { WorkbookInspectorRecordHistory } from "../../inspector/WorkbookInspectorRecordHistory";
import type {
  WorkbookInspectorErrorPresentation,
  WorkbookInspectorFeedback,
} from "../../inspector/workbookInspectorErrorModel";
import type { WorkbookInspectorSubject } from "../../inspector/workbookInspectorSubject";
import type { WorkbookRecordHistoryOwnerEffects } from "../../inspector/workbookRecordHistoryOwnerEffects";
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

export function GenericWorkbookInspector({
  config,
  currentIncidentRole,
  detailsContent,
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
  relationshipsContent,
  subject,
  surfaceTitle,
  workflowContent,
}: {
  readonly config: ViewContract["inspectorConfig"];
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly detailsContent: ReactNode;
  readonly disabledTokens: ReadonlySet<InspectorDisabledCondition>;
  readonly evidenceContent: ReactNode;
  readonly history: {
    readonly actions: ReadonlySet<"delete" | "restore" | "rollback">;
    readonly canMutate: boolean;
    readonly commands: RecordRouteCommandPort;
    readonly effects: WorkbookRecordHistoryOwnerEffects;
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
  readonly relationshipsContent: ReactNode;
  readonly subject: WorkbookInspectorSubject | null;
  readonly surfaceTitle: string;
  readonly workflowContent: ReactNode;
}) {
  function dispatchContextualAction(
    capability: InspectorContextualCapability,
  ): void {
    switch (capability.kind) {
      case "indicator":
        indicator?.select(
          resolveIndicatorInspectorHandler(
            config.viewSchemaId,
            capability.featureGroup,
          ),
        );
        return;
      case "create_related":
        indicator?.select(null);
        related.begin(capability.featureGroup);
    }
  }

  const panelContent = (panelId: InspectorPanelId, content?: ReactNode) => (
    <>
      {content}
      {subject?.kind === "live" && indicator?.handler?.panelId === panelId ? (
        <IndicatorInspectorWorkflow
          action={indicator.handler.action}
          indicatorRecordId={indicator.recordId}
          port={indicator.port}
          rowVersion={indicator.rowVersion}
          onMutationCommitted={indicator.onMutationCommitted}
        />
      ) : null}
      {subject?.kind === "live" &&
      related.state?.featureGroup.panelId === panelId ? (
        <InspectorCreateRelatedWorkflow
          referenceOptions={related.referenceOptions}
          state={related.state}
          onCancel={related.cancel}
          onSubmit={() => void related.submit()}
          onUpdateDraft={related.updateDraft}
        />
      ) : null}
    </>
  );

  return (
    <WorkbookInspectorShell
      accessibleLabel={`${surfaceTitle} inspector`}
      config={config}
      noRowHeading={`${surfaceTitle} inspector`}
      subject={subject}
      onClose={onClose}
    >
      <WorkbookInspectorDeclaredPanelList
        config={config}
        currentIncidentRole={currentIncidentRole}
        disabledTokens={disabledTokens}
        subject={subject}
        contentByPanel={{
          details:
            subject === null
              ? undefined
              : panelContent("details", detailsContent),
          evidence:
            subject === null
              ? undefined
              : panelContent("evidence", evidenceContent),
          history:
            subject === null ? undefined : (
              <WorkbookInspectorRecordHistory
                actions={history.actions}
                canMutate={history.canMutate}
                commands={history.commands}
                ownerEffects={history.effects}
                subject={subject}
              />
            ),
          relationships:
            subject === null
              ? undefined
              : panelContent("relationships", relationshipsContent),
          workflow: panelContent("workflow", workflowContent),
        }}
        onContextualAction={dispatchContextualAction}
      />
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
