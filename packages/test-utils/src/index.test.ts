// @vitest-environment jsdom

import { afterEach, vi } from "vitest";

import { registerActionSuite } from "./test-suites/actions";
import { registerContinuitySuite } from "./test-suites/continuity";
import { registerFacadeSuite } from "./test-suites/facade";
import { registerGroupingSuite } from "./test-suites/grouping";
import { registerMarkerSuite } from "./test-suites/marker";
import { registerVirtualTargetingSuite } from "./test-suites/virtual-targeting";

afterEach(() => {
  vi.useRealTimers();
  vi.restoreAllMocks();
  document.body.innerHTML = "";
});

registerFacadeSuite();
registerActionSuite();
registerGroupingSuite();
registerContinuitySuite();
registerVirtualTargetingSuite();
registerMarkerSuite();
