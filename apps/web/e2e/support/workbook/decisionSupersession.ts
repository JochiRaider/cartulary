import {
  decisionSupersessionTestId,
  gridShellTestId,
  rowCellTestId,
  workbookInspectorFeatureActionTestId,
  workbookInspectorToggleTestId,
} from "@cartulary/ui-contracts";
import { decisionsViewSchemaId } from "@cartulary/view-contracts";
import { expect, type Locator, type Page } from "@playwright/test";
import { createIncident } from "../incidents/fixtures";
import { uniqueIncidentKey, uniqueTxn } from "../runtime/fixtureIdentity";
import { createViewRow } from "./query";
export async function openDecisionReviewFixture(
  page: Page,
  navigate: (url: string) => Promise<unknown> = (url) => page.goto(url),
) {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("DECISION-REVIEW"),
    "Decision supersession review",
  );
  const target = await createViewRow(page, incidentId, decisionsViewSchemaId, {
    client_txn_id: uniqueTxn("target"),
    "decision.summary": "Isolate affected workstations",
    "decision.decision_type": "containment",
    "decision.status": "executed",
    "decision.rationale": "Initial containment evidence",
  });
  const replacement = await createViewRow(
    page,
    incidentId,
    decisionsViewSchemaId,
    {
      client_txn_id: uniqueTxn("replacement"),
      "decision.summary": "Isolate affected workstations",
      "decision.decision_type": "containment",
      "decision.status": "approved",
      "decision.rationale": "Reviewed containment evidence",
    },
  );
  await navigate(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(decisionsViewSchemaId)}`,
  );
  await expect(
    page.getByTestId(gridShellTestId(decisionsViewSchemaId)),
  ).toBeVisible();
  await page
    .getByTestId(workbookInspectorToggleTestId(decisionsViewSchemaId))
    .click();
  const cell = page
    .getByTestId(rowCellTestId(target.record_id, "decision.summary"))
    .locator("xpath=ancestor::*[@role='gridcell'][1]");
  await cell.dispatchEvent("mousedown", { button: 0 });
  await cell.focus();
  const start = page.getByTestId(
    workbookInspectorFeatureActionTestId(
      decisionsViewSchemaId,
      "decision.supersede",
    ),
  );
  await start.focus();
  await start.press("Enter");
  const replacementControl = page.getByTestId(
    decisionSupersessionTestId("replacement"),
  );
  await expect(
    replacementControl.locator(`option[value="${replacement.record_id}"]`),
  ).toHaveCount(1);
  await replacementControl.selectOption(replacement.record_id);
  await page
    .getByTestId(decisionSupersessionTestId("reason"))
    .fill(
      "Later evidence supports the revised containment scope.\nThe recorded execution remains part of the incident history.",
    );
  await page.getByTestId(decisionSupersessionTestId("review-action")).focus();
  await page.keyboard.press("Enter");
  await expect(
    page.getByTestId(decisionSupersessionTestId("review")),
  ).toContainText("remains executed");
  return { incidentId, target, replacement };
}
export async function expectDecisionControlReachable(
  page: Page,
  control: Locator,
) {
  await control.focus();
  await control.scrollIntoViewIfNeeded();
  await expect(control).toBeFocused();
  const box = await control.boundingBox(),
    viewport = page.viewportSize();
  if (!box || !viewport) throw new Error("Decision control viewport missing");
  expect(box.x).toBeGreaterThanOrEqual(-1);
  expect(box.x + box.width).toBeLessThanOrEqual(viewport.width + 1);
  expect(box.y).toBeGreaterThanOrEqual(-1);
  expect(box.y + box.height).toBeLessThanOrEqual(viewport.height + 1);
  expect(
    await page.evaluate(
      () =>
        document.documentElement.scrollWidth <=
        document.documentElement.clientWidth + 1,
    ),
  ).toBe(true);
}
