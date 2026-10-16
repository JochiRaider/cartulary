import assert from "node:assert/strict";
import {createHash} from "node:crypto";
import {chmodSync, copyFileSync, mkdirSync, readFileSync, renameSync, writeFileSync} from "node:fs";
import path from "node:path";
import {spawnSync} from "node:child_process";
import {fileURLToPath} from "node:url";
import {setTimeout as delay} from "node:timers/promises";
import {buildReceipt, goOutput, pins, verifyArchive} from "../harness/readiness/cryptographic-module.mjs";
import {referencePlatform} from "../harness/readiness/reference-platform.mjs";
import {assertPackagedRendererLayout} from "../report-renderer/stage-installation.mjs";

function command(binary,args,options={}) {
  const r=spawnSync(binary,args,{encoding:"utf8",timeout:180000,maxBuffer:16*1024*1024,...options});
  if(r.error || r.status!==0) {
    const diagnostic = /^Windows boundary probe failed: [a-zA-Z0-9 :_-]+$/.test(r.stderr?.trim() || "") ? `: ${r.stderr.trim()}` : "";
    const reasons = [...`${r.stdout || ""}\n${r.stderr || ""}`.matchAll(/(?:"reason_code"\s*:\s*"|reason_code=)([a-z0-9_]+)/g)].map(match => match[1]);
    const gate=[...`${r.stderr || ""}`.matchAll(/"gate":"([a-z_]+)"/g)].at(-1)?.[1] || "unspecified";
    throw new Error(`package platform command failed: ${path.basename(binary)} ${args[0] || ""} (status ${r.status}; gate ${gate}; reasons ${[...new Set(reasons)].join(",")})${diagnostic}`);
  }
  return r.stdout.trim();
}
const save=(artifacts,name,value)=>writeFileSync(path.join(artifacts,name),JSON.stringify(value,null,2)+"\n",{mode:0o600});

function rendererProbe(work, artifacts, project, image, host, lsm) {
  assert.equal(host.Privileged,false,"renderer qualification forbids privileged containers");
  assert.ok(!host.CapAdd?.length,"renderer qualification forbids added host capabilities");
  assert.ok(!host.SecurityOpt?.some(option=>/unconfined/.test(option)),"renderer qualification forbids unconfined policies");
  assert.deepEqual(host.CapDrop,["ALL"]);
  const profileFile=path.join(work,"release/assets/application-seccomp.v1.json"),profileBytes=readFileSync(profileFile);
  const profile=JSON.parse(profileBytes),policyDigest=createHash("sha256").update(profileBytes).digest("hex");
  const seccomp=host.SecurityOpt.find(option=>option.startsWith("seccomp="));
  assert.ok(seccomp,"application outer seccomp profile missing");
  assert.deepEqual(JSON.parse(seccomp.slice("seccomp=".length)),profile,"deployed seccomp profile differs from installed release");
  const args=["create","--network","none","--read-only","--label",`com.docker.compose.project=${project}`,
    "--tmpfs","/tmp:rw,nosuid,nodev,size=1g,mode=1777",
    "--mount",`type=bind,source=${path.join(work,"render-probe")},target=/render-probe,readonly`,
    "--mount",`type=bind,source=${path.join(work,"render-fixtures")},target=/probe-fixtures,readonly`,
    "--entrypoint","/render-probe"];
  for(const capability of host.CapDrop || []) args.push("--cap-drop",capability);
  for(const option of host.SecurityOpt || []) args.push("--security-opt",option.startsWith("seccomp=") ? `seccomp=${profileFile}` : option);
  const id=command("docker",[...args,image]);assert.match(id,/^[a-f0-9]{64}$/);
  try {
    const actualLSM=JSON.parse(command("docker",["inspect","--format",'{"apparmor":{{json .AppArmorProfile}},"process_label":{{json .ProcessLabel}}}',id]));
    assert.equal(actualLSM.apparmor,lsm.apparmor,"probe must use the deployed AppArmor policy");
    assert.equal(actualLSM.process_label.split(":").slice(0,3).join(":"),lsm.process_label.split(":").slice(0,3).join(":"),"probe must use the deployed SELinux domain");
    const execution=spawnSync("docker",["start","--attach",id],{encoding:"utf8",timeout:300000,maxBuffer:64*1024});
    const state=JSON.parse(command("docker",["inspect","--format","{{json .State}}",id]));
    const result=JSON.parse(execution.stdout || "{}");
    save(artifacts,"renderer-container.json",{image,container_id:id,exit_code:state.ExitCode,oom_killed:state.OOMKilled,
      security_options:host.SecurityOpt.map(option=>option.startsWith("seccomp=") ? `seccomp=sha256:${policyDigest}` : option),cap_drop:host.CapDrop,lsm:actualLSM,qualification:result});
    assert.equal(execution.error,undefined,"renderer container observation failed");
    assert.equal(execution.status,0,`renderer container rejected: ${result.reason_code || "unavailable"}`);
    assert.equal(state.ExitCode,0);assert.equal(state.OOMKilled,false);
    assert.equal(result.schema_id,"cartulary.packaged_renderer_qualification.v1");assert.equal(result.status,"pass");
    assert.equal(result.outer_policy_verified,true);
    assert.deepEqual(result.operations.map(item=>item.operation),["mermaid_svg","slidev_pdf"]);
    for(const item of result.operations) {assert.equal(item.fresh_attempts,2);assert.equal(item.native_loopback_observed,true);assert.equal(item.cleanup_verified,true);}
  } finally { command("docker",["rm","--force",id]); }
}

export function windowsProbe(work,origin,credentials={}) {
  const value=JSON.parse(command(path.join(work,"windows-probe.exe"),[],{input:JSON.stringify({origin,ca:readFileSync(path.join(work,"tls/ca.pem"),"utf8"),...credentials})}));
  const policy=pins();
  assert.equal(value.status,"pass");assert.equal(value.client_os,"windows");
  assert.equal(value.toolchain,policy.go_toolchain);assert.equal(value.module_version,policy.cryptographic_module.service_version);
  assert.equal(value.module_selector,policy.cryptographic_module.selector);assert.equal(value.module_enabled,true);
  assert.ok(Math.abs(Date.now()-value.unix_ms)<30000,"Windows/guest clock skew exceeds 30 seconds");
  return value;
}

async function main() {
 const [operation,work,artifacts,project,origin]=process.argv.slice(2);
 assert.ok(path.isAbsolute(work || "") && path.isAbsolute(artifacts || ""));
 assert.match(project,/^cartulary(?:mvpsmoke|mvprecoverysmk)[0-9]+$/);
 const go=process.env.GO || "go", policy=pins();
 const env={...process.env,GOTOOLCHAIN:policy.go_toolchain,GOFIPS140:policy.cryptographic_module.selector,CGO_ENABLED:"0",GOCACHE:process.env.GO_CACHE_DIR,GOMODCACHE:process.env.GO_MOD_CACHE_DIR,GOTMPDIR:process.env.GO_TMP_DIR};
 const compose=(...args)=>command("docker",["compose","--project-name",project,"--env-file",path.join(work,".env"),"-f",path.join(work,"release/assets/docker-compose.yml"),...args]);
 if(operation==="layout") {
   const id=origin;assert.match(id,/^[a-f0-9]{64}$/);
   const receiptFile=path.join(work,"image-renderer-receipt.json"),manifestFile=path.join(work,"image-renderer-manifest.json");
   command("docker",["cp",`${id}:/opt/cartulary/report-renderer/installation.json`,receiptFile]);
   const receiptBytes=readFileSync(receiptFile),digest=JSON.parse(receiptBytes).manifest_sha256;assert.match(digest,/^[a-f0-9]{64}$/);
   command("docker",["cp",`${id}:/opt/cartulary/report-renderer/${digest}.json`,manifestFile]);
   const admitted=assertPackagedRendererLayout(readFileSync(path.join(work,"image-files.txt"),"utf8").trimEnd().split("\n"),receiptBytes,readFileSync(manifestFile));
   save(artifacts,"renderer-image-layout.json",{installation_sha256:admitted,exact_file_set:true});
   return;
 }
 if(operation==="prepare") {
   verifyArchive(go,env);
   const client=path.join(work,"windows-probe.exe");
   goOutput(go,["build","-buildvcs=false","-o",client,"./tools/windowsprobe"],{...env,GOOS:"windows",GOARCH:"amd64"});
   save(artifacts,"windows-client.crypto.json",buildReceipt(go,client,env));
   const renderer=path.join(work,"render-probe");
   goOutput(go,["build","-buildvcs=false","-o",renderer,"./tools/renderprobe"],{...env,GOOS:"linux",GOARCH:"amd64"});
   chmodSync(renderer,0o555);
   save(artifacts,"renderer-probe.crypto.json",buildReceipt(go,renderer,env));
   const fixtures=path.join(work,"render-fixtures");mkdirSync(fixtures,{mode:0o755});chmodSync(fixtures,0o755);
   const vector=path.join(fixtures,"slidev-source.v2.json");
   copyFileSync("contracts/reporting/fixtures/slidev-source.v2.json",vector);chmodSync(vector,0o444);
   return;
 }
 if(operation!=="observe" && operation!=="replace") throw new Error("unsupported package platform operation");
 if(operation==="observe") {
   const environment={...referencePlatform(command),formal_cmvp_applicability:"unestablished",services:{}};
   for(const service of ["app","postgres","seaweedfs-s3"]) {
     const id=compose("ps","-q",service);assert.match(id,/^[a-f0-9]{64}$/);
     const selected=JSON.parse(command("docker",["inspect","--format",'{"image":{{json .Image}},"limits":{{json .HostConfig}},"mounts":{{json .Mounts}},"lsm":{"apparmor":{{json .AppArmorProfile}},"process_label":{{json .ProcessLabel}}}}',id]));
     environment.services[service]={container_id:id,image:JSON.parse(command("docker",["image","inspect","--format",'{{json .}}',selected.image])).Id,limits:{cpuset:selected.limits.CpusetCpus,nanocpus:selected.limits.NanoCpus,memory:selected.limits.Memory,memory_swap:selected.limits.MemorySwap},mounts:selected.mounts.map(({Type,Name,Source,Destination,RW})=>({Type,Name,Source,Destination,RW})),resolved_image:JSON.parse(command("docker",["image","inspect","--format",'{{json .}}',selected.image]))};
     const resolved=environment.services[service].resolved_image;
     environment.services[service].resolved_image={id:resolved.Id,repo_digests:resolved.RepoDigests,os:resolved.Os,architecture:resolved.Architecture};
     assert.equal(resolved.Os,"linux");assert.equal(resolved.Architecture,"amd64");
     if(service==="app") rendererProbe(work,artifacts,project,selected.image,selected.limits,selected.lsm);
     if(service==="app") for(const binary of ["server","migrate","operator"]) {
       const target=path.join(work,`qualified-${binary}`);
       command("docker",["cp",`${id}:/usr/local/bin/cartulary-${binary}`,target]);
       save(artifacts,`${binary}.crypto.json`,buildReceipt(go,target,env));
     }
   }
   save(artifacts,"environment.json",environment);
   save(artifacts,"windows-https.json",windowsProbe(work,origin));
   return;
 }
 const operate=(action)=>command(path.join(work,"release/assets/scripts/package.sh"),[action],{env:{...process.env,CARTULARY_MVP_DIR:work},timeout:300000});
 const stages=[];
 const checkpoint=(stage)=>{stages.push(stage);save(artifacts,"certificate-replacement-progress.json",{completed:stages});};
 const before=windowsProbe(work,origin);
 const formatIdentity=()=>compose("exec","-T","--user","postgres","postgres","psql","-U","postgres","-d","cartulary","-Atc","SELECT format_id FROM application_crypto_format");
 const admittedFormat=formatIdentity();
 assert.equal(admittedFormat,"cartulary.application_crypto_format.v1");
 operate("stop");
 for(const name of ["postgres","seaweed","application","migration","runtime","recovery","restore-migration","restore-recovery"]) for(const suffix of ["crt","key"]) {
   const target=path.join(work,"tls",`${name}.${suffix}`);
   copyFileSync(path.join(work,"tls",`${name}.replacement.${suffix}`),target+".next");renameSync(target+".next",target);
 }
 operate("start");
 checkpoint("installed startup renewed service transports");
 operate("stop");
 // A valid certificate for a different database role must fail authentication.
 for(const suffix of ["crt","key"]) {
   const target=path.join(work,"tls",`runtime.${suffix}`);
   copyFileSync(target,target+".admitted");
   copyFileSync(path.join(work,"tls",`migration.${suffix}`),target+".next");renameSync(target+".next",target);
 }
 const wrongPurpose=spawnSync("docker",["compose","--project-name",project,"--env-file",path.join(work,".env"),"-f",path.join(work,"release/assets/docker-compose.yml"),"run","--rm","--no-deps","app"],{encoding:"utf8",timeout:30000,maxBuffer:1024*1024});
 for(const suffix of ["crt","key"]) {
   const target=path.join(work,"tls",`runtime.${suffix}`);renameSync(target+".admitted",target);
 }
 assert.equal(wrongPurpose.status,2,"wrong database certificate purpose must reject actual server startup");
 assert.match(wrongPurpose.stdout,/postgres_admission_failed/);
 checkpoint("wrong certificate purpose rejected without format mutation");
 assert.equal(formatIdentity(),admittedFormat);
 compose("run","--rm","--no-deps","migrate");
 checkpoint("renewed migration credential admitted");
 let initialized;
 try { initialized=JSON.parse(compose("run","--rm","--no-deps","object-store-init")); }
 catch(error) {
   const log=spawnSync("docker",["compose","--project-name",project,"--env-file",path.join(work,".env"),"-f",path.join(work,"release/assets/docker-compose.yml"),"logs","--no-color","--tail","120","seaweedfs-s3"],{encoding:"utf8",timeout:10000,maxBuffer:1024*1024});
   const lines=`${log.stdout || ""}\n${log.stderr || ""}`.split("\n").filter(line=>!/(secret|credential|access.?key|authorization|token|password|BEGIN.*KEY)/i.test(line));
   writeFileSync(path.join(artifacts,"seaweed-replacement-diagnostic.log"),lines.join("\n"),{mode:0o600});
   const service=compose("ps","-q","seaweedfs-s3");
   if(service) save(artifacts,"seaweed-replacement-state.json",JSON.parse(command("docker",["inspect","--format",'{{json .State}}',service])));
   throw error;
 }
 assert.equal(initialized.already_exists,true,"renewal must preserve the existing bucket");
 assert.equal(initialized.created,false,"renewal must never replace retained storage with a fresh bucket");
 checkpoint("renewed recovery credential and S3 transport admitted");
 operate("start");
 let after;
 for(let attempt=0;attempt<60;attempt++) {try {after=windowsProbe(work,origin);break;} catch {await delay(1000);}}
 assert.ok(after,"renewed application did not become ready from Windows");
 assert.notEqual(before.leaf_sha256,after.leaf_sha256);
 const appID=compose("ps","-q","app");assert.match(appID,/^[a-f0-9]{64}$/);
 command("docker",["kill","--signal","KILL",appID]);operate("start");
 let restarted;
 for(let attempt=0;attempt<60;attempt++) {try {restarted=windowsProbe(work,origin);break;} catch {await delay(1000);}}
 assert.ok(restarted,"owned application did not recover after interruption");
 assert.equal(restarted.leaf_sha256,after.leaf_sha256);assert.equal(formatIdentity(),admittedFormat);
 save(artifacts,"certificate-replacement.json",{before,after,restarted,all_purpose_leaves_replaced:true,wrong_database_certificate_purpose_rejected:true,owned_services_recreated:true,owned_app_interruption_recovered:true,format_identity_preserved:true});
}
if (process.argv[1] && path.resolve(process.argv[1])===fileURLToPath(import.meta.url)) {
 try {await main();} catch(error) {process.stderr.write(`${error.message}\n`);process.exitCode=1;}
}
