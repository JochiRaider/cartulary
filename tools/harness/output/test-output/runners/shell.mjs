#!/usr/bin/env node
import { repoRoot } from "../../../contract/index.mjs";

import {
  existsSync,
  readFileSync,
  readdirSync,
  rmSync,
  statSync,
} from "node:fs";
import path from "node:path";
import {
  classifyExecutionFailure,
  classifyExecutionFailureReason,
  primaryPublicFailure,
} from "../../../contract/failure-taxonomy.mjs";
import { verboseOutput } from "../../tool-output.mjs";
import {
  testCoverageBuckets,
} from "../../../contract/test-output-context.mjs";
import {
  createBaseStepContext,
  writeStepArtifacts,
} from "../step-artifacts.mjs";
import { loadGovulncheckFindingsFile } from "../security-diagnostics.mjs";

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

function createCounts() {
  const counts = {
    tests: 0,
    failed: 0,
    non_test: 0,
    non_test_failed: 0,
    packages: 0,
  };
  for (const coverage of testCoverageBuckets) {
    counts[coverage] = 0;
    counts[`${coverage}_failed`] = 0;
  }
  return counts;
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

function splitLogLines(file) {
  if (!file || !existsSync(file)) {
    return [];
  }
  return readFileSync(file, "utf8").split(/\r?\n/);
}

function removeEmptyArtifact(file) {
  if (!file) {
    return;
  }
  let stat;
  try {
    stat = statSync(file);
  } catch (error) {
    if (error?.code === "ENOENT") {
      return;
    }
    throw error;
  }
  if (stat.size === 0) {
    rmSync(file, { force: true });
  }
}

function catalogOwnerFromEnvironment() {
  return optionalEnv("CARTULARY_CATALOG_OWNER_ID");
}

function firstActionableLine(lines) {
  for (const rawLine of lines) {
    const line = rawLine.trim();
    if (line === "") {
      continue;
    }
    if (
      line.startsWith("=== RUN") ||
      line.startsWith("--- PASS") ||
      line.startsWith("--- FAIL") ||
      line.startsWith("--- SKIP") ||
      line === "PASS" ||
      line === "FAIL" ||
      isShellOrchestrationLine(line) ||
      /^ok\s/.test(line) ||
      /^\?\s/.test(line)
    ) {
      continue;
    }
    return line;
  }
  return "";
}

function firstKnownToolDiagnosticLine(lines) {
  return firstShellCheckDiagnosticLine(lines) || firstBiomeDiagnosticLine(lines);
}

function firstShellCheckDiagnosticLine(lines) {
  let pendingLocation = null;

  for (const rawLine of lines) {
    const line = rawLine.trim();
    if (line === "") {
      continue;
    }

    const locationMatch = line.match(/^In (.+) line ([0-9]+):$/);
    if (locationMatch) {
      pendingLocation = {
        file: locationMatch[1],
        line: locationMatch[2],
      };
      continue;
    }

    const humanDiagnosticMatch = line.match(
      /^(?:\^--\s*)?(SC[0-9]+)\s+\([^)]+\):\s*(.+)$/,
    );
    if (humanDiagnosticMatch && pendingLocation) {
      return `ShellCheck ${humanDiagnosticMatch[1]} at ${pendingLocation.file}:${pendingLocation.line}: ${humanDiagnosticMatch[2]}`;
    }

    const gccDiagnosticMatch = line.match(
      /^(.+?):([0-9]+):[0-9]+:\s+[^:]+:\s+(.+?)\s+\[(SC[0-9]+)\]$/,
    );
    if (gccDiagnosticMatch) {
      return `ShellCheck ${gccDiagnosticMatch[4]} at ${gccDiagnosticMatch[1]}:${gccDiagnosticMatch[2]}: ${gccDiagnosticMatch[3]}`;
    }
  }

  return "";
}

