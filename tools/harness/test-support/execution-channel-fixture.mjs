// Authored producer-side protocol fixture, independent of graph aggregation.
import { readdirSync } from 'node:fs';
import path from 'node:path';
import { atomicLocalFile, readLocalFile } from '../runtime/secure-local-files.mjs';
export function openExecutionChannel(runtime, maximum = 4096) {
  const root=runtime.privatePath('execution-observations');
  atomicLocalFile(path.join(root,'limits.json'),JSON.stringify({maximum_execution_records:maximum,maximum_record_bytes:65536}));
  atomicLocalFile(path.join(root,'open'),'1\n');
  return {environment:{CARTULARY_HARNESS_EXECUTION_ENABLED:'1'},
    close(){atomicLocalFile(path.join(root,'open'),'0\n',{replace:true});},
    records(){return readdirSync(root).filter(n=>n.startsWith('slot-')).map(n=>JSON.parse(readLocalFile(path.join(root,n))));}};
}
