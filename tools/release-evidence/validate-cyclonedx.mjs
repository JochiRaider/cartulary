#!/usr/bin/env node
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { createRequire } from "node:module";
import { pathToFileURL } from "node:url";

import Ajv from "ajv";

const require = createRequire(import.meta.url);
// The pinned package's CommonJS entry is adjacent to its schema data. Resolve
// the actual dependency without executing an unrelated scanner CLI.
const dataDir = path.join(path.dirname(require.resolve("@cyclonedx/cdxgen")), "data");

function usage() {
  throw new Error("usage: validate-cyclonedx.mjs <bom-json> [<bom-json>...]");
}

function loadJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

function loadSchema() {
  const file = path.join(dataDir, "bom-1.7.schema.json");
  if (!existsSync(file)) {
    throw new Error("CycloneDX 1.7 schema missing from the pinned package");
  }
  return loadJSON(file);
}

function loadSupportSchemas(ajv) {
  for (const name of ["spdx.schema.json", "jsf-0.82.schema.json", "cryptography-defs.schema.json"]) {
    const file = path.join(dataDir, name);
    if (existsSync(file)) {
      ajv.addSchema(loadJSON(file), name);
    }
  }
}

export function validateBom(bom, label = "SBOM") {
  const specVersion = bom?.specVersion;
  if (specVersion === undefined || specVersion === null || specVersion === "") {
    throw new Error(`${label}: missing specVersion`);
  }
  if (specVersion !== "1.7") throw new Error(`${label}: unsupported CycloneDX specVersion (expected 1.7)`);

  const ajv = new Ajv({
    allErrors: false,
    strict: false,
    validateFormats: false,
    validateSchema: false,
  });
  loadSupportSchemas(ajv);
  const schema = loadSchema();
  const validate = ajv.compile(schema);
  if (!validate(bom)) {
    const details = ajv.errorsText(validate.errors, { separator: "\n  " }).slice(0, 4096);
    throw new Error(`${label}: CycloneDX ${specVersion} validation failed:\n  ${details}`);
  }
}

function validateFile(file) {
  const bom = loadJSON(path.resolve(file));
  validateBom(bom, file);
  const specVersion = bom.specVersion;
  console.log(`${file}: valid CycloneDX ${specVersion} JSON`);
}

function main() {
  const files = process.argv.slice(2);
  if (files.length === 0) {
    usage();
  }
  for (const file of files) {
    validateFile(file);
  }
}

if (import.meta.url === pathToFileURL(process.argv[1]).href) {
  try {
    main();
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    console.error(message);
    process.exit(1);
  }
}

export { validateFile };
