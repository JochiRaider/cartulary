import { PDFDocument, PDFHexString } from "pdf-lib";

import { RenderFailure } from "./failure.ts";

export interface PDFBinding {
  admittedAt: string;
  identitySHA256: string;
  expectedPages: number;
}

// Chromium emits wall-clock metadata. Bind that metadata to immutable admission,
// then compare complete independently rendered PDFs at the Reporting boundary.
// Normalizing metadata is not by itself a determinism test.
export async function normalizeRenderedPDF(
  bytes: Uint8Array,
  binding: PDFBinding,
): Promise<Uint8Array> {
  const date = new Date(binding.admittedAt);
  if (
    !/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d{1,6})?Z$/.test(
      binding.admittedAt,
    ) ||
    !Number.isFinite(date.valueOf()) ||
    date.toISOString().slice(0, 19) !== binding.admittedAt.slice(0, 19) ||
    !/^[a-f0-9]{64}$/.test(binding.identitySHA256) ||
    !Number.isSafeInteger(binding.expectedPages) ||
    binding.expectedPages < 1
  ) {
    throw new RenderFailure("slidev_export_failed", "invalid_pdf_binding");
  }
  try {
    const document = await PDFDocument.load(bytes, {
      updateMetadata: false,
      throwOnInvalidObject: true,
    });
    if (document.isEncrypted) {
      throw new RenderFailure("slidev_export_failed", "invalid_rendered_pdf");
    }
    if (document.getPageCount() !== binding.expectedPages) {
      throw new RenderFailure(
        "slidev_export_failed",
        "click_export_page_count_mismatch",
        {
          expected: binding.expectedPages,
          observed: document.getPageCount(),
          source: "pdf",
        },
      );
    }
    for (const page of document.getPages()) {
      const { width, height } = page.getSize();
      if (
        !Number.isFinite(width) ||
        !Number.isFinite(height) ||
        width <= 0 ||
        height <= 0
      ) {
        throw new RenderFailure("slidev_export_failed", "invalid_rendered_pdf");
      }
    }
    document.setCreationDate(date);
    document.setModificationDate(date);
    document.context.trailerInfo.ID = document.context.obj([
      PDFHexString.of(binding.identitySHA256),
      PDFHexString.of(binding.identitySHA256),
    ]);
    return await document.save({
      useObjectStreams: false,
      addDefaultPage: false,
    });
  } catch (error) {
    if (error instanceof RenderFailure) throw error;
    throw new RenderFailure("slidev_export_failed", "invalid_rendered_pdf");
  }
}
