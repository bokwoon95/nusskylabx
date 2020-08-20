// Package testutil contains helper functions for writing tests
package testutil

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/joho/godotenv"
)

var (
	_, f, _, _ = runtime.Caller(0)
	// testutil.RootDir will point to wherever this project's root directory is located on the user's computer
	RootDir = filepath.Join(filepath.Dir(f), ".."+string(os.PathSeparator)+"..") + string(os.PathSeparator)
)

func init() {
	// Try to read .env; if that fails, i.e. in Github Actions, read from
	// .env.default instead
	if err := godotenv.Load(RootDir + ".env"); err != nil {
		if err := godotenv.Load(RootDir + ".env.default"); err != nil {
			panic("unable to source from either .env or .env.default")
		}
	}
}

func Getenv(key string) string {
	return os.Getenv(key)
}

func ConvertResponseToRequest(t *testing.T, w *httptest.ResponseRecorder) *http.Request {
	cookies := w.Result().Cookies()
	r, err := http.NewRequest("GET", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, cookie := range cookies {
		r.AddCookie(cookie)
	}
	return r
}

func NewGet(path string, values url.Values) (*httptest.ResponseRecorder, *http.Request) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", path, strings.NewReader(values.Encode()))
	return w, r
}

func NewPost(path string, values url.Values) (*httptest.ResponseRecorder, *http.Request) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", path, strings.NewReader(values.Encode()))
	return w, r
}

func NewRequest(path string, values url.Values) *http.Request {
	r, _ := http.NewRequest("GET", path, strings.NewReader(values.Encode()))
	return r
}

func ResponseOK(w *httptest.ResponseRecorder) bool {
	invalidCodes := []int{
		http.StatusBadRequest,          // 400
		http.StatusUnauthorized,        // 401
		http.StatusForbidden,           // 403
		http.StatusNotExtended,         // 404
		http.StatusMethodNotAllowed,    // 405
		http.StatusInternalServerError, // 500
	}
	for _, invalidCode := range invalidCodes {
		if w.Code == invalidCode {
			return false
		}
	}
	return HasBody(w)
}

func HasBody(w *httptest.ResponseRecorder) bool {
	body := strings.TrimSpace(w.Body.String())
	return body != ""
}
