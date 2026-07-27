import { renderHook } from "@testing-library/react";
import { type ReactNode, StrictMode } from "react";
import { describe, expect, it, vi } from "vitest";
import { useInspectorLifecycleReset } from "./useInspectorLifecycleReset";

describe("useInspectorLifecycleReset", () => {
  it("does not replay the initial lifecycle reset or reset on same-key rerenders", () => {
    const reset = vi.fn();
    const { rerender } = renderHook(
      ({ resetKey }) => {
        useInspectorLifecycleReset(resetKey, reset);
      },
      {
        initialProps: {
          resetKey: "cartulary.view.timeline.v2:base",
        },
      },
    );

    expect(reset).not.toHaveBeenCalled();
    rerender({ resetKey: "cartulary.view.timeline.v2:base" });
    expect(reset).not.toHaveBeenCalled();
  });

  it("uses the latest callback and resets once before a changed lifecycle is observable", () => {
    const initialReset = vi.fn();
    const latestReset = vi.fn();
    const { rerender } = renderHook(
      ({ reset, resetKey }) => {
        useInspectorLifecycleReset(resetKey, reset);
      },
      {
        initialProps: {
          reset: initialReset,
          resetKey: "cartulary.view.hosts.v1:base",
        },
      },
    );

    rerender({
      reset: latestReset,
      resetKey: "cartulary.view.hosts.v1:saved-view",
    });

    expect(initialReset).not.toHaveBeenCalled();
    expect(latestReset).toHaveBeenCalledTimes(1);
  });

  it("treats empty and undefined reset keys as disabled lifecycles", () => {
    const reset = vi.fn();
    const { rerender } = renderHook(
      ({ resetKey }: { resetKey: string | undefined }) => {
        useInspectorLifecycleReset(resetKey, reset);
      },
      {
        initialProps: {
          resetKey: "",
        } as { resetKey: string | undefined },
      },
    );

    rerender({ resetKey: undefined });
    expect(reset).not.toHaveBeenCalled();

    rerender({ resetKey: "cartulary.view.identities.v1:base" });
    expect(reset).toHaveBeenCalledTimes(1);

    rerender({ resetKey: "" });
    expect(reset).toHaveBeenCalledTimes(1);
  });

  it("remains single-shot for a changed lifecycle under Strict Mode", () => {
    const reset = vi.fn();
    const wrapper = ({ children }: { children: ReactNode }) => (
      <StrictMode>{children}</StrictMode>
    );
    const { rerender } = renderHook(
      ({ resetKey }) => {
        useInspectorLifecycleReset(resetKey, reset);
      },
      {
        initialProps: {
          resetKey: "cartulary.view.assessments.v1:base",
        },
        wrapper,
      },
    );

    rerender({ resetKey: "cartulary.view.assessments.v1:saved-view" });
    expect(reset).toHaveBeenCalledTimes(1);
  });
});
