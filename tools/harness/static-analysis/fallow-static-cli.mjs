#!/usr/bin/env node

import { spawnSync } from "node:child_process";
import {
  existsSync,
  mkdirSync,
  readFileSync,
  statSync,
} from "node:fs";
import path from "node:path";

import {
  repoRoot,
  resolveRetainedArtifactIdentity,
  secureWriteFile,
  validateSchemaSync,
} from "../core/public-contract.mjs";
import {
  artifactRef,
  buildToolRunSummary,
  normalizeOutputMode,
  toolSummaryPath,
} from "../core/public-contract.mjs";

const target = "frontend-fallow-static";
const fallowSummarySchemaID = "cartulary.fallow_static_summary.v1";
const fallowScript = path.join(repoRoot, "node_modules", "fallow", "bin", "fallow");
const configPath = path.join(repoRoot, ".fallowrc.json");
const issueArrayKeys = new Set([
  "boundary_violations",
  "boundaryviolations",
  "circular_dependencies",
  "circulardependencies",
  "duplicate_exports",
  "duplicateexports",
  "policy_violations",
  "policyviolations",
  "re_export_cycles",
  "reexportcycles",
  "stale_suppressions",
  "stalesuppressions",
  "test_only_dependencies",
  "testonlydependencies",
  "type_only_dependencies",
  "typeonlydependencies",
  "unlisted_dependencies",
  "unlisteddependencies",
  "unresolved_imports",
  "unresolvedimports",
  "unused_dependencies",
  "unuseddependencies",
  "unused_dev_dependencies",
  "unuseddevdependencies",
  "unused_exports",
  "unusedexports",
  "unused_files",
  "unusedfiles",
  "unused_optional_dependencies",
  "unusedoptionaldependencies",
  "unused_types",
  "unusedtypes",
]);

function nowUTC() {
  return new Date().toISOString();
}

function monotonicMs() {
  return Number(process.hrtime.bigint() / 1_000_000n);
}

function repoRel(file) {
  const relative = path.relative(repoRoot, file).replaceAll("\\", "/");
  if (!relative.startsWith("../") && relative !== "..") {
    return relative;
  }
  return file.replaceAll("\\", "/");
}

function mkdir(dir) {
  mkdirSync(dir, { recursive: true, mode: 0o700 });
}

function readJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

function normalizeKey(key) {
  return String(key ?? "").replaceAll("-", "_").replaceAll("_", "").toLowerCase();
}

function increment(map, key, count = 1) {
  const normalized = String(key ?? "").trim() || "unknown";
  map[normalized] = (map[normalized] ?? 0) + count;
}

function mergeCounts(left, right) {
  for (const [key, value] of Object.entries(right ?? {})) {
    increment(left, key, value);
  }
}

function classifyRule(item, fallback) {
  return (
    item?.rule_id ??
    item?.ruleId ??
    item?.rule ??
    item?.issue_type ??
    item?.issueType ??
    item?.kind ??
    fallback
  );
}

function classifySeverity(item) {
  return item?.severity ?? item?.level ?? item?.rule_severity ?? "unknown";
}

function collectIssueCounts(value, result = { issueCount: 0, byRule: {}, bySeverity: {}, rawCountFields: {} }) {
  if (!value || typeof value !== "object") {
    return result;
  }
  if (Array.isArray(value)) {
    for (const item of value) {
      collectIssueCounts(item, result);
    }
    return result;
  }
  for (const [key, nested] of Object.entries(value)) {
    const normalizedKey = normalizeKey(key);
    if (Array.isArray(nested) && issueArrayKeys.has(normalizedKey)) {
      increment(result.rawCountFields, key, nested.length);
      result.issueCount += nested.length;
      for (const item of nested) {
        increment(result.byRule, classifyRule(item, key));
        increment(result.bySeverity, classifySeverity(item));
      }
      continue;
    }
    collectIssueCounts(nested, result);
  }
  return result;
}

function writeText(file, content) {
  secureWriteFile(file, content ?? "");
}

function runFallow(reportRoot, name, args, outputFile) {
  const stdoutFile = path.join(reportRoot, `${name}.stdout.log`);
  const stderrFile = path.join(reportRoot, `${name}.stderr.log`);
  const child = spawnSync(process.execPath, [fallowScript, ...args], {
    cwd: repoRoot,
    env: {
      ...process.env,
      PATH: `${path.dirname(process.execPath)}:${process.env.PATH ?? ""}`,
    },
    encoding: "utf8",
    maxBuffer: 128 * 1024 * 1024,
  });
  writeText(stdoutFile, child.stdout ?? "");
  writeText(stderrFile, child.stderr ?? "");
  if (child.error) {
    throw child.error;
  }
  return {
    name,
    command: ["fallow", ...args],
    outputFile,
    stdoutFile,
    stderrFile,
    status: (child.status ?? 1) === 0 ? "pass" : "fail",
    exitCode: child.status ?? 1,
  };
}

