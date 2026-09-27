import { mkdirSync, rmSync } from "node:fs";
import path from "node:path";
import { atomicLocalFile, privateDirectory, readLocalFile } from "../../runtime/secure-local-files.mjs";
import { parseStrictJSON } from "../../contract/index.mjs";
import { digest, jsonBytes } from "./session-files.mjs";
import { limits, schemaID, validate, ReviewFailure, emptyCounts } from "./contract.mjs";
import { decodePNG } from "./png.mjs";

export const components = () => ({ original: null, expected: null, actual: null, diff: null, observations: null, trace: null });
export function artifact(name, bytes, media_type) {
  if (!/^[a-z][a-z0-9_.-]*$/u.test(name)) throw new ReviewFailure("invalid_artifact");
  return { path: name, sha256: digest(bytes), bytes: bytes.length, media_type };
}
export function bundleBase(session, operationID) {
  return { schema_id: schemaID("bundle"), session_id: session.sessionID, bundle_id: `bundle-${operationID}`, classification: "private_diagnostic", tool_profile: session.profile, parents: [], source: null, observation: null, binding: null, components: components(), derived: [], analysis: null, limitations: [] };
}
export function references(bundle) {
  return [...Object.values(bundle.components).filter(Boolean), ...bundle.derived.flatMap((entry) => [entry.ref, ...entry.source_refs]), ...(bundle.source.frontend_receipt ? [bundle.source.frontend_receipt] : [])];
}
export async function validateFiles(bundle, files) {
  validate("bundle", bundle);
  const unique = new Map();
  for (const ref of references(bundle)) {
    if (unique.has(ref.path) && JSON.stringify(unique.get(ref.path)) !== JSON.stringify(ref)) throw new ReviewFailure("invalid_artifact");
    unique.set(ref.path, ref);
  }
  if (unique.size !== files.size || files.size + 1 > limits.files) throw new ReviewFailure("observation_limit");
  const total = [...files.values()].reduce((sum, bytes) => sum + bytes.length, 0);
  if (total > limits.bundle) throw new ReviewFailure("observation_limit");
  for (const [name, ref] of unique) {
    const bytes = files.get(name);
    if (!bytes || bytes.length !== ref.bytes || digest(bytes) !== ref.sha256) throw new ReviewFailure("invalid_artifact");
    if (ref.media_type === "image/png") await decodePNG(bytes);
    if (ref.media_type === "application/json") {
      if (bytes.length > limits.component) throw new ReviewFailure("observation_limit");
      const value = parseStrictJSON(new TextDecoder("utf-8", { fatal: true }).decode(bytes));
      if (bundle.components.observations?.path === name) validate("observations", value);
    }
  }
  return total;
}
export async function publishBundle(session, bundle, files) {
  session.abort.signal.throwIfAborted();
  session.bundles ??= new Map();
  if (session.bundles.size >= limits.bundles || session.bundles.has(bundle.bundle_id)) throw new ReviewFailure("capacity_exceeded");
  const size = await validateFiles(bundle, files), manifest = jsonBytes(bundle);
  if (manifest.length > limits.component) throw new ReviewFailure("observation_limit");
  const release = session.reserveBytes(size + manifest.length);
  const root = session.runtime.privatePath("bundles", bundle.bundle_id);
  let owned = false;
  try {
    // mkdir is exclusive; never remove or replace a pre-existing tree. The
    // immutable manifest is the publication commit point and is written last.
    privateDirectory(path.dirname(root));
    mkdirSync(root, { mode: 0o700 }); owned = true;
    for (const [name, bytes] of files) atomicLocalFile(path.join(root, name), bytes);
    session.abort.signal.throwIfAborted();
    atomicLocalFile(path.join(root, "bundle.json"), manifest);
    const identity = { root, sha256: digest(manifest) };
    session.bundles.set(bundle.bundle_id, identity);
    const counts = emptyCounts();
    counts.images = [...files.keys()].filter((name) => name.endsWith(".png")).length;
    if (bundle.components.observations) {
      const observations = JSON.parse(files.get(bundle.components.observations.path));
      counts.observed_elements = observations.elements.length;
      counts.axe_violations = observations.axe.violations.length; counts.axe_incomplete = observations.axe.incomplete.length;
      counts.console_errors = observations.console.records.filter((entry) => entry.level === "error").length;
      counts.failed_requests = observations.network.records.filter((entry) => entry.outcome === "failed").length;
    }
    return { bundle_id: bundle.bundle_id, private_refs: [{ kind: "bundle", absolute_path: path.join(root, "bundle.json") }], counts };
  } catch (error) { release(); if (owned) rmSync(root, { recursive: true, force: true }); throw error; }
}
export async function loadBundle(session, id) {
  const identity = session.bundles?.get(id);
  if (!identity) throw new ReviewFailure("invalid_request");
  try {
    const bytes = readLocalFile(path.join(identity.root, "bundle.json"), { maximum: limits.component });
    if (digest(bytes) !== identity.sha256) throw new ReviewFailure("invalid_artifact");
    const bundle = validate("bundle", parseStrictJSON(new TextDecoder("utf-8", { fatal: true }).decode(bytes)));
    if (bundle.bundle_id !== id || bundle.session_id !== session.sessionID) throw new ReviewFailure("invalid_artifact");
    const files = new Map();
    for (const ref of references(bundle)) {
      if (path.basename(ref.path) !== ref.path) throw new ReviewFailure("invalid_artifact");
      if (!files.has(ref.path)) files.set(ref.path, readLocalFile(path.join(identity.root, ref.path), { maximum: ref.media_type === "image/png" ? limits.png : limits.component }));
    }
    await validateFiles(bundle, files);
    return { bundle, files, sha256: identity.sha256, root: identity.root };
  } catch (cause) { if (cause instanceof ReviewFailure) throw cause; throw new ReviewFailure("invalid_artifact", { cause }); }
}
