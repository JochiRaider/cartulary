import {
  authTestId,
  genericCreateSubmitTestId,
  incidentLandingTestId,
} from "@cartulary/ui-contracts";
import {
  evidenceViewSchemaId as evidence,
  findingsViewSchemaId,
} from "@cartulary/view-contracts";
import { expect, test } from "./fixtures";
import {
  confirmLifecycle,
  currentLifecycle,
  lifecycleAction,
  openLifecycle,
} from "./support/incidentLifecycle";
import { createIncidentMemberUser } from "./support/incidents/memberships";
import { uniqueEmail, uniqueTxn } from "./support/runtime/fixtureIdentity";
import { installIncidentSocketMonitor } from "./support/transport/incidentSocket";
import {
  commitOrdinary,
  fillOrdinaryField,
  openOrdinaryFixture,
  ordinaryField,
  retainOrdinaryUncertainty,
  switchOrdinarySheet,
} from "./support/workbook/ordinaryCreate";
import { queryViewRows } from "./support/workbook/query";

test("Ordinary closure preserves copyable next authoring and requires explicit exact recovery before fresh creation", async ({
  page,
}) => {
  const { incident } = await openOrdinaryFixture(page);
  await fillOrdinaryField(
    page,
    evidence,
    "evidence.title",
    "Committed before closure",
  );
  const { bodies, recovery } = await retainOrdinaryUncertainty(page, incident);
  await fillOrdinaryField(
    page,
    evidence,
    "evidence.title",
    "Unsent after reopening",
  );
  const before = await currentLifecycle(page, incident);
  expect(
    (
      await lifecycleAction(page, incident, "closeIncident", {
        client_txn_id: uniqueTxn("ordinary-close"),
        base_incident_version: before.incident_version,
        reason: "Retain ordinary work",
      })
    ).ok,
  ).toBe(true);
  const retainedDraft = page.getByRole("textbox", {
    name: "Title retained authoring",
    exact: true,
  });
  await expect(retainedDraft).toHaveAttribute("readonly", "");
  await expect(retainedDraft).toHaveValue("Unsent after reopening");
  expect(bodies).toHaveLength(1);
  await recovery
    .getByRole("button", { name: "Recover submission", exact: true })
    .click();
  await expect(recovery).toContainText("Row accepted.");
  expect(bodies).toHaveLength(2);
  expect(bodies[1]).toBe(bodies[0]);
  const panel = await openLifecycle(page);
  await confirmLifecycle(page, "Reopen", "Explicit ordinary authoring resumes");
  await expect(
    panel.getByText("Reopen confirmed.", { exact: true }),
  ).toBeVisible();
  await page
    .getByRole("button", { name: "Close incident controls", exact: true })
    .click();
  const draft = await ordinaryField(page, evidence, "evidence.title");
  await expect(draft).toBeEditable();
  await expect(draft).toHaveValue("Unsent after reopening");
  expect(bodies).toHaveLength(2);
  expect(await queryViewRows(page, incident, evidence)).toHaveLength(1);
  await commitOrdinary(page, evidence);
  await expect.poll(() => bodies.length).toBe(3);
  expect(JSON.parse(bodies[2] ?? "{}").client_txn_id).not.toBe(
    JSON.parse(bodies[0] ?? "{}").client_txn_id,
  );
  await expect
    .poll(async () => (await queryViewRows(page, incident, evidence)).length)
    .toBe(2);
});

test("Ordinary role loss retains readable authoring and scoped revocation fences a late accepted receipt", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
  let memberId = "";
  const { incident } = await openOrdinaryFixture(
    page,
    evidence,
    async (url) => {
      const incident = new URL(url, "http://fixture").searchParams.get(
        "incident_id",
      );
      if (!incident) throw new Error("Missing incident");
      const member = await createIncidentMemberUser(page, incident, {
        email: uniqueEmail("ordinary-revoked"),
        display_name: "Ordinary author",
        initial_password: "OrdinaryMember1!",
        role: "editor",
        is_deployment_admin: false,
        mfa_required: false,
      });
      memberId = member.user_id;
      await sessionTracker.loginTrackedUser(page, {
        createdBy: "ordinary-revocation",
        email: member.email,
        password: member.initial_password,
        purpose: "retained ordinary revocation",
        userId: memberId,
      });
      const sockets = installIncidentSocketMonitor(page, incident);
      await page.goto(url);
      await sockets.waitForAcceptedSocket();
    },
  );
  await fillOrdinaryField(
    page,
    evidence,
    "evidence.title",
    "Protected ordinary work",
  );
  let committed = false;
  let release: () => void = () => {};
  const gate = new Promise<void>((resolve) => {
    release = resolve;
  });
  const suffix = `/views/${evidence}/rows`;
  await page.route(`**/incidents/${incident}${suffix}`, async (route) => {
    const response = await route.fetch();
    expect(response.ok()).toBe(true);
    committed = true;
    await gate;
    await route.fulfill({ response });
  });
  await commitOrdinary(page, evidence);
  await expect.poll(() => committed).toBe(true);
  await fillOrdinaryField(
    page,
    evidence,
    "evidence.title",
    "Private next draft",
  );
  const membershipPath = `/api/v1/incidents/${incident}/memberships/${memberId}`;
  const downgrade = await workerAdminRequest.patch(membershipPath, {
    data: { base_membership_version: 1, role: "viewer" },
  });
  expect(downgrade.ok()).toBe(true);
  // Membership updates are observed at the existing authority read boundary.
  // A fresh attempt on another schema must revalidate and dispatch nothing.
  const findings = findingsViewSchemaId;
  let unauthorizedWrites = 0;
  page.on("request", (request) => {
    if (
      request.method() === "POST" &&
      request.url().endsWith(`/views/${findings}/rows`)
    )
      unauthorizedWrites++;
  });
  await switchOrdinarySheet(page, findings);
  await fillOrdinaryField(
    page,
    findings,
    "finding.statement",
    "Cannot dispatch as a viewer",
  );
  await commitOrdinary(page, findings);
  await expect(
    page.getByRole("textbox", {
      name: "Statement retained authoring",
      exact: true,
    }),
  ).toHaveAttribute("readonly", "");
  expect(unauthorizedWrites).toBe(0);
  await switchOrdinarySheet(page, evidence);

  const draft = page.getByRole("textbox", {
    name: "Title retained authoring",
    exact: true,
  });
  await expect(draft).toHaveAttribute("readonly", "");
  await expect(draft).toHaveValue("Private next draft");
  await expect(
    page.getByTestId(genericCreateSubmitTestId(evidence)),
  ).toHaveCount(0);
  const response = await workerAdminRequest.get(
    `/api/v1/incidents/${incident}/memberships`,
  );
  const members = (await response.json()) as {
    data: { memberships: { user_id: string; membership_version: number }[] };
  };
  const member = members.data.memberships.find(
    (item) => item.user_id === memberId,
  );
  if (!member) throw new Error("Missing membership");
  expect(
    (
      await workerAdminRequest.delete(membershipPath, {
        data: { base_membership_version: member.membership_version },
      })
    ).status(),
  ).toBe(204);
  await expect(page.getByTestId(incidentLandingTestId("shell"))).toBeVisible();
  await expect(draft).toHaveCount(0);
  const late = page.waitForResponse((r) => r.url().endsWith(suffix));
  release();
  await late;
  await expect(
    page.getByRole("region", { name: "Row creation", exact: true }),
  ).toHaveCount(0);
  await expect(page.getByTestId(authTestId("shell"))).toHaveCount(0);
  await expect(page).not.toHaveURL(/incident_id=/u);
});
