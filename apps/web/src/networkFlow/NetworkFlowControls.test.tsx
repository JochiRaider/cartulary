import { fireEvent, render, screen } from "@testing-library/react";
import { createRef } from "react";
import { describe, expect, it, vi } from "vitest";

import {
  NetworkFlowButton,
  NetworkFlowChromeStyles,
  NetworkFlowField,
  NetworkFlowIconButton,
  NetworkFlowNumberInput,
  NetworkFlowSelect,
  NetworkFlowTextInput,
  networkFlowChromeCssText,
  networkFlowChromeRootClassName,
} from "./NetworkFlowControls";

describe("Network Flow control presentation seam", () => {
  it("preserves native button behavior and exposes deterministic variants", () => {
    const onClick = vi.fn();
    render(
      <div className={networkFlowChromeRootClassName}>
        <NetworkFlowButton onClick={onClick} variant="primary">
          Apply query
        </NetworkFlowButton>
        <NetworkFlowButton pending variant="danger">
          Retiring graph…
        </NetworkFlowButton>
      </div>,
    );

    const apply = screen.getByRole("button", { name: "Apply query" });
    fireEvent.click(apply);
    expect(onClick).toHaveBeenCalledOnce();
    expect(apply.getAttribute("data-network-flow-variant")).toBe("primary");

    const pending = screen.getByRole("button", { name: "Retiring graph…" });
    expect(pending.getAttribute("aria-busy")).toBe("true");
    expect((pending as HTMLButtonElement).disabled).toBe(true);
  });

  it("retains selected semantics and requires an accessible icon name", () => {
    render(
      <div className={networkFlowChromeRootClassName}>
        <NetworkFlowButton selected variant="mode">
          Graph
        </NetworkFlowButton>
        <NetworkFlowIconButton aria-label="Refresh graph">
          ↻
        </NetworkFlowIconButton>
      </div>,
    );

    expect(
      screen
        .getByRole("button", { name: "Graph" })
        .getAttribute("aria-pressed"),
    ).toBe("true");
    expect(screen.getByRole("button", { name: "Refresh graph" })).toBeTruthy();
  });

  it("forwards refs and preserves field, invalid, and numeric semantics", () => {
    const inputRef = createRef<HTMLInputElement>();
    render(
      <div className={networkFlowChromeRootClassName}>
        <NetworkFlowField
          error="A display name is required."
          errorId="display-name-error"
          htmlFor="display-name"
          label="Display name"
        >
          <NetworkFlowTextInput
            aria-describedby="display-name-error"
            aria-invalid="true"
            id="display-name"
            ref={inputRef}
          />
        </NetworkFlowField>
        <NetworkFlowNumberInput aria-label="Page size" />
        <NetworkFlowSelect aria-label="Profile">
          <option>Default</option>
        </NetworkFlowSelect>
      </div>,
    );

    expect(inputRef.current).toBe(screen.getByLabelText("Display name"));
    expect(
      screen.getByLabelText("Display name").getAttribute("aria-describedby"),
    ).toBe("display-name-error");
    expect((screen.getByLabelText("Page size") as HTMLInputElement).type).toBe(
      "number",
    );
    expect(screen.getByLabelText("Page size").getAttribute("inputmode")).toBe(
      "numeric",
    );
    expect(screen.getByRole("combobox", { name: "Profile" })).toBeTruthy();
  });

  it("ships scoped dark, focus, autofill, dialog, and responsive rules", () => {
    render(<NetworkFlowChromeStyles />);
    expect(document.querySelector("style")?.textContent).toBe(
      networkFlowChromeCssText,
    );
    expect(networkFlowChromeCssText).toContain("color-scheme: dark");
    expect(networkFlowChromeCssText).toContain(":focus-visible");
    expect(networkFlowChromeCssText).toContain(":-webkit-autofill");
    expect(networkFlowChromeCssText).toContain(".network-flow-dialog");
    expect(networkFlowChromeCssText).toContain("@media (max-width: 768px)");
  });
});
