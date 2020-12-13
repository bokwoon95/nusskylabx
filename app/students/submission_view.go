package students

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/app/skylab"

	"github.com/bokwoon95/nusskylabx/helpers/flash"
	"github.com/bokwoon95/nusskylabx/helpers/headers"
	"github.com/bokwoon95/nusskylabx/helpers/urlparams"
	"github.com/bokwoon95/nusskylabx/tables"
)

func (stu Students) SubmissionView(w http.ResponseWriter, r *http.Request) {
	stu.skylb.Log.TraceRequest(r)
	headers.DoNotCache(w)
	msgs := make(map[string][]string)
	submissionID, err := urlparams.Int(r, "submissionID")
	if err != nil {
		stu.skylb.BadRequest(w, r, err.Error())
		return
	}
	var submission skylab.Submission
	var editURL, submitURL string
	render := func() {
		r = stu.skylb.SetRoleSection(w, r, skylab.RoleStudent, stu.getSectionFromSubmissionID(submissionID))
		r, _ = stu.skylb.SetFlashMsgs(w, r, msgs)
		data := map[string]interface{}{
			"Submission": submission,
			"EditURL":    editURL,
			"SubmitURL":  submitURL,
		}
		stu.skylb.Render(w, r, data, "app/skylab/submission_view.html")
	}
	s := tables.V_SUBMISSIONS()
	err = sq.WithDefaultLog(sq.Lverbose).
		From(s).
		Where(s.SUBMISSION_ID.EqInt(submissionID)).
		SelectRowx(submission.RowMapper(s)).
		Fetch(stu.skylb.DB)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			stu.skylb.BadRequest(w, r, fmt.Sprintf("No submission found for submissionID %d", submissionID))
		default:
			msgs[flash.Error] = []string{err.Error()}
			w.WriteHeader(http.StatusInternalServerError)
			render()
		}
		return
	}
	canEdit, _ := r.Context().Value(skylab.ContextCanEditSubmission).(bool)
	if canEdit {
		editURL = skylab.StudentSubmission + "/" + strconv.Itoa(submissionID) + "/edit"
		submitURL = skylab.StudentSubmission + "/" + strconv.Itoa(submissionID) + "/submit"
	}
	render()
}

func (stu Students) CanViewSubmission(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stu.skylb.Log.TraceRequest(r)
		user, _ := r.Context().Value(skylab.ContextUser).(skylab.User)
		submissionID, err := urlparams.Int(r, "submissionID")
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

		// Get submission's teamID
		s := tables.SUBMISSIONS()
		var submissionTeamID int
		err = sq.From(s).Where(s.SUBMISSION_ID.EqInt(submissionID)).
			SelectRowx(func(row *sq.Row) { submissionTeamID = row.Int(s.TEAM_ID) }).
			Fetch(stu.skylb.DB)
		if err != nil {
			stu.skylb.BadRequest(w, r, fmt.Sprintf("submissionID %d doesn't exist", submissionID))
			return
		}

		// If user is visiting his own team's submission, he can view and edit
		// it
		if userTeamID == submissionTeamID {
			r = r.WithContext(context.WithValue(r.Context(), skylab.ContextCanViewSubmission, true))
			r = r.WithContext(context.WithValue(r.Context(), skylab.ContextCanEditSubmission, true))
			next.ServeHTTP(w, r)
			return
		}

		// Else check if user's team is an evaluator of submission's team
		tp := tables.TEAM_EVALUATION_PAIRS()
		rowsAffected, err := sq.WithDefaultLog(sq.Lstats).
			SelectOne().
			From(tp).
			Where(
				tp.EVALUATEE_TEAM_ID.EqInt(submissionTeamID),
				tp.EVALUATOR_TEAM_ID.EqInt(userTeamID),
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
		r = r.WithContext(context.WithValue(r.Context(), skylab.ContextCanViewSubmission, true))
		next.ServeHTTP(w, r)
	})
}

func (stu Students) getSectionFromSubmissionID(submissionID int) (section string) {
	s, f, p := tables.SUBMISSIONS(), tables.FORMS(), tables.PERIODS()
	var milestone string
	_ = sq.From(s).
		Join(f, f.FORM_ID.Eq(s.SUBMISSION_FORM_ID)).
		Join(p, p.PERIOD_ID.Eq(f.PERIOD_ID)).
		Where(s.SUBMISSION_ID.EqInt(submissionID)).
		SelectRowx(func(row *sq.Row) { milestone = row.String(p.MILESTONE) }).
		Fetch(stu.skylb.DB)
	switch milestone {
	case skylab.Milestone1:
		return skylab.StudentM1Submission
	case skylab.Milestone2:
		return skylab.StudentM2Submission
	case skylab.Milestone3:
		return skylab.StudentM3Submission
	default:
		return skylab.StudentM1Submission
	}
}
