import {
  currentIncidentRoleTestId,
  type IncidentControlsSection,
  incidentControlsMenuItemTestId,
  incidentControlsMenuTestId,
  incidentControlsTriggerTestId,
  type LandingAdminPanelToken,
  landingAdminMenuItemTestId,
  landingAdminPanelTestId,
  landingAdminShellTestId,
  landingIncidentCardTestId,
  landingIncidentOpenButtonTestId,
  phase1AccountTestId,
  phase1ErrorCodeTestId,
  phase1ErrorSummaryTestIds,
  phase1LandingTestId,
} from "@cartulary/ui-contracts";
import {
  ChevronDown,
  ChevronRight,
  FileClock,
  FolderOpen,
  LockKeyhole,
  Package,
  Palette,
  RefreshCw,
  Search,
  Upload,
  UserRound,
  UsersRound,
  X,
} from "lucide-react";
import {
  type CSSProperties,
  Fragment,
  type KeyboardEvent,
  type MutableRefObject,
  type ReactNode,
  useCallback,
  useEffect,
  useId,
  useRef,
  useState,
} from "react";

import {
  type APIError,
  clientTxnID,
  csrfCookieName,
  csrfHeaderName,
  extractError,
  fetchJSON,
  publicErrorView,
  readCookie,
} from "../services/browserApi";
import {
  type AccountPreferencesResource,
  type AccountProfileResource,
  type DensityMode,
  loadAccountPreferences,
  loadAccountProfile,
  patchAccountProfile,
  putAccountPreferences,
} from "./phase1Client";

export type IncidentData = {
  incident_id: string;
  incident_key: string;
  title: string;
  description: string | null;
  severity: string | null;
  tlp: string | null;
  current_phase: string | null;
  primary_external_case_ref: string | null;
  incident_version: number;
  status?: "active" | "closed";
  created_at?: string;
  updated_at?: string;
  closed_at?: string | null;
};

export type IncidentStatusFilter = "active" | "all" | "closed";

export type PagingMeta = {
  limit: number;
  has_more: boolean;
  next_cursor: string | null;
};

export type AppBootstrapState =
  | "loading"
  | "anonymous"
  | "authenticated"
  | "forbidden"
  | "revoked"
  | "public_error_envelope";

export type LandingRefreshState = "idle" | "loading" | "failed";

export type LandingAdminPanelDescriptor = {
  description: string;
  group: "account" | "deployment" | "primary";
  label: string;
  token: LandingAdminPanelToken;
};

export type AccountSettingsPanelToken =
  | "account-appearance"
  | "account-profile"
  | "account-security";

export type DeploymentAdministrationPanelToken =
  | "administrative-audit"
  | "deployment-users"
  | "incident-import"
  | "reference-packs";

export type LandingAdminShellProps = {
  accountMenu: ReactNode;
  activePanel: DeploymentAdministrationPanelToken;
  availablePanels: ReadonlyArray<LandingAdminPanelDescriptor>;
  children: ReactNode;
  currentUserLabel: string;
  onActivePanelChange: (panel: DeploymentAdministrationPanelToken) => void;
  statusText: string;
};

export type IncidentDirectoryShellProps = {
  accountMenu: ReactNode;
  children: ReactNode;
  currentUserLabel: string;
  statusText: string;
};

export type AccountApplicationMenuProps = {
  canOpenDeploymentAdministration: boolean;
  currentContext: "deployment-administration" | "incidents" | "workbook";
  currentUserLabel: string;
  currentIncidentRole?: string | null | undefined;
  incidentControls?:
    | {
        readonly activeSection: IncidentControlsSection;
        readonly items: ReadonlyArray<{
          readonly description: string;
          readonly label: string;
          readonly section: IncidentControlsSection;
        }>;
        readonly onSelectSection: (
          section: IncidentControlsSection,
          returnFocusTarget?: HTMLElement | null,
        ) => void;
      }
    | undefined;
  onOpenAccountSettings: (panel: AccountSettingsPanelToken) => void;
  onOpenDeploymentAdministration: () => void;
  onOpenIncidentDirectory: () => void;
  triggerTestId?: string | undefined;
};

export type IncidentLandingProps = {
  bootstrapState: AppBootstrapState;
  createIncidentCurrentPhase: string;
  createIncidentDescription: string;
  createIncidentExternalCase: string;
  createIncidentKey: string;
  createIncidentSeverity: string;
  createIncidentTitle: string;
  createIncidentTLP: string;
  error: APIError | null;
  hasMoreIncidents: boolean;
  incidents: IncidentData[];
  incidentSearch: string;
  incidentStatusFilter: IncidentStatusFilter;
  isRefreshing: boolean;
  onCreate: () => Promise<void> | void;
  onCreateIncidentCurrentPhaseChange: (value: string) => void;
  onCreateIncidentDescriptionChange: (value: string) => void;
  onCreateIncidentExternalCaseChange: (value: string) => void;
  onCreateIncidentKeyChange: (value: string) => void;
  onCreateIncidentSeverityChange: (value: string) => void;
  onCreateIncidentTitleChange: (value: string) => void;
  onCreateIncidentTLPChange: (value: string) => void;
  onLoadMore: () => Promise<void> | void;
  onOpenIncident: (incidentId: string) => void;
  onRefresh: () => Promise<void> | void;
  onSearchChange: (value: string) => void;
  onSearchSubmit: () => Promise<void> | void;
  onStatusFilterChange: (value: IncidentStatusFilter) => void;
  statusText: string;
};

type AdministrativeAuditEvent = {
  audit_event_id: string;
  occurred_at: string;
  actor_kind?: "operator" | "system" | "user";
  actor_user_id: string | null;
  source?: string;
  action_code: string;
  target_kind: string;
  target_id: string | null;
  changes: Array<{
    field_path: string;
    value_state: "redacted" | "visible";
    before: unknown;
    after: unknown;
  }>;
  reason_code: string | null;
};

type JobResource = {
  job_id?: string;
  status?: string;
  cancelable?: boolean;
  progress?: {
    completed?: number;
    total?: number | null;
  };
  result_summary?: { code?: string; resource_refs?: unknown[] } | null;
  error_summary?: { code?: string } | null;
};

type JobResourceRef = {
  kind?: unknown;
  id?: unknown;
};

const panelIcons: Record<LandingAdminPanelToken, typeof FolderOpen> = {
  incidents: FolderOpen,
  "deployment-users": UsersRound,
  "administrative-audit": FileClock,
  "reference-packs": Package,
  "incident-import": Upload,
  "account-profile": UserRound,
  "account-appearance": Palette,
  "account-security": LockKeyhole,
};

const terminalJobStates = new Set(["succeeded", "failed", "canceled"]);

function importedIncidentIDFromJob(job: JobResource | null): string | null {
  const refs = job?.result_summary?.resource_refs;
  if (!Array.isArray(refs)) {
    return null;
  }
  const incidentRefs = refs.filter((ref): ref is JobResourceRef => {
    if (typeof ref !== "object" || ref === null || Array.isArray(ref)) {
      return false;
    }
    const candidate = ref as JobResourceRef;
    return candidate.kind === "incident" && typeof candidate.id === "string";
  });
  if (incidentRefs.length !== 1) {
    return null;
  }
  const incidentID = incidentRefs[0]?.id;
  return typeof incidentID === "string" && incidentID.trim() !== ""
    ? incidentID
    : null;
}

export function LandingAdminShell({
  accountMenu,
  activePanel,
  availablePanels,
  children,
  currentUserLabel,
  onActivePanelChange,
  statusText,
}: LandingAdminShellProps) {
  const menuItemRefs = useRef(
    new Map<LandingAdminPanelToken, HTMLButtonElement>(),
  );
  const deploymentPanels = availablePanels.filter(
    (panel) => panel.group === "deployment",
  );
  const navigationPanels = deploymentPanels;

  function focusPanelMenuItem(panel: DeploymentAdministrationPanelToken) {
    const focus = () => {
      menuItemRefs.current.get(panel)?.focus();
    };
    if (typeof window.requestAnimationFrame === "function") {
      window.requestAnimationFrame(focus);
      return;
    }
    window.setTimeout(focus, 0);
  }

  function selectPanel(
    panel: DeploymentAdministrationPanelToken,
    focus = false,
  ) {
    onActivePanelChange(panel);
    if (focus) {
      focusPanelMenuItem(panel);
    }
  }

  function handleMenuKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    const currentIndex = navigationPanels.findIndex(
      (panel) => panel.token === activePanel,
    );
    const lastIndex = navigationPanels.length - 1;
    const selectByIndex = (index: number) => {
      event.preventDefault();
      selectPanel(
        (navigationPanels[index]?.token ??
          "deployment-users") as DeploymentAdministrationPanelToken,
        true,
      );
    };

    switch (event.key) {
      case "ArrowUp":
      case "ArrowLeft":
        selectByIndex(currentIndex <= 0 ? lastIndex : currentIndex - 1);
        return;
      case "ArrowDown":
      case "ArrowRight":
        selectByIndex(currentIndex >= lastIndex ? 0 : currentIndex + 1);
        return;
      case "Home":
        selectByIndex(0);
        return;
      case "End":
        selectByIndex(lastIndex);
        return;
      default:
        return;
    }
  }

  return (
    <section
      data-testid={landingAdminShellTestId("shell")}
      style={landingAdminShellStyle}
    >
      <header style={landingAdminHeaderStyle}>
        <div style={brandBlockStyle}>
          <p style={landingEyebrowStyle}>Cartulary</p>
          <h1 style={landingAdminTitleStyle}>Deployment administration</h1>
        </div>
        <dl style={landingAdminHeaderMetaStyle}>
          <div>
            <dt style={landingToolbarLabelStyle}>Session</dt>
            <dd
              data-testid={phase1LandingTestId("current-user")}
              style={landingAdminMetaValueStyle}
            >
              {currentUserLabel}
            </dd>
          </div>
        </dl>
        <div style={landingAccountNavStyle}>{accountMenu}</div>
      </header>

      <div style={landingAdminWorkspaceStyle}>
        <nav
          data-testid={landingAdminShellTestId("menu")}
          style={landingAdminMenuStyle}
          aria-label="Deployment administration"
          onKeyDown={handleMenuKeyDown}
        >
          <div style={landingAdminMenuItemsStyle}>
            {deploymentPanels.length > 0 ? (
              <MenuGroup
                title="Administration"
                panels={deploymentPanels}
                activePanel={activePanel}
                menuItemRefs={menuItemRefs}
                onSelect={selectPanel}
              />
            ) : null}
          </div>
        </nav>
        <div style={landingAdminContentStyle}>{children}</div>
      </div>

      <p aria-live="polite" role="status" style={visuallyHiddenStyle}>
        {statusText}
      </p>
    </section>
  );
}

