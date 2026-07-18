import {
  authTestId,
  incidentLandingTestId,
  landingIncidentCardTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";

import { expect, test } from "./fixtures";
import { createIncident } from "./support/incidents/fixtures";
import {
  createIncidentMembership,
  createIncidentMemberUser,
  createLocalUser,
} from "./support/incidents/memberships";
import {
  uniqueEmail,
  uniqueIncidentKey,
} from "./support/runtime/fixtureIdentity";

test("shows enterprise providers and begins provider sign-in from the anonymous shell", async ({
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
      authTestId("enterprise-provider-button"),
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

test("handles enterprise session root landing for zero, one, multiple, and disappearing incidents", async ({
  page,
  sessionTracker,
}) => {
  const zeroPassword = "EnterpriseZero1!";
  const zeroEmail = uniqueEmail("extension_profile-e1103-zero");
  const zeroMember = await createLocalUser(page, {
    email: zeroEmail,
    display_name: "Enterprise Zero Member",
    initial_password: zeroPassword,
  });

  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E1103"),
    "Enterprise Root Convergence",
  );
  const singlePassword = "EnterpriseRoot1!";
  const singleMember = await createIncidentMemberUser(page, incidentId, {
    email: uniqueEmail("extension_profile-e1103-member"),
    display_name: "Enterprise Root Member",
    initial_password: singlePassword,
    role: "admin",
  });

  const multipleIncidentA = await createIncident(
    page,
    uniqueIncidentKey("E1103A"),
    "Enterprise Multiple A",
  );
  const multipleIncidentB = await createIncident(
    page,
    uniqueIncidentKey("E1103B"),
    "Enterprise Multiple B",
  );
  const multiplePassword = "EnterpriseMultiple1!";
  const multipleMember = await createIncidentMemberUser(
    page,
    multipleIncidentA,
    {
      email: uniqueEmail("extension_profile-e1103-multiple"),
      display_name: "Enterprise Multiple Member",
      initial_password: multiplePassword,
      role: "admin",
    },
  );
  await createIncidentMembership(
    page,
    multipleIncidentB,
    multipleMember.email,
    "viewer",
  );

  await sessionTracker.loginTrackedUser(page, {
    createdBy: "extension_profile.enterprise-auth",
    email: zeroEmail,
    password: zeroPassword,
    purpose: "enterprise root zero incidents",
    userId: zeroMember.user_id,
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
  await expect(
    page.getByTestId(incidentLandingTestId("empty-state")),
  ).toBeVisible();
  await expect(page).not.toHaveURL(/incident_id=/);

  await sessionTracker.loginTrackedUser(page, {
    createdBy: "extension_profile.enterprise-auth",
    email: multipleMember.email,
    password: multiplePassword,
    purpose: "enterprise root multiple incidents",
    userId: multipleMember.user_id,
  });
  await page.goto("/");
  await expect(
    page.getByTestId(landingIncidentCardTestId(multipleIncidentA)),
  ).toBeVisible();
  await expect(
    page.getByTestId(landingIncidentCardTestId(multipleIncidentB)),
  ).toBeVisible();
  await expect(
    page.getByTestId(incidentLandingTestId("incidents-count")),
  ).toHaveText("2 loaded");
  await expect(page).not.toHaveURL(/incident_id=/);

  await sessionTracker.loginTrackedUser(page, {
    createdBy: "extension_profile.enterprise-auth",
    email: singleMember.email,
    password: singlePassword,
    purpose: "enterprise root convergence",
    userId: singleMember.user_id,
  });
  await page.goto("/");
  await expect(page).toHaveURL(new RegExp(`incident_id=${incidentId}`));
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();

  await page.route("**/api/v1/incidents**", async (route) => {
    const requestURL = new URL(route.request().url());
    if (
      route.request().method().toUpperCase() !== "GET" ||
      requestURL.pathname !== "/api/v1/incidents"
    ) {
      await route.fallback();
      return;
    }
    const response = await route.fetch();
    const body = await response.json();
    await route.fulfill({
      response,
      contentType: "application/json",
      body: JSON.stringify({
        ...body,
        data: {
          ...body.data,
          incidents: [],
        },
      }),
    });
  });
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page).not.toHaveURL(/incident_id=/);
  await expect(
    page.getByTestId(incidentLandingTestId("empty-state")),
  ).toBeVisible();
});
