#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"
CHECKER="$ROOT_DIR/tools/harness/static-analysis/frontend-import-boundary-check-cli.mjs"
cleanup_paths=()

cleanup() {
  local path
  for path in "${cleanup_paths[@]}"; do
    rm -rf "$path"
  done
}

trap cleanup EXIT

fail() {
  echo "$*" >&2
  exit 1
}

assert_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "$haystack" != *"$needle"* ]]; then
    fail "$label: expected output to contain [$needle], got [$haystack]"
  fi
}

assert_not_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "$haystack" == *"$needle"* ]]; then
    fail "$label: expected output not to contain [$needle], got [$haystack]"
  fi
}

assert_passes() {
  local label="$1"
  shift

  local output
  if ! output="$("$@" 2>&1)"; then
    fail "$label: expected success, got output: $output"
  fi
  printf '%s' "$output"
}

assert_fails() {
  local label="$1"
  shift

  local output
  local status
  set +e
  output="$("$@" 2>&1)"
  status=$?
  set -e

  if [[ "$status" -eq 0 ]]; then
    fail "$label: expected failure"
  fi
  printf '%s' "$output"
}

write_config() {
  local case_root="$1"

  mkdir -p "$case_root/tools"
  cat >"$case_root/tools/frontend_import_boundaries.json" <<'JSON'
{
  "schema_id": "cartulary.frontend_import_boundaries.v2",
  "scan_roots": [
    "apps/web/src",
    "apps/web/e2e",
    "packages/grid-adapter/src",
    "packages/protocol-ts/src",
    "packages/test-utils/src",
    "packages/ui-contracts/src",
    "packages/view-contracts/src"
  ],
  "scan_excludes": [
    "packages/protocol-ts/src/generated/**",
    "packages/ui-contracts/src/generated/**"
  ],
  "acyclic_import_graphs": [
    {
      "id": "frontend-test-utils-production-acyclic-import-graph",
      "level": "error",
      "message": "Production test-utils imports, type imports, re-exports, and dynamic imports must form an enumerable acyclic graph.",
      "applies_to": {
        "include": ["packages/test-utils/src/**"],
        "exclude": [
          "packages/test-utils/src/*.test.ts",
          "packages/test-utils/src/unsupported-entrypoints.compile.ts",
          "packages/test-utils/src/test-suites/**"
        ]
      }
    }
  ],
  "singleton_imports": [
    {
      "id": "frontend-rdg-stylesheet-singleton",
      "level": "error",
      "message": "Import the react-data-grid stylesheet exactly once from the grid adapter.",
      "specifier": "react-data-grid/lib/styles.css",
      "required_count": 1,
      "allowed_importers": ["packages/grid-adapter/src/**"]
    }
  ],
  "rules": [
    {
      "id": "frontend-grid-vendor-boundary",
      "level": "error",
      "message": "Import react-data-grid only through @cartulary/grid-adapter.",
      "applies_to": {
        "include": ["**"],
        "exclude": []
      },
      "allowed_importers": [
        "packages/grid-adapter/src/**"
      ],
      "restricted_imports": [
        {
          "kind": "package",
          "name": "react-data-grid",
          "include_subpaths": true
        }
      ]
    },
    {
      "id": "frontend-grid-dom-unit-binding-boundary",
      "level": "error",
      "message": "Import the non-virtualized grid binding only from package-local DOM-unit tests.",
      "applies_to": {
        "include": ["**"],
        "exclude": [
          "packages/grid-adapter/src/*.test.ts",
          "packages/grid-adapter/src/*.test.tsx",
          "packages/grid-adapter/src/**/*.test.ts",
          "packages/grid-adapter/src/**/*.test.tsx"
        ]
      },
      "allowed_importers": [],
      "restricted_imports": [
        {
          "kind": "path_prefix",
          "path": "packages/grid-adapter/src/domUnitBinding"
        }
      ]
    },
    {
      "id": "frontend-assessment-owner-no-timeline-imports",
      "level": "error",
      "message": "Assessment models and presentation must not depend on Timeline implementation; only the candidate query adapter may translate Timeline rows.",
      "applies_to": {
        "include": [
          "apps/web/src/workbook/components/AssessmentWorkbookSurface.tsx",
          "apps/web/src/workbook/hooks/useAssessmentSupportCandidates.ts"
        ],
        "exclude": []
      },
      "allowed_importers": [
        "apps/web/src/workbook/hooks/useAssessmentSupportCandidates.ts"
      ],
      "restricted_imports": [
        {
          "kind": "path_prefix",
          "path": "apps/web/src/workbook/timeline"
        }
      ]
    },
    {
      "id": "frontend-generated-protocol-boundary",
      "level": "error",
      "message": "Import generated protocol artifacts only through owner-declared Protocol-TS entrypoints.",
      "applies_to": {
        "include": ["**"],
        "exclude": []
      },
      "allowed_importers": [
        "packages/protocol-ts/src/entrypoints/collaboration.ts",
        "packages/protocol-ts/src/entrypoints/errors.ts",
        "packages/protocol-ts/src/entrypoints/extensions.ts",
        "packages/protocol-ts/src/entrypoints/http.ts",
        "packages/protocol-ts/src/entrypoints/import-targets.ts",
        "packages/protocol-ts/src/entrypoints/network-flow.ts",
        "packages/protocol-ts/src/entrypoints/view-schemas.ts"
      ],
      "restricted_imports": [
        {
          "kind": "package",
          "name": "@cartulary/protocol-ts/generated",
          "include_subpaths": true
        },
        {
          "kind": "path_prefix",
          "path": "packages/protocol-ts/src/generated"
        }
      ]
    },
    {
      "id": "frontend-generated-protocol-barrel-deny",
      "level": "error",
      "message": "Authored frontend code must not import the broad generated protocol barrel.",
      "applies_to": { "include": ["**"], "exclude": [] },
      "allowed_importers": [],
      "restricted_imports": [
        {
          "kind": "path_prefix",
          "path": "packages/protocol-ts/src/generated/index"
        }
      ]
    },
    {
      "id": "frontend-generated-protocol-audit-deny",
      "level": "error",
      "message": "Browser-authored code must not import generated audit artifacts.",
      "applies_to": { "include": ["**"], "exclude": [] },
      "allowed_importers": [],
      "restricted_imports": [
        {
          "kind": "path_prefix",
          "path": "packages/protocol-ts/src/generated/audit-artifacts"
        }
      ]
    },
    {
      "id": "frontend-generated-protocol-revisions-deny",
      "level": "error",
      "message": "Browser-authored code must not import generated revisions artifacts.",
      "applies_to": { "include": ["**"], "exclude": [] },
      "allowed_importers": [],
      "restricted_imports": [
        {
          "kind": "path_prefix",
          "path": "packages/protocol-ts/src/generated/revisions-artifacts"
        }
      ]
    },
    {
      "id": "frontend-generated-protocol-obsolete-artifact-projection-deny",
      "level": "error",
      "message": "Authored frontend code must not import removed aggregate Protocol-TS artifact projections.",
      "applies_to": { "include": ["**"], "exclude": [] },
      "allowed_importers": [],
      "restricted_imports": [
        {
          "kind": "path_prefix",
          "path": "packages/protocol-ts/src/generated/errors-artifacts"
        },
        {
          "kind": "path_prefix",
          "path": "packages/protocol-ts/src/generated/extensions-artifacts"
        },
        {
          "kind": "path_prefix",
          "path": "packages/protocol-ts/src/generated/network-flow-artifacts"
        },
        {
          "kind": "path_prefix",
          "path": "packages/protocol-ts/src/generated/ws-artifacts"
        }
      ]
    },
    {
      "id": "frontend-generated-protocol-validator-barrel-deny",
      "level": "error",
      "message": "Authored frontend code must import only its owner-specific generated validator module.",
      "applies_to": { "include": ["**"], "exclude": [] },
      "allowed_importers": [],
      "restricted_imports": [
        {
          "kind": "path_prefix",
          "path": "packages/protocol-ts/src/generated/protocol-validators"
        }
      ]
    },
    {
      "id": "frontend-generated-protocol-http-owner",
      "level": "error",
      "message": "Only the HTTP entrypoint may import generated Core HTTP validators and operation bindings.",
      "applies_to": { "include": ["**"], "exclude": [] },
      "allowed_importers": [
        "packages/protocol-ts/src/entrypoints/http.ts"
      ],
      "restricted_imports": [
        {
          "kind": "path_prefix",
          "path": "packages/protocol-ts/src/generated/http-operation-bindings"
        },
        {
          "kind": "path_prefix",
          "path": "packages/protocol-ts/src/generated/core-http-validators"
        }
      ]
    },
    {
      "id": "frontend-generated-protocol-collaboration-owner",
      "level": "error",
      "message": "Only the collaboration entrypoint may import generated collaboration types.",
      "applies_to": { "include": ["**"], "exclude": [] },
      "allowed_importers": [
        "packages/protocol-ts/src/entrypoints/collaboration.ts"
      ],
      "restricted_imports": [
        {
          "kind": "path_prefix",
          "path": "packages/protocol-ts/src/generated/collaboration-types"
        }
      ]
    },
    {
      "id": "frontend-generated-protocol-core-http-owner",
      "level": "error",
      "message": "Only the HTTP entrypoint may import generated Core HTTP types.",
      "applies_to": { "include": ["**"], "exclude": [] },
      "allowed_importers": [
        "packages/protocol-ts/src/entrypoints/http.ts"
      ],
      "restricted_imports": [
        {
          "kind": "path_prefix",
          "path": "packages/protocol-ts/src/generated/core-http-types"
        }
      ]
    },
    {
      "id": "frontend-generated-protocol-contract-artifact-owner",
      "level": "error",
      "message": "Only owner-declared family entrypoints may import approved generated artifacts and registries.",
      "applies_to": { "include": ["**"], "exclude": [] },
      "allowed_importers": [
        "packages/protocol-ts/src/entrypoints/errors.ts",
        "packages/protocol-ts/src/entrypoints/extensions.ts",
        "packages/protocol-ts/src/entrypoints/import-targets.ts",
        "packages/protocol-ts/src/entrypoints/network-flow.ts",
        "packages/protocol-ts/src/entrypoints/view-schemas.ts"
      ],
      "restricted_imports": [
        {
          "kind": "path_prefix",
          "path": "packages/protocol-ts/src/generated/import-target-registry"
        },
        {
          "kind": "path_prefix",
          "path": "packages/protocol-ts/src/generated/error-registry"
        },
        {
          "kind": "path_prefix",
          "path": "packages/protocol-ts/src/generated/extension-client-support-registry"
        },
        {
          "kind": "path_prefix",
          "path": "packages/protocol-ts/src/generated/extension-profile-registry"
        },
        {
          "kind": "path_prefix",
          "path": "packages/protocol-ts/src/generated/network-flow-error-registry"
        },
        {
          "kind": "path_prefix",
          "path": "packages/protocol-ts/src/generated/view-schema-registry"
        }
      ]
    },
    {
      "id": "frontend-generated-protocol-network-flow-owner",
      "level": "error",
      "message": "Only the Network Flow entrypoint may import generated Network Flow behavior projections.",
      "applies_to": { "include": ["**"], "exclude": [] },
      "allowed_importers": [
        "packages/protocol-ts/src/entrypoints/network-flow.ts"
      ],
      "restricted_imports": [
        {
          "kind": "path_prefix",
          "path": "packages/protocol-ts/src/generated/network-flow-types"
        },
        {
          "kind": "path_prefix",
          "path": "packages/protocol-ts/src/generated/network-flow-validators"
        }
      ]
    },
    {
      "id": "frontend-generated-design-token-boundary",
      "level": "error",
      "message": "Import generated design token artifacts only through the @cartulary/ui-contracts facade.",
      "applies_to": {
        "include": ["**"],
        "exclude": []
      },
      "allowed_importers": [
        "packages/ui-contracts/src/index.ts"
      ],
      "restricted_imports": [
        {
          "kind": "path_prefix",
          "path": "packages/ui-contracts/src/generated"
        }
      ]
    },
    {
      "id": "frontend-runtime-node-boundary",
      "level": "error",
      "message": "Browser runtime code must not import Node builtins.",
      "applies_to": {
        "include": ["apps/web/src/**", "packages/grid-adapter/src/**", "packages/protocol-ts/src/**", "packages/ui-contracts/src/**", "packages/view-contracts/src/**"],
        "exclude": ["apps/web/src/*.test.ts", "apps/web/src/*.test.tsx", "apps/web/src/**/*.test.ts", "apps/web/src/**/*.test.tsx", "apps/web/src/**/*TestSupport.ts", "packages/grid-adapter/src/*.test.tsx", "packages/grid-adapter/src/**/*.test.tsx", "packages/grid-adapter/src/test-support.tsx", "packages/ui-contracts/src/*.test.ts", "packages/ui-contracts/src/**/*.test.ts", "packages/view-contracts/src/*.test.ts", "packages/view-contracts/src/**/*.test.ts"]
      },
      "allowed_importers": [],
      "restricted_imports": [
        {
          "kind": "node_builtin",
          "names": ["*"]
        }
      ]
    },
    {
      "id": "frontend-test-utils-runtime-dependency-boundary",
      "level": "error",
      "message": "Production test-utils code must remain Playwright- and Node-independent.",
      "applies_to": {
        "include": ["packages/test-utils/src/**"],
        "exclude": [
          "packages/test-utils/src/*.test.ts",
          "packages/test-utils/src/unsupported-entrypoints.compile.ts",
          "packages/test-utils/src/test-suites/**"
        ]
      },
      "allowed_importers": [],
      "restricted_imports": [
        {
          "kind": "package",
          "name": "@playwright/test",
          "include_subpaths": true
        },
        {
          "kind": "node_builtin",
          "names": ["*"]
        }
      ]
    },
    {
      "id": "frontend-test-utils-observer-direction",
      "level": "error",
      "message": "Test-utils observers must not depend on setup or action modules.",
      "applies_to": {
        "include": [
          "packages/test-utils/src/focus.ts",
          "packages/test-utils/src/grid-observers.ts",
          "packages/test-utils/src/marker.ts"
        ],
        "exclude": []
      },
      "allowed_importers": [],
      "restricted_imports": [
        {
          "kind": "path_prefix",
          "path": "packages/test-utils/src/grid-actions"
        },
        {
          "kind": "path_prefix",
          "path": "packages/test-utils/src/grid-setup"
        }
      ]
    },
    {
      "id": "frontend-workspace-package-facade-boundary",
      "level": "error",
      "message": "Import workspace packages only through package names and declared package.json exports.",
      "applies_to": {
        "include": ["**"],
        "exclude": [
          "packages/test-utils/src/unsupported-entrypoints.compile.ts"
        ]
      },
      "allowed_importers": [],
      "restricted_imports": [
        {
          "kind": "workspace_package_facade",
          "package_roots": [
            "packages/grid-adapter",
            "packages/protocol-ts",
            "packages/test-utils",
            "packages/ui-contracts",
            "packages/view-contracts"
          ]
        }
      ]
    },
    {
      "id": "frontend-runtime-test-helper-boundary",
      "level": "error",
      "message": "Browser runtime code must not import test, e2e, or harness helper surfaces.",
      "applies_to": {
        "include": ["apps/web/src/**", "packages/grid-adapter/src/**", "packages/protocol-ts/src/**", "packages/ui-contracts/src/**", "packages/view-contracts/src/**"],
        "exclude": ["apps/web/src/*.test.ts", "apps/web/src/*.test.tsx", "apps/web/src/**/*.test.ts", "apps/web/src/**/*.test.tsx", "apps/web/src/**/*TestSupport.ts", "apps/web/src/testing/fetchMockTestSupport.ts", "apps/web/src/testing/timelineWorkbookTestSupport.ts", "packages/grid-adapter/src/*.test.tsx", "packages/grid-adapter/src/**/*.test.tsx", "packages/grid-adapter/src/test-support.tsx", "packages/ui-contracts/src/*.test.ts", "packages/ui-contracts/src/**/*.test.ts", "packages/view-contracts/src/*.test.ts", "packages/view-contracts/src/**/*.test.ts"]
      },
      "allowed_importers": [],
      "restricted_imports": [
        {
          "kind": "package",
          "name": "@cartulary/test-utils",
          "include_subpaths": true
        },
        {
          "kind": "package",
          "name": "@cartulary/grid-adapter/test-support",
          "include_subpaths": true
        },
        {
          "kind": "package",
          "name": "@playwright/test",
          "include_subpaths": true
        },
        {
          "kind": "package",
          "name": "@testing-library/react",
          "include_subpaths": true
        },
        {
          "kind": "package",
          "name": "vitest",
          "include_subpaths": true
        },
        {
          "kind": "path_prefix",
          "path": "apps/web/src/testing/appShellTestSupport"
        },
        {
          "kind": "path_prefix",
          "path": "apps/web/src/testing/fetchMockTestSupport"
        },
        {
          "kind": "path_prefix",
          "path": "apps/web/src/testing/timelineWorkbookTestSupport"
        }
      ]
    },
    {
      "id": "web-e2e-workbook-contract-boundary",
      "level": "error",
      "message": "E2E workbook support must consume generated-projection metadata from @cartulary/view-contracts instead of the web registry.",
      "applies_to": {
        "include": ["apps/web/e2e/**"],
        "exclude": []
      },
      "allowed_importers": [],
      "restricted_imports": [
        {
          "kind": "path_prefix",
          "path": "apps/web/src/workbook/models/workbookSurfaceRegistry"
        }
      ]
    },
    {
      "id": "web-e2e-test-utils-subpath-boundary",
      "level": "error",
      "message": "Import @cartulary/test-utils only through its semantic grid subpath.",
      "applies_to": {
        "include": ["apps/web/e2e/**"],
        "exclude": []
      },
      "allowed_importers": [],
      "restricted_imports": [
        {
          "kind": "package",
          "name": "@cartulary/test-utils",
          "include_subpaths": false
        }
      ]
    },
    {
      "id": "web-e2e-public-control-transport-boundary",
      "level": "error",
      "message": "Public JSON transport must not import privileged test-control transport.",
      "applies_to": {
        "include": ["apps/web/e2e/support/transport/publicJsonClient.ts"],
        "exclude": []
      },
      "allowed_importers": [],
      "restricted_imports": [
        {
          "kind": "path_prefix",
          "path": "apps/web/e2e/support/transport/testControlClient"
        }
      ]
    },
    {
      "id": "web-e2e-semantic-fixture-direction",
      "level": "error",
      "message": "Semantic fixtures must not depend on accessibility or visual evidence composition.",
      "applies_to": {
        "include": ["apps/web/e2e/support/evidence/**"],
        "exclude": []
      },
      "allowed_importers": [],
      "restricted_imports": [
        {
          "kind": "path_prefix",
          "path": "apps/web/e2e/support/visual"
        }
      ]
    },
    {
      "id": "web-e2e-core-support-direction",
      "level": "error",
      "message": "Core E2E support must not depend on higher semantic fixture or evidence owners.",
      "applies_to": {
        "include": ["apps/web/e2e/support/workbook/**"],
        "exclude": []
      },
      "allowed_importers": [],
      "restricted_imports": [
        {
          "kind": "path_prefix",
          "path": "apps/web/e2e/support/evidence"
        }
      ]
    },
    {
      "id": "frontend-synthetic-warning-boundary",
      "level": "warning",
      "message": "Synthetic warning-only import boundary fixture.",
      "applies_to": {
        "include": ["**"],
        "exclude": []
      },
      "allowed_importers": [],
      "restricted_imports": [
        {
          "kind": "package",
          "name": "cartulary-warning-only-fixture",
          "include_subpaths": true
        }
      ]
    }
  ],
  "raw_design_token_literal_checks": [
    {
      "id": "frontend-runtime-raw-design-color-literals",
      "level": "error",
      "message": "Runtime UI code must use generated Cartulary design token artifacts instead of raw machine-owned color token literals.",
      "design_document": "contracts/design/tokens.v1.json",
      "token_namespaces": ["colors"],
      "applies_to": {
        "include": ["apps/web/src/**", "packages/grid-adapter/src/**"],
        "exclude": ["apps/web/src/*.test.ts", "apps/web/src/*.test.tsx", "packages/grid-adapter/src/*.test.tsx"]
      }
    }
  ]
}
JSON
}

