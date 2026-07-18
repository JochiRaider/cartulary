const postgresFixturePolicyEnvAssignable = new Set([
  "template_clone",
  "package_reset",
  "transaction",
  "group_clone",
]);

function compareStrings(left, right) {
  return String(left).localeCompare(String(right));
}

function rowPackages(row) {
  if (row.package) {
    return [row.package];
  }
  return [...(row.packages ?? [])];
}

export function symbolFixtureDetail(row, symbol) {
  return row.symbol_fixture_details?.[symbol] ?? {
    fixture_policy: row.fixture_policy ?? {},
    fixture_budget: row.fixture_budget ?? {},
  };
}

function symbolPostgresFixturePolicy(row, symbol) {
  return symbolFixtureDetail(row, symbol).fixture_policy?.postgres ?? "";
}

function symbolPostgresFixtureBudget(row, symbol) {
  return symbolFixtureDetail(row, symbol).fixture_budget?.postgres ?? {};
}

function escapeRegex(value) {
  return String(value).replace(/[.*+?^${}()|[\]\\]/gu, String.raw`\$&`);
}

function exactRegex(values) {
  const escaped = values.map(escapeRegex);
  if (escaped.length === 0) {
    throw new Error("cannot build an exact regex from an empty value list");
  }
  if (escaped.length === 1) {
    return `^${escaped[0]}$`;
  }
  return `^(${escaped.join("|")})$`;
}

function buildUnionRegex(components) {
  const values = components.filter((component) => component !== "");
  if (values.length === 0) {
    throw new Error("cannot build aggregate regex from an empty selection");
  }
  if (values.length === 1) {
    return values[0];
  }
  return values.map((component) => `(${component})`).join("|");
}

export function aggregateRegex(rows) {
  const symbols = rows.flatMap((row) => row.symbols ?? []);
  const components = [];
  if (symbols.length > 0) {
    components.push(exactRegex(symbols.sort(compareStrings)));
  }
  for (const row of rows) {
    if (row.raw_selector) {
      components.push(row.raw_selector);
    }
  }
  return buildUnionRegex(components);
}

export function aggregatePackages(rows) {
  return Array.from(new Set(rows.flatMap(rowPackages))).sort(compareStrings);
}

export function fixturePolicyAssignments(rows, mode) {
  const assignments = [];
  for (const row of rows) {
    if (mode === "tests" && row.coverage !== "raw") {
      for (const symbol of row.symbols ?? []) {
        const policy = symbolPostgresFixturePolicy(row, symbol);
        if (!postgresFixturePolicyEnvAssignable.has(policy)) {
          continue;
        }
        assignments.push(`${symbol}=${policy}`);
      }
    }
    if (mode === "packages" && row.coverage === "raw") {
      const policy = row.fixture_policy?.postgres ?? "";
      if (!postgresFixturePolicyEnvAssignable.has(policy)) {
        continue;
      }
      for (const pkg of row.packages ?? []) {
        assignments.push(`${pkg}=${policy}`);
      }
    }
  }
  return assignments.sort(compareStrings);
}

export function resetTableAssignments(rows, mode) {
  const assignments = [];
  for (const row of rows) {
    if (mode === "tests" && row.coverage !== "raw") {
      for (const symbol of row.symbols ?? []) {
        const dirtyTables = symbolPostgresFixtureBudget(row, symbol).dirty_tables ?? [];
        if (dirtyTables.length === 0) {
          continue;
        }
        assignments.push(`${symbol}=${dirtyTables.join("|")}`);
      }
    }
    if (mode === "packages" && row.coverage === "raw") {
      const dirtyTables = row.fixture_budget?.postgres?.dirty_tables ?? [];
      if (dirtyTables.length === 0) {
        continue;
      }
      for (const pkg of row.packages ?? []) {
        assignments.push(`${pkg}=${dirtyTables.join("|")}`);
      }
    }
  }
  return assignments.sort(compareStrings);
}

function aggregateKey(row) {
  if (row.coverage === "raw") {
    return `raw:${row.id}`;
  }
  if (row.support_only) {
    return [
      "support",
      row.owner_id,
      row.execution_dependency,
      row.execution_family,
      row.execution_label,
    ].join("\u001f");
  }
  return [
    "manifest",
    row.owner_id,
    row.section,
    row.coverage,
    row.execution_dependency,
    row.execution_family,
    row.execution_label,
  ].join("\u001f");
}

export function collectAggregateEmissions(rows) {
  const groups = new Map();
  for (const row of rows) {
    const key = aggregateKey(row);
    if (!groups.has(key)) {
      groups.set(key, {
        mode:
          row.coverage === "raw"
            ? "raw"
            : row.support_only
              ? "support"
              : "manifest",
        label: row.execution_label,
        owner: row.owner_id,
        section: row.section,
        coverage: row.coverage,
        execution_dependency: row.execution_dependency,
        execution_family: row.execution_family,
        support_target: row.support_only ? row.execution_dependency : "",
        regex: row.raw_selector ?? "",
        ids: new Set(),
        packages: new Set(),
        symbols: [],
      });
    }
    const group = groups.get(key);
    group.ids.add(row.id);
    for (const pkg of rowPackages(row)) {
      group.packages.add(pkg);
    }
    if (row.support_only) {
      group.symbols.push(...(row.symbols ?? []));
    }
  }
  return Array.from(groups.values()).map((group) => {
    const symbols = group.symbols.sort(compareStrings);
    return {
      ...group,
      regex: group.mode === "support" ? exactRegex(symbols) : group.regex,
      ids: Array.from(group.ids).sort(compareStrings),
      packages: Array.from(group.packages).sort(compareStrings),
      symbols,
    };
  });
}
