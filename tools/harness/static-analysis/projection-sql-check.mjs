import { readFileSync } from "node:fs";
import { createSourceAdmission } from "./source-admission.mjs";
import path from "node:path";
import { inspectGoSQL, inspectGoImports } from "./go-import-inspection.mjs";
import { analyzedSQLStatements } from "./sql-relations.mjs";
import { providerPackageDefaults, declarationBindings, providerEntries, infrastructureEntries, foreignPackages, standardPackages } from "./projection-sql-policy.mjs";

// Read every production file of each imported local package, regardless of the
// current build tags. Generated Go is executable source and follows the same rule.
export function collectProjectionSources(root, entries, allowedForeign = foreignPackages) {
  if (!entries?.length) throw new Error("empty projection SQL scope");
  const files = new Map(), imports = new Map(), consumers = new Map(), foreign = new Set();
  const admission = createSourceAdmission(root);
  const modulePath = admission.file("go.mod", "repository module identity").content.match(/^module\s+(\S+)/mu)?.[1];
  if (!modulePath) throw new Error("missing repository module identity");
  function readPackage(pkg) {
    if (imports.has(pkg)) return imports.get(pkg);
    admission.directory(pkg, "projection source root");
    const sources=[];
    for (const entry of admission.entries(pkg, "projection source")) {
      if (/\.(?:s|S|c|cc|cpp|syso)$/u.test(entry.name)) throw new Error(`unclassified native source ${pkg}/${entry.name}`);
      if (!entry.name.endsWith('.go') || entry.name.endsWith('_test.go')) continue;
      const relative=`${pkg}/${entry.name}`;
      const {content}=admission.file(relative, "projection source");
      sources.push({relative,content});
    }
    if (!sources.length) throw new Error(`empty production package ${pkg}`);
    const dependencies=new Set();
    for (const source of sources) files.set(source.relative,source);
    for (const values of inspectGoImports([sources]).values()) for (const imported of values) {
      if (imported.startsWith(`${modulePath}/`)) dependencies.add(imported.slice(modulePath.length+1));
      else if (imported===modulePath || imported.startsWith('.') || imported==='C') throw new Error(`unresolved local or linked dependency ${pkg}: ${imported}`);
      else if (imported.split('/')[0].includes('.')) {
        if (!allowedForeign.has(imported)) throw new Error(`unadmitted foreign dependency ${pkg}: ${imported}`);
        foreign.add(imported);
      } else if (!standardPackages.has(imported)) throw new Error(`unadmitted standard dependency ${pkg}: ${imported}`);
    }
    imports.set(pkg,dependencies);
    return dependencies;
  }
  const seen=new Set();
  function visit(pkg, provider) {
    const key=`${pkg}:${provider ?? ''}`;
    if (seen.has(key)) return;
    seen.add(key);
    if (!consumers.has(pkg)) consumers.set(pkg,new Set());
    if (provider) consumers.get(pkg).add(provider);
    for (const dependency of readPackage(pkg)) visit(dependency,provider);
  }
  for (const entry of entries) visit(entry.package,entry.provider);
  return {files:[...files.values()].sort((a,b)=>a.relative.localeCompare(b.relative)), consumers, foreign:[...foreign].sort()};
}

