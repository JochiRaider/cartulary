#!/usr/bin/env node
import { createHash } from "node:crypto";
import {
  existsSync,
  mkdirSync,
  readFileSync,
  readdirSync,
  rmSync,
  statSync,
  writeFileSync,
} from "node:fs";
import os from "node:os";
import path from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const defaultRepoRoot = path.resolve(scriptDir, "../../..");
const manifestPath = "apps/web/public/assets/fonts/FONT_MANIFEST.json";
const fontsCSSPath = "apps/web/public/assets/fonts/fonts.css";
const noticePath = "apps/web/public/assets/fonts/NOTICE.fonts.md";
const remoteFontPattern =
  /\b(?:https?:)?\/\/(?:fonts\.googleapis\.com|fonts\.gstatic\.com|use\.typekit\.net|p\.typekit\.net|cdn\.jsdelivr\.net\/npm\/@fontsource|rsms\.me\/inter|unpkg\.com\/@fontsource)/iu;
const scanExtensions = new Set([
  ".css",
  ".go",
  ".html",
  ".js",
  ".jsx",
  ".mjs",
  ".ts",
  ".tsx",
]);
const ignoredDirectories = new Set([
  ".cache",
  ".git",
  ".pnpm-store",
  "coverage",
  "dist",
  "node_modules",
  "playwright-report",
  "test-results",
  "tmp",
]);
const activationStatuses = new Set([
  "active_default",
  "active_selector",
  "staged_inactive",
]);
const expectedFontRoleMetadata = new Map([
  [
    "Inter",
    {
      roles: ["ui", "grid", "grid-cell"],
      status: "active_default",
    },
  ],
  [
    "JetBrains Mono",
    {
      roles: ["mono"],
      status: "active_default",
    },
  ],
  [
    "Geist",
    {
      roles: ["alternate-ui"],
      status: "staged_inactive",
    },
  ],
  [
    "Geist Mono",
    {
      roles: ["alternate-mono"],
      status: "staged_inactive",
    },
  ],
  [
    "Source Serif 4",
    {
      roles: ["report-narrative"],
      status: "staged_inactive",
    },
  ],
  [
    "Atkinson Hyperlegible",
    {
      roles: ["accessible-reading"],
      status: "active_selector",
      selectors: ['[data-reading-profile="hyperlegible"]'],
      sourceNeedles: ["data-reading-profile", "hyperlegible"],
    },
  ],
  [
    "IBM Plex Sans Condensed",
    {
      roles: ["compact-metadata"],
      status: "active_selector",
      selectors: ['[data-density-role="narrow-metadata"]'],
      sourceNeedles: ["data-density-role", "narrow-metadata"],
    },
  ],
]);

function usage() {
  throw new Error("usage: check-font-bundle.mjs [--root <path>]");
}

function parseArgs(argv) {
  const options = {
    root: process.env.CARTULARY_REPO_ROOT ?? defaultRepoRoot,
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--root") {
      options.root = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    usage();
  }
  if (options.root.trim() === "") {
    usage();
  }
  return options;
}

function repoPath(root, relativePath) {
  return path.join(root, relativePath);
}

function repoRelative(root, file) {
  return path.relative(root, file).replaceAll(path.sep, "/") || ".";
}

function readJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

function sha256File(file) {
  return createHash("sha256").update(readFileSync(file)).digest("hex");
}

function collectFiles(root, relativeRoots) {
  const files = [];
  const visit = (file) => {
    if (!existsSync(file)) {
      return;
    }
    const stats = statSync(file);
    if (stats.isDirectory()) {
      if (ignoredDirectories.has(path.basename(file))) {
        return;
      }
      for (const name of readdirSync(file)) {
        visit(path.join(file, name));
      }
      return;
    }
    if (stats.isFile() && scanExtensions.has(path.extname(file))) {
      files.push(file);
    }
  };
  for (const relativeRoot of relativeRoots) {
    visit(repoPath(root, relativeRoot));
  }
  return files.sort((left, right) => left.localeCompare(right));
}

