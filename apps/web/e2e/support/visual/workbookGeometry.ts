import {
  dataTestIdPrefixSelector,
  dataTestIdSelector,
  gridRowGutterTestId,
  gridScrollportSelector,
  gridShellTestId,
  timelineInspectorTestId,
} from "@cartulary/ui-contracts";
import type { Page, TestInfo } from "@playwright/test";
import { expect, test } from "@playwright/test";
import { fontManifestSha256 } from "./fonts";

type GridVisualScrollLeft = "left" | "right" | number;

type GridVisualScrollState = {
  top: number;
  left: GridVisualScrollLeft;
};

export type WorkbookGridVisualScrollSnapshot = {
  shellTop: number;
  shellLeft: number;
  scrollportTop: number;
  scrollportLeft: number;
};

type GridVisualAnchor = {
  kind: "timelineEvidenceActions";
  rowId: string;
  top: number;
};

export function gridShellSelector(surface: string): string {
  return dataTestIdSelector(gridShellTestId(surface));
}

export async function attachVisualRenderDiagnostics(
  page: Page,
  name: string,
  surface?: string,
) {
  const browser = page.context().browser();
  const diagnostics = await readVisualRenderDiagnostics(page, surface);
  await test.info().attach(`${name}-render-diagnostics`, {
    body: JSON.stringify(
      {
        ...diagnostics,
        fontManifestSha256: fontManifestSha256(),
        playwright: {
          browserName: browser?.browserType().name() ?? null,
          browserVersion: browser?.version() ?? null,
          nodeArch: process.arch,
          nodePlatform: process.platform,
          nodeVersion: process.version,
          viewport: page.viewportSize(),
        },
      },
      null,
      2,
    ),
    contentType: "application/json",
  });
}

