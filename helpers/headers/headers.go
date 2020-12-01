// Package headers provides utilities for setting various headers
package headers

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"github.com/bokwoon95/nusskylabx/helpers/random"
)

func Funcs(funcs map[string]interface{}, r *http.Request) map[string]interface{} {
	funcs["HeadersCSPNonce"] = func() string {
		nonce, _ := r.Context().Value(CspNonceCtxkey).(string)
		return nonce
	}
	return funcs
}

// DoNotCache tells the browser to never cache the page being rendered.
// Performance will take a hit because the server must serve more requests, but
// sometimes it is vital that the data on a page stays fresh. Examples include
// the browser almost always displaying a cached result of the page when a user
// presses the back button, potentially displaying to the user stale data.
// Using DoNotCache will force the browser to request the server for a new page
// everytime, never caching it.
//
// TL;DR use this for data-sensitive pages where data updates regularly and
// it is not acceptable to show outdated information.
func DoNotCache(w http.ResponseWriter) {
	// https://stackoverflow.com/a/2068407
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
	w.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
	w.Header().Set("Expires", "0")                                         // Proxies.
}

type cspNonceCtxKey string

const CspNonceCtxkey cspNonceCtxKey = "CspNonceCtxkey"

// SecurityHeaders sets headers related to security to every outgoing response.
// The most relevant policy you might be interested in is
// Content-Security-Policy, which whitelists URLs that the site is able to
// access js/css/images from. This means all CDNs that we are using (e.g.
// bootstrap, jquery, javascript/css libraries) must be explicitly mentioned
// here, or they will be blocked by the browser.
//
// For more information about security related headers take a look at
// https://scotthelme.co.uk/introducing-securityheaders-io/
func SecurityHeaders(w http.ResponseWriter, r *http.Request) *http.Request {
	// We generate a nonce to use for our inline scripts. It will be injected
	// into the context for a template function to pick up later.
	// https://content-security-policy.com/examples/allow-inline-script/
	nonce := random.SecretKey()
	r = r.WithContext(context.WithValue(r.Context(), CspNonceCtxkey, nonce))

	// Any new URLs you wish to access should be added to the securityPolicies
	// whitelist below. You may need to do some reading up on which is the
	// relevant security policy to place the URL under (or you can just follow
	// what the error in the browser console tells you).
	//
	// Unfortunately the setting of 'unsafe-inline' is preventing us from
	// getting a A+ on https://securityheaders.com but is needed for several of
	// the external libraries and CDNs to work
	securityPolicies := []string{
		`script-src-elem
			'self'
			'unsafe-inline'
			'nonce-` + nonce + `'
			cdn.jsdelivr.net
			stackpath.bootstrapcdn.com
			cdn.datatables.net
			unpkg.com
			code.jquery.com
		`,
		`style-src-elem
			'self'
			cdn.jsdelivr.net
			stackpath.bootstrapcdn.com
			cdn.datatables.net
			unpkg.com
			fonts.googleapis.com
		`,
		`style-src 'unsafe-inline'`,
		`img-src
			'self'
			'unsafe-inline'
			cdn.datatables.net
			data:
			source.unsplash.com
			images.unsplash.com
		`,
		`font-src fonts.gstatic.com`,
		"default-src 'self'",
		"object-src 'self'",
		"media-src 'self'",
		"frame-ancestors 'self'",
		"connect-src 'self'",
	}
	// The multiline strings have ugly tabs and spaces in them, condense all
	// multi-whitespaces into a single whitespace so that the resultant header
	// string looks more presentable in the browser devtools.
	//
	// Every policy should also be joined together into a string, delimited by
	// a ; semicolon
	ContentSecurityPolicy := regexp.MustCompile(`\s+`).ReplaceAllString(strings.Join(securityPolicies, "; "), " ")
	w.Header().Set("Content-Security-Policy", ContentSecurityPolicy)

	// We're not going to use any of these features, so we can tell the browser
	// that.
	features := []string{
		`vibrate 'none'`,
		`microphone 'none'`,
		`camera 'none'`,
		`magnetometer 'none'`,
		`gyroscope 'none'`,
	}
	FeaturePolicy := regexp.MustCompile(`\s+`).ReplaceAllString(strings.Join(features, "; "), " ")
	w.Header().Set("Feature-Policy", FeaturePolicy)
	w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
	w.Header().Set("Referrer-Policy", "strict-origin")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "sameorigin")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	return r
}

// SecurityHeadersHandler is a http.Handler middleware wrapper around
// SecurityHeaders.
func SecurityHeadersHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = SecurityHeaders(w, r)
		next.ServeHTTP(w, r)
	})
}
