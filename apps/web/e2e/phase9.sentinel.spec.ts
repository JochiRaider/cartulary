import {
  applyFilterChip,
  removeFilterChip,
  sortByHeader,
} from "@cartulary/test-utils";
import {
  conflictMarkerTestId,
  dataTestIdSelector,
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  genericEditActionSelectTestId,
  genericEditFieldSelectTestId,
  genericEditRecordSelectTestId,
  genericEditValueTestId,
  gridFilterChipTestId,
  gridFilterFieldTestId,
  gridGroupingSelectTestId,
  gridGroupRowTestId,
  gridScrollportSelector,
  gridShellTestId,
  rowCellTestId,
  savedViewFamilySelector,
  savedViewOptionTestId,
  savedViewSelectorTestId,
  saveStateTestId,
  surfaceTabTestId,
  systemViewSwitcherMenuTestId,
  systemViewSwitcherOptionTestId,
  systemViewSwitcherTriggerTestId,
  timelineMutationSubstrateReadyTestId,
  workbookAddRowButtonTestId,
  workbookInspectorToggleTestId,
  workbookShellReadyTestId,
  workbookShellSlotTestId,
} from "@cartulary/ui-contracts";
import type { Page, Request } from "@playwright/test";

import { expect, test } from "./fixtures";
import {
  apiBase,
  createIncident,
  createSavedView,
  createViewRow,
  openSystemSurfaceBySwitcher,
  patchTimelineRecord,
  queryViewRows,
  uniqueIncidentKey,
  uniqueTxn,
  type ViewApiRow,
} from "./helpers";
import {
  assessmentsViewSchemaId,
  collectionItems,
  commLogViewSchemaId,
  createAssessmentViaUI,
  decisionsViewSchemaId,
  editGenericCell,
  evidenceViewSchemaId,
  expectAssessmentGridOrder,
  handoffViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  indicatorsViewSchemaId,
  lessonViewSchemaId,
  notesViewSchemaId,
  partiesViewSchemaId,
  statusReviewViewSchemaId,
  submitGenericEditAndWait,
  taskRequestsViewSchemaId,
  waitForViewRowByCell,
} from "./phase4Helpers";

const timelineViewSchemaId = "cartulary.view.timeline.v1";
const requiredBaseViewSchemaIds = [
  assessmentsViewSchemaId,
  commLogViewSchemaId,
  decisionsViewSchemaId,
  evidenceViewSchemaId,
  handoffViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  indicatorsViewSchemaId,
  lessonViewSchemaId,
  notesViewSchemaId,
  partiesViewSchemaId,
  statusReviewViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
] as const;
const optionalStandardizedSurfaceIds = [
  "cartulary.view.findings.v1",
  "cartulary.view.forensic_keywords.v1",
  "cartulary.view.investigative_queries.v1",
] as const;
const feP10WorkbookShellSurfaces = [
  {
    groupToken: "coordination",
    label: "Task Requests",
    viewSchemaId: taskRequestsViewSchemaId,
  },
  {
    groupToken: "coordination",
    label: "Decisions",
    viewSchemaId: decisionsViewSchemaId,
  },
  {
    groupToken: "scope-assessment",
    label: "Parties",
    viewSchemaId: partiesViewSchemaId,
  },
  {
    groupToken: "coordination",
    label: "Communications Log",
    viewSchemaId: commLogViewSchemaId,
  },
  {
    groupToken: "coordination",
    label: "Handoff",
    viewSchemaId: handoffViewSchemaId,
  },
  {
    groupToken: "review-learning",
    label: "Status Review",
    viewSchemaId: statusReviewViewSchemaId,
  },
  {
    groupToken: "review-learning",
    label: "Lesson",
    viewSchemaId: lessonViewSchemaId,
  },
] as const;
const findingsViewSchemaId = "cartulary.view.findings.v1";
const forensicKeywordsViewSchemaId = "cartulary.view.forensic_keywords.v1";
const investigativeQueriesViewSchemaId =
  "cartulary.view.investigative_queries.v1";
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

test("Phase 9 E-9-PASTE-02 pastes a representative 20x5 Timeline clipboard range", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E902"),
    "Phase 9 E-9-PASTE-02 clipboard paste",
  );
  const seed = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("e902-seed"),
    "timeline.summary": "Phase 9 paste seed",
  });

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toBeVisible();

  const seedSummary = page.getByTestId(
    rowCellTestId(seed.record_id as string, "timeline.summary"),
  );
  await seedSummary.focus();
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${timelineViewSchemaId}:${seed.record_id}:timeline.summary`,
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
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${timelineViewSchemaId}:${seed.record_id}:timeline.summary`,
  );
  await expect(
    page.getByTestId(
      rowCellTestId(seed.record_id as string, "timeline.summary"),
    ),
  ).toHaveValue("Phase 9 paste summary 1");
  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toBeVisible();

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

