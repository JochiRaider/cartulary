import path from "node:path";
import { defineConfig } from "@playwright/test";

import { webE2EBaseConfig } from "./playwright.shared.config";

export default defineConfig(
  webE2EBaseConfig({
    projects: [{ name: "chromium", use: { browserName: "chromium" } }],
    snapshotPathTemplate: process.env.CARTULARY_VISUAL_SNAPSHOT_ROOT
      ? path.join(
          process.env.CARTULARY_VISUAL_SNAPSHOT_ROOT,
          "{arg}{-snapshotSuffix}{ext}",
        )
      : "{snapshotDir}/{testFileDir}/{testFileName}-snapshots/{arg}{-snapshotSuffix}{ext}",
  }),
);
