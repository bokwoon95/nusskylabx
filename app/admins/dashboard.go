package admins

import (
	"net/http"

	"github.com/bokwoon95/nusskylabx/app/skylab"
)

func (adm Admins) Dashboard(w http.ResponseWriter, r *http.Request) {
	adm.skylb.Log.TraceRequest(r)
	r = adm.skylb.SetRoleSection(w, r, skylab.RoleAdmin, skylab.AdminDashboard)
	adm.skylb.Wender(w, r, nil, "app/admins/dashboard.html")
}
