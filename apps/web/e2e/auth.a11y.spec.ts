import {
  accountTestId,
  appRouteTestId,
  authTestId,
  incidentAdministrationTestId,
  incidentLandingTestId,
  landingAdminShellTestId,
  landingIncidentCardTestId,
  publicErrorCodeTestId,
  publicErrorSummaryTestIds,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import type { APIRequestContext, Page, Route } from "@playwright/test";
import { expect, test } from "./fixtures";
import { AccountSettings } from "./pages/accountSettings";
import { AuthGateway } from "./pages/authGateway";
import { openIncidentControls } from "./pages/deploymentAdministration";
import { IncidentDirectory } from "./pages/incidentDirectory";
import { expectAllInteractiveControlsNamed } from "./support/accessibility/assertions";
import {
  expectNoFocusTrap,
  expectTabOrderIncludes,
  expectVisibleFocus,
} from "./support/accessibility/keyboard";
import { expectAndRecordContrast } from "./support/accessibility/records";
import {
  expectAlertRole,
  expectCurrentIncidentRole,
  expectNoPrivateDiagnostics,
  expectStatusRole,
} from "./support/accessibility/status";
import { createDeploymentUser } from "./support/auth/deploymentUsers";
import { revokeAllSessions } from "./support/auth/sessions";
import { sessionCookieName } from "./support/auth/storageState";
import {
  enrollTotpViaBootstrap,
  generateTotpCode,
} from "./support/auth/suiteAdmin";
import { createIncident } from "./support/incidents/fixtures";
import { createIncidentMembership } from "./support/incidents/memberships";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import {
  holdBrowserRequest,
  installOwnedRoute,
  removeOwnedRoute,
} from "./support/transport/requestInterception";

type IncidentMembershipRecord = {
  membership_version: number;
  role: string;
  user_id: string;
};

async function fulfillPublicError(
  route: Route,
  options: {
    code: string;
    details?: Record<string, unknown>;
    message?: string;
    status: number;
  },
) {
  await route.fulfill({
    contentType: "application/json",
    status: options.status,
    body: JSON.stringify({
      error: {
        code: options.code,
        message: options.message,
        status: options.status,
        request_id: "raw-request-id-must-not-render",
        details: {
          reason_code: "safe_reason",
          ...(options.details ?? {}),
          bootstrap_token: "private-bootstrap-token",
          internal_path: "/home/cartulary/private/file.go",
          raw_request_id: "raw-request-id-must-not-render",
          secret_base32: "PRIVATESECRETBASE32",
          sql: "select * from private_table",
          stack: "stack frame at /home/cartulary/private/file.go:1",
        },
      },
    }),
  });
}

async function patchIncidentMembership(
  authRequests: APIRequestContext,
  incidentId: string,
  options: {
    baseMembershipVersion: number;
    role: string;
    userId: string;
  },
  headers: Record<string, string> = {},
) {
  const response = await authRequests.patch(
    `${apiBase}/api/v1/incidents/${incidentId}/memberships/${options.userId}`,
    {
      headers,
      data: {
        base_membership_version: options.baseMembershipVersion,
        role: options.role,
      },
    },
  );
  expect(response.ok()).toBeTruthy();
  return ((await response.json()) as { data: IncidentMembershipRecord }).data;
}

async function deleteIncidentMembership(
  authRequests: APIRequestContext,
  incidentId: string,
  userId: string,
  baseMembershipVersion: number,
  headers: Record<string, string> = {},
) {
  const response = await authRequests.delete(
    `${apiBase}/api/v1/incidents/${incidentId}/memberships/${userId}`,
    {
      headers,
      data: {
        base_membership_version: baseMembershipVersion,
      },
    },
  );
  expect(response.status()).toBe(204);
}

async function loadIncidentMembership(
  authRequests: APIRequestContext,
  incidentId: string,
  userId: string,
) {
  const response = await authRequests.get(
    `${apiBase}/api/v1/incidents/${incidentId}/memberships`,
  );
  expect(response.ok()).toBeTruthy();
  const body = (await response.json()) as {
    data: { memberships: IncidentMembershipRecord[] };
  };
  const membership =
    body.data.memberships.find((candidate) => candidate.user_id === userId) ??
    null;
  if (membership === null) {
    throw new Error(`missing incident membership for ${userId}`);
  }
  return membership;
}

async function createIncidentMembershipWithRequest(
  authRequests: APIRequestContext,
  incidentId: string,
  email: string,
  role: string,
) {
  const response = await authRequests.post(
    `${apiBase}/api/v1/incidents/${incidentId}/memberships`,
    {
      data: {
        client_txn_id: uniqueTxn("a11y-membership"),
        email,
        role,
      },
    },
  );
  expect(response.ok()).toBeTruthy();
}

async function createIncidentWithRequest(
  authRequests: APIRequestContext,
  incidentKey: string,
  title: string,
) {
  const response = await authRequests.post(`${apiBase}/api/v1/incidents`, {
    data: {
      client_txn_id: uniqueTxn("a11y-incident"),
      incident_key: incidentKey,
      title,
    },
  });
  expect(response.ok()).toBeTruthy();
  const body = (await response.json()) as { data: { incident_id: string } };
  return body.data.incident_id;
}

async function hasSessionCookie(page: Page) {
  const cookies = await page.context().cookies(apiBase);
  return cookies.some((cookie) => cookie.name === sessionCookieName);
}

async function expectIncidentSelectionSurfaceA11y(
  page: Page,
  options: {
    focusTestId?: string;
    tabStops?: readonly string[];
  } = {},
) {
  await expectAllInteractiveControlsNamed(page);
  await expectNoFocusTrap(page);
  if (options.focusTestId) {
    await expectVisibleFocus(page.getByTestId(options.focusTestId));
  }
  if (options.tabStops && options.tabStops.length > 0) {
    await expectTabOrderIncludes(page, options.tabStops);
  }
  await expectAndRecordContrast(page, [
    ...(options.focusTestId ? [options.focusTestId] : []),
    ...(options.tabStops ?? []),
  ]);
}

async function holdSinglePublicAPIResponse(
  page: Page,
  options: {
    method: string;
    path: string;
  },
) {
  return holdBrowserRequest(page, { ...options, once: true });
}

async function clearBrowserSession(page: Page) {
  await page.context().clearCookies();
}

test("a11y.incident-selection.row-01 deferred session loading exposes progress and keeps recovery controls keyboard reachable", async ({
  page,
}) => {
  await clearBrowserSession(page);
  const heldSession = await holdSinglePublicAPIResponse(page, {
    method: "GET",
    path: "/api/v1/auth/session",
  });
  await new AuthGateway(page).goto();
  await heldSession.waitForHit;

  await expect(page.getByTestId(authTestId("shell"))).toHaveAttribute(
    "data-bootstrap-state",
    "loading",
  );
  await expect(page.getByTestId(authTestId("shell"))).toHaveAttribute(
    "aria-busy",
    "true",
  );
  await expectStatusRole(page.getByTestId(authTestId("feedback")));
  await expect(page.getByTestId(authTestId("feedback"))).toContainText(
    "Checking current session",
  );
  await expectIncidentSelectionSurfaceA11y(page, {
    focusTestId: authTestId("login-username"),
    tabStops: [authTestId("login-username"), authTestId("login-password")],
  });

  await expect(page.getByTestId(authTestId("login-submit"))).toBeDisabled();
  heldSession.release();
  await expect(page.getByTestId(authTestId("shell"))).toHaveAttribute(
    "data-bootstrap-state",
    "anonymous",
  );

  await heldSession.dispose();
});

test("a11y.incident-selection.row-01 anonymous login after initial session_required reaches login controls and authenticated landing", async ({
  page,
  sessionTracker,
  testActorRequest,
}) => {
  const email = uniqueEmail("a11y-incident-selection-login");
  const password = "A11yIncidentSelectionLoginPass!";
  const user = await createDeploymentUser(testActorRequest, {
    email,
    display_name: "A11Y Incident Selection Login",
    initial_password: password,
    mfa_required: false,
    is_deployment_admin: false,
  });

  await clearBrowserSession(page);
  await new AuthGateway(page).goto();
  await expect(page.getByTestId(authTestId("shell"))).toHaveAttribute(
    "data-bootstrap-state",
    "anonymous",
  );
  await expectIncidentSelectionSurfaceA11y(page, {
    focusTestId: authTestId("login-username"),
    tabStops: [
      authTestId("login-username"),
      authTestId("login-password"),
      authTestId("login-submit"),
    ],
  });

  await new AuthGateway(page).login(email, password);
  await expect(page.getByTestId(incidentLandingTestId("shell"))).toBeVisible();
  await expectVisibleFocus(page.getByTestId(incidentLandingTestId("refresh")));
  await expectStatusRole(page.getByTestId(incidentLandingTestId("status")));
  await expectIncidentSelectionSurfaceA11y(page, {
    focusTestId: incidentLandingTestId("refresh"),
    tabStops: [
      incidentLandingTestId("search"),
      incidentLandingTestId("status-filter"),
      incidentLandingTestId("refresh"),
      incidentLandingTestId("create-open-button"),
    ],
  });
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "authentication accessibility",
    email,
    purpose: "a11y.incident-selection.row-01 anonymous login",
    userId: user.user_id,
  });
});