prepare_case_root() {
  local name="$1"
  local case_root="$tmp_dir/$name"

  mkdir -p \
    "$case_root/apps/web/src" \
    "$case_root/apps/web/e2e" \
    "$case_root/apps/web/e2e/support/evidence" \
    "$case_root/apps/web/e2e/support/transport" \
    "$case_root/apps/web/e2e/support/visual" \
    "$case_root/apps/web/e2e/support/workbook" \
    "$case_root/packages/grid-adapter/src" \
    "$case_root/packages/protocol-ts/src/entrypoints" \
    "$case_root/packages/protocol-ts/src/generated" \
    "$case_root/packages/test-utils/src" \
    "$case_root/packages/ui-contracts/src/generated" \
    "$case_root/packages/ui-contracts/src" \
    "$case_root/packages/view-contracts/src" \
    "$case_root/contracts/design"
  cp "$ROOT_DIR/contracts/design/tokens.v1.json" "$case_root/contracts/design/tokens.v1.json"
  cat >"$case_root/packages/grid-adapter/package.json" <<'JSON'
{
  "name": "@cartulary/grid-adapter",
  "exports": {
    ".": "./src/index.tsx",
    "./test-support": "./src/test-support.tsx"
  }
}
JSON
  cat >"$case_root/packages/protocol-ts/package.json" <<'JSON'
{
  "name": "@cartulary/protocol-ts",
  "exports": {
    ".": "./src/index.ts"
  }
}
JSON
  cat >"$case_root/packages/test-utils/package.json" <<'JSON'
{
  "name": "@cartulary/test-utils",
  "exports": {
    "./grid": "./src/grid.ts"
  }
}
JSON
  cat >"$case_root/packages/ui-contracts/package.json" <<'JSON'
{
  "name": "@cartulary/ui-contracts",
  "exports": {
    ".": "./src/index.ts"
  }
}
JSON
  cat >"$case_root/packages/view-contracts/package.json" <<'JSON'
{
  "name": "@cartulary/view-contracts",
  "exports": {
    ".": "./src/index.ts"
  }
}
JSON
  cat >"$case_root/packages/grid-adapter/src/index.tsx" <<'TS'
import "react-data-grid/lib/styles.css";

export const gridStylesheetLoaded = true;
TS
  write_config "$case_root"
  printf '%s\n' "$case_root"
}

