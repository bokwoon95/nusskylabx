package admins

import (
	"net/http"
	"strings"

	"github.com/bokwoon95/nusskylabx/app/skylab"

	"github.com/bokwoon95/nusskylabx/helpers/headers"
)

func (adm Admins) CreateUser(w http.ResponseWriter, r *http.Request) {
	adm.skylb.Log.TraceRequest(r)
	r = adm.skylb.SetRoleSection(w, r, skylab.RoleAdmin, skylab.AdminCreateUser)
	headers.DoNotCache(w)
	type Data struct {
		Rows string
	}
	var data Data
	var rows []string
	err := adm.skylb.DecodeVariableFromCookie(r, createUserRowsCookie, &rows)
	if err == nil {
		data.Rows = strings.Join(rows, "\n")
	}
	adm.skylb.Render(w, r, data, nil, "app/admins/create_user.html")
}
