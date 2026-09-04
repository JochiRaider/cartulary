package postgres

import "testing"

func TestAdmissionFactsRequireExactEngineAndSecurityPosture(t *testing.T) {
	tests := []struct {
		name   string
		facts  admissionFacts
		reason string
	}{
		{name: "exact baseline", facts: admissionFacts{sessionUser: "login", currentUser: "role", serverVersionNum: 180006, dataChecksums: "on"}},
		{name: "18.4", facts: admissionFacts{sessionUser: "login", currentUser: "role", serverVersionNum: 180004, dataChecksums: "on"}, reason: ReasonServerVersionMismatch},
		{name: "18.7", facts: admissionFacts{sessionUser: "login", currentUser: "role", serverVersionNum: 180007, dataChecksums: "on"}, reason: ReasonServerVersionMismatch},
		{name: "16", facts: admissionFacts{sessionUser: "login", currentUser: "role", serverVersionNum: 160000, dataChecksums: "on"}, reason: ReasonServerVersionMismatch},
		{name: "19", facts: admissionFacts{sessionUser: "login", currentUser: "role", serverVersionNum: 190000, dataChecksums: "on"}, reason: ReasonServerVersionMismatch},
		{name: "checksums off", facts: admissionFacts{sessionUser: "login", currentUser: "role", serverVersionNum: 180006, dataChecksums: "off"}, reason: ReasonDataChecksumsDisabled},
		{name: "wrong login", facts: admissionFacts{sessionUser: "other", currentUser: "role", serverVersionNum: 180006, dataChecksums: "on"}, reason: ReasonEffectiveRoleMismatch},
		{name: "wrong role", facts: admissionFacts{sessionUser: "login", currentUser: "other", serverVersionNum: 180006, dataChecksums: "on"}, reason: ReasonEffectiveRoleMismatch},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateAdmissionFacts(test.facts, "login", "role")
			if test.reason == "" {
				if err != nil {
					t.Fatalf("exact admission failed: %v", err)
				}
				return
			}
			configurationErr, ok := err.(*ConfigurationError)
			if !ok || configurationErr.Reason() != test.reason || configurationErr.Error() != test.reason {
				t.Fatalf("admission error = %T %v, want reason %q", err, err, test.reason)
			}
		})
	}
}

func TestAdmissionErrorRedactsUnderlyingFailure(t *testing.T) {
	err := admittedConnectionError(assertionError("password=secret host=private.example vendor detail"))
	if err.Error() != ReasonAdmissionFailed {
		t.Fatalf("admission error leaked detail: %q", err)
	}
}

type assertionError string

func (err assertionError) Error() string { return string(err) }
