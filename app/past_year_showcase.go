package app

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/bokwoon95/nusskylabx/helpers/urlparams"

	"github.com/bokwoon95/nusskylabx/app/skylab"
	"github.com/bokwoon95/nusskylabx/helpers/headers"
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

	data := make(map[string]interface{})
	data["Cohort"] = cohort
	data["ProjectLevel"] = projectlevel
	data["Numbers"] = make([]int, 100)
	rand.Seed(time.Now().UnixNano())
	data["RandomInt"] = func() int { return rand.Intn(100000) }
	ap.skylb.Wender(w, r, "app/past_year_showcase.html", data)
}
