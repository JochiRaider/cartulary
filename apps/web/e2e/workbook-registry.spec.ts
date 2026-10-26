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
  evidenceViewSchemaId,
  findingsViewSchemaId,
  forensicKeywordsViewSchemaId,
  handoffViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  indicatorsViewSchemaId,
  investigativeQueriesViewSchemaId,
  lessonViewSchemaId,
  listWorkbookSurfaceContracts,
  notesViewSchemaId,
  partiesViewSchemaId,
  requireViewContract,
  statusReviewViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { openSystemSurfaceBySwitcher } from "./pages/workbookInspector";

import { createIncident } from "./support/incidents/fixtures";
import { apiBase } from "./support/runtime/configuration";
import { uniqueIncidentKey } from "./support/runtime/fixtureIdentity";

async function expectRequiredAndOptionalRegistryExposed(page: Page) {
  const expected = [
    ...requiredBaseViewSchemaIds,
    ...optionalStandardizedSurfaceIds,
  ].sort();
  expect(
    listWorkbookSurfaceContracts()
      .map((entry) => entry.viewSchemaId)
      .sort(),
  ).toEqual(expected);
  for (const viewSchemaId of expected)
    expect(requireViewContract(viewSchemaId).viewSchemaId).toBe(viewSchemaId);
  const ids = await fetchPublicViewSchemaIds(page);
  expect(ids).toEqual(
    [...requiredBaseViewSchemaIds, ...optionalStandardizedSurfaceIds].sort(),
  );
  for (const viewSchemaId of [
    ...requiredBaseViewSchemaIds,
    ...optionalStandardizedSurfaceIds,
  ]) {
    const member = await page.request.get(
      `${apiBase}/api/v1/view-schemas/${viewSchemaId}`,
    );
    expect(member.ok()).toBeTruthy();
    const body = (await member.json()) as {
      data: { view_schema_id: string; fields: Array<{ field_key: string }> };
    };
    expect(body.data.view_schema_id).toBe(viewSchemaId);
    expect(body.data.fields.length).toBeGreaterThan(0);
  }
  const hypotheses = await page.request.get(
    `${apiBase}/api/v1/view-schemas/cartulary.view.hypotheses.v1`,
  );
  expect(hypotheses.status()).toBe(404);
}

async function fetchPublicViewSchemaIds(page: Page) {
  const response = await page.request.get(`${apiBase}/api/v1/view-schemas`);
  expect(response.ok()).toBeTruthy();
  const body = (await response.json()) as {
    data: { view_schemas: Array<{ view_schema_id: string }> };
  };
  return body.data.view_schemas.map((schema) => schema.view_schema_id);
}

const requiredBaseViewSchemaIds = [
  assessmentsViewSchemaId,
  commLogViewSchemaId,
  decisionsViewSchemaId,
  evidenceViewSchemaId,
  handoffViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  indicatorsViewSchemaId,
  lessonViewSchemaId,
  notesViewSchemaId,
  partiesViewSchemaId,
  statusReviewViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
] as const;

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
  await expectRequiredAndOptionalRegistryExposed(page);

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
