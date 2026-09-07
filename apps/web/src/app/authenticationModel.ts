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
import type { SessionObservation } from "./appSessionController";

export type EnterpriseAuthNavigation = {
  assign: (url: string) => void;
  returnTo: () => string;
};
type AuthEffects = {
  admitTransport: () => (() => void) | null;
  canAuthenticate: () => boolean;
  current: () => boolean;
  authenticated: (
    session: SessionData,
    signal: AbortSignal,
    current: () => boolean,
  ) => Promise<void> | void;
  inspectSession: (
    signal: AbortSignal,
    current: () => boolean,
  ) => Promise<SessionObservation>;
};
type AuthBanner = { message: string; tone: "error" | "info" | "success" };
type FieldErrors = {
  username?: string;
  password?: string;
  totpCode?: string;
  bootstrapCompleteCode?: string;
};
type Credentials = Readonly<{
  kind: "credentials";
  username: string;
  password: string;
  passwordVisible: boolean;
}>;
type MFA = Readonly<{
  kind: "mfa";
  username: string;
  password: string;
  passwordVisible: boolean;
  code: string;
  expiresAt: number;
}>;
type Enrollment = Readonly<{
  kind: "setup";
  username: string;
  code: string;
  expiresAt: number;
  enrollment:
    | { readonly kind: "unissued" }
    | { readonly kind: "ready"; readonly key: string };
}>;
type Flow = Credentials | MFA | Enrollment;
type Action = "login" | "enterprise" | "begin" | "complete";
type Operation =
  | { readonly kind: "idle" }
  | { readonly kind: "pending" | "uncertain"; readonly action: Action }
  | {
      readonly kind: "confirmed";
      readonly propagation: "pending" | "failed" | "ready";
    };
type State = Readonly<{
  flow: Flow;
  operation: Operation;
  sessionRead: "idle" | "pending" | "absent" | "unavailable";
  transportPending: boolean;
  banner: AuthBanner | null;
  error: APIError | null;
  fieldErrors: FieldErrors;
  enterpriseProviders: readonly EnterpriseAuthProvider[];
  providersStatus: "loading" | "ready" | "failed";
}>;
const credentials = (username = ""): Credentials => ({
  kind: "credentials",
  username,
  password: "",
  passwordVisible: false,
});
const initial = (): State => ({
  flow: credentials(),
  operation: { kind: "idle" },
  sessionRead: "idle",
  transportPending: false,
  banner: null,
  error: null,
  fieldErrors: {},
  enterpriseProviders: [],
  providersStatus: "loading",
});

