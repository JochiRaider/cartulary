import { readFileSync } from "node:fs";
import path from "node:path";
import { createFixture, repository } from "./fixtures.mjs";

export function createRulepackFixture(t) {
  const fixture = createFixture(t);
  fixture.write("package.json", {
    "private": true,
    "workspaces": [
      "apps/web",
      "packages/*"
    ],
    "dependencies": {
      "react-data-grid": "7.0.0-beta.59"
    }
  });
  fixture.write("apps/web/package.json", {
    "name": "@cartulary/web",
    "private": true,
    "dependencies": {
      "@cartulary/grid-adapter": "workspace:*"
    }
  });
  fixture.write("packages/grid-adapter/package.json", {
    "name": "@cartulary/grid-adapter",
    "private": true,
    "dependencies": {
      "react-data-grid": "7.0.0-beta.59"
    }
  });
  fixture.write("apps/web/src/bad.ts", "import DataGrid from \"react-data-grid\";\n\nexport const bad = DataGrid;\n");
  fixture.write("packages/grid-adapter/src/index.tsx", "import DataGrid from \"react-data-grid\";\n\nexport const allowed = DataGrid;\n");
  fixture.write(".fallowrc.json", {
    "$schema": "https://raw.githubusercontent.com/fallow-rs/fallow/main/schema.json",
    "entry": [
      "apps/web/src/bad.ts",
      "packages/grid-adapter/src/index.tsx"
    ],
    "rules": {
      "unused-files": "off",
      "unused-exports": "off",
      "unused-types": "off",
      "unused-dependencies": "off",
      "unlisted-dependencies": "off",
      "unresolved-imports": "off",
      "policy-violation": "warn"
    },
    "rulePacks": [
      "tools/fallow/cartulary-boundaries.rulepack.json"
    ]
  });
  fixture.write("tools/fallow/cartulary-boundaries.rulepack.json", readFileSync(path.join(repository, "tools/fallow/cartulary-boundaries.rulepack.json"), "utf8"));
  return fixture;
}

