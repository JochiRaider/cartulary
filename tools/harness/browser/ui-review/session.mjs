import { randomBytes } from "node:crypto";
import { accessSync, constants } from "node:fs";
import path from "node:path";
import { createSuiteRuntime, scanRetainedRoot } from "../../runtime/suite-runtime.mjs";
import { acquireHostAdmission, inheritedHostLease } from "../../runtime/host-admission.mjs";
import { readLocalFile } from "../../runtime/secure-local-files.mjs";
import { captureCapabilitySnapshot, resourceCapacities } from "../../scheduler/work-graph/capability.mjs";
import { buildSourceSnapshot } from "../../test-catalog/source-snapshot.mjs";
import { prepareReview } from "./preparation.mjs";
import { boundedCleanup, purgeReviewDetail, recordResource, recoveryResources, stopOwnedProcess } from "./ownership.mjs";
import { ReviewBrowser, browserReady } from "./browser.mjs";
import { commandID, emptyCounts, failureRecord, limits, result, ReviewFailure, schemaID } from "./contract.mjs";
import { newRunRoot, publishJSON, registerSession, updateSessionRecord } from "./session-files.mjs";
import { finishTerminal } from "./terminal.mjs";
import { ArtifactStore } from "./artifact-store.mjs";
import { repoRoot } from "./policy.mjs";