test("a11y.incident-selection.row-01 mfa_required challenge is keyboard reachable, visibly focused, named, and safely announced", async ({
  page,
  sessionTracker,
  testActorRequest,
}) => {
  const email = uniqueEmail("a11y-incident-selection-mfa");
  const password = "A11yIncidentSelectionMfaPass!";
  const user = await createDeploymentUser(testActorRequest, {
    email,
    display_name: "A11Y Incident Selection MFA",
    initial_password: password,
    mfa_required: true,
    is_deployment_admin: false,
  });
  const secretBase32 = await enrollTotpViaBootstrap(page, email, password);

  await clearBrowserSession(page);
  await new AuthGateway(page).goto();
  await new AuthGateway(page).login(email, password);
  await expect(page.getByTestId(authTestId("shell"))).toHaveAttribute(
    "data-bootstrap-state",
    "mfa_required",
  );
  await expectStatusRole(page.getByTestId(authTestId("feedback")));
  await expect(page.getByTestId(authTestId("feedback"))).toContainText(
    "Authenticator code",
  );
  await expectNoPrivateDiagnostics(page.getByTestId(authTestId("feedback")));
  expect(await hasSessionCookie(page)).toBeFalsy();
  await expectIncidentSelectionSurfaceA11y(page, {
    focusTestId: authTestId("login-totp-code"),
    tabStops: [
      authTestId("login-username"),
      authTestId("login-password"),
      authTestId("login-totp-code"),
      authTestId("login-submit"),
    ],
  });

  await new AuthGateway(page).login(
    email,
    password,
    generateTotpCode(secretBase32),
  );
  await expect(
    page.getByTestId(incidentLandingTestId("current-user")),
  ).toContainText("A11Y Incident Selection MFA");
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "authentication accessibility",
    email,
    purpose: "a11y.incident-selection.row-01 mfa_required retry",
    userId: user.user_id,
  });
});

