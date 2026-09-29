import { mkdirSync } from "node:fs";
import path from "node:path";
import { isDeepStrictEqual } from "node:util";
import { atomicLocalFile, privateDirectory, readLocalFile, removePrivateTree } from "../../runtime/secure-local-files.mjs";
import { parseStrictJSON } from "../../contract/index.mjs";
import { freeze } from "./immutable.mjs";
import { canonicalComparison } from "./canonical-import.mjs";
import { FilePayload, StoredFiles, bytesOf } from "./file-payload.mjs";
import { artifact, references, validateFiles } from "./bundles.mjs";
import { emptyCounts, limits, ReviewFailure, validate } from "./contract.mjs";
import { digest, jsonBytes } from "./session-files.mjs";

const safeName = (name) => /^[a-z][a-z0-9_.-]*$/u.test(name);

/** One owner for private output transactions, indexes and residual accounting.
 * Only a validated final commit file makes a transaction consumable. The IO
 * port permits boundary fault tests without weakening production path checks.
 */
export class ArtifactStore {
  #sessionID; #profile; #privatePath; #signal; #capacity; #io;
  #bundles = new Map(); #reports = new Map(); #reservations = new Map(); #pendingBundles = new Set(); #bytes = 0;
  constructor({ sessionID, profile, privatePath, signal, capacity = limits.storage, io = {}, index }) {
    this.#sessionID = sessionID; this.#profile = profile; this.#privatePath = privatePath; this.#signal = signal;
    if (!Number.isSafeInteger(capacity) || capacity < 1 || capacity > limits.storage) throw new ReviewFailure("invalid_request");
    this.#capacity = capacity;
    this.#io = { write: atomicLocalFile, remove: removePrivateTree, ...io };
    if (index) { this.#bundles = new Map(index.bundles); this.#reports = new Map(index.reports); }
  }
  get usage() { return Object.freeze({ bytes: this.#bytes, bundles: this.#bundles.size, reservations: this.#reservations.size }); }
  workIndex() { return { bundles: [...this.#bundles], reports: [...this.#reports] }; }
  workResult() { return { index: this.workIndex(), bytes: this.#bytes }; }
  beginWork(operationID) {
    this.#check();
    const bytes = this.#capacity - this.#bytes;
    if (bytes < 1) throw new ReviewFailure("capacity_exceeded");
    const root = this.#privatePath("artifacts", "jobs", `operation-${operationID}`);
    if (this.#reservations.has(root)) throw new ReviewFailure("invalid_request");
    privateDirectory(path.dirname(root)); mkdirSync(root, { mode: 0o700 });
    const reservation = { root, bytes, owned: true }; this.#reservations.set(root, reservation); this.#bytes += bytes;
    let closed = false;
    return {
      root, capacity: bytes, index: this.workIndex(),
      commit: (result) => {
        this.#check();
        if (closed || !Number.isSafeInteger(result.bytes) || result.bytes < 0 || result.bytes > bytes) throw new ReviewFailure("invalid_artifact");
        const bundles = new Map(result.index.bundles), reports = new Map(result.index.reports);
        for (const [id, identity] of this.#bundles) if (!isDeepStrictEqual(identity, bundles.get(id))) throw new ReviewFailure("invalid_artifact");
        for (const [id, identity] of this.#reports) if (!isDeepStrictEqual(identity, reports.get(id))) throw new ReviewFailure("invalid_artifact");
        for (const [id, identity] of [...bundles, ...reports]) {
          if (this.#bundles.has(id) || this.#reports.has(id)) continue;
          const relative = path.relative(root, identity.root);
          if (!relative || relative.startsWith("..") || path.isAbsolute(relative)) throw new ReviewFailure("invalid_artifact");
        }
        if (result.bytes === 0) { this.#io.remove(root); this.#release(root); }
        else { this.#bytes -= bytes - result.bytes; reservation.bytes = result.bytes; }
        this.#bundles = bundles; this.#reports = reports; closed = true;
      },
      rollback: () => {
        if (closed) return;
        try { this.#io.remove(reservation.root); this.#release(root); closed = true; }
        catch (error) { if (error.cleanupPath) reservation.root = error.cleanupPath; throw error; }
      },
    };
  }
  #check() { this.#signal.throwIfAborted(); }
  async #publish(collection, key, files, commitName, check) {
    this.#check();
    if (!safeName(key) || !files.has(commitName) || [...files.keys()].some((name) => !safeName(name))) throw new ReviewFailure("invalid_artifact");
    const root = this.#privatePath("artifacts", collection, key);
    const bytes = [...files.values()].reduce((sum, value) => sum + value.length, 0);
    if (!Number.isSafeInteger(bytes) || this.#bytes + bytes > this.#capacity || this.#reservations.has(root)) throw new ReviewFailure("capacity_exceeded");
    const reservation = { root, bytes, owned: false }; this.#reservations.set(root, reservation); this.#bytes += bytes;
    try {
      privateDirectory(path.dirname(root)); mkdirSync(root, { mode: 0o700 }); reservation.owned = true;
      for (const [name, value] of files) if (name !== commitName) { this.#check(); this.#io.write(path.join(root, name), value); }
      await check(root);
      this.#check(); this.#io.write(path.join(root, commitName), files.get(commitName));
      return { root, files: [...files].map(([name, value]) => artifact(name, value, "application/octet-stream")) };
    } catch (primary) {
      try { if (reservation.owned) this.#io.remove(reservation.root); this.#release(root); }
      catch (secondary) {
        if (secondary.cleanupPath) reservation.root = secondary.cleanupPath;
        (primary.cleanupFailures ??= []).push(new ReviewFailure("cleanup_failed", { cause: secondary }));
      }
      throw primary;
    }
  }
  #release(key) { const held = this.#reservations.get(key); if (held) { this.#bytes -= held.bytes; this.#reservations.delete(key); } }
  #checkParents(bundle) {
    if (bundle.session_id !== this.#sessionID || !isDeepStrictEqual(bundle.tool_profile, this.#profile)) throw new ReviewFailure("invalid_artifact");
    for (const parent of bundle.parents) {
      const stored = this.#bundles.get(parent.bundle_id);
      if (parent.session_id !== this.#sessionID || parent.bundle_id === bundle.bundle_id || !stored || parent.sha256 !== stored.sha256) throw new ReviewFailure("invalid_artifact");
      const bytes = readLocalFile(path.join(stored.root, "bundle.json"), { maximum: limits.component });
      if (digest(bytes) !== parent.sha256) throw new ReviewFailure("invalid_artifact");
      if (parent === bundle.parents[0]) {
        const left = JSON.parse(bytes);
        for (const key of ["source", "binding", "observation"]) if (!isDeepStrictEqual(bundle[key], left[key])) throw new ReviewFailure("invalid_artifact");
        for (const [key, ref] of Object.entries(left.components)) {
          const copy = bundle.components[key];
          if (Boolean(ref) !== Boolean(copy) || (ref && (ref.sha256 !== copy.sha256 || ref.bytes !== copy.bytes || ref.media_type !== copy.media_type))) throw new ReviewFailure("invalid_artifact");
        }
      }
    }
  }
  async publishBundle(input, files) {
    this.#check();
    if (this.#bundles.size + this.#pendingBundles.size >= limits.bundles || this.#bundles.has(input.bundle_id) || this.#pendingBundles.has(input.bundle_id)) throw new ReviewFailure("capacity_exceeded");
    const bundle = freeze(structuredClone(input)); files = new Map(files);
    validate("bundle", bundle); this.#checkParents(bundle);
    const manifest = jsonBytes(bundle);
    if (manifest.length > limits.component || files.has("bundle.json")) throw new ReviewFailure("observation_limit");
    this.#pendingBundles.add(bundle.bundle_id);
    try {
      await validateFiles(bundle, files);
      const staged = new Map(files); staged.set("bundle.json", manifest);
      const identity = await this.#publish("bundles", bundle.bundle_id, staged, "bundle.json", async (root) => {
        // Validate the bytes actually staged, not just producer buffers that can
        // change while asynchronous validation is running.
        const copied = this.#readFiles(root, bundle); await validateFiles(bundle, copied);
        for (const parent of bundle.parents) await this.loadBundle(parent.bundle_id);
      });
      identity.sha256 = digest(manifest); this.#bundles.set(bundle.bundle_id, identity);
      const counts = emptyCounts(); counts.images = [...files.keys()].filter((name) => name.endsWith(".png")).length;
      if (bundle.components.observations) {
        const observations = JSON.parse(bytesOf(files.get(bundle.components.observations.path)));
        counts.observed_elements = observations.elements.length;
        counts.axe_violations = observations.axe.violations.length; counts.axe_incomplete = observations.axe.incomplete.length;
        counts.console_errors = observations.console.records.filter((entry) => entry.level === "error").length;
        counts.failed_requests = observations.network.records.filter((entry) => entry.outcome === "failed").length;
      }
      return { bundle_id: bundle.bundle_id, private_refs: [{ kind: "bundle", absolute_path: path.join(identity.root, "bundle.json") }], counts };
    } finally { this.#pendingBundles.delete(bundle.bundle_id); }
  }
  #readFiles(root, bundle) {
    const files = new StoredFiles();
    for (const ref of references(bundle)) {
      if (!safeName(ref.path)) throw new ReviewFailure("invalid_artifact");
      if (!files.has(ref.path)) files.set(ref.path, new FilePayload(path.join(root, ref.path), ref));
    }
    return files;
  }
  async loadBundle(id) {
    this.#check(); const identity = this.#bundles.get(id);
    if (!identity) throw new ReviewFailure("invalid_request");
    try {
      const bytes = readLocalFile(path.join(identity.root, "bundle.json"), { maximum: limits.component });
      if (digest(bytes) !== identity.sha256) throw new ReviewFailure("invalid_artifact");
      const bundle = validate("bundle", parseStrictJSON(new TextDecoder("utf-8", { fatal: true }).decode(bytes)));
      if (bundle.bundle_id !== id) throw new ReviewFailure("invalid_artifact");
      this.#checkParents(bundle);
      const files = this.#readFiles(identity.root, bundle); await validateFiles(bundle, files);
      return Object.freeze({ bundle: freeze(bundle), comparisonIdentity: canonicalComparison(bundle.source), files, sha256: identity.sha256, root: identity.root });
    } catch (cause) { if (cause instanceof ReviewFailure) throw cause; throw new ReviewFailure("invalid_artifact", { cause }); }
  }
  async observations(operationID, observations) {
    validate("observations", observations); const bytes = jsonBytes(observations);
    if (bytes.length > limits.component) throw new ReviewFailure("observation_limit");
    const entry = await this.#publish("observations", `operation-${operationID}`, new Map([["observations.json", bytes]]), "observations.json", () => {});
    return [{ kind: "observations", absolute_path: path.join(entry.root, "observations.json") }];
  }
  async readModel(id) {
    const entry = await this.loadBundle(id);
    const selected = new Map(references(entry.bundle).map((ref) => [ref.path, ref]));
    const asset = (ref) => {
      if (!isDeepStrictEqual(selected.get(ref.path), ref)) throw new ReviewFailure("invalid_artifact");
      return entry.files.source(ref.path);
    };
    return Object.freeze({ bundle: entry.bundle, sha256: entry.sha256, references: Object.freeze([...selected.values()]),
      read: (ref) => bytesOf(asset(ref)), asset,
    });
  }
  async report(id, version, render) {
    const entry = await this.readModel(id), key = `${entry.sha256}-v${version}`;
    // Prefix keeps all storage keys in the same narrow safe-name grammar.
    const storageKey = `report-${key}`, cached = this.#reports.get(key);
    if (cached) {
      for (const ref of cached.files) {
        const bytes = readLocalFile(path.join(cached.root, ref.path), { maximum: ref.bytes });
        if (bytes.length !== ref.bytes || digest(bytes) !== ref.sha256) throw new ReviewFailure("invalid_artifact");
      }
      return { bundle_id: id, private_refs: [{ kind: "report", absolute_path: path.join(cached.root, "index.html") }] };
    }
    const files = await render(entry);
    if (!files.has("index.html") || files.get("index.html").length > limits.report) throw new ReviewFailure("observation_limit");
    const record = await this.#publish("reports", storageKey, files, "index.html", (root) => {
      for (const [name, bytes] of files) if (name !== "index.html" && digest(readLocalFile(path.join(root, name), { maximum: bytes.length })) !== (bytes instanceof FilePayload ? bytes.sha256 : digest(bytes))) throw new ReviewFailure("invalid_artifact");
    });
    this.#reports.set(key, record);
    return { bundle_id: id, private_refs: [{ kind: "report", absolute_path: path.join(record.root, "index.html") }] };
  }
  close() {
    const failures = [];
    for (const [key, reservation] of this.#reservations) {
      try { this.#io.remove(reservation.root); this.#release(key); }
      catch (error) { if (error.cleanupPath) reservation.root = error.cleanupPath; failures.push(error); }
    }
    if (failures.length) throw new AggregateError(failures, "artifact disposal failed");
    this.#bundles.clear(); this.#reports.clear();
  }
}
