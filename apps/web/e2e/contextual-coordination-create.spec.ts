import type { CreateViewRowResponse } from "@cartulary/protocol-ts/http";
import {
  authTestId,
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  gridShellTestId,
  incidentLandingTestId,
  surfaceTabTestId,
  workbookInspectorCloseButtonTestId,
} from "@cartulary/ui-contracts";
import {
  lessonViewSchemaId,
  statusReviewViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { createIncidentMemberUser } from "./support/incidents/memberships";
import { uniqueEmail, uniqueTxn } from "./support/runtime/fixtureIdentity";
import { installIncidentSocketMonitor } from "./support/transport/incidentSocket";
import {
  coordinationMatrix,
  fillCoordinationMinimum,
  openCoordinationFixture,
  retainCoordinationUncertainResult,
} from "./support/workbook/coordinationCreate";
import { fetchFullRecordHistory } from "./support/workbook/history";
import { createViewRow, queryViewRows } from "./support/workbook/query";
import { openRecoveryItem, recoveryEntry } from "./support/workbook/recovery";

test("All twelve contextual coordination actions create one owner-defined source link without semantic seeds", async ({
  page,
}) => {
  for (const [view, variants] of Object.entries(coordinationMatrix))
    for (const variant of variants) {
      const f = await openCoordinationFixture(page, variant, view);
      const before = await fetchFullRecordHistory(page, f.source.record_id);
      await f.form
        .getByTestId(genericCreateSubmitTestId(f.target.viewSchemaId))
        .press("Enter");
      await expect(f.form.getByRole("alert").first()).toBeVisible();
      await fillCoordinationMinimum(f);
      const response = page.waitForResponse(
        (r) =>
          r.request().method() === "POST" &&
          r.url().endsWith(`/views/${f.target.viewSchemaId}/rows`),
      );
      await f.form
        .getByTestId(genericCreateSubmitTestId(f.target.viewSchemaId))
        .press("Enter");
      const accepted = await response;
      expect(accepted.ok()).toBe(true);
      const receipt = (await accepted.json()) as CreateViewRowResponse;
      expect(receipt.data).toMatchObject({
        view_schema_id: f.target.viewSchemaId,
        source_record_id: f.source.record_id,
        link_type: "references_artifact",
      });
      expect(receipt.meta.request_id).toBe(accepted.headers()["x-request-id"]);
      await expect(f.form).toHaveCount(0);
      await expect(f.action).toBeFocused();
      const rows = await queryViewRows(page, f.incident, f.target.viewSchemaId);
      expect(rows).toHaveLength(1);
      expect(rows[0]?.record_id).toBe(receipt.data.row.record_id);
      for (const field of f.target.fields.filter(
        (field) => field.readKind === "collection",
      )) {
        const cell = rows[0]?.cells[field.fieldKey];
        expect(cell?.value).toMatchObject({
          kind: "collection_value_v1",
          items: [],
        });
      }
      const after = await fetchFullRecordHistory(page, f.source.record_id);
      const added = after.items.filter(
        (item) =>
          !before.items.some(
            (prior) => prior.history_item_ref === item.history_item_ref,
          ),
      );
      expect(added.length).toBeGreaterThan(0);
      expect(new Set(added.map((item) => item.change_set_id))).toEqual(
        new Set([receipt.data.change_set_id]),
      );
      const sources = await queryViewRows(page, f.incident, view);
      expect(
        sources.find((row) => row.record_id === f.source.record_id)
          ?.row_version,
      ).toBe(f.source.row_version);
    }
});

test("Coordination drafts survive navigation and source replacement or clear preserves every authored value", async ({
  page,
}) => {
  const f = await openCoordinationFixture(
    page,
    "lesson",
    taskRequestsViewSchemaId,
  );
  const replacement = await createViewRow(
    page,
    f.incident,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("replacement"),
      "timeline.activity_synopsis_text": "Chosen coordination source",
    },
  );
  await fillCoordinationMinimum(f);
  await f.form
    .getByRole("button", { name: "Choose source", exact: true })
    .click();
  const picker = f.form.getByRole("region", {
    name: "Choose source",
    exact: true,
  });
  await picker
    .getByRole("combobox", { name: "Reference surface", exact: true })
    .selectOption(timelineViewSchemaId);
  await expect(
    picker.getByRole("option", {
      name: "Chosen coordination source",
      exact: true,
    }),
  ).toBeAttached();
  await picker
    .getByRole("combobox", { name: "Source", exact: true })
    .selectOption(replacement.record_id);
  await picker.press("Escape");
  await expect(
    f.form.getByRole("button", { name: "Choose source", exact: true }),
  ).toBeFocused();
  await f.form
    .getByRole("button", { name: "Choose source", exact: true })
    .click();
  await picker
    .getByRole("combobox", { name: "Reference surface", exact: true })
    .selectOption(timelineViewSchemaId);
  await expect(
    picker.getByRole("option", {
      name: "Chosen coordination source",
      exact: true,
    }),
  ).toBeAttached();
  await picker
    .getByRole("combobox", { name: "Source", exact: true })
    .selectOption(replacement.record_id);
  await picker
    .getByRole("button", { name: "Apply references", exact: true })
    .click();
  await expect(f.form).toContainText("Chosen coordination source");
  await page.getByTestId(workbookInspectorCloseButtonTestId(f.view)).click();
  await switchSheet(page, statusReviewViewSchemaId);
  await openRecoveryItem(page, /^Coordination draft ·/);
  const recovery = page.getByRole("region", {
    name: "Retained Coordination authoring",
    exact: true,
  });
  await recovery
    .getByRole("button", { name: "Resume Coordination draft", exact: true })
    .click();
  await expect(
    recovery.getByTestId(genericCreateFieldTestId("lesson.summary")),
  ).toHaveValue("  Authored coordination summary  ");
  await expect(recovery).toContainText("Chosen coordination source");
  await recovery
    .getByRole("button", { name: "Clear source", exact: true })
    .click();
  await expect(recovery).toContainText("No source link will be saved.");
  const response = page.waitForResponse(
    (r) =>
      r.request().method() === "POST" &&
      r.url().endsWith(`/views/${lessonViewSchemaId}/rows`),
  );
  await recovery
    .getByTestId(genericCreateSubmitTestId(lessonViewSchemaId))
    .click();
  const receipt = (await (await response).json()) as CreateViewRowResponse;
  expect(receipt.data).not.toHaveProperty("source_record_id");
  expect(receipt.data.row.cells["lesson.summary"]?.value).toBe(
    "Authored coordination summary",
  );
});

