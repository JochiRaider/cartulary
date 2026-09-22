import { buildHTTPOperationPath } from "@cartulary/protocol-ts/http";
import {
  scrollGridCellIntoView,
  scrollGridTargetIntoView,
} from "@cartulary/test-utils/grid";
import {
  conflictMarkerTestId,
  draftCellTestId,
  gridFillHandleSelector,
  gridRowTestId,
  gridRowVersionAttribute,
  gridScrollportSelector,
  gridShellTestId,
  incidentLandingTestId,
  relationshipItemsTestId,
  rowCellTestId,
  rowHistoryOpenButtonTestId,
  rowHistoryPanelTestId,
  rowInspectButtonTestId,
  saveStateActionButtonTestId,
  saveStateTestId,
  timelineCollectionInputTestId,
  timelineInspectorTestId,
  timelineMutationSubstrateReadyTestId,
  timelineRowMarkReviewedButtonTestId,
  timelineRowSupersedeButtonTestId,
  timelineScalarEditorTestId,
  workbookEditRecoveryDiscardButtonTestId,
  workbookRowContextMenuTestId,
} from "@cartulary/ui-contracts";
import {
  requireViewContract,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Locator, Page } from "@playwright/test";
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
import { createTimelineFillers } from "./support/timeline/fixtures";
import { installIncidentSocketMonitor } from "./support/transport/incidentSocket";
import { holdBrowserRequest } from "./support/transport/requestInterception";
import { showTimelineCollectionColumns } from "./support/workbook/collections";
import {
  fetchFullRecordHistory,
  fetchRecordHistoryCount,
} from "./support/workbook/history";
import {
  createViewRow,
  queryViewRows,
  waitForViewRowByCell,
} from "./support/workbook/query";
import { openRecoveryItem } from "./support/workbook/recovery";

function required<T>(value: T | null | undefined): T {
  if (value === null || value === undefined)
    throw new Error("Expected fixture value is missing.");
  return value;
}

const synopsis = "timeline.activity_synopsis_text";
const source = "timeline.data_source_text";
const editor = (page: Page, recordId: string, fieldKey = synopsis) =>
  page.getByTestId(
    timelineScalarEditorTestId({ recordId, fieldKey, surface: "grid" }),
  );
const cell = (page: Page, recordId: string, fieldKey = synopsis) =>
  page
    .getByTestId(rowCellTestId(recordId, fieldKey))
    .locator("xpath=ancestor::*[@role='gridcell'][1]");

async function openTimeline(page: Page, incidentId: string) {
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toBeVisible();
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
}

async function seedTimeline(page: Page, count: number) {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("GRID-ENTRY"),
    "Timeline grid interaction regression",
  );
  for (let index = 0; index < count; index += 1)
    await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("grid-seed"),
      [synopsis]: `Fact ${index}`,
      [source]: `Source ${index}`,
    });
  await openTimeline(page, incidentId);
  const rows = await queryViewRows(page, incidentId, timelineViewSchemaId);
  return { incidentId, rows };
}

async function selectCell(page: Page, recordId: string, fieldKey = synopsis) {
  await scrollGridCellIntoView({
    page,
    surface: timelineViewSchemaId,
    recordId,
    cellKey: fieldKey,
  });
  await page.getByTestId(rowCellTestId(recordId, fieldKey)).click();
  if (fieldKey !== "timeline.capture_state")
    await expect(editor(page, recordId, fieldKey)).toBeFocused();
  await page.keyboard.press("Escape");
  await expect(cell(page, recordId, fieldKey)).toBeFocused();
}

async function clipboard(page: Page, text: string) {
  await page.context().grantPermissions(["clipboard-read", "clipboard-write"]);
  await page.evaluate((value) => navigator.clipboard.writeText(value), text);
  await page.keyboard.press("Control+v");
}

