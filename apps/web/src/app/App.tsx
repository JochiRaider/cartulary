import {
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
import type { WorkbookAccountApplicationMenuProps } from "../workbook/WorkbookShell";
import {
  AccountAppearancePanel,
  AccountApplicationMenu,
  AccountProfilePanel,
  type AccountSettingsPanelToken,
  AdministrativeAuditPanel,
  type AppBootstrapState,
  type DeploymentAdministrationPanelToken,
  type IncidentData,
  IncidentDirectoryShell,
  IncidentImportPanel,
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
import {
  ReferencePackAdminPanel,
  type ReferencePackJobResource,
} from "./ReferencePackAdminPanel";

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
  deploymentAdministration: boolean;
  manualIncidentDirectory: boolean;
};

type ShellRefreshOptions = {
  anonymousMessage?: string;
  landingNotice?: string | null;
  routeSnapshot: RouteState;
};

type AccountMenuContext =
  | "deployment-administration"
  | "incidents"
  | "workbook";

type AccountMenuOptions = {
  currentIncidentRole?: WorkbookAccountApplicationMenuProps["currentIncidentRole"];
  incidentControls?: WorkbookAccountApplicationMenuProps["incidentControls"];
  onOpenIncidentDirectory?: (() => void) | undefined;
  triggerTestId?: string | undefined;
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
const emptyIncidentsPaging: PagingMeta = {
  limit: 100,
  has_more: false,
  next_cursor: null,
};

function readRouteState(): RouteState {
  const params = new URLSearchParams(window.location.search);
  const deploymentAdministration =
    window.location.pathname === "/deployment-administration";
  const incidentId = deploymentAdministration
    ? ""
    : (params.get("incident_id") ?? "").trim();
  const historyState =
    typeof window.history.state === "object" && window.history.state !== null
      ? (window.history.state as { cartularyIncidentDirectory?: unknown })
      : null;
  return {
    incidentId,
    debugHarness:
      !deploymentAdministration && params.get("debug") === "harness",
    deploymentAdministration,
    manualIncidentDirectory:
      !deploymentAdministration &&
      incidentId === "" &&
      historyState?.cartularyIncidentDirectory === true,
  };
}

function writeRouteState(next: RouteState, mode: "push" | "replace") {
  const params = new URLSearchParams(window.location.search);
  const historyState = next.manualIncidentDirectory
    ? { cartularyIncidentDirectory: true }
    : {};
  if (next.deploymentAdministration) {
    params.delete("incident_id");
    params.delete("surface");
    params.delete("debug");
    const query = params.toString();
    const url =
      query === ""
        ? "/deployment-administration"
        : `/deployment-administration?${query}`;
    if (mode === "push") {
      window.history.pushState({}, "", url);
      return;
    }
    window.history.replaceState({}, "", url);
    return;
  }

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
    window.history.pushState(historyState, "", url);
    return;
  }
  window.history.replaceState(historyState, "", url);
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
  const [, setCredentialState] = useState<CredentialState | null>(null);
  const [credentialError, setCredentialError] = useState<APIError | null>(null);
  const [incidents, setIncidents] = useState<IncidentData[]>([]);
  const [incidentsPaging, setIncidentsPaging] =
    useState<PagingMeta>(emptyIncidentsPaging);
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
  const [activeDeploymentPanel, setActiveDeploymentPanel] =
    useState<DeploymentAdministrationPanelToken>("deployment-users");
  const [accountSettingsPanel, setAccountSettingsPanel] =
    useState<AccountSettingsPanelToken | null>(null);
  const [referencePackJob, setReferencePackJob] =
    useState<ReferencePackJobResource | null>(null);
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
  const suppressRootIncidentAutoOpenRef = useRef(false);
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
        setIncidentsPaging(emptyIncidentsPaging);
        setExtensionProfiles([]);
        setReferencePackJob(null);
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

      const nextSession = (sessionResult.payload as { data: SessionData }).data;
      const deploymentRouteDenied =
        options.routeSnapshot.deploymentAdministration &&
        !nextSession.is_deployment_admin;
      const effectiveRouteSnapshot = deploymentRouteDenied
        ? {
            incidentId: "",
            debugHarness: false,
            deploymentAdministration: false,
            manualIncidentDirectory: false,
          }
        : options.routeSnapshot;
      const shouldLoadIncidentDirectory =
        !effectiveRouteSnapshot.deploymentAdministration;
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
        shouldLoadIncidentDirectory
          ? fetchJSON<IncidentListEnvelope>(
              incidentListURL({ limit: 2, statusFilter: "all" }),
              { signal: controller.signal },
            )
          : Promise.resolve(null),
        shouldLoadIncidentDirectory
          ? fetchJSON<IncidentListEnvelope>(
              incidentListURL({
                limit: 100,
                search: directoryScope.search,
                statusFilter: directoryScope.statusFilter,
              }),
              { signal: controller.signal },
            )
          : Promise.resolve(null),
      ]);
      if (!canCommit()) {
        return;
      }

      const incidentsError =
        incidentsResult === null ? null : extractError(incidentsResult.payload);
      const bootstrapIncidentsError =
        bootstrapIncidentsResult === null
          ? null
          : extractError(bootstrapIncidentsResult.payload);
      const extensionsError = extractError(extensionsResult.payload);
      const nextCredentialError = extractError(credentialResult.payload);

      if (
        (!credentialResult.ok &&
          isSessionRequiredError(
            credentialResult.status,
            nextCredentialError,
          )) ||
        (!extensionsResult.ok &&
          isSessionRequiredError(extensionsResult.status, extensionsError)) ||
        (bootstrapIncidentsResult !== null &&
          !bootstrapIncidentsResult.ok &&
          isSessionRequiredError(
            bootstrapIncidentsResult.status,
            bootstrapIncidentsError,
          )) ||
        (incidentsResult !== null &&
          !incidentsResult.ok &&
          isSessionRequiredError(incidentsResult.status, incidentsError))
      ) {
        setSession(null);
        setCredentialState(null);
        setCredentialError(null);
        setIncidents([]);
        setIncidentsPaging(emptyIncidentsPaging);
        setExtensionProfiles([]);
        setReferencePackJob(null);
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
        (bootstrapIncidentsResult !== null && !bootstrapIncidentsResult.ok) ||
        (incidentsResult !== null && !incidentsResult.ok)
      ) {
        setSession(nextSession);
        setCredentialState(
          credentialResult.ok
            ? (credentialResult.payload as { data: CredentialState }).data
            : null,
        );
        setCredentialError(nextCredentialError);
        setIncidents([]);
        setIncidentsPaging(emptyIncidentsPaging);
        setExtensionProfiles([]);
        setReferencePackJob(null);
        setLandingNotice(options.landingNotice ?? null);
        setError(extensionsError ?? bootstrapIncidentsError ?? incidentsError);
        setAuthPrompt(defaultAuthPrompt);
        setAppBootstrapState(
          incidentsResult !== null &&
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
        setIncidentsPaging(emptyIncidentsPaging);
        setExtensionProfiles([]);
        setReferencePackJob(null);
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
      const nextIncidents =
        incidentsResult === null
          ? []
          : (incidentsResult.payload as IncidentListEnvelope).data.incidents;
      const nextIncidentsPaging =
        incidentsResult === null
          ? emptyIncidentsPaging
          : (incidentsResult.payload as IncidentListEnvelope).meta.paging;
      const nextBootstrapIncidents =
        bootstrapIncidentsResult === null
          ? []
          : (bootstrapIncidentsResult.payload as IncidentListEnvelope).data
              .incidents;
      const nextBootstrapPaging =
        bootstrapIncidentsResult === null
          ? emptyIncidentsPaging
          : (bootstrapIncidentsResult.payload as IncidentListEnvelope).meta
              .paging;
      const nextExtensionProfiles = (
        extensionsResult.payload as {
          data: { extensions: ExtensionProfileResource[] };
        }
      ).data.extensions;
      if (shouldLoadIncidentDirectory) {
        lastLoadedIncidentDirectoryScopeRef.current = directoryScope;
      }

      let nextRoute: RouteState | null = null;
      let nextLandingNotice = options.landingNotice ?? null;
      if (deploymentRouteDenied) {
        nextRoute = {
          incidentId: "",
          debugHarness: false,
          deploymentAdministration: false,
          manualIncidentDirectory: false,
        };
        nextLandingNotice =
          options.landingNotice ??
          "Deployment administration requires deployment admin access.";
      } else if (shouldLoadIncidentDirectory) {
        const suppressRootIncidentAutoOpen =
          effectiveRouteSnapshot.incidentId === "" &&
          suppressRootIncidentAutoOpenRef.current;
        const requestedIncidentStillVisible =
          effectiveRouteSnapshot.incidentId === "" ||
          nextBootstrapIncidents.some(
            (incident) =>
              incident.incident_id === effectiveRouteSnapshot.incidentId,
          );
        const rootIncidentID =
          effectiveRouteSnapshot.incidentId === "" &&
          !effectiveRouteSnapshot.debugHarness &&
          !effectiveRouteSnapshot.manualIncidentDirectory &&
          !suppressRootIncidentAutoOpen &&
          nextBootstrapIncidents.length === 1 &&
          !nextBootstrapPaging.has_more
            ? (nextBootstrapIncidents[0]?.incident_id ?? "")
            : "";
        nextRoute =
          rootIncidentID !== ""
            ? {
                incidentId: rootIncidentID,
                debugHarness: false,
                deploymentAdministration: false,
                manualIncidentDirectory: false,
              }
            : requestedIncidentStillVisible
              ? null
              : {
                  incidentId: "",
                  debugHarness: effectiveRouteSnapshot.debugHarness,
                  deploymentAdministration: false,
                  manualIncidentDirectory:
                    effectiveRouteSnapshot.manualIncidentDirectory,
                };
        nextLandingNotice =
          nextRoute !== null
            ? (options.landingNotice ?? defaultStaleIncidentMessage)
            : (options.landingNotice ?? null);
      }

      setSession(nextSession);
      setCredentialState(nextCredentialState);
      setCredentialError(nextCredentialError);
      setIncidents(nextIncidents);
      setIncidentsPaging(nextIncidentsPaging);
      setExtensionProfiles(nextExtensionProfiles);
      if (
        deploymentRouteDenied ||
        !extensionClaimed(nextExtensionProfiles, "reference_pack")
      ) {
        setReferencePackJob(null);
      }
      incidentDirectoryControlsTouchedRef.current = false;
      if (deploymentRouteDenied) {
        setActiveDeploymentPanel("deployment-users");
      }
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
        setIncidentsPaging(emptyIncidentsPaging);
        setExtensionProfiles([]);
        setReferencePackJob(null);
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
    const routeSnapshot = {
      debugHarness: route.debugHarness,
      deploymentAdministration: route.deploymentAdministration,
      incidentId: route.incidentId,
      manualIncidentDirectory: route.manualIncidentDirectory,
    };
    if (routeSnapshot.debugHarness) {
      activeRefreshRef.current.controller?.abort();
      return;
    }
    if (routeSnapshot.incidentId !== "" && sessionRef.current !== null) {
      return;
    }
    void refreshShell({
      routeSnapshot,
      landingNotice: null,
    });
  }, [
    route.debugHarness,
    route.deploymentAdministration,
    route.incidentId,
    route.manualIncidentDirectory,
    refreshShell,
  ]);

  useEffect(() => {
    if (
      sessionRef.current === null ||
      routeRef.current.incidentId !== "" ||
      routeRef.current.deploymentAdministration
    ) {
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
      if (!incidentDirectoryControlsTouchedRef.current) {
        return;
      }
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

  const handleLoadMoreIncidents = useCallback(async () => {
    const cursorToken = incidentsPaging.next_cursor;
    if (!incidentsPaging.has_more || cursorToken === null) {
      return;
    }
    incidentDirectoryControlsTouchedRef.current = false;
    setLandingRefreshState("loading");
    const result = await fetchJSON<IncidentListEnvelope>(
      incidentListURL({
        limit: incidentsPaging.limit,
        cursorToken,
        search: lastLoadedIncidentDirectoryScopeRef.current.search,
        statusFilter: lastLoadedIncidentDirectoryScopeRef.current.statusFilter,
      }),
    );
    const nextError = extractError(result.payload);
    if (!result.ok) {
      setError(nextError);
      setLandingNotice("Failed to load more incidents.");
      setAppBootstrapState(
        isSessionRequiredError(result.status, nextError)
          ? "revoked"
          : isForbiddenError(result.status, nextError)
            ? "forbidden"
            : "public_error_envelope",
      );
      if (isSessionRequiredError(result.status, nextError)) {
        setSession(null);
        setCredentialState(null);
        setCredentialError(null);
        setIncidents([]);
        setIncidentsPaging(emptyIncidentsPaging);
        setReferencePackJob(null);
        setAuthPrompt(defaultRevokedSessionMessage);
      }
      setLandingRefreshState("failed");
      return;
    }
    const envelope = result.payload as IncidentListEnvelope;
    setIncidents((current) => [...current, ...envelope.data.incidents]);
    setIncidentsPaging(envelope.meta.paging);
    setError(null);
    setLandingNotice(null);
    setAppBootstrapState("authenticated");
    setLandingRefreshState("idle");
  }, [
    incidentsPaging.has_more,
    incidentsPaging.limit,
    incidentsPaging.next_cursor,
  ]);

  const currentUserLabel = useMemo(() => {
    if (session === null) {
      return "Anonymous";
    }
    return `${session.display_name}${session.is_deployment_admin ? " · deployment admin" : ""}`;
  }, [session]);
  const landingStatusText =
    landingNotice ??
    (route.deploymentAdministration
      ? landingRefreshState === "loading"
        ? "Loading deployment administration."
        : landingRefreshState === "failed" || error !== null
          ? "Failed to load deployment administration."
          : "Deployment administration ready."
      : appBootstrapState === "forbidden"
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
  const availableDeploymentPanels = useMemo(() => {
    const panels: LandingAdminPanelDescriptor[] = [];
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
          token: "incident-import",
          label: "Incident import",
          description: "Create incident from bundle",
          group: "deployment",
        });
      }
    }
    return panels;
  }, [extensionProfiles, session?.is_deployment_admin]);
  useEffect(() => {
    if (
      !availableDeploymentPanels.some(
        (panel) => panel.token === activeDeploymentPanel,
      )
    ) {
      setActiveDeploymentPanel("deployment-users");
    }
  }, [activeDeploymentPanel, availableDeploymentPanels]);
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

  const openIncident = useCallback((incidentId: string) => {
    const nextRoute = {
      incidentId,
      debugHarness: false,
      deploymentAdministration: false,
      manualIncidentDirectory: false,
    };
    suppressRootIncidentAutoOpenRef.current = false;
    setLandingNotice(null);
    writeRouteState(nextRoute, "push");
    startTransition(() => {
      setRoute(nextRoute);
    });
  }, []);

  const navigateToIncidentDirectory = useCallback(() => {
    const nextRoute = {
      incidentId: "",
      debugHarness: false,
      deploymentAdministration: false,
      manualIncidentDirectory: true,
    };
    suppressRootIncidentAutoOpenRef.current = true;
    setLandingNotice(null);
    writeRouteState(nextRoute, "push");
    startTransition(() => {
      setRoute(nextRoute);
    });
  }, []);

  const navigateToDeploymentAdministration = useCallback(() => {
    if (!sessionRef.current?.is_deployment_admin) {
      const nextRoute = {
        incidentId: "",
        debugHarness: false,
        deploymentAdministration: false,
        manualIncidentDirectory: true,
      };
      suppressRootIncidentAutoOpenRef.current = false;
      setLandingNotice(
        "Deployment administration requires deployment admin access.",
      );
      writeRouteState(nextRoute, "replace");
      startTransition(() => {
        setRoute(nextRoute);
      });
      return;
    }
    const nextRoute = {
      incidentId: "",
      debugHarness: false,
      deploymentAdministration: true,
      manualIncidentDirectory: false,
    };
    suppressRootIncidentAutoOpenRef.current = false;
    setLandingNotice(null);
    setActiveDeploymentPanel("deployment-users");
    writeRouteState(nextRoute, "push");
    startTransition(() => {
      setRoute(nextRoute);
    });
  }, []);

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
        setReferencePackJob(null);
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
      debugHarness: false,
      deploymentAdministration: false,
      manualIncidentDirectory: true,
    };
    suppressRootIncidentAutoOpenRef.current = true;
    setLandingNotice("Returned to incident landing.");
    writeRouteState(nextRoute, "push");
    startTransition(() => {
      setRoute(nextRoute);
    });
  }, []);

  const renderAccountMenu = useCallback(
    (currentContext: AccountMenuContext, options: AccountMenuOptions = {}) => (
      <AccountApplicationMenu
        canOpenDeploymentAdministration={session?.is_deployment_admin ?? false}
        currentContext={currentContext}
        currentIncidentRole={options.currentIncidentRole}
        currentUserLabel={currentUserLabel}
        incidentControls={options.incidentControls}
        onOpenAccountSettings={setAccountSettingsPanel}
        onOpenDeploymentAdministration={navigateToDeploymentAdministration}
        onOpenIncidentDirectory={
          options.onOpenIncidentDirectory ?? navigateToIncidentDirectory
        }
        triggerTestId={options.triggerTestId}
      />
    ),
    [
      currentUserLabel,
      navigateToDeploymentAdministration,
      navigateToIncidentDirectory,
      session?.is_deployment_admin,
    ],
  );

  const renderAccountSettings = useCallback(() => {
    if (accountSettingsPanel === null || session === null) {
      return null;
    }
    const tabs: ReadonlyArray<{
      label: string;
      token: AccountSettingsPanelToken;
    }> = [
      { token: "account-profile", label: "Profile" },
      { token: "account-appearance", label: "Appearance" },
      { token: "account-security", label: "Security" },
    ];
    return (
      <div style={accountSettingsBackdropStyle}>
        <section
          aria-label="Account settings"
          role="dialog"
          style={accountSettingsDialogStyle}
        >
          <header style={accountSettingsHeaderStyle}>
            <div>
              <p style={accountSettingsEyebrowStyle}>Account settings</p>
              <h2 style={accountSettingsTitleStyle}>
                {tabs.find((tab) => tab.token === accountSettingsPanel)?.label}
              </h2>
            </div>
            <button
              style={accountSettingsCloseButtonStyle}
              type="button"
              onClick={() => {
                setAccountSettingsPanel(null);
              }}
            >
              Close
            </button>
          </header>
          <div style={accountSettingsTabsStyle} role="tablist">
            {tabs.map((tab) => (
              <button
                key={tab.token}
                aria-selected={accountSettingsPanel === tab.token}
                role="tab"
                style={
                  accountSettingsPanel === tab.token
                    ? accountSettingsTabSelectedStyle
                    : accountSettingsTabStyle
                }
                type="button"
                onClick={() => {
                  setAccountSettingsPanel(tab.token);
                }}
              >
                {tab.label}
              </button>
            ))}
          </div>
          <div style={accountSettingsPanelStyle}>
            {accountSettingsPanel === "account-profile" ? (
              <AccountProfilePanel onRefreshShell={refreshCurrentShell} />
            ) : null}
            {accountSettingsPanel === "account-appearance" ? (
              <AccountAppearancePanel />
            ) : null}
            {accountSettingsPanel === "account-security" ? (
              <Phase1AccountPanel
                credentialStateError={credentialError}
                onRefreshShell={refreshCurrentShell}
              />
            ) : null}
          </div>
        </section>
      </div>
    );
  }, [accountSettingsPanel, credentialError, refreshCurrentShell, session]);

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
              accountApplicationMenu={({
                currentIncidentRole,
                incidentControls,
              }) =>
                renderAccountMenu("workbook", {
                  currentIncidentRole,
                  incidentControls,
                  onOpenIncidentDirectory: handleReturnToLanding,
                  triggerTestId: phase1RouteTestId("workbook-current-user"),
                })
              }
              currentUserLabel={currentUserLabel}
              incidentId={route.incidentId}
              onIncidentAccessLost={handleIncidentAccessLost}
              onIncidentSnapshot={handleIncidentSnapshot}
            />
          </Suspense>
        </section>
        {renderAccountSettings()}
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
      {route.deploymentAdministration && session.is_deployment_admin ? (
        <LandingAdminShell
          accountMenu={renderAccountMenu("deployment-administration")}
          activePanel={activeDeploymentPanel}
          availablePanels={availableDeploymentPanels}
          currentUserLabel={currentUserLabel}
          onActivePanelChange={setActiveDeploymentPanel}
          statusText={landingStatusText}
        >
          <section
            id={landingAdminPanelTestId("deployment-users")}
            aria-labelledby={landingAdminMenuItemTestId("deployment-users")}
            data-testid={landingAdminPanelTestId("deployment-users")}
            hidden={activeDeploymentPanel !== "deployment-users"}
            style={landingAdminPanelRegionStyle}
          >
            <Phase1AdminPanel
              autoLoadUsers={activeDeploymentPanel === "deployment-users"}
              enterpriseAuthClaimed={extensionClaimed(
                extensionProfiles,
                "enterprise_authentication",
              )}
              onRefreshShell={refreshCurrentShell}
              session={session}
            />
          </section>
          <section
            id={landingAdminPanelTestId("administrative-audit")}
            aria-labelledby={landingAdminMenuItemTestId("administrative-audit")}
            data-testid={landingAdminPanelTestId("administrative-audit")}
            hidden={activeDeploymentPanel !== "administrative-audit"}
            style={landingAdminPanelRegionStyle}
          >
            <AdministrativeAuditPanel />
          </section>
          {extensionClaimed(extensionProfiles, "reference_pack") ? (
            <section
              id={landingAdminPanelTestId("reference-packs")}
              aria-labelledby={landingAdminMenuItemTestId("reference-packs")}
              data-testid={landingAdminPanelTestId("reference-packs")}
              hidden={activeDeploymentPanel !== "reference-packs"}
              style={landingAdminPanelRegionStyle}
            >
              <ReferencePackAdminPanel
                activeJob={referencePackJob}
                session={session}
                onJobChange={setReferencePackJob}
              />
            </section>
          ) : null}
          {extensionClaimed(extensionProfiles, "incident_portability") ? (
            <section
              id={landingAdminPanelTestId("incident-import")}
              aria-labelledby={landingAdminMenuItemTestId("incident-import")}
              data-testid={landingAdminPanelTestId("incident-import")}
              hidden={activeDeploymentPanel !== "incident-import"}
              style={landingAdminPanelRegionStyle}
            >
              <IncidentImportPanel onOpenImportedIncident={openIncident} />
            </section>
          ) : null}
        </LandingAdminShell>
      ) : (
        <IncidentDirectoryShell
          accountMenu={renderAccountMenu("incidents")}
          currentUserLabel={currentUserLabel}
          statusText={landingStatusText}
        >
          <section
            id={landingAdminPanelTestId("incidents")}
            aria-labelledby={landingAdminMenuItemTestId("incidents")}
            data-testid={landingAdminPanelTestId("incidents")}
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
              onLoadMore={handleLoadMoreIncidents}
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
        </IncidentDirectoryShell>
      )}
      {renderAccountSettings()}
    </main>
  );
}

