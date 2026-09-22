import { workbookShellReadyTestId } from "@cartulary/ui-contracts";
import {
  type ReactNode,
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import type { NetworkFlowImportSurfaceBinding } from "../app/useNetworkFlowImport";
import type { WorkbookImportSurfaceBinding } from "../app/useWorkbookImport";
import {
  IncidentCollaborationSession,
  useIncidentCollaborationSession,
} from "../collaboration/IncidentCollaborationSession";
import type { ExtensionDiscoveryProfile } from "../extensions/extensionAvailability";
import {
  importProfileId,
  importRouteFamily,
  networkAnalysisSheetRef,
  networkAnalysisWorkspaceKey,
  networkFlowActivityProfileId,
  networkFlowRouteFamily,
} from "../extensions/extensionWorkspaceIdentities";
import type { WorkbookImportController } from "../imports/WorkbookImportController";
import type { AuthorizationRecoveryPort } from "../shared/authorizationRecovery";
import type { IncidentResource } from "../shared/incidentResource";
import { WorkbookRecoveryBoundary } from "../shared/WorkbookRecoveryBoundary";
import { WorkbookWorkAreaOverlayProvider } from "../shared/WorkbookWorkAreaOverlay";
import { WorkbookRecoveryNavigation } from "../shared/workbookRecoveryNavigation";
import type {
  WorkbookAccountApplicationMenuProps,
  WorkbookAccountModel,
  WorkbookIncidentControlsRendererProps,
} from "../shared/workbookShellContracts";
import { WorkbookActiveSurfaceFrame } from "./components/WorkbookActiveSurfaceFrame";
import { WorkbookActiveSurfacePresentation } from "./components/WorkbookActiveSurfacePresentation";
import { WorkbookBatchRecovery } from "./components/WorkbookBatchRecovery";
import { WorkbookIncidentControlsPresentation } from "./components/WorkbookIncidentControlsPresentation";
import {
  WorkbookRecoveryEntry,
  WorkbookRecoveryPanel,
} from "./components/WorkbookRecoveryPanel";
import { WorkbookReferenceContext } from "./components/WorkbookReferenceControl";
import { WorkbookSaveAnnouncements } from "./components/WorkbookSaveAnnouncements";
import { workbookShellId } from "./components/WorkbookShellSlots";
import { WorkbookShellTopBar } from "./components/WorkbookShellTopBar";
import { workbookShellViewBarWorkingSet } from "./components/WorkbookShellViewBarControls";
import { WorkbookObservedStatusStrip } from "./components/WorkbookStatusStrip";
import { WorkbookSurfaceRefreshNotice } from "./components/WorkbookSurfaceRefreshNotice";
import { attachWorkbookMutationPresentation } from "./composition/attachWorkbookMutationPresentation";
import { createWorkbookMutationInfrastructure } from "./composition/createWorkbookMutationInfrastructure";
import { WorkbookMutationRuntimeBoundary } from "./composition/WorkbookMutationRuntimeBoundary";
import { AssessmentAppendRecovery } from "./features/assessments/AssessmentAppendRecovery";
import { ContextualCreateContext } from "./features/coordination/ContextualCreateContext";
import { ContextualCreateRecovery } from "./features/coordination/ContextualCreateRecovery";
import { CoordinationCreateContext } from "./features/coordination/CoordinationCreateContext";
import { CoordinationCreateRecovery } from "./features/coordination/CoordinationCreateRecovery";
import { DecisionSupersessionContext } from "./features/coordination/DecisionSupersessionContext";
import { WorkbookDecisionSupersessionRecovery } from "./features/coordination/WorkbookDecisionSupersessionRecovery";
import { WorkbookEntityMergeRecovery } from "./features/entities/WorkbookEntityMergeRecovery";
import {
  EvidenceAttachmentContext,
  TimelineFileContext,
} from "./features/evidence/EvidenceAttachmentContext";
import { TimelineRelatedEvidenceContext } from "./features/evidence/TimelineRelatedEvidenceContext";
import { TimelineRelatedEvidenceRecovery } from "./features/evidence/TimelineRelatedEvidenceRecovery";
import { IndicatorCreateContext } from "./features/indicators/IndicatorCreateContext";
import { IndicatorLifecycleContext } from "./features/indicators/IndicatorLifecycleContext";
import { ObservationContext } from "./features/indicators/ObservationContext";
import { WorkbookIndicatorCreateRecovery } from "./features/indicators/WorkbookIndicatorCreateRecovery";
import { WorkbookIndicatorLifecycleRecovery } from "./features/indicators/WorkbookIndicatorLifecycleRecovery";
import { WorkbookObservationRecovery } from "./features/indicators/WorkbookObservationRecovery";
import {
  type NetworkFlowImportController,
  NetworkFlowImportSurface,
  NetworkFlowIndicatorLinkSurface,
  NetworkFlowTableSurface,
  useNetworkFlowIndicatorLinkOwner,
  useNetworkFlowSavedGraphOwner,
  useNetworkFlowTableOwner,
} from "./features/NetworkFlowOperations";
import { NoteAssociationRecovery } from "./features/notes/NoteAssociationRecovery";
import { NoteCreateContext } from "./features/notes/NoteCreateContext";
import { NoteCreateRecovery } from "./features/notes/NoteCreateRecovery";
import { PartyLinkRecovery } from "./features/parties/PartyLinkRecovery";
import { WorkbookHistoryContext } from "./history/WorkbookHistoryContext";
import { WorkbookHistoryRecovery } from "./history/WorkbookHistoryRecovery";
import { useIncidentControlsDrawer } from "./hooks/useIncidentControlsDrawer";
import { useNetworkFlowImportBinding } from "./hooks/useNetworkFlowImportBinding";
import { useWorkbookAuthorizationState } from "./hooks/useWorkbookAuthorizationState";
import { WorkbookCandidateAuthorityContext } from "./hooks/useWorkbookCandidateDiscovery";
import { useWorkbookCollaborationLifecycle } from "./hooks/useWorkbookCollaborationLifecycle";
import {
  useWorkbookExtensionAvailability,
  useWorkbookExtensionFallback,
} from "./hooks/useWorkbookExtensionAvailability";
import { useWorkbookImportBinding } from "./hooks/useWorkbookImportBinding";
import { useWorkbookIncidentIdentity } from "./hooks/useWorkbookIncidentIdentity";
import { useWorkbookProjectionRefreshController } from "./hooks/useWorkbookProjectionRefreshController";
import { useWorkbookRecoveryFocus } from "./hooks/useWorkbookRecoveryFocus";
import {
  useWorkbookReferenceReader,
  useWorkbookShellInfrastructure,
} from "./hooks/useWorkbookShellInfrastructure";
import { useWorkbookSurfaceQueries } from "./hooks/useWorkbookSurfaceQueries";
import { useWorkbookLayoutFacade } from "./layout/useWorkbookLayoutFacade";
import type { AccountDensityMode } from "./layout/workbookDensity";
import {
  panelStyle,
  shellContentRegionStyle,
} from "./layout/workbookShellStyles";
import { workbookGridInteractionMode } from "./models/workbookGridState";
import type { WorkbookIncidentIdentity } from "./models/workbookIncidentIdentity";
import {
  isNetworkAnalysisSheetRef,
  workbookAccountPresentation,
  workbookActiveSystemSurfaceTitle,
} from "./models/workbookShellPresentation";
import { timelineViewSchemaId } from "./models/workbookSurfaceRegistry";
import type { WorkbookPreferenceController } from "./preferences/WorkbookPreferenceController";
import {
  useWorkbookPreferencesSnapshot,
  WorkbookPreferenceAnnouncements,
} from "./preferences/WorkbookPreferencesPanel";
import type { PreferenceWorkbookBinding } from "./preferences/workbookPreferenceModel";
import { WorkbookQueryBrowsingProvider } from "./query/WorkbookQueryBrowsingContext";
import type { WorkbookMutationRuntime } from "./runtime/WorkbookMutationRuntime";
import type { WorkbookMutationRuntimeRegistry } from "./runtime/WorkbookMutationRuntimeRegistry";
import type { SavedViewBinding } from "./savedviews/savedViewOperationModel";
import type { WorkbookSavedViewController } from "./savedviews/WorkbookSavedViewController";
import type { WorkbookSurfacesFacadeProps } from "./surfaces/WorkbookSurfacesFacade";
import { TimelineCaptureRecovery } from "./timeline/actions/TimelineCaptureRecovery";
import { TimelineMentionRecovery } from "./timeline/actions/TimelineMentionRecovery";

export type {
  WorkbookAccountApplicationMenuProps,
  WorkbookIncidentControlsRendererProps,
};

type WorkbookShellProps = {
  sessionIdentity: string | null;
  networkFlowImportController: NetworkFlowImportController;
  bindNetworkFlowImport: (
    binding: NetworkFlowImportSurfaceBinding | null,
  ) => void;
  importController: WorkbookImportController;
  bindWorkbookImport: (binding: WorkbookImportSurfaceBinding | null) => void;
  savedViewController: WorkbookSavedViewController;
  bindWorkbookSavedViews: (binding: SavedViewBinding | null) => void;
  preferenceController?: WorkbookPreferenceController | undefined;
  bindWorkbookPreferences?:
    | ((binding: PreferenceWorkbookBinding | null) => void)
    | undefined;
  onIncidentControlsSectionChange?:
    | ((
        section: WorkbookIncidentControlsRendererProps["activeSection"] | null,
      ) => void)
    | undefined;
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
  acceptedIncidentResource?: IncidentResource | null | undefined;
  onIncidentResourceObserved?:
    | ((resource: IncidentResource) => void)
    | undefined;
  extensionProfiles?: readonly ExtensionDiscoveryProfile[] | null | undefined;
  onSessionLost?: (() => void) | undefined;
  onIncidentAccessLost?: (() => void) | undefined;
  renderIncidentControls?:
    | ((props: WorkbookIncidentControlsRendererProps) => ReactNode)
    | undefined;
  mutationRuntimeRegistry: WorkbookMutationRuntimeRegistry;
};

type WorkbookShellContentProps = WorkbookShellProps & {
  mutationRuntime: WorkbookMutationRuntime;
};

const noExtensionProfiles: readonly ExtensionDiscoveryProfile[] = [];

function WorkbookShellContent({
  sessionIdentity,
  networkFlowImportController,
  bindNetworkFlowImport,
  importController,
  bindWorkbookImport,
  savedViewController,
  bindWorkbookSavedViews,
  preferenceController,
  bindWorkbookPreferences,
  onIncidentControlsSectionChange,
  authorizationRecovery,
  incidentId,
  apiBase,
  account,
  accountDensityMode,
  accountApplicationMenu,
  currentUserLabel,
  initialIncidentIdentity,
  acceptedIncidentResource,
  onIncidentResourceObserved,
  extensionProfiles = noExtensionProfiles,
  onIncidentAccessLost,
  onSessionLost,
  renderIncidentControls,
  mutationRuntime,
}: WorkbookShellContentProps) {
  const collaborationSession = useIncidentCollaborationSession();
  const extensionLifecycle = useWorkbookExtensionAvailability({
    clientInstanceId: collaborationSession.clientInstanceId,
    incidentId,
    profiles: extensionProfiles,
  });
  const authorization = useWorkbookAuthorizationState({
    onSessionLost,
    accountUserId: account?.user_id,
    authorizationRecovery,
    incidentId,
    onIncidentAccessLost,
  });
  const infrastructure = useWorkbookShellInfrastructure({
    acceptedAuthority: authorization.acceptedAuthority,
    sessionIdentity,
    savedViewOwner: savedViewController,
    bindWorkbookSavedViews,
    authorizationRecovered: authorization.acceptRecoveredAuthorization,
    apiBase,
    extensionAvailability: extensionLifecycle.controller,
    incidentId,
    mutationRuntime,
    onExtensionAvailabilityChange: extensionLifecycle.publishChange,
    onAuthorityUncertain: authorization.loadSessionRole,
  });
  const { commands, snapshot } = infrastructure.workbookRuntime;
  const referenceReader = useWorkbookReferenceReader(
    authorization.authorizationGeneration,
    infrastructure.viewQuery,
    apiBase,
    incidentId,
  );
  const { incidentIdentity, incidentIdentityError, acceptIncidentResource } =
    useWorkbookIncidentIdentity({
      collaborationSession,
      incidentPort: infrastructure.incidentPort,
      incidentId,
      initialIncidentIdentity,
      acceptedIncidentResource,
      onIncidentResourceObserved,
      onAuthorityUncertain: authorization.loadSessionRole,
    });
  useLayoutEffect(() => {
    const mergeAuthority =
      authorization.acceptedAuthority.userId &&
      authorization.acceptedAuthority.role !== null &&
      sessionIdentity !== null
        ? {
            actorId: authorization.acceptedAuthority.userId,
            sessionIdentity,
            incidentId,
            role: authorization.acceptedAuthority.role,
            closed: incidentIdentity?.status !== "active",
          }
        : null;
    infrastructure.mutationRuntime.setAuthority(mergeAuthority);
  }, [
    infrastructure.mutationRuntime,
    authorization.acceptedAuthority,
    incidentId,
    incidentIdentity?.status,
    sessionIdentity,
  ]);
  const networkFlowSavedGraphController = useNetworkFlowSavedGraphOwner({
    availability: extensionLifecycle.controller,
    apiBase,
    incidentId,
    actorId: authorization.currentUserId,
    sessionIdentity,
    role: authorization.currentIncidentRole,
    open: incidentIdentity?.status === "active",
  });
  const networkFlowIndicatorLinkController = useNetworkFlowIndicatorLinkOwner({
    availability: extensionLifecycle.controller,
    apiBase,
    incidentId,
    actorId: authorization.currentUserId,
    sessionIdentity,
    role: authorization.currentIncidentRole,
    open: incidentIdentity?.status === "active",
  });
  const networkFlowTableController = useNetworkFlowTableOwner({
    availability: extensionLifecycle.controller,
    apiBase,
    incidentId,
    actorId: authorization.currentUserId,
    sessionIdentity,
    role: authorization.currentIncidentRole,
    open: incidentIdentity?.status === "active",
    onMutationAdmitted: networkFlowIndicatorLinkController.onMutationAdmitted,
    onLocalChange: (change) => {
      if (
        change.resourceKind === "*" &&
        change.reasonCode === "authorization_lost"
      )
        void authorization.loadSessionRole();
      void networkFlowSavedGraphController.onResourceChange(change);
      networkFlowIndicatorLinkController.onResourceChange(change);
    },
  });
  const incidentRead = useRef<AbortController | null>(null);
  // biome-ignore lint/correctness/useExhaustiveDependencies: An incident replacement must abort the previous incident's resource read.
  useLayoutEffect(() => () => incidentRead.current?.abort(), [incidentId]);
  const recoverImportIncident = useCallback(async () => {
    incidentRead.current?.abort();
    const stop = new AbortController();
    incidentRead.current = stop;
    const result = await infrastructure.incidentPort.getIdentity({
      signal: stop.signal,
    });
    if (!stop.signal.aborted && result.kind === "accepted") {
      acceptIncidentResource(result.value);
      if (result.value.resource)
        onIncidentResourceObserved?.(result.value.resource);
    }
  }, [
    infrastructure.incidentPort,
    acceptIncidentResource,
    onIncidentResourceObserved,
  ]);
  const queries = useWorkbookSurfaceQueries({
    ordinaryCreateOwner: infrastructure.mutationRuntime.ordinaryCreate,
    explicitPatchOwner: infrastructure.mutationRuntime.explicitPatches,
    decisionOwner: infrastructure.mutationRuntime.decisionSupersession,
    indicatorOwner: infrastructure.mutationRuntime.indicatorRecords,
    assessmentOwner: infrastructure.mutationRuntime.assessmentAuthoring,
    activeContract: snapshot.activeContract,
    assessment: {
      setState: commands.setAssessmentQueryState,
      state: snapshot.assessmentQueryState,
    },
    generic: {
      setState: commands.setGenericQueryState,
      state: snapshot.genericQueryState,
    },
    hosts: {
      setState: commands.setHostQueryState,
      state: snapshot.hostQueryState,
    },
    identities: {
      setState: commands.setIdentityQueryState,
      state: snapshot.identityQueryState,
    },
    onAuthorityUncertain: authorization.loadSessionRole,
    sheetRef: snapshot.startupSheetRef,
    surface: snapshot.surface,
    timeline: {
      setState: commands.setTimelineQueryState,
      state: snapshot.timelineQueryState,
    },
    viewQuery: infrastructure.viewQuery,
  });
  useLayoutEffect(() => {
    const history = infrastructure.mutationRuntime.history;
    const surfaces = queries.facadeQueries;
    for (const row of [...surfaces.generic.rows, ...surfaces.assessment.rows])
      if (history.latestVersion(row.record_id) !== null)
        history.acceptVersion(row.record_id, row.row_version);
    for (const row of [
      ...surfaces.entities.hosts.rows,
      ...surfaces.entities.identities.rows,
    ]) {
      if (history.latestVersion(row.recordId) !== null)
        history.acceptVersion(row.recordId, row.rowVersion);
      const merge = infrastructure.mutationRuntime.entityMerge;
      if (merge.latestVersion(row.recordId) !== null)
        merge.acceptVersion(row.recordId, row.rowVersion);
    }
  }, [infrastructure.mutationRuntime, queries.facadeQueries]);
  useLayoutEffect(
    () =>
      attachWorkbookMutationPresentation({
        runtime: infrastructure.mutationRuntime,
        apiBase,
        actorId: authorization.currentUserId,
        sessionIdentity,
        surface: snapshot.surface,
        extensionWorkspace:
          snapshot.startupSheetRef.kind === "extension_workspace",
        refresh: {
          entities: queries.refreshProjection.entities,
          assessment: queries.refreshProjection.assessment,
          generic: queries.refreshProjection.generic,
        },
        authorityUncertain: authorization.loadSessionRole,
        authorizationRecovered: authorization.acceptRecoveredAuthorization,
      }),
    [
      infrastructure.mutationRuntime,
      apiBase,
      authorization.currentUserId,
      sessionIdentity,
      snapshot.surface,
      snapshot.startupSheetRef.kind,
      queries.refreshProjection.entities,
      queries.refreshProjection.assessment,
      queries.refreshProjection.generic,
      authorization.loadSessionRole,
      authorization.acceptRecoveredAuthorization,
    ],
  );
  const collaboration = useWorkbookCollaborationLifecycle({
    onSessionLost,
    activeSurfacePort: queries.activeSurfacePort,
    authorizationRecovery,
    cancelGridEntryFocus: commands.cancelGridEntryFocus,
    collaborationSession,
    extensionInvalidation: extensionLifecycle.invalidate,
    incidentId,
    mutationRuntime: infrastructure.mutationRuntime,
    onAuthorizationRecovered: authorization.acceptRecoveredAuthorization,
    onIncidentAccessLost,
    queryInvalidation: queries.invalidateAll,
    sheetRef: snapshot.startupSheetRef,
    sheetReloadToken: snapshot.sheetReloadToken,
    surface: snapshot.surface,
  });
  useEffect(() => {
    if (!incidentIdentity) return;
    if (incidentIdentity?.status === "closed")
      infrastructure.mutationRuntime.invalidate({ kind: "incident_closed" });
    else if (incidentIdentity?.status === "active")
      infrastructure.mutationRuntime.observeIncidentReopened();
  }, [incidentIdentity, infrastructure.mutationRuntime]);
  const interactionMode = workbookGridInteractionMode(
    incidentIdentity?.status,
    authorization.currentIncidentRole,
  );
  const workbookLayout = useWorkbookLayoutFacade({
    accountDensityMode,
    columnCommands: {
      onColumnHiddenChange: snapshot.activeLayoutControls.onColumnHiddenChange,
      onColumnMove: snapshot.activeLayoutControls.onColumnMove,
      onColumnReorder: snapshot.activeLayoutControls.onColumnReorder,
      onColumnSizingIntent: snapshot.activeLayoutControls.onColumnSizingIntent,
      bindColumnSizing: snapshot.activeLayoutControls.bindColumnSizing,
      sizing: snapshot.activeLayoutControls.sizing,
      freezing: snapshot.activeLayoutControls.freezing,
      onResetColumns: snapshot.activeLayoutControls.onResetColumns,
    },
    columnState: snapshot.activeLayoutState,
    incidentClosed: incidentIdentity?.status === "closed",
    interactionMode,
    viewSchemaId: snapshot.surface,
  });
  useWorkbookProjectionRefreshController({
    loadAssessmentSurface: queries.refreshProjection.assessment,
    loadEntities: queries.refreshProjection.entities,
    loadGenericSurface: queries.refreshProjection.generic,
    loadSessionRole: authorization.loadSessionRole,
    sheetReloadToken: snapshot.sheetReloadToken,
  });

  const networkAnalysisRef = useMemo(networkAnalysisSheetRef, []);
  const networkAnalysisActive = isNetworkAnalysisSheetRef(
    snapshot.startupSheetRef,
  );
  const networkAnalysisAvailable =
    incidentIdentity?.status !== "closed" &&
    extensionLifecycle.controller.isRenderable({
      extensionProfileId: networkFlowActivityProfileId,
      workspaceKey: networkAnalysisWorkspaceKey,
    });
  const preferenceState = useWorkbookPreferencesSnapshot(preferenceController);
  const inspectedHome = preferenceState?.inspectionActive
    ? preferenceState.home.resource?.home_sheet_ref
    : null;
  const inspectedDefault = preferenceState?.inspectionActive
    ? preferenceState.default.resource?.default_sheet_ref
    : null;
  const homeId = inspectedHome?.kind === "saved_view" ? inspectedHome.id : null;
  const defaultId =
    inspectedDefault?.kind === "saved_view" ? inspectedDefault.id : null;
  useLayoutEffect(() => {
    savedViewController.observePreference("home", homeId);
    savedViewController.observePreference("default", defaultId);
    return () => {
      savedViewController.observePreference("home", null);
      savedViewController.observePreference("default", null);
    };
  }, [savedViewController, homeId, defaultId]);
  const preferenceBinding = useRef(bindWorkbookPreferences);
  preferenceBinding.current = bindWorkbookPreferences;
  useLayoutEffect(() => {
    const selected = snapshot.startupSheetRef;
    const saved = snapshot.savedViewsResource.selectedSavedView;
    preferenceBinding.current?.({
      incidentId,
      actorId: authorization.currentUserId,
      apiBase,
      onAuthorizationRecovered: authorization.acceptRecoveredAuthorization,
      surface: {
        savedViewLabels: [
          ...savedViewController.getSnapshot().observations.values(),
        ].flatMap((observation) =>
          observation.resource
            ? [
                {
                  id: observation.resource.saved_view_id,
                  label: observation.resource.display_name,
                },
              ]
            : [],
        ),
        sheetRef: selected,
        label: networkAnalysisActive
          ? "Network Analysis"
          : (saved?.display_name ?? null),
        available:
          !snapshot.startupPending &&
          authorization.currentUserId !== null &&
          authorization.currentIncidentRole !== null &&
          authorization.currentIncidentRole !== "" &&
          (selected.kind === "extension_workspace"
            ? networkAnalysisActive && networkAnalysisAvailable
            : selected.kind !== "saved_view" ||
              (saved !== null && saved !== undefined)),
      },
    });
  });
  useLayoutEffect(() => () => preferenceBinding.current?.(null), []);
  const selectTimelineFallback = useCallback(() => {
    commands.selectWorkbookSurface(timelineViewSchemaId);
  }, [commands.selectWorkbookSurface]);
  useWorkbookExtensionFallback({
    active: networkAnalysisActive && !snapshot.startupPending,
    available: networkAnalysisAvailable,
    onFallback: selectTimelineFallback,
  });

  const selectBaseWorkbookSurface = useCallback(
    (
      viewSchemaId: string,
      options: { readonly focusFirstGridTarget?: boolean } = {},
    ) => {
      commands.selectWorkbookSurface(viewSchemaId, options);
    },
    [commands.selectWorkbookSurface],
  );
  const activeSurfaceFocusRef = useRef<HTMLElement | null>(null);
  const recoveryNavigation = useMemo(() => {
    void infrastructure.mutationRuntime;
    return new WorkbookRecoveryNavigation();
  }, [infrastructure.mutationRuntime]);
  const recoveryInvokerRef = useRef<HTMLElement | null>(null);
  const [recoveryDetailHost, setRecoveryDetailHost] =
    useState<HTMLDivElement | null>(null);
  useEffect(() => () => recoveryNavigation.dispose(), [recoveryNavigation]);
  const recoverySheetKey = JSON.stringify(snapshot.startupSheetRef);
  useLayoutEffect(() => {
    void recoverySheetKey;
    recoveryNavigation.close();
  }, [recoveryNavigation, recoverySheetKey]);
  const recoveryFocus = useWorkbookRecoveryFocus({
    activeSurfaceRef: activeSurfaceFocusRef,
    runtime: infrastructure.mutationRuntime,
    onSessionRecovery: authorization.loadSessionRole,
    navigation: recoveryNavigation,
    invokerRef: recoveryInvokerRef,
  });
  const importAssistantAvailable =
    extensionLifecycle.controller.isRouteAvailable(
      importProfileId,
      importRouteFamily,
    );
  useWorkbookImportBinding({
    controller: importController,
    bind: bindWorkbookImport,
    incidentId,
    apiBase,
    availability: extensionLifecycle.controller,
    available: importAssistantAvailable,
    role: authorization.currentIncidentRole,
    closed: incidentIdentity?.status !== "active",
    recoverAccess: authorization.loadSessionRole,
  });
  const incidentControls = useIncidentControlsDrawer(
    importAssistantAvailable,
    onIncidentControlsSectionChange,
  );
  useNetworkFlowImportBinding({
    controller: networkFlowImportController,
    bind: bindNetworkFlowImport,
    incidentId,
    apiBase,
    availability: extensionLifecycle.controller,
    available:
      importAssistantAvailable &&
      extensionLifecycle.controller.isRouteAvailable(
        networkFlowActivityProfileId,
        networkFlowRouteFamily,
      ),
    role: authorization.currentIncidentRole,
    closed: incidentIdentity?.status !== "active",
    lifecycleVersion: incidentIdentity?.incident_version,
    recoverIncident: recoverImportIncident,
    recoverAccess: authorization.loadSessionRole,
  });
  const accountApplication = accountApplicationMenu?.({
    currentIncidentRole: authorization.currentIncidentRole,
    incidentControls: incidentControls.accountIncidentControls,
  });
  const accountPresentation = workbookAccountPresentation(
    account,
    currentUserLabel,
  );
  const activeSystemSurfaceTitle = workbookActiveSystemSurfaceTitle(
    snapshot.surface,
    snapshot.activeContract.title,
    networkAnalysisActive,
  );
  const viewBarWorkingSet = workbookShellViewBarWorkingSet({
    chromeMode: workbookLayout.shell.chromeMode,
    currentIncidentRole: authorization.currentIncidentRole,
    currentUserId: authorization.currentUserId,
    incidentId,
    networkAnalysisActive,
    runtime: infrastructure.workbookRuntime,
    preferenceController,
    onInspectPreferences: (target) =>
      incidentControls.accountIncidentControls.onSelectSection(
        "summary",
        target,
      ),
  });
  const facadeProps: WorkbookSurfacesFacadeProps = {
    collaboration: { projection: collaboration.projection },
    continuity: { resetKey: collaboration.continuityResetKey },
    gridEntryFocus: {
      acknowledge: commands.acknowledgeGridEntryFocus,
      cancel: commands.cancelGridEntryFocus,
      request: snapshot.gridEntryFocusRequest,
    },
    incident: {
      apiBase,
      currentIncidentRole: authorization.currentIncidentRole,
      currentUserId: authorization.currentUserId,
      incidentPort: infrastructure.incidentPort,
      incidentId,
      onAuthorityUncertain: authorization.loadSessionRole,
    },
    inspector: { resetKey: collaboration.inspectorResetKey },
    layout: workbookLayout.surface,
    mutations: {
      clipboardPaste: infrastructure.clipboardPastePort,
      commands: infrastructure.mutationCommands,
      onActivateConflict: recoveryFocus.activate,
      runtime: infrastructure.mutationRuntime,
    },
    queries: queries.facadeQueries,
    viewState: {
      activeContract: snapshot.activeContract,
      viewBarWorkingSet,
      sheetRef: snapshot.startupSheetRef,
      sheetReloadToken: snapshot.sheetReloadToken,
      surface: snapshot.surface,
    },
  };
  const activeContent = (
    <WorkbookActiveSurfacePresentation
      extension={{
        availability: extensionLifecycle.controller,
        revision: extensionLifecycle.revision,
      }}
      extensionRenderer={{
        tableController: networkFlowTableController,
        savedGraphController: networkFlowSavedGraphController,
        indicatorLinkController: networkFlowIndicatorLinkController,
        importController: networkFlowImportController,
        currentUserId: authorization.currentUserId,
        workbookStatus: (
          <WorkbookObservedStatusStrip
            source={infrastructure.mutationRuntime.statusSource}
            sheetRef={snapshot.startupSheetRef}
            chromeMode={workbookLayout.shell.chromeMode}
            showPresence={false}
            onActivateConflict={recoveryFocus.activate}
            workbookFocusAnchor={null}
          />
        ),
        apiBase,
        currentIncidentRole: authorization.currentIncidentRole,
        incidentId,
        onAuthorityUncertain: authorization.loadSessionRole,
      }}
      sheetRef={snapshot.startupSheetRef}
      surface={facadeProps}
    />
  );

  return (
    <WorkbookCandidateAuthorityContext.Provider
      value={{
        identity: JSON.stringify([
          account?.user_id,
          sessionIdentity,
          incidentId,
          authorization.authorizationGeneration,
        ]),
        canRead:
          !!authorization.currentUserId &&
          !!authorization.currentIncidentRole &&
          sessionIdentity !== null,
        onAuthorityFailure: () => {
          void authorization.loadSessionRole();
        },
      }}
    >
      <TimelineFileContext.Provider
        value={infrastructure.mutationRuntime.timelineFiles}
      >
        <EvidenceAttachmentContext.Provider
          value={infrastructure.mutationRuntime.evidenceAttachments}
        >
          <WorkbookWorkAreaOverlayProvider>
            <WorkbookRecoveryBoundary
              navigation={recoveryNavigation}
              detailHost={recoveryDetailHost}
              invokerRef={recoveryInvokerRef}
            >
              <IndicatorCreateContext.Provider
                value={infrastructure.mutationRuntime.indicatorCreate}
              >
                <ObservationContext.Provider
                  value={infrastructure.mutationRuntime.indicatorObservations}
                >
                  <WorkbookHistoryContext.Provider
                    value={infrastructure.mutationRuntime}
                  >
                    <IndicatorLifecycleContext.Provider
                      value={infrastructure.mutationRuntime.indicatorLifecycle}
                    >
                      <DecisionSupersessionContext.Provider
                        value={
                          infrastructure.mutationRuntime.decisionSupersession
                        }
                      >
                        <TimelineRelatedEvidenceContext.Provider
                          value={{
                            owner:
                              infrastructure.mutationRuntime
                                .timelineRelatedEvidence,
                            sheetRef: snapshot.startupSheetRef,
                          }}
                        >
                          <CoordinationCreateContext.Provider
                            value={{
                              owner:
                                infrastructure.mutationRuntime
                                  .coordinationCreate,
                              sheetRef: snapshot.startupSheetRef,
                            }}
                          >
                            <NoteCreateContext.Provider
                              value={{
                                owner:
                                  infrastructure.mutationRuntime.noteCreate,
                                sheetRef: snapshot.startupSheetRef,
                              }}
                            >
                              <ContextualCreateContext.Provider
                                value={{
                                  owner:
                                    infrastructure.mutationRuntime
                                      .contextualCreate,
                                  sheetRef: snapshot.startupSheetRef,
                                }}
                              >
                                <section
                                  aria-label="Workbook shell"
                                  data-active-view-schema-id={snapshot.surface}
                                  data-testid={workbookShellReadyTestId()}
                                  data-workbook-shell-id={workbookShellId}
                                  data-cartulary-density={
                                    workbookLayout.shell.density
                                  }
                                  style={panelStyle}
                                >
                                  <WorkbookSaveAnnouncements
                                    runtime={infrastructure.mutationRuntime}
                                  />
                                  <NetworkFlowImportSurface
                                    controller={networkFlowImportController}
                                  />
                                  <NetworkFlowTableSurface
                                    controller={networkFlowTableController}
                                  />
                                  <NetworkFlowIndicatorLinkSurface
                                    controller={
                                      networkFlowIndicatorLinkController
                                    }
                                  />

                                  <WorkbookSurfaceRefreshNotice
                                    runtime={infrastructure.mutationRuntime}
                                  />

                                  <WorkbookBatchRecovery
                                    runtime={infrastructure.mutationRuntime}
                                    activateConflict={recoveryFocus.activate}
                                  />
                                  <WorkbookHistoryRecovery />
                                  <TimelineRelatedEvidenceRecovery
                                    owner={
                                      infrastructure.mutationRuntime
                                        .timelineRelatedEvidence
                                    }
                                  />
                                  <NoteAssociationRecovery
                                    owner={
                                      infrastructure.mutationRuntime
                                        .noteAssociations
                                    }
                                  />
                                  <NoteCreateRecovery
                                    owner={
                                      infrastructure.mutationRuntime.noteCreate
                                    }
                                  />
                                  <CoordinationCreateRecovery
                                    owner={
                                      infrastructure.mutationRuntime
                                        .coordinationCreate
                                    }
                                  />
                                  <ContextualCreateRecovery
                                    owner={
                                      infrastructure.mutationRuntime
                                        .contextualCreate
                                    }
                                  />
                                  <AssessmentAppendRecovery
                                    owner={
                                      infrastructure.mutationRuntime
                                        .assessmentAuthoring
                                    }
                                  />
                                  <PartyLinkRecovery
                                    owner={
                                      infrastructure.mutationRuntime.partyLinks
                                    }
                                  />
                                  <WorkbookIndicatorCreateRecovery
                                    owner={
                                      infrastructure.mutationRuntime
                                        .indicatorCreate
                                    }
                                  />
                                  <WorkbookObservationRecovery
                                    owner={
                                      infrastructure.mutationRuntime
                                        .indicatorObservations
                                    }
                                  />
                                  <WorkbookIndicatorLifecycleRecovery
                                    owner={
                                      infrastructure.mutationRuntime
                                        .indicatorLifecycle
                                    }
                                  />
                                  <TimelineMentionRecovery
                                    owner={infrastructure.timelineMentions}
                                  />
                                  <TimelineCaptureRecovery
                                    owner={infrastructure.timelineCapture}
                                  />
                                  <WorkbookDecisionSupersessionRecovery
                                    runtime={infrastructure.mutationRuntime}
                                  />
                                  <WorkbookEntityMergeRecovery
                                    runtime={infrastructure.mutationRuntime}
                                  />

                                  <WorkbookRecoveryPanel
                                    navigation={recoveryNavigation}
                                    registerDetailHost={setRecoveryDetailHost}
                                    invokerRef={recoveryInvokerRef}
                                    fallbackRef={activeSurfaceFocusRef}
                                  />
                                  <WorkbookShellTopBar
                                    recovery={
                                      <WorkbookRecoveryEntry
                                        navigation={recoveryNavigation}
                                        invokerRef={recoveryInvokerRef}
                                      />
                                    }
                                    account={{
                                      applicationMenu: accountApplication,
                                      displayName:
                                        accountPresentation.displayName,
                                      title: accountPresentation.title,
                                    }}
                                    activeSurfaceFocusRef={
                                      activeSurfaceFocusRef
                                    }
                                    activeSystemSurfaceTitle={
                                      activeSystemSurfaceTitle
                                    }
                                    collaboration={collaboration.snapshot}
                                    incidentIdentity={incidentIdentity}
                                    incidentIdentityError={
                                      incidentIdentityError
                                    }
                                    layout={workbookLayout.shell}
                                    networkAnalysisActive={
                                      networkAnalysisActive
                                    }
                                    networkAnalysisAvailable={
                                      networkAnalysisAvailable
                                    }
                                    onSelectNetworkAnalysis={() => {
                                      if (
                                        networkAnalysisRef.kind ===
                                        "extension_workspace"
                                      ) {
                                        commands.selectExtensionWorkspace(
                                          networkAnalysisRef,
                                        );
                                      }
                                    }}
                                    onSelectSurface={selectBaseWorkbookSurface}
                                    surface={snapshot.surface}
                                  />
                                  <div style={shellContentRegionStyle}>
                                    <WorkbookActiveSurfaceFrame
                                      activeContent={
                                        <WorkbookReferenceContext.Provider
                                          value={{
                                            reader: referenceReader,
                                            actorPresentation:
                                              authorization.currentUserId &&
                                              currentUserLabel
                                                ? {
                                                    userId:
                                                      authorization.currentUserId,
                                                    displayName:
                                                      currentUserLabel,
                                                  }
                                                : undefined,
                                            evidence:
                                              infrastructure.mutationRuntime
                                                .explicitPatches,
                                            onAuthorityFailure: () => {
                                              infrastructure.mutationRuntime.explicitPatches.suspend();
                                              void authorization.loadSessionRole();
                                            },
                                          }}
                                        >
                                          {activeContent}
                                        </WorkbookReferenceContext.Provider>
                                      }
                                      activeSurfaceRef={activeSurfaceFocusRef}
                                      apiBase={apiBase}
                                      focus={recoveryFocus}
                                      mutationRuntime={
                                        infrastructure.mutationRuntime
                                      }
                                      sheetRef={snapshot.startupSheetRef}
                                      onActivateOrigin={
                                        selectBaseWorkbookSurface
                                      }
                                    />
                                    {preferenceController ? (
                                      <WorkbookPreferenceAnnouncements
                                        controller={preferenceController}
                                      />
                                    ) : null}
                                    <WorkbookIncidentControlsPresentation
                                      onIncidentResourceAccepted={
                                        acceptIncidentResource
                                      }
                                      density={workbookLayout.shell.density}
                                      onAuthorizationRecovered={
                                        authorization.acceptRecoveredAuthorization
                                      }
                                      activeMenuItem={
                                        incidentControls.activeMenuItem
                                      }
                                      apiBase={apiBase}
                                      importController={importController}
                                      closeButtonRef={
                                        incidentControls.closeButtonRef
                                      }
                                      currentIncidentRole={
                                        authorization.currentIncidentRole
                                      }
                                      importAssistantAvailable={
                                        importAssistantAvailable
                                      }
                                      incidentId={incidentId}
                                      onClose={incidentControls.closeDrawer}
                                      onAuthorityUncertain={
                                        authorization.loadSessionRole
                                      }
                                      onNavigateToView={(viewSchemaId) => {
                                        commands.selectWorkbookSurface(
                                          viewSchemaId,
                                          {
                                            focusFirstGridTarget: true,
                                          },
                                        );
                                        incidentControls.closeDrawer({
                                          restoreTriggerFocus: false,
                                        });
                                      }}
                                      onSessionRoleChange={
                                        authorization.loadSessionRole
                                      }
                                      renderIncidentControls={
                                        renderIncidentControls
                                      }
                                      section={incidentControls.drawerSection}
                                    />
                                  </div>
                                </section>
                              </ContextualCreateContext.Provider>
                            </NoteCreateContext.Provider>
                          </CoordinationCreateContext.Provider>
                        </TimelineRelatedEvidenceContext.Provider>
                      </DecisionSupersessionContext.Provider>
                    </IndicatorLifecycleContext.Provider>
                  </WorkbookHistoryContext.Provider>
                </ObservationContext.Provider>
              </IndicatorCreateContext.Provider>
            </WorkbookRecoveryBoundary>
          </WorkbookWorkAreaOverlayProvider>
        </EvidenceAttachmentContext.Provider>
      </TimelineFileContext.Provider>
    </WorkbookCandidateAuthorityContext.Provider>
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
      <WorkbookQueryBrowsingProvider
        key={`${props.incidentId}:${props.sessionIdentity ?? "suspended"}`}
      >
        <CommittedWorkbookShell {...props} />
      </WorkbookQueryBrowsingProvider>
    </IncidentCollaborationSession>
  );
}

function CommittedWorkbookShell(props: WorkbookShellProps) {
  const { clientInstanceId } = useIncidentCollaborationSession();
  const create = useCallback(
    () =>
      createWorkbookMutationInfrastructure({
        apiBase: props.apiBase,
        incidentId: props.incidentId,
        clientInstanceId,
        partyAuthorization: props.authorizationRecovery,
      }),
    [
      props.apiBase,
      props.incidentId,
      props.authorizationRecovery,
      clientInstanceId,
    ],
  );
  return (
    <WorkbookMutationRuntimeBoundary
      registry={props.mutationRuntimeRegistry}
      incidentId={props.incidentId}
      clientInstanceId={clientInstanceId}
      create={create}
    >
      {(mutationRuntime) => (
        <WorkbookShellContent
          key={props.incidentId}
          {...props}
          mutationRuntime={mutationRuntime}
        />
      )}
    </WorkbookMutationRuntimeBoundary>
  );
}
