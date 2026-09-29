import { createHash } from "node:crypto";
import { readLocalChunks, readLocalFile } from "../../runtime/secure-local-files.mjs";
import { ReviewFailure } from "./failure.mjs";

/** Immutable borrowed bytes, revalidated on every read/copy. Publication streams
 * these descriptors; raster engines alone materialize the selected image.
 */
export class FilePayload {
  #file;
  constructor(file, ref) { this.#file = file; this.length = ref.bytes; this.sha256 = ref.sha256; Object.freeze(this); }
  *[Symbol.iterator]() {
    const hash = createHash("sha256"); let length = 0;
    for (const chunk of readLocalChunks(this.#file, { maximum: this.length })) { hash.update(chunk); length += chunk.length; yield chunk; }
    if (length !== this.length || hash.digest("hex") !== this.sha256) throw new ReviewFailure("invalid_artifact");
  }
  read() {
    const bytes = readLocalFile(this.#file, { maximum: this.length });
    if (bytes.length !== this.length || createHash("sha256").update(bytes).digest("hex") !== this.sha256) throw new ReviewFailure("invalid_artifact");
    return bytes;
  }
}
export const bytesOf = (value) => value instanceof FilePayload ? value.read() : value;
export class StoredFiles extends Map {
  get(name) { return bytesOf(super.get(name)); }
  source(name) { return super.get(name); }
}
