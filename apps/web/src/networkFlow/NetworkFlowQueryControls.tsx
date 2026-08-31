import { networkAnalysisTestId } from "@cartulary/ui-contracts";
import { useEffect, useMemo, useState } from "react";
import type {
  NetworkFlowFilter,
  NetworkFlowRejectedRowsQueryRequest,
} from "../services/networkFlowContractAdapter";
import {
  NetworkFlowActionGroup,
  NetworkFlowButton,
  NetworkFlowField,
  NetworkFlowNumberInput,
  NetworkFlowSelect,
  NetworkFlowTextInput,
} from "./NetworkFlowControls";
import {
  networkFlowColumnLabel,
  networkFlowPresentationColumns,
} from "./networkFlowPresentation";
import {
  emptyNetworkFlowAcceptedQuery,
  emptyNetworkFlowRejectedQuery,
  type NetworkFlowAcceptedQuery,
  type NetworkFlowRejectedQuery,
} from "./networkFlowQueryModel";

export function NetworkFlowAcceptedQueryControls({
  graphMode,
  onChange,
  query,
}: {
  readonly graphMode: boolean;
  readonly onChange: (query: NetworkFlowAcceptedQuery) => void;
  readonly query: NetworkFlowAcceptedQuery;
}) {
  const [startUTC, setStartUTC] = useState(query.timeWindow?.startUTC ?? "");
  const [endUTC, setEndUTC] = useState(query.timeWindow?.endUTC ?? "");
  const [endpoint, setEndpoint] = useState("");
  const [endpointOperator, setEndpointOperator] = useState<
    "eq" | "cidr_contains"
  >("eq");
  const [protocol, setProtocol] = useState("");
  const [bytesMinimum, setBytesMinimum] = useState("");
  const [packetsMinimum, setPacketsMinimum] = useState("");
  const [advanced, setAdvanced] = useState<readonly NetworkFlowFilter[]>([]);
  const [advancedField, setAdvancedField] = useState("network_flow.src_ip");
  const [advancedOperator, setAdvancedOperator] = useState("eq");
  const [advancedValue, setAdvancedValue] = useState("");
  const filterableColumns = useMemo(
    () =>
      networkFlowPresentationColumns("network_flow.accepted_rows.v1").filter(
        (column) =>
          !column.inspector_only && column.filter_operators.length > 0,
      ),
    [],
  );
  const selectedMetadata = filterableColumns.find(
    (column) => column.field_key === advancedField,
  );

  useEffect(() => {
    const endpointFilter = query.filters.find(
      (filter) => filter.field_key === "network_flow.endpoint_ip",
    );
    const protocolFilter = query.filters.find(
      (filter) => filter.field_key === "network_flow.ip_protocol",
    );
    const bytesFilter = query.filters.find(
      (filter) => filter.field_key === "network_flow.bytes_count",
    );
    const packetsFilter = query.filters.find(
      (filter) => filter.field_key === "network_flow.packets_count",
    );
    setStartUTC(query.timeWindow?.startUTC ?? "");
    setEndUTC(query.timeWindow?.endUTC ?? "");
    setEndpoint(filterScalarText(endpointFilter));
    setEndpointOperator(
      endpointFilter?.op === "cidr_contains" ? "cidr_contains" : "eq",
    );
    setProtocol(filterScalarText(protocolFilter));
    setBytesMinimum(filterRangeBoundText(bytesFilter, "gte"));
    setPacketsMinimum(filterRangeBoundText(packetsFilter, "gte"));
    setAdvanced(
      query.filters.filter(
        (filter) =>
          filter !== endpointFilter &&
          filter !== protocolFilter &&
          filter !== bytesFilter &&
          filter !== packetsFilter,
      ),
    );
  }, [query.filters, query.timeWindow]);

  const apply = () => {
    const filters: NetworkFlowFilter[] = [...advanced];
    if (endpoint.trim() !== "") {
      filters.push({
        field_key: "network_flow.endpoint_ip",
        op: endpointOperator,
        value: endpoint.trim(),
      });
    }
    if (protocol.trim() !== "") {
      filters.push({
        field_key: "network_flow.ip_protocol",
        op: "eq",
        value: Number(protocol),
      });
    }
    if (bytesMinimum.trim() !== "") {
      filters.push({
        field_key: "network_flow.bytes_count",
        op: "range",
        value: { gte: bytesMinimum.trim(), lte: null },
      });
    }
    if (packetsMinimum.trim() !== "") {
      filters.push({
        field_key: "network_flow.packets_count",
        op: "range",
        value: { gte: packetsMinimum.trim(), lte: null },
      });
    }
    onChange({
      filters,
      sort: query.sort,
      timeWindow:
        startUTC.trim() === "" && endUTC.trim() === ""
          ? null
          : {
              startUTC: startUTC.trim() === "" ? null : startUTC.trim(),
              endUTC: endUTC.trim() === "" ? null : endUTC.trim(),
            },
    });
  };

  const reset = () => {
    setStartUTC("");
    setEndUTC("");
    setEndpoint("");
    setProtocol("");
    setBytesMinimum("");
    setPacketsMinimum("");
    setAdvanced([]);
    setAdvancedValue("");
    onChange({ ...emptyNetworkFlowAcceptedQuery, sort: [] });
  };

  return (
    <section
      aria-label="Network Flow filters"
      className="network-flow-query-band"
      data-testid={networkAnalysisTestId("filters")}
    >
      <NetworkFlowField
        htmlFor="network-flow-query-table-scope"
        label="Table scope"
      >
        <NetworkFlowSelect
          defaultValue="active_table"
          disabled={!graphMode}
          id="network-flow-query-table-scope"
        >
          <option value="active_table">Active table</option>
          <option value="selected_tables">Selected tables</option>
          <option value="all_active_tables">All active tables</option>
        </NetworkFlowSelect>
      </NetworkFlowField>
      <NetworkFlowField
        htmlFor="network-flow-query-start"
        label="Flow overlap starts at"
      >
        <NetworkFlowTextInput
          id="network-flow-query-start"
          placeholder="2026-07-16T00:00:00Z"
          value={startUTC}
          onChange={(event) => setStartUTC(event.currentTarget.value)}
        />
      </NetworkFlowField>
      <NetworkFlowField
        htmlFor="network-flow-query-end"
        label="Flow overlap ends before"
      >
        <NetworkFlowTextInput
          id="network-flow-query-end"
          placeholder="2026-07-17T00:00:00Z"
          value={endUTC}
          onChange={(event) => setEndUTC(event.currentTarget.value)}
        />
      </NetworkFlowField>
      <NetworkFlowField
        htmlFor="network-flow-query-endpoint"
        label="Endpoint IP"
      >
        <span className="network-flow-inline-fields">
          <NetworkFlowSelect
            aria-label="Endpoint IP operator"
            value={endpointOperator}
            onChange={(event) =>
              setEndpointOperator(
                event.currentTarget.value as "eq" | "cidr_contains",
              )
            }
          >
            <option value="eq">equals</option>
            <option value="cidr_contains">in CIDR</option>
          </NetworkFlowSelect>
          <NetworkFlowTextInput
            aria-label="Endpoint IP value"
            id="network-flow-query-endpoint"
            value={endpoint}
            onChange={(event) => setEndpoint(event.currentTarget.value)}
          />
        </span>
      </NetworkFlowField>
      <NetworkFlowField htmlFor="network-flow-query-protocol" label="Protocol">
        <NetworkFlowNumberInput
          id="network-flow-query-protocol"
          min={0}
          max={255}
          value={protocol}
          onChange={(event) => setProtocol(event.currentTarget.value)}
        />
      </NetworkFlowField>
      <NetworkFlowField
        htmlFor="network-flow-query-minimum-bytes"
        label="Minimum bytes"
      >
        <NetworkFlowTextInput
          id="network-flow-query-minimum-bytes"
          inputMode="numeric"
          value={bytesMinimum}
          onChange={(event) => setBytesMinimum(event.currentTarget.value)}
        />
      </NetworkFlowField>
      <NetworkFlowField
        htmlFor="network-flow-query-minimum-packets"
        label="Minimum packets"
      >
        <NetworkFlowTextInput
          id="network-flow-query-minimum-packets"
          inputMode="numeric"
          value={packetsMinimum}
          onChange={(event) => setPacketsMinimum(event.currentTarget.value)}
        />
      </NetworkFlowField>
      <details
        className="network-flow-advanced"
        data-testid={networkAnalysisTestId("advanced-filters")}
      >
        <summary>Advanced field filters ({advanced.length})</summary>
        <div className="network-flow-popover network-flow-advanced__editor">
          <NetworkFlowField
            htmlFor="network-flow-query-advanced-field"
            label="Field"
          >
            <NetworkFlowSelect
              id="network-flow-query-advanced-field"
              value={advancedField}
              onChange={(event) => {
                setAdvancedField(event.currentTarget.value);
                setAdvancedOperator("eq");
              }}
            >
              {filterableColumns.map((column) => (
                <option key={column.field_key} value={column.field_key}>
                  {networkFlowColumnLabel(column.label_key)}
                </option>
              ))}
            </NetworkFlowSelect>
          </NetworkFlowField>
          <NetworkFlowField
            htmlFor="network-flow-query-advanced-operator"
            label="Operator"
          >
            <NetworkFlowSelect
              id="network-flow-query-advanced-operator"
              value={advancedOperator}
              onChange={(event) =>
                setAdvancedOperator(event.currentTarget.value)
              }
            >
              {(selectedMetadata?.filter_operators ?? ["eq"]).map(
                (operator) => (
                  <option key={operator} value={operator}>
                    {operator}
                  </option>
                ),
              )}
            </NetworkFlowSelect>
          </NetworkFlowField>
          <NetworkFlowField
            htmlFor="network-flow-query-advanced-value"
            label="Value"
          >
            <NetworkFlowTextInput
              id="network-flow-query-advanced-value"
              disabled={
                advancedOperator === "is_null" ||
                advancedOperator === "not_null"
              }
              value={advancedValue}
              onChange={(event) => setAdvancedValue(event.currentTarget.value)}
            />
          </NetworkFlowField>
          <NetworkFlowButton
            variant="secondary"
            onClick={() => {
              const filter = advancedFilter(
                advancedField,
                advancedOperator,
                advancedValue,
              );
              if (filter !== null) {
                setAdvanced((current) =>
                  current.some(
                    (candidate) =>
                      JSON.stringify(candidate) === JSON.stringify(filter),
                  )
                    ? current
                    : [...current, filter],
                );
                setAdvancedValue("");
              }
            }}
          >
            Add filter
          </NetworkFlowButton>
          {advanced.map((filter) => (
            <NetworkFlowButton
              className="network-flow-filter-chip"
              key={JSON.stringify(filter)}
              variant="ghost"
              onClick={() =>
                setAdvanced((current) =>
                  current.filter((candidate) => candidate !== filter),
                )
              }
            >
              Remove {filter.field_key} {filter.op}
            </NetworkFlowButton>
          ))}
        </div>
      </details>
      <NetworkFlowActionGroup>
        <NetworkFlowButton
          data-testid={networkAnalysisTestId("accepted-query-apply")}
          variant="primary"
          onClick={apply}
        >
          Apply query
        </NetworkFlowButton>
        <NetworkFlowButton
          data-testid={networkAnalysisTestId("accepted-query-clear")}
          variant="secondary"
          onClick={reset}
        >
          Clear query
        </NetworkFlowButton>
      </NetworkFlowActionGroup>
    </section>
  );
}

