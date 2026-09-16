import type { CreateViewRowResponse } from "@cartulary/protocol-ts/http";
import { scrollGridTargetIntoView } from "@cartulary/test-utils/grid";
import {
  entityInspectButtonTestId,
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  gridShellTestId,
  surfaceTabTestId,
  workbookInspectorCloseButtonTestId,
  workbookInspectorFeatureActionTestId,
  workbookInspectorPanelTestId,
  workbookInspectorToggleTestId,
  workbookShellReadyTestId,
  workbookSurfacesMenuOptionTestId,
  workbookSurfacesMenuTriggerTestId,
} from "@cartulary/ui-contracts";
import {
  assessmentsViewSchemaId,
  decisionsViewSchemaId,
  evidenceViewSchemaId,
  hostsViewSchemaId,
  partiesViewSchemaId,
  requireViewContract,
  taskRequestsViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { createIncident } from "./support/incidents/fixtures";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { fetchFullRecordHistory } from "./support/workbook/history";
import { createViewRow, queryViewRows } from "./support/workbook/query";
import { openRecoveryItem, recoveryEntry } from "./support/workbook/recovery";
import { openGenericInspectorForRecord } from "./support/workbook/rowMutations";

test("Entity contextual Task creation retains editable references and replays a lost response exactly after navigation", async ({
  page,
  workerAdmin,
}, info) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("CTD-ENTITY"),
    "Contextual Task recovery",
  );
  const source = await createViewRow(page, incident, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("host"),
    "host.display_name": "Originating host",
  });
  const support = await createViewRow(page, incident, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn("evidence"),
    "evidence.title": "Reviewed Evidence",
  });
  const party = await createViewRow(page, incident, partiesViewSchemaId, {
    client_txn_id: uniqueTxn("party"),
    "party.display_name": "Requesting Party",
    "party.party_kind": "person",
  });
  const decision = await createViewRow(page, incident, decisionsViewSchemaId, {
    client_txn_id: uniqueTxn("decision"),
    "decision.decision_type": "containment",
    "decision.summary": "Authorizing decision",
    "decision.rationale": "Review",
  });
  const sourceBefore = await fetchFullRecordHistory(page, source.record_id);
  const supportBefore = await fetchFullRecordHistory(page, support.record_id);
  const requests: string[] = [],
    receipts: CreateViewRowResponse[] = [],
    patches: string[] = [];
  let releaseReplay: (() => void) | undefined;
  const replayGate = new Promise<void>((resolve) => {
    releaseReplay = resolve;
  });
  page.on("request", (request) => {
    if (request.method() === "PATCH") patches.push(request.url());
  });
  await page.route(
    `**/views/${taskRequestsViewSchemaId}/rows`,
    async (route) => {
      requests.push(route.request().postData() ?? "");
      const response = await route.fetch();
      expect(response.ok()).toBeTruthy();
      receipts.push(await response.json());
      if (requests.length === 1) await route.abort("failed");
      else {
        await replayGate;
        await route.fulfill({ response });
      }
    },
  );
  await openSource(page, incident, hostsViewSchemaId, source.record_id);
  await begin(page, hostsViewSchemaId, "create_related.task_request");
  await page
    .getByTestId(genericCreateSubmitTestId(taskRequestsViewSchemaId))
    .click();
  await expect(
    page.getByRole("alert").filter({ hasText: "Title is required" }),
  ).toBeVisible();
  expect(requests).toHaveLength(0);
  await fillTask(page, "Retained contextual task");
  const linkedLabel = fieldLabel(
    taskRequestsViewSchemaId,
    "task.linked_record_ids",
  );
  await page
    .getByRole("button", { name: new RegExp(`^Remove ${linkedLabel}`) })
    .click();
  await choose(
    page,
    taskRequestsViewSchemaId,
    "task.linked_record_ids",
    support.record_id,
    evidenceViewSchemaId,
  );
  await choose(
    page,
    taskRequestsViewSchemaId,
    "task.owner_user_id",
    workerAdmin.user_id,
  );
  await choose(
    page,
    taskRequestsViewSchemaId,
    "task.requester_party_id",
    party.record_id,
  );
  await choose(
    page,
    taskRequestsViewSchemaId,
    "task.decision_record_id",
    decision.record_id,
  );
  await info.attach("contextual-task-form", {
    body: await page.screenshot(),
    contentType: "image/png",
  });
  await page
    .getByTestId(genericCreateSubmitTestId(taskRequestsViewSchemaId))
    .click();
  await expect.poll(() => receipts.length).toBe(1);
  const committed = receipts[0];
  if (!committed) throw new Error("Missing create receipt");
  const committedHistory = await fetchFullRecordHistory(
    page,
    committed.data.row.record_id,
  );
  await switchSurface(page, evidenceViewSchemaId);
  const selectedSurface = page.getByTestId(
    surfaceTabTestId(evidenceViewSchemaId),
  );
  await selectedSurface.focus();
  await recovery(page);
  const retained = page.getByRole("region", {
    name: "Retained contextual creation",
    exact: true,
  });
  await expect(retained).toContainText("Creation outcome unconfirmed.");
  await retained
    .getByRole("button", { name: "Recover original creation", exact: true })
    .click();
  await expect.poll(() => receipts.length).toBe(2);
  await selectedSurface.focus();
  releaseReplay?.();
  expect(requests[1]).toBe(requests[0]);
  expect(receipts[1]?.data).toEqual(committed.data);
  await expect(retained).toContainText(committed.data.change_set_id);
  await expect(selectedSurface).toBeFocused();
  await retained.getByText("Creation receipt", { exact: true }).click();
  await expect(
    retained.getByText(
      new RegExp(`Change set: ${committed.data.change_set_id}`),
    ),
  ).toBeVisible();
  await expect(page.getByTestId(workbookShellReadyTestId())).toHaveAttribute(
    "data-active-view-schema-id",
    evidenceViewSchemaId,
  );
  const rows = await queryViewRows(page, incident, taskRequestsViewSchemaId);
  expect(rows).toHaveLength(1);
  expect(rows[0]?.cells["task.status"]?.value).toBe("open");
  expect(rows[0]?.cells["task.priority"]?.value).toBe("normal");
  expect(rows[0]?.cells["task.owner_user_id"]?.value).toBe(workerAdmin.user_id);
  expect(
    JSON.stringify(rows[0]?.cells["task.linked_record_ids"]?.value),
  ).toContain(support.record_id);
  expect(
    JSON.stringify(rows[0]?.cells["task.linked_record_ids"]?.value),
  ).not.toContain(source.record_id);
  expect(rows[0]?.cells["task.requester_party_id"]?.value).toBe(
    party.record_id,
  );
  expect(rows[0]?.cells["task.decision_record_id"]?.value).toBe(
    decision.record_id,
  );
  expect(
    await fetchFullRecordHistory(page, committed.data.row.record_id),
  ).toEqual(committedHistory);
  expect(await fetchFullRecordHistory(page, source.record_id)).toEqual(
    sourceBefore,
  );
  const supportAfter = await fetchFullRecordHistory(page, support.record_id);
  const additions = supportAfter.items.filter(
    (item) =>
      !supportBefore.items.some(
        (prior) => prior.history_item_ref === item.history_item_ref,
      ),
  );
  expect(new Set(additions.map((item) => item.change_set_id))).toEqual(
    new Set([committed.data.change_set_id]),
  );
  expect(patches).toEqual([]);
});

