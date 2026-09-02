import { assessmentCreatePanelTestId } from "@cartulary/ui-contracts";
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
import { WorkbookInspectorShell } from "../../inspector/presentation/WorkbookInspectorShell";
import { WorkbookInspectorDeclaredPanelList } from "../../inspector/WorkbookInspectorDeclaredPanelList";
import { WorkbookInspectorRecordHistory } from "../../inspector/WorkbookInspectorRecordHistory";
import type { WorkbookInspectorFeedback } from "../../inspector/workbookInspectorErrorModel";
import type { WorkbookInspectorSubject } from "../../inspector/workbookInspectorSubject";
import type { WorkbookRecordHistoryOwnerEffects } from "../../inspector/workbookRecordHistoryOwnerEffects";
import type { GenericReferenceOptions } from "../../models/workbookReferenceOptions";
import type { RecordRouteCommandPort } from "../../mutations/workbookMutationCommandPorts";

export function AssessmentWorkbookInspector({
  config,
  currentIncidentRole,
  disabledTokens,
  draftMode,
  feedback,
  feedbackTestId,
  followOn,
  history,
  onClose,
  relationshipsContent,
  related,
  relatedFeedback,
  subject,
  workflowContent,
}: {
  readonly config: ViewContract["inspectorConfig"];
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly disabledTokens: ReadonlySet<InspectorDisabledCondition>;
  readonly draftMode: "follow_on" | "standalone";
  readonly feedback: WorkbookInspectorFeedback | null;
  readonly feedbackTestId: string;
  readonly followOn: {
    readonly canCreate: boolean;
    readonly open: () => boolean;
    readonly opened: () => void;
    readonly reject: (message: string) => void;
  };
  readonly history: {
    readonly actions: ReadonlySet<"delete" | "restore" | "rollback">;
    readonly canMutate: boolean;
    readonly commands: RecordRouteCommandPort;
    readonly effects: WorkbookRecordHistoryOwnerEffects;
  };
  readonly onClose: () => void;
  readonly relationshipsContent: ReactNode;
  readonly related: {
    readonly begin: (featureGroup: InspectorFeatureGroup) => boolean;
    readonly referenceOptions: GenericReferenceOptions;
    readonly state: InspectorRelatedRecordWorkflowState | null;
    readonly cancel: () => void;
    readonly submit: () => Promise<void>;
    readonly updateDraft: (fieldKey: string, value: string) => void;
  };
  readonly relatedFeedback: WorkbookInspectorFeedback | null;
  readonly subject: WorkbookInspectorSubject | null;
  readonly workflowContent: ReactNode;
}) {
  const dispatchContextualAction = (
    capability: InspectorContextualCapability,
  ): void => {
    if (capability.kind !== "create_related") return;
    const { featureGroup } = capability;
    if (featureGroup.featureGroupKey !== "create_related.assessment") {
      related.begin(featureGroup);
      return;
    }
    if (
      featureGroup.routeBinding.kind !== "view_row_create" ||
      featureGroup.routeBinding.owner !== "view_row_create_route" ||
      featureGroup.routeBinding.targetViewSchemaId !== config.viewSchemaId
    ) {
      followOn.reject("Assessment follow-on creation is unavailable.");
      return;
    }
    if (!followOn.canCreate) {
      followOn.reject("Assessment creation requires an active editor role.");
      return;
    }
    if (followOn.open()) followOn.opened();
  };
  const panelContent = (panelId: InspectorPanelId, content?: ReactNode) => (
    <>
      {content}
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
      accessibleLabel="Compromise Assessments inspector"
      config={config}
      eyebrow="Create"
      heading={
        draftMode === "follow_on"
          ? "Append follow-on assessment"
          : "Append assessment"
      }
      mode="creation"
      noRowHeading="Append assessment"
      subject={subject}
      testId={assessmentCreatePanelTestId()}
      onClose={onClose}
    >
      <WorkbookInspectorDeclaredPanelList
        config={config}
        currentIncidentRole={currentIncidentRole}
        disabledTokens={disabledTokens}
        subject={subject}
        contentByPanel={{
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
          workflow: panelContent(
            "workflow",
            <div style={creationSectionStyle}>
              {workflowContent}
              <WorkbookInspectorFeedbackView
                feedback={feedback}
                neutralStyle={feedbackStyle}
                testId={feedbackTestId}
              />
              <WorkbookInspectorFeedbackView
                feedback={relatedFeedback}
                neutralStyle={feedbackStyle}
              />
            </div>,
          ),
        }}
        onContextualAction={dispatchContextualAction}
      />
    </WorkbookInspectorShell>
  );
}

const creationSectionStyle = {
  display: "grid",
  gap: "0.75rem",
  marginBottom: "1rem",
};

const feedbackStyle = {
  color: "var(--ct-colors-ink-muted)",
  lineHeight: 1.5,
  margin: 0,
} satisfies CSSProperties;
