import {
  deploymentUserRowTestId,
  phase1AccountTestId,
  phase1AdminTestId,
  phase1AuthTestId,
  phase1ErrorCodeTestId,
  phase1ErrorSummaryTestIds,
} from "@cartulary/ui-contracts";
import { X } from "lucide-react";
import {
  type CSSProperties,
  forwardRef,
  useCallback,
  useEffect,
  useImperativeHandle,
  useRef,
  useState,
} from "react";

import {
  type APIError,
  extractError,
  publicErrorView,
} from "../services/browserApi";
import {
  adminResetPassword,
  adminResetTotp,
  adminRevokeAllSessions,
  beginEnterpriseAuth,
  beginTotpEnrollment,
  type CredentialState,
  changePassword,
  completeTotpEnrollment,
  createEnterpriseAuthBinding,
  createLocalUser,
  type EnterpriseAuthProvider,
  listEnterpriseAuthProviders,
  listUsers,
  loadUser,
  loginLocal,
  logoutCurrentSession,
  patchLocalUser,
  retireEnterpriseAuthBinding,
  rotateEnterpriseAuthBinding,
  type SessionData,
  type UserListEnvelope,
  type UserResource,
} from "./phase1Client";

type RefreshOptions = {
  anonymousMessage?: string;
};

type AuthSurfaceBootstrapState =
  | "loading"
  | "anonymous"
  | "revoked"
  | "public_error_envelope";

type Phase1AuthSurfaceProps = {
  bootstrapState: AuthSurfaceBootstrapState;
  message: string;
  onAuthenticated: () => Promise<void> | void;
  publicError?: APIError | null;
  readingProfile?: "default" | "hyperlegible" | undefined;
};

type Phase1AccountPanelProps = {
  credentialState: CredentialState | null;
  credentialStateError: APIError | null;
  onRefreshShell: (options?: RefreshOptions) => Promise<void> | void;
  session: SessionData;
};

type Phase1AdminPanelProps = {
  autoLoadUsers?: boolean | undefined;
  enterpriseAuthClaimed?: boolean | undefined;
  onCommandStateChange?:
    | ((state: Phase1AdminPanelCommandState) => void)
    | undefined;
  onRefreshShell: (options?: RefreshOptions) => Promise<void> | void;
  session: SessionData;
};

export type Phase1AccountPanelHandle = {
  refreshAccount: () => Promise<void>;
  signOut: () => Promise<void>;
};

export type Phase1AdminPanelHandle = {
  createUser: () => Promise<void>;
  loadTargetUser: () => Promise<void>;
  refreshUsers: () => Promise<void>;
  resetPassword: () => Promise<void>;
  resetTotp: () => Promise<void>;
  revokeAllSessions: () => Promise<void>;
  saveTargetUser: () => Promise<void>;
};

export type Phase1AdminPanelCommandState = {
  canLoadTargetUser: boolean;
  canSubmitTargetAction: boolean;
  canSubmitVersionedTargetAction: boolean;
  hasSelectedUser: boolean;
  targetOperationPending: boolean;
};

type TargetAdminOperation = "loading" | "mutating";
type CredentialDialogKind = "password" | "revoke" | "totp";

type AuthChallengeState = "mfa_required" | "mfa_setup_required";

type SafeAuthBindingSummary = NonNullable<
  UserResource["auth_bindings"]
>[number];
type EnterpriseAuthBindingSummary = Extract<
  SafeAuthBindingSummary,
  { provider_type: "oidc" | "saml" }
>;

function isEnterpriseAuthBinding(
  binding: SafeAuthBindingSummary,
): binding is EnterpriseAuthBindingSummary {
  return binding.provider_type !== "local";
}

let enterpriseAuthNavigate = (redirectURL: string) => {
  window.location.assign(redirectURL);
};

export function setEnterpriseAuthNavigateForTesting(
  navigate: (redirectURL: string) => void,
) {
  const previous = enterpriseAuthNavigate;
  enterpriseAuthNavigate = navigate;
  return () => {
    enterpriseAuthNavigate = previous;
  };
}

