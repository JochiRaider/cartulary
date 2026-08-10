import { createRequire } from "node:module";
import { readFileSync, readdirSync } from "node:fs";
import path from "node:path";

const require = createRequire(import.meta.url);
const Ajv = require("ajv/dist/2020");

const evidenceSchemaID = "cartulary.migration_history_evidence.v2";
const negativeSchemaID =
  "cartulary.migration_history_evidence_negative_fixtures.v2";
const expectedPaths = Object.freeze([
  "fixtures/migration-history-evidence-negative.v2.json",
  "fixtures/migration-history-evidence.v1.rejected.json",
  "fixtures/migration-history-evidence.v2.valid.json",
  "index.json",
  "migration-history-evidence.v2.schema.json",
]);
const expectedCases = Object.freeze([
  "v1_schema_rejected",
  "manifest_path_rejected",
  "manifest_source_path_rejected",
  "manifest_repository_path_rejected",
  "manifest_embedded_path_rejected",
  "manifest_file_rejected",
  "manifest_uri_rejected",
  "manifest_path_hash_rejected",
  "unknown_top_level_rejected",
  "absolute_detail_rejected",
  "absolute_service_ref_rejected",
]);

export function validateMigrationHistoryEvidenceContracts(root) {
  const contractRoot = path.join(root, "contracts/database-migrations");
  const actualPaths = collectFiles(contractRoot);
  if (actualPaths.join("\n") !== expectedPaths.join("\n")) {
    throw new Error(
      `database-migrations contract paths mismatch: ${actualPaths.join(", ")}`,
    );
  }

  const schema = readObject(
    path.join(contractRoot, "migration-history-evidence.v2.schema.json"),
  );
  if (
    schema.$id !== evidenceSchemaID ||
    schema.additionalProperties !== false
  ) {
    throw new Error("migration-history evidence schema must be closed v2");
  }
  assertAllObjectSchemasClosed(schema, "migration-history evidence schema");
  const ajv = new Ajv({
    allErrors: true,
    strict: false,
    validateFormats: false,
    validateSchema: true,
  });
  const validate = ajv.compile(schema);

  const valid = readObject(
    path.join(
      contractRoot,
      "fixtures/migration-history-evidence.v2.valid.json",
    ),
  );
  if (!validate(valid)) {
    throw new Error(
      `valid migration-history evidence fixture failed v2: ${ajv.errorsText(validate.errors)}`,
    );
  }

  const rejectedV1 = readObject(
    path.join(
      contractRoot,
      "fixtures/migration-history-evidence.v1.rejected.json",
    ),
  );
  if (validate(rejectedV1)) {
    throw new Error("v1 migration-history evidence fixture was admitted by v2");
  }

  const negative = readObject(
    path.join(
      contractRoot,
      "fixtures/migration-history-evidence-negative.v2.json",
    ),
  );
  if (negative.schema_id !== negativeSchemaID || !Array.isArray(negative.cases)) {
    throw new Error("migration-history negative fixture registry is invalid");
  }
  const caseIDs = negative.cases.map((entry) => entry.case_id);
  if (caseIDs.join("\n") !== expectedCases.join("\n")) {
    throw new Error("migration-history negative fixture cases are incomplete or reordered");
  }
  for (const current of negative.cases) {
    const mutated = structuredClone(valid);
    applyMutation(mutated, current);
    if (validate(mutated)) {
      throw new Error(
        `migration-history negative fixture ${current.case_id} was admitted`,
      );
    }
  }

  const attachments = readObject(
    path.join(root, "tools/harness_schema_attachments.json"),
  );
  if (
    JSON.stringify(attachments).includes(evidenceSchemaID) ||
    JSON.stringify(attachments).includes("contracts/database-migrations")
  ) {
    throw new Error(
      "database-migrations product schema must not enter harness schema attachments",
    );
  }

  return { schemaID: evidenceSchemaID, negativeCases: negative.cases.length };
}

function applyMutation(value, current) {
  switch (current.mutation) {
    case "schema_id_v1":
      value.schema_id = "cartulary.migration_history_evidence.v1";
      return;
    case "add_manifest_member":
      value.manifest[current.member] = current.value;
      return;
    case "add_top_level_member":
      value[current.member] = current.value;
      return;
    case "absolute_finding_detail":
      value.findings[0].detail = current.value;
      return;
    case "absolute_service_ref":
      value.database_binding.service_ref = current.value;
      return;
    default:
      throw new Error(`unknown migration-history evidence mutation ${current.mutation}`);
  }
}

function assertAllObjectSchemasClosed(value, label) {
  if (Array.isArray(value)) {
    value.forEach((entry, index) =>
      assertAllObjectSchemasClosed(entry, `${label}[${index}]`),
    );
    return;
  }
  if (!value || typeof value !== "object") return;
  if (value.type === "object" && value.additionalProperties !== false) {
    throw new Error(`${label} contains an open object schema`);
  }
  for (const [key, child] of Object.entries(value)) {
    assertAllObjectSchemasClosed(child, `${label}.${key}`);
  }
}

function collectFiles(root) {
  const results = [];
  const walk = (directory) => {
    for (const entry of readdirSync(directory, { withFileTypes: true })) {
      const absolute = path.join(directory, entry.name);
      if (entry.isDirectory()) {
        walk(absolute);
      } else if (entry.isFile()) {
        results.push(path.relative(root, absolute).split(path.sep).join("/"));
      }
    }
  };
  walk(root);
  return results.sort((left, right) => left.localeCompare(right));
}

function readObject(file) {
  const value = JSON.parse(readFileSync(file, "utf8"));
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${file} must contain one JSON object`);
  }
  return value;
}
