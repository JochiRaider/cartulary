import type { CreateViewRowResponse } from "@cartulary/protocol-ts/http";
import {
  authTestId,
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  gridShellTestId,
  incidentLandingTestId,
  surfaceTabTestId,
  workbookInspectorCloseButtonTestId,
  workbookInspectorToggleTestId,
  workbookSurfacesMenuOptionTestId,
  workbookSurfacesMenuTriggerTestId,
} from "@cartulary/ui-contracts";
import {
  evidenceViewSchemaId,
  hostsViewSchemaId,
  notesViewSchemaId,
  requireViewContract,
} from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { createIncidentMemberUser } from "./support/incidents/memberships";
import { uniqueEmail, uniqueTxn } from "./support/runtime/fixtureIdentity";
import { installIncidentSocketMonitor } from "./support/transport/incidentSocket";
import { fetchFullRecordHistory } from "./support/workbook/history";
import {
  noteSourceFields,
  openNoteFixture,
} from "./support/workbook/noteCreate";
import { createViewRow, queryViewRows } from "./support/workbook/query";

test("Incident revocation conceals retained Note authoring and late atomic acceptance without ending the account session", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
  let memberId = "";
  const f = await openNoteFixture(page, evidenceViewSchemaId, async (url) => {
    const incident = new URL(url, "http://fixture").searchParams.get(
      "incident_id",
    );
    if (!incident) throw new Error("Missing incident fixture");
    const member = await createIncidentMemberUser(page, incident, {
      email: uniqueEmail("linked-note-member"),
      display_name: "Note author",
      initial_password: "NoteMember1!",
      role: "editor",
      is_deployment_admin: false,
      mfa_required: false,
    });
    memberId = member.user_id;
    await sessionTracker.loginTrackedUser(page, {
      createdBy: "linked-note-revocation",
      email: member.email,
      password: member.initial_password,
      purpose: "retained Note incident revocation",
      userId: memberId,
    });
    const sockets = installIncidentSocketMonitor(page, incident);
    await page.goto(url);
    await sockets.waitForAcceptedSocket();
  });
  await f.form
    .getByRole("textbox", { name: "Title", exact: true })
    .fill("Protected unfinished Note");
  let committed = false;
  let release: () => void = () => {};
  const delayed = new Promise<void>((resolve) => {
    release = resolve;
  });
  await page.route(
    `**/records/${f.source.record_id}/linked-notes`,
    async (route) => {
      const response = await route.fetch();
      expect(response.ok()).toBe(true);
      committed = true;
      await delayed;
      await route.fulfill({ response });
    },
  );
  await f.form
    .getByRole("button", { name: "Create Note", exact: true })
    .click();
  await expect.poll(() => committed).toBe(true);
  const memberships = await workerAdminRequest.get(
    `/api/v1/incidents/${f.incident}/memberships`,
  );
  expect(memberships.ok()).toBe(true);
  const body = (await memberships.json()) as {
    data: { memberships: { user_id: string; membership_version: number }[] };
  };
  const membership = body.data.memberships.find(
    (entry) => entry.user_id === memberId,
  );
  if (!membership) throw new Error("Missing Note author membership");
  const revoked = await workerAdminRequest.delete(
    `/api/v1/incidents/${f.incident}/memberships/${memberId}`,
    {
      data: { base_membership_version: membership.membership_version },
    },
  );
  expect(revoked.status()).toBe(204);
  await expect(page.getByTestId(incidentLandingTestId("shell"))).toBeVisible();
  await expect(f.form).toHaveCount(0);
  await expect(
    page.getByText("Protected unfinished Note", { exact: true }),
  ).toHaveCount(0);
  const late = page.waitForResponse((response) =>
    response.url().endsWith(`/records/${f.source.record_id}/linked-notes`),
  );
  release();
  await late;
  await expect(page.getByTestId(authTestId("shell"))).toHaveCount(0);
  await expect(page.getByTestId(incidentLandingTestId("shell"))).toBeVisible();
  await expect(
    page.getByRole("region", { name: "Retained Note authoring", exact: true }),
  ).toHaveCount(0);
  await expect(page).not.toHaveURL(/incident_id=/u);
});

