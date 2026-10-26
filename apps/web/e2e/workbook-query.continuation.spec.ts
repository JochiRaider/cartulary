import { scrollGridCellIntoView } from "@cartulary/test-utils/grid";
import { gridRowTestId, rowCellTestId } from "@cartulary/ui-contracts";
import {
  notesViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";

import { expect, test } from "./fixtures";

import { createIncident } from "./support/incidents/fixtures";

import { uniqueIncidentKey } from "./support/runtime/fixtureIdentity";
import { observePageRequests } from "./support/runtime/observations";
import { switchOrdinarySheet } from "./support/workbook/ordinaryCreate";
import { observeQuery, seed } from "./support/workbook/queryBrowsing";
import { activateCommittedGridCell } from "./support/workbook/rowMutations";

test("Workbook browsing retains off-window drafts and sheet anchors through fourth-page eviction", async ({
  page,
  testActor,
}) => {
  test.setTimeout(180_000);
  const view = notesViewSchemaId;
  const incident = await createIncident(
    page,
    uniqueIncidentKey("WQC-WINDOW"),
    "Workbook window and anchor traversal",
  );
  await seed(page, incident, view, 301, testActor.user_id);
  const reads = await observeQuery(page, incident, view);
  let writes = 0;
  observePageRequests(page, (request) => {
    if (request.method() === "PATCH" && request.url().includes("/records/"))
      writes++;
  });
  await page.goto(`/?incident_id=${incident}&view_schema_id=${view}`);
  const controls = page.getByRole("group", { name: "Workbook browsing" });
  await expect(controls).toContainText("100 records loaded; more available.");
  const original = reads.at(-1)?.response.data.rows[0];
  if (!original) throw new Error("Missing original source");
  const originalCell = page.getByTestId(
    rowCellTestId(original.record_id, "note.title"),
  );
  await originalCell.click();
  const editor = page
    .getByTestId(gridRowTestId(view, original.record_id))
    .getByRole("textbox")
    .first();
  await editor.fill("Exact retained draft after query eviction");
  const more = controls.getByRole("button", { name: "Load more", exact: true });
  for (let index = 0; index < 3; index++) {
    const previous = reads.length;
    await more.click();
    await expect.poll(() => reads.length).toBe(previous + 1);
    await expect(more).toHaveAttribute(
      "aria-disabled",
      index === 2 ? "true" : "false",
    );
    const loaded = Number(
      (await controls.innerText()).match(/(\d+) records loaded/u)?.[1],
    );
    expect(loaded).toBeLessThanOrEqual(300);
  }
  expect(writes).toBe(0);
  await expect(
    page.getByTestId(gridRowTestId(view, original.record_id)),
  ).toHaveCount(0);
  await expect(
    controls.getByRole("button", { name: "Earlier rows" }),
  ).toHaveAttribute("aria-disabled", "false");
  await controls.getByRole("button", { name: "Earlier rows" }).click();
  await expect(controls).toContainText("100 records loaded; more available.");
  await expect(
    page
      .getByTestId(gridRowTestId(view, original.record_id))
      .getByRole("textbox"),
  ).toHaveCount(0);
  await originalCell.click();
  await expect(editor).toHaveValue("Exact retained draft after query eviction");
  await editor.press("Escape");
  expect(writes).toBe(0);
  for (let index = 0; index < 3; index++) {
    const previous = reads.length;
    await more.focus();
    await more.press("Enter");
    await expect.poll(() => reads.length).toBe(previous + 1);
    await expect(controls).not.toContainText("Loading");
    await expect(more).toBeFocused();
    const loaded = Number(
      (await controls.innerText()).match(/(\d+) records loaded/u)?.[1],
    );
    expect(loaded).toBeLessThanOrEqual(300);
  }
  await expect(controls).toContainText(
    "201 records loaded; end of current results.",
  );
  await expect(more).toHaveAttribute("aria-disabled", "true");
  const last = reads.at(-1)?.response.data.rows.at(-1);
  if (!last) throw new Error("Missing terminal record");
  await scrollGridCellIntoView({
    page,
    surface: view,
    recordId: last.record_id,
    cellKey: "note.title",
  });
  const lastCell = page.getByTestId(
    rowCellTestId(last.record_id, "note.title"),
  );
  await activateCommittedGridCell(
    lastCell.locator('xpath=ancestor::*[@role="gridcell"][1]'),
  );
  await switchOrdinarySheet(page, timelineViewSchemaId);
  const beforeReturn = reads.length;
  await switchOrdinarySheet(page, view);
  await expect(controls).toContainText(
    "1 records loaded; end of current results.",
  );
  expect(reads.length).toBe(beforeReturn + 1);
  expect(reads.at(-1)?.request.cursor_token).toBe(
    reads[beforeReturn - 1]?.request.cursor_token,
  );
  await expect(lastCell).toBeVisible();
  const earlier = controls.getByRole("button", { name: "Earlier rows" });
  await earlier.focus();
  await earlier.press("Enter");
  await expect(controls).toContainText("100 records loaded; more available.");
  await expect(earlier).toBeFocused();
  const refresh = controls.getByRole("button", {
    name: "Refresh",
    exact: true,
  });
  await refresh.focus();
  await refresh.press("Enter");
  await expect(controls).toContainText("100 records loaded; more available.");
  expect(reads.at(-1)?.request.cursor_token).toBeUndefined();
  await expect(refresh).toBeFocused();
});
