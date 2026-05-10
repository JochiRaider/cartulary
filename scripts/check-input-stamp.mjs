#!/usr/bin/env node
import { createHash } from "node:crypto";
import { existsSync, readFileSync, writeFileSync, mkdirSync } from "node:fs";
import path from "node:path";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");
const stampRoot = path.join(repoRoot, "tmp", "check-stamps");

const profiles = {
  lint_shell: {
    prefixes: ["scripts/"],
    suffixes: [".sh"],
    files: [
      "Makefile",
      "scripts/run-shellcheck.sh",
      "scripts/lib/generated-artifacts.sh",
      "tools/generated_artifact_policy.json",
    ],
    env: ["SHELLCHECK_BIN", "SHELLCHECK_VERSION", "LINT_SHELL_STRICT"],
  },
  migration_drift: {
    prefixes: ["cmd/migrate/", "db/migrations/", "internal/app/", "internal/platform/", "scripts/"],
    suffixes: [".go", ".sql", ".sh"],
    files: [
      "Makefile",
      "go.mod",
      "go.sum",
      "configs/dev/config.toml",
      "docker-compose.dev.yml",
      "scripts/check-migrations.sh",
      "scripts/dev-services.sh",
    ],
    env: ["GO", "CONFIG_FILE", "CARTULARY_MIGRATE_BIN"],
  },
  static_artifact_validation: {
    prefixes: ["contracts/", "scripts/", "tools/", "docs/spec/"],
    suffixes: [".json", ".mjs", ".js", ".sh", ".md"],
    files: ["Makefile", "package.json", "pnpm-lock.yaml", "pnpm-workspace.yaml"],
    env: ["NODE_BIN"],
  },
};

function usage() {
  console.error("usage: check-input-stamp.mjs --stamp-id <id> --profile <profile> -- <command> [args...]");
  process.exit(2);
}

function parseArgs(argv) {
  const options = { stampID: "", profile: "", command: [] };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--") {
      options.command = argv.slice(index + 1);
      break;
    }
    if (arg === "--stamp-id") {
      options.stampID = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--profile") {
      options.profile = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    usage();
  }
  if (!options.stampID || !options.profile || options.command.length === 0) {
    usage();
  }
  return options;
}

function gitFiles() {
  const result = spawnSync("git", ["ls-files", "-z"], {
    cwd: repoRoot,
    encoding: "utf8",
  });
  if (result.status !== 0) {
    throw new Error(`git ls-files failed: ${result.stderr || result.status}`);
  }
  return result.stdout.split("\0").filter(Boolean).sort((left, right) => left.localeCompare(right));
}

function matchesProfile(file, profile) {
  if (profile.files.includes(file)) {
    return true;
  }
  const prefixMatch = profile.prefixes.some((prefix) => file.startsWith(prefix));
  const suffixMatch = profile.suffixes.some((suffix) => file.endsWith(suffix));
  return prefixMatch && suffixMatch;
}

function hashFile(hash, file) {
  const absolute = path.join(repoRoot, file);
  if (!existsSync(absolute)) {
    return;
  }
  hash.update(`file\0${file}\0`);
  hash.update(readFileSync(absolute));
  hash.update("\0");
}

function computeDigest({ stampID, profileName, command }) {
  const profile = profiles[profileName];
  if (!profile) {
    throw new Error(`unknown local input stamp profile ${profileName}`);
  }
  const hash = createHash("sha256");
  hash.update(JSON.stringify({
    schema_id: "cartulary.check_input_stamp.v1",
    stamp_id: stampID,
    profile: profileName,
    command,
  }));
  hash.update("\0");
  for (const name of profile.env) {
    hash.update(`env\0${name}\0${process.env[name] ?? ""}\0`);
  }
  for (const file of gitFiles().filter((entry) => matchesProfile(entry, profile))) {
    hashFile(hash, file);
  }
  return `sha256:${hash.digest("hex")}`;
}

function sanitizeStampID(value) {
  return value.replace(/[^A-Za-z0-9._-]+/g, "-");
}

function readStamp(file) {
  try {
    return JSON.parse(readFileSync(file, "utf8"));
  } catch {
    return null;
  }
}

function writeStamp(file, data) {
  mkdirSync(path.dirname(file), { recursive: true });
  writeFileSync(file, `${JSON.stringify(data, null, 2)}\n`);
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  const disabled = process.env.CI === "1" || process.env.CARTULARY_CHECK_DISABLE_INPUT_STAMPS === "1";
  const digest = disabled ? "" : computeDigest({
    stampID: options.stampID,
    profileName: options.profile,
    command: options.command,
  });
  const stampFile = path.join(stampRoot, `${sanitizeStampID(options.stampID)}.json`);
  const existing = disabled ? null : readStamp(stampFile);
  if (!disabled && existing?.digest === digest) {
    console.log(`check input stamp hit: ${options.stampID}`);
    return;
  }

  if (disabled) {
    console.log(`check input stamp bypassed: ${options.stampID}`);
  } else {
    console.log(`check input stamp miss: ${options.stampID}`);
  }
  const [command, ...args] = options.command;
  const result = spawnSync(command, args, {
    cwd: repoRoot,
    env: process.env,
    stdio: "inherit",
  });
  const status = result.status ?? 1;
  if (status === 0 && !disabled) {
    writeStamp(stampFile, {
      schema_id: "cartulary.check_input_stamp.v1",
      stamp_id: options.stampID,
      profile: options.profile,
      digest,
      updated_at: new Date().toISOString(),
    });
  }
  process.exit(status);
}

try {
  main();
} catch (error) {
  console.error(`check input stamp failed: ${error.message}`);
  process.exit(1);
}
