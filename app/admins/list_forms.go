package admins

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/helpers/formutil"
	"github.com/bokwoon95/nusskylabx/helpers/urlparams"
	"github.com/bokwoon95/nusskylabx/tables"

	"github.com/bokwoon95/nusskylabx/app/skylab"

	"github.com/bokwoon95/nusskylabx/helpers/erro"
	"github.com/bokwoon95/nusskylabx/helpers/flash"
	"github.com/bokwoon95/nusskylabx/helpers/headers"
)

func (adm Admins) ListForms(w http.ResponseWriter, r *http.Request) {
	adm.skylb.Log.TraceRequest(r)
	r = adm.skylb.SetRoleSection(w, r, skylab.RoleAdmin, skylab.AdminListForms)
	headers.DoNotCache(w)

	// Get the last valid cohort
	cohort, _ := urlparams.PersistentString(w, r, "cohort", "_admin_list_forms_cohort")
	if cohort == "" || !skylab.Contains(adm.skylb.Cohorts(), cohort) {
		http.Redirect(w, r, skylab.AdminListForms+"/"+adm.skylb.CurrentCohort(), http.StatusMovedPermanently)
		return
	}

	type Data struct {
		Forms  []skylab.Form
		Cohort string
	}
	var data Data
	data.Cohort = cohort
	f, p := tables.FORMS(), tables.PERIODS()
	me, se := tables.MILESTONE_ENUM(), tables.STAGE_ENUM()
	var form skylab.Form
	err := sq.WithLog(adm.skylb.Log, sq.Lstats).
		From(f).
		Join(p, p.PERIOD_ID.Eq(f.PERIOD_ID)).
		Where(p.COHORT.EqString(cohort)).
		OrderBy(
			sq.Fieldf("array_position((SELECT array_agg(?) FROM ?), ?)", me.MILESTONE, me, p.MILESTONE),
			sq.Fieldf("array_position((SELECT array_agg(?) FROM ?), ?)", se.STAGE, se, p.STAGE),
		).
		Selectx(func(row *sq.Row) {
			form.Valid = row.IntValid(f.FORM_ID)
			form.FormID = row.Int(f.FORM_ID)
			form.Name = row.String(f.NAME)
			form.Subsection = row.String(f.SUBSECTION)
			form.Period.Valid = row.IntValid(p.PERIOD_ID)
			form.Period.PeriodID = row.Int(p.PERIOD_ID)
			form.Period.Cohort = row.String(p.COHORT)
			form.Period.Stage = row.String(p.STAGE)
			form.Period.Milestone = row.String(p.MILESTONE)
			form.Period.StartAt = row.NullTime(p.START_AT)
			form.Period.EndAt = row.NullTime(p.END_AT)
			row.ScanInto(&form.Questions, f.QUESTIONS)
		}, func() {
			data.Forms = append(data.Forms, form)
		}).
		Fetch(adm.skylb.DB)
	if err != nil {
		adm.skylb.InternalServerError(w, r, err)
		return
	}
	adm.skylb.Render(w, r, data, nil, "app/admins/list_forms.html")
}