test("Timeline row actions preserve native authoring and dismiss to the semantic destination", async ({
  page,
  workerAdminRequest,
  sessionTracker,
}) => {
  test.setTimeout(180_000);
  page.setDefaultTimeout(10_000);
  const { incidentId, rows } = await seedTimeline(page, 3);
  const id = required(rows[0]).record_id;
  const otherId = required(rows[1]).record_id;
  const menu = page.getByTestId(
    workbookRowContextMenuTestId(timelineViewSchemaId, id),
  );
  const writes: string[] = [];
  page.on("request", (request) => {
    if (
      ["PATCH", "POST"].includes(request.method()) &&
      (request.url().includes("/api/v1/records/") ||
        request.url().endsWith("/bulk-mutations"))
    )
      writes.push(request.url());
  });
  await selectCell(page, id);
  await cell(page, id).click();
  const input = editor(page, id);
  await input.press("End");
  await page.keyboard.type(" native draft");
  const draft = await input.inputValue();
  for (const invoke of [
    () => input.click({ button: "right" }),
    () => input.press("Shift+F10"),
    () => input.press("ContextMenu"),
  ]) {
    await invoke();
    await expect(menu).toHaveCount(0);
    await expect(input).toBeFocused();
    await expect(input).toHaveValue(draft);
    expect(writes).toEqual([]);
  }
  // Native editing history remains intact after the context-menu gestures.
  await input.press("Control+z");
  await expect(input).not.toHaveValue(draft);
  await input.press("Control+Shift+z");
  await expect(input).toHaveValue(draft);
  await input.press("Escape");
  expect(writes).toEqual([]);

  const bulk = page.getByRole("checkbox", {
    name: `Select record ${id}`,
    exact: true,
  });
  await bulk.check();
  await selectCell(page, id);
  await page.keyboard.press("Shift+ArrowDown");
  await page.context().grantPermissions(["clipboard-read", "clipboard-write"]);
  await page.keyboard.press("Control+c");
  const selectedText = await page.evaluate(() =>
    navigator.clipboard.readText(),
  );
  expect(selectedText).toContain("\n");
  await cell(page, id).click({ button: "right" });
  await expect(menu).toBeVisible();
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
  await page.keyboard.press("Escape");
  await expect(menu).toHaveCount(0);
  await expect(cell(page, id)).toBeFocused();
  await page.keyboard.press("Control+c");
  expect(await page.evaluate(() => navigator.clipboard.readText())).toBe(
    selectedText,
  );
  await expect(bulk).toBeChecked();
  await page.keyboard.press("Shift+F10");
  await expect(menu).toBeVisible();
  await page.keyboard.press("Tab");
  await expect(menu).toHaveCount(0);
  await cell(page, id).focus();
  await page.keyboard.press("ContextMenu");
  await expect(menu).toBeVisible();
  await page.keyboard.press("Shift+Tab");
  await expect(menu).toHaveCount(0);
  await cell(page, id).focus();
  await page.keyboard.press("ContextMenu");
  await expect(menu).toBeVisible();
  await cell(page, otherId).click();
  await expect(menu).toHaveCount(0);
  await expect(editor(page, otherId)).toBeFocused();
  await editor(page, otherId).press("Escape");

  await cell(page, id).focus();
  await page.keyboard.press("Shift+F10");
  await page.getByTestId(rowInspectButtonTestId(id)).click();
  const inspector = page.getByTestId(timelineInspectorTestId());
  await expect(inspector).toBeVisible();
  await expect
    .poll(() =>
      inspector.evaluate((node) => node.contains(document.activeElement)),
    )
    .toBe(true);
  await inspector.locator(`[data-inspector-edit-field="${synopsis}"]`).click();
  const inspectorDraft = page.getByTestId(
    timelineScalarEditorTestId({
      recordId: id,
      fieldKey: synopsis,
      surface: "inspector",
    }),
  );
  await inspectorDraft.fill("Unsubmitted Inspector authoring");
  await cell(page, otherId).click({ button: "right" });
  const otherMenu = page.getByTestId(
    workbookRowContextMenuTestId(timelineViewSchemaId, otherId),
  );
  await expect(otherMenu).toBeVisible();
  await expect(inspectorDraft).toHaveValue("Unsubmitted Inspector authoring");
  expect(writes).toEqual([]);
  await page.keyboard.press("Escape");
  await expect(inspectorDraft).toHaveValue("Unsubmitted Inspector authoring");
  await inspectorDraft.press("Escape");
  await cell(page, id).focus();
  await page.keyboard.press("Shift+F10");
  await page.keyboard.press("Escape");
  await expect(inspector).toBeVisible();
  await expect(cell(page, id)).toBeFocused();
  await page.keyboard.press("Shift+F10");
  await page.getByTestId(rowHistoryOpenButtonTestId(id)).click();
  await expect(page.getByTestId(rowHistoryPanelTestId())).toBeVisible();
  await expect
    .poll(() =>
      inspector.evaluate((node) => node.contains(document.activeElement)),
    )
    .toBe(true);
  expect(writes).toEqual([]);

  await showTimelineCollectionColumns(page);
  const tagsSummary = relationshipItemsTestId(id, "timeline.tags", "grid");
  await scrollGridTargetIntoView({
    page,
    surface: timelineViewSchemaId,
    targetTestId: tagsSummary,
  });
  const tags = page
    .getByTestId(tagsSummary)
    .locator("xpath=ancestor::*[@role='gridcell'][1]");
  await tags.getByRole("button", { name: "Add tags token" }).click();
  const token = page.getByTestId(
    timelineCollectionInputTestId(id, "timeline.tags", "grid"),
  );
  await token.fill("unsubmitted native token");
  for (const invoke of [
    () => token.click({ button: "right" }),
    () => token.press("Shift+F10"),
  ]) {
    await invoke();
    await expect(menu).toHaveCount(0);
    await expect(token).toBeFocused();
    await expect(token).toHaveValue("unsubmitted native token");
    expect(writes).toEqual([]);
  }
  await page
    .getByTestId(relationshipItemsTestId(otherId, "timeline.tags", "grid"))
    .click({ button: "right" });
  expect(writes).toEqual([]);
  await expect(otherMenu).toBeVisible();
  await page.keyboard.press("Escape");
  if (!(await token.count()))
    await tags.getByRole("button", { name: "Add tags token" }).click();
  await expect(token).toHaveValue("unsubmitted native token");
  expect(writes).toEqual([]);
  await token.press("Escape");
  await expect(inspector).toBeVisible();
  expect(writes).toEqual([]);

  await selectCell(page, id);
  await page.keyboard.press("Shift+F10");
  const displayedRow = page.getByTestId(
    gridRowTestId(timelineViewSchemaId, id),
  );
  const versionBefore = await displayedRow.getAttribute(
    gridRowVersionAttribute,
  );
  await externalPatch(
    page,
    incidentId,
    id,
    source,
    "Changed during menu browsing",
  );
  await expect(displayedRow).not.toHaveAttribute(
    gridRowVersionAttribute,
    versionBefore ?? "",
  );
  await expect(menu).toBeVisible();
  await page.getByTestId(timelineRowMarkReviewedButtonTestId(id)).focus();
  const currentForReview = required(
    (await queryViewRows(page, incidentId, timelineViewSchemaId)).find(
      (row) => row.record_id === id,
    ),
  );
  const reviewed = await page.request.post(
    `${apiBase}${buildHTTPOperationPath("markTimelineRecordReviewed", { record_id: id })}`,
    {
      headers: await csrfHeaders(page),
      data: {
        base_row_version: currentForReview.row_version,
        client_txn_id: uniqueTxn("row-menu-eligibility"),
      },
    },
  );
  expect(reviewed.ok()).toBe(true);
  await expect(
    page.getByTestId(timelineRowMarkReviewedButtonTestId(id)),
  ).toBeDisabled();
  await expect(menu).toBeVisible();
  await expect(
    page.getByTestId(timelineRowSupersedeButtonTestId(id)),
  ).toBeFocused();
  await page.keyboard.press("Escape");
  await expect(cell(page, id)).toBeFocused();
  await createTimelineFillers(page, incidentId, "Row menu virtualization", 55);
  await openTimeline(page, incidentId);
  await selectCell(page, id);
  await page.keyboard.press("Shift+F10");
  await expect(menu).toBeVisible();
  const scrollport = page
    .getByTestId(gridShellTestId(timelineViewSchemaId))
    .locator(gridScrollportSelector());
  const invokingElement = await cell(page, id).elementHandle();
  const scrolled = await scrollport.evaluate((node) => {
    node.scrollTop = node.scrollTop > 200 ? 0 : node.scrollHeight;
    return node.scrollTop;
  });
  await expect(menu).toHaveCount(0);
  await expect
    .poll(() => scrollport.evaluate((node) => node.scrollTop))
    .toBe(scrolled);
  await expect(scrollport).toBeFocused();
  // RDG keeps its selected cell mounted while the root owns focus. Moving to
  // another visible record releases that retention before the semantic return.
  const loadedRows = await queryViewRows(
    page,
    incidentId,
    timelineViewSchemaId,
  );
  const distantId = required(loadedRows.at(-1)).record_id;
  await selectCell(
    page,
    distantId === id ? required(loadedRows[0]).record_id : distantId,
  );
  await expect
    .poll(() => invokingElement?.evaluate((node) => node.isConnected))
    .toBe(false);
  await selectCell(page, id);
  await page.keyboard.press("Shift+F10");
  const current = required(
    (await queryViewRows(page, incidentId, timelineViewSchemaId)).find(
      (row) => row.record_id === id,
    ),
  );
  const removed = await page.request.delete(`${apiBase}/api/v1/records/${id}`, {
    headers: await csrfHeaders(page),
    data: {
      base_row_version: current.row_version,
      client_txn_id: uniqueTxn("row-menu-remove"),
    },
  });
  expect(removed.ok()).toBe(true);
  await expect(menu).toHaveCount(0);
  await expect(scrollport).toBeFocused();

  const member = await createIncidentMemberUser(page, incidentId, {
    display_name: "Menu authority reviewer",
    email: uniqueEmail("row-menu-reviewer"),
    initial_password: "RowMenuEditor1!",
    role: "reviewer",
    mfa_required: false,
    is_deployment_admin: false,
  });
  await sessionTracker.loginTrackedUser(page, {
    createdBy: "timeline-grid-entry",
    purpose: "Row menu authority invalidation",
    email: member.email,
    password: member.initial_password,
    userId: member.user_id,
  });
  const socket = installIncidentSocketMonitor(page, incidentId);
  await openTimeline(page, incidentId);
  await socket.waitForAcceptedSocket();
  await selectCell(page, otherId);
  await page.keyboard.press("Shift+F10");
  await expect(otherMenu).toBeVisible();
  await expect(
    page.getByTestId(timelineRowMarkReviewedButtonTestId(otherId)),
  ).toBeEnabled();
  const membershipPath = `/api/v1/incidents/${incidentId}/memberships/${member.user_id}`;
  expect(
    (
      await workerAdminRequest.patch(membershipPath, {
        data: { base_membership_version: 1, role: "viewer" },
      })
    ).ok(),
  ).toBe(true);
  // Role downgrades are observed through the existing action authorization
  // revalidation; no menu-specific polling or authority source is introduced.
  const denied = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response.url().includes(`/records/${otherId}/`),
  );
  await page.getByTestId(timelineRowMarkReviewedButtonTestId(otherId)).click();
  expect((await denied).status()).toBe(403);
  await expect(otherMenu).toHaveCount(0);
  await expect(page.getByTestId(timelineInspectorTestId())).toBeVisible();
  await cell(page, otherId).focus();
  await page.keyboard.press("Shift+F10");
  await expect(otherMenu).toBeVisible();
  await expect(
    page.getByTestId(timelineRowMarkReviewedButtonTestId(otherId)),
  ).toBeDisabled();
  await expect(
    page.getByTestId(timelineRowSupersedeButtonTestId(otherId)),
  ).toBeDisabled();
  expect(
    (
      await workerAdminRequest.delete(membershipPath, {
        data: { base_membership_version: 2 },
      })
    ).status(),
  ).toBe(204);
  await expect(page.getByTestId(incidentLandingTestId("shell"))).toBeVisible();
  await expect(otherMenu).toHaveCount(0);
  await expect(scrollport).toHaveCount(0);
});

