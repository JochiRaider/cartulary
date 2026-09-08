import { incidentLandingTestId } from "@cartulary/ui-contracts";
import { expect, test } from "./fixtures";
import { openIncidentControls } from "./pages/deploymentAdministration";
import { openIncidentFromLanding } from "./pages/incidentDirectory";
import { auditBrowserBarrier } from "./support/administrativeAudit";
import { csrfHeaders } from "./support/auth/browserSession";
import { createDeploymentUser } from "./support/auth/deploymentUsers";
import {
  installMembershipManagementPresentation,
  membershipBrowserMember,
  openMembershipManagement,
} from "./support/incidentMembershipManagement";
import { createIncident } from "./support/incidents/fixtures";
import { createIncidentMemberUser } from "./support/incidents/memberships";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { publicHttpOperation } from "./support/transport/publicHttpOperationClient";
import { atJsonOrigin } from "./support/transport/publicJsonClient";

test("membership management traverses live service pages and adds changes removes with audit visibility", async ({
  workerAdminPage: page,
}) => {
  test.setTimeout(180_000);
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("MM-PAGES"),
    "Membership page administration",
  );
  for (let start = 0; start < 100; start += 4) {
    await Promise.all(
      Array.from({ length: 4 }, (_, offset) =>
        createIncidentMemberUser(page, incidentId, {
          email: uniqueEmail(`mm-page-${start + offset}`),
          display_name: `Page member ${start + offset}`,
          initial_password: "MembershipPages1!",
          role: "viewer",
          is_deployment_admin: false,
          mfa_required: false,
        }),
      ),
    );
  }
  const email = uniqueEmail("mm-late");
  const target = await createDeploymentUser(page, {
    email,
    display_name: "Later member",
    initial_password: "MembershipPages1!",
    is_deployment_admin: false,
    mfa_required: false,
  });
  await openIncidentFromLanding(page, incidentId);
  const panel = await openMembershipManagement(page);
  await expect(panel.getByText(/Page 1: 100 members/u)).toBeVisible();
  await expect(panel.getByRole("article")).toHaveCount(100);
  await createIncidentMemberUser(page, incidentId, {
    email: uniqueEmail("mm-live"),
    display_name: "Live continuation member",
    initial_password: "MembershipPages1!",
    role: "viewer",
    is_deployment_admin: false,
    mfa_required: false,
  });
  await panel.getByRole("button", { name: "Next page", exact: true }).click();
  await expect(
    panel.getByText("Live continuation member", { exact: true }),
  ).toBeVisible();
  await panel
    .getByRole("button", { name: "Previous page", exact: true })
    .click();
  await expect(panel.getByRole("article")).toHaveCount(100);
  await panel
    .getByRole("button", { name: "Add existing account", exact: true })
    .click();
  await panel.getByLabel("User email", { exact: true }).fill(email);
  await panel
    .getByRole("button", { name: "Add membership", exact: true })
    .click();
  await expect(
    panel.getByText("Membership added.", { exact: true }),
  ).toBeVisible();
  await expect(panel.getByText(/Page 1: 100 members/u)).toBeVisible();
  await panel.getByRole("button", { name: "Next page", exact: true }).click();
  await expect(panel.getByText("Later member", { exact: true })).toBeVisible();
  await expect(panel.getByRole("article")).toHaveCount(3);
  await panel
    .getByRole("button", { name: "Previous page", exact: true })
    .click();
  await expect(panel.getByRole("article")).toHaveCount(100);
  await panel.getByRole("button", { name: "Next page", exact: true }).click();
  await panel
    .getByRole("button", {
      name: `Change role for Later member (${target.user_id})`,
      exact: true,
    })
    .click();
  await panel
    .getByRole("combobox", { name: /^Role for Later member/u })
    .selectOption("reviewer");
  await panel.getByRole("button", { name: "Save role", exact: true }).click();
  await expect(
    panel.getByText("Membership role saved.", { exact: true }),
  ).toBeVisible();
  await expect(panel.getByText(/Page 1: 100 members/u)).toBeVisible();
  await panel.getByRole("button", { name: "Next page", exact: true }).click();
  await panel
    .getByRole("button", {
      name: `Remove incident access for Later member (${target.user_id})`,
      exact: true,
    })
    .click();
  await expect(panel.getByText(/Reviewed version: 2/u)).toBeVisible();
  await panel
    .getByRole("button", { name: "Confirm removal", exact: true })
    .click();
  await expect(
    panel.getByText("Incident membership removed.", { exact: true }),
  ).toBeVisible();
  await page
    .getByRole("button", { name: "Close incident controls", exact: true })
    .click();
  await openIncidentControls(page, "membership-audit");
  const audit = page.getByRole("region", {
    name: "Incident membership audit browser",
    exact: true,
  });
  await audit
    .getByLabel("Target kind", { exact: true })
    .selectOption("incident_membership");
  await audit.getByLabel("Target ID", { exact: true }).fill(target.user_id);
  await audit
    .getByRole("button", { name: "Apply filters", exact: true })
    .click();
  await expect(audit.getByRole("status")).toContainText("Page 1: 3 events.");
  await expect(
    audit.getByText("Membership created", { exact: true }),
  ).toBeVisible();
  await expect(
    audit.getByText("Membership role changed", { exact: true }),
  ).toBeVisible();
  await expect(
    audit.getByText("Membership deleted", { exact: true }),
  ).toBeVisible();
});

