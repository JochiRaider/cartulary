import { describe, expect, it } from "vitest";
import {
  cartularyDesignPresentation,
  cartularyErrorPresentation,
  cartularyGridDataStatePresentation,
  cartularyGridInteractionModePresentation,
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
    expect(cartularyErrorPresentation("client_txn_conflict")).toMatchObject({
      actions: ["retry_with_new_request_id", "discard_blocked_edit"],
      live: "polite",
    });
  });

  it("contains every closed grid data state exactly once", () => {
    const states = cartularyDesignPresentation.gridDataStatePresentations.map(
      (presentation) => presentation.state,
    );

    expect(states).toEqual([
      "ready",
      "initial_loading",
      "refreshing",
      "empty",
      "filtered_empty",
      "stale_error",
      "unavailable",
      "permission_denied",
    ]);
    expect(new Set(states).size).toBe(8);
    expect(cartularyGridDataStatePresentation("stale_error")).toMatchObject({
      blocking: false,
      live: "assertive",
      rowRetention: "retain_previously_authorized",
    });
    expect(
      cartularyGridDataStatePresentation("permission_denied"),
    ).toMatchObject({
      actionRule: "none",
      draftRetention: "clear_protected_draft",
      rowRetention: "show_none",
    });
  });

  it("projects interaction state independently with one composition rule", () => {
    expect(
      cartularyDesignPresentation.gridInteractionModePresentations.map(
        (presentation) => presentation.mode,
      ),
    ).toEqual(["editable", "read_only"]);
    expect(cartularyGridInteractionModePresentation("read_only")).toMatchObject(
      {
        messageStrategy: "owner_label",
        visible: true,
      },
    );
    expect(cartularyDesignPresentation.gridStateComposition).toEqual({
      coDisplayInteractionMode: "read_only",
      liveRegionRule: "single_highest_priority_atomic_message",
      primary: "data_state",
      suppressInteractionForDataStates: ["permission_denied"],
    });
  });
});
