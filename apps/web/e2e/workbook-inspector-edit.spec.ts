import { scrollGridTargetIntoView } from "@cartulary/test-utils/grid";
import {
  entityInspectorTestId,
  genericEditFieldSelectTestId,
  genericEditSubmitTestId,
  genericEditValueTestId,
  gridShellTestId,
  rowCellTestId,
  workbookFocusAnchorTestId,
  workbookInspectorCloseButtonTestId,
  workbookInspectorToggleTestId,
} from "@cartulary/ui-contracts";
import {
  evidenceViewSchemaId,
  hostsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { createIncident } from "./support/incidents/fixtures";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { fetchRecordHistoryCount } from "./support/workbook/history";
import { switchOrdinarySheet } from "./support/workbook/ordinaryCreate";
import {
  createViewRow,
  patchRecord,
  queryViewRows,
} from "./support/workbook/query";
import { openGenericInspectorForRecord } from "./support/workbook/rowMutations";

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
  await page
    .getByTestId(genericEditFieldSelectTestId(view))
    .selectOption(field);
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
  const options = page
    .getByTestId(genericEditFieldSelectTestId(f.view))
    .locator("option");
  expect(
    await options.evaluateAll((items) =>
      items.map((item) => (item as HTMLOptionElement).value),
    ),
  ).not.toContain("host.fqdn");
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
  await expect(input).toHaveValue("");
  await openGenericInspectorForRecord(page, f.view, f.first.record_id);
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
  await cell.dispatchEvent("mousedown", { button: 0 });
  await cell.focus();
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
