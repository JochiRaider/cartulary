import { workbookShellReadyTestId } from "@cartulary/ui-contracts";
import { type ReactNode, useCallback, useEffect, useMemo, useRef } from "react";
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
} from "../extensions/extensionWorkspaceIdentities";
import type { AuthorizationRecoveryPort } from "../shared/authorizationRecovery";
import type {
  WorkbookAccountApplicationMenuProps,
  WorkbookAccountModel,
  WorkbookIncidentControlsRendererProps,
} from "../shared/workbookShellContracts";
import { WorkbookActiveSurfaceFrame } from "./components/WorkbookActiveSurfaceFrame";
import { WorkbookActiveSurfacePresentation } from "./components/WorkbookActiveSurfacePresentation";
import { WorkbookIncidentControlsPresentation } from "./components/WorkbookIncidentControlsPresentation";
import { WorkbookSaveAnnouncements } from "./components/WorkbookSaveAnnouncements";
import { workbookShellId } from "./components/WorkbookShellSlots";
import { WorkbookShellTopBar } from "./components/WorkbookShellTopBar";
import { workbookShellViewBarWorkingSet } from "./components/WorkbookShellViewBarControls";
import { WorkbookStatusStrip } from "./components/WorkbookStatusStrip";
import { useIncidentControlsDrawer } from "./hooks/useIncidentControlsDrawer";
import { useWorkbookAuthorizationState } from "./hooks/useWorkbookAuthorizationState";
import { useWorkbookCollaborationLifecycle } from "./hooks/useWorkbookCollaborationLifecycle";
import {
  useWorkbookExtensionAvailability,
  useWorkbookExtensionFallback,
} from "./hooks/useWorkbookExtensionAvailability";
import { useWorkbookIncidentIdentity } from "./hooks/useWorkbookIncidentIdentity";
import { useWorkbookProjectionRefreshController } from "./hooks/useWorkbookProjectionRefreshController";
import { useWorkbookRecoveryFocus } from "./hooks/useWorkbookRecoveryFocus";
import {
  useWorkbookReferenceQueryBroker,
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
import { WorkbookMutationRuntimeRegistry } from "./runtime/WorkbookMutationRuntimeRegistry";
import { projectWorkbookStatusForSurface } from "./runtime/workbookMutationStatusProjector";
import type { WorkbookSurfacesFacadeProps } from "./surfaces/WorkbookSurfacesFacade";

export type {
  WorkbookAccountApplicationMenuProps,
  WorkbookIncidentControlsRendererProps,
};

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
  extensionProfiles?: readonly ExtensionDiscoveryProfile[] | undefined;
  onIncidentAccessLost?: (() => void) | undefined;
  renderIncidentControls?:
    | ((props: WorkbookIncidentControlsRendererProps) => ReactNode)
    | undefined;
  mutationRuntimeRegistry?: WorkbookMutationRuntimeRegistry | undefined;
};

type WorkbookShellContentProps = WorkbookShellProps & {
  mutationRuntimeRegistry: WorkbookMutationRuntimeRegistry;
};

const noExtensionProfiles: readonly ExtensionDiscoveryProfile[] = [];

