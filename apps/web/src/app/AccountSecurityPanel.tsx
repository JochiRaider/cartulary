import {
  accountTestId,
  publicErrorCodeTestId,
  publicErrorSummaryTestIds,
} from "@cartulary/ui-contracts";
import { publicErrorView } from "../services/browserApi";
import type { AccountSecurityController } from "./accountSecurityModel";
import { PublicErrorSummary } from "./LandingAdminDisplay";
import { useAccountSecurity } from "./useAccountSecurity";

type AccountSecurityPanelProps = { controller: AccountSecurityController };

import type { CSSProperties } from "react";
import {
  buttonRowStyle,
  buttonStyle,
  cardHeaderStyle,
  cardStyle,
  detailGridStyle,
  errorStyle,
  formGridStyle,
  inputStyle,
  labelBlockStyle,
  labelStyle,
  monoTextStyle,
  secondaryButtonStyle,
  sectionTitleStyle,
  statusTextStyle,
  subsectionTitleStyle,
} from "./accountPanelStyles";
import { sectionEyebrowStyle } from "./landingAdminStyles";
export function AccountSecurityPanel(props: AccountSecurityPanelProps) {
  const { snapshot, commands } = useAccountSecurity(props.controller);
  const {
    credentialStateError,
    credentialState,
    credentialRead,
    fieldErrors,
    operation,
    transportPending,
    statusText,
    error,
    passwordCurrent,
    passwordNext,
    passwordFactorCode,
    totpCurrentPassword,
    totpCurrentFactorCode,
    enrollment,
    totpCompleteCode,
  } = snapshot;
  return (
    <section style={cardStyle}>
      <div style={cardHeaderStyle}>
        <div>
          <p style={sectionEyebrowStyle}>Account</p>
          <h2 style={sectionTitleStyle}>Session and credential security</h2>
        </div>
        <div style={buttonRowStyle}>
          <button
            data-testid={accountTestId("refresh-state")}
            style={secondaryButtonStyle}
            type="button"
            onClick={() => {
              void commands.refresh();
            }}
          >
            Refresh
          </button>
          <button
            data-testid={accountTestId("logout")}
            style={secondaryButtonStyle}
            type="button"
            onClick={() => {
              void commands.logout();
            }}
          >
            Sign out
          </button>
        </div>
      </div>

      <p>
        {credentialRead === "loading"
          ? "Loading credential state…"
          : credentialRead === "failed"
            ? "Credential state unavailable; refresh to check it."
            : `Authenticator: ${credentialState?.totp.state ?? "unavailable"}.`}
      </p>
      <form
        noValidate
        onSubmit={(event) => {
          event.preventDefault();
          commands.password();
        }}
        style={subsectionStyle}
      >
        <p style={subsectionTitleStyle}>Password change</p>
        <div style={formGridStyle}>
          <label htmlFor="account-password-current" style={labelBlockStyle}>
            Current password
          </label>
          <input
            data-testid={accountTestId("password-current")}
            id="account-password-current"
            autoComplete="current-password"
            aria-invalid={fieldErrors.passwordCurrent ? true : undefined}
            aria-describedby={
              fieldErrors.passwordCurrent
                ? "account-password-current-error"
                : undefined
            }
            style={inputStyle}
            type="password"
            value={passwordCurrent}
            onChange={(event) => {
              commands.change("passwordCurrent", event.target.value);
            }}
          />
          {fieldErrors.passwordCurrent ? (
            <p id="account-password-current-error" style={errorStyle}>
              {fieldErrors.passwordCurrent}
            </p>
          ) : null}
          <label htmlFor="account-password-next" style={labelBlockStyle}>
            New password
          </label>
          <input
            data-testid={accountTestId("password-next")}
            id="account-password-next"
            autoComplete="new-password"
            aria-invalid={fieldErrors.passwordNext ? true : undefined}
            aria-describedby={
              fieldErrors.passwordNext
                ? "account-password-next-error"
                : undefined
            }
            style={inputStyle}
            type="password"
            value={passwordNext}
            onChange={(event) => {
              commands.change("passwordNext", event.target.value);
            }}
          />
          {fieldErrors.passwordNext ? (
            <p id="account-password-next-error" style={errorStyle}>
              {fieldErrors.passwordNext}
            </p>
          ) : null}
          <label htmlFor="account-password-factor" style={labelBlockStyle}>
            Current TOTP code
          </label>
          <input
            data-testid={accountTestId("password-factor-code")}
            id="account-password-factor"
            autoComplete="one-time-code"
            aria-invalid={fieldErrors.passwordFactorCode ? true : undefined}
            aria-describedby={
              fieldErrors.passwordFactorCode
                ? "account-password-factor-error"
                : undefined
            }
            style={inputStyle}
            value={passwordFactorCode}
            onChange={(event) => {
              commands.change("passwordFactorCode", event.target.value);
            }}
          />
          {fieldErrors.passwordFactorCode ? (
            <p id="account-password-factor-error" style={errorStyle}>
              {fieldErrors.passwordFactorCode}
            </p>
          ) : null}
        </div>
        <div style={buttonRowStyle}>
          <button
            data-testid={accountTestId("password-change")}
            style={buttonStyle}
            type="submit"
            disabled={
              transportPending ||
              operation.kind === "pending" ||
              operation.kind === "uncertain"
            }
          >
            Change password
          </button>
        </div>
      </form>

      <section style={subsectionStyle}>
        <p style={subsectionTitleStyle}>TOTP replacement</p>
        <form
          noValidate
          onSubmit={(event) => {
            event.preventDefault();
            commands.begin();
          }}
        >
          <div style={formGridStyle}>
            <label
              htmlFor="account-totp-current-password"
              style={labelBlockStyle}
            >
              Current password
            </label>
            <input
              data-testid={accountTestId("totp-current-password")}
              id="account-totp-current-password"
              autoComplete="current-password"
              aria-invalid={fieldErrors.totpCurrentPassword ? true : undefined}
              aria-describedby={
                fieldErrors.totpCurrentPassword
                  ? "account-totp-current-password-error"
                  : undefined
              }
              style={inputStyle}
              type="password"
              value={totpCurrentPassword}
              onChange={(event) => {
                commands.change("totpCurrentPassword", event.target.value);
              }}
            />
            {fieldErrors.totpCurrentPassword ? (
              <p id="account-totp-current-password-error" style={errorStyle}>
                {fieldErrors.totpCurrentPassword}
              </p>
            ) : null}
            <label
              htmlFor="account-totp-current-factor"
              style={labelBlockStyle}
            >
              Current TOTP code
            </label>
            <input
              data-testid={accountTestId("totp-current-factor")}
              id="account-totp-current-factor"
              autoComplete="one-time-code"
              aria-invalid={
                fieldErrors.totpCurrentFactorCode ? true : undefined
              }
              aria-describedby={
                fieldErrors.totpCurrentFactorCode
                  ? "account-totp-current-factor-error"
                  : undefined
              }
              style={inputStyle}
              value={totpCurrentFactorCode}
              onChange={(event) => {
                commands.change("totpCurrentFactorCode", event.target.value);
              }}
            />
            {fieldErrors.totpCurrentFactorCode ? (
              <p id="account-totp-current-factor-error" style={errorStyle}>
                {fieldErrors.totpCurrentFactorCode}
              </p>
            ) : null}
          </div>
          <div style={buttonRowStyle}>
            <button
              data-testid={accountTestId("totp-begin")}
              style={buttonStyle}
              type="submit"
              disabled={
                transportPending ||
                operation.kind === "pending" ||
                operation.kind === "uncertain"
              }
            >
              Begin TOTP enrollment
            </button>
          </div>
        </form>
        {enrollment ? (
          <div style={detailGridStyle}>
            <div style={wideCellStyle}>
              <p>
                In your authenticator app, add an account using a setup key.
                Enter this key manually, then enter the six-digit code below.
              </p>
              <span style={labelStyle}>Setup key</span>
              <div
                data-testid={accountTestId("totp-setup-key")}
                style={monoTextStyle}
              >
                {enrollment.key}
              </div>
            </div>
          </div>
        ) : null}
        <form
          noValidate
          onSubmit={(event) => {
            event.preventDefault();
            commands.complete();
          }}
        >
          <div style={formGridStyle}>
            <label htmlFor="account-totp-complete-code" style={labelBlockStyle}>
              Replacement TOTP code
            </label>
            <input
              data-testid={accountTestId("totp-complete-code")}
              id="account-totp-complete-code"
              autoComplete="one-time-code"
              aria-invalid={fieldErrors.totpCompleteCode ? true : undefined}
              aria-describedby={
                fieldErrors.totpCompleteCode
                  ? "account-totp-complete-code-error"
                  : undefined
              }
              style={inputStyle}
              value={totpCompleteCode}
              onChange={(event) => {
                commands.change("totpCompleteCode", event.target.value);
              }}
            />
            {fieldErrors.totpCompleteCode ? (
              <p id="account-totp-complete-code-error" style={errorStyle}>
                {fieldErrors.totpCompleteCode}
              </p>
            ) : null}
          </div>
          <div style={buttonRowStyle}>
            <button
              data-testid={accountTestId("totp-complete")}
              style={buttonStyle}
              type="submit"
              disabled={
                enrollment === null ||
                transportPending ||
                operation.kind === "pending" ||
                operation.kind === "uncertain"
              }
            >
              Complete TOTP enrollment
            </button>
          </div>
        </form>
      </section>

      <p
        data-testid={accountTestId("status")}
        role={
          error === null && credentialStateError === null ? "status" : undefined
        }
        style={statusTextStyle}
      >
        {Object.values(fieldErrors).some(Boolean)
          ? "Review the highlighted fields."
          : statusText}
      </p>
      <p data-testid={publicErrorCodeTestId("account")} style={errorStyle}>
        {publicErrorView(error ?? credentialStateError)?.code ?? ""}
      </p>
      {operation.kind === "uncertain" ? (
        <button
          type="button"
          style={secondaryButtonStyle}
          disabled={transportPending || credentialRead !== "ready"}
          onClick={commands.review}
        >
          Review a new credential action
        </button>
      ) : null}
      {operation.kind === "confirmed" && operation.propagation === "failed" ? (
        <button
          type="button"
          style={secondaryButtonStyle}
          onClick={() => {
            void commands.refresh();
          }}
        >
          Retry refresh
        </button>
      ) : null}
      <PublicErrorSummary
        error={error ?? credentialStateError}
        testIds={publicErrorSummaryTestIds("account")}
      />
    </section>
  );
}

const subsectionStyle: CSSProperties = {
  marginTop: "1.5rem",
  paddingTop: "1.25rem",
  borderTop: "var(--ct-border-hairline)",
  minWidth: 0,
};

const wideCellStyle: CSSProperties = {
  gridColumn: "1 / -1",
};
