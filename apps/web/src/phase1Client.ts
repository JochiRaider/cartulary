import { type APIResult, clientTxnID, fetchJSON } from "./browserApi";

export type SessionMembership = {
  incident_id: string;
  role: string;
};

export type SessionData = {
  user_id: string;
  display_name: string;
  provider_type: string;
  mfa_state: string;
  is_deployment_admin: boolean;
  authenticated_at: string;
  idle_expires_at: string;
  absolute_expires_at: string;
  session_expires_at: string;
  memberships: SessionMembership[];
};

export type CredentialState = {
  user_id: string;
  auth_kind: string;
  recovery_model: string;
  password_changed_at: string | null;
  totp: {
    state: string;
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
};

export type TotpAuthMode = "bootstrap" | "session";

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

export function loginLocal(options: {
  apiBase?: string | undefined;
  password: string;
  secondFactorCode?: string;
  username: string;
}) {
  return fetchJSON<DataEnvelope<SessionData>>(
    apiPath(options.apiBase, "/api/v1/auth/login"),
    {
      method: "POST",
      body: JSON.stringify({
        username: options.username,
        password: options.password,
        second_factor: secondFactorPayload(options.secondFactorCode ?? ""),
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
  const requestInit: RequestInit = {
    method: "POST",
    credentials: options.authMode === "bootstrap" ? "omit" : "include",
    body: JSON.stringify({
      client_txn_id: options.clientTxnId ?? clientTxnID("phase1-ui-totp-begin"),
      ...(options.authMode === "session"
        ? {
            current_password: options.currentPassword ?? "",
            second_factor: secondFactorPayload(options.currentFactorCode ?? ""),
          }
        : {}),
    }),
  };
  if (options.authMode === "bootstrap" && options.bootstrapToken?.trim()) {
    requestInit.headers = {
      Authorization: `Bearer ${options.bootstrapToken.trim()}`,
    };
  }
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
  const requestInit: RequestInit = {
    method: "POST",
    credentials: options.authMode === "bootstrap" ? "omit" : "include",
    body: JSON.stringify({
      client_txn_id:
        options.clientTxnId ?? clientTxnID("phase1-ui-totp-complete"),
      enrollment_id: options.enrollmentId,
      code: options.code,
    }),
  };
  if (options.authMode === "bootstrap" && options.bootstrapToken?.trim()) {
    requestInit.headers = {
      Authorization: `Bearer ${options.bootstrapToken.trim()}`,
    };
  }
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
        display_name: options.displayName,
        mfa_required: options.mfaRequired,
        is_active: options.isActive,
        is_deployment_admin: options.isDeploymentAdmin,
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

export function setTestClockOffset(options: {
  apiBase?: string | undefined;
  offsetSeconds: number;
}) {
  return fetchJSON<{ data: { offset_seconds: number; now: string } }>(
    apiPath(options.apiBase, "/api/v1/test/clock/set"),
    {
      method: "POST",
      body: JSON.stringify({
        offset_seconds: options.offsetSeconds,
      }),
    },
  );
}

export type Phase1Response<T> = APIResult<DataEnvelope<T>>;
