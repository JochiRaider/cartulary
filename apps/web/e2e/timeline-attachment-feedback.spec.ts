import { Buffer } from "node:buffer";
import { scrollGridTargetIntoView } from "@cartulary/test-utils/grid";
import { gridShellTestId, rowCellTestId } from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import type { Locator, Page, TestInfo } from "@playwright/test";
import { expect, test } from "./fixtures";
import { installVisualPreferences } from "./support/auth/visualPreferences";
import { openTimelineAttachmentFeedback } from "./support/evidence/uploads";
import { createIncident } from "./support/incidents/fixtures";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { expectDecisionControlReachable } from "./support/workbook/decisionSupersession";
import { createViewRow } from "./support/workbook/query";

type Observation = {
  presentations: number;
  gridRenders: number;
  announcements: string[];
};

async function observe(page: Page) {
  await page.addInitScript(() => {
    type Fiber = {
      flags: number;
      memoizedProps?: {
        model?: { grid?: { timelineGridRows?: unknown } };
        timelineGridRows?: unknown;
      };
      child?: Fiber;
      sibling?: Fiber;
    };
    const observed = window as unknown as {
      attachmentObservation: Observation;
      __REACT_DEVTOOLS_GLOBAL_HOOK__: unknown;
    };
    observed.attachmentObservation = {
      presentations: 0,
      gridRenders: 0,
      announcements: [],
    };
    let previous = new WeakSet<object>();
    observed.__REACT_DEVTOOLS_GLOBAL_HOOK__ = {
      supportsFiber: true,
      inject: () => 1,
      onCommitFiberUnmount: () => {},
      onCommitFiberRoot: (_renderer: number, root: { current: Fiber }) => {
        const current = new WeakSet<object>();
        const visit = (fiber?: Fiber) => {
          if (!fiber) return;
          current.add(fiber);
          if (!previous.has(fiber) && (fiber.flags & 1) !== 0) {
            if (fiber.memoizedProps?.model?.grid?.timelineGridRows)
              observed.attachmentObservation.presentations++;
            if (fiber.memoizedProps?.timelineGridRows)
              observed.attachmentObservation.gridRenders++;
          }
          visit(fiber.child);
          visit(fiber.sibling);
        };
        visit(root.current);
        previous = current;
      },
    };
    document.addEventListener("DOMContentLoaded", () => {
      new MutationObserver((changes) => {
        const regions = new Set<Element>();
        for (const change of changes) {
          const target =
            change.target instanceof Element
              ? change.target
              : change.target.parentElement;
          const region = target?.closest(
            '[role="alert"], [role="status"], [aria-live]',
          );
          if (region && region.getAttribute("aria-live") !== "off")
            regions.add(region);
        }
        for (const region of regions) {
          const message = region.textContent?.trim();
          if (message)
            observed.attachmentObservation.announcements.push(message);
        }
      }).observe(document.body, {
        subtree: true,
        childList: true,
        characterData: true,
      });
    });
  });
}

async function gesture(page: Page, recordId: string, name: string, count = 1) {
  await scrollGridTargetIntoView({
    page,
    surface: timelineViewSchemaId,
    targetTestId: rowCellTestId(recordId, "timeline.activity_synopsis_text"),
  });
  await page
    .getByTestId(rowCellTestId(recordId, "timeline.activity_synopsis_text"))
    .evaluate(
      (element, input) => {
        const transfer = new DataTransfer();
        for (let i = 0; i < input.count; i++)
          transfer.items.add(
            new File(["attachment"], `${input.name}-${i}.txt`, {
              type: "text/plain",
            }),
          );
        element.dispatchEvent(
          new DragEvent("drop", {
            bubbles: true,
            cancelable: true,
            dataTransfer: transfer,
          }),
        );
      },
      { name, count },
    );
}

async function reachable(page: Page, control: Locator) {
  await expectDecisionControlReachable(page, control);
  // Include overflow clipping, not only bounding boxes inside the browser viewport.
  await expect(control).toBeInViewport({ ratio: 1 });
}

async function capture(page: Page, info: TestInfo, label: string) {
  const grid = await page
    .getByTestId(gridShellTestId(timelineViewSchemaId))
    .boundingBox();
  const region = page.getByRole("region", {
    name: "Timeline file work area",
    exact: true,
  });
  const workArea = await region.boundingBox();
  const observations = await page.evaluate(() => {
    const active = document.activeElement;
    return {
      ...(window as unknown as { attachmentObservation: Observation })
        .attachmentObservation,
      focus: {
        role: active?.getAttribute("role"),
        label: active?.getAttribute("aria-label"),
        tag: active?.tagName,
      },
    };
  });
  await info.attach(label, {
    body: Buffer.from(
      JSON.stringify(
        {
          grid,
          workArea,
          feedbackHeight: grid && workArea ? grid.y - workArea.y : null,
          observations,
          controls: await region.getByRole("button").allTextContents(),
        },
        null,
        2,
      ),
    ),
    contentType: "application/json",
  });
  await info.attach(`${label}-browser`, {
    body: await page.screenshot(),
    contentType: "image/png",
  });
  return { grid, workArea, observations };
}

