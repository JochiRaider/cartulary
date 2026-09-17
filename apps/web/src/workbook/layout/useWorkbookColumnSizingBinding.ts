import type { GridColumn, GridHandle } from "@cartulary/grid-adapter";
import { useLayoutEffect, useRef } from "react";
import { workbookColumnSizing } from "../models/workbookColumnSizing";
import type { WorkbookSurfaceLayoutOwner } from "./useWorkbookLayoutFacade";

/** Supplies defaults and a mounted capability, never another width store. */
export function useWorkbookColumnSizingBinding<Row>({
  columns,
  commands,
  gridHandleRef,
}: {
  readonly columns: readonly GridColumn<Row>[];
  readonly commands: WorkbookSurfaceLayoutOwner["commands"];
  readonly gridHandleRef: { readonly current: GridHandle | null };
}) {
  const latest = useRef(columns);
  latest.current = columns;
  const binding = useRef<{
    port: GridHandle["columnSizing"];
    bind: typeof commands.bindColumnSizing;
    dispose: () => void;
  } | null>(null);
  useLayoutEffect(() => {
    const port = gridHandleRef.current?.columnSizing;
    if (
      binding.current !== null &&
      binding.current.port === port &&
      binding.current.bind === commands.bindColumnSizing
    )
      return;
    binding.current?.dispose();
    binding.current = {
      port,
      bind: commands.bindColumnSizing,
      dispose: commands.bindColumnSizing({
        port,
        defaultWidth: (fieldKey) => {
          const width = latest.current.find(
            (column) => column.fieldKey === fieldKey,
          )?.width;
          return typeof width === "number" && Number.isFinite(width)
            ? Math.max(
                workbookColumnSizing.minimumWidthPx,
                Math.min(
                  workbookColumnSizing.maximumWidthPx,
                  Math.round(width),
                ),
              )
            : undefined;
        },
      }),
    };
  });
  useLayoutEffect(
    () => () => {
      binding.current?.dispose();
      binding.current = null;
    },
    [],
  );
}
