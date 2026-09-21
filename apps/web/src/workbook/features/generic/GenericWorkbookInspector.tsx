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
import { inspectorPanel } from "../../inspector/presentation/WorkbookInspectorPanelContent";
import { WorkbookInspectorShell } from "../../inspector/presentation/WorkbookInspectorShell";
import type { WorkbookInspectorDisabledReason } from "../../inspector/presentation/workbookInspectorPresentationModel";
import { WorkbookInspectorDeclaredPanelList } from "../../inspector/WorkbookInspectorDeclaredPanelList";
import { WorkbookInspectorRecordHistory } from "../../inspector/WorkbookInspectorRecordHistory";
import type {
  WorkbookInspectorErrorPresentation,
  WorkbookInspectorFeedback,
} from "../../inspector/workbookInspectorErrorModel";
import type { WorkbookInspectorSubject } from "../../inspector/workbookInspectorSubject";
import type { WorkbookRecordHistoryOwnerEffects } from "../../inspector/workbookRecordHistoryOwnerEffects";
import type { RecordRouteCommandPort } from "../../mutations/workbookMutationCommandPorts";
import { IndicatorInspectorWorkflow } from "../indicators/IndicatorInspectorWorkflow";
import { IndicatorLifecycleWorkflow } from "../indicators/IndicatorLifecycleWorkflow";
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
  related,
  relatedFeedback,
  relationshipsContent,
  subject,
  surfaceTitle,
  workflowContent,
  decisionSupersession,
}: {
  readonly config: ViewContract["inspectorConfig"];
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly detailsContent: ReactNode;
  readonly disabledTokens: ReadonlySet<InspectorDisabledCondition>;
  readonly evidenceContent: ReactNode;
  readonly history: {
    readonly beginMutation: () => () => void;
    readonly actions: ReadonlySet<"delete" | "restore" | "rollback">;
    readonly canMutate: boolean;
    readonly commands: RecordRouteCommandPort;
    readonly effects: WorkbookRecordHistoryOwnerEffects;
  };
  readonly indicator: {
    readonly handler: IndicatorInspectorHandler | null;
    readonly onMutationCommitted: () => Promise<void> | void;
    readonly recordId: string;
    readonly rowVersion: number;
    readonly select: (handler: IndicatorInspectorHandler | null) => void;
  } | null;
  readonly mutationError: WorkbookInspectorErrorPresentation | null;
  readonly onClose: () => void;
  readonly related: {
    readonly begin: (featureGroup: InspectorFeatureGroup) => boolean;
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
  readonly decisionSupersession?:
    | {
        readonly start: () => void;
        readonly content: ReactNode;
        readonly disabledReason: WorkbookInspectorDisabledReason | null;
      }
    | undefined;
}) {
  function dispatchContextualAction(
    capability: InspectorContextualCapability,
  ): void {
    switch (capability.kind) {
      case "decision_supersede":
        decisionSupersession?.start();
        return;
      case "indicator":
        indicator?.select(
          resolveIndicatorInspectorHandler(
            config.viewSchemaId,
            capability.featureGroup,
          ),
        );
        return;
      case "note_create":
      case "create_related":
        indicator?.select(null);
        related.begin(capability.featureGroup);
    }
  }

  const indicatorHandlerAdmitted =
    indicator?.handler !== null &&
    subject?.viewSchemaId === config.viewSchemaId &&
    config.featureGroups.some((feature) => {
      const canonical = resolveIndicatorInspectorHandler(
        config.viewSchemaId,
        feature,
      );
      return (
        canonical !== null &&
        canonical.action === indicator?.handler?.action &&
        canonical.panelId === indicator?.handler?.panelId
      );
    });
  const panelContent = (panelId: InspectorPanelId, content?: ReactNode) =>
    inspectorPanel(
      <>
        {content}
        {panelId === "workflow" ? (
          <p>Choose an available action for this record.</p>
        ) : null}
        {indicatorHandlerAdmitted &&
        subject?.kind === "live" &&
        indicator?.handler?.panelId === panelId ? (
          indicator.handler.action === "indicator.lifecycle.read" ||
          indicator.handler.action === "indicator.lifecycle.manage" ? (
            <IndicatorLifecycleWorkflow
              action={indicator.handler.action}
              subject={subject}
            />
          ) : (
            <IndicatorInspectorWorkflow
              action={indicator.handler.action}
              indicatorRecordId={indicator.recordId}
              onMutationCommitted={indicator.onMutationCommitted}
            />
          )
        ) : null}
        {subject?.kind === "live" &&
        related.state?.featureGroup.panelId === panelId ? (
          <InspectorCreateRelatedWorkflow
            state={related.state}
            onCancel={related.cancel}
            onSubmit={() => void related.submit()}
            onUpdateDraft={related.updateDraft}
          />
        ) : null}
      </>,
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
        additionalDisabledReasons={
          decisionSupersession?.disabledReason
            ? new Map([
                ["decision.supersede", decisionSupersession.disabledReason],
              ])
            : undefined
        }
        subject={subject}
        modelsByPanel={{
          details:
            subject === null
              ? undefined
              : panelContent(
                  "details",
                  <>
                    {detailsContent}
                    {mutationError ? (
                      <WorkbookInspectorPublicError error={mutationError} />
                    ) : null}
                  </>,
                ),
          evidence:
            subject === null
              ? undefined
              : panelContent("evidence", evidenceContent),
          history:
            subject === null
              ? undefined
              : panelContent(
                  "history",
                  <>
                    {decisionSupersession?.content}
                    <WorkbookInspectorRecordHistory
                      beginMutation={history.beginMutation}
                      actions={history.actions}
                      canMutate={history.canMutate}
                      commands={history.commands}
                      ownerEffects={history.effects}
                      subject={subject}
                    />
                  </>,
                ),
          relationships:
            subject === null
              ? undefined
              : panelContent("relationships", relationshipsContent),
          workflow: panelContent(
            "workflow",
            <>
              {workflowContent}
              <WorkbookInspectorFeedbackView
                feedback={relatedFeedback}
                neutralStyle={feedbackStyle}
              />
            </>,
          ),
        }}
        onContextualAction={dispatchContextualAction}
      />
    </WorkbookInspectorShell>
  );
}

const feedbackStyle = {
  color: "var(--ct-colors-ink-muted)",
  margin: 0,
} satisfies CSSProperties;
