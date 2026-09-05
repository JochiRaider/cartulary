import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { type RefObject, useRef, useState } from "react";
import { afterEach, describe, expect, it } from "vitest";
import { useRegisteredOverlayNavigation } from "./useRegisteredOverlayNavigation";

const itemKeys = ["first", "disabled", "last"] as const;

afterEach(cleanup);

function OverlayHarness({
  hideTrigger = false,
  keys = itemKeys,
  restoreFocusOnSubjectChange = true,
  subjectKey,
}: {
  readonly hideTrigger?: boolean;
  readonly keys?: readonly (typeof itemKeys)[number][];
  readonly restoreFocusOnSubjectChange?: boolean;
  readonly subjectKey: string;
}) {
  const [isOpen, setIsOpen] = useState(false);
  const triggerRef = useRef<HTMLButtonElement | null>(null);
  const fallbackRef = useRef<HTMLButtonElement | null>(null);
  const navigation = useRegisteredOverlayNavigation({
    reconcileItems: true,
    fallbackFocusRef: fallbackRef,
    initialItemKey: "first",
    isOpen,
    itemKeys: keys,
    onRequestClose: () => setIsOpen(false),
    subjectKey,
    restoreFocusOnSubjectChange,
    triggerRef,
  });

  return (
    <div>
      {hideTrigger ? null : (
        <button
          ref={triggerRef}
          type="button"
          onClick={() => {
            navigation.prepareOpen("first");
            setIsOpen(true);
          }}
        >
          Open
        </button>
      )}
      <button ref={fallbackRef} type="button">
        Fallback
      </button>
      {isOpen ? (
        <div
          role="menu"
          tabIndex={-1}
          onBlur={navigation.onOverlayBlur}
          onKeyDown={(event) => {
            if (navigation.activeKey !== null) {
              navigation.onItemKeyDown(event, navigation.activeKey);
            }
          }}
        >
          {keys.map((key) => (
            <button
              key={key}
              ref={navigation.registerItem(key)}
              disabled={key === "disabled"}
              role="menuitem"
              tabIndex={navigation.tabIndexFor(key)}
              type="button"
              onFocus={() => navigation.onItemFocus(key)}
              onKeyDown={(event) => navigation.onItemKeyDown(event, key)}
            >
              {key}
            </button>
          ))}
        </div>
      ) : null}
    </div>
  );
}

function UnmountingOverlayHarness({
  restoreFocusOnUnmount = true,
}: {
  readonly restoreFocusOnUnmount?: boolean;
}) {
  const [isOpen, setIsOpen] = useState(false);
  const returnRef = useRef<HTMLSpanElement | null>(null);
  const fallbackRef = useRef<HTMLButtonElement | null>(null);
  return (
    <div>
      <span ref={returnRef}>Pointer target</span>
      <button ref={fallbackRef} type="button" onClick={() => setIsOpen(true)}>
        Open unmounting overlay
      </button>
      {isOpen ? (
        <UnmountingOverlay
          fallbackRef={fallbackRef}
          onClose={() => setIsOpen(false)}
          returnRef={returnRef}
          restoreFocusOnUnmount={restoreFocusOnUnmount}
        />
      ) : null}
    </div>
  );
}

function UnmountingOverlay({
  fallbackRef,
  onClose,
  returnRef,
  restoreFocusOnUnmount,
}: {
  readonly fallbackRef: RefObject<HTMLElement | null>;
  readonly onClose: () => void;
  readonly returnRef: RefObject<HTMLElement | null>;
  readonly restoreFocusOnUnmount: boolean;
}) {
  const navigation = useRegisteredOverlayNavigation({
    reconcileItems: true,
    fallbackFocusRef: fallbackRef,
    initialItemKey: "only",
    isOpen: true,
    itemKeys: ["only"],
    onRequestClose: onClose,
    subjectKey: "unmounting",
    restoreFocusOnUnmount,
    triggerRef: returnRef,
  });
  return (
    <button
      ref={navigation.registerItem("only")}
      type="button"
      onKeyDown={(event) => navigation.onItemKeyDown(event, "only")}
    >
      Only action
    </button>
  );
}

