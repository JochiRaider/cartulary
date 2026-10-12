// Private browser/artifact producer. The native coordinator is the sole writer
// of session state, operation receipts and terminal evidence.
import { createHash } from "node:crypto";
import { readFileSync } from "node:fs";
import { execFileSync } from "../../workspace/child-process.mjs";
import net from "node:net";
import path from "node:path";
import { parseStrictJSON, validateSchemaSync } from "../../contract/index.mjs";
import { borrowRuntime } from "../../runtime/session-client.mjs";
import { readLocalFile } from "../../runtime/secure-local-files.mjs";
import { buildSourceSnapshot } from "../../test-catalog/index.mjs";
import { captureCapabilitySnapshot, resourceCapacities } from "../../scheduler/work-graph/capability.mjs";
import { buildReview, seedReview, checkReview, recipeEnvironment } from "../review-recipes.mjs";
import { ReviewBrowser, browserReady } from "./browser.mjs";
import { ArtifactStore } from "./artifact-store.mjs";
import { limits, emptyCounts, ReviewFailure, failureRecord, parseRequest } from "./contract.mjs";
import { preparationFailure } from "./failure.mjs";
import { repoRoot } from "./policy.mjs";

export class ReviewAdapter {
  constructor(identity) { this.identity = identity; this.abort = new AbortController(); this.parentLease = null; }
  async initialize({ input, profile, repository, run_root, runtime_root }) {
    if (repository !== repoRoot || runtime_root !== process.env.CARTULARY_HARNESS_SUITE_RUNTIME_ROOT) throw new ReviewFailure("invalid_request");
    this.input = input; this.profile = profile; this.runRoot = run_root;
    this.runtime = borrowRuntime({ repoRoot, runRoot: run_root, admissionToken: () => this.parentLease });
    if (input.mode !== "artifacts" && input.presentation === "ui") browserReady();
    if (input.presentation === "ui") this.store = new ArtifactStore({ sessionID: this.identity.session_id, profile, signal: this.abort.signal, privatePath: (...parts) => this.runtime.privatePath(...parts) });
    this.capacities = Object.fromEntries(resourceCapacities(captureCapabilitySnapshot({ root: repoRoot })));
    this.sourceDigest = buildSourceSnapshot(repoRoot).digest;
    this.preparationRoot = input.presentation === "design" ? run_root : this.runtime.privatePath("lifecycle", this.runtime.runID);
    this.sourceCommit = execFileSync("git", ["rev-parse", "HEAD"], { cwd: repoRoot, encoding: "utf8" }).trim();
    this.toolchainDigest = `sha256:${createHash("sha256").update(readFileSync(path.join(repoRoot, "tools/toolchain_pins.json"))).digest("hex")}`;
    this.environment = recipeEnvironment(repoRoot, this.preparationRoot, this.runtime, input);
    return this.facts();
  }
  facts() { return { epoch: this.browser?.epoch ?? null, capacities: this.capacities, source_digest: this.sourceDigest, source_commit: this.sourceCommit, toolchain_digest: this.toolchainDigest, environment: this.environment, review: this.seeded ? { public_origin: this.seeded.attached.CARTULARY_WEB_E2E_PUBLIC_ORIGIN, private_directory: this.seeded.privateDirectory, samples: this.seeded.review.samples, scenarios: this.seeded.review.scenarios } : null }; }
  async prepareBrowser({ admission_token }) {
    this.parentLease = admission_token;
    if (this.input.mode === "artifacts") return;
    let origin = this.input.origin, actors = {};
    if (this.input.mode === "seeded") {
      origin = this.seeded.attached.CARTULARY_WEB_E2E_PUBLIC_ORIGIN;
      const accounts = JSON.parse(readLocalFile(path.join(this.seeded.privateDirectory, "access.json"))).accounts;
      actors = { admin: accounts[0], empty: accounts[1], viewer: accounts[2], editor: accounts[3] };
    }
    this.browser = new ReviewBrowser({ origin, mode: this.input.mode, actors, environment: this.seeded?.attached,
      onOwnedResource: (resource) => this.runtime.record(resource), onLost: () => this.onLost?.() });
    try { await this.browser.start(); }
    catch (error) { throw preparationFailure(error, { phase: "browser_start", subject_id: "browser", condition: "child_failed", recovery_id: "inspect_failure" }); }
  }
  validate(command, request) {
    const kind = { "ui-browser": "action", "ui-capture": "capture_request", "ui-analyze": "analysis_request" }[command];
    if (kind) request = parseRequest(Buffer.from(JSON.stringify(request)), kind);
    if (command === "ui-browser") { if (!this.browser) throw new ReviewFailure("invalid_request"); this.browser.validateAction(request); }
    if (command === "ui-capture" && request.source === "page") {
      if (!this.browser) throw new ReviewFailure("invalid_request");
      if (request.expected_epoch !== this.browser.epoch || this.browser.needsSnapshot) throw new ReviewFailure("session_mismatch");
    }
    return request;
  }
  async execute({ command, request, bundle_id, operation_id }) {
    request = this.validate(command, request);
    this.abort.signal.throwIfAborted();
    if (command === "ui-browser") {
      if (request.action === "fill") this.runtime.registerSecret(request.parameters.text);
      let observed;
      if (request.action === "diagnostic_snapshot") {
        const { diagnosticSnapshot } = await import("./diagnostic.mjs");
        observed = await diagnosticSnapshot({ owner: this.browser, runtime: this.runtime, operationID: operation_id, signal: this.abort.signal,
          capacities: this.capacities, parentLease: this.parentLease, onResource: (resource) => this.runtime.record(resource) });
      } else observed = await this.browser.action(request);
      this.abort.signal.throwIfAborted();
      const private_refs = await this.store.observations(operation_id, observed), counts = emptyCounts();
      counts.observed_elements = observed.elements.length; counts.console_errors = observed.console.records.filter((r) => r.level === "error").length;
      counts.failed_requests = observed.network.records.filter((r) => r.outcome === "failed").length;
      return { counts, private_refs };
    }
    const { executeWork } = await import("./executor.mjs");
    return executeWork({ store: this.store, identity: { sessionID: this.identity.session_id, profile: this.profile }, onResource: (resource) => this.runtime.record(resource), signal: this.abort.signal,
      capacities: this.capacities, parentLease: this.parentLease, command, request: request ?? bundle_id, operationID: operation_id,
      observe: async (staging) => {
        const { pageSource } = await import("./source.mjs");
        const source = pageSource({ mode: this.input.mode, profile: this.input.profile, workspaceDigest: this.sourceDigest.replace(/^sha256:/u, ""), runID: this.runtime.runID, prepared: this.seeded, browserVersion: this.browser.version() }, request.binding);
        return this.browser.capture(request, { identity: { sessionID: this.identity.session_id, profile: this.profile }, source }, operation_id, staging);
      } });
  }
  async call(operation, payload) {
    switch (operation) {
      case "initialize": return { facts: await this.initialize(payload) };
      case "build": this.parentLease = payload.admission_token; this.environment.CARTULARY_HOST_ADMISSION_LEASE = this.parentLease; await buildReview({ root: repoRoot, environment: this.environment, input: this.input, runtime: this.runtime, signal: this.abort.signal }); return { facts: this.facts() };
      case "seed": this.seeded = await seedReview({ root: repoRoot, runRoot: this.preparationRoot, runtime: this.runtime, environment: payload.environment, input: this.input, sourceDigest: this.sourceDigest, signal: this.abort.signal }); return { facts: this.facts() };
      case "prepare_browser": await this.prepareBrowser(payload); return {};
      case "validate": this.validate(payload.command, payload.request); return {};
      case "execute": return { outcome: await this.execute(payload) };
      case "abort": this.abort.abort(new ReviewFailure("interrupted")); return {};
      case "close_browser": this.abort.abort(new ReviewFailure("interrupted")); await this.browser?.close(); return {};
      case "close_store": this.store?.close(); return {};
      case "check": if (this.seeded) await checkReview({ root: repoRoot, runtime: this.runtime, environment: this.seeded.attached, signal: this.abort.signal }); return {};
      case "close": this.runtime?.close(); return {};
      default: throw new ReviewFailure("invalid_request");
    }
  }
}

