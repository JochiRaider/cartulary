import { crc32, deflateRawSync } from "node:zlib";
import { cartularyImportDiscoveryPerformanceContract as contract } from "@cartulary/ui-contracts";
import type { ImportSourceFormat } from "./importProgressTimingSupport";

/** Fixed synthetic source workbook, constructed before the measurement begins. */
export function importDiscoveryWorkbook(format: ImportSourceFormat) {
  const rows: string[][] = [[...contract.workload.headers]];
  for (let index = 1; index <= contract.workload.data_rows; index++) {
    const ordinal = String(index).padStart(
      contract.workload.ordinal_width,
      "0",
    );
    rows.push([
      `${contract.workload.first_column_prefix}${ordinal}`,
      `${contract.workload.second_column_prefix}${ordinal}`,
    ]);
  }
  if (format === "csv")
    return {
      name: "discovery.csv",
      mimeType: "text/csv",
      buffer: Buffer.from(
        `${rows.map((row) => row.join(",")).join("\n")}\n`,
        "utf8",
      ),
    };
  const sheetRows = rows
    .map(
      (row, index) =>
        `<row r="${index + 1}">${row.map((value, column) => `<c r="${column === 0 ? "A" : "B"}${index + 1}" t="inlineStr"><is><t>${value}</t></is></c>`).join("")}</row>`,
    )
    .join("");
  const entries = [
    [
      "[Content_Types].xml",
      '<?xml version="1.0" encoding="UTF-8"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/></Types>',
    ],
    [
      "_rels/.rels",
      '<?xml version="1.0" encoding="UTF-8"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>',
    ],
    [
      "xl/workbook.xml",
      `<?xml version="1.0" encoding="UTF-8"?><workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="${contract.workload.worksheet}" sheetId="1" r:id="rId1"/></sheets></workbook>`,
    ],
    [
      "xl/_rels/workbook.xml.rels",
      '<?xml version="1.0" encoding="UTF-8"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/></Relationships>',
    ],
    [
      "xl/worksheets/sheet1.xml",
      `<?xml version="1.0" encoding="UTF-8"?><worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><dimension ref="A1:B${rows.length}"/><sheetData>${sheetRows}</sheetData></worksheet>`,
    ],
  ];
  // Minimal deterministic ZIP: fixed names, timestamps, CRC and raw DEFLATE.
  // It only emits these five trusted parts; it is not an archive reader.
  const local: Buffer[] = [],
    central: Buffer[] = [];
  let offset = 0;
  for (const [name, contents] of entries) {
    if (name === undefined || contents === undefined)
      throw new Error("Workbook part is missing");
    const filename = Buffer.from(name),
      data = Buffer.from(contents),
      compressed = deflateRawSync(data);
    const header = Buffer.alloc(30),
      directory = Buffer.alloc(46);
    header.writeUInt32LE(0x04034b50, 0);
    header.writeUInt16LE(20, 4);
    header.writeUInt16LE(8, 8);
    header.writeUInt16LE(0x21, 12);
    header.writeUInt32LE(crc32(data), 14);
    header.writeUInt32LE(compressed.length, 18);
    header.writeUInt32LE(data.length, 22);
    header.writeUInt16LE(filename.length, 26);
    directory.writeUInt32LE(0x02014b50, 0);
    directory.writeUInt16LE(20, 4);
    header.copy(directory, 6, 4, 28);
    directory.writeUInt32LE(offset, 42);
    local.push(header, filename, compressed);
    central.push(directory, filename);
    offset += header.length + filename.length + compressed.length;
  }
  const directory = Buffer.concat(central),
    end = Buffer.alloc(22);
  end.writeUInt32LE(0x06054b50, 0);
  end.writeUInt16LE(entries.length, 8);
  end.writeUInt16LE(entries.length, 10);
  end.writeUInt32LE(directory.length, 12);
  end.writeUInt32LE(offset, 16);
  return {
    name: "discovery.xlsx",
    mimeType:
      "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
    buffer: Buffer.concat([...local, directory, end]),
  };
}
