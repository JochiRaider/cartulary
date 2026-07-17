import {
  gridRowTestId,
  gridSavedRowsSelector,
  gridShellTestId,
} from "@cartulary/ui-contracts";
import { expect, type Page } from "@playwright/test";

import { assessmentsViewSchemaId } from "../contracts/workbookSurfaces";
import { readTimelineMutation } from "../workbook/rowMutations";

export async function createAssessmentViaUI(
  page: Page,
  options: {
    assessedAt: string;
    confidenceBand: string;
    rationale: string;
    state: string;
    supportRecordIds: string[];
  },
) {
  await page.getByTestId("assessment-create-state").selectOption(options.state);
  await page
    .getByTestId("assessment-create-confidence-band")
    .selectOption(options.confidenceBand);
  await page.getByTestId("assessment-create-rationale").fill(options.rationale);
  await page
    .getByTestId("assessment-create-assessed-at")
    .fill(options.assessedAt);
  if (options.supportRecordIds.length > 0) {
    await expect(
      page.getByTestId("assessment-create-support-refs").locator("option"),
    ).toHaveCount(options.supportRecordIds.length);
    await page
      .getByTestId("assessment-create-support-refs")
      .selectOption(options.supportRecordIds);
  }
  const responsePromise = waitForAssessmentCreate(page);
  await page.getByTestId("assessment-create-submit").click();
  const envelope = await readTimelineMutation(await responsePromise);
  await expect(
    page.getByTestId(
      gridRowTestId(assessmentsViewSchemaId, envelope.data.row.record_id),
    ),
  ).toBeVisible();
  return envelope.data.row;
}

export function waitForAssessmentCreate(page: Page) {
  return page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response.url().endsWith(`/views/${assessmentsViewSchemaId}/rows`),
  );
}

export async function expectAssessmentGridOrder(
  page: Page,
  expected: string[],
) {
  const grid = page.getByTestId(gridShellTestId(assessmentsViewSchemaId));
  await expect
    .poll(async () =>
      grid
        .locator(gridSavedRowsSelector())
        .evaluateAll((rows) =>
          rows.map((row) => row.getAttribute("data-grid-record-id") ?? ""),
        ),
    )
    .toEqual(expected);
}
