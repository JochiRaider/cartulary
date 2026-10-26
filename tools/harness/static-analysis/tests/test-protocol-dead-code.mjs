import assert from "node:assert/strict";
import path from "node:path";
import test from "node:test";
import { spawnSync } from "../../workspace/child-process.mjs";
import { createFixture, repository } from "./fixtures.mjs";

test("protocol analyzer preserves live and generated exports while enforcing authored dead code", (t) => {
  const fixture = createFixture(t);
  fixture.write("package.json", {
    "private": true,
    "workspaces": [
      "apps/web",
      "packages/*"
    ]
  });
  fixture.write(".fallowrc.json", {
    "$schema": "https://raw.githubusercontent.com/fallow-rs/fallow/main/schema.json",
    "entry": [
      "apps/web/src/main.ts"
    ],
    "workspaces": {
      "patterns": [
        "apps/web",
        "packages/*"
      ]
    },
    "ignorePatterns": [
      "packages/protocol-ts/src/generated/**"
    ],
    "ignoreExports": [
      {
        "file": "packages/protocol-ts/src/index.ts",
        "exports": [
          "*"
        ]
      }
    ],
    "publicPackages": [
      "@cartulary/protocol-ts"
    ],
    "rules": {
      "unused-files": "warn",
      "unused-exports": "warn",
      "unused-types": "warn",
      "unused-dependencies": "off"
    },
    "overrides": [
      {
        "files": [
          "packages/protocol-ts/src/generated/**"
        ],
        "rules": {
          "unused-files": "off",
          "unused-exports": "off",
          "unused-types": "off"
        }
      }
    ]
  });
  fixture.write("apps/web/package.json", {
    "name": "@cartulary/web",
    "private": true,
    "dependencies": {
      "@cartulary/protocol-ts": "workspace:*"
    }
  });
  fixture.write("apps/web/src/main.ts", "import { liveValue } from \"@cartulary/protocol-ts\";\nimport type { LiveType } from \"@cartulary/protocol-ts\";\n\nexport const appValue: LiveType = liveValue;\n");
  fixture.write("packages/protocol-ts/package.json", {
    "name": "@cartulary/protocol-ts",
    "private": true,
    "type": "module",
    "exports": {
      ".": "./src/index.ts",
      "./family": "./src/entrypoints/family.ts"
    }
  });
  fixture.write("packages/protocol-ts/src/index.ts", "export { liveValue, type LiveType, unusedExport } from \"./live\";\n");
  fixture.write("packages/protocol-ts/src/live.ts", "export const liveValue = \"live\";\nexport type LiveType = string;\nexport const unusedExport = \"unused export\";\nexport const unusedValue = \"unused value\";\nexport type UnusedType = { readonly unused: true };\n");
  fixture.write("packages/protocol-ts/src/unused-file.ts", "const unreachable = \"unused file\";\nvoid unreachable;\n");
  fixture.write("packages/protocol-ts/src/compatibility.compile.ts", "export const compileOnlyValue = \"compile fixture\";\nexport type CompileOnlyType = string;\n");
  fixture.write("packages/protocol-ts/src/generated/family-types.ts", "export type OwnerSelectedType = { readonly id: string };\nexport type OwnerSelectedFutureType = { readonly generation: number };\n");
  fixture.write("packages/protocol-ts/src/entrypoints/family.ts", "export type * from \"../generated/family-types.js\";\n");
  fixture.write("contracts/protocol-ts/frontend-entrypoints.v2.json", {
    "entrypoints": [
      {
        "specifier": "@cartulary/protocol-ts/family",
        "authored_path": "packages/protocol-ts/src/entrypoints/family.ts",
        "generated_module_allowlist": [
          "packages/protocol-ts/src/generated/family-types.ts"
        ]
      }
    ]
  });
  const child = spawnSync(process.execPath, [
    path.join(repository, "tools/harness/static-analysis/protocol-ts-dead-code-check.mjs"),
    "--root", fixture.root, "--fallow-bin", path.join(repository, "node_modules/fallow/bin/fallow"),
    "--output", path.join(fixture.root, "result.json"),
  ], { encoding: "utf8", timeout: 30000 });
  assert.equal(child.error, undefined);
  assert.equal(child.status, 1, child.stdout + child.stderr);
  const report = fixture.readJSON("result.json");
  const text = JSON.stringify(report.findings);
  for (const expected of ["unused-file.ts", "unusedExport", "unusedValue", "UnusedType"]) {
    assert.ok(text.includes(expected), expected);
  }
  for (const live of ["liveValue", "LiveType", "compileOnlyValue", "CompileOnlyType", "OwnerSelectedType", "OwnerSelectedFutureType"]) {
    assert.equal(text.includes(live), false, live);
  }
  assert.equal(report.authored_suppressions, 0);
  assert.equal(report.automated_compile_fixture_exports, 1);
  assert.equal(report.automated_generated_pass_through_exports, 2);
});