test("membership management separates confirmed refresh failure exact create replay and explicit conflict review", async ({
  workerAdminPage: page,
}) => {
  const fixture = await installMembershipManagementPresentation(page);
  const panel = await openMembershipManagement(page);
  await expect(
    panel.getByText("Response analyst", { exact: true }),
  ).toBeVisible();
  const add = async () => {
    await panel
      .getByRole("button", { name: "Add existing account", exact: true })
      .click();
    await panel
      .getByLabel("User email", { exact: true })
      .fill("existing@example.test");
  };
  await add();
  const pending = auditBrowserBarrier();
  fixture.gateWrite(pending.promise);
  fixture.mutation(201, membershipBrowserMember(fixture.incidentId));
  await panel
    .getByRole("button", { name: "Add membership", exact: true })
    .dblclick();
  await expect.poll(() => fixture.writes.length).toBe(1);
  await panel
    .getByLabel("User email", { exact: true })
    .fill("newer@example.test");
  fixture.failRead();
  pending.release();
  await expect(
    panel.getByText("Membership added.", { exact: true }),
  ).toBeVisible();
  await expect(
    panel.getByRole("button", { name: "Retry member refresh", exact: true }),
  ).toBeVisible();
  await expect(panel.getByLabel("User email", { exact: true })).toHaveValue(
    "newer@example.test",
  );
  fixture.setPage([membershipBrowserMember(fixture.incidentId)]);
  await panel
    .getByRole("button", { name: "Retry member refresh", exact: true })
    .click();
  await expect(panel.getByText(/Page 1:/u)).toBeVisible();
  expect(fixture.writes).toHaveLength(1);
  fixture.mutation(503);
  await panel
    .getByRole("button", { name: "Add membership", exact: true })
    .click();
  await expect(panel.getByText(/has an uncertain outcome/u)).toBeVisible();
  const original = fixture.writes[1]?.body;
  await panel
    .getByLabel("User email", { exact: true })
    .fill("changed@example.test");
  fixture.mutation(200, membershipBrowserMember(fixture.incidentId));
  await panel
    .getByRole("button", { name: "Replay exact add request", exact: true })
    .click();
  await expect(
    panel.getByText(/Membership confirmed: existing membership/u),
  ).toBeVisible();
  expect(fixture.writes[2]?.body).toEqual(original);
  await expect(panel.getByLabel("User email", { exact: true })).toHaveValue(
    "changed@example.test",
  );
  await panel
    .getByRole("button", { name: /^Change role for Response analyst/u })
    .click();
  await page
    .getByRole("button", { name: "Discard draft", exact: true })
    .click();
  fixture.mutation(409, null, "membership_version_conflict");
  await panel
    .getByRole("combobox", { name: /^Role for Response analyst/u })
    .selectOption("admin");
  await panel.getByRole("button", { name: "Save role", exact: true }).click();
  await expect(
    panel.getByText(/membership changed after it was reviewed/u),
  ).toBeVisible();
  fixture.setPage([
    membershipBrowserMember(fixture.incidentId, {
      role: "reviewer",
      membership_version: 2,
    }),
  ]);
  await panel
    .getByRole("button", { name: "Observe current membership", exact: true })
    .click();
  await expect(
    panel.getByText(/Observed role: reviewer; version: 2/u),
  ).toBeVisible();
  await panel
    .getByRole("button", { name: "Review observed membership", exact: true })
    .first()
    .click();
  expect(fixture.writes).toHaveLength(4);
  await expect(panel.getByText(/Reviewed version: 2/u)).toBeVisible();
  await expect(panel.getByRole("combobox")).toHaveValue("admin");
  fixture.setAccess("unavailable");
  await panel.getByRole("button", { name: "Refresh", exact: true }).click();
  await expect(
    panel.getByRole("button", { name: "Check access and reload", exact: true }),
  ).toBeVisible();
  await expect(panel.getByRole("combobox")).toHaveCount(0);
  await expect(panel.getByRole("article")).toHaveCount(0);
  fixture.setAccess("admin");
  await panel
    .getByRole("button", { name: "Check access and reload", exact: true })
    .click();
  await expect(panel.getByRole("combobox")).toHaveValue("admin");
});

