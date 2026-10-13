import type { Locator } from "@playwright/test";
import { expect, test } from "./fixtures";
import { activateCandidateIdentities } from "./support/workbook/references";

test("candidate activation waits for the requested control before choosing its interaction", async ({
  page,
}) => {
  for (const type of ["radio", "checkbox"] as const) {
    await page.setContent('<div id="candidates"></div>');
    const candidates = page.locator("#candidates");
    let requested!: () => void;
    const request = new Promise<void>((resolve) => {
      requested = resolve;
    });
    // Hold publication until the helper requests a candidate. The actual
    // selection uses Playwright and native DOM controls, with no timing sleep.
    const observed = new Proxy(candidates, {
      get(target, key) {
        if (key === "locator")
          return (...args: Parameters<Locator["locator"]>) => {
            requested();
            return target.locator(...args);
          };
        const value = Reflect.get(target, key);
        return typeof value === "function" ? value.bind(target) : value;
      },
    });
    const selecting = activateCandidateIdentities(observed, "chosen");
    await request;
    await candidates.evaluate((element, type) => {
      for (const value of ["other", "chosen"]) {
        const input = document.createElement("input");
        input.type = type;
        input.name = "candidate";
        input.value = value;
        element.append(input);
      }
    }, type);
    await selecting;
    await expect(candidates.locator('input[value="chosen"]')).toBeChecked();
    await expect(candidates.locator('input[value="other"]')).not.toBeChecked();
  }
});
