export type BrowserLocator = {
  blur?: () => Promise<void>;
  click: () => Promise<void>;
  dispatchEvent?: (type: string) => Promise<void>;
  evaluate?: (
    pageFunction: (element: Element, arg?: unknown) => unknown,
    arg?: unknown,
  ) => Promise<unknown>;
  fill: (value: string) => Promise<void>;
  isVisible?: () => Promise<boolean>;
  press?: (value: string) => Promise<void>;
  scrollIntoViewIfNeeded?: () => Promise<void>;
  selectOption?: (value: string | readonly string[]) => Promise<unknown>;
};

export type BrowserResponseLike = {
  ok: () => boolean;
  status?: () => number;
};

export type BrowserNetworkRequestLike = {
  method: () => string;
  postData?: () => string | null;
  postDataJSON?: () => unknown;
  url: () => string;
};

export type BrowserNetworkResponseLike = BrowserResponseLike & {
  json?: () => Promise<unknown>;
  request: () => BrowserNetworkRequestLike;
  url: () => string;
};

export type BrowserRequestLike = {
  post: (
    url: string,
    options: {
      data?: unknown;
      headers?: Record<string, string>;
    },
  ) => Promise<BrowserResponseLike>;
};

export type BrowserPageLike = {
  evaluate?: (
    pageFunction: (arg?: unknown) => unknown,
    arg?: unknown,
  ) => Promise<unknown>;
  getByTestId: (value: string) => BrowserLocator;
  request?: BrowserRequestLike;
  waitForRequest?: (
    predicate: (request: BrowserNetworkRequestLike) => boolean,
  ) => Promise<BrowserNetworkRequestLike>;
  waitForResponse?: (
    predicate: (response: BrowserNetworkResponseLike) => boolean,
  ) => Promise<BrowserNetworkResponseLike>;
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

export function requirePress(locator: BrowserLocator, value: string) {
  if (typeof locator.press !== "function") {
    throw new Error(`Grid browser-command ${value} requires locator.press()`);
  }
  return locator.press(value);
}

export function requireBlur(locator: BrowserLocator) {
  if (typeof locator.blur !== "function") {
    throw new Error("Grid browser-command blur requires locator.blur()");
  }
  return locator.blur();
}

export function requireDispatchEvent(locator: BrowserLocator, type: string) {
  if (typeof locator.dispatchEvent !== "function") {
    throw new Error(
      `Grid browser-command ${type} requires locator.dispatchEvent()`,
    );
  }
  return locator.dispatchEvent(type);
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
