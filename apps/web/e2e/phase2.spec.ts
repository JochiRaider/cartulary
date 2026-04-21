import type { Browser, Page } from "@playwright/test";

import { expect, test } from "./fixtures";
import {
  createIncident,
  createLocalUser,
  uniqueEmail,
  uniqueIncidentKey,
} from "./helpers";

test("E-2-01 creates an incident, bootstraps the creator as admin, and lands on the workbook surface", async ({
  page,
}) => {
  await page.goto("/");

  const incidentKey = uniqueIncidentKey("E201");
  await page.getByTestId("landing-incident-key").fill(incidentKey);
  await page.getByTestId("landing-incident-title").fill("Phase 2 E-2-01");
  await page.getByTestId("landing-create-button").click();

  await expect(page).toHaveURL(/incident_id=/);
  await expect(page.getByTestId("surface-tab-timeline")).toBeVisible();
  await expect(page.getByText("Current incident role: admin")).toBeVisible();
  await expect(page.getByTestId("incident-summary-key")).toHaveText(
    incidentKey,
  );
  await expect(page.getByTestId("incident-summary-title")).toHaveText(
    "Phase 2 E-2-01",
  );
  await expect(page.getByTestId("incident-summary-role")).toHaveText("admin");
  await expect(page.getByTestId("incident-pref-default-sheet-ref")).toHaveText(
    "Unset",
  );
  await expect(page.getByTestId("incident-pref-home-sheet-ref")).toHaveText(
    "Unset",
  );
});

test("E-2-02 shows incident discovery, raw querystring deep-link retrieval, and promoted-field-only patching on the ordinary incident shell", async ({
  page,
}) => {
  const incidentKey = uniqueIncidentKey("E202");
  const incidentId = await createIncident(page, incidentKey, "Phase 2 E-2-02");

  await page.goto("/");
  await expect(
    page.getByTestId(`landing-incident-${incidentId}`),
  ).toBeVisible();
  await expect(
    page.getByTestId(`landing-incident-${incidentId}`),
  ).toContainText(incidentKey);
  await page.getByTestId(`landing-open-${incidentId}`).click();

  await expect(page).toHaveURL(new RegExp(`incident_id=${incidentId}`));
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page).toHaveURL(new RegExp(`incident_id=${incidentId}`));
  await expect(page.getByTestId("incident-summary-key")).toHaveText(
    incidentKey,
  );
  await expect(page.getByTestId("incident-summary-title")).toHaveText(
    "Phase 2 E-2-02",
  );
  await expect(page.getByTestId("incident-summary-version")).toHaveText(
    "Version 1",
  );
  await expect(page.getByTestId("incident-summary-tlp")).toHaveText("Unset");
  await expect(page.getByTestId("incident-summary-current-phase")).toHaveText(
    "Unset",
  );
  await expect(
    page.getByTestId("incident-summary-primary-external-case-ref"),
  ).toHaveText("Unset");

  await page.getByTestId("incident-patch-tlp").fill("amber");
  await page.getByTestId("incident-patch-current-phase").fill("containment");
  await page
    .getByTestId("incident-patch-external-case")
    .fill("CASE-E202-PRIMARY");
  await page.getByTestId("incident-patch-button").click();

  await expect(page.getByTestId("incident-summary-version")).toHaveText(
    "Version 2",
  );
  await expect(page.getByTestId("incident-summary-tlp")).toHaveText("amber");
  await expect(page.getByTestId("incident-summary-current-phase")).toHaveText(
    "containment",
  );
  await expect(
    page.getByTestId("incident-summary-primary-external-case-ref"),
  ).toHaveText("CASE-E202-PRIMARY");

  await page.getByTestId("landing-return").click();
  await expect(
    page.getByTestId(`landing-incident-${incidentId}`),
  ).toContainText("containment");
  await expect(
    page.getByTestId(`landing-incident-${incidentId}`),
  ).toContainText("amber");
});

