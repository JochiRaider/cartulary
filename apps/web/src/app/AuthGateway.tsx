import {
  phase1AuthTestId,
  phase1ErrorCodeTestId,
  phase1ErrorSummaryTestIds,
} from "@cartulary/ui-contracts";
import { Eye, EyeOff } from "lucide-react";
import { type FormEvent, useEffect, useReducer, useRef } from "react";

import { type APIError, extractError } from "../services/browserApi";
import {
  beginEnterpriseAuth,
  beginTotpEnrollment,
  completeTotpEnrollment,
  type EnterpriseAuthProvider,
  listEnterpriseAuthProviders,
  loadSession,
  loginLocal,
} from "./phase1Client";

type AuthSurfaceBootstrapState =
  | "loading"
  | "anonymous"
  | "revoked"
  | "public_error_envelope";

type AuthChallengeState = "mfa_required" | "mfa_setup_required";

type Phase1AuthSurfaceProps = {
  bootstrapState: AuthSurfaceBootstrapState;
  message: string;
  onAuthenticated: () => Promise<void> | void;
  publicError?: APIError | null;
  readingProfile?: "default" | "hyperlegible" | undefined;
};

type AuthPhase = "credentials" | "mfa" | "setup";
type SetupAction = "beginning" | "completing" | "idle";
type BannerTone = "error" | "info" | "success";

type AuthBanner = {
  message: string;
  tone: BannerTone;
};

type FieldErrors = {
  bootstrapCompleteCode?: string;
  password?: string;
  totpCode?: string;
  username?: string;
};

type AuthState = {
  banner: AuthBanner | null;
  bootstrapCompleteCode: string;
  bootstrapEnrollmentId: string;
  bootstrapSecretBase32: string;
  bootstrapToken: string;
  enterprisePendingProviderKey: string | null;
  enterpriseProviders: EnterpriseAuthProvider[];
  error: APIError | null;
  fieldErrors: FieldErrors;
  password: string;
  passwordVisible: boolean;
  phase: AuthPhase;
  setupAction: SetupAction;
  submitting: boolean;
  totpCode: string;
  username: string;
};

type AuthAction =
  | { type: "auth_failure"; banner: AuthBanner; error: APIError | null }
  | { type: "enterprise_begin_started"; providerKey: string }
  | {
      type: "field";
      field: "bootstrapCompleteCode" | "password" | "totpCode" | "username";
      value: string;
    }
  | { type: "login_started" }
  | { type: "login_succeeded" }
  | { type: "mfa_required" }
  | {
      type: "mfa_setup_required";
      bootstrapToken: string;
      error: APIError | null;
    }
  | { type: "providers_loaded"; providers: EnterpriseAuthProvider[] }
  | { type: "setup_begin_failed"; banner: AuthBanner; error: APIError | null }
  | { type: "setup_begin_started" }
  | {
      type: "setup_begin_succeeded";
      enrollmentId: string;
      secretBase32: string;
    }
  | {
      type: "setup_complete_failed";
      banner: AuthBanner | null;
      error: APIError | null;
      fieldError?: string;
    }
  | { type: "setup_complete_started" }
  | { type: "setup_complete_succeeded" }
  | { type: "toggle_password_visibility" }
  | { type: "use_different_account" }
  | { type: "validation_failed"; fieldErrors: FieldErrors };

const initialAuthState: AuthState = {
  banner: null,
  bootstrapCompleteCode: "",
  bootstrapEnrollmentId: "",
  bootstrapSecretBase32: "",
  bootstrapToken: "",
  enterprisePendingProviderKey: null,
  enterpriseProviders: [],
  error: null,
  fieldErrors: {},
  password: "",
  passwordVisible: false,
  phase: "credentials",
  setupAction: "idle",
  submitting: false,
  totpCode: "",
  username: "",
};

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