const now = () => new Date().toISOString();
const pause = (milliseconds) => new Promise((resolve) => setTimeout(resolve, milliseconds));
export class ReviewSession {
  constructor(input, profile) {
    this.input = input; this.profile = profile; this.mode = input.UI_MODE;
    this.operationID = 0; this.total = emptyCounts(); this.started = now(); this.startedTick = performance.now();
    this.sessionID = `uireview-${randomBytes(16).toString("hex")}`; this.state = "preparing";
    this.failures = []; this.exitCode = 0; this.abort = new AbortController(); this.receipt = null;
    this.store = new ArtifactStore({ sessionID: this.sessionID, profile, signal: this.abort.signal, privatePath: (...parts) => this.runtime.privatePath(...parts) });
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
    if (this.preparing) return this.preparing;
    this.preparing = this.prepareMode();
    return this.preparing;
  }
  async prepareMode() {
    this.capacities = Object.fromEntries(resourceCapacities(captureCapabilitySnapshot({ root: repoRoot })));
    this.parentLease = inheritedHostLease();
    if (this.mode !== "artifacts") {
      try { this.hostLease = await acquireHostAdmission({ claims: { browser_stack: 1 }, capacities: this.capacities, parent: this.parentLease, signal: this.abort.signal }); }
      catch (cause) { throw new ReviewFailure("capacity_exceeded", { cause }); }
      this.workspaceDigest = buildSourceSnapshot(repoRoot).digest.replace(/^sha256:/u, "");
      let origin = this.input.UI_ORIGIN;
      let actors = {};
      if (this.mode === "seeded") {
        this.preparedOwner = prepareReview({ input: this.input, runtime: this.runtime, runID: this.runID, signal: this.abort.signal, onProcess: ({ boot, pid, start }) => this.hostLease.bind({ boot, pid, start }) });
        this.preparation = this.preparedOwner.done;
        const prepared = await this.preparedOwner.ready;
        this.abort.signal.throwIfAborted();
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
      this.abort.signal.throwIfAborted();
      this.browser = new ReviewBrowser({ origin, mode: this.mode, actors, onOwnedResource: async (resource) => { recordResource(this.runtime, resource); if (resource.state !== "released") { const { boot, pid, start } = resource.target; await this.hostLease.bind({ boot, pid, start }); } }, onLost: (error) => { void this.stop(error); } });
      await this.browser.start();
    }
    this.abort.signal.throwIfAborted(); this.state = "ready"; this.publishLocator();
    this.expiry = setTimeout(() => { void this.stop(); }, limits.lifetime);
  }
  value(command, fields = {}) {
    return result(command, { session_id: this.sessionID, state: this.state, epoch: ["closed", "failed"].includes(this.state) ? null : this.browser?.epoch ?? null, ...fields });
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
        this.abort.signal.throwIfAborted();
        const private_refs = await this.store.observations(operationID, observed);
        counts.observed_elements = observed.elements.length;
        counts.console_errors = observed.console.records.filter((record) => record.level === "error").length;
        counts.failed_requests = observed.network.records.filter((record) => record.outcome === "failed").length;
        outcome = this.value(command, { operation_id: operationID, private_refs });
      } else {
        const { executeWork } = await import("./executor.mjs");
        const produced = await executeWork({ store: this.store, identity: { sessionID: this.sessionID, profile: this.profile }, onResource: (resource) => recordResource(this.runtime, resource), signal: this.abort.signal, capacities: this.capacities, parentLease: this.hostLease?.token ?? this.parentLease, deadline: tick + limits.operation, command, request: request ?? bundleID, operationID,
          observe: async (staging) => {
            const { pageSource } = await import("./source.mjs");
            const source = pageSource({ mode: this.mode, profile: this.input.REVIEW_PROFILE, workspaceDigest: this.workspaceDigest, runID: this.runID, prepared: this.seeded, browserVersion: this.browser.version() }, request.binding);
            return this.browser.capture(request, { identity: { sessionID: this.sessionID, profile: this.profile }, source }, operationID, staging);
          } });
        this.abort.signal.throwIfAborted();
        counts = produced.counts ?? counts;
        outcome = this.value(command, { operation_id: operationID, bundle_id: produced.bundle_id, private_refs: produced.private_refs });
      }
    } catch (error) {
      const failure = this.abort.signal.aborted ? this.abort.signal.reason : error;
      outcome = this.fail(command, failure, { operation_id: operationID });
      if (error.cleanupFailures?.length) outcome.failures.push(failureRecord(new ReviewFailure("cleanup_failed")));
      if (error.cleanupFailures?.length) queueMicrotask(() => { void this.stop(failure instanceof ReviewFailure ? failure : new ReviewFailure("unsafe_artifact")); });
    }
    if (this.abort.signal.aborted && outcome.status === "ok") outcome = this.fail(command, this.abort.signal.reason, { operation_id: operationID });
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
    const attempt = async (release, timeout = 30000) => { try { await boundedCleanup(release, timeout); } catch (error) { cleanupFailures.push(error); } };
    await attempt(() => this.publishLocator());
    this.abort.abort(this.stopCause ?? new ReviewFailure("interrupted"));
    const preparedStop = this.preparedOwner?.stop(); preparedStop?.catch(() => {});
    await attempt(async () => { try { await this.preparing; } catch (error) { if (!this.abort.signal.aborted) throw error; } }, 250000);
    await attempt(() => this.browser?.close());
    try { await boundedCleanup(() => preparedStop ?? this.preparation, 250000); } catch (error) {
      // Acquisition/action failure already belongs to the primary lifecycle
      // outcome. Only failures of the preparation owner's release are secondary.
      cleanupFailures.push(...(error.cleanupFailures ?? (error.diagnostic === "cleanup_failed" ? [error] : [])));
      if (!this.failures.length && !this.abort.signal.aborted) { this.failures.push(failureRecord(new ReviewFailure("startup_failed"))); this.exitCode ||= 3; }
    }
    await attempt(async () => { await this.active; });
    await attempt(async () => {
      for (const resource of recoveryResources(this.runtime).filter((entry) => entry.kind.endsWith("_process"))) {
        await attempt(async () => { await stopOwnedProcess(resource.target); recordResource(this.runtime, { ...resource, state: "released" }); });
      }
    });
    await attempt(() => this.hostLease?.release());
    await attempt(() => this.store.close());
    await attempt(() => purgeReviewDetail(this.runtime));
    await attempt(async () => {
      const scan = await scanRetainedRoot(this.runRoot, { forbiddenValues: this.runtime.forbiddenValues(), removeUnsafe: true });
      if (scan.status !== "pass") throw new ReviewFailure("unsafe_artifact");
    });
    if (cleanupFailures.length === 0) await attempt(() => {
      // Keep minimum controller proof outside the private tree until terminal
      // publication succeeds, including when the tree is already gone.
      updateSessionRecord(this.record, { resources_released: true });
      return this.runtime.close();
    });
    if (cleanupFailures.length) {
      const artifact = cleanupFailures.some((failure) => failure instanceof ReviewFailure && failure.diagnostic === "unsafe_artifact");
      const failure = new ReviewFailure(artifact ? "unsafe_artifact" : "cleanup_failed");
      this.failures.push(failureRecord(failure)); this.exitCode ||= failure.exitCode;
    }
    Object.assign(this, await finishTerminal({ record: this.record, locator: { schema_id: schemaID("session"), session_id: this.sessionID, run_id: this.runID, mode: this.mode, created_at: this.started }, runRoot: this.runRoot, counts: this.total, failures: this.failures, exitCode: this.exitCode, duration: Math.floor(performance.now() - this.startedTick), cleanupFailed: cleanupFailures.length > 0, forbiddenValues: this.runtime.forbiddenValues() }));
  }
}