test("E-2-03 lets incident admins manage memberships and hides those controls from non-admin members on the ordinary shell", async ({
  browser,
  page,
  sessionTracker,
}) => {
  const memberEmail = uniqueEmail("phase2-e203-member");
  const memberPassword = "Phase2E203Pass!";
  const memberUser = await createLocalUser(page, {
    email: memberEmail,
    display_name: "Phase 2 E203 Member",
    initial_password: memberPassword,
  });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E203"),
    "Phase 2 E-2-03",
  );

  await openIncidentFromLanding(page, incidentId);

  await page.getByTestId("incident-membership-email").fill(memberEmail);
  await page.getByTestId("incident-membership-role").selectOption("viewer");
  await page.getByTestId("incident-membership-create").click();

  await expect(
    page.getByTestId(`incident-membership-row-${memberUser.user_id}`),
  ).toBeVisible();
  await expect(
    page.getByTestId(`incident-membership-role-input-${memberUser.user_id}`),
  ).toHaveValue("viewer");

  await page
    .getByTestId(`incident-membership-role-input-${memberUser.user_id}`)
    .selectOption("reviewer");
  await page
    .getByTestId(`incident-membership-patch-${memberUser.user_id}`)
    .click();

  await expect(
    page.getByTestId(`incident-membership-role-input-${memberUser.user_id}`),
  ).toHaveValue("reviewer");
  await expect(
    page.getByTestId(`incident-membership-version-${memberUser.user_id}`),
  ).toHaveText("Version 2");

  const memberPage = await openIncidentAsTrackedUser(browser, sessionTracker, {
    email: memberEmail,
    password: memberPassword,
    incidentId,
    userId: memberUser.user_id,
  });

  await expect(
    memberPage.getByTestId("incident-membership-admin-note"),
  ).toBeVisible();
  await expect(
    memberPage.getByTestId(`incident-membership-row-${memberUser.user_id}`),
  ).toBeVisible();
  await expect(
    memberPage.getByTestId(`incident-membership-role-${memberUser.user_id}`),
  ).toHaveText("reviewer");
  await expect(
    memberPage.getByTestId("incident-membership-create"),
  ).toHaveCount(0);
  await expect(
    memberPage.getByTestId(`incident-membership-patch-${memberUser.user_id}`),
  ).toHaveCount(0);
  await expect(
    memberPage.getByTestId(`incident-membership-delete-${memberUser.user_id}`),
  ).toHaveCount(0);
  await memberPage.context().close();

  await page
    .getByTestId(`incident-membership-delete-${memberUser.user_id}`)
    .click();
  await expect(
    page.getByTestId(`incident-membership-row-${memberUser.user_id}`),
  ).toHaveCount(0);
});

async function openIncidentFromLanding(page: Page, incidentId: string) {
  await page.goto("/");
  await expect(
    page.getByTestId(`landing-incident-${incidentId}`),
  ).toBeVisible();
  await page.getByTestId(`landing-open-${incidentId}`).click();
  await expect(page).toHaveURL(new RegExp(`incident_id=${incidentId}`));
}

async function openIncidentAsTrackedUser(
  browser: Browser,
  sessionTracker: {
    loginTrackedUser: (
      page: Page,
      details: {
        createdBy: string;
        email: string;
        password: string;
        purpose: string;
        userId: string;
      },
    ) => Promise<void>;
  },
  options: {
    email: string;
    password: string;
    incidentId: string;
    userId: string;
  },
) {
  const context = await browser.newContext();
  const page = await context.newPage();
  await sessionTracker.loginTrackedUser(page, {
    createdBy: "phase2 non-admin membership view",
    email: options.email,
    password: options.password,
    purpose: "phase2 e203 non-admin incident shell",
    userId: options.userId,
  });
  await openIncidentFromLanding(page, options.incidentId);
  return page;
}