test("Every contextual Note action creates one atomic source association with title-only or body-only input", async ({
  page,
}) => {
  for (const [index, view] of Object.keys(noteSourceFields).entries()) {
    const f = await openNoteFixture(page, view);
    await expect(
      f.form.getByRole("textbox", { name: "Title", exact: true }),
    ).toBeFocused();
    const before = await fetchFullRecordHistory(page, f.source.record_id);
    await f.form
      .getByTestId(genericCreateSubmitTestId(notesViewSchemaId))
      .click();
    await expect(f.form.getByRole("alert")).toContainText("title or body");
    const field = index % 2 ? "note.body" : "note.title";
    await f.form
      .getByTestId(genericCreateFieldTestId(field))
      .fill(`  Note from ${view}  `);
    await f.form
      .getByTestId(genericCreateFieldTestId("note.tags"))
      .fill(" Investigation ");
    const response = page.waitForResponse(
      (result) =>
        result.request().method() === "POST" &&
        result.url().endsWith(`/records/${f.source.record_id}/linked-notes`),
    );
    await f.form
      .getByTestId(genericCreateSubmitTestId(notesViewSchemaId))
      .press("Enter");
    const accepted = await response;
    expect(accepted.ok()).toBe(true);
    const receipt = (await accepted.json()) as CreateViewRowResponse;
    await expect(f.form).toHaveCount(0);
    await expect(f.action).toBeFocused();
    const notes = await queryViewRows(page, f.incident, notesViewSchemaId);
    expect(notes).toHaveLength(1);
    expect(notes[0]?.record_id).toBe(receipt.data.row.record_id);
    expect(notes[0]?.cells["note.linked_record_count"]?.value).toBe(1);
    expect(notes[0]?.cells[field]?.value).toBe(`Note from ${view}`);
    const after = await fetchFullRecordHistory(page, f.source.record_id);
    const additions = after.items.filter(
      (item) =>
        !before.items.some(
          (prior) => prior.history_item_ref === item.history_item_ref,
        ),
    );
    expect(additions.length).toBeGreaterThan(0);
    expect(new Set(additions.map((item) => item.change_set_id))).toEqual(
      new Set([receipt.data.change_set_id]),
    );
  }
});

test("Note source replacement and clearing preserve authoring across sheet navigation and picker cancellation", async ({
  page,
}) => {
  const f = await openNoteFixture(page, evidenceViewSchemaId);
  const replacement = await createViewRow(page, f.incident, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("replacement"),
    "host.display_name": "Chosen Host",
  });
  await f.form
    .getByTestId(genericCreateFieldTestId("note.title"))
    .fill("Retained source edit");
  const choose = f.form.getByRole("button", {
    name: "Choose source",
    exact: true,
  });
  await choose.click();
  await page
    .getByRole("combobox", { name: "Source sheet", exact: true })
    .selectOption(hostsViewSchemaId);
  const picker = page.getByRole("region", {
    name: "Choose Note source",
    exact: true,
  });
  await expect(
    picker.getByRole("option", { name: "Chosen Host", exact: true }),
  ).toBeAttached();
  await picker
    .getByRole("combobox", { name: "Note source", exact: true })
    .selectOption(replacement.record_id);
  await picker.press("Escape");
  await expect(choose).toBeFocused();
  await expect(f.form).toContainText("Reviewed investigation source");
  await choose.click();
  await page
    .getByRole("combobox", { name: "Source sheet", exact: true })
    .selectOption(hostsViewSchemaId);
  await expect(
    picker.getByRole("option", { name: "Chosen Host", exact: true }),
  ).toBeAttached();
  await picker
    .getByRole("combobox", { name: "Note source", exact: true })
    .selectOption(replacement.record_id);
  await picker
    .getByRole("button", { name: "Apply source", exact: true })
    .click();
  await expect(f.form).toContainText("Chosen Host");
  await page
    .getByTestId(workbookInspectorCloseButtonTestId(evidenceViewSchemaId))
    .click();
  await switchSheet(page, notesViewSchemaId);
  await expect(
    page.getByTestId(genericCreateFieldTestId("note.title")),
  ).toHaveValue("Retained source edit");
  const response = page.waitForResponse(
    (result) =>
      result.request().method() === "POST" &&
      result.url().endsWith(`/records/${replacement.record_id}/linked-notes`),
  );
  await page.getByTestId(genericCreateSubmitTestId(notesViewSchemaId)).click();
  expect((await response).ok()).toBe(true);
  await expect(
    page.getByTestId(genericCreateFieldTestId("note.title")),
  ).toHaveValue("");
  // Explicit clearing chooses ordinary creation while preserving the authored body.
  if (
    !(await page
      .getByRole("button", { name: "Choose source", exact: true })
      .count())
  )
    await page
      .getByTestId(workbookInspectorToggleTestId(notesViewSchemaId))
      .click();
  await page
    .getByRole("button", { name: "Choose source", exact: true })
    .click();
  await page
    .getByRole("combobox", { name: "Source sheet", exact: true })
    .selectOption(hostsViewSchemaId);
  await expect(
    page.getByRole("option", { name: "Chosen Host", exact: true }),
  ).toBeAttached();
  await page
    .getByRole("combobox", { name: "Note source", exact: true })
    .selectOption(replacement.record_id);
  await page.getByRole("button", { name: "Apply source", exact: true }).click();
  await page
    .getByTestId(genericCreateFieldTestId("note.body"))
    .fill("Ordinary body-only Note");
  await page.getByRole("button", { name: "Clear source", exact: true }).click();
  await expect(
    page.getByTestId(genericCreateFieldTestId("note.body")),
  ).toHaveValue("Ordinary body-only Note");
  const ordinary = page.waitForResponse(
    (result) =>
      result.request().method() === "POST" &&
      result.url().endsWith(`/views/${notesViewSchemaId}/rows`),
  );
  await page.getByTestId(genericCreateSubmitTestId(notesViewSchemaId)).click();
  expect((await ordinary).ok()).toBe(true);
  const notes = await queryViewRows(page, f.incident, notesViewSchemaId);
  expect(notes).toHaveLength(2);
  expect(
    notes.map((row) => row.cells["note.linked_record_count"]?.value).sort(),
  ).toEqual([0, 1]);
});

