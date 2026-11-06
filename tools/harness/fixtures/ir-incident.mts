import {
  closeSync,
  fstatSync,
  mkdirSync,
  readFileSync,
  readSync,
  rmSync,
} from "node:fs";
import { createRequire } from "node:module";
import path from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";
import {
  copyVerified,
  digest,
  openRegular,
  seal,
  secureDirectory,
  type FileRecord,
} from "./fixture-artifacts.mjs";

const repoRoot = fileURLToPath(new URL("../../../", import.meta.url));
const require = createRequire(import.meta.url);
const { foundationSchemaValidators } = require(
  path.join(repoRoot, "tools/harness/contract/generated/foundation-schema-validators.cjs"),
) as { foundationSchemaValidators: Record<string, (value: unknown) => boolean> };

export type IRSource = FileRecord & {
  records: number;
  columns: string[];
  use: "file_upload" | "clipboard" | "entity_alias_operations";
  field_keys: string[];
};
export type IRManifest = {
  schema_id: "cartulary.ir_fixture_manifest.v1";
  fixture_id: "ir-incident";
  revision: 1;
  source_files: IRSource[];
  expected_file: FileRecord;
  supplement_files: FileRecord[];
  dependencies: [{
    dependency_id: string;
    revision: number;
    manifest_path: string;
    manifest_sha256: string;
    datasets: string[];
  }];
};

function readBounded(root: string, name: string, limit: number): Buffer {
  const fd = openRegular(root, name);
  try {
    const size = fstatSync(fd).size;
    if (size > limit) throw new Error("Excessive IR fixture member");
    const buffer = Buffer.alloc(size + 1);
    let used = 0;
    while (used < buffer.length) {
      const count = readSync(fd, buffer, used, buffer.length - used, null);
      if (!count) break;
      used += count;
    }
    if (used !== size) throw new Error("IR fixture changed during admission");
    return buffer.subarray(0, used);
  } finally {
    closeSync(fd);
  }
}

async function strictJSON(data: Buffer): Promise<unknown> {
  const { parseStrictJSON } = (await import(
    pathToFileURL(path.join(repoRoot, "tools/harness/contract/strict-json.mjs")).href
  )) as { parseStrictJSON: (value: string) => unknown };
  return parseStrictJSON(
    new TextDecoder("utf-8", { fatal: true, ignoreBOM: true }).decode(data),
  );
}

// Called within a test, never at module load or discovery. Consumers receive
// only listed, verified copies. This helper neither creates product state nor
// interprets the supplemental record descriptions.
export async function admitIRCorpus(
  destination: string,
  fixtureRoot = path.join(repoRoot, "fixtures"),
) {
  secureDirectory(fixtureRoot);
  secureDirectory(path.dirname(destination), true);
  mkdirSync(destination, { mode: 0o700 });
  let admitted = false;
  try {
    const raw = readBounded(fixtureRoot, "ir-incident/v1/manifest.json", 1 << 20);
    const decoded = await strictJSON(raw);
    if (!foundationSchemaValidators["cartulary.ir_fixture_manifest.v1"]?.(decoded))
      throw new Error("Invalid IR fixture manifest");
    const manifest = decoded as IRManifest;
    const listed = new Map<string, FileRecord>();
    function add(prefix: string, file: FileRecord) {
      const name = prefix + file.logical_path;
      if (listed.has(name)) throw new Error("Duplicate IR fixture member");
      const entry = { ...file, logical_path: name };
      copyVerified(fixtureRoot, entry, destination);
      listed.set(name, entry);
    }
    for (const source of manifest.source_files) {
      const invalidMapping = source.use === "entity_alias_operations"
        ? source.field_keys.length !== 0
        : source.columns.length !== source.field_keys.length;
      if (invalidMapping) throw new Error("Invalid IR source mapping");
      add("ir-incident/v1/", source);
    }
    add("ir-incident/v1/", manifest.expected_file);
    for (const file of manifest.supplement_files) add("ir-incident/v1/", file);

    const dep = manifest.dependencies[0];
    const dependencyBytes = readBounded(fixtureRoot, dep.manifest_path, 1 << 20);
    if (digest(dependencyBytes) !== dep.manifest_sha256)
      throw new Error("IR dependency manifest changed");
    const orientation = (await strictJSON(dependencyBytes)) as {
      schema_id: string;
      expectation_revision: number;
      datasets: {
        name: string;
        csv: string;
        csv_sha256: string;
        expected: string;
        expected_sha256: string;
      }[];
    };
    if (
      orientation.schema_id !== "cartulary.network_flow.orientation_fixture.v1" ||
      orientation.expectation_revision !== dep.revision
    ) throw new Error("Incompatible orientation dependency");
    const found = new Set<string>();
    for (const dataset of orientation.datasets) {
      if (!dep.datasets.includes(dataset.name)) continue;
      if (found.has(dataset.name)) throw new Error("Duplicate dependency dataset");
      found.add(dataset.name);
      for (const [name, sha256] of [
        [dataset.csv, dataset.csv_sha256],
        [dataset.expected, dataset.expected_sha256],
      ] as const) {
        const data = readBounded(fixtureRoot, "network-flow-orientation/v1/" + name, 32 << 20);
        add("network-flow-orientation/v1/", { logical_path: name, size_bytes: data.length, sha256 });
      }
    }
    if (found.size !== dep.datasets.length) throw new Error("Missing dependency dataset");
    for (const [name, data] of [
      ["ir-incident/v1/manifest.json", raw],
      [dep.manifest_path, dependencyBytes],
    ] as const) {
      add("", { logical_path: name, size_bytes: data.length, sha256: digest(data) });
    }
    seal(destination);
    admitted = true;
    return {
      manifest,
      manifestSHA256: digest(raw),
      read(name: string): Buffer {
        if (!listed.has(name)) throw new Error("Unadmitted IR fixture member");
        return readFileSync(path.join(destination, name));
      },
    };
  } finally {
    if (!admitted) rmSync(destination, { recursive: true, force: true });
  }
}
