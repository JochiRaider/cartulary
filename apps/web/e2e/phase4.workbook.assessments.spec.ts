import { applyFilterChip, removeFilterChip } from "@cartulary/test-utils";
import {
  rowCellTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";

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
  expectAssessmentGridOrder,
  hostsViewSchemaId,
  timelineViewSchemaId,
  type ViewRow,
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
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
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
