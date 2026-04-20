import {
  startTransition,
  useCallback,
  useEffect,
  useId,
  useMemo,
  useState,
} from "react";

import {
  type APIError,
  clientTxnID,
  extractError,
  fetchJSON,
} from "./browserApi";
import { Phase1Harness } from "./Phase1Harness";
import {
  Phase1AccountPanel,
  Phase1AdminPanel,
  Phase1AuthSurface,
} from "./Phase1Surface";
import { Phase2Harness } from "./Phase2Harness";
import {
  type CredentialState,
  loadCredentialState,
  loadSession,
  type SessionData,
} from "./phase1Client";
import {
  buildCreatePayload,
  createDraftRow,
  ensureDraftRow,
  TimelineWorkbook,
  WorkbookShell,
} from "./WorkbookShell";

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

type IncidentLandingProps = {
  createIncidentKey: string;
  createIncidentTitle: string;
  currentUserLabel: string;
  error: APIError | null;
  incidents: IncidentData[];
  isLoading: boolean;
  onCreate: () => Promise<void> | void;
  onCreateIncidentKeyChange: (value: string) => void;
  onCreateIncidentTitleChange: (value: string) => void;
  onOpenIncident: (incidentId: string) => void;
  onRefresh: () => Promise<void> | void;
  statusText: string;
};

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

