package authcookietest

import (
	"fmt"
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
	result, err := validateAuthCookies(cookies)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func validateAuthCookies(cookies []*http.Cookie) (AuthCookies, error) {
	session, err := validateCookie(cookies, authn.SessionCookieName, true, true)
	if err != nil {
		return AuthCookies{}, err
	}
	csrf, err := validateCookie(cookies, authn.CSRFCookieName, false, false)
	if err != nil {
		return AuthCookies{}, err
	}
	return AuthCookies{Session: session, CSRF: csrf}, nil
}

func validateCookie(cookies []*http.Cookie, name string, httpOnly, allowStrict bool) (*http.Cookie, error) {
	for _, cookie := range cookies {
		if cookie == nil || cookie.Name != name {
			continue
		}
		if cookie.Value == "" {
			return nil, fmt.Errorf("%s cookie value must be populated", name)
		}
		if cookie.Path != "/" {
			return nil, fmt.Errorf("%s cookie Path must be /", name)
		}
		if cookie.HttpOnly != httpOnly {
			return nil, fmt.Errorf("%s cookie HttpOnly must be %t", name, httpOnly)
		}
		if !cookie.Secure {
			return nil, fmt.Errorf("%s cookie Secure must be true", name)
		}
		if cookie.SameSite != http.SameSiteLaxMode && !(allowStrict && cookie.SameSite == http.SameSiteStrictMode) {
			return nil, fmt.Errorf("%s cookie SameSite violates the authentication cookie policy", name)
		}
		return cookie, nil
	}
	return nil, fmt.Errorf("%s cookie must be set", name)
}
