import {
  type APIError,
  clientTxnID,
  type HTTPOperationResult,
} from "../services/browserApi";
import { validateProvisioningPassword } from "./accountInputValidation";
import {
  accountOperationError,
  observeAccountOperation,
} from "./accountOperation";
import {
  beginTotpEnrollment,
  changePassword,
  completeTotpEnrollment,
  loadCredentialState,
  logoutCurrentSession,
} from "./api/authAccountClient";
import type { CredentialState } from "./api/publicHttpTypes";

type AccountSessionEvent =
  | { readonly kind: "resource_refresh" }
  | {
      readonly kind:
        | "logout_confirmed"
        | "credentials_revoked"
        | "session_lost";
      readonly message: string;
    };
type SecurityEffects = {
  admitLogout: () => (() => void) | null;
  actor: string;
  current: () => boolean;
  event: (
    event: AccountSessionEvent,
    signal: AbortSignal,
    current: () => boolean,
  ) => Promise<void> | void;
};

type Action = "password" | "begin" | "complete" | "logout";
type Operation =
  | { kind: "idle" }
  | { kind: "pending" | "uncertain" | "rejected"; action: Action }
  | {
      kind: "confirmed";
      action: Action;
      propagation: "ready" | "pending" | "failed";
    };
const emptySecrets = {
  passwordCurrent: "",
  passwordNext: "",
  passwordFactorCode: "",
  totpCurrentPassword: "",
  totpCurrentFactorCode: "",
  totpCompleteCode: "",
};
type SecretField = keyof typeof emptySecrets;
type SecurityState = typeof emptySecrets & {
  enrollment: Readonly<{ key: string; expiresAt: number }> | null;
  credentialState: CredentialState | null;
  credentialRead: "loading" | "ready" | "failed";
  credentialStateError: APIError | null;
  error: APIError | null;
  statusText: string;
  fieldErrors: Partial<Record<SecretField, string>>;
  operation: Operation;
  transportPending: boolean;
};
const initial = (): SecurityState => ({
  ...emptySecrets,
  enrollment: null,
  credentialState: null,
  credentialRead: "loading",
  credentialStateError: null,
  error: null,
  statusText: "Loading account security.",
  fieldErrors: {},
  operation: { kind: "idle" },
  transportPending: false,
});
export class AccountSecurityController {
  private state = initial();
  private readonly listeners = new Set<() => void>();
  private epoch = 0;
  private form = 0;
  private active = false;
  private read = 0;
  private operation: object | null = null;
  private transport: object | null = null;
  private logoutTransport: object | null = null;
  private readonly transports = new Set<object>();
  private disposed = false;
  private enrollmentId = "";
  private cancelRead: (() => void) | null = null;
  private readonly observations = new Set<() => void>();
  private expiry: ReturnType<typeof setTimeout> | undefined;
  constructor(
    private readonly capture: () => SecurityEffects,
    private readonly api = {
      beginTotpEnrollment,
      changePassword,
      completeTotpEnrollment,
      loadCredentialState,
      logoutCurrentSession,
    },
  ) {}
  getSnapshot = (): Readonly<SecurityState> => this.state;
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  private publish(patch: Partial<SecurityState>) {
    if (patch.enrollment === null) this.enrollmentId = "";
    if (this.disposed) return;
    this.state = { ...this.state, ...patch };
    for (const listener of this.listeners) listener();
  }
  open = () => {
    if (this.disposed) return;
    this.active = true;
    const form = ++this.form;
    queueMicrotask(() => {
      if (this.active && this.form === form) void this.refreshCredentialState();
    });
  };
  close = () => {
    this.active = false;
    ++this.form;
    ++this.read;
    this.cancelRead?.();
    clearTimeout(this.expiry);
    this.publish({
      ...emptySecrets,
      enrollment: null,
      fieldErrors: {},
      error: null,
    });
  };
  retire = () => {
    this.close();
    ++this.epoch;
    for (const cancel of this.observations) cancel();
    this.observations.clear();
    this.operation = null;
    this.publish({
      ...initial(),
      transportPending: this.transports.size !== 0,
    });
  };
  dispose = () => {
    if (this.disposed) return;
    this.retire();
    this.disposed = true;
    this.listeners.clear();
  };
  change = (field: SecretField, value: string) => {
    if (this.active)
      this.publish({
        [field]: value,
        fieldErrors: { ...this.state.fieldErrors, [field]: undefined },
      });
  };
  private validFields(errors: Partial<Record<SecretField, string>>) {
    this.publish({ fieldErrors: errors, error: null });
    return Object.keys(errors).length === 0;
  }
  refreshCredentialState = async () => {
    if (!this.active) return false;
    this.cancelRead?.();
    const read = ++this.read;
    const form = this.form;
    const epoch = this.epoch;
    const effects = this.capture();
    const current = () =>
      this.active &&
      read === this.read &&
      form === this.form &&
      epoch === this.epoch &&
      effects.current();
    this.publish({ credentialRead: "loading", credentialStateError: null });
    const observation = observeAccountOperation((signal) =>
      this.api.loadCredentialState({ signal }),
    );
    this.cancelRead = observation.cancel;
    const outcome = await observation.result;
    if (!current()) return false;
    if (
      outcome.kind === "completed" &&
      outcome.value.ok &&
      outcome.value.payload.data.user_id === effects.actor
    ) {
      this.publish({
        credentialState: outcome.value.payload.data,
        credentialRead: "ready",
        credentialStateError: null,
        ...(this.state.operation.kind === "idle"
          ? { statusText: "Account security is current." }
          : {}),
      });
      return true;
    }
    const error =
      outcome.kind === "completed" && !outcome.value.ok
        ? outcome.value.payload.error
        : null;
    this.publish({
      credentialRead: "failed",
      credentialStateError: accountOperationError(error) ?? {
        code: "credential_state_unavailable",
        status: 503,
      },
      ...(this.state.operation.kind === "idle"
        ? {
            statusText:
              "Account security could not be loaded. Refresh to check current state.",
          }
        : {}),
    });
    if (error?.code === "session_required" && current())
      await this.propagate(
        effects,
        {
          kind: "session_lost",
          message: "Your session is no longer available. Sign in again.",
        },
        current,
      );
    return false;
  };
  private async propagate(
    effects: SecurityEffects,
    event: AccountSessionEvent,
    current: () => boolean,
  ) {
    if (!current()) return false;
    const observation = observeAccountOperation((signal) =>
      Promise.resolve(effects.event(event, signal, current)),
    );
    this.observations.add(observation.cancel);
    const outcome = await observation.result;
    this.observations.delete(observation.cancel);
    return current() && outcome.kind === "completed";
  }
  refresh = async () => {
    if (!this.active) return;
    const form = this.form;
    const epoch = this.epoch;
    const effects = this.capture();
    const current = () =>
      this.active &&
      this.form === form &&
      this.epoch === epoch &&
      effects.current();
    const operation = this.state.operation;
    const loaded = await this.refreshCredentialState();
    if (!current()) return;
    const propagated =
      loaded &&
      (await this.propagate(effects, { kind: "resource_refresh" }, current));
    if (!current()) return;
    if (operation.kind === "confirmed" && this.state.operation === operation)
      this.publish({
        operation: {
          ...operation,
          propagation: propagated ? "ready" : "failed",
        },
        statusText: propagated
          ? "Credential action confirmed. Refreshed account security."
          : "Credential action confirmed. Account refresh failed. Retry refresh.",
      });
    else if (operation.kind === "uncertain")
      this.publish({
        statusText:
          "Credential action remains uncertain. Current state is an observation, not confirmation of that request. Review a new action or sign in again.",
      });
    else
      this.publish({
        statusText: propagated
          ? "Refreshed account security."
          : "Account security refresh failed. Retry refresh.",
      });
  };
  review = () => {
    if (
      this.active &&
      !this.transport &&
      !this.operation &&
      this.state.credentialRead === "ready"
    )
      this.publish({
        operation: { kind: "idle" },
        error: null,
        statusText:
          "Current security state reviewed. Enter fresh credentials for a new action.",
      });
  };
  private async execute<T>(
    action: Action,
    request: (signal: AbortSignal) => Promise<HTTPOperationResult<T>>,
    confirmed: (
      data: T,
      formCurrent: () => boolean,
    ) => AccountSessionEvent | null,
    matchesActor: (data: T, actor: string) => boolean = () => true,
  ) {
    if (
      !this.active ||
      this.logoutTransport !== null ||
      (action !== "logout" &&
        (this.operation ||
          this.transport ||
          this.state.operation.kind === "uncertain"))
    )
      return;
    const effects = this.capture();
    if (!effects.current()) return;
    const releaseTransport =
      action === "logout" ? effects.admitLogout() : undefined;
    if (action === "logout" && !releaseTransport) {
      this.publish({
        ...emptySecrets,
        statusText:
          "An earlier authentication request is still settling. Reload if it does not finish.",
      });
      return;
    }
    const identity = {};
    this.transports.add(identity);
    if (action === "logout") {
      ++this.form;
      this.logoutTransport = identity;
      this.publish({ ...emptySecrets, enrollment: null });
      clearTimeout(this.expiry);
    }
    const form = this.form;
    const epoch = this.epoch;
    this.operation = identity;
    if (action !== "logout") this.transport = identity;
    const current = () =>
      epoch === this.epoch && this.operation === identity && effects.current();
    const formCurrent = () => current() && this.active && form === this.form;
    this.publish({
      operation: { kind: "pending", action },
      error: null,
      transportPending: true,
      statusText:
        action === "password"
          ? "Changing password"
          : action === "begin"
            ? "Beginning TOTP enrollment"
            : action === "complete"
              ? "Completing TOTP enrollment"
              : "Signing out",
    });
    const observation = observeAccountOperation(request);
    this.observations.add(observation.cancel);
    // Submitted credentials are now owned only by the transport, never a replay record.
    this.publish({
      passwordCurrent: "",
      passwordNext: "",
      passwordFactorCode: "",
      totpCurrentPassword: "",
      totpCurrentFactorCode: "",
      totpCompleteCode: "",
    });
    void observation.settled.then(() => {
      releaseTransport?.();
      if (this.transport === identity) this.transport = null;
      if (this.logoutTransport === identity) this.logoutTransport = null;
      this.transports.delete(identity);
      this.publish({ transportPending: this.transports.size !== 0 });
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
        this.publish({
          ...emptySecrets,
          enrollment: null,
          operation: { kind: "uncertain", action },
          statusText:
            "Credential action outcome is uncertain. Refresh current security state, then review a new action or sign in again. No request will be repeated automatically.",
        });
        if (formCurrent()) void this.refreshCredentialState();
        return;
      }
      const result = outcome.value;
      if (!result.ok) {
        this.publish({
          operation: { kind: "rejected", action },
          error: accountOperationError(result.payload.error),
          statusText:
            action === "password"
              ? "Password change failed"
              : action === "begin"
                ? "TOTP begin failed"
                : action === "complete"
                  ? "TOTP complete failed"
                  : "Sign out failed",
        });
        if (result.payload.error?.code === "session_required" && formCurrent())
          await this.propagate(
            effects,
            {
              kind: "session_lost",
              message: "Your session is no longer available. Sign in again.",
            },
            formCurrent,
          );
        return;
      }
      if (!matchesActor(result.payload, effects.actor)) {
        this.publish({
          ...emptySecrets,
          enrollment: null,
          operation: { kind: "uncertain", action },
          statusText:
            "Credential response did not match this account. Refresh current state before a new action.",
        });
        return;
      }
      const event = confirmed(result.payload, formCurrent);
      this.publish({
        operation: {
          kind: "confirmed",
          action,
          propagation:
            event === null ? "ready" : formCurrent() ? "pending" : "failed",
        },
        statusText:
          action === "begin"
            ? this.enrollmentId
              ? "Began TOTP enrollment"
              : "Enrollment request confirmed; setup is unavailable or expired. Review a new enrollment."
            : action === "complete"
              ? "TOTP enrollment completed."
              : action === "password"
                ? "Password changed."
                : "Signed out.",
      });
      if (event !== null && formCurrent()) {
        const propagated = await this.propagate(effects, event, formCurrent);
        if (!current()) return;
        this.publish({
          operation: {
            kind: "confirmed",
            action,
            propagation: propagated ? "ready" : "failed",
          },
          ...(propagated
            ? {}
            : {
                statusText:
                  "Credential action confirmed. Account refresh failed. Retry refresh.",
              }),
        });
      }
    } finally {
      this.observations.delete(observation.cancel);
      if (this.operation === identity) this.operation = null;
    }
  }
  logout = () => {
    void this.execute(
      "logout",
      (signal) => this.api.logoutCurrentSession({ signal }),
      () => ({
        kind: "logout_confirmed",
        message: this.transport
          ? "Signed out. An earlier credential request may still finish."
          : "Signed out.",
      }),
      (payload, actor) => payload.data.user_id === actor,
    );
  };
  password = () => {
    const s = this.state;
    const errors: Partial<Record<SecretField, string>> = {};
    if (!s.passwordCurrent)
      errors.passwordCurrent = "Enter your current password.";
    const nextError = validateProvisioningPassword(s.passwordNext);
    if (nextError) errors.passwordNext = nextError;
    if (
      (s.passwordFactorCode || s.credentialState?.totp.state === "active") &&
      !/^\d{6}$/.test(s.passwordFactorCode)
    )
      errors.passwordFactorCode = "Enter a six-digit authenticator code.";
    if (!this.validFields(errors)) return;
    const input = {
      currentPassword: s.passwordCurrent,
      newPassword: s.passwordNext,
      secondFactorCode: s.passwordFactorCode,
      clientTxnId: clientTxnID("security-password"),
    };
    void this.execute(
      "password",
      (signal) => this.api.changePassword({ ...input, signal }),
      () => ({
        kind: "credentials_revoked",
        message: "Password changed. Sign in again.",
      }),
      (payload, actor) => payload.data.user_id === actor,
    );
  };
  begin = () => {
    const s = this.state;
    const errors: Partial<Record<SecretField, string>> = {};
    if (!s.totpCurrentPassword)
      errors.totpCurrentPassword = "Enter your current password.";
    if (
      (s.totpCurrentFactorCode || s.credentialState?.totp.state === "active") &&
      !/^\d{6}$/.test(s.totpCurrentFactorCode)
    )
      errors.totpCurrentFactorCode = "Enter a six-digit authenticator code.";
    if (!this.validFields(errors)) return;
    const input = {
      authMode: "session" as const,
      currentPassword: s.totpCurrentPassword,
      currentFactorCode: s.totpCurrentFactorCode,
      clientTxnId: clientTxnID("security-totp-begin"),
    };
    void this.execute(
      "begin",
      (signal) => this.api.beginTotpEnrollment({ ...input, signal }),
      (payload, current) => {
        const data = payload.data;
        const expires = Date.parse(data.expires_at);
        if (current() && Number.isFinite(expires) && expires > Date.now()) {
          this.enrollmentId = data.enrollment_id;
          this.publish({
            enrollment: {
              key: data.totp_setup.secret_base32,
              expiresAt: expires,
            },
          });
          clearTimeout(this.expiry);
          this.expiry = setTimeout(
            () =>
              this.publish({
                enrollment: null,
                totpCompleteCode: "",
                statusText: "TOTP enrollment expired. Review a new enrollment.",
              }),
            Math.min(expires - Date.now(), 2_147_483_647),
          );
        }
        return null;
      },
    );
  };
  complete = () => {
    if (!this.enrollmentId) return;
    if (
      !this.validFields(
        /^\d{6}$/.test(this.state.totpCompleteCode)
          ? {}
          : { totpCompleteCode: "Enter a six-digit authenticator code." },
      )
    )
      return;
    const input = {
      authMode: "session" as const,
      code: this.state.totpCompleteCode,
      enrollmentId: this.enrollmentId,
      clientTxnId: clientTxnID("security-totp-complete"),
    };
    void this.execute(
      "complete",
      (signal) => this.api.completeTotpEnrollment({ ...input, signal }),
      (payload) => {
        clearTimeout(this.expiry);
        this.publish({
          enrollment: null,
          totpCompleteCode: "",
        });
        return payload.data.sessions_revoked
          ? {
              kind: "credentials_revoked",
              message: "TOTP enrollment completed. Sign in again.",
            }
          : { kind: "resource_refresh" };
      },
      (payload, actor) => payload.data.user_id === actor,
    );
  };
}
