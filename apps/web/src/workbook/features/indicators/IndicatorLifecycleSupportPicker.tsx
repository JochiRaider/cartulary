import { indicatorLifecycleTestId } from "@cartulary/ui-contracts";
import { requireViewContract } from "@cartulary/view-contracts";
import { useEffect, useMemo, useState, useSyncExternalStore } from "react";
import { WorkbookInspectorActionButton } from "../../inspector/presentation/WorkbookInspectorActions";
import { genericReferenceOptionsFromRows } from "../../models/genericWorkbookModel";
import {
  emptyWorkbookQueryState,
  type WorkbookQueryState,
} from "../../models/workbookQuery";
import {
  listWorkbookSurfaceRegistryEntries,
  timelineViewSchemaId,
} from "../../models/workbookSurfaceRegistry";
import type { IndicatorSupportReference } from "./indicatorLifecycleModel";
import type { IndicatorLifecycleOwnerPort } from "./indicatorLifecycleOperation";
import { IndicatorLifecyclePaging } from "./indicatorLifecyclePaging";
import {
  lifecycleField,
  lifecycleInput,
  lifecycleStack,
  lifecycleText,
} from "./indicatorLifecycleStyles";

export function IndicatorLifecycleSupportPicker({
  owner,
  selected,
  disabled,
  onChange,
}: {
  owner: IndicatorLifecycleOwnerPort;
  selected: readonly IndicatorSupportReference[];
  disabled: boolean;
  onChange: (value: readonly IndicatorSupportReference[]) => void;
}) {
  const [viewId, setViewId] = useState<string>(timelineViewSchemaId);
  const [query, setQuery] = useState<WorkbookQueryState>(
    emptyWorkbookQueryState,
  );
  const [text, setText] = useState("");
  const [fieldKey, setFieldKey] = useState("");
  const generation = owner.getSnapshot().generation;
  const contract = requireViewContract(viewId);
  const searchFields = contract.fields.filter(
    (field) =>
      field.filterOps.includes("full_text") ||
      field.filterOps.includes("prefix"),
  );
  const field =
    searchFields.find((field) => field.fieldKey === fieldKey) ??
    searchFields[0];
  const pages = useMemo(() => {
    void generation;
    return new IndicatorLifecyclePaging(
      (cursor, signal) => owner.records(viewId, query, cursor, signal),
      (row) => row.record_id,
    );
  }, [owner, viewId, query, generation]);
  const state = useSyncExternalStore(pages.subscribe, pages.getSnapshot);
  useEffect(() => {
    void pages.load();
    return () => pages.dispose();
  }, [pages]);
  const options = genericReferenceOptionsFromRows(viewId, [...state.items]);
  return (
    <fieldset
      data-testid={indicatorLifecycleTestId("support")}
      disabled={disabled}
      style={lifecycleStack}
    >
      <legend>Supporting records (optional)</legend>
      <p style={lifecycleText}>
        Choose records from this incident. You can leave support empty.
      </p>
      {selected.length ? (
        <ul>
          {selected.map((item) => (
            <li key={item.recordId}>
              {item.label}{" "}
              <WorkbookInspectorActionButton
                aria-label={`Remove support ${item.label}`}
                onClick={() =>
                  onChange(
                    selected.filter(
                      (candidate) => candidate.recordId !== item.recordId,
                    ),
                  )
                }
              >
                Remove
              </WorkbookInspectorActionButton>
            </li>
          ))}
        </ul>
      ) : (
        <p style={lifecycleText}>No supporting records selected.</p>
      )}
      <label style={lifecycleField}>
        Record surface
        <select
          style={lifecycleInput}
          value={viewId}
          onChange={(event) => {
            setViewId(event.target.value);
            setQuery(emptyWorkbookQueryState());
            setText("");
            setFieldKey("");
          }}
        >
          {listWorkbookSurfaceRegistryEntries().map((entry) => (
            <option key={entry.viewSchemaId} value={entry.viewSchemaId}>
              {entry.title}
            </option>
          ))}
        </select>
      </label>
      {field ? (
        <>
          <label style={lifecycleField}>
            Search field
            <select
              style={lifecycleInput}
              value={field.fieldKey}
              onChange={(event) => setFieldKey(event.target.value)}
            >
              {searchFields.map((candidate) => (
                <option key={candidate.fieldKey} value={candidate.fieldKey}>
                  {candidate.label}
                </option>
              ))}
            </select>
          </label>
          <label style={lifecycleField}>
            {field.filterOps.includes("full_text")
              ? "Search words"
              : "Starts with"}
            <input
              style={lifecycleInput}
              value={text}
              onChange={(event) => setText(event.target.value)}
            />
          </label>
          <WorkbookInspectorActionButton
            onClick={() =>
              setQuery({
                ...emptyWorkbookQueryState(),
                filters:
                  text === ""
                    ? []
                    : [
                        {
                          fieldKey: field.fieldKey,
                          op: field.filterOps.includes("full_text")
                            ? "full_text"
                            : "prefix",
                          arg: field.filterOps.includes("full_text")
                            ? { query: text }
                            : { value: text },
                        },
                      ],
              })
            }
          >
            Find supporting records
          </WorkbookInspectorActionButton>
        </>
      ) : null}
      {options.length ? (
        <ul>
          {options.map((item) => (
            <li key={item.recordId}>
              <label>
                <input
                  type="checkbox"
                  value={item.recordId}
                  checked={selected.some(
                    (candidate) => candidate.recordId === item.recordId,
                  )}
                  onChange={(event) =>
                    onChange(
                      event.target.checked
                        ? [...selected, item]
                        : selected.filter(
                            (candidate) => candidate.recordId !== item.recordId,
                          ),
                    )
                  }
                />{" "}
                {item.label}
              </label>
            </li>
          ))}
        </ul>
      ) : state.phase === "ready" ? (
        <p>No matching supporting records.</p>
      ) : null}
      <LifecyclePagingFeedback
        pages={pages}
        state={state}
        label="supporting records"
      />
    </fieldset>
  );
}

export function LifecyclePagingFeedback<T>({
  pages,
  state,
  label,
  controlsOnly = false,
}: {
  pages: IndicatorLifecyclePaging<T>;
  state: ReturnType<IndicatorLifecyclePaging<T>["getSnapshot"]>;
  label: string;
  controlsOnly?: boolean;
}) {
  return (
    <>
      {!controlsOnly &&
      ["initial_loading", "loading_more", "refreshing"].includes(
        state.phase,
      ) ? (
        <p role="status">
          {state.phase === "loading_more"
            ? "Loading more"
            : state.phase === "refreshing"
              ? "Refreshing"
              : "Loading"}{" "}
          {label}…
        </p>
      ) : null}
      {state.failure ? (
        <div>
          {controlsOnly ? null : (
            <p role="status">
              {state.failure.message}
              {state.items.length
                ? " Previously loaded results remain visible and may be incomplete or stale."
                : ""}
            </p>
          )}
          <WorkbookInspectorActionButton onClick={() => void pages.retry()}>
            {state.restartRequired ? `Restart ${label}` : `Retry ${label}`}
          </WorkbookInspectorActionButton>
        </div>
      ) : null}
      {state.hasMore && state.phase === "ready" ? (
        <WorkbookInspectorActionButton onClick={() => void pages.more()}>
          Load more {label}
        </WorkbookInspectorActionButton>
      ) : null}
    </>
  );
}
