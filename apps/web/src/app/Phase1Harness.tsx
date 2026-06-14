import {
  type CSSProperties,
  useCallback,
  useEffect,
  useRef,
  useState,
} from "react";

import { type APIError, extractError } from "../services/browserApi";
import {
  adminResetPassword,
  adminResetTotp,
  adminRevokeAllSessions,
  beginTotpEnrollment,
  type CredentialState,
  changePassword,
  completeTotpEnrollment,
  createLocalUser,
  loadCredentialState as loadCredentialStateRequest,
  loadSession as loadSessionRequest,
  loadUser as loadUserRequest,
  loginLocal,
  logoutCurrentSession,
  patchLocalUser,
  type SessionData,
  type UserResource,
} from "./phase1Client";

type ProbeResult = {
  status: number;
  body: unknown;
};

type TargetAdminOperation = "loading" | "mutating";

function prettyJSON(value: unknown): string {
  return JSON.stringify(value, null, 2);
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
  const [targetBaseVersion, setTargetBaseVersion] = useState("");
  const [patchDisplayName, setPatchDisplayName] = useState("");
  const [patchMFARequired, setPatchMFARequired] = useState(true);
  const [patchIsActive, setPatchIsActive] = useState(true);
  const [patchIsDeploymentAdmin, setPatchIsDeploymentAdmin] = useState(false);
  const [adminNewPassword, setAdminNewPassword] = useState("");
  const [adminReason, setAdminReason] = useState("phase1 harness action");
  const [selectedUser, setSelectedUser] = useState<UserResource | null>(null);
  const [targetAdminOperation, setTargetAdminOperation] =
    useState<TargetAdminOperation | null>(null);
  const targetAdminOperationRef = useRef<{
    id: number;
    kind: TargetAdminOperation;
  } | null>(null);
  const nextTargetAdminOperationID = useRef(0);

  const targetOperationPending = targetAdminOperation !== null;
  const loadedTargetIsCurrent =
    selectedUser !== null && targetUserID.trim() === selectedUser.user_id;
  const trimmedTargetBaseVersion = targetBaseVersion.trim();
  const targetBaseVersionIsValid = /^[1-9]\d*$/.test(trimmedTargetBaseVersion);
  const parsedTargetBaseVersion = Number.parseInt(trimmedTargetBaseVersion, 10);
  const canLoadTargetUser =
    !targetOperationPending && targetUserID.trim() !== "";
  const canSubmitTargetAction =
    !targetOperationPending && loadedTargetIsCurrent;
  const canSubmitVersionedTargetAction =
    canSubmitTargetAction && targetBaseVersionIsValid;

  function beginTargetAdminOperation(kind: TargetAdminOperation) {
    if (targetAdminOperationRef.current !== null) {
      return null;
    }
    const operation = {
      id: nextTargetAdminOperationID.current + 1,
      kind,
    };
    nextTargetAdminOperationID.current = operation.id;
    targetAdminOperationRef.current = operation;
    setTargetAdminOperation(kind);
    return operation.id;
  }

  function isCurrentTargetAdminOperation(operationID: number) {
    return targetAdminOperationRef.current?.id === operationID;
  }

  function finishTargetAdminOperation(operationID: number) {
    if (!isCurrentTargetAdminOperation(operationID)) {
      return;
    }
    targetAdminOperationRef.current = null;
    setTargetAdminOperation(null);
  }

  function clearSelectedUser() {
    setSelectedUser(null);
    setTargetBaseVersion("");
    setPatchDisplayName("");
    setPatchMFARequired(true);
    setPatchIsActive(true);
    setPatchIsDeploymentAdmin(false);
  }

  function applySelectedUser(user: UserResource) {
    setSelectedUser(user);
    setTargetUserID(user.user_id);
    setTargetBaseVersion(String(user.user_version));
    setPatchDisplayName(user.display_name);
    setPatchMFARequired(user.mfa_required);
    setPatchIsActive(user.is_active);
    setPatchIsDeploymentAdmin(user.is_deployment_admin);
  }

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
    const result = await loadSessionRequest(apiBase);
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
    const result = await loadCredentialStateRequest(apiBase);
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

  async function loadUser(
    userID: string,
    options?: {
      updateProbe?: boolean;
    },
  ) {
    const updateProbeOutput = options?.updateProbe ?? true;
    const targetUserID = userID.trim();
    if (targetUserID === "") {
      if (updateProbeOutput) {
        clearSelectedUser();
      }
      return null;
    }

    const operationID = beginTargetAdminOperation("loading");
    if (operationID === null) {
      return null;
    }
    if (updateProbeOutput) {
      setStatusText("Loading target user");
      setLastError(null);
    }

    try {
      const result = await loadUserRequest({ apiBase, userId: targetUserID });
      if (!isCurrentTargetAdminOperation(operationID)) {
        return null;
      }

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
          clearSelectedUser();
        }
        return null;
      }

      const user = (result.payload as { data: UserResource }).data;
      applySelectedUser(user);
      return user;
    } finally {
      finishTargetAdminOperation(operationID);
    }
  }

  useEffect(() => {
    void loadSession();
    void loadCredentialState();
  }, [loadCredentialState, loadSession]);

  async function handleLogin() {
    setStatusText("Signing in");
    const result = await loginLocal({
      apiBase,
      username: loginUsername,
      password: loginPassword,
      secondFactorCode: loginTotpCode,
    });
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
    const result = await logoutCurrentSession(apiBase);
    updateProbe(result.status, result.payload, "Signed out", "Sign out failed");
    await loadSession();
    await loadCredentialState();
  }

  async function handleBeginTOTP() {
    setStatusText("Beginning TOTP enrollment");
    const result = await beginTotpEnrollment({
      apiBase,
      authMode: totpAuthMode,
      bootstrapToken,
      currentPassword: totpCurrentPassword,
      currentFactorCode: totpCurrentFactorCode,
    });
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
    const result = await completeTotpEnrollment({
      apiBase,
      authMode: totpAuthMode,
      bootstrapToken,
      enrollmentId: totpEnrollmentId,
      code: totpCompleteCode,
    });
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
    const result = await changePassword({
      apiBase,
      currentPassword: passwordCurrent,
      newPassword: passwordNext,
      secondFactorCode: passwordFactorCode,
    });
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
    const operationID = beginTargetAdminOperation("mutating");
    if (operationID === null) {
      return;
    }
    setStatusText("Creating local user");
    setLastError(null);

    try {
      const result = await createLocalUser({
        apiBase,
        email: createEmail,
        displayName: createDisplayName,
        initialPassword: createInitialPassword,
        mfaRequired: createMFARequired,
        isDeploymentAdmin: createIsDeploymentAdmin,
      });
      if (!isCurrentTargetAdminOperation(operationID)) {
        return;
      }

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
      applySelectedUser(user);
    } finally {
      finishTargetAdminOperation(operationID);
    }
  }

  async function handlePatchUser() {
    if (!canSubmitVersionedTargetAction || selectedUser === null) {
      return;
    }
    const operationID = beginTargetAdminOperation("mutating");
    if (operationID === null) {
      return;
    }
    const targetUserID = selectedUser.user_id;
    setStatusText("Patching local user");
    setLastError(null);

    try {
      const result = await patchLocalUser({
        apiBase,
        userId: targetUserID,
        baseUserVersion: parsedTargetBaseVersion,
        displayName: patchDisplayName,
        mfaRequired: patchMFARequired,
        isActive: patchIsActive,
        isDeploymentAdmin: patchIsDeploymentAdmin,
      });
      if (!isCurrentTargetAdminOperation(operationID)) {
        return;
      }

      updateProbe(
        result.status,
        result.payload,
        "Patched local user",
        "Patch local user failed",
      );
      if (!result.ok) {
        return;
      }

      const user = (result.payload as { data: UserResource }).data;
      applySelectedUser(user);
    } finally {
      finishTargetAdminOperation(operationID);
    }
  }

  async function handleAdminPasswordReset() {
    if (!canSubmitVersionedTargetAction || selectedUser === null) {
      return;
    }
    const operationID = beginTargetAdminOperation("mutating");
    if (operationID === null) {
      return;
    }
    const targetUserID = selectedUser.user_id;
    setStatusText("Resetting user password");
    setLastError(null);

    try {
      const result = await adminResetPassword({
        apiBase,
        userId: targetUserID,
        baseUserVersion: parsedTargetBaseVersion,
        newPassword: adminNewPassword,
        reason: adminReason,
      });
      if (!isCurrentTargetAdminOperation(operationID)) {
        return;
      }

      updateProbe(
        result.status,
        result.payload,
        "Reset user password",
        "Reset user password failed",
      );
      if (!result.ok) {
        return;
      }

      const user = (result.payload as { data: UserResource }).data;
      applySelectedUser(user);
    } finally {
      finishTargetAdminOperation(operationID);
    }
  }

  async function handleAdminTOTPReset() {
    if (!canSubmitVersionedTargetAction || selectedUser === null) {
      return;
    }
    const operationID = beginTargetAdminOperation("mutating");
    if (operationID === null) {
      return;
    }
    const targetUserID = selectedUser.user_id;
    setStatusText("Resetting user TOTP");
    setLastError(null);

    try {
      const result = await adminResetTotp({
        apiBase,
        userId: targetUserID,
        baseUserVersion: parsedTargetBaseVersion,
        reason: adminReason,
      });
      if (!isCurrentTargetAdminOperation(operationID)) {
        return;
      }

      updateProbe(
        result.status,
        result.payload,
        "Reset user TOTP",
        "Reset user TOTP failed",
      );
      if (!result.ok) {
        return;
      }

      const user = (result.payload as { data: UserResource }).data;
      applySelectedUser(user);
    } finally {
      finishTargetAdminOperation(operationID);
    }
  }

  async function handleAdminRevokeAll() {
    if (!canSubmitTargetAction || selectedUser === null) {
      return;
    }
    const operationID = beginTargetAdminOperation("mutating");
    if (operationID === null) {
      return;
    }
    const targetUserID = selectedUser.user_id;
    setStatusText("Revoking every user session");
    setLastError(null);

    try {
      const result = await adminRevokeAllSessions({
        apiBase,
        userId: targetUserID,
        reason: adminReason,
      });
      if (!isCurrentTargetAdminOperation(operationID)) {
        return;
      }

      updateProbe(
        result.status,
        result.payload,
        "Revoked every user session",
        "Revoke-all failed",
      );
    } finally {
      finishTargetAdminOperation(operationID);
    }
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
                disabled={targetOperationPending}
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
                disabled={targetOperationPending}
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
                disabled={targetOperationPending}
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
                  disabled={targetOperationPending}
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
                  disabled={targetOperationPending}
                  type="checkbox"
                  onChange={(event) => {
                    setCreateIsDeploymentAdmin(event.target.checked);
                  }}
                />
                Deployment admin
              </label>
              <button
                data-testid="phase1-admin-create-user"
                disabled={targetOperationPending}
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
                disabled={targetOperationPending}
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
                disabled={!canLoadTargetUser}
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
                disabled={!canSubmitTargetAction}
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
                disabled={!canSubmitTargetAction}
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
                  disabled={!canSubmitTargetAction}
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
                  disabled={!canSubmitTargetAction}
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
                  disabled={!canSubmitTargetAction}
                  type="checkbox"
                  onChange={(event) => {
                    setPatchIsDeploymentAdmin(event.target.checked);
                  }}
                />
                Deployment admin
              </label>
              <button
                data-testid="phase1-admin-patch-user"
                disabled={!canSubmitVersionedTargetAction}
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
                  disabled={!canSubmitVersionedTargetAction}
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
                  disabled={!canSubmitVersionedTargetAction}
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
                  disabled={!canSubmitTargetAction}
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
  fontFamily: "var(--font-mono)",
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
