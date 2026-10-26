import {
  applyFilterChip,
  removeFilterChip,
  sortByHeader,
} from "@cartulary/test-utils/grid";
import {
  coordinationWorkflowTestId,
  decisionSupersessionTestId,
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  genericEditValueTestId,
  gridShellTestId,
  rowCellTestId,
  workbookFocusAnchorTestId,
  workbookInspectorCloseButtonTestId,
  workbookInspectorFeatureActionTestId,
  workbookInspectorToggleTestId,
} from "@cartulary/ui-contracts";
import {
  decisionsViewSchemaId,
  evidenceViewSchemaId,
  taskRequestsViewSchemaId,
} from "@cartulary/view-contracts";

import { expect, test } from "./fixtures";
import { csrfHeaders } from "./support/auth/browserSession";
import { collectionItems } from "./support/entities/mentions";
import { createIncident } from "./support/incidents/fixtures";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { publicHttpOperation } from "./support/transport/publicHttpOperationClient";
import { atJsonOrigin } from "./support/transport/publicJsonClient";
import { disableWorkbookSockets } from "./support/transport/socketIsolation";
import {
  activateSemanticGridCell,
  expectWorkbookSaved,
  openGenericSurface,
  setGenericCreateField,
  waitForReferenceCandidate,
} from "./support/workbook/authoring";
import {
  createViewRow,
  queryViewRows,
  waitForViewRowByCell,
} from "./support/workbook/query";
import { selectReferenceCandidates } from "./support/workbook/references";
import {
  editGenericCell,
  openGenericInspectorForRecord,
  submitGenericEditAndWait,
} from "./support/workbook/rowMutations";

test("Decision supersession and affected references stay native", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("TASK-DECISION-WORKFLOW"),
    "Workbook inspector task-decision-workflow Task and Decision workflows",
  );
  const support = await createViewRow(page, incidentId, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn("e906-support"),
    "evidence.title": "coordination-review supporting packet",
  });
  const targetDecision = await createViewRow(
    page,
    incidentId,
    decisionsViewSchemaId,
    {
      client_txn_id: uniqueTxn("e906-target-decision"),
      "decision.summary": "coordination-review target decision",
      "decision.decision_type": "containment",
      "decision.rationale": "Initial containment rationale.",
      "decision.support_refs": {
        kind: "collection_actions_v1",
        actions: [
          { op: "add_record_ref", linked_record_id: support.record_id },
        ],
      },
    },
  );
  const supersedingDecision = await createViewRow(
    page,
    incidentId,
    decisionsViewSchemaId,
    {
      client_txn_id: uniqueTxn("e906-superseding-decision"),
      "decision.summary": "coordination-review superseding decision",
      "decision.decision_type": "containment",
      "decision.rationale": "Updated containment rationale.",
      "decision.status": "approved",
      "decision.support_refs": {
        kind: "collection_actions_v1",
        actions: [
          { op: "add_record_ref", linked_record_id: support.record_id },
        ],
      },
    },
  );

  await disableWorkbookSockets(page);
  await openGenericSurface(page, incidentId, decisionsViewSchemaId);
  await expect(page).toHaveURL(
    new RegExp(`view_schema_id=${encodeURIComponent(decisionsViewSchemaId)}`),
  );
  await expect(
    page.getByTestId(gridShellTestId(decisionsViewSchemaId)),
  ).toBeVisible();
  await activateSemanticGridCell(
    page.getByTestId(
      rowCellTestId(targetDecision.record_id as string, "decision.summary"),
    ),
  );
  await expect(page.getByTestId(workbookFocusAnchorTestId())).toHaveText(
    `${decisionsViewSchemaId}:${targetDecision.record_id}:decision.summary`,
  );
  await page
    .getByTestId(
      workbookInspectorFeatureActionTestId(
        decisionsViewSchemaId,
        "decision.supersede",
      ),
    )
    .click();
  await expect(
    page
      .getByTestId(decisionSupersessionTestId("replacement"))
      .locator(`option[value="${supersedingDecision.record_id}"]`),
  ).toHaveCount(1);
  const supersedeResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response
        .url()
        .endsWith(`/api/v1/records/${targetDecision.record_id}/supersede`),
  );
  await page
    .getByTestId(decisionSupersessionTestId("replacement"))
    .selectOption(supersedingDecision.record_id as string);
  await page
    .getByTestId(decisionSupersessionTestId("reason"))
    .fill("coordination-review explicit supersession");
  await page.getByTestId(decisionSupersessionTestId("review-action")).click();
  await page.getByTestId(decisionSupersessionTestId("confirm")).click();
  const supersedeEnvelope = await (await supersedeResponse).json();
  expect(supersedeEnvelope.data.view_schema_id).toBe(decisionsViewSchemaId);
  expect(supersedeEnvelope.data.target_record_id).toBe(
    targetDecision.record_id,
  );
  expect(supersedeEnvelope.data.superseding_record_id).toBe(
    supersedingDecision.record_id,
  );
  await expectWorkbookSaved(page);
  let decisionRows = await queryViewRows(
    page,
    incidentId,
    decisionsViewSchemaId,
  );
  const refreshedTarget = decisionRows.find(
    (row) => row.record_id === targetDecision.record_id,
  );
  const refreshedSuperseding = decisionRows.find(
    (row) => row.record_id === supersedingDecision.record_id,
  );
  expect(refreshedTarget?.cells["decision.status"]?.value).toBe("superseded");
  expect(refreshedTarget?.cells["decision.is_superseded"]?.value).toBe(true);
  expect(
    refreshedSuperseding?.cells["decision.supersedes_record_id"]?.value,
  ).toBe(targetDecision.record_id);
  if (!refreshedSuperseding) {
    throw new Error("missing superseding decision row");
  }
  expect(
    collectionItems(refreshedSuperseding, "decision.support_refs"),
  ).toHaveLength(1);
  await openGenericInspectorForRecord(
    page,
    decisionsViewSchemaId,
    supersedingDecision.record_id as string,
  );
  await page
    .locator(`[data-inspector-edit-field="${"decision.affected_record_ids"}"]`)
    .click();
  await waitForReferenceCandidate(
    page,
    genericEditValueTestId(decisionsViewSchemaId),
    support.record_id as string,
    evidenceViewSchemaId,
  );
  await selectReferenceCandidates(
    page,
    genericEditValueTestId(decisionsViewSchemaId),
    support.record_id as string,
    evidenceViewSchemaId,
  );
  await submitGenericEditAndWait(
    page,
    decisionsViewSchemaId,
    supersedingDecision.record_id as string,
  );
  await expectWorkbookSaved(page);
  decisionRows = await queryViewRows(page, incidentId, decisionsViewSchemaId);
  const affectedDecision = decisionRows.find(
    (row) => row.record_id === supersedingDecision.record_id,
  );
  if (!affectedDecision) {
    throw new Error("missing affected decision row");
  }
  expect(
    collectionItems(affectedDecision, "decision.affected_record_ids"),
  ).toHaveLength(1);
  expect(affectedDecision.cells["decision.affected_record_count"]?.value).toBe(
    1,
  );
});

