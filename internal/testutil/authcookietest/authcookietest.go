package authcookietest

import (
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type AuthCookies struct {
	Session *http.Cookie
	CSRF    *http.Cookie
}

func RequireAuthCookies(t testing.TB, cookies []*http.Cookie) AuthCookies {
	t.Helper()

	authCookies := AuthCookies{
		Session: requireCookieByName(t, cookies, authn.SessionCookieName),
		CSRF:    requireCookieByName(t, cookies, authn.CSRFCookieName),
	}
	requireCookiePolicy(t, authCookies.Session, authCookiePolicy{
		name:     authn.SessionCookieName,
		httpOnly: true,
		secure:   true,
		path:     "/",
		sameSite: http.SameSiteLaxMode,
	})
	requireCookiePolicy(t, authCookies.CSRF, authCookiePolicy{
		name:     authn.CSRFCookieName,
		httpOnly: false,
		secure:   true,
		path:     "/",
		sameSite: http.SameSiteLaxMode,
	})
	return authCookies
}

type authCookiePolicy struct {
	name     string
	httpOnly bool
	secure   bool
	path     string
	sameSite http.SameSite
}

func requireCookieByName(t testing.TB, cookies []*http.Cookie, name string) *http.Cookie {
	t.Helper()
	for _, cookie := range cookies {
		if cookie.Name == name {
			if cookie.Value == "" {
				t.Fatalf("expected %s cookie value to be populated, got %#v", name, cookie)
			}
			return cookie
		}
	}
	t.Fatalf("expected %s cookie to be set, got %#v", name, cookies)
	return nil
}

func requireCookiePolicy(t testing.TB, cookie *http.Cookie, policy authCookiePolicy) {
	t.Helper()
	if cookie.Path != policy.path {
		t.Fatalf("unexpected %s cookie path: got %q want %q", policy.name, cookie.Path, policy.path)
	}
	if cookie.HttpOnly != policy.httpOnly {
		t.Fatalf("unexpected %s cookie HttpOnly: got %v want %v", policy.name, cookie.HttpOnly, policy.httpOnly)
	}
	if cookie.Secure != policy.secure {
		t.Fatalf("unexpected %s cookie Secure: got %v want %v", policy.name, cookie.Secure, policy.secure)
	}
	if cookie.SameSite != policy.sameSite {
		t.Fatalf("unexpected %s cookie SameSite: got %v want %v", policy.name, cookie.SameSite, policy.sameSite)
	}
}
