import { timelineInspectorSectionTestId } from "@cartulary/ui-contracts";
import { type ReactNode, useCallback } from "react";
import { InspectorCreateRelatedWorkflow } from "../../inspector/InspectorCreateRelatedWorkflow";
import { buildEvidenceCountDisplayViewModel } from "../../models/evidenceLifecycleViewModel";
import type { GenericReferenceOptions } from "../../models/workbookReferenceOptions";
import type { TimelineCreateRelatedWorkflowState } from "../hooks/useTimelineCreateRelatedWorkflow";
import type { TimelineInspectorHistorySubject } from "../hooks/useTimelineHistoryState";
import type {
  RecordHistoryItem,
  RecordHistoryState,
  RowHistoryPendingAction,
} from "../models/timelineHistoryModel";
import {
  readTimelineCellValue,
  type TimelineCollectionBinding,
  type TimelineScalarBinding,
  timelineCollectionBindings,
  timelineInspectorBindings,
  type WorkbookRow,
} from "../models/workbookTimelineModel";
import { TimelineEvidencePanel } from "./TimelineEvidencePanel";
import { TimelineHistoryPanel } from "./TimelineHistoryPanel";
import {
  bodyStyle,
  inspectorActionStackStyle,
  inspectorSectionStyle,
  sectionTitleStyle,
} from "./TimelineWorkbookStyles";

export function useTimelineWorkbookInspectorSections({
  cancelCreateRelatedWorkflow,
  confirmRowHistoryPendingAction,
  createRelatedWorkflow,
  currentHistoryRecordId,
  handleTimelineEvidenceFiles,
  inspectorHistorySubject,
  openRowHistory,
  previewRowHistoryDeleteRestore,
  previewRowHistoryRollback,
  renderTimelineCollectionInput,
  renderTimelineInspectorEditor,
  rowHistory,
  rowHistoryPendingAction,
  setRowHistoryPendingAction,
  submitCreateRelatedWorkflow,
  timelineCreateRelatedReferenceOptions,
  updateCreateRelatedWorkflowDraft,
}: {
  readonly cancelCreateRelatedWorkflow: () => void;
  readonly confirmRowHistoryPendingAction: () => void;
  readonly createRelatedWorkflow: TimelineCreateRelatedWorkflowState | null;
  readonly currentHistoryRecordId: string | null;
  readonly handleTimelineEvidenceFiles: (
    row: WorkbookRow,
    files: FileList | File[],
  ) => void;
  readonly inspectorHistorySubject: TimelineInspectorHistorySubject;
  readonly openRowHistory: (recordId: string) => void;
  readonly previewRowHistoryDeleteRestore: (
    operation: "delete" | "restore",
  ) => void;
  readonly previewRowHistoryRollback: (
    item: RecordHistoryItem,
    action: "change_set" | "history_entry" | "row_restore",
  ) => void;
  readonly renderTimelineCollectionInput: (
    row: WorkbookRow,
    binding: TimelineCollectionBinding,
  ) => ReactNode;
  readonly renderTimelineInspectorEditor: (
    row: WorkbookRow,
    binding: TimelineScalarBinding,
  ) => ReactNode;
  readonly rowHistory: RecordHistoryState;
  readonly rowHistoryPendingAction: RowHistoryPendingAction | null;
  readonly setRowHistoryPendingAction: (
    action: RowHistoryPendingAction | null,
  ) => void;
  readonly submitCreateRelatedWorkflow: () => Promise<void>;
  readonly timelineCreateRelatedReferenceOptions: GenericReferenceOptions;
  readonly updateCreateRelatedWorkflowDraft: (
    featureGroupKey: string,
    fieldKey: string,
    value: string,
  ) => void;
}) {
  const renderInspectorFieldEditors = useCallback(
    (row: WorkbookRow) => (
      <section
        data-testid={timelineInspectorSectionTestId("operational-text")}
        style={inspectorSectionStyle}
      >
        <h3 style={sectionTitleStyle}>Operational Text</h3>
        <div style={inspectorActionStackStyle}>
          {timelineInspectorBindings.map((binding) =>
            renderTimelineInspectorEditor(row, binding),
          )}
        </div>
      </section>
    ),
    [renderTimelineInspectorEditor],
  );

  const renderRelationshipEditors = useCallback(
    (row: WorkbookRow) => (
      <div style={inspectorActionStackStyle}>
        {timelineCollectionBindings.map((binding) =>
          renderTimelineCollectionInput(row, binding),
        )}
      </div>
    ),
    [renderTimelineCollectionInput],
  );

  const renderEvidenceAttachSection = useCallback(
    (row: WorkbookRow) => {
      const countDisplay = buildEvidenceCountDisplayViewModel({
        projectedCount: readTimelineCellValue(
          row.rawRow,
          "timeline.evidence_count",
        ),
        projectedHasEvidence: readTimelineCellValue(
          row.rawRow,
          "timeline.has_evidence",
        ),
      });
      return (
        <TimelineEvidencePanel
          countDisplay={countDisplay}
          row={row}
          onFilesSelected={handleTimelineEvidenceFiles}
        />
      );
    },
    [handleTimelineEvidenceFiles],
  );

  const renderWorkflowSection = useCallback(() => {
    if (createRelatedWorkflow === null) {
      return (
        <p style={bodyStyle}>
          Select a workflow action to create a related row.
        </p>
      );
    }
    return (
      <InspectorCreateRelatedWorkflow
        referenceOptions={timelineCreateRelatedReferenceOptions}
        state={createRelatedWorkflow}
        onCancel={cancelCreateRelatedWorkflow}
        onSubmit={() => {
          void submitCreateRelatedWorkflow();
        }}
        onUpdateDraft={(fieldKey, value) => {
          updateCreateRelatedWorkflowDraft(
            createRelatedWorkflow.featureGroup.featureGroupKey,
            fieldKey,
            value,
          );
        }}
      />
    );
  }, [
    cancelCreateRelatedWorkflow,
    createRelatedWorkflow,
    submitCreateRelatedWorkflow,
    timelineCreateRelatedReferenceOptions,
    updateCreateRelatedWorkflowDraft,
  ]);

  const renderRowHistorySection = useCallback(
    () => (
      <TimelineHistoryPanel
        currentRecordId={currentHistoryRecordId}
        history={rowHistory}
        pendingAction={rowHistoryPendingAction}
        selectedActiveRowRecordId={
          inspectorHistorySubject.kind === "live"
            ? inspectorHistorySubject.recordId
            : null
        }
        onCancelPendingAction={() => {
          setRowHistoryPendingAction(null);
        }}
        onConfirmPendingAction={confirmRowHistoryPendingAction}
        onOpenHistory={openRowHistory}
        onPreviewDeleteRestore={previewRowHistoryDeleteRestore}
        onPreviewRollback={previewRowHistoryRollback}
      />
    ),
    [
      confirmRowHistoryPendingAction,
      currentHistoryRecordId,
      inspectorHistorySubject,
      openRowHistory,
      previewRowHistoryDeleteRestore,
      previewRowHistoryRollback,
      rowHistory,
      rowHistoryPendingAction,
      setRowHistoryPendingAction,
    ],
  );

  return {
    renderEvidenceAttachSection,
    renderInspectorFieldEditors,
    renderRelationshipEditors,
    renderRowHistorySection,
    renderWorkflowSection,
  };
}