test("Task authoring priority queries and decision clearing stay native", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("TASK-DECISION-WORKFLOW"),
    "Workbook inspector task-decision-workflow Task and Decision workflows",
  );
  const support = await createViewRow(page, incidentId, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn("e906-support"),
    "evidence.title": "coordination-review supporting packet",
  });
  const targetDecision = await createViewRow(
    page,
    incidentId,
    decisionsViewSchemaId,
    {
      client_txn_id: uniqueTxn("e906-target-decision"),
      "decision.summary": "coordination-review target decision",
      "decision.decision_type": "containment",
      "decision.rationale": "Initial containment rationale.",
      "decision.support_refs": {
        kind: "collection_actions_v1",
        actions: [
          { op: "add_record_ref", linked_record_id: support.record_id },
        ],
      },
    },
  );
  const supersedingDecision = await createViewRow(
    page,
    incidentId,
    decisionsViewSchemaId,
    {
      client_txn_id: uniqueTxn("e906-superseding-decision"),
      "decision.summary": "coordination-review superseding decision",
      "decision.decision_type": "containment",
      "decision.rationale": "Updated containment rationale.",
      "decision.status": "approved",
      "decision.support_refs": {
        kind: "collection_actions_v1",
        actions: [
          { op: "add_record_ref", linked_record_id: support.record_id },
        ],
      },
    },
  );

  const prepared = await publicHttpOperation({
    operationID: "supersedeRecord",
    request: atJsonOrigin(page.request, apiBase),
    pathParameters: { record_id: targetDecision.record_id },
    headers: await csrfHeaders(page),
    body: {
      base_row_version: targetDecision.row_version,
      client_txn_id: uniqueTxn("task-decision-fixture"),
      replacement_record_id: supersedingDecision.record_id,
      reason: "Task authoring fixture",
    },
  });
  if (!prepared.ok)
    throw new Error("Could not prepare superseded Decision fixture");
  await disableWorkbookSockets(page);
  await openGenericSurface(page, incidentId, taskRequestsViewSchemaId);
  await expect(page).toHaveURL(
    new RegExp(
      `view_schema_id=${encodeURIComponent(taskRequestsViewSchemaId)}`,
    ),
  );
  await expect(
    page.getByTestId(gridShellTestId(taskRequestsViewSchemaId)),
  ).toBeVisible();
  const dueAt = "2026-05-19T16:00:00Z";
  await setGenericCreateField(
    page,
    "task.title",
    "coordination-review workbook task",
  );
  await setGenericCreateField(page, "task.task_kind", "collection");
  await setGenericCreateField(page, "task.workstream", "forensics");
  await setGenericCreateField(page, "task.due_at", dueAt);
  await setGenericCreateField(page, "task.external_ticket_ref", "SOC-E906");
  await waitForReferenceCandidate(
    page,
    genericCreateFieldTestId("task.decision_record_id"),
    supersedingDecision.record_id as string,
  );
  await setGenericCreateField(
    page,
    "task.decision_record_id",
    supersedingDecision.record_id as string,
  );
  await waitForReferenceCandidate(
    page,
    genericCreateFieldTestId("task.linked_record_ids"),
    support.record_id as string,
    evidenceViewSchemaId,
  );
  await setGenericCreateField(
    page,
    "task.linked_record_ids",
    support.record_id as string,
  );
  await page
    .getByTestId(genericCreateSubmitTestId(taskRequestsViewSchemaId))
    .click();
  const task = await waitForViewRowByCell(
    page,
    incidentId,
    taskRequestsViewSchemaId,
    "task.title",
    "coordination-review workbook task",
  );
  expect(task.cells["task.decision_record_id"]?.value).toBe(
    supersedingDecision.record_id,
  );
  expect(task.cells["task.external_ticket_ref"]?.value).toBe("SOC-E906");
  expect(Date.parse(task.cells["task.due_at"]?.value as string)).toBe(
    Date.parse(dueAt),
  );
  expect(task.cells["task.priority"]?.value).toBe("normal");
  expect(collectionItems(task, "task.linked_record_ids")).toHaveLength(1);

  await setGenericCreateField(
    page,
    "task.title",
    "coordination-review urgent queue task",
  );
  await setGenericCreateField(page, "task.task_kind", "follow_up");
  await setGenericCreateField(page, "task.priority", "urgent");
  await page
    .getByTestId(genericCreateSubmitTestId(taskRequestsViewSchemaId))
    .click();
  const urgentTask = await waitForViewRowByCell(
    page,
    incidentId,
    taskRequestsViewSchemaId,
    "task.title",
    "coordination-review urgent queue task",
  );
  expect(urgentTask.cells["task.priority"]?.value).toBe("urgent");

  const priorityRequestPromise = page.waitForRequest((request) => {
    if (
      request.method() !== "PATCH" ||
      !request.url().endsWith(`/api/v1/records/${task.record_id}`)
    ) {
      return false;
    }
    const body = request.postDataJSON() as {
      changes?: Array<{ field_key?: string; value?: unknown }>;
    };
    return (
      body.changes?.some(
        (change) =>
          change.field_key === "task.priority" && change.value === "high",
      ) ?? false
    );
  });
  await editGenericCell(
    page,
    taskRequestsViewSchemaId,
    task.record_id,
    "task.priority",
    "high",
  );
  const priorityRequest = await priorityRequestPromise;
  expect(priorityRequest.postDataJSON()).toMatchObject({
    view_schema_id: taskRequestsViewSchemaId,
    changes: [{ field_key: "task.priority", value: "high" }],
  });
  await expectWorkbookSaved(page);
  await expect(
    page.getByTestId(rowCellTestId(task.record_id, "task.priority")),
  ).toHaveText("high");

  await applyFilterChip(
    page,
    taskRequestsViewSchemaId,
    "task.priority",
    "high",
  );
  await expect(
    page.getByTestId(rowCellTestId(task.record_id, "task.priority")),
  ).toHaveText("high");
  await expect(
    page.getByTestId(rowCellTestId(urgentTask.record_id, "task.priority")),
  ).toHaveCount(0);
  await removeFilterChip(page, taskRequestsViewSchemaId, "task.priority");
  await expect(
    page.getByTestId(rowCellTestId(urgentTask.record_id, "task.priority")),
  ).toHaveText("urgent");

  await page
    .getByTestId(workbookInspectorCloseButtonTestId(taskRequestsViewSchemaId))
    .click();
  const priorityAscResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response
        .url()
        .endsWith(
          `/api/v1/incidents/${incidentId}/views/${taskRequestsViewSchemaId}/query`,
        ),
  );
  await sortByHeader(page, taskRequestsViewSchemaId, "task.priority");
  await priorityAscResponse;
  const priorityDescRequestPromise = page.waitForRequest((request) => {
    if (
      request.method() !== "POST" ||
      !request
        .url()
        .endsWith(
          `/api/v1/incidents/${incidentId}/views/${taskRequestsViewSchemaId}/query`,
        )
    ) {
      return false;
    }
    const body = request.postDataJSON() as {
      sort?: Array<{ direction?: string; field_key?: string }>;
    };
    return (
      body.sort?.some(
        (entry) =>
          entry.field_key === "task.priority" && entry.direction === "desc",
      ) ?? false
    );
  });
  await sortByHeader(page, taskRequestsViewSchemaId, "task.priority");
  const priorityDescRequest = await priorityDescRequestPromise;
  const priorityDescBody = priorityDescRequest.postDataJSON() as {
    sort?: Array<{ direction?: string; field_key?: string }>;
  };
  expect(priorityDescBody.sort).toContainEqual({
    direction: "desc",
    field_key: "task.priority",
  });
  await expect(
    page
      .getByTestId(gridShellTestId(taskRequestsViewSchemaId))
      .locator(
        '[role="row"][data-grid-record-id]:not([data-grid-record-id=""])',
      )
      .first(),
  ).toHaveAttribute("data-grid-record-id", urgentTask.record_id);

  const lifecycleResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "PATCH" &&
      response.url().endsWith(`/api/v1/records/${task.record_id}`),
  );
  await activateSemanticGridCell(
    page.getByTestId(rowCellTestId(task.record_id, "task.title")),
  );
  await page
    .getByTestId(workbookInspectorToggleTestId(taskRequestsViewSchemaId))
    .click();
  await page
    .getByTestId(coordinationWorkflowTestId("task-status"))
    .selectOption("blocked");
  await page
    .getByTestId(coordinationWorkflowTestId("task-blocked-reason"))
    .fill("Waiting on endpoint owner");
  await page
    .getByTestId(
      workbookInspectorFeatureActionTestId(
        taskRequestsViewSchemaId,
        "task.status.transition",
      ),
    )
    .click();
  await lifecycleResponse;
  await expectWorkbookSaved(page);
  let taskRows = await queryViewRows(
    page,
    incidentId,
    taskRequestsViewSchemaId,
  );
  let refreshedTask = taskRows.find((row) => row.record_id === task.record_id);
  expect(refreshedTask?.cells["task.status"]?.value).toBe("blocked");
  expect(refreshedTask?.cells["task.blocked_reason"]?.value).toBe(
    "Waiting on endpoint owner",
  );

  await applyFilterChip(
    page,
    taskRequestsViewSchemaId,
    "task.status",
    "blocked",
  );
  await expect(
    page.getByTestId(rowCellTestId(task.record_id, "task.status")),
  ).toHaveText("blocked");
  await applyFilterChip(
    page,
    taskRequestsViewSchemaId,
    "task.owner_user_id",
    String(refreshedTask?.cells["task.owner_user_id"]?.value),
  );
  await expect(
    page.getByTestId(rowCellTestId(task.record_id, "task.owner_user_id")),
  ).not.toHaveText("None");

  const clearRequestPromise = page.waitForRequest((request) => {
    if (
      request.method() !== "PATCH" ||
      !request.url().endsWith(`/api/v1/records/${task.record_id}`)
    ) {
      return false;
    }
    const body = request.postDataJSON() as {
      changes?: Array<{ field_key?: string; value?: unknown }>;
    };
    return (
      body.changes?.some(
        (change) => change.field_key === "task.decision_record_id",
      ) ?? false
    );
  });
  await openGenericInspectorForRecord(
    page,
    taskRequestsViewSchemaId,
    task.record_id,
  );
  await page
    .locator(`[data-inspector-edit-field="${"task.decision_record_id"}"]`)
    .click();
  await page
    .getByRole("button", { name: "Clear Decision", exact: true })
    .click();
  await submitGenericEditAndWait(
    page,
    taskRequestsViewSchemaId,
    task.record_id,
  );
  const clearRequest = await clearRequestPromise;
  expect(clearRequest.postDataJSON()).toMatchObject({
    view_schema_id: taskRequestsViewSchemaId,
    changes: [{ field_key: "task.decision_record_id", value: null }],
  });
  await expectWorkbookSaved(page);
  taskRows = await queryViewRows(page, incidentId, taskRequestsViewSchemaId);
  refreshedTask = taskRows.find((row) => row.record_id === task.record_id);
  expect(refreshedTask?.cells["task.decision_record_id"]?.value).toBeNull();
  expect(refreshedTask?.cells["task.blocked_reason"]?.value).toBe(
    "Waiting on endpoint owner",
  );

  const decisionRows = await queryViewRows(
    page,
    incidentId,
    decisionsViewSchemaId,
  );
  expect(
    decisionRows.find((row) => row.record_id === supersedingDecision.record_id)
      ?.cells["decision.supersedes_record_id"]?.value,
  ).toBe(targetDecision.record_id);
});
