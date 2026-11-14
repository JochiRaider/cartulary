#!/usr/bin/env node
import path from "node:path";
import { pathToFileURL } from "node:url";
import { parseStrictJSON, readRetainedArtifactBytes } from "../harness/contract/index.mjs";

import { validateBom } from "./validate-cyclonedx.mjs";

function deterministicUUID(digest) {
  const hex = digest.replace(/^sha256:/u, "").slice(0, 32).split("");
  hex[12] = "5";
  hex[16] = "8";
  return `${hex.slice(0, 8).join("")}-${hex.slice(8, 12).join("")}-${hex.slice(12, 16).join("")}-${hex.slice(16, 20).join("")}-${hex.slice(20).join("")}`;
}

export function validateReleaseBom(bom) {
  validateBom(bom);
  if (Object.hasOwn(bom.metadata ?? {}, "timestamp")) {
    throw new Error("canonical release SBOM must not contain metadata.timestamp");
  }
  const properties = bom.metadata?.component?.properties ?? [];
  const digestProperties = properties.filter(
    (entry) => entry.name === "cartulary:semantic_input_digest",
  );
  if (
    digestProperties.length !== 1 ||
    !/^sha256:[a-f0-9]{64}$/u.test(digestProperties[0].value)
  ) {
    throw new Error("canonical release SBOM must contain one semantic input digest");
  }
  const expectedSerial = `urn:uuid:${deterministicUUID(digestProperties[0].value)}`;
  if (bom.serialNumber !== expectedSerial) {
    throw new Error("canonical release SBOM serial does not match its semantic input digest");
  }
  return digestProperties[0].value;
}

function main() {
  const [file, ...extra] = process.argv.slice(2);
  if (!file || extra.length) throw new Error("usage: validate-release-sbom.mjs <sbom-json>");
  const absolute = path.resolve(file);
  const bytes = readRetainedArtifactBytes(path.dirname(absolute), path.basename(absolute), 16 * 1024 * 1024);
  validateReleaseBom(parseStrictJSON(new TextDecoder("utf-8", { fatal: true }).decode(bytes)));
  console.log(`${file}: valid canonical CycloneDX 1.7 JSON`);
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  try { main(); }
  catch { console.error("canonical release SBOM admission failed"); process.exitCode = 1; }
}
