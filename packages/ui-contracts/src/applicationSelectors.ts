import type { StableTestId } from "./selectorCore";
import {
  encodedTestId,
  encodeSelectorSegment,
  requireClosedToken,
  semanticSelectorTestId,
  stableEncodedTestId,
  stableSelectorTokenTestId,
  stableTestId,
  userScopedTestId,
} from "./selectorCore";

type PublicErrorSurface = "account" | "admin" | "auth" | "landing";

type AuthSelector =
  | "bootstrap-begin"
  | "bootstrap-complete"
  | "bootstrap-complete-code"
  | "bootstrap-enrollment-id"
  | "bootstrap-secret-base32"
  | "bootstrap-token"
  | "enterprise-provider-button"
  | "enterprise-provider-list"
  | "login-password"
  | "login-submit"
  | "login-totp-code"
  | "login-username"
  | "shell"
  | "shell-message"
  | "status";

type AccountSelector =
  | "appearance-density-mode"
  | "appearance-save"
  | "profile-display-name"
  | "profile-email"
  | "profile-save"
  | "credential-auth-kind"
  | "credential-password-changed-at"
  | "credential-pending-expires-at"
  | "credential-recovery-model"
  | "credential-totp-state"
  | "logout"
  | "password-change"
  | "password-current"
  | "password-factor-code"
  | "password-next"
  | "refresh-state"
  | "session-absolute-expires-at"
  | "session-authenticated-at"
  | "session-idle-expires-at"
  | "session-is-deployment-admin"
  | "session-memberships"
  | "session-mfa-state"
  | "session-provider-type"
  | "session-session-expires-at"
  | "session-user-id"
  | "status"
  | "totp-begin"
  | "totp-complete"
  | "totp-complete-code"
  | "totp-current-factor"
  | "totp-current-password"
  | "totp-enrollment-id"
  | "totp-secret-base32";

type DeploymentAdminSelector =
  | "access-note"
  | "create-display-name"
  | "create-email"
  | "create-is-deployment-admin"
  | "create-mfa-required"
  | "create-password"
  | "create-user"
  | "load-more-users"
  | "load-user"
  | "new-password"
  | "password-reset"
  | "patch-base-version"
  | "patch-display-name"
  | "patch-email"
  | "patch-is-active"
  | "patch-is-deployment-admin"
  | "patch-mfa-required"
  | "patch-user"
  | "reason"
  | "revoke-all"
  | "status"
  | "target-is-active"
  | "target-is-deployment-admin"
  | "target-user-id"
  | "target-user-id-input"
  | "target-user-version"
  | "totp-reset"
  | "user-filter"
  | "user-is-active-filter"
  | "user-is-deployment-admin-filter"
  | "user-list";

type IncidentLandingSelector =
  | "create-current-phase"
  | "create-description"
  | "create-external-case"
  | "create-open-button"
  | "create-severity"
  | "create-submit-button"
  | "create-tlp"
  | "current-user"
  | "empty-state"
  | "incident-key"
  | "incident-list"
  | "incident-title"
  | "incidents-count"
  | "loading"
  | "refresh"
  | "return"
  | "search"
  | "shell"
  | "status-filter"
  | "status";

type LandingAdminPanelSelectorToken =
  | "account-appearance"
  | "account-profile"
  | "account-security"
  | "administrative-audit"
  | "deployment-users"
  | "incident-import"
  | "incidents"
  | "reference-packs";

const landingAdminPanelTokens = [
  "incidents",
  "account-profile",
  "account-appearance",
  "account-security",
  "deployment-users",
  "administrative-audit",
  "reference-packs",
  "incident-import",
] as const satisfies readonly LandingAdminPanelSelectorToken[];

type LandingAdminShellSelector = "menu" | "shell" | "status-strip";

type IncidentControlsSectionSelectorToken =
  | "import-assistant"
  | "incident-fields"
  | "membership-audit"
  | "memberships"
  | "summary";

type IncidentAdministrationSelector =
  | "admin-action-message"
  | "admin-error-code"
  | "admin-status"
  | "close-button"
  | "lifecycle-reason"
  | "patch-button"
  | "patch-current-phase"
  | "patch-description"
  | "patch-external-case"
  | "patch-readonly-note"
  | "patch-severity"
  | "patch-tlp"
  | "pref-default-sheet-ref"
  | "pref-home-sheet-ref"
  | "reopen-button"
  | "summary-closed-at"
  | "summary-current-phase"
  | "summary-description"
  | "summary-key"
  | "summary-primary-external-case-ref"
  | "summary-role"
  | "summary-severity"
  | "summary-status"
  | "summary-title"
  | "summary-tlp"
  | "summary-version";

