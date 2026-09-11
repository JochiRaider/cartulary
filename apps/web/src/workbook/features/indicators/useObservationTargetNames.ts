import { useEffect, useMemo, useSyncExternalStore } from "react";
import { emptyWorkbookQueryState } from "../../models/workbookQuery";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import { ObservationCollection } from "./ObservationCollection";
import { observationIndicatorView } from "./observationModel";
import type { ObservationReadPort } from "./observationOperation";

/** One authorized lookup for all resolved targets in the loaded observation list. */
export function useObservationTargetNames(
  reader: ObservationReadPort,
  generation: number,
  ids: readonly string[],
) {
  const key = [...new Set(ids)].sort().join(":");
  const scope = useMemo(
    () => ({ reader, generation, key }),
    [reader, generation, key],
  );
  const pages = useMemo(
    () =>
      new ObservationCollection<WorkbookQueryRow>(
        (cursor, signal) =>
          scope.reader.records(
            observationIndicatorView,
            emptyWorkbookQueryState(),
            cursor,
            signal,
          ),
        (row) => row.record_id,
        (row) => row.row_version,
      ),
    [scope],
  );
  const state = useSyncExternalStore(pages.subscribe, pages.getSnapshot);
  useEffect(() => {
    if (key) void pages.load();
    return () => pages.dispose();
  }, [pages, key]);
  const labels = new Map(
    state.items.map((row) => [
      row.record_id,
      `${String(row.cells["indicator.display_value"]?.value ?? "Indicator")} · ${String(row.cells["indicator.indicator_type"]?.value ?? "")}`,
    ]),
  );
  useEffect(() => {
    if (
      state.phase === "ready" &&
      state.hasMore &&
      ids.some((id) => !state.items.some((row) => row.record_id === id))
    )
      void pages.more();
  }, [pages, state, ids]);
  return { labels, pages, state };
}
