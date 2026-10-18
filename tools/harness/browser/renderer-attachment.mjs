import { parseStrictJSON, validateSchemaSync, readRetainedArtifactBytes } from "../contract/index.mjs";
import { readLocalFile } from "../runtime/secure-local-files.mjs";
import { admittedBrowserTrust } from "./browser-trust.mjs";
import path from "node:path";
import { canonicalJSONString } from "../contract/index.mjs";

export function readRendererAttachment(environment) {
  const file = environment.CARTULARY_RENDERER_ATTACHMENT;
  if (!file) throw new Error("native renderer borrow attachment required");
  const attachment = parseStrictJSON(readLocalFile(file, { maximum: 1048576 }).toString());
  validateSchemaSync("cartulary.harness_renderer_attachment.v1", attachment);
  if (!path.isAbsolute(file) || file !== path.join(environment.CARTULARY_HARNESS_SUITE_RUNTIME_ROOT ?? "", "renderers",attachment.allocation_id,`borrow-${attachment.borrow_id}.json`) ||
      attachment.attestation_ref !== `_shared/renderers/${attachment.allocation_id}/renderer-profile-attestation.json` ||
      attachment.trust_ref !== `_shared/renderers/${attachment.allocation_id}/tls-trust-attestation.json` ||
      attachment.trust_digest !== attachment.trust.trust_bundle_sha256 ||
      canonicalJSONString(parseStrictJSON(attachment.environment.CARTULARY_VISUAL_RENDERER_PROFILE_JSON)) !== canonicalJSONString(attachment.attestation.profile)) throw new Error("renderer attachment binding mismatch");
  if (attachment.run_id !== environment.CARTULARY_TEST_RUN_ID || attachment.unit_id !== environment.CARTULARY_WORK_UNIT_ID ||
      attachment.runtime_profile_id !== environment.CARTULARY_BROWSER_RUNTIME_PROFILE_ID ||
      attachment.trust_digest !== admittedBrowserTrust(environment).sha256) throw new Error("renderer borrow identity mismatch");
  return { ...attachment, diagnosticBinding: attachment.diagnostic_binding, trustAttestation: attachment.trust, attestation: attachment.attestation, environment: attachment.environment };
}

// Retained readers resolve one physical attestation and the exact logical
// borrow. Private endpoints and connection bindings never enter this evidence.
export function readRendererEvidence(runRoot, group, runID) {
  const borrow = group.renderer_borrow;
  const base = `_shared/renderers/${borrow.allocation_id}`;
  if (borrow.attestation_ref !== `${base}/renderer-profile-attestation.json` || borrow.trust_ref !== `${base}/tls-trust-attestation.json`) throw new Error("foreign renderer evidence reference");
  const read = (relative, schema) => {
    const value = parseStrictJSON(readRetainedArtifactBytes(runRoot, relative, 1048576).toString());
    validateSchemaSync(schema, value);
    return value;
  };
  const event = read(`${base}/borrowed-${borrow.borrow_id}.json`, "cartulary.harness_renderer_event.v1");
  if (event.run_id !== runID || event.state !== "borrowed" || event.unit_id !== borrow.unit_id || event.allocation_id !== borrow.allocation_id || event.generation !== borrow.generation || event.borrow_id !== borrow.borrow_id || event.runtime_profile_id !== group.runtime_profile_id) throw new Error("renderer evidence borrow mismatch");
  const attestation = read(borrow.attestation_ref, "cartulary.frontend_visual_renderer_attestation.v1");
  const trust = read(borrow.trust_ref, "cartulary.browser_tls_trust.v1");
  return { attestation, trust, attestationPath: path.join(runRoot, borrow.attestation_ref) };
}
