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
	"github.com/bokwoon95/nusskylabx/helpers/urlparams"
)

func (stu Students) TeamEvaluationEdit(w http.ResponseWriter, r *http.Request) {
	stu.skylb.Log.TraceRequest(r)
	r = stu.skylb.SetRoleSection(w, r, skylab.RoleStudent, skylab.SectionPreserve)
	var data skylab.TeamEvaluationEdit
	teamEvaluationID, err := urlparams.Int(r, "teamEvaluationID")
	if err != nil {
		stu.skylb.InternalServerError(w, r, err)
		return
	}
	te := tables.V_TEAM_EVALUATIONS()
	err = sq.WithLog(stu.skylb.Log, sq.Lstats).
		From(te).
		Where(te.TEAM_EVALUATION_ID.EqInt(teamEvaluationID)).
		SelectRowx((&data.TeamEvaluation).RowMapper(te)).
		Fetch(stu.skylb.DB)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			stu.skylb.BadRequest(w, r, fmt.Sprintf("No such team evaluation with teamEvaluationID %d", teamEvaluationID))
		default:
			stu.skylb.InternalServerError(w, r, err)
		}
		return
	}
	data.SubmissionURL = skylab.StudentSubmission + "/" + strconv.Itoa(data.TeamEvaluation.Evaluatee.SubmissionID)
	data.PreviewURL = skylab.StudentTeamEvaluation + "/" + strconv.Itoa(teamEvaluationID) + "/preview"
	data.UpdateURL = skylab.StudentTeamEvaluation + "/" + strconv.Itoa(teamEvaluationID) + "/update"
	data.SubmitURL = skylab.StudentTeamEvaluation + "/" + strconv.Itoa(teamEvaluationID) + "/submit"
	funcs := formx.Funcs(nil, stu.skylb.Policy)
	stu.skylb.Render(w, r, data, funcs,
		"app/skylab/team_evaluation_edit.html",
		"helpers/formx/render_form.html",
		"helpers/formx/render_form_results.html",
	)
}

func (stu Students) CanViewTeamEvaluation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stu.skylb.Log.TraceRequest(r)
		user, _ := r.Context().Value(skylab.ContextUser).(skylab.User)
		teamEvaluationID, err := urlparams.Int(r, "teamEvaluationID")
		if err != nil {
			stu.skylb.InternalServerError(w, r, err)
			return
		}

		// Get user's teamID
		urs := tables.USER_ROLES_STUDENTS()
		var userTeamID int
		err = sq.WithLog(stu.skylb.Log, sq.Lstats).
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

		// Get the evaluator/evaluatee teamID for the given team evaluation
		var evaluatorTeamID, evaluateeTeamID int
		te, s := tables.TEAM_EVALUATIONS(), tables.SUBMISSIONS()
		err = sq.WithLog(stu.skylb.Log, sq.Lstats).
			From(te).
			Join(s, s.SUBMISSION_ID.Eq(te.EVALUATEE_SUBMISSION_ID)).
			Where(te.TEAM_EVALUATION_ID.EqInt(teamEvaluationID)).
			SelectRowx(func(row *sq.Row) {
				evaluatorTeamID = row.Int(te.EVALUATOR_TEAM_ID)
				evaluateeTeamID = row.Int(s.TEAM_ID)
			}).
			Fetch(stu.skylb.DB)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				stu.skylb.BadRequest(w, r, fmt.Sprintf("No such team evaluation with teamEvaluationID %d", teamEvaluationID))
			default:
				stu.skylb.InternalServerError(w, r, err)
			}
			return
		}

		switch {
		case userTeamID == evaluatorTeamID:
			r = r.WithContext(context.WithValue(r.Context(), skylab.ContextCanViewEvaluation, true))
			r = r.WithContext(context.WithValue(r.Context(), skylab.ContextCanEditEvaluation, true))
			next.ServeHTTP(w, r)
		case userTeamID == evaluateeTeamID:
			r = r.WithContext(context.WithValue(r.Context(), skylab.ContextCanViewEvaluation, true))
			next.ServeHTTP(w, r)
		default:
			stu.skylb.NotAuthorized(w, r)
		}
	})
}