run_checker() {
  local case_root="$1"
  shift

  "$NODE_BIN" "$CHECKER" --root "$case_root" --config tools/frontend_import_boundaries.json "$@"
}

mkdir -p "$ROOT_DIR/tmp"
tmp_dir="$(mktemp -d "$ROOT_DIR/tmp/frontend-import-boundaries.XXXXXX")"
cleanup_paths+=("$tmp_dir")

allowed_grid_root="$(prepare_case_root allowed-grid)"
cat >"$allowed_grid_root/packages/grid-adapter/src/index.tsx" <<'TS'
import { DataGrid } from "react-data-grid";
import "react-data-grid/lib/styles.css";

export const grid = DataGrid;
TS
allowed_grid_output="$(assert_passes "allowed grid adapter import" run_checker "$allowed_grid_root")"
assert_contains "$allowed_grid_output" "frontend import boundaries verified" "allowed grid adapter output"

blocked_grid_root="$(prepare_case_root blocked-grid)"
cat >"$blocked_grid_root/apps/web/src/GridLeak.tsx" <<'TS'
import { DataGrid } from "react-data-grid";

export const grid = DataGrid;
TS
blocked_grid_output="$(assert_fails "blocked app grid import" run_checker "$blocked_grid_root")"
assert_contains "$blocked_grid_output" "frontend-grid-vendor-boundary" "blocked grid rule"
assert_contains "$blocked_grid_output" "apps/web/src/GridLeak.tsx" "blocked grid file"

