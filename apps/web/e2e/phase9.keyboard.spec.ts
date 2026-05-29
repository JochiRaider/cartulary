import {
  rowCellTestId,
  rowHistoryPanelTestId,
  timelineCollectionInputTestId,
  timelineMutationSubstrateReadyTestId,
} from "@cartulary/ui-contracts";

import { expect, test } from "./fixtures";
import {
  createIncident,
  createViewRow,
  queryViewRows,
  uniqueIncidentKey,
  uniqueTxn,
} from "./helpers";
import {
  assessmentsViewSchemaId,
  collectionActionsPayload,
  commLogViewSchemaId,
  decisionsViewSchemaId,
  handoffViewSchemaId,
  hostRefsFieldKey,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  indicatorsViewSchemaId,
  lessonViewSchemaId,
  notesViewSchemaId,
  statusReviewViewSchemaId,
  taskRequestsViewSchemaId,
} from "./phase4Helpers";

const timelineViewSchemaId = "cartulary.view.timeline.v1";

function stringCell(
  row: { readonly cells?: Record<string, { readonly value?: unknown }> },
  fieldKey: string,
): string {
  return String(row.cells?.[fieldKey]?.value ?? "");
}

test("Phase 9 E-9-01 keyboard shortcuts keep workbook grid anchors without module switching", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E901"),
    "Phase 9 E-9-01 keyboard contract",
  );
  const alpha = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("e901-alpha"),
    "timeline.summary": "Phase 9 alpha",
    [hostRefsFieldKey]: collectionActionsPayload(["Phase9Host?"]),
  });
  const beta = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("e901-beta"),
    "timeline.summary": "Phase 9 beta",
  });

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toBeVisible();
  await expect(
    page.getByTestId(
      rowCellTestId(alpha.record_id as string, "timeline.summary"),
    ),
  ).toHaveValue("Phase 9 alpha");
  const initialURL = page.url();

  const alphaSummary = page.getByTestId(
    rowCellTestId(alpha.record_id as string, "timeline.summary"),
  );
  await alphaSummary.focus();
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${timelineViewSchemaId}:${alpha.record_id}:timeline.summary`,
  );

  await alphaSummary.press("ArrowDown");
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${timelineViewSchemaId}:${beta.record_id}:timeline.summary`,
  );
  await expect(
    page.getByTestId(
      rowCellTestId(beta.record_id as string, "timeline.summary"),
    ),
  ).toBeFocused();

  await page.keyboard.press("ArrowUp");
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${timelineViewSchemaId}:${alpha.record_id}:timeline.summary`,
  );
  await expect(alphaSummary).toBeFocused();

  await page.keyboard.press("Enter");
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${timelineViewSchemaId}:${beta.record_id}:timeline.summary`,
  );
  await expect(
    page.getByTestId(
      rowCellTestId(beta.record_id as string, "timeline.summary"),
    ),
  ).toBeFocused();

  await page.keyboard.press("Shift+Enter");
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${timelineViewSchemaId}:${alpha.record_id}:timeline.summary`,
  );
  await expect(alphaSummary).toBeFocused();

  await page.keyboard.press("Tab");
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${timelineViewSchemaId}:${alpha.record_id}:timeline.host_refs`,
  );
  await expect(
    page.getByTestId(
      timelineCollectionInputTestId(
        alpha.record_id as string,
        hostRefsFieldKey,
      ),
    ),
  ).toBeFocused();

  await page.keyboard.press("Control+K");
  await expect(page.getByTestId("timeline-inspector")).toContainText(
    "Phase9Host?",
  );
  expect(page.url()).toBe(initialURL);

  await page.keyboard.press("Space");
  await expect(page.getByTestId("timeline-inspector-message")).toContainText(
    "Linked evidence preview",
  );
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${timelineViewSchemaId}:${alpha.record_id}:timeline.host_refs`,
  );

  await page.keyboard.press("Alt+H");
  await expect(page.getByTestId(rowHistoryPanelTestId())).toContainText(
    String(alpha.record_id),
  );
  expect(page.url()).toBe(initialURL);

  await page.keyboard.press("Control+V");
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${timelineViewSchemaId}:${alpha.record_id}:timeline.host_refs`,
  );
  await page.keyboard.press("Escape");
  await expect(page.getByTestId(rowHistoryPanelTestId())).toHaveCount(0);
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${timelineViewSchemaId}:${alpha.record_id}:timeline.host_refs`,
  );
  expect(page.url()).toBe(initialURL);
});

test("Phase 9 E-9-GRIDANCHORS-01 shared grid keyboard anchors stay stable across workbook cells", async ({
  page,
  workerAdmin,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E9GRID01"),
    "Phase 9 E-9-GRIDANCHORS-01 keyboard anchor semantics",
  );
  const row = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("e9grid01"),
    "timeline.summary": "Phase 9 grid anchor",
  });

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toBeVisible();

  const summary = page.getByTestId(
    rowCellTestId(row.record_id as string, "timeline.summary"),
  );
  await summary.focus();
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${timelineViewSchemaId}:${row.record_id}:timeline.summary`,
  );

  await summary.press("ArrowRight");
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${timelineViewSchemaId}:${row.record_id}:timeline.host_refs`,
  );
  await expect(
    page.getByTestId(
      timelineCollectionInputTestId(row.record_id as string, hostRefsFieldKey),
    ),
  ).toBeFocused();

  await page.keyboard.press("ArrowLeft");
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${timelineViewSchemaId}:${row.record_id}:timeline.summary`,
  );
  await expect(summary).toBeFocused();

  const host = await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("e9grid01-host"),
    "host.display_name": "Phase 9 host anchor",
    "host.hostname": "phase9-host.example.test",
  });
  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      hostsViewSchemaId,
    )}`,
  );
  const hostName = page.getByTestId(
    rowCellTestId(host.record_id as string, "host.display_name"),
  );
  await expect(hostName).toHaveText("Phase 9 host anchor");
  await hostName.focus();
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${hostsViewSchemaId}:${host.record_id}:host.display_name`,
  );
  await page.keyboard.press("ArrowRight");
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${hostsViewSchemaId}:${host.record_id}:host.hostname`,
  );

  const identity = await createViewRow(
    page,
    incidentId,
    identitiesViewSchemaId,
    {
      client_txn_id: uniqueTxn("e9grid01-identity"),
      "identity.display_name": "Phase 9 identity anchor",
      "identity.upn": "phase9.identity@example.test",
    },
  );
  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      identitiesViewSchemaId,
    )}`,
  );
  const identityName = page.getByTestId(
    rowCellTestId(identity.record_id as string, "identity.display_name"),
  );
  await expect(identityName).toHaveText("Phase 9 identity anchor");
  await identityName.focus();
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${identitiesViewSchemaId}:${identity.record_id}:identity.display_name`,
  );
  await page.keyboard.press("ArrowRight");
  await expect(page.getByTestId("workbook-focus-anchor")).toContainText(
    `${identitiesViewSchemaId}:${identity.record_id}:`,
  );

  const assessment = await createViewRow(
    page,
    incidentId,
    assessmentsViewSchemaId,
    {
      client_txn_id: uniqueTxn("e9grid01-assessment"),
      "assessment.subject_ref": host.record_id,
      "assessment.subject_type": "host",
      "assessment.assessment_state": "confirmed",
      "assessment.confidence_score": 55,
      "assessment.rationale": "Phase 9 assessment anchor",
    },
  );
  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      assessmentsViewSchemaId,
    )}`,
  );
  const assessmentState = page.getByTestId(
    rowCellTestId(
      assessment.record_id as string,
      "assessment.assessment_state",
    ),
  );
  await expect(assessmentState).toHaveText("confirmed");
  await assessmentState.focus();
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${assessmentsViewSchemaId}:${assessment.record_id}:assessment.assessment_state`,
  );
  await page.keyboard.press("ArrowRight");
  await expect(page.getByTestId("workbook-focus-anchor")).toContainText(
    `${assessmentsViewSchemaId}:${assessment.record_id}:`,
  );

  const task = await createViewRow(page, incidentId, taskRequestsViewSchemaId, {
    client_txn_id: uniqueTxn("e9grid01-task"),
    "task.title": "Phase 9 task request anchor",
    "task.task_kind": "collection",
  });
  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      taskRequestsViewSchemaId,
    )}`,
  );
  const taskTitle = page.getByTestId(
    rowCellTestId(task.record_id as string, "task.title"),
  );
  await expect(taskTitle).toHaveText("Phase 9 task request anchor");
  await taskTitle.focus();
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${taskRequestsViewSchemaId}:${task.record_id}:task.title`,
  );
  await page.keyboard.press("ArrowRight");
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${taskRequestsViewSchemaId}:${task.record_id}:task.status`,
  );
  await expect(
    page.getByTestId(rowCellTestId(task.record_id as string, "task.status")),
  ).toBeFocused();

  const decision = await createViewRow(
    page,
    incidentId,
    decisionsViewSchemaId,
    {
      client_txn_id: uniqueTxn("e9grid01-decision"),
      "decision.summary": "Phase 9 decision anchor",
      "decision.decision_type": "containment",
      "decision.rationale": "Phase 9 decision grid anchor rationale",
    },
  );
  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      decisionsViewSchemaId,
    )}`,
  );
  const decisionSummary = page.getByTestId(
    rowCellTestId(decision.record_id as string, "decision.summary"),
  );
  await expect(decisionSummary).toHaveText("Phase 9 decision anchor");
  await decisionSummary.focus();
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${decisionsViewSchemaId}:${decision.record_id}:decision.summary`,
  );
  await page.keyboard.press("ArrowRight");
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${decisionsViewSchemaId}:${decision.record_id}:decision.status`,
  );
  await expect(
    page.getByTestId(
      rowCellTestId(decision.record_id as string, "decision.status"),
    ),
  ).toBeFocused();

  const note = await createViewRow(page, incidentId, notesViewSchemaId, {
    client_txn_id: uniqueTxn("e9grid01-note"),
    "note.title": "Phase 9 note anchor",
    "note.body": "Phase 9 generic surface anchor body",
  });
  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      notesViewSchemaId,
    )}`,
  );
  const noteTitle = page.getByTestId(
    rowCellTestId(note.record_id as string, "note.title"),
  );
  await expect(noteTitle).toHaveText("Phase 9 note anchor");
  await noteTitle.focus();
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${notesViewSchemaId}:${note.record_id}:note.title`,
  );
  await page.keyboard.press("ArrowRight");
  await expect(page.getByTestId("workbook-focus-anchor")).toContainText(
    `${notesViewSchemaId}:${note.record_id}:`,
  );

  const comm = await createViewRow(page, incidentId, commLogViewSchemaId, {
    client_txn_id: uniqueTxn("e9grid01-comm"),
    "comm_log.comm_type": "briefing",
    "comm_log.audience": "Phase 9 grid audience",
    "comm_log.channel_or_meeting": "Grid bridge",
    "comm_log.summary": "Phase 9 comm log anchor",
  });
  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      commLogViewSchemaId,
    )}`,
  );
  const commSummary = page.getByTestId(
    rowCellTestId(comm.record_id as string, "comm_log.summary"),
  );
  await expect(commSummary).toHaveText("Phase 9 comm log anchor");
  await commSummary.focus();
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${commLogViewSchemaId}:${comm.record_id}:comm_log.summary`,
  );
  await page.keyboard.press("ArrowRight");
  await expect(page.getByTestId("workbook-focus-anchor")).toContainText(
    `${commLogViewSchemaId}:${comm.record_id}:`,
  );
  await expect(page.getByTestId("workbook-focus-anchor")).not.toHaveText(
    `${commLogViewSchemaId}:${comm.record_id}:comm_log.summary`,
  );

  const handoff = await createViewRow(page, incidentId, handoffViewSchemaId, {
    client_txn_id: uniqueTxn("e9grid01-handoff"),
    "handoff.incoming_owner_user_id": workerAdmin.user_id,
    "handoff.current_state_summary": "Phase 9 handoff anchor",
  });
  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      handoffViewSchemaId,
    )}`,
  );
  const handoffSummary = page.getByTestId(
    rowCellTestId(handoff.record_id as string, "handoff.current_state_summary"),
  );
  await expect(handoffSummary).toHaveText("Phase 9 handoff anchor");
  await handoffSummary.focus();
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${handoffViewSchemaId}:${handoff.record_id}:handoff.current_state_summary`,
  );
  await page.keyboard.press("ArrowRight");
  await expect(page.getByTestId("workbook-focus-anchor")).toContainText(
    `${handoffViewSchemaId}:${handoff.record_id}:`,
  );
  await expect(page.getByTestId("workbook-focus-anchor")).not.toHaveText(
    `${handoffViewSchemaId}:${handoff.record_id}:handoff.current_state_summary`,
  );

  const statusReview = await createViewRow(
    page,
    incidentId,
    statusReviewViewSchemaId,
    {
      client_txn_id: uniqueTxn("e9grid01-status"),
      "status_review.current_state_summary": "Phase 9 status review anchor",
    },
  );
  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      statusReviewViewSchemaId,
    )}`,
  );
  const statusSummary = page.getByTestId(
    rowCellTestId(
      statusReview.record_id as string,
      "status_review.current_state_summary",
    ),
  );
  await expect(statusSummary).toHaveText("Phase 9 status review anchor");
  await statusSummary.focus();
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${statusReviewViewSchemaId}:${statusReview.record_id}:status_review.current_state_summary`,
  );
  await page.keyboard.press("ArrowRight");
  await expect(page.getByTestId("workbook-focus-anchor")).toContainText(
    `${statusReviewViewSchemaId}:${statusReview.record_id}:`,
  );
  await expect(page.getByTestId("workbook-focus-anchor")).not.toHaveText(
    `${statusReviewViewSchemaId}:${statusReview.record_id}:status_review.current_state_summary`,
  );

  const lesson = await createViewRow(page, incidentId, lessonViewSchemaId, {
    client_txn_id: uniqueTxn("e9grid01-lesson"),
    "lesson.summary": "Phase 9 lesson anchor",
  });
  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      lessonViewSchemaId,
    )}`,
  );
  const lessonSummary = page.getByTestId(
    rowCellTestId(lesson.record_id as string, "lesson.summary"),
  );
  await expect(lessonSummary).toHaveText("Phase 9 lesson anchor");
  await lessonSummary.focus();
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${lessonViewSchemaId}:${lesson.record_id}:lesson.summary`,
  );
  await page.keyboard.press("ArrowRight");
  await expect(page.getByTestId("workbook-focus-anchor")).toContainText(
    `${lessonViewSchemaId}:${lesson.record_id}:`,
  );
  await expect(page.getByTestId("workbook-focus-anchor")).not.toHaveText(
    `${lessonViewSchemaId}:${lesson.record_id}:lesson.summary`,
  );

  const indicator = await createViewRow(
    page,
    incidentId,
    indicatorsViewSchemaId,
    {
      client_txn_id: uniqueTxn("e9grid01-indicator"),
      "indicator.indicator_type": "ipv4_addr",
      "indicator.value_kind": "atomic",
      "indicator.display_value": "203.0.113.91",
    },
  );
  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      indicatorsViewSchemaId,
    )}`,
  );
  const indicatorType = page.getByTestId(
    rowCellTestId(indicator.record_id as string, "indicator.indicator_type"),
  );
  await expect(indicatorType).toHaveText("ipv4_addr");
  await indicatorType.focus();
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${indicatorsViewSchemaId}:${indicator.record_id}:indicator.indicator_type`,
  );
  await page.keyboard.press("ArrowRight");
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${indicatorsViewSchemaId}:${indicator.record_id}:indicator.value_kind`,
  );
});