export class AuthenticationController {
  private state = initial();
  private readonly listeners = new Set<() => void>();
  private epoch = 0;
  private active = false;
  private disposed = false;
  private operation: object | null = null;
  private transport: object | null = null;
  private sessionRead: object | null = null;
  private cancelOperation: (() => void) | null = null;
  private cancelDiscovery: (() => void) | null = null;
  private cancelSessionRead: (() => void) | null = null;
  private expiry: ReturnType<typeof setTimeout> | undefined;
  private expiresAt = 0;
  private discovery = 0;
  // Bootstrap authorization is protocol state, never a presentation or replay field.
  private bootstrapToken = "";
  private enrollmentId = "";
  constructor(
    private readonly capture: () => AuthEffects,
    private readonly navigation: EnterpriseAuthNavigation,
    private readonly api = {
      loginLocal,
      beginEnterpriseAuth,
      beginTotpEnrollment,
      completeTotpEnrollment,
      listEnterpriseAuthProviders,
    },
  ) {}
  getSnapshot = () => this.state;
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  private publish(patch: Partial<State>) {
    if (this.disposed) return;
    this.state = { ...this.state, ...patch };
    for (const listener of this.listeners) listener();
  }
  canStart = () =>
    this.active &&
    !this.disposed &&
    this.operation === null &&
    this.transport === null &&
    this.sessionRead === null &&
    !(
      this.state.operation.kind === "confirmed" &&
      this.state.operation.propagation !== "ready"
    ) &&
    this.capture().canAuthenticate();
  canInspectSession = () =>
    this.active &&
    !this.disposed &&
    this.sessionRead === null &&
    (this.transport !== null ||
      this.state.operation.kind === "uncertain" ||
      (this.state.operation.kind === "confirmed" &&
        this.state.operation.propagation === "failed"));
  open = () => {
    if (!this.disposed) this.active = true;
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
    this.cancelSessionRead?.();
    this.operation = null;
    this.sessionRead = null;
    this.cancelSessionRead = null;
    this.bootstrapToken = "";
    this.enrollmentId = "";
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
                "An earlier sign-in request is still settling. Check the current session or reload if it does not finish.",
            },
    });
  };
  dispose = () => {
    if (this.disposed) return;
    this.close();
    this.disposed = true;
    this.listeners.clear();
  };
  private clearSecrets() {
    clearTimeout(this.expiry);
    this.expiresAt = 0;
    this.bootstrapToken = "";
    this.enrollmentId = "";
    this.publish({ flow: credentials(this.state.flow.username) });
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
    if (!this.active || this.disposed) return;
    if (action.type === "use_different_account") {
      this.retire();
      return;
    }
    const flow = this.state.flow;
    if (action.type === "toggle_password_visibility") {
      if (flow.kind !== "setup")
        this.publish({
          flow: { ...flow, passwordVisible: !flow.passwordVisible },
        });
      return;
    }
    if (this.operation || this.transport || this.sessionRead) return;
    let next: Flow;
    switch (action.field) {
      case "username":
        if (flow.kind !== "credentials" && action.value !== flow.username) {
          this.retire();
          next = credentials(action.value);
        } else next = { ...flow, username: action.value };
        break;
      case "password":
        if (flow.kind === "setup") return;
        next = { ...flow, password: action.value };
        break;
      case "totpCode":
        if (flow.kind !== "mfa") return;
        next = { ...flow, code: action.value };
        break;
      case "bootstrapCompleteCode":
        if (flow.kind !== "setup") return;
        next = { ...flow, code: action.value };
        break;
    }
    this.publish({
      flow: next,
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
    action: Action,
    request: (signal: AbortSignal) => Promise<HTTPOperationResult<T>>,
    completed: (
      result: HTTPOperationResult<T>,
      effects: AuthEffects,
      current: () => boolean,
    ) => Promise<void> | void,
  ) {
    if (!this.canStart()) return;
    const effects = this.capture();
    if (!effects.current()) return;
    const release = effects.admitTransport();
    if (!release) {
      this.publish({
        banner: {
          tone: "info",
          message:
            "An earlier authentication request is still settling. Check the current session or reload if it does not finish.",
        },
      });
      return;
    }
    const token = {};
    const epoch = this.epoch;
    this.operation = token;
    this.transport = token;
    const current = () =>
      this.active &&
      !this.disposed &&
      this.epoch === epoch &&
      this.operation === token &&
      effects.current();
    this.publish({
      error: null,
      banner: null,
      fieldErrors: {},
      operation: { kind: "pending", action },
      transportPending: true,
    });
    const observation = observeAccountOperation(request);
    this.cancelOperation = observation.cancel;
    void observation.settled.then(() => {
      release();
      if (this.transport === token) {
        this.transport = null;
        this.publish({ transportPending: false });
      }
    });
    try {
      const outcome = await observation.result;
      if (!current()) return;
      if (
        outcome.kind !== "completed" ||
        (!outcome.value.ok &&
          (outcome.value.status >= 500 ||
            outcome.value.payload.error?.code ===
              "invalid_public_contract_response"))
      ) {
        this.clearSecrets();
        this.publish({
          operation: { kind: "uncertain", action },
          banner: {
            tone: "error",
            message:
              "Authentication outcome is uncertain. Check the current session. An outstanding request must settle before another sign-in; reload if it does not finish.",
          },
        });
        if (action === "login") void this.retrySession();
        return;
      }
      this.publish({ operation: { kind: "idle" } });
      await completed(outcome.value, effects, current);
    } catch {
      if (current()) {
        this.clearSecrets();
        this.publish({
          operation: { kind: "uncertain", action },
          banner: {
            tone: "error",
            message:
              "Authentication could not be completed. Check the current session before signing in again.",
          },
        });
      }
    } finally {
      if (this.operation === token) {
        this.operation = null;
        this.cancelOperation = null;
        this.publish({});
      }
    }
  }
  retrySession = async () => {
    if (!this.canInspectSession()) return;
    const token = {};
    const epoch = this.epoch;
    const effects = this.capture();
    const current = () =>
      this.active &&
      !this.disposed &&
      this.epoch === epoch &&
      this.sessionRead === token &&
      effects.current();
    if (!effects.current()) return;
    this.sessionRead = token;
    this.publish({ sessionRead: "pending" });
    const observation = observeAccountOperation((signal) =>
      effects.inspectSession(signal, () => current() && !signal.aborted),
    );
    this.cancelSessionRead = observation.cancel;
    const outcome = await observation.result;
    if (!current()) return;
    const result = outcome.kind === "completed" ? outcome.value : null;
    const accepted = result?.kind === "accepted";
    const confirmed = this.state.operation.kind === "confirmed";
    this.sessionRead = null;
    this.cancelSessionRead = null;
    this.publish({
      sessionRead: accepted
        ? "idle"
        : result?.kind === "session_lost"
          ? "absent"
          : "unavailable",
      ...(confirmed
        ? {
            operation: {
              kind: "confirmed",
              propagation: accepted ? "ready" : "failed",
            } as const,
          }
        : {}),
      banner: accepted
        ? null
        : {
            tone: "error",
            message: confirmed
              ? "Sign-in was confirmed, but the current session remains unavailable. Retry the session check or reload."
              : "The earlier authentication remains unconfirmed. Check the current session again or enter fresh credentials after the request settles; reload if it does not finish.",
          },
    });
  };
  login = () => {
    if (!this.canStart()) return;
    const flow = this.state.flow;
    if (flow.kind === "setup") return;
    const errors: FieldErrors = {};
    const email = validateAccountEmail(flow.username);
    if (email.error) errors.username = email.error;
    if (!flow.password) errors.password = "Enter your password.";
    if (flow.kind === "mfa" && !/^\d{6}$/.test(flow.code))
      errors.totpCode = "Enter a six-digit authenticator code.";
    if (Object.keys(errors).length) {
      this.publish({ fieldErrors: errors, banner: null });
      return;
    }
    if (flow.kind === "credentials" && !this.expireAt(Date.now() + 300_000))
      return;
    if (this.expiresAt && Date.now() >= this.expiresAt) return;
    const deadline = this.expiresAt;
    const input = {
      username: email.value,
      password: flow.password,
      ...(flow.kind === "mfa" ? { secondFactorCode: flow.code } : {}),
    };
    if (flow.kind === "mfa") this.publish({ flow: { ...flow, code: "" } });
    void this.execute(
      "login",
      (signal) => this.api.loginLocal({ ...input, signal }),
      async (result, effects, current) => {
        if (result.ok) {
          this.clearSecrets();
          this.publish({
            banner: null,
            operation: { kind: "confirmed", propagation: "pending" },
          });
          const publication = observeAccountOperation((signal) =>
            Promise.resolve(
              effects.authenticated(
                result.payload.data,
                signal,
                () => current() && !signal.aborted,
              ),
            ),
          );
          this.cancelOperation = publication.cancel;
          const outcome = await publication.result;
          if (!current()) return;
          this.publish({
            operation: {
              kind: "confirmed",
              propagation: outcome.kind === "completed" ? "ready" : "failed",
            },
            banner:
              outcome.kind === "completed"
                ? null
                : {
                    tone: "error",
                    message:
                      "Sign-in succeeded, but the current session could not be published. Retry the session check.",
                  },
          });
          return;
        }
        const error = result.payload.error;
        if (error?.code === "mfa_required") {
          this.publish({
            flow: {
              kind: "mfa",
              username: email.value,
              password: input.password,
              passwordVisible: false,
              code: "",
              expiresAt: deadline,
            },
            error: null,
            banner: null,
          });
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
          this.bootstrapToken = token;
          this.publish({
            flow: {
              kind: "setup",
              username: email.value,
              code: "",
              expiresAt: expires,
              enrollment: { kind: "unissued" },
            },
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
          this.state.flow.kind === "mfa"
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
          error: accountOperationError(error),
          banner: authBannerForError(error ?? null),
        });
      },
    );
  };
  enterprise = (providerKey: string) => {
    if (
      !this.canStart() ||
      !this.state.enterpriseProviders.some(
        (p) => p.provider_key === providerKey,
      )
    )
      return;
    this.clearSecrets();
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
    );
  };
  begin = () => {
    const flow = this.state.flow;
    if (
      flow.kind !== "setup" ||
      !this.bootstrapToken ||
      flow.enrollment.kind !== "unissued"
    )
      return;
    const token = this.bootstrapToken;
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
            error: accountOperationError(result.payload.error),
            banner: authBannerForError(result.payload.error ?? null),
          });
          return;
        }
        const data = result.payload.data;
        const expires = Math.min(flow.expiresAt, Date.parse(data.expires_at));
        if (!this.expireAt(expires)) return;
        this.enrollmentId = data.enrollment_id;
        this.publish({
          flow: {
            ...flow,
            expiresAt: expires,
            enrollment: {
              kind: "ready",
              key: data.totp_setup.secret_base32,
            },
          },
        });
      },
    );
  };
  complete = () => {
    const flow = this.state.flow;
    if (
      !this.canStart() ||
      flow.kind !== "setup" ||
      flow.enrollment.kind !== "ready"
    )
      return;
    if (!/^\d{6}$/.test(flow.code)) {
      this.publish({
        fieldErrors: {
          bootstrapCompleteCode: "Enter a six-digit authenticator code.",
        },
      });
      return;
    }
    const input = {
      authMode: "bootstrap" as const,
      bootstrapToken: this.bootstrapToken,
      enrollmentId: this.enrollmentId,
      code: flow.code,
      clientTxnId: clientTxnID("bootstrap-complete"),
    };
    this.publish({ flow: { ...flow, code: "" } });
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
            error: accountOperationError(result.payload.error),
            banner: authBannerForError(result.payload.error ?? null),
          });
          return;
        }
        this.clearSecrets();
        this.publish({
          banner: {
            tone: "success",
            message: "Authenticator setup is complete. Sign in again.",
          },
        });
      },
    );
  };
}
export function authBannerForError(error: APIError | null): AuthBanner {
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
