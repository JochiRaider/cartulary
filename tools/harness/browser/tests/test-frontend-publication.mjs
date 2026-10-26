import test from "node:test";
import assert from "node:assert/strict";
import { chmodSync, existsSync, lstatSync, linkSync, mkdtempSync, mkdirSync, readFileSync, rmSync, symlinkSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { createRuntime } from "../../runtime/session-client.mjs";
import { publishFrontendOutput, resolveFrontendArtifact, sealFrontendArtifact } from "../../readiness/frontend-artifact.mjs";
import { buildSourceSnapshot } from "../../test-catalog/index.mjs";
import { frontendPublicationEnvironment, startFrontendPublication } from "../frontend-publication.mjs";
const root = path.resolve(import.meta.dirname, "../../../..");
function frontendFixture(t) {
    const scratch = mkdtempSync(path.join(os.tmpdir(), "frontend-artifact-test-"));
    const runtimes = [];
    const publications = [];
    t.after(async () => {
        try {
            const stopped = await Promise.allSettled(publications.map(publication => publication.stop()));
            const errors = stopped.filter(result => result.status === "rejected").map(result => result.reason);
            if (errors.length)
                throw new AggregateError(errors, "publication test cleanup failed");
        }
        finally {
            for (const runtime of runtimes)
                runtime.close();
            rmSync(scratch, { recursive: true, force: true });
        }
    });
    const profile = { id: "production", producer_report_key: "build-web", producer_operation: "harness.build_web", entries: ["index.html"] };
    function fixture(runID, version) {
        const runRoot = path.join(scratch, runID);
        mkdirSync(runRoot, { mode: 0o700 });
        writeFileSync(path.join(runRoot, "run-manifest.json"), JSON.stringify({
            source_digest: `sha256:${"1".repeat(64)}`, toolchain_digest: `sha256:${"2".repeat(64)}`,
        }), { mode: 0o600 });
        const runtime = createRuntime({ repoRoot: root, runRoot, runID, scratchRoot: path.join(scratch, "private") });
        runtimes.push(runtime);
        const staging = runtime.privatePath("staging");
        mkdirSync(staging, { mode: 0o700 });
        writeFileSync(path.join(staging, "index.html"), `<script src="/lazy.js"></script>${version}`);
        writeFileSync(path.join(staging, "lazy.js"), `export default ${JSON.stringify(version)}`);
        const environment = {
            CARTULARY_TEST_RESULTS_DIR: scratch, CARTULARY_TEST_RUN_ID: runID,
            CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: runtime.root,
            CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: runtime.leaseID,
            CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: runID,
        };
        const directory = sealFrontendArtifact({ repoRoot: root, runRoot, runtime, profile, staging });
        return { directory, environment, runtime, runRoot };
    }
    function start(held, environment) {
        const publication = startFrontendPublication(held, { environment });
        publications.push(publication);
        return publication;
    }
    return { scratch, runtimes, profile, fixture, start };
}
test("independent frontend publication inherits capacity but clears invocation identity", () => {
    assert.deepEqual(frontendPublicationEnvironment({ PATH: "/tools", CARTULARY_HOST_ADMISSION_LEASE: "parent",
        CARTULARY_UNIT_CPU_TOKENS: "2", CARTULARY_HARNESS_CACHE_MODE: "off", CARTULARY_TEST_RUN_ID: "old-run",
        CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: "/private", CARTULARY_WORKSPACE_ADMISSION: "old",
        MAKEFLAGS: "old", MAKEOVERRIDES: "old", MFLAGS: "old", MAKELEVEL: "2" }), { PATH: "/tools", CARTULARY_HOST_ADMISSION_LEASE: "parent", CARTULARY_UNIT_CPU_TOKENS: "2", CARTULARY_HARNESS_CACHE_MODE: "off" });
});
test("frontend publication staging never enters executable source identity", () => {
    // Publishing another profile may overlap input validation. Neither its
    // staging files nor displaced output may enter the source snapshot.
    const sourceBefore = buildSourceSnapshot(root).digest;
    const publicationStage = mkdtempSync(path.join(root, "apps/web/.frontend-publish-"));
    const displacedOutput = `${publicationStage}-previous`;
    try {
        writeFileSync(path.join(publicationStage, "temporary.js.map"), "unpublished");
        mkdirSync(displacedOutput);
        writeFileSync(path.join(displacedOutput, "old.js"), "replaced");
        assert.equal(buildSourceSnapshot(root).digest, sourceBefore);
    }
    finally {
        rmSync(publicationStage, { recursive: true, force: true });
        rmSync(displacedOutput, { recursive: true, force: true });
    }
});
function publicationFixture(t, body) {
    const context = frontendFixture(t);
    const first = context.fixture("first", "original");
    const fakeBin = path.join(context.scratch, "bin");
    mkdirSync(fakeBin);
    writeFileSync(path.join(fakeBin, "make"), `#!${process.execPath}\n${body}\n`, { mode: 0o700 });
    return { ...context, first, environment: { ...process.env, ...first.environment, PATH: `${fakeBin}:${process.env.PATH}`,
            CARTULARY_HARNESS_RESOURCE_DEPENDENCIES: "[]" } };
}
test("frontend publication observes successful child completion", async (t) => {
    const { first, environment, start } = publicationFixture(t, "process.exitCode = 0;");
    const published = start(false, environment);
    const publishedRoot = await published.finished;
    assert.ok(publishedRoot.startsWith(path.join(first.runRoot, "frontend-publications")));
    await published.stop();
});
test("frontend publication cancellation settles one owned child exactly once", async (t) => {
    const { environment, start } = publicationFixture(t, "setTimeout(() => process.exit(0), 1000);");
    const cancelled = start(true, environment);
    await cancelled.stop();
    await assert.rejects(cancelled.finished, /Independent frontend Make build failed/u);
    assert.equal(cancelled.stop(), cancelled.stop(), "concurrent retirement shares one completion");
    await cancelled.stop();
});
test("frontend publication preserves failing child exit and stderr", async (t) => {
    const { environment, start } = publicationFixture(t, "process.stderr.write('build failed\\n'); process.exitCode = 7;");
    const failed = start(false, environment);
    await assert.rejects(failed.finished, /code=7.*\n.*build failed/su);
    await failed.stop();
});
test("frontend publication settles acquisition when the child never starts", async (t) => {
    const { scratch, environment, start } = publicationFixture(t, "process.exitCode = 0;");
    const absent = start(false, { ...environment, PATH: path.join(scratch, "absent") });
    await assert.rejects(absent.finished, /ENOENT/u);
    await absent.stop();
});
test("frontend publication retains admitted immutable assets and rejects tampered or foreign staging", (t) => {
    const { scratch, runtimes, profile, fixture } = frontendFixture(t);
    const first = fixture("first", "original");
    const second = fixture("second", "replacement");
    const destination = path.join(scratch, "published");
    const dangling = path.join(scratch, "dangling-publication");
    symlinkSync(path.join(scratch, "missing"), dangling);
    assert.throws(() => publishFrontendOutput(first.directory, dangling), /symlink/u);
    publishFrontendOutput(first.directory, destination);
    const before = resolveFrontendArtifact(root, "build-web", first.environment);
    publishFrontendOutput(second.directory, destination);
    for (let reload = 0; reload < 5; reload++) {
        assert.match(readFileSync(path.join(first.directory, "index.html"), "utf8"), /original/u);
        assert.match(readFileSync(path.join(first.directory, "lazy.js"), "utf8"), /original/u);
        assert.equal(resolveFrontendArtifact(root, "build-web", first.environment).receiptDigest, before.receiptDigest);
    }
    assert.match(readFileSync(path.join(destination, "index.html"), "utf8"), /replacement/u);
    assert.throws(() => resolveFrontendArtifact(root, "build-web-measurement", first.environment));
    const receiptFile = path.join(first.runRoot, before.receiptRef);
    const bytes = readFileSync(receiptFile);
    const bad = JSON.parse(bytes);
    bad.source_digest = `sha256:${"3".repeat(64)}`;
    writeFileSync(receiptFile, JSON.stringify(bad));
    assert.throws(() => resolveFrontendArtifact(root, "build-web", first.environment), /source_digest mismatch/u);
    writeFileSync(receiptFile, bytes);
    writeFileSync(path.join(first.directory, "lazy.js"), "tampered");
    assert.throws(() => resolveFrontendArtifact(root, "build-web", first.environment), /digest mismatch/u);
    const partial = first.runtime.privatePath("partial");
    mkdirSync(partial, { mode: 0o700 });
    assert.throws(() => sealFrontendArtifact({ repoRoot: root, runRoot: first.runRoot, runtime: first.runtime, profile, staging: partial }));
    symlinkSync(path.join(second.directory, "index.html"), path.join(partial, "index.html"));
    assert.throws(() => sealFrontendArtifact({ repoRoot: root, runRoot: first.runRoot, runtime: first.runtime, profile, staging: partial }), /invalid/u);
    rmSync(path.join(partial, "index.html"));
    chmodSync(path.join(second.directory, "index.html"), 0o644);
    linkSync(path.join(second.directory, "index.html"), path.join(partial, "index.html"));
    assert.throws(() => sealFrontendArtifact({ repoRoot: root, runRoot: first.runRoot, runtime: first.runtime, profile, staging: partial }), /share a file inode/u);
    assert.equal(lstatSync(path.join(second.directory, "index.html")).mode & 0o777, 0o644);
    chmodSync(path.join(second.directory, "index.html"), 0o600);
    rmSync(path.join(partial, "index.html"));
    // A valid foreign staging tree must fail before sealing or publication,
    // independently of missing entries or an already consumed producer claim.
    const foreign = path.join(scratch, "foreign-staging");
    mkdirSync(foreign, { mode: 0o700 });
    const foreignBytes = Buffer.from("foreign frontend");
    writeFileSync(path.join(foreign, "index.html"), foreignBytes, { mode: 0o600 });
    const foreignRun = path.join(scratch, "foreign-run");
    mkdirSync(foreignRun, { mode: 0o700 });
    writeFileSync(path.join(foreignRun, "run-manifest.json"), readFileSync(path.join(first.runRoot, "run-manifest.json")), { mode: 0o600 });
    const foreignRuntime = createRuntime({ repoRoot: root, runRoot: foreignRun, runID: "foreign-run", scratchRoot: path.join(scratch, "private") });
    runtimes.push(foreignRuntime);
    assert.throws(() => sealFrontendArtifact({ repoRoot: root, runRoot: foreignRun, runtime: foreignRuntime, profile, staging: foreign }));
    assert.deepEqual(readFileSync(path.join(foreign, "index.html")), foreignBytes);
    assert.equal(existsSync(path.join(foreignRun, "build-web/frontend-artifact.json")), false);
    assert.equal(existsSync(foreignRuntime.privatePath("frontend-production")), false);
    second.runtime.close();
    second.runtime.close();
    assert.throws(() => resolveFrontendArtifact(root, "build-web", second.environment));
});
