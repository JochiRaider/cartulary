#!/usr/bin/env node
import { repoRoot } from "../../contract/index.mjs";

import {
  existsSync,
  readFileSync,
} from "node:fs";
import path from "node:path";
import {
  failureFieldsForJSON,
  failuresFromDossiers,
  manifestMismatchFailureRecord,
} from "../../contract/failure-taxonomy.mjs";
import {
  compactJSONString,
  HarnessConfigError,
  prettyJSONString,
  secureMkdir,
  secureWriteFile,
  validateSchemaSync,
} from "../../contract/harness-contract.mjs";
import {
  artifactLine,
  fileArtifactRef,
  buildToolRunSummary,
  machineOutput,
  normalizeOutputMode,
  resultLine,
  suppressChildSuccess,
  terminalArtifactPath,
  toolSummaryPath,
  verboseOutput,
} from "../tool-output.mjs";
import {
  phaseSummarySchemaID,
  resolveResultsRoot,
  resolveRunId,
  timingBucketSet,
  validPhaseCountingModes,
} from "../../contract/test-output-context.mjs";
import { loadGovulncheckFindingsFile } from "./security-diagnostics.mjs";

const resultsRoot = resolveResultsRoot();

const runId = resolveRunId();

function normalizePath(value) {
  return value.replaceAll("\\", "/");
}

function relToRepo(value) {
  if (!value) {
    return "";
  }
  const normalized = normalizePath(value);
  if (!path.isAbsolute(value)) {
    return normalized;
  }
  const relative = normalizePath(path.relative(repoRoot, value));
  if (!relative.startsWith("../") && relative !== "..") {
    return relative;
  }
  return normalized;
}

function ensureDir(dir) {
  secureMkdir(dir);
}

function writeJson(file, value) {
  ensureDir(path.dirname(file));
  secureWriteFile(file, prettyJSONString(value));
}

function writeValidatedJson(file, schemaID, value) {
  validateSchemaSync(schemaID, value);
  writeJson(file, value);
}

function clampDurationMs(value) {
  if (!Number.isFinite(value) || value < 0) {
    return 0;
  }
  return value;
}

function normalizeAccountingMode(value) {
  if (value === "actual" || value === "reused" || value === "derived") {
    return value;
  }
  return "actual";
}

function normalizePhaseCountingMode(value) {
  if (validPhaseCountingModes.has(value)) {
    return value;
  }
  return "counted";
}

function resolveOutputMode() {
  return normalizeOutputMode();
}

function normalizeTimingBucket(value, runner = "") {
  if (value && timingBucketSet.has(value)) {
    return value;
  }
  if (runner === "go_test" || runner === "vitest" || runner === "playwright") {
    return "test_command";
  }
  return "test_command";
}

function requiredEnv(name) {
  const value = process.env[name];
  if (value === undefined || value === "") {
    throw new Error(`missing required environment variable ${name}`);
  }
  return value;
}

function optionalEnv(name, fallback = "") {
  return process.env[name] ?? fallback;
}

function parseInteger(name, fallback = 0) {
  const value = process.env[name];
  if (value === undefined || value === "") {
    return fallback;
  }
  const parsed = Number.parseInt(value, 10);
  if (Number.isNaN(parsed)) {
    throw new Error(`invalid integer ${name}=${value}`);
  }
  return parsed;
}

function optionalJSONEnv(name, fallback = null) {
  const value = optionalEnv(name);
  if (!value) {
    return fallback;
  }
  try {
    return JSON.parse(value);
  } catch {
    return fallback;
  }
}

function resolveArtifactPath(value) {
  if (!value) {
    return "";
  }
  return path.isAbsolute(value) ? value : path.join(repoRoot, value);
}

function finalizeLine(summary) {
  const actions = Array.isArray(summary.actions) ? summary.actions : [];
  const reused = actions.filter((action) => action.execution_state === "reused").length;
  const cacheHits = actions.filter((action) => action.cache?.state === "hit").length;
  return `${[
    `[FINALIZE] generated=${summary.generated?.status ?? "unknown"}`,
    `files=${summary.generated?.updated_file_count ?? 0}`,
    `duration=${summary.duration?.status ?? "skipped"}`,
    `run_checks=${summary.run_checks?.status ?? "skipped"}`,
    `reused=${reused}`,
    `cache_hits=${cacheHits}`,
    `results_dir=${summary.results_dir ?? "-"}`,
  ].join(" ")}\n`;
}

