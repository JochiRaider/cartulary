import { isDeepStrictEqual } from "node:util";
import { artifact, bundleBase, loadBundle, publishBundle } from "./bundles.mjs";
import { ReviewFailure } from "./contract.mjs";
import { runImageWork } from "./image-worker.mjs";
import { acquireHostAdmission } from "../../runtime/host-admission.mjs";

export function primary(bundle) {
  const ref = bundle.components.actual ?? bundle.components.original;
  if (!ref) throw new ReviewFailure("invalid_artifact");
  return ref;
}
export function compatible(left, right, kind) {
  if (kind === "reference") return;
  const context = (bundle) => {
    const metadata = bundle.source.import_ref?.metadata, fixture = metadata?.fixture;
    if (bundle.source.kind !== "canonical_visual" || !metadata || !fixture?.capture_scope || !Array.isArray(fixture.dynamic_masks) || typeof fixture.no_dynamic_regions !== "boolean" || (fixture.no_dynamic_regions ? fixture.dynamic_masks.length !== 0 : fixture.dynamic_masks.length === 0)) throw new ReviewFailure("invalid_artifact");
    return { renderer: bundle.source.renderer_profile_id, source_kind: bundle.source.kind, runtime: bundle.source.runtime_profile_id, profile: metadata.capture_intent.capture_profile, capture_id: metadata.capture_intent.capture_id, scenario: metadata.capture_intent.scenario_id, scope: fixture.capture_scope, masks: [...fixture.dynamic_masks].sort(), no_dynamic_regions: fixture.no_dynamic_regions };
  };
  if (!isDeepStrictEqual(context(left), context(right))) throw new ReviewFailure("invalid_artifact");
}
export async function execute(session, request, operationID) {
  let lease, primaryFailure;
  try {
    if (!session.hostLease) {
      try { lease = await acquireHostAdmission({ browsers: 0, browserCapacity: session.capacity, signal: session.abort.signal }); }
      catch (cause) { throw new ReviewFailure("capacity_exceeded", { cause }); }
      session.offlineLeases ??= new Set(); session.offlineLeases.add(lease);
    }
    const left = await loadBundle(session, request.bundle_id), source = primary(left.bundle);
    const right = request.comparison ? await loadBundle(session, request.comparison.bundle_id) : null;
    const secondary = right ? primary(right.bundle) : null;
    if (right) compatible(left.bundle, right.bundle, request.comparison.kind);
    const observed = left.bundle.components.observations ? JSON.parse(left.files.get(left.bundle.components.observations.path)) : null;
    const rectangles = observed?.elements.map((entry) => entry.visible_rect).filter(Boolean) ?? [];
    if (request.operations.includes("overlay") && (!rectangles.length || !left.bundle.observation)) throw new ReviewFailure("invalid_artifact");
    const bundle = bundleBase(session, operationID), files = new Map(), mapping = new Map();
    bundle.source = structuredClone(left.bundle.source); bundle.binding = structuredClone(left.bundle.binding); bundle.observation = structuredClone(left.bundle.observation); bundle.limitations = [...left.bundle.limitations];
    const copy = (name, ref) => { const bytes = left.files.get(ref.path); files.set(name, bytes); const copied = artifact(name, bytes, ref.media_type); mapping.set(ref.path, copied); return copied; };
    for (const [kind, ref] of Object.entries(left.bundle.components)) if (ref) bundle.components[kind] = copy(`${kind}.${ref.path.split(".").at(-1)}`, ref);
    if (bundle.source.frontend_receipt) bundle.source.frontend_receipt = copy("frontend-receipt.json", bundle.source.frontend_receipt);
    const parent = (entry) => ({ session_id: session.sessionID, bundle_id: entry.bundle.bundle_id, sha256: entry.sha256 });
    bundle.parents = [parent(left), ...(right && right.bundle.bundle_id !== left.bundle.bundle_id ? [parent(right)] : [])];
    let rightRef;
    if (right) { const bytes = right.files.get(secondary.path); files.set("comparison.png", bytes); rightRef = artifact("comparison.png", bytes, "image/png"); }
    const contactRefs = [left.bundle.components.expected, source, left.bundle.components.diff].filter(Boolean).filter((ref, index, refs) => refs.findIndex((item) => item.path === ref.path) === index);
    const computed = await runImageWork({ operations: request.operations, primary: left.files.get(source.path), secondary: right ? right.files.get(secondary.path) : null, crops: request.crops, rectangles, transform: left.bundle.observation?.coordinate_transform ?? null, contactInputs: contactRefs.map((ref) => left.files.get(ref.path)) }, session.abort.signal);
    const crops = [];
    for (const [index, output] of computed.outputs.entries()) {
      const name = `${output.kind}-${index + 1}.png`; files.set(name, output.bytes); const ref = artifact(name, output.bytes, "image/png");
      const source_refs = output.kind === "contact_sheet" ? [...contactRefs.map((item) => mapping.get(item.path)), ...crops] : output.kind === "exact_diff" ? [mapping.get(source.path), rightRef] : [mapping.get(source.path)];
      bundle.derived.push({ kind: output.kind, ref, source_refs, rectangle: output.rectangle }); if (output.kind === "crop") crops.push(ref);
    }
    const findings = [];
    for (const entry of observed?.elements ?? []) if (entry.overflow_candidate) findings.push({ code: "overflow_candidate", element_ref: entry.resolved_ref, artifact_ref: bundle.components.observations });
    if (observed?.axe.incomplete.length) findings.push({ code: "accessibility_incomplete", element_ref: null, artifact_ref: bundle.components.observations });
    if (computed.comparison?.different_pixels) findings.push({ code: "different_pixels", element_ref: null, artifact_ref: bundle.derived.find((entry) => entry.kind === "exact_diff").ref });
    if (request.comparison?.kind === "reference") bundle.limitations = [...new Set([...bundle.limitations, "cross_source_comparison"])].sort();
    bundle.analysis = { operations: request.operations, comparison: computed.comparison ? { kind: request.comparison.kind, left_bundle_id: left.bundle.bundle_id, right_bundle_id: right.bundle.bundle_id, ...computed.comparison } : null, findings };
    return await publishBundle(session, bundle, files);
  } catch (error) { primaryFailure = error; throw error; }
  finally {
    if (lease) try { await lease.release(); session.offlineLeases.delete(lease); }
    catch (cause) {
      const cleanup = new ReviewFailure("cleanup_failed", { cause });
      if (primaryFailure) primaryFailure.cleanupFailures = [...(primaryFailure.cleanupFailures ?? []), cleanup];
      else throw cleanup;
    }
  }
}
