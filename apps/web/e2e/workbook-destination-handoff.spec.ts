import { scrollGridCellIntoView } from "@cartulary/test-utils/grid";
import {
  draftCellTestId,
  genericCreateFieldTestId,
  gridScrollportSelector,
  gridShellTestId,
  rowCellTestId,
  surfaceTabTestId,
  timelineInspectorTestId,
  workbookColumnsMenuTestId,
  workbookColumnsMenuTriggerTestId,
  workbookGridEditorTestId,
  workbookInspectorCloseButtonTestId,
  workbookShellReadyTestId,
  workbookSurfacesMenuOptionTestId,
  workbookSurfacesMenuTriggerTestId,
} from "@cartulary/ui-contracts";
import {
  evidenceViewSchemaId,
  notesViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page, Request } from "@playwright/test";
import { expect, test } from "./fixtures";
import { loginLocalSession } from "./support/auth/browserSession";
import { createIncident } from "./support/incidents/fixtures";
import { createIncidentMemberUser } from "./support/incidents/memberships";
import {
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { holdBrowserRequest } from "./support/transport/requestInterception";
import { createViewRow } from "./support/workbook/query";
import {
  activateCommittedGridCell,
  ensureTimelineGridTargetVisible,
  openTimelineInspector,
} from "./support/workbook/rowMutations";

const title = "Destination handoff record";
const dateField = "timeline.date_entered_text";
const activityField = "timeline.activity_synopsis_text";
const cell = (page: Page, recordId: string, field = dateField) =>
  page
    .getByTestId(rowCellTestId(recordId, field))
    .locator('xpath=ancestor::*[@role="gridcell"][1]');

async function seed(page: Page) {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("HANDOFF"),
    "Destination handoff",
  );
  const record = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("handoff-record"),
    [activityField]: title,
  });
  return { incidentId, recordId: record.record_id };
}
async function enter(page: Page, incidentId: string) {
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
}
async function surface(page: Page, view: string) {
  const tab = page.getByTestId(surfaceTabTestId(view));
  if ((page.viewportSize()?.width ?? 1440) > 1024) await tab.click();
  else {
    await page.getByTestId(workbookSurfacesMenuTriggerTestId()).click();
    await page.getByTestId(workbookSurfacesMenuOptionTestId(view)).click();
  }
  await expect(page.getByTestId(gridShellTestId(view))).toBeVisible();
}
async function pin(page: Page, recordId: string) {
  await openTimelineInspector(page, recordId);
  await page
    .getByRole("button", { name: "Pin record to Work", exact: true })
    .click();
  await page
    .getByTestId(workbookInspectorCloseButtonTestId(timelineViewSchemaId))
    .click();
}
async function work(page: Page) {
  await page.getByRole("button", { name: "Work", exact: true }).press("Enter");
  await expect(
    page.getByRole("heading", { name: "Work", exact: true }),
  ).toBeFocused();
}
async function openPin(page: Page) {
  await work(page);
  await page.getByRole("button", { name: title, exact: true }).press("Enter");
}
async function returnToOrigin(page: Page, inspect = false) {
  if (
    !(await page
      .getByRole("button", { name: "Return", exact: true })
      .isVisible())
  )
    await page
      .getByRole("button", { name: "View options controls", exact: true })
      .click();
  if (inspect) {
    await page.getByLabel("Return options", { exact: true }).click();
    await page
      .getByRole("button", { name: "Return and inspect", exact: true })
      .press("Enter");
  } else
    await page
      .getByRole("button", { name: "Return", exact: true })
      .press("Enter");
}
async function navigationCell(page: Page, recordId: string, field = dateField) {
  await expect(
    page.getByRole("button", { name: "Close Work", exact: true }),
  ).toHaveCount(0);
  await expect(cell(page, recordId, field)).toBeFocused();
  await expect(cell(page, recordId, field)).toHaveAttribute(
    "aria-selected",
    "true",
  );
  await expect(page.locator('[data-grid-editing="true"]')).toHaveCount(0);
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
  expect(
    await cell(page, recordId, field).evaluate(
      (element) => element.closest("[inert]") === null,
    ),
  ).toBe(true);
}
function observeWrites(page: Page) {
  const writes: string[] = [];
  page.on("request", (request) => {
    if (
      ["POST", "PATCH", "PUT", "DELETE"].includes(request.method()) &&
      /\/(?:rows|records)(?:\/|$)/.test(new URL(request.url()).pathname)
    )
      writes.push(request.url());
  });
  return writes;
}

