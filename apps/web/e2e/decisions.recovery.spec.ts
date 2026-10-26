import { Buffer } from "node:buffer";

import { applyFilterChip } from "@cartulary/test-utils/grid";
import {
  decisionSupersessionTestId,
  rowCellTestId,
  workbookInspectorFeatureActionTestId,
  workbookInspectorPanelTestId,
} from "@cartulary/ui-contracts";
import { decisionsViewSchemaId } from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { expect, test } from "./fixtures";

import { createIncident } from "./support/incidents/fixtures";

import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";

import { installOwnedRoute } from "./support/transport/requestInterception";
import { disableWorkbookSockets } from "./support/transport/socketIsolation";
import {
  activateSemanticGridCell,
  openGenericSurface,
} from "./support/workbook/authoring";
import { fetchFullRecordHistory } from "./support/workbook/history";
import { createViewRow, queryViewRows } from "./support/workbook/query";
import { openRecoveryItem } from "./support/workbook/recovery";

async function exerciseDecisionRecovery(
  page: Page,
  status: "proposed" | "executed",
  loss: "before" | "after",
  attach: (data: unknown) => Promise<void>,
) {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("DECISION-RECOVERY"),
    "Decision supersession recovery",
  );
  const target = await createViewRow(page, incidentId, decisionsViewSchemaId, {
    client_txn_id: uniqueTxn("target"),
    "decision.summary": "Duplicate Decision label",
    "decision.decision_type": "containment",
    "decision.status": status,
    "decision.rationale": "Initial evidence",
  });
  const replacement = await createViewRow(
    page,
    incidentId,
    decisionsViewSchemaId,
    {
      client_txn_id: uniqueTxn("replacement"),
      "decision.summary": "Duplicate Decision label",
      "decision.decision_type": "containment",
      "decision.status": "approved",
      "decision.rationale": "Later evidence",
    },
  );
  const requests: string[] = [],
    receipts: unknown[] = [];
  await installOwnedRoute(
    page,
    `**/api/v1/records/${target.record_id}/supersede`,
    async (route) => {
      const body = route.request().postData();
      if (!body) throw new Error("Missing supersession body");
      requests.push(body);
      if (requests.length === 1 && loss === "before") {
        await route.abort("failed");
        return;
      }
      const response = await route.fetch();
      expect(response.status()).toBe(200);
      receipts.push((await response.json()).data);
      if (requests.length === 1) await route.abort("failed");
      else await route.fulfill({ response });
    },
  );
  await disableWorkbookSockets(page);
  await openGenericSurface(page, incidentId, decisionsViewSchemaId);
  await applyFilterChip(page, decisionsViewSchemaId, "decision.status", status);
  await expect(
    page.getByTestId(rowCellTestId(replacement.record_id, "decision.summary")),
  ).toHaveCount(0);
  await activateSemanticGridCell(
    page.getByTestId(rowCellTestId(target.record_id, "decision.summary")),
  );
  const historyPanel = page.getByTestId(
    workbookInspectorPanelTestId(decisionsViewSchemaId, "history"),
  );
  const start = page.getByTestId(
    workbookInspectorFeatureActionTestId(
      decisionsViewSchemaId,
      "decision.supersede",
    ),
  );
  await expect(start).toHaveCount(1);
  await expect(
    historyPanel.getByTestId(
      workbookInspectorFeatureActionTestId(
        decisionsViewSchemaId,
        "decision.supersede",
      ),
    ),
  ).toHaveCount(1);
  await start.click();
  const choices = page.getByTestId(decisionSupersessionTestId("replacement"));
  await expect(
    choices.locator(`option[value="${replacement.record_id}"]`),
  ).toContainText(`Duplicate Decision label (${replacement.record_id})`);
  await choices.selectOption(replacement.record_id);
  await page
    .getByTestId(decisionSupersessionTestId("reason"))
    .fill("  Later evidence\r\nReviewed reason  ");
  await page.getByTestId(decisionSupersessionTestId("review-action")).click();
  await expect(
    page.getByTestId(decisionSupersessionTestId("confirm")),
  ).toBeFocused();
  await expect(
    page.getByTestId(decisionSupersessionTestId("review")),
  ).toContainText(
    status === "executed" ? "remains executed" : "status changes to superseded",
  );
  await page.getByTestId(decisionSupersessionTestId("confirm")).click();
  await openRecoveryItem(page, /^Decision supersession ·/);
  await expect(
    page.getByRole("region", { name: "Decision action recovery", exact: true }),
  ).toContainText("The outcome is unknown. The server may have committed.");
  const beforeReplay = {
    rows: await queryViewRows(page, incidentId, decisionsViewSchemaId),
    targetHistory: await fetchFullRecordHistory(page, target.record_id),
    replacementHistory: await fetchFullRecordHistory(
      page,
      replacement.record_id,
    ),
  };
  let failRefresh = true;
  await installOwnedRoute(
    page,
    `**/views/${decisionsViewSchemaId}/query`,
    async (route) => {
      if (!failRefresh) return route.continue();
      failRefresh = false;
      await route.fulfill({
        status: 503,
        contentType: "application/json",
        body: JSON.stringify({
          error: {
            code: "service_unavailable",
            status: 503,
            retryable: true,
            message: "Refresh unavailable",
            request_id: "refresh-failure",
          },
        }),
      });
    },
  );
  await openRecoveryItem(page, /^Decision supersession ·/);
  const recovery = page.getByTestId(decisionSupersessionTestId("recovery"));
  await recovery
    .getByRole("button", { name: "Replay exact supersession request" })
    .click();
  await expect(
    recovery.getByText("Supersession accepted. Refresh is still required.", {
      exact: true,
    }),
  ).toBeVisible();
  expect(requests).toHaveLength(2);
  expect(requests[1]).toBe(requests[0]);
  if (loss === "after") expect(receipts[1]).toEqual(receipts[0]);
  await recovery.getByRole("button", { name: "Retry refresh" }).click();
  await expect(
    recovery.getByText(
      "Supersession accepted. Both Decisions and related projections refreshed.",
      { exact: true },
    ),
  ).toBeVisible();
  expect(requests).toHaveLength(2);
  const afterReplay = {
    rows: await queryViewRows(page, incidentId, decisionsViewSchemaId),
    targetHistory: await fetchFullRecordHistory(page, target.record_id),
    replacementHistory: await fetchFullRecordHistory(
      page,
      replacement.record_id,
    ),
  };
  const targetAfter = afterReplay.rows.find(
      (row) => row.record_id === target.record_id,
    ),
    replacementAfter = afterReplay.rows.find(
      (row) => row.record_id === replacement.record_id,
    );
  expect(targetAfter?.cells["decision.status"]?.value).toBe(
    status === "executed" ? "executed" : "superseded",
  );
  expect(targetAfter?.cells["decision.is_superseded"]?.value).toBe(true);
  expect(replacementAfter?.cells["decision.supersedes_record_id"]?.value).toBe(
    target.record_id,
  );
  expect(
    afterReplay.rows.filter(
      (row) =>
        row.cells["decision.supersedes_record_id"]?.value === target.record_id,
    ),
  ).toHaveLength(1);
  expect(targetAfter?.row_version).toBe(target.row_version + 1);
  expect(replacementAfter?.row_version).toBe(replacement.row_version + 1);
  if (loss === "after") expect(afterReplay).toEqual(beforeReplay);
  await attach({
    loss,
    status,
    target: target.record_id,
    replacement: replacement.record_id,
    requests,
    receipts,
    beforeReplay,
    afterReplay,
  });
}

test("Decision supersession recovers a precommit failure with exact identity in a filtered view", async ({
  page,
}, testInfo) => {
  await exerciseDecisionRecovery(page, "proposed", "before", async (data) => {
    await testInfo.attach("decision-recovery-evidence", {
      body: Buffer.from(JSON.stringify(data, null, 2)),
      contentType: "application/json",
    });
  });
});

test("Decision supersession recovers a lost executed receipt without advancing versions or change sets", async ({
  page,
}, testInfo) => {
  await exerciseDecisionRecovery(page, "executed", "after", async (data) => {
    await testInfo.attach("decision-recovery-evidence", {
      body: Buffer.from(JSON.stringify(data, null, 2)),
      contentType: "application/json",
    });
  });
});
