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
const defaultRepoRoot = path.resolve(scriptDir, "..");
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

  const cssURLMatches = [...fontsCSS.matchAll(/url\("([^"]+)"\)/giu)].map(
    (match) => decodeURIComponent(match[1]),
  );
  for (const cssURL of cssURLMatches) {
    const relativeFile = cssURL.replace(/^\.\//u, "");
    assert(
      existsSync(repoPath(root, `apps/web/public/assets/fonts/${relativeFile}`)),
      `${fontsCSSPath}: url ${cssURL} does not resolve to a vendored font file`,
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
    options.localSource
      ? '@font-face { font-family: "Inter"; src: local("Inter"), url("./inter/InterVariable.woff2") format("woff2"); } :root { --font-ui: "Inter"; --font-grid: "Inter"; --font-mono: "JetBrains Mono"; }\n'
      : '@font-face { font-family: "Inter"; src: url("./inter/InterVariable.woff2") format("woff2"); } :root { --font-ui: "Inter"; --font-grid: "Inter"; --font-mono: "JetBrains Mono"; }\n',
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

if (import.meta.url === pathToFileURL(process.argv[1]).href) {
  try {
    const { root } = parseArgs(process.argv.slice(2));
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
