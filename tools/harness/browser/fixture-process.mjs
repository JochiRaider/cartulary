import { spawn } from "node:child_process";
import { randomUUID } from "node:crypto";
import { StringDecoder } from "node:string_decoder";
import { redactString, redactValue, validateSchemaSync } from "../contract/index.mjs";

export { fixtureLifecycleAttachment } from "../contract/browser-fixture-diagnostic.mjs";

// This owner retains private pipe data until close, not just process exit. Only
// report() and its redacted error messages may cross into retained evidence.
export function createBrowserFixtureProcess({ command, args, cwd, env = process.env,
  startupMs = 90_000, cleanupMs = 35_000, signalMs = 5_000 }) {
  const attempt = randomUUID();
  const started = performance.now();
  const child = spawn(command, [...args, "--attempt-id", attempt], {
    cwd, env, detached: true, stdio: ["pipe", "pipe", "pipe"],
  });
  const secrets = new Set();
  const registerSecret = (value) => {
    secrets.add(value);
    // Retention can discard an earlier line of a credential. Each remaining
    // nonempty line still belongs to the same private value.
    for (const line of value.split(/\r?\n/u)) if (line) secrets.add(line);
  };
  for (const [key, value] of Object.entries(env)) {
    if (value && redactValue(value, key) !== value) registerSecret(value);
  }
  const sanitize = (text) => {
    let value = String(text);
    for (const secret of [...secrets].sort((a, b) => b.length - a.length))
      value = value.replaceAll(secret, "[REDACTED]");
    return redactString(value);
  };
  let stderr = "", line = "", stdout = "", droppingLine = false;
  let readySeen = false, closed = false, exitCode = null, signal = null;
  let forced = false, stopping = null, protocolError = null;
  const events = [];
  const decoder = new StringDecoder("utf8");
  const retainLine = (value) => {
    if (!value) return;
    if (value.startsWith('{"schema_id":"cartulary.browser_fixture_event.')) {
      try {
        const event = JSON.parse(value);
        validateSchemaSync("cartulary.browser_fixture_event.v1", event);
        if (event.attempt_id !== attempt) throw new Error("fixture attempt mismatch");
        if (events.length >= 256) throw new Error("fixture event limit exceeded");
        events.push(event);
      } catch { protocolError = "Invalid browser fixture lifecycle event"; }
    } else {
      // Keep whole lines until final redaction; truncating a credential before
      // redaction could turn it into an unrecognizable, exposed suffix.
      stderr += `${value}\n`;
      while (stderr.length > 32_768) stderr = stderr.slice(stderr.indexOf("\n") + 1);
    }
  };
  child.stderr.on("data", (chunk) => {
    for (const char of decoder.write(chunk)) {
      if (char === "\n") {
        if (!droppingLine) retainLine(line);
        line = ""; droppingLine = false;
      } else if (!droppingLine) {
        line += char;
        if (line.length > 32_768) { line = ""; droppingLine = true; }
      }
    }
  });
  let resolveReady, rejectReady;
  const ready = new Promise((resolve, reject) => { resolveReady = resolve; rejectReady = reject; });
  const timer = setTimeout(() => rejectReady(new Error("Browser fixture startup deadline exceeded")), startupMs);
  const finished = new Promise((resolve) => {
    child.once("close", (code, receivedSignal) => {
      if (!droppingLine) retainLine(line + decoder.end());
      closed = true; exitCode = code; signal = receivedSignal;
      clearTimeout(timer);
      if (!readySeen) rejectReady(new Error(`Browser fixture exited before ready: ${sanitize(stderr)}`));
      resolve();
    });
  });
  child.once("error", (error) => {
    clearTimeout(timer);
    rejectReady(new Error(sanitize(error.message)));
  });
  child.stdin.on("error", (error) => {
    if (error.code !== "EPIPE") retainLine(`fixture stdin: ${error.message}`);
  });
  child.stdout.on("data", (chunk) => {
    if (readySeen) return;
    stdout += chunk.toString("utf8");
    if (stdout.length > 262_144) {
      stdout = ""; clearTimeout(timer);
      rejectReady(new Error("Browser fixture ready payload exceeded its bound")); return;
    }
    const newline = stdout.indexOf("\n");
    if (newline < 0) return;
    try {
      const payload = JSON.parse(stdout.slice(0, newline));
      for (const [key, value] of Object.entries(payload)) {
        if (typeof value === "string" && value && redactValue(value, key) !== value) registerSecret(value);
      }
      readySeen = true; clearTimeout(timer); stdout = ""; resolveReady(payload);
    } catch {
      clearTimeout(timer); stdout = "";
      rejectReady(new Error("Invalid browser fixture ready payload"));
    }
  });
  const wait = async (duration) => {
    let timeout;
    await Promise.race([finished, new Promise((resolve) => { timeout = setTimeout(resolve, duration); })]);
    clearTimeout(timeout);
  };
  const kill = (name) => {
    if (!child.pid) return;
    try { process.kill(-child.pid, name); }
    catch (error) { if (error.code !== "ESRCH") throw error; }
  };
  const stop = () => stopping ??= (async () => {
    clearTimeout(timer);
    if (!closed) { child.stdin.end(); await wait(cleanupMs); }
    if (!closed) {
      forced = true; kill("SIGTERM"); await wait(signalMs);
      if (!closed) { kill("SIGKILL"); await wait(signalMs); }
    }
    if (readySeen && !events.some((event) => event.phase === "terminal")) protocolError ??= "Missing terminal fixture evidence";
    if (forced || !closed || exitCode !== 0 || protocolError) {
      const failed = events.filter((event) => event.outcome === "failed");
      throw new Error(sanitize(`Browser fixture cleanup failed: code=${exitCode} signal=${signal} forced=${forced} closed=${closed}\n${protocolError ?? ""}\n${failed.map((event) => `${event.stage}: ${event.message}`).join("\n")}\n${stderr}`));
    }
  })();
  return {
    ready, stop,
    report(failures = []) {
      const report = {
        schema_id: "cartulary.browser_fixture_lifecycle.v1", attempt_id: attempt,
        duration_ms: Math.round(performance.now() - started),
        process: { exit_code: exitCode, signal, closed, forced },
        events: events.map((event) => ({ ...event, message: sanitize(event.message).slice(0, 1024) })),
        stderr: sanitize(stderr),
        failures: failures.map(({ phase, failure_class, failure_reason, error }) => ({
          phase, failure_class, failure_reason, message: sanitize(String(error?.message ?? error)).slice(0, 16_384),
        })),
      };
      validateSchemaSync(report.schema_id, report);
      return report;
    },
  };
}
