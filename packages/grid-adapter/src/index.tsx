export { gridRowIdentitiesEqual } from "./core";

import type { ForwardedRef } from "react";
import { forwardRef } from "react";
import "react-data-grid/lib/styles.css";
import "./styles.css";
import type { GridViewportProps } from "./core";
import { resolveGridViewportStyle } from "./viewportStyle";

export type {
  ClipboardDecodeResult,
  ClipboardRepresentations,
} from "./clipboardCodec";
export {
  clipboardFailure,
  clipboardLimits,
  decodeDelimitedClipboard,
  decodeGridClipboard,
  encodeClipboardTable,
} from "./clipboardCodec";
export type {
  GridActionsColumn,
  GridCellAnchor,
  GridCellNavigationOptions,
  GridCellNavigationResult,
  GridCellPasteIntent,
  GridCellRange,
  GridCellStateInput,
  GridCellTarget,
  GridClearIntent,
  GridClipboardDimensions,
  GridClipboardInput,
  GridClipboardPasteContract,
  GridColumn,
  GridColumnMeasurement,
  GridColumnSizingIntent,
  GridColumnSizingPort,
  GridCoreRecordBulkSelection,
  GridDataRow,
  GridDataState,
  GridDataStateAction,
  GridDensity,
  GridDraftRow,
  GridEditCommitOutcome,
  GridEditorAdapter,
  GridEditorFocusTarget,
  GridFillIntent,
  GridFocusResult,
  GridFocusTarget,
  GridGroupingDescriptor,
  GridGroupingScalar,
  GridHandle,
  GridInteractionMode,
  GridNavigationIntent,
  GridNavigationKey,
  GridPasteTargetResolution,
  GridPresentationPort,
  GridPresentationSnapshot,
  GridRowGutter,
  GridRowIdentity,
  GridRowStateInput,
  GridSortEntry,
  GridSurfaceIdentity,
} from "./core";
export { SemanticDataGrid } from "./SemanticDataGrid";
export { gridAnchorKey } from "./semanticPresentation";

export const GridViewport = forwardRef<HTMLDivElement, GridViewportProps>(
  function GridViewport(
    {
      blockSizing = "standalone",
      children,
      chrome = "sheet",
      className,
      style,
      testId,
    }: GridViewportProps,
    ref: ForwardedRef<HTMLDivElement>,
  ) {
    return (
      <div
        className={className}
        data-grid-chrome={chrome}
        data-testid={testId}
        ref={ref}
        style={resolveGridViewportStyle(style, chrome, blockSizing)}
      >
        {children}
      </div>
    );
  },
);