test("Phase 9 E-9-GRIDHOST-01 Host entity-origin clipboard paste reuses exact matches and creates stubs", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E9GRIDHOSTPASTE"),
    "Phase 9 E-9-GRIDHOST-01 host paste",
  );
  const existing = await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("e9grid-host-existing"),
    "host.display_name": "Phase 9 reusable host",
    "host.hostname": "phase9-host-reuse.example.test",
  });
  const postURLs: string[] = [];
  page.on("request", (request) => {
    if (request.method() === "POST") {
      postURLs.push(request.url());
    }
  });

  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      hostsViewSchemaId,
    )}`,
  );
  const displayName = page.getByTestId(
    rowCellTestId(existing.record_id as string, "host.display_name"),
  );
  await expect(displayName).toHaveText("Phase 9 reusable host");
  await displayName.focus();
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${hostsViewSchemaId}:${existing.record_id}:host.display_name`,
  );

  const pasteResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response
        .url()
        .includes(
          `/api/v1/incidents/${incidentId}/views/${hostsViewSchemaId}/clipboard-paste`,
        ),
  );
  await displayName.evaluate((element) => {
    const data = new DataTransfer();
    data.setData(
      "text/plain",
      [
        "Phase 9 pasted host reuse\tphase9-host-reuse.example.test",
        "Phase 9 pasted host create\tphase9-host-create.example.test",
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
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${hostsViewSchemaId}:${existing.record_id}:host.display_name`,
  );
  await expect(
    page.getByTestId(
      rowCellTestId(existing.record_id as string, "host.display_name"),
    ),
  ).toHaveText("Phase 9 pasted host reuse");

  const rows = await queryViewRows(page, incidentId, hostsViewSchemaId);
  const reused = rows.find((row) => row.record_id === existing.record_id);
  expect(stringCell(reused ?? {}, "host.display_name")).toBe(
    "Phase 9 pasted host reuse",
  );
  const created = rows.find(
    (row) =>
      row.record_id !== existing.record_id &&
      stringCell(row, "host.hostname") === "phase9-host-create.example.test",
  );
  expect(created).toBeTruthy();
  if (created) {
    await expect(
      page.getByTestId(rowCellTestId(created.record_id, "host.display_name")),
    ).toHaveText("Phase 9 pasted host create");
  }
  expect(postURLs.some((url) => url.includes("/imports"))).toBeFalsy();
});

test("Phase 9 E-9-GRIDIDENTITY-01 Identity entity-origin clipboard paste reuses exact matches and creates stubs", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E9GRIDIDENTITYPASTE"),
    "Phase 9 E-9-GRIDIDENTITY-01 identity paste",
  );
  const existing = await createViewRow(
    page,
    incidentId,
    identitiesViewSchemaId,
    {
      client_txn_id: uniqueTxn("e9grid-identity-existing"),
      "identity.display_name": "Phase 9 reusable identity",
      "identity.upn": "phase9.identity.reuse@example.test",
    },
  );
  const postURLs: string[] = [];
  page.on("request", (request) => {
    if (request.method() === "POST") {
      postURLs.push(request.url());
    }
  });

  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      identitiesViewSchemaId,
    )}`,
  );
  const displayName = page.getByTestId(
    rowCellTestId(existing.record_id as string, "identity.display_name"),
  );
  await expect(displayName).toHaveText("Phase 9 reusable identity");
  await displayName.focus();
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${identitiesViewSchemaId}:${existing.record_id}:identity.display_name`,
  );

  const pasteResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response
        .url()
        .includes(
          `/api/v1/incidents/${incidentId}/views/${identitiesViewSchemaId}/clipboard-paste`,
        ),
  );
  await displayName.evaluate((element) => {
    const data = new DataTransfer();
    data.setData(
      "text/plain",
      [
        "Phase 9 pasted identity reuse\tphase9.identity.reuse@example.test",
        "Phase 9 pasted identity create\tphase9.identity.create@example.test",
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
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${identitiesViewSchemaId}:${existing.record_id}:identity.display_name`,
  );
  await expect(
    page.getByTestId(
      rowCellTestId(existing.record_id as string, "identity.display_name"),
    ),
  ).toHaveText("Phase 9 pasted identity reuse");

  const rows = await queryViewRows(page, incidentId, identitiesViewSchemaId);
  const reused = rows.find((row) => row.record_id === existing.record_id);
  expect(stringCell(reused ?? {}, "identity.display_name")).toBe(
    "Phase 9 pasted identity reuse",
  );
  const created = rows.find(
    (row) =>
      row.record_id !== existing.record_id &&
      stringCell(row, "identity.upn") === "phase9.identity.create@example.test",
  );
  expect(created).toBeTruthy();
  if (created) {
    await expect(
      page.getByTestId(
        rowCellTestId(created.record_id, "identity.display_name"),
      ),
    ).toHaveText("Phase 9 pasted identity create");
  }
  expect(postURLs.some((url) => url.includes("/imports"))).toBeFalsy();
});
