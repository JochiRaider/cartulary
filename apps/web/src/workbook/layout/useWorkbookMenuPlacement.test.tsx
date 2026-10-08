import { cleanup, render, screen } from "@testing-library/react";
import { useRef } from "react";
import { afterEach, expect, it } from "vitest";
import { useWorkbookMenuPlacement } from "./useWorkbookMenuPlacement";

afterEach(cleanup);

it("placement retirement preserves the owner's current visibility across reopening", () => {
  function Menu({ open }: { readonly open: boolean }) {
    const anchor = useRef<HTMLButtonElement>(null);
    const panel = useRef<HTMLDivElement>(null);
    useWorkbookMenuPlacement(open, panel, anchor);
    return (
      <>
        <button ref={anchor} type="button">
          Open menu
        </button>
        <div
          ref={panel}
          role="menu"
          aria-label="Retained menu"
          style={{ display: open ? "grid" : "none", position: "absolute" }}
        >
          Menu content
        </div>
      </>
    );
  }
  const { rerender } = render(<Menu open />);
  const panel = screen.getByRole("menu", { name: "Retained menu" });
  rerender(<Menu open={false} />);
  expect(screen.queryByRole("menu", { name: "Retained menu" })).toBeNull();
  expect(panel.style.display).toBe("none");
  expect(panel.style.position).toBe("absolute");
  rerender(<Menu open />);
  expect(screen.getByRole("menu", { name: "Retained menu" })).toBe(panel);
  rerender(<Menu open={false} />);
  expect(screen.queryByRole("menu", { name: "Retained menu" })).toBeNull();
});