test("Timeline native clipboard preserves scalar and rectangular values and rejects malformed representations", async ({
  page,
}) => {
  const { incidentId, rows } = await seedTimeline(page, 4);
  const ids = rows.map((row) => row.record_id);
  const first = required(ids[0]);
  const second = required(ids[1]);
  const third = required(ids[2]);
  const fourth = required(ids[3]);
  await externalPatch(page, incidentId, first, synopsis, "a,b");
  await externalPatch(page, incidentId, second, synopsis, 'c,"d"');
  await externalPatch(page, incidentId, first, source, "00123");
  await externalPatch(page, incidentId, second, source, "2026-09-15");
  const raw = "timeline.raw_activity_text";
  const richScalar = '=SUM(A1)\t"quoted"\n界😀 e\u0301\r\nend';
  await externalPatch(page, incidentId, first, raw, richScalar);
  await openTimeline(page, incidentId);
  await page.context().grantPermissions(["clipboard-read", "clipboard-write"]);
  const attempts: string[] = [];
  page.on("request", (request) => {
    if (request.url().endsWith("/clipboard-paste"))
      attempts.push(required(request.postData()));
  });
  // Real browser copy and paste; the one-column comma regression.
  await selectCell(page, first);
  await page.keyboard.press("Shift+ArrowDown");
  await page.keyboard.press("Control+c");
  const copied = await page.evaluate(async () => {
    const item = (await navigator.clipboard.read())[0];
    return item
      ? {
          plain: await (await item.getType("text/plain")).text(),
          html: await (await item.getType("text/html")).text(),
        }
      : null;
  });
  expect(copied?.plain).toBe('a,b\n"c,""d"""');
  expect(copied?.html).toContain('data-cartulary-clipboard="1"');
  await selectCell(page, third);
  await page.keyboard.press("Control+v");
  await expect
    .poll(
      async () =>
        (await queryViewRows(page, incidentId, timelineViewSchemaId)).find(
          (row) => row.record_id === fourth,
        )?.cells[synopsis]?.value,
    )
    .toBe('c,"d"');
  let persisted = await queryViewRows(page, incidentId, timelineViewSchemaId);
  expect(
    persisted.find((row) => row.record_id === third)?.cells[source]?.value,
  ).toBe(required(rows[2]).cells[source]?.value);
  expect(attempts).toHaveLength(1);
  expect(JSON.parse(required(attempts[0]))).toMatchObject({
    format: "tsv",
    header_mode: "none",
    columns: [synopsis],
  });
  // Two columns, leading zeros and date-looking strings.
  await selectCell(page, first);
  await page.keyboard.press("Shift+ArrowRight");
  await page.keyboard.press("Shift+ArrowDown");
  await page.keyboard.press("Control+c");
  await selectCell(page, third);
  await page.keyboard.press("Control+v");
  await expect
    .poll(
      async () =>
        (await queryViewRows(page, incidentId, timelineViewSchemaId)).find(
          (row) => row.record_id === fourth,
        )?.cells[source]?.value,
    )
    .toBe("2026-09-15");
  persisted = await queryViewRows(page, incidentId, timelineViewSchemaId);
  expect(
    persisted.find((row) => row.record_id === third)?.cells[source]?.value,
  ).toBe("00123");
  // Marked scalar reverses the export apostrophe, keeping formulas as text.
  const sourceRaw = persisted.find((row) => row.record_id === first)?.cells[raw]
    ?.value;
  await selectCell(page, first, raw);
  await page.keyboard.press("Control+c");
  expect(await page.evaluate(() => navigator.clipboard.readText())).toBe(
    `'${sourceRaw}`,
  );
  await selectCell(page, third, raw);
  await page.keyboard.press("Control+v");
  await expect
    .poll(
      async () =>
        (await queryViewRows(page, incidentId, timelineViewSchemaId)).find(
          (row) => row.record_id === third,
        )?.cells[raw]?.value,
    )
    .toBe(sourceRaw);
  // Active editors keep native selection/caret semantics and cancellation.
  await page.getByTestId(rowCellTestId(third, raw)).click();
  await expect(editor(page, third, raw)).toBeFocused();
  await page.keyboard.press("Control+a");
  await clipboard(page, 'native, "literal"');
  await expect(editor(page, third, raw)).toHaveValue('native, "literal"');
  await page.keyboard.press("Escape");
  const before = await queryViewRows(page, incidentId, timelineViewSchemaId);
  const attemptCount = attempts.length;
  for (const offered of [
    { "text/csv": '"unterminated', "text/plain": "must not fall back" },
    { "text/plain": "one\ttwo\nshort" },
    {
      "text/html": '<table><tr><td colspan="2">merged</td></tr></table>',
      "text/plain": "must not fall back",
    },
    {
      "text/html":
        '<table data-cartulary-clipboard="99"><tr><td>unknown</td></tr></table>',
    },
    { "text/plain": "x".repeat(8_388_609) },
  ]) {
    await cell(page, third, raw).evaluate((element, representations) => {
      const data = new DataTransfer();
      for (const [type, value] of Object.entries(representations))
        if (value !== undefined) data.setData(type, value);
      element.dispatchEvent(
        new ClipboardEvent("paste", {
          bubbles: true,
          cancelable: true,
          clipboardData: data,
        }),
      );
    }, offered);
    await expect(
      page.locator('.cartulary-grid-live-region[role="alert"]'),
    ).toBeAttached();
    await expect(cell(page, third, raw)).toBeFocused();
  }
  expect(attempts).toHaveLength(attemptCount);
  expect(await queryViewRows(page, incidentId, timelineViewSchemaId)).toEqual(
    before,
  );
  // An explicit empty copied cell clears through the destination field contract.
  await selectCell(page, fourth, raw);
  await page.keyboard.press("Control+c");
  await selectCell(page, third, raw);
  await page.keyboard.press("Control+v");
  await expect
    .poll(
      async () =>
        (await queryViewRows(page, incidentId, timelineViewSchemaId)).find(
          (row) => row.record_id === third,
        )?.cells[raw]?.value,
    )
    .toBe("");
  await selectCell(page, third);
  await page.evaluate(async () =>
    navigator.clipboard.write([
      new ClipboardItem({
        "text/html": new Blob(
          [
            "<table><tr><td><span>a<br>b</span></td><td>0042</td></tr><tr><td>=1+1</td><td></td></tr></table>",
          ],
          { type: "text/html" },
        ),
        "text/plain": new Blob(["unused fallback"], { type: "text/plain" }),
      }),
    ]),
  );
  await page.keyboard.press("Control+v");
  await expect
    .poll(
      async () =>
        (await queryViewRows(page, incidentId, timelineViewSchemaId)).find(
          (row) => row.record_id === fourth,
        )?.cells[synopsis]?.value,
    )
    .toBe("=1+1");
  persisted = await queryViewRows(page, incidentId, timelineViewSchemaId);
  expect(
    persisted.find((row) => row.record_id === third)?.cells[synopsis]?.value,
  ).toBe("a\nb");
  expect(
    persisted.find((row) => row.record_id === third)?.cells[source]?.value,
  ).toBe("0042");
  expect(
    persisted.find((row) => row.record_id === fourth)?.cells[source]?.value,
  ).toBe("");
});

async function tabTo(page: Page, target: Locator) {
  for (let step = 0; step < 128; step += 1) {
    if (
      await target.evaluateAll((elements) =>
        elements.some((element) => element === document.activeElement),
      )
    )
      return;
    await page.keyboard.press("Tab");
    await page.evaluate(
      () =>
        new Promise<void>((resolve) => requestAnimationFrame(() => resolve())),
    );
  }
  await expect(target).toBeFocused();
}

async function externalPatch(
  page: Page,
  incidentId: string,
  recordId: string,
  fieldKey: string,
  value: string,
) {
  const current = required(
    (await queryViewRows(page, incidentId, timelineViewSchemaId)).find(
      (row) => row.record_id === recordId,
    ),
  );
  const response = await page.request.patch(
    `${apiBase}/api/v1/records/${recordId}`,
    {
      headers: await csrfHeaders(page),
      data: {
        view_schema_id: timelineViewSchemaId,
        client_txn_id: uniqueTxn("concurrent-edit"),
        base_row_version: current.row_version,
        changes: [{ field_key: fieldKey, value }],
      },
    },
  );
  expect(response.ok()).toBeTruthy();
}

test("Timeline pointer transitions preserve one edit and wait for acceptance", async ({
  page,
}) => {
  const { incidentId, rows } = await seedTimeline(page, 2);
  const id = required(rows[0]).record_id;
  await selectCell(page, id);
  const requests: string[] = [];
  page.on("request", (request) => {
    if (
      request.method() === "PATCH" &&
      request.url().endsWith(`/records/${id}`)
    )
      requests.push(request.postData() ?? "");
  });
  await page.getByTestId(rowCellTestId(id, synopsis)).click();
  await expect(editor(page, id)).toBeFocused();
  expect(
    await editor(page, id).evaluate(
      (element: HTMLInputElement | HTMLTextAreaElement) => ({
        start: element.selectionStart,
        end: element.selectionEnd,
        length: element.value.length,
      }),
    ),
  ).toEqual({
    start: String(required(rows[0]).cells[synopsis]?.value).length,
    end: String(required(rows[0]).cells[synopsis]?.value).length,
    length: String(required(rows[0]).cells[synopsis]?.value).length,
  });
  await page.keyboard.press("Escape");
  await page.getByTestId(rowCellTestId(id, synopsis)).dblclick();
  await expect(editor(page, id)).toBeFocused();
  // A second click may reposition the caret, but never starts another session.
  await page.keyboard.press("End");
  await page.keyboard.type(" pointer");
  const held = await holdBrowserRequest(page, {
    method: "PATCH",
    path: `/api/v1/records/${id}`,
  });
  try {
    await page.getByTestId(rowCellTestId(id, source)).click();
    await held.waitForHit;
    await expect(editor(page, id)).toHaveValue(
      `${required(rows[0]).cells[synopsis]?.value} pointer`,
    );
    await expect(editor(page, id, source)).toHaveCount(0);
    held.release();
    await expect(editor(page, id, source)).toBeFocused();
    expect(requests).toHaveLength(1);
    await page.keyboard.press("Escape");
    await expect(cell(page, id, source)).toBeFocused();
    await waitForViewRowByCell(
      page,
      incidentId,
      timelineViewSchemaId,
      synopsis,
      `${required(rows[0]).cells[synopsis]?.value} pointer`,
    );
    await page
      .getByRole("checkbox", { name: `Select record ${id}`, exact: true })
      .click();
    await expect(editor(page, id)).toHaveCount(0);
    expect(requests).toHaveLength(1);
  } finally {
    await held.dispose();
  }
  await selectCell(page, id);
  await page.keyboard.type("Accepted after Escape");
  const late = await holdBrowserRequest(page, {
    method: "PATCH",
    path: `/api/v1/records/${id}`,
  });
  try {
    await page.keyboard.press("Tab");
    await late.waitForHit;
    await page.keyboard.press("Escape");
    await expect(cell(page, id)).toBeFocused();
    await page.keyboard.press("ArrowDown");
    await expect(cell(page, required(rows[1]).record_id)).toBeFocused();
    late.release();
    await waitForViewRowByCell(
      page,
      incidentId,
      timelineViewSchemaId,
      synopsis,
      "Accepted after Escape",
    );
    await expect(cell(page, required(rows[1]).record_id)).toBeFocused();
    expect(requests).toHaveLength(2);
    expect(await fetchRecordHistoryCount(page, id)).toBe(3);
  } finally {
    await late.dispose();
  }
});

