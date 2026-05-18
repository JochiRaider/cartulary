import { applyFilterChip, removeFilterChip } from "@cartulary/test-utils";
import {
  conflictMarkerTestId,
  gridScrollportSelector,
  gridShellTestId,
  rowCellTestId,
} from "@cartulary/ui-contracts";
import type { Page, Request } from "@playwright/test";

import { expect, test } from "./fixtures";
import {
  createIncident,
  createViewRow,
  patchTimelineRecord,
  queryViewRows,
  uniqueIncidentKey,
  uniqueTxn,
} from "./helpers";
import {
  assessmentsViewSchemaId,
  commLogViewSchemaId,
  createAssessmentViaUI,
  evidenceViewSchemaId,
  editGenericCell,
  expectAssessmentGridOrder,
  hostsViewSchemaId,
  indicatorsViewSchemaId,
  notesViewSchemaId,
  partiesViewSchemaId,
  taskRequestsViewSchemaId,
  waitForViewRowByCell,
} from "./phase4Helpers";

const phase9Sprint0SentinelMessage =
  "Phase 9 Sprint 0 blocker sentinel: this is not behavior completion evidence; replace this sentinel with real Phase 9 implementation evidence before claiming the row complete.";

const timelineViewSchemaId = "cartulary.view.timeline.v1";

async function disableWorkbookSockets(page: Page) {
  await page.addInitScript(() => {
    class Phase9ClosedWebSocket {
      static readonly CONNECTING = 0;
      static readonly OPEN = 1;
      static readonly CLOSING = 2;
      static readonly CLOSED = 3;

      readonly url: string;
      readyState = Phase9ClosedWebSocket.CONNECTING;
      onclose: ((event: CloseEvent) => void) | null = null;
      onerror: ((event: Event) => void) | null = null;
      onmessage: ((event: MessageEvent) => void) | null = null;
      onopen: ((event: Event) => void) | null = null;

      constructor(url: string | URL) {
        this.url = String(url);
      }

      close() {
        this.readyState = Phase9ClosedWebSocket.CLOSED;
        this.onclose?.(new CloseEvent("close"));
      }

      send() {}
    }

    Object.defineProperty(window, "WebSocket", {
      configurable: true,
      value: Phase9ClosedWebSocket,
    });
  });
}

test("Phase 9 E-9-02 pastes a representative 20x5 Timeline clipboard range", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E902"),
    "Phase 9 E-9-02 clipboard paste",
  );
  const seed = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("e902-seed"),
    "timeline.summary": "Phase 9 paste seed",
  });

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByText("Timeline mutation substrate")).toBeVisible();

  const seedSummary = page.getByTestId(
    rowCellTestId(seed.record_id as string, "summary"),
  );
  await seedSummary.focus();
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `timeline:${seed.record_id}:timeline.summary`,
  );

  const pasteRows = Array.from({ length: 20 }, (_, index) => {
    const ordinal = index + 1;
    return [
      `Phase 9 paste summary ${ordinal}`,
      `phase9-host-${ordinal}.example.test`,
      `phase9-user-${ordinal}@example.test`,
      `readonly-evidence-${ordinal}`,
      `phase9-tag-${ordinal}`,
    ].join("\t");
  });
  const pastePayload = pasteRows.join("\n");

  const pasteResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response
        .url()
        .includes(
          `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/clipboard-paste`,
        ),
  );
  await seedSummary.evaluate((element, text) => {
    const data = new DataTransfer();
    data.setData("text/plain", text);
    element.dispatchEvent(
      new ClipboardEvent("paste", {
        bubbles: true,
        cancelable: true,
        clipboardData: data,
      }),
    );
  }, pastePayload);
  await expect((await pasteResponse).ok()).toBeTruthy();
  await expect(page.getByTestId("save-state")).toHaveText("Saved");
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `timeline:${seed.record_id}:timeline.summary`,
  );
  await expect(page.getByText(`Timeline row ${seed.record_id}`)).toBeVisible();
  await expect(
    page.getByTestId(rowCellTestId(seed.record_id as string, "summary")),
  ).toHaveValue("Phase 9 paste summary 1");
  await expect(page.getByTestId("timeline-grid-shell")).toBeVisible();

  const rows = await queryViewRows(page, incidentId, timelineViewSchemaId);
  const matchingRows = rows.filter((row) => {
    const cells = row.cells as Record<string, { value: unknown }>;
    return String(cells["timeline.summary"]?.value ?? "").startsWith(
      "Phase 9 paste summary ",
    );
  });
  expect(matchingRows).toHaveLength(20);
  const first = matchingRows.find((row) => row.record_id === seed.record_id);
  expect(first).toBeTruthy();
  expect(first?.cells["timeline.summary"]?.value).toBe(
    "Phase 9 paste summary 1",
  );
  expect(
    collectionDisplayTexts(first?.cells["timeline.host_refs"]?.value),
  ).toContain("phase9-host-1.example.test");
  expect(
    collectionDisplayTexts(first?.cells["timeline.identity_refs"]?.value),
  ).toContain("phase9-user-1@example.test");
  expect(first?.cells["timeline.evidence_count"]?.value).toBe(0);
  expect(
    collectionDisplayTexts(first?.cells["timeline.tags"]?.value),
  ).toContain("phase9-tag-1");
  const twentieth = matchingRows.find((row) => {
    const cells = row.cells as Record<string, { value: unknown }>;
    return cells["timeline.summary"]?.value === "Phase 9 paste summary 20";
  });
  expect(twentieth).toBeTruthy();
  expect(
    collectionDisplayTexts(twentieth?.cells["timeline.host_refs"]?.value),
  ).toContain("phase9-host-20.example.test");
  expect(
    collectionDisplayTexts(twentieth?.cells["timeline.identity_refs"]?.value),
  ).toContain("phase9-user-20@example.test");
  expect(twentieth?.cells["timeline.evidence_count"]?.value).toBe(0);
  expect(
    collectionDisplayTexts(twentieth?.cells["timeline.tags"]?.value),
  ).toContain("phase9-tag-20");
});