blocked_assessment_timeline_root="$(prepare_case_root blocked-assessment-timeline)"
mkdir -p \
  "$blocked_assessment_timeline_root/apps/web/src/workbook/components" \
  "$blocked_assessment_timeline_root/apps/web/src/workbook/timeline/models"
cat >"$blocked_assessment_timeline_root/apps/web/src/workbook/components/AssessmentWorkbookSurface.tsx" <<'TS'
import { normalizeTimelineFullRow } from "../timeline/models/workbookTimelineModel";

export const leaked = normalizeTimelineFullRow;
TS
blocked_assessment_timeline_output="$(assert_fails "blocked assessment Timeline import" run_checker "$blocked_assessment_timeline_root")"
assert_contains "$blocked_assessment_timeline_output" "frontend-assessment-owner-no-timeline-imports" "blocked assessment Timeline rule"
assert_contains "$blocked_assessment_timeline_output" "apps/web/src/workbook/components/AssessmentWorkbookSurface.tsx" "blocked assessment Timeline file"

allowed_assessment_candidate_root="$(prepare_case_root allowed-assessment-candidate)"
mkdir -p \
  "$allowed_assessment_candidate_root/apps/web/src/workbook/hooks" \
  "$allowed_assessment_candidate_root/apps/web/src/workbook/timeline/models"
cat >"$allowed_assessment_candidate_root/apps/web/src/workbook/hooks/useAssessmentSupportCandidates.ts" <<'TS'
import { normalizeTimelineFullRow } from "../timeline/models/workbookTimelineModel";

export const candidateAdapter = normalizeTimelineFullRow;
TS
allowed_assessment_candidate_output="$(assert_passes "allowed assessment candidate Timeline adapter" run_checker "$allowed_assessment_candidate_root")"
assert_contains "$allowed_assessment_candidate_output" "frontend import boundaries verified" "allowed assessment candidate adapter output"

allowed_dom_unit_root="$(prepare_case_root allowed-dom-unit)"
cat >"$allowed_dom_unit_root/packages/grid-adapter/src/domUnitBinding.test.tsx" <<'TS'
import { SemanticDataGrid } from "./domUnitBinding";

export const grid = SemanticDataGrid;
TS
allowed_dom_unit_output="$(assert_passes "allowed package DOM-unit import" run_checker "$allowed_dom_unit_root")"
assert_contains "$allowed_dom_unit_output" "frontend import boundaries verified" "allowed package DOM-unit output"

blocked_dom_unit_root="$(prepare_case_root blocked-dom-unit)"
cat >"$blocked_dom_unit_root/apps/web/src/GridDomUnitLeak.test.tsx" <<'TS'
import { SemanticDataGrid } from "../../../packages/grid-adapter/src/domUnitBinding";

