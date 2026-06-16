import {
  dataTestIdSelector,
  type LandingAdminPanelToken,
  landingAdminCommandTestId,
  landingAdminPanelTestId,
  landingAdminShellTestId,
  landingAdminTabTestId,
  landingIncidentCardTestId,
  landingIncidentOpenButtonTestId,
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
  extractError,
  fetchJSON,
  publicErrorView,
} from "../services/browserApi";
import {
  Phase1AccountPanel,
  type Phase1AccountPanelHandle,
  Phase1AdminPanel,
  type Phase1AdminPanelCommandState,
  type Phase1AdminPanelHandle,
  Phase1AuthSurface,
} from "./Phase1Surface";
import {
  type CredentialState,
  loadCredentialState,
  loadSession,
  type SessionData,
} from "./phase1Client";
import {
  ReferencePackAdminPanel,
  type ReferencePackAdminPanelHandle,
} from "./ReferencePackAdminPanel";

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
  createIncidentKey: string;
  createIncidentTitle: string;
  currentUserLabel: string;
  error: APIError | null;
  incidents: IncidentData[];
  isRefreshing: boolean;
  onCreate: () => Promise<void> | void;
  onCreateIncidentKeyChange: (value: string) => void;
  onCreateIncidentTitleChange: (value: string) => void;
  onOpenIncident: (incidentId: string) => void;
  onRefresh: () => Promise<void> | void;
  onSelectIncident: (incidentId: string) => void;
  selectedIncidentId: string | null;
  statusText: string;
};

type AppProps = {
  readonly readingProfile?: CartularyReadingProfile | undefined;
  readonly themeId?: string | undefined;
};

type LandingAdminCommand = {
  id: string;
  label: string;
  onClick: () => Promise<void> | void;
  disabled?: boolean;
  variant?: "danger" | "primary" | "secondary";
};

type LandingAdminShellProps = {
  activePanel: LandingAdminPanelToken;
  children: ReactNode;
  commands: LandingAdminCommand[];
  currentUserLabel: string;
  incidentCount: number;
  onActivePanelChange: (panel: LandingAdminPanelToken) => void;
  statusText: string;
};

