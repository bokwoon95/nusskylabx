package advisers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/app/skylab"
	"github.com/bokwoon95/nusskylabx/tables"

	"github.com/bokwoon95/nusskylabx/helpers/erro"
	"github.com/bokwoon95/nusskylabx/helpers/flash"
	"github.com/bokwoon95/nusskylabx/helpers/formutil"
	"github.com/bokwoon95/nusskylabx/helpers/formx"
	"github.com/bokwoon95/nusskylabx/helpers/urlparams"
)

func (adv Advisers) CanViewUserEvaluation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		adv.skylb.Log.TraceRequest(r)
		user, _ := r.Context().Value(skylab.ContextUser).(skylab.User)
		userEvaluationID, err := urlparams.Int(r, "userEvaluationID")
		if err != nil {
			adv.skylb.InternalServerError(w, r, err)
			return
		}
		ue := tables.USER_EVALUATIONS()
		rowsAffected, err := sq.WithDefaultLog(sq.Lverbose).
			SelectOne().
			From(ue).
			Where(
				ue.USER_EVALUATION_ID.EqInt(userEvaluationID),
				ue.EVALUATOR_USER_ROLE_ID.EqInt(user.Roles[skylab.RoleAdviser]),
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
		next.ServeHTTP(w, r)
	})
}

func (adv Advisers) CanEditUserEvaluation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		adv.skylb.Log.TraceRequest(r)
		user, _ := r.Context().Value(skylab.ContextUser).(skylab.User)
		userEvaluationID, err := urlparams.Int(r, "userEvaluationID")
		if err != nil {
			adv.skylb.BadRequest(w, r, err.Error())
			return
		}
		ue := tables.USER_EVALUATIONS()
		rowsAffected, err := sq.WithDefaultLog(sq.Lstats).
			SelectOne().
			From(ue).
			Where(
				ue.USER_EVALUATION_ID.EqInt(userEvaluationID),
				ue.EVALUATOR_USER_ROLE_ID.EqInt(user.Roles[skylab.RoleAdviser]),
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
		next.ServeHTTP(w, r)
	})
}

func (adv Advisers) UserEvaluationCreate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		adv.skylb.Log.TraceRequest(r)
		user, _ := r.Context().Value(skylab.ContextUser).(skylab.User)
		submissionID, err := formutil.Int(r, "submissionID")
		if err != nil {
			adv.skylb.BadRequest(w, r, err.Error())
			return
		}
		milestone, err := formutil.String(r, "milestone")
		if errors.Is(err, formutil.ErrValueMissing) {
			adv.skylb.BadRequest(w, r, fmt.Sprintf("milestone form value not provided"))
			return
		} else if milestone == "" || !skylab.Contains(skylab.Milestones(), milestone) {
			adv.skylb.BadRequest(w, r, fmt.Sprintf("'%s' is not a valid skylab milestone", milestone))
			return
		}

		// Get the formID for the milestone evaluation
		p, f := tables.PERIODS(), tables.FORMS()
		var formID int
		err = sq.WithDefaultLog(sq.Lverbose).
			From(f).
			Join(p, p.PERIOD_ID.Eq(f.PERIOD_ID)).
			Where(
				p.COHORT.EqString(adv.skylb.CurrentCohort()),
				p.STAGE.EqString(skylab.StageEvaluation),
				p.MILESTONE.EqString(milestone),
				f.NAME.EqString(""),
				f.SUBSECTION.EqString(""),
			).
			SelectRowx(func(row *sq.Row) { formID = row.Int(f.FORM_ID) }).
			Fetch(adv.skylb.DB)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				adv.skylb.BadRequest(w, r, "Unfortunately it seems like the administrator has not created the submission form yet.")
			default:
				adv.skylb.InternalServerError(w, r, err)
			}
			return
		}

		// Insert or select a user evaluation, returning the userEvaluationID
		ue := tables.USER_EVALUATIONS()
		var userEvaluationID int
		err = sq.WithDefaultLog(sq.Lverbose).
			InsertInto(ue).
			Columns(ue.EVALUATOR_USER_ROLE_ID, ue.EVALUATEE_SUBMISSION_ID, ue.EVALUATION_FORM_ID).
			Values(user.Roles[skylab.RoleAdviser], submissionID, formID).
			OnConflict(ue.EVALUATOR_USER_ROLE_ID, ue.EVALUATEE_SUBMISSION_ID).DoNothing().
			ReturningRowx(func(row *sq.Row) { userEvaluationID = row.Int(ue.USER_EVALUATION_ID) }).
			Fetch(adv.skylb.DB)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			adv.skylb.InternalServerError(w, r, err)
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			err = sq.WithDefaultLog(sq.Lverbose).
				From(ue).
				Where(
					ue.EVALUATOR_USER_ROLE_ID.EqInt(user.Roles[skylab.RoleAdviser]),
					ue.EVALUATEE_SUBMISSION_ID.EqInt(submissionID),
				).
				SelectRowx(func(row *sq.Row) { userEvaluationID = row.Int(ue.USER_EVALUATION_ID) }).
				Fetch(adv.skylb.DB)
			if err != nil {
				adv.skylb.InternalServerError(w, r, err)
				return
			}
		}

		r = urlparams.SetInt(r, "userEvaluationID", userEvaluationID)
		next.ServeHTTP(w, r)
	})
}

