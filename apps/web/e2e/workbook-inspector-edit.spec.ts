import { scrollGridTargetIntoView } from "@cartulary/test-utils/grid";
import {
  entityInspectorTestId,
  genericEditSubmitTestId,
  genericEditValueTestId,
  gridShellTestId,
  rowCellTestId,
  workbookFocusAnchorTestId,
  workbookInspectorCloseButtonTestId,
  workbookInspectorPanelTestId,
  workbookInspectorToggleTestId,
} from "@cartulary/ui-contracts";
import {
  evidenceViewSchemaId,
  hostsViewSchemaId,
  indicatorsViewSchemaId,
  notesViewSchemaId,
  partiesViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { csrfHeaders } from "./support/auth/browserSession";
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
import { fetchRecordHistoryCount } from "./support/workbook/history";
import { switchOrdinarySheet } from "./support/workbook/ordinaryCreate";
import {
  createViewRow,
  patchRecord,
  queryViewRows,
} from "./support/workbook/query";
import {
  activateCommittedGridCell,
  openGenericInspectorForRecord,
} from "./support/workbook/rowMutations";

async function fixture(page: Page, view: string = hostsViewSchemaId) {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("IER"),
    "Inspector editing recovery",
  );
  const field =
    view === hostsViewSchemaId ? "host.display_name" : "evidence.title";
  const first = await createViewRow(page, incident, view, {
    client_txn_id: uniqueTxn("inspector-a"),
    [field]: "Inspected A",
  });
  const second = await createViewRow(page, incident, view, {
    client_txn_id: uniqueTxn("inspector-b"),
    [field]: "Inspected B",
  });
  await page.goto(`/?incident_id=${incident}&view_schema_id=${view}`);
  await expect(page.getByTestId(gridShellTestId(view))).toBeVisible();
  await openGenericInspectorForRecord(page, view, first.record_id);
  return { incident, view, field, first, second };
}
async function editField(page: Page, view: string, field: string) {
  await page.locator(`[data-inspector-edit-field="${field}"]`).click();
  return page.getByTestId(genericEditValueTestId(view));
}

test("Inspector edits bind the selected record and retain dirty fields through saved changes and explicit return", async ({
  page,
}) => {
  const f = await fixture(page),
    input = await editField(page, f.view, "host.location");
  await expect(page.getByRole("combobox", { name: "Edit record" })).toHaveCount(
    0,
  );
  await expect(
    page.locator('[data-inspector-edit-field="host.fqdn"]'),
  ).toHaveCount(0);
  await input.fill("  unfinished location  ");
  await patchRecord(page, f.first.record_id, {
    view_schema_id: f.view,
    base_row_version: 1,
    client_txn_id: uniqueTxn("unrelated"),
    changes: [{ field_key: "host.display_name", value: "Inspected A renamed" }],
  });
  await expect(page.getByTestId(entityInspectorTestId("host"))).toContainText(
    "Inspected A renamed",
  );
  await expect(input).toHaveValue("  unfinished location  ");
  await expect(page.getByTestId(genericEditSubmitTestId(f.view))).toBeEnabled();
  await editField(page, f.view, "host.business_owner");
  await editField(page, f.view, "host.location");
  await expect(
    page.getByTestId(genericEditSubmitTestId(f.view)),
  ).toBeDisabled();
  await page.getByRole("button", { name: "Resume draft", exact: true }).click();
  await openGenericInspectorForRecord(page, f.view, f.second.record_id);
  await editField(page, f.view, "host.location");
  await expect(input).toHaveValue("");
  await openGenericInspectorForRecord(page, f.view, f.first.record_id);
  await editField(page, f.view, "host.location");
  await page.getByRole("button", { name: "Resume draft", exact: true }).click();
  await patchRecord(page, f.first.record_id, {
    view_schema_id: f.view,
    base_row_version: 2,
    client_txn_id: uniqueTxn("same-field"),
    changes: [{ field_key: "host.location", value: "Concurrent location" }],
  });
  await expect(
    page.getByRole("button", { name: "Keep draft Location", exact: true }),
  ).toBeVisible();
  await expect(input).toHaveValue("  unfinished location  ");
  await page
    .getByRole("button", { name: "Keep draft Location", exact: true })
    .click();
  const sent = page.waitForResponse(
    (response) =>
      response.request().method() === "PATCH" &&
      response.url().endsWith(`/records/${f.first.record_id}`),
  );
  await page.getByTestId(genericEditSubmitTestId(f.view)).click();
  expect((await sent).ok()).toBeTruthy();
  await expect(
    page.getByRole("region", { name: "Inspector changes" }),
  ).toContainText("Saved, version 4");
  const rows = await queryViewRows(page, f.incident, f.view);
  expect(
    rows.find((row) => row.record_id === f.first.record_id)?.cells[
      "host.location"
    ]?.value,
  ).toBe("unfinished location");
  expect(
    rows.find((row) => row.record_id === f.second.record_id)?.row_version,
  ).toBe(1);
});