export const grid = SemanticDataGrid;
TS
blocked_dom_unit_output="$(assert_fails "blocked app DOM-unit import" run_checker "$blocked_dom_unit_root")"
assert_contains "$blocked_dom_unit_output" "frontend-grid-dom-unit-binding-boundary" "blocked DOM-unit rule"
assert_contains "$blocked_dom_unit_output" "apps/web/src/GridDomUnitLeak.test.tsx" "blocked DOM-unit file"

missing_stylesheet_root="$(prepare_case_root missing-stylesheet)"
cat >"$missing_stylesheet_root/packages/grid-adapter/src/index.tsx" <<'TS'
export const gridStylesheetLoaded = false;
TS
missing_stylesheet_output="$(assert_fails "missing RDG stylesheet singleton" run_checker "$missing_stylesheet_root")"
assert_contains "$missing_stylesheet_output" "frontend-rdg-stylesheet-singleton" "missing RDG stylesheet singleton rule"
assert_contains "$missing_stylesheet_output" "expected exactly 1, found 0" "missing RDG stylesheet singleton count"

duplicate_stylesheet_root="$(prepare_case_root duplicate-stylesheet)"
cat >"$duplicate_stylesheet_root/packages/grid-adapter/src/duplicate.tsx" <<'TS'
import "react-data-grid/lib/styles.css";

export const duplicateGridStylesheetLoaded = true;
TS
duplicate_stylesheet_output="$(assert_fails "duplicate RDG stylesheet singleton" run_checker "$duplicate_stylesheet_root")"
assert_contains "$duplicate_stylesheet_output" "frontend-rdg-stylesheet-singleton" "duplicate RDG stylesheet singleton rule"
assert_contains "$duplicate_stylesheet_output" "expected exactly 1, found 2" "duplicate RDG stylesheet singleton count"

outside_stylesheet_root="$(prepare_case_root outside-stylesheet)"
cat >"$outside_stylesheet_root/apps/web/src/GridStylesheetLeak.tsx" <<'TS'
import "react-data-grid/lib/styles.css";

export const appGridStylesheetLoaded = true;
TS
outside_stylesheet_output="$(assert_fails "outside RDG stylesheet singleton" run_checker "$outside_stylesheet_root")"
assert_contains "$outside_stylesheet_output" "frontend-rdg-stylesheet-singleton" "outside RDG stylesheet singleton rule"
assert_contains "$outside_stylesheet_output" "apps/web/src/GridStylesheetLeak.tsx" "outside RDG stylesheet importer"

generated_package_root="$(prepare_case_root generated-package)"
cat >"$generated_package_root/apps/web/src/generatedProtocol.ts" <<'TS'
import { contractArtifactIndex } from "@cartulary/protocol-ts/generated";

export const contracts = contractArtifactIndex;
TS
generated_package_output="$(assert_fails "generated package error" run_checker "$generated_package_root")"
assert_contains "$generated_package_output" "error: frontend-generated-protocol-boundary" "generated package error"

generated_relative_root="$(prepare_case_root generated-relative)"
cat >"$generated_relative_root/apps/web/e2e/support.ts" <<'TS'
import { contractArtifactIndex } from "../../../packages/protocol-ts/src/generated/contracts";

export const contracts = contractArtifactIndex;
TS
generated_relative_output="$(assert_fails "generated relative error" run_checker "$generated_relative_root")"
assert_contains "$generated_relative_output" "error: frontend-generated-protocol-boundary" "generated relative error"
assert_contains "$generated_relative_output" "apps/web/e2e/support.ts" "generated relative file"

warning_fixture_root="$(prepare_case_root warning-fixture)"
cat >"$warning_fixture_root/apps/web/src/warningFixture.ts" <<'TS'
import { fixture } from "cartulary-warning-only-fixture/subpath";

export const warningFixture = fixture;
TS
warning_fixture_output="$(assert_passes "synthetic warning" run_checker "$warning_fixture_root")"
assert_contains "$warning_fixture_output" "warning: frontend-synthetic-warning-boundary" "synthetic warning"
warning_fixture_error_output="$(assert_fails "synthetic warning as error" run_checker "$warning_fixture_root" --warnings-as-errors)"
assert_contains "$warning_fixture_error_output" "error: frontend-synthetic-warning-boundary" "synthetic warning promoted to error"

family_root="$(prepare_case_root family)"
cat >"$family_root/packages/protocol-ts/package.json" <<'JSON'
{
  "name": "@cartulary/protocol-ts",
  "exports": {
    "./http": "./src/entrypoints/http.ts"
  }
}
JSON
cat >"$family_root/apps/web/src/contracts.ts" <<'TS'
import { buildHTTPOperationPath } from "@cartulary/protocol-ts/http";

export const buildPath = buildHTTPOperationPath;
TS
family_output="$(assert_passes "protocol family import" run_checker "$family_root")"
assert_contains "$family_output" "frontend import boundaries verified" "family import output"
assert_not_contains "$family_output" "frontend-generated-protocol-boundary" "family import must not warn"

unsupported_root="$(prepare_case_root unsupported-root)"
cat >"$unsupported_root/packages/protocol-ts/package.json" <<'JSON'
{
  "name": "@cartulary/protocol-ts",
  "exports": {
    "./http": "./src/entrypoints/http.ts"
  }
}
JSON
cat >"$unsupported_root/apps/web/src/contracts.ts" <<'TS'
import { legacy } from "@cartulary/protocol-ts";

export const leaked = legacy;
TS
unsupported_root_output="$(assert_fails "unsupported protocol root import" run_checker "$unsupported_root")"
assert_contains "$unsupported_root_output" "frontend-workspace-package-facade-boundary" "unsupported protocol root rule"

protocol_entrypoint_root="$(prepare_case_root protocol-entrypoint)"
cat >"$protocol_entrypoint_root/packages/protocol-ts/src/entrypoints/errors.ts" <<'TS'
import { errorRegistry } from "../generated/error-registry.js";

export { errorRegistry };
TS
protocol_entrypoint_output="$(assert_passes "protocol entrypoint generated import" run_checker "$protocol_entrypoint_root")"
assert_contains "$protocol_entrypoint_output" "frontend import boundaries verified" "protocol entrypoint generated import output"

protocol_root_bypass_root="$(prepare_case_root protocol-root-bypass)"
cat >"$protocol_root_bypass_root/packages/protocol-ts/src/index.ts" <<'TS'
import { errorArtifacts } from "./generated/errors-artifacts.js";

export const artifacts = errorArtifacts;
TS
protocol_root_bypass_output="$(assert_fails "protocol root generated import" run_checker "$protocol_root_bypass_root")"
assert_contains "$protocol_root_bypass_output" "frontend-generated-protocol-boundary" "protocol root generated import rule"

protocol_wrong_entrypoint_root="$(prepare_case_root protocol-wrong-entrypoint)"
cat >"$protocol_wrong_entrypoint_root/packages/protocol-ts/src/entrypoints/collaboration.ts" <<'TS'
import { errorRegistry } from "../generated/error-registry.js";

