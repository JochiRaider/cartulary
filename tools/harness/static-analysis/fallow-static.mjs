import { spawn } from "../workspace/child-process.mjs";
import {
  closeSync,
  existsSync,
  mkdirSync,
  openSync,
  readFileSync,
  statSync,
} from "node:fs";
import path from "node:path";
import {
  collectIssueCounts, summarizeFallowReports, runFallowBatch, validateFallowReport,
  readPackageSurfaceBindings, collectBlockingPackageSurfaceFindings,
  collectBlockingWebSurfaceFindings,
} from "./fallow-policy.mjs";

import {
  resolveRetainedArtifactIdentity,
  secureWriteFile,
  validateSchemaSync,
} from "../contract/index.mjs";
import {
  fileArtifactRef,
  buildToolRunSummary,
  normalizeOutputMode,
  toolSummaryPath,
} from "../output/index.mjs";
import {
  buildBlockingPackageSurfaceConfig,
  buildResolvedFallowConfig,
  defaultFallowReachabilityOwnerPath,
} from "./fallow-reachability.mjs";

const target = "frontend-fallow-static";
const fallowSummarySchemaID = "cartulary.fallow_static_summary.v2";
function nowUTC() {
  return new Date().toISOString();
}

function monotonicMs() {
  return Number(process.hrtime.bigint() / 1_000_000n);
}

