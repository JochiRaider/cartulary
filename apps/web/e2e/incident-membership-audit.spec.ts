import {
  incidentMembershipAuditDetailTestId,
  incidentMembershipAuditRowTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import { expect, test } from "./fixtures";
import { openIncidentControls } from "./pages/deploymentAdministration";
import {
  openIncidentAsTrackedUser,
  openIncidentFromLanding,
} from "./pages/incidentDirectory";
import {
  auditBrowserBarrier,
  auditBrowserEventId,
} from "./support/administrativeAudit";
import {
  installMembershipAuditPresentation,
  membershipBrowserEvent,
  openMembershipAudit,
} from "./support/incidentMembershipAudit";
import { createIncident } from "./support/incidents/fixtures";
import { createIncidentMemberUser } from "./support/incidents/memberships";
import {
  uniqueEmail,
  uniqueIncidentKey,
} from "./support/runtime/fixtureIdentity";
import { publicHttpOperation } from "./support/transport/publicHttpOperationClient";

test("membership audit browses 105 real exact role changes with accepted query pagination and instant boundaries", async ({
  workerAdminPage: page,
  workerAdmin,
  workerAdminRequest,
  browser,
  sessionTracker,
}, testInfo) => {
  test.setTimeout(120_000);
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("AUDIT-PAGING"),
    "Membership audit pagination",
  );
  const member = await createIncidentMemberUser(page, incidentId, {
    email: uniqueEmail("audit-member"),
    display_name: "Audit pagination member",
    initial_password: "AuditMember!2026",
    role: "viewer",
    mfa_required: false,
    is_deployment_admin: false,
  });
  let version = 1;
  for (let index = 0; index < 105; index++) {
    const result = await workerAdminRequest.patch(
      `/api/v1/incidents/${incidentId}/memberships/${member.user_id}`,
      {
        data: {
          base_membership_version: version,
          role: index % 2 === 0 ? "editor" : "viewer",
        },
      },
    );
    expect(result.ok()).toBe(true);
    version = (
      (await result.json()) as { data: { membership_version: number } }
    ).data.membership_version;
  }
  const query = {
    actor_user_id: workerAdmin.user_id,
    action_code: "membership_role_changed",
    target_kind: "incident_membership",
    target_id: member.user_id,
    limit: 100,
  };
  const read = (extra: Record<string, string> = {}) =>
    publicHttpOperation({
      operationID: "listIncidentMembershipAuditEvents",
      request: workerAdminRequest,
      pathParameters: { incident_id: incidentId },
      query: { ...query, ...extra },
    });
  const first = await read();
  if (!first.ok) throw new Error("Missing audit first page");
  expect(first.payload.data.audit_events).toHaveLength(100);
  const cursor = first.payload.meta.paging?.next_cursor;
  const newest = first.payload.data.audit_events[0];
  if (!cursor || !newest) throw new Error("Missing continuation or event");
  const second = await read({ cursor_token: cursor });
  const inclusive = await read({ occurred_at_gte: newest.occurred_at });
  const exclusive = await read({ occurred_at_lt: newest.occurred_at });
  if (!second.ok || !inclusive.ok || !exclusive.ok)
    throw new Error("Audit boundary read failed");
  expect(second.payload.data.audit_events).toHaveLength(5);
  expect(
    inclusive.payload.data.audit_events.map((event) => event.audit_event_id),
  ).toContain(newest.audit_event_id);
  expect(
    exclusive.payload.data.audit_events.map((event) => event.audit_event_id),
  ).not.toContain(newest.audit_event_id);
  await openIncidentFromLanding(page, incidentId);
  const requests: URL[] = [];
  page.on("request", (request) => {
    if (request.url().includes("/membership-audit-events"))
      requests.push(new URL(request.url()));
  });
  const panel = await openMembershipAudit(page);
  await expect(panel.getByRole("status")).toContainText("Page 1:");
  await panel
    .getByLabel("Actor user ID", { exact: true })
    .fill(query.actor_user_id);
  await panel
    .getByLabel("Action code", { exact: true })
    .selectOption(query.action_code);
  await panel
    .getByLabel("Target kind", { exact: true })
    .selectOption(query.target_kind);
  await panel.getByLabel("Target ID", { exact: true }).fill(query.target_id);
  await panel.getByLabel("Target ID", { exact: true }).press("Enter");
  await expect(panel.getByRole("status")).toContainText("Page 1: 100 events.");
  await expect(panel.getByRole("button", { name: /^Inspect / })).toHaveCount(
    100,
  );
  await panel
    .getByLabel("Action code", { exact: true })
    .selectOption("membership_deleted");
  const next = panel.getByRole("button", { name: "Next page", exact: true });
  await next.focus();
  await page.keyboard.press("Enter");
  await expect(panel.getByRole("status")).toContainText("Page 2: 5 events.");
  await expect(next).toBeFocused();
  await expect(next).toBeDisabled();
  expect(requests.at(-1)?.searchParams.get("action_code")).toBe(
    query.action_code,
  );
  for (const event of second.payload.data.audit_events)
    await expect(
      panel.getByTestId(incidentMembershipAuditRowTestId(event.audit_event_id)),
    ).toHaveCount(1);
  await expect(panel.getByRole("button", { name: /^Inspect / })).toHaveCount(5);
  await panel
    .getByRole("button", { name: "Previous page", exact: true })
    .click();
  await expect(panel.getByRole("status")).toContainText("Page 1: 100 events.");
  await panel.getByRole("button", { name: "Refresh", exact: true }).click();
  await expect(panel.getByRole("status")).toContainText("Page 1: 100 events.");
  expect(requests.at(-1)?.searchParams.get("action_code")).toBe(
    query.action_code,
  );
  await panel
    .getByLabel("Action code", { exact: true })
    .selectOption(query.action_code);
  await panel.getByLabel("Occurred at or after").fill(newest.occurred_at);
  await panel.getByLabel("Occurred at or after").press("Enter");
  await expect(panel.getByRole("button", { name: /^Inspect / })).toHaveCount(
    inclusive.payload.data.audit_events.length,
  );
  await panel
    .getByRole("button", { name: /^Inspect / })
    .first()
    .click();
  await expect(
    panel.getByTestId(
      incidentMembershipAuditDetailTestId(newest.audit_event_id),
    ),
  ).toContainText('"editor"');
  await page.keyboard.press("Escape");
  await expect(
    page.getByLabel("Account and application navigation"),
  ).toBeFocused();
  const memberPage = await openIncidentAsTrackedUser(browser, sessionTracker, {
    createdBy: "membership audit",
    purpose: "non-admin audit denial",
    email: member.email,
    password: member.initial_password,
    userId: member.user_id,
    incidentId,
  });
  const denied = await openMembershipAudit(memberPage);
  await expect(denied.getByText(/Only incident admins/)).toBeVisible();
  await expect(denied.getByRole("button", { name: /^Inspect / })).toHaveCount(
    0,
  );
  await memberPage.context().close();
  await testInfo.attach("real-membership-audit-pagination", {
    body: JSON.stringify({
      generated_matching_events: 105,
      page_sizes: [100, 5],
      unique_ids: new Set(
        [
          ...first.payload.data.audit_events,
          ...second.payload.data.audit_events,
        ].map((event) => event.audit_event_id),
      ).size,
      inclusive_lower: true,
      exclusive_upper: true,
      draft_continuation_unchanged: true,
    }),
    contentType: "application/json",
  });
});

