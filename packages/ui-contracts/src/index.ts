import { listViewSchemaRegistryEntries } from "@cartulary/protocol-ts";

export {
  type CartularyDefaultThemeId,
  type CartularyDesignTokenVarName,
  cartularyDefaultThemeId,
  cartularyDesignThemeCssText,
  cartularyDesignTokenVars,
} from "./generated/design-tokens";

export type WorkbookSurface = string;

declare const stableTestIdBrand: unique symbol;

export type StableTestId = string & {
  readonly [stableTestIdBrand]: "StableTestId";
};

export type EntityType = "host" | "identity";

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
  | "details"
  | "evidence"
  | "history"
  | "relationships";

export const timelineInspectorSections = [
  "details",
  "relationships",
  "evidence",
  "history",
] as const satisfies readonly TimelineInspectorSection[];

export type Phase1ErrorSurface = "account" | "admin" | "auth" | "landing";

export type Phase1AuthSelector =
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

export type Phase1AccountSelector =
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

export type Phase1AdminSelector =
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

export type Phase1LandingSelector =
  | "create-button"
  | "create-current-phase"
  | "create-description"
  | "create-external-case"
  | "create-severity"
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
  | "incident-bundle-import"
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
  "incident-bundle-import",
] as const satisfies readonly LandingAdminPanelToken[];

export type LandingAdminShellSelector = "menu" | "shell" | "status-strip";

export type IncidentControlsSection =
  | "incident-fields"
  | "memberships"
  | "summary";

export const incidentControlsSections = [
  "summary",
  "incident-fields",
  "memberships",
] as const satisfies readonly IncidentControlsSection[];

export type Phase1RouteSelector =
  | "app-shell"
  | "debug-harness-loading"
  | "workbook-current-user"
  | "workbook-loading";

export type WorkbookShellSlot =
  | "inspector"
  | "primary-grid"
  | "status-strip"
  | "top-bar"
  | "view-bar";

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
  | "scope-assessment";

export const systemViewSwitcherGroupTokens = [
  "scope-assessment",
  "coordination",
  "review-learning",
  "optional-artifact-surfaces",
] as const satisfies readonly SystemViewSwitcherGroupToken[];

export type Phase1ErrorSummaryTestIds = {
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

const phase1AuthTestIds = Object.freeze({
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
} satisfies Record<Phase1AuthSelector, string>);

const phase1AccountTestIds = Object.freeze({
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
} satisfies Record<Phase1AccountSelector, string>);

const phase1AdminTestIds = Object.freeze({
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
} satisfies Record<Phase1AdminSelector, string>);

const phase1LandingTestIds = Object.freeze({
  "create-button": "landing-create-button",
  "create-current-phase": "landing-create-current-phase",
  "create-description": "landing-create-description",
  "create-external-case": "landing-create-external-case",
  "create-severity": "landing-create-severity",
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
} satisfies Record<Phase1LandingSelector, string>);

const landingAdminShellTestIds = Object.freeze({
  menu: "landing-admin-menu",
  shell: "landing-admin-shell",
  "status-strip": "landing-admin-status-strip",
} satisfies Record<LandingAdminShellSelector, string>);

const landingAdminPanelTokenSet = Object.freeze(
  new Set<LandingAdminPanelToken>(landingAdminPanelTokens),
);

const phase1RouteTestIds = Object.freeze({
  "app-shell": "app-shell",
  "debug-harness-loading": "debug-harness-loading",
  "workbook-current-user": "workbook-current-user",
  "workbook-loading": "workbook-loading",
} satisfies Record<Phase1RouteSelector, string>);

const phase1ErrorSurfaces = Object.freeze(
  new Set<Phase1ErrorSurface>(["account", "admin", "auth", "landing"]),
);

export function phase1AuthTestId(selector: Phase1AuthSelector): StableTestId {
  return phase1SelectorTestId(
    phase1AuthTestIds,
    selector,
    "phase1 auth selector",
  );
}

export function phase1AccountTestId(
  selector: Phase1AccountSelector,
): StableTestId {
  return phase1SelectorTestId(
    phase1AccountTestIds,
    selector,
    "phase1 account selector",
  );
}

export function phase1AdminTestId(selector: Phase1AdminSelector): StableTestId {
  return phase1SelectorTestId(
    phase1AdminTestIds,
    selector,
    "phase1 admin selector",
  );
}

export function phase1LandingTestId(
  selector: Phase1LandingSelector,
): StableTestId {
  return phase1SelectorTestId(
    phase1LandingTestIds,
    selector,
    "phase1 landing selector",
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

export function phase1RouteTestId(selector: Phase1RouteSelector): StableTestId {
  return phase1SelectorTestId(
    phase1RouteTestIds,
    selector,
    "phase1 route selector",
  );
}

export function phase1ErrorCodeTestId(
  surface: Phase1ErrorSurface,
): StableTestId {
  return stableTestId(`${requirePhase1ErrorSurface(surface)}-error-code`);
}

export function phase1ErrorSummaryTestIds(
  surface: Phase1ErrorSurface,
): Phase1ErrorSummaryTestIds {
  const prefix = requirePhase1ErrorSurface(surface);
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

export function workbookShellReadyTestId(): string {
  return "workbook-shell-ready";
}

export function workbookShellSlotTestId(slot: WorkbookShellSlot): StableTestId {
  return stableTestId(`workbook-shell-slot-${requireWorkbookShellSlot(slot)}`);
}

export function workbookShellSlotLabel(slot: WorkbookShellSlot): string {
  return workbookShellSlotLabels[requireWorkbookShellSlot(slot)];
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

export function phase2IncidentRowTestId(incidentId: string): string {
  return encodedTestId("incident-row", incidentId, "incident_id");
}

export function phase2SelectIncidentButtonTestId(incidentId: string): string {
  return encodedTestId("select-incident", incidentId, "incident_id");
}

export function phase2MembershipRowTestId(userId: string): string {
  return phase2MembershipControlTestId("row", userId);
}

export function phase2MembershipRoleInputTestId(userId: string): string {
  return phase2MembershipControlTestId("roleInput", userId);
}

export function phase2MembershipVersionTestId(userId: string): string {
  return phase2MembershipControlTestId("version", userId);
}

export function phase2MembershipPatchButtonTestId(userId: string): string {
  return phase2MembershipControlTestId("patch", userId);
}

export function phase2MembershipDeleteButtonTestId(userId: string): string {
  return phase2MembershipControlTestId("delete", userId);
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
  return '[role="row"][data-grid-record-id=""]';
}

export function conflictMarkerTestId(
  recordId: string,
  fieldKey: string,
): string {
  return recordFieldTestId("conflict-marker", recordId, fieldKey);
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

export function statusStripQueueCountTestId(): string {
  return "status-strip-queue-count";
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
  return options.surface === "inspector" ? `${base}-inspector` : base;
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

function phase1SelectorTestId<T extends string>(
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

const phase2MembershipControlPrefixes = {
  row: "membership-row",
  roleInput: "membership-role-input",
  version: "membership-version",
  patch: "patch-membership",
  delete: "delete-membership",
} as const;

type Phase2MembershipControl = keyof typeof phase2MembershipControlPrefixes;

function phase2MembershipControlTestId(
  control: Phase2MembershipControl,
  userId: string,
): string {
  return userScopedTestId(phase2MembershipControlPrefixes[control], userId);
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

function requirePhase1ErrorSurface(
  value: Phase1ErrorSurface,
): Phase1ErrorSurface {
  if (phase1ErrorSurfaces.has(value)) {
    return value;
  }
  throw new Error(`Invalid phase1 error surface selector token: ${value}`);
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
