import type {
  CreateDeploymentUserRequest,
  CreateIncidentRequest,
  CreateIncidentResponse,
  GetCurrentAccountPreferencesResponse,
  GetCurrentAccountProfileResponse,
  GetCurrentSessionResponse,
  GetDeploymentUserResponse,
  ListDeploymentExtensionsResponse,
  ListDeploymentUsersResponse,
  PatchCurrentAccountProfileRequest,
  PatchCurrentAccountProfileResponse,
  PatchDeploymentUserRequest,
  PutCurrentAccountPreferencesRequest,
  PutCurrentAccountPreferencesResponse,
  ResetDeploymentUserPasswordRequest,
  ResetDeploymentUserTOTPRequest,
  RevokeAllDeploymentUserSessionsRequest,
} from "@cartulary/protocol-ts/http";
import {
  type APIResult,
  clientTxnID,
  fetchHTTPOperation,
  fetchJSON,
} from "../../services/browserApi";
import type { AuthorizationRecoveryPort } from "../../shared/authorizationRecovery";

export type SessionMembership =
  GetCurrentSessionResponse["data"]["memberships"][number];
export type SessionProviderType =
  GetCurrentSessionResponse["data"]["provider_type"];
export type SessionMFAState = GetCurrentSessionResponse["data"]["mfa_state"];
export type CredentialAuthKind = "local";
export type CredentialRecoveryModel = "admin_assisted";
export type CredentialTOTPState = "not_enrolled" | "pending" | "active";

export type SessionData = GetCurrentSessionResponse["data"];

export type CredentialState = {
  user_id: string;
  auth_kind: CredentialAuthKind;
  recovery_model: CredentialRecoveryModel;
  password_changed_at: string | null;
  totp: {
    state: CredentialTOTPState;
    enrolled_at?: string | null;
    pending_expires_at?: string | null;
  };
};

export type UserResource = GetDeploymentUserResponse["data"];

export type AccountProfileResource = GetCurrentAccountProfileResponse["data"];
export type AccountPreferencesResource =
  GetCurrentAccountPreferencesResponse["data"];
export type DensityMode = Exclude<
  AccountPreferencesResource["density_mode"],
  null
>;

export type ExtensionProfileResource =
  ListDeploymentExtensionsResponse["data"]["extensions"][number];

export type PagingMeta = NonNullable<
  ListDeploymentUsersResponse["meta"]["paging"]
>;

export type UserListEnvelope = ListDeploymentUsersResponse;

export type TotpAuthMode = "bootstrap" | "session";

export type AuthBindingSummary = UserResource["auth_bindings"][number];

export type EnterpriseAuthProvider = {
  provider_key: string;
  provider_type: "oidc" | "saml";
  display_name: string;
};

type EnterpriseAuthBeginResponse = {
  provider_key: string;
  provider_type: "oidc" | "saml";
  redirect_url: string;
  expires_at: string;
};

type DataEnvelope<T> = {
  data: T;
};

type TotpBeginResponse = {
  enrollment_id: string;
  totp_setup: {
    secret_base32: string;
  };
};

type ShellGetOptions = {
  apiBase?: string | undefined;
  signal?: AbortSignal;
};

function apiPath(base: string | undefined, path: string): string {
  const trimmedBase = (base ?? "").trim();
  if (trimmedBase === "") {
    return path;
  }
  return `${trimmedBase.replace(/\/$/, "")}${path}`;
}

function resolveShellGetOptions(
  apiBaseOrOptions?: string | ShellGetOptions,
): ShellGetOptions {
  if (
    typeof apiBaseOrOptions === "undefined" ||
    typeof apiBaseOrOptions === "string"
  ) {
    return {
      apiBase: apiBaseOrOptions,
    };
  }
  return apiBaseOrOptions;
}

function secondFactorPayload(code: string) {
  if (code.trim() === "") {
    return undefined;
  }
  return {
    kind: "totp",
    assertion: {
      code,
    },
  };
}