const incidentControlsSections = [
  "summary",
  "import-assistant",
  "incident-fields",
  "memberships",
  "membership-audit",
] as const satisfies readonly IncidentControlsSectionSelectorToken[];

type AppRouteSelector =
  | "app-shell"
  | "debug-harness-loading"
  | "debug-harness-shell"
  | "workbook-current-user"
  | "workbook-loading";

type PublicErrorSummaryTestIds = {
  readonly container: StableTestId;
  readonly details: StableTestId;
  readonly message: StableTestId;
};

const authTestIds = Object.freeze({
  "bootstrap-begin": "auth-bootstrap-begin",
  "bootstrap-complete": "auth-bootstrap-complete",
  "bootstrap-complete-code": "auth-bootstrap-complete-code",
  "bootstrap-enrollment-id": "auth-bootstrap-enrollment-id",
  "bootstrap-secret-base32": "auth-bootstrap-secret-base32",
  "bootstrap-token": "auth-bootstrap-token",
  "enterprise-provider-button": "auth-enterprise-provider-button",
  "enterprise-provider-list": "auth-enterprise-provider-list",
  "login-password": "auth-login-password",
  "login-submit": "auth-login-submit",
  "login-totp-code": "auth-login-totp-code",
  "login-username": "auth-login-username",
  shell: "auth-shell",
  "shell-message": "auth-shell-message",
  status: "auth-status",
} satisfies Record<AuthSelector, string>);

const accountTestIds = Object.freeze({
  "appearance-density-mode": "account-appearance-density-mode",
  "appearance-save": "account-appearance-save",
  "credential-auth-kind": "account-credential-auth-kind",
  "credential-password-changed-at": "account-credential-password-changed-at",
  "credential-pending-expires-at": "account-credential-pending-expires-at",
  "credential-recovery-model": "account-credential-recovery-model",
  "credential-totp-state": "account-credential-totp-state",
  logout: "account-logout",
  "password-change": "account-password-change",
  "password-current": "account-password-current",
  "password-factor-code": "account-password-factor-code",
  "password-next": "account-password-next",
  "profile-display-name": "account-profile-display-name",
  "profile-email": "account-profile-email",
  "profile-save": "account-profile-save",
  "refresh-state": "account-refresh-state",
  "session-absolute-expires-at": "account-session-absolute-expires-at",
  "session-authenticated-at": "account-session-authenticated-at",
  "session-idle-expires-at": "account-session-idle-expires-at",
  "session-is-deployment-admin": "account-session-is-deployment-admin",
  "session-memberships": "account-session-memberships",
  "session-mfa-state": "account-session-mfa-state",
  "session-provider-type": "account-session-provider-type",
  "session-session-expires-at": "account-session-session-expires-at",
  "session-user-id": "account-session-user-id",
  status: "account-status",
  "totp-begin": "account-totp-begin",
  "totp-complete": "account-totp-complete",
  "totp-complete-code": "account-totp-complete-code",
  "totp-current-factor": "account-totp-current-factor",
  "totp-current-password": "account-totp-current-password",
  "totp-enrollment-id": "account-totp-enrollment-id",
  "totp-secret-base32": "account-totp-secret-base32",
} satisfies Record<AccountSelector, string>);

const deploymentAdminTestIds = Object.freeze({
  "access-note": "admin-access-note",
  "create-display-name": "admin-create-display-name",
  "create-email": "admin-create-email",
  "create-is-deployment-admin": "admin-create-is-deployment-admin",
  "create-mfa-required": "admin-create-mfa-required",
  "create-password": "admin-create-password",
  "create-user": "admin-create-user",
  "load-more-users": "admin-load-more-users",
  "load-user": "admin-load-user",
  "new-password": "admin-new-password",
  "password-reset": "admin-password-reset",
  "patch-base-version": "admin-patch-base-version",
  "patch-display-name": "admin-patch-display-name",
  "patch-email": "admin-patch-email",
  "patch-is-active": "admin-patch-is-active",
  "patch-is-deployment-admin": "admin-patch-is-deployment-admin",
  "patch-mfa-required": "admin-patch-mfa-required",
  "patch-user": "admin-patch-user",
  reason: "admin-reason",
  "revoke-all": "admin-revoke-all",
  status: "admin-status",
  "target-is-active": "admin-target-is-active",
  "target-is-deployment-admin": "admin-target-is-deployment-admin",
  "target-user-id": "admin-target-user-id",
  "target-user-id-input": "admin-target-user-id-input",
  "target-user-version": "admin-target-user-version",
  "totp-reset": "admin-totp-reset",
  "user-filter": "admin-user-filter",
  "user-is-active-filter": "admin-user-is-active-filter",
  "user-is-deployment-admin-filter": "admin-user-is-deployment-admin-filter",
  "user-list": "admin-user-list",
} satisfies Record<DeploymentAdminSelector, string>);

