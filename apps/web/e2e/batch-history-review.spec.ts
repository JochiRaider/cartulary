import {
  applyFilterChip,
  scrollGridCellIntoView,
} from "@cartulary/test-utils/grid";
import {
  gridScrollportSelector,
  rowCellTestId,
  rowHistoryActionTestId,
  rowHistoryItemTestId,
  rowHistoryRollbackConfirmButtonTestId,
  saveStateTestId,
  timelineInspectorTestId,
  timelineMutationSubstrateReadyTestId,
  timelineScalarEditorTestId,
  workbookColumnsMenuTestId,
  workbookColumnsMenuTriggerTestId,
} from "@cartulary/ui-contracts";
import {
  requireViewContract,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { createIncident } from "./support/incidents/fixtures";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { fetchFullRecordHistory } from "./support/workbook/history";
import {
  createViewRow,
  patchRecord,
  queryViewRows,
} from "./support/workbook/query";
import { openRecoveryItem } from "./support/workbook/recovery";

const synopsis = "timeline.activity_synopsis_text";
const source = "timeline.data_source_text";
async function reviewBatch(
  page: Page,
  family: "clear" | "paste" | "fill" | "tag",
) {
  test.setTimeout(180_000);
  const incident = await createIncident(
    page,
    uniqueIncidentKey("BATCH-REVIEW"),
    `Batch ${family} review`,
  );
  for (let index = 0; index < 3; index++)
    await createViewRow(page, incident, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("batch-review-row"),
      [synopsis]: `Review fact ${index}`,
      [source]: `Source ${index}`,
    });
  const rows = await queryViewRows(page, incident, timelineViewSchemaId);
  const first = required(rows[0]),
    second = required(rows[1]),
    third = required(rows[2]);
  await page.goto(`/?incident_id=${incident}`);
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toBeVisible();
  await page
    .getByTestId(workbookColumnsMenuTriggerTestId(timelineViewSchemaId))
    .click();
  const columns = page.getByTestId(
    workbookColumnsMenuTestId(timelineViewSchemaId),
  );
  const fields = requireViewContract(timelineViewSchemaId).fields;
  const synopsisField =
    requireViewContract(timelineViewSchemaId).fieldMap[synopsis];
  if (!synopsisField) throw new Error("Missing Synopsis fixture");
  for (const field of fields.slice(
    0,
    fields.findIndex((field) => field.fieldKey === synopsis) + 1,
  )) {
    if (field.fieldKey === "record_id" || field.fieldKey === "row_version")
      continue;
    await columns
      .getByRole("button", { name: `Width for ${field.label}`, exact: true })
      .click();
    await columns
      .getByRole("textbox", { name: "Width in CSS pixels" })
      .fill(field.fieldKey === synopsis ? "220" : "40");
    await columns
      .getByRole("button", { name: "Apply width", exact: true })
      .click();
    await columns.getByRole("button", { name: "Cancel", exact: true }).click();
  }
  await columns
    .getByRole("button", {
      name: `Freeze through ${synopsisField.label}`,
      exact: true,
    })
    .click();
  await columns
    .getByRole("button", { name: "Close columns", exact: true })
    .click();
  await expect(page.locator(gridScrollportSelector())).toHaveAttribute(
    "data-grid-freeze-state",
    "active",
  );
  const historyReads: string[] = [],
    writes: { path: string; body: string }[] = [];
  const gridPatches: string[] = [];
  page.on("request", (request) => {
    if (
      request.method() === "PATCH" &&
      new URL(request.url()).pathname.startsWith("/api/v1/records/")
    )
      gridPatches.push(request.url());
    if (new URL(request.url()).pathname.endsWith("/history"))
      historyReads.push(request.url());
    if (
      ["PATCH", "POST", "DELETE"].includes(request.method()) &&
      /\/(rollback|bulk-mutations|clipboard-paste)$/.test(
        new URL(request.url()).pathname,
      )
    )
      writes.push({ path: request.url(), body: request.postData() ?? "" });
  });
  await scrollGridCellIntoView({
    page,
    surface: timelineViewSchemaId,
    recordId: first.record_id,
    cellKey: source,
  });
  await page.getByTestId(rowCellTestId(first.record_id, source)).click();
  await page.keyboard.press("Escape");
  await page.keyboard.press("Shift+ArrowDown");
  const response = page.waitForResponse(
    (value) =>
      value.request().method() === "POST" &&
      value
        .url()
        .endsWith(family === "paste" ? "/clipboard-paste" : "/bulk-mutations"),
  );
  if (family === "clear") await page.keyboard.press("Delete");
  if (family === "fill") await page.keyboard.press("Control+d");
  if (family === "paste") {
    await page
      .context()
      .grantPermissions(["clipboard-read", "clipboard-write"]);
    await page.evaluate(() =>
      navigator.clipboard.writeText("Pasted one\nPasted two"),
    );
    await page.keyboard.press("Control+v");
  }
  if (family === "tag") {
    for (const row of [first, second])
      await page
        .getByRole("checkbox", {
          name: `Select record ${row.record_id}`,
          exact: true,
        })
        .check();
    await page
      .getByRole("textbox", { name: "Tag for selected Timeline records" })
      .fill("review-tag");
    await page.getByRole("button", { name: "Assign tag", exact: true }).click();
  }
  const accepted = await response;
  expect(accepted.ok()).toBe(true);
  const receipt = (await accepted.json()).data;
  expect(receipt.change_set_id).toBeTruthy();
  expect(receipt.rows).toHaveLength(family === "fill" ? 1 : 2);
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  expect(historyReads).toEqual([]);
  await expect(
    page.getByRole("region", { name: "Recovery navigation", exact: true }),
  ).toHaveCount(0);
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
  const requestedRecord = receipt.rows[0].record_id as string;
  let current = await fetchFullRecordHistory(page, requestedRecord);
  const original = required(
    current.items.find((item) => item.change_set_id === receipt.change_set_id),
  );
  if (family === "clear") {
    // A current live page must be searched; the batch is no longer the newest entry.
    let version = current.row_version;
    for (let index = 0; index < 101; index++) {
      version = (
        await patchRecord(page, requestedRecord, {
          client_txn_id: uniqueTxn("later-history"),
          view_schema_id: timelineViewSchemaId,
          base_row_version: version,
          changes: [
            {
              field_key: "timeline.raw_activity_text",
              value: `Later observation ${index}`,
            },
          ],
        })
      ).row_version;
    }
    await applyFilterChip(
      page,
      timelineViewSchemaId,
      "timeline.capture_state",
      "reviewed",
    );
    await expect(
      page.getByTestId(rowCellTestId(requestedRecord, synopsis)),
    ).toHaveCount(0);
    current = await fetchFullRecordHistory(page, requestedRecord);
  }
  const draft = page.getByTestId(
    timelineScalarEditorTestId({
      recordId: third.record_id,
      fieldKey: synopsis,
      surface: "grid",
    }),
  );
  if (family === "paste") {
    await scrollGridCellIntoView({
      page,
      surface: timelineViewSchemaId,
      recordId: third.record_id,
      cellKey: synopsis,
    });
    await page.getByTestId(rowCellTestId(third.record_id, synopsis)).click();
    await draft.fill("Unsubmitted review draft");
  }
  const labels = {
    clear: "Clear contents",
    paste: "Paste",
    fill: "Fill",
    tag: "Tag assignment",
  };
  await openRecoveryItem(page, new RegExp(`^${labels[family]} ·`));
  expect(historyReads).toEqual([]);
  const choices = page.getByRole("region", {
    name: "Returned Timeline records",
  });
  await expect(
    choices.getByRole("button", { name: /^Review this change:/ }),
  ).toHaveCount(family === "fill" ? 1 : 2);
  const choice = choices
    .getByRole("button", { name: /^Review this change:/ })
    .first();
  await choice.focus();
  await choice.press("Enter");
  const review = page.getByRole("region", {
    name: "Review this change",
    exact: true,
  });
  await expect(
    review.getByText("Requested change", { exact: true }),
  ).toHaveCount(
    current.items.filter((item) => item.change_set_id === receipt.change_set_id)
      .length,
  );
  await expect(review.locator(":scope > h3")).toBeFocused();
  expect(historyReads).toHaveLength(family === "clear" ? 2 : 1);
  expect(
    historyReads.every(
      (url) =>
        new URL(url).pathname === `/api/v1/records/${requestedRecord}/history`,
    ),
  ).toBe(true);
  expect(writes).toHaveLength(1);
  expect(gridPatches).toEqual([]);
  if (family === "paste")
    await expect(draft).toHaveValue("Unsubmitted review draft");
  await test.info().attach("batch-review-desktop", {
    body: await page
      .getByRole("region", { name: "Workbook recovery", exact: true })
      .screenshot(),
    contentType: "image/png",
  });
  const matching = review.getByTestId(
    rowHistoryItemTestId({ historyItemRef: original.history_item_ref }),
  );
  await expect(matching).toContainText("Requested change");
  await review.getByRole("button", { name: "Show requested entries" }).click();
  await expect(
    matching.getByText("Requested change", { exact: true }),
  ).toBeFocused();
  expect(historyReads).toHaveLength(family === "clear" ? 2 : 1);
  const reverse = matching.getByTestId(
    rowHistoryActionTestId({
      action: "change_set",
      historyItemRef: original.history_item_ref,
    }),
  );
  const currentItem = required(
    current.items.find(
      (item) => item.history_item_ref === original.history_item_ref,
    ),
  );
  if (family === "clear") {
    expect(currentItem.available_rollback_actions).not.toContain("change_set");
    await expect(reverse).toHaveCount(0);
    for (const action of currentItem.available_rollback_actions)
      await expect(
        matching.getByTestId(
          rowHistoryActionTestId({
            action,
            historyItemRef: original.history_item_ref,
          }),
        ),
      ).toBeEnabled();
    await review.getByRole("button", { name: "Close change review" }).click();
    await expect(choice).toBeFocused();
    await expect(
      page.getByTestId(rowCellTestId(requestedRecord, synopsis)),
    ).toHaveCount(0);
    expect(writes).toHaveLength(1);
    return;
  }
  await reverse.click();
  await expect(review).toContainText(
    "all reversible changes in this change set",
  );
  await expect(review).toContainText(
    "beyond the displayed row or returned record list",
  );
  expect(writes).toHaveLength(1);
  const rollbackResponse = page.waitForResponse(
    (value) =>
      value.url().endsWith(`/records/${requestedRecord}/rollback`) &&
      value.request().method() === "POST",
  );
  await review
    .getByTestId(
      rowHistoryRollbackConfirmButtonTestId({
        action: "change_set",
        historyItemRef: original.history_item_ref,
      }),
    )
    .click();
  const rolledBack = await rollbackResponse;
  expect(rolledBack.ok()).toBe(true);
  expect(JSON.parse(rolledBack.request().postData() ?? "{}")).toMatchObject({
    base_row_version: current.row_version,
    target: { kind: "change_set", change_set_id: receipt.change_set_id },
  });
  expect(writes).toHaveLength(2);
  const reversal = (await rolledBack.json()).data;
  const after = await fetchFullRecordHistory(page, requestedRecord);
  expect(after.row_version).toBeGreaterThan(current.row_version);
  expect(
    after.items.find(
      (item) => item.history_item_ref === original.history_item_ref,
    )?.change_set_id,
  ).toBe(receipt.change_set_id);
  expect(
    after.items.some(
      (item) =>
        item.change_set_id === reversal.rollback_change_set_id &&
        item.actor_user_id === original.actor_user_id,
    ),
  ).toBe(true);
  const restored = await queryViewRows(page, incident, timelineViewSchemaId);
  for (const row of [first, second]) {
    const saved = required(
      restored.find((item) => item.record_id === row.record_id),
    );
    expect(saved.cells[source]?.value).toBe(row.cells[source]?.value);
    if (family === "tag")
      expect(saved.cells["timeline.tags"]?.value).toEqual(
        row.cells["timeline.tags"]?.value,
      );
  }
  await expect(review).toBeVisible();
  await page.setViewportSize({ width: 390, height: 500 });
  await page.addStyleTag({
    content:
      "* { line-height: 1.5 !important; letter-spacing: 0.12em !important; word-spacing: 0.16em !important; } p { margin-bottom: 2em !important; }",
  });
  await expect(
    review.getByRole("button", { name: "Close change review" }),
  ).toBeEnabled();
  const panel = page.getByRole("region", {
    name: "Workbook recovery",
    exact: true,
  });
  expect(
    await panel.evaluate(
      (element) => element.scrollWidth <= element.clientWidth + 1,
    ),
  ).toBe(true);
  await test.info().attach("batch-review-narrow-spacing", {
    body: await panel.screenshot(),
    contentType: "image/png",
  });
  await review.getByRole("button", { name: "Close change review" }).click();
  await expect(choice).toBeFocused();
  expect(writes).toHaveLength(2);
  expect(gridPatches).toEqual([]);
  if (family === "paste")
    await expect(draft).toHaveValue("Unsubmitted review draft");
}
function required<T>(value: T | undefined): T {
  if (value === undefined) throw new Error("Missing batch fixture");
  return value;
}
test("Timeline clear batch reviews an exact later-page filtered record using current action metadata", async ({
  page,
}) => reviewBatch(page, "clear"));
test("Timeline paste batch reviews an explicit record and preserves attributed history", async ({
  page,
}) => reviewBatch(page, "paste"));
test("Timeline fill batch reviews recipients and preserves attributed history", async ({
  page,
}) => reviewBatch(page, "fill"));
test("Timeline tag batch reviews an explicit record and reverses all returned targets", async ({
  page,
}) => reviewBatch(page, "tag"));
