import { listViewSchemaRegistryEntries } from "@cartulary/protocol-ts";
import { cartularyDesignTokenVars } from "./generated/design-tokens";

export {
  type CartularyDefaultThemeId,
  type CartularyDesignTokenVarName,
  cartularyDefaultThemeId,
  cartularyDesignThemeCssText,
  cartularyDesignTokenVars,
} from "./generated/design-tokens";

export type WorkbookSurface = string;

export type WorkbookGridDensity = "compact" | "default" | "comfortable";

export function workbookGridRowHeightPx(density: WorkbookGridDensity): number {
  const tokenName = `--ct-density-${density}-rowHeight` as const;
  const token = cartularyDesignTokenVars[tokenName];
  const match = /^(\d+)px$/.exec(token);
  if (match === null) {
    throw new Error(
      `Invalid fixed grid row-height token ${tokenName}: ${token}`,
    );
  }
  return Number(match[1]);
}

declare const stableTestIdBrand: unique symbol;

export type StableTestId = string & {
  readonly [stableTestIdBrand]: "StableTestId";
};

export type EntityType = "host" | "identity";

export type EntityMergeControl =
  | "confirm"
  | "loser-record"
  | "message"
  | "plan"
  | "reason"
  | "start";

export type AssessmentCreateControl =
  | "assessed-at"
  | "confidence-band"
  | "message"
  | "rationale"
  | "state"
  | "subject"
  | "subject-type"
  | "submit"
  | "support-refs";

export type GenericWorkbookSelector =
  | "mutation-error"
  | "note-source-record"
  | "reference-load-error";

export type CoordinationWorkflowSelector =
  | "decision-reason"
  | "decision-replacement"
  | "decision-submit"
  | "decision-target"
  | "party-clear-both"
  | "party-clear-link"
  | "party-clear-text"
  | "party-create-from-text"
  | "party-existing"
  | "party-link-existing"
  | "party-pair"
  | "party-partial-completion"
  | "party-retry-created-link"
  | "task-blocked-reason"
  | "task-status"
  | "task-submit"
  | "task-target";

export const entityTypes = [
  "host",
  "identity",
] as const satisfies readonly EntityType[];

export type EntityMentionResolutionStatus =
  | "dismissed"
  | "resolved"
  | "unresolved";

export const entityMentionResolutionStatuses = [
  "unresolved",
  "resolved",
  "dismissed",
] as const satisfies readonly EntityMentionResolutionStatus[];

export type RowHistoryRollbackAction =
  | "change_set"
  | "history_entry"
  | "row_restore";

export const rowHistoryRollbackActions = [
  "change_set",
  "history_entry",
  "row_restore",
] as const satisfies readonly RowHistoryRollbackAction[];

export type RowHistoryDestructiveOperation = "delete" | "restore";

export const rowHistoryDestructiveOperations = [
  "delete",
  "restore",
] as const satisfies readonly RowHistoryDestructiveOperation[];

export type TimelineScalarEditorSurface = "grid" | "inspector";

export const timelineScalarEditorSurfaces = [
  "grid",
  "inspector",
] as const satisfies readonly TimelineScalarEditorSurface[];

export type TimelineInspectorSection =
  | "operational-text"
  | "evidence"
  | "history"
  | "relationships";

export const timelineInspectorSections = [
  "operational-text",
  "relationships",
  "evidence",
  "history",
] as const satisfies readonly TimelineInspectorSection[];

export type WorkbookInspectorPanelId =
  | "details"
  | "relationships"
  | "evidence"
  | "history"
  | "workflow";

export const workbookInspectorPanelIds = [
  "details",
  "relationships",
  "evidence",
  "history",
  "workflow",
] as const satisfies readonly WorkbookInspectorPanelId[];

export type PublicErrorSurface = "account" | "admin" | "auth" | "landing";

export type AuthSelector =
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

export type AccountSelector =
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

export type DeploymentAdminSelector =
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

export type IncidentLandingSelector =
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

export type LandingAdminPanelToken =
  | "account-appearance"
  | "account-profile"
  | "account-security"
  | "administrative-audit"
  | "deployment-users"
  | "incident-import"
  | "incidents"
  | "reference-packs";

export const landingAdminPanelTokens = [
  "incidents",
  "account-profile",
  "account-appearance",
  "account-security",
  "deployment-users",
  "administrative-audit",
  "reference-packs",
  "incident-import",
] as const satisfies readonly LandingAdminPanelToken[];

export type LandingAdminShellSelector = "menu" | "shell" | "status-strip";

export type IncidentControlsSection =
  | "import-assistant"
  | "incident-fields"
  | "membership-audit"
  | "memberships"
  | "summary";

export type IncidentControlsLoadState =
  | "loading"
  | "partial"
  | "synced"
  | "unavailable";

export type IncidentAdministrationSelector =
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

export type WorkbookConflictControl =
  | "activate-origin"
  | "apply-collection"
  | "close"
  | "keep-saved"
  | "merged-value"
  | "paste-navigator"
  | "paste-next"
  | "paste-position"
  | "paste-previous"
  | "use-merged"
  | "use-server-suggestion"
  | "use-unsaved";

export const incidentControlsSections = [
  "summary",
  "import-assistant",
  "incident-fields",
  "memberships",
  "membership-audit",
] as const satisfies readonly IncidentControlsSection[];

export type AppRouteSelector =
  | "app-shell"
  | "debug-harness-loading"
  | "debug-harness-shell"
  | "workbook-current-user"
  | "workbook-loading";

export type WorkbookShellSlot =
  | "inspector"
  | "primary-grid"
  | "status-strip"
  | "top-bar"
  | "view-bar";

