#!/usr/bin/env node
import { createHash } from "node:crypto";
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..", "..", "..");
const profileIDPattern = /^[a-z][a-z0-9_]*$/u;
const supportedKinds = new Set(["default", "network_flow_claimed"]);

function usage() {
  throw new Error(
    "usage: browser-runtime-profile.mjs resolve <execution-topology-manifest> <runtime-profile-id>",
  );
}

function requireString(value, label) {
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(`${label} must be a non-empty string`);
  }
  return value.trim();
}

function canonicalJSON(value) {
  if (Array.isArray(value)) {
    return `[${value.map(canonicalJSON).join(",")}]`;
  }
  if (value && typeof value === "object") {
    return `{${Object.keys(value)
      .sort()
      .map((key) => `${JSON.stringify(key)}:${canonicalJSON(value[key])}`)
      .join(",")}}`;
  }
  return JSON.stringify(value);
}

export function loadBrowserRuntimeProfiles(manifestPath) {
  const manifest = JSON.parse(readFileSync(manifestPath, "utf8"));
  const profiles = manifest.browser_e2e_batch?.runtime_profiles;
  if (!Array.isArray(profiles) || profiles.length === 0) {
    throw new Error("browser_e2e_batch.runtime_profiles must be a non-empty array");
  }
  const byID = new Map();
  for (const [index, raw] of profiles.entries()) {
    const label = `browser_e2e_batch.runtime_profiles[${index + 1}]`;
    if (!raw || typeof raw !== "object" || Array.isArray(raw)) {
      throw new Error(`${label} must be an object`);
    }
    const id = requireString(raw.id, `${label}.id`);
    const kind = requireString(raw.kind, `${label}.kind`);
    if (!profileIDPattern.test(id)) {
      throw new Error(`${label}.id must be a snake_case profile identifier`);
    }
    if (!supportedKinds.has(kind)) {
      throw new Error(`${label}.kind ${kind} is not supported by the browser runtime`);
    }
    if (byID.has(id)) {
      throw new Error(`duplicate browser runtime profile ${id}`);
    }
    const manifestFile =
      raw.key_ring_manifest_path === undefined
        ? ""
        : requireString(raw.key_ring_manifest_path, `${label}.key_ring_manifest_path`);
    if (kind === "default" && manifestFile !== "") {
      throw new Error(`${label} default profile cannot declare a key-ring manifest`);
    }
    if (kind === "network_flow_claimed" && manifestFile === "") {
      throw new Error(`${label} claimed profile must declare key_ring_manifest_path`);
    }
    const fingerprintInput = {
      id,
      kind,
      key_ring_manifest_path: manifestFile,
    };
    byID.set(id, {
      id,
      kind,
      keyRingManifestPath: manifestFile,
      fingerprint: createHash("sha256")
        .update(canonicalJSON(fingerprintInput))
        .digest("hex"),
    });
  }
  if (byID.get("default")?.kind !== "default") {
    throw new Error("browser runtime profiles must declare id=default with kind=default");
  }
  return byID;
}

function main(argv) {
  const [command, configuredManifestPath, profileID, extra] = argv;
  if (command !== "resolve" || !configuredManifestPath || !profileID || extra !== undefined) {
    usage();
  }
  const manifestPath = path.isAbsolute(configuredManifestPath)
    ? configuredManifestPath
    : path.join(repoRoot, configuredManifestPath);
  const profiles = loadBrowserRuntimeProfiles(manifestPath);
  const profile = profiles.get(profileID);
  if (!profile) {
    throw new Error(`unknown browser runtime profile ${profileID}`);
  }
  process.stdout.write(
    [
      profile.id,
      profile.kind,
      profile.keyRingManifestPath || "-",
      profile.fingerprint,
    ].join("\t") + "\n",
  );
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  try {
    main(process.argv.slice(2));
  } catch (error) {
    process.stderr.write(`${error instanceof Error ? error.message : String(error)}\n`);
    process.exit(2);
  }
}
