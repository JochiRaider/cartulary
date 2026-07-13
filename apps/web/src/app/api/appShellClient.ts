import type {
  AccountPreferencesResource,
  AccountProfileResource,
  DensityMode,
} from "@cartulary/protocol-ts";
import {
  type APIResult,
  clientTxnID,
  fetchJSON,
} from "../../services/browserApi";

export type SessionMembership = {
  incident_id: string;
  role: string;
};

export type SessionProviderType = "local" | "oidc" | "saml";
export type SessionMFAState = "not_required" | "satisfied";
export type CredentialAuthKind = "local";
export type CredentialRecoveryModel = "admin_assisted";
export type CredentialTOTPState = "not_enrolled" | "pending" | "active";

export type SessionData = {
  user_id: string;
  display_name: string;
  provider_type: SessionProviderType;
  mfa_state: SessionMFAState;
  is_deployment_admin: boolean;
  authenticated_at: string;
  idle_expires_at: string;
  absolute_expires_at: string;
  session_expires_at: string;
  memberships: SessionMembership[];
};

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

export type UserResource = {
  user_id: string;
  email: string;
  display_name: string;
  user_version: number;
  is_active: boolean;
  mfa_required: boolean;
  is_deployment_admin: boolean;
  auth_bindings?: AuthBindingSummary[];
  created_at?: string;
  updated_at?: string;
  updated_by_user_id?: string | null;
  last_login_at?: string | null;
};

export type { AccountPreferencesResource, AccountProfileResource, DensityMode };

export type ExtensionProfileResource = {
  profile_id: string;
  claimed: boolean;
  route_families: string[];
};

export type PagingMeta = {
  limit: number;
  has_more: boolean;
  next_cursor: string | null;
};

export type UserListEnvelope = {
  data: {
    users: UserResource[];
  };
  meta: {
    paging: PagingMeta;
    request_id: string;
  };
};

export type TotpAuthMode = "bootstrap" | "session";

export type AuthBindingSummary =
  | {
      provider_type: "local";
      provider_key: "local";
      username: string;
      created_at: string;
    }
  | {
      auth_binding_id: string;
      provider_type: "oidc" | "saml";
      provider_key: string;
      provider_subject: string;
      created_at: string;
      last_auth_at: string | null;
    };

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
  return fetchJSON<DataEnvelope<SessionData>>(
    apiPath(options.apiBase, "/api/v1/auth/session"),
    requestInit,
  );
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
  return fetchJSON<DataEnvelope<AccountProfileResource>>(
    apiPath(options?.apiBase, "/api/v1/account/profile"),
    typeof options?.signal === "undefined"
      ? undefined
      : {
          signal: options.signal,
        },
  );
}

export function patchAccountProfile(options: {
  apiBase?: string | undefined;
  baseUserVersion: number;
  clientTxnId?: string;
  displayName: string;
}) {
  return fetchJSON<DataEnvelope<AccountProfileResource>>(
    apiPath(options.apiBase, "/api/v1/account/profile"),
    {
      method: "PATCH",
      body: JSON.stringify({
        base_user_version: options.baseUserVersion,
        client_txn_id:
          options.clientTxnId ?? clientTxnID("account-profile-patch"),
        display_name: options.displayName,
      }),
    },
  );
}

export function loadAccountPreferences(options?: ShellGetOptions) {
  return fetchJSON<DataEnvelope<AccountPreferencesResource>>(
    apiPath(options?.apiBase, "/api/v1/account/preferences"),
    typeof options?.signal === "undefined"
      ? undefined
      : {
          signal: options.signal,
        },
  );
}

export function putAccountPreferences(options: {
  apiBase?: string | undefined;
  basePreferencesVersion: number;
  clientTxnId?: string;
  densityMode: DensityMode | null;
}) {
  return fetchJSON<DataEnvelope<AccountPreferencesResource>>(
    apiPath(options.apiBase, "/api/v1/account/preferences"),
    {
      method: "PUT",
      body: JSON.stringify({
        base_preferences_version: options.basePreferencesVersion,
        client_txn_id:
          options.clientTxnId ?? clientTxnID("account-preferences-put"),
        density_mode: options.densityMode,
      }),
    },
  );
}

