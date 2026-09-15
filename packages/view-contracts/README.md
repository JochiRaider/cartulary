# View contracts

The public facade projects authored workbook view, surface and reference contracts.
Adopted Core owners define behavior; this package does not introduce HTTP routes.

`src/references.ts` exposes the 27 ordinary existing-record field bindings from
`contracts/view-references/index.json`, through the generated protocol registry.
Target record kinds and declared candidate surfaces remain separate. Member IDs,
Party IDs and general record IDs have distinct identities. Collection operation
tokens are projected; removal still uses the returned collection item_ref.
Clearability and field editing capabilities remain on the existing view field.

Edit authored contract inputs and use `make generate`. Never hand-edit generated
registries. Choose verification with
`make task-guide ROLE=module-author OWNER=package.view_contracts`.
