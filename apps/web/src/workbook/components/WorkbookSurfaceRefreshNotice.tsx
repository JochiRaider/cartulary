import { getViewContract } from "@cartulary/view-contracts";
import { useState, useSyncExternalStore } from "react";
import {
  useWorkbookRecoverySource,
  WorkbookRecoveryDetail,
} from "../../shared/WorkbookRecoveryBoundary";
import type { WorkbookRecoveryItem } from "../../shared/workbookRecoveryNavigation";
import type { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";

export function WorkbookSurfaceRefreshNotice({
  runtime,
}: {
  readonly runtime: WorkbookMutationRuntime;
}) {
  const debts = useSyncExternalStore(
    runtime.subscribe,
    runtime.getRefreshRecoverySnapshot,
  );
  const [reading, setReading] = useState(false);
  const items: readonly WorkbookRecoveryItem[] = debts.map((id, order) => ({
    id,
    label: "Refresh saved view",
    summary: "Saved changes; view refresh required",
    origin: getViewContract(id)?.title ?? "Workbook",
    sheetRef: { kind: "view_schema", id },
    attention: "attention",
    order,
    refreshOnlyView: id,
  }));
  const selected = useWorkbookRecoverySource("surface-refresh", items);
  return (
    <WorkbookRecoveryDetail source="surface-refresh" item={selected}>
      <p>Saved changes; view refresh pending.</p>
      <button
        type="button"
        disabled={reading}
        onClick={() => {
          if (!selected) return;
          setReading(true);
          void runtime
            .refreshSurface(selected)
            .finally(() => setReading(false));
        }}
      >
        {reading ? "Refreshing view…" : "Refresh saved view"}
      </button>
    </WorkbookRecoveryDetail>
  );
}