test("Coordination response loss after server commit recovers the exact request without another artifact", async ({
  page,
}) => {
  const f = await openCoordinationFixture(page);
  await fillCoordinationMinimum(f);
  const bodies: string[] = [];
  page.on("request", (request) => {
    if (
      request.method() === "POST" &&
      request.url().endsWith(`/views/${f.target.viewSchemaId}/rows`)
    )
      bodies.push(request.postData() ?? "");
  });
  const { path, recovery } = await retainCoordinationUncertainResult(page, f);
  expect(
    await queryViewRows(page, f.incident, f.target.viewSchemaId),
  ).toHaveLength(1);
  await page.unroute(path);
  const response = page.waitForResponse(
    (r) =>
      r.request().method() === "POST" &&
      r.url().endsWith(`/views/${f.target.viewSchemaId}/rows`),
  );
  await recovery
    .getByRole("button", { name: "Recover submission", exact: true })
    .press("Enter");
  const replay = await response;
  expect(replay.status()).toBe(200);
  expect(bodies).toHaveLength(2);
  expect(bodies[1]).toBe(bodies[0]);
  expect(
    await queryViewRows(page, f.incident, f.target.viewSchemaId),
  ).toHaveLength(1);
  await expect(recoveryEntry(page)).toHaveCount(0);
});

