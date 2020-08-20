package applicants

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"strings"

	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/app/skylab"
	"github.com/bokwoon95/nusskylabx/tables"

	"github.com/bokwoon95/nusskylabx/helpers/erro"
	"github.com/bokwoon95/nusskylabx/helpers/flash"
	"github.com/bokwoon95/nusskylabx/helpers/formutil"
	"github.com/bokwoon95/nusskylabx/helpers/formx"
)

func (apt Applicants) Application(w http.ResponseWriter, r *http.Request) {
	apt.skylb.Log.TraceRequest(r)
	r = apt.skylb.SetRoleSection(w, r, skylab.RoleApplicant, skylab.SectionPreserve)
	user, _ := r.Context().Value(skylab.ContextUser).(skylab.User)
	a := tables.V_APPLICATIONS()
	rowsAffected, err := sq.WithLog(apt.skylb.Log, sq.Lstats).
		SelectOne().
		From(a).
		Where(
			a.COHORT.EqString(apt.skylb.CurrentCohort()),
			sq.Int(user.UserID).In(sq.Fields{a.APPLICANT1_USER_ID, a.APPLICANT2_USER_ID}),
		).
		Exec(apt.skylb.DB, sq.ErowsAffected)
	if err != nil {
		apt.skylb.InternalServerError(w, r, err)
		return
	}
	if rowsAffected == 0 {
		http.Redirect(w, r, "/applicant", http.StatusMovedPermanently)
		return
	}
	var data skylab.ApplicationEdit
	data.ApplicantUserID = user.UserID
	err = sq.WithLog(apt.skylb.Log, sq.Lverbose).
		From(a).
		Where(
			sq.String(user.Email).In(sq.Fields{a.APPLICANT1_EMAIL, a.APPLICANT2_EMAIL}),
			a.COHORT.EqString(apt.skylb.CurrentCohort()),
		).
		SelectRowx((&data.Application).RowMapper(a)).
		Fetch(apt.skylb.DB)
	if err != nil {
		apt.skylb.InternalServerError(w, r, err)
		return
	}
	funcs := template.FuncMap{}
	funcs = formx.Funcs(funcs, apt.skylb.Policy)
	apt.skylb.Render(w, r, data, funcs, "app/applicants/application.html", "helpers/formx/render_form.html")
}

