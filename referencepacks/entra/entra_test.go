package producer

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	entrainputs "github.com/JochiRaider/cartulary/referencepacks/entra/inputs"
)

func TestEntraSourceBoundaries_Unit(t *testing.T) {
	heading := "Synthetic applications"
	source := EntraSource{ID: "S3", Heading: &heading, Header: []string{"Application Name", "Application IDs"}, Rows: 1}
	input := "## Unselected\n|Application Name|Application IDs|\n|--|--|\n|Other|not-a-guid|\n## Synthetic applications\n|Application Name|Application IDs|\n|--|--|\n|Two names|00000003-0000-0000-c000-000000000000<br>00000003-0000-0ff1-ce00-000000000000| \n## Next\n|Application Name|Application IDs|\n|--|--|\n|Later|unselected|\n"
	rows, unselected, err := extractEntra(source, []byte(input))
	if err != nil || len(rows) != 1 || len(unselected) != 2 || rows[0].line != 8 || !strings.HasSuffix(rows[0].raw, " ") {
		t.Fatal("section boundary or exact locator", rows, unselected, err)
	}
	crlf, _, err := extractEntra(source, []byte(strings.ReplaceAll(input, "\n", "\r\n")))
	if err != nil || crlf[0].line != rows[0].line || crlf[0].raw != rows[0].raw {
		t.Fatal("CRLF provenance", err)
	}
	for _, bad := range []string{strings.Replace(input, "|Application Name|Application IDs|", "|Name|IDs|", -1), strings.Replace(input, "|Two names|", "|Two|columns|", 1), input + "## Synthetic applications\n", strings.Replace(input, "## Synthetic applications", "## Renamed section", 1), strings.Replace(input, "|Two names|", "|Two names\x00|", 1)} {
		got, _, e := extractEntra(source, []byte(bad))
		if e == nil && len(got) > 0 {
			_, e = entraCell(got[0].cells[0], false)
		}
		if e == nil {
			t.Fatal("changed structure/control accepted")
		}
	}
	for _, bad := range []string{"<script>x</script>", "[name](https://example.invalid)", "name\x00", "**name**"} {
		if _, err := entraCell(bad, false); err == nil {
			t.Fatal("markup/control accepted", bad)
		}
	}
	if _, err := entraCell("ordinary<br/>description", true); err != nil {
		t.Fatal(err)
	}
	if _, err := entraCell("ordinary<br>description", true); err == nil {
		t.Fatal("unadopted markup accepted")
	}
	for _, id := range []string{"00000003-0000-0000-c000-000000000000", "00000003-0000-0ff1-ce00-000000000000"} {
		if !entraUUID.MatchString(id) {
			t.Fatal("Microsoft lexical UUID rejected")
		}
	}
	for _, id := range []string{"39d63e7-7fa3-4b2b-94ea-ee256fdb8c2f", "{00000003-0000-0000-c000-000000000000}", "urn:uuid:00000003-0000-0000-c000-000000000000", " 00000003-0000-0000-c000-000000000000"} {
		if entraUUID.MatchString(id) {
			t.Fatal("UUID repaired")
		}
	}
	if entraASCII("ABCÉ") != "abcÉ" {
		t.Fatal("non-ASCII folding")
	}
	for _, bad := range []string{"301", "-1", "9223372036854775807", "tomorrow"} {
		if _, err := entraRetryDelay(bad, time.Now()); err == nil {
			t.Fatal("unbounded retry", bad)
		}
	}
}

type corruptEntraSource struct{}

func (corruptEntraSource) FetchEntra(context.Context, EntraSource) (EntraDownload, error) {
	return EntraDownload{[]byte("synthetic corrupt bytes"), "2026-01-01T00:00:00Z", "text/plain"}, nil
}

func TestEntraSourceIntegrity_Unit(t *testing.T) {
	inventory, err := entrainputs.Files.ReadFile("inventory.v2.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entraInventory(inventory); err != nil {
		t.Fatal(err)
	}
	if _, err := entraInventory(append(inventory, ' ')); err == nil {
		t.Fatal("changed immutable input accepted")
	}
	root, err := os.OpenRoot(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	if _, err := EntraAcquire(t.Context(), root, inventory, corruptEntraSource{}); err == nil {
		t.Fatal("bad source digest accepted")
	}
	if _, err := root.Lstat("source-lock.json"); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("failed acquisition published complete lock", err)
	}
	if _, err := EntraCheck(root, inventory); err == nil {
		t.Fatal("missing lock accepted")
	}
	if err := root.WriteFile("source-lock.json", []byte(strings.Repeat("x", 65537)), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := EntraCheck(root, inventory); err == nil {
		t.Fatal("unbounded source lock admitted")
	}
	if _, err := (EntraHTTPFetcher{}).FetchEntra(t.Context(), EntraSource{ID: "unlisted", URL: "https://example.invalid"}); err == nil {
		t.Fatal("unlisted source admitted")
	}

}
