import {
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  timelineInspectorSectionTestId,
} from "@cartulary/ui-contracts";
import { type ReactNode, useCallback } from "react";
import { GenericMutationControl } from "../../components/GenericMutationControl";
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
  actionButtonStyle,
  bodyStyle,
  inlineButtonRowStyle,
  inspectorActionStackStyle,
  inspectorSectionStyle,
  labelStyle,
  secondaryActionButtonStyle,
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
    const workflow = createRelatedWorkflow;
    const writableFields = workflow.targetContract.fields.filter(
      (field) => field.writeKind !== "read_only",
    );
    return (
      <div style={inspectorActionStackStyle}>
        <p style={bodyStyle}>{workflow.targetContract.viewSchemaId}</p>
        {writableFields.map((field) => {
          const controlId = `timeline-create-related-${workflow.featureGroup.featureGroupKey}-${field.fieldKey}`;
          return (
            <label htmlFor={controlId} key={field.fieldKey} style={labelStyle}>
              {field.label}
              <GenericMutationControl
                collectionMode="add"
                field={field}
                id={controlId}
                referenceOptions={timelineCreateRelatedReferenceOptions}
                testId={genericCreateFieldTestId(field.fieldKey)}
                value={workflow.draft[field.fieldKey] ?? ""}
                onChange={(value) => {
                  updateCreateRelatedWorkflowDraft(
                    workflow.featureGroup.featureGroupKey,
                    field.fieldKey,
                    value,
                  );
                }}
              />
            </label>
          );
        })}
        <div style={inlineButtonRowStyle}>
          <button
            data-testid={genericCreateSubmitTestId(
              workflow.targetContract.viewSchemaId,
            )}
            disabled={workflow.isSubmitting}
            style={secondaryActionButtonStyle}
            type="button"
            onClick={() => {
              void submitCreateRelatedWorkflow();
            }}
          >
            Create related row
          </button>
          <button
            disabled={workflow.isSubmitting}
            style={actionButtonStyle}
            type="button"
            onClick={() => {
              cancelCreateRelatedWorkflow();
            }}
          >
            Cancel
          </button>
        </div>
        {workflow.message ? (
          <p role="alert" style={bodyStyle}>
            {workflow.message}
          </p>
        ) : null}
      </div>
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
