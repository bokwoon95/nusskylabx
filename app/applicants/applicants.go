package applicants

import (
	"database/sql"
	"errors"
	"net/http"

	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/helpers/timeutil"
	"github.com/bokwoon95/nusskylabx/tables"

	"github.com/bokwoon95/nusskylabx/app/db"
	"github.com/bokwoon95/nusskylabx/app/skylab"

	"github.com/bokwoon95/nusskylabx/helpers/headers"
)

type Applicants struct {
	skylb skylab.Skylab
	d     db.DB
}

func New(skylb skylab.Skylab) Applicants {
	return Applicants{
		skylb: skylb,
		d:     db.New(skylb),
	}
}

func (apt Applicants) CheckIfOpen(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apt.skylb.Log.TraceRequest(r)
		headers.DoNotCache(w)
		var startAt, endAt sql.NullTime
		p := tables.PERIODS()
		err := sq.
			From(p).
			Where(
				p.COHORT.EqString(apt.skylb.CurrentCohort()),
				p.STAGE.EqString(skylab.StageApplication),
				p.MILESTONE.EqString(skylab.MilestoneNull),
			).
			SelectRowx(func(row *sq.Row) {
				startAt = row.NullTime(p.START_AT)
				endAt = row.NullTime(p.END_AT)
			}).
			Fetch(apt.skylb.DB)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				http.Redirect(w, r, "/applicant/application/closed", http.StatusMovedPermanently)
			default:
				apt.skylb.InternalServerError(w, r, err)
			}
			return
		}
		status := timeutil.ResolveTimestatus(startAt, endAt)
		if !status.IsOpen {
			http.Redirect(w, r, "/applicant/application/closed", http.StatusMovedPermanently)
			return
		}
		next.ServeHTTP(w, r)
	})
}
