import { createHash } from "node:crypto";
import {
  chmodSync,
  mkdirSync,
  mkdtempSync,
  readdirSync,
  readFileSync,
  renameSync,
  rmSync,
  statSync,
  symlinkSync,
  writeFileSync,
} from "node:fs";
import path from "node:path";
import { expect, it } from "vitest";
import type { FixtureExecutionSummary } from "./networkFlowFixtureTestSupport";
import {
  decodeFixtureJSON,
  FixtureExecutionError,
  runNetworkFlowFixture,
  selectedFixtureRunRoot,
} from "./networkFlowFixtureTestSupport";

const fixtureID = "NF-FIX-999-admission-controls";
const csv = "a,b\n1,2\n";
const hash = (data: Uint8Array | string): string =>
  createHash("sha256").update(data).digest("hex");
type Entry = { logical_path: string; size_bytes: number; sha256: string };
const framed = (files: Entry[]): string =>
  hash(
    files
      .map(
        (file) =>
          file.logical_path +
          "\0" +
          file.sha256 +
          "\0" +
          file.size_bytes +
          "\n",
      )
      .join(""),
  );

function seed(corpus: string) {
  const root = path.join(corpus, fixtureID);
  for (const directory of ["source", "expected", "transcripts"])
    mkdirSync(path.join(root, directory), { recursive: true });
  function file(logicalPath: string, data: string) {
    writeFileSync(path.join(root, logicalPath), data);
    return {
      logical_path: logicalPath,
      size_bytes: Buffer.byteLength(data),
      sha256: hash(data),
    };
  }
  const manifest = {
    schema_id: "cartulary.network_flow_fixture_manifest.v2",
    manifest_version: 2,
    fixture_id: fixtureID,
    profile_id: "network_flow_activity",
    freeze: {
      status: "frozen",
      revision: 1,
      change_policy: "new_fixture_revision_required",
    },
    source_files: [
      {
        ...file("source/controls.csv", csv),
        media_type: "text/csv; charset=utf-8",
        role: "input",
        newline_policy: "lf_required",
      },
      {
        ...file(
          "source/controls.scenario.json",
          JSON.stringify({
            schema_id: "cartulary.network_flow_fixture_scenario.v2",
            fixture_id: fixtureID,
            summary:
              "Synthetic file admission controls, without product interpretation.",
          }),
        ),
        media_type: "application/json; charset=utf-8",
        role: "support",
        newline_policy: "lf_required",
      },
    ],
    expected_artifacts: [
      {
        ...file("expected/controls.expected.json", "{}\n"),
        media_type: "application/json; charset=utf-8",
        role: "mapping",
      },
    ],
    transcript_files: [
      {
        ...file("transcripts/controls.transcript.json", "{}\n"),
        media_type: "application/json; charset=utf-8",
        transcript_kind: "preview",
      },
    ],
    source_bundle_sha256: "",
    expected_bundle_sha256: "",
  };
  return { root, manifest };
}
type Seed = ReturnType<typeof seed>;
function publish({ root, manifest }: Seed): void {
  manifest.source_bundle_sha256 = framed(manifest.source_files);
  manifest.expected_bundle_sha256 = framed([
    ...manifest.expected_artifacts,
    ...manifest.transcript_files,
  ]);
  writeFileSync(
    path.join(root, "manifest.json"),
    JSON.stringify(manifest) + "\n",
  );
}
function source(value: Seed, index = 0) {
  const file = value.manifest.source_files[index];
  if (!file) throw new Error("Synthetic fixture source missing");
  return file;
}
async function failedRun(
  config: { fixtureID: string; revision: number; corpusRoot: string },
  compare: Parameters<typeof runNetworkFlowFixture>[1],
) {
  try {
    await runNetworkFlowFixture(config, compare);
  } catch (error) {
    expect(error).toBeInstanceOf(FixtureExecutionError);
    if (!(error instanceof FixtureExecutionError)) throw error;
    return {
      error,
      summary: (await decodeFixtureJSON(
        readFileSync(error.summaryPath),
      )) as FixtureExecutionSummary,
    };
  }
  throw new Error("Invalid fixture unexpectedly qualified");
}

