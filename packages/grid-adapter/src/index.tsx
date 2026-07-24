import type { ForwardedRef } from "react";
import { forwardRef } from "react";
import "react-data-grid/lib/styles.css";
import "./styles.css";
import type { GridViewportProps } from "./core";
import { resolveGridViewportStyle } from "./viewportStyle";

export type {
  GridActionsColumn,
  GridBlockSizing,
  GridCellAnchor,
  GridCellCopyIntent,
  GridCellPasteIntent,
  GridCellRange,
  GridCellRenderContext,
  GridCellStateContext,
  GridCellStateInput,
  GridCellTarget,
  GridChrome,
  GridClipboardDimensions,
  GridClipboardInput,
  GridClipboardPasteContract,
  GridColumn,
  GridCoreRecordBulkSelection,
  GridDataRow,
  GridDataState,
  GridDataStateAction,
  GridDensity,
  GridDraftCellRenderContext,
  GridDraftRow,
  GridEditCommitIntent,
  GridEditCommitOutcome,
  GridEditorActivation,
  GridEditorAdapter,
  GridEditorFocusTarget,
  GridEditorRenderContext,
  GridExpandedCellRange,
  GridFillIntent,
  GridGroupingDescriptor,
  GridGroupingScalar,
  GridHandle,
  GridInteractionMode,
  GridMutationIdentity,
  GridNavigationIntent,
  GridNavigationKey,
  GridPasteRowTarget,
  GridPasteTargetResolution,
  GridRowGutter,
  GridRowIdentity,
  GridRowStateInput,
  GridSemanticStateInput,
  GridSortDirection,
  GridSortEntry,
  GridStateValidation,
  GridSurfaceIdentity,
  GridViewportProps,
  SemanticDataGridProps,
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
