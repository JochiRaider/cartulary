package graphprojection

import "testing"

func TestGraphProjectionUnitBehaviorFixtures_Unit(t *testing.T) {
	fixtures := []struct {
		id  string
		run func(*testing.T)
	}{
		{"GP-FIX-001", TestGPFIX001Remediation},
		{"GP-FIX-002", TestGPFIX002Remediation},
		{"GP-FIX-003", TestGPFIX003Remediation},
		{"GP-FIX-005", TestGPFIX005Remediation},
		{"GP-FIX-006", TestGPFIX006Remediation},
		{"GP-FIX-007", TestGPFIX007Remediation},
		{"GP-FIX-008", TestGPFIX008Remediation},
		{"GP-FIX-009", TestGPFIX009Remediation},
		{"GP-FIX-010", TestGPFIX010Remediation},
		{"GP-FIX-011", TestGPFIX011Remediation},
		{"GP-FIX-012", TestGPFIX012Remediation},
		{"GP-FIX-013", TestGPFIX013Remediation},
		{"GP-FIX-014", TestGPFIX014CanonicalJSONString},
		{"GP-FIX-015", TestGPFIX015IntegerLexicalBoundaries},
		{"GP-FIX-016", TestGPFIX016TimestampCalendarValidity},
		{"GP-FIX-017", TestGPFIX017IdentifierUnicodeWhitespace},
		{"GP-FIX-022", TestGPFIX022MinimalGraphDigestTranscript},
		{"GP-FIX-023", TestGPFIX023HostGraphDigestTranscript},
		{"GP-FIX-024", TestGPFIX024Remediation},
		{"GP-FIX-025", TestGPFIX025Remediation},
		{"GP-FIX-026", TestGPFIX026Remediation},
		{"GP-FIX-027", TestGPFIX027Remediation},
		{"GP-FIX-028", TestGPFIX028Remediation},
		{"GP-FIX-034", TestGPFIX034Remediation},
		{"GP-FIX-036", TestGPFIX036DistinctSortedArrayCanonicalOrder},
	}
	for _, fixture := range fixtures {
		t.Run(fixture.id, fixture.run)
	}
}

func TestGPFIX001Remediation(t *testing.T) {
	verifyUnitFixture(t, "GP-FIX-001")
}

func TestGPFIX002Remediation(t *testing.T) {
	verifyUnitFixture(t, "GP-FIX-002")
}

func TestGPFIX003Remediation(t *testing.T) {
	verifyUnitFixture(t, "GP-FIX-003")
}

func TestGPFIX005Remediation(t *testing.T) {
	verifyUnitFixture(t, "GP-FIX-005")
}

func TestGPFIX006Remediation(t *testing.T) {
	verifyUnitFixture(t, "GP-FIX-006")
}

func TestGPFIX007Remediation(t *testing.T) {
	verifyUnitFixture(t, "GP-FIX-007")
}

func TestGPFIX008Remediation(t *testing.T) {
	verifyUnitFixture(t, "GP-FIX-008")
}

func TestGPFIX009Remediation(t *testing.T) {
	verifyUnitFixture(t, "GP-FIX-009")
}

func TestGPFIX010Remediation(t *testing.T) {
	verifyUnitFixture(t, "GP-FIX-010")
}

func TestGPFIX011Remediation(t *testing.T) {
	verifyUnitFixture(t, "GP-FIX-011")
}

func TestGPFIX012Remediation(t *testing.T) {
	verifyUnitFixture(t, "GP-FIX-012")
}

func TestGPFIX013Remediation(t *testing.T) {
	verifyUnitFixture(t, "GP-FIX-013")
}

func TestGPFIX024Remediation(t *testing.T) {
	verifyUnitFixture(t, "GP-FIX-024")
}

func TestGPFIX025Remediation(t *testing.T) {
	verifyUnitFixture(t, "GP-FIX-025")
}

func TestGPFIX026Remediation(t *testing.T) {
	verifyUnitFixture(t, "GP-FIX-026")
}

func TestGPFIX027Remediation(t *testing.T) {
	verifyUnitFixture(t, "GP-FIX-027")
}

func TestGPFIX028Remediation(t *testing.T) {
	verifyUnitFixture(t, "GP-FIX-028")
}

func TestGPFIX034Remediation(t *testing.T) {
	verifyUnitFixture(t, "GP-FIX-034")
}