export function createBasePhaseContext(runner) {
  const phaseDir = requiredEnv("CARTULARY_PHASE_DIR");
  ensureDir(phaseDir);
  const accountingMode =
    optionalEnv("CARTULARY_REPORT_SLICE") === "1"
      ? normalizeAccountingMode(
          optionalEnv("CARTULARY_PHASE_ACCOUNTING_MODE", "actual"),
        )
      : "actual";
  const countingMode = normalizePhaseCountingMode(
    optionalEnv("CARTULARY_PHASE_COUNTING_MODE", "counted"),
  );
  const logicalDurationMs = clampDurationMs(
    parseInteger("CARTULARY_PHASE_LOGICAL_DURATION_MS", 0),
  );
  const executedDurationMs = clampDurationMs(
    parseInteger(
      "CARTULARY_PHASE_EXECUTED_DURATION_MS",
      accountingMode === "actual" ? logicalDurationMs : 0,
    ),
  );
  return {
    label: requiredEnv("CARTULARY_PHASE_LABEL"),
    phaseDir,
    target: optionalEnv("CARTULARY_TEST_TARGET", "adhoc"),
    command: requiredEnv("CARTULARY_PHASE_COMMAND"),
    commandArgv: optionalJSONEnv("CARTULARY_PHASE_COMMAND_ARGV", null),
    runner,
    timingBucket: normalizeTimingBucket(
      optionalEnv("CARTULARY_PHASE_TIMING_BUCKET"),
      runner,
    ),
    startTime: requiredEnv("CARTULARY_PHASE_START_TIME"),
    endTime: requiredEnv("CARTULARY_PHASE_END_TIME"),
    accountingMode,
    countingMode,
    executedDurationMs,
    logicalDurationMs,
    reusedDurationMs: accountingMode === "reused" ? logicalDurationMs : 0,
    derivedDurationMs: accountingMode === "derived" ? logicalDurationMs : 0,
    wallDurationMs: clampDurationMs(
      parseInteger("CARTULARY_PHASE_WALL_DURATION_MS", logicalDurationMs),
    ),
    exitStatus: parseInteger("CARTULARY_PHASE_EXIT_STATUS", 0),
  };
}