export function IncidentDirectoryShell({
  accountMenu,
  children,
  currentUserLabel,
  statusText,
}: IncidentDirectoryShellProps) {
  return (
    <section
      data-testid={landingAdminShellTestId("shell")}
      style={incidentDirectoryShellStyle}
    >
      <header style={landingAdminHeaderStyle}>
        <div style={brandBlockStyle}>
          <p style={landingEyebrowStyle}>Cartulary</p>
          <h1 style={landingAdminTitleStyle}>Incident directory</h1>
        </div>
        <dl style={landingAdminHeaderMetaStyle}>
          <div>
            <dt style={landingToolbarLabelStyle}>Session</dt>
            <dd
              data-testid={phase1LandingTestId("current-user")}
              style={landingAdminMetaValueStyle}
            >
              {currentUserLabel}
            </dd>
          </div>
        </dl>
        <div style={landingAccountNavStyle}>{accountMenu}</div>
      </header>
      <div style={landingAdminContentStyle}>{children}</div>
      <p aria-live="polite" role="status" style={visuallyHiddenStyle}>
        {statusText}
      </p>
    </section>
  );
}

export function AccountApplicationMenu({
  canOpenDeploymentAdministration,
  currentContext,
  currentUserLabel,
  currentIncidentRole,
  incidentControls,
  onOpenAccountSettings,
  onOpenDeploymentAdministration,
  onOpenIncidentDirectory,
  triggerTestId,
}: AccountApplicationMenuProps) {
  const [open, setOpen] = useState(false);
  const [controlsOpen, setControlsOpen] = useState(false);
  const containerRef = useRef<HTMLFieldSetElement | null>(null);
  const triggerRef = useRef<HTMLButtonElement | null>(null);
  const closeMenu = useCallback(() => {
    setOpen(false);
    setControlsOpen(false);
  }, []);

  return (
    <fieldset
      ref={containerRef}
      aria-label="Account and application menu"
      style={accountMenuAnchorStyle}
      onBlur={(event) => {
        const nextFocus = event.relatedTarget;
        if (
          nextFocus instanceof Node &&
          containerRef.current?.contains(nextFocus)
        ) {
          return;
        }
        closeMenu();
      }}
    >
      <button
        ref={triggerRef}
        aria-controls={open ? "account-application-menu" : undefined}
        aria-expanded={open}
        aria-haspopup="menu"
        aria-label="Account and application navigation"
        data-testid={triggerTestId}
        style={accountMenuTriggerStyle}
        type="button"
        onClick={() => {
          setOpen((current) => {
            const nextOpen = !current;
            if (!nextOpen) {
              setControlsOpen(false);
            }
            return nextOpen;
          });
        }}
      >
        <UserRound aria-hidden="true" size={16} />
        <span style={accountMenuTriggerTextStyle}>{currentUserLabel}</span>
        <ChevronDown aria-hidden="true" size={15} />
      </button>
      {open ? (
        <div id="account-application-menu" role="menu" style={accountMenuStyle}>
          {currentContext === "workbook" ? (
            <div
              aria-disabled="true"
              data-testid={currentIncidentRoleTestId()}
              role="menuitem"
              style={accountMenuStatusItemStyle}
              tabIndex={-1}
            >
              Current incident role: {currentIncidentRole || "viewer"}
            </div>
          ) : null}
          <button
            aria-current={currentContext === "incidents" ? "page" : undefined}
            data-testid={
              currentContext === "workbook"
                ? phase1LandingTestId("return")
                : undefined
            }
            role="menuitem"
            style={accountMenuItemStyle}
            type="button"
            onClick={() => {
              closeMenu();
              onOpenIncidentDirectory();
            }}
          >
            Incidents
          </button>
          {canOpenDeploymentAdministration ? (
            <button
              aria-current={
                currentContext === "deployment-administration"
                  ? "page"
                  : undefined
              }
              role="menuitem"
              style={accountMenuItemStyle}
              type="button"
              onClick={() => {
                closeMenu();
                onOpenDeploymentAdministration();
              }}
            >
              Deployment administration
            </button>
          ) : null}
          {incidentControls ? (
            <>
              <button
                aria-controls={
                  controlsOpen ? incidentControlsMenuTestId() : undefined
                }
                aria-expanded={controlsOpen}
                aria-haspopup="menu"
                data-testid={incidentControlsTriggerTestId()}
                role="menuitem"
                style={accountMenuItemStyle}
                type="button"
                onClick={() => {
                  setControlsOpen((current) => !current);
                }}
              >
                Controls
              </button>
              {controlsOpen ? (
                <div
                  data-testid={incidentControlsMenuTestId()}
                  id={incidentControlsMenuTestId()}
                  role="menu"
                  style={accountSubmenuStyle}
                >
                  {incidentControls.items.map((item) => (
                    <button
                      key={item.section}
                      data-testid={incidentControlsMenuItemTestId(item.section)}
                      role="menuitem"
                      style={
                        item.section === incidentControls.activeSection
                          ? accountSubmenuItemSelectedStyle
                          : accountSubmenuItemStyle
                      }
                      type="button"
                      onClick={() => {
                        const returnFocusTarget = triggerRef.current;
                        closeMenu();
                        incidentControls.onSelectSection(
                          item.section,
                          returnFocusTarget,
                        );
                      }}
                    >
                      <span style={accountSubmenuItemLabelStyle}>
                        {item.label}
                      </span>
                      <span style={accountSubmenuItemDescriptionStyle}>
                        {item.description}
                      </span>
                    </button>
                  ))}
                </div>
              ) : null}
            </>
          ) : null}
          <div aria-hidden="true" style={accountMenuSeparatorStyle} />
          <button
            role="menuitem"
            style={accountMenuItemStyle}
            type="button"
            onClick={() => {
              closeMenu();
              onOpenAccountSettings("account-profile");
            }}
          >
            Account settings
          </button>
        </div>
      ) : null}
    </fieldset>
  );
}

function MenuGroup({
  activePanel,
  menuItemRefs,
  onSelect,
  panels,
  title,
}: {
  activePanel: DeploymentAdministrationPanelToken;
  menuItemRefs: MutableRefObject<
    Map<LandingAdminPanelToken, HTMLButtonElement>
  >;
  onSelect: (
    panel: DeploymentAdministrationPanelToken,
    focus?: boolean,
  ) => void;
  panels: ReadonlyArray<LandingAdminPanelDescriptor>;
  title: string;
}) {
  if (panels.length === 0) {
    return null;
  }
  return (
    <div style={menuGroupStyle}>
      <p style={menuGroupTitleStyle}>{title}</p>
      <div style={menuGroupItemsStyle}>
        {panels.map((panel) => {
          const token = panel.token as DeploymentAdministrationPanelToken;
          const selected = token === activePanel;
          return (
            <PanelButton
              key={panel.token}
              panel={panel}
              selected={selected}
              refCallback={(element) => {
                if (element === null) {
                  menuItemRefs.current.delete(panel.token);
                  return;
                }
                menuItemRefs.current.set(panel.token, element);
              }}
              onClick={() => {
                onSelect(token);
              }}
            />
          );
        })}
      </div>
    </div>
  );
}

function PanelButton({
  compact = false,
  onClick,
  panel,
  refCallback,
  selected,
}: {
  compact?: boolean;
  onClick: () => void;
  panel: LandingAdminPanelDescriptor;
  refCallback?: (element: HTMLButtonElement | null) => void;
  selected: boolean;
}) {
  const Icon = panelIcons[panel.token];
  const style = compact
    ? selected
      ? landingAccountNavButtonSelectedStyle
      : landingAccountNavButtonStyle
    : selected
      ? landingAdminMenuItemSelectedStyle
      : landingAdminMenuItemStyle;
  return (
    <button
      id={landingAdminMenuItemTestId(panel.token)}
      ref={refCallback}
      aria-controls={landingAdminPanelTestId(panel.token)}
      aria-pressed={selected}
      data-testid={landingAdminMenuItemTestId(panel.token)}
      style={style}
      type="button"
      onClick={onClick}
    >
      <Icon size={compact ? 15 : 17} strokeWidth={2.2} />
      <span style={landingAdminMenuItemTextStyle}>
        <span style={landingAdminMenuItemLabelStyle}>{panel.label}</span>
        {compact ? null : (
          <span style={landingAdminMenuItemDescriptionStyle}>
            {panel.description}
          </span>
        )}
      </span>
    </button>
  );
}