test("Timeline rectangle paste keyboard fill and pointer fill preserve targets", async ({
  page,
}) => {
  const { incidentId, rows } = await seedTimeline(page, 3);
  const [first, second, third] = rows.map((row) => row.record_id) as [
    string,
    string,
    string,
  ];
  await selectCell(page, first);
  await clipboard(page, "Pasted one\tShared source\nPasted two\tSecond source");
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  await waitForViewRowByCell(
    page,
    incidentId,
    timelineViewSchemaId,
    synopsis,
    "Pasted two",
  );
  await selectCell(page, first, source);
  await page.keyboard.press("Shift+ArrowDown");
  await page.keyboard.press("Control+d");
  await expect(page.getByTestId(rowCellTestId(second, source))).toHaveText(
    "Shared source",
  );
  await expect(cell(page, second, source)).toBeFocused();
  await selectCell(page, first, source);
  const handle = page.locator(gridFillHandleSelector());
  await expect(handle).toHaveAttribute("aria-label", "Drag to fill this value");
  await handle.scrollIntoViewIfNeeded();
  const start = required(await handle.boundingBox());
  const end = required(await cell(page, third, source).boundingBox());
  expect(
    await handle.evaluate((element) => {
      const rect = element.getBoundingClientRect();
      return (
        document
          .elementFromPoint(rect.x + rect.width / 2, rect.y + rect.height / 2)
          ?.closest('[data-cartulary-fill-handle="true"]') === element
      );
    }),
  ).toBe(true);
  await page.mouse.move(start.x + start.width / 2, start.y + start.height / 2);
  await page.mouse.down();
  await page.mouse.move(end.x + end.width / 2, end.y + end.height / 2, {
    steps: 8,
  });
  await page.mouse.up();
  await expect(page.getByTestId(rowCellTestId(third, source))).toHaveText(
    "Shared source",
  );
  const persisted = await queryViewRows(page, incidentId, timelineViewSchemaId);
  for (const id of [first, second, third])
    expect(
      persisted.find((row) => row.record_id === id)?.cells[source]?.value,
    ).toBe("Shared source");
  expect(
    persisted.find((row) => row.record_id === first)?.cells[synopsis]?.value,
  ).toBe("Pasted one");
  expect(
    persisted.find((row) => row.record_id === second)?.cells[synopsis]?.value,
  ).toBe("Pasted two");
  await selectCell(page, first, source);
  let fills = 0;
  page.on("request", (request) => {
    if (request.url().endsWith("/bulk-mutations")) fills += 1;
  });
  await page.keyboard.press("Shift+ArrowLeft");
  await page.keyboard.press("Shift+ArrowDown");
  await page.keyboard.press("Control+d");
  await expect(
    page
      .getByRole("alert")
      .filter({ hasText: "Select a writable one-column range" }),
  ).toHaveText("Select a writable one-column range before using fill down.");
  await page.evaluate(
    () =>
      new Promise<void>((resolve) =>
        requestAnimationFrame(() => requestAnimationFrame(() => resolve())),
      ),
  );
  expect(fills).toBe(0);
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
});

test("Timeline rejected edits keep correction local and preserve rough date text", async ({
  page,
}) => {
  const { incidentId, rows } = await seedTimeline(page, 2);
  const id = required(rows[0]).record_id;
  await selectCell(page, id);
  await page.keyboard.type("Unsaved");
  await editor(page, id).fill("Invalid\u0001text");
  await page.keyboard.press("Tab");
  await expect(editor(page, id)).toBeFocused();
  await expect(editor(page, id)).toHaveValue("Invalid\u0001text");
  expect(
    (await queryViewRows(page, incidentId, timelineViewSchemaId)).find(
      (row) => row.record_id === id,
    )?.cells[synopsis]?.value,
  ).toBe(required(rows[0]).cells[synopsis]?.value);
  await page.getByTestId(saveStateActionButtonTestId()).click();
  await page.getByTestId(workbookEditRecoveryDiscardButtonTestId()).click();
  await page
    .getByRole("button", { name: "Close recovery", exact: true })
    .click();
  await selectCell(page, id);
  await page.keyboard.type("Corrected fact");
  await page.keyboard.press("Enter");
  await waitForViewRowByCell(
    page,
    incidentId,
    timelineViewSchemaId,
    synopsis,
    "Corrected fact",
  );
  await selectCell(page, id, "timeline.activity_utc_text");
  await page.keyboard.type("not-a-timestamp");
  await page.keyboard.press("Enter");
  await waitForViewRowByCell(
    page,
    incidentId,
    timelineViewSchemaId,
    "timeline.activity_utc_text",
    "not-a-timestamp",
  );
  await selectCell(page, id);
  await page.keyboard.type("My conflict draft");
  const held = await holdBrowserRequest(page, {
    method: "PATCH",
    path: `/api/v1/records/${id}`,
  });
  try {
    await page.keyboard.press("Tab");
    await held.waitForHit;
    await externalPatch(
      page,
      incidentId,
      id,
      synopsis,
      "Concurrent saved fact",
    );
    held.release();
    await expect(editor(page, id)).toHaveValue("My conflict draft");
    await expect(editor(page, id)).toBeFocused();
    await expect(
      page.getByTestId(conflictMarkerTestId(id, synopsis)),
    ).toBeVisible();
    expect(
      (await queryViewRows(page, incidentId, timelineViewSchemaId)).find(
        (row) => row.record_id === id,
      )?.cells[synopsis]?.value,
    ).toBe("Concurrent saved fact");
    await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
    await page.getByTestId(conflictMarkerTestId(id, synopsis)).click();
    await page
      .getByRole("button", { name: "Close conflict recovery", exact: true })
      .click();
    await page.getByTestId(conflictMarkerTestId(id, synopsis)).click();
    await page
      .getByRole("button", { name: "Discard local draft", exact: true })
      .click();
    await expect(page.getByTestId(rowCellTestId(id, synopsis))).toHaveText(
      "Concurrent saved fact",
    );
    await expect(cell(page, id)).toBeFocused();
    await page.keyboard.type("Continued after conflict");
    await page.keyboard.press("Enter");
    await waitForViewRowByCell(
      page,
      incidentId,
      timelineViewSchemaId,
      synopsis,
      "Continued after conflict",
    );
  } finally {
    await held.dispose();
  }
});

