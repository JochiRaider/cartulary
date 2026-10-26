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

async function exerciseSurface(page: Page, view: string, actorId: string) {
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
  await scrollGridCellIntoView({
    page,
    surface: view,
    recordId: row.record_id,
    cellKey: fixture.field,
  });
  await expect(
    page.getByTestId(rowCellTestId(row.record_id, fixture.field)),
  ).toContainText(String(row.cells[fixture.field]?.value));
  if (view === timelineViewSchemaId)
    await openTimelineInspector(page, row.record_id);
  else await openGenericInspectorForRecord(page, view, row.record_id);
  await expect(
    page.getByTestId(workbookInspectorCloseButtonTestId(view)),
  ).toBeVisible();
  await page.getByTestId(workbookInspectorCloseButtonTestId(view)).click();
  // Assessment rows are append-only; Indicator display values have their own lifecycle editor.
  if (view !== assessmentsViewSchemaId && !view.includes("indicators")) {
    await scrollGridCellIntoView({
      page,
      surface: view,
      recordId: row.record_id,
      cellKey: fixture.field,
    });
    const cell = page
      .getByTestId(rowCellTestId(row.record_id, fixture.field))
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
    await editor.fill(`${String(row.cells[fixture.field]?.value)} edited`);
    await editor.press("Enter");
    expect((await accepted).ok()).toBe(true);
    await expect(
      page.getByTestId(gridRowTestId(view, row.record_id)),
    ).toHaveAttribute("data-grid-row-version", "2");
  }
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

test("Workbook continuation reaches and returns later cartulary.view.timeline.v2 records through the real route", async ({
  page,
  testActor,
}) => {
  await exerciseSurface(page, timelineViewSchemaId, testActor.user_id);
});

test("Workbook continuation reaches and returns later cartulary.view.hosts.v1 records through the real route", async ({
  page,
  testActor,
}) => {
  await exerciseSurface(page, hostsViewSchemaId, testActor.user_id);
});

test("Workbook continuation reaches and returns later cartulary.view.identities.v1 records through the real route", async ({
  page,
  testActor,
}) => {
  await exerciseSurface(page, identitiesViewSchemaId, testActor.user_id);
});

test("Workbook continuation reaches and returns later cartulary.view.notes.v1 records through the real route", async ({
  page,
  testActor,
}) => {
  await exerciseSurface(page, notesViewSchemaId, testActor.user_id);
});

test("Workbook continuation reaches and returns later cartulary.view.evidence.v1 records through the real route", async ({
  page,
  testActor,
}) => {
  await exerciseSurface(page, evidenceViewSchemaId, testActor.user_id);
});

test("Workbook continuation reaches and returns later cartulary.view.indicators.v1 records through the real route", async ({
  page,
  testActor,
}) => {
  await exerciseSurface(page, indicatorsViewSchemaId, testActor.user_id);
});

test("Workbook continuation reaches and returns later cartulary.view.parties.v1 records through the real route", async ({
  page,
  testActor,
}) => {
  await exerciseSurface(page, partiesViewSchemaId, testActor.user_id);
});

test("Workbook continuation reaches and returns later cartulary.view.task_requests.v1 records through the real route", async ({
  page,
  testActor,
}) => {
  await exerciseSurface(page, taskRequestsViewSchemaId, testActor.user_id);
});

test("Workbook continuation reaches and returns later cartulary.view.decisions.v1 records through the real route", async ({
  page,
  testActor,
}) => {
  await exerciseSurface(page, decisionsViewSchemaId, testActor.user_id);
});

test("Workbook continuation reaches and returns later cartulary.view.comm_log.v1 records through the real route", async ({
  page,
  testActor,
}) => {
  await exerciseSurface(page, commLogViewSchemaId, testActor.user_id);
});

test("Workbook continuation reaches and returns later cartulary.view.handoff.v1 records through the real route", async ({
  page,
  testActor,
}) => {
  await exerciseSurface(page, handoffViewSchemaId, testActor.user_id);
});

test("Workbook continuation reaches and returns later cartulary.view.status_review.v1 records through the real route", async ({
  page,
  testActor,
}) => {
  await exerciseSurface(page, statusReviewViewSchemaId, testActor.user_id);
});

test("Workbook continuation reaches and returns later cartulary.view.lesson.v1 records through the real route", async ({
  page,
  testActor,
}) => {
  await exerciseSurface(page, lessonViewSchemaId, testActor.user_id);
});

test("Workbook continuation reaches and returns later cartulary.view.findings.v1 records through the real route", async ({
  page,
  testActor,
}) => {
  await exerciseSurface(page, findingsViewSchemaId, testActor.user_id);
});

test("Workbook continuation reaches and returns later cartulary.view.investigative_queries.v1 records through the real route", async ({
  page,
  testActor,
}) => {
  await exerciseSurface(
    page,
    investigativeQueriesViewSchemaId,
    testActor.user_id,
  );
});

test("Workbook continuation reaches and returns later cartulary.view.forensic_keywords.v1 records through the real route", async ({
  page,
  testActor,
}) => {
  await exerciseSurface(page, forensicKeywordsViewSchemaId, testActor.user_id);
});

test("Workbook continuation reaches and returns later cartulary.view.assessments.v1 records through the real route", async ({
  page,
  testActor,
}) => {
  await exerciseSurface(page, assessmentsViewSchemaId, testActor.user_id);
});