async function readVisualRenderDiagnostics(page: Page, surface?: string) {
  return page.evaluate(
    ({ chipSelector, scrollportSelector, shellSelector }) => {
      const round = (value: number) => Number(value.toFixed(2));
      const rectFor = (element: Element) => {
        const rect = element.getBoundingClientRect();
        return {
          bottom: round(rect.bottom),
          height: round(rect.height),
          left: round(rect.left),
          right: round(rect.right),
          top: round(rect.top),
          width: round(rect.width),
          x: round(rect.x),
          y: round(rect.y),
        };
      };
      const textBoundsFor = (element: HTMLElement) => {
        const range = document.createRange();
        range.selectNodeContents(element);
        const rect = range.getBoundingClientRect();
        range.detach();
        return {
          bottom: round(rect.bottom),
          height: round(rect.height),
          left: round(rect.left),
          right: round(rect.right),
          text: (element.textContent ?? "").replace(/\s+/g, " ").trim(),
          top: round(rect.top),
          width: round(rect.width),
        };
      };
      const fontStyleFor = (element: HTMLElement | null) => {
        if (element === null) {
          return null;
        }
        const style = getComputedStyle(element);
        return {
          fontFamily: style.fontFamily,
          fontFeatureSettings: style.fontFeatureSettings,
          fontKerning: style.fontKerning,
          fontSize: style.fontSize,
          fontStretch: style.fontStretch,
          fontStyle: style.fontStyle,
          fontSynthesis: style.fontSynthesis,
          fontVariantLigatures: style.fontVariantLigatures,
          fontVariationSettings: style.fontVariationSettings,
          fontWeight: style.fontWeight,
          letterSpacing: style.letterSpacing,
          lineHeight: style.lineHeight,
          textRendering: style.textRendering,
          webkitFontSmoothing: style.getPropertyValue("-webkit-font-smoothing"),
        };
      };
      const shell =
        shellSelector === null
          ? null
          : document.querySelector<HTMLElement>(shellSelector);
      const scrollport =
        shell?.querySelector<HTMLElement>(scrollportSelector) ?? null;
      const gridStyleSource = scrollport ?? shell;
      const gridTextSamples = shell
        ? Array.from(
            shell.querySelectorAll<HTMLElement>(
              '[role="columnheader"], [role="rowheader"], [role="gridcell"]',
            ),
          )
            .filter((element) => {
              const rect = element.getBoundingClientRect();
              return (
                rect.width > 0 &&
                rect.height > 0 &&
                (element.textContent ?? "").trim() !== ""
              );
            })
            .slice(0, 8)
            .map((element) => ({
              ariaColIndex: element.getAttribute("aria-colindex"),
              fieldKey: element.getAttribute("data-grid-field-key"),
              rect: rectFor(element),
              role: element.getAttribute("role"),
              testId: element.getAttribute("data-testid"),
              textBounds: textBoundsFor(element),
              typography: fontStyleFor(element),
            }))
        : [];
      const chipSamples = Array.from(
        document.querySelectorAll<HTMLElement>(chipSelector),
      )
        .filter((element) => {
          const rect = element.getBoundingClientRect();
          return rect.width > 0 && rect.height > 0;
        })
        .slice(0, 8)
        .map((element) => ({
          ariaLabel: element.getAttribute("aria-label"),
          rect: rectFor(element),
          role: element.getAttribute("role"),
          testId: element.getAttribute("data-testid"),
          textBounds: textBoundsFor(element),
          typography: fontStyleFor(element),
        }));
      const userAgentData =
        "userAgentData" in navigator
          ? {
              brands: (
                navigator as Navigator & {
                  userAgentData?: {
                    brands?: unknown;
                    mobile?: unknown;
                    platform?: unknown;
                  };
                }
              ).userAgentData?.brands,
              mobile: (
                navigator as Navigator & {
                  userAgentData?: {
                    brands?: unknown;
                    mobile?: unknown;
                    platform?: unknown;
                  };
                }
              ).userAgentData?.mobile,
              platform: (
                navigator as Navigator & {
                  userAgentData?: {
                    brands?: unknown;
                    mobile?: unknown;
                    platform?: unknown;
                  };
                }
              ).userAgentData?.platform,
            }
          : null;
      return {
        computed: {
          chip: fontStyleFor(
            chipSamples.length > 0
              ? document.querySelector<HTMLElement>(chipSelector)
              : null,
          ),
          gridCell: fontStyleFor(
            gridTextSamples.length > 0
              ? (shell?.querySelector<HTMLElement>(
                  '[role="gridcell"], [role="rowheader"], [role="columnheader"]',
                ) ?? null)
              : null,
          ),
          scrollport: fontStyleFor(scrollport),
          shell: fontStyleFor(shell),
        },
        cssVars: gridStyleSource
          ? {
              cellPadding: getComputedStyle(gridStyleSource)
                .getPropertyValue("--cartulary-grid-cell-padding")
                .trim(),
              density: getComputedStyle(gridStyleSource)
                .getPropertyValue("--cartulary-grid-density")
                .trim(),
              gridCellLineHeight: getComputedStyle(gridStyleSource)
                .getPropertyValue("--ct-typography-grid-cell-lineHeight")
                .trim(),
              rowHeight: getComputedStyle(gridStyleSource)
                .getPropertyValue("--cartulary-grid-row-height")
                .trim(),
            }
          : null,
        devicePixelRatio: window.devicePixelRatio,
        fontFaces: Array.from(document.fonts).map((face) => ({
          display: face.display,
          family: face.family,
          status: face.status,
          stretch: face.stretch,
          style: face.style,
          weight: face.weight,
        })),
        fontChecks: {
          inter400: document.fonts.check('400 12px "Inter"'),
          inter500: document.fonts.check('500 12px "Inter"'),
          jetBrainsMono400: document.fonts.check('400 12px "JetBrains Mono"'),
        },
        gridTextSamples,
        navigator: {
          hardwareConcurrency: navigator.hardwareConcurrency,
          language: navigator.language,
          languages: navigator.languages,
          platform: navigator.platform,
          userAgent: navigator.userAgent,
          userAgentData,
        },
        shell: shell ? rectFor(shell) : null,
        chipSamples,
        scrollport: scrollport ? rectFor(scrollport) : null,
        visualViewport:
          window.visualViewport === null
            ? null
            : {
                height: round(window.visualViewport.height),
                offsetLeft: round(window.visualViewport.offsetLeft),
                offsetTop: round(window.visualViewport.offsetTop),
                pageLeft: round(window.visualViewport.pageLeft),
                pageTop: round(window.visualViewport.pageTop),
                scale: window.visualViewport.scale,
                width: round(window.visualViewport.width),
              },
        viewport: {
          innerHeight: window.innerHeight,
          innerWidth: window.innerWidth,
          outerHeight: window.outerHeight,
          outerWidth: window.outerWidth,
          screenHeight: window.screen.height,
          screenWidth: window.screen.width,
        },
      };
    },
    {
      chipSelector: dataTestIdPrefixSelector("chip-"),
      scrollportSelector: gridScrollportSelector(),
      shellSelector: surface === undefined ? null : gridShellSelector(surface),
    },
  );
}

