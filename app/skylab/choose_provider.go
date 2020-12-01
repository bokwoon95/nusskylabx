package skylab

import (
	"net/http"
	"strings"

	"github.com/bokwoon95/nusskylabx/helpers/auth"
)

// ChooseProvider will check if there is a valid authentication provider
// (either openid or oauth) in the current request's queryparams and if there
// isn't, it will render a page for the user to choose their preferred
// provider. The page will contain the same links that it was called with,
// effectively redirecting to itself (except this time with a valid provider)
func (skylb Skylab) ChooseProvider(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		skylb.Log.TraceRequest(r)
		magicstring := strings.TrimSpace(r.FormValue("magicstring"))
		provider := r.FormValue("provider")
		if auth.IsValidProvider(provider) {
			next.ServeHTTP(w, r)
			return
		}
		type Data struct {
			Magicstring string
			RequestURL  string
		}
		data := Data{
			Magicstring: magicstring,
			RequestURL:  r.RequestURI,
		}
		skylb.Render(w, r, data, auth.AddConstProvider(map[string]interface{}{}), "app/skylab/choose_provider.html")
	})
}
