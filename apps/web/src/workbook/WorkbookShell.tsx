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
import { apiPath } from "../services/browserApi";
import type { GeneratedExtensionProfileResource } from "../services/extensionContractAdapter";
import { fetchWorkbookJSON, readEnvelope } from "../services/workbookApi";
import { workbookSheetRefKey } from "../shared/workbookSheetRef";
import type {
  WorkbookAccountApplicationMenuProps,
  WorkbookAccountModel,
  WorkbookIncidentControlsRendererProps,
  WorkbookIncidentRole,
  WorkbookIncidentSnapshot,
} from "../shared/workbookShellContracts";
import { ActiveSurfaceSavedViewSelector } from "./components/ActiveSurfaceSavedViewSelector";
import { IncidentControlsDrawer } from "./components/IncidentControlsDrawer";
import { SystemViewSwitcher } from "./components/SystemViewSwitcher";
import { WorkbookActiveSurface } from "./components/WorkbookActiveSurface";
import { WorkbookGridControls } from "./components/WorkbookGridControls";
import {
  WorkbookShellSlotRegion,
  workbookShellId,
} from "./components/WorkbookShellSlots";
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
  topBarQuerySlotStyle,
} from "./components/WorkbookShellStyles";
import { useIncidentControlsDrawer } from "./hooks/useIncidentControlsDrawer";
import { useWorkbookIncidentIdentity } from "./hooks/useWorkbookIncidentIdentity";
import { useWorkbookPendingGridFocus } from "./hooks/useWorkbookPendingGridFocus";
import { useWorkbookProjectionRefreshController } from "./hooks/useWorkbookProjectionRefreshController";
import { useWorkbookResponsiveLayout } from "./hooks/useWorkbookResponsiveLayout";
import { useWorkbookShellRuntime } from "./hooks/useWorkbookShellRuntime";
import { useWorkbookSurfaceLoaders } from "./hooks/useWorkbookSurfaceLoaders";
import {
  type AccountDensityMode,
  resolveEffectiveWorkbookDensity,
} from "./models/workbookDensity";
import { workbookGridInteractionMode } from "./models/workbookGridState";
import type { WorkbookIncidentIdentity } from "./models/workbookIncidentIdentity";
import {
  requiredBuiltInWorkbookSurfaceIds,
  timelineViewSchemaId,
} from "./models/workbookSurfaceRegistry";
import { displayInitials } from "./utils/workbookPresence";

export type {
  WorkbookAccountApplicationMenuProps,
  WorkbookIncidentControlsRendererProps,
};

type IncidentRole = WorkbookIncidentRole;