export async function createIncident(options: {
  request: Readonly<CreateIncidentRequest>;
  signal: AbortSignal;
}) {
  const {
    client_txn_id,
    incident_key,
    title,
    description,
    severity,
    tlp,
    current_phase,
    primary_external_case_ref,
  } = options.request;
  const request = {
    client_txn_id,
    incident_key,
    title,
    ...(description === undefined ? {} : { description }),
    ...(severity === undefined ? {} : { severity }),
    ...(tlp === undefined ? {} : { tlp }),
    ...(current_phase === undefined ? {} : { current_phase }),
    ...(primary_external_case_ref === undefined
      ? {}
      : { primary_external_case_ref }),
  } satisfies CreateIncidentRequest;
  const result = await fetchHTTPOperation<CreateIncidentResponse>({
    operationID: "createIncident",
    init: {
      method: "POST",
      body: JSON.stringify(request),
      signal: options.signal,
    },
  });
  if (result.ok && result.status !== 200 && result.status !== 201) {
    return {
      ok: false as const,
      status: 502,
      payload: {
        error: {
          code: "invalid_public_contract_response",
          status: 502,
          retryable: true,
        },
      },
    };
  }
  return result;
}

function bootstrapAuthorizationHeader(options: {
  authMode: TotpAuthMode;
  bootstrapToken?: string | undefined;
}): HeadersInit | undefined {
  if (options.authMode !== "bootstrap") {
    return undefined;
  }
  const token = options.bootstrapToken?.trim() ?? "";
  if (token === "") {
    return undefined;
  }
  return {
    Authorization: `Bearer ${token}`,
  };
}

function totpEnrollmentRequestInit(options: {
  authMode: TotpAuthMode;
  bootstrapToken?: string | undefined;
  body: Record<string, unknown>;
}): RequestInit {
  const headers = bootstrapAuthorizationHeader(options);
  return {
    method: "POST",
    credentials: options.authMode === "bootstrap" ? "omit" : "include",
    body: JSON.stringify(options.body),
    ...(headers === undefined ? {} : { headers }),
  };
}

export function loadSession(
  apiBase?: string,
): Promise<APIResult<DataEnvelope<SessionData>>>;
export function loadSession(
  options?: ShellGetOptions,
): Promise<APIResult<DataEnvelope<SessionData>>>;
export function loadSession(apiBaseOrOptions?: string | ShellGetOptions) {
  const options = resolveShellGetOptions(apiBaseOrOptions);
  const requestInit =
    typeof options.signal === "undefined"
      ? undefined
      : {
          signal: options.signal,
        };
  return fetchHTTPOperation<GetCurrentSessionResponse>({
    apiBase: options.apiBase,
    init: requestInit,
    operationID: "getCurrentSession",
  });
}

export function createAppAuthorizationRecoveryPort(
  options: {
    readonly loadCurrentSession?:
      | ((signal: AbortSignal) => Promise<APIResult<DataEnvelope<SessionData>>>)
      | undefined;
    readonly onSessionRecovered?: ((session: SessionData) => void) | undefined;
  } = {},
): AuthorizationRecoveryPort {
  return {
    async recover({ incidentId, signal }) {
      try {
        const result = await (
          options.loadCurrentSession ??
          ((nextSignal: AbortSignal) => loadSession({ signal: nextSignal }))
        )(signal);
        if (!result.ok) return { kind: "unavailable" };
        const session = (result.payload as DataEnvelope<SessionData>).data;
        const membership = session.memberships.find(
          (entry) => entry.incident_id === incidentId,
        );
        if (membership === undefined) return { kind: "access_lost" };
        options.onSessionRecovered?.(session);
        return {
          kind: "authorized",
          role: membership.role,
          userId: session.user_id,
        };
      } catch {
        return { kind: "unavailable" };
      }
    },
  };
}

export function loadCredentialState(
  apiBase?: string,
): Promise<APIResult<DataEnvelope<CredentialState>>>;
export function loadCredentialState(
  options?: ShellGetOptions,
): Promise<APIResult<DataEnvelope<CredentialState>>>;
export function loadCredentialState(
  apiBaseOrOptions?: string | ShellGetOptions,
) {
  const options = resolveShellGetOptions(apiBaseOrOptions);
  const requestInit =
    typeof options.signal === "undefined"
      ? undefined
      : {
          signal: options.signal,
        };
  return fetchJSON<DataEnvelope<CredentialState>>(
    apiPath(options.apiBase, "/api/v1/auth/credential-state"),
    requestInit,
  );
}

export function loadAccountProfile(options?: ShellGetOptions) {
  return fetchHTTPOperation<GetCurrentAccountProfileResponse>({
    apiBase: options?.apiBase,
    init:
      typeof options?.signal === "undefined"
        ? undefined
        : {
            signal: options.signal,
          },
    operationID: "getCurrentAccountProfile",
  });
}

