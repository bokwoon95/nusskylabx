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
	"github.com/bokwoon95/nusskylabx/tables"

	"github.com/bokwoon95/nusskylabx/helpers/erro"
	"github.com/bokwoon95/nusskylabx/helpers/flash"
	"github.com/bokwoon95/nusskylabx/helpers/formutil"
	"github.com/bokwoon95/nusskylabx/helpers/formx"
	"github.com/bokwoon95/nusskylabx/helpers/headers"
	"github.com/bokwoon95/nusskylabx/helpers/urlparams"
)

func (stu Students) SubmissionEdit(w http.ResponseWriter, r *http.Request) {
	stu.skylb.Log.TraceRequest(r)
	headers.DoNotCache(w)
	var data skylab.SubmissionEditData
	var msgs = make(map[string][]string)
	submissionID, err := urlparams.Int(r, "submissionID")
	if err != nil {
		stu.skylb.BadRequest(w, r, err.Error())
		return
	}
	render := func(data skylab.SubmissionEditData, msgs map[string][]string) {
		var funcs map[string]interface{}
		funcs = formx.Funcs(funcs, stu.skylb.Policy)
		r = stu.skylb.SetRoleSection(w, r, skylab.RoleStudent, stu.getSectionFromSubmissionID(submissionID))
		r, _ = stu.skylb.SetFlashMsgs(w, r, msgs)
		stu.skylb.Render(w, r, data, funcs, "app/skylab/submission_edit.html", "helpers/formx/render_form.html")
	}
	s := tables.V_SUBMISSIONS()
	err = sq.WithDefaultLog(sq.Lstats).
		From(s).
		Where(s.SUBMISSION_ID.EqInt(submissionID)).
		SelectRowx((&data.Submission).RowMapper(s)).
		Fetch(stu.skylb.DB)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			stu.skylb.BadRequest(w, r, fmt.Sprintf("No submission found for submissionID %d", submissionID))
		default:
			msgs[flash.Error] = []string{err.Error()}
			w.WriteHeader(http.StatusInternalServerError)
			render(data, msgs)
		}
		return
	}
	data.PreviewURL = skylab.StudentSubmission + "/" + strconv.Itoa(submissionID) + "/preview"
	data.UpdateURL = skylab.StudentSubmission + "/" + strconv.Itoa(submissionID) + "/update"
	data.SubmitURL = skylab.StudentSubmission + "/" + strconv.Itoa(submissionID) + "/submit"

	// Team evaluations
	te := tables.V_TEAM_EVALUATIONS()
	var teamEvaluation skylab.TeamEvaluation
	err = sq.WithDefaultLog(sq.Lstats).
		From(te).
		Where(te.SUBMISSION_ID.EqInt(submissionID)).
		Selectx((&teamEvaluation).RowMapper(te), func() {
			data.PeerEvaluations = append(data.PeerEvaluations, teamEvaluation)
		}).
		Fetch(stu.skylb.DB)
	if err != nil {
		stu.skylb.InternalServerError(w, r, err)
		return
	}

	// Adviser + Mentor Evaluations
	ue := tables.V_USER_EVALUATIONS()
	var userEvaluation skylab.UserEvaluation
	err = sq.WithDefaultLog(sq.Lstats).
		From(ue).
		Where(ue.SUBMISSION_ID.EqInt(submissionID)).
		Selectx((&userEvaluation).RowMapper(ue), func() {
			switch userEvaluation.Role {
			case skylab.RoleAdviser:
				data.AdviserEvaluation = userEvaluation
			case skylab.RoleMentor:
				data.MentorEvaluation = userEvaluation
			}
		}).
		Fetch(stu.skylb.DB)
	if err != nil {
		stu.skylb.InternalServerError(w, r, err)
		return
	}
	render(data, msgs)
}