function reportFromRun(run) {
  const data = existsSync(run.outputFile) ? readJSON(run.outputFile) : {};
  const counts = collectIssueCounts(data);
  return {
    name: run.name,
    command: run.command,
    status: run.status,
    exit_code: run.exitCode,
    artifact: artifactRef(`${run.name}_json`, repoRel(run.outputFile), "json"),
    issue_count: counts.issueCount,
    by_rule: counts.byRule,
    by_severity: counts.bySeverity,
    raw_count_fields: counts.rawCountFields,
  };
}

function hasUsableFallowOutput(run) {
  if (!existsSync(run.outputFile)) {
    return false;
  }
  if (!run.outputFile.endsWith(".json")) {
    return statSync(run.outputFile).size > 0;
  }
  try {
    const data = readJSON(run.outputFile);
    return data?.error !== true;
  } catch {
    return false;
  }
}

function failureReasonForRuns(runs) {
  const combined = runs.map((run) => {
    const stderr = existsSync(run.stderrFile) ? readFileSync(run.stderrFile, "utf8") : "";
    const stdout = existsSync(run.stdoutFile) ? readFileSync(run.stdoutFile, "utf8") : "";
    return `${run.name}\n${stdout}\n${stderr}`;
  }).join("\n").toLowerCase();
  if (combined.includes("config") || combined.includes(".fallowrc")) {
    return { failureClass: "config", failureReason: "configuration_error" };
  }
  return { failureClass: "harness", failureReason: "tool_diagnostic_failure" };
}

function makeToolSummary({ identity, status, startedAt, durationMs, summaryArtifacts, logArtifacts, failures = [], warnings = [], failureClass = null, failureReason = null }) {
  const targetRoot = path.join(identity.run_root, target);
  const summary = buildToolRunSummary({
    target,
    command: ["make", target],
    status,
    startedAt,
    completedAt: nowUTC(),
    durationMs,
    outputMode: normalizeOutputMode(),
    resultRoot: repoRel(identity.result_root),
    runId: identity.run_id,
    runRoot: repoRel(identity.run_root),
    summaryArtifacts,
    logArtifacts,
    counts: {
      non_test: 1,
      non_test_failed: status === "pass" ? 0 : 1,
    },
    failureClass,
    failureReason,
    failures,
    warnings,
    rerunCommands: [`make ${target}`],
  });
  validateSchemaSync("cartulary.tool_run_summary.v3", summary);
  secureWriteFile(toolSummaryPath(targetRoot), `${JSON.stringify(summary, null, 2)}\n`);
  return summary;
}

