import { workbookRecordHistoryLoadedData } from "../../inspector/workbookRecordHistoryModel";
import { useTimelineWorkbookInspectorSections } from "../components/TimelineWorkbookInspectorSections";

type TimelineInspectorSectionsInput = Parameters<
  typeof useTimelineWorkbookInspectorSections
>[0];

export function useTimelineInspectorPresentation<
  const TModel extends Record<string, unknown>,
>({
  currentHistoryDeleted,
  currentHistoryRecordId,
  isOpen,
  model,
  sections,
}: {
  readonly currentHistoryDeleted: boolean;
  readonly currentHistoryRecordId: string | null;
  readonly isOpen: boolean;
  readonly model: TModel;
  readonly sections: TimelineInspectorSectionsInput;
}) {
  const {
    detailsOwner,
    renderEvidenceAttachSection,
    renderInspectorFieldEditors,
    renderRelationshipEditor,
    renderRowHistorySection,
    renderFeatureWorkflow,
  } = useTimelineWorkbookInspectorSections(sections);
  if (!isOpen) return null;

  const rowHistoryData = workbookRecordHistoryLoadedData(sections.rowHistory);
  return {
    ...model,
    detailsOwner,
    renderEvidenceAttachSection,
    renderInspectorFieldEditors,
    renderRelationshipEditor,
    renderRowHistorySection,
    renderFeatureWorkflow,
    rowHistoryRecordId: currentHistoryDeleted ? currentHistoryRecordId : null,
    rowHistoryRowVersion:
      currentHistoryDeleted &&
      rowHistoryData?.record_id === currentHistoryRecordId
        ? rowHistoryData.row_version
        : null,
  };
}
