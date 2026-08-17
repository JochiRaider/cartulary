#!/usr/bin/env node
import { readFileSync } from "node:fs";

import { validateFile } from "./validate-cyclonedx.mjs";

function deterministicUUID(digest) {
  const hex = digest.replace(/^sha256:/u, "").slice(0, 32).split("");
  hex[12] = "5";
  hex[16] = "8";
  return `${hex.slice(0, 8).join("")}-${hex.slice(8, 12).join("")}-${hex.slice(12, 16).join("")}-${hex.slice(16, 20).join("")}-${hex.slice(20).join("")}`;
}

function main() {
  const [file] = process.argv.slice(2);
  if (!file) throw new Error("usage: validate-release-sbom.mjs <sbom-json>");
  validateFile(file);
  const bom = JSON.parse(readFileSync(file, "utf8"));
  if (Object.hasOwn(bom.metadata ?? {}, "timestamp")) {
    throw new Error(`${file}: canonical release SBOM must not contain metadata.timestamp`);
  }
  const properties = bom.metadata?.component?.properties ?? [];
  const digestProperties = properties.filter(
    (entry) => entry.name === "cartulary:semantic_input_digest",
  );
  if (
    digestProperties.length !== 1 ||
    !/^sha256:[a-f0-9]{64}$/u.test(digestProperties[0].value)
  ) {
    throw new Error(`${file}: canonical release SBOM must contain one semantic input digest`);
  }
  const expectedSerial = `urn:uuid:${deterministicUUID(digestProperties[0].value)}`;
  if (bom.serialNumber !== expectedSerial) {
    throw new Error(`${file}: canonical release SBOM serial does not match its semantic input digest`);
  }
}

try {
  main();
} catch (error) {
  console.error(error instanceof Error ? error.message : String(error));
  process.exit(1);
}
