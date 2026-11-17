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
      path.join(
        repoRoot,
        "tools/harness/static-analysis/font-bundle-check-cli.mjs",
      ),
    ).href
  )) as FontBundleModule;
}

describe("vendored font bundle", () => {
  test("validates the checked-in manifest, checksums, notice, CSS, and source metadata", async () => {
    const { checkFontBundle, createFontBundleFixture } =
      await loadFontBundleModule();

    expect(checkFontBundle(repoRoot)).toEqual([]);
    // The unmutated fixture is the baseline every rejection case perturbs once.
    const baseline = createFontBundleFixture();
    try {
      expect(checkFontBundle(baseline)).toEqual([]);
    } finally {
      rmSync(baseline, { force: true, recursive: true });
    }
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
    ["badHash", ["sha256 mismatch"]],
    ["badBytes", ["byte size mismatch"]],
    ["missingLicense", ["missing LICENSE.txt or OFL.txt"]],
    ["localSource", ["must not use local(...)"]],
    ["remoteFont", ["must not reference a remote font CDN"]],
    [
      // Inter is the required default family, so dropping its activation
      // metadata also fails its required role and status checks.
      "missingActivationMetadata",
      [
        "Inter: activation_status is required",
        "Inter: role_ids must be non-empty",
        "Inter: activation_status must be active_default",
        "Inter: role_ids must include ui",
        "Inter: role_ids must include grid",
        "Inter: role_ids must include grid-cell",
      ],
    ],
    [
      "missingCssFontReference",
      ["manifest font file inter/InterVariable.woff2 must be referenced"],
    ],
    ["deadFontVariable", ["unused --font-unused"]],
    [
      // Staging the required default family also fails its default checks.
      "undocumentedStagedFamily",
      [
        "Inter: activation_status must be active_default",
        "Inter: role_ids must include ui",
        "Inter: role_ids must include grid",
        "Inter: role_ids must include grid-cell",
        "staged_inactive family Inter must declare staging_reason",
      ],
    ],
  ])("rejects %s fixture", async (option, expectedMessages) => {
    const { checkFontBundle, createFontBundleFixture } =
      await loadFontBundleModule();
    const fixtureRoot = createFontBundleFixture({ [option]: true });
    try {
      const errors = checkFontBundle(fixtureRoot);
      expect(errors, errors.join("\n")).toHaveLength(expectedMessages.length);
      for (const [index, message] of expectedMessages.entries())
        expect(errors[index]).toContain(message);
    } finally {
      rmSync(fixtureRoot, { force: true, recursive: true });
    }
  });
});