test("a11y.incident-selection.row-01 mfa_setup_required enrollment is keyboard reachable and public errors hide private setup diagnostics", async ({
  page,
  testActorRequest,
}) => {
  const email = uniqueEmail("a11y-incident-selection-mfa-setup");
  const password = "A11yIncidentSelectionSetupPass!";
  await createDeploymentUser(testActorRequest, {
    email,
    display_name: "A11Y Incident Selection MFA Setup",
    initial_password: password,
    mfa_required: true,
    is_deployment_admin: false,
  });

  await clearBrowserSession(page);
  await new AuthGateway(page).goto();
  await new AuthGateway(page).login(email, password);
  await expect(page.getByTestId(authTestId("shell"))).toHaveAttribute(
    "data-bootstrap-state",
    "mfa_setup_required",
  );
  await expect(page.getByTestId(authTestId("feedback"))).toHaveText(
    "Authenticator setup is required before sign-in.",
  );
  await expect(page.getByText("Setup token", { exact: true })).toHaveCount(0);
  await expectNoPrivateDiagnostics(page.getByTestId(authTestId("feedback")));
  await expectIncidentSelectionSurfaceA11y(page, {
    focusTestId: authTestId("bootstrap-begin"),
    tabStops: [authTestId("bootstrap-begin")],
  });

  await new AuthGateway(page).beginBootstrapEnrollment();
  await expectStatusRole(page.getByTestId(authTestId("feedback")));
  const secretBase32 = await new AuthGateway(page).requireText(
    authTestId("bootstrap-setup-key"),
  );
  await expectIncidentSelectionSurfaceA11y(page, {
    focusTestId: authTestId("bootstrap-complete-code"),
    tabStops: [
      authTestId("bootstrap-complete-code"),
      authTestId("bootstrap-complete"),
    ],
  });
  await new AuthGateway(page).completeBootstrapEnrollment(
    generateTotpCode(secretBase32),
  );
  await expect(
    page.getByText("Authenticator setup is complete. Sign in again.").first(),
  ).toBeVisible();
  expect(await hasSessionCookie(page)).toBeFalsy();
});

