import { mkdirSync, readFileSync, writeFileSync } from "node:fs";
import path from "node:path";
import { digest } from "../ui-review/session-files.mjs";
import { repoRoot } from "../ui-review/toolchain.mjs";
import { captureCapabilitySnapshot } from "../../scheduler/work-graph/capability.mjs";

// Synthetic producer evidence uses current machine owners and fabricated process
// identities. No retained run, private browser observation, or Markdown is input.
export function importFixture(root) {
  const read = (file) => JSON.parse(readFileSync(path.join(repoRoot, file)));
  const fileDigest = (file) => digest(readFileSync(path.join(repoRoot, file)));
  const sha = `sha256:${"a".repeat(64)}`, timestamp = "2026-09-27T00:00:00.000Z";
  const row = read("tools/test_families/module.auth.json").rows.find((entry) => entry.runner === "playwright" && entry.selector.stage === "visual" && entry.selector.titles.some((title) => title.startsWith("Capture auth gateway initial,")));
  const goldenPath = "apps/web/e2e/workbook.visual.spec.ts-snapshots/auth-initial-linux.png";
  const fixture = read("tools/frontend_visual_fixture_registry.json").fixtures.find((entry) => entry.golden_artifacts.includes(goldenPath));
  const renderer = read("tools/frontend_visual_renderer_profile.json");
  const registered = fixture?.capture_profiles[goldenPath] ?? { browser_zoom_percent: 100, density_id: null, device_scale_factor: 1, theme_id: "dark_graphite", viewport_css_px: "1440x900", surface_kind: "application_shell" };
  const capture = {
    capture_id: `visual.capture.${digest(Buffer.from(JSON.stringify([row.selector.project_id, row.selector.titles[0], goldenPath]))).slice(0, 20)}`,
    owner_id: row.owner_id, row_id: row.row_id, scenario_id: row.selector.scenario_ids[0], project_id: row.selector.project_id, renderer_profile_id: renderer.profile_id,
    screenshot_assertion_location: "synthetic:1", assertion_file: row.selector.file, capture_intent: "auth-initial", screenshot_name_source: "auth-initial", expected_golden_path: goldenPath, test_title: row.selector.titles[0],
    capture_profile: { ...registered, color_scheme: "light", reduced_motion: true, project_id: row.selector.project_id, snapshot_path_template: "{snapshotDir}/{testFileDir}/{testFileName}-snapshots/{arg}{-snapshotSuffix}{ext}", snapshot_suffix: "linux", expected_density_id: registered.density_id, expected_theme_id: registered.theme_id },
  };
  const sourcePaths = { catalog_owner: "tools/test_catalog_owner.json", family_manifest: "tools/test_families/module.auth.json", fixture_registry: "tools/frontend_visual_fixture_registry.json", renderer_profile: "tools/frontend_visual_renderer_profile.json", golden_manifest: "tools/frontend_visual_golden_manifest.json", playwright_config: "apps/web/playwright.config.ts", screenshot_helper: "apps/web/e2e/workbook.visual.spec.ts" };
  const { schema_id: _schema, ...profile } = renderer;
  const reconciliation = {
    schema_id: "cartulary.frontend_visual_reconciliation.v3", status: "fail", renderer: { ...profile, attestation_count: 1 },
    golden_manifest: { path: sourcePaths.golden_manifest, sha256: fileDigest(sourcePaths.golden_manifest), renderer_profile_id: renderer.profile_id, status: "pass" },
    source_refs: Object.entries(sourcePaths).map(([kind, file]) => ({ kind, path: file, sha256: fileDigest(file), project_ids: [], snapshot_path_template: null, symbols: [] })), capture_intents: [capture],
    goldens: [{ golden_path: goldenPath, sha256: fileDigest(goldenPath), consumer_capture_ids: [capture.capture_id], catalog_row_ids: [row.row_id], owner_ids: [row.owner_id], scenario_ids: [capture.scenario_id], project_ids: [row.selector.project_id], fixture_ids: fixture ? [fixture.fixture_id] : [], design_contract_ids: [], consumer_refs: [], classification: "active", permitted_action: "retain" }],
    counts: { capture_intents: 1, committed_goldens: 1, active: 1, orphan: 0, missing_golden: 0, ambiguous_mapping: 0, registered_fixtures: fixture ? 1 : 0, unresolved_registered_fixtures: 0 }, artifact_refs: [], errors: [],
  };
  const manifest = { schema_id: "cartulary.harness_run_manifest.v1", run_id: path.basename(root), command_id: "cartulary.harness.command.browser_e2e_visual.v1", target: "browser-e2e-visual", declared_inputs: {}, source_commit: "a".repeat(40), source_state: "dirty", source_digest: sha, toolchain_digest: sha, system_digest: sha, graph_digest: sha, capability_snapshot: captureCapabilitySnapshot({ root: repoRoot }), cache_mode: "normal", started_at: timestamp };
  const receipt = { schema_id: "cartulary.frontend_build_artifact.v1", run_id: manifest.run_id, artifact_id: "production", producer_unit_id: "target:build-web", source_digest: sha, toolchain_digest: sha, content_digest: sha };
  const process = { status: "pass", pid: 1, process_group_id: 1, boot_id: "synthetic", start_time_ticks: 1, effective_uid: 0, executable_device: 1, executable_inode: 1, executable_sha256: sha };
  const stack = {
    schema_id: "cartulary.web_e2e_stack.v7", suite_id: "synthetic", browser_session_id: "synthetic", service_mode: "owned", runtime_profile_id: row.runtime_profile_id, configuration_fingerprint: sha,
    service_admission_ref: "admission.json", service_admission_sha256: sha, postgres_identity: { database_name: "synthetic", template_database: "synthetic", schema_hash: sha, fixture_capability: "postgres_dedicated" }, object_store_identity: { endpoint_origin: "http://127.0.0.1:9000", bucket: "synthetic", fixture_generation: sha },
    backend: { ...process, origin: "http://127.0.0.1:3000", port: 3000, ready_at: timestamp }, frontend: { ...process, origin: "http://127.0.0.1:4000", port: 4000, frontend_mode: "preview", frontend_command_kind: "vite-preview", build_artifact_ref: "frontend.json", build_receipt_sha256: sha, build_artifact_sha256: sha, ready_at: timestamp },
    fixture_identity: { fixture_capability: "browser_stack", fixture_id: sha, scenario_id: "synthetic" }, startup_diagnostics_ref: "startup.json", startup_diagnostics_sha256: sha, lease_ref: "lease.json", lease_sha256: sha, ready_at: timestamp,
  };
  const payload = { schema_id: "cartulary.frontend_visual_capture_intent.v2", test_file: capture.assertion_file, test_title: capture.test_title, capture_profile: capture.capture_profile, capture_id: capture.capture_id, capture_intent: capture.capture_intent, expected_golden_path: goldenPath, project_id: capture.project_id, renderer_profile_id: capture.renderer_profile_id, screenshot_assertion_location: capture.screenshot_assertion_location };
  const result = { status: "failed", attachments: [{ name: `cartulary-visual-capture-intent-${capture.capture_id}.json`, contentType: "application/json", body: Buffer.from(JSON.stringify(payload)).toString("base64") }] };
  const report = { config: { version: "1.59.1", rootDir: path.join(repoRoot, "apps/web/e2e") }, suites: [{ specs: [{ title: capture.test_title, file: "workbook.visual.spec.ts", tests: [{ projectName: capture.project_id, results: [result] }] }] }] };
  const group = { schema_id: "cartulary.browser_group_result.v6", target_id: "browser-e2e-visual", stage_id: "visual", group_id: "synthetic", browser_session_id: "synthetic", runtime_profile_id: row.runtime_profile_id, service_requirement: "test-services", fixture_capabilities: ["browser_stack"], service_dependencies: ["object_store", "postgres"], resource_profile_ids: ["browser_isolated"], selected_rows: [row.row_id], started_at: timestamp, finished_at: timestamp, duration_ms: 0, status: "fail", exit_code: 10, row_results: [{ row_id: row.row_id, terminal_state: "failed", duration_ms: 0, exit_code: 10, failure_class: "product", failure_reason: "test_assertion_failure", failure_diagnostic: null }], session_artifacts: [], artifacts: { playwright_report: "report.json", stdout: "stdout.log", stderr: "stderr.log" } };
  const target = { schema_id: "cartulary.browser_target_result.v4", target_id: "browser-e2e-visual", status: "fail", group_results: [], sessions: [], artifacts: [], generated_at: timestamp };
  const write = (file, value) => { const bytes = Buffer.from(`${JSON.stringify(value)}\n`); mkdirSync(path.dirname(path.join(root, file)), { recursive: true, mode: 0o700 }); writeFileSync(path.join(root, file), bytes, { mode: 0o600 }); return `sha256:${digest(bytes)}`; };
  const publish = () => {
    write("run-manifest.json", manifest); write("report.json", report);
    stack.frontend.build_receipt_sha256 = write("frontend.json", receipt);
    group.session_artifacts = [{ kind: "stack_v7", ref: "stack.json", sha256: write("stack.json", stack) }, { kind: "startup_diagnostics_v2", ref: "startup.json", sha256: sha }];
    target.group_results = [{ group_id: group.group_id, browser_session_id: group.browser_session_id, ref: "group.json", sha256: write("group.json", group) }];
    target.sessions = [{ browser_session_id: group.browser_session_id, runtime_profile_id: group.runtime_profile_id, service_requirement: group.service_requirement, artifacts: group.session_artifacts }];
    target.artifacts = [{ kind: "frontend_visual_reconciliation_v3", ref: "reconciliation.json", sha256: write("reconciliation.json", reconciliation) }];
    write("browser-e2e-visual/browser-target-result.json", target);
  };
  publish(); return { request: { source: "canonical_visual", run_root: root, capture_id: capture.capture_id }, manifest, receipt, stack, group, target, reconciliation, report, result, publish, goldenPath };
}
