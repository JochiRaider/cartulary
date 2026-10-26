import { scrollGridTargetIntoView } from "@cartulary/test-utils/grid";
import {
  genericCreateFieldTestId,
  genericEditValueTestId,
  gridShellTestId,
  saveStateTestId,
  workbookInspectorToggleTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import { expect, type Locator, type Page } from "@playwright/test";

import {
  activateCandidateIdentities,
  openReferenceCandidates,
  selectReferenceCandidates,
} from "./references";
import {
  activateCommittedGridCell,
  openGenericInspectorForRecord,
  submitGenericEditAndWait,
} from "./rowMutations";

export async function activateSemanticGridCell(content: Locator) {
  const cell = content.locator("xpath=ancestor::*[@role='gridcell'][1]");
  await activateCommittedGridCell(cell);
  return cell;
}

export async function openGenericSurface(
  page: Page,
  incidentId: string,
  viewSchemaId: string,
) {
  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      viewSchemaId,
    )}`,
  );
  await expect(page.getByTestId(gridShellTestId(viewSchemaId))).toBeVisible();
  await page.getByTestId(workbookInspectorToggleTestId(viewSchemaId)).click();
}

export async function expectWorkbookSaved(page: Page) {
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
}

export async function setGenericCreateField(
  page: Page,
  fieldKey: string,
  value: string | string[],
) {
  const fieldTestId = genericCreateFieldTestId(fieldKey);
  await scrollGridTargetIntoView({
    page,
    surface: currentWorkbookSurface(page),
    targetTestId: fieldTestId,
  });
  const input = page.getByTestId(fieldTestId);
  if (
    await page
      .getByRole("region", { name: /^Choose /u })
      .or(page.getByRole("dialog", { name: /^Choose /u }))
      .isVisible()
  ) {
    await selectReferenceCandidates(page, fieldTestId, value);
    return;
  }
  const tagName = await input.evaluate((element) => element.tagName);
  const inputType = await input.getAttribute("type");
  if (inputType === "checkbox") {
    if (value === "true") {
      await input.check();
    } else {
      await input.uncheck();
    }
    return;
  }
  if (tagName === "DIV") {
    if (
      !(await input
        .getByRole("button", { name: "Apply references", exact: true })
        .count())
    )
      await input
        .getByRole("button", { name: /^Choose /u })
        .first()
        .click();
    const choices = input.locator(
      "[data-workbook-single-candidates], [data-workbook-multi-candidates]",
    );
    await activateCandidateIdentities(choices, value);
    await input
      .getByRole("button", { name: "Apply references", exact: true })
      .click();
    return;
  }
  if (tagName === "SELECT") {
    await input.selectOption(value);
    return;
  }
  await input.fill(Array.isArray(value) ? value.join("\n") : value);
}

function currentWorkbookSurface(page: Page): string {
  const viewSchemaId = new URL(page.url()).searchParams.get("view_schema_id");
  return viewSchemaId ?? timelineViewSchemaId;
}

export async function waitForReferenceCandidate(
  page: Page,
  testId: string,
  value: string,
  referenceView?: string,
) {
  const { candidates } = await openReferenceCandidates(
    page,
    testId,
    referenceView,
  );
  await expect(
    candidates.locator(`[value="${value}"], [value$=":${value}"]`),
  ).toHaveCount(1);
}

export async function editExtendedSurfaceCell(
  page: Page,
  viewSchemaId: string,
  recordId: string,
  fieldKey: string,
  value: string | string[],
) {
  await openGenericInspectorForRecord(page, viewSchemaId, recordId);
  await page.locator(`[data-inspector-edit-field="${fieldKey}"]`).click();
  const input = page.getByTestId(genericEditValueTestId(viewSchemaId));
  const tagName = await input.evaluate((element) => element.tagName);
  const inputType = await input.getAttribute("type");
  if (inputType === "checkbox") {
    if (value === "true") {
      await input.check();
    } else {
      await input.uncheck();
    }
    await submitGenericEditAndWait(page, viewSchemaId, recordId);
    return;
  }
  if (tagName === "SELECT") {
    if (typeof value === "string") {
      await expect(input.locator(`option[value="${value}"]`)).toHaveCount(1);
    }
    await input.selectOption(value);
  } else {
    await input.fill(Array.isArray(value) ? value.join("\n") : value);
  }
  await submitGenericEditAndWait(page, viewSchemaId, recordId);
}
