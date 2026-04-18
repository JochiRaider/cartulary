import { type CSSProperties, useCallback, useEffect, useState } from "react";

const csrfCookieName = "cartulary_csrf";
const csrfHeaderName = "X-CSRF-Token";

type APIError = {
  code: string;
  details?: Record<string, unknown>;
  message?: string;
  request_id?: string;
  status?: number;
};

type ProbeResult = {
  status: number;
  body: unknown;
};

type SessionMembership = {
  incident_id: string;
  role: string;
};

type SessionData = {
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

type CredentialState = {
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

type UserResource = {
  user_id: string;
  email: string;
  display_name: string;
  user_version: number;
  is_active: boolean;
  mfa_required: boolean;
  is_deployment_admin: boolean;
};

function readCookie(name: string): string | null {
  if (typeof document === "undefined") {
    return null;
  }

  const prefix = `${name}=`;
  for (const segment of document.cookie.split(";")) {
    const trimmed = segment.trim();
    if (trimmed.startsWith(prefix)) {
      return decodeURIComponent(trimmed.slice(prefix.length));
    }
  }
  return null;
}

async function fetchJSON<T>(
  input: RequestInfo | URL,
  init?: RequestInit,
): Promise<{
  ok: boolean;
  status: number;
  payload: T | { error?: APIError };
}> {
  const method = (init?.method ?? "GET").toUpperCase();
  const headers: Record<string, string> = {
    ...(init?.headers as Record<string, string> | undefined),
  };
  if (method !== "GET" && method !== "HEAD" && method !== "OPTIONS") {
    headers["Content-Type"] = "application/json";
    const csrfToken = readCookie(csrfCookieName);
    if (
      (init?.credentials ?? "include") === "include" &&
      csrfToken !== null &&
      csrfToken !== ""
    ) {
      headers[csrfHeaderName] = csrfToken;
    }
  }

  const response = await fetch(input, {
    credentials: "include",
    ...init,
    headers,
  });
  const contentType = response.headers.get("Content-Type") ?? "";
  const payload = contentType.includes("application/json")
    ? ((await response.json()) as T | { error?: APIError })
    : ((await response.text()) as unknown as T | { error?: APIError });
  return { ok: response.ok, status: response.status, payload };
}

function apiPath(base: string | undefined, path: string): string {
  const trimmedBase = (base ?? "").trim();
  if (trimmedBase === "") {
    return path;
  }
  return `${trimmedBase.replace(/\/$/, "")}${path}`;
}

function extractError(payload: unknown): APIError | null {
  if (!payload || typeof payload !== "object") {
    return null;
  }
  return (payload as { error?: APIError }).error ?? null;
}

function prettyJSON(value: unknown): string {
  return JSON.stringify(value, null, 2);
}

function uniqueTxn(prefix: string) {
  return `${prefix}-${Date.now()}-${Math.random().toString(16).slice(2, 8)}`;
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

export function Phase1Harness({ apiBase }: { apiBase?: string }) {
  const [statusText, setStatusText] = useState("Ready");
  const [lastError, setLastError] = useState<APIError | null>(null);
  const [lastProbe, setLastProbe] = useState<ProbeResult | null>(null);

  const [session, setSession] = useState<SessionData | null>(null);
  const [sessionError, setSessionError] = useState<APIError | null>(null);
  const [credentialState, setCredentialState] =
    useState<CredentialState | null>(null);
  const [credentialStateError, setCredentialStateError] =
    useState<APIError | null>(null);
  const [bootstrapToken, setBootstrapToken] = useState("");
  const [totpAuthMode, setTotpAuthMode] = useState<"bootstrap" | "session">(
    "bootstrap",
  );
  const [totpEnrollmentId, setTotpEnrollmentID] = useState("");
  const [totpSecretBase32, setTotpSecretBase32] = useState("");

  const [loginUsername, setLoginUsername] = useState(
    "bootstrap-admin@example.test",
  );
  const [loginPassword, setLoginPassword] = useState("BootstrapPass1!");
  const [loginTotpCode, setLoginTotpCode] = useState("");

  const [totpCurrentPassword, setTotpCurrentPassword] = useState("");
  const [totpCurrentFactorCode, setTotpCurrentFactorCode] = useState("");
  const [totpCompleteCode, setTotpCompleteCode] = useState("");

  const [passwordCurrent, setPasswordCurrent] = useState("");
  const [passwordNext, setPasswordNext] = useState("");
  const [passwordFactorCode, setPasswordFactorCode] = useState("");

  const [createEmail, setCreateEmail] = useState("");
  const [createDisplayName, setCreateDisplayName] = useState("");
  const [createInitialPassword, setCreateInitialPassword] = useState("");
  const [createMFARequired, setCreateMFARequired] = useState(true);
  const [createIsDeploymentAdmin, setCreateIsDeploymentAdmin] = useState(false);

  const [targetUserID, setTargetUserID] = useState("");
  const [targetBaseVersion, setTargetBaseVersion] = useState("1");
  const [patchDisplayName, setPatchDisplayName] = useState("");
  const [patchMFARequired, setPatchMFARequired] = useState(true);
  const [patchIsActive, setPatchIsActive] = useState(true);
  const [patchIsDeploymentAdmin, setPatchIsDeploymentAdmin] = useState(false);
  const [adminNewPassword, setAdminNewPassword] = useState("");
  const [adminReason, setAdminReason] = useState("phase1 harness action");
  const [selectedUser, setSelectedUser] = useState<UserResource | null>(null);

  const updateProbe = useCallback(
    (
      status: number,
      payload: unknown,
      successStatusText: string,
      failureStatusText: string,
    ) => {
      const error = extractError(payload);
      setLastProbe({ status, body: payload });
      setLastError(error);
      setStatusText(error === null ? successStatusText : failureStatusText);
      return error;
    },
    [],
  );

  const loadSession = useCallback(async () => {
    const result = await fetchJSON<{ data: SessionData }>(
      apiPath(apiBase, "/api/v1/auth/session"),
    );
    const error = extractError(result.payload);
    if (!result.ok) {
      setSession(null);
      setSessionError(error);
      return error;
    }

    setSessionError(null);
    setSession((result.payload as { data: SessionData }).data);
    return null;
  }, [apiBase]);

  const loadCredentialState = useCallback(async () => {
    const result = await fetchJSON<{ data: CredentialState }>(
      apiPath(apiBase, "/api/v1/auth/credential-state"),
    );
    const error = extractError(result.payload);
    if (!result.ok) {
      setCredentialState(null);
      setCredentialStateError(error);
      return error;
    }

    setCredentialStateError(null);
    setCredentialState((result.payload as { data: CredentialState }).data);
    return null;
  }, [apiBase]);

  const loadUser = useCallback(
    async (
      userID: string,
      options?: {
        updateProbe?: boolean;
      },
    ) => {
      const updateProbeOutput = options?.updateProbe ?? true;
      if (userID.trim() === "") {
        if (updateProbeOutput) {
          setSelectedUser(null);
        }
        return null;
      }

      const result = await fetchJSON<{ data: UserResource }>(
        apiPath(apiBase, `/api/v1/users/${userID}`),
      );
      if (updateProbeOutput) {
        updateProbe(
          result.status,
          result.payload,
          "Loaded target user",
          "Load target user failed",
        );
      }
      if (!result.ok) {
        if (updateProbeOutput) {
          setSelectedUser(null);
        }
        return null;
      }

      const user = (result.payload as { data: UserResource }).data;
      setSelectedUser(user);
      setTargetUserID(user.user_id);
      setTargetBaseVersion(String(user.user_version));
      setPatchDisplayName(user.display_name);
      setPatchMFARequired(user.mfa_required);
      setPatchIsActive(user.is_active);
      setPatchIsDeploymentAdmin(user.is_deployment_admin);
      return user;
    },
    [apiBase, updateProbe],
  );

  useEffect(() => {
    void loadSession();
    void loadCredentialState();
  }, [loadCredentialState, loadSession]);

  async function handleLogin() {
    setStatusText("Signing in");
    const result = await fetchJSON<{ data: SessionData }>(
      apiPath(apiBase, "/api/v1/auth/login"),
      {
        method: "POST",
        body: JSON.stringify({
          username: loginUsername,
          password: loginPassword,
          second_factor: secondFactorPayload(loginTotpCode),
        }),
      },
    );
    const error = updateProbe(
      result.status,
      result.payload,
      "Signed in",
      "Sign in failed",
    );
    if (error?.code === "mfa_setup_required") {
      const token = error.details?.bootstrap_token;
      if (typeof token === "string") {
        setBootstrapToken(token);
      }
    }
    if (!result.ok) {
      setSession(null);
      setCredentialState(null);
      return;
    }

    setBootstrapToken("");
    await loadSession();
    await loadCredentialState();
  }

  async function handleLogout() {
    setStatusText("Signing out");
    const result = await fetchJSON(apiPath(apiBase, "/api/v1/auth/logout"), {
      method: "POST",
      body: JSON.stringify({}),
    });
    updateProbe(result.status, result.payload, "Signed out", "Sign out failed");
    await loadSession();
    await loadCredentialState();
  }

  async function handleExpireCurrentSession() {
    setStatusText("Expiring current session");
    const result = await fetchJSON(
      apiPath(apiBase, "/api/v1/test/auth/expire-current"),
      {
        method: "POST",
        body: JSON.stringify({}),
      },
    );
    updateProbe(
      result.status,
      result.payload,
      "Current session expired",
      "Expire current session failed",
    );
    await loadSession();
  }

  async function handleBeginTOTP() {
    setStatusText("Beginning TOTP enrollment");
    const body: Record<string, unknown> = {
      client_txn_id: uniqueTxn("phase1-ui-totp-begin"),
    };
    if (totpAuthMode === "session") {
      body.current_password = totpCurrentPassword;
      body.second_factor = secondFactorPayload(totpCurrentFactorCode);
    }

    const headers =
      totpAuthMode === "bootstrap" && bootstrapToken.trim() !== ""
        ? { Authorization: `Bearer ${bootstrapToken}` }
        : undefined;
    const requestInit: RequestInit = {
      method: "POST",
      credentials: totpAuthMode === "bootstrap" ? "omit" : "include",
      body: JSON.stringify(body),
    };
    if (headers) {
      requestInit.headers = headers;
    }
    const result = await fetchJSON<{
      data: {
        enrollment_id: string;
        totp_setup: { secret_base32: string };
      };
    }>(apiPath(apiBase, "/api/v1/auth/mfa/totp/begin"), requestInit);
    updateProbe(
      result.status,
      result.payload,
      "Began TOTP enrollment",
      "TOTP begin failed",
    );
    if (!result.ok) {
      return;
    }

    const data = (
      result.payload as {
        data: { enrollment_id: string; totp_setup: { secret_base32: string } };
      }
    ).data;
    setTotpEnrollmentID(data.enrollment_id);
    setTotpSecretBase32(data.totp_setup.secret_base32);
  }

  async function handleCompleteTOTP() {
    setStatusText("Completing TOTP enrollment");
    const headers =
      totpAuthMode === "bootstrap" && bootstrapToken.trim() !== ""
        ? { Authorization: `Bearer ${bootstrapToken}` }
        : undefined;
    const requestInit: RequestInit = {
      method: "POST",
      credentials: totpAuthMode === "bootstrap" ? "omit" : "include",
      body: JSON.stringify({
        client_txn_id: uniqueTxn("phase1-ui-totp-complete"),
        enrollment_id: totpEnrollmentId,
        code: totpCompleteCode,
      }),
    };
    if (headers) {
      requestInit.headers = headers;
    }
    const result = await fetchJSON(
      apiPath(apiBase, "/api/v1/auth/mfa/totp/complete"),
      requestInit,
    );
    updateProbe(
      result.status,
      result.payload,
      "Completed TOTP enrollment",
      "TOTP complete failed",
    );
    await loadSession();
    await loadCredentialState();
  }

  async function handlePasswordChange() {
    setStatusText("Changing password");
    const result = await fetchJSON(
      apiPath(apiBase, "/api/v1/auth/password/change"),
      {
        method: "POST",
        body: JSON.stringify({
          client_txn_id: uniqueTxn("phase1-ui-password-change"),
          current_password: passwordCurrent,
          new_password: passwordNext,
          second_factor: secondFactorPayload(passwordFactorCode),
        }),
      },
    );
    updateProbe(
      result.status,
      result.payload,
      "Changed password",
      "Password change failed",
    );
    await loadSession();
    await loadCredentialState();
  }

  async function handleCreateUser() {
    setStatusText("Creating local user");
    const result = await fetchJSON<{ data: UserResource }>(
      apiPath(apiBase, "/api/v1/users"),
      {
        method: "POST",
        body: JSON.stringify({
          client_txn_id: uniqueTxn("phase1-ui-user-create"),
          auth_kind: "local",
          email: createEmail,
          display_name: createDisplayName,
          initial_password: createInitialPassword,
          mfa_required: createMFARequired,
          is_deployment_admin: createIsDeploymentAdmin,
        }),
      },
    );
    updateProbe(
      result.status,
      result.payload,
      "Created local user",
      "Create local user failed",
    );
    if (!result.ok) {
      return;
    }

    const user = (result.payload as { data: UserResource }).data;
    setSelectedUser(user);
    setTargetUserID(user.user_id);
    setTargetBaseVersion(String(user.user_version));
    setPatchDisplayName(user.display_name);
    setPatchMFARequired(user.mfa_required);
    setPatchIsActive(user.is_active);
    setPatchIsDeploymentAdmin(user.is_deployment_admin);
  }

  async function handlePatchUser() {
    setStatusText("Patching local user");
    const result = await fetchJSON<{ data: UserResource }>(
      apiPath(apiBase, `/api/v1/users/${targetUserID}`),
      {
        method: "PATCH",
        body: JSON.stringify({
          base_user_version: Number.parseInt(targetBaseVersion, 10),
          display_name: patchDisplayName,
          mfa_required: patchMFARequired,
          is_active: patchIsActive,
          is_deployment_admin: patchIsDeploymentAdmin,
        }),
      },
    );
    updateProbe(
      result.status,
      result.payload,
      "Patched local user",
      "Patch local user failed",
    );
    if (!result.ok) {
      return;
    }

    await loadUser(targetUserID, { updateProbe: false });
  }

  async function handleAdminPasswordReset() {
    setStatusText("Resetting user password");
    const result = await fetchJSON<{ data: UserResource }>(
      apiPath(apiBase, `/api/v1/users/${targetUserID}/password/reset`),
      {
        method: "POST",
        body: JSON.stringify({
          base_user_version: Number.parseInt(targetBaseVersion, 10),
          client_txn_id: uniqueTxn("phase1-ui-password-reset"),
          new_password: adminNewPassword,
          reason: adminReason,
        }),
      },
    );
    updateProbe(
      result.status,
      result.payload,
      "Reset user password",
      "Reset user password failed",
    );
    if (!result.ok) {
      return;
    }

    await loadUser(targetUserID, { updateProbe: false });
  }

  async function handleAdminTOTPReset() {
    setStatusText("Resetting user TOTP");
    const result = await fetchJSON<{ data: UserResource }>(
      apiPath(apiBase, `/api/v1/users/${targetUserID}/mfa/totp/reset`),
      {
        method: "POST",
        body: JSON.stringify({
          base_user_version: Number.parseInt(targetBaseVersion, 10),
          client_txn_id: uniqueTxn("phase1-ui-totp-reset"),
          reason: adminReason,
        }),
      },
    );
    updateProbe(
      result.status,
      result.payload,
      "Reset user TOTP",
      "Reset user TOTP failed",
    );
    if (!result.ok) {
      return;
    }

    await loadUser(targetUserID, { updateProbe: false });
  }

  async function handleAdminRevokeAll() {
    setStatusText("Revoking every user session");
    const result = await fetchJSON(
      apiPath(apiBase, `/api/v1/users/${targetUserID}/sessions/revoke-all`),
      {
        method: "POST",
        body: JSON.stringify({
          client_txn_id: uniqueTxn("phase1-ui-revoke-all"),
          reason: adminReason,
        }),
      },
    );
    updateProbe(
      result.status,
      result.payload,
      "Revoked every user session",
      "Revoke-all failed",
    );
    if (!result.ok) {
      return;
    }

    await loadUser(targetUserID, { updateProbe: false });
  }

  return (
    <section style={shellStyle}>
      <header style={sectionHeaderStyle}>
        <div>
          <p style={eyebrowStyle}>Phase 1 Harness</p>
          <h1 style={headlineStyle}>
            Authentication, sessions, and local admin
          </h1>
          <p style={bodyStyle}>
            Browser-visible verification for local login, session inspection,
            credential state, TOTP setup and replacement, password change, and
            deployment-local user administration.
          </p>
        </div>
        <div style={statusCardStyle}>
          <span style={labelStyle}>Status</span>
          <strong data-testid="phase1-status">{statusText}</strong>
        </div>
      </header>

      <section style={cardStyle}>
        <h2 style={subheadStyle}>Login</h2>
        <div style={formRowStyle}>
          <input
            data-testid="phase1-login-username"
            placeholder="user@example.test"
            style={inputStyle}
            type="text"
            value={loginUsername}
            onChange={(event) => {
              setLoginUsername(event.target.value);
            }}
          />
          <input
            data-testid="phase1-login-password"
            placeholder="Password"
            style={inputStyle}
            type="password"
            value={loginPassword}
            onChange={(event) => {
              setLoginPassword(event.target.value);
            }}
          />
          <input
            data-testid="phase1-login-totp-code"
            placeholder="123456"
            style={narrowInputStyle}
            type="text"
            value={loginTotpCode}
            onChange={(event) => {
              setLoginTotpCode(event.target.value);
            }}
          />
          <button
            data-testid="phase1-login"
            style={buttonStyle}
            type="button"
            onClick={() => {
              void handleLogin();
            }}
          >
            Sign in
          </button>
          <button
            data-testid="phase1-logout"
            style={secondaryButtonStyle}
            type="button"
            onClick={() => {
              void handleLogout();
            }}
          >
            Sign out
          </button>
        </div>
        <div style={detailGridStyle}>
          <div>
            <span style={labelStyle}>Bootstrap token</span>
            <div data-testid="phase1-bootstrap-token" style={monoTextStyle}>
              {bootstrapToken || ""}
            </div>
          </div>
          <div>
            <span style={labelStyle}>Last error</span>
            <div data-testid="phase1-last-error-code">
              {lastError?.code ?? ""}
            </div>
          </div>
        </div>
      </section>

      <section style={cardStyle}>
        <h2 style={subheadStyle}>Session</h2>
        <div style={formRowStyle}>
          <button
            data-testid="phase1-refresh-session"
            style={secondaryButtonStyle}
            type="button"
            onClick={() => {
              void loadSession();
            }}
          >
            Refresh session
          </button>
          <button
            data-testid="phase1-expire-session"
            style={secondaryButtonStyle}
            type="button"
            onClick={() => {
              void handleExpireCurrentSession();
            }}
          >
            Force idle expiry
          </button>
        </div>
        {session ? (
          <div style={gridStyle}>
            <div>
              <span style={labelStyle}>User ID</span>
              <div data-testid="phase1-session-user-id" style={monoTextStyle}>
                {session.user_id}
              </div>
            </div>
            <div>
              <span style={labelStyle}>Provider</span>
              <div data-testid="phase1-session-provider-type">
                {session.provider_type}
              </div>
            </div>
            <div>
              <span style={labelStyle}>MFA State</span>
              <div data-testid="phase1-session-mfa-state">
                {session.mfa_state}
              </div>
            </div>
            <div>
              <span style={labelStyle}>Deployment Admin</span>
              <div data-testid="phase1-session-is-deployment-admin">
                {String(session.is_deployment_admin)}
              </div>
            </div>
            <div>
              <span style={labelStyle}>Authenticated At</span>
              <div data-testid="phase1-session-authenticated-at">
                {session.authenticated_at}
              </div>
            </div>
            <div>
              <span style={labelStyle}>Idle Expires At</span>
              <div data-testid="phase1-session-idle-expires-at">
                {session.idle_expires_at}
              </div>
            </div>
            <div>
              <span style={labelStyle}>Absolute Expires At</span>
              <div data-testid="phase1-session-absolute-expires-at">
                {session.absolute_expires_at}
              </div>
            </div>
            <div>
              <span style={labelStyle}>Session Expires At</span>
              <div data-testid="phase1-session-session-expires-at">
                {session.session_expires_at}
              </div>
            </div>
            <div style={wideCellStyle}>
              <span style={labelStyle}>Memberships</span>
              <ul
                data-testid="phase1-session-memberships"
                style={plainListStyle}
              >
                {session.memberships.map((membership) => (
                  <li key={`${membership.incident_id}-${membership.role}`}>
                    {membership.incident_id} - {membership.role}
                  </li>
                ))}
              </ul>
            </div>
          </div>
        ) : (
          <p data-testid="phase1-session-error-code" style={bodyStyle}>
            {sessionError?.code ?? "No active session"}
          </p>
        )}
      </section>

      <section style={cardStyle}>
        <h2 style={subheadStyle}>Credential State</h2>
        <div style={formRowStyle}>
          <button
            data-testid="phase1-refresh-credential-state"
            style={secondaryButtonStyle}
            type="button"
            onClick={() => {
              void loadCredentialState();
            }}
          >
            Refresh credential state
          </button>
        </div>
        {credentialState ? (
          <div style={gridStyle}>
            <div>
              <span style={labelStyle}>User ID</span>
              <div
                data-testid="phase1-credential-user-id"
                style={monoTextStyle}
              >
                {credentialState.user_id}
              </div>
            </div>
            <div>
              <span style={labelStyle}>Auth Kind</span>
              <div data-testid="phase1-credential-auth-kind">
                {credentialState.auth_kind}
              </div>
            </div>
            <div>
              <span style={labelStyle}>Recovery Model</span>
              <div data-testid="phase1-credential-recovery-model">
                {credentialState.recovery_model}
              </div>
            </div>
            <div>
              <span style={labelStyle}>Password Changed At</span>
              <div data-testid="phase1-credential-password-changed-at">
                {credentialState.password_changed_at ?? ""}
              </div>
            </div>
            <div>
              <span style={labelStyle}>TOTP State</span>
              <div data-testid="phase1-credential-totp-state">
                {credentialState.totp.state}
              </div>
            </div>
            <div>
              <span style={labelStyle}>Pending Expires At</span>
              <div data-testid="phase1-credential-pending-expires-at">
                {credentialState.totp.pending_expires_at ?? ""}
              </div>
            </div>
          </div>
        ) : (
          <p data-testid="phase1-credential-error-code" style={bodyStyle}>
            {credentialStateError?.code ?? "No credential state available"}
          </p>
        )}
      </section>

      <section style={cardStyle}>
        <h2 style={subheadStyle}>TOTP Enrollment</h2>
        <div style={formRowStyle}>
          <label style={labelBlockStyle}>
            <span style={labelStyle}>Auth mode</span>
            <select
              data-testid="phase1-totp-auth-mode"
              style={selectStyle}
              value={totpAuthMode}
              onChange={(event) => {
                setTotpAuthMode(
                  event.target.value === "session" ? "session" : "bootstrap",
                );
              }}
            >
              <option value="bootstrap">bootstrap</option>
              <option value="session">session</option>
            </select>
          </label>
          <input
            data-testid="phase1-totp-current-password"
            placeholder="Current password"
            style={inputStyle}
            type="password"
            value={totpCurrentPassword}
            onChange={(event) => {
              setTotpCurrentPassword(event.target.value);
            }}
          />
          <input
            data-testid="phase1-totp-current-factor"
            placeholder="Current TOTP"
            style={narrowInputStyle}
            type="text"
            value={totpCurrentFactorCode}
            onChange={(event) => {
              setTotpCurrentFactorCode(event.target.value);
            }}
          />
          <button
            data-testid="phase1-totp-begin"
            style={buttonStyle}
            type="button"
            onClick={() => {
              void handleBeginTOTP();
            }}
          >
            Begin
          </button>
        </div>
        <div style={detailGridStyle}>
          <div>
            <span style={labelStyle}>Enrollment ID</span>
            <div data-testid="phase1-totp-enrollment-id" style={monoTextStyle}>
              {totpEnrollmentId}
            </div>
          </div>
          <div>
            <span style={labelStyle}>Secret Base32</span>
            <div data-testid="phase1-totp-secret-base32" style={monoTextStyle}>
              {totpSecretBase32}
            </div>
          </div>
        </div>
        <div style={formRowStyle}>
          <input
            data-testid="phase1-totp-complete-code"
            placeholder="123456"
            style={narrowInputStyle}
            type="text"
            value={totpCompleteCode}
            onChange={(event) => {
              setTotpCompleteCode(event.target.value);
            }}
          />
          <button
            data-testid="phase1-totp-complete"
            style={buttonStyle}
            type="button"
            onClick={() => {
              void handleCompleteTOTP();
            }}
          >
            Complete
          </button>
        </div>
      </section>

      <section style={cardStyle}>
        <h2 style={subheadStyle}>Password Change</h2>
        <div style={formRowStyle}>
          <input
            data-testid="phase1-password-current"
            placeholder="Current password"
            style={inputStyle}
            type="password"
            value={passwordCurrent}
            onChange={(event) => {
              setPasswordCurrent(event.target.value);
            }}
          />
          <input
            data-testid="phase1-password-next"
            placeholder="New password"
            style={inputStyle}
            type="password"
            value={passwordNext}
            onChange={(event) => {
              setPasswordNext(event.target.value);
            }}
          />
          <input
            data-testid="phase1-password-factor-code"
            placeholder="123456"
            style={narrowInputStyle}
            type="text"
            value={passwordFactorCode}
            onChange={(event) => {
              setPasswordFactorCode(event.target.value);
            }}
          />
          <button
            data-testid="phase1-password-change"
            style={buttonStyle}
            type="button"
            onClick={() => {
              void handlePasswordChange();
            }}
          >
            Change password
          </button>
        </div>
      </section>

      <section style={cardStyle}>
        <h2 style={subheadStyle}>Deployment-Local User Admin</h2>
        <div style={gridStyle}>
          <div style={wideCellStyle}>
            <span style={labelStyle}>Create local user</span>
            <div style={stackStyle}>
              <input
                data-testid="phase1-admin-create-email"
                placeholder="new.user@example.test"
                style={inputStyle}
                type="text"
                value={createEmail}
                onChange={(event) => {
                  setCreateEmail(event.target.value);
                }}
              />
              <input
                data-testid="phase1-admin-create-display-name"
                placeholder="Display name"
                style={inputStyle}
                type="text"
                value={createDisplayName}
                onChange={(event) => {
                  setCreateDisplayName(event.target.value);
                }}
              />
              <input
                data-testid="phase1-admin-create-password"
                placeholder="Initial password"
                style={inputStyle}
                type="password"
                value={createInitialPassword}
                onChange={(event) => {
                  setCreateInitialPassword(event.target.value);
                }}
              />
              <label style={checkboxStyle}>
                <input
                  data-testid="phase1-admin-create-mfa-required"
                  checked={createMFARequired}
                  type="checkbox"
                  onChange={(event) => {
                    setCreateMFARequired(event.target.checked);
                  }}
                />
                MFA required
              </label>
              <label style={checkboxStyle}>
                <input
                  data-testid="phase1-admin-create-is-deployment-admin"
                  checked={createIsDeploymentAdmin}
                  type="checkbox"
                  onChange={(event) => {
                    setCreateIsDeploymentAdmin(event.target.checked);
                  }}
                />
                Deployment admin
              </label>
              <button
                data-testid="phase1-admin-create-user"
                style={buttonStyle}
                type="button"
                onClick={() => {
                  void handleCreateUser();
                }}
              >
                Create user
              </button>
            </div>
          </div>

          <div style={wideCellStyle}>
            <span style={labelStyle}>Target user</span>
            <div style={stackStyle}>
              <input
                data-testid="phase1-admin-target-user-id-input"
                placeholder="Target user UUID"
                style={inputStyle}
                type="text"
                value={targetUserID}
                onChange={(event) => {
                  setTargetUserID(event.target.value);
                }}
              />
              <button
                data-testid="phase1-admin-load-user"
                style={secondaryButtonStyle}
                type="button"
                onClick={() => {
                  void loadUser(targetUserID);
                }}
              >
                Load user
              </button>
              <div
                data-testid="phase1-admin-target-user-id"
                style={monoTextStyle}
              >
                {selectedUser?.user_id ?? ""}
              </div>
              <div data-testid="phase1-admin-target-user-version">
                {selectedUser?.user_version ?? ""}
              </div>
              <div data-testid="phase1-admin-target-is-active">
                {selectedUser ? String(selectedUser.is_active) : ""}
              </div>
              <div data-testid="phase1-admin-target-is-deployment-admin">
                {selectedUser ? String(selectedUser.is_deployment_admin) : ""}
              </div>
            </div>
          </div>

          <div style={wideCellStyle}>
            <span style={labelStyle}>Patch target user</span>
            <div style={stackStyle}>
              <input
                data-testid="phase1-admin-patch-base-version"
                placeholder="Base user version"
                style={narrowInputStyle}
                type="number"
                value={targetBaseVersion}
                onChange={(event) => {
                  setTargetBaseVersion(event.target.value);
                }}
              />
              <input
                data-testid="phase1-admin-patch-display-name"
                placeholder="Display name"
                style={inputStyle}
                type="text"
                value={patchDisplayName}
                onChange={(event) => {
                  setPatchDisplayName(event.target.value);
                }}
              />
              <label style={checkboxStyle}>
                <input
                  data-testid="phase1-admin-patch-mfa-required"
                  checked={patchMFARequired}
                  type="checkbox"
                  onChange={(event) => {
                    setPatchMFARequired(event.target.checked);
                  }}
                />
                MFA required
              </label>
              <label style={checkboxStyle}>
                <input
                  data-testid="phase1-admin-patch-is-active"
                  checked={patchIsActive}
                  type="checkbox"
                  onChange={(event) => {
                    setPatchIsActive(event.target.checked);
                  }}
                />
                Active
              </label>
              <label style={checkboxStyle}>
                <input
                  data-testid="phase1-admin-patch-is-deployment-admin"
                  checked={patchIsDeploymentAdmin}
                  type="checkbox"
                  onChange={(event) => {
                    setPatchIsDeploymentAdmin(event.target.checked);
                  }}
                />
                Deployment admin
              </label>
              <button
                data-testid="phase1-admin-patch-user"
                style={buttonStyle}
                type="button"
                onClick={() => {
                  void handlePatchUser();
                }}
              >
                Patch user
              </button>
            </div>
          </div>

          <div style={wideCellStyle}>
            <span style={labelStyle}>Credential actions</span>
            <div style={stackStyle}>
              <input
                data-testid="phase1-admin-new-password"
                placeholder="Replacement password"
                style={inputStyle}
                type="password"
                value={adminNewPassword}
                onChange={(event) => {
                  setAdminNewPassword(event.target.value);
                }}
              />
              <input
                data-testid="phase1-admin-reason"
                placeholder="Reason"
                style={inputStyle}
                type="text"
                value={adminReason}
                onChange={(event) => {
                  setAdminReason(event.target.value);
                }}
              />
              <div style={formRowStyle}>
                <button
                  data-testid="phase1-admin-password-reset"
                  style={buttonStyle}
                  type="button"
                  onClick={() => {
                    void handleAdminPasswordReset();
                  }}
                >
                  Password reset
                </button>
                <button
                  data-testid="phase1-admin-totp-reset"
                  style={secondaryButtonStyle}
                  type="button"
                  onClick={() => {
                    void handleAdminTOTPReset();
                  }}
                >
                  TOTP reset
                </button>
                <button
                  data-testid="phase1-admin-revoke-all"
                  style={secondaryButtonStyle}
                  type="button"
                  onClick={() => {
                    void handleAdminRevokeAll();
                  }}
                >
                  Revoke all
                </button>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section style={cardStyle}>
        <h2 style={subheadStyle}>Probe Output</h2>
        <div style={detailGridStyle}>
          <div>
            <span style={labelStyle}>HTTP Status</span>
            <div data-testid="phase1-last-http-status">
              {lastProbe?.status ?? ""}
            </div>
          </div>
          <div>
            <span style={labelStyle}>Error Details</span>
            <pre data-testid="phase1-last-error-details" style={codeBlockStyle}>
              {prettyJSON(lastError?.details ?? {})}
            </pre>
          </div>
        </div>
        <pre data-testid="phase1-last-response" style={codeBlockStyle}>
          {prettyJSON(lastProbe?.body ?? {})}
        </pre>
      </section>
    </section>
  );
}

const shellStyle: CSSProperties = {
  display: "grid",
  gap: "1rem",
};

const sectionHeaderStyle: CSSProperties = {
  display: "flex",
  justifyContent: "space-between",
  gap: "1rem",
  alignItems: "flex-start",
  flexWrap: "wrap",
};

const eyebrowStyle: CSSProperties = {
  margin: 0,
  letterSpacing: "0.1em",
  textTransform: "uppercase",
  fontSize: "0.75rem",
  color: "#52607a",
};

const headlineStyle: CSSProperties = {
  margin: "0.35rem 0",
  fontSize: "1.85rem",
  lineHeight: 1.1,
};

const bodyStyle: CSSProperties = {
  margin: 0,
  color: "#334155",
  maxWidth: "64ch",
};

const statusCardStyle: CSSProperties = {
  minWidth: "14rem",
  padding: "0.9rem 1rem",
  borderRadius: "1rem",
  background: "linear-gradient(135deg, #f7fbff 0%, #eff6ff 100%)",
  border: "1px solid #bfdbfe",
  boxShadow: "0 10px 30px rgba(30, 64, 175, 0.08)",
};

const labelStyle: CSSProperties = {
  display: "block",
  fontSize: "0.76rem",
  fontWeight: 700,
  letterSpacing: "0.04em",
  textTransform: "uppercase",
  color: "#52607a",
  marginBottom: "0.35rem",
};

const labelBlockStyle: CSSProperties = {
  display: "grid",
  gap: "0.35rem",
};

const cardStyle: CSSProperties = {
  padding: "1rem",
  borderRadius: "1rem",
  background: "rgba(255, 255, 255, 0.94)",
  border: "1px solid #dbeafe",
  boxShadow: "0 20px 45px rgba(15, 23, 42, 0.07)",
};

const subheadStyle: CSSProperties = {
  marginTop: 0,
  marginBottom: "0.8rem",
  fontSize: "1.1rem",
};

const formRowStyle: CSSProperties = {
  display: "flex",
  gap: "0.75rem",
  flexWrap: "wrap",
  alignItems: "center",
};

const stackStyle: CSSProperties = {
  display: "grid",
  gap: "0.65rem",
};

const inputStyle: CSSProperties = {
  minWidth: "14rem",
  padding: "0.7rem 0.85rem",
  borderRadius: "0.85rem",
  border: "1px solid #cbd5e1",
  background: "#f8fafc",
};

const narrowInputStyle: CSSProperties = {
  ...inputStyle,
  minWidth: "8rem",
};

const selectStyle: CSSProperties = {
  ...inputStyle,
  minWidth: "10rem",
};

const buttonStyle: CSSProperties = {
  padding: "0.72rem 1rem",
  borderRadius: "999px",
  border: "none",
  background: "#0f172a",
  color: "#f8fafc",
  cursor: "pointer",
  fontWeight: 700,
};

const secondaryButtonStyle: CSSProperties = {
  ...buttonStyle,
  background: "#e2e8f0",
  color: "#0f172a",
};

const gridStyle: CSSProperties = {
  display: "grid",
  gap: "0.85rem",
  gridTemplateColumns: "repeat(auto-fit, minmax(13rem, 1fr))",
};

const detailGridStyle: CSSProperties = {
  display: "grid",
  gap: "0.85rem",
  gridTemplateColumns: "repeat(auto-fit, minmax(18rem, 1fr))",
};

const wideCellStyle: CSSProperties = {
  gridColumn: "1 / -1",
};

const plainListStyle: CSSProperties = {
  margin: 0,
  paddingLeft: "1rem",
};

const monoTextStyle: CSSProperties = {
  fontFamily: "ui-monospace, SFMono-Regular, SF Mono, Menlo, monospace",
  wordBreak: "break-all",
};

const codeBlockStyle: CSSProperties = {
  margin: 0,
  padding: "0.85rem",
  borderRadius: "0.85rem",
  background: "#0f172a",
  color: "#e2e8f0",
  overflowX: "auto",
  fontSize: "0.82rem",
};

const checkboxStyle: CSSProperties = {
  display: "flex",
  gap: "0.55rem",
  alignItems: "center",
};
