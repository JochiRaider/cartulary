import {
  networkAnalysisTestId,
  surfaceTabTestId,
  workbookIncidentIdentityTestId,
  workbookResponsiveBandTestId,
  workbookShellReadyTestId,
  workbookSurfacesMenuOptionTestId,
  workbookSurfacesMenuTestId,
  workbookSurfacesMenuTriggerTestId,
} from "@cartulary/ui-contracts";
import { requireViewContract } from "@cartulary/view-contracts";
import {
  lazy,
  type ReactNode,
  Suspense,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import {
  IncidentCollaborationSession,
  useIncidentCollaborationSession,
} from "../collaboration/IncidentCollaborationSession";
import { ExtensionAvailabilityProvider } from "../extensions/ExtensionAvailabilityContext";
import { ExtensionAvailabilityController } from "../extensions/extensionAvailability";
import {
  importProfileId,
  importRouteFamily,
  networkAnalysisSheetRef,
  networkAnalysisWorkspaceKey,
  networkFlowActivityProfileId,
} from "../extensions/extensionWorkspaceIdentities";
import type { GeneratedExtensionProfileResource } from "../services/extensionContractAdapter";
import type { AuthorizationRecoveryPort } from "../shared/authorizationRecovery";
import { sheetRefKey } from "../shared/sheetRef";
import type {
  WorkbookAccountApplicationMenuProps,
  WorkbookAccountModel,
  WorkbookIncidentControlsRendererProps,
  WorkbookIncidentRole,
  WorkbookIncidentSnapshot,
} from "../shared/workbookShellContracts";
import { createWorkbookIncidentAdapter } from "./adapters/createWorkbookIncidentAdapter";
import { createWorkbookPendingMutationAdapter } from "./adapters/createWorkbookPendingMutationAdapter";
import { createWorkbookPreferenceAdapter } from "./adapters/createWorkbookPreferenceAdapter";
import { createWorkbookSavedViewAdapter } from "./adapters/createWorkbookSavedViewAdapter";
import { createWorkbookStartupAdapter } from "./adapters/createWorkbookStartupAdapter";
import { createWorkbookViewQueryAdapter } from "./adapters/createWorkbookViewQueryAdapter";
import { useWorkbookCollaborationCoordinatorSession } from "./collaboration/useWorkbookCollaborationCoordinator";
import { WorkbookCollaborationCoordinator } from "./collaboration/WorkbookCollaborationCoordinator";
import type { WorkbookActiveSurfacePort } from "./collaboration/workbookSurfacePort";
import { ActiveSurfaceSavedViewSelector } from "./components/ActiveSurfaceSavedViewSelector";
import { IncidentControlsDrawer } from "./components/IncidentControlsDrawer";
import { SystemViewSwitcher } from "./components/SystemViewSwitcher";
import { WorkbookConflictResolver } from "./components/WorkbookConflictResolver";
import { WorkbookGridControls } from "./components/WorkbookGridControls";
import {
  WorkbookShellSlotRegion,
  workbookShellId,
} from "./components/WorkbookShellSlots";
import { WorkbookPresenceSummary } from "./components/WorkbookStatusStrip";
import { useIncidentControlsDrawer } from "./hooks/useIncidentControlsDrawer";
import { useWorkbookIncidentIdentity } from "./hooks/useWorkbookIncidentIdentity";
import { useWorkbookPendingGridFocus } from "./hooks/useWorkbookPendingGridFocus";
import { useWorkbookProjectionRefreshController } from "./hooks/useWorkbookProjectionRefreshController";
import { useWorkbookShellRuntime } from "./hooks/useWorkbookShellRuntime";
import { useWorkbookLayoutFacade } from "./layout/useWorkbookLayoutFacade";
import type { AccountDensityMode } from "./layout/workbookDensity";
import {
  activeSystemViewTitleStyle,
  currentUserChipStyle,
  currentUserSlotStyle,
  panelStyle,
  shellActiveSurfaceStyle,
  shellContentNoticeStyle,
  shellContentRegionStyle,
  shellIncidentIdentityStyle,
  shellIncidentTitleStyle,
  shellTopBarActionsStyle,
  shellTopBarStyle,
  shellTopBarUnsupportedStyle,
  shellTopBarValueStyle,
  surfaceMenuTriggerStyle,
  surfacesMenuFrameStyle,
  surfacesMenuItemSelectedStyle,
  surfacesMenuItemStyle,
  surfacesMenuStyle,
  surfaceTabActiveStyle,
  surfaceTabStyle,
  systemViewSlotStyle,
  tabStripStyle,
} from "./layout/workbookShellStyles";
import { workbookGridInteractionMode } from "./models/workbookGridState";
import type { WorkbookIncidentIdentity } from "./models/workbookIncidentIdentity";
import {
  assessmentsViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  requiredBuiltInWorkbookSurfaceIds,
  timelineViewSchemaId,
} from "./models/workbookSurfaceRegistry";
import { createWorkbookMutationCommandPorts } from "./mutations/createWorkbookMutationCommandPorts";
import { createBrowserSecureTransactionIdPort } from "./mutations/secureTransactionId";
import { useAssessmentSurfaceQuery } from "./query/useAssessmentSurfaceQuery";
import { useEntitySurfaceQuery } from "./query/useEntitySurfaceQuery";
import { useGenericSurfaceQuery } from "./query/useGenericSurfaceQuery";
import { useWorkbookMutationRuntime } from "./runtime/useWorkbookMutationRuntime";
import { WorkbookMutationRuntime } from "./runtime/WorkbookMutationRuntime";
import {
  createReferenceQueryBroker,
  type ReferenceQueryBrokerPort,
} from "./services/referenceQueryBroker";
import { WorkbookSurfacesFacade } from "./surfaces/WorkbookSurfacesFacade";
import { displayInitials } from "./utils/workbookPresence";

export type {
  WorkbookAccountApplicationMenuProps,
  WorkbookIncidentControlsRendererProps,
};

type IncidentRole = WorkbookIncidentRole;

function recordWorkbookPendingMutationTiming(
  name: string,
  details: Readonly<Record<string, unknown>> = {},
) {
  if (typeof performance === "undefined") return;
  performance.mark(`cartulary.workbook.${name}`, { detail: details });
}

type WorkbookShellProps = {
  authorizationRecovery: AuthorizationRecoveryPort;
  incidentId: string;
  apiBase?: string | undefined;
  account?: WorkbookAccountModel | undefined;
  accountDensityMode?: AccountDensityMode | undefined;
  accountApplicationMenu?:
    | ((props: WorkbookAccountApplicationMenuProps) => ReactNode)
    | undefined;
  currentUserLabel?: string | undefined;
  initialIncidentIdentity?: WorkbookIncidentIdentity | undefined;
  extensionProfiles?: readonly GeneratedExtensionProfileResource[] | undefined;
  onIncidentSnapshot?:
    | ((incident: WorkbookIncidentSnapshot) => void)
    | undefined;
  onIncidentAccessLost?: (() => void) | undefined;
  renderIncidentControls?:
    | ((props: WorkbookIncidentControlsRendererProps) => ReactNode)
    | undefined;
};

type ExtensionWorkspaceRendererProps = {
  readonly apiBase: string | undefined;
  readonly currentIncidentRole: IncidentRole | null;
  readonly incidentId: string;
  readonly onIncidentAccessLost: (() => void) | undefined;
};

function extensionWorkspaceRegistryKey(
  extensionProfileId: string,
  workspaceKey: string,
): string {
  return `${extensionProfileId}:${workspaceKey}`;
}

const extensionWorkspaceRenderers: Readonly<
  Record<string, (props: ExtensionWorkspaceRendererProps) => ReactNode>
> = {
  [extensionWorkspaceRegistryKey(
    networkFlowActivityProfileId,
    networkAnalysisWorkspaceKey,
  )]: (props) => (
    <Suspense fallback={null}>
      <LazyNetworkFlowFeature {...props} />
    </Suspense>
  ),
};

const LazyNetworkFlowFeature = lazy(async () => {
  const feature = await import("./features/NetworkFlowFeature");
  return { default: feature.NetworkFlowFeature };
});

const LazyImportAssistantFeature = lazy(async () => {
  const feature = await import("./features/ImportAssistantFeature");
  return { default: feature.ImportAssistantFeature };
});

function WorkbookShellContent({
  authorizationRecovery,
  incidentId,
  apiBase,
  account,
  accountDensityMode,
  accountApplicationMenu,
  currentUserLabel,
  initialIncidentIdentity,
  extensionProfiles = [],
  onIncidentSnapshot,
  onIncidentAccessLost,
  renderIncidentControls,
}: WorkbookShellProps) {
  const collaborationSession = useIncidentCollaborationSession();
  const extensionAvailability = useMemo(
    () =>
      new ExtensionAvailabilityController({
        clientInstanceId: collaborationSession.clientInstanceId,
        incidentId,
      }),
    [collaborationSession.clientInstanceId, incidentId],
  );
  const [extensionAvailabilityRevision, setExtensionAvailabilityRevision] =
    useState(0);
  extensionAvailability.setDiscovery(extensionProfiles);
  const handleExtensionAvailabilityChange = useCallback(() => {
    setExtensionAvailabilityRevision((current) => current + 1);
  }, []);
  const transactionIds = useMemo(createBrowserSecureTransactionIdPort, []);
  const pendingMutationPort = useMemo(
    () =>
      createWorkbookPendingMutationAdapter({
        apiBase,
        incidentId,
        recordTiming: recordWorkbookPendingMutationTiming,
      }),
    [apiBase, incidentId],
  );
  const mutationRuntime = useMemo(
    () =>
      new WorkbookMutationRuntime(
        {
          clientInstanceId: collaborationSession.clientInstanceId,
          incidentId,
        },
        transactionIds,
        pendingMutationPort,
      ),
    [
      collaborationSession.clientInstanceId,
      incidentId,
      pendingMutationPort,
      transactionIds,
    ],
  );
  const mutationCommands = useMemo(
    () =>
      createWorkbookMutationCommandPorts({
        apiBase,
        incidentId,
        transactionIds,
      }),
    [apiBase, incidentId, transactionIds],
  );
  const mutationSnapshot = useWorkbookMutationRuntime(mutationRuntime);
  const surfaceSelectionVersionRef = useRef(0);
  const incidentPort = useMemo(
    () => createWorkbookIncidentAdapter({ apiBase, incidentId }),
    [apiBase, incidentId],
  );
  const preferencePort = useMemo(
    () => createWorkbookPreferenceAdapter({ apiBase, incidentId }),
    [apiBase, incidentId],
  );
  const startupPort = useMemo(
    () => createWorkbookStartupAdapter({ apiBase, incidentId }),
    [apiBase, incidentId],
  );
  const savedViewPort = useMemo(
    () => createWorkbookSavedViewAdapter({ apiBase, incidentId }),
    [apiBase, incidentId],
  );
  const workbookRuntime = useWorkbookShellRuntime({
    incidentId,
    onIncidentAccessLost,
    surfaceSelectionVersionRef,
    extensionAvailability,
    onExtensionAvailabilityChange: handleExtensionAvailabilityChange,
    preferencePort,
    savedViewPort,
    startupPort,
  });
  const {
    activeContract,
    activeLayoutControls,
    activeLayoutState,
    activeQueryControls,
    activeSavedViewModified,
    assessmentQueryState,
    genericQueryState,
    hostQueryState,
    identityQueryState,
    pendingGridFocusSurface,
    savedViews,
    sheetReloadToken,
    startupSheetRef,
    surface,
    timelineQueryState,
  } = workbookRuntime.snapshot;
  const {
    createSavedView,
    deleteSavedView,
    duplicateSavedView,
    selectSavedView,
    selectExtensionWorkspace,
    selectWorkbookSurface,
    setAssessmentQueryState,
    setGenericQueryState,
    setHostQueryState,
    setIdentityQueryState,
    setPendingGridFocusSurface,
    setTimelineQueryState,
    setWorkbookDefaultSheetRef,
    setWorkbookHomeSheetRef,
    updateSavedView,
  } = workbookRuntime.commands;
  const networkAnalysisRef = networkAnalysisSheetRef();
  const networkAnalysisActive =
    startupSheetRef.kind === "extension_workspace" &&
    startupSheetRef.extension_profile_id === networkFlowActivityProfileId &&
    startupSheetRef.workspace_key === networkAnalysisWorkspaceKey;
  const [currentUserId, setCurrentUserId] = useState<string | null>(
    () => account?.user_id ?? null,
  );
  const [currentIncidentRole, setCurrentIncidentRole] =
    useState<IncidentRole | null>(null);
  const authorizationGeneration = `${currentUserId ?? "anonymous"}:${currentIncidentRole ?? "none"}`;
  const viewQuery = useMemo(
    () => createWorkbookViewQueryAdapter({ apiBase, incidentId }),
    [apiBase, incidentId],
  );
  const referenceQueryBroker = useMemo<ReferenceQueryBrokerPort>(
    () =>
      createReferenceQueryBroker({
        authorizationGeneration,
        viewQuery,
      }),
    [authorizationGeneration, viewQuery],
  );
  useEffect(
    () => () => {
      referenceQueryBroker.dispose();
    },
    [referenceQueryBroker],
  );
  const { incidentIdentity, incidentIdentityError } =
    useWorkbookIncidentIdentity({
      incidentPort,
      incidentId,
      initialIncidentIdentity,
      onIncidentAccessLost,
      onIncidentSnapshot,
    });
  const [surfacesMenuOpen, setSurfacesMenuOpen] = useState(false);
  const surfacesMenuTriggerRef = useRef<HTMLButtonElement>(null);
  const importAssistantAvailable = extensionAvailability.isRouteAvailable(
    importProfileId,
    importRouteFamily,
  );
  const {
    accountIncidentControls,
    activeMenuItem: activeIncidentControlsMenuItem,
    closeButtonRef: incidentControlsCloseButtonRef,
    closeDrawer: closeIncidentControlsDrawer,
    drawerSection: incidentControlsDrawerSection,
  } = useIncidentControlsDrawer(importAssistantAvailable);
  const genericSurfaceActive =
    surface !== timelineViewSchemaId &&
    surface !== hostsViewSchemaId &&
    surface !== identitiesViewSchemaId &&
    surface !== assessmentsViewSchemaId;
  const {
    applyRecordChanged: applyGenericRecordChanged,
    invalidate: invalidateGenericQuery,
    loadState: genericLoadState,
    refresh: loadGenericSurface,
    rows: genericRows,
  } = useGenericSurfaceQuery({
    active: genericSurfaceActive,
    contract: activeContract,
    onIncidentAccessLost,
    queryState: genericQueryState,
    viewQuery,
    viewSchemaId: surface,
  });
  const {
    applyRecordChanged: applyAssessmentRecordChanged,
    invalidate: invalidateAssessmentQuery,
    loadState: assessmentLoadState,
    refresh: loadAssessmentSurface,
    rows: assessmentRows,
  } = useAssessmentSurfaceQuery({
    active: surface === assessmentsViewSchemaId,
    onIncidentAccessLost,
    queryState: assessmentQueryState,
    viewQuery,
  });
  const {
    applyRecordChanged: applyEntityRecordChanged,
    invalidate: invalidateEntityQuery,
    entityIndex,
    hostRows,
    identityRows,
    loadState: entityLoadState,
    refresh: loadEntities,
  } = useEntitySurfaceQuery({
    hostQueryState,
    identityQueryState,
    onIncidentAccessLost,
    viewQuery,
  });
  const [evidenceInvalidationGeneration, setEvidenceInvalidationGeneration] =
    useState(0);
  const [inspectorInvalidationGeneration, setInspectorInvalidationGeneration] =
    useState(0);
  const [
    continuityInvalidationGeneration,
    setContinuityInvalidationGeneration,
  ] = useState(0);
  const continuityInvalidation = useCallback(() => {
    setPendingGridFocusSurface(() => null);
    setContinuityInvalidationGeneration((current) => current + 1);
  }, [setPendingGridFocusSurface]);
  const evidenceInvalidation = useCallback(() => {
    setEvidenceInvalidationGeneration((current) => current + 1);
  }, []);
  const inspectorInvalidation = useCallback(() => {
    setInspectorInvalidationGeneration((current) => current + 1);
  }, []);
  const extensionInvalidation = useCallback(() => {
    extensionAvailability.invalidate();
    handleExtensionAvailabilityChange();
  }, [extensionAvailability, handleExtensionAvailabilityChange]);
  const queryInvalidation = useCallback(
    (reason: Parameters<typeof invalidateGenericQuery>[0]) => {
      invalidateGenericQuery(reason);
      invalidateAssessmentQuery(reason);
      invalidateEntityQuery(reason);
    },
    [invalidateAssessmentQuery, invalidateEntityQuery, invalidateGenericQuery],
  );
  const onAuthorizationRecovered = useCallback(
    (result: { readonly role: IncidentRole; readonly userId: string }) => {
      setCurrentUserId(result.userId || null);
      setCurrentIncidentRole(result.role);
    },
    [],
  );
  const collaborationProjection = useMemo(
    () =>
      new WorkbookCollaborationCoordinator({
        authorizationRecovery,
        continuityInvalidation,
        evidenceInvalidation,
        extensionInvalidation,
        incidentId,
        initialSheetRef: {
          kind: "view_schema",
          id: timelineViewSchemaId,
        },
        inspectorInvalidation,
        mutationRuntime,
        onAuthorizationRecovered,
        onIncidentAccessLost,
        queryInvalidation,
      }),
    [
      authorizationRecovery,
      continuityInvalidation,
      evidenceInvalidation,
      extensionInvalidation,
      incidentId,
      inspectorInvalidation,
      mutationRuntime,
      onAuthorizationRecovered,
      onIncidentAccessLost,
      queryInvalidation,
    ],
  );
  const collaborationSnapshot = useWorkbookCollaborationCoordinatorSession({
    projection: collaborationProjection,
    session: collaborationSession,
    sheetRef: startupSheetRef,
  });
  const activeLoaderSurfacePort =
    useMemo<WorkbookActiveSurfacePort | null>(() => {
      if (
        startupSheetRef.kind === "extension_workspace" ||
        surface === timelineViewSchemaId
      ) {
        return null;
      }
      const refresh =
        surface === assessmentsViewSchemaId
          ? loadAssessmentSurface
          : surface === hostsViewSchemaId || surface === identitiesViewSchemaId
            ? loadEntities
            : loadGenericSurface;
      const entitySurface =
        surface === hostsViewSchemaId || surface === identitiesViewSchemaId;
      const assessmentSurface = surface === assessmentsViewSchemaId;
      return {
        identity: {
          sheetRef: startupSheetRef,
          viewSchemaId: surface,
        },
        applyRecordChanged: (payload) =>
          entitySurface
            ? applyEntityRecordChanged(payload, surface)
            : assessmentSurface
              ? applyAssessmentRecordChanged(payload)
              : applyGenericRecordChanged(payload),
        invalidate: entitySurface
          ? invalidateEntityQuery
          : assessmentSurface
            ? invalidateAssessmentQuery
            : invalidateGenericQuery,
        refresh: async () => {
          await refresh();
        },
      };
    }, [
      applyAssessmentRecordChanged,
      applyEntityRecordChanged,
      invalidateAssessmentQuery,
      invalidateEntityQuery,
      invalidateGenericQuery,
      applyGenericRecordChanged,
      loadAssessmentSurface,
      loadEntities,
      loadGenericSurface,
      startupSheetRef,
      surface,
    ]);
  useEffect(() => {
    if (activeLoaderSurfacePort === null) return;
    return collaborationProjection.registerActiveSurface(
      activeLoaderSurfacePort,
    );
  }, [activeLoaderSurfacePort, collaborationProjection]);
  const interactionMode = workbookGridInteractionMode(
    incidentIdentity?.status,
    currentIncidentRole,
  );
  const workbookLayout = useWorkbookLayoutFacade({
    accountDensityMode,
    columnCommands: {
      onColumnHiddenChange: activeLayoutControls.onColumnHiddenChange,
      onColumnMove: activeLayoutControls.onColumnMove,
      onColumnReorder: activeLayoutControls.onColumnReorder,
      onColumnWidthChange: activeLayoutControls.onColumnWidthChange,
      onResetColumns: activeLayoutControls.onResetColumns,
    },
    columnState: activeLayoutState,
    interactionMode,
    viewSchemaId: surface,
  });
  const responsiveBand = workbookLayout.shell.chromeMode;
  useEffect(() => {
    if (account?.user_id) {
      setCurrentUserId(account.user_id);
    }
  }, [account?.user_id]);

  const networkFlowActivityAvailable = extensionAvailability.isRenderable({
    extensionProfileId: networkFlowActivityProfileId,
    workspaceKey: networkAnalysisWorkspaceKey,
  });

  useEffect(() => {
    if (!networkFlowActivityAvailable && networkAnalysisActive) {
      selectWorkbookSurface(timelineViewSchemaId);
    }
  }, [
    networkAnalysisActive,
    networkFlowActivityAvailable,
    selectWorkbookSurface,
  ]);

  const loadSessionRole = useCallback(async () => {
    const result = await authorizationRecovery.recover({
      incidentId,
      signal: new AbortController().signal,
    });
    if (result.kind !== "authorized") {
      setCurrentUserId(null);
      setCurrentIncidentRole("");
      onIncidentAccessLost?.();
      return;
    }
    setCurrentUserId(result.userId || null);
    setCurrentIncidentRole(result.role);
  }, [authorizationRecovery, incidentId, onIncidentAccessLost]);

  useWorkbookProjectionRefreshController({
    loadAssessmentSurface,
    loadEntities,
    loadGenericSurface,
    loadSessionRole,
    sheetReloadToken,
  });

  useWorkbookPendingGridFocus({
    pendingGridFocusSurface,
    setPendingGridFocusSurface,
    surface,
  });

  const activeSurfaceIsBuiltIn = requiredBuiltInWorkbookSurfaceIds.some(
    (viewSchemaId) => viewSchemaId === surface,
  );
  const activeSystemSurfaceTitle =
    activeSurfaceIsBuiltIn || networkAnalysisActive
      ? null
      : activeContract.title;
  const incidentKeyLabel = incidentIdentity?.incident_key ?? "Incident";
  const incidentTitleLabel = incidentIdentity?.title ?? "Loading incident";
  const accountDisplayName =
    account?.display_name ?? currentUserLabel ?? "Unknown user";
  const accountTitle = account
    ? `${account.display_name}${account.is_deployment_admin ? " (deployment administrator)" : ""}`
    : accountDisplayName;

  const activeSavedViewSelector = networkAnalysisActive ? null : (
    <ActiveSurfaceSavedViewSelector
      activeViewSchemaId={surface}
      chromeMode={responsiveBand}
      currentIncidentRole={currentIncidentRole}
      currentUserId={currentUserId}
      isModified={activeSavedViewModified}
      onCreateSavedView={createSavedView}
      onDeleteSavedView={deleteSavedView}
      onDuplicateSavedView={duplicateSavedView}
      onResetToSavedView={selectSavedView}
      onSelectBaseSurface={selectWorkbookSurface}
      onSelectSavedView={selectSavedView}
      onSetDefaultSheetRef={setWorkbookDefaultSheetRef}
      onSetHomeSheetRef={setWorkbookHomeSheetRef}
      onUpdateSavedView={updateSavedView}
      savedViews={savedViews}
      selectedSheetRef={startupSheetRef}
    />
  );
  const activeViewBarQueryControls =
    networkAnalysisActive ||
    responsiveBand === "below_supported_minimum" ? null : (
      <WorkbookGridControls
        chromeMode={responsiveBand}
        contract={activeQueryControls.contract}
        filterDraft={activeQueryControls.filterDraft}
        layoutState={activeLayoutState}
        onApplyFilter={activeQueryControls.onApplyFilter}
        onClearAll={activeQueryControls.onClearAll}
        onFilterDraftChange={activeQueryControls.onFilterDraftChange}
        onGroupByChange={activeQueryControls.onGroupByChange}
        onColumnHiddenChange={activeLayoutControls.onColumnHiddenChange}
        onColumnMove={activeLayoutControls.onColumnMove}
        onResetColumns={activeLayoutControls.onResetColumns}
        onRemoveFilter={activeQueryControls.onRemoveFilter}
        onSortChange={activeQueryControls.onSortChange}
        queryState={activeQueryControls.queryState}
        surface={activeQueryControls.surface}
      />
    );
  const workbookAccountApplicationMenu = accountApplicationMenu?.({
    currentIncidentRole,
    incidentControls: accountIncidentControls,
  });
  const incidentControlsDrawer =
    incidentControlsDrawerSection ===
    null ? null : incidentControlsDrawerSection === "import-assistant" &&
      importAssistantAvailable ? (
      <Suspense fallback={<p role="status">Loading import assistant…</p>}>
        <LazyImportAssistantFeature
          apiBase={apiBase}
          availability={extensionAvailability}
          currentIncidentRole={currentIncidentRole}
          incidentId={incidentId}
          onNavigateToView={(viewSchemaId) => {
            selectWorkbookSurface(viewSchemaId, {
              focusFirstGridTarget: true,
            });
            closeIncidentControlsDrawer({
              restoreTriggerFocus: false,
            });
          }}
        />
      </Suspense>
    ) : (
      (renderIncidentControls?.({
        activeSection: incidentControlsDrawerSection,
        apiBase,
        currentIncidentRole,
        incidentId,
        onIncidentAccessLost,
        onIncidentSnapshot,
        onSessionRoleChange: loadSessionRole,
      }) ?? null)
    );
  const inspectorResetKey = `${surface}:${sheetRefKey(startupSheetRef)}:${sheetReloadToken}:${inspectorInvalidationGeneration}:${evidenceInvalidationGeneration}`;
  const continuityResetKey = `${surface}:${sheetRefKey(startupSheetRef)}:${sheetReloadToken}:${continuityInvalidationGeneration}`;
  const activeExtensionWorkspace = (() => {
    if (startupSheetRef.kind !== "extension_workspace") {
      return null;
    }
    if (
      !extensionAvailability.isRenderable({
        extensionProfileId: startupSheetRef.extension_profile_id,
        workspaceKey: startupSheetRef.workspace_key,
      })
    ) {
      return (
        <p style={shellContentNoticeStyle}>
          This extension workspace is not currently available.
        </p>
      );
    }
    const renderer =
      extensionWorkspaceRenderers[
        extensionWorkspaceRegistryKey(
          startupSheetRef.extension_profile_id,
          startupSheetRef.workspace_key,
        )
      ];
    if (!renderer) {
      return (
        <p style={shellContentNoticeStyle}>
          This extension workspace is not available in this client.
        </p>
      );
    }
    const lifecycleKey = `${extensionAvailability.currentTag()?.epochId ?? "disabled"}:${extensionAvailabilityRevision}`;
    return (
      <ExtensionAvailabilityProvider
        controller={extensionAvailability}
        key={lifecycleKey}
      >
        {renderer({
          apiBase,
          currentIncidentRole,
          incidentId,
          onIncidentAccessLost,
        })}
      </ExtensionAvailabilityProvider>
    );
  })();
  const selectBaseWorkbookSurface = useCallback(
    (
      viewSchemaId: string,
      options: { readonly focusFirstGridTarget?: boolean } = {},
    ) => {
      selectWorkbookSurface(viewSchemaId, options);
    },
    [selectWorkbookSurface],
  );

  return (
    <section
      aria-label="Workbook shell"
      data-active-view-schema-id={surface}
      data-testid={workbookShellReadyTestId()}
      data-workbook-shell-id={workbookShellId}
      style={panelStyle}
    >
      <WorkbookShellSlotRegion
        slot="top-bar"
        style={{
          ...shellTopBarStyle,
          ...(responsiveBand === "below_supported_minimum"
            ? shellTopBarUnsupportedStyle
            : null),
        }}
        viewSchemaId={surface}
      >
        <div
          data-testid={workbookIncidentIdentityTestId()}
          style={shellIncidentIdentityStyle}
          title={
            incidentIdentity === null
              ? (incidentIdentityError ?? "Loading incident")
              : `${incidentIdentity.incident_key} ${incidentIdentity.title}`
          }
        >
          <strong style={shellTopBarValueStyle}>{incidentKeyLabel}</strong>
          <span style={shellIncidentTitleStyle}>{incidentTitleLabel}</span>
        </div>
        <span
          aria-hidden="true"
          data-testid={workbookResponsiveBandTestId()}
          data-workbook-block-mode={workbookLayout.shell.blockMode}
          data-workbook-responsive-band={responsiveBand}
          hidden
        />
        {responsiveBand === "base" ? (
          <nav aria-label="Built-in workbook surfaces" style={tabStripStyle}>
            {requiredBuiltInWorkbookSurfaceIds.map((viewSchemaID) => {
              const contract = requireViewContract(viewSchemaID);
              return (
                <button
                  aria-current={
                    !networkAnalysisActive && surface === viewSchemaID
                      ? "page"
                      : undefined
                  }
                  key={viewSchemaID}
                  data-testid={surfaceTabTestId(viewSchemaID)}
                  data-view-schema-id={viewSchemaID}
                  data-workbook-tab-index={String(
                    requiredBuiltInWorkbookSurfaceIds.indexOf(viewSchemaID),
                  )}
                  style={{
                    ...surfaceTabStyle,
                    ...(surface === viewSchemaID && !networkAnalysisActive
                      ? surfaceTabActiveStyle
                      : null),
                  }}
                  type="button"
                  onClick={() => {
                    selectBaseWorkbookSurface(viewSchemaID);
                  }}
                >
                  {contract.title}
                </button>
              );
            })}
          </nav>
        ) : (
          <div style={surfacesMenuFrameStyle}>
            <button
              aria-controls={
                surfacesMenuOpen ? workbookSurfacesMenuTestId() : undefined
              }
              aria-expanded={surfacesMenuOpen}
              aria-haspopup="menu"
              data-testid={workbookSurfacesMenuTriggerTestId()}
              ref={surfacesMenuTriggerRef}
              style={surfaceMenuTriggerStyle}
              type="button"
              onClick={() => {
                setSurfacesMenuOpen((current) => !current);
              }}
              onKeyDown={(event) => {
                if (event.key !== "ArrowDown") return;
                event.preventDefault();
                setSurfacesMenuOpen(true);
                window.requestAnimationFrame(() => {
                  document
                    .getElementById(workbookSurfacesMenuTestId())
                    ?.querySelector<HTMLElement>('[role="menuitemradio"]')
                    ?.focus({ preventScroll: true });
                });
              }}
            >
              Surfaces
            </button>
            {surfacesMenuOpen ? (
              <div
                data-testid={workbookSurfacesMenuTestId()}
                id={workbookSurfacesMenuTestId()}
                role="menu"
                style={surfacesMenuStyle}
                onKeyDown={(event) => {
                  if (event.key === "Escape") {
                    event.preventDefault();
                    event.stopPropagation();
                    setSurfacesMenuOpen(false);
                    surfacesMenuTriggerRef.current?.focus({
                      preventScroll: true,
                    });
                    return;
                  }
                  if (
                    event.key !== "ArrowDown" &&
                    event.key !== "ArrowUp" &&
                    event.key !== "Home" &&
                    event.key !== "End"
                  ) {
                    return;
                  }
                  const items = Array.from(
                    event.currentTarget.querySelectorAll<HTMLElement>(
                      '[role="menuitemradio"]',
                    ),
                  );
                  if (items.length === 0) return;
                  const activeIndex =
                    document.activeElement instanceof HTMLElement
                      ? items.indexOf(document.activeElement)
                      : -1;
                  let nextIndex = 0;
                  if (event.key === "End") {
                    nextIndex = items.length - 1;
                  } else if (event.key === "ArrowUp") {
                    nextIndex =
                      activeIndex <= 0 ? items.length - 1 : activeIndex - 1;
                  } else if (event.key === "ArrowDown") {
                    nextIndex =
                      activeIndex < 0 || activeIndex === items.length - 1
                        ? 0
                        : activeIndex + 1;
                  }
                  event.preventDefault();
                  items[nextIndex]?.focus({ preventScroll: true });
                }}
              >
                {requiredBuiltInWorkbookSurfaceIds.map((viewSchemaID) => {
                  const contract = requireViewContract(viewSchemaID);
                  const isSelected =
                    !networkAnalysisActive && surface === viewSchemaID;
                  return (
                    <button
                      key={viewSchemaID}
                      aria-checked={isSelected}
                      data-testid={workbookSurfacesMenuOptionTestId(
                        viewSchemaID,
                      )}
                      data-view-schema-id={viewSchemaID}
                      role="menuitemradio"
                      style={{
                        ...surfacesMenuItemStyle,
                        ...(isSelected ? surfacesMenuItemSelectedStyle : null),
                      }}
                      type="button"
                      onClick={() => {
                        setSurfacesMenuOpen(false);
                        selectBaseWorkbookSurface(viewSchemaID);
                      }}
                    >
                      {contract.title}
                    </button>
                  );
                })}
              </div>
            ) : null}
          </div>
        )}
        <div style={systemViewSlotStyle}>
          {networkFlowActivityAvailable ? (
            <button
              aria-current={networkAnalysisActive ? "page" : undefined}
              data-testid={networkAnalysisTestId("tab")}
              style={{
                ...surfaceTabStyle,
                ...(networkAnalysisActive ? surfaceTabActiveStyle : null),
              }}
              type="button"
              onClick={() => {
                if (networkAnalysisRef.kind === "extension_workspace") {
                  selectExtensionWorkspace(networkAnalysisRef);
                }
              }}
            >
              Network Analysis
            </button>
          ) : null}
          <SystemViewSwitcher
            activeViewSchemaId={surface}
            onSelect={(viewSchemaId) => {
              selectWorkbookSurface(viewSchemaId, {
                focusFirstGridTarget: true,
              });
            }}
          />
          {activeSystemSurfaceTitle ? (
            <span style={activeSystemViewTitleStyle}>
              {activeSystemSurfaceTitle}
            </span>
          ) : null}
        </div>
        <div style={shellTopBarActionsStyle}>
          {responsiveBand === "base" || responsiveBand === "narrow_desktop" ? (
            <WorkbookPresenceSummary
              records={collaborationSnapshot.activeSheetPresenceRecords}
            />
          ) : null}
          <div style={currentUserSlotStyle}>
            {workbookAccountApplicationMenu ?? (
              <span style={currentUserChipStyle} title={accountTitle}>
                {displayInitials(accountDisplayName)}
              </span>
            )}
          </div>
        </div>
      </WorkbookShellSlotRegion>

      <div style={shellContentRegionStyle}>
        <div style={shellActiveSurfaceStyle}>
          <div
            aria-hidden={
              activeExtensionWorkspace === null &&
              mutationSnapshot.conflictPanelOpen
                ? true
                : undefined
            }
            inert={
              activeExtensionWorkspace === null &&
              mutationSnapshot.conflictPanelOpen
                ? true
                : undefined
            }
            style={{ display: "contents" }}
          >
            {activeExtensionWorkspace ?? (
              <WorkbookSurfacesFacade
                collaboration={{ projection: collaborationProjection }}
                continuity={{ resetKey: continuityResetKey }}
                incident={{
                  apiBase,
                  currentIncidentRole,
                  currentUserId,
                  incidentPort,
                  incidentId,
                  onIncidentAccessLost,
                }}
                inspector={{ resetKey: inspectorResetKey }}
                layout={workbookLayout.surface}
                mutations={{
                  commands: mutationCommands,
                  pending: pendingMutationPort,
                  runtime: mutationRuntime,
                }}
                queries={{
                  assessment: {
                    loadState: assessmentLoadState,
                    refresh: loadAssessmentSurface,
                    rows: assessmentRows,
                    setState: setAssessmentQueryState,
                    state: assessmentQueryState,
                  },
                  entities: {
                    hosts: {
                      rows: hostRows,
                      setState: setHostQueryState,
                      state: hostQueryState,
                    },
                    identities: {
                      rows: identityRows,
                      setState: setIdentityQueryState,
                      state: identityQueryState,
                    },
                    index: entityIndex,
                    loadState: entityLoadState,
                    refresh: loadEntities,
                  },
                  generic: {
                    loadState: genericLoadState,
                    refresh: loadGenericSurface,
                    rows: genericRows,
                    setState: setGenericQueryState,
                    state: genericQueryState,
                  },
                  referenceBroker: referenceQueryBroker,
                  viewQuery,
                  timeline: {
                    setState: setTimelineQueryState,
                    state: timelineQueryState,
                  },
                }}
                viewState={{
                  activeContract,
                  queryControls: activeViewBarQueryControls,
                  savedViewSelector: activeSavedViewSelector,
                  sheetRef: startupSheetRef,
                  sheetReloadToken,
                  surface,
                }}
              />
            )}
          </div>
          {activeExtensionWorkspace === null ? (
            <WorkbookConflictResolver
              apiBase={apiBase}
              mutationRuntime={mutationRuntime}
              onActivateOrigin={selectBaseWorkbookSurface}
            />
          ) : null}
        </div>

        {incidentControlsDrawerSection !== null ? (
          <IncidentControlsDrawer
            activeMenuItem={activeIncidentControlsMenuItem}
            closeButtonRef={incidentControlsCloseButtonRef}
            onClose={closeIncidentControlsDrawer}
          >
            {incidentControlsDrawer}
          </IncidentControlsDrawer>
        ) : null}
      </div>
    </section>
  );
}

export function WorkbookShell(props: WorkbookShellProps) {
  return (
    <IncidentCollaborationSession
      apiBase={props.apiBase}
      incidentId={props.incidentId}
      initialPresence={{
        sheet_ref: { kind: "view_schema", id: timelineViewSchemaId },
        mode: "viewing",
      }}
    >
      <WorkbookShellContent {...props} />
    </IncidentCollaborationSession>
  );
}