test("Phase 9 E-9-02 groups paste conflicts and preserves selection continuity", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E902-CONFLICT"),
    "Phase 9 E-9-02 grouped paste conflicts",
  );
  const first = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("e902-conflict-first"),
    "timeline.summary": "Phase 9 conflict first base",
  });
  const second = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("e902-conflict-second"),
    "timeline.summary": "Phase 9 conflict second base",
  });

  await disableWorkbookSockets(page);
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByText("Timeline mutation substrate")).toBeVisible();

  const firstSummary = page.getByTestId(
    rowCellTestId(first.record_id as string, "summary"),
  );
  await firstSummary.focus();
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `timeline:${first.record_id}:timeline.summary`,
  );

  await patchTimelineRecord(page, first.record_id as string, {
    view_schema_id: timelineViewSchemaId,
    base_row_version: first.row_version,
    client_txn_id: uniqueTxn("e902-conflict-first-server"),
    changes: [
      {
        field_key: "timeline.summary",
        value: "Phase 9 server first",
      },
    ],
  });
  await patchTimelineRecord(page, second.record_id as string, {
    view_schema_id: timelineViewSchemaId,
    base_row_version: second.row_version,
    client_txn_id: uniqueTxn("e902-conflict-second-server"),
    changes: [
      {
        field_key: "timeline.summary",
        value: "Phase 9 server second",
      },
    ],
  });

  const pasteResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response
        .url()
        .includes(
          `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/clipboard-paste`,
        ),
  );
  await firstSummary.evaluate((element) => {
    const data = new DataTransfer();
    data.setData(
      "text/plain",
      [
        "Phase 9 client first",
        "Phase 9 client second",
        "Phase 9 conflict create",
      ].join("\n"),
    );
    element.dispatchEvent(
      new ClipboardEvent("paste", {
        bubbles: true,
        cancelable: true,
        clipboardData: data,
      }),
    );
  });
  await expect((await pasteResponse).ok()).toBeTruthy();

  await expect(page.getByTestId("timeline-grid-shell")).toBeVisible();
  await expect(page.getByTestId("save-state")).toHaveText("Conflict");
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `timeline:${first.record_id}:timeline.summary`,
  );
  await expect(
    page.getByTestId(
      conflictMarkerTestId(first.record_id as string, "timeline.summary"),
    ),
  ).toBeVisible();
  await expect(
    page.getByTestId(
      conflictMarkerTestId(second.record_id as string, "timeline.summary"),
    ),
  ).toBeVisible();
  await expect(page.getByTestId("paste-conflict-navigator")).toBeVisible();
  await expect(page.getByTestId("paste-conflict-position")).toHaveText(
    "1 of 2",
  );
  await expect(page.getByTestId("conflict-local-value")).toHaveValue(
    "Phase 9 client first",
  );
  await page.getByTestId("paste-conflict-next").click();
  await expect(page.getByTestId("paste-conflict-position")).toHaveText(
    "2 of 2",
  );
  await expect(page.getByTestId("conflict-local-value")).toHaveValue(
    "Phase 9 client second",
  );
  await page.getByTestId("conflict-close").click();
  await expect(page.getByTestId("conflict-resolver")).toHaveCount(0);
  await expect(page.getByTestId("save-state")).toHaveText("Conflict");
});