export function patchAccountProfile(options: {
  apiBase?: string | undefined;
  baseUserVersion: number;
  clientTxnId?: string;
  displayName: string;
}) {
  const request = {
    base_user_version: options.baseUserVersion,
    client_txn_id: options.clientTxnId ?? clientTxnID("account-profile-patch"),
    display_name: options.displayName,
  } satisfies PatchCurrentAccountProfileRequest;
  return fetchHTTPOperation<PatchCurrentAccountProfileResponse>({
    apiBase: options.apiBase,
    init: {
      method: "PATCH",
      body: JSON.stringify(request),
    },
    operationID: "patchCurrentAccountProfile",
  });
}

export function loadAccountPreferences(options?: ShellGetOptions) {
  return fetchHTTPOperation<GetCurrentAccountPreferencesResponse>({
    apiBase: options?.apiBase,
    init:
      typeof options?.signal === "undefined"
        ? undefined
        : {
            signal: options.signal,
          },
    operationID: "getCurrentAccountPreferences",
  });
}

export function putAccountPreferences(options: {
  apiBase?: string | undefined;
  basePreferencesVersion: number;
  clientTxnId?: string;
  densityMode: DensityMode | null;
}) {
  const request = {
    base_preferences_version: options.basePreferencesVersion,
    client_txn_id:
      options.clientTxnId ?? clientTxnID("account-preferences-put"),
    density_mode: options.densityMode,
  } satisfies PutCurrentAccountPreferencesRequest;
  return fetchHTTPOperation<PutCurrentAccountPreferencesResponse>({
    apiBase: options.apiBase,
    init: {
      method: "PUT",
      body: JSON.stringify(request),
    },
    operationID: "putCurrentAccountPreferences",
  });
}

export function loadExtensions(options?: ShellGetOptions) {
  return fetchHTTPOperation<ListDeploymentExtensionsResponse>({
    apiBase: options?.apiBase,
    init:
      typeof options?.signal === "undefined"
        ? undefined
        : {
            signal: options.signal,
          },
    operationID: "listDeploymentExtensions",
  });
}

export function loginLocal(options: {
  apiBase?: string | undefined;
  password: string;
  secondFactorCode?: string;
  username: string;
}) {
  const secondFactor = secondFactorPayload(options.secondFactorCode ?? "");
  return fetchJSON<DataEnvelope<SessionData>>(
    apiPath(options.apiBase, "/api/v1/auth/login"),
    {
      method: "POST",
      body: JSON.stringify({
        username: options.username,
        password: options.password,
        ...(secondFactor === undefined
          ? {}
          : {
              second_factor: secondFactor,
            }),
      }),
    },
  );
}

export function listEnterpriseAuthProviders(options?: ShellGetOptions) {
  return fetchJSON<DataEnvelope<{ providers: EnterpriseAuthProvider[] }>>(
    apiPath(options?.apiBase, "/api/v1/auth/providers"),
    typeof options?.signal === "undefined"
      ? undefined
      : {
          signal: options.signal,
        },
  );
}

export function beginEnterpriseAuth(options: {
  apiBase?: string | undefined;
  providerKey: string;
  returnTo?: string | undefined;
}) {
  return fetchJSON<DataEnvelope<EnterpriseAuthBeginResponse>>(
    apiPath(
      options.apiBase,
      `/api/v1/auth/providers/${encodeURIComponent(options.providerKey)}/begin`,
    ),
    {
      method: "POST",
      body: JSON.stringify({
        return_to: options.returnTo ?? "/",
      }),
    },
  );
}

export function logoutCurrentSession(apiBase?: string) {
  return fetchJSON(apiPath(apiBase, "/api/v1/auth/logout"), {
    method: "POST",
    body: JSON.stringify({}),
  });
}

export function beginTotpEnrollment(options: {
  apiBase?: string | undefined;
  authMode: TotpAuthMode;
  bootstrapToken?: string;
  clientTxnId?: string;
  currentFactorCode?: string;
  currentPassword?: string;
}) {
  const requestInit = totpEnrollmentRequestInit({
    authMode: options.authMode,
    bootstrapToken: options.bootstrapToken,
    body: {
      client_txn_id:
        options.clientTxnId ?? clientTxnID("authentication-ui-totp-begin"),
      ...(options.authMode === "session"
        ? {
            current_password: options.currentPassword ?? "",
            second_factor: secondFactorPayload(options.currentFactorCode ?? ""),
          }
        : {}),
    },
  });
  return fetchJSON<DataEnvelope<TotpBeginResponse>>(
    apiPath(options.apiBase, "/api/v1/auth/mfa/totp/begin"),
    requestInit,
  );
}

