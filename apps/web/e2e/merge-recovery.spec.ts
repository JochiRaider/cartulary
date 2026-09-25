import { Buffer } from "node:buffer";
import {
  type MergeEntityRecordResponse,
  validateHTTPOperationResponse,
} from "@cartulary/protocol-ts/http";
import { scrollGridTargetIntoView } from "@cartulary/test-utils/grid";
import {
  entityInspectButtonTestId,
  entityMergeControlTestId,
  gridRowTestId,
  gridShellTestId,
  rowHistoryActionTestId,
  rowHistoryRollbackConfirmButtonTestId,
  saveStateTestId,
  workbookInspectorCloseButtonTestId,
  workbookShellReadyTestId,
  workbookSurfacesMenuOptionTestId,
  workbookSurfacesMenuTriggerTestId,
} from "@cartulary/ui-contracts";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import { expect, type Page, type TestInfo } from "@playwright/test";
import { test } from "./fixtures";
import {
  aliasCollectionActionsPayload,
  collectionActionsPayload,
  collectionItems,
  findRow,
  requireItemByRawText,
} from "./support/entities/mentions";
import { createIncident } from "./support/incidents/fixtures";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import {
  fetchFullRecordHistory,
  openHistoryEventDetails,
} from "./support/workbook/history";
import {
  createViewRow,
  patchRecord,
  queryViewRows,
} from "./support/workbook/query";
import { openRecoveryItem, recoveryEntry } from "./support/workbook/recovery";

test("recovers host merge loss before dispatch with exact replay and existing history rollback", async ({
  page,
}, testInfo) => exerciseRecovery(page, testInfo, "host", "before dispatch"));
test("recovers host merge loss after commit with exact replay and existing history rollback", async ({
  page,
}, testInfo) => exerciseRecovery(page, testInfo, "host", "after commit"));
test("recovers identity merge loss before dispatch with exact replay and existing history rollback", async ({
  page,
}, testInfo) =>
  exerciseRecovery(page, testInfo, "identity", "before dispatch"));
