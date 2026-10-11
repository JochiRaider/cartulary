import assert from "node:assert/strict";
import test from "node:test";
import { publishMeasurementFixture } from "./measurement-fixture.mjs";
import { approvedPlan, treatmentOrder, validatePlan } from "../qualification/policy.mjs";
import { compareCell, assessWorkload, mean, median, mad, p90, intervalUnion, eligible } from "../qualification/analysis.mjs";
import { assembleMeasurement, summarizeMemory, finalizationIntervals, requireBasicTreatment } from "../qualification/measurement.mjs";
const plan=approvedPlan();
const attempt=(block,overrides={})=>({block,phase:"measured",mode:"off",readers:0,exit_code:0,failure:null,measurement:{outer_ms:10000,canonical_ms:8000,workload_cpu_us:1000000,reader_cpu_us:0,steady_rss_bytes:1000000,sampled_peak_rss_bytes:3000000,cpu_error_us:1,rss_error_bytes:4096,outer_error_ms:0,memory_samples:80,maximum_memory_gap_ms:102,finalization_intervals:[],coverage:{waited_descendants:true,reader_lifecycle:true,memory_samples:true,bounded_error:true,exclusive_quietness:true,complete_finalization:true},...overrides}});

test("qualification arithmetic uses a signed CPU mean, diagnostic MAD, elapsed percentiles and union bounds",()=>{
 assert.equal(mean([3,1,2,9]),3.75);assert.throws(()=>mean([]),/empty/u);
 assert.equal(median([3,1,2,9]),2.5);assert.equal(mad([1,2,3,4,5]),1);assert.equal(p90([4,1,3,2,5]),5);
 assert.equal(intervalUnion([{start_ms:10,end_ms:40},{start_ms:30,end_ms:50},{start_ms:70,end_ms:90}]),60);
 const reference=Array.from({length:5},(_,i)=>attempt(i+1));
 const candidate=[-10000,0,10000,20000,30000].map((cpu,i)=>attempt(i+1,{workload_cpu_us:1000000+cpu,steady_rss_bytes:900000}));
 const result=compareCell(reference,candidate,plan.budgets);
 assert.equal(result.stats.cpu_mean_percent_one_core,0.1);
 assert.equal(result.stats.cpu_median_percent_one_core,0.1);
 assert.equal(result.stats.cpu_mad_percent_one_core,0.1);
 assert.equal(result.stats.cpu_error_bound_percent_one_core,0.00002);
 assert.equal(result.stats.cpu_upper_percent_one_core,0.40002000000000004);
 assert.equal(result.stats.rss_upper_bytes,-91808);assert.ok(result.passed);assert.equal(result.signed_pairs[0].cpu,-0.1);
 candidate[0].measurement.reader_cpu_us=1500000;assert.equal(compareCell(reference,candidate,plan.budgets).passed,false);
});
test("CPU uses equal-weight paired percentages, preserves signs and rejects equality with its exclusive limit",()=>{
 const reference=Array.from({length:5},(_,i)=>attempt(i+1));
 const compare=(percentages)=>compareCell(reference,percentages.map((cpu,i)=>attempt(i+1,{workload_cpu_us:1000000+cpu*100000})),plan.budgets);
 const skewed=compare([1,1,1,1,5]);assert.equal(skewed.stats.cpu_mean_percent_one_core,1.8);assert.equal(skewed.stats.cpu_median_percent_one_core,1);assert.equal(skewed.passed,true);
 const boundary=compare([0,0,0,0,12.5]);assert.equal(boundary.stats.cpu_mean_percent_one_core,2.5);assert.deepEqual(boundary.failures,["cpu_mean_percent_one_core"]);
 const below=compare([2.499,2.499,2.499,2.499,2.499]);assert.equal(below.passed,true);
 const signed=compare([-3,-1,0,1,2]);assert.equal(signed.stats.cpu_mean_percent_one_core,-0.2);assert.equal(signed.signed_pairs[0].cpu,-3);
 const variedReference=[attempt(1,{outer_ms:1000}),attempt(2,{outer_ms:10000})];
 const variedCandidate=[attempt(1,{outer_ms:1000,workload_cpu_us:1010000}),attempt(2,{outer_ms:10000,workload_cpu_us:1010000})];
 assert.equal(compareCell(variedReference,variedCandidate,plan.budgets).stats.cpu_mean_percent_one_core,0.55);
 const noisy=compare([-3,-1,1,3,5]);assert.equal(noisy.stats.cpu_mean_percent_one_core,1);assert.ok(noisy.stats.cpu_upper_percent_one_core>2);assert.equal(noisy.passed,true);
 const bounded=reference.map((entry)=>({...entry,measurement:{...entry.measurement,cpu_error_us:1000000}}));
 const diagnostic=compareCell(reference,bounded,plan.budgets);assert.equal(diagnostic.passed,true);assert.ok(diagnostic.stats.cpu_error_bound_percent_one_core>2);
});
test("CPU qualification gates each cell separately and applies leave-one-block-out stability to its mean",()=>{
 const makeAttempts=(percentages)=>{const values=[];for(let block=1;block<=5;block++)for(const mode of plan.modes)for(const readers of plan.reader_counts)values.push({...attempt(block,{reader_cpu_us:mode==="basic"&&readers===4?percentages[block-1]*100000:0}),mode,readers});return values;};
 const failed=assessWorkload(makeAttempts([2.5,2.5,2.5,2.5,2.5]),plan);assert.equal(failed.passed,false);assert.equal(failed.unstable,false);
 assert.deepEqual(failed.cells.filter((cell)=>!cell.passed).map(({mode,readers,failures})=>({mode,readers,failures})),[{mode:"basic",readers:4,failures:["cpu_mean_percent_one_core"]}]);
 const unstable=assessWorkload(makeAttempts([1,1,1,1,7]),plan);assert.equal(unstable.needs_sixth,true);assert.equal(unstable.passed,false);
});
test("qualification cannot use self-counters, missing lifecycle, missing memory or unbounded error as zero",()=>{
 for(const field of Object.keys(attempt(1).measurement.coverage)){const value=attempt(1);value.measurement.coverage[field]=false;assert.equal(eligible(value),false)}
 for(const patch of [{cpu_error_us:null},{rss_error_bytes:null},{memory_samples:1},{maximum_memory_gap_ms:201},{outer_ms:0}])assert.equal(eligible(attempt(1,patch)),false);
 assert.equal(compareCell([attempt(1)],[{...attempt(1),measurement:null}],plan.budgets).eligible,false);
});
test("RSS uses time-weighted simultaneous samples in the central eighty percent, never process peaks",()=>{
 const samples=Array.from({length:11},(_,index)=>({at_ms:index*100,rss:index<5?1000:3000,count:2,span_ms:index?1:0,availability:"available",transition_bytes:index===0?1000:index===5?2000:0}));
 const value=summarizeMemory(samples,1000,{bounded:true,pageBytes:4096});
 assert.equal(value.steady_rss_bytes,2000);assert.equal(value.sampled_peak_rss_bytes,3000);assert.equal(value.memory_samples,9);assert.equal(value.rss_error_bytes,8194.5);assert.equal(value.complete,true);
 assert.equal(summarizeMemory(samples,1000).rss_error_bytes,null);
 assert.equal(summarizeMemory(samples.slice(3).map((s,i)=>({...s,transition_bytes:i?s.transition_bytes:s.rss})),1000).complete,false);
 assert.throws(()=>summarizeMemory([{at_ms:20,rss:-1,count:1,span_ms:1}],100));
});


