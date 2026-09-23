import { type RefCallback, useCallback, useSyncExternalStore } from "react";
import type { RecordHistoryItem } from "../../adapters/workbookHistoryResponse";
import { InspectorCreateRelatedWorkflow } from "../../inspector/InspectorCreateRelatedWorkflow";
import type { InspectorRelatedRecordWorkflowState } from "../../inspector/inspectorRelatedRecordModel";
import { ownedInspectorRegion } from "../../inspector/presentation/WorkbookInspectorPanelContent";
import type { WorkbookInspectorAttention } from "../../inspector/presentation/workbookInspectorPresentationModel";

import type { HistoryBrowsingControls } from "../../inspector/WorkbookInspectorRecordHistory";
import { workbookInspectorOrdinaryAttention } from "../../inspector/workbookInspectorOrdinaryAttention";
import type { WorkbookRecordHistoryState } from "../../inspector/workbookRecordHistoryModel";
import { buildEvidenceCountDisplayViewModel } from "../../models/evidenceLifecycleViewModel";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
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
    row: WorkbookRow,
    files: FileList | File[],
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
  useSyncExternalStore(
    detailsOwner.drafts.subscribe,
    detailsOwner.drafts.getSnapshot,
  );
  useSyncExternalStore(
    detailsOwner.patches.subscribe,
    detailsOwner.patches.getSnapshot,
  );
  const inspectorAttentionForRow = (
    row: WorkbookRow,
  ): readonly WorkbookInspectorAttention[] =>
    workbookInspectorOrdinaryAttention(
      detailsOwner.drafts,
      detailsOwner.patches,
      timelineViewSchemaId,
      row.rawRow ?? null,
    );
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
    inspectorAttentionForRow,
    renderEvidenceAttachSection,
    renderInspectorFieldEditors,
    renderRelationshipEditor,
    renderRowHistorySection,
    renderFeatureWorkflow,
  };
}
