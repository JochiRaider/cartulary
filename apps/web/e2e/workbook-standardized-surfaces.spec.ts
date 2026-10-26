import { applyFilterChip } from "@cartulary/test-utils/grid";
import {
  genericCreateSubmitTestId,
  gridGroupingSelectTestId,
  gridGroupRowTestId,
  workbookQueryEntryTestId,
} from "@cartulary/ui-contracts";
import {
  findingsViewSchemaId,
  forensicKeywordsViewSchemaId,
  investigativeQueriesViewSchemaId,
  notesViewSchemaId,
} from "@cartulary/view-contracts";

import { expect, test } from "./fixtures";

import { collectionItems } from "./support/entities/mentions";
import { createIncident } from "./support/incidents/fixtures";

import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import {
  editExtendedSurfaceCell,
  expectWorkbookSaved,
  openGenericSurface,
  setGenericCreateField,
} from "./support/workbook/authoring";
import {
  createViewRow,
  queryViewRows,
  waitForViewRowByCell,
} from "./support/workbook/query";

test("Findings authoring closes supported hypotheses and groups confidence", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("OPTIONAL-SURFACES"),
    "Workbook inspector workbook-interaction optional surfaces",
  );
  const supportingNote = await createViewRow(
    page,
    incidentId,
    notesViewSchemaId,
    {
      client_txn_id: uniqueTxn("e907-note"),
      "note.title": "workbook-interaction supporting note",
    },
  );

  await openGenericSurface(page, incidentId, findingsViewSchemaId);
  await expect(page).toHaveURL(
    new RegExp(`view_schema_id=${encodeURIComponent(findingsViewSchemaId)}`),
  );
  await setGenericCreateField(
    page,
    "finding.statement",
    "Hypothesis: exfiltration used compressed archives",
  );
  await setGenericCreateField(page, "finding.kind", "hypothesis");
  await setGenericCreateField(page, "finding.confidence_score", "72");
  await page
    .getByTestId(genericCreateSubmitTestId(findingsViewSchemaId))
    .click();
  const finding = await waitForViewRowByCell(
    page,
    incidentId,
    findingsViewSchemaId,
    "finding.statement",
    "Hypothesis: exfiltration used compressed archives",
  );
  await editExtendedSurfaceCell(
    page,
    findingsViewSchemaId,
    finding.record_id,
    "finding.state",
    "closed",
  );
  await expectWorkbookSaved(page);
  await editExtendedSurfaceCell(
    page,
    findingsViewSchemaId,
    finding.record_id,
    "finding.supporting_refs",
    supportingNote.record_id as string,
  );
  await expectWorkbookSaved(page);
  const findingRows = await queryViewRows(
    page,
    incidentId,
    findingsViewSchemaId,
  );
  const refreshedFinding = findingRows.find(
    (row) => row.record_id === finding.record_id,
  );
  if (!refreshedFinding) {
    throw new Error("missing refreshed finding");
  }
  expect(refreshedFinding.cells["finding.kind"]?.value).toBe("hypothesis");
  expect(refreshedFinding.cells["finding.state"]?.value).toBe("closed");
  expect(refreshedFinding.cells["finding.confidence_band"]?.value).toBe("high");
  expect(typeof refreshedFinding.cells["finding.closed_at"]?.value).toBe(
    "string",
  );
  expect(
    collectionItems(refreshedFinding, "finding.supporting_refs"),
  ).toHaveLength(1);
  await applyFilterChip(
    page,
    findingsViewSchemaId,
    "finding.confidence_band",
    "high",
  );
  await expect(
    page.getByTestId(
      workbookQueryEntryTestId(
        findingsViewSchemaId,
        "filter",
        "finding.confidence_band",
      ),
    ),
  ).toContainText("high");
  await page
    .getByTestId(gridGroupingSelectTestId(findingsViewSchemaId))
    .selectOption("finding.kind");
  await expect(
    page.getByTestId(
      gridGroupRowTestId(findingsViewSchemaId, "finding.kind", "hypothesis"),
    ),
  ).toBeVisible();
});

test("Investigative Queries authoring preserves query text and platform filters", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("OPTIONAL-SURFACES"),
    "Workbook inspector workbook-interaction optional surfaces",
  );

  await openGenericSurface(page, incidentId, investigativeQueriesViewSchemaId);
  await setGenericCreateField(page, "investigative_query.platform", "Kusto");
  await setGenericCreateField(
    page,
    "investigative_query.purpose",
    "Locate archive staging",
  );
  await setGenericCreateField(
    page,
    "investigative_query.query_text",
    "DeviceFileEvents | where FileName endswith '.zip'",
  );
  await page
    .getByTestId(genericCreateSubmitTestId(investigativeQueriesViewSchemaId))
    .click();
  const investigativeQuery = await waitForViewRowByCell(
    page,
    incidentId,
    investigativeQueriesViewSchemaId,
    "investigative_query.purpose",
    "Locate archive staging",
  );
  expect(
    typeof investigativeQuery.cells["investigative_query.created_by_user_id"]
      ?.value,
  ).toBe("string");
  await applyFilterChip(
    page,
    investigativeQueriesViewSchemaId,
    "investigative_query.platform",
    "Kusto",
  );
  await expect(
    page.getByTestId(
      workbookQueryEntryTestId(
        investigativeQueriesViewSchemaId,
        "filter",
        "investigative_query.platform",
      ),
    ),
  ).toContainText("kusto");
});

test("Forensic Keywords authoring preserves regex and case sensitivity", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("OPTIONAL-SURFACES"),
    "Workbook inspector workbook-interaction optional surfaces",
  );

  await openGenericSurface(page, incidentId, forensicKeywordsViewSchemaId);
  await setGenericCreateField(page, "forensic_keyword.pattern", "7z\\.exe");
  await setGenericCreateField(
    page,
    "forensic_keyword.reason",
    "Archive utility execution",
  );
  await setGenericCreateField(page, "forensic_keyword.match_mode", "regex");
  await setGenericCreateField(page, "forensic_keyword.case_sensitive", "true");
  await page
    .getByTestId(genericCreateSubmitTestId(forensicKeywordsViewSchemaId))
    .click();
  const forensicKeyword = await waitForViewRowByCell(
    page,
    incidentId,
    forensicKeywordsViewSchemaId,
    "forensic_keyword.pattern",
    "7z\\.exe",
  );
  expect(forensicKeyword.cells["forensic_keyword.match_mode"]?.value).toBe(
    "regex",
  );
  expect(forensicKeyword.cells["forensic_keyword.case_sensitive"]?.value).toBe(
    true,
  );
});
