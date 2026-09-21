import {
  applyFilterChip,
  changeGrouping,
  collapseGridGroup,
  expandGridGroup,
  removeFilterChip,
  scrollGridCellIntoView,
  sortByHeader,
} from "@cartulary/test-utils/grid";
import {
  authTestId,
  entityInspectorTestId,
  genericEditValueTestId,
  gridGroupRowTestId,
  gridScrollportSelector,
  gridShellTestId,
  rowCellTestId,
  workbookColumnsMenuTestId,
  workbookColumnsMenuTriggerTestId,
  workbookGridEditorTestId,
  workbookInspectorToggleTestId,
} from "@cartulary/ui-contracts";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
  requireViewContract,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { APIRequestContext, Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { csrfHeaders } from "./support/auth/browserSession";
import { revokeAllSessions } from "./support/auth/sessions";
import { installVisualPreferences } from "./support/auth/visualPreferences";
import { installPatchController } from "./support/collaboration/replay";
import { createIncident } from "./support/incidents/fixtures";
import { createIncidentMemberUser } from "./support/incidents/memberships";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { installIncidentSocketMonitor } from "./support/transport/incidentSocket";
import { publicHttpOperation } from "./support/transport/publicHttpOperationClient";
import { atJsonOrigin } from "./support/transport/publicJsonClient";
import { holdBrowserRequest } from "./support/transport/requestInterception";
import {
  commitOrdinary,
  ordinaryField,
  switchOrdinarySheet,
} from "./support/workbook/ordinaryCreate";
import {
  createViewRow,
  patchRecord,
  queryViewRows,
} from "./support/workbook/query";
import {
  createSavedView,
  selectSavedView,
} from "./support/workbook/savedViews";

const surfaces = [
  { view: hostsViewSchemaId, type: "host" as const, primary: "hostname" },
  { view: identitiesViewSchemaId, type: "identity" as const, primary: "upn" },
];
type EntitySurface = (typeof surfaces)[number];
const grid = (page: Page) => page.locator(gridScrollportSelector());
const cell = (page: Page, id: string, field: string) =>
  page.getByTestId(rowCellTestId(id, field));
const editor = (page: Page, id: string, field: string) =>
  page.getByTestId(workbookGridEditorTestId(id, field));

async function showField(page: Page, f: EntitySurface, label: string) {
  await page.getByTestId(workbookColumnsMenuTriggerTestId(f.view)).click();
  await page
    .getByTestId(workbookColumnsMenuTestId(f.view))
    .getByRole("checkbox", { name: label, exact: true })
    .check();
  await page
    .getByRole("button", { name: "Close columns", exact: true })
    .click();
}
async function reveal(page: Page, f: EntitySurface, id: string, field: string) {
  await scrollGridCellIntoView({
    page,
    surface: f.view,
    recordId: id,
    cellKey: field,
  });
  return cell(page, id, field);
}
async function seed(page: Page, f: EntitySurface, count = 3) {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("EFN"),
    "Entity Find evidence",
  );
  const rows: { record_id: string; row_version: number }[] = [];
  let next = 0;
  await Promise.all(
    Array.from({ length: 8 }, async () => {
      for (;;) {
        const index = next++;
        if (index >= count) return;
        rows[index] = await createViewRow(page, incident, f.view, {
          client_txn_id: uniqueTxn("entity-find"),
          [`${f.type}.display_name`]: `${String(index).padStart(4, "0")} needle Café`,
          [`${f.type}.${f.primary}`]: `primary-${index}.example`,
          ...(f.type === "host"
            ? {
                "host.fqdn": `fallback-${index}.example`,
                "host.location": "hidden-only",
              }
            : {
                "identity.email": `email-${index}@example.test`,
                "identity.sam_account_name": `hidden-only-${index}`,
              }),
        });
      }
    }),
  );
  const first = rows[0];
  if (!first) throw new Error("Missing first entity fixture");
  rows[0] = await patchRecord(page, first.record_id, {
    view_schema_id: f.view,
    base_row_version: first.row_version,
    client_txn_id: uniqueTxn("entity-alias"),
    changes: [
      {
        field_key: `${f.type}.aliases`,
        action_payload: {
          kind: "collection_actions_v1",
          actions: [
            { op: "add_alias", alias_text: "Alpha needle" },
            { op: "add_alias", alias_text: "Beta separate" },
          ],
        },
      },
    ],
  });
  const sockets = installIncidentSocketMonitor(page, incident);
  await page.goto(`/?incident_id=${incident}&view_schema_id=${f.view}`);
  await sockets.waitForAcceptedSocket();
  await expect(page.getByTestId(gridShellTestId(f.view))).toBeVisible();
  await sortByHeader(page, f.view, `${f.type}.display_name`);
  await expect(
    cell(page, first.record_id, `${f.type}.display_name`),
  ).toHaveText("0000 needle Café");
  return {
    ...f,
    sockets,
    incident,
    rows,
    first: first.record_id,
    last: rows[count - 1]?.record_id ?? first.record_id,
  };
}

