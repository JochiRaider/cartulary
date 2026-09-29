import {
  getReferenceFieldContract,
  requireViewContract,
  type ViewContract,
} from "@cartulary/view-contracts";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import { GenericInspectorReferenceSummary } from "./GenericInspectorReferenceSummary";

afterEach(cleanup);

function row(cells: WorkbookQueryRow["cells"]): WorkbookQueryRow {
  return { record_id: "saved-reference-subject", row_version: 1, cells };
}

function referenceLabels(contract: ViewContract, evidenceOnly = false) {
  return contract.fields.flatMap((field) => {
    const reference = getReferenceFieldContract(
      contract.viewSchemaId,
      field.fieldKey,
    );
    return reference &&
      (!evidenceOnly || reference.targetRecordTypes.includes("evidence"))
      ? [field.label]
      : [];
  });
}

describe("Generic inspector reference summary", () => {
  it("reads accepted direct Party values and opens only the selected ordinary field", () => {
    const contract = requireViewContract("cartulary.view.evidence.v1");
    const onEdit = vi.fn();
    render(
      <GenericInspectorReferenceSummary
        contract={contract}
        row={row({
          "evidence.collector_party_id": { value: "Accepted collector" },
          "evidence.source_party_id": { value: null },
        })}
        canEdit
        onEdit={onEdit}
      />,
    );

    expect(screen.getAllByRole("term").map((term) => term.textContent)).toEqual(
      referenceLabels(contract),
    );
    expect(
      screen.getAllByRole("definition").map((definition) => definition.textContent),
    ).toEqual([
      "Accepted collector",
      "Edit Collector Party",
      "None",
      "Edit Source Party",
    ]);
    const collector = screen.getByRole("button", {
      name: "Edit Collector Party",
    });
    expect(collector.getAttribute("data-inspector-action-tone")).toBe("quiet");
    fireEvent.click(collector);
    expect(onEdit).toHaveBeenCalledExactlyOnceWith(
      "evidence.collector_party_id",
    );
  });

  it("keeps declared order, collection values, and unavailable projections distinct in normal and Evidence summaries", () => {
    const contract = requireViewContract("cartulary.view.lesson.v1");
    const onEdit = vi.fn();
    const accepted = "A long <script>literal</script> evidence title ".repeat(4);
    const saved = row({
      "lesson.follow_up_task_ids": { value: { items: [] } },
      "lesson.evidence_refs": {
        value: { items: [{ display_text: accepted }] },
      },
    });
    const { rerender, container } = render(
      <GenericInspectorReferenceSummary
        contract={contract}
        row={saved}
        canEdit
        onEdit={onEdit}
      />,
    );

    expect(screen.getAllByRole("term").map((term) => term.textContent)).toEqual(
      referenceLabels(contract),
    );
    expect(screen.getByText(accepted)).toBeTruthy();
    expect(container.querySelector("script")).toBeNull();
    expect(screen.getByText("None")).toBeTruthy();
    expect(screen.getByText("Not available in this row.")).toBeTruthy();
    expect(screen.getByRole("button", { name: "Manage Evidence" })).not.toHaveProperty(
      "disabled",
      true,
    );
    expect(
      screen.getByRole("button", { name: "Manage Follow-up Tasks" }),
    ).not.toHaveProperty("disabled", true);
    expect(screen.getByRole("button", { name: "Edit Owner" })).toHaveProperty(
      "disabled",
      true,
    );

    rerender(
      <GenericInspectorReferenceSummary
        contract={contract}
        row={saved}
        evidenceOnly
        canEdit
        onEdit={onEdit}
      />,
    );
    expect(screen.getAllByRole("term").map((term) => term.textContent)).toEqual(
      referenceLabels(contract, true),
    );
    fireEvent.click(screen.getByRole("button", { name: "Manage Evidence" }));
    expect(onEdit).toHaveBeenCalledExactlyOnceWith("lesson.evidence_refs");
  });

  it("keeps truthful no-field messages and non-authoring viewer actions", () => {
    const onEdit = vi.fn();
    const evidence = requireViewContract("cartulary.view.evidence.v1");
    const { rerender } = render(
      <GenericInspectorReferenceSummary
        contract={evidence}
        row={row({ "evidence.collector_party_id": { value: "Collector" } })}
        canEdit={false}
        onEdit={onEdit}
      />,
    );
    expect(screen.getByRole("button", { name: "Edit Collector Party" })).toHaveProperty(
      "disabled",
      true,
    );
    fireEvent.click(screen.getByRole("button", { name: "Edit Collector Party" }));
    expect(onEdit).not.toHaveBeenCalled();

    rerender(
      <GenericInspectorReferenceSummary
        contract={evidence}
        row={row({})}
        evidenceOnly
        canEdit={false}
        onEdit={onEdit}
      />,
    );
    expect(
      screen.getByText(/This record has no evidence reference field/u),
    ).toBeTruthy();

    rerender(
      <GenericInspectorReferenceSummary
        contract={requireViewContract("cartulary.view.hosts.v1")}
        row={row({})}
        canEdit={false}
        onEdit={onEdit}
      />,
    );
    expect(screen.getByText(/This record has no reference fields/u)).toBeTruthy();
  });
});