function licenseFileForFamily(root, directory) {
  const dir = repoPath(root, `apps/web/public/assets/fonts/${directory}`);
  if (!existsSync(dir)) {
    return null;
  }
  return (
    readdirSync(dir)
      .filter((name) => /^(?:OFL|LICENSE)(?:\.[^.]+)?$/iu.test(name))
      .sort()[0] ?? null
  );
}

function assert(condition, message, failures) {
  if (!condition) {
    failures.push(message);
  }
}

function collectFontFiles(root) {
  const base = repoPath(root, "apps/web/public/assets/fonts");
  const files = [];
  const visit = (file) => {
    if (!existsSync(file)) {
      return;
    }
    const stats = statSync(file);
    if (stats.isDirectory()) {
      for (const name of readdirSync(file)) {
        visit(path.join(file, name));
      }
      return;
    }
    if (stats.isFile() && path.extname(file) === ".woff2") {
      files.push(repoRelative(base, file));
    }
  };
  visit(base);
  return files.sort((left, right) => left.localeCompare(right));
}

function collectSourceText(root) {
  return collectFiles(root, ["apps/web/src"])
    .filter((file) => !/\.test\.[cm]?[jt]sx?$/u.test(file))
    .map((file) => readFileSync(file, "utf8"))
    .join("\n");
}