test("Entity committed presentation characterization preserves labels aliases drafts and independent accepted sheets", async ({
  page,
}, info) => {
  const observations = [];
  for (const f of surfaces) {
    const data = await seed(page, f);
    const primary = await reveal(page, f, data.first, `${f.type}.${f.primary}`);
    await expect(primary).toHaveText(
      f.type === "host" ? "primary-0.example" : "email-0@example.test",
    );
    const primaryText = await primary.textContent();
    const aliases = await reveal(page, f, data.first, `${f.type}.aliases`);
    await expect(aliases).toContainText("Alpha needle");
    await expect(aliases).toContainText("Beta separate");
    const aliasText = await aliases.textContent();
    await showField(page, f, "Reusable Identifiers");
    await expect(
      await reveal(page, f, data.first, `${f.type}.reusable_identifiers`),
    ).toHaveText("None");
    const display = await reveal(page, f, data.first, `${f.type}.display_name`);
    await display.click();
    const raw = editor(page, data.first, `${f.type}.display_name`);
    await expect(raw).toBeFocused();
    await raw.fill("  exact raw work  ");
    await page.getByTestId(workbookColumnsMenuTriggerTestId(f.view)).click();
    await expect(raw).toHaveValue("  exact raw work  ");
    await page
      .getByRole("button", { name: "Close columns", exact: true })
      .click();
    await raw.focus();
    await raw.press("Escape");
    await page.getByTestId(workbookInspectorToggleTestId(f.view)).click();
    const inspector = page.getByTestId(entityInspectorTestId(f.type));
    await expect(inspector).toBeVisible();
    await inspector
      .getByRole("textbox", { name: "Alias text" })
      .fill("  unsubmitted alias  ");
    await expect(
      inspector.getByRole("textbox", { name: "Alias text" }),
    ).toHaveValue("  unsubmitted alias  ");
    observations.push({
      view: f.view,
      primary: primaryText,
      aliases: aliasText,
      loaded: data.rows.length,
      mountedRows: await grid(page).locator("[data-record-id]").count(),
    });
  }
  await info.attach("entity-presentation-observations", {
    body: JSON.stringify(observations, null, 2),
    contentType: "application/json",
  });
});

const focusCell = (page: Page, id: string, field: string) =>
  cell(page, id, field).locator("xpath=ancestor::*[@role='gridcell'][1]");
const entry = (page: Page) =>
  page.getByRole("button", { name: "Find in loaded rows", exact: true });
const input = (page: Page) =>
  page.getByRole("textbox", { name: "Find in loaded rows", exact: true });
const status = (page: Page) =>
  page.getByRole("status").filter({
    hasText:
      /(?:matching cell|No matches in loaded rows|Enter text to find|Finding in loaded rows|Waiting for the edit)/,
  });
async function find(page: Page, term: string, count: number) {
  await entry(page).click();
  await input(page).fill(term);
  await expect(status(page)).toContainText(
    count
      ? `${count} matching ${count === 1 ? "cell" : "cells"}`
      : "No matches in loaded rows.",
  );
}
async function navigate(page: Page, direction = "Enter") {
  await input(page).press(direction);
  await expect(input(page)).toHaveCount(0);
}
function observe(page: Page, incident: string) {
  const requests: string[] = [];
  page.on("request", (request) => {
    if (
      (request.url().includes(`/incidents/${incident}/views/`) &&
        request.url().endsWith("/query")) ||
      request.url().includes("/saved-views") ||
      request.method() === "PATCH" ||
      request.url().endsWith("/rows")
    )
      requests.push(`${request.method()} ${request.url()}`);
  });
  return requests;
}
async function setField(
  page: Page,
  f: EntitySurface,
  field: string,
  visible: boolean,
) {
  await page.getByTestId(workbookColumnsMenuTriggerTestId(f.view)).click();
  await page
    .getByTestId(workbookColumnsMenuTestId(f.view))
    .getByRole("checkbox", {
      name: requireViewContract(f.view).fieldMap[field]?.label ?? "missing",
      exact: true,
    })
    .setChecked(visible);
  await page
    .getByRole("button", { name: "Close columns", exact: true })
    .click();
}

