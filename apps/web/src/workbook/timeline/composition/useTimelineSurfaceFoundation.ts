import { useCallback, useMemo, useState } from "react";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { createTimelineClipboardPasteAdapter } from "../adapters/createTimelineClipboardPasteAdapter";
import { createTimelineEvidenceAttachmentAdapter } from "../adapters/createTimelineEvidenceAttachmentAdapter";
import { createTimelineHistoryAdapter } from "../adapters/createTimelineHistoryAdapter";
import { createTimelineMentionAdapter } from "../adapters/createTimelineMentionAdapter";
import { createTimelineRecordActionAdapter } from "../adapters/createTimelineRecordActionAdapter";
import { useTimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import { useTimelineMentions } from "../hooks/useTimelineMentions";
import { useTimelinePendingSaves } from "../hooks/useTimelinePendingSaves";
import { useTimelineRows } from "../hooks/useTimelineRows";
import { useTimelineWorkbookRuntime } from "../hooks/useTimelineWorkbookRuntime";
import type { PendingReplayRuntimeMeta } from "../models/timelineControllerPorts";
import type { TimelineWorkbookSurfaceRuntime } from "../models/timelineWorkbookSurfaceRuntime";

type TimelineSurfaceFoundationInput = {
  readonly apiBase: string | undefined;
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
  const mentionPort = useMemo(
    () => createTimelineMentionAdapter({ apiBase, incidentId }),
    [apiBase, incidentId],
  );
  const clipboardPastePort = useMemo(
    () => createTimelineClipboardPasteAdapter({ apiBase, incidentId }),
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
  const pendingSaves = useTimelinePendingSaves<PendingReplayRuntimeMeta>({
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
      clipboardPaste: clipboardPastePort,
      evidenceAttachment: evidenceAttachmentPort,
      history: historyPort,
      mention: mentionPort,
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
