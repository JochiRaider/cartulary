import { createHash } from "node:crypto";
import { RenderFailure } from "./failure.ts";

export interface StageAsset {
  path: string;
  media_type: string;
  bytes: string; // Canonical padded base64 of exact file octets.
  sha256: string;
}

export function stageAssetURL(path: string): string {
  return path
    .split("/")
    .map((part) =>
      encodeURIComponent(part).replace(
        /[!'()*]/gu,
        (char) => `%${char.charCodeAt(0).toString(16).toUpperCase()}`,
      ),
    )
    .join("/");
}

export function stageAssetsDigest(assets: readonly StageAsset[]): string {
  const hash = createHash("sha256").update(
    "cartulary.reporting_render_assets.v2\0",
  );
  for (const asset of assets) {
    const bytes = Buffer.from(asset.bytes, "base64");
    hash.update(
      `${Buffer.byteLength(asset.path)}:${asset.path}${Buffer.byteLength(asset.media_type)}:${asset.media_type}${bytes.byteLength}:`,
    );
    hash.update(bytes);
  }
  return hash.digest("hex");
}

function localAssetPath(value: string, media: string): boolean {
  if (
    Buffer.from(value).toString("utf8") !== value ||
    value.normalize("NFC") !== value ||
    Buffer.byteLength(value) > 512 ||
    /[\\:%?#]/u.test(value) ||
    [...value].some(
      (char) => (char.codePointAt(0) ?? 0) < 32 || char.codePointAt(0) === 127,
    )
  )
    return false;
  const parts = value.split("/");
  if (
    parts.length !== 3 ||
    parts[0] !== "assets" ||
    parts.some((part) => part === "" || part === "." || part === "..")
  )
    return false;
  const id = parts[1] ?? "";
  const filename = parts[2] ?? "";
  if (
    !(
      (media === "image/png" && filename.endsWith(".png")) ||
      (media === "image/jpeg" &&
        (filename.endsWith(".jpg") || filename.endsWith(".jpeg"))) ||
      (media === "image/svg+xml" && filename.endsWith(".svg"))
    )
  )
    return false;
  return (
    id.trim() === id && [...id].length <= 128 && !/[#\u007f-\u009f]/u.test(id)
  );
}

function themeAssetPath(value: string, media: string): boolean {
  if (media === "text/css; charset=utf-8") return value === "theme/style.css";
  if (media === "font/woff2")
    return /^theme\/fonts\/[a-z][a-z0-9_-]{0,63}\.woff2$/.test(value);
  if (media === "image/png")
    return /^theme\/images\/[a-z][a-z0-9_-]{0,63}\.png$/.test(value);
  return (
    media === "image/svg+xml" &&
    /^theme\/images\/[a-z][a-z0-9_-]{0,63}\.svg$/.test(value)
  );
}

// This validates the private transport, not the public image safety contract.
// Reporting admits passive SVG and raster dimensions before entering the stage;
// the stage cannot widen paths, content hashes or native sandbox capabilities.
export function validateStageAssets(
  value: unknown,
): asserts value is StageAsset[] {
  const invalid = (): never => {
    throw new RenderFailure("slidev_source_invalid", "stage_request_invalid");
  };
  if (!Array.isArray(value) || value.length > 600) return invalid();
  let previous = "";
  let localCount = 0,
    localBytes = 0,
    diagramCount = 0,
    diagramBytes = 0;
  let themeCount = 0,
    themeCSS = 0,
    themeFonts = 0,
    themeImages = 0;
  for (const entry of value) {
    if (!entry || typeof entry !== "object" || Array.isArray(entry))
      return invalid();
    const asset = entry as Record<string, unknown>;
    if (
      Object.keys(asset).length !== 4 ||
      typeof asset.path !== "string" ||
      typeof asset.media_type !== "string" ||
      typeof asset.bytes !== "string" ||
      typeof asset.sha256 !== "string" ||
      !/^[a-f0-9]{64}$/.test(asset.sha256) ||
      Buffer.compare(Buffer.from(asset.path), Buffer.from(previous)) <= 0 ||
      asset.bytes.length > Math.ceil((50 * 1024 * 1024) / 3) * 4
    )
      return invalid();
    const bytes = Buffer.from(asset.bytes, "base64");
    if (
      bytes.length === 0 ||
      bytes.toString("base64") !== asset.bytes ||
      createHash("sha256").update(bytes).digest("hex") !== asset.sha256
    )
      return invalid();
    if (
      /^diagrams\/[A-Za-z0-9][A-Za-z0-9_.-]{0,127}\.svg$/.test(asset.path) &&
      asset.media_type === "image/svg+xml"
    ) {
      diagramCount++;
      diagramBytes += bytes.length;
      if (
        diagramCount > 100 ||
        bytes.length > 8 * 1024 * 1024 ||
        diagramBytes > 32 * 1024 * 1024
      )
        return invalid();
    } else if (
      (localAssetPath(asset.path, asset.media_type) &&
        ["image/png", "image/jpeg", "image/svg+xml"].includes(
          asset.media_type,
        )) ||
      themeAssetPath(asset.path, asset.media_type)
    ) {
      localCount++;
      localBytes += bytes.length;
      if (
        localCount > 500 ||
        bytes.length > 50 * 1024 * 1024 ||
        localBytes > 250 * 1024 * 1024
      )
        return invalid();
      if (
        (asset.media_type === "font/woff2" &&
          bytes.length > 16 * 1024 * 1024) ||
        (asset.media_type === "text/css; charset=utf-8" && bytes.length > 8192)
      )
        return invalid();
    } else return invalid();
    previous = asset.path;
    if (asset.path.startsWith("theme/")) {
      themeCount++;
      if (asset.media_type === "text/css; charset=utf-8") themeCSS++;
      else if (asset.media_type === "font/woff2") themeFonts++;
      else themeImages++;
    }
  }
  if (themeCount && (themeCSS !== 1 || themeFonts > 2 || themeImages > 1))
    return invalid();
}
