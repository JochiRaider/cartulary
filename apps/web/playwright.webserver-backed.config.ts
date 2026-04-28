import { defineConfig } from "@playwright/test";

import { webE2EBaseConfig } from "./playwright.shared.config";

const functionalGrep = process.env.CARTULARY_PLAYWRIGHT_FUNCTIONAL_GREP;
const functionalFiles = process.env.CARTULARY_PLAYWRIGHT_FUNCTIONAL_FILES;
const supportGrep = process.env.CARTULARY_PLAYWRIGHT_SUPPORT_GREP;
const supportFiles = process.env.CARTULARY_PLAYWRIGHT_SUPPORT_FILES;

if (!functionalGrep || functionalGrep.trim() === "") {
  throw new Error("CARTULARY_PLAYWRIGHT_FUNCTIONAL_GREP is required");
}

if (!functionalFiles || functionalFiles.trim() === "") {
  throw new Error("CARTULARY_PLAYWRIGHT_FUNCTIONAL_FILES is required");
}

if (!supportGrep || supportGrep.trim() === "") {
  throw new Error("CARTULARY_PLAYWRIGHT_SUPPORT_GREP is required");
}

if (!supportFiles || supportFiles.trim() === "") {
  throw new Error("CARTULARY_PLAYWRIGHT_SUPPORT_FILES is required");
}

function normalizeFunctionalMatch(file: string) {
  return file
    .trim()
    .replace(/^apps\/web\/e2e\//u, "")
    .replace(/^e2e\//u, "");
}

export default defineConfig(webE2EBaseConfig({
  projects: [
    {
      name: "functional",
      grep: new RegExp(functionalGrep),
      testMatch: functionalFiles
        .split(/\r?\n/u)
        .map(normalizeFunctionalMatch)
        .filter((file) => file.length > 0),
    },
    {
      name: "support",
      grep: new RegExp(supportGrep),
      testMatch: supportFiles
        .split(/\r?\n/u)
        .map(normalizeFunctionalMatch)
        .filter((file) => file.length > 0),
    },
  ],
}));
