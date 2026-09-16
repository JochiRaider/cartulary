import { scrollGridTargetIntoView } from "@cartulary/test-utils/grid";
import {
  entityInspectButtonTestId,
  gridRowTestId,
  gridShellTestId,
  rowCellTestId,
  rowHistoryDeleteButtonTestId,
  rowHistoryDestructiveConfirmButtonTestId,
  rowHistoryOpenButtonTestId,
  rowHistoryRestoreButtonTestId,
  saveStateTestId,
  workbookInspectorCloseButtonTestId,
  workbookInspectorToggleTestId,
} from "@cartulary/ui-contracts";
import {
  assessmentsViewSchemaId,
  evidenceViewSchemaId,
  hostsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { csrfHeaders } from "./support/auth/browserSession";
import { createIncident } from "./support/incidents/fixtures";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { publicHttpOperation } from "./support/transport/publicHttpOperationClient";
import { atJsonOrigin } from "./support/transport/publicJsonClient";
import { fetchFullRecordHistory } from "./support/workbook/history";
import { createViewRow, patchRecord } from "./support/workbook/query";
import { openRecoveryItem, recoveryEntry } from "./support/workbook/recovery";
import {
  clickTimelineRowAction,
  openGenericInspectorForRecord,
} from "./support/workbook/rowMutations";

type Surface = "Timeline" | "Generic" | "Entity" | "Assessment";
async function prepare(page: Page, surface: Surface) {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("HISTORY-RECOVERY"),
    `History recovery ${surface}`,
  );
  const host =
    surface === "Assessment" || surface === "Entity"
      ? await createViewRow(page, incidentId, hostsViewSchemaId, {
          client_txn_id: uniqueTxn("history-host"),
          "host.display_name": "History recovery host",
          "host.hostname": "history-recovery.example.test",
        })
      : null;
  const viewSchemaId =
    surface === "Timeline"
      ? timelineViewSchemaId
      : surface === "Generic"
        ? evidenceViewSchemaId
        : surface === "Entity"
          ? hostsViewSchemaId
          : assessmentsViewSchemaId;
  const row =
    surface === "Entity" && host
      ? host
      : await createViewRow(page, incidentId, viewSchemaId, {
          client_txn_id: uniqueTxn("history-row"),
          ...(surface === "Timeline"
            ? { "timeline.activity_synopsis_text": "History recovery row" }
            : surface === "Generic"
              ? {
                  "evidence.collector_party_text": "History recovery collector",
                  "evidence.title": "History recovery row",
                }
              : {
                  "assessment.subject_ref": host?.record_id,
                  "assessment.subject_type": "host",
                  "assessment.assessment_state": "confirmed",
                  "assessment.confidence_score": 85,
                  "assessment.rationale": "History recovery rationale",
                  "assessment.assessed_at": "2026-09-10T00:00:00Z",
                }),
        });
  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(viewSchemaId)}`,
  );
  await expect(page.getByTestId(gridShellTestId(viewSchemaId))).toBeVisible();
  if (surface === "Timeline")
    await clickTimelineRowAction(
      page,
      row.record_id,
      rowHistoryOpenButtonTestId(row.record_id),
    );
  else {
    if (surface === "Generic")
      await openGenericInspectorForRecord(page, viewSchemaId, row.record_id);
    else if (surface === "Entity") {
      const targetTestId = entityInspectButtonTestId("host", row.record_id);
      await scrollGridTargetIntoView({
        page,
        surface: viewSchemaId,
        targetTestId,
      });
      await page.getByTestId(targetTestId).click();
    } else {
      await page
        .getByTestId(
          rowCellTestId(row.record_id, "assessment.assessment_state"),
        )
        .click();
      await page
        .getByTestId(workbookInspectorToggleTestId(viewSchemaId))
        .click();
    }
    await page
      .getByRole("button", { name: "Open history", exact: true })
      .click();
  }
  await expect(page.getByTestId(rowHistoryDeleteButtonTestId())).toBeVisible();
  return { incidentId, row, viewSchemaId };
}

async function recoverDelete(
  page: Page,
  surface: Surface,
  mode:
    | "after_commit"
    | "before_commit"
    | "malformed"
    | "historical" = "after_commit",
) {
  const { row, viewSchemaId } = await prepare(page, surface);
  const before = await fetchFullRecordHistory(page, row.record_id);
  const bodies: string[] = [];
  let finishObserved = () => {};
  const observed = new Promise<void>((resolve) => {
    finishObserved = resolve;
  });
  await page.route(`**/api/v1/records/${row.record_id}`, async (route) => {
    if (route.request().method() !== "DELETE") {
      await route.continue();
      return;
    }
    bodies.push(route.request().postData() ?? "");
    if (bodies.length > 1) {
      await route.continue();
      return;
    }
    if (mode === "before_commit") {
      await route.abort("failed");
      finishObserved();
      return;
    }
    const response = await route.fetch();
    expect(response.ok()).toBe(true);
    if (mode === "malformed")
      await route.fulfill({ response, body: JSON.stringify({ data: {} }) });
    else await route.abort("failed");
    finishObserved();
  });
  await page.getByTestId(rowHistoryDeleteButtonTestId()).click();
  await page
    .getByTestId(
      rowHistoryDestructiveConfirmButtonTestId({ operation: "delete" }),
    )
    .click();
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Syncing");
  await expect(recoveryEntry(page)).toBeVisible();
  await observed;
  const close = page.getByTestId(
    workbookInspectorCloseButtonTestId(viewSchemaId),
  );
  if (await close.isVisible()) await close.click();
  const settledBeforeReplay = await fetchFullRecordHistory(page, row.record_id);
  expect(settledBeforeReplay.row_version).toBe(
    mode === "before_commit" ? before.row_version : before.row_version + 1,
  );
  if (mode === "historical") {
    const restore = await publicHttpOperation({
      operationID: "restoreRecord",
      pathParameters: { record_id: row.record_id },
      body: {
        base_row_version: settledBeforeReplay.row_version,
        client_txn_id: uniqueTxn("history-peer-restore"),
        reason: "Restore before historical replay",
      },
      headers: await csrfHeaders(page),
      request: atJsonOrigin(page.request, apiBase),
    });
    expect(restore.ok).toBe(true);
    if (!restore.ok) throw new Error("restore did not succeed");
    await patchRecord(page, row.record_id, {
      view_schema_id: viewSchemaId,
      base_row_version: restore.payload.data.row_version,
      client_txn_id: uniqueTxn("history-peer-edit"),
      changes: [
        {
          field_key: "timeline.activity_synopsis_text",
          value: "Newer accepted row",
        },
      ],
    });
    await expect(
      page.getByTestId(
        rowCellTestId(row.record_id, "timeline.activity_synopsis_text"),
      ),
    ).toHaveText("Newer accepted row");
  }
  const replayBaseline = await fetchFullRecordHistory(page, row.record_id);
  const trigger = recoveryEntry(page);
  await trigger.focus();
  await openRecoveryItem(
    page,
    /^(Soft-delete row|Restore[^·]*|Reverse[^·]*) ·/,
  );
  const recovery = page.getByRole("region", {
    name: "History action recovery",
    exact: true,
  });
  await expect(recovery).toContainText("Outcome unknown");
  await expect(
    page
      .getByRole("region", { name: "Recovery navigation", exact: true })
      .locator(":scope > h2"),
  ).toBeFocused();
  await test.info().attach("history-recovery", {
    body: await recovery.screenshot(),
    contentType: "image/png",
  });
  await recovery.getByRole("button", { name: "Replay exact action" }).click();
  await expect(recovery).toContainText("Action completed.");
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  expect(bodies).toHaveLength(2);
  expect(bodies[1]).toBe(bodies[0]);
  const recovered = await fetchFullRecordHistory(page, row.record_id);
  if (mode === "before_commit") {
    expect(recovered.row_version).toBe(before.row_version + 1);
    expect(
      new Set(
        recovered.items
          .filter(
            (item) =>
              !before.items.some(
                (old) => old.history_item_ref === item.history_item_ref,
              ),
          )
          .map((item) => item.change_set_id),
      ).size,
    ).toBe(1);
  } else {
    // Full current history equality proves replay added no mutation, change set or revision.
    expect(recovered).toEqual(replayBaseline);
  }
  if (mode === "historical") {
    await expect(
      page.getByTestId(
        rowCellTestId(row.record_id, "timeline.activity_synopsis_text"),
      ),
    ).toHaveText("Newer accepted row");
    return;
  }
  await recovery
    .getByRole("button", { name: "Review current history" })
    .click();
  await recovery.getByTestId(rowHistoryRestoreButtonTestId()).click();
  await recovery
    .getByTestId(
      rowHistoryDestructiveConfirmButtonTestId({ operation: "restore" }),
    )
    .click();
  await expect(
    page.getByTestId(gridRowTestId(viewSchemaId, row.record_id)),
  ).toBeVisible();
  const restored = await fetchFullRecordHistory(page, row.record_id);
  expect(restored.deleted).toBe(false);
  expect(restored.row_version).toBe(recovered.row_version + 1);
}

test("Timeline history recovers a committed delete after inspector closure", async ({
  page,
}) => recoverDelete(page, "Timeline"));
test("Generic history recovers a committed delete after inspector closure", async ({
  page,
}) => recoverDelete(page, "Generic"));
test("Entity history recovers a committed delete after inspector closure", async ({
  page,
}) => recoverDelete(page, "Entity"));
test("Assessment history recovers a committed delete after inspector closure", async ({
  page,
}) => recoverDelete(page, "Assessment"));
test("Timeline history recovers loss before commit without duplicate history", async ({
  page,
}) => recoverDelete(page, "Timeline", "before_commit"));
test("Timeline history recovers malformed success without duplicate history", async ({
  page,
}) => recoverDelete(page, "Timeline", "malformed"));
test("Timeline historical replay preserves a newer accepted row", async ({
  page,
}) => recoverDelete(page, "Timeline", "historical"));

test("Timeline acknowledged delete survives a failed history refresh", async ({
  page,
}) => {
  const { row, viewSchemaId } = await prepare(page, "Timeline");
  let acknowledged = false;
  let failRefresh = true;
  let mutations = 0;
  await page.route(`**/api/v1/records/${row.record_id}`, async (route) => {
    if (route.request().method() !== "DELETE") {
      await route.continue();
      return;
    }
    mutations++;
    const response = await route.fetch();
    expect(response.ok()).toBe(true);
    acknowledged = true;
    await route.fulfill({ response });
  });
  await page.route(
    `**/api/v1/records/${row.record_id}/history`,
    async (route) => {
      if (acknowledged && failRefresh) {
        await route.fulfill({
          status: 503,
          contentType: "application/json",
          body: JSON.stringify({ unavailable: true }),
        });
        return;
      }
      await route.continue();
    },
  );
  await page.getByTestId(rowHistoryDeleteButtonTestId()).click();
  await page
    .getByTestId(
      rowHistoryDestructiveConfirmButtonTestId({ operation: "delete" }),
    )
    .click();
  await openRecoveryItem(
    page,
    /^(Soft-delete row|Restore[^·]*|Reverse[^·]*) ·/,
  );
  const recovery = page.getByRole("region", {
    name: "History action recovery",
    exact: true,
  });
  await expect(recovery).toContainText("Action completed; refresh required.");
  const committed = await fetchFullRecordHistory(page, row.record_id);
  failRefresh = false;
  await recovery
    .getByRole("button", { name: "Refresh completed action" })
    .click();
  await expect(recovery).toContainText("Action completed.");
  await expect(
    page.getByTestId(gridRowTestId(viewSchemaId, row.record_id)),
  ).toHaveCount(0);
  expect(mutations).toBe(1);
  expect(await fetchFullRecordHistory(page, row.record_id)).toEqual(committed);
});
