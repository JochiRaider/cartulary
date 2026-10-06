import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { spawnSync } from "node:child_process";
import { existsSync, linkSync, mkdirSync, mkdtempSync, readFileSync, readdirSync, rmSync, statSync, symlinkSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
const root=path.resolve(import.meta.dirname,"../../..");
const work=mkdtempSync(path.join(tmpdir(),"cartulary-release-contract-"));
const release=path.join(work,"release"), bin=path.join(work,"bin");
const hash=bytes=>createHash("sha256").update(bytes).digest("hex");
function write(relative,value,mode=0o644) { const f=path.join(release,relative);mkdirSync(path.dirname(f),{recursive:true});writeFileSync(f,typeof value==="object" && !Buffer.isBuffer(value)?JSON.stringify(value):value,{mode}); }
try {
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
