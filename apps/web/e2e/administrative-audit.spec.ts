import { randomUUID } from "node:crypto";
import {
  administrativeAuditDetailTestId,
  administrativeAuditEventTestId,
} from "@cartulary/ui-contracts";
import { expect, test } from "./fixtures";
import { DeploymentAdministration } from "./pages/deploymentAdministration";
import {
  auditBrowserBarrier,
  auditBrowserEvent,
  auditBrowserEventId,
  auditBrowserPath,
  installAuditPresentation,
  openAdministrativeAudit,
} from "./support/administrativeAudit";
import { createDeploymentUser } from "./support/auth/deploymentUsers";
import { patchUser } from "./support/auth/sessions";
import { publicHttpOperation } from "./support/transport/publicHttpOperationClient";

test("administrative audit browses more than one hundred real exact matching events and timestamp boundaries", async ({
  workerAdminPage: page,
  workerAdmin,
  workerAdminRequest,
}, testInfo) => {
  test.setTimeout(120_000);
  let target = await createDeploymentUser(workerAdminRequest, {
    email: `audit-${randomUUID()}@example.test`,
    display_name: "Audit pagination fixture",
    initial_password: "AuditPaginationFixture!2026",
    is_deployment_admin: false,
    mfa_required: false,
  });
  try {
    for (let index = 0; index < 105; index++)
      target = await patchUser(workerAdminRequest, target.user_id, {
        base_user_version: target.user_version,
        display_name: `Audit pagination event ${index}`,
      });
    const query = {
      actor_user_id: workerAdmin.user_id,
      action_code: "user_profile_updated",
      target_kind: "user",
      target_id: target.user_id,
      limit: 100,
    };
    const first = await publicHttpOperation({
      operationID: "listAdministrativeAuditEvents",
      request: workerAdminRequest,
      query,
    });
    if (!first.ok) throw new Error(`Audit first page failed: ${first.status}`);
    expect(first.payload.data.audit_events).toHaveLength(100);
    const cursor = first.payload.meta.paging?.next_cursor;
    expect(cursor).toBeTruthy();
    const second = await publicHttpOperation({
      operationID: "listAdministrativeAuditEvents",
      request: workerAdminRequest,
      query: { ...query, cursor_token: cursor },
    });
    if (!second.ok)
      throw new Error(`Audit second page failed: ${second.status}`);
    expect(second.payload.data.audit_events).toHaveLength(5);
    expect(second.payload.meta.paging?.has_more).toBe(false);
    const events = [
      ...first.payload.data.audit_events,
      ...second.payload.data.audit_events,
    ];
    const newest = events[0];
    if (!newest) throw new Error("Missing generated event");
    const inclusive = await publicHttpOperation({
      operationID: "listAdministrativeAuditEvents",
      request: workerAdminRequest,
      query: { ...query, occurred_at_gte: newest.occurred_at },
    });
    const exclusive = await publicHttpOperation({
      operationID: "listAdministrativeAuditEvents",
      request: workerAdminRequest,
      query: { ...query, occurred_at_lt: newest.occurred_at },
    });
    if (!inclusive.ok || !exclusive.ok)
      throw new Error("Timestamp boundary read failed");
    expect(
      inclusive.payload.data.audit_events.map((event) => event.audit_event_id),
    ).toContain(newest.audit_event_id);
    expect(
      exclusive.payload.data.audit_events.map((event) => event.audit_event_id),
    ).not.toContain(newest.audit_event_id);
    const mismatch = await publicHttpOperation({
      operationID: "listAdministrativeAuditEvents",
      request: workerAdminRequest,
      query: { ...query, cursor_token: cursor, limit: 99 },
    });
    expect(mismatch.ok).toBe(false);
    expect(mismatch.status).toBe(400);

    const requests: URL[] = [];
    page.on("request", (request) => {
      if (new URL(request.url()).pathname.startsWith("/api/"))
        requests.push(new URL(request.url()));
    });
    const panel = await openAdministrativeAudit(page);
    await expect(panel.getByRole("status")).toContainText("Page 1:");
    await panel
      .getByLabel("Actor user ID", { exact: true })
      .fill(workerAdmin.user_id);
    await panel
      .getByLabel("Action code", { exact: true })
      .selectOption(query.action_code);
    await panel
      .getByLabel("Target kind", { exact: true })
      .selectOption(query.target_kind);
    await panel.getByLabel("Target ID", { exact: true }).fill(target.user_id);
    await panel.getByLabel("Target ID", { exact: true }).press("Enter");
    await expect(panel.getByRole("status")).toContainText(
      "Page 1: 100 events. More events available.",
    );
    for (const event of first.payload.data.audit_events)
      await expect(
        panel.getByTestId(administrativeAuditEventTestId(event.audit_event_id)),
      ).toHaveCount(1);
    const next = panel.getByRole("button", { name: "Next page", exact: true });
    await next.focus();
    await page.keyboard.press("Enter");
    await expect(panel.getByRole("status")).toContainText("Page 2: 5 events.");
    await expect(next).toBeDisabled();
    for (const event of second.payload.data.audit_events)
      await expect(
        panel.getByTestId(administrativeAuditEventTestId(event.audit_event_id)),
      ).toHaveCount(1);
    await expect(panel.getByRole("button", { name: /^Inspect / })).toHaveCount(
      5,
    );
    await panel
      .getByRole("button", { name: "Previous page", exact: true })
      .click();
    await expect(panel.getByRole("status")).toContainText(
      "Page 1: 100 events.",
    );
    await panel.getByLabel("Occurred at or after").fill(newest.occurred_at);
    await panel
      .getByRole("button", { name: "Apply filters", exact: true })
      .click();
    await expect(panel.getByRole("button", { name: /^Inspect / })).toHaveCount(
      inclusive.payload.data.audit_events.length,
    );
    const eventRow = panel.getByTestId(
      administrativeAuditEventTestId(newest.audit_event_id),
    );
    await eventRow.getByRole("button").click();
    await expect(
      panel.getByTestId(administrativeAuditDetailTestId(newest.audit_event_id)),
    ).toContainText("display_name");
    expect(
      requests.filter((url) => /membership|journal/u.test(url.pathname)),
    ).toEqual([]);
    expect(
      requests
        .filter((url) => url.pathname === auditBrowserPath)
        .every((url) => !url.searchParams.has("search")),
    ).toBe(true);
    await testInfo.attach("real-audit-pagination", {
      body: JSON.stringify({
        generated_matching_events: events.length,
        unique_ids: new Set(events.map((event) => event.audit_event_id)).size,
        page_sizes: [100, 5],
        inclusive_boundary: true,
        exclusive_boundary: true,
        cursor_mismatch_status: mismatch.status,
      }),
      contentType: "application/json",
    });
  } finally {
    await patchUser(workerAdminRequest, target.user_id, {
      base_user_version: target.user_version,
      is_active: false,
    });
  }
});

