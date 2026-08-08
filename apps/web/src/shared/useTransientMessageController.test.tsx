import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
} from "@testing-library/react";
import { useState } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { useTransientMessageController } from "./useTransientMessageController";

function Harness({ actionAvailable = false }: { actionAvailable?: boolean }) {
  const [visible, setVisible] = useState(true);
  const controller = useTransientMessageController({
    actionAvailable,
    enabled: visible,
    messageKey: "confirmation-1",
    onDismiss: () => setVisible(false),
  });
  return visible ? (
    <div role="status" tabIndex={-1} {...controller}>
      Saved.
      <button type="button">Target</button>
    </div>
  ) : null;
}

afterEach(() => {
  cleanup();
  vi.useRealTimers();
});

describe("useTransientMessageController", () => {
  it("dismisses after 5,000ms of visible unpaused time", () => {
    vi.useFakeTimers();
    render(<Harness />);
    act(() => vi.advanceTimersByTime(4999));
    expect(screen.queryByRole("status")).not.toBeNull();
    act(() => vi.advanceTimersByTime(1));
    expect(screen.queryByRole("status")).toBeNull();
  });

  it("pauses for pointer hover and keyboard focus and resumes the remainder", () => {
    vi.useFakeTimers();
    render(<Harness />);
    const message = screen.getByRole("status");
    act(() => vi.advanceTimersByTime(2000));
    fireEvent.mouseEnter(message);
    act(() => vi.advanceTimersByTime(5000));
    expect(screen.queryByRole("status")).not.toBeNull();
    fireEvent.mouseLeave(message);
    act(() => vi.advanceTimersByTime(1000));
    fireEvent.focus(screen.getByRole("button"));
    act(() => vi.advanceTimersByTime(5000));
    expect(screen.queryByRole("status")).not.toBeNull();
    fireEvent.blur(screen.getByRole("button"));
    act(() => vi.advanceTimersByTime(1999));
    expect(screen.queryByRole("status")).not.toBeNull();
    act(() => vi.advanceTimersByTime(1));
    expect(screen.queryByRole("status")).toBeNull();
  });

  it("does not dismiss a message with a still-valid action", () => {
    vi.useFakeTimers();
    render(<Harness actionAvailable />);
    act(() => vi.advanceTimersByTime(10000));
    expect(screen.queryByRole("status")).not.toBeNull();
  });

  it("pauses while the document is hidden and resumes the exact remainder", () => {
    vi.useFakeTimers();
    const originalHidden = document.hidden;
    Object.defineProperty(document, "hidden", {
      configurable: true,
      value: false,
    });
    try {
      render(<Harness />);
      act(() => vi.advanceTimersByTime(2000));
      Object.defineProperty(document, "hidden", {
        configurable: true,
        value: true,
      });
      fireEvent(document, new Event("visibilitychange"));
      act(() => vi.advanceTimersByTime(5000));
      expect(screen.queryByRole("status")).not.toBeNull();
      Object.defineProperty(document, "hidden", {
        configurable: true,
        value: false,
      });
      fireEvent(document, new Event("visibilitychange"));
      act(() => vi.advanceTimersByTime(2999));
      expect(screen.queryByRole("status")).not.toBeNull();
      act(() => vi.advanceTimersByTime(1));
      expect(screen.queryByRole("status")).toBeNull();
    } finally {
      Object.defineProperty(document, "hidden", {
        configurable: true,
        value: originalHidden,
      });
    }
  });
});
