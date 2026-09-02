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
    renderEvidenceAttachSection,
    renderInspectorFieldEditors,
    renderRelationshipEditors,
    renderRowHistorySection,
    renderWorkflowSection,
  } = useTimelineWorkbookInspectorSections(sections);
  if (!isOpen) return null;

  const rowHistoryData = workbookRecordHistoryLoadedData(sections.rowHistory);
  return {
    ...model,
    renderEvidenceAttachSection,
    renderInspectorFieldEditors,
    renderRelationshipEditors,
    renderRowHistorySection,
    renderWorkflowSection,
    rowHistoryRecordId: currentHistoryDeleted ? currentHistoryRecordId : null,
    rowHistoryRowVersion:
      currentHistoryDeleted &&
      rowHistoryData?.record_id === currentHistoryRecordId
        ? rowHistoryData.row_version
        : null,
  };
}