test("sample timing preserves overlapping transitions and missing memory remains unavailable", () => {
 const samples=Array.from({length:11},(_,i)=>({at_ms:i*100,rss:12000,count:2,span_ms:i?10:0,availability:"available",transition_bytes:i===0?12000:i===5?8000:0}));
 // Two residents exchange 4,000 bytes each. Aggregate RSS is unchanged, but
 // 8,000 byte-transitions over 10 ms / 800 ms adds 100 bytes of timing error.
 const value=summarizeMemory(samples,1000,{bounded:true,pageBytes:4096});
 assert.equal(value.steady_rss_bytes,12000);assert.equal(value.rss_error_bytes,8292);
 const missing=samples.map((s,i)=>i===5?{...s,rss:null,availability:"unavailable"}:s);
 const unavailable=summarizeMemory(missing,1000,{bounded:true,pageBytes:4096});
 assert.equal(unavailable.complete,false);assert.equal(unavailable.steady_rss_bytes,null);assert.equal(unavailable.rss_error_bytes,null);
 assert.throws(()=>summarizeMemory(samples.map((s,i)=>i===5?{...s,rss:30000}:s),1000),/transition bound/u);
 assert.throws(()=>summarizeMemory(samples.map((s,i)=>i===5?{...s,span_ms:501}:s),1000));
});
test("memory coverage uses the fixed steady window and retains all outside observations", () => {
 const samples=Array.from({length:11},(_,i)=>({at_ms:i*100,rss:10000,count:1,span_ms:0,availability:"available",transition_bytes:i?0:10000}));
 const outside=samples.map((s,i)=>i===0||i===9||i===10?{...s,rss:null,availability:"unavailable"}:s);
 const value=summarizeMemory(outside,1000,{bounded:true,pageBytes:4096});
 assert.equal(value.complete,true);assert.equal(value.steady_rss_bytes,10000);assert.equal(value.rss_error_bytes,4096);assert.equal(value.sampled_peak_rss_bytes,10000);
 assert.equal(outside.filter((s)=>s.rss===null).length,3);
 const uncertainty=outside.map((s,i)=>i===9?{...s,span_ms:1}:s);
 assert.equal(summarizeMemory(uncertainty,1000,{bounded:true,pageBytes:4096}).complete,false);
 const beginning=outside.map((s,i)=>i===1?{...s,rss:null,availability:"unavailable"}:s);
 assert.equal(summarizeMemory(beginning,1000,{bounded:true,pageBytes:4096}).steady_rss_bytes,null);
 assert.equal(summarizeMemory(samples.map((s)=>({...s,rss:null,availability:"unavailable"})),1000).sampled_peak_rss_bytes,null);
});

