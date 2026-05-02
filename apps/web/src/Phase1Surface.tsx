import { type CSSProperties, useRef, useState } from "react";

import { type APIError, extractError } from "./browserApi";
import {
  adminResetPassword,
  adminResetTotp,
  adminRevokeAllSessions,
  beginTotpEnrollment,
  type CredentialState,
  changePassword,
  completeTotpEnrollment,
  createLocalUser,
  loadUser,
  loginLocal,
  logoutCurrentSession,
  patchLocalUser,
  type SessionData,
  type UserResource,
} from "./phase1Client";

type RefreshOptions = {
  anonymousMessage?: string;
};

type Phase1AuthSurfaceProps = {
  isCheckingSession: boolean;
  message: string;
  onAuthenticated: () => Promise<void> | void;
};

type Phase1AccountPanelProps = {
  credentialState: CredentialState | null;
  credentialStateError: APIError | null;
  onRefreshShell: (options?: RefreshOptions) => Promise<void> | void;
  session: SessionData;
};

type Phase1AdminPanelProps = {
  onRefreshShell: (options?: RefreshOptions) => Promise<void> | void;
  session: SessionData;
};

type TargetAdminOperation = "loading" | "mutating";

export function Phase1AuthSurface({
  isCheckingSession,
  message,
  onAuthenticated,
}: Phase1AuthSurfaceProps) {
  const [statusText, setStatusText] = useState("Ready to sign in.");
  const [error, setError] = useState<APIError | null>(null);
  const [bootstrapToken, setBootstrapToken] = useState("");
  const [bootstrapEnrollmentId, setBootstrapEnrollmentId] = useState("");
  const [bootstrapSecretBase32, setBootstrapSecretBase32] = useState("");

  const [username, setUsername] = useState("bootstrap-admin@example.test");
  const [password, setPassword] = useState("BootstrapPass1!");
  const [totpCode, setTotpCode] = useState("");
  const [bootstrapCompleteCode, setBootstrapCompleteCode] = useState("");

  async function handleLogin() {
    setStatusText("Signing in");
    const result = await loginLocal({
      username,
      password,
      secondFactorCode: totpCode,
    });
    const nextError = extractError(result.payload);
    setError(nextError);
    if (!result.ok) {
      if (nextError?.code === "mfa_setup_required") {
        const token = nextError.details?.bootstrap_token;
        if (typeof token === "string") {
          setBootstrapToken(token);
        }
        setStatusText("TOTP enrollment is required before sign-in.");
        return;
      }
      setStatusText("Sign in failed");
      return;
    }

    setBootstrapToken("");
    setBootstrapEnrollmentId("");
    setBootstrapSecretBase32("");
    setError(null);
    setStatusText("Signed in");
    await onAuthenticated();
  }

  async function handleBeginBootstrapEnrollment() {
    setStatusText("Beginning TOTP enrollment");
    const result = await beginTotpEnrollment({
      authMode: "bootstrap",
      bootstrapToken,
    });
    const nextError = extractError(result.payload);
    setError(nextError);
    if (!result.ok) {
      setStatusText("TOTP begin failed");
      return;
    }

    const data = (
      result.payload as {
        data: { enrollment_id: string; totp_setup: { secret_base32: string } };
      }
    ).data;
    setBootstrapEnrollmentId(data.enrollment_id);
    setBootstrapSecretBase32(data.totp_setup.secret_base32);
    setStatusText("Began TOTP enrollment");
  }

  async function handleCompleteBootstrapEnrollment() {
    setStatusText("Completing TOTP enrollment");
    const result = await completeTotpEnrollment({
      authMode: "bootstrap",
      bootstrapToken,
      code: bootstrapCompleteCode,
      enrollmentId: bootstrapEnrollmentId,
    });
    const nextError = extractError(result.payload);
    setError(nextError);
    if (!result.ok) {
      setStatusText("TOTP complete failed");
      return;
    }

    setBootstrapToken("");
    setBootstrapEnrollmentId("");
    setBootstrapSecretBase32("");
    setBootstrapCompleteCode("");
    setTotpCode("");
    setError(null);
    setStatusText("TOTP enrollment completed. Sign in with your TOTP code.");
  }

  return (
    <main style={shellStyle}>
      <section style={shellPanelStyle}>
        <header style={sectionHeaderStyle}>
          <div>
            <p style={eyebrowStyle}>Cartulary</p>
            <h1 style={headlineStyle}>Local sign-in</h1>
            <p data-testid="auth-shell-message" style={bodyStyle}>
              {message}
            </p>
          </div>
          <div style={statusCardStyle}>
            <span style={labelStyle}>Status</span>
            <strong data-testid="auth-status">
              {isCheckingSession ? "Checking current session…" : statusText}
            </strong>
          </div>
        </header>

        <section style={cardStyle}>
          <div style={cardHeaderStyle}>
            <div>
              <p style={sectionEyebrowStyle}>Authentication</p>
              <h2 style={sectionTitleStyle}>
                Sign in to the ordinary app shell
              </h2>
            </div>
          </div>
          <div style={formGridStyle}>
            <label htmlFor="auth-login-username" style={labelBlockStyle}>
              Email
            </label>
            <input
              data-testid="auth-login-username"
              id="auth-login-username"
              style={inputStyle}
              value={username}
              onChange={(event) => {
                setUsername(event.target.value);
              }}
            />
            <label htmlFor="auth-login-password" style={labelBlockStyle}>
              Password
            </label>
            <input
              data-testid="auth-login-password"
              id="auth-login-password"
              style={inputStyle}
              type="password"
              value={password}
              onChange={(event) => {
                setPassword(event.target.value);
              }}
            />
            <label htmlFor="auth-login-totp-code" style={labelBlockStyle}>
              TOTP code
            </label>
            <input
              data-testid="auth-login-totp-code"
              id="auth-login-totp-code"
              style={inputStyle}
              value={totpCode}
              onChange={(event) => {
                setTotpCode(event.target.value);
              }}
            />
          </div>
          <div style={buttonRowStyle}>
            <button
              data-testid="auth-login-submit"
              style={buttonStyle}
              type="button"
              onClick={() => {
                void handleLogin();
              }}
            >
              Sign in
            </button>
          </div>
          <p data-testid="auth-error-code" style={errorStyle}>
            {error?.code ?? ""}
          </p>
        </section>

        {bootstrapToken !== "" ? (
          <section style={cardStyle}>
            <div style={cardHeaderStyle}>
              <div>
                <p style={sectionEyebrowStyle}>Bootstrap</p>
                <h2 style={sectionTitleStyle}>Complete TOTP enrollment</h2>
              </div>
            </div>
            <p style={bodyStyle}>
              This account requires TOTP enrollment before it can create a
              session.
            </p>
            <div style={detailGridStyle}>
              <div>
                <span style={labelStyle}>Bootstrap token</span>
                <div data-testid="auth-bootstrap-token" style={monoTextStyle}>
                  {bootstrapToken}
                </div>
              </div>
              <div>
                <span style={labelStyle}>Enrollment id</span>
                <div
                  data-testid="auth-bootstrap-enrollment-id"
                  style={monoTextStyle}
                >
                  {bootstrapEnrollmentId}
                </div>
              </div>
              <div style={wideCellStyle}>
                <span style={labelStyle}>Secret base32</span>
                <div
                  data-testid="auth-bootstrap-secret-base32"
                  style={monoTextStyle}
                >
                  {bootstrapSecretBase32}
                </div>
              </div>
            </div>
            <div style={buttonRowStyle}>
              <button
                data-testid="auth-bootstrap-begin"
                style={buttonStyle}
                type="button"
                onClick={() => {
                  void handleBeginBootstrapEnrollment();
                }}
              >
                Begin enrollment
              </button>
            </div>
            <div style={formGridStyle}>
              <label
                htmlFor="auth-bootstrap-complete-code"
                style={labelBlockStyle}
              >
                TOTP code
              </label>
              <input
                data-testid="auth-bootstrap-complete-code"
                id="auth-bootstrap-complete-code"
                style={inputStyle}
                value={bootstrapCompleteCode}
                onChange={(event) => {
                  setBootstrapCompleteCode(event.target.value);
                }}
              />
            </div>
            <div style={buttonRowStyle}>
              <button
                data-testid="auth-bootstrap-complete"
                style={buttonStyle}
                type="button"
                onClick={() => {
                  void handleCompleteBootstrapEnrollment();
                }}
              >
                Complete enrollment
              </button>
            </div>
          </section>
        ) : null}
      </section>
    </main>
  );
}

