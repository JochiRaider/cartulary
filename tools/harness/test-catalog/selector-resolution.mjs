import { lstatSync, readFileSync, readdirSync, realpathSync } from "node:fs";
import path from "node:path";
import { HarnessConfigError } from "../contract/index.mjs";
import { readExecutableRepositoryJSON } from "./restricted-input-boundary.mjs";
import { goRunnerBuildContext } from "./go-build-context.mjs";
import { discoverRunnerSources, excludedSourcePath, frontendSourceProject, sourceExclusionPolicy } from "./source-discovery.mjs";


const goOperatingSystems = new Set([
  "aix",
  "android",
  "darwin",
  "dragonfly",
  "freebsd",
  "illumos",
  "ios",
  "js",
  "linux",
  "netbsd",
  "openbsd",
  "plan9",
  "solaris",
  "wasip1",
  "windows",
]);
const goArchitectures = new Set([
  "386",
  "amd64",
  "arm",
  "arm64",
  "loong64",
  "mips",
  "mips64",
  "mips64le",
  "mipsle",
  "ppc64",
  "ppc64le",
  "riscv64",
  "s390x",
  "wasm",
]);

function asciiCompare(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function pathUnder(candidate, parent) {
  return candidate === parent || candidate.startsWith(`${parent}${path.sep}`);
}

function assertNoSymlinkComponents(root, relativePath, label) {
  let current = root;
  for (const segment of relativePath.split(path.posix.sep)) {
    current = path.join(current, segment);
    if (lstatSync(current).isSymbolicLink()) {
      throw new Error(`${label} must not traverse a symbolic link`);
    }
  }
}

function sourcePathExists(root, relativePath, label) {
  try { assertNoSymlinkComponents(root, relativePath, label); return true; }
  catch (error) { if (error.code === "ENOENT") return false; throw error; }
}

function assertSourceBoundary(relative, approvedRoots, label) {
  if (excludedSourcePath(relative)) throw new Error(`${label} is an excluded source path`);
  if (!approvedRoots.some((approved) => relative === approved || relative.startsWith(`${approved}/`))) {
    throw new Error(`${label} is outside approved runner roots`);
  }
}

function containedFile(root, relativePath, approvedRoots, label, scope) {
  if (path.isAbsolute(relativePath) || relativePath.includes("\\")) {
    throw new Error(`${label} must be a normalized repository-relative path`);
  }
  const normalized = path.posix.normalize(relativePath);
  if (normalized !== relativePath || normalized.startsWith("../")) {
    throw new Error(`${label} contains traversal or normalization drift`);
  }
  if (/[*?{}[\]]/u.test(relativePath)) {
    throw new Error(`${label} must not contain a glob`);
  }
  assertSourceBoundary(relativePath, approvedRoots, label);
  const lexical = path.resolve(root, relativePath);
  const resolvedRoot = realpathSync(root);
  if (!sourcePathExists(root, relativePath, label)) throw new Error(`${label} does not exist: ${relativePath}`);
  const stat = lstatSync(lexical);
  if (!stat.isFile() || stat.isSymbolicLink()) {
    throw new Error(`${label} must resolve to a non-symlink regular file`);
  }
  const resolved = realpathSync(lexical);
  if (!pathUnder(resolved, resolvedRoot)) {
    throw new Error(`${label} escapes the repository`);
  }
  if (
    approvedRoots.length > 0 &&
    !approvedRoots.some((approvedRoot) =>
      pathUnder(resolved, realpathSync(path.resolve(root, approvedRoot))),
    )
  ) {
    throw new Error(`${label} is outside approved runner roots`);
  }
  return { lexical };
}

// A Playwright title is its title path in Playwright's display form: suite
// components and the test title joined by " › ". No component may contain
// the separator, so the path is exact and needs no source parsing.
const playwrightSeparator = " \u203a ";

export function playwrightTitlePath(title, label = "Playwright title") {
  const parts = String(title).split(playwrightSeparator);
  if (parts.some((part) => !part || part.trim() !== part || /[\r\n\u203a]/u.test(part))) {
    throw new Error(`${label} ${JSON.stringify(title)} is not an exact Playwright title path`);
  }
  return parts;
}

// The test's own title, as Playwright reports it for a spec.
export function playwrightLeafTitle(title) {
  return playwrightTitlePath(title).at(-1);
}

// The exact test-list selection of Playwright rows, from their catalog titles.
export function playwrightSelection(rows) {
  return rows.flatMap((row) => {
    if (row.runner !== "playwright") throw new Error("browser selection requires Playwright rows");
    return row.selector.titles.map((title) => ({
      row_id: row.row_id, project_id: row.selector.project_id, file: row.selector.file,
      title_path: playwrightTitlePath(title, row.row_id),
    }));
  });
}

function packageDirectory(root, packagePath, approvedRoots, label) {
  if (!packagePath.startsWith("./") || /[*?{}[\]\\]/u.test(packagePath)) {
    throw new Error(`${label} must be an exact repository package`);
  }
  const relative = packagePath.slice(2);
  if (path.posix.normalize(relative) !== relative || relative.split("/").includes("..")) {
    throw new Error(`${label} contains traversal or normalization drift`);
  }
  assertSourceBoundary(relative, approvedRoots, label);
  const lexical = path.resolve(root, relative);
  if (!sourcePathExists(root, relative, label)) {
    throw new Error(`${label} does not exist: ${packagePath}`);
  }
  const stat = lstatSync(lexical);
  if (!stat.isDirectory() || stat.isSymbolicLink()) {
    throw new Error(`${label} must resolve to a non-symlink directory`);
  }
  const resolved = realpathSync(lexical);
  const resolvedRoot = realpathSync(root);
  if (!pathUnder(resolved, resolvedRoot)) {
    throw new Error(`${label} escapes the repository`);
  }
  if (
    !approvedRoots.some((approvedRoot) =>
      pathUnder(resolved, realpathSync(path.resolve(root, approvedRoot))),
    )
  ) {
    throw new Error(`${label} is outside approved runner roots`);
  }
  return lexical;
}

function fileNameMatchesGoBuildContext(fileName, context) {
  const stem = fileName.replace(/_test\.go$/u, "");
  const segments = stem.split("_");
  const last = segments.at(-1);
  const previous = segments.at(-2);
  if (goOperatingSystems.has(previous) && goArchitectures.has(last)) {
    return previous === context.goos && last === context.goarch;
  }
  if (goOperatingSystems.has(last)) return last === context.goos;
  if (goArchitectures.has(last)) return last === context.goarch;
  return true;
}

function tokenizeGoBuildExpression(expression, label) {
  const tokens = [];
  let cursor = 0;
  while (cursor < expression.length) {
    if (/\s/u.test(expression[cursor])) {
      cursor += 1;
      continue;
    }
    const operator = expression.slice(cursor, cursor + 2);
    if (operator === "&&" || operator === "||") {
      tokens.push(operator);
      cursor += 2;
      continue;
    }
    if (expression[cursor] === "!" || expression[cursor] === "(" || expression[cursor] === ")") {
      tokens.push(expression[cursor]);
      cursor += 1;
      continue;
    }
    const match = /^[A-Za-z0-9_.]+/u.exec(expression.slice(cursor));
    if (!match) throw new Error(`${label} contains invalid Go build constraint syntax`);
    tokens.push(match[0]);
    cursor += match[0].length;
  }
  return tokens;
}

function evaluateGoBuildExpression(expression, context, label) {
  const tokens = tokenizeGoBuildExpression(expression, label);
  if (tokens.some((tag) => tag.startsWith("goexperiment."))) {
    throw new Error(`${label} uses an unsupported experimental Go build context`);
  }
  let cursor = 0;
  const current = () => tokens[cursor];
  const consume = () => tokens[cursor++];
  const parseUnary = () => {
    if (current() === "!") {
      consume();
      return !parseUnary();
    }
    if (current() === "(") {
      consume();
      const value = parseOr();
      if (consume() !== ")") throw new Error(`${label} has an unmatched parenthesis`);
      return value;
    }
    const tag = consume();
    if (!tag || ["&&", "||", ")"].includes(tag)) {
      throw new Error(`${label} contains an incomplete Go build constraint`);
    }
    return context.tags.has(tag);
  };
  const parseAnd = () => {
    let value = parseUnary();
    while (current() === "&&") {
      consume();
      const right = parseUnary();
      value = value && right;
    }
    return value;
  };
  const parseOr = () => {
    let value = parseAnd();
    while (current() === "||") {
      consume();
      const right = parseAnd();
      value = value || right;
    }
    return value;
  };
  if (tokens.length === 0) throw new Error(`${label} has an empty Go build constraint`);
  const result = parseOr();
  if (cursor !== tokens.length) throw new Error(`${label} contains trailing Go build constraint tokens`);
  return result;
}

function sourceMatchesGoBuildContext(source, context, label) {
  const expressions = [...source.matchAll(/^\/\/go:build[ \t]+(.+)$/gmu)].map((match) => match[1].trim());
  if (expressions.length > 1) throw new Error(`${label} contains multiple //go:build lines`);
  if (expressions.length === 1) {
    return evaluateGoBuildExpression(expressions[0], context, label);
  }
  if (/^\/\/ \+build[ \t]+/mu.test(source)) {
    throw new Error(`${label} uses unsupported legacy // +build syntax`);
  }
  return true;
}

function goSymbols(root, directory, scope, discovery = false) {
  const cacheKey = `${discovery ? "discovery" : "selection"}:${directory}`;
  if (scope.packages.has(cacheKey)) return scope.packages.get(cacheKey);
  const context = scope.goContext ??= goRunnerBuildContext(root);
  const activeCounts = new Map();
  const excludedFiles = new Map();
  const excluded = discovery ? (scope.excluded ??= sourceExclusionPolicy(root)) : () => false;
  for (const entry of readdirSync(directory, { withFileTypes: true })) {
    const relative = path.relative(root, path.join(directory, entry.name)).split(path.sep).join("/");
    if (excluded(relative)) continue;
    if (!entry.isFile() || !entry.name.endsWith("_test.go") || /^[._]/u.test(entry.name)) {
      continue;
    }
    const source = readFileSync(path.join(directory, entry.name), "utf8");
    const active = fileNameMatchesGoBuildContext(entry.name, context) &&
      sourceMatchesGoBuildContext(source, context, entry.name);
    // Comments and literals can contain apparent declarations. Preserve lines
    // while masking them; only source-level declarations may become selectors.
    const declarations = source.replace(/\/\*[\s\S]*?\*\/|\/\/[^\n]*|`[^`]*`|"(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*'/gu,
      (literal) => literal.replace(/[^\n]/gu, " "));
    for (const match of declarations.matchAll(/^func\s+(Test[\p{L}\p{Nd}_]*)\s*\(/gmu)) {
      // Go treats the suffix as a test unless its first rune is lowercase.
      // The current operational selector contract admits ASCII names only;
      // unsupported executable identities must be visible, never omitted.
      if (/^Test\p{Ll}/u.test(match[1])) continue;
      if (active && !/^Test[A-Za-z0-9_]*$/u.test(match[1])) {
        throw new Error(`${relative}: unsupported Go test identity ${match[1]}; explicit runner contract admission is required`);
      }
      if (active) activeCounts.set(match[1], (activeCounts.get(match[1]) ?? 0) + 1);
      else {
        const files = excludedFiles.get(match[1]) ?? [];
        files.push(entry.name);
        excludedFiles.set(match[1], files);
      }
    }
  }
  const symbols = { activeCounts, context, excludedFiles };
  scope.packages.set(cacheKey, symbols);
  return symbols;
}

function resolveScopedRowSelector({ root, row, runner, operationIDs }, scope) {
  const label = `${row.row_id}.selector`;
  if (row.runner === "go") {
    const directory = packageDirectory(root, row.selector.package, runner.approved_roots, `${label}.package`);
    const symbols = goSymbols(root, directory, scope);
    for (const symbol of row.selector.tests) {
      if ((symbols.activeCounts.get(symbol) ?? 0) === 0 && symbols.excludedFiles.has(symbol)) {
        throw new Error(
          `${label}.tests ${symbol} is excluded from the Go runner build context ` +
          `${symbols.context.label}: ${symbols.excludedFiles.get(symbol).sort(asciiCompare).join(", ")}`,
        );
      }
      if ((symbols.activeCounts.get(symbol) ?? 0) !== 1) {
        throw new Error(`${label}.tests ${symbol} must resolve exactly once`);
      }
    }
    return row.selector.tests.map((symbol) => `go:${row.selector.package}:${symbol}`);
  }
  if (row.runner === "vitest") {
    // Rows carry exact full titles. The pinned runner selects them by exact
    // name and fails a title it does not collect exactly once, so the catalog
    // admits the file and project without parsing the source.
    containedFile(root, row.selector.file, runner.approved_roots, `${label}.file`, scope);
    if (runner.project_ids && !runner.project_ids.includes(row.selector.project_id)) throw new Error(`${label}.project_id is not registered`);
    if (row.selector.project_id !== frontendSourceProject(row.selector.file)) throw new Error(`${label}.project_id does not run ${row.selector.file}`);
    return row.selector.titles.map((title) => `vitest:${row.selector.project_id}:${row.selector.file}:${title}`);
  }
  if (row.runner === "playwright") {
    // Titles are exact title paths; the pinned runner fails a path it does
    // not report exactly once and any test it runs outside the selection.
    containedFile(root, row.selector.file, runner.approved_roots, `${label}.file`, scope);
    for (const title of row.selector.titles) playwrightTitlePath(title, `${label}.titles`);
    if (!runner.project_ids.includes(row.selector.project_id)) {
      throw new Error(`${label}.project_id is not registered`);
    }
    if (!runner.stages.includes(row.selector.stage)) {
      throw new Error(`${label}.stage is not registered`);
    }
    if (row.selector.scenario_ids.length !== row.selector.titles.length) {
      throw new Error(`${label} scenario_ids and titles must have equal length`);
    }
    return row.selector.scenario_ids.map(
      (scenarioID) => `playwright:${row.selector.project_id}:${row.selector.stage}:${scenarioID}`,
    ).concat(row.selector.titles.map((title) =>
      `playwright-source:${row.selector.project_id}:${row.selector.file}:${title}`,
    ));
  }
  if (row.runner === "shell") {
    if (!operationIDs.has(row.selector.operation_id)) {
      throw new Error(`${label}.operation_id is not registered`);
    }
    return [`shell:${row.selector.operation_id}`];
  }
  throw new Error(`${row.row_id}.runner is unsupported`);
}

// Facts belong to one synchronous catalog load. Sharing them across its rows
// avoids rescanning packages and reparsing source; a fresh load observes edits,
// imported registration suites and changed toolchain pins without invalidation
// heuristics or process-global source retention. Path checks remain per selector.
export function createRowSelectorResolver(root) {
  const scope = { packages: new Map(), goContext: null };
  const resolve = (options) => resolveScopedRowSelector({ ...options, root }, scope);
  // An exact Playwright test-list line that names a suite prefix of another
  // cataloged path would select that path's tests too.
  resolve.assertTitlePaths = (selected) => {
    const paths = Map.groupBy([...selected].filter((key) => key.startsWith("playwright-source:")).map((key) => {
      const [, project, file, ...title] = key.split(":");
      return { scope: `${project}:${file}`, path: playwrightTitlePath(title.join(":")) };
    }), (entry) => entry.scope);
    for (const [scope, entries] of paths) {
      for (const entry of entries) {
        if (entries.some((other) => other.path.length > entry.path.length && entry.path.every((part, index) => other.path[index] === part))) {
          throw new Error(`${scope}: Playwright title path ${JSON.stringify(entry.path)} is a suite prefix of another cataloged path`);
        }
      }
    }
  };
  // Go coverage is per test symbol; Vitest and Playwright coverage is per
  // file: every discovered test file has a row.
  resolve.assertSourceCoverage = (runners, selected) => {
    const missing = [];
    for (const runner of runners.values()) {
      if (runner.runner === "shell") continue;
      const files = discoverRunnerSources(root, runner);
      if (runner.runner === "go") {
        const packages = [...new Set(files.map((file) => `./${path.posix.dirname(file)}`))];
        for (const packageName of packages) {
          const directory = packageDirectory(root, packageName, runner.approved_roots, "source coverage");
          const { activeCounts } = goSymbols(root, directory, scope, true);
          for (const [symbol, count] of activeCounts) {
            if (symbol === "TestMain") continue;
            if (count !== 1 || !selected.has(`go:${packageName}:${symbol}`)) missing.push(`go:${packageName}:${symbol}`);
          }
        }
      } else {
        const prefix = runner.runner === "vitest" ? "vitest:" : "playwright-source:";
        const covered = new Set([...selected].filter((key) => key.startsWith(prefix)).map((key) => key.split(":")[2]));
        for (const file of files) {
          if (!covered.has(file)) missing.push(runner.runner === "vitest" ? `vitest:${frontendSourceProject(file)}:${file}` : `playwright-source:chromium:${file}`);
        }
      }
    }
    if (missing.length) {
      missing.sort(asciiCompare);
      throw new HarnessConfigError(`${missing.length} source tests are not cataloged exactly once:\n${missing.join("\n")}`, {
        diagnosticCause: "catalog_coverage",
        subjects: missing,
      });
    }
  };
  return resolve;
}

export function resolveRowSelector({ root, ...options }) {
  return createRowSelectorResolver(root)(options);
}

export function assertSortedUnique(values, label) {
  const sorted = [...values].sort(asciiCompare);
  if (JSON.stringify(values) !== JSON.stringify(sorted)) {
    throw new Error(`${label} must be ASCII-sorted`);
  }
  if (new Set(values).size !== values.length) {
    throw new Error(`${label} must not contain duplicates`);
  }
}
