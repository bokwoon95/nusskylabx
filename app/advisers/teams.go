package advisers

import (
	"net/http"

	"github.com/bokwoon95/nusskylabx/app/skylab"
)

func (adv Advisers) Teams(w http.ResponseWriter, r *http.Request) {
	adv.skylb.Log.TraceRequest(r)
	r = adv.skylb.SetRoleSection(w, r, skylab.RoleAdviser, skylab.AdviserTeams)
	adv.skylb.Render(w, r, nil, "app/advisers/teams.html")
}
