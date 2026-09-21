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
import type { WorkbookInspectorAttention } from "../../inspector/presentation/workbookInspectorPresentationModel";
import { WorkbookInspectorDeclaredPanelList } from "../../inspector/WorkbookInspectorDeclaredPanelList";
import { WorkbookInspectorRecordHistory } from "../../inspector/WorkbookInspectorRecordHistory";
import type { WorkbookInspectorFeedback } from "../../inspector/workbookInspectorErrorModel";
import type { WorkbookRecordHistoryOwnerEffects } from "../../inspector/workbookRecordHistoryOwnerEffects";
import type { WorkbookRecordSubject } from "../../ports/WorkbookRecordSubject";

export function EntityWorkbookInspector({
  actionFeedback,
  config,
  currentIncidentRole,
  detailsContent,
  attention = [],
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
  readonly attention?: readonly WorkbookInspectorAttention[];
  readonly evidenceContent: ReactNode;
  readonly disabledTokens: ReadonlySet<InspectorDisabledCondition>;
  readonly feedbackTestId: string;
  readonly history: {
    readonly actions: ReadonlySet<"delete" | "restore" | "rollback">;
    readonly canMutate: boolean;
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
  readonly subject: WorkbookRecordSubject | null;
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
    attention: panelId === "details" ? attention : [],
    featureContent: Object.fromEntries(
      config.featureGroups.flatMap((feature) => {
        if (feature.panelId !== panelId || subject?.kind !== "live") return [];
        const active =
          related.state?.featureGroup.featureGroupKey ===
            feature.featureGroupKey &&
          related.state.subject.recordId === subject.recordId &&
          related.state.subject.viewSchemaId === subject.viewSchemaId;
        const notice =
          actionFeedback?.destination?.kind === "feature" &&
          actionFeedback.destination.featureGroupKey ===
            feature.featureGroupKey &&
          (!actionFeedback.sourceRecordId ||
            actionFeedback.sourceRecordId === subject.recordId)
            ? actionFeedback
            : null;
        if (!active && !notice) return [];
        return [
          [
            feature.featureGroupKey,
            <>
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
                testId={feedbackTestId}
              />
            </>,
          ],
        ];
      }),
    ),
    feedback:
      panelId === "relationships" ? (
        <>
          {relationshipFeedback}
          {actionFeedback?.destination?.kind !== "feature" ? (
            <WorkbookInspectorFeedbackView
              feedback={actionFeedback}
              neutralStyle={feedbackStyle}
              testId={feedbackTestId}
            />
          ) : null}
        </>
      ) : null,
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
          {...(subject
            ? { mode: "saved", subject, sections }
            : { mode: "empty", heading: `${surfaceTitle} inspector` })}
          testId={testId}
          onClose={onClose}
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
