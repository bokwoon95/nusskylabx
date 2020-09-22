package advisers

import (
	"context"
	"net/http"

	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/app/skylab"
	"github.com/bokwoon95/nusskylabx/tables"

	"github.com/bokwoon95/nusskylabx/helpers/urlparams"
)

func (adv Advisers) CanViewSubmission(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		adv.skylb.Log.TraceRequest(r)
		user, _ := r.Context().Value(skylab.ContextUser).(skylab.User)
		submissionID, err := urlparams.Int(r, "submissionID")
		if err != nil {
			adv.skylb.BadRequest(w, r, err.Error())
			return
		}
		t, s := tables.TEAMS(), tables.SUBMISSIONS()
		rowsAffected, err := sq.WithDefaultLog(sq.Lstats).
			SelectOne().
			From(t).
			Where(
				t.TEAM_ID.In(sq.Select(s.TEAM_ID).From(s).Where(s.SUBMISSION_ID.EqInt(submissionID))),
				t.ADVISER_USER_ROLE_ID.EqInt(user.Roles[skylab.RoleAdviser]),
			).
			Exec(adv.skylb.DB, sq.ErowsAffected)
		if err != nil {
			adv.skylb.InternalServerError(w, r, err)
			return
		}
		if rowsAffected == 0 {
			adv.skylb.NotAuthorized(w, r)
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), skylab.ContextCanViewSubmission, true))
		next.ServeHTTP(w, r)
	})
}