export function Phase1AuthSurface({
  bootstrapState,
  message,
  onAuthenticated,
  publicError = null,
  readingProfile = "default",
}: Phase1AuthSurfaceProps) {
  const [statusText, setStatusText] = useState("Ready to sign in.");
  const [error, setError] = useState<APIError | null>(null);
  const [authChallengeState, setAuthChallengeState] =
    useState<AuthChallengeState | null>(null);
  const [bootstrapToken, setBootstrapToken] = useState("");
  const [bootstrapEnrollmentId, setBootstrapEnrollmentId] = useState("");
  const [bootstrapSecretBase32, setBootstrapSecretBase32] = useState("");
  const [enterpriseProviders, setEnterpriseProviders] = useState<
    EnterpriseAuthProvider[]
  >([]);

  const [username, setUsername] = useState("bootstrap-admin@example.test");
  const [password, setPassword] = useState("BootstrapPass1!");
  const [totpCode, setTotpCode] = useState("");
  const [bootstrapCompleteCode, setBootstrapCompleteCode] = useState("");

  useEffect(() => {
    if (bootstrapState === "loading") {
      return;
    }
    const controller = new AbortController();
    void (async () => {
      const result = await listEnterpriseAuthProviders({
        signal: controller.signal,
      });
      if (controller.signal.aborted) {
        return;
      }
      const nextError = extractError(result.payload);
      if (!result.ok) {
        if (
          result.status === 404 ||
          nextError?.code === "extension_profile_not_claimed"
        ) {
          setEnterpriseProviders([]);
          return;
        }
        setEnterpriseProviders([]);
        return;
      }
      const data = (
        result.payload as { data: { providers: EnterpriseAuthProvider[] } }
      ).data;
      setEnterpriseProviders(data.providers);
    })();
    return () => {
      controller.abort();
    };
  }, [bootstrapState]);

  async function handleLogin() {
    setStatusText("Signing in");
    setAuthChallengeState(null);
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
        setBootstrapEnrollmentId("");
        setBootstrapSecretBase32("");
        setAuthChallengeState("mfa_setup_required");
        setStatusText("TOTP enrollment is required before sign-in.");
        return;
      }
      if (nextError?.code === "mfa_required") {
        setBootstrapToken("");
        setBootstrapEnrollmentId("");
        setBootstrapSecretBase32("");
        setAuthChallengeState("mfa_required");
        setStatusText("TOTP code is required before sign-in.");
        return;
      }
      setBootstrapToken("");
      setBootstrapEnrollmentId("");
      setBootstrapSecretBase32("");
      setStatusText("Sign in failed");
      return;
    }

    setBootstrapToken("");
    setBootstrapEnrollmentId("");
    setBootstrapSecretBase32("");
    setAuthChallengeState(null);
    setError(null);
    setStatusText("Signed in");
    await onAuthenticated();
  }

  async function handleEnterpriseBegin(providerKey: string) {
    setStatusText("Starting enterprise sign-in");
    const returnTo =
      `${window.location.pathname}${window.location.search}`.trim() || "/";
    const result = await beginEnterpriseAuth({
      providerKey,
      returnTo,
    });
    const nextError = extractError(result.payload);
    setError(nextError);
    if (!result.ok) {
      setStatusText("Enterprise sign-in failed");
      return;
    }
    const data = (result.payload as { data: { redirect_url: string } }).data;
    enterpriseAuthNavigate(data.redirect_url);
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
    setAuthChallengeState(null);
    setError(null);
    setStatusText("TOTP enrollment completed. Sign in with your TOTP code.");
  }

  const displayedError = error ?? publicError;
  const displayedBootstrapState = authChallengeState ?? bootstrapState;
  const rootShellStyle =
    readingProfile === "hyperlegible"
      ? {
          ...shellStyle,
          fontFamily: "var(--ct-typography-accessible-reading-fontFamily)",
        }
      : shellStyle;

  return (
    <main
      aria-busy={bootstrapState === "loading"}
      className="cartulary-shell"
      data-bootstrap-state={displayedBootstrapState}
      data-reading-profile={
        readingProfile === "hyperlegible" ? "hyperlegible" : undefined
      }
      data-testid={phase1AuthTestId("shell")}
      style={rootShellStyle}
    >
      <section style={shellPanelStyle}>
        <header style={sectionHeaderStyle}>
          <div>
            <p style={eyebrowStyle}>Cartulary</p>
            <h1 style={headlineStyle}>Local sign-in</h1>
            <p
              data-testid={phase1AuthTestId("shell-message")}
              style={bodyStyle}
            >
              {message}
            </p>
          </div>
          <div style={statusCardStyle}>
            <span style={labelStyle}>Status</span>
            <strong
              aria-live="polite"
              data-testid={phase1AuthTestId("status")}
              role="status"
            >
              {bootstrapState === "loading"
                ? "Checking current session…"
                : statusText}
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
              data-testid={phase1AuthTestId("login-username")}
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
              data-testid={phase1AuthTestId("login-password")}
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
              data-testid={phase1AuthTestId("login-totp-code")}
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
              data-testid={phase1AuthTestId("login-submit")}
              style={buttonStyle}
              type="button"
              onClick={() => {
                void handleLogin();
              }}
            >
              Sign in
            </button>
          </div>
          <p
            aria-live="assertive"
            data-testid={phase1ErrorCodeTestId("auth")}
            role={displayedError === null ? undefined : "alert"}
            style={errorStyle}
          >
            {publicErrorView(displayedError)?.code ?? ""}
          </p>
          <PublicErrorSummary
            error={displayedError}
            testIds={phase1ErrorSummaryTestIds("auth")}
          />
        </section>

        {enterpriseProviders.length > 0 ? (
          <section style={cardStyle}>
            <div style={cardHeaderStyle}>
              <div>
                <p style={sectionEyebrowStyle}>Enterprise</p>
                <h2 style={sectionTitleStyle}>Provider sign-in</h2>
              </div>
            </div>
            <div
              data-testid={phase1AuthTestId("enterprise-provider-list")}
              style={buttonRowStyle}
            >
              {enterpriseProviders.map((provider) => (
                <button
                  key={provider.provider_key}
                  data-provider-key={provider.provider_key}
                  data-testid={phase1AuthTestId("enterprise-provider-button")}
                  style={secondaryButtonStyle}
                  type="button"
                  onClick={() => {
                    void handleEnterpriseBegin(provider.provider_key);
                  }}
                >
                  {provider.display_name}
                </button>
              ))}
            </div>
          </section>
        ) : null}

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
                <span style={labelStyle}>Setup token</span>
                <div
                  data-testid={phase1AuthTestId("bootstrap-token")}
                  style={monoTextStyle}
                >
                  Stored for TOTP setup requests.
                </div>
              </div>
              <div>
                <span style={labelStyle}>Enrollment id</span>
                <div
                  data-testid={phase1AuthTestId("bootstrap-enrollment-id")}
                  style={monoTextStyle}
                >
                  {bootstrapEnrollmentId}
                </div>
              </div>
              <div style={wideCellStyle}>
                <span style={labelStyle}>Secret base32</span>
                <div
                  data-testid={phase1AuthTestId("bootstrap-secret-base32")}
                  style={monoTextStyle}
                >
                  {bootstrapSecretBase32}
                </div>
              </div>
            </div>
            <div style={buttonRowStyle}>
              <button
                data-testid={phase1AuthTestId("bootstrap-begin")}
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
                data-testid={phase1AuthTestId("bootstrap-complete-code")}
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
                data-testid={phase1AuthTestId("bootstrap-complete")}
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

export const Phase1AccountPanel = forwardRef<
  Phase1AccountPanelHandle,
  Phase1AccountPanelProps
>(function Phase1AccountPanel(
  { credentialState, credentialStateError, onRefreshShell, session },
  ref,
) {
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

  useImperativeHandle(ref, () => ({
    refreshAccount: handleRefreshAccount,
    signOut: handleLogout,
  }));

  return (
    <section style={cardStyle}>
      <div style={cardHeaderStyle}>
        <div>
          <p style={sectionEyebrowStyle}>Account</p>
          <h2 style={sectionTitleStyle}>Session and credential security</h2>
        </div>
        <div style={buttonRowStyle}>
          <button
            data-testid={phase1AccountTestId("refresh-state")}
            style={secondaryButtonStyle}
            type="button"
            onClick={() => {
              void handleRefreshAccount();
            }}
          >
            Refresh
          </button>
          <button
            data-testid={phase1AccountTestId("logout")}
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
          <div
            data-testid={phase1AccountTestId("session-user-id")}
            style={monoTextStyle}
          >
            {session.user_id}
          </div>
        </div>
        <div>
          <span style={labelStyle}>Provider</span>
          <div data-testid={phase1AccountTestId("session-provider-type")}>
            {session.provider_type}
          </div>
        </div>
        <div>
          <span style={labelStyle}>MFA state</span>
          <div data-testid={phase1AccountTestId("session-mfa-state")}>
            {session.mfa_state}
          </div>
        </div>
        <div>
          <span style={labelStyle}>Deployment admin</span>
          <div data-testid={phase1AccountTestId("session-is-deployment-admin")}>
            {String(session.is_deployment_admin)}
          </div>
        </div>
        <div>
          <span style={labelStyle}>Authenticated at</span>
          <div data-testid={phase1AccountTestId("session-authenticated-at")}>
            {session.authenticated_at}
          </div>
        </div>
        <div>
          <span style={labelStyle}>Idle expires at</span>
          <div data-testid={phase1AccountTestId("session-idle-expires-at")}>
            {session.idle_expires_at}
          </div>
        </div>
        <div>
          <span style={labelStyle}>Absolute expires at</span>
          <div data-testid={phase1AccountTestId("session-absolute-expires-at")}>
            {session.absolute_expires_at}
          </div>
        </div>
        <div>
          <span style={labelStyle}>Session expires at</span>
          <div data-testid={phase1AccountTestId("session-session-expires-at")}>
            {session.session_expires_at}
          </div>
        </div>
        <div style={wideCellStyle}>
          <span style={labelStyle}>Incident memberships</span>
          <ul
            data-testid={phase1AccountTestId("session-memberships")}
            style={plainListStyle}
          >
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
          <div data-testid={phase1AccountTestId("credential-auth-kind")}>
            {credentialState?.auth_kind ?? ""}
          </div>
        </div>
        <div>
          <span style={labelStyle}>Recovery model</span>
          <div data-testid={phase1AccountTestId("credential-recovery-model")}>
            {credentialState?.recovery_model ?? ""}
          </div>
        </div>
        <div>
          <span style={labelStyle}>Password changed at</span>
          <div
            data-testid={phase1AccountTestId("credential-password-changed-at")}
          >
            {credentialState?.password_changed_at ?? ""}
          </div>
        </div>
        <div>
          <span style={labelStyle}>TOTP state</span>
          <div data-testid={phase1AccountTestId("credential-totp-state")}>
            {credentialState?.totp.state ?? ""}
          </div>
        </div>
        <div style={wideCellStyle}>
          <span style={labelStyle}>Pending expires at</span>
          <div
            data-testid={phase1AccountTestId("credential-pending-expires-at")}
          >
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
            data-testid={phase1AccountTestId("password-current")}
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
            data-testid={phase1AccountTestId("password-next")}
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
            data-testid={phase1AccountTestId("password-factor-code")}
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
            data-testid={phase1AccountTestId("password-change")}
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
            data-testid={phase1AccountTestId("totp-current-password")}
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
            data-testid={phase1AccountTestId("totp-current-factor")}
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
            data-testid={phase1AccountTestId("totp-begin")}
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
            <div
              data-testid={phase1AccountTestId("totp-enrollment-id")}
              style={monoTextStyle}
            >
              {totpEnrollmentId}
            </div>
          </div>
          <div style={wideCellStyle}>
            <span style={labelStyle}>Secret base32</span>
            <div
              data-testid={phase1AccountTestId("totp-secret-base32")}
              style={monoTextStyle}
            >
              {totpSecretBase32}
            </div>
          </div>
        </div>
        <div style={formGridStyle}>
          <label htmlFor="account-totp-complete-code" style={labelBlockStyle}>
            Replacement TOTP code
          </label>
          <input
            data-testid={phase1AccountTestId("totp-complete-code")}
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
            data-testid={phase1AccountTestId("totp-complete")}
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

      <p
        aria-live="polite"
        data-testid={phase1AccountTestId("status")}
        role="status"
        style={statusTextStyle}
      >
        {statusText}
      </p>
      <p
        aria-live="assertive"
        data-testid={phase1ErrorCodeTestId("account")}
        role={
          error === null && credentialStateError === null ? undefined : "alert"
        }
        style={errorStyle}
      >
        {publicErrorView(error ?? credentialStateError)?.code ?? ""}
      </p>
      <PublicErrorSummary
        error={error ?? credentialStateError}
        testIds={phase1ErrorSummaryTestIds("account")}
      />
    </section>
  );
});

function upsertUserResource(
  users: UserResource[],
  nextUser: UserResource,
): UserResource[] {
  const existingIndex = users.findIndex(
    (user) => user.user_id === nextUser.user_id,
  );
  if (existingIndex === -1) {
    return [...users, nextUser].sort((a, b) =>
      a.user_id.localeCompare(b.user_id),
    );
  }
  const next = [...users];
  next[existingIndex] = nextUser;
  return next;
}

function mergeUserResources(
  users: UserResource[],
  nextUsers: UserResource[],
): UserResource[] {
  return nextUsers.reduce(upsertUserResource, users);
}

export const Phase1AdminPanel = forwardRef<
  Phase1AdminPanelHandle,
  Phase1AdminPanelProps
>(function Phase1AdminPanel(
  {
    autoLoadUsers = false,
    enterpriseAuthClaimed = false,
    onCommandStateChange,
    onRefreshShell,
    session,
  },
  ref,
) {
  const [statusText, setStatusText] = useState(
    session.is_deployment_admin
      ? "Deployment user administration is ready."
      : "Deployment admin access is required for user administration.",
  );
  const [error, setError] = useState<APIError | null>(null);
  const [selectedUser, setSelectedUser] = useState<UserResource | null>(null);
  const [users, setUsers] = useState<UserResource[]>([]);
  const [userFilter, setUserFilter] = useState("");
  const [userActiveFilter, setUserActiveFilter] = useState("all");
  const [userAdminFilter, setUserAdminFilter] = useState("all");
  const [usersNextCursor, setUsersNextCursor] = useState<string | null>(null);
  const [usersHasMore, setUsersHasMore] = useState(false);

  const [createEmail, setCreateEmail] = useState("");
  const [createDisplayName, setCreateDisplayName] = useState("");
  const [createInitialPassword, setCreateInitialPassword] = useState("");
  const [createMfaRequired, setCreateMfaRequired] = useState(true);
  const [createIsDeploymentAdmin, setCreateIsDeploymentAdmin] = useState(false);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);

  const [targetUserId, setTargetUserId] = useState("");
  const [targetBaseVersion, setTargetBaseVersion] = useState("");
  const [patchEmail, setPatchEmail] = useState("");
  const [patchDisplayName, setPatchDisplayName] = useState("");
  const [patchMfaRequired, setPatchMfaRequired] = useState(true);
  const [patchIsActive, setPatchIsActive] = useState(true);
  const [patchIsDeploymentAdmin, setPatchIsDeploymentAdmin] = useState(false);
  const [adminNewPassword, setAdminNewPassword] = useState("");
  const [adminReason, setAdminReason] = useState("");
  const [credentialDialog, setCredentialDialog] =
    useState<CredentialDialogKind | null>(null);
  const [enterpriseProviders, setEnterpriseProviders] = useState<
    EnterpriseAuthProvider[]
  >([]);
  const [bindingProviderKey, setBindingProviderKey] = useState("");
  const [bindingProviderSubject, setBindingProviderSubject] = useState("");
  const [bindingTargetID, setBindingTargetID] = useState("");
  const [bindingNewSubject, setBindingNewSubject] = useState("");
  const [bindingReason, setBindingReason] = useState("");
  const [targetAdminOperation, setTargetAdminOperation] =
    useState<TargetAdminOperation | null>(null);
  const targetAdminOperationRef = useRef<{
    id: number;
    kind: TargetAdminOperation;
  } | null>(null);
  const nextTargetAdminOperationID = useRef(0);
  const userListRequestIDRef = useRef(0);
  const acceptedUserQueryRef = useRef({
    isActive: "all",
    isDeploymentAdmin: "all",
    search: "",
  });
  const userFiltersTouchedRef = useRef(false);
  const autoLoadUsersStartedRef = useRef(false);

  const targetOperationPending = targetAdminOperation !== null;
  const loadedTargetIsCurrent = selectedUser !== null;
  const parsedTargetBaseVersion = selectedUser?.user_version ?? 0;
  const canLoadTargetUser =
    !targetOperationPending && targetUserId.trim() !== "";
  const canSubmitTargetAction =
    !targetOperationPending && loadedTargetIsCurrent;
  const canSubmitVersionedTargetAction =
    canSubmitTargetAction && selectedUser !== null;

  useEffect(() => {
    onCommandStateChange?.({
      canLoadTargetUser,
      canSubmitTargetAction,
      canSubmitVersionedTargetAction,
      hasSelectedUser: selectedUser !== null,
      targetOperationPending,
    });
  }, [
    canLoadTargetUser,
    canSubmitTargetAction,
    canSubmitVersionedTargetAction,
    onCommandStateChange,
    selectedUser,
    targetOperationPending,
  ]);

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
    setPatchEmail("");
    setPatchDisplayName("");
    setPatchMfaRequired(true);
    setPatchIsActive(true);
    setPatchIsDeploymentAdmin(false);
    setAdminNewPassword("");
    setAdminReason("");
    setCredentialDialog(null);
    setBindingTargetID("");
    setBindingNewSubject("");
  }

  function applySelectedUser(user: UserResource) {
    const firstEnterpriseBinding = user.auth_bindings?.find(
      isEnterpriseAuthBinding,
    );
    setSelectedUser(user);
    setUsers((current) => upsertUserResource(current, user));
    setTargetUserId(user.user_id);
    setTargetBaseVersion(String(user.user_version));
    setPatchEmail(user.email);
    setPatchDisplayName(user.display_name);
    setPatchMfaRequired(user.mfa_required);
    setPatchIsActive(user.is_active);
    setPatchIsDeploymentAdmin(user.is_deployment_admin);
    setBindingTargetID(firstEnterpriseBinding?.auth_binding_id ?? "");
    setBindingNewSubject("");
    setAdminNewPassword("");
    setAdminReason("");
    setCredentialDialog(null);
  }

  const refreshUsers = useCallback(async () => {
    if (!session.is_deployment_admin) {
      return;
    }
    const requestID = userListRequestIDRef.current + 1;
    userListRequestIDRef.current = requestID;
    const filterActive =
      userFilter.trim() !== "" ||
      userActiveFilter !== "all" ||
      userAdminFilter !== "all";
    setStatusText(
      filterActive
        ? "Searching deployment users"
        : "Refreshing deployment users",
    );
    setError(null);
    const result = await listUsers({
      limit: 100,
      search: userFilter,
      isActive: userActiveFilter === "all" ? null : userActiveFilter === "true",
      isDeploymentAdmin:
        userAdminFilter === "all" ? null : userAdminFilter === "true",
    });
    if (userListRequestIDRef.current !== requestID) {
      return;
    }
    const nextError = extractError(result.payload);
    setError(nextError);
    if (!result.ok) {
      setStatusText("Deployment users unavailable");
      return;
    }
    const payload = result.payload as UserListEnvelope;
    setUsers(payload.data.users);
    setUsersNextCursor(payload.meta.paging.next_cursor);
    setUsersHasMore(payload.meta.paging.has_more);
    acceptedUserQueryRef.current = {
      isActive: userActiveFilter,
      isDeploymentAdmin: userAdminFilter,
      search: userFilter.trim(),
    };
    userFiltersTouchedRef.current = false;
    setStatusText("Deployment users loaded");
  }, [
    session.is_deployment_admin,
    userActiveFilter,
    userAdminFilter,
    userFilter,
  ]);

  const loadMoreUsers = useCallback(async () => {
    if (!session.is_deployment_admin || !usersHasMore) {
      return;
    }
    setStatusText("Loading more deployment users");
    setError(null);
    const acceptedQuery = acceptedUserQueryRef.current;
    const result = await listUsers({
      cursorToken: usersNextCursor,
      limit: 100,
      search: acceptedQuery.search,
      isActive:
        acceptedQuery.isActive === "all"
          ? null
          : acceptedQuery.isActive === "true",
      isDeploymentAdmin:
        acceptedQuery.isDeploymentAdmin === "all"
          ? null
          : acceptedQuery.isDeploymentAdmin === "true",
    });
    const nextError = extractError(result.payload);
    setError(nextError);
    if (!result.ok) {
      setStatusText("Load more users failed");
      return;
    }
    const payload = result.payload as UserListEnvelope;
    setUsers((current) => mergeUserResources(current, payload.data.users));
    setUsersNextCursor(payload.meta.paging.next_cursor);
    setUsersHasMore(payload.meta.paging.has_more);
    setStatusText("Loaded more deployment users");
  }, [session.is_deployment_admin, usersHasMore, usersNextCursor]);

  useEffect(() => {
    if (!autoLoadUsers || !session.is_deployment_admin) {
      autoLoadUsersStartedRef.current = false;
      return;
    }
    if (autoLoadUsersStartedRef.current) {
      return;
    }
    autoLoadUsersStartedRef.current = true;
    void refreshUsers();
  }, [autoLoadUsers, refreshUsers, session.is_deployment_admin]);

  useEffect(() => {
    if (
      !autoLoadUsers ||
      !session.is_deployment_admin ||
      !userFiltersTouchedRef.current
    ) {
      return;
    }
    const timeout = window.setTimeout(() => {
      void refreshUsers();
    }, 180);
    return () => {
      window.clearTimeout(timeout);
    };
  }, [autoLoadUsers, refreshUsers, session.is_deployment_admin]);

  useEffect(() => {
    if (!session.is_deployment_admin || !enterpriseAuthClaimed) {
      setEnterpriseProviders([]);
      return;
    }
    const controller = new AbortController();
    void (async () => {
      const result = await listEnterpriseAuthProviders({
        signal: controller.signal,
      });
      if (controller.signal.aborted) {
        return;
      }
      if (!result.ok) {
        setEnterpriseProviders([]);
        return;
      }
      const data = (
        result.payload as { data: { providers: EnterpriseAuthProvider[] } }
      ).data;
      setEnterpriseProviders(data.providers);
    })();
    return () => {
      controller.abort();
    };
  }, [enterpriseAuthClaimed, session.is_deployment_admin]);

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
      setCreateInitialPassword("");
      if (!result.ok) {
        setStatusText("Create local user failed");
        return;
      }

      const user = (result.payload as { data: UserResource }).data;
      applySelectedUser(user);
      setStatusText("Created local user");
      setError(null);
      setCreateDialogOpen(false);
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
        email: patchEmail,
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
      setAdminNewPassword("");
      if (!result.ok) {
        setStatusText("Reset user password failed");
        return;
      }

      const user = (result.payload as { data: UserResource }).data;
      applySelectedUser(user);
      setStatusText("Reset user password");
      setError(null);
      setAdminReason("");
      setCredentialDialog(null);
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
      setAdminNewPassword("");
      setAdminReason("");
      setCredentialDialog(null);
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
      setAdminNewPassword("");
      setAdminReason("");
      setCredentialDialog(null);
      await onRefreshShell();
    } finally {
      finishTargetAdminOperation(operationID);
    }
  }

  async function handleCreateEnterpriseBinding() {
    if (!canSubmitVersionedTargetAction || selectedUser === null) {
      return;
    }
    const providerKey = bindingProviderKey.trim();
    const providerSubject = bindingProviderSubject.trim();
    if (providerKey === "" || providerSubject === "") {
      setStatusText("Provider key and provider subject are required.");
      return;
    }
    const operationID = beginTargetAdminOperation("mutating");
    if (operationID === null) {
      return;
    }
    setStatusText("Creating enterprise auth binding");
    setError(null);

    try {
      const result = await createEnterpriseAuthBinding({
        userId: selectedUser.user_id,
        baseUserVersion: parsedTargetBaseVersion,
        providerKey,
        providerSubject,
        reason: bindingReason,
      });
      if (!isCurrentTargetAdminOperation(operationID)) {
        return;
      }

      const nextError = extractError(result.payload);
      setError(nextError);
      if (!result.ok) {
        setStatusText("Create enterprise auth binding failed");
        return;
      }

      const user = (result.payload as { data: UserResource }).data;
      applySelectedUser(user);
      setBindingProviderSubject("");
      setStatusText("Created enterprise auth binding");
      setError(null);
      await onRefreshShell();
    } finally {
      finishTargetAdminOperation(operationID);
    }
  }

  async function handleRotateEnterpriseBinding() {
    if (
      !canSubmitVersionedTargetAction ||
      selectedUser === null ||
      bindingTargetID.trim() === ""
    ) {
      return;
    }
    const newProviderSubject = bindingNewSubject.trim();
    if (newProviderSubject === "") {
      setStatusText("New provider subject is required.");
      return;
    }
    const operationID = beginTargetAdminOperation("mutating");
    if (operationID === null) {
      return;
    }
    setStatusText("Rotating enterprise auth binding");
    setError(null);

    try {
      const result = await rotateEnterpriseAuthBinding({
        userId: selectedUser.user_id,
        authBindingId: bindingTargetID.trim(),
        baseUserVersion: parsedTargetBaseVersion,
        newProviderSubject,
        reason: bindingReason,
      });
      if (!isCurrentTargetAdminOperation(operationID)) {
        return;
      }

      const nextError = extractError(result.payload);
      setError(nextError);
      if (!result.ok) {
        setStatusText("Rotate enterprise auth binding failed");
        return;
      }

      const user = (result.payload as { data: UserResource }).data;
      applySelectedUser(user);
      setStatusText("Rotated enterprise auth binding");
      setError(null);
      await onRefreshShell();
    } finally {
      finishTargetAdminOperation(operationID);
    }
  }

  async function handleRetireEnterpriseBinding() {
    if (
      !canSubmitVersionedTargetAction ||
      selectedUser === null ||
      bindingTargetID.trim() === ""
    ) {
      return;
    }
    const operationID = beginTargetAdminOperation("mutating");
    if (operationID === null) {
      return;
    }
    setStatusText("Retiring enterprise auth binding");
    setError(null);

    try {
      const result = await retireEnterpriseAuthBinding({
        userId: selectedUser.user_id,
        authBindingId: bindingTargetID.trim(),
        baseUserVersion: parsedTargetBaseVersion,
        reason: bindingReason,
      });
      if (!isCurrentTargetAdminOperation(operationID)) {
        return;
      }

      const nextError = extractError(result.payload);
      setError(nextError);
      if (!result.ok) {
        setStatusText("Retire enterprise auth binding failed");
        return;
      }

      const user = (result.payload as { data: UserResource }).data;
      applySelectedUser(user);
      setStatusText("Retired enterprise auth binding");
      setError(null);
      await onRefreshShell();
    } finally {
      finishTargetAdminOperation(operationID);
    }
  }

  useImperativeHandle(ref, () => ({
    createUser: handleCreateUser,
    loadTargetUser: () => loadSelectedUser(targetUserId),
    refreshUsers,
    resetPassword: handleAdminPasswordReset,
    resetTotp: handleAdminTotpReset,
    revokeAllSessions: handleAdminRevokeAll,
    saveTargetUser: handlePatchUser,
  }));

  if (!session.is_deployment_admin) {
    return (
      <section style={cardStyle}>
        <div style={cardHeaderStyle}>
          <div>
            <p style={sectionEyebrowStyle}>Deployment users</p>
            <h2 style={sectionTitleStyle}>User administration</h2>
          </div>
        </div>
        <p data-testid={phase1AdminTestId("access-note")} style={bodyStyle}>
          Deployment admin access is required for user creation, patching, and
          credential actions. Incident-admin membership alone does not unlock
          these controls.
        </p>
      </section>
    );
  }

  const selectedEnterpriseBindings =
    selectedUser?.auth_bindings?.filter(isEnterpriseAuthBinding) ?? [];
  const selectedLocalBindings =
    selectedUser?.auth_bindings?.filter(
      (binding) => binding.provider_type === "local",
    ) ?? [];

  return (
    <section style={cardStyle}>
      <div style={cardHeaderStyle}>
        <div>
          <p style={sectionEyebrowStyle}>Deployment users</p>
          <h2 style={sectionTitleStyle}>User administration</h2>
        </div>
        <button
          data-testid={
            createDialogOpen ? undefined : phase1AdminTestId("create-user")
          }
          disabled={targetOperationPending}
          style={buttonStyle}
          type="button"
          onClick={() => {
            setCreateDialogOpen(true);
          }}
        >
          Create user
        </button>
      </div>

      <div style={adminWorkspaceStyle}>
        <section style={userListPaneStyle}>
          <div style={compactPanelHeaderStyle}>
            <div>
              <p style={sectionEyebrowStyle}>Directory</p>
              <h3 style={subsectionTitleStyle}>Loaded users</h3>
            </div>
            <button
              style={secondaryButtonStyle}
              type="button"
              onClick={() => {
                void refreshUsers();
              }}
            >
              Refresh users
            </button>
          </div>
          <label htmlFor="admin-user-filter" style={labelBlockStyle}>
            Search users
          </label>
          <input
            data-testid={phase1AdminTestId("user-filter")}
            id="admin-user-filter"
            style={inputStyle}
            value={userFilter}
            onChange={(event) => {
              userFiltersTouchedRef.current = true;
              setUserFilter(event.target.value);
            }}
            onKeyDown={(event) => {
              if (event.key === "Enter") {
                void refreshUsers();
              }
            }}
            placeholder="Name, email, id, status"
          />
          <div style={checkboxRowStyle}>
            <label style={labelBlockStyle}>
              Active
              <select
                data-testid={phase1AdminTestId("user-is-active-filter")}
                style={inputStyle}
                value={userActiveFilter}
                onChange={(event) => {
                  userFiltersTouchedRef.current = true;
                  setUserActiveFilter(event.target.value);
                }}
              >
                <option value="all">All</option>
                <option value="true">Active</option>
                <option value="false">Disabled</option>
              </select>
            </label>
            <label style={labelBlockStyle}>
              Deployment admin
              <select
                data-testid={phase1AdminTestId(
                  "user-is-deployment-admin-filter",
                )}
                style={inputStyle}
                value={userAdminFilter}
                onChange={(event) => {
                  userFiltersTouchedRef.current = true;
                  setUserAdminFilter(event.target.value);
                }}
              >
                <option value="all">All</option>
                <option value="true">Admins</option>
                <option value="false">Standard users</option>
              </select>
            </label>
          </div>
          <div
            data-testid={phase1AdminTestId("user-list")}
            style={userListStyle}
          >
            {users.map((user) => {
              const selected = selectedUser?.user_id === user.user_id;
              return (
                <button
                  key={user.user_id}
                  aria-pressed={selected}
                  data-testid={deploymentUserRowTestId(user.user_id)}
                  style={
                    selected ? selectedUserRowButtonStyle : userRowButtonStyle
                  }
                  type="button"
                  onClick={() => {
                    void loadSelectedUser(user.user_id);
                  }}
                >
                  <span style={userRowPrimaryStyle}>{user.display_name}</span>
                  <span style={userRowSecondaryStyle}>{user.email}</span>
                  <span style={userRowMetaStyle}>
                    v{user.user_version} ·{" "}
                    {user.is_active ? "active" : "disabled"} ·{" "}
                    {user.is_deployment_admin
                      ? "deployment admin"
                      : "standard user"}
                  </span>
                </button>
              );
            })}
            {users.length === 0 ? (
              <p style={bodyStyle}>No deployment users loaded.</p>
            ) : null}
          </div>
          <button
            data-testid={phase1AdminTestId("load-more-users")}
            disabled={!usersHasMore || targetOperationPending}
            style={secondaryButtonStyle}
            type="button"
            onClick={() => {
              void loadMoreUsers();
            }}
          >
            Load more users
          </button>
        </section>

        <div style={userInspectorPaneStyle}>
          {selectedUser === null ? (
            <section style={emptyInspectorStyle}>
              <p style={subsectionTitleStyle}>User detail</p>
              <p style={bodyStyle}>
                Select a user from the directory to view profile fields, account
                state, credential actions, and extension-owned bindings.
              </p>
            </section>
          ) : (
            <>
              <section style={inspectorSectionStyle}>
                <div style={compactPanelHeaderStyle}>
                  <div>
                    <p style={sectionEyebrowStyle}>Selected user</p>
                    <h3 style={subsectionTitleStyle}>
                      {selectedUser.display_name}
                    </h3>
                  </div>
                  <button
                    style={secondaryButtonStyle}
                    type="button"
                    onClick={clearSelectedUser}
                  >
                    Clear
                  </button>
                </div>
                <div style={detailGridStyle}>
                  <div>
                    <span style={labelStyle}>Loaded user id</span>
                    <div
                      data-testid={phase1AdminTestId("target-user-id")}
                      style={monoTextStyle}
                    >
                      {selectedUser.user_id}
                    </div>
                  </div>
                  <div>
                    <span style={labelStyle}>User version</span>
                    <div data-testid={phase1AdminTestId("target-user-version")}>
                      {selectedUser.user_version}
                    </div>
                  </div>
                  <div>
                    <span style={labelStyle}>Is active</span>
                    <div data-testid={phase1AdminTestId("target-is-active")}>
                      {String(selectedUser.is_active)}
                    </div>
                  </div>
                  <div>
                    <span style={labelStyle}>Deployment admin</span>
                    <div
                      data-testid={phase1AdminTestId(
                        "target-is-deployment-admin",
                      )}
                    >
                      {String(selectedUser.is_deployment_admin)}
                    </div>
                  </div>
                  <div>
                    <span style={labelStyle}>Base user version</span>
                    <div data-testid={phase1AdminTestId("patch-base-version")}>
                      {targetBaseVersion}
                    </div>
                  </div>
                </div>
                <div style={formGridStyle}>
                  <label htmlFor="admin-patch-email" style={labelBlockStyle}>
                    Email
                    <input
                      data-testid={phase1AdminTestId("patch-email")}
                      id="admin-patch-email"
                      disabled={!canSubmitTargetAction}
                      style={inputStyle}
                      value={patchEmail}
                      onChange={(event) => {
                        setPatchEmail(event.target.value);
                      }}
                    />
                  </label>
                  <label
                    htmlFor="admin-patch-display-name"
                    style={labelBlockStyle}
                  >
                    Display name
                    <input
                      data-testid={phase1AdminTestId("patch-display-name")}
                      id="admin-patch-display-name"
                      disabled={!canSubmitTargetAction}
                      style={inputStyle}
                      value={patchDisplayName}
                      onChange={(event) => {
                        setPatchDisplayName(event.target.value);
                      }}
                    />
                  </label>
                </div>
                <div style={checkboxRowStyle}>
                  <label style={checkboxLabelStyle}>
                    <input
                      data-testid={phase1AdminTestId("patch-mfa-required")}
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
                      data-testid={phase1AdminTestId("patch-is-active")}
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
                      data-testid={phase1AdminTestId(
                        "patch-is-deployment-admin",
                      )}
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
                    data-testid={phase1AdminTestId("patch-user")}
                    disabled={!canSubmitVersionedTargetAction}
                    style={buttonStyle}
                    type="button"
                    onClick={() => {
                      void handlePatchUser();
                    }}
                  >
                    Save user
                  </button>
                </div>
              </section>

              <section style={inspectorSectionStyle}>
                <p style={subsectionTitleStyle}>Credential actions</p>
                <p style={bodyStyle}>
                  Credential actions run only after confirmation and may revoke
                  active sessions for the selected user.
                </p>
                <div style={buttonRowStyle}>
                  <button
                    data-testid={
                      credentialDialog === "password"
                        ? undefined
                        : phase1AdminTestId("password-reset")
                    }
                    disabled={!canSubmitVersionedTargetAction}
                    style={destructiveButtonStyle}
                    type="button"
                    onClick={() => {
                      setCredentialDialog("password");
                    }}
                  >
                    Reset password
                  </button>
                  <button
                    data-testid={
                      credentialDialog === "totp"
                        ? undefined
                        : phase1AdminTestId("totp-reset")
                    }
                    disabled={!canSubmitVersionedTargetAction}
                    style={destructiveButtonStyle}
                    type="button"
                    onClick={() => {
                      setCredentialDialog("totp");
                    }}
                  >
                    Reset TOTP
                  </button>
                  <button
                    data-testid={
                      credentialDialog === "revoke"
                        ? undefined
                        : phase1AdminTestId("revoke-all")
                    }
                    disabled={!canSubmitTargetAction}
                    style={destructiveButtonStyle}
                    type="button"
                    onClick={() => {
                      setCredentialDialog("revoke");
                    }}
                  >
                    Revoke all sessions
                  </button>
                </div>
              </section>

              {enterpriseAuthClaimed ? (
                <section style={inspectorSectionStyle}>
                  <p style={subsectionTitleStyle}>Enterprise bindings</p>
                  <div style={detailGridStyle}>
                    <div>
                      <span style={labelStyle}>Local identity</span>
                      <div style={monoTextStyle}>
                        {selectedLocalBindings.length === 0
                          ? "No local binding loaded"
                          : selectedLocalBindings
                              .map((binding) =>
                                binding.provider_type === "local"
                                  ? `${binding.provider_key}: ${binding.username}`
                                  : "",
                              )
                              .filter((value) => value !== "")
                              .join(", ")}
                      </div>
                    </div>
                    <div>
                      <span style={labelStyle}>Enterprise bindings</span>
                      <div style={monoTextStyle}>
                        {selectedEnterpriseBindings.length}
                      </div>
                    </div>
                    <div>
                      <span style={labelStyle}>
                        Configured providers discovered
                      </span>
                      <div style={monoTextStyle}>
                        {enterpriseProviders.length}
                      </div>
                    </div>
                  </div>

                  <div style={userListStyle}>
                    {selectedEnterpriseBindings.map((binding) => (
                      <label
                        key={binding.auth_binding_id}
                        style={
                          bindingTargetID === binding.auth_binding_id
                            ? selectedUserRowButtonStyle
                            : userRowButtonStyle
                        }
                      >
                        <span style={checkboxLabelStyle}>
                          <input
                            type="radio"
                            name="enterprise-auth-binding-target"
                            checked={
                              bindingTargetID === binding.auth_binding_id
                            }
                            disabled={!canSubmitTargetAction}
                            onChange={() => {
                              setBindingTargetID(binding.auth_binding_id);
                            }}
                          />
                          <span style={userRowPrimaryStyle}>
                            {binding.provider_type.toUpperCase()} ·{" "}
                            {binding.provider_key}
                          </span>
                        </span>
                        <span style={userRowSecondaryStyle}>
                          Subject: {binding.provider_subject}
                        </span>
                        <span style={userRowMetaStyle}>
                          Created {binding.created_at}; last authenticated{" "}
                          {binding.last_auth_at ?? "not recorded"}
                        </span>
                      </label>
                    ))}
                    {selectedEnterpriseBindings.length === 0 ? (
                      <p style={bodyStyle}>
                        No enterprise bindings are attached to this user.
                      </p>
                    ) : null}
                  </div>

                  <datalist id="admin-enterprise-provider-options">
                    {enterpriseProviders.map((provider) => (
                      <option
                        key={provider.provider_key}
                        value={provider.provider_key}
                      >
                        {provider.display_name} ({provider.provider_type})
                      </option>
                    ))}
                  </datalist>

                  <div style={formGridStyle}>
                    <label
                      htmlFor="admin-enterprise-provider-key"
                      style={labelBlockStyle}
                    >
                      Provider key
                      <input
                        id="admin-enterprise-provider-key"
                        list="admin-enterprise-provider-options"
                        disabled={!canSubmitTargetAction}
                        style={inputStyle}
                        value={bindingProviderKey}
                        onChange={(event) => {
                          setBindingProviderKey(event.target.value);
                        }}
                      />
                    </label>
                    <label
                      htmlFor="admin-enterprise-provider-subject"
                      style={labelBlockStyle}
                    >
                      Provider subject
                      <input
                        id="admin-enterprise-provider-subject"
                        disabled={!canSubmitTargetAction}
                        style={inputStyle}
                        value={bindingProviderSubject}
                        onChange={(event) => {
                          setBindingProviderSubject(event.target.value);
                        }}
                      />
                    </label>
                    <label
                      htmlFor="admin-enterprise-new-subject"
                      style={labelBlockStyle}
                    >
                      New provider subject
                      <input
                        id="admin-enterprise-new-subject"
                        disabled={
                          !canSubmitTargetAction || bindingTargetID === ""
                        }
                        style={inputStyle}
                        value={bindingNewSubject}
                        onChange={(event) => {
                          setBindingNewSubject(event.target.value);
                        }}
                      />
                    </label>
                    <label
                      htmlFor="admin-enterprise-binding-reason"
                      style={labelBlockStyle}
                    >
                      Reason
                      <input
                        id="admin-enterprise-binding-reason"
                        style={inputStyle}
                        value={bindingReason}
                        onChange={(event) => {
                          setBindingReason(event.target.value);
                        }}
                      />
                    </label>
                  </div>
                  <div style={buttonRowStyle}>
                    <button
                      disabled={!canSubmitVersionedTargetAction}
                      style={buttonStyle}
                      type="button"
                      onClick={() => {
                        void handleCreateEnterpriseBinding();
                      }}
                    >
                      Create binding
                    </button>
                    <button
                      disabled={
                        !canSubmitVersionedTargetAction ||
                        bindingTargetID === ""
                      }
                      style={buttonStyle}
                      type="button"
                      onClick={() => {
                        void handleRotateEnterpriseBinding();
                      }}
                    >
                      Rotate subject
                    </button>
                    <button
                      disabled={
                        !canSubmitVersionedTargetAction ||
                        bindingTargetID === ""
                      }
                      style={secondaryButtonStyle}
                      type="button"
                      onClick={() => {
                        void handleRetireEnterpriseBinding();
                      }}
                    >
                      Retire binding
                    </button>
                  </div>
                </section>
              ) : null}
            </>
          )}
        </div>
      </div>

      {createDialogOpen ? (
        <div style={dialogBackdropStyle}>
          <section
            aria-label="Create local user"
            aria-modal="true"
            role="dialog"
            style={dialogStyle}
          >
            <header style={dialogHeaderStyle}>
              <div>
                <p style={sectionEyebrowStyle}>Deployment users</p>
                <h3 style={subsectionTitleStyle}>Create local user</h3>
              </div>
              <button
                aria-label="Close create user"
                style={iconButtonStyle}
                type="button"
                onClick={() => setCreateDialogOpen(false)}
              >
                <X aria-hidden="true" size={16} />
              </button>
            </header>
            <div style={formGridStyle}>
              <label htmlFor="admin-create-email" style={labelBlockStyle}>
                Email
                <input
                  data-testid={phase1AdminTestId("create-email")}
                  id="admin-create-email"
                  disabled={targetOperationPending}
                  style={inputStyle}
                  value={createEmail}
                  onChange={(event) => {
                    setCreateEmail(event.target.value);
                  }}
                />
              </label>
              <label
                htmlFor="admin-create-display-name"
                style={labelBlockStyle}
              >
                Display name
                <input
                  data-testid={phase1AdminTestId("create-display-name")}
                  id="admin-create-display-name"
                  disabled={targetOperationPending}
                  style={inputStyle}
                  value={createDisplayName}
                  onChange={(event) => {
                    setCreateDisplayName(event.target.value);
                  }}
                />
              </label>
              <label htmlFor="admin-create-password" style={labelBlockStyle}>
                Initial password
                <input
                  data-testid={phase1AdminTestId("create-password")}
                  id="admin-create-password"
                  disabled={targetOperationPending}
                  style={inputStyle}
                  type="password"
                  value={createInitialPassword}
                  onChange={(event) => {
                    setCreateInitialPassword(event.target.value);
                  }}
                />
              </label>
            </div>
            <div style={checkboxRowStyle}>
              <label style={checkboxLabelStyle}>
                <input
                  data-testid={phase1AdminTestId("create-mfa-required")}
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
                  data-testid={phase1AdminTestId("create-is-deployment-admin")}
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
            <div style={dialogButtonRowStyle}>
              <button
                style={secondaryButtonStyle}
                type="button"
                onClick={() => setCreateDialogOpen(false)}
              >
                Cancel
              </button>
              <button
                data-testid={phase1AdminTestId("create-user")}
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
        </div>
      ) : null}

      {credentialDialog !== null && selectedUser !== null ? (
        <div style={dialogBackdropStyle}>
          <section
            aria-label="Confirm credential action"
            aria-modal="true"
            role="dialog"
            style={dialogStyle}
          >
            <header style={dialogHeaderStyle}>
              <div>
                <p style={sectionEyebrowStyle}>Credential action</p>
                <h3 style={subsectionTitleStyle}>
                  {credentialDialog === "password"
                    ? "Reset password"
                    : credentialDialog === "totp"
                      ? "Reset TOTP"
                      : "Revoke all sessions"}
                </h3>
              </div>
              <button
                aria-label="Close credential action"
                style={iconButtonStyle}
                type="button"
                onClick={() => setCredentialDialog(null)}
              >
                <X aria-hidden="true" size={16} />
              </button>
            </header>
            <p style={bodyStyle}>
              {credentialDialog === "password"
                ? "This resets the selected user's local password, revokes active sessions, and invalidates any pending TOTP bootstrap."
                : credentialDialog === "totp"
                  ? "This clears the selected user's active TOTP credential and revokes active sessions."
                  : "This revokes every active session for the selected user."}
            </p>
            <div style={formGridStyle}>
              {credentialDialog === "password" ? (
                <label htmlFor="admin-new-password" style={labelBlockStyle}>
                  New password
                  <input
                    data-testid={phase1AdminTestId("new-password")}
                    id="admin-new-password"
                    style={inputStyle}
                    type="password"
                    value={adminNewPassword}
                    onChange={(event) => {
                      setAdminNewPassword(event.target.value);
                    }}
                  />
                </label>
              ) : null}
              <label htmlFor="admin-reason" style={labelBlockStyle}>
                Reason
                <input
                  data-testid={phase1AdminTestId("reason")}
                  id="admin-reason"
                  style={inputStyle}
                  value={adminReason}
                  onChange={(event) => {
                    setAdminReason(event.target.value);
                  }}
                />
              </label>
            </div>
            <div style={dialogButtonRowStyle}>
              <button
                style={secondaryButtonStyle}
                type="button"
                onClick={() => setCredentialDialog(null)}
              >
                Cancel
              </button>
              <button
                data-testid={
                  credentialDialog === "password"
                    ? phase1AdminTestId("password-reset")
                    : credentialDialog === "totp"
                      ? phase1AdminTestId("totp-reset")
                      : phase1AdminTestId("revoke-all")
                }
                disabled={
                  credentialDialog === "revoke"
                    ? !canSubmitTargetAction
                    : !canSubmitVersionedTargetAction
                }
                style={destructiveButtonStyle}
                type="button"
                onClick={() => {
                  if (credentialDialog === "password") {
                    void handleAdminPasswordReset();
                    return;
                  }
                  if (credentialDialog === "totp") {
                    void handleAdminTotpReset();
                    return;
                  }
                  void handleAdminRevokeAll();
                }}
              >
                Confirm
              </button>
            </div>
          </section>
        </div>
      ) : null}

      <p
        aria-live="polite"
        data-testid={phase1AdminTestId("status")}
        role="status"
        style={statusTextStyle}
      >
        {statusText}
      </p>
      <p
        aria-live="assertive"
        data-testid={phase1ErrorCodeTestId("admin")}
        role={error === null ? undefined : "alert"}
        style={errorStyle}
      >
        {publicErrorView(error)?.code ?? ""}
      </p>
      <PublicErrorSummary
        error={error}
        testIds={phase1ErrorSummaryTestIds("admin")}
      />
    </section>
  );
});

