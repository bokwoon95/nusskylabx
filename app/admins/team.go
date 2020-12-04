package admins

import (
	"net/http"

	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/app/skylab"
	"github.com/bokwoon95/nusskylabx/helpers/urlparams"
	"github.com/bokwoon95/nusskylabx/tables"
)

func (adm Admins) TeamView(w http.ResponseWriter, r *http.Request) {
	adm.skylb.Log.TraceRequest(r)
	r = adm.skylb.SetRoleSection(w, r, skylab.RolePreserve, skylab.SectionPreserve)
	teamID, err := urlparams.Int(r, "teamID")
	if err != nil {
		adm.skylb.BadRequest(w, r, err.Error())
		return
	}
	t := tables.V_TEAMS()
	var team skylab.Team
	err = sq.WithDefaultLog(sq.Lverbose).
		From(t).
		Where(t.TEAM_ID.EqInt(teamID)).
		SelectRowx(team.RowMapper(t)).
		Fetch(adm.skylb.DB)
	if err != nil {
		adm.skylb.InternalServerError(w, r, err)
		return
	}
	data := map[string]interface{}{
		"Team":        team,
		"UserBaseURL": skylab.AdminUser,
	}
	adm.skylb.Wender(w, r, data, "app/skylab/team_view.html")
}
