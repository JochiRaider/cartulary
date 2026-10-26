import assert from "node:assert/strict";
import path from "node:path";
import { WorkGraphCompiler } from "../index.mjs";
import test from "node:test";
import { planBrowserLanes } from "../browser-lanes.mjs";
const root = path.resolve(import.meta.dirname, "../../../../..");
test("browser suites project all groups and retained visual finalizers", async (t) => {
    const compiler = new WorkGraphCompiler(root);
    for (const target of [
        "browser-e2e-webserver-backed",
        "browser-e2e-stateful",
        "browser-e2e-measurement",
        "browser-e2e-a11y",
        "browser-e2e-visual",
    ]) {
        const suite = ({ "browser-e2e-webserver-backed": "functional", "browser-e2e-a11y": "accessibility" })[target] ?? target.slice("browser-e2e-".length);
        const graph = compiler.compileCommand("browser-e2e", { SUITE: suite }).graph;
        assert.ok(graph.units.some((unit) => unit.unit_id.startsWith("browser_group:")), `${target} must expose browser groups as units`);
        assert.ok(graph.units.some((unit) => unit.unit_id.startsWith("browser_target_summary:")), `${target} must expose its summary projection as a unit`);
        assert.equal(graph.units.some((unit) => unit.command.args.some((arg) => arg.includes("run-browser-e2e-batch"))), false);
    }
    const visualGraph = compiler.compile({
        kind: "target",
        target: "harness.suite_browser_visual",
    });
    const visualTargetFinalizer = visualGraph.units.find((unit) => unit.unit_id === "browser_target_summary:browser-e2e-visual");
    assert.ok(visualTargetFinalizer, "visual target must expose its finalizer");
    assert.ok(visualTargetFinalizer.current_run_evidence_outputs.includes("browser-e2e-visual/frontend-visual-reconciliation.json"), "visual target finalizer must retain the reconciliation artifact");
    for (const unit of visualGraph.units.filter((entry) => entry.unit_id.startsWith("browser_group:visual:"))) {
        assert.ok(unit.current_run_evidence_outputs.some((output) => output.endsWith("/browser-group-result.json")), `${unit.unit_id} must retain its logical borrow and physical attestation reference`);
    }
    const visualUpdateGraph = compiler.compile({
        kind: "target",
        target: "browser-e2e-visual-update",
    });
    const visualUpdateFinalizer = visualUpdateGraph.units.find((unit) => unit.unit_id === "browser_target_summary:browser-e2e-visual-update");
    assert.ok(visualUpdateFinalizer?.current_run_evidence_outputs.includes("browser-e2e-visual-update/frontend-visual-reconciliation.json"), "visual update must reconcile its complete scratch candidate before promotion");
});
test("ordinary browser selection owns cohorts lanes preparation and accounting", async (t) => {
    const compiler = new WorkGraphCompiler(root);
    // partition cohorts and physical affinity. Reset is consuming preparation.
    const functionalRows = compiler.catalog.rows.filter((row) => row.runner === "playwright" && row.selector.stage === "webserver_backed");
    const statefulRows = compiler.catalog.rows.filter((row) => row.runner === "playwright" && row.selector.stage === "stateful" && row.aggregate_placement !== "explicit_only");
    for (const rows of [functionalRows, statefulRows, [...functionalRows, ...statefulRows]]) {
        const selected = rows.map((row) => row.row_id).sort();
        const graph = compiler.compile({ kind: "rows", row_ids: selected });
        const groups = graph.units.filter((unit) => unit.unit_id.startsWith("browser_group:"));
        assert.deepEqual(groups.flatMap((unit) => unit.row_bindings.map(({ row_id }) => row_id)).sort(), selected);
        assert.equal(graph.units.some((unit) => unit.unit_id.startsWith("browser_reset:") || unit.unit_id.startsWith("browser_lifecycle:")), false);
        assert.ok(new Set(groups.map((unit) => unit.command.environment.CARTULARY_BROWSER_LANE_ID)).size <= 4);
        for (const unit of groups) {
            const environment = unit.command.environment;
            assert.equal(unit.owner_id, "harness.browser");
            assert.equal(environment.CARTULARY_BROWSER_RESOURCE_PROFILE_ID, "browser_ordinary");
            assert.match(environment.CARTULARY_BROWSER_LANE_ID, /^browser-[1-4]$/u);
            assert.match(environment.CARTULARY_BROWSER_PREPARATION_LABEL, /^cohort-/u);
            assert.equal(environment.CARTULARY_BROWSER_GROUP_GENERATION, undefined);
            assert.equal(environment.CARTULARY_BROWSER_FUNCTIONAL_LANE_ID, undefined);
            assert.equal(unit.resource_claims.postgres, 2);
            assert.equal(unit.resource_claims.object_store, 1);
            assert.deepEqual(unit.service_dependencies, ["object_store", "postgres"]);
            assert.deepEqual(unit.physical_allocations.browser_stack, compiler.topology.physical_resource_claims.browser_stack);
            assert.deepEqual(unit.physical_allocations.renderer, compiler.topology.physical_resource_claims.renderer);
            assert.equal(unit.renderer_requirement, "pooled");
            assert.equal(unit.failure_policy.block_descendants, false);
            const selectedRows = unit.row_bindings.map(({ row_id }) => compiler.catalog.rowByID.get(row_id));
            assert.equal(new Set(selectedRows.map((row) => row.runtime_profile_id)).size, 1);
            if (selectedRows.length > 1)
                assert.ok(selectedRows.every((row) => row.browser_state_scope === "test_owned"));
        }
        for (const [affinity, members] of Map.groupBy(groups, (unit) => unit.affinity_key)) {
            assert.equal(new Set(members.map((unit) => unit.command.environment.CARTULARY_BROWSER_RUNTIME_PROFILE_ID)).size, 1, affinity);
            const bodies = members.reduce((total, unit) => total + unit.row_bindings.reduce((cost, { row_id }) => {
                const row = compiler.catalog.rowByID.get(row_id);
                return cost + row.selector.titles.length * compiler.owner.evidence_estimates_ms[row.evidence_class];
            }, 0), 0);
            const policy = compiler.topology.browser_e2e_batch.invocation_policy;
            assert.equal(members.reduce((total, unit) => total + unit.estimated_work_ms, 0), bodies +
                members.length * (policy.startup_ms + policy.cleanup_ms) + compiler.owner.lifecycle_estimates_ms.browser_allocation_acquire +
                (members.length - 1) * compiler.owner.lifecycle_estimates_ms.browser_reset + compiler.owner.lifecycle_estimates_ms.browser_allocation_release);
        }
    }
    const mixed = compiler.compile({ kind: "rows", row_ids: [...functionalRows, ...statefulRows].map((row) => row.row_id).sort() });
    assert.ok(mixed.units.some((unit) => new Set((unit.row_bindings ?? []).map(({ row_id }) => compiler.catalog.rowByID.get(row_id).selector.stage)).size > 1), "compatible selected rows must pack across logical reporting stages");
    for (const target of ["browser-e2e-webserver-backed", "browser-e2e-stateful"]) {
        const finalizer = mixed.units.find((unit) => unit.unit_id === `browser_target_summary:${target}`);
        assert.ok(finalizer);
        assert.equal(finalizer.command.args.includes("--reset-prefix"), false);
        assert.ok(finalizer.command.args.includes("--group-units"));
        assert.ok(finalizer.needs.every((id) => mixed.units.some((unit) => unit.unit_id === id)));
    }
    const one = compiler.compile({ kind: "rows", row_ids: [functionalRows[0].row_id] });
    assert.equal(one.units.filter((unit) => unit.row_bindings?.length).length, 1);
    const isolatedID = "module.networkflow.browser_stateful.pagination_expiry_bridge";
    const claimedIDs = [isolatedID, "module.extensions.browser_stateful.bc015_availability_continuity_d538000c38",
        "module.networkflow.browser_stateful.pagination_recovery", "module.networkflow.browser_stateful.saved_graph_exact_result_lifecycle",
        "module.networkflow.browser_stateful.verify_protected_network_analysis_state_is_disca_21a5de1ebf"].sort();
    const claimed = compiler.compile({ kind: "rows", row_ids: claimedIDs });
    const claimedGroups = claimed.units.filter((unit) => unit.row_bindings?.length);
    assert.deepEqual(claimedGroups.flatMap((unit) => unit.row_bindings.map(({ row_id }) => row_id)).sort(), claimedIDs);
    assert.ok(claimedGroups.some((unit) => unit.row_bindings.length === 1 && unit.row_bindings[0].row_id === isolatedID));
    assert.equal(new Set(claimedGroups.map((unit) => unit.affinity_key)).size, claimedGroups.length);
    const syntheticGroups = Array.from({ length: 9 }, (_, i) => ({ name: `group-${i}`, runtimeProfileID: `incompatible-${i}`, estimate: 20 - i }));
    const plan = (groups) => planBrowserLanes(groups, { maxLanes: 2, estimateGroup: (group) => group.estimate });
    assert.deepEqual(plan(syntheticGroups), plan([...syntheticGroups].reverse()));
    assert.equal(plan(syntheticGroups).length, 2, "incompatible partitions queue within the lane limit");
    assert.equal(new Set(plan(syntheticGroups).flatMap((lane) => lane.groups.map((item) => item.group.name))).size, 9);
});
test("measurement selections bind one builder and quiet consumers with post-cleanup finalizers", async (t) => {
    const compiler = new WorkGraphCompiler(root);
    const measurement = compiler.compile({
        kind: "target",
        target: "harness.suite_browser_measurement",
    });
    const snapshotBuilders = measurement.units.filter((entry) => entry.unit_id.startsWith("fixture_snapshot:default:timeline_large_grid_snapshot_v2:"));
    assert.equal(snapshotBuilders.length, 1, "measurement rows sharing the large-grid profile use one builder");
    assert.equal(snapshotBuilders[0].kind, "fixture_builder");
    assert.equal(snapshotBuilders[0].fixture_lease, "postgres_dedicated");
    assert.deepEqual(snapshotBuilders[0].service_dependencies, ["postgres"]);
    assert.deepEqual(snapshotBuilders[0].resource_claims, {
        cpu: 1,
        io: 1,
        memory_mb: 512,
        postgres: 1,
        process: 1,
    });
    assert.equal(snapshotBuilders[0].command.environment.CARTULARY_FIXTURE_BUILDER_RESOURCE_PROFILE_ID, "performance_fixture_builder");
    for (const forbiddenResource of ["browser_stack", "object_store", "port_lane"]) {
        assert.equal(Object.hasOwn(snapshotBuilders[0].resource_claims, forbiddenResource), false, `fixture builder must not claim ${forbiddenResource}`);
    }
    assert.deepEqual(snapshotBuilders[0].shared_locks, ["host_activity"]);
    assert.deepEqual(snapshotBuilders[0].exclusive_locks, []);
    assert.match(snapshotBuilders[0].snapshot_key, /^[a-f0-9]{64}$/u);
    assert.deepEqual(snapshotBuilders[0].current_run_evidence_outputs, [
        `performance-fixtures/${snapshotBuilders[0].snapshot_key}/build-diagnostics.json`,
        `performance-fixtures/${snapshotBuilders[0].snapshot_key}/snapshot-build.json`,
        `unit-results/${snapshotBuilders[0].unit_id.replaceAll(/[^A-Za-z0-9_.-]+/gu, "-")}.json`,
    ]);
    const activeFixtureProfiles = [...compiler.catalog.fixtureProfiles.profiles.values()].filter(profile => profile.status === "active");
    assert.deepEqual(measurement.units.filter(unit => unit.kind === "fixture_builder").map(unit => unit.fixture_profile_id).sort(), activeFixtureProfiles.map(profile => profile.fixture_profile_id).sort());
    for (const profile of activeFixtureProfiles) {
        const profiledRowIDs = compiler.catalog.rows.filter(row => row.fixture_profile_id === profile.fixture_profile_id).map(row => row.row_id).sort();
        assert.ok(profiledRowIDs.length > 0);
        const fixtureIdentity = graph => {
            const builders = graph.units.filter(unit => unit.kind === "fixture_builder" && unit.fixture_profile_id === profile.fixture_profile_id);
            assert.equal(builders.length, 1, "profiled route must resolve exactly one builder for each selected profile");
            return { builder_unit_id: builders[0].unit_id, fixture_profile_id: builders[0].fixture_profile_id, snapshot_key: builders[0].snapshot_key };
        };
        const expectedFixtureIdentity = fixtureIdentity(measurement);
        const owners = [...new Set(compiler.catalog.rows.filter(row => row.fixture_profile_id === profile.fixture_profile_id).map(row => row.owner_id))];
        for (const graph of [
            compiler.compile({ kind: "rows", row_ids: [profiledRowIDs[0]] }),
            compiler.compile({ kind: "rows", row_ids: profiledRowIDs }),
            ...owners.map(owner_id => compiler.compile({ kind: "owner", owner_id })),
            compiler.compile({ kind: "aggregate", target: "release-check" }),
        ])
            assert.deepEqual(fixtureIdentity(graph), expectedFixtureIdentity, "direct, row, owner and release routes must preserve each fixture identity");
    }
    const measurementSummaries = measurement.units.filter((entry) => entry.unit_id.startsWith("browser_measurement_summary:measurement:"));
    assert.equal(measurementSummaries.length, activeFixtureProfiles.reduce((count, profile) => count + profile.verification_bindings.length, 0), "each admitted measurement row has one post-cleanup summary finalizer");
    for (const summary of measurementSummaries) {
        assert.equal(summary.kind, "finalizer");
        assert.equal(summary.fixture_lease, "none");
        assert.ok(Object.keys(summary.resource_claims).length > 0);
        assert.deepEqual(summary.shared_locks, ["host_activity"]);
        assert.deepEqual(summary.exclusive_locks, []);
        assert.equal(summary.needs.length, 1);
        assert.ok(summary.needs[0].startsWith("browser_group:measurement:"));
        assert.match(summary.current_run_evidence_outputs[0], /frontend-measurement-summary\.v5\.json$/u);
    }
    const measurementTargetFinalizer = measurement.units.find((entry) => entry.unit_id === "browser_target_summary:browser-e2e-measurement");
    assert.ok(measurementTargetFinalizer);
    for (const summary of measurementSummaries) {
        assert.ok(measurementTargetFinalizer.needs.includes(summary.unit_id), "the aggregate must wait for every row finalizer");
    }
    const timelineGridConsumers = measurement.units.filter((entry) => entry.unit_id.startsWith("browser_group:measurement:") &&
        entry.row_bindings.some((binding) => compiler.catalog.rowByID.get(binding.row_id).selector.file === "apps/web/e2e/measurement/timeline-grid.spec.ts"));
    assert.equal(timelineGridConsumers.length, 4, "all four Timeline grid measurements retain dedicated consumers");
    for (const firstConsumer of timelineGridConsumers) {
        assert.ok(firstConsumer.needs.includes(snapshotBuilders[0].unit_id), firstConsumer.unit_id + " must depend on the shared snapshot builder");
        assert.equal(firstConsumer.fixture_profile_id, "timeline_large_grid_snapshot_v2");
        assert.equal(firstConsumer.snapshot_key, snapshotBuilders[0].snapshot_key);
        assert.match(firstConsumer.command.environment.CARTULARY_FIXTURE_ROW_ID, /^module\.timeline\.measurement\./u);
        assert.match(firstConsumer.command.environment.CARTULARY_FIXTURE_PREDICATE_ID, /^perf\./u);
    }
    for (const unit of measurement.units) {
        if (Object.keys(unit.resource_claims).length === 0)
            continue;
        assert.equal(unit.shared_locks.includes("host_activity") +
            unit.exclusive_locks.includes("host_activity"), 1, `${unit.unit_id} must declare exactly one host_activity mode`);
    }
    for (const unit of measurement.units.filter((entry) => entry.unit_id.startsWith("browser_group:measurement:") ||
        entry.unit_id === "browser_target_summary:browser-e2e-measurement")) {
        assert.ok(unit.exclusive_locks.includes("host_activity"), `${unit.unit_id} must hold the quiet profile exclusively`);
    }
    for (const unit of measurement.units.filter((entry) => entry.unit_id.startsWith("browser_group:measurement:"))) {
        assert.equal(unit.command.environment.CARTULARY_BROWSER_RESOURCE_PROFILE_ID, "browser_measurement_quiet");
        assert.equal(unit.timeout_ms, 3600000);
        assert.equal(unit.failure_policy.block_descendants, false, `${unit.unit_id} must allow the evidence finalizer to run after failure`);
    }
    const functional = compiler.compile({
        kind: "target",
        target: "harness.suite_browser_functional",
    });
    for (const unit of functional.units.filter((entry) => entry.unit_id.startsWith("browser_group:"))) {
        assert.ok(unit.shared_locks.includes("host_activity"));
        assert.equal(unit.command.environment.CARTULARY_BROWSER_RESOURCE_PROFILE_ID, "browser_ordinary");
    }
});
test("workbook owner selects browser groups without duplicate row runners", async (t) => {
    const compiler = new WorkGraphCompiler(root);
    const workbookOwner = compiler.compile({
        kind: "owner",
        owner_id: "module.workbook",
    });
    const selectedWorkbookBrowserRows = compiler.catalog.rows
        .filter((row) => row.owner_id === "module.workbook" && row.runner === "playwright" && row.aggregate_placement !== "explicit_only")
        .map((row) => row.row_id)
        .sort();
    const projectedWorkbookBrowserRows = workbookOwner.units
        .filter((unit) => unit.unit_id.startsWith("browser_group:"))
        .flatMap((unit) => unit.current_run_evidence_outputs
        .filter((output) => output.startsWith("rows/"))
        .map((output) => output.slice("rows/".length, -".json".length)))
        .sort();
    assert.deepEqual(projectedWorkbookBrowserRows, selectedWorkbookBrowserRows, "owner selections must project Playwright rows through browser group units");
    assert.equal(workbookOwner.units.some((unit) => unit.unit_id.startsWith("row:") &&
        selectedWorkbookBrowserRows.includes(unit.unit_id.slice("row:".length))), false, "owner selections must not schedule Playwright row runners directly");
});
for (const suite of ["contract", "regression", "full"]) {
    test(`harness ${suite} selects each browser assurance operation exactly once`, () => {
        const compiler = new WorkGraphCompiler(root);
        const graph = compiler.compileCommand("harness-contract", { SUITE: suite }).graph;
        const observed = graph.units.filter(unit => ["row:harness.browser.behavior.artifact_assurance", "row:harness.test_catalog.behavior.routing_contract"].includes(unit.unit_id)).map(unit => unit.unit_id).sort();
        assert.deepEqual(observed, ["row:harness.browser.behavior.artifact_assurance", "row:harness.test_catalog.behavior.routing_contract"]);
    });
}
