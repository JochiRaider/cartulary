// Historical fixture creation is harness-only. Every restore and maintenance
// action executes the immutable installed package against destroyed source state.
import assert from "node:assert/strict";
import {randomUUID} from "node:crypto";
import {chmodSync,copyFileSync,mkdirSync,readFileSync,unlinkSync,writeFileSync} from "node:fs";
import path from "node:path";
import {spawnSync} from "node:child_process";
import {goOutput,pins} from "../harness/readiness/cryptographic-module.mjs";
import {windowsProbe} from "./package-platform.mjs";

const [work,project,origin,artifacts]=process.argv.slice(2);
assert.match(project,/^cartularymvprecoverysmk[0-9]+$/);
const command=(binary,args,options={})=>{
 const r=spawnSync(binary,args,{encoding:"utf8",timeout:300000,maxBuffer:16*1024*1024,...options});
 if(r.status!==0) {
  const reasons=[...`${r.stdout || ""}\n${r.stderr || ""}`.matchAll(/"reason_code"\s*:\s*"([a-z0-9_]+)"/g)].map(x=>x[1]);
  throw new Error(`portable package command failed: ${path.basename(binary)} ${args[0]} status=${r.status} reasons=${reasons.join(",")}`);
 }
 return r.stdout.trim();
};
const compose=(directory,name,...args)=>command("docker",["compose","--project-name",name,"--env-file",path.join(directory,".env"),"-f",path.join(directory,"release/assets/docker-compose.yml"),...args],{env:{...process.env,CARTULARY_INSTALLATION_DIR:directory}});
const operate=(directory,...args)=>command(path.join(directory,"release/assets/scripts/package.sh"),args,{env:{...process.env,CARTULARY_MVP_DIR:directory}});
const policy=pins(),helper=path.join(work,"historical-input");
goOutput(process.env.GO || "go",["build","-buildvcs=false","-o",helper,"./tools/packagerecovery"],{...process.env,GOTOOLCHAIN:policy.go_toolchain,GOFIPS140:policy.cryptographic_module.selector,CGO_ENABLED:"0",GOCACHE:process.env.GO_CACHE_DIR,GOMODCACHE:process.env.GO_MOD_CACHE_DIR,GOTMPDIR:process.env.GO_TMP_DIR});
chmodSync(helper,0o755);
const transfers=path.join(work,"independent-transfers");mkdirSync(transfers,{mode:0o777});chmodSync(transfers,0o777);
// First verify normal installed export selection before historical fixture input.
const fresh=JSON.parse(operate(work,"backup-export",path.join(transfers,"fresh")));
assert.equal(fresh.result,"succeeded");
operate(work,"stop");
const stale=JSON.parse(compose(work,project,"run","--rm","--no-deps","--volume",`${helper}:/run/historical-input:ro`,"--volume",`${transfers}:/transfer-output`,"--entrypoint","/run/historical-input","recovery-operator"));
assert.ok(Date.now()-Date.parse(stale.consistency_point_at)>39*24*3600000);
const target=path.join(work,"destination"),targetProject=`${project}destination`;
command(path.join(work,"release/assets/scripts/install.sh"),[target]);
for(const name of ["config.toml","bootstrap-admin.json","restore-verification-target.toml","revisions-conflict-token-key-ring.json"]) {copyFileSync(path.join(work,name),path.join(target,name));chmodSync(path.join(target,name),0o644);}
writeFileSync(path.join(target,".env"),readFileSync(path.join(work,".env"),"utf8").replace(/^CARTULARY_MVP_COMPOSE_PROJECT_NAME=.*$/m,`CARTULARY_MVP_COMPOSE_PROJECT_NAME=${targetProject}`));
compose(work,project,"down","--volumes","--remove-orphans");
unlinkSync(path.join(work,"config.toml"));unlinkSync(path.join(work,".env"));
assert.equal(command("docker",["volume","ls","-q","--filter",`label=com.docker.compose.project=${project}`]),"");
const id=stale.backup_set_id,operation=randomUUID();
const args=["restore-bundle",path.join(transfers,"stale"),id,"--operation-id",operation];
const rejected=spawnSync(path.join(target,"release/assets/scripts/package.sh"),args,{encoding:"utf8",timeout:300000,env:{...process.env,CARTULARY_MVP_DIR:target}});
assert.notEqual(rejected.status,0);
assert.equal(JSON.parse(rejected.stdout).error.reason_code,"stale_backup_unacknowledged");
const acknowledged=[...args,"--acknowledge-stale-backup",id];
const restored=JSON.parse(operate(target,...acknowledged));assert.equal(restored.result,"succeeded");
assert.deepEqual(JSON.parse(operate(target,...acknowledged)),restored);
const sql=q=>compose(target,targetProject,"exec","-T","--user","postgres","postgres","psql","-U","postgres","-d","cartulary","-Atc",q);
assert.equal(sql("SELECT count(*) FROM users WHERE email='admin@example.test'"),"1");
assert.equal(sql("SELECT count(*) FROM backup_sets"),"0");
operate(target,"start");
assert.equal(sql("SELECT count(*) FROM backup_sets WHERE consistency_point_at > now()-interval '24 hours'"),"1");
assert.equal(JSON.parse(operate(target,"restore-verify-due")).result,"no_op");
copyFileSync(path.join(work,"windows-probe.exe"),path.join(target,"windows-probe.exe"));mkdirSync(path.join(target,"tls"));copyFileSync(path.join(work,"tls/ca.pem"),path.join(target,"tls/ca.pem"));
const windows=windowsProbe(target,origin);
writeFileSync(path.join(artifacts,"portable-source-loss.json"),JSON.stringify({historical_fixture:stale.fixture,backup_set_id:id,source_services_destroyed:true,source_configuration_removed:true,source_retention_expired:true,stale_acknowledgement_required:true,exact_replay:true,fresh_backup_gate:true,windows},null,2)+"\n",{mode:0o600});
