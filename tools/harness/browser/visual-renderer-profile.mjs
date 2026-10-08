import { createHash } from "node:crypto";
import { readFileSync } from "node:fs";
import path from "node:path";
import { isDeepStrictEqual } from "node:util";
import { validateSchemaSync } from "../contract/index.mjs";

export const visualRendererProfilePath = "tools/frontend_visual_renderer_profile.json";
const fontManifestPath = "assets/fonts/FONT_MANIFEST.json";
export const sha256 = (bytes) => createHash("sha256").update(bytes).digest("hex");

export function loadVisualRendererProfile(root) {
  const profile = JSON.parse(readFileSync(path.join(root, visualRendererProfilePath), "utf8"));
  validateSchemaSync("cartulary.frontend_visual_renderer_profile.v2", profile);
  return Object.freeze(profile);
}

export function verifyRendererFonts(root, profile, publicRoot = path.join(root, "apps/web/public")) {
  const manifestBytes = readFileSync(path.join(publicRoot, fontManifestPath));
  if (sha256(manifestBytes) !== profile.font_manifest_sha256) throw new Error("renderer font manifest mismatch");
  const manifest = JSON.parse(manifestBytes);
  const files = manifest.families.flatMap((family) => family.files);
  const observedFiles = [];
  if (!files.length) throw new Error("empty renderer font manifest");
  for (const file of files) {
    if (path.isAbsolute(file.path) || file.path.split(/[\\/]/u).includes("..")) throw new Error("invalid font path");
    const bytes = readFileSync(path.join(publicRoot, "assets/fonts", file.path));
    if (bytes.length !== file.bytes || sha256(bytes) !== file.sha256) throw new Error("renderer font bytes mismatch");
    observedFiles.push({ path: file.path, bytes: bytes.length, sha256: sha256(bytes) });
  }
  return { manifestBytes, files: observedFiles };
}

export function validateRendererAttestation(attestation, profile, fontFiles) {
  validateSchemaSync("cartulary.frontend_visual_renderer_attestation.v1", attestation);
  if (!isDeepStrictEqual(attestation.profile, profile)) throw new Error("renderer attestation profile mismatch");
  const observed = attestation.observed;
  if (!isDeepStrictEqual(observed.font_files, fontFiles)) throw new Error("renderer observed font files mismatch");
  for (const [field, expected] of Object.entries({ container_image: profile.container_image, platform: profile.platform, playwright_version: profile.playwright_version, core_version: profile.playwright_version, chromium_revision: profile.chromium_revision, chromium_version: profile.chromium_version, font_manifest_sha256: profile.font_manifest_sha256 })) {
    if (observed[field] !== expected) throw new Error(`renderer observed ${field} mismatch`);
  }
  return attestation;
}