export function Phase1AccountPanel({
  credentialState,
  credentialStateError,
  onRefreshShell,
  session,
}: Phase1AccountPanelProps) {
  const [statusText, setStatusText] = useState("Account security is current.");
  const [error, setError] = useState<APIError | null>(null);

  const [passwordCurrent, setPasswordCurrent] = useState("");
  const [passwordNext, setPasswordNext] = useState("");
  const [passwordFactorCode, setPasswordFactorCode] = useState("");

  const [totpCurrentPassword, setTotpCurrentPassword] = useState("");
  const [totpCurrentFactorCode, setTotpCurrentFactorCode] = useState("");
  const [totpEnrollmentId, setTotpEnrollmentId] = useState("");
  const [totpSecretBase32, setTotpSecretBase32] = useState("");
  const [totpCompleteCode, setTotpCompleteCode] = useState("");

  async function handleLogout() {
    setStatusText("Signing out");
    const result = await logoutCurrentSession();
    const nextError = extractError(result.payload);
    setError(nextError);
    if (!result.ok) {
      setStatusText("Sign out failed");
      return;
    }
    await onRefreshShell({ anonymousMessage: "Signed out." });
  }

  async function handleRefreshAccount() {
    setStatusText("Refreshing account security");
    setError(null);
    await onRefreshShell();
    setStatusText("Refreshed account security.");
  }

  async function handlePasswordChange() {
    setStatusText("Changing password");
    const result = await changePassword({
      currentPassword: passwordCurrent,
      newPassword: passwordNext,
      secondFactorCode: passwordFactorCode,
    });
    const nextError = extractError(result.payload);
    setError(nextError);
    if (!result.ok) {
      setStatusText("Password change failed");
      return;
    }
    await onRefreshShell({
      anonymousMessage: "Password changed. Sign in again.",
    });
  }

  async function handleBeginTotpReplacement() {
    setStatusText("Beginning TOTP enrollment");
    const result = await beginTotpEnrollment({
      authMode: "session",
      currentPassword: totpCurrentPassword,
      currentFactorCode: totpCurrentFactorCode,
    });
    const nextError = extractError(result.payload);
    setError(nextError);
    if (!result.ok) {
      setStatusText("TOTP begin failed");
      return;
    }
    const data = (
      result.payload as {
        data: { enrollment_id: string; totp_setup: { secret_base32: string } };
      }
    ).data;
    setTotpEnrollmentId(data.enrollment_id);
    setTotpSecretBase32(data.totp_setup.secret_base32);
    setStatusText("Began TOTP enrollment");
  }

  async function handleCompleteTotpReplacement() {
    setStatusText("Completing TOTP enrollment");
    const result = await completeTotpEnrollment({
      authMode: "session",
      code: totpCompleteCode,
      enrollmentId: totpEnrollmentId,
    });
    const nextError = extractError(result.payload);
    setError(nextError);
    if (!result.ok) {
      setStatusText("TOTP complete failed");
      return;
    }
    await onRefreshShell({
      anonymousMessage: "TOTP enrollment completed. Sign in again.",
    });
  }

  return (
    <section style={cardStyle}>
      <div style={cardHeaderStyle}>
        <div>
          <p style={sectionEyebrowStyle}>Account</p>
          <h2 style={sectionTitleStyle}>Session and credential security</h2>
        </div>
        <div style={buttonRowStyle}>
          <button
            data-testid="account-refresh-state"
            style={secondaryButtonStyle}
            type="button"
            onClick={() => {
              void handleRefreshAccount();
            }}
          >
            Refresh
          </button>
          <button
            data-testid="account-logout"
            style={secondaryButtonStyle}
            type="button"
            onClick={() => {
              void handleLogout();
            }}
          >
            Sign out
          </button>
        </div>
      </div>

      <div style={detailGridStyle}>
        <div>
          <span style={labelStyle}>User id</span>
          <div data-testid="account-session-user-id" style={monoTextStyle}>
            {session.user_id}
          </div>
        </div>
        <div>
          <span style={labelStyle}>Provider</span>
          <div data-testid="account-session-provider-type">
            {session.provider_type}
          </div>
        </div>
        <div>
          <span style={labelStyle}>MFA state</span>
          <div data-testid="account-session-mfa-state">{session.mfa_state}</div>
        </div>
        <div>
          <span style={labelStyle}>Deployment admin</span>
          <div data-testid="account-session-is-deployment-admin">
            {String(session.is_deployment_admin)}
          </div>
        </div>
        <div>
          <span style={labelStyle}>Authenticated at</span>
          <div data-testid="account-session-authenticated-at">
            {session.authenticated_at}
          </div>
        </div>
        <div>
          <span style={labelStyle}>Idle expires at</span>
          <div data-testid="account-session-idle-expires-at">
            {session.idle_expires_at}
          </div>
        </div>
        <div>
          <span style={labelStyle}>Absolute expires at</span>
          <div data-testid="account-session-absolute-expires-at">
            {session.absolute_expires_at}
          </div>
        </div>
        <div>
          <span style={labelStyle}>Session expires at</span>
          <div data-testid="account-session-session-expires-at">
            {session.session_expires_at}
          </div>
        </div>
        <div style={wideCellStyle}>
          <span style={labelStyle}>Incident memberships</span>
          <ul data-testid="account-session-memberships" style={plainListStyle}>
            {session.memberships.map((membership) => (
              <li key={`${membership.incident_id}:${membership.role}`}>
                {membership.incident_id} · {membership.role}
              </li>
            ))}
          </ul>
        </div>
      </div>

      <div style={detailGridStyle}>
        <div>
          <span style={labelStyle}>Auth kind</span>
          <div data-testid="account-credential-auth-kind">
            {credentialState?.auth_kind ?? ""}
          </div>
        </div>
        <div>
          <span style={labelStyle}>Recovery model</span>
          <div data-testid="account-credential-recovery-model">
            {credentialState?.recovery_model ?? ""}
          </div>
        </div>
        <div>
          <span style={labelStyle}>Password changed at</span>
          <div data-testid="account-credential-password-changed-at">
            {credentialState?.password_changed_at ?? ""}
          </div>
        </div>
        <div>
          <span style={labelStyle}>TOTP state</span>
          <div data-testid="account-credential-totp-state">
            {credentialState?.totp.state ?? ""}
          </div>
        </div>
        <div style={wideCellStyle}>
          <span style={labelStyle}>Pending expires at</span>
          <div data-testid="account-credential-pending-expires-at">
            {credentialState?.totp.pending_expires_at ?? ""}
          </div>
        </div>
      </div>

      <section style={subsectionStyle}>
        <p style={subsectionTitleStyle}>Password change</p>
        <div style={formGridStyle}>
          <label htmlFor="account-password-current" style={labelBlockStyle}>
            Current password
          </label>
          <input
            data-testid="account-password-current"
            id="account-password-current"
            style={inputStyle}
            type="password"
            value={passwordCurrent}
            onChange={(event) => {
              setPasswordCurrent(event.target.value);
            }}
          />
          <label htmlFor="account-password-next" style={labelBlockStyle}>
            New password
          </label>
          <input
            data-testid="account-password-next"
            id="account-password-next"
            style={inputStyle}
            type="password"
            value={passwordNext}
            onChange={(event) => {
              setPasswordNext(event.target.value);
            }}
          />
          <label htmlFor="account-password-factor" style={labelBlockStyle}>
            Current TOTP code
          </label>
          <input
            data-testid="account-password-factor-code"
            id="account-password-factor"
            style={inputStyle}
            value={passwordFactorCode}
            onChange={(event) => {
              setPasswordFactorCode(event.target.value);
            }}
          />
        </div>
        <div style={buttonRowStyle}>
          <button
            data-testid="account-password-change"
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

      <section style={subsectionStyle}>
        <p style={subsectionTitleStyle}>TOTP replacement</p>
        <div style={formGridStyle}>
          <label
            htmlFor="account-totp-current-password"
            style={labelBlockStyle}
          >
            Current password
          </label>
          <input
            data-testid="account-totp-current-password"
            id="account-totp-current-password"
            style={inputStyle}
            type="password"
            value={totpCurrentPassword}
            onChange={(event) => {
              setTotpCurrentPassword(event.target.value);
            }}
          />
          <label htmlFor="account-totp-current-factor" style={labelBlockStyle}>
            Current TOTP code
          </label>
          <input
            data-testid="account-totp-current-factor"
            id="account-totp-current-factor"
            style={inputStyle}
            value={totpCurrentFactorCode}
            onChange={(event) => {
              setTotpCurrentFactorCode(event.target.value);
            }}
          />
        </div>
        <div style={buttonRowStyle}>
          <button
            data-testid="account-totp-begin"
            style={buttonStyle}
            type="button"
            onClick={() => {
              void handleBeginTotpReplacement();
            }}
          >
            Begin TOTP enrollment
          </button>
        </div>
        <div style={detailGridStyle}>
          <div>
            <span style={labelStyle}>Enrollment id</span>
            <div data-testid="account-totp-enrollment-id" style={monoTextStyle}>
              {totpEnrollmentId}
            </div>
          </div>
          <div style={wideCellStyle}>
            <span style={labelStyle}>Secret base32</span>
            <div data-testid="account-totp-secret-base32" style={monoTextStyle}>
              {totpSecretBase32}
            </div>
          </div>
        </div>
        <div style={formGridStyle}>
          <label htmlFor="account-totp-complete-code" style={labelBlockStyle}>
            Replacement TOTP code
          </label>
          <input
            data-testid="account-totp-complete-code"
            id="account-totp-complete-code"
            style={inputStyle}
            value={totpCompleteCode}
            onChange={(event) => {
              setTotpCompleteCode(event.target.value);
            }}
          />
        </div>
        <div style={buttonRowStyle}>
          <button
            data-testid="account-totp-complete"
            style={buttonStyle}
            type="button"
            onClick={() => {
              void handleCompleteTotpReplacement();
            }}
          >
            Complete TOTP enrollment
          </button>
        </div>
      </section>

      <p data-testid="account-status" style={statusTextStyle}>
        {statusText}
      </p>
      <p data-testid="account-error-code" style={errorStyle}>
        {error?.code ?? credentialStateError?.code ?? ""}
      </p>
    </section>
  );
}