test("Entity Find literal renderer fragments counts order and native keyboard ownership work on both schemas", async ({
  page,
}) => {
  for (const schema of surfaces) {
    const f = await seed(page, schema);
    const field = `${f.type}.display_name`;
    await setField(
      page,
      f,
      `${f.type}.${f.type === "host" ? "location" : "sam_account_name"}`,
      false,
    );
    const requests = observe(page, f.incident);
    await find(page, "needle", 4);
    await navigate(page);
    await expect(focusCell(page, f.first, field)).toBeFocused();
    await entry(page).click();
    await navigate(page);
    await expect(focusCell(page, f.first, `${f.type}.aliases`)).toBeFocused();
    await entry(page).click();
    await navigate(page, "Shift+Enter");
    await expect(focusCell(page, f.first, field)).toBeFocused();
    await entry(page).click();
    await navigate(page, "Shift+Enter");
    await expect(focusCell(page, f.last, field)).toBeFocused();
    await expect(status(page)).toContainText("Wrapped to end.");
    await find(page, "primary-0.example", f.type === "host" ? 1 : 0);
    if (f.type === "identity") {
      // Both the UPN renderer's current fallback and the Email cell display it.
      await find(page, "email-0@example.test", 2);
    }
    for (const term of [
      "Alpha needle, Beta separate",
      "No aliases",
      "None",
      "hidden-only",
      f.first,
      "needle.*",
    ]) {
      await find(page, term, 0);
    }
    await find(page, "Beta separate", 1);
    await find(page, "CAFE\u0301", 3);
    await page
      .getByRole("checkbox", { name: "Match case", exact: true })
      .check();
    await expect(status(page)).toHaveText("No matches in loaded rows.");
    await input(page).fill("Café");
    await expect(status(page)).toContainText("3 matching cells");
    await input(page).fill(" Café ");
    await expect(status(page)).toHaveText("No matches in loaded rows.");
    await input(page).fill("");
    await expect(status(page)).toHaveText("Enter text to find in loaded rows.");
    expect(
      await input(page).evaluate((element) =>
        element.dispatchEvent(
          new KeyboardEvent("keydown", {
            key: "Enter",
            isComposing: true,
            bubbles: true,
            cancelable: true,
          }),
        ),
      ),
    ).toBe(true);
    await input(page).press("Escape");
    await (await reveal(page, f, f.first, field)).click();
    const raw = editor(page, f.first, field);
    expect(
      await raw.evaluate((element) =>
        element.dispatchEvent(
          new KeyboardEvent("keydown", {
            key: "f",
            ctrlKey: true,
            bubbles: true,
            cancelable: true,
          }),
        ),
      ),
    ).toBe(true);
    await raw.press("Escape");
    await focusCell(page, f.first, field).press("Control+f");
    await expect(input(page)).toBeFocused();
    await expect(input(page)).toHaveValue("");
    expect(requests).toEqual([]);
  }
});

test("Entity Find reveals offscreen rows and columns with semantic focus and bounded mounted records on both schemas", async ({
  page,
}, info) => {
  for (const schema of surfaces) {
    const f = await seed(page, schema, 85);
    const field = `${f.type}.display_name`;
    const far = `${f.type}.${f.type === "host" ? "fqdn" : "sam_account_name"}`;
    await showField(
      page,
      f,
      requireViewContract(f.view).fieldMap[far]?.label ?? "missing",
    );
    await page.getByTestId(workbookColumnsMenuTriggerTestId(f.view)).click();
    const columns = page.getByTestId(workbookColumnsMenuTestId(f.view));
    const label = requireViewContract(f.view).fieldMap[far]?.label;
    const later = columns.getByRole("button", {
      name: `Move ${label} later`,
      exact: true,
    });
    for (let i = 0; i < 40 && (await later.isEnabled()); i++)
      await later.click();
    await columns
      .getByRole("button", {
        name: `Freeze through ${requireViewContract(f.view).fieldMap[field]?.label}`,
        exact: true,
      })
      .click();
    await columns
      .getByRole("button", { name: "Close columns", exact: true })
      .click();
    const requests = observe(page, f.incident);
    await find(page, "needle", 86);
    expect(
      await grid(page).locator("[data-grid-record-id]").count(),
    ).toBeLessThan(60);
    await expect(cell(page, f.last, field)).toHaveCount(0);
    await input(page).fill("0084 needle");
    await expect(status(page)).toContainText("1 matching cell");
    await navigate(page);
    await expect(focusCell(page, f.last, field)).toBeFocused();
    await expect(editor(page, f.last, field)).toHaveCount(0);
    await expect(cell(page, f.last, far)).toHaveCount(0);
    await find(
      page,
      f.type === "host" ? "fallback-84.example" : "hidden-only-84",
      1,
    );
    await navigate(page);
    const target = focusCell(page, f.last, far);
    await expect(target).toBeFocused();
    await expect(target).toBeInViewport();
    await expect(target).toHaveAttribute(
      "aria-description",
      /Current Find match/,
    );
    await expect(
      target.getByRole("img", { name: "Current Find match", exact: true }),
    ).toBeVisible();
    const frozenRight = await grid(page)
      .locator('.cartulary-grid-frozen-data[role="columnheader"]')
      .last()
      .evaluate((node) => node.getBoundingClientRect().right);
    expect((await target.boundingBox())?.x).toBeGreaterThanOrEqual(frozenRight);
    expect(requests).toEqual([]);
    await info.attach(`entity-find-virtualized-${f.type}`, {
      body: await page.screenshot(),
      contentType: "image/png",
    });
  }
});

