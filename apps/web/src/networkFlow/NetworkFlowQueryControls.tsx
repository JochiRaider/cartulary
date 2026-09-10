import { networkAnalysisTestId } from "@cartulary/ui-contracts";
import {
  type ReactNode,
  useId,
  useLayoutEffect,
  useRef,
  useState,
} from "react";
import type { NetworkFlowTable } from "../services/networkFlowContractAdapter";
import { networkFlowQueryMetadata } from "../services/networkFlowContractAdapter";
import {
  NetworkFlowActionGroup,
  NetworkFlowButton,
  NetworkFlowChoice,
  NetworkFlowField,
  NetworkFlowSelect,
  NetworkFlowTextInput,
} from "./NetworkFlowControls";
import {
  networkFlowColumnLabel,
  networkFlowPresentationColumns,
} from "./networkFlowPresentation";
import {
  acceptedQueryIdentity,
  canonicalFilterKey,
  compileAcceptedDraft,
  compilePredicate,
  defaultGraphQuerySettings,
  emptyPredicateInput,
  type GraphQuerySettings,
  type NetworkFlowAcceptedDraft,
  type NetworkFlowAcceptedQuery,
  type NetworkFlowRejectedDraft,
  type NetworkFlowRejectedQuery,
  predicateSummary,
  type QueryBasicSlot,
  type QueryInput,
  type QueryIssue,
  type QueryPredicateDraft,
  queryField,
  queryOperator,
} from "./networkFlowQueryModel";

type Feedback = {
  readonly issues: readonly QueryIssue[];
  readonly status: "applied" | "pending" | "failed";
};
function fieldLabel(field: string): string {
  return networkFlowColumnLabel(
    "network_flow.column." + field.replace("network_flow.", ""),
  );
}
const basics = [
  {
    slot: "endpoint",
    field: "network_flow.endpoint_ip",
    label: "Endpoint IP",
    id: "endpoint",
  },
  {
    slot: "protocol",
    field: "network_flow.ip_protocol",
    label: "Protocol",
    id: "protocol",
  },
  {
    slot: "bytesMinimum",
    field: "network_flow.bytes_count",
    label: "Minimum bytes",
    id: "minimum-bytes",
  },
  {
    slot: "packetsMinimum",
    field: "network_flow.packets_count",
    label: "Minimum packets",
    id: "minimum-packets",
  },
] as const;

