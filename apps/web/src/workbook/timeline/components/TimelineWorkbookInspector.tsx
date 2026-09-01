import {
  timelineInspectorMessageTestId,
  timelineInspectorTestId,
} from "@cartulary/ui-contracts";
import type {
  InspectorConfig,
  InspectorFeatureGroup,
} from "@cartulary/view-contracts";
import type { ReactNode } from "react";
import type { WorkbookIncidentRole } from "../../../shared/workbookShellContracts";
import type { MentionResolutionAction } from "../../collaboration/workbookCollaborationMessages";
import {
  WorkbookInspectorContextualActions,
  WorkbookInspectorPanelSection,
  WorkbookInspectorShell,
} from "../../inspector/presentation/WorkbookInspectorPresentation";
import {
  type WorkbookInspectorSubjectPresentation,
  workbookInspectorSafePublicMessage,
} from "../../inspector/presentation/workbookInspectorPresentationModel";
import type { InspectorDisabledToken } from "../../inspector/semanticInspectorDispatcher";
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
  readonly inspectorMessage: string | null;
  readonly inspectorMentions: readonly InspectorMention[];
  readonly onResolveTargetChange: (value: string) => void;
  readonly onSelectMention: (rowRecordId: string, itemRef: string) => void;
  readonly onSetInspectorMessage: (message: string) => void;
  readonly onClose: () => void;
  readonly onFeatureAction: (featureGroup: InspectorFeatureGroup) => void;
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
  const disabledTokens = new Set<InspectorDisabledToken>();
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
            featureGroups={inspectorConfig.featureGroups.filter(
              (featureGroup) =>
                featureGroup.panelId === panelId &&
                (featureGroup.routeBinding.kind === "view_row_create" ||
                  featureGroup.routeBinding.kind === "indicator_observations" ||
                  featureGroup.routeBinding.kind === "indicator_lifecycle"),
            )}
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
          <InspectorMessage message={inspectorMessage} />
        </>
      ) : currentHistoryDeleted && rowHistoryRecordId !== null ? (
        <>
          {inspectorConfig.panels
            .filter((panel) => panel.panelId === "history")
            .map((panel) => renderPanel(panel.panelId))}
          <InspectorMessage message={inspectorMessage} />
        </>
      ) : (
        <InspectorMessage message={inspectorMessage} />
      )}
    </WorkbookInspectorShell>
  );
}

function InspectorMessage({ message }: { readonly message: string | null }) {
  return message ? (
    <p data-testid={timelineInspectorMessageTestId()} style={bodyStyle}>
      {workbookInspectorSafePublicMessage(message)}
    </p>
  ) : null;
}