test("Inspector uncertain recovery replays exact requests without consuming newer authoring", async ({
  page,
}) => {
  for (const [view, malformed] of [
    [hostsViewSchemaId, false],
    [evidenceViewSchemaId, true],
  ] as const) {
    const f = await fixture(page, view),
      input = await editField(page, view, f.field),
      bodies: string[] = [];
    await page.route(
      `**/api/v1/records/${f.first.record_id}`,
      async (route) => {
        if (route.request().method() !== "PATCH") {
          await route.continue();
          return;
        }
        bodies.push(route.request().postData() ?? "");
        const response = await route.fetch();
        expect(response.ok()).toBeTruthy();
        if (bodies.length > 1) await route.fulfill({ response });
        else if (malformed)
          await route.fulfill({
            response,
            body: JSON.stringify({
              data: { row: {} },
              meta: { request_id: "invalid-inspector-receipt" },
            }),
          });
        else await route.abort("failed");
      },
    );
    await input.fill("Captured inspector value");
    await page.getByTestId(genericEditSubmitTestId(view)).click();
    const retry = page.getByRole("button", {
      name: "Retry original change",
      exact: true,
    });
    await expect(retry).toBeVisible();
    await input.fill("Newer unfinished value");
    await page.getByTestId(workbookInspectorCloseButtonTestId(view)).click();
    await retry.focus();
    await retry.press("Enter");
    await expect(
      page.getByRole("region", { name: "Inspector changes" }),
    ).toContainText("Saved, version 2");
    expect(bodies).toHaveLength(2);
    expect(bodies[1]).toBe(bodies[0]);
    expect(await fetchRecordHistoryCount(page, f.first.record_id)).toBe(2);
    await expect(
      page.getByTestId(workbookInspectorCloseButtonTestId(view)),
    ).toHaveCount(0);
    await switchOrdinarySheet(page, timelineViewSchemaId);
    await switchOrdinarySheet(page, view);
    await openGenericInspectorForRecord(page, view, f.first.record_id);
    await editField(page, view, f.field);
    await page
      .getByRole("button", { name: "Resume draft", exact: true })
      .click();
    await expect(input).toHaveValue("Newer unfinished value");
  }
});

test("Inspector acknowledged refresh recovery preserves newer grid focus and never dispatches another patch", async ({
  page,
}) => {
  const f = await fixture(page),
    input = await editField(page, f.view, "host.location");
  let release: () => void = () => {},
    committed = false,
    failRefresh = true,
    count = 0;
  await page.route(`**/api/v1/records/${f.first.record_id}`, async (route) => {
    if (route.request().method() !== "PATCH") {
      await route.continue();
      return;
    }
    count++;
    const response = await route.fetch();
    expect(response.ok()).toBeTruthy();
    committed = true;
    await new Promise<void>((resolve) => {
      release = resolve;
    });
    await route.fulfill({ response });
  });
  await page.route(`**/views/${f.view}/query`, async (route) => {
    if (committed && failRefresh) await route.abort("failed");
    else await route.continue();
  });
  await input.fill("Accepted location");
  await page.getByTestId(genericEditSubmitTestId(f.view)).click();
  await expect.poll(() => committed).toBeTruthy();
  await page.getByTestId(workbookInspectorCloseButtonTestId(f.view)).click();
  const id = rowCellTestId(f.second.record_id, "host.display_name");
  await scrollGridTargetIntoView({ page, surface: f.view, targetTestId: id });
  const cell = page
    .getByTestId(id)
    .locator("xpath=ancestor::*[@role='gridcell'][1]");
  await activateCommittedGridCell(cell);
  release();
  const recovery = page.getByRole("button", {
    name: "Refresh saved change",
    exact: true,
  });
  await expect(recovery).toBeVisible();
  await expect(cell).toBeFocused();
  await expect(page.getByTestId(workbookFocusAnchorTestId())).toHaveText(
    `${f.view}:${f.second.record_id}:host.display_name`,
  );
  await expect(
    page.getByTestId(workbookInspectorCloseButtonTestId(f.view)),
  ).toHaveCount(0);
  failRefresh = false;
  await recovery.click();
  await expect(recovery).toHaveCount(0);
  expect(count).toBe(1);
  expect(await fetchRecordHistoryCount(page, f.first.record_id)).toBe(2);
});

