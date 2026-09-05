import { incidentLandingTestId } from "@cartulary/ui-contracts";
import { expect, type Locator, type Page } from "@playwright/test";
import { IncidentDirectory } from "../../pages/incidentDirectory";

export function responseBarrier() {
  let release = () => {};
  const promise = new Promise<void>((resolve) => {
    release = resolve;
  });
  return { promise, release };
}

export async function openCreationPresentation(page: Page) {
  await new IncidentDirectory(page).goto();
  // Keep authorized background fixtures deterministic without replacing the query.
  const queried = page.waitForResponse(
    (response) =>
      response.request().method() === "GET" &&
      new URL(response.url()).searchParams.get("search") ===
        "creation-presentation-no-existing-results",
  );
  await page
    .getByTestId(incidentLandingTestId("search"))
    .fill("creation-presentation-no-existing-results");
  await page.getByTestId(incidentLandingTestId("search")).press("Enter");
  await queried;
  await expect(
    page.getByTestId(incidentLandingTestId("empty-state")),
  ).toBeVisible();
  await expect(page.getByTestId(incidentLandingTestId("loading"))).toBeHidden();
  await new IncidentDirectory(page).openCreation();
  return page.getByRole("form", { name: "New incident" });
}

export async function expectCreationControlReachable(
  page: Page,
  control: Locator,
) {
  if (await control.evaluate((element) => element === document.activeElement)) {
    // Re-enter using native keyboard navigation after viewport/zoom changes.
    // Calling focus() on an already-focused element does not scroll it again.
    await page.keyboard.press("Shift+Tab");
    await page.keyboard.press("Tab");
  } else {
    await control.focus();
  }
  await expect(control).toBeFocused();
  await expect
    .poll(() =>
      control.evaluate((element) => {
        const box = element.getBoundingClientRect();
        return (
          box.left >= 0 &&
          box.right <= window.innerWidth + 1 &&
          box.top >= 0 &&
          box.bottom <= window.innerHeight + 1
        );
      }),
    )
    .toBe(true);
  expect(
    await page.evaluate(
      () =>
        document.documentElement.scrollWidth <=
        document.documentElement.clientWidth + 1,
    ),
  ).toBe(true);
}
