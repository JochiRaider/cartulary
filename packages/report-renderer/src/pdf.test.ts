import { PDFDocument } from "pdf-lib";
import { describe, expect, it } from "vitest";

import { normalizeRenderedPDF } from "./pdf.ts";

const binding = {
  admittedAt: "2026-10-14T01:02:03.123456Z",
  identitySHA256: "a".repeat(64),
  expectedPages: 2,
};

async function examplePDF(date: string) {
  const document = await PDFDocument.create();
  document.addPage([1920, 1080]).drawText("Cartulary rendered page one");
  document.addPage([1920, 1080]).drawText("Cartulary rendered page two");
  document.setCreationDate(new Date(date));
  document.setModificationDate(new Date(date));
  return document.save();
}

describe("rendered PDF admission", () => {
  it("normalizes wall-clock metadata while retaining real pages and content", async () => {
    const first = await normalizeRenderedPDF(
      await examplePDF("2026-10-14T10:00:00Z"),
      binding,
    );
    const second = await normalizeRenderedPDF(
      await examplePDF("2026-10-15T10:00:00Z"),
      binding,
    );
    expect(first).toEqual(second);
    const document = await PDFDocument.load(first, { updateMetadata: false });
    expect(document.getPageCount()).toBe(2);
    expect(document.getCreationDate()?.toISOString()).toBe(
      "2026-10-14T01:02:03.000Z",
    );
    expect(document.getModificationDate()?.toISOString()).toBe(
      "2026-10-14T01:02:03.000Z",
    );
    expect(document.context.trailerInfo.ID?.toString()).toContain(
      binding.identitySHA256,
    );
    expect(document.getPage(0).node.Contents()).toBeDefined();
  });

  it("rejects placeholder PDF bytes and incorrect click page counts", async () => {
    await expect(
      normalizeRenderedPDF(
        new TextEncoder().encode("%PDF-1.4\nplaceholder\n%%EOF\n"),
        binding,
      ),
    ).rejects.toMatchObject({ reasonCode: "invalid_rendered_pdf" });
    await expect(
      normalizeRenderedPDF(await examplePDF(binding.admittedAt), {
        ...binding,
        expectedPages: 3,
      }),
    ).rejects.toMatchObject({ reasonCode: "click_export_page_count_mismatch" });
  });

  it("rejects invalid admission bindings before interpreting output", async () => {
    const bytes = await examplePDF(binding.admittedAt);
    for (const changed of [
      { admittedAt: "now" },
      { admittedAt: "2026-02-30T01:02:03Z" },
      { admittedAt: "2026-10-14T01:02:03+01:00" },
      { identitySHA256: "missing" },
      { expectedPages: 0 },
      { expectedPages: 1.5 },
    ]) {
      await expect(
        normalizeRenderedPDF(bytes, { ...binding, ...changed }),
      ).rejects.toMatchObject({
        reasonCode: "invalid_pdf_binding",
      });
    }
  });
});
