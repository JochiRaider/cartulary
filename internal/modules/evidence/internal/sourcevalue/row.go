// Package sourcevalue owns the complete Evidence source row shared by retained
// snapshots and portable current state. It does not admit database coercion.
package sourcevalue

import (
	"encoding/json"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/modules/evidence/internal/policy"
	"github.com/JochiRaider/cartulary/internal/modules/evidence/internal/textpolicy"
	shape "github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/google/uuid"
)

type Row struct {
	RecordID           uuid.UUID  `json:"record_id"`
	IncidentID         uuid.UUID  `json:"incident_id"`
	Title              *string    `json:"title"`
	LifecycleState     string     `json:"lifecycle_state"`
	RequestedAt        *time.Time `json:"requested_at"`
	ReceivedAt         *time.Time `json:"received_at"`
	StorageRef         *string    `json:"storage_ref"`
	BlobHash           *string    `json:"blob_hash"`
	CollectorPartyText *string    `json:"collector_party_text"`
	CollectorPartyID   *uuid.UUID `json:"collector_party_id"`
	SourcePartyText    *string    `json:"source_party_text"`
	SourcePartyID      *uuid.UUID `json:"source_party_id"`
	UploadState        string     `json:"upload_state"`
	ObjectBlobID       *uuid.UUID `json:"object_blob_id"`
	OwnerUserID        *uuid.UUID `json:"owner_user_id"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// Parse admits every member before decoding so omitted values cannot become null
// and strings cannot be silently interpreted as a different scalar type.
func Parse(raw map[string]any) (Row, bool) {
	if shape.ValidateSourceObject(raw, raw,
		shape.SourceFieldGroup{Names: "record_id incident_id", Kind: shape.SourceUUID},
		shape.SourceFieldGroup{Names: "lifecycle_state upload_state", Kind: shape.SourceString},
		shape.SourceFieldGroup{Names: "title storage_ref blob_hash collector_party_text source_party_text", Kind: shape.SourceString, Nullable: true},
		shape.SourceFieldGroup{Names: "requested_at received_at", Kind: shape.SourceTimestamp, Nullable: true},
		shape.SourceFieldGroup{Names: "collector_party_id source_party_id object_blob_id owner_user_id", Kind: shape.SourceUUID, Nullable: true},
		shape.SourceFieldGroup{Names: "created_at updated_at", Kind: shape.SourceTimestamp},
	) != nil {
		return Row{}, false
	}
	for _, value := range raw {
		if text, ok := value.(string); ok && (!utf8.ValidString(text) || strings.ContainsRune(text, '\x00')) {
			return Row{}, false
		}
	}
	for _, name := range []string{"title", "storage_ref", "collector_party_text", "source_party_text"} {
		field, ok := viewschema.LookupField("cartulary.view.evidence.v1", "evidence."+name)
		if !ok || field.StringContractID == nil {
			return Row{}, false
		}
		if raw[name] == nil && field.Clearable {
			continue
		}
		text, ok := raw[name].(string)
		if !ok {
			return Row{}, false
		}
		normalized, valid := textpolicy.Normalize(text, *field.StringContractID)
		if !valid || normalized == "" || normalized != text {
			return Row{}, false
		}
	}
	payload, err := json.Marshal(raw)
	var row Row
	if err != nil || json.Unmarshal(payload, &row) != nil {
		return Row{}, false
	}
	row.UTCClocks()
	return row, true
}

func (r Row) LifecycleLegal() bool {
	return policy.ValidEvidenceLifecycle(r.LifecycleState) && policy.ValidBlobUploadState(r.UploadState) &&
		!policy.ViolatesEvidenceBlobBridge(r.LifecycleState, r.ObjectBlobID != nil, r.UploadState)
}
func (r Row) StorageReferenceLegal() bool {
	return r.StorageRef == nil || policy.ServerManagedStorageRefMatchesAssociation(*r.StorageRef, r.ObjectBlobID)
}
func (r *Row) UTCClocks() {
	r.CreatedAt = r.CreatedAt.UTC()
	r.UpdatedAt = r.UpdatedAt.UTC()
	for _, clock := range []**time.Time{&r.RequestedAt, &r.ReceivedAt} {
		if *clock != nil {
			value := (**clock).UTC()
			*clock = &value
		}
	}
}
func (r Row) WithLocalActor(actor uuid.UUID) Row {
	if r.OwnerUserID != nil {
		r.OwnerUserID = &actor
	}
	return r
}
