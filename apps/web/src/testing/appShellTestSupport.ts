import type {
  CredentialState,
  EnterpriseAuthProvider,
  SessionData,
} from "../app/phase1Client";
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
  incident_id: string;
  incident_key: string;
  incident_version: number;
  primary_external_case_ref: string | null;
  severity: string | null;
  title: string;
  tlp: string | null;
};

type InstallLandingShellFetchOptions = {
  credentialState?: MaybeHandler<CredentialState>;
  enterpriseProviders?: MaybeHandler<{ providers: EnterpriseAuthProvider[] }>;
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
    if (
      request.url === "/api/v1/auth/credential-state" &&
      request.method === "GET"
    ) {
      return dataResponse(
        options.credentialState ?? credentialStateResource(),
        request,
      );
    }
    if (request.url === "/api/v1/incidents" && request.method === "GET") {
      const incidents = await resolveMaybeHandler(
        options.incidents ?? [],
        request,
      );
      if (incidents instanceof Response) {
        return incidents;
      }
      return jsonResponse({
        data: {
          incidents,
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
  return {
    input,
    init,
    method: (init?.method ?? "GET").toUpperCase(),
    url: String(input),
  };
}

async function dataResponse<TValue extends object>(
  source: MaybeHandler<TValue>,
  request: RouteRequest,
) {
  const value = await resolveMaybeHandler(source, request);
  if (value instanceof Response) {
    return value;
  }
  return jsonResponse({ data: value });
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
