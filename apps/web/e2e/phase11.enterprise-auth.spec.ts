import {
  phase1AuthTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";

import { expect, test } from "./fixtures";
import {
  createIncident,
  createIncidentMemberUser,
  uniqueEmail,
  uniqueIncidentKey,
} from "./helpers";

test("E-11-02 shows enterprise providers and begins provider sign-in from the anonymous shell", async ({
  browser,
}) => {
  const context = await browser.newContext();
  const page = await context.newPage();
  let beginBody: unknown = null;

  try {
    await page.route("**/api/v1/auth/session", async (route) => {
      await route.fulfill({
        status: 401,
        contentType: "application/json",
        body: JSON.stringify({
          error: {
            code: "session_required",
            status: 401,
            details: {},
          },
        }),
      });
    });
    await page.route("**/api/v1/auth/providers", async (route) => {
      expect(route.request().method()).toBe("GET");
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          data: {
            providers: [
              {
                provider_key: "corp-oidc",
                provider_type: "oidc",
                display_name: "Corporate OIDC",
              },
            ],
          },
        }),
      });
    });
    await page.route(
      "**/api/v1/auth/providers/corp-oidc/begin",
      async (route) => {
        expect(route.request().method()).toBe("POST");
        beginBody = route.request().postDataJSON();
        await route.fulfill({
          status: 200,
          contentType: "application/json",
          body: JSON.stringify({
            data: {
              provider_key: "corp-oidc",
              provider_type: "oidc",
              redirect_url: "/enterprise-idp/start?state=e-11-02",
              expires_at: "2026-06-13T22:30:00Z",
            },
          }),
        });
      },
    );
    await page.route("**/enterprise-idp/start**", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "text/html",
        body: "<!doctype html><title>Enterprise IdP</title>",
      });
    });

    await page.goto("/");
    const providerButton = page.getByTestId(
      phase1AuthTestId("enterprise-provider-button"),
    );
    await expect(providerButton).toBeVisible();
    await expect(providerButton).toContainText("Corporate OIDC");

    await providerButton.click();
    await expect
      .poll(() => beginBody)
      .toEqual({
        return_to: "/",
      });
    await expect(page).toHaveURL(/\/enterprise-idp\/start\?state=e-11-02$/);
  } finally {
    await context.close();
  }
});

test("E-11-03 opens the only visible incident after enterprise session root convergence", async ({
  page,
  sessionTracker,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E1103"),
    "Enterprise Root Convergence",
  );
  const memberPassword = "EnterpriseRoot1!";
  const member = await createIncidentMemberUser(page, incidentId, {
    email: uniqueEmail("phase11-e1103-member"),
    display_name: "Enterprise Root Member",
    initial_password: memberPassword,
    role: "admin",
  });
  await sessionTracker.loginTrackedUser(page, {
    createdBy: "phase11.enterprise-auth",
    email: member.email,
    password: memberPassword,
    purpose: "enterprise root convergence",
    userId: member.user_id,
  });

  await page.route("**/api/v1/auth/session", async (route) => {
    const response = await route.fetch();
    const body = await response.json();
    await route.fulfill({
      response,
      contentType: "application/json",
      body: JSON.stringify({
        ...body,
        data: {
          ...body.data,
          provider_type: "oidc",
        },
      }),
    });
  });

  await page.goto("/");
  await expect(page).toHaveURL(new RegExp(`incident_id=${incidentId}`));
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
});
