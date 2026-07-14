import type {
  AccountPreferencesResource,
  CredentialState,
  EnterpriseAuthProvider,
  ExtensionProfileResource,
  SessionData,
} from "../app/api/appShellClient";
import { jsonResponse } from "./fetchMockTestSupport";

type FetchMock = {
  mockImplementation: (
    handler: (
      input: RequestInfo | URL,
      init?: RequestInit,
    ) => Response | Promise<Response>,
  ) => unknown;
};

type RouteRequest = {
  input: RequestInfo | URL;
  init: RequestInit | undefined;
  method: string;
  path: string;
  query: URLSearchParams;
  url: string;
};

type MaybeHandler<TValue extends object> =
  | TValue
  | Response
  | Promise<Response>
  | ((request: RouteRequest) => TValue | Response | Promise<Response>);

type ExtraRoute = {
  handler: (request: RouteRequest) => Response | Promise<Response>;
  method: string;
  url: string;
};

export type IncidentResource = {
  current_phase: string | null;
  description: string | null;
  closed_at?: string | null;
  incident_id: string;
  incident_key: string;
  incident_version: number;
  primary_external_case_ref: string | null;
  severity: string | null;
  status?: "active" | "closed";
  title: string;
  tlp: string | null;
  updated_at?: string;
};

type InstallLandingShellFetchOptions = {
  accountPreferences?: MaybeHandler<AccountPreferencesResource>;
  credentialState?: MaybeHandler<CredentialState>;
  enterpriseProviders?: MaybeHandler<{ providers: EnterpriseAuthProvider[] }>;
  extensions?: MaybeHandler<{ extensions: ExtensionProfileResource[] }>;
  extraRoutes?: ExtraRoute[];
  incidents?: MaybeHandler<IncidentResource[]>;
  onCreateIncident?: MaybeHandler<IncidentResource>;
  session: MaybeHandler<SessionData>;
};

export function installLandingShellFetch(
  fetchMock: FetchMock,
  options: InstallLandingShellFetchOptions,
) {
  fetchMock.mockImplementation(async (input, init) => {
    const request = routeRequest(input, init);
    const extraRoute = options.extraRoutes?.find(
      (route) =>
        route.url === request.url &&
        route.method.toUpperCase() === request.method,
    );
    if (extraRoute) {
      return extraRoute.handler(request);
    }

    if (request.url === "/api/v1/auth/session" && request.method === "GET") {
      return dataResponse(options.session, request);
    }
    if (request.url === "/api/v1/auth/providers" && request.method === "GET") {
      return dataResponse(
        options.enterpriseProviders ?? { providers: [] },
        request,
      );
    }
    if (request.path === "/api/v1/extensions" && request.method === "GET") {
      await waitForAbortWindow(request.init?.signal ?? undefined);
      const value = await resolveMaybeHandler(
        options.extensions ?? { extensions: [] },
        request,
      );
      if (value instanceof Response) {
        return value;
      }
      return jsonResponse({
        data: value,
        meta: { request_id: "request-test" },
      });
    }
    if (
      request.url === "/api/v1/account/preferences" &&
      request.method === "GET"
    ) {
      return dataResponse(
        options.accountPreferences ?? accountPreferencesResource(),
        request,
      );
    }
    if (
      request.url === "/api/v1/auth/credential-state" &&
      request.method === "GET"
    ) {
      return dataResponse(
        options.credentialState ?? credentialStateResource(),
        request,
      );
    }
    if (request.path === "/api/v1/incidents" && request.method === "GET") {
      const incidents = await resolveMaybeHandler(
        options.incidents ?? [],
        request,
      );
      if (incidents instanceof Response) {
        return incidents;
      }
      const limit = Number.parseInt(request.query.get("limit") ?? "100", 10);
      const boundedLimit = Number.isFinite(limit) && limit > 0 ? limit : 100;
      const cursorToken = request.query.get("cursor_token");
      const cursorOffset =
        cursorToken?.startsWith("cursor-") === true
          ? Number.parseInt(cursorToken.slice("cursor-".length), 10)
          : 0;
      const boundedCursorOffset =
        Number.isFinite(cursorOffset) && cursorOffset > 0 ? cursorOffset : 0;
      const status = request.query.get("status");
      const search = (request.query.get("search") ?? "").trim().toLowerCase();
      const filteredIncidents = incidents.filter((incident) => {
        if (status === "active" && (incident.status ?? "active") !== "active") {
          return false;
        }
        if (status === "closed" && incident.status !== "closed") {
          return false;
        }
        if (search === "") {
          return true;
        }
        return [
          incident.incident_key,
          incident.title,
          incident.severity,
          incident.tlp,
          incident.current_phase,
          incident.primary_external_case_ref,
        ].some((value) => value?.toLowerCase().includes(search) ?? false);
      });
      const visibleIncidents = filteredIncidents.slice(
        boundedCursorOffset,
        boundedCursorOffset + boundedLimit,
      );
      const nextOffset = boundedCursorOffset + visibleIncidents.length;
      const hasMore = filteredIncidents.length > nextOffset;
      return jsonResponse({
        data: {
          incidents: visibleIncidents,
        },
        meta: {
          paging: {
            limit: boundedLimit,
            has_more: hasMore,
            next_cursor: hasMore ? `cursor-${nextOffset}` : null,
          },
        },
      });
    }
    if (
      request.url === "/api/v1/incidents" &&
      request.method === "POST" &&
      options.onCreateIncident
    ) {
      return dataResponse(options.onCreateIncident, request);
    }
    throw new Error(`unexpected fetch: ${request.method} ${request.url}`);
  });
}

