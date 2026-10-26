import { Buffer } from "node:buffer";
import { createHash } from "node:crypto";
import { readFileSync } from "node:fs";
import type { Page } from "@playwright/test";
import { test } from "@playwright/test";
import {
  attestedVisualRendererProfile,
  waitForLoadedVendoredFonts,
} from "../runtime/visualRenderer";

const verifiedFontPages = new WeakMap<Page, string>();

export async function waitForVendoredFonts(page: Page) {
  const profile = attestedVisualRendererProfile();
  if (verifiedFontPages.get(page) !== profile.profile_id) {
    const manifestResponse = await page.request.get(
      new URL("/assets/fonts/FONT_MANIFEST.json", page.url()).href,
    );
    const manifestBytes = await manifestResponse.body();
    if (
      !manifestResponse.ok() ||
      createHash("sha256").update(manifestBytes).digest("hex") !==
        profile.font_manifest_sha256
    )
      throw new Error("served font manifest mismatch");
    const manifest = JSON.parse(manifestBytes.toString("utf8")) as {
      families: { files: { path: string; bytes: number; sha256: string }[] }[];
    };
    const files = manifest.families.flatMap((family) => family.files);
    for (const file of files) {
      const response = await page.request.get(
        new URL(`/assets/fonts/${file.path}`, page.url()).href,
      );
      const bytes = await response.body();
      if (
        !response.ok() ||
        bytes.length !== file.bytes ||
        createHash("sha256").update(bytes).digest("hex") !== file.sha256
      )
        throw new Error("served font integrity mismatch");
    }
    verifiedFontPages.set(page, profile.profile_id);
  }
  await page.evaluate(waitForLoadedVendoredFonts);
}

export async function attachFontManifestDigest() {
  await test.info().attach("font-manifest-sha256", {
    body: Buffer.from(`${fontManifestSha256()}\n`, "utf8"),
    contentType: "text/plain",
  });
}

export function fontManifestSha256(): string {
  const manifest = readFileSync(
    new URL("../../../public/assets/fonts/FONT_MANIFEST.json", import.meta.url),
  );
  return createHash("sha256").update(manifest).digest("hex");
}
