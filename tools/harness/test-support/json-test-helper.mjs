#!/usr/bin/env node
import { readFileSync } from "node:fs";

function usage() {
  console.error("usage: json-test-helper.mjs get <file> <dot.path>");
  console.error("       json-test-helper.mjs assert-null <file> <dot.path> <label>");
  process.exit(2);
}

function jsonPathValue(object, dottedPath) {
  return dottedPath
    .split(".")
    .reduce((current, key) => current?.[key], object);
}

function readValue(file, dottedPath) {
  return jsonPathValue(JSON.parse(readFileSync(file, "utf8")), dottedPath);
}

const [command, file, dottedPath, label] = process.argv.slice(2);
if (!command || !file || !dottedPath) {
  usage();
}

const value = readValue(file, dottedPath);

switch (command) {
  case "get":
    if (label !== undefined) {
      usage();
    }
    if (value === undefined || value === null) {
      process.exit(1);
    }
    process.stdout.write(String(value));
    break;
  case "assert-null":
    if (label === undefined) {
      usage();
    }
    if (value !== null) {
      console.error(`${label}: expected null, got [${String(value)}]`);
      process.exit(1);
    }
    break;
  default:
    usage();
}