export function createReachabilityFixture(t) {
  const fixture = createFixture(t);
  fixture.write("package.json", {
    "private": true,
    "workspaces": [
      "apps/web",
      "packages/*"
    ],
    "devDependencies": {
      "@cyclonedx/cdxgen": "12.3.1",
      "unused-tool": "1.0.0"
    }
  });
  fixture.write("apps/web/package.json", {
    "name": "@cartulary/web",
    "private": true,
    "devDependencies": {
      "vitest": "4.1.4",
      "vite": "8.0.8"
    }
  });
  fixture.write("packages/example/package.json", {
    "name": "@cartulary/example",
    "private": true
  });
  fixture.write(".fallowrc.json", {
    "$schema": "https://raw.githubusercontent.com/fallow-rs/fallow/main/schema.json",
    "entry": [
      "apps/web/src/main.tsx"
    ],
    "workspaces": {
      "patterns": [
        "apps/web",
        "packages/*"
      ]
    },
    "rules": {
      "unused-files": "warn",
      "unused-exports": "warn",
      "unused-types": "warn",
      "unused-dependencies": "off",
      "unused-dev-dependencies": "warn",
      "unlisted-dependencies": "off",
      "unresolved-imports": "warn",
      "policy-violation": "off"
    }
  });
  fixture.write("apps/web/vite.config.ts", "export default {};\n");
  fixture.write("apps/web/index.html", "<html>\n  <head>\n    <link rel=\"preload\" href=\"/assets/fonts/inter/InterVariable.woff2\" as=\"font\" />\n    <link rel=\"stylesheet\" href=\"/assets/fonts/fonts.css\" />\n  </head>\n  <body>\n    <script type=\"module\" src=\"/src/main.tsx\"></script>\n  </body>\n</html>\n");
  fixture.write("apps/web/measurement.html", "<html><body><script src=\"./src/measurement/main.ts\" type=\"module\"></script></body></html>\n");
  fixture.write("apps/web/src/measurement/main.ts", "import { fixture } from \"./fixture\";\nconsole.log(fixture);\n");
  fixture.write("apps/web/src/measurement/fixture.ts", "export const fixture = \"measurement\";\n");
  fixture.write("apps/web/src/measurement/unreachable.ts", "export const unusedMeasurement = true;\n");
  fixture.write("apps/web/public/assets/fonts/fonts.css", "@font-face {\n  font-family: Inter;\n  src: url(\"./inter/InterVariable.woff2\") format(\"woff2\");\n}\n");
  fixture.write("apps/web/src/main.tsx", "import { connect, usedValue, type Reader } from \"./surface\";\nconst reader: Reader = { read: () => usedValue };\nconnect(reader, (value) => console.log(value));\nexport const app = \"cartulary\";\n");
  fixture.write("apps/web/src/surface.ts", "export interface Reader { read(): string }\nexport function connect(reader: Reader, callback: (value: string) => void) {\n  callback(reader.read());\n}\nexport const usedValue = \"used\";\nexport const unusedValue = \"dead\";\nexport type UnusedType = { dead: true };\n");
  fixture.write("apps/web/src/fixtureSupport.ts", "export const testOnly = \"test\";\n");
  fixture.write("apps/web/src/surface.test.ts", "import { testOnly } from \"./fixtureSupport\";\nif (testOnly !== \"test\") throw new Error(\"fixture failed\");\n");
  fixture.write("apps/web/src/testing/testSetup.ts", "export const setup = \"vitest\";\n");
  fixture.write("apps/web/src/testing/testSetup.dom.ts", "export const setupDom = \"vitest-dom\";\n");
  fixture.write("tools/harness/static-analysis/example-cli.mjs", "export const exampleCli = true;\n");
  fixture.write("tools/harness/static-analysis/example-test.mjs", "export const exampleTest = true;\n");
  fixture.write("tools/harness/test-support/example-dynamic.mjs", "export const importedDynamicPeer = true;\nexport const modeledDynamicExport = true;\nexport const ordinaryUnusedExport = true;\n");
  fixture.write("tools/harness/test-support/example-direct.mjs", "import { importedDynamicPeer } from \"./example-dynamic.mjs\";\n\nexport const exampleDirect = importedDynamicPeer;\n");
  fixture.write("tools/release-evidence/generate-sbom-license-evidence.mjs", "import { spawnSync } from \"node:child_process\";\n\nexport function runCdxgen(pnpm = \"pnpm\") {\n  return spawnSync(pnpm, [\"exec\", \"cdxgen\", \"--version\"]);\n}\n");
  fixture.write("packages/example/tsconfig.json", {
    "include": [
      "src"
    ]
  });
  fixture.write("packages/example/src/contract.compile.ts", "// Compile-only consumer: no runtime entry or test title.\nexport type ContractFixture = { required: string };\n");
  fixture.write("packages/example/src/orphan.ts", "export const unusedProjectSource = true;\n");
  fixture.write("tools/harness/static-analysis/shell-child.mjs", "console.log(\"real shell-launched helper\");\n");
  fixture.write("tools/harness/static-analysis/orphan.mjs", "export const unusedTool = true;\n");
  fixture.write("tools/task_surface_owner.json", {
    "schema_id": "cartulary.task_surface_owner.v1",
    "targets": [
      {
        "name": "example-tool",
        "backing_scripts": [
          "tools/harness/static-analysis/example-cli.mjs",
          "tools/harness/static-analysis/shell-child.mjs"
        ]
      }
    ],
    "harness_checks": [
      {
        "name": "harness-smoke-example",
        "backing_scripts": [
          "tools/harness/static-analysis/example-test.mjs"
        ],
        "command": [
          "node",
          "./tools/harness/static-analysis/example-test.mjs"
        ]
      }
    ],
    "make_recipes": {
      "example-tool": {
        "type": "node_tool"
      }
    }
  });
  fixture.write("tools/fallow/reachability_owner.json", {
    "schema_id": "cartulary.fallow_reachability_owner.v1",
    "base_config": {
      "path": ".fallowrc.json"
    },
    "blocking_web_surface": {
      "root": "apps/web",
      "rules": [
        "unused_files",
        "unused_exports",
        "unused_types"
      ]
    },
    "blocking_package_surfaces": {
      "packages": [
        {
          "package_name": "@cartulary/example",
          "entrypoint": "apps/web/src/main.tsx",
          "own_test_globs": [
            "packages/example/src/**/*.test.ts"
          ]
        }
      ],
      "rules": [
        "unused_exports",
        "unused_types"
      ]
    },
    "task_surface": {
      "owner_path": "tools/task_surface_owner.json",
      "script_extensions": [
        ".cjs",
        ".cts",
        ".js",
        ".jsx",
        ".mjs",
        ".mts",
        ".ts",
        ".tsx"
      ],
      "required_node_tool_backing_scripts": true
    },
    "harness_entrypoints": {
      "files": [
        "tools/harness/test-support/example-direct.mjs"
      ]
    },
    "typescript": {
      "project_file": "packages/example/tsconfig.json"
    },
    "harness_dynamic_exports": [
      {
        "file": "tools/harness/test-support/example-dynamic.mjs",
        "exports": [
          "modeledDynamicExport"
        ],
        "owner": "fixture shell dynamic import",
        "evidence": "fixture coverage for reachability-owner dynamic export handling",
        "removal_trigger": "remove when fixture no longer needs dynamic export modeling"
      }
    ],
    "vitest": {
      "config_file": "apps/web/vite.config.ts",
      "setup_files": [
        "apps/web/src/testing/testSetup.ts",
        "apps/web/src/testing/testSetup.dom.ts"
      ],
      "test_entry_globs": [
        "apps/web/src/**/*.test.ts",
        "apps/web/src/**/*.test.tsx"
      ]
    },
    "vite_public_assets": {
      "public_root": "apps/web/public",
      "html_entry_files": [
        "apps/web/index.html",
        "apps/web/measurement.html"
      ],
      "url_prefixes": [
        "/assets/"
      ],
      "always_used_files": [
        "apps/web/public/assets/fonts/fonts.css"
      ]
    },
    "executable_tooling_dependencies": [
      {
        "package_name": "@cyclonedx/cdxgen",
        "owner_script": "tools/release-evidence/generate-sbom-license-evidence.mjs",
        "command": [
          "pnpm",
          "exec",
          "cdxgen"
        ]
      }
    ],
    "static_policy": {
      "runtime_enabled": false,
      "per_file_suppression_growth": false
    },
    "extensions": {}
  });
  fixture.write("apps/web/public/assets/fonts/inter/InterVariable.woff2", "font\n");
  return fixture;
}

