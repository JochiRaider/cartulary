import { removeResourceClaims } from "./state.mjs";

export function createRetainedClaimTracker(activeClaims) {
  const retainedClaims = new Map();

  return {
    removeFinishedUnitClaims(unit) {
      const retained = unit.retainedResourceClaims ?? new Map();
      const releasable = new Map();
      for (const [resource, amount] of unit.resourceClaims.entries()) {
        const retainedAmount = retained.get(resource) ?? 0;
        const next = amount - retainedAmount;
        if (next > 0) {
          releasable.set(resource, next);
        }
        if (retainedAmount > 0) {
          retainedClaims.set(
            resource,
            (retainedClaims.get(resource) ?? 0) + retainedAmount,
          );
        }
      }
      removeResourceClaims({ resourceClaims: releasable }, activeClaims);
    },

    releaseRetainedClaims() {
      if (retainedClaims.size === 0) {
        return;
      }
      removeResourceClaims({ resourceClaims: retainedClaims }, activeClaims);
      retainedClaims.clear();
    },

    releaseRetainedClaimsForUnit(unit) {
      const claims = unit.releaseRetainedResourceClaims ?? new Map();
      if (claims.size === 0) {
        return;
      }
      const releasable = new Map();
      for (const [resource, amount] of claims.entries()) {
        const retainedAmount = retainedClaims.get(resource) ?? 0;
        const next = Math.min(retainedAmount, amount);
        if (next <= 0) {
          continue;
        }
        if (retainedAmount === next) {
          retainedClaims.delete(resource);
        } else {
          retainedClaims.set(resource, retainedAmount - next);
        }
        releasable.set(resource, next);
      }
      if (releasable.size > 0) {
        removeResourceClaims({ resourceClaims: releasable }, activeClaims);
      }
    },
  };
}
