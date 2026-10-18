import { readFileSync } from "node:fs";
import path from "node:path";
import { hasMakeNodeTool } from "../../command-surface/make-node-tools.mjs";
import { semanticJSONDigest } from "../../contract/index.mjs";

/** Operation definitions belong to the graph owner, independently of command names. */
export function readOperationOwner(root) {
  const owner = JSON.parse(readFileSync(path.join(root, "tools/harness_work_graph_owner.json"), "utf8"));
  for (const [id, operation] of Object.entries(owner.operations ?? {})) {
    operationDefinition(owner, id);
    if (["node_tool", "owner_command"].includes(operation.execution.type) && !hasMakeNodeTool(operation.execution.tool_id)) throw new Error(`${id}: unknown private tool binding ${operation.execution.tool_id}`);
  }
  return owner;
}

export function operationDefinition(owner, operationID) {
  if (typeof operationID !== "string" || !/^harness\.[a-z][a-z0-9_]*$/u.test(operationID) || !Object.hasOwn(owner.operations ?? {}, operationID)) {
    throw new Error(`unknown harness operation ${operationID}`);
  }
  return owner.operations[operationID];
}

export function operationExecutionDigest(operation) {
  // The owner already separates reporting identity from executable behavior.
  // Hash that complete boundary so new execution fields cannot silently evade
  // dispatch freshness or reuse identity (including evidence publication).
  return semanticJSONDigest(operation.execution);
}

// Public wrappers project the selected operation contract. Graph dispatch always
// resolves the operation owner; reporting labels never select executable behavior.
export function projectOperationBindings(surface, owner) {
  const projected = structuredClone(surface);
  const privateBindings = operationBindingNames(owner);
  for (const [name, binding] of Object.entries(projected.make_recipes)) {
    if (["work_graph", "ui_review"].includes(binding.type)) binding.binding_names = privateBindings;
    if (binding.type === "bootstrap") {
      const selection = projected.targets.find((entry) => entry.name === name)?.selection_binding;
      const ids = [...new Set([...Object.values(selection?.variants ?? {}), selection?.default_selection])];
      for (const id of ids) {
        if (operationDefinition(owner, id).execution.type !== "installation") throw new Error("bootstrap bindings require installation operations");
      }
      binding.binding_names = operationBindingNames({ operations: Object.fromEntries(ids.map((id) => [id, owner.operations[id]])) });
      continue;
    }
    if (binding.type !== "operation") continue;
    if (Object.keys(binding).some((key) => !["type", "operation_id"].includes(key))) {
      throw new Error(`${name}: operation bindings cannot override executable behavior`);
    }
    const operation = operationDefinition(owner, binding.operation_id);
    projected.make_recipes[name] = {
      ...structuredClone(operation.execution),
      operation_id: binding.operation_id,
    };
  }
  return projected;
}

export function operationBindingNames(owner) {
  const names = new Set();
  for (const operation of Object.values(owner.operations)) {
    for (const match of JSON.stringify(operation.execution).matchAll(/\$\(([A-Z][A-Z0-9_]*)\)/gu)) {
      names.add(match[1]);
    }
  }
  return [...names].sort();
}

function expand(value, environment) {
  // No shell, Make expansion, or recursive expansion is permitted at dispatch.
  if (typeof value !== "string" || value.replace(/\$\(([A-Z][A-Z0-9_]*)\)/gu, "").includes("$(")) {
    throw new Error("operation templates accept only named environment bindings");
  }
  const result = value.replace(/\$\(([A-Z][A-Z0-9_]*)\)/gu, (_, name) => {
    if (!Object.hasOwn(environment, name) || typeof environment[name] !== "string") throw new Error(`missing operation binding ${name}`);
    return environment[name];
  });
  return result;
}

function literalToken(token) {
  return token.startsWith('"') && token.endsWith('"') ? token.slice(1, -1) : token;
}

export function operationEnvironment(operation, environment) {
  const env = { ...environment, CARTULARY_TEST_TARGET: operation.report_key };
  for (const [name, value] of Object.entries({ ...operation.execution.exports, ...operation.execution.env })) env[name] = expand(value, environment);
  return env;
}

// Installation is a single lifecycle operation: Make produces its declared
// files, then an optional acquisition action runs. It never schedules graph work.
export function buildOperationSequence(owner, operationID, environment) {
  const operation = operationDefinition(owner, operationID);
  if (operation.execution.type !== "installation") return [buildOperationInvocation(owner, operationID, environment)];
  const withType = (type) => ({ operations: { [operationID]: { ...operation, execution: { ...operation.execution, type } } } });
  return [buildOperationInvocation(withType("artifact_binding"), operationID, environment),
    ...(operation.execution.command ? [buildOperationInvocation(withType("step_command"), operationID, environment)] : [])];
}

export function buildOperationInvocation(owner, operationID, environment) {
  const { execution, report_key: reportKey } = operationDefinition(owner, operationID);
  const env = operationEnvironment({ execution, report_key: reportKey }, environment);
  if (execution.type === "artifact_binding") {
    const files = execution.prerequisites.map((file) => expand(file, environment));
    if (files.some((file) => !path.isAbsolute(file) || file.includes("\0") || file.includes("\n"))) {
      throw new Error(`${operationID}: artifact operations require absolute file targets`);
    }
    return { command: environment.MAKE || "make", args: ["--silent", "--no-print-directory", "--", ...files], env };
  }
  if (execution.type !== "step_command") {
    throw new Error(`${operationID}: ${execution.type} requires a typed tool invocation`);
  }
  const tokens = execution.mode === "node"
    ? ["$(NODE_BIN)", execution.script, ...(execution.args ?? [])]
    : execution.command;
  const argv = tokens.flatMap((token) => {
    // These are formatting flags owned by Make, not user forwarding arguments.
    if (["$(BIOME_CHECK_FLAGS)", "$(BIOME_SCRIPT_CHECK_FLAGS)"].includes(token)) {
      return expand(token, environment).split(/\s+/u).filter(Boolean);
    }
    return [expand(literalToken(token), environment)];
  });
  if (!argv[0]) throw new Error(`${operationID}: missing executable`);
  if (tokens[0] === "$(MAKE)" || argv[0] === "make") {
    if (argv.slice(1).some((arg) => !["--silent", "--no-print-directory", "--"].includes(arg) && !path.isAbsolute(arg))) {
      throw new Error(`${operationID}: Make may produce files only; operation dependencies belong to the graph`);
    }
  }
  if (execution.mode === "run_step") {
    if (!environment.RUN_STEP_SCRIPT) throw new Error("missing operation binding RUN_STEP_SCRIPT");
    return { command: "bash", args: [environment.RUN_STEP_SCRIPT, execution.step_label, "--", ...argv], env };
  }
  return { command: argv[0], args: argv.slice(1), env };
}