test("recovers identity merge loss after commit with exact replay and existing history rollback", async ({
  page,
}, testInfo) => exerciseRecovery(page, testInfo, "identity", "after commit"));
async function exerciseRecovery(
  page: Page,
  testInfo: TestInfo,
  entityType: "host" | "identity",
  loss: "before dispatch" | "after commit",
) {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("MERGE-RECOVERY"),
    "Merge recovery service evidence",
  );
  const viewSchemaId =
    entityType === "host" ? hostsViewSchemaId : identitiesViewSchemaId;
  const identifierClass = entityType === "host" ? "hostname" : "email";
  const identifierField = `${entityType}.${identifierClass}`;
  const refs =
    entityType === "host" ? "timeline.host_refs" : "timeline.identity_refs";
  const label = "Duplicate label";
  const survivorValue =
    entityType === "host" ? "survivor.example.test" : "survivor@example.test";
  const loserValue =
    entityType === "host" ? "loser.example.test" : "loser@example.test";
  const survivor = await createViewRow(page, incidentId, viewSchemaId, {
    client_txn_id: uniqueTxn("survivor"),
    [`${entityType}.display_name`]: label,
    [identifierField]: survivorValue,
  });
  const loser = await createViewRow(page, incidentId, viewSchemaId, {
    client_txn_id: uniqueTxn("loser"),
    [`${entityType}.display_name`]: label,
    [identifierField]: loserValue,
    [`${entityType}.aliases`]: aliasCollectionActionsPayload([
      "Recovery alias",
    ]),
  });
  const dependent = await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("dependent"),
      "timeline.activity_synopsis_text": "Recovery dependent row",
      [refs]: collectionActionsPayload(["Recovery mention"]),
    },
  );
  const mention = requireItemByRawText(
    collectionItems(dependent, refs),
    "Recovery mention",
  );
  await patchRecord(page, dependent.record_id, {
    view_schema_id: timelineViewSchemaId,
    base_row_version: dependent.row_version,
    client_txn_id: uniqueTxn("resolve"),
    changes: [
      {
        field_key: refs,
        action_payload: {
          kind: "collection_actions_v1",
          actions: [
            {
              op: "resolve_item",
              item_ref: String(mention.item_ref),
              resolved_record_id: loser.record_id,
            },
          ],
        },
      },
    ],
  });
  const before = await fetchFullRecordHistory(page, survivor.record_id);
  const requests: string[] = [];
  const receipts: MergeEntityRecordResponse["data"][] = [];
  await page.route(
    `**/api/v1/records/${survivor.record_id}/merge`,
    async (route) => {
      const body = route.request().postData();
      if (!body) throw new Error("Merge request body missing");
      requests.push(body);
      if (requests.length === 1 && loss === "before dispatch") {
        await route.abort("failed");
        return;
      }
      const response = await route.fetch();
      expect(response.status()).toBe(200);
      const envelope: unknown = await response.json();
      expect(
        validateHTTPOperationResponse("mergeEntityRecord", envelope).ok,
      ).toBe(true);
      receipts.push((envelope as MergeEntityRecordResponse).data);
      if (requests.length === 1) await route.abort("failed");
      else await route.fulfill({ response });
    },
  );
  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(viewSchemaId)}`,
  );
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await scrollGridTargetIntoView({
    page,
    surface: viewSchemaId,
    targetTestId: entityInspectButtonTestId(entityType, survivor.record_id),
  });
  await page
    .getByTestId(entityInspectButtonTestId(entityType, survivor.record_id))
    .click();
  await page.getByTestId(entityMergeControlTestId("start")).click();
  const choices = page.getByTestId(entityMergeControlTestId("loser-record"));
  await expect(
    choices.locator(`option[value="${loser.record_id}"]`),
  ).toContainText(`${label} (${loser.record_id})`);
  await choices.selectOption(loser.record_id);
  await page
    .getByTestId(entityMergeControlTestId("reason"))
    .fill("Reviewed duplicate with retained exact recovery");
  const plan = page.getByTestId(entityMergeControlTestId("plan"));
  await expect(plan).toContainText(survivor.record_id);
  await expect(plan).toContainText(loser.record_id);
  await expect(plan).toContainText(`Carry as reusable ${loserValue}`);
  await expect(plan).toContainText("historical loser");
  await page.getByTestId(entityMergeControlTestId("review")).click();
  await expect(
    page.getByTestId(entityMergeControlTestId("cancel")),
  ).toBeFocused();
  await page.getByTestId(entityMergeControlTestId("confirm")).click();
  const recoveryTrigger = recoveryEntry(page);
  await expect(recoveryTrigger).toBeVisible();
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Syncing");
  expect(requests).toHaveLength(1);
  await page
    .getByTestId(workbookInspectorCloseButtonTestId(viewSchemaId))
    .click();
  const navigate = async (view: string) => {
    const menu = page.getByTestId(workbookSurfacesMenuTriggerTestId());
    if (await menu.isVisible()) {
      await menu.click();
      await page.getByTestId(workbookSurfacesMenuOptionTestId(view)).click();
    } else {
      const name =
        view === timelineViewSchemaId
          ? "Timeline"
          : entityType === "host"
            ? "Hosts"
            : "Identities";
      await page
        .getByRole("navigation", { name: "Built-in workbook surfaces" })
        .getByRole("button", { name, exact: true })
        .click();
    }
    await expect(page.getByTestId(gridShellTestId(view))).toBeVisible();
  };
  await navigate(timelineViewSchemaId);
  await navigate(viewSchemaId);
  await expect(
    page.getByTestId(gridRowTestId(viewSchemaId, survivor.record_id)),
  ).toBeVisible();
  await expect(
    page.getByTestId(gridRowTestId(viewSchemaId, loser.record_id)),
  ).toHaveCount(loss === "after commit" ? 0 : 1);
  const committedRows = await queryViewRows(page, incidentId, viewSchemaId);
  const committedTimeline = await queryViewRows(
    page,
    incidentId,
    timelineViewSchemaId,
  );
  const committedHistory = await fetchFullRecordHistory(
    page,
    survivor.record_id,
  );
  const committedLoserHistory = await fetchFullRecordHistory(
    page,
    loser.record_id,
  );
  let failRefresh = true;
  let failedRefreshReads = 0;
  await page.route(`**/views/${viewSchemaId}/query`, async (route) => {
    if (!failRefresh) return route.continue();
    failedRefreshReads += 1;
    await route.fulfill({
      status: 500,
      contentType: "application/json",
      body: JSON.stringify({
        error: {
          status: 500,
          code: "internal_error",
          message: "Deliberate projection refresh loss",
          request_id: "merge-refresh-loss",
          retryable: true,
        },
      }),
    });
  });
  await recoveryTrigger.focus();
  await openRecoveryItem(page, /^Entity merge ·/);
  const recovery = page.getByRole("region", {
    name: "Merge action recovery",
    exact: true,
  });
  await expect(
    page
      .getByRole("region", { name: "Recovery navigation", exact: true })
      .locator(":scope > h2"),
  ).toBeFocused();
  await recovery
    .getByRole("button", { name: "Replay exact merge request" })
    .click();
  await expect(recovery).toContainText("Refresh is still required");
  expect(failedRefreshReads).toBeGreaterThan(0);
  expect(requests).toHaveLength(2);
  expect(requests[1]).toBe(requests[0]);
  const receipt = receipts.at(-1);
  if (!receipt) throw new Error("Expected service merge receipt");
  if (loss === "after commit") {
    expect(receipts).toHaveLength(2);
    expect(receipts[1]).toEqual(receipts[0]);
  } else expect(receipts).toHaveLength(1);
  await expect(recovery).toContainText(receipt.change_set_id);
  await recovery.getByText("Merge receipt", { exact: true }).click();
  await expect(
    recovery.getByRole("table", { name: "Reusable identifiers by class" }),
  ).toBeVisible();
  await testInfo.attach("merge-acknowledgement-refresh-required", {
    body: await recovery.screenshot(),
    contentType: "image/png",
  });
  failRefresh = false;
  await recovery
    .getByRole("button", { name: "Refresh completed merge" })
    .click();
  await expect(recovery).toContainText("current projections refreshed");
  expect(requests).toHaveLength(2);
  const rows = await queryViewRows(page, incidentId, viewSchemaId);
  const timeline = await queryViewRows(page, incidentId, timelineViewSchemaId);
  const history = await fetchFullRecordHistory(page, survivor.record_id);
  const loserHistory = await fetchFullRecordHistory(page, loser.record_id);
  if (loss === "after commit") {
    expect(rows).toEqual(committedRows);
    expect(timeline).toEqual(committedTimeline);
    expect(history).toEqual(committedHistory);
    expect(loserHistory).toEqual(committedLoserHistory);
  }
  expect(rows.some((row) => row.record_id === loser.record_id)).toBe(false);
  const survivorAfter = findRow(rows, survivor.record_id);
  expect(survivorAfter.row_version).toBe(survivor.row_version + 1);
  expect(loserHistory.row_version).toBe(loser.row_version + 1);
  expect(survivorAfter.cells[identifierField]?.value).toBe(survivorValue);
  expect(
    collectionItems(survivorAfter, `${entityType}.reusable_identifiers`),
  ).toEqual(
    expect.arrayContaining([
      expect.objectContaining({
        identifier_class: identifierClass,
        normalized_value: loserValue,
      }),
    ]),
  );
  expect(collectionItems(survivorAfter, `${entityType}.aliases`)).toEqual(
    expect.arrayContaining([
      expect.objectContaining({ alias_text: "Recovery alias" }),
    ]),
  );
  expect(
    requireItemByRawText(
      collectionItems(findRow(timeline, dependent.record_id), refs),
      "Recovery mention",
    ).resolved_record_id,
  ).toBe(survivor.record_id);
  const oldRefs = new Set(before.items.map((item) => item.history_item_ref));
  const newItems = history.items.filter(
    (item) => !oldRefs.has(item.history_item_ref),
  );
  expect([...new Set(newItems.map((item) => item.change_set_id))]).toEqual([
    receipt.change_set_id,
  ]);
  const rollbackItem = newItems.find((item) =>
    item.available_rollback_actions.includes("change_set"),
  );
  if (!rollbackItem)
    throw new Error("Existing change-set rollback action missing");
  await recovery.getByRole("button", { name: "Review merge history" }).click();
  const anchor = {
    action: "change_set" as const,
    historyItemRef: rollbackItem.history_item_ref,
  };
  await openHistoryEventDetails(recovery, rollbackItem.history_item_ref);
  await recovery.getByTestId(rowHistoryActionTestId(anchor)).click();
  const rollbackResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response.url().endsWith(`/api/v1/records/${survivor.record_id}/rollback`),
  );
  await recovery
    .getByTestId(rowHistoryRollbackConfirmButtonTestId(anchor))
    .click();
  const rolledBack = await rollbackResponse;
  expect(rolledBack.status()).toBe(200);
  expect(JSON.parse(rolledBack.request().postData() ?? "{}").target).toEqual({
    kind: "change_set",
    change_set_id: receipt.change_set_id,
  });
  await expect(
    page.getByTestId(gridRowTestId(viewSchemaId, loser.record_id)),
  ).toBeVisible();
  const restoredRows = await queryViewRows(page, incidentId, viewSchemaId);
  const restoredSurvivor = findRow(restoredRows, survivor.record_id);
  expect(restoredSurvivor.row_version).toBeGreaterThan(
    receipt.survivor_row_version,
  );
  expect(restoredSurvivor.cells[identifierField]?.value).toBe(survivorValue);
  expect(
    collectionItems(restoredSurvivor, `${entityType}.reusable_identifiers`),
  ).toEqual([]);
  const restoredTimeline = await queryViewRows(
    page,
    incidentId,
    timelineViewSchemaId,
  );
  expect(
    requireItemByRawText(
      collectionItems(findRow(restoredTimeline, dependent.record_id), refs),
      "Recovery mention",
    ).resolved_record_id,
  ).toBe(loser.record_id);
  expect(requests).toHaveLength(2);
  await testInfo.attach("merge-exact-replay-service-evidence", {
    body: Buffer.from(
      JSON.stringify(
        {
          loss,
          entityType,
          requests,
          receipts,
          committedRows,
          rows,
          committedTimeline,
          timeline,
          newItems,
          restoredRows,
          restoredTimeline,
        },
        null,
        2,
      ),
    ),
    contentType: "application/json",
  });
  await page.keyboard.press("Escape");
  await expect(recoveryTrigger).toBeFocused();
}