test("Phase 9 E-9-03 Notes tab creates artifact-backed linked notes", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E903"),
    "Phase 9 E-9-03 Notes linked create",
  );
  const source = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("e903-source"),
    "timeline.summary": "Phase 9 linked note source",
  });

  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      notesViewSchemaId,
    )}`,
  );
  await expect(
    page.getByTestId(gridShellTestId(notesViewSchemaId)),
  ).toBeVisible();
  await expect(
    page
      .getByTestId("generic-create-note-source-record")
      .locator(`option[value="${source.record_id}"]`),
  ).toHaveCount(1);
  await page
    .getByTestId("generic-create-field-note.title")
    .fill("Phase 9 E-9-03 linked note");
  await page
    .getByTestId("generic-create-field-note.body")
    .fill("Created from the Notes tab with a source record link.");
  await page
    .getByTestId("generic-create-note-source-record")
    .selectOption(source.record_id as string);

  const responsePromise = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response
        .url()
        .endsWith(`/api/v1/records/${source.record_id}/linked-notes`),
  );
  await page.getByTestId(`generic-create-submit-${notesViewSchemaId}`).click();
  const response = await responsePromise;
  expect(response.ok()).toBeTruthy();
  const envelope = (await response.json()) as {
    data: { row: { record_id: string } };
  };
  const noteRecordId = envelope.data.row.record_id;
  await expect(
    page.getByTestId(rowCellTestId(noteRecordId, "note.title")),
  ).toHaveText("Phase 9 E-9-03 linked note");

  const rows = await queryViewRows(page, incidentId, notesViewSchemaId);
  const noteRow = rows.find((row) => row.record_id === noteRecordId);
  expect(noteRow).toBeTruthy();
  expect(noteRow?.cells["note.linked_record_count"]?.value).toBe(1);
});

test("Phase 9 E-9-04 Party create and link preserve raw text on the workbook surface", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E904"),
    "Phase 9 E-9-04 party create link",
  );
  const evidence = await createViewRow(page, incidentId, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn("e904-evidence"),
    "evidence.title": "Phase 9 party evidence",
    "evidence.collector_party_text": "Browser Collector Raw",
    "evidence.source_party_text": "Browser Source Raw",
  });
  const existingParty = await createViewRow(
    page,
    incidentId,
    partiesViewSchemaId,
    {
      client_txn_id: uniqueTxn("e904-existing-party"),
      "party.display_name": "Existing Source Party",
      "party.party_kind": "organization",
    },
  );

  await openGenericSurface(page, incidentId, evidenceViewSchemaId, "Evidence");
  await page
    .getByTestId(`generic-edit-record-${evidenceViewSchemaId}`)
    .selectOption(evidence.record_id as string);
  const typedCollectorText = "Browser Collector Raw <collector@example.test>";
  await editGenericCell(
    page,
    evidenceViewSchemaId,
    evidence.record_id as string,
    "evidence.collector_party_text",
    typedCollectorText,
  );
  await expect(page.getByTestId("generic-mutation-state")).toHaveText("Saved");
  let rows = await queryViewRows(page, incidentId, evidenceViewSchemaId);
  let refreshedEvidence = rows.find(
    (row) => row.record_id === evidence.record_id,
  );
  expect(refreshedEvidence?.cells["evidence.collector_party_text"]?.value).toBe(
    typedCollectorText,
  );
  expect(
    refreshedEvidence?.cells["evidence.collector_party_id"]?.value,
  ).toBeNull();
  await page
    .getByTestId(rowCellTestId(evidence.record_id as string, "evidence.title"))
    .focus();
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${evidenceViewSchemaId}:${evidence.record_id}:evidence.title`,
  );
  const preservedScroll = await setGenericGridScroll(
    page,
    evidenceViewSchemaId,
  );
  const assertEvidenceContextStable = async () => {
    await expect(page.getByRole("heading", { name: "Evidence" })).toBeVisible();
    await expect(page).toHaveURL(
      new RegExp(`view_schema_id=${encodeURIComponent(evidenceViewSchemaId)}`),
    );
    await expect(
      page.getByTestId(gridShellTestId(evidenceViewSchemaId)),
    ).toBeVisible();
    await expect(
      page.getByTestId(`generic-edit-record-${evidenceViewSchemaId}`),
    ).toHaveValue(evidence.record_id as string);
    await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
      `${evidenceViewSchemaId}:${evidence.record_id}:evidence.title`,
    );
    await expectGenericGridScroll(page, evidenceViewSchemaId, preservedScroll);
  };
  await page
    .getByTestId("party-link-pair")
    .selectOption("evidence.collector_party_text:evidence.collector_party_id");
  await page.getByTestId("party-link-create-from-text").click();
  const createdParty = await waitForViewRowByCell(
    page,
    incidentId,
    partiesViewSchemaId,
    "party.display_name",
    typedCollectorText,
  );
  await assertEvidenceContextStable();

  rows = await queryViewRows(page, incidentId, evidenceViewSchemaId);
  refreshedEvidence = rows.find(
    (row) => row.record_id === evidence.record_id,
  );
  expect(refreshedEvidence?.cells["evidence.collector_party_text"]?.value).toBe(
    typedCollectorText,
  );
  expect(refreshedEvidence?.cells["evidence.collector_party_id"]?.value).toBe(
    createdParty.record_id,
  );

  await page
    .getByTestId("party-link-pair")
    .selectOption("evidence.source_party_text:evidence.source_party_id");
  await page
    .getByTestId("party-link-existing-party")
    .selectOption(existingParty.record_id as string);
  await page.getByTestId("party-link-link-existing").click();
  await assertEvidenceContextStable();
  rows = await queryViewRows(page, incidentId, evidenceViewSchemaId);
  refreshedEvidence = rows.find((row) => row.record_id === evidence.record_id);
  expect(refreshedEvidence?.cells["evidence.source_party_text"]?.value).toBe(
    "Browser Source Raw",
  );
  expect(refreshedEvidence?.cells["evidence.source_party_id"]?.value).toBe(
    existingParty.record_id,
  );

  await page.getByTestId("party-link-clear-link").click();
  await assertEvidenceContextStable();
  rows = await queryViewRows(page, incidentId, evidenceViewSchemaId);
  refreshedEvidence = rows.find((row) => row.record_id === evidence.record_id);
  expect(refreshedEvidence?.cells["evidence.source_party_text"]?.value).toBe(
    "Browser Source Raw",
  );
  expect(
    refreshedEvidence?.cells["evidence.source_party_id"]?.value,
  ).toBeNull();

  await page
    .getByTestId("party-link-existing-party")
    .selectOption(existingParty.record_id as string);
  await page.getByTestId("party-link-link-existing").click();
  await assertEvidenceContextStable();
  await page.getByTestId("party-link-clear-text").click();
  await assertEvidenceContextStable();
  rows = await queryViewRows(page, incidentId, evidenceViewSchemaId);
  refreshedEvidence = rows.find((row) => row.record_id === evidence.record_id);
  expect(
    refreshedEvidence?.cells["evidence.source_party_text"]?.value,
  ).toBeNull();
  expect(refreshedEvidence?.cells["evidence.source_party_id"]?.value).toBe(
    existingParty.record_id,
  );

  await page
    .getByTestId("party-link-existing-party")
    .selectOption(existingParty.record_id as string);
  const clearBothPatchRequests: Request[] = [];
  const captureClearBothPatch = (request: Request) => {
    if (
      request.method() === "PATCH" &&
      request.url().endsWith(`/api/v1/records/${evidence.record_id}`)
    ) {
      clearBothPatchRequests.push(request);
    }
  };
  page.on("request", captureClearBothPatch);
  await page.getByTestId("party-link-clear-both").click();
  await assertEvidenceContextStable();
  page.off("request", captureClearBothPatch);
  expect(clearBothPatchRequests).toHaveLength(1);
  expect(clearBothPatchRequests[0]?.postDataJSON()).toMatchObject({
    view_schema_id: evidenceViewSchemaId,
    changes: [
      { field_key: "evidence.source_party_text", value: null },
      { field_key: "evidence.source_party_id", value: null },
    ],
  });
  rows = await queryViewRows(page, incidentId, evidenceViewSchemaId);
  refreshedEvidence = rows.find((row) => row.record_id === evidence.record_id);
  expect(
    refreshedEvidence?.cells["evidence.source_party_text"]?.value,
  ).toBeNull();
  expect(
    refreshedEvidence?.cells["evidence.source_party_id"]?.value,
  ).toBeNull();

  await expect(page.getByTestId("generic-mutation-state")).toHaveText("Saved");
  await assertEvidenceContextStable();

  const task = await createViewRow(page, incidentId, taskRequestsViewSchemaId, {
    client_txn_id: uniqueTxn("e904-task"),
    "task.title": "Phase 9 party task",
    "task.task_kind": "request",
    "task.requester_party_text": "Browser Requester Raw",
  });
  await openGenericSurface(
    page,
    incidentId,
    taskRequestsViewSchemaId,
    "Task Requests",
  );
  await page
    .getByTestId(`generic-edit-record-${taskRequestsViewSchemaId}`)
    .selectOption(task.record_id as string);
  const typedRequesterText = "Browser Requester Raw <requester@example.test>";
  await editGenericCell(
    page,
    taskRequestsViewSchemaId,
    task.record_id as string,
    "task.requester_party_text",
    typedRequesterText,
  );
  await expect(page.getByTestId("generic-mutation-state")).toHaveText("Saved");
  rows = await queryViewRows(page, incidentId, taskRequestsViewSchemaId);
  let refreshedTask = rows.find((row) => row.record_id === task.record_id);
  expect(refreshedTask?.cells["task.requester_party_text"]?.value).toBe(
    typedRequesterText,
  );
  expect(refreshedTask?.cells["task.requester_party_id"]?.value).toBeNull();
  await page
    .getByTestId(rowCellTestId(task.record_id as string, "task.title"))
    .focus();
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${taskRequestsViewSchemaId}:${task.record_id}:task.title`,
  );
  const taskScroll = await setGenericGridScroll(page, taskRequestsViewSchemaId);
  const assertTaskContextStable = async () => {
    await expect(
      page.getByRole("heading", { name: "Task Requests" }),
    ).toBeVisible();
    await expect(page).toHaveURL(
      new RegExp(
        `view_schema_id=${encodeURIComponent(taskRequestsViewSchemaId)}`,
      ),
    );
    await expect(
      page.getByTestId(gridShellTestId(taskRequestsViewSchemaId)),
    ).toBeVisible();
    await expect(
      page.getByTestId(`generic-edit-record-${taskRequestsViewSchemaId}`),
    ).toHaveValue(task.record_id as string);
    await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
      `${taskRequestsViewSchemaId}:${task.record_id}:task.title`,
    );
    await expectGenericGridScroll(page, taskRequestsViewSchemaId, taskScroll);
  };

  await page
    .getByTestId("party-link-pair")
    .selectOption("task.requester_party_text:task.requester_party_id");
  await page.getByTestId("party-link-create-from-text").click();
  const createdRequester = await waitForViewRowByCell(
    page,
    incidentId,
    partiesViewSchemaId,
    "party.display_name",
    typedRequesterText,
  );
  await assertTaskContextStable();
  rows = await queryViewRows(page, incidentId, taskRequestsViewSchemaId);
  refreshedTask = rows.find((row) => row.record_id === task.record_id);
  expect(refreshedTask?.cells["task.requester_party_text"]?.value).toBe(
    typedRequesterText,
  );
  expect(refreshedTask?.cells["task.requester_party_id"]?.value).toBe(
    createdRequester.record_id,
  );

  await page
    .getByTestId("party-link-existing-party")
    .selectOption(existingParty.record_id as string);
  await page.getByTestId("party-link-link-existing").click();
  await assertTaskContextStable();
  rows = await queryViewRows(page, incidentId, taskRequestsViewSchemaId);
  refreshedTask = rows.find((row) => row.record_id === task.record_id);
  expect(refreshedTask?.cells["task.requester_party_text"]?.value).toBe(
    typedRequesterText,
  );
  expect(refreshedTask?.cells["task.requester_party_id"]?.value).toBe(
    existingParty.record_id,
  );

  await page.getByTestId("party-link-clear-link").click();
  await assertTaskContextStable();
  rows = await queryViewRows(page, incidentId, taskRequestsViewSchemaId);
  refreshedTask = rows.find((row) => row.record_id === task.record_id);
  expect(refreshedTask?.cells["task.requester_party_text"]?.value).toBe(
    typedRequesterText,
  );
  expect(refreshedTask?.cells["task.requester_party_id"]?.value).toBeNull();

  await page
    .getByTestId("party-link-existing-party")
    .selectOption(existingParty.record_id as string);
  await page.getByTestId("party-link-link-existing").click();
  await assertTaskContextStable();
  await page.getByTestId("party-link-clear-text").click();
  await assertTaskContextStable();
  rows = await queryViewRows(page, incidentId, taskRequestsViewSchemaId);
  refreshedTask = rows.find((row) => row.record_id === task.record_id);
  expect(refreshedTask?.cells["task.requester_party_text"]?.value).toBeNull();
  expect(refreshedTask?.cells["task.requester_party_id"]?.value).toBe(
    existingParty.record_id,
  );

  await page
    .getByTestId("party-link-existing-party")
    .selectOption(existingParty.record_id as string);
  const taskClearBothPatchRequests: Request[] = [];
  const captureTaskClearBothPatch = (request: Request) => {
    if (
      request.method() === "PATCH" &&
      request.url().endsWith(`/api/v1/records/${task.record_id}`)
    ) {
      taskClearBothPatchRequests.push(request);
    }
  };
  page.on("request", captureTaskClearBothPatch);
  await page.getByTestId("party-link-clear-both").click();
  await assertTaskContextStable();
  page.off("request", captureTaskClearBothPatch);
  expect(taskClearBothPatchRequests).toHaveLength(1);
  expect(taskClearBothPatchRequests[0]?.postDataJSON()).toMatchObject({
    view_schema_id: taskRequestsViewSchemaId,
    changes: [
      { field_key: "task.requester_party_text", value: null },
      { field_key: "task.requester_party_id", value: null },
    ],
  });
  rows = await queryViewRows(page, incidentId, taskRequestsViewSchemaId);
  refreshedTask = rows.find((row) => row.record_id === task.record_id);
  expect(refreshedTask?.cells["task.requester_party_text"]?.value).toBeNull();
  expect(refreshedTask?.cells["task.requester_party_id"]?.value).toBeNull();

  const commLog = await createViewRow(page, incidentId, commLogViewSchemaId, {
    client_txn_id: uniqueTxn("e904-comm"),
    "comm_log.comm_type": "briefing",
    "comm_log.audience": "Browser Leadership Audience",
    "comm_log.channel_or_meeting": "Bridge",
    "comm_log.summary": "Phase 9 party comm log",
  });
  await openGenericSurface(
    page,
    incidentId,
    commLogViewSchemaId,
    "Communications Log",
  );
  await page
    .getByTestId(`generic-edit-record-${commLogViewSchemaId}`)
    .selectOption(commLog.record_id as string);
  await page
    .getByTestId(rowCellTestId(commLog.record_id as string, "comm_log.summary"))
    .focus();
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${commLogViewSchemaId}:${commLog.record_id}:comm_log.summary`,
  );
  const commScroll = await setGenericGridScroll(page, commLogViewSchemaId);
  const assertCommContextStable = async () => {
    await expect(
      page.getByRole("heading", { name: "Communications Log" }),
    ).toBeVisible();
    await expect(page).toHaveURL(
      new RegExp(`view_schema_id=${encodeURIComponent(commLogViewSchemaId)}`),
    );
    await expect(
      page.getByTestId(gridShellTestId(commLogViewSchemaId)),
    ).toBeVisible();
    await expect(
      page.getByTestId(`generic-edit-record-${commLogViewSchemaId}`),
    ).toHaveValue(commLog.record_id as string);
    await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
      `${commLogViewSchemaId}:${commLog.record_id}:comm_log.summary`,
    );
    await expectGenericGridScroll(page, commLogViewSchemaId, commScroll);
  };
  const addAndRemoveCommPartyRef = async (fieldKey: string) => {
    await page
      .getByTestId(`generic-edit-field-${commLogViewSchemaId}`)
      .selectOption(fieldKey);
    await page
      .getByTestId(`generic-edit-action-${commLogViewSchemaId}`)
      .selectOption("add");
    await page
      .getByTestId(`generic-edit-value-${commLogViewSchemaId}`)
      .selectOption(existingParty.record_id as string);
    await page
      .getByTestId(`generic-edit-submit-${commLogViewSchemaId}`)
      .click();
    await assertCommContextStable();

    rows = await queryViewRows(page, incidentId, commLogViewSchemaId);
    let refreshedComm = rows.find((row) => row.record_id === commLog.record_id);
    expect(refreshedComm?.cells["comm_log.audience"]?.value).toBe(
      "Browser Leadership Audience",
    );
    expect(collectionItemRefs(refreshedComm?.cells[fieldKey]?.value)).toContain(
      `party_ref:${existingParty.record_id}`,
    );

    await page
      .getByTestId(`generic-edit-action-${commLogViewSchemaId}`)
      .selectOption("remove");
    await page
      .getByTestId(`generic-edit-value-${commLogViewSchemaId}`)
      .selectOption(`party_ref:${existingParty.record_id}`);
    await page
      .getByTestId(`generic-edit-submit-${commLogViewSchemaId}`)
      .click();
    await assertCommContextStable();

    rows = await queryViewRows(page, incidentId, commLogViewSchemaId);
    refreshedComm = rows.find((row) => row.record_id === commLog.record_id);
    expect(refreshedComm?.cells["comm_log.audience"]?.value).toBe(
      "Browser Leadership Audience",
    );
    expect(
      collectionItemRefs(refreshedComm?.cells[fieldKey]?.value),
    ).not.toContain(`party_ref:${existingParty.record_id}`);
  };

  await addAndRemoveCommPartyRef("comm_log.audience_party_ids");
  await addAndRemoveCommPartyRef("comm_log.attendee_party_ids");
});