func (adv Advisers) UserEvaluationUpdate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		adv.skylb.Log.TraceRequest(r)
		msgs := make(map[string][]string)
		userEvaluationID, err := urlparams.Int(r, "userEvaluationID")
		if err != nil {
			adv.skylb.BadRequest(w, r, err.Error())
			return
		}
		var questions formx.Questions
		var answers formx.Answers
		ue, f := tables.USER_EVALUATIONS(), tables.FORMS()
		err = sq.WithDefaultLog(sq.Lstats).
			From(ue).
			Join(f, f.FORM_ID.Eq(ue.EVALUATION_FORM_ID)).
			Where(ue.USER_EVALUATION_ID.EqInt(userEvaluationID)).
			SelectRowx(func(row *sq.Row) { row.ScanInto(&questions, f.QUESTIONS) }).
			Fetch(adv.skylb.DB)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				adv.skylb.BadRequest(w, r, fmt.Sprintf("No such evaluation {userEvaluationID:%d}", userEvaluationID))
			default:
				adv.skylb.InternalServerError(w, r, err)
			}
			return
		}
		_ = formutil.ParseForm(r)
		answers = formx.ExtractAnswers(r.Form, questions)
		// If no answers are present at all (which is different from answers having
		// blank values), do not proceed with the data update as that is not what
		// we want under any circumstance. If a user wishes to clear out an answer,
		// they would at least provide an empty string.
		if answers.IsEmpty() {
			next.ServeHTTP(w, r)
			return
		}
		_, err = sq.WithDefaultLog(sq.Lstats).
			Update(ue).
			Set(ue.EVALUATION_DATA.Set(answers)).
			Where(ue.USER_EVALUATION_ID.EqInt(userEvaluationID)).
			Exec(adv.skylb.DB, sq.ErowsAffected)
		if err != nil {
			msgs[flash.Error] = []string{erro.Wrap(err).Error()}
		} else {
			msgs[flash.Success] = []string{"Updated!"}
		}
		r, _ = adv.skylb.SetFlashMsgs(w, r, msgs)
		next.ServeHTTP(w, r)
	})
}

func (adv Advisers) UserEvaluationSubmit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		adv.skylb.Log.TraceRequest(r)
		msgs := make(map[string][]string)
		userEvaluationID, err := urlparams.Int(r, "userEvaluationID")
		if err != nil {
			adv.skylb.BadRequest(w, r, err.Error())
			return
		}
		ue := tables.USER_EVALUATIONS()
		_, err = sq.WithDefaultLog(sq.Lstats).
			Update(ue).
			Set(ue.SUBMITTED.SetBool(true)).
			Where(ue.USER_EVALUATION_ID.EqInt(userEvaluationID)).
			Exec(adv.skylb.DB, sq.ErowsAffected)
		if err != nil {
			msgs[flash.Error] = []string{erro.Wrap(err).Error()}
		} else {
			msgs[flash.Success] = []string{"Submitted!"}
		}
		r, _ = adv.skylb.SetFlashMsgs(w, r, msgs)
		next.ServeHTTP(w, r)
	})
}
