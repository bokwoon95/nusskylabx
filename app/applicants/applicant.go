// Package applicants implements the applicant-facing pages on Skylab
package applicants

import (
	"errors"
	"fmt"
	"net/http"

	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/app/skylab"
	"github.com/bokwoon95/nusskylabx/tables"
)

func (apt Applicants) Applicant(w http.ResponseWriter, r *http.Request) {
	apt.skylb.Log.TraceRequest(r)
	r = apt.skylb.SetRoleSection(w, r, skylab.RoleApplicant, skylab.SectionPreserve)
	user, _ := r.Context().Value(skylab.ContextUser).(skylab.User)
	if !user.Valid {
		apt.skylb.BadRequest(w, r, fmt.Sprintf("Unable to obtain user from context"))
		return
	}
	a := tables.V_APPLICATIONS()
	rowsAffected, err := sq.WithDefaultLog(sq.Lstats).
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
	if rowsAffected != 0 {
		http.Redirect(w, r, "/applicant/application", http.StatusMovedPermanently)
		return
	}
	apt.skylb.Wender(w, r, nil, "app/applicants/applicant.html")
}

func (apt Applicants) IdempotentCreateApplicant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apt.skylb.Log.TraceRequest(r)
		displayname, ok2 := r.Context().Value("displayname").(string)
		email, ok3 := r.Context().Value("email").(string)
		user := skylab.User{Displayname: displayname, Email: email}
		if !ok2 || !ok3 {
			apt.skylb.BadRequest(w, r, fmt.Sprintf("Incomplete user retrieved from context: %+v", user))
			return
		}
		var err error
		user.Roles = map[string]int{skylab.RoleApplicant: 0}
		user, err = apt.d.CreateUser(user, apt.skylb.CurrentCohort())
		if err != nil {
			switch {
			case errors.Is(err, skylab.ErrEmailEmpty):
				apt.skylb.BadRequest(w, r, "Email cannot be empty")
			default:
				apt.skylb.InternalServerError(w, r, err)
			}
			return
		}
		apt.skylb.Log.Printf("%+v", user)
		next.ServeHTTP(w, r)
	})
}
