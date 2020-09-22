package advisers

import (
	"context"
	"net/http"

	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/app/skylab"
	"github.com/bokwoon95/nusskylabx/helpers/urlparams"
	"github.com/bokwoon95/nusskylabx/tables"
)

func (adv Advisers) CanViewTeamEvaluation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		adv.skylb.Log.TraceRequest(r)
		user, _ := r.Context().Value(skylab.ContextUser).(skylab.User)
		teamEvaluationID, err := urlparams.Int(r, "teamEvaluationID")
		if err != nil {
			adv.skylb.BadRequest(w, r, err.Error())
			return
		}
		adviserTeamIDs, err := adv.getTeamIDs(user)
		if err != nil {
			adv.skylb.InternalServerError(w, r, err)
			return
		}
		te, s := tables.TEAM_EVALUATIONS(), tables.SUBMISSIONS()
		rowsAffected, err := sq.WithDefaultLog(sq.Lstats).
			SelectOne().
			From(te).
			Join(s, s.SUBMISSION_ID.Eq(te.EVALUATEE_SUBMISSION_ID)).
			Where(
				te.TEAM_EVALUATION_ID.EqInt(teamEvaluationID),
				te.EVALUATOR_TEAM_ID.In(adviserTeamIDs),
				s.TEAM_ID.In(adviserTeamIDs),
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
		r = r.WithContext(context.WithValue(r.Context(), skylab.ContextCanViewEvaluation, true))
		next.ServeHTTP(w, r)
	})
}