export async function normalizeWorkbookGridVisualState(
  page: Page,
  surface: string,
  options: GridVisualRegressionOptions,
) {
  if ("anchor" in options) {
    await normalizeWorkbookGridAnchorVisualState(page, surface, options.anchor);
    return;
  }
  const { scroll } = options;
  await setWorkbookGridScroll(page, surface, scroll);
  await waitForVisualLayoutFrame(page);
  const expected = await setWorkbookGridScroll(page, surface, scroll);
  await expect
    .poll(() => readWorkbookGridScroll(page, surface), {
      message: `Expected ${surface} grid visual scroll to normalize shell and scrollport state`,
    })
    .toEqual(expected);
  await expect
    .poll(
      async () => {
        const headerText = await page
          .getByTestId(gridShellTestId(surface))
          .locator('[role="columnheader"]')
          .allTextContents();
        return headerText.some((text) => text.trim() !== "");
      },
      {
        message: `Expected ${surface} grid visual headers to finish rendering after scroll normalization`,
      },
    )
    .toBe(true);
  await waitForVisualLayoutFrame(page);
}

async function normalizeWorkbookGridAnchorVisualState(
  page: Page,
  surface: string,
  anchor: GridVisualAnchor,
) {
  await setWorkbookGridAnchor(page, surface, anchor);
  await waitForVisualLayoutFrame(page);
  await expect
    .poll(
      async () => {
        const expected = await setWorkbookGridAnchor(page, surface, anchor);
        const state = await readWorkbookGridAnchorState(page, surface, anchor);
        return (
          state.ready &&
          state.diagnostics.scroll.shell.left === expected.shellLeft &&
          state.diagnostics.scroll.shell.top === expected.shellTop &&
          state.diagnostics.scroll.scrollport.left ===
            expected.scrollportLeft &&
          state.diagnostics.scroll.scrollport.top === expected.scrollportTop
        );
      },
      {
        message: `Expected ${surface} grid visual anchor ${anchor.kind} to reach stable geometry with normalized shell and scrollport state`,
        timeout: 6_000,
      },
    )
    .toBe(true);
  await waitForVisualLayoutFrame(page);
}

export async function waitForVisualLayoutFrame(page: Page) {
  await page.evaluate(() => {
    return new Promise<void>((resolve) => {
      requestAnimationFrame(() => {
        requestAnimationFrame(() => resolve());
      });
    });
  });
}

