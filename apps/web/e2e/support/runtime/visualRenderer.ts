type RendererProfile = {
  schema_id: "cartulary.frontend_visual_renderer_profile.v2";
  profile_id: string;
  chromium_version: string;
  font_manifest_sha256: string;
  locale: string;
  device_scale_factor: number;
  color_scheme: "light" | "dark" | "no-preference";
};

// The Make-owned lease validates the complete active profile before issuing
// this private attestation. Browser consumers receive data, not harness code.
export function attestedVisualRendererProfile(): Readonly<RendererProfile> {
  const profile = JSON.parse(
    process.env.CARTULARY_VISUAL_RENDERER_PROFILE_JSON ?? "null",
  ) as RendererProfile | null;
  if (
    process.env.CARTULARY_VISUAL_RENDERER_ATTESTED !== "1" ||
    profile?.schema_id !== "cartulary.frontend_visual_renderer_profile.v2" ||
    profile.profile_id !== process.env.CARTULARY_VISUAL_RENDERER_PROFILE_ID
  ) {
    throw new Error("missing or mismatched visual renderer attestation");
  }
  return Object.freeze(profile);
}

// Executed in the page: a resolved FontFaceSet.ready does not prove every
// required face loaded successfully. Failed and missing faces close the gate.
export async function waitForLoadedVendoredFonts(): Promise<void> {
  await Promise.all([
    document.fonts.load('400 12px "Inter"'),
    document.fonts.load('400 12px "JetBrains Mono"'),
  ]);
  await document.fonts.ready;
  const faces = Array.from(document.fonts);
  for (const family of ["Inter", "JetBrains Mono"]) {
    const familyFaces = faces.filter((face) => face.family === family);
    if (familyFaces.length === 0) {
      throw new Error(`missing vendored font-face for ${family}`);
    }
    const failedFace = familyFaces.find((face) => face.status === "error");
    if (failedFace) {
      throw new Error(`vendored font ${family} failed to load`);
    }
    if (!document.fonts.check(`400 12px "${family}"`)) {
      throw new Error(`vendored font ${family} is not ready`);
    }
  }
}