export type NetworkAnalysisSelector =
  | "accepted-grid"
  | "accepted-query-apply"
  | "accepted-query-clear"
  | "column-menu"
  | "contributor-close"
  | "contributor-grid"
  | "contributor-drawer"
  | "diagnostics-summary"
  | "delete-cancel"
  | "delete-confirm"
  | "delete-confirmation"
  | "delete-dialog"
  | "delete-trigger"
  | "filters"
  | "graph-panel"
  | "graph-live-region"
  | "graph-scope"
  | "import-input"
  | "import-trigger"
  | "inspector"
  | "inspector-close"
  | "indicator-link-cancel"
  | "indicator-link-confirmation"
  | "indicator-link-dialog"
  | "indicator-link-existing-id"
  | "indicator-link-submit"
  | "layout-reset"
  | "load-fixture"
  | "mapping-apply"
  | "mapping-dialog"
  | "mapping-display-name"
  | "mapping-preview"
  | "mapping-preview-summary"
  | "mapping-profile"
  | "mapping-timestamp-mode"
  | "mapping-timezone"
  | "mapping-unknown-policy"
  | "mode-graph"
  | "mode-rejected"
  | "mode-rows"
  | "page-next"
  | "page-previous"
  | "page-status"
  | "refresh"
  | "rename-cancel"
  | "rename-dialog"
  | "rename-input"
  | "rename-submit"
  | "rename-trigger"
  | "rejected-grid"
  | "rejected-query-apply"
  | "rejected-query-clear"
  | "stale-state"
  | "status-strip"
  | "table-panel"
  | "tab"
  | "workspace"
  | "workspace-header";

export const workbookShellSlots = [
  "top-bar",
  "view-bar",
  "primary-grid",
  "inspector",
  "status-strip",
] as const satisfies readonly WorkbookShellSlot[];

export const workbookShellSlotLabels = {
  inspector: "Inspector",
  "primary-grid": "Primary grid",
  "status-strip": "Status strip",
  "top-bar": "Workbook top bar",
  "view-bar": "View controls",
} as const satisfies Record<WorkbookShellSlot, string>;

export type SystemViewSwitcherGroupToken =
  | "coordination"
  | "optional-artifact-surfaces"
  | "review-learning"
  | "scope-indicators";

export const systemViewSwitcherGroupTokens = [
  "scope-indicators",
  "coordination",
  "review-learning",
  "optional-artifact-surfaces",
] as const satisfies readonly SystemViewSwitcherGroupToken[];

export type PublicErrorSummaryTestIds = {
  readonly container: StableTestId;
  readonly details: StableTestId;
  readonly message: StableTestId;
};

export type RowHistoryItemAnchor = {
  readonly historyItemRef: string;
};

export type RowHistoryActionAnchor = RowHistoryItemAnchor & {
  readonly action: RowHistoryRollbackAction;
};

export type RowHistoryDestructiveAnchor = {
  readonly operation: RowHistoryDestructiveOperation;
};

const registeredViewSchemaIds = Object.freeze(
  new Set(listViewSchemaRegistryEntries().map((entry) => entry.view_schema_id)),
);

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
  new Set<LandingAdminPanelToken>(landingAdminPanelTokens),
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

const workbookConflictControlTestIds = Object.freeze({
  "activate-origin": "conflict-activate-origin",
  "apply-collection": "conflict-apply-collection",
  close: "conflict-close",
  "keep-saved": "conflict-keep-saved",
  "merged-value": "conflict-merged-value",
  "paste-navigator": "paste-conflict-navigator",
  "paste-next": "paste-conflict-next",
  "paste-position": "paste-conflict-position",
  "paste-previous": "paste-conflict-previous",
  "use-merged": "conflict-use-merged",
  "use-server-suggestion": "conflict-use-server-suggestion",
  "use-unsaved": "conflict-use-unsaved",
} satisfies Record<WorkbookConflictControl, string>);

const entityMergeControlTestIds = Object.freeze({
  confirm: "merge-confirm",
  "loser-record": "merge-loser-record",
  message: "merge-message",
  plan: "merge-plan",
  reason: "merge-reason",
  start: "merge-start",
} satisfies Record<EntityMergeControl, string>);

const assessmentCreateControlTestIds = Object.freeze({
  "assessed-at": "assessment-create-assessed-at",
  "confidence-band": "assessment-create-confidence-band",
  message: "assessment-create-message",
  rationale: "assessment-create-rationale",
  state: "assessment-create-state",
  subject: "assessment-create-subject",
  "subject-type": "assessment-create-subject-type",
  submit: "assessment-create-submit",
  "support-refs": "assessment-create-support-refs",
} satisfies Record<AssessmentCreateControl, string>);

const genericWorkbookTestIds = Object.freeze({
  "mutation-error": "generic-mutation-error",
  "note-source-record": "generic-create-note-source-record",
  "reference-load-error": "generic-reference-load-error",
} satisfies Record<GenericWorkbookSelector, string>);

const coordinationWorkflowTestIds = Object.freeze({
  "decision-reason": "decision-supersede-reason",
  "decision-replacement": "decision-supersede-replacement",
  "decision-submit": "decision-supersede-submit",
  "decision-target": "decision-supersede-target",
  "party-clear-both": "party-link-clear-both",
  "party-clear-link": "party-link-clear-link",
  "party-clear-text": "party-link-clear-text",
  "party-create-from-text": "party-link-create-from-text",
  "party-existing": "party-link-existing-party",
  "party-link-existing": "party-link-link-existing",
  "party-pair": "party-link-pair",
  "party-partial-completion": "party-link-partial-completion",
  "party-retry-created-link": "party-link-retry-created",
  "task-blocked-reason": "task-lifecycle-blocked-reason",
  "task-status": "task-lifecycle-status",
  "task-submit": "task-lifecycle-submit",
  "task-target": "task-lifecycle-target",
} satisfies Record<CoordinationWorkflowSelector, string>);

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

export function workbookConflictControlTestId(
  control: WorkbookConflictControl,
): StableTestId {
  return semanticSelectorTestId(
    workbookConflictControlTestIds,
    control,
    "workbook conflict control",
  );
}

export function entityMergeControlTestId(
  control: EntityMergeControl,
): StableTestId {
  return semanticSelectorTestId(
    entityMergeControlTestIds,
    control,
    "entity merge control",
  );
}

export function assessmentCreateControlTestId(
  control: AssessmentCreateControl,
): StableTestId {
  return semanticSelectorTestId(
    assessmentCreateControlTestIds,
    control,
    "assessment create control",
  );
}

export function genericWorkbookTestId(
  selector: GenericWorkbookSelector,
): StableTestId {
  return semanticSelectorTestId(
    genericWorkbookTestIds,
    selector,
    "generic workbook selector",
  );
}