describe("registered overlay navigation", () => {
  it("cancels requested restoration on unmount when the consumer owns that policy", () => {
    render(<UnmountingOverlayHarness restoreFocusOnUnmount={false} />);
    fireEvent.click(
      screen.getByRole("button", { name: "Open unmounting overlay" }),
    );
    fireEvent.keyDown(screen.getByRole("button", { name: "Only action" }), {
      key: "Escape",
    });
    expect(screen.queryByRole("button", { name: "Only action" })).toBeNull();
    expect(document.activeElement).not.toBe(
      screen.getByRole("button", { name: "Open unmounting overlay" }),
    );
  });

  it("reconciles a removed focused item without using its detached element", () => {
    const { rerender } = render(<OverlayHarness subjectKey="surface-a" />);
    fireEvent.click(screen.getByRole("button", { name: "Open" }));
    rerender(
      <OverlayHarness keys={["disabled", "last"]} subjectKey="surface-a" />,
    );
    expect(document.activeElement).toBe(
      screen.getByRole("menuitem", { name: "last" }),
    );
  });

  it("allows a destination owner to suppress subject-change restoration", () => {
    const { rerender } = render(
      <OverlayHarness
        restoreFocusOnSubjectChange={false}
        subjectKey="surface-a"
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Open" }));
    rerender(
      <OverlayHarness
        restoreFocusOnSubjectChange={false}
        subjectKey="surface-b"
      />,
    );
    expect(screen.queryByRole("menu")).toBeNull();
    expect(document.activeElement).not.toBe(
      screen.getByRole("button", { name: "Open" }),
    );
  });

  it("focuses on open, wraps, skips disabled items, and restores on Escape", () => {
    render(<OverlayHarness subjectKey="surface-a" />);
    const trigger = screen.getByRole("button", { name: "Open" });
    fireEvent.click(trigger);
    const first = screen.getByRole("menuitem", { name: "first" });
    const last = screen.getByRole("menuitem", { name: "last" });
    expect(document.activeElement).toBe(first);

    fireEvent.keyDown(first, { key: "ArrowDown" });
    expect(document.activeElement).toBe(last);
    fireEvent.keyDown(last, { key: "ArrowDown" });
    expect(document.activeElement).toBe(first);
    fireEvent.keyDown(first, { key: "End" });
    expect(document.activeElement).toBe(last);
    fireEvent.keyDown(last, { key: "Home" });
    expect(document.activeElement).toBe(first);

    fireEvent.keyDown(first, { key: "Escape" });
    expect(screen.queryByRole("menu")).toBeNull();
    expect(document.activeElement).toBe(trigger);
  });

  it("closes without restoring when focus leaves the registered surface", () => {
    render(<OverlayHarness subjectKey="surface-a" />);
    fireEvent.click(screen.getByRole("button", { name: "Open" }));
    fireEvent.blur(screen.getByRole("menuitem", { name: "first" }), {
      relatedTarget: document.body,
    });
    expect(screen.queryByRole("menu")).toBeNull();
  });

  it("keeps a programmatically focused overlay root inside the focus boundary", () => {
    render(<OverlayHarness subjectKey="surface-a" />);
    const trigger = screen.getByRole("button", { name: "Open" });
    fireEvent.click(trigger);
    const menu = screen.getByRole("menu");
    fireEvent.focus(menu);
    expect(screen.queryByRole("menu")).toBe(menu);
    fireEvent.keyDown(menu, { key: "Escape" });
    expect(screen.queryByRole("menu")).toBeNull();
    expect(document.activeElement).toBe(trigger);
  });

  it("closes on subject change and restores to a registered fallback", () => {
    const { rerender } = render(<OverlayHarness subjectKey="surface-a" />);
    fireEvent.click(screen.getByRole("button", { name: "Open" }));
    rerender(<OverlayHarness hideTrigger subjectKey="surface-b" />);
    expect(screen.queryByRole("menu")).toBeNull();
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: "Fallback" }),
    );
  });

  it("restores through the fallback when closing unmounts the navigation owner", () => {
    render(<UnmountingOverlayHarness />);
    const fallback = screen.getByRole("button", {
      name: "Open unmounting overlay",
    });
    fireEvent.click(fallback);
    const action = screen.getByRole("button", { name: "Only action" });
    expect(document.activeElement).toBe(action);
    fireEvent.keyDown(action, { key: "Escape" });
    expect(screen.queryByRole("button", { name: "Only action" })).toBeNull();
    expect(document.activeElement).toBe(fallback);
  });
});
