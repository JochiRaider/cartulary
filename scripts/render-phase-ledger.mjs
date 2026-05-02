import { collectEntries, collectSupportGoEntries, loadManifest } from "./lib/phase-manifest.mjs";
import {
  activePhaseRegistryEntries,
  activePhaseRegistryEntry,
} from "./lib/phase-registry.mjs";

const supportTargetDisplay = new Map([
  ["backend_unit", "backend-unit"],
  ["backend_integration_support", "backend-integration-support"],
]);
const supportTargetOrder = new Map([
  ["backend_unit", 0],
  ["backend_integration_support", 1],
]);

export function phaseLedgerOutputPath(phase) {
  return `docs/testing/${phase}_coverage_ledger.md`;
}

export function phaseLedgerOutputs(root = process.cwd()) {
  return activePhaseRegistryEntries(root).map((entry) => {
    const phase = entry.phase;
    const { manifestPath, manifest } = loadManifest(root, phase);
    if (manifest.ledger === undefined) {
      throw new Error(`${manifestPath} must declare ledger metadata for phase ledger rendering`);
    }
    return { phase, outputPath: entry.ledger_path };
  });
}

function renderEvidence(entry) {
  if (entry.runner === "go_test") {
    const symbols = entry.symbols ?? [entry.symbol];
    return symbols
      .map((symbol, index) =>
        index === 0 ? `\`${entry.file}::${symbol}\`` : `\`${symbol}\``,
      )
      .join(", ");
  }
  return `\`${entry.file}::${entry.title}\``;
}

function renderExecution(entry) {
  return `\`${entry.execution_dependency}\``;
}

function phaseNumberFromKey(phase) {
  const match = /^phase(0|[1-9]\d*)$/.exec(phase);
  if (!match) {
    throw new Error(`unsupported phase key ${phase}`);
  }
  return match[1];
}

function sectionPrefix(section) {
  switch (section) {
    case "unit":
      return "U";
    case "integration":
      return "I";
    case "e2e":
      return "E";
    default:
      throw new Error(`unsupported support section ${section}`);
  }
}

function requireLedgerString(ledger, field, phase) {
  const value = ledger[field];
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(`${phase} ledger.${field} must be a non-empty string`);
  }
  return value;
}

function optionalLedgerStringArray(ledger, field, phase) {
  const value = ledger[field];
  if (value === undefined) {
    return [];
  }
  if (!Array.isArray(value) || value.some((entry) => typeof entry !== "string")) {
    throw new Error(`${phase} ledger.${field} must be an array of strings`);
  }
  return value;
}

function requireLedgerStringArray(ledger, field, phase) {
  const value = optionalLedgerStringArray(ledger, field, phase);
  if (value.length === 0) {
    throw new Error(`${phase} ledger.${field} must be a non-empty array of strings`);
  }
  return value;
}

function ledgerConfig(manifest, phase, registryEntry) {
  const ledger = manifest.ledger;
  if (ledger === null || Array.isArray(ledger) || typeof ledger !== "object") {
    throw new Error(`${phase} manifest must declare ledger metadata object`);
  }
  if (ledger.sections === null || Array.isArray(ledger.sections) || typeof ledger.sections !== "object") {
    throw new Error(`${phase} ledger.sections must be an object`);
  }
  const sections = Object.entries(ledger.sections);
  if (sections.length === 0) {
    throw new Error(`${phase} ledger.sections must not be empty`);
  }
  for (const [section, title] of sections) {
    if (!["unit", "integration", "e2e"].includes(section)) {
      throw new Error(`${phase} ledger.sections has unsupported section ${section}`);
    }
    if (typeof title !== "string" || title.trim() === "") {
      throw new Error(`${phase} ledger.sections.${section} must be a non-empty string`);
    }
  }

  return {
    title: requireLedgerString(ledger, "title", phase),
    manifestPath: registryEntry.manifest_path,
    scope: requireLedgerString(registryEntry, "scope", phase),
    normativeOwners: requireLedgerString(registryEntry, "normative_owners", phase),
    notes: optionalLedgerStringArray(ledger, "notes", phase),
    authoritativeExecution: requireLedgerStringArray(ledger, "authoritative_execution", phase),
    supportExecutionExtras: optionalLedgerStringArray(ledger, "support_execution_extras", phase),
    sections,
    sharedHarness: requireLedgerStringArray(ledger, "shared_harness", phase),
    supportOnly: requireLedgerStringArray(ledger, "support_only", phase),
  };
}

