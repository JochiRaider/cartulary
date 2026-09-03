import { useCallback, useMemo, useState } from "react";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { createTimelineBulkTagCommandAdapter } from "../adapters/createTimelineBulkTagCommandAdapter";
import { createTimelineEvidenceAttachmentAdapter } from "../adapters/createTimelineEvidenceAttachmentAdapter";
import { createTimelineHistoryAdapter } from "../adapters/createTimelineHistoryAdapter";
import { createTimelineMentionEntityCreationAdapter } from "../adapters/createTimelineMentionEntityCreationAdapter";
import { createTimelineMentionResolutionAdapter } from "../adapters/createTimelineMentionResolutionAdapter";
import { createTimelineRecordActionAdapter } from "../adapters/createTimelineRecordActionAdapter";
import { useTimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import { useTimelineMentions } from "../hooks/useTimelineMentions";
import { useTimelinePendingSaves } from "../hooks/useTimelinePendingSaves";
import { useTimelineRows } from "../hooks/useTimelineRows";
import { useTimelineWorkbookRuntime } from "../hooks/useTimelineWorkbookRuntime";
import type { TimelineWorkbookSurfaceRuntime } from "../models/timelineWorkbookSurfaceRuntime";

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
  mutationCommands,
  mutationRuntime,
  query,
}: TimelineSurfaceFoundationInput) {
  const historyPort = useMemo(
    () => createTimelineHistoryAdapter({ apiBase }),
    [apiBase],
  );
  const recordActionPort = useMemo(
    () => createTimelineRecordActionAdapter({ apiBase }),
    [apiBase],
  );
  const mentionPorts = useMemo(
    () => ({
      entityCreation: createTimelineMentionEntityCreationAdapter({
        apiBase,
        incidentId,
      }),
      resolution: createTimelineMentionResolutionAdapter({ apiBase }),
    }),
    [apiBase, incidentId],
  );
  const evidenceAttachmentPort = useMemo(
    () =>
      createTimelineEvidenceAttachmentAdapter({
        apiBase,
        createClientTxnId: mutationCommands.identity.createLogicalActionId,
        incidentId,
      }),
    [apiBase, incidentId, mutationCommands.identity],
  );
  const bulkTagPort = useMemo(
    () =>
      createTimelineBulkTagCommandAdapter({
        apiBase,
        createClientTxnId: mutationCommands.identity.createLogicalActionId,
        incidentId,
      }),
    [apiBase, incidentId, mutationCommands.identity],
  );
  const runtime = useTimelineWorkbookRuntime({
    filterDraft: query.filterDraft,
    queryState: query.state,
    setFilterDraft: query.setFilterDraft,
    setQueryState: query.setState,
  });
  const [loadAccessLost, setLoadAccessLost] = useState(false);
  const [initialLoadGenerationKey, setInitialLoadGenerationKey] = useState(0);
  const [activeCollectionInputKey, setActiveCollectionInputKey] = useState<
    string | null
  >(null);
  const rows = useTimelineRows();
  const mentions = useTimelineMentions();
  const pendingSaves = useTimelinePendingSaves({
    mutationRuntime,
  });
  const editorDraftRegistry =
    useTimelineEditorDraftRegistry(timelineViewSchemaId);
  const recordTiming = useCallback(
    (name: string, details: Record<string, unknown> = {}) => {
      if (typeof performance === "undefined") {
        return;
      }
      performance.mark(`cartulary.workbook.${name}`, { detail: details });
    },
    [],
  );
  const activateCollectionInput = useCallback((focusKey: string) => {
    setActiveCollectionInputKey(focusKey);
  }, []);
  const deactivateCollectionInput = useCallback((focusKey: string) => {
    setActiveCollectionInputKey((current) =>
      current === focusKey ? null : current,
    );
  }, []);

  return {
    commands: {
      editor: {
        activateCollectionInput,
        deactivateCollectionInput,
      },
      lifecycle: {
        setInitialLoadGenerationKey,
        setIsInitialLoading: runtime.lifecycle.setIsInitialLoading,
        setIsRefreshing: runtime.lifecycle.setIsRefreshing,
        setLoadAccessLost,
        setLoadError: runtime.lifecycle.setLoadError,
        setRefreshError: runtime.lifecycle.setRefreshError,
      },
      mentions: mentions.commands,
      pendingSaves: pendingSaves.commands,
      query: runtime.query,
      recordTiming,
      rows: {
        allocateDraftIndex: rows.nextDraftIndex,
        ...rows.commands,
      },
    },
    ports: {
      bulkTag: bulkTagPort,
      clipboardPaste,
      evidenceAttachment: evidenceAttachmentPort,
      history: historyPort,
      mentions: mentionPorts,
      recordActions: recordActionPort,
    },
    refs: {
      editorDraftRegistry,
      pendingSaves: pendingSaves.refs,
      rows: rows.rowsRef,
    },
    snapshot: {
      editor: { activeCollectionInputKey },
      initialLoadGenerationKey,
      lifecycle: {
        isInitialLoading: runtime.lifecycle.isInitialLoading,
        isRefreshing: runtime.lifecycle.isRefreshing,
        loadAccessLost,
        loadError: runtime.lifecycle.loadError,
        refreshError: runtime.lifecycle.refreshError,
      },
      mentions: mentions.snapshot,
      pendingQueue: pendingSaves.snapshot.pendingQueueSnapshot,
      query: {
        filterDraft: runtime.query.filterDraft,
        queryState: runtime.query.queryState,
      },
      rows: rows.rows,
    },
  };
}
