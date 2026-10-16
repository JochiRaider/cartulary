import type { ClipboardEvent, KeyboardEvent } from "react";
import { clipboardRepresentations } from "./clipboardCodec";
import type { GridClipboardPasteContract } from "./core";
import {
  type GridSemanticPresentationModel,
  planDraftPasteTargets,
} from "./semanticPresentation";

/** Shared event boundary for a recordless insertion cell and its native editor. */
export function draftClipboardHandlers<Row>({
  editable,
  draftVisible,
  clipboardPaste,
  presentation,
  announce,
}: {
  readonly editable: boolean;
  readonly draftVisible: boolean;
  readonly clipboardPaste: GridClipboardPasteContract | undefined;
  readonly presentation: () => GridSemanticPresentationModel<Row>;
  readonly announce: (message: string) => void;
}) {
  return {
    onKeyDownCapture: (event: KeyboardEvent<HTMLElement>) => {
      if (
        !editable ||
        clipboardPaste?.onDraftPaste === undefined ||
        !draftVisible ||
        !(event.target instanceof HTMLElement)
      )
        return;
      const cell = event.target.closest<HTMLElement>('[role="gridcell"]');
      if (!cell?.closest('[data-cartulary-grid-draft-row="true"]')) return;
      // An empty native editor can yield to the recordless insertion cell.
      // Nonempty drafts and native clipboard editing remain source-owned.
      if (
        event.key === "Escape" &&
        !event.nativeEvent.isComposing &&
        (event.target instanceof HTMLInputElement ||
          event.target instanceof HTMLTextAreaElement) &&
        event.target.value === "" &&
        !event.target.readOnly &&
        !event.target.disabled
      ) {
        event.preventDefault();
        event.stopPropagation();
        // Vendor summary cells are not necessarily programmatically focusable.
        // This owner-only target stays out of the sequential tab order.
        cell.tabIndex = -1;
        cell.focus();
        announce(
          "Blank-row insertion cell. Paste a table to create rows, or press Enter to edit this cell.",
        );
      } else if (event.key === "Enter" && event.target === cell) {
        event.preventDefault();
        event.stopPropagation();
        cell
          .querySelector<HTMLElement>(
            "input:not([readonly]), textarea:not([readonly])",
          )
          ?.focus();
      }
    },
    onPasteCapture: (event: ClipboardEvent<HTMLElement>) => {
      if (
        !(event.target instanceof HTMLElement) ||
        event.target.getAttribute("role") !== "gridcell" ||
        !event.target.closest('[data-cartulary-grid-draft-row="true"]') ||
        clipboardPaste?.onDraftPaste === undefined
      )
        return;
      event.preventDefault();
      event.stopPropagation();
      if (!editable || !draftVisible) return;
      const fieldKey =
        event.target.dataset.gridFieldKey ??
        event.target.querySelector<HTMLElement>("[data-grid-field-key]")
          ?.dataset.gridFieldKey;
      if (!fieldKey) return;
      const decoded = clipboardPaste.decode(
        clipboardRepresentations(event.clipboardData),
      );
      if (decoded.kind === "noop") return;
      if (decoded.kind === "failure") {
        clipboardPaste.onError?.(decoded.message);
        announce(decoded.message);
        return;
      }
      if (decoded.kind !== "table") {
        announce("Press Enter to paste text into the blank-row editor.");
        return;
      }
      const targetResolution = planDraftPasteTargets(presentation(), fieldKey, {
        rowCount: decoded.values.length,
        columnCount: decoded.values[0]?.length ?? 0,
        fieldKeys: decoded.fieldKeys,
      });
      if (!targetResolution) {
        const message =
          "The clipboard does not fit the available writable insertion cells.";
        clipboardPaste.onError?.(message);
        announce(message);
        return;
      }
      clipboardPaste.onDraftPaste({
        input: decoded,
        fieldKey,
        targetResolution,
      });
    },
  };
}