test("a11y.inspector retained editing and recovery remain named keyboard reachable and bounded in narrow layouts", async ({
  page,
}, info) => {
  const f = await fixture(page),
    input = await editField(page, f.view, "host.location");
  await input.fill("Keyboard retained location");
  await page.getByTestId(workbookInspectorCloseButtonTestId(f.view)).click();
  await page.getByTestId(workbookInspectorToggleTestId(f.view)).click();
  await editField(page, f.view, "host.location");
  const resume = page.getByRole("button", {
    name: "Resume draft",
    exact: true,
  });
  await resume.focus();
  await resume.press("Enter");
  await expect(input).toBeFocused();
  await expect(input).toHaveValue("Keyboard retained location");
  await page.emulateMedia({ reducedMotion: "reduce" });
  for (const [width, zoom] of [
    [1280, 1],
    [390, 1],
    [1280, 2],
  ] as const) {
    await page.setViewportSize({ width, height: 720 });
    await page.evaluate((zoom) => {
      document.documentElement.style.zoom = String(zoom);
      if (zoom > 1) {
        document.body.style.lineHeight = "1.5";
        document.body.style.letterSpacing = "0.12em";
        document.body.style.wordSpacing = "0.16em";
      }
    }, zoom);
    await input.scrollIntoViewIfNeeded();
    await input.focus();
    await expect(input).toBeFocused();
    await expect(input).toBeInViewport({ ratio: 1 });
    await expect(
      page.getByRole("button", { name: "Discard draft", exact: true }),
    ).toHaveAccessibleName("Discard draft");
    await info.attach(`inspector-edit-${width}-${zoom}`, {
      body: await page.screenshot({ animations: "disabled", caret: "hide" }),
      contentType: "image/png",
    });
  }
  await info.attach("inspector-edit-accessibility-tree", {
    body: await page.getByTestId(entityInspectorTestId("host")).ariaSnapshot(),
    contentType: "text/plain",
  });
});