const landingAdminPanelRegionStyle: CSSProperties = {
  minWidth: 0,
  minHeight: 0,
  padding: "var(--ct-spacing-md)",
};

const accountSettingsBackdropStyle: CSSProperties = {
  position: "fixed",
  inset: 0,
  zIndex: 40,
  display: "grid",
  placeItems: "center",
  padding: "var(--ct-spacing-lg)",
  background: "rgba(12, 16, 24, 0.42)",
};

const accountSettingsDialogStyle: CSSProperties = {
  width: "min(60rem, calc(100vw - 2rem))",
  maxHeight: "min(52rem, calc(100vh - 2rem))",
  display: "grid",
  gridTemplateRows: "auto auto minmax(0, 1fr)",
  minWidth: 0,
  overflow: "hidden",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  background: "var(--ct-colors-surface-1)",
  boxShadow: "var(--ct-elevation-panel)",
};

const accountSettingsHeaderStyle: CSSProperties = {
  display: "flex",
  justifyContent: "space-between",
  alignItems: "center",
  gap: "var(--ct-spacing-md)",
  padding: "var(--ct-spacing-md)",
  borderBottom: "var(--ct-border-hairline)",
};

const accountSettingsEyebrowStyle: CSSProperties = {
  margin: 0,
  fontSize: "0.72rem",
  letterSpacing: "0.14em",
  textTransform: "uppercase",
  color: "var(--ct-colors-accent)",
};

const accountSettingsTitleStyle: CSSProperties = {
  margin: "0.2rem 0 0",
  fontSize: "1.15rem",
};

const accountSettingsCloseButtonStyle: CSSProperties = {
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-ink)",
  padding: "0.48rem 0.7rem",
  fontWeight: 700,
  cursor: "pointer",
};

const accountSettingsTabsStyle: CSSProperties = {
  display: "flex",
  gap: "0.35rem",
  padding: "var(--ct-spacing-sm) var(--ct-spacing-md)",
  borderBottom: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
};

const accountSettingsTabStyle: CSSProperties = {
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  background: "transparent",
  color: "var(--ct-colors-ink-muted)",
  padding: "0.45rem 0.65rem",
  fontWeight: 700,
  cursor: "pointer",
};

const accountSettingsTabSelectedStyle: CSSProperties = {
  ...accountSettingsTabStyle,
  background: "var(--ct-colors-surface-1)",
  color: "var(--ct-colors-ink)",
};

const accountSettingsPanelStyle: CSSProperties = {
  minWidth: 0,
  minHeight: 0,
  overflow: "auto",
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