export { errorRegistry };
TS
protocol_wrong_entrypoint_output="$(assert_fails "protocol wrong entrypoint import" run_checker "$protocol_wrong_entrypoint_root")"
assert_contains "$protocol_wrong_entrypoint_output" "frontend-generated-protocol-contract-artifact-owner" "protocol wrong entrypoint rule"

protocol_barrel_root="$(prepare_case_root protocol-barrel)"
cat >"$protocol_barrel_root/packages/protocol-ts/src/entrypoints/errors.ts" <<'TS'
import { errorArtifacts } from "../generated/index.js";

export const artifacts = errorArtifacts;
TS
protocol_barrel_output="$(assert_fails "protocol barrel import" run_checker "$protocol_barrel_root")"
assert_contains "$protocol_barrel_output" "frontend-generated-protocol-barrel-deny" "protocol barrel deny rule"

protocol_obsolete_artifact_root="$(prepare_case_root protocol-obsolete-artifact)"
cat >"$protocol_obsolete_artifact_root/packages/protocol-ts/src/entrypoints/errors.ts" <<'TS'
import { errorArtifacts } from "../generated/errors-artifacts.js";

export { errorArtifacts };
TS
protocol_obsolete_artifact_output="$(assert_fails "protocol obsolete artifact import" run_checker "$protocol_obsolete_artifact_root")"
assert_contains "$protocol_obsolete_artifact_output" "frontend-generated-protocol-obsolete-artifact-projection-deny" "protocol obsolete artifact deny rule"

protocol_validator_barrel_root="$(prepare_case_root protocol-validator-barrel)"
cat >"$protocol_validator_barrel_root/packages/protocol-ts/src/entrypoints/http.ts" <<'TS'
export { validateCartularyCoreHttpErrorEnvelopeV1 } from "../generated/protocol-validators.js";
TS
protocol_validator_barrel_output="$(assert_fails "protocol validator barrel import" run_checker "$protocol_validator_barrel_root")"
assert_contains "$protocol_validator_barrel_output" "frontend-generated-protocol-validator-barrel-deny" "protocol validator barrel deny rule"

protocol_audit_root="$(prepare_case_root protocol-audit)"
cat >"$protocol_audit_root/packages/protocol-ts/src/entrypoints/errors.ts" <<'TS'
import { auditArtifacts } from "../generated/audit-artifacts.js";

export const artifacts = auditArtifacts;
TS
protocol_audit_output="$(assert_fails "protocol audit import" run_checker "$protocol_audit_root")"
assert_contains "$protocol_audit_output" "frontend-generated-protocol-audit-deny" "protocol audit deny rule"

protocol_revisions_root="$(prepare_case_root protocol-revisions)"
cat >"$protocol_revisions_root/packages/protocol-ts/src/entrypoints/errors.ts" <<'TS'
import { revisionsArtifacts } from "../generated/revisions-artifacts.js";

export const artifacts = revisionsArtifacts;
TS
protocol_revisions_output="$(assert_fails "protocol revisions import" run_checker "$protocol_revisions_root")"
assert_contains "$protocol_revisions_output" "frontend-generated-protocol-revisions-deny" "protocol revisions deny rule"

design_token_owner_root="$(prepare_case_root design-token-owner)"
cat >"$design_token_owner_root/packages/ui-contracts/src/index.ts" <<'TS'
export { cartularyDesignTokenVars } from "./generated/design-tokens";
TS
design_token_owner_output="$(assert_passes "ui-contracts generated token import" run_checker "$design_token_owner_root")"
assert_contains "$design_token_owner_output" "frontend import boundaries verified" "ui-contracts generated token import output"

design_token_relative_root="$(prepare_case_root design-token-relative)"
cat >"$design_token_relative_root/apps/web/src/designTokenBypass.ts" <<'TS'
import { cartularyDesignTokenVars } from "../../../packages/ui-contracts/src/generated/design-tokens";

export const vars = cartularyDesignTokenVars;
TS
design_token_relative_output="$(assert_fails "generated design token relative error" run_checker "$design_token_relative_root")"
assert_contains "$design_token_relative_output" "frontend-generated-design-token-boundary" "generated design token relative rule"
assert_contains "$design_token_relative_output" "apps/web/src/designTokenBypass.ts" "generated design token relative file"

raw_design_literal_root="$(prepare_case_root raw-design-literal)"
cat >"$raw_design_literal_root/apps/web/src/rawDesignLiteral.tsx" <<'TS'
export const accent = "#FACC15";
TS
raw_design_literal_output="$(assert_fails "raw design color literal error" run_checker "$raw_design_literal_root")"
assert_contains "$raw_design_literal_output" "frontend-runtime-raw-design-color-literals" "raw design color literal rule"
assert_contains "$raw_design_literal_output" "#FACC15" "raw design color literal value"

node_runtime_root="$(prepare_case_root node-runtime)"
cat >"$node_runtime_root/apps/web/src/nodeRuntimeLeak.ts" <<'TS'
import path from "node:path";

export const joined = path.join("browser", "runtime");
TS
node_runtime_output="$(assert_fails "node builtin runtime error" run_checker "$node_runtime_root")"
assert_contains "$node_runtime_output" "frontend-runtime-node-boundary" "node builtin runtime rule"
assert_contains "$node_runtime_output" "node:path" "node builtin runtime specifier"

test_utils_dependency_root="$(prepare_case_root test-utils-runtime-dependency)"
cat >"$test_utils_dependency_root/packages/test-utils/src/grid.ts" <<'TS'
import type { Page } from "@playwright/test";
import path from "node:path";

export const leaked = { path, page: undefined as Page | undefined };
TS
test_utils_dependency_output="$(assert_fails "test-utils runtime dependency error" run_checker "$test_utils_dependency_root")"
assert_contains "$test_utils_dependency_output" "frontend-test-utils-runtime-dependency-boundary" "test-utils runtime dependency rule"
assert_contains "$test_utils_dependency_output" "@playwright/test" "test-utils Playwright dependency specifier"
assert_contains "$test_utils_dependency_output" "node:path" "test-utils Node dependency specifier"

test_utils_cycle_root="$(prepare_case_root test-utils-cycle)"
cat >"$test_utils_cycle_root/packages/test-utils/src/grid.ts" <<'TS'
import { action } from "./grid-actions";

export const grid = action;
TS
cat >"$test_utils_cycle_root/packages/test-utils/src/grid-actions.ts" <<'TS'
import { grid } from "./grid";

export const action = grid;
TS
test_utils_cycle_output="$(assert_fails "test-utils production cycle" run_checker "$test_utils_cycle_root")"
assert_contains "$test_utils_cycle_output" "frontend-test-utils-production-acyclic-import-graph" "test-utils cycle rule"