test("Phase 9 E-9-05 assessment workflow keeps invalid timestamp drafts local", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E905"),
    "Phase 9 E-9-05 assessment workflow",
  );
  const subjectA = await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("e905-host-a"),
    "host.display_name": "Phase 9 Assessment Host A",
    "host.hostname": "phase9-assessment-a.example.test",
  });
  const subjectB = await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("e905-host-b"),
    "host.display_name": "Phase 9 Assessment Host B",
    "host.hostname": "phase9-assessment-b.example.test",
  });
  const support = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("e905-support"),
    "timeline.summary": "Phase 9 assessment support event",
  });

  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      assessmentsViewSchemaId,
    )}`,
  );
  await expect(page.getByText("Timeline workbook shell")).toBeVisible();
  await expect(page.getByTestId("assessment-create-panel")).toBeVisible();
  await expect(page.getByTestId("assessment-create-subject")).toHaveValue(
    subjectA.record_id as string,
  );

  const invalidTimestamp = "2026-04-24 12:00:00";
  await page.getByTestId("assessment-create-state").selectOption("confirmed");
  await page
    .getByTestId("assessment-create-confidence-band")
    .selectOption("high");
  await page
    .getByTestId("assessment-create-rationale")
    .fill("Invalid timestamp must remain a draft.");
  await page.getByTestId("assessment-create-assessed-at").fill(invalidTimestamp);
  const failedCreate = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response.url().endsWith(`/views/${assessmentsViewSchemaId}/rows`),
  );
  await page.getByTestId("assessment-create-submit").click();
  expect((await failedCreate).status()).toBe(400);
  await expect(page.getByTestId("assessment-create-message")).toContainText(
    "invalid_mutation_payload",
  );
  await expect(page.getByTestId("assessment-create-assessed-at")).toHaveValue(
    invalidTimestamp,
  );
  expect(
    await queryViewRows(page, incidentId, assessmentsViewSchemaId),
  ).toHaveLength(0);

  const createdUnknown = await createAssessmentViaUI(page, {
    assessedAt: "2026-04-24T10:00:00Z",
    confidenceBand: "unset",
    rationale: "Phase 9 unknown rationale.",
    state: "unknown",
    supportRecordIds: [],
  });
  const createdSuspected = await createAssessmentViaUI(page, {
    assessedAt: "2026-04-24T11:00:00Z",
    confidenceBand: "low",
    rationale: "Phase 9 suspected rationale.",
    state: "suspected",
    supportRecordIds: [],
  });
  const createdConfirmed = await createAssessmentViaUI(page, {
    assessedAt: "2026-04-24T12:00:00Z",
    confidenceBand: "medium",
    rationale: "Phase 9 confirmed rationale.",
    state: "confirmed",
    supportRecordIds: [support.record_id as string],
  });

  await page
    .getByTestId("assessment-create-subject")
    .selectOption(subjectB.record_id as string);
  const createdDisproven = await createAssessmentViaUI(page, {
    assessedAt: "2026-04-24T13:00:00Z",
    confidenceBand: "medium",
    rationale: "Phase 9 disproven rationale.",
    state: "disproven",
    supportRecordIds: [],
  });
  const createdCleared = await createAssessmentViaUI(page, {
    assessedAt: "2026-04-24T14:00:00Z",
    confidenceBand: "high",
    rationale: "Phase 9 cleared rationale.",
    state: "cleared",
    supportRecordIds: [],
  });

  await expectAssessmentGridOrder(page, [
    createdCleared.record_id,
    createdDisproven.record_id,
    createdConfirmed.record_id,
    createdSuspected.record_id,
    createdUnknown.record_id,
  ]);
  await expect(
    page.getByTestId(
      rowCellTestId(createdDisproven.record_id, "assessment.subject_ref"),
    ),
  ).toHaveText(subjectB.record_id as string);
  await expect(
    page.getByTestId(
      rowCellTestId(
        createdConfirmed.record_id,
        "assessment.supporting_link_count",
      ),
    ),
  ).toHaveText("1");
  await expect(
    page.getByTestId(
      rowCellTestId(createdUnknown.record_id, "assessment.confidence_band"),
    ),
  ).toHaveText("unset");
  await expect(
    page.getByTestId(
      rowCellTestId(createdSuspected.record_id, "assessment.confidence_band"),
    ),
  ).toHaveText("low");
  await expect(
    page.getByTestId(
      rowCellTestId(createdConfirmed.record_id, "assessment.confidence_band"),
    ),
  ).toHaveText("medium");
  await expect(
    page.getByTestId(
      rowCellTestId(createdCleared.record_id, "assessment.confidence_band"),
    ),
  ).toHaveText("high");

  await applyFilterChip(
    page,
    assessmentsViewSchemaId,
    "assessment.assessment_state",
    "disproven",
  );
  await expectAssessmentGridOrder(page, [createdDisproven.record_id]);
  await removeFilterChip(
    page,
    assessmentsViewSchemaId,
    "assessment.assessment_state",
  );
  await applyFilterChip(
    page,
    assessmentsViewSchemaId,
    "assessment.assessment_state",
    "cleared",
  );
  await expectAssessmentGridOrder(page, [createdCleared.record_id]);
  await removeFilterChip(
    page,
    assessmentsViewSchemaId,
    "assessment.assessment_state",
  );
  await applyFilterChip(
    page,
    assessmentsViewSchemaId,
    "assessment.confidence_band",
    "high",
  );
  await expectAssessmentGridOrder(page, [createdCleared.record_id]);
});

test("Phase 9 E-9-06 Sprint 0 blocker sentinel", async () => {
  expect(phase9Sprint0SentinelMessage).toContain("blocker sentinel");
});

test("Phase 9 E-9-07 Sprint 0 blocker sentinel", async () => {
  expect(phase9Sprint0SentinelMessage).toContain("blocker sentinel");
});

test("Phase 9 E-9-08 Notes and Indicators open by canonical view schema IDs", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E908"),
    "Phase 9 E-9-08 registry identities",
  );
  await createViewRow(page, incidentId, notesViewSchemaId, {
    client_txn_id: uniqueTxn("e908-note"),
    "note.title": "Phase 9 registry note",
  });
  const indicator = await createViewRow(
    page,
    incidentId,
    indicatorsViewSchemaId,
    {
      client_txn_id: uniqueTxn("e908-indicator"),
      "indicator.indicator_type": "ipv4_addr",
      "indicator.value_kind": "atomic",
      "indicator.display_value": "203.0.113.92",
    },
  );

  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      notesViewSchemaId,
    )}`,
  );
  await expect(page.getByTestId("surface-tab-notes")).toBeVisible();
  await expect(
    page.getByTestId(gridShellTestId(notesViewSchemaId)),
  ).toBeVisible();
  await expect(page).toHaveURL(
    new RegExp(`view_schema_id=${encodeURIComponent(notesViewSchemaId)}`),
  );

  await page
    .getByTestId("system-view-selector")
    .selectOption(indicatorsViewSchemaId);
  await expect(
    page.getByTestId(gridShellTestId(indicatorsViewSchemaId)),
  ).toBeVisible();
  await expect(page).toHaveURL(
    new RegExp(`view_schema_id=${encodeURIComponent(indicatorsViewSchemaId)}`),
  );
  await expect(
    page.getByTestId(
      rowCellTestId(indicator.record_id as string, "indicator.display_value"),
    ),
  ).toHaveText("203.0.113.92");
});