type WorkbookShellProps = {
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

type SessionEnvelope = {
  data: {
    user_id: string;
    memberships: Array<{
      incident_id: string;
      role: IncidentRole;
    }>;
  };
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
  const responsiveLayout = useWorkbookResponsiveLayout();
  const responsiveBand = responsiveLayout.chromeMode;
  const surfaceSelectionVersionRef = useRef(0);
  const workbookRuntime = useWorkbookShellRuntime({
    apiBase,
    incidentId,
    onIncidentAccessLost,
    surfaceSelectionVersionRef,
    extensionAvailability,
    onExtensionAvailabilityChange: handleExtensionAvailabilityChange,
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
  const effectiveDensity = useMemo(
    () => resolveEffectiveWorkbookDensity(surface, accountDensityMode),
    [accountDensityMode, surface],
  );
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
  const { incidentIdentity, incidentIdentityError } =
    useWorkbookIncidentIdentity({
      apiBase,
      incidentId,
      initialIncidentIdentity,
      onIncidentAccessLost,
      onIncidentSnapshot,
    });
  const [surfacesMenuOpen, setSurfacesMenuOpen] = useState(false);
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
  const {
    assessmentLoadState,
    assessmentRows,
    entityIndex,
    entityLoadState,
    genericLoadState,
    genericRows,
    hostRows,
    identityRows,
    loadAssessmentSurface,
    loadEntities,
    loadGenericSurface,
  } = useWorkbookSurfaceLoaders({
    activeContract,
    apiBase,
    assessmentQueryState,
    genericQueryState,
    hostQueryState,
    identityQueryState,
    incidentId,
    onIncidentAccessLost,
    surface,
  });
  const interactionMode = workbookGridInteractionMode(
    incidentIdentity?.status,
    currentIncidentRole,
  );
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

  useEffect(() => {
    collaborationSession.publishPresence({
      sheet_ref: startupSheetRef,
      mode: "viewing",
    });
  }, [collaborationSession, startupSheetRef]);

  const loadSessionRole = useCallback(async () => {
    const result = await fetchWorkbookJSON<SessionEnvelope>(
      apiPath(apiBase, "/api/v1/auth/session"),
    );
    if (!result.ok) {
      setCurrentUserId(null);
      setCurrentIncidentRole("");
      onIncidentAccessLost?.();
      return;
    }
    const envelope = readEnvelope<SessionEnvelope>(result.payload);
    setCurrentUserId(envelope.data.user_id || null);
    const membership =
      envelope.data.memberships.find(
        (entry) => entry.incident_id === incidentId,
      ) ?? null;
    if (membership === null) {
      onIncidentAccessLost?.();
    }
    setCurrentIncidentRole(membership?.role ?? "");
  }, [apiBase, incidentId, onIncidentAccessLost]);

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
  const inspectorResetKey = `${surface}:${workbookSheetRefKey(startupSheetRef)}:${sheetReloadToken}`;
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
          data-workbook-block-mode={responsiveLayout.blockMode}
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
              style={surfaceMenuTriggerStyle}
              type="button"
              onClick={() => {
                setSurfacesMenuOpen((current) => !current);
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
        {responsiveBand === "below_supported_minimum" ||
        networkAnalysisActive ? null : (
          <div style={topBarQuerySlotStyle}>
            <WorkbookGridControls
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
          </div>
        )}
        <div style={shellTopBarActionsStyle}>
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
          {activeExtensionWorkspace ?? (
            <WorkbookActiveSurface
              activeContract={activeContract}
              apiBase={apiBase}
              assessmentLoadState={assessmentLoadState}
              assessmentQueryState={assessmentQueryState}
              assessmentRows={assessmentRows}
              authorizationEpoch={`${currentUserId ?? "anonymous"}:${currentIncidentRole ?? "none"}`}
              currentIncidentRole={currentIncidentRole}
              currentUserId={currentUserId}
              density={effectiveDensity}
              entityIndex={entityIndex}
              entityLoadState={entityLoadState}
              genericLoadState={genericLoadState}
              genericQueryState={genericQueryState}
              genericRows={genericRows}
              hostQueryState={hostQueryState}
              hostRows={hostRows}
              identityQueryState={identityQueryState}
              identityRows={identityRows}
              incidentId={incidentId}
              interactionMode={interactionMode}
              inspectorResetKey={inspectorResetKey}
              layoutState={activeLayoutState}
              loadAssessmentSurface={loadAssessmentSurface}
              loadEntities={loadEntities}
              loadGenericSurface={loadGenericSurface}
              onColumnHiddenChange={activeLayoutControls.onColumnHiddenChange}
              onColumnMove={activeLayoutControls.onColumnMove}
              onColumnReorder={activeLayoutControls.onColumnReorder}
              onColumnWidthChange={activeLayoutControls.onColumnWidthChange}
              onIncidentAccessLost={onIncidentAccessLost}
              onResetColumns={activeLayoutControls.onResetColumns}
              savedViewSelector={activeSavedViewSelector}
              setAssessmentQueryState={setAssessmentQueryState}
              setGenericQueryState={setGenericQueryState}
              setHostQueryState={setHostQueryState}
              setIdentityQueryState={setIdentityQueryState}
              setTimelineQueryState={setTimelineQueryState}
              sheetRef={startupSheetRef}
              sheetReloadToken={sheetReloadToken}
              surface={surface}
              timelineQueryState={timelineQueryState}
            />
          )}
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