test("Phase 9 E-9-CONFLICT-02 groups paste conflicts and preserves selection continuity", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E902-CONFLICT"),
    "Phase 9 E-9-CONFLICT-02 grouped paste conflicts",
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
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toBeVisible();

  const firstSummary = page.getByTestId(
    rowCellTestId(first.record_id as string, "timeline.summary"),
  );
  await firstSummary.focus();
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${timelineViewSchemaId}:${first.record_id}:timeline.summary`,
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

  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Conflict");
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${timelineViewSchemaId}:${first.record_id}:timeline.summary`,
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
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Conflict");
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
  await page
    .getByTestId(workbookInspectorToggleTestId(notesViewSchemaId))
    .click();
  await expect(
    page
      .getByTestId("generic-create-note-source-record")
      .locator(`option[value="${source.record_id}"]`),
  ).toHaveCount(1, { timeout: 15_000 });
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
  await page.getByTestId(genericCreateSubmitTestId(notesViewSchemaId)).click();
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
    .getByTestId(genericEditRecordSelectTestId(evidenceViewSchemaId))
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
    await expect(
      page.getByTestId(gridShellTestId(evidenceViewSchemaId)),
    ).toBeVisible();
    await expect(page).toHaveURL(
      new RegExp(`view_schema_id=${encodeURIComponent(evidenceViewSchemaId)}`),
    );
    await expect(
      page.getByTestId(gridShellTestId(evidenceViewSchemaId)),
    ).toBeVisible();
    await expect(
      page.getByTestId(genericEditRecordSelectTestId(evidenceViewSchemaId)),
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
  await expect(page.getByTestId("generic-mutation-state")).toHaveText("Saved");
  await assertEvidenceContextStable();

  rows = await queryViewRows(page, incidentId, evidenceViewSchemaId);
  refreshedEvidence = rows.find((row) => row.record_id === evidence.record_id);
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
  refreshedEvidence = await applyPartyPatchAndWait(page, {
    incidentId,
    viewSchemaId: evidenceViewSchemaId,
    recordId: evidence.record_id as string,
    clickTestId: "party-link-link-existing",
    changes: [
      {
        field_key: "evidence.source_party_id",
        value: existingParty.record_id,
      },
    ],
    expectedCells: {
      "evidence.source_party_text": "Browser Source Raw",
      "evidence.source_party_id": existingParty.record_id,
    },
  });
  await assertEvidenceContextStable();
  expect(refreshedEvidence?.cells["evidence.source_party_text"]?.value).toBe(
    "Browser Source Raw",
  );
  expect(refreshedEvidence?.cells["evidence.source_party_id"]?.value).toBe(
    existingParty.record_id,
  );

  refreshedEvidence = await applyPartyPatchAndWait(page, {
    incidentId,
    viewSchemaId: evidenceViewSchemaId,
    recordId: evidence.record_id as string,
    clickTestId: "party-link-clear-link",
    changes: [{ field_key: "evidence.source_party_id", value: null }],
    expectedCells: {
      "evidence.source_party_text": "Browser Source Raw",
      "evidence.source_party_id": null,
    },
  });
  await assertEvidenceContextStable();
  expect(refreshedEvidence?.cells["evidence.source_party_text"]?.value).toBe(
    "Browser Source Raw",
  );
  expect(
    refreshedEvidence?.cells["evidence.source_party_id"]?.value,
  ).toBeNull();

  await page
    .getByTestId("party-link-existing-party")
    .selectOption(existingParty.record_id as string);
  refreshedEvidence = await applyPartyPatchAndWait(page, {
    incidentId,
    viewSchemaId: evidenceViewSchemaId,
    recordId: evidence.record_id as string,
    clickTestId: "party-link-link-existing",
    changes: [
      {
        field_key: "evidence.source_party_id",
        value: existingParty.record_id,
      },
    ],
    expectedCells: {
      "evidence.source_party_text": "Browser Source Raw",
      "evidence.source_party_id": existingParty.record_id,
    },
  });
  await assertEvidenceContextStable();
  refreshedEvidence = await applyPartyPatchAndWait(page, {
    incidentId,
    viewSchemaId: evidenceViewSchemaId,
    recordId: evidence.record_id as string,
    clickTestId: "party-link-clear-text",
    changes: [{ field_key: "evidence.source_party_text", value: null }],
    expectedCells: {
      "evidence.source_party_text": null,
      "evidence.source_party_id": existingParty.record_id,
    },
  });
  await assertEvidenceContextStable();
  expect(
    refreshedEvidence?.cells["evidence.source_party_text"]?.value,
  ).toBeNull();
  expect(refreshedEvidence?.cells["evidence.source_party_id"]?.value).toBe(
    existingParty.record_id,
  );

  await page
    .getByTestId("party-link-existing-party")
    .selectOption(existingParty.record_id as string);
  refreshedEvidence = await applyPartyPatchAndWait(page, {
    incidentId,
    viewSchemaId: evidenceViewSchemaId,
    recordId: evidence.record_id as string,
    clickTestId: "party-link-clear-both",
    changes: [
      { field_key: "evidence.source_party_text", value: null },
      { field_key: "evidence.source_party_id", value: null },
    ],
    expectedCells: {
      "evidence.source_party_text": null,
      "evidence.source_party_id": null,
    },
  });
  await assertEvidenceContextStable();
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
    .getByTestId(genericEditRecordSelectTestId(taskRequestsViewSchemaId))
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
      page.getByTestId(gridShellTestId(taskRequestsViewSchemaId)),
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
      page.getByTestId(genericEditRecordSelectTestId(taskRequestsViewSchemaId)),
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
  await expect(page.getByTestId("generic-mutation-state")).toHaveText("Saved");
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
  refreshedTask = await applyPartyPatchAndWait(page, {
    incidentId,
    viewSchemaId: taskRequestsViewSchemaId,
    recordId: task.record_id as string,
    clickTestId: "party-link-link-existing",
    changes: [
      {
        field_key: "task.requester_party_id",
        value: existingParty.record_id,
      },
    ],
    expectedCells: {
      "task.requester_party_text": typedRequesterText,
      "task.requester_party_id": existingParty.record_id,
    },
  });
  await assertTaskContextStable();
  expect(refreshedTask?.cells["task.requester_party_text"]?.value).toBe(
    typedRequesterText,
  );
  expect(refreshedTask?.cells["task.requester_party_id"]?.value).toBe(
    existingParty.record_id,
  );

  refreshedTask = await applyPartyPatchAndWait(page, {
    incidentId,
    viewSchemaId: taskRequestsViewSchemaId,
    recordId: task.record_id as string,
    clickTestId: "party-link-clear-link",
    changes: [{ field_key: "task.requester_party_id", value: null }],
    expectedCells: {
      "task.requester_party_text": typedRequesterText,
      "task.requester_party_id": null,
    },
  });
  await assertTaskContextStable();
  expect(refreshedTask?.cells["task.requester_party_text"]?.value).toBe(
    typedRequesterText,
  );
  expect(refreshedTask?.cells["task.requester_party_id"]?.value).toBeNull();

  await page
    .getByTestId("party-link-existing-party")
    .selectOption(existingParty.record_id as string);
  refreshedTask = await applyPartyPatchAndWait(page, {
    incidentId,
    viewSchemaId: taskRequestsViewSchemaId,
    recordId: task.record_id as string,
    clickTestId: "party-link-link-existing",
    changes: [
      {
        field_key: "task.requester_party_id",
        value: existingParty.record_id,
      },
    ],
    expectedCells: {
      "task.requester_party_text": typedRequesterText,
      "task.requester_party_id": existingParty.record_id,
    },
  });
  await assertTaskContextStable();
  refreshedTask = await applyPartyPatchAndWait(page, {
    incidentId,
    viewSchemaId: taskRequestsViewSchemaId,
    recordId: task.record_id as string,
    clickTestId: "party-link-clear-text",
    changes: [{ field_key: "task.requester_party_text", value: null }],
    expectedCells: {
      "task.requester_party_text": null,
      "task.requester_party_id": existingParty.record_id,
    },
  });
  await assertTaskContextStable();
  expect(refreshedTask?.cells["task.requester_party_text"]?.value).toBeNull();
  expect(refreshedTask?.cells["task.requester_party_id"]?.value).toBe(
    existingParty.record_id,
  );

  await page
    .getByTestId("party-link-existing-party")
    .selectOption(existingParty.record_id as string);
  refreshedTask = await applyPartyPatchAndWait(page, {
    incidentId,
    viewSchemaId: taskRequestsViewSchemaId,
    recordId: task.record_id as string,
    clickTestId: "party-link-clear-both",
    changes: [
      { field_key: "task.requester_party_text", value: null },
      { field_key: "task.requester_party_id", value: null },
    ],
    expectedCells: {
      "task.requester_party_text": null,
      "task.requester_party_id": null,
    },
  });
  await assertTaskContextStable();
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
    .getByTestId(genericEditRecordSelectTestId(commLogViewSchemaId))
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
      page.getByTestId(gridShellTestId(commLogViewSchemaId)),
    ).toBeVisible();
    await expect(page).toHaveURL(
      new RegExp(`view_schema_id=${encodeURIComponent(commLogViewSchemaId)}`),
    );
    await expect(
      page.getByTestId(gridShellTestId(commLogViewSchemaId)),
    ).toBeVisible();
    await expect(
      page.getByTestId(genericEditRecordSelectTestId(commLogViewSchemaId)),
    ).toHaveValue(commLog.record_id as string);
    await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
      `${commLogViewSchemaId}:${commLog.record_id}:comm_log.summary`,
    );
    await expectGenericGridScroll(page, commLogViewSchemaId, commScroll);
  };
  const addAndRemoveCommPartyRef = async (fieldKey: string) => {
    await page
      .getByTestId(genericEditFieldSelectTestId(commLogViewSchemaId))
      .selectOption(fieldKey);
    await page
      .getByTestId(genericEditActionSelectTestId(commLogViewSchemaId))
      .selectOption("add");
    await page
      .getByTestId(genericEditValueTestId(commLogViewSchemaId))
      .selectOption(existingParty.record_id as string);
    await submitGenericEditAndWait(
      page,
      commLogViewSchemaId,
      commLog.record_id as string,
    );
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
      .getByTestId(genericEditActionSelectTestId(commLogViewSchemaId))
      .selectOption("remove");
    await page
      .getByTestId(genericEditValueTestId(commLogViewSchemaId))
      .selectOption(`party_ref:${existingParty.record_id}`);
    await submitGenericEditAndWait(
      page,
      commLogViewSchemaId,
      commLog.record_id as string,
    );
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
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await page
    .getByTestId(workbookAddRowButtonTestId(assessmentsViewSchemaId))
    .click();
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
  await page
    .getByTestId("assessment-create-assessed-at")
    .fill(invalidTimestamp);
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

test("Phase 9 E-9-TASKDECISION-06 Task Request and Decision workbook workflows stay native", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E906"),
    "Phase 9 E-9-TASKDECISION-06 Task and Decision workflows",
  );
  const support = await createViewRow(page, incidentId, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn("e906-support"),
    "evidence.title": "E-9-06 supporting packet",
  });
  const targetDecision = await createViewRow(
    page,
    incidentId,
    decisionsViewSchemaId,
    {
      client_txn_id: uniqueTxn("e906-target-decision"),
      "decision.summary": "E-9-06 target decision",
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
      "decision.summary": "E-9-06 superseding decision",
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
  await openGenericSurface(
    page,
    incidentId,
    decisionsViewSchemaId,
    "Decisions",
  );
  await expect(page).toHaveURL(
    new RegExp(`view_schema_id=${encodeURIComponent(decisionsViewSchemaId)}`),
  );
  await expect(
    page.getByTestId(gridShellTestId(decisionsViewSchemaId)),
  ).toBeVisible();
  await expect(
    page
      .getByTestId("decision-supersede-replacement")
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
    .getByTestId("decision-supersede-target")
    .selectOption(targetDecision.record_id as string);
  await page
    .getByTestId("decision-supersede-replacement")
    .selectOption(supersedingDecision.record_id as string);
  await page
    .getByTestId("decision-supersede-reason")
    .fill("E-9-06 explicit supersession");
  await page.getByTestId("decision-supersede-submit").click();
  const supersedeEnvelope = await (await supersedeResponse).json();
  expect(supersedeEnvelope.data.view_schema_id).toBe(decisionsViewSchemaId);
  expect(supersedeEnvelope.data.target_record_id).toBe(
    targetDecision.record_id,
  );
  expect(supersedeEnvelope.data.superseding_record_id).toBe(
    supersedingDecision.record_id,
  );
  await expect(page.getByTestId("generic-mutation-state")).toHaveText("Saved");
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
  await page
    .getByTestId(genericEditRecordSelectTestId(decisionsViewSchemaId))
    .selectOption(supersedingDecision.record_id as string);
  await page
    .getByTestId(genericEditFieldSelectTestId(decisionsViewSchemaId))
    .selectOption("decision.affected_record_ids");
  await waitForPhase9GenericOption(
    page,
    genericEditValueTestId(decisionsViewSchemaId),
    support.record_id as string,
  );
  await page
    .getByTestId(genericEditValueTestId(decisionsViewSchemaId))
    .selectOption(support.record_id as string);
  await submitGenericEditAndWait(
    page,
    decisionsViewSchemaId,
    supersedingDecision.record_id as string,
  );
  await expect(page.getByTestId("generic-mutation-state")).toHaveText("Saved");
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

  await openGenericSurface(
    page,
    incidentId,
    taskRequestsViewSchemaId,
    "Task Requests",
  );
  await expect(page).toHaveURL(
    new RegExp(
      `view_schema_id=${encodeURIComponent(taskRequestsViewSchemaId)}`,
    ),
  );
  await expect(
    page.getByTestId(gridShellTestId(taskRequestsViewSchemaId)),
  ).toBeVisible();
  const dueAt = "2026-05-19T16:00:00Z";
  await setPhase9GenericCreateField(page, "task.title", "E-9-06 workbook task");
  await setPhase9GenericCreateField(page, "task.task_kind", "collection");
  await setPhase9GenericCreateField(page, "task.workstream", "forensics");
  await setPhase9GenericCreateField(page, "task.due_at", dueAt);
  await setPhase9GenericCreateField(
    page,
    "task.external_ticket_ref",
    "SOC-E906",
  );
  await waitForPhase9GenericOption(
    page,
    "generic-create-field-task.decision_record_id",
    supersedingDecision.record_id as string,
  );
  await setPhase9GenericCreateField(
    page,
    "task.decision_record_id",
    supersedingDecision.record_id as string,
  );
  await waitForPhase9GenericOption(
    page,
    "generic-create-field-task.linked_record_ids",
    support.record_id as string,
  );
  await setPhase9GenericCreateField(
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
    "E-9-06 workbook task",
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

  await setPhase9GenericCreateField(
    page,
    "task.title",
    "E-9-06 urgent queue task",
  );
  await setPhase9GenericCreateField(page, "task.task_kind", "follow_up");
  await setPhase9GenericCreateField(page, "task.priority", "urgent");
  await page
    .getByTestId(genericCreateSubmitTestId(taskRequestsViewSchemaId))
    .click();
  const urgentTask = await waitForViewRowByCell(
    page,
    incidentId,
    taskRequestsViewSchemaId,
    "task.title",
    "E-9-06 urgent queue task",
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
  await expect(page.getByTestId("generic-mutation-state")).toHaveText("Saved");
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
  await page
    .getByTestId("task-lifecycle-target")
    .selectOption(task.record_id as string);
  await page.getByTestId("task-lifecycle-status").selectOption("blocked");
  await page
    .getByTestId("task-lifecycle-blocked-reason")
    .fill("Waiting on endpoint owner");
  await page.getByTestId("task-lifecycle-submit").click();
  await lifecycleResponse;
  await expect(page.getByTestId("generic-mutation-state")).toHaveText("Saved");
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
  await editGenericCell(
    page,
    taskRequestsViewSchemaId,
    task.record_id,
    "task.decision_record_id",
    "",
  );
  const clearRequest = await clearRequestPromise;
  expect(clearRequest.postDataJSON()).toMatchObject({
    view_schema_id: taskRequestsViewSchemaId,
    changes: [{ field_key: "task.decision_record_id", value: null }],
  });
  await expect(page.getByTestId("generic-mutation-state")).toHaveText("Saved");
  taskRows = await queryViewRows(page, incidentId, taskRequestsViewSchemaId);
  refreshedTask = taskRows.find((row) => row.record_id === task.record_id);
  expect(refreshedTask?.cells["task.decision_record_id"]?.value).toBeNull();
  expect(refreshedTask?.cells["task.blocked_reason"]?.value).toBe(
    "Waiting on endpoint owner",
  );

  decisionRows = await queryViewRows(page, incidentId, decisionsViewSchemaId);
  expect(
    decisionRows.find((row) => row.record_id === supersedingDecision.record_id)
      ?.cells["decision.supersedes_record_id"]?.value,
  ).toBe(targetDecision.record_id);
});

test("Phase 9 E-9-COORDINATION-06 coordination workbook workflows stay native", async ({
  page,
  workerAdmin,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E906COORD"),
    "Phase 9 E-9-COORDINATION-06 coordination workflows",
  );
  const party = await createViewRow(page, incidentId, partiesViewSchemaId, {
    client_txn_id: uniqueTxn("e906coord-party"),
    "party.display_name": "E-9-06 Coordination Lead",
    "party.party_kind": "team",
  });
  const evidence = await createViewRow(page, incidentId, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn("e906coord-evidence"),
    "evidence.title": "E-9-06 coordination evidence",
  });
  const decision = await createViewRow(
    page,
    incidentId,
    decisionsViewSchemaId,
    {
      client_txn_id: uniqueTxn("e906coord-decision"),
      "decision.summary": "E-9-06 coordination decision",
      "decision.decision_type": "containment",
      "decision.rationale": "Coordination workflow decision.",
    },
  );
  const task = await createViewRow(page, incidentId, taskRequestsViewSchemaId, {
    client_txn_id: uniqueTxn("e906coord-task"),
    "task.title": "E-9-06 coordination task",
    "task.task_kind": "follow_up",
  });

  await disableWorkbookSockets(page);
  await openGenericSurface(
    page,
    incidentId,
    commLogViewSchemaId,
    "Communications Log",
  );
  await expect(page).toHaveURL(
    new RegExp(`view_schema_id=${encodeURIComponent(commLogViewSchemaId)}`),
  );
  await setPhase9GenericCreateField(page, "comm_log.comm_type", "briefing");
  await setPhase9GenericCreateField(
    page,
    "comm_log.audience",
    "E-9-06 coordination audience",
  );
  await setPhase9GenericCreateField(
    page,
    "comm_log.channel_or_meeting",
    "Coordination bridge",
  );
  await setPhase9GenericCreateField(
    page,
    "comm_log.summary",
    "E-9-06 coordination log",
  );
  await waitForPhase9GenericOption(
    page,
    "generic-create-field-comm_log.decision_ids",
    decision.record_id as string,
  );
  await setPhase9GenericCreateField(
    page,
    "comm_log.decision_ids",
    decision.record_id as string,
  );
  await waitForPhase9GenericOption(
    page,
    "generic-create-field-comm_log.action_task_ids",
    task.record_id as string,
  );
  await setPhase9GenericCreateField(
    page,
    "comm_log.action_task_ids",
    task.record_id as string,
  );
  await waitForPhase9GenericOption(
    page,
    "generic-create-field-comm_log.audience_party_ids",
    party.record_id as string,
  );
  await setPhase9GenericCreateField(
    page,
    "comm_log.audience_party_ids",
    party.record_id as string,
  );
  await page
    .getByTestId(genericCreateSubmitTestId(commLogViewSchemaId))
    .click();
  const comm = await waitForViewRowByCell(
    page,
    incidentId,
    commLogViewSchemaId,
    "comm_log.summary",
    "E-9-06 coordination log",
  );
  expect(comm.cells["comm_log.audience"]?.value).toBe(
    "E-9-06 coordination audience",
  );
  expect(collectionItems(comm, "comm_log.decision_ids")).toHaveLength(1);
  expect(collectionItems(comm, "comm_log.action_task_ids")).toHaveLength(1);
  expect(collectionItems(comm, "comm_log.audience_party_ids")).toHaveLength(1);

  await openGenericSurface(page, incidentId, handoffViewSchemaId, "Handoff");
  await expect(page).toHaveURL(
    new RegExp(`view_schema_id=${encodeURIComponent(handoffViewSchemaId)}`),
  );
  await setPhase9GenericCreateField(
    page,
    "handoff.current_state_summary",
    "E-9-06 handoff state",
  );
  await page
    .getByTestId(genericCreateSubmitTestId(handoffViewSchemaId))
    .click();
  await expect(page.getByTestId("generic-mutation-error")).toContainText(
    "Incoming owner",
  );
  await setPhase9GenericCreateField(
    page,
    "handoff.incoming_owner_user_id",
    workerAdmin.user_id,
  );
  await page
    .getByTestId(genericCreateSubmitTestId(handoffViewSchemaId))
    .click();
  const handoff = await waitForViewRowByCell(
    page,
    incidentId,
    handoffViewSchemaId,
    "handoff.current_state_summary",
    "E-9-06 handoff state",
  );
  await editGenericCell(
    page,
    handoffViewSchemaId,
    handoff.record_id,
    "handoff.next_checks",
    "Verify next checkpoint before shift change",
  );
  await expect(page.getByTestId("generic-mutation-state")).toHaveText("Saved");
  await editGenericCell(
    page,
    handoffViewSchemaId,
    handoff.record_id,
    "handoff.open_risk_refs",
    "Residual endpoint access risk",
  );
  await expect(page.getByTestId("generic-mutation-state")).toHaveText("Saved");
  await editGenericCell(
    page,
    handoffViewSchemaId,
    handoff.record_id,
    "handoff.acknowledged_at",
    "2026-05-26T14:00:00Z",
  );
  await expect(page.getByTestId("generic-mutation-state")).toHaveText("Saved");
  const handoffRows = await queryViewRows(
    page,
    incidentId,
    handoffViewSchemaId,
  );
  const refreshedHandoff = handoffRows.find(
    (row) => row.record_id === handoff.record_id,
  );
  expect(refreshedHandoff?.cells["handoff.next_checks"]?.value).toBe(
    "Verify next checkpoint before shift change",
  );
  expect(refreshedHandoff?.cells["handoff.ack_state"]?.value).toBe(
    "acknowledged",
  );
  if (!refreshedHandoff) {
    throw new Error("missing refreshed handoff");
  }
  expect(
    collectionItems(refreshedHandoff, "handoff.open_risk_refs"),
  ).toHaveLength(1);
  await applyFilterChip(
    page,
    handoffViewSchemaId,
    "handoff.ack_state",
    "acknowledged",
  );
  await expect(
    page.getByTestId(
      gridFilterChipTestId(handoffViewSchemaId, "handoff.ack_state"),
    ),
  ).toContainText("acknowledged");
  await removeFilterChip(page, handoffViewSchemaId, "handoff.ack_state");

  await openGenericSurface(
    page,
    incidentId,
    statusReviewViewSchemaId,
    "Status Review",
  );
  await setPhase9GenericCreateField(
    page,
    "status_review.current_state_summary",
    "E-9-06 status review state",
  );
  await page
    .getByTestId(genericCreateSubmitTestId(statusReviewViewSchemaId))
    .click();
  const status = await waitForViewRowByCell(
    page,
    incidentId,
    statusReviewViewSchemaId,
    "status_review.current_state_summary",
    "E-9-06 status review state",
  );
  await editPhase9GenericCell(
    page,
    statusReviewViewSchemaId,
    status.record_id,
    "status_review.blocked_task_ids",
    task.record_id as string,
  );
  await expect(page.getByTestId("generic-mutation-state")).toHaveText("Saved");
  await editPhase9GenericCell(
    page,
    statusReviewViewSchemaId,
    status.record_id,
    "status_review.pending_evidence_ids",
    evidence.record_id as string,
  );
  await expect(page.getByTestId("generic-mutation-state")).toHaveText("Saved");
  await editPhase9GenericCell(
    page,
    statusReviewViewSchemaId,
    status.record_id,
    "status_review.open_decision_ids",
    decision.record_id as string,
  );
  await expect(page.getByTestId("generic-mutation-state")).toHaveText("Saved");
  await editGenericCell(
    page,
    statusReviewViewSchemaId,
    status.record_id,
    "status_review.next_report_at",
    "2026-05-27T15:30:00Z",
  );
  await expect(page.getByTestId("generic-mutation-state")).toHaveText("Saved");
  const statusRows = await queryViewRows(
    page,
    incidentId,
    statusReviewViewSchemaId,
  );
  const refreshedStatus = statusRows.find(
    (row) => row.record_id === status.record_id,
  );
  if (!refreshedStatus) {
    throw new Error("missing refreshed status review");
  }
  expect(
    collectionItems(refreshedStatus, "status_review.blocked_task_ids"),
  ).toHaveLength(1);
  expect(
    collectionItems(refreshedStatus, "status_review.pending_evidence_ids"),
  ).toHaveLength(1);
  expect(
    collectionItems(refreshedStatus, "status_review.open_decision_ids"),
  ).toHaveLength(1);
  expect(
    Date.parse(
      refreshedStatus.cells["status_review.next_report_at"]?.value as string,
    ),
  ).toBe(Date.parse("2026-05-27T15:30:00Z"));
  await applyFilterChip(
    page,
    statusReviewViewSchemaId,
    "status_review.next_report_day",
    "2026-05-27",
  );
  await expect(
    page.getByTestId(
      gridFilterChipTestId(
        statusReviewViewSchemaId,
        "status_review.next_report_day",
      ),
    ),
  ).toContainText("2026-05-27");
  await removeFilterChip(
    page,
    statusReviewViewSchemaId,
    "status_review.next_report_day",
  );

  await openGenericSurface(page, incidentId, lessonViewSchemaId, "Lesson");
  await setPhase9GenericCreateField(
    page,
    "lesson.summary",
    "E-9-06 lesson workflow",
  );
  await page.getByTestId(genericCreateSubmitTestId(lessonViewSchemaId)).click();
  const lesson = await waitForViewRowByCell(
    page,
    incidentId,
    lessonViewSchemaId,
    "lesson.summary",
    "E-9-06 lesson workflow",
  );
  await editPhase9GenericCell(
    page,
    lessonViewSchemaId,
    lesson.record_id,
    "lesson.follow_up_task_ids",
    task.record_id as string,
  );
  await expect(page.getByTestId("generic-mutation-state")).toHaveText("Saved");
  await editPhase9GenericCell(
    page,
    lessonViewSchemaId,
    lesson.record_id,
    "lesson.evidence_refs",
    evidence.record_id as string,
  );
  await expect(page.getByTestId("generic-mutation-state")).toHaveText("Saved");
  await editGenericCell(
    page,
    lessonViewSchemaId,
    lesson.record_id,
    "lesson.closure_state",
    "closed",
  );
  await expect(page.getByTestId("generic-mutation-state")).toHaveText("Saved");
  const lessonRows = await queryViewRows(page, incidentId, lessonViewSchemaId);
  const refreshedLesson = lessonRows.find(
    (row) => row.record_id === lesson.record_id,
  );
  if (!refreshedLesson) {
    throw new Error("missing refreshed lesson");
  }
  expect(refreshedLesson.cells["lesson.closure_state"]?.value).toBe("closed");
  expect(
    collectionItems(refreshedLesson, "lesson.follow_up_task_ids"),
  ).toHaveLength(1);
  expect(collectionItems(refreshedLesson, "lesson.evidence_refs")).toHaveLength(
    1,
  );
  await applyFilterChip(
    page,
    lessonViewSchemaId,
    "lesson.closure_state",
    "closed",
  );
  await expect(
    page.getByTestId(
      gridFilterChipTestId(lessonViewSchemaId, "lesson.closure_state"),
    ),
  ).toContainText("closed");
});

test("Phase 9 E-9-07 optional standardized surfaces are workbook-native when exposed", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E907"),
    "Phase 9 E-9-07 optional surfaces",
  );
  const supportingNote = await createViewRow(
    page,
    incidentId,
    notesViewSchemaId,
    {
      client_txn_id: uniqueTxn("e907-note"),
      "note.title": "E-9-07 supporting note",
    },
  );

  await expectOptionalStandardizedSurfacesExposed(page);

  await openGenericSurface(page, incidentId, findingsViewSchemaId, "Findings");
  await expect(page).toHaveURL(
    new RegExp(`view_schema_id=${encodeURIComponent(findingsViewSchemaId)}`),
  );
  await setPhase9GenericCreateField(
    page,
    "finding.statement",
    "Hypothesis: exfiltration used compressed archives",
  );
  await setPhase9GenericCreateField(page, "finding.kind", "hypothesis");
  await setPhase9GenericCreateField(page, "finding.confidence_score", "72");
  await page
    .getByTestId(genericCreateSubmitTestId(findingsViewSchemaId))
    .click();
  const finding = await waitForViewRowByCell(
    page,
    incidentId,
    findingsViewSchemaId,
    "finding.statement",
    "Hypothesis: exfiltration used compressed archives",
  );
  await editPhase9GenericCell(
    page,
    findingsViewSchemaId,
    finding.record_id,
    "finding.state",
    "closed",
  );
  await expect(page.getByTestId("generic-mutation-state")).toHaveText("Saved");
  await editPhase9GenericCell(
    page,
    findingsViewSchemaId,
    finding.record_id,
    "finding.supporting_refs",
    supportingNote.record_id as string,
  );
  await expect(page.getByTestId("generic-mutation-state")).toHaveText("Saved");
  const findingRows = await queryViewRows(
    page,
    incidentId,
    findingsViewSchemaId,
  );
  const refreshedFinding = findingRows.find(
    (row) => row.record_id === finding.record_id,
  );
  if (!refreshedFinding) {
    throw new Error("missing refreshed finding");
  }
  expect(refreshedFinding.cells["finding.kind"]?.value).toBe("hypothesis");
  expect(refreshedFinding.cells["finding.state"]?.value).toBe("closed");
  expect(refreshedFinding.cells["finding.confidence_band"]?.value).toBe("high");
  expect(typeof refreshedFinding.cells["finding.closed_at"]?.value).toBe(
    "string",
  );
  expect(
    collectionItems(refreshedFinding, "finding.supporting_refs"),
  ).toHaveLength(1);
  await applyFilterChip(
    page,
    findingsViewSchemaId,
    "finding.confidence_band",
    "high",
  );
  await expect(
    page.getByTestId(
      gridFilterChipTestId(findingsViewSchemaId, "finding.confidence_band"),
    ),
  ).toContainText("high");
  await page
    .getByTestId(gridGroupingSelectTestId(findingsViewSchemaId))
    .selectOption("finding.kind");
  await expect(
    page.getByTestId(
      gridGroupRowTestId(findingsViewSchemaId, "finding.kind", "hypothesis"),
    ),
  ).toBeVisible();

  await openGenericSurface(
    page,
    incidentId,
    investigativeQueriesViewSchemaId,
    "Investigative Queries",
  );
  await setPhase9GenericCreateField(
    page,
    "investigative_query.platform",
    "Kusto",
  );
  await setPhase9GenericCreateField(
    page,
    "investigative_query.purpose",
    "Locate archive staging",
  );
  await setPhase9GenericCreateField(
    page,
    "investigative_query.query_text",
    "DeviceFileEvents | where FileName endswith '.zip'",
  );
  await page
    .getByTestId(genericCreateSubmitTestId(investigativeQueriesViewSchemaId))
    .click();
  const investigativeQuery = await waitForViewRowByCell(
    page,
    incidentId,
    investigativeQueriesViewSchemaId,
    "investigative_query.purpose",
    "Locate archive staging",
  );
  expect(
    typeof investigativeQuery.cells["investigative_query.created_by_user_id"]
      ?.value,
  ).toBe("string");
  await applyFilterChip(
    page,
    investigativeQueriesViewSchemaId,
    "investigative_query.platform",
    "Kusto",
  );
  await expect(
    page.getByTestId(
      gridFilterChipTestId(
        investigativeQueriesViewSchemaId,
        "investigative_query.platform",
      ),
    ),
  ).toContainText("Kusto");

  await openGenericSurface(
    page,
    incidentId,
    forensicKeywordsViewSchemaId,
    "Forensic Keywords",
  );
  await setPhase9GenericCreateField(
    page,
    "forensic_keyword.pattern",
    "7z\\.exe",
  );
  await setPhase9GenericCreateField(
    page,
    "forensic_keyword.reason",
    "Archive utility execution",
  );
  await setPhase9GenericCreateField(
    page,
    "forensic_keyword.match_mode",
    "regex",
  );
  await setPhase9GenericCreateField(
    page,
    "forensic_keyword.case_sensitive",
    "true",
  );
  await page
    .getByTestId(genericCreateSubmitTestId(forensicKeywordsViewSchemaId))
    .click();
  const forensicKeyword = await waitForViewRowByCell(
    page,
    incidentId,
    forensicKeywordsViewSchemaId,
    "forensic_keyword.pattern",
    "7z\\.exe",
  );
  expect(forensicKeyword.cells["forensic_keyword.match_mode"]?.value).toBe(
    "regex",
  );
  expect(forensicKeyword.cells["forensic_keyword.case_sensitive"]?.value).toBe(
    true,
  );

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(
    page.getByTestId(systemViewSwitcherTriggerTestId()),
  ).toBeVisible();

  const optionValues = await systemViewSelectorValues(page);
  for (const viewSchemaId of optionalStandardizedSurfaceIds) {
    expect(optionValues).toContain(viewSchemaId);
  }
  for (const viewSchemaId of [
    assessmentsViewSchemaId,
    commLogViewSchemaId,
    decisionsViewSchemaId,
    indicatorsViewSchemaId,
    handoffViewSchemaId,
    lessonViewSchemaId,
    partiesViewSchemaId,
    statusReviewViewSchemaId,
    taskRequestsViewSchemaId,
  ]) {
    expect(optionValues).toContain(viewSchemaId);
  }
});

