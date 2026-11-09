import { readFileSync } from 'node:fs';
import { createHash } from 'node:crypto';
import { decodeGoPackages, validateApplicationClosure } from './closure.mjs';
const packages = decodeGoPackages(readFileSync(process.argv[2], 'utf8'));
validateApplicationClosure(packages);
const production = packages.filter(pkg => pkg.Name === 'main').map(pkg => pkg.ImportPath).sort();
if (production.length === 0) throw new Error('application executables missing from dependency closure');
const closure = [...new Set(packages.map(pkg => pkg.ImportPath))].sort();
process.stdout.write(`${JSON.stringify({ schema_id: 'cartulary.reference_pack_runtime_boundary.v1', status: 'pass', production_roots: production, packages: closure.length, dependency_closure_sha256: createHash('sha256').update(JSON.stringify(closure)).digest('hex'), dependency_resolution: 'readonly', workspace: 'off', network_resolution: 'off' })}\n`);
