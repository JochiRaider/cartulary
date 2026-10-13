import { createReadStream } from "node:fs";
import { createInterface } from "node:readline";

import { validateSchemaSync } from "../../contract/index.mjs";

function asciiCompare(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function regexEscape(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/gu, "\\$&");
}

function exactAlternation(values) {
  return `^(?:${values.map(regexEscape).join("|")})$`;
}

export function buildGoInvocations(rows, workers, command = process.env.GO || "go") {
  const byPackage = new Map();
  for (const row of rows) {
    const values = byPackage.get(row.selector.package) ?? [];
    values.push(row);
    byPackage.set(row.selector.package, values);
  }
  return [...byPackage.entries()]
    .sort(([left], [right]) => asciiCompare(left, right))
    .map(([packageName, packageRows]) => ({
      command,
      args: [
        "test",
        "-json",
        "-count=1",
        `-p=${Math.min(workers, 16)}`,
        "-run",
        exactAlternation(
          [...new Set(packageRows.flatMap((row) => row.selector.tests))].sort(asciiCompare),
        ),
        packageName,
      ],
      rows: packageRows.map((row) => ({
        row_id: row.row_id,
        selectors: [...row.selector.tests],
      })),
    }));
}

const failureMarker = "CARTULARY_HARNESS_TEST_FAILURE=";

function observeEvent(event, terminal, markers) {
  if (
    typeof event.Test === "string" &&
    ["pass", "fail", "skip"].includes(event.Action)
  ) {
    terminal.set(event.Test, {
      action: event.Action,
      duration_ms: Math.max(0, Math.round(Number(event.Elapsed ?? 0) * 1000)),
    });
  }
  if (typeof event.Test === "string" && typeof event.Output === "string") {
    const markerIndex = event.Output.indexOf(failureMarker);
    if (markerIndex >= 0) {
      const encoded = event.Output.slice(markerIndex + failureMarker.length).trim();
      try {
        const envelope = JSON.parse(encoded);
        validateSchemaSync("cartulary.harness_test_failure.v1", envelope);
        markers.push({ test: event.Test, envelope, malformed: false });
      } catch {
        markers.push({ test: event.Test, envelope: null, malformed: true });
      }
    }
  }
}

function parseEvents(stdout) {
  const terminal = new Map();
  const markers = [];
  for (const line of String(stdout).split("\n")) {
    if (line.trim() === "") continue;
    let event;
    try {
      event = JSON.parse(line);
    } catch {
      continue;
    }
    observeEvent(event, terminal, markers);
  }
  return { terminal, markers };
}

async function parseEventsFile(file, diagnostic) {
  const terminal = new Map();
  const markers = [];
  const input = createReadStream(file, { encoding: "utf8" });
  const lines = createInterface({ input, crlfDelay: Number.POSITIVE_INFINITY });
  try {
    for await (const line of lines) {
      if (line.trim() === "") continue;
      if (Buffer.byteLength(line, "utf8") > 1024 * 1024) {
        throw new Error("Go JSON event exceeds the 1 MiB line contract");
      }
      try {
        const event = JSON.parse(line);
        observeEvent(event, terminal, markers);
        diagnostic?.observe(event);
      } catch {
        // `go test -json` may relay non-JSON process diagnostics. Selector
        // closure below classifies missing terminal events without retaining
        // the unbounded raw stream.
      }
    }
  } finally {
    lines.close();
    input.destroy();
  }
  return { terminal, markers };
}

function markerFailure(selectors, markers) {
  const relevant = markers.filter(({ test }) =>
    selectors.some((selector) => test === selector || test.startsWith(`${selector}/`)),
  );
  if (relevant.length === 0) return null;
  if (relevant.some((entry) => entry.malformed)) {
    return {
      terminal_state: "failed",
      exit_code: 11,
      failure_class: "harness",
      failure_reason: "scheduler_accounting_error",
      failure_diagnostic: null,
    };
  }
  const unique = new Map(
    relevant.map((entry) => [JSON.stringify(entry.envelope), entry.envelope]),
  );
  if (unique.size !== 1) {
    return {
      terminal_state: "failed",
      exit_code: 11,
      failure_class: "harness",
      failure_reason: "scheduler_accounting_error",
      failure_diagnostic: null,
    };
  }
  const envelope = [...unique.values()][0];
  if (envelope.failure_class === "interrupted") {
    return {
      terminal_state: "cancelled",
      exit_code: 130,
      failure_class: envelope.failure_class,
      failure_reason: envelope.failure_reason,
      failure_diagnostic: envelope,
    };
  }
  return {
    terminal_state: "infrastructure_failed",
    exit_code: 3,
    failure_class: envelope.failure_class,
    failure_reason: envelope.failure_reason,
    failure_diagnostic: envelope,
  };
}