test("Phase 9 E-9-08 required registry identities stay canonical with optional additions", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E908"),
    "Phase 9 E-9-08 registry identities",
  );
  await expectRequiredAndOptionalRegistryExposed(page);

  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      notesViewSchemaId,
    )}`,
  );
  await expect(
    page.getByTestId(surfaceTabTestId(notesViewSchemaId)),
  ).toBeVisible();
  await expect(
    page.getByTestId(gridShellTestId(notesViewSchemaId)),
  ).toBeVisible();
  await expect(page).toHaveURL(
    new RegExp(`view_schema_id=${encodeURIComponent(notesViewSchemaId)}`),
  );

  await openSystemSurfaceBySwitcher(page, indicatorsViewSchemaId, {
    actionTimeoutMs: 2500,
    attempts: 2,
    totalTimeoutMs: 7000,
  });
  await expect(
    page.getByTestId(gridShellTestId(indicatorsViewSchemaId)),
  ).toBeVisible();
  await expect(page).toHaveURL(
    new RegExp(`view_schema_id=${encodeURIComponent(indicatorsViewSchemaId)}`),
  );

  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      findingsViewSchemaId,
    )}`,
  );
  await expect(
    page.getByTestId(gridShellTestId(findingsViewSchemaId)),
  ).toBeVisible();
  await expect(page).toHaveURL(
    new RegExp(`view_schema_id=${encodeURIComponent(findingsViewSchemaId)}`),
  );
});

