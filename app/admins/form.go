package admins

import (
	"net/http"
	"strconv"

	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/app/skylab"
	"github.com/bokwoon95/nusskylabx/helpers/flash"
	"github.com/bokwoon95/nusskylabx/helpers/formutil"
	"github.com/bokwoon95/nusskylabx/helpers/formx"
	"github.com/bokwoon95/nusskylabx/helpers/urlparams"
	"github.com/bokwoon95/nusskylabx/tables"
)

func (adm Admins) FormEdit(w http.ResponseWriter, r *http.Request) {
	adm.skylb.Log.TraceRequest(r)
	r = adm.skylb.SetRoleSection(w, r, skylab.RolePreserve, skylab.SectionPreserve)
	formID, err := urlparams.Int(r, "formID")
	if err != nil {
		adm.skylb.BadRequest(w, r, err.Error())
		return
	}
	f, p := tables.FORMS(), tables.PERIODS()
	var form skylab.Form
	err = sq.
		From(f).
		Join(p, p.PERIOD_ID.Eq(f.PERIOD_ID)).
		Where(f.FORM_ID.EqInt(formID)).
		SelectRowx(func(row *sq.Row) {
			// Form
			form.Valid = row.IntValid(f.FORM_ID)
			form.FormID = row.Int(f.FORM_ID)
			form.Name = row.String(f.NAME)
			form.Subsection = row.String(f.SUBSECTION)
			row.ScanInto(&form.Questions, f.QUESTIONS)
			// Period
			form.Period.Valid = row.IntValid(p.PERIOD_ID)
			form.Period.PeriodID = row.Int(p.PERIOD_ID)
			form.Period.Cohort = row.String(p.COHORT)
			form.Period.Stage = row.String(p.STAGE)
			form.Period.Milestone = row.String(p.MILESTONE)
			form.Period.StartAt = row.NullTime(p.START_AT)
			form.Period.EndAt = row.NullTime(p.END_AT)
		}).
		Fetch(adm.skylb.DB)
	if err != nil {
		adm.skylb.BadRequest(w, r, err.Error())
		return
	}
	data := map[string]interface{}{
		"Title":            form.Title(),
		"Form":             form,
		"QuestionsAnswers": formx.MergeQuestionsAnswers(form.Questions, formx.Answers{}),
		"PreviewURL":       skylab.AdminForm + "/" + strconv.Itoa(formID) + "/preview",
		"UpdateURL":        skylab.AdminForm + "/" + strconv.Itoa(formID) + "/update",
	}
	adm.skylb.Wender(w, r, data, "app/skylab/form_edit.html")
}

func (adm Admins) FormUpdate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		formID, err := urlparams.Int(r, "formID")
		if err != nil {
			adm.skylb.BadRequest(w, r, err.Error())
			return
		}
		msgs := make(map[string][]string)
		_ = formutil.ParseForm(r)
		data := r.FormValue("data")
		f := tables.FORMS()
		_, err = sq.WithDefaultLog(sq.Lstats).
			Update(f).
			Set(f.QUESTIONS.Set(data)).
			Where(f.FORM_ID.EqInt(formID)).
			Exec(adm.skylb.DB, sq.ErowsAffected)
		if err != nil {
			msgs[flash.Error] = []string{err.Error()}
		} else {
			msgs[flash.Success] = []string{"Form updated!"}
		}
		r, _ = adm.skylb.SetFlashMsgs(w, r, msgs)
		next.ServeHTTP(w, r)
	})
}

func (adm Admins) FormView(w http.ResponseWriter, r *http.Request) {
	adm.skylb.Log.TraceRequest(r)
	r = adm.skylb.SetRoleSection(w, r, skylab.RolePreserve, skylab.SectionPreserve)
	formID, err := urlparams.Int(r, "formID")
	if err != nil {
		adm.skylb.BadRequest(w, r, err.Error())
		return
	}
	var questions formx.Questions
	f := tables.FORMS()
	err = sq.
		From(f).
		Where(f.FORM_ID.EqInt(formID)).
		SelectRowx(func(row *sq.Row) {
			row.ScanInto(&questions, f.QUESTIONS)
		}).
		Fetch(adm.skylb.DB)
	if err != nil {
		adm.skylb.BadRequest(w, r, err.Error())
		return
	}
	data := map[string]interface{}{
		"Title":            "",
		"Form":             skylab.Form{},
		"QuestionsAnswers": formx.MergeQuestionsAnswers(questions, formx.Answers{}),
		"EditURL":          skylab.AdminForm + "/" + strconv.Itoa(formID) + "/edit",
		"UpdateURL":        skylab.AdminForm + "/" + strconv.Itoa(formID) + "/update",
	}
	adm.skylb.Wender(w, r, data, "app/skylab/form_view.html")
}
