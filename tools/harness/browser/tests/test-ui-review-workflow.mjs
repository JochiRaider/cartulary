import "./test-ui-diagnostic.mjs";
import test from "node:test";
import assert from "node:assert/strict";
import { publicWorkflow } from "./ui-review-public-workflow.mjs";

test("public dev and browser-free artifact workflows keep private detail out of retained receipts and expire all links", async () => { const results = await Promise.allSettled([publicWorkflow({ runID: "workflow-left" }), publicWorkflow({ runID: "workflow-right" })]); for (const result of results) if (result.status === "rejected") throw result.reason; });

test("canonical font readiness rejects failed browser font loads before capture", async () => {
  const { chromium } = await import("playwright");
  const { readFileSync } = await import("node:fs");
  const { waitForLoadedVendoredFonts } = await import("../../../../apps/web/e2e/support/runtime/visualRenderer.ts");
  const fonts = new URL("../../../../apps/web/public/assets/fonts/", import.meta.url);
  const manifest = JSON.parse(readFileSync(new URL("FONT_MANIFEST.json", fonts), "utf8"));
  const families = ["Inter", "JetBrains Mono"].map((family) => ({ family, bytes: [...readFileSync(new URL(manifest.families.find((entry) => entry.family === family).files[0].path, fonts))] }));
  const browser = await chromium.launch({ headless: true });
  try {
    const page = await browser.newPage();
    await page.evaluate(async (families) => {
      for (const { family, bytes } of families) {
        const face = new FontFace(family, new Uint8Array(bytes), { weight: "400" });
        document.fonts.add(face); await face.load();
      }
    }, families);
    await page.evaluate(waitForLoadedVendoredFonts);
    await page.evaluate(async () => {
      const failed = new FontFace("Inter", new Uint8Array([0, 1, 2]), { weight: "400" });
      document.fonts.add(failed); await failed.load().catch(() => {}); await document.fonts.ready;
    });
    await assert.rejects(page.evaluate(waitForLoadedVendoredFonts));
    await page.evaluate(() => document.fonts.clear());
    await assert.rejects(page.evaluate(waitForLoadedVendoredFonts), /missing vendored font-face/u);
  } finally { await browser.close(); }
});