test("Timeline active drafts survive in-app refresh and stale refresh failure", async ({
  page,
}) => {
  const { incidentId, rows } = await seedTimeline(page, 3);
  const id = required(rows.at(-1)).record_id;
  await selectCell(page, id);
  await page.keyboard.type("Draft kept through refresh");
  const scrollport = page
    .getByTestId(gridShellTestId(timelineViewSchemaId))
    .locator(gridScrollportSelector());
  const before = await scrollport.evaluate((element) => ({
    top: element.scrollTop,
    left: element.scrollLeft,
  }));
  const queryPath = `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/query`;
  let releaseRefresh!: () => void;
  let queryCaptured!: () => void;
  const refreshGate = new Promise<void>((resolve) => {
    releaseRefresh = resolve;
  });
  const captured = new Promise<void>((resolve) => {
    queryCaptured = resolve;
  });
  await page.route(
    `**${queryPath}`,
    async (route) => {
      const response = await route.fetch();
      queryCaptured();
      await refreshGate;
      await route.fulfill({ response });
    },
    { times: 1 },
  );
  try {
    await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("live-create"),
      [synopsis]: "Concurrent new row",
    });
    await captured;
    await externalPatch(
      page,
      incidentId,
      id,
      source,
      "Newer source during refresh",
    );
    await expect(page.getByTestId(rowCellTestId(id, source))).toHaveText(
      "Newer source during refresh",
    );
    await expect(editor(page, id)).toHaveValue("Draft kept through refresh");
    await expect(editor(page, id)).toBeFocused();
    releaseRefresh();
    await expect(
      page.locator('[data-grid-data-state="refreshing"]'),
    ).toHaveCount(0);
    await expect(editor(page, id)).toHaveValue("Draft kept through refresh");
    await expect(editor(page, id)).toBeFocused();
    expect(
      await scrollport.evaluate((element) => ({
        top: element.scrollTop,
        left: element.scrollLeft,
      })),
    ).toEqual(before);
    await expect(page.getByTestId(rowCellTestId(id, source))).toHaveText(
      "Newer source during refresh",
    );
  } finally {
    releaseRefresh();
  }
  await page.keyboard.press("Home");
  await page.keyboard.press("ArrowRight");
  // An authoritative sort-key update moves the active record from the end
  // to the beginning of the query while its synopsis remains uncommitted.
  await externalPatch(
    page,
    incidentId,
    id,
    "timeline.activity_utc_text",
    "2026-04-10T12:00:00Z",
  );
  await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("reorder-refresh"),
    [synopsis]: "Concurrent sort refresh",
  });
  await expect(
    page.locator('[role="row"][data-grid-record-id]').first(),
  ).toHaveAttribute("data-grid-record-id", id);
  await expect(editor(page, id)).toHaveValue("Draft kept through refresh");
  await expect(editor(page, id)).toBeFocused();
  expect(
    await editor(page, id).evaluate(
      (element: HTMLInputElement | HTMLTextAreaElement) =>
        element.selectionStart,
    ),
  ).toBe(1);
  const reordered = await queryViewRows(page, incidentId, timelineViewSchemaId);
  expect(required(reordered[0]).record_id).toBe(id);
  expect(required(reordered[0]).cells[synopsis]?.value).toBe(
    required(rows.at(-1)).cells[synopsis]?.value,
  );
  await page.route(
    `**${queryPath}`,
    async (route) =>
      route.fulfill({
        status: 503,
        contentType: "application/json",
        body: JSON.stringify({
          error: {
            code: "internal_error",
            message: "Temporary refresh failure",
          },
        }),
      }),
    { times: 1 },
  );
  const failedRefresh = page.waitForResponse(
    (response) =>
      response.url().endsWith(queryPath) && response.status() === 503,
  );
  await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("live-create-failure"),
    [synopsis]: "Another concurrent row",
  });
  await failedRefresh;
  await expect(editor(page, id)).toHaveValue("Draft kept through refresh");
  await expect(editor(page, id)).toBeFocused();
  await page.keyboard.press("Enter");
  await waitForViewRowByCell(
    page,
    incidentId,
    timelineViewSchemaId,
    synopsis,
    "Draft kept through refresh",
  );
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
});

test("Timeline virtualization retains the draft and semantic cell without mounting every row", async ({
  page,
}) => {
  const { incidentId, rows } = await seedTimeline(page, 70);
  const id = required(rows[0]).record_id;
  await selectCell(page, id);
  await page.keyboard.type("Virtualized retained draft");
  const scrollport = page
    .getByTestId(gridShellTestId(timelineViewSchemaId))
    .locator(gridScrollportSelector());
  await scrollport.hover();
  await page.mouse.wheel(0, 100_000);
  // RDG pins one active row to preserve its editor, while ordinary rows unmount.
  await expect(
    page.locator(
      `[role="row"][data-grid-record-id="${required(rows[1]).record_id}"]`,
    ),
  ).toHaveCount(0);
  await expect
    .poll(() => page.locator('[role="row"][data-grid-record-id]').count())
    .toBeLessThan(40);
  expect(
    await editor(page, id).evaluate((element) => {
      const viewport = element
        .closest('[role="grid"]')
        ?.getBoundingClientRect();
      const rect = element.getBoundingClientRect();
      return viewport !== undefined && rect.bottom <= viewport.top;
    }),
  ).toBe(true);
  await page.mouse.wheel(0, -100_000);
  await expect
    .poll(() => scrollport.evaluate((element) => element.scrollTop))
    .toBe(0);
  await expect(editor(page, id)).toHaveValue("Virtualized retained draft");
  await expect(editor(page, id)).toBeFocused();
  await page.keyboard.type(" continued");
  await page.keyboard.press("Enter");
  await waitForViewRowByCell(
    page,
    incidentId,
    timelineViewSchemaId,
    synopsis,
    "Virtualized retained draft continued",
  );
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
});

test("Timeline spreadsheet keys commit and navigate across multiple rows", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("GRID-KEYS"),
    "Timeline spreadsheet keyboard capture",
  );
  for (let index = 0; index < 3; index += 1) {
    await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("grid-key-seed"),
      [synopsis]: `Seed ${index}`,
    });
  }
  await openTimeline(page, incidentId);
  const rows = await queryViewRows(page, incidentId, timelineViewSchemaId);
  const first = required(rows[0]).record_id;
  await tabTo(page, cell(page, first));
  for (let index = 0; index < rows.length; index += 1) {
    const id = required(rows[index]).record_id;
    await expect(cell(page, id)).toBeFocused();
    await page.keyboard.type(`Keyboard row ${index}`);
    await expect(editor(page, id)).toHaveValue(`Keyboard row ${index}`);
    await page.keyboard.press("Tab");
    await expect(cell(page, id, source)).toBeFocused();
    await page.keyboard.type(`Source ${index}`);
    await page.keyboard.press("Shift+Tab");
    await expect(cell(page, id)).toBeFocused();
    await page.keyboard.press("Enter");
  }
  const summaryDraft = page.getByTestId(draftCellTestId(synopsis));
  await expect(summaryDraft).toBeFocused();
  await page.keyboard.press("Shift+Enter");
  await expect(summaryDraft).toBeFocused();
  await expect(summaryDraft).toHaveValue("\n");
  await summaryDraft.fill("");
  await page.keyboard.press("Tab");
  await expect(page.getByTestId(draftCellTestId(source))).toBeFocused();
  await page.keyboard.press("Shift+Enter");
  await expect(cell(page, required(rows[2]).record_id, source)).toBeFocused();
  await page.keyboard.press("ArrowLeft");
  await expect(cell(page, required(rows[2]).record_id)).toBeFocused();
  await page.keyboard.press("ArrowUp");
  await expect(cell(page, required(rows[1]).record_id)).toBeFocused();
  const fields = requireViewContract(timelineViewSchemaId).defaultVisibleFields;
  const firstField = required(fields[0]);
  const lastField = required(fields.at(-1));
  await page.keyboard.press("End");
  await expect(
    cell(page, required(rows[1]).record_id, lastField),
  ).toBeFocused();
  await page.keyboard.press("Tab");
  await expect(
    cell(page, required(rows[2]).record_id, firstField),
  ).toBeFocused();
  await page.keyboard.press("Shift+Tab");
  await expect(
    cell(page, required(rows[1]).record_id, lastField),
  ).toBeFocused();
  await page.keyboard.press("Enter");
  await page.keyboard.press("Tab");
  const draftFields = fields.filter(
    (field) =>
      requireViewContract(timelineViewSchemaId).fieldMap[field]?.writeKind !==
      "read_only",
  );
  for (const field of draftFields) {
    await expect(page.getByTestId(draftCellTestId(field))).toBeFocused();
    await page.keyboard.press("Tab");
  }
  await expect
    .poll(() =>
      page
        .getByTestId(gridShellTestId(timelineViewSchemaId))
        .evaluate((element) => element.contains(document.activeElement)),
    )
    .toBe(false);
  expect(
    await queryViewRows(page, incidentId, timelineViewSchemaId),
  ).toHaveLength(3);
  const saved = await queryViewRows(page, incidentId, timelineViewSchemaId);
  for (let index = 0; index < rows.length; index += 1) {
    const row = required(
      saved.find(
        (candidate) => candidate.record_id === required(rows[index]).record_id,
      ),
    );
    expect(row.cells[synopsis]?.value).toBe(`Keyboard row ${index}`);
    expect(row.cells[source]?.value).toBe(`Source ${index}`);
  }
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
});

