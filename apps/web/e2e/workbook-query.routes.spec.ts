import type { ViewRow } from "@cartulary/protocol-ts/http";
import { scrollGridCellIntoView } from "@cartulary/test-utils/grid";
import {
  gridRowTestId,
  gridSavedRowsSelector,
  gridShellTestId,
  rowCellTestId,
  workbookInspectorCloseButtonTestId,
} from "@cartulary/ui-contracts";
import {
  assessmentsViewSchemaId,
  commLogViewSchemaId,
  decisionsViewSchemaId,
  evidenceViewSchemaId,
  findingsViewSchemaId,
  forensicKeywordsViewSchemaId,
  handoffViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  indicatorsViewSchemaId,
  investigativeQueriesViewSchemaId,
  lessonViewSchemaId,
  notesViewSchemaId,
  partiesViewSchemaId,
  statusReviewViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { expect, test } from "./fixtures";

import { createIncident } from "./support/incidents/fixtures";

import { uniqueIncidentKey } from "./support/runtime/fixtureIdentity";
import { rejectUnexpectedPageErrors } from "./support/runtime/observations";
import { installIncidentSocketMonitor } from "./support/transport/incidentSocket";
import { observeQuery, seed } from "./support/workbook/queryBrowsing";
import {
  openGenericInspectorForRecord,
  openTimelineInspector,
} from "./support/workbook/rowMutations";

async function exerciseContinuation(page: Page, view: string, actorId: string) {
  test.setTimeout(180_000);
  rejectUnexpectedPageErrors(page);
  const count = 101;
  const incident = await createIncident(
    page,
    uniqueIncidentKey("WQC"),
    `Workbook continuation ${view}`,
  );
  const url = `/?incident_id=${incident}&view_schema_id=${encodeURIComponent(view)}`;
  const fixtureSocket = installIncidentSocketMonitor(page, incident);
  await page.goto(url);
  await fixtureSocket.waitForAcceptedSocket();
  // A committed fixture write may still await collaboration sequencing. Observe
  // each publication before measuring a fresh page's bounded startup reads.
  const fixture = await seed(page, incident, view, count, actorId, (recordId) =>
    fixtureSocket.waitForMessage("record_changed", {
      matches: (message) => message.payload.record_id === recordId,
    }),
  );
  await page.goto("about:blank");
  const reads = await observeQuery(page, incident, view);
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${encodeURIComponent(view)}`,
  );
  const controls = page.getByRole("group", { name: "Workbook browsing" });
  await expect(controls).toContainText("100 records loaded; more available.");
  expect(reads.length).toBeLessThanOrEqual(3);
  const first = reads.at(-1);
  if (!first) throw new Error("Missing first production response");
  expect(first.request.cursor_token).toBeUndefined();
  expect(first.request.limit).toBe(100);
  if (view === timelineViewSchemaId) {
    await page
      .getByRole("checkbox", { name: "Select all loaded records", exact: true })
      .check();
    await expect(
      page.getByText("100 records selected.", { exact: true }),
    ).toBeVisible();
  }
  const more = controls.getByRole("button", { name: "Load more", exact: true });
  await more.focus();
  await more.press("Enter");
  await expect(controls).toContainText(
    `${Math.min(200, count)} records loaded`,
  );
  await expect(more).toBeFocused();
  if (view === timelineViewSchemaId)
    await expect(
      page.getByText("100 records selected.", { exact: true }),
    ).toBeVisible();
  const later = reads.at(-1);
  if (!later) throw new Error("Missing continuation response");
  const producing = reads
    .slice(0, reads.indexOf(later))
    .reverse()
    .find(
      (read) =>
        read.response.meta.paging.next_cursor === later.request.cursor_token,
    );
  expect(
    producing,
    "Continuation preserves a producing response's cursor bytes",
  ).toBeDefined();
  expect(later.response.meta.query).toEqual(producing?.response.meta.query);
  expect(later.request.sort).toBeUndefined();
  expect(later.request.group_by).toBeUndefined();
  const row = later.response.data.rows[0];
  if (!row) throw new Error("Fixture must cross the route page boundary");
  await exerciseRow(
    page,
    view,
    fixture.field,
    row,
    String(row.cells[fixture.field]?.value),
  );
  await expect(controls).toContainText(
    "101 records loaded; end of current results.",
  );
  // Inspector departure releases the exhausted continuation action's focus.
  await expect(more).toHaveCount(0);
  await expect(
    page.getByTestId(gridShellTestId(view)).locator(gridSavedRowsSelector()),
  ).not.toHaveCount(0);
  expect(
    await page
      .getByTestId(gridShellTestId(view))
      .locator(gridSavedRowsSelector())
      .count(),
  ).toBeLessThan(100);
  // One initial read, explicit continuation, and bounded edit reconciliation.
  expect(reads.length).toBeLessThanOrEqual(12);
}

async function exerciseRow(
  page: Page,
  view: string,
  field: string,
  row: ViewRow,
  value: string,
) {
  await scrollGridCellIntoView({
    page,
    surface: view,
    recordId: row.record_id,
    cellKey: field,
  });
  await expect(
    page.getByTestId(rowCellTestId(row.record_id, field)),
  ).toContainText(value);
  if (view === timelineViewSchemaId)
    await openTimelineInspector(page, row.record_id);
  else await openGenericInspectorForRecord(page, view, row.record_id);
  await expect(
    page.getByTestId(workbookInspectorCloseButtonTestId(view)),
  ).toBeVisible();
  await page.getByTestId(workbookInspectorCloseButtonTestId(view)).click();
  // Assessment rows are append-only; Indicator display values have their own lifecycle editor.
  if (view !== assessmentsViewSchemaId && view !== indicatorsViewSchemaId) {
    await scrollGridCellIntoView({
      page,
      surface: view,
      recordId: row.record_id,
      cellKey: field,
    });
    const cell = page
      .getByTestId(rowCellTestId(row.record_id, field))
      .locator('xpath=ancestor-or-self::*[@role="gridcell"][1]');
    await cell.click();
    const editor = page
      .getByTestId(gridRowTestId(view, row.record_id))
      .getByRole("textbox")
      .first();
    await expect(editor).toBeVisible();
    const accepted = page.waitForResponse(
      (response) =>
        response.request().method() === "PATCH" &&
        response.url().endsWith(`/records/${row.record_id}`),
    );
    const editedValue = `${value} edited`;
    await editor.fill(editedValue);
    await editor.press("Enter");
    expect((await accepted).ok()).toBe(true);
    await expect(
      page.getByTestId(gridRowTestId(view, row.record_id)),
    ).toHaveAttribute("data-grid-row-version", "2");
    await expect(
      page.getByTestId(rowCellTestId(row.record_id, field)),
    ).toContainText(editedValue);
  }
}

async function exercisePublicSurface(
  page: Page,
  view: string,
  actorId: string,
) {
  rejectUnexpectedPageErrors(page);
  const incident = await createIncident(
    page,
    uniqueIncidentKey("WQS"),
    `Workbook surface ${view}`,
  );
  const fixture = await seed(page, incident, view, 2, actorId);
  const reads = await observeQuery(page, incident, view);
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${encodeURIComponent(view)}`,
  );
  await expect(
    page.getByRole("group", { name: "Workbook browsing" }),
  ).toContainText("2 records loaded; end of current results.");
  const read = reads.at(-1);
  if (!read) throw new Error("Missing public surface query response");
  expect(read.response.data.incident_id).toBe(incident);
  expect(read.response.data.view_schema_id).toBe(view);
  const rows = read.response.data.rows;
  expect(rows).toHaveLength(2);
  expect(new Set(rows.map((row) => row.record_id)).size).toBe(2);
  // Authored fixture values establish rendering identity independently of the
  // query response. Default ordering is route-owned, not creation ordering.
  const values =
    view === indicatorsViewSchemaId
      ? ["wqc-00000.example.test", "wqc-00001.example.test"]
      : ["WQC 00000", "WQC 00001"];
  expect(rows.map((row) => row.cells[fixture.field]?.value).sort()).toEqual(
    values,
  );
  for (const row of rows) {
    expect(row.row_version).toBe(1);
    await scrollGridCellIntoView({
      page,
      surface: view,
      recordId: row.record_id,
      cellKey: fixture.field,
    });
    await expect(
      page.getByTestId(rowCellTestId(row.record_id, fixture.field)),
    ).toContainText(String(row.cells[fixture.field]?.value));
  }
  const row = rows[1];
  if (!row) throw new Error("Missing second surface fixture record");
  await exerciseRow(
    page,
    view,
    fixture.field,
    row,
    String(row.cells[fixture.field]?.value),
  );
}

