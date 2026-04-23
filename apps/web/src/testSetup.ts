import { afterEach, vi } from "vitest";

function cssEscape(value: string) {
  return String(value).replace(/[^a-zA-Z0-9_-]/g, (char) => `\\${char}`);
}

afterEach(() => {
  vi.useRealTimers();
  vi.clearAllTimers();
});

if (globalThis.CSS === undefined) {
  Object.defineProperty(globalThis, "CSS", {
    configurable: true,
    value: { escape: cssEscape },
  });
} else if (typeof globalThis.CSS.escape !== "function") {
  Object.defineProperty(globalThis.CSS, "escape", {
    configurable: true,
    value: cssEscape,
  });
}