function firstBiomeDiagnosticLine(lines) {
  for (const rawLine of lines) {
    const line = rawLine.trim();
    if (line === "") {
      continue;
    }

    const ruleDiagnosticMatch = line.match(
      /^(.+?):([0-9]+):([0-9]+)\s+((?:assist|lint)\/[A-Za-z0-9/_-]+)\b/u,
    );
    if (ruleDiagnosticMatch) {
      return `Biome ${ruleDiagnosticMatch[4]} at ${ruleDiagnosticMatch[1]}:${ruleDiagnosticMatch[2]}:${ruleDiagnosticMatch[3]}`;
    }

    const formatDiagnosticMatch = line.match(/^(.+?)\s+format\s+━/u);
    if (formatDiagnosticMatch) {
      return `Biome format at ${formatDiagnosticMatch[1]}`;
    }
  }

  return "";
}

function isShellOrchestrationLine(line) {
  return /^\[TARGET\]\s+start\s+\S+\s/.test(line);
}

function printBlock(header, fields) {
  const lines = [header];
  for (const [key, value] of Object.entries(fields)) {
    lines.push(`${key}=${value === "" ? "-" : value}`);
  }
  process.stderr.write(`${lines.join("\n")}\n`);
}

function showStepDetailOutput(context) {
  return verboseOutput() || context.target === "adhoc";
}

function finalizeShellStep(context, stdoutLog, stderrLog, details) {
  removeEmptyArtifact(stdoutLog);
  removeEmptyArtifact(stderrLog);

  writeStepArtifacts(context, {
    ...details,
    artifacts: {
      stdout_log: existsSync(stdoutLog) ? stdoutLog : "",
      stderr_log: existsSync(stderrLog) ? stderrLog : "",
    },
  });

  if (details.status === "pass") {
    return 0;
  }

  if (showStepDetailOutput(context)) {
    for (const dossier of details.dossiers) {
      printBlock(`failure: ${context.label}`, dossier);
    }
  }
  return 1;
}

function browserStartupDiagnosticFailureDetails(context) {
  const servicesRoot = path.resolve(
    repoRoot,
    requiredEnv("CARTULARY_TEST_RESULTS_DIR"),
    requiredEnv("CARTULARY_TEST_RUN_ID"),
    "_shared",
    "test-services",
  );
  const diagnostics = [];
  if (existsSync(servicesRoot)) {
    for (const suite of readdirSync(servicesRoot, { withFileTypes: true })) {
      const sessionsRoot = path.join(servicesRoot, suite.name, "browser-sessions");
      if (!suite.isDirectory() || !existsSync(sessionsRoot)) continue;
      for (const session of readdirSync(sessionsRoot, { withFileTypes: true })) {
        if (!session.isDirectory()) continue;
        const file = path.join(sessionsRoot, session.name, "startup-diagnostics.json");
        if (!existsSync(file)) continue;
        try {
          const payload = JSON.parse(readFileSync(file, "utf8"));
          if (
            payload?.schema_id === "cartulary.browser_startup_diagnostics.v2" &&
            payload?.status === "failed" &&
            typeof payload?.failure_class === "string" &&
            typeof payload?.failure_reason === "string"
          ) {
            diagnostics.push({
              failure_class: payload.failure_class,
              failure_reason: payload.failure_reason,
              kind: "browser_startup",
              source: "browser_session",
              target: context.target,
              step: context.label,
              message: payload.message ?? "",
              artifact: normalizePath(path.relative(repoRoot, file)),
            });
          }
        } catch {
          // Malformed artifacts are classified by the ordinary artifact boundary.
        }
      }
    }
  }
  const primary = primaryPublicFailure(diagnostics);
  return primary
    ? {
        failure_class: primary.failure_class,
        failure_reason: primary.failure_reason,
      }
    : null;
}

