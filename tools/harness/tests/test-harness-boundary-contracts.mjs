import { runContractSuite } from "./contract-suite-support.mjs";
runContractSuite("boundaries");

import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";
import test from "node:test";
import { validateSchemaSync } from "../contract/index.mjs";
import { validateInspectorFieldLayouts } from "../generated-artifacts/design-presentation/design-presentation.mjs";

test("inspector presentation v2 rejects invalid authored field layouts", () => {
  const root = path.resolve(import.meta.dirname, "../../..");
  const presentation = JSON.parse(readFileSync(path.join(root, "contracts/design/presentation.v2.json"), "utf8"));
  const layouts = presentation.inspector.field_layout_overrides;
  const viewSchemas = path.join(root, "contracts/view-schemas");
  validateSchemaSync("cartulary.design_presentation.v2", presentation);
  validateInspectorFieldLayouts(layouts, viewSchemas);
  for (const candidate of [
    [...layouts, { ...layouts[0], layout: "narrative" }],
    [{ ...layouts[0], view_schema_id: "unknown" }],
    [{ ...layouts[0], field_key: "timeline.nonexistent" }],
    [{ ...layouts[0], layout: "unknown" }],
  ]) assert.throws(() => validateInspectorFieldLayouts(candidate, viewSchemas));
  const unknownLayout = structuredClone(presentation);
  unknownLayout.inspector.field_layout_overrides[0].layout = "unknown";
  assert.throws(() => validateSchemaSync("cartulary.design_presentation.v2", unknownLayout));
});
