import path from "node:path";
import { publishCommandFailure, readCommandFailure } from "./command-failure.mjs";

try {
  if (process.argv.length !== 4) throw new Error("invalid diagnostic arguments");
  if (!readCommandFailure(path.resolve(import.meta.dirname, "../../.."))) publishCommandFailure(path.resolve(import.meta.dirname, "../../.."), { failure_class: process.argv[2], failure_reason: process.argv[3] });
} catch { process.exitCode = 11; }
