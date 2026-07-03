const defaultSummaryLimit = 5;
const defaultHumanLimit = 3;

function normalizeString(value) {
  return typeof value === "string" ? value.trim() : "";
}

function sortedUnique(values) {
  const source = Array.isArray(values) ? values : [];
  return Array.from(new Set(source.map(normalizeString).filter(Boolean))).sort(
    (left, right) => left.localeCompare(right),
  );
}

function reasonTokens(reason) {
  return sortedUnique(
    String(reason ?? "")
      .split(",")
      .map((entry) => entry.trim())
      .filter((entry) => entry && entry !== "none"),
  );
}

function blockedUnitRecords(blockedUnits) {
  if (!Array.isArray(blockedUnits)) {
    return [];
  }
  return blockedUnits
    .filter((unit) => unit && typeof unit === "object" && !Array.isArray(unit))
    .map((unit) => ({
      workUnit: normalizeString(unit.work_unit),
      waitingOn: sortedUnique(unit.waiting_on ?? []),
      blockedResources: sortedUnique(unit.blocked_resources ?? []),
    }));
}

function fallbackObservations({ blockedResources, waitingOn }) {
  return [
    ...blockedResources.map((name) => ({ kind: "resource", name })),
    ...waitingOn.map((name) => ({ kind: "dependency", name })),
  ];
}

function blockerKey({ kind, name }) {
  return `${kind}:${name}`;
}

export function schedulerBlockedDiagnostics({
  reason = "none",
  blockedResources = [],
  waitingOn = [],
  blockedUnits = [],
} = {}) {
  const resources = sortedUnique(blockedResources);
  const dependencies = sortedUnique(waitingOn);
  const explanations = new Set();
  const observations = [];

  for (const token of reasonTokens(reason)) {
    if (token === "dependencies") {
      explanations.add("dependencies");
      continue;
    }
    if (token !== "resources") {
      explanations.add(token);
    }
  }
  for (const resource of resources) {
    explanations.add(resource);
  }
  if (dependencies.length > 0) {
    explanations.add("dependencies");
  }

  for (const unit of blockedUnitRecords(blockedUnits)) {
    if (unit.waitingOn.length > 0) {
      explanations.add("dependencies");
    }
    for (const dependency of unit.waitingOn) {
      observations.push({ kind: "dependency", name: dependency });
    }
    for (const resource of unit.blockedResources) {
      explanations.add(resource);
      observations.push({ kind: "resource", name: resource });
    }
  }

  const observed =
    observations.length > 0
      ? observations
      : fallbackObservations({
          blockedResources: resources,
          waitingOn: dependencies,
        });
  const materialParts = sortedUnique(
    observed.length > 0
      ? observed.map(blockerKey)
      : Array.from(explanations).map(
          (explanation) => `explanation:${explanation}`,
        ),
  );

  return {
    explanations: sortedUnique(Array.from(explanations)),
    observations: observed,
    materialKey: materialParts.join("|") || "none",
  };
}

export function addTopBlockerObservations(counts, observations) {
  for (const observation of observations) {
    const kind = normalizeString(observation?.kind);
    const name = normalizeString(observation?.name);
    if (!kind || !name) {
      continue;
    }
    const key = blockerKey({ kind, name });
    counts.set(key, {
      blocker: key,
      kind,
      name,
      count: (counts.get(key)?.count ?? 0) + 1,
    });
  }
}

export function topBlockerRecords(counts, limit = defaultSummaryLimit) {
  return Array.from(counts.values())
    .sort((left, right) =>
      right.count - left.count ||
      left.kind.localeCompare(right.kind) ||
      left.name.localeCompare(right.name),
    )
    .slice(0, limit)
    .map((entry) => ({
      blocker: entry.blocker,
      kind: entry.kind,
      name: entry.name,
      count: entry.count,
    }));
}

export function formatTopBlockers(blockers, limit = defaultHumanLimit) {
  if (!Array.isArray(blockers) || blockers.length === 0) {
    return "none";
  }
  const displayed = blockers
    .slice(0, limit)
    .map((entry) => `${entry.kind}:${entry.name}(${entry.count})`);
  const suffix = blockers.length > limit ? `,+${blockers.length - limit}` : "";
  return `${displayed.join(",")}${suffix}`;
}
