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
  type ReactNode,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { NetworkAnalysisWorkspace } from "../networkFlow/NetworkAnalysisWorkspace";
import {
  networkAnalysisURLSelected,
  writeNetworkAnalysisURL,
} from "../networkFlow/networkFlowClient";
import { apiPath } from "../services/browserApi";
import { fetchJSON, readEnvelope } from "../services/workbookApi";
import type {
  WorkbookAccountApplicationMenuProps,
  WorkbookAccountModel,
  WorkbookIncidentControlsMenuItem,
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
import { useWorkbookResponsiveLayout } from "./hooks/useWorkbookResponsiveLayout";
import { useWorkbookShellRuntime } from "./hooks/useWorkbookShellRuntime";
import { useWorkbookSurfaceLoaders } from "./hooks/useWorkbookSurfaceLoaders";
import {
  type AccountDensityMode,
  resolveEffectiveWorkbookDensity,
} from "./models/workbookDensity";
import type { WorkbookIncidentIdentity } from "./models/workbookIncidentIdentity";
import { requiredBuiltInWorkbookSurfaceIds } from "./models/workbookSurfaceRegistry";
import { displayInitials } from "./utils/workbookPresence";

export type { WorkbookIncidentIdentity } from "./models/workbookIncidentIdentity";
export type {
  WorkbookAccountApplicationMenuProps,
  WorkbookAccountModel,
  WorkbookIncidentControlsMenuItem,
  WorkbookIncidentControlsRendererProps,
  WorkbookIncidentRole,
  WorkbookIncidentSnapshot,
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
  networkFlowActivityClaimed?: boolean | undefined;
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

export function WorkbookShell({
  incidentId,
  apiBase,
  account,
  accountDensityMode,
  accountApplicationMenu,
  currentUserLabel,
  initialIncidentIdentity,
  networkFlowActivityClaimed = false,
  onIncidentSnapshot,
  onIncidentAccessLost,
  renderIncidentControls,
}: WorkbookShellProps) {
  const responsiveLayout = useWorkbookResponsiveLayout();
  const responsiveBand = responsiveLayout.chromeMode;
  const surfaceSelectionVersionRef = useRef(0);
  const workbookRuntime = useWorkbookShellRuntime({
    apiBase,
    incidentId,
    onIncidentAccessLost,
    surfaceSelectionVersionRef,
  });
  const {
    activeContract,
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
  const [networkAnalysisActive, setNetworkAnalysisActive] = useState(() =>
    networkFlowActivityClaimed
      ? networkAnalysisURLSelected(new URLSearchParams(window.location.search))
      : false,
  );
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
  const {
    accountIncidentControls,
    activeMenuItem: activeIncidentControlsMenuItem,
    closeButtonRef: incidentControlsCloseButtonRef,
    closeDrawer: closeIncidentControlsDrawer,
    drawerSection: incidentControlsDrawerSection,
  } = useIncidentControlsDrawer();
  const {
    assessmentLoadError,
    assessmentRows,
    entityIndex,
    entityLoadError,
    genericLoadError,
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
  useEffect(() => {
    if (account?.user_id) {
      setCurrentUserId(account.user_id);
    }
  }, [account?.user_id]);

  useEffect(() => {
    if (!networkFlowActivityClaimed && networkAnalysisActive) {
      setNetworkAnalysisActive(false);
    }
  }, [networkAnalysisActive, networkFlowActivityClaimed]);

  useEffect(() => {
    if (networkAnalysisActive) {
      writeNetworkAnalysisURL(incidentId);
    }
  }, [incidentId, networkAnalysisActive]);

  const loadSessionRole = useCallback(async () => {
    const result = await fetchJSON<SessionEnvelope>(
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

  useEffect(() => {
    void Promise.all([loadEntities(), loadSessionRole()]);
  }, [loadEntities, loadSessionRole]);

  useEffect(() => {
    if (sheetReloadToken === 0) {
      return;
    }
    void loadEntities();
  }, [loadEntities, sheetReloadToken]);

  useEffect(() => {
    void sheetReloadToken;
    void loadGenericSurface();
  }, [loadGenericSurface, sheetReloadToken]);

  useEffect(() => {
    void sheetReloadToken;
    void loadAssessmentSurface();
  }, [loadAssessmentSurface, sheetReloadToken]);

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
    incidentControlsDrawerSection === null
      ? null
      : (renderIncidentControls?.({
          activeSection: incidentControlsDrawerSection,
          apiBase,
          currentIncidentRole,
          incidentId,
          onIncidentAccessLost,
          onIncidentSnapshot,
          onSessionRoleChange: loadSessionRole,
        }) ?? null);
  const inspectorResetKey = `${surface}:${startupSheetRef.kind}:${startupSheetRef.id}:${sheetReloadToken}`;
  const selectBaseWorkbookSurface = useCallback(
    (
      viewSchemaId: string,
      options: { readonly focusFirstGridTarget?: boolean } = {},
    ) => {
      setNetworkAnalysisActive(false);
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
          {networkFlowActivityClaimed ? (
            <button
              aria-current={networkAnalysisActive ? "page" : undefined}
              data-testid={networkAnalysisTestId("tab")}
              style={{
                ...surfaceTabStyle,
                ...(networkAnalysisActive ? surfaceTabActiveStyle : null),
              }}
              type="button"
              onClick={() => {
                setNetworkAnalysisActive(true);
              }}
            >
              Network Analysis
            </button>
          ) : null}
          <SystemViewSwitcher
            activeViewSchemaId={surface}
            onSelect={(viewSchemaId) => {
              setNetworkAnalysisActive(false);
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
              onApplyFilter={activeQueryControls.onApplyFilter}
              onClearAll={activeQueryControls.onClearAll}
              onFilterDraftChange={activeQueryControls.onFilterDraftChange}
              onGroupByChange={activeQueryControls.onGroupByChange}
              onRemoveFilter={activeQueryControls.onRemoveFilter}
              onToggleSort={activeQueryControls.onToggleSort}
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
        {entityLoadError ? (
          <p data-testid="entity-load-error" style={shellContentNoticeStyle}>
            {entityLoadError}
          </p>
        ) : null}

        <div style={shellActiveSurfaceStyle}>
          {networkAnalysisActive ? (
            <NetworkAnalysisWorkspace
              apiBase={apiBase}
              currentIncidentRole={currentIncidentRole}
              incidentId={incidentId}
              onIncidentAccessLost={onIncidentAccessLost}
            />
          ) : (
            <WorkbookActiveSurface
              activeContract={activeContract}
              apiBase={apiBase}
              assessmentLoadError={assessmentLoadError}
              assessmentQueryState={assessmentQueryState}
              assessmentRows={assessmentRows}
              currentIncidentRole={currentIncidentRole}
              currentUserId={currentUserId}
              density={effectiveDensity}
              entityIndex={entityIndex}
              genericLoadError={genericLoadError}
              genericQueryState={genericQueryState}
              genericRows={genericRows}
              hostQueryState={hostQueryState}
              hostRows={hostRows}
              identityQueryState={identityQueryState}
              identityRows={identityRows}
              incidentId={incidentId}
              inspectorResetKey={inspectorResetKey}
              loadAssessmentSurface={loadAssessmentSurface}
              loadEntities={loadEntities}
              loadGenericSurface={loadGenericSurface}
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