export function IncidentLanding({
  bootstrapState,
  createIncidentCurrentPhase,
  createIncidentDescription,
  createIncidentExternalCase,
  createIncidentKey,
  createIncidentSeverity,
  createIncidentTitle,
  createIncidentTLP,
  error,
  hasMoreIncidents,
  incidents,
  incidentSearch,
  incidentStatusFilter,
  isRefreshing,
  onCreate,
  onCreateIncidentCurrentPhaseChange,
  onCreateIncidentDescriptionChange,
  onCreateIncidentExternalCaseChange,
  onCreateIncidentKeyChange,
  onCreateIncidentSeverityChange,
  onCreateIncidentTitleChange,
  onCreateIncidentTLPChange,
  onLoadMore,
  onOpenIncident,
  onRefresh,
  onSearchChange,
  onSearchSubmit,
  onStatusFilterChange,
  statusText,
}: IncidentLandingProps) {
  const incidentKeyFieldId = useId();
  const incidentTitleFieldId = useId();
  const [createOpen, setCreateOpen] = useState(false);
  const hasIncidents = incidents.length > 0;
  const hasActiveQuery =
    incidentSearch.trim() !== "" || incidentStatusFilter !== "all";

  async function submitCreate() {
    await onCreate();
  }

  return (
    <section
      aria-busy={isRefreshing}
      data-bootstrap-state={bootstrapState}
      data-testid={phase1LandingTestId("shell")}
      style={surfacePanelStyle}
    >
      <header style={surfaceHeaderStyle}>
        <div>
          <p style={sectionEyebrowStyle}>Incidents</p>
          <h2 style={sectionTitleStyle}>Visible incidents</h2>
        </div>
        <div style={headerActionRowStyle}>
          <p
            data-testid={phase1LandingTestId("incidents-count")}
            style={countPillStyle}
          >
            {incidents.length} loaded{hasMoreIncidents ? " +" : ""}
          </p>
          <button
            data-testid={phase1LandingTestId("refresh")}
            style={secondaryButtonStyle}
            type="button"
            onClick={() => {
              void onRefresh();
            }}
          >
            <RefreshCw size={15} />
            Refresh
          </button>
          {!createOpen ? (
            <button
              data-testid={phase1LandingTestId("create-open-button")}
              style={primaryButtonStyle}
              type="button"
              onClick={() => {
                setCreateOpen(true);
              }}
            >
              New incident
            </button>
          ) : null}
        </div>
      </header>

      <div style={incidentWorkspaceStyle}>
        <section style={incidentDirectoryStyle}>
          <div style={toolbarGridStyle}>
            <label htmlFor="incident-filter" style={labelBlockStyle}>
              Search visible incidents
              <span style={searchInputShellStyle}>
                <Search size={16} />
                <input
                  id="incident-filter"
                  data-testid={phase1LandingTestId("search")}
                  style={searchInputStyle}
                  value={incidentSearch}
                  onChange={(event) => {
                    onSearchChange(event.target.value);
                  }}
                  onKeyDown={(event) => {
                    if (event.key === "Enter") {
                      void onSearchSubmit();
                    }
                  }}
                  placeholder="Key, title, severity, TLP, phase, external case"
                />
              </span>
            </label>
            <label htmlFor="incident-status-filter" style={labelBlockStyle}>
              Status
              <select
                id="incident-status-filter"
                data-testid={phase1LandingTestId("status-filter")}
                style={inputStyle}
                value={incidentStatusFilter}
                onChange={(event) => {
                  onStatusFilterChange(
                    event.target.value as IncidentStatusFilter,
                  );
                }}
              >
                <option value="all">All</option>
                <option value="active">Active</option>
                <option value="closed">Closed</option>
              </select>
            </label>
          </div>

          {isRefreshing ? (
            <p
              aria-live="polite"
              data-testid={phase1LandingTestId("loading")}
              role="status"
              style={inlineStatusStyle}
            >
              Searching visible incidents…
            </p>
          ) : null}

          {!isRefreshing && !hasIncidents ? (
            <p
              data-testid={phase1LandingTestId("empty-state")}
              style={emptyStateStyle}
            >
              {hasActiveQuery
                ? "No visible incidents match this query."
                : "No incidents are visible for this session yet."}
            </p>
          ) : null}

          {hasIncidents ? (
            <div
              data-testid={phase1LandingTestId("incident-list")}
              style={tableShellStyle}
            >
              <table style={dataTableStyle}>
                <thead>
                  <tr>
                    <th style={tableHeaderCellStyle}>Incident</th>
                    <th style={tableHeaderCellStyle}>Status</th>
                    <th style={tableHeaderCellStyle}>Phase</th>
                    <th style={tableHeaderCellStyle}>Severity</th>
                    <th style={tableHeaderCellStyle}>TLP</th>
                    <th style={tableHeaderCellStyle}>External case</th>
                    <th style={tableHeaderCellStyle}>Updated</th>
                    <th style={tableHeaderActionCellStyle}>Action</th>
                  </tr>
                </thead>
                <tbody>
                  {incidents.map((incident) => (
                    <tr
                      key={incident.incident_id}
                      data-testid={landingIncidentCardTestId(
                        incident.incident_id,
                      )}
                      style={tableRowStyle}
                    >
                      <td style={primaryCellStyle}>
                        <p style={monoMutedStyle}>{incident.incident_key}</p>
                        <p style={strongTextStyle}>{incident.title}</p>
                        <p style={metadataTextStyle}>
                          v{incident.incident_version}
                        </p>
                      </td>
                      <td style={tableCellStyle}>
                        <StatusBadge value={incident.status ?? "active"} />
                      </td>
                      <td style={tableCellStyle}>
                        <MutedUnset value={incident.current_phase} />
                      </td>
                      <td style={tableCellStyle}>
                        <MutedUnset value={incident.severity} />
                      </td>
                      <td style={tableCellStyle}>
                        <MutedUnset value={incident.tlp} />
                      </td>
                      <td style={tableCellStyle}>
                        <MutedUnset
                          value={incident.primary_external_case_ref}
                        />
                      </td>
                      <td
                        style={tableCellStyle}
                        title={incident.updated_at ?? undefined}
                      >
                        {formatNullableDateTime(incident.updated_at)}
                      </td>
                      <td style={tableActionCellStyle}>
                        <button
                          data-testid={landingIncidentOpenButtonTestId(
                            incident.incident_id,
                          )}
                          style={tableLinkButtonStyle}
                          type="button"
                          onClick={() => {
                            onOpenIncident(incident.incident_id);
                          }}
                        >
                          Open
                          <ChevronRight size={15} />
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : null}

          {hasMoreIncidents ? (
            <button
              style={secondaryButtonStyle}
              type="button"
              onClick={() => {
                void onLoadMore();
              }}
            >
              Load more incidents
            </button>
          ) : null}
        </section>
      </div>

      {createOpen ? (
        <div style={dialogBackdropStyle}>
          <section
            aria-label="New incident"
            aria-modal="true"
            role="dialog"
            style={createDialogStyle}
          >
            <header style={dialogHeaderStyle}>
              <div>
                <p style={sectionEyebrowStyle}>Create</p>
                <h3 style={subsectionTitleStyle}>New incident</h3>
              </div>
              <button
                aria-label="Close new incident"
                style={iconButtonStyle}
                type="button"
                onClick={() => {
                  setCreateOpen(false);
                }}
              >
                <X aria-hidden="true" size={16} />
              </button>
            </header>
            <div style={formGridStyle}>
              <label htmlFor={incidentKeyFieldId} style={labelBlockStyle}>
                Incident key
                <input
                  data-testid={phase1LandingTestId("incident-key")}
                  id={incidentKeyFieldId}
                  style={inputStyle}
                  value={createIncidentKey}
                  onChange={(event) => {
                    onCreateIncidentKeyChange(event.target.value);
                  }}
                  placeholder="IR-2026-001"
                />
              </label>
              <label htmlFor={incidentTitleFieldId} style={labelBlockStyle}>
                Title
                <input
                  data-testid={phase1LandingTestId("incident-title")}
                  id={incidentTitleFieldId}
                  style={inputStyle}
                  value={createIncidentTitle}
                  onChange={(event) => {
                    onCreateIncidentTitleChange(event.target.value);
                  }}
                  placeholder="Credential theft investigation"
                />
              </label>
            </div>
            <details style={detailsStyle}>
              <summary style={detailsSummaryStyle}>More details</summary>
              <div style={formGridStyle}>
                <label
                  htmlFor="incident-create-description"
                  style={labelBlockStyle}
                >
                  Description
                  <textarea
                    data-testid={phase1LandingTestId("create-description")}
                    id="incident-create-description"
                    style={textAreaStyle}
                    value={createIncidentDescription}
                    onChange={(event) => {
                      onCreateIncidentDescriptionChange(event.target.value);
                    }}
                  />
                </label>
                <label
                  htmlFor="incident-create-severity"
                  style={labelBlockStyle}
                >
                  Severity
                  <input
                    data-testid={phase1LandingTestId("create-severity")}
                    id="incident-create-severity"
                    style={inputStyle}
                    value={createIncidentSeverity}
                    onChange={(event) => {
                      onCreateIncidentSeverityChange(event.target.value);
                    }}
                  />
                </label>
                <label htmlFor="incident-create-tlp" style={labelBlockStyle}>
                  TLP
                  <select
                    data-testid={phase1LandingTestId("create-tlp")}
                    id="incident-create-tlp"
                    style={inputStyle}
                    value={createIncidentTLP}
                    onChange={(event) => {
                      onCreateIncidentTLPChange(event.target.value);
                    }}
                  >
                    <option value="">Unset</option>
                    <option value="TLP:CLEAR">Clear</option>
                    <option value="TLP:GREEN">Green</option>
                    <option value="TLP:AMBER">Amber</option>
                    <option value="TLP:AMBER+STRICT">Amber strict</option>
                    <option value="TLP:RED">Red</option>
                  </select>
                </label>
                <label
                  htmlFor="incident-create-current-phase"
                  style={labelBlockStyle}
                >
                  Current phase
                  <input
                    data-testid={phase1LandingTestId("create-current-phase")}
                    id="incident-create-current-phase"
                    style={inputStyle}
                    value={createIncidentCurrentPhase}
                    onChange={(event) => {
                      onCreateIncidentCurrentPhaseChange(event.target.value);
                    }}
                  />
                </label>
                <label
                  htmlFor="incident-create-external-case"
                  style={labelBlockStyle}
                >
                  External case
                  <input
                    data-testid={phase1LandingTestId("create-external-case")}
                    id="incident-create-external-case"
                    style={inputStyle}
                    value={createIncidentExternalCase}
                    onChange={(event) => {
                      onCreateIncidentExternalCaseChange(event.target.value);
                    }}
                  />
                </label>
              </div>
            </details>
            <button
              data-testid={phase1LandingTestId("create-submit-button")}
              style={primaryButtonStyle}
              type="button"
              onClick={() => {
                void submitCreate();
              }}
            >
              Create and open
            </button>
          </section>
        </div>
      ) : null}

      <p
        aria-live="polite"
        data-testid={phase1LandingTestId("status")}
        role="status"
        style={statusTextStyle}
      >
        {statusText}
      </p>
      <p
        aria-live="assertive"
        data-testid={phase1ErrorCodeTestId("landing")}
        role={error === null ? undefined : "alert"}
        style={errorTextStyle}
      >
        {publicErrorView(error)?.code ?? ""}
      </p>
      <PublicErrorSummary
        error={error}
        testIds={phase1ErrorSummaryTestIds("landing")}
      />
    </section>
  );
}

function MutedUnset({ value }: { value: string | null | undefined }) {
  if (value === null || typeof value === "undefined" || value === "") {
    return <span style={unsetValueStyle}>Not set</span>;
  }
  return <span>{value}</span>;
}

export function AccountProfilePanel({
  onRefreshShell,
}: {
  onRefreshShell: () => Promise<void> | void;
}) {
  const [profile, setProfile] = useState<AccountProfileResource | null>(null);
  const [displayName, setDisplayName] = useState("");
  const [status, setStatus] = useState("Loading account profile.");
  const [error, setError] = useState<APIError | null>(null);

  const loadProfile = useCallback(async () => {
    const result = await loadAccountProfile();
    const nextError = extractError(result.payload);
    setError(nextError);
    if (!result.ok) {
      setStatus("Account profile unavailable.");
      return;
    }
    const nextProfile = (result.payload as { data: AccountProfileResource })
      .data;
    setProfile(nextProfile);
    setDisplayName(nextProfile.display_name);
    setStatus("Account profile loaded.");
  }, []);

  useEffect(() => {
    void loadProfile();
  }, [loadProfile]);

  async function saveProfile() {
    if (profile === null) {
      return;
    }
    setStatus("Saving account profile.");
    const result = await patchAccountProfile({
      baseUserVersion: profile.user_version,
      displayName,
    });
    const nextError = extractError(result.payload);
    setError(nextError);
    if (!result.ok) {
      setStatus("Account profile save failed.");
      return;
    }
    const nextProfile = (result.payload as { data: AccountProfileResource })
      .data;
    setProfile(nextProfile);
    setDisplayName(nextProfile.display_name);
    setStatus("Account profile saved.");
    await onRefreshShell();
  }

  return (
    <section style={surfacePanelStyle}>
      <p style={sectionEyebrowStyle}>Profile</p>
      <h2 style={sectionTitleStyle}>Account profile</h2>
      <div style={definitionPanelStyle}>
        <div>
          <span style={definitionLabelStyle}>Email</span>
          <div
            data-testid={phase1AccountTestId("profile-email")}
            id="account-profile-email"
            style={definitionValueStyle}
          >
            {profile?.email ?? ""}
          </div>
        </div>
        <label htmlFor="account-profile-display-name" style={labelBlockStyle}>
          Display name
          <input
            data-testid={phase1AccountTestId("profile-display-name")}
            id="account-profile-display-name"
            style={inputStyle}
            value={displayName}
            onChange={(event) => {
              setDisplayName(event.target.value);
            }}
          />
        </label>
        <button
          data-testid={phase1AccountTestId("profile-save")}
          disabled={profile === null}
          style={primaryButtonStyle}
          type="button"
          onClick={() => {
            void saveProfile();
          }}
        >
          Save profile
        </button>
      </div>
      <p aria-live="polite" role="status" style={statusTextStyle}>
        {status}
      </p>
      <p
        aria-live="assertive"
        role={error === null ? undefined : "alert"}
        style={errorTextStyle}
      >
        {publicErrorView(error)?.code ?? ""}
      </p>
    </section>
  );
}

export function AccountAppearancePanel() {
  const [preferences, setPreferences] =
    useState<AccountPreferencesResource | null>(null);
  const [densityMode, setDensityMode] = useState<DensityMode | "">("");
  const [status, setStatus] = useState("Loading account appearance.");
  const [error, setError] = useState<APIError | null>(null);

  const loadPreferences = useCallback(async () => {
    const result = await loadAccountPreferences();
    const nextError = extractError(result.payload);
    setError(nextError);
    if (!result.ok) {
      setStatus("Account appearance unavailable.");
      return;
    }
    const nextPreferences = (
      result.payload as { data: AccountPreferencesResource }
    ).data;
    setPreferences(nextPreferences);
    setDensityMode(nextPreferences.density_mode ?? "");
    setStatus("Account appearance loaded.");
  }, []);

  useEffect(() => {
    void loadPreferences();
  }, [loadPreferences]);

  async function savePreferences() {
    if (preferences === null) {
      return;
    }
    setStatus("Saving account appearance.");
    const result = await putAccountPreferences({
      basePreferencesVersion: preferences.preferences_version,
      densityMode: densityMode === "" ? null : densityMode,
    });
    const nextError = extractError(result.payload);
    setError(nextError);
    if (!result.ok) {
      setStatus("Account appearance save failed.");
      return;
    }
    const nextPreferences = (
      result.payload as { data: AccountPreferencesResource }
    ).data;
    setPreferences(nextPreferences);
    setDensityMode(nextPreferences.density_mode ?? "");
    setStatus("Account appearance saved.");
  }

  return (
    <section style={surfacePanelStyle}>
      <p style={sectionEyebrowStyle}>Appearance</p>
      <h2 style={sectionTitleStyle}>Density</h2>
      <div style={segmentedFormStyle}>
        <label htmlFor="account-density-mode" style={labelBlockStyle}>
          Density
          <select
            data-testid={phase1AccountTestId("appearance-density-mode")}
            id="account-density-mode"
            style={inputStyle}
            value={densityMode}
            onChange={(event) => {
              setDensityMode(event.target.value as DensityMode | "");
            }}
          >
            <option value="">Use surface default</option>
            <option value="compact">Compact</option>
            <option value="default">Default</option>
            <option value="comfortable">Comfortable</option>
          </select>
        </label>
        <button
          data-testid={phase1AccountTestId("appearance-save")}
          disabled={preferences === null}
          style={primaryButtonStyle}
          type="button"
          onClick={() => {
            void savePreferences();
          }}
        >
          Save appearance
        </button>
      </div>
      <p aria-live="polite" role="status" style={statusTextStyle}>
        {status}
      </p>
      <p
        aria-live="assertive"
        role={error === null ? undefined : "alert"}
        style={errorTextStyle}
      >
        {publicErrorView(error)?.code ?? ""}
      </p>
    </section>
  );
}

export function AdministrativeAuditPanel() {
  const [events, setEvents] = useState<AdministrativeAuditEvent[]>([]);
  const [expandedAuditEventID, setExpandedAuditEventID] = useState<
    string | null
  >(null);
  const [actorUserID, setActorUserID] = useState("");
  const [actionCode, setActionCode] = useState("");
  const [targetKind, setTargetKind] = useState("");
  const [targetID, setTargetID] = useState("");
  const [occurredAtGTE, setOccurredAtGTE] = useState("");
  const [occurredAtLT, setOccurredAtLT] = useState("");
  const [status, setStatus] = useState("Administrative audit idle.");
  const [error, setError] = useState<APIError | null>(null);
  const initialAuditLoadedRef = useRef(false);

  const loadAudit = useCallback(async () => {
    setStatus("Loading administrative audit.");
    const params = new URLSearchParams({ limit: "100" });
    for (const [key, value] of [
      ["actor_user_id", actorUserID],
      ["action_code", actionCode],
      ["target_kind", targetKind],
      ["target_id", targetID],
      ["occurred_at_gte", occurredAtGTE],
      ["occurred_at_lt", occurredAtLT],
    ] as const) {
      const trimmed = value.trim();
      if (trimmed !== "") {
        params.set(key, trimmed);
      }
    }
    const result = await fetchJSON<{
      data:
        | { administrative_audit_events: AdministrativeAuditEvent[] }
        | { audit_events: AdministrativeAuditEvent[] };
    }>(`/api/v1/administrative-audit-events?${params.toString()}`);
    const nextError = extractError(result.payload);
    setError(nextError);
    if (!result.ok) {
      setStatus("Administrative audit unavailable.");
      return;
    }
    const data = (
      result.payload as {
        data:
          | { administrative_audit_events: AdministrativeAuditEvent[] }
          | { audit_events: AdministrativeAuditEvent[] };
      }
    ).data;
    setEvents(
      "audit_events" in data
        ? data.audit_events
        : data.administrative_audit_events,
    );
    setStatus("Administrative audit loaded.");
  }, [
    actionCode,
    actorUserID,
    occurredAtGTE,
    occurredAtLT,
    targetID,
    targetKind,
  ]);

  useEffect(() => {
    if (initialAuditLoadedRef.current) {
      return;
    }
    initialAuditLoadedRef.current = true;
    void loadAudit();
  }, [loadAudit]);

  return (
    <section style={surfacePanelStyle}>
      <div style={surfaceHeaderStyle}>
        <div>
          <p style={sectionEyebrowStyle}>Administrative audit</p>
          <h2 style={sectionTitleStyle}>Deployment events</h2>
        </div>
        <button
          style={secondaryButtonStyle}
          type="button"
          onClick={() => {
            void loadAudit();
          }}
        >
          <RefreshCw size={15} />
          Refresh
        </button>
      </div>
      <div style={auditFilterGridStyle}>
        <label style={labelBlockStyle}>
          Actor user ID
          <input
            aria-label="Actor user id"
            style={inputStyle}
            value={actorUserID}
            onChange={(event) => setActorUserID(event.target.value)}
          />
        </label>
        <label style={labelBlockStyle}>
          Action code
          <input
            aria-label="Action code"
            style={inputStyle}
            value={actionCode}
            onChange={(event) => setActionCode(event.target.value)}
          />
        </label>
        <label style={labelBlockStyle}>
          Target kind
          <select
            aria-label="Target kind"
            style={inputStyle}
            value={targetKind}
            onChange={(event) => setTargetKind(event.target.value)}
          >
            <option value="">Any target kind</option>
            <option value="user">User</option>
            <option value="account_preferences">Account preferences</option>
            <option value="auth_binding">Auth binding</option>
            <option value="backup_set">Backup set</option>
            <option value="restore_operation">Restore operation</option>
            <option value="legacy_administrative_event">Legacy event</option>
          </select>
        </label>
        <label style={labelBlockStyle}>
          Target ID
          <input
            aria-label="Target id"
            style={inputStyle}
            value={targetID}
            onChange={(event) => setTargetID(event.target.value)}
          />
        </label>
        <label style={labelBlockStyle}>
          Occurred at or after
          <input
            aria-label="Occurred at or after"
            style={inputStyle}
            value={occurredAtGTE}
            onChange={(event) => setOccurredAtGTE(event.target.value)}
          />
        </label>
        <label style={labelBlockStyle}>
          Occurred before
          <input
            aria-label="Occurred before"
            style={inputStyle}
            value={occurredAtLT}
            onChange={(event) => setOccurredAtLT(event.target.value)}
          />
        </label>
      </div>
      <div style={tableShellStyle}>
        <table style={dataTableStyle}>
          <thead>
            <tr>
              <th style={tableHeaderCellStyle}>Occurred</th>
              <th style={tableHeaderCellStyle}>Action</th>
              <th style={tableHeaderCellStyle}>Actor</th>
              <th style={tableHeaderCellStyle}>Target</th>
              <th style={tableHeaderCellStyle}>Reason</th>
              <th style={tableHeaderCellStyle}>Changes</th>
              <th style={tableHeaderActionCellStyle}>Details</th>
            </tr>
          </thead>
          <tbody>
            {events.map((event) => {
              const expanded = expandedAuditEventID === event.audit_event_id;
              return (
                <Fragment key={event.audit_event_id}>
                  <tr key={event.audit_event_id} style={tableRowStyle}>
                    <td style={primaryCellStyle}>
                      {formatNullableDateTime(event.occurred_at)}
                      <p style={metadataTextStyle}>{event.source ?? ""}</p>
                    </td>
                    <td style={tableCellStyle}>
                      {formatAuditActionLabel(event.action_code)}
                    </td>
                    <td style={tableCellStyle}>
                      {formatAuditActor(event.actor_kind, event.actor_user_id)}
                    </td>
                    <td style={tableCellStyle}>{formatAuditTarget(event)}</td>
                    <td style={tableCellStyle}>
                      {event.reason_code ?? "No reason"}
                    </td>
                    <td style={tableCellStyle}>{event.changes.length}</td>
                    <td style={tableActionCellStyle}>
                      <button
                        aria-expanded={expanded}
                        style={tableLinkButtonStyle}
                        type="button"
                        onClick={() => {
                          setExpandedAuditEventID(
                            expanded ? null : event.audit_event_id,
                          );
                        }}
                      >
                        {expanded ? "Hide" : "Inspect"}
                        <ChevronRight size={15} />
                      </button>
                    </td>
                  </tr>
                  {expanded ? (
                    <tr key={`${event.audit_event_id}-details`}>
                      <td colSpan={7} style={auditDetailCellStyle}>
                        <div style={auditDetailPanelStyle}>
                          <div style={auditDetailMetaGridStyle}>
                            <div>
                              <span style={definitionLabelStyle}>
                                Audit event ID
                              </span>
                              <div style={monoMutedStyle}>
                                {event.audit_event_id}
                              </div>
                            </div>
                            <div>
                              <span style={definitionLabelStyle}>
                                Raw action code
                              </span>
                              <div style={monoMutedStyle}>
                                {event.action_code}
                              </div>
                            </div>
                            <div>
                              <span style={definitionLabelStyle}>
                                Target kind
                              </span>
                              <div style={monoMutedStyle}>
                                {event.target_kind}
                              </div>
                            </div>
                            <div>
                              <span style={definitionLabelStyle}>
                                Target ID
                              </span>
                              <div style={monoMutedStyle}>
                                {event.target_id ?? "No target ID"}
                              </div>
                            </div>
                          </div>
                          {event.changes.length > 0 ? (
                            <table style={auditChangeTableStyle}>
                              <thead>
                                <tr>
                                  <th style={tableHeaderCellStyle}>Field</th>
                                  <th style={tableHeaderCellStyle}>Before</th>
                                  <th style={tableHeaderCellStyle}>After</th>
                                </tr>
                              </thead>
                              <tbody>
                                {event.changes.map((change) => (
                                  <tr
                                    key={`${event.audit_event_id}-${change.field_path}`}
                                  >
                                    <td style={primaryCellStyle}>
                                      <span style={monoMutedStyle}>
                                        {change.field_path}
                                      </span>
                                    </td>
                                    <td style={tableCellStyle}>
                                      {renderAuditValue(
                                        change.value_state,
                                        change.before,
                                      )}
                                    </td>
                                    <td style={tableCellStyle}>
                                      {renderAuditValue(
                                        change.value_state,
                                        change.after,
                                      )}
                                    </td>
                                  </tr>
                                ))}
                              </tbody>
                            </table>
                          ) : (
                            <p style={emptyStateStyle}>
                              No field-level changes were published for this
                              event.
                            </p>
                          )}
                        </div>
                      </td>
                    </tr>
                  ) : null}
                </Fragment>
              );
            })}
          </tbody>
        </table>
      </div>
      {events.length === 0 ? (
        <p style={emptyStateStyle}>No administrative audit events loaded.</p>
      ) : null}
      <p aria-live="polite" role="status" style={statusTextStyle}>
        {status}
      </p>
      <p
        aria-live="assertive"
        role={error === null ? undefined : "alert"}
        style={errorTextStyle}
      >
        {publicErrorView(error)?.code ?? ""}
      </p>
    </section>
  );
}

export function IncidentImportPanel({
  onOpenImportedIncident,
}: {
  onOpenImportedIncident: (incidentId: string) => void;
}) {
  const [file, setFile] = useState<File | null>(null);
  const [importDialogOpen, setImportDialogOpen] = useState(false);
  const [job, setJob] = useState<JobResource | null>(null);
  const [status, setStatus] = useState("Incident import idle.");
  const [error, setError] = useState<APIError | null>(null);
  const pollTimer = useRef<number | null>(null);
  const importedIncidentID = importedIncidentIDFromJob(job);

  const loadJob = useCallback(async (jobID: string) => {
    const result = await fetchJSON<{ data: JobResource }>(
      `/api/v1/jobs/${jobID}`,
    );
    if (!result.ok) {
      setError(extractError(result.payload));
      return;
    }
    const nextJob = (result.payload as { data: JobResource }).data;
    setJob(nextJob);
    setStatus(`Incident import job ${nextJob.status ?? "updated"}.`);
  }, []);

  useEffect(() => {
    if (
      job?.job_id === undefined ||
      job.status === undefined ||
      terminalJobStates.has(job.status)
    ) {
      return;
    }
    pollTimer.current = window.setTimeout(() => {
      void loadJob(job.job_id ?? "");
    }, 1000);
    return () => {
      if (pollTimer.current !== null) {
        window.clearTimeout(pollTimer.current);
        pollTimer.current = null;
      }
    };
  }, [job, loadJob]);

  async function submitImport() {
    if (file === null) {
      setStatus("Select an incident bundle first.");
      return;
    }
    const form = new FormData();
    form.append(
      "metadata",
      new Blob(
        [
          JSON.stringify({
            client_txn_id: clientTxnID("incident-import"),
          }),
        ],
        { type: "application/json" },
      ),
    );
    form.append("file", file);
    const headers = new Headers();
    const csrfToken = readCookie(csrfCookieName);
    if (csrfToken !== null && csrfToken !== "") {
      headers.set(csrfHeaderName, csrfToken);
    }
    setStatus("Submitting incident import.");
    const response = await fetch("/api/v1/incident-bundles/import", {
      method: "POST",
      credentials: "include",
      headers,
      body: form,
    });
    const payload = (await response.json()) as {
      data?: JobResource;
      error?: APIError;
    };
    if (!response.ok) {
      setError(extractError(payload));
      setStatus("Incident import failed to start.");
      return;
    }
    const nextJob = payload.data ?? null;
    setError(null);
    setJob(nextJob);
    setStatus(
      `Incident import queued${nextJob?.job_id ? `: ${nextJob.job_id}` : "."}`,
    );
    setImportDialogOpen(false);
    setFile(null);
    if (nextJob?.job_id) {
      void loadJob(nextJob.job_id);
    }
  }

  return (
    <section style={surfacePanelStyle}>
      <div style={surfaceHeaderStyle}>
        <div>
          <p style={sectionEyebrowStyle}>Incident portability</p>
          <h2 style={sectionTitleStyle}>Incident import</h2>
        </div>
        <button
          style={primaryButtonStyle}
          type="button"
          onClick={() => setImportDialogOpen(true)}
        >
          Import incident
        </button>
      </div>
      <div style={jobPanelStyle}>
        <p style={strongTextStyle}>Import progress</p>
        <p style={metadataTextStyle}>
          {job === null
            ? "No import job is active."
            : `${job.status ?? "queued"} · ${job.progress?.completed ?? 0}/${job.progress?.total ?? "?"}`}
        </p>
        {importedIncidentID !== null ? (
          <button
            style={primaryButtonStyle}
            type="button"
            onClick={() => {
              onOpenImportedIncident(importedIncidentID);
            }}
          >
            Open imported incident
          </button>
        ) : null}
      </div>
      {importDialogOpen ? (
        <div style={dialogBackdropStyle}>
          <section
            aria-label="Import incident bundle"
            aria-modal="true"
            role="dialog"
            style={createDialogStyle}
          >
            <header style={dialogHeaderStyle}>
              <div>
                <p style={sectionEyebrowStyle}>Staged import</p>
                <h3 style={subsectionTitleStyle}>Select incident bundle</h3>
              </div>
              <button
                aria-label="Close incident import"
                style={iconButtonStyle}
                type="button"
                onClick={() => setImportDialogOpen(false)}
              >
                <X aria-hidden="true" size={16} />
              </button>
            </header>
            <div style={formGridStyle}>
              <label style={labelBlockStyle}>
                Bundle file
                <input
                  aria-label="Incident bundle file"
                  style={inputStyle}
                  type="file"
                  onChange={(event) => {
                    setFile(event.currentTarget.files?.[0] ?? null);
                  }}
                />
              </label>
            </div>
            <div style={buttonRowEndStyle}>
              <button
                style={secondaryButtonStyle}
                type="button"
                onClick={() => setImportDialogOpen(false)}
              >
                Cancel
              </button>
              <button
                style={primaryButtonStyle}
                type="button"
                onClick={() => {
                  void submitImport();
                }}
              >
                Start import
              </button>
            </div>
          </section>
        </div>
      ) : null}
      <p aria-live="polite" role="status" style={statusTextStyle}>
        {status}
      </p>
      <p
        aria-live="assertive"
        role={error === null ? undefined : "alert"}
        style={errorTextStyle}
      >
        {publicErrorView(error)?.code ?? ""}
      </p>
    </section>
  );
}

function StatusBadge({ value }: { value: string }) {
  const isClosed = value === "closed";
  return (
    <span style={isClosed ? closedBadgeStyle : activeBadgeStyle}>{value}</span>
  );
}

function PublicErrorSummary({
  error,
  testIds,
}: {
  error: APIError | null;
  testIds: {
    readonly container: string;
    readonly details: string;
    readonly message: string;
  };
}) {
  const view = publicErrorView(error);
  return (
    <div
      data-testid={testIds.container}
      role={view === null ? undefined : "alert"}
      style={publicErrorStyle}
    >
      <p data-testid={testIds.message} style={errorMessageStyle}>
        {view?.statusText ?? ""}
      </p>
      <p data-testid={testIds.details} style={errorDetailStyle}>
        {view?.details
          .map((detail) => `${detail.label}: ${detail.value}`)
          .join(" · ") ?? ""}
      </p>
    </div>
  );
}

function renderAuditValue(valueState: "redacted" | "visible", value: unknown) {
  if (valueState === "redacted") {
    return <span style={redactedBadgeStyle}>Redacted</span>;
  }
  if (value === null) {
    return <span style={nullValueStyle}>null</span>;
  }
  if (typeof value === "string") {
    return <span>{value}</span>;
  }
  return <span>{JSON.stringify(value)}</span>;
}

function formatAuditActionLabel(actionCode: string) {
  const label = auditActionLabels[actionCode];
  return label ?? actionCode;
}

function formatAuditActor(
  actorKind: AdministrativeAuditEvent["actor_kind"],
  actorUserID: string | null,
) {
  if (actorUserID !== null && actorUserID !== "") {
    return actorUserID;
  }
  return actorKind ?? "system";
}

function formatAuditTarget(event: AdministrativeAuditEvent) {
  const targetID = event.target_id ?? "";
  switch (event.target_kind) {
    case "user":
      return targetID === "" ? "User" : `User ${targetID}`;
    case "account_preferences":
      return targetID === ""
        ? "Account preferences"
        : `Account preferences for ${targetID}`;
    case "auth_binding":
      return targetID === "" ? "Enterprise authentication binding" : targetID;
    case "backup_set":
      return targetID === "" ? "Backup set" : `Backup set ${targetID}`;
    case "restore_operation":
      return targetID === ""
        ? "Restore operation"
        : `Restore operation ${targetID}`;
    case "legacy_administrative_event":
      return "Legacy event";
    default:
      return targetID === ""
        ? event.target_kind
        : `${event.target_kind}: ${targetID}`;
  }
}

function formatNullableDateTime(value: string | null | undefined) {
  if (value === null || typeof value === "undefined" || value === "") {
    return "Not recorded";
  }
  return value;
}

const landingAdminShellStyle: CSSProperties = {
  width: "100%",
  minHeight: "100vh",
  display: "grid",
  gridTemplateRows: "auto minmax(0, 1fr)",
  background: "var(--ct-colors-canvas)",
  overflow: "hidden",
};

const incidentDirectoryShellStyle: CSSProperties = {
  ...landingAdminShellStyle,
  overflow: "auto",
};

const landingAdminHeaderStyle: CSSProperties = {
  display: "grid",
  gridTemplateColumns: "minmax(12rem, 1fr) auto minmax(14rem, 1fr)",
  gap: "var(--ct-spacing-lg)",
  alignItems: "center",
  minHeight: "var(--ct-layout-topBarHeight)",
  padding: "var(--ct-spacing-sm) var(--ct-spacing-lg)",
  borderBottom: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
};

const brandBlockStyle: CSSProperties = {
  minWidth: 0,
};

const landingEyebrowStyle: CSSProperties = {
  margin: 0,
  fontSize: "0.72rem",
  letterSpacing: 0,
  textTransform: "uppercase",
  color: "var(--ct-colors-accent)",
};

const landingAdminTitleStyle: CSSProperties = {
  margin: "0.18rem 0 0",
  fontSize: "var(--ct-typography-surface-title-fontSize)",
  lineHeight: "var(--ct-typography-surface-title-lineHeight)",
};

const landingAdminHeaderMetaStyle: CSSProperties = {
  display: "grid",
  gridTemplateColumns: "repeat(2, minmax(9rem, auto))",
  gap: "var(--ct-spacing-lg)",
  margin: 0,
};

const landingAdminMetaValueStyle: CSSProperties = {
  margin: "0.2rem 0 0",
  color: "var(--ct-colors-ink)",
  fontWeight: 700,
};

const landingToolbarLabelStyle: CSSProperties = {
  margin: 0,
  fontSize: "0.68rem",
  letterSpacing: 0,
  textTransform: "uppercase",
  color: "var(--ct-colors-ink-subtle)",
};

const landingAccountNavStyle: CSSProperties = {
  display: "flex",
  flexWrap: "wrap",
  justifyContent: "flex-end",
  gap: "var(--ct-spacing-xs)",
};

const accountMenuAnchorStyle: CSSProperties = {
  position: "relative",
  display: "inline-flex",
  justifyContent: "flex-end",
  minInlineSize: 0,
  margin: 0,
  padding: 0,
  border: 0,
};

const accountMenuTriggerStyle: CSSProperties = {
  display: "inline-flex",
  alignItems: "center",
  maxWidth: "min(22rem, 100%)",
  gap: "0.45rem",
  padding: "0.5rem 0.65rem",
  borderRadius: "var(--ct-rounded-sm)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-ink)",
  fontWeight: 700,
  cursor: "pointer",
};

const accountMenuTriggerTextStyle: CSSProperties = {
  minWidth: 0,
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap",
};

const accountMenuStyle: CSSProperties = {
  position: "absolute",
  insetBlockStart: "calc(100% + 0.35rem)",
  insetInlineEnd: 0,
  zIndex: 20,
  minWidth: "18rem",
  display: "grid",
  gap: "0.2rem",
  padding: "0.4rem",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  background: "var(--ct-colors-surface-2)",
  boxShadow: "var(--ct-elevation-panel)",
};

const accountMenuItemStyle: CSSProperties = {
  width: "100%",
  padding: "0.55rem 0.65rem",
  border: "none",
  borderRadius: "var(--ct-rounded-sm)",
  background: "transparent",
  color: "var(--ct-colors-ink-muted)",
  font: "inherit",
  fontWeight: 700,
  textAlign: "left",
  cursor: "pointer",
};

const accountMenuStatusItemStyle: CSSProperties = {
  width: "100%",
  padding: "0.55rem 0.65rem",
  borderRadius: "var(--ct-rounded-sm)",
  background: "var(--ct-colors-surface-1)",
  color: "var(--ct-colors-ink-muted)",
  fontWeight: 700,
  textAlign: "left",
};

const accountSubmenuStyle: CSSProperties = {
  display: "grid",
  gap: "0.2rem",
  padding: "0.2rem 0 0.25rem 0.55rem",
};

const accountSubmenuItemStyle: CSSProperties = {
  display: "grid",
  gap: "0.16rem",
  width: "100%",
  padding: "0.5rem 0.6rem",
  border: "none",
  borderRadius: "var(--ct-rounded-sm)",
  background: "transparent",
  color: "var(--ct-colors-ink-muted)",
  font: "inherit",
  textAlign: "left",
  cursor: "pointer",
};

const accountSubmenuItemSelectedStyle: CSSProperties = {
  ...accountSubmenuItemStyle,
  background: "var(--ct-colors-surface-3)",
  color: "var(--ct-colors-ink)",
};

const accountSubmenuItemLabelStyle: CSSProperties = {
  fontWeight: 700,
};

const accountSubmenuItemDescriptionStyle: CSSProperties = {
  color: "var(--ct-colors-ink-subtle)",
  fontSize: "0.76rem",
};

const accountMenuSeparatorStyle: CSSProperties = {
  height: 1,
  margin: "0.25rem 0",
  background: "var(--ct-colors-border-muted)",
};

const landingAccountNavButtonStyle: CSSProperties = {
  display: "inline-flex",
  alignItems: "center",
  gap: "0.42rem",
  padding: "0.5rem 0.65rem",
  borderRadius: "var(--ct-rounded-sm)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-ink-muted)",
  fontWeight: 700,
  cursor: "pointer",
};

const landingAccountNavButtonSelectedStyle: CSSProperties = {
  ...landingAccountNavButtonStyle,
  border: "var(--ct-border-strong)",
  color: "var(--ct-colors-ink)",
  background: "var(--ct-colors-surface-3)",
};

const landingAdminWorkspaceStyle: CSSProperties = {
  minHeight: 0,
  display: "grid",
  gridTemplateColumns: "16rem minmax(0, 1fr)",
  background: "var(--ct-colors-canvas)",
};

const landingAdminMenuStyle: CSSProperties = {
  minHeight: 0,
  padding: "var(--ct-spacing-md)",
  borderRight: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
  overflow: "auto",
};

const landingAdminMenuItemsStyle: CSSProperties = {
  display: "grid",
  gap: "var(--ct-spacing-lg)",
};

const menuGroupStyle: CSSProperties = {
  display: "grid",
  gap: "var(--ct-spacing-sm)",
};

const menuGroupTitleStyle: CSSProperties = {
  margin: 0,
  padding: "0 var(--ct-spacing-xs)",
  fontSize: "0.68rem",
  letterSpacing: "0.14em",
  textTransform: "uppercase",
  color: "var(--ct-colors-ink-subtle)",
};

const menuGroupItemsStyle: CSSProperties = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
};

const landingAdminMenuItemStyle: CSSProperties = {
  display: "grid",
  gridTemplateColumns: "1.1rem minmax(0, 1fr)",
  gap: "var(--ct-spacing-sm)",
  alignItems: "start",
  width: "100%",
  minWidth: 0,
  padding: "var(--ct-spacing-sm)",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  background: "transparent",
  color: "var(--ct-colors-ink-muted)",
  textAlign: "left",
  cursor: "pointer",
};

const landingAdminMenuItemSelectedStyle: CSSProperties = {
  ...landingAdminMenuItemStyle,
  border: "var(--ct-border-strong)",
  background: "var(--ct-colors-surface-3)",
  color: "var(--ct-colors-ink)",
  boxShadow: "inset 3px 0 0 var(--ct-colors-accent)",
};

const landingAdminMenuItemTextStyle: CSSProperties = {
  display: "grid",
  gap: "0.2rem",
  minWidth: 0,
};

const landingAdminMenuItemLabelStyle: CSSProperties = {
  fontWeight: 700,
  overflowWrap: "anywhere",
};

const landingAdminMenuItemDescriptionStyle: CSSProperties = {
  color: "var(--ct-colors-ink-subtle)",
  fontSize: "0.76rem",
  overflowWrap: "anywhere",
};

const landingAdminContentStyle: CSSProperties = {
  minWidth: 0,
  minHeight: 0,
  overflow: "auto",
  width: "100%",
  maxWidth: "1600px",
  justifySelf: "center",
  boxSizing: "border-box",
};

const surfacePanelStyle: CSSProperties = {
  minWidth: 0,
  minHeight: 0,
  display: "grid",
  gap: "var(--ct-spacing-md)",
  padding: "var(--ct-spacing-md)",
  color: "var(--ct-colors-ink)",
};

const surfaceHeaderStyle: CSSProperties = {
  display: "flex",
  justifyContent: "space-between",
  alignItems: "center",
  gap: "var(--ct-spacing-md)",
  padding: "0 0 var(--ct-spacing-sm)",
  borderBottom: "var(--ct-border-hairline)",
};

const headerActionRowStyle: CSSProperties = {
  display: "flex",
  alignItems: "center",
  gap: "var(--ct-spacing-sm)",
};

const sectionEyebrowStyle: CSSProperties = {
  margin: 0,
  fontSize: "0.72rem",
  letterSpacing: "0.14em",
  textTransform: "uppercase",
  color: "var(--ct-colors-ink-subtle)",
};

const sectionTitleStyle: CSSProperties = {
  margin: "0.2rem 0 0",
  fontSize: "1.25rem",
};

const subsectionTitleStyle: CSSProperties = {
  margin: "0.2rem 0 0",
  fontSize: "var(--ct-typography-section-heading-fontSize)",
};

const incidentWorkspaceStyle: CSSProperties = {
  minWidth: 0,
  minHeight: 0,
  display: "grid",
  gridTemplateColumns: "minmax(0, 1fr)",
  gap: "var(--ct-spacing-md)",
  alignItems: "start",
};

const incidentDirectoryStyle: CSSProperties = {
  minWidth: 0,
  display: "grid",
  gap: "var(--ct-spacing-md)",
};

const toolbarGridStyle: CSSProperties = {
  display: "grid",
  gridTemplateColumns: "minmax(18rem, 1fr) minmax(10rem, 14rem)",
  gap: "var(--ct-spacing-sm)",
  alignItems: "end",
};

const labelBlockStyle: CSSProperties = {
  display: "grid",
  gap: "0.35rem",
  minWidth: 0,
  fontSize: "0.82rem",
  fontWeight: 700,
  color: "var(--ct-colors-ink-muted)",
};

const searchInputShellStyle: CSSProperties = {
  display: "grid",
  gridTemplateColumns: "1rem minmax(0, 1fr)",
  gap: "0.55rem",
  alignItems: "center",
  padding: "0.72rem 0.85rem",
  borderRadius: "var(--ct-component-text-input-rounded)",
  border: "var(--ct-component-text-input-border)",
  background: "var(--ct-component-text-input-backgroundColor)",
};

const searchInputStyle: CSSProperties = {
  minWidth: 0,
  width: "100%",
  border: "none",
  outline: "none",
  padding: 0,
  background: "transparent",
  color: "var(--ct-component-text-input-textColor)",
};

const inputStyle: CSSProperties = {
  boxSizing: "border-box",
  width: "100%",
  maxWidth: "100%",
  minWidth: 0,
  padding: "var(--ct-component-text-input-padding)",
  borderRadius: "var(--ct-component-text-input-rounded)",
  border: "var(--ct-component-text-input-border)",
  background: "var(--ct-component-text-input-backgroundColor)",
  color: "var(--ct-component-text-input-textColor)",
  fontSize: "0.92rem",
};

const textAreaStyle: CSSProperties = {
  ...inputStyle,
  minHeight: "6rem",
  resize: "vertical",
};

const formGridStyle: CSSProperties = {
  display: "grid",
  gap: "var(--ct-spacing-sm)",
};

const segmentedFormStyle: CSSProperties = {
  display: "grid",
  gridTemplateColumns: "minmax(16rem, 24rem) auto",
  gap: "var(--ct-spacing-md)",
  alignItems: "end",
  marginTop: "var(--ct-spacing-md)",
};

const detailsStyle: CSSProperties = {
  marginTop: "0.25rem",
};

const detailsSummaryStyle: CSSProperties = {
  cursor: "pointer",
  color: "var(--ct-colors-ink-muted)",
  fontWeight: 700,
};

const dialogBackdropStyle: CSSProperties = {
  position: "fixed",
  inset: 0,
  zIndex: 40,
  display: "grid",
  placeItems: "center",
  padding: "var(--ct-spacing-lg)",
  background: "rgba(10, 13, 18, 0.68)",
};

const createDialogStyle: CSSProperties = {
  width: "min(48rem, 100%)",
  maxHeight: "calc(100vh - 3rem)",
  overflow: "auto",
  display: "grid",
  gap: "var(--ct-spacing-md)",
  padding: "var(--ct-spacing-lg)",
  border: "var(--ct-border-strong)",
  borderRadius: "var(--ct-rounded-md)",
  background: "var(--ct-colors-surface-1)",
  boxShadow: "var(--ct-elevation-panel)",
};

const dialogHeaderStyle: CSSProperties = {
  display: "flex",
  alignItems: "flex-start",
  justifyContent: "space-between",
  gap: "var(--ct-spacing-md)",
};

const iconButtonStyle: CSSProperties = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  width: "2rem",
  height: "2rem",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-ink)",
  cursor: "pointer",
};

const buttonRowEndStyle: CSSProperties = {
  display: "flex",
  justifyContent: "flex-end",
  flexWrap: "wrap",
  gap: "var(--ct-spacing-sm)",
};

const primaryButtonStyle: CSSProperties = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  gap: "0.45rem",
  padding: "var(--ct-component-button-primary-padding)",
  borderRadius: "var(--ct-component-button-primary-rounded)",
  border: "none",
  background: "var(--ct-component-button-primary-backgroundColor)",
  color: "var(--ct-component-button-primary-textColor)",
  fontWeight: 700,
  cursor: "pointer",
};

