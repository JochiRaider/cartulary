import {
  applyFilterChip,
  removeFilterChip,
  rowCellTestId,
} from "@cartulary/test-utils";

import { expect, test } from "./fixtures";
import {
  createIncident,
  createViewRow,
  uniqueIncidentKey,
  uniqueTxn,
} from "./helpers";
import {
  assessmentsViewSchemaId,
  createAssessmentViaUI,
  decisionsViewSchemaId,
  editGenericCell,
  expectAssessmentGridOrder,
  hostsViewSchemaId,
  notesViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
  type ViewRow,
  waitForViewRowByCell,
} from "./phase4Helpers";

test("E-4-05 creates append-only assessment history through the workbook UI", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E4ASSESS"),
    "Phase 4 assessment workbook E2E",
  );
  const subject = (await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("assessment-host"),
    "host.display_name": "Assessment Host",
    "host.hostname": "assessment-host.example.test",
  })) as ViewRow;
  const support = (await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("assessment-support"),
    "timeline.summary": "Assessment support event",
  })) as ViewRow;

  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      assessmentsViewSchemaId,
    )}`,
  );
  await expect(page.getByText("Timeline workbook shell")).toBeVisible();
  await expect(page.getByTestId("assessment-create-panel")).toBeVisible();
  await expect(page.getByTestId("assessment-create-subject")).toHaveValue(
    subject.record_id,
  );

  const created: Record<string, ViewRow> = {};
  for (const entry of [
    {
      state: "unknown",
      band: "unset",
      assessedAt: "2026-04-24T10:00:00Z",
      supportRefs: [] as string[],
    },
    {
      state: "suspected",
      band: "low",
      assessedAt: "2026-04-24T11:00:00Z",
      supportRefs: [] as string[],
    },
    {
      state: "confirmed",
      band: "medium",
      assessedAt: "2026-04-24T12:00:00Z",
      supportRefs: [support.record_id],
    },
    {
      state: "disproven",
      band: "medium",
      assessedAt: "2026-04-24T13:00:00Z",
      supportRefs: [] as string[],
    },
    {
      state: "cleared",
      band: "high",
      assessedAt: "2026-04-24T14:00:00Z",
      supportRefs: [] as string[],
    },
  ]) {
    created[entry.state] = await createAssessmentViaUI(page, {
      assessedAt: entry.assessedAt,
      confidenceBand: entry.band,
      rationale: `Assessment ${entry.state} rationale.`,
      state: entry.state,
      supportRecordIds: entry.supportRefs,
    });
  }

  await expectAssessmentGridOrder(page, [
    created.cleared.record_id,
    created.disproven.record_id,
    created.confirmed.record_id,
    created.suspected.record_id,
    created.unknown.record_id,
  ]);

  await expect(
    page.getByTestId(
      rowCellTestId(created.unknown.record_id, "assessment.confidence_band"),
    ),
  ).toHaveText("unset");
  await expect(
    page.getByTestId(
      rowCellTestId(created.suspected.record_id, "assessment.confidence_band"),
    ),
  ).toHaveText("low");
  await expect(
    page.getByTestId(
      rowCellTestId(created.confirmed.record_id, "assessment.confidence_band"),
    ),
  ).toHaveText("medium");
  await expect(
    page.getByTestId(
      rowCellTestId(created.cleared.record_id, "assessment.confidence_band"),
    ),
  ).toHaveText("high");
  await expect(
    page.getByTestId(
      rowCellTestId(
        created.confirmed.record_id,
        "assessment.supporting_link_count",
      ),
    ),
  ).toHaveText("1");

  await applyFilterChip(
    page,
    assessmentsViewSchemaId,
    "assessment.confidence_band",
    "high",
  );
  await expectAssessmentGridOrder(page, [created.cleared.record_id]);
  await removeFilterChip(
    page,
    assessmentsViewSchemaId,
    "assessment.confidence_band",
  );
  await expectAssessmentGridOrder(page, [
    created.cleared.record_id,
    created.disproven.record_id,
    created.confirmed.record_id,
    created.suspected.record_id,
    created.unknown.record_id,
  ]);

  await applyFilterChip(
    page,
    assessmentsViewSchemaId,
    "assessment.assessment_state",
    "disproven",
  );
  await expectAssessmentGridOrder(page, [created.disproven.record_id]);
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
  await expectAssessmentGridOrder(page, [created.cleared.record_id]);
});

test("E-4-06 creates and edits Notes, Task Requests, and Decisions through generic workbook surfaces", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E4GENERIC"),
    "Phase 4 generic workbook mutation E2E",
  );
  const support = (await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("generic-support"),
    "timeline.summary": "Generic surface support event",
  })) as ViewRow;

  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      notesViewSchemaId,
    )}`,
  );
  await expect(page.getByRole("heading", { name: "Notes" })).toBeVisible();
  await page.getByTestId(`generic-create-submit-${notesViewSchemaId}`).click();
  await expect(page.getByTestId("generic-mutation-error")).toContainText(
    "invalid_mutation_payload",
  );
  await page
    .getByTestId("generic-create-field-note.title")
    .fill("Browser note");
  await page
    .getByTestId("generic-create-field-note.body")
    .fill("Created from generic workbook UI");
  await page.getByTestId("generic-create-field-note.tags").fill("browser");
  await page.getByTestId(`generic-create-submit-${notesViewSchemaId}`).click();
  const note = await waitForViewRowByCell(
    page,
    incidentId,
    notesViewSchemaId,
    "note.title",
    "Browser note",
  );
  await expect(
    page.getByTestId(rowCellTestId(note.record_id, "note.title")),
  ).toHaveText("Browser note");
  await editGenericCell(
    page,
    notesViewSchemaId,
    note.record_id,
    "note.body",
    "Updated browser note",
  );
  await expect(
    page.getByTestId(rowCellTestId(note.record_id, "note.body")),
  ).toHaveText("Updated browser note");

  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      decisionsViewSchemaId,
    )}`,
  );
  await expect(page.getByRole("heading", { name: "Decisions" })).toBeVisible();
  await page
    .getByTestId("generic-create-field-decision.summary")
    .fill("Browser decision");
  await page
    .getByTestId("generic-create-field-decision.decision_type")
    .fill("containment");
  await page
    .getByTestId("generic-create-field-decision.rationale")
    .fill("Generic UI decision rationale.");
  await page
    .getByTestId("generic-create-field-decision.support_refs")
    .fill(support.record_id);
  await page
    .getByTestId(`generic-create-submit-${decisionsViewSchemaId}`)
    .click();
  const decision = await waitForViewRowByCell(
    page,
    incidentId,
    decisionsViewSchemaId,
    "decision.summary",
    "Browser decision",
  );
  await expect(
    page.getByTestId(rowCellTestId(decision.record_id, "decision.status")),
  ).toHaveText("proposed");
  await editGenericCell(
    page,
    decisionsViewSchemaId,
    decision.record_id,
    "decision.summary",
    "Updated browser decision",
  );
  await expect(
    page.getByTestId(rowCellTestId(decision.record_id, "decision.summary")),
  ).toHaveText("Updated browser decision");

  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      taskRequestsViewSchemaId,
    )}`,
  );
  await expect(
    page.getByRole("heading", { name: "Task Requests" }),
  ).toBeVisible();
  await page
    .getByTestId("generic-create-field-task.title")
    .fill("Browser task");
  await page
    .getByTestId("generic-create-field-task.task_kind")
    .fill("collection");
  await page
    .getByTestId("generic-create-field-task.decision_record_id")
    .fill(decision.record_id);
  await page
    .getByTestId("generic-create-field-task.linked_record_ids")
    .fill(support.record_id);
  await page
    .getByTestId(`generic-create-submit-${taskRequestsViewSchemaId}`)
    .click();
  const task = await waitForViewRowByCell(
    page,
    incidentId,
    taskRequestsViewSchemaId,
    "task.title",
    "Browser task",
  );
  await expect(
    page.getByTestId(rowCellTestId(task.record_id, "task.status")),
  ).toHaveText("open");
  await expect(
    page.getByTestId(rowCellTestId(task.record_id, "task.priority")),
  ).toHaveText("normal");
  await expect(
    page.getByTestId(rowCellTestId(task.record_id, "task.linked_record_count")),
  ).toHaveText("1");
  await editGenericCell(
    page,
    taskRequestsViewSchemaId,
    task.record_id,
    "task.title",
    "Updated browser task",
  );
  await expect(
    page.getByTestId(rowCellTestId(task.record_id, "task.title")),
  ).toHaveText("Updated browser task");
});