function collectionDisplayTexts(value: unknown): string[] {
  if (!value || typeof value !== "object" || !("items" in value)) {
    return [];
  }
  const items = (value as { items?: unknown }).items;
  if (!Array.isArray(items)) {
    return [];
  }
  return items.flatMap((item) => {
    if (!item || typeof item !== "object") {
      return [];
    }
    const record = item as Record<string, unknown>;
    if (typeof record.display_text === "string") {
      return [record.display_text];
    }
    if (typeof record.raw_text === "string") {
      return [record.raw_text];
    }
    return [];
  });
}

function collectionItemRefs(value: unknown): string[] {
  if (!value || typeof value !== "object" || !("items" in value)) {
    return [];
  }
  const items = (value as { items?: unknown }).items;
  if (!Array.isArray(items)) {
    return [];
  }
  return items.flatMap((item) => {
    if (!item || typeof item !== "object") {
      return [];
    }
    const record = item as Record<string, unknown>;
    return typeof record.item_ref === "string" ? [record.item_ref] : [];
  });
}

async function openGenericSurface(
  page: Page,
  incidentId: string,
  viewSchemaId: string,
  heading: string,
) {
  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      viewSchemaId,
    )}`,
  );
  await expect(page.getByRole("heading", { name: heading })).toBeVisible();
}

async function setGenericGridScroll(page: Page, viewSchemaId: string) {
  return page
    .locator(
      `[data-testid="${gridShellTestId(viewSchemaId)}"] ${gridScrollportSelector()}`,
    )
    .evaluate((element) => {
      element.scrollTop = 16;
      element.scrollLeft = 32;
      return { top: element.scrollTop, left: element.scrollLeft };
    });
}

async function expectGenericGridScroll(
  page: Page,
  viewSchemaId: string,
  expected: { top: number; left: number },
) {
  await expect
    .poll(async () =>
      page
        .locator(
          `[data-testid="${gridShellTestId(viewSchemaId)}"] ${gridScrollportSelector()}`,
        )
        .evaluate((element) => ({
          top: element.scrollTop,
          left: element.scrollLeft,
        })),
    )
    .toEqual(expected);
}
