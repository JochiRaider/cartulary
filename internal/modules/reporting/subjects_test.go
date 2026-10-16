package reporting

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func testCanonicalSubjectAdmission(t *testing.T, snapshot SnapshotModel) {
	subject := snapshot.Subjects[0]
	if subject.CanonicalDisplayValueSHA256 != "db312f57547c3b4bf166b8e719421abc78ee308d54fa5619cf1bccaf95d24bd4" {
		t.Fatal("display digest differs from canonical byte vector")
	}
	raw, err := canonicalJSON(subject)
	if err != nil || bytes.Contains(raw, []byte("Host A")) || bytes.Contains(raw, []byte(`"display_name"`)) || bytes.Contains(raw, []byte(`"subject_kind"`)) {
		t.Fatal("canonical subject retained a raw or retired member", err)
	}
	var members map[string]json.RawMessage
	if json.Unmarshal(raw, &members) != nil || len(members) != 10 {
		t.Fatal("subject member set differs from Table 13-B")
	}
	// Required nullable members must survive as explicit null. Every omission
	// fails independently, even if the attacker recomputes the enclosing hash.
	for key := range members {
		var changed map[string]json.RawMessage
		_ = json.Unmarshal(raw, &changed)
		delete(changed, key)
		data, _ := json.Marshal(changed)
		var decoded TokenizableSubject
		if json.Unmarshal(data, &decoded) == nil {
			t.Fatal("missing subject member accepted", key)
		}
	}
	for _, key := range []string{"subject_kind", "display_name", "source_record_ref"} {
		var changed map[string]json.RawMessage
		_ = json.Unmarshal(raw, &changed)
		changed[key] = json.RawMessage(`"private"`)
		data, _ := json.Marshal(changed)
		var decoded TokenizableSubject
		if json.Unmarshal(data, &decoded) == nil {
			t.Fatal("retired subject shape accepted", key)
		}
	}
	for name, change := range map[string]func(*TokenizableSubject){
		"altered digest":      func(s *TokenizableSubject) { s.CanonicalDisplayValueSHA256 = strings.Repeat("f", 64) },
		"null source id":      func(s *TokenizableSubject) { s.SourceRecordID = nil },
		"null source type":    func(s *TokenizableSubject) { s.SourceRecordType = nil },
		"mixed identity":      func(s *TokenizableSubject) { id := "mention"; s.EntityMentionID = &id },
		"wrong source family": func(s *TokenizableSubject) { kind := "identity"; s.SourceRecordType = &kind },
		"lost source binding": func(s *TokenizableSubject) { s.SourceRefs = []string{} },
		"invented support":    func(s *TokenizableSubject) { s.SupportRefs = []string{"assignment-is-not-support"} },
		"lost restrictions":   func(s *TokenizableSubject) { s.DisclosurePartitionRefs = []string{} },
	} {
		t.Run(name, func(t *testing.T) {
			candidate := snapshot
			candidate.Subjects = append([]TokenizableSubject{}, snapshot.Subjects...)
			change(&candidate.Subjects[0])
			if validateSnapshotModelIdentity(candidate) == nil {
				t.Fatal("altered subject admitted")
			}
			data, _ := canonicalJSON(candidate)
			decoded, _, err := decodePortableObject[SnapshotModel](data, reportingObjectDigest(candidate.SchemaID, data))
			if err == nil && validateSnapshotModelIdentity(decoded) == nil {
				t.Fatal("self-consistent tampered subject recovered")
			}
		})
	}
	// Nullable identity is not inferred from a stable-ref string.
	var decoded TokenizableSubject
	if json.Unmarshal(bytes.Replace(raw, []byte(`"entity_mention_id":null`), []byte(`"entity_mention_id":null,"entity_mention_id":"other"`), 1), &decoded) == nil {
		t.Fatal("duplicate subject member accepted")
	}
}
