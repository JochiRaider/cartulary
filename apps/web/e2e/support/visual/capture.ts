import {
  dataTestIdSelector,
  incidentControlsScrollportTestId,
} from "@cartulary/ui-contracts";
import { expect, type Locator, type Page, test } from "@playwright/test";
import { withPresentation } from "../accessibility/presentation";

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

import { workbookResponsiveBandTestId } from "@cartulary/ui-contracts";
import { emitVisualCaptureIntent } from "./captureIntent";
import { attachFontManifestDigest, waitForVendoredFonts } from "./fonts";
import {
  applyVisualMetadata,
  restoreVisualMetadata,
  verifyVisualMetadata,
} from "./metadataNormalization";
import { verifyTimelineCaptureData } from "./timelineVisualFixture";
import { attachVisualRenderDiagnostics } from "./workbookGeometry";

async function visualCaptureStep<T>(
  capture: string,
  stage: string,
  work: () => Promise<T>,
): Promise<T> {
  return await test.step(stage, async () => {
    const started = Date.now();
    const errorsBefore = test.info().errors.length;
    let failed = false;
    try {
      return await work();
    } catch (error) {
      failed = true;
      throw error;
    } finally {
      try {
        await test.info().attach(`${capture}-${stage}`, {
          body: JSON.stringify({
            stage,
            elapsed_ms: Date.now() - started,
            status:
              failed || test.info().errors.length > errorsBefore
                ? "fail"
                : "pass",
          }),
          contentType: "application/json",
        });
      } catch {
        /* Retain the primary failure if attachment cleanup also fails. */
      }
    }
  });
}

type CaptureOptions = {
  renderSurface?: string;
  anchor?: VisualAnchor;
  verifyFraming?: () => Promise<void>;
};
type CaptureMode =
  | {
      kind: "locator";
      locator: Locator;
      maxDiffPixels?: number;
      prepareState?: () => Promise<void>;
    }
  | {
      kind: "viewport";
      ready?: () => Promise<void>;
      mask?: Locator[];
      prepareAnchor?: () => Promise<void>;
    };

export async function assertVisualRegression(
  page: Page,
  name: string,
  locator = page.getByRole("main"),
  options: CaptureOptions & {
    maxDiffPixels?: number;
    prepareState?: () => Promise<void>;
  } = {},
) {
  return captureVisual(
    page,
    name,
    { kind: "locator", locator, ...options },
    options,
  );
}

export async function assertViewportVisualRegression(
  page: Page,
  name: string,
  options: CaptureOptions & {
    ready?: () => Promise<void>;
    mask?: Locator[];
    prepareAnchor?: () => Promise<void>;
  } = {},
) {
  return captureVisual(page, name, { kind: "viewport", ...options }, options);
}