function repoRel(root, file) {
  const relative = path.relative(root, file).replaceAll("\\", "/");
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

function fallowArtifactRef(role, file, kind = "json") {
  return {
    role,
    kind,
    path: file,
  };
}

function runFallow({ root, env }, reportRoot, name, args, outputFile) {
  const stdoutFile = path.join(reportRoot, `${name}.stdout.log`);
  const stderrFile = path.join(reportRoot, `${name}.stderr.log`);
  return new Promise((resolve) => {
    const stdoutFD = openSync(stdoutFile, "w", 0o600);
    let stderrFD;
    try {
      stderrFD = openSync(stderrFile, "w", 0o600);
    } catch (error) {
      closeSync(stdoutFD);
      throw error;
    }
    let settled = false;
    const finish = (status) => {
      if (settled) return;
      settled = true;
      closeSync(stdoutFD);
      closeSync(stderrFD);
      const exitCode = status ?? 1;
      resolve({
        name,
        command: ["fallow", ...args],
        outputFile,
        stdoutFile,
        stderrFile,
        status: exitCode === 0 ? "pass" : "fail",
        exitCode,
      });
    };
    try {
      const child = spawn(process.execPath, [path.join(root, "node_modules/fallow/bin/fallow"), ...args], {
        cwd: root,
        env: {
          ...env,
          PATH: `${path.dirname(process.execPath)}:${env.PATH ?? ""}`,
        },
        stdio: ["ignore", stdoutFD, stderrFD],
      });
      child.on("error", () => finish(127));
      child.on("close", (status, signal) => finish(signal ? 128 : status));
    } catch {
      finish(127);
    }
  });
}

function reportFromRun(root, run, data) {
  const counts = collectIssueCounts(data);
  const artifactRole = `${run.name.replaceAll("-", "_")}_json`;
  return {
    name: run.name,
    command: run.command,
    status: run.status,
    exit_code: run.exitCode,
    artifact: fallowArtifactRef(artifactRole, repoRel(root, run.outputFile), "json"),
    issue_count: counts.issueCount,
    by_rule: counts.byRule,
    by_severity: counts.bySeverity,
    raw_count_fields: counts.rawCountFields,
  };
}

function makeToolSummary({ root, env, identity, status, startedAt, durationMs, summaryArtifacts, logArtifacts, failures = [], warnings = [] }) {
  const targetRoot = path.join(identity.run_root, target);
  const summary = buildToolRunSummary({
    target,
    command: ["make", target],
    status,
    startedAt,
    completedAt: nowUTC(),
    durationMs,
    outputMode: normalizeOutputMode(env),
    resultRoot: repoRel(root, identity.result_root),
    runId: identity.run_id,
    runRoot: repoRel(root, identity.run_root),
    summaryArtifacts,
    logArtifacts,
    counts: {
      non_test: 1,
      non_test_failed: status === "pass" ? 0 : 1,
    },
    failures,
    warnings,
    rerunCommands: [`make ${target}`],
  });
  validateSchemaSync("cartulary.tool_run_summary.v5", summary);
  secureWriteFile(toolSummaryPath(targetRoot), `${JSON.stringify(summary, null, 2)}\n`);
  return summary;
}

export async function runFallowStatic({
  root,
  env = process.env,
  execute = (...args) => runFallow({ root, env }, ...args),
}) {
  const fallowScript = path.join(root, "node_modules", "fallow", "bin", "fallow");
  const configPath = path.join(root, ".fallowrc.json");
  const reachabilityOwnerPath = path.join(root, defaultFallowReachabilityOwnerPath);

  const startedAt = nowUTC();
  const startedMs = monotonicMs();
  const identity = resolveRetainedArtifactIdentity(target, env, {
    root,
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
  const failures = [];

  try {
    if (!existsSync(configPath)) {
      throw Object.assign(new Error(".fallowrc.json is required"), {
        failureClass: "config",
        failureReason: "configuration_error",
      });
    }
    if (!existsSync(reachabilityOwnerPath)) {
      throw Object.assign(new Error(`${defaultFallowReachabilityOwnerPath} is required`), {
        failureClass: "config",
        failureReason: "configuration_error",
      });
    }
    if (!existsSync(fallowScript)) {
      throw Object.assign(new Error("fallow package binary was not found; run make bootstrap PART=frontend"), {
        failureClass: "config",
        failureReason: "configuration_error",
      });
    }

    const resolvedConfig = path.join(reportRoot, "resolved-fallowrc.json");
    let reachability;
    let bindings;
    try {
      reachability = buildResolvedFallowConfig({
        root,
        outputFile: resolvedConfig,
      });
      bindings = readPackageSurfaceBindings(root, reachability.owner.blocking_package_surfaces);
    } catch (error) {
      throw Object.assign(error, {
        failureClass: "config",
        failureReason: "configuration_error",
      });
    }
    const resolvedConfigRel = repoRel(root, resolvedConfig);
    const packageSurfaceConfig = path.join(
      reportRoot,
      "package-surface-fallowrc.json",
    );
    secureWriteFile(
      packageSurfaceConfig,
      `${JSON.stringify(
        buildBlockingPackageSurfaceConfig({
          config: reachability.config,
          owner: reachability.owner,
        }),
        null,
        2,
      )}\n`,
    );
    const packageSurfaceConfigRel = repoRel(root, packageSurfaceConfig);

    const deadCodeJSON = path.join(reportRoot, "dead-code.json");
    const deadCodeSARIF = path.join(reportRoot, "dead-code.sarif");
    const deadCodeMarkdown = path.join(reportRoot, "dead-code.md");
    const dupesJSON = path.join(reportRoot, "dupes.json");
    const healthJSON = path.join(reportRoot, "health.json");
    const packageSurfaceJSON = path.join(
      reportRoot,
      "package-surface-dead-code.json",
    );

    const artifactFiles = [
      ["resolved_fallow_config", resolvedConfig, "json"],
      ["dead_code_json", deadCodeJSON, "json"],
      ["dead_code_sarif", deadCodeSARIF, "sarif"],
      ["dead_code_markdown", deadCodeMarkdown, "markdown"],
      ["dupes_json", dupesJSON, "json"],
      ["health_json", healthJSON, "json"],
      ["package_surface_config", packageSurfaceConfig, "json"],
      ["package_surface_dead_code", packageSurfaceJSON, "json"],
    ];
    runs.push(...await runFallowBatch([
      {
        reportRoot,
        name: "dead-code",
        args: [
          "dead-code",
          "--config",
          resolvedConfigRel,
          "--format",
          "json",
          "--quiet",
          "--no-cache",
          "--output-file",
          deadCodeJSON,
          "--sarif-file",
          deadCodeSARIF,
        ],
        outputFile: deadCodeJSON,
      },
      {
        reportRoot,
        name: "dupes",
        args: [
          "dupes",
          "--config",
          resolvedConfigRel,
          "--format",
          "json",
          "--quiet",
          "--no-cache",
          "--output-file",
          dupesJSON,
        ],
        outputFile: dupesJSON,
      },
      {
        reportRoot,
        name: "health",
        args: [
          "health",
          "--config",
          resolvedConfigRel,
          "--format",
          "json",
          "--quiet",
          "--no-cache",
          "--output-file",
          healthJSON,
        ],
        outputFile: healthJSON,
      },
      {
        reportRoot,
        name: "package-surface-dead-code",
        args: [
          "dead-code",
          "--config",
          packageSurfaceConfigRel,
          "--format",
          "json",
          "--quiet",
          "--no-cache",
          "--output-file",
          packageSurfaceJSON,
        ],
        outputFile: packageSurfaceJSON,
      },
    ], execute));

    for (const run of runs) {
      if (existsSync(run.stdoutFile) && statSync(run.stdoutFile).size > 0) {
        logArtifacts.push(fileArtifactRef(`${run.name}_stdout`, repoRel(root, run.stdoutFile), "log"));
      }
      if (existsSync(run.stderrFile) && statSync(run.stderrFile).size > 0) {
        logArtifacts.push(fileArtifactRef(`${run.name}_stderr`, repoRel(root, run.stderrFile), "log"));
      }
    }

    for (const [role, file, kind] of artifactFiles) {
      if (existsSync(file)) summaryArtifacts.push(fileArtifactRef(role, repoRel(root, file), kind));
    }
    const reports = [];
    let packageSurfaceFindings = [];
    let webSurfaceFindings = [];
    let deadCodeReport;
    let completeReports = 0;
    for (const [index, run] of runs.entries()) {
      const record = (failureClass, failureReason, headline, artifact = run.outputFile, label = run.name) => ({
        target, label, headline, failure_class: failureClass, failure_reason: failureReason,
        child_registry_order: index, artifact: repoRel(root, artifact),
      });
      if (run.exitCode !== 0 && run.exitCode !== 1) {
        failures.push(record(run.exitCode === 2 ? "config" : "harness",
          run.exitCode === 2 ? "configuration_error" : "tool_diagnostic_failure",
          `${run.name} exited ${run.exitCode}`, run.stderrFile));
        continue;
      }
      try {
        const data = readJSON(run.outputFile);
        validateFallowReport(data, run.name === "package-surface-dead-code" ? "dead-code" : run.name);
        if (run.name === "dead-code") {
          deadCodeReport = data;
          webSurfaceFindings = collectBlockingWebSurfaceFindings(data, reachability.owner.blocking_web_surface);
          failures.push(...webSurfaceFindings.map((finding) => record("harness", "tool_diagnostic_failure",
            `${finding.rule}: ${finding.path}${finding.symbol ? ` (${finding.symbol})` : ""}`,
            run.outputFile, `${finding.path}:${finding.symbol ?? finding.rule}`)));
        } else if (run.name === "package-surface-dead-code") {
          packageSurfaceFindings = collectBlockingPackageSurfaceFindings(data, reachability.owner.blocking_package_surfaces, bindings);
          failures.push(...packageSurfaceFindings.map((finding) => record("harness", "tool_diagnostic_failure",
            `unused manual package export ${finding.symbol}`, run.outputFile, `${finding.package_name}:${finding.symbol}`)));
        }
        if (run.name !== "package-surface-dead-code") reports.push(reportFromRun(root, run, data));
        completeReports += 1;
        if (run.exitCode === 1) warnings.push({
          kind: "fallow_nonblocking_exit",
          message: `${run.name} exited 1; the retained report is evaluated by the bounded static policies.`,
        });
      } catch (error) {
        failures.push(record("artifact", "artifact_error", error.message));
      }
    }
    if (deadCodeReport) {
      try {
        const sarif = readJSON(deadCodeSARIF);
        if (sarif?.version !== "2.1.0" || !Array.isArray(sarif.runs)) {
          throw new Error("Fallow dead-code SARIF report is invalid");
        }
      } catch (error) {
        failures.push({ target, label: "dead-code-sarif", headline: error.message,
          failure_class: "artifact", failure_reason: "artifact_error", child_registry_order: 0,
          artifact: repoRel(root, deadCodeSARIF) });
      }
    }
    if (completeReports !== runs.length || failures.some((failure) => failure.failure_reason === "artifact_error")) {
      return makeToolSummary({ root, env, identity, status: "fail", startedAt,
        durationMs: monotonicMs() - startedMs, summaryArtifacts, logArtifacts, failures, warnings,
      }).exit_code;
    }

    // Render the human summary from the same result; never scan a second time
    // or consume Markdown as verification input.
    const deadCodeCounts = collectIssueCounts(deadCodeReport);
    secureWriteFile(deadCodeMarkdown, [
      "# Fallow static reachability", "",
      "Detailed findings and actions are retained in dead-code.json.", "",
      "| Rule | Findings |", "| --- | --- |",
      ...Object.entries(deadCodeCounts.byRule).sort(([a], [b]) => a.localeCompare(b))
        .map(([rule, count]) => `| ${rule} | ${count} |`), "",
    ].join("\n"));
    const totals = summarizeFallowReports(reports);
    if (totals.issue_count > 0) {
      warnings.push({
        kind: "fallow_findings",
        issue_count: totals.issue_count,
        message: "Findings outside the bounded web and package policies remain advisory static-analysis evidence.",
      });
    }

    const fallowArtifacts = artifactFiles.map(([role, artifactPath, kind]) =>
      fallowArtifactRef(role, repoRel(root, artifactPath), kind),
    );
    const fallowSummary = {
      schema_id: fallowSummarySchemaID,
      target,
      generated_at: nowUTC(),
      mode: "static_reachability_report",
      config: {
        path: repoRel(root, resolvedConfig),
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
        blocking: true,
        failure_on_issues: true,
      },
      artifacts: fallowArtifacts,
      warnings,
      extensions: {
        "cartulary.fallow_reachability": {
          owner_path: defaultFallowReachabilityOwnerPath,
          base_config_path: ".fallowrc.json",
          resolved_config_path: repoRel(root, resolvedConfig),
          stats: reachability.stats,
          web_surface: {
            ...reachability.owner.blocking_web_surface,
            finding_count: webSurfaceFindings.length,
            findings: webSurfaceFindings,
            report_path: repoRel(root, deadCodeJSON),
          },
          package_surface: {
            blocking: true,
            failure_on_issues: true,
            finding_count: packageSurfaceFindings.length,
            findings: packageSurfaceFindings,
            packages: reachability.owner.blocking_package_surfaces.packages.map(
              (entry) => ({
                entrypoint: entry.entrypoint,
                package_name: entry.package_name,
              }),
            ),
            report_path: repoRel(root, packageSurfaceJSON),
          },
        },
      },
    };
    validateSchemaSync(fallowSummarySchemaID, fallowSummary);
    const fallowSummaryFile = path.join(targetRoot, "fallow-static-summary.json");
    secureWriteFile(fallowSummaryFile, `${JSON.stringify(fallowSummary, null, 2)}\n`);
    summaryArtifacts.push(
      fileArtifactRef("fallow_static_summary", repoRel(root, fallowSummaryFile), "json"),
      fileArtifactRef("dead_code_markdown", repoRel(root, deadCodeMarkdown), "markdown"),
    );
    return makeToolSummary({ root, env, identity, status: failures.length ? "fail" : "pass",
      startedAt, durationMs: monotonicMs() - startedMs, summaryArtifacts, logArtifacts, failures, warnings,
    }).exit_code;
  } catch (error) {
    const failureClass = error.failureClass ?? "artifact";
    const failureReason = error.failureReason ?? "artifact_error";
    return makeToolSummary({
        root, env,
      identity,
      status: "fail",
      startedAt,
      durationMs: monotonicMs() - startedMs,
      summaryArtifacts,
      logArtifacts,
      failures: [
        ...failures,
        {
          target,
          label: target,
          failure_class: failureClass,
          failure_reason: failureReason,
          headline: error.message,
          artifact: "",
        },
      ],
      warnings,
    }).exit_code;
  }
}

