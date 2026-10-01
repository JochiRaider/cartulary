import {
  gridRowGutterTestId,
  gridScrollportSelector,
  gridShellTestId,
  rowCellTestId,
  saveStateTestId,
  timelineInspectorTestId,
  workbookShellSlotTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import { expect, type Page } from "@playwright/test";

export async function focusTimelineShellOrigin(page: Page, recordId: string) {
  const grid = page.getByTestId(gridShellTestId(timelineViewSchemaId));
  await grid.evaluate((element, scrollport) => {
    window.scrollTo({ top: 0, left: 0, behavior: "instant" });
    for (
      let node: HTMLElement | null = element as HTMLElement;
      node;
      node = node.parentElement
    ) {
      node.scrollLeft = 0;
      node.scrollTop = 0;
    }
    const port = element.querySelector<HTMLElement>(scrollport);
    if (!port) throw new Error("Missing Timeline scrollport");
    port.scrollLeft = 0;
    port.scrollTop = 0;
  }, gridScrollportSelector());
  const content = page.getByTestId(
    rowCellTestId(recordId, "timeline.date_entered_text"),
  );
  await expect(content).toBeVisible();
  await content.evaluate((element) => {
    const cell = element.closest<HTMLElement>('[role="gridcell"]');
    if (!cell) throw new Error("Missing committed Date Entered cell");
    cell.focus({ preventScroll: true });
  });
}

export async function observeTimelineShellOrigin(page: Page, recordId: string) {
  const ids = {
    target: rowCellTestId(recordId, "timeline.date_entered_text"),
    inspector: timelineInspectorTestId(),
    save: saveStateTestId(),
    landmarks: [
      ...(["top-bar", "view-bar", "primary-grid", "status-strip"] as const).map(
        workbookShellSlotTestId,
      ),
      gridRowGutterTestId(timelineViewSchemaId, recordId),
    ],
  };
  return page.evaluate(
    async ({ ids, recordId }) => {
      const find = (id: string) =>
        document.querySelector<HTMLElement>(
          `[data-testid=${JSON.stringify(id)}]`,
        );
      const rect = (node: Element | null) => {
        if (!node) return null;
        const r = node.getBoundingClientRect();
        return { x: r.x, y: r.y, width: r.width, height: r.height };
      };
      const visible = (node: Element | null) => {
        const r = node?.getBoundingClientRect();
        return (
          !!r &&
          r.width > 0 &&
          r.height > 0 &&
          r.left >= 0 &&
          r.top >= 0 &&
          r.right <= innerWidth &&
          r.bottom <= innerHeight
        );
      };
      const frames: {
        cell: ReturnType<typeof rect>;
        landmarks: ReturnType<typeof rect>[];
        offsets: number[][];
        valid: boolean;
        fonts: boolean;
        selected: boolean;
        focused: boolean;
        save: string | null;
      }[] = [];
      for (let i = 0; i < 3; i++) {
        await new Promise<void>((resolve) =>
          requestAnimationFrame(() => resolve()),
        );
        const element = find(ids.target);
        const cell = element?.closest<HTMLElement>('[role="gridcell"]') ?? null;
        const row = element?.closest<HTMLElement>("[data-grid-record-id]");
        const offsets = [[window.scrollX, window.scrollY]];
        for (
          let parent = cell?.parentElement;
          parent;
          parent = parent.parentElement
        )
          offsets.push([parent.scrollLeft, parent.scrollTop]);
        const fonts = ["Inter", "JetBrains Mono"].every((family) =>
          Array.from(document.fonts).some(
            (face) =>
              face.family.replaceAll('"', "") === family &&
              face.status === "loaded",
          ),
        );
        frames.push({
          cell: rect(cell),
          landmarks: ids.landmarks.map((id) => rect(find(id))),
          offsets,
          fonts,
          selected: row?.getAttribute("data-inspector-active") === "true",
          focused: document.activeElement === cell,
          save: find(ids.save)?.textContent ?? null,
          valid:
            !!cell &&
            visible(cell) &&
            ids.landmarks.every((id) => {
              const r = find(id)?.getBoundingClientRect();
              return (
                !!r &&
                r.width > 0 &&
                r.height > 0 &&
                r.right > 0 &&
                r.bottom > 0 &&
                r.left < innerWidth &&
                r.top < innerHeight
              );
            }) &&
            fonts &&
            document.fonts.status === "loaded" &&
            document.activeElement === cell &&
            row?.dataset.gridRecordId === recordId &&
            row.getAttribute("data-inspector-active") === "true" &&
            offsets.every(([x, y]) => x === 0 && y === 0) &&
            !find(ids.inspector) &&
            find(ids.save)?.textContent === "Saved" &&
            !document.querySelector(
              '[data-grid-editing="true"], [data-grid-data-state="loading"], [data-grid-data-state="refreshing"], [data-grid-data-state="stale_error"], [data-grid-data-state="unavailable"]',
            ),
        });
      }
      return {
        ready: frames.every(
          (frame) =>
            frame.valid && JSON.stringify(frame) === JSON.stringify(frames[0]),
        ),
        frames,
      };
    },
    { ids, recordId },
  );
}

export async function verifyTimelineShellOrigin(page: Page, recordId: string) {
  let observation:
    | Awaited<ReturnType<typeof observeTimelineShellOrigin>>
    | undefined;
  try {
    await expect
      .poll(
        async () => {
          observation = await observeTimelineShellOrigin(page, recordId);
          return observation.ready;
        },
        {
          message:
            "Declared Timeline data, fonts, landmarks, Date Entered focus and top-left origin must remain correct for three frames",
        },
      )
      .toBe(true);
  } catch (cause) {
    throw new Error(
      `Timeline capture preparation failed: ${JSON.stringify(observation)}`,
      { cause },
    );
  }
}