test("Accepted coordination refresh recovery sends reads only and keeps the source selection", async ({
  page,
}) => {
  const f = await openCoordinationFixture(page);
  await fillCoordinationMinimum(f);
  let committed = false,
    creates = 0;
  const query = `**/incidents/${f.incident}/views/${f.target.viewSchemaId}/query`;
  await page.route(query, async (route) => {
    if (committed) await route.abort("failed");
    else await route.continue();
  });
  await page.route(
    `**/incidents/${f.incident}/views/${f.target.viewSchemaId}/rows`,
    async (route) => {
      creates++;
      const response = await route.fetch();
      expect(response.ok()).toBe(true);
      committed = true;
      await route.fulfill({ response });
    },
  );
  await f.form
    .getByTestId(genericCreateSubmitTestId(f.target.viewSchemaId))
    .click();
  await openRecoveryItem(page, /^Coordination creation ·/);
  const recovery = page.getByRole("region", {
    name: "Retained Coordination authoring",
    exact: true,
  });
  await expect(recovery).toContainText("views need refresh");
  await page.unroute(query);
  await recovery
    .getByRole("button", { name: "Retry refresh", exact: true })
    .press("Enter");
  await expect(recoveryEntry(page)).toHaveText("Recovery (0)");
  expect(creates).toBe(1);
  await expect(page.getByTestId(gridShellTestId(f.view))).toBeVisible();
  expect(
    await queryViewRows(page, f.incident, f.target.viewSchemaId),
  ).toHaveLength(1);
});

test("Incident revocation conceals coordination drafts and late acceptance while preserving the account session", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
  let memberId = "";
  const f = await openCoordinationFixture(
    page,
    "lesson",
    taskRequestsViewSchemaId,
    async (url) => {
      const incident = new URL(url, "http://fixture").searchParams.get(
        "incident_id",
      );
      if (!incident) throw new Error("Missing incident");
      const member = await createIncidentMemberUser(page, incident, {
        email: uniqueEmail("coordination-revoked"),
        display_name: "Coordination author",
        initial_password: "CoordinationMember1!",
        role: "editor",
        is_deployment_admin: false,
        mfa_required: false,
      });
      memberId = member.user_id;
      await sessionTracker.loginTrackedUser(page, {
        createdBy: "coordination-revocation",
        email: member.email,
        password: member.initial_password,
        purpose: "retained coordination revocation",
        userId: memberId,
      });
      const sockets = installIncidentSocketMonitor(page, incident);
      await page.goto(url);
      await sockets.waitForAcceptedSocket();
    },
  );
  await fillCoordinationMinimum(f);
  let committed = false;
  let release: () => void = () => {};
  const delayed = new Promise<void>((resolve) => {
    release = resolve;
  });
  const suffix = `/views/${f.target.viewSchemaId}/rows`;
  await page.route(`**/incidents/${f.incident}${suffix}`, async (route) => {
    const response = await route.fetch();
    expect(response.ok()).toBe(true);
    committed = true;
    await delayed;
    await route.fulfill({ response });
  });
  await f.form
    .getByTestId(genericCreateSubmitTestId(f.target.viewSchemaId))
    .click();
  await expect.poll(() => committed).toBe(true);
  const response = await workerAdminRequest.get(
    `/api/v1/incidents/${f.incident}/memberships`,
  );
  const members = (await response.json()) as {
    data: { memberships: { user_id: string; membership_version: number }[] };
  };
  const member = members.data.memberships.find(
    (item) => item.user_id === memberId,
  );
  if (!member) throw new Error("Missing membership");
  const revoked = await workerAdminRequest.delete(
    `/api/v1/incidents/${f.incident}/memberships/${memberId}`,
    { data: { base_membership_version: member.membership_version } },
  );
  expect(revoked.status()).toBe(204);
  await expect(page.getByTestId(incidentLandingTestId("shell"))).toBeVisible();
  await expect(f.form).toHaveCount(0);
  const late = page.waitForResponse((r) => r.url().endsWith(suffix));
  release();
  await late;
  await expect(page.getByTestId(authTestId("shell"))).toHaveCount(0);
  await expect(
    page.getByRole("region", {
      name: "Retained Coordination authoring",
      exact: true,
    }),
  ).toHaveCount(0);
  await expect(page).not.toHaveURL(/incident_id=/u);
});
async function switchSheet(page: Page, view: string) {
  const tab = page.getByTestId(surfaceTabTestId(view));
  if (await tab.count()) await tab.click();
  else {
    await page
      .getByRole("button", { name: "System views", exact: true })
      .click();
    await page
      .locator(`[role="menuitemradio"][data-view-schema-id="${view}"]`)
      .click();
  }
  await expect(page.getByTestId(gridShellTestId(view))).toBeVisible();
}
