package admins

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/bokwoon95/nusskylabx/app/skylab"

	"github.com/bokwoon95/nusskylabx/helpers/dbutil"
	"github.com/bokwoon95/nusskylabx/helpers/flash"
	"github.com/bokwoon95/nusskylabx/helpers/formutil"
	"github.com/bokwoon95/nusskylabx/helpers/urlparams"
	"github.com/davecgh/go-spew/spew"
)

func (adm Admins) ListCohorts(w http.ResponseWriter, r *http.Request) {
	adm.skylb.Log.TraceRequest(r)
	r = adm.skylb.SetRoleSection(w, r, skylab.RoleAdmin, skylab.AdminListCohorts)
	var msgs = make(map[string][]string)
	var nextCohort string
	latestCohort, _ := strconv.Atoi(adm.skylb.LatestCohort())
	if latestCohort != 0 {
		nextCohort = strconv.Itoa(latestCohort + 1)
	}
	if nextCohort == "" {
		msgs[flash.Error] = []string{"Could not compute the next cohort after " + adm.skylb.LatestCohort()}
	}
	r, _ = adm.skylb.SetFlashMsgs(w, r, msgs)
	data := map[string]interface{}{
		"NextCohort": nextCohort,
		"Cohorts":    adm.skylb.Cohorts(),
	}
	adm.skylb.Render(w, r, data, "app/admins/list_cohorts.html")
}

func (adm Admins) ListCohortsCreate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		adm.skylb.Log.TraceRequest(r)
		msgs := make(map[string][]string)
		cohort, err := urlparams.String(r, "cohort")
		if err != nil {
			adm.skylb.BadRequest(w, r, err.Error())
		}
		query := `INSERT INTO cohort_enum (cohort) VALUES ($1)`
		_, err = adm.skylb.DB.Exec(query, cohort)
		if err != nil {
			dberr := dbutil.NewDBError(err, query, cohort)
			msgs[flash.Error] = append(msgs[flash.Error], fmt.Sprintf("%s<br>%s", dberr.Query, dberr.Error()))
		} else {
			msgs[flash.Success] = append(msgs[flash.Success], "Created cohort "+cohort)
		}
		err = adm.skylb.RefreshCohorts()
		if err != nil {
			msgs[flash.Error] = append(msgs[flash.Error], err.Error())
		}
		r, _ = adm.skylb.SetFlashMsgs(w, r, msgs)
		next.ServeHTTP(w, r)
	})
}

func (adm Admins) ListCohortsDelete(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		adm.skylb.Log.TraceRequest(r)
		msgs := make(map[string][]string)
		_ = formutil.ParseForm(r)
		pass := func(w http.ResponseWriter, r *http.Request, msgs map[string][]string) {
			r, _ = adm.skylb.SetFlashMsgs(w, r, msgs)
			next.ServeHTTP(w, r)
		}
		// var placeholders []string // not sure what this is for, maybe delete
		var cohorts []interface{}
		for _, cohort := range r.Form["cohort"] {
			if skylab.Contains(adm.skylb.Cohorts(), cohort) {
				// placeholders = append(placeholders, "?")
				cohorts = append(cohorts, cohort)
			}
		}
		if len(cohorts) == 0 {
			msgs[flash.Warning] = append(msgs[flash.Warning], fmt.Sprintf("no cohorts were passed in for deletion<br><pre>%s</pre>", spew.Sdump(r.Form)))
			pass(w, r, msgs)
			return
		}
		query := `DELETE FROM cohort_enum WHERE cohort = $1`
		for _, cohort := range r.Form["cohort"] {
			_, err := adm.skylb.DB.Exec(query, cohort)
			if err != nil {
				adm.skylb.Log.Println(dbutil.AsDBError(err))
				if dberr, ok := dbutil.AsDBError(err); ok {
					switch dberr.Code {
					case dbutil.ErrForeignKeyViolation:
						msgs[flash.Error] = append(msgs[flash.Error], fmt.Sprintf("Cannot delete cohort %s as it is currently in use", cohort))
					default:
						msgs[flash.Error] = append(msgs[flash.Error], fmt.Sprintf("%s<br>%s", dberr.Query, dberr.Error()))
					}
				} else {
					msgs[flash.Error] = append(msgs[flash.Error], err.Error())
				}
			} else {
				msgs[flash.Success] = append(msgs[flash.Success], fmt.Sprintf("Deleted cohort %s", cohort))
			}
		}
		err := adm.skylb.RefreshCohorts()
		if err != nil {
			msgs[flash.Error] = append(msgs[flash.Error], err.Error())
		}
		pass(w, r, msgs)
	})
}

func (adm Admins) ListCohortsRefresh(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		adm.skylb.Log.TraceRequest(r)
		msgs := make(map[string][]string)
		err := adm.skylb.RefreshCohorts()
		if err != nil {
			msgs[flash.Error] = append(msgs[flash.Error], err.Error())
		} else {
			msgs[flash.Success] = append(msgs[flash.Success], "Cohorts Refreshed!")
		}
		r, _ = adm.skylb.SetFlashMsgs(w, r, msgs)
		next.ServeHTTP(w, r)
	})
}
