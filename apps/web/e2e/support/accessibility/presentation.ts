import type { Page } from "@playwright/test";

// Scope temporary styles to one observation, including its preparation failures.
export async function withPresentation<T>(
  page: Page,
  options: {
    zoom?: string;
    textSpacing?: string;
    visualSnapshot?: boolean;
    viewport?: { width: number; height: number };
  },
  observe: () => Promise<T>,
  reportCleanup: (errors: unknown[]) => Promise<void> = async (errors) => {
    const { test } = await import("@playwright/test");
    await test.info().attach("presentation-cleanup-failure", {
      body: JSON.stringify({ failures: errors.length }),
      contentType: "application/json",
    });
  },
): Promise<T> {
  const viewport = page.viewportSize();
  const before = await page.evaluate(() => ({
    zoom: document.documentElement.style.getPropertyValue("zoom"),
    priority: document.documentElement.style.getPropertyPriority("zoom"),
    visualSnapshot: document.documentElement.getAttribute(
      "data-visual-snapshot",
    ),
  }));
  let spacing: Awaited<ReturnType<Page["addStyleTag"]>> | undefined;
  let failed = false;
  let result: T | undefined;
  let cleanupErrors: unknown[] = [];
  try {
    if (options.viewport) await page.setViewportSize(options.viewport);
    await page.evaluate((options) => {
      if (options.zoom !== undefined)
        document.documentElement.style.setProperty("zoom", options.zoom);
      if (options.visualSnapshot !== undefined)
        document.documentElement.dataset.visualSnapshot = String(
          options.visualSnapshot,
        );
    }, options);
    if (options.textSpacing)
      spacing = await page.addStyleTag({ content: options.textSpacing });
    result = await observe();
  } catch (error) {
    failed = true;
    throw error;
  } finally {
    const errors: unknown[] = [];
    const restore = async (work: () => Promise<unknown>) => {
      try {
        await work();
      } catch (error) {
        errors.push(error);
      }
    };
    if (!page.isClosed()) {
      if (spacing)
        await restore(async () =>
          spacing?.evaluate((element) =>
            element.parentNode?.removeChild(element),
          ),
        );
      await restore(() =>
        page.evaluate((before) => {
          const root = document.documentElement;
          if (before.zoom)
            root.style.setProperty("zoom", before.zoom, before.priority);
          else root.style.removeProperty("zoom");
          if (before.visualSnapshot === null)
            root.removeAttribute("data-visual-snapshot");
          else root.setAttribute("data-visual-snapshot", before.visualSnapshot);
        }, before),
      );
      if (viewport) await restore(() => page.setViewportSize(viewport));
    }
    if (errors.length) {
      try {
        await reportCleanup(errors);
      } catch {
        /* Preserve the primary failure. */
      }
      cleanupErrors = errors;
    }
  }
  if (!failed && cleanupErrors.length)
    throw new AggregateError(
      cleanupErrors,
      "Temporary presentation restoration failed",
    );
  return result as T;
}