func (stu Students) CanEditTeamEvaluation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stu.skylb.Log.TraceRequest(r)
		user, _ := r.Context().Value(skylab.ContextUser).(skylab.User)
		teamEvaluationID, err := urlparams.Int(r, "teamEvaluationID")
		if err != nil {
			stu.skylb.BadRequest(w, r, err.Error())
			return
		}

		// Get user's teamID
		urs := tables.USER_ROLES_STUDENTS()
		var userTeamID int
		err = sq.WithLog(stu.skylb.Log, sq.Lstats).
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

		// Get the evaluator/evaluatee teamID for the given team evaluation
		var evaluatorTeamID, evaluateeTeamID int
		te, s := tables.TEAM_EVALUATIONS(), tables.SUBMISSIONS()
		err = sq.WithLog(stu.skylb.Log, sq.Lstats).
			From(te).
			Join(s, s.SUBMISSION_ID.Eq(te.EVALUATEE_SUBMISSION_ID)).
			Where(te.TEAM_EVALUATION_ID.EqInt(teamEvaluationID)).
			SelectRowx(func(row *sq.Row) {
				evaluatorTeamID = row.Int(te.EVALUATOR_TEAM_ID)
				evaluateeTeamID = row.Int(s.TEAM_ID)
			}).
			Fetch(stu.skylb.DB)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				stu.skylb.BadRequest(w, r, fmt.Sprintf("No such team evaluation with teamEvaluationID %d", teamEvaluationID))
			default:
				stu.skylb.InternalServerError(w, r, err)
			}
			return
		}

		switch {
		case userTeamID == evaluatorTeamID:
			r = r.WithContext(context.WithValue(r.Context(), skylab.ContextCanViewEvaluation, true))
			r = r.WithContext(context.WithValue(r.Context(), skylab.ContextCanEditEvaluation, true))
			next.ServeHTTP(w, r)
		case userTeamID == evaluateeTeamID:
			fallthrough
		default:
			stu.skylb.NotAuthorized(w, r)
		}
	})
}

