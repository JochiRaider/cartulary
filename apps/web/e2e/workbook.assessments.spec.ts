import type { ViewRow } from "@cartulary/protocol-ts/http";
import {
  applyFilterChip,
  removeFilterChip,
  scrollGridCellIntoView,
} from "@cartulary/test-utils/grid";
import {
  assessmentCreateControlTestId,
  assessmentCreatePanelTestId,
  gridRowTestId,
  rowCellTestId,
  workbookAddRowButtonTestId,
  workbookInspectorFeatureActionTestId,
  workbookInspectorToggleTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import {
  assessmentsViewSchemaId,
  hostsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import { expect, test } from "./fixtures";
import {
  createAssessmentViaUI,
  expectAssessmentGridOrder,
  waitForAssessmentCreate,
} from "./support/assessments/fixtures";
import { createIncident } from "./support/incidents/fixtures";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { observePageRequests } from "./support/runtime/observations";
import { installIncidentSocketMonitor } from "./support/transport/incidentSocket";
import { fetchFullRecordHistory } from "./support/workbook/history";
import { createViewRow, queryViewRows } from "./support/workbook/query";

test("creates append-only assessment history through the workbook UI", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("WORKBOOK-ASSESSMENTS"),
    "Record relationships assessment workbook E2E",
  );
  const subject = await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("assessment-host"),
    "host.display_name": "Assessment Host",
    "host.hostname": "assessment-host.example.test",
  });
  const support = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("assessment-support"),
    "timeline.activity_synopsis_text": "Assessment support event",
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
  await expect(page.getByTestId(assessmentCreatePanelTestId())).toBeVisible();
  await expect(
    page
      .getByTestId(assessmentCreateControlTestId("subject"))
      .getByRole("radio", { checked: true }),
  ).toHaveCount(0);

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
      subjectRecordId: subject.record_id,
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
  await scrollGridCellIntoView({
    cellKey: "assessment.supporting_link_count",
    page,
    recordId: created.confirmed.record_id,
    surface: assessmentsViewSchemaId,
  });
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