func (adm Admins) ListFormsCreate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		adm.skylb.Log.TraceRequest(r)
		msgs := make(map[string][]string)
		_ = formutil.ParseForm(r)
		cohort := r.FormValue("cohort")
		stage := r.FormValue("stage")
		milestone := r.FormValue("milestone")
		name := r.FormValue("name")
		subsection := r.FormValue("subsection")
		pass := func(w http.ResponseWriter, r *http.Request, msgs map[string][]string, formID int) {
			r = urlparams.SetInt(r, "formID", formID)
			r, _ = adm.skylb.SetFlashMsgs(w, r, msgs)
			next.ServeHTTP(w, r)
		}
		fail := func(w http.ResponseWriter, r *http.Request, msgs map[string][]string) {
			r, _ = adm.skylb.SetFlashMsgs(w, r, msgs)
			adm.skylb.Redirect(skylab.AdminListForms)(w, r)
		}
		if cohort == "" {
			msgs[flash.Error] = []string{"A cohort must be specified"}
			fail(w, r, msgs)
			return
		}

		// select or create periodID with cohort, stage, milestone
		var periodID int
		p, f := tables.PERIODS(), tables.FORMS()
		err := sq.WithLog(adm.skylb.Log, sq.Lverbose).
			From(p).
			Where(
				p.COHORT.EqString(cohort),
				p.STAGE.EqString(stage),
				p.MILESTONE.EqString(milestone),
			).
			SelectRowx(func(row *sq.Row) { periodID = row.Int(p.PERIOD_ID) }).
			Fetch(adm.skylb.DB)
		if errors.Is(err, sql.ErrNoRows) {
			err = sq.WithLog(adm.skylb.Log, sq.Lverbose).
				InsertInto(p).Columns(p.COHORT, p.STAGE, p.MILESTONE).Values(cohort, stage, milestone).
				ReturningRowx(func(row *sq.Row) { periodID = row.Int(p.PERIOD_ID) }).
				Fetch(adm.skylb.DB)
		}
		if err != nil {
			msgs[flash.Error] = []string{err.Error()}
			fail(w, r, msgs)
			return
		}

		// select or create formID with periodID, name, subsection
		var formID int
		err = sq.WithLog(adm.skylb.Log, sq.Lverbose).
			From(f).
			Where(
				f.PERIOD_ID.EqInt(periodID),
				f.NAME.EqString(name),
				f.SUBSECTION.EqString(subsection),
			).
			SelectRowx(func(row *sq.Row) { formID = row.Int(f.FORM_ID) }).
			Fetch(adm.skylb.DB)
		if errors.Is(err, sql.ErrNoRows) {
			err = sq.WithLog(adm.skylb.Log, sq.Lverbose).
				InsertInto(f).Columns(f.PERIOD_ID, f.NAME, f.SUBSECTION).Values(periodID, name, subsection).
				ReturningRowx(func(row *sq.Row) { formID = row.Int(f.FORM_ID) }).
				Fetch(adm.skylb.DB)
		}
		if err != nil {
			msgs[flash.Error] = []string{err.Error()}
			fail(w, r, msgs)
			return
		}
		pass(w, r, msgs, formID)
	})
}

var errFormAlreadyExists = errors.New("Form already exists for target cohort and was not duplicated")

func (adm Admins) ListFormsDuplicate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		adm.skylb.Log.TraceRequest(r)
		_ = formutil.ParseForm(r)
		msgs := make(map[string][]string)
		nextHandler := func(w http.ResponseWriter, r *http.Request, msgs map[string][]string) {
			r, _ = adm.skylb.SetFlashMsgs(w, r, msgs)
			next.ServeHTTP(w, r)
		}
		cohort := r.FormValue("cohort")
		if cohort == "" {
			msgs[flash.Error] = append(msgs[flash.Error], "cohort cannot be blank")
			r = urlparams.SetString(r, "cohort", adm.skylb.CurrentCohort())
			nextHandler(w, r, msgs)
			return
		}
		r = urlparams.SetString(r, "cohort", cohort)

		// Duplicate one form
		if len(r.Form["stage"]) != 0 && len(r.Form["milestone"]) != 0 {
			stage := r.FormValue("stage")
			milestone := r.FormValue("milestone")
			formID := r.FormValue("formID")
			err := adm.duplicateForm(formID, cohort, stage, milestone)
			if err != nil {
				switch {
				case errors.Is(err, errFormAlreadyExists):
					msgs[flash.Warning] = append(msgs[flash.Warning], err.Error())
				default:
					msgs[flash.Error] = append(msgs[flash.Error], err.Error())
					log.Printf(err.Error())
				}
				log.Printf(err.Error())
				nextHandler(w, r, msgs)
				return
			}
			msgs[flash.Success] = append(msgs[flash.Success], fmt.Sprintf("1 form duplicated for cohort: %s, stage: %s, milestone: %s", cohort, stage, milestone))
			nextHandler(w, r, msgs)
			return
		}

		// Duplicate many forms
		var formsDuplicated int
		var formsAlreadyExisted int
		for i := range r.Form["formID"] {
			err := adm.duplicateForm(r.Form["formID"][i], cohort, "", "")
			if err != nil {
				switch {
				case errors.Is(err, errFormAlreadyExists):
					formsAlreadyExisted++
				default:
					msgs[flash.Error] = append(msgs[flash.Error], err.Error())
					log.Printf(err.Error())
				}
				continue
			}
			formsDuplicated++
		}
		if formsDuplicated > 0 {
			msgs[flash.Success] = append(msgs[flash.Success], fmt.Sprintf("%d forms duplicated for cohort %s", formsDuplicated, cohort))
		}
		if formsAlreadyExisted > 0 {
			msgs[flash.Warning] = append(msgs[flash.Warning], fmt.Sprintf("%d forms were not duplicated for cohort %s because they already exist", formsAlreadyExisted, cohort))
		}
		nextHandler(w, r, msgs)
	})
}

