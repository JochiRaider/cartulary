import { timelineViewSchemaId } from "@cartulary/view-contracts";
import { useCallback, useEffect, useMemo, useState } from "react";
import { timelineMentionOwnerFor } from "../actions/timelineMentionOwnerFor";
import { createTimelineBulkTagCommandAdapter } from "../adapters/createTimelineBulkTagCommandAdapter";
import { createTimelineMentionCandidateReader } from "../adapters/createTimelineMentionCandidateReader";
import { createTimelineRecordActionAdapter } from "../adapters/createTimelineRecordActionAdapter";
import { createTimelineBulkTagReadiness } from "../bulk/createTimelineBulkTagReadiness";
import { useTimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import { useTimelineCommittedRows } from "../hooks/useTimelineCommittedRows";
import { useTimelineMentions } from "../hooks/useTimelineMentions";
import { useTimelineRows } from "../hooks/useTimelineRows";
import { useTimelineWorkbookRuntime } from "../hooks/useTimelineWorkbookRuntime";
import { timelinePendingSavesRefsFor } from "../models/timelinePendingSaves";
import type { TimelineWorkbookSurfaceRuntime } from "../models/timelineWorkbookSurfaceRuntime";
import { timelineMutationOwnerFor } from "../mutations/WorkbookTimelineMutationOwner";

type TimelineSurfaceFoundationInput = {
  readonly apiBase: string | undefined;
  readonly clipboardPaste: TimelineWorkbookSurfaceRuntime["clipboardPaste"];
  readonly incidentId: string;
  readonly mutationCommands: TimelineWorkbookSurfaceRuntime["mutationCommands"];
  readonly mutationRuntime: TimelineWorkbookSurfaceRuntime["mutationRuntime"];
  readonly query: Pick<
    TimelineWorkbookSurfaceRuntime["query"],
    "filterDraft" | "setFilterDraft" | "setState" | "state"
  >;
};

export function useTimelineSurfaceFoundation({
  apiBase,
  clipboardPaste,
  incidentId,
  mutationRuntime,
  query,
}: TimelineSurfaceFoundationInput) {
  const recordActionPort = useMemo(
    () =>
      createTimelineRecordActionAdapter({
        apiBase,
        readScope: () => mutationRuntime.recordReadScope,
      }),
    [apiBase, mutationRuntime],
  );
  const mentionOwner = useMemo(
    () => timelineMentionOwnerFor(mutationRuntime),
    [mutationRuntime],
  );
  const mentionCandidates = useMemo(
    () => createTimelineMentionCandidateReader({ apiBase, incidentId }),
    [apiBase, incidentId],
  );
  const evidenceAttachmentPort = mutationRuntime.timelineFiles;
  const bulkTagPort = useMemo(
    () => createTimelineBulkTagCommandAdapter(mutationRuntime.batches),
    [mutationRuntime],
  );
  const runtime = useTimelineWorkbookRuntime({
    filterDraft: query.filterDraft,
    queryState: query.state,
    setFilterDraft: query.setFilterDraft,
    setQueryState: query.setState,
  });
  const [loadAccessLost, setLoadAccessLost] = useState(false);
  const [initialLoadGenerationKey, setInitialLoadGenerationKey] = useState(0);
  const mutationOwner = timelineMutationOwnerFor(mutationRuntime);
  const rows = useTimelineRows(mutationOwner);
  const mentions = useTimelineMentions(mentionOwner);
  const pendingSaves = timelinePendingSavesRefsFor(
    mutationRuntime,
    mutationRuntime.pendingQueue(),
  );
  const editorDraftRegistry = useTimelineEditorDraftRegistry(
    mutationRuntime.localDraftsForSurface(timelineViewSchemaId),
    mutationOwner.capture,
  );
  const bulkTagReadiness = useMemo(
    () =>
      createTimelineBulkTagReadiness({
        runtime: mutationRuntime,
        drafts: editorDraftRegistry,
        pending: pendingSaves,
        rows: rows.rowsRef,
      }),
    [mutationRuntime, editorDraftRegistry, pendingSaves, rows.rowsRef],
  );
  const committedRows = useTimelineCommittedRows({
    rowsRef: rows.rowsRef,
    mutationRuntime,
    materializeRow: editorDraftRegistry.materializeRow,
  });
  const clearCommittedRows = committedRows.commands.clearProtectedRows;
  useEffect(() => {
    if (loadAccessLost) clearCommittedRows();
  }, [clearCommittedRows, loadAccessLost]);
  const recordTiming = useCallback(
    (name: string, details: Record<string, unknown> = {}) => {
      if (typeof performance === "undefined") {
        return;
      }
      performance.mark(`cartulary.workbook.${name}`, { detail: details });
    },
    [],
  );

  return {
    commands: {
      lifecycle: {
        setInitialLoadGenerationKey,
        setIsInitialLoading: runtime.lifecycle.setIsInitialLoading,
        setIsRefreshing: runtime.lifecycle.setIsRefreshing,
        setLoadAccessLost,
        setLoadError: runtime.lifecycle.setLoadError,
        setRefreshError: runtime.lifecycle.setRefreshError,
        setOperationError: runtime.lifecycle.setOperationError,
        setMutationError: runtime.lifecycle.setMutationError,
      },
      mentions: mentions.commands,
      query: runtime.query,
      recordTiming,
      rows: {
        allocateDraftIndex: rows.nextDraftIndex,
        ...rows.commands,
      },
    },
    ports: {
      committedRows: committedRows.commands,
      bulkTag: bulkTagPort,
      bulkTagReadiness,
      clipboardPaste,
      evidenceAttachment: evidenceAttachmentPort,
      mentionOwner,
      mentionCandidates,
      recordActions: recordActionPort,
    },
    refs: {
      editorDraftRegistry,
      pendingSaves,
      rows: rows.rowsRef,
    },
    snapshot: {
      initialLoadGenerationKey,
      lifecycle: {
        isInitialLoading: runtime.lifecycle.isInitialLoading,
        isRefreshing: runtime.lifecycle.isRefreshing,
        loadAccessLost,
        loadError: runtime.lifecycle.loadError,
        refreshError: runtime.lifecycle.refreshError,
        operationError: runtime.lifecycle.operationError,
      },
      mentions: mentions.snapshot,
      query: {
        filterDraft: runtime.query.filterDraft,
        queryState: runtime.query.queryState,
      },
      rows: rows.rows,
    },
  };
}
