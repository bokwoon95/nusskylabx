package admins

import (
	"net/http"

	"github.com/bokwoon95/nusskylabx/app/skylab"
)

func (adm Admins) ListFeedbacks(w http.ResponseWriter, r *http.Request) {
	adm.skylb.Log.TraceRequest(r)
	r = adm.skylb.SetRoleSection(w, r, skylab.RoleAdmin, skylab.AdminListFeedbacks)
	type Data struct {
	}
	var data Data
	adm.skylb.Render(w, r, data, nil, "app/admins/list_feedbacks.html")
}
