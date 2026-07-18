import { defineConfig } from "@playwright/test";

import { webE2EBaseConfig } from "./playwright.shared.config";

export default defineConfig(
  webE2EBaseConfig({
    projects: [{ name: "chromium", use: { browserName: "chromium" } }],
  }),
);
