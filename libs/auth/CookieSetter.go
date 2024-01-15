package auth

import (
	"net/http"
	"time"
)

type CookieSetter interface {
	GetSessionCookie(request *http.Request) (value string)
	SetSessionCookie(writer http.ResponseWriter, value string, expiredIn time.Duration)
	DeleteSessionCookie(writer http.ResponseWriter)
}

type DefaultCookieSetter struct {
	CookieSameSite http.SameSite
	CookieSecure   bool
	CookieHttpOnly bool
	CookieDomain   string
}

func (d *DefaultCookieSetter) GetSessionCookie(request *http.Request) (value string) {
	token, err := request.Cookie("token")
	if err != nil { // The only possible error is cookie not found.
		return ""
	}
	return token.Value
}

func (d *DefaultCookieSetter) SetSessionCookie(writer http.ResponseWriter, value string, expiredIn time.Duration) {
	http.SetCookie(writer, &http.Cookie{
		Name:     "token",
		Value:    value,
		Path:     "/",
		Expires:  time.Now().Add(expiredIn),
		HttpOnly: d.CookieHttpOnly,
		SameSite: d.CookieSameSite,
		Secure:   d.CookieSecure,
		Domain:   d.CookieDomain,
	})
}

func (d *DefaultCookieSetter) DeleteSessionCookie(writer http.ResponseWriter) {
	http.SetCookie(writer, &http.Cookie{
		Name:     "token",
		MaxAge:   -1,
		Path:     "/",
		HttpOnly: d.CookieHttpOnly,
		SameSite: d.CookieSameSite,
		Secure:   d.CookieSecure,
		Domain:   d.CookieDomain,
	})
}
