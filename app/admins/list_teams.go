package admins

import (
	"net/http"

	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/helpers/urlparams"
	"github.com/bokwoon95/nusskylabx/tables"

	"github.com/bokwoon95/nusskylabx/app/skylab"

	"github.com/bokwoon95/nusskylabx/helpers/headers"
)

func (adm Admins) ListTeams(w http.ResponseWriter, r *http.Request) {
	adm.skylb.Log.TraceRequest(r)
	r = adm.skylb.SetRoleSection(w, r, skylab.RoleAdmin, skylab.AdminListTeams)
	headers.DoNotCache(w)

	// Get the last valid cohort
	cohort, _ := urlparams.PersistentString(w, r, "cohort", "_admin_list_teams_cohort")
	if cohort == "" || !skylab.Contains(adm.skylb.Cohorts(), cohort) {
		http.Redirect(w, r, skylab.AdminListTeams+"/"+adm.skylb.CurrentCohort(), http.StatusMovedPermanently)
		return
	}

	t := tables.V_TEAMS()
	var team skylab.Team
	var teams []skylab.Team
	err := sq.WithDefaultLog(sq.Lverbose).
		From(t).
		Where(t.COHORT.EqString(cohort)).
		Selectx(team.RowMapper(t), func() { teams = append(teams, team) }).
		Fetch(adm.skylb.DB)
	if err != nil {
		adm.skylb.InternalServerError(w, r, err)
		return
	}
	data := map[string]interface{}{
		"Teams":  teams,
		"Cohort": cohort,
	}
	adm.skylb.Wender(w, r, data, "app/admins/list_teams.html")
}
