import { loadExecutionTopology, resolveBrowserFrontendArtifact } from "../../generated-artifacts/execution-topology.mjs";
import path from "node:path";
import { semanticJSONDigest, resolveCommandInputs } from "../../contract/index.mjs";

import {
  loadTestCatalog,
  resolveVitestTitles,
  targetForCatalogRow,
} from "../../test-catalog/index.mjs";
import { packInvocationCohorts } from "./cohorts.mjs";
import { planGoLPTShards } from "../../backend/go-lpt-shards.mjs";
import { buildWorkGraph, loadWorkGraphOwner } from "./model.mjs";
import { operationDefinition, operationExecutionDigest } from "./operations.mjs";
import { policyDependencyRowIDs } from "./policy-rows.mjs";
import { resolveCommandSelection, resolvePrivateSelection } from "./command-selection.mjs";
import {
  selectBrowserRowStages,
  compileBrowserStageGraph,
} from "./browser.mjs";
import {
  assertFixtureServiceDependencies,
  assertServiceDependencies,
  topologyResourceClaims,
} from "./resource-claims.mjs";

function compareASCII(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function goRowProcessIsolation(row) {
  return row.fixture_capability === "managed_process" || row.process_isolation === "exclusive"
    ? "exclusive"
    : "shared";
}

function assertSortedUniqueInput(values, label) {
  if (new Set(values).size !== values.length) {
    throw new Error(`${label} contains duplicate values`);
  }
  return [...values].sort(compareASCII);
}


function fixtureCapability(row) {
  const capability = row.fixture_capability;
  if (!capability) throw new Error(`${row.row_id} has no explicit fixture capability`);
  return capability;
}

function resourceClaims(row, topology) {
  assertFixtureServiceDependencies(
    row.fixture_capability,
    row.service_dependencies,
    row.row_id,
  );
  assertServiceDependencies(
    topology,
    row.runtime_profile_id,
    row.service_dependencies,
    row.row_id,
  );
  return topologyResourceClaims(
    topology,
    row.resource_profile_id,
    row.service_dependencies,
    row.row_id,
  );
}

function fixtureEnvironment(capability, serviceDependencies) {
  const postgresPolicies = {
    postgres_transaction: "transaction",
    postgres_dedicated: "template_clone",
    // Migration leases also admit current-head isolation checks in the same
    // semantic row. MigrationDatabaseT owns its fresh replay identity; any
    // PrepareIsolatedDatabaseT call remains explicitly assigned to a clone.
    postgres_migration: "template_clone",
  };
  return {
    CARTULARY_FIXTURE_CAPABILITY: capability,
    CARTULARY_HARNESS_SERVICE_DEPENDENCIES: serviceDependencies.join(","),
    ...(postgresPolicies[capability]
      ? { CARTULARY_POSTGRES_FIXTURE_POLICY_DEFAULT: postgresPolicies[capability] }
      : {}),
  };
}

function rowCommand(row, target, runtimeEnvironment = {}) {
  return {
    executable: "node",
    args: ["tools/harness/execution/runners/row-runner-cli.mjs", "--row-id", row.row_id],
    environment: {
      CARTULARY_TEST_ROWS: row.row_id,
      ...(target ? { CARTULARY_TEST_TARGET: target } : {}),
      ...fixtureEnvironment(row.fixture_capability, row.service_dependencies),
      ...runtimeEnvironment,
    },
  };
}

function rowCachePolicy(row) {
  return (
    !new Set(["go", "vitest"]).has(row.runner) ||
    row.fixture_capability !== "none" ||
    row.service_dependencies.length > 0 ||
    new Set(["measurement", "security", "visual", "release"]).has(row.evidence_class)
      ? "none"
      : "content_addressed"
  );
}

function boundedGoClaims(claims, requestedParallelism, availablePostgresLanes) {
  if (!Object.hasOwn(claims, "postgres")) {
    return { claims, parallelism: requestedParallelism };
  }
  const parallelism = Math.max(
    1,
    Math.min(
      requestedParallelism,
      availablePostgresLanes,
      claims.postgres,
    ),
  );
  return {
    claims: Object.fromEntries(
      Object.entries({ ...claims, cpu: parallelism, postgres: parallelism }).sort(
        ([left], [right]) => compareASCII(left, right),
      ),
    ),
    parallelism,
  };
}

function policyEvidenceOutputs(target) {
  if (target === "test-evidence-audit") return ["test-evidence-audit/audit.json"];
  if (target === "build-web" || target === "build-web-measurement") {
    return [`${target}/frontend-artifact.json`];
  }
  if (target === "go-vulncheck") {
    return [
      "unit-artifacts/operation-harness.go_vulncheck/govulncheck-findings.json",
      "unit-artifacts/operation-harness.go_vulncheck/govulncheck-output.jsonstream",
    ];
  }
  return [];
}

function rowUnit(row, target, owner, topology, needs, runtimeEnvironment) {
  return {
    unit_id: `row:${row.row_id}`,
    owner_id: row.owner_id,
    kind: "runner",
    command: rowCommand(row, target, {
      ...runtimeEnvironment,
      ...(row.runner === "shell" ? {
        CARTULARY_OPERATION_DIGEST: operationExecutionDigest(operationDefinition(owner, row.selector.operation_id)),
      } : {}),
    }),
    needs,
    resource_claims: resourceClaims(row, topology),
    fixture_lease: fixtureCapability(row),
    postgres_isolation: row.postgres_isolation,
    service_dependencies: row.service_dependencies,
    cache_policy: rowCachePolicy(row),
    timeout_ms: owner.default_timeout_ms,
    current_run_evidence_outputs: [`rows/${row.row_id}.json`],
    row_bindings: [{ row_id: row.row_id, runner: row.runner }],
    failure_policy: {
      block_descendants: true,
      continue_independent: true,
      aggregate_effect: "required",
    },
    estimated_work_ms: owner.evidence_estimates_ms[row.evidence_class],
  };
}

function goShardUnit(
  shard,
  rows,
  target,
  owner,
  topology,
  needs,
  runtimeEnvironment,
  availablePostgresLanes,
) {
  const rowIDs = shard.item_ids;
  const first = rows[0];
  const baseClaims = resourceClaims(first, topology);
  baseClaims.cpu = shard.cpu_tokens;
  const { claims, parallelism } = boundedGoClaims(
    baseClaims,
    shard.cpu_tokens,
    availablePostgresLanes,
  );
  return {
    unit_id: shard.shard_id,
    owner_id: "harness.backend",
    kind: "runner",
    command: {
      executable: "node",
      args: ["tools/harness/execution/runners/row-runner-cli.mjs", "--row-ids", rowIDs.join(",")],
      environment: {
        CARTULARY_TEST_ROWS: rowIDs.join(","),
        CARTULARY_TEST_TARGET: target,
        CARTULARY_UNIT_CPU_TOKENS: String(parallelism),
        GOMAXPROCS: String(parallelism),
        ...fixtureEnvironment(first.fixture_capability, first.service_dependencies),
        ...runtimeEnvironment,
      },
    },
    needs,
    resource_claims: claims,
    fixture_lease: fixtureCapability(first),
    postgres_isolation: first.postgres_isolation,
    service_dependencies: first.service_dependencies,
    cache_policy: rows.every((row) => rowCachePolicy(row) === "content_addressed")
      ? "content_addressed"
      : "none",
    timeout_ms: owner.default_timeout_ms,
    current_run_evidence_outputs: rowIDs
      .map((rowID) => `rows/${rowID}.json`)
      .sort(compareASCII),
    row_bindings: rows.map(({ row_id, runner }) => ({ row_id, runner })),
    failure_policy: {
      block_descendants: true,
      continue_independent: true,
      aggregate_effect: "required",
    },
    // Package transactions and buckets can serialize even with multiple CPU tokens.
    // Sum selected work; charge process startup/settlement only once.
    estimated_work_ms: owner.invocation_policies.go.startup_ms + shard.estimated_work_ms +
      owner.invocation_policies.go.cleanup_ms,
  };
}

export class WorkGraphCompiler {
  buildGraph(units) { return buildWorkGraph(units, this.topology); }
  constructor(root, { targetInputs = {} } = {}) {
    this.root = root;
    this.targetInputs = targetInputs;
    this.owner = loadWorkGraphOwner(root);
    const topology = loadExecutionTopology({
      root,
      manifestPath: path.resolve(root, "tools/execution_topology_manifest.json"),
    });
    this.taskSurface = topology.taskSurface;
    this.topology = topology.raw;
    this.browserLaneLimit = this.topology.browser_lane_policy.default;
    this.allowExplicitRows = true;
    this.runtimeBinaryProducer = new Map(
      this.topology.runtime_binaries.map((binary) => [binary.id, binary.producer_operation]),
    );
    this.familyRuntimeBinaries = new Map(
      this.topology.go_targets.family_runtime_binaries.map((family) => [family.family_id, family.runtime_binary_ids]),
    );
    this.availableGoLanes = 4;
    this.availablePostgresLanes = 4;
    this._catalog = null;
    this._targetRows = null;
    this._rowTargets = null;
  }

  ensureCatalog() {
    if (this._catalog) return;
    this._catalog = loadTestCatalog(this.root);
    this._targetRows = new Map();
    this._rowTargets = new Map();
    for (const row of this._catalog.rows) {
      const target = targetForCatalogRow(row);
      this._rowTargets.set(row.row_id, target);
      const rows = this._targetRows.get(target) ?? [];
      rows.push(row.row_id);
      this._targetRows.set(target, rows);
    }
  }

  get catalog() {
    this.ensureCatalog();
    return this._catalog;
  }

  get targetRows() {
    this.ensureCatalog();
    return this._targetRows;
  }

  get rowTargets() {
    this.ensureCatalog();
    return this._rowTargets;
  }

  compileRows(rowIDs, stack = [], { browserMode = "validation", goShuffleSeed } = {}) {
    const sortedRowIDs = assertSortedUniqueInput(rowIDs, "row selection");
    const rows = sortedRowIDs.map((rowID) => {
      const row = this.catalog.rowByID.get(rowID);
      if (!row) throw new Error(`unknown active row ${rowID}`);
      if (!this.allowExplicitRows && row.aggregate_placement === "explicit_only") {
        throw new Error(`policy dependency reintroduces explicit-only row ${rowID}`);
      }
      return row;
    });
    const profiles = new Set(rows.filter((row) => row.runner === "playwright" && row.resource_profile_id === "browser_functional").map((row) => row.runtime_profile_id));
    if (profiles.size > this.browserLaneLimit) throw new Error(`browser functional closure has ${profiles.size} runtime profiles but only ${this.browserLaneLimit} lanes`);
    const browserArtifactsByRow = new Map();
    const selectedBrowserStages = selectBrowserRowStages(this.root, rows.filter((row) => row.runner === "playwright").map((row) => row.row_id));
    for (const stage of selectedBrowserStages) {
      const artifact = resolveBrowserFrontendArtifact(this.root, stage.name, this.topology);
      for (const group of stage.groups) for (const rowID of group.selectedRowIDs) browserArtifactsByRow.set(rowID, artifact);
    }
    const prerequisiteTargetsForRow = (row) => {
      const targets = new Set(row.preparation_operations ?? []);
      if (row.runner === "vitest") targets.add("harness.frontend_install");
      if (row.runner === "vitest" && row.selector.project_id === "renderer-node" && row.evidence_class === "integration") {
        targets.add("harness.playwright_install");
      }
      if (row.runner === "playwright") {
        for (const target of [
          "harness.playwright_install",
          browserArtifactsByRow.get(row.row_id).producer_operation,
          "harness.build_server_harness",
          "harness.build_migrate",
          "harness.test_service_images",
        ]) {
          targets.add(target);
        }
      }
      if (row.service_dependencies.length > 0) targets.add("harness.test_service_images");
      for (const binaryID of this.familyRuntimeBinaries.get(row.family_id) ?? []) {
        const producer = this.runtimeBinaryProducer.get(binaryID);
        if (!producer) throw new Error(`${row.family_id} references unknown runtime binary ${binaryID}`);
        targets.add(producer);
      }
      return [...targets].sort(compareASCII);
    };
    const prerequisiteTargets = new Set();
    for (const row of rows) {
      for (const target of prerequisiteTargetsForRow(row)) prerequisiteTargets.add(target);
    }
    const prerequisiteGraphs = new Map([...prerequisiteTargets]
      .sort(compareASCII)
      .map((target) => [target, this.compilePolicyTarget(target)]));
    const prerequisiteUnits = [...prerequisiteGraphs.values()].flatMap((graph) => graph.units);
    const declaredGraphs = new Map(rows.map((row) => {
      const target = this.rowTargets.get(row.row_id);
      return [row.row_id, this.compileDependencies(target, stack)];
    }));
    for (const graph of declaredGraphs.values()) prerequisiteUnits.push(...graph.units);
    const terminalIDs = (graph) => {
      const dependencies = new Set(graph.units.flatMap((unit) => unit.needs));
      return graph.units
        .filter((unit) => !dependencies.has(unit.unit_id))
        .map((unit) => unit.unit_id)
        .sort(compareASCII);
    };
    const needsForRow = (row) => [
      ...prerequisiteTargetsForRow(row).flatMap((target) => terminalIDs(prerequisiteGraphs.get(target))),
      ...terminalIDs(declaredGraphs.get(row.row_id)),
    ]
      .filter((unitID, index, values) => values.indexOf(unitID) === index)
      .sort(compareASCII);
    const runtimeEnvironmentForRow = (row) => Object.fromEntries(
      (this.familyRuntimeBinaries.get(row.family_id) ?? [])
        .map((binaryID) => this.topology.runtime_binaries.find((binary) => binary.id === binaryID))
        .map((binary) => [
          binary.consumer_env,
          path.resolve(this.root, process.env[binary.output_make_variable] || binary.default_output_path),
        ])
        .sort(([left], [right]) => compareASCII(left, right)),
    );
    const goRows = rows.filter((row) => row.runner === "go");
    const explicitContexts = this.owner.row_execution_contexts ?? {};
    const otherRows = rows.filter(
      (row) => row.runner !== "go" && row.runner !== "playwright" && row.runner !== "vitest",
    );
    const vitestGroups = new Map();
    const exactVitestTitles = resolveVitestTitles(this.root, rows.filter((row) => row.runner === "vitest"));
    for (const row of rows.filter((row) => row.runner === "vitest")) {
      const base = rowUnit(row, this.rowTargets.get(row.row_id), this.owner, this.topology, needsForRow(row), runtimeEnvironmentForRow(row));
      const { CARTULARY_TEST_ROWS: _selection, ...environment } = base.command.environment;
      // Reporting ownership does not alter the physical execution contract.
      const { owner_id: _owner, unit_id: _id, command: _command, current_run_evidence_outputs: _outputs, row_bindings: _rows, estimated_work_ms: _estimate, ...contract } = base;
      const key = semanticJSONDigest({ contract, environment, project: row.selector.project_id,
        evidence_class: row.evidence_class, runtime_profile_id: row.runtime_profile_id, resource_profile_id: row.resource_profile_id,
        fixture_profile_id: row.fixture_profile_id ?? null,
        isolation: row.process_isolation === "exclusive" || row.fixture_capability === "managed_process" ? row.row_id : null });
      const group = vitestGroups.get(key) ?? { key, base, rows: [] };
      group.rows.push(row);
      vitestGroups.set(key, group);
    }
    const vitestPolicy = this.owner.invocation_policies.vitest;
    const vitestUnits = [...vitestGroups.values()].flatMap(({ key, base, rows: selected }) => {
      const files = new Map();
      for (const row of selected) {
        const item = files.get(row.selector.file) ?? { id: row.selector.file, cost: 0, rows: [] };
        item.cost += this.owner.evidence_estimates_ms[row.evidence_class];
        item.rows.push(row);
        files.set(item.id, item);
      }
      return packInvocationCohorts([...files.values()], vitestPolicy, base.timeout_ms).map((cohort) => {
        const selectedRows = cohort.items.flatMap((item) => item.rows).sort((a, b) => compareASCII(a.row_id, b.row_id));
        const ids = selectedRows.map((row) => row.row_id);
        const selections = selectedRows.map((row) => ({ row_id: row.row_id, project: row.selector.project_id,
          file: row.selector.file, titles: exactVitestTitles.get(row.row_id) }));
        const claims = { ...base.resource_claims };
        const { CARTULARY_TEST_ROWS: _selection, ...environment } = base.command.environment;
        for (const [resource, amount] of Object.entries(vitestPolicy.coordinator_claims)) claims[resource] = (claims[resource] ?? 0) + amount;
        return { ...base, owner_id: new Set(selectedRows.map((row) => row.owner_id)).size === 1 ? selectedRows[0].owner_id : "harness.command_surface", unit_id: `vitest_cohort:${semanticJSONDigest({ key, selections, policy: vitestPolicy, workers: 1, isolate: true, settlement: "public_fork_stop" }).slice("sha256:".length)}`,
          command: { ...base.command, args: ["tools/harness/execution/runners/row-runner-cli.mjs", "--unit-rows"],
            environment: { ...environment, VITEST_MAX_WORKERS: "1" } },
          resource_claims: Object.fromEntries(Object.entries(claims).sort(([a], [b]) => compareASCII(a, b))),
          current_run_evidence_outputs: ids.map((id) => `rows/${id}.json`),
          row_bindings: selectedRows.map(({ row_id, runner }) => ({ row_id, runner })),
          estimated_work_ms: cohort.cost };
      });
    });
    const goUnits = [];
    const goSelections = [goRows.filter((row) => !explicitContexts[row.row_id]),
      ...goRows.filter((row) => explicitContexts[row.row_id]).map((row) => [row])];
    for (const selectedGoRows of goSelections.filter((selection) => selection.length > 0)) {
      const context = explicitContexts[selectedGoRows[0].row_id];
      if (context && goRowProcessIsolation(selectedGoRows[0]) !== "exclusive") {
        throw new Error(`${selectedGoRows[0].row_id} requires exclusive process isolation for a fixed execution context`);
      }
      const plan = planGoLPTShards(
        selectedGoRows.map((row) => ({
          id: row.row_id,
          estimated_work_ms: this.owner.evidence_estimates_ms[row.evidence_class],
          compatibility: {
            target: this.rowTargets.get(row.row_id),
            package_selection: row.selector.package,
            evidence_class: row.evidence_class,
            runtime_profile_id: row.runtime_profile_id,
            resource_profile_id: row.resource_profile_id,
            fixture_capability: row.fixture_capability,
            postgres_isolation: row.postgres_isolation,
            fixture_profile_id: row.fixture_profile_id ?? null,
            process_isolation: goRowProcessIsolation(row),
            service_dependencies: row.service_dependencies,
            preparation_operations: row.preparation_operations ?? [],
            runtime_binary_ids: this.familyRuntimeBinaries.get(row.family_id) ?? [],
            ...(context ? { execution_context: context } : {}),
          },
          // Dedicated databases, migration databases, and ordinary object-store
          // namespaces are owned and eagerly finalized by each testing.T. They
          // may share a compatible Go process without sharing the resource.
          // managed_process is different: process lifecycle is the resource
          // under test, so each selected row retains its own child process.
          isolated: goRowProcessIsolation(row) === "exclusive",
        })),
        { availableGoLanes: context?.cpu_tokens ?? this.availableGoLanes },
      );
      for (const shard of plan.shards) {
        const shardRows = shard.item_ids.map((rowID) => this.catalog.rowByID.get(rowID));
        const unit = goShardUnit(
          shard,
          shardRows,
          this.rowTargets.get(shard.item_ids[0]),
          this.owner,
          this.topology,
          needsForRow(shardRows[0]),
          { ...runtimeEnvironmentForRow(shardRows[0]), CARTULARY_GO_RACE: context?.go_race ? "1" : "0", CARTULARY_GO_SHUFFLE_SEED: goShuffleSeed === undefined ? "" : String(goShuffleSeed) },
          this.availablePostgresLanes,
        );
        if (goShuffleSeed !== undefined) unit.unit_id += `:order-${goShuffleSeed}`;
        goUnits.push(context ? { ...unit, cache_policy: context.cache_policy } : unit);
      }
    }
    const browserUnits = selectedBrowserStages.flatMap((stage) => {
      const graph = compileBrowserStageGraph(this.root, this.owner, stage, { mode: browserMode, catalog: this.catalog, maxLanes: this.browserLaneLimit });
      const prerequisites = [...new Set(stage.groups.flatMap((group) => group.selectedRowIDs)
        .flatMap((rowID) => needsForRow(this.catalog.rowByID.get(rowID))))].sort(compareASCII);
      return graph.units.map((unit) => unit.needs.length === 0 ? { ...unit, needs: prerequisites } : unit);
    });
    return this.buildGraph([
      ...prerequisiteUnits,
      ...goUnits,
      ...vitestUnits,
      ...browserUnits,
      ...otherRows.map((row) =>
        rowUnit(
          row,
          this.rowTargets.get(row.row_id),
          this.owner,
          this.topology,
          needsForRow(row),
          runtimeEnvironmentForRow(row),
        ),
      ),
    ]);
  }

  compileOwner(ownerID, rowIDs) {
    const ownerRows = this.catalog.rows
      .filter((row) => row.owner_id === ownerID)
      .map((row) => row.row_id);
    if (ownerRows.length === 0) throw new Error(`unknown active owner ${ownerID}`);
    if (rowIDs === undefined) return this.withSelectionPolicy(false, () => this.compileRows(ownerRows.filter((id) => this.catalog.rowByID.get(id).aggregate_placement !== "explicit_only")));
    for (const rowID of rowIDs) {
      if (!ownerRows.includes(rowID)) {
        throw new Error(`${rowID} does not belong to ${ownerID}`);
      }
    }
    return this.compileRows(rowIDs);
  }

  compileTarget(target) {
    if (this.owner.target_members[target]) return this.compilePrivateSelection(target);
    if (this.owner.operations[target]) return this.compilePolicyTarget(target);
    const entry = this.taskSurface.targets.find((value) => value.name === target);
    if (!entry || entry.target_class !== "public") throw new Error(`unknown public command ${target}`);
    if (this.owner.aggregate_tiers[target]) return this.compileAggregate(target);
    if (entry.selection_binding) return this.compileCommand(target, this.targetInputs[target] ?? {}).graph;
    const recipe = this.taskSurface.make_recipes[target];
    if (recipe?.operation_id) {
      const inputs = resolveCommandInputs(target, this.targetInputs[target] ?? {}, this.taskSurface);
      return this.compilePolicyTarget(recipe.operation_id, [], this.forwardedInputs(entry, inputs));
    }
    throw new Error(`command ${target} has no execution selection`);
  }

  compileMember(member) {
    if (typeof member === "string") return this.compilePrivateSelection(member);
    if (member.row_id) return this.compileRows([member.row_id]);
    if (member.operation_id) return this.compilePolicyTarget(member.operation_id);
    throw new Error("unknown work-graph member binding");
  }

  compilePrivateSelection(selection, stack = []) {
    const selected = resolvePrivateSelection(this, selection);
    const units = selected.rows.length ? this.compileRows(selected.rows, stack).units : [];
    for (const id of selected.operations) units.push(...this.compilePolicyTarget(id, stack).units);
    return this.attachDependencies(this.buildGraph(units), this.compileDependencies(selection, stack));
  }

  compileCommand(target, inputs) {
    const selection = resolveCommandSelection(this, target, inputs);
    return this.withSelectionPolicy(selection.explicit, () => this.compileResolvedCommand(target, selection));
  }

  withSelectionPolicy(explicit, compile) {
    const previous = this.allowExplicitRows;
    this.allowExplicitRows = previous && explicit;
    try { return compile(); } finally { this.allowExplicitRows = previous; }
  }

  resolveBrowserLaneLimit(value) {
    const policy = this.topology.browser_lane_policy;
    const limit = value === undefined ? policy.default : typeof value === "string" && /^[1-9][0-9]*$/u.test(value) ? Number(value) : value;
    if (!Number.isSafeInteger(limit) || limit < 1 || limit > policy.maximum) throw new Error("invalid BROWSER_LANES");
    this.browserLaneLimit = limit;
    return limit;
  }

  compileResolvedCommand(target, selection) {
    this.resolveBrowserLaneLimit(selection.inputs.BROWSER_LANES);
    const binding = this.taskSurface.targets.find((entry) => entry.name === target).selection_binding;
    const units = selection.rows.length ? this.compileRows(selection.rows, [], { browserMode: binding.browser_mode ?? "validation", goShuffleSeed: selection.inputs.GO_SHUFFLE_SEED }).units : [];
    const entry = this.taskSurface.targets.find((value) => value.name === target);
    const forwardedInputs = this.forwardedInputs(entry, selection.inputs);
    for (const id of selection.operations) units.push(...this.compilePolicyTarget(id, [], forwardedInputs).units);
    return { graph: this.buildGraph(units), inputs: selection.inputs };
  }

  forwardedInputs(entry, inputs) {
    const controls = new Set(["CHECK", "COMPONENT", "SUITE", "OWNER", "ROWS", "SCOPE", "JSON", "DETAIL", "BROWSER_LANES", "GO_SHUFFLE_SEED"]);
    // Owner is executable input for the evidence audit, but a selection control
    // for owner slices. Only the declared private action consumes it.
    if (!entry.selection_binding) controls.delete("OWNER");
    return Object.fromEntries((entry.input_contract?.inputs ?? [])
      .filter((input) => input.child_forwarding === "argv" && input.summary_emission !== "redacted_value" &&
        !controls.has(input.name) && Object.hasOwn(inputs, input.name))
      .map((input) => [input.name, inputs[input.name]]));
  }

  policyKeyForTarget(operationID) { return operationID; }
  policyForTarget(operationID) { return this.owner.policy_units[operationID]; }

  compileOperation(operationID, { forwardedInputs = {}, policy } = {}) {
    const operation = operationDefinition(this.owner, operationID);
    const target = operation.report_key;
    policy ??= this.owner.policy_units[operationID];
    const fixtureLease = policy?.fixture_lease ?? "none";
    const serviceDependencies = policy?.service_dependencies ?? [];
    assertFixtureServiceDependencies(fixtureLease, serviceDependencies, `policy ${target}`);
    const resourceClaims = topologyResourceClaims(
      this.topology,
      policy?.resource_profile_id ?? "standard",
      serviceDependencies,
      `policy ${target}`,
    );
    return this.buildGraph([
      {
        unit_id: `operation:${operationID}`,
        owner_id: "harness.command_surface",
        kind: operation.kind,
        command: {
          executable: "node",
          args: ["tools/harness/execution/operation-runner-cli.mjs", operationID],
          environment: {
            ...forwardedInputs,
            CARTULARY_TEST_TARGET: operation.report_key,
            CARTULARY_OPERATION_DIGEST: operationExecutionDigest(operation),
            CARTULARY_HARNESS_SERVICE_DEPENDENCIES: serviceDependencies.join(","),
            ...(operation.execution.type === "artifact_binding"
              ? { CARTULARY_HARNESS_GRAPH_ARTIFACT_CHILD: "1" }
              : {}),
          },
        },
        needs: [],
        resource_claims: resourceClaims,
        fixture_lease: fixtureLease,
        postgres_isolation: policy?.postgres_isolation ?? "none",
        service_dependencies: serviceDependencies,
        cache_policy: policy?.cache_policy ?? "none",
        timeout_ms: Object.hasOwn(policy ?? {}, "timeout_ms")
          ? policy.timeout_ms
          : this.owner.default_timeout_ms,
        current_run_evidence_outputs: policyEvidenceOutputs(operation.report_key),
        reusable_artifact_outputs: policy?.reusable_artifact_outputs ?? [],
        failure_policy: {
          block_descendants: true,
          continue_independent: true,
          aggregate_effect: "required",
        },
        estimated_work_ms: policy?.estimated_work_ms ?? 1000,
      },
    ]);
  }

  compileDependencies(target, stack = []) {
    if (stack.includes(target)) {
      throw new Error(`policy dependency cycle ${[...stack, target].join(" -> ")}`);
    }
    const definition = this.policyForTarget(target);
    if (!definition) return this.buildGraph([]);
    const dependencyGraphs = [
      ...definition.needs.map((dependency) =>
        this.owner.target_members[dependency]
          ? this.compilePrivateSelection(dependency, [...stack, target])
          : this.compilePolicyTarget(dependency, [...stack, target]),
      ),
      ...((definition.required_rows ?? []).length > 0
        ? [this.compileRows(policyDependencyRowIDs(this.owner, this.policyKeyForTarget(target), this.catalog.rows), [...stack, target])]
        : []),
    ];
    return this.buildGraph(dependencyGraphs.flatMap((graph) => graph.units));
  }

  compilePolicyTarget(operationID, stack = [], forwardedInputs = {}) {
    if (this.owner.target_members[operationID]) return this.compilePrivateSelection(operationID, stack);
    operationDefinition(this.owner, operationID);
    const rowIDs = this.targetRows.get(operationID) ?? [];
    if (rowIDs.length > 0) return this.compileRows(rowIDs, stack);
    return this.attachDependencies(this.compileOperation(operationID, { forwardedInputs }), this.compileDependencies(operationID, stack));
  }

  attachDependencies(base, dependencies) {
    const dependencyUnits = dependencies.units;
    const neededByAnother = new Set(
      dependencyUnits.flatMap((unit) => unit.needs),
    );
    const dependencyTerminals = dependencyUnits
      .filter((unit) => !neededByAnother.has(unit.unit_id))
      .map((unit) => unit.unit_id)
      .sort(compareASCII);
    const baseRoots = new Set(
      base.units.filter((unit) => unit.needs.length === 0).map((unit) => unit.unit_id),
    );
    const attachedBase = base.units.map((unit) =>
      baseRoots.has(unit.unit_id) && dependencyTerminals.length > 0
        ? { ...unit, needs: [...dependencyTerminals] }
        : unit,
    );
    return this.buildGraph([...dependencyUnits, ...attachedBase]);
  }

  aggregatePolicyRoots(target, seen = new Set()) {
    if (seen.has(target)) throw new Error(`aggregate policy inheritance cycle at ${target}`);
    const next = new Set(seen).add(target);
    return [
      ...(this.owner.aggregate_policy_roots[target] ?? []),
      ...(this.owner.aggregate_policy_inherits[target] ?? []).flatMap((parent) =>
        this.aggregatePolicyRoots(parent, next),
      ),
    ].sort(compareASCII);
  }

  aggregateProjections(target, graph) {
    const byID = new Map(graph.units.map((unit) => [unit.unit_id, unit]));
    const members = new Map();
    const add = (projection, unitID) => {
      if (!projection) return;
      const values = members.get(projection) ?? new Set();
      values.add(unitID);
      members.set(projection, values);
    };
    for (const unit of graph.units) {
      if (unit.unit_id.startsWith("row:")) {
        add(this.rowTargets.get(unit.unit_id.slice("row:".length)), unit.unit_id);
      } else if (
        unit.current_run_evidence_outputs.some((output) => output.startsWith("rows/"))
      ) {
        for (const output of unit.current_run_evidence_outputs.filter((value) =>
          value.startsWith("rows/"),
        )) {
          const rowID = output.slice("rows/".length, -".json".length);
          add(this.rowTargets.get(rowID), unit.unit_id);
        }
      } else if (unit.unit_id.startsWith("operation:")) {
        add(unit.unit_id.slice("operation:".length), unit.unit_id);
      } else if (unit.unit_id.startsWith("browser_group:")) {
        add(unit.command.environment.CARTULARY_TEST_TARGET, unit.unit_id);
      } else if (unit.unit_id.startsWith("browser_evidence:")) {
        add(unit.unit_id.slice("browser_evidence:".length), unit.unit_id);
      } else if (unit.unit_id.startsWith("browser_target_summary:")) {
        add(unit.unit_id.slice("browser_target_summary:".length), unit.unit_id);
      }
    }
    for (const check of this.taskSurface.harness_checks.filter((entry) => entry.gate_smoke_role)) {
      const candidate = this.compilePrivateSelection(check.selection);
      if (candidate.units.every((unit) => byID.get(unit.unit_id)?.semantic_digest === unit.semantic_digest)) {
        for (const unit of candidate.units) add(check.name, unit.unit_id);
      }
    }
    // A public default view exists only when its complete, identical operation
    // graph is already present. Reporting never adds executable work.
    for (const entry of this.taskSurface.targets) {
      const recipe = this.taskSurface.make_recipes[entry.name];
      if (entry.name === target || this.owner.aggregate_tiers[entry.name] ||
          entry.selection_binding?.kind === "owner_or_suite" ||
          entry.selection_binding?.browser_mode ||
          (entry.input_contract?.inputs ?? []).some((input) => input.required) ||
          (entry.side_effects ?? []).some((effect) => effect.class === "authored_source_write") ||
          !(entry.selection_binding && recipe?.type === "work_graph" || recipe?.operation_id && recipe.graph_entry === true)) continue;
      const candidate = this.compileTarget(entry.name);
      if (candidate.units.length && candidate.units.every((unit) => byID.get(unit.unit_id)?.semantic_digest === unit.semantic_digest)) {
        for (const unit of candidate.units) add(entry.name, unit.unit_id);
      }
    }
    const includeDependencies = (unitID, values) => {
      for (const dependency of byID.get(unitID).needs) {
        if (values.has(dependency)) continue;
        values.add(dependency);
        includeDependencies(dependency, values);
      }
    };
    for (const values of members.values()) {
      for (const unitID of [...values]) includeDependencies(unitID, values);
    }
    return {
      [target]: graph.units.map((unit) => unit.unit_id),
      ...Object.fromEntries(
        [...members.entries()]
          .sort(([left], [right]) => compareASCII(left, right))
          .map(([name, values]) => [name, [...values].sort(compareASCII)]),
      ),
    };
  }

  compileAggregatePlan(target) {
    const graph = this.compileAggregate(target);
    return {
      graph,
      projections: this.aggregateProjections(target, graph),
    };
  }

  compileAggregate(target) {
    const inputs = resolveCommandInputs(target, this.targetInputs[target] ?? {}, this.taskSurface);
    this.resolveBrowserLaneLimit(inputs.BROWSER_LANES);
    return this.withSelectionPolicy(false, () => this.compileAggregateRows(target));
  }

  compileAggregateRows(target) {
    const maximumTier = this.owner.aggregate_tiers[target];
    if (!maximumTier) throw new Error(`unknown aggregate ${target}`);
    const maximumRank = this.owner.tier_order.indexOf(maximumTier);
    const admitted = (entry) => {
      if (entry.aggregate_placement === "explicit_only") return false;
      const rank = this.owner.tier_order.indexOf(entry.aggregate_placement);
      if (rank < 0) throw new Error(`invalid aggregate placement for ${entry.row_id ?? entry.id}`);
      return rank <= maximumRank;
    };
    const rowIDs = this.catalog.rows
      .filter(admitted)
      .map((row) => row.row_id);
    const rows = rowIDs.map((rowID) => this.catalog.rowByID.get(rowID));
    const policyTargets = new Set(this.aggregatePolicyRoots(target));
    const policyRows = new Map();
    const ordinaryRows = [];
    for (const row of rows) {
      const rowTarget = this.rowTargets.get(row.row_id);
      if (row.runner !== "playwright" && this.owner.policy_units[rowTarget]) {
        policyTargets.add(rowTarget);
        const selected = policyRows.get(rowTarget) ?? [];
        selected.push(row.row_id);
        policyRows.set(rowTarget, selected);
      }
      else ordinaryRows.push(row.row_id);
    }
    const expandPolicy = (policyTarget) => {
      if (policyTargets.has(policyTarget)) {
        // The target may have been present before its dependency closure was walked.
      } else {
        policyTargets.add(policyTarget);
      }
      for (const dependency of this.owner.policy_units[policyTarget]?.needs ?? []) {
        if (!policyTargets.has(dependency)) {
          policyTargets.add(dependency);
          expandPolicy(dependency);
        }
      }
    };
    for (const policyTarget of [...policyTargets]) expandPolicy(policyTarget);
    const forcedRowIDs = [...new Set(
      [...policyTargets].filter((policyTarget) => this.owner.policy_units[policyTarget]).flatMap((policyTarget) =>
        policyDependencyRowIDs(this.owner, policyTarget, this.catalog.rows),
      ),
    )].sort(compareASCII);
    const forcedRows = new Set(forcedRowIDs);
    const ordinaryTierRows = ordinaryRows.filter((rowID) => !forcedRows.has(rowID));
    const units = ordinaryTierRows.length > 0 ? this.compileRows(ordinaryTierRows).units : [];
    // Each semantic consumer is constructed with its dependencies by the same
    // path used for direct and owner selection. Never reattach dependencies to
    // roots of an expanded graph (those roots can be the producers themselves).
    for (const policyTarget of [...policyTargets].sort(compareASCII)) {
      const selectedRows = policyRows.get(policyTarget);
      const graph = selectedRows
        ? this.compileRows(selectedRows)
        : this.owner.target_members[policyTarget]
          ? this.compileTarget(policyTarget)
          : this.compilePolicyTarget(policyTarget);
      units.push(...graph.units);
    }
    return this.buildGraph(units);
  }

  compile(selection) {
    if (selection.kind === "target") return this.compileTarget(selection.target);
    if (selection.kind === "aggregate") return this.compileAggregate(selection.target);
    if (selection.kind === "owner") {
      return this.compileOwner(selection.owner_id, selection.row_ids);
    }
    if (selection.kind === "rows") return this.compileRows(selection.row_ids);
    throw new Error(`unsupported graph selection ${selection.kind}`);
  }
}