function PublicErrorSummary({
  error,
  testIds,
}: {
  error: APIError | null;
  testIds: {
    readonly container: string;
    readonly details: string;
    readonly message: string;
  };
}) {
  const view = publicErrorView(error);
  return (
    <div
      data-testid={testIds.container}
      role={view === null ? undefined : "alert"}
      style={publicErrorStyle}
    >
      <p data-testid={testIds.message} style={errorMessageStyle}>
        {view?.statusText ?? ""}
      </p>
      <p data-testid={testIds.details} style={errorDetailStyle}>
        {view?.details
          .map((detail) => `${detail.label}: ${detail.value}`)
          .join(" · ") ?? ""}
      </p>
    </div>
  );
}

const shellStyle: CSSProperties = {
  minHeight: "100vh",
  padding: "2rem",
  background: "var(--ct-colors-canvas)",
  color: "var(--ct-colors-ink)",
  fontFamily: "var(--ct-typography-ui-fontFamily)",
};

const shellPanelStyle: CSSProperties = {
  width: "min(56rem, 100%)",
  margin: "0 auto",
  padding: "2rem",
  borderRadius: "var(--ct-rounded-lg)",
  background: "var(--ct-colors-surface-1)",
  boxShadow: "var(--ct-elevation-panel)",
  border: "var(--ct-border-hairline)",
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
  color: "var(--ct-colors-accent)",
};

