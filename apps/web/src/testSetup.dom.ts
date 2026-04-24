if (globalThis.ResizeObserver === undefined) {
  Object.defineProperty(globalThis, "ResizeObserver", {
    configurable: true,
    value: class ResizeObserver {
      disconnect() {}
      observe() {}
      unobserve() {}
    },
  });
}

if (typeof window !== "undefined") {
  Object.defineProperty(window, "focus", {
    configurable: true,
    value: () => {},
  });
}

if (
  typeof HTMLElement !== "undefined" &&
  typeof HTMLElement.prototype.scrollIntoView !== "function"
) {
  Object.defineProperty(HTMLElement.prototype, "scrollIntoView", {
    configurable: true,
    value: () => {},
  });
}

type TestLayoutWindow = Window & {
  __cartularyTestLayout?: {
    readonly height?: number;
    readonly width?: number;
  };
};

function parsePixelValue(value: string | null | undefined) {
  if (!value) {
    return null;
  }
  const match = /^\s*(\d+(?:\.\d+)?)px\s*$/u.exec(value);
  return match ? Number(match[1]) : null;
}

function sumGridTemplatePixels(value: string | null | undefined) {
  if (!value) {
    return null;
  }
  const widths = Array.from(value.matchAll(/(\d+(?:\.\d+)?)px/gu)).map(
    (match) => Number(match[1]),
  );
  if (widths.length === 0) {
    return null;
  }
  return widths.reduce((total, width) => total + width, 0);
}

function configuredLayoutSize(element: HTMLElement, axis: "height" | "width") {
  const explicit = Number(
    element.getAttribute(`data-cartulary-test-${axis}`) ??
      document.documentElement.getAttribute(`data-cartulary-test-${axis}`) ??
      "",
  );
  if (Number.isFinite(explicit) && explicit > 0) {
    return explicit;
  }

  const windowValue = (window as TestLayoutWindow).__cartularyTestLayout?.[
    axis
  ];
  if (typeof windowValue === "number" && Number.isFinite(windowValue)) {
    return windowValue;
  }

  const inlineStyle = element.style;
  const direct = parsePixelValue(inlineStyle[axis]);
  if (direct !== null) {
    return direct;
  }

  if (axis === "width") {
    const minWidth = parsePixelValue(inlineStyle.minWidth);
    if (minWidth !== null) {
      return minWidth;
    }
    if (element.getAttribute("role") === "grid") {
      return 1;
    }
    const gridWidth = sumGridTemplatePixels(inlineStyle.gridTemplateColumns);
    if (gridWidth !== null) {
      return gridWidth;
    }
    return 224;
  }

  return element.getAttribute("role") === "row" ? 48 : 40;
}

function fallbackRect(element: HTMLElement) {
  const width = configuredLayoutSize(element, "width");
  const height = configuredLayoutSize(element, "height");
  return {
    bottom: height,
    height,
    left: 0,
    right: width,
    top: 0,
    width,
    x: 0,
    y: 0,
    toJSON() {
      return this;
    },
  } as DOMRect;
}

if (typeof HTMLElement !== "undefined") {
  const originalGetBoundingClientRect =
    HTMLElement.prototype.getBoundingClientRect;
  Object.defineProperty(HTMLElement.prototype, "getBoundingClientRect", {
    configurable: true,
    value(this: HTMLElement) {
      const measured = originalGetBoundingClientRect.call(this);
      if (measured.width > 0 || measured.height > 0) {
        return measured;
      }
      return fallbackRect(this);
    },
  });

  const installDimensionFallback = (
    property:
      | "clientHeight"
      | "clientWidth"
      | "offsetHeight"
      | "offsetWidth"
      | "scrollHeight"
      | "scrollWidth",
    axis: "height" | "width",
  ) => {
    const descriptor = Object.getOwnPropertyDescriptor(
      HTMLElement.prototype,
      property,
    );
    if (descriptor?.configurable === false) {
      return;
    }
    const originalGet = descriptor?.get;
    Object.defineProperty(HTMLElement.prototype, property, {
      configurable: true,
      get(this: HTMLElement) {
        const measured = originalGet?.call(this);
        if (typeof measured === "number" && measured > 0) {
          return measured;
        }
        return configuredLayoutSize(this, axis);
      },
    });
  };

  installDimensionFallback("clientHeight", "height");
  installDimensionFallback("clientWidth", "width");
  installDimensionFallback("offsetHeight", "height");
  installDimensionFallback("offsetWidth", "width");
  installDimensionFallback("scrollHeight", "height");
  installDimensionFallback("scrollWidth", "width");
}