test_utils_observer_root="$(prepare_case_root test-utils-observer-direction)"
cat >"$test_utils_observer_root/packages/test-utils/src/grid-actions.ts" <<'TS'
export const mutateGrid = true;
TS
cat >"$test_utils_observer_root/packages/test-utils/src/grid-setup.ts" <<'TS'
export const prepareGrid = true;
TS
cat >"$test_utils_observer_root/packages/test-utils/src/marker.ts" <<'TS'
import { mutateGrid } from "./grid-actions";
import { prepareGrid } from "./grid-setup";

export const observer = { mutateGrid, prepareGrid };
TS
test_utils_observer_output="$(assert_fails "test-utils observer direction" run_checker "$test_utils_observer_root")"
assert_contains "$test_utils_observer_output" "frontend-test-utils-observer-direction" "test-utils observer direction rule"
assert_contains "$test_utils_observer_output" "./grid-actions" "test-utils observer action dependency"
assert_contains "$test_utils_observer_output" "./grid-setup" "test-utils observer setup dependency"

node_e2e_root="$(prepare_case_root node-e2e)"
cat >"$node_e2e_root/apps/web/e2e/nodeHarness.ts" <<'TS'
import path from "node:path";

export const joined = path.join("e2e", "harness");
TS
node_e2e_output="$(assert_passes "node builtin e2e allowed" run_checker "$node_e2e_root")"
assert_contains "$node_e2e_output" "frontend import boundaries verified" "node builtin e2e output"

package_subpath_root="$(prepare_case_root package-subpath)"
cat >"$package_subpath_root/apps/web/src/viewContractBypass.ts" <<'TS'
import { internal } from "@cartulary/view-contracts/src/index";

export const leaked = internal;
TS
package_subpath_output="$(assert_fails "workspace package subpath error" run_checker "$package_subpath_root")"
assert_contains "$package_subpath_output" "frontend-workspace-package-facade-boundary" "workspace package subpath rule"
assert_contains "$package_subpath_output" "@cartulary/view-contracts/src/index" "workspace package subpath specifier"

relative_package_root="$(prepare_case_root relative-package)"
cat >"$relative_package_root/apps/web/src/viewContractRelativeBypass.ts" <<'TS'
import { internal } from "../../../packages/view-contracts/src/index";

export const leaked = internal;
TS
relative_package_output="$(assert_fails "workspace relative source error" run_checker "$relative_package_root")"
assert_contains "$relative_package_output" "frontend-workspace-package-facade-boundary" "workspace relative source rule"
assert_contains "$relative_package_output" "apps/web/src/viewContractRelativeBypass.ts" "workspace relative source file"

declared_package_export_root="$(prepare_case_root declared-package-export)"
cat >"$declared_package_export_root/apps/web/src/GridAdapter.test.tsx" <<'TS'
import { SemanticDataGrid } from "@cartulary/grid-adapter/test-support";
import { render } from "@testing-library/react";
import { it } from "vitest";

export const testGrid = { SemanticDataGrid, render, it };
TS
declared_package_export_output="$(assert_passes "declared package export in test allowed" run_checker "$declared_package_export_root")"
assert_contains "$declared_package_export_output" "frontend import boundaries verified" "declared package export output"

test_helper_runtime_root="$(prepare_case_root test-helper-runtime)"
mkdir -p "$test_helper_runtime_root/apps/web/src/testing"
cat >"$test_helper_runtime_root/apps/web/src/testing/timelineWorkbookTestSupport.ts" <<'TS'
export const support = true;
TS
cat >"$test_helper_runtime_root/apps/web/src/runtimeHelperLeak.tsx" <<'TS'
import { helper } from "@cartulary/test-utils";
import { SemanticDataGrid } from "@cartulary/grid-adapter/test-support";
import { test } from "@playwright/test";
import { render } from "@testing-library/react";
import { describe } from "vitest";
import { support } from "./testing/timelineWorkbookTestSupport";

export const leaked = { SemanticDataGrid, describe, helper, render, support, test };
TS
test_helper_runtime_output="$(assert_fails "runtime test helper error" run_checker "$test_helper_runtime_root")"
assert_contains "$test_helper_runtime_output" "frontend-runtime-test-helper-boundary" "runtime test helper rule"
for specifier in "@cartulary/test-utils" "@cartulary/grid-adapter/test-support" "@playwright/test" "@testing-library/react" "vitest" "./testing/timelineWorkbookTestSupport"; do
  assert_contains "$test_helper_runtime_output" "$specifier" "runtime test helper specifier $specifier"
done

test_helper_allowed_root="$(prepare_case_root test-helper-allowed)"
cat >"$test_helper_allowed_root/apps/web/src/runtimeHelperAllowed.test.tsx" <<'TS'
import { helper } from "@cartulary/test-utils/grid";
import { render } from "@testing-library/react";
import { describe } from "vitest";

export const allowed = { describe, helper, render };
TS
cat >"$test_helper_allowed_root/apps/web/e2e/runtimeHarness.ts" <<'TS'
import { test } from "@playwright/test";
import { helper } from "@cartulary/test-utils/grid";

export const allowed = { helper, test };
TS
test_helper_allowed_output="$(assert_passes "test helper imports allowed in tests" run_checker "$test_helper_allowed_root")"
assert_contains "$test_helper_allowed_output" "frontend import boundaries verified" "test helper allowed output"

e2e_semantic_allowed_root="$(prepare_case_root e2e-semantic-allowed)"
cat >"$e2e_semantic_allowed_root/apps/web/e2e/runtimeHarness.ts" <<'TS'
import { helper } from "@cartulary/test-utils/grid";
import { metadata } from "./support/workbook/query";

export const allowed = { helper, metadata };
TS
e2e_semantic_allowed_output="$(assert_passes "semantic E2E imports allowed" run_checker "$e2e_semantic_allowed_root")"
assert_contains "$e2e_semantic_allowed_output" "frontend import boundaries verified" "semantic E2E allowed output"

e2e_root_test_utils_root="$(prepare_case_root e2e-root-test-utils)"
cat >"$e2e_root_test_utils_root/apps/web/e2e/runtimeHarness.ts" <<'TS'
import { helper } from "@cartulary/test-utils";

export const leaked = helper;
TS
e2e_root_test_utils_output="$(assert_fails "root E2E test-utils import" run_checker "$e2e_root_test_utils_root")"
assert_contains "$e2e_root_test_utils_output" "web-e2e-test-utils-subpath-boundary" "root E2E test-utils rule"

e2e_removed_test_utils_alias_root="$(prepare_case_root e2e-removed-test-utils-alias)"
cat >"$e2e_removed_test_utils_alias_root/apps/web/e2e/runtimeHarness.ts" <<'TS'
import { helper } from "@cartulary/test-utils/visual";

export const leaked = helper;
TS
e2e_removed_test_utils_alias_output="$(assert_fails "removed E2E test-utils alias" run_checker "$e2e_removed_test_utils_alias_root")"
assert_contains "$e2e_removed_test_utils_alias_output" "frontend-workspace-package-facade-boundary" "removed E2E test-utils alias rule"

