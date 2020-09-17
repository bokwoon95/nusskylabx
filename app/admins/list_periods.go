package admins

import (
	"fmt"
	"log"
	"net/http"

	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/app/skylab"
	"github.com/bokwoon95/nusskylabx/tables"

	"github.com/bokwoon95/nusskylabx/helpers/erro"
	"github.com/bokwoon95/nusskylabx/helpers/flash"
	"github.com/bokwoon95/nusskylabx/helpers/formutil"
	"github.com/bokwoon95/nusskylabx/helpers/headers"
	"github.com/bokwoon95/nusskylabx/helpers/timeutil"
	"github.com/bokwoon95/nusskylabx/helpers/urlparams"
)

func (adm Admins) ListPeriods(w http.ResponseWriter, r *http.Request) {
	adm.skylb.Log.TraceRequest(r)
	r = adm.skylb.SetRoleSection(w, r, skylab.RoleAdmin, skylab.AdminListPeriods)
	headers.DoNotCache(w)

	// Get the last valid cohort
	cohort, _ := urlparams.PersistentString(w, r, "cohort", "_admin_list_periods_cohort")
	if cohort == "" || !skylab.Contains(adm.skylb.Cohorts(), cohort) {
		http.Redirect(w, r, skylab.AdminListPeriods+"/"+adm.skylb.CurrentCohort(), http.StatusMovedPermanently)
		return
	}

	type Data struct {
		Periods []skylab.Period
		Cohort  string
	}
	var data Data
	data.Cohort = cohort
	p := tables.PERIODS()
	me, se := tables.MILESTONE_ENUM(), tables.STAGE_ENUM()
	var period skylab.Period
	err := sq.
		WithLog(adm.skylb.Log, sq.Lverbose).
		From(p).
		Where(p.COHORT.EqString(cohort)).
		OrderBy(
			sq.Fieldf("array_position((SELECT array_agg(?) FROM ?), ?)", me.MILESTONE, me, p.MILESTONE),
			sq.Fieldf("array_position((SELECT array_agg(?) FROM ?), ?)", se.STAGE, se, p.STAGE),
		).
		Selectx(func(row *sq.Row) {
			period.Valid = row.IntValid(p.PERIOD_ID)
			period.PeriodID = row.Int(p.PERIOD_ID)
			period.Cohort = row.String(p.COHORT)
			period.Stage = row.String(p.STAGE)
			period.Milestone = row.String(p.MILESTONE)
			period.StartAt = row.NullTime(p.START_AT)
			period.EndAt = row.NullTime(p.END_AT)
		}, func() {
			data.Periods = append(data.Periods, period)
		}).
		Fetch(adm.skylb.DB)
	if err != nil {
		adm.skylb.InternalServerError(w, r, err)
		return
	}
	funcs := adm.skylb.AddInputSelects(nil)
	funcs = timeutil.Funcs(funcs)
	adm.skylb.Render(w, r, data, funcs, "app/admins/list_periods.html")
}

func (adm Admins) ListPeriodsDelete(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = formutil.ParseForm(r)
		msgs := make(map[string][]string)
		periodIDs := r.Form["periodID"]
		p := tables.PERIODS()
		rowsAffected, err := sq.WithLog(adm.skylb.Log, sq.Lverbose).
			DeleteFrom(p).
			Where(p.PERIOD_ID.In(periodIDs)).
			Exec(adm.skylb.DB, sq.ErowsAffected)
		if err != nil {
			if pqerr, ok := erro.AsPqError(err); ok && pqerr.Code == erro.PqForeignKeyViolation {
				msgs[flash.Error] = append(msgs[flash.Error],
					fmt.Sprintf("Unable to delete periods %+v because some are still referenced in the database", periodIDs),
				)
			} else {
				log.Printf("[ERROR] deleting from period with periodIDs %+v: %s\n", periodIDs, err)
				msgs[flash.Error] = append(msgs[flash.Error], fmt.Sprintf("Unable to delete periods %+v: %s", periodIDs, err))
			}
		}
		if rowsAffected > 0 {
			msgs[flash.Success] = append(msgs[flash.Success], fmt.Sprintf("%d period(s) deleted", rowsAffected))
		}
		r, _ = adm.skylb.SetFlashMsgs(w, r, msgs)
		next.ServeHTTP(w, r)
	})
}

