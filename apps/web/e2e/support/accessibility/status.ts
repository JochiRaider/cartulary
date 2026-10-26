import {
  currentIncidentRoleTestId,
  saveStateTestId,
} from "@cartulary/ui-contracts";
import type { Locator, Page } from "@playwright/test";
import { expect } from "../../fixtures";

const privateDiagnosticPatterns = [
  /bootstrap[_ -]?token/i,
  /secret[_ -]?base32/i,
  /PRIVATESECRET|ERRORSECRET|CREATESECRET|MFAREQUIREDSECRET/i,
  /otpauth/i,
  /request[_ -]?id/i,
  /\bstack\b/i,
  /\btraceback\b/i,
  /\/(?:home|var|tmp|usr|app|workspace)\//i,
  /\bselect\b[\s\S]{0,80}\bfrom\b/i,
  /\binsert\b[\s\S]{0,80}\binto\b/i,
  /\bupdate\b[\s\S]{0,80}\bset\b/i,
  /internal_path/i,
  /raw_request/i,
];

export async function expectCurrentIncidentRole(page: Page, roleText: string) {
  const accountMenuTrigger = page.getByRole("button", {
    name: "Account and application navigation",
  });
  await accountMenuTrigger.click();
  await expect(page.getByTestId(currentIncidentRoleTestId())).toContainText(
    roleText,
  );
  await accountMenuTrigger.click();
}

export async function expectWorkbookSavePresentation(page: Page) {
  const primary = page.getByTestId(saveStateTestId());
  await expect(primary).toBeVisible();
  await expect(primary).not.toHaveAttribute("role");
  await expect(primary).not.toHaveAttribute("aria-live");
  await expect(
    page.getByRole("status", { name: "Workbook save updates" }),
  ).toHaveCount(1);
  await expect(
    page.getByRole("alert", { name: "Workbook save conflicts" }),
  ).toHaveCount(1);
}

export async function expectStatusRole(
  locator: Locator,
  options: { readonly visible?: boolean } = {},
) {
  if (options.visible === false) {
    await expect(locator).toHaveCount(1);
  } else {
    await expect(locator).toBeVisible();
  }
  await expect(locator).toHaveAttribute("role", "status");
  await expect(locator).not.toHaveText("");
}

export async function expectAlertRole(locator: Locator) {
  await expect(locator).toBeVisible();
  await expect(locator).toHaveAttribute("role", "alert");
  await expect(locator).not.toHaveText("");
}

export async function expectNoPrivateDiagnostics(locator: Locator) {
  const text = (await locator.textContent()) ?? "";
  for (const pattern of privateDiagnosticPatterns) {
    expect(text).not.toMatch(pattern);
  }
}