func (stu Students) TeamEvaluationCreate(w http.ResponseWriter, r *http.Request) {
	stu.skylb.Log.TraceRequest(r)
	user, _ := r.Context().Value(skylab.ContextUser).(skylab.User)
	submissionID, err := strconv.Atoi(r.FormValue("submissionID"))
	if err != nil {
		if r.FormValue("submissionID") == "" {
			stu.skylb.BadRequest(w, r, fmt.Sprintf("submissionID was not found or is blank"))
		} else {
			stu.skylb.BadRequest(w, r, fmt.Sprintf("'%s' is not a valid team submission id (need integer)", r.FormValue("submissionID")))
		}
		return
	}

	// Get user's (evaluator) teamID
	urs := tables.USER_ROLES_STUDENTS()
	var evaluatorTeamID int
	err = sq.WithLog(stu.skylb.Log, sq.Lstats).
		From(urs).
		Where(urs.USER_ROLE_ID.EqInt(user.Roles[skylab.RoleStudent])).
		SelectRowx(func(row *sq.Row) { evaluatorTeamID = row.Int(urs.TEAM_ID) }).
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

	// Ensure the user's team is authorized to evaluate this submission's team
	s, tp := tables.SUBMISSIONS(), tables.TEAM_EVALUATION_PAIRS()
	rowsAffected, err := sq.WithLog(stu.skylb.Log, sq.Lstats).
		From(s).
		Join(tp, tp.EVALUATEE_TEAM_ID.Eq(s.TEAM_ID)).
		Where(
			s.SUBMISSION_ID.EqInt(submissionID),
			tp.EVALUATOR_TEAM_ID.EqInt(evaluatorTeamID),
		).
		SelectOne().
		Exec(stu.skylb.DB, sq.ErowsAffected)
	if err != nil {
		stu.skylb.InternalServerError(w, r, err)
		return
	}
	if rowsAffected == 0 {
		stu.skylb.Log.Printf(fmt.Sprintf("either teamID %[1]d is not authorized to evaluate submissionID %[2]d, or submissionID %[2]d doesn't exist", evaluatorTeamID, submissionID))
		stu.skylb.NotAuthorized(w, r)
		return
	}

	// See if the teamEvaluationID already exists. If so, redirect to it.
	te := tables.TEAM_EVALUATIONS()
	var teamEvaluationID int
	err = sq.WithLog(stu.skylb.Log, sq.Lstats).
		From(te).
		Where(
			te.EVALUATOR_TEAM_ID.EqInt(evaluatorTeamID),
			te.EVALUATEE_SUBMISSION_ID.EqInt(submissionID),
		).
		SelectRowx(func(row *sq.Row) {
			teamEvaluationID = row.Int(te.TEAM_EVALUATION_ID)
		}).
		Fetch(stu.skylb.DB)
	if err == nil {
		// existing teamEvaluationID was found
		url := skylab.StudentTeamEvaluation + "/" + strconv.Itoa(teamEvaluationID) + "/edit"
		http.Redirect(w, r, url, http.StatusMovedPermanently)
		return
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		stu.skylb.InternalServerError(w, r, err)
		return
	}

	// Get evaluation formID
	milestone := r.FormValue("milestone")
	if milestone == "" {
		stu.skylb.BadRequest(w, r, fmt.Sprintf("milestone not specified!"))
		return
	}
	if !skylab.Contains(skylab.Milestones(), milestone) {
		stu.skylb.BadRequest(w, r, fmt.Sprintf("milestone '%s' is not a valid Skylab milestone", milestone))
		return
	}
	f, p := tables.FORMS(), tables.PERIODS()
	var evaluationFormID int
	err = sq.WithLog(stu.skylb.Log, sq.Lstats).
		From(f).
		Join(p, p.PERIOD_ID.Eq(f.PERIOD_ID)).
		Where(
			p.COHORT.EqString(stu.skylb.CurrentCohort()),
			p.STAGE.EqString(skylab.StageEvaluation),
			p.MILESTONE.EqString(milestone),
			f.NAME.EqString(""),
			f.SUBSECTION.EqString(""),
		).
		SelectRowx(func(row *sq.Row) { evaluationFormID = row.Int(f.FORM_ID) }).
		Fetch(stu.skylb.DB)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			stu.skylb.BadRequest(w, r, fmt.Sprintf("Unfortunately it seems like the administrator has not created the form yet"))
		default:
			stu.skylb.InternalServerError(w, r, err)
		}
		return
	}

	err = sq.WithLog(stu.skylb.Log, sq.Lverbose).
		InsertInto(te).
		Columns(te.EVALUATOR_TEAM_ID, te.EVALUATEE_SUBMISSION_ID, te.EVALUATION_FORM_ID).
		Values(evaluatorTeamID, submissionID, evaluationFormID).
		ReturningRowx(func(row *sq.Row) { teamEvaluationID = row.Int(te.TEAM_EVALUATION_ID) }).
		Fetch(stu.skylb.DB)
	if err != nil {
		stu.skylb.InternalServerError(w, r, err)
		return
	}
	url := skylab.StudentTeamEvaluation + "/" + strconv.Itoa(teamEvaluationID) + "/edit"
	http.Redirect(w, r, url, http.StatusMovedPermanently)
}

func (stu Students) TeamEvaluationUpdate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stu.skylb.Log.TraceRequest(r)
		msgs := make(map[string][]string)
		teamEvaluationID, err := urlparams.Int(r, "teamEvaluationID")
		if err != nil {
			stu.skylb.InternalServerError(w, r, err)
			return
		}
		_ = formutil.ParseForm(r)
		err = stu.UpdateEvaluationAnswers(teamEvaluationID, r.Form)
		if err != nil {
			msgs[flash.Error] = []string{erro.Wrap(err).Error()}
		} else {
			msgs[flash.Success] = []string{"Updated!"}
		}
		r, _ = stu.skylb.SetFlashMsgs(w, r, msgs)
		next.ServeHTTP(w, r)
	})
}

func (stu Students) TeamEvaluationSubmit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stu.skylb.Log.TraceRequest(r)
		msgs := make(map[string][]string)
		teamEvaluationID, err := urlparams.Int(r, "teamEvaluationID")
		if err != nil {
			stu.skylb.InternalServerError(w, r, err)
			return
		}
		te := tables.TEAM_EVALUATIONS()
		_, err = sq.WithLog(stu.skylb.Log, sq.Lstats).
			Update(te).
			Set(te.SUBMITTED.SetBool(true)).
			Where(te.TEAM_EVALUATION_ID.EqInt(teamEvaluationID)).
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