test("Timeline attachment feedback keeps completed history bounded and presentation independent", async ({
  page,
  workerAdmin,
}, info) => {
  await page.setViewportSize({ width: 1280, height: 720 });
  await observe(page);
  const incident = await createIncident(
    page,
    uniqueIncidentKey("TAF-CLEANUP"),
    "Attachment feedback",
  );
  const rows = [];
  for (let i = 0; i < 4; i++)
    rows.push(
      await createViewRow(page, incident, timelineViewSchemaId, {
        client_txn_id: uniqueTxn("attachment-source"),
        "timeline.activity_synopsis_text": `Attachment source ${i}`,
      }),
    );
  const preferences = await installVisualPreferences(page, workerAdmin.user_id);
  const writes: string[] = [];
  page.on("request", (request) => {
    if (
      ["POST", "PUT", "PATCH", "DELETE"].includes(request.method()) &&
      !request.url().endsWith("/query")
    )
      writes.push(`${request.method()} ${request.url()}`);
  });
  for (const density of ["compact", "default", "comfortable"] as const) {
    preferences.select(density);
    await page.setViewportSize({ width: 1280, height: 720 });
    await page.goto(
      `/?incident_id=${incident}&view_schema_id=${timelineViewSchemaId}`,
    );
    const grid = page.getByTestId(gridShellTestId(timelineViewSchemaId));
    const feedback = page.getByRole("region", {
      name: "Timeline attachments",
      exact: true,
    });
    const trigger = page.getByRole("button", { name: /^Attachments:/ });
    await expect(grid).toBeVisible();
    await expect(
      page.locator("[data-cartulary-density]").first(),
    ).toHaveAttribute("data-cartulary-density", density);
    let firstHeight = 0;
    for (const [index, row] of rows.slice(0, 3).entries()) {
      await gesture(page, row.record_id, `completed-${index}`);
      await expect(trigger).toHaveText(
        `Attachments: 0 need attention, 0 in progress, ${index + 1} completed`,
      );
      await expect(trigger).toHaveAttribute("aria-expanded", "false");
      await expect(feedback.getByRole("group")).toHaveCount(0);
      const result = await capture(
        page,
        info,
        `${density}-completed-${index + 1}`,
      );
      if (!result.grid) throw new Error("Missing grid geometry");
      if (index === 0) firstHeight = result.grid.height;
      else expect(result.grid.height).toBe(firstHeight);
    }
    const source = rows[3];
    if (!source) throw new Error("Missing upload source");
    let release = () => {};
    let transfers = 0;
    const held = new Promise<void>((resolve) => {
      release = resolve;
    });
    const path = "**/api/v1/object-uploads/*";
    await page.route(path, async (route) => {
      transfers++;
      await held;
      await route.continue();
    });
    let beforeAcceptance: Observation | null = null;
    try {
      await gesture(
        page,
        source.record_id,
        "pending-with-a-long-original-source-filename-for-narrow-layout-controls",
      );
      await expect.poll(() => transfers).toBe(1);
      await expect(trigger).toContainText("1 in progress");
      const pending = await capture(page, info, `${density}-upload-pending`);
      await gesture(page, source.record_id, "rejected", 2);
      await expect(
        feedback.getByText("Choose one file at a time.", { exact: true }),
      ).toBeVisible();
      const rejected = await capture(
        page,
        info,
        `${density}-admission-rejection`,
      );
      // The first admission notice changes work-area geometry. Its resize can
      // legitimately propagate; repeat the same event without geometry changes
      // to isolate status-only observation from accepted rows and layout.
      expect(rejected.grid?.height).toBeLessThan(pending.grid?.height ?? 0);
      const announcements = rejected.observations.announcements.filter(
        (value) => value === "Choose one file at a time.",
      ).length;
      await gesture(page, source.record_id, "rejected", 2);
      await expect
        .poll(async () =>
          page.evaluate(
            () =>
              (
                window as unknown as { attachmentObservation: Observation }
              ).attachmentObservation.announcements.filter(
                (value) => value === "Choose one file at a time.",
              ).length,
          ),
        )
        .toBe(announcements + 1);
      const repeatedRejection = await capture(
        page,
        info,
        `${density}-repeated-admission`,
      );
      expect(repeatedRejection.observations.presentations).toBe(
        rejected.observations.presentations,
      );
      expect(repeatedRejection.observations.gridRenders).toBe(
        rejected.observations.gridRenders,
      );
      const before = [...writes];
      await trigger.focus();
      await trigger.press("Enter");
      await expect(
        feedback.getByRole("group", { name: /^File recovery: pending/ }),
      ).toContainText("Uploading file. You can keep editing.");
      const history = feedback.getByText("Completed attachments (3)", {
        exact: true,
      });
      await expect(history.locator("..")).not.toHaveAttribute("open", "");
      await history.click();
      await expect(
        feedback.getByRole("group", {
          name: "File recovery: completed-0-0.txt",
          exact: true,
        }),
      ).toBeVisible();
      await expect(
        history
          .locator("..")
          .getByRole("button", { name: "Discard retained file work" }),
      ).toHaveCount(0);
      await history.click();
      for (const viewport of [
        { width: 1280, height: 720 },
        { width: 1024, height: 720 },
        { width: 768, height: 640 },
        { width: 390, height: 480 },
      ]) {
        await page.setViewportSize(viewport);
        await reachable(page, trigger);
        const workArea = await page
          .getByRole("region", { name: "Timeline file work area", exact: true })
          .boundingBox();
        const box = await feedback.boundingBox();
        if (!box || !workArea) throw new Error("Missing bounded feedback");
        expect(box.height).toBeLessThanOrEqual(workArea.height * 0.25 + 1);
        expect((await grid.boundingBox())?.height).toBeGreaterThan(0);
        const stop = feedback.getByRole("button", {
          name: "Discard retained file work",
          exact: true,
        });
        await reachable(page, stop);
        await capture(page, info, `${density}-details-${viewport.width}`);
      }
      await page.setViewportSize({ width: 1280, height: 720 });
      await page.evaluate(() => {
        document.documentElement.style.zoom = "200%";
      });
      await reachable(page, trigger);
      await reachable(
        page,
        feedback.getByRole("button", {
          name: "Discard retained file work",
          exact: true,
        }),
      );
      await capture(page, info, `${density}-zoom-200`);
      await page.evaluate(() => {
        document.documentElement.style.zoom = "";
      });
      const spacing = await page.addStyleTag({
        content:
          "#root * { letter-spacing: 0.12em !important; line-height: 1.5 !important; word-spacing: 0.16em !important; }",
      });
      await reachable(
        page,
        feedback.getByRole("button", {
          name: "Discard retained file work",
          exact: true,
        }),
      );
      await capture(page, info, `${density}-text-spacing`);
      await spacing.evaluate((element) =>
        element.parentNode?.removeChild(element),
      );
      await trigger.focus();
      await trigger.press("Escape");
      await expect(trigger).toHaveAttribute("aria-expanded", "false");
      await expect(trigger).toBeFocused();
      expect(writes).toEqual(before);
      beforeAcceptance = (
        await capture(page, info, `${density}-before-acceptance`)
      ).observations;
    } finally {
      release();
    }
    await expect(trigger).toHaveText(
      "Attachments: 0 need attention, 0 in progress, 4 completed",
    );
    await expect(trigger).toBeFocused();
    await page.unroute(path);
    const accepted = await capture(page, info, `${density}-all-completed`);
    if (!beforeAcceptance) throw new Error("Missing held-write observation");
    expect(accepted.observations.presentations).toBeGreaterThan(
      beforeAcceptance.presentations,
    );
    expect(accepted.observations.gridRenders).toBeGreaterThan(
      beforeAcceptance.gridRenders,
    );
    const row = rows[0];
    if (!row) throw new Error("Missing repeated source");
    // A new attachment on the same source has fresh identity and visible progress.
    let finish = () => {};
    const repeated = new Promise<void>((resolve) => {
      finish = resolve;
    });
    await page.route(path, async (route) => {
      await repeated;
      await route.continue();
    });
    try {
      await gesture(page, row.record_id, "replacement");
      await expect(trigger).toHaveText(
        "Attachments: 0 need attention, 1 in progress, 3 completed",
      );
      await expect(trigger).toHaveAttribute("aria-expanded", "false");
      await openTimelineAttachmentFeedback(page);
      await expect(
        feedback.getByRole("group", {
          name: "File recovery: replacement-0.txt",
          exact: true,
        }),
      ).toBeVisible();
    } finally {
      finish();
    }
    await expect(trigger).toHaveText(
      "Attachments: 0 need attention, 0 in progress, 4 completed",
    );
    await page.unroute(path);
  }
});
