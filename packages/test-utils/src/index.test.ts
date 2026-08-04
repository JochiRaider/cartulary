// @vitest-environment jsdom

import { afterEach, vi } from "vitest";

import { registerContinuitySuite } from "./test-suites/continuity";
import { registerMarkerSuite } from "./test-suites/marker";
import { registerSelectorGroupingSuite } from "./test-suites/selector-grouping";
import { registerVirtualTargetingSuite } from "./test-suites/virtual-targeting";

afterEach(() => {
  vi.useRealTimers();
  vi.restoreAllMocks();
  document.body.innerHTML = "";
});

registerSelectorGroupingSuite();
registerContinuitySuite();
registerVirtualTargetingSuite();
registerMarkerSuite();
