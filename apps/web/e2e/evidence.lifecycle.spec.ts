import { Buffer } from "node:buffer";
import type {
  AttachBlobToEvidenceRecordResponse,
  ViewRow,
} from "@cartulary/protocol-ts/http";
import { scrollGridTargetIntoView } from "@cartulary/test-utils/grid";
import {
  evidenceAttachFileInputTestId,
  evidencePreviewButtonTestId,
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  gridShellTestId,
  rowCellTestId,
} from "@cartulary/ui-contracts";
import {
  evidenceViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { expect, test } from "./fixtures";

import { collectionItems } from "./support/entities/mentions";
import {
  openEvidenceSurface,
  openTimelineSurface,
} from "./support/evidence/browser";
import { chooseEvidenceFile } from "./support/evidence/uploads";
import { createIncident } from "./support/incidents/fixtures";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { rejectUnexpectedPageErrors } from "./support/runtime/observations";
import {
  createViewRow,
  patchRecord,
  queryViewRows,
  waitForViewRowByCell,
} from "./support/workbook/query";
import { editGenericCell } from "./support/workbook/rowMutations";

async function setGenericCreateField(
  page: Page,
  fieldKey: string,
  value: string,
) {
  const fieldTestId = genericCreateFieldTestId(fieldKey);
  await scrollGridTargetIntoView({
    page,
    surface: evidenceViewSchemaId,
    targetTestId: fieldTestId,
  });
  await page.getByTestId(fieldTestId).fill(value);
}

async function waitForEvidenceRow(
  page: Page,
  incidentId: string,
  title: string,
): Promise<ViewRow> {
  const row = await waitForViewRowByCell(
    page,
    incidentId,
    evidenceViewSchemaId,
    "evidence.title",
    title,
    { timeout: 5_000 },
  );
  await scrollGridTargetIntoView({
    page,
    surface: evidenceViewSchemaId,
    targetTestId: rowCellTestId(row.record_id, "evidence.title"),
  });
  await expect(
    page.getByTestId(rowCellTestId(row.record_id, "evidence.title")),
  ).toBeVisible();
  return row;
}

test.beforeEach(({ page }) => {
  rejectUnexpectedPageErrors(page);
});

test("tracks requested evidence before a blob exists and later advances it", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("EVIDENCE-LIFECYCLE"),
    "Evidence requested evidence",
  );
  const timelineRow = await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("e5-requested-timeline"),
      "timeline.activity_synopsis_text": "Requested package tracking",
    },
  );

  await openEvidenceSurface(page, incidentId);
  await setGenericCreateField(page, "evidence.title", "Requested package");
  await setGenericCreateField(page, "evidence.storage_ref", "ticket://E5-04");
  const submitTestId = genericCreateSubmitTestId(evidenceViewSchemaId);
  await scrollGridTargetIntoView({
    page,
    surface: evidenceViewSchemaId,
    targetTestId: submitTestId,
  });
  await page.getByTestId(submitTestId).click();

  const requested = await waitForEvidenceRow(
    page,
    incidentId,
    "Requested package",
  );
  await scrollGridTargetIntoView({
    page,
    surface: evidenceViewSchemaId,
    targetTestId: rowCellTestId(requested.record_id, "evidence.title"),
  });
  await page
    .getByTestId(rowCellTestId(requested.record_id, "evidence.title"))
    .focus();
  await expect(
    page.getByTestId(rowCellTestId(requested.record_id, "evidence.title")),
  ).toHaveText("Requested package");
  await scrollGridTargetIntoView({
    page,
    surface: evidenceViewSchemaId,
    targetTestId: rowCellTestId(
      requested.record_id,
      "evidence.lifecycle_state",
    ),
  });
  await expect(
    page.getByTestId(
      rowCellTestId(requested.record_id, "evidence.lifecycle_state"),
    ),
  ).toHaveText("requested");
  await scrollGridTargetIntoView({
    page,
    surface: evidenceViewSchemaId,
    targetTestId: rowCellTestId(requested.record_id, "evidence.upload_state"),
  });
  await expect(
    page.getByTestId(
      rowCellTestId(requested.record_id, "evidence.upload_state"),
    ),
  ).toHaveText("pending");

  const linkedTimeline = await patchRecord(page, timelineRow.record_id, {
    view_schema_id: timelineViewSchemaId,
    base_row_version: timelineRow.row_version,
    client_txn_id: uniqueTxn("e5-requested-link"),
    changes: [
      {
        field_key: "timeline.attached_evidence_ids",
        action_payload: {
          kind: "collection_actions_v1",
          actions: [
            {
              op: "add_record_ref",
              linked_record_id: requested.record_id,
            },
          ],
        },
      },
    ],
  });
  expect(
    collectionItems(linkedTimeline, "timeline.attached_evidence_ids").some(
      (item) => item.linked_record_id === requested.record_id,
    ),
  ).toBe(true);
  expect(linkedTimeline.cells["timeline.evidence_count"]?.value).toBe(0);
  expect(linkedTimeline.cells["timeline.has_evidence"]?.value).toBe(false);
  await expect
    .poll(async () => {
      const rows = await queryViewRows(page, incidentId, timelineViewSchemaId);
      const row = rows.find(
        (candidate) => candidate.record_id === timelineRow.record_id,
      );
      return [
        row?.cells["timeline.evidence_count"]?.value,
        row?.cells["timeline.has_evidence"]?.value,
      ];
    })
    .toEqual([0, false]);

  await openTimelineSurface(page, incidentId);
  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toBeVisible();

  await openEvidenceSurface(page, incidentId);
  await scrollGridTargetIntoView({
    page,
    surface: evidenceViewSchemaId,
    targetTestId: evidencePreviewButtonTestId(requested.record_id),
  });
  const attachedResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response
        .url()
        .endsWith(`/evidence-records/${requested.record_id}/attach-blob`),
  );
  await chooseEvidenceFile(
    page.getByTestId(evidenceAttachFileInputTestId(requested.record_id)),
    {
      name: "requested.txt",
      mimeType: "text/plain",
      buffer: Buffer.from(
        "evidence_lifecycle requested evidence payload",
        "utf8",
      ),
    },
  );

  const attached = await attachedResponse;
  expect(attached.ok()).toBe(true);
  const advanced = (
    (await attached.json()) as AttachBlobToEvidenceRecordResponse
  ).data.row;
  expect(advanced.record_id).toBe(requested.record_id);
  expect(advanced.row_version).toBeGreaterThan(requested.row_version);
  expect(advanced.cells["evidence.lifecycle_state"]?.value).toBe("requested");
  await scrollGridTargetIntoView({
    page,
    surface: evidenceViewSchemaId,
    targetTestId: rowCellTestId(advanced.record_id, "evidence.lifecycle_state"),
  });
  await expect(
    page.getByTestId(
      rowCellTestId(advanced.record_id, "evidence.lifecycle_state"),
    ),
  ).toHaveText("requested");
  await scrollGridTargetIntoView({
    page,
    surface: evidenceViewSchemaId,
    targetTestId: rowCellTestId(advanced.record_id, "evidence.upload_state"),
  });
  await expect(
    page.getByTestId(
      rowCellTestId(advanced.record_id, "evidence.upload_state"),
    ),
  ).toHaveText("available");
  await editGenericCell(
    page,
    evidenceViewSchemaId,
    advanced.record_id,
    "evidence.lifecycle_state",
    "available",
  );
  await expect(
    page.getByTestId(
      rowCellTestId(advanced.record_id, "evidence.lifecycle_state"),
    ),
  ).toHaveText("available");
  expect(
    (await queryViewRows(page, incidentId, evidenceViewSchemaId)).filter(
      (row) => row.record_id === requested.record_id,
    ),
  ).toHaveLength(1);
  await expect
    .poll(async () => {
      const rows = await queryViewRows(page, incidentId, timelineViewSchemaId);
      const row = rows.find(
        (candidate) => candidate.record_id === timelineRow.record_id,
      );
      return [
        row?.cells["timeline.evidence_count"]?.value,
        row?.cells["timeline.has_evidence"]?.value,
      ];
    })
    .toEqual([1, true]);
  await openTimelineSurface(page, incidentId);
  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toBeVisible();
  const timelineRows = await queryViewRows(
    page,
    incidentId,
    timelineViewSchemaId,
  );
  expect(
    timelineRows.filter((row) => row.record_id === timelineRow.record_id),
  ).toHaveLength(1);
});