export function completeTotpEnrollment(options: {
  apiBase?: string | undefined;
  authMode: TotpAuthMode;
  bootstrapToken?: string;
  clientTxnId?: string;
  code: string;
  enrollmentId: string;
}) {
  const requestInit = totpEnrollmentRequestInit({
    authMode: options.authMode,
    bootstrapToken: options.bootstrapToken,
    body: {
      client_txn_id:
        options.clientTxnId ?? clientTxnID("authentication-ui-totp-complete"),
      enrollment_id: options.enrollmentId,
      code: options.code,
    },
  });
  return fetchJSON(
    apiPath(options.apiBase, "/api/v1/auth/mfa/totp/complete"),
    requestInit,
  );
}

export function changePassword(options: {
  apiBase?: string | undefined;
  clientTxnId?: string;
  currentPassword: string;
  newPassword: string;
  secondFactorCode?: string;
}) {
  return fetchJSON(apiPath(options.apiBase, "/api/v1/auth/password/change"), {
    method: "POST",
    body: JSON.stringify({
      client_txn_id:
        options.clientTxnId ?? clientTxnID("authentication-ui-password-change"),
      current_password: options.currentPassword,
      new_password: options.newPassword,
      second_factor: secondFactorPayload(options.secondFactorCode ?? ""),
    }),
  });
}

export function createLocalUser(options: {
  apiBase?: string | undefined;
  clientTxnId?: string;
  displayName: string;
  email: string;
  initialPassword: string;
  isDeploymentAdmin: boolean;
  mfaRequired: boolean;
}) {
  const request = {
    client_txn_id:
      options.clientTxnId ?? clientTxnID("authentication-ui-user-create"),
    auth_kind: "local",
    email: options.email,
    display_name: options.displayName,
    initial_password: options.initialPassword,
    mfa_required: options.mfaRequired,
    is_deployment_admin: options.isDeploymentAdmin,
  } satisfies CreateDeploymentUserRequest;
  return fetchHTTPOperation<DataEnvelope<UserResource>>({
    apiBase: options.apiBase,
    init: {
      method: "POST",
      body: JSON.stringify(request),
    },
    operationID: "createDeploymentUser",
  });
}

export function listUsers(options?: {
  apiBase?: string | undefined;
  cursorToken?: string | null;
  isActive?: boolean | null | undefined;
  isDeploymentAdmin?: boolean | null | undefined;
  limit?: number | undefined;
  search?: string | undefined;
  signal?: AbortSignal | undefined;
}) {
  const query: Record<string, string | number> = {
    limit: options?.limit ?? 100,
  };
  const cursorToken = options?.cursorToken?.trim() ?? "";
  if (cursorToken !== "") {
    query.cursor_token = cursorToken;
  }
  const search = options?.search?.trim() ?? "";
  if (search !== "") {
    query.search = search;
  }
  if (typeof options?.isActive === "boolean") {
    query.is_active = String(options.isActive);
  }
  if (typeof options?.isDeploymentAdmin === "boolean") {
    query.is_deployment_admin = String(options.isDeploymentAdmin);
  }
  const requestInit =
    typeof options?.signal === "undefined"
      ? undefined
      : {
          signal: options.signal,
        };
  return fetchHTTPOperation<UserListEnvelope>({
    apiBase: options?.apiBase,
    init: requestInit,
    operationID: "listDeploymentUsers",
    query,
  });
}

export function loadUser(options: {
  apiBase?: string | undefined;
  userId: string;
}) {
  return fetchHTTPOperation<DataEnvelope<UserResource>>({
    apiBase: options.apiBase,
    operationID: "getDeploymentUser",
    pathParameters: { user_id: options.userId },
  });
}

export function patchLocalUser(options: {
  apiBase?: string | undefined;
  baseUserVersion: number;
  displayName: string;
  email: string;
  isActive: boolean;
  isDeploymentAdmin: boolean;
  mfaRequired: boolean;
  userId: string;
}) {
  const request = {
    base_user_version: options.baseUserVersion,
    email: options.email,
    display_name: options.displayName,
    mfa_required: options.mfaRequired,
    is_active: options.isActive,
    is_deployment_admin: options.isDeploymentAdmin,
  } satisfies PatchDeploymentUserRequest;
  return fetchHTTPOperation<DataEnvelope<UserResource>>({
    apiBase: options.apiBase,
    init: {
      method: "PATCH",
      body: JSON.stringify(request),
    },
    operationID: "patchDeploymentUser",
    pathParameters: { user_id: options.userId },
  });
}

