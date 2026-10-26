// @vitest-environment jsdom
import { afterEach, describe, expect, it, vi } from "vitest";
import { observeContrast } from "./contrast";

function required<T extends keyof HTMLElementTagNameMap>(tag: T) {
  const element = document.querySelector(tag);
  if (!element) throw new Error(`Missing contrast fixture ${tag}`);
  return element;
}
function fixture(style = "") {
  document.documentElement.style.backgroundColor = "rgb(255, 255, 255)";
  document.body.innerHTML = `<section style="background: rgba(0, 0, 0, .5)"><button id="target" style="background: transparent; border: 0; color: black; ${style}">Action</button></section>`;
  vi.spyOn(HTMLElement.prototype, "getClientRects").mockReturnValue([
    {},
  ] as unknown as DOMRectList);
}
afterEach(() => {
  vi.restoreAllMocks();
  document.body.replaceChildren();
  document.documentElement.removeAttribute("style");
});
describe("required contrast observations", () => {
  it("rejects missing ambiguous and hidden targets instead of dropping them", () => {
    fixture();
    expect(
      observeContrast([{ id: "missing", selector: ".missing" }])[0],
    ).toMatchObject({
      result: "unsupported",
      error: expect.stringContaining("found 0"),
    });
    document.body.append(required("button").cloneNode(true));
    expect(
      observeContrast([{ id: "duplicate", selector: "button" }])[0],
    ).toMatchObject({
      result: "unsupported",
      error: expect.stringContaining("found 2"),
    });
    fixture("display: none");
    expect(
      observeContrast([{ id: "hidden", selector: "button" }])[0]?.result,
    ).toBe("unsupported");
  });
  it("composites translucent foreground and every background layer before applying unrounded declared thresholds", () => {
    fixture("color: white");
    const text = observeContrast([{ id: "text", selector: "button" }])[0];
    expect(text).toMatchObject({
      result: "fail",
      threshold: 4.5,
      background: { r: 127.5, g: 127.5, b: 127.5, a: 1 },
    });
    expect(text?.ratio).toBeCloseTo(3.976653, 5);
    expect(
      observeContrast([
        { id: "state", selector: "button", kind: "nontext" },
      ])[0],
    ).toMatchObject({ result: "pass", threshold: 3 });
    expect(
      observeContrast([{ id: "strict", selector: "button", threshold: 7 }])[0]
        ?.result,
    ).toBe("fail");
    required("section").style.backgroundColor = "white";
    required("button").style.color = "rgba(0, 0, 0, .5)";
    expect(
      observeContrast([{ id: "alpha-text", selector: "button" }])[0]?.ratio,
    ).toBeCloseTo(3.976653, 5);
  });
  it("checks every composited hard-stripe color while rejecting interpolated gradients", () => {
    fixture(
      "background-color: white; color: black; background-image: repeating-linear-gradient(135deg, rgba(0, 0, 0, 0), rgba(0, 0, 0, 0) 7px, rgba(0, 0, 0, .5) 7px, rgba(0, 0, 0, .5) 8px), none",
    );
    const striped = observeContrast([{ id: "striped", selector: "button" }])[0];
    expect(striped, JSON.stringify(striped)).toMatchObject({
      result: "pass",
      backgrounds: [
        { r: 255, g: 255, b: 255, a: 1 },
        { r: 127.5, g: 127.5, b: 127.5, a: 1 },
      ],
    });
    expect(striped?.ratio).toBeCloseTo(5.280823, 5);
    required("button").style.backgroundImage =
      "repeating-linear-gradient(135deg, rgb(255, 255, 255), rgb(255, 255, 255) 6px, rgb(0, 0, 0) 7px, rgb(0, 0, 0) 8px)";
    expect(
      observeContrast([{ id: "interpolated", selector: "button" }])[0]?.result,
    ).toBe("unsupported");
  });
  it("requires focus contrast against both adjacent paints and rejects unsupported effects", () => {
    fixture(
      "background: white; outline-style: solid; outline-width: 2px; outline-color: white",
    );
    required("section").style.backgroundColor = "black";
    required("button").focus();
    expect(
      observeContrast([{ id: "focus", selector: "button", kind: "focus" }])[0],
    ).toMatchObject({ result: "fail", threshold: 3, adjacent_ratios: [1, 21] });
    const button = required("button");
    button.style.outlineOffset = "-4px";
    expect(
      observeContrast([{ id: "inset", selector: "button", kind: "focus" }])[0],
    ).toMatchObject({ result: "fail", adjacent_ratios: [1] });
    button.style.outlineColor = "rgba(0, 0, 0, .5)";
    expect(
      observeContrast([
        { id: "inset-alpha", selector: "button", kind: "focus" },
      ])[0]?.ratio,
    ).toBeCloseTo(3.976653, 5);
    button.style.outlineOffset = "-2px";
    button.style.outlineColor = "white";
    expect(
      observeContrast([{ id: "edge", selector: "button", kind: "focus" }])[0],
    ).toMatchObject({ result: "fail", adjacent_ratios: [1, 21] });
    button.style.borderTop = "3px solid black";
    expect(
      observeContrast([{ id: "border", selector: "button", kind: "focus" }])[0],
    ).toMatchObject({ result: "fail", adjacent_ratios: [21, 1] });
    button.style.borderTop = "0px";
    button.style.outlineColor = "transparent";
    button.style.boxShadow = "rgb(255, 255, 255) 0px 0px 0px 2px";
    expect(
      observeContrast([
        { id: "shadow-ring", selector: "button", kind: "focus" },
      ])[0],
    ).toMatchObject({ result: "fail", adjacent_ratios: [1, 21] });
    button.style.boxShadow = "rgb(255, 255, 255) 0px 0px 3px 2px";
    expect(
      observeContrast([
        { id: "blurred-ring", selector: "button", kind: "focus" },
      ])[0]?.result,
    ).toBe("unsupported");
    for (const style of [
      "background-image: linear-gradient(black,white)",
      "filter: blur(1px)",
      "mix-blend-mode: multiply",
      "opacity: .5",
    ]) {
      fixture(style);
      expect(
        observeContrast([{ id: style, selector: "button" }])[0]?.result,
      ).toBe("unsupported");
    }
  });
});
