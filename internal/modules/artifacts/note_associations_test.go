package artifacts

import (
	"bytes"
	"strings"
	"testing"
)

func TestNoteAssociationAdmission(t *testing.T) {
	const add = `{"op":"add","counterpart_record_id":"11111111-2222-4333-8444-555555555555"}`
	valid := `{"kind":"source","base_row_version":2,"client_txn_id":"association-1","actions":[` + add + `]}`
	admitted, failure := AdmitNoteAssociations(strings.NewReader(valid))
	if failure != nil || !admitted.valid() {
		t.Fatalf("valid association rejected: %v", failure)
	}
	for _, payload := range []string{
		strings.Replace(valid, `"kind":"source"`, `"kind":"arbitrary"`, 1),
		strings.Replace(valid, `"base_row_version":2`, `"base_row_version":0`, 1),
		strings.Replace(valid, add, `{"op":"remove","item_ref":"forged"}`, 1),
		strings.Replace(valid, add, `{"op":"add","counterpart_record_id":"11111111-2222-4333-8444-555555555555","link_type":"derived_from"}`, 1),
		strings.Replace(valid, `[`+add+`]`, `[]`, 1),
		strings.Replace(valid, `[`+add+`]`, `[`+strings.TrimSuffix(strings.Repeat(add+",", 65), ",")+`]`, 1),
		strings.Replace(valid, `"kind":"source"`, `"kind":"source","kind":"evidence"`, 1),
		valid + `{}`,
	} {
		if _, err := AdmitNoteAssociations(strings.NewReader(payload)); err == nil {
			t.Fatalf("invalid association admitted: %s", payload)
		}
	}
	other, failure := AdmitNoteAssociations(strings.NewReader(strings.Replace(valid, `"source"`, `"evidence"`, 1)))
	if failure != nil || bytes.Equal(admitted.requestHash(), other.requestHash()) {
		t.Fatal("association kind must be part of request identity")
	}
	if (NoteAssociationAdmission{}).valid() {
		t.Fatal("zero admission was accepted")
	}
}
