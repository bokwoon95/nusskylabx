// Package cookies provides cookie related utilities
package cookies

import (
	"errors"
	"net/http"
	"time"

	"github.com/bokwoon95/nusskylabx/helpers/auth"
	"github.com/bokwoon95/nusskylabx/helpers/erro"
)

const (
	ErrNoCookie erro.BaseError = "Cookie with name '%s' not found"
)

// CookieOpt are cookie options that change the nature of the cookie being set.
type CookieOpt func(*http.Cookie) *http.Cookie

// AllowJS sets whether the cookie can be read by client side javascript. If
// there is no reason for javascript to access the cookie, do not turn it on to
// make the cookie more secure. Default is false.
func AllowJS(allowJS bool) CookieOpt {
	return func(cookie *http.Cookie) *http.Cookie {
		cookie.HttpOnly = !allowJS
		return cookie
	}
}

// MustHTTPS sets whether the cookie must be sent over a HTTPS connection.
// Default is false.
func MustHTTPS(mustHttps bool) CookieOpt {
	return func(cookie *http.Cookie) *http.Cookie {
		cookie.Secure = mustHttps
		return cookie
	}
}

// Duration sets how long the cookie is valid for. Default is three months.
func Duration(duration time.Duration) CookieOpt {
	return func(cookie *http.Cookie) *http.Cookie {
		cookie.MaxAge = int(duration.Seconds())
		return cookie
	}
}

// SetCookie will set a cookie of the specified name and value in the client
// browser.
//
// It takes in additional options which may change various settings of the
// cookie, such as the duration of the cookie and whether the cookie should be
// readable by client side javascript.
func SetCookie(w http.ResponseWriter, name, value string, opts ...CookieOpt) {
	threemonths := time.Hour * 24 * 30 * 3
	// default cookie configuration
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		HttpOnly: true,                       // disallow JavaScript access to cookie
		Secure:   false,                      // allow HTTP (instead of restricting it to HTTPS)
		SameSite: http.SameSiteLaxMode,       // CSRF countermeasure
		MaxAge:   int(threemonths.Seconds()), // three months
		Path:     "/",
	}
	// apply cookie options
	for _, opt := range opts {
		if opt != nil { // guard against runtime panics
			cookie = opt(cookie)
		}
	}
	http.SetCookie(w, cookie)
}

// RequestAddCookie will add a cookie directly to a request. Normally SetCookie
// will set a cookie in the client browser, which will then be sent back to the
// server for every request. RequestAddCookie operates differently by adding
// the cookie for a single request. This is useful in tests where there is no
// client browser to set cookies for, in which case you can just attach the
// cookie directly to a test request instead.
func RequestAddCookie(r *http.Request, name, value string, opts ...CookieOpt) {
	threemonths := time.Hour * 24 * 30 * 3
	// default cookie configuration
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		HttpOnly: true,                       // disallow JavaScript access to cookie
		Secure:   false,                      // allow HTTP (instead of restricting it to HTTPS)
		SameSite: http.SameSiteLaxMode,       // CSRF countermeasure
		MaxAge:   int(threemonths.Seconds()), // three months
		Path:     "/",
	}
	// apply cookie options
	for _, opt := range opts {
		if opt != nil { // guard against runtime panics
			cookie = opt(cookie)
		}
	}
	r.AddCookie(cookie)
}

// GetCookieValue will read a value from a cookie of the specified name. If the
// cookie doesn't exist, an empty string is returned.
//
// It is a very simple wrapper over directly calling (*http.Request).Cookie().
func GetCookieValue(r *http.Request, name string) string {
	c, _ := r.Cookie(name)
	if c != nil {
		return c.Value
	}
	return ""
}

// SetCookieOneMinute is a wrapper around SetCookie with the duration of the
// cookie set to one minute
func SetCookieOneMinute(w http.ResponseWriter, name, value string, opts ...CookieOpt) {
	opts = append([]CookieOpt{Duration(time.Second * 60)}, opts...)
	SetCookie(w, name, value, opts...)
}

// DeleteCookie will delete a cookie of the specified name from the client
// browser.
func DeleteCookie(w http.ResponseWriter, name string) {
	SetCookie(w, name, "", Duration(time.Second*-1))
}

// Encoder encapsulates the secret key that will be used to securely sign the
// cookies set by EncodeVariableInCookie/ DecodeVariableFromCookie. If you are
// just setting plain cookies with SetCookie/SetCookieOneMinute, you do not
// need this.
type Encoder struct {
	Key string
}

// NewEncoder returns a new Encoder.
func NewEncoder(key string) Encoder {
	return Encoder{Key: key}
}

// EncodeVariableInCookie will serialize any go variable into a string and
// store it in a client side cookie, where it can later be retrieved from the
// cookie with DecodeVariableFromCookie.
func (ce Encoder) EncodeVariableInCookie(w http.ResponseWriter, cookiename string, variable interface{}, opts ...CookieOpt) error {
	value, err := auth.Serialize(ce.Key, variable)
	if err != nil {
		return erro.Wrap(err)
	}
	SetCookie(w, cookiename, value, opts...)
	return nil
}

// DecodeVariableFromCookie will retrieve a go variable from a client side
// cookie that was previously set by EncodeVariableInCookie. If the named
// cookie does not exist, it will fail with a ErrNoCookie error. Note that you
// need to pass a pointer to the variable you wish to decode the cookie value
// into, not the variable itself.
//
// The cookie's digital hash signature will be checked for any tampering. If
// the signature is wrong, DecodeVariableFromCookie will fail with an
// auth.ErrDeserializeOutputInvalid error.
func (ce Encoder) DecodeVariableFromCookie(r *http.Request, cookiename string, variablePtr interface{}) error {
	var cookie *http.Cookie
	cookie, err := r.Cookie(cookiename)
	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			return erro.Wrap(erro.Errorf(ErrNoCookie, cookiename))
		default:
			return erro.Wrap(err)
		}
	}
	err = auth.Deserialize(ce.Key, cookie.Value, variablePtr)
	if err != nil {
		return erro.Wrap(err)
	}
	return nil
}
