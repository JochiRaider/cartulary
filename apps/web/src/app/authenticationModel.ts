import {
  type FormEvent,
  useEffect,
  useLayoutEffect,
  useRef,
  useSyncExternalStore,
} from "react";
import {
  type APIError,
  clientTxnID,
  type HTTPOperationResult,
} from "../services/browserApi";
import { validateAccountEmail } from "./accountInputValidation";
import {
  accountOperationError,
  observeAccountOperation,
} from "./accountOperation";
import {
  beginEnterpriseAuth,
  beginTotpEnrollment,
  completeTotpEnrollment,
  listEnterpriseAuthProviders,
  loginLocal,
} from "./api/authAccountClient";
import type {
  EnterpriseAuthProvider,
  SessionData,
} from "./api/publicHttpTypes";

type AuthSurfaceBootstrapState =
  | "loading"
  | "anonymous"
  | "revoked"
  | "public_error_envelope";
type AuthChallengeState = "mfa_required" | "mfa_setup_required";
export type EnterpriseAuthNavigation = {
  assign: (url: string) => void;
  returnTo: () => string;
};
type AuthEffects = {
  admitTransport?: () => (() => void) | null;
  canAuthenticate?: () => boolean;
  current: () => boolean;
  authenticated: (session: SessionData) => Promise<void> | void;
  uncertain: () => Promise<boolean>;
};
export type AuthGatewayProps = {
  controller: AuthenticationController;
  bootstrapState: AuthSurfaceBootstrapState;
  message: string;
  publicError?: APIError | null;
  readingProfile?: "default" | "hyperlegible" | undefined;
};
type AuthPhase = "credentials" | "mfa" | "setup";
type AuthBanner = { message: string; tone: "error" | "info" | "success" };
type FieldErrors = {
  bootstrapCompleteCode?: string;
  password?: string;
  totpCode?: string;
  username?: string;
};
type AuthState = {
  confirmation: "idle" | "pending" | "failed";
  banner: AuthBanner | null;
  bootstrapCompleteCode: string;
  bootstrapEnrollmentId: string;
  bootstrapSecretBase32: string;
  bootstrapToken: string;
  enterprisePendingProviderKey: string | null;
  enterpriseProviders: EnterpriseAuthProvider[];
  providersStatus: "loading" | "ready" | "failed";
  error: APIError | null;
  fieldErrors: FieldErrors;
  password: string;
  passwordVisible: boolean;
  phase: AuthPhase;
  setupAction: "beginning" | "completing" | "idle";
  submitting: boolean;
  transportPending: boolean;
  totpCode: string;
  username: string;
};
const emptySecrets = {
  password: "",
  totpCode: "",
  bootstrapCompleteCode: "",
  bootstrapEnrollmentId: "",
  bootstrapSecretBase32: "",
  bootstrapToken: "",
  passwordVisible: false,
};
const initial = (): AuthState => ({
  confirmation: "idle",
  ...emptySecrets,
  banner: null,
  error: null,
  fieldErrors: {},
  phase: "credentials",
  setupAction: "idle",
  submitting: false,
  transportPending: false,
  username: "",
  enterprisePendingProviderKey: null,
  enterpriseProviders: [],
  providersStatus: "loading",
});
const defaultNavigation: EnterpriseAuthNavigation = {
  assign: (url) => window.location.assign(url),
  returnTo: () => `${window.location.pathname}${window.location.search}` || "/",
};

