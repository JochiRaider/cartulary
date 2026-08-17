function compareASCII(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function safeID(value) {
  return value.replaceAll(/[^A-Za-z0-9_.+-]+/gu, "-");
}

function assertPositiveInteger(value, label) {
  if (!Number.isSafeInteger(value) || value < 1) {
    throw new Error(`${label} must be a positive safe integer`);
  }
}

export function planBrowserFunctionalLanes(
  groups,
  { estimateGroup, lanePrefix = "browser-functional", maxLanes = 4 } = {},
) {
  if (!Array.isArray(groups) || groups.length === 0) {
    throw new Error("browser functional lane planning requires selected groups");
  }
  if (typeof estimateGroup !== "function") {
    throw new Error("browser functional lane planning requires estimateGroup");
  }
  assertPositiveInteger(maxLanes, "browser functional maxLanes");

  const names = new Set();
  const partitionsByProfile = new Map();
  for (const group of groups) {
    if (!group || typeof group.name !== "string" || group.name.trim() === "") {
      throw new Error("browser functional group must have a stable name");
    }
    if (names.has(group.name)) {
      throw new Error(`duplicate browser functional group ${group.name}`);
    }
    names.add(group.name);
    if (group.resourceProfileID !== "browser_functional") {
      throw new Error(
        `browser functional group ${group.name} must use browser_functional`,
      );
    }
    const runtimeProfileID = String(group.runtimeProfileID ?? "").trim();
    if (!runtimeProfileID) {
      throw new Error(`browser functional group ${group.name} has no runtime profile`);
    }
    const estimatedWorkMs = estimateGroup(group);
    assertPositiveInteger(
      estimatedWorkMs,
      `browser functional group ${group.name} estimate`,
    );
    const partition = partitionsByProfile.get(runtimeProfileID) ?? {
      runtimeProfileID,
      groups: [],
      totalEstimatedWorkMs: 0,
      laneCount: 1,
    };
    partition.groups.push({ group, estimatedWorkMs });
    partition.totalEstimatedWorkMs += estimatedWorkMs;
    partitionsByProfile.set(runtimeProfileID, partition);
  }

  const partitions = [...partitionsByProfile.values()].sort((left, right) =>
    compareASCII(left.runtimeProfileID, right.runtimeProfileID),
  );
  if (partitions.length > maxLanes) {
    throw new Error(
      `browser functional closure has ${partitions.length} runtime profiles but only ${maxLanes} lanes`,
    );
  }

  const wantedLanes = Math.min(maxLanes, groups.length);
  let allocatedLanes = partitions.length;
  while (allocatedLanes < wantedLanes) {
    const candidates = partitions
      .filter((partition) => partition.laneCount < partition.groups.length)
      .sort((left, right) => {
        const leftScaled = left.totalEstimatedWorkMs * right.laneCount;
        const rightScaled = right.totalEstimatedWorkMs * left.laneCount;
        return rightScaled - leftScaled ||
          compareASCII(left.runtimeProfileID, right.runtimeProfileID);
      });
    const selected = candidates[0];
    if (!selected) break;
    selected.laneCount += 1;
    allocatedLanes += 1;
  }

  const lanes = [];
  for (const partition of partitions) {
    const compatibleLanes = Array.from(
      { length: partition.laneCount },
      (_, index) => ({
        laneID: `${safeID(lanePrefix)}-${safeID(partition.runtimeProfileID)}-${index + 1}`,
        runtimeProfileID: partition.runtimeProfileID,
        estimatedWorkMs: 0,
        groups: [],
      }),
    );
    const sortedGroups = [...partition.groups].sort(
      (left, right) =>
        right.estimatedWorkMs - left.estimatedWorkMs ||
        compareASCII(left.group.name, right.group.name),
    );
    for (const item of sortedGroups) {
      const lane = [...compatibleLanes].sort(
        (left, right) =>
          left.estimatedWorkMs - right.estimatedWorkMs ||
          compareASCII(left.laneID, right.laneID),
      )[0];
      lane.groups.push({
        group: item.group,
        estimatedWorkMs: item.estimatedWorkMs,
        generation: lane.groups.length + 1,
      });
      lane.estimatedWorkMs += item.estimatedWorkMs;
    }
    lanes.push(...compatibleLanes);
  }
  return lanes.sort((left, right) => compareASCII(left.laneID, right.laneID));
}
