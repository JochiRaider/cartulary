#!/usr/bin/env node
import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { createHash } from "node:crypto";
import {
  existsSync,
  mkdirSync,
  mkdtempSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import path from "node:path";
import { test } from "node:test";
import { fileURLToPath } from "node:url";

import {
  loadExecutionTopology,
  renderBrowserBatchManifest,
  renderCheckScheduleManifest,
  renderTaskSurfaceManifest,
} from "../generated-artifacts/index.mjs";
import {
  expandServiceBackedSchedule,
  expandServiceBackedScheduleForCheck,
} from "../execution/service-backed/index.mjs";
import {
  collectTaskSurfaceManifestErrors,
  renderTaskSurfaceMake,
} from "../generated-artifacts/index.mjs";
import { collectFrontendGuideTargetRestatementErrors } from "../phase-accounting/index.mjs";
import {
  collectPlaywrightTitleObservationsForTarget,
  collectVitestTitleObservations,
  frontendScenarioStatus,
} from "../output/test-output/frontend-row-evidence.mjs";
import {
  artifactRef,
  buildToolRunSummary,
  machineOutput,
  terminalArtifactPath,
  toolRunSummarySchemaID,
} from "../output/index.mjs";
import {
  HarnessConfigError,
  generateTestRouteToken,
  normalizeFailureClass,
  primaryPublicFailure,
  preflightPublicTarget,
  redactString,
  redactValue,
  runCleanup,
  testRouteTokenValid,
  validateSchema,
} from "../contract/index.mjs";
import { collectGoShardsForTarget } from "../backend/backend-shard-plan.mjs";
import { renderServiceBackedScheduleManifest } from "../generated-artifacts/index.mjs";
import {
  ownerFacadePathLists,
  unsupportedPrivateHelperRules,
} from "../static-analysis/harness-helper-ownership-registry.mjs";
import { collectHarnessImportBoundaryViolations } from "../static-analysis/harness-import-boundary.mjs";
import { validateSchedulerEventOrderFile } from "../scheduler/scheduler/event-order.mjs";
import {
  schedulerCapacityProfilesByFamily,
  schedulerCapacityProfileValues,
  schedulerFamilyForCapacityProfile,
  schedulerFamilyValues,
} from "../scheduler/scheduler-family-contract.mjs";
import { schedulerSummaryLine } from "../scheduler/scheduler-reporting.mjs";
import { validateSchedulerResourceRegistrySemantics } from "../scheduler/scheduler-resources.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "../../..");

function readJSON(relativePath) {
  return JSON.parse(readFileSync(path.join(repoRoot, relativePath), "utf8"));
}

function writeJSONFile(file, value) {
  mkdirSync(path.dirname(file), { recursive: true });
  writeFileSync(file, `${JSON.stringify(value, null, 2)}\n`);
}

function writeFixtureFile(root, relativePath, content) {
  const file = path.join(root, relativePath);
  mkdirSync(path.dirname(file), { recursive: true });
  writeFileSync(file, content);
}

const fixtureImportKeyword = ["im", "port"].join("");
const fixtureExportKeyword = ["ex", "port"].join("");

function fixtureImport(specifier) {
  return `${fixtureImportKeyword} "${specifier}";\n`;
}

function fixtureDynamicImport(specifier) {
  return `${fixtureImportKeyword}("${specifier}")`;
}

function fixtureExportFrom(symbol, specifier) {
  return `${fixtureExportKeyword} { ${symbol} } from "${specifier}";\n`;
}

function legacyPlanningImport(file) {
  return ["..", "planning", file].join("/");
}

function legacyPlanningPath(file) {
  return ["tools", "harness", "planning", file].join("/");
}

test("network flow fixture manifest schema is closed and byte-addressed", async () => {
  function sha256(content) {
    return createHash("sha256").update(content).digest("hex");
  }

  function bundleHash(entries) {
    const hash = createHash("sha256");
    for (const entry of entries) {
      hash.update(entry.logical_path, "utf8");
      hash.update(Buffer.from([0]));
      hash.update(entry.sha256, "utf8");
      hash.update(Buffer.from([0]));
      hash.update(String(entry.size_bytes), "utf8");
      hash.update("\n", "utf8");
    }
    return hash.digest("hex");
  }

  const sourceContent = "src_ip,dst_ip\n10.0.0.1,10.0.0.2\n";
  const expectedContent = "{\"rows\":[]}\n";
  const transcriptContent = "{\"status\":\"pass\"}\n";
  const sourceFile = {
    logical_path: "source/cisco-sna-minimal.csv",
    media_type: "text/csv",
    size_bytes: Buffer.byteLength(sourceContent),
    sha256: sha256(sourceContent),
    role: "input",
    newline_policy: "lf_required",
  };
  const expectedFile = {
    logical_path: "expected/rows.json",
    media_type: "application/json",
    size_bytes: Buffer.byteLength(expectedContent),
    sha256: sha256(expectedContent),
    role: "expected_rows",
  };
  const transcriptFile = {
    logical_path: "transcripts/apply.jsonl",
    media_type: "application/jsonl",
    size_bytes: Buffer.byteLength(transcriptContent),
    sha256: sha256(transcriptContent),
    transcript_kind: "apply",
  };
  const manifest = {
    schema_id: "cartulary.network_flow_fixture_manifest.v1",
    manifest_version: 1,
    fixture_id: "NF-FIX-001-cisco-sna-minimal",
    profile_id: "network_flow_activity",
    freeze: {
      status: "frozen",
      revision: 1,
      change_policy: "new_fixture_revision_required",
    },
    owner_refs: [
      {
        document: "docs/network-flow-activity-nlspec.md",
        requirement_ids: ["NF-REQ-177"],
        acceptance_ids: ["NF-AC-052"],
      },
    ],
    source_files: [sourceFile],
    expected_artifacts: [expectedFile],
    transcript_files: [transcriptFile],
    acceptance_ids: ["NF-AC-052"],
    execution_selectors: ["phase12/network-flow/NF-AC-052"],
    source_bundle_sha256: bundleHash([sourceFile]),
    expected_bundle_sha256: bundleHash([expectedFile, transcriptFile]),
  };

  await validateSchema("cartulary.network_flow_fixture_manifest.v1", manifest);

  await assert.rejects(
    validateSchema("cartulary.network_flow_fixture_manifest.v1", {
      ...manifest,
      unexpected: true,
    }),
    /must NOT have additional properties/u,
  );

  const invalidDigest = structuredClone(manifest);
  invalidDigest.source_files[0].sha256 = "ABC";
  await assert.rejects(
    validateSchema("cartulary.network_flow_fixture_manifest.v1", invalidDigest),
    /must match pattern/u,
  );

  const root = mkdtempSync(path.join(repoRoot, "tmp", "network-flow-fixture."));
  try {
    const fixtureDir = path.join(root, manifest.fixture_id);
    writeFixtureFile(fixtureDir, sourceFile.logical_path, sourceContent);
    writeFixtureFile(fixtureDir, expectedFile.logical_path, expectedContent);
    writeFixtureFile(fixtureDir, transcriptFile.logical_path, transcriptContent);
    const manifestPath = path.join(fixtureDir, "manifest.json");
    writeFileSync(manifestPath, `${JSON.stringify(manifest, null, 2)}\n`);

    const checker = path.join(
      repoRoot,
      "tools/harness/generated-artifacts/check-json-shapes.mjs",
    );
    const pass = spawnSync(
      process.execPath,
      [checker, "--kind", "network-flow-fixture-manifest", "--file", manifestPath],
      { encoding: "utf8" },
    );
    assert.equal(pass.status, 0, pass.stderr);

    const mismatched = structuredClone(manifest);
    mismatched.source_files[0].sha256 = "0".repeat(64);
    mismatched.source_bundle_sha256 = bundleHash(mismatched.source_files);
    writeFileSync(manifestPath, `${JSON.stringify(mismatched, null, 2)}\n`);
    const fail = spawnSync(
      process.execPath,
      [checker, "--kind", "network-flow-fixture-manifest", "--file", manifestPath],
      { encoding: "utf8" },
    );
    assert.notEqual(fail.status, 0);
    assert.match(fail.stderr, /sha256 does not match file bytes/u);

    writeFileSync(manifestPath, `${JSON.stringify(manifest, null, 2)}\n`);
    writeFixtureFile(fixtureDir, "source/unlisted.csv", "extra\n");
    const unlisted = spawnSync(
      process.execPath,
      [checker, "--kind", "network-flow-fixture-manifest", "--file", manifestPath],
      { encoding: "utf8" },
    );
    assert.notEqual(unlisted.status, 0);
    assert.match(unlisted.stderr, /unlisted fixture file source\/unlisted\.csv/u);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("contract family registry schema is closed and restricts planned families", async () => {
  const registry = readJSON("contracts/index.json");

  await validateSchema("cartulary.contract_family_registry.v1", registry);

  await assert.rejects(
    validateSchema("cartulary.contract_family_registry.v1", {
      ...registry,
      unexpected: true,
    }),
    /must NOT have additional properties/u,
  );

  const activatedNetworkFlow = structuredClone(registry);
  activatedNetworkFlow.families = activatedNetworkFlow.families.map((family) =>
    family.family_id === "network-flow"
      ? { ...family, generation_status: "active", activation_dependency_ids: [] }
      : family,
  );
  await validateSchema(
    "cartulary.contract_family_registry.v1",
    activatedNetworkFlow,
  );

  const checker = path.join(
    repoRoot,
    "tools/harness/generated-artifacts/check-json-shapes.mjs",
  );
  const artifactPath = path.join(repoRoot, "contracts/index.json");
  const pass = spawnSync(
    process.execPath,
    [checker, "--kind", "contract-family-registry", "--file", artifactPath],
    { encoding: "utf8" },
  );
  assert.equal(pass.status, 0, pass.stderr);

  const root = mkdtempSync(path.join(repoRoot, "tmp", "contract-family."));
  try {
    const mutableArtifactPath = path.join(root, "index.json");
    writeFileSync(
      mutableArtifactPath,
      `${JSON.stringify(activatedNetworkFlow, null, 2)}\n`,
    );
    const activePass = spawnSync(
      process.execPath,
      [checker, "--kind", "contract-family-registry", "--file", mutableArtifactPath],
      { encoding: "utf8" },
    );
    assert.equal(activePass.status, 0, activePass.stderr);

    const unexpectedPlanned = structuredClone(registry);
    unexpectedPlanned.families = unexpectedPlanned.families.map((family) =>
      family.family_id === "errors"
        ? {
            ...family,
            generation_status: "planned",
            activation_dependency_ids: ["NFA-GEN-002"],
          }
        : family,
    );
    writeFileSync(
      mutableArtifactPath,
      `${JSON.stringify(unexpectedPlanned, null, 2)}\n`,
    );
    const fail = spawnSync(
      process.execPath,
      [checker, "--kind", "contract-family-registry", "--file", mutableArtifactPath],
      { encoding: "utf8" },
    );
    assert.notEqual(fail.status, 0);
    assert.match(fail.stderr, /active output_order must be openapi/u);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("network flow activity accounting is closed and fails drift gaps", async () => {
  const accounting = readJSON("tools/network_flow_activity_accounting.json");

  await validateSchema("cartulary.network_flow_activity_accounting.v1", accounting);

  await assert.rejects(
    validateSchema("cartulary.network_flow_activity_accounting.v1", {
      ...accounting,
      unexpected: true,
    }),
    /must NOT have additional properties/u,
  );

  const checker = path.join(
    repoRoot,
    "tools/harness/generated-artifacts/check-json-shapes.mjs",
  );
  const artifactPath = path.join(
    repoRoot,
    "tools/network_flow_activity_accounting.json",
  );
  const pass = spawnSync(
    process.execPath,
    [checker, "--kind", "network-flow-activity-accounting", "--file", artifactPath],
    { encoding: "utf8" },
  );
  assert.equal(pass.status, 0, pass.stderr);

  const root = mkdtempSync(path.join(repoRoot, "tmp", "network-flow-accounting."));
  try {
    const sourceWithTodo = readFileSync(
      path.join(repoRoot, accounting.source_spec),
      "utf8",
    ).replace(
      "Adopted current-profile Core 00 revision at owner artifact `155b5f64`",
      "TODO: adopted Core 00 version",
    );
    const sourceWithTodoFile = path.join(root, "network-flow-with-todo.md");
    writeFileSync(sourceWithTodoFile, sourceWithTodo);
    const unresolvedLocator = structuredClone(accounting);
    unresolvedLocator.source_spec = path
      .relative(repoRoot, sourceWithTodoFile)
      .split(path.sep)
      .join(path.posix.sep);
    const unresolvedLocatorFile = path.join(root, "unresolved-locator.json");
    writeFileSync(
      unresolvedLocatorFile,
      `${JSON.stringify(unresolvedLocator, null, 2)}\n`,
    );
    const unresolvedLocatorResult = spawnSync(
      process.execPath,
      [
        checker,
        "--kind",
        "network-flow-activity-accounting",
        "--file",
        unresolvedLocatorFile,
      ],
      { encoding: "utf8" },
    );
    assert.notEqual(unresolvedLocatorResult.status, 0);
    assert.match(
      unresolvedLocatorResult.stderr,
      /locator for Core 00 must not contain TODO:/u,
    );

    const missingCopyPath = structuredClone(accounting);
    missingCopyPath.drift_accounting.required_copy_paths = [
      ...missingCopyPath.drift_accounting.required_copy_paths,
      "tools/missing-network-flow-accounting-input.json",
    ];
    const missingCopyPathFile = path.join(root, "missing-copy-path.json");
    writeFileSync(
      missingCopyPathFile,
      `${JSON.stringify(missingCopyPath, null, 2)}\n`,
    );
    const missingCopyPathResult = spawnSync(
      process.execPath,
      [
        checker,
        "--kind",
        "network-flow-activity-accounting",
        "--file",
        missingCopyPathFile,
      ],
      { encoding: "utf8" },
    );
    assert.notEqual(missingCopyPathResult.status, 0);
    assert.match(
      missingCopyPathResult.stderr,
      /copy_paths must include tools\/missing-network-flow-accounting-input\.json/u,
    );

    const plannedRegistry = readJSON("contracts/index.json");
    plannedRegistry.families = plannedRegistry.families.map((family) =>
      family.family_id === "network-flow"
        ? {
            ...family,
            generation_status: "planned",
            activation_dependency_ids:
              accounting.contract_registry.planned_activation_dependency_ids,
          }
        : family,
    );
    const plannedRegistryFile = path.join(root, "contracts-index.json");
    writeFileSync(
      plannedRegistryFile,
      `${JSON.stringify(plannedRegistry, null, 2)}\n`,
    );
    const stalePlannedRegistry = structuredClone(accounting);
    stalePlannedRegistry.contract_registry.path = path
      .relative(repoRoot, plannedRegistryFile)
      .split(path.sep)
      .join(path.posix.sep);
    const stalePlannedRegistryFile = path.join(root, "stale-planned-registry.json");
    writeFileSync(
      stalePlannedRegistryFile,
      `${JSON.stringify(stalePlannedRegistry, null, 2)}\n`,
    );
    const stalePlannedRegistryResult = spawnSync(
      process.execPath,
      [
        checker,
        "--kind",
        "network-flow-activity-accounting",
        "--file",
        stalePlannedRegistryFile,
      ],
      { encoding: "utf8" },
    );
    assert.notEqual(stalePlannedRegistryResult.status, 0);
    assert.match(
      stalePlannedRegistryResult.stderr,
      /planned but generated outputs contain Network Flow markers/u,
    );

    const activeWithDependencies = readJSON("contracts/index.json");
    activeWithDependencies.families = activeWithDependencies.families.map((family) =>
      family.family_id === "network-flow"
        ? { ...family, activation_dependency_ids: ["NFA-GEN-004"] }
        : family,
    );
    const activeWithDependenciesFile = path.join(
      root,
      "active-with-dependencies.json",
    );
    writeFileSync(
      activeWithDependenciesFile,
      `${JSON.stringify(activeWithDependencies, null, 2)}\n`,
    );
    const activeWithDependenciesAccounting = structuredClone(accounting);
    activeWithDependenciesAccounting.contract_registry.path = path
      .relative(repoRoot, activeWithDependenciesFile)
      .split(path.sep)
      .join(path.posix.sep);
    const activeWithDependenciesAccountingFile = path.join(
      root,
      "active-with-dependencies-accounting.json",
    );
    writeFileSync(
      activeWithDependenciesAccountingFile,
      `${JSON.stringify(activeWithDependenciesAccounting, null, 2)}\n`,
    );
    const activeWithDependenciesResult = spawnSync(
      process.execPath,
      [
        checker,
        "--kind",
        "network-flow-activity-accounting",
        "--file",
        activeWithDependenciesAccountingFile,
      ],
      { encoding: "utf8" },
    );
    assert.notEqual(activeWithDependenciesResult.status, 0);
    assert.match(
      activeWithDependenciesResult.stderr,
      /active but activation dependencies remain/u,
    );
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("network flow authored contracts are closed and index-owned", async () => {
  const contractIndex = readJSON("contracts/network-flow/index.json");

  await validateSchema("cartulary.network_flow_contract_index.v1", contractIndex);

  await assert.rejects(
    validateSchema("cartulary.network_flow_contract_index.v1", {
      ...contractIndex,
      unexpected: true,
    }),
    /must NOT have additional properties/u,
  );

  const checker = path.join(
    repoRoot,
    "tools/harness/generated-artifacts/check-json-shapes.mjs",
  );
  const artifactPath = path.join(repoRoot, "contracts/network-flow/index.json");
  const pass = spawnSync(
    process.execPath,
    [checker, "--kind", "network-flow-contract-index", "--file", artifactPath],
    { encoding: "utf8" },
  );
  assert.equal(pass.status, 0, pass.stderr);

  const root = mkdtempSync(
    path.join(repoRoot, "contracts/network-flow/tmp-contract."),
  );
  const relativePath = (file) =>
    path.relative(repoRoot, file).split(path.sep).join(path.posix.sep);
  const routes = readJSON("contracts/network-flow/routes.v1.json");
  const schemas = readJSON("contracts/network-flow/schemas.v1.json");
  const errors = readJSON("contracts/network-flow/errors.v1.json");
  const tempIndexPath = path.join(root, "index.json");
  const tempRoutesPath = path.join(root, "routes.v1.json");
  const tempSchemasPath = path.join(root, "schemas.v1.json");
  const tempErrorsPath = path.join(root, "errors.v1.json");
  const tempIndex = {
    ...contractIndex,
    contract_files: {
      routes: relativePath(tempRoutesPath),
      schemas: relativePath(tempSchemasPath),
      errors: relativePath(tempErrorsPath),
      timezone_provenance: contractIndex.contract_files.timezone_provenance,
    },
  };
  const writeTempContracts = ({
    routeContracts = routes,
    schemaBundle = schemas,
    errorContracts = errors,
  } = {}) => {
    writeJSONFile(tempRoutesPath, routeContracts);
    writeJSONFile(tempSchemasPath, schemaBundle);
    writeJSONFile(tempErrorsPath, errorContracts);
    writeJSONFile(tempIndexPath, tempIndex);
  };
  const runChecker = () =>
    spawnSync(
      process.execPath,
      [checker, "--kind", "network-flow-contract-index", "--file", tempIndexPath],
      { encoding: "utf8" },
    );

  try {
    writeTempContracts();
    const tempPass = runChecker();
    assert.equal(tempPass.status, 0, tempPass.stderr);

    const routePathDrift = structuredClone(routes);
    routePathDrift.routes[0].path =
      "/api/v1/incidents/{incident_id}/network-flow/profiles";
    writeTempContracts({ routeContracts: routePathDrift });
    const routePathResult = runChecker();
    assert.notEqual(routePathResult.status, 0);
    assert.match(routePathResult.stderr, /routes\[1\]\.path/u);

    const unknownPrimaryError = structuredClone(routes);
    unknownPrimaryError.routes[0].primary_errors = ["network_flow_missing"];
    writeTempContracts({ routeContracts: unknownPrimaryError });
    const unknownPrimaryErrorResult = runChecker();
    assert.notEqual(unknownPrimaryErrorResult.status, 0);
    assert.match(unknownPrimaryErrorResult.stderr, /primary_errors/u);

    const openSchema = structuredClone(schemas);
    openSchema.$defs.TableQueryRequest.additionalProperties = true;
    writeTempContracts({ schemaBundle: openSchema });
    const openSchemaResult = runChecker();
    assert.notEqual(openSchemaResult.status, 0);
    assert.match(openSchemaResult.stderr, /additionalProperties must not be true/u);

    const mismatchedSchemaID = structuredClone(schemas);
    mismatchedSchemaID.$defs.TableQueryRequest.properties.schema_id.const =
      "cartulary.network_flow.table_query_request.v0";
    writeTempContracts({ schemaBundle: mismatchedSchemaID });
    const mismatchedSchemaIDResult = runChecker();
    assert.notEqual(mismatchedSchemaIDResult.status, 0);
    assert.match(mismatchedSchemaIDResult.stderr, /schema_id\.const must match/u);

    const missingPublicSchemaID = structuredClone(schemas);
    delete missingPublicSchemaID.$defs.TableQueryRequest.x_schema_id;
    writeTempContracts({ schemaBundle: missingPublicSchemaID });
    const missingPublicSchemaIDResult = runChecker();
    assert.notEqual(missingPublicSchemaIDResult.status, 0);
    assert.match(missingPublicSchemaIDResult.stderr, /public schema IDs mismatch/u);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("network flow timezone provenance schema is closed and immutable-source scoped", async () => {
  const provenance = readJSON(
    "contracts/network-flow/timezone/tzdb-2026c.provenance.json",
  );

  await validateSchema(
    "cartulary.network_flow_timezone_ruleset_provenance.v1",
    provenance,
  );

  await assert.rejects(
    validateSchema("cartulary.network_flow_timezone_ruleset_provenance.v1", {
      ...provenance,
      unexpected: true,
    }),
    /must NOT have additional properties/u,
  );

  const mutableSource = structuredClone(provenance);
  mutableSource.source_archive.url =
    "https://data.iana.org/time-zones/repository/tzdata-latest.tar.gz";
  await assert.rejects(
    validateSchema(
      "cartulary.network_flow_timezone_ruleset_provenance.v1",
      mutableSource,
    ),
    /must be equal to constant/u,
  );

  const hostAuthoritative = structuredClone(provenance);
  hostAuthoritative.conformance_policy.host_timezone_database_authoritative = true;
  await assert.rejects(
    validateSchema(
      "cartulary.network_flow_timezone_ruleset_provenance.v1",
      hostAuthoritative,
    ),
    /must be equal to constant/u,
  );

  const checker = path.join(
    repoRoot,
    "tools/harness/generated-artifacts/check-json-shapes.mjs",
  );
  const artifactPath = path.join(
    repoRoot,
    "contracts/network-flow/timezone/tzdb-2026c.provenance.json",
  );
  const pass = spawnSync(
    process.execPath,
    [checker, "--kind", "network-flow-timezone-provenance", "--file", artifactPath],
    { encoding: "utf8" },
  );
  assert.equal(pass.status, 0, pass.stderr);

  const root = mkdtempSync(path.join(repoRoot, "tmp", "network-flow-tzdb."));
  try {
    const mutableArtifact = structuredClone(provenance);
    mutableArtifact.source_archive.sha256 = "0".repeat(64);
    const mutableArtifactPath = path.join(root, "tzdb.provenance.json");
    writeFileSync(
      mutableArtifactPath,
      `${JSON.stringify(mutableArtifact, null, 2)}\n`,
    );
    const fail = spawnSync(
      process.execPath,
      [
        checker,
        "--kind",
        "network-flow-timezone-provenance",
        "--file",
        mutableArtifactPath,
      ],
      { encoding: "utf8" },
    );
    assert.notEqual(fail.status, 0);
    assert.match(fail.stderr, /source_archive\.sha256/u);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("network flow fault-control schema is closed and boundary-scoped", async () => {
  const response = {
    schema_id: "cartulary.test.network_flow_fault_control.v1",
    fault_id: "fault-1",
    boundary: "network_flow.import.before_transaction_commit",
    fault_kind: "return_error",
    error_code: "network_flow_fault_probe",
    correlation_key: "apply:job-1",
    consume_once: true,
  };

  await validateSchema("cartulary.test.network_flow_fault_control.v1", response);

  await assert.rejects(
    validateSchema("cartulary.test.network_flow_fault_control.v1", {
      ...response,
      unexpected: true,
    }),
    /must NOT have additional properties/u,
  );

  await assert.rejects(
    validateSchema("cartulary.test.network_flow_fault_control.v1", {
      ...response,
      boundary: "network_flow.import.unknown",
    }),
    /must be equal to one of the allowed values/u,
  );

  await assert.rejects(
    validateSchema("cartulary.test.network_flow_fault_control.v1", {
      ...response,
      consume_once: false,
    }),
    /must be equal to constant/u,
  );
});

test("network flow randomness-control schema is closed and stream-scoped", async () => {
  const response = {
    schema_id: "cartulary.test.network_flow_randomness_control.v1",
    control_id: "random-1",
    stream: "network_flow.table_id",
    value_kind: "uuid",
    value_count: 2,
    remaining_count: 2,
    consume_once: true,
    exhaustion: "fail_closed",
  };

  await validateSchema(
    "cartulary.test.network_flow_randomness_control.v1",
    response,
  );

  await assert.rejects(
    validateSchema("cartulary.test.network_flow_randomness_control.v1", {
      ...response,
      unexpected: true,
    }),
    /must NOT have additional properties/u,
  );

  await assert.rejects(
    validateSchema("cartulary.test.network_flow_randomness_control.v1", {
      ...response,
      stream: "network_flow.unknown",
    }),
    /must be equal to one of the allowed values/u,
  );

  await assert.rejects(
    validateSchema("cartulary.test.network_flow_randomness_control.v1", {
      ...response,
      consume_once: false,
    }),
    /must be equal to constant/u,
  );
});

test("network flow auth-transition-control schema is closed and hidden-resource scoped", async () => {
  const response = {
    schema_id: "cartulary.test.network_flow_auth_transition_control.v1",
    control_id: "auth-transition-1",
    boundary: "network_flow.route.after_authorization_before_lookup",
    transition_kind: "incident_membership_revoked",
    actor_ref: "actor:analyst-1",
    incident_ref: "incident:alpha",
    resource_kind: "network_flow_table",
    resource_ref: "network-flow-table:table-1",
    hidden_response_kind: "not_found",
    must_not_disclose_resource: true,
    correlation_key: "query:page-1",
    consume_once: true,
  };

  await validateSchema(
    "cartulary.test.network_flow_auth_transition_control.v1",
    response,
  );

  await assert.rejects(
    validateSchema("cartulary.test.network_flow_auth_transition_control.v1", {
      ...response,
      unexpected: true,
    }),
    /must NOT have additional properties/u,
  );

  await assert.rejects(
    validateSchema("cartulary.test.network_flow_auth_transition_control.v1", {
      ...response,
      hidden_response_kind: "raw_resource",
    }),
    /must be equal to one of the allowed values/u,
  );

  await assert.rejects(
    validateSchema("cartulary.test.network_flow_auth_transition_control.v1", {
      ...response,
      must_not_disclose_resource: false,
    }),
    /must be equal to constant/u,
  );
});

test("network flow audit-assertion-control schema is closed and count-scoped", async () => {
  const response = {
    schema_id: "cartulary.test.network_flow_audit_assertion_control.v1",
    assertion_id: "audit-assertion-1",
    assertion_kind: "no_audit_replay",
    event_code: "network_flow_table_created",
    operation_ref: "import:apply-1",
    actor_ref: "actor:analyst-1",
    incident_ref: "incident:alpha",
    resource_kind: "network_flow_table",
    resource_ref: "network-flow-table:table-1",
    baseline_count: 1,
    expected_final_count: 1,
    expected_replay_increment: 0,
    correlation_key: "apply:job-1",
    consume_once: true,
  };

  await validateSchema(
    "cartulary.test.network_flow_audit_assertion_control.v1",
    response,
  );

  await assert.rejects(
    validateSchema("cartulary.test.network_flow_audit_assertion_control.v1", {
      ...response,
      unexpected: true,
    }),
    /must NOT have additional properties/u,
  );

  await assert.rejects(
    validateSchema("cartulary.test.network_flow_audit_assertion_control.v1", {
      ...response,
      event_code: "network_flow_secret_viewed",
    }),
    /must be equal to one of the allowed values/u,
  );

  await assert.rejects(
    validateSchema("cartulary.test.network_flow_audit_assertion_control.v1", {
      ...response,
      consume_once: false,
    }),
    /must be equal to constant/u,
  );
});

test("test clock-control schema is closed and mode-scoped", async () => {
  const response = {
    schema_id: "cartulary.test.clock_control.v1",
    mode: "fixed",
    now: "2026-11-01T05:30:00.123456789Z",
    offset_seconds: 0,
    fixed_now: "2026-11-01T05:30:00.123456789Z",
  };

  await validateSchema("cartulary.test.clock_control.v1", response);

  await assert.rejects(
    validateSchema("cartulary.test.clock_control.v1", {
      ...response,
      unexpected: true,
    }),
    /must NOT have additional properties/u,
  );

  await assert.rejects(
    validateSchema("cartulary.test.clock_control.v1", {
      ...response,
      mode: "frozen",
    }),
    /must be equal to one of the allowed values/u,
  );

  await assert.rejects(
    validateSchema("cartulary.test.clock_control.v1", {
      ...response,
      schema_id: "cartulary.test.clock_control.v2",
    }),
    /must be equal to constant/u,
  );
});

test("frontend guide target restatements reject stale explicit targets", () => {
  const rowTargets = new Map([
    ["FE-A11Y-P9-01", new Set(["browser-e2e-a11y"])],
  ]);
  const matchingGuideRow =
    "| FE-A11Y-P9-01 | Accessibility | Verify inspector controls. | UI/UX guide Sections 9, 12, 14 | `N/A` | `D-AC-009` | `make browser-e2e-a11y` | `design_direction` |";
  assert.deepEqual(
    collectFrontendGuideTargetRestatementErrors(
      matchingGuideRow,
      rowTargets,
      "guide.md",
    ),
    [],
  );

  const staleGuideRow =
    "| FE-A11Y-P9-01 | Accessibility | Verify inspector controls. | UI/UX guide Sections 9, 12, 14 | `N/A` | `D-AC-009` | `make browser-e2e-a11y-preflight` | `design_direction` |";
  const errors = collectFrontendGuideTargetRestatementErrors(
    staleGuideRow,
    rowTargets,
    "guide.md",
  );
  assert.equal(errors.length, 1);
  assert.match(errors[0], /guide\.md:1/);
  assert.match(errors[0], /FE-A11Y-P9-01/);
  assert.match(errors[0], /browser-e2e-a11y-preflight/);
  assert.match(errors[0], /browser-e2e-a11y/);
});

test("frontend phase-accounting facade does not re-export test-output indexes", () => {
  const facade = readFileSync(
    path.join(repoRoot, "tools/harness/phase-accounting/frontend/index.mjs"),
    "utf8",
  );
  assert.doesNotMatch(facade, /output\/test-output\/frontend-indexes/);
});

test("frontend row evidence normalizes runner observations before accounting", () => {
  const root = mkdtempSync(path.join(repoRoot, "tmp", "frontend-row-evidence."));
  try {
    const vitestRunner = path.join(root, "vitest-runner.json");
    writeFileSync(
      vitestRunner,
      JSON.stringify({
        testResults: [
          {
            name: path.join(
              repoRoot,
              "apps/web/src/workbook/WorkbookShell.phase8.test.tsx",
            ),
            assertionResults: [
              {
                title: "FE-U-P8-01 renders timeline rows",
                status: "passed",
              },
              {
                title: "FE-U-P8-02 persists timeline edits",
                status: "failed",
              },
            ],
          },
        ],
      }),
    );

    const vitestObservations = collectVitestTitleObservations(vitestRunner);
    assert.deepEqual(
      vitestObservations.get("FE-U-P8-01 renders timeline rows"),
      [
        {
          file: "apps/web/src/workbook/WorkbookShell.phase8.test.tsx",
          status: "passed",
        },
      ],
    );
    assert.equal(
      frontendScenarioStatus(
        vitestObservations.get("FE-U-P8-02 persists timeline edits"),
      ),
      "failed",
    );

    const playwrightTargetDir = path.join(root, "browser-e2e-visual");
    mkdirSync(path.join(playwrightTargetDir, "visual"), { recursive: true });
    writeFileSync(
      path.join(playwrightTargetDir, "visual", "phase-summary.json"),
      JSON.stringify({
        inventory: [
          {
            symbol_or_title: "FE-V-P8-01 keeps workbook visual baseline",
            package_or_file: "workbook.visual.spec.ts",
          },
        ],
        failures: [
          {
            symbol_or_title: "FE-V-P8-02 reports visual diff",
            package_or_file: "apps/web/e2e/workbook.visual.spec.ts",
          },
        ],
      }),
    );

    const playwrightObservations =
      collectPlaywrightTitleObservationsForTarget(playwrightTargetDir);
    assert.deepEqual(
      playwrightObservations.get("FE-V-P8-01 keeps workbook visual baseline"),
      [
        {
          file: "apps/web/e2e/workbook.visual.spec.ts",
          status: "passed",
        },
      ],
    );
    assert.deepEqual(
      playwrightObservations.get("FE-V-P8-02 reports visual diff"),
      [
        {
          file: "apps/web/e2e/workbook.visual.spec.ts",
          status: "failed",
        },
      ],
    );

    assert.equal(frontendScenarioStatus([]), "missing");
    assert.equal(
      frontendScenarioStatus([
        { file: "apps/web/e2e/a.spec.ts", status: "skipped" },
      ]),
      "skipped",
    );
    assert.equal(
      frontendScenarioStatus([
        { file: "apps/web/e2e/a.spec.ts", status: "unknown" },
      ]),
      "unknown",
    );
    assert.equal(
      frontendScenarioStatus([
        { file: "apps/web/e2e/a.spec.ts", status: "passed" },
        { file: "apps/web/e2e/b.spec.ts", status: "failed" },
      ]),
      "failed",
    );
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

function runVitestPhaseSummaryFixture({ root, runnerJSON, sidecarJSON = "" }) {
  const phaseDir = path.join(root, sidecarJSON ? "phase-sidecar" : "phase-fallback");
  const resultsDir = path.relative(repoRoot, path.join(root, "results"));
  const result = spawnSync(
    process.execPath,
    [path.join(repoRoot, "tools/harness/output/test-output.mjs"), "vitest-phase"],
    {
      cwd: repoRoot,
      encoding: "utf8",
      env: {
        ...process.env,
        CARTULARY_TEST_RESULTS_DIR: resultsDir,
        CARTULARY_TEST_RUN_ID: "vitest-sidecar-fixture",
        CARTULARY_TEST_TARGET: "frontend-unit",
        CARTULARY_PHASE_DIR: phaseDir,
        CARTULARY_PHASE_LABEL: "frontend-unit vitest",
        CARTULARY_PHASE_COMMAND: "pnpm --dir apps/web exec vitest run",
        CARTULARY_PHASE_START_TIME: "2026-01-01T00:00:00.000Z",
        CARTULARY_PHASE_END_TIME: "2026-01-01T00:00:01.000Z",
        CARTULARY_PHASE_DURATION_MS: "1000",
        CARTULARY_PHASE_EXIT_STATUS: "1",
        CARTULARY_PHASE_RUNNER_LOG: runnerJSON,
        CARTULARY_PHASE_STDOUT_LOG: path.join(root, "stdout.log"),
        CARTULARY_PHASE_STDERR_LOG: path.join(root, "stderr.log"),
        ...(sidecarJSON
          ? { CARTULARY_PHASE_VITEST_FAILURE_DETAILS: sidecarJSON }
          : {}),
      },
    },
  );
  assert.equal(
    result.status,
    1,
    `vitest phase fixture should fail for the synthetic assertion: ${result.stderr}${result.stdout}`,
  );
  return JSON.parse(
    readFileSync(path.join(phaseDir, "phase-summary.json"), "utf8"),
  );
}

function browserWorkerSlotCount(group) {
  if (group?.kind === "functional_shard" || group?.kind === "support") {
    return 1;
  }
  const workers = group?.workers ?? "1";
  if (workers === "default") {
    return 1;
  }
  const parsed = Number.parseInt(String(workers), 10);
  assert.ok(
    Number.isInteger(parsed) && parsed > 0 && String(parsed) === String(workers),
    `browser group ${group?.id} workers must be a positive integer or default`,
  );
  return parsed;
}

function assertBrowserWorkerSlots(units, label) {
  assert.ok(units.length > 0, `${label} must include browser groups`);
  const expectedTotal = units.reduce(
    (sum, unit) => sum + browserWorkerSlotCount(unit.browser_group),
    0,
  );
  const occupied = new Set();
  for (const unit of units) {
    const env = unit.env ?? {};
    assert.equal(
      env.CARTULARY_PLAYWRIGHT_WORKER_COUNT,
      String(expectedTotal),
      `${label} ${unit.id} must receive the service-session worker count`,
    );
    assert.match(
      env.CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET ?? "",
      /^(0|[1-9][0-9]*)$/,
      `${label} ${unit.id} must receive an explicit worker offset`,
    );
    const offset = Number.parseInt(
      env.CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET,
      10,
    );
    const slots = browserWorkerSlotCount(unit.browser_group);
    for (let slot = offset; slot < offset + slots; slot += 1) {
      assert.ok(!occupied.has(slot), `${label} worker slot ${slot} overlaps`);
      occupied.add(slot);
    }
    if (unit.browser_group?.kind === "support") {
      assert.equal(env.PLAYWRIGHT_WORKERS, "1");
    }
  }
  assert.deepEqual(
    [...occupied].sort((left, right) => left - right),
    Array.from({ length: expectedTotal }, (_value, index) => index),
    `${label} worker slots must be contiguous`,
  );
}

test("Vitest failure sidecar overrides STACK_TRACE_ERROR summary fallback", () => {
  const root = mkdtempSync(path.join(repoRoot, "tmp", "vitest-sidecar."));
  try {
    const ownerPath = "apps/web/src/SyntheticSidecar.test.tsx";
    const title = "reports the retained assertion message";
    const runnerJSON = path.join(root, "runner.json");
    const sidecarJSON = path.join(root, "vitest-failure-details.json");
    writeFileSync(path.join(root, "stdout.log"), "");
    writeFileSync(path.join(root, "stderr.log"), "");
    writeFileSync(
      runnerJSON,
      JSON.stringify({
        testResults: [
          {
            name: path.join(repoRoot, ownerPath),
            status: "failed",
            assertionResults: [
              {
                title,
                status: "failed",
                failureMessages: [
                  `Error: STACK_TRACE_ERROR\n    at ${path.join(repoRoot, ownerPath)}:10:1`,
                ],
              },
            ],
          },
        ],
      }),
    );
    writeFileSync(
      sidecarJSON,
      JSON.stringify({
        schema_id: "cartulary.vitest_failure_details.v1",
        runner_json: path.relative(repoRoot, runnerJSON),
        stdout_log: "",
        stderr_log: "",
        generated_at: "2026-01-01T00:00:00.000Z",
        failures: [
          {
            owner_path: ownerPath,
            title,
            status: "failed",
            message: "expected fetchMock to have exactly 2 calls, got 3",
            message_source: "synthetic_sidecar_assertion_message",
            raw_messages: [
              "expected fetchMock to have exactly 2 calls, got 3",
            ],
            diagnostic_tags: ["vitest_stack_trace_error"],
            first_app_frame: `${ownerPath}:10`,
          },
        ],
      }),
    );

    const sidecarSummary = runVitestPhaseSummaryFixture({
      root,
      runnerJSON,
      sidecarJSON,
    });
    assert.equal(
      sidecarSummary.dossiers[0].message,
      "expected fetchMock to have exactly 2 calls, got 3",
    );
    assert.ok(
      sidecarSummary.dossiers[0].diagnostic_tags.includes(
        "vitest_failure_sidecar",
      ),
      "sidecar-backed failures must carry a diagnostic tag",
    );
    assert.match(
      sidecarSummary.dossiers[0].raw,
      /vitest-failure-details\.json/,
      "sidecar-backed failures must retain the sidecar artifact ref",
    );

    const fallbackSummary = runVitestPhaseSummaryFixture({
      root,
      runnerJSON,
    });
    assert.match(
      fallbackSummary.dossiers[0].message,
      /Vitest reporter emitted STACK_TRACE_ERROR/,
      "missing sidecar must keep the current fallback diagnostic",
    );
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

function renderedArtifacts() {
  const topology = loadExecutionTopology();
  const taskSurface = renderTaskSurfaceManifest(topology);
  const browserBatch = renderBrowserBatchManifest(topology);
  const serviceBacked = renderServiceBackedScheduleManifest({
    topology: "tools/execution_topology_manifest.json",
    topologyObject: topology,
  });
  return {
    topology,
    taskSurface,
    browserBatch,
    serviceBacked,
    checkSchedule: renderCheckScheduleManifest(topology),
    expandedCheckSchedule: renderCheckScheduleManifest(topology, {
      serviceBackedScheduleManifest: serviceBacked,
      expandServiceBackedScheduleForCheck,
    }),
    taskSurfaceMake: renderTaskSurfaceMake(taskSurface),
  };
}

function splitMarkdownRow(line) {
  return line
    .slice(1, -1)
    .split("|")
    .map((cell) => cell.trim());
}

function parseHarnessPublicRegistry() {
  const text = readFileSync(
    path.join(repoRoot, "docs/testing-harness-nlspec.md"),
    "utf8",
  );
  const lines = text.split("\n");
  const rows = new Map();
  let inTable = false;
  for (const line of lines) {
    if (line.startsWith("| Target | Command ID | Family ID |")) {
      inTable = true;
      continue;
    }
    if (inTable && line.startsWith("**TH-HARNESS-REQ-059**")) {
      break;
    }
    if (!inTable || !line.startsWith("| `")) {
      continue;
    }
    const cells = splitMarkdownRow(line);
    const target = cells[0].replaceAll("`", "");
    rows.set(target, {
      outputClass: cells[4].replaceAll("`", ""),
      sideEffects: cells[7]
        .split(",")
        .map((entry) => entry.trim().replaceAll("`", ""))
        .filter(Boolean)
        .sort((left, right) => left.localeCompare(right)),
    });
  }
  return rows;
}

function markdownCodeTokens(cell) {
  return [...String(cell ?? "").matchAll(/`([^`]+)`/gu)].map((match) => match[1]);
}

function normalizeSpecAllowedSources(cell) {
  const text = String(cell ?? "").toLowerCase();
  const sources = [];
  if (text.includes("make command line")) {
    sources.push("make_command_line");
  }
  if (text.includes("environment")) {
    sources.push("environment");
  }
  if (text.includes("makefile default")) {
    sources.push("makefile_default");
  }
  if (text.includes("internal default")) {
    sources.push("internal_default");
  }
  if (text.includes("manifest")) {
    sources.push("manifest");
  }
  return sources;
}

function normalizeSpecDefault(cell) {
  const text = String(cell ?? "").trim();
  if (text === "none") {
    return null;
  }
  const token = markdownCodeTokens(text)[0];
  if (token === undefined) {
    return null;
  }
  if (token === "false") {
    return false;
  }
  if (/^(0|[1-9][0-9]*)$/u.test(token)) {
    return Number.parseInt(token, 10);
  }
  if (/^(?:0|[1-9][0-9]*)\.[0-9]+$/u.test(token)) {
    return Number(token);
  }
  return token;
}

function normalizeSpecEmptyString(cell) {
  const text = String(cell ?? "").toLowerCase();
  if (text.includes("false")) {
    return "false";
  }
  if (text.includes("invalid")) {
    return "invalid";
  }
  return "omitted";
}

function normalizeSpecInvalidReason(cell) {
  return markdownCodeTokens(cell)[0] ?? "";
}

function normalizeSpecChildForwarding(cell) {
  return String(cell ?? "").trim().toLowerCase().replaceAll(" ", "_");
}

function normalizeSpecValuesAndBounds(cell, type) {
  const text = String(cell ?? "");
  if (type === "enum") {
    return { values: markdownCodeTokens(text) };
  }
  const range = text.match(/`?([0-9]+)\.\.([0-9]+)`?/u);
  if (range) {
    return {
      min: Number.parseInt(range[1], 10),
      max: Number.parseInt(range[2], 10),
    };
  }
  const min = text.match(/`?>=\s*([0-9]+(?:\.[0-9]+)?)`?/u);
  if (min) {
    return { min: Number(min[1]) };
  }
  return {};
}

function parseHarnessInputMatrix() {
  const text = readFileSync(
    path.join(repoRoot, "docs/testing-harness-nlspec.md"),
    "utf8",
  );
  const lines = text.split("\n");
  const byTarget = new Map();
  let inMatrix = false;
  for (const line of lines) {
    if (line.startsWith("| Target(s) | Input | Type |")) {
      inMatrix = true;
      continue;
    }
    if (inMatrix && line.startsWith("`fixture-report` remains")) {
      break;
    }
    if (!inMatrix || !line.startsWith("| `")) {
      continue;
    }
    const cells = splitMarkdownRow(line);
    const targets = markdownCodeTokens(cells[0]);
    const name = markdownCodeTokens(cells[1])[0];
    const type = markdownCodeTokens(cells[2])[0];
    const entry = {
      name,
      binding: "make_variable",
      allowed_sources: normalizeSpecAllowedSources(cells[4]),
      required: cells[3] === "yes",
      type,
      default: normalizeSpecDefault(cells[5]),
      empty_string: normalizeSpecEmptyString(cells[7]),
      normalization: markdownCodeTokens(cells[8])[0] ?? "",
      invalid_reason: normalizeSpecInvalidReason(cells[10]),
      summary_emission: String(cells[11] ?? "").trim(),
      child_forwarding: normalizeSpecChildForwarding(cells[12]),
      ...normalizeSpecValuesAndBounds(cells[9], type),
    };
    for (const target of targets) {
      if (!byTarget.has(target)) {
        byTarget.set(target, []);
      }
      byTarget.get(target).push(entry);
    }
  }
  return byTarget;
}

function normalizeManifestInput(input) {
  const normalized = {
    name: input.name,
    binding: input.binding,
    allowed_sources: input.allowed_sources,
    required: input.required,
    type: input.type,
    default: input.default,
    empty_string: input.empty_string,
    normalization: input.normalization,
    invalid_reason: input.invalid_reason,
    summary_emission: input.summary_emission,
    child_forwarding: input.child_forwarding,
  };
  if (input.values !== undefined) {
    normalized.values = input.values;
  }
  if (input.min !== undefined) {
    normalized.min = input.min;
  }
  if (input.max !== undefined) {
    normalized.max = input.max;
  }
  return normalized;
}

function normalizeInputList(inputs = []) {
  return inputs
    .map((input) => ({ ...input }))
    .sort((left, right) => left.name.localeCompare(right.name));
}

test("fast harness smoke is role-complete and intentionally small", () => {
  const manifest = readJSON("tools/task_surface_manifest.json");
  const checksByName = new Map(
    manifest.harness_checks.map((check) => [check.name, check]),
  );
  const fastChecks = manifest.harness_tiers.fast.checks;
  assert.deepEqual(fastChecks, [
    "harness-smoke-public-make-wrapper",
    "harness-smoke-check-scheduler-smoke",
    "harness-smoke-service-backed-scheduler-smoke",
  ]);
  assert.deepEqual(
    fastChecks.map((name) => checksByName.get(name)?.gate_smoke_role),
    [
      "public_make_wrapper",
      "check_scheduler_semantic",
      "service_backed_scheduler_semantic",
    ],
  );
});

test("fast harness smoke scratch fixtures stay outside the repository", () => {
  const manifest = readJSON("tools/task_surface_manifest.json");
  const checksByName = new Map(
    manifest.harness_checks.map((check) => [check.name, check]),
  );
  const fastShellScripts = new Set(
    manifest.harness_tiers.fast.checks.flatMap((name) =>
      (checksByName.get(name)?.backing_scripts ?? []).filter((script) =>
        script.endsWith(".sh"),
      ),
    ),
  );
  const requiredScratchHelpers = new Map([
    [
      "tools/harness/tests/test-public-make-wrapper-smoke.sh",
      ["cartulary_harness_mktemp_dir \"public-make-wrapper.XXXXXX\""],
    ],
    [
      "tools/harness/scheduler/tests/test-check-scheduler.sh",
      [
        "cartulary_harness_mktemp_dir \"check-scheduler-smoke.XXXXXX\"",
        "cartulary_harness_mktemp_dir \"check-scheduler-smoke-service-timing.XXXXXX\"",
      ],
    ],
    [
      "tools/harness/scheduler/tests/test-service-backed-scheduler.sh",
      [
        "cartulary_harness_mktemp_dir \"service-backed-scheduler-smoke.XXXXXX\"",
      ],
    ],
  ]);
  for (const script of fastShellScripts) {
    const content = readFileSync(path.join(repoRoot, script), "utf8");
    assert.match(
      content,
      /tools\/harness\/test-support\/harness-scratch\.sh|test-support\/harness-scratch\.sh/,
      `${script} must source harness-scratch.sh for fast smoke scratch`,
    );
    for (const required of requiredScratchHelpers.get(script) ?? []) {
      assert.ok(content.includes(required), `${script} missing ${required}`);
    }
    assert.doesNotMatch(
      content,
      /mktemp -d "\$\{?ROOT_DIR\}?\/(?:tmp|\.cartulary\/tmp)\/(?:public-make-wrapper|check-scheduler-smoke|check-scheduler-smoke-service-timing|service-backed-scheduler-smoke)\.XXXXXX"/,
      `${script} must not create fast smoke fixtures with raw mktemp templates`,
    );
    assert.doesNotMatch(
      content,
      /CARTULARY_HARNESS_SCRATCH_ROOT:-\$\{ROOT_DIR\}\/\.cartulary\/tmp/,
      `${script} must not default harness scratch inside the repository`,
    );
  }
});

test("dev service lifecycle guards are mutation-safe", () => {
  const result = spawnSync("bash", ["./tools/harness/readiness/tests/test-dev-services-lifecycle.sh"], {
    cwd: repoRoot,
    encoding: "utf8",
    env: { ...process.env },
    maxBuffer: 16 * 1024 * 1024,
  });
  assert.equal(
    result.status,
    0,
    [
      "tools/harness/readiness/tests/test-dev-services-lifecycle.sh failed",
      "--- stdout ---",
      result.stdout,
      "--- stderr ---",
      result.stderr,
    ].join("\n"),
  );
});

test("task-surface validation rejects fast harness smoke drift", () => {
  const { taskSurface, browserBatch, serviceBacked } = renderedArtifacts();
  const invalid = structuredClone(taskSurface);
  invalid.harness_tiers.fast.checks.push("harness-smoke-execution-topology");
  assert.match(
    collectTaskSurfaceManifestErrors(invalid, {
      browserBatchManifest: browserBatch,
      serviceBackedScheduleManifest: serviceBacked,
    }).join("\n"),
    /harness_tiers\.fast\.checks must contain exactly 3 gate smoke checks/,
  );

  const missingRole = structuredClone(taskSurface);
  delete missingRole.harness_checks.find(
    (check) => check.name === "harness-smoke-public-make-wrapper",
  ).gate_smoke_role;
  assert.match(
    collectTaskSurfaceManifestErrors(missingRole, {
      browserBatchManifest: browserBatch,
      serviceBackedScheduleManifest: serviceBacked,
    }).join("\n"),
    /harness-smoke-public-make-wrapper\.gate_smoke_role is required for fast harness smoke/,
  );
});

test("task-surface validation rejects Node-backed wrappers without Node readiness", () => {
  const { taskSurface, browserBatch, serviceBacked } = renderedArtifacts();
  for (const target of ["check", "check-harness-smoke"]) {
    const invalid = structuredClone(taskSurface);
    invalid.make_recipes[target].prerequisites = [];
    assert.match(
      collectTaskSurfaceManifestErrors(invalid, {
        browserBatchManifest: browserBatch,
        serviceBackedScheduleManifest: serviceBacked,
      }).join("\n"),
      new RegExp(
        `make_recipes\\.${target}\\.prerequisites must include \\$\\(NODE_BIN\\)`,
      ),
      `${target} must require explicit Node readiness`,
    );
  }
});

test("full harness tier composes fast, extended, lifecycle, and full-only diagnostics", () => {
  const manifest = readJSON("tools/task_surface_manifest.json");
  const fullOnlyChecks = new Set(["harness-smoke-tool-output-real-targets"]);
  const expectedFullBase = [
    ...manifest.harness_tiers.fast.checks,
    ...manifest.harness_tiers.extended.checks,
    ...manifest.harness_tiers.lifecycle.checks,
  ];
  const filteredFull = manifest.harness_tiers.full.checks.filter(
    (check) => !fullOnlyChecks.has(check),
  );
  assert.deepEqual(filteredFull, expectedFullBase);
  for (const check of fullOnlyChecks) {
    assert.ok(manifest.harness_tiers.full.checks.includes(check));
  }
});

test("execution harness smoke is a narrow execution-wrapper subset", () => {
  const manifest = readJSON("tools/task_surface_manifest.json");
  const expectedExecutionChecks = [
    "harness-smoke-run-make-sequence-fast",
    "harness-smoke-cartulary-runner-service-backed-target",
    "harness-smoke-make-node-tools",
    "harness-smoke-run-phase",
    "harness-smoke-run-vitest-phase",
    "harness-smoke-run-vitest-manifest-phase",
    "harness-smoke-run-frontend-unit",
  ];
  assert.deepEqual(manifest.harness_tiers.execution.checks, expectedExecutionChecks);
  const extendedChecks = new Set(manifest.harness_tiers.extended.checks);
  for (const check of expectedExecutionChecks) {
    assert.ok(extendedChecks.has(check), `${check} must remain in extended smoke`);
  }
  const target = manifest.targets.find((entry) => entry.name === "run-harness-smoke-execution");
  assert.equal(target?.target_class, "internal_helper");
  assert.deepEqual(target?.default_inclusion_sets, []);
});

test("check scheduler restores node packages before run-phase validation", () => {
  const { checkSchedule, taskSurface } = renderedArtifacts();
  const schedule = checkSchedule.schedules.find(
    (entry) => entry.target === "check",
  );
  assert.deepEqual(
    taskSurface.make_recipes.check.prerequisites,
    ["$(NODE_BIN)"],
    "check wrapper must only bootstrap the scheduler Node runtime",
  );
  const checkFrontendInstall = schedule.work_units.find(
    (unit) => unit.target === "check-frontend-install",
  );
  const toolchainDrift = schedule.work_units.find(
    (unit) => unit.target === "toolchain-drift",
  );
  const jsonShapeCheck = schedule.work_units.find(
    (unit) => unit.target === "json-shape-check",
  );
  assert.deepEqual(
    schedule.work_units
      .filter((unit) => (unit.needs ?? []).length === 0)
      .map((unit) => unit.target)
      .sort(),
    ["check-frontend-install"],
    "check-frontend-install must be the only dependency-free check unit",
  );
  assert.ok(
    checkFrontendInstall,
    "check schedule must include check-frontend-install",
  );
  assert.deepEqual(
    checkFrontendInstall.needs ?? [],
    [],
    "check-frontend-install must be able to run before run-phase children",
  );
  assert.ok(toolchainDrift, "check schedule must include toolchain-drift");
  assert.ok(
    (toolchainDrift.needs ?? []).includes("check-frontend-install"),
    "toolchain-drift must wait for installed node package dependencies",
  );
  assert.ok(jsonShapeCheck, "check schedule must include json-shape-check");
  assert.ok(
    (jsonShapeCheck.needs ?? []).includes("check-frontend-install"),
    "json-shape-check must wait for installed node package dependencies",
  );
  for (const target of [
    "harness-contract",
    "harness-contract-tests",
    "toolchain-drift",
    "json-shape-check",
  ]) {
    assert.ok(
      taskSurface.make_recipes[target].prerequisites.includes(
        "$(FRONTEND_INSTALL_STAMP)",
      ),
      `${target} direct wrapper must bootstrap installed node package dependencies`,
    );
  }
});

test("check scheduler defers own schema validation until package readiness", () => {
  const schedulerScript = readFileSync(
    path.join(repoRoot, "tools/harness/scheduler/check-schedule-cli.mjs"),
    "utf8",
  );
  const schedulerEngine = readFileSync(
    path.join(repoRoot, "tools/harness/scheduler/scheduler/engine.mjs"),
    "utf8",
  );
  assert.match(
    schedulerScript,
    /schemaValidationEnabled:\s*!deferSchemaValidationForPackageReadiness/u,
  );
  assert.match(
    schedulerScript,
    /context\.reporter\.setSchemaValidationEnabled\(true\)/u,
  );
  assert.match(schedulerEngine, /deferredSchemaRecords/u);
});

test("generated task surface and Make wrapper keep harness projection wiring", () => {
  const { taskSurface, browserBatch, serviceBacked, taskSurfaceMake } =
    renderedArtifacts();
  assert.deepEqual(
    collectTaskSurfaceManifestErrors(taskSurface, {
      browserBatchManifest: browserBatch,
      serviceBackedScheduleManifest: serviceBacked,
    }),
    [],
  );
  assert.deepEqual(taskSurface.make_recipes.check?.prerequisites, [
    "$(NODE_BIN)",
  ]);
  assert.deepEqual(
    taskSurface.make_recipes["check-harness-smoke"]?.prerequisites,
    ["$(NODE_BIN)"],
  );
  const checkBlock = targetRecipeBlock(taskSurfaceMake, "check").join("\n");
  const preflightIndex = checkBlock.indexOf(
    "$(RUN_HARNESS_PREFLIGHT) check",
  );
  const prerequisiteIndex = checkBlock.indexOf(
    "$(MAKE) --silent --no-print-directory $(NODE_BIN); fi",
  );
  const schedulerIndex = checkBlock.indexOf(
    "$(NODE_BIN) $(RUN_CHECK_SCHEDULE_SCRIPT)",
  );
  assert.ok(preflightIndex >= 0, "check must render public preflight");
  assert.ok(
    prerequisiteIndex >= 0,
    "check must render Node readiness before owner CLI preflight",
  );
  assert.ok(
    preflightIndex > prerequisiteIndex,
    "check must run owner CLI preflight after Node readiness",
  );
  assert.ok(
    schedulerIndex > preflightIndex,
    "check must launch the scheduler after preflight",
  );
  assert.equal(
    taskSurface.make_recipes["check-harness-smoke"]?.child_target,
    "run-harness-smoke-fast",
  );
  assert.equal(
    taskSurface.make_recipes["check-harness-smoke"]?.projection,
    "check-harness-smoke",
  );
  assert.match(
    taskSurfaceMake,
    /\$\(RUN_HARNESS_SMOKE_SCRIPT\) --tier fast --jobs "\$\(HARNESS_SMOKE_JOBS\)"/,
  );
  assert.match(
    taskSurfaceMake,
    /summary-target --target check-harness-smoke --child-target run-harness-smoke-fast --status pass/,
  );
  const targetEntries = new Map(
    taskSurface.targets.map((entry) => [entry.name, entry]),
  );
  const producedSummaryTargets = new Set(
    Object.values(taskSurface.sequences ?? {}).flatMap((sequence) =>
      (sequence.steps ?? []).flatMap(
        (step) => step.produces_summary_targets ?? [],
      ),
    ),
  );
  for (const target of producedSummaryTargets) {
    const recipe = taskSurface.make_recipes[target];
    const entry = targetEntries.get(target);
    if (
      recipe?.mode !== "run_phase" ||
      entry?.output_policy?.summary_schema !== "cartulary.tool_run_summary.v3"
    ) {
      continue;
    }
    assert.match(
      taskSurfaceMake,
      new RegExp(`RUN_RETAINED_TARGET_SUMMARY,${target},pass`),
      `${target} run_phase recipe must retain a passing target summary`,
    );
    assert.match(
      taskSurfaceMake,
      new RegExp(`RUN_RETAINED_TARGET_SUMMARY,${target},fail`),
      `${target} run_phase recipe must retain a failing target summary`,
    );
  }
});

test("harness NLSpec registry mirrors public target output classes and side effects", () => {
  const { taskSurface } = renderedArtifacts();
  const specRows = parseHarnessPublicRegistry();
  const publicTargets = taskSurface.targets.filter(
    (entry) => entry.target_class === "public",
  );
  assert.equal(
    specRows.size,
    publicTargets.length,
    "NLSpec public target registry row count must match manifest public target count",
  );
  for (const target of publicTargets) {
    const spec = specRows.get(target.name);
    assert.ok(spec, `${target.name} must appear in the NLSpec public registry`);
    assert.equal(
      spec.outputClass,
      target.output_policy.output_class,
      `${target.name} output class must match NLSpec registry`,
    );
    assert.deepEqual(
      spec.sideEffects,
      target.side_effects
        .map((entry) => entry.class)
        .sort((left, right) => left.localeCompare(right)),
      `${target.name} side effects must match NLSpec registry`,
    );
  }
});

test("harness NLSpec input matrix mirrors public target input contracts", () => {
  const { taskSurface } = renderedArtifacts();
  const specInputs = parseHarnessInputMatrix();
  const publicTargets = taskSurface.targets.filter(
    (entry) => entry.target_class === "public",
  );
  for (const target of publicTargets) {
    const expected = normalizeInputList(specInputs.get(target.name) ?? []);
    const actual = normalizeInputList(
      (target.input_contract?.inputs ?? []).map(normalizeManifestInput),
    );
    assert.deepEqual(
      actual,
      expected,
      `${target.name} input_contract must match the NLSpec input matrix`,
    );
  }

  const synthetic = structuredClone(taskSurface);
  const drift = synthetic.targets.find(
    (target) => target.name === "scheduler-summary-timing-drift",
  );
  drift.input_contract.inputs.find(
    (input) => input.name === "SCHEDULER_WARM_CHECK_BUDGET_MS",
  ).default = null;
  const expected = normalizeInputList(
    specInputs.get("scheduler-summary-timing-drift") ?? [],
  );
  assert.notDeepEqual(
    normalizeInputList(drift.input_contract.inputs.map(normalizeManifestInput)),
    expected,
    "matrix parity check must detect implementation default drift",
  );
});

test("scheduler manifest exercises every required command type", () => {
  const schedulerManifest = readJSON("tools/scheduler_manifest.json");
  const expected = new Set([
    "make_target",
    "service_session_start",
    "browser_stage_session_start",
    "browser_group",
    "browser_stage_complete",
    "go_shard",
    "go_shard_finalize",
    "service_complete",
  ]);
  for (const schedule of schedulerManifest.schedules ?? []) {
    for (const unit of schedule.work_units ?? []) {
      expected.delete(unit.command?.type);
    }
  }
  assert.deepEqual([...expected].sort(), [], "every required scheduler command type must have a live fixture");
});

test("scheduler family facade matches schema registry and generated manifests", () => {
  const expectedFamilies = [...schedulerFamilyValues];
  const expectedFamilySet = new Set(expectedFamilies);
  const schedulerSchema = readJSON(
    "tools/schemas/cartulary.scheduler_manifest.v2.schema.json",
  );
  assert.deepEqual(
    schedulerSchema.$defs.schedule.properties.scheduler_kind.enum,
    expectedFamilies,
    "scheduler manifest schema enum must match scheduler family facade",
  );

  const resourceRegistry = readJSON("tools/scheduler_resource_registry.json");
  validateSchedulerResourceRegistrySemantics(
    resourceRegistry,
    "tools/scheduler_resource_registry.json",
  );
  const registryFamilyReferences = [];
  for (const resource of resourceRegistry.resources ?? []) {
    for (const scheduler of resource.schedulers ?? []) {
      registryFamilyReferences.push([`resource ${resource.name}`, scheduler]);
    }
  }
  for (const template of resourceRegistry.templates ?? []) {
    for (const scheduler of template.schedulers ?? []) {
      registryFamilyReferences.push([`template ${template.name}`, scheduler]);
    }
  }
  for (const profile of resourceRegistry.capacity_profiles ?? []) {
    registryFamilyReferences.push([
      `capacity profile ${profile.name}`,
      profile.scheduler,
    ]);
  }
  for (const [source, scheduler] of registryFamilyReferences) {
    assert.ok(
      expectedFamilySet.has(scheduler),
      `${source} references unknown scheduler family ${scheduler}`,
    );
  }
  assert.deepEqual(
    (resourceRegistry.capacity_profiles ?? [])
      .map((profile) => profile.name)
      .sort(),
    [...schedulerCapacityProfileValues].sort(),
    "registry capacity profiles must match the scheduler facade",
  );
  for (const [family, profiles] of Object.entries(schedulerCapacityProfilesByFamily)) {
    for (const profileName of profiles) {
      assert.equal(
        schedulerFamilyForCapacityProfile(profileName),
        family,
        `${profileName} must map back to ${family}`,
      );
    }
  }

  const schedulerManifest = readJSON("tools/scheduler_manifest.json");
  for (const schedule of schedulerManifest.schedules ?? []) {
    assert.ok(
      expectedFamilySet.has(schedule.scheduler_kind),
      `${schedule.target} references unknown scheduler family ${schedule.scheduler_kind}`,
    );
    assert.equal(
      schedulerFamilyForCapacityProfile(schedule.capacity_profile),
      schedule.scheduler_kind,
      `${schedule.target} capacity profile must match scheduler_kind`,
    );
  }
});

test("service-backed Go shard units are executable by their declared targets", () => {
  const { serviceBacked } = renderedArtifacts();
  const shardNamesByTarget = new Map();
  function shardsForTarget(target) {
    if (!shardNamesByTarget.has(target)) {
      shardNamesByTarget.set(
        target,
        new Set(collectGoShardsForTarget(repoRoot, target).map((shard) => shard.name)),
      );
    }
    return shardNamesByTarget.get(target);
  }

  for (const schedule of serviceBacked.schedules ?? []) {
    for (const unit of schedule.work_units ?? []) {
      if (unit.kind !== "go_shard") {
        continue;
      }
      assert.ok(
        shardsForTarget(unit.target).has(unit.shard),
        `${schedule.target ?? schedule.name} schedules ${unit.shard} for ${unit.target}, but that target cannot execute the shard`,
      );
    }
  }
});

test("Go shard finalizers declare exactly their selected shard needs", () => {
  const schedulerManifest = readJSON("tools/scheduler_manifest.json");
  for (const schedule of schedulerManifest.schedules ?? []) {
    const completionKeys = new Set(
      (schedule.work_units ?? []).flatMap((unit) => unit.completion_keys ?? [unit.id]),
    );
    for (const unit of schedule.work_units ?? []) {
      if (unit.command?.type !== "go_shard_finalize") {
        continue;
      }
      assert.ok(
        Array.isArray(unit.shard_names) && unit.shard_names.length > 0,
        `${schedule.target} ${unit.id} must declare selected shard_names`,
      );
      const expectedNeeds = unit.shard_names.map((shardName) => `go_shard:${shardName}`);
      for (const expectedNeed of expectedNeeds) {
        assert.ok(
          (unit.needs ?? []).includes(expectedNeed),
          `${schedule.target} ${unit.id} missing selected shard need ${expectedNeed}`,
        );
        assert.ok(
          completionKeys.has(expectedNeed),
          `${schedule.target} ${unit.id} selected shard need ${expectedNeed} has no producer`,
        );
      }
      for (const need of (unit.needs ?? []).filter((entry) => entry.startsWith("go_shard:"))) {
        assert.ok(
          expectedNeeds.includes(need),
          `${schedule.target} ${unit.id} has shard need ${need} outside shard_names`,
        );
      }
    }
  }
});

function targetRecipeBlock(renderedMake, target) {
  const lines = renderedMake.split("\n");
  const headerPattern = new RegExp(`^${target.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")}:`);
  const start = lines.findIndex(
    (line) => headerPattern.test(line) && !line.includes(": export "),
  );
  assert.notEqual(start, -1, `${target} must have a rendered recipe`);
  const block = [];
  for (let index = start; index < lines.length; index += 1) {
    if (index > start && /^[A-Za-z0-9_.-]+:/.test(lines[index])) {
      break;
    }
    block.push(lines[index]);
  }
  return block;
}

test("public targets declare command identity and semantic value", () => {
  const { taskSurface, browserBatch, serviceBacked } = renderedArtifacts();
  const publicTargets = taskSurface.targets.filter(
    (entry) => entry.target_class === "public",
  );
  assert.ok(publicTargets.length > 0, "public target registry must not be empty");
  const commandIDs = new Set();
  for (const target of publicTargets) {
    assert.match(
      target.command_id,
      /^cartulary\.harness\.command\.[a-z][a-z0-9_]*\.v1$/,
      `${target.name} must declare stable command_id`,
    );
    assert.ok(!commandIDs.has(target.command_id), `${target.name} command_id must be unique`);
    commandIDs.add(target.command_id);
    assert.match(
      target.family_id,
      /^[a-z][a-z0-9_]*$/,
      `${target.name} must declare family_id`,
    );
    assert.equal(
      target.lifecycle_state,
      "public_active",
      `${target.name} must declare a current public lifecycle state`,
    );
    assert.ok(
      Array.isArray(target.semantic_behaviors) &&
        target.semantic_behaviors.length > 0,
      `${target.name} must declare semantic behaviors`,
    );
    assert.ok(
      Array.isArray(target.side_effects) && target.side_effects.length > 0,
      `${target.name} must declare side effects`,
    );
    assert.ok(
      target.input_contract &&
        target.input_contract.undeclared_make_command_line === "usage_error" &&
        target.input_contract.undeclared_inherited_env === "ignore" &&
        Array.isArray(target.input_contract.inputs),
      `${target.name} must declare a closed public input contract`,
    );
    for (const entry of target.semantic_behaviors) {
      assert.match(entry.owner_section, /^Section (?:[1-9]|1[0-9])(?:\.[0-9]+)?$/);
    }
    for (const entry of target.side_effects) {
      assert.match(entry.owner_section, /^Section (?:[1-9]|1[0-9])(?:\.[0-9]+)?$/);
    }
  }

  const duplicateID = structuredClone(taskSurface);
  duplicateID.targets.find((entry) => entry.name === "help-all").command_id =
    duplicateID.targets.find((entry) => entry.name === "help").command_id;
  assert.match(
    collectTaskSurfaceManifestErrors(duplicateID, {
      browserBatchManifest: browserBatch,
      serviceBackedScheduleManifest: serviceBacked,
    }).join("\n"),
    /help-all\.command_id duplicates help/,
  );

  const malformedID = structuredClone(taskSurface);
  malformedID.targets.find((entry) => entry.name === "help").command_id =
    "cartulary.harness.command.help.latest";
  assert.match(
    collectTaskSurfaceManifestErrors(malformedID, {
      browserBatchManifest: browserBatch,
      serviceBackedScheduleManifest: serviceBacked,
    }).join("\n"),
    /help\.command_id must match cartulary\.harness\.command\.<name>\.v1/,
  );

  const missingSemantic = structuredClone(taskSurface);
  missingSemantic.targets.find((entry) => entry.name === "help").semantic_behaviors = [];
  assert.match(
    collectTaskSurfaceManifestErrors(missingSemantic, {
      browserBatchManifest: browserBatch,
      serviceBackedScheduleManifest: serviceBacked,
    }).join("\n"),
    /help\.semantic_behaviors must declare at least one semantic behavior/,
  );

  const missingOwner = structuredClone(taskSurface);
  missingOwner.targets.find((entry) => entry.name === "help").semantic_behaviors = [
    { behavior: "diagnostic_synthesis", owner_section: "" },
  ];
  assert.match(
    collectTaskSurfaceManifestErrors(missingOwner, {
      browserBatchManifest: browserBatch,
      serviceBackedScheduleManifest: serviceBacked,
    }).join("\n"),
    /help\.semantic_behaviors\[1\]\.owner_section must be a Section reference/,
  );

  const missingSideEffects = structuredClone(taskSurface);
  delete missingSideEffects.targets.find((entry) => entry.name === "help").side_effects;
  assert.match(
    collectTaskSurfaceManifestErrors(missingSideEffects, {
      browserBatchManifest: browserBatch,
      serviceBackedScheduleManifest: serviceBacked,
    }).join("\n"),
    /help\.side_effects must declare at least one side-effect class/,
  );

  const missingInputContract = structuredClone(taskSurface);
  delete missingInputContract.targets.find((entry) => entry.name === "help").input_contract;
  assert.match(
    collectTaskSurfaceManifestErrors(missingInputContract, {
      browserBatchManifest: browserBatch,
      serviceBackedScheduleManifest: serviceBacked,
    }).join("\n"),
    /help\.input_contract must be declared for public targets/,
  );

  const misplacedInputPolicy = structuredClone(taskSurface);
  misplacedInputPolicy.targets.find(
    (entry) => entry.name === "target-plan",
  ).input_contract.undeclared_make_command_line = "ignore";
  assert.match(
    collectTaskSurfaceManifestErrors(misplacedInputPolicy, {
      browserBatchManifest: browserBatch,
      serviceBackedScheduleManifest: serviceBacked,
    }).join("\n"),
    /target-plan\.input_contract\.undeclared_make_command_line must be usage_error/,
  );

  const invalidSideEffects = structuredClone(taskSurface);
  invalidSideEffects.targets.find((entry) => entry.name === "help").side_effects = [
    { class: "none", owner_section: "Section 4" },
    { class: "retained_artifacts", owner_section: "Section 8" },
  ];
  assert.match(
    collectTaskSurfaceManifestErrors(invalidSideEffects, {
      browserBatchManifest: browserBatch,
      serviceBackedScheduleManifest: serviceBacked,
    }).join("\n"),
    /help\.side_effects\[2\]\.artifact_policy must be declared for retained_artifacts[\s\S]*help\.side_effects none is mutually exclusive with other classes/,
  );

  const duplicateSideEffects = structuredClone(taskSurface);
  duplicateSideEffects.targets.find((entry) => entry.name === "format").side_effects.push(
    structuredClone(
      duplicateSideEffects.targets
        .find((entry) => entry.name === "format")
        .side_effects.find((entry) => entry.class === "authored_source_write"),
    ),
  );
  assert.match(
    collectTaskSurfaceManifestErrors(duplicateSideEffects, {
      browserBatchManifest: browserBatch,
      serviceBackedScheduleManifest: serviceBacked,
    }).join("\n"),
    /format\.side_effects contains duplicate authored_source_write/,
  );

  const legacyFields = structuredClone(taskSurface);
  const legacyHelp = legacyFields.targets.find((entry) => entry.name === "help");
  legacyHelp.classification = "public";
  legacyHelp.included_in = ["helper_only"];
  assert.match(
    collectTaskSurfaceManifestErrors(legacyFields, {
      browserBatchManifest: browserBatch,
      serviceBackedScheduleManifest: serviceBacked,
    }).join("\n"),
    /help\.classification is obsolete; use target_class[\s\S]*help\.included_in is obsolete; use default_inclusion_sets/,
  );

  const privateHelperOnly = structuredClone(taskSurface);
  privateHelperOnly.targets.find(
    (entry) => entry.name === "run-harness-smoke-fast",
  ).default_inclusion_sets = ["helper_only"];
  assert.match(
    collectTaskSurfaceManifestErrors(privateHelperOnly, {
      browserBatchManifest: browserBatch,
      serviceBackedScheduleManifest: serviceBacked,
    }).join("\n"),
    /run-harness-smoke-fast\.default_inclusion_sets helper_only is only valid for public direct-invocation targets/,
  );

  const shallowAlias = structuredClone(taskSurface);
  shallowAlias.targets.push({
    name: "synthetic-shallow-wrapper",
    target_class: "public",
    default_inclusion_sets: ["helper_only"],
    family_id: "help_discovery",
    lifecycle_state: "public_active",
    command_id: "cartulary.harness.command.synthetic_shallow_wrapper.v1",
    semantic_behaviors: [],
    side_effects: [{ class: "none", owner_section: "Section 4" }],
    output_policy: structuredClone(
      shallowAlias.targets.find((entry) => entry.name === "help").output_policy,
    ),
  });
  shallowAlias.help_tiers[0].entries.push({
    target: "synthetic-shallow-wrapper",
    description: "synthetic shallow wrapper",
  });
  shallowAlias.make_recipes["synthetic-shallow-wrapper"] = {
    type: "alias",
    prerequisites: ["help"],
  };
  assert.match(
    collectTaskSurfaceManifestErrors(shallowAlias, {
      browserBatchManifest: browserBatch,
      serviceBackedScheduleManifest: serviceBacked,
    }).join("\n"),
    /synthetic-shallow-wrapper\.semantic_behaviors must declare at least one semantic behavior/,
  );
});

test("public non-interactive wrappers run preflight before child work", () => {
  const { taskSurface, taskSurfaceMake } = renderedArtifacts();
  const recipes = taskSurface.make_recipes;
  for (const target of taskSurface.targets) {
    if (
      target.target_class !== "public" ||
      target.output_policy?.output_class === "interactive_raw"
    ) {
      continue;
    }
    const block = targetRecipeBlock(taskSurfaceMake, target.name);
    const recipeLines = block.filter((line) => line.startsWith("\t"));
    assert.ok(recipeLines.length > 0, `${target.name} must render recipe lines`);
    const escapedName = target.name.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
    if (target.name === "bootstrap-node-runtime") {
      assert.match(
        recipeLines[0],
        /\$\(MAKE\) --silent --no-print-directory \$\(NODE_BIN\)/,
        "bootstrap-node-runtime must provision Node through the owner readiness path",
      );
      assert.ok(
        !recipeLines.some((line) => line.includes("$(RUN_HARNESS_PREFLIGHT)")),
        "bootstrap-node-runtime must not require Node-backed preflight before Node exists",
      );
      continue;
    }
    const preflightLineIndex = recipeLines.findIndex((line) =>
      new RegExp(
        `^\\t\\$\\(Q\\)env .* \\$\\(RUN_HARNESS_PREFLIGHT\\) ${escapedName}$`,
      ).test(line),
    );
    assert.match(
      recipeLines[preflightLineIndex] ?? "",
      new RegExp(`\\$\\(RUN_HARNESS_PREFLIGHT\\) ${escapedName}$`),
      `${target.name} must run public preflight before child work`,
    );
    const prerequisites = recipes[target.name]?.prerequisites ?? [];
    if (prerequisites.includes("$(NODE_BIN)")) {
      assert.match(
        recipeLines.slice(0, preflightLineIndex).join("\n"),
        /\$\(MAKE\) --silent --no-print-directory \$\(NODE_BIN\)/,
        `${target.name} must make Node ready before owner CLI preflight`,
      );
    }
    if (prerequisites.some((prerequisite) => prerequisite !== "$(NODE_BIN)")) {
      assert.match(
        recipeLines.slice(preflightLineIndex + 1).join("\n"),
        /CARTULARY_CHECK_SCHEDULER_SKIP_PREREQUISITES/,
        `${target.name} prerequisite work must follow preflight`,
      );
    }
  }
});

test("per-target input contract rejects misplaced Make variables and ignores ambient env", () => {
  assert.throws(
    () =>
      preflightPublicTarget("target-plan", {
        PHASE: "phase4",
        CARTULARY_MAKE_ORIGIN_PHASE: "command line",
      }),
    (error) =>
      error instanceof HarnessConfigError &&
      error.failure_reason === "usage_error" &&
      /PHASE is not declared for target target-plan/.test(error.message),
  );
  assert.doesNotThrow(() =>
    preflightPublicTarget("target-plan", {
      PHASE: "phase4",
      CARTULARY_MAKE_ORIGIN_PHASE: "environment",
    }),
  );
  assert.throws(
    () =>
      preflightPublicTarget("target-plan", {
        TASK_SURFACE_MANIFEST: "/tmp/override.json",
        CARTULARY_MAKE_ORIGIN_TASK_SURFACE_MANIFEST: "command line",
      }),
    (error) =>
      error instanceof HarnessConfigError &&
      error.failure_reason === "configuration_error" &&
      /TASK_SURFACE_MANIFEST is an internal harness input/.test(error.message),
  );
  assert.doesNotThrow(() =>
    preflightPublicTarget("target-plan", {
      TASK_SURFACE_MANIFEST: "/tmp/override.json",
      CARTULARY_MAKE_ORIGIN_TASK_SURFACE_MANIFEST: "environment",
    }),
  );
  assert.throws(
    () =>
      preflightPublicTarget("target-plan", {
        OPERATOR_BIN: "/tmp/operator",
        CARTULARY_MAKE_ORIGIN_OPERATOR_BIN: "command line",
      }),
    (error) =>
      error instanceof HarnessConfigError &&
      error.failure_reason === "usage_error" &&
      /OPERATOR_BIN is not declared for target target-plan/.test(error.message),
  );
  assert.throws(
    () =>
      preflightPublicTarget("target-plan", {
        CARTULARY_OPERATOR_BIN: "/tmp/operator",
        CARTULARY_MAKE_ORIGIN_CARTULARY_OPERATOR_BIN: "command line",
      }),
    (error) =>
      error instanceof HarnessConfigError &&
      error.failure_reason === "configuration_error" &&
      /CARTULARY_OPERATOR_BIN is an internal harness input/.test(error.message),
  );
  assert.doesNotThrow(() =>
    preflightPublicTarget("target-plan", {
      OPERATOR_BIN: "/tmp/operator",
      CARTULARY_MAKE_ORIGIN_OPERATOR_BIN: "environment",
      CARTULARY_OPERATOR_BIN: "/tmp/operator",
      CARTULARY_MAKE_ORIGIN_CARTULARY_OPERATOR_BIN: "environment",
    }),
  );
  assert.doesNotThrow(() =>
    preflightPublicTarget("frontend-unit", {
      VITEST_MAX_WORKERS: "4",
      CARTULARY_MAKE_ORIGIN_VITEST_MAX_WORKERS: "command line",
    }),
  );
  assert.throws(
    () =>
      preflightPublicTarget("frontend-unit", {
        VITEST_MAX_WORKERS: "17",
        CARTULARY_MAKE_ORIGIN_VITEST_MAX_WORKERS: "command line",
      }),
    (error) =>
      error instanceof HarnessConfigError &&
      error.failure_reason === "usage_error" &&
      /VITEST_MAX_WORKERS must be a positive integer <= 16/.test(error.message),
  );
  assert.throws(
    () =>
      preflightPublicTarget("frontend-unit", {
        VITEST_FLAGS: "apps/web/src/example.test.tsx",
        CARTULARY_MAKE_ORIGIN_VITEST_FLAGS: "command line",
      }),
    (error) =>
      error instanceof HarnessConfigError &&
      error.failure_reason === "usage_error" &&
      /VITEST_FLAGS is not declared for target frontend-unit/.test(error.message),
  );
  assert.doesNotThrow(() =>
    preflightPublicTarget("frontend-unit", {
      VITEST_FLAGS: "apps/web/src/example.test.tsx",
      CARTULARY_MAKE_ORIGIN_VITEST_FLAGS: "environment",
    }),
  );
  assert.throws(
    () =>
      preflightPublicTarget("db-reset", {
        CARTULARY_DESTRUCTIVE_CONFIRM: "object-store-reset",
        CARTULARY_MAKE_ORIGIN_CARTULARY_DESTRUCTIVE_CONFIRM: "command line",
      }),
    (error) =>
      error instanceof HarnessConfigError &&
      error.failure_reason === "usage_error" &&
      /CARTULARY_DESTRUCTIVE_CONFIRM must be one of db-reset/.test(error.message),
  );
  assert.doesNotThrow(() =>
    preflightPublicTarget("db-reset", {
      CARTULARY_DESTRUCTIVE_CONFIRM: "db-reset",
      CARTULARY_MAKE_ORIGIN_CARTULARY_DESTRUCTIVE_CONFIRM: "command line",
    }),
  );
  assert.doesNotThrow(() =>
    preflightPublicTarget("db-reset", {
      CARTULARY_DESTRUCTIVE_CONFIRM: "db-reset",
      CARTULARY_MAKE_ORIGIN_CARTULARY_DESTRUCTIVE_CONFIRM: "environment",
    }),
  );
});

test("extended harness contracts are explicit and outside default local check", () => {
  const { checkSchedule, taskSurface } = renderedArtifacts();
  const check = checkSchedule.schedules.find(
    (schedule) => schedule.target === "check",
  );
  assert.ok(check, "rendered check schedule must include check");
  const harnessContracts = check.work_units.find(
    (unit) => unit.target === "harness-contract-tests",
  );
  assert.equal(harnessContracts, undefined, "default local check must omit deep harness contract tests");
  const contractTarget = taskSurface.targets.find((target) => target.name === "harness-contract");
  assert.ok(contractTarget, "task surface must expose the explicit harness-contract target");
  assert.deepEqual(
    contractTarget.default_inclusion_sets,
    ["ci", "release-check"],
    "harness-contract must be selected by extended gates only",
  );
});

test("default check service-backed browser work uses declared session groups", () => {
  const { serviceBacked, expandedCheckSchedule } = renderedArtifacts();
  const serviceCheck = serviceBacked.schedules.find(
    (schedule) => schedule.target === "check-service-backed",
  );
  assert.ok(serviceCheck, "service-backed sources must include check-service-backed");
  const browserSources = serviceCheck.work_unit_sources.filter(
    (source) => source.type === "browser_stage",
  );
  assert.deepEqual(
    new Map(browserSources.map((source) => [source.browser_stage, source.browser_session_group])),
    new Map([
      ["webserver-backed", "default-check-browser-shared"],
      ["stateful", "default-check-stateful-isolated"],
    ]),
  );
  assert.equal(
    browserSources.find((source) => source.browser_stage === "stateful")
      ?.browser_session_isolation_reason,
    "stateful browser evidence mutates persisted runtime state and remains isolated from shared default-check browser work",
  );

  const check = expandedCheckSchedule.schedules.find(
    (schedule) => schedule.target === "check",
  );
  const browserSessions = check.work_units.filter(
    (unit) => unit.kind === "browser_stage_session",
  );
  const expectedStatefulSessionGroups = [
    "stateful-phase1",
    "stateful-phase6",
    "stateful-phase8",
  ].map((group) => `default-check-stateful-isolated-${group}`);
  assert.equal(browserSessions.length, 1 + expectedStatefulSessionGroups.length);
  assert.deepEqual(
    browserSessions.map((unit) => unit.browser_session_group).sort(),
    ["default-check-browser-shared", ...expectedStatefulSessionGroups].sort(),
  );
  for (const session of browserSessions.filter((unit) => unit.browser_stage === "stateful")) {
    assert.equal(
      session.browser_session_isolation_reason,
      "stateful browser evidence mutates persisted runtime state and remains isolated from shared default-check browser work",
    );
  }
  const webserverBrowserGroups = check.work_units.filter(
    (unit) =>
      unit.kind === "browser_group" &&
      unit.aggregate_target === "browser-e2e-webserver-backed",
  );
  const functionalBrowserGroups = webserverBrowserGroups.filter(
    (unit) => unit.browser_group?.kind === "functional_shard",
  );
  const supportBrowserGroups = webserverBrowserGroups.filter(
    (unit) => unit.browser_group?.kind === "support",
  );
  const functionalShardCounts = new Set(
    functionalBrowserGroups.map((unit) => unit.browser_group?.shard_count),
  );
  assert.deepEqual([...functionalShardCounts], [functionalBrowserGroups.length]);
  assert.equal(supportBrowserGroups.length, 1);
  for (const excludedBrowserTarget of [
    "browser-e2e-measurement",
    "browser-e2e-visual",
    "browser-e2e-a11y",
  ]) {
    assert.equal(
      check.work_units.some(
        (unit) =>
          unit.aggregate_target === excludedBrowserTarget ||
          unit.target === excludedBrowserTarget,
      ),
      false,
      `${excludedBrowserTarget} must remain outside default local check`,
    );
  }
  assert.equal(
    check.work_units.filter(
      (unit) =>
        unit.kind === "browser_group" &&
        unit.aggregate_target === "browser-e2e-stateful",
    ).length,
    expectedStatefulSessionGroups.length,
  );
  const statefulBrowserGroups = check.work_units.filter(
    (unit) =>
      unit.kind === "browser_group" &&
      unit.aggregate_target === "browser-e2e-stateful",
  );
  assert.deepEqual(
    statefulBrowserGroups
      .map((unit) => [unit.browser_group?.name, unit.browser_group?.selected_row_ids ?? []])
      .sort(([left], [right]) => left.localeCompare(right))
      .map(([, rowIDs]) => rowIDs),
    [["E-1-04", "E-1-05"], ["E-6-03"], ["E-8-05"]],
  );
  for (const unit of statefulBrowserGroups) {
    assert.equal(unit.env.CARTULARY_FRONTEND_ROW_ACCOUNTING_SCOPE, "disabled");
    assert.equal(unit.env.CARTULARY_FRONTEND_ROW_ACCOUNTING_PHASE_NAMESPACE, "base");
  }
  const defaultFunctionalEntryIDs = functionalBrowserGroups.flatMap(
    (unit) => unit.browser_group?.entry_ids ?? [],
  );
  for (const explicitRowID of [
    "FE-E-P4-01",
    "FE-E-P5-01",
    "FE-E-P6-01",
    "FE-E-P7-01",
    "FE-E-P8-01",
    "FE-E-P9-01",
    "FE-E-P9-02",
    "FE-E-P9-03",
    "FE-E-P10-01",
  ]) {
    assert.equal(
      defaultFunctionalEntryIDs.includes(explicitRowID),
      false,
      `${explicitRowID} must stay outside default check browser projection`,
    );
  }
  assert.deepEqual(
    [...new Set(defaultFunctionalEntryIDs.filter((id) => id.startsWith("FE-")))].sort(),
    ["FE-B-P2-01", "FE-B-P2-02", "FE-E-P1-01", "FE-E-P2-01"].sort(),
  );
  assertBrowserWorkerSlots(
    check.work_units.filter(
      (unit) =>
        unit.kind === "browser_group" &&
        unit.service_session?.target === "check-service-backed",
    ),
    "default check service-backed browser groups",
  );
  const serviceBackedCheckSource = serviceBacked.schedules.find(
    (schedule) => schedule.target === "check-service-backed",
  );
  assert.ok(
    serviceBackedCheckSource,
    "rendered service-backed artifact must include check-service-backed",
  );
  const serviceBackedCheckUnits = expandServiceBackedSchedule({
    repoRoot,
    serviceSchedule: serviceBackedCheckSource,
  });
  assertBrowserWorkerSlots(
    serviceBackedCheckUnits.filter((unit) => unit.kind === "browser_group"),
    "direct check-service-backed browser groups",
  );
});

test("harness import boundary has no forbidden planning edges", () => {
  const report = collectHarnessImportBoundaryViolations(repoRoot);
  assert.deepEqual(report.violations, []);
  assert.deepEqual(report.forbidden_sccs, []);
});

test("backend module boundary rejects Revisions source SQL, mappings, and provider imports", () => {
  const root = mkdtempSync(path.join(repoRoot, "tmp", "backend-boundary."));
  try {
    writeFixtureFile(
      root,
      "internal/modules/revisions/source_sql.go",
      "package revisions\n\nconst sourceSQL = `UPDATE hosts SET display_name = $1`\n",
    );
    writeFixtureFile(
      root,
      "internal/modules/revisions/source_mapping.go",
      'package revisions\n\nvar sourceMapping = map[string]string{"host.display_name": "display_name"}\n',
    );
    writeFixtureFile(
      root,
      "internal/modules/revisions/provider_import.go",
      'package revisions\n\nimport _ "github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity/rollbackprovider"\n',
    );

    const result = spawnSync(
      process.execPath,
      [
        path.join(repoRoot, "tools/harness/static-analysis/backend-module-boundary-check-cli.mjs"),
        "--manifest",
        path.join(repoRoot, "tools/backend_module_boundaries.json"),
        "--root",
        root,
      ],
      { cwd: repoRoot, encoding: "utf8" },
    );
    assert.notEqual(result.status, 0, `synthetic violations unexpectedly passed: ${result.stdout}`);
    const report = JSON.parse(result.stdout.trim());
    assert.ok(
      report.violations.some(
        (violation) =>
          violation.code === "source_table_access" &&
          violation.path === "internal/modules/revisions/source_sql.go" &&
          violation.symbol_or_import === "hosts",
      ),
      `source SQL violation missing: ${result.stdout}`,
    );
    assert.ok(
      report.violations.some(
        (violation) =>
          violation.code === "sql_table_allowlist" &&
          violation.path === "internal/modules/revisions/source_sql.go" &&
          violation.symbol_or_import === "hosts",
      ),
      `SQL allowlist violation missing: ${result.stdout}`,
    );
    assert.ok(
      report.violations.some(
        (violation) =>
          violation.code === "source_mapping" &&
          violation.path === "internal/modules/revisions/source_mapping.go" &&
          violation.symbol_or_import === "host.",
      ),
      `source mapping violation missing: ${result.stdout}`,
    );
    assert.ok(
      report.violations.some(
        (violation) =>
          violation.code === "owner_port_only_import" &&
          violation.path === "internal/modules/revisions/provider_import.go" &&
          violation.symbol_or_import.endsWith("/entities/hostidentity/rollbackprovider"),
      ),
      `provider import violation missing: ${result.stdout}`,
    );
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("backend module boundary consumes test support inventory scan exclusions", () => {
  const root = mkdtempSync(path.join(repoRoot, "tmp", "backend-boundary-support."));
  try {
    writeFixtureFile(
      root,
      "internal/modules/future/production.go",
      'package future\n\nimport _ "github.com/JochiRaider/cartulary/internal/modules/workbook/routes"\n',
    );
    writeFixtureFile(
      root,
      "internal/modules/future/testsupport/routetest/support.go",
      'package routetest\n\nimport _ "github.com/JochiRaider/cartulary/internal/modules/workbook/routes"\n',
    );
    writeFixtureFile(
      root,
      "tools/test_support_inventory.json",
      JSON.stringify(
        {
          schema_id: "cartulary.test_support_inventory.v1",
          go_support_roots: [
            {
              path: "internal/modules/future/testsupport",
              owner: "future",
              posture: "owner_local",
              runtime_scan: "excluded",
              support_scan: "included",
              service_starting: false,
              rationale: "Synthetic support root excluded from production boundary scans.",
            },
          ],
          shared_data_roots: [],
        },
        null,
        2,
      ),
    );

    const result = spawnSync(
      process.execPath,
      [
        path.join(repoRoot, "tools/harness/static-analysis/backend-module-boundary-check-cli.mjs"),
        "--manifest",
        path.join(repoRoot, "tools/backend_module_boundaries.json"),
        "--support-inventory",
        path.join(root, "tools/test_support_inventory.json"),
        "--root",
        root,
      ],
      { cwd: repoRoot, encoding: "utf8" },
    );
    assert.notEqual(result.status, 0, `synthetic production violation unexpectedly passed: ${result.stdout}`);
    const report = JSON.parse(result.stdout.trim());
    assert.ok(
      report.effective_scan_excludes.includes("internal/modules/future/testsupport/**"),
      `support inventory exclusion missing: ${result.stdout}`,
    );
    assert.deepEqual(
      report.violations.map((violation) => violation.path),
      ["internal/modules/future/production.go"],
      `support-root violation should be excluded: ${result.stdout}`,
    );
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("harness import boundary consumes the authored helper ownership registry", () => {
  const report = collectHarnessImportBoundaryViolations(repoRoot);
  const manifestText = [
    "tools/task_surface_manifest.json",
    "tools/execution_topology_manifest.json",
  ]
    .map((relativePath) => readFileSync(path.join(repoRoot, relativePath), "utf8"))
    .join("\n");
  assert.deepEqual(
    Object.keys(report.owner_facades).sort(),
    Object.keys(ownerFacadePathLists).sort(),
  );
  for (const [owner, paths] of Object.entries(ownerFacadePathLists)) {
    assert.deepEqual(
      report.owner_facades[owner],
      [...paths].sort((left, right) => left.localeCompare(right)),
      `${owner} facade paths must come from the helper ownership registry`,
    );
    for (const ownerPath of paths) {
      assert.ok(
        existsSync(path.join(repoRoot, ownerPath)),
        `${owner} facade path must exist: ${ownerPath}`,
      );
    }
  }
  assert.deepEqual(
    report.unsupported_private_rules,
    unsupportedPrivateHelperRules.map((rule) => ({
      id: rule.id,
      exact: [...rule.exact].sort((left, right) => left.localeCompare(right)),
      prefixes: [...rule.prefixes].sort((left, right) => left.localeCompare(right)),
    })),
  );
  for (const rule of unsupportedPrivateHelperRules) {
    for (const ownerPath of rule.exact) {
      assert.equal(
        manifestText.includes(ownerPath),
        false,
        `${ownerPath} must not appear in public task/topology manifests`,
      );
    }
    for (const prefix of rule.prefixes) {
      assert.equal(
        manifestText.includes(prefix),
        false,
        `${prefix} must not appear in public task/topology manifests`,
      );
    }
  }
});

test("harness import boundary rejects legacy planning imports and cycles", () => {
  const root = mkdtempSync(path.join(repoRoot, "tmp", "harness-boundary."));
  try {
    writeFixtureFile(
      root,
      "tools/harness/backend/backend-shard-plan.mjs",
      `${fixtureImport("../backend/go-shard-plan.mjs")}export const backendShardPlan = true;\n`,
    );
    writeFixtureFile(
      root,
      "tools/harness/backend/backend-target-plan.mjs",
      "export const backendTargetPlan = true;\n",
    );
    writeFixtureFile(
      root,
      "tools/harness/backend/backend-duration-accounting.mjs",
      "export const backendDurationAccounting = true;\n",
    );
    writeFixtureFile(
      root,
      "tools/harness/backend/target-execution/cli.mjs",
      "export const backendTargetExecutionCli = true;\n",
    );
    writeFixtureFile(
      root,
      "tools/harness/backend/backend-target-execution.mjs",
      fixtureExportFrom("backendTargetExecutionCli", "./target-execution/cli.mjs"),
    );
    writeFixtureFile(
      root,
      "tools/harness/phase-accounting/phase-manifest.mjs",
      "export const phaseManifest = true;\n",
    );
    writeFixtureFile(
      root,
      "tools/harness/phase-accounting/phase-registry.mjs",
      "export const phaseRegistry = true;\n",
    );
    writeFixtureFile(
      root,
      "tools/harness/phase-accounting/frontend/phase-artifacts.mjs",
      "export const frontendPhaseValidation = true;\n",
    );
    writeFixtureFile(
      root,
      "tools/harness/phase-accounting/frontend-phase-manifest.mjs",
      fixtureExportFrom("frontendPhaseValidation", "./frontend/phase-artifacts.mjs"),
    );
    writeFixtureFile(
      root,
      "tools/harness/phase-accounting/frontend-row-accounting.mjs",
      "export const frontendRowAccounting = true;\n",
    );
    writeFixtureFile(
      root,
      "tools/harness/phase-accounting/frontend-readiness.mjs",
      "export const frontendReadiness = true;\n",
    );
    writeFixtureFile(
      root,
      "tools/harness/execution/summary-topology.mjs",
      "export const summaryTopology = true;\n",
    );
    writeFixtureFile(
      root,
      "tools/harness/backend/target-plan.mjs",
      "export const targetPlan = true;\n",
    );
    writeFixtureFile(
      root,
      "tools/harness/backend/go-shard-plan.mjs",
      `export async function inspect() { return ${fixtureDynamicImport("./target-plan.mjs")}; }\n`,
    );
    writeFixtureFile(
      root,
      "tools/harness/backend/go-target-runner.mjs",
      `${fixtureImport("./backend-target-plan.mjs")}export const runner = true;\n`,
    );
    writeFixtureFile(
      root,
      "tools/harness/browser/browser-shard-plan.mjs",
      `export async function inspect() { return ${fixtureDynamicImport("../phase-accounting/phase-manifest.mjs")}; }\n`,
    );
    writeFixtureFile(
      root,
      "tools/harness/browser/browser-batch-manifest.mjs",
      "export const browserBatchManifest = true;\n",
    );
    writeFixtureFile(
      root,
      "tools/harness/browser/browser-duration-accounting.mjs",
      "export const browserDurationAccounting = true;\n",
    );
    writeFixtureFile(
      root,
      "tools/harness/browser/browser-duration-discovery.mjs",
      "export const privateBrowserDurationDiscovery = true;\n",
    );
    writeFixtureFile(
      root,
      "tools/harness/scheduler/adapters/browser.mjs",
      [
        fixtureExportFrom("browserBatchManifest", "../../browser/browser-batch-manifest.mjs").trimEnd(),
        fixtureExportFrom(
          "privateBrowserDurationDiscovery",
          "../../browser/browser-duration-discovery.mjs",
        ).trimEnd(),
        "",
      ].join("\n"),
    );
    writeFixtureFile(
      root,
      "tools/harness/phase-accounting/phase-slice-plan.mjs",
      [
        fixtureExportFrom("browserSchedulerAdapter", "../scheduler/adapters/browser.mjs").trimEnd(),
        fixtureExportFrom("backendShardPlan", "../backend/backend-shard-plan.mjs").trimEnd(),
        "",
      ].join("\n"),
    );
    writeFixtureFile(
      root,
      "tools/harness/execution/service-backed/schedule-planning.mjs",
      [
        fixtureExportFrom("backendShardPlan", "../../backend/backend-shard-plan.mjs").trimEnd(),
        fixtureExportFrom("browserSchedulerAdapter", "../../scheduler/adapters/browser.mjs").trimEnd(),
        "",
      ].join("\n"),
    );
    writeFixtureFile(
      root,
      "tools/harness/duration-accounting/index.mjs",
      "export const durationAccounting = true;\n",
    );
    writeFixtureFile(
      root,
      "tools/harness/scheduler/scheduler/event-order.mjs",
      "export const schedulerEventOrder = true;\n",
    );
    writeFixtureFile(
      root,
      "tools/harness/scheduler/phase-slice-execution.mjs",
      "export const phaseSliceExecution = true;\n",
    );
    writeFixtureFile(
      root,
      "tools/harness/output/test-output/frontend-row-evidence.mjs",
      "export const frontendRowEvidence = true;\n",
    );
    writeFixtureFile(
      root,
      "tools/harness/generated-artifacts/duration-facade.mjs",
      fixtureExportFrom(
        "backendDurationAccounting",
        "../backend/backend-duration-accounting.mjs",
      ),
    );
    writeFixtureFile(
      root,
      "tools/harness/generated-artifacts/browser-duration-facade.mjs",
      fixtureExportFrom(
        "browserDurationAccounting",
        "../browser/browser-duration-accounting.mjs",
      ),
    );
    writeFixtureFile(
      root,
      "tools/harness/diagnostics/execution-facade.mjs",
      fixtureExportFrom(
        "backendTargetExecutionCli",
        "../backend/backend-target-execution.mjs",
      ),
    );

    const clean = collectHarnessImportBoundaryViolations(root);
    assert.deepEqual(clean.violations, []);
    assert.deepEqual(clean.forbidden_sccs, []);
    assert.ok(
      clean.owner_facades.scheduler.includes("tools/harness/scheduler/scheduler-runner.mjs"),
      "scheduler runner facade must be classified",
    );
    assert.ok(
      clean.owner_facades.scheduler.includes(
        "tools/harness/scheduler/scheduler-family-contract.mjs",
      ),
      "scheduler family facade must be classified",
    );
    assert.ok(
      clean.owner_facades.scheduler.includes("tools/harness/scheduler/scheduler-manifest.mjs"),
      "scheduler manifest facade must be classified",
    );
    assert.ok(
      clean.owner_facades.scheduler.includes("tools/harness/scheduler/scheduler-resources.mjs"),
      "scheduler resources facade must be classified",
    );
    assert.ok(
      clean.owner_facades.scheduler.includes(
        "tools/harness/scheduler/scheduler-resource-policy.mjs",
      ),
      "scheduler resource policy facade must be classified",
    );
    assert.ok(
      clean.owner_facades.scheduler.includes("tools/harness/scheduler/scheduler-reporting.mjs"),
      "scheduler reporting facade must be classified",
    );
    assert.ok(
      clean.owner_facades.scheduler.includes("tools/harness/scheduler/process-executor.mjs"),
      "scheduler process adapter facade must be classified",
    );
    assert.ok(
      clean.owner_facades.scheduler.includes("tools/harness/scheduler/scheduler-runtime.mjs"),
      "scheduler runtime facade must be classified",
    );
    assert.ok(
      clean.owner_facades.scheduler.includes(
        "tools/harness/scheduler/phase-slice-execution.mjs",
      ),
      "phase-slice scheduler execution facade must be classified",
    );
    assert.ok(
      clean.owner_facades.test_output.includes(
        "tools/harness/output/test-output/frontend-row-evidence.mjs",
      ),
      "frontend row evidence test-output facade must be classified",
    );
    assert.ok(
      clean.owner_facades.test_output.includes(
        "tools/harness/output/test-output/frontend-indexes.mjs",
      ),
      "frontend manifest test-output index facade must be classified",
    );
    assert.ok(
      clean.owner_facades.browser.includes(
        "tools/harness/browser/browser-lifecycle-adapter.sh",
      ),
      "browser lifecycle adapter facade must be classified",
    );
    assert.ok(
      clean.owner_facades.phase_accounting.includes(
        "tools/harness/phase-accounting/index.mjs",
      ),
      "phase-accounting index facade must be classified",
    );
    assert.ok(
      clean.owner_facades.phase_accounting.includes(
        "tools/harness/phase-accounting/phase-slice-plan.mjs",
      ),
      "phase-slice planning facade must be classified",
    );
    for (const phaseAccountingFacade of [
      "tools/harness/phase-accounting/phase-manifest.mjs",
      "tools/harness/phase-accounting/phase-registry.mjs",
      "tools/harness/phase-accounting/frontend-phase-manifest.mjs",
      "tools/harness/phase-accounting/frontend-row-accounting.mjs",
      "tools/harness/phase-accounting/frontend-readiness.mjs",
    ]) {
      assert.ok(
        clean.owner_facades.phase_accounting.includes(phaseAccountingFacade),
        `${phaseAccountingFacade} must be classified as a phase-accounting facade`,
      );
    }
    assert.ok(
      clean.owner_facades.service_backed_execution.includes(
        "tools/harness/execution/service-backed/index.mjs",
      ),
      "service-backed schedule planning index facade must be classified",
    );
    assert.ok(
      clean.owner_facades.service_backed_execution.includes(
        "tools/harness/execution/service-backed/schedule-planning.mjs",
      ),
      "service-backed schedule planning facade must be classified",
    );
    assert.ok(
      clean.owner_facades.execution_runtime.includes(
        "tools/harness/execution/phase-runtime.sh",
      ),
      "phase execution runtime facade must be classified",
    );
    assert.ok(
      clean.owner_facades.command_surface.includes(
        "tools/harness/command-surface/make-node-tools.mjs",
      ),
      "Make-node command-surface facade must be classified",
    );
    assert.ok(
      clean.owner_facades.generated_artifacts.includes(
        "tools/harness/generated-artifacts/index.mjs",
      ),
      "generated-artifacts facade must be classified",
    );
    assert.ok(
      clean.owner_facades.readiness.includes(
        "tools/harness/readiness/cache-policy.sh",
      ),
      "readiness cache policy facade must be classified",
    );
    assert.ok(
      clean.owner_facades.duration_accounting.includes("tools/harness/duration-accounting/index.mjs"),
      "duration accounting facade must be classified",
    );
    assert.ok(
      clean.owner_facades.contract.includes("tools/harness/contract/index.mjs"),
      "contract index facade must be classified",
    );
    assert.ok(
      clean.owner_facades.output.includes("tools/harness/output/index.mjs"),
      "output index facade must be classified",
    );
    assert.ok(
      clean.owner_facades.scheduler_diagnostics.includes(
        "tools/harness/scheduler/scheduler/event-order.mjs",
      ),
      "scheduler event drift facade must be classified",
    );
    assert.ok(
      clean.owner_facades.scheduler_diagnostics.includes(
        "tools/harness/scheduler/scheduler/summary-timing-drift.mjs",
      ),
      "scheduler summary timing drift facade must be classified",
    );
    assert.ok(
      !clean.owner_facades.scheduler.includes(
        "tools/harness/phase-accounting/phase-slice-plan.mjs",
      ),
      "phase-slice planning facade must not remain in scheduler bucket",
    );
    assert.ok(
      !clean.owner_facades.scheduler.includes(
        "tools/harness/execution/service-backed/schedule-planning.mjs",
      ),
      "service-backed planning facade must not remain in scheduler bucket",
    );
    assert.ok(
      !clean.owner_facades.scheduler.includes(
        "tools/harness/scheduler/scheduler/event-order.mjs",
      ),
      "scheduler event drift facade must not remain in scheduler bucket",
    );
    assert.ok(
      !clean.owner_facades.scheduler.includes(
        "tools/harness/scheduler/scheduler/summary-timing-drift.mjs",
      ),
      "scheduler summary timing drift facade must not remain in scheduler bucket",
    );
    for (const unsupportedRule of [
      "legacy_backend_database_contract_drift",
      "legacy_backend_duration_and_shard_helpers",
      "legacy_backend_security_findings_helper",
      "legacy_frontend_catch_all_directory",
      "legacy_scheduler_backend_adapters",
      "legacy_scheduler_phase_slice_and_service_backed_helpers",
      "legacy_scheduler_duration_helpers",
      "legacy_scheduler_process_and_evidence_drift_helpers",
      "legacy_execution_phase_runtime_and_node_registry",
    ]) {
      assert.ok(
        clean.unsupported_private_rules.some((rule) => rule.id === unsupportedRule),
        `${unsupportedRule} unsupported-private rule must be reported`,
      );
    }
    for (const legacySchedulerHelper of [
      "tools/harness/scheduler/adapters/backend.mjs",
      "tools/harness/scheduler/adapters/schedule-context.mjs",
      "tools/harness/scheduler/check-service-backed-expansion.mjs",
      "tools/harness/scheduler/service-backed-schedule-manifest.mjs",
      "tools/harness/scheduler/service-backed-schedule-topology.mjs",
      "tools/harness/scheduler/phase-slice-plan.mjs",
      "tools/harness/scheduler/phase-slice-cli.mjs",
      "tools/harness/scheduler/execution-dependencies.mjs",
      "tools/harness/scheduler/scheduler/process-executor.mjs",
      "tools/harness/scheduler/duration-baseline-cli.mjs",
      "tools/harness/scheduler/duration-drift.mjs",
      "tools/harness/scheduler/target-duration-baselines.mjs",
      "tools/harness/scheduler/service-backed-make-target-durations-cli.mjs",
      "tools/harness/scheduler/harness-smoke-durations-cli.mjs",
      "tools/harness/scheduler/duration-baseline-drift-suite.sh",
      "tools/harness/scheduler/scheduler-event-order-drift-cli.mjs",
      "tools/harness/scheduler/scheduler-summary-timing-drift-cli.mjs",
      "tools/harness/execution/run-phase-common.sh",
      "tools/harness/execution/make-node-tools.mjs",
    ]) {
      assert.ok(
        clean.unsupported_private_helpers.includes(legacySchedulerHelper),
        `${legacySchedulerHelper} must remain unsupported_private rather than a stable shim`,
      );
    }

    writeFixtureFile(
      root,
      "tools/harness/backend/direct-target-plan.mjs",
      `${fixtureImport(legacyPlanningImport("target-plan.mjs"))}export const directTargetPlan = true;\n`,
    );
    writeFixtureFile(
      root,
      "tools/harness/browser/direct-phase-manifest.mjs",
      `${fixtureImport(legacyPlanningImport("phase-manifest.mjs"))}export const directPhaseManifest = true;\n`,
    );
    const direct = collectHarnessImportBoundaryViolations(root);
    const directSources = new Set(
      direct.violations
        .filter((violation) => violation.rule === "forbidden_planning_import")
        .map((violation) => violation.source),
    );
    assert.ok(directSources.has("tools/harness/backend/direct-target-plan.mjs"));
    assert.ok(directSources.has("tools/harness/browser/direct-phase-manifest.mjs"));

    writeFixtureFile(
      root,
      "tools/harness/scheduler/direct-backend-target-plan.mjs",
      `${fixtureImport("../backend/target-plan.mjs")}export const directBackendTargetPlan = true;\n`,
    );
    writeFixtureFile(
      root,
      "tools/harness/generated-artifacts/direct-legacy-duration.mjs",
      `${fixtureImport("../backend/duration/baselines.mjs")}export const directLegacyDuration = true;\n`,
    );
    writeFixtureFile(
      root,
      "tools/harness/generated-artifacts/direct-legacy-scheduler-phase-slice.mjs",
      `${fixtureImport("../scheduler/phase-slice-plan.mjs")}export const directLegacySchedulerPhaseSlice = true;\n`,
    );
    writeFixtureFile(
      root,
      "tools/harness/generated-artifacts/direct-legacy-execution-runtime.mjs",
      `${fixtureImport("../execution/run-phase-common.sh")}export const directLegacyExecutionRuntime = true;\n`,
    );
    writeFixtureFile(
      root,
      "tools/harness/generated-artifacts/direct-legacy-execution-runtime-source.sh",
      '#!/usr/bin/env bash\nsource "${ROOT_DIR}/tools/harness/execution/run-phase-common.sh"\n',
    );
    writeFixtureFile(
      root,
      "tools/harness/generated-artifacts/direct-legacy-make-node-tools.mjs",
      `${fixtureImport("../execution/make-node-tools.mjs")}export const directLegacyMakeNodeTools = true;\n`,
    );
    writeFixtureFile(
      root,
      "tools/harness/diagnostics/direct-go-target-runner.mjs",
      `${fixtureImport("../backend/go-target-runner.mjs")}export const directGoTargetRunner = true;\n`,
    );
    writeFixtureFile(
      root,
      "tools/harness/scheduler/direct-target-execution-helper.mjs",
      `${fixtureImport("../backend/target-execution/cli.mjs")}export const directTargetExecutionHelper = true;\n`,
    );
    writeFixtureFile(
      root,
      "tools/harness/scheduler/direct-browser-batch.mjs",
      `${fixtureImport("../browser/browser-batch-manifest.mjs")}export const directBrowserBatch = true;\n`,
    );
    writeFixtureFile(
      root,
      "tools/harness/generated-artifacts/direct-browser-duration-discovery.mjs",
      `${fixtureImport("../browser/browser-duration-discovery.mjs")}export const directBrowserDurationDiscovery = true;\n`,
    );
    writeFixtureFile(
      root,
      "tools/harness/diagnostics/direct-frontend-evidence.mjs",
      `${fixtureImport("../frontend/evidence/index.mjs")}export const directFrontendEvidence = true;\n`,
    );
    writeFixtureFile(
      root,
      "tools/harness/diagnostics/direct-frontend-phase-validation.mjs",
      `${fixtureImport("../phase-accounting/frontend/phase-artifacts.mjs")}export const directFrontendPhaseValidation = true;\n`,
    );
    const backendBoundary = collectHarnessImportBoundaryViolations(root);
    assert.ok(
      backendBoundary.violations.some(
        (violation) =>
          violation.rule === "forbidden_private_backend_import" &&
          violation.source === "tools/harness/scheduler/direct-backend-target-plan.mjs" &&
          violation.target === "tools/harness/backend/target-plan.mjs",
      ),
      "non-owner target-plan import must be reported",
    );
    assert.ok(
      backendBoundary.violations.some(
        (violation) =>
          violation.rule === "forbidden_unsupported_private_helper_import" &&
          violation.unsupported_private_rule ===
            "legacy_backend_duration_and_shard_helpers" &&
          violation.source === "tools/harness/generated-artifacts/direct-legacy-duration.mjs" &&
          violation.target === "tools/harness/backend/duration/baselines.mjs",
      ),
      "unsupported helper import must be reported",
    );
    assert.ok(
      backendBoundary.violations.some(
        (violation) =>
          violation.rule === "forbidden_unsupported_private_helper_import" &&
          violation.unsupported_private_rule ===
            "legacy_scheduler_phase_slice_and_service_backed_helpers" &&
          violation.source ===
            "tools/harness/generated-artifacts/direct-legacy-scheduler-phase-slice.mjs" &&
          violation.target === "tools/harness/scheduler/phase-slice-plan.mjs",
      ),
      "unsupported scheduler catch-all helper import must be reported",
    );
    assert.ok(
      backendBoundary.violations.some(
        (violation) =>
          violation.rule === "forbidden_unsupported_private_helper_import" &&
          violation.unsupported_private_rule ===
            "legacy_execution_phase_runtime_and_node_registry" &&
          violation.source ===
            "tools/harness/generated-artifacts/direct-legacy-execution-runtime.mjs" &&
          violation.target === "tools/harness/execution/run-phase-common.sh",
      ),
      "unsupported legacy phase runtime helper import must be reported",
    );
    assert.ok(
      backendBoundary.violations.some(
        (violation) =>
          violation.rule === "forbidden_unsupported_private_helper_import" &&
          violation.unsupported_private_rule ===
            "legacy_execution_phase_runtime_and_node_registry" &&
          violation.source ===
            "tools/harness/generated-artifacts/direct-legacy-execution-runtime-source.sh" &&
          violation.target === "tools/harness/execution/run-phase-common.sh",
      ),
      "unsupported legacy phase runtime shell source must be reported",
    );
    assert.ok(
      backendBoundary.violations.some(
        (violation) =>
          violation.rule === "forbidden_unsupported_private_helper_import" &&
          violation.unsupported_private_rule ===
            "legacy_execution_phase_runtime_and_node_registry" &&
          violation.source ===
            "tools/harness/generated-artifacts/direct-legacy-make-node-tools.mjs" &&
          violation.target === "tools/harness/execution/make-node-tools.mjs",
      ),
      "unsupported legacy Make-node registry import must be reported",
    );
    assert.ok(
      backendBoundary.violations.some(
        (violation) =>
          violation.rule === "forbidden_private_backend_import" &&
          violation.source === "tools/harness/diagnostics/direct-go-target-runner.mjs" &&
          violation.target === "tools/harness/backend/go-target-runner.mjs",
      ),
      "non-owner runner import must be reported",
    );
    assert.ok(
      backendBoundary.violations.some(
        (violation) =>
          violation.rule === "forbidden_private_backend_import" &&
          violation.source === "tools/harness/scheduler/direct-target-execution-helper.mjs" &&
          violation.target === "tools/harness/backend/target-execution/cli.mjs",
      ),
      "non-owner target-execution helper import must be reported",
    );
    assert.ok(
      backendBoundary.violations.some(
        (violation) =>
          violation.rule === "forbidden_scheduler_private_browser_import" &&
          violation.source === "tools/harness/scheduler/direct-browser-batch.mjs" &&
          violation.target === "tools/harness/browser/browser-batch-manifest.mjs",
      ),
      "scheduler must use browser adapter rather than direct browser helper imports",
    );
    assert.ok(
      backendBoundary.violations.some(
        (violation) =>
          violation.rule === "forbidden_private_browser_import" &&
          violation.source ===
            "tools/harness/generated-artifacts/direct-browser-duration-discovery.mjs" &&
          violation.target === "tools/harness/browser/browser-duration-discovery.mjs",
      ),
      "non-owner browser duration discovery import must be reported",
    );
    assert.ok(
      backendBoundary.violations.some(
        (violation) =>
          violation.rule === "forbidden_unsupported_private_helper_import" &&
          violation.unsupported_private_rule === "legacy_frontend_catch_all_directory" &&
          violation.source === "tools/harness/diagnostics/direct-frontend-evidence.mjs" &&
          violation.target === "tools/harness/frontend/evidence/index.mjs",
      ),
      "unsupported frontend catch-all helper import must be reported",
    );
    assert.ok(
      backendBoundary.violations.some(
        (violation) =>
          violation.rule === "forbidden_private_phase_accounting_import" &&
          violation.source ===
            "tools/harness/diagnostics/direct-frontend-phase-validation.mjs" &&
          violation.target === "tools/harness/phase-accounting/frontend/phase-artifacts.mjs",
      ),
      "non-owner phase-accounting private import must be reported",
    );

    writeFixtureFile(
      root,
      legacyPlanningPath("cycle.mjs"),
      `${fixtureImport("../backend/cycle.mjs")}export const cyclePlanning = true;\n`,
    );
    writeFixtureFile(
      root,
      "tools/harness/backend/cycle.mjs",
      `${fixtureImport(legacyPlanningImport("cycle.mjs"))}export const cycleBackend = true;\n`,
    );
    const cyclic = collectHarnessImportBoundaryViolations(root);
    assert.ok(
      cyclic.forbidden_sccs.some(
        (scc) =>
          scc.files.includes("tools/harness/backend/cycle.mjs") &&
          scc.files.includes(legacyPlanningPath("cycle.mjs")),
      ),
      "forbidden backend/planning cycle must be reported",
    );
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("harness import boundary rejects private core imports", () => {
  const root = mkdtempSync(path.join(repoRoot, "tmp", "harness-core-boundary."));
  const privateCoreImportPath = "../core/public-contract.mjs";
  const privateCoreTarget = ["tools", "harness", "core", "public-contract.mjs"].join("/");
  try {
    writeFixtureFile(
      root,
      "tools/harness/contract/index.mjs",
      "export const contractFacade = true;\n",
    );
    writeFixtureFile(
      root,
      "tools/harness/output/index.mjs",
      "export const outputFacade = true;\n",
    );
    writeFixtureFile(
      root,
      "tools/harness/backend/uses-facade.mjs",
      `${fixtureImport("../contract/index.mjs")}export const backend = true;\n`,
    );
    writeFixtureFile(
      root,
      "tools/harness/browser/uses-output.mjs",
      `${fixtureImport("../output/index.mjs")}export const browser = true;\n`,
    );

    const clean = collectHarnessImportBoundaryViolations(root);
    assert.deepEqual(clean.violations, []);

    writeFixtureFile(
      root,
      "tools/harness/backend/uses-core.mjs",
      `${fixtureImport(privateCoreImportPath)}export const backendCore = true;\n`,
    );
    const report = collectHarnessImportBoundaryViolations(root);
    assert.ok(
      report.violations.some(
        (violation) =>
          violation.rule === "forbidden_private_core_import" &&
          violation.source === "tools/harness/backend/uses-core.mjs" &&
          violation.target === privateCoreTarget,
      ),
      "direct private core import must be reported",
    );
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("contract output and scheduler summaries preserve normalized public surface", () => {
  const runRoot = ".cartulary/test-results/contract-surface";
  const summary = buildToolRunSummary({
    target: "harness-contract",
    command: ["make", "harness-contract"],
    status: "fail",
    startedAt: "2026-01-01T00:00:00.000Z",
    completedAt: "2026-01-01T00:00:01.250Z",
    durationMs: 1250.4,
    outputMode: "machine",
    resultRoot: ".cartulary/test-results",
    runId: "contract-surface",
    runRoot,
    summaryArtifacts: [
      artifactRef(
        "target_summary",
        `${runRoot}/harness-contract/target-summary.json`,
      ),
      artifactRef(
        "tool_run_summary",
        `${runRoot}/harness-contract/tool-run-summary.json`,
      ),
      artifactRef(
        "scheduler_summary",
        `${runRoot}/harness-contract/scheduler-summary.json`,
      ),
    ],
    logArtifacts: [
      artifactRef(
        "scheduler_events",
        `${runRoot}/harness-contract/scheduler-events.jsonl`,
        "jsonl",
      ),
    ],
    workUnits: [
      { id: "unit-z", status: "pass" },
      { id: "unit-a", status: "fail" },
    ],
    failures: [
      {
        failure_class: "product",
        failure_reason: "test_assertion_failure",
        target: "backend-unit",
        lifecycle_step: "semantic_target_behavior",
        scheduler_event_sequence: 2,
        artifact: `${runRoot}/backend-unit/target-summary.json`,
      },
    ],
    slowest: [
      { id: "slow-b", duration_ms: 25 },
      { id: "slow-a", duration_ms: 25 },
    ],
  });
  assert.equal(summary.schema_id, toolRunSummarySchemaID);
  assert.equal(summary.output_mode, "machine");
  assert.equal(summary.exit_code, 10);
  assert.equal(summary.failure_class, "product");
  assert.equal(summary.failure_reason, "test_assertion_failure");
  assert.deepEqual(
    summary.summary_artifacts.map((artifact) => artifact.role),
    ["scheduler_summary", "target_summary", "tool_run_summary"],
  );
  assert.deepEqual(
    summary.work_units.map((unit) => unit.id),
    ["unit-a", "unit-z"],
  );
  assert.deepEqual(
    summary.slowest.map((entry) => entry.id),
    ["slow-a", "slow-b"],
  );
  assert.equal(machineOutput({ CARTULARY_OUTPUT_MODE: "machine" }), true);
  assert.equal(
    terminalArtifactPath(runRoot, `${runRoot}/tool-run-summary.json`),
    "tool-run-summary.json",
  );

  const schedulerLine = schedulerSummaryLine({
    target: "check",
    status: "fail",
    completed: 2,
    total: 3,
    failed: 1,
    failureClass: "product",
    failureReason: "test_assertion_failure",
    skipped: 1,
    finalizerFailures: 1,
    totalWallTimeMs: 1250,
    slowest: [{ label: "unit-a", duration_ms: 1250 }],
  });
  assert.match(schedulerLine, /^\[SUMMARY\] /u);
  assert.match(schedulerLine, /target=check/u);
  assert.match(schedulerLine, /failure_class=product/u);
  assert.match(schedulerLine, /reason=test_assertion_failure/u);
  assert.match(schedulerLine, /work_units=2\/3/u);
  assert.match(schedulerLine, /total_wall_time=1\.25s/u);
  assert.match(schedulerLine, /skipped=1/u);
  assert.match(schedulerLine, /finalizer_failures=1/u);

  const root = mkdtempSync(path.join(repoRoot, "tmp", "scheduler-events."));
  const eventFile = path.join(root, "scheduler-events.jsonl");
  try {
    writeFileSync(
      eventFile,
      [
        JSON.stringify({
          schema_id: "cartulary.scheduler_event.v6",
          target: "check",
          event: "scheduler-start",
          seq: 1,
          monotonic_ms: 0,
          emitted_at: "2026-01-01T00:00:00.000Z",
        }),
        JSON.stringify({
          schema_id: "cartulary.scheduler_event.v6",
          target: "check",
          event: "scheduler-finish",
          seq: 2,
          monotonic_ms: 5,
          emitted_at: "2026-01-01T00:00:00.005Z",
        }),
      ].join("\n") + "\n",
    );
    assert.deepEqual(validateSchedulerEventOrderFile(eventFile), []);

    writeFileSync(
      eventFile,
      [
        JSON.stringify({
          schema_id: "cartulary.scheduler_event.v6",
          target: "check",
          event: "scheduler-start",
          seq: 1,
          monotonic_ms: 10,
          emitted_at: "2026-01-01T00:00:00.010Z",
        }),
        JSON.stringify({
          schema_id: "cartulary.scheduler_event.v6",
          target: "check",
          event: "scheduler-finish",
          seq: 2,
          monotonic_ms: 5,
          emitted_at: "2026-01-01T00:00:00.020Z",
        }),
      ].join("\n") + "\n",
    );
    assert.match(
      validateSchedulerEventOrderFile(eventFile).join("\n"),
      /monotonic_ms regressed/u,
    );
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("primary public failure uses closed deterministic tie breakers", () => {
  assert.deepEqual(
    primaryPublicFailure([
      {
        failure_class: "harness",
        failure_reason: "cleanup_error",
        lifecycle_step: "cleanup_finalizers",
        artifact: "z.log",
      },
      {
        failure_class: "product",
        failure_reason: "test_assertion_failure",
        lifecycle_step: "semantic_target_behavior",
        scheduler_event_sequence: 8,
        child_registry_order: 2,
        artifact: "b.log",
      },
      {
        failure_class: "product",
        failure_reason: "test_assertion_failure",
        lifecycle_step: "semantic_target_behavior",
        scheduler_event_sequence: 7,
        child_registry_order: 3,
        artifact: "a.log",
      },
    ]),
    {
      failure_class: "product",
      failure_reason: "test_assertion_failure",
      kind: "failure",
      source: "",
      target: "",
      phase: "",
      runner: "",
      label: "",
      message: "",
      artifact: "a.log",
      lifecycle_step: "semantic_target_behavior",
      scheduler_event_sequence: 7,
      child_registry_order: 3,
    },
  );

  assert.equal(
    primaryPublicFailure([
      {
        failure_class: "harness",
        failure_reason: "scheduler_accounting_error",
        lifecycle_step: "artifact_validation",
        artifact: "b.log",
      },
      {
        failure_class: "harness",
        failure_reason: "fixture_error",
        lifecycle_step: "artifact_validation",
        artifact: "a.log",
      },
    ])?.failure_reason,
    "fixture_error",
  );

  assert.equal(
    primaryPublicFailure([
      {
        failure_class: "harness",
        failure_reason: "fixture_error",
        lifecycle_step: "artifact_validation",
        label: "late-lifecycle",
      },
      {
        failure_class: "harness",
        failure_reason: "fixture_error",
        lifecycle_step: "configuration_resolution",
        label: "early-lifecycle",
      },
    ])?.label,
    "early-lifecycle",
  );
  assert.equal(
    primaryPublicFailure([
      {
        failure_class: "harness",
        failure_reason: "fixture_error",
        lifecycle_step: "semantic_target_behavior",
        scheduler_event_sequence: 4,
        child_registry_order: 1,
        label: "later-event",
      },
      {
        failure_class: "harness",
        failure_reason: "fixture_error",
        lifecycle_step: "semantic_target_behavior",
        scheduler_event_sequence: 3,
        child_registry_order: 9,
        label: "earlier-event",
      },
    ])?.label,
    "earlier-event",
  );
  assert.equal(
    primaryPublicFailure([
      {
        failure_class: "harness",
        failure_reason: "fixture_error",
        lifecycle_step: "semantic_target_behavior",
        scheduler_event_sequence: 3,
        child_registry_order: 7,
        label: "later-child",
      },
      {
        failure_class: "harness",
        failure_reason: "fixture_error",
        lifecycle_step: "semantic_target_behavior",
        scheduler_event_sequence: 3,
        child_registry_order: 2,
        label: "earlier-child",
      },
    ])?.label,
    "earlier-child",
  );
  assert.equal(
    primaryPublicFailure([
      {
        failure_class: "harness",
        failure_reason: "fixture_error",
        lifecycle_step: "semantic_target_behavior",
        scheduler_event_sequence: 3,
        child_registry_order: 2,
        artifact: "z.log",
        label: "later-artifact",
      },
      {
        failure_class: "harness",
        failure_reason: "fixture_error",
        lifecycle_step: "semantic_target_behavior",
        scheduler_event_sequence: 3,
        child_registry_order: 2,
        artifact: "a.log",
        label: "earlier-artifact",
      },
    ])?.label,
    "earlier-artifact",
  );
  assert.equal(
    primaryPublicFailure([
      {
        failure_class: "harness",
        failure_reason: "tool_diagnostic_failure",
        lifecycle_step: "semantic_target_behavior",
        scheduler_event_sequence: 3,
        child_registry_order: 2,
        artifact: "a.log",
        label: "later-reason",
      },
      {
        failure_class: "harness",
        failure_reason: "fixture_error",
        lifecycle_step: "semantic_target_behavior",
        scheduler_event_sequence: 3,
        child_registry_order: 2,
        artifact: "a.log",
        label: "earlier-reason",
      },
    ])?.label,
    "earlier-reason",
  );
  assert.equal(
    primaryPublicFailure([
      {
        failure_class: "harness",
        failure_reason: "fixture_error",
        lifecycle_step: "semantic_target_behavior",
        scheduler_event_sequence: 3,
        child_registry_order: 2,
        artifact: "a.log",
        label: "first-input",
      },
      {
        failure_class: "harness",
        failure_reason: "fixture_error",
        lifecycle_step: "semantic_target_behavior",
        scheduler_event_sequence: 3,
        child_registry_order: 2,
        artifact: "a.log",
        label: "second-input",
      },
    ])?.label,
    "first-input",
  );
});

test("failure class normalization rejects legacy aliases for current artifacts", () => {
  assert.equal(normalizeFailureClass("product"), "product");
  assert.equal(normalizeFailureClass("infra"), "infra");
  assert.equal(normalizeFailureClass("test", "unknown"), "unknown");
  assert.equal(normalizeFailureClass("helper", "unknown"), "unknown");
  assert.equal(normalizeFailureClass("infrastructure", "unknown"), "unknown");
});

test("cleanup guard protects closed roots and permits cleanup-owned paths", () => {
  const output = [];
  const stdout = { write: (value) => output.push(value) };
  runCleanup({
    scope: "clean",
    candidates: ["go.mod", "db/migrations"],
    includeTmp: false,
    dryRun: true,
    stdout,
  });
  assert.match(output.join(""), /DRY-RUN retain go\.mod protected_root/);
  assert.match(output.join(""), /DRY-RUN retain db\/migrations protected_root/);
  assert.match(
    output.join(""),
    /DRY-RUN remove-children internal\/platform\/httpapi\/webassets\/dist registered_embedded_web_assets_preserve_keep/,
  );

  const tempRoot = mkdtempSync(path.join(repoRoot, "tmp", "cleanup-owned."));
  const owned = path.relative(repoRoot, tempRoot).replaceAll("\\", "/");
  writeFileSync(path.join(tempRoot, "artifact.txt"), "temporary");
  const beforeCleanupOutput = output.length;
  runCleanup({
    candidates: [owned, "tmp/missing-cleanup-owned-path"],
    includeTmp: false,
    dryRun: false,
    stdout,
  });
  assert.equal(existsSync(tempRoot), false, "cleanup-owned temp path must be removed");
  assert.equal(
    output.slice(beforeCleanupOutput).join(""),
    `removing ${owned}\n`,
    "missing cleanup-owned path must be a no-op",
  );
  rmSync(tempRoot, { recursive: true, force: true });
});

test("test route token generation and validation follow closed attach rules", () => {
  const generated = generateTestRouteToken();
  assert.equal(generated.length, 43);
  assert.match(generated, /^[A-Za-z0-9_-]{43}$/u);
  assert.equal(testRouteTokenValid(generated), true);
  assert.equal(testRouteTokenValid("short"), false);
  assert.equal(testRouteTokenValid("token"), false);
  assert.equal(testRouteTokenValid("a".repeat(43)), false);
  assert.equal(testRouteTokenValid(`${"a".repeat(42)}\n`), false);
});

test("redaction uses closed structured keys and raw secret families", () => {
  const structured = redactValue({
    service_sessions: [
      {
        session_target: "browser-stage-token-name",
        cleanup_status: "pass",
        setup_duration_ms: 12,
        healthy: true,
        count: 3,
        absent: null,
        lease_file: "tmp/session/lease-file.json",
        session_token: "nested-session-token",
      },
    ],
    browser_stage_sessions: [
      {
        session_target: "browser-e2e-webserver-backed",
        cleanup_status: "skipped_no_lease",
        lease_file: "tmp/browser/session.json",
      },
    ],
    X_Cartulary_Test_Route_Token: "route-secret",
    CARTULARY_S3TEST_SECRET_ACCESS_KEY: "object-store-secret",
    session_target: "not-redacted-token-substring",
  });
  assert.equal(structured.service_sessions[0].session_target, "browser-stage-token-name");
  assert.equal(structured.service_sessions[0].cleanup_status, "pass");
  assert.equal(structured.service_sessions[0].setup_duration_ms, 12);
  assert.equal(structured.service_sessions[0].healthy, true);
  assert.equal(structured.service_sessions[0].count, 3);
  assert.equal(structured.service_sessions[0].absent, null);
  assert.equal(structured.service_sessions[0].lease_file, "tmp/session/lease-file.json");
  assert.equal(structured.service_sessions[0].session_token, "[REDACTED]");
  assert.equal(structured.browser_stage_sessions[0].session_target, "browser-e2e-webserver-backed");
  assert.equal(structured.browser_stage_sessions[0].cleanup_status, "skipped_no_lease");
  assert.equal(structured.browser_stage_sessions[0].lease_file, "tmp/browser/session.json");
  assert.equal(structured.X_Cartulary_Test_Route_Token, "[REDACTED]");
  assert.equal(structured.CARTULARY_S3TEST_SECRET_ACCESS_KEY, "[REDACTED]");
  assert.equal(structured.session_target, "not-redacted-token-substring");
  assert.deepEqual(
    redactValue(["--token", "cli-secret", "--target", "backend-unit"]),
    ["--token", "[REDACTED]", "--target", "backend-unit"],
  );

  const raw = redactString([
    "postgres://cartulary:supersecret@127.0.0.1:5432/postgres password=supersecret",
    "https://user:secret@example.test/path",
    "Authorization: Bearer abc.def.ghi",
    "X-Cartulary-Test-Route-Token: route-secret",
    "minio_secret_access_key=minio-secret",
    "-----BEGIN PRIVATE KEY-----\nsecret\n-----END PRIVATE KEY-----",
  ].join("\n"));
  for (const leaked of [
    "supersecret",
    "secret@example",
    "abc.def.ghi",
    "route-secret",
    "minio-secret",
    "BEGIN PRIVATE KEY",
  ]) {
    assert.equal(raw.includes(leaked), false, `raw redaction leaked ${leaked}: ${raw}`);
  }
});