test("FE-B-P10-01 Verify Task Requests, Decisions, Parties, Communications Log, Handoff, Status Review, and Lesson open inside the same workbook shell and retain view controls.", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("FEBP1001"),
    "FE-B-P10-01 workbook shell coordination surfaces",
  );
  const savedViewIdsBySurface = new Map<string, string>();
  for (const surface of feP10WorkbookShellSurfaces) {
    const savedView = await createSavedView(page, incidentId, {
      display_name: `FE-B-P10-01 ${surface.label} saved view`,
      scope: "shared",
      view_schema_id: surface.viewSchemaId,
    });
    savedViewIdsBySurface.set(surface.viewSchemaId, savedView.saved_view_id);
  }

  const firstSurface = feP10WorkbookShellSurfaces[0];
  if (firstSurface === undefined) {
    throw new Error("FE-P10 workbook shell surface list must not be empty");
  }
  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      firstSurface.viewSchemaId,
    )}`,
  );

  const shell = page.getByTestId(workbookShellReadyTestId());
  await expect(shell).toBeVisible();
  const shellId = await shell.getAttribute("data-workbook-shell-id");
  expect(shellId).toBe(workbookShellReadyTestId());

  const topBar = shell.locator(
    dataTestIdSelector(workbookShellSlotTestId("top-bar")),
  );
  const viewBar = shell.locator(
    dataTestIdSelector(workbookShellSlotTestId("view-bar")),
  );
  const statusStrip = shell.locator(
    dataTestIdSelector(workbookShellSlotTestId("status-strip")),
  );
  await expect(topBar).toBeVisible();
  await expect(viewBar).toBeVisible();
  await expect(statusStrip).toBeVisible();
  await expect(
    topBar.getByTestId(systemViewSwitcherTriggerTestId()),
  ).toBeVisible();
  await expect(topBar.locator(savedViewFamilySelector())).toHaveCount(0);
  await expect(statusStrip.getByTestId(saveStateTestId())).toBeVisible();

  await topBar.getByTestId(systemViewSwitcherTriggerTestId()).click();
  const menu = page.getByTestId(systemViewSwitcherMenuTestId());
  await expect(menu).toBeVisible();
  for (const surface of feP10WorkbookShellSurfaces) {
    await expect(
      menu.getByTestId(
        systemViewSwitcherOptionTestId(
          surface.groupToken,
          surface.viewSchemaId,
        ),
      ),
    ).toHaveAttribute("data-view-schema-id", surface.viewSchemaId);
  }
  await page.keyboard.press("Escape");
  await expect(menu).toHaveCount(0);

  for (const [index, surface] of feP10WorkbookShellSurfaces.entries()) {
    if (index > 0) {
      const surfaceQuery = page.waitForRequest((request) => {
        return (
          request.method() === "POST" &&
          request
            .url()
            .endsWith(
              `/api/v1/incidents/${incidentId}/views/${surface.viewSchemaId}/query`,
            )
        );
      });
      await topBar.getByTestId(systemViewSwitcherTriggerTestId()).click();
      const option = page.getByTestId(
        systemViewSwitcherOptionTestId(
          surface.groupToken,
          surface.viewSchemaId,
        ),
      );
      await expect(option).toHaveAttribute(
        "data-view-schema-id",
        surface.viewSchemaId,
      );
      await option.click();
      await surfaceQuery;
    }

    await expect(shell).toHaveAttribute(
      "data-workbook-shell-id",
      shellId ?? "",
    );
    await expect(shell).toHaveAttribute(
      "data-active-view-schema-id",
      surface.viewSchemaId,
    );
    await expect(page).toHaveURL(
      new RegExp(`view_schema_id=${encodeURIComponent(surface.viewSchemaId)}`),
    );
    await expect(
      page.getByTestId(gridShellTestId(surface.viewSchemaId)),
    ).toBeVisible();
    await expect(
      viewBar.getByTestId(savedViewSelectorTestId(surface.viewSchemaId)),
    ).toHaveCount(1);
    await expect(
      viewBar.getByTestId(gridFilterFieldTestId(surface.viewSchemaId)),
    ).toBeVisible();
    await expect(
      viewBar.getByTestId(gridGroupingSelectTestId(surface.viewSchemaId)),
    ).toBeVisible();

    const activeSelector = viewBar.getByTestId(
      savedViewSelectorTestId(surface.viewSchemaId),
    );
    const savedViewId = savedViewIdsBySurface.get(surface.viewSchemaId);
    if (savedViewId === undefined) {
      throw new Error(`missing saved view for ${surface.viewSchemaId}`);
    }
    await expect(
      activeSelector.getByTestId(
        savedViewOptionTestId(surface.viewSchemaId, savedViewId),
      ),
    ).toHaveAttribute("data-view-schema-id", surface.viewSchemaId);
    for (const [
      otherViewSchemaId,
      otherSavedViewId,
    ] of savedViewIdsBySurface.entries()) {
      if (otherViewSchemaId === surface.viewSchemaId) {
        continue;
      }
      await expect(
        activeSelector.getByTestId(
          savedViewOptionTestId(surface.viewSchemaId, otherSavedViewId),
        ),
      ).toHaveCount(0);
    }
    await expect(topBar.locator(savedViewFamilySelector())).toHaveCount(0);
    await expect(statusStrip.getByTestId(saveStateTestId())).toBeVisible();
  }
});

type PartyPatchChange = {
  field_key: string;
  value: unknown;
};

type PartyPatchExpectation = {
  changes: PartyPatchChange[];
  clickTestId: string;
  expectedCells: Record<string, unknown>;
  incidentId: string;
  recordId: string;
  viewSchemaId: string;
};

// RCA note: retained run 20260605T002603Z-p3350228 first failed after
// Clear party link at the task requester assertion. The trace showed the
// direct-write PATCH with base_row_version 4 and value null was still pending,
// so party-link assertions wait for request body, response row, and query
// round-trip evidence before reading workbook cells.
async function applyPartyPatchAndWait(
  page: Page,
  expectation: PartyPatchExpectation,
): Promise<ViewApiRow> {
  const isTargetPatchRequest = (request: Request) =>
    request.method() === "PATCH" &&
    request.url().endsWith(`/api/v1/records/${expectation.recordId}`);
  const patchRequests: Request[] = [];
  const capturePatchRequest = (request: Request) => {
    if (isTargetPatchRequest(request)) {
      patchRequests.push(request);
    }
  };
  page.on("request", capturePatchRequest);
  try {
    const [response] = await Promise.all([
      page.waitForResponse((candidate) =>
        isTargetPatchRequest(candidate.request()),
      ),
      page.getByTestId(expectation.clickTestId).click(),
    ]);

    expect(patchRequests).toHaveLength(1);
    expect(patchRequests[0]?.postDataJSON()).toMatchObject({
      view_schema_id: expectation.viewSchemaId,
      changes: expectation.changes,
    });
    expect(response.ok()).toBeTruthy();
    const body = (await response.json()) as { data: { row: ViewApiRow } };
    expectPartyPatchCells(body.data.row, expectation.expectedCells);
    const row = await waitForAuthoritativeViewRowCells(
      page,
      expectation.incidentId,
      expectation.viewSchemaId,
      expectation.recordId,
      expectation.expectedCells,
    );
    await expect(page.getByTestId("generic-mutation-state")).toHaveText(
      "Saved",
    );
    return row;
  } finally {
    page.off("request", capturePatchRequest);
  }
}

async function waitForAuthoritativeViewRowCells(
  page: Page,
  incidentId: string,
  viewSchemaId: string,
  recordId: string,
  expectedCells: Record<string, unknown>,
): Promise<ViewApiRow> {
  let matched: ViewApiRow | null = null;
  await expect
    .poll(
      async () => {
        const rows = await queryViewRows(page, incidentId, viewSchemaId);
        matched = rows.find((row) => row.record_id === recordId) ?? null;
        if (matched === null) {
          return null;
        }
        return Object.fromEntries(
          Object.keys(expectedCells).map((fieldKey) => [
            fieldKey,
            matched?.cells[fieldKey]?.value,
          ]),
        );
      },
      {
        message: `${viewSchemaId} row ${recordId} should round-trip party fields`,
      },
    )
    .toEqual(expectedCells);
  if (matched === null) {
    throw new Error(`missing ${viewSchemaId} row ${recordId}`);
  }
  return matched;
}

function expectPartyPatchCells(
  row: ViewApiRow,
  expectedCells: Record<string, unknown>,
) {
  for (const [fieldKey, expected] of Object.entries(expectedCells)) {
    expect(row.cells[fieldKey]?.value).toEqual(expected);
  }
}

async function fetchPublicViewSchemaIds(page: Page) {
  const response = await page.request.get(`${apiBase}/api/v1/view-schemas`);
  expect(response.ok()).toBeTruthy();
  const body = (await response.json()) as {
    data: { view_schemas: Array<{ view_schema_id: string }> };
  };
  return body.data.view_schemas.map((schema) => schema.view_schema_id);
}

async function expectOptionalStandardizedSurfacesExposed(page: Page) {
  const ids = await fetchPublicViewSchemaIds(page);
  expect(ids).toEqual(
    [...requiredBaseViewSchemaIds, ...optionalStandardizedSurfaceIds].sort(),
  );
  for (const viewSchemaId of optionalStandardizedSurfaceIds) {
    expect(ids).toContain(viewSchemaId);
    const member = await page.request.get(
      `${apiBase}/api/v1/view-schemas/${viewSchemaId}`,
    );
    expect(member.ok()).toBeTruthy();
    const body = (await member.json()) as {
      data: { view_schema_id: string; fields: Array<{ field_key: string }> };
    };
    expect(body.data.view_schema_id).toBe(viewSchemaId);
    expect(body.data.fields.length).toBeGreaterThan(0);
  }
  const hypotheses = await page.request.get(
    `${apiBase}/api/v1/view-schemas/cartulary.view.hypotheses.v1`,
  );
  expect(hypotheses.status()).toBe(404);
}

async function expectRequiredAndOptionalRegistryExposed(page: Page) {
  const ids = await fetchPublicViewSchemaIds(page);
  expect(ids).toEqual(
    [...requiredBaseViewSchemaIds, ...optionalStandardizedSurfaceIds].sort(),
  );
  for (const viewSchemaId of [
    ...requiredBaseViewSchemaIds,
    ...optionalStandardizedSurfaceIds,
  ]) {
    const member = await page.request.get(
      `${apiBase}/api/v1/view-schemas/${viewSchemaId}`,
    );
    expect(member.ok()).toBeTruthy();
    const body = (await member.json()) as {
      data: { view_schema_id: string; fields: Array<{ field_key: string }> };
    };
    expect(body.data.view_schema_id).toBe(viewSchemaId);
    expect(body.data.fields.length).toBeGreaterThan(0);
  }
  const hypotheses = await page.request.get(
    `${apiBase}/api/v1/view-schemas/cartulary.view.hypotheses.v1`,
  );
  expect(hypotheses.status()).toBe(404);
}

async function systemViewSelectorValues(page: Page) {
  await page.getByTestId(systemViewSwitcherTriggerTestId()).click();
  await expect(page.getByTestId(systemViewSwitcherMenuTestId())).toBeVisible();
  const values = await page
    .getByTestId(systemViewSwitcherMenuTestId())
    .locator("[data-view-schema-id]")
    .evaluateAll((options) =>
      options
        .map((option) => option.getAttribute("data-view-schema-id"))
        .filter((value): value is string => value !== null),
    );
  await page.keyboard.press("Escape");
  return values;
}

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
  void heading;
  await expect(page.getByTestId(gridShellTestId(viewSchemaId))).toBeVisible();
  await page.getByTestId(workbookInspectorToggleTestId(viewSchemaId)).click();
}

async function setPhase9GenericCreateField(
  page: Page,
  fieldKey: string,
  value: string | string[],
) {
  const input = page.getByTestId(genericCreateFieldTestId(fieldKey));
  const tagName = await input.evaluate((element) => element.tagName);
  const inputType = await input.getAttribute("type");
  if (inputType === "checkbox") {
    if (value === "true") {
      await input.check();
    } else {
      await input.uncheck();
    }
    return;
  }
  if (tagName === "SELECT") {
    await input.selectOption(value);
    return;
  }
  await input.fill(Array.isArray(value) ? value.join("\n") : value);
}

async function editPhase9GenericCell(
  page: Page,
  viewSchemaId: string,
  recordId: string,
  fieldKey: string,
  value: string | string[],
) {
  await page.getByTestId(workbookInspectorToggleTestId(viewSchemaId)).click();
  await page
    .getByTestId(genericEditRecordSelectTestId(viewSchemaId))
    .selectOption(recordId);
  await page
    .getByTestId(genericEditFieldSelectTestId(viewSchemaId))
    .selectOption(fieldKey);
  const input = page.getByTestId(genericEditValueTestId(viewSchemaId));
  const tagName = await input.evaluate((element) => element.tagName);
  const inputType = await input.getAttribute("type");
  if (inputType === "checkbox") {
    if (value === "true") {
      await input.check();
    } else {
      await input.uncheck();
    }
    await submitGenericEditAndWait(page, viewSchemaId, recordId);
    return;
  }
  if (tagName === "SELECT") {
    if (typeof value === "string") {
      await waitForPhase9GenericOption(
        page,
        genericEditValueTestId(viewSchemaId),
        value,
      );
    }
    await input.selectOption(value);
  } else {
    await input.fill(Array.isArray(value) ? value.join("\n") : value);
  }
  await submitGenericEditAndWait(page, viewSchemaId, recordId);
}

async function waitForPhase9GenericOption(
  page: Page,
  testId: string,
  value: string,
) {
  await expect(
    page.getByTestId(testId).locator(`option[value="${value}"]`),
  ).toHaveCount(1, { timeout: 15_000 });
}

async function setGenericGridScroll(page: Page, viewSchemaId: string) {
  return page
    .locator(
      `${dataTestIdSelector(gridShellTestId(viewSchemaId))} ${gridScrollportSelector()}`,
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
          `${dataTestIdSelector(gridShellTestId(viewSchemaId))} ${gridScrollportSelector()}`,
        )
        .evaluate((element) => ({
          top: element.scrollTop,
          left: element.scrollLeft,
        })),
    )
    .toEqual(expected);
}