test("Entity Find borrows exact grid work and gates rejected accepted and superseded destinations on both schemas", async ({
  page,
}) => {
  for (const schema of surfaces) {
    const f = await seed(page, schema);
    const field = `${f.type}.display_name`;
    await cell(page, f.first, field).click();
    const raw = editor(page, f.first, field);
    await raw.fill("  Exact unfinished draft  ");
    const hold = await holdBrowserRequest(page, {
      method: "PATCH",
      path: `/api/v1/records/${f.first}`,
    });
    try {
      await find(page, "Exact unfinished draft", 0);
      await input(page).fill("0002 needle");
      await expect(status(page)).toContainText("1 matching cell");
      expect(hold.hitCount()).toBe(0);
      await expect(raw).toHaveValue("  Exact unfinished draft  ");
      await input(page).press("Escape");
      await expect(raw).toBeFocused();
      expect(hold.hitCount()).toBe(0);
      await find(page, "0002 needle", 1);
      await input(page).press("Enter");
      await hold.waitForHit;
      await expect(raw).toHaveValue("  Exact unfinished draft  ");
      await expect(status(page)).toContainText("Waiting for the edit to save");
      await page.getByRole("button", { name: "Close Find" }).click();
      await expect(raw).toBeFocused();
      hold.release();
      await expect(cell(page, f.first, field)).toHaveText(
        "Exact unfinished draft",
      );
      await find(page, "0002 needle", 1);
      await navigate(page);
      await expect(focusCell(page, f.last, field)).toBeFocused();
      expect(hold.hitCount()).toBe(1);
    } finally {
      await hold.dispose();
    }
    await cell(page, f.first, field).click();
    await raw.fill("Accepted without cancellation");
    await find(page, "0002 needle", 1);
    await navigate(page);
    await expect(focusCell(page, f.last, field)).toBeFocused();
    await find(page, "Accepted without cancellation", 1);
    await navigate(page);
    await cell(page, f.first, field).click();
    await raw.fill("Superseded destination edit");
    const replacement = await holdBrowserRequest(page, {
      method: "PATCH",
      path: `/api/v1/records/${f.first}`,
    });
    try {
      await find(page, "0002 needle", 1);
      await input(page).press("Enter");
      await replacement.waitForHit;
      await input(page).fill("0001 needle");
      replacement.release();
      await expect(cell(page, f.first, field)).toHaveText(
        "Superseded destination edit",
      );
      await expect(input(page)).toBeFocused();
      await expect(focusCell(page, f.last, field)).not.toBeFocused();
      expect(replacement.hitCount()).toBe(1);
      await navigate(page);
      const middle = f.rows[1];
      if (!middle) throw new Error("Missing middle entity fixture");
      await expect(focusCell(page, middle.record_id, field)).toBeFocused();
    } finally {
      await replacement.dispose();
    }
    await cell(page, f.first, field).click();
    await raw.fill("  rejected raw work  ");
    const rejected = await installPatchController(page);
    try {
      rejected.failNextPatch(422, "invalid_request", { recordId: f.first });
      await find(page, "0002 needle", 1);
      await input(page).press("Enter");
      await expect(raw).toHaveValue("  rejected raw work  ");
      await expect(raw).toBeFocused();
      await expect(status(page)).toContainText("Correct the original edit");
      await raw.press("Escape");
    } finally {
      await rejected.dispose();
    }
  }
});

