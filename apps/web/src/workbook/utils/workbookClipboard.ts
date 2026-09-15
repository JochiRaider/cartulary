import {
  type ClipboardDecodeResult,
  type ClipboardRepresentations,
  clipboardFailure,
  clipboardLimits,
  decodeGridClipboard,
  type GridClipboardPasteContract,
} from "@cartulary/grid-adapter";

/** Non-Timeline consumers have no header row to remove. */
export function decodeWorkbookClipboardInput(
  offered: ClipboardRepresentations,
): ClipboardDecodeResult {
  const input = decodeGridClipboard(offered);
  return input.kind === "table" && input.values.length > clipboardLimits.maxRows
    ? clipboardFailure("too_many_rows")
    : input;
}

export function workbookClipboardPasteContract(
  onPaste: GridClipboardPasteContract["onPaste"],
  onError?: (message: string) => void,
): GridClipboardPasteContract {
  return {
    decode: decodeWorkbookClipboardInput,
    onPaste,
    ...(onError ? { onError } : {}),
  };
}