test("Workbook continuation reaches and returns later cartulary.view.timeline.v2 records through the real route", async ({
  page,
  testActor,
}) => {
  await exerciseContinuation(page, timelineViewSchemaId, testActor.user_id);
});

test("Workbook continuation reaches and returns later cartulary.view.hosts.v1 records through the real route", async ({
  page,
  testActor,
}) => {
  await exerciseContinuation(page, hostsViewSchemaId, testActor.user_id);
});

test("Workbook continuation reaches and returns later cartulary.view.identities.v1 records through the real route", async ({
  page,
  testActor,
}) => {
  await exerciseContinuation(page, identitiesViewSchemaId, testActor.user_id);
});

test("Workbook continuation reaches and returns later cartulary.view.notes.v1 records through the real route", async ({
  page,
  testActor,
}) => {
  await exerciseContinuation(page, notesViewSchemaId, testActor.user_id);
});

test("Workbook surface cartulary.view.evidence.v1 loads and inspects records through the real route", async ({
  page,
  testActor,
}) => {
  await exercisePublicSurface(page, evidenceViewSchemaId, testActor.user_id);
});

test("Workbook surface cartulary.view.indicators.v1 loads and inspects records through the real route", async ({
  page,
  testActor,
}) => {
  await exercisePublicSurface(page, indicatorsViewSchemaId, testActor.user_id);
});

