package skylab

import (
	"html/template"
	"net/http"
)

// NavbarFuncs contain the Template Functions required for rendering the navbar "app/skylab/navbar.html"
func (skylb Skylab) NavbarFuncs(funcs template.FuncMap, w http.ResponseWriter, r *http.Request) template.FuncMap {
	if funcs == nil {
		funcs = template.FuncMap{}
	}
	funcs = addConstRole(funcs)
	funcs["SkylabCsrfToken"] = csrfToken(r)
	funcs["SkylabIsProd"] = func() bool { return skylb.IsProd }
	funcs["SkylabCurrentRole"] = getCurrentRole(r)
	funcs["SkylabCurrentSection"] = currentSection(r)
	funcs["SkylabUser"] = getUser(r)
	funcs["SkylabAdmin"] = getAdmin(r)
	funcs["SkylabUserIsRole"] = userIsRole
	funcs["SkylabUserIsApplicantOnly"] = userIsApplicantOnly
	funcs["AdminCreateUser"] = func() string { return AdminCreateUser }
	return funcs
}

// userIsRole is a Template Function that checks if the given User is a
// role
func userIsRole(user User, role string) bool {
	if !user.Valid || len(user.Roles) == 0 {
		return false
	}
	return user.Roles[role] != 0
}

// userIsApplicantOnly is a Template Function that checks if the given
// User only is a RoleApplicant and no other role
func userIsApplicantOnly(user User) bool {
	if !user.Valid || len(user.Roles) == 0 {
		return false
	}
	if len(user.Roles) > 1 || user.Roles[RoleApplicant] == 0 {
		return false
	}
	return true
}

// getCurrentRole is a Template Function that retrieves the user's current role
// from the context
func getCurrentRole(r *http.Request) func() string {
	currentRole, _ := r.Context().Value(ContextCurrentRole).(string)
	return func() string {
		return currentRole
	}
}

// currentSection is a Template Function that retrieves the user's current
// section from the context
func currentSection(r *http.Request) func() string {
	currentSection, _ := r.Context().Value(ContextCurrentSection).(string)
	return func() string {
		return currentSection
	}
}

// getUser is a Template Function that retrieves a User from
// the context for the template to use
func getUser(r *http.Request) func() User {
	user, _ := r.Context().Value(ContextUser).(User)
	return func() User {
		return user
	}
}

// getAdmin is a Template Function that retrieves an Admin from
// the context for the template to use
func getAdmin(r *http.Request) func() User {
	admin, _ := r.Context().Value(ContextAdmin).(User)
	return func() User {
		return admin
	}
}