test("appends a subject-only follow-on while preserving keyboard selection", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("WORKBOOK-ASSESSMENT-FOLLOW-ON"),
    "Assessment follow-on workbook E2E",
  );
  const subject = await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("assessment-follow-on-host"),
    "host.display_name": "Follow-on Host",
    "host.hostname": "assessment-follow-on.example.test",
  });
  const original = await createViewRow(
    page,
    incidentId,
    assessmentsViewSchemaId,
    {
      client_txn_id: uniqueTxn("assessment-original"),
      "assessment.subject_ref": subject.record_id,
      "assessment.subject_type": "host",
      "assessment.assessment_state": "confirmed",
      "assessment.confidence_score": 85,
      "assessment.rationale": "Original assessment rationale.",
      "assessment.assessed_at": "2026-04-24T12:00:00Z",
    },
  );

  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      assessmentsViewSchemaId,
    )}`,
  );
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await page
    .getByTestId(
      rowCellTestId(original.record_id, "assessment.assessment_state"),
    )
    .click();
  const originalRow = page.getByTestId(
    gridRowTestId(assessmentsViewSchemaId, original.record_id),
  );
  await expect(originalRow).toHaveAttribute("data-inspector-active", "true");

  const originalHistory = await fetchFullRecordHistory(
    page,
    original.record_id,
  );
  await applyFilterChip(
    page,
    assessmentsViewSchemaId,
    "assessment.assessment_state",
    "confirmed",
  );
  const inspectorToggle = page.getByTestId(
    workbookInspectorToggleTestId(assessmentsViewSchemaId),
  );
  const followOnAction = page.getByTestId(
    workbookInspectorFeatureActionTestId(
      assessmentsViewSchemaId,
      "create_related.assessment",
    ),
  );
  await inspectorToggle.click();
  await followOnAction.click();
  await expect(
    page
      .getByTestId(assessmentCreateControlTestId("subject"))
      .locator(`input[value="${subject.record_id}"]`),
  ).toBeChecked();
  await expect(
    page.getByTestId(assessmentCreateControlTestId("subject-type")),
  ).toHaveValue("host");
  await expect(
    page.getByTestId(assessmentCreateControlTestId("state")),
  ).toHaveValue("unknown");
  await expect(
    page.getByTestId(assessmentCreateControlTestId("confidence-band")),
  ).toHaveValue("unset");
  await expect(
    page.getByTestId(assessmentCreateControlTestId("rationale")),
  ).toHaveValue("");
  await expect(
    page.getByTestId(assessmentCreateControlTestId("assessed-at")),
  ).toHaveValue("");
  await expect(
    page.getByRole("region", { name: "Assessment supporting records" }),
  ).toContainText("No supporting records selected.");

  await page
    .getByTestId(assessmentCreateControlTestId("rationale"))
    .fill("Cancelled draft.");
  await page.keyboard.press("Escape");
  await expect(page.getByTestId(assessmentCreatePanelTestId())).toHaveCount(0);
  await expect(inspectorToggle).toBeFocused();
  await expect(originalRow).toHaveAttribute("data-inspector-active", "true");

  await inspectorToggle.click();
  await page
    .getByRole("button", { name: "Resume assessment draft", exact: true })
    .click();
  await expect(
    page.getByTestId(assessmentCreateControlTestId("rationale")),
  ).toHaveValue("Cancelled draft.");
  await page
    .getByRole("button", { name: "Discard assessment draft", exact: true })
    .click();
  await followOnAction.click();
  await page
    .getByTestId(assessmentCreateControlTestId("state"))
    .selectOption("suspected");
  await page
    .getByTestId(assessmentCreateControlTestId("rationale"))
    .fill("Fresh follow-on rationale.");
  const createResponsePromise = waitForAssessmentCreate(page);
  await page.getByTestId(assessmentCreateControlTestId("submit")).click();
  const createResponse = await createResponsePromise;
  expect(createResponse.request().postDataJSON()).toMatchObject({
    "assessment.subject_ref": subject.record_id,
    "assessment.subject_type": "host",
    "assessment.assessment_state": "suspected",
    "assessment.confidence_score": null,
    "assessment.rationale": "Fresh follow-on rationale.",
  });
  expect(createResponse.request().postDataJSON()).not.toHaveProperty(
    "assessment.assessed_at",
  );
  expect(createResponse.request().postDataJSON()).not.toHaveProperty(
    "assessment.support_refs",
  );
  expect(createResponse.request().postDataJSON()).not.toHaveProperty(
    "supersedes",
  );
  const createEnvelope = (await createResponse.json()) as {
    data: { row: ViewRow };
  };
  expect(createEnvelope.data.row.record_id).not.toBe(original.record_id);
  await expectAssessmentGridOrder(page, [original.record_id]);
  expect(await fetchFullRecordHistory(page, original.record_id)).toEqual(
    originalHistory,
  );
  await removeFilterChip(
    page,
    assessmentsViewSchemaId,
    "assessment.assessment_state",
  );
  await expect(
    page.getByTestId(
      gridRowTestId(assessmentsViewSchemaId, createEnvelope.data.row.record_id),
    ),
  ).toBeVisible();
  await expect(originalRow).toHaveAttribute("data-inspector-active", "true");
});

test("assessment workflow keeps invalid timestamp drafts local", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("ASSESSMENT-WORKFLOW"),
    "Workbook inspector assessment-workflow assessment workflow",
  );
  const fixtureSocket = installIncidentSocketMonitor(page, incidentId);
  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      assessmentsViewSchemaId,
    )}`,
  );
  await fixtureSocket.waitForAcceptedSocket();
  const subjectA = await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("e905-host-a"),
    "host.display_name": "Workbook inspector Assessment Host A",
    "host.hostname": "workbook_interaction-assessment-a.example.test",
  });
  const subjectB = await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("e905-host-b"),
    "host.display_name": "Workbook inspector Assessment Host B",
    "host.hostname": "workbook_interaction-assessment-b.example.test",
  });
  const support = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("e905-support"),
    "timeline.activity_synopsis_text":
      "Workbook inspector assessment support event",
  });

  // Commit does not establish collaboration publication. Finish fixture effects
  // before opening a draft whose selected candidates require explicit review
  // after a relevant record_changed observation.
  for (const row of [subjectA, subjectB, support]) {
    await fixtureSocket.waitForMessage("record_changed", {
      matches: (message) =>
        message.payload.record_id === row.record_id &&
        message.payload.row_version === row.row_version,
    });
  }
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await page
    .getByTestId(workbookAddRowButtonTestId(assessmentsViewSchemaId))
    .click();
  await expect(page.getByTestId(assessmentCreatePanelTestId())).toBeVisible();
  await expect(
    page
      .getByTestId(assessmentCreateControlTestId("subject"))
      .getByRole("radio", { checked: true }),
  ).toHaveCount(0);
  await page
    .getByTestId(assessmentCreateControlTestId("subject"))
    .locator(`input[type="radio"][value="${subjectA.record_id}"]`)
    .check();

  const invalidTimestamp = "2026-04-24 12:00:00";
  await page
    .getByTestId(assessmentCreateControlTestId("state"))
    .selectOption("confirmed");
  await page
    .getByTestId(assessmentCreateControlTestId("confidence-band"))
    .selectOption("high");
  await page
    .getByTestId(assessmentCreateControlTestId("rationale"))
    .fill("Invalid timestamp must remain a draft.");
  await page
    .getByTestId(assessmentCreateControlTestId("assessed-at"))
    .fill(invalidTimestamp);
  const assessmentWrites: string[] = [];
  observePageRequests(page, (request) => {
    if (
      request.method() === "POST" &&
      request.url().endsWith(`/views/${assessmentsViewSchemaId}/rows`)
    )
      assessmentWrites.push(request.postData() ?? "");
  });
  await page.getByTestId(assessmentCreateControlTestId("submit")).click();
  await expect(
    page.getByTestId(assessmentCreateControlTestId("message")),
  ).toContainText("Enter an RFC3339 timestamp");
  expect(assessmentWrites).toEqual([]);
  await expect(
    page.getByTestId(assessmentCreateControlTestId("assessed-at")),
  ).toHaveValue(invalidTimestamp);
  expect(
    await queryViewRows(page, incidentId, assessmentsViewSchemaId),
  ).toHaveLength(0);

  const createdUnknown = await createAssessmentViaUI(page, {
    subjectRecordId: subjectA.record_id,
    assessedAt: "2026-04-24T10:00:00Z",
    confidenceBand: "unset",
    rationale: "Workbook inspector unknown rationale.",
    state: "unknown",
    supportRecordIds: [],
  });
  const createdSuspected = await createAssessmentViaUI(page, {
    subjectRecordId: subjectA.record_id,
    assessedAt: "2026-04-24T11:00:00Z",
    confidenceBand: "low",
    rationale: "Workbook inspector suspected rationale.",
    state: "suspected",
    supportRecordIds: [],
  });
  const createdConfirmed = await createAssessmentViaUI(page, {
    subjectRecordId: subjectA.record_id,
    assessedAt: "2026-04-24T12:00:00Z",
    confidenceBand: "medium",
    rationale: "Workbook inspector confirmed rationale.",
    state: "confirmed",
    supportRecordIds: [support.record_id as string],
  });

  const createdDisproven = await createAssessmentViaUI(page, {
    subjectRecordId: subjectB.record_id,
    assessedAt: "2026-04-24T13:00:00Z",
    confidenceBand: "medium",
    rationale: "Workbook inspector disproven rationale.",
    state: "disproven",
    supportRecordIds: [],
  });
  const createdCleared = await createAssessmentViaUI(page, {
    subjectRecordId: subjectB.record_id,
    assessedAt: "2026-04-24T14:00:00Z",
    confidenceBand: "high",
    rationale: "Workbook inspector cleared rationale.",
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
  await scrollGridCellIntoView({
    cellKey: "assessment.subject_ref",
    page,
    recordId: createdDisproven.record_id,
    surface: assessmentsViewSchemaId,
  });
  await expect(
    page.getByTestId(
      rowCellTestId(createdDisproven.record_id, "assessment.subject_ref"),
    ),
  ).toHaveText(subjectB.record_id as string);
  await scrollGridCellIntoView({
    cellKey: "assessment.supporting_link_count",
    page,
    recordId: createdConfirmed.record_id,
    surface: assessmentsViewSchemaId,
  });
  await expect(
    page.getByTestId(
      rowCellTestId(
        createdConfirmed.record_id,
        "assessment.supporting_link_count",
      ),
    ),
  ).toHaveText("1");
  await scrollGridCellIntoView({
    cellKey: "assessment.confidence_band",
    page,
    recordId: createdUnknown.record_id,
    surface: assessmentsViewSchemaId,
  });
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