test("membership management retains pending work on close and requires deliberate route departure", async ({
  workerAdminPage: page,
}) => {
  const fixture = await installMembershipManagementPresentation(page);
  const panel = await openMembershipManagement(page);
  await expect(
    panel.getByText("Response analyst", { exact: true }),
  ).toBeVisible();
  await panel
    .getByRole("button", {
      name: /^Remove incident access for Response analyst/u,
    })
    .click();
  const pending = auditBrowserBarrier();
  fixture.gateWrite(pending.promise);
  fixture.mutation(204);
  await panel
    .getByRole("button", { name: "Confirm removal", exact: true })
    .click();
  await expect.poll(() => fixture.writes.length).toBe(1);
  await page
    .getByRole("button", { name: "Close incident controls", exact: true })
    .click();
  await openMembershipManagement(page);
  await expect(
    panel.getByText(/Saving the reviewed membership action/u),
  ).toBeVisible();
  await page
    .getByRole("button", { name: "Close incident controls", exact: true })
    .click();
  const leave = async () => {
    await page.getByLabel("Account and application navigation").click();
    await page
      .getByRole("menuitem", { name: "Incidents", exact: true })
      .click();
  };
  await leave();
  await expect(
    page.getByRole("dialog", { name: "Leave membership work?" }),
  ).toBeVisible();
  await page.getByRole("button", { name: "Stay", exact: true }).click();
  expect(new URL(page.url()).searchParams.get("incident_id")).not.toBe("");
  await leave();
  await page
    .getByRole("button", { name: "Leave and forget recovery", exact: true })
    .click();
  await expect(page.getByTestId(incidentLandingTestId("shell"))).toBeVisible();
  pending.release();
  await expect(panel).toHaveCount(0);
  expect(fixture.writes).toHaveLength(1);
});

test("membership management self-demotion and self-removal recover current authority on closed incidents", async ({
  workerAdminPage: page,
}) => {
  const session = await page.request.get(`${apiBase}/api/v1/auth/session`);
  expect(session.ok()).toBe(true);
  const actor = (await session.json()).data as {
    user_id: string;
    display_name: string;
  };
  for (const action of ["demote", "remove"] as const) {
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey(`MM-SELF-${action}`),
      "Closed incident membership review",
    );
    await createIncidentMemberUser(page, incidentId, {
      email: uniqueEmail(`mm-other-admin-${action}`),
      display_name: "Other incident administrator",
      initial_password: "MembershipSelfReview1!",
      role: "admin",
      is_deployment_admin: false,
      mfa_required: false,
    });
    const closed = await publicHttpOperation({
      request: atJsonOrigin(page.request, apiBase),
      headers: await csrfHeaders(page),
      operationID: "closeIncident",
      pathParameters: { incident_id: incidentId },
      body: {
        base_incident_version: 1,
        client_txn_id: uniqueTxn("close-membership-review"),
        reason: "Review membership after closure",
      },
    });
    expect(closed.ok).toBe(true);
    await openIncidentFromLanding(page, incidentId);
    const panel = await openMembershipManagement(page);
    await expect(panel.getByText(actor.user_id, { exact: true })).toBeVisible();
    if (action === "demote") {
      await panel
        .getByRole("button", {
          name: `Change role for ${actor.display_name} (${actor.user_id})`,
          exact: true,
        })
        .click();
      await panel
        .getByRole("combobox", {
          name: `Role for ${actor.display_name} (${actor.user_id})`,
          exact: true,
        })
        .selectOption("viewer");
      await expect(
        panel.getByText(/lose membership administration/u),
      ).toBeVisible();
      await panel
        .getByRole("button", { name: "Save role", exact: true })
        .click();
      await expect(panel.getByText(/Only incident admins/u)).toBeVisible();
      await expect(
        panel.getByText(actor.user_id, { exact: true }),
      ).toBeVisible();
      await expect(
        panel.getByRole("button", {
          name: "Add existing account",
          exact: true,
        }),
      ).toHaveCount(0);
    } else {
      await panel
        .getByRole("button", {
          name: `Remove incident access for ${actor.display_name} (${actor.user_id})`,
          exact: true,
        })
        .click();
      await expect(
        panel.getByText(/lose access to this incident/u),
      ).toBeVisible();
      await panel
        .getByRole("button", { name: "Confirm removal", exact: true })
        .click();
      await expect(
        page.getByTestId(incidentLandingTestId("shell")),
      ).toBeVisible();
      await expect(panel).toHaveCount(0);
    }
    const observed = await page.request.get(`${apiBase}/api/v1/auth/session`);
    expect(observed.ok()).toBe(true);
    expect((await observed.json()).data.user_id).toBe(actor.user_id);
  }
});
