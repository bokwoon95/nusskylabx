package skylab

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/bokwoon95/nusskylabx/helpers/random"
	"github.com/bokwoon95/nusskylabx/helpers/testutil"
	"github.com/matryer/is"
)

func TestSkylab_InternalServerError(t *testing.T) {
	t.Parallel()
	is := is.New(t)
	skylb := NewTestDefault(t)
	path := random.URL()
	skylb.Mux.Get(path, func(w http.ResponseWriter, r *http.Request) {
		skylb.InternalServerError(w, r, fmt.Errorf("some error"))
	})
	w, r := testutil.NewGet(path, nil)
	skylb.Mux.ServeHTTP(w, r)
	is.True(testutil.HasBody(w))
	is.Equal(w.Code, http.StatusInternalServerError)
}
