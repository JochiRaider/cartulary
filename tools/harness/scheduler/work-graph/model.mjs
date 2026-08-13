import { readFileSync } from "node:fs";
import path from "node:path";

import { validateSchemaSync } from "../../contract/index.mjs";
import {
  canonicalJSONString,
  semanticJSONDigest,
} from "../../test-catalog/index.mjs";

const graphSchemaID = "cartulary.harness_work_graph.v2";

export function loadWorkGraphOwner(root) {
  const owner = JSON.parse(
    readFileSync(path.join(root, "tools/harness_work_graph_owner.json"), "utf8"),
  );
  validateSchemaSync(owner.schema_id, owner);
  return owner;
}

function semanticUnit(unit) {
  const { semantic_digest: _digest, ...semantic } = unit;
  return semantic;
}

export function unitSemanticDigest(unit) {
  return semanticJSONDigest(semanticUnit(unit));
}

export function finalizeUnit(unit) {
  const complete = { ...unit, semantic_digest: "" };
  complete.semantic_digest = unitSemanticDigest(complete);
  return complete;
}

export function graphSemanticDigest(units) {
  return semanticJSONDigest({ units });
}

function assertSortedUnique(values, label) {
  const sorted = [...values].sort((left, right) =>
    left < right ? -1 : left > right ? 1 : 0,
  );
  if (canonicalJSONString(values) !== canonicalJSONString(sorted)) {
    throw new Error(`${label} must be sorted`);
  }
  if (new Set(values).size !== values.length) {
    throw new Error(`${label} contains a duplicate`);
  }
}

export function validateWorkGraph(graph, { capacities } = {}) {
  validateSchemaSync(graphSchemaID, graph);
  const unitIDs = graph.units.map((unit) => unit.unit_id);
  assertSortedUnique(unitIDs, "graph.units.unit_id");
  const byID = new Map(graph.units.map((unit) => [unit.unit_id, unit]));

  for (const unit of graph.units) {
    if (Object.keys(unit.resource_claims).length === 0) {
      throw new Error(`${unit.unit_id}.resource_claims must bound executable work`);
    }
    assertSortedUnique(unit.needs, `${unit.unit_id}.needs`);
    assertSortedUnique(
      unit.service_dependencies,
      `${unit.unit_id}.service_dependencies`,
    );
    assertSortedUnique(unit.shared_locks ?? [], `${unit.unit_id}.shared_locks`);
    assertSortedUnique(unit.exclusive_locks ?? [], `${unit.unit_id}.exclusive_locks`);
    for (const lock of unit.shared_locks ?? []) {
      if ((unit.exclusive_locks ?? []).includes(lock)) {
        throw new Error(`${unit.unit_id} cannot hold ${lock} as both shared and exclusive`);
      }
    }
    if (Object.keys(unit.resource_claims).length > 0) {
      const hostModes = Number((unit.shared_locks ?? []).includes("host_activity")) +
        Number((unit.exclusive_locks ?? []).includes("host_activity"));
      if (hostModes !== 1) {
        throw new Error(`${unit.unit_id} must hold exactly one host_activity lock mode`);
      }
    }
    assertSortedUnique(
      unit.evidence_outputs,
      `${unit.unit_id}.evidence_outputs`,
    );
    for (const output of unit.evidence_outputs) {
      if (
        output.startsWith("/") ||
        output.includes("\\") ||
        output.split("/").includes("..")
      ) {
        throw new Error(`${unit.unit_id} has unsafe evidence output ${output}`);
      }
    }
    if (unit.semantic_digest !== unitSemanticDigest(unit)) {
      throw new Error(`${unit.unit_id}.semantic_digest is stale`);
    }
    for (const dependency of unit.needs) {
      if (!byID.has(dependency)) {
        throw new Error(`${unit.unit_id} needs unknown unit ${dependency}`);
      }
      if (dependency === unit.unit_id) {
        throw new Error(`${unit.unit_id} cannot depend on itself`);
      }
    }
    if (capacities) {
      for (const [resource, amount] of Object.entries(unit.resource_claims)) {
        const limit = capacities.get(resource);
        if (!Number.isInteger(limit) || limit < amount) {
          throw new Error(
            `${unit.unit_id} has infeasible claim ${resource}=${amount}; capacity=${limit ?? "unknown"}`,
          );
        }
      }
    }
  }

  const visiting = new Set();
  const visited = new Set();
  function visit(unitID) {
    if (visiting.has(unitID)) throw new Error(`work graph contains cycle at ${unitID}`);
    if (visited.has(unitID)) return;
    visiting.add(unitID);
    for (const dependency of byID.get(unitID).needs) visit(dependency);
    visiting.delete(unitID);
    visited.add(unitID);
  }
  for (const unitID of unitIDs) visit(unitID);

  if (graph.graph_digest !== graphSemanticDigest(graph.units)) {
    throw new Error("graph.graph_digest is stale");
  }
  const expectedSelector = `units:${semanticJSONDigest(unitIDs)}`;
  if (graph.selector !== expectedSelector) {
    throw new Error("graph.selector does not identify its canonical unit set");
  }
  return graph;
}

export function buildWorkGraph(units) {
  const byID = new Map();
  for (const rawUnit of units) {
    if (!Array.isArray(rawUnit.service_dependencies)) {
      throw new Error(`${rawUnit.unit_id}.service_dependencies is required`);
    }
    assertSortedUnique(
      rawUnit.service_dependencies,
      `${rawUnit.unit_id}.service_dependencies`,
    );
    const safeID = rawUnit.unit_id.replaceAll(/[^A-Za-z0-9_.-]+/gu, "-");
    const unitResult = `unit-results/${safeID}.json`;
    const sharedLocks = [...(rawUnit.shared_locks ?? [])];
    const exclusiveLocks = [...(rawUnit.exclusive_locks ?? [])];
    const hostModes = Number(sharedLocks.includes("host_activity")) +
      Number(exclusiveLocks.includes("host_activity"));
    if (Object.keys(rawUnit.resource_claims).length > 0 && hostModes === 0) {
      sharedLocks.push("host_activity");
    } else if (hostModes > 1) {
      throw new Error(`${rawUnit.unit_id} cannot hold multiple host_activity lock modes`);
    }
    const unit = finalizeUnit({
      ...rawUnit,
      service_dependencies: [...rawUnit.service_dependencies],
      shared_locks: [...new Set(sharedLocks)].sort(),
      exclusive_locks: [...new Set(exclusiveLocks)].sort(),
      evidence_outputs: [...new Set([...rawUnit.evidence_outputs, unitResult])].sort(),
    });
    const prior = byID.get(unit.unit_id);
    if (prior && canonicalJSONString(prior) !== canonicalJSONString(unit)) {
      throw new Error(`conflicting semantic units share ${unit.unit_id}`);
    }
    byID.set(unit.unit_id, unit);
  }
  const canonicalUnits = [...byID.values()].sort((left, right) =>
    left.unit_id < right.unit_id ? -1 : left.unit_id > right.unit_id ? 1 : 0,
  );
  const unitIDs = canonicalUnits.map((unit) => unit.unit_id);
  const graph = {
    schema_id: graphSchemaID,
    selector: `units:${semanticJSONDigest(unitIDs)}`,
    units: canonicalUnits,
    graph_digest: graphSemanticDigest(canonicalUnits),
  };
  return validateWorkGraph(graph);
}