const incidentLandingTestIds = Object.freeze({
  "create-current-phase": "landing-create-current-phase",
  "create-description": "landing-create-description",
  "create-external-case": "landing-create-external-case",
  "create-open-button": "landing-create-open-button",
  "create-severity": "landing-create-severity",
  "create-submit-button": "landing-create-submit-button",
  "create-tlp": "landing-create-tlp",
  "current-user": "landing-current-user",
  "empty-state": "landing-empty-state",
  "incident-key": "landing-incident-key",
  "incident-list": "landing-incident-list",
  "incident-title": "landing-incident-title",
  "incidents-count": "landing-incidents-count",
  loading: "landing-loading",
  refresh: "landing-refresh",
  return: "landing-return",
  search: "landing-search",
  shell: "incident-landing",
  "status-filter": "landing-status-filter",
  status: "landing-status",
} satisfies Record<IncidentLandingSelector, string>);

const landingAdminShellTestIds = Object.freeze({
  menu: "landing-admin-menu",
  shell: "landing-admin-shell",
  "status-strip": "landing-admin-status-strip",
} satisfies Record<LandingAdminShellSelector, string>);

const landingAdminPanelTokenSet = Object.freeze(
  new Set<LandingAdminPanelSelectorToken>(landingAdminPanelTokens),
);

const appRouteTestIds = Object.freeze({
  "app-shell": "app-shell",
  "debug-harness-loading": "debug-harness-loading",
  "debug-harness-shell": "debug-harness-shell",
  "workbook-current-user": "workbook-current-user",
  "workbook-loading": "workbook-loading",
} satisfies Record<AppRouteSelector, string>);

const incidentAdministrationTestIds = Object.freeze({
  "admin-action-message": "incident-admin-action-message",
  "admin-error-code": "incident-admin-error-code",
  "admin-status": "incident-admin-status",
  "close-button": "incident-close-button",
  "lifecycle-reason": "incident-lifecycle-reason",
  "patch-button": "incident-patch-button",
  "patch-current-phase": "incident-patch-current-phase",
  "patch-description": "incident-patch-description",
  "patch-external-case": "incident-patch-external-case",
  "patch-readonly-note": "incident-patch-readonly-note",
  "patch-severity": "incident-patch-severity",
  "patch-tlp": "incident-patch-tlp",
  "pref-default-sheet-ref": "incident-pref-default-sheet-ref",
  "pref-home-sheet-ref": "incident-pref-home-sheet-ref",
  "reopen-button": "incident-reopen-button",
  "summary-closed-at": "incident-summary-closed-at",
  "summary-current-phase": "incident-summary-current-phase",
  "summary-description": "incident-summary-description",
  "summary-key": "incident-summary-key",
  "summary-primary-external-case-ref":
    "incident-summary-primary-external-case-ref",
  "summary-role": "incident-summary-role",
  "summary-severity": "incident-summary-severity",
  "summary-status": "incident-summary-status",
  "summary-title": "incident-summary-title",
  "summary-tlp": "incident-summary-tlp",
  "summary-version": "incident-summary-version",
} satisfies Record<IncidentAdministrationSelector, string>);

const publicErrorSurfaces = Object.freeze(
  new Set<PublicErrorSurface>(["account", "admin", "auth", "landing"]),
);

export function authTestId(selector: AuthSelector): StableTestId {
  return semanticSelectorTestId(authTestIds, selector, "auth selector");
}

export function accountTestId(selector: AccountSelector): StableTestId {
  return semanticSelectorTestId(accountTestIds, selector, "account selector");
}

export function deploymentAdminTestId(
  selector: DeploymentAdminSelector,
): StableTestId {
  return semanticSelectorTestId(
    deploymentAdminTestIds,
    selector,
    "deployment admin selector",
  );
}

