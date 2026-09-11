import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, expect, it } from "vitest";
import {
  observationOwnerFixture,
  observationSource,
} from "../../../testing/observationTestSupport";
import { IndicatorInspectorWorkflow } from "./IndicatorInspectorWorkflow";
import { ObservationContext } from "./ObservationContext";

afterEach(cleanup);
it("Observation capture exposes committed source text for accessible selection", () => {
  const t = observationOwnerFixture();
  render(
    <ObservationContext.Provider value={t.owner}>
      <IndicatorInspectorWorkflow
        action="indicator.observations.manage"
        sourceRecordId={observationSource.recordId}
        source={t.source}
      />
    </ObservationContext.Provider>,
  );
  expect(
    screen
      .getByRole("textbox", { name: "Saved source text" })
      .getAttribute("aria-readonly"),
  ).toBe("true");
  expect(
    screen.queryByRole("spinbutton", { name: "Span start byte" }),
  ).toBeNull();
});
