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
import { WorkbookInspectorFeedbackView } from "../../inspector/presentation/WorkbookInspectorFeedback";
import {
  inspectorPanel,
  ownedInspectorRegion,
  savedInspectorRegion,
  type WorkbookInspectorRegion,
} from "../../inspector/presentation/WorkbookInspectorPanelContent";
import { WorkbookInspectorShell } from "../../inspector/presentation/WorkbookInspectorShell";
import { WorkbookInspectorDeclaredPanelList } from "../../inspector/WorkbookInspectorDeclaredPanelList";
import { WorkbookInspectorRecordHistory } from "../../inspector/WorkbookInspectorRecordHistory";
import type { WorkbookInspectorFeedback } from "../../inspector/workbookInspectorErrorModel";
import type { WorkbookInspectorSubject } from "../../inspector/workbookInspectorSubject";
import type { WorkbookRecordHistoryOwnerEffects } from "../../inspector/workbookRecordHistoryOwnerEffects";
import type { RecordRouteCommandPort } from "../../mutations/workbookMutationCommandPorts";

export function EntityWorkbookInspector({
  actionFeedback,
  config,
  currentIncidentRole,
  detailsContent,
  evidenceContent,
  disabledTokens,
  feedbackTestId,
  history,
  mergeFeedback,
  mergePreconditionDetails,
  onClose,
  related,
  relationshipsContent,
  subject,
  surfaceTitle,
  testId,
}: {
  readonly actionFeedback: WorkbookInspectorFeedback | null;
  readonly config: ViewContract["inspectorConfig"];
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly detailsContent: ReactNode;
  readonly evidenceContent: ReactNode;
  readonly disabledTokens: ReadonlySet<InspectorDisabledCondition>;
  readonly feedbackTestId: string;
  readonly history: {
    readonly beginMutation: () => () => void;
    readonly actions: ReadonlySet<"delete" | "restore" | "rollback">;
    readonly canMutate: boolean;
    readonly commands: RecordRouteCommandPort;
    readonly effects: WorkbookRecordHistoryOwnerEffects;
  };
  readonly mergeFeedback: WorkbookInspectorFeedback | null;
  readonly mergePreconditionDetails: ReactNode;
  readonly onClose: () => void;
  readonly related: {
    readonly begin: (featureGroup: InspectorFeatureGroup) => boolean;
    readonly state: InspectorRelatedRecordWorkflowState | null;
    readonly cancel: () => void;
    readonly submit: () => Promise<void>;
    readonly updateDraft: (fieldKey: string, value: string) => void;
  };
  readonly relationshipsContent: readonly [
    WorkbookInspectorRegion,
    ...WorkbookInspectorRegion[],
  ];
  readonly subject: WorkbookInspectorSubject | null;
  readonly surfaceTitle: string;
  readonly testId?: string | undefined;
}) {
  if (currentIncidentRole === null) return null;
  const dispatchContextualAction = (
    capability: InspectorContextualCapability,
  ) => {
    if (
      capability.kind === "create_related" ||
      capability.kind === "note_create"
    ) {
      related.begin(capability.featureGroup);
    }
  };
  const panelContent = (
    panelId: InspectorPanelId,
    ...regions: [WorkbookInspectorRegion, ...WorkbookInspectorRegion[]]
  ) => ({
    ...inspectorPanel(...regions),
    authoring: (
      <>
        {panelId === "relationships" ? (
          <>
            {relationshipFeedback}
            <WorkbookInspectorFeedbackView
              feedback={actionFeedback}
              neutralStyle={feedbackStyle}
              testId={feedbackTestId}
            />
          </>
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
  const relationshipFeedback =
    mergeFeedback === null ? null : (
      <div style={feedbackBlockStyle}>
        <WorkbookInspectorFeedbackView
          feedback={mergeFeedback}
          neutralStyle={feedbackStyle}
          testId={feedbackTestId}
        />
        {mergePreconditionDetails}
      </div>
    );

  return (
    <WorkbookInspectorDeclaredPanelList
      config={config}
      currentIncidentRole={currentIncidentRole}
      disabledTokens={disabledTokens}
      subject={subject}
      modelsByPanel={{
        evidence:
          subject === null
            ? undefined
            : inspectorPanel(
                savedInspectorRegion("evidence-metadata", {
                  kind: "populated",
                  content: evidenceContent,
                }),
              ),
        details:
          subject === null
            ? undefined
            : panelContent(
                "details",
                savedInspectorRegion("saved-fields", {
                  kind: "populated",
                  content: detailsContent,
                }),
              ),
        history:
          subject === null
            ? undefined
            : inspectorPanel(
                ownedInspectorRegion("record-history", (present) => (
                  <WorkbookInspectorRecordHistory
                    present={present}
                    beginMutation={history.beginMutation}
                    actions={history.actions}
                    canMutate={history.canMutate}
                    commands={history.commands}
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
          subject === null
            ? undefined
            : panelContent(
                "workflow",
                savedInspectorRegion("workflow", {
                  kind: "empty",
                  message: "Choose an available action for this record.",
                }),
              ),
      }}
      onContextualAction={dispatchContextualAction}
    >
      {(sections) => (
        <WorkbookInspectorShell
          accessibleLabel={`${surfaceTitle} inspector`}
          config={config}
          noRowHeading={`${surfaceTitle} inspector`}
          subject={subject}
          testId={testId}
          onClose={onClose}
          sections={sections}
        ></WorkbookInspectorShell>
      )}
    </WorkbookInspectorDeclaredPanelList>
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
