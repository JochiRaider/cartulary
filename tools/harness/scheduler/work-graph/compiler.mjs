import { readFileSync } from "node:fs";
import path from "node:path";

import {
  loadTestCatalog,
  targetForCatalogRow,
} from "../../test-catalog/index.mjs";
import { planGoLPTShards } from "../../backend/go-lpt-shards.mjs";
import { buildWorkGraph, loadWorkGraphOwner } from "./model.mjs";
import {
  browserTargetStage,
  compileBrowserRowSelectionGraph,
  compileBrowserStageGraph,
} from "./browser.mjs";

function readJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

function compareASCII(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function assertSortedUniqueInput(values, label) {
  if (new Set(values).size !== values.length) {
    throw new Error(`${label} contains duplicate values`);
  }
  return [...values].sort(compareASCII);
}

function commandTargetMap(taskSurface) {
  return new Map(
    taskSurface.targets
      .filter((target) => target.command_id)
      .map((target) => [target.command_id, target.name]),
  );
}

function fixtureCapability(row) {
  const capability = row.fixture_capability;
  if (!capability) throw new Error(`${row.row_id} has no explicit fixture capability`);
  return capability;
}

function resourceClaims(row, owner) {
  const claims = { ...owner.runner_resource_claims[row.runner] };
  const capability = fixtureCapability(row);
  if (capability.startsWith("postgres_")) claims.postgres = 1;
  if (capability === "object_store_namespace") claims.object_store = 1;
  if (capability === "managed_process") claims.service_stack = 1;
  return Object.fromEntries(Object.entries(claims).sort(([left], [right]) => compareASCII(left, right)));
}

function fixtureEnvironment(capability) {
  const postgresPolicies = {
    postgres_transaction: "transaction",
    postgres_group: "group_clone",
    postgres_dedicated: "template_clone",
    // Migration leases also admit current-head isolation checks in the same
    // semantic row. MigrationDatabaseT owns its fresh replay identity; any
    // PrepareIsolatedDatabaseT call remains explicitly assigned to a clone.
    postgres_migration: "template_clone",
  };
  return {
    CARTULARY_FIXTURE_CAPABILITY: capability,
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
      ...fixtureEnvironment(row.fixture_capability),
      ...runtimeEnvironment,
    },
  };
}

function rowCachePolicy(row) {
  return (
    !new Set(["go", "vitest"]).has(row.runner) ||
    new Set(["managed_process", "postgres_migration", "browser_stack"]).has(row.fixture_capability) ||
    new Set(["measurement", "visual", "release"]).has(row.evidence_class)
      ? "none"
      : "content_addressed"
  );
}

function policyEvidenceOutputs(target) {
  if (target === "go-vulncheck") {
    return [
      "unit-artifacts/target-go-vulncheck/govulncheck-findings.json",
      "unit-artifacts/target-go-vulncheck/govulncheck-output.jsonstream",
    ];
  }
  return [];
}

function rowUnit(row, target, owner, needs, runtimeEnvironment) {
  return {
    unit_id: `row:${row.row_id}`,
    owner_id: row.owner_id,
    kind: "runner",
    command: rowCommand(row, target, runtimeEnvironment),
    needs,
    resource_claims: resourceClaims(row, owner),
    fixture_lease: fixtureCapability(row),
    cache_policy: rowCachePolicy(row),
    timeout_ms: owner.default_timeout_ms,
    evidence_outputs: [`rows/${row.row_id}.json`],
    failure_policy: {
      block_descendants: true,
      continue_independent: true,
      aggregate_effect: "required",
    },
    estimated_work_ms: owner.evidence_estimates_ms[row.evidence_class],
  };
}

function goShardUnit(shard, rows, target, owner, needs, runtimeEnvironment) {
  const rowIDs = shard.item_ids;
  const first = rows[0];
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
        ...fixtureEnvironment(first.fixture_capability),
        ...runtimeEnvironment,
      },
    },
    needs,
    resource_claims: resourceClaims(first, owner),
    fixture_lease: fixtureCapability(first),
    cache_policy: rows.every((row) => rowCachePolicy(row) === "content_addressed")
      ? "content_addressed"
      : "none",
    timeout_ms: owner.default_timeout_ms,
    evidence_outputs: rowIDs.map((rowID) => `rows/${rowID}.json`).sort(compareASCII),
    failure_policy: {
      block_descendants: true,
      continue_independent: true,
      aggregate_effect: "required",
    },
    estimated_work_ms: shard.estimated_work_ms,
  };
}

