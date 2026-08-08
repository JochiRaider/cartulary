import { describe, expect, it } from "vitest";
import {
  cartularyDesignPresentation,
  cartularyErrorPresentation,
} from "./designPresentation";

describe("design presentation projection", () => {
  it("exports the adopted loading and transient timing rules", () => {
    expect(cartularyDesignPresentation.initialLoading).toMatchObject({
      announceOncePerGeneration: true,
      delayMs: 2000,
      message: "Still loading this surface",
      retryOnDelay: false,
    });
    expect(cartularyDesignPresentation.transientConfirmation).toMatchObject({
      resumeResetsElapsed: false,
      stillValidActionPreventsDismissal: true,
      visibleUnpausedMs: 5000,
    });
  });

  it("contains every closed error family exactly once", () => {
    const families = cartularyDesignPresentation.errorPresentations.map(
      (presentation) => presentation.family,
    );
    expect(new Set(families).size).toBe(11);
    expect(cartularyErrorPresentation("unknown_future_error")).toMatchObject({
      actions: [],
      live: "assertive",
    });
  });
});
