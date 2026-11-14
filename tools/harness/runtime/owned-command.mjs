import path from "node:path";
import { constants } from "node:os";
import { spawn } from "../workspace/child-process.mjs";
import { workspaceLayout } from "../../workspace_layout.generated.mjs";
import { parseStrictJSON } from "../contract/index.mjs";
import { CommandFailure } from "./command-failure.mjs";
import { parseCleanupDiagnostic, withCleanupDiagnostic } from "./cleanup-diagnostic.mjs";

const keysAre = (value, keys) => value && typeof value === "object" && !Array.isArray(value) &&
  Object.keys(value).sort().join(",") === [...keys].sort().join(",");
const positivePID = value => Number.isSafeInteger(value) && value > 1;
const terminationSignal = value => Number.isSafeInteger(value) && value >= 1 && value <= 64;
const signalName = number => Object.keys(constants.signals).find(name => constants.signals[name] === number) ?? `SIG${number}`;
const failedAcquisition = error => ({ code: null, signal: null, started: false, settled: true, error });
function validProof(proof) {
  return keysAre(proof, ["boot", "pid", "start", "group", "pgid", "sid"]) &&
    typeof proof.boot === "string" && /^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$/u.test(proof.boot) &&
    positivePID(proof.pid) && proof.group === true && proof.pgid === proof.pid && positivePID(proof.sid) &&
    typeof proof.start === "string" && /^[1-9][0-9]*$/u.test(proof.start);
}
function protocolFailure() {
  return withCleanupDiagnostic(new CommandFailure("native command ownership unresolved", {
    failure_class: "harness", failure_reason: "cleanup_error",
  }), "process_protocol");
}
function captureOutputFailure(kind) {
  return Object.assign(new CommandFailure(`private child capture ${kind === "limit" ? "byte limit exceeded" : "write failed"}`, {
    failure_class: "artifact", failure_reason: "artifact_error",
  }), { captureFailure: kind });
}