const secondaryButtonStyle: CSSProperties = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  gap: "0.45rem",
  padding: "var(--ct-component-button-secondary-padding)",
  borderRadius: "var(--ct-component-button-secondary-rounded)",
  border: "var(--ct-component-button-secondary-border)",
  background: "var(--ct-component-button-secondary-backgroundColor)",
  color: "var(--ct-component-button-secondary-textColor)",
  fontWeight: 700,
  cursor: "pointer",
};

const countPillStyle: CSSProperties = {
  margin: 0,
  minWidth: "2.5rem",
  padding: "0.45rem 0.75rem",
  borderRadius: "var(--ct-rounded-pill)",
  background: "var(--ct-colors-surface-3)",
  textAlign: "center",
  fontWeight: 700,
};

const inlineStatusStyle: CSSProperties = {
  margin: 0,
  color: "var(--ct-colors-ink-muted)",
};

const emptyStateStyle: CSSProperties = {
  margin: "var(--ct-spacing-md) 0 0",
  color: "var(--ct-colors-ink-muted)",
};

const tableShellStyle: CSSProperties = {
  minWidth: 0,
  overflowX: "auto",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  background: "var(--ct-colors-surface-1)",
};

const dataTableStyle: CSSProperties = {
  width: "100%",
  borderCollapse: "collapse",
  minWidth: "62rem",
};