export function writePhaseArtifacts(context, details) {
  const playwrightTimingPath = details.playwrightTiming
    ? path.join(context.phaseDir, "playwright-timing.json")
    : "";
  if (details.playwrightTiming) {
    writeJson(playwrightTimingPath, details.playwrightTiming);
  }

  const artifacts = {};
  for (const [key, value] of Object.entries({
    ...(details.artifacts ?? {}),
    shellcheck_inventory: existsSync(
      path.join(context.phaseDir, "shellcheck-inventory.txt"),
    )
      ? path.join(context.phaseDir, "shellcheck-inventory.txt")
      : "",
    security_profiles: existsSync(
      path.join(context.phaseDir, "security-profiles.jsonl"),
    )
      ? path.join(context.phaseDir, "security-profiles.jsonl")
      : "",
    govulncheck_findings: existsSync(
      path.join(context.phaseDir, "govulncheck-findings.json"),
    )
      ? path.join(context.phaseDir, "govulncheck-findings.json")
      : "",
    playwright_timing: playwrightTimingPath,
  })) {
    if (!value) {
      continue;
    }
    artifacts[key] = relToRepo(value);
  }
  const targetRunRoot =
    context.target === "adhoc"
      ? context.phaseDir
      : path.dirname(context.phaseDir);
  const finalizeSummaryPath = path.join(targetRunRoot, "finalize-summary.json");
  const finalizeSummaryRel = existsSync(finalizeSummaryPath)
    ? relToRepo(finalizeSummaryPath)
    : "";
  const finalizeSummary = finalizeSummaryRel
    ? readJsonIfExists(finalizeSummaryPath)
    : null;
  const existingTargetSummary = readJsonIfExists(
    path.join(targetRunRoot, "target-summary.json"),
  );
  const existingTargetExtensionsRaw =
    existingTargetSummary && typeof existingTargetSummary.extensions === "object"
      ? existingTargetSummary.extensions
      : {};
  if (Object.hasOwn(existingTargetExtensionsRaw, "cartulary.frontend_row_accounting")) {
    throw new HarnessConfigError(
      'current target summary contains unsupported extensions["cartulary.frontend_row_accounting"]',
      { reason: "artifact_error" },
    );
  }
  const existingTargetExtensions = existingTargetExtensionsRaw;
  const failureRecords = [
    ...failuresFromDossiers(details.dossiers ?? [], {
      target: context.target,
      label: context.label,
      command: context.command,
      runner: context.runner,
      phase: details.phase,
    }),
    ...(details.manifestMismatch
      ? [
          manifestMismatchFailureRecord(details.manifestMismatch, {
            target: context.target,
            phase: details.phase,
            runner: context.runner,
          }),
        ]
      : []),
    ...(finalizeSummary?.failures ?? []).map((failure) => ({
      failure_class: failure.failure_class,
      failure_reason: failure.failure_reason,
      target: context.target,
      child_target: failure.target ?? undefined,
      label: failure.substep_id ?? failure.action_id,
      headline: failure.headline,
      reproduce: context.command,
      artifacts: failure.summary_json ? [failure.summary_json] : [],
    })),
    ...(details.failures ?? []),
  ];
  const failureFields = failureFieldsForJSON(
    failureRecords,
    details.counts ?? {},
  );

  const meta = {
    label: context.label,
    runner: context.runner,
    command: context.command,
    start_time: context.startTime,
    end_time: context.endTime,
    exit_status: context.exitStatus,
    counting_mode: context.countingMode,
    accounting_mode: context.accountingMode,
    executed_duration_ms: context.executedDurationMs,
    logical_duration_ms: context.logicalDurationMs,
    reused_duration_ms: context.reusedDurationMs,
    derived_duration_ms: context.derivedDurationMs,
    wall_duration_ms: context.wallDurationMs,
    critical_path_wall_duration_ms: context.wallDurationMs,
    teardown_duration_ms:
      context.timingBucket === "teardown" ? context.wallDurationMs : 0,
    timing_bucket: context.timingBucket,
    status: details.status,
    counts: details.counts,
    failure_class: failureFields.failure_class,
    failure_classes: failureFields.failure_classes,
    failure_headline: failureFields.failure_headline,
  };

  writeJson(path.join(context.phaseDir, "meta.json"), meta);

  if (details.manifestSummary) {
    writeJson(
      path.join(context.phaseDir, "manifest-summary.json"),
      details.manifestSummary,
    );
  }
  if (details.manifestMismatch) {
    writeJson(
      path.join(context.phaseDir, "manifest-mismatch.json"),
      details.manifestMismatch,
    );
  }

  const summary = {
    schema_id: phaseSummarySchemaID,
    label: context.label,
    target: context.target,
    runner: context.runner,
    status: details.status,
    phase: details.phase,
    command: context.command,
    start_time: context.startTime,
    end_time: context.endTime,
    accounting_mode: context.accountingMode,
    executed_duration_ms: context.executedDurationMs,
    logical_duration_ms: context.logicalDurationMs,
    reused_duration_ms: context.reusedDurationMs,
    derived_duration_ms: context.derivedDurationMs,
    wall_duration_ms: context.wallDurationMs,
    critical_path_wall_duration_ms: context.wallDurationMs,
    teardown_duration_ms:
      context.timingBucket === "teardown" ? context.wallDurationMs : 0,
    timing_bucket: context.timingBucket,
    exit_status: context.exitStatus,
    counting_mode: context.countingMode,
    artifacts,
    counts: details.counts,
    ...failureFields,
    owners: details.owners ?? [],
    inventory: details.inventory ?? [],
    dossiers: details.dossiers ?? [],
    manifest_mismatch: details.manifestMismatch ?? null,
  };
  writeValidatedJson(
    path.join(context.phaseDir, "phase-summary.json"),
    phaseSummarySchemaID,
    summary,
  );
  const runRootAbs = path.join(resultsRoot, runId);
  const runRoot = relToRepo(runRootAbs);
  const toolSummaryFile = toolSummaryPath(targetRunRoot);
  const toolSummaryRel = relToRepo(toolSummaryFile);
  const shellcheckInventoryPath = artifacts.shellcheck_inventory
    ? resolveArtifactPath(artifacts.shellcheck_inventory)
    : "";
  const securityProfilesPath = artifacts.security_profiles
    ? resolveArtifactPath(artifacts.security_profiles)
    : "";
  const govulncheckFindingsPath = artifacts.govulncheck_findings
    ? resolveArtifactPath(artifacts.govulncheck_findings)
    : "";
  const govulncheckFindingsResult = loadGovulncheckFindingsFile(
    govulncheckFindingsPath,
  );
  const shellcheckFilesChecked =
    shellcheckInventoryPath && existsSync(shellcheckInventoryPath)
      ? readFileSync(shellcheckInventoryPath, "utf8")
          .split(/\r?\n/u)
          .filter(Boolean).length
      : null;
  const securityProfileCount =
    securityProfilesPath && existsSync(securityProfilesPath)
      ? readFileSync(securityProfilesPath, "utf8")
          .split(/\r?\n/u)
          .filter(Boolean).length
      : null;
  const govulncheckFindings = govulncheckFindingsResult.findings;
  const securityExtension = {};
  if (securityProfileCount !== null) {
    securityExtension.profile_count = securityProfileCount;
  }
  if (govulncheckFindings) {
    securityExtension.govulncheck = {
      status: govulncheckFindings.status ?? "",
      finding_count: govulncheckFindings.counts?.finding_count ?? 0,
      blocking_count: govulncheckFindings.counts?.blocking_count ?? 0,
      blocking_vulnerability_ids:
        govulncheckFindings.blocking_vulnerability_ids ?? [],
    };
  }
  const toolSummary = buildToolRunSummary({
    target: context.target,
    command:
      Array.isArray(context.commandArgv) && context.commandArgv.length > 0
        ? context.commandArgv
        : context.command,
    status: details.status,
    exitCode: details.status === "pass" ? 0 : context.exitStatus || 1,
    startedAt: context.startTime,
    completedAt: context.endTime,
    durationMs: context.wallDurationMs,
    outputMode: resolveOutputMode(),
    resultRoot: relToRepo(resultsRoot),
    runId,
    runRoot,
    summaryArtifacts: [
      fileArtifactRef("tool_run_summary", toolSummaryRel),
      fileArtifactRef(
        "phase_summary",
        relToRepo(path.join(context.phaseDir, "phase-summary.json")),
      ),
      existingTargetSummary?.artifacts?.frontend_row_accounting
        ? fileArtifactRef(
            "frontend_row_accounting",
            existingTargetSummary.artifacts.frontend_row_accounting,
          )
        : null,
      details.manifestSummary
        ? fileArtifactRef(
            "manifest_summary",
            relToRepo(path.join(context.phaseDir, "manifest-summary.json")),
          )
        : null,
      details.manifestMismatch
        ? fileArtifactRef(
            "manifest_mismatch",
            relToRepo(path.join(context.phaseDir, "manifest-mismatch.json")),
          )
        : null,
      finalizeSummaryRel
        ? fileArtifactRef("finalize_summary", finalizeSummaryRel)
        : null,
      artifacts.shellcheck_inventory
        ? fileArtifactRef(
            "shellcheck_inventory",
            artifacts.shellcheck_inventory,
            "text",
          )
        : null,
      artifacts.security_profiles
        ? fileArtifactRef("security_profiles", artifacts.security_profiles, "jsonl")
        : null,
      artifacts.govulncheck_findings
        ? fileArtifactRef(
            "govulncheck_findings",
            artifacts.govulncheck_findings,
            "json",
          )
        : null,
      artifacts.release_readiness_evidence
        ? fileArtifactRef(
            "release_readiness_evidence",
            artifacts.release_readiness_evidence,
            "json",
          )
        : null,
    ],
    logArtifacts: Object.entries(artifacts)
      .filter(([key]) => key.endsWith("_log"))
      .map(([key, value]) => fileArtifactRef(key, value, "log")),
    workUnits: [],
    evidenceTargets: [
      {
        target: context.target,
        status: details.status,
        run_root: runRoot,
      },
    ],
    helperUnits: [],
    counts: details.counts ?? {},
    failureClass: failureFields.failure_class,
    failures: failureRecords,
    slowest: [],
    rerunCommands: context.command ? [context.command] : [],
    extensions: {
      ...existingTargetExtensions,
      ...(shellcheckFilesChecked !== null
        ? { "cartulary.lint_shell": { files_checked: shellcheckFilesChecked } }
        : {}),
      ...(Object.keys(securityExtension).length > 0
        ? { "cartulary.security": securityExtension }
        : {}),
    },
  });
  writeToolSummary(toolSummaryFile, toolSummary);

  if (details.status !== "pass" && !suppressChildSuccess()) {
    if (machineOutput()) {
      process.stdout.write(compactJSONString(toolSummary));
    } else if (!verboseOutput()) {
      const logArtifact =
        toolSummary.log_artifacts?.find(
          (artifact) => artifact.role === "stderr_log",
        )?.path ??
        toolSummary.log_artifacts?.[0]?.path ??
        "-";
      const failure = failureRecords[0] ?? {};
      process.stderr.write(
        `[FAIL] target=${context.target} exit_code=${toolSummary.exit_code} failure_class=${toolSummary.failure_class ?? "harness"} reason=${toolSummary.failure_reason ?? "unknown_failure"} work_unit=- child_target=- duration_ms=${toolSummary.duration_ms} headline="${failure.headline ?? `${context.label} failed`}"\n`,
      );
      process.stderr.write(
        `[ARTIFACTS] target=${context.target} root=${toolSummary.run_root} summary_json=${terminalArtifactPath(toolSummary.run_root, toolSummaryRel)} log_artifact=${terminalArtifactPath(toolSummary.run_root, logArtifact)} scheduler_json=- progress_log=-${finalizeSummaryRel ? ` finalize_json=${terminalArtifactPath(toolSummary.run_root, finalizeSummaryRel)}` : ""}\n`,
      );
      process.stderr.write(`[RERUN] command="${context.command}"\n`);
      process.stderr.write(
        `[INVESTIGATE] command="make explain-run RESULTS_DIR=${relToRepo(path.join(resultsRoot, runId))} TARGET=${context.target} DETAIL=logs"\n`,
      );
    }
  }

  if (details.status === "pass" && !suppressChildSuccess()) {
    if (machineOutput()) {
      process.stdout.write(compactJSONString(toolSummary));
    } else {
      if (finalizeSummary) {
        process.stdout.write(finalizeLine(finalizeSummary));
      }
      process.stdout.write(resultLine(toolSummary, toolSummaryRel));
      process.stdout.write(
        artifactLine(toolSummary, toolSummaryRel, {
          extraFields: finalizeSummaryRel
            ? [
                `finalize_json=${terminalArtifactPath(toolSummary.run_root, finalizeSummaryRel)}`,
              ]
            : [],
          investigate: `make explain-run RESULTS_DIR=${relToRepo(path.join(resultsRoot, runId))} TARGET=${context.target}`,
        }),
      );
    }
  }
}

function readJsonIfExists(file) {
  if (!existsSync(file)) {
    return null;
  }
  try {
    return JSON.parse(readFileSync(file, "utf8"));
  } catch {
    return null;
  }
}

function writeToolSummary(file, summary) {
  writeValidatedJson(file, summary.schema_id, summary);
  return relToRepo(file);
}

function govulncheckFindingsPath(context) {
  const commandContext = [context.target, context.label, context.command]
    .join("\n")
    .toLowerCase();
  if (
    !commandContext.includes("go-vulncheck") &&
    !commandContext.includes("govulncheck")
  ) {
    return "";
  }
  return path.join(context.phaseDir, "govulncheck-findings.json");
}
