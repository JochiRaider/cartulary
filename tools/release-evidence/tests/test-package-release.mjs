import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { spawnSync } from "node:child_process";
import { existsSync, linkSync, mkdirSync, mkdtempSync, readFileSync, readdirSync, rmSync, statSync, symlinkSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import { assertPackagedRendererLayout, stageRendererInstallation } from "../../report-renderer/stage-installation.mjs";
import { restrictedExecutableInputRoots } from "../../harness/test-catalog/restricted-input-boundary.mjs";
const root=path.resolve(import.meta.dirname,"../../..");
const work=mkdtempSync(path.join(tmpdir(),"cartulary-release-contract-"));
const release=path.join(work,"release"), bin=path.join(work,"bin");
const hash=bytes=>createHash("sha256").update(bytes).digest("hex");
function write(relative,value,mode=0o644) { const f=path.join(release,relative);mkdirSync(path.dirname(f),{recursive:true});writeFileSync(f,typeof value==="object" && !Buffer.isBuffer(value)?JSON.stringify(value):value,{mode}); }
try {
 const seccomp=JSON.parse(readFileSync(path.join(root,"deploy/mvp/application-seccomp.v1.json"),"utf8"));
 assert.equal(seccomp.defaultAction,"SCMP_ACT_ERRNO");assert.equal(seccomp.defaultErrnoRet,1);
 assert.deepEqual(seccomp.architectures,["SCMP_ARCH_X86_64"]);
 const decision=(name,args=[])=>seccomp.syscalls.find(rule=>rule.names.includes(name) && (rule.args || []).every(test=>{
  const value=BigInt(args[test.index] || 0),expected=BigInt(test.value);
  if(test.op==="SCMP_CMP_EQ") return value===expected;
  assert.equal(test.op,"SCMP_CMP_MASKED_EQ");return (value&expected)===BigInt(test.valueTwo || 0);
 })) || {action:seccomp.defaultAction,errnoRet:seccomp.defaultErrnoRet};
 for(const syscall of ["bpf","ptrace","setns","io_uring_enter","io_uring_register","keyctl","open_by_handle_at","init_module","userfaultfd"]) assert.equal(decision(syscall).action,"SCMP_ACT_ERRNO",syscall);
 for(const domain of [17,38,40]) assert.equal(decision("socket",[domain]).action,"SCMP_ACT_ERRNO");
 for(const syscall of ["clone3","io_uring_setup"]) assert.equal(decision(syscall).errnoRet,38);
 assert.equal(decision("clone",[17]).action,"SCMP_ACT_ALLOW");
 assert.equal(decision("clone",[2114060305]).action,"SCMP_ACT_ALLOW");
 assert.equal(decision("clone",[1073741841]).action,"SCMP_ACT_ERRNO");
 assert.equal(decision("unshare",[268435456]).action,"SCMP_ACT_ERRNO");
 assert.equal(decision("unshare",[1073741824]).action,"SCMP_ACT_ERRNO");
 // Shipping is an exact file-set copy. No retired installation, unrelated file,
 // documentation, linked entry or altered byte can enter the build context.
 for (const mutation of ["none", "extra", "missing", "changed", "symlink", "hardlink", "human", "duplicate_receipt", "receipt_escape"]) {
  const source=path.join(work,`renderer-${mutation}`), destination=path.join(work,`context-${mutation}`);
  mkdirSync(source);
  const bytes=Buffer.from("synthetic runtime bytes");
  const manifest={schema_id:"cartulary.reporting_runtime_installation.v1",platform:"linux/amd64",node_version:"1.0.0",package_manager:"pnpm@1.0.0",chromium_version:"synthetic",bubblewrap_version:"1.0.0",lockfile_sha256:hash(bytes),package_store_digest:hash(bytes),font_manifest_sha256:hash(bytes),chromium_executable_sha256:hash(bytes),chromium_executable_path:"runtime/browser/chrome-headless-shell",packages:Object.fromEntries(["@slidev/cli","@slidev/theme-default","markdown-it","mermaid","pdf-lib","playwright-chromium"].map(name=>[name,"1.0.0"])),files:[{path:"runtime/executable",sha256:hash(bytes)}]};
  const encoded=Buffer.from(JSON.stringify(manifest)), identity=hash(encoded), selected=path.join(source,`${identity}.root`);
  mkdirSync(path.join(selected,"runtime"),{recursive:true});
  const executable=path.join(selected,"runtime/executable");
  writeFileSync(executable,bytes,{mode:0o555});
  writeFileSync(path.join(source,`${identity}.json`),encoded);
  const receipt={schema_id:"cartulary.reporting_runtime_receipt.v1",manifest_sha256:identity,root:`${identity}.root`,manifest:`${identity}.json`};
  if(mutation==="receipt_escape") receipt.root="../elsewhere";
  let receiptBytes=JSON.stringify(receipt);
  if(mutation==="duplicate_receipt") receiptBytes=receiptBytes.replace('{','{"root":"unexpected",');
  writeFileSync(path.join(source,"installation.json"),receiptBytes);
  if(mutation==="extra") writeFileSync(path.join(selected,"unexpected"),"unbound");
  if(mutation==="changed") {rmSync(executable);writeFileSync(executable,"altered");}
  if(mutation==="missing") rmSync(executable);
  if(mutation==="symlink") {rmSync(executable);symlinkSync(path.join(work,"absent"),executable);}
  if(mutation==="hardlink") linkSync(executable,path.join(source,"alias"));
  if(mutation==="human") for(const restricted of restrictedExecutableInputRoots(root)) symlinkSync(path.join(work,"absent"),path.join(selected,restricted));
  if(mutation==="none") {
   const prefix="opt/cartulary/report-renderer/";
   const entries=[prefix,`${prefix}installation.json`,`${prefix}${identity}.json`,`${prefix}${identity}.root/`,`${prefix}${identity}.root/runtime/`,`${prefix}${identity}.root/runtime/executable`];
   assert.equal(assertPackagedRendererLayout(entries,Buffer.from(receiptBytes),encoded),identity);
   for(const extra of ["usr/bin/node","usr/bin/npm","opt/other/node_modules/unbound.js",`${prefix}${"b".repeat(64)}.root/`,`${prefix}${identity}.root/runtime/unbound.js`]) assert.throws(()=>assertPackagedRendererLayout([...entries,extra],Buffer.from(receiptBytes),encoded));
   assert.throws(()=>assertPackagedRendererLayout(entries.slice(0,-1),Buffer.from(receiptBytes),encoded));
   assert.throws(()=>assertPackagedRendererLayout([...entries,entries.at(-1)],Buffer.from(receiptBytes),encoded));
   mkdirSync(path.join(source,"retired.root"));
   writeFileSync(path.join(source,"retired.root/never-ship"),"retired");
   assert.equal(stageRendererInstallation(source,destination),identity);
   assert.deepEqual(readdirSync(destination).sort(),[`${identity}.json`,`${identity}.root`,"installation.json"].sort());
   assert.equal(statSync(path.join(destination,`${identity}.root/runtime/executable`)).mode&0o777,0o555);
  } else assert.throws(()=>stageRendererInstallation(source,destination),undefined,mutation);
 }
 mkdirSync(bin); mkdirSync(release);
 for(const name of ["install.sh","release-verify.sh"]) { write(`assets/scripts/${name}`,readFileSync(path.join(root,"deploy/mvp/scripts",name)),0o755); }
 write("assets/.env.example","SETTING=literal\n"); write("assets/config.toml.example","# supplied installation state\n");
 const images=["application","postgres","seaweedfs"].map((role,i)=>({role,config_sha256:String(i+1).repeat(64),id:`sha256:${String(i+1).repeat(64)}`,reference:`cartulary/${role}:immutable`,archive:`images/${role}.tar`,inventory:`inventories/${role}.syft.json`}));
 const binaries={};
 for(const name of ["migrate","operator","server"]) { binaries[name]={sha256:"a".repeat(64),receipt:`receipts/${name}.crypto.json`};write(binaries[name].receipt,{schema_id:"cartulary.cryptographic_build_receipt.v1",binary_sha256:binaries[name].sha256,selector:"current"}); }
 for(const image of images) { write(image.archive,"fixture image");write(image.inventory,{descriptor:{name:"syft",version:"1.44.0"},source:{type:"image",metadata:{imageID:image.id}},artifacts:[{name:"component",licenses:[]}]}); }
 write("inventories/scanner.json",{schema_id:"cartulary.image_inventory_tool.v1",module:"github.com/anchore/syft",version:"v1.44.0",binary_sha256:"e".repeat(64)});
 write("inventories/notices.json",{schema_id:"cartulary.distribution_notices.v1",review_disposition:"unestablished",images:images.map(image=>({image_id:image.id}))});
 const files=(dir,prefix="")=>readdirSync(dir).flatMap(name=>{const rel=prefix+name,full=path.join(dir,name),s=statSync(full);return s.isDirectory()?files(full,rel+"/"):[{path:rel,sha256:hash(readFileSync(full)),size_bytes:s.size,mode:(s.mode&0o777).toString(8)}]});
 const manifest={schema_id:"cartulary.local_release_manifest.v1",platform:"linux/amd64",binaries,images,assets:files(release).sort((a,b)=>a.path<b.path?-1:1)};
 write("release-manifest.json",manifest);
 const original=readFileSync(path.join(release,"release-manifest.json"));
 writeFileSync(path.join(bin,"docker"),`#!/usr/bin/env bash
set -eu
if [[ "$1" == info ]]; then echo "\${FIXTURE_PLATFORM:-linux/amd64}"; exit; fi
if [[ "$2" == load ]]; then exit; fi
arg="\${!#}"
if [[ "$arg" == sha256:* ]]; then echo "$arg linux/amd64"; exit; fi
if [[ "\${CONTAMINATED:-0}" == 1 ]]; then echo sha256:wrong; exit; fi
exit 1
`,{mode:0o755});
 for(const name of ["go","node","pnpm","npm"]) writeFileSync(path.join(bin,name),"#!/bin/sh\nexit 99\n",{mode:0o755});
 const env={PATH:`${bin}:/usr/bin:/bin`};
 const verify=(extra={})=>spawnSync(path.join(release,"assets/scripts/release-verify.sh"),[release,"images"],{env:{...env,...extra},encoding:"utf8"});
 assert.equal(verify().status,0,verify().stderr);
 assert.notEqual(verify({FIXTURE_PLATFORM:"linux/arm64"}).status,0);
 assert.notEqual(verify({CONTAMINATED:"1"}).status,0);
 for(const member of ["assets/config.toml.example",images[0].inventory,images[0].archive,binaries.operator.receipt]) {const file=path.join(release,member),before=readFileSync(file);writeFileSync(file,"altered");assert.notEqual(verify().status,0,member);writeFileSync(file,before);}
 write("unexpected","unbound");assert.notEqual(verify().status,0);rmSync(path.join(release,"unexpected"));
 symlinkSync("config.toml.example",path.join(release,"assets/link"));assert.notEqual(verify().status,0);rmSync(path.join(release,"assets/link"));
 linkSync(path.join(release,"assets/config.toml.example"),path.join(work,"linked"));assert.notEqual(verify().status,0);rmSync(path.join(work,"linked"));
 const installed=path.join(work,"installation");const install=()=>spawnSync(path.join(release,"assets/scripts/install.sh"),[installed],{env,encoding:"utf8"});
 const result=install();assert.equal(result.status,0,result.stderr);
 assert.equal(readFileSync(path.join(installed,"release/release-manifest.json")).equals(original),true);
 assert.equal(existsSync(path.join(installed,".env")),true);
 assert.equal(existsSync(path.join(installed,"runtime")),true);
 assert.notEqual(install().status,0,"occupied destination cannot be replaced");
} finally {rmSync(work,{recursive:true,force:true});}
