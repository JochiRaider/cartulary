#!/usr/bin/env node

import { readFileSync } from "node:fs";
import path from "node:path";

import { validateSchemaSync } from "../contract/index.mjs";
import { WorkGraphCompiler } from "../scheduler/work-graph/index.mjs";

const root = path.resolve(import.meta.dirname, "../../..");
const aggregateTargets = new Set(["test-fast", "test", "check", "ci", "release-check"]);

function usage() {
  throw new Error("usage: target-plan-cli.mjs [--json] [--detail] [--target <target>]");
}

function parseArgs(argv) {
  const options = { detail: false, json: false, target: "check" };
  for (let index = 0; index < argv.length; index += 1) {
    if (argv[index] === "--json") options.json = true;
    else if (argv[index] === "--detail") options.detail = true;
    else if (argv[index] === "--target") options.target = argv[++index] ?? "";
    else usage();
  }
  if (!options.target) usage();
  return options;
}

function planFor(target) {
  const taskSurface = JSON.parse(readFileSync(path.join(root, "tools/task_surface_owner.json"), "utf8"));
  if (!taskSurface.targets.some((entry) => entry.name === target)) {
    throw new Error(`unknown target ${target}`);
  }
  const compiler = new WorkGraphCompiler(root);
  const compiled = aggregateTargets.has(target)
    ? compiler.compileAggregatePlan(target)
    : { graph: compiler.compile({ kind: "target", target }), projections: null };
  const projections = compiled.projections ?? {
    [target]: compiled.graph.units.map((unit) => unit.unit_id),
  };
  const plan = {
    schema_id: "cartulary.harness_target_plan.v2",
    target,
    graph_digest: compiled.graph.graph_digest,
    projections,
    units: compiled.graph.units.map((unit) => ({
      unit_id: unit.unit_id,
      owner_id: unit.owner_id,
      kind: unit.kind,
      needs: unit.needs,
      resource_claims: unit.resource_claims,
      fixture_capability: unit.fixture_lease,
      service_dependencies: unit.service_dependencies,
      cache_policy: unit.cache_policy,
      estimated_work_ms: unit.estimated_work_ms,
      semantic_digest: unit.semantic_digest,
      evidence_outputs: unit.evidence_outputs,
    })),
  };
  validateSchemaSync(plan.schema_id, plan);
  return plan;
}

function human(plan, detail) {
  if (!detail) {
    const cached = plan.units.filter((unit) => unit.cache_policy !== "none").length;
    const fixtures = plan.units.filter((unit) => unit.fixture_capability !== "none").length;
    return `target=${plan.target} units=${plan.units.length} cached=${cached} fixtures=${fixtures} digest=${plan.graph_digest}`;
  }
  return [
    `target ${plan.target}`,
    `graph ${plan.graph_digest}`,
    ...plan.units.map((unit) =>
      `${unit.unit_id} kind=${unit.kind} owner=${unit.owner_id} needs=${unit.needs.join(",") || "-"} fixture=${unit.fixture_capability} services=${unit.service_dependencies.join(",") || "-"} cache=${unit.cache_policy} work_ms=${unit.estimated_work_ms}`,
    ),
  ].join("\n");
}

try {
  const options = parseArgs(process.argv.slice(2));
  const plan = planFor(options.target);
  process.stdout.write(options.json ? `${JSON.stringify(plan, null, 2)}\n` : `${human(plan, options.detail)}\n`);
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exitCode = 2;
}
