import { Buffer } from "node:buffer";
import { createHash } from "node:crypto";
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import {
  cartularyDefaultThemeId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { test } from "@playwright/test";
import { visualPreferences } from "../auth/visualPreferences";
import { attestedVisualRendererProfile } from "../runtime/visualRenderer";
import { assertMetadataCaptureIdentity } from "./metadataNormalization";
import { assertVisualPresentation } from "./profile";

type FrontendVisualFixture = {
  capture_profiles: Record<
    string,
    {
      viewport_css_px: string;
      device_scale_factor: number;
      browser_zoom_percent: number;
      theme_id: string;
      density_id: string | null;
      surface_kind: string;
    }
  >;
  blocked_reason: string;
  capture_scope: { kind: string; selector?: string };
  design_contract_id?: string;
  dynamic_masks: string[];
  fixture_id: string;
  fixture_title: string;
  focus_state: Record<string, unknown>;
  editor_state: Record<string, unknown>;
  golden_artifacts: string[];
  golden_filename: string;
  inspector_state: Record<string, unknown>;
  no_dynamic_regions: boolean;
  catalog_row_ids: string[];
  playwright_scenario_title: string;
  replacement_fixture_id: string;
  scroll_normalization: { kind: string; anchor?: string; reason?: string };
  seed_id: string;
  status: string;
};

type FrontendVisualFixtureRegistry = {
  fixtures: FrontendVisualFixture[];
  owner_id: string;
  schema_id: string;
  verification_id: string;
};

function findRepoRoot(): string {
  let candidate = process.cwd();
  while (true) {
    if (
      existsSync(
        path.join(candidate, "tools", "frontend_visual_fixture_registry.json"),
      )
    ) {
      return candidate;
    }
    const parent = path.dirname(candidate);
    if (parent === candidate) {
      throw new Error(
        "could not find tools/frontend_visual_fixture_registry.json",
      );
    }
    candidate = parent;
  }
}

function repoPath(relativePath: string): string {
  return path.join(findRepoRoot(), relativePath);
}

function loadFrontendVisualFixtureRegistry(): FrontendVisualFixtureRegistry {
  return JSON.parse(
    readFileSync(
      repoPath("tools/frontend_visual_fixture_registry.json"),
      "utf8",
    ),
  ) as FrontendVisualFixtureRegistry;
}

export async function emitVisualCaptureIntent(
  page: Page,
  captureIntent: string,
  screenshotAssertionLocation: string,
) {
  const testInfo = test.info();
  const viewport = page.viewportSize();
  if (viewport === null) {
    throw new Error(`visual capture ${captureIntent} requires a viewport`);
  }
  const densityApplies =
    (await page.getByTestId(workbookShellReadyTestId()).count()) > 0;
  const surfaceDefault =
    densityApplies &&
    (await page
      .getByTestId(workbookShellReadyTestId())
      .getAttribute("data-active-view-schema-id")) === timelineViewSchemaId
      ? "compact"
      : "default";
  const expectedDensity = densityApplies
    ? (visualPreferences(page).read().density_mode ?? surfaceDefault)
    : null;
  const expectedTheme = cartularyDefaultThemeId;
  const browserProfile = await page.evaluate(() => {
    const zoom = Number.parseFloat(document.documentElement.style.zoom);
    return {
      browser_zoom_percent: Number.isFinite(zoom) ? Math.round(zoom) : 100,
      color_scheme: window.matchMedia("(prefers-color-scheme: dark)").matches
        ? "dark"
        : "light",
      density_id:
        document
          .querySelector("[data-cartulary-density]")
          ?.getAttribute("data-cartulary-density") ?? null,
      device_scale_factor: window.devicePixelRatio,
      reduced_motion: window.matchMedia("(prefers-reduced-motion: reduce)")
        .matches,
      theme_id:
        document
          .querySelector("[data-cartulary-theme]")
          ?.getAttribute("data-cartulary-theme") ?? null,
    };
  });
  const presentation = {
    surface_kind: densityApplies
      ? ("workbook_shell" as const)
      : ("application_shell" as const),
    density_id: expectedDensity,
    theme_id: expectedTheme,
  };
  assertVisualPresentation(presentation, browserProfile);
  const rendererProfile = attestedVisualRendererProfile();
  const rendererProfileId =
    process.env.CARTULARY_VISUAL_RENDERER_PROFILE_ID ?? "";
  if (
    process.env.CARTULARY_VISUAL_RENDERER_ATTESTED !== "1" ||
    rendererProfileId !== rendererProfile.profile_id ||
    page.context().browser()?.version() !== rendererProfile.chromium_version
  ) {
    throw new Error("visual capture requires the attested pinned renderer");
  }
  const snapshotPath = testInfo.snapshotPath(`${captureIntent}.png`);
  const expectedGoldenPath = process.env.CARTULARY_VISUAL_SNAPSHOT_ROOT
    ? `apps/web/e2e/visual-goldens/${path.basename(snapshotPath)}`
    : path.relative(findRepoRoot(), snapshotPath).replaceAll(path.sep, "/");
  for (const fixture of loadFrontendVisualFixtureRegistry().fixtures) {
    if (!fixture.golden_artifacts.includes(expectedGoldenPath)) continue;
    const registered = fixture.capture_profiles[expectedGoldenPath];
    const observed = {
      ...browserProfile,
      surface_kind: presentation.surface_kind,
      viewport_css_px: `${viewport.width}x${viewport.height}`,
    };
    if (
      !registered ||
      Object.entries(registered).some(
        ([key, value]) => observed[key as keyof typeof observed] !== value,
      )
    ) {
      const error = new Error(
        `Registered capture profile mismatch for ${fixture.fixture_id}: ${captureIntent}`,
      );
      error.name = "CartularyVisualCaptureError";
      throw error;
    }
  }
  const testFile = path
    .relative(findRepoRoot(), testInfo.file)
    .replaceAll(path.sep, "/");
  const selected: {
    row_id: string;
    scenario_id: string;
    file: string;
    title: string;
    project_id: string;
  }[] = JSON.parse(process.env.CARTULARY_VISUAL_SELECTED_CASES ?? "[]");
  const matches = selected.filter(
    (entry) =>
      entry.file === testFile &&
      entry.title === testInfo.title &&
      entry.project_id === testInfo.project.name,
  );
  const binding = matches[0];
  if (matches.length !== 1 || !binding)
    throw new Error(
      "visual capture requires exactly one selected catalog row/scenario",
    );
  const captureId = `visual.capture.${createHash("sha256")
    .update(
      JSON.stringify([
        binding.row_id,
        binding.scenario_id,
        testInfo.project.name,
        captureIntent,
      ]),
    )
    .digest("hex")
    .slice(0, 20)}`;
  assertMetadataCaptureIdentity(captureIntent, captureId, expectedGoldenPath);
  await testInfo.attach(`cartulary-visual-capture-intent-${captureId}.json`, {
    body: Buffer.from(
      `${JSON.stringify({
        schema_id: "cartulary.frontend_visual_capture_intent.v3",
        capture_id: captureId,
        row_id: binding.row_id,
        scenario_id: binding.scenario_id,
        capture_intent: captureIntent,
        expected_golden_path: expectedGoldenPath,
        project_id: testInfo.project.name,
        renderer_profile_id: rendererProfileId,
        screenshot_assertion_location: screenshotAssertionLocation,
        capture_profile: {
          ...browserProfile,
          surface_kind: presentation.surface_kind,
          expected_density_id: expectedDensity,
          expected_theme_id: expectedTheme,
          project_id: testInfo.project.name,
          snapshot_path_template:
            "{testDir}/visual-goldens/{arg}{-snapshotSuffix}{ext}",
          snapshot_suffix: testInfo.snapshotSuffix,
          viewport_css_px: `${viewport.width}x${viewport.height}`,
        },
        test_file: path
          .relative(findRepoRoot(), testInfo.file)
          .replaceAll(path.sep, "/"),
        test_title: testInfo.title,
      })}\n`,
      "utf8",
    ),
    contentType: "application/json",
  });
}