async function captureVisual(
  page: Page,
  name: string,
  mode: CaptureMode,
  options: CaptureOptions,
) {
  return withPresentation(page, { visualSnapshot: true }, async () => {
    const started = Date.now();
    if (mode.kind === "locator") await expect(mode.locator).toBeVisible();
    else
      await visualCaptureStep(
        name,
        "scenario and drawer readiness",
        async () => {
          await mode.ready?.();
        },
      );
    await visualCaptureStep(name, "presentation and font readiness", () =>
      prepareVisualPresentation(page),
    );
    if (mode.kind === "locator")
      await visualCaptureStep(name, "prepare declared surface", async () => {
        await mode.prepareState?.();
      });
    const data = await verifyTimelineCaptureData(
      page,
      name,
      mode.kind === "locator" ? mode.locator : undefined,
    );
    if (data)
      await test.info().attach(`${name}-timeline-data`, {
        body: JSON.stringify(data),
        contentType: "application/json",
      });
    let failed = false;
    let cleanupError: unknown;
    try {
      const normalization = await visualCaptureStep(
        name,
        "admit scoped metadata normalization",
        () => applyVisualMetadata(page, name),
      );
      await test.info().attach(`${name}-metadata-normalization`, {
        body: JSON.stringify(normalization),
        contentType: "application/json",
      });
      await visualCaptureStep(name, "settle layout geometry", () =>
        settleVisualGeometry(
          page,
          mode.kind === "locator" ? options.anchor : undefined,
        ),
      );
      if (mode.kind === "viewport") await mode.prepareAnchor?.();
      await options.verifyFraming?.();
      if (mode.kind === "viewport") {
        const geometry = await visualCaptureStep(
          name,
          "establish and verify visual anchor",
          () => settleVisualGeometry(page, options.anchor),
        );
        await test.info().attach(`${name}-capture-geometry`, {
          body: JSON.stringify({
            stage: "anchored",
            elapsed_ms: Date.now() - started,
            geometry,
          }),
          contentType: "application/json",
        });
      }
      await attachVisualRenderDiagnostics(page, name, options.renderSurface);
      await visualCaptureStep(name, "validate capture evidence", () =>
        emitVisualCaptureIntent(
          page,
          name,
          `apps/web/e2e/support/visual/capture.ts#${mode.kind === "locator" ? "assertVisualRegression" : "assertViewportVisualRegression"}`,
        ),
      );
      // Soft comparisons retain subsequent capture intents for exact reconciliation.
      await visualCaptureStep(name, "compare visual screenshot", async () => {
        await verifyVisualMetadata(page);
        if (mode.kind === "locator") {
          await expect.soft(mode.locator).toHaveScreenshot(`${name}.png`, {
            animations: "disabled",
            caret: "hide",
            ...(mode.maxDiffPixels === undefined
              ? {}
              : { maxDiffPixels: mode.maxDiffPixels }),
          });
        } else {
          await expect.soft(page).toHaveScreenshot(`${name}.png`, {
            animations: "disabled",
            caret: "hide",
            fullPage: false,
            ...(mode.mask ? { mask: mode.mask } : {}),
          });
        }
        await verifyVisualMetadata(page);
        if (mode.kind === "locator") await options.verifyFraming?.();
        await verifyVisualGeometry(page, options.anchor);
        if (mode.kind === "viewport") await options.verifyFraming?.();
      });
    } catch (error) {
      failed = true;
      throw error;
    } finally {
      try {
        await restoreVisualMetadata(page);
      } catch (error) {
        try {
          await test.info().attach("metadata-cleanup-failure", {
            body: JSON.stringify({ failed: true }),
            contentType: "application/json",
          });
        } catch {
          /* Preserve the original failure. */
        }
        cleanupError = error;
      }
    }
    if (!failed && cleanupError) throw cleanupError;
  });
}

export async function prepareVisualPresentation(page: Page) {
  const responsiveBand = page.getByTestId(workbookResponsiveBandTestId());
  if ((await responsiveBand.count()) > 0) {
    const viewportWidth = await page.evaluate(() => {
      const viewportWidth = window.visualViewport?.width ?? window.innerWidth;
      const rootWidth = document.documentElement.clientWidth;
      const zoomValue =
        document.documentElement.style.zoom ||
        getComputedStyle(document.documentElement).zoom;
      const parsedZoom = Number.parseFloat(zoomValue);
      const zoomScale =
        !Number.isFinite(parsedZoom) || parsedZoom <= 0
          ? 1
          : zoomValue.trim().endsWith("%")
            ? parsedZoom / 100
            : parsedZoom;
      const zoomAdjustedWidth = viewportWidth / zoomScale;
      return rootWidth > 0
        ? Math.min(zoomAdjustedWidth, rootWidth)
        : zoomAdjustedWidth;
    });
    const expectedBand =
      viewportWidth >= 1280
        ? "base"
        : viewportWidth >= 1024
          ? "narrow_desktop"
          : viewportWidth >= 768
            ? "compact_desktop"
            : "below_supported_minimum";
    await expect(responsiveBand).toHaveAttribute(
      "data-workbook-responsive-band",
      expectedBand,
    );
  }
  await waitForVendoredFonts(page);
  await attachFontManifestDigest();
}
