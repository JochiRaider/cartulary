#!/usr/bin/env node
import { validateSchedulerEventOrder } from "../tools/harness/scheduler/scheduler/event-order.mjs";

function usage() {
  process.stderr.write(
    "usage: check-scheduler-event-order-drift.mjs [--target <target>] <results-dir|run-dir|scheduler-events.jsonl>\n",
  );
  process.exit(2);
}

function parseArgs(argv) {
  const options = { target: "", resultsDir: "" };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--target") {
      options.target = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg.startsWith("--")) {
      usage();
    }
    if (options.resultsDir) {
      usage();
    }
    options.resultsDir = arg;
  }
  if (!options.resultsDir) {
    usage();
  }
  return options;
}

const options = parseArgs(process.argv.slice(2));
const { files, errors } = validateSchedulerEventOrder(options.resultsDir, { target: options.target });
if (files.length === 0) {
  process.stderr.write("no scheduler-events.jsonl files found\n");
  process.exit(1);
}
if (errors.length > 0) {
  process.stderr.write("scheduler event order drift detected:\n");
  for (const error of errors) {
    process.stderr.write(`  ${error}\n`);
  }
  process.exit(1);
}
process.stdout.write(`scheduler event order verified for ${files.length} event stream(s)\n`);

