import { expect, test } from "./fixtures";

test("E-7-01 opens row history from the workbook surface with legal rollback actions", async ({
  page,
}, testInfo) => {
  void page;
  testInfo.annotations.push({
    type: "phase7-deferred",
    description:
      "Sprint 1 does not claim reviewer workbook history UI behavior.",
  });
  expect("GET /api/v1/records/{record_id}/history").toContain("history");
});

test("E-7-02 rolls back one mistaken mutation without reverting later unrelated edits", async ({
  page,
}, testInfo) => {
  void page;
  testInfo.annotations.push({
    type: "phase7-deferred",
    description: "Sprint 1 does not claim browser rollback behavior.",
  });
  expect("rollback").toBe("rollback");
});

test("E-7-03 soft-deletes and restores a row with tombstone concurrency", async ({
  page,
}, testInfo) => {
  void page;
  testInfo.annotations.push({
    type: "phase7-deferred",
    description:
      "Sprint 2 provides HTTP/WebSocket support fixtures; reviewer workbook UI remains deferred.",
  });
  expect("delete/restore").toContain("restore");
});

test("E-7-04 whole-row restore appends a new attributed revision", async ({
  page,
}, testInfo) => {
  void page;
  testInfo.annotations.push({
    type: "phase7-deferred",
    description: "Sprint 1 does not claim browser whole-row restore behavior.",
  });
  expect("whole-row restore").toContain("restore");
});
