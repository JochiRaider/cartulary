import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { NetworkFlowAcceptedQueryControls } from "./NetworkFlowQueryControls";
import {
  compileAcceptedDraft,
  type NetworkFlowAcceptedQuery,
  reconstructAcceptedQuery,
} from "./networkFlowQueryModel";

describe("Network Flow query authoring integrity", () => {
  it("retains protocol lists and complete counter predicates on unchanged Apply", () => {
    const query: NetworkFlowAcceptedQuery = {
      filters: [
        { field_key: "network_flow.ip_protocol", op: "in", value: [6, 17] },
        {
          field_key: "network_flow.bytes_count",
          op: "eq",
          value: "18446744073709551615",
        },
        {
          field_key: "network_flow.packets_count",
          op: "range",
          value: { gte: null, lte: "42" },
        },
      ],
      sort: [],
      timeWindow: null,
    };
    const onChange = vi.fn();
    const draft = reconstructAcceptedQuery(query);
    render(
      <NetworkFlowAcceptedQueryControls
        graphMode={false}
        tableLabel="Table A"
        draft={draft}
        onDraftChange={vi.fn()}
        appliedQuery={query}
        issues={[]}
        status="applied"
        onClear={vi.fn()}
        onApply={() => {
          const result = compileAcceptedDraft(draft);
          if (result.ok) onChange(result.value);
          return result.ok;
        }}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Apply query" }));
    expect(onChange).toHaveBeenCalledWith(query);
  });
});
