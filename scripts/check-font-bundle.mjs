#!/usr/bin/env node

import { pathToFileURL } from "node:url";

import {
  runFontBundleCheckCLI,
} from "../tools/harness/frontend/font-bundle-check-cli.mjs";

export {
  checkFontBundle,
  createFontBundleFixture,
  runFontBundleCheckCLI,
} from "../tools/harness/frontend/font-bundle-check-cli.mjs";

if (import.meta.url === pathToFileURL(process.argv[1]).href) {
  runFontBundleCheckCLI();
}
