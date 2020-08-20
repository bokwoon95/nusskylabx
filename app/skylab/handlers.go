package skylab

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/helpers/erro"
	"github.com/bokwoon95/nusskylabx/tables"
	"github.com/go-chi/chi"
)

// Redirect returns a http.HandlerFunc that redirects to the given url
func (skylb Skylab) Redirect(url string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		skylb.Log.TraceRequest(r)
		matchgrps := regexp.MustCompile(`{([^{}]+)}`).FindAllStringSubmatch(url, -1)
		newUrl := url
		skylb.Log.Printf("Should redirect to: %s", url)
		// Graft the chi URL Params from the initial url onto the newURL. If
		// /users/{uid}/edit should redirect to /users/{uid}, /users/1/edit ->
		// graft {uid}=1 -> should redirect to /users/1
		for _, matchgrp := range matchgrps {
			curlyBraceParam := matchgrp[0]
			key := matchgrp[1]
			// URL params can optionally take a regex after a colon ':' e.g.
			// `{id:\d+}`. Hence the key is actually everything before the
			// colon
			if colonIdx := strings.Index(key, ":"); colonIdx > 0 {
				key = key[:colonIdx]
			}
			value := chi.URLParam(r, key)
			if value == "" {
				skylb.InternalServerError(w, r, erro.Wrap(fmt.Errorf(
					"should redirect to '%s' but unable to find '%s' from route context."+
						" Did you call urlparams.SetString/urlparams.SetInt?",
					url, curlyBraceParam,
				)))
				return
			}
			newUrl = strings.ReplaceAll(newUrl, curlyBraceParam, value)
			skylb.Log.Printf("Substitute %s with %s", curlyBraceParam, value)
		}
		skylb.Log.Printf("Redirecting to: %s", newUrl)
		http.Redirect(w, r, newUrl, http.StatusMovedPermanently)
	}
}

// AddProdContext will inject Skylab's environment (production or development)
// into the request context for handlers down the chain to pick up on
func (skylb Skylab) AddProdContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, ContextIsProd, skylb.IsProd)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (skylb Skylab) RedirectUserrole(w http.ResponseWriter, r *http.Request) {
	user, _ := r.Context().Value(ContextUser).(User)
	if !user.Valid {
		skylb.NotLoggedIn(w, r)
		return
	}
	if user.Roles == nil {
		user.Roles = make(map[string]int)
	}
	var role string
	var userRoleID int
	u, ur := tables.USERS(), tables.USER_ROLES()
	err := sq.WithLog(skylb.Log, sq.Lverbose).
		From(u).
		Join(ur, ur.USER_ID.Eq(u.USER_ID)).
		Where(
			u.EMAIL.EqString(user.Email),
			ur.COHORT.EqString(skylb.CurrentCohort()),
		).
		Selectx(func(row *sq.Row) {
			role = row.String(ur.ROLE)
			userRoleID = row.Int(ur.USER_ROLE_ID)
		}, func() {
			user.Roles[role] = userRoleID
		}).
		Fetch(skylb.DB)
	if err != nil {
		skylb.InternalServerError(w, r, err)
		return
	}
	switch {
	case user.Roles[RoleAdmin] != 0:
		http.Redirect(w, r, "/admin", http.StatusMovedPermanently)
	case user.Roles[RoleMentor] != 0:
		http.Redirect(w, r, "/mentor", http.StatusMovedPermanently)
	case user.Roles[RoleAdviser] != 0:
		http.Redirect(w, r, "/adviser", http.StatusMovedPermanently)
	case user.Roles[RoleStudent] != 0:
		http.Redirect(w, r, "/student", http.StatusMovedPermanently)
	case user.Roles[RoleApplicant] != 0:
		http.Redirect(w, r, "/applicant", http.StatusMovedPermanently)
	default:
		skylb.NotLoggedIn(w, r)
	}
}

func (skylb Skylab) AllowIfDevelopment(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		skylb.Log.TraceRequest(r)
		isProd, _ := r.Context().Value(ContextIsProd).(bool)
		if !isProd {
			skylb.Log.Printf("environment is development, allowing user through")
			next.ServeHTTP(w, r)
			return
		}
		admin, _ := r.Context().Value(ContextAdmin).(User)
		if !admin.Valid {
			skylb.Log.Printf("admin is not a valid admin %+v", admin)
			skylb.NotAnAdmin(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (skylb Skylab) HasValidRole(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(ContextUser).(User)
		if !user.Valid {
			skylb.NotLoggedIn(w, r)
			return
		}
		var hasRoleThatIsValid bool
		for role := range user.Roles {
			if Contains(Roles(), role) {
				hasRoleThatIsValid = true
				break
			}
		}
		if !hasRoleThatIsValid {
			skylb.NotAuthorized(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