test("Reference selection reaches later real targets and retains staged choices through source and refresh failures", async ({
  page,
  workerAdmin,
}) => {
  test.setTimeout(180_000);
  const taskView = taskRequestsViewSchemaId,
    noteView = notesViewSchemaId,
    indicatorView = indicatorsViewSchemaId;
  const incident = await createIncident(
    page,
    uniqueIncidentKey("RSR"),
    "Reference selection recovery",
  );
  const task = await createViewRow(page, incident, taskView, {
    client_txn_id: uniqueTxn("rsr-task"),
    "task.title": "Reference owner",
    "task.task_kind": "question",
    "task.owner_user_id": workerAdmin.user_id,
  });
  const notes = [];
  for (let offset = 0; offset < 105; offset += 5) {
    notes.push(
      ...(await Promise.all(
        Array.from({ length: 5 }, (_, index) =>
          createViewRow(page, incident, noteView, {
            client_txn_id: uniqueTxn("rsr-note"),
            "note.title": `Reference ${String(offset + index).padStart(3, "0")}`,
          }),
        ),
      )),
    );
  }
  const indicator = await createViewRow(page, incident, indicatorView, {
    client_txn_id: uniqueTxn("rsr-indicator"),
    "indicator.display_value": "reference.example",
    "indicator.indicator_type": "domain_name",
    "indicator.value_kind": "atomic",
  });
  const retained = notes[0];
  if (!retained) throw new Error("Missing retained Note");
  let initialFailure = true,
    continuationFailure = true,
    refreshFailure = false;
  const reads: Record<string, unknown>[] = [],
    writes: Record<string, unknown>[] = [];
  await page.route(`**/views/${noteView}/query`, async (route) => {
    const body = route.request().postDataJSON();
    reads.push(body);
    if (!body.cursor_token && initialFailure) {
      initialFailure = false;
      return route.abort("failed");
    }
    if (body.cursor_token && continuationFailure) {
      continuationFailure = false;
      return route.abort("failed");
    }
    await route.continue();
  });
  await page.route(`**/views/${taskView}/query`, (route) =>
    refreshFailure ? route.abort("failed") : route.continue(),
  );
  await page.route(`**/api/v1/records/${task.record_id}`, async (route) => {
    if (route.request().method() !== "PATCH") return route.continue();
    writes.push(route.request().postDataJSON());
    const response = await route.fetch();
    expect(response.ok()).toBe(true);
    refreshFailure = true;
    await route.fulfill({ response });
  });
  await page.goto(`/?incident_id=${incident}&view_schema_id=${taskView}`);
  await openGenericInspectorForRecord(page, taskView, task.record_id);
  const input = await editField(page, taskView, "task.linked_record_ids");
  await input.fill(retained.record_id);
  const choose = page
    .getByTestId(workbookInspectorPanelTestId(taskView, "details"))
    .getByRole("button", { name: "Choose linked records", exact: true });
  await choose.focus();
  await choose.press("Enter");
  const popup = page.getByRole("dialog", {
    name: "Choose linked records",
    exact: true,
  });
  await popup.getByLabel("Reference surface").selectOption(noteView);
  await expect(popup.getByRole("alert")).toContainText("could not be loaded");
  await expect(input).toHaveValue(retained.record_id);
  await popup
    .getByRole("button", { name: "Retry references", exact: true })
    .click();
  await expect(popup).toContainText("Page 1: 100 candidates; more available");
  const list = popup.getByRole("listbox", {
    name: "Linked Records candidates",
  });
  const first = await list.locator("option").first().getAttribute("value");
  if (!first) throw new Error("Missing page one candidate");
  const selectedOnPage = () =>
    list.evaluate((element) =>
      Array.from(
        (element as HTMLSelectElement).selectedOptions,
        (option) => option.value,
      ),
    );
  await list.selectOption([...(await selectedOnPage()), first]);
  await popup.getByRole("button", { name: "Next", exact: true }).click();
  await expect(popup.getByRole("alert")).toContainText(
    "accepted page is retained",
  );
  await expect(list.locator("option")).toHaveCount(100);
  await popup
    .getByRole("button", { name: "Retry references", exact: true })
    .click();
  await expect(popup).toContainText("Page 2: 5 candidates; end of this source");
  const later = await list.locator("option").last().getAttribute("value");
  if (!later) throw new Error("Missing later candidate");
  await list.selectOption([...(await selectedOnPage()), later]);
  await popup
    .getByLabel("Reference filter field")
    .selectOption("note.created_by_user_id");
  const beforeFilter = reads.length;
  await popup
    .getByLabel("Reference filter value")
    .fill("00000000-0000-4000-8000-000000000999");
  expect(reads).toHaveLength(beforeFilter);
  await popup
    .getByRole("button", { name: "Apply filter", exact: true })
    .click();
  await expect(popup).toContainText("Page 1: 0 candidates; end of this source");
  await popup.getByLabel("Reference filter value").fill(workerAdmin.user_id);
  await popup
    .getByRole("button", { name: "Apply filter", exact: true })
    .click();
  await expect(popup).toContainText("Page 1: 100 candidates; more available");
  await expect(list.locator("option")).toHaveCount(100);
  await popup.getByLabel("Reference surface").selectOption(indicatorView);
  await expect(list.locator("option")).toHaveCount(1);
  await list.selectOption(`record:${indicator.record_id}`);
  expect(writes).toHaveLength(0);
  await expect(input).toHaveValue(retained.record_id);
  await popup
    .getByRole("button", { name: "Use selection", exact: true })
    .click();
  const selected = (await input.inputValue()).split("\n");
  expect(selected).toContain(retained.record_id);
  expect(selected).toContain(first.replace("record:", ""));
  expect(selected).toContain(later.replace("record:", ""));
  expect(selected).toContain(indicator.record_id);
  expect(writes).toHaveLength(0);
  await choose.click();
  await popup
    .getByRole("button", { name: "Cancel references", exact: true })
    .click();
  expect((await input.inputValue()).split("\n")).toEqual(selected);
  await page.getByTestId(genericEditSubmitTestId(taskView)).click();
  await expect(
    page.getByRole("region", { name: "Inspector changes" }),
  ).toContainText("Saved");
  expect(writes).toHaveLength(1);
  expect(writes[0]?.changes).toEqual([
    {
      field_key: "task.linked_record_ids",
      action_payload: {
        kind: "collection_actions_v1",
        actions: selected.map((linked_record_id) => ({
          op: "add_record_ref",
          linked_record_id,
        })),
      },
    },
  ]);
  // A separate candidate retry after the acknowledgement is still read-only.
  await choose.click();
  await popup.getByLabel("Reference surface").selectOption(noteView);
  await expect(popup).toContainText("Page 1: 100 candidates");
  await popup.getByRole("button", { name: "Next", exact: true }).click();
  await expect(popup).toContainText("Page 2: 5 candidates");
  await popup
    .getByRole("button", { name: "Cancel references", exact: true })
    .click();
  expect(writes).toHaveLength(1);
  expect(reads.length).toBeLessThanOrEqual(10);
  refreshFailure = false;
  await page
    .getByRole("button", { name: "Refresh saved change", exact: true })
    .click();
  await expect(
    page.getByRole("button", { name: "Refresh saved change", exact: true }),
  ).toHaveCount(0);
  expect(writes).toHaveLength(1);
  const saved = (await queryViewRows(page, incident, taskView)).find(
    (row) => row.record_id === task.record_id,
  );
  const value = saved?.cells["task.linked_record_ids"]?.value as {
    items: { item_ref: string }[];
  };
  expect(value.items).toHaveLength(selected.length);
  await page
    .getByRole("combobox", { name: "Collection edit action" })
    .selectOption("remove");
  const removal = page.getByTestId(genericEditValueTestId(taskView));
  const itemRef = await removal.locator("option").first().getAttribute("value");
  expect(value.items.map((item) => item.item_ref)).toContain(itemRef);
  if (!itemRef) throw new Error("Missing removal item_ref");
  await removal.selectOption(itemRef);
  const removalAccepted = page.waitForResponse(
    (response) =>
      response.request().method() === "PATCH" &&
      response.url().endsWith(`/records/${task.record_id}`),
  );
  await page.getByTestId(genericEditSubmitTestId(taskView)).click();
  expect((await removalAccepted).ok()).toBe(true);
  await expect(
    page.getByRole("region", { name: "Inspector changes" }),
  ).toContainText("Saved, version 3");
  await expect.poll(() => writes.length).toBe(2);
  expect(writes[1]?.changes).toEqual([
    {
      field_key: "task.linked_record_ids",
      action_payload: {
        kind: "collection_actions_v1",
        actions: [{ op: "remove_record_ref", item_ref: itemRef }],
      },
    },
  ]);
});

