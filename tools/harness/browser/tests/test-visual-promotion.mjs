import { publishBrowserTargetResult } from "../browser-target-finalizer-cli.mjs";
import assert from "node:assert/strict";
import { chmodSync, lstatSync, existsSync, mkdirSync, mkdtempSync, readFileSync, renameSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { promoteVisualSnapshotCandidate, stageVisualSnapshotCandidate } from "../visual-snapshot-promotion.mjs";
import test from "node:test";
const root = path.resolve(import.meta.dirname, "../../../..");
test("visual snapshot promotion rolls back a failed manifest swap and preserves private candidates", async (t) => {
    const promotionRoot = mkdtempSync(path.join(os.tmpdir(), "cartulary-visual-promotion."));
    try {
        const sourceSnapshots = path.join(promotionRoot, "source-snapshots");
        const candidateSnapshots = path.join(promotionRoot, "candidate-snapshots");
        const sourceManifest = path.join(promotionRoot, "source-manifest.json");
        const candidateManifest = path.join(promotionRoot, "candidate-manifest.json");
        const snapshotBackup = path.join(promotionRoot, "snapshot-backup");
        const manifestBackup = path.join(promotionRoot, "manifest-backup.json");
        mkdirSync(sourceSnapshots);
        writeFileSync(path.join(sourceSnapshots, "fixture.png"), "old-golden", { mode: 0o644 });
        chmodSync(path.join(sourceSnapshots, "fixture.png"), 0o644);
        stageVisualSnapshotCandidate(sourceSnapshots, candidateSnapshots);
        assert.equal(lstatSync(path.join(candidateSnapshots, "fixture.png")).mode & 0o777, 0o600);
        assert.equal(readFileSync(path.join(candidateSnapshots, "fixture.png"), "utf8"), "old-golden");
        stageVisualSnapshotCandidate(sourceSnapshots, candidateSnapshots);
        writeFileSync(path.join(candidateSnapshots, "fixture.png"), "new-golden");
        writeFileSync(sourceManifest, "old-manifest");
        writeFileSync(candidateManifest, "new-manifest");
        let renameCount = 0;
        assert.throws(() => promoteVisualSnapshotCandidate({
            sourceSnapshots,
            candidateSnapshots,
            sourceManifest,
            candidateManifest,
            snapshotBackup,
            manifestBackup,
        }, {
            existsSync,
            renameSync(from, to) {
                renameCount += 1;
                if (renameCount === 4)
                    throw new Error("injected manifest promotion failure");
                renameSync(from, to);
            },
            rmSync,
        }), /injected manifest promotion failure/u);
        assert.equal(readFileSync(path.join(sourceSnapshots, "fixture.png"), "utf8"), "old-golden");
        assert.equal(readFileSync(sourceManifest, "utf8"), "old-manifest");
        assert.equal(readFileSync(path.join(candidateSnapshots, "fixture.png"), "utf8"), "new-golden");
        assert.equal(readFileSync(candidateManifest, "utf8"), "new-manifest");
        assert.equal(existsSync(snapshotBackup), false);
        assert.equal(existsSync(manifestBackup), false);
        promoteVisualSnapshotCandidate({
            sourceSnapshots,
            candidateSnapshots,
            sourceManifest,
            candidateManifest,
            snapshotBackup,
            manifestBackup,
        });
        assert.equal(readFileSync(path.join(sourceSnapshots, "fixture.png"), "utf8"), "new-golden");
        assert.equal(readFileSync(sourceManifest, "utf8"), "new-manifest");
        assert.equal(existsSync(snapshotBackup), false);
        assert.equal(existsSync(manifestBackup), false);
    }
    finally {
        rmSync(promotionRoot, { force: true, recursive: true });
    }
});
// Exercise the publication boundary itself, including a real failed file write.
// Promotion is observable; no test depends on the finalizer's source spelling.
test("target publication failure cannot promote a visual candidate", () => {
    const directory = mkdtempSync(path.join(os.tmpdir(), "visual-target-publication-"));
    try {
        const digest = `sha256:${"a".repeat(64)}`;
        const payload = { schema_id: "cartulary.browser_target_result.v9", target_id: "browser-e2e-visual-update", status: "pass",
            generated_at: "2026-10-26T00:00:00Z", group_results: [{ group_id: "visual", browser_session_id: "visual", ref: "group.json", sha256: digest }],
            sessions: [{ browser_session_id: "visual", runtime_profile_id: "default", service_requirement: "test-services", artifacts: [
                        { kind: "stack_v7", ref: "stack.json", sha256: digest }, { kind: "startup_diagnostics_v2", ref: "startup.json", sha256: digest },
                    ] }] };
        const destination = path.join(directory, "target.json");
        const candidate = { candidate: "exact candidate" };
        let promoted = 0;
        const promote = (value) => {
            assert.equal(value, candidate);
            assert.deepEqual(JSON.parse(readFileSync(destination, "utf8")), payload);
            promoted++;
        };
        assert.throws(() => publishBrowserTargetResult(destination, { ...payload, status: "invalid" }, candidate, { promote }), /validation failed/);
        assert.equal(existsSync(destination), false);
        assert.equal(promoted, 0);
        mkdirSync(destination);
        assert.throws(() => publishBrowserTargetResult(destination, payload, candidate, { promote }));
        assert.equal(promoted, 0);
        rmSync(destination, { recursive: true });
        publishBrowserTargetResult(destination, { ...payload, status: "fail" }, candidate, { promote });
        assert.equal(promoted, 0);
        publishBrowserTargetResult(destination, payload, null, { promote });
        assert.equal(promoted, 0);
        publishBrowserTargetResult(destination, payload, candidate, { promote });
        assert.equal(promoted, 1);
    }
    finally {
        rmSync(directory, { recursive: true, force: true });
    }
});
