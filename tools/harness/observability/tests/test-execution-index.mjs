import assert from "node:assert/strict";
import test from "node:test";
import { mkdtempSync, rmSync, symlinkSync, readdirSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { readExecutionIndex } from "../execution-reader.mjs";
import { analyzeExecution } from "../execution-analysis.mjs";
import { executionRecordsDigest, executionObservation, resumeExecutionObservation } from "../../runtime/execution-observations.mjs";
import { atomicLocalFile, readLocalFile } from "../../runtime/secure-local-files.mjs";
import { createRuntime } from "../../runtime/session-client.mjs";
import { createLaunchContext } from "../../runtime/launch-context.mjs";
import { openExecutionChannel } from "../../test-support/execution-channel-fixture.mjs";

const digest = `sha256:${"a".repeat(64)}`;
const id = "00000000-0000-4000-8000-000000000001";
const launch = { kind: "launch", ref: `launch:${id}`, run_id: "execution", invocation_id: id, parent_invocation_id: null,
  unit_id: "u", command_id: null, row_ids: ["r"], producer: "go", launched: true, outcome: "passed" };
function fixture(t) {
  const base = mkdtempSync(path.join(os.tmpdir(), "cartulary-execution-")), root = path.join(base, "execution");
  t.after(() => rmSync(base, { recursive: true, force: true }));
  const manifest = { run_id: "execution", source_digest: digest, graph_digest: digest, instrumentation: { mode: "basic", policy_digest: digest } };
  const run = { manifest, registrations: new Map([["u", { row_ids: ["r"] }]]) };
  const index = { schema_id: "cartulary.harness_execution_index.v5", host_admission: { status: "complete", omitted_records: 0, records: [], records_digest: executionRecordsDigest([]) }, allocation_lifecycle: { status: "complete", omitted_records: 0, omitted_refs: [], records: [], records_digest: executionRecordsDigest([]) }, graph_clock: null, run_id: "execution", source_digest: digest, graph_digest: digest,
    policy_digest: digest, status: "complete", omitted_records: 0, omitted_refs: [], records: [structuredClone(launch)] };
  const file = path.join(root, "diagnostics/execution-index.json");
  const save = () => atomicLocalFile(file, JSON.stringify({ ...index, records_digest: executionRecordsDigest(index.records) }), { replace: true });
  save();
  return { root, run, index, file, save };
}

test("execution reader rejects forged identity, duplicate records, cycles and false completeness", (t) => {
  for (const mutate of [
    (v) => { v.run_id = "foreign"; }, (v) => { v.source_digest = `sha256:${"b".repeat(64)}`; },
    (v) => { v.graph_digest = `sha256:${"b".repeat(64)}`; }, (v) => { v.policy_digest = `sha256:${"b".repeat(64)}`; },
    (v) => v.records.push(v.records[0]), (v) => { v.records[0].parent_invocation_id = id; },
    (v) => { v.records[0].unit_id = "foreign"; }, (v) => { v.records[0].row_ids = ["foreign"]; },
    (v) => { v.records[0].launched = false; }, (v) => { v.records[0].outcome = "incomplete"; },
    (v) => { v.records[0].parent_invocation_id = "00000000-0000-4000-8000-000000000002"; },
  ]) {
    const f = fixture(t); mutate(f.index); f.save(); assert.throws(() => readExecutionIndex(f.root, f.run));
  }
});

test("explicit omissions preserve incomplete parentage without manufacturing a launch", (t) => {
  const f = fixture(t), parent = "00000000-0000-4000-8000-000000000002";
  f.index.records[0].parent_invocation_id = parent;
  f.index.status = "partial"; f.index.omitted_records = 1; f.index.omitted_refs = [`launch:${parent}`]; f.save();
  const index = readExecutionIndex(f.root, f.run);
  assert.equal(index.records.length, 1);
  assert.equal(analyzeExecution(index).completeness, "partial");
});

test("execution reader rejects digest corruption, unsafe files, duplicate JSON keys and byte overflow", (t) => {
  for (const transform of [
    (f) => atomicLocalFile(f.file, JSON.stringify({ ...f.index, records_digest: digest }), { replace: true }),
    (f) => atomicLocalFile(f.file, '{"schema_id":1,"schema_id":2}', { replace: true }),
    (f) => atomicLocalFile(f.file, " ".repeat(4 * 1024 ** 2 + 1), { replace: true }),
    (f) => { const target = path.join(f.root, "elsewhere"); atomicLocalFile(target, "{}"); rmSync(f.file); symlinkSync(target, f.file); },
  ]) { const f = fixture(t); transform(f); assert.throws(() => readExecutionIndex(f.root, f.run)); }
});

test("failed acquisition cannot invent a successful lease or measured unavailable activity", (t) => {
  const f = fixture(t);
  f.index.records.push({ kind: "activity", local_start_ms: null, local_end_ms: null, anchors: [], ref: "activity:a", activity: "migration", unit_id: "u", invocation_id: id, allocation_ref: null,
    outcome: "failed", clock: "unavailable", clock_identity: null, resolution_ms: null, duration_ms: 0, availability: "unavailable" });
  f.save(); assert.throws(() => readExecutionIndex(f.root, f.run));
  f.index.records.pop();
  f.index.records.push({ kind: "allocation", ref: "allocation:a", allocation_ref: "allocation:a", unit_id: "u", ownership: null, capability: "browser_stack", outcome: "failed" },
    { kind: "lease", ref: "lease:a", lease_ref: "a", allocation_ref: "allocation:a", unit_id: "u", ownership: "owned", capability: "browser_stack" });
  atomicLocalFile(path.join(f.root, "_shared/fixture-leases/a.json"), JSON.stringify({ schema_id: "cartulary.harness_fixture_lease.v5", lease_id: "a", capability: "browser_stack", postgres_isolation: "exclusive", ownership: "owned", state: "leased", cleanup_outcome: "not_required", cleanup_failure_reason: null, resource_ids: [], created_at: "2026-10-09T00:00:00.000Z" }));
  f.save(); assert.throws(() => readExecutionIndex(f.root, f.run), /successful allocation/u);
});

test("bounded presentation retains outcomes, unavailable durations and exact list omissions", (t) => {
  const f = fixture(t);
  f.index.records = Array.from({ length: 25 }, (_, i) => ({ kind: "activity", local_start_ms: null, local_end_ms: null, anchors: [], ref: `activity:${i}`, activity: "migration", unit_id: "u", invocation_id: id,
    allocation_ref: null, outcome: "incomplete", clock: "unavailable", clock_identity: null, resolution_ms: null, duration_ms: null, availability: "not_observed" }));
  f.index.status = "partial";
  const view = analyzeExecution(f.index);
  assert.equal(view.activities.records.length, 20); assert.equal(view.activities.omitted, 5);
  assert.equal(view.activities.records[0].duration_ms, null);
});

test("private beginnings obey the byte bound and survive invalid replacement", (t) => {
  const f = fixture(t), repoRoot = path.resolve(import.meta.dirname, "../../../..");
  rmSync(f.file);
  const runtime = createRuntime({ repoRoot, runRoot: f.root, runID: "execution" });
  t.after(() => runtime.close());
  const session = openExecutionChannel(runtime);
  const environment = { CARTULARY_TEST_RUN_ID: "execution", CARTULARY_TEST_RESULTS_DIR: path.dirname(f.root),
    CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: runtime.root, CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: runtime.leaseID,
    CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: "execution", ...session.environment };
  const options = { repoRoot, environment }, root = runtime.privatePath("execution-observations");
  const oversized = { ...launch, row_ids: Array.from({ length: 1000 }, (_, i) => `row-${i}-${"a".repeat(70)}`) };
  assert.equal(executionObservation(options, oversized).token, null);
  assert.equal(readdirSync(root).some((name) => name.startsWith("slot-")), false);
  const initial = { ...launch, launched: false, outcome: "incomplete" };
  const observation = executionObservation(options, initial);
  const file = path.join(root, observation.token.slot);
  assert.deepEqual(JSON.parse(readLocalFile(file)), initial);
  observation.update({ ...initial, outcome: "unknown" });
  assert.deepEqual(JSON.parse(readLocalFile(file)), initial);
  observation.update(oversized);
  assert.deepEqual(JSON.parse(readLocalFile(file)), initial);
  const resumed = resumeExecutionObservation(options, observation.token);
  resumed.update(oversized);
  assert.deepEqual(JSON.parse(readLocalFile(file)), initial);
  observation.update(launch, {});
  assert.deepEqual(JSON.parse(readLocalFile(file)), initial);
  session.close(); observation.update(launch); assert.deepEqual(session.records(), [initial]);
});

test("spawn proof survives a short launch privately without retroactive retained process details", (t) => {
  const f = fixture(t), repoRoot = path.resolve(import.meta.dirname, "../../../..");
  rmSync(f.file);
  const runtime = createRuntime({ repoRoot, runRoot: f.root, runID: "execution" });
  t.after(() => runtime.close());
  const session = openExecutionChannel(runtime);
  const environment = { CARTULARY_TEST_RUN_ID: "execution", CARTULARY_TEST_RESULTS_DIR: path.dirname(f.root),
    CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: runtime.root, CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: runtime.leaseID,
    CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: "execution", ...session.environment };
  const child = createLaunchContext({ repoRoot, environment, unitID: "u", producer: "go" });
  child.spawned(process.pid); child.settle("passed");
  session.close();
  const [value]=session.records();
  assert.equal(value.process_proof.pid,process.pid);
  assert.equal(value.process_proof.start,value.process_proof.stat.start);
  assert.equal(value.record.invocation_id,child.identity.invocation_id);
  assert.equal(value.record.outcome,"passed");
  assert.equal(JSON.stringify(value.record).includes('"process_proof"'),false);
});

test("allocation lifecycle rejects conflicts, chronology, missing coverage and unsafe borrows", (t) => {
  const lifecycle = () => ({ allocation_ref:"allocation:a", borrows:[{lease_ref:"lease-a",unit_id:"u"}],clock_identity:"fixture-broker-monotonic",
    acquisition:{start_ms:0,end_ms:1,duration_ms:1},retirement:{start_ms:3,end_ms:5,duration_ms:2},retirement_reason:"run_finalization",triggering_unit_id:null,settlement:"settled",coverage:"complete" });
  for (const reason of ["acquisition_failure","unhealthy_borrower","terminal_borrower_release","capacity_pressure","exclusive_admission","invocation_cancellation","run_finalization"]) {
    const f=fixture(t);const record=lifecycle();record.retirement_reason=reason;record.settlement="unresolved";
    f.index.allocation_lifecycle.records=[record];f.index.allocation_lifecycle.records_digest=executionRecordsDigest([record]);f.save();
    assert.equal(readExecutionIndex(f.root,f.run).allocation_lifecycle.records[0].retirement_reason,reason);
  }
  for (const mutate of [
    s=>s.records.push(s.records[0]), s=>{s.records[0].triggering_unit_id="foreign";},
    s=>{s.records[0].borrows[0].unit_id="foreign";},s=>{s.records[0].acquisition.duration_ms=9;},
    s=>{s.records[0].retirement.start_ms=0;s.records[0].retirement.duration_ms=5;},
    s=>{s.omitted_records=1;},s=>{s.records[0].retirement_reason="guess";},
  ]) {
    const f=fixture(t);f.index.allocation_lifecycle.records=[lifecycle()];mutate(f.index.allocation_lifecycle);
    f.index.allocation_lifecycle.records_digest=executionRecordsDigest(f.index.allocation_lifecycle.records);f.save();
    assert.throws(()=>readExecutionIndex(f.root,f.run));
  }
});

test("admission observations validate bounded facts, identities and canonical clocks",(t)=>{
 const f=fixture(t);
 f.run.phases=new Map([["host",{unit_id:"u",phase:"host_admission",start_ms:0,end_ms:20}]]);
 const phase={unit_id:"u",request_ref:null,clock_identity:"graph-monotonic",kind:"phase",phase:"admission_wait",start_ms:1,end_ms:10,duration_ms:4,observations:2};
 const save=(records,status="complete",omitted_records=0)=>{f.index.host_admission={records,status,omitted_records,records_digest:executionRecordsDigest(records)};f.save()};
 save([phase]);assert.equal(readExecutionIndex(f.root,f.run).host_admission.records[0].duration_ms,4);
 for(const change of [r=>r.unit_id="foreign",r=>r.end_ms=0,r=>r.duration_ms=30,r=>r.end_ms=40,r=>r.request_ref=`request:${"a".repeat(64)}`,r=>r.clock_identity="other",r=>r.private_path="/private/lease"]){const r=structuredClone(phase);change(r);save([r]);assert.throws(()=>readExecutionIndex(f.root,f.run));}
 save([phase,phase]);assert.throws(()=>readExecutionIndex(f.root,f.run),/overlapping/u);
 save([phase],"complete",1);assert.throws(()=>readExecutionIndex(f.root,f.run),/omissions/u);
 save([phase],"truncated",1);assert.equal(readExecutionIndex(f.root,f.run).host_admission.status,"truncated");
 const request_ref=`request:${"a".repeat(64)}`;
 const decision={...phase,kind:"decision",outcome:"committed",request_ref,duration_ms:0,decision:{request_ref,sequence:1,queue_position:1,mode:"shared",granted:false,requested_claims:{cpu:2},effective_capacities:{cpu:2},current_deficits:{cpu:1},persistent_deficits:{},required_ancestors:[],active_holders:[],residents:[],prospective:[],priority_blockers:[],exclusive_blockers:[]}};delete decision.phase;
 save([decision]);assert.equal(readExecutionIndex(f.root,f.run).host_admission.records.length,1);
 decision.decision.granted=true;save([decision]);assert.throws(()=>readExecutionIndex(f.root,f.run),/blockers/u);
 const ref=`resident:${"b".repeat(64)}`;
 const evaluation={...decision,kind:"retirement_evaluation",evaluation:{request_ref,resident_refs:[ref],before_deficits:{memory_mb:100},after_deficits:{memory_mb:80},active_references:[],eligible:true}};delete evaluation.decision;
 save([evaluation]);assert.equal(readExecutionIndex(f.root,f.run).host_admission.records[0].evaluation.eligible,true);
 for(const mutate of [r=>r.evaluation.eligible=false,r=>r.evaluation.after_deficits.process=1,r=>r.evaluation.resident_refs.push(ref),r=>r.evaluation.request_ref=`request:${"c".repeat(64)}`]) { const r=structuredClone(evaluation); mutate(r); save([r]); assert.throws(()=>readExecutionIndex(f.root,f.run)); }
 const noRelief=structuredClone(evaluation);noRelief.evaluation.after_deficits={memory_mb:100};noRelief.evaluation.eligible=false;save([noRelief]);assert.equal(readExecutionIndex(f.root,f.run).host_admission.records[0].evaluation.eligible,false);
 const retirement={...phase,kind:"retirement",duration_ms:9,observations:1,resident_refs:[],reason:"renderer_slot_pressure",outcome:"unresolved"};delete retirement.phase;
 save([retirement]);assert.equal(readExecutionIndex(f.root,f.run).host_admission.records[0].outcome,"unresolved");
});