export function coordinationWorkflowTestId(
  selector: CoordinationWorkflowSelector,
): StableTestId {
  return semanticSelectorTestId(
    coordinationWorkflowTestIds,
    selector,
    "coordination workflow selector",
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

export function landingAdminMenuItemTestId(
  panel: LandingAdminPanelToken,
): StableTestId {
  return stableTestId(
    `landing-admin-menu-item-${requireLandingAdminPanel(panel)}`,
  );
}

export function landingAdminPanelTestId(
  panel: LandingAdminPanelToken,
): StableTestId {
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

export function gridShellTestId(viewSchemaId: WorkbookSurface): string {
  return viewFirstTestId(viewSchemaId, "grid-shell");
}

export function surfaceTabTestId(viewSchemaId: string): string {
  return viewScopedTestId("surface-tab", viewSchemaId);
}

export function systemViewSelectorTestId(): string {
  return systemViewSwitcherTriggerTestId();
}

export function systemViewSwitcherTriggerTestId(): StableTestId {
  return stableTestId("system-view-selector");
}

export function systemViewSwitcherMenuTestId(): StableTestId {
  return stableTestId("system-view-switcher-menu");
}

export function systemViewSwitcherGroupTestId(
  groupToken: SystemViewSwitcherGroupToken,
): StableTestId {
  return stableTestId(
    `system-view-switcher-group-${requireSystemViewSwitcherGroupToken(groupToken)}`,
  );
}

export function systemViewSwitcherOptionTestId(
  groupToken: SystemViewSwitcherGroupToken,
  viewSchemaId: string,
): StableTestId {
  return stableTestId(
    `system-view-switcher-option-${requireSystemViewSwitcherGroupToken(
      groupToken,
    )}-${requireViewSchemaId(viewSchemaId)}`,
  );
}

export function workbookIncidentIdentityTestId(): StableTestId {
  return stableTestId("workbook-incident-identity");
}

export function workbookResponsiveBandTestId(): StableTestId {
  return stableTestId("workbook-responsive-band");
}

export function workbookSurfacesMenuTriggerTestId(): StableTestId {
  return stableTestId("workbook-surfaces-menu-trigger");
}

export function workbookSurfacesMenuTestId(): StableTestId {
  return stableTestId("workbook-surfaces-menu");
}

export function workbookSurfacesMenuOptionTestId(
  viewSchemaId: string,
): StableTestId {
  return stableTestId(
    viewScopedTestId("workbook-surfaces-menu-option", viewSchemaId),
  );
}

export function workbookViewBarQueryControlsTestId(
  viewSchemaId: WorkbookSurface,
): StableTestId {
  return stableTestId(viewFirstTestId(viewSchemaId, "view-bar-query"));
}

export function workbookSortMenuTriggerTestId(
  viewSchemaId: WorkbookSurface,
): StableTestId {
  return stableTestId(viewFirstTestId(viewSchemaId, "sort-menu-trigger"));
}

export function workbookSortMenuTestId(
  viewSchemaId: WorkbookSurface,
): StableTestId {
  return stableTestId(viewFirstTestId(viewSchemaId, "sort-menu"));
}

export function workbookSortOptionTestId(
  viewSchemaId: WorkbookSurface,
  fieldKey: string,
): StableTestId {
  return stableTestId(
    viewFirstTestId(viewSchemaId, `sort-option-${requireFieldKey(fieldKey)}`),
  );
}

export function workbookFilterPopoverTriggerTestId(
  viewSchemaId: WorkbookSurface,
): StableTestId {
  return stableTestId(viewFirstTestId(viewSchemaId, "filter-popover-trigger"));
}

export function workbookFilterPopoverTestId(
  viewSchemaId: WorkbookSurface,
): StableTestId {
  return stableTestId(viewFirstTestId(viewSchemaId, "filter-popover"));
}

export function workbookShellReadyTestId(): string {
  return "workbook-shell-ready";
}

export function workbookShellSlotTestId(slot: WorkbookShellSlot): StableTestId {
  return stableTestId(`workbook-shell-slot-${requireWorkbookShellSlot(slot)}`);
}

export function workbookShellSlotLabel(slot: WorkbookShellSlot): string {
  return workbookShellSlotLabels[requireWorkbookShellSlot(slot)];
}

export function networkAnalysisTestId(
  selector: NetworkAnalysisSelector,
): StableTestId {
  return stableTestId(`network-flow-analysis-${selector}`);
}

export function networkAnalysisTableTabTestId(
  networkFlowTableId: string,
): StableTestId {
  return stableTestId(
    encodedTestId("network-flow-table-tab", networkFlowTableId, "table_id"),
  );
}

export function networkAnalysisEdgeTestId(edgeId: string): StableTestId {
  return stableTestId(encodedTestId("network-flow-edge", edgeId, "edge_id"));
}

export function networkAnalysisVertexTestId(vertexId: string): StableTestId {
  return stableTestId(
    encodedTestId("network-flow-vertex", vertexId, "vertex_id"),
  );
}

export function networkAnalysisRowTestId(rowId: string): StableTestId {
  return stableTestId(encodedTestId("network-flow-row", rowId, "row_id"));
}

export function networkAnalysisRowCellTestId(
  rowId: string,
  fieldKey: string,
): StableTestId {
  return stableTestId(
    `${encodedTestId("network-flow-row-cell", rowId, "row_id")}-${encodeSelectorSegment(fieldKey, "field_key")}`,
  );
}

export function networkAnalysisDiagnosticTestId(
  diagnosticId: string,
): StableTestId {
  return stableTestId(
    encodedTestId("network-flow-diagnostic", diagnosticId, "diagnostic_id"),
  );
}

export function networkAnalysisDiagnosticCellTestId(
  diagnosticId: string,
  fieldKey: string,
): StableTestId {
  return stableTestId(
    `${encodedTestId("network-flow-diagnostic-cell", diagnosticId, "diagnostic_id")}-${encodeSelectorSegment(fieldKey, "field_key")}`,
  );
}

export function networkAnalysisColumnActionTestId(
  fieldKey: string,
  action: "move-earlier" | "move-later" | "toggle",
): StableTestId {
  return stableTestId(
    `network-flow-column-${encodeSelectorSegment(fieldKey, "field_key")}-${action}`,
  );
}

export function networkAnalysisMappingColumnTestId(
  sourceColumnOrdinal: number,
): StableTestId {
  if (!Number.isSafeInteger(sourceColumnOrdinal) || sourceColumnOrdinal < 1) {
    throw new Error("source_column_ordinal must be a positive safe integer");
  }
  return stableTestId(`network-flow-mapping-column-${sourceColumnOrdinal}`);
}

export function workbookAddRowButtonTestId(
  viewSchemaId: WorkbookSurface,
): StableTestId {
  return stableTestId(viewFirstTestId(viewSchemaId, "add-row"));
}

export function workbookInspectorToggleTestId(
  viewSchemaId: WorkbookSurface,
): StableTestId {
  return stableTestId(viewFirstTestId(viewSchemaId, "inspector-toggle"));
}

export function workbookInspectorCloseButtonTestId(
  viewSchemaId: WorkbookSurface,
): StableTestId {
  return stableTestId(viewFirstTestId(viewSchemaId, "inspector-close"));
}

export function workbookInspectorPanelTestId(
  viewSchemaId: WorkbookSurface,
  panelId: WorkbookInspectorPanelId,
): StableTestId {
  return stableTestId(
    viewFirstTestId(
      viewSchemaId,
      `inspector-panel-${requireWorkbookInspectorPanelId(panelId)}`,
    ),
  );
}

export function workbookInspectorFeatureGroupTestId(
  viewSchemaId: WorkbookSurface,
  featureGroupKey: string,
): StableTestId {
  return stableTestId(
    viewFirstTestId(
      viewSchemaId,
      `inspector-feature-${requireFeatureGroupKey(featureGroupKey)}`,
    ),
  );
}

export function workbookInspectorFeatureActionTestId(
  viewSchemaId: WorkbookSurface,
  featureGroupKey: string,
): StableTestId {
  return stableTestId(
    viewFirstTestId(
      viewSchemaId,
      `inspector-feature-action-${requireFeatureGroupKey(featureGroupKey)}`,
    ),
  );
}

export function workbookInlineDraftRowTestId(
  viewSchemaId: WorkbookSurface,
): StableTestId {
  return stableTestId(viewFirstTestId(viewSchemaId, "inline-draft-row"));
}

export function workbookRowActionMenuButtonTestId(
  viewSchemaId: WorkbookSurface,
  recordId: string,
): StableTestId {
  return stableTestId(
    viewFirstTestId(
      viewSchemaId,
      `row-action-menu-${requireRecordId(recordId)}`,
    ),
  );
}

export function workbookRowContextMenuTestId(
  viewSchemaId: WorkbookSurface,
  recordId: string,
): StableTestId {
  return stableTestId(
    viewFirstTestId(
      viewSchemaId,
      `row-context-menu-${requireRecordId(recordId)}`,
    ),
  );
}

export function timelineMutationSubstrateReadyTestId(): string {
  return "timeline-mutation-substrate-ready";
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

export function incidentControlsMenuItemTestId(
  section: IncidentControlsSection,
): StableTestId {
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

export function timelineInspectorSectionTestId(
  section: TimelineInspectorSection,
): StableTestId {
  return stableTestId(
    `timeline-inspector-section-${requireTimelineInspectorSection(section)}`,
  );
}

export function timelineInspectorMessageTestId(): StableTestId {
  return stableTestId("timeline-inspector-message");
}

export function timelineInspectorTestId(): StableTestId {
  return stableTestId("timeline-inspector");
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

export function gridScrollportClassName(): string {
  return "cartulary-grid-scrollport";
}

export function gridScrollportSelector(): string {
  return `.${gridScrollportClassName()}`;
}

export function gridActionsHeaderTestId(viewSchemaId: WorkbookSurface): string {
  return viewFirstTestId(viewSchemaId, "actions-header");
}

export function gridRowGutterTestId(
  viewSchemaId: WorkbookSurface,
  recordId: string,
): string {
  return viewFirstTestId(
    viewSchemaId,
    `row-gutter-${requireRecordId(recordId)}`,
  );
}

/**
 * Scope this selector through `gridShellTestId(surface)` when targeting
 * workbook rows. Do not rely on raw table markup or renderer classes.
 */
export function gridSavedRowsSelector(): string {
  return '[role="row"][data-grid-record-id]:not([data-grid-record-id=""])';
}

/**
 * Scope these selectors through an owner grid shell. They describe adapter-
 * owned semantic data rows and cells without exposing vendor classes or
 * positional coordinates.
 */
export function gridDataRowsSelector(): string {
  return '[role="row"][data-cartulary-grid-row-kind="data"]';
}

export function gridDataCellsSelector(): string {
  return `${gridDataRowsSelector()} [role="gridcell"]`;
}

export function gridSavedRowSelector(recordId: string): string {
  return `[role="row"][data-grid-record-id="${cssAttributeValue(
    requireNonEmptySelectorValue(recordId, "record_id"),
  )}"]`;
}

/**
 * Scope this selector through `gridShellTestId(surface)` when targeting the
 * workbook draft row. Do not rely on raw table markup or renderer classes.
 */
export function gridDraftRowSelector(): string {
  return '[role="row"][data-cartulary-grid-draft-row="true"]';
}

export function conflictMarkerTestId(
  recordId: string,
  fieldKey: string,
): string {
  return recordFieldTestId("conflict-marker", recordId, fieldKey);
}

export function workbookConflictResolverTestId(): string {
  return "workbook-conflict-resolver";
}

export function workbookConflictSummaryTestId(): string {
  return "workbook-conflict-summary";
}

export function workbookConflictSavedValueTestId(): string {
  return "workbook-conflict-saved-value";
}

export function workbookConflictLocalValueTestId(): string {
  return "workbook-conflict-local-value";
}

export function workbookEditRecoveryTestId(): string {
  return "workbook-edit-recovery";
}

export function workbookEditRecoveryRetryButtonTestId(): string {
  return "workbook-edit-recovery-retry";
}

export function workbookEditRecoveryDiscardButtonTestId(): string {
  return "workbook-edit-recovery-discard";
}

export function rowPresenceMarkerTestId(recordId: string): string {
  return `presence-row-${requireRecordId(recordId)}`;
}

export function cellPresenceMarkerTestId(
  recordId: string,
  fieldKey: string,
): string {
  return recordFieldTestId("presence-cell", recordId, fieldKey);
}

export function saveStateTestId(): string {
  return "save-state";
}

export function saveStateActionButtonTestId(): string {
  return "save-state-action";
}

export function statusStripQueueCountTestId(): string {
  return "status-strip-queue-count";
}

export function workbookFocusAnchorTestId(): string {
  return "workbook-focus-anchor";
}

export function workbookPresenceSummaryTestId(): string {
  return "presence-header";
}

export function pendingQueueNoticeTestId(): string {
  return "pending-queue-notice";
}

export function pendingQueueCountTestId(): string {
  return "pending-queue-count";
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

export function gridSortHeaderTestId(
  viewSchemaId: WorkbookSurface,
  fieldKey: string,
): string {
  return viewFirstTestId(viewSchemaId, `sort-${requireFieldKey(fieldKey)}`);
}

export function gridFilterChipTestId(
  viewSchemaId: WorkbookSurface,
  fieldKey: string,
): string {
  return viewFirstTestId(
    viewSchemaId,
    `filter-chip-${requireFieldKey(fieldKey)}`,
  );
}

export function gridFilterFieldTestId(viewSchemaId: WorkbookSurface): string {
  return viewFirstTestId(viewSchemaId, "filter-field");
}

export function gridFilterValueTestId(viewSchemaId: WorkbookSurface): string {
  return viewFirstTestId(viewSchemaId, "filter-value");
}

export function gridFilterApplyTestId(viewSchemaId: WorkbookSurface): string {
  return viewFirstTestId(viewSchemaId, "filter-apply");
}

export function gridGroupingSelectTestId(
  viewSchemaId: WorkbookSurface,
): string {
  return viewFirstTestId(viewSchemaId, "group-by");
}

export function gridGroupRowTestId(
  viewSchemaId: WorkbookSurface,
  fieldKey: string,
  value: string,
): string {
  return `${gridGroupRowTestIdPrefix(viewSchemaId, fieldKey)}${encodeSelectorSegment(value, "group value")}`;
}

export function gridGroupRowTestIdPrefix(
  viewSchemaId: WorkbookSurface,
  fieldKey: string,
): string {
  return `${viewFirstTestId(viewSchemaId, `group-${requireFieldKey(fieldKey)}`)}-`;
}

export function gridGroupRowsSelector(
  viewSchemaId: WorkbookSurface,
  fieldKey: string,
): string {
  return dataTestIdPrefixSelector(
    gridGroupRowTestIdPrefix(viewSchemaId, fieldKey),
  );
}

export function gridRowTestId(
  viewSchemaId: WorkbookSurface,
  recordId: string,
): string {
  return `${viewScopedTestId("grid-row", viewSchemaId)}-${requireRecordId(recordId)}`;
}

export function rowCellTestId(recordId: string, fieldKey: string): string {
  return recordFieldTestId("row", recordId, fieldKey);
}

export function timelineScalarEditorTestId(options: {
  readonly fieldKey: string;
  readonly recordId: string | null;
  readonly surface?: TimelineScalarEditorSurface | undefined;
}): string {
  const base =
    options.recordId === null
      ? draftCellTestId(options.fieldKey)
      : rowCellTestId(options.recordId, options.fieldKey);
  if (options.surface === "inspector") return `${base}-inspector`;
  return options.recordId === null ? base : `${base}-grid-editor`;
}

export function rowInspectorFieldTestId(
  recordId: string,
  fieldKey: string,
): string {
  return recordFieldTestId("row", recordId, fieldKey, "inspector");
}

export function rowInspectButtonTestId(recordId: string): string {
  return recordTestId("row", recordId, "inspect");
}

export function rowHistoryOpenButtonTestId(recordId: string): string {
  return recordTestId("row-history-open", recordId);
}

export function rowHistoryOpenInspectorButtonTestId(recordId: string): string {
  return `${rowHistoryOpenButtonTestId(recordId)}-inspector`;
}

export function rowHistoryPanelTestId(): string {
  return "row-history-panel";
}

export function rowHistoryOpenSelectedButtonTestId(): string {
  return "row-history-open-selected";
}

export function rowHistoryLoadingTestId(): string {
  return "row-history-loading";
}

export function rowHistoryMessageTestId(): string {
  return "row-history-message";
}

export function rowHistoryDeleteButtonTestId(): string {
  return "row-history-delete";
}

export function rowHistoryRestoreButtonTestId(): string {
  return "row-history-restore";
}

export function rowHistoryDestructiveConfirmPanelTestId(
  anchor: RowHistoryDestructiveAnchor,
): string {
  return `row-history-destructive-confirm-${requireRowHistoryDestructiveOperation(anchor.operation)}`;
}

export function rowHistoryDestructiveConfirmButtonTestId(
  anchor: RowHistoryDestructiveAnchor,
): string {
  return `${rowHistoryDestructiveConfirmPanelTestId(anchor)}-confirm`;
}

export function rowHistoryDestructiveCancelButtonTestId(
  anchor: RowHistoryDestructiveAnchor,
): string {
  return `${rowHistoryDestructiveConfirmPanelTestId(anchor)}-cancel`;
}

export function rowHistoryItemTestId(anchor: RowHistoryItemAnchor): string {
  return `row-history-item-${encodeSelectorSegment(
    rowHistoryItemIdentity(anchor),
    "row history item identity",
  )}`;
}

export function rowHistoryActionTestId(anchor: RowHistoryActionAnchor): string {
  const action = requireRowHistoryRollbackAction(anchor.action);
  return `row-history-action-${encodeSelectorSegment(
    rowHistoryActionIdentity(anchor),
    "row history action identity",
  )}-${action}`;
}

export function rowHistoryRollbackPreviewTestId(
  anchor: RowHistoryActionAnchor,
): string {
  const action = requireRowHistoryRollbackAction(anchor.action);
  return `row-history-rollback-preview-${encodeSelectorSegment(
    rowHistoryActionIdentity(anchor),
    "row history action identity",
  )}-${action}`;
}

export function rowHistoryRollbackConfirmButtonTestId(
  anchor: RowHistoryActionAnchor,
): string {
  return `${rowHistoryRollbackPreviewTestId(anchor)}-confirm`;
}

export function rowHistoryRollbackCancelButtonTestId(
  anchor: RowHistoryActionAnchor,
): string {
  return `${rowHistoryRollbackPreviewTestId(anchor)}-cancel`;
}

export function draftCellTestId(fieldKey: string): string {
  return draftFieldTestId(fieldKey);
}

export function draftRowCreateButtonTestId(): string {
  return "draft-row-create";
}

export function relationshipItemsTestId(
  recordId: string,
  fieldKey: string,
): string {
  return timelineCollectionFieldControlTestId(recordId, fieldKey, "items");
}

export function relationshipOverflowButtonTestId(
  recordId: string,
  fieldKey: string,
): string {
  return timelineCollectionFieldControlTestId(recordId, fieldKey, "overflow");
}

export function draftRelationshipItemsTestId(fieldKey: string): string {
  return timelineCollectionFieldControlTestId(null, fieldKey, "items");
}

export function timelineCollectionInputTestId(
  recordId: string,
  fieldKey: string,
): string {
  return timelineCollectionFieldControlTestId(recordId, fieldKey, "input");
}

export function draftTimelineCollectionInputTestId(fieldKey: string): string {
  return timelineCollectionFieldControlTestId(null, fieldKey, "input");
}

export function timelineRowVersionTestId(recordId: string): string {
  return rowCellTestId(recordId, "row_version");
}

export function timelineRowMarkReviewedButtonTestId(recordId: string): string {
  return recordTestId("row", recordId, "mark-reviewed");
}

export function timelineRowReplacementInputTestId(recordId: string): string {
  return recordTestId("row", recordId, "replacement-id");
}

export function timelineRowSupersedeButtonTestId(recordId: string): string {
  return recordTestId("row", recordId, "supersede");
}

export function timelineEvidenceFileInputTestId(recordId: string): string {
  return recordTestId("timeline-evidence-file", recordId);
}

export function timelineDraftEvidenceFileInputTestId(): string {
  return "timeline-evidence-file-draft";
}

export function timelineEvidenceAttachSectionTestId(recordId: string): string {
  return recordTestId("timeline-evidence-attach", recordId);
}

export function timelineDraftEvidenceAttachSectionTestId(): string {
  return "timeline-evidence-attach-draft";
}

export function timelinePreviewRowTestId(recordId: string): string {
  return recordTestId("timeline-preview-row", recordId);
}

export function relationshipChipTestId(itemRef: string): string {
  return itemRefTestId("chip", itemRef);
}

export function mentionItemTestId(itemRef: string): string {
  return itemRefTestId("mention", itemRef);
}

export function autoResolutionNoticeTestId(itemRef: string): string {
  return itemRefTestId("auto-resolution-notice", itemRef);
}

export function autoResolutionNoticeFamilySelector(): string {
  return dataTestIdPrefixSelector("auto-resolution-notice-");
}

export function autoResolutionUndoButtonTestId(itemRef: string): string {
  return `${autoResolutionNoticeTestId(itemRef)}-undo`;
}

export function autoResolutionReviewButtonTestId(itemRef: string): string {
  return `${autoResolutionNoticeTestId(itemRef)}-review`;
}

export function pasteConflictItemTestId(itemKey: string): string {
  return `paste-conflict-item-${encodeSelectorSegment(itemKey, "paste conflict key")}`;
}

export function entityMentionResolutionStatusTestId(value: string): string {
  return `entity-mention-resolution-status-${requireEntityMentionResolutionStatus(value)}`;
}

export function entityInspectButtonTestId(
  entityType: EntityType,
  recordId: string,
): string {
  return `inspect-${requireEntityType(entityType)}-${requireRecordId(recordId)}`;
}

export function entityInspectorTestId(entityType: EntityType): string {
  return `${requireEntityType(entityType)}-inspector`;
}

export function entityInspectorSubjectTestId(
  entityType: EntityType,
  recordId: string,
): string {
  return recordTestId(
    `${requireEntityType(entityType)}-inspector-subject`,
    recordId,
  );
}

export function entityReusableIdentifiersSectionTestId(
  entityType: EntityType,
  recordId: string,
): string {
  return recordTestId(
    `${requireEntityType(entityType)}-reusable-identifiers`,
    recordId,
  );
}

export function entityReusableIdentifierItemTestId(
  entityType: EntityType,
  recordId: string,
  itemRef: string,
): string {
  return tokenScopedTestId(
    entityReusableIdentifiersSectionTestId(entityType, recordId),
    requireItemRef(itemRef),
  );
}

export function entityMergePreconditionDetailsTestId(
  entityType: EntityType,
  recordId: string,
): string {
  return recordTestId(
    `${requireEntityType(entityType)}-merge-precondition-details`,
    recordId,
  );
}

export function assessmentCreatePanelTestId(): string {
  return "assessment-create-panel";
}

export function evidencePreviewButtonTestId(recordId: string): string {
  return evidenceRecordControlTestId("preview", recordId);
}

export function evidenceDownloadButtonTestId(recordId: string): string {
  return evidenceRecordControlTestId("download", recordId);
}

export function evidenceAttachFileInputTestId(recordId: string): string {
  return evidenceRecordControlTestId("attach-file", recordId);
}

export function evidenceAccessMessageTestId(recordId: string): string {
  return evidenceRecordControlTestId("access-message", recordId);
}

export function evidencePreviewFrameTestId(recordId: string): string {
  return evidenceRecordControlTestId("preview-frame", recordId);
}

export function evidencePreviewPanelTestId(): string {
  return "evidence-preview-panel";
}

export function genericCreateFieldTestId(fieldKey: string): string {
  return `generic-create-field-${requireFieldKey(fieldKey)}`;
}

export function genericCreateSubmitTestId(viewSchemaId: string): string {
  return viewScopedTestId("generic-create-submit", viewSchemaId);
}

export function genericEditRecordSelectTestId(viewSchemaId: string): string {
  return genericEditControlTestId("record", viewSchemaId);
}

export function genericEditFieldSelectTestId(viewSchemaId: string): string {
  return genericEditControlTestId("field", viewSchemaId);
}

export function genericEditActionSelectTestId(viewSchemaId: string): string {
  return genericEditControlTestId("action", viewSchemaId);
}

export function genericEditValueTestId(viewSchemaId: string): string {
  return genericEditControlTestId("value", viewSchemaId);
}

export function genericEditSubmitTestId(viewSchemaId: string): string {
  return genericEditControlTestId("submit", viewSchemaId);
}

export function mentionResolveTargetSelectTestId(): string {
  return "inspector-resolve-target";
}

export function mentionResolveExistingButtonTestId(): string {
  return "inspector-resolve-existing";
}

export function mentionCreateEntityButtonTestId(
  entityType: EntityType,
): string {
  return `inspector-create-${requireEntityType(entityType)}`;
}

export function mentionDismissButtonTestId(): string {
  return "inspector-dismiss-mention";
}

export function mentionRestoreUnresolvedButtonTestId(): string {
  return "inspector-restore-unresolved";
}

export function savedViewFamilySelector(): string {
  return dataTestIdPrefixSelector("saved-view-");
}

export function savedViewSelectorTestId(viewSchemaId: string): StableTestId {
  return stableTestId(viewScopedTestId("saved-view-selector", viewSchemaId));
}

export function savedViewOptionTestId(
  viewSchemaId: string,
  savedViewId: string,
): StableTestId {
  return stableTestId(
    `${viewScopedTestId("saved-view-option", viewSchemaId)}-${encodeSelectorSegment(
      savedViewId,
      "saved_view_id",
    )}`,
  );
}

export function savedViewNameInputTestId(viewSchemaId: string): StableTestId {
  return stableTestId(viewScopedTestId("saved-view-name", viewSchemaId));
}

export function savedViewScopeSelectTestId(viewSchemaId: string): StableTestId {
  return stableTestId(viewScopedTestId("saved-view-scope", viewSchemaId));
}

export function savedViewActionMenuTriggerTestId(
  viewSchemaId: string,
): StableTestId {
  return stableTestId(
    viewScopedTestId("saved-view-action-menu-trigger", viewSchemaId),
  );
}

export function savedViewActionMenuTestId(viewSchemaId: string): StableTestId {
  return stableTestId(viewScopedTestId("saved-view-action-menu", viewSchemaId));
}

export function savedViewCreateButtonTestId(
  viewSchemaId: string,
): StableTestId {
  return stableTestId(viewScopedTestId("saved-view-create", viewSchemaId));
}

export function savedViewDuplicateButtonTestId(
  viewSchemaId: string,
  savedViewId: string,
): StableTestId {
  return stableTestId(
    `${viewScopedTestId("saved-view-duplicate", viewSchemaId)}-${encodeSelectorSegment(
      savedViewId,
      "saved_view_id",
    )}`,
  );
}

export function savedViewUpdateButtonTestId(
  viewSchemaId: string,
  savedViewId: string,
): StableTestId {
  return stableTestId(
    `${viewScopedTestId("saved-view-update", viewSchemaId)}-${encodeSelectorSegment(
      savedViewId,
      "saved_view_id",
    )}`,
  );
}

export function savedViewDeleteButtonTestId(
  viewSchemaId: string,
  savedViewId: string,
): StableTestId {
  return stableTestId(
    `${viewScopedTestId("saved-view-delete", viewSchemaId)}-${encodeSelectorSegment(
      savedViewId,
      "saved_view_id",
    )}`,
  );
}

export function savedViewSetHomeButtonTestId(
  viewSchemaId: string,
): StableTestId {
  return stableTestId(viewScopedTestId("saved-view-set-home", viewSchemaId));
}

export function savedViewSetDefaultButtonTestId(
  viewSchemaId: string,
): StableTestId {
  return stableTestId(viewScopedTestId("saved-view-set-default", viewSchemaId));
}

export function savedViewModifiedTestId(viewSchemaId: string): StableTestId {
  return stableTestId(viewScopedTestId("saved-view-modified", viewSchemaId));
}

export function savedViewRenameButtonTestId(
  viewSchemaId: string,
  savedViewId: string,
): StableTestId {
  return stableTestId(
    `${viewScopedTestId("saved-view-rename", viewSchemaId)}-${encodeSelectorSegment(
      savedViewId,
      "saved_view_id",
    )}`,
  );
}

export function savedViewManageSharingButtonTestId(
  viewSchemaId: string,
  savedViewId: string,
): StableTestId {
  return stableTestId(
    `${viewScopedTestId(
      "saved-view-manage-sharing",
      viewSchemaId,
    )}-${encodeSelectorSegment(savedViewId, "saved_view_id")}`,
  );
}

export function savedViewResetButtonTestId(
  viewSchemaId: string,
  savedViewId: string,
): StableTestId {
  return stableTestId(
    `${viewScopedTestId("saved-view-reset", viewSchemaId)}-${encodeSelectorSegment(
      savedViewId,
      "saved_view_id",
    )}`,
  );
}

export function savedViewStatusTestId(viewSchemaId: string): StableTestId {
  return stableTestId(viewScopedTestId("saved-view-status", viewSchemaId));
}

export function dataTestIdPrefixSelector(testIdPrefix: string): string {
  return `[data-testid^="${cssAttributeValue(
    requireNonEmptySelectorValue(testIdPrefix, "data-testid prefix"),
  )}"]`;
}

export function dataTestIdSelector(testId: string): string {
  return `[data-testid="${cssAttributeValue(
    requireNonEmptySelectorValue(testId, "data-testid"),
  )}"]`;
}

function requireViewSchemaId(value: string): string {
  const token = requireNonEmptySelectorValue(value, "view_schema_id");
  if (
    !/^cartulary\.view\.[a-z][a-z0-9_]*(?:\.[a-z0-9_]+)*\.v[1-9][0-9]*$/u.test(
      token,
    )
  ) {
    throw new Error(`Invalid view_schema_id selector token: ${value}`);
  }
  if (!registeredViewSchemaIds.has(token)) {
    throw new Error(`Unknown view_schema_id selector token: ${value}`);
  }
  const encoded = encodeSelectorSegment(token, "view_schema_id");
  return encoded;
}

function requireFieldKey(value: string): string {
  const encoded = encodeSelectorSegment(value, "field_key");
  if (!/^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/u.test(value)) {
    throw new Error(`Invalid field_key selector token: ${value}`);
  }
  return encoded;
}

function requireFeatureGroupKey(value: string): string {
  const encoded = encodeSelectorSegment(value, "feature_group_key");
  if (!/^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/u.test(value)) {
    throw new Error(`Invalid feature_group_key selector token: ${value}`);
  }
  return encoded;
}

function requireRecordId(value: string): string {
  return encodeSelectorSegment(value, "record_id");
}

function requireItemRef(value: string): string {
  return encodeSelectorSegment(value, "item_ref");
}

function encodeSelectorSegment(value: string, label: string): string {
  return encodeURIComponent(requireNonEmptySelectorValue(value, label));
}

function requireNonEmptySelectorValue(value: string, label: string): string {
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(`Invalid ${label} selector token: ${value}`);
  }
  return value;
}

function requireSelectorToken<T extends string>(
  tokens: Readonly<Record<T, string>>,
  value: T,
  label: string,
): string {
  const token = tokens[value];
  if (token === undefined) {
    throw new Error(`Invalid ${label} token: ${String(value)}`);
  }
  return token;
}

function stableSelectorTokenTestId<T extends string>(
  tokens: Readonly<Record<T, string>>,
  value: T,
  label: string,
): StableTestId {
  return stableTestId(requireSelectorToken(tokens, value, label));
}

function semanticSelectorTestId<T extends string>(
  tokens: Readonly<Record<T, string>>,
  value: T,
  label: string,
): StableTestId {
  return stableSelectorTokenTestId(tokens, value, label);
}

function stableTestId(value: string): StableTestId {
  return value as StableTestId;
}

function encodedTestId(prefix: string, value: string, label: string): string {
  return `${prefix}-${encodeSelectorSegment(value, label)}`;
}

function userScopedTestId(prefix: string, userId: string): string {
  return encodedTestId(prefix, userId, "user_id");
}

function requireLandingAdminPanel(
  panel: LandingAdminPanelToken,
): LandingAdminPanelToken {
  if (landingAdminPanelTokenSet.has(panel)) {
    return panel;
  }
  throw new Error(`Invalid landing admin panel token: ${String(panel)}`);
}

function requireIncidentControlsSection(
  section: IncidentControlsSection,
): IncidentControlsSection {
  return requireClosedToken(
    incidentControlsSections,
    section,
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

function stableEncodedTestId(
  prefix: string,
  value: string,
  label: string,
): StableTestId {
  return stableTestId(encodedTestId(prefix, value, label));
}

function viewScopedTestId(prefix: string, viewSchemaId: string): string {
  return `${prefix}-${requireViewSchemaId(viewSchemaId)}`;
}

function genericEditControlTestId(
  control: string,
  viewSchemaId: string,
): string {
  return viewScopedTestId(`generic-edit-${control}`, viewSchemaId);
}

function viewFirstTestId(
  viewSchemaId: WorkbookSurface,
  suffix: string,
): string {
  return `${requireViewSchemaId(viewSchemaId)}-${suffix}`;
}

function suffixedTestId(base: string, suffix?: string): string {
  return suffix === undefined ? base : `${base}-${suffix}`;
}

function tokenScopedTestId(
  prefix: string,
  token: string,
  suffix?: string,
): string {
  return suffixedTestId(`${prefix}-${token}`, suffix);
}

function recordTestId(
  prefix: string,
  recordId: string,
  suffix?: string,
): string {
  return tokenScopedTestId(prefix, requireRecordId(recordId), suffix);
}

function evidenceRecordControlTestId(
  control: string,
  recordId: string,
): string {
  return recordTestId(`evidence-${control}`, recordId);
}

function recordFieldTestId(
  prefix: string,
  recordId: string,
  fieldKey: string,
  suffix?: string,
): string {
  return tokenScopedTestId(
    recordTestId(prefix, recordId),
    requireFieldKey(fieldKey),
    suffix,
  );
}

function draftFieldTestId(fieldKey: string, suffix?: string): string {
  return tokenScopedTestId("draft-row", requireFieldKey(fieldKey), suffix);
}

function timelineCollectionFieldControlTestId(
  recordId: string | null,
  fieldKey: string,
  suffix: "items" | "input" | "overflow",
): string {
  return recordId === null
    ? draftFieldTestId(fieldKey, suffix)
    : recordFieldTestId("row", recordId, fieldKey, suffix);
}

function itemRefTestId(
  prefix: string,
  itemRef: string,
  suffix?: string,
): string {
  return tokenScopedTestId(prefix, requireItemRef(itemRef), suffix);
}

function requireClosedToken<T extends string>(
  tokens: readonly T[],
  value: T,
  label: string,
): T {
  if ((tokens as readonly string[]).includes(value)) {
    return value;
  }
  throw new Error(`Invalid ${label} token: ${String(value)}`);
}

function requireSystemViewSwitcherGroupToken(
  groupToken: SystemViewSwitcherGroupToken,
): SystemViewSwitcherGroupToken {
  return requireClosedToken(
    systemViewSwitcherGroupTokens,
    groupToken,
    "system view switcher group",
  );
}

function requireWorkbookShellSlot(slot: WorkbookShellSlot): WorkbookShellSlot {
  return requireClosedToken(workbookShellSlots, slot, "workbook shell slot");
}

function requireTimelineInspectorSection(
  section: TimelineInspectorSection,
): TimelineInspectorSection {
  return requireClosedToken(
    timelineInspectorSections,
    section,
    "timeline inspector section",
  );
}

function requireWorkbookInspectorPanelId(
  panelId: WorkbookInspectorPanelId,
): WorkbookInspectorPanelId {
  return requireClosedToken(
    workbookInspectorPanelIds,
    panelId,
    "workbook inspector panel",
  );
}

function requirePublicErrorSurface(
  value: PublicErrorSurface,
): PublicErrorSurface {
  if (publicErrorSurfaces.has(value)) {
    return value;
  }
  throw new Error(`Invalid public error surface selector token: ${value}`);
}

function cssAttributeValue(value: string): string {
  return value
    .replace(/\\/gu, "\\\\")
    .replace(/\n/gu, "\\a ")
    .replace(/\r/gu, "\\d ")
    .replace(/\f/gu, "\\c ")
    .replace(/"/gu, '\\"');
}

function requireEntityType(value: EntityType): EntityType {
  if (value === "host" || value === "identity") {
    return value;
  }
  throw new Error(`Invalid entity type selector token: ${value}`);
}

function requireEntityMentionResolutionStatus(
  value: string,
): EntityMentionResolutionStatus {
  if (value === "unresolved" || value === "resolved" || value === "dismissed") {
    return value;
  }
  throw new Error(`Invalid entity_mentions.resolution_status token: ${value}`);
}

function requireRowHistoryRollbackAction(
  value: string,
): RowHistoryRollbackAction {
  if (
    value === "change_set" ||
    value === "history_entry" ||
    value === "row_restore"
  ) {
    return value;
  }
  throw new Error(`Invalid row history rollback action token: ${value}`);
}

function requireRowHistoryDestructiveOperation(
  value: string,
): RowHistoryDestructiveOperation {
  if (value === "delete" || value === "restore") {
    return value;
  }
  throw new Error(`Invalid row history destructive operation token: ${value}`);
}

function rowHistoryItemIdentity(anchor: RowHistoryItemAnchor): string {
  return requireNonEmptySelectorValue(
    anchor.historyItemRef,
    "history_item_ref",
  );
}

function rowHistoryActionIdentity(anchor: RowHistoryActionAnchor): string {
  return requireNonEmptySelectorValue(
    anchor.historyItemRef,
    "history_item_ref",
  );
}