export function NetworkFlowRejectedQueryControls({
  onChange,
  query,
}: {
  readonly onChange: (query: NetworkFlowRejectedQuery) => void;
  readonly query: NetworkFlowRejectedQuery;
}) {
  const [errorCodes, setErrorCodes] = useState(query.errorCodes.join(", "));
  const [fieldKey, setFieldKey] = useState(query.fieldKeys[0] ?? "");
  const [rowStart, setRowStart] = useState(
    query.sourceRowRange?.gte?.toString() ?? "",
  );
  const [rowEnd, setRowEnd] = useState(
    query.sourceRowRange?.lte?.toString() ?? "",
  );
  useEffect(() => {
    setErrorCodes(query.errorCodes.join(", "));
    setFieldKey(query.fieldKeys[0] ?? "");
    setRowStart(query.sourceRowRange?.gte?.toString() ?? "");
    setRowEnd(query.sourceRowRange?.lte?.toString() ?? "");
  }, [query.errorCodes, query.fieldKeys, query.sourceRowRange]);
  const fieldOptions = networkFlowPresentationColumns(
    "network_flow.accepted_rows.v1",
  ).filter(
    (column) => !column.inspector_only && column.filter_operators.length > 0,
  );
  const apply = () => {
    const nextFieldKeys = fieldKey === "" ? [] : [fieldKey];
    onChange({
      errorCodes: errorCodes
        .split(",")
        .map((value) => value.trim())
        .filter((value) => value !== ""),
      fieldKeys:
        nextFieldKeys as NetworkFlowRejectedRowsQueryRequest["field_keys"] &
          readonly string[],
      sourceRowRange:
        rowStart === "" && rowEnd === ""
          ? null
          : {
              gte: rowStart === "" ? null : Number(rowStart),
              lte: rowEnd === "" ? null : Number(rowEnd),
            },
    });
  };
  return (
    <section
      aria-label="Diagnostic filters"
      className="network-flow-query-band"
    >
      <NetworkFlowField
        htmlFor="network-flow-diagnostic-error-codes"
        label="Error codes"
      >
        <NetworkFlowTextInput
          id="network-flow-diagnostic-error-codes"
          placeholder="Comma-separated codes"
          value={errorCodes}
          onChange={(event) => setErrorCodes(event.currentTarget.value)}
        />
      </NetworkFlowField>
      <NetworkFlowField
        htmlFor="network-flow-diagnostic-field-key"
        label="Field key"
      >
        <NetworkFlowTextInput
          id="network-flow-diagnostic-field-key"
          list="network-flow-diagnostic-field-keys"
          value={fieldKey}
          onChange={(event) => setFieldKey(event.currentTarget.value)}
        />
        <datalist id="network-flow-diagnostic-field-keys">
          {fieldOptions.map((column) => (
            <option key={column.field_key} value={column.field_key} />
          ))}
        </datalist>
      </NetworkFlowField>
      <NetworkFlowField
        htmlFor="network-flow-diagnostic-row-start"
        label="First source row"
      >
        <NetworkFlowNumberInput
          id="network-flow-diagnostic-row-start"
          min={1}
          value={rowStart}
          onChange={(event) => setRowStart(event.currentTarget.value)}
        />
      </NetworkFlowField>
      <NetworkFlowField
        htmlFor="network-flow-diagnostic-row-end"
        label="Last source row"
      >
        <NetworkFlowNumberInput
          id="network-flow-diagnostic-row-end"
          min={1}
          value={rowEnd}
          onChange={(event) => setRowEnd(event.currentTarget.value)}
        />
      </NetworkFlowField>
      <NetworkFlowActionGroup>
        <NetworkFlowButton
          data-testid={networkAnalysisTestId("rejected-query-apply")}
          variant="primary"
          onClick={apply}
        >
          Apply diagnostics query
        </NetworkFlowButton>
        <NetworkFlowButton
          data-testid={networkAnalysisTestId("rejected-query-clear")}
          variant="secondary"
          onClick={() => {
            setErrorCodes("");
            setFieldKey("");
            setRowStart("");
            setRowEnd("");
            onChange(emptyNetworkFlowRejectedQuery);
          }}
        >
          Clear diagnostics query
        </NetworkFlowButton>
      </NetworkFlowActionGroup>
    </section>
  );
}

