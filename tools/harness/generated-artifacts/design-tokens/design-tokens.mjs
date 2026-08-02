import { createHash, randomUUID } from "node:crypto";
import {
  chmodSync,
  closeSync,
  fsyncSync,
  lstatSync,
  mkdirSync,
  openSync,
  readFileSync,
  renameSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import path from "node:path";

import {
  parseStrictJSON,
  validateSchemaSync,
} from "../../contract/index.mjs";

const schemaID = "cartulary.design_token_registry.v1";
const ownerID = "package.ui";
const verificationID = "package.ui.verification.machine_owned_tokens";
const defaultThemeID = "dark_graphite";
const generatorID = "cartulary.design_token_generation.v2";
const generatedSourceMode = 0o644;
const tokenNamespaces = Object.freeze([
  "border",
  "colors",
  "component",
  "density",
  "elevation",
  "layout",
  "motion",
  "rounded",
  "spacing",
  "typography",
]);
const tokenNamePattern = new RegExp(
  `^--ct-(${tokenNamespaces.join("|")})-[A-Za-z0-9]+(?:-[A-Za-z0-9]+)*$`,
  "u",
);
const safeScalarPattern = /^[A-Za-z0-9_#(),.%+\- /]+$/u;
const allowedFunctions = new Set(["cubic-bezier", "min", "rgba"]);

export class DesignTokenValidationError extends Error {
  constructor(failures) {
    super(
      `design token validation failed: ${failures
        .map((failure) => `${failure.class}: ${failure.message}`)
        .join("; ")}`,
    );
    this.name = "DesignTokenValidationError";
    this.failures = failures;
  }
}

export function compareUnicodeCodePoints(left, right) {
  const leftCodePoints = Array.from(left, (value) => value.codePointAt(0));
  const rightCodePoints = Array.from(right, (value) => value.codePointAt(0));
  const sharedLength = Math.min(leftCodePoints.length, rightCodePoints.length);
  for (let index = 0; index < sharedLength; index += 1) {
    const difference = leftCodePoints[index] - rightCodePoints[index];
    if (difference !== 0) {
      return difference;
    }
  }
  return leftCodePoints.length - rightCodePoints.length;
}

function machineRegistryPathFailure(filePath) {
  const resolved = path.resolve(filePath);
  const pathSegments = resolved.split(path.sep);
  if (pathSegments.includes("docs") || path.extname(resolved) !== ".json") {
    return {
      class: "forbidden_machine_registry_path",
      message: `${resolved}: design token input must be a JSON machine projection outside docs/`,
    };
  }
  return undefined;
}

function tokenNamespace(variableName) {
  return tokenNamePattern.exec(variableName)?.[1];
}

function scalarFailures(name, raw) {
  const failures = [];
  if (raw !== raw.trim() || raw.length === 0 || raw.length > 512) {
    failures.push({
      class: "invalid_scalar",
      message: `${name}: values must be non-empty, trimmed strings of at most 512 characters`,
    });
    return failures;
  }
  if (/var\s*\(|\{[^}]*\}|\$[A-Za-z_-]/u.test(raw)) {
    failures.push({
      class: "unresolved_reference",
      message: `${name}: the machine projection must contain an already-resolved scalar`,
    });
    return failures;
  }
  if (
    /[\u0000-\u001f\u007f-\u009f;{}'"`\\]/u.test(raw) ||
    raw.includes("/*") ||
    raw.includes("*/") ||
    !safeScalarPattern.test(raw)
  ) {
    failures.push({
      class: "unsafe_css_value",
      message: `${name}: value contains syntax outside the design-owned CSS scalar grammar`,
    });
    return failures;
  }
  let depth = 0;
  for (const character of raw) {
    if (character === "(") {
      depth += 1;
    } else if (character === ")") {
      depth -= 1;
      if (depth < 0) {
        break;
      }
    }
  }
  const functions = [...raw.matchAll(/([A-Za-z][A-Za-z0-9-]*)\s*\(/gu)].map(
    (match) => match[1],
  );
  if (depth !== 0 || functions.some((nameValue) => !allowedFunctions.has(nameValue))) {
    failures.push({
      class: "unsafe_css_value",
      message: `${name}: value has unbalanced parentheses or an unapproved CSS function`,
    });
  }
  return failures;
}

function validateRegistrySemantics(registry) {
  const failures = [];
  for (const [field, actual, expected] of [
    ["schema_id", registry.schema_id, schemaID],
    ["owner_id", registry.owner_id, ownerID],
    ["verification_id", registry.verification_id, verificationID],
    ["default_theme_id", registry.default_theme_id, defaultThemeID],
  ]) {
    if (actual !== expected) {
      failures.push({
        class: "invalid_metadata",
        message: `${field} must equal ${JSON.stringify(expected)}`,
      });
    }
  }
  for (const [name, raw] of Object.entries(registry.token_vars ?? {}).sort(
    ([left], [right]) => compareUnicodeCodePoints(left, right),
  )) {
    if (!tokenNamePattern.test(name)) {
      failures.push({
        class: "invalid_token_name",
        message: `${name}: token name is outside the closed CSS-variable namespaces`,
      });
      continue;
    }
    if (typeof raw === "string") {
      failures.push(...scalarFailures(name, raw));
    }
  }
  return failures;
}

export function loadDesignTokenDocument(filePath) {
  const pathFailure = machineRegistryPathFailure(filePath);
  if (pathFailure !== undefined) {
    throw new DesignTokenValidationError([pathFailure]);
  }

  let inputBytes;
  let registry;
  try {
    inputBytes = readFileSync(filePath);
    const source = new TextDecoder("utf-8", { fatal: true }).decode(inputBytes);
    registry = parseStrictJSON(source, filePath);
    validateSchemaSync(schemaID, registry);
  } catch (error) {
    throw new DesignTokenValidationError([
      {
        class: "invalid_machine_registry",
        message: `${filePath}: ${error instanceof Error ? error.message : String(error)}`,
      },
    ]);
  }

  const semanticFailures = validateRegistrySemantics(registry);
  if (semanticFailures.length > 0) {
    throw new DesignTokenValidationError(semanticFailures);
  }
  const entries = Object.entries(registry.token_vars).sort(([left], [right]) =>
    compareUnicodeCodePoints(left, right),
  );
  return {
    metadata: {
      defaultThemeId: registry.default_theme_id,
      generatorId: generatorID,
      inputSha256: createHash("sha256").update(inputBytes).digest("hex"),
      ownerId: registry.owner_id,
      schemaId: registry.schema_id,
      verificationId: registry.verification_id,
    },
    tokenMap: new Map(
      entries.map(([name, raw]) => [
        name,
        { name, namespace: tokenNamespace(name), raw },
      ]),
    ),
    tokenVars: new Map(entries),
  };
}

export function renderDesignTokenTypeScript(document) {
  const tokenVarEntries = [...document.tokenVars.entries()].sort(([left], [right]) =>
    compareUnicodeCodePoints(left, right),
  );
  const cssText = renderCssText(document.metadata.defaultThemeId, tokenVarEntries);
  const tokenVarObject = Object.fromEntries(tokenVarEntries);

  return [
    `// Code generated by ${generatorID}; DO NOT EDIT.`,
    `// Input SHA-256: ${document.metadata.inputSha256}`,
    "",
    `export const cartularyDefaultThemeId = ${JSON.stringify(document.metadata.defaultThemeId)} as const;`,
    "",
    `export const cartularyDesignTokenVars = ${JSON.stringify(tokenVarObject, null, 2)} as const;`,
    "",
    `export const cartularyDesignThemeCssText = ${JSON.stringify(cssText)} as const;`,
    "",
    "export type CartularyDesignTokenVarName = keyof typeof cartularyDesignTokenVars;",
    "export type CartularyDefaultThemeId = typeof cartularyDefaultThemeId;",
    "",
  ].join("\n");
}

function renderCssText(themeId, tokenVarEntries) {
  const declarations = tokenVarEntries.map(([name, value]) => `  ${name}: ${value};`);
  return [`:root,`, `[data-cartulary-theme="${themeId}"] {`, ...declarations, `}`].join(
    "\n",
  );
}

function outputCollision(message) {
  return new DesignTokenValidationError([
    { class: "output_collision", message },
  ]);
}

export function replaceFileAtomically(outputPath, output, options = {}) {
  const resolvedOutput = path.resolve(outputPath);
  const outputDirectory = path.dirname(resolvedOutput);
  let existingOutput;
  try {
    const existing = lstatSync(resolvedOutput);
    if (!existing.isFile()) {
      throw outputCollision(`${resolvedOutput}: existing output is not a regular file`);
    }
    existingOutput = existing;
  } catch (error) {
    if (error?.code !== "ENOENT") {
      throw error;
    }
  }
  mkdirSync(outputDirectory, { recursive: true });
  if (existingOutput && readFileSync(resolvedOutput, "utf8") === output) {
    if ((existingOutput.mode & 0o7777) !== generatedSourceMode) {
      chmodSync(resolvedOutput, generatedSourceMode);
    }
    return;
  }
  const temporaryPath = path.resolve(
    options.temporaryPath ??
      path.join(
        outputDirectory,
        `.${path.basename(resolvedOutput)}.${process.pid}.${randomUUID()}.tmp`,
      ),
  );
  if (
    path.dirname(temporaryPath) !== outputDirectory ||
    temporaryPath === resolvedOutput
  ) {
    throw outputCollision(`${temporaryPath}: temporary output must be a distinct sibling`);
  }

  let descriptor;
  let temporaryCreated = false;
  try {
    descriptor = openSync(temporaryPath, "wx", 0o600);
    temporaryCreated = true;
    writeFileSync(descriptor, output, "utf8");
    fsyncSync(descriptor);
    chmodSync(temporaryPath, generatedSourceMode);
    closeSync(descriptor);
    descriptor = undefined;
    renameSync(temporaryPath, resolvedOutput);
    temporaryCreated = false;
    const directoryDescriptor = openSync(outputDirectory, "r");
    try {
      fsyncSync(directoryDescriptor);
    } finally {
      closeSync(directoryDescriptor);
    }
  } catch (error) {
    if (descriptor !== undefined) {
      closeSync(descriptor);
    }
    if (temporaryCreated) {
      rmSync(temporaryPath, { force: true });
    }
    if (error?.code === "EEXIST") {
      throw outputCollision(`${temporaryPath}: sibling temporary output already exists`);
    }
    throw error;
  }
}