test("Entity Find preserves independent inspector scalar alias and recordless authoring without submission on both schemas", async ({
  page,
}) => {
  for (const schema of surfaces) {
    const f = await seed(page, schema);
    const field = `${f.type}.display_name`;
    await cell(page, f.first, field).click();
    await editor(page, f.first, field).press("Escape");
    await page.getByTestId(workbookInspectorToggleTestId(f.view)).click();
    const inspector = page.getByTestId(entityInspectorTestId(f.type));
    await inspector.locator(`[data-inspector-edit-field="${field}"]`).click();
    const scalar = page.getByTestId(genericEditValueTestId(f.view));
    await scalar.fill("  Inspector scalar raw  ");
    await inspector.getByText("Manage aliases", { exact: true }).click();
    const alias = inspector.getByRole("textbox", {
      name: "Alias text",
      exact: true,
    });
    await alias.fill("  Inspector alias raw  ");
    const requests = observe(page, f.incident);
    await find(page, "Inspector alias raw", 0);
    await input(page).press("Escape");
    await expect(alias).toBeFocused();
    await scalar.focus();
    await find(page, "Inspector scalar raw", 0);
    await input(page).press("Escape");
    await expect(scalar).toBeFocused();
    await find(page, "0002 needle", 1);
    await navigate(page);
    await expect(focusCell(page, f.last, field)).toBeFocused();
    await expect(inspector).toContainText("0000 needle Café");
    await expect(alias).toHaveValue("  Inspector alias raw  ");
    await expect(scalar).toHaveValue("  Inspector scalar raw  ");
    await entry(page).click();
    await input(page).press("Escape");
    await expect(focusCell(page, f.last, field)).toBeFocused();
    await expect(alias).toHaveValue("  Inspector alias raw  ");
    const draft = await ordinaryField(page, f.view, field);
    await draft.fill("  recordless needle  ");
    await find(page, "recordless needle", 0);
    await input(page).fill("0002 needle");
    await expect(status(page)).toContainText("1 matching cell");
    await navigate(page);
    await expect(draft).toHaveValue("  recordless needle  ");
    expect(requests).toEqual([]);
  }
});

test("Entity Find follows grouping hidden fields saved views live deletion and surface retirement on both schemas", async ({
  page,
}) => {
  for (const schema of surfaces) {
    const f = await seed(page, schema);
    const field = `${f.type}.display_name`;
    await find(page, "needle", 4);
    await navigate(page);
    await setField(page, f, `${f.type}.aliases`, false);
    await entry(page).click();
    await expect(input(page)).toHaveValue("needle");
    await expect(status(page)).toContainText("1 of 3 matching cells");
    await setField(page, f, `${f.type}.aliases`, true);
    const groupField = `${f.type}.${f.type === "host" ? "host_state" : "identity_state"}`;
    const groupValue = String(
      (await queryViewRows(page, f.incident, f.view))[0]?.cells[groupField]
        ?.value,
    );
    await changeGrouping(page, f.view, groupField);
    const groupId = gridGroupRowTestId(f.view, groupField, groupValue);
    await collapseGridGroup({ page, surface: f.view, groupTestId: groupId });
    await entry(page).click();
    await expect(status(page)).toHaveText("No matches in loaded rows.");
    await expandGridGroup({ page, surface: f.view, groupTestId: groupId });
    await entry(page).click();
    await expect(status(page)).toContainText("4 matching cells");
    const saved = await createSavedView(page, f.incident, {
      display_name: "Entity Find hidden names",
      view_schema_id: f.view,
      layout_json: {
        layout_schema_id: "cartulary.layout.v1",
        column_order: requireViewContract(f.view).fields.map(
          (field) => field.fieldKey,
        ),
        column_widths: [],
        hidden_field_keys: [field],
      },
    });
    await switchOrdinarySheet(page, timelineViewSchemaId);
    await entry(page).click();
    await expect(input(page)).toHaveValue("");
    await switchOrdinarySheet(page, f.view);
    await entry(page).click();
    await expect(input(page)).toHaveValue("");
    await input(page).fill("needle");
    await expect(status(page)).toContainText("4 matching cells");
    await selectSavedView(page, f.view, saved.saved_view_id);
    await entry(page).click();
    await expect(input(page)).toHaveValue("needle");
    await expect(status(page)).toContainText("1 matching cell");
    await selectSavedView(page, f.view, "");
    await setField(page, f, field, true);
    await sortByHeader(page, f.view, field);
    await find(page, "0002 needle", 1);
    await navigate(page);
    await expect(focusCell(page, f.last, field)).toBeFocused();
    const changed = await patchRecord(page, f.last, {
      view_schema_id: f.view,
      base_row_version: f.rows[2]?.row_version ?? 1,
      client_txn_id: uniqueTxn("entity-find-live"),
      changes: [{ field_key: field, value: "Changed committed value" }],
    });
    await expect(cell(page, f.last, field)).toHaveText(
      "Changed committed value",
    );
    await expect(status(page)).toHaveText("No matches in loaded rows.");
    await find(page, "Changed committed value", 1);
    await navigate(page);
    const start = f.sockets.messageCount();
    const removed = await publicHttpOperation({
      operationID: "deleteRecord",
      request: atJsonOrigin(page.request, apiBase),
      headers: await csrfHeaders(page),
      pathParameters: { record_id: f.last },
      body: {
        base_row_version: changed.row_version,
        client_txn_id: uniqueTxn("entity-find-delete"),
        reason: "Find fixture deletion",
      },
    });
    expect(removed.ok).toBe(true);
    const deletion = await f.sockets.waitForMessage("record_changed", {
      startAt: start,
      matches: (message) => message.payload.record_id === f.last,
    });
    expect(deletion.payload.changed_field_keys).toEqual([]);
    await expect(status(page)).toHaveText("No matches in loaded rows.");
    await expect(cell(page, f.last, field)).toHaveCount(0);
  }
});