it("Node fixture admission seals verified inputs and retains bounded comparison evidence", async () => {
  const runRoot = selectedFixtureRunRoot();
  const corpus = mkdtempSync(path.join(runRoot, "node-fixture-controls-"));
  const config = { fixtureID, revision: 1, corpusRoot: corpus };
  function reset(): Seed {
    rmSync(path.join(corpus, fixtureID), { recursive: true, force: true });
    return seed(corpus);
  }
  try {
    const negatives: [string, (value: Seed) => void][] = [
      [
        "traversal",
        (value) => {
          source(value).logical_path = "source/../outside.csv";
        },
      ],
      [
        "absolute",
        (value) => {
          source(value).logical_path = "/outside.csv";
        },
      ],
      [
        "dot",
        (value) => {
          source(value).logical_path = "source/./controls.csv";
        },
      ],
      [
        "duplicate",
        (value) => {
          value.manifest.source_files.push({ ...source(value) });
        },
      ],
      [
        "cross-inventory-duplicate",
        (value) => {
          const entry = value.manifest.expected_artifacts[0];
          if (!entry) throw new Error("Synthetic expected artifact missing");
          entry.logical_path = source(value).logical_path;
        },
      ],
      [
        "order",
        (value) => {
          value.manifest.source_files.reverse();
        },
      ],
      [
        "size",
        (value) => {
          source(value).size_bytes += 1;
        },
      ],
      [
        "hash",
        (value) => {
          source(value).sha256 = "0".repeat(64);
        },
      ],
      [
        "missing",
        (value) => {
          rmSync(path.join(value.root, "source/controls.csv"));
        },
      ],
      [
        "extra",
        (value) => {
          writeFileSync(path.join(value.root, "extra.csv"), csv);
        },
      ],
      [
        "leaf-symlink",
        (value) => {
          renameSync(
            path.join(value.root, "source/controls.csv"),
            path.join(corpus, "outside.csv"),
          );
          symlinkSync(
            path.join(corpus, "outside.csv"),
            path.join(value.root, "source/controls.csv"),
          );
        },
      ],
      [
        "parent-symlink",
        (value) => {
          renameSync(
            path.join(value.root, "source"),
            path.join(value.root, "source-real"),
          );
          symlinkSync("source-real", path.join(value.root, "source"));
        },
      ],
      [
        "closed-manifest",
        (value) => {
          Object.assign(value.manifest, { unknown: true });
        },
      ],
      [
        "identity",
        (value) => {
          value.manifest.fixture_id = "NF-FIX-998-wrong";
        },
      ],
      [
        "revision",
        (value) => {
          value.manifest.freeze.revision += 1;
        },
      ],
      [
        "draft",
        (value) => {
          value.manifest.freeze.status = "draft";
        },
      ],
      [
        "scenario",
        (value) => {
          const data = JSON.stringify({
            schema_id: "cartulary.network_flow_fixture_scenario.v2",
            fixture_id: "NF-FIX-998-wrong",
            summary: "Wrong identity.",
          });
          writeFileSync(
            path.join(value.root, source(value, 1).logical_path),
            data,
          );
          Object.assign(source(value, 1), {
            size_bytes: Buffer.byteLength(data),
            sha256: hash(data),
          });
        },
      ],
      [
        "closed-scenario",
        (value) => {
          const data = JSON.stringify({
            schema_id: "cartulary.network_flow_fixture_scenario.v2",
            fixture_id: fixtureID,
            summary: "Wrong member.",
            unknown: true,
          });
          writeFileSync(
            path.join(value.root, source(value, 1).logical_path),
            data,
          );
          Object.assign(source(value, 1), {
            size_bytes: Buffer.byteLength(data),
            sha256: hash(data),
          });
        },
      ],
    ];
    for (const [name, mutate] of negatives) {
      const value = reset();
      mutate(value);
      publish(value);
      let callbacks = 0;
      const { summary } = await failedRun(config, () => {
        callbacks += 1;
      });
      expect(callbacks, name).toBe(0);
      expect(summary.failure_phase, name).toBe("admission");
      expect(summary.comparison_status, name).toBe(
        name === "draft" ? "blocked" : "failed",
      );
      expect(summary.produced_artifacts, name).toEqual([]);
    }
    for (const corrupt of [
      (text: string) =>
        text.replace(
          /"source_bundle_sha256":"[a-f0-9]{64}"/u,
          '"source_bundle_sha256":"' + "0".repeat(64) + '"',
        ),
      (text: string) =>
        text.replace(
          /"expected_bundle_sha256":"[a-f0-9]{64}"/u,
          '"expected_bundle_sha256":"' + "0".repeat(64) + '"',
        ),
      (text: string) =>
        text.replace(
          '"manifest_version":2',
          '"manifest_version":2,"manifest_version":2',
        ),
      (text: string) => text.replace('"network_flow_activity"', '"\\ud800"'),
      () => Buffer.from([0xff]),
      () => " ".repeat(1024 * 1024 + 1),
      (text: string) => "\uFEFF" + text,
    ]) {
      const value = reset();
      publish(value);
      const manifest = path.join(value.root, "manifest.json");
      writeFileSync(manifest, corrupt(readFileSync(manifest, "utf8")));
      let callbacks = 0;
      const { summary } = await failedRun(config, () => {
        callbacks += 1;
      });
      expect(callbacks).toBe(0);
      expect(summary.failure_phase).toBe("admission");
    }

    const value = reset();
    publish(value);
    const result = await runNetworkFlowFixture(config, (bundle) => {
      writeFileSync(
        path.join(value.root, "source/controls.csv"),
        "changed source",
      );
      expect(bundle.read("source/controls.csv").toString()).toBe(csv);
      expect(() => bundle.read("manifest.json")).toThrow("Unlisted");
      expect(() => bundle.record("../escape", csv)).toThrow("Invalid");
      bundle.record("actual.json", "{}\n");
      expect(() => bundle.record("actual.json", "replace")).toThrow();
    });
    expect(result.summary.comparison_status).toBe("passed");
    expect(result.summary.failure_phase).toBeNull();
    expect(result.summary.manifest_sha256).toBe(
      hash(readFileSync(path.join(value.root, "manifest.json"))),
    );
    const inputs = path.join(
      runRoot,
      result.summary.materialized_input_root ?? "",
    );
    expect(statSync(inputs).mode & 0o777).toBe(0o500);
    expect(
      statSync(path.join(inputs, "source/controls.csv")).mode & 0o777,
    ).toBe(0o400);
    expect(result.summary.produced_artifacts).toEqual([
      {
        logical_path: path.relative(
          runRoot,
          path.join(path.dirname(result.summaryPath), "produced/actual.json"),
        ),
        size_bytes: 3,
        sha256: hash("{}\n"),
      },
    ]);
    expect(statSync(result.summaryPath).mode & 0o777).toBe(0o600);

    const valid = reset();
    publish(valid);
    const linkedCorpus = path.join(corpus, "linked-corpus");
    symlinkSync(corpus, linkedCorpus);
    let linkedCallbacks = 0;
    const linked = await failedRun(
      { ...config, corpusRoot: linkedCorpus },
      () => {
        linkedCallbacks += 1;
      },
    );
    expect(linkedCallbacks).toBe(0);
    expect(linked.summary.failure_phase).toBe("admission");
    rmSync(linkedCorpus);
    const wrong = await failedRun(config, (bundle) => {
      bundle.record("actual.json", '{"wrong":true}\n');
      throw new Error("Deliberately wrong expectation");
    });
    expect(wrong.summary.comparison_status).toBe("failed");
    expect(wrong.summary.failure_phase).toBe("comparison");
    expect(wrong.summary.produced_artifacts).toHaveLength(1);
    const bounded = await runNetworkFlowFixture(config, (bundle) => {
      for (let i = 0; i < 64; i += 1) bundle.record("artifact-" + i, "");
      expect(() => bundle.record("overflow", "")).toThrow("excessive");
    });
    expect(bounded.summary.produced_artifacts).toHaveLength(64);
    const parallel = await Promise.all(
      [0, 1].map(async (index) =>
        runNetworkFlowFixture(config, async (bundle) => {
          await Promise.resolve();
          bundle.record("same-name.json", JSON.stringify({ index }));
        }),
      ),
    );
    expect(new Set(parallel.map((item) => item.summaryPath)).size).toBe(2);
    expect(
      new Set(parallel.map((item) => item.summary.materialized_input_root))
        .size,
    ).toBe(2);

    const previous = new Set(readdirSync(runRoot));
    let publicationDirectory = "";
    await expect(
      runNetworkFlowFixture(config, () => {
        const added = readdirSync(runRoot).filter(
          (entry) =>
            entry.startsWith(fixtureID + "-node-") && !previous.has(entry),
        );
        expect(added).toHaveLength(1);
        publicationDirectory = path.join(
          runRoot,
          added[0] ?? "",
          "comparison-summary.json",
        );
        mkdirSync(publicationDirectory);
      }),
    ).rejects.toBeInstanceOf(FixtureExecutionError);
    rmSync(publicationDirectory, { recursive: true });
  } finally {
    // Mutable attack corpora are scratch; verified copies and summaries remain.
    chmodSync(corpus, 0o700);
    rmSync(corpus, { recursive: true, force: true });
  }
});