test("a11y.incident-selection.row-01 authenticated landing exposes account, admin, incident, retry, and visible incident controls", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("A11YLAND"),
    "A11Y authenticated landing",
  );

  await new IncidentDirectory(page).goto();
  await expect(page.getByTestId(incidentLandingTestId("shell"))).toBeVisible();
  await expect(
    page.getByTestId(landingIncidentCardTestId(incidentId)),
  ).toBeVisible();
  await expectVisibleFocus(
    page.getByTestId(incidentLandingTestId("create-open-button")),
  );
  await expectStatusRole(page.getByTestId(incidentLandingTestId("status")));
  await expectIncidentSelectionSurfaceA11y(page, {
    focusTestId: incidentLandingTestId("create-open-button"),
    tabStops: [
      incidentLandingTestId("search"),
      incidentLandingTestId("status-filter"),
      incidentLandingTestId("refresh"),
      incidentLandingTestId("create-open-button"),
    ],
  });
});

test("a11y.incident-selection.row-01 incident empty, list, selected, stale-selection, and incident-error states expose keyboard recovery", async ({
  page,
  sessionTracker,
  testActorRequest,
}) => {
  const email = uniqueEmail("a11y-incident-selection-incident");
  const password = "A11yIncidentSelectionIncidentPass!";
  const user = await createDeploymentUser(testActorRequest, {
    email,
    display_name: "A11Y Incident Selection Incident",
    initial_password: password,
    mfa_required: false,
    is_deployment_admin: false,
  });

  await clearBrowserSession(page);
  await new AuthGateway(page).goto();
  await new AuthGateway(page).login(email, password);
  await expect(
    page.getByTestId(incidentLandingTestId("empty-state")),
  ).toContainText("No incidents are visible");
  await expectStatusRole(page.getByTestId(incidentLandingTestId("status")));

  const selectedIncidentId = await createIncidentWithRequest(
    testActorRequest,
    uniqueIncidentKey("A11YEMPTYA"),
    "A11Y selected incident",
  );
  await createIncidentMembershipWithRequest(
    testActorRequest,
    selectedIncidentId,
    email,
    "admin",
  );
  const alternateIncidentId = await createIncidentWithRequest(
    testActorRequest,
    uniqueIncidentKey("A11YEMPTYB"),
    "A11Y alternate incident",
  );
  await createIncidentMembershipWithRequest(
    testActorRequest,
    alternateIncidentId,
    email,
    "admin",
  );

  await new IncidentDirectory(page).refresh();
  await expect(
    page.getByTestId(landingIncidentCardTestId(selectedIncidentId)),
  ).toBeVisible();
  await expect(
    page.getByTestId(landingIncidentCardTestId(alternateIncidentId)),
  ).toBeVisible();

  await new IncidentDirectory(page).openIncident(selectedIncidentId);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expectCurrentIncidentRole(page, "admin");
  await openIncidentControls(page, "incident-fields");
  const metadataPanel = page.getByRole("region", {
    name: "Promoted incident fields",
    exact: true,
  });
  await expectStatusRole(metadataPanel.getByRole("status"));
  await expectVisibleFocus(
    page.getByTestId(incidentAdministrationTestId("patch-button")),
  );
  await expectIncidentSelectionSurfaceA11y(page, {
    focusTestId: appRouteTestId("workbook-current-user"),
    tabStops: [appRouteTestId("workbook-current-user")],
  });

  await new IncidentDirectory(page).open();
  await new IncidentDirectory(page).openIncident(selectedIncidentId);
  const staleIncidentURL = page.url();

  const selectedMembership = await loadIncidentMembership(
    testActorRequest,
    selectedIncidentId,
    user.user_id,
  );
  await deleteIncidentMembership(
    testActorRequest,
    selectedIncidentId,
    user.user_id,
    selectedMembership.membership_version,
  );
  // Exercise explicit stale-route entry even if live revocation already
  // returned this page to the directory before the next browser action.
  await page.goto(staleIncidentURL);
  await expect(page).not.toHaveURL(
    new RegExp(`incident_id=${selectedIncidentId}`),
  );
  await expect(page).not.toHaveURL(/incident_id=/);
  await expect(page.getByTestId(landingAdminShellTestId("heading"))).toHaveText(
    "Incident directory",
  );
  await expect(
    page.getByTestId(landingAdminShellTestId("heading")),
  ).toBeFocused();
  await new IncidentDirectory(page).openIncident(alternateIncidentId);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expectCurrentIncidentRole(page, "admin");
  await openIncidentControls(page, "incident-fields");
  await expect(
    page.getByTestId(incidentAdministrationTestId("patch-tlp")),
  ).toBeVisible();
  const alternateMembership = await loadIncidentMembership(
    testActorRequest,
    alternateIncidentId,
    user.user_id,
  );
  await patchIncidentMembership(testActorRequest, alternateIncidentId, {
    baseMembershipVersion: alternateMembership.membership_version,
    role: "editor",
    userId: user.user_id,
  });
  await new IncidentDirectory(page).patchIncidentFields({
    currentPhase: "containment",
    externalCase: "CASE-A11Y",
    tlp: "TLP:AMBER",
  });
  await expect(
    page.getByTestId(incidentAdministrationTestId("patch-readonly-note")),
  ).toBeVisible();
  await expectStatusRole(metadataPanel.getByRole("status"));
  await expect(metadataPanel.getByRole("status")).toContainText(
    "Save was not sent.",
  );
  await expectNoPrivateDiagnostics(metadataPanel);
  await expect(
    metadataPanel.getByRole("button", { name: "Refresh", exact: true }),
  ).toBeFocused();
  await expectVisibleFocus(
    page.getByTestId(appRouteTestId("workbook-current-user")),
  );
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "authentication accessibility",
    email,
    purpose: "a11y.incident-selection.row-01 incident states",
    userId: user.user_id,
  });
});

