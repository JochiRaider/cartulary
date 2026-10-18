import { loadTestCatalog } from "../test-catalog/index.mjs";

const ownerIDPattern = /^(?:module|platform|app|web|package|harness)\.[a-z][a-z0-9_]{0,62}$/u;

function asciiCompare(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function sortedCounts(values) {
  const counts = new Map();
  for (const value of values) counts.set(value, (counts.get(value) ?? 0) + 1);
  return Object.fromEntries([...counts.entries()].sort(([left], [right]) => asciiCompare(left, right)));
}

function requireOwner(catalog, ownerID) {
  const normalized = String(ownerID ?? "").trim();
  if (!ownerIDPattern.test(normalized)) throw new Error("OWNER is missing or malformed");
  const owner = catalog.registry.owners.find((entry) => entry.owner_id === normalized && entry.status === "active");
  if (!owner) throw new Error(`unknown active test owner ${normalized}`);
  return owner;
}

function ownerProjection(root, ownerID) {
  const catalog = loadTestCatalog(root);
  const owner = requireOwner(catalog, ownerID);
  const rows = catalog.rows
    .filter((row) => row.owner_id === owner.owner_id)
    .sort((left, right) => asciiCompare(left.row_id, right.row_id));
  if (rows.length === 0) throw new Error(`active test owner ${owner.owner_id} has no rows`);
  const serviceRows = rows.filter((row) => row.service_dependencies.length > 0);
  return { catalog, owner, rows, serviceRows };
}

function broaderCommands(projection) {
  const gate = { fast: "test-fast", standard: "check", full: "ci", release: "release-check" };
  return [...new Set(projection.rows.map((row) => `make ${gate[row.minimum_tier]}`))].sort(asciiCompare);
}

export function buildOwnerTaskGuide(root, ownerID) {
  const projection = ownerProjection(root, ownerID);
  const evidenceClasses = new Set(projection.rows.map((row) => row.evidence_class));
  const generatedCommands = [];
  if (
    projection.owner.owner_id === "harness.generated_artifacts" ||
    projection.rows.some((row) => row.runner === "shell" && row.selector.operation_id.includes("generate"))
  ) {
    generatedCommands.push("make source-check CHECK=generated", "make source-check CHECK=schema");
  }
  const releaseRequired = evidenceClasses.has("release");
  return {
    schema_id: "cartulary.task_guide_summary.v4",
    evidence_epoch: projection.catalog.summary.evidence_epoch,
    owner_id: projection.owner.owner_id,
    test_catalog_digest: projection.catalog.summary.test_catalog_digest,
    verification_routing_digest: projection.catalog.summary.verification_routing_digest,
    focused_commands: [
      `make test-slice OWNER=${projection.owner.owner_id}`,
      ...(projection.serviceRows.length > 0
        ? [`make test-slice OWNER=${projection.owner.owner_id} SCOPE=services`]
        : []),
    ],
    generated_commands: generatedCommands.sort(asciiCompare),
    broader_commands: broaderCommands(projection),
    release_gate: releaseRequired ? "make release-check" : null,
    rows: projection.rows.map((row) => ({ row_id: row.row_id, family_id: row.family_id, runtime_profile_id: row.runtime_profile_id, resource_profile_id: row.resource_profile_id, runner: row.runner, evidence_class: row.evidence_class,
      minimum_tier: row.minimum_tier, fixture_capability: row.fixture_capability,
      service_dependencies: row.service_dependencies,
      command: `make test-slice OWNER=${projection.owner.owner_id} ROWS=${row.row_id}` })),
    coverage: { row_count: projection.rows.length, service_backed_row_count: projection.serviceRows.length,
      families: sortedCounts(projection.rows.map((row) => row.family_id)),
      runtime_profiles: sortedCounts(projection.rows.map((row) => row.runtime_profile_id)),
      resource_profiles: sortedCounts(projection.rows.map((row) => row.resource_profile_id)),
      runners: sortedCounts(projection.rows.map((row) => row.runner)),
      evidence_classes: sortedCounts(projection.rows.map((row) => row.evidence_class)),
      minimum_tiers: sortedCounts(projection.rows.map((row) => row.minimum_tier)),
      fixture_capabilities: sortedCounts(projection.rows.map((row) => row.fixture_capability)) },
  };
}