test("Timeline first input creates once and retains typing through acknowledgement", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("GRID-CREATE"),
    "Timeline continuous rough capture",
  );
  await openTimeline(page, incidentId);
  const path = `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/rows`;
  const held = await holdBrowserRequest(page, { method: "POST", path });
  const submissions: unknown[] = [];
  page.on("request", (request) => {
    if (
      ["POST", "PATCH"].includes(request.method()) &&
      (request.url().endsWith(path) ||
        request.url().includes("/api/v1/records/"))
    )
      submissions.push({
        method: request.method(),
        payload: request.postDataJSON(),
      });
  });
  try {
    const draft = page.getByTestId(draftCellTestId(synopsis));
    await scrollGridTargetIntoView({
      page,
      surface: timelineViewSchemaId,
      targetTestId: draftCellTestId(synopsis),
    });
    await tabTo(page, draft);
    const originalDraftId = await draft.getAttribute("id");
    await page.keyboard.type("First");
    await held.waitForHit;
    await page.keyboard.press("Enter");
    await page.keyboard.type(" continuous fact");
    await expect(draft).toHaveValue("First continuous fact");
    expect(held.hitCount()).toBe(1);
    await page.keyboard.press("Home");
    await page.keyboard.press("ArrowRight");
    expect(
      await draft.evaluate(
        (element: HTMLTextAreaElement) => element.selectionStart,
      ),
    ).toBe(1);
    held.release();
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    const rows = await queryViewRows(page, incidentId, timelineViewSchemaId);
    expect(rows).toHaveLength(1);
    const id = required(rows[0]).record_id;
    await expect(editor(page, id)).toBeFocused();
    await expect(editor(page, id)).toHaveValue("First continuous fact");
    expect(
      await editor(page, id).evaluate(
        (element: HTMLTextAreaElement) => element.selectionStart,
      ),
    ).toBe(1);
    await page.keyboard.press("End");
    await page.keyboard.type(" continued");
    await page.keyboard.press("Enter");
    await expect(page.getByTestId(draftCellTestId(synopsis))).toBeFocused();
    await expect(editor(page, id)).toHaveCount(0);
    await expect(
      page.getByTestId(draftCellTestId(synopsis)),
    ).not.toHaveAttribute("id", required(originalDraftId));
    await waitForViewRowByCell(
      page,
      incidentId,
      timelineViewSchemaId,
      synopsis,
      "First continuous fact continued",
    );
    expect(
      await queryViewRows(page, incidentId, timelineViewSchemaId),
    ).toHaveLength(1);
    for (const label of ["Second keyboard capture", "Third keyboard capture"]) {
      const holdDuplicate = label.startsWith("Second");
      const createGate = holdDuplicate
        ? await holdBrowserRequest(page, { method: "POST", path })
        : null;
      const patchGate = holdDuplicate
        ? await holdBrowserRequest(page, {
            method: "PATCH",
            path: "/api/v1/records/*",
          })
        : null;
      try {
        await expect(page.getByTestId(draftCellTestId(synopsis))).toBeFocused();
        await page.keyboard.type(label);
        if (createGate !== null && patchGate !== null) {
          await createGate.waitForHit;
          createGate.release();
          await patchGate.waitForHit;
          const capturing = required(
            (await queryViewRows(page, incidentId, timelineViewSchemaId)).find(
              (row) => row.record_id !== id,
            ),
          );
          await expect(editor(page, capturing.record_id)).toHaveValue(label);
          await expect(editor(page, capturing.record_id)).toBeFocused();
        }
        // Tab must join the identical follow-on patch already in flight.
        await page.keyboard.press("Tab");
        patchGate?.release();
        const created = await waitForViewRowByCell(
          page,
          incidentId,
          timelineViewSchemaId,
          synopsis,
          label,
        );
        expect(created.record_id, JSON.stringify(submissions)).not.toBe(id);
        await expect(cell(page, created.record_id, source)).toBeFocused();
        await expect(editor(page, created.record_id)).toHaveCount(0);
        if (patchGate !== null) expect(patchGate.hitCount()).toBe(1);
        await page.keyboard.type(`Source for ${label}`);
        await page.keyboard.press("Shift+Tab");
        await expect(cell(page, created.record_id)).toBeFocused();
        await page.keyboard.press("Enter");
        await waitForViewRowByCell(
          page,
          incidentId,
          timelineViewSchemaId,
          source,
          `Source for ${label}`,
        );
      } finally {
        await createGate?.dispose();
        await patchGate?.dispose();
      }
    }
    expect(
      await queryViewRows(page, incidentId, timelineViewSchemaId),
    ).toHaveLength(3);
    await expect(page.getByTestId(draftCellTestId(synopsis))).toBeFocused();
    const navigationGate = await holdBrowserRequest(page, {
      method: "POST",
      path,
    });
    try {
      await page
        .getByTestId(draftCellTestId(synopsis))
        .fill("Accepted draft navigation");
      await navigationGate.waitForHit;
      await page.keyboard.press("Enter");
      navigationGate.release();
      const created = await waitForViewRowByCell(
        page,
        incidentId,
        timelineViewSchemaId,
        synopsis,
        "Accepted draft navigation",
      );
      await expect(editor(page, created.record_id)).toHaveCount(0);
      await expect(page.getByTestId(draftCellTestId(synopsis))).toBeFocused();
      expect(await fetchRecordHistoryCount(page, created.record_id)).toBe(1);
      expect(
        await queryViewRows(page, incidentId, timelineViewSchemaId),
      ).toHaveLength(4);
    } finally {
      await navigationGate.dispose();
    }
    await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
  } finally {
    await held.dispose();
  }
});

test("Timeline uncertain creation replays exactly through refresh without losing input", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("GRID-RECOVER"),
    "Timeline uncertain capture recovery",
  );
  await openTimeline(page, incidentId);
  const path = `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/rows`;
  const attempts: string[] = [];
  const changeSets: string[] = [];
  let createdRecordId = "";
  let releaseReplay!: () => void;
  const replayGate = new Promise<void>((resolve) => {
    releaseReplay = resolve;
  });
  let reportLoss!: () => void;
  const responseLost = new Promise<void>((resolve) => {
    reportLoss = resolve;
  });
  await page.route(`**${path}`, async (route) => {
    attempts.push(required(route.request().postData()));
    if (attempts.length > 1) await replayGate;
    const response = await route.fetch();
    expect(response.status()).toBe(attempts.length === 1 ? 201 : 200);
    const body = await response.json();
    changeSets.push(body.data.change_set_id);
    createdRecordId = body.data.row.record_id;
    if (attempts.length === 1) {
      await route.abort("connectionfailed");
      reportLoss();
    } else await route.fulfill({ response });
  });
  try {
    const draft = page.getByTestId(draftCellTestId(synopsis));
    await scrollGridTargetIntoView({
      page,
      surface: timelineViewSchemaId,
      targetTestId: draftCellTestId(synopsis),
    });
    await draft.focus();
    await page.keyboard.insertText("Uncertain fact");
    await responseLost;
    await expect(draft).toHaveValue("Uncertain fact");
    await expect(draft).toBeFocused();
    await page.keyboard.type(" still typing");
    const queryPath = `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/query`;
    const refreshed = page.waitForResponse(
      (response) => response.url().endsWith(queryPath) && response.ok(),
    );
    await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("refresh-uncertain"),
      [synopsis]: "Other analyst fact",
    });
    await refreshed;
    await expect(draft).toHaveValue("Uncertain fact still typing");
    await expect(draft).toBeFocused();
    releaseReplay();
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    expect(attempts.length).toBeGreaterThanOrEqual(2);
    expect(new Set(attempts).size).toBe(1);
    expect(new Set(changeSets).size).toBe(1);
    await expect(editor(page, createdRecordId)).toHaveValue(
      "Uncertain fact still typing",
    );
    await expect(editor(page, createdRecordId)).toBeFocused();
    await page.keyboard.press("Escape");
    await expect(cell(page, createdRecordId)).toBeFocused();
    await expect(
      page.getByTestId(rowCellTestId(createdRecordId, synopsis)),
    ).toHaveText("Uncertain fact still typing");
    const saved = await queryViewRows(page, incidentId, timelineViewSchemaId);
    expect(saved).toHaveLength(2);
    expect(
      saved.find((row) => row.record_id === createdRecordId)?.cells[synopsis]
        ?.value,
    ).toBe("Uncertain fact still typing");
    expect(await fetchRecordHistoryCount(page, createdRecordId)).toBe(2);
    await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
  } finally {
    releaseReplay();
    if (!page.isClosed()) await page.unroute(`**${path}`);
  }
});

