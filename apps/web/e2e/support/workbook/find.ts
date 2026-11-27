import type { Locator } from "@playwright/test";
import { expect } from "../../fixtures";

/**
 * Moves Workbook Find to its next or previous match and waits for the panel
 * to collapse before the match takes focus.
 *
 * Navigation is unavailable while Find recomputes (design Workbook Find
 * presentation). A saved edit's follow-up view read can replace the loaded
 * rows after the status already reports the matches, and the rescan that
 * follows ignores a keystroke that lands during it. An ignored keystroke has
 * no effect, so repeat it until navigation collapses the panel.
 */
export async function navigateFind(
  input: Locator,
  direction: "Enter" | "Shift+Enter" = "Enter",
) {
  await expect(async () => {
    if ((await input.count()) > 0)
      await input.press(direction, { timeout: 1_000 });
    await expect(input).toHaveCount(0, { timeout: 2_000 });
  }).toPass({ timeout: 10_000 });
}