test("Entity Find searches accepted bounded windows through continuation eviction and failed replacement on both schemas", async ({
  page,
}, info) => {
  for (const schema of surfaces) {
    const f = await seed(page, schema, 405);
    const browsing = page.getByRole("group", { name: "Workbook browsing" });
    const more = browsing.getByRole("button", {
      name: "Load more",
      exact: true,
    });
    const requests = observe(page, f.incident);
    await find(page, "needle", 101);
    expect(requests).toEqual([]);
    for (const count of [201, 301]) {
      await more.click();
      await entry(page).click();
      await expect(status(page)).toContainText(`${count} matching cells`);
    }
    const baseline = [...requests];
    await input(page).fill("obsolete term");
    await input(page).fill("needle");
    await expect(status(page)).toContainText("301 matching cells");
    expect(requests).toEqual(baseline);
    expect(
      await grid(page).locator("[data-grid-record-id]").count(),
    ).toBeLessThan(60);
    await more.click();
    await entry(page).click();
    await expect(status(page)).toContainText("300 matching cells");
    await input(page).fill("0000 needle");
    await expect(status(page)).toHaveText("No matches in loaded rows.");
    await input(page).fill("needle");
    await expect(status(page)).toContainText("300 matching cells");
    const route = `**/api/v1/incidents/${f.incident}/views/${f.view}/query`;
    await page.route(route, (request) => request.abort("failed"), { times: 1 });
    await browsing
      .getByRole("button", { name: "Refresh", exact: true })
      .click();
    await expect(browsing).toContainText("Retry");
    await entry(page).click();
    await expect(status(page)).toContainText("300 matching cells");
    await expect(
      page.getByRole("region", { name: "Find in loaded rows" }),
    ).toContainText("Showing retained rows");
    await info.attach(`entity-find-window-${f.type}`, {
      body: await page.screenshot(),
      contentType: "image/png",
    });
    const retry = browsing.getByRole("button", { name: "Retry", exact: true });
    await retry.focus();
    await expect(input(page)).toHaveCount(0);
    await retry.press("Enter");
    await entry(page).click();
    await expect(input(page)).toHaveValue("needle");
    await expect(status(page)).toContainText("101 matching cells");
    await applyFilterChip(
      page,
      f.view,
      `${f.type}.${f.type === "host" ? "host_state" : "identity_state"}`,
      "retired",
    );
    await entry(page).click();
    await expect(status(page)).toHaveText("No matches in loaded rows.");
    const createdName = "0000 creation-only needle";
    const createInput = await ordinaryField(
      page,
      f.view,
      `${f.type}.display_name`,
    );
    await createInput.fill(createdName);
    await find(page, "creation-only", 0);
    await input(page).press("Escape");
    await expect(createInput).toHaveValue(createdName);
    const receipt = page.waitForResponse(
      (response) =>
        response.request().method() === "POST" &&
        response.url().endsWith(`/views/${f.view}/rows`),
    );
    await commitOrdinary(page, f.view);
    const created = (await (await receipt).json()).data.row;
    expect(created.cells[`${f.type}.display_name`].value).toBe(createdName);
    // Ordinary entities admit receipts only through their query owner. Unlike
    // Timeline creation, they do not materialize an out-of-query creation pin.
    await expect(
      cell(page, created.record_id, `${f.type}.display_name`),
    ).toHaveCount(0);
    await find(page, "creation-only", 0);
    await removeFilterChip(
      page,
      f.view,
      `${f.type}.${f.type === "host" ? "host_state" : "identity_state"}`,
    );
    await entry(page).click();
    await expect(status(page)).toContainText("1 matching cell");
    await input(page).fill("needle");
    await expect(status(page)).toContainText("101 matching cells");
  }
});

