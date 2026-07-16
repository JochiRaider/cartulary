import type { CSSProperties, ForwardedRef } from "react";
import { forwardRef } from "react";
import "react-data-grid/lib/styles.css";
import "./styles.css";
import type { GridBlockSizing, GridChrome, GridViewportProps } from "./core";

export {
  assertGridRows,
  buildGridPresentationRows,
  formatGridClipboardTSV,
  type GridActionsColumn,
  type GridBlockSizing,
  type GridCellAnchor,
  type GridCellCopyIntent,
  type GridCellMutationIntent,
  type GridCellRange,
  type GridCellRenderContext,
  type GridCellSelection,
  type GridCellStateContext,
  type GridCellStateInput,
  type GridCellTarget,
  type GridChrome,
  type GridColumn,
  type GridCoreRecordBulkSelection,
  type GridDataRow,
  type GridDataState,
  type GridDataStateAction,
  type GridDensity,
  type GridDraftCellRenderContext,
  type GridDraftRow,
  type GridEditCommitIntent,
  type GridEditCommitOutcome,
  type GridEditorAdapter,
  type GridEditorRenderContext,
  type GridExpandedCellRange,
  type GridFillIntent,
  type GridGroupingDescriptor,
  type GridGroupingScalar,
  type GridHandle,
  type GridInteractionMode,
  type GridMutationIdentity,
  type GridNavigationIntent,
  type GridNavigationKey,
  type GridPasteRowTarget,
  type GridPasteTargetResolution,
  type GridRowGutter,
  type GridRowIdentity,
  type GridRowStateInput,
  type GridSemanticRow,
  type GridSemanticStateInput,
  type GridSortDirection,
  type GridSortEntry,
  type GridStateValidation,
  type GridSurfaceIdentity,
  type GridViewportProps,
  gridClipboardDimensions,
  isGridColumnEditable,
  navigateGridCellAnchor,
  parseGridClipboardTable,
  resolveGridCellAnchor,
  resolveGridCellRange,
  resolveGridPasteTargets,
  type SemanticDataGridProps,
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
        style={resolveViewportStyle(style, chrome, blockSizing)}
      >
        {children}
      </div>
    );
  },
);

function resolveViewportStyle(
  style?: CSSProperties,
  chrome: GridChrome = "sheet",
  blockSizing: GridBlockSizing = "standalone",
): CSSProperties {
  return {
    background: "var(--ct-colors-surface-1)",
    blockSize: blockSizing === "fill" ? "100%" : "min(70vh, 46rem)",
    border: "var(--ct-border-hairline)",
    borderRadius: chrome === "framed" ? "var(--ct-rounded-sm)" : 0,
    boxSizing: blockSizing === "fill" ? "border-box" : undefined,
    color: "var(--ct-colors-ink)",
    minBlockSize: blockSizing === "fill" ? 0 : "18rem",
    minInlineSize: 0,
    overflow: "hidden",
    position: "relative",
    ...(blockSizing === "fill"
      ? { blockSize: "100%", display: "flex", flexDirection: "column" }
      : null),
    ...style,
  };
}