test("Timeline paste retains committed creates through lost response navigation and failed refresh", async ({
  page,
}) => {
  const { incidentId, rows } = await seedTimeline(page, 2);
  const unrelated = required(rows[0]).record_id;
  const target = required(rows[1]).record_id;
  const path = `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/clipboard-paste`;
  const queryPath = `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/query`;
  const attempts: string[] = [];
  const changes: string[] = [];
  let failReads = false;
  let reportLoss!: () => void;
  const lost = new Promise<void>((resolve) => {
    reportLoss = resolve;
  });
  let releaseReplay!: () => void;
  const gate = new Promise<void>((resolve) => {
    releaseReplay = resolve;
  });
  await page.route(`**${queryPath}`, async (route) => {
    if (!failReads) {
      await route.continue();
      return;
    }
    await route.fulfill({
      status: 503,
      contentType: "application/json",
      body: JSON.stringify({
        error: { code: "internal_error", message: "Temporary read failure" },
      }),
    });
  });
  await page.route(`**${path}`, async (route) => {
    attempts.push(required(route.request().postData()));
    const response = await route.fetch();
    expect(response.status()).toBe(200);
    const result = await response.json();
    changes.push(result.data.change_set_id);
    if (attempts.length === 1) {
      await route.abort("connectionfailed");
      reportLoss();
    } else {
      await gate;
      failReads = true;
      await route.fulfill({ response });
    }
  });
  try {
    await selectCell(page, target);
    await clipboard(page, "Captured update\nCaptured new record");
    await lost;
    await page.getByRole("button", { name: "Hosts", exact: true }).click();
    await page.getByRole("button", { name: "Timeline", exact: true }).click();
    await expect(
      page.getByTestId(timelineMutationSubstrateReadyTestId()),
    ).toBeVisible();
    await externalPatch(page, incidentId, target, source, "Intervening source");
    const before = await fetchRecordHistoryCount(page, target);
    await openRecoveryItem(page, /^(Paste|Fill|Tag assignment) ·/);
    const retry = page.getByRole("button", {
      name: "Retry paste",
      exact: true,
    });
    await tabTo(page, retry);
    await page.keyboard.press("Enter");
    await expect.poll(() => attempts.length).toBe(2);
    await page
      .getByRole("button", { name: "Close recovery", exact: true })
      .click();
    await scrollGridCellIntoView({
      page,
      surface: timelineViewSchemaId,
      recordId: unrelated,
      cellKey: synopsis,
    });
    await page.getByTestId(rowCellTestId(unrelated, synopsis)).click();
    await editor(page, unrelated).fill("Newer typing survives");
    releaseReplay();
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expect(editor(page, unrelated)).toBeFocused();
    await expect(editor(page, unrelated)).toHaveValue("Newer typing survives");
    await openRecoveryItem(page, /^(Paste|Fill|Tag assignment) ·/);
    failReads = false;
    const refresh = page.getByRole("button", {
      name: "Retry refresh",
      exact: true,
    });
    await tabTo(page, refresh);
    await page.keyboard.press("Enter");
    await expect(refresh).toHaveCount(0);
    expect(attempts).toEqual([attempts[0], attempts[0]]);
    expect(changes).toEqual([changes[0], changes[0]]);
    const saved = await queryViewRows(page, incidentId, timelineViewSchemaId);
    expect(saved).toHaveLength(3);
    expect(
      saved.find((row) => row.record_id === target)?.cells[source]?.value,
    ).toBe("Intervening source");
    expect(await fetchRecordHistoryCount(page, target)).toBe(before);
    expect(attempts).toHaveLength(2);
  } finally {
    releaseReplay();
    if (!page.isClosed()) {
      await page.unroute(`**${path}`);
      await page.unroute(`**${queryPath}`);
    }
  }
});

test("Timeline paste keeps ordered grouped conflicts and per-cell attributed correction", async ({
  page,
}) => {
  const { incidentId, rows } = await seedTimeline(page, 2);
  const first = required(rows[0]).record_id;
  const second = required(rows[1]).record_id;
  const path = `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/clipboard-paste`;
  const held = await holdBrowserRequest(page, { method: "POST", path });
  try {
    await selectCell(page, first);
    await clipboard(page, "Client first\nClient second\nAccepted create");
    await held.waitForHit;
    await externalPatch(page, incidentId, first, synopsis, "Server first");
    await externalPatch(page, incidentId, second, synopsis, "Server second");
    held.release();
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Conflict");
    await page
      .getByRole("button", { name: "Open conflict recovery", exact: true })
      .click();
    await expect(
      page.getByRole("navigation", { name: "Workbook conflict navigator" }),
    ).toBeVisible();
    await expect(page.getByText("1 of 2", { exact: true })).toBeVisible();
    await expect(
      page.getByTestId(conflictMarkerTestId(second, synopsis)),
    ).toBeVisible();
    const saved = await queryViewRows(page, incidentId, timelineViewSchemaId);
    expect(saved).toHaveLength(3);
    expect(
      saved.find((row) => row.record_id === first)?.cells[synopsis]?.value,
    ).toBe("Server first");
    await openRecoveryItem(page, /^(Paste|Fill|Tag assignment) ·/);
    await page
      .getByRole("button", { name: "Review conflicts", exact: true })
      .click();
    const useMine = page.getByRole("button", {
      name: "Use my unsaved value",
      exact: true,
    });
    await tabTo(page, useMine);
    await page.keyboard.press("Enter");
    await waitForViewRowByCell(
      page,
      incidentId,
      timelineViewSchemaId,
      synopsis,
      "Client first",
    );
    await openRecoveryItem(page, /^Paste ·/);
    await page
      .getByRole("button", { name: "Review conflicts", exact: true })
      .click();
    await expect(
      page
        .getByRole("region", { name: "Your unsaved value", exact: true })
        .getByRole("code"),
    ).toBeVisible();
    await tabTo(page, useMine);
    await page.keyboard.press("Enter");
    await waitForViewRowByCell(
      page,
      incidentId,
      timelineViewSchemaId,
      synopsis,
      "Client second",
    );
    expect(await fetchRecordHistoryCount(page, first)).toBe(3);
    expect(await fetchRecordHistoryCount(page, second)).toBe(3);
    await expect(
      page.getByRole("button", { name: /for all conflicts/ }),
    ).toHaveCount(0);
  } finally {
    await held.dispose();
  }
});

test("Timeline exact headers and duplicate clipboard delivery preserve one semantic action", async ({
  page,
}) => {
  const { incidentId, rows } = await seedTimeline(page, 2);
  const first = required(rows[0]).record_id;
  const fields = requireViewContract(timelineViewSchemaId).fields.filter(
    (field) => !field.defaultHidden && field.gridEditable,
  );
  const text = [
    fields.map((field) => field.label).join("\t"),
    ...["Header first", "Header second"].map((value) =>
      fields
        .map((field) => (field.fieldKey === synopsis ? value : ""))
        .join("\t"),
    ),
  ].join("\n");
  const attempts: string[] = [];
  page.on("request", (request) => {
    if (request.url().endsWith("/clipboard-paste"))
      attempts.push(required(request.postData()));
  });
  await selectCell(page, first, source);
  await cell(page, first, source).evaluate((element, clipboardText) => {
    const data = new DataTransfer();
    data.setData("text/plain", clipboardText);
    const event = new ClipboardEvent("paste", {
      bubbles: true,
      cancelable: true,
      clipboardData: data,
    });
    element.dispatchEvent(event);
    element.dispatchEvent(event);
  }, text);
  await waitForViewRowByCell(
    page,
    incidentId,
    timelineViewSchemaId,
    synopsis,
    "Header second",
  );
  expect(attempts).toHaveLength(1);
  expect(JSON.parse(required(attempts[0])).targets).toHaveLength(2);
  expect(JSON.parse(required(attempts[0])).columns).toEqual(
    fields.map((field) => field.fieldKey),
  );
  await selectCell(page, first);
  await clipboard(page, "Deliberate repeat\nAnother deliberate row");
  await waitForViewRowByCell(
    page,
    incidentId,
    timelineViewSchemaId,
    synopsis,
    "Another deliberate row",
  );
  expect(attempts).toHaveLength(2);
  await selectCell(page, first);
  const repeatedResponse = page.waitForResponse(
    (response) =>
      response.url().endsWith("/clipboard-paste") &&
      response.request().method() === "POST",
  );
  await clipboard(page, "Deliberate repeat\nAnother deliberate row");
  expect((await repeatedResponse).ok()).toBe(true);
  expect(attempts).toHaveLength(3);
  expect(JSON.parse(required(attempts[1])).client_txn_id).not.toBe(
    JSON.parse(required(attempts[2])).client_txn_id,
  );
  expect(JSON.parse(required(attempts[0])).client_txn_id).not.toBe(
    JSON.parse(required(attempts[1])).client_txn_id,
  );
  expect(
    await queryViewRows(page, incidentId, timelineViewSchemaId),
  ).toHaveLength(2);
});

