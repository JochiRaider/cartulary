export const entityMergeIdentifierFields = {
  host: [
    {
      key: "host.aad_device_id",
      label: "AAD Device ID",
      identifierClass: "aad_device_id",
    },
    { key: "host.fqdn", label: "FQDN", identifierClass: "fqdn" },
    { key: "host.hostname", label: "Hostname", identifierClass: "hostname" },
  ],
  identity: [
    {
      key: "identity.aad_object_id",
      label: "AAD Object ID",
      identifierClass: "aad_object_id",
    },
    { key: "identity.sid", label: "SID", identifierClass: "sid" },
    { key: "identity.upn", label: "UPN", identifierClass: "upn" },
    { key: "identity.email", label: "Email", identifierClass: "email" },
    {
      key: "identity.sam_account_name",
      label: "SAM Account Name",
      identifierClass: "sam_account_name",
    },
  ],
} as const;
