package authcookietest

import (
	"net/http"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func validCookies() []*http.Cookie {
	return []*http.Cookie{
		{Name: authn.SessionCookieName, Value: "private-session-value", Path: "/", HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode},
		{Name: authn.CSRFCookieName, Value: "private-csrf-value", Path: "/", Secure: true, SameSite: http.SameSiteLaxMode},
	}
}

func TestRequireAuthCookies(t *testing.T) {
	for _, tc := range []struct {
		name string
		mode http.SameSite
	}{
		{"lax session", http.SameSiteLaxMode},
		{"strict session", http.SameSiteStrictMode},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cookies := validCookies()
			cookies[0].SameSite = tc.mode
			got := RequireAuthCookies(t, cookies)
			if got.Session != cookies[0] || got.CSRF != cookies[1] {
				t.Fatal("authentication cookies must retain the issued cookie pair")
			}
		})
	}
}

func TestAuthCookiePolicyRejectsInvalidCookies(t *testing.T) {
	for _, target := range []struct {
		name  string
		index int
	}{
		{authn.SessionCookieName, 0},
		{authn.CSRFCookieName, 1},
	} {
		for _, tc := range []struct {
			name     string
			property string
			mutate   func(*http.Cookie)
		}{
			{"missing", "set", func(c *http.Cookie) { c.Name = "unrelated" }},
			{"empty", "value", func(c *http.Cookie) { c.Value = "" }},
			{"insecure", "Secure", func(c *http.Cookie) { c.Secure = false }},
			{"wrong path", "Path", func(c *http.Cookie) { c.Path = "/restricted" }},
			{"wrong visibility", "HttpOnly", func(c *http.Cookie) { c.HttpOnly = !c.HttpOnly }},
			{"unspecified same site", "SameSite", func(c *http.Cookie) { c.SameSite = 0 }},
			{"default same site", "SameSite", func(c *http.Cookie) { c.SameSite = http.SameSiteDefaultMode }},
			{"cross site", "SameSite", func(c *http.Cookie) { c.SameSite = http.SameSiteNoneMode }},
		} {
			t.Run(target.name+"/"+tc.name, func(t *testing.T) {
				cookies := validCookies()
				tc.mutate(cookies[target.index])
				got, err := validateAuthCookies(cookies)
				if err == nil || got.Session != nil || got.CSRF != nil {
					t.Fatal("invalid cookie policy must fail without returning credentials")
				}
				if !strings.Contains(err.Error(), target.name) || !strings.Contains(err.Error(), tc.property) {
					t.Fatal("policy failure must identify its cookie and property")
				}
				if strings.Contains(err.Error(), "private-session-value") || strings.Contains(err.Error(), "private-csrf-value") {
					t.Fatal("policy failure disclosed credential material")
				}
			})
		}
	}
	t.Run("csrf remains lax", func(t *testing.T) {
		cookies := validCookies()
		cookies[1].SameSite = http.SameSiteStrictMode
		if _, err := validateAuthCookies(cookies); err == nil {
			t.Fatal("CSRF cookie must retain the supported double-submit policy")
		}
	})
}
