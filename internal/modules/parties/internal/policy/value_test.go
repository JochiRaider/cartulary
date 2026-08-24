package policy_test

import (
	"encoding/json"
	"testing"

	contractparties "github.com/JochiRaider/cartulary/internal/gen/contractparties"
	"github.com/JochiRaider/cartulary/internal/modules/parties/internal/policy"
)

func TestNormalizationCorpus(t *testing.T) {
	t.Parallel()
	var corpus struct {
		Cases []struct {
			CaseID             string  `json:"case_id"`
			FieldKey           string  `json:"field_key"`
			Input              *string `json:"input"`
			Accepted           bool    `json:"accepted"`
			ReasonCode         string  `json:"reason_code"`
			StoredValue        *string `json:"stored_value"`
			EqualityValue      *string `json:"equality_value"`
			ClaimValue         *string `json:"claim_value"`
			CanonicalHashValue any     `json:"canonical_hash_value"`
		} `json:"cases"`
	}
	if err := json.Unmarshal([]byte(contractparties.Artifacts[1].JSON), &corpus); err != nil {
		t.Fatalf("decode generated Party normalization corpus: %v", err)
	}
	for _, testCase := range corpus.Cases {
		testCase := testCase
		t.Run(testCase.CaseID, func(t *testing.T) {
			t.Parallel()
			value, admissionErr := policy.Admit(testCase.FieldKey, testCase.Input)
			if !testCase.Accepted {
				if admissionErr == nil || admissionErr.ReasonCode != testCase.ReasonCode {
					t.Fatalf("admission error = %#v, want reason %q", admissionErr, testCase.ReasonCode)
				}
				return
			}
			if admissionErr != nil {
				t.Fatalf("admit: %v", admissionErr)
			}
			requireOptional(t, "stored", value.StoredValue, testCase.StoredValue)
			requireOptional(t, "equality", value.EqualityValue, testCase.EqualityValue)
			requireOptional(t, "claim", value.ExactMatchClaimValue, testCase.ClaimValue)
			if got, want := value.CanonicalHashValue(), testCase.CanonicalHashValue; got != want {
				t.Fatalf("canonical hash value = %#v, want %#v", got, want)
			}
		})
	}
}

func requireOptional(t *testing.T, name string, getter func() (string, bool), want *string) {
	t.Helper()
	got, present := getter()
	if want == nil {
		if present {
			t.Fatalf("%s value = %q, want null", name, got)
		}
		return
	}
	if !present || got != *want {
		t.Fatalf("%s value = %q/%t, want %q", name, got, present, *want)
	}
}
