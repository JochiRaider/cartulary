import { randomUUID } from "node:crypto";
import { dataTestIdSelector } from "@cartulary/ui-contracts";
import type { Locator, Page } from "@playwright/test";

export type ContrastTarget =
  | string
  | {
      readonly id: string;
      readonly selector: string;
      readonly required?: true;
      readonly text?: string;
      readonly kind?: "text" | "nontext" | "focus";
      readonly threshold?: number;
    };
type PaintTarget = Exclude<ContrastTarget, string>;

// Serializable browser observation. Required targets are never filtered out.
export function observeContrast(targets: readonly PaintTarget[]) {
  type Rgba = { r: number; g: number; b: number; a: number };
  function parse(value: string): Rgba {
    if (value === "transparent") return { r: 0, g: 0, b: 0, a: 0 };
    const srgb =
      /^color\(srgb ([\d.]+) ([\d.]+) ([\d.]+)(?: \/ ([\d.]+))?\)$/.exec(
        value.trim(),
      );
    if (srgb)
      return {
        r: Number(srgb[1]) * 255,
        g: Number(srgb[2]) * 255,
        b: Number(srgb[3]) * 255,
        a: Number(srgb[4] ?? 1),
      };
    const match =
      /^rgba?\((\d*\.?\d+),\s*(\d*\.?\d+),\s*(\d*\.?\d+)(?:,\s*(\d*\.?\d+))?\)$/i.exec(
        value.trim(),
      );
    if (!match) throw new Error(`Unsupported color ${value}`);
    return {
      r: Number(match[1]),
      g: Number(match[2]),
      b: Number(match[3]),
      a: Number(match[4] ?? 1),
    };
  }
  function over(front: Rgba, back: Rgba): Rgba {
    const a = front.a + back.a * (1 - front.a);
    if (a === 0) return { r: 0, g: 0, b: 0, a: 0 };
    const channel = (key: "r" | "g" | "b") =>
      (front[key] * front.a + back[key] * back.a * (1 - front.a)) / a;
    return { r: channel("r"), g: channel("g"), b: channel("b"), a };
  }
  function luminance(color: Rgba) {
    const channel = (value: number) => {
      const normalized = value / 255;
      return normalized <= 0.04045
        ? normalized / 12.92
        : ((normalized + 0.055) / 1.055) ** 2.4;
    };
    return (
      0.2126 * channel(color.r) +
      0.7152 * channel(color.g) +
      0.0722 * channel(color.b)
    );
  }
  function ratio(front: Rgba, back: Rgba) {
    const f = luminance(over(front, back)),
      b = luminance(back);
    return (Math.max(f, b) + 0.05) / (Math.min(f, b) + 0.05);
  }
  function unique(colors: Rgba[]) {
    return [
      ...new Map(
        colors.map((color) => [JSON.stringify(color), color]),
      ).values(),
    ];
  }
  function composite(front: Rgba[], back: Rgba[]) {
    if (front.length * back.length > 64)
      throw new Error("Unsupported paint-layer complexity");
    return unique(front.flatMap((f) => back.map((b) => over(f, b))));
  }
  function splitLayers(value: string) {
    const parts: string[] = [];
    let depth = 0,
      start = 0;
    for (let i = 0; i < value.length; i++) {
      if (value[i] === "(") depth++;
      else if (value[i] === ")") depth--;
      else if (value[i] === "," && depth === 0) {
        parts.push(value.slice(start, i).trim());
        start = i + 1;
      }
    }
    parts.push(value.slice(start).trim());
    return parts;
  }
  function imagePalette(value: string): Rgba[] {
    if (value === "none") return [{ r: 0, g: 0, b: 0, a: 0 }];
    // Grid state cues use two solid bands, not interpolated gradients. Inspect
    // every possible band color and layer combination; never sample away a cue.
    const gradient = /^repeating-linear-gradient\((.*)\)$/.exec(value);
    if (!gradient) throw new Error("Unsupported background image");
    const parts = splitLayers(gradient[1] ?? "");
    if (!/^-?[\d.]+deg$/.test(parts.shift() ?? "") || parts.length !== 4)
      throw new Error("Unsupported background gradient geometry");
    const stops = parts.map((part, index) => {
      const stop =
        /^(rgba?\([^)]+\)|color\(srgb [^)]+\)|transparent)(?: ([\d.]+)(px|%))?$/.exec(
          part,
        );
      if (!stop || (index !== 0 && stop[2] === undefined))
        throw new Error("Unsupported gradient stop");
      return {
        color: parse(stop[1] ?? ""),
        at: Number(stop[2] ?? 0),
        unit: stop[3],
      };
    });
    const [a, b, c, d] = stops;
    if (
      !a ||
      !b ||
      !c ||
      !d ||
      a.at !== 0 ||
      b.at !== c.at ||
      b.at <= 0 ||
      d.at <= c.at ||
      (a.unit !== undefined && a.unit !== b.unit) ||
      b.unit !== c.unit ||
      c.unit !== d.unit ||
      JSON.stringify(a.color) !== JSON.stringify(b.color) ||
      JSON.stringify(c.color) !== JSON.stringify(d.color)
    )
      throw new Error("Unsupported interpolated gradient");
    return unique([a.color, c.color]);
  }
  function background(element: Element | null): Rgba[] {
    const layers: Rgba[][] = [];
    let opaque = false;
    for (
      let candidate = element;
      candidate;
      candidate = candidate.parentElement
    ) {
      const style = getComputedStyle(candidate);
      if (
        style.display === "none" ||
        style.visibility === "hidden" ||
        style.visibility === "collapse" ||
        Number(style.opacity || 1) !== 1
      )
        throw new Error("Hidden target or unsupported ancestor opacity");
      if (
        (style.filter && style.filter !== "none") ||
        (style.getPropertyValue("backdrop-filter") &&
          style.getPropertyValue("backdrop-filter") !== "none") ||
        (style.mixBlendMode && style.mixBlendMode !== "normal")
      )
        throw new Error("Unsupported ancestor filter or blending");
      if (opaque) continue;
      let colors = [parse(style.backgroundColor)];
      for (const layer of splitLayers(
        style.backgroundImage || "none",
      ).reverse())
        colors = composite(imagePalette(layer), colors);
      layers.push(colors);
      opaque = colors.every((color) => color.a === 1);
    }
    if (!opaque) throw new Error("No opaque painted background found");
    return layers
      .reverse()
      .reduce(
        (back, front) => composite(front, back),
        [{ r: 0, g: 0, b: 0, a: 0 }],
      );
  }
  return targets.map((target) => {
    const kind = target.kind ?? "text";
    const threshold = target.threshold ?? (kind === "text" ? 4.5 : 3);
    const base = { target: target.id, threshold };
    try {
      if (
        !Number.isFinite(threshold) ||
        threshold < (kind === "text" ? 4.5 : 3)
      )
        throw new Error("Invalid declared contrast threshold");
      const matches = document.querySelectorAll(target.selector);
      if (matches.length !== 1)
        throw new Error(
          `Expected exactly one target ${target.selector}; found ${matches.length}`,
        );
      const element = matches[0];
      if (
        !(element instanceof HTMLElement) ||
        element.getClientRects().length === 0
      )
        throw new Error("Target is not rendered");
      if (target.text !== undefined && element.textContent !== target.text)
        throw new Error(
          "Target does not contain the expected source text or marker",
        );
      const style = getComputedStyle(element);
      const inside = background(element);
      let foreground: Rgba;
      let ratios: number[];
      if (kind === "focus") {
        if (element !== document.activeElement)
          throw new Error("Focus target is not focused");
        foreground = parse(style.outlineColor);
        const outside = background(element.parentElement);
        let offset = Number.parseFloat(style.outlineOffset || "0");
        let width = Number.parseFloat(style.outlineWidth);
        if (
          style.outlineStyle === "none" ||
          !(width > 0) ||
          foreground.a === 0
        ) {
          // A single unblurred, unshifted outer shadow is a rectangular ring.
          // Selected mentions use this ring with a transparent forced-color outline.
          const shadow = /^(rgba?\([^)]+\)) 0px 0px 0px (\d*\.?\d+)px$/.exec(
            style.boxShadow,
          );
          if (!shadow || !(Number(shadow[2]) > 0))
            throw new Error(
              "Unsupported focus indicator: require a visible outline or solid outer shadow ring",
            );
          foreground = parse(shadow[1] ?? "");
          if (foreground.a === 0)
            throw new Error("Transparent focus indicator");
          width = Number(shadow[2]);
          offset = 0;
        }
        if (!Number.isFinite(offset))
          throw new Error("Unsupported focus offset");
        // Depth increases into the border box. An outline occupies the band
        // [-offset-width, -offset], including when inset into a grid header.
        // Composite each edge over its actual paint before comparing it with
        // the paint immediately beside that edge.
        const innerDepth = -offset;
        const outerDepth = innerDepth - width;
        const epsilon = Math.min(width / 4, 0.001);
        const observed: number[] = [];
        for (const side of ["top", "right", "bottom", "left"]) {
          const borderWidth = Number.parseFloat(
            style.getPropertyValue(`border-${side}-width`) || "0",
          );
          const borderStyle = style.getPropertyValue(`border-${side}-style`);
          const paintAt = (depth: number) => {
            if (depth < 0) return outside;
            if (
              depth < borderWidth &&
              borderStyle !== "none" &&
              borderStyle !== "hidden"
            ) {
              if (borderStyle !== "solid")
                throw new Error("Unsupported patterned focus-adjacent border");
              return composite(
                [parse(style.getPropertyValue(`border-${side}-color`))],
                inside,
              );
            }
            return inside;
          };
          for (const [under, adjacent] of [
            [paintAt(innerDepth - epsilon), paintAt(innerDepth + epsilon)],
            [paintAt(outerDepth + epsilon), paintAt(outerDepth - epsilon)],
          ] as const) {
            for (const paint of under)
              for (const neighbor of adjacent)
                observed.push(ratio(over(foreground, paint), neighbor));
          }
        }
        ratios = [...new Set(observed)];
      } else {
        if (kind === "text") {
          const fill = style.getPropertyValue("-webkit-text-fill-color");
          if (
            (fill && fill !== style.color) ||
            (style.textShadow && style.textShadow !== "none")
          )
            throw new Error("Unsupported text paint override or shadow");
        }
        foreground = parse(style.color);
        ratios = inside.map((color) => ratio(foreground, color));
      }
      const minimum = Math.min(...ratios);
      return {
        ...base,
        result: minimum >= threshold ? "pass" : "fail",
        ratio: minimum,
        adjacent_ratios: ratios,
        foreground,
        background: inside[0],
        backgrounds: inside,
      };
    } catch (error) {
      return {
        ...base,
        result: "unsupported",
        ratio: null,
        error: String(error),
      };
    }
  });
}

