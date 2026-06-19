import {
  type LandingAdminPanelToken,
  landingAdminMenuItemTestId,
  landingAdminPanelTestId,
  phase1RouteTestId,
} from "@cartulary/ui-contracts";
import {
  type CSSProperties,
  lazy,
  Suspense,
  startTransition,
  useCallback,
  useEffect,
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
  AccountAppearancePanel,
  AccountProfilePanel,
  AdministrativeAuditPanel,
  type AppBootstrapState,
  IncidentBundleImportPanel,
  type IncidentData,
  IncidentLanding,
  type IncidentStatusFilter,
  type LandingAdminPanelDescriptor,
  LandingAdminShell,
  type LandingRefreshState,
  type PagingMeta,
} from "./LandingAdminSurface";
import {
  Phase1AccountPanel,
  Phase1AdminPanel,
  Phase1AuthSurface,
} from "./Phase1Surface";
import {
  type CredentialState,
  type ExtensionProfileResource,
  loadCredentialState,
  loadExtensions,
  loadSession,
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

type AppProps = {
  readonly readingProfile?: CartularyReadingProfile | undefined;
  readonly themeId?: string | undefined;
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
    setIncidentSearch(value);
  }, []);

  const handleIncidentSearchSubmit = useCallback(
    () =>
      refreshShell({
        routeSnapshot: routeRef.current,
        landingNotice: null,
      }),
    [refreshShell],
  );

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
          ? "Searching visible incidents…"
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
            onSearchSubmit={handleIncidentSearchSubmit}
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
              enterpriseAuthClaimed={extensionClaimed(
                extensionProfiles,
                "enterprise_authentication",
              )}
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

const landingAdminPanelRegionStyle: CSSProperties = {
  minWidth: 0,
  minHeight: 0,
  padding: "var(--ct-spacing-md)",
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

const routeLoadingStyle = {
  margin: "1rem 0 0",
  padding: "1rem 1.2rem",
  borderRadius: "var(--ct-rounded-lg)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-ink-muted)",
};
