import {mkdirSync, readFileSync, writeFileSync} from "node:fs";
import path from "node:path";
import os from "node:os";
import {spawnSync} from "node:child_process";
import {buildReceipt, goOutput, pins, verifyArchive} from "./cryptographic-module.mjs";
import {referencePlatform} from "./reference-platform.mjs";

const [operation, work, artifacts, appID, origin] = process.argv.slice(2);
if (!["prepare","measure"].includes(operation) || !path.isAbsolute(work || "") || !path.isAbsolute(artifacts || "")) throw new Error("invalid capacity assessment invocation");
const policy = pins();
const env = {...process.env, GOTOOLCHAIN:policy.go_toolchain, GOFIPS140:policy.cryptographic_module.selector, GODEBUG:"fips140=on", GOOS:"linux", GOARCH:"amd64", CGO_ENABLED:"0", GOCACHE:process.env.GO_CACHE_DIR, GOMODCACHE:process.env.GO_MOD_CACHE_DIR, GOTMPDIR:process.env.GO_TMP_DIR};
const go = process.env.GO || "go";
const client = path.join(work,"capacity-client");
const save = (name, value) => writeFileSync(path.join(artifacts,name), `${JSON.stringify(value,null,2)}\n`,{mode:0o600});
const command = (binary,args,options={}) => {
 const r = spawnSync(binary,args,{env,encoding:"utf8",maxBuffer:16*1024*1024,timeout:60000,...options});
 if (r.error || r.status !== 0) throw new Error(`assessment command failed: ${path.basename(binary)} ${args[0] || ""}`);
 return r.stdout.trim();
};
try {
 for (const dir of [env.GOCACHE,env.GOMODCACHE,env.GOTMPDIR]) mkdirSync(dir,{recursive:true});
 verifyArchive(go,env);
 if (operation === "prepare") {
  goOutput(go,["build","-buildvcs=false","-o",client,"./tools/credentialassessment"],env);
  save("capacity-client.crypto.json",buildReceipt(go,client,env));
 } else {
  if (!/^[a-f0-9]{64}$/.test(appID || "") || !/^https:\/\/127\.0\.0\.1:[0-9]+$/.test(origin || "")) throw new Error("invalid measured application identity");
  const inspect = JSON.parse(command("docker",["inspect",appID]))[0];
  const {CpusetCpus,NanoCpus,Memory,MemorySwap} = inspect.HostConfig;
  if (CpusetCpus!=="0,1" || NanoCpus!==2000000000 || Memory!==2147483648 || MemorySwap!==2147483648) throw new Error("frozen application resource profile mismatch");
  const cgroup = JSON.parse(command("docker",["exec",appID,"/run/cartulary-capacity-client","--resource-report"]));
  const quota = cgroup["cpu.max"].trim().split(/\s+/).map(Number);
  if (quota.length!==2 || quota[0]/quota[1]!==2 || cgroup["memory.max"].trim()!=="2147483648" || cgroup["memory.swap.max"].trim()!=="0" || !["0-1","0,1"].includes(cgroup["cpuset.cpus.effective"].trim())) throw new Error("effective application cgroup differs from frozen profile");
  const identity = {...referencePlatform(command), effective_limits:{CpusetCpus,NanoCpus,Memory,MemorySwap,cgroup}, disposition:"current_package_credential_assessment", formal_cmvp_applicability:"unestablished", application_image:JSON.parse(command("docker",["image","inspect",inspect.Image])).map(({Id,RepoDigests,Architecture,Os})=>({Id,RepoDigests,Architecture,Os}))};
  const password = JSON.parse(readFileSync(path.join(work,"bootstrap-admin.json"),"utf8")).initial_password;
  const start = process.hrtime.bigint(), wall=Date.now();
  const measurement = spawnSync(client,[],{env:{...env,CARTULARY_ASSESSMENT_ORIGIN:origin,CARTULARY_ASSESSMENT_PASSWORD:password,CARTULARY_ASSESSMENT_ROOT_CERTIFICATE_PATH:path.join(work,"tls/ca.pem")},encoding:"utf8",timeout:180000});
  identity.elapsed_monotonic_ms=Number(process.hrtime.bigint()-start)/1e6;
  identity.elapsed_wall_ms=Date.now()-wall;
  identity.load_after=os.loadavg();
  save("environment.json",identity);
  if (measurement.stdout.trim()) save("observations.json",JSON.parse(measurement.stdout));
  if (measurement.error || measurement.status!==0) throw new Error(measurement.stderr.trim() || "credential measurement failed");
  if (Math.abs(identity.elapsed_wall_ms-identity.elapsed_monotonic_ms)>1000) throw new Error("measurement clock discontinuity invalidated the run");
  process.stdout.write("credential capacity assessment passed on the fixed WSL2 package profile\n");
 }
} catch(error) {
 save("failure.json",{message:error.message});
 process.stderr.write(`${error.message}\n`); process.exitCode=1;
}
