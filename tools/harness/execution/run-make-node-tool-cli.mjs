#!/usr/bin/env node

import { spawnSync } from "node:child_process";
import {
  existsSync,
  readFileSync,
  statSync,
} from "node:fs";
import path from "node:path";

import {
  buildMakeNodeToolChildEnv,
  buildMakeNodeToolInvocation,
  makeNodeToolNames,
  UsageError,
} from "../command-surface/make-node-tools.mjs";
import {
  classifyExecutionFailure,
  classifyExecutionFailureReason,
  compactJSONString,
  prettyJSONString,
  publicExitCodeForSummary,
  redactString,
  resolveRetainedArtifactIdentity,
  secureMkdir,
  secureWriteFile,
  validateSchema,
} from "../contract/index.mjs";
import {
  artifactLine,
  fileArtifactRef,
  buildToolRunSummary,
  machineOutput,
  normalizeOutputMode,
  quietLikeOutput,
  resultLine,
  terminalArtifactPath,
  toolRunSummarySchemaID,
  toolSummaryPath,
} from "../output/index.mjs";
import { finalizeObservabilitySafely, observabilityRequiredTarget } from "../observability/observability.mjs";

function nowUTC() {
  return new Date().toISOString();
}

function monotonicMs() {
  return Number(process.hrtime.bigint() / 1_000_000n);
}

function relToCwd(value) {
  const relative = path.relative(process.cwd(), value).replaceAll("\\", "/");
  if (!relative.startsWith("../") && relative !== "..") {
    return relative;
  }
  return value.replaceAll("\\", "/");
}

function loadTargetPolicy(target) {
  const manifestPath =
    process.env.TASK_SURFACE_MANIFEST || "tools/task_surface_manifest.json";
  const resolved = path.isAbsolute(manifestPath)
    ? manifestPath
    : path.join(process.cwd(), manifestPath);
  try {
    const manifest = JSON.parse(readFileSync(resolved, "utf8"));
    return manifest.targets?.find((entry) => entry.name === target)?.output_policy ?? null;
  } catch {
    return null;
  }
}

function shouldWrapTarget(target) {
  const policy = loadTargetPolicy(target);
  return policy?.summary_schema === toolRunSummarySchemaID;
}

function writeIfNonEmpty(file, content) {
  secureWriteFile(file, redactString(content));
  if (!content) {
    return null;
  }
  return file;
}

function runRelativePath(runRoot, file) {
  const relative = path.relative(runRoot, file).replaceAll("\\", "/");
  if (!relative || relative === ".." || relative.startsWith("../") || path.isAbsolute(relative)) {
    throw new Error(`${file} must stay under run root ${runRoot}`);
  }
  return relative;
}

function artifactIfExists(role, file, runRoot, kind = "log") {
  if (!file || !existsSync(file) || statSync(file).size === 0) {
    return null;
  }
  return fileArtifactRef(role, runRelativePath(runRoot, file), kind);
}

function readToolSummary(file, target) {
  if (!existsSync(file)) {
    return null;
  }
  try {
    const summary = JSON.parse(readFileSync(file, "utf8"));
    if (
      summary?.schema_id === toolRunSummarySchemaID &&
      summary.target === target &&
      ["pass", "fail"].includes(summary.status)
    ) {
      return summary;
    }
  } catch {
    return null;
  }
  return null;
}

function addUniqueArtifacts(existing, additions) {
  const artifacts = [];
  const seen = new Set();
  for (const artifact of [...(existing ?? []), ...additions.filter(Boolean)]) {
    const key = `${artifact.role}\0${artifact.path_kind}\0${artifact.format ?? ""}\0${artifact.path}`;
    if (seen.has(key)) {
      continue;
    }
    seen.add(key);
    artifacts.push(artifact);
  }
  return artifacts.sort((left, right) =>
    `${left.role}\0${left.path_kind}\0${left.format ?? ""}\0${left.path}`.localeCompare(
      `${right.role}\0${right.path_kind}\0${right.format ?? ""}\0${right.path}`,
    ),
  );
}

