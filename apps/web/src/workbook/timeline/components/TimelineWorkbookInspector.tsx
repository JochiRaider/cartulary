import {
  timelineInspectorMessageTestId,
  timelineInspectorTestId,
} from "@cartulary/ui-contracts";
import type {
  InspectorConfig,
  InspectorDisabledCondition,
} from "@cartulary/view-contracts";
import type { ReactNode } from "react";
import type { WorkbookIncidentRole } from "../../../shared/workbookShellContracts";
import type { MentionResolutionAction } from "../../collaboration/workbookCollaborationMessages";
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
import type { WorkbookInspectorFeedback } from "../../inspector/workbookInspectorErrorModel";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { InspectorMention } from "../models/workbookMentionChips";
import type { WorkbookRow } from "../models/workbookTimelineModel";
import {
  type MentionEntityOption,
  TimelineMentionsPanel,
} from "./TimelineMentionsPanel";
import { bodyStyle } from "./TimelineWorkbookStyles";

export function TimelineWorkbookInspector({
  canManageMentions,
  currentHistoryDeleted,
  currentIncidentRole,
  incidentClosed,
  entityIndex,
  getRelationshipLabel,
  hostEntities,
  identityEntities,
  inspectorConfig,
  inspectorMessage,
  inspectorMentions,
  onResolveTargetChange,
  onSelectMention,
  onSetInspectorMessage,
  onClose,
  onFeatureAction,
  onCreateEntityFromMention,
  onSubmitMentionAction,
  renderEvidenceAttachSection,
  renderInspectorFieldEditors,
  renderPanelSupplement,
  renderRelationshipEditors,
  renderWorkflowSection,
  renderRowHistorySection,
  rowHistoryRecordId,
  rowHistoryRowVersion,
  selectedMention,
  selectedResolveTargetId,
  selectedRow,
}: {
  readonly canManageMentions: boolean;
  readonly currentHistoryDeleted: boolean;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly incidentClosed: boolean;
  readonly entityIndex: Record<string, { label: string }>;
  readonly getRelationshipLabel: (
    fieldKey: InspectorMention["fieldKey"],
  ) => string;
  readonly hostEntities: readonly MentionEntityOption[];
  readonly identityEntities: readonly MentionEntityOption[];
  readonly inspectorConfig: InspectorConfig;
  readonly inspectorMessage: WorkbookInspectorFeedback | null;
  readonly inspectorMentions: readonly InspectorMention[];
  readonly onResolveTargetChange: (value: string) => void;
  readonly onSelectMention: (rowRecordId: string, itemRef: string) => void;
  readonly onSetInspectorMessage: (message: WorkbookInspectorFeedback) => void;
  readonly onClose: () => void;
  readonly onFeatureAction: (capability: InspectorContextualCapability) => void;
  readonly onCreateEntityFromMention: (mention: InspectorMention) => void;
  readonly onSubmitMentionAction: (
    mention: InspectorMention,
    action: MentionResolutionAction,
    resolvedRecordId?: string,
  ) => void;
  readonly renderEvidenceAttachSection: (row: WorkbookRow) => ReactNode;
  readonly renderInspectorFieldEditors: (row: WorkbookRow) => ReactNode;
  readonly renderPanelSupplement: (
    panelId: InspectorConfig["panels"][number]["panelId"],
  ) => ReactNode;
  readonly renderRelationshipEditors: (row: WorkbookRow) => ReactNode;
  readonly renderWorkflowSection: () => ReactNode;
  readonly renderRowHistorySection: () => ReactNode;
  readonly rowHistoryRecordId: string | null;
  readonly rowHistoryRowVersion: number | null;
  readonly selectedMention: InspectorMention | null;
  readonly selectedResolveTargetId: string;
  readonly selectedRow: WorkbookRow | null;
}) {
  const disabledTokens = new Set<InspectorDisabledCondition>();
  if (!selectedRow?.recordId && !currentHistoryDeleted) {
    disabledTokens.add("no_row_selected");
  }
  if (currentHistoryDeleted) {
    disabledTokens.add("record_deleted");
  } else if (selectedRow?.recordId) {
    disabledTokens.add("record_not_deleted");
  }
  disabledTokens.add("rollback_target_unavailable");
  if (incidentClosed) {
    disabledTokens.add("incident_closed");
  }
  const subjectPresentation: WorkbookInspectorSubjectPresentation | null =
    selectedRow?.recordId
      ? {
          label:
            selectedRow.values.activitySynopsisText.trim() ||
            "Selected timeline row",
          recordId: selectedRow.recordId,
          rowVersion: selectedRow.rowVersion,
          surfaceLabel: "Timeline",
        }
      : currentHistoryDeleted &&
          rowHistoryRecordId !== null &&
          rowHistoryRowVersion !== null
        ? {
            label: "Deleted timeline row",
            recordId: rowHistoryRecordId,
            rowVersion: rowHistoryRowVersion,
            stateLabel: "Deleted",
            surfaceLabel: "Timeline",
          }
        : null;
  const renderPanel = (
    panelId: (typeof inspectorConfig.panels)[number]["panelId"],
  ) => {
    let content: ReactNode = null;
    if (selectedRow?.recordId) {
      switch (panelId) {
        case "details":
          content = renderInspectorFieldEditors(selectedRow);
          break;
        case "evidence":
          content = renderEvidenceAttachSection(selectedRow);
          break;
        case "history":
          content = renderRowHistorySection();
          break;
        case "relationships":
          content = (
            <TimelineMentionsPanel
              canManageMentions={canManageMentions}
              entityIndex={entityIndex}
              getRelationshipLabel={getRelationshipLabel}
              hostEntities={hostEntities}
              identityEntities={identityEntities}
              inspectorMentions={inspectorMentions}
              relationshipEditors={renderRelationshipEditors(selectedRow)}
              onResolveTargetChange={onResolveTargetChange}
              onSelectMention={onSelectMention}
              onSetInspectorMessage={onSetInspectorMessage}
              onCreateEntityFromMention={onCreateEntityFromMention}
              onSubmitMentionAction={onSubmitMentionAction}
              selectedMention={selectedMention}
              selectedResolveTargetId={selectedResolveTargetId}
            />
          );
          break;
        case "workflow":
          content = renderWorkflowSection();
          break;
      }
    } else if (
      currentHistoryDeleted &&
      rowHistoryRecordId !== null &&
      panelId === "history"
    ) {
      content = renderRowHistorySection();
    }
    const supplement = renderPanelSupplement(panelId);
    return (
      <WorkbookInspectorPanelSection
        config={inspectorConfig}
        key={panelId}
        panelId={panelId}
      >
        {subjectPresentation === null ? null : (
          <WorkbookInspectorContextualActions
            config={inspectorConfig}
            currentIncidentRole={currentIncidentRole}
            disabledTokens={disabledTokens}
            capabilities={inspectorContextualCapabilities({
              config: inspectorConfig,
              panelId,
            })}
            subject={subjectPresentation}
            onAction={onFeatureAction}
          />
        )}
        {content}
        {supplement}
      </WorkbookInspectorPanelSection>
    );
  };
  return (
    <WorkbookInspectorShell
      accessibleLabel="Timeline inspector"
      noRowHeading="Timeline inspector"
      subject={subjectPresentation}
      testId={timelineInspectorTestId()}
      viewSchemaId={timelineViewSchemaId}
      onClose={onClose}
    >
      {selectedRow?.recordId ? (
        <>
          {inspectorConfig.panels.map((panel) => renderPanel(panel.panelId))}
          <WorkbookInspectorFeedbackView
            feedback={inspectorMessage}
            neutralStyle={bodyStyle}
            testId={timelineInspectorMessageTestId()}
          />
        </>
      ) : currentHistoryDeleted && rowHistoryRecordId !== null ? (
        <>
          {inspectorConfig.panels
            .filter((panel) => panel.panelId === "history")
            .map((panel) => renderPanel(panel.panelId))}
          <WorkbookInspectorFeedbackView
            feedback={inspectorMessage}
            neutralStyle={bodyStyle}
            testId={timelineInspectorMessageTestId()}
          />
        </>
      ) : (
        <WorkbookInspectorFeedbackView
          feedback={inspectorMessage}
          neutralStyle={bodyStyle}
          testId={timelineInspectorMessageTestId()}
        />
      )}
    </WorkbookInspectorShell>
  );
}
