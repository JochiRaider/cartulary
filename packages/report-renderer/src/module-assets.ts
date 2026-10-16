import { constants } from "node:fs";
import { lstat, open } from "node:fs/promises";
import path from "node:path";

// A closed package-module surface. Validate before touching the filesystem:
// documentation, dependency trees, maps and arbitrary descendants have no route.
export function rendererModulePath(url: string): string | null {
  const match =
    /^\/renderer\/((?:chunks\/mermaid\.(?:core|esm)\/)?[A-Za-z0-9_-][A-Za-z0-9_.-]*\.mjs)$/.exec(
      url,
    );
  return match?.[1] ?? null;
}

export async function readRendererModule(
  root: string,
  url: string,
): Promise<Buffer> {
  const relative = rendererModulePath(url);
  if (relative === null) throw new Error("module_not_admitted");
  const parts = relative.split("/");
  let current = root;
  for (const [index, part] of parts.entries()) {
    current = path.join(current, part);
    const entry = await lstat(current);
    if (
      entry.isSymbolicLink() ||
      (index < parts.length - 1 ? !entry.isDirectory() : !entry.isFile())
    )
      throw new Error("module_not_admitted");
  }
  const file = await open(current, constants.O_RDONLY | constants.O_NOFOLLOW);
  try {
    const info = await file.stat();
    if (!info.isFile() || info.size > 64 * 1024 * 1024)
      throw new Error("module_not_admitted");
    const bytes = await file.readFile();
    if (bytes.byteLength > 64 * 1024 * 1024)
      throw new Error("module_not_admitted");
    return bytes;
  } finally {
    await file.close();
  }
}