export function createEnterpriseAuthBinding(options: {
  apiBase?: string | undefined;
  baseUserVersion: number;
  clientTxnId?: string;
  providerKey: string;
  providerSubject: string;
  reason: string;
  userId: string;
}) {
  return fetchJSON<DataEnvelope<UserResource>>(
    apiPath(options.apiBase, `/api/v1/users/${options.userId}/auth-bindings`),
    {
      method: "POST",
      body: JSON.stringify({
        base_user_version: options.baseUserVersion,
        client_txn_id:
          options.clientTxnId ??
          clientTxnID("authentication-ui-auth-binding-create"),
        provider_key: options.providerKey,
        provider_subject: options.providerSubject,
        reason: options.reason,
      }),
    },
  );
}

export function rotateEnterpriseAuthBinding(options: {
  apiBase?: string | undefined;
  authBindingId: string;
  baseUserVersion: number;
  clientTxnId?: string;
  newProviderSubject: string;
  reason: string;
  userId: string;
}) {
  return fetchJSON<DataEnvelope<UserResource>>(
    apiPath(
      options.apiBase,
      `/api/v1/users/${options.userId}/auth-bindings/${options.authBindingId}/rotate`,
    ),
    {
      method: "POST",
      body: JSON.stringify({
        base_user_version: options.baseUserVersion,
        client_txn_id:
          options.clientTxnId ??
          clientTxnID("authentication-ui-auth-binding-rotate"),
        new_provider_subject: options.newProviderSubject,
        reason: options.reason,
      }),
    },
  );
}

export function retireEnterpriseAuthBinding(options: {
  apiBase?: string | undefined;
  authBindingId: string;
  baseUserVersion: number;
  clientTxnId?: string;
  reason: string;
  userId: string;
}) {
  return fetchJSON<DataEnvelope<UserResource>>(
    apiPath(
      options.apiBase,
      `/api/v1/users/${options.userId}/auth-bindings/${options.authBindingId}`,
    ),
    {
      method: "DELETE",
      body: JSON.stringify({
        base_user_version: options.baseUserVersion,
        client_txn_id:
          options.clientTxnId ??
          clientTxnID("authentication-ui-auth-binding-retire"),
        reason: options.reason,
      }),
    },
  );
}

export function adminResetPassword(options: {
  apiBase?: string | undefined;
  baseUserVersion: number;
  clientTxnId?: string;
  newPassword: string;
  reason: string;
  userId: string;
}) {
  const request = {
    base_user_version: options.baseUserVersion,
    client_txn_id:
      options.clientTxnId ?? clientTxnID("authentication-ui-password-reset"),
    new_password: options.newPassword,
    reason: options.reason,
  } satisfies ResetDeploymentUserPasswordRequest;
  return fetchHTTPOperation<DataEnvelope<UserResource>>({
    apiBase: options.apiBase,
    init: {
      method: "POST",
      body: JSON.stringify(request),
    },
    operationID: "resetDeploymentUserPassword",
    pathParameters: { user_id: options.userId },
  });
}

export function adminResetTotp(options: {
  apiBase?: string | undefined;
  baseUserVersion: number;
  clientTxnId?: string;
  reason: string;
  userId: string;
}) {
  const request = {
    base_user_version: options.baseUserVersion,
    client_txn_id:
      options.clientTxnId ?? clientTxnID("authentication-ui-totp-reset"),
    reason: options.reason,
  } satisfies ResetDeploymentUserTOTPRequest;
  return fetchHTTPOperation<DataEnvelope<UserResource>>({
    apiBase: options.apiBase,
    init: {
      method: "POST",
      body: JSON.stringify(request),
    },
    operationID: "resetDeploymentUserTOTP",
    pathParameters: { user_id: options.userId },
  });
}

export function adminRevokeAllSessions(options: {
  apiBase?: string | undefined;
  clientTxnId?: string;
  reason: string;
  userId: string;
}) {
  const request = {
    client_txn_id:
      options.clientTxnId ?? clientTxnID("authentication-ui-revoke-all"),
    reason: options.reason,
  } satisfies RevokeAllDeploymentUserSessionsRequest;
  return fetchHTTPOperation({
    apiBase: options.apiBase,
    init: {
      method: "POST",
      body: JSON.stringify(request),
    },
    operationID: "revokeAllDeploymentUserSessions",
    pathParameters: { user_id: options.userId },
  });
}
