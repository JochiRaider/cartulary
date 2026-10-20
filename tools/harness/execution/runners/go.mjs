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

export function buildGoInvocations(rows, workers, command, context) {
  if (context.goShuffleSeed !== undefined && (!/^[1-9][0-9]*$/u.test(String(context.goShuffleSeed)) ||
    Number(context.goShuffleSeed) > 2147483647)) throw new Error("invalid Go shuffle seed");
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
      environment: { ...context.environment, ...(context.goRace ? { CGO_ENABLED: "1" } : {}) },
      args: [
        "test",
        "-json",
        "-count=1",
        ...(context.goRace ? ["-race"] : []),
        ...(context.goShuffleSeed === undefined ? [] : [`-shuffle=${context.goShuffleSeed}`]),
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

function goEventObserver() {
  const terminal = new Map(), markers = [], partial = new Map();
  const malformed = (test) => markers.push({ test, envelope: null, malformed: true });
  const finishLine = (state, terminated) => {
    if (state.encoded === null) return;
    if (!terminated || state.overflow) { malformed(state.test); return; }
    try {
      const envelope = JSON.parse(state.encoded.trim());
      validateSchemaSync("cartulary.harness_test_failure.v4", envelope);
      markers.push({ test: state.test, envelope, malformed: false });
    } catch { malformed(state.test); }
  };
  const empty = (test) => ({ test, tail: "", encoded: null, bytes: 0, overflow: false });
  return {
    observe(event) {
      if (typeof event.Test !== "string") return;
      // Package identity also prevents a malformed mixed-package stream from
      // joining fragments belonging to equal test names in different packages.
      const key = JSON.stringify([event.Package ?? "", event.Test]);
      if (typeof event.Output === "string") {
        let state = partial.get(key) ?? empty(event.Test);
        partial.delete(key);
        const pieces = event.Output.split("\n");
        for (let index = 0; index < pieces.length; index++) {
          const piece = pieces[index];
          if (state.encoded === null) {
            const candidate = state.tail + piece;
            const markerIndex = candidate.indexOf(failureMarker);
            if (markerIndex >= 0) {
              state.encoded = candidate.slice(markerIndex + failureMarker.length);
              state.bytes = Buffer.byteLength(state.encoded);
              state.tail = "";
            } else state.tail = candidate.slice(-(failureMarker.length - 1));
          } else if (!state.overflow) {
            state.bytes += Buffer.byteLength(piece);
            if (state.bytes <= 64 * 1024) state.encoded += piece;
          }
          if (state.bytes > 64 * 1024) { state.overflow = true; state.encoded = ""; }
          if (index < pieces.length - 1) {
            finishLine(state, true);
            state = empty(event.Test);
          }
        }
        if (state.encoded !== null || state.tail !== "") {
          if (partial.size >= 128) malformed(event.Test);
          else partial.set(key, state);
        }
      }
      if (["pass", "fail", "skip"].includes(event.Action)) {
        const state = partial.get(key);
        if (state) finishLine(state, false);
        partial.delete(key);
        terminal.set(event.Test, {
          action: event.Action,
          duration_ms: Math.max(0, Math.round(Number(event.Elapsed ?? 0) * 1000)),
        });
      }
    },
    finish() {
      for (const state of partial.values()) finishLine(state, false);
      partial.clear();
      return { terminal, markers };
    },
  };
}

function parseGoEventLine(line) {
  if (Buffer.byteLength(line, "utf8") > 1024 * 1024) {
    throw new Error("Go JSON event exceeds the 1 MiB line contract");
  }
  try {
    const event = JSON.parse(line);
    return event !== null && typeof event === "object" && !Array.isArray(event) ? event : null;
  } catch { return null; }
}

function parseEvents(stdout) {
  const observer = goEventObserver();
  for (const line of String(stdout).split("\n")) {
    if (line.trim() === "") continue;
    const event = parseGoEventLine(line);
    if (event) observer.observe(event);
  }
  return observer.finish();
}

async function parseEventsFile(file, diagnostic) {
  const observer = goEventObserver();
  const input = createReadStream(file, { encoding: "utf8" });
  const lines = createInterface({ input, crlfDelay: Number.POSITIVE_INFINITY });
  try {
    for await (const line of lines) {
      if (line.trim() === "") continue;
      const event = parseGoEventLine(line);
      // Process diagnostics and non-event JSON cannot complete a fragment.
      if (!event) continue;
      observer.observe(event);
      diagnostic?.observe(event);
    }
  } finally {
    lines.close();
    input.destroy();
  }
  return observer.finish();
}

function markerFailure(selectors, markers, terminal) {
  const relevant = markers.filter(({ test }) =>
    selectors.some((selector) => test === selector || test.startsWith(`${selector}/`)),
  );
  if (relevant.length === 0) return null;
  const accountingFailure = () => ({
    terminal_state: "failed", exit_code: 11, failure_class: "harness",
    failure_reason: "scheduler_accounting_error", failure_diagnostic: null,
  });
  if (relevant.some((entry) => entry.malformed)) return accountingFailure();
  const setup = new Map(relevant.filter(({ envelope }) => envelope.phase === "setup")
    .map(({ envelope }) => [JSON.stringify(envelope), envelope]));
  if (setup.size > 1) return accountingFailure();
  const cleanupEntries = relevant.filter(({ envelope }) => envelope.phase === "cleanup");
  const cleanup = [...new Map(cleanupEntries.map(({ envelope }) =>
    [JSON.stringify(envelope), envelope])).values()];
  if (cleanup.length > 32) return accountingFailure();
  // A later cleanup sees the failure raised by an earlier cleanup. Its prior
  // flag must not manufacture a product assertion. Descendant failures are
  // likewise already attributed before a parent's cleanup begins.
  const knownFailures = relevant.filter(({ envelope }) => envelope.phase === "setup").map(({ test }) => test);
  let priorProductFailure = false;
  for (const entry of cleanupEntries) {
    const accounted = knownFailures.some((test) => test === entry.test || test.startsWith(`${entry.test}/`));
    if (entry.envelope.prior_failure && !accounted) priorProductFailure = true;
    knownFailures.push(entry.test);
  }
  const envelope = [...setup.values()][0] ?? null;
  const diagnostic = {
    schema_id: "cartulary.harness_go_failure_diagnostic.v3",
    setup: envelope, cleanup,
  };
  validateSchemaSync(diagnostic.schema_id, diagnostic);
  if (!selectors.some((selector) => terminal.get(selector)?.action === "fail")) return accountingFailure();
  // Parent fail events can reflect their typed descendants. A failing sibling
  // or independent selector without a typed cause is still a product assertion
  // and retains the ordinary class precedence over setup and cleanup failures.
  priorProductFailure ||= [...terminal].some(([test, observation]) => observation.action === "fail" &&
    selectors.some((selector) => test === selector || test.startsWith(`${selector}/`)) &&
    !knownFailures.some((known) => known === test || known.startsWith(`${test}/`)));
  if (priorProductFailure) {
    return {
      terminal_state: "failed", exit_code: 10, failure_class: "product",
      failure_reason: "test_assertion_failure", failure_diagnostic: diagnostic,
    };
  }
  if (envelope) {
    return {
      terminal_state: envelope.failure_class === "interrupted" ? "cancelled" : "infrastructure_failed",
      exit_code: envelope.failure_class === "interrupted" ? 130 : 3,
      failure_class: envelope.failure_class, failure_reason: envelope.failure_reason,
      failure_diagnostic: diagnostic,
    };
  }
  return {
    terminal_state: "failed", exit_code: 12,
    failure_class: "harness", failure_reason: "cleanup_error",
    failure_diagnostic: diagnostic,
  };
}

function adaptParsedGoInvocation(invocation, result, { terminal, markers }) {
  return invocation.rows.map((row) => {
    const observations = row.selectors.map((selector) => terminal.get(selector));
    const missing = observations.some((entry) => entry === undefined);
    const skipped = observations.some((entry) => entry?.action === "skip");
    const failed = observations.some((entry) => entry?.action === "fail");
    const typedFailure = markerFailure(row.selectors, markers, terminal);
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