function adaptParsedGoInvocation(invocation, result, { terminal, markers }) {
  return invocation.rows.map((row) => {
    const observations = row.selectors.map((selector) => terminal.get(selector));
    const missing = observations.some((entry) => entry === undefined);
    const skipped = observations.some((entry) => entry?.action === "skip");
    const failed = observations.some((entry) => entry?.action === "fail");
    const typedFailure = markerFailure(row.selectors, markers);
    const terminalState = typedFailure?.terminal_state ?? (missing || skipped
      ? "infrastructure_failed"
      : failed
        ? "failed"
        : "passed");
    return {
      row_id: row.row_id,
      terminal_state: terminalState,
      duration_ms: observations.reduce((sum, entry) => sum + (entry?.duration_ms ?? 0), 0),
      exit_code: typedFailure?.exit_code ?? (terminalState === "passed" ? 0 : terminalState === "failed" ? 10 : 3),
      failure_class: typedFailure?.failure_class ?? (missing || skipped
        ? "infra"
        : failed
          ? "product"
          : null),
      failure_reason: typedFailure?.failure_reason ?? (missing
        ? "missing_selector_result"
        : skipped
          ? "unauthorized_skip"
          : failed
            ? "test_assertion_failure"
            : null),
      failure_diagnostic: typedFailure?.failure_diagnostic ?? null,
    };
  });
}

export function adaptGoInvocation(invocation, result) {
  return adaptParsedGoInvocation(invocation, result, parseEvents(result.stdout));
}

// Diagnostics are collected during the same parse that establishes row outcomes.
// Keep at most 128 active 4 KiB tails and eight completed failure tails; later
// passing output cannot evict already observed failures. These excerpts are not
// accounting evidence and the caller must redact before retained publication.
function failureDetails(invocation) {
  const selectors = invocation.rows.flatMap((row) => row.selectors);
  const relevant = (name) => typeof name === "string" && selectors.some((selector) => name === selector || name.startsWith(`${selector}/`));
  const active = new Map(), failed = new Map();
  return {
    observe(event) {
      if (!relevant(event.Test)) return;
      const name = event.Test;
      if (typeof event.Output === "string" && Buffer.byteLength(name) <= 512) {
        const before = active.get(name) ?? Buffer.alloc(0);
        active.delete(name);
        if (active.size === 128) active.delete(active.keys().next().value);
        active.set(name, Buffer.from(Buffer.concat([before, Buffer.from(event.Output)]).subarray(-4096)));
      }
      if (["pass", "fail", "skip"].includes(event.Action)) {
        if (event.Action === "fail" && failed.size < 8) failed.set(name, active.get(name));
        active.delete(name);
      }
    },
    render(terminal) {
      const names = [...terminal].filter(([name, event]) => event.action === "fail" && relevant(name)).map(([name]) => name);
      if (!names.length) return "";
      const selected = names.slice(0, 8);
      return selected.map((name) => `[GO FAILURE] ${Buffer.from(name).subarray(0, 512).toString()}\n${failed.get(name)?.toString() ?? "detail unavailable within active-test bound"}\n`).join("") +
        `[GO FAILURE DETAIL] selected=${selected.length} omitted=${names.length - selected.length}; at most 4 KiB per test\n`;
    },
  };
}

export async function adaptGoInvocationFile(invocation, result, onFailureOutput = () => {}) {
  const diagnostic = failureDetails(invocation);
  const parsed = await parseEventsFile(result.stdoutPath, diagnostic);
  onFailureOutput(diagnostic.render(parsed.terminal));
  return adaptParsedGoInvocation(invocation, result, parsed);
}