export function loadExtensions(options?: ShellGetOptions) {
  return fetchJSON<DataEnvelope<{ extensions: ExtensionProfileResource[] }>>(
    apiPath(options?.apiBase, "/api/v1/extensions"),
    typeof options?.signal === "undefined"
      ? undefined
      : {
          signal: options.signal,
        },
  );
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
      client_txn_id: options.clientTxnId ?? clientTxnID("phase1-ui-totp-begin"),
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
        options.clientTxnId ?? clientTxnID("phase1-ui-totp-complete"),
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
        options.clientTxnId ?? clientTxnID("phase1-ui-password-change"),
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
  return fetchJSON<DataEnvelope<UserResource>>(
    apiPath(options.apiBase, "/api/v1/users"),
    {
      method: "POST",
      body: JSON.stringify({
        client_txn_id:
          options.clientTxnId ?? clientTxnID("phase1-ui-user-create"),
        auth_kind: "local",
        email: options.email,
        display_name: options.displayName,
        initial_password: options.initialPassword,
        mfa_required: options.mfaRequired,
        is_deployment_admin: options.isDeploymentAdmin,
      }),
    },
  );
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
  const params = new URLSearchParams();
  params.set("limit", String(options?.limit ?? 100));
  const cursorToken = options?.cursorToken?.trim() ?? "";
  if (cursorToken !== "") {
    params.set("cursor_token", cursorToken);
  }
  const search = options?.search?.trim() ?? "";
  if (search !== "") {
    params.set("search", search);
  }
  if (typeof options?.isActive === "boolean") {
    params.set("is_active", String(options.isActive));
  }
  if (typeof options?.isDeploymentAdmin === "boolean") {
    params.set("is_deployment_admin", String(options.isDeploymentAdmin));
  }
  const requestInit =
    typeof options?.signal === "undefined"
      ? undefined
      : {
          signal: options.signal,
        };
  return fetchJSON<UserListEnvelope>(
    apiPath(options?.apiBase, `/api/v1/users?${params.toString()}`),
    requestInit,
  );
}

export function loadUser(options: {
  apiBase?: string | undefined;
  userId: string;
}) {
  return fetchJSON<DataEnvelope<UserResource>>(
    apiPath(options.apiBase, `/api/v1/users/${options.userId}`),
  );
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
  return fetchJSON<DataEnvelope<UserResource>>(
    apiPath(options.apiBase, `/api/v1/users/${options.userId}`),
    {
      method: "PATCH",
      body: JSON.stringify({
        base_user_version: options.baseUserVersion,
        email: options.email,
        display_name: options.displayName,
        mfa_required: options.mfaRequired,
        is_active: options.isActive,
        is_deployment_admin: options.isDeploymentAdmin,
      }),
    },
  );
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
          options.clientTxnId ?? clientTxnID("phase1-ui-auth-binding-create"),
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
          options.clientTxnId ?? clientTxnID("phase1-ui-auth-binding-rotate"),
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
          options.clientTxnId ?? clientTxnID("phase1-ui-auth-binding-retire"),
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
  return fetchJSON<DataEnvelope<UserResource>>(
    apiPath(options.apiBase, `/api/v1/users/${options.userId}/password/reset`),
    {
      method: "POST",
      body: JSON.stringify({
        base_user_version: options.baseUserVersion,
        client_txn_id:
          options.clientTxnId ?? clientTxnID("phase1-ui-password-reset"),
        new_password: options.newPassword,
        reason: options.reason,
      }),
    },
  );
}

export function adminResetTotp(options: {
  apiBase?: string | undefined;
  baseUserVersion: number;
  clientTxnId?: string;
  reason: string;
  userId: string;
}) {
  return fetchJSON<DataEnvelope<UserResource>>(
    apiPath(options.apiBase, `/api/v1/users/${options.userId}/mfa/totp/reset`),
    {
      method: "POST",
      body: JSON.stringify({
        base_user_version: options.baseUserVersion,
        client_txn_id:
          options.clientTxnId ?? clientTxnID("phase1-ui-totp-reset"),
        reason: options.reason,
      }),
    },
  );
}

export function adminRevokeAllSessions(options: {
  apiBase?: string | undefined;
  clientTxnId?: string;
  reason: string;
  userId: string;
}) {
  return fetchJSON(
    apiPath(
      options.apiBase,
      `/api/v1/users/${options.userId}/sessions/revoke-all`,
    ),
    {
      method: "POST",
      body: JSON.stringify({
        client_txn_id:
          options.clientTxnId ?? clientTxnID("phase1-ui-revoke-all"),
        reason: options.reason,
      }),
    },
  );
}
