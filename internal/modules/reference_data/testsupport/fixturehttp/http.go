package fixturehttp

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

// ImportPackFixture drives public import and Job observation. Callers own
// fresh runtime state and explicitly choose activation after this staged import.
func ImportPackFixture(t testing.TB, server *httptestx.Server, login authcookietest.AuthCookies, fixture testsupport.PortablePackFixture, transaction string) {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreatePart(textproto.MIMEHeader{"Content-Disposition": {`form-data; name="metadata"`}, "Content-Type": {"application/json"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = fmt.Fprintf(part, `{"client_txn_id":%q}`, transaction); err != nil {
		t.Fatal(err)
	}
	part, err = writer.CreatePart(textproto.MIMEHeader{"Content-Disposition": {`form-data; name="file"; filename="fixture.tar"`}, "Content-Type": {"application/x-tar"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = io.Copy(part, bytes.NewReader(fixture.Container)); err != nil {
		t.Fatal(err)
	}
	if err = writer.Close(); err != nil {
		t.Fatal(err)
	}
	request, err := http.NewRequest(http.MethodPost, server.HTTP.URL+"/api/v1/reference-packs/import", &body)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set(authn.CSRFHeaderName, login.CSRF.Value)
	request.AddCookie(login.Session)
	request.AddCookie(login.CSRF)
	job := httptestx.RequireSuccessEnvelope(t, httptestx.Do(t, http.DefaultClient, request), http.StatusAccepted)["data"].(map[string]any)
	httptestx.PollJob(t, server.HTTP.URL, job["job_id"].(string), httptestx.PollOptions{
		Timeout: 120 * time.Second, Interval: 25 * time.Millisecond, Request: []func(*http.Request){httptestx.WithCookies(login.Session)},
	}, "succeeded")
}
