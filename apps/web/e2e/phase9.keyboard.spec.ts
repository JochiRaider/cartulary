import { rowCellTestId } from "@cartulary/ui-contracts";

import { expect, test } from "./fixtures";
import {
  createIncident,
  createViewRow,
  uniqueIncidentKey,
  uniqueTxn,
} from "./helpers";
import {
  assessmentsViewSchemaId,
  collectionActionsPayload,
  hostRefsFieldKey,
  hostsViewSchemaId,
  notesViewSchemaId,
  taskRequestsViewSchemaId,
} from "./phase4Helpers";

const timelineViewSchemaId = "cartulary.view.timeline.v1";

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
  await expect(page.getByText("Timeline mutation substrate")).toBeVisible();
  await expect(
    page.getByTestId(rowCellTestId(alpha.record_id as string, "summary")),
  ).toHaveValue("Phase 9 alpha");
  const initialURL = page.url();

  const alphaSummary = page.getByTestId(
    rowCellTestId(alpha.record_id as string, "summary"),
  );
  await alphaSummary.focus();
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `timeline:${alpha.record_id}:timeline.summary`,
  );

  await alphaSummary.press("ArrowDown");
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `timeline:${beta.record_id}:timeline.summary`,
  );
  await expect(
    page.getByTestId(rowCellTestId(beta.record_id as string, "summary")),
  ).toBeFocused();

  await page.keyboard.press("ArrowUp");
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `timeline:${alpha.record_id}:timeline.summary`,
  );
  await expect(alphaSummary).toBeFocused();

  await page.keyboard.press("Enter");
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `timeline:${beta.record_id}:timeline.summary`,
  );
  await expect(
    page.getByTestId(rowCellTestId(beta.record_id as string, "summary")),
  ).toBeFocused();

  await page.keyboard.press("Shift+Enter");
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `timeline:${alpha.record_id}:timeline.summary`,
  );
  await expect(alphaSummary).toBeFocused();

  await page.keyboard.press("Tab");
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `timeline:${alpha.record_id}:timeline.host_refs`,
  );
  await expect(
    page.getByTestId(
      rowCellTestId(alpha.record_id as string, "hostRefs-input"),
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
    `timeline:${alpha.record_id}:timeline.host_refs`,
  );

  await page.keyboard.press("Alt+H");
  await expect(page.getByTestId("row-history-panel")).toContainText(
    String(alpha.record_id),
  );
  expect(page.url()).toBe(initialURL);

  await page.keyboard.press("Control+V");
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `timeline:${alpha.record_id}:timeline.host_refs`,
  );
  await page.keyboard.press("Escape");
  await expect(page.getByTestId("row-history-panel")).toHaveCount(0);
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `timeline:${alpha.record_id}:timeline.host_refs`,
  );
  expect(page.url()).toBe(initialURL);
});

test("Phase 9 E-9-GRID-01 shared grid keyboard anchors stay stable across workbook cells", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E9GRID01"),
    "Phase 9 E-9-GRID-01 keyboard anchor semantics",
  );
  const row = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("e9grid01"),
    "timeline.summary": "Phase 9 grid anchor",
  });

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByText("Timeline mutation substrate")).toBeVisible();

  const summary = page.getByTestId(
    rowCellTestId(row.record_id as string, "summary"),
  );
  await summary.focus();
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `timeline:${row.record_id}:timeline.summary`,
  );

  await summary.press("ArrowRight");
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `timeline:${row.record_id}:timeline.host_refs`,
  );
  await expect(
    page.getByTestId(rowCellTestId(row.record_id as string, "hostRefs-input")),
  ).toBeFocused();

  await page.keyboard.press("ArrowLeft");
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `timeline:${row.record_id}:timeline.summary`,
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
    `hosts:${host.record_id}:host.display_name`,
  );
  await page.keyboard.press("ArrowRight");
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `hosts:${host.record_id}:host.hostname`,
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
});