const tableHeaderCellStyle: CSSProperties = {
  padding: "0.7rem 0.85rem",
  borderBottom: "var(--ct-border-hairline)",
  color: "var(--ct-colors-ink-subtle)",
  fontSize: "0.68rem",
  letterSpacing: "0.12em",
  textTransform: "uppercase",
  textAlign: "left",
  whiteSpace: "nowrap",
};

const tableHeaderActionCellStyle: CSSProperties = {
  ...tableHeaderCellStyle,
  textAlign: "right",
};

const tableRowStyle: CSSProperties = {
  borderBottom: "var(--ct-border-hairline)",
};

const tableCellStyle: CSSProperties = {
  padding: "0.75rem 0.85rem",
  color: "var(--ct-colors-ink-muted)",
  verticalAlign: "top",
  fontSize: "0.84rem",
};

const primaryCellStyle: CSSProperties = {
  ...tableCellStyle,
  color: "var(--ct-colors-ink)",
  minWidth: "14rem",
};

const tableActionCellStyle: CSSProperties = {
  ...tableCellStyle,
  textAlign: "right",
};

const tableLinkButtonStyle: CSSProperties = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  gap: "0.35rem",
  padding: "0.25rem 0",
  border: "none",
  background: "transparent",
  color: "var(--ct-colors-ink)",
  fontWeight: 700,
  cursor: "pointer",
};

