package cookies

import (
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/bokwoon95/nusskylabx/helpers/random"

	"github.com/bokwoon95/nusskylabx/helpers/testutil"
	"github.com/google/uuid"
	"github.com/matryer/is"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func TestSetCookie(t *testing.T) {
	is := is.New(t)
	var w *httptest.ResponseRecorder
	var name, value string
	var n int
	var cookies []*http.Cookie

	w, name, value = httptest.NewRecorder(), uuid.New().String(), uuid.New().String()
	SetCookie(w, name, value, AllowJS(true), MustHTTPS(false))
	cookies = w.Result().Cookies()
	is.True(len(cookies) > 0)
	is.Equal(cookies[0].Name, name)   // Name matches
	is.Equal(cookies[0].Value, value) // Value matches

	w, name, value = httptest.NewRecorder(), uuid.New().String(), uuid.New().String()
	n = rand.Intn(100)
	SetCookie(w, name, value, Duration(time.Second*time.Duration(n)))
	cookies = w.Result().Cookies()
	is.True(len(cookies) > 0)
	is.Equal(cookies[0].Name, name)   // Name matches
	is.Equal(cookies[0].Value, value) // Value matches
	is.Equal(cookies[0].MaxAge, n)    // MaxAge matches n

	w, name, value = httptest.NewRecorder(), uuid.New().String(), uuid.New().String()
	SetCookieOneMinute(w, name, value)
	cookies = w.Result().Cookies()
	is.Equal(cookies[0].Name, name)   // Name matches
	is.Equal(cookies[0].Value, value) // Value matches
	is.Equal(cookies[0].MaxAge, 60)   // MaxAge matches one minute
}

func TestGetCookieValue(t *testing.T) {
	tests := []struct {
		DontSet bool
		Name    string
		Value   string
	}{
		{false, uuid.New().String(), uuid.New().String()},
		{false, uuid.New().String(), uuid.New().String()},
		{true, "", ""},
		{true, uuid.New().String(), uuid.New().String()},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(strconv.Itoa(i+1), func(t *testing.T) {
			t.Parallel()
			is := is.New(t)
			r, _ := http.NewRequest("GET", "", nil)
			if !tt.DontSet {
				r.AddCookie(&http.Cookie{Name: tt.Name, Value: tt.Value})
				is.Equal(GetCookieValue(r, tt.Name), tt.Value)
			} else {
				is.Equal(GetCookieValue(r, tt.Name), "")
			}
		})
	}
}

func TestDeleteCookie(t *testing.T) {
	is := is.New(t)
	w, name := httptest.NewRecorder(), uuid.New().String()
	DeleteCookie(w, name)
	cookies := w.Result().Cookies()
	is.True(len(cookies) > 0)
	is.Equal(cookies[0].Name, name) // Name matches
	is.Equal(cookies[0].Value, "")  // Value is empty
	is.True(cookies[0].MaxAge < 1)  // MaxAge expired
}

func TestEncodeDecodeCookie(t *testing.T) {
	is := is.New(t)
	type Person struct {
		Name       string
		Age        int
		Occupation string
	}
	ce := NewEncoder(uuid.New().String())

	// Encode a person into a cookie
	w, name := httptest.NewRecorder(), uuid.New().String()
	input := Person{"Bob", 25, "Cryptographer"}
	err := ce.EncodeVariableInCookie(w, name, input)
	is.NoErr(err)

	// Decode a person from the cookie value
	r := testutil.ConvertResponseToRequest(t, w)
	var output Person
	err = ce.DecodeVariableFromCookie(r, name, &output)
	is.NoErr(err)
	is.Equal(input, output)
}

func TestRequestAddCookie(t *testing.T) {
	type args struct {
		r     *http.Request
		name  string
		value string
		opts  []CookieOpt
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{"1", args{
			testutil.NewRequest("", nil),
			random.Word(),
			random.Word(),
			[]CookieOpt{AllowJS(true)},
		}},
		{"2", args{
			testutil.NewRequest("", nil),
			random.Word(),
			random.Word(),
			[]CookieOpt{MustHTTPS(false), Duration(time.Second * 10)},
		}},
		{"3",
			args{
				testutil.NewRequest("", nil),
				random.Word(),
				random.Word(),
				nil,
			}},
	}
	for _, tt := range tests {
		is := is.New(t)
		RequestAddCookie(tt.args.r, tt.args.name, tt.args.value, tt.args.opts...)
		cookie, err := tt.args.r.Cookie(tt.args.name)
		is.NoErr(err)
		is.True(cookie != nil)
		is.Equal(cookie.Value, tt.args.value)
	}
}
