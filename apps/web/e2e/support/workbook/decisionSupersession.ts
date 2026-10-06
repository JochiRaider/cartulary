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
import { activateCommittedGridCell } from "./rowMutations";
export async function openDecisionReviewFixture(
  page: Page,
  navigate: (url: string) => Promise<unknown> = (url) => page.goto(url),
  decidedAt?: { target: string; replacement: string },
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
    ...(decidedAt ? { "decision.decided_at": decidedAt.target } : {}),
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
      ...(decidedAt ? { "decision.decided_at": decidedAt.replacement } : {}),
    },
  );
  if (decidedAt) {
    expect(target.cells["decision.decided_at"]?.value).toBe(decidedAt.target);
    expect(replacement.cells["decision.decided_at"]?.value).toBe(
      decidedAt.replacement,
    );
  }
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
  await activateCommittedGridCell(cell);
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
  const viewport = page.viewportSize();
  if (!viewport) throw new Error("Decision control viewport missing");
  // Recovery completion can replace layout while native focus/scroll settles.
  // Observe one frame atomically; do not repair focus or scroll during retries.
  await expect(async () => {
    const geometry = await control.evaluate((element) => {
      const box = element.getBoundingClientRect();
      return {
        focused: document.activeElement === element,
        left: box.left,
        right: box.right,
        top: box.top,
        bottom: box.bottom,
        documentWidth: document.documentElement.scrollWidth,
        documentClientWidth: document.documentElement.clientWidth,
      };
    });
    expect(geometry.focused).toBe(true);
    expect(geometry.left).toBeGreaterThanOrEqual(-1);
    expect(geometry.right, JSON.stringify(geometry)).toBeLessThanOrEqual(
      viewport.width + 1,
    );
    expect(geometry.top).toBeGreaterThanOrEqual(-1);
    expect(geometry.bottom).toBeLessThanOrEqual(viewport.height + 1);
    expect(geometry.documentWidth).toBeLessThanOrEqual(
      geometry.documentClientWidth + 1,
    );
  }).toPass({ timeout: 5_000 });
}
