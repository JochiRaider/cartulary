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
import {
  inspectorPanel,
  ownedInspectorRegion,
  savedInspectorRegion,
  type WorkbookInspectorRegion,
} from "../../inspector/presentation/WorkbookInspectorPanelContent";
import { WorkbookInspectorShell } from "../../inspector/presentation/WorkbookInspectorShell";
import type { WorkbookInspectorDisabledReason } from "../../inspector/presentation/workbookInspectorPresentationModel";
import { WorkbookInspectorDeclaredPanelList } from "../../inspector/WorkbookInspectorDeclaredPanelList";
import { WorkbookInspectorRecordHistory } from "../../inspector/WorkbookInspectorRecordHistory";
import type {
  WorkbookInspectorErrorPresentation,
  WorkbookInspectorFeedback,
} from "../../inspector/workbookInspectorErrorModel";
import type { WorkbookRecordHistoryOwnerEffects } from "../../inspector/workbookRecordHistoryOwnerEffects";
import type { WorkbookRecordSubject } from "../../ports/WorkbookRecordSubject";
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
  creationAttachment,
}: {
  readonly config: ViewContract["inspectorConfig"];
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly detailsContent: ReactNode;
  readonly disabledTokens: ReadonlySet<InspectorDisabledCondition>;
  readonly evidenceContent: readonly [
    WorkbookInspectorRegion,
    ...WorkbookInspectorRegion[],
  ];
  readonly history: {
    readonly actions: ReadonlySet<"delete" | "restore" | "rollback">;
    readonly canMutate: boolean;
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
  readonly relationshipsContent: readonly [
    WorkbookInspectorRegion,
    ...WorkbookInspectorRegion[],
  ];
  readonly subject: WorkbookRecordSubject | null;
  readonly surfaceTitle: string;
  readonly workflowContent: ReactNode;
  readonly creationAttachment?: string | undefined;
  readonly decisionSupersession?:
    | {
        readonly start: () => void;
        readonly content: ReactNode;
        readonly disabledReason: WorkbookInspectorDisabledReason | null;
      }
    | undefined;
}) {
  if (currentIncidentRole === null) return null;
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
  const panelContent = (
    panelId: InspectorPanelId,
    ...regions: [WorkbookInspectorRegion, ...WorkbookInspectorRegion[]]
  ) => ({
    ...inspectorPanel(...regions),
    authoring: (
      <>
        {panelId === "workflow" ? (
          <>
            {workflowContent}
            <WorkbookInspectorFeedbackView
              feedback={relatedFeedback}
              neutralStyle={feedbackStyle}
            />
          </>
        ) : null}
        {panelId === "history" ? decisionSupersession?.content : null}
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
      </>
    ),
  });

  return (
    <WorkbookInspectorDeclaredPanelList
      creationAttachment={
        creationAttachment
          ? { id: creationAttachment, viewSchemaId: config.viewSchemaId }
          : undefined
      }
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
                savedInspectorRegion("saved-fields", {
                  kind: "populated",
                  content: (
                    <>
                      {detailsContent}
                      {mutationError ? (
                        <WorkbookInspectorPublicError error={mutationError} />
                      ) : null}
                    </>
                  ),
                }),
              ),
        evidence:
          subject === null
            ? undefined
            : panelContent("evidence", ...evidenceContent),
        history:
          subject === null
            ? undefined
            : panelContent(
                "history",
                ownedInspectorRegion("record-history", (present) => (
                  <WorkbookInspectorRecordHistory
                    present={present}
                    actions={history.actions}
                    canMutate={history.canMutate}
                    ownerEffects={history.effects}
                    subject={subject}
                  />
                )),
              ),
        relationships:
          subject === null
            ? undefined
            : panelContent("relationships", ...relationshipsContent),
        workflow:
          subject || creationAttachment
            ? panelContent(
                "workflow",
                savedInspectorRegion("workflow", {
                  kind: "empty",
                  message: "Choose an available action for this record.",
                }),
              )
            : undefined,
      }}
      onContextualAction={dispatchContextualAction}
    >
      {(sections) => (
        <WorkbookInspectorShell
          accessibleLabel={`${surfaceTitle} inspector`}
          config={config}
          {...(subject
            ? { mode: "saved", subject, sections }
            : creationAttachment
              ? {
                  mode: "creation",
                  context: {
                    id: creationAttachment,
                    viewSchemaId: config.viewSchemaId,
                    heading: `Create ${surfaceTitle}`,
                  },
                  sections,
                }
              : { mode: "empty", heading: `${surfaceTitle} inspector` })}
          onClose={onClose}
        ></WorkbookInspectorShell>
      )}
    </WorkbookInspectorDeclaredPanelList>
  );
}

const feedbackStyle = {
  color: "var(--ct-colors-ink-muted)",
  margin: 0,
} satisfies CSSProperties;
