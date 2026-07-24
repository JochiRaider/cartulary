import { spawnSync } from "node:child_process";
import { existsSync, readFileSync, realpathSync } from "node:fs";
import path from "node:path";

import { validateSchemaSync } from "../contract/index.mjs";

const exceptionSchemaID = "cartulary.documentation_read_exceptions.v1";
const exceptionPath = "tools/documentation_read_exceptions.json";
const sourceExtensions = new Set([
  ".cjs",
  ".go",
  ".js",
  ".mjs",
  ".sh",
  ".ts",
  ".tsx",
]);
const operationNames = {
  existsSync: "stat_path",
  lstatSync: "stat_path",
  readFile: "read_file",
  readFileSync: "read_file",
  readdir: "enumerate_directory",
  readdirSync: "enumerate_directory",
  realpath: "resolve_realpath",
  realpathSync: "resolve_realpath",
  stat: "stat_path",
  statSync: "stat_path",
};

function readJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

function normalize(value) {
  return value.replaceAll("\\", "/").replace(/^\.\//u, "");
}

function escapeRegExp(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/gu, "\\$&");
}

function trackedExecutableSources(root) {
  const result = spawnSync(
    "git",
    ["ls-files", "--cached", "--others", "--exclude-standard", "-z"],
    {
      cwd: root,
      encoding: "buffer",
    },
  );
  if (result.status !== 0) {
    throw new Error(`git ls-files failed: ${String(result.stderr)}`);
  }
  return result.stdout
    .toString("utf8")
    .split("\0")
    .filter(Boolean)
    .filter((file) => existsSync(path.join(root, file)))
    .filter((file) => {
      const normalized = normalize(file);
      const basename = path.posix.basename(normalized);
      const makeSource = basename === "Makefile" || normalized.endsWith(".mk");
      const packageScriptSource = basename === "package.json";
      return (
        makeSource ||
        packageScriptSource ||
        (sourceExtensions.has(path.extname(normalized)) &&
          /^(?:cmd|internal|apps|packages|scripts|tools)\//u.test(normalized))
      );
    })
    .sort((left, right) => left.localeCompare(right));
}

function documentationPathFromExpression(expression) {
  const direct = expression.match(/(?:^|["'`/])docs\/(?:spec\/)?[^"'`\s),;]*/u);
  if (direct) {
    return normalize(direct[0].replace(/^["'`/]/u, ""));
  }
  if (/(["'`])docs\1/u.test(expression)) {
    return "docs/";
  }
  return null;
}

function lineForIndex(source, index) {
  return source.slice(0, index).split("\n").length;
}

export function scanDocumentationReadSource(consumerPath, source) {
  const findings = [];
  const documentationVariables = new Map();
  for (const match of source.matchAll(
    /\b(?:const|let|var)\s+([A-Za-z_$][A-Za-z0-9_$]*)\s*=\s*([^;\n]+(?:\n[^;]+)?);/gu,
  )) {
    const documentationPath = documentationPathFromExpression(match[2]);
    if (documentationPath) {
      documentationVariables.set(match[1], documentationPath);
    }
  }

  const operationPattern = new RegExp(
    `\\b(${Object.keys(operationNames).join("|")})\\s*\\(([^;]{0,800})`,
    "gu",
  );
  for (const match of source.matchAll(operationPattern)) {
    const expression = match[2].split(/\n?\}\)?\s*;|\n\s*(?:const|let|return)\b/u)[0];
    let documentationPath = documentationPathFromExpression(expression);
    if (!documentationPath) {
      for (const [variable, variablePath] of documentationVariables) {
        if (new RegExp(`\\b${variable}\\b`, "u").test(expression)) {
          documentationPath = variablePath;
          break;
        }
      }
    }
    if (!documentationPath) {
      continue;
    }
    findings.push({
      consumer_path: normalize(consumerPath),
      documentation_path: documentationPath,
      operation: operationNames[match[1]],
      line: lineForIndex(source, match.index),
    });
  }

  const helperOperations = new Map();
  for (const match of source.matchAll(
    /\bfunction\s+([A-Za-z_$][A-Za-z0-9_$]*)\s*\(([^)]*)\)\s*\{([^}]{0,1600})\}/gu,
  )) {
    const parameters = match[2]
      .split(",")
      .map((parameter) => parameter.trim())
      .filter((parameter) => /^[A-Za-z_$][A-Za-z0-9_$]*$/u.test(parameter));
    for (const [operationName, operation] of Object.entries(operationNames)) {
      if (
        parameters.some((parameter) =>
          new RegExp(
            `\\b${operationName}\\s*\\(\\s*${escapeRegExp(parameter)}\\b`,
            "u",
          ).test(match[3]),
        )
      ) {
        helperOperations.set(match[1], operation);
      }
    }
  }
  for (const [helper, operation] of helperOperations) {
    const helperCallPattern = new RegExp(`\\b${helper}\\s*\\(([^;]{0,800})`, "gu");
    for (const match of source.matchAll(helperCallPattern)) {
      const expression = match[1].split(/\n?\}\)?\s*;|\n\s*(?:const|let|return)\b/u)[0];
      let documentationPath = documentationPathFromExpression(expression);
      if (!documentationPath) {
        for (const [variable, variablePath] of documentationVariables) {
          if (new RegExp(`\\b${variable}\\b`, "u").test(expression)) {
            documentationPath = variablePath;
            break;
          }
        }
      }
      if (!documentationPath) {
        continue;
      }
      findings.push({
        consumer_path: normalize(consumerPath),
        documentation_path: documentationPath,
        operation,
        line: lineForIndex(source, match.index),
      });
    }
  }

  for (const match of source.matchAll(
    /\b(?:cat|cp|readlink|realpath|stat|test\s+-[efd])\s+([^\n]*docs\/[^\n]*)/gu,
  )) {
    findings.push({
      consumer_path: normalize(consumerPath),
      documentation_path: documentationPathFromExpression(match[1]) ?? "docs/",
      operation: match[0].startsWith("cat ") ? "read_file" : "stat_path",
      line: lineForIndex(source, match.index),
    });
  }

  return findings;
}

function exceptionAllows(exception, finding) {
  if (exception.consumer_path !== finding.consumer_path) {
    return false;
  }
  if (!exception.operations.includes(finding.operation)) {
    return false;
  }
  return new RegExp(exception.documentation_pattern, "u").test(
    finding.documentation_path,
  );
}

export function loadDocumentationReadExceptions(root) {
  const value = readJSON(path.join(root, exceptionPath));
  validateSchemaSync(exceptionSchemaID, value);
  const keys = value.exceptions.map(
    (entry) =>
      `${entry.consumer_path}\0${entry.documentation_pattern}\0${entry.operations.join(",")}`,
  );
  const sorted = [...keys].sort((left, right) => left.localeCompare(right));
  if (JSON.stringify(keys) !== JSON.stringify(sorted)) {
    throw new Error(`${exceptionPath}.exceptions must be ASCII-sorted`);
  }
  if (new Set(keys).size !== keys.length) {
    throw new Error(`${exceptionPath}.exceptions must not contain duplicates`);
  }
  return value;
}

export function scanExecutableDocumentationReads(root) {
  const exceptions = loadDocumentationReadExceptions(root);
  const findings = [];
  for (const file of trackedExecutableSources(root)) {
    const source = readFileSync(path.join(root, file), "utf8");
    for (const finding of scanDocumentationReadSource(file, source)) {
      if (!exceptions.exceptions.some((entry) => exceptionAllows(entry, finding))) {
        findings.push(finding);
      }
    }
  }
  if (findings.length > 0) {
    findings.sort((left, right) =>
      `${left.consumer_path}\0${left.line}\0${left.operation}`.localeCompare(
        `${right.consumer_path}\0${right.line}\0${right.operation}`,
      ),
    );
    const details = findings
      .map(
        (finding) =>
          `${finding.consumer_path}:${finding.line} ${finding.operation} ${finding.documentation_path}`,
      )
      .join("; ");
    throw new Error(`unauthorized documentation access: ${details}`);
  }
  return findings;
}

function pathUnder(candidate, parent) {
  return candidate === parent || candidate.startsWith(`${parent}${path.sep}`);
}

export function assertDocumentationAccessAllowed({
  root,
  consumerPath,
  operation,
  candidatePath,
  exceptions = null,
}) {
  const resolvedRoot = realpathSync(root);
  const docsRoot = path.join(resolvedRoot, "docs");
  const lexical = path.resolve(resolvedRoot, candidatePath);
  let resolved = lexical;
  if (existsSync(lexical)) {
    resolved = realpathSync(lexical);
  }
  if (!pathUnder(lexical, docsRoot) && !pathUnder(resolved, docsRoot)) {
    return;
  }
  const finding = {
    consumer_path: normalize(consumerPath),
    documentation_path: normalize(path.relative(resolvedRoot, resolved)),
    operation,
  };
  const policy = exceptions ?? loadDocumentationReadExceptions(resolvedRoot);
  if (policy.exceptions.some((entry) => exceptionAllows(entry, finding))) {
    return;
  }
  throw new Error(
    `boundary_policy_violation: ${finding.consumer_path} may not ${operation} ${finding.documentation_path}`,
  );
}
