import { assessmentCreatePanelTestId } from "@cartulary/ui-contracts";
import type {
  InspectorDisabledCondition,
  InspectorFeatureGroup,
  InspectorPanelId,
  ViewContract,
} from "@cartulary/view-contracts";
import type { ReactNode } from "react";
import type { WorkbookIncidentRole } from "../../../shared/workbookShellContracts";
import {
  workbookFormFieldsStyle as creationSectionStyle,
  workbookFormMessageStyle as feedbackStyle,
  workbookFormHeadingStyle,
} from "../../components/workbookFormStyles";
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
import type { WorkbookRecordHistoryOwnerEffects } from "../../inspector/workbookRecordHistoryOwnerEffects";
import type { WorkbookRecordSubject } from "../../ports/WorkbookRecordSubject";

export function AssessmentWorkbookInspector({
  config,
  currentIncidentRole,
  detailsContent,
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
  readonly detailsContent: ReactNode;
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
    readonly effects: WorkbookRecordHistoryOwnerEffects;
  };
  readonly onClose: () => void;
  readonly relationshipsContent: ReactNode;
  readonly related: {
    readonly begin: (featureGroup: InspectorFeatureGroup) => boolean;
    readonly state: InspectorRelatedRecordWorkflowState | null;
    readonly cancel: () => void;
    readonly submit: () => Promise<void>;
    readonly updateDraft: (fieldKey: string, value: string) => void;
  };
  readonly relatedFeedback: WorkbookInspectorFeedback | null;
  readonly subject: WorkbookRecordSubject | null;
  readonly workflowContent: ReactNode;
}) {
  if (currentIncidentRole === null) return null;
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
  const assessmentAuthoring = (
    <div style={creationSectionStyle}>
      <h4 style={workbookFormHeadingStyle}>
        {draftMode === "follow_on"
          ? "Append follow-on assessment"
          : "Append assessment"}
      </h4>
      {workflowContent}
      <WorkbookInspectorFeedbackView
        feedback={feedback}
        neutralStyle={feedbackStyle}
        testId={feedbackTestId}
      />
    </div>
  );
  const panelContent = (
    panelId: InspectorPanelId,
    ...regions: [WorkbookInspectorRegion, ...WorkbookInspectorRegion[]]
  ) => ({
    ...inspectorPanel(...regions),
    featureContent: Object.fromEntries(
      config.featureGroups.flatMap((feature) => {
        if (feature.panelId !== panelId || subject?.kind !== "live") return [];
        const followOnActive =
          draftMode === "follow_on" &&
          feature.featureGroupKey === "create_related.assessment";
        const active =
          related.state?.featureGroup.featureGroupKey ===
            feature.featureGroupKey &&
          related.state.subject.recordId === subject.recordId &&
          related.state.subject.viewSchemaId === subject.viewSchemaId;
        const notice =
          relatedFeedback?.destination?.kind === "feature" &&
          relatedFeedback.destination.featureGroupKey ===
            feature.featureGroupKey &&
          (!relatedFeedback.sourceRecordId ||
            relatedFeedback.sourceRecordId === subject.recordId)
            ? relatedFeedback
            : null;
        if (!followOnActive && !active && !notice) return [];
        return [
          [
            feature.featureGroupKey,
            <>
              {followOnActive ? assessmentAuthoring : null}
              {active && related.state ? (
                <InspectorCreateRelatedWorkflow
                  state={related.state}
                  onCancel={related.cancel}
                  onSubmit={() => void related.submit()}
                  onUpdateDraft={related.updateDraft}
                />
              ) : null}
              <WorkbookInspectorFeedbackView
                feedback={notice}
                neutralStyle={feedbackStyle}
              />
            </>,
          ],
        ];
      }),
    ),
    feedback:
      panelId === "workflow" &&
      relatedFeedback?.destination?.kind !== "feature" ? (
        <WorkbookInspectorFeedbackView
          feedback={relatedFeedback}
          neutralStyle={feedbackStyle}
        />
      ) : null,
  });

  return (
    <WorkbookInspectorDeclaredPanelList
      creationAttachment={{
        id: `assessment-${draftMode}`,
        viewSchemaId: config.viewSchemaId,
      }}
      config={config}
      currentIncidentRole={currentIncidentRole}
      disabledTokens={disabledTokens}
      subject={subject}
      modelsByPanel={{
        details:
          subject === null
            ? undefined
            : inspectorPanel(
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
            : panelContent(
                "relationships",
                savedInspectorRegion("assessment-relationships", {
                  kind: "populated",
                  content: relationshipsContent,
                }),
              ),
        workflow: panelContent(
          "workflow",
          savedInspectorRegion("assessment-authoring", {
            kind: "populated",
            content:
              subject?.kind === "live" && draftMode === "follow_on"
                ? null
                : assessmentAuthoring,
          }),
        ),
      }}
      onContextualAction={dispatchContextualAction}
    >
      {(sections) => (
        <WorkbookInspectorShell
          accessibleLabel="Compromise Assessments inspector"
          config={config}
          {...(subject
            ? { mode: "saved", subject, sections }
            : {
                mode: "creation",
                sections,
                context: {
                  id: `assessment-${draftMode}`,
                  viewSchemaId: config.viewSchemaId,
                  heading:
                    draftMode === "follow_on"
                      ? "Append follow-on assessment"
                      : "Append assessment",
                },
              })}
          testId={assessmentCreatePanelTestId()}
          onClose={onClose}
        ></WorkbookInspectorShell>
      )}
    </WorkbookInspectorDeclaredPanelList>
  );
}