export function evaluateProjectionSQL(declarations, providers, owners, defaults = providerPackageDefaults, bindings = declarationBindings, complete = false, consumers = null) {
  const failures = [];
  const seenProviders = new Set();
  const seenSources = new Set();
  const usedBindings = new Set();
  const usedDefaults = new Set();
  const allTables = new Set(providers.flatMap(p=>p.projection_table_ids));
  const allAuthorities = new Set(providers.flatMap(p=>p.source_authority_modules));
  for (const [file, entries] of declarations) {
    const pkg = path.posix.dirname(file);
    for (const entry of entries) {
      let statements;
      try { statements = analyzedSQLStatements([entry]); }
      catch (error) { failures.push(`${file}:${entry.name}: ${error.message}`); continue; }
      if (statements.some(statement=>statement.controls?.length)) {failures.push(`${file}:${entry.name}: provider dependencies cannot change SQL session authority`);continue;}
      // SQL vocabulary and value-only fragments confer no relation authority.
      // Unresolved construction has already failed above, including at sinks.
      if (!statements.some(statement=>statement.accesses.length)) continue;
      const role = bindings[pkg]?.[entry.name] ?? defaults[pkg];
      if (bindings[pkg]?.[entry.name]) usedBindings.add(`${pkg}:${entry.name}`);
      else if (defaults[pkg]) usedDefaults.add(pkg);
      if (!role) { failures.push(`${file}:${entry.name}: unassigned SQL declaration`); continue; }
      const provider = providers.find(p=>p.provider_id===role);
      if (provider) seenProviders.add(role);
      else if (!["compiled_projection", "restore_incidents", "projection_count", "shared_read"].includes(role)) { failures.push(`${file}:${entry.name}: unknown role ${role}`); continue; }
      const recipients = [...(consumers?.get(pkg) ?? [])];
      if (role === "shared_read" && !recipients.length) { failures.push(`${file}:${entry.name}: shared read capability has no provider provenance`); continue; }
      for (const access of statements.flatMap(statement=>statement.accesses)) {
        const fail = (message)=>failures.push(`${file}:${entry.name}: ${message}`);
        if (access.dynamic || access.table === "<dynamic>") {
          fail("unresolved provider relation");
          continue;
        }
        const table = access.table;
        const tableOwners = owners.get(table) ?? [];
        if (tableOwners.length !== 1) { fail(`${table}: expected one schema owner`); continue; }
        const owner = tableOwners[0];
        for (const recipient of recipients) {
          const consumer=providers.find(p=>p.provider_id===recipient);
          if (!consumer) {fail(`unknown provider provenance ${recipient}`);continue;}
          if (access.operation !== "read") fail(`provider ${recipient}: reachable ${access.operation} capability`);
          else if (owner === "projections" ? !consumer.projection_table_ids.includes(table) : !consumer.source_authority_modules.includes(owner)) fail(`provider ${recipient}: ${table} outside consumer authority`);
          else if (owner !== "projections") {seenSources.add(recipient);seenProviders.add(recipient);}
        }
        if (role === "projection_count") {
          if (!allTables.has(table) || access.operation !== "read") fail(`${table}: outside projection count maintenance`);
          continue;
        }
        if (role === "restore_incidents") {
          if (table !== "incidents" || access.operation !== "read") fail(`${table}: outside restore incident enumeration`);
          continue;
        }
        if (role === "shared_read") continue;
        const tables = provider ? new Set(provider.projection_table_ids) : allTables;
        const authorities = provider ? new Set(provider.source_authority_modules) : allAuthorities;
        if (owner === "projections") {
          if (!tables.has(table)) fail(`${table}: undeclared projection table`);
        } else if (access.operation !== "read") fail(`${table}: ${access.operation} of authoritative source`);
        else if (!authorities.has(owner)) fail(`${table}: undeclared source authority ${owner}`);
      }
    }
  }
  for (const provider of providers) if (!seenProviders.has(provider.provider_id)) failures.push(`provider ${provider.provider_id}: no SQL declaration coverage`);
  if (consumers) for (const provider of providers) if (!seenSources.has(provider.provider_id)) failures.push(`provider ${provider.provider_id}: no qualified source SQL coverage`);
  if (complete) for (const [pkg, entries] of Object.entries(bindings)) {
    for (const name of Object.keys(entries)) if (!usedBindings.has(`${pkg}:${name}`)) failures.push(`${pkg}:${name}: stale SQL binding`);
  }
  if (complete) for (const pkg of Object.keys(defaults)) if (!usedDefaults.has(pkg)) failures.push(`${pkg}: stale SQL package capability`);
  return failures.sort();
}

export function inspectProjectionSQL(root) {
  const registry = JSON.parse(readFileSync(path.join(root,"contracts/projection-providers/index.json"),"utf8"));
  if (registry.authority !== "validation_only_code_backed_registry_authoritative") throw new Error("projection descriptor mirror authority mismatch");
  const providers = registry.providers.filter(p=>p.status==="active");
  const manifest = JSON.parse(readFileSync(path.join(root,"tools/schema_object_ownership_manifest.json"),"utf8"));
  const owners = new Map();
  for (const entry of manifest.entries) {
    if (entry.management_class !== "cartulary_authored" || !["table","view"].includes(entry.object_kind)) continue;
    if (!entry.qualified_name.startsWith("public.")) throw new Error("unqualified schema relation");
    const table = entry.qualified_name.slice(7);
    owners.set(table, [...(owners.get(table) ?? []),entry.source_owner]);
  }
  const active=new Set(providers.map(p=>p.provider_id));
  for (const provider of active) if (!providerEntries[provider]?.length) throw new Error(`provider ${provider}: missing implementation entry`);
  for (const provider of Object.keys(providerEntries)) if (!active.has(provider)) throw new Error(`stale provider entry ${provider}`);
  const entries=Object.entries(providerEntries).flatMap(([provider,packages])=>packages.map(pkg=>({package:pkg,provider})));
  entries.push(...infrastructureEntries.map(pkg=>({package:pkg})));
  const closure=collectProjectionSources(root,entries);
  return evaluateProjectionSQL(inspectGoSQL([closure.files]),providers,owners,providerPackageDefaults,declarationBindings,true,closure.consumers);
}