function advancedFilter(
  fieldKey: string,
  operator: string,
  rawValue: string,
): NetworkFlowFilter | null {
  const field_key = fieldKey as NetworkFlowFilter["field_key"];
  if (operator === "is_null" || operator === "not_null") {
    return { field_key, op: operator };
  }
  const value = rawValue.trim();
  if (value === "") {
    return null;
  }
  if (operator === "in") {
    return {
      field_key,
      op: "in",
      value: value.split(",").map((entry) => semanticScalar(fieldKey, entry)),
    };
  }
  if (operator === "gte" || operator === "lte" || operator === "between") {
    const [first, second] = value.split(",", 2);
    const timestamp = fieldKey.endsWith("_utc");
    return {
      field_key,
      op: "range",
      value: timestamp
        ? {
            gte: operator === "lte" ? null : first,
            lt: operator === "gte" ? null : (second ?? first),
          }
        : {
            gte:
              operator === "lte" ? null : semanticScalar(fieldKey, first ?? ""),
            lte:
              operator === "gte"
                ? null
                : semanticScalar(fieldKey, second ?? first ?? ""),
          },
    };
  }
  return {
    field_key,
    op: operator as NetworkFlowFilter["op"],
    value: semanticScalar(fieldKey, value),
  };
}

function semanticScalar(fieldKey: string, value: string): string | number {
  return fieldKey.endsWith("_port") ||
    fieldKey === "network_flow.ip_protocol" ||
    fieldKey === "source_row_number"
    ? Number(value.trim())
    : value.trim();
}

function filterScalarText(filter: NetworkFlowFilter | undefined): string {
  const value = filter?.value;
  return typeof value === "string" || typeof value === "number"
    ? String(value)
    : "";
}

function filterRangeBoundText(
  filter: NetworkFlowFilter | undefined,
  bound: "gte" | "lte" | "lt",
): string {
  const value = filter?.value;
  if (value === null || typeof value !== "object" || Array.isArray(value)) {
    return "";
  }
  const candidate = (value as Record<string, unknown>)[bound];
  return typeof candidate === "string" || typeof candidate === "number"
    ? String(candidate)
    : "";
}