test("a11y.references native popup preserves keyboard focus and fits narrow zoomed and spaced layouts", async ({
  page,
}, info) => {
  const f = await fixture(page, evidenceViewSchemaId);
  await createViewRow(page, f.incident, partiesViewSchemaId, {
    client_txn_id: uniqueTxn("rsr-a11y"),
    "party.display_name": "Keyboard accessible Party",
    "party.party_kind": "person",
  });
  const input = await editField(page, f.view, "evidence.source_party_id");
  await page.emulateMedia({ reducedMotion: "reduce" });
  for (const [width, zoom] of [
    [1280, 1],
    [390, 1],
    [1280, 2],
  ] as const) {
    await page.setViewportSize({ width, height: 720 });
    await page.evaluate((zoom) => {
      document.documentElement.style.zoom = String(zoom);
      document.body.style.lineHeight = "1.5";
      document.body.style.letterSpacing = "0.12em";
      document.body.style.wordSpacing = "0.16em";
    }, zoom);
    const trigger = page
      .getByTestId(workbookInspectorPanelTestId(f.view, "details"))
      .getByRole("button", { name: "Choose source party", exact: true });
    await trigger.scrollIntoViewIfNeeded();
    await trigger.focus();
    await trigger.press("Enter");
    const popup = page.getByRole("dialog", {
      name: "Choose source party",
      exact: true,
    });
    await expect(popup).toContainText(
      "Page 1: 1 candidates; end of this source",
    );
    expect(
      await popup.evaluate((element) => element.matches(":popover-open")),
    ).toBe(true);
    await expect(popup).toBeInViewport({ ratio: 1 });
    const list = popup.getByRole("listbox", {
      name: "Source Party candidates",
    });
    await list.focus();
    await expect(list).toBeFocused();
    const cancel = popup.getByRole("button", {
      name: "Cancel references",
      exact: true,
    });
    await cancel.focus();
    await cancel.press("Tab");
    expect(
      await popup.evaluate((element) =>
        element.contains(document.activeElement),
      ),
    ).toBe(true);
    await info.attach(`reference-picker-${width}-${zoom}`, {
      body: await page.screenshot({ animations: "disabled", caret: "hide" }),
      contentType: "image/png",
    });
    await info.attach(`reference-picker-tree-${width}-${zoom}`, {
      body: await popup.ariaSnapshot(),
      contentType: "text/plain",
    });
    await page.keyboard.press("Escape");
    await expect(popup).toHaveCount(0);
    await expect(input).toBeFocused();
  }
});

