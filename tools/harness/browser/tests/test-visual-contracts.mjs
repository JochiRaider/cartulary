import assert from "node:assert/strict";
import { mkdirSync, mkdtempSync, readFileSync, rmSync, symlinkSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { classifyFrontendVisualGoldens, validateCapturePayload, validateCaptureAttribution, resolveRegisteredFixtures } from "../frontend-visual-reconciliation.mjs";
import { buildFrontendVisualGoldenManifest, validateFrontendVisualGoldenManifest } from "../frontend-visual-golden-manifest.mjs";
import { assertVisualRendererEnvironmentIsPrivate, loadVisualRendererProfile } from "../visual-renderer-lease.mjs";
import test from "node:test";
const root = path.resolve(import.meta.dirname, "../../../..");
test("visual renderer input is private and manifest matches exact PNG identities", async (t) => {
    const rendererProfile = loadVisualRendererProfile(root);
    assert.equal(rendererProfile.playwright_version, JSON.parse(readFileSync(path.join(root, "tools/toolchain_pins.json"), "utf8")).ui_review.playwright);
    assert.throws(() => assertVisualRendererEnvironmentIsPrivate({
        CARTULARY_VISUAL_RENDERER_WS_ENDPOINT: "ws://127.0.0.1:1/host-input",
    }), /harness-private/u, "visual renderer endpoints must not be inherited user inputs");
    assert.doesNotThrow(() => assertVisualRendererEnvironmentIsPrivate({}));
    const currentGoldenManifest = buildFrontendVisualGoldenManifest({ root });
    assert.deepEqual(validateFrontendVisualGoldenManifest(currentGoldenManifest, { root }), [], "generated golden manifest must exhaustively match committed PNG identities");
    assert.match(validateFrontendVisualGoldenManifest({
        ...currentGoldenManifest,
        goldens: currentGoldenManifest.goldens.map((golden, index) => index === 0 ? { ...golden, sha256: "0".repeat(64) } : golden),
    }, { root }).join("\n"), /do not match committed PNGs/u, "edited PNG identity must fail golden-manifest validation");
});
test("visual reconciliation classifies every golden and admits selected fixtures exactly", async (t) => {
    const capture = (captureID, goldenPath) => ({
        capture_id: captureID,
        capture_intent: captureID,
        owner_id: "harness.browser",
        row_id: `row.${captureID}`,
        scenario_id: `scenario.${captureID}`,
        project_id: "chromium",
        expected_golden_path: goldenPath,
    });
    const classifiedGoldens = classifyFrontendVisualGoldens({
        captureIntents: [
            capture("active", "snapshots/active.png"),
            capture("missing", "snapshots/missing.png"),
            capture("ambiguous-a", "snapshots/ambiguous.png"),
            capture("ambiguous-b", "snapshots/ambiguous.png"),
        ],
        committedGoldens: new Map([
            ["snapshots/active.png", "a".repeat(64)],
            ["snapshots/ambiguous.png", "b".repeat(64)],
            ["snapshots/orphan.png", "c".repeat(64)],
        ]),
        fixtures: [
            {
                fixture_id: "visual.fixture.active",
                golden_artifacts: ["snapshots/active.png"],
            },
        ],
    });
    assert.deepEqual(Object.fromEntries(classifiedGoldens.map((golden) => [
        golden.golden_path,
        golden.classification,
    ])), {
        "snapshots/active.png": "active",
        "snapshots/ambiguous.png": "ambiguous_mapping",
        "snapshots/missing.png": "missing_golden",
        "snapshots/orphan.png": "orphan",
    }, "visual reconciliation must distinguish active, ambiguous, missing, and orphan goldens");
    const fixtureCatalogEntries = [
        {
            row_id: "row.selected",
            title: "Selected fixture",
        },
        {
            row_id: "row.unselected",
            title: "Unselected fixture",
        },
    ];
    const fixtureGoldens = [
        {
            golden_path: "snapshots/selected.png",
            sha256: "d".repeat(64),
            catalog_row_ids: ["row.selected"],
        },
        {
            golden_path: "snapshots/unselected.png",
            sha256: "e".repeat(64),
            catalog_row_ids: [],
        },
    ];
    assert.deepEqual(resolveRegisteredFixtures([
        {
            fixture_id: "visual.fixture.selected",
            catalog_row_ids: ["row.selected"],
            playwright_scenario_title: "Selected fixture",
            golden_artifacts: ["snapshots/selected.png"],
        },
        {
            fixture_id: "visual.fixture.unselected",
            catalog_row_ids: ["row.unselected"],
            playwright_scenario_title: "Unselected fixture",
            golden_artifacts: ["snapshots/unselected.png"],
        },
    ], fixtureGoldens, fixtureCatalogEntries, [
        {
            row_id: "row.selected",
        },
    ]), [], "an owner slice must validate selected fixtures from capture intent and unselected fixtures from exact catalog/path existence");
    assert.deepEqual(resolveRegisteredFixtures([
        {
            fixture_id: "visual.fixture.selected",
            catalog_row_ids: ["row.selected"],
            playwright_scenario_title: "Selected fixture",
            golden_artifacts: ["snapshots/unselected.png"],
        },
    ], fixtureGoldens, fixtureCatalogEntries, [
        {
            row_id: "row.selected",
        },
    ]).map((fixture) => fixture.fixture_id), ["visual.fixture.selected"], "a selected fixture must resolve its exact runtime catalog row");
});
test("capture attribution binds report catalog source project and safe paths", async (t) => {
    const rendererProfile = loadVisualRendererProfile(root);
    const validCapturePayload = {
        schema_id: "cartulary.frontend_visual_capture_intent.v3", capture_id: "visual.capture.e78d7d3c42c82475cbdc",
        row_id: "visual.row", scenario_id: "scenario_0123456789ab",
        capture_intent: "test", expected_golden_path: "apps/web/e2e/visual-goldens/test.png",
        project_id: "chromium", renderer_profile_id: rendererProfile.profile_id, screenshot_assertion_location: "apps/web/e2e/support/visual/capture.ts#assertVisualRegression",
        test_file: "apps/web/e2e/auth.visual.spec.ts", test_title: "test",
        capture_profile: { browser_zoom_percent: 100, color_scheme: "dark", device_scale_factor: 1, reduced_motion: true, project_id: "chromium", snapshot_path_template: "{testDir}/visual-goldens/{arg}{-snapshotSuffix}{ext}", snapshot_suffix: "linux", viewport_css_px: "1280x720", surface_kind: "workbook_shell", density_id: "compact", expected_density_id: "compact", theme_id: "dark_graphite", expected_theme_id: "dark_graphite" },
    };
    validateCapturePayload(validCapturePayload);
    for (const density of [undefined, null, "", "unknown", "comfortable"]) {
        assert.throws(() => validateCapturePayload({ ...validCapturePayload, capture_profile: { ...validCapturePayload.capture_profile, density_id: density } }));
    }
    validateCapturePayload({ ...validCapturePayload, capture_profile: { ...validCapturePayload.capture_profile, surface_kind: "application_shell", density_id: null, expected_density_id: null } });
    assert.throws(() => validateCapturePayload({ ...validCapturePayload, schema_id: "cartulary.frontend_visual_capture_intent.v1" }));
    assert.throws(() => validateCapturePayload({ ...validCapturePayload, capture_id: "visual.capture." + "b".repeat(20) }), /capture identity/, "a forged capture digest must not be accepted");
    const attribution = {
        captureBindings: { test: { capture_id: validCapturePayload.capture_id, golden_path: validCapturePayload.expected_golden_path } },
        root, catalogEntries: [{ row_id: "visual.row", scenario_id: "scenario_0123456789ab", project_id: "chromium", file: "apps/web/e2e/auth.visual.spec.ts", title: "test" }],
        selectedRowIDs: ["visual.row"], report: { config: { rootDir: path.join(root, "apps/web/e2e") } }, spec: { file: "auth.visual.spec.ts", title: "test" }, playwrightTest: { projectName: "chromium" },
    };
    assert.equal(validateCaptureAttribution(validCapturePayload, attribution), attribution.catalogEntries[0]);
    for (const mutate of [
        (value) => { value.selectedRowIDs = []; },
        (value) => { value.captureBindings.test.golden_path = "apps/web/e2e/visual-goldens/wrong.png"; },
        (value) => { value.catalogEntries.push(value.catalogEntries[0]); },
        (value) => { value.catalogEntries[0].scenario_id = "scenario_aaaaaaaaaaaa"; },
        (value) => { value.spec.file = "evidence.visual.spec.ts"; },
        (value) => { value.spec.title = "wrong title"; },
        (value) => { value.playwrightTest.projectName = "wrong project"; },
        (value) => { value.report.config.rootDir = path.dirname(root); },
    ]) {
        const invalid = structuredClone(attribution);
        mutate(invalid);
        assert.throws(() => validateCaptureAttribution(validCapturePayload, invalid));
    }
    for (const mutate of [
        (value) => { value.test_file = "apps/web/e2e/../auth.visual.spec.ts"; },
        (value) => { value.test_file = "/apps/web/e2e/auth.visual.spec.ts"; },
        (value) => { value.expected_golden_path = "apps/web/e2e/visual-goldens/../test.png"; },
        (value) => { value.expected_golden_path = "apps/web/e2e/visual-goldens/alias\\test.png"; },
        (value) => { value.screenshot_assertion_location = "apps/web/e2e/auth.visual.spec.ts#assertVisualRegression"; },
        (value) => { value.capture_profile.project_id = "wrong"; },
        (value) => { value.schema_id = "cartulary.frontend_visual_capture_intent.v2"; },
    ]) {
        const invalid = structuredClone(validCapturePayload);
        mutate(invalid);
        assert.throws(() => validateCapturePayload(invalid));
    }
    const captureSourceRoot = mkdtempSync(path.join(os.tmpdir(), "visual-source-"));
    try {
        mkdirSync(path.join(captureSourceRoot, "apps/web/e2e"), { recursive: true });
        symlinkSync(path.join(root, validCapturePayload.test_file), path.join(captureSourceRoot, validCapturePayload.test_file));
        assert.throws(() => validateCaptureAttribution(validCapturePayload, { ...attribution, root: captureSourceRoot, report: { config: { rootDir: path.join(captureSourceRoot, "apps/web/e2e") } } }), /ELOOP|unsafe|symlink/u);
    }
    finally {
        rmSync(captureSourceRoot, { recursive: true, force: true });
    }
    // File organization is independent of capture identity; the catalog/report join still owns admission.
    validateCapturePayload({ ...validCapturePayload, test_file: "apps/web/e2e/features/account.presentation.spec.ts" });
    assert.throws(() => validateCapturePayload({ ...validCapturePayload, test_file: "apps/web/e2e/./auth.visual.spec.ts" }));
});
