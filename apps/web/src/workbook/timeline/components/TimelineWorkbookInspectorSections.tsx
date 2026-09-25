import { type RefCallback, useCallback } from "react";
import type { RecordHistoryItem } from "../../adapters/workbookHistoryResponse";
import type { TimelineFileSource } from "../../features/evidence/timelineFileOperation";
import { InspectorCreateRelatedWorkflow } from "../../inspector/InspectorCreateRelatedWorkflow";
import type { InspectorRelatedRecordWorkflowState } from "../../inspector/inspectorRelatedRecordModel";
import { ownedInspectorRegion } from "../../inspector/presentation/WorkbookInspectorPanelContent";

import type { HistoryBrowsingControls } from "../../inspector/WorkbookInspectorRecordHistory";
import type { WorkbookRecordHistoryState } from "../../inspector/workbookRecordHistoryModel";
import { buildEvidenceCountDisplayViewModel } from "../../models/evidenceLifecycleViewModel";
import type { WorkbookRecordSubject } from "../../ports/WorkbookRecordSubject";
import {
  type CollectionFieldKey,
  timelineCollectionBindings,
} from "../models/timelineFieldRegistry";
import {
  readTimelineCellValue,
  type WorkbookRow,
} from "../models/timelineRowModel";
import { TimelineEvidencePanel } from "./TimelineEvidencePanel";
import { TimelineHistoryPanel } from "./TimelineHistoryPanel";
import {
  TimelineInspectorDetails,
  type TimelineInspectorDetailsOwner,
} from "./TimelineInspectorDetails";
import type { RenderTimelineCollectionInput } from "./TimelineWorkbookRendererTypes";

export function useTimelineWorkbookInspectorSections({
  cancelRowHistoryPendingAction,
  canMutateHistory,
  confirmRowHistoryPendingAction,
  createRelatedWorkflow,
  handleTimelineEvidenceFiles,
  inspectorHistorySubject,
  openRowHistory,
  previewRowHistoryDeleteRestore,
  previewRowHistoryRollback,
  historyBrowsingControls,
  renderTimelineCollectionInput,
  detailsOwner,
  rowHistory,
}: {
  readonly cancelRowHistoryPendingAction: () => void;
  readonly canMutateHistory: boolean;
  readonly confirmRowHistoryPendingAction: () => void;
  readonly createRelatedWorkflow: InspectorRelatedRecordWorkflowState | null;
  readonly handleTimelineEvidenceFiles: (
    source: TimelineFileSource,
    files: FileList | readonly File[],
  ) => void;
  readonly inspectorHistorySubject: WorkbookRecordSubject | null;
  readonly openRowHistory: (recordId: string) => void;
  readonly previewRowHistoryDeleteRestore: (
    operation: "delete" | "restore",
  ) => void;
  readonly previewRowHistoryRollback: (
    item: RecordHistoryItem,
    action: "change_set" | "history_entry" | "row_restore",
  ) => void;
  readonly renderTimelineCollectionInput: RenderTimelineCollectionInput;
  readonly detailsOwner: TimelineInspectorDetailsOwner;
  readonly historyBrowsingControls: HistoryBrowsingControls;
  readonly rowHistory: WorkbookRecordHistoryState;
}) {
  const renderInspectorFieldEditors = useCallback(
    (
      row: WorkbookRow,
      collectionDestinations: Readonly<
        Record<string, (() => void) | undefined>
      >,
    ) => (
      <TimelineInspectorDetails
        key={row.recordId ?? row.key}
        row={row}
        owner={detailsOwner}
        collectionDestinations={collectionDestinations}
      />
    ),
    [detailsOwner],
  );

  const renderRelationshipEditor = useCallback(
    (row: WorkbookRow, fieldKey: CollectionFieldKey) => {
      const binding = timelineCollectionBindings.find(
        (candidate) => candidate.fieldKey === fieldKey,
      );
      if (!binding)
        throw new Error(
          `Missing Timeline collection contribution: ${fieldKey}`,
        );
      return renderTimelineCollectionInput(
        row,
        binding,
        undefined,
        "inspector",
      );
    },
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

  const renderFeatureWorkflow = useCallback(
    (featureGroupKey: string) => {
      if (
        createRelatedWorkflow === null ||
        createRelatedWorkflow.featureGroup.featureGroupKey !== featureGroupKey
      ) {
        return null;
      }
      return <InspectorCreateRelatedWorkflow state={createRelatedWorkflow} />;
    },
    [createRelatedWorkflow],
  );

  const renderRowHistorySection = useCallback(
    (elementRef?: RefCallback<HTMLElement>) =>
      ownedInspectorRegion("record-history", (present) => (
        <TimelineHistoryPanel
          present={present}
          canMutate={canMutateHistory}
          elementRef={elementRef}
          history={rowHistory}
          browsingControls={historyBrowsingControls}
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
      )),
    [
      cancelRowHistoryPendingAction,
      canMutateHistory,
      confirmRowHistoryPendingAction,
      inspectorHistorySubject,
      openRowHistory,
      previewRowHistoryDeleteRestore,
      previewRowHistoryRollback,
      historyBrowsingControls,
      rowHistory,
    ],
  );

  return {
    detailsOwner,
    renderEvidenceAttachSection,
    renderInspectorFieldEditors,
    renderRelationshipEditor,
    renderRowHistorySection,
    renderFeatureWorkflow,
  };
}