function main() {
  const startedAt = nowUTC();
  const startedMs = monotonicMs();
  const identity = resolveRetainedArtifactIdentity(target, process.env, {
    allowExistingRunRoot: true,
    materializeGeneratedRunId: true,
  });
  const targetRoot = path.join(identity.run_root, target);
  const reportRoot = path.join(targetRoot, "fallow");
  mkdir(reportRoot);

  const summaryArtifacts = [];
  const logArtifacts = [];
  const warnings = [];
  const runs = [];

  try {
    if (!existsSync(configPath)) {
      throw Object.assign(new Error(".fallowrc.json is required"), {
        failureClass: "config",
        failureReason: "configuration_error",
      });
    }
    if (!existsSync(fallowScript)) {
      throw Object.assign(new Error("fallow package binary was not found; run make frontend-install"), {
        failureClass: "config",
        failureReason: "configuration_error",
      });
    }

    const deadCodeJSON = path.join(reportRoot, "dead-code.json");
    const deadCodeSARIF = path.join(reportRoot, "dead-code.sarif");
    const deadCodeMarkdown = path.join(reportRoot, "dead-code.md");
    const dupesJSON = path.join(reportRoot, "dupes.json");
    const healthJSON = path.join(reportRoot, "health.json");

    runs.push(runFallow(reportRoot, "dead-code", [
      "dead-code",
      "--format",
      "json",
      "--quiet",
      "--no-cache",
      "--output-file",
      deadCodeJSON,
      "--sarif-file",
      deadCodeSARIF,
    ], deadCodeJSON));
    runs.push(runFallow(reportRoot, "dead-code-markdown", [
      "dead-code",
      "--format",
      "markdown",
      "--quiet",
      "--no-cache",
      "--output-file",
      deadCodeMarkdown,
    ], deadCodeMarkdown));
    runs.push(runFallow(reportRoot, "dupes", [
      "dupes",
      "--format",
      "json",
      "--quiet",
      "--no-cache",
      "--output-file",
      dupesJSON,
    ], dupesJSON));
    runs.push(runFallow(reportRoot, "health", [
      "health",
      "--format",
      "json",
      "--quiet",
      "--no-cache",
      "--output-file",
      healthJSON,
    ], healthJSON));

    for (const run of runs) {
      if (existsSync(run.stdoutFile) && statSync(run.stdoutFile).size > 0) {
        logArtifacts.push(artifactRef(`${run.name}_stdout`, repoRel(run.stdoutFile), "log"));
      }
      if (existsSync(run.stderrFile) && statSync(run.stderrFile).size > 0) {
        logArtifacts.push(artifactRef(`${run.name}_stderr`, repoRel(run.stderrFile), "log"));
      }
    }

    const failedRuns = runs.filter((run) => (
      run.status !== "pass" && !hasUsableFallowOutput(run)
    ));
    if (failedRuns.length > 0) {
      const failure = failureReasonForRuns(failedRuns);
      const failureRecords = failedRuns.map((run) => ({
        target,
        label: run.name,
        failure_class: failure.failureClass,
        failure_reason: failure.failureReason,
        headline: `${run.name} failed`,
        artifact: repoRel(run.stderrFile),
      }));
      makeToolSummary({
        identity,
        status: "fail",
        startedAt,
        durationMs: monotonicMs() - startedMs,
        summaryArtifacts,
        logArtifacts,
        failures: failureRecords,
        failureClass: failure.failureClass,
        failureReason: failure.failureReason,
      });
      return failure.failureReason === "configuration_error" ? 2 : 1;
    }

    const nonBlockingExitRuns = runs.filter((run) => (
      run.status !== "pass" && hasUsableFallowOutput(run)
    ));
    for (const run of nonBlockingExitRuns) {
      warnings.push({
        kind: "fallow_nonblocking_exit",
        message: `${run.name} exited ${run.exitCode}; Phase A retained the valid report without failing the target.`,
      });
    }

    const reports = runs
      .filter((run) => run.outputFile.endsWith(".json"))
      .map(reportFromRun);
    const totals = {
      reports: reports.length,
      issue_count: 0,
      by_rule: {},
      by_severity: {},
    };
    for (const report of reports) {
      totals.issue_count += report.issue_count;
      mergeCounts(totals.by_rule, report.by_rule);
      mergeCounts(totals.by_severity, report.by_severity);
    }
    if (totals.issue_count > 0) {
      warnings.push({
        kind: "fallow_findings",
        issue_count: totals.issue_count,
        message: "Fallow Phase A findings are retained as non-blocking static-analysis evidence.",
      });
    }

    const fallowArtifacts = [
      artifactRef("dead_code_json", repoRel(deadCodeJSON), "json"),
      artifactRef("dead_code_sarif", repoRel(deadCodeSARIF), "sarif"),
      artifactRef("dead_code_markdown", repoRel(deadCodeMarkdown), "markdown"),
      artifactRef("dupes_json", repoRel(dupesJSON), "json"),
      artifactRef("health_json", repoRel(healthJSON), "json"),
    ];
    const fallowSummary = {
      schema_id: fallowSummarySchemaID,
      target,
      generated_at: nowUTC(),
      mode: "phase_a_report",
      config: {
        path: ".fallowrc.json",
        static_layer: "open",
        runtime_enabled: false,
      },
      reports,
      totals,
      baseline: {
        mode: "not_configured",
        artifacts: [],
      },
      enforcement: {
        blocking: false,
        failure_on_issues: false,
      },
      artifacts: fallowArtifacts,
      warnings,
      extensions: {},
    };
    validateSchemaSync(fallowSummarySchemaID, fallowSummary);
    const fallowSummaryFile = path.join(targetRoot, "fallow-static-summary.json");
    secureWriteFile(fallowSummaryFile, `${JSON.stringify(fallowSummary, null, 2)}\n`);
    summaryArtifacts.push(
      artifactRef("fallow_static_summary", repoRel(fallowSummaryFile), "json"),
      ...fallowArtifacts,
    );

    makeToolSummary({
      identity,
      status: "pass",
      startedAt,
      durationMs: monotonicMs() - startedMs,
      summaryArtifacts,
      logArtifacts,
      warnings,
    });
    return 0;
  } catch (error) {
    const failureClass = error.failureClass ?? "artifact";
    const failureReason = error.failureReason ?? "artifact_error";
    makeToolSummary({
      identity,
      status: "fail",
      startedAt,
      durationMs: monotonicMs() - startedMs,
      summaryArtifacts,
      logArtifacts,
      failures: [
        {
          target,
          label: target,
          failure_class: failureClass,
          failure_reason: failureReason,
          headline: error.message,
          artifact: "",
        },
      ],
      failureClass,
      failureReason,
    });
    return failureReason === "configuration_error" ? 2 : 11;
  }
}

process.exit(main());
