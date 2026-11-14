package deleterestore

import shape "github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract"

func ValidateHostSnapshotSource(record, source map[string]any) error {
	return shape.ValidateSourceObject(record, source,
		shape.SourceFieldGroup{Names: "record_id incident_id created_by_user_id updated_by_user_id", Kind: shape.SourceUUID, Nullable: false},
		shape.SourceFieldGroup{Names: "display_name entity_origin host_state", Kind: shape.SourceString, Nullable: false},
		shape.SourceFieldGroup{Names: "hostname aad_device_id fqdn location os_platform business_owner criticality containment_status", Kind: shape.SourceString, Nullable: true},
		shape.SourceFieldGroup{Names: "seed_entity_mention_id merged_into_record_id", Kind: shape.SourceUUID, Nullable: true},
		shape.SourceFieldGroup{Names: "row_version", Kind: shape.SourceInteger, Nullable: false},
		shape.SourceFieldGroup{Names: "created_at updated_at", Kind: shape.SourceTimestamp, Nullable: false},
	)
}

func ValidateIdentitySnapshotSource(record, source map[string]any) error {
	return shape.ValidateSourceObject(record, source,
		shape.SourceFieldGroup{Names: "record_id incident_id created_by_user_id updated_by_user_id", Kind: shape.SourceUUID, Nullable: false},
		shape.SourceFieldGroup{Names: "display_name entity_origin identity_state", Kind: shape.SourceString, Nullable: false},
		shape.SourceFieldGroup{Names: "upn email sam_account_name aad_object_id sid privilege_level mfa_state reset_status", Kind: shape.SourceString, Nullable: true},
		shape.SourceFieldGroup{Names: "seed_entity_mention_id merged_into_record_id", Kind: shape.SourceUUID, Nullable: true},
		shape.SourceFieldGroup{Names: "row_version", Kind: shape.SourceInteger, Nullable: false},
		shape.SourceFieldGroup{Names: "created_at updated_at", Kind: shape.SourceTimestamp, Nullable: false},
	)
}
