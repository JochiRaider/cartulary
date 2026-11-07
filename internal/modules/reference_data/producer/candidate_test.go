package producer

import (
	"context"
	"os"
	"testing"
)

func TestATTACKNativeCandidate(t *testing.T) {
	root, err := os.OpenRoot(contractRoot)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	files, err := ReadCandidate(root, fixtureBytes(t, "candidate-index.v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cases struct {
		Schema  string `json:"schema_id"`
		Purpose string `json:"purpose"`
		Cases   []struct {
			Name   string `json:"name"`
			Schema string `json:"schema_file"`
			Value  any    `json:"value"`
			Valid  bool   `json:"valid"`
		} `json:"cases"`
	}
	if err := DecodeClosed(files["fixtures/native-cases.v1.json"], &cases); err != nil {
		t.Fatal(err)
	}
	if cases.Schema != "cartulary.attack_native_cases.v1" || len(cases.Cases) == 0 {
		t.Fatal("native fixture identity invalid")
	}
	seen := map[string]bool{}
	for _, tc := range cases.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			if seen[tc.Name] {
				t.Fatal("duplicate native expectation")
			}
			seen[tc.Name] = true
			schema, err := CompileCandidate(files[tc.Schema], "")
			if err != nil {
				t.Fatal(err)
			}
			if err = schema.Validate(tc.Value); (err == nil) != tc.Valid {
				t.Fatalf("candidate valid=%t, error=%v", tc.Valid, err)
			}
		})
	}
	retained := testRoot(t)
	lock, fetch := sourceFixture(t, "valid")
	if _, err := Acquire(context.Background(), retained, lock, fetch); err != nil {
		t.Fatal(err)
	}
	report, err := ValidateNativeCandidate(context.Background(), retained, lock, fixtureBytes(t, "candidate-index.v1.json"), []byte("{"))
	if err == nil || report.Status != "proposal_shape_failed" {
		t.Fatal("schema compilation failure claimed a passing proposal", err)
	}
	// Independent SHA-256 witness, not a value produced by the code under test.
	if Digest([]byte("abc")) != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Fatal("SHA-256 witness failed")
	}
}
