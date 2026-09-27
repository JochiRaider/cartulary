import sharp from "sharp";
import { ReviewFailure, limits } from "./contract.mjs";

const signature = Buffer.from([137, 80, 78, 71, 13, 10, 26, 10]);
const crcTable = Array.from({ length: 256 }, (_, initial) => {
  let value = initial;
  for (let bit = 0; bit < 8; bit++) value = value & 1 ? 0xedb88320 ^ (value >>> 1) : value >>> 1;
  return value >>> 0;
});
export function crc32(bytes) {
  let value = 0xffffffff;
  for (const byte of bytes) value = crcTable[(value ^ byte) & 255] ^ (value >>> 8);
  return (value ^ 0xffffffff) >>> 0;
}
export function pngHeader(bytes) {
  const invalid = () => { throw new ReviewFailure("invalid_artifact"); };
  if (!Buffer.isBuffer(bytes) || bytes.length > limits.png || bytes.length < 45 || !bytes.subarray(0, 8).equals(signature)) invalid();
  let offset = 8, header, data = false, ended = false, dataEnded = false;
  const unique = new Set();
  while (offset < bytes.length) {
    if (offset + 12 > bytes.length) invalid();
    const length = bytes.readUInt32BE(offset), end = offset + length + 12;
    if (end > bytes.length) invalid();
    const kind = bytes.toString("ascii", offset + 4, offset + 8), payload = bytes.subarray(offset + 8, end - 4);
    if (!/^[A-Za-z]{4}$/u.test(kind) || crc32(bytes.subarray(offset + 4, end - 4)) !== bytes.readUInt32BE(end - 4)) invalid();
    if (!header && kind !== "IHDR") invalid();
    if (["IHDR", "IEND", "sRGB", "gAMA", "cHRM", "PLTE"].includes(kind)) { if (unique.has(kind)) invalid(); unique.add(kind); }
    if (kind === "IHDR") {
      if (length !== 13 || payload[8] !== 8 || ![2, 6].includes(payload[9]) || payload[10] !== 0 || payload[11] !== 0 || payload[12] > 1) invalid();
      header = { width: payload.readUInt32BE(0), height: payload.readUInt32BE(4) };
      if (!header.width || !header.height || header.width > limits.dimension || header.height > limits.dimension || header.width * header.height > limits.pixels) invalid();
    } else if (kind === "IDAT") { if (dataEnded) invalid(); data = true; }
    else if (data) dataEnded = true;
    if (["acTL", "fcTL", "fdAT", "iCCP", "cICP", "mDCv", "cLLi"].includes(kind)) invalid();
    if (kind === "sRGB" && (length !== 1 || payload[0] > 3)) invalid();
    if (kind === "gAMA" && (length !== 4 || payload.readUInt32BE() !== 45455)) invalid();
    if (kind === "cHRM" && (length !== 32 || ![31270, 32900, 64000, 33000, 30000, 60000, 15000, 6000].every((value, index) => payload.readUInt32BE(index * 4) === value))) invalid();
    if (kind[0] === kind[0].toUpperCase() && !["IHDR", "IDAT", "IEND", "PLTE"].includes(kind)) invalid();
    if (kind === "IEND") { if (length || !data || end !== bytes.length) invalid(); ended = true; }
    offset = end;
  }
  if (!ended) invalid();
  return header;
}
export async function decodePNG(bytes) {
  const header = pngHeader(bytes);
  try {
    const { data, info } = await sharp(bytes, { limitInputPixels: limits.pixels, failOn: "warning", sequentialRead: true }).ensureAlpha().raw().toBuffer({ resolveWithObject: true });
    if (info.width !== header.width || info.height !== header.height || info.channels !== 4 || data.length !== header.width * header.height * 4) throw new Error("decoded shape");
    return { ...header, data };
  } catch (cause) { throw new ReviewFailure("invalid_artifact", { cause }); }
}
export async function encodePNG(image) {
  try {
    const bytes = await sharp(image.data, { raw: { width: image.width, height: image.height, channels: 4 } }).png({ palette: false, compressionLevel: 9 }).toBuffer();
    pngHeader(bytes); return bytes;
  } catch (cause) { throw new ReviewFailure("invalid_artifact", { cause }); }
}