test("raw measurement closure rejects forged bindings, duplicate keys, late writes, exhausted records and unsafe files", async(t)=>{
 const {mkdtempSync,rmSync,readFileSync,symlinkSync}=await import("node:fs");const {default:path}=await import("node:path");const {default:os}=await import("node:os");
 const {loadMeasurementTrace}=await import("../qualification/trace.mjs");const {atomicLocalFile}=await import("../../runtime/secure-local-files.mjs");const {digest}=await import("../qualification/policy.mjs");
 const root=mkdtempSync(path.join(os.tmpdir(),"cartulary-raw-measurement-"));t.after(()=>rmSync(root,{recursive:true,force:true}));
 const d=`sha256:${"a".repeat(64)}`,identity={attempt_id:"fixture",run_id:"run",source_digest:d,plan_digest:d,exit_code:0,readers:0};
 const raw={outer_ms:1000,outer_error_ms:2,exit_code:0,workload_cpu_us:12345,reader_cpu_us:0,cpu_error_us:4,reaped_processes:2,adopted_descendants:1,waited_descendants:true,reader_lifecycle:true,reader_starts:[],reader_skips:0,reader_failures:0,rss_bounded:true,page_bytes:4096,cancelled:false,losses:[],envelope_previous_ms:0,envelope_seen_ms:-1,support_cpu_us:-10,support_cpu_error_us:20,support_peak_rss_bytes:4096,support_read_bytes:0,support_write_bytes:0,samples:Array.from({length:10},(_,i)=>({at_ms:i*100,rss:12000,count:2,span_ms:0,availability:"available",transition_bytes:i?0:12000}))};
 const reference=publishMeasurementFixture(root,identity,raw),file=path.join(root,reference.path),original=readFileSync(file);
 const loaded=loadMeasurementTrace(root,identity,reference);assert.deepEqual(loaded,raw);assert.equal(Object.isFrozen(loaded.samples[0]),true);
 assert.throws(()=>loadMeasurementTrace(root,{...identity,source_digest:`sha256:${"b".repeat(64)}`},reference),/identity/u);
 assert.throws(()=>loadMeasurementTrace(root,identity,{...reference,digest:d}),/digest/u);
 assert.throws(()=>loadMeasurementTrace(root,identity,{...reference,path:"../outside"}),/reference/u);
 for(const bytes of [Buffer.from(original.toString().replace('"kind":"header"','"kind":"header","kind":"header"')),Buffer.concat([original,Buffer.from('{}\n')]),original.subarray(0,original.length-2),Buffer.from(`${" ".repeat(65536)}{}\n`)]){
  atomicLocalFile(file,bytes,{replace:true});assert.throws(()=>loadMeasurementTrace(root,identity,{...reference,bytes:bytes.length,digest:digest(bytes)}));
 }
 atomicLocalFile(file,original,{replace:true});rmSync(file);symlinkSync("/dev/null",file);assert.throws(()=>loadMeasurementTrace(root,identity,reference));rmSync(file);
 // Availability is independently checked even when the shape and digest close.
 const invalid=publishMeasurementFixture(root,identity,{...raw,samples:[{...raw.samples[0],rss:null}]});
 assert.throws(()=>loadMeasurementTrace(root,identity,invalid),/availability/u);
});
test("complete qualification keeps the authored plan and deterministic randomized treatment order",()=>{
 assert.equal(validatePlan(plan).length,71);assert.throws(()=>validatePlan({...plan,reader_cadence_ms:1000}));
 assert.equal(plan.schema_id,"cartulary.harness_instrumentation_qualification_plan.v3");
 assert.equal(plan.budgets.cpu_mean_percent_one_core,2.5);
 assert.throws(()=>validatePlan({...plan,schema_id:"cartulary.harness_instrumentation_qualification_plan.v2"}));
 assert.throws(()=>validatePlan({...plan,schema_id:"cartulary.harness_instrumentation_qualification_plan.v1"}));
 assert.throws(()=>validatePlan({...plan,budgets:{...plan.budgets,cpu_mean_percent_one_core:3}}));
 const first=treatmentOrder(plan,"go","measured",1);assert.deepEqual(treatmentOrder(plan,"go","measured",1),first);assert.notDeepEqual(treatmentOrder(plan,"go","measured",2),first);assert.equal(new Set(first.map((v)=>`${v.mode}:${v.readers}`)).size,6);
 const attempts=[];for(let block=1;block<=5;block++)for(const mode of plan.modes)for(const readers of plan.reader_counts)attempts.push({...attempt(block),mode,readers});
 assert.equal(assessWorkload(attempts,plan).passed,true);
 const partial=assessWorkload(attempts.filter((entry)=>entry.block<5),plan);assert.equal(partial.passed,false);assert.ok(partial.cells.every((cell)=>!cell.eligible && cell.failures.includes("incomplete_blocks")));
 const noisy=attempts.find((v)=>v.block===5&&v.mode==="basic"&&v.readers===4);noisy.measurement.outer_ms=11000;
 assert.equal(assessWorkload(attempts,plan).needs_sixth,true);
});
test("external receipt publication and diagnostic drain are bounded without counting overlaps twice",()=>{
 const raw={outer_ms:1000,envelope_previous_ms:760,envelope_seen_ms:860};
 const retained={terminalReceipt:{outer_duration_ms:900},envelope:{intervals:[{phase:"collector_shutdown",start_ms:700,end_ms:730}]}};
 const value=finalizationIntervals(raw,retained,"basic");assert.deepEqual(value,{complete:true,intervals:[{start_ms:730,end_ms:860},{start_ms:890,end_ms:1000}]});assert.equal(intervalUnion(value.intervals),240);
 assert.equal(finalizationIntervals(raw,{...retained,terminalReceipt:null},"basic").complete,false);
 assert.deepEqual(finalizationIntervals(raw,{},"off"),{intervals:[],complete:true});
});