export function incidentLandingTestId(
  selector: IncidentLandingSelector,
): StableTestId {
  return semanticSelectorTestId(
    incidentLandingTestIds,
    selector,
    "incident landing selector",
  );
}

export function incidentAdministrationTestId(
  selector: IncidentAdministrationSelector,
): StableTestId {
  return semanticSelectorTestId(
    incidentAdministrationTestIds,
    selector,
    "incident administration selector",
  );
}

export function landingAdminShellTestId(
  selector: LandingAdminShellSelector,
): StableTestId {
  return stableSelectorTokenTestId(
    landingAdminShellTestIds,
    selector,
    "landing admin shell selector",
  );
}

export function landingAdminMenuItemTestId(panel: string): StableTestId {
  return stableTestId(
    `landing-admin-menu-item-${requireLandingAdminPanel(panel)}`,
  );
}

export function landingAdminPanelTestId(panel: string): StableTestId {
  return stableTestId(`landing-admin-panel-${requireLandingAdminPanel(panel)}`);
}

export function appRouteTestId(selector: AppRouteSelector): StableTestId {
  return semanticSelectorTestId(
    appRouteTestIds,
    selector,
    "app route selector",
  );
}

export function publicErrorCodeTestId(
  surface: PublicErrorSurface,
): StableTestId {
  return stableTestId(`${requirePublicErrorSurface(surface)}-error-code`);
}

export function publicErrorSummaryTestIds(
  surface: PublicErrorSurface,
): PublicErrorSummaryTestIds {
  const prefix = requirePublicErrorSurface(surface);
  return {
    container: stableTestId(`${prefix}-error-public`),
    details: stableTestId(`${prefix}-error-details`),
    message: stableTestId(`${prefix}-error-message`),
  };
}

export function currentIncidentRoleTestId(): string {
  return "current-incident-role";
}

export function incidentControlsTriggerTestId(): StableTestId {
  return stableTestId("incident-controls-trigger");
}

export function incidentControlsMenuTestId(): StableTestId {
  return stableTestId("incident-controls-menu");
}

export function incidentControlsMenuItemTestId(section: string): StableTestId {
  return stableTestId(
    `incident-controls-menu-item-${requireIncidentControlsSection(section)}`,
  );
}

export function incidentControlsPanelTestId(): StableTestId {
  return stableTestId("incident-controls-panel");
}

export function workbookImportAssistantTestId(): StableTestId {
  return stableTestId("workbook-import-assistant");
}

export function incidentControlsSurfaceTestId(): StableTestId {
  return stableTestId("incident-controls-surface");
}

export function incidentControlsStatusTestId(): StableTestId {
  return incidentAdministrationTestId("admin-status");
}

export function incidentControlsActionMessageTestId(): StableTestId {
  return incidentAdministrationTestId("admin-action-message");
}

export function incidentControlsCloseButtonTestId(): StableTestId {
  return stableTestId("incident-controls-close");
}

export function landingIncidentCardTestId(incidentId: string): StableTestId {
  return stableEncodedTestId("landing-incident", incidentId, "incident_id");
}

export function landingIncidentOpenButtonTestId(
  incidentId: string,
): StableTestId {
  return stableEncodedTestId("landing-open", incidentId, "incident_id");
}

export function deploymentUserRowTestId(userId: string): StableTestId {
  return stableEncodedTestId("deployment-user-row", userId, "user_id");
}

export function incidentMembershipEmailInputTestId(): string {
  return "incident-membership-email";
}

export function incidentMembershipRoleSelectTestId(): string {
  return "incident-membership-role";
}

export function incidentMembershipCreateButtonTestId(): string {
  return "incident-membership-create";
}

export function incidentMembershipAdminNoteTestId(): string {
  return "incident-membership-admin-note";
}

export function incidentMembershipListTestId(): string {
  return "incident-membership-list";
}

export function incidentMembershipRowTestId(userId: string): string {
  return incidentMembershipControlTestId("row", userId);
}

export function incidentMembershipVersionTestId(userId: string): string {
  return incidentMembershipControlTestId("version", userId);
}

export function incidentMembershipRoleInputTestId(userId: string): string {
  return incidentMembershipControlTestId("roleInput", userId);
}

export function incidentMembershipPatchButtonTestId(userId: string): string {
  return incidentMembershipControlTestId("patch", userId);
}