func (stu Students) IdempotentSubmissionCreate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stu.skylb.Log.TraceRequest(r)
		headers.DoNotCache(w)

		// Try to see which milestone we are creating a submission for
		milestone := r.FormValue("milestone")
		if milestone == "" {
			// If milestone cannot be found in r.FormValue(), try r.Context()
			var ok bool
			milestone, ok = r.Context().Value(skylab.ContextCurrentMilestone).(string)
			if !ok {
				stu.skylb.InternalServerError(w, r, fmt.Errorf("milestone not specified!"))
				return
			}
		}

		// Get the user's teamID
		user, _ := r.Context().Value(skylab.ContextUser).(skylab.User)
		ur, urs := tables.USER_ROLES(), tables.USER_ROLES_STUDENTS()
		var teamID int
		err := sq.WithDefaultLog(sq.Lverbose).
			From(ur).
			LeftJoin(urs, urs.USER_ROLE_ID.Eq(ur.USER_ROLE_ID)).
			Where(
				ur.COHORT.EqString(stu.skylb.CurrentCohort()),
				ur.ROLE.EqString(skylab.RoleStudent),
				ur.USER_ID.EqInt(user.UserID),
			).
			SelectRowx(func(row *sq.Row) { teamID = row.Int(urs.TEAM_ID) }).
			Fetch(stu.skylb.DB)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				stu.skylb.NotAuthorized(w, r)
			default:
				stu.skylb.InternalServerError(w, r, err)
			}
			return
		}
		if teamID == 0 {
			stu.StudentNoTeam(w, r)
			return
		}

		// Get the formID for the milestone submission
		p, f := tables.PERIODS(), tables.FORMS()
		var formID int
		err = sq.WithDefaultLog(sq.Lverbose).
			From(f).
			Join(p, p.PERIOD_ID.Eq(f.PERIOD_ID)).
			Where(
				p.COHORT.EqString(stu.skylb.CurrentCohort()),
				p.STAGE.EqString(skylab.StageSubmission),
				p.MILESTONE.EqString(milestone),
				f.NAME.EqString(""),
				f.SUBSECTION.EqString(""),
			).
			SelectRowx(func(row *sq.Row) { formID = row.Int(f.FORM_ID) }).
			Fetch(stu.skylb.DB)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				stu.skylb.BadRequest(w, r, "Unfortunately it seems like the administrator has not created the submission form yet.")
			default:
				stu.skylb.InternalServerError(w, r, err)
			}
			return
		}

		// Create a new submission entry with teamID and formID, if we didn't
		// return any submissionID from the insert it means that there was an
		// insert conflict i.e. the submission already exists. In which case we
		// simply SELECT for it instead.
		s := tables.SUBMISSIONS()
		var submissionID int
		err = sq.WithDefaultLog(sq.Lverbose).
			InsertInto(s).
			Columns(s.TEAM_ID, s.SUBMISSION_FORM_ID).
			Values(teamID, formID).
			OnConflict(s.TEAM_ID, s.SUBMISSION_FORM_ID).DoNothing().
			ReturningRowx(func(row *sq.Row) { submissionID = row.Int(s.SUBMISSION_ID) }).
			Fetch(stu.skylb.DB)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			stu.skylb.InternalServerError(w, r, err)
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			err = sq.
				From(s).
				Where(
					s.TEAM_ID.EqInt(teamID),
					s.SUBMISSION_FORM_ID.EqInt(formID),
				).
				SelectRowx(func(row *sq.Row) { submissionID = row.Int(s.SUBMISSION_ID) }).
				Fetch(stu.skylb.DB)
			if err != nil {
				stu.skylb.InternalServerError(w, r, err)
				return
			}
		}

		// At this point a valid submissionID should have been found/created. Inject it
		// into the URL param context.
		r = urlparams.SetInt(r, "submissionID", submissionID)
		next.ServeHTTP(w, r)
	})
}

func (stu Students) SubmissionUpdate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stu.skylb.Log.TraceRequest(r)
		msgs := make(map[string][]string)
		submissionID, err := urlparams.Int(r, "submissionID")
		if err != nil {
			stu.skylb.BadRequest(w, r, err.Error())
			return
		}
		_ = formutil.ParseForm(r)
		err = stu.UpdateSubmissionAnswers(submissionID, r.Form)
		if err != nil {
			msgs[flash.Error] = []string{err.Error()}
			goto Redirect
		}
		msgs[flash.Success] = []string{"Updated!"}
	Redirect:
		r, _ = stu.skylb.SetFlashMsgs(w, r, msgs)
		next.ServeHTTP(w, r)
	})
}

func (stu Students) SubmissionSubmit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stu.skylb.Log.TraceRequest(r)
		msgs := make(map[string][]string)
		submissionID, err := urlparams.Int(r, "submissionID")
		if err != nil {
			stu.skylb.InternalServerError(w, r, err)
			return
		}
		s := tables.SUBMISSIONS()
		_, err = sq.WithDefaultLog(sq.Lverbose).
			Update(s).
			Set(s.SUBMITTED.SetBool(true)).
			Where(s.SUBMISSION_ID.EqInt(submissionID)).
			Exec(stu.skylb.DB, sq.ErowsAffected)
		if err != nil {
			msgs[flash.Error] = []string{erro.Wrap(err).Error()}
		} else {
			msgs[flash.Success] = []string{"Submitted!"}
		}
		r, _ = stu.skylb.SetFlashMsgs(w, r, msgs)
		next.ServeHTTP(w, r)
	})
}

func (stu Students) CanEditSubmission(next http.Handler) http.Handler {
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
		err = sq.
			From(urs).
			Where(urs.USER_ROLE_ID.EqInt(user.Roles[skylab.RoleStudent])).
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
		err = sq.
			From(s).
			Where(s.SUBMISSION_ID.EqInt(submissionID)).
			SelectRowx(func(row *sq.Row) { submissionTeamID = row.Int(s.TEAM_ID) }).
			Fetch(stu.skylb.DB)
		if err != nil {
			stu.skylb.BadRequest(w, r, fmt.Sprintf("submissionID %d doesn't exist", submissionID))
			return
		}

		// If user is not accessing his own team's submission, he cannot edit
		// it
		if userTeamID != submissionTeamID {
			stu.skylb.NotAuthorized(w, r)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), skylab.ContextCanViewSubmission, true))
		r = r.WithContext(context.WithValue(r.Context(), skylab.ContextCanEditSubmission, true))
		next.ServeHTTP(w, r)
	})
}
