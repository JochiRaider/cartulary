import assert from "node:assert/strict";
import fs from "node:fs";
import { syncBuiltinESMExports } from "node:module";
import { linkSync, mkdirSync, mkdtempSync, readFileSync, rmSync, symlinkSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { loadPerformanceFixtureSnapshotRegistry } from "../index.mjs";
import test from "node:test";
const root = path.resolve(import.meta.dirname, "../../../..");
test("fixture registry rejects forbidden inputs before probing and rejects physical aliases", async (t) => {
    const registryBoundaryRoot = mkdtempSync(path.join(os.tmpdir(), "cartulary-registry-boundary-"));
    try {
        const registry = JSON.parse(readFileSync(path.join(root, "tools/performance_fixture_snapshot_owner.json"), "utf8"));
        const registryPath = path.join(registryBoundaryRoot, "registry.json");
        function selectSource(relative) {
            const value = structuredClone(registry);
            value.profiles[0].source_contract_refs = [{ ...value.profiles[0].source_contract_refs[0], path: relative }];
            writeFileSync(registryPath, JSON.stringify(value));
        }
        const forbiddenDirectory = "docs";
        const forbidden = [`${forbiddenDirectory}/source.json`, `contracts/${forbiddenDirectory}/source.json`, "contracts/source.MD", "contracts/README"];
        const probes = ["openSync", "statSync", "lstatSync", "realpathSync", "readdirSync"];
        const originals = new Map(probes.map(name => [name, fs[name]]));
        let forbiddenProbes = 0;
        try {
            for (const name of probes)
                fs[name] = (file, ...args) => {
                    const token = String(file);
                    if (token.split(path.sep).includes(forbiddenDirectory) || /(?:\.(?:md|markdown)$|(?:^|\/)readme(?:\.|$))/iu.test(token)) {
                        forbiddenProbes++;
                        throw new Error("unexpected forbidden filesystem probe");
                    }
                    return originals.get(name)(file, ...args);
                };
            syncBuiltinESMExports();
            for (const relative of forbidden) {
                selectSource(relative);
                assert.throws(() => loadPerformanceFixtureSnapshotRegistry(root, { registryPath }), /under contracts|documentation/u);
            }
            assert.equal(forbiddenProbes, 0);
        }
        finally {
            for (const [name, original] of originals)
                fs[name] = original;
            syncBuiltinESMExports();
        }
        mkdirSync(path.join(registryBoundaryRoot, "contracts"));
        writeFileSync(path.join(registryBoundaryRoot, "source.json"), "{}");
        symlinkSync("../source.json", path.join(registryBoundaryRoot, "contracts/alias.json"));
        selectSource("contracts/alias.json");
        assert.throws(() => loadPerformanceFixtureSnapshotRegistry(registryBoundaryRoot, { registryPath }), /ELOOP/u);
        symlinkSync(registryBoundaryRoot, path.join(registryBoundaryRoot, "contracts/alias"));
        selectSource("contracts/alias/source.json");
        assert.throws(() => loadPerformanceFixtureSnapshotRegistry(registryBoundaryRoot, { registryPath }), /ENOTDIR|ELOOP/u);
        linkSync(path.join(registryBoundaryRoot, "source.json"), path.join(registryBoundaryRoot, "contracts/hardlink.json"));
        selectSource("contracts/hardlink.json");
        assert.throws(() => loadPerformanceFixtureSnapshotRegistry(registryBoundaryRoot, { registryPath }), /unaliased/u);
    }
    finally {
        rmSync(registryBoundaryRoot, { recursive: true, force: true });
    }
});