function authReducer(state: AuthState, action: AuthAction): AuthState {
  switch (action.type) {
    case "auth_failure":
      return {
        ...state,
        banner: action.banner,
        error: action.error,
        enterprisePendingProviderKey: null,
        fieldErrors: {},
        submitting: false,
      };
    case "enterprise_begin_started":
      return {
        ...state,
        banner: null,
        enterprisePendingProviderKey: action.providerKey,
        error: null,
        fieldErrors: {},
      };
    case "field":
      return {
        ...state,
        [action.field]: action.value,
        banner: null,
        error: null,
        fieldErrors: {
          ...state.fieldErrors,
          [action.field]: undefined,
        },
      };
    case "login_started":
      return {
        ...state,
        banner: null,
        error: null,
        fieldErrors: {},
        submitting: true,
      };
    case "login_succeeded":
      return {
        ...state,
        banner: null,
        bootstrapCompleteCode: "",
        bootstrapEnrollmentId: "",
        bootstrapSecretBase32: "",
        bootstrapToken: "",
        error: null,
        fieldErrors: {},
        submitting: false,
        totpCode: "",
      };
    case "mfa_required":
      return {
        ...state,
        banner: null,
        bootstrapCompleteCode: "",
        bootstrapEnrollmentId: "",
        bootstrapSecretBase32: "",
        bootstrapToken: "",
        error: null,
        fieldErrors: {},
        phase: "mfa",
        submitting: false,
        totpCode: "",
      };
    case "mfa_setup_required":
      return {
        ...state,
        banner: {
          message: "Authenticator setup is required before sign-in.",
          tone: "info",
        },
        bootstrapCompleteCode: "",
        bootstrapEnrollmentId: "",
        bootstrapSecretBase32: "",
        bootstrapToken: action.bootstrapToken,
        error: action.error,
        fieldErrors: {},
        phase: "setup",
        setupAction: "idle",
        submitting: false,
        totpCode: "",
      };
    case "providers_loaded":
      return {
        ...state,
        enterpriseProviders: action.providers,
      };
    case "setup_begin_failed":
      return {
        ...state,
        banner: action.banner,
        error: action.error,
        setupAction: "idle",
      };
    case "setup_begin_started":
      return {
        ...state,
        banner: null,
        error: null,
        fieldErrors: {},
        setupAction: "beginning",
      };
    case "setup_begin_succeeded":
      return {
        ...state,
        banner: {
          message: "Authenticator enrollment started.",
          tone: "info",
        },
        bootstrapEnrollmentId: action.enrollmentId,
        bootstrapSecretBase32: action.secretBase32,
        error: null,
        setupAction: "idle",
      };
    case "setup_complete_failed":
      return {
        ...state,
        banner: action.banner,
        error: action.error,
        fieldErrors:
          typeof action.fieldError === "string"
            ? {
                ...state.fieldErrors,
                bootstrapCompleteCode: action.fieldError,
              }
            : state.fieldErrors,
        setupAction: "idle",
      };
    case "setup_complete_started":
      return {
        ...state,
        banner: null,
        error: null,
        fieldErrors: {},
        setupAction: "completing",
      };
    case "setup_complete_succeeded":
      return {
        ...state,
        banner: {
          message: "Authenticator setup is complete. Sign in again.",
          tone: "success",
        },
        bootstrapCompleteCode: "",
        bootstrapEnrollmentId: "",
        bootstrapSecretBase32: "",
        bootstrapToken: "",
        error: null,
        fieldErrors: {},
        password: "",
        phase: "credentials",
        setupAction: "idle",
        submitting: false,
        totpCode: "",
      };
    case "toggle_password_visibility":
      return {
        ...state,
        passwordVisible: !state.passwordVisible,
      };
    case "use_different_account":
      return {
        ...state,
        banner: null,
        bootstrapCompleteCode: "",
        bootstrapEnrollmentId: "",
        bootstrapSecretBase32: "",
        bootstrapToken: "",
        error: null,
        fieldErrors: {},
        password: "",
        phase: "credentials",
        setupAction: "idle",
        totpCode: "",
      };
    case "validation_failed":
      return {
        ...state,
        banner: null,
        fieldErrors: action.fieldErrors,
        submitting: false,
      };
  }
}

function normalizeTotpCode(value: string): string {
  return value.replace(/\D/gu, "").slice(0, 6);
}

function validateEmail(value: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/u.test(value.trim());
}

function validateLoginState(state: AuthState): FieldErrors {
  const fieldErrors: FieldErrors = {};
  if (!validateEmail(state.username)) {
    fieldErrors.username = "Enter a valid email address.";
  }
  if (state.password.length === 0) {
    fieldErrors.password = "Enter your password.";
  }
  if (state.phase === "mfa" && !/^\d{6}$/u.test(state.totpCode)) {
    fieldErrors.totpCode = "Enter a six-digit authenticator code.";
  }
  return fieldErrors;
}

function hasFieldErrors(fieldErrors: FieldErrors): boolean {
  return Object.values(fieldErrors).some(
    (value) => typeof value === "string" && value.length > 0,
  );
}

function authBannerForError(error: APIError | null): AuthBanner {
  if (error?.code === "invalid_credentials") {
    return {
      message: "Email or password is incorrect.",
      tone: "error",
    };
  }
  if (error?.code === "invalid_auth_request") {
    return {
      message: "Sign-in request could not be completed.",
      tone: "error",
    };
  }
  if (
    error?.code === "session_required" ||
    error?.code === "auth_required" ||
    error?.code === "credential_bootstrap_rejected"
  ) {
    return {
      message: "Sign in again to continue.",
      tone: "error",
    };
  }
  return {
    message: "Authentication is temporarily unavailable. Try again.",
    tone: "error",
  };
}

function setupBannerForError(error: APIError | null): AuthBanner {
  if (error?.code === "invalid_second_factor") {
    return {
      message: "The verification code is incorrect or expired.",
      tone: "error",
    };
  }
  if (error?.code === "totp_setup_not_pending") {
    return {
      message: "Authenticator setup expired. Start setup again.",
      tone: "error",
    };
  }
  return {
    message: "Authenticator setup could not be completed. Try again.",
    tone: "error",
  };
}

