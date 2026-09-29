import { parseStrictJSON } from "../../contract/index.mjs";
import { FilePayload, bytesOf } from "./file-payload.mjs";
import { digest } from "./session-files.mjs";
import { limits, schemaID, validate, ReviewFailure } from "./contract.mjs";
import { validateObservationChannels } from "./bundle-semantics.mjs";

export const components = () => ({ original: null, expected: null, actual: null, diff: null, observations: null, trace: null });
export function artifact(name, bytes, media_type) {
  if (!/^[a-z][a-z0-9_.-]*$/u.test(name)) throw new ReviewFailure("invalid_artifact");
  return { path: name, sha256: bytes instanceof FilePayload ? bytes.sha256 : digest(bytes), bytes: bytes.length, media_type };
}
export function bundleBase(identity, operationID) {
  return { schema_id: schemaID("bundle"), session_id: identity.sessionID, bundle_id: `bundle-${operationID}`, classification: "private_diagnostic", tool_profile: identity.profile, parents: [], source: null, observation: null, binding: null, components: components(), derived: [], analysis: null, limitations: [] };
}
export function references(bundle) {
  return [...Object.values(bundle.components).filter(Boolean), ...bundle.derived.flatMap((entry) => [entry.ref, ...entry.source_refs]), ...(bundle.source.frontend_receipt ? [bundle.source.frontend_receipt] : [])];
}
export function validateBundleBudget(files) {
  const total = [...files.values()].reduce((sum, bytes) => sum + bytes.length, 0);
  if (files.size + 1 > limits.files || total > limits.bundle) throw new ReviewFailure("observation_limit");
  return total;
}
export async function validateFiles(bundle, files) {
  validate("bundle", bundle);
  const unique = new Map();
  for (const ref of references(bundle)) {
    if (unique.has(ref.path) && JSON.stringify(unique.get(ref.path)) !== JSON.stringify(ref)) throw new ReviewFailure("invalid_artifact");
    unique.set(ref.path, ref);
  }
  if (unique.size !== files.size || files.size + 1 > limits.files) throw new ReviewFailure("observation_limit");
  const total = validateBundleBudget(files);
  for (const [name, ref] of unique) {
    const bytes = bytesOf(files.get(name));
    if (!bytes || bytes.length !== ref.bytes || digest(bytes) !== ref.sha256) throw new ReviewFailure("invalid_artifact");
    if (ref.media_type === "image/png") {
      const { decodePNG } = await import("./png.mjs");
      const image = await decodePNG(bytes);
      if (bundle.components.original?.path === name && bundle.observation && (image.width !== bundle.observation.image_dimensions.width || image.height !== bundle.observation.image_dimensions.height)) throw new ReviewFailure("invalid_artifact");
    }
    if (ref.media_type === "application/json") {
      if (bytes.length > limits.component) throw new ReviewFailure("observation_limit");
      const value = parseStrictJSON(new TextDecoder("utf-8", { fatal: true }).decode(bytes));
      if (bundle.components.observations?.path === name) {
        validate("observations", value);
        try { validateObservationChannels(bundle, value); }
        catch (cause) { throw new ReviewFailure("invalid_artifact", { cause }); }
      }
    }
  }
  return total;
}
