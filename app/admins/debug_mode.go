package admins

import (
	"io/ioutil"
	"net/http"
	"os"

	"github.com/bokwoon95/nusskylabx/app/skylab"

	"github.com/bokwoon95/nusskylabx/helpers/headers"
)

func (adm Admins) DebugMode(w http.ResponseWriter, r *http.Request) {
	adm.skylb.Log.TraceRequest(r)
	r = adm.skylb.SetRoleSection(w, r, skylab.RoleAdmin, skylab.AdminDebugMode)
	headers.DoNotCache(w)
	type Data struct {
		DebugMode bool
	}
	var data Data
	data.DebugMode = adm.skylb.Log.Writer() != ioutil.Discard
	adm.skylb.Render(w, r, data, nil, "app/admins/debug_mode.html")
}

func (adm Admins) DebugOn(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		adm.skylb.Log.TraceRequest(r)
		adm.skylb.Log.SetOutput(os.Stdout)
		next.ServeHTTP(w, r)
	})
}

func (adm Admins) DebugOff(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		adm.skylb.Log.TraceRequest(r)
		adm.skylb.Log.SetOutput(ioutil.Discard)
		next.ServeHTTP(w, r)
	})
}
