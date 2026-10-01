import { readFileSync } from "node:fs";
import { createRequire } from "node:module";
import path from "node:path";

const json = (file) => JSON.parse(readFileSync(file, "utf8"));

// Resolve from each consumer, including its private dependencies. Hoisting and
// package-manager store layout are deliberately outside this contract.
export function resolvePlaywrightPackages(root) {
  const pins = json(path.join(root, "tools/toolchain_pins.json")).ui_review;
  const expected = pins.playwright;
  if (pins["playwright-core"] !== expected) throw new Error("Playwright pins disagree");
  const locate = (from, name) => {
    const manifest = createRequire(from).resolve(`${name}/package.json`);
    const value = json(manifest);
    if (value.version !== expected) throw new Error(`resolved ${name} version mismatch`);
    return { manifest, directory: path.dirname(manifest), value };
  };
  const rootManifest = path.join(root, "package.json");
  const webManifest = path.join(root, "apps/web/package.json");
  for (const [file, names] of [[rootManifest, ["playwright", "playwright-core"]], [webManifest, ["@playwright/test"]]]) {
    for (const name of names) if (json(file).devDependencies?.[name] !== expected) throw new Error(`authored ${name} version mismatch`);
  }
  const playwright = locate(rootManifest, "playwright");
  const core = locate(rootManifest, "playwright-core");
  const test = locate(webManifest, "@playwright/test");
  const testPlaywright = locate(test.manifest, "playwright");
  const engines = [core, locate(playwright.manifest, "playwright-core"), locate(testPlaywright.manifest, "playwright-core")];
  for (const engine of engines) {
    const descriptor = json(path.join(engine.directory, "browsers.json")).browsers.find((entry) => entry.name === "chromium");
    if (!descriptor || engines.some((other) => {
      const actual = json(path.join(other.directory, "browsers.json")).browsers.find((entry) => entry.name === "chromium");
      return actual?.revision !== descriptor.revision || actual?.browserVersion !== descriptor.browserVersion;
    })) throw new Error("resolved Chromium descriptors disagree");
  }
  return { playwrightPath: playwright.directory, corePath: engines[1].directory, testPath: test.directory,
    version: expected, chromium: json(path.join(core.directory, "browsers.json")).browsers.find((entry) => entry.name === "chromium") };
}
