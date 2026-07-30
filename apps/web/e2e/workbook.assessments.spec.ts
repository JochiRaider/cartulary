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

import { expect, test } from "./fixtures";
import {
  createAssessmentViaUI,
  expectAssessmentGridOrder,
  waitForAssessmentCreate,
} from "./support/assessments/fixtures";
import {
  assessmentsViewSchemaId,
  hostsViewSchemaId,
  timelineViewSchemaId,
} from "./support/contracts/workbookSurfaces";
import type { ViewRow } from "./support/entities/mentions";
import { createIncident } from "./support/incidents/fixtures";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { createViewRow } from "./support/workbook/query";

test("creates append-only assessment history through the workbook UI", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("WORKBOOK-ASSESSMENTS"),
    "Record relationships assessment workbook E2E",
  );
  const subject = (await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("assessment-host"),
    "host.display_name": "Assessment Host",
    "host.hostname": "assessment-host.example.test",
  })) as ViewRow;
  const support = (await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("assessment-support"),
    "timeline.activity_synopsis_text": "Assessment support event",
  })) as ViewRow;

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
    page.getByTestId(assessmentCreateControlTestId("subject")),
  ).toHaveValue(subject.record_id);

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
  const subject = (await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("assessment-follow-on-host"),
    "host.display_name": "Follow-on Host",
    "host.hostname": "assessment-follow-on.example.test",
  })) as ViewRow;
  const original = (await createViewRow(
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
  )) as ViewRow;

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
    page.getByTestId(assessmentCreateControlTestId("subject")),
  ).toHaveValue(subject.record_id);
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
    page.getByTestId(assessmentCreateControlTestId("support-refs")),
  ).toHaveValues([]);

  await page
    .getByTestId(assessmentCreateControlTestId("rationale"))
    .fill("Cancelled draft.");
  await page.keyboard.press("Escape");
  await expect(page.getByTestId(assessmentCreatePanelTestId())).toHaveCount(0);
  await expect(inspectorToggle).toBeFocused();
  await expect(originalRow).toHaveAttribute("data-inspector-active", "true");

  await inspectorToggle.click();
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
  await expect(
    page.getByTestId(
      gridRowTestId(assessmentsViewSchemaId, createEnvelope.data.row.record_id),
    ),
  ).toBeVisible();
  await expect(originalRow).toHaveAttribute("data-inspector-active", "true");
});
