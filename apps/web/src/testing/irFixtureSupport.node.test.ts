import {
  chmodSync,
  cpSync,
  existsSync,
  mkdtempSync,
  readdirSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { expect, it } from "vitest";
import { admitIRCorpus } from "../../../../tools/harness/fixtures/ir-incident.mjs";

function writable(directory: string): void {
  chmodSync(directory, 0o700);
  for (const entry of readdirSync(directory, { withFileTypes: true }))
    if (entry.isDirectory()) writable(path.join(directory, entry.name));
}

it("IR fixture admission verifies the shared originals and dependencies before browser setup", async () => {
  const root = fileURLToPath(new URL("../../../../fixtures/", import.meta.url));
  const workspace = mkdtempSync(path.join(tmpdir(), "cartulary-ir-admission-"));
  try {
    const corpus = await admitIRCorpus(path.join(workspace, "valid"), root);
    expect(
      corpus.manifest.source_files.map((source) => source.records),
    ).toEqual([222, 30, 20, 1500, 1500]);
    const expected = JSON.parse(
      corpus.read("ir-incident/v1/expected.json").toString(),
    );
    expect(expected.flow_observations).toBe(4500);
    expect(
      expected.timeline_cases.find(
        (entry: { source_record: number }) => entry.source_record === 183,
      ).cells[7],
    ).toContain("\r\n");
    expect(() => corpus.read("../../unlisted")).toThrow("Unadmitted");
    for (const member of [
      "ir-incident/v1/timeline.csv",
      "ir-incident/v1/expected.json",
      "network-flow-orientation/v1/manifest.json",
      "network-flow-orientation/v1/day2.csv",
    ]) {
      const fixtures = path.join(workspace, "changed");
      for (const owner of ["ir-incident", "network-flow-orientation"])
        cpSync(path.join(root, owner), path.join(fixtures, owner), {
          recursive: true,
        });
      writeFileSync(path.join(fixtures, member), "tampered");
      const destination = path.join(workspace, "rejected");
      await expect(admitIRCorpus(destination, fixtures)).rejects.toThrow();
      expect(existsSync(destination)).toBe(false);
      rmSync(fixtures, { recursive: true });
    }
  } finally {
    writable(workspace);
    rmSync(workspace, { recursive: true });
  }
});
