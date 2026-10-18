import { mkdirSync, readFileSync, writeFileSync } from "node:fs";
import path from "node:path";

// Current synthetic physical/borrow facts for retained-reader tests.
export function publishRendererFixture(root, group, runID) {
  const repo = path.resolve(import.meta.dirname, "../../../..");
  const profile = JSON.parse(readFileSync(path.join(repo, "tools/frontend_visual_renderer_profile.json")));
  const borrow = group.renderer_borrow, base = `_shared/renderers/${borrow.allocation_id}`;
  const write = (relative, value) => {
    const file = path.join(root, relative);
    mkdirSync(path.dirname(file), {recursive:true,mode:0o700});
    writeFileSync(file, `${JSON.stringify(value)}\n`, {mode:0o600});
  };
  const observed = Object.fromEntries(["container_image","platform","playwright_version","chromium_revision","chromium_version","font_manifest_sha256"].map(key=>[key,profile[key]]));
  Object.assign(observed,{image_id:`sha256:${"a".repeat(64)}`,core_version:profile.playwright_version,font_files:[{path:"fixture.woff2",bytes:1,sha256:"a".repeat(64)}]});
  write(borrow.attestation_ref,{schema_id:"cartulary.frontend_visual_renderer_attestation.v1",profile,observed});
  write(borrow.trust_ref,{schema_id:"cartulary.browser_tls_trust.v1",tool_sha256:"a".repeat(64),trust_bundle_sha256:"b".repeat(64),authorities_sha256:["c".repeat(64)]});
  write(`${base}/borrowed-${borrow.borrow_id}.json`,{schema_id:"cartulary.harness_renderer_event.v1",run_id:runID,allocation_id:borrow.allocation_id,generation:borrow.generation,borrow_id:borrow.borrow_id,slot_id:1,runtime_profile_id:group.runtime_profile_id,state:"borrowed",unit_id:borrow.unit_id});
}