const defaultAdminCommandState: Phase1AdminPanelCommandState = {
  canLoadTargetUser: false,
  canSubmitTargetAction: false,
  canSubmitVersionedTargetAction: false,
  hasSelectedUser: false,
  targetOperationPending: false,
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

export function IncidentLanding({
  bootstrapState,
  createIncidentKey,
  createIncidentTitle,
  currentUserLabel,
  error,
  incidents,
  isRefreshing,
  onCreate,
  onCreateIncidentKeyChange,
  onCreateIncidentTitleChange,
  onOpenIncident,
  onRefresh,
  onSelectIncident,
  selectedIncidentId,
  statusText,
}: IncidentLandingProps) {
  const incidentKeyFieldId = useId();
  const incidentTitleFieldId = useId();
  const [incidentFilter, setIncidentFilter] = useState("");
  const hasIncidents = incidents.length > 0;
  const visibleIncidents = useMemo(() => {
    const query = incidentFilter.trim().toLowerCase();
    if (query === "") {
      return incidents;
    }
    return incidents.filter((incident) =>
      [
        incident.incident_key,
        incident.title,
        incident.current_phase ?? "",
        incident.tlp ?? "",
        incident.primary_external_case_ref ?? "",
      ]
        .join(" ")
        .toLowerCase()
        .includes(query),
    );
  }, [incidentFilter, incidents]);

  return (
    <section
      aria-busy={isRefreshing}
      data-bootstrap-state={bootstrapState}
      data-testid={phase1LandingTestId("shell")}
      style={landingPanelStyle}
    >
      <div style={landingHeroStyle}>
        <p style={landingEyebrowStyle}>Cartulary</p>
        <h1 style={landingHeadlineStyle}>Incident landing</h1>
        <p style={landingBodyStyle}>
          Open a visible incident or create a new one to enter the workbook. The
          debug harness remains available behind `?debug=harness`.
        </p>
      </div>

      <div style={landingToolbarStyle}>
        <div>
          <p style={landingToolbarLabelStyle}>Current session</p>
          <p style={landingToolbarValueStyle}>{currentUserLabel}</p>
        </div>
        <button
          data-testid={phase1LandingTestId("refresh")}
          style={landingSecondaryButtonStyle}
          type="button"
          onClick={() => {
            void onRefresh();
          }}
        >
          Refresh incidents
        </button>
      </div>

      <section style={landingSectionStyle}>
        <div style={landingSectionHeaderStyle}>
          <div>
            <p style={landingSectionEyebrowStyle}>Create</p>
            <h2 style={landingSectionTitleStyle}>New incident</h2>
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

      <section style={landingSectionStyle}>
        <div style={landingSectionHeaderStyle}>
          <div>
            <p style={landingSectionEyebrowStyle}>Visible incidents</p>
            <h2 style={landingSectionTitleStyle}>Workbook access</h2>
          </div>
          <p
            data-testid={phase1LandingTestId("incidents-count")}
            style={landingCountStyle}
          >
            {visibleIncidents.length}
          </p>
        </div>

        <label htmlFor="incident-filter" style={landingLabelStyle}>
          Search visible incidents
        </label>
        <input
          id="incident-filter"
          style={landingInputStyle}
          value={incidentFilter}
          onChange={(event) => {
            setIncidentFilter(event.target.value);
          }}
          placeholder="Key, title, phase, TLP, external case"
        />

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

        {!isRefreshing && hasIncidents && visibleIncidents.length === 0 ? (
          <p style={landingBodyStyle}>
            No visible incidents match this filter.
          </p>
        ) : null}

        {hasIncidents && visibleIncidents.length > 0 ? (
          <div
            data-testid={phase1LandingTestId("incident-list")}
            style={landingListStyle}
          >
            {visibleIncidents.map((incident) => {
              const selected = selectedIncidentId === incident.incident_id;
              return (
                <article
                  key={incident.incident_id}
                  data-testid={landingIncidentCardTestId(incident.incident_id)}
                  data-selected={selected ? "true" : undefined}
                  style={
                    selected
                      ? selectedLandingIncidentCardStyle
                      : landingIncidentCardStyle
                  }
                >
                  <button
                    aria-pressed={selected}
                    style={landingIncidentSelectButtonStyle}
                    type="button"
                    onClick={() => {
                      onSelectIncident(incident.incident_id);
                    }}
                  >
                    <p style={landingIncidentKeyStyle}>
                      {incident.incident_key}
                    </p>
                    <h3 style={landingIncidentTitleStyle}>{incident.title}</h3>
                    <p style={landingIncidentMetaStyle}>
                      Version {incident.incident_version}
                      {incident.current_phase
                        ? ` · ${incident.current_phase}`
                        : ""}
                      {incident.tlp ? ` · ${incident.tlp}` : ""}
                    </p>
                  </button>
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
              );
            })}
          </div>
        ) : null}
      </section>

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
  const [appBootstrapState, setAppBootstrapState] =
    useState<AppBootstrapState>("loading");
  const [landingRefreshState, setLandingRefreshState] =
    useState<LandingRefreshState>("idle");
  const [landingNotice, setLandingNotice] = useState<string | null>(null);
  const [error, setError] = useState<APIError | null>(null);
  const [authPrompt, setAuthPrompt] = useState(defaultAuthPrompt);
  const [activeAdminPanel, setActiveAdminPanel] =
    useState<LandingAdminPanelToken>("incidents");
  const [adminCommandState, setAdminCommandState] =
    useState<Phase1AdminPanelCommandState>(defaultAdminCommandState);
  const [selectedLandingIncidentId, setSelectedLandingIncidentId] = useState<
    string | null
  >(null);
  const [createIncidentKey, setCreateIncidentKey] = useState("");
  const [createIncidentTitle, setCreateIncidentTitle] = useState("");
  const accountPanelRef = useRef<Phase1AccountPanelHandle | null>(null);
  const adminPanelRef = useRef<Phase1AdminPanelHandle | null>(null);
  const referencePackPanelRef = useRef<ReferencePackAdminPanelHandle | null>(
    null,
  );
  const routeRef = useRef(route);
  const sessionRef = useRef(session);
  const activeRefreshRef = useRef<{
    controller: AbortController | null;
    requestID: number;
  }>({
    controller: null,
    requestID: 0,
  });

  routeRef.current = route;
  sessionRef.current = session;

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

      const [credentialResult, incidentsResult] = await Promise.all([
        loadCredentialState({
          signal: controller.signal,
        }),
        fetchJSON<{ data: { incidents: IncidentData[] } }>(
          "/api/v1/incidents",
          {
            signal: controller.signal,
          },
        ),
      ]);
      if (!canCommit()) {
        return;
      }

      const incidentsError = extractError(incidentsResult.payload);
      const nextCredentialError = extractError(credentialResult.payload);
      const nextSession = (sessionResult.payload as { data: SessionData }).data;

      if (
        (!credentialResult.ok &&
          isSessionRequiredError(
            credentialResult.status,
            nextCredentialError,
          )) ||
        (!incidentsResult.ok &&
          isSessionRequiredError(incidentsResult.status, incidentsError))
      ) {
        setSession(null);
        setCredentialState(null);
        setCredentialError(null);
        setIncidents([]);
        setLandingNotice(null);
        setError(nextCredentialError ?? incidentsError);
        setAuthPrompt(options.anonymousMessage ?? defaultRevokedSessionMessage);
        setAppBootstrapState("revoked");
        setLandingRefreshState("idle");
        return;
      }

      if (!incidentsResult.ok) {
        setSession(nextSession);
        setCredentialState(
          credentialResult.ok
            ? (credentialResult.payload as { data: CredentialState }).data
            : null,
        );
        setCredentialError(nextCredentialError);
        setIncidents([]);
        setLandingNotice(options.landingNotice ?? null);
        setError(incidentsError);
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
      const nextIncidents = (
        incidentsResult.payload as { data: { incidents: IncidentData[] } }
      ).data.incidents;
      const requestedIncidentStillVisible =
        options.routeSnapshot.incidentId === "" ||
        nextIncidents.some(
          (incident) =>
            incident.incident_id === options.routeSnapshot.incidentId,
        );
      const enterpriseRootIncidentID =
        options.routeSnapshot.incidentId === "" &&
        !options.routeSnapshot.debugHarness &&
        nextSession.provider_type !== "local" &&
        nextIncidents.length === 1
          ? (nextIncidents[0]?.incident_id ?? "")
          : "";
      const nextRoute =
        enterpriseRootIncidentID !== ""
          ? {
              incidentId: enterpriseRootIncidentID,
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
    if (incidents.length === 0) {
      setSelectedLandingIncidentId(null);
      return;
    }
    if (
      selectedLandingIncidentId === null ||
      !incidents.some(
        (incident) => incident.incident_id === selectedLandingIncidentId,
      )
    ) {
      setSelectedLandingIncidentId(incidents[0]?.incident_id ?? null);
    }
  }, [incidents, selectedLandingIncidentId]);

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
              : `Loaded ${incidents.length} visible incident${incidents.length === 1 ? "" : "s"}.`);
  const readingProfileAttribute =
    readingProfile === "hyperlegible" ? readingProfile : undefined;
  const rootPageStyle =
    readingProfile === "hyperlegible"
      ? {
          ...pageStyle,
          fontFamily: "var(--ct-typography-accessible-reading-fontFamily)",
        }
      : pageStyle;

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
    setIncidents((current) => upsertIncident(current, incident));
    setLandingNotice(null);
    setError(null);
    setAppBootstrapState("authenticated");
    openIncident(incident.incident_id);
  }, [createIncidentKey, createIncidentTitle, openIncident]);

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

  const selectedLandingIncident = useMemo(
    () =>
      selectedLandingIncidentId === null
        ? null
        : (incidents.find(
            (incident) => incident.incident_id === selectedLandingIncidentId,
          ) ?? null),
    [incidents, selectedLandingIncidentId],
  );
  const sessionIsDeploymentAdmin = session?.is_deployment_admin ?? false;
  const handleAdminCommandStateChange = useCallback(
    (state: Phase1AdminPanelCommandState) => {
      setAdminCommandState(state);
    },
    [],
  );

  const activeAdminPanelCommands = useMemo<LandingAdminCommand[]>(() => {
    switch (activeAdminPanel) {
      case "incidents":
        return [
          {
            id: "refresh",
            label: "Refresh",
            variant: "secondary",
            onClick: () =>
              refreshShell({
                routeSnapshot: routeRef.current,
                landingNotice: null,
              }),
          },
          {
            id: "new incident",
            label: "New incident",
            variant: "secondary",
            onClick: () => {
              const input = document.querySelector<HTMLInputElement>(
                dataTestIdSelector(phase1LandingTestId("incident-key")),
              );
              input?.focus();
            },
          },
          {
            id: "open selected",
            label: "Open selected",
            variant: "primary",
            disabled: selectedLandingIncident === null,
            onClick: () => {
              if (selectedLandingIncident !== null) {
                openIncident(selectedLandingIncident.incident_id);
              }
            },
          },
        ];
      case "account-security":
        return [
          {
            id: "refresh account",
            label: "Refresh account",
            variant: "secondary",
            onClick: () => accountPanelRef.current?.refreshAccount(),
          },
          {
            id: "sign out",
            label: "Sign out",
            variant: "secondary",
            onClick: () => accountPanelRef.current?.signOut(),
          },
        ];
      case "deployment-users":
        return [
          {
            id: "refresh users",
            label: "Refresh users",
            variant: "secondary",
            disabled: !sessionIsDeploymentAdmin,
            onClick: () => adminPanelRef.current?.refreshUsers(),
          },
          {
            id: "create user",
            label: "Create user",
            variant: "primary",
            disabled: !sessionIsDeploymentAdmin,
            onClick: () => adminPanelRef.current?.createUser(),
          },
          {
            id: "load by id",
            label: "Load by ID",
            variant: "secondary",
            disabled:
              !sessionIsDeploymentAdmin || !adminCommandState.canLoadTargetUser,
            onClick: () => adminPanelRef.current?.loadTargetUser(),
          },
          {
            id: "save target",
            label: "Save target",
            variant: "secondary",
            disabled:
              !sessionIsDeploymentAdmin ||
              !adminCommandState.canSubmitVersionedTargetAction,
            onClick: () => adminPanelRef.current?.saveTargetUser(),
          },
          {
            id: "reset password",
            label: "Reset password",
            variant: "danger",
            disabled:
              !sessionIsDeploymentAdmin ||
              !adminCommandState.canSubmitVersionedTargetAction,
            onClick: () => adminPanelRef.current?.resetPassword(),
          },
          {
            id: "reset totp",
            label: "Reset TOTP",
            variant: "danger",
            disabled:
              !sessionIsDeploymentAdmin ||
              !adminCommandState.canSubmitVersionedTargetAction,
            onClick: () => adminPanelRef.current?.resetTotp(),
          },
          {
            id: "revoke sessions",
            label: "Revoke sessions",
            variant: "danger",
            disabled:
              !sessionIsDeploymentAdmin ||
              !adminCommandState.canSubmitTargetAction,
            onClick: () => adminPanelRef.current?.revokeAllSessions(),
          },
        ];
      case "reference-packs":
        return [
          {
            id: "refresh",
            label: "Refresh",
            variant: "secondary",
            disabled: !sessionIsDeploymentAdmin,
            onClick: () => referencePackPanelRef.current?.refreshPacks(),
          },
          {
            id: "import",
            label: "Import",
            variant: "primary",
            disabled: !sessionIsDeploymentAdmin,
            onClick: () => referencePackPanelRef.current?.importBundle(),
          },
          {
            id: "refresh all",
            label: "Refresh all",
            variant: "secondary",
            disabled: !sessionIsDeploymentAdmin,
            onClick: () => referencePackPanelRef.current?.refreshAll(),
          },
          {
            id: "refresh selected",
            label: "Refresh selected",
            variant: "secondary",
            disabled: !sessionIsDeploymentAdmin,
            onClick: () => referencePackPanelRef.current?.refreshSelected(),
          },
          {
            id: "cancel job",
            label: "Cancel job",
            variant: "secondary",
            disabled: !sessionIsDeploymentAdmin,
            onClick: () => referencePackPanelRef.current?.cancelJob(),
          },
        ];
      default:
        return [];
    }
  }, [
    activeAdminPanel,
    adminCommandState,
    openIncident,
    refreshShell,
    selectedLandingIncident,
    sessionIsDeploymentAdmin,
  ]);

  if (route.incidentId !== "" && session !== null) {
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
        <section style={landingPanelStyle}>
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
        commands={activeAdminPanelCommands}
        currentUserLabel={currentUserLabel}
        incidentCount={incidents.length}
        onActivePanelChange={setActiveAdminPanel}
        statusText={landingStatusText}
      >
        {activeAdminPanel === "incidents" ? (
          <IncidentLanding
            bootstrapState={appBootstrapState}
            createIncidentKey={createIncidentKey}
            createIncidentTitle={createIncidentTitle}
            currentUserLabel={currentUserLabel}
            error={error}
            incidents={incidents}
            isRefreshing={landingRefreshState === "loading"}
            onCreate={handleCreateIncident}
            onCreateIncidentKeyChange={setCreateIncidentKey}
            onCreateIncidentTitleChange={setCreateIncidentTitle}
            onOpenIncident={openIncident}
            onRefresh={() => {
              void refreshShell({
                routeSnapshot: routeRef.current,
                landingNotice: null,
              });
            }}
            onSelectIncident={setSelectedLandingIncidentId}
            selectedIncidentId={selectedLandingIncidentId}
            statusText={landingStatusText}
          />
        ) : null}
        <div hidden={activeAdminPanel !== "account-security"}>
          <Phase1AccountPanel
            ref={accountPanelRef}
            credentialState={credentialState}
            credentialStateError={credentialError}
            onRefreshShell={refreshCurrentShell}
            session={session}
          />
        </div>
        <div hidden={activeAdminPanel !== "deployment-users"}>
          <Phase1AdminPanel
            ref={adminPanelRef}
            autoLoadUsers={activeAdminPanel === "deployment-users"}
            onCommandStateChange={handleAdminCommandStateChange}
            onRefreshShell={refreshCurrentShell}
            session={session}
          />
        </div>
        <div hidden={activeAdminPanel !== "reference-packs"}>
          <ReferencePackAdminPanel
            ref={referencePackPanelRef}
            session={session}
          />
        </div>
      </LandingAdminShell>
    </main>
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

const landingAdminPanels = [
  {
    token: "incidents",
    label: "Incidents",
    description: "Open and create workbook access",
  },
  {
    token: "account-security",
    label: "Account security",
    description: "Session and credentials",
  },
  {
    token: "deployment-users",
    label: "Deployment users",
    description: "Local user administration",
  },
  {
    token: "reference-packs",
    label: "Reference packs",
    description: "Pack operations",
  },
] as const satisfies ReadonlyArray<{
  description: string;
  label: string;
  token: LandingAdminPanelToken;
}>;

function LandingAdminShell({
  activePanel,
  children,
  commands,
  currentUserLabel,
  incidentCount,
  onActivePanelChange,
  statusText,
}: LandingAdminShellProps) {
  const tabRefs = useRef(new Map<LandingAdminPanelToken, HTMLButtonElement>());
  const activePanelMeta =
    landingAdminPanels.find((panel) => panel.token === activePanel) ??
    landingAdminPanels[0];
  const panelId = landingAdminPanelTestId(activePanel);
  const tabId = landingAdminTabTestId(activePanel);

  function focusPanelTab(panel: LandingAdminPanelToken) {
    const focus = () => {
      tabRefs.current.get(panel)?.focus();
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
      focusPanelTab(panel);
    }
  }

  function handleTabKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    const currentIndex = landingAdminPanels.findIndex(
      (panel) => panel.token === activePanel,
    );
    const lastIndex = landingAdminPanels.length - 1;
    const selectByIndex = (index: number) => {
      event.preventDefault();
      selectPanel(landingAdminPanels[index]?.token ?? "incidents", true);
    };

    switch (event.key) {
      case "ArrowLeft":
        selectByIndex(currentIndex <= 0 ? lastIndex : currentIndex - 1);
        return;
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
          <h1 style={landingAdminTitleStyle}>Incident administration</h1>
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
            <dt style={landingToolbarLabelStyle}>Visible incidents</dt>
            <dd style={landingAdminMetaValueStyle}>{incidentCount}</dd>
          </div>
        </dl>
      </header>

      <nav style={landingAdminRibbonStyle} aria-label="Admin panels">
        <div
          data-testid={landingAdminShellTestId("tablist")}
          role="tablist"
          aria-label="Incident administration panels"
          style={landingAdminTabsStyle}
          onKeyDown={handleTabKeyDown}
        >
          {landingAdminPanels.map((panel) => {
            const selected = panel.token === activePanel;
            return (
              <button
                key={panel.token}
                id={landingAdminTabTestId(panel.token)}
                ref={(element) => {
                  if (element === null) {
                    tabRefs.current.delete(panel.token);
                    return;
                  }
                  tabRefs.current.set(panel.token, element);
                }}
                aria-controls={landingAdminPanelTestId(panel.token)}
                aria-selected={selected}
                data-testid={landingAdminTabTestId(panel.token)}
                role="tab"
                style={
                  selected ? landingAdminTabSelectedStyle : landingAdminTabStyle
                }
                tabIndex={selected ? 0 : -1}
                type="button"
                onClick={() => {
                  selectPanel(panel.token);
                }}
              >
                <span style={landingAdminTabLabelStyle}>{panel.label}</span>
                <span style={landingAdminTabDescriptionStyle}>
                  {panel.description}
                </span>
              </button>
            );
          })}
        </div>

        <div
          data-testid={landingAdminShellTestId("command-strip")}
          style={landingAdminCommandStripStyle}
        >
          <div style={landingAdminCommandLabelStyle}>
            {activePanelMeta.label}
          </div>
          <div style={landingAdminCommandButtonsStyle}>
            {commands.map((command) => (
              <button
                key={command.id}
                data-testid={landingAdminCommandTestId(activePanel, command.id)}
                disabled={command.disabled}
                style={commandButtonStyle(command.variant)}
                type="button"
                onClick={() => {
                  void command.onClick();
                }}
              >
                {command.label}
              </button>
            ))}
          </div>
        </div>
      </nav>

      <section
        id={panelId}
        aria-labelledby={tabId}
        data-testid={panelId}
        role="tabpanel"
        style={landingAdminPanelRegionStyle}
      >
        {children}
      </section>

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

function commandButtonStyle(
  variant: LandingAdminCommand["variant"] = "secondary",
): CSSProperties {
  if (variant === "primary") {
    return landingAdminPrimaryCommandStyle;
  }
  if (variant === "danger") {
    return landingAdminDangerCommandStyle;
  }
  return landingAdminSecondaryCommandStyle;
}

const landingAdminShellStyle: CSSProperties = {
  width: "min(96rem, 100%)",
  minHeight: "calc(100vh - 4rem)",
  margin: "0 auto",
  display: "grid",
  gridTemplateRows:
    "auto auto minmax(0, 1fr) var(--ct-layout-statusStripHeight)",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-lg)",
  background: "var(--ct-colors-surface-1)",
  boxShadow: "var(--ct-elevation-panel)",
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

const landingAdminRibbonStyle: CSSProperties = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
  padding: "var(--ct-spacing-sm) var(--ct-spacing-lg)",
  borderBottom: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
};

const landingAdminTabsStyle: CSSProperties = {
  display: "flex",
  gap: "var(--ct-spacing-xs)",
  overflowX: "auto",
};

const landingAdminTabStyle: CSSProperties = {
  display: "grid",
  gap: "var(--ct-spacing-xxs)",
  minWidth: "11rem",
  padding: "var(--ct-spacing-sm) var(--ct-spacing-md)",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-md)",
  background: "var(--ct-colors-surface-1)",
  color: "var(--ct-colors-ink-muted)",
  textAlign: "left",
  cursor: "pointer",
};

const landingAdminTabSelectedStyle: CSSProperties = {
  ...landingAdminTabStyle,
  border: "var(--ct-border-strong)",
  background: "var(--ct-colors-surface-3)",
  color: "var(--ct-colors-ink)",
  boxShadow: "inset 0 -2px 0 var(--ct-colors-accent)",
};

const landingAdminTabLabelStyle: CSSProperties = {
  fontWeight: 700,
};

const landingAdminTabDescriptionStyle: CSSProperties = {
  color: "var(--ct-colors-ink-subtle)",
  fontSize: "0.78rem",
};

const landingAdminCommandStripStyle: CSSProperties = {
  display: "flex",
  alignItems: "center",
  justifyContent: "space-between",
  gap: "var(--ct-spacing-md)",
  minHeight: "var(--ct-layout-viewBarHeight)",
};

const landingAdminCommandLabelStyle: CSSProperties = {
  color: "var(--ct-colors-ink-muted)",
  fontSize: "var(--ct-typography-metadata-fontSize)",
  fontWeight: 700,
};

const landingAdminCommandButtonsStyle: CSSProperties = {
  display: "flex",
  flexWrap: "wrap",
  justifyContent: "flex-end",
  gap: "var(--ct-spacing-xs)",
};

const landingAdminPanelRegionStyle: CSSProperties = {
  minWidth: 0,
  minHeight: 0,
  padding: "var(--ct-spacing-lg)",
  overflow: "auto",
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

const landingAdminSecondaryCommandStyle: CSSProperties = {
  borderRadius: "var(--ct-component-button-secondary-rounded)",
  border: "var(--ct-component-button-secondary-border)",
  padding: "var(--ct-component-button-secondary-padding)",
  background: "var(--ct-component-button-secondary-backgroundColor)",
  color: "var(--ct-component-button-secondary-textColor)",
  font: "inherit",
  fontWeight: 700,
  cursor: "pointer",
};

const landingAdminPrimaryCommandStyle: CSSProperties = {
  ...landingAdminSecondaryCommandStyle,
  border: "none",
  background: "var(--ct-component-button-primary-backgroundColor)",
  color: "var(--ct-component-button-primary-textColor)",
};

const landingAdminDangerCommandStyle: CSSProperties = {
  ...landingAdminSecondaryCommandStyle,
  color: "var(--ct-colors-semantic-destructive)",
};

const pageStyle = {
  minHeight: "100vh",
  margin: 0,
  padding: "2rem",
  background: "var(--ct-colors-canvas)",
  color: "var(--ct-colors-ink)",
  fontFamily: "var(--ct-typography-ui-fontFamily)",
  fontSize: "var(--ct-typography-ui-fontSize)",
  lineHeight: "var(--ct-typography-ui-lineHeight)",
};

const landingPanelStyle = {
  width: "min(78rem, 100%)",
  margin: "0 auto",
  padding: "2rem",
  borderRadius: "var(--ct-rounded-lg)",
  background: "var(--ct-colors-surface-1)",
  boxShadow: "var(--ct-elevation-panel)",
  border: "var(--ct-border-hairline)",
};

const workbookFrameStyle = {
  width: "min(96rem, 100%)",
  margin: "0 auto",
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

const landingToolbarStyle = {
  display: "flex",
  justifyContent: "space-between",
  alignItems: "center",
  gap: "1rem",
  marginBottom: "1.5rem",
  padding: "1rem 1.2rem",
  borderRadius: "var(--ct-rounded-lg)",
  background: "var(--ct-colors-surface-2)",
};

const landingToolbarLabelStyle = {
  margin: 0,
  fontSize: "0.72rem",
  letterSpacing: "0.14em",
  textTransform: "uppercase" as const,
  color: "var(--ct-colors-ink-subtle)",
};

const landingToolbarValueStyle = {
  margin: "0.3rem 0 0",
  fontWeight: 600,
};

const landingSectionStyle = {
  marginTop: "1.5rem",
  padding: "1.25rem",
  borderRadius: "var(--ct-rounded-lg)",
  background: "var(--ct-colors-surface-2)",
  border: "var(--ct-border-hairline)",
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
  gap: "0.9rem",
};

const landingIncidentCardStyle = {
  display: "flex",
  justifyContent: "space-between",
  alignItems: "center",
  gap: "1rem",
  padding: "1rem",
  borderRadius: "var(--ct-rounded-lg)",
  background: "var(--ct-colors-surface-1)",
  border: "var(--ct-border-hairline)",
};

const selectedLandingIncidentCardStyle = {
  ...landingIncidentCardStyle,
  border: "var(--ct-border-strong)",
  background: "var(--ct-colors-surface-3)",
  boxShadow: "inset 3px 0 0 var(--ct-colors-accent)",
};

const landingIncidentSelectButtonStyle = {
  minWidth: 0,
  flex: "1 1 auto",
  padding: 0,
  border: "none",
  background: "transparent",
  color: "inherit",
  textAlign: "left" as const,
  cursor: "pointer",
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

const landingIncidentMetaStyle = {
  margin: "0.4rem 0 0",
  color: "var(--ct-colors-ink-muted)",
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