const headlineStyle: CSSProperties = {
  margin: "0.35rem 0 0",
  fontSize: "clamp(2rem, 3vw, 2.8rem)",
  lineHeight: 1.05,
};

const bodyStyle: CSSProperties = {
  margin: "0.75rem 0 0",
  color: "var(--ct-colors-ink-muted)",
  maxWidth: "42rem",
  overflowWrap: "anywhere",
};

const statusCardStyle: CSSProperties = {
  minWidth: "14rem",
  padding: "0.9rem 1rem",
  borderRadius: "var(--ct-rounded-lg)",
  background: "var(--ct-colors-surface-2)",
  border: "var(--ct-border-hairline)",
};

const cardStyle: CSSProperties = {
  padding: "1.4rem",
  borderRadius: "var(--ct-rounded-lg)",
  background: "var(--ct-colors-surface-2)",
  border: "var(--ct-border-hairline)",
  color: "var(--ct-colors-ink)",
  minWidth: 0,
  boxSizing: "border-box",
};

const cardHeaderStyle: CSSProperties = {
  display: "flex",
  justifyContent: "space-between",
  gap: "1rem",
  alignItems: "center",
  marginBottom: "1rem",
};

const compactPanelHeaderStyle: CSSProperties = {
  display: "flex",
  justifyContent: "space-between",
  gap: "0.75rem",
  alignItems: "center",
  marginBottom: "1rem",
};

