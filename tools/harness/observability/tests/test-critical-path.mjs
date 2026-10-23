import assert from "node:assert/strict";
import test from "node:test";
import { mkdtempSync, rmSync, readdirSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { criticalPathQuery, criticalPathModel, criticalPathExplanation, formatCriticalPath } from "../critical-path.mjs";
import { writeCanonicalFixture } from "./retained-fixture.mjs";

function pathFixture(count = 45) {
  const ids = Array.from({length: count}, (_, i) => `unit:${String(i).padStart(3,"0")}`);
  const run = { manifest: { run_id: "path", target: "ci", source_digest: `sha256:${"a".repeat(64)}`, graph_digest: `sha256:${"b".repeat(64)}` },
    targetSummaries: new Map([["ci", {target:"ci", unit_ids:ids}], ["internal", {target:"internal",unit_ids:ids.slice(0,5)}], ["empty",{target:"empty",unit_ids:[]}]]),
    registrations: new Map(ids.map((id,i) => [id,{unit_id:id,kind:"runner",needs:i ? [ids[i-1]]:[]}])),
    intervals: new Map(ids.map((id,i)=>[id,{start:i*20+5,end:i*20+15,queue_ms:5}])),
    waits: new Map(ids.map(id=>[id,{count:1,by_reason:{rank:2,resources:3,exclusive_barrier:0,scheduler_stop:0}}])), phases:new Map() };
  run.phases.set("p1",{unit_id:ids[0],phase:"host_admission",start_ms:5,end_ms:10,outcome:"passed"});
  run.phases.set("p2",{unit_id:ids[0],phase:"runner",start_ms:8,end_ms:12,outcome:"passed"});
  return {run,ids};
}
const page = (run,options={},target="ci") => JSON.parse(formatCriticalPath(criticalPathModel(run,null,target,criticalPathQuery(options)).model,{json:true}));
test("critical path pagination reconstructs long paths and preserves complete totals",()=>{
 const {run,ids}=pathFixture(); let offset=0, observed=[];
 do { const p=page(run,{offset}); observed.push(...p.records.map(r=>r.unit_id)); assert.deepEqual(p.path_totals,{unit_count:45,queue_ms:225,lifetime_ms:450,host_admission_ms:5,host_phase_ms:{logical_slot:0,admission_transaction:0,admission_wait:0,pressure_retirement:0},host_uncovered_ms:5}); offset=p.next_offset; } while(offset!==null);
 assert.deepEqual(observed,ids);
 const filtered=page(run,{unitID:ids[0]}); assert.equal(filtered.total_count,1);assert.equal(filtered.path_totals.unit_count,45);assert.equal(filtered.filtered_totals.unit_count,1);
 assert.equal(filtered.records[0].phase_union_ms,7);assert.equal(filtered.records[0].phase_overlap_ms,2);assert.equal(filtered.records[0].uncovered_lifetime_ms,3);assert.equal(filtered.records[0].queue_ms,5);
 assert.deepEqual(page(run,{offset:999}).records,[]);assert.equal(page(run,{offset:999}).next_offset,null);
 assert.equal(page(run,{},"internal").path_totals.unit_count,5);assert.equal(page(run,{},"empty").total_count,0);
 assert.throws(()=>page(run,{unitID:"absent"}),{exit_code:2});assert.throws(()=>page(run,{},"absent"),{exit_code:2});
});
test("critical path rejects invalid selectors and shortens only at record boundaries",()=>{
 for(const q of [{view:"other"},{offset:-1},{offset:1.1},{offset:"1x"},{offset:"01"},{offset:Number.MAX_SAFE_INTEGER+1},{limit:0},{limit:101},{unitID:""}]) assert.throws(()=>criticalPathQuery(q),{exit_code:2});
 const {run}=pathFixture();const model=criticalPathModel(run,null,"ci",criticalPathQuery()).model;
 const p=JSON.parse(formatCriticalPath(model,{json:true,maximumBytes:2600}));assert.ok(p.returned_count>0 && p.returned_count<20);assert.equal(p.next_offset,p.returned_count);assert.equal(p.total_count,45);
 assert.throws(()=>formatCriticalPath(model,{json:true,maximumBytes:100}),/exceeds/u);
 const text=formatCriticalPath(model);assert.ok(Buffer.byteLength(text)<=8000);assert.match(text,/next_offset=/u);
});
test("critical path ties remain deterministic and allocations follow path ordering",()=>{
 const {run,ids}=pathFixture(3);run.registrations.get(ids[2]).needs=[];
 const tied=pathFixture(2);tied.run.registrations.get(tied.ids[1]).needs=[];assert.deepEqual(page(tied.run).records.map(r=>r.unit_id),[tied.ids[0]]);
 const initial=page(run);run.registrations=new Map([...run.registrations].reverse());assert.deepEqual(page(run),initial);
 const lifecycle=(allocation_ref,unit)=>({allocation_ref,borrows:[{lease_ref:allocation_ref,unit_id:unit}],triggering_unit_id:null});
 const execution={allocation_lifecycle:{status:"truncated",omitted_records:2,records:[lifecycle("allocation:z",ids[1]),lifecycle("allocation:a",ids[0])]}};
 const model=criticalPathModel(run,execution,"ci",criticalPathQuery({view:"allocations"})).model;
 assert.deepEqual(model.records.map(r=>r.lifecycle.allocation_ref),["allocation:a","allocation:z"]);assert.equal(model.coverage.allocation_omissions,2);
});
test("canonical wait pages reconcile exact reasons and remain read-only",async(t)=>{
 const dir=mkdtempSync(path.join(os.tmpdir(),"cartulary-path-"));t.after(()=>rmSync(dir,{recursive:true,force:true}));writeCanonicalFixture(dir);
 const before=readdirSync(dir,{recursive:true});
 const result=JSON.parse(formatCriticalPath(await criticalPathExplanation(dir,{view:"waits"}),{json:true}));
 assert.equal(result.total_count,1);assert.equal(result.records[0].duration_ms,2);assert.equal(result.records[0].wait_reason,"rank");assert.deepEqual(result.records[0].blocking_unit_ids,[]);
 assert.equal(result.path_totals.queue_ms,2);assert.deepEqual(readdirSync(dir,{recursive:true}),before);
 rmSync(path.join(dir,"diagnostics/execution-index.json"));
 const unavailable=await criticalPathExplanation(dir,{view:"allocations"});assert.equal(unavailable.coverage.allocation_lifecycle,"unavailable");assert.equal(unavailable.total_count,0);
});

test("critical path rejects nonclosing queue partitions",()=>{
 const {run,ids}=pathFixture(1);run.waits.get(ids[0]).by_reason.resources=4;
 for(const view of ["units","waits","allocations"]) assert.throws(()=>page(run,{view}),/queue reason partition/u);
});

test("host admission pages retain complete phase totals without turning snapshots into intervals",()=>{
 const {run,ids}=pathFixture();
 const records=ids.map((unit_id,i)=>({unit_id,request_ref:null,clock_identity:"graph-monotonic",kind:"phase",phase:"admission_wait",start_ms:i*20+5,end_ms:i*20+10,duration_ms:2,observations:2}));
 for(const r of records) run.phases.set(`host:${r.unit_id}`,{unit_id:r.unit_id,phase:"host_admission",start_ms:r.start_ms,end_ms:r.end_ms,outcome:"passed"});
 run.phases.delete("p1");
 const execution={host_admission:{status:"partial",omitted_records:3,records}};
 let offset=0,seen=[];
 do { const p=JSON.parse(formatCriticalPath(criticalPathModel(run,execution,"ci",criticalPathQuery({view:"host-admission",offset})).model,{json:true}));
   seen.push(...p.records.map(r=>r.unit_id));assert.equal(p.coverage.host_admission,"partial");assert.equal(p.coverage.host_admission_omissions,3);assert.equal(p.path_totals.host_phase_ms.admission_wait,90);offset=p.next_offset;
 } while(offset!==null);
 assert.deepEqual(seen,ids);
 const filtered=criticalPathModel(run,execution,"ci",criticalPathQuery({view:"host-admission",unitID:ids[0]})).model;
 assert.equal(filtered.total_count,1);assert.equal(filtered.path_totals.host_phase_ms.admission_wait,90);assert.equal(filtered.filtered_totals.host_phase_ms.admission_wait,2);
 assert.equal(page(run,{view:"host-admission"}).coverage.host_admission,"unavailable");
 assert.deepEqual(page(run,{view:"host-admission"}).records,[]);
});
