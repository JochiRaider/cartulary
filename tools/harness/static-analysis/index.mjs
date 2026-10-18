import { buildResolvedFallowConfig } from "./fallow-reachability.mjs";

// Configuration integrity is blocking source validation. Report execution is
// a separate operation and never occurs while resolving this read-only fact.
export function validateStaticConfiguration(root) {
  buildResolvedFallowConfig({ root });
}