const strongTextStyle: CSSProperties = {
  margin: 0,
  color: "var(--ct-colors-ink)",
  fontWeight: 700,
  overflowWrap: "anywhere",
};

const metadataTextStyle: CSSProperties = {
  margin: "0.24rem 0 0",
  color: "var(--ct-colors-ink-subtle)",
  fontSize: "0.78rem",
  overflowWrap: "anywhere",
};

const monoMutedStyle: CSSProperties = {
  margin: 0,
  color: "var(--ct-colors-ink-subtle)",
  fontFamily: "var(--ct-typography-mono-fontFamily)",
  fontSize: "0.78rem",
  overflowWrap: "anywhere",
};

const unsetValueStyle: CSSProperties = {
  color: "var(--ct-colors-ink-subtle)",
  fontStyle: "italic",
};

const activeBadgeStyle: CSSProperties = {
  display: "inline-flex",
  padding: "0.18rem 0.5rem",
  borderRadius: "var(--ct-rounded-pill)",
  color: "var(--ct-colors-semantic-success)",
  background:
    "color-mix(in srgb, var(--ct-colors-semantic-success) 14%, transparent)",
  fontWeight: 700,
};

const closedBadgeStyle: CSSProperties = {
  ...activeBadgeStyle,
  color: "var(--ct-colors-ink-muted)",
  background: "var(--ct-colors-surface-3)",
};

