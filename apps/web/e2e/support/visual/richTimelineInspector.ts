import {
  timelineInspectorTestId,
  workbookInspectorCloseButtonTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import { expect, type Page, type TestInfo } from "@playwright/test";
import { waitForLoadedVendoredFonts } from "../runtime/visualRenderer";
import { openTimelineInspector } from "../workbook/rowMutations";
import { settleVisualGeometry } from "./capture";

// Supporting observations only. The sparse D-VFIX-002 fixture remains independent.
export async function captureRichTimelineInspector(
  page: Page,
  info: TestInfo,
  recordId: string,
) {
  for (const viewport of [
    { width: 1440, height: 900 },
    { width: 768, height: 640 },
  ]) {
    await page.setViewportSize(viewport);
    await openTimelineInspector(page, recordId);
    const inspector = page.getByTestId(timelineInspectorTestId());
    for (const section of ["Details", "Relationships", "Evidence"] as const) {
      const chooser = inspector.getByRole("button", { name: /^Sections:/ });
      if (await chooser.isVisible()) await chooser.click();
      await inspector
        .getByRole("button", { name: section, exact: true })
        .click();
      if (section === "Details") {
        await expect(
          inspector.locator(
            '[data-inspector-saved-field="timeline.raw_activity_text"]',
          ),
        ).toContainText("collection=authentication_excerpt,process_excerpt");
        await expect(
          inspector.locator(
            '[data-inspector-saved-field="timeline.analyst_text"]',
          ),
        ).toContainText("Jordan Ellis");
      } else if (section === "Relationships") {
        const correction = inspector.getByText("Correction and resolution", {
          exact: true,
        });
        if (
          (await correction.count()) &&
          (await correction.locator("..").getAttribute("open")) !== null
        )
          await correction.click();
        await expect(inspector).toContainText("Archive node");
        await expect(inspector).toContainText(
          "unidentified collection workstation",
        );
        // Dismissal performed before navigation is durable History evidence;
        // the Relationships owner retains only dismissals observed in this session.
        await expect(
          inspector.getByRole("button", { name: /Dismissed mention/ }),
        ).toHaveCount(0);
      } else {
        await expect(inspector).toContainText("Attached evidence count: 2");
        await expect(inspector).toContainText(
          "Authentication acquisition excerpt",
        );
        await expect(inspector).toContainText("Process acquisition excerpt");
      }
      await page.evaluate(waitForLoadedVendoredFonts);
      await settleVisualGeometry(page);
      const geometry = await inspector.evaluate((element) => {
        const box = element.getBoundingClientRect();
        const body = element.querySelector<HTMLElement>(
          "[data-inspector-scroll-body]",
        );
        return {
          visible:
            box.width > 0 &&
            box.height > 0 &&
            box.left >= 0 &&
            box.top >= 0 &&
            box.right <= innerWidth &&
            box.bottom <= innerHeight,
          bodyFits: !!body && body.scrollWidth <= body.clientWidth + 1,
        };
      });
      expect(geometry).toEqual({ visible: true, bodyFits: true });
      await info.attach(
        `rich-inspector-${section.toLowerCase()}-${viewport.width}x${viewport.height}`,
        {
          body: await page.screenshot({ animations: "disabled" }),
          contentType: "image/png",
        },
      );
    }
    await page
      .getByTestId(workbookInspectorCloseButtonTestId(timelineViewSchemaId))
      .click();
  }
}