test("membership audit fences delayed reads across query drawer section and visibility with explicit cursor recovery", async ({
  workerAdminPage: page,
}) => {
  const fixture = await installMembershipAuditPresentation(page);
  const event = membershipBrowserEvent(fixture.incidentId);
  fixture.setPage([event], "next-page");
  const panel = await openMembershipAudit(page);
  await expect(panel.getByRole("status")).toContainText("Page 1:");
  await panel.getByRole("button", { name: /^Inspect / }).click();
  await expect(
    panel.getByTestId(incidentMembershipAuditDetailTestId(auditBrowserEventId)),
  ).toContainText("Redacted");
  const pending = auditBrowserBarrier();
  fixture.fail();
  fixture.gateRead(pending.promise);
  const count = fixture.requests.length;
  await panel.getByRole("button", { name: "Next page", exact: true }).click();
  await expect.poll(() => fixture.requests.length).toBe(count + 1);
  await expect(
    panel.getByRole("button", { name: "Next page", exact: true }),
  ).toBeDisabled();
  fixture.setPage([]);
  await panel
    .getByLabel("Action code", { exact: true })
    .selectOption("membership_deleted");
  await panel
    .getByRole("button", { name: "Apply filters", exact: true })
    .click();
  await expect(
    panel.getByText("No membership audit events match the applied filters."),
  ).toBeVisible();
  pending.release();
  await expect(panel.getByRole("status")).toContainText("Page 1: 0 events.");
  fixture.setPage([event], "next-page");
  await panel
    .getByRole("button", { name: "Clear filters", exact: true })
    .click();
  await expect(
    panel.getByRole("button", { name: "Next page", exact: true }),
  ).toBeEnabled();
  fixture.fail("invalid_pagination_request", "cursor_expired");
  await panel.getByRole("button", { name: "Next page", exact: true }).click();
  await expect(
    panel.getByRole("button", { name: "Reload first page", exact: true }),
  ).toBeVisible();
  await page.keyboard.press("Escape");
  const rejectedReads = fixture.requests.length;
  fixture.setPage([event]);
  await openMembershipAudit(page);
  await expect(
    panel.getByRole("button", { name: "Reload first page", exact: true }),
  ).toBeEnabled();
  expect(fixture.requests).toHaveLength(rejectedReads);
  await panel
    .getByRole("button", { name: "Reload first page", exact: true })
    .click();
  await expect(panel.getByRole("status")).toContainText("Page 1: 1 event.");
  for (const boundary of ["close", "section", "visibility"] as const) {
    const gate = auditBrowserBarrier();
    fixture.setPage([
      membershipBrowserEvent(fixture.incidentId, {
        action_code: "late_action",
      }),
    ]);
    fixture.gateRead(gate.promise);
    const before = fixture.requests.length;
    await panel.getByRole("button", { name: "Refresh", exact: true }).click();
    await expect.poll(() => fixture.requests.length).toBe(before + 1);
    if (boundary === "visibility")
      await page.evaluate(() => {
        Object.defineProperty(document, "visibilityState", {
          configurable: true,
          value: "hidden",
        });
        document.dispatchEvent(new Event("visibilitychange"));
      });
    else {
      await page.keyboard.press("Escape");
      if (boundary === "section") await openIncidentControls(page, "summary");
    }
    gate.release();
    fixture.setPage([event]);
    if (boundary === "visibility")
      await page.evaluate(() => {
        Object.defineProperty(document, "visibilityState", {
          configurable: true,
          value: "visible",
        });
        document.dispatchEvent(new Event("visibilitychange"));
      });
    else {
      if (boundary === "section") await page.keyboard.press("Escape");
      await openMembershipAudit(page);
    }
    await expect(panel.getByRole("status")).toContainText("Page 1: 1 event.");
    await expect(panel.getByText("late_action", { exact: true })).toHaveCount(
      0,
    );
  }
});

