import { assessmentCreatePanelTestId } from "@cartulary/ui-contracts";
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

export function AssessmentWorkbookInspector({
  children,
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
  viewSchemaId,
}: {
  readonly children: ReactNode;
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
    readonly effects: RecordHistoryEffects;
    readonly subject: WorkbookRecordHistorySubject | null;
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
  readonly subject: WorkbookInspectorSubjectPresentation | null;
  readonly viewSchemaId: string;
}) {
  const liveSubject = history.subject?.kind === "live" ? subject : null;
  function dispatchContextualAction(
    capability: InspectorContextualCapability,
  ): void {
    if (capability.kind !== "create_related") return;
    const { featureGroup } = capability;
    if (featureGroup.featureGroupKey !== "create_related.assessment") {
      related.begin(featureGroup);
      return;
    }
    if (
      featureGroup.routeBinding.kind !== "view_row_create" ||
      featureGroup.routeBinding.owner !== "view_row_create_route" ||
      featureGroup.routeBinding.targetViewSchemaId !== viewSchemaId
    ) {
      followOn.reject("Assessment follow-on creation is unavailable.");
      return;
    }
    if (!followOn.canCreate) {
      followOn.reject("Assessment creation requires an active editor role.");
      return;
    }
    if (followOn.open()) followOn.opened();
  }

  return (
    <WorkbookInspectorShell
      accessibleLabel="Compromise Assessments inspector"
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
            {panel.panelId === "relationships" ? relationshipsContent : null}
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
      <WorkbookInspectorFeedbackView
        feedback={relatedFeedback}
        neutralStyle={feedbackStyle}
      />
      {history.subject?.kind === "deleted" ? null : (
        <div style={creationSectionStyle}>
          {children}
          <WorkbookInspectorFeedbackView
            feedback={feedback}
            neutralStyle={feedbackStyle}
            testId={feedbackTestId}
          />
        </div>
      )}
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