function WorkbookShellContent({
  authorizationRecovery,
  incidentId,
  apiBase,
  account,
  accountDensityMode,
  accountApplicationMenu,
  currentUserLabel,
  initialIncidentIdentity,
  extensionProfiles = noExtensionProfiles,
  onIncidentAccessLost,
  renderIncidentControls,
  mutationRuntimeRegistry,
}: WorkbookShellContentProps) {
  const collaborationSession = useIncidentCollaborationSession();
  const extensionLifecycle = useWorkbookExtensionAvailability({
    clientInstanceId: collaborationSession.clientInstanceId,
    incidentId,
    profiles: extensionProfiles,
  });
  const infrastructure = useWorkbookShellInfrastructure({
    apiBase,
    clientInstanceId: collaborationSession.clientInstanceId,
    extensionAvailability: extensionLifecycle.controller,
    incidentId,
    mutationRuntimeRegistry,
    onExtensionAvailabilityChange: extensionLifecycle.publishChange,
    onIncidentAccessLost,
  });
  const { commands, snapshot } = infrastructure.workbookRuntime;
  const authorization = useWorkbookAuthorizationState({
    accountUserId: account?.user_id,
    authorizationRecovery,
    incidentId,
    onIncidentAccessLost,
  });
  const referenceQueryBroker = useWorkbookReferenceQueryBroker(
    authorization.authorizationGeneration,
    infrastructure.viewQuery,
  );
  const { incidentIdentity, incidentIdentityError } =
    useWorkbookIncidentIdentity({
      incidentPort: infrastructure.incidentPort,
      incidentId,
      initialIncidentIdentity,
      onIncidentAccessLost,
    });
  const queries = useWorkbookSurfaceQueries({
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
    onIncidentAccessLost,
    referenceBroker: referenceQueryBroker,
    sheetRef: snapshot.startupSheetRef,
    surface: snapshot.surface,
    timeline: {
      setState: commands.setTimelineQueryState,
      state: snapshot.timelineQueryState,
    },
    viewQuery: infrastructure.viewQuery,
  });
  const collaboration = useWorkbookCollaborationLifecycle({
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
      onColumnWidthChange: snapshot.activeLayoutControls.onColumnWidthChange,
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
  const networkAnalysisAvailable = extensionLifecycle.controller.isRenderable({
    extensionProfileId: networkFlowActivityProfileId,
    workspaceKey: networkAnalysisWorkspaceKey,
  });
  const selectTimelineFallback = useCallback(() => {
    commands.selectWorkbookSurface(timelineViewSchemaId);
  }, [commands.selectWorkbookSurface]);
  useWorkbookExtensionFallback({
    active: networkAnalysisActive,
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
  const activeStatus = projectWorkbookStatusForSurface(
    infrastructure.mutationSnapshot,
    snapshot.startupSheetRef,
  );
  const recoveryFocus = useWorkbookRecoveryFocus({
    activeSurfaceRef: activeSurfaceFocusRef,
    runtime: infrastructure.mutationRuntime,
    snapshot: activeStatus,
    onSessionRecovery: authorization.loadSessionRole,
  });
  const importAssistantAvailable =
    extensionLifecycle.controller.isRouteAvailable(
      importProfileId,
      importRouteFamily,
    );
  const incidentControls = useIncidentControlsDrawer(importAssistantAvailable);
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
  });
  const facadeProps: WorkbookSurfacesFacadeProps = {
    collaboration: { projection: collaboration.projection },
    continuity: { resetKey: collaboration.continuityResetKey },
    gridEntryFocus: {
      acknowledge: commands.acknowledgeGridEntryFocus,
      request: snapshot.gridEntryFocusRequest,
    },
    incident: {
      apiBase,
      currentIncidentRole: authorization.currentIncidentRole,
      currentUserId: authorization.currentUserId,
      incidentPort: infrastructure.incidentPort,
      incidentId,
      onIncidentAccessLost,
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
        workbookStatus: (
          <WorkbookStatusStrip
            status={activeStatus}
            chromeMode={workbookLayout.shell.chromeMode}
            showPresence={false}
            onActivateConflict={recoveryFocus.activate}
            workbookFocusAnchor={null}
          />
        ),
        apiBase,
        currentIncidentRole: authorization.currentIncidentRole,
        incidentId,
        onIncidentAccessLost,
      }}
      sheetRef={snapshot.startupSheetRef}
      surface={facadeProps}
    />
  );

  return (
    <section
      aria-label="Workbook shell"
      data-active-view-schema-id={snapshot.surface}
      data-testid={workbookShellReadyTestId()}
      data-workbook-shell-id={workbookShellId}
      style={panelStyle}
    >
      <WorkbookSaveAnnouncements runtime={infrastructure.mutationRuntime} />
      <WorkbookShellTopBar
        account={{
          applicationMenu: accountApplication,
          displayName: accountPresentation.displayName,
          title: accountPresentation.title,
        }}
        activeSurfaceFocusRef={activeSurfaceFocusRef}
        activeSystemSurfaceTitle={activeSystemSurfaceTitle}
        collaboration={collaboration.snapshot}
        incidentIdentity={incidentIdentity}
        incidentIdentityError={incidentIdentityError}
        layout={workbookLayout.shell}
        networkAnalysisActive={networkAnalysisActive}
        networkAnalysisAvailable={networkAnalysisAvailable}
        onSelectNetworkAnalysis={() => {
          if (networkAnalysisRef.kind === "extension_workspace") {
            commands.selectExtensionWorkspace(networkAnalysisRef);
          }
        }}
        onSelectSurface={selectBaseWorkbookSurface}
        surface={snapshot.surface}
      />
      <div style={shellContentRegionStyle}>
        <WorkbookActiveSurfaceFrame
          activeContent={activeContent}
          activeSurfaceRef={activeSurfaceFocusRef}
          apiBase={apiBase}
          focus={recoveryFocus}
          mutationRuntime={infrastructure.mutationRuntime}
          mutationSnapshot={activeStatus}
          onActivateOrigin={selectBaseWorkbookSurface}
        />
        <WorkbookIncidentControlsPresentation
          activeMenuItem={incidentControls.activeMenuItem}
          apiBase={apiBase}
          availability={extensionLifecycle.controller}
          closeButtonRef={incidentControls.closeButtonRef}
          currentIncidentRole={authorization.currentIncidentRole}
          importAssistantAvailable={importAssistantAvailable}
          incidentId={incidentId}
          onClose={incidentControls.closeDrawer}
          onIncidentAccessLost={onIncidentAccessLost}
          onNavigateToView={(viewSchemaId) => {
            commands.selectWorkbookSurface(viewSchemaId, {
              focusFirstGridTarget: true,
            });
            incidentControls.closeDrawer({ restoreTriggerFocus: false });
          }}
          onSessionRoleChange={authorization.loadSessionRole}
          renderIncidentControls={renderIncidentControls}
          section={incidentControls.drawerSection}
        />
      </div>
    </section>
  );
}

export function WorkbookShell(props: WorkbookShellProps) {
  const localMutationRuntimeRegistry = useMemo(
    () => new WorkbookMutationRuntimeRegistry(),
    [],
  );
  const mutationRuntimeRegistry =
    props.mutationRuntimeRegistry ?? localMutationRuntimeRegistry;
  useEffect(
    () => () => {
      if (props.mutationRuntimeRegistry === undefined) {
        localMutationRuntimeRegistry.dispose();
      }
    },
    [localMutationRuntimeRegistry, props.mutationRuntimeRegistry],
  );
  return (
    <IncidentCollaborationSession
      apiBase={props.apiBase}
      incidentId={props.incidentId}
      initialPresence={{
        sheet_ref: { kind: "view_schema", id: timelineViewSchemaId },
        mode: "viewing",
      }}
    >
      <WorkbookShellContent
        {...props}
        mutationRuntimeRegistry={mutationRuntimeRegistry}
      />
    </IncidentCollaborationSession>
  );
}