export async function setWorkbookGridScroll(
  page: Page,
  surface: string,
  scroll: GridVisualScrollState,
): Promise<WorkbookGridVisualScrollSnapshot> {
  return page.evaluate(
    ({ left, scrollportSelector, shellSelector, surface, top }) => {
      const shell = document.querySelector<HTMLElement>(shellSelector);
      if (shell === null) {
        throw new Error(`Expected ${surface} grid shell to exist`);
      }
      const scrollports = Array.from(
        shell.querySelectorAll<HTMLElement>(scrollportSelector),
      );
      if (scrollports.length !== 1 || scrollports[0] === undefined) {
        throw new Error(
          `Expected ${surface} grid shell to contain exactly one ${scrollportSelector} scrollport, received ${scrollports.length}`,
        );
      }
      const scrollport = scrollports[0];
      const maxLeft = Math.max(
        0,
        scrollport.scrollWidth - scrollport.clientWidth,
      );
      const maxTop = Math.max(
        0,
        scrollport.scrollHeight - scrollport.clientHeight,
      );
      const expectedLeft =
        left === "left" ? 0 : left === "right" ? maxLeft : left;
      const expectedTop = Math.min(Math.max(0, top), maxTop);
      shell.scrollTop = 0;
      shell.scrollLeft = 0;
      scrollport.scrollTop = expectedTop;
      scrollport.scrollLeft = Math.min(Math.max(0, expectedLeft), maxLeft);
      shell.scrollTop = 0;
      shell.scrollLeft = 0;
      return {
        shellTop: Math.round(shell.scrollTop),
        shellLeft: Math.round(shell.scrollLeft),
        scrollportTop: Math.round(scrollport.scrollTop),
        scrollportLeft: Math.round(scrollport.scrollLeft),
      };
    },
    {
      left: scroll.left,
      scrollportSelector: gridScrollportSelector(),
      shellSelector: gridShellSelector(surface),
      surface,
      top: scroll.top,
    },
  );
}

function buildWorkbookGridAnchorSelectors(
  surface: string,
  anchor: GridVisualAnchor,
) {
  switch (anchor.kind) {
    case "timelineEvidenceActions":
      return {
        fieldKeys: [],
        scrollTargetTestId: gridRowGutterTestId(surface, anchor.rowId),
        requiredTestIds: {
          rowGutter: gridRowGutterTestId(surface, anchor.rowId),
        },
      };
  }
}

async function setWorkbookGridAnchor(
  page: Page,
  surface: string,
  anchor: GridVisualAnchor,
): Promise<WorkbookGridVisualScrollSnapshot> {
  return page.evaluate(
    ({ anchor, scrollportSelector, selectors, shellSelector, surface }) => {
      const shell = document.querySelector<HTMLElement>(shellSelector);
      if (shell === null) {
        throw new Error(`Expected ${surface} grid shell to exist`);
      }
      const scrollports = Array.from(
        shell.querySelectorAll<HTMLElement>(scrollportSelector),
      );
      if (scrollports.length !== 1 || scrollports[0] === undefined) {
        throw new Error(
          `Expected ${surface} grid shell to contain exactly one ${scrollportSelector} scrollport, received ${scrollports.length}`,
        );
      }
      const scrollport = scrollports[0];
      const maxLeft = Math.max(
        0,
        scrollport.scrollWidth - scrollport.clientWidth,
      );
      const maxTop = Math.max(
        0,
        scrollport.scrollHeight - scrollport.clientHeight,
      );
      const expectedTop = Math.min(Math.max(0, anchor.top), maxTop);
      const byTestId = (testId: string) =>
        Array.from(shell.querySelectorAll<HTMLElement>("[data-testid]")).find(
          (element) => element.getAttribute("data-testid") === testId,
        ) ?? null;

      shell.scrollTop = 0;
      shell.scrollLeft = 0;
      scrollport.scrollTop = expectedTop;
      // The sticky evidence gutter is visible at every horizontal offset.
      // Establish its declared left-edge framing before locating the anchor.
      scrollport.scrollLeft = 0;
      const scrollTarget = byTestId(selectors.scrollTargetTestId);
      if (scrollTarget === null) {
        shell.scrollLeft = Math.max(0, shell.scrollWidth - shell.clientWidth);
        scrollport.scrollLeft = maxLeft;
      } else {
        scrollTarget.scrollIntoView({ block: "nearest", inline: "nearest" });
      }
      shell.scrollTop = 0;

      const requiredElements = Object.values(selectors.requiredTestIds)
        .map((testId) => byTestId(testId))
        .filter((element): element is HTMLElement => element !== null);
      if (requiredElements.length > 0) {
        const shellRect = shell.getBoundingClientRect();
        const leftMost = Math.min(
          ...requiredElements.map(
            (element) => element.getBoundingClientRect().left,
          ),
        );
        const rightMost = Math.max(
          ...requiredElements.map(
            (element) => element.getBoundingClientRect().right,
          ),
        );
        const padding = 8;
        if (leftMost < shellRect.left + padding) {
          shell.scrollLeft = Math.max(
            0,
            shell.scrollLeft - (shellRect.left + padding - leftMost),
          );
        } else if (rightMost > shellRect.right - padding) {
          shell.scrollLeft = Math.min(
            Math.max(0, shell.scrollWidth - shell.clientWidth),
            shell.scrollLeft + (rightMost - (shellRect.right - padding)),
          );
        }
      }
      shell.scrollTop = 0;

      return {
        shellTop: Math.round(shell.scrollTop),
        shellLeft: Math.round(shell.scrollLeft),
        scrollportTop: Math.round(scrollport.scrollTop),
        scrollportLeft: Math.round(scrollport.scrollLeft),
      };
    },
    {
      anchor,
      scrollportSelector: gridScrollportSelector(),
      selectors: buildWorkbookGridAnchorSelectors(surface, anchor),
      shellSelector: gridShellSelector(surface),
      surface,
    },
  );
}

