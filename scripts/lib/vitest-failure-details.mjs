#!/usr/bin/env node

import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  prettyJSONString,
  secureWriteFile,
  validateSchemaSync,
} from "../../tools/harness/core/harness-contract.mjs";

const schemaID = "cartulary.vitest_failure_details.v1";
const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..", "..");

function normalizePath(value) {
  return String(value ?? "").replaceAll("\\", "/");
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

function isVitestFileResult(value) {
  return Boolean(
    value &&
      typeof value === "object" &&
      typeof value.name === "string" &&
      Array.isArray(value.assertionResults),
  );
}

function appendVitestFileResults(value, fileResults, visited) {
  if (!value || typeof value !== "object") {
    return;
  }
  if (isVitestFileResult(value)) {
    fileResults.push(value);
    return;
  }
  if (visited.has(value)) {
    return;
  }
  visited.add(value);
  if (Array.isArray(value)) {
    for (const entry of value) {
      appendVitestFileResults(entry, fileResults, visited);
    }
    return;
  }
  for (const key of ["testResults", "projectResults", "projects", "results"]) {
    if (!Array.isArray(value[key])) {
      continue;
    }
    for (const entry of value[key]) {
      appendVitestFileResults(entry, fileResults, visited);
    }
  }
}

function collectVitestFileResults(report) {
  const fileResults = [];
  appendVitestFileResults(report, fileResults, new Set());
  return fileResults;
}

function normalizeFailureMessages(failureMessage, failureMessages = []) {
  const messages = Array.isArray(failureMessages)
    ? failureMessages.filter((entry) => typeof entry === "string")
    : [];
  if (messages.length > 0) {
    return messages;
  }
  return typeof failureMessage === "string" && failureMessage !== ""
    ? [failureMessage]
    : [];
}

function firstVitestAppFrame(message) {
  const frame = String(message ?? "")
    .split("\n")
    .map((line) => line.trim())
    .find((line) =>
      /(?:^at\s+|^\()\/?home\/.*\/cartulary\/apps\/web\/src\//.test(line),
    );
  if (!frame) {
    return "";
  }
  const match = frame.match(/(apps\/web\/src\/[^:)]+):([0-9]+)(?::[0-9]+)?/);
  if (!match) {
    return "";
  }
  return `${match[1]}:${match[2]}`;
}

function diagnosticTags(messageOrMessages) {
  const message = Array.isArray(messageOrMessages)
    ? messageOrMessages.join("\n")
    : String(messageOrMessages ?? "");
  const tags = [];
  if (message.includes("STACK_TRACE_ERROR")) {
    tags.push("vitest_stack_trace_error");
  }
  if (
    message.includes('Unable to find an element by: [data-testid="row-') ||
    message.includes("Expected workbook rows for surface")
  ) {
    tags.push("workbook_row_hydration_wait");
  }
  if (
    message.includes("controlled_input_replacement_mismatch") ||
    message.includes("Expected input value")
  ) {
    tags.push("controlled_input_replacement");
  }
  return [...new Set(tags)].sort();
}

function bestMessage({ fallback, failureMessage = "", failureMessages = [] }) {
  const messages = normalizeFailureMessages(failureMessage, failureMessages);
  for (const message of messages) {
    for (const line of message.split("\n")) {
      const trimmed = line.trim();
      if (
        trimmed &&
        trimmed !== "Error: STACK_TRACE_ERROR" &&
        !trimmed.startsWith("at ")
      ) {
        return {
          message: trimmed,
          messageSource: "runner_json_assertion_message",
        };
      }
    }
  }
  const combined = messages.join("\n");
  if (combined.includes("STACK_TRACE_ERROR")) {
    return {
      message:
        "Vitest reporter emitted STACK_TRACE_ERROR before preserving the assertion message",
      messageSource: "runner_json_stack_trace_error",
    };
  }
  return {
    message: fallback,
    messageSource: "runner_json_fallback",
  };
}

function failureRecord({
  ownerPath,
  title,
  status,
  failureMessage,
  failureMessages,
  fallback,
}) {
  const rawMessages = normalizeFailureMessages(failureMessage, failureMessages);
  const combined = rawMessages.join("\n");
  const message = bestMessage({ fallback, failureMessage, failureMessages });
  const firstAppFrame = firstVitestAppFrame(combined);
  const renderedMessage =
    message.messageSource === "runner_json_stack_trace_error"
      ? [
          message.message,
          `file=${ownerPath || "(unknown)"}`,
          `title=${title || "(unknown)"}`,
          firstAppFrame ? `first_app_frame=${firstAppFrame}` : "",
        ]
          .filter(Boolean)
          .join("; ")
      : message.message;
  return {
    owner_path: ownerPath,
    title,
    status,
    message: renderedMessage,
    message_source: message.messageSource,
    raw_messages: rawMessages,
    diagnostic_tags: diagnosticTags(rawMessages),
    first_app_frame: firstAppFrame,
  };
}

function collectFailures(report) {
  const failures = [];
  for (const fileResult of collectVitestFileResults(report)) {
    const ownerPath = relToRepo(fileResult.name ?? "");
    const assertions = fileResult.assertionResults ?? [];
    const executedAssertions = assertions.filter(
      (assertion) => assertion.status !== "skipped",
    );
    if (executedAssertions.length === 0 && fileResult.status === "failed") {
      failures.push(
        failureRecord({
          ownerPath,
          title: "(suite load)",
          status: fileResult.status ?? "failed",
          failureMessage: fileResult.message ?? "",
          failureMessages: [],
          fallback: "test file failed before a top-level test was attributed",
        }),
      );
      continue;
    }
    for (const assertion of assertions) {
      if (
        assertion.status === "passed" ||
        assertion.status === "skipped" ||
        assertion.status === "todo"
      ) {
        continue;
      }
      const failureMessages = Array.isArray(assertion.failureMessages)
        ? assertion.failureMessages
        : [];
      failures.push(
        failureRecord({
          ownerPath,
          title: assertion.title ?? "(missing title)",
          status: assertion.status ?? "failed",
          failureMessage: failureMessages[0] ?? "",
          failureMessages,
          fallback: `${assertion.title ?? "vitest assertion"} failed`,
        }),
      );
    }
  }
  return failures;
}

function usage() {
  console.error(
    "usage: vitest-failure-details.mjs <runner-json> <output-json> [stdout-log] [stderr-log]",
  );
}

const [runnerJSON, outputJSON, stdoutLog = "", stderrLog = ""] =
  process.argv.slice(2);
if (!runnerJSON || !outputJSON) {
  usage();
  process.exit(2);
}
if (!existsSync(runnerJSON)) {
  console.error(`runner JSON does not exist: ${runnerJSON}`);
  process.exit(1);
}

const report = JSON.parse(readFileSync(runnerJSON, "utf8"));
const details = {
  schema_id: schemaID,
  runner_json: relToRepo(runnerJSON),
  stdout_log: existsSync(stdoutLog) ? relToRepo(stdoutLog) : "",
  stderr_log: existsSync(stderrLog) ? relToRepo(stderrLog) : "",
  generated_at: new Date().toISOString(),
  failures: collectFailures(report),
};

validateSchemaSync(schemaID, details);
secureWriteFile(outputJSON, prettyJSONString(details));