const sectionEyebrowStyle: CSSProperties = {
  margin: 0,
  fontSize: "0.72rem",
  letterSpacing: "0.14em",
  textTransform: "uppercase",
  color: "var(--ct-colors-ink-subtle)",
};

const sectionTitleStyle: CSSProperties = {
  margin: "0.35rem 0 0",
  fontSize: "1.15rem",
};

const subsectionStyle: CSSProperties = {
  marginTop: "1.5rem",
  paddingTop: "1.25rem",
  borderTop: "var(--ct-border-hairline)",
  minWidth: 0,
};

const inspectorSectionStyle: CSSProperties = {
  marginTop: "0",
  marginBottom: "1rem",
  padding: "var(--ct-spacing-md)",
  borderRadius: "var(--ct-rounded-md)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
  minWidth: 0,
};

const emptyInspectorStyle: CSSProperties = {
  ...inspectorSectionStyle,
  minHeight: "12rem",
  display: "grid",
  alignContent: "center",
};

const adminWorkspaceStyle: CSSProperties = {
  display: "grid",
  gridTemplateColumns: "repeat(auto-fit, minmax(min(100%, 24rem), 1fr))",
  gap: "var(--ct-spacing-lg)",
  alignItems: "start",
};

const userListPaneStyle: CSSProperties = {
  minWidth: 0,
  padding: "var(--ct-spacing-md)",
  borderRadius: "var(--ct-rounded-md)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
};

