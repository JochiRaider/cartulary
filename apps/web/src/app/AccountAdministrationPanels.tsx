import {
  accountTestId,
  deploymentAdminTestId,
  deploymentUserRowTestId,
  publicErrorCodeTestId,
  publicErrorSummaryTestIds,
} from "@cartulary/ui-contracts";
import { X } from "lucide-react";
import type { CSSProperties } from "react";
import { type APIError, publicErrorView } from "../services/browserApi";
import {
  type AccountSecurityPanelProps,
  useAccountSecurity,
} from "./accountSecurityModel";
import { DeploymentUserActionDialog } from "./DeploymentUserActionDialog";
import { DeploymentUserLeaveDialog } from "./DeploymentUserLeaveDialog";
import {
  type DeploymentUsersPanelProps,
  isEnterpriseAuthBinding,
  useDeploymentUsers,
} from "./deploymentUsersModel";

export type { AccountSessionEvent } from "./accountSecurityModel";

export function AccountSecurityPanel(props: AccountSecurityPanelProps) {
  const {
    credentialStateError,
    credentialState,
    credentialRead,
    fieldErrors,
    operation,
    transportPending,
    review,
    statusText,
    error,
    passwordCurrent,
    setPasswordCurrent,
    passwordNext,
    setPasswordNext,
    passwordFactorCode,
    setPasswordFactorCode,
    totpCurrentPassword,
    setTotpCurrentPassword,
    totpCurrentFactorCode,
    setTotpCurrentFactorCode,
    totpEnrollmentId,
    totpSecretBase32,
    totpCompleteCode,
    setTotpCompleteCode,
    handleLogout,
    handleRefreshAccount,
    handlePasswordChange,
    handleBeginTotpReplacement,
    handleCompleteTotpReplacement,
  } = useAccountSecurity(props);
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
              void handleRefreshAccount();
            }}
          >
            Refresh
          </button>
          <button
            data-testid={accountTestId("logout")}
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

      <p role="status">
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
          handlePasswordChange();
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
              setPasswordCurrent(event.target.value);
            }}
          />
          {fieldErrors.passwordCurrent ? (
            <p
              id="account-password-current-error"
              role="alert"
              style={errorStyle}
            >
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
              setPasswordNext(event.target.value);
            }}
          />
          {fieldErrors.passwordNext ? (
            <p id="account-password-next-error" role="alert" style={errorStyle}>
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
              setPasswordFactorCode(event.target.value);
            }}
          />
          {fieldErrors.passwordFactorCode ? (
            <p
              id="account-password-factor-error"
              role="alert"
              style={errorStyle}
            >
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
            handleBeginTotpReplacement();
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
                setTotpCurrentPassword(event.target.value);
              }}
            />
            {fieldErrors.totpCurrentPassword ? (
              <p
                id="account-totp-current-password-error"
                role="alert"
                style={errorStyle}
              >
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
                setTotpCurrentFactorCode(event.target.value);
              }}
            />
            {fieldErrors.totpCurrentFactorCode ? (
              <p
                id="account-totp-current-factor-error"
                role="alert"
                style={errorStyle}
              >
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
        <div style={detailGridStyle}>
          <div>
            <span style={labelStyle}>Enrollment id</span>
            <div
              data-testid={accountTestId("totp-enrollment-id")}
              style={monoTextStyle}
            >
              {totpEnrollmentId}
            </div>
          </div>
          <div style={wideCellStyle}>
            <span style={labelStyle}>Secret base32</span>
            <div
              data-testid={accountTestId("totp-secret-base32")}
              style={monoTextStyle}
            >
              {totpSecretBase32}
            </div>
          </div>
        </div>
        <form
          noValidate
          onSubmit={(event) => {
            event.preventDefault();
            handleCompleteTotpReplacement();
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
                setTotpCompleteCode(event.target.value);
              }}
            />
            {fieldErrors.totpCompleteCode ? (
              <p
                id="account-totp-complete-code-error"
                role="alert"
                style={errorStyle}
              >
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
                !totpEnrollmentId ||
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
        aria-live="polite"
        data-testid={accountTestId("status")}
        role="status"
        style={statusTextStyle}
      >
        {statusText}
      </p>
      <p
        aria-live="assertive"
        data-testid={publicErrorCodeTestId("account")}
        role={
          error === null && credentialStateError === null ? undefined : "alert"
        }
        style={errorStyle}
      >
        {publicErrorView(error ?? credentialStateError)?.code ?? ""}
      </p>
      {operation.kind === "uncertain" ? (
        <button
          type="button"
          style={secondaryButtonStyle}
          disabled={transportPending || credentialRead !== "ready"}
          onClick={review}
        >
          Review a new credential action
        </button>
      ) : null}
      {operation.kind === "confirmed" && operation.propagation === "failed" ? (
        <button
          type="button"
          style={secondaryButtonStyle}
          onClick={() => {
            void handleRefreshAccount();
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

export function DeploymentUsersPanel(props: DeploymentUsersPanelProps) {
  const { controller, enterpriseAuthClaimed = false } = props;
  const {
    statusText,
    error,
    selectedUser,
    users,
    userFilter,
    setUserFilter,
    userActiveFilter,
    setUserActiveFilter,
    userAdminFilter,
    setUserAdminFilter,
    createEmail,
    setCreateEmail,
    createDisplayName,
    setCreateDisplayName,
    createInitialPassword,
    setCreateInitialPassword,
    createMfaRequired,
    setCreateMfaRequired,
    createIsDeploymentAdmin,
    setCreateIsDeploymentAdmin,
    createDialogOpen,
    setCreateDialogOpen,
    patchEmail,
    setPatchEmail,
    patchDisplayName,
    setPatchDisplayName,
    patchMfaRequired,
    setPatchMfaRequired,
    patchIsActive,
    setPatchIsActive,
    patchIsDeploymentAdmin,
    setPatchIsDeploymentAdmin,
    adminNewPassword,
    setAdminNewPassword,
    adminReason,
    setAdminReason,
    credentialDialog,
    setCredentialDialog,
    enterpriseProviders,
    bindingProviderKey,
    setBindingProviderKey,
    bindingProviderSubject,
    setBindingProviderSubject,
    bindingTargetID,
    setBindingTargetID,
    bindingNewSubject,
    setBindingNewSubject,
    bindingReason,
    setBindingReason,
    clearSelectedUser,
    loadSelectedUser,
    handleCreateUser,
    handlePatchUser,
    handleAdminPasswordReset,
    handleAdminTotpReset,
    handleAdminRevokeAll,
    handleCreateEnterpriseBinding,
    handleRotateEnterpriseBinding,
    handleRetireEnterpriseBinding,
    refreshUsers,
    loadMoreUsers,
    targetOperationPending,
    canSubmitTargetAction,
    authorized,
    draft,
    operation,
    operationObserved,
    transportPending,
    fieldErrors,
    queryStatus,
    targetStatus,
    acceptedQuery,
    providersStatus,
    canSave,
    canPage,
    dirty,
    reviewRequired,
  } = useDeploymentUsers(props);
  if (!authorized) {
    return (
      <section style={cardStyle}>
        <div style={cardHeaderStyle}>
          <div>
            <p style={sectionEyebrowStyle}>Deployment users</p>
            <h2 style={sectionTitleStyle}>User administration</h2>
          </div>
        </div>
        <p data-testid={deploymentAdminTestId("access-note")} style={bodyStyle}>
          Deployment admin access is required for user creation, patching, and
          credential actions. Incident-admin membership alone does not unlock
          these controls.
        </p>
        <p role="status" data-testid={deploymentAdminTestId("status")}>
          {statusText}
        </p>
        <p role="alert" data-testid={publicErrorCodeTestId("admin")}>
          {publicErrorView(error)?.code ?? ""}
        </p>
        <PublicErrorSummary
          error={error}
          testIds={publicErrorSummaryTestIds("admin")}
        />
        <button
          type="button"
          style={secondaryButtonStyle}
          onClick={() => {
            void controller.refreshUsers();
          }}
        >
          Check administration access
        </button>
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
      <DeploymentUserLeaveDialog controller={controller} />
      <div style={cardHeaderStyle}>
        <div>
          <p style={sectionEyebrowStyle}>Deployment users</p>
          <h2 style={sectionTitleStyle}>User administration</h2>
        </div>
        <button
          data-testid={
            createDialogOpen ? undefined : deploymentAdminTestId("create-user")
          }
          disabled={targetOperationPending}
          style={buttonStyle}
          type="button"
          onClick={() => {
            void setCreateDialogOpen(true);
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
          {queryStatus === "loading" || queryStatus === "pending" ? (
            <p role="status">
              Updating users. The last accepted results remain visible.
            </p>
          ) : null}
          {queryStatus === "failed" ? (
            <p role="status">
              User query failed. Showing results for{" "}
              {acceptedQuery?.search || "the last accepted query"}; refresh to
              retry.
            </p>
          ) : null}
          <label htmlFor="admin-user-filter" style={labelBlockStyle}>
            Search users
          </label>
          <input
            data-testid={deploymentAdminTestId("user-filter")}
            id="admin-user-filter"
            style={inputStyle}
            value={userFilter}
            onChange={(event) => {
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
                data-testid={deploymentAdminTestId("user-is-active-filter")}
                style={inputStyle}
                value={userActiveFilter}
                onChange={(event) => {
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
                data-testid={deploymentAdminTestId(
                  "user-is-deployment-admin-filter",
                )}
                style={inputStyle}
                value={userAdminFilter}
                onChange={(event) => {
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
            data-testid={deploymentAdminTestId("user-list")}
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
            data-testid={deploymentAdminTestId("load-more-users")}
            disabled={!canPage}
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
              <form
                noValidate
                onSubmit={(event) => {
                  event.preventDefault();
                  void handlePatchUser();
                }}
                style={inspectorSectionStyle}
              >
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
                      data-testid={deploymentAdminTestId("target-user-id")}
                      style={monoTextStyle}
                    >
                      {selectedUser.user_id}
                    </div>
                  </div>
                  <div>
                    <span style={labelStyle}>User version</span>
                    <div
                      data-testid={deploymentAdminTestId("target-user-version")}
                    >
                      {selectedUser.user_version}
                    </div>
                  </div>
                  <div>
                    <span style={labelStyle}>Is active</span>
                    <div
                      data-testid={deploymentAdminTestId("target-is-active")}
                    >
                      {String(selectedUser.is_active)}
                    </div>
                  </div>
                  <div>
                    <span style={labelStyle}>Deployment admin</span>
                    <div
                      data-testid={deploymentAdminTestId(
                        "target-is-deployment-admin",
                      )}
                    >
                      {String(selectedUser.is_deployment_admin)}
                    </div>
                  </div>
                  <div>
                    <span style={labelStyle}>Base user version</span>
                    <div
                      data-testid={deploymentAdminTestId("patch-base-version")}
                    >
                      {draft?.baseVersion}
                    </div>
                  </div>
                </div>
                <div style={formGridStyle}>
                  <div>
                    <label htmlFor="admin-patch-email" style={labelBlockStyle}>
                      Email
                      <input
                        data-testid={deploymentAdminTestId("patch-email")}
                        id="admin-patch-email"
                        aria-invalid={fieldErrors.email ? true : undefined}
                        aria-describedby={
                          fieldErrors.email
                            ? "admin-patch-email-error"
                            : undefined
                        }
                        style={inputStyle}
                        value={patchEmail}
                        onChange={(event) => {
                          setPatchEmail(event.target.value);
                        }}
                      />
                    </label>
                    {fieldErrors.email ? (
                      <p
                        id="admin-patch-email-error"
                        role="alert"
                        style={errorStyle}
                      >
                        {fieldErrors.email}
                      </p>
                    ) : null}
                  </div>
                  <div>
                    <label
                      htmlFor="admin-patch-display-name"
                      style={labelBlockStyle}
                    >
                      Display name
                      <input
                        data-testid={deploymentAdminTestId(
                          "patch-display-name",
                        )}
                        id="admin-patch-display-name"
                        aria-invalid={
                          fieldErrors.display_name ? true : undefined
                        }
                        aria-describedby={
                          fieldErrors.display_name
                            ? "admin-patch-display-name-error"
                            : undefined
                        }
                        style={inputStyle}
                        value={patchDisplayName}
                        onChange={(event) => {
                          setPatchDisplayName(event.target.value);
                        }}
                      />
                    </label>
                    {fieldErrors.display_name ? (
                      <p
                        id="admin-patch-display-name-error"
                        role="alert"
                        style={errorStyle}
                      >
                        {fieldErrors.display_name}
                      </p>
                    ) : null}
                  </div>
                </div>
                <div style={checkboxRowStyle}>
                  <label style={checkboxLabelStyle}>
                    <input
                      data-testid={deploymentAdminTestId("patch-mfa-required")}
                      type="checkbox"
                      checked={patchMfaRequired}
                      onChange={(event) => {
                        setPatchMfaRequired(event.target.checked);
                      }}
                    />
                    MFA required
                  </label>
                  <label style={checkboxLabelStyle}>
                    <input
                      data-testid={deploymentAdminTestId("patch-is-active")}
                      type="checkbox"
                      checked={patchIsActive}
                      onChange={(event) => {
                        setPatchIsActive(event.target.checked);
                      }}
                    />
                    Active
                  </label>
                  <label style={checkboxLabelStyle}>
                    <input
                      data-testid={deploymentAdminTestId(
                        "patch-is-deployment-admin",
                      )}
                      type="checkbox"
                      checked={patchIsDeploymentAdmin}
                      onChange={(event) => {
                        setPatchIsDeploymentAdmin(event.target.checked);
                      }}
                    />
                    Deployment admin
                  </label>
                </div>
                {reviewRequired ? (
                  <p role="status">
                    The accepted user changed or the action needs review.
                    Refresh this user, then review remaining edits before
                    another save.
                  </p>
                ) : null}
                <div style={buttonRowStyle}>
                  <button
                    data-testid={deploymentAdminTestId("patch-user")}
                    disabled={!canSave}
                    style={buttonStyle}
                    type="submit"
                  >
                    Save user
                  </button>
                  {dirty ? (
                    <button
                      type="button"
                      style={secondaryButtonStyle}
                      onClick={controller.discard}
                    >
                      Discard user edits
                    </button>
                  ) : null}
                  <button
                    type="button"
                    style={secondaryButtonStyle}
                    onClick={() => {
                      void controller.refreshTarget();
                    }}
                  >
                    Refresh current user
                  </button>
                  {reviewRequired ? (
                    <button
                      type="button"
                      style={secondaryButtonStyle}
                      disabled={targetOperationPending || transportPending}
                      onClick={controller.review}
                    >
                      Review remaining edits
                    </button>
                  ) : null}
                </div>
              </form>

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
                        : deploymentAdminTestId("password-reset")
                    }
                    disabled={!canSubmitTargetAction}
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
                        : deploymentAdminTestId("totp-reset")
                    }
                    disabled={!canSubmitTargetAction}
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
                        : deploymentAdminTestId("revoke-all")
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
                    <div>
                      <label
                        htmlFor="admin-enterprise-provider-key"
                        style={labelBlockStyle}
                      >
                        Provider key
                        <input
                          id="admin-enterprise-provider-key"
                          aria-invalid={
                            fieldErrors.providerKey ? true : undefined
                          }
                          aria-describedby={
                            fieldErrors.providerKey
                              ? "admin-enterprise-provider-key-error"
                              : undefined
                          }
                          list="admin-enterprise-provider-options"
                          disabled={!canSubmitTargetAction}
                          style={inputStyle}
                          value={bindingProviderKey}
                          onChange={(event) => {
                            setBindingProviderKey(event.target.value);
                          }}
                        />
                      </label>
                      {fieldErrors.providerKey ? (
                        <p
                          id="admin-enterprise-provider-key-error"
                          role="alert"
                        >
                          {fieldErrors.providerKey}
                        </p>
                      ) : null}
                    </div>
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
                      disabled={!canSubmitTargetAction}
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
                        !canSubmitTargetAction || bindingTargetID === ""
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
                        !canSubmitTargetAction || bindingTargetID === ""
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

      {targetStatus === "failed" ? (
        <p role="status">
          Current user could not be refreshed. The accepted resource and draft
          remain available; refresh this user to retry.
        </p>
      ) : null}
      {operation.kind !== "idle" ? (
        <p role="status">
          {operation.kind === "confirmed"
            ? "Confirmed. "
            : operation.kind === "uncertain"
              ? "Outcome uncertain. "
              : operation.kind === "rejected"
                ? "Action rejected. "
                : "Action pending. "}
          {operation.intent.target
            ? `Action target: ${operation.intent.label} (${operation.intent.target}).`
            : `Create action: ${operation.intent.label}.`}
        </p>
      ) : null}
      {operation.kind === "uncertain" ? (
        <div style={buttonRowStyle}>
          {operation.attempt ? (
            <button
              type="button"
              style={buttonStyle}
              disabled={transportPending}
              onClick={controller.replay}
            >
              Replay exact action
            </button>
          ) : (
            <>
              <button
                type="button"
                style={secondaryButtonStyle}
                onClick={() => {
                  if (operation.intent.target)
                    void controller.select(operation.intent.target);
                  else void controller.refreshUsers();
                }}
              >
                Inspect previous action target
              </button>
              <button
                type="button"
                style={secondaryButtonStyle}
                disabled={!operationObserved || transportPending}
                onClick={controller.review}
              >
                Review a new action
              </button>
            </>
          )}
        </div>
      ) : null}
      {operation.kind === "confirmed" && operation.propagation === "failed" ? (
        <button
          type="button"
          style={secondaryButtonStyle}
          onClick={() => {
            void controller.retryRefresh();
          }}
        >
          Retry session refresh
        </button>
      ) : null}
      {enterpriseAuthClaimed && providersStatus === "failed" ? (
        <p role="status">
          Provider suggestions could not be loaded. Configured provider keys can
          still be entered.{" "}
          <button
            type="button"
            onClick={() => {
              void controller.discoverProviders();
            }}
          >
            Retry provider suggestions
          </button>
        </p>
      ) : null}
      {createDialogOpen ? (
        <div style={dialogBackdropStyle}>
          <DeploymentUserActionDialog
            label="Create local user"
            style={dialogStyle}
            onClose={() => {
              void setCreateDialogOpen(false);
            }}
            onSubmit={() => {
              void handleCreateUser();
            }}
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
                data-dialog-close=""
              >
                <X aria-hidden="true" size={16} />
              </button>
            </header>
            <div style={formGridStyle}>
              <div>
                <label htmlFor="admin-create-email" style={labelBlockStyle}>
                  Email
                  <input
                    data-testid={deploymentAdminTestId("create-email")}
                    id="admin-create-email"
                    aria-invalid={fieldErrors.createEmail ? true : undefined}
                    aria-describedby={
                      fieldErrors.createEmail
                        ? "admin-create-email-error"
                        : undefined
                    }
                    disabled={targetOperationPending}
                    style={inputStyle}
                    value={createEmail}
                    onChange={(event) => {
                      setCreateEmail(event.target.value);
                    }}
                  />
                </label>
                {fieldErrors.createEmail ? (
                  <p
                    id="admin-create-email-error"
                    role="alert"
                    style={errorStyle}
                  >
                    {fieldErrors.createEmail}
                  </p>
                ) : null}
              </div>
              <div>
                <label
                  htmlFor="admin-create-display-name"
                  style={labelBlockStyle}
                >
                  Display name
                  <input
                    data-testid={deploymentAdminTestId("create-display-name")}
                    id="admin-create-display-name"
                    aria-invalid={
                      fieldErrors.createDisplayName ? true : undefined
                    }
                    aria-describedby={
                      fieldErrors.createDisplayName
                        ? "admin-create-display-name-error"
                        : undefined
                    }
                    disabled={targetOperationPending}
                    style={inputStyle}
                    value={createDisplayName}
                    onChange={(event) => {
                      setCreateDisplayName(event.target.value);
                    }}
                  />
                </label>
                {fieldErrors.createDisplayName ? (
                  <p
                    id="admin-create-display-name-error"
                    role="alert"
                    style={errorStyle}
                  >
                    {fieldErrors.createDisplayName}
                  </p>
                ) : null}
              </div>
              <div>
                <label htmlFor="admin-create-password" style={labelBlockStyle}>
                  Initial password
                  <input
                    data-testid={deploymentAdminTestId("create-password")}
                    id="admin-create-password"
                    aria-invalid={fieldErrors.createPassword ? true : undefined}
                    aria-describedby={
                      fieldErrors.createPassword
                        ? "admin-create-password-error"
                        : undefined
                    }
                    disabled={targetOperationPending}
                    style={inputStyle}
                    type="password"
                    value={createInitialPassword}
                    onChange={(event) => {
                      setCreateInitialPassword(event.target.value);
                    }}
                  />
                </label>
                {fieldErrors.createPassword ? (
                  <p
                    id="admin-create-password-error"
                    role="alert"
                    style={errorStyle}
                  >
                    {fieldErrors.createPassword}
                  </p>
                ) : null}
              </div>
            </div>
            <div style={checkboxRowStyle}>
              <label style={checkboxLabelStyle}>
                <input
                  data-testid={deploymentAdminTestId("create-mfa-required")}
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
                  data-testid={deploymentAdminTestId(
                    "create-is-deployment-admin",
                  )}
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
                data-dialog-close=""
              >
                Cancel
              </button>
              <button
                data-testid={deploymentAdminTestId("create-user")}
                disabled={targetOperationPending}
                style={buttonStyle}
                type="submit"
              >
                Create user
              </button>
            </div>
            <p role="status">{statusText}</p>
            {error ? <p role="alert">{publicErrorView(error)?.code}</p> : null}
          </DeploymentUserActionDialog>
        </div>
      ) : null}

      {credentialDialog !== null && selectedUser !== null ? (
        <div style={dialogBackdropStyle}>
          <DeploymentUserActionDialog
            label="Confirm credential action"
            style={dialogStyle}
            onClose={() => setCredentialDialog(null)}
            onSubmit={() => {
              if (credentialDialog === "password")
                void handleAdminPasswordReset();
              else if (credentialDialog === "totp") void handleAdminTotpReset();
              else void handleAdminRevokeAll();
            }}
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
                data-dialog-close=""
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
                <div>
                  <label htmlFor="admin-new-password" style={labelBlockStyle}>
                    New password
                    <input
                      data-testid={deploymentAdminTestId("new-password")}
                      id="admin-new-password"
                      aria-invalid={fieldErrors.password ? true : undefined}
                      aria-describedby={
                        fieldErrors.password
                          ? "admin-new-password-error"
                          : undefined
                      }
                      style={inputStyle}
                      type="password"
                      value={adminNewPassword}
                      onChange={(event) => {
                        setAdminNewPassword(event.target.value);
                      }}
                    />
                  </label>
                  {fieldErrors.password ? (
                    <p
                      id="admin-new-password-error"
                      role="alert"
                      style={errorStyle}
                    >
                      {fieldErrors.password}
                    </p>
                  ) : null}
                </div>
              ) : null}
              <label htmlFor="admin-reason" style={labelBlockStyle}>
                Reason
                <input
                  data-testid={deploymentAdminTestId("reason")}
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
                data-dialog-close=""
              >
                Cancel
              </button>
              <button
                data-testid={
                  credentialDialog === "password"
                    ? deploymentAdminTestId("password-reset")
                    : credentialDialog === "totp"
                      ? deploymentAdminTestId("totp-reset")
                      : deploymentAdminTestId("revoke-all")
                }
                disabled={!canSubmitTargetAction}
                style={destructiveButtonStyle}
                type="submit"
              >
                Confirm
              </button>
            </div>
            <p role="status">{statusText}</p>
            {error ? <p role="alert">{publicErrorView(error)?.code}</p> : null}
          </DeploymentUserActionDialog>
        </div>
      ) : null}

      <p
        aria-live="polite"
        data-testid={deploymentAdminTestId("status")}
        role="status"
        style={statusTextStyle}
      >
        {statusText}
      </p>
      <p
        aria-live="assertive"
        data-testid={publicErrorCodeTestId("admin")}
        role={error === null ? undefined : "alert"}
        style={errorStyle}
      >
        {publicErrorView(error)?.code ?? ""}
      </p>
      <PublicErrorSummary
        error={error}
        testIds={publicErrorSummaryTestIds("admin")}
      />
    </section>
  );
}

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

const bodyStyle: CSSProperties = {
  margin: "0.75rem 0 0",
  color: "var(--ct-colors-ink-muted)",
  maxWidth: "42rem",
  overflowWrap: "anywhere",
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
