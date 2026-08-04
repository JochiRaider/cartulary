export type BrowserLocator = {
  click: () => Promise<void>;
  evaluate?: (
    pageFunction: (element: Element, arg?: unknown) => unknown,
    arg?: unknown,
  ) => Promise<unknown>;
  fill: (value: string) => Promise<void>;
  isVisible?: () => Promise<boolean>;
  selectOption?: (value: string | readonly string[]) => Promise<unknown>;
};

export type BrowserPageLike = {
  getByTestId: (value: string) => BrowserLocator;
};

export type BrowserEvaluate = NonNullable<BrowserLocator["evaluate"]>;

export function delay(durationMs: number) {
  return new Promise<void>((resolve) => {
    setTimeout(resolve, durationMs);
  });
}

export function requireEvaluate(
  locator: BrowserLocator,
  message: string,
): BrowserEvaluate {
  if (typeof locator.evaluate !== "function") {
    throw new Error(message);
  }
  return (pageFunction, arg) =>
    locator.evaluate?.(pageFunction, arg) as Promise<unknown>;
}

export function requireSelectOption(
  locator: BrowserLocator,
  message: string,
): NonNullable<BrowserLocator["selectOption"]> {
  if (typeof locator.selectOption !== "function") {
    throw new Error(message);
  }
  return (value) => locator.selectOption?.(value) as Promise<unknown>;
}

export async function isLocatorVisible(locator: BrowserLocator) {
  if (typeof locator.isVisible === "function") {
    return locator.isVisible();
  }
  try {
    const evaluate = requireEvaluate(
      locator,
      "isLocatorVisible requires locator.evaluate() support",
    );
    return Boolean(
      await evaluate((element) => {
        const rect = element.getBoundingClientRect();
        return (
          element.isConnected &&
          rect.width > 0 &&
          rect.height > 0 &&
          getComputedStyle(element).visibility !== "hidden"
        );
      }),
    );
  } catch {
    return false;
  }
}

export function supportsVisibilityCheck(locator: BrowserLocator) {
  return typeof locator.isVisible === "function";
}
