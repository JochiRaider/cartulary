import { scrollGridTargetIntoView } from "@cartulary/test-utils/grid";
import {
  currentIncidentRoleTestId,
  entityInspectButtonTestId,
  gridRowTestId,
  timelinePreviewRowTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import { expect, type Page, type Response } from "@playwright/test";

import { timelineViewSchemaId } from "../contracts/workbookSurfaces";
import { queryViewRows } from "../workbook/query";
import {
  collectionItems,
  findRow,
  requireItemByRawText,
  type ViewRow,
} from "./mentions";

type MergeEnvelope = {
  data: {
    survivor_record_id: string;
    loser_record_id: string;
    merged_into_record_id: string;
    merge_summary: {
      record_type: string;
      repointed_mention_resolution_count: number;
      repointed_link_count: number;
    };
  };
};

export async function exerciseEntityMerge(
  page: Page,
  options: {
    dependentRow: ViewRow;
    entityType: "host" | "identity";
    expectAdminRole?: boolean;
    fieldKey: string;
    incidentId: string;
    loser: ViewRow;
    loserLabel: string;
    mergeReason: string;
    rawText: string;
    recordType: string;
    resolvedLabel: string;
    survivor: ViewRow;
    survivorLabel: string;
    viewSchemaId: string;
  },
) {
  await page.goto(
    `/?incident_id=${options.incidentId}&view_schema_id=${encodeURIComponent(
      options.viewSchemaId,
    )}`,
  );
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  if (options.expectAdminRole === true) {
    await expectCurrentIncidentRole(page, "Current incident role: admin");
  }
  await expect(
    page.getByTestId(
      gridRowTestId(options.viewSchemaId, options.survivor.record_id),
    ),
  ).toBeVisible();
  await expect(
    page.getByTestId(
      gridRowTestId(options.viewSchemaId, options.loser.record_id),
    ),
  ).toBeVisible();
  const inspectButtonTestId = entityInspectButtonTestId(
    options.entityType,
    options.survivor.record_id,
  );
  await scrollGridTargetIntoView({
    page,
    surface: options.viewSchemaId,
    targetTestId: inspectButtonTestId,
  });
  await page.getByTestId(inspectButtonTestId).click();
  const inspectorTestId = entityInspectorLocalTestId(options.entityType);
  await expect(page.getByTestId(inspectorTestId)).toContainText(
    options.survivorLabel,
  );
  await expect(page.getByTestId("merge-start")).toBeVisible();
  await page.getByTestId("merge-start").click();
  await page
    .getByTestId("merge-loser-record")
    .selectOption(options.loser.record_id);
  await page.getByTestId("merge-reason").fill(options.mergeReason);
  await expect(page.getByTestId("merge-plan")).toContainText(
    `Survivor ${options.survivorLabel}`,
  );
  await expect(page.getByTestId("merge-plan")).toContainText(
    `loser ${options.loserLabel}`,
  );
  await expect(page.getByTestId("merge-plan")).toContainText(
    options.survivor.record_id,
  );
  await expect(page.getByTestId("merge-plan")).toContainText(
    options.loser.record_id,
  );
  const mergeResponsePromise = waitForMergeResponse(
    page,
    options.survivor.record_id,
  );
  await page.getByTestId("merge-confirm").click();
  const mergeEnvelope = await readMergeEnvelope(await mergeResponsePromise);
  await expect(
    page.getByTestId(
      gridRowTestId(options.viewSchemaId, options.loser.record_id),
    ),
  ).toHaveCount(0);
  await expect(page.getByTestId(inspectorTestId)).toContainText(
    options.survivorLabel,
  );
  await expect(page.getByTestId("merge-message")).toContainText(
    `Merged ${options.loserLabel} into ${options.survivorLabel}`,
  );
  await expect(
    page.getByTestId(timelinePreviewRowTestId(options.dependentRow.record_id)),
  ).toBeVisible();
  await expect(
    page
      .getByTestId(timelinePreviewRowTestId(options.dependentRow.record_id))
      .getByLabel(`Resolved ${options.resolvedLabel}`),
  ).toBeVisible();
  const entityRowsAfter = (await queryViewRows(
    page,
    options.incidentId,
    options.viewSchemaId,
  )) as ViewRow[];
  const timelineRowsAfter = (await queryViewRows(
    page,
    options.incidentId,
    timelineViewSchemaId,
  )) as ViewRow[];
  const dependentRowAfter = findRow(
    timelineRowsAfter,
    options.dependentRow.record_id,
  );
  const dependentItemAfter = requireItemByRawText(
    collectionItems(dependentRowAfter, options.fieldKey),
    options.rawText,
  );
  expect(mergeEnvelope.data.survivor_record_id).toBe(
    options.survivor.record_id,
  );
  expect(mergeEnvelope.data.loser_record_id).toBe(options.loser.record_id);
  expect(mergeEnvelope.data.merged_into_record_id).toBe(
    options.survivor.record_id,
  );
  expect(mergeEnvelope.data.merge_summary.record_type).toBe(options.recordType);
  expect(
    mergeEnvelope.data.merge_summary.repointed_mention_resolution_count,
  ).toBeGreaterThan(0);
  expect(mergeEnvelope.data.merge_summary.repointed_link_count).toBeGreaterThan(
    0,
  );
  expect(
    entityRowsAfter.some((row) => row.record_id === options.survivor.record_id),
  ).toBeTruthy();
  expect(
    entityRowsAfter.some((row) => row.record_id === options.loser.record_id),
  ).toBeFalsy();
  expect(String(dependentItemAfter.item_kind)).toBe("resolved_ref");
  expect(String(dependentItemAfter.raw_text)).toBe(options.rawText);
  expect(String(dependentItemAfter.resolved_record_id)).toBe(
    options.survivor.record_id,
  );
  return {
    dependentItemAfter,
    dependentRowAfter,
    entityRowsAfter,
    mergeEnvelope,
    timelineRowsAfter,
  };
}

function waitForMergeResponse(page: Page, survivorRecordId: string) {
  return page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response.url().endsWith(`/api/v1/records/${survivorRecordId}/merge`),
  );
}

async function readMergeEnvelope(response: Response) {
  expect(response.ok()).toBeTruthy();
  return (await response.json()) as MergeEnvelope;
}

async function expectCurrentIncidentRole(page: Page, roleText: string) {
  const accountMenuTrigger = page.getByLabel(
    "Account and application navigation",
  );
  await accountMenuTrigger.click();
  await expect(page.getByTestId(currentIncidentRoleTestId())).toHaveText(
    roleText,
  );
  await accountMenuTrigger.click();
}

function entityInspectorLocalTestId(entityType: "host" | "identity") {
  return entityType === "host" ? "host-inspector" : "identity-inspector";
}
