import { applyFilterChip, removeFilterChip } from "@cartulary/test-utils/grid";
import {
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  workbookQueryEntryTestId,
} from "@cartulary/ui-contracts";
import {
  commLogViewSchemaId,
  decisionsViewSchemaId,
  evidenceViewSchemaId,
  handoffViewSchemaId,
  lessonViewSchemaId,
  partiesViewSchemaId,
  statusReviewViewSchemaId,
  taskRequestsViewSchemaId,
} from "@cartulary/view-contracts";

import { expect, test } from "./fixtures";

import { collectionItems } from "./support/entities/mentions";
import { createIncident } from "./support/incidents/fixtures";

import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { disableWorkbookSockets } from "./support/transport/socketIsolation";
import {
  editExtendedSurfaceCell,
  expectWorkbookSaved,
  openGenericSurface,
  setGenericCreateField,
  waitForReferenceCandidate,
} from "./support/workbook/authoring";
import {
  createViewRow,
  queryViewRows,
  waitForViewRowByCell,
} from "./support/workbook/query";
import { editGenericCell } from "./support/workbook/rowMutations";

test("Communications Log links decisions tasks and audience parties", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("COORDINATION-WORKFLOW"),
    "Workbook inspector workbook-interaction coordination workflows",
  );
  const party = await createViewRow(page, incidentId, partiesViewSchemaId, {
    client_txn_id: uniqueTxn("e906coord-party"),
    "party.display_name": "Coordination Review Lead",
    "party.party_kind": "team",
  });

  const decision = await createViewRow(
    page,
    incidentId,
    decisionsViewSchemaId,
    {
      client_txn_id: uniqueTxn("e906coord-decision"),
      "decision.summary": "coordination-review decision",
      "decision.decision_type": "containment",
      "decision.rationale": "Coordination workflow decision.",
    },
  );
  const task = await createViewRow(page, incidentId, taskRequestsViewSchemaId, {
    client_txn_id: uniqueTxn("e906coord-task"),
    "task.title": "coordination-review task",
    "task.task_kind": "follow_up",
  });

  await disableWorkbookSockets(page);
  await openGenericSurface(page, incidentId, commLogViewSchemaId);
  await expect(page).toHaveURL(
    new RegExp(`view_schema_id=${encodeURIComponent(commLogViewSchemaId)}`),
  );
  await setGenericCreateField(page, "comm_log.comm_type", "briefing");
  await setGenericCreateField(
    page,
    "comm_log.audience",
    "coordination-review audience",
  );
  await setGenericCreateField(
    page,
    "comm_log.channel_or_meeting",
    "Coordination bridge",
  );
  await setGenericCreateField(
    page,
    "comm_log.summary",
    "coordination-review log",
  );
  await waitForReferenceCandidate(
    page,
    genericCreateFieldTestId("comm_log.decision_ids"),
    decision.record_id as string,
    decisionsViewSchemaId,
  );
  await setGenericCreateField(
    page,
    "comm_log.decision_ids",
    decision.record_id as string,
  );
  await waitForReferenceCandidate(
    page,
    genericCreateFieldTestId("comm_log.action_task_ids"),
    task.record_id as string,
    taskRequestsViewSchemaId,
  );
  await setGenericCreateField(
    page,
    "comm_log.action_task_ids",
    task.record_id as string,
  );
  await waitForReferenceCandidate(
    page,
    genericCreateFieldTestId("comm_log.audience_party_ids"),
    party.record_id as string,
    partiesViewSchemaId,
  );
  await setGenericCreateField(
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
    "coordination-review log",
  );
  expect(comm.cells["comm_log.audience"]?.value).toBe(
    "coordination-review audience",
  );
  expect(collectionItems(comm, "comm_log.decision_ids")).toHaveLength(1);
  expect(collectionItems(comm, "comm_log.action_task_ids")).toHaveLength(1);
  expect(collectionItems(comm, "comm_log.audience_party_ids")).toHaveLength(1);
});

test("Handoff authoring validates ownership and persists acknowledgment", async ({
  page,
  testActor,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("COORDINATION-WORKFLOW"),
    "Workbook inspector workbook-interaction coordination workflows",
  );

  await disableWorkbookSockets(page);
  await openGenericSurface(page, incidentId, handoffViewSchemaId);
  await expect(page).toHaveURL(
    new RegExp(`view_schema_id=${encodeURIComponent(handoffViewSchemaId)}`),
  );
  await setGenericCreateField(
    page,
    "handoff.current_state_summary",
    "coordination-review handoff state",
  );
  await page
    .getByTestId(genericCreateSubmitTestId(handoffViewSchemaId))
    .click();
  await expect(
    page.getByRole("alert").filter({ hasText: "Incoming Owner" }),
  ).toBeVisible();
  await setGenericCreateField(
    page,
    "handoff.incoming_owner_user_id",
    testActor.user_id,
  );
  await page
    .getByTestId(genericCreateSubmitTestId(handoffViewSchemaId))
    .click();
  const handoff = await waitForViewRowByCell(
    page,
    incidentId,
    handoffViewSchemaId,
    "handoff.current_state_summary",
    "coordination-review handoff state",
  );
  await editGenericCell(
    page,
    handoffViewSchemaId,
    handoff.record_id,
    "handoff.next_checks",
    "Verify next checkpoint before shift change",
  );
  await expectWorkbookSaved(page);
  await editGenericCell(
    page,
    handoffViewSchemaId,
    handoff.record_id,
    "handoff.open_risk_refs",
    "Residual endpoint access risk",
  );
  await expectWorkbookSaved(page);
  await editGenericCell(
    page,
    handoffViewSchemaId,
    handoff.record_id,
    "handoff.acknowledged_at",
    "2026-05-26T14:00:00Z",
  );
  await expectWorkbookSaved(page);
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
      workbookQueryEntryTestId(
        handoffViewSchemaId,
        "filter",
        "handoff.ack_state",
      ),
    ),
  ).toContainText("acknowledged");
  await removeFilterChip(page, handoffViewSchemaId, "handoff.ack_state");
});

