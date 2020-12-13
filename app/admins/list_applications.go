package admins

import (
	"net/http"

	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/helpers/urlparams"
	"github.com/bokwoon95/nusskylabx/tables"

	"github.com/bokwoon95/nusskylabx/app/skylab"
	"github.com/bokwoon95/nusskylabx/helpers/headers"
)

func (adm Admins) ListApplications(w http.ResponseWriter, r *http.Request) {
	adm.skylb.Log.TraceRequest(r)
	r = adm.skylb.SetRoleSection(w, r, skylab.RoleAdmin, skylab.AdminListApplications)
	headers.DoNotCache(w)

	// Get the last valid cohort
	cohort, _ := urlparams.PersistentString(w, r, "cohort", "_admin_list_applications_cohort")
	if cohort == "" || !skylab.Contains(adm.skylb.Cohorts(), cohort) {
		http.Redirect(w, r, skylab.AdminListApplications+"/"+adm.skylb.CurrentCohort(), http.StatusMovedPermanently)
		return
	}

	var msgs = make(map[string][]string)
	var application skylab.Application
	var applications []skylab.Application
	a := tables.V_APPLICATIONS()
	err := sq.WithDefaultLog(sq.Lverbose).
		From(a).
		Where(
			a.COHORT.EqString(cohort),
			a.APPLICANT1_USER_ID.IsNotNull(),
			a.APPLICANT2_USER_ID.IsNotNull(),
		).
		Selectx(func(row *sq.Row) {
			application = skylab.Application{
				Valid:         row.IntValid(a.APPLICATION_ID),
				ApplicationID: row.Int(a.APPLICATION_ID),
				Cohort:        row.String(a.COHORT),
				Status:        row.String(a.STATUS),
				ProjectLevel:  row.String(a.PROJECT_LEVEL),
				Magicstring:   row.NullString(a.MAGICSTRING),
				Submitted:     row.Bool(a.SUBMITTED),
				Applicant1: skylab.User{
					Valid:       row.IntValid(a.APPLICANT1_USER_ID),
					UserID:      row.Int(a.APPLICANT1_USER_ID),
					Displayname: row.String(a.APPLICANT1_DISPLAYNAME),
				},
				Applicant2: skylab.User{
					Valid:       row.IntValid(a.APPLICANT2_USER_ID),
					UserID:      row.Int(a.APPLICANT2_USER_ID),
					Displayname: row.String(a.APPLICANT2_DISPLAYNAME),
				},
			}
		}, func() {
			applications = append(applications, application)
		}).
		Fetch(adm.skylb.DB)
	if err != nil {
		adm.skylb.InternalServerError(w, r, err)
		return
	}
	r, _ = adm.skylb.SetFlashMsgs(w, r, msgs)
	data := map[string]interface{}{
		"Applications": applications,
		"Cohort":       cohort,
	}
	adm.skylb.Render(w, r, data, "app/admins/list_applications.html")
}