export function sessionResource(overrides?: Partial<SessionData>): SessionData {
  return {
    user_id: "user-1",
    display_name: "Operator",
    provider_type: "local",
    mfa_state: "not_required",
    is_deployment_admin: false,
    authenticated_at: "2026-04-20T12:00:00Z",
    idle_expires_at: "2026-04-20T12:30:00Z",
    absolute_expires_at: "2026-04-20T20:00:00Z",
    session_expires_at: "2026-04-20T12:30:00Z",
    memberships: [],
    ...overrides,
  };
}

export function credentialStateResource(
  overrides?: Partial<CredentialState>,
): CredentialState {
  const baseTotp = {
    state: "not_enrolled" as const,
    enrolled_at: null,
    pending_expires_at: null,
  };
  return {
    user_id: "user-1",
    auth_kind: "local",
    recovery_model: "admin_assisted",
    password_changed_at: "2026-04-20T12:00:00Z",
    ...overrides,
    totp: {
      ...baseTotp,
      ...(overrides?.totp ?? {}),
    },
  };
}

function accountPreferencesResource(
  overrides?: Partial<AccountPreferencesResource>,
): AccountPreferencesResource {
  return {
    user_id: "user-1",
    density_mode: null,
    preferences_version: 1,
    created_at: "2026-04-20T12:00:00Z",
    updated_at: "2026-04-20T12:00:00Z",
    ...overrides,
  };
}

export function incidentResource(
  incidentId: string,
  incidentKey: string,
  title: string,
  overrides?: Partial<IncidentResource>,
): IncidentResource {
  return {
    incident_id: incidentId,
    incident_key: incidentKey,
    title,
    description: null,
    severity: null,
    tlp: null,
    current_phase: null,
    primary_external_case_ref: null,
    incident_version: 1,
    ...overrides,
  };
}

function routeRequest(
  input: RequestInfo | URL,
  init: RequestInit | undefined,
): RouteRequest {
  const rawURL = String(input);
  const parsedURL = new URL(rawURL, "http://cartulary.test");
  return {
    input,
    init,
    method: (init?.method ?? "GET").toUpperCase(),
    path: parsedURL.pathname,
    query: parsedURL.searchParams,
    url: rawURL,
  };
}

async function dataResponse<TValue extends object>(
  source: MaybeHandler<TValue>,
  request: RouteRequest,
) {
  await waitForAbortWindow(request.init?.signal ?? undefined);
  const value = await resolveMaybeHandler(source, request);
  if (value instanceof Response) {
    return value;
  }
  return jsonResponse({ data: value });
}

function waitForAbortWindow(signal: AbortSignal | undefined) {
  return new Promise<void>((resolve, reject) => {
    if (signal?.aborted) {
      reject(new DOMException("Aborted", "AbortError"));
      return;
    }
    let timeout: ReturnType<typeof setTimeout> | null = null;
    const cleanup = () => {
      if (timeout !== null) {
        clearTimeout(timeout);
      }
      signal?.removeEventListener("abort", abort);
    };
    const abort = () => {
      cleanup();
      reject(new DOMException("Aborted", "AbortError"));
    };
    timeout = setTimeout(() => {
      cleanup();
      resolve();
    }, 0);
    signal?.addEventListener("abort", abort, { once: true });
  });
}

function resolveMaybeHandler<TValue extends object>(
  source: MaybeHandler<TValue>,
  request: RouteRequest,
) {
  if (typeof source === "function") {
    return source(request);
  }
  return source;
}