test("Assessment contextual Decision draft survives inspector closure and accepted refresh failure recovers with reads only", async ({
  page,
}, info) => {
  await page.setViewportSize({ width: 768, height: 800 });
  const incident = await createIncident(
    page,
    uniqueIncidentKey("CTD-ASSESSMENT"),
    "Contextual Decision recovery",
  );
  const subject = await createViewRow(page, incident, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("host"),
    "host.display_name": "Assessment subject",
  });
  const source = await createViewRow(page, incident, assessmentsViewSchemaId, {
    client_txn_id: uniqueTxn("assessment"),
    "assessment.subject_ref": subject.record_id,
    "assessment.subject_type": "host",
    "assessment.assessment_state": "unknown",
    "assessment.rationale": "Contextual source",
  });
  const before = await fetchFullRecordHistory(page, source.record_id);
  const requests: string[] = [],
    receipts: CreateViewRowResponse[] = [];
  let failedReads = false;
  await page.route(`**/views/${decisionsViewSchemaId}/query`, async (route) => {
    if (failedReads) await route.abort("failed");
    else await route.continue();
  });
  await page.route(`**/views/${decisionsViewSchemaId}/rows`, async (route) => {
    requests.push(route.request().postData() ?? "");
    const response = await route.fetch();
    expect(response.ok()).toBeTruthy();
    receipts.push(await response.json());
    failedReads = true;
    await route.fulfill({ response });
  });
  await openSource(page, incident, assessmentsViewSchemaId, source.record_id);
  await begin(page, assessmentsViewSchemaId, "create_related.decision");
  await fillDecision(page, "Retained decision");
  await page
    .getByTestId(workbookInspectorCloseButtonTestId(assessmentsViewSchemaId))
    .click();
  await switchSurface(page, hostsViewSchemaId);
  await recovery(page);
  const retained = page.getByRole("region", {
    name: "Retained contextual creation",
    exact: true,
  });
  await retained
    .getByRole("button", { name: "Resume contextual draft", exact: true })
    .click();
  await expect(
    page.getByTestId(genericCreateFieldTestId("decision.summary")),
  ).toHaveValue("Retained decision");
  await page
    .getByTestId(genericCreateSubmitTestId(decisionsViewSchemaId))
    .click();
  await expect.poll(() => receipts.length).toBe(1);
  const receipt = receipts[0];
  if (!receipt) throw new Error("Missing decision receipt");
  const retry = retained.getByRole("button", {
    name: "Retry creation refresh",
    exact: true,
  });
  await expect(retry).toBeEnabled();
  await expect(
    page.getByRole("status", { name: "Workbook save updates", exact: true }),
  ).toContainText("Saved");
  await info.attach("contextual-decision-refresh", {
    body: await page.screenshot(),
    contentType: "image/png",
  });
  failedReads = false;
  await retry.click();
  await expect(retry).toHaveCount(0);
  expect(requests).toHaveLength(1);
  await expect(page.getByTestId(workbookShellReadyTestId())).toHaveAttribute(
    "data-active-view-schema-id",
    hostsViewSchemaId,
  );
  const rows = await queryViewRows(page, incident, decisionsViewSchemaId);
  expect(rows).toHaveLength(1);
  expect(
    JSON.stringify(rows[0]?.cells["decision.support_refs"]?.value),
  ).toContain(source.record_id);
  expect(rows[0]?.cells["decision.status"]?.value).toBe("proposed");
  const after = await fetchFullRecordHistory(page, source.record_id);
  const changes = after.items.filter(
    (item) =>
      !before.items.some(
        (prior) => prior.history_item_ref === item.history_item_ref,
      ),
  );
  expect(new Set(changes.map((item) => item.change_set_id))).toEqual(
    new Set([receipt.data.change_set_id]),
  );
  const bounds = await retained.boundingBox();
  expect(bounds?.x).toBeGreaterThanOrEqual(0);
  expect((bounds?.x ?? 0) + (bounds?.width ?? 0)).toBeLessThanOrEqual(768);
  await retained.press("Escape");
  await expect(retained).not.toBeVisible();
  await expect(recoveryEntry(page)).toBeFocused();
});

