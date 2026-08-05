import { describe, expect, it } from "vitest";

import {
  accountTestId,
  appRouteTestId,
  authTestId,
  currentIncidentRoleTestId,
  debugIncidentRowTestId,
  debugMembershipDeleteButtonTestId,
  debugMembershipPatchButtonTestId,
  debugMembershipRoleInputTestId,
  debugMembershipRowTestId,
  debugMembershipVersionTestId,
  debugSelectIncidentButtonTestId,
  deploymentAdminTestId,
  deploymentUserRowTestId,
  extensionProfileRowTestId,
  incidentAdministrationTestId,
  incidentControlsActionMessageTestId,
  incidentControlsCloseButtonTestId,
  incidentControlsMenuItemTestId,
  incidentControlsMenuTestId,
  incidentControlsPanelTestId,
  incidentControlsStatusTestId,
  incidentControlsSurfaceTestId,
  incidentControlsTriggerTestId,
  incidentLandingTestId,
  incidentMembershipAdminNoteTestId,
  incidentMembershipAuditRowTestId,
  incidentMembershipCreateButtonTestId,
  incidentMembershipDeleteButtonTestId,
  incidentMembershipEmailInputTestId,
  incidentMembershipListTestId,
  incidentMembershipPatchButtonTestId,
  incidentMembershipRoleDisplayTestId,
  incidentMembershipRoleInputTestId,
  incidentMembershipRoleSelectTestId,
  incidentMembershipRowTestId,
  incidentMembershipVersionTestId,
  landingAdminMenuItemTestId,
  landingAdminPanelTestId,
  landingAdminShellTestId,
  landingIncidentCardTestId,
  landingIncidentOpenButtonTestId,
  publicErrorCodeTestId,
  publicErrorSummaryTestIds,
  timelineInspectorMessageTestId,
  timelineInspectorSectionTestId,
  timelineInspectorTestId,
  timelineMutationSubstrateReadyTestId,
  workbookFilterPopoverTestId,
  workbookFilterPopoverTriggerTestId,
  workbookImportAssistantTestId,
  workbookIncidentIdentityTestId,
  workbookResponsiveBandTestId,
  workbookShellReadyTestId,
  workbookShellSlotLabel,
  workbookShellSlots,
  workbookShellSlotTestId,
  workbookSortMenuTestId,
  workbookSortMenuTriggerTestId,
  workbookSortOptionTestId,
  workbookSurfacesMenuOptionTestId,
  workbookSurfacesMenuTestId,
  workbookSurfacesMenuTriggerTestId,
  workbookViewBarQueryControlsTestId,
} from "./index";

const incidentControlsSections = [
  "summary",
  "import-assistant",
  "incident-fields",
  "memberships",
  "membership-audit",
] as const;

const timelineInspectorSections = [
  "operational-text",
  "relationships",
  "evidence",
  "history",
] as const;

