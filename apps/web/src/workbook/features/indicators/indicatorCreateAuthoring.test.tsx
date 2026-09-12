import { requireViewContract } from "@cartulary/view-contracts";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { useReducer } from "react";
import { afterEach, expect, it, vi } from "vitest";
import { testObservation } from "../../../testing/observationTestSupport";
import { IndicatorCanonicalAuthoring } from "./IndicatorCanonicalAuthoring";
import { IndicatorCreateDraftStore } from "./IndicatorCreateDraftStore";
import {
  changeIndicatorCreateValue,
  indicatorCreateSeed,
} from "./indicatorCreateModel";

afterEach(cleanup);
const contract = requireViewContract("cartulary.view.indicators.v1");
it("Indicator canonical authoring preserves observed text and requires an explicit value kind", () => {
  const submit = vi.fn();
  let store: IndicatorCreateDraftStore;
  function Editor() {
    const [, redraw] = useReducer((value) => value + 1, 0);
    store ??= new IndicatorCreateDraftStore(redraw);
    return (
      <IndicatorCanonicalAuthoring
        observation={testObservation}
        contract={contract}
        draft={store.ensure(testObservation)}
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
  ).toBe(testObservation.normalized_candidate);
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