const statusTextStyle: CSSProperties = {
  margin: "0.2rem 0 0",
  minHeight: "1.4rem",
  color: "var(--ct-colors-ink-muted)",
};

const errorTextStyle: CSSProperties = {
  margin: 0,
  minHeight: "1.25rem",
  color: "var(--ct-colors-semantic-conflict)",
  fontWeight: 700,
};

const publicErrorStyle: CSSProperties = {
  marginTop: "0.25rem",
};

const errorMessageStyle: CSSProperties = {
  margin: 0,
  minHeight: "1.25rem",
  color: "var(--ct-colors-semantic-conflict)",
};

const errorDetailStyle: CSSProperties = {
  margin: "0.2rem 0 0",
  minHeight: "1.25rem",
  color: "var(--ct-colors-semantic-conflict)",
  overflowWrap: "anywhere",
};

const definitionPanelStyle: CSSProperties = {
  display: "grid",
  gridTemplateColumns: "minmax(12rem, 24rem) minmax(12rem, 24rem) auto",
  gap: "var(--ct-spacing-md)",
  alignItems: "end",
  marginTop: "var(--ct-spacing-md)",
};

const definitionLabelStyle: CSSProperties = {
  display: "block",
  color: "var(--ct-colors-ink-subtle)",
  fontSize: "0.72rem",
  letterSpacing: "0.12em",
  textTransform: "uppercase",
  marginBottom: "0.35rem",
};