async function readWorkbookGridAnchorState(
  page: Page,
  surface: string,
  anchor: GridVisualAnchor,
) {
  return page.evaluate(
    async ({
      anchor,
      inspectorSelector,
      scrollportSelector,
      selectors,
      shellSelector,
      surface,
    }) => {
      const readDiagnostics = () => {
        const shell = document.querySelector<HTMLElement>(shellSelector);
        if (shell === null) {
          throw new Error(`Expected ${surface} grid shell to exist`);
        }
        const scrollports = Array.from(
          shell.querySelectorAll<HTMLElement>(scrollportSelector),
        );
        if (scrollports.length !== 1 || scrollports[0] === undefined) {
          throw new Error(
            `Expected ${surface} grid shell to contain exactly one ${scrollportSelector} scrollport, received ${scrollports.length}`,
          );
        }
        const scrollport = scrollports[0];
        const shellRect = shell.getBoundingClientRect();
        const byTestId = (testId: string) =>
          Array.from(shell.querySelectorAll<HTMLElement>("[data-testid]")).find(
            (element) => element.getAttribute("data-testid") === testId,
          ) ?? null;
        const roundedRect = (element: HTMLElement | null) => {
          if (element === null) {
            return null;
          }
          const rect = element.getBoundingClientRect();
          const visible =
            rect.width > 0 &&
            rect.height > 0 &&
            rect.right <= shellRect.right - 1 &&
            rect.left >= shellRect.left + 1 &&
            rect.bottom >= shellRect.top + 1 &&
            rect.top <= shellRect.bottom - 1;
          return {
            bottom: Math.round(rect.bottom),
            height: Math.round(rect.height),
            left: Math.round(rect.left),
            right: Math.round(rect.right),
            top: Math.round(rect.top),
            visible,
            width: Math.round(rect.width),
          };
        };
        const requiredRects = Object.fromEntries(
          Object.entries(selectors.requiredTestIds).map(([key, testId]) => [
            key,
            roundedRect(byTestId(testId)),
          ]),
        );
        const visibleFieldKeys = Array.from(
          shell.querySelectorAll<HTMLElement>("[data-grid-field-key]"),
        )
          .filter((element) => {
            const rect = element.getBoundingClientRect();
            return (
              rect.width > 0 &&
              rect.height > 0 &&
              rect.right >= shellRect.left + 1 &&
              rect.left <= shellRect.right - 1 &&
              rect.bottom >= shellRect.top + 1 &&
              rect.top <= shellRect.bottom - 1
            );
          })
          .map((element) => element.getAttribute("data-grid-field-key") ?? "")
          .filter((fieldKey, index, fieldKeys) => {
            return fieldKey !== "" && fieldKeys.indexOf(fieldKey) === index;
          });
        const missingTestIds = Object.entries(selectors.requiredTestIds)
          .filter(([, testId]) => byTestId(testId) === null)
          .map(([key]) => key);
        const hiddenTestIds = Object.entries(requiredRects)
          .filter(([, rect]) => rect === null || !rect.visible)
          .map(([key]) => key);
        const missingFieldKeys = selectors.fieldKeys.filter(
          (fieldKey) => !visibleFieldKeys.includes(fieldKey),
        );
        return {
          activeElementTestId:
            document.activeElement instanceof HTMLElement
              ? document.activeElement.getAttribute("data-testid")
              : null,
          anchorKind: anchor.kind,
          inspectorOpen: document.querySelector(inspectorSelector) !== null,
          missingFieldKeys,
          missingTestIds,
          ready:
            missingTestIds.length === 0 &&
            hiddenTestIds.length === 0 &&
            missingFieldKeys.length === 0,
          requiredRects,
          screenshotTargetTestId: `${surface}-grid-shell`,
          scroll: {
            shell: {
              clientHeight: shell.clientHeight,
              clientWidth: shell.clientWidth,
              left: Math.round(shell.scrollLeft),
              maxLeft: Math.max(0, shell.scrollWidth - shell.clientWidth),
              maxTop: Math.max(0, shell.scrollHeight - shell.clientHeight),
              scrollHeight: shell.scrollHeight,
              scrollWidth: shell.scrollWidth,
              top: Math.round(shell.scrollTop),
            },
            scrollport: {
              clientHeight: scrollport.clientHeight,
              clientWidth: scrollport.clientWidth,
              left: Math.round(scrollport.scrollLeft),
              maxLeft: Math.max(
                0,
                scrollport.scrollWidth - scrollport.clientWidth,
              ),
              maxTop: Math.max(
                0,
                scrollport.scrollHeight - scrollport.clientHeight,
              ),
              scrollHeight: scrollport.scrollHeight,
              scrollWidth: scrollport.scrollWidth,
              top: Math.round(scrollport.scrollTop),
            },
          },
          surface,
          visibleFieldKeys,
        };
      };
      const nextFrame = () =>
        new Promise<void>((resolve) => {
          requestAnimationFrame(() => resolve());
        });
      const samples = [readDiagnostics()];
      await nextFrame();
      samples.push(readDiagnostics());
      await nextFrame();
      samples.push(readDiagnostics());
      const signature = (sample: (typeof samples)[number]) =>
        JSON.stringify({
          rects: sample.requiredRects,
          scroll: sample.scroll,
          visibleFieldKeys: sample.visibleFieldKeys,
        });
      const firstSample = samples[0];
      if (firstSample === undefined) {
        throw new Error("Expected grid visual anchor diagnostics to sample");
      }
      const lastSample = samples[samples.length - 1];
      if (lastSample === undefined) {
        throw new Error(
          "Expected grid visual anchor diagnostics to retain a final sample",
        );
      }
      return {
        diagnostics: lastSample,
        ready:
          samples.every((sample) => sample.ready) &&
          samples.every(
            (sample) => signature(sample) === signature(firstSample),
          ),
        samples,
      };
    },
    {
      anchor,
      inspectorSelector: dataTestIdSelector(timelineInspectorTestId()),
      scrollportSelector: gridScrollportSelector(),
      selectors: buildWorkbookGridAnchorSelectors(surface, anchor),
      shellSelector: gridShellSelector(surface),
      surface,
    },
  );
}

