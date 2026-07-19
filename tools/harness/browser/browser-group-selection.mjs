function asciiCompare(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

export function selectedBrowserGroupRowIDs(group, env = process.env) {
  if (!Object.hasOwn(env, "CARTULARY_BROWSER_SELECTED_ROW_IDS")) {
    return [...group.selectedRowIDs];
  }
  const selected = String(env.CARTULARY_BROWSER_SELECTED_ROW_IDS)
    .split(",")
    .map((rowID) => rowID.trim())
    .filter(Boolean);
  const sorted = [...selected].sort(asciiCompare);
  if (
    selected.length === 0 ||
    new Set(selected).size !== selected.length ||
    JSON.stringify(selected) !== JSON.stringify(sorted)
  ) {
    throw new Error(
      `browser group ${group.name} scheduled row selection must be non-empty, sorted, and unique`,
    );
  }
  const available = new Set(group.selectedRowIDs);
  const unsupported = selected.filter((rowID) => !available.has(rowID));
  if (unsupported.length > 0) {
    throw new Error(
      `browser group ${group.name} scheduled row selection contains rows outside the generated group: ${unsupported.join(", ")}`,
    );
  }
  return selected;
}
