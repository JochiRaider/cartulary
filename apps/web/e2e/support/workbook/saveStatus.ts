import { expect, type Page } from "@playwright/test";

type SaveEvent = { priority: string; message: string };
type ObservedWindow = Window & { workbookSaveEvents: SaveEvent[] };

export async function observeSaveEvents(page: Page) {
  await expect(
    page.getByRole("status", { name: "Workbook save updates" }),
  ).toHaveText("");
  await page.evaluate(() => {
    const state = window as unknown as ObservedWindow;
    state.workbookSaveEvents = [];
    for (const name of ["Workbook save updates", "Workbook save conflicts"]) {
      const region = document.querySelector(`[aria-label="${name}"]`);
      if (region === null)
        throw new Error(`Missing production save announcement host: ${name}`);
      let previous = region.textContent ?? "";
      new MutationObserver(() => {
        const message = region.textContent ?? "";
        if (message !== previous && message !== "")
          state.workbookSaveEvents.push({
            priority: region.getAttribute("aria-live") ?? "",
            message,
          });
        previous = message;
      }).observe(region, {
        childList: true,
        subtree: true,
        characterData: true,
      });
    }
  });
}

export function saveEvents(page: Page) {
  return page.evaluate(
    () => (window as unknown as ObservedWindow).workbookSaveEvents,
  );
}
