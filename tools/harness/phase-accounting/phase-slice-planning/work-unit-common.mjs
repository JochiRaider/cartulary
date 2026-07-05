export const goCPUResource = "go_cpu";
export const goIOResource = "go_io";
export const browserStackResource = "browser_stack";

export function compareStrings(left, right) {
  return String(left).localeCompare(String(right));
}

export function uniqueSorted(values) {
  return Array.from(new Set(values.filter(Boolean))).sort(compareStrings);
}

export function targetWeight(rows) {
  return Math.max(1, rows.length) * 1000;
}

export function mergeClaims(...claimMaps) {
  const merged = new Map();
  for (const claims of claimMaps) {
    for (const [resource, amount] of claims.entries()) {
      merged.set(resource, (merged.get(resource) ?? 0) + amount);
    }
  }
  return merged;
}
