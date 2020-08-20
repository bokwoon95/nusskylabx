// Package openid partially implements the OpenID 2.0 specification for the Relying Party (RP) in Stateless Mode
package openid

import (
	"context"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"

	"github.com/bokwoon95/nusskylabx/helpers/headers"
)

const (
	ProviderNUS = "nus"
)

var endpoints = map[string]string{
	ProviderNUS: "https://openid.nus.edu.sg/server/",
}

func IsValidProvider(provider string) bool {
	providers := make(map[string]bool)
	providers[ProviderNUS] = true
	return providers[provider]
}

func AddConstProvider(funcs template.FuncMap) template.FuncMap {
	if funcs == nil {
		funcs = template.FuncMap{}
	}
	funcs["ProviderNUS"] = func() string { return ProviderNUS }
	return funcs
}

func Redirect(provider, returnTo string, errorHandler func(http.ResponseWriter, *http.Request, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		headers.DoNotCache(w)
		magicstring := r.FormValue("magicstring")
		params := url.Values{}
		if magicstring != "" {
			params.Add("magicstring", magicstring)
		}

		// Construct OpenID url to redirect user to
		// https://openid.net/specs/openid-authentication-2_0.html#anchor27
		req, err := http.NewRequest("GET", endpoints[provider], nil)
		if err != nil {
			errorHandler(w, r, err)
		}
		q := req.URL.Query()
		q.Add("openid.ns", "http://specs.openid.net/auth/2.0")
		q.Add("openid.mode", "checkid_setup")
		q.Add("openid.claimed_id", "http://specs.openid.net/auth/2.0/identifier_select")
		q.Add("openid.identity", "http://specs.openid.net/auth/2.0/identifier_select")
		q.Add("openid.return_to", returnTo+"?"+params.Encode())
		q.Add("openid.sreg.required", "email,nickname,fullname")
		req.URL.RawQuery = q.Encode()
		url := req.URL.String()
		http.Redirect(w, r, url, http.StatusMovedPermanently)
	}
}

func Authenticate(provider string, errorHandler func(http.ResponseWriter, *http.Request, error)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// https://openid.net/specs/openid-authentication-2_0.html#responding_to_authentication
			// TODO: follow all four verification checks in
			// https://openid.net/specs/openid-authentication-2_0.html#verification
			// (Section 11)
			username := r.FormValue("openid.sreg.nickname")
			displayname := r.FormValue("openid.sreg.fullname")
			email := r.FormValue("openid.sreg.email")
			if username == "" || displayname == "" || email == "" {
				errorHandler(w, r, fmt.Errorf("Either nickname[%s], fullname[%s] or email[%s] is empty", username, displayname, email))
				return
			}
			// https://openid.net/specs/openid-authentication-2_0.html#verifying_signatures (Section 11.4.2)
			// Not storing any associations, we are verifying it directly with the OpenID Provider (OP)
			if provider != ProviderNUS {
				// OpenID 2.0 is obsolete, we will likely never implement it
				// for any other provider other than NUS OpenID. So we will
				// just throw an error if the provider is not NUS OpenID.
				errorHandler(w, r, fmt.Errorf("NUS is the only supported OpenID 2.0 provider for now"))
				return
			}
			queries := r.URL.Query()
			queries["openid.mode"] = []string{"check_authentication"}
			resp, err := http.PostForm(endpoints[provider], queries)
			if err != nil {
				errorHandler(w, r, err)
				return
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				errorHandler(w, r, err)
				return
			}
			// https://openid.net/specs/openid-authentication-2_0.html#verifying_signatures (Section 11.4.2.2)
			sbody := string(body)
			match := regexp.MustCompile(`is_valid:(\w+)`).FindStringSubmatch(sbody)
			if match == nil {
				errorHandler(w, r, fmt.Errorf("is_valid missing from OpenID response %s", sbody))
				return
			}
			if match[1] != "true" {
				errorHandler(w, r, fmt.Errorf("is_valid is not true from OpenID response %s", sbody))
				return
			}
			ctx := r.Context()
			ctx = context.WithValue(ctx, "username", username)
			ctx = context.WithValue(ctx, "displayname", displayname)
			ctx = context.WithValue(ctx, "email", email)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