test("Workbook surface cartulary.view.parties.v1 loads and inspects records through the real route", async ({
  page,
  testActor,
}) => {
  await exercisePublicSurface(page, partiesViewSchemaId, testActor.user_id);
});

test("Workbook surface cartulary.view.task_requests.v1 loads and inspects records through the real route", async ({
  page,
  testActor,
}) => {
  await exercisePublicSurface(
    page,
    taskRequestsViewSchemaId,
    testActor.user_id,
  );
});

test("Workbook surface cartulary.view.decisions.v1 loads and inspects records through the real route", async ({
  page,
  testActor,
}) => {
  await exercisePublicSurface(page, decisionsViewSchemaId, testActor.user_id);
});

test("Workbook surface cartulary.view.comm_log.v1 loads and inspects records through the real route", async ({
  page,
  testActor,
}) => {
  await exercisePublicSurface(page, commLogViewSchemaId, testActor.user_id);
});

test("Workbook surface cartulary.view.handoff.v1 loads and inspects records through the real route", async ({
  page,
  testActor,
}) => {
  await exercisePublicSurface(page, handoffViewSchemaId, testActor.user_id);
});

test("Workbook surface cartulary.view.status_review.v1 loads and inspects records through the real route", async ({
  page,
  testActor,
}) => {
  await exercisePublicSurface(
    page,
    statusReviewViewSchemaId,
    testActor.user_id,
  );
});

test("Workbook surface cartulary.view.lesson.v1 loads and inspects records through the real route", async ({
  page,
  testActor,
}) => {
  await exercisePublicSurface(page, lessonViewSchemaId, testActor.user_id);
});

test("Workbook surface cartulary.view.findings.v1 loads and inspects records through the real route", async ({
  page,
  testActor,
}) => {
  await exercisePublicSurface(page, findingsViewSchemaId, testActor.user_id);
});

test("Workbook surface cartulary.view.investigative_queries.v1 loads and inspects records through the real route", async ({
  page,
  testActor,
}) => {
  await exercisePublicSurface(
    page,
    investigativeQueriesViewSchemaId,
    testActor.user_id,
  );
});

test("Workbook surface cartulary.view.forensic_keywords.v1 loads and inspects records through the real route", async ({
  page,
  testActor,
}) => {
  await exercisePublicSurface(
    page,
    forensicKeywordsViewSchemaId,
    testActor.user_id,
  );
});

test("Workbook continuation reaches and returns later cartulary.view.assessments.v1 records through the real route", async ({
  page,
  testActor,
}) => {
  await exerciseContinuation(page, assessmentsViewSchemaId, testActor.user_id);
});
