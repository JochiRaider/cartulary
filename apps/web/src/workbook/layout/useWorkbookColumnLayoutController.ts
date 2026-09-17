import type { GridColumnSizingIntent } from "@cartulary/grid-adapter";
import type { ViewContract } from "@cartulary/view-contracts";
import {
  useEffect,
  useLayoutEffect,
  useMemo,
  useSyncExternalStore,
} from "react";
import {
  WorkbookColumnLayoutController,
  type WorkbookColumnSizingBinding,
} from "./WorkbookColumnLayoutController";

export function useWorkbookColumnLayoutController({
  activeContract,
  contextKey = activeContract.viewSchemaId,
}: {
  readonly activeContract: ViewContract;
  readonly contextKey?: string | undefined;
}) {
  const owner = useMemo(() => new WorkbookColumnLayoutController(), []);
  const snapshot = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  useLayoutEffect(() => owner.activate(contextKey), [owner, contextKey]);
  useEffect(() => () => owner.dispose(), [owner]);
  const id = activeContract.viewSchemaId;
  const commands = useMemo(
    () => ({
      onColumnHiddenChange: (field: string, hidden: boolean) =>
        owner.hide(id, field, hidden),
      onColumnMove: (field: string, direction: "earlier" | "later") =>
        owner.move(id, field, direction),
      onColumnReorder: (from: string, to: string) =>
        owner.reorder(id, from, to),
      onColumnSizingIntent: (intent: GridColumnSizingIntent) =>
        owner.onIntent(id, intent),
      onRestoreColumnDefault: (field: string) =>
        owner.restoreDefault(id, field),
      onCancelColumnSizing: owner.cancel,
      bindColumnSizing: (binding: WorkbookColumnSizingBinding) =>
        owner.bind(id, binding),
      readColumnSizing: (field: string) => owner.read(id, field),
      onResetColumns: () => owner.reset(id),
    }),
    [owner, id],
  );
  const activeLayoutState = owner.currentLayoutStateForSurface(id);
  return {
    commands: {
      applyLayoutStateForSurface: owner.applyLayoutStateForSurface,
      currentLayoutStateForSurface: owner.currentLayoutStateForSurface,
    },
    snapshot: {
      activeLayoutState,
      activeLayoutControls: {
        ...commands,
        layoutState: activeLayoutState,
        sizing: {
          read: commands.readColumnSizing,
          onIntent: commands.onColumnSizingIntent,
          restoreDefault: commands.onRestoreColumnDefault,
          cancel: commands.onCancelColumnSizing,
          pendingField: snapshot.pendingField,
          notice: snapshot.notice,
        },
      },
    },
  };
}
