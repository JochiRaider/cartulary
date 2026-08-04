import {
  assessmentCreateControlTestId,
  gridRowTestId,
  gridSavedRowsSelector,
  gridShellTestId,
} from "@cartulary/ui-contracts";
import { assessmentsViewSchemaId } from "@cartulary/view-contracts";
import { expect, type Page } from "@playwright/test";
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
  await page
    .getByTestId(assessmentCreateControlTestId("state"))
    .selectOption(options.state);
  await page
    .getByTestId(assessmentCreateControlTestId("confidence-band"))
    .selectOption(options.confidenceBand);
  await page
    .getByTestId(assessmentCreateControlTestId("rationale"))
    .fill(options.rationale);
  await page
    .getByTestId(assessmentCreateControlTestId("assessed-at"))
    .fill(options.assessedAt);
  if (options.supportRecordIds.length > 0) {
    await expect(
      page
        .getByTestId(assessmentCreateControlTestId("support-refs"))
        .locator("option"),
    ).toHaveCount(options.supportRecordIds.length);
    await page
      .getByTestId(assessmentCreateControlTestId("support-refs"))
      .selectOption(options.supportRecordIds);
  }
  const responsePromise = waitForAssessmentCreate(page);
  await page.getByTestId(assessmentCreateControlTestId("submit")).click();
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
