package headers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bokwoon95/nusskylabx/helpers/testutil"
	"github.com/matryer/is"
)

func TestDoNotCache(t *testing.T) {
	w := httptest.NewRecorder()
	DoNotCache(w)
}

func TestSecurityHeaders(t *testing.T) {
	w, r := testutil.NewGet("", nil)
	SecurityHeaders(w, r)
}

func TestSecurityHeadersHandler(t *testing.T) {
	is := is.New(t)
	w, r := testutil.NewGet("", nil)
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello world"))
	})
	SecurityHeadersHandler(h).ServeHTTP(w, r)
	is.True(testutil.HasBody(w))
}