function classifyShellFailureDetails(context, stdoutLines, stderrLines, message) {
  const startupDiagnostic = browserStartupDiagnosticFailureDetails(context);
  if (startupDiagnostic) {
    return startupDiagnostic;
  }
  if (govulncheckFindingsValidationError(context)) {
    return {
      failure_class: "artifact",
      failure_reason: "artifact_error",
    };
  }

  const text = [
    context.target,
    context.label,
    context.command,
    message,
    ...stderrLines,
    ...stdoutLines,
  ]
    .join("\n")
    .toLowerCase();

  if (
    text.includes("built frontend artifact missing") ||
    text.includes("run make build-web before browser e2e") ||
    text.includes("repo-local pnpm was not found") ||
    text.includes("node runtime bootstrap failed") ||
    text.includes("node archive checksum mismatch") ||
    text.includes("missing committed node checksum") ||
    text.includes("unsupported node bootstrap platform")
  ) {
    return { failure_class: "config", failure_reason: "configuration_error" };
  }
  if (
    text.includes("port must differ") ||
    text.includes("is already in use") ||
    text.includes("address already in use") ||
    text.includes("eaddrinuse") ||
    text.includes("failed to allocate an available") ||
    text.includes("listener conflict")
  ) {
    return { failure_class: "infra", failure_reason: "resource_conflict" };
  }
  if (
    text.includes("timed out waiting for frontend") ||
    text.includes("timed out waiting for backend") ||
    text.includes("timed out waiting for service") ||
    text.includes("service readiness timeout")
  ) {
    return {
      failure_class: "infra",
      failure_reason: "service_readiness_timeout",
    };
  }
  if (
    text.includes("system limit for number of file watchers reached") ||
    (text.includes("enospc") && text.includes("watch")) ||
    text.includes("frontend exited before readiness") ||
    text.includes("backend exited before readiness") ||
    text.includes("exited immediately after readiness") ||
    text.includes("exited unexpectedly during browser e2e supervision")
  ) {
    return { failure_class: "infra", failure_reason: "service_start_error" };
  }
  if (isSecurityScannerFindingFailure(context, text)) {
    return {
      failure_class: "security",
      failure_reason: "security_finding",
    };
  }
  if (isToolDiagnosticShellFailure(context, text)) {
    return {
      failure_class: "harness",
      failure_reason: "tool_diagnostic_failure",
    };
  }

  return {
    failure_class: classifyExecutionFailure(
      context.target,
      context.label,
      context.command,
    ),
    failure_reason: classifyExecutionFailureReason(
      context.target,
      context.label,
      context.command,
    ),
  };
}

function firstStructuredToolFailureLine(context) {
  return govulncheckArtifactErrorLine(context) || govulncheckFailureLine(context);
}

function isSecurityScannerFindingFailure(context, text) {
  return isGovulncheckSecurityFinding(context) || isGosecSecurityFinding(context, text);
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
  return path.join(context.stepDir, "govulncheck-findings.json");
}

function readGovulncheckFindings(context) {
  const file = govulncheckFindingsPath(context);
  if (!file || !existsSync(file)) {
    return null;
  }
  return loadGovulncheckFindingsFile(file).findings;
}

function govulncheckFindingsValidationError(context) {
  const file = govulncheckFindingsPath(context);
  if (!file || !existsSync(file)) {
    return "";
  }
  const result = loadGovulncheckFindingsFile(file);
  return result.error ?? "";
}

function govulncheckArtifactErrorLine(context) {
  const error = govulncheckFindingsValidationError(context);
  if (!error) {
    return "";
  }
  return `Govulncheck findings artifact is invalid: ${error}`;
}

function govulncheckFailureLine(context) {
  const findings = readGovulncheckFindings(context);
  const blockingCount = findings?.counts?.blocking_count ?? 0;
  if (blockingCount <= 0) {
    return "";
  }
  const ids = (findings.blocking_vulnerability_ids ?? []).slice(0, 5);
  const suffix = ids.length > 0 ? `: ${ids.join(",")}` : "";
  return `govulncheck found ${blockingCount} symbol-reachable vulnerabilities${suffix}`;
}

function isGovulncheckSecurityFinding(context) {
  const findings = readGovulncheckFindings(context);
  return (findings?.counts?.blocking_count ?? 0) > 0;
}

function isGosecSecurityFinding(context, text) {
  const commandContext = [context.target, context.label, context.command]
    .join("\n")
    .toLowerCase();
  if (
    !commandContext.includes("go-gosec-targeted") &&
    !commandContext.includes("gosec")
  ) {
    return false;
  }
  return /\bg[0-9]{3}\b/u.test(text) || /\bissues?:\s*[1-9][0-9]*\b/u.test(text);
}

