import { requireViewContract } from "@cartulary/view-contracts";
import { vi } from "vitest";
import { createIndicatorCreateTransport } from "../workbook/adapters/createIndicatorCreateTransport";
import type { CreateViewRowResponse } from "../workbook/adapters/indicatorCreateProtocol";
import type { IndicatorCreateTransport } from "../workbook/features/indicators/indicatorCreateOperation";
import { WorkbookIndicatorCreateOwner } from "../workbook/features/indicators/WorkbookIndicatorCreateOwner";
import { normalizeWorkbookViewRows } from "../workbook/models/workbookContractRows";
import {
  observationAuthority,
  observationReaderFixture,
  observationTargetId,
  testObservation,
} from "./observationTestSupport";

export const canonicalCreateContract = requireViewContract(
  "cartulary.view.indicators.v1",
);
export const canonicalCreateValues = {
  "indicator.indicator_type": "domain_name",
  "indicator.value_kind": "atomic",
  "indicator.display_value": "EXAMPLE[.]COM",
};
export function canonicalCreateResponse(): CreateViewRowResponse {
  const cells = Object.fromEntries(
    canonicalCreateContract.fields.map((field) => [
      field.fieldKey,
      { value: field.readKind === "number" ? 0 : null },
    ]),
  ) as Record<string, { value: string | number | null }>;
  cells["indicator.indicator_type"] = { value: "domain_name" };
  cells["indicator.value_kind"] = { value: "atomic" };
  cells["indicator.display_value"] = { value: "example.com" };
  cells["indicator.normalized_value"] = { value: "example.com" };
  return {
    data: {
      view_schema_id: "cartulary.view.indicators.v1",
      change_set_id: "60000000-0000-4000-8000-000000000021",
      row: {
        record_id: observationTargetId,
        row_version: 1,
        cells,
        group_values: {
          "indicator.indicator_type": "domain_name",
          "indicator.value_kind": "atomic",
          "indicator.lifecycle_summary": null,
        },
      },
    },
    meta: { request_id: "canonical-create" },
  };
}
export function canonicalCreateFixture() {
  const reader = observationReaderFixture(),
    response = canonicalCreateResponse();
  const [row] = normalizeWorkbookViewRows(
    canonicalCreateContract,
    [response.data.row],
    "canonical create fixture",
  );
  if (!row) throw new Error("Canonical create fixture row missing");
  const receipt = {
    status: 201 as const,
    response,
    row,
  };
  let sequence = 0;
  const ids = { create: vi.fn(() => `canonical-secure-${++sequence}`) },
    accepted = vi.fn();
  const transport = {
    ...createIndicatorCreateTransport({
      apiBase: "https://original.test",
      incidentId: observationAuthority.incidentId,
    }),
    send: vi.fn<IndicatorCreateTransport["send"]>(async () => ({
      kind: "accepted",
      receipt,
    })),
  };
  const owner = new WorkbookIndicatorCreateOwner(
      observationAuthority.incidentId,
      ids,
      accepted,
    ),
    reconcile = vi.fn(async () => {});
  owner.configure(reader, transport);
  owner.setAuthority(observationAuthority);
  owner.registerReconciliation(reconcile);
  const binding = {
    isCurrent: vi.fn(() => true),
    matchesDraft: vi.fn(() => true),
  };
  const admit = () => {
    const result = owner.admit(
      testObservation,
      canonicalCreateContract,
      canonicalCreateValues,
      binding,
    );
    if (result.kind !== "admitted")
      throw new Error(`Unexpected admission: ${result.kind}`);
    return result.attempt;
  };
  return {
    owner,
    reader,
    transport,
    ids,
    accepted,
    reconcile,
    binding,
    receipt,
    response,
    admit,
    entry: () => owner.getSnapshot().entries.at(-1),
  };
}
