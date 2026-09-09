import { authTestId, cartularyDefaultThemeId } from "@cartulary/ui-contracts";
import { Eye, EyeOff } from "lucide-react";
import { normalizeTotpCode, useAuthentication } from "./useAuthentication";

type AuthGatewayProps = Parameters<typeof useAuthentication>[0];

export function AuthGateway(props: AuthGatewayProps) {
  const { bootstrapState, readingProfile = "default" } = props;
  const {
    snapshot: state,
    commands,
    usernameRef,
    passwordRef,
    totpCodeRef,
    bootstrapBeginRef,
    bootstrapCompleteCodeRef,
    handleLoginSubmit,
    displayedBootstrapState,
    displayedBanner,
    authAlertText,
    authLiveRole,
    activeErrorCode,
    statusText,
    rootClassName,
    supportText,
    title,
    submitLabel,
    canSubmit,
  } = useAuthentication(props);
  const observing = state.sessionRead === "pending";
  const progressing =
    observing ||
    bootstrapState === "loading" ||
    state.operation.kind === "pending";
  const feedbackText = observing
    ? "Checking current session..."
    : progressing
      ? statusText
      : authAlertText || statusText;
  const feedbackRole = progressing ? "status" : (authLiveRole ?? "status");
  return (
    <main
      aria-busy={bootstrapState === "loading"}
      className={rootClassName}
      data-bootstrap-state={displayedBootstrapState}
      data-reading-profile={
        readingProfile === "hyperlegible" ? "hyperlegible" : undefined
      }
      data-testid={authTestId("shell")}
      data-cartulary-theme={cartularyDefaultThemeId}
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
            data-testid={authTestId("shell-message")}
          >
            {supportText}
          </p>

          <p
            className={
              feedbackText
                ? "cartulary-auth-banner"
                : "cartulary-auth-visually-hidden"
            }
            data-tone={
              feedbackRole === "alert"
                ? "error"
                : (displayedBanner?.tone ?? "info")
            }
            data-testid={authTestId("feedback")}
            data-error-code={activeErrorCode}
            role={feedbackRole}
          >
            {feedbackText}
          </p>

          {state.operation.kind === "uncertain" ||
          state.transportPending ||
          (state.operation.kind === "confirmed" &&
            state.operation.propagation !== "ready") ? (
            <button
              type="button"
              className="cartulary-auth-secondary-button"
              disabled={
                state.sessionRead === "pending" || !commands.canInspectSession()
              }
              onClick={() => {
                void props.controller.retrySession();
              }}
            >
              Retry session check
            </button>
          ) : null}
          {state.flow.kind === "setup" ? (
            <form
              noValidate
              onSubmit={(event) => {
                event.preventDefault();
                commands.complete();
              }}
              className="cartulary-auth-setup"
              aria-label="Authenticator setup"
            >
              <p className="cartulary-auth-setup-copy">
                This account requires authenticator enrollment before it can
                sign in.
              </p>
              {state.flow.enrollment.kind === "ready" ? (
                <div className="cartulary-auth-detail-list">
                  <p>
                    In your authenticator app, add an account using a setup key.
                    Enter this key manually, then enter the six-digit code from
                    the app below.
                  </p>
                  <div>
                    <span className="cartulary-auth-detail-label">
                      Setup key
                    </span>
                    <div
                      className="cartulary-auth-mono"
                      data-testid={authTestId("bootstrap-setup-key")}
                    >
                      {state.flow.enrollment.key}
                    </div>
                  </div>
                </div>
              ) : (
                <button
                  ref={bootstrapBeginRef}
                  aria-disabled={!canSubmit}
                  className="cartulary-auth-primary-button"
                  data-testid={authTestId("bootstrap-begin")}
                  disabled={!canSubmit}
                  type="button"
                  onClick={() => {
                    void commands.begin();
                  }}
                >
                  {state.operation.kind === "pending" &&
                  state.operation.action === "begin"
                    ? "Beginning setup..."
                    : "Begin enrollment"}
                </button>
              )}
              {state.flow.enrollment.kind === "ready" ? (
                <>
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
                        state.fieldErrors.bootstrapCompleteCode
                          ? true
                          : undefined
                      }
                      autoComplete="one-time-code"
                      data-testid={authTestId("bootstrap-complete-code")}
                      id="auth-bootstrap-complete-code"
                      inputMode="numeric"
                      maxLength={6}
                      pattern="[0-9]*"
                      type="text"
                      value={state.flow.kind === "setup" ? state.flow.code : ""}
                      onChange={(event) => {
                        commands.dispatch({
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
                    >
                      {state.fieldErrors.bootstrapCompleteCode}
                    </p>
                  ) : null}
                  <button
                    aria-disabled={
                      state.operation.kind === "pending" ||
                      state.transportPending
                    }
                    className="cartulary-auth-primary-button"
                    data-testid={authTestId("bootstrap-complete")}
                    disabled={
                      state.operation.kind === "pending" ||
                      state.transportPending
                    }
                    type="submit"
                  >
                    {state.operation.kind === "pending" &&
                    state.operation.action === "complete"
                      ? "Completing setup..."
                      : "Complete enrollment"}
                  </button>
                </>
              ) : null}
              <button
                className="cartulary-auth-secondary-button"
                type="button"
                onClick={() =>
                  commands.dispatch({ type: "use_different_account" })
                }
              >
                Use a different account
              </button>
            </form>
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
                  data-testid={authTestId("login-username")}
                  id="auth-login-username"
                  type="email"
                  value={state.flow.username}
                  onChange={(event) => {
                    commands.dispatch({
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
                >
                  {state.fieldErrors.username}
                </p>
              ) : null}

              <label
                className="cartulary-auth-field"
                htmlFor="auth-login-password"
              >
                <span id="auth-login-password-label">Password</span>
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
                    data-testid={authTestId("login-password")}
                    id="auth-login-password"
                    aria-labelledby="auth-login-password-label"
                    type={state.flow.passwordVisible ? "text" : "password"}
                    value={state.flow.password}
                    onChange={(event) => {
                      commands.dispatch({
                        type: "field",
                        field: "password",
                        value: event.target.value,
                      });
                    }}
                  />
                  <button
                    aria-label={
                      state.flow.passwordVisible
                        ? "Hide password"
                        : "Show password"
                    }
                    className="cartulary-auth-icon-button"
                    type="button"
                    onClick={() =>
                      commands.dispatch({ type: "toggle_password_visibility" })
                    }
                  >
                    {state.flow.passwordVisible ? (
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
                >
                  {state.fieldErrors.password}
                </p>
              ) : null}

              {state.flow.kind === "mfa" ? (
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
                      data-testid={authTestId("login-totp-code")}
                      id="auth-login-totp-code"
                      inputMode="numeric"
                      maxLength={6}
                      pattern="[0-9]*"
                      type="text"
                      value={state.flow.kind === "mfa" ? state.flow.code : ""}
                      onChange={(event) => {
                        commands.dispatch({
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
                    >
                      {state.fieldErrors.totpCode}
                    </p>
                  ) : null}
                </>
              ) : null}

              <button
                aria-busy={
                  state.operation.kind === "pending" &&
                  state.operation.action === "login"
                    ? true
                    : undefined
                }
                aria-disabled={!canSubmit}
                className="cartulary-auth-primary-button"
                data-testid={authTestId("login-submit")}
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

          {state.providersStatus === "failed" ? (
            <p role="status">
              Enterprise sign-in options could not be loaded. Local sign-in
              remains available.{" "}
              <button
                type="button"
                className="cartulary-auth-secondary-button"
                onClick={() => {
                  void props.controller.discover();
                }}
              >
                Retry enterprise options
              </button>
            </p>
          ) : null}
          {state.enterpriseProviders.length > 0 ? (
            <section
              className="cartulary-auth-enterprise"
              aria-label="Enterprise sign-in"
            >
              <p className="cartulary-auth-enterprise-label">
                Enterprise sign-in
              </p>
              <div data-testid={authTestId("enterprise-provider-list")}>
                {state.enterpriseProviders.map((provider) => (
                  <button
                    key={provider.provider_key}
                    className="cartulary-auth-secondary-button"
                    data-provider-key={provider.provider_key}
                    data-testid={authTestId("enterprise-provider-button")}
                    disabled={!canSubmit}
                    type="button"
                    onClick={() => {
                      void commands.enterprise(provider.provider_key);
                    }}
                  >
                    {provider.display_name}
                  </button>
                ))}
              </div>
            </section>
          ) : null}
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
