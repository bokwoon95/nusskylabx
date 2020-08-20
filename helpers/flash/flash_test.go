package flash

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bokwoon95/nusskylabx/helpers/random"

	"github.com/bokwoon95/nusskylabx/helpers/testutil"
	"github.com/matryer/is"
)

func TestSetGetFlashMsgs(t *testing.T) {
	is := is.New(t)
	fe := NewEncoder(random.Word())
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "", nil)

	// Set FlashMsgs
	msgs := map[string][]string{
		Success: {
			"Successfully Updated Name",
			"Successfully Updated Email",
			"Successfully Updated Phone Number",
		},
		Error: {
			"User 1 Already Exists",
			"Display Name Already taken",
			"Password cannot be blank",
		},
	}
	_, err := fe.SetFlashMsgs(w, r, msgs)
	is.NoErr(err)

	// Get FlashMsgs
	r = testutil.ConvertResponseToRequest(t, w)
	flashmsgsMap, err := fe.GetFlashMsgs(w, r)
	is.NoErr(err)
	is.Equal(flashmsgsMap, appendFlashMsgs(nil, msgs))
}

func TestUnsetFlashMsgsHandler(t *testing.T) {
	is := is.New(t)
	fe := NewEncoder(random.Word())
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "", nil)

	// Set FlashMsgs
	msgs := map[string][]string{
		Success: {"This message should be preserved"},
		Error:   {"Delete this message"},
		Warning: {"Also delete this message"},
	}
	r, _ = fe.SetFlashMsgs(w, r, msgs)

	// testing handler
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		delete(msgs, Error)
		delete(msgs, Warning)
		flashmsgs, err := fe.GetFlashMsgs(w, r)
		is.NoErr(err)
		is.Equal(flashmsgs, appendFlashMsgs(nil, msgs))
	})

	fe.UnsetFlashMsgsHandler(Error, Warning)(next).ServeHTTP(w, r)
}
