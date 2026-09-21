package artifacts

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/strictjson"
	"github.com/google/uuid"
)

type NoteAssociationKind string

const (
	NoteAssociationSource      NoteAssociationKind = "source"
	NoteAssociationEvidence    NoteAssociationKind = "evidence"
	NoteAssociationRelatedNote NoteAssociationKind = "related_note"
)

func ParseNoteAssociationKind(value string) (NoteAssociationKind, bool) {
	kind := NoteAssociationKind(value)
	return kind, kind == NoteAssociationSource || kind == NoteAssociationEvidence || kind == NoteAssociationRelatedNote
}

type noteAssociationAction struct {
	Op            string    `json:"op"`
	CounterpartID uuid.UUID `json:"counterpart_record_id,omitempty"`
	LinkID        uuid.UUID `json:"link_id,omitempty"`
}

type NoteAssociationAdmission struct {
	kind        NoteAssociationKind
	baseVersion int64
	clientTxnID string
	actions     []noteAssociationAction
	hash        [sha256.Size]byte
}

func (a NoteAssociationAdmission) ClientTxnID() string { return a.clientTxnID }
func (a NoteAssociationAdmission) valid() bool {
	return a.baseVersion > 0 && len(a.actions) > 0 && a.clientTxnID != ""
}
func (a NoteAssociationAdmission) requestHash() []byte { return append([]byte(nil), a.hash[:]...) }

func associationItemRef(id uuid.UUID) string { return base64.RawURLEncoding.EncodeToString(id[:]) }
func associationItemID(ref string) (uuid.UUID, bool) {
	b, err := base64.RawURLEncoding.DecodeString(ref)
	if err != nil || len(b) != 16 {
		return uuid.Nil, false
	}
	id, err := uuid.FromBytes(b)
	return id, err == nil && id != uuid.Nil && associationItemRef(id) == ref
}

func AdmitNoteAssociations(reader io.Reader) (NoteAssociationAdmission, *AdmissionError) {
	raw, err := strictjson.DecodeObject(reader)
	if err != nil {
		return NoteAssociationAdmission{}, newAdmissionError("payload", admissionRequestNotObject)
	}
	for key := range raw {
		if key != "kind" && key != "base_row_version" && key != "client_txn_id" && key != "actions" {
			return NoteAssociationAdmission{}, newAdmissionError(key, admissionUnknownField)
		}
	}
	var a NoteAssociationAdmission
	var kind string
	if json.Unmarshal(raw["kind"], &kind) != nil {
		return a, newAdmissionError("kind", admissionInvalidValue)
	}
	var valid bool
	if a.kind, valid = ParseNoteAssociationKind(kind); !valid {
		return a, newAdmissionError("kind", admissionInvalidValue)
	}
	if json.Unmarshal(raw["base_row_version"], &a.baseVersion) != nil || a.baseVersion < 1 {
		return a, newAdmissionError("base_row_version", admissionInvalidBaseRowVersion)
	}
	if json.Unmarshal(raw["client_txn_id"], &a.clientTxnID) != nil || strings.TrimSpace(a.clientTxnID) == "" || len(a.clientTxnID) > 128 {
		return a, newAdmissionError("client_txn_id", admissionInvalidValue)
	}
	var actions []json.RawMessage
	if json.Unmarshal(raw["actions"], &actions) != nil || len(actions) < 1 || len(actions) > 64 {
		return a, newAdmissionError("actions", admissionInvalidValue)
	}
	for _, item := range actions {
		fields, err := strictjson.DecodeObject(strings.NewReader(string(item)))
		if err != nil || len(fields) != 2 {
			return NoteAssociationAdmission{}, newAdmissionError("actions", admissionInvalidValue)
		}
		var action noteAssociationAction
		if json.Unmarshal(fields["op"], &action.Op) != nil {
			return NoteAssociationAdmission{}, newAdmissionError("actions", admissionInvalidValue)
		}
		switch action.Op {
		case "add":
			id, ok := artifactUUIDActionField(fields, "counterpart_record_id")
			if !ok || id == uuid.Nil {
				return NoteAssociationAdmission{}, newAdmissionError("actions", admissionInvalidValue)
			}
			action.CounterpartID = id
		case "remove":
			var ref string
			if json.Unmarshal(fields["item_ref"], &ref) != nil {
				return NoteAssociationAdmission{}, newAdmissionError("actions", admissionInvalidValue)
			}
			id, ok := associationItemID(ref)
			if !ok {
				return NoteAssociationAdmission{}, newAdmissionError("actions", admissionInvalidValue)
			}
			action.LinkID = id
		default:
			return NoteAssociationAdmission{}, newAdmissionError("actions", admissionInvalidValue)
		}
		a.actions = append(a.actions, action)
	}
	// Preserve action order. Transaction identity is separately bound by the key.
	a.hash = hashArtifactMutationPayload(map[string]any{"kind": a.kind, "base_row_version": a.baseVersion, "actions": a.actions})
	return a, nil
}
