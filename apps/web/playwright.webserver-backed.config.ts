import { defineConfig } from "@playwright/test";

import { webE2EBaseConfig } from "./playwright.shared.config";

const functionalGrep = process.env.CARTULARY_PLAYWRIGHT_FUNCTIONAL_GREP;

if (!functionalGrep || functionalGrep.trim() === "") {
  throw new Error("CARTULARY_PLAYWRIGHT_FUNCTIONAL_GREP is required");
}

export default defineConfig(webE2EBaseConfig({
  projects: [
    {
      name: "functional",
      grep: new RegExp(functionalGrep),
      testMatch: [
        "phase1.spec.ts",
        "phase2.spec.ts",
        "phase3.spec.ts",
        "phase4.spec.ts",
      ],
    },
    {
      name: "support",
      testMatch: ["phase2.support.spec.ts", "phase3.support.spec.ts"],
    },
  ],
}));
