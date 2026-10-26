import { Buffer } from "node:buffer";

import { applyFilterChip, removeFilterChip } from "@cartulary/test-utils/grid";
import {
  rowCellTestId,
  workbookConflictResolverTestId,
  workbookInspectorCloseButtonTestId,
  workbookInspectorFeatureActionTestId,
} from "@cartulary/ui-contracts";
import { taskRequestsViewSchemaId } from "@cartulary/view-contracts";
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
  expectWorkbookSaved,
  openGenericSurface,
} from "./support/workbook/authoring";
import { fetchFullRecordHistory } from "./support/workbook/history";
import {
  createViewRow,
  patchRecord,
  queryViewRows,
} from "./support/workbook/query";
import { openRecoveryItem } from "./support/workbook/recovery";
import { selectReferenceCandidates } from "./support/workbook/references";
import { openGenericInspectorForRecord } from "./support/workbook/rowMutations";

async function exerciseTaskPatchRecovery(
  page: Page,
  loss: "before" | "after" | "malformed",
  attach: (data: unknown) => Promise<void>,
) {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("TASK-RECOVERY"),
    "Task patch recovery",
  );
  const target = await createViewRow(
    page,
    incidentId,
    taskRequestsViewSchemaId,
    {
      client_txn_id: uniqueTxn("task"),
      "task.title": "Recover this Task",
      "task.task_kind": "collection",
    },
  );
  await disableWorkbookSockets(page);
  const requests: string[] = [],
    receipts: unknown[] = [];
  let failRefresh = false;
  await installOwnedRoute(
    page,
    `**/api/v1/records/${target.record_id}`,
    async (route) => {
      if (route.request().method() !== "PATCH") return route.continue();
      const body = route.request().postData();
      if (!body) throw new Error("Missing patch body");
      requests.push(body);
      if (requests.length === 1 && loss === "before")
        return route.abort("failed");
      const response = await route.fetch();
      expect(response.status()).toBe(200);
      const envelope = await response.json();
      receipts.push(envelope.data);
      if (requests.length === 1) {
        if (loss === "malformed")
          return route.fulfill({
            response,
            json: {
              ...envelope,
              data: { ...envelope.data, change_set_id: "" },
            },
          });
        return route.abort("failed");
      }
      failRefresh = true;
      return route.fulfill({ response });
    },
  );
  await installOwnedRoute(
    page,
    `**/views/${taskRequestsViewSchemaId}/query`,
    async (route) => {
      if (!failRefresh) return route.continue();
      failRefresh = false;
      return route.fulfill({
        status: 503,
        contentType: "application/json",
        body: JSON.stringify({
          error: {
            code: "service_unavailable",
            status: 503,
            retryable: true,
            message: "Refresh unavailable",
            request_id: "task-refresh-failure",
          },
        }),
      });
    },
  );
  await openGenericSurface(page, incidentId, taskRequestsViewSchemaId);
  await applyFilterChip(page, taskRequestsViewSchemaId, "task.status", "open");
  await activateSemanticGridCell(
    page.getByTestId(rowCellTestId(target.record_id, "task.title")),
  );
  await page.getByLabel("Task lifecycle status").selectOption("done");
  await page
    .getByTestId(
      workbookInspectorFeatureActionTestId(
        taskRequestsViewSchemaId,
        "task.status.transition",
      ),
    )
    .click();
  await expect(
    page.getByRole("button", { name: "Retry original Task change" }),
  ).toBeVisible();
  await page
    .getByTestId(workbookInspectorCloseButtonTestId(taskRequestsViewSchemaId))
    .click();
  const beforeReplay = {
    rows: await queryViewRows(page, incidentId, taskRequestsViewSchemaId),
    history: await fetchFullRecordHistory(page, target.record_id),
  };
  await page
    .getByRole("button", { name: "Retry original Task change" })
    .click();
  await expect(
    page.getByRole("button", { name: "Refresh Task view" }),
  ).toBeVisible();
  expect(requests).toHaveLength(2);
  expect(requests[1]).toBe(requests[0]);
  if (loss !== "before") expect(receipts[1]).toEqual(receipts[0]);
  await page.getByRole("button", { name: "Refresh Task view" }).click();
  await expect(
    page.getByText(
      "This Task is no longer in the current view. Your filters are unchanged.",
    ),
  ).toBeVisible();
  expect(requests).toHaveLength(2);
  const afterReplay = {
    rows: await queryViewRows(page, incidentId, taskRequestsViewSchemaId),
    history: await fetchFullRecordHistory(page, target.record_id),
  };
  const saved = afterReplay.rows.find(
    (row) => row.record_id === target.record_id,
  );
  expect(saved?.row_version).toBe(target.row_version + 1);
  expect(saved?.cells["task.status"]?.value).toBe("done");
  if (loss !== "before") expect(afterReplay).toEqual(beforeReplay);
  else
    expect(afterReplay.history.items.length).toBe(
      beforeReplay.history.items.length + 1,
    );
  await attach({ loss, requests, receipts, beforeReplay, afterReplay });
}