export async function attachWorkbookGridVisualDiagnostics(
  page: Page,
  name: string,
  surface: string,
  options: GridVisualRegressionOptions,
) {
  const testInfo = options.testInfo ?? test.info();
  const diagnostics =
    "anchor" in options
      ? await readWorkbookGridAnchorState(page, surface, options.anchor)
      : await readWorkbookGridDiagnostics(page, surface);
  await testInfo.attach(`${name}-grid-diagnostics`, {
    body: JSON.stringify(
      {
        ...diagnostics,
        presentation: await readWorkbookGridPresentationDiagnostics(
          page,
          surface,
        ),
      },
      null,
      2,
    ),
    contentType: "application/json",
  });
}

async function readWorkbookGridPresentationDiagnostics(
  page: Page,
  surface: string,
) {
  return page.evaluate(
    ({ chipSelector, scrollportSelector, shellSelector, surface }) => {
      const round = (value: number) => Number(value.toFixed(2));
      const rectFor = (element: Element) => {
        const rect = element.getBoundingClientRect();
        return {
          bottom: round(rect.bottom),
          height: round(rect.height),
          left: round(rect.left),
          right: round(rect.right),
          top: round(rect.top),
          width: round(rect.width),
        };
      };
      const typographyFor = (element: HTMLElement | null) => {
        if (element === null) {
          return null;
        }
        const style = getComputedStyle(element);
        return {
          fontFamily: style.fontFamily,
          fontSize: style.fontSize,
          fontWeight: style.fontWeight,
          letterSpacing: style.letterSpacing,
          lineHeight: style.lineHeight,
          paddingBlockEnd: style.paddingBlockEnd,
          paddingBlockStart: style.paddingBlockStart,
          paddingInlineEnd: style.paddingInlineEnd,
          paddingInlineStart: style.paddingInlineStart,
        };
      };
      const shell = document.querySelector<HTMLElement>(shellSelector);
      if (shell === null) {
        throw new Error(`Expected ${surface} grid shell to exist`);
      }
      const scrollport = shell.querySelector<HTMLElement>(scrollportSelector);
      if (scrollport === null) {
        throw new Error(
          `Expected ${surface} grid shell to contain ${scrollportSelector}`,
        );
      }
      const gridStyle = getComputedStyle(scrollport);
      const rowHeightVar = gridStyle
        .getPropertyValue("--cartulary-grid-row-height")
        .trim();
      const firstCell = shell.querySelector<HTMLElement>(
        '[role="gridcell"], [role="rowheader"], [role="columnheader"]',
      );
      const firstDataCell =
        shell.querySelector<HTMLElement>(
          '[data-grid-record-id] [role="gridcell"], [data-grid-record-id] [role="rowheader"]',
        ) ?? firstCell;
      const chipBounds = Array.from(
        shell.querySelectorAll<HTMLElement>(chipSelector),
      )
        .filter((element) => {
          const rect = element.getBoundingClientRect();
          return rect.width > 0 && rect.height > 0;
        })
        .slice(0, 8)
        .map((element) => ({
          ariaLabel: element.getAttribute("aria-label"),
          rect: rectFor(element),
          testId: element.getAttribute("data-testid"),
          text: (element.textContent ?? "").replace(/\s+/g, " ").trim(),
          typography: typographyFor(element),
        }));
      return {
        chipBounds,
        computedLineHeight: firstCell
          ? getComputedStyle(firstCell).lineHeight
          : null,
        computedRowHeight:
          rowHeightVar === "" ? null : round(Number.parseFloat(rowHeightVar)),
        densityId: gridStyle
          .getPropertyValue("--cartulary-grid-density")
          .trim(),
        gridCell: typographyFor(firstCell),
        observedFirstCellHeight: firstDataCell
          ? round(firstDataCell.getBoundingClientRect().height)
          : null,
        rowHeightVar,
        cellPaddingVar: gridStyle
          .getPropertyValue("--cartulary-grid-cell-padding")
          .trim(),
        tokenGridCellLineHeight: gridStyle
          .getPropertyValue("--ct-typography-grid-cell-lineHeight")
          .trim(),
      };
    },
    {
      chipSelector: dataTestIdPrefixSelector("chip-"),
      scrollportSelector: gridScrollportSelector(),
      shellSelector: gridShellSelector(surface),
      surface,
    },
  );
}

