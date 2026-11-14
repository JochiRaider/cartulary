import { scrollGridTargetIntoView } from "@cartulary/test-utils/grid";
import {
  currentIncidentRoleTestId,
  entityInspectButtonTestId,
  entityInspectorTestId,
  entityMergeControlTestId,
  gridRowTestId,
  gridScrollportSelector,
  timelinePreviewRowTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
} from "@cartulary/view-contracts";
import { expect, type Page } from "@playwright/test";
import { readHttpOperationResponse } from "../transport/publicHttpOperationClient";

export async function exerciseEntityMerge(
  page: Page,
  options: {
    dependentRecordId: string;
    entityType: "host" | "identity";
    expectAdminRole?: boolean;
    incidentId: string;
    loserRecordId: string;
    loserLabel: string;
    mergeReason: string;
    resolvedLabel: string;
    survivorRecordId: string;
    survivorLabel: string;
  },
): Promise<void> {
  const viewSchemaId =
    options.entityType === "host" ? hostsViewSchemaId : identitiesViewSchemaId;
  await page.goto(
    `/?incident_id=${options.incidentId}&view_schema_id=${encodeURIComponent(
      viewSchemaId,
    )}`,
  );
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  if (options.expectAdminRole === true) {
    await expectCurrentIncidentRole(page, "Current incident role: admin");
  }
  await expect(
    page.getByTestId(gridRowTestId(viewSchemaId, options.survivorRecordId)),
  ).toBeVisible();
  await expect(
    page.getByTestId(gridRowTestId(viewSchemaId, options.loserRecordId)),
  ).toBeVisible();
  const inspectButtonTestId = entityInspectButtonTestId(
    options.entityType,
    options.survivorRecordId,
  );
  await scrollGridTargetIntoView({
    page,
    surface: viewSchemaId,
    targetTestId: inspectButtonTestId,
  });
  await page.getByTestId(inspectButtonTestId).click();
  const inspectorTestId = entityInspectorTestId(options.entityType);
  await expect(page.getByTestId(inspectorTestId)).toContainText(
    options.survivorLabel,
  );
  await expect(
    page.getByTestId(entityMergeControlTestId("start")),
  ).toBeVisible();
  await page.getByTestId(entityMergeControlTestId("start")).click();
  await page
    .getByTestId(entityMergeControlTestId("loser-record"))
    .selectOption(options.loserRecordId);
  await page
    .getByTestId(entityMergeControlTestId("reason"))
    .fill(options.mergeReason);
  await expect(
    page.getByTestId(entityMergeControlTestId("plan")),
  ).toContainText(`Survivor ${options.survivorLabel}`);
  await expect(
    page.getByTestId(entityMergeControlTestId("plan")),
  ).toContainText(`loser ${options.loserLabel}`);
  await expect(
    page.getByTestId(entityMergeControlTestId("plan")),
  ).toContainText(options.survivorRecordId);
  await expect(
    page.getByTestId(entityMergeControlTestId("plan")),
  ).toContainText(options.loserRecordId);
  await page.getByTestId(entityMergeControlTestId("review")).click();
  await expect(
    page.getByTestId(entityMergeControlTestId("cancel")),
  ).toBeFocused();
  const mergeResponsePromise = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response
        .url()
        .endsWith(`/api/v1/records/${options.survivorRecordId}/merge`),
  );
  const scrollport = page.locator(gridScrollportSelector());
  const scrollBefore = await scrollport.evaluate((node) => ({
    left: node.scrollLeft,
    top: node.scrollTop,
  }));
  await page.getByTestId(entityMergeControlTestId("confirm")).click();
  const mergeEnvelope = await readHttpOperationResponse(
    await mergeResponsePromise,
    "mergeEntityRecord",
  );
  expect(mergeEnvelope.data.survivor_record_id).toBe(options.survivorRecordId);
  expect(mergeEnvelope.data.loser_record_id).toBe(options.loserRecordId);
  expect(mergeEnvelope.data.merged_into_record_id).toBe(
    options.survivorRecordId,
  );
  await expect(
    page.getByTestId(gridRowTestId(viewSchemaId, options.loserRecordId)),
  ).toHaveCount(0);
  await expect(page.getByTestId(inspectorTestId)).toContainText(
    options.survivorLabel,
  );
  await expect(
    page.getByTestId(entityMergeControlTestId("message")),
  ).toContainText(
    `Merged ${options.loserLabel} (${options.loserRecordId}) into ${options.survivorLabel} (${options.survivorRecordId})`,
  );
  await expect(
    page.getByTestId(timelinePreviewRowTestId(options.dependentRecordId)),
  ).toBeVisible();
  await expect(
    page
      .getByTestId(timelinePreviewRowTestId(options.dependentRecordId))
      .getByRole("note", {
        name: `Resolved ${options.entityType}: ${options.resolvedLabel}`,
        exact: true,
      }),
  ).toBeVisible();
  await expect(
    page.getByRole("region", { name: "Entity merge review", exact: true }),
  ).toBeFocused();
  expect(
    await scrollport.evaluate((node) => ({
      left: node.scrollLeft,
      top: node.scrollTop,
    })),
  ).toEqual(scrollBefore);
}

async function expectCurrentIncidentRole(page: Page, roleText: string) {
  const accountMenuTrigger = page.getByRole("button", {
    name: "Account and application navigation",
  });
  await accountMenuTrigger.click();
  await expect(page.getByTestId(currentIncidentRoleTestId())).toHaveText(
    roleText,
  );
  await accountMenuTrigger.click();
}