test("Reference target deletion merge and membership removal preserve exact choices and owner admission", async ({
  page,
  workerAdmin,
  workerAdminRequest,
}) => {
  const f = await fixture(page, evidenceViewSchemaId);
  const partyView = partiesViewSchemaId,
    taskView = taskRequestsViewSchemaId;
  const party = await createViewRow(page, f.incident, partyView, {
    client_txn_id: uniqueTxn("rsr-deleted-party"),
    "party.display_name": "Chosen before deletion",
    "party.party_kind": "person",
  });
  const input = await editField(page, f.view, "evidence.source_party_id");
  const panel = page.getByTestId(
    workbookInspectorPanelTestId(f.view, "details"),
  );
  await panel.getByRole("button", { name: "Choose source party" }).click();
  const picker = page.getByRole("dialog", { name: "Choose source party" });
  await picker
    .getByRole("listbox", { name: "Source Party candidates" })
    .selectOption(`party:${party.record_id}`);
  await picker.getByRole("button", { name: "Use selection" }).click();
  const deleted = await publicHttpOperation({
    operationID: "deleteRecord",
    pathParameters: { record_id: party.record_id },
    body: {
      base_row_version: 1,
      client_txn_id: uniqueTxn("rsr-delete-target"),
      reason: "Reference eligibility evidence",
    },
    headers: await csrfHeaders(page),
    request: atJsonOrigin(page.request, apiBase),
  });
  expect(deleted.ok).toBe(true);
  const response = page.waitForResponse(
    (response) =>
      response.request().method() === "PATCH" &&
      response.url().endsWith(`/records/${f.first.record_id}`),
  );
  await page.getByTestId(genericEditSubmitTestId(f.view)).click();
  expect((await response).ok()).toBe(false);
  await expect(input).toHaveValue(party.record_id);
  await panel.getByRole("button", { name: "Choose source party" }).click();
  await expect(picker).toContainText(
    "Page 1: 0 candidates; end of this source",
  );
  await expect(picker).toContainText("Chosen before deletion");
  await picker.getByRole("button", { name: "Cancel references" }).click();
  await expect(input).toHaveValue(party.record_id);
  const task = await createViewRow(page, f.incident, taskView, {
    client_txn_id: uniqueTxn("rsr-member-task"),
    "task.title": "Member target",
    "task.task_kind": "question",
    "task.owner_user_id": workerAdmin.user_id,
  });
  const member = await createIncidentMemberUser(page, f.incident, {
    email: uniqueEmail("rsr-membership"),
    display_name: "Selected incident member",
    initial_password: "ReferenceMembership1!",
    role: "editor",
    is_deployment_admin: false,
    mfa_required: false,
  });
  await switchOrdinarySheet(page, taskView);
  await openGenericInspectorForRecord(page, taskView, task.record_id);
  const owner = await editField(page, taskView, "task.owner_user_id");
  const taskPanel = page.getByTestId(
    workbookInspectorPanelTestId(taskView, "details"),
  );
  await taskPanel
    .getByRole("button", { name: "Choose owner", exact: true })
    .click();
  const members = page.getByRole("dialog", {
    name: "Choose owner",
    exact: true,
  });
  await members
    .getByRole("listbox", { name: "Owner candidates" })
    .selectOption(`incident_member:${member.user_id}`);
  await members.getByRole("button", { name: "Use selection" }).click();
  expect(
    (
      await workerAdminRequest.delete(
        `/api/v1/incidents/${f.incident}/memberships/${member.user_id}`,
        { data: { base_membership_version: 1 } },
      )
    ).status(),
  ).toBe(204);
  const rejected = page.waitForResponse(
    (response) =>
      response.request().method() === "PATCH" &&
      response.url().endsWith(`/records/${task.record_id}`),
  );
  await page.getByTestId(genericEditSubmitTestId(taskView)).click();
  expect((await rejected).ok()).toBe(false);
  await expect(owner).toHaveValue(member.user_id);
  await taskPanel
    .getByRole("button", { name: "Choose owner", exact: true })
    .click();
  await expect(
    members.getByRole("option", { name: new RegExp(member.user_id) }),
  ).toHaveCount(0);
  await expect(members).toContainText("Selected incident member");
  await members.getByRole("button", { name: "Cancel references" }).click();
  await expect(owner).toHaveValue(member.user_id);
  const loser = await createViewRow(page, f.incident, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("rsr-merge-loser"),
    "host.display_name": "Chosen before merge",
  });
  const survivor = await createViewRow(page, f.incident, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("rsr-merge-survivor"),
    "host.display_name": "Surviving target",
  });
  const linked = await editField(page, taskView, "task.linked_record_ids");
  await taskPanel
    .getByRole("button", { name: "Choose linked records", exact: true })
    .click();
  const records = page.getByRole("dialog", {
    name: "Choose linked records",
    exact: true,
  });
  await records.getByLabel("Reference surface").selectOption(hostsViewSchemaId);
  await records
    .getByRole("listbox", { name: "Linked Records candidates" })
    .selectOption(`record:${loser.record_id}`);
  await records.getByRole("button", { name: "Use selection" }).click();
  const merged = await publicHttpOperation({
    operationID: "mergeEntityRecord",
    pathParameters: { survivor_record_id: survivor.record_id },
    body: {
      client_txn_id: uniqueTxn("rsr-merge"),
      loser_record_id: loser.record_id,
      loser_base_row_version: 1,
      survivor_base_row_version: 1,
      reason: "Reference target merge evidence",
    },
    headers: await csrfHeaders(page),
    request: atJsonOrigin(page.request, apiBase),
  });
  expect(merged.ok).toBe(true);
  await taskPanel
    .getByRole("button", { name: "Choose linked records", exact: true })
    .click();
  await expect(records).toContainText(
    "Page 1: 1 candidates; end of this source",
  );
  await expect(
    records.getByRole("option", { name: new RegExp(loser.record_id) }),
  ).toHaveCount(0);
  await expect(records).toContainText("Chosen before merge");
  await records.getByRole("button", { name: "Cancel references" }).click();
  await expect(linked).toHaveValue(loser.record_id);
  const mergeResult = page.waitForResponse(
    (response) =>
      response.request().method() === "PATCH" &&
      response.url().endsWith(`/records/${task.record_id}`),
  );
  await page.getByTestId(genericEditSubmitTestId(taskView)).click();
  const mergeResponse = await mergeResult;
  expect(mergeResponse.ok()).toBe(true);
  expect(
    mergeResponse.request().postDataJSON().changes[0].action_payload.actions,
  ).toEqual([{ op: "add_record_ref", linked_record_id: loser.record_id }]);
  const savedTask = (await queryViewRows(page, f.incident, taskView)).find(
    (row) => row.record_id === task.record_id,
  );
  const links = savedTask?.cells["task.linked_record_ids"]?.value as {
    items: { item_ref: string }[];
  };
  expect(links.items.map((item) => item.item_ref)).toEqual([
    `record_ref:${loser.record_id}`,
  ]);
});
