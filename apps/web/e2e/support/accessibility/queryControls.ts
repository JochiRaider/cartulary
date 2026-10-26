import {
  gridShellTestId,
  type SystemViewSwitcherGroupToken,
  systemViewSwitcherMenuTestId,
  systemViewSwitcherOptionTestId,
  systemViewSwitcherTriggerTestId,
} from "@cartulary/ui-contracts";
import type { Page } from "@playwright/test";
import { expect } from "../../fixtures";
import { expectVisibleFocus } from "./keyboard";

export async function openA11ySystemSurface(
  page: Page,
  options: {
    groupToken: SystemViewSwitcherGroupToken;
    viewSchemaId: string;
  },
) {
  const trigger = page.getByTestId(systemViewSwitcherTriggerTestId());
  await expectVisibleFocus(trigger);
  await trigger.press("Enter");
  await expect(trigger).toHaveAttribute("aria-expanded", "true");
  const menu = page.getByTestId(systemViewSwitcherMenuTestId());
  await expect(menu).toBeVisible();
  await expect(menu).toHaveAttribute("role", "menu");

  const option = page.getByTestId(
    systemViewSwitcherOptionTestId(options.groupToken, options.viewSchemaId),
  );
  await expect(option).toHaveAttribute("role", "menuitemradio");
  await expect(option).toHaveAttribute(
    "data-view-schema-id",
    options.viewSchemaId,
  );
  await expect(option).not.toHaveText("");
  const menuOptions = menu.getByRole("menuitemradio");
  await expect
    .poll(() =>
      menuOptions.evaluateAll((entries) => {
        return entries.some((entry) => entry === document.activeElement);
      }),
    )
    .toBeTruthy();
  const optionCount = await menuOptions.count();
  const activeIndex = await menuOptions.evaluateAll((entries) => {
    const activeElement = document.activeElement;
    if (
      !(activeElement instanceof HTMLElement) &&
      !(activeElement instanceof SVGElement)
    ) {
      return -1;
    }
    return entries.indexOf(activeElement);
  });
  const targetIndex = await menuOptions.evaluateAll(
    (entries, viewSchemaId) =>
      entries.findIndex(
        (entry) => entry.getAttribute("data-view-schema-id") === viewSchemaId,
      ),
    options.viewSchemaId,
  );
  expect(activeIndex).toBeGreaterThanOrEqual(0);
  expect(targetIndex).toBeGreaterThanOrEqual(0);
  const arrowCount = (targetIndex - activeIndex + optionCount) % optionCount;
  for (let index = 1; index <= arrowCount; index += 1) {
    await page.keyboard.press("ArrowDown");
    await expect(
      menuOptions.nth((activeIndex + index) % optionCount),
    ).toBeFocused();
  }
  await expect(option).toBeFocused();
  await expectVisibleFocus(option);
  await option.press("Enter");
  await expect(menu).toHaveCount(0);
  await expect(trigger).toHaveAttribute("aria-expanded", "false");
  await expect(trigger).toHaveAttribute(
    "data-view-schema-id",
    options.viewSchemaId,
  );
  await expect(
    page.getByTestId(gridShellTestId(options.viewSchemaId)),
  ).toBeVisible();
}
