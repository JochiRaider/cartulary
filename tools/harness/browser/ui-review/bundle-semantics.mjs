import { freeze } from "./immutable.mjs";
import { fraction } from "./image-math.mjs";

// Semantic machine projection of the adopted source/channel contract. Envelopes
// remain V1; these are impossible-state corrections, not compatibility readers.
export const sourceVariants = freeze({
  sealed_review: Object.freeze({ required: ["original", "observations"], forbidden: ["expected", "actual", "diff", "trace"], observation: true }),
  live_unattested: Object.freeze({ required: ["original", "observations"], forbidden: ["expected", "actual", "diff", "trace"], observation: true }),
  reference_image: Object.freeze({ required: ["original"], forbidden: ["expected", "actual", "diff", "observations", "trace"], observation: false }),
  canonical_visual: Object.freeze({ required: [], forbidden: ["original", "observations"], observation: false }),
});
export function validateBundleSemantics(bundle) {
  const kind = bundle.source.kind, rule = sourceVariants[kind], parts = bundle.components;
  const require = (valid) => { if (!valid) throw new Error("invalid bundle variant"); };
  const token = (name, present) => require(bundle.limitations.includes(name) === present);
  require(Boolean(rule));
  for (const name of rule.required) require(Boolean(parts[name]));
  for (const name of rule.forbidden) require(parts[name] === null);
  require(Boolean(bundle.observation) === rule.observation);
  token("reference_only", kind === "reference_image"); token("live_unattested", kind === "live_unattested");
  token("no_dom", !rule.observation); token("no_trace", parts.trace === null);
  token("rendered_nodes_only", rule.observation);
  token("no_actual", kind === "canonical_visual" && parts.actual === null);
  if (!rule.observation) { token("no_axe", true); token("truncated_console", false); token("truncated_network", false); }
  if (["reference_image", "live_unattested"].includes(kind)) require(bundle.binding === null);
  if (kind === "reference_image") require(parts.original.sha256 === bundle.source.import_ref.input_sha256);
  if (kind === "canonical_visual") {
    require(Boolean(bundle.binding) && Boolean(parts.expected || parts.actual));
    require(!parts.diff || Boolean(parts.expected && parts.actual));
  }
  for (const [name, ref] of Object.entries(parts)) if (ref) require(ref.media_type === (name === "observations" ? "application/json" : name === "trace" ? "application/zip" : "image/png"));
  if (bundle.source.frontend_receipt) require(bundle.source.frontend_receipt.media_type === "application/json");
  require(new Set(bundle.parents.map((parent) => parent.bundle_id)).size === bundle.parents.length);
  for (const parent of bundle.parents) require(parent.session_id === bundle.session_id && parent.bundle_id !== bundle.bundle_id);
  token("cross_source_comparison", bundle.analysis?.comparison?.kind === "reference");
  if (bundle.analysis === null) require(bundle.parents.length === 0 && bundle.derived.length === 0);
  else {
    const { operations, comparison } = bundle.analysis;
    require(operations.length > 0 && bundle.parents.length > 0);
    require(operations.every((op, index) => index === 0 || op > operations[index - 1]));
    require(operations.includes("exact_diff") === Boolean(comparison));
    for (const operation of operations) require(bundle.derived.some((entry) => entry.kind === operation));
    for (const derived of bundle.derived) {
      require(operations.includes(derived.kind) && derived.ref.media_type === "image/png");
      require(derived.source_refs.length > 0 && derived.source_refs.every((ref) => ref.media_type === "image/png" && ref.path !== derived.ref.path));
    }
    for (const operation of ["overlay", "contact_sheet", "exact_diff"]) require(bundle.derived.filter((entry) => entry.kind === operation).length <= 1);
    require(bundle.derived.filter((entry) => entry.kind === "crop").length <= 16);
    if (operations.includes("overlay")) require(rule.observation);
    if (comparison) {
      require(bundle.parents.some((parent) => parent.bundle_id === comparison.left_bundle_id));
      require(bundle.parents.some((parent) => parent.bundle_id === comparison.right_bundle_id));
      require(comparison.total_pixels === comparison.width * comparison.height && comparison.different_pixels <= comparison.total_pixels && comparison.different_fraction === fraction(comparison.different_pixels, comparison.total_pixels));
    }
  }
}

export function validateObservationChannels(bundle, observed) {
  const require = (valid) => { if (!valid) throw new Error("observation channel mismatch"); };
  require(bundle.limitations.includes("no_axe") === (observed.axe.status !== "completed"));
  require(bundle.limitations.includes("truncated_console") === observed.console.truncated);
  require(bundle.limitations.includes("truncated_network") === observed.network.truncated);
}
