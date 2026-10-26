// biome-ignore lint/correctness/noUndeclaredDependencies: the workspace root pins the browser accessibility engine used by the Make-owned harness.
import AxeBuilder from "@axe-core/playwright";
import type { Page } from "@playwright/test";
import { expect, test } from "../../fixtures";

export async function expectIncidentHeadingAndHeaders(
  page: Page,
  gutterLabel?: string,
) {
  const heading = page.getByRole("heading", { level: 1 });
  await expect(heading).toHaveCount(1);
  const identity = page.getByRole("button", { name: /^Incident details:/ });
  await expect(heading.getByRole("button")).toHaveCount(1);
  await expect(heading).toHaveAccessibleName(
    (await identity.getAttribute("aria-label")) ?? "",
  );
  if (gutterLabel !== undefined) {
    const header = page.getByRole("columnheader", {
      name: gutterLabel,
      exact: true,
    });
    await expect(header).toBeVisible();
    await expect(header).toHaveText(gutterLabel);
  }
  const result = await new AxeBuilder({ page })
    .withRules(["empty-table-header", "page-has-heading-one"])
    .analyze();
  await test.info().attach("incident-heading-and-table-headers", {
    body: JSON.stringify(result),
    contentType: "application/json",
  });
  expect(result.violations).toEqual([]);
  expect(result.incomplete).toEqual([]);
  expect(result.passes.map((rule) => rule.id)).toContain(
    "page-has-heading-one",
  );
  if (gutterLabel !== undefined)
    expect(result.passes.map((rule) => rule.id)).toContain(
      "empty-table-header",
    );
}
