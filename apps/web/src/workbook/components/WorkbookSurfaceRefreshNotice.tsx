import { useState, useSyncExternalStore } from "react";
import type { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";

/** Presentation of the existing retained surface refresh obligation. */
export function WorkbookSurfaceRefreshNotice({
  runtime,
  viewSchemaId,
}: {
  readonly runtime: WorkbookMutationRuntime;
  readonly viewSchemaId: string;
}) {
  useSyncExternalStore(runtime.subscribe, runtime.getSnapshot);
  const [reading, setReading] = useState(false);
  if (!runtime.surfaceRefreshRequired(viewSchemaId)) return null;
  return (
    <div role="status" data-grid-editor-external-action="true">
      Saved changes; view refresh pending.{" "}
      <button
        type="button"
        disabled={reading}
        onClick={() => {
          setReading(true);
          void runtime
            .refreshSurface(viewSchemaId)
            .finally(() => setReading(false));
        }}
      >
        {reading ? "Refreshing view…" : "Refresh saved view"}
      </button>
    </div>
  );
}