function cssURLSet(fontsCSS) {
  return new Set(
    [...fontsCSS.matchAll(/url\("([^"]+)"\)/giu)].map((match) =>
      decodeURIComponent(match[1]).replace(/^\.\//u, ""),
    ),
  );
}

function declaredFontVariables(fontsCSS) {
  return [
    ...new Set(
      [...fontsCSS.matchAll(/(--font-[a-z0-9-]+)\s*:/giu)].map(
        (match) => match[1],
      ),
    ),
  ].sort((left, right) => left.localeCompare(right));
}

function cssVariableIsConsumed(fontsCSS, variableName) {
  const escapedVariableName = variableName.replace(
    /[.*+?^${}()|[\]\\]/gu,
    "\\$&",
  );
  return new RegExp(`var\\(\\s*${escapedVariableName}\\b`, "u").test(
    fontsCSS,
  );
}

function stagedCssVariables(manifest) {
  return new Set(
    (manifest.families ?? [])
      .filter((family) => family?.activation_status === "staged_inactive")
      .flatMap((family) => family?.staged_css_variables ?? []),
  );
}

function validateRoleMetadata({
  family,
  fontsCSS,
  manifestPathLabel,
  notice,
  sourceText,
  failures,
}) {
  const label = family?.family ?? "<missing family>";
  const roleIDs = Array.isArray(family?.role_ids) ? family.role_ids : [];
  const expected = expectedFontRoleMetadata.get(family?.family);

  assert(
    activationStatuses.has(family?.activation_status),
    `${label}: activation_status is required and must be active_default, active_selector, or staged_inactive`,
    failures,
  );
  assert(roleIDs.length > 0, `${label}: role_ids must be non-empty`, failures);
  if (expected) {
    assert(
      family.activation_status === expected.status,
      `${label}: activation_status must be ${expected.status}`,
      failures,
    );
    for (const role of expected.roles) {
      assert(
        roleIDs.includes(role),
        `${label}: role_ids must include ${role}`,
        failures,
      );
    }
  }

  if (family?.activation_status === "active_default") {
    assert(
      family.active_by_default === true,
      `${label}: active_default families must set active_by_default true`,
      failures,
    );
    assert(
      Array.isArray(family.activation_selectors) &&
        family.activation_selectors.length > 0,
      `${label}: active_default families must declare activation_selectors`,
      failures,
    );
  }

  if (family?.activation_status === "active_selector") {
    assert(
      family.active_by_default === false,
      `${label}: active_selector families must not be active_by_default`,
      failures,
    );
    assert(
      Array.isArray(family.activation_selectors) &&
        family.activation_selectors.length > 0,
      `${label}: active_selector families must declare activation_selectors`,
      failures,
    );
    for (const selector of expected?.selectors ?? family.activation_selectors ?? []) {
      assert(
        fontsCSS.includes(selector),
        `${label}: active selector ${selector} must exist in ${fontsCSSPath}`,
        failures,
      );
    }
    for (const needle of expected?.sourceNeedles ?? []) {
      assert(
        sourceText.includes(needle),
        `${label}: active selector source must render ${needle}`,
        failures,
      );
    }
  }

  if (family?.activation_status === "staged_inactive") {
    assert(
      family.active_by_default === false,
      `${label}: staged_inactive families must not be active_by_default`,
      failures,
    );
    assert(
      typeof family.staging_reason === "string" &&
        family.staging_reason.trim() !== "",
      `${manifestPathLabel}: staged_inactive family ${label} must declare staging_reason`,
      failures,
    );
    assert(
      !Array.isArray(family.activation_selectors) ||
        family.activation_selectors.length === 0,
      `${label}: staged_inactive families must not declare activation_selectors`,
      failures,
    );
    assert(
      notice.toLowerCase().includes("staged") && notice.includes(label),
      `${noticePath}: staged_inactive family ${label} must be documented as staged`,
      failures,
    );
  }
}

export function checkFontBundle(root = defaultRepoRoot) {
  const failures = [];
  const manifestFile = repoPath(root, manifestPath);
  const fontsCSSFile = repoPath(root, fontsCSSPath);
  const noticeFile = repoPath(root, noticePath);

  assert(existsSync(manifestFile), `${manifestPath}: missing`, failures);
  assert(existsSync(fontsCSSFile), `${fontsCSSPath}: missing`, failures);
  assert(existsSync(noticeFile), `${noticePath}: missing`, failures);
  if (failures.length > 0) {
    return failures;
  }

  const manifest = readJSON(manifestFile);
  const fontsCSS = readFileSync(fontsCSSFile, "utf8");
  const notice = readFileSync(noticeFile, "utf8");
  const sourceText = collectSourceText(root);

  assert(
    manifest.schema_id === "cartulary.font_manifest.v1",
    `${manifestPath}: schema_id must be cartulary.font_manifest.v1`,
    failures,
  );
  assert(
    Array.isArray(manifest.families) && manifest.families.length > 0,
    `${manifestPath}: families must be a non-empty array`,
    failures,
  );
  assert(!/local\s*\(/iu.test(fontsCSS), `${fontsCSSPath}: must not use local(...)`, failures);

  const defaultTokenBlock = fontsCSS.match(
    /--font-(?:ui|grid|mono):[^;]+;/giu,
  )?.join("\n") ?? "";
  for (const inactiveFamily of [
    "Geist",
    "Geist Mono",
    "Source Serif 4",
    "Atkinson Hyperlegible",
    "IBM Plex Sans Condensed",
  ]) {
    assert(
      !defaultTokenBlock.includes(`"${inactiveFamily}"`),
      `${fontsCSSPath}: inactive family ${inactiveFamily} must not appear in active ui/grid/mono tokens`,
      failures,
    );
  }

  const cssURLs = cssURLSet(fontsCSS);
  for (const relativeFile of cssURLs) {
    assert(
      existsSync(repoPath(root, `apps/web/public/assets/fonts/${relativeFile}`)),
      `${fontsCSSPath}: url ./${relativeFile} does not resolve to a vendored font file`,
      failures,
    );
  }

  const manifestFontFiles = new Set(
    (manifest.families ?? []).flatMap((family) =>
      (family?.files ?? []).map((fileRecord) => fileRecord?.path),
    ),
  );
  for (const fontFile of collectFontFiles(root)) {
    assert(
      manifestFontFiles.has(fontFile),
      `${fontFile}: committed .woff2 must be listed in ${manifestPath}`,
      failures,
    );
  }
  for (const fontFile of manifestFontFiles) {
    assert(
      cssURLs.has(fontFile),
      `${manifestPath}: manifest font file ${fontFile} must be referenced by ${fontsCSSPath}`,
      failures,
    );
  }

  const stagedVariables = stagedCssVariables(manifest);
  for (const variableName of declaredFontVariables(fontsCSS)) {
    assert(
      cssVariableIsConsumed(fontsCSS, variableName) ||
        stagedVariables.has(variableName),
      `${fontsCSSPath}: unused ${variableName} must be consumed or documented as staged/inactive manifest metadata`,
      failures,
    );
  }

  for (const family of manifest.families ?? []) {
    const label = family?.family ?? "<missing family>";
    assert(typeof family.family === "string" && family.family.trim() !== "", `${manifestPath}: family name missing`, failures);
    assert(typeof family.directory === "string" && family.directory.trim() !== "", `${label}: directory missing`, failures);
    assert(family.license === "OFL-1.1", `${label}: license must be OFL-1.1`, failures);
    assert(notice.includes(family.family), `${noticePath}: missing family ${family.family}`, failures);
    assert(
      licenseFileForFamily(root, family.directory) !== null,
      `${family.directory}: missing LICENSE.txt or OFL.txt`,
      failures,
    );
    const sourceFile = repoPath(root, `apps/web/public/assets/fonts/${family.directory}/SOURCE.json`);
    assert(existsSync(sourceFile), `${family.directory}: missing SOURCE.json`, failures);
    if (existsSync(sourceFile)) {
      const source = readJSON(sourceFile);
      assert(source.family === family.family, `${family.directory}/SOURCE.json: family mismatch`, failures);
      assert(source.license === family.license, `${family.directory}/SOURCE.json: license mismatch`, failures);
      assert(typeof source.upstream === "string" && source.upstream.startsWith("https://"), `${family.directory}/SOURCE.json: upstream must be https URL`, failures);
    }
    assert(Array.isArray(family.files) && family.files.length > 0, `${label}: files must be non-empty`, failures);
    validateRoleMetadata({
      family,
      failures,
      fontsCSS,
      manifestPathLabel: manifestPath,
      notice,
      sourceText,
    });
    for (const fileRecord of family.files ?? []) {
      const file = repoPath(root, `apps/web/public/assets/fonts/${fileRecord.path}`);
      assert(existsSync(file), `${fileRecord.path}: missing font file`, failures);
      assert(notice.includes(fileRecord.path), `${noticePath}: missing file ${fileRecord.path}`, failures);
      if (existsSync(file)) {
        const stats = statSync(file);
        assert(stats.size === fileRecord.bytes, `${fileRecord.path}: byte size mismatch`, failures);
        assert(sha256File(file) === fileRecord.sha256, `${fileRecord.path}: sha256 mismatch`, failures);
      }
    }
  }

  for (const file of collectFiles(root, [
    "apps/web/index.html",
    "apps/web/src",
    "packages",
    "internal/modules/reporting",
  ])) {
    const text = readFileSync(file, "utf8");
    assert(
      !remoteFontPattern.test(text),
      `${repoRelative(root, file)}: must not reference a remote font CDN`,
      failures,
    );
  }

  return failures;
}

export function createFontBundleFixture(options = {}) {
  const root = path.join(os.tmpdir(), `cartulary-font-fixture-${process.pid}-${Date.now()}-${Math.random().toString(16).slice(2)}`);
  const fontRoot = path.join(root, "apps/web/public/assets/fonts");
  mkdirSync(path.join(fontRoot, "inter"), { recursive: true });
  mkdirSync(path.join(root, "apps/web/src"), { recursive: true });
  mkdirSync(path.join(root, "packages"), { recursive: true });
  mkdirSync(path.join(root, "internal/modules/reporting"), { recursive: true });
  const bytes = Buffer.from("fixture-font", "utf8");
  const digest = createHash("sha256").update(bytes).digest("hex");
  writeFileSync(path.join(fontRoot, "inter/InterVariable.woff2"), bytes);
  writeFileSync(path.join(fontRoot, "inter/OFL.txt"), "OFL fixture\n");
  writeFileSync(
    path.join(fontRoot, "inter/SOURCE.json"),
    JSON.stringify(
      {
        family: "Inter",
        upstream: "https://example.test/inter",
        version_or_ref: "fixture",
        license: "OFL-1.1",
        files: [{ path: "InterVariable.woff2", sha256: digest, bytes: bytes.length }],
        retrieved_at: "2026-01-01T00:00:00.000Z",
        retrieved_by: "fixture",
      },
      null,
      2,
    ),
  );
  writeFileSync(
    path.join(fontRoot, "FONT_MANIFEST.json"),
    JSON.stringify(
      {
        schema_id: "cartulary.font_manifest.v1",
        generated_at: "2026-01-01T00:00:00.000Z",
        base_path: "apps/web/public/assets/fonts",
        families: [
          {
            family: "Inter",
            directory: "inter",
            active_by_default: true,
            ...(options.missingActivationMetadata
              ? {}
              : {
                  role_ids: options.undocumentedStagedFamily
                    ? ["alternate-ui"]
                    : ["ui", "grid", "grid-cell"],
                  activation_status: options.undocumentedStagedFamily
                    ? "staged_inactive"
                    : "active_default",
                  activation_selectors: options.undocumentedStagedFamily
                    ? []
                    : [":root --font-ui", "body"],
                }),
            ...(options.undocumentedStagedFamily
              ? { active_by_default: false }
              : {}),
            license: "OFL-1.1",
            version_or_ref: "fixture",
            source: "https://example.test/inter",
            files: [
              {
                path: "inter/InterVariable.woff2",
                sha256: options.badHash ? "0".repeat(64) : digest,
                bytes: options.badBytes ? bytes.length + 1 : bytes.length,
              },
            ],
          },
        ],
      },
      null,
      2,
    ),
  );
  writeFileSync(
    path.join(fontRoot, "NOTICE.fonts.md"),
    options.missingNotice
      ? "# Notice\n\nInter\n"
      : "# Notice\n\nInter\n\ninter/InterVariable.woff2\n",
  );
  writeFileSync(
    path.join(fontRoot, "fonts.css"),
    `${
      options.missingCssFontReference
        ? ""
        : options.localSource
          ? '@font-face { font-family: "Inter"; src: local("Inter"), url("./inter/InterVariable.woff2") format("woff2"); } '
          : '@font-face { font-family: "Inter"; src: url("./inter/InterVariable.woff2") format("woff2"); } '
    }:root { --font-ui: "Inter"; --font-grid: "Inter"; --font-mono: "JetBrains Mono"; ${
      options.deadFontVariable ? '--font-unused: "Inter"; ' : ""
    }} body, .cartulary-shell { font-family: var(--font-ui); } .cartulary-grid { font-family: var(--font-grid); } code { font-family: var(--font-mono); }\n`,
  );
  writeFileSync(path.join(root, "apps/web/index.html"), "<div id=\"root\"></div>\n");
  writeFileSync(
    path.join(root, "apps/web/src/fixture.ts"),
    options.remoteFont ? "export const css = 'https://fonts.googleapis.com/css2?family=Inter';\n" : "export const ok = true;\n",
  );
  if (options.missingLicense) {
    rmSync(path.join(fontRoot, "inter/OFL.txt"));
  }
  return root;
}

export function runFontBundleCheckCLI(argv = process.argv.slice(2)) {
  try {
    const { root } = parseArgs(argv);
    const failures = checkFontBundle(path.resolve(root));
    if (failures.length > 0) {
      for (const failure of failures) {
        console.error(failure);
      }
      process.exit(1);
    }
    console.log("font bundle verified");
  } catch (error) {
    console.error(error instanceof Error ? error.message : String(error));
    process.exit(2);
  }
}

if (import.meta.url === pathToFileURL(process.argv[1]).href) {
  runFontBundleCheckCLI();
}