const userInspectorPaneStyle: CSSProperties = {
  minWidth: 0,
};

const subsectionTitleStyle: CSSProperties = {
  margin: 0,
  fontWeight: 700,
  color: "var(--ct-colors-ink)",
};

const formGridStyle: CSSProperties = {
  display: "grid",
  gridTemplateColumns: "minmax(0, 1fr)",
  gap: "0.75rem",
  marginTop: "1rem",
  minWidth: 0,
};

const detailGridStyle: CSSProperties = {
  display: "grid",
  gridTemplateColumns: "repeat(auto-fit, minmax(12rem, 1fr))",
  gap: "0.85rem",
  marginTop: "1rem",
  minWidth: 0,
};

const userListStyle: CSSProperties = {
  display: "grid",
  gap: "var(--ct-spacing-sm)",
  margin: "var(--ct-spacing-md) 0",
};

const userRowButtonStyle: CSSProperties = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
  width: "100%",
  minWidth: 0,
  padding: "var(--ct-spacing-sm)",
  borderRadius: "var(--ct-rounded-sm)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-ink)",
  textAlign: "left",
  cursor: "pointer",
};

const selectedUserRowButtonStyle: CSSProperties = {
  ...userRowButtonStyle,
  border: "var(--ct-border-strong)",
  background: "var(--ct-colors-surface-3)",
  boxShadow: "inset 3px 0 0 var(--ct-colors-accent)",
};

