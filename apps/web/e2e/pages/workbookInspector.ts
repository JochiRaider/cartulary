import {
  gridDraftRowSelector,
  gridSavedRowsSelector,
  gridShellTestId,
  systemViewSwitcherMenuTestId,
  systemViewSwitcherTriggerTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import type { Page } from "@playwright/test";

import { expect } from "@playwright/test";

export function gridSavedRows(page: Page, surface: string) {
  return page
    .getByTestId(gridShellTestId(surface))
    .locator(gridSavedRowsSelector());
}

export function gridDraftRows(page: Page, surface: string) {
  return page
    .getByTestId(gridShellTestId(surface))
    .locator(gridDraftRowSelector());
}

export async function openSystemSurfaceBySwitcher(
  page: Page,
  viewSchemaId: string,
  options: {
    actionTimeoutMs?: number;
    attempts?: number;
    totalTimeoutMs?: number;
  } = {},
) {
  const attempts = options.attempts ?? 3;
  const actionTimeoutMs = options.actionTimeoutMs ?? 2500;
  const totalTimeoutMs = options.totalTimeoutMs ?? 8000;
  const startedAt = Date.now();
  const deadline = startedAt + totalTimeoutMs;
  let lastError: unknown = null;
  const nextOperationTimeout = () => {
    const remainingMs = deadline - Date.now();
    if (remainingMs <= 0) {
      throw new Error(
        `system view switcher deadline expired after ${totalTimeoutMs}ms`,
      );
    }
    return Math.min(actionTimeoutMs, remainingMs);
  };

  for (let attempt = 1; attempt <= attempts; attempt += 1) {
    try {
      await page
        .getByTestId(systemViewSwitcherTriggerTestId())
        .click({ timeout: nextOperationTimeout() });
      const menu = page.getByTestId(systemViewSwitcherMenuTestId());
      await expect(menu).toBeVisible({ timeout: nextOperationTimeout() });
      const option = menu.locator(`[data-view-schema-id="${viewSchemaId}"]`);
      await expect(option).toHaveCount(1, {
        timeout: nextOperationTimeout(),
      });
      await option.click({ timeout: nextOperationTimeout() });
      await expect(
        page.getByTestId(workbookShellReadyTestId()),
      ).toHaveAttribute("data-active-view-schema-id", viewSchemaId, {
        timeout: nextOperationTimeout(),
      });
      await expect(page).toHaveURL(
        new RegExp(`view_schema_id=${encodeURIComponent(viewSchemaId)}`),
        { timeout: nextOperationTimeout() },
      );
      await expect(page.getByTestId(gridShellTestId(viewSchemaId))).toBeVisible(
        {
          timeout: nextOperationTimeout(),
        },
      );
      return;
    } catch (error) {
      lastError = error;
      await page.keyboard.press("Escape").catch(() => {});
      const remainingMs = deadline - Date.now();
      if (attempt >= attempts || remainingMs <= 0) {
        break;
      }
      await page
        .waitForTimeout(Math.min(100 * attempt, remainingMs))
        .catch(() => {});
    }
  }

  const shell = page.getByTestId(workbookShellReadyTestId());
  const trigger = page.getByTestId(systemViewSwitcherTriggerTestId());
  const menu = page.getByTestId(systemViewSwitcherMenuTestId());
  const diagnostics = {
    actionTimeoutMs,
    activeSurface: await shell
      .getAttribute("data-active-view-schema-id")
      .catch(() => null),
    currentUrl: page.url(),
    elapsedMs: Date.now() - startedAt,
    lastError:
      lastError instanceof Error ? lastError.message : String(lastError),
    menuAttached: (await menu.count().catch(() => 0)) > 0,
    menuVisible: await menu.isVisible().catch(() => false),
    requestedViewSchemaId: viewSchemaId,
    totalTimeoutMs,
    triggerEnabled: await trigger.isEnabled().catch(() => false),
    triggerVisible: await trigger.isVisible().catch(() => false),
    visibleOptionViewSchemaIds: await menu
      .locator("[data-view-schema-id]")
      .evaluateAll((options) =>
        options
          .map((option) => option.getAttribute("data-view-schema-id"))
          .filter((value): value is string => value !== null),
      )
      .catch(() => []),
  };
  throw new Error(
    `System view switcher did not open ${viewSchemaId}: ${JSON.stringify(
      diagnostics,
    )}`,
  );
}