function isToolDiagnosticShellFailure(context, text) {
  return (
    isShellCheckShellFailure(context, text) ||
    isBiomeShellFailure(context, text) ||
    isGovulncheckToolDiagnosticFailure(context, text)
  );
}

function isGovulncheckToolDiagnosticFailure(context, text) {
  const commandContext = [context.target, context.label, context.command]
    .join("\n")
    .toLowerCase();
  if (
    !commandContext.includes("go-vulncheck") &&
    !commandContext.includes("govulncheck")
  ) {
    return false;
  }
  return text.includes("govulncheck json parse failed");
}

function isShellCheckShellFailure(context, text) {
  const commandContext = [
    context.target,
    context.label,
    context.command,
  ]
    .join("\n")
    .toLowerCase();
  if (
    !commandContext.includes("lint-shell") &&
    !commandContext.includes("run-shellcheck.sh") &&
    !commandContext.includes("shellcheck")
  ) {
    return false;
  }
  return (
    /\bsc[0-9]{4}\s+\((?:error|warning|info|style)\):/u.test(text) ||
    /\[[ ]*sc[0-9]{4}[ ]*\]/u.test(text)
  );
}

function isBiomeShellFailure(context, text) {
  const commandContext = [
    context.target,
    context.label,
    context.command,
  ]
    .join("\n")
    .toLowerCase();
  if (
    !commandContext.includes("lint-biome") &&
    !commandContext.includes("run-frontend-biome.sh") &&
    !commandContext.includes("biome")
  ) {
    return false;
  }
  return (
    /\b(?:assist|lint)\/[a-z0-9/_-]+/u.test(text) ||
    /\breporter\/(?:format|violations)\b/u.test(text) ||
    text.includes("formatter would have printed") ||
    text.includes("some warnings were emitted while running checks") ||
    /\bfound [0-9]+ (?:errors?|warnings?)\b/u.test(text)
  );
}

export function handleShellStep() {
  const context = createBaseStepContext("shell");
  const stdoutLog = requiredEnv("CARTULARY_STEP_STDOUT_LOG");
  const stderrLog = requiredEnv("CARTULARY_STEP_STDERR_LOG");

  if (context.exitStatus === 0) {
    return finalizeShellStep(context, stdoutLog, stderrLog, {
      status: "pass",
      step: catalogOwnerFromEnvironment(),
      counts: {
        ...createCounts(),
      },
      owners: [],
      inventory: [],
      dossiers: [],
    });
  }

  const stderrLines = splitLogLines(stderrLog);
  const stdoutLines = splitLogLines(stdoutLog);
  const structuredFailureLine = firstStructuredToolFailureLine(context);
  const messageBase =
    firstKnownToolDiagnosticLine(stderrLines) ||
    firstKnownToolDiagnosticLine(stdoutLines) ||
    structuredFailureLine ||
    firstActionableLine(stderrLines) ||
    firstActionableLine(stdoutLines) ||
    `command exited with status ${context.exitStatus}`;
  const failureNote = optionalEnv("CARTULARY_STEP_FAILURE_NOTE");
  const message =
    failureNote === ""
      ? messageBase
      : `${messageBase} | remediation: ${failureNote}`;
  const failureDetails = classifyShellFailureDetails(
    context,
    stdoutLines,
    stderrLines,
    message,
  );
  return finalizeShellStep(context, stdoutLog, stderrLog, {
    status: "fail",
    step: catalogOwnerFromEnvironment(),
    counts: {
      ...createCounts(),
      failed: 1,
      non_test: 1,
      non_test_failed: 1,
    },
    owners: [],
    inventory: [],
    dossiers: [
      {
        failure_class: failureDetails.failure_class,
        failure_reason: failureDetails.failure_reason,
        coverage: "non_test",
        step: catalogOwnerFromEnvironment(),
        id: "",
        runner: "shell",
        package_or_file: "(shell command)",
        symbol_or_title: "(shell command)",
        message,
        reproduce: context.command,
        raw: renderRawList([stdoutLog, stderrLog]),
      },
    ],
  });
}

function renderRawList(paths) {
  return paths
    .filter((entry) => entry && existsSync(entry))
    .map((entry) => relToRepo(entry))
    .join(";");
}