test("qualification manifest independently closes every attempt, selection, digest and treatment",async(t)=>{
 const {mkdtempSync,rmSync,readFileSync,symlinkSync}=await import("node:fs");const {default:path}=await import("node:path");const {default:os}=await import("node:os");
 const {atomicLocalFile}=await import("../../runtime/secure-local-files.mjs");const {writeCanonicalFixture}=await import("./retained-fixture.mjs");
 const {validateQualification}=await import("../qualification/validator.mjs");const {digest,manifestSchema,resultSchema}=await import("../qualification/policy.mjs");const {loadRetainedObservability}=await import("../observability.mjs");
 const root=mkdtempSync(path.join(os.tmpdir(),"cartulary-qualification-fixture-"));t.after(()=>rmSync(root,{recursive:true,force:true}));
 const write=(file,value)=>atomicLocalFile(path.join(root,file),`${JSON.stringify(value)}\n`,{replace:true});
 const d=`sha256:${"a".repeat(64)}`,planDigest=validatePlan(plan);
 const manifest={schema_id:manifestSchema,source_commit:"a".repeat(40),source_digest:d,toolchain_digest:d,system_digest:d,policy_digest:d,plan_digest:planDigest,seed:plan.seed,environment:"wsl2_linux",accounting:"linux_subreaper_wait4_proc",attempts:[],exclusions:[],status:"complete"};
 write("plan.json",plan);
 for(const workload of plan.workloads)for(const [phase,block]of[["cold",0],["warmup",0],...[1,2,3,4,5].map((i)=>["measured",i])])for(const treatment of treatmentOrder(plan,workload.id,phase,block)){
  const id=`${workload.id}-${phase}-${block}-${treatment.mode}-${treatment.readers}`,runRef=`runs/${id}`,runRoot=path.join(root,runRef);
  const {manifest:runManifest,summary}=writeCanonicalFixture(runRoot);
  runManifest.source_state="clean";runManifest.target=workload.target;runManifest.declared_inputs={OWNER:workload.owner,ROWS:workload.rows.join(",")};runManifest.cache_mode=phase==="cold"?"cold":"off";runManifest.instrumentation.mode=treatment.mode;
  write(`${runRef}/run-manifest.json`,runManifest);
  if(workload.target!==summary.target){const target=JSON.parse(readFileSync(path.join(runRoot,"target-summaries/test-slice.json"),"utf8"));target.target=workload.target;write(`${runRef}/target-summaries/${workload.target}.json`,target);summary.target=workload.target;summary.artifact_refs=[`target-summaries/${workload.target}.json`];write(`${runRef}/run-summary.json`,summary);rmSync(path.join(runRoot,"target-summaries/test-slice.json"))}
  if(treatment.mode==="off")rmSync(path.join(runRoot,"diagnostics/execution-index.json"));
  else{
   atomicLocalFile(path.join(runRoot,"diagnostics/resource-samples.ndjson"),"");
   write(`${runRef}/diagnostics/resource-index.json`,{schema_id:"cartulary.harness_resource_index.v3",run_id:id,source_digest:d,graph_digest:d,policy_digest:d,status:"complete",clock:"graph_process_monotonic",coverage:"observed_partial",samples:0,sample_bytes:0,sample_digest:digest(""),omitted_observations:0,discovery_truncations:0,failed_sweeps:0,sweeps:0,processes:[],containers:{availability:"not_observed",identities:[],omitted_observations:0,deadline_overruns:0},observer:{qualification:"component_diagnostic_only",collector_thread_cpu_user_us:0,collector_thread_cpu_system_us:0,collector_js_heap_peak_bytes:0,read_bytes:0,write_bytes:0,maximum_sweep_ms:0,shutdown_ms:0,outside_component_accounting:"startup_parent_helpers_readers_publication"}});
   const envelope={schema_id:"cartulary.harness_invocation_envelope.v2",run_id:id,source_digest:d,graph_digest:d,invocation_digest:d,boundary:"graph_main_entry_to_pre_scan",clock:"graph_process_monotonic",canonical_ref:"unit-events.ndjson",intervals:[{phase:"collector_shutdown",start_ms:0,end_ms:10}],exclusions:["make_and_preflight","module_imports","retained_secret_scan","command_return","receipt_publication"]};
   write(`${runRef}/diagnostics/invocation-envelope.json`,envelope);const scan={schema_id:"cartulary.harness_retained_secret_scan.v1",status:"pass",scanned_files:1,policy_digest:d};write(`${runRef}/retained-secret-scan.json`,scan);
   write(`${runRef}/diagnostics/terminal-receipt.json`,{schema_id:"cartulary.harness_terminal_receipt.v1",invocation_digest:d,source_digest:d,outer_duration_ms:10000,child_exit_code:0,outcome:"passed",publication_start_ms:10000,graph_memberships:[{run_digest:digest(id),graph_digest:d,envelope_digest:digest(`${JSON.stringify(envelope)}\n`),scan_digest:digest(`${JSON.stringify(scan)}\n`),scan_outcome:"passed",scan_interval:{start_ms:9990,end_ms:10000}}]});
  }
  const result={...attempt(block,{canonical_ms:10}),schema_id:resultSchema,attempt_id:id,run_id:id,workload:workload.id,...treatment,phase,order:manifest.attempts.length,source_digest:d,plan_digest:planDigest};
  const raw={outer_ms:10000,outer_error_ms:0,exit_code:0,workload_cpu_us:1000000,reader_cpu_us:0,cpu_error_us:2,reaped_processes:1,adopted_descendants:0,waited_descendants:true,reader_lifecycle:true,reader_starts:Array(treatment.readers).fill(1),reader_skips:0,reader_failures:0,rss_bounded:true,page_bytes:4096,cancelled:false,losses:[],envelope_previous_ms:9910,envelope_seen_ms:9911,support_cpu_us:1,support_cpu_error_us:2,support_peak_rss_bytes:1,support_read_bytes:0,support_write_bytes:0,samples:Array.from({length:100},(_,i)=>({at_ms:i*100,rss:1000000,count:1,span_ms:0,availability:"available",transition_bytes:i?0:1000000}))};
  result.measurement_trace=publishMeasurementFixture(root,result,raw);result.support_publication={cpu_us:1,elapsed_ms:1};
  result.measurement=assembleMeasurement(raw,await loadRetainedObservability(runRoot),treatment.mode);
  const ref=`results/${id}.json`;write(ref,result);manifest.attempts.push({result_ref:ref,result_digest:digest(`${JSON.stringify(result)}\n`),run_ref:runRef});if(phase!=="measured")manifest.exclusions.push({attempt_id:id,reason:phase});
 }
 write("qualification.json",manifest);const file=path.join(root,"qualification.json");assert.equal((await validateQualification(file)).passed,true);
 const basic=manifest.attempts.find((entry)=>entry.run_ref.endsWith("basic-0"));const resourceFile=path.join(root,basic.run_ref,"diagnostics/resource-index.json");const resourceBytes=readFileSync(resourceFile);rmSync(resourceFile);await assert.rejects(validateQualification(file),/treatment is incomplete/u);atomicLocalFile(resourceFile,resourceBytes);
 const prefix={...manifest,status:"incomplete",attempts:manifest.attempts.slice(0,3),exclusions:manifest.exclusions.slice(0,3)};write("qualification.json",prefix);assert.equal((await validateQualification(file)).passed,false);write("qualification.json",manifest);
 const original=manifest.attempts[0].result_digest;manifest.attempts[0].result_digest=`sha256:${"b".repeat(64)}`;write("qualification.json",manifest);await assert.rejects(validateQualification(file),/digest/u);manifest.attempts[0].result_digest=original;
 manifest.attempts[0].result_ref="../outside.json";write("qualification.json",manifest);await assert.rejects(validateQualification(file),/unsafe/u);
 manifest.attempts[0].result_ref="missing.json";write("qualification.json",manifest);await assert.rejects(validateQualification(file));
 symlinkSync("/dev/null",path.join(root,"unsafe.json"));manifest.attempts[0].result_ref="unsafe.json";write("qualification.json",manifest);await assert.rejects(validateQualification(file));
});




