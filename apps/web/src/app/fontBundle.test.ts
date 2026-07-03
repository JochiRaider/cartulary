import { readFileSync, rmSync } from "node:fs";
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
  "..",
);

async function loadFontBundleModule(): Promise<FontBundleModule> {
  return (await import(
    pathToFileURL(
      path.join(repoRoot, "tools/harness/frontend/font-bundle-check-cli.mjs"),
    ).href
  )) as FontBundleModule;
}

describe("vendored font bundle", () => {
  test("validates the checked-in manifest, checksums, notice, CSS, and source metadata", async () => {
    const { checkFontBundle } = await loadFontBundleModule();

    expect(checkFontBundle(repoRoot)).toEqual([]);
  });

  test("documents selector-active and staged font roles in the manifest", () => {
    const manifest = JSON.parse(
      readFileSync(
        path.join(repoRoot, "apps/web/public/assets/fonts/FONT_MANIFEST.json"),
        "utf8",
      ),
    ) as {
      families: Array<{
        activation_selectors?: string[];
        activation_status?: string;
        family: string;
        role_ids?: string[];
        staging_reason?: string;
      }>;
    };
    const familyByName = new Map(
      manifest.families.map((family) => [family.family, family]),
    );

    expect(familyByName.get("Inter")).toMatchObject({
      activation_status: "active_default",
      role_ids: expect.arrayContaining(["ui", "grid", "grid-cell"]),
    });
    expect(familyByName.get("JetBrains Mono")).toMatchObject({
      activation_status: "active_default",
      role_ids: expect.arrayContaining(["mono"]),
    });
    expect(familyByName.get("Atkinson Hyperlegible")).toMatchObject({
      activation_selectors: ['[data-reading-profile="hyperlegible"]'],
      activation_status: "active_selector",
      role_ids: ["accessible-reading"],
    });
    expect(familyByName.get("IBM Plex Sans Condensed")).toMatchObject({
      activation_selectors: ['[data-density-role="narrow-metadata"]'],
      activation_status: "active_selector",
      role_ids: ["compact-metadata"],
    });
    for (const familyName of ["Geist", "Geist Mono", "Source Serif 4"]) {
      expect(familyByName.get(familyName)).toMatchObject({
        activation_status: "staged_inactive",
        staging_reason: expect.any(String),
      });
    }
  });

  test.each([
    ["badHash", "sha256 mismatch"],
    ["badBytes", "byte size mismatch"],
    ["missingLicense", "missing LICENSE.txt or OFL.txt"],
    ["missingNotice", "missing file inter/InterVariable.woff2"],
    ["localSource", "must not use local(...)"],
    ["remoteFont", "must not reference a remote font CDN"],
    ["missingActivationMetadata", "activation_status is required"],
    [
      "missingCssFontReference",
      "manifest font file inter/InterVariable.woff2 must be referenced",
    ],
    ["deadFontVariable", "unused --font-unused"],
    [
      "undocumentedStagedFamily",
      "staged_inactive family Inter must declare staging_reason",
    ],
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
