import { requireViewContract } from "@cartulary/view-contracts";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { useReducer } from "react";
import { afterEach, expect, it, vi } from "vitest";
import { testObservation } from "../../../testing/observationTestSupport";
import { IndicatorCanonicalAuthoring } from "./IndicatorCanonicalAuthoring";
import { IndicatorCreateDraftStore } from "./IndicatorCreateDraftStore";
import {
  changeIndicatorCreateValue,
  indicatorCreateErrors,
  indicatorCreateRequest,
  indicatorCreateSeed,
} from "./indicatorCreateModel";

afterEach(cleanup);
const contract = requireViewContract("cartulary.view.indicators.v1");
it("Indicator canonical authoring preserves observed text and requires an explicit value kind", () => {
  const observation = {
    ...testObservation,
    parsed_indicator_type: "url" as const,
    normalized_candidate: "https://example.test/",
  };
  const submit = vi.fn();
  let store: IndicatorCreateDraftStore;
  function Editor() {
    const [, redraw] = useReducer((value) => value + 1, 0);
    store ??= new IndicatorCreateDraftStore(redraw);
    return (
      <IndicatorCanonicalAuthoring
        observation={observation}
        contract={contract}
        draft={store.ensure(observation)}
        disabled={false}
        onChange={(key, value) =>
          store.update(testObservation.observation_id, key, value)
        }
        onSubmit={submit}
      />
    );
  }
  render(<Editor />);
  expect(
    (screen.getByLabelText("Canonical value") as HTMLInputElement).value,
  ).toBe(observation.normalized_candidate);
  fireEvent.click(
    screen.getByRole("button", { name: "Create canonical Indicator" }),
  );
  expect(submit).not.toHaveBeenCalled();
  expect(screen.getByLabelText("Value kind").getAttribute("aria-invalid")).toBe(
    "true",
  );
  fireEvent.change(screen.getByLabelText("Value kind"), {
    target: { value: "atomic" },
  });
  fireEvent.click(
    screen.getByRole("button", { name: "Create canonical Indicator" }),
  );
  expect(submit).toHaveBeenCalledOnce();
  fireEvent.change(screen.getByLabelText("Indicator type"), {
    target: { value: "ipv4_addr" },
  });
  expect(
    (screen.getByLabelText("Canonical value") as HTMLInputElement).value,
  ).toBe("");
  expect((screen.getByLabelText("Value kind") as HTMLSelectElement).value).toBe(
    "atomic",
  );
  expect(screen.queryByLabelText("Hash value")).toBeNull();
  expect(
    screen.getByText(testObservation.observed_text, { selector: "strong" }),
  ).toBeTruthy();
});
it("Indicator canonical type changes invalidate derived details without rewriting a retained draft", () => {
  const store = new IndicatorCreateDraftStore();
  const initial = store.ensure(testObservation);
  for (const key of [
    "indicator.hash_algorithm",
    "indicator.hash_value",
    "indicator.normalized_value",
    "indicator.defanged_value",
    "indicator.stix_pattern",
  ])
    store.update(initial.observationId, key, "old");
  const prior = store.get(initial.observationId);
  store.update(initial.observationId, "indicator.indicator_type", "text");
  expect(store.get(initial.observationId)?.values).toEqual({
    "indicator.indicator_type": "text",
    "indicator.value_kind": "",
  });
  expect(prior?.values["indicator.stix_pattern"]).toBe("old");
  expect(
    changeIndicatorCreateValue(
      initial.values,
      "indicator.indicator_type",
      "ipv6_addr",
    )["indicator.value_kind"],
  ).toBe("atomic");
  expect(
    indicatorCreateSeed({
      ...testObservation,
      parsed_indicator_type: null,
      normalized_candidate: null,
    }),
  ).toEqual({
    "indicator.indicator_type": "",
    "indicator.value_kind": "",
    "indicator.display_value": "",
  });
  store.clear();
  expect(store.get(initial.observationId)).toBeUndefined();
});

it("Indicator requests preserve raw values for the shared owner algorithms", () => {
  const raw = "\uFEFF A\r\nB\tC\rD ";
  const values = {
    "indicator.indicator_type": "text",
    "indicator.value_kind": "atomic",
    "indicator.display_value": raw,
  };
  expect(indicatorCreateRequest(contract, values, "raw-text")).toEqual({
    client_txn_id: "raw-text",
    ...values,
  });
  expect(
    indicatorCreateErrors(contract, {
      ...values,
      "indicator.indicator_type": "url",
      "indicator.value_kind": "pattern",
    }),
  ).toHaveProperty("indicator.value_kind");
  expect(
    indicatorCreateErrors(contract, {
      ...values,
      "indicator.display_value": "😀".repeat(8192),
    }),
  ).toEqual({});
  expect(
    indicatorCreateErrors(contract, {
      ...values,
      "indicator.display_value": "😀".repeat(8193),
    }),
  ).toHaveProperty("indicator.display_value");
});
