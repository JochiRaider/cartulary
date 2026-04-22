if (globalThis.ResizeObserver === undefined) {
  Object.defineProperty(globalThis, "ResizeObserver", {
    configurable: true,
    value: class ResizeObserver {
      disconnect() {}
      observe() {}
      unobserve() {}
    },
  });
}

if (typeof window !== "undefined") {
  Object.defineProperty(window, "focus", {
    configurable: true,
    value: () => {},
  });
}

if (
  typeof HTMLElement !== "undefined" &&
  typeof HTMLElement.prototype.scrollIntoView !== "function"
) {
  Object.defineProperty(HTMLElement.prototype, "scrollIntoView", {
    configurable: true,
    value: () => {},
  });
}