function fieldErrorForAuthRequest(
  error: APIError | null,
  phase: AuthPhase,
): FieldErrors | null {
  if (error?.code !== "invalid_auth_request") {
    return null;
  }
  const field = error.details?.field;
  if (field === "username") {
    return { username: "Enter a valid email address." };
  }
  if (
    phase === "mfa" &&
    (field === "second_factor.assertion.code" ||
      field === "second_factor" ||
      field === "second_factor.assertion")
  ) {
    return { totpCode: "Enter a six-digit authenticator code." };
  }
  return null;
}

function firstFieldError(fieldErrors: FieldErrors): string {
  return (
    fieldErrors.username ??
    fieldErrors.password ??
    fieldErrors.totpCode ??
    fieldErrors.bootstrapCompleteCode ??
    ""
  );
}

export function AuthGateway({
  bootstrapState,
  message,
  onAuthenticated,
  publicError = null,
  readingProfile = "default",
}: Phase1AuthSurfaceProps) {
  const [state, dispatch] = useReducer(authReducer, initialAuthState);
  const usernameRef = useRef<HTMLInputElement | null>(null);
  const passwordRef = useRef<HTMLInputElement | null>(null);
  const totpCodeRef = useRef<HTMLInputElement | null>(null);
  const bootstrapBeginRef = useRef<HTMLButtonElement | null>(null);
  const bootstrapCompleteCodeRef = useRef<HTMLInputElement | null>(null);

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
      if (
        !result.ok &&
        (result.status === 404 ||
          nextError?.code === "extension_profile_not_claimed")
      ) {
        dispatch({ type: "providers_loaded", providers: [] });
        return;
      }
      if (!result.ok) {
        dispatch({ type: "providers_loaded", providers: [] });
        return;
      }
      const data = (
        result.payload as { data: { providers: EnterpriseAuthProvider[] } }
      ).data;
      dispatch({ type: "providers_loaded", providers: data.providers });
    })();
    return () => {
      controller.abort();
    };
  }, [bootstrapState]);

  useEffect(() => {
    if (state.phase === "mfa") {
      totpCodeRef.current?.focus();
      return;
    }
    if (state.phase === "setup") {
      if (state.bootstrapEnrollmentId === "") {
        bootstrapBeginRef.current?.focus();
        return;
      }
      bootstrapCompleteCodeRef.current?.focus();
    }
  }, [state.bootstrapEnrollmentId, state.phase]);

  useEffect(() => {
    if (state.fieldErrors.username) {
      usernameRef.current?.focus();
      return;
    }
    if (state.fieldErrors.password) {
      passwordRef.current?.focus();
      return;
    }
    if (state.fieldErrors.totpCode) {
      totpCodeRef.current?.focus();
      return;
    }
    if (state.fieldErrors.bootstrapCompleteCode) {
      bootstrapCompleteCodeRef.current?.focus();
    }
  }, [state.fieldErrors]);

  async function handleLoginSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (state.submitting) {
      return;
    }
    const validationErrors = validateLoginState(state);
    if (hasFieldErrors(validationErrors)) {
      dispatch({ type: "validation_failed", fieldErrors: validationErrors });
      return;
    }

    dispatch({ type: "login_started" });
    try {
      const result = await loginLocal({
        username: state.username,
        password: state.password,
        ...(state.phase === "mfa" ? { secondFactorCode: state.totpCode } : {}),
      });
      const nextError = extractError(result.payload);
      if (!result.ok) {
        if (nextError?.code === "mfa_required") {
          dispatch({ type: "mfa_required" });
          return;
        }
        if (nextError?.code === "mfa_setup_required") {
          const token = nextError.details?.bootstrap_token;
          dispatch({
            type: "mfa_setup_required",
            bootstrapToken: typeof token === "string" ? token : "",
            error: nextError,
          });
          return;
        }
        if (nextError?.code === "invalid_second_factor") {
          dispatch({
            type: "validation_failed",
            fieldErrors: {
              totpCode: "The verification code is incorrect or expired.",
            },
          });
          return;
        }
        const requestFieldError = fieldErrorForAuthRequest(
          nextError,
          state.phase,
        );
        if (requestFieldError !== null) {
          dispatch({
            type: "validation_failed",
            fieldErrors: requestFieldError,
          });
          return;
        }
        dispatch({
          type: "auth_failure",
          banner: authBannerForError(nextError),
          error: nextError,
        });
        return;
      }

      dispatch({ type: "login_succeeded" });
      await onAuthenticated();
    } catch {
      const sessionResult = await loadSession();
      if (sessionResult.ok) {
        dispatch({ type: "login_succeeded" });
        await onAuthenticated();
        return;
      }
      dispatch({
        type: "auth_failure",
        banner: {
          message: "Sign-in response could not be confirmed. Try again.",
          tone: "error",
        },
        error: null,
      });
    }
  }

  async function handleEnterpriseBegin(providerKey: string) {
    if (state.enterprisePendingProviderKey !== null) {
      return;
    }
    dispatch({ type: "enterprise_begin_started", providerKey });
    const returnTo =
      `${window.location.pathname}${window.location.search}`.trim() || "/";
    const result = await beginEnterpriseAuth({
      providerKey,
      returnTo,
    });
    const nextError = extractError(result.payload);
    if (!result.ok) {
      dispatch({
        type: "auth_failure",
        banner: {
          message: "Enterprise sign-in could not be started.",
          tone: "error",
        },
        error: nextError,
      });
      return;
    }
    const data = (result.payload as { data: { redirect_url: string } }).data;
    enterpriseAuthNavigate(data.redirect_url);
  }

  async function handleBeginBootstrapEnrollment() {
    if (state.setupAction !== "idle") {
      return;
    }
    dispatch({ type: "setup_begin_started" });
    try {
      const result = await beginTotpEnrollment({
        authMode: "bootstrap",
        bootstrapToken: state.bootstrapToken,
      });
      const nextError = extractError(result.payload);
      if (!result.ok) {
        dispatch({
          type: "setup_begin_failed",
          banner: setupBannerForError(nextError),
          error: nextError,
        });
        return;
      }
      const data = (
        result.payload as {
          data: {
            enrollment_id: string;
            totp_setup: { secret_base32: string };
          };
        }
      ).data;
      dispatch({
        type: "setup_begin_succeeded",
        enrollmentId: data.enrollment_id,
        secretBase32: data.totp_setup.secret_base32,
      });
    } catch {
      dispatch({
        type: "setup_begin_failed",
        banner: {
          message: "Authenticator setup could not be started. Try again.",
          tone: "error",
        },
        error: null,
      });
    }
  }

  async function handleCompleteBootstrapEnrollment() {
    if (state.setupAction !== "idle") {
      return;
    }
    if (!/^\d{6}$/u.test(state.bootstrapCompleteCode)) {
      dispatch({
        type: "setup_complete_failed",
        banner: null,
        error: null,
        fieldError: "Enter a six-digit authenticator code.",
      });
      return;
    }
    dispatch({ type: "setup_complete_started" });
    try {
      const result = await completeTotpEnrollment({
        authMode: "bootstrap",
        bootstrapToken: state.bootstrapToken,
        code: state.bootstrapCompleteCode,
        enrollmentId: state.bootstrapEnrollmentId,
      });
      const nextError = extractError(result.payload);
      if (!result.ok) {
        const fieldError =
          nextError?.code === "invalid_second_factor"
            ? "The verification code is incorrect or expired."
            : undefined;
        dispatch({
          type: "setup_complete_failed",
          banner:
            fieldError === undefined ? setupBannerForError(nextError) : null,
          error: nextError,
          ...(fieldError === undefined ? {} : { fieldError }),
        });
        return;
      }
      dispatch({ type: "setup_complete_succeeded" });
    } catch {
      dispatch({
        type: "setup_complete_failed",
        banner: {
          message: "Authenticator setup could not be completed. Try again.",
          tone: "error",
        },
        error: null,
      });
    }
  }

  const displayedBootstrapState:
    | AuthChallengeState
    | AuthSurfaceBootstrapState =
    state.phase === "mfa"
      ? "mfa_required"
      : state.phase === "setup"
        ? "mfa_setup_required"
        : bootstrapState;
  const externalBanner =
    publicError === null || state.banner !== null
      ? null
      : authBannerForError(publicError);
  const displayedBanner = state.banner ?? externalBanner;
  const currentFieldError = firstFieldError(state.fieldErrors);
  const authAlertText = displayedBanner?.message ?? currentFieldError;
  const authLiveRole =
    authAlertText === ""
      ? undefined
      : displayedBanner?.tone === "error" || currentFieldError !== ""
        ? "alert"
        : "status";
  const authLivePoliteness = authLiveRole === "alert" ? "assertive" : "polite";
  const activeErrorCode = state.error?.code ?? publicError?.code ?? "";
  const statusText =
    bootstrapState === "loading"
      ? "Checking current session..."
      : state.submitting
        ? "Signing in..."
        : state.setupAction === "beginning"
          ? "Beginning authenticator enrollment..."
          : state.setupAction === "completing"
            ? "Completing authenticator enrollment..."
            : state.phase === "mfa"
              ? "Authenticator code required."
              : state.phase === "setup"
                ? "Authenticator setup required."
                : "";
  const rootClassName =
    readingProfile === "hyperlegible"
      ? "cartulary-shell cartulary-auth-shell cartulary-auth-hyperlegible"
      : "cartulary-shell cartulary-auth-shell";
  const supportText =
    state.phase === "mfa"
      ? "Enter the authenticator code for this account."
      : state.phase === "setup"
        ? "Complete authenticator setup before signing in."
        : message;
  const title =
    state.phase === "mfa"
      ? "Verify your identity"
      : state.phase === "setup"
        ? "Set up authenticator"
        : "Sign in to Cartulary";
  const submitLabel = state.submitting
    ? "Signing in..."
    : state.phase === "mfa"
      ? "Verify and sign in"
      : "Sign in";
  const canSubmit = !state.submitting;

  return (
    <main
      aria-busy={bootstrapState === "loading"}
      className={rootClassName}
      data-bootstrap-state={displayedBootstrapState}
      data-reading-profile={
        readingProfile === "hyperlegible" ? "hyperlegible" : undefined
      }
      data-testid={phase1AuthTestId("shell")}
    >
      <style>{authGatewayStyleText}</style>
      <section className="cartulary-auth-identity" aria-label="Cartulary">
        <div className="cartulary-auth-wordmark">CARTULARY</div>
        <div className="cartulary-auth-identity-copy">
          <p>Workbook-native incident response.</p>
        </div>
        <div className="cartulary-auth-motif" aria-hidden="true">
          <span />
        </div>
      </section>

      <section className="cartulary-auth-rail" aria-labelledby="auth-title">
        <div className="cartulary-auth-rail-inner">
          <p className="cartulary-auth-eyebrow">LOCAL ACCOUNT</p>
          <h1 id="auth-title" className="cartulary-auth-title">
            {title}
          </h1>
          <p
            className="cartulary-auth-support"
            data-testid={phase1AuthTestId("shell-message")}
          >
            {supportText}
          </p>

          {displayedBanner !== null ? (
            <div
              className="cartulary-auth-banner"
              data-tone={displayedBanner.tone}
              role={displayedBanner.tone === "error" ? "alert" : "status"}
            >
              {displayedBanner.message}
            </div>
          ) : null}

          {state.phase === "setup" ? (
            <section
              className="cartulary-auth-setup"
              aria-label="Authenticator setup"
            >
              <p className="cartulary-auth-setup-copy">
                This account requires authenticator enrollment before it can
                sign in.
              </p>
              <div className="cartulary-auth-detail-list">
                <div>
                  <span className="cartulary-auth-detail-label">
                    Setup token
                  </span>
                  <div data-testid={phase1AuthTestId("bootstrap-token")}>
                    Stored for TOTP setup requests.
                  </div>
                </div>
                <div>
                  <span className="cartulary-auth-detail-label">
                    Enrollment id
                  </span>
                  <div
                    className="cartulary-auth-mono"
                    data-testid={phase1AuthTestId("bootstrap-enrollment-id")}
                  >
                    {state.bootstrapEnrollmentId}
                  </div>
                </div>
                <div>
                  <span className="cartulary-auth-detail-label">
                    Secret base32
                  </span>
                  <div
                    className="cartulary-auth-mono"
                    data-testid={phase1AuthTestId("bootstrap-secret-base32")}
                  >
                    {state.bootstrapSecretBase32}
                  </div>
                </div>
              </div>
              <button
                ref={bootstrapBeginRef}
                aria-disabled={state.setupAction !== "idle"}
                className="cartulary-auth-primary-button"
                data-testid={phase1AuthTestId("bootstrap-begin")}
                disabled={state.setupAction !== "idle"}
                type="button"
                onClick={() => {
                  void handleBeginBootstrapEnrollment();
                }}
              >
                {state.setupAction === "beginning"
                  ? "Beginning setup..."
                  : "Begin enrollment"}
              </button>
              <label
                className="cartulary-auth-field"
                htmlFor="auth-bootstrap-complete-code"
              >
                Authenticator code
                <input
                  ref={bootstrapCompleteCodeRef}
                  aria-describedby={
                    state.fieldErrors.bootstrapCompleteCode
                      ? "auth-bootstrap-complete-code-error"
                      : undefined
                  }
                  aria-invalid={
                    state.fieldErrors.bootstrapCompleteCode ? true : undefined
                  }
                  autoComplete="one-time-code"
                  data-testid={phase1AuthTestId("bootstrap-complete-code")}
                  id="auth-bootstrap-complete-code"
                  inputMode="numeric"
                  maxLength={6}
                  pattern="[0-9]*"
                  type="text"
                  value={state.bootstrapCompleteCode}
                  onChange={(event) => {
                    dispatch({
                      type: "field",
                      field: "bootstrapCompleteCode",
                      value: normalizeTotpCode(event.target.value),
                    });
                  }}
                />
              </label>
              {state.fieldErrors.bootstrapCompleteCode ? (
                <p
                  className="cartulary-auth-field-error"
                  id="auth-bootstrap-complete-code-error"
                  role="alert"
                >
                  {state.fieldErrors.bootstrapCompleteCode}
                </p>
              ) : null}
              <button
                aria-disabled={
                  state.setupAction !== "idle" ||
                  state.bootstrapEnrollmentId === ""
                }
                className="cartulary-auth-primary-button"
                data-testid={phase1AuthTestId("bootstrap-complete")}
                disabled={
                  state.setupAction !== "idle" ||
                  state.bootstrapEnrollmentId === ""
                }
                type="button"
                onClick={() => {
                  void handleCompleteBootstrapEnrollment();
                }}
              >
                {state.setupAction === "completing"
                  ? "Completing setup..."
                  : "Complete enrollment"}
              </button>
              <button
                className="cartulary-auth-secondary-button"
                type="button"
                onClick={() => dispatch({ type: "use_different_account" })}
              >
                Use a different account
              </button>
            </section>
          ) : (
            <form
              className="cartulary-auth-form"
              noValidate
              onSubmit={(event) => {
                void handleLoginSubmit(event);
              }}
            >
              <label
                className="cartulary-auth-field"
                htmlFor="auth-login-username"
              >
                Email
                <input
                  ref={usernameRef}
                  aria-describedby={
                    state.fieldErrors.username
                      ? "auth-login-username-error"
                      : undefined
                  }
                  aria-invalid={state.fieldErrors.username ? true : undefined}
                  autoComplete="username"
                  data-testid={phase1AuthTestId("login-username")}
                  id="auth-login-username"
                  type="email"
                  value={state.username}
                  onChange={(event) => {
                    dispatch({
                      type: "field",
                      field: "username",
                      value: event.target.value,
                    });
                  }}
                />
              </label>
              {state.fieldErrors.username ? (
                <p
                  className="cartulary-auth-field-error"
                  id="auth-login-username-error"
                  role="alert"
                >
                  {state.fieldErrors.username}
                </p>
              ) : null}

              <label
                className="cartulary-auth-field"
                htmlFor="auth-login-password"
              >
                Password
                <span className="cartulary-auth-password-control">
                  <input
                    ref={passwordRef}
                    aria-describedby={
                      state.fieldErrors.password
                        ? "auth-login-password-error"
                        : undefined
                    }
                    aria-invalid={state.fieldErrors.password ? true : undefined}
                    autoComplete="current-password"
                    data-testid={phase1AuthTestId("login-password")}
                    id="auth-login-password"
                    type={state.passwordVisible ? "text" : "password"}
                    value={state.password}
                    onChange={(event) => {
                      dispatch({
                        type: "field",
                        field: "password",
                        value: event.target.value,
                      });
                    }}
                  />
                  <button
                    aria-label={
                      state.passwordVisible ? "Hide password" : "Show password"
                    }
                    className="cartulary-auth-icon-button"
                    type="button"
                    onClick={() =>
                      dispatch({ type: "toggle_password_visibility" })
                    }
                  >
                    {state.passwordVisible ? (
                      <EyeOff aria-hidden="true" size={18} />
                    ) : (
                      <Eye aria-hidden="true" size={18} />
                    )}
                  </button>
                </span>
              </label>
              {state.fieldErrors.password ? (
                <p
                  className="cartulary-auth-field-error"
                  id="auth-login-password-error"
                  role="alert"
                >
                  {state.fieldErrors.password}
                </p>
              ) : null}

              {state.phase === "mfa" ? (
                <>
                  <label
                    className="cartulary-auth-field"
                    htmlFor="auth-login-totp-code"
                  >
                    Authenticator code
                    <input
                      ref={totpCodeRef}
                      aria-describedby={
                        state.fieldErrors.totpCode
                          ? "auth-login-totp-code-error"
                          : undefined
                      }
                      aria-invalid={
                        state.fieldErrors.totpCode ? true : undefined
                      }
                      autoComplete="one-time-code"
                      data-testid={phase1AuthTestId("login-totp-code")}
                      id="auth-login-totp-code"
                      inputMode="numeric"
                      maxLength={6}
                      pattern="[0-9]*"
                      type="text"
                      value={state.totpCode}
                      onChange={(event) => {
                        dispatch({
                          type: "field",
                          field: "totpCode",
                          value: normalizeTotpCode(event.target.value),
                        });
                      }}
                    />
                  </label>
                  {state.fieldErrors.totpCode ? (
                    <p
                      className="cartulary-auth-field-error"
                      id="auth-login-totp-code-error"
                      role="alert"
                    >
                      {state.fieldErrors.totpCode}
                    </p>
                  ) : null}
                </>
              ) : null}

              <button
                aria-busy={state.submitting ? true : undefined}
                aria-disabled={!canSubmit}
                className="cartulary-auth-primary-button"
                data-testid={phase1AuthTestId("login-submit")}
                disabled={!canSubmit}
                type="submit"
              >
                {submitLabel}
              </button>
            </form>
          )}

          <p className="cartulary-auth-help">
            Need account access? Contact a deployment administrator.
          </p>

          {state.enterpriseProviders.length > 0 ? (
            <section
              className="cartulary-auth-enterprise"
              aria-label="Enterprise sign-in"
            >
              <p className="cartulary-auth-enterprise-label">
                Enterprise sign-in
              </p>
              <div data-testid={phase1AuthTestId("enterprise-provider-list")}>
                {state.enterpriseProviders.map((provider) => (
                  <button
                    key={provider.provider_key}
                    className="cartulary-auth-secondary-button"
                    data-provider-key={provider.provider_key}
                    data-testid={phase1AuthTestId("enterprise-provider-button")}
                    disabled={
                      state.enterprisePendingProviderKey !== null &&
                      state.enterprisePendingProviderKey !==
                        provider.provider_key
                    }
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

          <p
            aria-live="polite"
            className="cartulary-auth-visually-hidden"
            data-testid={phase1AuthTestId("status")}
            role="status"
          >
            {statusText}
          </p>
          <p
            aria-live={authLivePoliteness}
            className="cartulary-auth-visually-hidden"
            data-error-code={activeErrorCode}
            data-testid={phase1ErrorCodeTestId("auth")}
            role={authLiveRole}
          >
            {authAlertText}
          </p>
          <div
            className="cartulary-auth-visually-hidden"
            data-error-code={activeErrorCode}
            data-testid={phase1ErrorSummaryTestIds("auth").container}
            role={authLiveRole}
          >
            <p data-testid={phase1ErrorSummaryTestIds("auth").message}>
              {authAlertText}
            </p>
            <p data-testid={phase1ErrorSummaryTestIds("auth").details} />
          </div>
        </div>
      </section>
    </main>
  );
}

const authGatewayStyleText = `
.cartulary-auth-shell {
  min-height: var(--ct-app-viewport-block-size, 100vh);
  padding: 0;
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(460px, 500px);
  background: var(--ct-colors-canvas);
  color: var(--ct-colors-ink);
  font-family: var(--ct-typography-ui-fontFamily);
  overflow: auto;
}

.cartulary-auth-hyperlegible {
  font-family: var(--ct-typography-accessible-reading-fontFamily);
}

.cartulary-auth-identity {
  position: relative;
  min-height: var(--ct-app-viewport-block-size, 100vh);
  padding: 64px;
  box-sizing: border-box;
  display: grid;
  align-content: center;
  gap: 28px;
  overflow: hidden;
}

.cartulary-auth-wordmark {
  position: absolute;
  inset-block-start: 48px;
  inset-inline-start: 64px;
  font-size: 13px;
  font-weight: 700;
  color: var(--ct-colors-ink);
}

.cartulary-auth-identity-copy {
  position: relative;
  max-width: 520px;
  z-index: 1;
}

.cartulary-auth-identity-copy p {
  margin: 0;
  font-size: 42px;
  line-height: 1.08;
  font-weight: 600;
  color: var(--ct-colors-ink);
}

.cartulary-auth-motif {
  position: absolute;
  inset: 118px 72px 72px 72px;
  opacity: 0.72;
  background-image:
    linear-gradient(var(--ct-colors-hairline) 1px, transparent 1px),
    linear-gradient(90deg, var(--ct-colors-hairline) 1px, transparent 1px);
  background-size: 56px 56px;
  mask-image: linear-gradient(90deg, transparent, black 18%, black 78%, transparent);
}

.cartulary-auth-motif span {
  position: absolute;
  inset-block-start: 168px;
  inset-inline-start: 112px;
  width: 156px;
  height: 56px;
  border: 2px solid var(--ct-colors-accent);
  border-radius: var(--ct-rounded-sm);
  border-inline-end-color: var(--ct-colors-hairline);
  border-block-end-color: var(--ct-colors-hairline);
}

.cartulary-auth-rail {
  min-height: var(--ct-app-viewport-block-size, 100vh);
  border-left: var(--ct-border-hairline);
  background: var(--ct-colors-surface-1);
  box-sizing: border-box;
  padding: 48px;
  display: grid;
  align-content: center;
}

.cartulary-auth-rail-inner {
  width: 100%;
  max-width: 392px;
  transform: translateY(-24px);
}

.cartulary-auth-eyebrow,
.cartulary-auth-enterprise-label,
.cartulary-auth-detail-label {
  margin: 0;
  font-size: 12px;
  font-weight: 700;
  color: var(--ct-colors-ink-subtle);
}

.cartulary-auth-title {
  margin: 8px 0 0;
  font-size: 32px;
  line-height: 1.18;
  font-weight: 600;
}

.cartulary-auth-support,
.cartulary-auth-help,
.cartulary-auth-setup-copy {
  margin: 10px 0 0;
  font-size: 14px;
  line-height: 1.45;
  color: var(--ct-colors-ink-muted);
}

.cartulary-auth-banner {
  margin-block-start: 24px;
  padding: 10px 12px;
  border-radius: var(--ct-rounded-sm);
  background: var(--ct-colors-surface-2);
  border: var(--ct-border-hairline);
  color: var(--ct-colors-ink);
  font-size: 14px;
}

.cartulary-auth-banner[data-tone="error"] {
  border-color: var(--ct-colors-semantic-conflict);
}

.cartulary-auth-banner[data-tone="success"] {
  border-color: var(--ct-colors-semantic-success);
}

.cartulary-auth-form,
.cartulary-auth-setup {
  margin-block-start: 28px;
  display: grid;
  gap: 16px;
}

.cartulary-auth-field {
  display: grid;
  gap: 7px;
  color: var(--ct-colors-ink-muted);
  font-size: 13px;
  font-weight: 600;
}

.cartulary-auth-field input {
  box-sizing: border-box;
  width: 100%;
  min-width: 0;
  min-height: 44px;
  padding: 9px 10px;
  border-radius: var(--ct-rounded-sm);
  border: var(--ct-border-hairline);
  background: var(--ct-colors-surface-1);
  color: var(--ct-colors-ink);
  font: inherit;
  font-weight: 400;
}

.cartulary-auth-field input[aria-invalid="true"] {
  border: 2px solid var(--ct-colors-semantic-conflict);
}

.cartulary-auth-password-control {
  position: relative;
  display: block;
}

.cartulary-auth-password-control input {
  padding-inline-end: 46px;
}

.cartulary-auth-icon-button {
  position: absolute;
  inset-block: 5px;
  inset-inline-end: 5px;
  width: 34px;
  border: 0;
  border-radius: var(--ct-rounded-sm);
  background: transparent;
  color: var(--ct-colors-ink-muted);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.cartulary-auth-primary-button,
.cartulary-auth-secondary-button {
  min-height: 44px;
  width: 100%;
  border-radius: var(--ct-rounded-md);
  padding: 10px 12px;
  font-family: var(--ct-typography-button-fontFamily);
  font-size: var(--ct-typography-button-fontSize);
  line-height: 1.2;
  font-weight: var(--ct-typography-button-fontWeight);
  cursor: pointer;
}

.cartulary-auth-primary-button {
  border: 0;
  background: var(--ct-colors-accent);
  color: var(--ct-colors-on-accent);
}

.cartulary-auth-primary-button[aria-disabled="true"],
.cartulary-auth-secondary-button[aria-disabled="true"] {
  cursor: default;
  background: var(--ct-colors-surface-2);
  color: var(--ct-colors-ink-tertiary);
  border: var(--ct-border-hairline);
}

.cartulary-auth-primary-button:disabled,
.cartulary-auth-secondary-button:disabled {
  cursor: default;
}

.cartulary-auth-secondary-button {
  border: var(--ct-border-hairline);
  background: var(--ct-colors-surface-2);
  color: var(--ct-colors-ink);
}

.cartulary-auth-field-error {
  margin: -8px 0 0;
  color: var(--ct-colors-semantic-conflict);
  font-size: 13px;
  line-height: 1.35;
}

.cartulary-auth-help {
  margin-block-start: 22px;
}

.cartulary-auth-enterprise {
  margin-block-start: 28px;
  padding-block-start: 22px;
  border-block-start: var(--ct-border-hairline);
}

.cartulary-auth-enterprise > div {
  margin-block-start: 12px;
  display: grid;
  gap: 10px;
}

.cartulary-auth-detail-list {
  display: grid;
  gap: 12px;
  padding: 12px 0;
}

.cartulary-auth-detail-list > div {
  min-width: 0;
}

.cartulary-auth-detail-list div div {
  margin-block-start: 5px;
  color: var(--ct-colors-ink-muted);
  overflow-wrap: anywhere;
}

.cartulary-auth-mono {
  font-family: var(--ct-typography-mono-fontFamily);
  font-size: 12px;
}

.cartulary-auth-visually-hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

@media (max-width: 1099px) {
  .cartulary-auth-shell {
    grid-template-columns: minmax(0, 1fr) minmax(400px, 440px);
  }

  .cartulary-auth-identity {
    padding: 48px;
  }

  .cartulary-auth-wordmark {
    inset-block-start: 36px;
    inset-inline-start: 48px;
  }

  .cartulary-auth-identity-copy p {
    font-size: 30px;
  }
}

@media (max-width: 767px) {
  .cartulary-auth-shell {
    display: block;
    min-height: var(--ct-app-viewport-block-size, 100vh);
  }

  .cartulary-auth-identity {
    min-height: auto;
    padding: 28px 20px 18px;
    display: block;
  }

  .cartulary-auth-wordmark {
    position: static;
  }

  .cartulary-auth-identity-copy {
    margin-block-start: 18px;
  }

  .cartulary-auth-identity-copy p {
    font-size: 24px;
    line-height: 1.18;
  }

  .cartulary-auth-motif {
    display: none;
  }

  .cartulary-auth-rail {
    min-height: auto;
    border-left: 0;
    border-top: var(--ct-border-hairline);
    padding: 30px 20px 42px;
    display: block;
  }

  .cartulary-auth-rail-inner {
    max-width: none;
    transform: none;
  }

  .cartulary-auth-title {
    font-size: 28px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .cartulary-auth-rail-inner {
    transform: none;
  }
}
`;

export type { Phase1AuthSurfaceProps };