describe("@cartulary/ui-contracts application selectors", () => {
  it("provides shared builders for app shell and incident membership selectors", () => {
    expect(workbookShellReadyTestId()).toBe("workbook-shell-ready");
    expect(workbookIncidentIdentityTestId()).toBe("workbook-incident-identity");
    expect(workbookResponsiveBandTestId()).toBe("workbook-responsive-band");
    expect(workbookSurfacesMenuTriggerTestId()).toBe(
      "workbook-surfaces-menu-trigger",
    );
    expect(workbookSurfacesMenuTestId()).toBe("workbook-surfaces-menu");
    expect(workbookSurfacesMenuOptionTestId("cartulary.view.timeline.v2")).toBe(
      "workbook-surfaces-menu-option-cartulary.view.timeline.v2",
    );
    expect(
      workbookViewBarQueryControlsTestId("cartulary.view.timeline.v2"),
    ).toBe("cartulary.view.timeline.v2-view-bar-query");
    expect(workbookSortMenuTriggerTestId("cartulary.view.timeline.v2")).toBe(
      "cartulary.view.timeline.v2-sort-menu-trigger",
    );
    expect(workbookSortMenuTestId("cartulary.view.timeline.v2")).toBe(
      "cartulary.view.timeline.v2-sort-menu",
    );
    expect(
      workbookSortOptionTestId(
        "cartulary.view.timeline.v2",
        "timeline.activity_synopsis_text",
      ),
    ).toBe(
      "cartulary.view.timeline.v2-sort-option-timeline.activity_synopsis_text",
    );
    expect(
      workbookFilterPopoverTriggerTestId("cartulary.view.timeline.v2"),
    ).toBe("cartulary.view.timeline.v2-filter-popover-trigger");
    expect(workbookFilterPopoverTestId("cartulary.view.timeline.v2")).toBe(
      "cartulary.view.timeline.v2-filter-popover",
    );
    expect(workbookShellSlots).toEqual([
      "top-bar",
      "view-bar",
      "primary-grid",
      "inspector",
      "status-strip",
    ]);
    expect(workbookShellSlotTestId("top-bar")).toBe(
      "workbook-shell-slot-top-bar",
    );
    expect(workbookShellSlotTestId("view-bar")).toBe(
      "workbook-shell-slot-view-bar",
    );
    expect(workbookShellSlotLabel("top-bar")).toBe("Workbook top bar");
    expect(workbookShellSlotLabel("primary-grid")).toBe("Primary grid");
    expect(() => workbookShellSlotTestId("toolbar" as never)).toThrow(
      "Invalid workbook shell slot token: toolbar",
    );
    expect(() => workbookShellSlotLabel("toolbar" as never)).toThrow(
      "Invalid workbook shell slot token: toolbar",
    );
    expect(timelineMutationSubstrateReadyTestId()).toBe(
      "timeline-mutation-substrate-ready",
    );
    expect(currentIncidentRoleTestId()).toBe("current-incident-role");
    expect(incidentControlsTriggerTestId()).toBe("incident-controls-trigger");
    expect(incidentControlsMenuTestId()).toBe("incident-controls-menu");
    expect(incidentControlsSections).toEqual([
      "summary",
      "import-assistant",
      "incident-fields",
      "memberships",
      "membership-audit",
    ]);
    expect(incidentControlsMenuItemTestId("summary")).toBe(
      "incident-controls-menu-item-summary",
    );
    expect(incidentControlsMenuItemTestId("import-assistant")).toBe(
      "incident-controls-menu-item-import-assistant",
    );
    expect(workbookImportAssistantTestId()).toBe("workbook-import-assistant");
    expect(incidentControlsMenuItemTestId("incident-fields")).toBe(
      "incident-controls-menu-item-incident-fields",
    );
    expect(incidentControlsMenuItemTestId("memberships")).toBe(
      "incident-controls-menu-item-memberships",
    );
    expect(incidentControlsMenuItemTestId("membership-audit")).toBe(
      "incident-controls-menu-item-membership-audit",
    );
    expect(() => incidentControlsMenuItemTestId("audit" as never)).toThrow(
      "Invalid incident controls section token: audit",
    );
    expect(incidentControlsPanelTestId()).toBe("incident-controls-panel");
    expect(incidentControlsSurfaceTestId()).toBe("incident-controls-surface");
    expect(incidentControlsStatusTestId()).toBe("incident-admin-status");
    expect(incidentControlsActionMessageTestId()).toBe(
      "incident-admin-action-message",
    );
    expect(
      (
        [
          "admin-action-message",
          "admin-error-code",
          "admin-status",
          "close-button",
          "lifecycle-reason",
          "patch-button",
          "patch-current-phase",
          "patch-description",
          "patch-external-case",
          "patch-readonly-note",
          "patch-severity",
          "patch-tlp",
          "pref-default-sheet-ref",
          "pref-home-sheet-ref",
          "reopen-button",
          "summary-closed-at",
          "summary-current-phase",
          "summary-description",
          "summary-key",
          "summary-primary-external-case-ref",
          "summary-role",
          "summary-severity",
          "summary-status",
          "summary-title",
          "summary-tlp",
          "summary-version",
        ] as const
      ).map((selector) => incidentAdministrationTestId(selector)),
    ).toEqual([
      "incident-admin-action-message",
      "incident-admin-error-code",
      "incident-admin-status",
      "incident-close-button",
      "incident-lifecycle-reason",
      "incident-patch-button",
      "incident-patch-current-phase",
      "incident-patch-description",
      "incident-patch-external-case",
      "incident-patch-readonly-note",
      "incident-patch-severity",
      "incident-patch-tlp",
      "incident-pref-default-sheet-ref",
      "incident-pref-home-sheet-ref",
      "incident-reopen-button",
      "incident-summary-closed-at",
      "incident-summary-current-phase",
      "incident-summary-description",
      "incident-summary-key",
      "incident-summary-primary-external-case-ref",
      "incident-summary-role",
      "incident-summary-severity",
      "incident-summary-status",
      "incident-summary-title",
      "incident-summary-tlp",
      "incident-summary-version",
    ]);
    expect(() =>
      incidentAdministrationTestId("delete-incident" as never),
    ).toThrow(
      "Invalid incident administration selector token: delete-incident",
    );
    expect(
      incidentMembershipAuditRowTestId("00000000-0000-4000-8000-000000002001"),
    ).toBe("membership-audit-row-00000000-0000-4000-8000-000000002001");
    expect(incidentControlsCloseButtonTestId()).toBe("incident-controls-close");
    expect(timelineInspectorSections).toEqual([
      "operational-text",
      "relationships",
      "evidence",
      "history",
    ]);
    expect(timelineInspectorSectionTestId("operational-text")).toBe(
      "timeline-inspector-section-operational-text",
    );
    expect(timelineInspectorSectionTestId("relationships")).toBe(
      "timeline-inspector-section-relationships",
    );
    expect(timelineInspectorSectionTestId("evidence")).toBe(
      "timeline-inspector-section-evidence",
    );
    expect(timelineInspectorSectionTestId("history")).toBe(
      "timeline-inspector-section-history",
    );
    expect(timelineInspectorMessageTestId()).toBe("timeline-inspector-message");
    expect(timelineInspectorTestId()).toBe("timeline-inspector");
    expect(timelineInspectorSectionTestId("operational-text")).toBe(
      timelineInspectorSectionTestId("operational-text"),
    );
    expect(() => timelineInspectorSectionTestId("summary" as never)).toThrow(
      "Invalid timeline inspector section token: summary",
    );
    expect(landingIncidentCardTestId("incident-1")).toBe(
      "landing-incident-incident-1",
    );
    expect(landingIncidentOpenButtonTestId("incident-1")).toBe(
      "landing-open-incident-1",
    );
    expect(landingIncidentCardTestId("incident:1")).toBe(
      "landing-incident-incident%3A1",
    );
    expect(incidentMembershipEmailInputTestId()).toBe(
      "incident-membership-email",
    );
    expect(incidentMembershipRoleSelectTestId()).toBe(
      "incident-membership-role",
    );
    expect(incidentMembershipCreateButtonTestId()).toBe(
      "incident-membership-create",
    );
    expect(incidentMembershipAdminNoteTestId()).toBe(
      "incident-membership-admin-note",
    );
    expect(incidentMembershipListTestId()).toBe("incident-membership-list");
    expectSelectorCases<
      "delete" | "patch" | "roleDisplay" | "roleInput" | "row" | "version"
    >(
      (control) => {
        const testIdFor = {
          row: incidentMembershipRowTestId,
          version: incidentMembershipVersionTestId,
          roleInput: incidentMembershipRoleInputTestId,
          patch: incidentMembershipPatchButtonTestId,
          delete: incidentMembershipDeleteButtonTestId,
          roleDisplay: incidentMembershipRoleDisplayTestId,
        }[control];
        return testIdFor("user-2");
      },
      [
        ["row", "incident-membership-row-user-2"],
        ["version", "incident-membership-version-user-2"],
        ["roleInput", "incident-membership-role-input-user-2"],
        ["patch", "incident-membership-patch-user-2"],
        ["delete", "incident-membership-delete-user-2"],
        ["roleDisplay", "incident-membership-role-user-2"],
      ],
    );
    expect(debugIncidentRowTestId("incident-2")).toBe(
      "incident-row-incident-2",
    );
    expect(debugSelectIncidentButtonTestId("incident-2")).toBe(
      "select-incident-incident-2",
    );
    expectSelectorCases<"delete" | "patch" | "roleInput" | "row" | "version">(
      (control) => {
        const testIdFor = {
          row: debugMembershipRowTestId,
          roleInput: debugMembershipRoleInputTestId,
          version: debugMembershipVersionTestId,
          patch: debugMembershipPatchButtonTestId,
          delete: debugMembershipDeleteButtonTestId,
        }[control];
        return testIdFor("user-2");
      },
      [
        ["row", "membership-row-user-2"],
        ["roleInput", "membership-role-input-user-2"],
        ["version", "membership-version-user-2"],
        ["patch", "patch-membership-user-2"],
        ["delete", "delete-membership-user-2"],
      ],
    );
    expect(extensionProfileRowTestId("profile:core")).toBe(
      "extension-profile%3Acore",
    );
  });

  it("provides stable authentication bootstrap, landing, session, admin, and error selectors", () => {
    expectSelectorCases(authTestId, [
      ["shell", "auth-shell"],
      ["shell-message", "auth-shell-message"],
      ["status", "auth-status"],
      ["login-username", "auth-login-username"],
      ["login-password", "auth-login-password"],
      ["login-totp-code", "auth-login-totp-code"],
      ["login-submit", "auth-login-submit"],
      ["bootstrap-token", "auth-bootstrap-token"],
      ["bootstrap-enrollment-id", "auth-bootstrap-enrollment-id"],
      ["bootstrap-secret-base32", "auth-bootstrap-secret-base32"],
      ["bootstrap-begin", "auth-bootstrap-begin"],
      ["bootstrap-complete-code", "auth-bootstrap-complete-code"],
      ["bootstrap-complete", "auth-bootstrap-complete"],
    ]);

    expectSelectorCases(incidentLandingTestId, [
      ["shell", "incident-landing"],
      ["current-user", "landing-current-user"],
      ["refresh", "landing-refresh"],
      ["incident-key", "landing-incident-key"],
      ["incident-title", "landing-incident-title"],
      ["create-open-button", "landing-create-open-button"],
      ["create-submit-button", "landing-create-submit-button"],
      ["incidents-count", "landing-incidents-count"],
      ["loading", "landing-loading"],
      ["empty-state", "landing-empty-state"],
      ["incident-list", "landing-incident-list"],
      ["status", "landing-status"],
      ["return", "landing-return"],
    ]);

    expectSelectorCases(landingAdminShellTestId, [
      ["shell", "landing-admin-shell"],
      ["menu", "landing-admin-menu"],
      ["status-strip", "landing-admin-status-strip"],
    ]);
    expect(landingAdminMenuItemTestId("administrative-audit")).toBe(
      "landing-admin-menu-item-administrative-audit",
    );
    expect(landingAdminMenuItemTestId("deployment-users")).toBe(
      "landing-admin-menu-item-deployment-users",
    );
    expect(landingAdminMenuItemTestId("incident-import")).toBe(
      "landing-admin-menu-item-incident-import",
    );
    expect(landingAdminMenuItemTestId("reference-packs")).toBe(
      "landing-admin-menu-item-reference-packs",
    );
    expect(landingAdminPanelTestId("administrative-audit")).toBe(
      "landing-admin-panel-administrative-audit",
    );
    expect(landingAdminPanelTestId("deployment-users")).toBe(
      "landing-admin-panel-deployment-users",
    );
    expect(landingAdminPanelTestId("incident-import")).toBe(
      "landing-admin-panel-incident-import",
    );
    expect(landingAdminPanelTestId("reference-packs")).toBe(
      "landing-admin-panel-reference-packs",
    );
    expect(deploymentUserRowTestId("user:1")).toBe(
      "deployment-user-row-user%3A1",
    );

    expectSelectorCases(appRouteTestId, [
      ["app-shell", "app-shell"],
      ["workbook-current-user", "workbook-current-user"],
      ["workbook-loading", "workbook-loading"],
      ["debug-harness-loading", "debug-harness-loading"],
      ["debug-harness-shell", "debug-harness-shell"],
    ]);

    expectSelectorCases(accountTestId, [
      ["refresh-state", "account-refresh-state"],
      ["logout", "account-logout"],
      ["session-user-id", "account-session-user-id"],
      ["session-provider-type", "account-session-provider-type"],
      ["session-mfa-state", "account-session-mfa-state"],
      ["session-is-deployment-admin", "account-session-is-deployment-admin"],
      ["session-authenticated-at", "account-session-authenticated-at"],
      ["session-idle-expires-at", "account-session-idle-expires-at"],
      ["session-absolute-expires-at", "account-session-absolute-expires-at"],
      ["session-session-expires-at", "account-session-session-expires-at"],
      ["session-memberships", "account-session-memberships"],
      ["credential-auth-kind", "account-credential-auth-kind"],
      ["credential-recovery-model", "account-credential-recovery-model"],
      [
        "credential-password-changed-at",
        "account-credential-password-changed-at",
      ],
      ["credential-totp-state", "account-credential-totp-state"],
      [
        "credential-pending-expires-at",
        "account-credential-pending-expires-at",
      ],
      ["password-current", "account-password-current"],
      ["password-next", "account-password-next"],
      ["password-factor-code", "account-password-factor-code"],
      ["password-change", "account-password-change"],
      ["totp-current-password", "account-totp-current-password"],
      ["totp-current-factor", "account-totp-current-factor"],
      ["totp-begin", "account-totp-begin"],
      ["totp-enrollment-id", "account-totp-enrollment-id"],
      ["totp-secret-base32", "account-totp-secret-base32"],
      ["totp-complete-code", "account-totp-complete-code"],
      ["totp-complete", "account-totp-complete"],
      ["status", "account-status"],
    ]);

    expectSelectorCases(deploymentAdminTestId, [
      ["access-note", "admin-access-note"],
      ["create-email", "admin-create-email"],
      ["create-display-name", "admin-create-display-name"],
      ["create-password", "admin-create-password"],
      ["create-mfa-required", "admin-create-mfa-required"],
      ["create-is-deployment-admin", "admin-create-is-deployment-admin"],
      ["create-user", "admin-create-user"],
      ["user-filter", "admin-user-filter"],
      ["user-list", "admin-user-list"],
      ["load-more-users", "admin-load-more-users"],
      ["target-user-id-input", "admin-target-user-id-input"],
      ["load-user", "admin-load-user"],
      ["target-user-id", "admin-target-user-id"],
      ["target-user-version", "admin-target-user-version"],
      ["target-is-active", "admin-target-is-active"],
      ["target-is-deployment-admin", "admin-target-is-deployment-admin"],
      ["patch-base-version", "admin-patch-base-version"],
      ["patch-display-name", "admin-patch-display-name"],
      ["patch-mfa-required", "admin-patch-mfa-required"],
      ["patch-is-active", "admin-patch-is-active"],
      ["patch-is-deployment-admin", "admin-patch-is-deployment-admin"],
      ["patch-user", "admin-patch-user"],
      ["new-password", "admin-new-password"],
      ["reason", "admin-reason"],
      ["password-reset", "admin-password-reset"],
      ["totp-reset", "admin-totp-reset"],
      ["revoke-all", "admin-revoke-all"],
      ["status", "admin-status"],
    ]);

    expect(publicErrorCodeTestId("auth")).toBe("auth-error-code");
    expect(publicErrorSummaryTestIds("auth")).toEqual({
      container: "auth-error-public",
      details: "auth-error-details",
      message: "auth-error-message",
    });
    expect(publicErrorCodeTestId("account")).toBe("account-error-code");
    expect(publicErrorSummaryTestIds("account")).toEqual({
      container: "account-error-public",
      details: "account-error-details",
      message: "account-error-message",
    });
    expect(publicErrorCodeTestId("admin")).toBe("admin-error-code");
    expect(publicErrorSummaryTestIds("admin")).toEqual({
      container: "admin-error-public",
      details: "admin-error-details",
      message: "admin-error-message",
    });
    expect(publicErrorCodeTestId("landing")).toBe("landing-error-code");
    expect(publicErrorSummaryTestIds("landing")).toEqual({
      container: "landing-error-public",
      details: "landing-error-details",
      message: "landing-error-message",
    });
  });

  it("keeps authentication selector identity on semantic state and stable field identifiers", () => {
    const renamedSession = {
      displayLabel: "Current operator",
      field: "session-user-id" as const,
    };
    const relabeledSession = {
      displayLabel: "Signed-in user",
      field: "session-user-id" as const,
    };
    const authControls = [
      { field: "login-password" as const, label: "Password" },
      { field: "login-username" as const, label: "Email address" },
    ];

    expect(accountTestId(renamedSession.field)).toBe(
      accountTestId(relabeledSession.field),
    );
    expect(
      authTestId(
        authControls.find((control) => control.label === "Email address")
          ?.field ?? "login-password",
      ),
    ).toBe("auth-login-username");
    expect(
      authTestId(
        [...authControls]
          .reverse()
          .find((control) => control.field === "login-username")?.field ??
          "login-password",
      ),
    ).toBe("auth-login-username");
    expect(landingIncidentCardTestId("incident:stable")).toBe(
      "landing-incident-incident%3Astable",
    );
  });

  it("rejects invalid authentication selector vocabularies", () => {
    expect(() => authTestId("username" as never)).toThrow(
      "Invalid auth selector token: username",
    );
    expect(() => accountTestId("user-id" as never)).toThrow(
      "Invalid account selector token: user-id",
    );
    expect(() => deploymentAdminTestId("target-user" as never)).toThrow(
      "Invalid deployment admin selector token: target-user",
    );
    expect(() => incidentLandingTestId("incident-card" as never)).toThrow(
      "Invalid incident landing selector token: incident-card",
    );
    expect(() => landingAdminShellTestId("tabs" as never)).toThrow(
      "Invalid landing admin shell selector token: tabs",
    );
    expect(() => landingAdminMenuItemTestId("users" as never)).toThrow(
      "Invalid landing admin panel token: users",
    );
    expect(() => appRouteTestId("shell" as never)).toThrow(
      "Invalid app route selector token: shell",
    );
    expect(() => publicErrorCodeTestId("session" as never)).toThrow(
      "Invalid public error surface selector token: session",
    );
  });
});

function expectSelectorCases<T extends string>(
  testIdFor: (selector: T) => string,
  cases: ReadonlyArray<readonly [T, string]>,
): void {
  for (const [selector, expected] of cases) {
    expect(testIdFor(selector)).toBe(expected);
  }
}
