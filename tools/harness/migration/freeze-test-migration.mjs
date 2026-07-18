import { createHash } from "node:crypto";
import { execFileSync } from "node:child_process";
import { existsSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { createRequire } from "node:module";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../..");
const baselinePath = path.join(root, "tools/test_migration_baseline.json");
const crosswalkPath = path.join(root, "tools/test_migration_crosswalk.json");
const requireFromRoot = createRequire(path.join(root, "package.json"));

function fail(message) {
  throw new Error(message);
}

function canonical(value) {
  if (Array.isArray(value)) return `[${value.map(canonical).join(",")}]`;
  if (value && typeof value === "object") {
    return `{${Object.keys(value).sort().map((key) => `${JSON.stringify(key)}:${canonical(value[key])}`).join(",")}}`;
  }
  return JSON.stringify(value);
}

function digest(value) {
  const bytes = Buffer.isBuffer(value) ? value : Buffer.from(typeof value === "string" ? value : canonical(value));
  return `sha256:${createHash("sha256").update(bytes).digest("hex")}`;
}

function json(relative) {
  return JSON.parse(readFileSync(path.join(root, relative), "utf8"));
}

function git(...args) {
  return execFileSync("git", args, { cwd: root, encoding: "utf8" }).trim();
}

function trackedPaths() {
  return git("ls-files", "-z").split("\0").filter(Boolean).sort();
}

function writeJson(file, value) {
  mkdirSync(path.dirname(file), { recursive: true });
  writeFileSync(file, `${JSON.stringify(value, null, 2)}\n`);
}

function jsonValue(value) {
  return JSON.parse(JSON.stringify(value));
}

function validateMigrationArtifact(schemaName, value) {
  const Ajv2020 = requireFromRoot("ajv/dist/2020").default;
  const schema = json(`tools/harness/migration/schemas/${schemaName}.schema.json`);
  const validator = new Ajv2020({ allErrors: true, strict: false }).compile(schema);
  if (!validator(value)) {
    fail(`${schemaName} validation failed: ${JSON.stringify(validator.errors)}`);
  }
}

function goSelector(row) {
  const symbols = [...new Set([...(row.symbols ?? []), ...(row.symbol ? [row.symbol] : [])])].sort();
  return { package: row.package, file: row.file, symbols };
}

function backendSelector(row) {
  if (row.runner === "playwright") {
    return {
      file: row.file ?? null,
      scenario_symbols: [...new Set(row.scenario_symbols ?? [])].sort(),
      titles: [...new Set([...(row.titles ?? []), ...(row.title ? [row.title] : [])])].sort()
    };
  }
  return goSelector(row);
}

function frontendRunner(row) {
  const targetNames = (row.targets ?? []).map((target) => target.target_name);
  if (targetNames.some((target) => target.startsWith("browser-e2e"))) return "playwright";
  if (row.scenario_titles?.length) return "vitest";
  return "shell";
}

function frontendSelector(row) {
  const targets = [...new Set((row.targets ?? []).map((target) => target.target_name))].sort();
  const scenarioTitles = [...new Set(row.scenario_titles ?? [])].sort();
  return scenarioTitles.length ? { targets, scenario_titles: scenarioTitles } : { command_ids: targets };
}

const tracked = trackedPaths();
const backendFiles = Array.from({ length: 13 }, (_, index) => `tools/phase${index}_test_map.json`);
const frontendFiles = tracked.filter((file) => /^tools\/frontend_phase_maps\/fe_p\d+_test_map\.json$/.test(file));
const graphFile = "tools/subsystem_test_maps/graph_projection_test_map.json";
const identities = [];
const backendSupport = [];

for (const sourcePath of backendFiles) {
  const manifest = json(sourcePath);
  for (const group of ["unit", "integration", "e2e", "visual"]) {
    for (const row of manifest[group] ?? []) {
      const selector = backendSelector(row);
      identities.push({
        source_registry_id: "backend_phase_maps",
        legacy_row_id: row.id,
        source_path: sourcePath,
        source_group: group,
        runner: row.runner === "playwright" ? "playwright" : "go",
        selector,
        selector_digest: digest(selector)
      });
    }
  }
  for (const [index, row] of (manifest.support_go_targets ?? []).entries()) {
    const selector = goSelector(row);
    backendSupport.push({
      candidate_id: `${manifest.phase}:support:${String(index + 1).padStart(3, "0")}`,
      source_path: sourcePath,
      classification: row.evidence_class ?? "implementation_support",
      selector,
      selector_digest: digest(selector)
    });
  }
}

for (const sourcePath of frontendFiles) {
  const manifest = json(sourcePath);
  for (const row of manifest.rows) {
    const selector = frontendSelector(row);
    identities.push({
      source_registry_id: "frontend_phase_maps",
      legacy_row_id: row.id,
      source_path: sourcePath,
      source_group: row.layer,
      runner: frontendRunner(row),
      selector,
      selector_digest: digest(selector)
    });
  }
}

const graph = json(graphFile);
for (const group of ["unit", "integration"]) {
  for (const row of graph[group] ?? []) {
    const selector = goSelector(row);
    identities.push({
      source_registry_id: "graph_projection_subsystem",
      legacy_row_id: row.id,
      source_path: graphFile,
      source_group: group,
      runner: "go",
      selector,
      selector_digest: digest(selector)
    });
  }
}

identities.sort((a, b) => `${a.source_registry_id}\0${a.legacy_row_id}`.localeCompare(`${b.source_registry_id}\0${b.legacy_row_id}`));
backendSupport.sort((a, b) => a.candidate_id.localeCompare(b.candidate_id));

const classifications = json("tools/test_accounting_classification.json").vitest.map((row, index) => {
  const selector = {
    target: row.target,
    ...(row.file ? { file: row.file } : {}),
    ...(row.file_pattern ? { file_pattern: row.file_pattern } : {})
  };
  return {
    candidate_id: `vitest:${String(index + 1).padStart(3, "0")}`,
    source_path: "tools/test_accounting_classification.json",
    classification: row.coverage,
    selector,
    selector_digest: digest(selector)
  };
});

if (identities.length !== 548) fail(`expected 548 authoritative identities, found ${identities.length}`);
if (backendSupport.length !== 37) fail(`expected 37 backend support candidates, found ${backendSupport.length}`);
if (classifications.length !== 228) fail(`expected 228 Vitest classifications, found ${classifications.length}`);

const ledgers = tracked.filter((file) => /^docs\/testing\/(?:frontend_phase_coverage_ledgers\/fe_p\d+|phase\d+)_coverage_ledger\.md$/.test(file));
if (ledgers.length !== 26) fail(`expected 26 phase ledgers, found ${ledgers.length}`);

const phasePattern = /(?:^|[/_.-])(?:phase\d+|fe[_-]?p\d+|fe[_-]?v[_-]?p\d+)(?:[/_.-]|$)/i;
const phaseIdentityPaths = tracked.filter((file) => phasePattern.test(file));
const visualGoldens = tracked
  .filter((file) => /^apps\/web\/e2e\//.test(file) && /(?:\.png|\.webp|\.jpg|\.jpeg)$/i.test(file))
  .map((file) => ({ path: file, digest: digest(readFileSync(path.join(root, file))) }));
const executableRoots = /^(?:cmd|internal|apps|packages|tools|contracts)\//;
const documentationReadCandidates = tracked.filter((file) => {
  if (!executableRoots.test(file) || !/\.(?:go|js|mjs|cjs|ts|tsx|sh|json)$/.test(file)) return false;
  const text = readFileSync(path.join(root, file), "utf8");
  return /docs\/(?:spec\/)?[^\s"'`]+(?:\.md|\/)/.test(text) && /readFile|ReadFile|realpath|stat|readdir|owner_document|source_spec|guide_path/.test(text);
});
function productionOwner(file) {
  if (/incidents_phase2|incidents\/phase2/.test(file)) return "module.incidents";
  if (/recovery_phase10/.test(file)) return "module.recovery";
  if (/reporting_phase11/.test(file)) return "module.reporting";
  if (/savedviews_phase8/.test(file)) return "module.savedviews";
  if (/timeline_phase3/.test(file)) return "module.timeline";
  if (/workbook_startup_phase8/.test(file)) return "module.workbook";
  fail(`production follow-up owner is not classified: ${file}`);
}

const productionFollowups = tracked
  .filter((file) => /^(?:db\/queries|internal\/gen\/sql|internal\/modules)\//.test(file) && !/(?:_test\.go|\/testdata\/)/.test(file) && phasePattern.test(file))
  .map((file) => ({
    path: file,
    classification: file.startsWith("internal/gen/") ? "generated_descendant" : "authored_historical_delivery_metadata",
    owner_id: productionOwner(file),
    reason: file.startsWith("internal/gen/")
      ? "Generated descendant; update only through the matching authored SQL owner in a separately authorized module change."
      : "Historical delivery metadata has no runtime phase meaning; decouple live test selectors and retain a module-owner rename follow-up."
  }));
const generatedPaths = tracked.filter((file) => /^(?:internal\/gen|packages\/(?:protocol-ts|ui-contracts)\/src\/generated|tools\/task_surface\.generated\.mk)/.test(file) && phasePattern.test(file));

const sourceFiles = [...backendFiles, ...frontendFiles, graphFile, "tools/test_accounting_classification.json"].sort();
const sourceProjection = sourceFiles.map((file) => ({ file, digest: digest(readFileSync(path.join(root, file))) }));
const baseline = jsonValue({
  schema_id: "cartulary.test_migration_baseline.v1",
  baseline_commit: git("rev-parse", "HEAD"),
  baseline_tree: git("rev-parse", "HEAD^{tree}"),
  source_digest: digest(sourceProjection),
  authoritative_population: 548,
  source_counts: { backend: 456, frontend: 87, graph_projection: 5 },
  identities,
  auxiliary_candidates: {
    backend_support: backendSupport,
    vitest_classifications: classifications
  },
  inventories: {
    ledgers,
    phase_identity_paths: phaseIdentityPaths,
    visual_goldens: visualGoldens,
    documentation_read_candidates: documentationReadCandidates,
    production_followups: productionFollowups,
    generated_paths: generatedPaths
  }
});
const baselineDigest = digest(baseline);
const pendingKeys = identities.map(({ source_registry_id, legacy_row_id }) => ({ source_registry_id, legacy_row_id }));
const crosswalk = jsonValue({
  schema_id: "cartulary.test_migration_crosswalk.v1",
  baseline_digest: baselineDigest,
  status: "in_progress",
  pending_baseline_keys: pendingKeys,
  dispositions: [],
  auxiliary_dispositions: [],
  new_rows: []
});

if (process.argv.includes("--check")) {
  if (!existsSync(baselinePath) || !existsSync(crosswalkPath)) fail("migration baseline or crosswalk is missing");
  const storedBaseline = json("tools/test_migration_baseline.json");
  const currentCrosswalk = json("tools/test_migration_crosswalk.json");
  validateMigrationArtifact("cartulary.test_migration_baseline.v1", storedBaseline);
  validateMigrationArtifact("cartulary.test_migration_crosswalk.v1", currentCrosswalk);
  const storedBaselineDigest = digest(storedBaseline);
  const migrationStarted =
    currentCrosswalk.dispositions.length > 0 ||
    currentCrosswalk.auxiliary_dispositions.length > 0 ||
    currentCrosswalk.new_rows.length > 0;
  if (!migrationStarted && canonical(storedBaseline) !== canonical(baseline)) {
    fail("test migration baseline drift");
  }
  if (currentCrosswalk.baseline_digest !== storedBaselineDigest) {
    fail("test migration crosswalk baseline digest drift");
  }
  const represented = currentCrosswalk.pending_baseline_keys.length + currentCrosswalk.dispositions.length;
  if (represented !== 548) fail(`crosswalk represents ${represented}/548 baseline identities`);
  const frozenKeys = new Set(
    storedBaseline.identities.map((entry) => `${entry.source_registry_id}\0${entry.legacy_row_id}`),
  );
  const representedKeys = [
    ...currentCrosswalk.pending_baseline_keys,
    ...currentCrosswalk.dispositions,
  ].map((entry) => `${entry.source_registry_id}\0${entry.legacy_row_id}`);
  if (new Set(representedKeys).size !== representedKeys.length) {
    fail("test migration crosswalk represents a baseline identity more than once");
  }
  if (representedKeys.some((key) => !frozenKeys.has(key))) {
    fail("test migration crosswalk contains an unknown baseline identity");
  }
  const frozenAuxiliaryIDs = new Set(
    Object.values(storedBaseline.auxiliary_candidates)
      .flat()
      .map((entry) => entry.candidate_id),
  );
  const auxiliaryIDs = currentCrosswalk.auxiliary_dispositions.map((entry) => entry.candidate_id);
  if (new Set(auxiliaryIDs).size !== auxiliaryIDs.length) {
    fail("test migration crosswalk contains a duplicate auxiliary disposition");
  }
  if (auxiliaryIDs.some((candidateID) => !frozenAuxiliaryIDs.has(candidateID))) {
    fail("test migration crosswalk contains an unknown auxiliary candidate");
  }
  if (migrationStarted && existsSync(path.join(root, "tools/test_catalog_owner.json"))) {
    const { loadTestCatalog } = await import("../test-catalog/test-catalog.mjs");
    const catalog = loadTestCatalog(root);
    const authorizedRows = new Map();
    for (const entry of [
      ...currentCrosswalk.dispositions.filter((item) => item.disposition !== "deleted"),
      ...currentCrosswalk.new_rows,
    ]) {
      if (authorizedRows.has(entry.row_id)) {
        fail(`test migration crosswalk authorizes duplicate row ${entry.row_id}`);
      }
      authorizedRows.set(entry.row_id, entry);
    }
    for (const row of catalog.rows) {
      const authorization = authorizedRows.get(row.row_id);
      if (!authorization) fail(`catalog row ${row.row_id} has no migration authorization`);
      if (authorization.owner_id !== row.owner_id) {
        fail(`catalog row ${row.row_id} owner differs from migration authorization`);
      }
      if (canonical(authorization.verification_ids) !== canonical(row.verification_ids)) {
        fail(`catalog row ${row.row_id} verifications differ from migration authorization`);
      }
      authorizedRows.delete(row.row_id);
    }
    if (authorizedRows.size > 0) {
      fail(`migration authorization has no catalog row: ${[...authorizedRows.keys()][0]}`);
    }
  }
  process.stdout.write(`${JSON.stringify({ schema_id: "cartulary.test_migration_baseline_check.v1", status: "passed", authoritative_population: 548, pending: currentCrosswalk.pending_baseline_keys.length, dispositions: currentCrosswalk.dispositions.length, backend_support: storedBaseline.auxiliary_candidates.backend_support.length, vitest_classifications: storedBaseline.auxiliary_candidates.vitest_classifications.length, auxiliary_dispositions: auxiliaryIDs.length })}\n`);
} else {
  writeJson(baselinePath, baseline);
  writeJson(crosswalkPath, crosswalk);
  process.stdout.write(`wrote ${path.relative(root, baselinePath)} and ${path.relative(root, crosswalkPath)}\n`);
}