func (adm Admins) duplicateForm(formID string, cohort, stage, milestone string) error {
	p1, p2 := tables.PERIODS().As("p1"), tables.PERIODS().As("p2")
	f1, f2 := tables.FORMS().As("f1"), tables.FORMS().As("f2")
	var stageField, milestoneField sq.Field
	if stage != "" {
		stageField = sq.String(stage)
	} else {
		stageField = p2.STAGE
	}
	if milestone != "" {
		milestoneField = sq.String(milestone)
	} else {
		milestoneField = p2.MILESTONE
	}
	var periodID int
	err := sq.WithLog(adm.skylb.Log, sq.Lverbose).
		InsertInto(p1).
		Columns(p1.COHORT, p1.STAGE, p1.MILESTONE, p1.START_AT, p1.END_AT).
		Select(sq.
			Select(
				sq.String(cohort),
				stageField,
				milestoneField,
				// https://stackoverflow.com/a/56276474
				sq.Fieldf(
					"? + MAKE_INTERVAL(YEARS := ?::INT - EXTRACT(YEAR FROM ?)::INT)",
					p2.START_AT, cohort, p2.START_AT,
				),
				sq.Fieldf(
					"? + MAKE_INTERVAL(YEARS := ?::INT - EXTRACT(YEAR FROM ?)::INT)",
					p2.END_AT, cohort, p2.END_AT,
				),
			).
			From(p2).
			Join(f1, f1.PERIOD_ID.Eq(p2.PERIOD_ID)).
			Where(sq.Predicatef("? = ?", f1.FORM_ID, formID)).
			Limit(1),
		).
		OnConflict(p1.COHORT, p1.STAGE, p1.MILESTONE).
		DoUpdateSet(
			p1.START_AT.Set(sq.Excluded(p1.START_AT)),
			p1.END_AT.Set(sq.Excluded(p1.END_AT)),
		).
		ReturningRowx(func(row *sq.Row) { periodID = row.Int(p1.PERIOD_ID) }).
		Fetch(adm.skylb.DB)
	if err != nil {
		return erro.Wrap(err)
	}
	rowsAffected, err := sq.WithLog(adm.skylb.Log, sq.Lverbose).
		InsertInto(f1).
		Columns(f1.PERIOD_ID, f1.NAME, f1.SUBSECTION, f1.QUESTIONS).
		Select(sq.
			Select(sq.Int(periodID), f2.NAME, f2.SUBSECTION, f2.QUESTIONS).
			From(f2).Where(sq.Predicatef("? = ?", f2.FORM_ID, formID)),
		).
		OnConflict().DoNothing().
		Exec(adm.skylb.DB, sq.ErowsAffected)
	if err != nil {
		return erro.Wrap(err)
	}
	if rowsAffected == 0 {
		return errFormAlreadyExists
	}
	return nil
}

func (adm Admins) ListFormsDelete(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		adm.skylb.Log.TraceRequest(r)
		_ = formutil.ParseForm(r)
		msgs := make(map[string][]string)
		nextHandler := func(w http.ResponseWriter, r *http.Request, msgs map[string][]string) {
			r, _ = adm.skylb.SetFlashMsgs(w, r, msgs)
			next.ServeHTTP(w, r)
		}
		formIDs := r.Form["formID"]
		f := tables.FORMS()
		rowsAffected, err := sq.WithLog(adm.skylb.Log, sq.Lverbose).
			DeleteFrom(f).
			Where(f.FORM_ID.In(formIDs)).
			Exec(adm.skylb.DB, sq.ErowsAffected)
		if err != nil {
			if pqerr, ok := erro.AsPqError(err); ok && pqerr.Code == erro.PqForeignKeyViolation {
				msgs[flash.Error] = append(msgs[flash.Error],
					fmt.Sprintf("Unable to delete forms %+v because some are still referenced in the database", formIDs),
				)
			} else {
				log.Printf("[ERROR] deleting from forms with formIDs %+v: %s\n", formIDs, err)
				msgs[flash.Error] = append(msgs[flash.Error], fmt.Sprintf("Unable to delete forms %+v: %s", formIDs, err))
			}
		}
		if rowsAffected > 0 {
			msgs[flash.Success] = append(msgs[flash.Success], fmt.Sprintf("%d form(s) deleted", rowsAffected))
		}
		nextHandler(w, r, msgs)
	})
}
