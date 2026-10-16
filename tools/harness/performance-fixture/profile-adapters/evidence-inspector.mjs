import { createHash } from "node:crypto";
import { canonicalJSONString, validateSchemaSync } from "../../contract/index.mjs";
import { readPerformanceContract } from "./large-grid.mjs";

export function validateEvidenceInspectorObservation(root, profile, observation) {
  const contract=readPerformanceContract(root,"contracts/performance/ac045.v2.json","cartulary.performance_evidence_inspector_contract.v1");
  const digest=`sha256:${createHash("sha256").update(canonicalJSONString(contract.fixture)).digest("hex")}`;
  if (profile.fixture_profile_id!=="evidence_heavy_snapshot_v2" || observation.fixture_profile_id!==profile.fixture_profile_id ||
      observation.criterion_id!==contract.criterion_id || observation.predicate_id!==contract.predicate_id ||
      observation.fixture_id!==contract.fixture.fixture_id || observation.fixture_digest!==digest ||
      observation.measurement_policy_id!==contract.measurement_policy_id || observation.threshold_ms!==contract.threshold_ms ||
      observation.statistics.kind!=="percentile" || observation.statistics.percentile!==contract.percentile ||
      JSON.stringify(observation.traffic)!==JSON.stringify({counts:[{fact_id:"analyst_sessions",value:1},{fact_id:"background_sessions",value:0}],rates:[],conditions:[]}))
    throw new Error("Evidence inspector observation differs from its exact workload");
  const proof=observation.predicate_evidence;
  if (proof!==null) {
    validateSchemaSync("cartulary.evidence_inspector_measurement_proof.v1",proof);
    if (proof.samples.length!==observation.samples.length || new Set(proof.items.map(item=>item.record_id)).size!==5 ||
        proof.items.some(item=>item.record_id===proof.target.record_id))
      throw new Error("Evidence inspector proof bindings or sample count diverged");
    proof.samples.forEach((sample,index)=>{
      const retained=observation.samples[index];
      if (sample.sample_index!==index || sample.warmup!==(index===0) || retained.sample_index!==index || retained.warmup!==sample.warmup ||
          sample.accepted_at>sample.metadata_started_at || sample.metadata_started_at>sample.metadata_completed_at ||
          sample.metadata_completed_at>sample.first_frame_at || sample.first_frame_at>=sample.second_frame_at ||
          sample.total_ms!==sample.second_frame_at-sample.accepted_at || retained.total_ms!==sample.total_ms ||
          JSON.stringify(retained.stages_ms)!==JSON.stringify([{fact_id:"accepted_to_visible_paint",value:sample.total_ms}]))
        throw new Error("Evidence inspector fresh request, consecutive paint or duration proof is inconsistent");
    });
  }
  const warmups=observation.samples.filter(sample=>sample.warmup).length;
  const values=observation.samples.filter(sample=>!sample.warmup).map(sample=>sample.total_ms).sort((a,b)=>a-b);
  const p50=values.length?values[Math.ceil(values.length*0.5)-1]:null;
  const p95=values.length?values[Math.ceil(values.length*0.95)-1]:null;
  if (warmups!==observation.warmup_samples || values.length!==observation.measured_samples || observation.statistics.p50_ms!==p50 || observation.statistics.p95_ms!==p95)
    throw new Error("Evidence inspector retained statistics differ from samples");
  if (["passed","threshold_failed"].includes(observation.outcome) &&
      (proof===null || warmups!==contract.warmup_passes || values.length!==contract.measured_samples ||
       observation.outcome!==(p95<=contract.threshold_ms?"passed":"threshold_failed")))
    throw new Error("Evidence inspector qualifying observation has incomplete sampling or wrong threshold outcome");
}
