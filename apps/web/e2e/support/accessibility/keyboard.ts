import type { Locator, Page } from "@playwright/test";
import { expect } from "@playwright/test";
import { expectFocusContrast } from "./assertions";

const keyboardSentinelId = "a11y-keyboard-sentinel";

async function activeTestId(page: Page) {
  return page.evaluate(() => {
    const active = document.activeElement;
    if (!(active instanceof HTMLElement)) {
      return "";
    }
    return (
      active.getAttribute("data-testid") ??
      active.querySelector("[data-testid]")?.getAttribute("data-testid") ??
      active.closest("[data-testid]")?.getAttribute("data-testid") ??
      ""
    );
  });
}

async function activeElementOwnTestId(page: Page) {
  return page.evaluate(() => {
    const active = document.activeElement;
    if (!(active instanceof HTMLElement)) {
      return "";
    }
    return active.getAttribute("data-testid") ?? "";
  });
}

async function activeElementSignature(page: Page) {
  return page.evaluate(() => {
    const active = document.activeElement;
    if (!(active instanceof HTMLElement) || active === document.body) {
      return "";
    }
    const testId =
      active.closest("[data-testid]")?.getAttribute("data-testid") ?? "";
    const id = active.id === "" ? "" : `#${active.id}`;
    const tag = active.tagName.toLowerCase();
    const role = active.getAttribute("role") ?? "";
    const nameParts = [
      active.getAttribute("aria-label")?.trim() ?? "",
      active.getAttribute("title")?.trim() ?? "",
      active.textContent?.trim() ?? "",
    ].filter(Boolean);
    return [
      tag,
      id,
      testId === "" ? "" : ` data-testid:${testId}`,
      role === "" ? "" : ` role:${role}`,
      nameParts.length === 0 ? "" : `:${nameParts.join(" ")}`,
    ].join("");
  });
}

export async function blurActiveElement(page: Page) {
  await page.evaluate(() => {
    const active = document.activeElement;
    if (active instanceof HTMLElement) {
      active.blur();
    }
  });
}

async function focusKeyboardSentinel(page: Page) {
  await page.evaluate((sentinelId) => {
    let sentinel = document.getElementById(sentinelId);
    if (sentinel === null) {
      sentinel = document.createElement("button");
      sentinel.id = sentinelId;
      sentinel.setAttribute("aria-label", "Keyboard traversal sentinel");
      sentinel.setAttribute("type", "button");
      sentinel.style.position = "fixed";
      sentinel.style.inlineSize = "1px";
      sentinel.style.blockSize = "1px";
      sentinel.style.opacity = "0";
      sentinel.style.pointerEvents = "none";
      document.body.prepend(sentinel);
    }
    sentinel.focus();
  }, keyboardSentinelId);
}

async function removeKeyboardSentinel(page: Page) {
  await page.evaluate((sentinelId) => {
    document.getElementById(sentinelId)?.remove();
  }, keyboardSentinelId);
}

export async function expectTabOrderIncludes(
  page: Page,
  testIds: readonly (string | RegExp)[],
  maxTabs = 180,
) {
  await focusKeyboardSentinel(page);
  try {
    const remaining = [...testIds];
    for (let index = 0; index < maxTabs && remaining.length > 0; index += 1) {
      await page.keyboard.press("Tab");
      const active = await activeTestId(page);
      const reachedIndex = remaining.findIndex((target) =>
        typeof target === "string" ? target === active : target.test(active),
      );
      if (reachedIndex >= 0) {
        remaining.splice(reachedIndex, 1);
      }
    }
    expect(remaining.map(String)).toEqual([]);
  } finally {
    await removeKeyboardSentinel(page);
  }
}

export async function expectNoFocusTrap(page: Page) {
  await focusKeyboardSentinel(page);
  try {
    await page.keyboard.press("Tab");
    const first = await activeElementSignature(page);
    await page.keyboard.press("Tab");
    const second = await activeElementSignature(page);
    expect(first).not.toBe("");
    expect(second).not.toBe("");
    expect(second).not.toBe(first);
    await page.keyboard.press("Shift+Tab");
    expect(await activeElementSignature(page)).toBe(first);
  } finally {
    await removeKeyboardSentinel(page);
  }
}

export async function expectTabTraversalAdvancesFrom(
  page: Page,
  origin: Locator,
  tabCount = 4,
  backwards = false,
) {
  await expectVisibleFocus(origin);
  const visited: string[] = [];
  for (let index = 0; index < tabCount; index += 1) {
    await page.keyboard.press(backwards ? "Shift+Tab" : "Tab");
    visited.push(await activeElementSignature(page));
  }
  expect(visited.every((signature) => signature !== "")).toBeTruthy();
  expect(new Set(visited).size).toBeGreaterThan(1);
}

export async function expectKeyboardFocusReachesTestId(
  page: Page,
  testId: string,
  maxTabs = 20,
) {
  const visited: string[] = [];
  for (let index = 0; index < maxTabs; index += 1) {
    const current = await activeElementOwnTestId(page);
    if (current === testId) {
      return;
    }
    visited.push(current);
    await page.keyboard.press("Tab");
  }
  const current = await activeElementOwnTestId(page);
  if (current === testId) {
    return;
  }
  visited.push(current);
  expect(visited).toContain(testId);
}

export async function expectVisibleFocus(locator: Locator) {
  await expect(locator).toHaveCount(1);
  await expect(locator).toBeVisible();
  await expect
    .poll(async () => {
      await locator.focus();
      return locator.evaluate((element) => element === document.activeElement);
    })
    .toBeTruthy();
  await expect
    .poll(() =>
      locator.evaluate((element) => {
        const style = window.getComputedStyle(element);
        const outlineVisible =
          style.outlineStyle !== "none" &&
          style.outlineWidth !== "0px" &&
          style.outlineColor !== "transparent";
        const shadowVisible = style.boxShadow !== "none";
        return outlineVisible || shadowVisible;
      }),
    )
    .toBeTruthy();
  await expectFocusContrast(locator);
}
