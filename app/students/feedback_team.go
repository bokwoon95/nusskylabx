package students

import (
	"net/http"

	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/app/skylab"
	"github.com/bokwoon95/nusskylabx/helpers/headers"
	"github.com/bokwoon95/nusskylabx/helpers/urlparams"
	"github.com/bokwoon95/nusskylabx/tables"
)

func (stu Students) TeamFeedbackEdit(w http.ResponseWriter, r *http.Request) {
	stu.skylb.Log.TraceRequest(r)
	r = stu.skylb.SetRoleSection(w, r, skylab.RoleStudent, skylab.StudentTeamFeedback)
	headers.DoNotCache(w)

	var teamFeedback skylab.TeamFeedback
	var err error
	teamFeedback.FeedbackIDOnTeam, err = urlparams.Int(r, "feedbackIDOnTeam")
	if err != nil {
		stu.skylb.BadRequest(w, r, err.Error())
		return
	}
	data := map[string]interface{}{
		"TeamFeedback": teamFeedback,
	}
	stu.skylb.Render(w, r, data, "app/students/feedback_team.html")
}

func (stu Students) CanEditTeamFeedback(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(skylab.ContextUser).(skylab.User)
		feedbackIDOnTeam, err := urlparams.Int(r, "feedbackIDOnTeam")
		if err != nil {
			stu.skylb.BadRequest(w, r, err.Error())
			return
		}
		t, ft := tables.V_TEAMS(), tables.FEEDBACK_ON_TEAMS()
		rowsAffected, err := sq.WithDefaultLog(sq.Lverbose).
			SelectOne().
			From(ft).
			Join(t, t.TEAM_ID.Eq(ft.EVALUATOR_TEAM_ID)).
			Where(
				ft.FEEDBACK_ID_ON_TEAM.EqInt(feedbackIDOnTeam),
				sq.Int(user.UserID).In(sq.Fields{
					t.STUDENT1_USER_ID,
					t.STUDENT2_USER_ID,
				}),
			).
			Exec(stu.skylb.DB, sq.ErowsAffected)
		if err != nil {
			stu.skylb.InternalServerError(w, r, err)
			return
		}
		if rowsAffected == 0 {
			// Evaluating team and evaluated team are not under the same adviser
			stu.skylb.NotAuthorized(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
