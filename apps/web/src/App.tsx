import {
  landingIncidentCardTestId,
  landingIncidentOpenButtonTestId,
  phase1ErrorCodeTestId,
  phase1ErrorSummaryTestIds,
  phase1LandingTestId,
  phase1RouteTestId,
} from "@cartulary/ui-contracts";
import {
  lazy,
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
} from "./browserApi";
import {
  Phase1AccountPanel,
  Phase1AdminPanel,
  Phase1AuthSurface,
} from "./Phase1Surface";
import {
  type CredentialState,
  loadCredentialState,
  loadSession,
  type SessionData,
} from "./phase1Client";
import { ReferencePackAdminPanel } from "./ReferencePackAdminPanel";

const LazyWorkbookShell = lazy(async () => {
  const module = await import("./WorkbookShell");
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
  statusText: string;
};

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
          <p
            data-testid={phase1LandingTestId("current-user")}
            style={landingToolbarValueStyle}
          >
            {currentUserLabel}
          </p>
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
            {incidents.length}
          </p>
        </div>

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
                <div style={landingIncidentTextStyle}>
                  <p style={landingIncidentKeyStyle}>{incident.incident_key}</p>
                  <h3 style={landingIncidentTitleStyle}>{incident.title}</h3>
                  <p style={landingIncidentMetaStyle}>
                    Version {incident.incident_version}
                    {incident.current_phase
                      ? ` · ${incident.current_phase}`
                      : ""}
                    {incident.tlp ? ` · ${incident.tlp}` : ""}
                  </p>
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

export function App() {
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
  const [createIncidentKey, setCreateIncidentKey] = useState("");
  const [createIncidentTitle, setCreateIncidentTitle] = useState("");
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
      const nextRoute = requestedIncidentStillVisible
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

  if (route.incidentId !== "" && session !== null) {
    return (
      <main
        aria-busy={appBootstrapState === "loading"}
        className="cartulary-shell"
        data-bootstrap-state={appBootstrapState}
        data-testid={phase1RouteTestId("app-shell")}
        style={pageStyle}
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
      <main className="cartulary-shell" style={pageStyle}>
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
      />
    );
  }

  return (
    <main
      aria-busy={appBootstrapState === "loading"}
      className="cartulary-shell"
      data-bootstrap-state={appBootstrapState}
      data-testid={phase1RouteTestId("app-shell")}
      style={pageStyle}
    >
      <div style={shellStackStyle}>
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
          statusText={landingStatusText}
        />
        <section style={supportPanelGridStyle}>
          <Phase1AccountPanel
            credentialState={credentialState}
            credentialStateError={credentialError}
            onRefreshShell={refreshCurrentShell}
            session={session}
          />
          <Phase1AdminPanel
            onRefreshShell={refreshCurrentShell}
            session={session}
          />
          <ReferencePackAdminPanel session={session} />
        </section>
      </div>
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

const pageStyle = {
  minHeight: "100vh",
  margin: 0,
  padding: "2rem",
  background:
    "linear-gradient(135deg, rgb(242 236 225) 0%, rgb(229 235 225) 50%, rgb(214 229 224) 100%)",
  color: "rgb(24 38 35)",
  fontFamily: "var(--font-ui)",
};

const landingPanelStyle = {
  width: "min(78rem, 100%)",
  margin: "0 auto",
  padding: "2rem",
  borderRadius: "1.5rem",
  background: "rgb(255 251 244 / 0.94)",
  boxShadow: "0 24px 80px rgb(29 78 70 / 0.12)",
  border: "1px solid rgb(185 204 196 / 0.8)",
};

const workbookFrameStyle = {
  width: "min(96rem, 100%)",
  margin: "0 auto",
};

const shellStackStyle = {
  width: "min(96rem, 100%)",
  margin: "0 auto",
  display: "grid",
  gap: "1.5rem",
};

const supportPanelGridStyle = {
  display: "grid",
  gridTemplateColumns: "repeat(auto-fit, minmax(22rem, 1fr))",
  gap: "1.5rem",
  alignItems: "start",
};

const landingHeroStyle = {
  marginBottom: "1.5rem",
};

const landingEyebrowStyle = {
  margin: 0,
  fontSize: "0.78rem",
  letterSpacing: "0.18em",
  textTransform: "uppercase" as const,
  color: "rgb(55 92 86)",
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
  color: "rgb(67 90 85)",
};

const landingToolbarStyle = {
  display: "flex",
  justifyContent: "space-between",
  alignItems: "center",
  gap: "1rem",
  marginBottom: "1.5rem",
  padding: "1rem 1.2rem",
  borderRadius: "1rem",
  background: "rgb(233 241 236)",
};

const landingToolbarLabelStyle = {
  margin: 0,
  fontSize: "0.72rem",
  letterSpacing: "0.14em",
  textTransform: "uppercase" as const,
  color: "rgb(70 103 96)",
};

const landingToolbarValueStyle = {
  margin: "0.3rem 0 0",
  fontWeight: 600,
};

const landingSectionStyle = {
  marginTop: "1.5rem",
  padding: "1.25rem",
  borderRadius: "1.1rem",
  background: "rgb(248 244 236)",
  border: "1px solid rgb(218 227 219)",
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
  color: "rgb(75 108 100)",
};

const landingSectionTitleStyle = {
  margin: "0.25rem 0 0",
  fontSize: "1.35rem",
};

const landingCountStyle = {
  margin: 0,
  minWidth: "2.5rem",
  padding: "0.5rem 0.75rem",
  borderRadius: "999px",
  background: "rgb(215 228 221)",
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
  borderRadius: "0.85rem",
  border: "1px solid rgb(191 207 199)",
  background: "white",
  color: "inherit",
};

const landingPrimaryButtonStyle = {
  padding: "0.8rem 1rem",
  borderRadius: "0.85rem",
  border: "none",
  background: "rgb(37 89 79)",
  color: "white",
  fontWeight: 600,
  cursor: "pointer",
};

const landingSecondaryButtonStyle = {
  padding: "0.8rem 1rem",
  borderRadius: "0.85rem",
  border: "1px solid rgb(157 182 174)",
  background: "rgb(250 253 250)",
  color: "rgb(28 64 58)",
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
  borderRadius: "1rem",
  background: "rgb(255 255 255 / 0.92)",
  border: "1px solid rgb(211 223 216)",
};

const landingIncidentTextStyle = {
  minWidth: 0,
};

const landingIncidentKeyStyle = {
  margin: 0,
  fontSize: "0.78rem",
  letterSpacing: "0.14em",
  textTransform: "uppercase" as const,
  color: "rgb(75 108 100)",
};

const landingIncidentTitleStyle = {
  margin: "0.35rem 0 0",
  fontSize: "1.05rem",
};

const landingIncidentMetaStyle = {
  margin: "0.4rem 0 0",
  color: "rgb(89 109 103)",
};

const landingStatusStyle = {
  margin: "1.25rem 0 0",
  minHeight: "1.5rem",
  color: "rgb(45 82 75)",
};

const routeLoadingStyle = {
  margin: "1rem 0 0",
  padding: "1rem 1.2rem",
  borderRadius: "1rem",
  background: "rgb(233 241 236)",
  color: "rgb(45 82 75)",
};

const landingErrorStyle = {
  margin: "0.25rem 0 0",
  minHeight: "1.25rem",
  color: "rgb(147 47 47)",
  fontWeight: 600,
};

const publicErrorStyle = {
  marginTop: "0.25rem",
};

const errorMessageStyle = {
  margin: 0,
  minHeight: "1.25rem",
  color: "rgb(126 45 45)",
};

const errorDetailStyle = {
  margin: "0.2rem 0 0",
  minHeight: "1.25rem",
  color: "rgb(126 45 45)",
  overflowWrap: "anywhere" as const,
};