async function readWorkbookGridDiagnostics(page: Page, surface: string) {
  return page.evaluate(
    ({ inspectorSelector, scrollportSelector, shellSelector, surface }) => {
      const shell = document.querySelector<HTMLElement>(shellSelector);
      if (shell === null) {
        throw new Error(`Expected ${surface} grid shell to exist`);
      }
      const scrollports = Array.from(
        shell.querySelectorAll<HTMLElement>(scrollportSelector),
      );
      if (scrollports.length !== 1 || scrollports[0] === undefined) {
        throw new Error(
          `Expected ${surface} grid shell to contain exactly one ${scrollportSelector} scrollport, received ${scrollports.length}`,
        );
      }
      const scrollport = scrollports[0];
      const scrollportRect = scrollport.getBoundingClientRect();
      const visibleFieldKeys = Array.from(
        shell.querySelectorAll<HTMLElement>("[data-grid-field-key]"),
      )
        .filter((element) => {
          const rect = element.getBoundingClientRect();
          return (
            rect.width > 0 &&
            rect.height > 0 &&
            rect.right >= scrollportRect.left + 1 &&
            rect.left <= scrollportRect.right - 1 &&
            rect.bottom >= scrollportRect.top + 1 &&
            rect.top <= scrollportRect.bottom - 1
          );
        })
        .map((element) => element.getAttribute("data-grid-field-key") ?? "")
        .filter((fieldKey, index, fieldKeys) => {
          return fieldKey !== "" && fieldKeys.indexOf(fieldKey) === index;
        });
      return {
        activeElementTestId:
          document.activeElement instanceof HTMLElement
            ? document.activeElement.getAttribute("data-testid")
            : null,
        inspectorOpen: document.querySelector(inspectorSelector) !== null,
        ready: true,
        requiredRects: {},
        screenshotTargetTestId: `${surface}-grid-shell`,
        scroll: {
          shell: {
            clientHeight: shell.clientHeight,
            clientWidth: shell.clientWidth,
            left: Math.round(shell.scrollLeft),
            maxLeft: Math.max(0, shell.scrollWidth - shell.clientWidth),
            maxTop: Math.max(0, shell.scrollHeight - shell.clientHeight),
            scrollHeight: shell.scrollHeight,
            scrollWidth: shell.scrollWidth,
            top: Math.round(shell.scrollTop),
          },
          scrollport: {
            clientHeight: scrollport.clientHeight,
            clientWidth: scrollport.clientWidth,
            left: Math.round(scrollport.scrollLeft),
            maxLeft: Math.max(
              0,
              scrollport.scrollWidth - scrollport.clientWidth,
            ),
            maxTop: Math.max(
              0,
              scrollport.scrollHeight - scrollport.clientHeight,
            ),
            scrollHeight: scrollport.scrollHeight,
            scrollWidth: scrollport.scrollWidth,
            top: Math.round(scrollport.scrollTop),
          },
        },
        surface,
        visibleFieldKeys,
      };
    },
    {
      inspectorSelector: dataTestIdSelector(timelineInspectorTestId()),
      scrollportSelector: gridScrollportSelector(),
      shellSelector: gridShellSelector(surface),
      surface,
    },
  );
}