test("Entity Find accepted edits retain newer work through failed refresh and recover with reads without write replay on both schemas", async ({
  page,
}) => {
  for (const schema of surfaces) {
    const f = await seed(page, schema);
    const field = `${f.type}.display_name`;
    let committed = false,
      damage = true,
      writes = 0,
      reads = 0;
    let release = () => {};
    const ack = new Promise<void>((resolve) => {
      release = resolve;
    });
    const queryRoute = `**/api/v1/incidents/${f.incident}/views/${f.view}/query`;
    const patchRoute = `**/api/v1/records/${f.first}`;
    await page.route(queryRoute, async (route) => {
      reads++;
      if (writes > 0 && damage) return route.abort("failed");
      await route.continue();
    });
    await page.route(patchRoute, async (route) => {
      if (route.request().method() !== "PATCH") return route.continue();
      writes++;
      const response = await route.fetch();
      expect(response.ok()).toBe(true);
      committed = true;
      await ack;
      await route.fulfill({ response });
    });
    try {
      await cell(page, f.first, field).click();
      const raw = editor(page, f.first, field);
      await raw.fill("Accepted despite refresh");
      await find(page, "0002 needle", 1);
      await input(page).press("Enter");
      await expect.poll(() => committed).toBe(true);
      await raw.click();
      await expect(raw).toBeFocused();
      await raw.fill("  newer exact raw work  ");
      release();
      await expect.poll(() => reads).toBeGreaterThan(0);
      await expect(raw).toHaveValue("  newer exact raw work  ");
      await expect(raw).toBeFocused();
      await find(page, "Accepted despite refresh", 1);
      await expect(
        page.getByRole("region", { name: "Find in loaded rows" }),
      ).toContainText("Showing retained rows");
      await switchOrdinarySheet(page, timelineViewSchemaId);
      damage = false;
      const before = reads;
      await switchOrdinarySheet(page, f.view);
      await expect.poll(() => reads).toBeGreaterThan(before);
      await expect(cell(page, f.first, field)).toHaveText(
        "Accepted despite refresh",
      );
      await cell(page, f.first, field).click();
      await expect(raw).toHaveValue("  newer exact raw work  ");
      await raw.press("Escape");
      expect(writes).toBe(1);
      await find(page, "0002 needle", 1);
      await navigate(page);
      await expect(focusCell(page, f.last, field)).toBeFocused();
      expect(writes).toBe(1);
    } finally {
      release();
      await page.unroute(queryRoute);
      await page.unroute(patchRoute);
    }
  }
});

test("Entity Find supports density viewport zoom text spacing announcements and recovery menu focus on both schemas", async ({
  page,
  workerAdmin,
}, info) => {
  const preferences = await installVisualPreferences(page, workerAdmin.user_id);
  for (const schema of surfaces) {
    const f = await seed(page, schema);
    const field = `${f.type}.display_name`;
    for (const [density, width, height, zoom] of [
      ["compact", 1024, 720, 1],
      ["default", 768, 640, 1],
      ["comfortable", 1440, 900, 2],
      ["compact", 1440, 900, 1],
    ] as const) {
      preferences.select(density);
      await page.reload();
      await expect(cell(page, f.first, field)).toBeVisible();
      await page.setViewportSize({
        width: width * zoom,
        height: height * zoom,
      });
      await page.evaluate((value) => {
        document.documentElement.style.zoom = String(value);
      }, zoom);
      const spacing = await page.addStyleTag({
        content:
          "* {letter-spacing:0.12em !important;word-spacing:0.16em !important;line-height:1.5 !important;}",
      });
      await find(page, "0000 needle", 1);
      await expect(input(page)).toBeFocused();
      await expect(status(page)).not.toContainText("0000 needle");
      const panel = page.getByRole("region", { name: "Find in loaded rows" });
      const rect = await panel.boundingBox();
      if (!rect) throw new Error("Missing panel");
      expect(rect.x).toBeGreaterThanOrEqual(0);
      expect(rect.x + rect.width).toBeLessThanOrEqual(width * zoom);
      await expect(
        page.getByRole("button", { name: "Previous", exact: true }),
      ).toBeVisible();
      await expect(
        page.getByRole("button", { name: "Close Find", exact: true }),
      ).toBeVisible();
      await info.attach(`entity-find-${f.type}-${density}-${width}-${zoom}`, {
        body: await page.screenshot(),
        contentType: "image/png",
      });
      await navigate(page);
      await expect(focusCell(page, f.first, field)).toBeFocused();
      await expect(
        focusCell(page, f.first, field).getByRole("img", {
          name: "Current Find match",
          exact: true,
        }),
      ).toBeVisible();
      await spacing.evaluate((element) =>
        element.parentNode?.removeChild(element),
      );
    }
    await entry(page).click();
    await page
      .getByRole("button", {
        name: "Account and application navigation",
        exact: true,
      })
      .click();
    await expect(input(page)).toHaveCount(0);
    await page.keyboard.press("Escape");
    await expect(
      page.getByRole("button", {
        name: "Account and application navigation",
        exact: true,
      }),
    ).toBeFocused();
  }
  expect(preferences.unexpectedWrites).toEqual([]);
});