e2e_private_test_utils_path_root="$(prepare_case_root e2e-private-test-utils-path)"
cat >"$e2e_private_test_utils_path_root/apps/web/e2e/runtimeHarness.ts" <<'TS'
import { helper } from "@cartulary/test-utils/browser";

export const leaked = helper;
TS
e2e_private_test_utils_path_output="$(assert_fails "private E2E test-utils path" run_checker "$e2e_private_test_utils_path_root")"
assert_contains "$e2e_private_test_utils_path_output" "frontend-workspace-package-facade-boundary" "private E2E test-utils path rule"

e2e_app_registry_root="$(prepare_case_root e2e-app-registry)"
mkdir -p "$e2e_app_registry_root/apps/web/src/workbook/models"
cat >"$e2e_app_registry_root/apps/web/e2e/runtimeHarness.ts" <<'TS'
import { registry } from "../src/workbook/models/workbookSurfaceRegistry";

export const leaked = registry;
TS
e2e_app_registry_output="$(assert_fails "E2E app workbook registry import" run_checker "$e2e_app_registry_root")"
assert_contains "$e2e_app_registry_output" "web-e2e-workbook-contract-boundary" "E2E app workbook registry rule"

e2e_public_control_root="$(prepare_case_root e2e-public-control)"
cat >"$e2e_public_control_root/apps/web/e2e/support/transport/publicJsonClient.ts" <<'TS'
import { TestControlClient } from "./testControlClient";

export const leaked = TestControlClient;
TS
e2e_public_control_output="$(assert_fails "public transport control import" run_checker "$e2e_public_control_root")"
assert_contains "$e2e_public_control_output" "web-e2e-public-control-transport-boundary" "public transport control rule"

e2e_semantic_visual_root="$(prepare_case_root e2e-semantic-visual)"
cat >"$e2e_semantic_visual_root/apps/web/e2e/support/evidence/fixtures.ts" <<'TS'
import { visualFixture } from "../visual/fixtures";

export const leaked = visualFixture;
TS
e2e_semantic_visual_output="$(assert_fails "semantic fixture visual import" run_checker "$e2e_semantic_visual_root")"
assert_contains "$e2e_semantic_visual_output" "web-e2e-semantic-fixture-direction" "semantic fixture visual rule"

e2e_core_evidence_root="$(prepare_case_root e2e-core-evidence)"
cat >"$e2e_core_evidence_root/apps/web/e2e/support/workbook/query.ts" <<'TS'
import { evidenceFixture } from "../evidence/fixtures";

export const leaked = evidenceFixture;
TS
e2e_core_evidence_output="$(assert_fails "core support evidence import" run_checker "$e2e_core_evidence_root")"
assert_contains "$e2e_core_evidence_output" "web-e2e-core-support-direction" "core support direction rule"

protocol_entrypoint_allowed_root="$(prepare_case_root protocol-entrypoint-allowed)"
mkdir -p "$protocol_entrypoint_allowed_root/contracts/protocol-ts"
cat >"$protocol_entrypoint_allowed_root/contracts/protocol-ts/frontend-entrypoints.v2.json" <<'JSON'
{
  "schema_id": "cartulary.protocol_ts_frontend_entrypoints.v2",
  "entrypoints": [
    {
      "specifier": "@cartulary/protocol-ts/collaboration",
      "authored_path": "packages/protocol-ts/src/entrypoints/collaboration.ts",
      "generated_module_allowlist": [
        "packages/protocol-ts/src/generated/collaboration-types.ts"
      ]
    }
  ]
}
JSON
cat >"$protocol_entrypoint_allowed_root/packages/protocol-ts/src/entrypoints/collaboration.ts" <<'TS'
export type { IncidentStreamMessage } from "../generated/collaboration-types.js";
TS
cat >"$protocol_entrypoint_allowed_root/packages/protocol-ts/package.json" <<'JSON'
{
  "name": "@cartulary/protocol-ts",
  "exports": {
    "./collaboration": "./src/entrypoints/collaboration.ts"
  }
}
JSON
protocol_entrypoint_allowed_output="$(assert_passes "owner-allowed protocol entrypoint import" run_checker "$protocol_entrypoint_allowed_root")"
assert_contains "$protocol_entrypoint_allowed_output" "frontend import boundaries verified" "owner-allowed protocol entrypoint output"

protocol_entrypoint_forbidden_root="$(prepare_case_root protocol-entrypoint-forbidden)"
mkdir -p "$protocol_entrypoint_forbidden_root/contracts/protocol-ts"
cp "$protocol_entrypoint_allowed_root/contracts/protocol-ts/frontend-entrypoints.v2.json" "$protocol_entrypoint_forbidden_root/contracts/protocol-ts/frontend-entrypoints.v2.json"
cp "$protocol_entrypoint_allowed_root/packages/protocol-ts/package.json" "$protocol_entrypoint_forbidden_root/packages/protocol-ts/package.json"
cat >"$protocol_entrypoint_forbidden_root/packages/protocol-ts/src/entrypoints/collaboration.ts" <<'TS'
export { validateCartularyCoreHttpErrorEnvelopeV1 } from "../generated/core-http-validators.js";
TS
protocol_entrypoint_forbidden_output="$(assert_fails "owner-forbidden protocol entrypoint import" run_checker "$protocol_entrypoint_forbidden_root")"
assert_contains "$protocol_entrypoint_forbidden_output" "frontend-protocol-ts-entrypoint-generated-allowlist" "owner-forbidden protocol entrypoint rule"

protocol_export_map_mismatch_root="$(prepare_case_root protocol-export-map-mismatch)"
mkdir -p "$protocol_export_map_mismatch_root/contracts/protocol-ts"
cp "$protocol_entrypoint_allowed_root/contracts/protocol-ts/frontend-entrypoints.v2.json" "$protocol_export_map_mismatch_root/contracts/protocol-ts/frontend-entrypoints.v2.json"
cat >"$protocol_export_map_mismatch_root/packages/protocol-ts/package.json" <<'JSON'
{
  "name": "@cartulary/protocol-ts",
  "exports": {
    ".": "./src/index.ts",
    "./collaboration": "./src/entrypoints/collaboration.ts"
  }
}
JSON
cat >"$protocol_export_map_mismatch_root/packages/protocol-ts/src/entrypoints/collaboration.ts" <<'TS'
export type { IncidentStreamMessage } from "../generated/collaboration-types.js";
TS
protocol_export_map_mismatch_output="$(assert_fails "owner-mismatched protocol export map" run_checker "$protocol_export_map_mismatch_root")"
assert_contains "$protocol_export_map_mismatch_output" "frontend-protocol-ts-entrypoint-export-map" "owner-mismatched protocol export map rule"
