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
	var rows []string
	err := adm.skylb.DecodeVariableFromCookie(r, createUserRowsCookie, &rows)
	if err != nil {
		adm.skylb.InternalServerError(w, r, err)
		return
	}
	data := map[string]interface{}{
		"Rows": strings.Join(rows, "\n"),
	}
	adm.skylb.Render(w, r, data, "app/admins/create_user.html")
}
