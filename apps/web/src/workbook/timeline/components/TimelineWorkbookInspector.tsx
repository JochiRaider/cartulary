import {
  timelineInspectorMessageTestId,
  timelineInspectorTestId,
} from "@cartulary/ui-contracts";
import type {
  InspectorConfig,
  InspectorDisabledCondition,
  InspectorPanelId,
} from "@cartulary/view-contracts";
import type { ReactNode, RefCallback } from "react";
import type { WorkbookIncidentRole } from "../../../shared/workbookShellContracts";
import type { InspectorContextualCapability } from "../../inspector/inspectorCapabilityResolver";
import { WorkbookInspectorFeedbackView } from "../../inspector/presentation/WorkbookInspectorFeedback";
import {
  inspectorPanel,
  savedInspectorRegion,
  type WorkbookInspectorRegion,
} from "../../inspector/presentation/WorkbookInspectorPanelContent";
import { WorkbookInspectorShell } from "../../inspector/presentation/WorkbookInspectorShell";
import type { WorkbookInspectorDisabledReason } from "../../inspector/presentation/workbookInspectorPresentationModel";
import { WorkbookInspectorDeclaredPanelList } from "../../inspector/WorkbookInspectorDeclaredPanelList";
import type { WorkbookInspectorFeedback } from "../../inspector/workbookInspectorErrorModel";
import { buildWorkbookInspectorSubject } from "../../inspector/workbookInspectorSubject";
import type { TimelineInspectorElementRegistry } from "../focus/timelineInspectorElementRegistry";
import type { CollectionFieldKey } from "../models/timelineFieldRegistry";
import type { WorkbookRow } from "../models/timelineRowModel";
import type { InspectorMention } from "../models/workbookMentionChips";
import type { TimelineMentionActions } from "./TimelineMentionActionControls";
import { TimelineMentionsPanel } from "./TimelineMentionsPanel";
import { bodyStyle } from "./TimelineWorkbookStyles";

