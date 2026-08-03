import type { GridDensity, GridInteractionMode } from "@cartulary/grid-adapter";
import { requireViewContract } from "@cartulary/view-contracts";
import {
  type Dispatch,
  type ReactNode,
  type SetStateAction,
  useCallback,
  useMemo,
  useState,
} from "react";
import { createAppAuthorizationRecoveryPort } from "../app/api/appShellClient";
import type { SheetRef } from "../shared/sheetRef";
import { createWorkbookIncidentAdapter } from "../workbook/adapters/createWorkbookIncidentAdapter";
import { createWorkbookPendingMutationAdapter } from "../workbook/adapters/createWorkbookPendingMutationAdapter";
import { createWorkbookViewQueryAdapter } from "../workbook/adapters/createWorkbookViewQueryAdapter";
import { WorkbookCollaborationCoordinator } from "../workbook/collaboration/WorkbookCollaborationCoordinator";
import { WorkbookConflictResolver } from "../workbook/components/WorkbookConflictResolver";
import {
  defaultWorkbookLayoutState,
  moveWorkbookColumn,
  reorderWorkbookColumns,
  setWorkbookColumnHidden,
  setWorkbookColumnWidth,
  type WorkbookResolvedLayoutState,
} from "../workbook/layout/workbookColumnLayout";
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
import { WorkbookMutationRuntime } from "../workbook/runtime/WorkbookMutationRuntime";
import { TimelineWorkbook } from "../workbook/timeline/components/TimelineWorkbook";
import type {
  TimelineWorkbookEntityRow,
  TimelineWorkbookIncidentRole,
} from "../workbook/timeline/models/timelineWorkbookSurfaceRuntime";

const timelineContract = requireViewContract(timelineViewSchemaId);

export type TimelineWorkbookRuntimeFixtureProps = {
  readonly incidentId?: string | undefined;
  readonly apiBase?: string | undefined;
  readonly currentUserId?: string | null | undefined;
  readonly sheetRef?: SheetRef | undefined;
  readonly inspectorResetKey?: string | undefined;
  readonly reloadToken?: number | undefined;
  readonly renderInlineQueryControls?: boolean | undefined;
  readonly chromeMode?: WorkbookChromeMode | undefined;
  readonly savedViewSelector?: ReactNode | undefined;
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
  readonly onColumnWidthChange?:
    | ((fieldKey: string, width: number) => void)
    | undefined;
  readonly onResetColumns?: (() => void) | undefined;
  readonly onRefreshEntities?: (() => Promise<void> | void) | undefined;
  readonly interactionMode?: GridInteractionMode | undefined;
  readonly onIncidentAccessLost?: (() => void) | undefined;
};

export function TimelineWorkbookRuntimeFixture({
  incidentId = "10000000-0000-4000-8000-000000000001",
  apiBase,
  currentUserId = null,
  sheetRef = {
    kind: "view_schema",
    id: timelineViewSchemaId,
  },
  inspectorResetKey = timelineViewSchemaId,
  reloadToken = 0,
  renderInlineQueryControls = true,
  chromeMode = "base",
  savedViewSelector,
  showStatusPresence = true,
  filterDraft: providedFilterDraft,
  onFilterDraftChange,
  onQueryStateChange,
  queryState: providedQueryState,
  hostEntities = [],
  identityEntities = [],
  entityIndex = {},
  currentIncidentRole = "",
  density = "compact",
  layoutState: providedLayoutState,
  onColumnHiddenChange,
  onColumnMove,
  onColumnReorder,
  onColumnWidthChange,
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
  const [layoutState, setLayoutState] = useState<WorkbookResolvedLayoutState>(
    providedLayoutState ?? defaultWorkbookLayoutState(timelineContract),
  );
  const setColumnHidden = useCallback((fieldKey: string, hidden: boolean) => {
    setLayoutState((current) =>
      setWorkbookColumnHidden(timelineContract, current, fieldKey, hidden),
    );
  }, []);
  const moveColumn = useCallback(
    (fieldKey: string, direction: "earlier" | "later") => {
      setLayoutState((current) =>
        moveWorkbookColumn(timelineContract, current, fieldKey, direction),
      );
    },
    [],
  );
  const reorderColumn = useCallback(
    (sourceFieldKey: string, targetFieldKey: string) => {
      setLayoutState((current) =>
        reorderWorkbookColumns(
          timelineContract,
          current,
          sourceFieldKey,
          targetFieldKey,
        ),
      );
    },
    [],
  );
  const setColumnWidth = useCallback((fieldKey: string, width: number) => {
    setLayoutState((current) =>
      setWorkbookColumnWidth(timelineContract, current, fieldKey, width),
    );
  }, []);
  const resetColumns = useCallback(() => {
    setLayoutState(defaultWorkbookLayoutState(timelineContract));
  }, []);
  const [runtimeAssembly] = useState(() => {
    const transactionIds = createBrowserSecureTransactionIdPort();
    const pendingMutationPort = createWorkbookPendingMutationAdapter({
      apiBase,
      incidentId,
    });
    return {
      mutationRuntime: new WorkbookMutationRuntime(
        {
          clientInstanceId: "timeline-runtime-fixture",
          incidentId,
        },
        transactionIds,
        pendingMutationPort,
      ),
      mutationCommands: createWorkbookMutationCommandPorts({
        apiBase,
        incidentId,
        transactionIds,
      }),
      pendingMutationPort,
    };
  });
  const { mutationCommands, mutationRuntime, pendingMutationPort } =
    runtimeAssembly;
  const viewQuery = useMemo(
    () => createWorkbookViewQueryAdapter({ apiBase, incidentId }),
    [apiBase, incidentId],
  );
  const incidentPort = useMemo(
    () => createWorkbookIncidentAdapter({ apiBase, incidentId }),
    [apiBase, incidentId],
  );
  const [collaborationProjection] = useState(
    () =>
      new WorkbookCollaborationCoordinator({
        authorizationRecovery: createAppAuthorizationRecoveryPort(),
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
      }),
  );

  return (
    <div style={{ position: "relative", blockSize: "100%" }}>
      <TimelineWorkbook
        runtime={{
          attachCollaborationSession: true,
          collaborationProjection,
          mutationRuntime,
          pendingMutationPort,
          mutationCommands: mutationCommands.timeline,
          indicatorWorkflow: mutationCommands.indicators,
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
            savedViewSelector,
            viewBarQueryControls: undefined,
          },
          entities: {
            hosts: hostEntities,
            identities: identityEntities,
            index: entityIndex,
            refresh: onRefreshEntities,
          },
          layout: {
            commands: {
              onColumnHiddenChange: onColumnHiddenChange ?? setColumnHidden,
              onColumnMove: onColumnMove ?? moveColumn,
              onColumnReorder: onColumnReorder ?? reorderColumn,
              onColumnWidthChange: onColumnWidthChange ?? setColumnWidth,
              onResetColumns: onResetColumns ?? resetColumns,
            },
            snapshot: {
              chromeMode,
              density,
              interactionMode,
              showStatusPresence,
              state: providedLayoutState ?? layoutState,
            },
          },
          onIncidentAccessLost,
        }}
      />
      <WorkbookConflictResolver
        apiBase={apiBase}
        mutationRuntime={mutationRuntime}
        onActivateOrigin={() => undefined}
      />
    </div>
  );
}