export function Phase1AdminPanel({
  onRefreshShell,
  session,
}: Phase1AdminPanelProps) {
  const [statusText, setStatusText] = useState(
    session.is_deployment_admin
      ? "Deployment user administration is ready."
      : "Deployment admin access is required for user administration.",
  );
  const [error, setError] = useState<APIError | null>(null);
  const [selectedUser, setSelectedUser] = useState<UserResource | null>(null);

  const [createEmail, setCreateEmail] = useState("");
  const [createDisplayName, setCreateDisplayName] = useState("");
  const [createInitialPassword, setCreateInitialPassword] = useState("");
  const [createMfaRequired, setCreateMfaRequired] = useState(true);
  const [createIsDeploymentAdmin, setCreateIsDeploymentAdmin] = useState(false);

  const [targetUserId, setTargetUserId] = useState("");
  const [targetBaseVersion, setTargetBaseVersion] = useState("");
  const [patchDisplayName, setPatchDisplayName] = useState("");
  const [patchMfaRequired, setPatchMfaRequired] = useState(true);
  const [patchIsActive, setPatchIsActive] = useState(true);
  const [patchIsDeploymentAdmin, setPatchIsDeploymentAdmin] = useState(false);
  const [adminNewPassword, setAdminNewPassword] = useState("");
  const [adminReason, setAdminReason] = useState("ordinary admin action");
  const [targetAdminOperation, setTargetAdminOperation] =
    useState<TargetAdminOperation | null>(null);
  const targetAdminOperationRef = useRef<{
    id: number;
    kind: TargetAdminOperation;
  } | null>(null);
  const nextTargetAdminOperationID = useRef(0);

  const targetOperationPending = targetAdminOperation !== null;
  const loadedTargetIsCurrent =
    selectedUser !== null && targetUserId.trim() === selectedUser.user_id;
  const trimmedTargetBaseVersion = targetBaseVersion.trim();
  const targetBaseVersionIsValid = /^[1-9]\d*$/.test(trimmedTargetBaseVersion);
  const parsedTargetBaseVersion = Number.parseInt(trimmedTargetBaseVersion, 10);
  const canLoadTargetUser =
    !targetOperationPending && targetUserId.trim() !== "";
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
    setPatchMfaRequired(true);
    setPatchIsActive(true);
    setPatchIsDeploymentAdmin(false);
  }

  function applySelectedUser(user: UserResource) {
    setSelectedUser(user);
    setTargetUserId(user.user_id);
    setTargetBaseVersion(String(user.user_version));
    setPatchDisplayName(user.display_name);
    setPatchMfaRequired(user.mfa_required);
    setPatchIsActive(user.is_active);
    setPatchIsDeploymentAdmin(user.is_deployment_admin);
  }

  async function loadSelectedUser(
    userId: string,
    options?: {
      preserveStatus?: boolean;
    },
  ) {
    const targetUserID = userId.trim();
    if (targetUserID === "") {
      clearSelectedUser();
      return;
    }

    const operationID = beginTargetAdminOperation("loading");
    if (operationID === null) {
      return;
    }
    setStatusText("Loading target user");
    setError(null);

    try {
      const result = await loadUser({ userId: targetUserID });
      if (!isCurrentTargetAdminOperation(operationID)) {
        return;
      }

      const nextError = extractError(result.payload);
      setError(nextError);
      if (!result.ok) {
        if (!options?.preserveStatus) {
          setStatusText("Load target user failed");
        }
        clearSelectedUser();
        return;
      }

      const user = (result.payload as { data: UserResource }).data;
      applySelectedUser(user);
      if (!options?.preserveStatus) {
        setStatusText("Loaded target user");
      }
      setError(null);
    } finally {
      finishTargetAdminOperation(operationID);
    }
  }

  async function handleCreateUser() {
    const operationID = beginTargetAdminOperation("mutating");
    if (operationID === null) {
      return;
    }
    setStatusText("Creating local user");
    setError(null);

    try {
      const result = await createLocalUser({
        email: createEmail,
        displayName: createDisplayName,
        initialPassword: createInitialPassword,
        mfaRequired: createMfaRequired,
        isDeploymentAdmin: createIsDeploymentAdmin,
      });
      if (!isCurrentTargetAdminOperation(operationID)) {
        return;
      }

      const nextError = extractError(result.payload);
      setError(nextError);
      if (!result.ok) {
        setStatusText("Create local user failed");
        return;
      }

      const user = (result.payload as { data: UserResource }).data;
      applySelectedUser(user);
      setStatusText("Created local user");
      setError(null);
      await onRefreshShell();
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
    setError(null);

    try {
      const result = await patchLocalUser({
        userId: targetUserID,
        baseUserVersion: parsedTargetBaseVersion,
        displayName: patchDisplayName,
        mfaRequired: patchMfaRequired,
        isActive: patchIsActive,
        isDeploymentAdmin: patchIsDeploymentAdmin,
      });
      if (!isCurrentTargetAdminOperation(operationID)) {
        return;
      }

      const nextError = extractError(result.payload);
      setError(nextError);
      if (!result.ok) {
        setStatusText("Patch local user failed");
        return;
      }

      const user = (result.payload as { data: UserResource }).data;
      applySelectedUser(user);
      setStatusText("Patched local user");
      setError(null);
      await onRefreshShell();
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
    setError(null);

    try {
      const result = await adminResetPassword({
        userId: targetUserID,
        baseUserVersion: parsedTargetBaseVersion,
        newPassword: adminNewPassword,
        reason: adminReason,
      });
      if (!isCurrentTargetAdminOperation(operationID)) {
        return;
      }

      const nextError = extractError(result.payload);
      setError(nextError);
      if (!result.ok) {
        setStatusText("Reset user password failed");
        return;
      }

      const user = (result.payload as { data: UserResource }).data;
      applySelectedUser(user);
      setStatusText("Reset user password");
      setError(null);
      await onRefreshShell();
    } finally {
      finishTargetAdminOperation(operationID);
    }
  }

  async function handleAdminTotpReset() {
    if (!canSubmitVersionedTargetAction || selectedUser === null) {
      return;
    }
    const operationID = beginTargetAdminOperation("mutating");
    if (operationID === null) {
      return;
    }
    const targetUserID = selectedUser.user_id;
    setStatusText("Resetting user TOTP");
    setError(null);

    try {
      const result = await adminResetTotp({
        userId: targetUserID,
        baseUserVersion: parsedTargetBaseVersion,
        reason: adminReason,
      });
      if (!isCurrentTargetAdminOperation(operationID)) {
        return;
      }

      const nextError = extractError(result.payload);
      setError(nextError);
      if (!result.ok) {
        setStatusText("Reset user TOTP failed");
        return;
      }

      const user = (result.payload as { data: UserResource }).data;
      applySelectedUser(user);
      setStatusText("Reset user TOTP");
      setError(null);
      await onRefreshShell();
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
    setError(null);

    try {
      const result = await adminRevokeAllSessions({
        userId: targetUserID,
        reason: adminReason,
      });
      if (!isCurrentTargetAdminOperation(operationID)) {
        return;
      }

      const nextError = extractError(result.payload);
      setError(nextError);
      if (!result.ok) {
        setStatusText("Revoke-all failed");
        return;
      }

      setStatusText("Revoked every user session");
      setError(null);
      await onRefreshShell();
    } finally {
      finishTargetAdminOperation(operationID);
    }
  }

  if (!session.is_deployment_admin) {
    return (
      <section style={cardStyle}>
        <div style={cardHeaderStyle}>
          <div>
            <p style={sectionEyebrowStyle}>Deployment users</p>
            <h2 style={sectionTitleStyle}>User administration</h2>
          </div>
        </div>
        <p data-testid="admin-access-note" style={bodyStyle}>
          Deployment admin access is required for user creation, patching, and
          credential actions. Incident-admin membership alone does not unlock
          these controls.
        </p>
      </section>
    );
  }

  return (
    <section style={cardStyle}>
      <div style={cardHeaderStyle}>
        <div>
          <p style={sectionEyebrowStyle}>Deployment users</p>
          <h2 style={sectionTitleStyle}>User administration</h2>
        </div>
      </div>

      <section style={subsectionStyle}>
        <p style={subsectionTitleStyle}>Create local user</p>
        <div style={formGridStyle}>
          <label htmlFor="admin-create-email" style={labelBlockStyle}>
            Email
          </label>
          <input
            data-testid="admin-create-email"
            id="admin-create-email"
            disabled={targetOperationPending}
            style={inputStyle}
            value={createEmail}
            onChange={(event) => {
              setCreateEmail(event.target.value);
            }}
          />
          <label htmlFor="admin-create-display-name" style={labelBlockStyle}>
            Display name
          </label>
          <input
            data-testid="admin-create-display-name"
            id="admin-create-display-name"
            disabled={targetOperationPending}
            style={inputStyle}
            value={createDisplayName}
            onChange={(event) => {
              setCreateDisplayName(event.target.value);
            }}
          />
          <label htmlFor="admin-create-password" style={labelBlockStyle}>
            Initial password
          </label>
          <input
            data-testid="admin-create-password"
            id="admin-create-password"
            disabled={targetOperationPending}
            style={inputStyle}
            type="password"
            value={createInitialPassword}
            onChange={(event) => {
              setCreateInitialPassword(event.target.value);
            }}
          />
        </div>
        <div style={checkboxRowStyle}>
          <label style={checkboxLabelStyle}>
            <input
              data-testid="admin-create-mfa-required"
              type="checkbox"
              disabled={targetOperationPending}
              checked={createMfaRequired}
              onChange={(event) => {
                setCreateMfaRequired(event.target.checked);
              }}
            />
            MFA required
          </label>
          <label style={checkboxLabelStyle}>
            <input
              data-testid="admin-create-is-deployment-admin"
              type="checkbox"
              disabled={targetOperationPending}
              checked={createIsDeploymentAdmin}
              onChange={(event) => {
                setCreateIsDeploymentAdmin(event.target.checked);
              }}
            />
            Deployment admin
          </label>
        </div>
        <div style={buttonRowStyle}>
          <button
            data-testid="admin-create-user"
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
      </section>

      <section style={subsectionStyle}>
        <p style={subsectionTitleStyle}>Load target user</p>
        <div style={formGridStyle}>
          <label htmlFor="admin-target-user-id-input" style={labelBlockStyle}>
            User id
          </label>
          <input
            data-testid="admin-target-user-id-input"
            id="admin-target-user-id-input"
            disabled={targetOperationPending}
            style={inputStyle}
            value={targetUserId}
            onChange={(event) => {
              setTargetUserId(event.target.value);
            }}
          />
        </div>
        <div style={buttonRowStyle}>
          <button
            data-testid="admin-load-user"
            disabled={!canLoadTargetUser}
            style={secondaryButtonStyle}
            type="button"
            onClick={() => {
              void loadSelectedUser(targetUserId);
            }}
          >
            Load user
          </button>
        </div>
        <div style={detailGridStyle}>
          <div>
            <span style={labelStyle}>Loaded user id</span>
            <div data-testid="admin-target-user-id" style={monoTextStyle}>
              {selectedUser?.user_id ?? ""}
            </div>
          </div>
          <div>
            <span style={labelStyle}>User version</span>
            <div data-testid="admin-target-user-version">
              {selectedUser?.user_version ?? ""}
            </div>
          </div>
          <div>
            <span style={labelStyle}>Is active</span>
            <div data-testid="admin-target-is-active">
              {selectedUser ? String(selectedUser.is_active) : ""}
            </div>
          </div>
          <div>
            <span style={labelStyle}>Deployment admin</span>
            <div data-testid="admin-target-is-deployment-admin">
              {selectedUser ? String(selectedUser.is_deployment_admin) : ""}
            </div>
          </div>
        </div>
      </section>

      <section style={subsectionStyle}>
        <p style={subsectionTitleStyle}>Patch target user</p>
        <div style={formGridStyle}>
          <label htmlFor="admin-patch-base-version" style={labelBlockStyle}>
            Base user version
          </label>
          <input
            data-testid="admin-patch-base-version"
            id="admin-patch-base-version"
            disabled={!canSubmitTargetAction}
            style={inputStyle}
            value={targetBaseVersion}
            onChange={(event) => {
              setTargetBaseVersion(event.target.value);
            }}
          />
          <label htmlFor="admin-patch-display-name" style={labelBlockStyle}>
            Display name
          </label>
          <input
            data-testid="admin-patch-display-name"
            id="admin-patch-display-name"
            disabled={!canSubmitTargetAction}
            style={inputStyle}
            value={patchDisplayName}
            onChange={(event) => {
              setPatchDisplayName(event.target.value);
            }}
          />
        </div>
        <div style={checkboxRowStyle}>
          <label style={checkboxLabelStyle}>
            <input
              data-testid="admin-patch-mfa-required"
              type="checkbox"
              disabled={!canSubmitTargetAction}
              checked={patchMfaRequired}
              onChange={(event) => {
                setPatchMfaRequired(event.target.checked);
              }}
            />
            MFA required
          </label>
          <label style={checkboxLabelStyle}>
            <input
              data-testid="admin-patch-is-active"
              type="checkbox"
              disabled={!canSubmitTargetAction}
              checked={patchIsActive}
              onChange={(event) => {
                setPatchIsActive(event.target.checked);
              }}
            />
            Active
          </label>
          <label style={checkboxLabelStyle}>
            <input
              data-testid="admin-patch-is-deployment-admin"
              type="checkbox"
              disabled={!canSubmitTargetAction}
              checked={patchIsDeploymentAdmin}
              onChange={(event) => {
                setPatchIsDeploymentAdmin(event.target.checked);
              }}
            />
            Deployment admin
          </label>
        </div>
        <div style={buttonRowStyle}>
          <button
            data-testid="admin-patch-user"
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
      </section>

      <section style={subsectionStyle}>
        <p style={subsectionTitleStyle}>Credential actions</p>
        <div style={formGridStyle}>
          <label htmlFor="admin-new-password" style={labelBlockStyle}>
            New password
          </label>
          <input
            data-testid="admin-new-password"
            id="admin-new-password"
            style={inputStyle}
            type="password"
            value={adminNewPassword}
            onChange={(event) => {
              setAdminNewPassword(event.target.value);
            }}
          />
          <label htmlFor="admin-reason" style={labelBlockStyle}>
            Reason
          </label>
          <input
            data-testid="admin-reason"
            id="admin-reason"
            style={inputStyle}
            value={adminReason}
            onChange={(event) => {
              setAdminReason(event.target.value);
            }}
          />
        </div>
        <div style={buttonRowStyle}>
          <button
            data-testid="admin-password-reset"
            disabled={!canSubmitVersionedTargetAction}
            style={buttonStyle}
            type="button"
            onClick={() => {
              void handleAdminPasswordReset();
            }}
          >
            Reset password
          </button>
          <button
            data-testid="admin-totp-reset"
            disabled={!canSubmitVersionedTargetAction}
            style={buttonStyle}
            type="button"
            onClick={() => {
              void handleAdminTotpReset();
            }}
          >
            Reset TOTP
          </button>
          <button
            data-testid="admin-revoke-all"
            disabled={!canSubmitTargetAction}
            style={buttonStyle}
            type="button"
            onClick={() => {
              void handleAdminRevokeAll();
            }}
          >
            Revoke all sessions
          </button>
        </div>
      </section>

      <p data-testid="admin-status" style={statusTextStyle}>
        {statusText}
      </p>
      <p data-testid="admin-error-code" style={errorStyle}>
        {error?.code ?? ""}
      </p>
    </section>
  );
}

const shellStyle: CSSProperties = {
  minHeight: "100vh",
  padding: "2rem",
};

const shellPanelStyle: CSSProperties = {
  width: "min(56rem, 100%)",
  margin: "0 auto",
  padding: "2rem",
  borderRadius: "1.5rem",
  background: "rgb(255 251 244 / 0.94)",
  boxShadow: "0 24px 80px rgb(29 78 70 / 0.12)",
  border: "1px solid rgb(185 204 196 / 0.8)",
};

const sectionHeaderStyle: CSSProperties = {
  display: "flex",
  justifyContent: "space-between",
  gap: "1rem",
  alignItems: "flex-start",
  marginBottom: "1.5rem",
};

const eyebrowStyle: CSSProperties = {
  margin: 0,
  letterSpacing: "0.14em",
  textTransform: "uppercase",
  fontSize: "0.78rem",
  color: "rgb(75 108 100)",
};

const headlineStyle: CSSProperties = {
  margin: "0.35rem 0 0",
  fontSize: "clamp(2rem, 3vw, 2.8rem)",
  lineHeight: 1.05,
};

const bodyStyle: CSSProperties = {
  margin: "0.75rem 0 0",
  color: "rgb(73 95 90)",
  maxWidth: "42rem",
};

const statusCardStyle: CSSProperties = {
  minWidth: "14rem",
  padding: "0.9rem 1rem",
  borderRadius: "1rem",
  background: "rgb(239 245 240)",
  border: "1px solid rgb(197 214 206)",
};

const cardStyle: CSSProperties = {
  padding: "1.4rem",
  borderRadius: "1.1rem",
  background: "rgb(255 255 255 / 0.92)",
  border: "1px solid rgb(211 223 216)",
};

const cardHeaderStyle: CSSProperties = {
  display: "flex",
  justifyContent: "space-between",
  gap: "1rem",
  alignItems: "center",
  marginBottom: "1rem",
};

const sectionEyebrowStyle: CSSProperties = {
  margin: 0,
  fontSize: "0.72rem",
  letterSpacing: "0.14em",
  textTransform: "uppercase",
  color: "rgb(75 108 100)",
};

const sectionTitleStyle: CSSProperties = {
  margin: "0.35rem 0 0",
  fontSize: "1.15rem",
};

const subsectionStyle: CSSProperties = {
  marginTop: "1.5rem",
  paddingTop: "1.25rem",
  borderTop: "1px solid rgb(222 230 225)",
};

const subsectionTitleStyle: CSSProperties = {
  margin: 0,
  fontWeight: 700,
  color: "rgb(45 82 75)",
};

const formGridStyle: CSSProperties = {
  display: "grid",
  gridTemplateColumns: "1fr",
  gap: "0.75rem",
  marginTop: "1rem",
};

const detailGridStyle: CSSProperties = {
  display: "grid",
  gridTemplateColumns: "repeat(auto-fit, minmax(12rem, 1fr))",
  gap: "0.85rem",
  marginTop: "1rem",
};

const labelStyle: CSSProperties = {
  display: "block",
  fontSize: "0.72rem",
  letterSpacing: "0.12em",
  textTransform: "uppercase",
  color: "rgb(75 108 100)",
  marginBottom: "0.35rem",
};

const labelBlockStyle: CSSProperties = {
  fontSize: "0.84rem",
  fontWeight: 600,
  color: "rgb(45 82 75)",
};

const inputStyle: CSSProperties = {
  width: "100%",
  marginTop: "0.35rem",
  padding: "0.72rem 0.85rem",
  borderRadius: "0.9rem",
  border: "1px solid rgb(191 207 199)",
  background: "rgb(255 255 255)",
  fontSize: "0.95rem",
};

const buttonRowStyle: CSSProperties = {
  display: "flex",
  flexWrap: "wrap",
  gap: "0.75rem",
  marginTop: "1rem",
};

const buttonStyle: CSSProperties = {
  padding: "0.72rem 1rem",
  borderRadius: "999px",
  border: "none",
  background: "rgb(30 98 86)",
  color: "white",
  fontWeight: 700,
  cursor: "pointer",
};

const secondaryButtonStyle: CSSProperties = {
  ...buttonStyle,
  background: "rgb(232 240 235)",
  color: "rgb(30 74 67)",
  border: "1px solid rgb(197 214 206)",
};

const monoTextStyle: CSSProperties = {
  fontFamily: '"IBM Plex Mono", "SFMono-Regular", Consolas, monospace',
  overflowWrap: "anywhere",
};

const plainListStyle: CSSProperties = {
  margin: 0,
  paddingLeft: "1rem",
};

const wideCellStyle: CSSProperties = {
  gridColumn: "1 / -1",
};

const checkboxRowStyle: CSSProperties = {
  display: "flex",
  flexWrap: "wrap",
  gap: "1rem",
  marginTop: "1rem",
};

const checkboxLabelStyle: CSSProperties = {
  display: "inline-flex",
  alignItems: "center",
  gap: "0.45rem",
  color: "rgb(45 82 75)",
};

const statusTextStyle: CSSProperties = {
  margin: "1rem 0 0",
  minHeight: "1.5rem",
  color: "rgb(45 82 75)",
};

const errorStyle: CSSProperties = {
  margin: "0.25rem 0 0",
  minHeight: "1.25rem",
  color: "rgb(147 47 47)",
  fontWeight: 600,
};