test("Evidence contextual authoring requires explicit discard before replacing a retained draft", async ({
  page,
}) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("CTD-EVIDENCE"),
    "Contextual draft replacement",
  );
  const source = await createViewRow(page, incident, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn("evidence"),
    "evidence.title": "Evidence source",
  });
  await openSource(page, incident, evidenceViewSchemaId, source.record_id);
  await begin(page, evidenceViewSchemaId, "create_related.task_request");
  await fillTask(page, "Preserve this draft");
  await page
    .getByRole("button", { name: "Keep draft and close", exact: true })
    .click();
  await begin(page, evidenceViewSchemaId, "create_related.decision", false);
  await recovery(page);
  const retained = page.getByRole("region", {
    name: "Retained contextual creation",
    exact: true,
  });
  await retained
    .getByRole("button", { name: "Resume contextual draft", exact: true })
    .click();
  await expect(
    page.getByTestId(genericCreateFieldTestId("task.title")),
  ).toHaveValue("Preserve this draft");
  await retained
    .getByRole("button", { name: "Discard draft", exact: true })
    .click();
  await begin(page, evidenceViewSchemaId, "create_related.decision");
  await fillDecision(page, "Deliberate replacement");
  await page
    .getByTestId(genericCreateSubmitTestId(decisionsViewSchemaId))
    .click();
  await expect
    .poll(
      async () =>
        (await queryViewRows(page, incident, decisionsViewSchemaId)).length,
    )
    .toBe(1);
  expect(
    await queryViewRows(page, incident, taskRequestsViewSchemaId),
  ).toHaveLength(0);
  const rows = await queryViewRows(page, incident, decisionsViewSchemaId);
  expect(
    JSON.stringify(rows[0]?.cells["decision.support_refs"]?.value),
  ).toContain(source.record_id);
});