function rawGoUnit(entry, owner, needs) {
  const claims = { ...owner.runner_resource_claims.go };
  if (entry.fixture_capability.startsWith("postgres_")) claims.postgres = 1;
  return {
    unit_id: `raw_go:${entry.id}`,
    owner_id: "harness.backend",
    kind: "runner",
    command: {
      executable: process.env.GO || "go",
      args: [
        "test",
        "-json",
        "-count=1",
        "-p=1",
        "-run",
        entry.selection_pattern,
        ...entry.packages,
      ],
      environment: {
        CARTULARY_TEST_TARGET: entry.target,
        ...fixtureEnvironment(entry.fixture_capability),
      },
    },
    needs,
    resource_claims: Object.fromEntries(Object.entries(claims).sort(([left], [right]) => compareASCII(left, right))),
    fixture_lease: entry.fixture_capability,
    cache_policy: "none",
    timeout_ms: owner.default_timeout_ms,
    evidence_outputs: [],
    failure_policy: {
      block_descendants: true,
      continue_independent: true,
      aggregate_effect: "required",
    },
    estimated_work_ms: entry.estimated_work_ms,
  };
}

export class WorkGraphCompiler {
  constructor(root) {
    this.root = root;
    this.owner = loadWorkGraphOwner(root);
    this.taskSurface = readJSON(path.join(root, "tools/task_surface_owner.json"));
    this.topology = readJSON(path.join(root, "tools/execution_topology_manifest.json"));
    this.runtimeBinaryProducer = new Map(
      this.topology.runtime_binaries.map((binary) => [binary.id, binary.producer_target]),
    );
    this.familyRuntimeBinaries = new Map(
      this.topology.go_targets.family_runtime_binaries.map((family) => [family.family_id, family.runtime_binary_ids]),
    );
    this.rawGoAggregates = this.topology.go_targets.raw_go_aggregates ?? [];
    this.commandTargets = commandTargetMap(this.taskSurface);
    this.availableGoLanes = 4;
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
      const target = targetForCatalogRow(row, {
        commandTargetByID: this.commandTargets,
      });
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

  compileRows(rowIDs) {
    const sortedRowIDs = assertSortedUniqueInput(rowIDs, "row selection");
    const rows = sortedRowIDs.map((rowID) => {
      const row = this.catalog.rowByID.get(rowID);
      if (!row) throw new Error(`unknown active row ${rowID}`);
      return row;
    });
    const prerequisiteTargetsForRow = (row) => {
      const targets = new Set();
      if (row.runner === "vitest") targets.add("frontend-install");
      if (row.runner === "playwright") {
        for (const target of [
          "playwright-install",
          "build-web",
          "build-server-harness",
          "build-migrate",
          "test-service-images",
        ]) {
          targets.add(target);
        }
      }
      if (row.fixture_capability !== "none") targets.add("test-service-images");
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
    const terminalIDs = (graph) => {
      const dependencies = new Set(graph.units.flatMap((unit) => unit.needs));
      return graph.units
        .filter((unit) => !dependencies.has(unit.unit_id))
        .map((unit) => unit.unit_id)
        .sort(compareASCII);
    };
    const needsForRow = (row) => prerequisiteTargetsForRow(row)
      .flatMap((target) => terminalIDs(prerequisiteGraphs.get(target)))
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
    const playwrightRows = rows.filter((row) => row.runner === "playwright");
    const otherRows = rows.filter(
      (row) => row.runner !== "go" && row.runner !== "playwright",
    );
    const goUnits = [];
    if (goRows.length > 0) {
      const plan = planGoLPTShards(
        goRows.map((row) => ({
          id: row.row_id,
          estimated_work_ms: this.owner.evidence_estimates_ms[row.evidence_class],
          compatibility: {
            target: this.rowTargets.get(row.row_id),
            runtime_profile_id: row.runtime_profile_id,
            resource_profile_id: row.resource_profile_id,
            fixture_capability: row.fixture_capability,
            runtime_binary_ids: this.familyRuntimeBinaries.get(row.family_id) ?? [],
          },
          isolated: new Set(["managed_process", "postgres_dedicated", "postgres_migration"]).has(row.fixture_capability),
        })),
        { availableGoLanes: this.availableGoLanes },
      );
      for (const shard of plan.shards) {
        const shardRows = shard.item_ids.map((rowID) => this.catalog.rowByID.get(rowID));
        goUnits.push(goShardUnit(
          shard,
          shardRows,
          this.rowTargets.get(shard.item_ids[0]),
          this.owner,
          needsForRow(shardRows[0]),
          runtimeEnvironmentForRow(shardRows[0]),
        ));
      }
    }
    const browserGraph =
      playwrightRows.length === 0
        ? buildWorkGraph([])
        : compileBrowserRowSelectionGraph(
            this.root,
            this.owner,
            playwrightRows.map((row) => row.row_id),
          );
    const browserRoots = new Set(
      browserGraph.units
        .filter((unit) => unit.needs.length === 0)
        .map((unit) => unit.unit_id),
    );
    const browserPrerequisites = playwrightRows
      .flatMap((row) => needsForRow(row))
      .filter((unitID, index, values) => values.indexOf(unitID) === index)
      .sort(compareASCII);
    return buildWorkGraph([
      ...prerequisiteUnits,
      ...goUnits,
      ...browserGraph.units.map((unit) =>
        browserRoots.has(unit.unit_id)
          ? { ...unit, needs: browserPrerequisites }
          : unit,
      ),
      ...otherRows.map((row) =>
        rowUnit(
          row,
          this.rowTargets.get(row.row_id),
          this.owner,
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
    if (rowIDs === undefined) return this.compileRows(ownerRows);
    for (const rowID of rowIDs) {
      if (!ownerRows.includes(rowID)) {
        throw new Error(`${rowID} does not belong to ${ownerID}`);
      }
    }
    return this.compileRows(rowIDs);
  }

  compileTarget(target) {
    if (this.owner.target_members[target]) {
      return buildWorkGraph(
        this.owner.target_members[target]
          .flatMap((member) => this.compileTarget(member).units),
      );
    }
    if (this.owner.policy_units[target]) {
      return this.compilePolicyTarget(target);
    }
    return this.compileBaseTarget(target);
  }

  compileBaseTarget(target, { policyOnly = false } = {}) {
    const taskTarget = this.taskSurface.targets.find((entry) => entry.name === target);
    if (!taskTarget) throw new Error(`unknown task target ${target}`);
    const browser = browserTargetStage(this.root, target);
    if (browser) {
      const readinessTargets = [
        "playwright-install",
        "build-web",
        "build-server-harness",
        "build-migrate",
        "test-service-images",
      ];
      const readinessUnits = readinessTargets
        .flatMap((readinessTarget) => this.compilePolicyTarget(readinessTarget).units);
      const dependedOn = new Set(readinessUnits.flatMap((unit) => unit.needs));
      const terminals = readinessUnits
        .filter((unit) => !dependedOn.has(unit.unit_id))
        .map((unit) => unit.unit_id)
        .sort(compareASCII);
      const browserGraph = compileBrowserStageGraph(this.root, this.owner, browser.stage, {
        mode: browser.mode,
      });
      const roots = new Set(browserGraph.units.filter((unit) => unit.needs.length === 0).map((unit) => unit.unit_id));
      return buildWorkGraph([
        ...readinessUnits,
        ...browserGraph.units.map((unit) => roots.has(unit.unit_id) ? { ...unit, needs: terminals } : unit),
      ]);
    }
    const rowIDs = policyOnly ? [] : this.targetRows.get(target) ?? [];
    const rawGo = policyOnly ? buildWorkGraph([]) : this.compileRawGoTarget(target);
    if (rowIDs.length > 0) return buildWorkGraph([...this.compileRows(rowIDs).units, ...rawGo.units]);
    if (rawGo.units.length > 0) return rawGo;
    const recipe = this.taskSurface.make_recipes[target];
    const kind = recipe?.type === "artifact_binding" || (taskTarget.side_effects ?? []).some(
      (effect) => effect.class === "generated_artifacts" || effect.class === "build_outputs",
    )
      ? "artifact"
      : "policy";
    const policy = this.owner.policy_units[target];
    const fixtureLease = policy?.fixture_lease ?? "none";
    const resourceClaims = { cpu: 1, io: 1, memory_mb: 128, process: 1 };
    if (fixtureLease.startsWith("postgres_")) resourceClaims.postgres = 1;
    if (fixtureLease === "object_store_namespace") resourceClaims.object_store = 1;
    if (fixtureLease === "managed_process") resourceClaims.service_stack = 1;
    return buildWorkGraph([
      {
        unit_id: `target:${target}`,
        owner_id: "harness.command_surface",
        kind,
        command: {
          executable: "make",
          args: ["--silent", "--no-print-directory", target],
          environment: {
            CARTULARY_TEST_TARGET: target,
            ...(recipe?.type === "artifact_binding"
              ? { CARTULARY_HARNESS_GRAPH_ARTIFACT_CHILD: "1" }
              : {}),
          },
        },
        needs: [],
        resource_claims: resourceClaims,
        fixture_lease: fixtureLease,
        cache_policy: policy?.cache_policy ?? "none",
        timeout_ms: this.owner.default_timeout_ms,
        evidence_outputs: policyEvidenceOutputs(target),
        failure_policy: {
          block_descendants: true,
          continue_independent: true,
          aggregate_effect: "required",
        },
        estimated_work_ms: policy?.estimated_work_ms ?? 1000,
      },
    ]);
  }

  compileRawGoTarget(target) {
    const entries = this.rawGoAggregates.filter((entry) => entry.target === target);
    if (entries.length === 0) return buildWorkGraph([]);
    const needsServices = entries.some((entry) => entry.fixture_capability !== "none");
    const readiness = needsServices ? this.compilePolicyTarget("test-service-images") : buildWorkGraph([]);
    const dependencies = new Set(readiness.units.flatMap((unit) => unit.needs));
    const terminals = readiness.units
      .filter((unit) => !dependencies.has(unit.unit_id))
      .map((unit) => unit.unit_id)
      .sort(compareASCII);
    return buildWorkGraph([
      ...readiness.units,
      ...entries.map((entry) => rawGoUnit(entry, this.owner, entry.fixture_capability === "none" ? [] : terminals)),
    ]);
  }

  compilePolicyTarget(target, stack = []) {
    if (stack.includes(target)) {
      throw new Error(`policy dependency cycle ${[...stack, target].join(" -> ")}`);
    }
    const definition = this.owner.policy_units[target];
    if (!definition) return this.compileBaseTarget(target);
    const base = this.compileBaseTarget(target, { policyOnly: true });
    const dependencyGraphs = [
      ...definition.needs.map((dependency) =>
        this.owner.target_members[dependency]
          ? this.compileTarget(dependency)
          : this.compilePolicyTarget(dependency, [...stack, target]),
      ),
      ...(definition.owner_slices ?? []).map((ownerID) => this.compileOwner(ownerID)),
    ];
    const dependencyUnits = dependencyGraphs.flatMap((graph) => graph.units);
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
    return buildWorkGraph([...dependencyUnits, ...attachedBase]);
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
      } else if (unit.evidence_outputs.some((output) => output.startsWith("rows/"))) {
        for (const output of unit.evidence_outputs.filter((value) => value.startsWith("rows/"))) {
          const rowID = output.slice("rows/".length, -".json".length);
          add(this.rowTargets.get(rowID), unit.unit_id);
        }
      } else if (unit.unit_id.startsWith("target:")) {
        add(unit.unit_id.slice("target:".length), unit.unit_id);
      } else if (unit.unit_id.startsWith("raw_go:")) {
        add(unit.command.environment.CARTULARY_TEST_TARGET, unit.unit_id);
      } else if (unit.unit_id.startsWith("browser_group:")) {
        add(unit.command.environment.CARTULARY_TEST_TARGET, unit.unit_id);
      } else if (unit.unit_id.startsWith("browser_evidence:")) {
        add(unit.unit_id.slice("browser_evidence:".length), unit.unit_id);
      } else if (unit.unit_id.startsWith("browser_target_summary:")) {
        add(unit.unit_id.slice("browser_target_summary:".length), unit.unit_id);
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
    const maximumTier = this.owner.aggregate_tiers[target];
    if (!maximumTier) throw new Error(`unknown aggregate ${target}`);
    const maximumRank = this.owner.tier_order.indexOf(maximumTier);
    const rowIDs = this.catalog.rows
      .filter(
        (row) => this.owner.tier_order.indexOf(row.minimum_tier) <= maximumRank,
      )
      .map((row) => row.row_id);
    const rows = rowIDs.map((rowID) => this.catalog.rowByID.get(rowID));
    const selectedTargets = new Set(rows.map((row) => this.rowTargets.get(row.row_id)));
    const browserTargets = new Set();
    const policyTargets = new Set(this.aggregatePolicyRoots(target));
    const policyRows = new Map();
    const ordinaryRows = [];
    for (const row of rows) {
      const rowTarget = this.rowTargets.get(row.row_id);
      if (row.runner === "playwright") browserTargets.add(rowTarget);
      else if (this.owner.policy_units[rowTarget]) {
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
    const forcedOwnerIDs = [...new Set(
      [...policyTargets].flatMap((policyTarget) =>
        this.owner.policy_units[policyTarget]?.owner_slices ?? [],
      ),
    )].sort(compareASCII);
    const forcedOwnerRowIDs = new Set(
      this.catalog.rows
        .filter((row) => forcedOwnerIDs.includes(row.owner_id))
        .map((row) => row.row_id),
    );
    const ordinaryTierRows = ordinaryRows.filter((rowID) => !forcedOwnerRowIDs.has(rowID));
    const units = ordinaryTierRows.length > 0 ? this.compileRows(ordinaryTierRows).units : [];
    const forcedOwnerGraphs = new Map(
      forcedOwnerIDs.map((ownerID) => [ownerID, this.compileOwner(ownerID)]),
    );
    for (const graph of forcedOwnerGraphs.values()) units.push(...graph.units);
    for (const selectedTarget of [...selectedTargets].sort(compareASCII)) {
      units.push(...this.compileRawGoTarget(selectedTarget).units);
    }
    for (const browserTarget of [...browserTargets].sort(compareASCII)) {
      units.push(...this.compileTarget(browserTarget).units);
    }
    const policyGraphs = new Map();
    for (const policyTarget of [...policyTargets].sort(compareASCII)) {
      const selectedRows = policyRows.get(policyTarget);
      const graph = selectedRows
        ? this.compileRows(selectedRows)
        : this.owner.target_members[policyTarget]
          ? this.compileTarget(policyTarget)
          : this.compileBaseTarget(policyTarget);
      policyGraphs.set(policyTarget, graph);
    }
    const terminalIDs = (graph) => {
      const dependencies = new Set(graph.units.flatMap((unit) => unit.needs));
      return graph.units
        .filter((unit) => !dependencies.has(unit.unit_id))
        .map((unit) => unit.unit_id)
        .sort(compareASCII);
    };
    for (const [policyTarget, graph] of policyGraphs) {
      const definition = this.owner.policy_units[policyTarget];
      const dependencyIDs = [
        ...(definition?.needs ?? [])
          .flatMap((dependency) => terminalIDs(policyGraphs.get(dependency))),
        ...(definition?.owner_slices ?? [])
          .flatMap((ownerID) => terminalIDs(forcedOwnerGraphs.get(ownerID))),
      ].filter((unitID, index, values) => values.indexOf(unitID) === index)
        .sort(compareASCII);
      const roots = new Set(
        graph.units.filter((unit) => unit.needs.length === 0).map((unit) => unit.unit_id),
      );
      units.push(
        ...graph.units.map((unit) =>
          roots.has(unit.unit_id) && dependencyIDs.length > 0
            ? { ...unit, needs: dependencyIDs }
            : unit,
        ),
      );
    }
    return buildWorkGraph(units);
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