func (apt Applicants) MagicstringVerifier(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apt.skylb.Log.TraceRequest(r)
		msgs := make(map[string][]string)
		magicstring := strings.TrimSpace(r.FormValue("magicstring"))
		redirect := func(w http.ResponseWriter, r *http.Request, msgs map[string][]string) {
			// If user logged in, redirect them to /applicant. If user not
			// logged in, show them invalid_magicstring.html instead.
			r, _ = apt.skylb.SetFlashMsgs(w, r, msgs)
			user, _ := r.Context().Value(skylab.ContextUser).(skylab.User)
			if user.Valid {
				http.Redirect(w, r, "/applicant", http.StatusMovedPermanently)
				return
			}
			apt.skylb.Render(w, r, nil, nil, "app/applicants/invalid_magicstring.html")
		}
		// if magicstring cannot be obtained from form.Value, check context as well
		if magicstring == "" {
			var ok bool
			magicstring, ok = r.Context().Value("magicstring").(string)
			if !ok {
				msgs[flash.Error] = []string{"magicstring cannot be blank"}
				redirect(w, r, msgs)
				return
			}
		}
		magicstringRegex := regexp.MustCompile("^[a-zA-Z0-9]{32}$")
		if !magicstringRegex.MatchString(magicstring) {
			msgs[flash.Error] = []string{"invalid magicstring"}
			if l := len(magicstring); l > 32 && magicstringRegex.MatchString(magicstring[l-32:]) {
				msgs[flash.Error][0] += ", did you perhaps mean " + magicstring[l-32:] + "?"
			}
			redirect(w, r, msgs)
			return
		}
		a := tables.APPLICATIONS()
		rowsAffected, err := sq.WithLog(apt.skylb.Log, sq.Lstats).
			SelectOne().
			From(a).
			Where(a.MAGICSTRING.EqString(magicstring)).
			Exec(apt.skylb.DB, sq.ErowsAffected)
		if err != nil {
			apt.skylb.InternalServerError(w, r, err)
			return
		}
		if rowsAffected == 0 {
			msgs[flash.Error] = []string{"invalid magicstring"}
			redirect(w, r, msgs)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (apt Applicants) IdempotentCreateApplication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apt.skylb.Log.TraceRequest(r)
		user, _ := r.Context().Value(skylab.ContextUser).(skylab.User)
		if !user.Valid {
			apt.skylb.NotLoggedIn(w, r)
			return
		}
		_, err := sq.WithLog(apt.skylb.Log, sq.Lstats).
			Select(tables.IDEMPOTENT_CREATE_APPLICATION(user.Displayname, user.Email)).
			Exec(apt.skylb.DB, sq.ErowsAffected)
		if err != nil {
			apt.skylb.InternalServerError(w, r, err)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (apt Applicants) JoinApplication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apt.skylb.Log.TraceRequest(r)
		displayname, ok2 := r.Context().Value("displayname").(string)
		email, ok3 := r.Context().Value("email").(string)
		user := skylab.User{Displayname: displayname, Email: email}
		if !ok2 || !ok3 {
			apt.skylb.BadRequest(w, r, fmt.Sprintf("Incomplete user retrieved from context: %+v", user))
			return
		}
		magicstring := r.FormValue("magicstring")
		err := apt.JoinApplicationModel(user, magicstring)
		if err != nil {
			apt.skylb.InternalServerError(w, r, err)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (apt Applicants) JoinIfLoggedin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apt.skylb.Log.TraceRequest(r)
		user, _ := r.Context().Value(skylab.ContextUser).(skylab.User)
		if !user.Valid {
			next.ServeHTTP(w, r)
			return
		}
		msgs := map[string][]string{}
		magicstring := r.FormValue("magicstring")
		err := apt.JoinApplicationModel(user, magicstring)
		if err != nil {
			if pqerr, ok := erro.AsPqError(err); ok {
				switch pqerr.Code {
				case skylab.ErrApplicantJoinedOwnApplication.PqCode():
					msgs[flash.Error] = append(msgs[flash.Error], "You cannot join your own application")
					r, _ = apt.skylb.SetFlashMsgs(w, r, msgs)
					http.Redirect(w, r, "/applicant/application", http.StatusMovedPermanently)
					return
				}
			}
			apt.skylb.InternalServerError(w, r, err)
			return
		}
		http.Redirect(w, r, "/applicant", http.StatusMovedPermanently)
	})
}

func (apt Applicants) LeaveApplication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apt.skylb.Log.TraceRequest(r)
		user, _ := r.Context().Value(skylab.ContextUser).(skylab.User)
		if !user.Valid {
			apt.skylb.BadRequest(w, r, fmt.Sprintf("Unable to obtain user from context"))
			return
		}
		err := apt.LeaveApplicationModel(user)
		if err != nil {
			apt.skylb.InternalServerError(w, r, err)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (apt Applicants) UpdateApplication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apt.skylb.Log.TraceRequest(r)
		user, _ := r.Context().Value(skylab.ContextUser).(skylab.User)
		_ = formutil.ParseForm(r)
		p, f := tables.PERIODS(), tables.FORMS()
		// Get application Questions
		var applicationQuestions formx.Questions
		err := sq.From(p).Join(f, f.PERIOD_ID.Eq(p.PERIOD_ID)).Where(
			p.COHORT.EqString(apt.skylb.CurrentCohort()),
			p.STAGE.EqString(skylab.StageApplication),
			p.MILESTONE.EqString(skylab.MilestoneNull),
			f.NAME.EqString(""),
			f.SUBSECTION.EqString(skylab.ApplicationSubsectionApplication), // <= "application"
		).SelectRowx(func(row *sq.Row) {
			row.ScanInto(&applicationQuestions, f.QUESTIONS)
		}).Fetch(apt.skylb.DB)
		if err != nil {
			apt.skylb.InternalServerError(w, r, err)
			return
		}
		// Get application Answers
		applicationAnswers := formx.ExtractAnswers(r.Form, applicationQuestions)
		// Get applicant Questions
		var applicantQuestions formx.Questions
		err = sq.From(p).Join(f, f.PERIOD_ID.Eq(p.PERIOD_ID)).Where(
			p.COHORT.EqString(apt.skylb.CurrentCohort()),
			p.STAGE.EqString(skylab.StageApplication),
			p.MILESTONE.EqString(skylab.MilestoneNull),
			f.NAME.EqString(""),
			f.SUBSECTION.EqString(skylab.ApplicationSubsectionApplicant), // <= "applicant"
		).SelectRowx(func(row *sq.Row) {
			row.ScanInto(&applicantQuestions, f.QUESTIONS)
		}).Fetch(apt.skylb.DB)
		if err != nil {
			apt.skylb.InternalServerError(w, r, err)
			return
		}
		// Get applicant Answers
		applicantAnswers := formx.ExtractAnswers(r.Form, applicantQuestions)
		// Upsert application + applicant form data
		userRoleID := user.Roles[skylab.RoleApplicant]
		_, err = sq.WithLog(apt.skylb.Log, sq.Lverbose).
			Select(tables.UPSERT_APPLICATION_DATA(userRoleID, applicantAnswers, applicationAnswers)).
			Exec(apt.skylb.DB, sq.ErowsAffected)
		if err != nil {
			apt.skylb.InternalServerError(w, r, err)
			return
		}
		var msgs = make(map[string][]string)
		msgs[flash.Success] = []string{"Application updated"}
		r, _ = apt.skylb.SetFlashMsgs(w, r, msgs)
		next.ServeHTTP(w, r)
	})
}