async function runWrapped(target, invocation) {
  const identity = resolveRetainedArtifactIdentity(
    target,
    process.env,
  );
  const resultsRoot = identity.result_root;
  const runID = identity.run_id;
  process.env.CARTULARY_TEST_RESULTS_DIR = resultsRoot;
  process.env.CARTULARY_TEST_RUN_ID = runID;
  const runRootAbs = path.join(resultsRoot, runID);
  const targetRootAbs = path.join(runRootAbs, target);
  secureMkdir(targetRootAbs);
  const stdoutLog = path.join(targetRootAbs, "stdout.log");
  const stderrLog = path.join(targetRootAbs, "stderr.log");
  const startedAt = nowUTC();
  const startedMs = monotonicMs();
  const child = spawnSync(process.execPath, [invocation.script, ...invocation.args], {
    env: buildMakeNodeToolChildEnv(target, process.env),
    encoding: "utf8",
    maxBuffer: 64 * 1024 * 1024,
    stdio: quietLikeOutput() ? ["ignore", "pipe", "pipe"] : "inherit",
  });
  if (child.error) {
    throw child.error;
  }
  const status = child.status ?? 1;
  const stdoutFile = quietLikeOutput()
    ? writeIfNonEmpty(stdoutLog, child.stdout ?? "")
    : null;
  const stderrFile = quietLikeOutput()
    ? writeIfNonEmpty(stderrLog, child.stderr ?? "")
    : null;
  const usageFailure = status === 2;
  const fallbackFailureClass = usageFailure
    ? "config"
    : classifyExecutionFailure(target, target, invocation.script);
  const fallbackFailureReason = usageFailure
    ? "usage_error"
    : classifyExecutionFailureReason(target, target, invocation.script);
  const fallbackFailureHeadline = usageFailure
    ? `${target} rejected invalid usage before child work`
    : `${target} failed`;
  const runRoot = relToCwd(runRootAbs);
  const summaryFile = toolSummaryPath(targetRootAbs);
  const summaryRel = runRelativePath(runRootAbs, summaryFile);
  const childSummary = readToolSummary(summaryFile, target);
  const summary = childSummary ?? buildToolRunSummary({
    target,
    command: ["make", target],
    status: status === 0 ? "pass" : "fail",
    exitCode: status,
    startedAt,
    completedAt: nowUTC(),
    durationMs: monotonicMs() - startedMs,
    outputMode: normalizeOutputMode(),
    resultRoot: relToCwd(resultsRoot),
    runId: runID,
    runRoot,
    summaryArtifacts: [fileArtifactRef("tool_run_summary", summaryRel)],
    counts: {},
    failureClass: status === 0 ? null : fallbackFailureClass,
    failureReason: status === 0 ? null : fallbackFailureReason,
    failures:
      status === 0
        ? []
        : [
            {
              target,
              label: target,
              failure_class: fallbackFailureClass,
              failure_reason: fallbackFailureReason,
              headline: fallbackFailureHeadline,
              artifact: stderrFile ? relToCwd(stderrFile) : "",
            },
          ],
    rerunCommands: [`make ${target}`],
  });
  summary.summary_artifacts = addUniqueArtifacts(summary.summary_artifacts, [
    fileArtifactRef("tool_run_summary", summaryRel),
  ]);
  summary.log_artifacts = addUniqueArtifacts(summary.log_artifacts, [
    artifactIfExists("stdout_log", stdoutFile, runRootAbs),
    artifactIfExists("stderr_log", stderrFile, runRootAbs),
  ]);
  summary.result_root ||= relToCwd(resultsRoot);
  summary.run_id ||= runID;
  summary.run_root ||= runRoot;
  summary.output_mode = normalizeOutputMode();
  summary.exit_code = publicExitCodeForSummary(summary);
  await validateSchema(toolRunSummarySchemaID, summary);
  secureWriteFile(summaryFile, prettyJSONString(summary));
  if (observabilityRequiredTarget(target)) {
    const observability = finalizeObservabilitySafely(runRootAbs, {
      target,
      status: summary.status === "pass" ? "passed" : "failed",
    });
    if (observability.status === "partial") {
      summary.warnings = [
        ...(summary.warnings ?? []),
        {
          kind: "harness_observability",
          status: "partial",
          diagnostic: observability.diagnostic,
        },
      ];
      await validateSchema(toolRunSummarySchemaID, summary);
      secureWriteFile(summaryFile, prettyJSONString(summary));
    }
  }

  if (machineOutput()) {
    process.stdout.write(compactJSONString(summary));
  } else if (summary.status === "pass") {
    process.stdout.write(resultLine(summary, summaryRel));
    process.stdout.write(
      artifactLine(summary, summaryRel, {
        investigate: `make explain-run RESULTS_DIR=${relToCwd(path.join(resultsRoot, runID))} TARGET=${target}`,
        log: stdoutFile ? relToCwd(stdoutFile) : null,
      }),
    );
  } else {
    const logArtifact = stderrFile ? relToCwd(stderrFile) : stdoutFile ? relToCwd(stdoutFile) : "-";
    process.stderr.write(
      `[FAIL] target=${target} exit_code=${summary.exit_code} failure_class=${summary.failure_class ?? "harness"} reason=${summary.failure_reason ?? "unknown_failure"} work_unit=- child_target=- duration_ms=${summary.duration_ms} headline="${target} failed"\n`,
    );
    process.stderr.write(
      `[ARTIFACTS] target=${target} root=${summary.run_root} summary_json=${terminalArtifactPath(summary.run_root, summaryRel)} log_artifact=${terminalArtifactPath(summary.run_root, logArtifact)} scheduler_json=- progress_log=-\n`,
    );
    process.stderr.write(`[RERUN] command="make ${target}"\n`);
    process.stderr.write(
      `[INVESTIGATE] command="make explain-run RESULTS_DIR=${relToCwd(path.join(resultsRoot, runID))} TARGET=${target} DETAIL=logs"\n`,
    );
  }
  return summary.exit_code;
}

function usage() {
  process.stderr.write(`usage: run-make-node-tool.mjs <${makeNodeToolNames().join("|")}>\n`);
}

function main(argv) {
  const [target, ...extra] = argv;
  if (!target || extra.length > 0) {
    usage();
    return 2;
  }

  let invocation;
  try {
    invocation = buildMakeNodeToolInvocation(target, process.env);
  } catch (error) {
    if (error instanceof UsageError) {
      process.stderr.write(`${error.usage ?? error.message}\n`);
      return 2;
    }
    throw error;
  }

  const child = spawnSync(process.execPath, [invocation.script, ...invocation.args], {
    env: buildMakeNodeToolChildEnv(target, process.env),
    stdio: "inherit",
  });
  if (child.error) {
    throw child.error;
  }
  return child.status ?? 1;
}

try {
  const [target, ...extra] = process.argv.slice(2);
  if (target && extra.length === 0) {
    let invocation;
    try {
      invocation = buildMakeNodeToolInvocation(target, process.env);
    } catch {
      invocation = null;
    }
    if (invocation && shouldWrapTarget(target)) {
      process.exit(await runWrapped(target, invocation));
    }
  }
  process.exit(main(process.argv.slice(2)));
} catch (error) {
  process.stderr.write(`make node tool failed: ${error.message}\n`);
  process.exit(1);
}