test("executable provenance and source digest exclude human documentation before status inspection",async(t)=>{
 const {mkdtempSync,mkdirSync,writeFileSync,readFileSync,rmSync}=await import("node:fs");const {default:path}=await import("node:path");const {default:os}=await import("node:os");const {execFileSync}=await import("node:child_process");
 const {repositoryRoot}=await import("../qualification/policy.mjs");const {executableProvenance,buildSourceSnapshot}=await import("../../test-catalog/index.mjs");
 const root=mkdtempSync(path.join(os.tmpdir(),"cartulary-provenance-"));t.after(()=>rmSync(root,{recursive:true,force:true}));
 const policy=readFileSync(path.join(repositoryRoot,"tools/executable_input_policy.json"));mkdirSync(path.join(root,"tools"));writeFileSync(path.join(root,"tools/executable_input_policy.json"),policy);
 const human=JSON.parse(policy).restricted_roots[0];mkdirSync(path.join(root,human));mkdirSync(path.join(root,"nested"));
 const files=[path.join(human,"owner.txt"),"README","notes.md", "nested/note.MD"];
 writeFileSync(path.join(root,"main.js"),"export const n=1;\n");for(const file of files)writeFileSync(path.join(root,file),"initial\n");
 const git=(...args)=>execFileSync("git",args,{cwd:root,stdio:"pipe"});git("init","-q");git("add",".");git("-c","user.name=Fixture","-c","user.email=fixture@example.invalid","commit","-qm","fixture");
 const digest=buildSourceSnapshot(root).digest;for(const file of files)writeFileSync(path.join(root,file),"updated\n");
 assert.equal(executableProvenance(root).source_state,"clean");assert.equal(buildSourceSnapshot(root).digest,digest);
 writeFileSync(path.join(root,"main.js"),"export const n=2;\n");assert.equal(executableProvenance(root).source_state,"dirty");assert.notEqual(buildSourceSnapshot(root).digest,digest);
});

test("qualification requires the actual basic observer, including closed resource publication", () => {
  const retained = { execution: {}, envelope: {}, terminalReceipt: {}, resourceIndex: { failed_sweeps: 0, containers: { deadline_overruns: 0, availability: "not_observed" } }, resources: { mode: "basic", availability: "available", completeness: "partial" } };
  assert.doesNotThrow(() => requireBasicTreatment(retained));
  for (const key of ["execution", "envelope", "terminalReceipt", "resources", "resourceIndex"]) assert.throws(() => requireBasicTreatment({ ...retained, [key]: null }));
  assert.throws(() => requireBasicTreatment({ ...retained, resourceIndex: { ...retained.resourceIndex, failed_sweeps: 1 } }));
  for (const completeness of ["collector_failed", "unsupported", "truncated"]) assert.throws(() => requireBasicTreatment({ ...retained, resources: { ...retained.resources, completeness } }));
  assert.throws(() => requireBasicTreatment({ ...retained, resources: { mode: "basic", availability: "not_observed" } }));
});