export class AuthenticationController {
  private state = initial();
  private readonly listeners = new Set<() => void>();
  private epoch = 0;
  private active = false;
  private operation: object | null = null;
  private transport: object | null = null;
  private cancelOperation: (() => void) | null = null;
  private cancelDiscovery: (() => void) | null = null;
  private expiry: ReturnType<typeof setTimeout> | undefined;
  private expiresAt = 0;
  private discovery = 0;
  constructor(
    private readonly capture: () => AuthEffects,
    private readonly navigation: EnterpriseAuthNavigation = defaultNavigation,
    private readonly api = {
      loginLocal,
      beginEnterpriseAuth,
      beginTotpEnrollment,
      completeTotpEnrollment,
      listEnterpriseAuthProviders,
    },
  ) {}
  getSnapshot = () => this.state;
  canStart = () =>
    this.active &&
    this.operation === null &&
    this.transport === null &&
    this.state.confirmation === "idle" &&
    (this.capture().canAuthenticate?.() ?? true);
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  private publish(patch: Partial<AuthState>) {
    this.state = { ...this.state, ...patch };
    for (const listener of this.listeners) listener();
  }
  open = () => {
    this.active = true;
  };
  close = () => {
    this.active = false;
    this.retire();
  };
  retire = () => {
    ++this.epoch;
    ++this.discovery;
    this.cancelOperation?.();
    this.cancelDiscovery?.();
    this.operation = null;
    clearTimeout(this.expiry);
    this.expiresAt = 0;
    this.publish({
      ...initial(),
      transportPending: this.transport !== null,
      banner:
        this.transport === null
          ? null
          : {
              tone: "info",
              message:
                "An earlier sign-in request is still settling. Reload if it does not finish.",
            },
    });
  };
  dispose = () => {
    this.close();
    this.listeners.clear();
  };
  private clearSecrets() {
    clearTimeout(this.expiry);
    this.expiresAt = 0;
    this.publish(emptySecrets);
  }
  private expireAt(deadline: number) {
    clearTimeout(this.expiry);
    this.expiresAt = deadline;
    const expire = () => {
      this.retire();
      this.publish({
        banner: {
          tone: "info",
          message: "Authentication expired. Enter your credentials again.",
        },
      });
    };
    if (!Number.isFinite(deadline) || deadline <= Date.now()) {
      expire();
      return false;
    }
    this.expiry = setTimeout(
      expire,
      Math.min(deadline - Date.now(), 2_147_483_647),
    );
    return true;
  }
  dispatch = (
    action:
      | {
          type: "field";
          field: "username" | "password" | "totpCode" | "bootstrapCompleteCode";
          value: string;
        }
      | { type: "toggle_password_visibility" }
      | { type: "use_different_account" },
  ) => {
    if (!this.active) return;
    if (action.type === "use_different_account") {
      this.retire();
      return;
    }
    if (action.type === "toggle_password_visibility") {
      this.publish({ passwordVisible: !this.state.passwordVisible });
      return;
    }
    if (this.operation !== null || this.transport !== null) return;
    if (
      action.field === "username" &&
      action.value !== this.state.username &&
      this.state.phase !== "credentials"
    ) {
      this.retire();
    }
    this.publish({
      [action.field]: action.value,
      error: null,
      banner: null,
      fieldErrors: { ...this.state.fieldErrors, [action.field]: undefined },
    });
  };
  discover = async () => {
    if (!this.active) return;
    this.cancelDiscovery?.();
    const request = ++this.discovery;
    const epoch = this.epoch;
    this.publish({ providersStatus: "loading" });
    const observation = observeAccountOperation((signal) =>
      this.api.listEnterpriseAuthProviders({ signal }),
    );
    this.cancelDiscovery = observation.cancel;
    const outcome = await observation.result;
    if (!this.active || request !== this.discovery || epoch !== this.epoch)
      return;
    if (outcome.kind === "completed") {
      const result = outcome.value;
      if (result.ok) {
        this.publish({
          providersStatus: "ready",
          enterpriseProviders: result.payload.data.providers,
        });
        return;
      }
      if (result.payload.error?.code === "extension_profile_not_claimed") {
        this.publish({ providersStatus: "ready", enterpriseProviders: [] });
        return;
      }
    }
    this.publish({ providersStatus: "failed" });
  };
  private async execute<T>(
    kind: "login" | "enterprise" | "begin" | "complete",
    request: (signal: AbortSignal) => Promise<HTTPOperationResult<T>>,
    completed: (
      result: HTTPOperationResult<T>,
      effects: AuthEffects,
      current: () => boolean,
    ) => Promise<void> | void,
    providerKey?: string,
  ) {
    if (!this.active || this.operation !== null || this.transport !== null)
      return;
    const effects = this.capture();
    if (!effects.current()) return;
    const identity = {};
    const epoch = this.epoch;
    this.operation = identity;
    this.transport = identity;
    const releaseTransport = effects.admitTransport?.();
    if (effects.admitTransport && !releaseTransport) {
      this.operation = null;
      this.transport = null;
      this.publish({
        banner: {
          tone: "info",
          message:
            "An earlier authentication request is still settling. Reload if it does not finish.",
        },
      });
      return;
    }
    const current = () =>
      this.active &&
      this.epoch === epoch &&
      this.operation === identity &&
      effects.current();
    this.publish({
      error: null,
      banner: null,
      fieldErrors: {},
      submitting: kind === "login",
      setupAction:
        kind === "begin"
          ? "beginning"
          : kind === "complete"
            ? "completing"
            : "idle",
      enterprisePendingProviderKey: providerKey ?? null,
      transportPending: true,
    });
    const observation = observeAccountOperation(request);
    this.cancelOperation = observation.cancel;
    void observation.settled.then(() => {
      releaseTransport?.();
      if (this.transport === identity) {
        this.transport = null;
        this.publish({ transportPending: false });
      }
    });
    try {
      const outcome = await observation.result;
      if (!current()) return;
      this.publish({
        submitting: false,
        setupAction: "idle",
        enterprisePendingProviderKey: null,
      });
      if (
        outcome.kind !== "completed" ||
        (!outcome.value.ok &&
          (outcome.value.status >= 500 ||
            outcome.value.payload.error?.code ===
              "invalid_public_contract_response"))
      ) {
        this.clearSecrets();
        this.publish({
          phase: "credentials",
          banner: {
            tone: "error",
            message:
              kind === "login"
                ? "Sign-in response could not be confirmed. Checking the current session. An outstanding request must settle before another sign-in; reload if it does not finish."
                : "Authentication outcome is uncertain. Start again with fresh credentials after the request settles; reload if it does not finish.",
          },
        });
        if (kind === "login") {
          let accepted = false;
          try {
            accepted = await effects.uncertain();
          } catch {
            /* Keep uncertainty visible. */
          }
          if (!current()) return;
          if (accepted) this.publish({ banner: null });
          else
            this.publish({
              banner: {
                tone: "error",
                message:
                  "Sign-in response could not be confirmed. Try again after the request settles; reload if it does not finish.",
              },
            });
        }
        return;
      }
      await completed(outcome.value, effects, current);
    } catch {
      if (current()) {
        this.clearSecrets();
        this.publish({
          phase: "credentials",
          banner: {
            tone: "error",
            message:
              "Authentication could not be completed. Check the current session before signing in again.",
          },
        });
      }
    } finally {
      if (this.operation === identity) {
        this.operation = null;
        this.cancelOperation = null;
        this.publish({
          submitting: false,
          setupAction: "idle",
          enterprisePendingProviderKey: null,
        });
      }
    }
  }
  retrySession = async () => {
    if (
      !this.active ||
      this.operation ||
      this.transport ||
      this.state.confirmation !== "failed"
    )
      return;
    const epoch = this.epoch;
    const effects = this.capture();
    const current = () =>
      this.active && this.epoch === epoch && effects.current();
    if (!current()) return;
    this.publish({ confirmation: "pending" });
    const observation = observeAccountOperation(() => effects.uncertain());
    this.cancelOperation = observation.cancel;
    const outcome = await observation.result;
    if (!current()) return;
    const accepted = outcome.kind === "completed" && outcome.value;
    this.publish({
      confirmation: accepted ? "idle" : "failed",
      banner: accepted
        ? null
        : {
            tone: "error",
            message:
              "Sign-in was confirmed, but the current session remains unavailable. Retry the session check or reload.",
          },
    });
  };
  login = () => {
    if (!this.canStart()) return;
    const state = this.state;
    const fieldErrors: FieldErrors = {};
    const email = validateAccountEmail(state.username);
    if (email.error) fieldErrors.username = email.error;
    if (!state.password) fieldErrors.password = "Enter your password.";
    if (state.phase === "mfa" && !/^\d{6}$/.test(state.totpCode))
      fieldErrors.totpCode = "Enter a six-digit authenticator code.";
    if (Object.keys(fieldErrors).length) {
      this.publish({ fieldErrors, banner: null });
      return;
    }
    if (state.phase === "credentials" && !this.expireAt(Date.now() + 300_000))
      return;
    if (
      state.phase === "setup" ||
      (this.expiresAt && Date.now() >= this.expiresAt)
    )
      return;
    const input = {
      username: email.value,
      password: state.password,
      ...(state.phase === "mfa" ? { secondFactorCode: state.totpCode } : {}),
    };
    this.publish({ totpCode: "" });
    void this.execute(
      "login",
      (signal) => this.api.loginLocal({ ...input, signal }),
      async (result, effects, current) => {
        if (result.ok) {
          this.clearSecrets();
          this.publish({ banner: null });
          this.publish({ confirmation: "pending" });
          try {
            await effects.authenticated(result.payload.data);
          } catch {
            if (current())
              this.publish({
                confirmation: "failed",
                banner: {
                  tone: "error",
                  message:
                    "Sign-in succeeded, but the current session could not be published. Retry the session check.",
                },
              });
            return;
          }
          if (current()) this.publish({ confirmation: "idle" });
          return;
        }
        const error = result.payload.error;
        if (error?.code === "mfa_required") {
          this.publish({ phase: "mfa", error: null, banner: null });
          return;
        }
        if (error?.code === "mfa_setup_required") {
          const token = error.details?.bootstrap_token;
          const expires = Date.parse(
            String(error.details?.bootstrap_expires_at ?? ""),
          );
          this.clearSecrets();
          if (typeof token !== "string" || !token || !this.expireAt(expires))
            return;
          this.publish({
            phase: "setup",
            bootstrapToken: token,
            error: accountOperationError(error),
            banner: {
              tone: "info",
              message: "Authenticator setup is required before sign-in.",
            },
          });
          return;
        }
        if (
          error?.code === "invalid_second_factor" &&
          this.state.phase === "mfa"
        ) {
          this.publish({
            fieldErrors: {
              totpCode: "The verification code is incorrect or expired.",
            },
          });
          return;
        }
        this.clearSecrets();
        this.publish({
          phase: "credentials",
          error: accountOperationError(error),
          banner: authBannerForError(error ?? null),
        });
      },
    );
  };
  enterprise = (providerKey: string) => {
    if (!this.canStart()) return;
    if (
      !this.state.enterpriseProviders.some(
        (p) => p.provider_key === providerKey,
      )
    )
      return;
    this.clearSecrets();
    this.publish({ phase: "credentials" });
    void this.execute(
      "enterprise",
      (signal) =>
        this.api.beginEnterpriseAuth({
          providerKey,
          returnTo: this.navigation.returnTo(),
          signal,
        }),
      (result, _effects, current) => {
        if (!result.ok) {
          this.publish({
            error: accountOperationError(result.payload.error),
            banner: {
              tone: "error",
              message: "Enterprise sign-in could not be started.",
            },
          });
          return;
        }
        if (current()) this.navigation.assign(result.payload.data.redirect_url);
      },
      providerKey,
    );
  };
  begin = () => {
    if (
      this.state.phase !== "setup" ||
      !this.state.bootstrapToken ||
      this.state.bootstrapEnrollmentId
    )
      return;
    const token = this.state.bootstrapToken;
    void this.execute(
      "begin",
      (signal) =>
        this.api.beginTotpEnrollment({
          authMode: "bootstrap",
          bootstrapToken: token,
          clientTxnId: clientTxnID("bootstrap-begin"),
          signal,
        }),
      (result) => {
        if (!result.ok) {
          this.clearSecrets();
          this.publish({
            phase: "credentials",
            error: accountOperationError(result.payload.error),
            banner: authBannerForError(result.payload.error ?? null),
          });
          return;
        }
        const data = result.payload.data;
        if (
          !this.expireAt(Math.min(this.expiresAt, Date.parse(data.expires_at)))
        )
          return;
        this.publish({
          bootstrapEnrollmentId: data.enrollment_id,
          bootstrapSecretBase32: data.totp_setup.secret_base32,
        });
      },
    );
  };
  complete = () => {
    if (
      this.state.phase !== "setup" ||
      !this.state.bootstrapEnrollmentId ||
      this.operation ||
      this.transport
    )
      return;
    if (!/^\d{6}$/.test(this.state.bootstrapCompleteCode)) {
      this.publish({
        fieldErrors: {
          bootstrapCompleteCode: "Enter a six-digit authenticator code.",
        },
      });
      return;
    }
    const input = {
      authMode: "bootstrap" as const,
      bootstrapToken: this.state.bootstrapToken,
      enrollmentId: this.state.bootstrapEnrollmentId,
      code: this.state.bootstrapCompleteCode,
      clientTxnId: clientTxnID("bootstrap-complete"),
    };
    this.publish({ bootstrapCompleteCode: "" });
    void this.execute(
      "complete",
      (signal) => this.api.completeTotpEnrollment({ ...input, signal }),
      (result) => {
        if (!result.ok) {
          if (result.payload.error?.code === "invalid_second_factor") {
            this.publish({
              error: accountOperationError(result.payload.error),
              fieldErrors: {
                bootstrapCompleteCode:
                  "The verification code is incorrect or expired.",
              },
            });
            return;
          }
          this.clearSecrets();
          this.publish({
            phase: "credentials",
            error: accountOperationError(result.payload.error),
            banner: authBannerForError(result.payload.error ?? null),
          });
          return;
        }
        this.clearSecrets();
        this.publish({
          phase: "credentials",
          banner: {
            tone: "success",
            message: "Authenticator setup is complete. Sign in again.",
          },
        });
      },
    );
  };
}
export function normalizeTotpCode(value: string) {
  return value.replace(/\D/gu, "").slice(0, 6);
}
function authBannerForError(error: APIError | null): AuthBanner {
  return {
    tone: "error",
    message:
      error?.code === "invalid_credentials"
        ? "Email or password is incorrect."
        : error?.code === "invalid_auth_request"
          ? "Sign-in request could not be completed."
          : [
                "session_required",
                "auth_required",
                "credential_bootstrap_rejected",
              ].includes(error?.code ?? "")
            ? "Sign in again to continue."
            : "Authentication is temporarily unavailable. Try again.",
  };
}
function firstFieldError(errors: FieldErrors) {
  return (
    errors.username ??
    errors.password ??
    errors.totpCode ??
    errors.bootstrapCompleteCode ??
    ""
  );
}
export function useAuthentication({
  controller,
  bootstrapState,
  message,
  publicError = null,
  readingProfile = "default",
}: AuthGatewayProps) {
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  const usernameRef = useRef<HTMLInputElement>(null);
  const passwordRef = useRef<HTMLInputElement>(null);
  const totpCodeRef = useRef<HTMLInputElement>(null);
  const bootstrapBeginRef = useRef<HTMLButtonElement>(null);
  const bootstrapCompleteCodeRef = useRef<HTMLInputElement>(null);
  useLayoutEffect(() => {
    controller.open();
    return controller.close;
  }, [controller]);
  useEffect(() => {
    if (bootstrapState !== "loading") void controller.discover();
  }, [controller, bootstrapState]);
  useEffect(() => {
    if (state.fieldErrors.username) usernameRef.current?.focus();
    else if (state.fieldErrors.password) passwordRef.current?.focus();
    else if (state.fieldErrors.totpCode) totpCodeRef.current?.focus();
    else if (state.fieldErrors.bootstrapCompleteCode)
      bootstrapCompleteCodeRef.current?.focus();
    else if (state.phase === "mfa") totpCodeRef.current?.focus();
    else if (state.phase === "setup") {
      if (!state.bootstrapEnrollmentId) bootstrapBeginRef.current?.focus();
      else bootstrapCompleteCodeRef.current?.focus();
    }
  }, [state.fieldErrors, state.phase, state.bootstrapEnrollmentId]);
  const dispatch = controller.dispatch;
  const handleLoginSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    controller.login();
  };
  const handleEnterpriseBegin = controller.enterprise;
  const handleBeginBootstrapEnrollment = controller.begin;
  const handleCompleteBootstrapEnrollment = controller.complete;
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
  const authLivePoliteness: "assertive" | "polite" =
    authLiveRole === "alert" ? "assertive" : "polite";
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
  const canSubmit = controller.canStart() && bootstrapState !== "loading";

  return {
    state,
    dispatch,
    usernameRef,
    passwordRef,
    totpCodeRef,
    bootstrapBeginRef,
    bootstrapCompleteCodeRef,
    handleLoginSubmit,
    handleEnterpriseBegin,
    handleBeginBootstrapEnrollment,
    handleCompleteBootstrapEnrollment,
    displayedBootstrapState,
    displayedBanner,
    authAlertText,
    authLiveRole,
    authLivePoliteness,
    activeErrorCode,
    statusText,
    rootClassName,
    supportText,
    title,
    submitLabel,
    canSubmit,
  };
}
