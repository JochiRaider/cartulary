import { randomBytes } from "node:crypto";
import { accessSync, constants } from "node:fs";
import path from "node:path";
import { createSuiteRuntime, scanRetainedRoot } from "../../runtime/suite-runtime.mjs";
import { acquireHostAdmission } from "../../runtime/host-admission.mjs";
import { atomicLocalFile, privateDirectory, readLocalFile } from "../../runtime/secure-local-files.mjs";
import { captureCapabilitySnapshot } from "../../scheduler/work-graph/capability.mjs";
import { buildSourceSnapshot } from "../../test-catalog/source-snapshot.mjs";
import { runPreparedReview } from "../review-preparation.mjs";
import { ReviewBrowser, browserReady } from "./browser.mjs";
import { commandID, emptyCounts, failureRecord, limits, result, ReviewFailure, schemaID } from "./contract.mjs";
import { jsonBytes, newRunRoot, publishJSON, registerSession, unregisterSession } from "./session-files.mjs";
import { repoRoot } from "./toolchain.mjs";

const now = () => new Date().toISOString();
const pause = (milliseconds) => new Promise((resolve) => setTimeout(resolve, milliseconds));
export class ReviewSession {
  constructor(input, profile) {
    this.input = input; this.profile = profile; this.mode = input.UI_MODE;
    this.operationID = 0; this.total = emptyCounts(); this.started = now(); this.startedTick = performance.now();
    this.sessionID = `uireview-${randomBytes(16).toString("hex")}`; this.state = "preparing";
    this.failures = []; this.exitCode = 0; this.abort = new AbortController(); this.receipt = null;
    this.privateBytes = 0;
  }
  initialize(environment) {
    if (this.mode !== "artifacts") {
      browserReady();
      try {
        accessSync("/usr/bin/flock", constants.X_OK);
        if (this.mode === "seeded") for (const file of ["server-harness", "migrate", "tmp/toolbin/cartulary-test-services"]) accessSync(path.join(repoRoot, file), constants.X_OK);
      } catch (cause) { throw new ReviewFailure("tool_configuration", { cause }); }
    }
    const { runRoot, runID } = newRunRoot(environment); this.runRoot = runRoot; this.runID = runID;
    this.locatorFile = path.join(runRoot, "ui-review/session.json");
    this.runtime = createSuiteRuntime({ repoRoot, runRoot, runID });
    this.record = registerSession({ locatorFile: this.locatorFile, sessionID: this.sessionID, runtime: this.runtime });
    this.publishLocator(); return this.record;
  }
  publishLocator() {
    publishJSON(this.locatorFile, "session", { schema_id: schemaID("session"), session_id: this.sessionID, run_id: this.runID, mode: this.mode, state: this.state, created_at: this.started, updated_at: now(), terminal_receipt: this.receipt }, { replace: true });
  }
  async prepare() {
    this.capacity = captureCapabilitySnapshot({ root: repoRoot }).port_lanes;
    if (this.mode !== "artifacts") {
      try { this.hostLease = await acquireHostAdmission({ browsers: 1, browserCapacity: this.capacity, signal: this.abort.signal }); }
      catch (cause) { throw new ReviewFailure("capacity_exceeded", { cause }); }
      this.workspaceDigest = buildSourceSnapshot(repoRoot).digest.replace(/^sha256:/u, "");
      let origin = this.input.UI_ORIGIN;
      let actors = {};
      if (this.mode === "seeded") {
        const ready = Promise.withResolvers();
        const hold = Promise.withResolvers(); this.releasePreparation = hold.resolve;
        const lifecycleRoot = this.runtime.privatePath("lifecycle", this.runID);
        this.preparation = runPreparedReview({ environment: { ...process.env, REVIEW_PROFILE: this.input.REVIEW_PROFILE }, signal: this.abort.signal, target: "ui-review", runID: this.runID, runRoot: lifecycleRoot, runtime: this.runtime, provision: false, retainDetail: false, writeOutput: () => {}, onReady: ready.resolve, hold: () => hold.promise });
        this.preparation.catch(ready.reject);
        const prepared = await ready.promise;
        origin = prepared.attached.CARTULARY_WEB_E2E_PUBLIC_ORIGIN;
        const accounts = JSON.parse(readLocalFile(path.join(prepared.privateDirectory, "access.json"))).accounts;
        actors = { admin: accounts[0], empty: accounts[1], viewer: accounts[2], editor: accounts[3] };
        this.seeded = prepared;
        this.health = setInterval(() => { if (!this.checking && this.state === "ready") { this.checking = true; prepared.check().catch(() => this.stop(new ReviewFailure("session_lost"))).finally(() => { this.checking = false; }); } }, 3000);
      } else {
        const started = performance.now(); let reached = false;
        while (performance.now() - started < 30000) {
          this.abort.signal.throwIfAborted();
          try { const response = await fetch(origin, { redirect: "manual", signal: AbortSignal.any([this.abort.signal, AbortSignal.timeout(1000)]) });
            if (response.status >= 300 && response.status < 400 && new URL(response.headers.get("location"), origin).origin !== origin) throw new ReviewFailure("navigation_boundary");
            if (response.status < 500) { reached = true; await response.body?.cancel(); break; }
            await response.body?.cancel();
          } catch (error) { if (error instanceof ReviewFailure) throw error; }
          await pause(250);
        }
        if (!reached) throw new ReviewFailure("readiness_expired");
      }
      this.browser = new ReviewBrowser({ origin, mode: this.mode, actors, onLost: (error) => { void this.stop(error); } });
      await this.browser.start();
    }
    this.abort.signal.throwIfAborted(); this.state = "ready"; this.publishLocator();
    this.expiry = setTimeout(() => { void this.stop(); }, limits.lifetime);
  }
  value(command, fields = {}) {
    return result(command, { session_id: this.sessionID, state: this.state, epoch: ["closed", "failed"].includes(this.state) ? null : this.browser?.epoch ?? null, ...fields });
  }
  reserveBytes(bytes) {
    if (!Number.isSafeInteger(bytes) || bytes < 0 || this.privateBytes + bytes > limits.storage) throw new ReviewFailure("capacity_exceeded");
    this.privateBytes += bytes;
    let released = false;
    return () => { if (!released) { this.privateBytes -= bytes; released = true; } };
  }
  fail(command, error, fields = {}) {
    const normalized = error instanceof ReviewFailure ? error : new ReviewFailure("unsafe_artifact", { cause: error });
    return this.value(command, { status: "error", exit_code: normalized.exitCode, failures: [failureRecord(normalized)], ...fields });
  }
  async handle(command, request, bundleID) {
    if (command === "ui-review-status") return this.value(command, { receipt: this.receipt });
    if (command === "ui-review-stop") { await this.stop(); return this.terminalValue(command); }
    const waiting = performance.now();
    while (this.active && this.state === "busy") {
      if (performance.now() - waiting >= limits.lock) return this.fail(command, new ReviewFailure("capacity_exceeded"));
      await pause(20);
    }
    if (this.state !== "ready") return this.fail(command, new ReviewFailure("session_mismatch"));
    try {
      if (command === "ui-browser") {
        if (!this.browser) throw new ReviewFailure("invalid_request");
        this.browser.validateAction(request);
      }
      if (command === "ui-capture" && request.source === "page") {
        if (!this.browser) throw new ReviewFailure("invalid_request");
        if (request.expected_epoch !== this.browser.epoch || this.browser.needsSnapshot) throw new ReviewFailure("session_mismatch");
      }
    } catch (error) { return this.fail(command, error); }
    const operationID = ++this.operationID;
    this.state = "busy"; this.publishLocator();
    const active = this.operation(command, request, bundleID, operationID);
    this.active = active;
    // Normal Playwright action timeouts leave the page available for a fresh
    // snapshot. An undrainable protocol call closes the owned browser instead
    // of allowing a late operation to overlap the next request.
    const watchdog = setTimeout(() => { void this.stop(new ReviewFailure("operation_expired")); }, limits.operation);
    try { return await active; }
    finally { clearTimeout(watchdog); this.active = null; if (this.state === "busy") { this.state = "ready"; this.publishLocator(); } }
  }
  async operation(command, request, bundleID, operationID) {
    const started = now(), tick = performance.now(); let counts = emptyCounts(); let outcome;
    try {
      this.abort.signal.throwIfAborted();
      if (command === "ui-browser") {
        if (request.action === "fill") this.runtime.registerSecret(request.parameters.text);
        const observed = await this.browser.action(request);
        const file = this.runtime.privatePath(`operation-${operationID}`, "observations.json");
        const bytes = jsonBytes(observed), release = this.reserveBytes(bytes.length);
        try { atomicLocalFile(file, bytes); } catch (error) { release(); throw error; }
        counts.observed_elements = observed.elements.length;
        counts.console_errors = observed.console.records.filter((record) => record.level === "error").length;
        counts.failed_requests = observed.network.records.filter((record) => record.outcome === "failed").length;
        outcome = this.value(command, { operation_id: operationID, private_refs: [{ kind: "observations", absolute_path: file }] });
      } else {
        // Capture/import and analysis owners are installed by the following slices.
        const owner = command === "ui-capture" ? await import("./capture.mjs") : command === "ui-analyze" ? await import("./analysis.mjs") : await import("./report.mjs");
        const produced = await owner.execute(this, request ?? bundleID, operationID);
        counts = produced.counts ?? counts;
        outcome = this.value(command, { operation_id: operationID, bundle_id: produced.bundle_id, private_refs: produced.private_refs });
      }
    } catch (error) {
      const failure = this.abort.signal.aborted ? this.abort.signal.reason : error;
      outcome = this.fail(command, failure, { operation_id: operationID });
      if (error.cleanupFailures?.length) outcome.failures.push(failureRecord(new ReviewFailure("cleanup_failed")));
      if (this.offlineLeases?.size) queueMicrotask(() => { void this.stop(failure instanceof ReviewFailure ? failure : new ReviewFailure("unsafe_artifact")); });
    }
    if (outcome.status === "ok") for (const [key, value] of Object.entries(counts)) this.total[key] += value;
    const receipt = { schema_id: schemaID("receipt"), command_id: commandID(command), session_id: this.sessionID, operation_id: operationID, mode: this.mode, state: "ready", status: outcome.status, exit_code: outcome.exit_code, started_at: started, finished_at: now(), duration_ms: Math.floor(performance.now() - tick), failures: outcome.failures, counts, bundle_id: outcome.bundle_id, cleanup: "not_terminal" };
    try {
      const relative = `ui-review/operations/${operationID}/receipt.json`;
      const ref = publishJSON(path.join(this.runRoot, relative), "receipt", receipt);
      return { ...outcome, state: this.state === "busy" ? "ready" : this.state, receipt: { path: relative, ...ref }, ...(this.state === "stopping" ? { private_refs: [] } : {}) };
    } catch (cause) {
      // Stop asynchronously: waiting here would make stop wait for this operation.
      queueMicrotask(() => { void this.stop(new ReviewFailure("unsafe_artifact", { cause })); });
      return this.fail(command, new ReviewFailure("unsafe_artifact"), { operation_id: operationID });
    }
  }
  terminalValue(command) { return this.value(command, { status: this.failures.length ? "error" : "ok", exit_code: this.exitCode, failures: this.failures, receipt: this.receipt }); }
  stop(error) {
    if (this.stopping) return this.stopping;
    if (error) { this.failures.push(failureRecord(error)); this.exitCode = error.exitCode ?? 11; }
    this.stopCause = error;
    this.stopping = this.finish(); return this.stopping;
  }
  async finish() {
    this.state = "stopping"; clearTimeout(this.expiry); clearInterval(this.health);
    const cleanupFailures = [];
    const attempt = async (release) => { try { await release(); } catch (error) { cleanupFailures.push(error); } };
    await attempt(() => this.publishLocator());
    this.abort.abort(this.stopCause ?? new ReviewFailure("interrupted"));
    await attempt(() => this.browser?.close());
    this.releasePreparation?.();
    try { await this.preparation; } catch (error) {
      // Acquisition/action failure already belongs to the primary lifecycle
      // outcome. Only failures of the preparation owner's release are secondary.
      cleanupFailures.push(...(error.cleanupFailures ?? []));
      if (!this.failures.length && !this.abort.signal.aborted) { this.failures.push(failureRecord(new ReviewFailure("startup_failed"))); this.exitCode ||= 3; }
    }
    await attempt(async () => { await this.active; });
    await attempt(() => this.hostLease?.release());
    for (const lease of this.offlineLeases ?? []) await attempt(async () => { await lease.release(); this.offlineLeases.delete(lease); });
    await attempt(async () => {
      const scan = await scanRetainedRoot(this.runRoot, { forbiddenValues: this.runtime.forbiddenValues(), removeUnsafe: true });
      if (scan.status !== "pass") throw new ReviewFailure("unsafe_artifact");
    });
    if (cleanupFailures.length === 0) await attempt(() => this.runtime.close());
    if (cleanupFailures.length) {
      const artifact = cleanupFailures.some((failure) => failure instanceof ReviewFailure && failure.diagnostic === "unsafe_artifact");
      const failure = new ReviewFailure(artifact ? "unsafe_artifact" : "cleanup_failed");
      this.failures.push(failureRecord(failure)); this.exitCode ||= failure.exitCode;
    }
    this.failures = [...new Map(this.failures.map((failure) => [JSON.stringify(failure), failure])).values()];
    this.state = this.failures.length ? "failed" : "closed";
    const receipt = { schema_id: schemaID("receipt"), command_id: commandID("ui-review"), session_id: this.sessionID, operation_id: null, mode: this.mode, state: this.state, status: this.failures.length ? "error" : "ok", exit_code: this.exitCode, started_at: this.started, finished_at: now(), duration_ms: Math.floor(performance.now() - this.startedTick), failures: this.failures, counts: this.total, bundle_id: null, cleanup: cleanupFailures.length ? "failed" : "complete" };
    try { this.receipt = { path: "ui-review/terminal.json", ...publishJSON(path.join(this.runRoot, "ui-review/terminal.json"), "receipt", receipt) }; }
    catch { this.state = "failed"; this.receipt = null; this.failures.push(failureRecord(new ReviewFailure("unsafe_artifact"))); this.exitCode ||= 11; }
    try { this.publishLocator(); }
    catch { this.state = "failed"; this.receipt = null; this.failures.push(failureRecord(new ReviewFailure("unsafe_artifact"))); this.exitCode ||= 11; }
    if (!cleanupFailures.length && this.record) unregisterSession(this.record);
  }
}
