package students

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/app/skylab"
	"github.com/bokwoon95/nusskylabx/tables"

	"github.com/bokwoon95/nusskylabx/helpers/urlparams"
)

func (stu Students) CanViewUserEvaluation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stu.skylb.Log.TraceRequest(r)
		user, _ := r.Context().Value(skylab.ContextUser).(skylab.User)
		userEvaluationID, err := urlparams.Int(r, "userEvaluationID")
		if err != nil {
			stu.skylb.BadRequest(w, r, err.Error())
			return
		}

		// Get user's teamID
		urs := tables.USER_ROLES_STUDENTS()
		var userTeamID int
		err = sq.From(urs).Where(urs.USER_ROLE_ID.EqInt(user.Roles[skylab.RoleStudent])).
			SelectRowx(func(row *sq.Row) { userTeamID = row.Int(urs.TEAM_ID) }).
			Fetch(stu.skylb.DB)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				stu.StudentNoTeam(w, r)
			default:
				stu.skylb.InternalServerError(w, r, err)
			}
			return
		}

		// Get evaluations's evaluatorUserRoleID
		ue := tables.USER_EVALUATIONS()
		var evaluatorUserRoleID int
		err = sq.From(ue).Where(ue.USER_EVALUATION_ID.EqInt(userEvaluationID)).
			SelectRowx(func(row *sq.Row) { evaluatorUserRoleID = row.Int(ue.EVALUATOR_USER_ROLE_ID) }).
			Fetch(stu.skylb.DB)
		if err != nil {
			stu.skylb.BadRequest(w, r, fmt.Sprintf("userEvaluationID %d doesn't exist", userEvaluationID))
			return
		}

		// Check if the evaluatorUserRoleID is either the team's adviser or mentor
		t := tables.TEAMS()
		rowsAffected, err := sq.WithLog(stu.skylb.Log, sq.Lstats).
			SelectOne().
			From(t).
			Where(
				t.TEAM_ID.EqInt(userTeamID),
				sq.Int(evaluatorUserRoleID).In(sq.Fields{t.ADVISER_USER_ROLE_ID, t.MENTOR_USER_ROLE_ID}),
			).
			Exec(stu.skylb.DB, sq.ErowsAffected)
		if err != nil {
			stu.skylb.InternalServerError(w, r, err)
			return
		}
		if rowsAffected == 0 {
			stu.skylb.NotAuthorized(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
