import { IndicatorInspectorWorkflow } from "../../features/indicators/IndicatorInspectorWorkflow";
import { TimelineWorkbookInspector } from "../components/TimelineWorkbookInspector";
import type { TimelineWorkbookPresentationModel } from "./useTimelineWorkbookPresentation";

type TimelineWorkbookInspectorRegionModel = NonNullable<
  TimelineWorkbookPresentationModel["inspector"]
>;

export function TimelineWorkbookInspectorRegion({
  model,
}: {
  readonly model: TimelineWorkbookInspectorRegionModel;
}) {
  return (
    <TimelineWorkbookInspector
      canManageMentions={model.canManageMentions}
      currentHistoryDeleted={model.currentHistoryDeleted}
      currentIncidentRole={model.currentIncidentRole}
      incidentClosed={model.incidentClosed}
      entityIndex={model.entityIndex}
      elementRegistry={model.elementRegistry}
      getRelationshipLabel={model.getRelationshipLabel}
      hostEntities={model.hostEntities}
      identityEntities={model.identityEntities}
      inspectorConfig={model.inspectorConfig}
      inspectorMessage={model.inspectorMessage}
      inspectorMentions={model.inspectorMentions}
      onClose={model.onClose}
      onCreateEntityFromMention={model.onCreateEntityFromMention}
      onFeatureAction={model.onFeatureAction}
      onResolveTargetChange={model.onResolveTargetChange}
      onSelectMention={model.onSelectMention}
      onSetInspectorMessage={model.onSetInspectorMessage}
      onSubmitMentionAction={model.onSubmitMentionAction}
      renderEvidenceAttachSection={model.renderEvidenceAttachSection}
      renderInspectorFieldEditors={model.renderInspectorFieldEditors}
      renderPanelSupplement={(panelId) =>
        model.indicatorInspectorHandler?.panelId === panelId &&
        model.selectedRow?.recordId &&
        model.selectedRow.rowVersion !== null ? (
          <IndicatorInspectorWorkflow
            action={model.indicatorInspectorHandler.action}
            port={model.indicatorWorkflow}
            rowVersion={model.selectedRow.rowVersion}
            sourceFields={model.sourceFields}
            sourceRecordId={model.selectedRow.recordId}
            onMutationCommitted={() => model.loadRows({ showLoading: false })}
          />
        ) : null
      }
      renderRelationshipEditors={model.renderRelationshipEditors}
      renderRowHistorySection={model.renderRowHistorySection}
      renderWorkflowSection={() => model.renderWorkflowSection()}
      rowHistoryRecordId={model.rowHistoryRecordId}
      rowHistoryRowVersion={model.rowHistoryRowVersion}
      selectedMention={model.selectedMention}
      selectedResolveTargetId={model.selectedResolveTargetId}
      selectedRow={model.selectedRow}
    />
  );
}
