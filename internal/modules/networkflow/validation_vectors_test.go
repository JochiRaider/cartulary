package networkflow

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
)

type endpointValidationVectors struct {
	IPs []struct {
		Name     string  `json:"name"`
		Input    *string `json:"input"`
		Accepted bool    `json:"accepted"`
	} `json:"ip_cases"`
	Pairs []struct {
		Name     string  `json:"name"`
		AIP      *string `json:"a_ip"`
		APort    *int32  `json:"a_port"`
		BIP      *string `json:"b_ip"`
		BPort    *int32  `json:"b_port"`
		Accepted bool    `json:"accepted"`
	} `json:"pair_cases"`
}

func readValidationVectors(t testing.TB, name string, target any) {
	t.Helper()
	raw, err := os.ReadFile("../../../fixtures/network-flow-validation/v1/" + name + ".json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, target); err != nil {
		t.Fatal(err)
	}
}

func TestNetworkFlowCanonicalEndpointVectors(t *testing.T) {
	var vectors endpointValidationVectors
	readValidationVectors(t, "endpoints", &vectors)
	for _, v := range vectors.IPs {
		t.Run(v.Name, func(t *testing.T) {
			accepted := false
			if v.Input != nil {
				ip, err := parseIPLiteral(*v.Input)
				accepted = err == nil && ip == *v.Input
			}
			if accepted != v.Accepted {
				t.Fatalf("canonical admission = %v, want %v", accepted, v.Accepted)
			}
		})
	}
	for _, v := range vectors.Pairs {
		t.Run("pair/"+v.Name, func(t *testing.T) {
			a, b := flowEndpoint{Port: v.APort}, flowEndpoint{Port: v.BPort}
			if v.AIP != nil {
				a.IP = *v.AIP
			}
			if v.BIP != nil {
				b.IP = *v.BIP
			}
			if got := canonicalEndpointsValid(a, b); got != v.Accepted {
				t.Fatalf("pair admission = %v, want %v", got, v.Accepted)
			}
		})
	}
}

// Both adapters consume independently authored expectations, not each other's output.
func AssertCanonicalEndpointSQLVectors(t testing.TB, db dbport.DB) {
	t.Helper()
	var vectors endpointValidationVectors
	readValidationVectors(t, "endpoints", &vectors)
	for _, v := range vectors.IPs {
		var accepted bool
		if err := db.QueryRow(t.Context(), `SELECT public.network_flow_ip_literal_is_canonical_v1($1)`, v.Input).Scan(&accepted); err != nil {
			t.Fatal(v.Name, err)
		}
		if accepted != v.Accepted {
			t.Fatalf("SQL %s = %v, want %v", v.Name, accepted, v.Accepted)
		}
	}
	for _, v := range vectors.Pairs {
		var accepted bool
		if err := db.QueryRow(t.Context(), `SELECT public.network_flow_endpoints_are_valid_v1($1,$2,$3,$4)`, v.AIP, v.APort, v.BIP, v.BPort).Scan(&accepted); err != nil {
			t.Fatal(v.Name, err)
		}
		if accepted != v.Accepted {
			t.Fatalf("SQL pair %s = %v, want %v", v.Name, accepted, v.Accepted)
		}
	}
}

func TestNetworkFlowDiagnosticIDV2KnownAnswers(t *testing.T) {
	var vectors []struct {
		Name        string  `json:"name"`
		Source      string  `json:"source_sha256"`
		Fingerprint string  `json:"mapping_fingerprint"`
		Row         int64   `json:"row"`
		Ordinal     *int64  `json:"ordinal"`
		Field       *string `json:"field"`
		Code        string  `json:"code"`
		Reason      string  `json:"reason"`
		Expected    string  `json:"expected"`
		Retired     string  `json:"retired"`
	}
	readValidationVectors(t, "diagnostic-ids", &vectors)
	for _, v := range vectors {
		t.Run(v.Name, func(t *testing.T) {
			got := diagnosticID(v.Source, v.Fingerprint, v.Row, v.Ordinal, v.Field, v.Code, v.Reason)
			if got != v.Expected {
				t.Fatalf("ID %s, want %s", got, v.Expected)
			}
			d := diagnostic(v.Row, v.Ordinal, nil, v.Field, v.Code, v.Reason, "untrusted")
			if v.Ordinal != nil {
				d.RawHeaderSHA256 = stringPtr(v.Source)
			}
			d.DiagnosticID = v.Expected
			if err := validateDiagnostic(d, v.Source, v.Fingerprint); err != nil {
				t.Fatal(err)
			}
			d.DiagnosticID = v.Retired
			if validateDiagnostic(d, v.Source, v.Fingerprint) == nil {
				t.Fatal("retired v1 identity admitted")
			}
		})
	}
}