const definitionValueStyle: CSSProperties = {
  minHeight: "2.45rem",
  display: "flex",
  alignItems: "center",
  padding: "0.3rem 0",
  color: "var(--ct-colors-ink)",
  overflowWrap: "anywhere",
};

const auditFilterGridStyle: CSSProperties = {
  display: "grid",
  gridTemplateColumns: "repeat(auto-fit, minmax(13rem, 1fr))",
  gap: "var(--ct-spacing-sm)",
};

const auditDetailCellStyle: CSSProperties = {
  padding: 0,
  borderBottom: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
};

const auditDetailPanelStyle: CSSProperties = {
  display: "grid",
  gap: "var(--ct-spacing-md)",
  padding: "var(--ct-spacing-md)",
};

const auditDetailMetaGridStyle: CSSProperties = {
  display: "grid",
  gridTemplateColumns: "repeat(auto-fit, minmax(14rem, 1fr))",
  gap: "var(--ct-spacing-sm)",
};

const auditChangeTableStyle: CSSProperties = {
  ...dataTableStyle,
  minWidth: "44rem",
};

const redactedBadgeStyle: CSSProperties = {
  display: "inline-flex",
  padding: "0.08rem 0.38rem",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-pill)",
  color: "var(--ct-colors-semantic-caution)",
  fontWeight: 700,
};

const nullValueStyle: CSSProperties = {
  color: "var(--ct-colors-ink-subtle)",
  fontStyle: "italic",
};

const jobPanelStyle: CSSProperties = {
  padding: "var(--ct-spacing-md)",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  background: "var(--ct-colors-surface-1)",
};

const visuallyHiddenStyle: CSSProperties = {
  position: "absolute",
  width: "1px",
  height: "1px",
  padding: 0,
  margin: "-1px",
  overflow: "hidden",
  clip: "rect(0, 0, 0, 0)",
  whiteSpace: "nowrap",
  border: 0,
};

const auditActionLabels: Record<string, string> = {
  account_preferences_updated: "Account preferences updated",
  auth_binding_created: "Authentication binding created",
  auth_binding_retired: "Authentication binding retired",
  auth_binding_rotated: "Authentication binding rotated",
  backup_created: "Backup created",
  bootstrap_admin_created: "Bootstrap admin created",
  deployment_admin_granted: "Deployment admin granted",
  deployment_admin_revoked: "Deployment admin revoked",
  legacy_administrative_event: "Legacy administrative event",
  password_changed: "Password changed",
  password_reset: "Password reset",
  restore_completed: "Restore completed",
  restore_failed: "Restore failed",
  restore_started: "Restore started",
  restore_verification_completed: "Restore verification completed",
  sessions_revoked: "Sessions revoked",
  totp_enrollment_begun: "TOTP enrollment begun",
  totp_enrollment_completed: "TOTP enrollment completed",
  totp_reset: "TOTP reset",
  user_created: "User created",
  user_profile_updated: "User profile updated",
  user_status_changed: "User status changed",
};