function renderBulletLines(lines) {
  return lines.map((line) => `- ${line}`);
}

function renderIntroduction(phase, config) {
  const phaseLabel = phase.replace(/^phase/, "Phase ");
  return [
    `This ledger is generated from \`${config.manifestPath}\`. Update the manifest row metadata first, then regenerate this file.`,
    "",
    `- Scope: ${config.scope}`,
    `- Normative owners: ${config.normativeOwners}`,
    `- Authority: \`tools/${phase}_test_map.json\` is the enforced ${phaseLabel} traceability source. This ledger is a rendered companion and does not control the mechanical row inventory.`,
    ...renderBulletLines(config.notes),
    "",
    "## Authoritative Execution",
    "",
    ...renderBulletLines(config.authoritativeExecution),
  ];
}

function renderSupportExecutionLines(phase, manifest, extras = []) {
  const phaseNumber = phaseNumberFromKey(phase);
  const supportLines = collectSupportGoEntries(manifest)
    .sort((left, right) => {
      const orderDiff =
        (supportTargetOrder.get(left.target) ?? Number.MAX_SAFE_INTEGER) -
        (supportTargetOrder.get(right.target) ?? Number.MAX_SAFE_INTEGER);
      if (orderDiff !== 0) {
        return orderDiff;
      }
      return left.file.localeCompare(right.file);
    })
    .map((entry) => {
      const target = supportTargetDisplay.get(entry.target);
      if (!target) {
        throw new Error(`unsupported support target ${entry.target}`);
      }
      return `- \`${entry.file}\` runs through \`${target}\` with \`${entry.selection_pattern}\` and is forbidden from claiming \`${sectionPrefix(entry.section)}-${phaseNumber}-*\` identifiers.`;
    });
  return [...new Set(supportLines), ...renderBulletLines(extras)];
}

function renderSection(title, entries) {
  const lines = [
    `## ${title}`,
    "",
    "| Row | Evidence | Execution | Claim | Out of scope |",
    "| --- | --- | --- | --- | --- |",
  ];

  for (const entry of entries) {
    lines.push(
      `| \`${entry.id}\` | ${renderEvidence(entry)} | ${renderExecution(entry)} | ${entry.claim} | ${entry.out_of_scope} |`,
    );
  }

  return lines;
}

export function renderPhaseLedger(root, phase) {
  const { manifest } = loadManifest(root, phase);
  const registryEntry = activePhaseRegistryEntry(root, phase);
  if (!registryEntry) {
    throw new Error(`unknown active phase ${phase}`);
  }
  const config = ledgerConfig(manifest, phase, registryEntry);
  const entries = collectEntries(manifest).filter(
    (entry) => entry.coverage === "authoritative",
  );

  const lines = [`# ${config.title}`, "", ...renderIntroduction(phase, config)];
  const supportExecutionLines = renderSupportExecutionLines(
    phase,
    manifest,
    config.supportExecutionExtras,
  );
  if (supportExecutionLines.length > 0) {
    lines.push("", "## Support-Only Execution", "", ...supportExecutionLines);
  }

  for (const [sectionKey, title] of config.sections) {
    const sectionEntries = entries.filter((entry) => entry.section === sectionKey);
    lines.push("", ...renderSection(title, sectionEntries));
  }

  lines.push("", "## Shared Harness Coverage", "", ...config.sharedHarness);
  lines.push("", "## Support-Only Evidence", "", ...renderBulletLines(config.supportOnly));

  return `${lines.join("\n")}\n`;
}

function main(argv) {
  const [phase] = argv;
  if (!phase) {
    throw new Error("usage: render-phase-ledger.mjs <phase>");
  }
  process.stdout.write(renderPhaseLedger(process.cwd(), phase));
}

if (import.meta.url === `file://${process.argv[1]}`) {
  try {
    main(process.argv.slice(2));
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    process.stderr.write(`${message}\n`);
    process.exit(1);
  }
}
