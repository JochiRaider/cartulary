import { describe, expect, it } from "vitest";

import {
  assertGroupRowPresentationOnly,
  collapseGridGroup,
  expandGridGroup,
} from "../grid";
import { testTimelineViewSchemaId } from "./browser-fixtures";

export function registerGroupingSuite() {
  describe("@cartulary/test-utils selector choreography", () => {
    it("toggles group outline expansion by aria state", async () => {
      let ariaExpanded = "true";
      let clickCount = 0;
      const element = {
        getAttribute(name: string) {
          return name === "aria-expanded" ? ariaExpanded : null;
        },
      } as Element;
      const page = {
        getByTestId(value: string) {
          expect(value).toBe("group-row");
          return {
            click: async () => {
              clickCount += 1;
              ariaExpanded = ariaExpanded === "true" ? "false" : "true";
            },
            evaluate: async (
              pageFunction: (element: Element, arg?: unknown) => unknown,
              arg?: unknown,
            ) => pageFunction(element, arg),
            fill: async () => undefined,
          };
        },
      };

      await collapseGridGroup({
        groupTestId: "group-row",
        page,
        surface: testTimelineViewSchemaId,
      });
      expect(ariaExpanded).toBe("false");
      expect(clickCount).toBe(1);

      await collapseGridGroup({
        groupTestId: "group-row",
        page,
        surface: testTimelineViewSchemaId,
      });
      expect(clickCount).toBe(1);

      await expandGridGroup({
        groupTestId: "group-row",
        page,
        surface: testTimelineViewSchemaId,
      });
      expect(ariaExpanded).toBe("true");
      expect(clickCount).toBe(2);
    });

    it("asserts group rows remain presentation-only", async () => {
      const page = {
        getByTestId(value: string) {
          const element = Array.from(
            document.querySelectorAll<HTMLElement>("[data-testid]"),
          ).find((candidate) => candidate.dataset.testid === value);
          if (element === undefined) {
            throw new Error(`Missing test id ${value}`);
          }
          return {
            click: async () => undefined,
            evaluate: async (
              pageFunction: (element: Element, arg?: unknown) => unknown,
              arg?: unknown,
            ) => pageFunction(element, arg),
            fill: async () => undefined,
          };
        },
      };

      document.body.innerHTML = `
        <div role="row" aria-level="1" aria-expanded="true">
          <div role="gridcell">
            <button aria-expanded="true" data-testid="group-row" type="button">reviewed</button>
          </div>
        </div>
      `;
      await expect(
        assertGroupRowPresentationOnly({
          groupTestId: "group-row",
          page,
          surface: testTimelineViewSchemaId,
        }),
      ).resolves.toBeUndefined();

      document.body.innerHTML = `
        <div role="row" aria-level="1" aria-expanded="true" data-grid-record-id="record-1">
          <div role="gridcell">
            <button aria-expanded="true" data-testid="group-row" type="button">reviewed</button>
          </div>
        </div>
      `;
      await expect(
        assertGroupRowPresentationOnly({
          groupTestId: "group-row",
          page,
          surface: testTimelineViewSchemaId,
        }),
      ).rejects.toThrow(/omit data-grid-record-id/);
    });
  });
}
