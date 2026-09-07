import {
  deploymentAdminTestId,
  deploymentUserRowTestId,
  publicErrorCodeTestId,
  publicErrorSummaryTestIds,
} from "@cartulary/ui-contracts";
import { X } from "lucide-react";
import { publicErrorView } from "../services/browserApi";
import { DeploymentUserActionDialog } from "./DeploymentUserActionDialog";
import { DeploymentUserLeaveDialog } from "./DeploymentUserLeaveDialog";
import type { DeploymentUsersController } from "./deploymentUsersModel";
import { isEnterpriseAuthBinding } from "./deploymentUsersModel";
import { PublicErrorSummary } from "./LandingAdminDisplay";
import { useDeploymentUsers } from "./useDeploymentUsers";

type DeploymentUsersPanelProps = {
  controller: DeploymentUsersController;
  fallbackFocusRef?: RefObject<HTMLElement | null> | undefined;
} & Parameters<typeof useDeploymentUsers>[1];

import type { CSSProperties, RefObject } from "react";
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
export function DeploymentUsersPanel(props: DeploymentUsersPanelProps) {
  const { controller } = props;
  const { snapshot, commands } = useDeploymentUsers(controller, props);
  const {
    statusText,
    error,
    selected,
    users,
    queryDraft,
    create,
    createOpen,
    credentialDialog,
    password,
    reason,
    binding: bindingDraft,
    providers,
    authorized,
    draft,
    operation,
    enterpriseClaimed,
    transportPending,
    fieldErrors,
    queryStatus,
    targetStatus,
    acceptedQuery,
    providersStatus,
  } = snapshot;
  const values = selected ? { ...selected, ...draft?.changes } : null;
  const targetOperationPending =
    targetStatus === "loading" || operation.kind === "pending";
  const canSubmitTargetAction = commands.canTargetAction();
  const canSave = commands.canSave();
  const canPage = commands.canPage();
  const dirty = commands.hasDirtyDraft();
  const reviewRequired = commands.needsReview();
  const feedback = (
    <>
      <p
        data-testid={deploymentAdminTestId("status")}
        role={error === null ? "status" : undefined}
        style={statusTextStyle}
      >
        {Object.values(fieldErrors).some(Boolean)
          ? "Review the highlighted fields."
          : statusText}
      </p>
      <p data-testid={publicErrorCodeTestId("admin")} style={errorStyle}>
        {publicErrorView(error)?.code ?? ""}
      </p>
      <PublicErrorSummary
        error={error}
        testIds={publicErrorSummaryTestIds("admin")}
      />
    </>
  );
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
        {feedback}
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
    selected?.auth_bindings?.filter(isEnterpriseAuthBinding) ?? [];
  const selectedLocalBindings =
    selected?.auth_bindings?.filter(
      (binding) => binding.provider_type === "local",
    ) ?? [];

  return (
    <section style={cardStyle}>
      <DeploymentUserLeaveDialog
        controller={controller}
        fallbackFocusRef={props.fallbackFocusRef}
      />
      <div style={cardHeaderStyle}>
        <div>
          <p style={sectionEyebrowStyle}>Deployment users</p>
          <h2 style={sectionTitleStyle}>User administration</h2>
        </div>
        <button
          data-testid={
            createOpen ? undefined : deploymentAdminTestId("create-user")
          }
          disabled={targetOperationPending}
          style={buttonStyle}
          type="button"
          onClick={() => {
            void commands.createDialog(true);
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
                void commands.refreshUsers();
              }}
            >
              Refresh users
            </button>
          </div>
          {queryStatus === "loading" || queryStatus === "pending" ? (
            <p>Updating users. The last accepted results remain visible.</p>
          ) : null}
          {queryStatus === "failed" ? (
            <p>
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
            value={queryDraft.search}
            onChange={(event) => {
              commands.changeQuery({ search: event.target.value });
            }}
            onKeyDown={(event) => {
              if (event.key === "Enter") {
                void commands.refreshUsers();
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
                value={
                  queryDraft.isActive === null
                    ? "all"
                    : String(queryDraft.isActive)
                }
                onChange={(event) => {
                  commands.changeQuery({
                    isActive:
                      event.target.value === "all"
                        ? null
                        : event.target.value === "true",
                  });
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
                value={
                  queryDraft.isAdmin === null
                    ? "all"
                    : String(queryDraft.isAdmin)
                }
                onChange={(event) => {
                  commands.changeQuery({
                    isAdmin:
                      event.target.value === "all"
                        ? null
                        : event.target.value === "true",
                  });
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
              const isSelected = selected?.user_id === user.user_id;
              return (
                <button
                  key={user.user_id}
                  aria-pressed={isSelected}
                  data-testid={deploymentUserRowTestId(user.user_id)}
                  style={
                    isSelected ? selectedUserRowButtonStyle : userRowButtonStyle
                  }
                  type="button"
                  onClick={() => {
                    void commands.select(user.user_id);
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
              void commands.loadMore();
            }}
          >
            Load more users
          </button>
        </section>

        <div style={userInspectorPaneStyle}>
          {selected === null ? (
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
                  void commands.save();
                }}
                style={inspectorSectionStyle}
              >
                <div style={compactPanelHeaderStyle}>
                  <div>
                    <p style={sectionEyebrowStyle}>Selected user</p>
                    <h3 style={subsectionTitleStyle}>
                      {selected.display_name}
                    </h3>
                  </div>
                  <button
                    style={secondaryButtonStyle}
                    type="button"
                    onClick={() => {
                      void commands.select("");
                    }}
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
                      {selected.user_id}
                    </div>
                  </div>
                  <div>
                    <span style={labelStyle}>User version</span>
                    <div
                      data-testid={deploymentAdminTestId("target-user-version")}
                    >
                      {selected.user_version}
                    </div>
                  </div>
                  <div>
                    <span style={labelStyle}>Is active</span>
                    <div
                      data-testid={deploymentAdminTestId("target-is-active")}
                    >
                      {String(selected.is_active)}
                    </div>
                  </div>
                  <div>
                    <span style={labelStyle}>Deployment admin</span>
                    <div
                      data-testid={deploymentAdminTestId(
                        "target-is-deployment-admin",
                      )}
                    >
                      {String(selected.is_deployment_admin)}
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
                        value={values?.email ?? ""}
                        onChange={(event) => {
                          commands.changeDraft("email", event.target.value);
                        }}
                      />
                    </label>
                    {fieldErrors.email ? (
                      <p id="admin-patch-email-error" style={errorStyle}>
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
                        value={values?.display_name ?? ""}
                        onChange={(event) => {
                          commands.changeDraft(
                            "display_name",
                            event.target.value,
                          );
                        }}
                      />
                    </label>
                    {fieldErrors.display_name ? (
                      <p id="admin-patch-display-name-error" style={errorStyle}>
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
                      checked={values?.mfa_required ?? false}
                      onChange={(event) => {
                        commands.changeDraft(
                          "mfa_required",
                          event.target.checked,
                        );
                      }}
                    />
                    MFA required
                  </label>
                  <label style={checkboxLabelStyle}>
                    <input
                      data-testid={deploymentAdminTestId("patch-is-active")}
                      type="checkbox"
                      checked={values?.is_active ?? false}
                      onChange={(event) => {
                        commands.changeDraft("is_active", event.target.checked);
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
                      checked={values?.is_deployment_admin ?? false}
                      onChange={(event) => {
                        commands.changeDraft(
                          "is_deployment_admin",
                          event.target.checked,
                        );
                      }}
                    />
                    Deployment admin
                  </label>
                </div>
                {reviewRequired ? (
                  <p>
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
                      commands.credentialDialog("password");
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
                      commands.credentialDialog("totp");
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
                      commands.credentialDialog("revoke");
                    }}
                  >
                    Revoke all sessions
                  </button>
                </div>
              </section>

              {enterpriseClaimed ? (
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
                      <div style={monoTextStyle}>{providers.length}</div>
                    </div>
                  </div>

                  <div style={userListStyle}>
                    {selectedEnterpriseBindings.map((binding) => (
                      <label
                        key={binding.auth_binding_id}
                        style={
                          bindingDraft.target === binding.auth_binding_id
                            ? selectedUserRowButtonStyle
                            : userRowButtonStyle
                        }
                      >
                        <span style={checkboxLabelStyle}>
                          <input
                            type="radio"
                            name="enterprise-auth-binding-target"
                            checked={
                              bindingDraft.target === binding.auth_binding_id
                            }
                            disabled={!canSubmitTargetAction}
                            onChange={() => {
                              commands.changeBinding(
                                "target",
                                binding.auth_binding_id,
                              );
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
                    {providers.map((provider) => (
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
                          value={bindingDraft.providerKey}
                          onChange={(event) => {
                            commands.changeBinding(
                              "providerKey",
                              event.target.value,
                            );
                          }}
                        />
                      </label>
                      {fieldErrors.providerKey ? (
                        <p id="admin-enterprise-provider-key-error">
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
                        value={bindingDraft.subject}
                        onChange={(event) => {
                          commands.changeBinding("subject", event.target.value);
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
                          !canSubmitTargetAction || bindingDraft.target === ""
                        }
                        style={inputStyle}
                        value={bindingDraft.newSubject}
                        onChange={(event) => {
                          commands.changeBinding(
                            "newSubject",
                            event.target.value,
                          );
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
                        value={bindingDraft.reason}
                        onChange={(event) => {
                          commands.changeBinding("reason", event.target.value);
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
                        void commands.safeAction("bindingCreate");
                      }}
                    >
                      Create binding
                    </button>
                    <button
                      disabled={
                        !canSubmitTargetAction || bindingDraft.target === ""
                      }
                      style={buttonStyle}
                      type="button"
                      onClick={() => {
                        void commands.safeAction("bindingRotate");
                      }}
                    >
                      Rotate subject
                    </button>
                    <button
                      disabled={
                        !canSubmitTargetAction || bindingDraft.target === ""
                      }
                      style={secondaryButtonStyle}
                      type="button"
                      onClick={() => {
                        void commands.safeAction("bindingRetire");
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
        <p>
          Current user could not be refreshed. The accepted resource and draft
          remain available; refresh this user to retry.
        </p>
      ) : null}
      {operation.kind !== "idle" ? (
        <p>
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
          {operation.recovery.kind === "exact_replay" ? (
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
                disabled={
                  operation.recovery.observation !== "ready" || transportPending
                }
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
      {enterpriseClaimed && providersStatus === "failed" ? (
        <p>
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
      {createOpen ? (
        <DeploymentUserActionDialog
          fallbackFocusRef={props.fallbackFocusRef}
          label="Create local user"
          style={dialogStyle}
          onClose={() => {
            void commands.createDialog(false);
          }}
          onSubmit={() => {
            void commands.create();
          }}
        >
          {(dismiss) => (
            <>
              <header style={dialogHeaderStyle}>
                <div>
                  <p style={sectionEyebrowStyle}>Deployment users</p>
                  <h3 style={subsectionTitleStyle}>Create local user</h3>
                </div>
                <button
                  aria-label="Close create user"
                  style={iconButtonStyle}
                  type="button"
                  onClick={dismiss}
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
                      value={create.email}
                      onChange={(event) => {
                        commands.changeCreate("email", event.target.value);
                      }}
                    />
                  </label>
                  {fieldErrors.createEmail ? (
                    <p id="admin-create-email-error" style={errorStyle}>
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
                      value={create.displayName}
                      onChange={(event) => {
                        commands.changeCreate(
                          "displayName",
                          event.target.value,
                        );
                      }}
                    />
                  </label>
                  {fieldErrors.createDisplayName ? (
                    <p id="admin-create-display-name-error" style={errorStyle}>
                      {fieldErrors.createDisplayName}
                    </p>
                  ) : null}
                </div>
                <div>
                  <label
                    htmlFor="admin-create-password"
                    style={labelBlockStyle}
                  >
                    Initial password
                    <input
                      data-testid={deploymentAdminTestId("create-password")}
                      id="admin-create-password"
                      aria-invalid={
                        fieldErrors.createPassword ? true : undefined
                      }
                      aria-describedby={
                        fieldErrors.createPassword
                          ? "admin-create-password-error"
                          : undefined
                      }
                      disabled={targetOperationPending}
                      style={inputStyle}
                      type="password"
                      value={create.password}
                      onChange={(event) => {
                        commands.changeCreate("password", event.target.value);
                      }}
                    />
                  </label>
                  {fieldErrors.createPassword ? (
                    <p id="admin-create-password-error" style={errorStyle}>
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
                    checked={create.mfaRequired}
                    onChange={(event) => {
                      commands.changeCreate(
                        "mfaRequired",
                        event.target.checked,
                      );
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
                    checked={create.admin}
                    onChange={(event) => {
                      commands.changeCreate("admin", event.target.checked);
                    }}
                  />
                  Deployment admin
                </label>
              </div>
              <div style={dialogButtonRowStyle}>
                <button
                  style={secondaryButtonStyle}
                  type="button"
                  onClick={dismiss}
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
              {feedback}
            </>
          )}
        </DeploymentUserActionDialog>
      ) : null}

      {credentialDialog !== null && selected !== null ? (
        <DeploymentUserActionDialog
          fallbackFocusRef={props.fallbackFocusRef}
          label="Confirm credential action"
          style={dialogStyle}
          onClose={() => commands.credentialDialog(null)}
          onSubmit={() => {
            if (credentialDialog === "password") void commands.resetPassword();
            else if (credentialDialog === "totp")
              void commands.safeAction("totp");
            else void commands.safeAction("revoke");
          }}
        >
          {(dismiss) => (
            <>
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
                  onClick={dismiss}
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
                        value={password}
                        onChange={(event) => {
                          commands.changePassword(event.target.value);
                        }}
                      />
                    </label>
                    {fieldErrors.password ? (
                      <p id="admin-new-password-error" style={errorStyle}>
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
                    value={reason}
                    onChange={(event) => {
                      commands.changeReason(event.target.value);
                    }}
                  />
                </label>
              </div>
              <div style={dialogButtonRowStyle}>
                <button
                  style={secondaryButtonStyle}
                  type="button"
                  onClick={dismiss}
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
              {feedback}
            </>
          )}
        </DeploymentUserActionDialog>
      ) : null}

      {!createOpen && credentialDialog === null && !snapshot.leavePrompt
        ? feedback
        : null}
    </section>
  );
}

const adminWorkspaceStyle: CSSProperties = {
  display: "grid",
  gridTemplateColumns: "repeat(auto-fit, minmax(min(100%, 24rem), 1fr))",
  gap: "var(--ct-spacing-lg)",
  alignItems: "start",
};

const bodyStyle: CSSProperties = {
  margin: "0.75rem 0 0",
  color: "var(--ct-colors-ink-muted)",
  maxWidth: "42rem",
  overflowWrap: "anywhere",
};

const checkboxLabelStyle: CSSProperties = {
  display: "inline-flex",
  alignItems: "center",
  gap: "0.45rem",
  color: "var(--ct-colors-ink-muted)",
};

const checkboxRowStyle: CSSProperties = {
  display: "flex",
  flexWrap: "wrap",
  gap: "1rem",
  marginTop: "1rem",
};

const compactPanelHeaderStyle: CSSProperties = {
  display: "flex",
  justifyContent: "space-between",
  gap: "0.75rem",
  alignItems: "center",
  marginBottom: "1rem",
};

const destructiveButtonStyle: CSSProperties = {
  ...buttonStyle,
  background: "var(--ct-colors-semantic-conflict)",
  color: "var(--ct-colors-surface-1)",
  border: "none",
};

const dialogButtonRowStyle: CSSProperties = {
  display: "flex",
  justifyContent: "flex-end",
  flexWrap: "wrap",
  gap: "0.75rem",
};

const dialogHeaderStyle: CSSProperties = {
  display: "flex",
  justifyContent: "space-between",
  gap: "1rem",
  alignItems: "flex-start",
};

const dialogStyle: CSSProperties = {
  width: "min(44rem, 100%)",
  overflow: "auto",
  display: "grid",
  gap: "1rem",
  padding: "1.25rem",
  borderRadius: "var(--ct-rounded-md)",
  border: "var(--ct-border-strong)",
  background: "var(--ct-colors-surface-1)",
  boxShadow: "var(--ct-elevation-panel)",
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

const userInspectorPaneStyle: CSSProperties = {
  minWidth: 0,
};

const userListPaneStyle: CSSProperties = {
  minWidth: 0,
  padding: "var(--ct-spacing-md)",
  borderRadius: "var(--ct-rounded-md)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
};

const userListStyle: CSSProperties = {
  display: "grid",
  gap: "var(--ct-spacing-sm)",
  margin: "var(--ct-spacing-md) 0",
};

const userRowMetaStyle: CSSProperties = {
  color: "var(--ct-colors-ink-subtle)",
  fontSize: "0.82rem",
};

const userRowPrimaryStyle: CSSProperties = {
  fontWeight: 700,
  overflowWrap: "anywhere",
};

const userRowSecondaryStyle: CSSProperties = {
  color: "var(--ct-colors-ink-muted)",
  overflowWrap: "anywhere",
};