test("administrative audit fences delayed pages and errors with keyboard filter recovery and focus continuity", async ({
  workerAdminPage: page,
}) => {
  const fixture = await installAuditPresentation(page);
  fixture.setPage([auditBrowserEvent()], "opaque-page-two");
  const panel = await openAdministrativeAudit(page);
  await expect(panel.getByRole("status")).toContainText("Page 1:");
  const inspect = panel.getByRole("button", { name: /^Inspect / });
  await inspect.focus();
  await page.keyboard.press("Enter");
  await expect(
    panel.getByTestId(administrativeAuditDetailTestId(auditBrowserEventId)),
  ).toContainText("Redacted");
  const pending = auditBrowserBarrier();
  fixture.gateRead(pending.promise);
  fixture.setPage([
    auditBrowserEvent({
      audit_event_id: "00000000-0000-4000-8000-000000002000",
    }),
  ]);
  const readCount = fixture.requests.length;
  await panel.getByRole("button", { name: "Next page", exact: true }).click();
  await expect.poll(() => fixture.requests.length).toBe(readCount + 1);
  await expect(
    panel.getByRole("button", { name: "Next page", exact: true }),
  ).toBeDisabled();
  fixture.setPage([]);
  await panel.getByLabel("Target ID", { exact: true }).fill("exact-target");
  await panel.getByLabel("Target ID", { exact: true }).press("Enter");
  await expect(panel.getByLabel("Target kind", { exact: true })).toBeFocused();
  await expect(
    panel.getByLabel("Target kind", { exact: true }),
  ).toHaveAttribute("aria-invalid", "true");
  await panel.getByLabel("Target kind", { exact: true }).selectOption("user");
  await panel
    .getByRole("button", { name: "Apply filters", exact: true })
    .click();
  await expect(
    panel.getByText(/No deployment events match the applied filters/),
  ).toBeVisible();
  pending.release();
  await expect(panel.getByRole("button", { name: /^Inspect / })).toHaveCount(0);
  fixture.setPage([auditBrowserEvent()]);
  await panel
    .getByRole("button", { name: "Clear filters", exact: true })
    .click();
  await expect(panel.getByRole("status")).toContainText("Page 1: 1 event");
  await panel.getByRole("button", { name: /^Inspect / }).click();
  const replacing = auditBrowserBarrier();
  fixture.setPage([]);
  fixture.gateRead(replacing.promise);
  await panel.getByRole("button", { name: "Refresh", exact: true }).click();
  await panel.getByRole("button", { name: /^Hide / }).focus();
  replacing.release();
  await expect(
    panel.getByRole("button", { name: "Refresh", exact: true }),
  ).toBeFocused();
  const staleError = auditBrowserBarrier();
  fixture.fail();
  fixture.gateRead(staleError.promise);
  await panel.getByRole("button", { name: "Refresh", exact: true }).click();
  await expect.poll(() => fixture.requests.length).toBe(readCount + 5);
  await new DeploymentAdministration(page).selectPanel("deployment-users");
  const hiddenReads = fixture.requests.length;
  staleError.release();
  fixture.setPage([auditBrowserEvent()]);
  await new DeploymentAdministration(page).selectPanel("administrative-audit");
  await expect(panel.getByRole("status")).toContainText("Page 1: 1 event");
  expect(fixture.requests.length).toBe(hiddenReads + 1);
  fixture.fail("invalid_pagination_request", "cursor_query_mismatch");
  await panel.getByRole("button", { name: "Refresh", exact: true }).click();
  const reload = panel.getByRole("button", {
    name: "Reload first page",
    exact: true,
  });
  await expect(reload).toBeVisible();
  const rejectedReads = fixture.requests.length;
  const accessReads = fixture.accessReads;
  await new DeploymentAdministration(page).selectPanel("deployment-users");
  await new DeploymentAdministration(page).selectPanel("administrative-audit");
  await expect.poll(() => fixture.accessReads).toBe(accessReads + 1);
  await expect(panel.getByRole("status")).toContainText(
    "continuation is no longer usable",
  );
  expect(fixture.requests.length).toBe(rejectedReads);
  fixture.setPage([auditBrowserEvent()]);
  await reload.focus();
  await page.keyboard.press("Enter");
  await expect(panel.getByRole("status")).toContainText("Page 1: 1 event");
  await expect(panel).not.toContainText("Server diagnostic");
  fixture.setPage([
    auditBrowserEvent({
      action_code: "constructor",
      target_kind: "__proto__",
      target_id: "published_".repeat(20),
    }),
  ]);
  await panel.getByRole("button", { name: "Refresh", exact: true }).click();
  await panel.getByRole("button", { name: /^Inspect constructor / }).click();
  const details = panel.getByRole("region", {
    name: /^Details for constructor /,
  });
  for (const value of ['""', "false", "0", "null", "Redacted"])
    await expect(
      details.getByRole("cell", { name: value, exact: true }).first(),
    ).toBeVisible();
  await expect(details).toContainText('"<script>inert text</script>"');
  await expect(details.locator("script")).toHaveCount(0);
  await expect(
    panel
      .getByLabel("Action code", { exact: true })
      .getByRole("option", { name: "constructor", exact: true }),
  ).toHaveCount(0);
});