const userRowPrimaryStyle: CSSProperties = {
  fontWeight: 700,
  overflowWrap: "anywhere",
};

const userRowSecondaryStyle: CSSProperties = {
  color: "var(--ct-colors-ink-muted)",
  overflowWrap: "anywhere",
};

const userRowMetaStyle: CSSProperties = {
  color: "var(--ct-colors-ink-subtle)",
  fontSize: "0.82rem",
};

const labelStyle: CSSProperties = {
  display: "block",
  fontSize: "0.72rem",
  letterSpacing: "0.12em",
  textTransform: "uppercase",
  color: "var(--ct-colors-ink-subtle)",
  marginBottom: "0.35rem",
};

const labelBlockStyle: CSSProperties = {
  fontSize: "0.84rem",
  fontWeight: 600,
  color: "var(--ct-colors-ink-muted)",
  minWidth: 0,
};

const inputStyle: CSSProperties = {
  boxSizing: "border-box",
  width: "100%",
  maxWidth: "100%",
  minWidth: 0,
  marginTop: "0.35rem",
  padding: "var(--ct-component-text-input-padding)",
  borderRadius: "var(--ct-component-text-input-rounded)",
  border: "var(--ct-component-text-input-border)",
  background: "var(--ct-component-text-input-backgroundColor)",
  color: "var(--ct-component-text-input-textColor)",
  fontSize: "0.95rem",
};

