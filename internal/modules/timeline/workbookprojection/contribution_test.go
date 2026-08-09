package workbookprojection

import "testing"

func TestNewContributionRequiresSource(t *testing.T) {
	t.Parallel()
	if _, err := NewContribution(nil); err == nil {
		t.Fatal("source-less Timeline projection contribution unexpectedly constructed")
	}
}