export function TimelineWorkbookInspector({
  additionalDisabledReasons,
  currentHistoryDeleted,
  currentIncidentRole,
  incidentClosed,
  entityIndex,
  mentionActions,
  getRelationshipLabel,
  inspectorConfig,
  inspectorMessage,
  inspectorMentions,
  elementRegistry,
  onSelectMention,
  onClose,
  onFeatureAction,
  renderEvidenceAttachSection,
  renderInspectorFieldEditors,
  renderPanelSupplement,
  renderRelationshipEditor,
  renderWorkflowSection,
  renderRowHistorySection,
  rowHistoryRecordId,
  rowHistoryRowVersion,
  selectedMention,
  selectedRow,
}: {
  readonly additionalDisabledReasons?:
    | ReadonlyMap<string, WorkbookInspectorDisabledReason>
    | undefined;
  readonly currentHistoryDeleted: boolean;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly incidentClosed: boolean;
  readonly mentionActions: TimelineMentionActions;
  readonly entityIndex: Record<string, { label: string }>;
  readonly getRelationshipLabel: (
    fieldKey: InspectorMention["fieldKey"],
  ) => string;
  readonly inspectorConfig: InspectorConfig;
  readonly inspectorMessage: WorkbookInspectorFeedback | null;
  readonly inspectorMentions: readonly InspectorMention[];
  readonly elementRegistry: TimelineInspectorElementRegistry;
  readonly onSelectMention: (rowRecordId: string, itemRef: string) => void;
  readonly onClose: () => void;
  readonly onFeatureAction: (capability: InspectorContextualCapability) => void;
  readonly renderEvidenceAttachSection: (
    row: WorkbookRow,
    elementRef?: RefCallback<HTMLElement>,
  ) => ReactNode;
  readonly renderInspectorFieldEditors: (row: WorkbookRow) => ReactNode;
  readonly renderPanelSupplement: (panelId: InspectorPanelId) => ReactNode;
  readonly renderRelationshipEditor: (
    row: WorkbookRow,
    fieldKey: CollectionFieldKey,
  ) => ReactNode;
  readonly renderWorkflowSection: () => ReactNode;
  readonly renderRowHistorySection: (
    elementRef?: RefCallback<HTMLElement>,
  ) => WorkbookInspectorRegion;
  readonly rowHistoryRecordId: string | null;
  readonly rowHistoryRowVersion: number | null;
  readonly selectedMention: InspectorMention | null;
  readonly selectedRow: WorkbookRow | null;
}) {
  if (currentIncidentRole === null) return null;
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
  if (incidentClosed) disabledTokens.add("incident_closed");

  const subject = selectedRow?.recordId
    ? buildWorkbookInspectorSubject({
        config: inspectorConfig,
        kind: "live",
        label:
          selectedRow.values.activitySynopsisText.trim() ||
          "Selected timeline row",
        recordId: selectedRow.recordId,
        rowVersion: selectedRow.rowVersion,
        surfaceLabel: "Timeline",
      })
    : currentHistoryDeleted
      ? buildWorkbookInspectorSubject({
          config: inspectorConfig,
          kind: "deleted",
          label: "Deleted timeline row",
          recordId: rowHistoryRecordId,
          rowVersion: rowHistoryRowVersion,
          stateLabel: "Deleted",
          surfaceLabel: "Timeline",
        })
      : null;

  const visibleFeedback =
    inspectorMessage?.sourceRecordId &&
    inspectorMessage.sourceRecordId !== subject?.recordId
      ? null
      : inspectorMessage;
  const localFeedback = (panelId: InspectorPanelId) =>
    visibleFeedback?.destination &&
    "panel" in visibleFeedback.destination &&
    visibleFeedback.destination.panel === panelId &&
    visibleFeedback.destination.kind !== "relationship_item" ? (
      <WorkbookInspectorFeedbackView
        feedback={visibleFeedback}
        neutralStyle={bodyStyle}
        testId={timelineInspectorMessageTestId()}
      />
    ) : null;
  const withSupplement = (
    panelId: InspectorPanelId,
    region: WorkbookInspectorRegion,
  ) => ({
    ...inspectorPanel(region),
    authoring: (
      <>
        {subject?.kind === "live" ? renderPanelSupplement(panelId) : null}
        {localFeedback(panelId)}
      </>
    ),
  });
  const liveRow = subject?.kind === "live" ? selectedRow : null;
  const relationships =
    liveRow === null ? null : (
      <TimelineMentionsPanel
        feedback={
          visibleFeedback?.destination?.kind === "relationship_item"
            ? visibleFeedback
            : null
        }
        sourceRecordId={liveRow.recordId}
        entityIndex={entityIndex}
        actions={mentionActions}
        getRelationshipLabel={getRelationshipLabel}
        inspectorMentions={inspectorMentions}
        relationshipEditors={{
          "timeline.host_refs": renderRelationshipEditor(
            liveRow,
            "timeline.host_refs",
          ),
          "timeline.identity_refs": renderRelationshipEditor(
            liveRow,
            "timeline.identity_refs",
          ),
          "timeline.tags": renderRelationshipEditor(liveRow, "timeline.tags"),
        }}
        registerMention={elementRegistry.registerMention}
        registerCollectionItem={elementRegistry.registerCollectionItem}
        onSelectMention={onSelectMention}
        selectedMention={selectedMention}
      />
    );

  return (
    <WorkbookInspectorShell
      accessibleLabel="Timeline inspector"
      config={inspectorConfig}
      elementRef={elementRegistry.registerRoot}
      noRowHeading="Timeline inspector"
      subject={subject}
      testId={timelineInspectorTestId()}
      onClose={onClose}
    >
      <WorkbookInspectorDeclaredPanelList
        config={inspectorConfig}
        currentIncidentRole={currentIncidentRole}
        disabledTokens={disabledTokens}
        additionalDisabledReasons={additionalDisabledReasons}
        panelRef={(panelId, element) => {
          if (panelId !== "evidence" && panelId !== "history") {
            elementRegistry.registerPanel(panelId, element);
          }
        }}
        subject={subject}
        modelsByPanel={{
          details:
            liveRow === null
              ? undefined
              : withSupplement(
                  "details",
                  savedInspectorRegion("saved-fields", {
                    kind: "populated",
                    content: renderInspectorFieldEditors(liveRow),
                  }),
                ),
          evidence:
            liveRow === null
              ? undefined
              : withSupplement(
                  "evidence",
                  savedInspectorRegion("evidence-metadata", {
                    kind: "populated",
                    content: renderEvidenceAttachSection(liveRow, (element) =>
                      elementRegistry.registerPanel("evidence", element),
                    ),
                  }),
                ),
          history:
            subject === null
              ? undefined
              : withSupplement(
                  "history",
                  renderRowHistorySection((element) =>
                    elementRegistry.registerPanel("history", element),
                  ),
                ),
          relationships:
            liveRow === null
              ? undefined
              : withSupplement(
                  "relationships",
                  savedInspectorRegion("mentions", {
                    kind: "populated",
                    content: relationships,
                  }),
                ),
          workflow:
            liveRow === null
              ? undefined
              : {
                  ...withSupplement(
                    "workflow",
                    savedInspectorRegion("workflow", {
                      kind: "empty",
                      message: "Choose an available workflow action.",
                    }),
                  ),
                  authoring: (
                    <>
                      {renderWorkflowSection()}
                      {renderPanelSupplement("workflow")}
                      {localFeedback("workflow")}
                    </>
                  ),
                },
        }}
        onContextualAction={onFeatureAction}
      />
      <WorkbookInspectorFeedbackView
        feedback={
          visibleFeedback?.destination &&
          visibleFeedback.destination.kind !== "inspector"
            ? null
            : visibleFeedback
        }
        neutralStyle={bodyStyle}
        testId={timelineInspectorMessageTestId()}
      />
    </WorkbookInspectorShell>
  );
}
