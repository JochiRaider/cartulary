import { test } from "@playwright/test";

test.skip("E-4-01 resolves and creates entities from Timeline mentions in the inspector", async ({
  page,
}) => {
  void page;
  test.info().annotations.push({
    type: "todo",
    description:
      "TODO Phase 4 E-4-01: enable when the browser harness exposes Timeline relationship cells plus inspector resolve/create-from-mention controls.",
  });
});

test.skip("E-4-02 dismisses and ordinarily restores a mention without relinking", async ({
  page,
}) => {
  void page;
  test.info().annotations.push({
    type: "todo",
    description:
      "TODO Phase 4 E-4-02: enable when the browser harness exposes dismiss and ordinary restore controls for mentions.",
  });
});

test.skip("E-4-03 merges duplicate entities from the inspector and preserves survivor identity", async ({
  page,
}) => {
  void page;
  test.info().annotations.push({
    type: "todo",
    description:
      "TODO Phase 4 E-4-03: enable when the browser harness exposes entity inspector merge flow.",
  });
});

test.skip("E-4-04 auto-resolves only eligible exact-match Timeline tokens", async ({
  page,
}) => {
  void page;
  test.info().annotations.push({
    type: "todo",
    description:
      "TODO Phase 4 E-4-04: enable when the browser harness exposes auto-resolution disclosure and unresolved mention UI.",
  });
});