export async function readWorkbookGridScroll(
  page: Page,
  surface: string,
): Promise<WorkbookGridVisualScrollSnapshot> {
  return page.evaluate(
    ({ scrollportSelector, shellSelector, surface }) => {
      const shell = document.querySelector<HTMLElement>(shellSelector);
      if (shell === null) {
        throw new Error(`Expected ${surface} grid shell to exist`);
      }
      const scrollports = Array.from(
        shell.querySelectorAll<HTMLElement>(scrollportSelector),
      );
      if (scrollports.length !== 1 || scrollports[0] === undefined) {
        throw new Error(
          `Expected ${surface} grid shell to contain exactly one ${scrollportSelector} scrollport, received ${scrollports.length}`,
        );
      }
      return {
        shellTop: Math.round(shell.scrollTop),
        shellLeft: Math.round(shell.scrollLeft),
        scrollportTop: Math.round(scrollports[0].scrollTop),
        scrollportLeft: Math.round(scrollports[0].scrollLeft),
      };
    },
    {
      scrollportSelector: gridScrollportSelector(),
      shellSelector: gridShellSelector(surface),
      surface,
    },
  );
}

export type GridVisualRegressionOptions = {
  maxDiffPixels?: number;
  testInfo?: TestInfo;
} & (
  | { scroll: GridVisualScrollState; anchor?: never }
  | { anchor: GridVisualAnchor; scroll?: never }
);