test("membership audit distinguishes temporary observation failure role downgrade hidden incident and session loss", async ({
  workerAdminPage: page,
}) => {
  const fixture = await installMembershipAuditPresentation(page);
  const panel = await openMembershipAudit(page);
  await expect(panel.getByRole("status")).toContainText("Page 1:");
  fixture.setAccess("unavailable");
  await panel.getByRole("button", { name: "Refresh", exact: true }).click();
  await expect(panel.getByRole("status")).toContainText(
    "access could not be checked",
  );
  await expect(panel.getByRole("button", { name: /^Inspect / })).toHaveCount(1);
  fixture.setAccess("admin");
  await panel
    .getByRole("button", { name: "Try the read again", exact: true })
    .click();
  await expect(panel.getByRole("status")).toContainText("Page 1:");
  fixture.setAccess("viewer");
  await panel.getByRole("button", { name: "Refresh", exact: true }).click();
  await expect(panel.getByText(/Only incident admins/)).toBeVisible();
  await expect(panel.getByRole("button", { name: /^Inspect / })).toHaveCount(0);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  expect(new URL(page.url()).searchParams.get("incident_id")).toBe(
    fixture.incidentId,
  );
  fixture.setAccess("admin");
  await page.reload();
  await openMembershipAudit(page);
  await expect(panel.getByRole("status")).toContainText("Page 1:");
  fixture.setAccess("hidden");
  await panel.getByRole("button", { name: "Refresh", exact: true }).click();
  await expect(page).not.toHaveURL(/incident_id=/);
  await expect(panel).toHaveCount(0);
  fixture.setAccess("admin");
  await openIncidentFromLanding(page, fixture.incidentId);
  await openMembershipAudit(page);
  await expect(panel.getByRole("status")).toContainText("Page 1:");
  fixture.setAccess("session");
  await panel.getByRole("button", { name: "Refresh", exact: true }).click();
  await expect(panel).toHaveCount(0);
  await expect(page.getByRole("button", { name: /Sign in/i })).toBeVisible();
});