async function openSource(
  page: Page,
  incident: string,
  view: string,
  id: string,
) {
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${encodeURIComponent(view)}`,
  );
  await expect(page.getByTestId(gridShellTestId(view))).toBeVisible();
  if (view === hostsViewSchemaId) {
    const targetTestId = entityInspectButtonTestId("host", id);
    await scrollGridTargetIntoView({ page, surface: view, targetTestId });
    await page.getByTestId(targetTestId).click();
    await page.getByTestId(workbookInspectorToggleTestId(view)).click();
  } else await openGenericInspectorForRecord(page, view, id);
}
async function begin(
  page: Page,
  source: string,
  feature: string,
  expectForm = true,
) {
  const panel = page.getByTestId(
    workbookInspectorPanelTestId(source, "workflow"),
  );
  await expect(panel).toBeVisible();
  await page
    .getByTestId(workbookInspectorFeatureActionTestId(source, feature))
    .click();
  if (expectForm)
    await expect(
      page.getByTestId(
        genericCreateSubmitTestId(
          feature.endsWith("decision")
            ? decisionsViewSchemaId
            : taskRequestsViewSchemaId,
        ),
      ),
    ).toBeVisible();
}
async function fillTask(page: Page, title: string) {
  await page.getByTestId(genericCreateFieldTestId("task.title")).fill(title);
  await page
    .getByTestId(genericCreateFieldTestId("task.task_kind"))
    .selectOption("follow_up");
}
async function fillDecision(page: Page, summary: string) {
  await page
    .getByTestId(genericCreateFieldTestId("decision.summary"))
    .fill(summary);
  await page
    .getByTestId(genericCreateFieldTestId("decision.decision_type"))
    .selectOption("containment");
  await page
    .getByTestId(genericCreateFieldTestId("decision.rationale"))
    .fill("Reviewed contextual rationale");
}
function fieldLabel(view: string, field: string) {
  const result = requireViewContract(view).fieldMap[field];
  if (!result) throw new Error("Missing target field");
  return result.label;
}
async function choose(
  page: Page,
  target: string,
  field: string,
  id: string,
  surface?: string,
) {
  await page
    .getByRole("button", {
      name: `Choose ${fieldLabel(target, field)}`,
      exact: true,
    })
    .click();
  const picker = page.getByRole("region", {
    name: `Choose ${fieldLabel(target, field)}`,
    exact: true,
  });
  if (surface)
    await picker
      .getByRole("combobox", { name: "Reference surface", exact: true })
      .selectOption(surface);
  const select = picker.getByRole(
    requireViewContract(target).fieldMap[field]?.readKind === "collection"
      ? "listbox"
      : "combobox",
    { name: fieldLabel(target, field), exact: true },
  );
  await expect(select).toBeEnabled();
  await select.selectOption(id);
  await picker
    .getByRole("button", { name: "Cancel references", exact: true })
    .click();
  await page
    .getByRole("button", {
      name: `Choose ${fieldLabel(target, field)}`,
      exact: true,
    })
    .click();
  if (surface)
    await picker
      .getByRole("combobox", { name: "Reference surface", exact: true })
      .selectOption(surface);
  await expect(select).toBeEnabled();
  await select.selectOption(id);
  await picker
    .getByRole("button", { name: "Apply references", exact: true })
    .click();
}
async function recovery(page: Page) {
  await openRecoveryItem(page, /^(Task Requests|Decisions) (draft|creation) ·/);
}
async function switchSurface(page: Page, view: string) {
  const tab = page.getByTestId(surfaceTabTestId(view));
  if (await tab.isVisible()) await tab.click();
  else {
    await page.getByTestId(workbookSurfacesMenuTriggerTestId()).click();
    await page.getByTestId(workbookSurfacesMenuOptionTestId(view)).click();
  }
  await expect(page.getByTestId(gridShellTestId(view))).toBeVisible();
}
