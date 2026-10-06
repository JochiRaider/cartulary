import assert from "node:assert/strict";
import {copyFileSync, readFileSync, renameSync, writeFileSync} from "node:fs";
import path from "node:path";
import {spawnSync} from "node:child_process";
import {fileURLToPath} from "node:url";
import {setTimeout as delay} from "node:timers/promises";
import {buildReceipt, goOutput, pins, verifyArchive} from "../harness/readiness/cryptographic-module.mjs";
import {referencePlatform} from "../harness/readiness/reference-platform.mjs";

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
 if(operation==="prepare") {
   verifyArchive(go,env);
   const client=path.join(work,"windows-probe.exe");
   goOutput(go,["build","-buildvcs=false","-o",client,"./tools/windowsprobe"],{...env,GOOS:"windows",GOARCH:"amd64"});
   save(artifacts,"windows-client.crypto.json",buildReceipt(go,client,env));
   return;
 }
 if(operation!=="observe" && operation!=="replace") throw new Error("unsupported package platform operation");
 if(operation==="observe") {
   const environment={...referencePlatform(command),formal_cmvp_applicability:"unestablished",services:{}};
   for(const service of ["app","postgres","seaweedfs-s3"]) {
     const id=compose("ps","-q",service);assert.match(id,/^[a-f0-9]{64}$/);
     const selected=JSON.parse(command("docker",["inspect","--format",'{"image":{{json .Image}},"limits":{{json .HostConfig}},"mounts":{{json .Mounts}}}',id]));
     environment.services[service]={container_id:id,image:JSON.parse(command("docker",["image","inspect","--format",'{{json .}}',selected.image])).Id,limits:{cpuset:selected.limits.CpusetCpus,nanocpus:selected.limits.NanoCpus,memory:selected.limits.Memory,memory_swap:selected.limits.MemorySwap},mounts:selected.mounts.map(({Type,Name,Source,Destination,RW})=>({Type,Name,Source,Destination,RW})),resolved_image:JSON.parse(command("docker",["image","inspect","--format",'{{json .}}',selected.image]))};
     const resolved=environment.services[service].resolved_image;
     environment.services[service].resolved_image={id:resolved.Id,repo_digests:resolved.RepoDigests,os:resolved.Os,architecture:resolved.Architecture};
     assert.equal(resolved.Os,"linux");assert.equal(resolved.Architecture,"amd64");
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
