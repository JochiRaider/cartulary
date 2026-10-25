// Logical queues bound concurrency. Compatibility controls physical affinity,
// never whether an otherwise valid selected closure can be admitted.
export function planBrowserLanes(groups, { estimateGroup, lanePrefix = "browser", maxLanes } = {}) {
  if (!Array.isArray(groups) || !groups.length || typeof estimateGroup !== "function" ||
      !Number.isSafeInteger(maxLanes) || maxLanes < 1) throw new Error("invalid browser lane planning input");
  const names = new Set();
  const items = groups.map((group) => {
    const cost = estimateGroup(group);
    if (!group.name || names.has(group.name) || !Number.isSafeInteger(cost) || cost < 1) {
      throw new Error("browser lanes require unique groups and positive estimates");
    }
    names.add(group.name);
    return { group, estimatedWorkMs: cost };
  }).sort((a, b) => b.estimatedWorkMs - a.estimatedWorkMs || a.group.name.localeCompare(b.group.name, "en"));
  const lanes = Array.from({ length: Math.min(maxLanes, groups.length) }, (_, i) => ({
    laneID: `${lanePrefix}-${i + 1}`, estimatedWorkMs: 0, groups: [],
  }));
  for (const item of items) {
    const lane = [...lanes].sort((a, b) => a.estimatedWorkMs - b.estimatedWorkMs || a.laneID.localeCompare(b.laneID, "en"))[0];
    lane.groups.push(item);
    lane.estimatedWorkMs += item.estimatedWorkMs;
  }
  return lanes;
}
