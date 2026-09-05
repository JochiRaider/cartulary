import { timelineInspectorSectionTestId } from "@cartulary/ui-contracts";
import { type ReactNode, type RefCallback, useCallback } from "react";
import { InspectorCreateRelatedWorkflow } from "../../inspector/InspectorCreateRelatedWorkflow";
import type { InspectorRelatedRecordWorkflowState } from "../../inspector/inspectorRelatedRecordModel";
import type { WorkbookInspectorSubject } from "../../inspector/workbookInspectorSubject";
import type {
  RecordHistoryItem,
  WorkbookRecordHistoryState,
} from "../../inspector/workbookRecordHistoryModel";
import { buildEvidenceCountDisplayViewModel } from "../../models/evidenceLifecycleViewModel";
import type { GenericReferenceOptions } from "../../models/workbookReferenceOptions";
import {
  type TimelineScalarBinding,
  timelineCollectionBindings,
  timelineInspectorBindings,
} from "../models/timelineFieldRegistry";
import {
  readTimelineCellValue,
  type WorkbookRow,
} from "../models/timelineRowModel";
import { TimelineEvidencePanel } from "./TimelineEvidencePanel";
import { TimelineHistoryPanel } from "./TimelineHistoryPanel";
import type { RenderTimelineCollectionInput } from "./TimelineWorkbookRendererTypes";
import {
  bodyStyle,
  inspectorActionStackStyle,
  inspectorSectionStyle,
  sectionTitleStyle,
} from "./TimelineWorkbookStyles";

export function useTimelineWorkbookInspectorSections({
  cancelCreateRelatedWorkflow,
  cancelRowHistoryPendingAction,
  canMutateHistory,
  confirmRowHistoryPendingAction,
  createRelatedWorkflow,
  handleTimelineEvidenceFiles,
  inspectorHistorySubject,
  openRowHistory,
  previewRowHistoryDeleteRestore,
  previewRowHistoryRollback,
  renderTimelineCollectionInput,
  renderTimelineInspectorEditor,
  rowHistory,
  submitCreateRelatedWorkflow,
  timelineCreateRelatedReferenceOptions,
  updateCreateRelatedWorkflowDraft,
}: {
  readonly cancelCreateRelatedWorkflow: () => void;
  readonly cancelRowHistoryPendingAction: () => void;
  readonly canMutateHistory: boolean;
  readonly confirmRowHistoryPendingAction: () => void;
  readonly createRelatedWorkflow: InspectorRelatedRecordWorkflowState | null;
  readonly handleTimelineEvidenceFiles: (
    row: WorkbookRow,
    files: FileList | File[],
  ) => void;
  readonly inspectorHistorySubject: WorkbookInspectorSubject | null;
  readonly openRowHistory: (recordId: string) => void;
  readonly previewRowHistoryDeleteRestore: (
    operation: "delete" | "restore",
  ) => void;
  readonly previewRowHistoryRollback: (
    item: RecordHistoryItem,
    action: "change_set" | "history_entry" | "row_restore",
  ) => void;
  readonly renderTimelineCollectionInput: RenderTimelineCollectionInput;
  readonly renderTimelineInspectorEditor: (
    row: WorkbookRow,
    binding: TimelineScalarBinding,
  ) => ReactNode;
  readonly rowHistory: WorkbookRecordHistoryState;
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
          renderTimelineCollectionInput(row, binding, undefined, "inspector"),
        )}
      </div>
    ),
    [renderTimelineCollectionInput],
  );

  const renderEvidenceAttachSection = useCallback(
    (row: WorkbookRow, elementRef?: RefCallback<HTMLElement>) => {
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
          elementRef={elementRef}
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
    (elementRef?: RefCallback<HTMLElement>) => (
      <TimelineHistoryPanel
        canMutate={canMutateHistory}
        elementRef={elementRef}
        history={rowHistory}
        selectedActiveRowRecordId={
          inspectorHistorySubject?.kind === "live"
            ? inspectorHistorySubject.recordId
            : null
        }
        onCancelPendingAction={cancelRowHistoryPendingAction}
        onConfirmPendingAction={confirmRowHistoryPendingAction}
        onOpenHistory={openRowHistory}
        onPreviewDeleteRestore={previewRowHistoryDeleteRestore}
        onPreviewRollback={previewRowHistoryRollback}
      />
    ),
    [
      cancelRowHistoryPendingAction,
      canMutateHistory,
      confirmRowHistoryPendingAction,
      inspectorHistorySubject,
      openRowHistory,
      previewRowHistoryDeleteRestore,
      previewRowHistoryRollback,
      rowHistory,
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