async function exerciseHandoff(
  page: Page,
  viewport: { width: number; height: number },
) {
  const { incidentId, recordId } = await seed(page);
  await page.setViewportSize(viewport);
  await enter(page, incidentId);
  await pin(page, recordId);
  const writes = observeWrites(page);
  await surface(page, notesViewSchemaId);
  await openPin(page);
  await navigationCell(page, recordId);
  await openPin(page);
  await navigationCell(page, recordId);
  // Arrow navigation proves the committed cell is usable, not merely selected.
  await page.keyboard.press("ArrowRight");
  await expect(cell(page, recordId)).not.toBeFocused();
  await ensureTimelineGridTargetVisible(
    page,
    rowCellTestId(recordId, activityField),
  );
  await activateCommittedGridCell(cell(page, recordId, activityField));
  await work(page);
  await page.keyboard.press("Escape");
  await expect(
    page.getByRole("button", { name: "Close Work", exact: true }),
  ).toHaveCount(0);
  await work(page);
  await page
    .getByRole("button", { name: "Task Requests", exact: true })
    .press("Enter");
  await expect(
    page.getByTestId(gridShellTestId(taskRequestsViewSchemaId)),
  ).toBeVisible();
  await expect(
    page.getByRole("button", { name: "Close Work", exact: true }),
  ).toHaveCount(0);
  await expect
    .poll(async () =>
      page
        .getByTestId(gridShellTestId(taskRequestsViewSchemaId))
        .evaluate((element) => element.contains(document.activeElement)),
    )
    .toBe(true);
  await work(page);
  await returnToOrigin(page);
  await navigationCell(page, recordId, activityField);
  await work(page);
  await page
    .getByRole("button", { name: "Task Requests", exact: true })
    .press("Enter");
  await expect(
    page.getByTestId(gridShellTestId(taskRequestsViewSchemaId)),
  ).toBeVisible();
  await work(page);
  await returnToOrigin(page, true);
  await expect(
    page.getByRole("button", { name: "Close Work", exact: true }),
  ).toHaveCount(0);
  await expect(page.getByTestId(timelineInspectorTestId())).toBeVisible();
  await expect(page.getByTestId(timelineInspectorTestId())).toContainText(
    title,
  );
  await expect
    .poll(async () =>
      page
        .getByTestId(timelineInspectorTestId())
        .evaluate((element) => element.contains(document.activeElement)),
    )
    .toBe(true);
  await expect(page.locator('[data-grid-editing="true"]')).toHaveCount(0);
  expect(writes).toEqual([]);
}

test("Work and Return hand off usable destinations at 1440 pixels.", async ({
  page,
}) => {
  await exerciseHandoff(page, { width: 1440, height: 900 });
});
test("Work and Return hand off usable destinations at 1024 pixels.", async ({
  page,
}) => {
  await exerciseHandoff(page, { width: 1024, height: 768 });
});
test("Work and Return hand off usable destinations at 768 pixels.", async ({
  page,
}) => {
  await exerciseHandoff(page, { width: 768, height: 768 });
});

test("Record handoff falls back through hidden fields to the root and retains unsubmitted work.", async ({
  page,
}) => {
  const { incidentId, recordId } = await seed(page);
  const evidence = await createViewRow(page, incidentId, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn("handoff-retained"),
    "evidence.title": "Retained invalid authoring",
  });
  await page.setViewportSize({ width: 1440, height: 900 });
  await enter(page, incidentId);
  await pin(page, recordId);
  const writes = observeWrites(page);
  await page
    .getByTestId(workbookColumnsMenuTriggerTestId(timelineViewSchemaId))
    .click();
  const columns = page.getByTestId(
    workbookColumnsMenuTestId(timelineViewSchemaId),
  );
  const checked = columns.getByRole("checkbox", { checked: true });
  // Leave activity as the only eligible data column; a pin has no field hint.
  while (await checked.count()) await checked.first().uncheck();
  await columns
    .getByRole("checkbox", { name: "Activity Synopsis", exact: true })
    .check();
  await page.keyboard.press("Escape");
  await page.setViewportSize({ width: 1024, height: 768 });
  await surface(page, notesViewSchemaId);
  const note = page.getByTestId(genericCreateFieldTestId("note.title"));
  await note.fill("Unsubmitted note survives navigation");
  await openPin(page);
  await navigationCell(page, recordId, activityField);
  await page.setViewportSize({ width: 1440, height: 900 });
  await page
    .getByTestId(workbookColumnsMenuTriggerTestId(timelineViewSchemaId))
    .click();
  while (await checked.count()) await checked.first().uncheck();
  await page.keyboard.press("Escape");
  await page.setViewportSize({ width: 1024, height: 768 });
  await surface(page, notesViewSchemaId);
  await expect(note).toHaveValue("Unsubmitted note survives navigation");
  await openPin(page);
  await expect(
    page.getByRole("button", { name: "Close Work", exact: true }),
  ).toHaveCount(0);
  const root = page
    .getByTestId(gridShellTestId(timelineViewSchemaId))
    .locator(gridScrollportSelector());
  await expect(root).toBeFocused();
  await page.keyboard.press("Tab");
  await expect(root).not.toBeFocused();
  await page.setViewportSize({ width: 1440, height: 900 });
  await surface(page, evidenceViewSchemaId);
  const field = "evidence.requested_at";
  const activate = async () => {
    await scrollGridCellIntoView({
      page,
      surface: evidenceViewSchemaId,
      recordId: evidence.record_id,
      cellKey: field,
    });
    await page.getByTestId(rowCellTestId(evidence.record_id, field)).click();
  };
  await activate();
  const editor = page.getByTestId(
    workbookGridEditorTestId(evidence.record_id, field),
  );
  await editor.fill(" 2026-02-30T12:00:00 ");
  await editor.press("Enter");
  await expect(editor).toBeFocused();
  await openPin(page);
  await expect(root).toBeFocused();
  await work(page);
  await returnToOrigin(page);
  await expect(editor).toHaveCount(0);
  await expect(cell(page, evidence.record_id, field)).toBeFocused();
  await activate();
  await expect(editor).toHaveValue(" 2026-02-30T12:00:00 ");
  expect(writes).toEqual([]);
});

