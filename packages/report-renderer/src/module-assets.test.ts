import {
  mkdir,
  mkdtemp,
  readFile,
  rm,
  symlink,
  writeFile,
} from "node:fs/promises";
import { tmpdir } from "node:os";
import path from "node:path";

import { expect, it } from "vitest";

import { readRendererModule, rendererModulePath } from "./module-assets.ts";

it("renderer module admission excludes documentation traversal and symlink targets", async () => {
  // Exercise every prohibited root from the operational boundary contract;
  // none is opened, traversed or used as a specification input.
  const policy = JSON.parse(
    await readFile(
      new URL("../../../tools/executable_input_policy.json", import.meta.url),
      "utf8",
    ),
  ) as { restricted_roots: string[] };
  for (const input of [
    ...policy.restricted_roots.map((root) => `/renderer/${root}/index.mjs`),
    "/renderer/README.md",
    "/renderer/a.mjs.map",
    "/renderer/../outside.mjs",
    "/renderer/%2e%2e/outside.mjs",
    "/renderer/a.mjs?secret=x",
    "/renderer/chunks/mermaid.esm/../../../outside.mjs",
    "/renderer/node_modules/tool.mjs",
    "/renderer/chunks/other/module.mjs",
    "/renderer/a\\b.mjs",
    "/renderer//a.mjs",
  ]) {
    expect(rendererModulePath(input)).toBeNull();
    // An absent root is deliberate: rejected routes need no filesystem access.
    await expect(
      readRendererModule("/absent-renderer-root", input),
    ).rejects.toThrow("module_not_admitted");
  }
  const root = await mkdtemp(path.join(tmpdir(), "cartulary-render-module-"));
  try {
    await writeFile(path.join(root, "entry.mjs"), "export default 1;", {
      mode: 0o600,
    });
    expect(
      (await readRendererModule(root, "/renderer/entry.mjs")).toString(),
    ).toBe("export default 1;");
    await symlink("README.md", path.join(root, "redirect.mjs"));
    await expect(
      readRendererModule(root, "/renderer/redirect.mjs"),
    ).rejects.toThrow("module_not_admitted");
    await mkdir(path.join(root, "chunks"));
    await symlink(root, path.join(root, "chunks", "mermaid.esm"));
    await expect(
      readRendererModule(root, "/renderer/chunks/mermaid.esm/entry.mjs"),
    ).rejects.toThrow("module_not_admitted");
  } finally {
    await rm(root, { recursive: true, force: true });
  }
});