export function createE2EReachabilityFixture(t) {
  const fixture = createReachabilityFixture(t);
  const authoredConfig = JSON.parse(readFileSync(path.join(repository, ".fallowrc.json"), "utf8"));
  const authoredOwner = JSON.parse(readFileSync(path.join(repository, "tools/fallow/reachability_owner.json"), "utf8"));
  const config = fixture.readJSON(".fallowrc.json");
  config.entry.push(...authoredConfig.entry.filter((entry) => entry.startsWith("apps/web/e2e/")));
  fixture.write(".fallowrc.json", config);
  const owner = fixture.readJSON("tools/fallow/reachability_owner.json");
  owner.vitest.test_entry_globs.push(...authoredOwner.vitest.test_entry_globs.filter((entry) => entry.startsWith("apps/web/e2e/")));
  fixture.write("tools/fallow/reachability_owner.json", owner);
  fixture.write("apps/web/e2e/workbook.spec.ts", [
    'import { live, type Subject } from "./support/client";',
    'import * as namespace from "./support/namespace";',
    'import { forwarded } from "./support/facade";',
    'const subject: Subject = { id: live };',
    'console.log(subject, namespace.value, forwarded);',
    'void import("./support/dynamic").then(({ dynamicValue }) => console.log(dynamicValue));',
  ].join("\n"));
  fixture.write("apps/web/e2e/workbook.spec.tsx", 'import { jsxValue } from "./support/jsx"; console.log(jsxValue);');
  fixture.write("apps/web/e2e/global-setup.ts", 'import { setup } from "./support/setup"; export default setup;');
  fixture.write("apps/web/e2e/global-teardown.ts", 'import { teardown } from "./support/teardown"; export default teardown;');
  fixture.write("apps/web/e2e/support/client.test.ts", 'import { testOnly } from "./testOnly"; console.log(testOnly);');
  fixture.write("apps/web/e2e/support/client.ts", 'export const live = "record"; export type Subject = { id: string }; export const unusedValue = 1; export type UnusedType = string;');
  fixture.write("apps/web/e2e/support/namespace.ts", 'export const value = 1;');
  fixture.write("apps/web/e2e/support/facade.ts", 'export { forwarded } from "./forwarded";');
  fixture.write("apps/web/e2e/support/forwarded.ts", 'export const forwarded = 1;');
  fixture.write("apps/web/e2e/support/dynamic.ts", 'export const dynamicValue = 1;');
  fixture.write("apps/web/e2e/support/jsx.ts", 'export const jsxValue = 1;');
  fixture.write("apps/web/e2e/support/setup.ts", 'export function setup() {}');
  fixture.write("apps/web/e2e/support/teardown.ts", 'export function teardown() {}');
  fixture.write("apps/web/e2e/support/testOnly.ts", 'export const testOnly = 1;');
  fixture.write("apps/web/e2e/support/orphan.ts", 'export const orphan = 1;');
  return fixture;
}