test("Task lifecycle binds selected rows and atomically blocks completes and reopens", async ({
  page,
  testActor,
}, testInfo) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("TASK-LIFECYCLE"),
    "Task lifecycle editing",
  );
  const target = await createViewRow(
    page,
    incidentId,
    taskRequestsViewSchemaId,
    {
      client_txn_id: uniqueTxn("terminal-task"),
      "task.title": "Selected Task",
      "task.task_kind": "collection",
      "task.status": "canceled",
    },
  );
  const sibling = await createViewRow(
    page,
    incidentId,
    taskRequestsViewSchemaId,
    {
      client_txn_id: uniqueTxn("sibling"),
      "task.title": "Selected Task",
      "task.task_kind": "collection",
    },
  );
  await openGenericSurface(page, incidentId, taskRequestsViewSchemaId);
  await activateSemanticGridCell(
    page.getByTestId(rowCellTestId(target.record_id, "task.title")),
  );
  const submit = page.getByTestId(
    workbookInspectorFeatureActionTestId(
      taskRequestsViewSchemaId,
      "task.status.transition",
    ),
  );
  const status = page.getByLabel("Task lifecycle status");
  await expect(submit).toHaveCount(1);
  await expect(status).toHaveValue("canceled");
  await expect(status.locator('option[value="done"]')).toHaveJSProperty(
    "disabled",
    true,
  );
  await status.selectOption("blocked");
  await expect(submit).toBeDisabled();
  await page
    .getByLabel("Blocked reason", { exact: true })
    .fill("Waiting on evidence");
  await selectReferenceCandidates(
    page,
    "task-lifecycle-owner",
    testActor.user_id,
  );
  const ownerValue = page.getByLabel("Task lifecycle owner value", {
    exact: true,
  });
  const ownerPickerTrigger = page.getByRole("button", {
    name: "Choose task lifecycle owner",
    exact: true,
  });
  await ownerPickerTrigger.focus();
  await ownerPickerTrigger.press("Enter");
  const ownerPicker = page.getByRole("dialog", {
    name: "Choose task lifecycle owner",
    exact: true,
  });
  const ownerCandidates = ownerPicker.getByRole("listbox", {
    name: "Task lifecycle owner candidates",
  });
  const removeOwner = ownerPicker.getByRole("button", {
    name: /^Remove selected /,
  });
  await expect(ownerCandidates).toBeEnabled();
  await removeOwner.focus();
  await removeOwner.press("Space");
  await expect(removeOwner).toHaveCount(0);
  await expect(ownerCandidates).toBeFocused();
  await page.keyboard.press("Escape");
  await expect(ownerPicker).toHaveCount(0);
  await expect(ownerValue).toBeFocused();
  await expect(ownerValue).toHaveValue(testActor.user_id);
  await status.focus();
  await page.keyboard.press("Tab");
  await expect(
    page.getByLabel("Task lifecycle owner value", { exact: true }),
  ).toBeFocused();
  await page.keyboard.press("Tab");
  await expect(
    page.getByRole("button", {
      name: "Choose task lifecycle owner",
      exact: true,
    }),
  ).toBeFocused();
  await page.keyboard.press("Tab");
  await expect(
    page.getByLabel("Blocked reason", { exact: true }),
  ).toBeFocused();
  await page.keyboard.press("Tab");
  await expect(submit).toBeFocused();
  await expect(submit).toHaveCSS("outline-style", "solid");
  await testInfo.attach("task-lifecycle-editor-accessibility", {
    body: await page
      .getByRole("group", { name: "Task status transition", exact: true })
      .ariaSnapshot(),
    contentType: "text/plain",
  });
  await testInfo.attach("task-lifecycle-editor", {
    body: await page.screenshot(),
    contentType: "image/png",
  });
  const blockedResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "PATCH" &&
      response.url().endsWith(`/api/v1/records/${target.record_id}`),
  );
  await submit.focus();
  await submit.press("Enter");
  const blocked = await blockedResponse;
  expect(blocked.ok()).toBeTruthy();
  expect(blocked.request().postDataJSON().changes).toEqual(
    expect.arrayContaining([
      { field_key: "task.status", value: "blocked" },
      { field_key: "task.blocked_reason", value: "Waiting on evidence" },
      { field_key: "task.owner_user_id", value: testActor.user_id },
    ]),
  );
  await expectWorkbookSaved(page);
  await applyFilterChip(
    page,
    taskRequestsViewSchemaId,
    "task.status",
    "blocked",
  );
  await status.selectOption("done");
  const completedResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "PATCH" &&
      response.url().endsWith(`/api/v1/records/${target.record_id}`),
  );
  await submit.click();
  const completed = await completedResponse;
  const receipt = (await completed.json()).data;
  expect(completed.request().postDataJSON().changes).toEqual([
    { field_key: "task.status", value: "done" },
  ]);
  expect(receipt.row.cells["task.blocked_reason"].value).toBeNull();
  expect(
    Number.isFinite(Date.parse(receipt.row.cells["task.completed_at"].value)),
  ).toBe(true);
  await expect(
    page.getByText(
      "This Task is no longer in the current view. Your filters are unchanged.",
    ),
  ).toBeVisible();
  await expect(
    page.getByTestId(rowCellTestId(target.record_id, "task.title")),
  ).toHaveCount(0);
  await expect(
    page.getByRole("complementary").filter({ has: status }),
  ).toHaveAttribute("data-record-id", target.record_id);
  await expect(status).toHaveValue("done");
  await removeFilterChip(page, taskRequestsViewSchemaId, "task.status");
  await activateSemanticGridCell(
    page.getByTestId(rowCellTestId(target.record_id, "task.title")),
  );
  await expect(status).toHaveValue("done");
  await expect(page.getByLabel("Task completion time")).toHaveValue(
    receipt.row.cells["task.completed_at"].value,
  );
  await status.selectOption("open");
  await submit.click();
  await expectWorkbookSaved(page);
  const rows = await queryViewRows(page, incidentId, taskRequestsViewSchemaId);
  const reopened = rows.find((row) => row.record_id === target.record_id);
  expect(reopened?.cells["task.status"]?.value).toBe("open");
  expect(reopened?.cells["task.completed_at"]?.value).toBeNull();
  expect(
    rows.find((row) => row.record_id === sibling.record_id)?.row_version,
  ).toBe(sibling.row_version);
  await testInfo.attach("task-lifecycle-service-evidence", {
    body: Buffer.from(
      JSON.stringify(
        { blocked: (await blocked.json()).data, completed: receipt, reopened },
        null,
        2,
      ),
    ),
    contentType: "application/json",
  });
});

