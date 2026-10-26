import assert from "node:assert/strict";
import fs from "node:fs";
import { syncBuiltinESMExports } from "node:module";
import { closeSync, linkSync, mkdirSync, mkdtempSync, openSync, renameSync, rmSync, symlinkSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { assertRetainedArtifactFile, readRetainedArtifactBytes } from "../index.mjs";
import test from "node:test";
const root = path.resolve(import.meta.dirname, "../../../..");
function artifactFiles(t) {
    const artifactReaderRoot = mkdtempSync(path.join(os.tmpdir(), "cartulary-evidence-reader-"));
    t.after(() => rmSync(artifactReaderRoot, { recursive: true, force: true }));
    const admitted = path.join(artifactReaderRoot, "admitted");
    const displaced = path.join(artifactReaderRoot, "displaced");
    const outside = path.join(artifactReaderRoot, "outside");
    mkdirSync(admitted);
    mkdirSync(outside);
    writeFileSync(path.join(admitted, "evidence.json"), "original");
    writeFileSync(path.join(outside, "evidence.json"), "redirected");
    const read = () => readRetainedArtifactBytes(artifactReaderRoot, "admitted/evidence.json", 8).toString();
    return { artifactReaderRoot, admitted, displaced, outside, read };
}
test("retained artifacts reject unsafe paths and aliases", (t) => {
    const { artifactReaderRoot, admitted, read } = artifactFiles(t);
    assert.equal(read(), "original");
    assertRetainedArtifactFile(artifactReaderRoot, "admitted/evidence.json");
    assert.throws(() => readRetainedArtifactBytes(artifactReaderRoot, "admitted/evidence.json", 7), /bounded/u);
    assert.throws(() => readRetainedArtifactBytes(artifactReaderRoot, "admitted", 8), /regular file/u);
    assert.throws(() => assertRetainedArtifactFile(artifactReaderRoot, "admitted"), /regular file/u);
    for (const relative of ["../outside", "admitted/../outside", "/absolute", "admitted//evidence.json", "admitted\\evidence.json"]) {
        assert.throws(() => readRetainedArtifactBytes(artifactReaderRoot, relative, 8), /normalized and relative/u);
    }
    symlinkSync(admitted, path.join(artifactReaderRoot, "alias"));
    symlinkSync(path.join(admitted, "evidence.json"), path.join(admitted, "alias.json"));
    for (const relative of ["alias/evidence.json", "admitted/alias.json"]) {
        assert.throws(() => readRetainedArtifactBytes(artifactReaderRoot, relative, 8), /ELOOP|ENOTDIR/u);
        assert.throws(() => assertRetainedArtifactFile(artifactReaderRoot, relative), /ELOOP|ENOTDIR/u);
    }
    assert.throws(() => readRetainedArtifactBytes(path.join(artifactReaderRoot, "alias"), "evidence.json", 8), /ELOOP|ENOTDIR/u);
    linkSync(path.join(admitted, "evidence.json"), path.join(admitted, "hardlink.json"));
    assert.throws(read, /unaliased/u);
    assert.throws(() => assertRetainedArtifactFile(artifactReaderRoot, "admitted/hardlink.json"), /unaliased/u);
    rmSync(path.join(admitted, "hardlink.json"));
});
test("descriptor-relative reads reject concurrent mutation and close descriptors after failure", (t) => {
    const { artifactReaderRoot, admitted, displaced, outside, read } = artifactFiles(t);
    const originalOpen = fs.openSync;
    const originalRead = fs.readSync;
    const originalClose = fs.closeSync;
    const opened = new Set();
    try {
        // No forbidden documentation component may even reach the filesystem.
        fs.openSync = () => { throw new Error("unexpected filesystem traversal"); };
        syncBuiltinESMExports();
        // These are hostile path tokens, not references to repository documents.
        const forbiddenDirectory = "docs";
        for (const relative of [`${forbiddenDirectory}/evidence.json`, "evidence.md", "nested/README", "nested/readme.txt"]) {
            assert.throws(() => readRetainedArtifactBytes(artifactReaderRoot, relative, 8), /documentation/u);
            assert.throws(() => assertRetainedArtifactFile(artifactReaderRoot, relative), /documentation/u);
        }
        assert.throws(() => readRetainedArtifactBytes(path.join(artifactReaderRoot, forbiddenDirectory), "evidence.json", 8), /documentation/u);
        fs.openSync = (file, flags) => {
            if (file.endsWith("/evidence.json")) {
                renameSync(admitted, displaced);
                symlinkSync(outside, admitted);
            }
            const descriptor = originalOpen(file, flags);
            opened.add(descriptor);
            return descriptor;
        };
        fs.closeSync = (descriptor) => { opened.delete(descriptor); return originalClose(descriptor); };
        syncBuiltinESMExports();
        assert.equal(read(), "original", "renaming an ancestor cannot redirect a descriptor-relative read");
        assert.equal(opened.size, 0);
        rmSync(admitted);
        renameSync(displaced, admitted);
        fs.openSync = (file, flags) => {
            const descriptor = originalOpen(file, flags);
            opened.add(descriptor);
            return descriptor;
        };
        let mutated = false;
        fs.readSync = (...args) => {
            if (!mutated) {
                mutated = true;
                writeFileSync(path.join(admitted, "evidence.json"), "changed");
            }
            return originalRead(...args);
        };
        syncBuiltinESMExports();
        assert.throws(read, /changed during read/u);
        assert.equal(opened.size, 0, "failed reads close every descriptor");
    }
    finally {
        fs.openSync = originalOpen;
        fs.readSync = originalRead;
        fs.closeSync = originalClose;
        syncBuiltinESMExports();
    }
});
