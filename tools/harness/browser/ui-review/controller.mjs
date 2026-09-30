import { preparationFailure } from "./failure.mjs";
import { fork } from "node:child_process";
import { chmodSync } from "node:fs";
import net from "node:net";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { parseStrictJSON } from "../../contract/index.mjs";
import { parseRequest, result, ReviewFailure, failureRecord, limits } from "./contract.mjs";
import { emitResult } from "./output.mjs";
import { readInput, readLocator, resolveSession, terminalResult } from "./session-files.mjs";

function terminal(identity, command) {
  if (!["ui-review-status", "ui-review-stop"].includes(command)) throw new ReviewFailure("session_mismatch");
  return terminalResult(identity, command, (receipt, ref) => result(command, { session_id: identity.locator.session_id, state: identity.locator.state, receipt: ref, status: receipt.status, exit_code: receipt.exit_code, failures: receipt.failures }));
}
async function send(record, message) {
  return new Promise((resolve, reject) => {
    const connection = net.createConnection(record.socket); let data = Buffer.alloc(0), completed = false;
    const timer = setTimeout(() => { connection.destroy(new ReviewFailure("session_lost")); }, message.command === "ui-review-stop" ? 360000 : limits.operation + limits.lock + 10000);
    const interrupt = () => connection.destroy(new ReviewFailure("interrupted"));
    process.once("SIGINT", interrupt); process.once("SIGTERM", interrupt);
    const finish = (error, response) => { if (completed) return; completed = true; clearTimeout(timer); process.off("SIGINT", interrupt); process.off("SIGTERM", interrupt); connection.destroy(); if (error) reject(error); else resolve(response); };
    connection.once("connect", () => connection.write(`${JSON.stringify({ token: record.token, session_id: record.session_id, ...message })}\n`));
    connection.on("data", (part) => {
      if (data.length + part.length > limits.request) { finish(new ReviewFailure("invalid_artifact")); return; }
      data = Buffer.concat([data, part]);
      if (data.at(-1) === 10) { try { finish(null, parseStrictJSON(new TextDecoder("utf-8", { fatal: true }).decode(data))); } catch (cause) { finish(new ReviewFailure("invalid_artifact", { cause })); } }
    });
    connection.on("error", (cause) => finish(cause instanceof ReviewFailure ? cause : new ReviewFailure("session_lost", { cause })));
    connection.once("end", () => { if (!completed) finish(new ReviewFailure("session_lost")); });
  });
}
export async function execute(command, input, profile) {
  if (command === "ui-review") return start(input, profile);
  const identity = readLocator(input.UI_SESSION);
  if (["closed", "failed"].includes(identity.locator.state)) {
    let response = terminal(identity, command);
    if (command === "ui-review-stop" && identity.locator.state === "failed") {
      const record = resolveSession(identity, { allowDead: true, allowAbsent: true });
      if (record?.dead) {
        const { recoverSession } = await import("./recovery.mjs");
        response = { ...await recoverSession(record), command_id: response.command_id };
      }
    }
    emitResult(response, input.output); process.exitCode = response.exit_code; return;
  }
  let request;
  if (input.UI_REQUEST) request = parseRequest(readInput(input.UI_REQUEST, { extension: ".json", maximum: limits.request }), command === "ui-browser" ? "action" : command === "ui-capture" ? "capture_request" : "analysis_request");
  const record = resolveSession(identity, { allowDead: command === "ui-review-stop" });
  if (record.dead) {
    const { recoverSession } = await import("./recovery.mjs");
    const response = { ...await recoverSession(record), command_id: result(command).command_id };
    emitResult(response, input.output); process.exitCode = response.exit_code; return;
  }
  const response = await send(record, { command, request: request ?? null, bundle_id: input.UI_BUNDLE ?? null });
  emitResult(response, input.output); process.exitCode = response.exit_code;
}
async function start(input, profile) {
  const environment = { ...process.env };
  for (const name of Object.keys(environment)) {
    if (/^(?:UI_|REVIEW_PROFILE$|MAKEFLAGS$|MAKEOVERRIDES$|MFLAGS$|NODE_OPTIONS$|NODE_PATH$|NODE_V8_COVERAGE$|DEBUG$|PWDEBUG$|OTEL_|CARTULARY_(?:MAKE_|HARNESS_|BROWSER_|WEB_E2E_|PGTEST_|S3_)|CARTULARY__)/u.test(name)) delete environment[name];
  }
  const worker = fork(fileURLToPath(import.meta.url), ["--controller"], { detached: true, stdio: ["ignore", "ignore", "ignore", "ipc"], execArgv: [], env: environment });
  let record, completed = false;
  const forward = (signal) => { if (worker.connected) worker.send({ stop: signal }); };
  const interrupt = () => forward("SIGINT"), terminate = () => forward("SIGTERM");
  process.on("SIGINT", interrupt); process.on("SIGTERM", terminate);
  worker.send({ input, profile });
  await new Promise((resolve, reject) => {
    worker.on("message", (message) => {
      if (message.record) { record = message.record; process.stdout.write(`UI review preparing. UI_SESSION=${record.locator}\n`); worker.send({ acknowledged: true }); }
      if (message.ready) process.stdout.write(`UI review ready (${input.UI_MODE}). UI_SESSION=${message.locator}\nPrivate review links expire on stop or session failure.\n`);
      if (message.terminal) { completed = true; emitResult(message.terminal, input.output); process.exitCode = message.terminal.exit_code; }
    });
    worker.once("error", reject);
    worker.once("exit", async () => {
      process.off("SIGINT", interrupt); process.off("SIGTERM", terminate);
      if (!completed) {
        try {
          if (record) {
            const identity = readLocator(record.locator);
            const { recoverSession } = await import("./recovery.mjs");
            const response = ["closed", "failed"].includes(identity.locator.state) ? terminal(identity, "ui-review-stop") : await recoverSession(record);
            emitResult(response, input.output); process.exitCode = response.exit_code;
          } else throw preparationFailure(new Error("preparation controller unavailable"));
        } catch (error) { reject(error); return; }
      }
      resolve();
    });
  });
}
async function serve() {
  const { input, profile } = await new Promise((resolve) => process.once("message", resolve));
  const { ReviewSession } = await import("./session.mjs");
  const session = new ReviewSession(input, profile);
  let server, stopping;
  const end = (error) => {
    stopping ??= (async () => {
      try {
        await session.stop(error);
        if (process.connected) process.send({ terminal: session.terminalValue("ui-review") });
      } finally {
        server?.close();
        if (process.connected) process.disconnect();
      }
    })();
    return stopping;
  };
  process.once("disconnect", () => { if (!stopping) void end(new ReviewFailure("session_lost")); });
  process.on("SIGINT", () => { void end(new ReviewFailure("interrupted", { exitCode: 130 })); });
  process.on("SIGTERM", () => { void end(new ReviewFailure("interrupted", { exitCode: 143 })); });
  try {
    const record = session.initialize(process.env);
    const acknowledged = new Promise((resolve) => process.once("message", resolve));
    process.send({ record }); await acknowledged;
    process.on("message", (message) => { if (message.stop) void end(new ReviewFailure("interrupted", { exitCode: message.stop === "SIGINT" ? 130 : 143 })); });
    server = net.createServer((socket) => {
      let bytes = Buffer.alloc(0), submitted = false, answered = false;
      const timer = setTimeout(() => socket.destroy(), 5000);
      socket.on("error", () => {});
      socket.on("close", () => { clearTimeout(timer); if (submitted && !answered) void end(new ReviewFailure("interrupted")); });
      socket.on("data", async (part) => {
        if (submitted || bytes.length + part.length > limits.request + 1024) { socket.destroy(); return; }
        bytes = Buffer.concat([bytes, part]); if (bytes.at(-1) !== 10) return;
        let envelope;
        try {
          envelope = parseStrictJSON(new TextDecoder("utf-8", { fatal: true }).decode(bytes));
          if (Object.keys(envelope).sort().join(",") !== "bundle_id,command,request,session_id,token" || envelope.token !== record.token || envelope.session_id !== session.sessionID || !["ui-review-status", "ui-browser", "ui-capture", "ui-analyze", "ui-review-report", "ui-review-stop"].includes(envelope.command)) throw new Error("unauthorized controller request");
        } catch { socket.destroy(); return; }
        submitted = true; clearTimeout(timer);
        let response;
        try {
          const requestKind = { "ui-browser": "action", "ui-capture": "capture_request", "ui-analyze": "analysis_request" }[envelope.command];
          if (requestKind) {
            if (envelope.bundle_id !== null) throw new ReviewFailure("invalid_request");
            envelope.request = parseRequest(Buffer.from(JSON.stringify(envelope.request)), requestKind);
          } else if (envelope.request !== null || (envelope.command === "ui-review-report" ? !/^bundle-[1-9][0-9]*$/u.test(envelope.bundle_id) : envelope.bundle_id !== null)) throw new ReviewFailure("invalid_request");
          response = await session.handle(envelope.command, envelope.request, envelope.bundle_id);
        }
        catch (error) { response = session.fail(envelope.command, error); }
        answered = true; socket.end(`${JSON.stringify(response)}\n`);
        if (["closed", "failed"].includes(session.state)) void end();
      });
    });
    await new Promise((resolve, reject) => { server.once("error", reject); server.listen(record.socket, resolve); }); chmodSync(record.socket, 0o600);
    await session.prepare();
    if (process.connected) process.send({ ready: true, locator: session.locatorFile });
    // Browser loss and lifetime expiry may originate inside the session.
    const watch = setInterval(() => { if (["closed", "failed"].includes(session.state)) { clearInterval(watch); void end(); } }, 100);
  } catch (error) {
    if (session.runtime) await end(preparationFailure(error));
    else {
      if (process.connected) { process.send({ terminal: result("ui-review", { status: "error", exit_code: preparationFailure(error).exitCode, failures: [failureRecord(preparationFailure(error))] }) }); process.disconnect(); }
    }
  }
}
if (process.argv[2] === "--controller") await serve();