func (apt Applicants) SubmitApplication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apt.skylb.Log.TraceRequest(r)
		user, _ := r.Context().Value(skylab.ContextUser).(skylab.User)
		msgs := make(map[string][]string)
		a, va := tables.APPLICATIONS(), tables.V_APPLICATIONS()
		var applicationID int
		var applicant1UserID, applicant2UserID sql.NullInt64
		err := sq.From(va).Where(
			sq.Int(user.UserID).In(sq.Fields{
				va.APPLICANT1_USER_ID,
				va.APPLICANT2_USER_ID,
			}),
		).SelectRowx(func(row *sq.Row) {
			applicationID = row.Int(va.APPLICATION_ID)
			applicant1UserID = row.NullInt64(va.APPLICANT1_USER_ID)
			applicant2UserID = row.NullInt64(va.APPLICANT2_USER_ID)
		}).Fetch(apt.skylb.DB)
		if err != nil {
			apt.skylb.InternalServerError(w, r, err)
			return
		}
		if !applicant1UserID.Valid || !applicant2UserID.Valid {
			msgs[flash.Error] = []string{"You cannot submit your application as you are still missing a second team member"}
			goto Next
		}
		_, err = sq.WithLog(apt.skylb.Log, sq.Lstats).
			Update(a).
			Set(a.SUBMITTED.SetBool(true)).
			Where(a.APPLICATION_ID.EqInt(applicationID)).
			Exec(apt.skylb.DB, sq.ErowsAffected)
		if err != nil {
			apt.skylb.InternalServerError(w, r, err)
			return
		}
		msgs[flash.Success] = []string{"Application submitted"}
	Next:
		r, _ = apt.skylb.SetFlashMsgs(w, r, msgs)
		next.ServeHTTP(w, r)
	})
}