test("a11y.incident-selection.row-01 forbidden access-denied public envelope is announced and exposes recovery without private diagnostics", async ({
  page,
}) => {
  const routePattern = "**/api/v1/incidents**";
  const routeHandler = async (route: Route) => {
    if (route.request().method().toUpperCase() !== "GET") {
      await route.fallback();
      return;
    }
    await fulfillPublicError(route, {
      code: "authorization_denied",
      details: {
        required_role: "incident_admin",
      },
      message: "Access denied.",
      status: 403,
    });
  };

  await installOwnedRoute(page, routePattern, routeHandler);
  await page.goto("/");
  await expect(page.getByTestId(incidentLandingTestId("shell"))).toBeVisible();
  await expect(
    page.getByTestId(incidentLandingTestId("shell")),
  ).toHaveAttribute("data-bootstrap-state", "forbidden");
  await expectStatusRole(page.getByTestId(incidentLandingTestId("status")));
  await expectAlertRole(page.getByTestId(publicErrorCodeTestId("landing")));
  await expect(page.getByTestId(publicErrorCodeTestId("landing"))).toHaveText(
    "authorization_denied",
  );
  await expectNoPrivateDiagnostics(
    page.getByTestId(publicErrorSummaryTestIds("landing").container),
  );
  await expectVisibleFocus(page.getByTestId(incidentLandingTestId("refresh")));
  await expectIncidentSelectionSurfaceA11y(page, {
    focusTestId: incidentLandingTestId("refresh"),
    tabStops: [incidentLandingTestId("refresh")],
  });

  await removeOwnedRoute(page, routePattern, routeHandler);
  await page.getByTestId(incidentLandingTestId("refresh")).click();
  await expect(page.getByTestId(publicErrorCodeTestId("landing"))).toHaveText(
    "",
  );
});

