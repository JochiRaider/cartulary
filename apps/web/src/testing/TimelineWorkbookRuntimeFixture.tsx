import type { GridDensity, GridInteractionMode } from "@cartulary/grid-adapter";
import { requireViewContract } from "@cartulary/view-contracts";
import {
  type Dispatch,
  type SetStateAction,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import type { SheetRef } from "../shared/sheetRef";
import { WorkbookRecoveryNavigation } from "../shared/workbookRecoveryNavigation";
import { createWorkbookBatchTransport } from "../workbook/adapters/createWorkbookBatchTransport";
import { createWorkbookClipboardPasteAdapter } from "../workbook/adapters/createWorkbookClipboardPasteAdapter";
import { createWorkbookIncidentAdapter } from "../workbook/adapters/createWorkbookIncidentAdapter";
import { createWorkbookPendingMutationAdapter } from "../workbook/adapters/createWorkbookPendingMutationAdapter";
import { createWorkbookRecordHistoryAdapter } from "../workbook/adapters/createWorkbookRecordHistoryAdapter";
import { createWorkbookViewQueryAdapter } from "../workbook/adapters/createWorkbookViewQueryAdapter";
import { createWorkbookCollaborationCoordinator } from "../workbook/collaboration/WorkbookCollaborationCoordinator";
import {
  systemWorkbookCollaborationClock,
  systemWorkbookCollaborationScheduler,
} from "../workbook/collaboration/workbookCollaborationTiming";
import { WorkbookActiveSurfaceFrame } from "../workbook/components/WorkbookActiveSurfaceFrame";
import { WorkbookBatchRecovery } from "../workbook/components/WorkbookBatchRecovery";
import { WorkbookHistoryContext } from "../workbook/history/WorkbookHistoryContext";
import { useWorkbookRecoveryFocus } from "../workbook/hooks/useWorkbookRecoveryFocus";
import { useWorkbookColumnLayoutController } from "../workbook/layout/useWorkbookColumnLayoutController";
import type { WorkbookResolvedLayoutState } from "../workbook/layout/workbookColumnLayout";
import type { WorkbookChromeMode } from "../workbook/layout/workbookResponsiveLayout";
import {
  defaultFilterDraft,
  emptyWorkbookQueryState,
  type FilterDraft,
  type WorkbookQueryState,
} from "../workbook/models/workbookQuery";
import { timelineViewSchemaId } from "../workbook/models/workbookSurfaceRegistry";
import { createWorkbookMutationCommandPorts } from "../workbook/mutations/createWorkbookMutationCommandPorts";
import { createBrowserSecureTransactionIdPort } from "../workbook/mutations/secureTransactionId";
import type { WorkbookReadScope } from "../workbook/query/WorkbookQueryRow";
import { createWorkbookMutationRuntime } from "../workbook/runtime/createWorkbookMutationRuntime";
import { reconcileTimelineCaptureReceipt } from "../workbook/timeline/actions/reconcileTimelineCaptureReceipt";
import { reconcileTimelineMentionReceipt } from "../workbook/timeline/actions/reconcileTimelineMentionReceipt";
import { TimelineCaptureRecovery } from "../workbook/timeline/actions/TimelineCaptureRecovery";
import { TimelineMentionRecovery } from "../workbook/timeline/actions/TimelineMentionRecovery";
import { timelineCaptureOwnerFor } from "../workbook/timeline/actions/timelineCaptureOwnerFor";
import { timelineMentionOwnerFor } from "../workbook/timeline/actions/timelineMentionOwnerFor";
import { createTimelineCandidateReader } from "../workbook/timeline/adapters/createTimelineCandidateReader";
import { createTimelineMentionEntityCreationAdapter } from "../workbook/timeline/adapters/createTimelineMentionEntityCreationAdapter";
import { createTimelineMentionResolutionAdapter } from "../workbook/timeline/adapters/createTimelineMentionResolutionAdapter";
import { createTimelineMentionSourceReader } from "../workbook/timeline/adapters/createTimelineMentionSourceReader";
import { createTimelineRecordActionAdapter } from "../workbook/timeline/adapters/createTimelineRecordActionAdapter";
import { TimelineWorkbook } from "../workbook/timeline/components/TimelineWorkbook";
import type {
  TimelineWorkbookEntityRow,
  TimelineWorkbookIncidentRole,
} from "../workbook/timeline/models/timelineWorkbookSurfaceRuntime";
import { WorkbookRecoveryFixture } from "./WorkbookRecoveryFixture";
import { workbookAuthorizationRecovery } from "./workbookAuthorizationTestSupport";

const timelineContract = requireViewContract(timelineViewSchemaId);
const idleGridEntryFocus = {
  acknowledge: () => undefined,
  cancel: () => undefined,
  request: { kind: "idle" as const },
};

export type TimelineWorkbookRuntimeFixtureProps = {
  readonly incidentId?: string | undefined;
  readonly apiBase?: string | undefined;
  readonly currentUserId?: string | null | undefined;
  readonly sheetRef?: SheetRef | undefined;
  readonly inspectorResetKey?: string | undefined;
  readonly reloadToken?: number | undefined;
  readonly renderInlineQueryControls?: boolean | undefined;
  readonly chromeMode?: WorkbookChromeMode | undefined;
  readonly incidentClosed?: boolean | undefined;
  readonly showStatusPresence?: boolean | undefined;
  readonly filterDraft?: FilterDraft | undefined;
  readonly onFilterDraftChange?:
    | Dispatch<SetStateAction<FilterDraft>>
    | undefined;
  readonly onQueryStateChange?:
    | Dispatch<SetStateAction<WorkbookQueryState>>
    | undefined;
  readonly queryState?: WorkbookQueryState | undefined;
  readonly hostEntities?: readonly TimelineWorkbookEntityRow[] | undefined;
  readonly identityEntities?: readonly TimelineWorkbookEntityRow[] | undefined;
  readonly entityIndex?: Record<string, TimelineWorkbookEntityRow> | undefined;
  readonly currentIncidentRole?:
    | TimelineWorkbookIncidentRole
    | null
    | undefined;
  readonly density?: GridDensity | undefined;
  readonly layoutState?: WorkbookResolvedLayoutState | undefined;
  readonly onColumnHiddenChange?:
    | ((fieldKey: string, hidden: boolean) => void)
    | undefined;
  readonly onColumnMove?:
    | ((fieldKey: string, direction: "earlier" | "later") => void)
    | undefined;
  readonly onColumnReorder?:
    | ((sourceFieldKey: string, targetFieldKey: string) => void)
    | undefined;
  readonly onResetColumns?: (() => void) | undefined;
  readonly onRefreshEntities?: (() => Promise<void> | void) | undefined;
  readonly interactionMode?: GridInteractionMode | undefined;
  readonly onIncidentAccessLost?: (() => void) | undefined;
};

export function TimelineWorkbookRuntimeFixture({
  incidentId = "10000000-0000-4000-8000-000000000001",
  apiBase,
  currentUserId = "fixture-actor",
  sheetRef = {
    kind: "view_schema",
    id: timelineViewSchemaId,
  },
  inspectorResetKey = timelineViewSchemaId,
  reloadToken = 0,
  renderInlineQueryControls = true,
  chromeMode = "base",
  incidentClosed = false,
  showStatusPresence = true,
  filterDraft: providedFilterDraft,
  onFilterDraftChange,
  onQueryStateChange,
  queryState: providedQueryState,
  hostEntities = [],
  identityEntities = [],
  entityIndex = {},
  currentIncidentRole = "editor",
  density = "compact",
  layoutState: providedLayoutState,
  onColumnHiddenChange,
  onColumnMove,
  onColumnReorder,
  onResetColumns,
  onRefreshEntities,
  interactionMode = { kind: "editable" },
  onIncidentAccessLost,
}: TimelineWorkbookRuntimeFixtureProps) {
  const [queryState, setQueryState] = useState<WorkbookQueryState>(
    providedQueryState ?? emptyWorkbookQueryState(),
  );
  const [filterDraft, setFilterDraft] = useState<FilterDraft>(
    providedFilterDraft ?? defaultFilterDraft(timelineContract),
  );
  const layoutOwner = useWorkbookColumnLayoutController({
    activeContract: timelineContract,
    contextKey: incidentId,
  });
  const columnControls = layoutOwner.snapshot.activeLayoutControls;
  const layoutState = columnControls.layoutState;
  useLayoutEffect(() => {
    if (providedLayoutState)
      layoutOwner.commands.applyLayoutStateForSurface(
        timelineContract.viewSchemaId,
        providedLayoutState,
      );
  }, [providedLayoutState, layoutOwner.commands.applyLayoutStateForSurface]);
  const [runtimeAssembly] = useState(() => {
    const transactionIds = createBrowserSecureTransactionIdPort();
    const pendingMutationPort = createWorkbookPendingMutationAdapter({
      apiBase,
      incidentId,
      readScope: (): WorkbookReadScope | null =>
        mutationRuntime.recordReadScope,
    });
    const mutationRuntime = createWorkbookMutationRuntime(
      { clientInstanceId: "timeline-runtime-fixture", incidentId },
      transactionIds,
      pendingMutationPort,
    );
    mutationRuntime.batches.configure(
      createWorkbookBatchTransport({
        apiBase,
        incidentId,
        readScope: () => mutationRuntime.recordReadScope,
      }),
    );
    return {
      clipboardPaste: createWorkbookClipboardPasteAdapter(
        mutationRuntime.batches,
      ),
      mutationRuntime,
      mutationCommands: createWorkbookMutationCommandPorts({
        readScope: () => mutationRuntime.recordReadScope,
        apiBase,
        incidentId,
        transactionIds,
        batches: mutationRuntime.batches,
      }),
    };
  });
  const { clipboardPaste, mutationCommands, mutationRuntime } = runtimeAssembly;
  useLayoutEffect(() => {
    mutationRuntime.setAuthority({
      actorId: currentUserId ?? "fixture-actor",
      sessionIdentity: "fixture-session",
      incidentId,
      role: currentIncidentRole ?? "",
      closed: incidentClosed,
    });
  }, [
    mutationRuntime,
    currentUserId,
    incidentId,
    currentIncidentRole,
    incidentClosed,
  ]);
  const timelineMentions = useMemo(
    () => timelineMentionOwnerFor(mutationRuntime),
    [mutationRuntime],
  );
  useLayoutEffect(() => {
    timelineMentions.configure(
      createTimelineMentionResolutionAdapter({ apiBase }),
    );
    timelineMentions.configureCreation(
      createTimelineMentionEntityCreationAdapter({ apiBase }),
    );
    return timelineMentions.registerReconciliation(async (receipt, scope) => {
      await reconcileTimelineMentionReceipt(
        timelineMentions,
        createTimelineMentionSourceReader({
          apiBase,
          incidentId,
          readScope: () => mutationRuntime.recordReadScope,
        }),
        receipt,
        scope,
      );
    });
  }, [timelineMentions, mutationRuntime, apiBase, incidentId]);
  useLayoutEffect(() => {
    if (!onRefreshEntities) return;
    return timelineMentions.registerCreationReconciliation(async (scope) => {
      if (!scope.isCurrent()) throw new Error("Entity refresh detached");
      await onRefreshEntities();
    });
  }, [timelineMentions, onRefreshEntities]);
  const timelineCapture = useMemo(
    () => timelineCaptureOwnerFor(mutationRuntime),
    [mutationRuntime],
  );
  useLayoutEffect(() => {
    timelineCapture.configure(
      createTimelineRecordActionAdapter({
        apiBase,
        readScope: () => mutationRuntime.recordReadScope,
      }),
      createTimelineCandidateReader({ apiBase, incidentId }),
      onIncidentAccessLost,
    );
    return timelineCapture.registerReconciliation(async (receipt, scope) => {
      await reconcileTimelineCaptureReceipt(
        timelineCapture,
        mutationRuntime.history,
        receipt,
        scope,
      );
      if (!scope.isCurrent()) throw new Error("Timeline fixture detached");
      await mutationRuntime.history.refreshSurface(timelineViewSchemaId);
    });
  }, [
    timelineCapture,
    apiBase,
    incidentId,
    onIncidentAccessLost,
    mutationRuntime,
  ]);
  useLayoutEffect(() => {
    mutationRuntime.history.configure(
      createWorkbookRecordHistoryAdapter({ apiBase, incidentId }),
    );
  }, [mutationRuntime, apiBase, incidentId]);
  useLayoutEffect(
    () => () => mutationRuntime.invalidate({ kind: "runtime_disposed" }),
    [mutationRuntime],
  );
  const activeSurfaceRef = useRef<HTMLElement | null>(null);
  const navigation = useMemo(() => {
    void mutationRuntime;
    return new WorkbookRecoveryNavigation();
  }, [mutationRuntime]);
  const invokerRef = useRef<HTMLElement | null>(null);
  useLayoutEffect(() => () => navigation.dispose(), [navigation]);
  const recoveryFocus = useWorkbookRecoveryFocus({
    activeSurfaceRef,
    runtime: mutationRuntime,
    onSessionRecovery: async () => undefined,
    navigation,
    invokerRef,
  });
  const viewQuery = useMemo(
    () =>
      createWorkbookViewQueryAdapter({
        apiBase,
        incidentId,
        readScope: () => mutationRuntime.recordReadScope,
      }),
    [apiBase, incidentId, mutationRuntime],
  );
  const incidentPort = useMemo(
    () => createWorkbookIncidentAdapter({ apiBase, incidentId }),
    [apiBase, incidentId],
  );
  const [collaborationProjection] = useState(() =>
    createWorkbookCollaborationCoordinator({
      authorizationRecovery: workbookAuthorizationRecovery(),
      clock: systemWorkbookCollaborationClock,
      continuityInvalidation: () => undefined,
      evidenceInvalidation: () => undefined,
      extensionInvalidation: () => undefined,
      incidentId,
      initialSheetRef: sheetRef,
      inspectorInvalidation: () => undefined,
      mutationRuntime,
      onAuthorizationRecovered: () => undefined,
      onIncidentAccessLost,
      queryInvalidation: () => undefined,
      scheduler: systemWorkbookCollaborationScheduler,
    }),
  );

  return (
    <WorkbookRecoveryFixture
      navigation={navigation}
      invokerRef={invokerRef}
      fallbackRef={activeSurfaceRef}
      standalone={false}
    >
      <WorkbookHistoryContext.Provider value={mutationRuntime}>
        <WorkbookBatchRecovery
          runtime={mutationRuntime}
          activateConflict={recoveryFocus.activate}
        />
        <TimelineCaptureRecovery owner={timelineCapture} />
        <TimelineMentionRecovery owner={timelineMentions} />
        <WorkbookActiveSurfaceFrame
          activeSurfaceRef={activeSurfaceRef}
          apiBase={apiBase}
          focus={recoveryFocus}
          mutationRuntime={mutationRuntime}
          sheetRef={sheetRef}
          onActivateOrigin={() => undefined}
          activeContent={
            <TimelineWorkbook
              runtime={{
                attachCollaborationSession: true,
                clipboardPaste,
                collaborationProjection,
                mutationRuntime,
                mutationCommands: mutationCommands.timeline,
                gridEntryFocus: idleGridEntryFocus,
                incident: {
                  id: incidentId,
                  apiBase,
                  continuityResetKey: inspectorResetKey,
                  currentUserId,
                  currentRole: currentIncidentRole,
                  incidentPort,
                  sheetRef,
                  inspectorResetKey,
                  reloadToken,
                },
                query: {
                  viewQuery,
                  state: providedQueryState ?? queryState,
                  setState: onQueryStateChange ?? setQueryState,
                  filterDraft: providedFilterDraft ?? filterDraft,
                  setFilterDraft: onFilterDraftChange ?? setFilterDraft,
                  renderInlineControls: renderInlineQueryControls,
                  viewBarWorkingSet: null,
                },
                entities: {
                  hosts: hostEntities,
                  identities: identityEntities,
                  index: entityIndex,
                  refresh: onRefreshEntities,
                },
                layout: {
                  commands: {
                    onColumnHiddenChange:
                      onColumnHiddenChange ??
                      columnControls.onColumnHiddenChange,
                    onColumnMove: onColumnMove ?? columnControls.onColumnMove,
                    onColumnReorder:
                      onColumnReorder ?? columnControls.onColumnReorder,
                    onColumnSizingIntent: columnControls.onColumnSizingIntent,
                    bindColumnSizing: columnControls.bindColumnSizing,
                    sizing: columnControls.sizing,
                    freezing: columnControls.freezing,
                    onResetColumns:
                      onResetColumns ?? columnControls.onResetColumns,
                  },
                  snapshot: {
                    chromeMode,
                    density,
                    incidentClosed,
                    interactionMode,
                    showStatusPresence,
                    state: providedLayoutState ?? layoutState,
                  },
                },
                onActivateConflict: recoveryFocus.activate,
                onAuthorityUncertain: onIncidentAccessLost,
              }}
            />
          }
        />
      </WorkbookHistoryContext.Provider>
    </WorkbookRecoveryFixture>
  );
}
