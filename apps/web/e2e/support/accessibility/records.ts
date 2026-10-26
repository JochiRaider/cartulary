import { mkdirSync, writeFileSync } from "node:fs";
import path from "node:path";
import type { Locator, Page } from "@playwright/test";
import { expect, test } from "@playwright/test";
import {
  type ContrastTarget,
  collectContrastChecks,
  withContrastTarget,
} from "./contrast";

function contrastRecordPath(title: string) {
  const dir = process.env.CARTULARY_FRONTEND_ACCESSIBILITY_CONTRAST_DIR;
  if (!dir) {
    return "";
  }
  const slug = title
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-|-$/g, "")
    .slice(0, 120);
  mkdirSync(dir, { recursive: true });
  return path.join(dir, `${slug}.json`);
}

export async function expectAndRecordContrast(
  page: Page,
  testIds: readonly ContrastTarget[],
  context = "",
) {
  const title = test.info().title;
  const checks = await collectContrastChecks(page, testIds);

  const recordPath = contrastRecordPath(
    context ? `${context}-${title}` : title,
  );
  if (recordPath) {
    writeFileSync(
      recordPath,
      `${JSON.stringify(
        {
          scenario_title: title,
          ...(context ? { context } : {}),
          checks,
        },
        null,
        2,
      )}\n`,
      "utf8",
    );
  }
  expect(checks.length).toBeGreaterThan(0);
  const required = testIds.filter(
    (target) => typeof target !== "string" && target.required,
  );
  expect(
    checks.filter((check) =>
      required.some(
        (target) => typeof target !== "string" && target.id === check?.target,
      ),
    ),
  ).toHaveLength(required.length);
  expect(
    checks.filter((check) => check?.result !== "pass"),
    context,
  ).toEqual([]);
}

export async function expectLocatorContrast(
  locator: Locator,
  id: string,
  options: Omit<Exclude<ContrastTarget, string>, "id" | "selector"> = {},
) {
  await withContrastTarget(locator, (selector) =>
    expectAndRecordContrast(locator.page(), [{ id, selector, ...options }], id),
  );
}
