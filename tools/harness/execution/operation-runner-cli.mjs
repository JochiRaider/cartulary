#!/usr/bin/env node
import { spawnSync } from "../workspace/child-process.mjs";
import path from "node:path";
import { reportCommandFailure } from "../runtime/command-failure.mjs";
import { publicExitCodeForFailure, redactString } from "../contract/index.mjs";
import { buildOperationSequence, operationDefinition, operationEnvironment, operationExecutionDigest, readOperationOwner } from "../scheduler/work-graph/operations.mjs";
import { buildMakeNodeToolInvocation, buildMakeNodeToolChildEnv } from "../command-surface/make-node-tools.mjs";

// Executes exactly one already-scheduled operation. It never acquires graph
// dependencies or starts another graph. Cancellation belongs to the native owner.
try {
  const [operationID, ...extra] = process.argv.slice(2);
  if (!operationID || extra.length) throw new Error("usage: operation-runner-cli.mjs <operation-id>");
  const owner = readOperationOwner(process.cwd());
  const operation = operationDefinition(owner, operationID);
  if (process.env.CARTULARY_OPERATION_DIGEST && process.env.CARTULARY_OPERATION_DIGEST !== operationExecutionDigest(operation)) {
    throw new Error("operation definition changed after graph compilation");
  }
  const recipe = operation.execution;
  const environment = { ...process.env, CURDIR: process.cwd(), NODE_BIN: process.execPath };
  let invocations;
  if (["node_tool", "owner_command"].includes(recipe.type)) {
    const env = operationEnvironment(operation, environment);
    const tool = buildMakeNodeToolInvocation(recipe.tool_id, env);
    invocations = [{ command: process.execPath, args: [tool.script, ...tool.args], env: buildMakeNodeToolChildEnv(recipe.tool_id, env) }];
  } else {
    invocations = buildOperationSequence(owner, operationID, environment);
  }
  let status = 0;
  for (const invocation of invocations) {
    const child = spawnSync(invocation.command, invocation.args, { env: invocation.env, stdio: "inherit" });
    if (child.error) throw child.error;
    status = child.status ?? 1;
    if (status !== 0) break;
  }
  if (status === 0 && (recipe.success_summary || recipe.type === "artifact_binding" && recipe.test_target === "self")) {
    const summary = spawnSync(process.execPath, [path.resolve("tools/harness/execution/cartulary-runner-cli.mjs"), "target-summary", operation.report_key, "pass"], { env: invocations.at(-1).env, stdio: "inherit" });
    if (summary.error) throw summary.error;
    status = summary.status ?? 1;
  }
  process.exitCode = status;
} catch (error) {
  process.stderr.write(`operation dispatch failed: ${redactString(error.message)}\n`);
  process.exitCode = publicExitCodeForFailure(reportCommandFailure(process.cwd(), error, { failure_class: "config", failure_reason: "configuration_error" }));
}
