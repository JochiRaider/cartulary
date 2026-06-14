import { describe, expect, it } from "vitest";
import {
  captureViewportAnchor,
  computeRestoredViewportScroll,
  isRectFullyVisibleWithinContainer,
  type RectLike,
} from "./workbookContinuity";

const containerRect: RectLike = {
  top: 100,
  left: 40,
  right: 440,
  bottom: 400,
  width: 400,
  height: 300,
};

describe("workbookContinuity", () => {
  it("captures and restores an unchanged anchor without moving the grid", () => {
    const elementRect: RectLike = {
      top: 170,
      left: 85,
      right: 225,
      bottom: 210,
      width: 140,
      height: 40,
    };

    const anchor = captureViewportAnchor(containerRect, elementRect);

    expect(anchor).toEqual({
      top: 70,
      left: 45,
      width: 140,
      height: 40,
    });
    expect(
      computeRestoredViewportScroll({
        preservedScroll: { top: 240, left: 18 },
        currentScroll: { top: 240, left: 18 },
        preservedAnchor: anchor,
        containerRect,
        elementRect,
      }),
    ).toEqual({ top: 240, left: 18 });
  });

  it("applies only the vertical delta needed to restore the preserved anchor", () => {
    const preservedAnchor = {
      top: 70,
      left: 45,
      width: 140,
      height: 40,
    };

    expect(
      computeRestoredViewportScroll({
        preservedScroll: { top: 240, left: 18 },
        currentScroll: { top: 240, left: 18 },
        preservedAnchor,
        containerRect,
        elementRect: {
          top: 230,
          left: 85,
          right: 225,
          bottom: 270,
          width: 140,
          height: 40,
        },
      }),
    ).toEqual({ top: 300, left: 18 });
  });

  it("adds the minimum extra scroll needed when anchor restoration would clip the row", () => {
    const preservedAnchor = {
      top: 240,
      left: 45,
      width: 140,
      height: 40,
    };

    expect(
      computeRestoredViewportScroll({
        preservedScroll: { top: 240, left: 18 },
        currentScroll: { top: 240, left: 18 },
        preservedAnchor,
        containerRect,
        elementRect: {
          top: 340,
          left: 85,
          right: 225,
          bottom: 440,
          width: 140,
          height: 100,
        },
      }),
    ).toEqual({ top: 280, left: 18 });
  });

  it("preserves horizontal scroll until the focused element would be clipped", () => {
    const elementRect: RectLike = {
      top: 170,
      left: 390,
      right: 470,
      bottom: 210,
      width: 80,
      height: 40,
    };

    expect(
      computeRestoredViewportScroll({
        preservedScroll: { top: 240, left: 18 },
        currentScroll: { top: 240, left: 18 },
        preservedAnchor: {
          top: 70,
          left: 45,
          width: 140,
          height: 40,
        },
        containerRect,
        elementRect,
      }),
    ).toEqual({ top: 240, left: 48 });
    expect(
      isRectFullyVisibleWithinContainer(containerRect, {
        top: 170,
        left: 85,
        right: 225,
        bottom: 210,
        width: 140,
        height: 40,
      }),
    ).toBe(true);
    expect(isRectFullyVisibleWithinContainer(containerRect, elementRect)).toBe(
      false,
    );
  });

  it("adds the minimum horizontal and vertical deltas when anchor restoration would clip both axes", () => {
    expect(
      computeRestoredViewportScroll({
        preservedScroll: { top: 240, left: 18 },
        currentScroll: { top: 240, left: 18 },
        preservedAnchor: {
          top: 240,
          left: 45,
          width: 140,
          height: 40,
        },
        containerRect,
        elementRect: {
          top: 340,
          left: 390,
          right: 470,
          bottom: 440,
          width: 80,
          height: 100,
        },
      }),
    ).toEqual({ top: 280, left: 48 });
  });

  it("bases visibility repair on the actual clamped scroll after restoring the preserved scroll", () => {
    expect(
      computeRestoredViewportScroll({
        preservedScroll: { top: 800, left: 18 },
        currentScroll: { top: 500, left: 18 },
        preservedAnchor: {
          top: 160,
          left: 45,
          width: 140,
          height: 40,
        },
        containerRect,
        elementRect: {
          top: -20,
          left: 85,
          right: 225,
          bottom: 20,
          width: 140,
          height: 40,
        },
      }),
    ).toEqual({ top: 220, left: 18 });
  });
});