export function NetworkFlowAcceptedQueryControls({
  draft,
  onDraftChange,
  onApply,
  onClear,
  issues,
  status,
  appliedQuery,
  graphMode,
  graphControls,
  graphDirty = false,
  temporal = false,
  tableLabel,
}: {
  readonly draft: NetworkFlowAcceptedDraft;
  readonly onDraftChange: (draft: NetworkFlowAcceptedDraft) => void;
  readonly onApply: () => boolean;
  readonly onClear: () => void;
  readonly appliedQuery: NetworkFlowAcceptedQuery;
  readonly graphDirty?: boolean;
  readonly graphMode: boolean;
  readonly graphControls?: ReactNode;
  readonly temporal?: boolean;
  readonly tableLabel: string;
} & Feedback) {
  const prefix = useId();
  const nextID = useRef(0);
  const [editing, setEditing] = useState<string | null>(null);
  const root = useRef<HTMLElement | null>(null);
  const focusRequested = useRef(false);
  useLayoutEffect(() => {
    if (!focusRequested.current || issues.length === 0) return;
    focusRequested.current = false;
    const target = root.current?.querySelector<HTMLElement>(
      '[aria-invalid="true"]',
    );
    if (target) {
      let parent: HTMLElement | null = target.parentElement;
      while (parent) {
        if (parent instanceof HTMLDetailsElement) parent.open = true;
        parent = parent.parentElement;
      }
      target.focus();
    }
  }, [issues]);
  const issue = (path: string) => issues.find((v) => v.path === path)?.message;
  const updateEntry = (entry: QueryPredicateDraft) =>
    onDraftChange({
      ...draft,
      predicates: draft.predicates.map((p) => (p.id === entry.id ? entry : p)),
    });
  const changeBasic = (
    slot: QueryBasicSlot,
    text: string,
    op?: "eq" | "cidr_contains",
  ) => {
    const basic = basics.find((b) => b.slot === slot);
    if (!basic) return;
    const previous = draft.predicates.find((p) => p.slot === slot);
    const predicates = draft.predicates.filter((p) => p !== previous);
    if (text !== "" || op !== undefined) {
      const entry: QueryPredicateDraft = {
        id: previous?.id ?? prefix + "-" + nextID.current++,
        slot,
        field: basic.field,
        op:
          slot === "endpoint"
            ? (op ??
              (previous?.op === "cidr_contains" ? "cidr_contains" : "eq"))
            : slot === "protocol"
              ? "eq"
              : "range",
        input:
          slot === "bytesMinimum" || slot === "packetsMinimum"
            ? { kind: "range", lower: text, upper: "" }
            : { kind: "scalar", text },
      };
      if (previous)
        onDraftChange({
          ...draft,
          predicates: draft.predicates.map((p) => (p === previous ? entry : p)),
        });
      else onDraftChange({ ...draft, predicates: [...predicates, entry] });
    } else onDraftChange({ ...draft, predicates });
  };
  const compiled = compileAcceptedDraft(draft, graphMode ? "graph" : "rows");
  const dirty =
    !compiled.ok ||
    acceptedQueryIdentity(compiled.value) !==
      acceptedQueryIdentity(appliedQuery);
  return (
    <section
      ref={root}
      aria-label="Network Flow filters"
      className="network-flow-query-band"
      data-testid={networkAnalysisTestId("filters")}
    >
      {graphMode ? graphControls : null}
      <div className="network-flow-query-meta">
        {!graphMode ? (
          <span>
            Table scope: {tableLabel}. Column sorting applies immediately.
          </span>
        ) : null}
        <DraftFeedback
          issues={issues}
          status={status}
          dirty={dirty || (graphMode && graphDirty)}
        />
        <details className="network-flow-applied-query">
          <summary>
            Applied filters ({appliedQuery.filters.length})
            {appliedQuery.timeWindow ? " · time window" : ""}
          </summary>
          {appliedQuery.filters.length === 0 ? (
            <span>No field filters.</span>
          ) : (
            <ul>
              {appliedQuery.filters.map((filter) => (
                <li key={canonicalFilterKey(filter)}>
                  {predicateSummary(filter)}
                </li>
              ))}
            </ul>
          )}
          {appliedQuery.timeWindow ? (
            <p>
              Time window: {appliedQuery.timeWindow.startUTC ?? "unbounded"} to{" "}
              {appliedQuery.timeWindow.endUTC ?? "unbounded"} (end exclusive).
            </p>
          ) : null}
        </details>
      </div>
      {(["startUTC", "endUTC"] as const).map((key) => (
        <QueryText
          key={key}
          id={"network-flow-query-" + (key === "startUTC" ? "start" : "end")}
          label={
            temporal
              ? key === "startUTC"
                ? "Flow starts at or after"
                : "Flow starts before"
              : key === "startUTC"
                ? "Flow overlap starts at"
                : "Flow overlap ends before"
          }
          value={draft[key]}
          error={issue(key)}
          onChange={(value) => onDraftChange({ ...draft, [key]: value })}
        />
      ))}
      {basics.map((basic) => {
        const entry = draft.predicates.find((p) => p.slot === basic.slot);
        const value =
          entry?.input.kind === "scalar"
            ? entry.input.text
            : entry?.input.kind === "range"
              ? entry.input.lower
              : "";
        const error = entry
          ? (issue(entry.id + ".value") ?? issue(entry.id + ".lower"))
          : undefined;
        return (
          <NetworkFlowField
            key={basic.slot}
            htmlFor={"network-flow-query-" + basic.id}
            label={basic.label}
            error={error}
            errorId={"network-flow-query-" + basic.id + "-error"}
          >
            <div
              className={
                basic.slot === "endpoint"
                  ? "network-flow-endpoint-input"
                  : undefined
              }
            >
              {basic.slot === "endpoint" ? (
                <NetworkFlowSelect
                  aria-label="Endpoint IP operator"
                  value={entry?.op ?? "eq"}
                  onChange={(e) => {
                    const op = e.currentTarget.value;
                    if (op === "eq" || op === "cidr_contains")
                      changeBasic("endpoint", value, op);
                  }}
                >
                  <option value="eq">equals</option>
                  <option value="cidr_contains">in CIDR</option>
                </NetworkFlowSelect>
              ) : null}
              <NetworkFlowTextInput
                id={"network-flow-query-" + basic.id}
                aria-label={
                  basic.slot === "endpoint" ? "Endpoint IP value" : undefined
                }
                inputMode={basic.slot === "endpoint" ? "text" : "numeric"}
                value={value}
                aria-invalid={!!error}
                aria-describedby={
                  error
                    ? "network-flow-query-" + basic.id + "-error"
                    : undefined
                }
                onChange={(e) => changeBasic(basic.slot, e.currentTarget.value)}
              />
            </div>
          </NetworkFlowField>
        );
      })}
      <details
        className="network-flow-advanced"
        data-testid={networkAnalysisTestId("advanced-filters")}
      >
        <summary>
          Advanced field filters (
          {draft.predicates.filter((p) => p.slot === "advanced").length})
        </summary>
        <div className="network-flow-popover network-flow-advanced__editor">
          <NetworkFlowButton
            onClick={() => {
              const id = prefix + "-" + nextID.current++;
              onDraftChange({
                ...draft,
                predicates: [
                  ...draft.predicates,
                  {
                    id,
                    field: "network_flow.src_ip",
                    op: "eq",
                    input: { kind: "scalar", text: "" },
                    slot: "advanced",
                  },
                ],
              });
              setEditing(id);
            }}
          >
            Add filter
          </NetworkFlowButton>
          {draft.predicates
            .filter((p) => p.slot === "advanced")
            .map((entry) => {
              const result = compilePredicate(entry);
              const summary = result.ok
                ? predicateSummary(result.value)
                : fieldLabel(entry.field) + " " + entry.op + " — incomplete";
              const hasError = issues.some((v) =>
                v.path.startsWith(entry.id + "."),
              );
              return (
                <fieldset key={entry.id} aria-label={"Filter: " + summary}>
                  <span>{summary}</span>
                  <NetworkFlowButton
                    variant="ghost"
                    aria-label={"Edit " + summary}
                    onClick={() => setEditing(entry.id)}
                  >
                    Edit
                  </NetworkFlowButton>
                  <NetworkFlowButton
                    variant="ghost"
                    aria-label={"Remove " + summary}
                    onClick={(event) => {
                      onDraftChange({
                        ...draft,
                        predicates: draft.predicates.filter(
                          (p) => p.id !== entry.id,
                        ),
                      });
                      setEditing(null);
                      event.currentTarget
                        .closest("details")
                        ?.querySelector<HTMLButtonElement>("button")
                        ?.focus();
                    }}
                  >
                    Remove
                  </NetworkFlowButton>
                  {editing === entry.id || hasError ? (
                    <PredicateEditor
                      entry={entry}
                      issues={issues}
                      onChange={updateEntry}
                    />
                  ) : null}
                </fieldset>
              );
            })}
        </div>
      </details>
      <NetworkFlowActionGroup>
        <NetworkFlowButton
          data-testid={networkAnalysisTestId("accepted-query-apply")}
          variant="primary"
          onClick={() => {
            focusRequested.current = true;
            if (onApply()) focusRequested.current = false;
          }}
        >
          Apply query
        </NetworkFlowButton>
        <NetworkFlowButton
          data-testid={networkAnalysisTestId("accepted-query-clear")}
          onClick={() => {
            setEditing(null);
            onClear();
          }}
        >
          Clear query
        </NetworkFlowButton>
      </NetworkFlowActionGroup>
    </section>
  );
}
function PredicateEditor({
  entry,
  issues,
  onChange,
}: {
  readonly entry: QueryPredicateDraft;
  readonly issues: readonly QueryIssue[];
  readonly onChange: (entry: QueryPredicateDraft) => void;
}) {
  const columns = networkFlowPresentationColumns(
    "network_flow.accepted_rows.v1",
  ).filter((c) => c.filter_operators.length > 0);
  const selected = queryField(entry.field);
  const offered =
    columns.find((c) => c.field_key === entry.field)?.filter_operators ??
    selected?.operators ??
    [];
  const operators = [...offered];
  if (!operators.some((op) => op === entry.op)) operators.push(entry.op);
  const error = (part: string) =>
    issues.find((i) => i.path === entry.id + "." + part)?.message ??
    (part !== "op" || entry.input.kind === "null"
      ? issues.find((i) => i.path === entry.id + ".value")?.message
      : undefined);
  return (
    <div className="network-flow-predicate-editor">
      <NetworkFlowField htmlFor={entry.id + "-field"} label="Field">
        <NetworkFlowSelect
          id={entry.id + "-field"}
          value={entry.field}
          onChange={(e) => {
            const field = queryField(e.currentTarget.value);
            if (!field) return;
            const first = queryOperator(
              columns.find((c) => c.field_key === field.field_key)
                ?.filter_operators[0] ?? field.operators[0],
            );
            if (!first) return;
            onChange({
              ...entry,
              field: field.field_key,
              op: first,
              input: emptyPredicateInput(first),
            });
          }}
        >
          {!columns.some((c) => c.field_key === entry.field) ? (
            <option value={entry.field}>{fieldLabel(entry.field)}</option>
          ) : null}
          {columns.map((c) => (
            <option key={c.field_key} value={c.field_key}>
              {networkFlowColumnLabel(c.label_key)}
            </option>
          ))}
        </NetworkFlowSelect>
      </NetworkFlowField>
      <NetworkFlowField
        htmlFor={entry.id + "-op"}
        label="Operator"
        errorId={entry.id + "-op-error"}
        error={error("op")}
      >
        <NetworkFlowSelect
          id={entry.id + "-op"}
          value={entry.op}
          aria-invalid={!!error("op")}
          aria-describedby={error("op") ? entry.id + "-op-error" : undefined}
          onChange={(e) => {
            const op = queryOperator(e.currentTarget.value);
            if (op) onChange({ ...entry, op, input: emptyPredicateInput(op) });
          }}
        >
          {operators.map((op) => (
            <option key={op} value={op}>
              {op === "range" ? "range" : op}
            </option>
          ))}
        </NetworkFlowSelect>
      </NetworkFlowField>
      <PredicateInput
        input={entry.input}
        id={entry.id}
        timestamp={selected?.kind === "timestamp"}
        error={error}
        onChange={(input) => onChange({ ...entry, input })}
      />
    </div>
  );
}
function PredicateInput({
  input,
  id,
  timestamp,
  error,
  onChange,
}: {
  readonly input: QueryInput;
  readonly id: string;
  readonly timestamp: boolean;
  readonly error: (part: string) => string | undefined;
  readonly onChange: (input: QueryInput) => void;
}) {
  if (input.kind === "null") return <span>This operation has no value.</span>;
  if (input.kind === "list")
    return (
      <QueryList
        id={id}
        label="Value"
        members={input.members}
        error={error}
        onChange={(members) => onChange({ kind: "list", members })}
      />
    );
  if (input.kind === "range")
    return (
      <>
        <QueryText
          id={id + "-lower"}
          label="At least (inclusive)"
          value={input.lower}
          error={error("lower")}
          onChange={(lower) => onChange({ ...input, lower })}
        />
        <QueryText
          id={id + "-upper"}
          label={timestamp ? "Before (exclusive)" : "At most (inclusive)"}
          value={input.upper}
          error={error("upper")}
          onChange={(upper) => onChange({ ...input, upper })}
        />
      </>
    );
  return (
    <QueryText
      id={id + "-value"}
      label="Value"
      value={input.text}
      error={error("value")}
      onChange={(text) => onChange({ kind: "scalar", text })}
    />
  );
}
function QueryText({
  id,
  label,
  value,
  error,
  onChange,
}: {
  readonly id: string;
  readonly label: string;
  readonly value: string;
  readonly error: string | undefined;
  readonly onChange: (value: string) => void;
}) {
  return (
    <NetworkFlowField
      htmlFor={id}
      label={label}
      error={error}
      errorId={id + "-error"}
    >
      <NetworkFlowTextInput
        id={id}
        value={value}
        aria-invalid={!!error}
        aria-describedby={error ? id + "-error" : undefined}
        onChange={(e) => onChange(e.currentTarget.value)}
      />
    </NetworkFlowField>
  );
}
function QueryList({
  id,
  label,
  members,
  error,
  onChange,
  choices,
}: {
  readonly id: string;
  readonly label: string;
  readonly members: readonly string[];
  readonly error: (part: string) => string | undefined;
  readonly onChange: (members: readonly string[]) => void;
  readonly choices?: readonly string[];
}) {
  const prefix = useId();
  const nextId = useRef(0);
  const memberIds = useRef<string[]>([]);
  while (memberIds.current.length < members.length)
    memberIds.current.push(prefix + "-" + nextId.current++);
  memberIds.current.length = members.length;
  return (
    <fieldset aria-label={label + " list"}>
      {members.map((member, index) => (
        <div
          key={memberIds.current[index]}
          className="network-flow-inline-fields"
        >
          {choices ? (
            <NetworkFlowField
              htmlFor={id + "-" + index}
              label={label + " " + (index + 1)}
              error={error("member-" + index)}
              errorId={id + "-" + index + "-error"}
            >
              <NetworkFlowSelect
                id={id + "-" + index}
                value={member}
                aria-invalid={!!error("member-" + index)}
                aria-describedby={
                  error("member-" + index)
                    ? id + "-" + index + "-error"
                    : undefined
                }
                onChange={(e) =>
                  onChange(
                    members.map((v, i) =>
                      i === index ? e.currentTarget.value : v,
                    ),
                  )
                }
              >
                <option value="">Select a value</option>
                {choices.map((c) => (
                  <option key={c} value={c}>
                    {c}
                  </option>
                ))}
              </NetworkFlowSelect>
            </NetworkFlowField>
          ) : (
            <QueryText
              id={id + "-" + index}
              label={label + " " + (index + 1)}
              value={member}
              error={error("member-" + index)}
              onChange={(text) =>
                onChange(members.map((v, i) => (i === index ? text : v)))
              }
            />
          )}
          <NetworkFlowButton
            aria-label={"Remove " + label + " " + (index + 1)}
            onClick={(e) => {
              const group = e.currentTarget.parentElement?.parentElement;
              memberIds.current.splice(index, 1);
              onChange(members.filter((_, i) => i !== index));
              group
                ?.querySelector<HTMLButtonElement>("[data-query-add-member]")
                ?.focus();
            }}
          >
            Remove
          </NetworkFlowButton>
        </div>
      ))}
      <NetworkFlowButton
        data-query-add-member=""
        onClick={() => onChange([...members, ""])}
      >
        Add {label.toLowerCase()}
      </NetworkFlowButton>
    </fieldset>
  );
}
function DraftFeedback({
  issues,
  status,
  dirty,
}: { readonly dirty: boolean } & Feedback) {
  return (
    <div className="network-flow-query-feedback">
      <output>
        {status === "pending"
          ? "Applying query…"
          : status === "failed"
            ? "Query failed. The last successful query remains applied."
            : dirty
              ? "Unapplied draft changes."
              : "Query applied."}
      </output>
      {issues.length ? (
        <div role="alert">
          {issues.map((issue) => (
            <p key={issue.path + issue.message}>{issue.message}</p>
          ))}
        </div>
      ) : null}
    </div>
  );
}
export function NetworkFlowRejectedQueryControls({
  draft,
  onDraftChange,
  onApply,
  onClear,
  issues,
  status,
  appliedQuery,
}: {
  readonly draft: NetworkFlowRejectedDraft;
  readonly onDraftChange: (draft: NetworkFlowRejectedDraft) => void;
  readonly onApply: () => boolean;
  readonly onClear: () => void;
  readonly appliedQuery: NetworkFlowRejectedQuery;
} & Feedback) {
  const root = useRef<HTMLElement | null>(null);
  const focusRequested = useRef(false);
  useLayoutEffect(() => {
    if (focusRequested.current && issues.length) {
      focusRequested.current = false;
      root.current
        ?.querySelector<HTMLElement>('[aria-invalid="true"]')
        ?.focus();
    }
  }, [issues]);
  const error = (path: string) => issues.find((i) => i.path === path)?.message;
  return (
    <section
      ref={root}
      aria-label="Diagnostic filters"
      className="network-flow-query-band"
    >
      <DraftFeedback
        issues={issues}
        status={status}
        dirty={
          draft.lower !== String(appliedQuery.sourceRowRange?.gte ?? "") ||
          draft.upper !== String(appliedQuery.sourceRowRange?.lte ?? "") ||
          JSON.stringify(draft.fieldKeys) !==
            JSON.stringify(appliedQuery.fieldKeys) ||
          JSON.stringify(draft.errorCodes) !==
            JSON.stringify(appliedQuery.errorCodes)
        }
      />
      <QueryList
        id="network-flow-diagnostic-errors"
        label="Error code"
        members={draft.errorCodes}
        choices={networkFlowQueryMetadata.diagnosticErrors}
        error={(part) => error("errorCodes." + part)}
        onChange={(errorCodes) => onDraftChange({ ...draft, errorCodes })}
      />
      <QueryList
        id="network-flow-diagnostic-fields"
        label="Field key"
        members={draft.fieldKeys}
        choices={networkFlowQueryMetadata.diagnosticFields}
        error={(part) => error("fieldKeys." + part)}
        onChange={(fieldKeys) => onDraftChange({ ...draft, fieldKeys })}
      />
      <QueryText
        id="network-flow-diagnostic-row-start"
        label="First source row"
        value={draft.lower}
        error={error("lower")}
        onChange={(lower) => onDraftChange({ ...draft, lower })}
      />
      <QueryText
        id="network-flow-diagnostic-row-end"
        label="Last source row"
        value={draft.upper}
        error={error("upper")}
        onChange={(upper) => onDraftChange({ ...draft, upper })}
      />
      <NetworkFlowActionGroup>
        <NetworkFlowButton
          data-testid={networkAnalysisTestId("rejected-query-apply")}
          variant="primary"
          onClick={() => {
            focusRequested.current = true;
            if (onApply()) focusRequested.current = false;
          }}
        >
          Apply diagnostics query
        </NetworkFlowButton>
        <NetworkFlowButton
          data-testid={networkAnalysisTestId("rejected-query-clear")}
          onClick={onClear}
        >
          Clear diagnostics query
        </NetworkFlowButton>
      </NetworkFlowActionGroup>
      <details>
        <summary>Applied diagnostic filters</summary>
        <p>
          Error codes: {appliedQuery.errorCodes.join(", ") || "all"}; fields:{" "}
          {appliedQuery.fieldKeys.join(", ") || "all"}; source rows:{" "}
          {appliedQuery.sourceRowRange?.gte ?? "unbounded"} through{" "}
          {appliedQuery.sourceRowRange?.lte ?? "unbounded"}.
        </p>
      </details>
    </section>
  );
}
export function NetworkFlowGraphQueryControls({
  draft,
  applied,
  onChange,
  tables,
  activeTableId,
  issues,
}: {
  readonly draft: GraphQuerySettings;
  readonly applied: GraphQuerySettings;
  readonly onChange: (draft: GraphQuerySettings) => void;
  readonly tables: readonly NetworkFlowTable[];
  readonly activeTableId: string | null;
  readonly issues: readonly QueryIssue[];
}) {
  const scopeError = issues.find((v) => v.path === "scope")?.message;
  const label = (id: string) =>
    tables.find((t) => t.network_flow_table_id === id)?.display_name ??
    "Unavailable table (" + id + ")";
  const appliedScope =
    applied.scopeMode === "all_active_tables"
      ? "All active tables"
      : applied.scopeMode === "selected_tables"
        ? applied.selectedTableIds.map(label).join(", ")
        : label(activeTableId ?? "");
  return (
    <div className="network-flow-graph-query-controls">
      <fieldset data-testid={networkAnalysisTestId("graph-scope")}>
        <legend>Graph scope (Apply to query)</legend>
        {(
          ["active_table", "selected_tables", "all_active_tables"] as const
        ).map((mode) => (
          <label key={mode} htmlFor={"graph-scope-" + mode}>
            <NetworkFlowChoice
              id={"graph-scope-" + mode}
              type="radio"
              name="network-flow-graph-scope"
              checked={draft.scopeMode === mode}
              aria-invalid={!!scopeError}
              aria-describedby={
                scopeError ? "network-flow-scope-error" : undefined
              }
              onChange={() =>
                onChange({
                  ...draft,
                  scopeMode: mode,
                  selectedTableIds:
                    mode === "selected_tables" &&
                    draft.selectedTableIds.length === 0 &&
                    activeTableId
                      ? [activeTableId]
                      : draft.selectedTableIds,
                })
              }
            />
            {mode === "active_table"
              ? "Active table"
              : mode === "selected_tables"
                ? "Selected tables"
                : "All active tables"}
          </label>
        ))}
        {draft.scopeMode === "selected_tables"
          ? [
              ...tables.map((t) => t.network_flow_table_id),
              ...draft.selectedTableIds.filter(
                (id) => !tables.some((t) => t.network_flow_table_id === id),
              ),
            ].map((id) => (
              <label key={id} htmlFor={"graph-table-" + id}>
                <NetworkFlowChoice
                  id={"graph-table-" + id}
                  checked={draft.selectedTableIds.includes(id)}
                  onChange={(e) =>
                    onChange({
                      ...draft,
                      selectedTableIds: e.currentTarget.checked
                        ? [...draft.selectedTableIds, id]
                        : draft.selectedTableIds.filter((v) => v !== id),
                    })
                  }
                />
                {label(id)}
              </label>
            ))
          : null}
        {scopeError ? (
          <span id="network-flow-scope-error">{scopeError}</span>
        ) : null}
      </fieldset>
      <fieldset>
        <legend>Aggregation (Apply to query)</legend>
        <label htmlFor="graph-aggregation-default">
          <NetworkFlowChoice
            id="graph-aggregation-default"
            type="radio"
            name="network-flow-aggregation"
            checked={draft.aggregation.mode === "default_flow_edge_v1"}
            onChange={() =>
              onChange({
                ...draft,
                aggregation: defaultGraphQuerySettings.aggregation,
              })
            }
          />
          Default flow edges
        </label>
        <label htmlFor="graph-aggregation-temporal">
          <NetworkFlowChoice
            id="graph-aggregation-temporal"
            type="radio"
            name="network-flow-aggregation"
            checked={draft.aggregation.mode === "time_bucket_v1"}
            onChange={() =>
              onChange({
                ...draft,
                aggregation: {
                  mode: "time_bucket_v1",
                  bucket_width_seconds: 3600,
                  include_example_row_refs: true,
                },
              })
            }
          />
          Time buckets
        </label>
        {draft.aggregation.mode === "time_bucket_v1" ? (
          <NetworkFlowField
            htmlFor="network-flow-bucket-width"
            label="Bucket width"
          >
            <NetworkFlowSelect
              id="network-flow-bucket-width"
              value={draft.aggregation.bucket_width_seconds}
              onChange={(e) => {
                const width = networkFlowQueryMetadata.bucketWidths.find(
                  (v) => String(v) === e.currentTarget.value,
                );
                if (width)
                  onChange({
                    ...draft,
                    aggregation: {
                      mode: "time_bucket_v1",
                      bucket_width_seconds: width,
                      include_example_row_refs: true,
                    },
                  });
              }}
            >
              {networkFlowQueryMetadata.bucketWidths.map((v) => (
                <option key={v} value={v}>
                  {v === 60
                    ? "1 minute"
                    : v === 300
                      ? "5 minutes"
                      : v === 900
                        ? "15 minutes"
                        : v === 3600
                          ? "1 hour"
                          : v === 21600
                            ? "6 hours"
                            : "1 day"}
                </option>
              ))}
            </NetworkFlowSelect>
          </NetworkFlowField>
        ) : null}
      </fieldset>
      <span>
        Applied graph scope: {appliedScope}. Aggregation:{" "}
        {applied.aggregation.mode === "time_bucket_v1"
          ? applied.aggregation.bucket_width_seconds + " second buckets"
          : "default flow edges"}
        .
      </span>
    </div>
  );
}