test("Note response loss after commit replays exact bytes after navigation and accepted refresh recovery sends reads only", async ({
  page,
}) => {
  const f = await openNoteFixture(page, hostsViewSchemaId);
  const requests: string[] = [],
    receipts: CreateViewRowResponse[] = [];
  let failReads = false;
  await page.route(`**/views/${notesViewSchemaId}/query`, async (route) => {
    if (failReads) await route.abort("failed");
    else await route.continue();
  });
  await page.route(
    `**/records/${f.source.record_id}/linked-notes`,
    async (route) => {
      requests.push(route.request().postData() ?? "");
      const response = await route.fetch();
      expect(response.ok()).toBe(true);
      receipts.push(await response.json());
      if (requests.length === 1) await route.abort("failed");
      else {
        failReads = true;
        await route.fulfill({ response });
      }
    },
  );
  await f.form
    .getByTestId(genericCreateFieldTestId("note.title"))
    .fill("Recover exactly once");
  await f.form
    .getByTestId(genericCreateSubmitTestId(notesViewSchemaId))
    .click();
  await expect(
    page.locator("summary").filter({ hasText: /^Note recovery$/ }),
  ).toBeVisible();
  await expect(
    f.form.getByRole("textbox", { name: "Title", exact: true }),
  ).toBeDisabled();
  await page
    .getByTestId(workbookInspectorCloseButtonTestId(hostsViewSchemaId))
    .click();
  await switchSheet(page, evidenceViewSchemaId);
  await page
    .locator("summary")
    .filter({ hasText: /^Note recovery$/ })
    .click();
  const recovery = page.getByRole("region", {
    name: "Retained Note authoring",
    exact: true,
  });
  await recovery
    .getByRole("button", { name: "Recover submission", exact: true })
    .click();
  await expect(
    recovery.getByText(/Note created, but views need refresh/),
  ).toBeVisible();
  expect(requests).toHaveLength(2);
  expect(requests[1]).toBe(requests[0]);
  expect(receipts[1]?.data).toEqual(receipts[0]?.data);
  failReads = false;
  await recovery
    .getByRole("button", { name: "Retry refresh", exact: true })
    .click();
  await expect(recovery).toHaveCount(0);
  await expect(
    page.getByRole("region", {
      name: "Active workbook surface focus target",
      exact: true,
    }),
  ).toBeFocused();
  expect(requests).toHaveLength(2);
  const notes = await queryViewRows(page, f.incident, notesViewSchemaId);
  expect(notes).toHaveLength(1);
  expect(notes[0]?.cells["note.linked_record_count"]?.value).toBe(1);
});

async function switchSheet(page: Page, view: string) {
  const tab = page.getByTestId(surfaceTabTestId(view));
  if (await tab.count()) await tab.click();
  else {
    await page.getByTestId(workbookSurfacesMenuTriggerTestId()).click();
    await page
      .getByTestId(
        workbookSurfacesMenuOptionTestId(
          requireViewContract(view).viewSchemaId,
        ),
      )
      .click();
  }
  await expect(page.getByTestId(gridShellTestId(view))).toBeVisible();
}