test("Status Review links pending work and filters the next report day", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("COORDINATION-WORKFLOW"),
    "Workbook inspector workbook-interaction coordination workflows",
  );

  const evidence = await createViewRow(page, incidentId, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn("e906coord-evidence"),
    "evidence.title": "coordination-review evidence",
  });
  const decision = await createViewRow(
    page,
    incidentId,
    decisionsViewSchemaId,
    {
      client_txn_id: uniqueTxn("e906coord-decision"),
      "decision.summary": "coordination-review decision",
      "decision.decision_type": "containment",
      "decision.rationale": "Coordination workflow decision.",
    },
  );
  const task = await createViewRow(page, incidentId, taskRequestsViewSchemaId, {
    client_txn_id: uniqueTxn("e906coord-task"),
    "task.title": "coordination-review task",
    "task.task_kind": "follow_up",
  });

  await disableWorkbookSockets(page);
  await openGenericSurface(page, incidentId, statusReviewViewSchemaId);
  await setGenericCreateField(
    page,
    "status_review.current_state_summary",
    "coordination-review status review state",
  );
  await page
    .getByTestId(genericCreateSubmitTestId(statusReviewViewSchemaId))
    .click();
  const status = await waitForViewRowByCell(
    page,
    incidentId,
    statusReviewViewSchemaId,
    "status_review.current_state_summary",
    "coordination-review status review state",
  );
  await editExtendedSurfaceCell(
    page,
    statusReviewViewSchemaId,
    status.record_id,
    "status_review.blocked_task_ids",
    task.record_id as string,
  );
  await expectWorkbookSaved(page);
  await editExtendedSurfaceCell(
    page,
    statusReviewViewSchemaId,
    status.record_id,
    "status_review.pending_evidence_ids",
    evidence.record_id as string,
  );
  await expectWorkbookSaved(page);
  await editExtendedSurfaceCell(
    page,
    statusReviewViewSchemaId,
    status.record_id,
    "status_review.open_decision_ids",
    decision.record_id as string,
  );
  await expectWorkbookSaved(page);
  await editGenericCell(
    page,
    statusReviewViewSchemaId,
    status.record_id,
    "status_review.next_report_at",
    "2026-05-27T15:30:00Z",
  );
  await expectWorkbookSaved(page);
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
      workbookQueryEntryTestId(
        statusReviewViewSchemaId,
        "filter",
        "status_review.next_report_day",
      ),
    ),
  ).toContainText("2026-05-27");
  await removeFilterChip(
    page,
    statusReviewViewSchemaId,
    "status_review.next_report_day",
  );
});

test("Lesson authoring links follow-up work and closes independently", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("COORDINATION-WORKFLOW"),
    "Workbook inspector workbook-interaction coordination workflows",
  );

  const evidence = await createViewRow(page, incidentId, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn("e906coord-evidence"),
    "evidence.title": "coordination-review evidence",
  });

  const task = await createViewRow(page, incidentId, taskRequestsViewSchemaId, {
    client_txn_id: uniqueTxn("e906coord-task"),
    "task.title": "coordination-review task",
    "task.task_kind": "follow_up",
  });

  await disableWorkbookSockets(page);
  await openGenericSurface(page, incidentId, lessonViewSchemaId);
  await setGenericCreateField(
    page,
    "lesson.summary",
    "coordination-review lesson workflow",
  );
  await page.getByTestId(genericCreateSubmitTestId(lessonViewSchemaId)).click();
  const lesson = await waitForViewRowByCell(
    page,
    incidentId,
    lessonViewSchemaId,
    "lesson.summary",
    "coordination-review lesson workflow",
  );
  await editExtendedSurfaceCell(
    page,
    lessonViewSchemaId,
    lesson.record_id,
    "lesson.follow_up_task_ids",
    task.record_id as string,
  );
  await expectWorkbookSaved(page);
  await editExtendedSurfaceCell(
    page,
    lessonViewSchemaId,
    lesson.record_id,
    "lesson.evidence_refs",
    evidence.record_id as string,
  );
  await expectWorkbookSaved(page);
  await editGenericCell(
    page,
    lessonViewSchemaId,
    lesson.record_id,
    "lesson.closure_state",
    "closed",
  );
  await expectWorkbookSaved(page);
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
      workbookQueryEntryTestId(
        lessonViewSchemaId,
        "filter",
        "lesson.closure_state",
      ),
    ),
  ).toContainText("closed");
});
