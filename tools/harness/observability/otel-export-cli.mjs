#!/usr/bin/env node

import { lstatSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  loadRetainedObservability,
  resolveExactRunDir,
} from "./observability.mjs";

export const exporterTimeoutMs = 10_000;
const usage = "usage: otel-export-cli.mjs --results-dir <root|run-dir> [--run-id <id>] --endpoint <url> [--headers-file <0600-json-file>]";

export function parseExporterArgs(argv) {
  const options = { resultsDir: "", runID: "", endpoint: "", headersFile: "" };
  const names = new Map([
    ["--results-dir", "resultsDir"],
    ["--run-id", "runID"],
    ["--endpoint", "endpoint"],
    ["--headers-file", "headersFile"],
  ]);
  for (let index = 0; index < argv.length; index += 1) {
    const field = names.get(argv[index]);
    const value = argv[index + 1];
    if (!field || !value) throw new Error(usage);
    options[field] = value;
    index += 1;
  }
  if (!options.resultsDir || !options.endpoint) throw new Error(usage);
  return options;
}

function isIPv4Loopback(hostname) {
  const octets = hostname.split(".");
  return octets.length === 4 &&
    octets[0] === "127" &&
    octets.every((octet) => /^\d{1,3}$/u.test(octet) && Number(octet) <= 255);
}

export function validatedEndpoint(value) {
  const authority = /^[A-Za-z][A-Za-z0-9+.-]*:\/\/([^/?#]*)/u.exec(value)?.[1] ?? "";
  if (!authority || /[%\\\u0000-\u001f\u007f]/u.test(authority)) {
    throw new Error("HARNESS_OTLP_ENDPOINT contains an ambiguous authority");
  }
  let endpoint;
  try {
    endpoint = new URL(value);
  } catch {
    throw new Error("HARNESS_OTLP_ENDPOINT must be an absolute URL");
  }
  if (endpoint.username || endpoint.password || endpoint.search || endpoint.hash) {
    throw new Error("HARNESS_OTLP_ENDPOINT forbids credentials, query, and fragment");
  }
  const loopback = endpoint.hostname === "localhost" ||
    endpoint.hostname === "::1" ||
    endpoint.hostname === "[::1]" ||
    isIPv4Loopback(endpoint.hostname);
  if (endpoint.protocol !== "https:" && !(endpoint.protocol === "http:" && loopback)) {
    throw new Error("HARNESS_OTLP_ENDPOINT requires HTTPS except for loopback HTTP");
  }
  return endpoint;
}

export function signalURL(endpoint, signal) {
  const next = new URL(endpoint.href);
  const basePath = next.pathname.replace(/\/+$/u, "").replace(/\/v1\/(?:traces|metrics)$/u, "");
  next.pathname = `${basePath}/v1/${signal}`.replace(/^\/+/u, "/");
  return next;
}

export function headersFromFile(file) {
  if (!file) return {};
  const resolved = path.resolve(file);
  const stats = lstatSync(resolved);
  const ownedByCurrentUser = typeof process.getuid !== "function" || stats.uid === process.getuid();
  if (!stats.isFile() || stats.isSymbolicLink() || !ownedByCurrentUser || (stats.mode & 0o777) !== 0o600) {
    throw new Error("HARNESS_OTLP_HEADERS_FILE must be a regular owner-only 0600 JSON file");
  }
  if (stats.size > 65_536) throw new Error("HARNESS_OTLP_HEADERS_FILE exceeds 64 KiB");
  const value = JSON.parse(readFileSync(resolved, "utf8"));
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error("HARNESS_OTLP_HEADERS_FILE must contain one JSON object");
  }
  if (Object.keys(value).length > 32) throw new Error("HARNESS_OTLP_HEADERS_FILE contains more than 32 headers");
  const headers = {};
  const forbiddenProtocolHeaders = new Set([
    "connection",
    "content-length",
    "content-type",
    "host",
    "keep-alive",
    "proxy-connection",
    "te",
    "trailer",
    "transfer-encoding",
    "upgrade",
  ]);
  for (const [name, headerValue] of Object.entries(value)) {
    if (
      !/^[A-Za-z0-9!#$%&'*+.^_`|~-]{1,64}$/u.test(name) ||
      typeof headerValue !== "string" ||
      Buffer.byteLength(headerValue, "utf8") > 4096 ||
      /[\r\n]/u.test(headerValue)
    ) {
      throw new Error("HARNESS_OTLP_HEADERS_FILE contains an unsafe header entry");
    }
    if (forbiddenProtocolHeaders.has(name.toLowerCase())) {
      throw new Error("HARNESS_OTLP_HEADERS_FILE cannot override protocol headers");
    }
    headers[name] = headerValue;
  }
  return headers;
}

export async function deliver(
  endpoint,
  payload,
  headers,
  deliveryTimeoutMs = exporterTimeoutMs,
  fetchImpl = globalThis.fetch,
) {
  const response = await fetchImpl(endpoint, {
    method: "POST",
    redirect: "error",
    signal: AbortSignal.timeout(deliveryTimeoutMs),
    headers: { ...headers, "content-type": "application/json" },
    body: JSON.stringify(payload),
  });
  if (response.redirected || (response.status >= 300 && response.status < 400)) {
    throw new Error("collector redirects are forbidden");
  }
  if (!response.ok) throw new Error(`collector rejected payload with HTTP ${response.status}`);
}

export function loadExporterInput(options) {
  const endpoint = validatedEndpoint(options.endpoint);
  const headers = headersFromFile(options.headersFile);
  const runDir = resolveExactRunDir(options.resultsDir, options.runID);
  const retained = loadRetainedObservability(runDir);
  return { endpoint, headers, retained };
}

export async function exportRetainedObservability(
  input,
  { deliveryTimeoutMs = exporterTimeoutMs, fetchImpl = globalThis.fetch } = {},
) {
  let signals = 0;
  for (const { result: invocation } of input.retained.built) {
    await deliver(
      signalURL(input.endpoint, "traces"),
      invocation.traceOTLP,
      input.headers,
      deliveryTimeoutMs,
      fetchImpl,
    );
    signals += 1;
    await deliver(
      signalURL(input.endpoint, "metrics"),
      invocation.metricsOTLP,
      input.headers,
      deliveryTimeoutMs,
      fetchImpl,
    );
    signals += 1;
  }
  return { invocations: input.retained.built.length, signals };
}

async function main() {
  let input;
  try {
    input = loadExporterInput(parseExporterArgs(process.argv.slice(2)));
  } catch {
    process.stderr.write("harness-otel-export FAIL failure_class=config reason=configuration_error diagnostic=invalid-export-configuration\n");
    process.exitCode = 2;
    return;
  }
  try {
    const result = await exportRetainedObservability(input);
    process.stdout.write(`harness-otel-export PASS invocations=${result.invocations} signals=${result.signals}\n`);
  } catch {
    process.stderr.write("harness-otel-export FAIL failure_class=harness reason=tool_diagnostic_failure diagnostic=delivery-failed\n");
    process.exitCode = 1;
  }
}

if (path.resolve(process.argv[1] ?? "") === fileURLToPath(import.meta.url)) {
  await main();
}
