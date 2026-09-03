import type { ForwardedRef } from "react";
import { forwardRef } from "react";
import "react-data-grid/lib/styles.css";
import "./styles.css";
import type { GridViewportProps } from "./core";
import { resolveGridViewportStyle } from "./viewportStyle";

export type {
  GridActionsColumn,
  GridCellAnchor,
  GridCellPasteIntent,
  GridCellRange,
  GridCellStateInput,
  GridCellTarget,
  GridClipboardDimensions,
  GridClipboardInput,
  GridClipboardPasteContract,
  GridColumn,
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
  GridGroupingDescriptor,
  GridHandle,
  GridInteractionMode,
  GridNavigationIntent,
  GridNavigationKey,
  GridPasteTargetResolution,
  GridRowGutter,
  GridRowStateInput,
  GridSortEntry,
  GridSurfaceIdentity,
} from "./core";
export { SemanticDataGrid } from "./SemanticDataGrid";

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