test("a11y.incident-selection.row-01 revoked session after prior authentication announces session end and supports re-authentication", async ({
  page,
  sessionTracker,
  testActorRequest,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("A11YREVOKE"),
    "A11Y revoked incident",
  );
  const email = uniqueEmail("a11y-incident-selection-revoked");
  const password = "A11yIncidentSelectionRevokedPass!";
  const user = await createDeploymentUser(testActorRequest, {
    email,
    display_name: "A11Y Incident Selection Revoked",
    initial_password: password,
    mfa_required: false,
    is_deployment_admin: false,
  });
  await createIncidentMembership(page, incidentId, email, "viewer");

  await clearBrowserSession(page);
  await new AuthGateway(page).goto();
  await new AuthGateway(page).login(email, password);
  await new IncidentDirectory(page).openIncident(incidentId);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expect(
    page.getByTestId(appRouteTestId("workbook-current-user")),
  ).toContainText("A11Y Incident Selection Revoked");
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "authentication accessibility",
    email,
    purpose: "a11y.incident-selection.row-01 revoked session before revoke-all",
    userId: user.user_id,
  });

  await revokeAllSessions(
    testActorRequest,
    user.user_id,
    "a11y.incident-selection.row-01 revoked-session",
  );
  await expect(page.getByTestId(authTestId("shell"))).toHaveAttribute(
    "data-bootstrap-state",
    "revoked",
  );
  await expect(page.getByTestId(authTestId("shell-message"))).toContainText(
    "Sign in again",
  );
  await expectIncidentSelectionSurfaceA11y(page, {
    focusTestId: authTestId("login-submit"),
    tabStops: [
      authTestId("login-username"),
      authTestId("login-password"),
      authTestId("login-submit"),
    ],
  });

  await new AuthGateway(page).login(email, password);
  await new IncidentDirectory(page).openIncident(incidentId);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expect(
    page.getByTestId(appRouteTestId("workbook-current-user")),
  ).toContainText("A11Y Incident Selection Revoked");
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "authentication accessibility",
    email,
    purpose: "a11y.incident-selection.row-01 revoked session re-auth",
    userId: user.user_id,
  });
});

test("a11y.incident-selection.row-01 generic public error envelope renders safe diagnostics and keyboard error recovery", async ({
  page,
}) => {
  const routePattern = "**/api/v1/auth/credential-state";
  const routeHandler = async (route: Route) => {
    await fulfillPublicError(route, {
      code: "credential_state_unavailable",
      details: {
        field: "credential_state",
        reason_code: "temporary_failure",
      },
      message: "select private_column from local_credentials",
      status: 500,
    });
  };

  await installOwnedRoute(page, routePattern, routeHandler);
  await page.goto("/");
  await new AccountSettings(page).openSecurity();
  await page.getByTestId(accountTestId("refresh-state")).focus();
  await expectVisibleFocus(page.getByTestId(accountTestId("refresh-state")));
  await expect(page.getByTestId(publicErrorCodeTestId("account"))).toHaveText(
    "credential_state_unavailable",
  );
  await expectAlertRole(
    page.getByTestId(publicErrorSummaryTestIds("account").message),
  );
  await expect(
    page.getByTestId(publicErrorCodeTestId("account")),
  ).not.toHaveAttribute("role", "alert");
  await expect(
    page.getByTestId(publicErrorSummaryTestIds("account").container),
  ).not.toHaveAttribute("role", "alert");
  await expect(
    page.getByRole("dialog", { name: "Account settings" }).getByRole("alert"),
  ).toHaveCount(1);
  await expect(
    page.getByTestId(publicErrorSummaryTestIds("account").message),
  ).toHaveText("Request failed.");
  await expect(
    page.getByTestId(publicErrorSummaryTestIds("account").details),
  ).toContainText("Reason: temporary_failure");
  await expect(
    page.getByTestId(publicErrorSummaryTestIds("account").details),
  ).toContainText("Field: credential_state");
  await expectNoPrivateDiagnostics(
    page.getByTestId(publicErrorSummaryTestIds("account").container),
  );
  await expectIncidentSelectionSurfaceA11y(page, {
    focusTestId: accountTestId("refresh-state"),
    tabStops: [accountTestId("refresh-state"), accountTestId("logout")],
  });

  await removeOwnedRoute(page, routePattern, routeHandler);
  await page.getByTestId(accountTestId("refresh-state")).click();
  await expect(page.getByTestId(accountTestId("status"))).toHaveText(
    "Refreshed account security.",
  );
});