test("Timeline fill and tagging retain conflicts-only receipts and independent local recovery", async ({
  page,
}) => {
  const { incidentId, rows } = await seedTimeline(page, 3);
  const first = required(rows[0]).record_id;
  const fillTarget = required(rows[1]).record_id;
  const tagTarget = required(rows[2]).record_id;
  const path = `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/bulk-mutations`;
  let held = await holdBrowserRequest(page, { method: "POST", path });
  const responseForBatch = () =>
    page.waitForResponse(
      (response) =>
        response.url().endsWith(path) && response.request().method() === "POST",
    );
  try {
    await selectCell(page, first, source);
    await page.keyboard.press("Shift+ArrowDown");
    const fillResponse = responseForBatch();
    await page.keyboard.press("Control+d");
    await held.waitForHit;
    await externalPatch(
      page,
      incidentId,
      fillTarget,
      source,
      "Concurrent fill target",
    );
    held.release();
    const filled = await (await fillResponse).json();
    expect(filled.data.rows).toEqual([]);
    expect(filled.data.change_set_id).toBeUndefined();
    expect(filled.data.conflicts).toHaveLength(1);
    await held.dispose();
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Conflict");
    await page
      .getByRole("button", { name: "Open conflict recovery", exact: true })
      .click();
    await page
      .getByRole("button", { name: "Close recovery", exact: true })
      .click();
    held = await holdBrowserRequest(page, { method: "POST", path });
    await page
      .getByRole("checkbox", { name: `Select record ${tagTarget}` })
      .check();
    await page
      .getByRole("textbox", { name: "Tag for selected Timeline records" })
      .fill("client-tag");
    const tagResponse = responseForBatch();
    await page.getByRole("button", { name: "Assign tag", exact: true }).click();
    await held.waitForHit;
    const current = required(
      (await queryViewRows(page, incidentId, timelineViewSchemaId)).find(
        (row) => row.record_id === tagTarget,
      ),
    );
    const patched = await page.request.patch(
      `${apiBase}/api/v1/records/${tagTarget}`,
      {
        headers: await csrfHeaders(page),
        data: {
          view_schema_id: timelineViewSchemaId,
          client_txn_id: uniqueTxn("concurrent-tag"),
          base_row_version: current.row_version,
          changes: [
            {
              field_key: "timeline.tags",
              action_payload: {
                kind: "collection_actions_v1",
                actions: [{ op: "add_tag", tag_name: "server-tag" }],
              },
            },
          ],
        },
      },
    );
    expect(patched.ok()).toBe(true);
    held.release();
    const tagged = await (await tagResponse).json();
    expect(tagged.data.rows).toEqual([]);
    expect(tagged.data.change_set_id).toBeUndefined();
    expect(tagged.data.conflicts[0].conflict_resolution_class).toBe(
      "collection_review",
    );
    await openRecoveryItem(page, /^Tag assignment ·/);
    const tagSection = page.getByRole("region", { name: "Tag assignment 2" });
    await tagSection.getByRole("button", { name: "Review conflicts" }).click();
    const apply = page.getByRole("button", {
      name: "Apply reviewed collection",
      exact: true,
    });
    await tabTo(page, apply);
    await page.keyboard.press("Enter");
    await expect
      .poll(async () =>
        JSON.stringify(
          (await queryViewRows(page, incidentId, timelineViewSchemaId)).find(
            (row) => row.record_id === tagTarget,
          )?.cells["timeline.tags"]?.value,
        ),
      )
      .toContain("client-tag");
    const history = await fetchFullRecordHistory(page, tagTarget);
    expect(history.row_version).toBe(3);
    expect(new Set(history.items.map((item) => item.change_set_id)).size).toBe(
      3,
    );
    expect(history.items.every((item) => item.actor_user_id.length > 0)).toBe(
      true,
    );
    expect(
      (await queryViewRows(page, incidentId, timelineViewSchemaId)).find(
        (row) => row.record_id === fillTarget,
      )?.cells[source]?.value,
    ).toBe("Concurrent fill target");
  } finally {
    await held.dispose();
  }
});

test("Timeline clear retries exact lost receipts and recovers acknowledged reads without another write", async ({
  page,
}) => {
  const { incidentId, rows } = await seedTimeline(page, 2),
    target = required(rows[0]).record_id;
  const path = `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/bulk-mutations`;
  const queryPath = `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/query`;
  const attempts: string[] = [],
    changes: string[] = [];
  let failReads = false;
  let reportLoss = () => {};
  const lost = new Promise<void>((resolve) => {
    reportLoss = resolve;
  });
  await page.route(`**${queryPath}`, async (route) => {
    if (!failReads) {
      await route.continue();
      return;
    }
    await route.fulfill({
      status: 503,
      contentType: "application/json",
      body: JSON.stringify({
        error: { code: "internal_error", message: "Temporary read failure" },
      }),
    });
  });
  await page.route(`**${path}`, async (route) => {
    attempts.push(required(route.request().postData()));
    const response = await route.fetch();
    expect(response.status()).toBe(200);
    const result = await response.json();
    changes.push(result.data.change_set_id);
    if (attempts.length === 1) {
      await route.abort("connectionfailed");
      reportLoss();
    } else {
      failReads = true;
      await route.fulfill({ response });
    }
  });
  try {
    await selectCell(page, target);
    await page.keyboard.press("Delete");
    await lost;
    await page.getByRole("button", { name: "Hosts", exact: true }).click();
    await page.getByRole("button", { name: "Timeline", exact: true }).click();
    await expect(
      page.getByTestId(timelineMutationSubstrateReadyTestId()),
    ).toBeVisible();
    const revisions = await fetchRecordHistoryCount(page, target);
    expect(revisions).toBe(2);
    await openRecoveryItem(page, /^Clear contents ·/);
    const retry = page.getByRole("button", {
      name: "Retry clear contents",
      exact: true,
    });
    await tabTo(page, retry);
    await page.keyboard.press("Enter");
    await expect.poll(() => attempts.length).toBe(2);
    const refresh = page.getByRole("button", {
      name: "Retry refresh",
      exact: true,
    });
    await expect(refresh).toBeVisible();
    failReads = false;
    await tabTo(page, refresh);
    await page.keyboard.press("Enter");
    await expect(refresh).toHaveCount(0);
    expect(attempts).toEqual([attempts[0], attempts[0]]);
    expect(changes).toEqual([changes[0], changes[0]]);
    expect(await fetchRecordHistoryCount(page, target)).toBe(revisions);
    expect(
      (await queryViewRows(page, incidentId, timelineViewSchemaId)).find(
        (row) => row.record_id === target,
      )?.cells[synopsis]?.value,
    ).toBeNull();
  } finally {
    failReads = false;
    if (!page.isClosed()) {
      await page.unroute(`**${path}`);
      await page.unroute(`**${queryPath}`);
    }
  }
});

test("Timeline clear retains partial null conflicts and resolves the captured value", async ({
  page,
}) => {
  const { incidentId, rows } = await seedTimeline(page, 2),
    first = required(rows[0]).record_id,
    second = required(rows[1]).record_id;
  const path = `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/bulk-mutations`;
  const held = await holdBrowserRequest(page, { method: "POST", path });
  try {
    await selectCell(page, first);
    await page.keyboard.press("Shift+ArrowDown");
    const response = page.waitForResponse(
      (response) =>
        response.url().endsWith(path) && response.request().method() === "POST",
    );
    await page.keyboard.press("Delete");
    await held.waitForHit;
    await externalPatch(page, incidentId, first, synopsis, "Concurrent source");
    held.release();
    const result = await (await response).json();
    expect(result.data.rows).toHaveLength(1);
    expect(result.data.rows[0].record_id).toBe(second);
    expect(result.data.conflicts).toHaveLength(1);
    expect(result.data.conflicts[0].client_value).toBeNull();
    await openRecoveryItem(page, /^Clear contents ·/);
    await expect(
      page.getByText(/1 original conflict; 1 unresolved/),
    ).toBeVisible();
    await page
      .getByRole("button", { name: "Review conflicts", exact: true })
      .click();
    await expect(
      page.getByRole("region", { name: "Your unsaved value", exact: true }),
    ).toContainText("Cleared (null)");
    const resolved = page.waitForResponse(
      (response) =>
        response.url().includes("/resolve") &&
        response.request().method() === "POST",
    );
    await page
      .getByRole("button", { name: "Use my unsaved value", exact: true })
      .click();
    const request = (await resolved).request().postDataJSON();
    expect(request.resolved_value).toBeNull();
    await expect
      .poll(async () =>
        (await queryViewRows(page, incidentId, timelineViewSchemaId)).map(
          (row) => row.cells[synopsis]?.value,
        ),
      )
      .toEqual([null, null]);
    expect(await fetchRecordHistoryCount(page, first)).toBe(3);
    expect(await fetchRecordHistoryCount(page, second)).toBe(2);
  } finally {
    await held.dispose();
  }
});
