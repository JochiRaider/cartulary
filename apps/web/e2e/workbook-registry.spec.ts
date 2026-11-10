import {
  gridShellTestId,
  surfaceTabTestId,
  systemViewSwitcherMenuTestId,
  systemViewSwitcherTriggerTestId,
} from "@cartulary/ui-contracts";
import {
  assessmentsViewSchemaId,
  commLogViewSchemaId,
  decisionsViewSchemaId,
  findingsViewSchemaId,
  forensicKeywordsViewSchemaId,
  handoffViewSchemaId,
  indicatorsViewSchemaId,
  investigativeQueriesViewSchemaId,
  lessonViewSchemaId,
  notesViewSchemaId,
  partiesViewSchemaId,
  statusReviewViewSchemaId,
  taskRequestsViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { openSystemSurfaceBySwitcher } from "./pages/workbookInspector";

import { createIncident } from "./support/incidents/fixtures";
import { uniqueIncidentKey } from "./support/runtime/fixtureIdentity";

const optionalStandardizedSurfaceIds = [
  findingsViewSchemaId,
  forensicKeywordsViewSchemaId,
  investigativeQueriesViewSchemaId,
] as const;

async function assertSwitcherMembership(page: Page) {
  const values = await systemViewSelectorValues(page);
  for (const viewSchemaId of [
    ...optionalStandardizedSurfaceIds,
    assessmentsViewSchemaId,
    commLogViewSchemaId,
    decisionsViewSchemaId,
    indicatorsViewSchemaId,
    handoffViewSchemaId,
    lessonViewSchemaId,
    partiesViewSchemaId,
    statusReviewViewSchemaId,
    taskRequestsViewSchemaId,
  ]) {
    expect(values).toContain(viewSchemaId);
  }
}

async function systemViewSelectorValues(page: Page) {
  await page.getByTestId(systemViewSwitcherTriggerTestId()).click();
  await expect(page.getByTestId(systemViewSwitcherMenuTestId())).toBeVisible();
  const values = await page
    .getByTestId(systemViewSwitcherMenuTestId())
    .locator("[data-view-schema-id]")
    .evaluateAll((options) =>
      options
        .map((option) => option.getAttribute("data-view-schema-id"))
        .filter((value): value is string => value !== null),
    );
  await page.keyboard.press("Escape");
  return values;
}

test("required registry identities stay canonical with optional additions", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("REGISTRY-IDENTITIES"),
    "Workbook inspector workbook-interaction registry identities",
  );

  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      notesViewSchemaId,
    )}`,
  );
  await expect(
    page.getByTestId(surfaceTabTestId(notesViewSchemaId)),
  ).toBeVisible();
  await expect(
    page.getByTestId(gridShellTestId(notesViewSchemaId)),
  ).toBeVisible();
  await expect(page).toHaveURL(
    new RegExp(`view_schema_id=${encodeURIComponent(notesViewSchemaId)}`),
  );

  await openSystemSurfaceBySwitcher(page, indicatorsViewSchemaId, {
    actionTimeoutMs: 2500,
    attempts: 2,
    totalTimeoutMs: 7000,
  });
  await expect(
    page.getByTestId(gridShellTestId(indicatorsViewSchemaId)),
  ).toBeVisible();
  await expect(page).toHaveURL(
    new RegExp(`view_schema_id=${encodeURIComponent(indicatorsViewSchemaId)}`),
  );

  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      findingsViewSchemaId,
    )}`,
  );
  await expect(
    page.getByTestId(gridShellTestId(findingsViewSchemaId)),
  ).toBeVisible();
  await expect(page).toHaveURL(
    new RegExp(`view_schema_id=${encodeURIComponent(findingsViewSchemaId)}`),
  );

  await assertSwitcherMembership(page);
});
