import { mkdir, writeFile } from "node:fs/promises";
import { createRequire } from "node:module";
import { dirname } from "node:path";
import { test as base } from "../fixtures";
import { attestedVisualRendererProfile } from "../support/runtime/visualRenderer";

const require = createRequire(import.meta.url);
const playwrightVersion: string =
  require("@playwright/test/package.json").version;

// Measurement-only evidence: ordinary browser suites do not acquire a browser
// or retain environment artifacts merely because they share the base fixtures.
export const test = base.extend<{ executionEnvironment: undefined }>({
  executionEnvironment: [
    async ({ browser, browserName }, use, info) => {
      if (browserName !== "chromium") {
        throw new Error("Measurement environment receipt requires Chromium");
      }
      const renderer = attestedVisualRendererProfile();
      if (browser.version() !== renderer.chromium_version) {
        throw new Error(
          "Measurement browser differs from its attested renderer",
        );
      }
      const session = await browser.newBrowserCDPSession();
      let browserMode: "headless" | "headed";
      try {
        // The browser-level response is independent of page/context UA overrides.
        // Chromium's headless shell does not expose getBrowserCommandLine.
        const version = await session.send("Browser.getVersion");
        const product = /^(HeadlessChrome|Chrome)\/(\d+\.\d+\.\d+\.\d+)$/u.exec(
          version.product,
        );
        const agent =
          /(?:^| )(HeadlessChrome|Chrome)\/\d+\.\d+\.\d+\.\d+(?: |$)/u.exec(
            version.userAgent,
          );
        if (
          !product ||
          !agent ||
          product[2] !== renderer.chromium_version ||
          (product[1] === "HeadlessChrome" && agent[1] !== "HeadlessChrome")
        ) {
          throw new Error(
            "Chromium did not provide an unambiguous measurement identity",
          );
        }
        browserMode = agent[1] === "HeadlessChrome" ? "headless" : "headed";
      } finally {
        await session.detach();
      }
      const receipt = {
        schema_id: "cartulary.browser_execution_environment.v1",
        browser_engine: browserName,
        browser_build: browser.version(),
        browser_mode: browserMode,
        playwright_version: playwrightVersion,
        node_version: process.version,
        platform: process.platform,
        architecture: process.arch,
      };
      const file = info.outputPath("browser-execution-environment.v1.json");
      const bytes = Buffer.from(`${JSON.stringify(receipt)}\n`);
      await mkdir(dirname(file), { recursive: true, mode: 0o700 });
      await writeFile(file, bytes, { flag: "wx", mode: 0o600 });
      await info.attach(receipt.schema_id, {
        body: bytes,
        contentType: "application/json",
      });
      await use(undefined);
    },
    { auto: true },
  ],
});

export { expect } from "../fixtures";
