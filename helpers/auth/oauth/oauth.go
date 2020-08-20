// Package oauth is a wrapper around x/oauth2 to provide oauth2 authetication for various providers
package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/bokwoon95/nusskylabx/helpers/headers"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"
)

const (
	ProviderGoogle   = "google"
	ProviderFacebook = "facebook"
)

func IsValidProvider(provider string) bool {
	providers := map[string]bool{
		ProviderGoogle:   true,
		ProviderFacebook: true,
	}
	return providers[provider]
}

func AddConstProvider(funcs template.FuncMap) template.FuncMap {
	if funcs == nil {
		funcs = template.FuncMap{}
	}
	funcs["ProviderGoogle"] = func() string { return ProviderGoogle }
	funcs["ProviderFacebook"] = func() string { return ProviderFacebook }
	return funcs
}

const csrftokenName = "_oauth_csrf"

type User struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Link          string `json:"link"`
	Picture       string `json:"picture"`
	Gender        string `json:"gender"`
	Locale        string `json:"locale"`
}

type State struct {
	Csrftoken string
	Provider  string
}

func generateRandomString() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func initializeOauthConfig(provider string, returnTo string) *oauth2.Config {
	switch provider {
	case ProviderGoogle:
		return &oauth2.Config{
			RedirectURL:  returnTo,
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.profile",
				"https://www.googleapis.com/auth/userinfo.email",
			},
			Endpoint: google.Endpoint,
		}
	case ProviderFacebook:
		return &oauth2.Config{
			RedirectURL:  returnTo,
			ClientID:     os.Getenv("FACEBOOK_CLIENT_ID"),
			ClientSecret: os.Getenv("FACEBOOK_CLIENT_SECRET"),
			Scopes: []string{
				"public_profile",
				"email",
			},
			Endpoint: facebook.Endpoint,
		}
	default:
		return &oauth2.Config{}
	}
}

func accessTokenEndpoint(provider, accesstoken string) string {
	accesstoken = url.QueryEscape(accesstoken)
	switch provider {
	case ProviderGoogle:
		return "https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accesstoken
	case ProviderFacebook:
		return "https://graph.facebook.com/me?access_token=" + accesstoken
	default:
		return ""
	}
}

func Redirect(provider, returnTo string, errorHandler func(http.ResponseWriter, *http.Request, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		headers.DoNotCache(w)
		if !IsValidProvider(provider) {
			errorHandler(w, r, fmt.Errorf("Invalid provider: %s", provider))
			return
		}
		csrftoken, err := generateRandomString()
		if err != nil {
			errorHandler(w, r, err)
			return
		}
		state, err := json.Marshal(State{
			Csrftoken: csrftoken,
			Provider:  provider,
		})
		if err != nil {
			errorHandler(w, r, err)
			return
		}
		oauthConfig := initializeOauthConfig(provider, returnTo)
		url := oauthConfig.AuthCodeURL(base64.URLEncoding.EncodeToString(state), oauth2.SetAuthURLParam("prompt", "consent"))
		http.SetCookie(w, &http.Cookie{
			Name:     csrftokenName,
			Value:    csrftoken,
			HttpOnly: true,
		})
		http.Redirect(w, r, url, http.StatusMovedPermanently)
	}
}

func Authenticate(returnTo string, errorHandler func(http.ResponseWriter, *http.Request, error)) func(http.Handler) http.Handler {
	if errorHandler == nil {
		errorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			panic("errorHandler cannot be nil")
		}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			statebytes, err := base64.URLEncoding.DecodeString(r.FormValue("state"))
			if err != nil {
				errorHandler(w, r, err)
				return
			}
			var state State
			err = json.Unmarshal(statebytes, &state)
			if err != nil {
				errorHandler(w, r, err)
				return
			}
			cookieCsrf, err := r.Cookie(csrftokenName)
			if err != nil {
				errorHandler(w, r, err)
				return
			}
			if state.Csrftoken != cookieCsrf.Value {
				errorHandler(w, r, fmt.Errorf("state.Csrftoken[%s] doesn't match cookieCsrf.Value[%s]", state.Csrftoken, cookieCsrf.Value))
				return
			}
			provider := state.Provider
			code := r.FormValue("code")
			oauthConfig := initializeOauthConfig(provider, returnTo)
			token, err := oauthConfig.Exchange(context.Background(), code)
			if err != nil {
				errorHandler(w, r, err)
				return
			}
			if !token.Valid() {
				errorHandler(w, r, fmt.Errorf("oauth token is invalid %+v", token))
				return
			}
			response, err := http.Get(accessTokenEndpoint(provider, token.AccessToken))
			if err != nil {
				errorHandler(w, r, err)
			}
			defer response.Body.Close()
			contents, err := ioutil.ReadAll(response.Body)
			if err != nil {
				errorHandler(w, r, err)
			}
			var user User
			err = json.Unmarshal(contents, &user)
			if err != nil {
				errorHandler(w, r, err)
			}
			ctx := r.Context()
			ctx = context.WithValue(ctx, "username", user.Email)
			ctx = context.WithValue(ctx, "displayname", user.Name)
			ctx = context.WithValue(ctx, "email", user.Email)
			// w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