async function main() {
  process.umask(0o077);
  const schema = "cartulary.harness_review_adapter.v1";
  const identity = { session_id: process.env.CARTULARY_REVIEW_SESSION_ID, run_id: process.env.CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID, runtime_lease_id: process.env.CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID };
  const channel = new net.Socket({ fd: 3, readable: true, writable: true });
  const adapter = new ReviewAdapter(identity); let buffered = Buffer.alloc(0), closing = false, sequence = 0;
  const send = (message) => { validateSchemaSync(schema, message); const bytes = Buffer.from(`${JSON.stringify(message)}\n`); if (bytes.length > limits.frame) throw new ReviewFailure("invalid_artifact"); channel.write(bytes); };
  adapter.onLost = () => send({ schema_id: schema, ...identity, request_id: 0, kind: "event", operation: "lost", failure: failureRecord(new ReviewFailure("session_lost")) });
  channel.on("error", () => process.exit(15)); channel.on("end", () => process.exit(15));
  process.on("SIGTERM", () => adapter.abort.abort(new ReviewFailure("interrupted"))); process.on("SIGINT", () => adapter.abort.abort(new ReviewFailure("interrupted")));
  channel.on("data", (part) => {
    buffered = Buffer.concat([buffered, part]);
    while (true) {
      const at = buffered.indexOf(10); if (at < 0) { if (buffered.length >= limits.frame) process.exit(11); return; }
      if (at + 1 > limits.frame || closing) { process.exit(11); return; }
      const bytes = buffered.subarray(0, at); buffered = buffered.subarray(at + 1);
      let message;
      try { message = parseStrictJSON(new TextDecoder("utf-8", { fatal: true }).decode(bytes)); validateSchemaSync(schema, message);
        if (message.kind !== "request" || Object.entries(identity).some(([key, value]) => message[key] !== value)) throw new Error("binding"); }
      catch { process.exit(11); return; }
      if (message.operation !== "abort") { if (message.request_id <= sequence) { process.exit(11); return; } sequence = message.request_id; }
      if (message.operation === "close") closing = true;
      void (async () => {
        let response = { facts: null }, outcome = { epoch: null, counts: emptyCounts(), bundle_id: null, private_refs: [], failures: [], exit_code: 0, stop: false };
        try { response = await adapter.call(message.operation, message.payload); Object.assign(outcome, response.outcome); }
        catch (error) {
          if (message.operation === "execute" && adapter.abort.signal.aborted) error = adapter.abort.signal.reason;
          const normalized = ["initialize", "build", "seed", "prepare_browser"].includes(message.operation) ? preparationFailure(error) : error instanceof ReviewFailure ? error : new ReviewFailure("unsafe_artifact");
          outcome.failures = [failureRecord(normalized)]; outcome.exit_code = normalized.exitCode;
          if (error.cleanupFailures?.length) { outcome.failures.push(failureRecord(new ReviewFailure("cleanup_failed"))); outcome.stop = true; }
        }
        outcome.epoch = adapter.browser?.epoch ?? null;
        try { send({ schema_id: schema, ...identity, request_id: message.request_id, operation: message.operation, kind: "response", facts: response.facts ?? null, outcome });
          if (message.operation === "close") channel.end(() => process.exit(0)); }
        catch { process.exit(11); }
      })();
    }
  });
}
if (process.argv[2] === "--private-adapter") await main();
