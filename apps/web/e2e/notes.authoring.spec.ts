import {
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  gridShellTestId,
  rowCellTestId,
  workbookInspectorToggleTestId,
} from "@cartulary/ui-contracts";
import {
  notesViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";

import { expect, test } from "./fixtures";

import { createIncident } from "./support/incidents/fixtures";

import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";

import { createViewRow, queryViewRows } from "./support/workbook/query";

test("Notes tab creates artifact-backed linked notes", async ({ page }) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("LINKED-NOTES"),
    "Workbook inspector linked-notes Notes linked create",
  );
  const source = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("e903-source"),
    "timeline.activity_synopsis_text": "Workbook inspector linked note source",
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
  await page
    .getByTestId(genericCreateFieldTestId("note.title"))
    .fill("Workbook inspector linked-notes linked note");
  await page
    .getByTestId(genericCreateFieldTestId("note.body"))
    .fill("Created from the Notes tab with a source record link.");
  await page
    .getByRole("button", { name: "Choose source", exact: true })
    .click();
  await expect(
    page
      .getByRole("group", { name: "Note source", exact: true })
      .locator(`input[type="radio"][value="${source.record_id}"]`),
  ).toHaveCount(1, { timeout: 15_000 });
  await page
    .getByRole("group", { name: "Note source", exact: true })
    .locator(`input[type="radio"][value="${source.record_id}"]`)
    .check();
  await page.getByRole("button", { name: "Apply source", exact: true }).click();

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
  ).toHaveText("Workbook inspector linked-notes linked note");

  const rows = await queryViewRows(page, incidentId, notesViewSchemaId);
  const noteRow = rows.find((row) => row.record_id === noteRecordId);
  expect(noteRow).toBeTruthy();
  expect(noteRow?.cells["note.linked_record_count"]?.value).toBe(1);
});
