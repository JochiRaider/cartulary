import { rmSync } from "node:fs";
import path from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";
import { describe, expect, test } from "vitest";

type FontBundleModule = {
  checkFontBundle: (root?: string) => string[];
  createFontBundleFixture: (options?: Record<string, boolean>) => string;
};

const repoRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "..",
  "..",
  "..",
);

async function loadFontBundleModule(): Promise<FontBundleModule> {
  return (await import(
    pathToFileURL(path.join(repoRoot, "scripts/check-font-bundle.mjs")).href
  )) as FontBundleModule;
}

describe("vendored font bundle", () => {
  test("validates the checked-in manifest, checksums, notice, CSS, and source metadata", async () => {
    const { checkFontBundle } = await loadFontBundleModule();

    expect(checkFontBundle(repoRoot)).toEqual([]);
  });

  test.each([
    ["badHash", "sha256 mismatch"],
    ["badBytes", "byte size mismatch"],
    ["missingLicense", "missing LICENSE.txt or OFL.txt"],
    ["missingNotice", "missing file inter/InterVariable.woff2"],
    ["localSource", "must not use local(...)"],
    ["remoteFont", "must not reference a remote font CDN"],
  ])("rejects %s fixture", async (option, expectedMessage) => {
    const { checkFontBundle, createFontBundleFixture } =
      await loadFontBundleModule();
    const fixtureRoot = createFontBundleFixture({ [option]: true });
    try {
      expect(checkFontBundle(fixtureRoot).join("\n")).toContain(
        expectedMessage,
      );
    } finally {
      rmSync(fixtureRoot, { force: true, recursive: true });
    }
  });
});