test("membership audit rejects malformed scope and preserves additive inert JSON inspection", async ({
  workerAdminPage: page,
}) => {
  const fixture = await installMembershipAuditPresentation(page);
  const panel = await openMembershipAudit(page);
  await expect(panel.getByRole("status")).toContainText("Page 1:");
  fixture.setPage([
    membershipBrowserEvent("00000000-0000-4000-8000-000000009999", {
      action_code: "wrong_scope",
    }),
  ]);
  await panel.getByRole("button", { name: "Refresh", exact: true }).click();
  await expect(panel.getByRole("status")).toContainText("refresh failed");
  await expect(panel.getByText("wrong_scope", { exact: true })).toHaveCount(0);
  fixture.malformed({
    data: { audit_events: [] },
    meta: { request_id: "missing-paging" },
  });
  await panel
    .getByRole("button", { name: "Try the read again", exact: true })
    .click();
  await expect(panel.getByRole("status")).toContainText("refresh failed");
  fixture.setPage([
    membershipBrowserEvent(fixture.incidentId, {
      action_code: "future_action",
      target_kind: "future_target",
      target_id: "long_target_".repeat(40),
      changes: [
        {
          field_path: "a.string",
          value_state: "visible",
          before: null,
          after: "null",
        },
        {
          field_path: "b.JSON",
          value_state: "visible",
          before: [false, 0, ""],
          after: { message: "<script>inert</script>" },
        },
        {
          field_path: "c.credential",
          value_state: "redacted",
          before: null,
          after: null,
        },
      ],
    }),
  ]);
  await panel
    .getByRole("button", { name: "Try the read again", exact: true })
    .click();
  await expect(panel.getByRole("status")).toContainText("Page 1:");
  await panel.getByRole("button", { name: /^Inspect future_action/ }).click();
  const details = panel.getByRole("region", { name: /^Details for / });
  await expect(details.getByText('"null"', { exact: true })).toBeVisible();
  await expect(details.getByText("null", { exact: true })).toBeVisible();
  await expect(details).toContainText("false");
  await expect(details).toContainText("<script>inert</script>");
  await expect(details.getByText("Redacted", { exact: true })).toHaveCount(2);
  await expect(panel.locator("script")).toHaveCount(0);
  await expect(
    panel
      .getByLabel("Action code", { exact: true })
      .getByRole("option", { name: "future_action" }),
  ).toHaveCount(0);
  await expect(
    panel
      .getByLabel("Target kind", { exact: true })
      .getByRole("option", { name: "future_target" }),
  ).toHaveCount(0);
});
