import type { Page } from "@playwright/test";

export async function injectDesignFixture(
  page: Page,
  options: {
    ariaLabel: string;
    fixtureName: string;
    html: string;
    missingMainMessage: string;
    styleText: string;
  },
) {
  await page.evaluate((fixtureOptions) => {
    document
      .querySelector(
        `style[data-design-fixture-style='${fixtureOptions.fixtureName}']`,
      )
      ?.remove();
    document
      .querySelector(`[data-design-fixture='${fixtureOptions.fixtureName}']`)
      ?.remove();

    const main = document.querySelector("main.cartulary-shell");
    if (!(main instanceof HTMLElement)) {
      throw new Error(fixtureOptions.missingMainMessage);
    }

    const style = document.createElement("style");
    style.dataset.designFixtureStyle = fixtureOptions.fixtureName;
    style.textContent = fixtureOptions.styleText;
    document.head.append(style);

    const fixture = document.createElement("section");
    fixture.dataset.designFixture = fixtureOptions.fixtureName;
    fixture.setAttribute("aria-label", fixtureOptions.ariaLabel);
    fixture.innerHTML = fixtureOptions.html;
    main.append(fixture);
  }, options);
}