export async function collectContrastChecks(
  page: Page,
  targets: readonly ContrastTarget[],
) {
  return page.evaluate(
    observeContrast,
    [...new Set(targets)].map((target) =>
      typeof target === "string"
        ? { id: target, selector: dataTestIdSelector(target) }
        : target,
    ),
  );
}

// Resolve rich Playwright locators once, then give the browser observer an
// unambiguous temporary selector. Restore caller-owned attributes on every exit.
export async function withContrastTarget<T>(
  locator: Locator,
  observe: (selector: string) => Promise<T>,
): Promise<T> {
  const count = await locator.count();
  if (count !== 1)
    throw new Error(`Expected exactly one contrast target; found ${count}`);
  const element = await locator.elementHandle();
  if (!element)
    throw new Error("Contrast target disappeared before observation");
  const attribute = "data-a11y-contrast-observation";
  const marker = randomUUID();
  const errors: unknown[] = [];
  let value: T | undefined;
  let before: string | null = null;
  let captured = false;
  try {
    before = await element.getAttribute(attribute);
    captured = true;
    await element.evaluate(
      (element, { attribute, marker }) =>
        element.setAttribute(attribute, marker),
      { attribute, marker },
    );
    value = await observe(`[${attribute}="${marker}"]`);
  } catch (error) {
    errors.push(error);
  } finally {
    // Restore the exact observed node, even if a render detached it. Never
    // re-resolve the locator and mutate a replacement control during cleanup.
    if (captured && !locator.page().isClosed()) {
      try {
        await element.evaluate(
          (element, { attribute, before }) => {
            if (before === null) element.removeAttribute(attribute);
            else element.setAttribute(attribute, before);
          },
          { attribute, before },
        );
      } catch (cleanup) {
        errors.push(cleanup);
      }
    }
    try {
      await element.dispose();
    } catch (cleanup) {
      errors.push(cleanup);
    }
  }
  if (errors.length === 1) throw errors[0];
  if (errors.length > 1)
    throw new AggregateError(
      errors,
      "Contrast observation and marker cleanup failed",
    );
  return value as T;
}