export function incidentMembershipDeleteButtonTestId(userId: string): string {
  return incidentMembershipControlTestId("delete", userId);
}

export function incidentMembershipRoleDisplayTestId(userId: string): string {
  return incidentMembershipControlTestId("roleDisplay", userId);
}

export function incidentMembershipAuditRowTestId(
  auditEventId: string,
): StableTestId {
  return stableEncodedTestId(
    "membership-audit-row",
    auditEventId,
    "audit_event_id",
  );
}

export function debugIncidentRowTestId(incidentId: string): string {
  return encodedTestId("incident-row", incidentId, "incident_id");
}

export function debugSelectIncidentButtonTestId(incidentId: string): string {
  return encodedTestId("select-incident", incidentId, "incident_id");
}

export function debugMembershipRowTestId(userId: string): string {
  return debugMembershipControlTestId("row", userId);
}

export function debugMembershipRoleInputTestId(userId: string): string {
  return debugMembershipControlTestId("roleInput", userId);
}

export function debugMembershipVersionTestId(userId: string): string {
  return debugMembershipControlTestId("version", userId);
}

export function debugMembershipPatchButtonTestId(userId: string): string {
  return debugMembershipControlTestId("patch", userId);
}

export function debugMembershipDeleteButtonTestId(userId: string): string {
  return debugMembershipControlTestId("delete", userId);
}

export function extensionProfileRowTestId(profileId: string): string {
  return encodedTestId("extension", profileId, "extension_profile_id");
}

export function referencePackAdminPanelTestId(): string {
  return "reference-pack-admin-panel";
}

export function referencePackFileInputTestId(): string {
  return "reference-pack-file";
}

export function referencePackImportButtonTestId(): string {
  return "reference-pack-import";
}

export function referencePackJobStatusTestId(): string {
  return "reference-pack-job-status";
}

export function referencePackReloadButtonTestId(): string {
  return "reference-pack-reload";
}

export function referencePackCancelButtonTestId(): string {
  return "reference-pack-cancel";
}

export function referencePackRefreshAllButtonTestId(): string {
  return "reference-pack-refresh-all";
}

export function referencePackRefreshSelectedButtonTestId(): string {
  return "reference-pack-refresh-selected";
}

export function referencePackRowTestId(
  packKey: string,
  packVersion: string,
): string {
  return `reference-pack-row-${encodeSelectorSegment(
    packKey,
    "pack_key",
  )}-${encodeSelectorSegment(packVersion, "pack_version")}`;
}

export function referencePackErrorTestId(): string {
  return "reference-pack-error";
}

function requireLandingAdminPanel(
  panel: string,
): LandingAdminPanelSelectorToken {
  if (landingAdminPanelTokenSet.has(panel as LandingAdminPanelSelectorToken)) {
    return panel as LandingAdminPanelSelectorToken;
  }
  throw new Error(`Invalid landing admin panel token: ${String(panel)}`);
}

function requireIncidentControlsSection(
  section: string,
): IncidentControlsSectionSelectorToken {
  return requireClosedToken(
    incidentControlsSections,
    section as IncidentControlsSectionSelectorToken,
    "incident controls section",
  );
}

const incidentMembershipControlPrefixes = {
  row: "incident-membership-row",
  version: "incident-membership-version",
  roleInput: "incident-membership-role-input",
  patch: "incident-membership-patch",
  delete: "incident-membership-delete",
  roleDisplay: "incident-membership-role",
} as const;

type IncidentMembershipControl = keyof typeof incidentMembershipControlPrefixes;

function incidentMembershipControlTestId(
  control: IncidentMembershipControl,
  userId: string,
): string {
  return userScopedTestId(incidentMembershipControlPrefixes[control], userId);
}

const debugMembershipControlPrefixes = {
  row: "membership-row",
  roleInput: "membership-role-input",
  version: "membership-version",
  patch: "patch-membership",
  delete: "delete-membership",
} as const;

type DebugMembershipControl = keyof typeof debugMembershipControlPrefixes;

function debugMembershipControlTestId(
  control: DebugMembershipControl,
  userId: string,
): string {
  return userScopedTestId(debugMembershipControlPrefixes[control], userId);
}

function requirePublicErrorSurface(
  value: PublicErrorSurface,
): PublicErrorSurface {
  if (publicErrorSurfaces.has(value)) {
    return value;
  }
  throw new Error(`Invalid public error surface selector token: ${value}`);
}