const buttonRowStyle: CSSProperties = {
  display: "flex",
  flexWrap: "wrap",
  gap: "0.75rem",
  marginTop: "1rem",
};

const buttonStyle: CSSProperties = {
  padding: "var(--ct-component-button-primary-padding)",
  borderRadius: "var(--ct-component-button-primary-rounded)",
  border: "none",
  background: "var(--ct-component-button-primary-backgroundColor)",
  color: "var(--ct-component-button-primary-textColor)",
  fontWeight: 600,
  cursor: "pointer",
};

const secondaryButtonStyle: CSSProperties = {
  ...buttonStyle,
  background: "var(--ct-component-button-secondary-backgroundColor)",
  color: "var(--ct-component-button-secondary-textColor)",
  border: "var(--ct-component-button-secondary-border)",
};

const destructiveButtonStyle: CSSProperties = {
  ...buttonStyle,
  background: "var(--ct-colors-semantic-conflict)",
  color: "var(--ct-colors-surface-1)",
  border: "none",
};

const dialogBackdropStyle: CSSProperties = {
  position: "fixed",
  inset: 0,
  zIndex: 40,
  display: "grid",
  placeItems: "center",
  padding: "1.5rem",
  background: "rgba(10, 13, 18, 0.68)",
};

const dialogStyle: CSSProperties = {
  width: "min(44rem, 100%)",
  maxHeight: "calc(100vh - 3rem)",
  overflow: "auto",
  display: "grid",
  gap: "1rem",
  padding: "1.25rem",
  borderRadius: "var(--ct-rounded-md)",
  border: "var(--ct-border-strong)",
  background: "var(--ct-colors-surface-1)",
  boxShadow: "var(--ct-elevation-panel)",
};

const dialogHeaderStyle: CSSProperties = {
  display: "flex",
  justifyContent: "space-between",
  gap: "1rem",
  alignItems: "flex-start",
};

const iconButtonStyle: CSSProperties = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  width: "2rem",
  height: "2rem",
  borderRadius: "var(--ct-rounded-sm)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-ink)",
  cursor: "pointer",
};

const dialogButtonRowStyle: CSSProperties = {
  display: "flex",
  justifyContent: "flex-end",
  flexWrap: "wrap",
  gap: "0.75rem",
};

const monoTextStyle: CSSProperties = {
  fontFamily: "var(--ct-typography-mono-fontFamily)",
  overflowWrap: "anywhere",
  minWidth: 0,
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
  color: "var(--ct-colors-ink-muted)",
};

const statusTextStyle: CSSProperties = {
  margin: "1rem 0 0",
  minHeight: "1.5rem",
  color: "var(--ct-colors-ink-muted)",
};

const errorStyle: CSSProperties = {
  margin: "0.25rem 0 0",
  minHeight: "1.25rem",
  color: "var(--ct-colors-semantic-conflict)",
  fontWeight: 600,
};

const publicErrorStyle: CSSProperties = {
  marginTop: "0.25rem",
};

const errorMessageStyle: CSSProperties = {
  margin: 0,
  minHeight: "1.25rem",
  color: "var(--ct-colors-semantic-conflict)",
};

const errorDetailStyle: CSSProperties = {
  margin: "0.2rem 0 0",
  minHeight: "1.25rem",
  color: "var(--ct-colors-semantic-conflict)",
  overflowWrap: "anywhere",
};
