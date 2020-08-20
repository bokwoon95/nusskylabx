package app

import (
	"html/template"
	"math/rand"
	"net/http"
	"time"

	"github.com/bokwoon95/nusskylabx/helpers/urlparams"

	"github.com/bokwoon95/nusskylabx/app/skylab"
	"github.com/bokwoon95/nusskylabx/helpers/headers"
	"github.com/bokwoon95/nusskylabx/helpers/random"
)

func (ap App) PastYearShowcase(w http.ResponseWriter, r *http.Request) {
	ap.skylb.Log.TraceRequest(r)
	r = ap.skylb.SetRoleSection(w, r, skylab.RolePreserve, "")
	headers.DoNotCache(w)

	// Get the last valid cohort
	cohort, _ := urlparams.PersistentString(w, r, "cohort", "_past_year_showcase_cohort")
	if cohort == "" || !skylab.Contains(ap.skylb.Cohorts(), cohort) {
		http.Redirect(w, r, "/showcase/"+ap.skylb.CurrentCohort(), http.StatusMovedPermanently)
		return
	}

	// Get the last valid projectlevel
	projectlevel, _ := urlparams.PersistentString(w, r, "projectlevel", "_past_year_showcase_project_level")
	if !skylab.Contains(skylab.ProjectLevels(), projectlevel) {
		http.Redirect(w, r, "/showcase/"+cohort+"/"+skylab.ProjectLevelApollo, http.StatusMovedPermanently)
		return
	}

	type Data struct {
		Cohort       string
		ProjectLevel string
		Numbers      []int
	}
	var data Data
	data.Cohort = cohort
	data.ProjectLevel = projectlevel
	data.Numbers = make([]int, 100)
	rand.Seed(time.Now().UnixNano())
	funcs := template.FuncMap{
		"RandomTeamName": random.TeamName,
		"RandomInt":      func() int { return rand.Intn(100000) },
	}
	ap.skylb.Render(w, r, data, funcs, "app/past_year_showcase.html")
}
