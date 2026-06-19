import {
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
  phase1RouteTestId,
} from "@cartulary/ui-contracts";
import {
  type CSSProperties,
  type KeyboardEvent,
  lazy,
  type ReactNode,
  Suspense,
  startTransition,
  useCallback,
  useEffect,
  useId,
  useMemo,
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
  Phase1AccountPanel,
  Phase1AdminPanel,
  Phase1AuthSurface,
} from "./Phase1Surface";
import {
  type AccountPreferencesResource,
  type AccountProfileResource,
  type CredentialState,
  type DensityMode,
  type ExtensionProfileResource,
  loadAccountPreferences,
  loadAccountProfile,
  loadCredentialState,
  loadExtensions,
  loadSession,
  patchAccountProfile,
  putAccountPreferences,
  type SessionData,
} from "./phase1Client";
import { ReferencePackAdminPanel } from "./ReferencePackAdminPanel";

const LazyWorkbookShell = lazy(async () => {
  const module = await import("../workbook/WorkbookShell");
  return { default: module.WorkbookShell };
});

const LazyDebugHarnessShell = lazy(async () => {
  const module = await import("./DebugHarnessShell");
  return { default: module.DebugHarnessShell };
});

type IncidentData = {
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

type IncidentStatusFilter = "active" | "all" | "closed";

type PagingMeta = {
  limit: number;
  has_more: boolean;
  next_cursor: string | null;
};

type IncidentListEnvelope = {
  data: {
    incidents: IncidentData[];
  };
  meta: {
    paging: PagingMeta;
  };
};

type RouteState = {
  incidentId: string;
  debugHarness: boolean;
};

type ShellRefreshOptions = {
  anonymousMessage?: string;
  landingNotice?: string | null;
  routeSnapshot: RouteState;
};

type AppBootstrapState =
  | "loading"
  | "anonymous"
  | "authenticated"
  | "forbidden"
  | "revoked"
  | "public_error_envelope";

type LandingRefreshState = "idle" | "loading" | "failed";

type IncidentLandingProps = {
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
  onOpenIncident: (incidentId: string) => void;
  onRefresh: () => Promise<void> | void;
  onSearchChange: (value: string) => void;
  onStatusFilterChange: (value: IncidentStatusFilter) => void;
  statusText: string;
};

type AppProps = {
  readonly readingProfile?: CartularyReadingProfile | undefined;
  readonly themeId?: string | undefined;
};

type LandingAdminShellProps = {
  activePanel: LandingAdminPanelToken;
  availablePanels: ReadonlyArray<LandingAdminPanelDescriptor>;
  children: ReactNode;
  currentUserLabel: string;
  incidentCount: number;
  onActivePanelChange: (panel: LandingAdminPanelToken) => void;
  statusText: string;
};

type LandingAdminPanelDescriptor = {
  description: string;
  group: "account" | "deployment" | "primary";
  label: string;
  token: LandingAdminPanelToken;
};

export type CartularyReadingProfile = "default" | "hyperlegible";

const defaultAuthPrompt =
  "Sign in with your local account to open incidents, manage account security, and administer deployment users.";
const defaultStaleIncidentMessage =
  "The requested incident is no longer visible.";
const accessLostLandingNotice =
  "The current incident is no longer visible. Returned to the landing screen.";
const defaultRevokedSessionMessage =
  "The current session ended. Sign in again to continue.";

function readRouteState(): RouteState {
  const params = new URLSearchParams(window.location.search);
  return {
    incidentId: (params.get("incident_id") ?? "").trim(),
    debugHarness: params.get("debug") === "harness",
  };
}

function writeRouteState(next: RouteState, mode: "push" | "replace") {
  const params = new URLSearchParams(window.location.search);
  if (next.incidentId === "") {
    params.delete("incident_id");
    params.delete("surface");
  } else {
    params.set("incident_id", next.incidentId);
  }
  if (next.debugHarness) {
    params.set("debug", "harness");
  } else {
    params.delete("debug");
  }

  const query = params.toString();
  const url = query === "" ? "/" : `/?${query}`;
  if (mode === "push") {
    window.history.pushState({}, "", url);
    return;
  }
  window.history.replaceState({}, "", url);
}

function upsertIncident(incidents: IncidentData[], nextIncident: IncidentData) {
  const existingIndex = incidents.findIndex(
    (incident) => incident.incident_id === nextIncident.incident_id,
  );
  if (existingIndex === -1) {
    return [nextIncident, ...incidents];
  }

  const next = [...incidents];
  next[existingIndex] = nextIncident;
  return next;
}

function isAbortError(error: unknown): boolean {
  if (typeof DOMException !== "undefined" && error instanceof DOMException) {
    return error.name === "AbortError";
  }
  return (
    typeof error === "object" &&
    error !== null &&
    "name" in error &&
    (error as { name?: unknown }).name === "AbortError"
  );
}

function isSessionRequiredError(status: number, error: APIError | null) {
  return (
    status === 401 && (error === null || error.code === "session_required")
  );
}

function isForbiddenError(status: number, error: APIError | null) {
  return status === 403 || error?.code === "authorization_denied";
}

function incidentListURL(options: {
  cursorToken?: string | null;
  limit: number;
  search?: string;
  statusFilter?: IncidentStatusFilter;
}) {
  const params = new URLSearchParams();
  params.set("limit", String(options.limit));
  const cursorToken = options.cursorToken?.trim() ?? "";
  if (cursorToken !== "") {
    params.set("cursor_token", cursorToken);
  }
  const search = options.search?.trim() ?? "";
  if (search !== "") {
    params.set("search", search);
  }
  if (options.statusFilter === "active" || options.statusFilter === "closed") {
    params.set("status", options.statusFilter);
  }
  return `/api/v1/incidents?${params.toString()}`;
}

function extensionClaimed(
  profiles: ExtensionProfileResource[],
  profileId: string,
) {
  return profiles.some(
    (profile) => profile.profile_id === profileId && profile.claimed,
  );
}

function incidentCreateOptionalBody(fields: {
  currentPhase: string;
  description: string;
  externalCase: string;
  severity: string;
  tlp: string;
}) {
  const body: Record<string, string> = {};
  const optionalFields = [
    ["description", fields.description],
    ["severity", fields.severity],
    ["tlp", fields.tlp],
    ["current_phase", fields.currentPhase],
    ["primary_external_case_ref", fields.externalCase],
  ] as const;
  for (const [field, raw] of optionalFields) {
    const value = raw.trim();
    if (value !== "") {
      body[field] = value;
    }
  }
  return body;
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
  onOpenIncident,
  onRefresh,
  onSearchChange,
  onStatusFilterChange,
  statusText,
}: IncidentLandingProps) {
  const incidentKeyFieldId = useId();
  const incidentTitleFieldId = useId();
  const hasIncidents = incidents.length > 0;

  return (
    <section
      aria-busy={isRefreshing}
      data-bootstrap-state={bootstrapState}
      data-testid={phase1LandingTestId("shell")}
      style={landingPanelStyle}
    >
      <header style={landingPanelHeaderStyle}>
        <div>
          <p style={landingSectionEyebrowStyle}>Incidents</p>
          <h2 style={landingSectionTitleStyle}>Workbook access</h2>
        </div>
        <div style={landingPanelHeaderActionsStyle}>
          <p
            data-testid={phase1LandingTestId("incidents-count")}
            style={landingCountStyle}
          >
            {incidents.length} loaded{hasMoreIncidents ? " +" : ""}
          </p>
          <button
            data-testid={phase1LandingTestId("refresh")}
            style={landingSecondaryButtonStyle}
            type="button"
            onClick={() => {
              void onRefresh();
            }}
          >
            Refresh
          </button>
        </div>
      </header>

      <div style={landingWorkspaceStyle}>
        <section style={landingIncidentListPanelStyle}>
          <div style={landingSectionHeaderStyle}>
            <div>
              <p style={landingSectionEyebrowStyle}>Visible incidents</p>
              <h3 style={landingSubsectionTitleStyle}>Open workbook</h3>
            </div>
          </div>

          <label htmlFor="incident-filter" style={landingLabelStyle}>
            Search visible incidents
          </label>
          <input
            id="incident-filter"
            data-testid={phase1LandingTestId("search")}
            style={landingInputStyle}
            value={incidentSearch}
            onChange={(event) => {
              onSearchChange(event.target.value);
            }}
            placeholder="Key, title, severity, TLP, phase, external case"
          />
          <label htmlFor="incident-status-filter" style={landingLabelStyle}>
            Status
          </label>
          <select
            id="incident-status-filter"
            data-testid={phase1LandingTestId("status-filter")}
            style={landingInputStyle}
            value={incidentStatusFilter}
            onChange={(event) => {
              onStatusFilterChange(event.target.value as IncidentStatusFilter);
            }}
          >
            <option value="all">All</option>
            <option value="active">Active</option>
            <option value="closed">Closed</option>
          </select>

          {isRefreshing ? (
            <p
              aria-live="polite"
              data-testid={phase1LandingTestId("loading")}
              role="status"
              style={landingBodyStyle}
            >
              Loading visible incidents…
            </p>
          ) : null}

          {!isRefreshing && !hasIncidents ? (
            <p
              data-testid={phase1LandingTestId("empty-state")}
              style={landingBodyStyle}
            >
              No incidents are visible for this session yet.
            </p>
          ) : null}

          {!isRefreshing && !hasIncidents && incidentSearch.trim() !== "" ? (
            <p style={landingBodyStyle}>
              No visible incidents match this query.
            </p>
          ) : null}

          {hasIncidents ? (
            <div
              data-testid={phase1LandingTestId("incident-list")}
              style={landingListStyle}
            >
              {incidents.map((incident) => (
                <article
                  key={incident.incident_id}
                  data-testid={landingIncidentCardTestId(incident.incident_id)}
                  style={landingIncidentCardStyle}
                >
                  <div style={landingIncidentSummaryStyle}>
                    <p style={landingIncidentKeyStyle}>
                      {incident.incident_key}
                    </p>
                    <h3 style={landingIncidentTitleStyle}>{incident.title}</h3>
                    <div style={landingIncidentMetaGridStyle}>
                      <span>v{incident.incident_version}</span>
                      <span>{incident.status ?? "active"}</span>
                      <span>{incident.current_phase ?? "No phase"}</span>
                      <span>{incident.severity ?? "No severity"}</span>
                      <span>{incident.tlp ?? "No TLP"}</span>
                      <span>
                        {incident.primary_external_case_ref ??
                          "No external case"}
                      </span>
                    </div>
                  </div>
                  <button
                    data-testid={landingIncidentOpenButtonTestId(
                      incident.incident_id,
                    )}
                    style={landingPrimaryButtonStyle}
                    type="button"
                    onClick={() => {
                      onOpenIncident(incident.incident_id);
                    }}
                  >
                    Open workbook
                  </button>
                </article>
              ))}
            </div>
          ) : null}
        </section>

        <section style={landingCreatePanelStyle}>
          <div style={landingSectionHeaderStyle}>
            <div>
              <p style={landingSectionEyebrowStyle}>Create</p>
              <h3 style={landingSubsectionTitleStyle}>New incident</h3>
            </div>
          </div>
          <div style={landingFormGridStyle}>
            <label htmlFor={incidentKeyFieldId} style={landingLabelStyle}>
              Incident key
            </label>
            <input
              data-testid={phase1LandingTestId("incident-key")}
              id={incidentKeyFieldId}
              style={landingInputStyle}
              value={createIncidentKey}
              onChange={(event) => {
                onCreateIncidentKeyChange(event.target.value);
              }}
              placeholder="IR-2026-001"
            />
            <label htmlFor={incidentTitleFieldId} style={landingLabelStyle}>
              Title
            </label>
            <input
              data-testid={phase1LandingTestId("incident-title")}
              id={incidentTitleFieldId}
              style={landingInputStyle}
              value={createIncidentTitle}
              onChange={(event) => {
                onCreateIncidentTitleChange(event.target.value);
              }}
              placeholder="Credential theft investigation"
            />
          </div>
          <details style={landingDetailsStyle}>
            <summary style={landingDetailsSummaryStyle}>More details</summary>
            <div style={landingFormGridStyle}>
              <label
                htmlFor="incident-create-description"
                style={landingLabelStyle}
              >
                Description
              </label>
              <textarea
                data-testid={phase1LandingTestId("create-description")}
                id="incident-create-description"
                style={landingTextAreaStyle}
                value={createIncidentDescription}
                onChange={(event) => {
                  onCreateIncidentDescriptionChange(event.target.value);
                }}
              />
              <label
                htmlFor="incident-create-severity"
                style={landingLabelStyle}
              >
                Severity
              </label>
              <input
                data-testid={phase1LandingTestId("create-severity")}
                id="incident-create-severity"
                style={landingInputStyle}
                value={createIncidentSeverity}
                onChange={(event) => {
                  onCreateIncidentSeverityChange(event.target.value);
                }}
              />
              <label htmlFor="incident-create-tlp" style={landingLabelStyle}>
                TLP
              </label>
              <select
                data-testid={phase1LandingTestId("create-tlp")}
                id="incident-create-tlp"
                style={landingInputStyle}
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
              <label
                htmlFor="incident-create-current-phase"
                style={landingLabelStyle}
              >
                Current phase
              </label>
              <input
                data-testid={phase1LandingTestId("create-current-phase")}
                id="incident-create-current-phase"
                style={landingInputStyle}
                value={createIncidentCurrentPhase}
                onChange={(event) => {
                  onCreateIncidentCurrentPhaseChange(event.target.value);
                }}
              />
              <label
                htmlFor="incident-create-external-case"
                style={landingLabelStyle}
              >
                External case
              </label>
              <input
                data-testid={phase1LandingTestId("create-external-case")}
                id="incident-create-external-case"
                style={landingInputStyle}
                value={createIncidentExternalCase}
                onChange={(event) => {
                  onCreateIncidentExternalCaseChange(event.target.value);
                }}
              />
            </div>
          </details>
          <button
            data-testid={phase1LandingTestId("create-button")}
            style={landingPrimaryButtonStyle}
            type="button"
            onClick={() => {
              void onCreate();
            }}
          >
            Create and open
          </button>
        </section>
      </div>

      <p
        aria-live="polite"
        data-testid={phase1LandingTestId("status")}
        role="status"
        style={landingStatusStyle}
      >
        {statusText}
      </p>
      <p
        aria-live="assertive"
        data-testid={phase1ErrorCodeTestId("landing")}
        role={error === null ? undefined : "alert"}
        style={landingErrorStyle}
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

export function App({ readingProfile = "default", themeId }: AppProps = {}) {
  const [route, setRoute] = useState<RouteState>(() => readRouteState());
  const [session, setSession] = useState<SessionData | null>(null);
  const [credentialState, setCredentialState] =
    useState<CredentialState | null>(null);
  const [credentialError, setCredentialError] = useState<APIError | null>(null);
  const [incidents, setIncidents] = useState<IncidentData[]>([]);
  const [incidentsPaging, setIncidentsPaging] = useState<PagingMeta>({
    limit: 100,
    has_more: false,
    next_cursor: null,
  });
  const [extensionProfiles, setExtensionProfiles] = useState<
    ExtensionProfileResource[]
  >([]);
  const [appBootstrapState, setAppBootstrapState] =
    useState<AppBootstrapState>("loading");
  const [landingRefreshState, setLandingRefreshState] =
    useState<LandingRefreshState>("idle");
  const [landingNotice, setLandingNotice] = useState<string | null>(null);
  const [error, setError] = useState<APIError | null>(null);
  const [authPrompt, setAuthPrompt] = useState(defaultAuthPrompt);
  const [activeAdminPanel, setActiveAdminPanel] =
    useState<LandingAdminPanelToken>("incidents");
  const [incidentSearch, setIncidentSearch] = useState("");
  const [incidentStatusFilter, setIncidentStatusFilter] =
    useState<IncidentStatusFilter>("all");
  const [createIncidentKey, setCreateIncidentKey] = useState("");
  const [createIncidentTitle, setCreateIncidentTitle] = useState("");
  const [createIncidentDescription, setCreateIncidentDescription] =
    useState("");
  const [createIncidentSeverity, setCreateIncidentSeverity] = useState("");
  const [createIncidentTLP, setCreateIncidentTLP] = useState("");
  const [createIncidentCurrentPhase, setCreateIncidentCurrentPhase] =
    useState("");
  const [createIncidentExternalCase, setCreateIncidentExternalCase] =
    useState("");
  const routeRef = useRef(route);
  const sessionRef = useRef(session);
  const incidentSearchRef = useRef(incidentSearch);
  const incidentStatusFilterRef = useRef(incidentStatusFilter);
  const lastLoadedIncidentDirectoryScopeRef = useRef<{
    search: string;
    statusFilter: IncidentStatusFilter;
  }>({
    search: "",
    statusFilter: "all",
  });
  const incidentDirectoryControlsTouchedRef = useRef(false);
  const activeRefreshRef = useRef<{
    controller: AbortController | null;
    requestID: number;
  }>({
    controller: null,
    requestID: 0,
  });

  routeRef.current = route;
  sessionRef.current = session;
  incidentSearchRef.current = incidentSearch;
  incidentStatusFilterRef.current = incidentStatusFilter;

  const refreshShell = useCallback(async (options: ShellRefreshOptions) => {
    const requestID = activeRefreshRef.current.requestID + 1;
    const previousController = activeRefreshRef.current.controller;
    const controller = new AbortController();
    activeRefreshRef.current = {
      controller,
      requestID,
    };
    previousController?.abort();

    const canCommit = () =>
      activeRefreshRef.current.requestID === requestID &&
      activeRefreshRef.current.controller === controller &&
      !controller.signal.aborted;

    const hadSessionAtRefreshStart = sessionRef.current !== null;

    if (!hadSessionAtRefreshStart) {
      setAppBootstrapState("loading");
      setLandingRefreshState("idle");
    } else {
      setAppBootstrapState("authenticated");
      setLandingRefreshState("loading");
    }
    setError(null);
    setLandingNotice(options.landingNotice ?? null);

    try {
      const sessionResult = await loadSession({
        signal: controller.signal,
      });
      if (!canCommit()) {
        return;
      }

      if (!sessionResult.ok) {
        const sessionError = extractError(sessionResult.payload);
        setSession(null);
        setCredentialState(null);
        setCredentialError(null);
        setIncidents([]);
        setIncidentsPaging({ limit: 100, has_more: false, next_cursor: null });
        setExtensionProfiles([]);
        setLandingNotice(null);
        setError(
          hadSessionAtRefreshStart ||
            !isSessionRequiredError(sessionResult.status, sessionError)
            ? sessionError
            : null,
        );
        setAuthPrompt(
          options.anonymousMessage ??
            (hadSessionAtRefreshStart
              ? defaultRevokedSessionMessage
              : defaultAuthPrompt),
        );
        setAppBootstrapState(
          isSessionRequiredError(sessionResult.status, sessionError)
            ? hadSessionAtRefreshStart
              ? "revoked"
              : "anonymous"
            : "public_error_envelope",
        );
        setLandingRefreshState("idle");
        return;
      }

      const directoryScope = {
        search: incidentSearchRef.current.trim(),
        statusFilter: incidentStatusFilterRef.current,
      };
      const [
        credentialResult,
        extensionsResult,
        bootstrapIncidentsResult,
        incidentsResult,
      ] = await Promise.all([
        loadCredentialState({
          signal: controller.signal,
        }),
        loadExtensions({ signal: controller.signal }),
        fetchJSON<IncidentListEnvelope>(
          incidentListURL({ limit: 2, statusFilter: "all" }),
          { signal: controller.signal },
        ),
        fetchJSON<IncidentListEnvelope>(
          incidentListURL({
            limit: 100,
            search: directoryScope.search,
            statusFilter: directoryScope.statusFilter,
          }),
          { signal: controller.signal },
        ),
      ]);
      if (!canCommit()) {
        return;
      }

      const incidentsError = extractError(incidentsResult.payload);
      const bootstrapIncidentsError = extractError(
        bootstrapIncidentsResult.payload,
      );
      const extensionsError = extractError(extensionsResult.payload);
      const nextCredentialError = extractError(credentialResult.payload);
      const nextSession = (sessionResult.payload as { data: SessionData }).data;

      if (
        (!credentialResult.ok &&
          isSessionRequiredError(
            credentialResult.status,
            nextCredentialError,
          )) ||
        (!extensionsResult.ok &&
          isSessionRequiredError(extensionsResult.status, extensionsError)) ||
        (!bootstrapIncidentsResult.ok &&
          isSessionRequiredError(
            bootstrapIncidentsResult.status,
            bootstrapIncidentsError,
          )) ||
        (!incidentsResult.ok &&
          isSessionRequiredError(incidentsResult.status, incidentsError))
      ) {
        setSession(null);
        setCredentialState(null);
        setCredentialError(null);
        setIncidents([]);
        setIncidentsPaging({ limit: 100, has_more: false, next_cursor: null });
        setExtensionProfiles([]);
        setLandingNotice(null);
        setError(
          nextCredentialError ??
            extensionsError ??
            bootstrapIncidentsError ??
            incidentsError,
        );
        setAuthPrompt(options.anonymousMessage ?? defaultRevokedSessionMessage);
        setAppBootstrapState("revoked");
        setLandingRefreshState("idle");
        return;
      }

      if (
        !extensionsResult.ok ||
        !bootstrapIncidentsResult.ok ||
        !incidentsResult.ok
      ) {
        setSession(nextSession);
        setCredentialState(
          credentialResult.ok
            ? (credentialResult.payload as { data: CredentialState }).data
            : null,
        );
        setCredentialError(nextCredentialError);
        setIncidents([]);
        setIncidentsPaging({ limit: 100, has_more: false, next_cursor: null });
        setExtensionProfiles([]);
        setLandingNotice(options.landingNotice ?? null);
        setError(extensionsError ?? bootstrapIncidentsError ?? incidentsError);
        setAuthPrompt(defaultAuthPrompt);
        setAppBootstrapState(
          isForbiddenError(incidentsResult.status, incidentsError)
            ? "forbidden"
            : "public_error_envelope",
        );
        setLandingRefreshState("failed");
        return;
      }

      if (!credentialResult.ok) {
        setSession(nextSession);
        setCredentialState(null);
        setCredentialError(nextCredentialError);
        setIncidents([]);
        setIncidentsPaging({ limit: 100, has_more: false, next_cursor: null });
        setExtensionProfiles([]);
        setLandingNotice(options.landingNotice ?? null);
        setError(nextCredentialError);
        setAuthPrompt(defaultAuthPrompt);
        setAppBootstrapState("public_error_envelope");
        setLandingRefreshState("failed");
        return;
      }

      const nextCredentialState = credentialResult.ok
        ? (credentialResult.payload as { data: CredentialState }).data
        : null;
      const nextIncidents = (incidentsResult.payload as IncidentListEnvelope)
        .data.incidents;
      const nextIncidentsPaging = (
        incidentsResult.payload as IncidentListEnvelope
      ).meta.paging;
      const nextBootstrapIncidents = (
        bootstrapIncidentsResult.payload as IncidentListEnvelope
      ).data.incidents;
      const nextBootstrapPaging = (
        bootstrapIncidentsResult.payload as IncidentListEnvelope
      ).meta.paging;
      const nextExtensionProfiles = (
        extensionsResult.payload as {
          data: { extensions: ExtensionProfileResource[] };
        }
      ).data.extensions;
      lastLoadedIncidentDirectoryScopeRef.current = directoryScope;
      const requestedIncidentStillVisible =
        options.routeSnapshot.incidentId === "" ||
        nextBootstrapIncidents.some(
          (incident) =>
            incident.incident_id === options.routeSnapshot.incidentId,
        );
      const rootIncidentID =
        options.routeSnapshot.incidentId === "" &&
        !options.routeSnapshot.debugHarness &&
        nextBootstrapIncidents.length === 1 &&
        !nextBootstrapPaging.has_more
          ? (nextBootstrapIncidents[0]?.incident_id ?? "")
          : "";
      const nextRoute =
        rootIncidentID !== ""
          ? {
              incidentId: rootIncidentID,
              debugHarness: false,
            }
          : requestedIncidentStillVisible
            ? null
            : {
                incidentId: "",
                debugHarness: options.routeSnapshot.debugHarness,
              };
      const nextLandingNotice =
        nextRoute !== null
          ? (options.landingNotice ?? defaultStaleIncidentMessage)
          : (options.landingNotice ?? null);

      setSession(nextSession);
      setCredentialState(nextCredentialState);
      setCredentialError(nextCredentialError);
      setIncidents(nextIncidents);
      setIncidentsPaging(nextIncidentsPaging);
      setExtensionProfiles(nextExtensionProfiles);
      incidentDirectoryControlsTouchedRef.current = false;
      setLandingNotice(nextLandingNotice);
      setError(null);
      setAuthPrompt(defaultAuthPrompt);
      setAppBootstrapState("authenticated");
      setLandingRefreshState("idle");

      if (nextRoute !== null) {
        writeRouteState(nextRoute, "replace");
        startTransition(() => {
          setRoute(nextRoute);
        });
      }
    } catch (error) {
      if (isAbortError(error) || !canCommit()) {
        return;
      }
      if (sessionRef.current === null) {
        setSession(null);
        setCredentialState(null);
        setCredentialError(null);
        setIncidents([]);
        setIncidentsPaging({ limit: 100, has_more: false, next_cursor: null });
        setExtensionProfiles([]);
        setLandingNotice(null);
        setError(null);
        setAuthPrompt(options.anonymousMessage ?? defaultAuthPrompt);
        setAppBootstrapState("anonymous");
        setLandingRefreshState("idle");
        return;
      }

      setError(null);
      setAuthPrompt(defaultAuthPrompt);
      setAppBootstrapState("public_error_envelope");
      setLandingRefreshState("failed");
    }
  }, []);

  const refreshCurrentShell = useCallback(
    (options?: { anonymousMessage?: string }) =>
      refreshShell({
        routeSnapshot: routeRef.current,
        landingNotice: null,
        ...(typeof options?.anonymousMessage === "string"
          ? {
              anonymousMessage: options.anonymousMessage,
            }
          : {}),
      }),
    [refreshShell],
  );

  useEffect(() => {
    const handlePopState = () => {
      setRoute(readRouteState());
    };
    window.addEventListener("popstate", handlePopState);
    return () => {
      window.removeEventListener("popstate", handlePopState);
    };
  }, []);

  useEffect(() => {
    return () => {
      activeRefreshRef.current.controller?.abort();
    };
  }, []);

  useEffect(() => {
    if (route.debugHarness) {
      activeRefreshRef.current.controller?.abort();
      return;
    }
    void refreshShell({
      routeSnapshot: routeRef.current,
      landingNotice: null,
    });
  }, [route.debugHarness, refreshShell]);

  useEffect(() => {
    if (sessionRef.current === null || routeRef.current.incidentId !== "") {
      return;
    }
    if (!incidentDirectoryControlsTouchedRef.current) {
      return;
    }
    const nextScope = {
      search: incidentSearch.trim(),
      statusFilter: incidentStatusFilter,
    };
    const lastLoadedScope = lastLoadedIncidentDirectoryScopeRef.current;
    if (
      nextScope.search === lastLoadedScope.search &&
      nextScope.statusFilter === lastLoadedScope.statusFilter
    ) {
      incidentDirectoryControlsTouchedRef.current = false;
      return;
    }
    const timeout = window.setTimeout(() => {
      void refreshShell({
        routeSnapshot: routeRef.current,
        landingNotice: null,
      });
    }, 180);
    return () => {
      window.clearTimeout(timeout);
    };
  }, [incidentSearch, incidentStatusFilter, refreshShell]);

  const handleIncidentSearchChange = useCallback((value: string) => {
    incidentDirectoryControlsTouchedRef.current = true;
    setIncidentSearch(value);
  }, []);

  const handleIncidentStatusFilterChange = useCallback(
    (value: IncidentStatusFilter) => {
      incidentDirectoryControlsTouchedRef.current = true;
      setIncidentStatusFilter(value);
    },
    [],
  );

  const currentUserLabel = useMemo(() => {
    if (session === null) {
      return "Anonymous";
    }
    return `${session.display_name}${session.is_deployment_admin ? " · deployment admin" : ""}`;
  }, [session]);
  const landingStatusText =
    landingNotice ??
    (appBootstrapState === "forbidden"
      ? "Access to visible incidents is denied."
      : appBootstrapState === "public_error_envelope" && error !== null
        ? (publicErrorView(error)?.statusText ??
          "Failed to load visible incidents.")
        : landingRefreshState === "loading"
          ? "Loading visible incidents…"
          : landingRefreshState === "failed" || error !== null
            ? "Failed to load visible incidents."
            : incidents.length === 0
              ? "No visible incidents yet."
              : `Loaded ${incidents.length} incident${incidents.length === 1 ? "" : "s"}${incidentsPaging.has_more ? "; more available." : "."}`);
  const availableLandingPanels = useMemo(() => {
    const panels: LandingAdminPanelDescriptor[] = [
      {
        token: "incidents",
        label: "Incidents",
        description: "Open and create workbook access",
        group: "primary",
      },
      {
        token: "account-profile",
        label: "Profile",
        description: "Email and display name",
        group: "account",
      },
      {
        token: "account-appearance",
        label: "Appearance",
        description: "Density preference",
        group: "account",
      },
      {
        token: "account-security",
        label: "Security",
        description: "Session and credentials",
        group: "account",
      },
    ];
    if (session?.is_deployment_admin) {
      panels.push(
        {
          token: "deployment-users",
          label: "Deployment users",
          description: "Local user administration",
          group: "deployment",
        },
        {
          token: "administrative-audit",
          label: "Administrative audit",
          description: "Deployment audit events",
          group: "deployment",
        },
      );
      if (extensionClaimed(extensionProfiles, "reference_pack")) {
        panels.push({
          token: "reference-packs",
          label: "Reference packs",
          description: "Pack operations",
          group: "deployment",
        });
      }
      if (extensionClaimed(extensionProfiles, "incident_portability")) {
        panels.push({
          token: "incident-bundle-import",
          label: "Incident bundle import",
          description: "Create incident from bundle",
          group: "deployment",
        });
      }
    }
    return panels;
  }, [extensionProfiles, session?.is_deployment_admin]);
  useEffect(() => {
    if (
      !availableLandingPanels.some((panel) => panel.token === activeAdminPanel)
    ) {
      setActiveAdminPanel("incidents");
    }
  }, [activeAdminPanel, availableLandingPanels]);
  const readingProfileAttribute =
    readingProfile === "hyperlegible" ? readingProfile : undefined;
  const rootPageStyle =
    readingProfile === "hyperlegible"
      ? {
          ...pageStyle,
          fontFamily: "var(--ct-typography-accessible-reading-fontFamily)",
        }
      : pageStyle;
  const workbookRootPageStyle =
    readingProfile === "hyperlegible"
      ? {
          ...workbookRoutePageStyle,
          fontFamily: "var(--ct-typography-accessible-reading-fontFamily)",
        }
      : workbookRoutePageStyle;

  const openIncident = useCallback(
    (incidentId: string) => {
      const nextRoute = { incidentId, debugHarness: route.debugHarness };
      setLandingNotice(null);
      writeRouteState(nextRoute, "push");
      startTransition(() => {
        setRoute(nextRoute);
      });
    },
    [route.debugHarness],
  );

  const handleCreateIncident = useCallback(async () => {
    const incidentKey = createIncidentKey.trim();
    const title = createIncidentTitle.trim();
    if (incidentKey === "" || title === "") {
      setLandingNotice("Incident key and title are required.");
      return;
    }

    setLandingNotice(null);
    setError(null);
    const response = await fetchJSON<{ data: IncidentData }>(
      "/api/v1/incidents",
      {
        method: "POST",
        body: JSON.stringify({
          client_txn_id: clientTxnID("landing-incident"),
          incident_key: incidentKey,
          title,
          ...incidentCreateOptionalBody({
            currentPhase: createIncidentCurrentPhase,
            description: createIncidentDescription,
            externalCase: createIncidentExternalCase,
            severity: createIncidentSeverity,
            tlp: createIncidentTLP,
          }),
        }),
      },
    );
    if (!response.ok) {
      const nextError = extractError(response.payload);
      setError(nextError);
      setAppBootstrapState(
        isSessionRequiredError(response.status, nextError)
          ? "revoked"
          : isForbiddenError(response.status, nextError)
            ? "forbidden"
            : "public_error_envelope",
      );
      if (isSessionRequiredError(response.status, nextError)) {
        setSession(null);
        setCredentialState(null);
        setCredentialError(null);
        setIncidents([]);
        setAuthPrompt(defaultRevokedSessionMessage);
      }
      setLandingNotice("Incident create failed.");
      return;
    }

    const incident = (response.payload as { data: IncidentData }).data;
    setCreateIncidentKey("");
    setCreateIncidentTitle("");
    setCreateIncidentDescription("");
    setCreateIncidentSeverity("");
    setCreateIncidentTLP("");
    setCreateIncidentCurrentPhase("");
    setCreateIncidentExternalCase("");
    setIncidents((current) => upsertIncident(current, incident));
    setLandingNotice(null);
    setError(null);
    setAppBootstrapState("authenticated");
    openIncident(incident.incident_id);
  }, [
    createIncidentCurrentPhase,
    createIncidentDescription,
    createIncidentExternalCase,
    createIncidentKey,
    createIncidentSeverity,
    createIncidentTLP,
    createIncidentTitle,
    openIncident,
  ]);

  const handleIncidentAccessLost = useCallback(() => {
    void refreshShell({
      routeSnapshot: routeRef.current,
      landingNotice: accessLostLandingNotice,
    });
  }, [refreshShell]);

  const handleIncidentSnapshot = useCallback((incident: IncidentData) => {
    setIncidents((current) => upsertIncident(current, incident));
  }, []);

  const handleReturnToLanding = useCallback(() => {
    const nextRoute = {
      incidentId: "",
      debugHarness: route.debugHarness,
    };
    setLandingNotice("Returned to incident landing.");
    writeRouteState(nextRoute, "push");
    startTransition(() => {
      setRoute(nextRoute);
    });
  }, [route.debugHarness]);

  if (route.incidentId !== "" && session !== null) {
    return (
      <main
        aria-busy={appBootstrapState === "loading"}
        className="cartulary-shell"
        data-bootstrap-state={appBootstrapState}
        data-cartulary-theme={themeId}
        data-reading-profile={readingProfileAttribute}
        data-testid={phase1RouteTestId("app-shell")}
        style={workbookRootPageStyle}
      >
        <section style={workbookFrameStyle}>
          <Suspense
            fallback={
              <p
                aria-live="polite"
                data-testid={phase1RouteTestId("workbook-loading")}
                role="status"
                style={routeLoadingStyle}
              >
                Loading workbook…
              </p>
            }
          >
            <LazyWorkbookShell
              currentUserLabel={currentUserLabel}
              incidentId={route.incidentId}
              onIncidentAccessLost={handleIncidentAccessLost}
              onIncidentSnapshot={handleIncidentSnapshot}
              onReturnToLanding={handleReturnToLanding}
            />
          </Suspense>
        </section>
      </main>
    );
  }

  if (route.debugHarness) {
    return (
      <main
        className="cartulary-shell"
        data-cartulary-theme={themeId}
        data-reading-profile={readingProfileAttribute}
        style={rootPageStyle}
      >
        <section style={utilityPanelStyle}>
          <div style={landingHeroStyle}>
            <p style={landingEyebrowStyle}>Cartulary</p>
            <h1 style={landingHeadlineStyle}>Debug harness shell</h1>
            <p style={landingBodyStyle}>
              Phase 1 and Phase 2 harness controls now live behind the explicit
              `?debug=harness` flag so the default root path can behave like the
              real incident landing.
            </p>
          </div>

          <Suspense
            fallback={
              <p
                aria-live="polite"
                data-testid={phase1RouteTestId("debug-harness-loading")}
                role="status"
                style={routeLoadingStyle}
              >
                Loading debug harness…
              </p>
            }
          >
            <LazyDebugHarnessShell />
          </Suspense>
        </section>
      </main>
    );
  }

  if (session === null) {
    return (
      <Phase1AuthSurface
        bootstrapState={
          appBootstrapState === "revoked" ||
          appBootstrapState === "public_error_envelope"
            ? appBootstrapState
            : appBootstrapState === "loading"
              ? "loading"
              : "anonymous"
        }
        message={authPrompt}
        onAuthenticated={async () => {
          await refreshShell({
            routeSnapshot: routeRef.current,
            landingNotice: null,
          });
        }}
        publicError={error}
        readingProfile={readingProfile}
      />
    );
  }

  return (
    <main
      aria-busy={appBootstrapState === "loading"}
      className="cartulary-shell"
      data-bootstrap-state={appBootstrapState}
      data-cartulary-theme={themeId}
      data-reading-profile={readingProfileAttribute}
      data-testid={phase1RouteTestId("app-shell")}
      style={rootPageStyle}
    >
      <LandingAdminShell
        activePanel={activeAdminPanel}
        availablePanels={availableLandingPanels}
        currentUserLabel={currentUserLabel}
        incidentCount={incidents.length}
        onActivePanelChange={setActiveAdminPanel}
        statusText={landingStatusText}
      >
        <section
          id={landingAdminPanelTestId("incidents")}
          aria-labelledby={landingAdminMenuItemTestId("incidents")}
          data-testid={landingAdminPanelTestId("incidents")}
          hidden={activeAdminPanel !== "incidents"}
          style={landingAdminPanelRegionStyle}
        >
          <IncidentLanding
            bootstrapState={appBootstrapState}
            createIncidentCurrentPhase={createIncidentCurrentPhase}
            createIncidentDescription={createIncidentDescription}
            createIncidentExternalCase={createIncidentExternalCase}
            createIncidentKey={createIncidentKey}
            createIncidentSeverity={createIncidentSeverity}
            createIncidentTitle={createIncidentTitle}
            createIncidentTLP={createIncidentTLP}
            error={error}
            hasMoreIncidents={incidentsPaging.has_more}
            incidents={incidents}
            incidentSearch={incidentSearch}
            incidentStatusFilter={incidentStatusFilter}
            isRefreshing={landingRefreshState === "loading"}
            onCreate={handleCreateIncident}
            onCreateIncidentCurrentPhaseChange={setCreateIncidentCurrentPhase}
            onCreateIncidentDescriptionChange={setCreateIncidentDescription}
            onCreateIncidentExternalCaseChange={setCreateIncidentExternalCase}
            onCreateIncidentKeyChange={setCreateIncidentKey}
            onCreateIncidentSeverityChange={setCreateIncidentSeverity}
            onCreateIncidentTitleChange={setCreateIncidentTitle}
            onCreateIncidentTLPChange={setCreateIncidentTLP}
            onOpenIncident={openIncident}
            onRefresh={() => {
              void refreshShell({
                routeSnapshot: routeRef.current,
                landingNotice: null,
              });
            }}
            onSearchChange={handleIncidentSearchChange}
            onStatusFilterChange={handleIncidentStatusFilterChange}
            statusText={landingStatusText}
          />
        </section>
        <section
          id={landingAdminPanelTestId("account-profile")}
          aria-labelledby={landingAdminMenuItemTestId("account-profile")}
          data-testid={landingAdminPanelTestId("account-profile")}
          hidden={activeAdminPanel !== "account-profile"}
          style={landingAdminPanelRegionStyle}
        >
          <AccountProfilePanel onRefreshShell={refreshCurrentShell} />
        </section>
        <section
          id={landingAdminPanelTestId("account-appearance")}
          aria-labelledby={landingAdminMenuItemTestId("account-appearance")}
          data-testid={landingAdminPanelTestId("account-appearance")}
          hidden={activeAdminPanel !== "account-appearance"}
          style={landingAdminPanelRegionStyle}
        >
          <AccountAppearancePanel />
        </section>
        <section
          id={landingAdminPanelTestId("account-security")}
          aria-labelledby={landingAdminMenuItemTestId("account-security")}
          data-testid={landingAdminPanelTestId("account-security")}
          hidden={activeAdminPanel !== "account-security"}
          style={landingAdminPanelRegionStyle}
        >
          <Phase1AccountPanel
            credentialState={credentialState}
            credentialStateError={credentialError}
            onRefreshShell={refreshCurrentShell}
            session={session}
          />
        </section>
        {session.is_deployment_admin ? (
          <section
            id={landingAdminPanelTestId("deployment-users")}
            aria-labelledby={landingAdminMenuItemTestId("deployment-users")}
            data-testid={landingAdminPanelTestId("deployment-users")}
            hidden={activeAdminPanel !== "deployment-users"}
            style={landingAdminPanelRegionStyle}
          >
            <Phase1AdminPanel
              autoLoadUsers={activeAdminPanel === "deployment-users"}
              onRefreshShell={refreshCurrentShell}
              session={session}
            />
          </section>
        ) : null}
        {session.is_deployment_admin ? (
          <section
            id={landingAdminPanelTestId("administrative-audit")}
            aria-labelledby={landingAdminMenuItemTestId("administrative-audit")}
            data-testid={landingAdminPanelTestId("administrative-audit")}
            hidden={activeAdminPanel !== "administrative-audit"}
            style={landingAdminPanelRegionStyle}
          >
            <AdministrativeAuditPanel />
          </section>
        ) : null}
        {session.is_deployment_admin &&
        extensionClaimed(extensionProfiles, "reference_pack") ? (
          <section
            id={landingAdminPanelTestId("reference-packs")}
            aria-labelledby={landingAdminMenuItemTestId("reference-packs")}
            data-testid={landingAdminPanelTestId("reference-packs")}
            hidden={activeAdminPanel !== "reference-packs"}
            style={landingAdminPanelRegionStyle}
          >
            <ReferencePackAdminPanel session={session} />
          </section>
        ) : null}
        {session.is_deployment_admin &&
        extensionClaimed(extensionProfiles, "incident_portability") ? (
          <section
            id={landingAdminPanelTestId("incident-bundle-import")}
            aria-labelledby={landingAdminMenuItemTestId(
              "incident-bundle-import",
            )}
            data-testid={landingAdminPanelTestId("incident-bundle-import")}
            hidden={activeAdminPanel !== "incident-bundle-import"}
            style={landingAdminPanelRegionStyle}
          >
            <IncidentBundleImportPanel />
          </section>
        ) : null}
      </LandingAdminShell>
    </main>
  );
}

function AccountProfilePanel({
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
    <section style={landingIncidentListPanelStyle}>
      <p style={landingSectionEyebrowStyle}>Profile</p>
      <h2 style={landingSectionTitleStyle}>Account profile</h2>
      <div style={landingFormGridStyle}>
        <label htmlFor="account-profile-email" style={landingLabelStyle}>
          Email
        </label>
        <input
          data-testid={phase1AccountTestId("profile-email")}
          id="account-profile-email"
          readOnly
          style={landingInputStyle}
          value={profile?.email ?? ""}
        />
        <label htmlFor="account-profile-display-name" style={landingLabelStyle}>
          Display name
        </label>
        <input
          data-testid={phase1AccountTestId("profile-display-name")}
          id="account-profile-display-name"
          style={landingInputStyle}
          value={displayName}
          onChange={(event) => {
            setDisplayName(event.target.value);
          }}
        />
        <button
          data-testid={phase1AccountTestId("profile-save")}
          disabled={profile === null}
          style={landingPrimaryButtonStyle}
          type="button"
          onClick={() => {
            void saveProfile();
          }}
        >
          Save profile
        </button>
      </div>
      <p aria-live="polite" role="status" style={landingStatusStyle}>
        {status}
      </p>
      <p
        aria-live="assertive"
        role={error === null ? undefined : "alert"}
        style={landingErrorStyle}
      >
        {publicErrorView(error)?.code ?? ""}
      </p>
    </section>
  );
}

function AccountAppearancePanel() {
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
    <section style={landingIncidentListPanelStyle}>
      <p style={landingSectionEyebrowStyle}>Appearance</p>
      <h2 style={landingSectionTitleStyle}>Density</h2>
      <div style={landingFormGridStyle}>
        <label htmlFor="account-density-mode" style={landingLabelStyle}>
          Density
        </label>
        <select
          data-testid={phase1AccountTestId("appearance-density-mode")}
          id="account-density-mode"
          style={landingInputStyle}
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
        <button
          data-testid={phase1AccountTestId("appearance-save")}
          disabled={preferences === null}
          style={landingPrimaryButtonStyle}
          type="button"
          onClick={() => {
            void savePreferences();
          }}
        >
          Save appearance
        </button>
      </div>
      <p aria-live="polite" role="status" style={landingStatusStyle}>
        {status}
      </p>
      <p
        aria-live="assertive"
        role={error === null ? undefined : "alert"}
        style={landingErrorStyle}
      >
        {publicErrorView(error)?.code ?? ""}
      </p>
    </section>
  );
}

type AdministrativeAuditEvent = {
  audit_event_id: string;
  occurred_at: string;
  actor_user_id: string | null;
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

function AdministrativeAuditPanel() {
  const [events, setEvents] = useState<AdministrativeAuditEvent[]>([]);
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
      data: { administrative_audit_events: AdministrativeAuditEvent[] };
    }>(`/api/v1/administrative-audit-events?${params.toString()}`);
    const nextError = extractError(result.payload);
    setError(nextError);
    if (!result.ok) {
      setStatus("Administrative audit unavailable.");
      return;
    }
    setEvents(
      (
        result.payload as {
          data: { administrative_audit_events: AdministrativeAuditEvent[] };
        }
      ).data.administrative_audit_events,
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
    <section style={landingIncidentListPanelStyle}>
      <div style={landingSectionHeaderStyle}>
        <div>
          <p style={landingSectionEyebrowStyle}>Administrative audit</p>
          <h2 style={landingSectionTitleStyle}>Deployment events</h2>
        </div>
        <button
          style={landingSecondaryButtonStyle}
          type="button"
          onClick={() => {
            void loadAudit();
          }}
        >
          Refresh
        </button>
      </div>
      <div style={auditFilterGridStyle}>
        <input
          style={landingInputStyle}
          value={actorUserID}
          onChange={(event) => setActorUserID(event.target.value)}
          placeholder="Actor user id"
        />
        <input
          style={landingInputStyle}
          value={actionCode}
          onChange={(event) => setActionCode(event.target.value)}
          placeholder="Action code"
        />
        <select
          style={landingInputStyle}
          value={targetKind}
          onChange={(event) => setTargetKind(event.target.value)}
        >
          <option value="">Target kind</option>
          <option value="deployment">Deployment</option>
          <option value="incident">Incident</option>
          <option value="user">User</option>
        </select>
        <input
          style={landingInputStyle}
          value={targetID}
          onChange={(event) => setTargetID(event.target.value)}
          placeholder="Target id"
        />
        <input
          style={landingInputStyle}
          value={occurredAtGTE}
          onChange={(event) => setOccurredAtGTE(event.target.value)}
          placeholder="Occurred at or after"
        />
        <input
          style={landingInputStyle}
          value={occurredAtLT}
          onChange={(event) => setOccurredAtLT(event.target.value)}
          placeholder="Occurred before"
        />
      </div>
      <div style={landingListStyle}>
        {events.map((event) => (
          <article key={event.audit_event_id} style={landingIncidentCardStyle}>
            <div style={landingIncidentSummaryStyle}>
              <p style={landingIncidentKeyStyle}>{event.action_code}</p>
              <h3 style={landingIncidentTitleStyle}>
                {event.target_kind}: {event.target_id ?? "deployment"}
              </h3>
              <div style={landingIncidentMetaGridStyle}>
                <span>{event.occurred_at}</span>
                <span>{event.actor_user_id ?? "system"}</span>
                <span>{event.reason_code ?? "No reason"}</span>
              </div>
              <div style={auditChangesStyle}>
                {event.changes.map((change) => (
                  <span key={change.field_path}>
                    {change.field_path}:{" "}
                    {change.value_state === "redacted"
                      ? "redacted"
                      : `${String(change.before ?? "")} -> ${String(change.after ?? "")}`}
                  </span>
                ))}
              </div>
            </div>
          </article>
        ))}
        {events.length === 0 ? (
          <p style={landingBodyStyle}>No administrative audit events loaded.</p>
        ) : null}
      </div>
      <p aria-live="polite" role="status" style={landingStatusStyle}>
        {status}
      </p>
      <p
        aria-live="assertive"
        role={error === null ? undefined : "alert"}
        style={landingErrorStyle}
      >
        {publicErrorView(error)?.code ?? ""}
      </p>
    </section>
  );
}

function IncidentBundleImportPanel() {
  const [file, setFile] = useState<File | null>(null);
  const [status, setStatus] = useState("Incident bundle import idle.");
  const [error, setError] = useState<APIError | null>(null);

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
            client_txn_id: clientTxnID("incident-bundle-import"),
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
    setStatus("Submitting incident bundle import.");
    const response = await fetch("/api/v1/incident-bundles/import", {
      method: "POST",
      credentials: "include",
      headers,
      body: form,
    });
    const payload = (await response.json()) as {
      data?: { job_id?: string };
      error?: APIError;
    };
    if (!response.ok) {
      setError(extractError(payload));
      setStatus("Incident bundle import failed to start.");
      return;
    }
    setError(null);
    setStatus(
      `Incident bundle import queued${payload.data?.job_id ? `: ${payload.data.job_id}` : "."}`,
    );
  }

  return (
    <section style={landingIncidentListPanelStyle}>
      <p style={landingSectionEyebrowStyle}>Incident portability</p>
      <h2 style={landingSectionTitleStyle}>Import incident bundle</h2>
      <div style={landingFormGridStyle}>
        <input
          aria-label="Incident bundle file"
          style={landingInputStyle}
          type="file"
          onChange={(event) => {
            setFile(event.currentTarget.files?.[0] ?? null);
          }}
        />
        <button
          style={landingPrimaryButtonStyle}
          type="button"
          onClick={() => {
            void submitImport();
          }}
        >
          Import bundle
        </button>
      </div>
      <p aria-live="polite" role="status" style={landingStatusStyle}>
        {status}
      </p>
      <p
        aria-live="assertive"
        role={error === null ? undefined : "alert"}
        style={landingErrorStyle}
      >
        {publicErrorView(error)?.code ?? ""}
      </p>
    </section>
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

function LandingAdminShell({
  activePanel,
  availablePanels,
  children,
  currentUserLabel,
  incidentCount,
  onActivePanelChange,
  statusText,
}: LandingAdminShellProps) {
  const menuItemRefs = useRef(
    new Map<LandingAdminPanelToken, HTMLButtonElement>(),
  );
  const accountPanels = availablePanels.filter(
    (panel) => panel.group === "account",
  );
  const navigationPanels = availablePanels.filter(
    (panel) => panel.group !== "account",
  );

  function focusPanelMenuItem(panel: LandingAdminPanelToken) {
    const focus = () => {
      menuItemRefs.current.get(panel)?.focus();
    };
    if (typeof window.requestAnimationFrame === "function") {
      window.requestAnimationFrame(focus);
      return;
    }
    window.setTimeout(focus, 0);
  }

  function selectPanel(panel: LandingAdminPanelToken, focus = false) {
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
      selectPanel(navigationPanels[index]?.token ?? "incidents", true);
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
        <div>
          <p style={landingEyebrowStyle}>Cartulary</p>
          <h1 style={landingAdminTitleStyle}>Incidents</h1>
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
          <div>
            <dt style={landingToolbarLabelStyle}>Loaded incidents</dt>
            <dd style={landingAdminMetaValueStyle}>{incidentCount}</dd>
          </div>
        </dl>
        <nav style={landingAccountNavStyle} aria-label="Account settings">
          {accountPanels.map((panel) => (
            <button
              key={panel.token}
              id={landingAdminMenuItemTestId(panel.token)}
              aria-controls={landingAdminPanelTestId(panel.token)}
              aria-pressed={panel.token === activePanel}
              data-testid={landingAdminMenuItemTestId(panel.token)}
              style={
                panel.token === activePanel
                  ? landingAccountNavButtonSelectedStyle
                  : landingAccountNavButtonStyle
              }
              type="button"
              onClick={() => {
                selectPanel(panel.token);
              }}
            >
              {panel.label}
            </button>
          ))}
        </nav>
      </header>

      <div style={landingAdminWorkspaceStyle}>
        <nav
          data-testid={landingAdminShellTestId("menu")}
          style={landingAdminMenuStyle}
          aria-label="Incident administration panels"
          onKeyDown={handleMenuKeyDown}
        >
          <div style={landingAdminMenuItemsStyle}>
            {navigationPanels.map((panel) => {
              const selected = panel.token === activePanel;
              return (
                <button
                  key={panel.token}
                  id={landingAdminMenuItemTestId(panel.token)}
                  ref={(element) => {
                    if (element === null) {
                      menuItemRefs.current.delete(panel.token);
                      return;
                    }
                    menuItemRefs.current.set(panel.token, element);
                  }}
                  aria-controls={landingAdminPanelTestId(panel.token)}
                  aria-pressed={selected}
                  data-testid={landingAdminMenuItemTestId(panel.token)}
                  style={
                    selected
                      ? landingAdminMenuItemSelectedStyle
                      : landingAdminMenuItemStyle
                  }
                  type="button"
                  onClick={() => {
                    selectPanel(panel.token);
                  }}
                >
                  <span style={landingAdminMenuItemLabelStyle}>
                    {panel.label}
                  </span>
                  <span style={landingAdminMenuItemDescriptionStyle}>
                    {panel.description}
                  </span>
                </button>
              );
            })}
          </div>
        </nav>
        <div style={landingAdminContentStyle}>{children}</div>
      </div>

      <footer
        aria-live="polite"
        data-testid={landingAdminShellTestId("status-strip")}
        role="status"
        style={landingAdminStatusStripStyle}
      >
        <span style={landingAdminStatusPrimaryStyle}>Ready</span>
        <span style={landingAdminStatusSecondaryStyle}>{statusText}</span>
      </footer>
    </section>
  );
}

const landingAdminShellStyle: CSSProperties = {
  width: "100%",
  minHeight: "100vh",
  display: "grid",
  gridTemplateRows: "auto minmax(0, 1fr) var(--ct-layout-statusStripHeight)",
  background: "var(--ct-colors-canvas)",
  overflow: "hidden",
};

const landingAdminHeaderStyle: CSSProperties = {
  display: "flex",
  justifyContent: "space-between",
  gap: "var(--ct-spacing-lg)",
  alignItems: "center",
  minHeight: "var(--ct-layout-topBarHeight)",
  padding: "var(--ct-spacing-md) var(--ct-spacing-lg)",
  borderBottom: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
};

const landingAccountNavStyle: CSSProperties = {
  display: "flex",
  flexWrap: "wrap",
  justifyContent: "flex-end",
  gap: "var(--ct-spacing-xs)",
};

const landingAccountNavButtonStyle: CSSProperties = {
  padding: "0.55rem 0.75rem",
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

const landingAdminTitleStyle: CSSProperties = {
  margin: "0.2rem 0 0",
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
  margin: "0.25rem 0 0",
  color: "var(--ct-colors-ink)",
  fontWeight: 700,
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
  gap: "var(--ct-spacing-sm)",
};

const landingAdminMenuItemStyle: CSSProperties = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
  width: "100%",
  minWidth: 0,
  padding: "var(--ct-spacing-sm) var(--ct-spacing-md)",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  background: "var(--ct-colors-surface-2)",
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

const landingAdminMenuItemLabelStyle: CSSProperties = {
  fontWeight: 700,
};

const landingAdminMenuItemDescriptionStyle: CSSProperties = {
  color: "var(--ct-colors-ink-subtle)",
  fontSize: "0.78rem",
};

const landingAdminContentStyle: CSSProperties = {
  minWidth: 0,
  minHeight: 0,
  overflow: "auto",
};

const landingAdminPanelRegionStyle: CSSProperties = {
  minWidth: 0,
  minHeight: 0,
  padding: "var(--ct-spacing-md)",
};

const landingAdminStatusStripStyle: CSSProperties = {
  display: "flex",
  alignItems: "center",
  gap: "var(--ct-spacing-md)",
  minHeight: "var(--ct-layout-statusStripHeight)",
  padding: "0 var(--ct-spacing-lg)",
  borderTop: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
  color: "var(--ct-colors-ink-muted)",
  fontSize: "var(--ct-typography-metadata-fontSize)",
};

const landingAdminStatusPrimaryStyle: CSSProperties = {
  color: "var(--ct-colors-semantic-success)",
  fontWeight: 700,
};

const landingAdminStatusSecondaryStyle: CSSProperties = {
  minWidth: 0,
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap",
};

const pageStyle = {
  minHeight: "100vh",
  margin: 0,
  padding: 0,
  background: "var(--ct-colors-canvas)",
  color: "var(--ct-colors-ink)",
  fontFamily: "var(--ct-typography-ui-fontFamily)",
  fontSize: "var(--ct-typography-ui-fontSize)",
  lineHeight: "var(--ct-typography-ui-lineHeight)",
};

const workbookRoutePageStyle = {
  ...pageStyle,
  blockSize: "var(--ct-app-viewport-block-size)",
  minBlockSize: 0,
  minHeight: 0,
  overflow: "hidden",
};

const landingPanelStyle = {
  minHeight: "100%",
  minWidth: 0,
  display: "grid",
  gridTemplateRows: "auto minmax(0, 1fr) auto auto auto",
  gap: "var(--ct-spacing-md)",
  padding: "var(--ct-spacing-md)",
  background: "var(--ct-colors-surface-1)",
  color: "var(--ct-colors-ink)",
};

const utilityPanelStyle = {
  width: "min(78rem, 100%)",
  margin: "2rem auto",
  padding: "2rem",
  borderRadius: "var(--ct-rounded-lg)",
  background: "var(--ct-colors-surface-1)",
  boxShadow: "var(--ct-elevation-panel)",
  border: "var(--ct-border-hairline)",
};

const workbookFrameStyle = {
  display: "grid",
  gridTemplateRows: "minmax(0, 1fr)",
  blockSize: "100%",
  minBlockSize: 0,
  minWidth: 0,
  overflow: "hidden",
};

const landingHeroStyle = {
  marginBottom: "1.5rem",
};

const landingEyebrowStyle = {
  margin: 0,
  fontSize: "0.78rem",
  letterSpacing: "0.18em",
  textTransform: "uppercase" as const,
  color: "var(--ct-colors-accent)",
};

const landingHeadlineStyle = {
  margin: "0.45rem 0 0",
  fontSize: "clamp(2rem, 4vw, 3rem)",
  lineHeight: 1.05,
};

const landingBodyStyle = {
  margin: "0.9rem 0 0",
  maxWidth: "42rem",
  lineHeight: 1.6,
  color: "var(--ct-colors-ink-muted)",
};

const landingPanelHeaderStyle = {
  display: "flex",
  justifyContent: "space-between",
  alignItems: "center",
  gap: "var(--ct-spacing-md)",
  padding: "var(--ct-spacing-md)",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-md)",
  background: "var(--ct-colors-surface-2)",
};

const landingPanelHeaderActionsStyle = {
  display: "flex",
  alignItems: "center",
  gap: "var(--ct-spacing-sm)",
};

const landingToolbarLabelStyle = {
  margin: 0,
  fontSize: "0.72rem",
  letterSpacing: "0.14em",
  textTransform: "uppercase" as const,
  color: "var(--ct-colors-ink-subtle)",
};

const landingWorkspaceStyle = {
  minWidth: 0,
  minHeight: 0,
  display: "grid",
  gridTemplateColumns: "minmax(28rem, 1fr) minmax(20rem, 24rem)",
  gap: "var(--ct-spacing-md)",
  alignItems: "start",
};

const landingIncidentListPanelStyle = {
  minWidth: 0,
  minHeight: 0,
  padding: "var(--ct-spacing-md)",
  borderRadius: "var(--ct-rounded-md)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
};

const landingCreatePanelStyle = {
  minWidth: 0,
  padding: "var(--ct-spacing-md)",
  borderRadius: "var(--ct-rounded-md)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
};

const landingSectionHeaderStyle = {
  display: "flex",
  justifyContent: "space-between",
  alignItems: "flex-start",
  gap: "1rem",
  marginBottom: "1rem",
};

const landingSectionEyebrowStyle = {
  margin: 0,
  fontSize: "0.72rem",
  letterSpacing: "0.14em",
  textTransform: "uppercase" as const,
  color: "var(--ct-colors-ink-subtle)",
};

const landingSectionTitleStyle = {
  margin: "0.25rem 0 0",
  fontSize: "1.35rem",
};

const landingSubsectionTitleStyle = {
  margin: "0.25rem 0 0",
  fontSize: "var(--ct-typography-section-heading-fontSize)",
  lineHeight: "var(--ct-typography-section-heading-lineHeight)",
};

const landingCountStyle = {
  margin: 0,
  minWidth: "2.5rem",
  padding: "0.5rem 0.75rem",
  borderRadius: "var(--ct-rounded-pill)",
  background: "var(--ct-colors-surface-3)",
  textAlign: "center" as const,
  fontWeight: 600,
};

const landingFormGridStyle = {
  display: "grid",
  gridTemplateColumns: "minmax(0, 1fr)",
  gap: "0.65rem",
};

const landingDetailsStyle = {
  marginTop: "var(--ct-spacing-md)",
};

const landingDetailsSummaryStyle = {
  cursor: "pointer",
  color: "var(--ct-colors-ink-muted)",
  fontWeight: 700,
};

const landingLabelStyle = {
  fontSize: "0.9rem",
  fontWeight: 600,
};

const landingInputStyle = {
  width: "100%",
  padding: "0.8rem 0.9rem",
  borderRadius: "var(--ct-component-text-input-rounded)",
  border: "var(--ct-component-text-input-border)",
  background: "var(--ct-component-text-input-backgroundColor)",
  color: "var(--ct-component-text-input-textColor)",
};

const landingTextAreaStyle = {
  ...landingInputStyle,
  minHeight: "6rem",
  resize: "vertical" as const,
};

const landingPrimaryButtonStyle = {
  padding: "0.8rem 1rem",
  borderRadius: "var(--ct-component-button-primary-rounded)",
  border: "none",
  background: "var(--ct-component-button-primary-backgroundColor)",
  color: "var(--ct-component-button-primary-textColor)",
  fontWeight: 600,
  cursor: "pointer",
};

const landingSecondaryButtonStyle = {
  padding: "0.8rem 1rem",
  borderRadius: "var(--ct-component-button-secondary-rounded)",
  border: "var(--ct-component-button-secondary-border)",
  background: "var(--ct-component-button-secondary-backgroundColor)",
  color: "var(--ct-component-button-secondary-textColor)",
  fontWeight: 600,
  cursor: "pointer",
};

const landingListStyle = {
  display: "grid",
  gap: "var(--ct-spacing-sm)",
  marginTop: "var(--ct-spacing-md)",
};

const landingIncidentCardStyle = {
  display: "grid",
  gridTemplateColumns: "minmax(0, 1fr) auto",
  alignItems: "center",
  gap: "var(--ct-spacing-md)",
  minWidth: 0,
  padding: "var(--ct-spacing-md)",
  borderRadius: "var(--ct-rounded-md)",
  background: "var(--ct-colors-surface-1)",
  border: "var(--ct-border-hairline)",
};

const landingIncidentSummaryStyle = {
  minWidth: 0,
};

const landingIncidentKeyStyle = {
  margin: 0,
  fontSize: "0.78rem",
  letterSpacing: "0.14em",
  textTransform: "uppercase" as const,
  color: "var(--ct-colors-ink-subtle)",
};

const landingIncidentTitleStyle = {
  margin: "0.35rem 0 0",
  fontSize: "1.05rem",
};

const landingIncidentMetaGridStyle = {
  display: "grid",
  gridTemplateColumns: "repeat(auto-fit, minmax(7rem, max-content))",
  gap: "var(--ct-spacing-sm)",
  margin: "0.4rem 0 0",
  color: "var(--ct-colors-ink-muted)",
  fontSize: "var(--ct-typography-metadata-fontSize)",
};

const auditFilterGridStyle = {
  display: "grid",
  gridTemplateColumns: "repeat(auto-fit, minmax(14rem, 1fr))",
  gap: "var(--ct-spacing-sm)",
  marginBottom: "var(--ct-spacing-md)",
};

const auditChangesStyle = {
  display: "grid",
  gap: "0.25rem",
  marginTop: "var(--ct-spacing-sm)",
  color: "var(--ct-colors-ink-muted)",
  fontSize: "var(--ct-typography-metadata-fontSize)",
};

const landingStatusStyle = {
  margin: "1.25rem 0 0",
  minHeight: "1.5rem",
  color: "var(--ct-colors-ink-muted)",
};

const routeLoadingStyle = {
  margin: "1rem 0 0",
  padding: "1rem 1.2rem",
  borderRadius: "var(--ct-rounded-lg)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-ink-muted)",
};

const landingErrorStyle = {
  margin: "0.25rem 0 0",
  minHeight: "1.25rem",
  color: "var(--ct-colors-semantic-conflict)",
  fontWeight: 600,
};

const publicErrorStyle = {
  marginTop: "0.25rem",
};

const errorMessageStyle = {
  margin: 0,
  minHeight: "1.25rem",
  color: "var(--ct-colors-semantic-conflict)",
};

const errorDetailStyle = {
  margin: "0.2rem 0 0",
  minHeight: "1.25rem",
  color: "var(--ct-colors-semantic-conflict)",
  overflowWrap: "anywhere" as const,
};
