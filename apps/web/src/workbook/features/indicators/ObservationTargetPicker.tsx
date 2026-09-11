import {
  useEffect,
  useId,
  useMemo,
  useState,
  useSyncExternalStore,
} from "react";
import { emptyWorkbookQueryState } from "../../models/workbookQuery";
import { ObservationCollection } from "./ObservationCollection";
import { ObservationPagingFeedback } from "./ObservationPagingFeedback";
import {
  type ObservationTarget,
  observationIndicatorView,
  observationTypes,
} from "./observationModel";
import type { ObservationReadPort } from "./observationOperation";
import {
  observationField,
  observationInput,
  observationStack,
  observationText,
} from "./observationStyles";

export function ObservationTargetPicker({
  reader,
  generation,
  selected,
  onChange,
  disabled = false,
  label = "Indicator target (optional)",
}: {
  reader: ObservationReadPort;
  generation: number;
  selected: ObservationTarget | null;
  onChange: (target: ObservationTarget | null) => void;
  disabled?: boolean;
  label?: string;
}) {
  const [type, setType] = useState("");
  const typeId = useId(),
    targetId = useId();
  const scope = useMemo(
    () => ({ reader, generation, type }),
    [reader, generation, type],
  );
  const pages = useMemo(() => {
    return new ObservationCollection(
      (cursor, signal) =>
        scope.reader.records(
          observationIndicatorView,
          {
            ...emptyWorkbookQueryState(),
            filters: scope.type
              ? [
                  {
                    fieldKey: "indicator.indicator_type",
                    op: "eq",
                    arg: { value: scope.type },
                  },
                ]
              : [],
          },
          cursor,
          signal,
        ),
      (row) => row.record_id,
      (row) => row.row_version,
    );
  }, [scope]);
  const state = useSyncExternalStore(pages.subscribe, pages.getSnapshot);
  useEffect(() => {
    void pages.load();
    return () => pages.dispose();
  }, [pages]);
  const options = state.items.map((row) => ({
    recordId: row.record_id,
    label: String(row.cells["indicator.display_value"]?.value ?? ""),
    type: String(row.cells["indicator.indicator_type"]?.value ?? ""),
  }));
  const shown =
    selected && !options.some((item) => item.recordId === selected.recordId)
      ? [selected, ...options]
      : options;
  return (
    <fieldset style={observationStack} disabled={disabled}>
      <legend>{label}</legend>
      <div style={observationField}>
        <label htmlFor={typeId}>Indicator type</label>
        <select
          id={typeId}
          style={observationInput}
          value={type}
          onChange={(event) => setType(event.target.value)}
        >
          <option value="">All types</option>
          {observationTypes.map((value) => (
            <option key={value} value={value}>
              {value}
            </option>
          ))}
        </select>
      </div>
      <div style={observationField}>
        <label htmlFor={targetId}>Existing Indicator</label>
        <select
          id={targetId}
          style={observationInput}
          value={selected?.recordId ?? ""}
          onChange={(event) =>
            onChange(
              shown.find((item) => item.recordId === event.target.value) ??
                null,
            )
          }
        >
          <option value="">No Indicator selected</option>
          {shown.map((item) => (
            <option key={item.recordId} value={item.recordId}>
              {item.label} · {item.type}
            </option>
          ))}
        </select>
      </div>
      {selected ? (
        <p style={observationText}>
          Selected: {selected.label} · {selected.type}
        </p>
      ) : null}
      {state.phase === "ready" && !state.items.length ? (
        <p>No matching Indicators.</p>
      ) : null}
      <ObservationPagingFeedback
        pages={pages}
        state={state}
        label="Indicators"
      />
    </fieldset>
  );
}