function incidentSummaryText(incident: IncidentData): string {
  return [incident.title, incident.incident_key]
    .filter((segment) => segment.trim() !== "")
    .join(" · ");
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

export function IncidentLanding({
  createIncidentKey,
  createIncidentTitle,
  currentUserLabel,
  error,
  incidents,
  isLoading,
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
    <section style={landingPanelStyle}>
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
            data-testid="landing-current-user"
            style={landingToolbarValueStyle}
          >
            {currentUserLabel}
          </p>
        </div>
        <button
          data-testid="landing-refresh"
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
            data-testid="landing-incident-key"
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
            data-testid="landing-incident-title"
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
          data-testid="landing-create-button"
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
          <p data-testid="landing-incidents-count" style={landingCountStyle}>
            {incidents.length}
          </p>
        </div>

        {isLoading ? (
          <p data-testid="landing-loading" style={landingBodyStyle}>
            Loading visible incidents…
          </p>
        ) : null}

        {!isLoading && !hasIncidents ? (
          <p data-testid="landing-empty-state" style={landingBodyStyle}>
            No incidents are visible for this session yet.
          </p>
        ) : null}

        {hasIncidents ? (
          <div data-testid="landing-incident-list" style={landingListStyle}>
            {incidents.map((incident) => (
              <article
                key={incident.incident_id}
                data-testid={`landing-incident-${incident.incident_id}`}
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
                  data-testid={`landing-open-${incident.incident_id}`}
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

      <p data-testid="landing-status" style={landingStatusStyle}>
        {statusText}
      </p>
      <p data-testid="landing-error-code" style={landingErrorStyle}>
        {error?.code ?? ""}
      </p>
    </section>
  );
}

export function App() {
  const defaultAuthPrompt =
    "Sign in with your local account to open incidents, manage account security, and administer deployment users.";
  const [route, setRoute] = useState<RouteState>(() => readRouteState());
  const [session, setSession] = useState<SessionData | null>(null);
  const [credentialState, setCredentialState] =
    useState<CredentialState | null>(null);
  const [credentialError, setCredentialError] = useState<APIError | null>(null);
  const [incidents, setIncidents] = useState<IncidentData[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [statusText, setStatusText] = useState("Loading visible incidents…");
  const [pinnedStatusText, setPinnedStatusText] = useState<string | null>(null);
  const [error, setError] = useState<APIError | null>(null);
  const [authPrompt, setAuthPrompt] = useState(defaultAuthPrompt);
  const [createIncidentKey, setCreateIncidentKey] = useState("");
  const [createIncidentTitle, setCreateIncidentTitle] = useState("");

  const loadLanding = useCallback(
    async (options?: { anonymousMessage?: string; staleMessage?: string }) => {
      setIsLoading(true);

      const sessionResult = await loadSession();
      if (!sessionResult.ok) {
        setSession(null);
        setCredentialState(null);
        setCredentialError(null);
        setIncidents([]);
        setError(null);
        setStatusText(
          options?.anonymousMessage ??
            "Sign in required for the incident shell.",
        );
        setAuthPrompt(options?.anonymousMessage ?? defaultAuthPrompt);
        setIsLoading(false);
        return;
      }

      const [credentialResult, incidentsResult] = await Promise.all([
        loadCredentialState(),
        fetchJSON<{ data: { incidents: IncidentData[] } }>("/api/v1/incidents"),
      ]);

      const incidentsError = extractError(incidentsResult.payload);
      const nextCredentialError = extractError(credentialResult.payload);
      if (!incidentsResult.ok) {
        setSession((sessionResult.payload as { data: SessionData }).data);
        setCredentialState(
          credentialResult.ok
            ? (credentialResult.payload as { data: CredentialState }).data
            : null,
        );
        setCredentialError(nextCredentialError);
        setIncidents([]);
        setError(incidentsError);
        setStatusText("Failed to load visible incidents.");
        setIsLoading(false);
        return;
      }

      const nextSession = (sessionResult.payload as { data: SessionData }).data;
      const nextCredentialState = credentialResult.ok
        ? (credentialResult.payload as { data: CredentialState }).data
        : null;
      const nextIncidents = (
        incidentsResult.payload as { data: { incidents: IncidentData[] } }
      ).data.incidents;
      const requestedIncidentStillVisible =
        route.incidentId === "" ||
        nextIncidents.some(
          (incident) => incident.incident_id === route.incidentId,
        );

      setSession(nextSession);
      setCredentialState(nextCredentialState);
      setCredentialError(nextCredentialError);
      setIncidents(nextIncidents);
      setError(null);
      setAuthPrompt(defaultAuthPrompt);
      setIsLoading(false);

      if (!requestedIncidentStillVisible) {
        const nextRoute = { incidentId: "", debugHarness: route.debugHarness };
        const staleMessage =
          options?.staleMessage ??
          "The requested incident is no longer visible.";
        writeRouteState(nextRoute, "replace");
        startTransition(() => {
          setRoute(nextRoute);
        });
        setPinnedStatusText(staleMessage);
        setStatusText(staleMessage);
        return;
      }

      if (options?.staleMessage) {
        setPinnedStatusText(options.staleMessage);
        setStatusText(options.staleMessage);
        return;
      }
      if (pinnedStatusText !== null) {
        setStatusText(pinnedStatusText);
        return;
      }
      setStatusText(
        nextIncidents.length === 0
          ? "No visible incidents yet."
          : `Loaded ${nextIncidents.length} visible incident${nextIncidents.length === 1 ? "" : "s"}.`,
      );
    },
    [pinnedStatusText, route.debugHarness, route.incidentId],
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
    void loadLanding();
  }, [loadLanding]);

  const currentUserLabel = useMemo(() => {
    if (session === null) {
      return "Anonymous";
    }
    return `${session.display_name}${session.is_deployment_admin ? " · deployment admin" : ""}`;
  }, [session]);

  const openIncident = useCallback(
    (incidentId: string) => {
      const nextRoute = { incidentId, debugHarness: route.debugHarness };
      setPinnedStatusText(null);
      writeRouteState(nextRoute, "push");
      startTransition(() => {
        setRoute(nextRoute);
      });
      setStatusText("Opened workbook.");
    },
    [route.debugHarness],
  );

  const handleCreateIncident = useCallback(async () => {
    const incidentKey = createIncidentKey.trim();
    const title = createIncidentTitle.trim();
    if (incidentKey === "" || title === "") {
      setStatusText("Incident key and title are required.");
      return;
    }

    setPinnedStatusText(null);
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
      setError(extractError(response.payload));
      setStatusText("Incident create failed.");
      return;
    }

    const incident = (response.payload as { data: IncidentData }).data;
    setCreateIncidentKey("");
    setCreateIncidentTitle("");
    setIncidents((current) => upsertIncident(current, incident));
    setError(null);
    setStatusText(`Created incident ${incidentSummaryText(incident)}.`);
    openIncident(incident.incident_id);
  }, [createIncidentKey, createIncidentTitle, openIncident]);

  const handleIncidentAccessLost = useCallback(() => {
    void loadLanding({
      staleMessage:
        "The current incident is no longer visible. Returned to the landing screen.",
    });
  }, [loadLanding]);

  const handleIncidentSnapshot = useCallback((incident: IncidentData) => {
    setIncidents((current) => upsertIncident(current, incident));
  }, []);

  if (route.incidentId !== "" && session !== null) {
    return (
      <main style={pageStyle}>
        <section style={workbookFrameStyle}>
          <div style={workbookToolbarStyle}>
            <div>
              <p style={landingToolbarLabelStyle}>Workbook</p>
              <p
                data-testid="workbook-current-user"
                style={landingToolbarValueStyle}
              >
                {currentUserLabel}
              </p>
            </div>
            <button
              data-testid="landing-return"
              style={landingSecondaryButtonStyle}
              type="button"
              onClick={() => {
                const nextRoute = {
                  incidentId: "",
                  debugHarness: route.debugHarness,
                };
                writeRouteState(nextRoute, "push");
                startTransition(() => {
                  setRoute(nextRoute);
                });
                setStatusText("Returned to incident landing.");
              }}
            >
              Incident landing
            </button>
          </div>

          <WorkbookShell
            incidentId={route.incidentId}
            onIncidentAccessLost={handleIncidentAccessLost}
            onIncidentSnapshot={handleIncidentSnapshot}
          />
        </section>
      </main>
    );
  }

  if (route.debugHarness) {
    return (
      <main style={pageStyle}>
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

          <Phase1Harness />
          <Phase2Harness />
        </section>
      </main>
    );
  }

  if (session === null) {
    return (
      <Phase1AuthSurface
        isLoading={isLoading}
        message={authPrompt}
        onAuthenticated={async () => {
          setPinnedStatusText(null);
          await loadLanding();
        }}
      />
    );
  }

  return (
    <main style={pageStyle}>
      <div style={shellStackStyle}>
        <IncidentLanding
          createIncidentKey={createIncidentKey}
          createIncidentTitle={createIncidentTitle}
          currentUserLabel={currentUserLabel}
          error={error}
          incidents={incidents}
          isLoading={isLoading}
          onCreate={handleCreateIncident}
          onCreateIncidentKeyChange={setCreateIncidentKey}
          onCreateIncidentTitleChange={setCreateIncidentTitle}
          onOpenIncident={openIncident}
          onRefresh={() => {
            setPinnedStatusText(null);
            void loadLanding();
          }}
          statusText={statusText}
        />
        <section style={supportPanelGridStyle}>
          <Phase1AccountPanel
            credentialState={credentialState}
            credentialStateError={credentialError}
            onRefreshShell={loadLanding}
            session={session}
          />
          <Phase1AdminPanel onRefreshShell={loadLanding} session={session} />
        </section>
      </div>
    </main>
  );
}

const pageStyle = {
  minHeight: "100vh",
  margin: 0,
  padding: "2rem",
  background:
    "linear-gradient(135deg, rgb(242 236 225) 0%, rgb(229 235 225) 50%, rgb(214 229 224) 100%)",
  color: "rgb(24 38 35)",
  fontFamily: '"IBM Plex Sans", "Segoe UI", sans-serif',
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

const workbookToolbarStyle = {
  ...landingToolbarStyle,
  marginBottom: "1rem",
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

const landingErrorStyle = {
  margin: "0.25rem 0 0",
  minHeight: "1.25rem",
  color: "rgb(147 47 47)",
  fontWeight: 600,
};

export { buildCreatePayload, createDraftRow, ensureDraftRow, TimelineWorkbook };
