import { primaryImage, requireComparable } from "./bundle-model.mjs";
import { artifact, bundleBase } from "./bundles.mjs";
import { ReviewFailure, limits } from "./contract.mjs";
import { computeImages } from "./image-algorithms.mjs";

export async function execute(operation, request, operationID) {
  const left = await operation.store.loadBundle(request.bundle_id), source = primaryImage(left.bundle);
  const right = request.comparison ? await operation.store.loadBundle(request.comparison.bundle_id) : null;
  const secondary = right ? primaryImage(right.bundle) : null;
  if (right) requireComparable(left, right, request.comparison.kind);
  const observed = left.bundle.components.observations ? JSON.parse(left.files.get(left.bundle.components.observations.path)) : null;
  const rectangles = observed?.elements.map((entry) => entry.visible_rect).filter(Boolean) ?? [];
  if (request.operations.includes("overlay") && (!rectangles.length || !left.bundle.observation)) throw new ReviewFailure("invalid_artifact");
  const bundle = bundleBase(operation, operationID), files = new Map(), mapping = new Map();
  bundle.source = structuredClone(left.bundle.source); bundle.binding = structuredClone(left.bundle.binding); bundle.observation = structuredClone(left.bundle.observation); bundle.limitations = [...left.bundle.limitations];
  const copy = (name, ref) => { const bytes = left.files.source(ref.path); files.set(name, bytes); const copied = artifact(name, bytes, ref.media_type); mapping.set(ref.path, copied); return copied; };
  for (const [kind, ref] of Object.entries(left.bundle.components)) if (ref) bundle.components[kind] = copy(`${kind}.${ref.path.split(".").at(-1)}`, ref);
  if (bundle.source.frontend_receipt) bundle.source.frontend_receipt = copy("frontend-receipt.json", bundle.source.frontend_receipt);
  const parent = (entry) => ({ session_id: operation.sessionID, bundle_id: entry.bundle.bundle_id, sha256: entry.sha256 });
  bundle.parents = [parent(left), ...(right && right.bundle.bundle_id !== left.bundle.bundle_id ? [parent(right)] : [])];
  let rightRef;
  if (right) { const bytes = right.files.source(secondary.path); files.set("comparison.png", bytes); rightRef = artifact("comparison.png", bytes, "image/png"); }
  const contactRefs = [left.bundle.components.expected, source, left.bundle.components.diff].filter(Boolean).filter((ref, index, refs) => refs.findIndex((item) => item.path === ref.path) === index);
  const maximumBytes = limits.bundle - [...files.values()].reduce((sum, bytes) => sum + bytes.length, 0);
  if (maximumBytes < 0) throw new ReviewFailure("observation_limit");
  const computed = await computeImages({ operations: request.operations, primary: left.files.get(source.path), secondary: right ? right.files.get(secondary.path) : null, crops: request.crops, rectangles, transform: left.bundle.observation?.coordinate_transform ?? null, contactInputs: contactRefs.map((ref) => left.files.get(ref.path)), maximumBytes });
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
  return await operation.store.publishBundle(bundle, files);
}
