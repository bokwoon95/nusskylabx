package skylab

import (
	"crypto/sha256"
	"net/http"
	"testing"

	"github.com/bokwoon95/nusskylabx/helpers/random"
	"github.com/bokwoon95/nusskylabx/helpers/testutil"
	"github.com/gorilla/csrf"
	"github.com/matryer/is"
)

func TestSkylab_BadRequest(t *testing.T) {
	t.Parallel()
	is := is.New(t)
	skylb := NewTestDefault(t)
	path := random.URL()
	skylb.Mux.Get(path, func(w http.ResponseWriter, r *http.Request) {
		skylb.BadRequest(w, r, random.Sentence(10))
	})
	w, r := testutil.NewGet(path, nil)
	skylb.Mux.ServeHTTP(w, r)
	is.True(testutil.HasBody(w))
	is.Equal(w.Code, http.StatusBadRequest)
}

func TestSkylab_NotLoggedIn(t *testing.T) {
	t.Parallel()
	is := is.New(t)
	skylb := NewTestDefault(t)
	path := random.URL()
	skylb.Mux.Get(path, skylb.NotLoggedIn)
	w, r := testutil.NewGet(path, nil)
	skylb.Mux.ServeHTTP(w, r)
	is.True(testutil.HasBody(w))
	is.Equal(w.Code, http.StatusUnauthorized)
}

func TestSkylab_NotARole_X(t *testing.T) {
	t.Parallel()
	is := is.New(t)
	skylb := NewTestDefault(t)
	path := random.URL()

	tests := []struct {
		handler func(http.ResponseWriter, *http.Request)
	}{
		{skylb.NotAuthorized},
		{skylb.NotARole(RoleAdmin)},
		{skylb.NotAnApplicant},
		{skylb.NotAStudent},
		{skylb.NotAnAdviser},
		{skylb.NotAMentor},
		{skylb.NotAnAdmin},
	}
	for _, tt := range tests {
		skylb.Mux.Get(path, tt.handler)
		w, r := testutil.NewGet(path, nil)
		skylb.Mux.ServeHTTP(w, r)
		is.True(testutil.HasBody(w))
		is.Equal(w.Code, http.StatusForbidden)
	}
}

func TestSkylab_CsrfTokenInvalid(t *testing.T) {
	t.Parallel()
	is := is.New(t)
	skylb := NewTestDefault(t)
	authKey := sha256.Sum256([]byte(random.SecretKey()))
	skylb.Mux.Use(csrf.Protect(authKey[:])) // Add CSRF protection
	path := random.URL()
	skylb.Mux.Post(path, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("200 OK"))
	})
	// Send a POST without any CSRF token
	w, r := testutil.NewPost(path, nil)
	skylb.Mux.ServeHTTP(w, r)
	is.True(testutil.HasBody(w))
	is.Equal(w.Code, http.StatusForbidden)
}

func TestSkylab_NotFound(t *testing.T) {
	t.Parallel()
	is := is.New(t)
	skylb := NewTestDefault(t)
	path := random.URL()
	w, r := testutil.NewGet(path, nil)
	skylb.Mux.Get(path, skylb.NotFound)
	skylb.Mux.ServeHTTP(w, r)
	is.True(testutil.HasBody(w))
	is.Equal(w.Code, http.StatusNotFound)
}

func TestSkylab_MethodNotAllowed(t *testing.T) {
	t.Parallel()
	is := is.New(t)
	skylb := NewTestDefault(t)
	path := random.URL()
	skylb.Mux.Get(path, skylb.MethodNotAllowed)
	w, r := testutil.NewGet(path, nil)
	skylb.Mux.ServeHTTP(w, r)
	is.True(testutil.HasBody(w))
	is.Equal(w.Code, http.StatusMethodNotAllowed)
}