test("Task lifecycle replays a precommit loss with its exact request", async ({
  page,
}, testInfo) => {
  await exerciseTaskPatchRecovery(page, "before", async (data) =>
    testInfo.attach("task-recovery-evidence", {
      body: Buffer.from(JSON.stringify(data, null, 2)),
      contentType: "application/json",
    }),
  );
});

test("Task lifecycle replays a lost committed receipt without another version or change set", async ({
  page,
}, testInfo) => {
  await exerciseTaskPatchRecovery(page, "after", async (data) =>
    testInfo.attach("task-recovery-evidence", {
      body: Buffer.from(JSON.stringify(data, null, 2)),
      contentType: "application/json",
    }),
  );
});

test("Task lifecycle recovers a malformed success and retries failed refresh without patching", async ({
  page,
}, testInfo) => {
  await exerciseTaskPatchRecovery(page, "malformed", async (data) =>
    testInfo.attach("task-recovery-evidence", {
      body: Buffer.from(JSON.stringify(data, null, 2)),
      contentType: "application/json",
    }),
  );
});

test("Task lifecycle compound conflicts keep saved without losing the atomic draft", async ({
  page,
}, testInfo) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("TASK-CONFLICT"),
    "Task compound conflict",
  );
  const target = await createViewRow(
    page,
    incidentId,
    taskRequestsViewSchemaId,
    {
      client_txn_id: uniqueTxn("task"),
      "task.title": "Conflicted Task",
      "task.task_kind": "collection",
    },
  );
  await disableWorkbookSockets(page);
  await openGenericSurface(page, incidentId, taskRequestsViewSchemaId);
  await activateSemanticGridCell(
    page.getByTestId(rowCellTestId(target.record_id, "task.title")),
  );
  let first = true;
  await installOwnedRoute(
    page,
    `**/api/v1/records/${target.record_id}`,
    async (route) => {
      if (route.request().method() !== "PATCH" || !first)
        return route.continue();
      first = false;
      await patchRecord(page, target.record_id, {
        base_row_version: target.row_version,
        view_schema_id: taskRequestsViewSchemaId,
        client_txn_id: uniqueTxn("concurrent-task"),
        changes: [{ field_key: "task.status", value: "in_progress" }],
      });
      return route.fulfill({ response: await route.fetch() });
    },
  );
  await page.getByLabel("Task lifecycle status").selectOption("blocked");
  await page
    .getByLabel("Blocked reason", { exact: true })
    .fill("Retained atomic reason");
  await page
    .getByTestId(
      workbookInspectorFeatureActionTestId(
        taskRequestsViewSchemaId,
        "task.status.transition",
      ),
    )
    .click();
  await openRecoveryItem(page, /^Same-field conflict ·/);
  const resolver = page.getByTestId(workbookConflictResolverTestId());
  await expect(resolver).toBeVisible();
  await expect(
    resolver.getByRole("button", { name: "Use my unsaved value" }),
  ).toBeDisabled();
  const beforeKeep = await fetchFullRecordHistory(page, target.record_id);
  await resolver
    .getByRole("button", { name: "Keep saved", exact: true })
    .click();
  await expect(resolver).toHaveCount(0);
  const afterKeep = await fetchFullRecordHistory(page, target.record_id);
  expect(afterKeep).toEqual(beforeKeep);
  await openGenericInspectorForRecord(
    page,
    taskRequestsViewSchemaId,
    target.record_id,
  );
  await expect(page.getByLabel("Blocked reason", { exact: true })).toHaveValue(
    "Retained atomic reason",
  );
  await page
    .getByRole("button", { name: "Keep draft Status", exact: true })
    .click();
  await page
    .getByTestId(
      workbookInspectorFeatureActionTestId(
        taskRequestsViewSchemaId,
        "task.status.transition",
      ),
    )
    .click();
  await expectWorkbookSaved(page);
  const rows = await queryViewRows(page, incidentId, taskRequestsViewSchemaId);
  const saved = rows.find((row) => row.record_id === target.record_id);
  expect(saved?.cells["task.status"]?.value).toBe("blocked");
  expect(saved?.cells["task.blocked_reason"]?.value).toBe(
    "Retained atomic reason",
  );
  expect(saved?.row_version).toBe(target.row_version + 2);
  await testInfo.attach("task-compound-conflict-evidence", {
    body: Buffer.from(
      JSON.stringify({ beforeKeep, afterKeep, saved }, null, 2),
    ),
    contentType: "application/json",
  });
});
