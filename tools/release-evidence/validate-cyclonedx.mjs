#!/usr/bin/env node
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

import Ajv from "ajv";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "../..");

function usage() {
  throw new Error("usage: validate-cyclonedx.mjs <bom-json> [<bom-json>...]");
}

function loadJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

function schemaPath(specVersion) {
  return path.join(
    repoRoot,
    "node_modules",
    "@cyclonedx",
    "cdxgen",
    "data",
    `bom-${specVersion}.schema.json`,
  );
}

function loadSchema(specVersion) {
  const file = schemaPath(specVersion);
  if (!existsSync(file)) {
    throw new Error(`CycloneDX ${specVersion} schema missing at ${file}`);
  }
  return loadJSON(file);
}

function loadSupportSchemas(ajv) {
  const dataDir = path.join(repoRoot, "node_modules", "@cyclonedx", "cdxgen", "data");
  for (const name of ["spdx.schema.json", "jsf-0.82.schema.json", "cryptography-defs.schema.json"]) {
    const file = path.join(dataDir, name);
    if (existsSync(file)) {
      ajv.addSchema(loadJSON(file), name);
    }
  }
}

function validateFile(file) {
  const absolute = path.resolve(file);
  const bom = loadJSON(absolute);
  const specVersion = String(bom.specVersion ?? "");
  if (!specVersion) {
    throw new Error(`${file}: missing specVersion`);
  }

  const ajv = new Ajv({
    allErrors: true,
    strict: false,
    validateFormats: false,
    validateSchema: false,
  });
  loadSupportSchemas(ajv);
  const schema = loadSchema(specVersion);
  const validate = ajv.compile(schema);
  if (!validate(bom)) {
    const details = ajv.errorsText(validate.errors, { separator: "\n  " });
    throw new Error(`${file}: CycloneDX ${specVersion} validation failed:\n  ${details}`);
  }
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