test("Delayed record reads preserve Work and cannot supersede newer interaction.", async ({
  page,
}) => {
  const { incidentId, recordId } = await seed(page);
  await page.setViewportSize({ width: 1024, height: 768 });
  await enter(page, incidentId);
  await pin(page, recordId);
  await surface(page, notesViewSchemaId);
  const writes = observeWrites(page);
  const locatePath = `**/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/locate`;
  await page.route(locatePath, (route) => route.abort("failed"));
  await openPin(page);
  await page
    .getByRole("button", { name: "Navigation", exact: true })
    .press("Enter");
  await expect(
    page.getByRole("dialog", { name: "Navigation details" }),
  ).toBeFocused();
  await expect(
    page.getByRole("button", { name: "Retry navigation", exact: true }),
  ).toBeVisible();
  await expect(
    page.getByRole("button", { name: "Close Work", exact: true }),
  ).toBeVisible();
  await expect(
    page.getByTestId(gridShellTestId(notesViewSchemaId)),
  ).toBeVisible();
  await page.unroute(locatePath);
  const held = await holdBrowserRequest(page, {
    method: "POST",
    path: `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/locate`,
  });
  try {
    await page
      .getByRole("button", { name: "Retry navigation", exact: true })
      .press("Enter");
    await held.waitForHit;
    await expect(
      page.getByRole("button", { name: "Close Work", exact: true }),
    ).toBeVisible();
    await expect(
      page.getByTestId(gridShellTestId(notesViewSchemaId)),
    ).toBeVisible();
    const settled = new Promise<void>((resolve) => {
      const onSettled = (request: Request) => {
        if (!request.url().endsWith(`/views/${timelineViewSchemaId}/locate`))
          return;
        page.off("requestfinished", onSettled);
        page.off("requestfailed", onSettled);
        resolve();
      };
      page.on("requestfinished", onSettled);
      page.on("requestfailed", onSettled);
    });
    await page
      .getByRole("button", { name: "Task Requests", exact: true })
      .press("Enter");
    await expect(
      page.getByTestId(gridShellTestId(taskRequestsViewSchemaId)),
    ).toBeVisible();
    held.release();
    await settled;
    await expect(
      page.getByTestId(gridShellTestId(taskRequestsViewSchemaId)),
    ).toBeVisible();
    await expect
      .poll(async () =>
        page
          .getByTestId(gridShellTestId(taskRequestsViewSchemaId))
          .evaluate((element) => element.contains(document.activeElement)),
      )
      .toBe(true);
    await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
    await work(page);
    await returnToOrigin(page);
    await expect(
      page
        .getByTestId(gridShellTestId(notesViewSchemaId))
        .locator(gridScrollportSelector()),
    ).toBeFocused();
    await expect(
      page.getByTestId(genericCreateFieldTestId("note.title")),
    ).not.toBeFocused();
    expect(writes).toEqual([]);
  } finally {
    await held.dispose();
  }
});

test("Viewer record handoff selects a committed cell without enabling writes.", async ({
  page,
}) => {
  const { incidentId, recordId } = await seed(page);
  const viewer = await createIncidentMemberUser(page, incidentId, {
    display_name: "Handoff Viewer",
    email: uniqueEmail("handoff-viewer"),
    initial_password: "HandoffViewer1!",
    role: "viewer",
    is_deployment_admin: false,
    mfa_required: false,
  });
  await loginLocalSession(page, viewer.email, viewer.initial_password);
  await page.setViewportSize({ width: 768, height: 768 });
  await enter(page, incidentId);
  await pin(page, recordId);
  const writes = observeWrites(page);
  await surface(page, notesViewSchemaId);
  await openPin(page);
  await navigationCell(page, recordId);
  await page.keyboard.press("Enter");
  await expect(page.locator('[data-grid-editing="true"]')).toHaveCount(0);
  await expect(page.getByTestId(draftCellTestId(dateField))).toHaveCount(0);
  expect(writes).toEqual([]);
});
