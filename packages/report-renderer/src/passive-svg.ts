// Runs inside the renderer's browser. This constructs the release representation
// from the pinned engine's DOM; it is not the final SVG admission validator. The
// Reporting owner still validates the serialized bytes independently.
export function materializePassiveSVG(selector: string): string {
  const candidate = document.querySelector(selector);
  if (!(candidate instanceof SVGSVGElement))
    throw new Error("missing_svg_root");
  const root: SVGSVGElement = candidate;
  const namespace = "http://www.w3.org/2000/svg";
  const allowed = new Set([
    "svg",
    "g",
    "defs",
    "rect",
    "circle",
    "ellipse",
    "line",
    "polyline",
    "polygon",
    "path",
    "text",
    "tspan",
    "title",
    "desc",
  ]);
  const geometry = new Set([
    "x",
    "y",
    "dx",
    "dy",
    "x1",
    "y1",
    "x2",
    "y2",
    "cx",
    "cy",
    "r",
    "rx",
    "ry",
    "width",
    "height",
    "viewBox",
    "d",
    "points",
    "transform",
  ]);
  const presentation = [
    "fill",
    "stroke",
    "stroke-width",
    "font-family",
    "font-size",
    "font-weight",
    "text-anchor",
    "dominant-baseline",
    "opacity",
    "fill-opacity",
    "stroke-opacity",
  ];
  const node = (name: string) => document.createElementNS(namespace, name);
  const finite = (value: number) => {
    if (!Number.isFinite(value)) throw new Error("invalid_svg_geometry");
    return Object.is(value, -0) ? "0" : String(value);
  };
  const paint = (value: string) => {
    if (value === "none") return { color: value, alpha: 1 };
    const rgb = /^rgba?\((\d+), (\d+), (\d+)(?:, ([0-9.]+))?\)$/.exec(value);
    if (!rgb) throw new Error("unsupported_svg_paint");
    const alpha = rgb[4] === undefined ? 1 : Number(rgb[4]);
    if (!Number.isFinite(alpha) || alpha < 0 || alpha > 1)
      throw new Error("unsupported_svg_paint");
    return {
      color: `#${rgb
        .slice(1, 4)
        .map((part) => Number(part).toString(16).padStart(2, "0"))
        .join("")}`,
      alpha,
    };
  };
  const copyPresentation = (source: Element, target: Element) => {
    const style = getComputedStyle(source);
    if (
      style.filter !== "none" ||
      style.clipPath !== "none" ||
      style.maskImage !== "none"
    )
      throw new Error("unsupported_svg_effect");
    for (const attribute of presentation) {
      const value = style.getPropertyValue(attribute);
      if (!value) continue;
      if (attribute !== "fill" && attribute !== "stroke")
        target.setAttribute(attribute, value);
    }
    for (const attribute of ["fill", "stroke"]) {
      const { color, alpha } = paint(style.getPropertyValue(attribute));
      target.setAttribute(attribute, color);
      target.setAttribute(
        `${attribute}-opacity`,
        finite(alpha * Number(style.getPropertyValue(`${attribute}-opacity`))),
      );
    }
    return style;
  };

  function marker(
    source: SVGGeometryElement,
    position: "start" | "end",
    value: string,
  ): SVGElement | null {
    if (value === "none") return null;
    // The browser expands fragment URLs to the local document URL. Require that
    // exact document before resolving an engine-owned marker; never fetch it.
    const match = /^url\("?([^"()]+)"?\)$/.exec(value);
    if (!match?.[1]) throw new Error("invalid_svg_marker");
    const url = new URL(match[1], document.URL);
    const documentURL = new URL(document.URL);
    if (
      url.origin !== documentURL.origin ||
      url.pathname !== documentURL.pathname ||
      url.search !== documentURL.search ||
      !url.hash
    )
      throw new Error("external_svg_marker");
    const original = [...root.querySelectorAll("marker")].find(
      (item) => item.id === url.hash.slice(1),
    );
    if (
      !(original instanceof SVGMarkerElement) ||
      !["auto", "auto-start-reverse"].includes(
        original.getAttribute("orient") ?? "",
      )
    )
      throw new Error("unsupported_svg_marker");
    const length = source.getTotalLength();
    if (!(length > 0)) throw new Error("invalid_svg_marker_path");
    const point = source.getPointAtLength(position === "start" ? 0 : length);
    const near = source.getPointAtLength(
      position === "start"
        ? Math.min(length, 0.01)
        : Math.max(0, length - 0.01),
    );
    let angle =
      (Math.atan2(
        position === "start" ? near.y - point.y : point.y - near.y,
        position === "start" ? near.x - point.x : point.x - near.x,
      ) *
        180) /
      Math.PI;
    if (
      position === "start" &&
      original.getAttribute("orient") === "auto-start-reverse"
    )
      angle += 180;
    const box = original.viewBox.baseVal;
    const width = original.markerWidth.baseVal.value;
    const height = original.markerHeight.baseVal.value;
    const refX = original.refX.baseVal.value;
    const refY = original.refY.baseVal.value;
    if (
      !(width > 0 && height > 0) ||
      (original.hasAttribute("viewBox") && !(box.width > 0 && box.height > 0))
    )
      throw new Error("unsupported_svg_marker_viewport");
    const units = original.markerUnits.baseVal;
    const strokeScale =
      units === SVGMarkerElement.SVG_MARKERUNITS_STROKEWIDTH
        ? Number.parseFloat(getComputedStyle(source).strokeWidth)
        : units === SVGMarkerElement.SVG_MARKERUNITS_USERSPACEONUSE
          ? 1
          : Number.NaN;
    // refX/refY are in the marker's content coordinate system. Viewport
    // alignment translations cancel when that reference point is anchored at
    // the edge endpoint; the viewBox scale must still be materialized.
    let scaleX = original.hasAttribute("viewBox") ? width / box.width : 1;
    let scaleY = original.hasAttribute("viewBox") ? height / box.height : 1;
    if (
      original.preserveAspectRatio.baseVal.align !==
      SVGPreserveAspectRatio.SVG_PRESERVEASPECTRATIO_NONE
    ) {
      const uniform =
        original.preserveAspectRatio.baseVal.meetOrSlice ===
        SVGPreserveAspectRatio.SVG_MEETORSLICE_SLICE
          ? Math.max(scaleX, scaleY)
          : Math.min(scaleX, scaleY);
      scaleX = uniform;
      scaleY = uniform;
    }
    const output = node("g");
    output.setAttribute(
      "transform",
      `translate(${finite(point.x)} ${finite(point.y)}) rotate(${finite(angle)}) scale(${finite(strokeScale * scaleX)} ${finite(strokeScale * scaleY)}) translate(${finite(-refX)} ${finite(-refY)})`,
    );
    for (const child of original.children) output.append(convert(child));
    return output;
  }

  function convert(source: Element): SVGElement {
    if (source.namespaceURI !== namespace || !allowed.has(source.localName))
      throw new Error("unsupported_svg_element");
    const output = node(source.localName);
    for (const attribute of source.attributes) {
      if (
        attribute.localName.toLowerCase().startsWith("on") ||
        ["href", "src"].includes(attribute.localName)
      )
        throw new Error("active_svg_attribute");
      if (geometry.has(attribute.name)) {
        if (
          source instanceof SVGTextPositioningElement &&
          ["x", "y", "dx", "dy"].includes(attribute.name)
        ) {
          const lengths = new Map([
            ["x", source.x],
            ["y", source.y],
            ["dx", source.dx],
            ["dy", source.dy],
          ]).get(attribute.name)?.baseVal;
          if (!lengths || lengths.numberOfItems !== 1)
            throw new Error("unsupported_svg_text_position");
          output.setAttribute(attribute.name, finite(lengths.getItem(0).value));
        } else output.setAttribute(attribute.name, attribute.value);
      }
    }
    const style = copyPresentation(source, output);
    if (style.display === "none" || style.visibility !== "visible")
      throw new Error("unsupported_svg_visibility");
    for (const child of source.childNodes) {
      if (child.nodeType === Node.TEXT_NODE) {
        if (["text", "tspan", "title", "desc"].includes(source.localName))
          output.append(document.createTextNode(child.textContent ?? ""));
        else if (child.textContent?.trim())
          throw new Error("unexpected_svg_text");
      } else if (child instanceof Element) {
        // Styles are evaluated above. Definitions are consumed only through the
        // closed marker conversion below; no references survive serialization.
        if (
          child.localName === "style" ||
          child.localName === "defs" ||
          child.localName === "marker"
        )
          continue;
        output.append(convert(child));
      } else throw new Error("unsupported_svg_node");
    }
    if (source instanceof SVGGeometryElement) {
      if (style.markerMid !== "none")
        throw new Error("unsupported_svg_mid_marker");
      let content: SVGElement = output;
      const pattern =
        style.strokeDasharray === "none"
          ? []
          : style.strokeDasharray
              .split(",")
              .map((value) => Number.parseFloat(value));
      if (pattern.some((value) => !Number.isFinite(value) || value < 0))
        throw new Error("unsupported_svg_dash_pattern");
      if (pattern.length % 2) pattern.push(...pattern);
      // Zero-length gaps (including Mermaid marker strokes) are solid lines.
      if (pattern.some((value, index) => index % 2 === 1 && value > 0)) {
        if (Number.parseFloat(style.strokeDashoffset) !== 0)
          throw new Error("unsupported_svg_dash_offset");
        type Point = { x: number; y: number };
        let points: Point[];
        if (source instanceof SVGLineElement) {
          points = [
            { x: source.x1.baseVal.value, y: source.y1.baseVal.value },
            { x: source.x2.baseVal.value, y: source.y2.baseVal.value },
          ];
        } else if (source instanceof SVGPathElement) {
          // The admitted flowchart profile selects linear routing. Preserve all
          // vertices exactly; never approximate an unknown curved path.
          const data = source.getAttribute("d") ?? "";
          const number =
            "[-+]?(?:[0-9]+(?:\\.[0-9]*)?|\\.[0-9]+)(?:[eE][-+]?[0-9]+)?";
          const command = new RegExp(
            `([ML])\\s*(${number})[ ,]+(${number})`,
            "gy",
          );
          points = [];
          let offset = 0;
          while (offset < data.length) {
            while (data[offset] === " ") offset += 1;
            command.lastIndex = offset;
            const part = command.exec(data);
            if (!part || part[1] !== (points.length ? "L" : "M"))
              throw new Error("unsupported_svg_dash_geometry");
            points.push({ x: Number(part[2]), y: Number(part[3]) });
            offset = command.lastIndex;
          }
        } else throw new Error("unsupported_svg_dash_geometry");
        if (points.length < 2) throw new Error("unsupported_svg_dash_geometry");
        const distances = [0];
        for (let index = 1; index < points.length; index += 1) {
          const a = points[index - 1];
          const b = points[index];
          if (!a || !b) throw new Error("invalid_svg_geometry");
          distances.push(
            (distances[index - 1] ?? 0) + Math.hypot(b.x - a.x, b.y - a.y),
          );
        }
        const length = distances.at(-1) ?? 0;
        const pointAt = (distance: number): Point => {
          for (let index = 1; index < points.length; index += 1) {
            const end = distances[index] ?? 0;
            const start = distances[index - 1] ?? 0;
            const a = points[index - 1];
            const b = points[index];
            if (a && b && end > start && distance <= end) {
              const fraction = (distance - start) / (end - start);
              return {
                x: a.x + (b.x - a.x) * fraction,
                y: a.y + (b.y - a.y) * fraction,
              };
            }
          }
          const last = points.at(-1);
          if (!last) throw new Error("invalid_svg_geometry");
          return last;
        };
        const group = node("g");
        const transform = source.getAttribute("transform");
        if (transform) group.setAttribute("transform", transform);
        let cursor = 0;
        let index = 0;
        while (cursor < length) {
          const end = Math.min(
            length,
            cursor + (pattern[index % pattern.length] ?? 0),
          );
          if (index % 2 === 0 && end > cursor) {
            const a = pointAt(cursor);
            const b = pointAt(end);
            const path = node("path");
            copyPresentation(source, path);
            path.setAttribute("fill", "none");
            const retained = points.filter(
              (_, index) =>
                (distances[index] ?? 0) > cursor &&
                (distances[index] ?? 0) < end,
            );
            path.setAttribute(
              "d",
              `M${finite(a.x)} ${finite(a.y)}` +
                [...retained, b]
                  .map((point) => `L${finite(point.x)} ${finite(point.y)}`)
                  .join(""),
            );
            group.append(path);
          }
          cursor = end;
          index += 1;
          if (index > 10000) throw new Error("svg_dash_limit");
        }
        content = group;
      }
      const start = marker(source, "start", style.markerStart);
      const end = marker(source, "end", style.markerEnd);
      if (start || end) {
        const group = node("g");
        const transform = source.getAttribute("transform");
        if (transform) {
          group.setAttribute("transform", transform);
          content.removeAttribute("transform");
        }
        group.append(content);
        if (start) group.append(start);
        if (end) group.append(end);
        return group;
      }
      return content;
    }
    return output;
  }
  const output = convert(root);
  output.setAttribute("xmlns", namespace);
  // Stable attribute order and XML escaping are part of the output contract.
  for (const element of [output, ...output.querySelectorAll("*")]) {
    const attributes = [...element.attributes]
      .map(({ name, value }) => [name, value] as const)
      .sort(([a], [b]) => (a < b ? -1 : a > b ? 1 : 0));
    for (const attribute of [...element.attributes])
      element.removeAttribute(attribute.name);
    for (const [name, value] of attributes) element.setAttribute(name, value);
  }
  return `${new XMLSerializer().serializeToString(output)}\n`;
}