type FindMember = { email: string; initial_password: string; user_id: string };
async function authorityCase(
  page: Page,
  workerAdminRequest: APIRequestContext,
  schema: EntitySurface,
  authenticate: (member: FindMember, recovery: boolean) => Promise<void>,
) {
  const f = await seed(page, schema);
  const member = await createIncidentMemberUser(page, f.incident, {
    email: uniqueEmail("entity-find-authority"),
    display_name: "Entity Find viewer",
    initial_password: "EntityFind1!",
    role: "viewer",
    is_deployment_admin: false,
    mfa_required: false,
  });
  const login = (recovery = false) => authenticate(member, recovery);
  await login();
  const sockets = installIncidentSocketMonitor(page, f.incident);
  await page.goto(`/?incident_id=${f.incident}&view_schema_id=${f.view}`);
  await sockets.waitForAcceptedSocket();
  await find(page, "needle", 4);
  await navigate(page);
  await expect(grid(page).locator('[role="gridcell"]:focus')).toHaveAttribute(
    "aria-readonly",
    "true",
  );
  await entry(page).click();
  await input(page).fill("Protected absent term");
  await expect(status(page)).toHaveText("No matches in loaded rows.");
  await revokeAllSessions(
    workerAdminRequest,
    member.user_id,
    "Entity Find suspension evidence",
  );
  const revoked = await sockets.waitForMessage("session_revoked");
  await sockets.waitForClose(revoked.socketIndex);
  await expect(page.getByTestId(authTestId("shell"))).toHaveAttribute(
    "data-bootstrap-state",
    "revoked",
  );
  await expect(input(page)).toHaveCount(0);
  await expect(
    page.getByRole("img", { name: "Current Find match", exact: true }),
  ).toHaveCount(0);
  const start = sockets.messageCount();
  await login(true);
  await sockets.waitForAcceptedSocket({ startAt: start });
  await expect(entry(page)).toBeEnabled();
  await entry(page).click();
  await expect(input(page)).toHaveValue("");
  await expect(status(page)).toHaveText("Enter text to find in loaded rows.");
  await input(page).fill("needle");
  await expect(status(page)).toContainText("4 matching cells");
  const revokedStart = sockets.messageCount();
  const membershipResponse = await workerAdminRequest.get(
    `/api/v1/incidents/${f.incident}/memberships`,
  );
  expect(membershipResponse.ok()).toBe(true);
  const membership = (await membershipResponse.json()).data.memberships.find(
    (item: { user_id: string }) => item.user_id === member.user_id,
  );
  expect(membership).toBeDefined();
  const removed = await workerAdminRequest.delete(
    `/api/v1/incidents/${f.incident}/memberships/${member.user_id}`,
    { data: { base_membership_version: membership.membership_version } },
  );
  expect(removed.status()).toBe(204);
  await sockets.waitForMessage("session_revoked", {
    startAt: revokedStart,
    matches: (message) =>
      message.payload.reason_code === "incident_access_revoked",
  });
  await expect(page).not.toHaveURL(new RegExp(f.incident));
  await expect(input(page)).toHaveCount(0);
  await expect(
    page.getByRole("img", { name: "Find match", exact: true }),
  ).toHaveCount(0);
}
test("Entity Find retires protected host search during session suspension and reauthentication for read-only viewers", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
  await authorityCase(
    page,
    workerAdminRequest,
    { view: hostsViewSchemaId, type: "host", primary: "hostname" },
    (member, recovery) =>
      sessionTracker.loginTrackedUser(page, {
        recovery,
        createdBy: "entity-find",
        email: member.email,
        password: member.initial_password,
        purpose: "Entity Find authority",
        userId: member.user_id,
      }),
  );
});
test("Entity Find retires protected identity search during session suspension and reauthentication for read-only viewers", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
  await authorityCase(
    page,
    workerAdminRequest,
    { view: identitiesViewSchemaId, type: "identity", primary: "upn" },
    (member, recovery) =>
      sessionTracker.loginTrackedUser(page, {
        recovery,
        createdBy: "entity-find",
        email: member.email,
        password: member.initial_password,
        purpose: "Entity Find authority",
        userId: member.user_id,
      }),
  );
});
