import { readFileSync } from "node:fs";
import path from "node:path";
import { prepareBoundaryPolicy, evaluateBoundaryPolicy } from "../backend-module-boundary-check.mjs";
import { repository } from "./fixtures.mjs";

export function emptyPolicy(overrides = {}) {
  return {
    schema_id: "cartulary.backend_module_boundaries.v4", scan_roots: [], scan_excludes: [], exact_file_sets: [],
    migration_source_boundary: { scan_roots: [], restricted_imports: [], forbidden_declarations: [], exact_calls: [] },
    owner_port_only_imports: [], raw_ndjson_targets: [], forbidden_route_dependencies: [], forbidden_go_imports: [],
    go_import_allowlists: [], source_table_access: [], forbidden_source_mappings: [], sql_table_allowlists: [],
    sql_table_access: [], forbidden_go_calls: [], command_root_shape: { root: "cmd", allowed_go_files: [] },
    forbidden_source_tokens: [], forbidden_test_build_tokens: [],
    generated_root_writes: { scan_roots: [], generated_roots: [], write_symbols: [], allowed_paths: [] },
    ...overrides,
  };
}

export function repositoryPolicy() {
  return prepareBoundaryPolicy(JSON.parse(readFileSync(path.join(repository, "tools/backend_module_boundaries.json"), "utf8")));
}

export function inspect(files, overrides, inputs = {}) {
  return evaluateBoundaryPolicy({ ...prepareBoundaryPolicy(emptyPolicy()), ...overrides }, { files, ...inputs }).violations;
}
