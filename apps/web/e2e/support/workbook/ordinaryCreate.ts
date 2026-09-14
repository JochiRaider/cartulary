import { scrollGridTargetIntoView } from "@cartulary/test-utils/grid";
import {
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  gridShellTestId,
  surfaceTabTestId,
  workbookColumnsMenuTestId,
  workbookColumnsMenuTriggerTestId,
  workbookInspectorToggleTestId,
} from "@cartulary/ui-contracts";
import { requireViewContract } from "@cartulary/view-contracts";
import { expect, type Page } from "@playwright/test";
import { createIncident } from "../incidents/fixtures";
import { uniqueIncidentKey } from "../runtime/fixtureIdentity";

export const ordinaryEvidenceView = "cartulary.view.evidence.v1";
export async function openOrdinaryFixture(
  page: Page,
  view = ordinaryEvidenceView,
  navigate: (url: string) => Promise<unknown> = (url) => page.goto(url),
) {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("ORC"),
    "Ordinary workbook authoring",
  );
  await navigate(
    `/?incident_id=${incident}&view_schema_id=${encodeURIComponent(view)}`,
  );
  await expect(page.getByTestId(gridShellTestId(view))).toBeVisible();
  return { incident, view };
}
export async function switchOrdinarySheet(page: Page, view: string) {
  const tab = page.getByTestId(surfaceTabTestId(view));
  if (await tab.count()) await tab.click();
  else {
    await page
      .getByRole("button", { name: "System views", exact: true })
      .click();
    await page
      .locator(`[role="menuitemradio"][data-view-schema-id="${view}"]`)
      .click();
  }
  await expect(page.getByTestId(gridShellTestId(view))).toBeVisible();
}
export async function ordinaryField(page: Page, view: string, key: string) {
  const field = requireViewContract(view).fieldMap[key];
  if (!field?.createWritable) throw new Error(`Missing ordinary field ${key}`);
  const testId = genericCreateFieldTestId(key);
  if (!(await page.getByTestId(testId).count()) && field.defaultHidden) {
    if (
      view === "cartulary.view.hosts.v1" ||
      view === "cartulary.view.identities.v1"
    ) {
      await page.getByTestId(workbookColumnsMenuTriggerTestId(view)).click();
      const menu = page.getByTestId(workbookColumnsMenuTestId(view));
      const column = menu.getByRole("menuitemcheckbox", {
        name: field.label,
        exact: true,
      });
      if ((await column.getAttribute("aria-checked")) !== "true")
        await column.click();
      await page.getByTestId(workbookColumnsMenuTriggerTestId(view)).click();
    } else await page.getByTestId(workbookInspectorToggleTestId(view)).click();
  }
  if (!(await page.getByTestId(testId).count()))
    await scrollGridTargetIntoView({
      page,
      surface: view,
      targetTestId: testId,
    });
  return page.getByTestId(testId);
}
export async function fillOrdinaryField(
  page: Page,
  view: string,
  key: string,
  value: string,
) {
  const input = await ordinaryField(page, view, key);
  const field = requireViewContract(view).fieldMap[key];
  if (!field) throw new Error("Missing field");
  const tag = await input.evaluate((element) => element.tagName);
  if (tag === "DIV") {
    await input
      .getByRole("button", { name: /^Choose /u })
      .first()
      .click();
    await input
      .getByRole(
        field.writeKind === "action_payload" ? "listbox" : "combobox",
        { name: field.label, exact: true },
      )
      .selectOption(value);
    await input
      .getByRole("button", { name: "Apply references", exact: true })
      .click();
  } else if (tag === "SELECT") await input.selectOption(value);
  else if (field.readKind === "boolean")
    await input.setChecked(value === "true");
  else await input.fill(value);
  return input;
}
export async function commitOrdinary(page: Page, view: string) {
  const id = genericCreateSubmitTestId(view);
  if (!(await page.getByTestId(id).count()))
    await scrollGridTargetIntoView({ page, surface: view, targetTestId: id });
  const button = page.getByTestId(id);
  await button.click();
}
export async function retainOrdinaryUncertainty(
  page: Page,
  incident: string,
  view = ordinaryEvidenceView,
  malformed = false,
) {
  const bodies: string[] = [];
  await page.route(
    `**/incidents/${incident}/views/${view}/rows`,
    async (route) => {
      bodies.push(route.request().postData() ?? "");
      const response = await route.fetch();
      expect(response.ok()).toBe(true);
      if (bodies.length > 1) await route.fulfill({ response });
      else if (malformed)
        await route.fulfill({
          response,
          body: JSON.stringify({
            data: { row: {} },
            meta: { request_id: "broken-receipt" },
          }),
        });
      else await route.abort("failed");
    },
  );
  await commitOrdinary(page, view);
  const recovery = page.getByRole("region", {
    name: "Row creation",
    exact: true,
  });
  await expect(
    recovery.getByRole("button", { name: "Recover submission", exact: true }),
  ).toBeVisible();
  return { bodies, recovery };
}
