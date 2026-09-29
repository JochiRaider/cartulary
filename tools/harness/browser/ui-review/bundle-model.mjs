import { isDeepStrictEqual } from "node:util";
import { ReviewFailure } from "./failure.mjs";

export function primaryImage(bundle) {
  const ref = bundle.components.actual ?? bundle.components.original;
  if (!ref) throw new ReviewFailure("invalid_artifact");
  return ref;
}
export function requireComparable(left, right, kind) {
  if (kind === "reference") return;
  if (!left.comparisonIdentity || !right.comparisonIdentity || !isDeepStrictEqual(left.comparisonIdentity, right.comparisonIdentity)) throw new ReviewFailure("invalid_artifact");
}
