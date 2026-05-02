import { applyFilterChip, removeFilterChip } from "@cartulary/test-utils";
import { rowCellTestId } from "@cartulary/ui-contracts";
import type { Page } from "@playwright/test";

import { expect, test } from "./fixtures";
import {
  createIncident,
  createViewRow,
  uniqueIncidentKey,
  uniqueTxn,
} from "./helpers";
import {
  assessmentsViewSchemaId,
  commLogViewSchemaId,
  createAssessmentViaUI,
  decisionsViewSchemaId,
  editGenericCell,
  evidenceViewSchemaId,
  expectAssessmentGridOrder,
  handoffViewSchemaId,
  hostsViewSchemaId,
  lessonViewSchemaId,
  notesViewSchemaId,
  partiesViewSchemaId,
  statusReviewViewSchemaId,
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

  type AssessmentState =
    | "cleared"
    | "confirmed"
    | "disproven"
    | "suspected"
    | "unknown";
  const assessmentEntries: Array<{
    state: AssessmentState;
    band: string;
    assessedAt: string;
    supportRefs: string[];
  }> = [
    {
      state: "unknown",
      band: "unset",
      assessedAt: "2026-04-24T10:00:00Z",
      supportRefs: [],
    },
    {
      state: "suspected",
      band: "low",
      assessedAt: "2026-04-24T11:00:00Z",
      supportRefs: [],
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
      supportRefs: [],
    },
    {
      state: "cleared",
      band: "high",
      assessedAt: "2026-04-24T14:00:00Z",
      supportRefs: [],
    },
  ];
  const created = {} as Record<AssessmentState, ViewRow>;
  for (const entry of assessmentEntries) {
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

test("E-4-06 creates and edits required workbook mutation surfaces through typed generic controls", async ({
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

  await openGenericSurface(page, incidentId, partiesViewSchemaId, "Parties");
  await expectGenericCreateMinimum(page, partiesViewSchemaId, "Display name");
  await setGenericCreateField(page, "party.display_name", "Browser Party");
  await setGenericCreateField(page, "party.party_kind", "organization");
  await page
    .getByTestId(`generic-create-submit-${partiesViewSchemaId}`)
    .click();
  const party = await waitForViewRowByCell(
    page,
    incidentId,
    partiesViewSchemaId,
    "party.display_name",
    "Browser Party",
  );
  await editGenericCell(
    page,
    partiesViewSchemaId,
    party.record_id,
    "party.primary_email",
    "ir@example.test",
  );
  await expect(
    page.getByTestId(rowCellTestId(party.record_id, "party.primary_email")),
  ).toHaveText("ir@example.test");

  await openGenericSurface(page, incidentId, notesViewSchemaId, "Notes");
  await expect(page.getByRole("heading", { name: "Notes" })).toBeVisible();
  await expectGenericCreateMinimum(page, notesViewSchemaId, "Title or body");
  await setGenericCreateField(page, "note.title", "Browser note");
  await setGenericCreateField(
    page,
    "note.body",
    "Created from generic workbook UI",
  );
  await setGenericCreateField(page, "note.tags", "browser");
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

  await openGenericSurface(
    page,
    incidentId,
    decisionsViewSchemaId,
    "Decisions",
  );
  await expect(page.getByRole("heading", { name: "Decisions" })).toBeVisible();
  await expectGenericCreateMinimum(
    page,
    decisionsViewSchemaId,
    "Summary, decision type",
  );
  await setGenericCreateField(page, "decision.summary", "Browser decision");
  await setGenericCreateField(page, "decision.decision_type", "containment");
  await setGenericCreateField(
    page,
    "decision.rationale",
    "Generic UI decision rationale.",
  );
  await waitForGenericOption(
    page,
    "generic-create-field-decision.support_refs",
    support.record_id,
  );
  await setGenericCreateField(page, "decision.support_refs", support.record_id);
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

  await openGenericSurface(
    page,
    incidentId,
    taskRequestsViewSchemaId,
    "Task Requests",
  );
  await expect(
    page.getByRole("heading", { name: "Task Requests" }),
  ).toBeVisible();
  await expectGenericCreateMinimum(page, taskRequestsViewSchemaId, "Title");
  await setGenericCreateField(page, "task.title", "Browser task");
  await setGenericCreateField(page, "task.task_kind", "collection");
  await waitForGenericOption(
    page,
    "generic-create-field-task.decision_record_id",
    decision.record_id,
  );
  await setGenericCreateField(
    page,
    "task.decision_record_id",
    decision.record_id,
  );
  await waitForGenericOption(
    page,
    "generic-create-field-task.linked_record_ids",
    support.record_id,
  );
  await setGenericCreateField(
    page,
    "task.linked_record_ids",
    support.record_id,
  );
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

  await openGenericSurface(page, incidentId, evidenceViewSchemaId, "Evidence");
  await expectGenericCreateMinimum(
    page,
    evidenceViewSchemaId,
    "Evidence needs",
  );
  await setGenericCreateField(page, "evidence.title", "Browser evidence");
  await setGenericCreateField(page, "evidence.storage_ref", "slot/browser");
  await waitForGenericOption(
    page,
    "generic-create-field-evidence.collector_party_id",
    party.record_id,
  );
  await setGenericCreateField(
    page,
    "evidence.collector_party_id",
    party.record_id,
  );
  await page
    .getByTestId(`generic-create-submit-${evidenceViewSchemaId}`)
    .click();
  const evidence = await waitForViewRowByCell(
    page,
    incidentId,
    evidenceViewSchemaId,
    "evidence.title",
    "Browser evidence",
  );
  await expect(
    page.getByTestId(rowCellTestId(evidence.record_id, "evidence.storage_ref")),
  ).toHaveText("slot/browser");

  await openGenericSurface(
    page,
    incidentId,
    commLogViewSchemaId,
    "Communications Log",
  );
  await expectGenericCreateMinimum(page, commLogViewSchemaId, "Type");
  await setGenericCreateField(page, "comm_log.comm_type", "briefing");
  await setGenericCreateField(page, "comm_log.audience", "leadership");
  await setGenericCreateField(page, "comm_log.channel_or_meeting", "Bridge");
  await setGenericCreateField(page, "comm_log.summary", "Browser comm log");
  await waitForGenericOption(
    page,
    "generic-create-field-comm_log.audience_party_ids",
    party.record_id,
  );
  await setGenericCreateField(
    page,
    "comm_log.audience_party_ids",
    party.record_id,
  );
  await page
    .getByTestId(`generic-create-submit-${commLogViewSchemaId}`)
    .click();
  const commLog = await waitForViewRowByCell(
    page,
    incidentId,
    commLogViewSchemaId,
    "comm_log.summary",
    "Browser comm log",
  );
  await editGenericCell(
    page,
    commLogViewSchemaId,
    commLog.record_id,
    "comm_log.summary",
    "Updated browser comm log",
  );
  await expect(
    page.getByTestId(rowCellTestId(commLog.record_id, "comm_log.summary")),
  ).toHaveText("Updated browser comm log");

  await openGenericSurface(page, incidentId, handoffViewSchemaId, "Handoff");
  await expectGenericCreateMinimum(page, handoffViewSchemaId, "current state");
  await setGenericCreateField(
    page,
    "handoff.current_state_summary",
    "Browser handoff",
  );
  await waitForGenericOption(
    page,
    "generic-create-field-handoff.open_task_ids",
    task.record_id,
  );
  await setGenericCreateField(page, "handoff.open_task_ids", task.record_id);
  await waitForGenericOption(
    page,
    "generic-create-field-handoff.open_decision_ids",
    decision.record_id,
  );
  await setGenericCreateField(
    page,
    "handoff.open_decision_ids",
    decision.record_id,
  );
  await setGenericCreateField(page, "handoff.open_risk_refs", "Browser risk");
  await page
    .getByTestId(`generic-create-submit-${handoffViewSchemaId}`)
    .click();
  const handoff = await waitForViewRowByCell(
    page,
    incidentId,
    handoffViewSchemaId,
    "handoff.current_state_summary",
    "Browser handoff",
  );
  await expect(
    page.getByTestId(
      rowCellTestId(handoff.record_id, "handoff.current_state_summary"),
    ),
  ).toHaveText("Browser handoff");

  await openGenericSurface(
    page,
    incidentId,
    statusReviewViewSchemaId,
    "Status Review",
  );
  await expectGenericCreateMinimum(
    page,
    statusReviewViewSchemaId,
    "Current state",
  );
  await setGenericCreateField(
    page,
    "status_review.current_state_summary",
    "Browser status review",
  );
  await waitForGenericOption(
    page,
    "generic-create-field-status_review.pending_evidence_ids",
    evidence.record_id,
  );
  await setGenericCreateField(
    page,
    "status_review.pending_evidence_ids",
    evidence.record_id,
  );
  await setGenericCreateField(
    page,
    "status_review.blocked_task_ids",
    task.record_id,
  );
  await page
    .getByTestId(`generic-create-submit-${statusReviewViewSchemaId}`)
    .click();
  const statusReview = await waitForViewRowByCell(
    page,
    incidentId,
    statusReviewViewSchemaId,
    "status_review.current_state_summary",
    "Browser status review",
  );
  await expect(
    page.getByTestId(
      rowCellTestId(
        statusReview.record_id,
        "status_review.current_state_summary",
      ),
    ),
  ).toHaveText("Browser status review");

  await openGenericSurface(page, incidentId, lessonViewSchemaId, "Lesson");
  await expectGenericCreateMinimum(page, lessonViewSchemaId, "Summary");
  await setGenericCreateField(page, "lesson.summary", "Browser lesson");
  await waitForGenericOption(
    page,
    "generic-create-field-lesson.evidence_refs",
    evidence.record_id,
  );
  await setGenericCreateField(page, "lesson.evidence_refs", evidence.record_id);
  await setGenericCreateField(
    page,
    "lesson.follow_up_task_ids",
    task.record_id,
  );
  await page.getByTestId(`generic-create-submit-${lessonViewSchemaId}`).click();
  const lesson = await waitForViewRowByCell(
    page,
    incidentId,
    lessonViewSchemaId,
    "lesson.summary",
    "Browser lesson",
  );
  await editGenericCell(
    page,
    lessonViewSchemaId,
    lesson.record_id,
    "lesson.closure_state",
    "closed",
  );
  await expect(
    page.getByTestId(rowCellTestId(lesson.record_id, "lesson.closure_state")),
  ).toHaveText("closed");
});

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

async function expectGenericCreateMinimum(
  page: Page,
  viewSchemaId: string,
  message: string,
) {
  await page.getByTestId(`generic-create-submit-${viewSchemaId}`).click();
  await expect(page.getByTestId("generic-mutation-error")).toContainText(
    message,
  );
}

async function setGenericCreateField(
  page: Page,
  fieldKey: string,
  value: string | string[],
) {
  const input = page.getByTestId(`generic-create-field-${fieldKey}`);
  const tagName = await input.evaluate((element) => element.tagName);
  if (tagName === "SELECT") {
    await input.selectOption(value);
    return;
  }
  await input.fill(Array.isArray(value) ? value.join("\n") : value);
}

async function waitForGenericOption(page: Page, testId: string, value: string) {
  await expect(
    page.getByTestId(testId).locator(`option[value="${value}"]`),
  ).toHaveCount(1);
}