// The native child owns workload launch, signals, descendant settlement and
// reaping. This adapter only registers its proof and consumes bounded facts.
// The result keeps workload status separate from transport/registration failure;
// only settled=true permits a caller to retire its process journal entries.
export async function runOwnedCommand(command, args, { cwd, environment, signal, output = "ignore", captureLimitBytes, onReady = () => {} }) {
  if (typeof command !== "string" || !command || command.includes("\0") || !Array.isArray(args) || args.some(arg => typeof arg !== "string" || arg.includes("\0"))) {
    return failedAcquisition(new Error("invalid owned command invocation"));
  }
  let streams;
  if (["ignore", "inherit"].includes(output)) streams = [output, output];
  else if (keysAre(output, ["stdout", "stderr"]) && Object.values(output).every(fd => Number.isSafeInteger(fd) && fd >= 0 && fd <= 2147483647)) streams = [output.stdout, output.stderr];
  else return failedAcquisition(new Error("invalid owned command output policy"));
  if (captureLimitBytes !== undefined && (typeof output !== "object" || !Number.isSafeInteger(captureLimitBytes) || captureLimitBytes < 1 || captureLimitBytes > 256 * 1024 * 1024)) {
    return failedAcquisition(new Error("invalid owned command capture limit"));
  }
  if (signal?.aborted) return failedAcquisition(signal.reason);
  const root = path.resolve(import.meta.dirname, "../../..");
  const binary = environment.CARTULARY_HARNESS_RUNTIME_BIN || path.join(root, workspaceLayout.toolbin, "cartulary-harness");
  let child;
  try {
    const mode = captureLimitBytes === undefined ? ["pipe"] : ["capture", String(captureLimitBytes)];
    child = spawn(binary, ["owned-command", ...mode, command, ...args], {
      cwd, env: environment, stdio: ["pipe", ...streams, "pipe"], detached: true,
    });
  } catch (error) { return failedAcquisition(error); }
  let ready, terminal, error, invalid = false, registered = false, stopRequested = false, stopSent = false, bytes = 0, pending = Buffer.alloc(0), frames = 0;
  const remember = failure => {
    if (!error) error = failure;
    else if (error !== failure) (error.cleanupFailures ??= []).push(failure);
  };
  const sendStop = () => {
    if (stopSent) return;
    stopSent = true;
    if (!child.stdin.destroyed) child.stdin.write("stop\n");
  };
  const stop = () => {
    stopRequested = true;
    // Cancellation before readiness must not race the parent's proof lookup.
    // Native adoption remains pending until that synchronous lookup finishes.
    if (registered) sendStop();
  };
  const rejectProtocol = () => {
    if (!invalid) remember(protocolFailure());
    invalid = true;
    pending = Buffer.alloc(0);
    sendStop();
  };
  child.stdin.on("error", () => {
    // A late stop may race a valid terminal receipt. The closed result below
    // still requires exact terminal evidence; EPIPE itself never is evidence.
  });
  const completion = new Promise(resolve => {
    child.once("error", remember);
    child.once("close", (code, terminatedBy) => resolve({ code, terminatedBy }));
  });
  child.stdio[3].on("error", rejectProtocol);
  child.stdio[3].on("data", data => {
    bytes += data.length;
    if (invalid) return;
    if (bytes > 8192) { rejectProtocol(); return; }
    pending = Buffer.concat([pending, data]);
    while (pending.includes(10)) {
      const boundary = pending.indexOf(10);
      const line = pending.subarray(0, boundary);
      pending = pending.subarray(boundary + 1);
      if (++frames > 2 || line.length > 4096) { rejectProtocol(); return; }
      let event;
      try { event = parseStrictJSON(new TextDecoder("utf-8", { fatal: true }).decode(line)); }
      catch { rejectProtocol(); return; }
      const captureFailure = event?.capture_failure;
      const captureKeys = Object.hasOwn(event ?? {}, "capture_failure") ? ["capture_failure"] : [];
      if (captureKeys.length && (captureLimitBytes === undefined || !["limit", "io"].includes(captureFailure))) { rejectProtocol(); return; }
      if (terminal) { rejectProtocol(); return; }
      if (event?.event === "ready" && !ready && keysAre(event, ["event", "proof"]) && validProof(event.proof)) {
        ready = event.proof;
        try {
          const result = onReady(ready);
          if (result?.then) throw new Error("owned command registration must be synchronous");
        } catch (failure) { remember(failure); stop(); }
        registered = true;
        if (stopRequested) sendStop();
        else child.stdin.write("adopt\n");
      } else if (event?.event === "settled" && ready && keysAre(event, [...(Object.hasOwn(event, "signal") ? ["event", "exit", "cause", "signal"] : ["event", "exit", "cause"]), ...captureKeys]) &&
          Number.isSafeInteger(event.exit) && event.exit >= 0 && event.exit <= 255 &&
          (!Object.hasOwn(event, "signal") || terminationSignal(event.signal) && event.exit === 128 + event.signal) &&
          ["exited", "stopped", "control_lost", "parent_lost", "protocol", "adoption_timeout", "signal", "capture"].includes(event.cause) &&
          (event.cause !== "capture" || captureFailure)) {
        terminal = event;
      } else if (event?.event === "failed" && keysAre(event, [...(event.diagnostic ? ["event", "cause", "diagnostic"] : ["event", "cause"]), ...captureKeys]) &&
          (ready ? event.cause === "cleanup" : ["parent", "admission", "launch", "launch_missing", "cleanup"].includes(event.cause))) {
        const diagnostic = event.diagnostic ? parseCleanupDiagnostic(Buffer.from(JSON.stringify(event.diagnostic))) : null;
        if (event.diagnostic && !diagnostic) { rejectProtocol(); return; }
        terminal = event;
        if (captureFailure) remember(captureOutputFailure(captureFailure));
        const failure = event.cause === "launch_missing" ? Object.assign(new Error("owned command executable missing"), { code: "ENOENT" })
          : event.cause === "launch" ? new Error("owned command launch failed") : protocolFailure();
        if (diagnostic) Object.defineProperty(failure, "cleanupDiagnostic", { value: diagnostic, configurable: true });
        remember(failure);
      } else { rejectProtocol(); return; }
    }
    if (pending.length > 4096) rejectProtocol();
  });
  signal?.addEventListener("abort", stop, { once: true });
  if (signal?.aborted) stop();
  const native = await completion;
  signal?.removeEventListener("abort", stop);
  child.stdin.destroy();
  if (!child.pid) return failedAcquisition(error ?? protocolFailure());
  const expected = terminal?.event === "settled" && ["exited", "stopped"].includes(terminal.cause) ? 0 : 12;
  if (pending.length || !terminal || native.terminatedBy || native.code !== expected) rejectProtocol();
  const completed = !invalid && terminal?.event === "settled";
  const unacquired = !invalid && terminal?.event === "failed" && !ready && ["parent", "admission", "launch", "launch_missing"].includes(terminal.cause);
  if (completed && terminal.capture_failure) remember(captureOutputFailure(terminal.capture_failure));
  if (completed && !["exited", "stopped", "capture"].includes(terminal.cause)) remember(protocolFailure());
  return { code: completed ? terminal.exit : null, signal: completed && terminal.signal ? signalName(terminal.signal) : null, started: Boolean(ready), settled: completed || unacquired, error };
}
