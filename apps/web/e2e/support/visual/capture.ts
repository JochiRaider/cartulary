import {
  dataTestIdSelector,
  incidentControlsScrollportTestId,
} from "@cartulary/ui-contracts";
import { expect, type Locator, type Page, test } from "@playwright/test";

export type VisualAnchor = {
  locator: Locator;
  align: "start" | "center";
  focus?: boolean;
  outerScroll?: "drawer_end";
  scrollportSelector?: string;
};

export async function settleVisualGeometry(page: Page, anchor?: VisualAnchor) {
  const target = anchor?.locator ?? page.locator("body");
  await expect(target).toBeVisible();
  if (anchor) await page.mouse.move(0, 0);
  if (anchor?.focus) {
    await target.evaluate((element) =>
      (element as HTMLElement).focus({ preventScroll: true }),
    );
  }
  const selector =
    anchor?.scrollportSelector ??
    dataTestIdSelector(incidentControlsScrollportTestId());
  if (anchor) {
    await target.evaluate(
      (element, { selector, align, outerScroll }) => {
        const container = element.closest<HTMLElement>(selector);
        if (!container)
          throw new Error("visual anchor requires its declared scrollport");
        window.scrollTo({ left: 0, top: 0, behavior: "instant" });
        for (
          let outer = container.parentElement;
          outer;
          outer = outer.parentElement
        ) {
          outer.scrollLeft = 0;
          outer.scrollTop = 0;
        }
        container.scrollLeft = 0;
        container.scrollTop = 0;
        if (outerScroll === "drawer_end") {
          window.scrollTo({
            top: Math.max(
              0,
              container.getBoundingClientRect().bottom - window.innerHeight,
            ),
            left: 0,
            behavior: "instant",
          });
        }
        const box = element.getBoundingClientRect();
        const port = container.getBoundingClientRect();
        const style = getComputedStyle(container);
        const borderBoxHeight =
          Number.parseFloat(style.height) +
          (style.boxSizing === "border-box"
            ? 0
            : [
                style.paddingTop,
                style.paddingBottom,
                style.borderTopWidth,
                style.borderBottomWidth,
              ].reduce((sum, value) => sum + Number.parseFloat(value), 0));
        const zoom = port.height / borderBoxHeight;
        const visibleTop = Math.max(0, port.top + container.clientTop * zoom);
        const visibleBottom = Math.min(
          window.innerHeight,
          port.top + (container.clientTop + container.clientHeight) * zoom,
        );
        const top = (box.top - visibleTop) / zoom;
        container.scrollTop = Math.max(
          0,
          Math.min(
            container.scrollHeight - container.clientHeight,
            top -
              (align === "center"
                ? (visibleBottom - visibleTop - box.height) / zoom / 2
                : 0),
          ),
        );
      },
      { selector, align: anchor.align, outerScroll: anchor.outerScroll },
    );
  }
  return await verifyVisualGeometry(page, anchor);
}

// Observation only: a stable frame at the wrong anchor remains a failure.
export async function verifyVisualGeometry(page: Page, anchor?: VisualAnchor) {
  const target = anchor?.locator ?? page.locator("body");
  const selector =
    anchor?.scrollportSelector ??
    dataTestIdSelector(incidentControlsScrollportTestId());
  let observation: unknown;
  try {
    await expect
      .poll(
        async () => {
          const frames = await target.evaluate(
            async (element, { selector, align, focused }) => {
              const observe = () => {
                const container = element.closest<HTMLElement>(selector);
                const identity = (node: Element | null) => {
                  if (!node) return null;
                  const parts: string[] = [];
                  for (
                    let current: Element | null = node;
                    current;
                    current = current.parentElement
                  ) {
                    const testId = current.getAttribute("data-testid");
                    if (testId) {
                      parts.unshift(`testid:${JSON.stringify(testId)}`);
                      break;
                    }
                    const siblings = current.parentElement
                      ? Array.from(current.parentElement.children).filter(
                          (sibling) => sibling.tagName === current?.tagName,
                        )
                      : [current];
                    parts.unshift(
                      `${current.tagName.toLowerCase()}:nth-of-type(${siblings.indexOf(current) + 1})`,
                    );
                  }
                  return parts.join(" > ");
                };
                const rect = (node: Element) => {
                  const r = node.getBoundingClientRect();
                  return { x: r.x, y: r.y, width: r.width, height: r.height };
                };
                const box = element.getBoundingClientRect();
                const port = container?.getBoundingClientRect();
                let anchorCorrect = !align;
                if (align && container && port) {
                  const style = getComputedStyle(container);
                  const borderBoxHeight =
                    Number.parseFloat(style.height) +
                    (style.boxSizing === "border-box"
                      ? 0
                      : [
                          style.paddingTop,
                          style.paddingBottom,
                          style.borderTopWidth,
                          style.borderBottomWidth,
                        ].reduce(
                          (sum, value) => sum + Number.parseFloat(value),
                          0,
                        ));
                  const zoom = port.height / borderBoxHeight;
                  const visibleTop = Math.max(
                    0,
                    port.top + container.clientTop * zoom,
                  );
                  const visibleBottom = Math.min(
                    window.innerHeight,
                    port.top +
                      (container.clientTop + container.clientHeight) * zoom,
                  );
                  const contentTop =
                    (box.top - visibleTop) / zoom + container.scrollTop;
                  const desired = Math.max(
                    0,
                    Math.min(
                      container.scrollHeight - container.clientHeight,
                      contentTop -
                        (align === "center"
                          ? (visibleBottom - visibleTop - box.height) / zoom / 2
                          : 0),
                    ),
                  );
                  anchorCorrect =
                    Math.abs(container.scrollTop - desired) <= 1 &&
                    box.top < visibleBottom &&
                    box.bottom > visibleTop &&
                    (align !== "center" ||
                      (box.top >= visibleTop - 1 &&
                        box.bottom <= visibleBottom + 1));
                }
                return {
                  anchor: rect(element),
                  anchorIdentity: identity(element),
                  scrollport: container ? rect(container) : null,
                  scrollTop: container?.scrollTop ?? null,
                  scrollLeft: container?.scrollLeft ?? null,
                  scrollHeight: container?.scrollHeight ?? null,
                  documentScroll: [window.scrollX, window.scrollY],
                  activeElement: identity(document.activeElement),
                  focusCorrect: !focused || document.activeElement === element,
                  anchorCorrect,
                };
              };
              const frames = [];
              for (let frame = 0; frame < 3; frame++) {
                await new Promise<void>((resolve) =>
                  requestAnimationFrame(() => resolve()),
                );
                frames.push(observe());
              }
              return frames;
            },
            {
              selector,
              align: anchor?.align ?? null,
              focused: anchor?.focus ?? false,
            },
          );
          observation = frames[2];
          const first = JSON.stringify(frames[0]);
          return frames.every(
            (frame) =>
              frame.anchorCorrect &&
              frame.focusCorrect &&
              JSON.stringify(frame) === first,
          )
            ? 3
            : 0;
        },
        {
          message:
            "visual geometry must satisfy its anchor for three consecutive frames",
          intervals: [0],
        },
      )
      .toBeGreaterThanOrEqual(3);
  } catch (error) {
    try {
      await test.info().attach("failed-visual-geometry", {
        body: JSON.stringify(observation),
        contentType: "application/json",
      });
    } catch {
      /* Preserve the geometry failure. */
    }
    throw error;
  }
  return observation;
}
