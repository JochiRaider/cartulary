import { IndicatorInspectorWorkflow } from "../../features/indicators/IndicatorInspectorWorkflow";
import { TimelineSupersessionEditor } from "../actions/TimelineSupersessionEditor";
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
      additionalDisabledReasons={model.additionalDisabledReasons}
      currentHistoryDeleted={model.currentHistoryDeleted}
      currentIncidentRole={model.currentIncidentRole}
      incidentClosed={model.incidentClosed}
      entityIndex={model.entityIndex}
      mentionActions={model.mentionActions}
      elementRegistry={model.elementRegistry}
      getRelationshipLabel={model.getRelationshipLabel}
      inspectorConfig={model.inspectorConfig}
      inspectorMessage={model.inspectorMessage}
      inspectorMentions={model.inspectorMentions}
      onClose={model.onClose}
      onFeatureAction={model.onFeatureAction}
      onSelectMention={model.onSelectMention}
      renderEvidenceAttachSection={model.renderEvidenceAttachSection}
      renderInspectorFieldEditors={model.renderInspectorFieldEditors}
      inspectorAttentionForRow={model.inspectorAttentionForRow}
      renderFeatureSupplement={(feature) =>
        (feature.featureGroupKey === "timeline.supersede" &&
          (model.captureEditor ||
            model.captureResult?.review.action === "supersede")) ||
        (feature.featureGroupKey === "timeline.mark_reviewed" &&
          model.captureResult?.review.action === "mark-reviewed") ? (
          <>
            {model.captureResult?.receipt &&
            feature.featureGroupKey ===
              `timeline.${model.captureResult.review.action.replace("-", "_")}` ? (
              <section aria-label="Timeline action result">
                <p>
                  {model.captureResult.receipt.operation === "mark-reviewed"
                    ? "Marked reviewed."
                    : "Superseded."}{" "}
                  Saved version {model.captureResult.receipt.data.row_version}.
                </p>
                {model.captureResult.review.action === "supersede" ? (
                  <p>
                    Replacement:{" "}
                    {model.captureResult.review.replacement
                      ? `${model.captureResult.review.replacement.label} · ${model.captureResult.review.replacement.context} · ${model.captureResult.review.replacement.recordId}`
                      : "No replacement"}
                  </p>
                ) : null}
                {model.captureResult.reconciliation !== "complete" ? (
                  <p>
                    Refresh is still required. Open Recovery to refresh this
                    result.
                  </p>
                ) : null}
              </section>
            ) : null}
            {model.captureEditor &&
            feature.featureGroupKey === "timeline.supersede" ? (
              <TimelineSupersessionEditor
                key={model.captureEditor.originKey}
                {...model.captureEditor}
              />
            ) : null}
          </>
        ) : model.indicatorInspectorHandler?.panelId === feature.panelId &&
          model.indicatorInspectorHandler.action ===
            feature.routeBinding.actionKey &&
          model.selectedRow?.recordId &&
          model.selectedRow.rowVersion !== null ? (
          <IndicatorInspectorWorkflow
            action={model.indicatorInspectorHandler.action}
            source={model.observationSource}
            sourceRecordId={model.selectedRow.recordId}
            onMutationCommitted={() => model.loadRows({ showLoading: false })}
          />
        ) : null
      }
      renderRelationshipEditor={model.renderRelationshipEditor}
      renderRowHistorySection={model.renderRowHistorySection}
      renderFeatureWorkflow={model.renderFeatureWorkflow}
      rowHistoryRecordId={model.rowHistoryRecordId}
      rowHistoryRowVersion={model.rowHistoryRowVersion}
      selectedMention={model.selectedMention}
      selectedRow={model.selectedRow}
    />
  );
}