func (adm Admins) ListPeriodsCreate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = formutil.ParseForm(r)
		msgs := make(map[string][]string)
		cohort := r.FormValue("cohort")
		stage := r.FormValue("stage")
		milestone := r.FormValue("milestone")
		startdate := r.FormValue("startdate")
		starttime := r.FormValue("starttime")
		enddate := r.FormValue("enddate")
		endtime := r.FormValue("endtime")
		start := timeutil.ParseDateTimeString(startdate, starttime)
		end := timeutil.ParseDateTimeString(enddate, endtime)
		p := tables.PERIODS()
		var count int
		stmt := sq.
			InsertInto(p).
			Columns(p.COHORT, p.STAGE, p.MILESTONE, p.START_AT, p.END_AT).
			Values(cohort, stage, milestone, start, end).
			OnConflict().DoNothing().ReturningOne().
			CTE("stmt")
		err := sq.WithLog(adm.skylb.Log, sq.Lverbose).
			From(stmt).
			SelectRowx(func(row *sq.Row) {
				count = row.Int(sq.Count())
			}).
			Fetch(adm.skylb.DB)
		if err != nil {
			msgs[flash.Error] = append(msgs[flash.Success], erro.Wrap(err).Error())
		}
		if count == 0 {
			msgs[flash.Warning] = append(msgs[flash.Warning], fmt.Sprintf(
				"Period with cohort: %s, stage: %s, milestone: %s not created as it already exists",
				cohort, stage, milestone,
			))
		}
		r = urlparams.SetString(r, "cohort", cohort)
		r, _ = adm.skylb.SetFlashMsgs(w, r, msgs)
		next.ServeHTTP(w, r)
	})
}

func (adm Admins) ListPeriodsDuplicate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = formutil.ParseForm(r)
		msgs := make(map[string][]string)
		nextHandler := func(w http.ResponseWriter, r *http.Request, msgs map[string][]string) {
			r, _ = adm.skylb.SetFlashMsgs(w, r, msgs)
			next.ServeHTTP(w, r)
		}
		cohort := r.FormValue("cohort")
		if cohort == "" {
			msgs[flash.Error] = append(msgs[flash.Error], "cohort cannot be blank")
			nextHandler(w, r, msgs)
			return
		}
		if !skylab.Contains(adm.skylb.Cohorts(), cohort) {
			msgs[flash.Error] = append(msgs[flash.Error], "Invalid cohort: "+cohort)
			nextHandler(w, r, msgs)
			return
		}
		r = urlparams.SetString(r, "cohort", cohort)
		periodIDs := r.Form["periodID"]
		p1, p2, p3 := tables.PERIODS().As("p1"), tables.PERIODS().As("p2"), tables.PERIODS().As("p3")
		rowsAffected, err := sq.WithLog(adm.skylb.Log, sq.Lverbose).
			InsertInto(p1).
			Columns(p1.COHORT, p1.STAGE, p1.MILESTONE, p1.START_AT, p1.END_AT).
			Select(
				sq.Select(
					sq.String(cohort),
					p2.STAGE,
					p2.MILESTONE,
					sq.Fieldf(
						"? + MAKE_INTERVAL(YEARS := ?::INT - EXTRACT(YEAR FROM ?)::INT)",
						p2.START_AT, cohort, p2.START_AT,
					),
					sq.Fieldf(
						"? + MAKE_INTERVAL(YEARS := ?::INT - EXTRACT(YEAR FROM ?)::INT)",
						p2.END_AT, cohort, p2.END_AT,
					)).
					From(p2).
					Where(
						p2.PERIOD_ID.In(r.Form["periodID"]),
						sq.Not(sq.Exists(
							sq.SelectOne().From(p3).Where(
								p3.COHORT.EqString(cohort),
								p3.STAGE.Eq(p2.STAGE),
								p3.MILESTONE.Eq(p2.MILESTONE),
							),
						)),
					),
			).
			OnConflict().
			DoNothing().
			Exec(adm.skylb.DB, sq.ErowsAffected)
		if err != nil {
			msgs[flash.Error] = append(msgs[flash.Error], erro.Wrap(err).Error())
			nextHandler(w, r, msgs)
			return
		}
		if rowsAffected > 0 {
			msgs[flash.Success] = append(msgs[flash.Success], fmt.Sprintf("%d period(s) duplicated", rowsAffected))
		}
		if diff := len(periodIDs) - int(rowsAffected); diff > 0 {
			msgs[flash.Warning] = append(msgs[flash.Warning], fmt.Sprintf("%d period(s) not duplicated because they already exist", diff))
		}
		nextHandler(w, r, msgs)
	})
}
