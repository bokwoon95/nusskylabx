package skylab

import (
	"context"
	"html/template"
	"net/http"

	"github.com/bokwoon95/nusskylabx/helpers/cookies"
	"github.com/bokwoon95/nusskylabx/helpers/headers"
)

const (
	SectionPreserve = "preserve"

	StudentDashboard      = "/student/dashboard"
	StudentSubmission     = "/student/submission"
	StudentUserEvaluation = "/student/user-evaluation"
	StudentTeamEvaluation = "/student/team-evaluation"
	StudentM1Submission   = "/student/milestone1/submission"
	StudentM1Evaluation   = "/student/milestone1/evaluation"
	StudentM2Submission   = "/student/milestone2/submission"
	StudentM2Evaluation   = "/student/milestone2/evaluation"
	StudentM3Submission   = "/student/milestone3/submission"
	StudentM3Evaluation   = "/student/milestone3/evaluation"
	StudentTeamFeedback   = "/student/feedback/team"
	StudentUserFeedback   = "/student/feedback/user"
	StudentListFeedbacks  = "/student/feedbacks"
	StudentTeam           = "/student/teams"

	AdviserDashboard           = "/adviser/dashboard"
	AdviserTeams               = "/adviser/teams"
	AdviserEvaluateeEvaluators = "/adviser/evaluatee-evaluators"
	AdviserEvaluatorEvaluatees = "/adviser/evaluator-evaluatees"
	AdviserSubmission          = "/adviser/submission"
	AdviserUserEvaluation      = "/adviser/user-evaluation"
	AdviserTeamEvaluation      = "/adviser/team-evaluation"
	AdviserM1MakeEvaluation    = "/adviser/milestone1/evaluation"
	AdviserM1ViewEvaluation    = "/adviser/milestone1/evaluations"
	AdviserM2MakeEvaluation    = "/adviser/milestone2/evaluation"
	AdviserM2ViewEvaluation    = "/adviser/milestone2/evaluations"
	AdviserM3MakeEvaluation    = "/adviser/milestone3/evaluation"
	AdviserM3ViewEvaluation    = "/adviser/milestone3/evaluations"

	MentorDashboard = "/mentor/dashboard"

	AdminDashboard         = "/admin/dashboard"
	AdminCreateUser        = "/admin/create-user"
	AdminCreateUserConfirm = "/admin/create-user/confirm"
	AdminListCohorts       = "/admin/cohorts"
	AdminListUsers         = "/admin/users"
	AdminUser              = "/admin/user"
	AdminListPeriods       = "/admin/periods"
	AdminListForms         = "/admin/forms"
	AdminForm              = "/admin/form"
	AdminListTeams         = "/admin/teams"
	AdminTeam              = "/admin/team"
	AdminListApplications  = "/admin/applications"
	AdminApplication       = "/admin/application"
	AdminListFeedbacks     = "/admin/feedbacks"
	AdminDumpJson          = "/dump-json"
	AdminTestmail          = "/testmail" // experimental
)

var sectionSymbols = map[string]string{
	StudentDashboard:      "StudentDashboard",
	StudentTeam:           "StudentTeam",
	StudentSubmission:     "StudentSubmission",
	StudentUserEvaluation: "StudentUserEvaluation",
	StudentTeamEvaluation: "StudentTeamEvaluation",
	StudentM1Submission:   "StudentM1Submission",
	StudentM1Evaluation:   "StudentM1Evaluation",
	StudentM2Submission:   "StudentM2Submission",
	StudentM2Evaluation:   "StudentM2Evaluation",
	StudentM3Submission:   "StudentM3Submission",
	StudentM3Evaluation:   "StudentM3Evaluation",
	StudentTeamFeedback:   "StudentTeamFeedback",
	StudentUserFeedback:   "StudentUserFeedback",
	StudentListFeedbacks:  "StudentListFeedbacks",

	AdviserDashboard:           "AdviserDashboard",
	AdviserTeams:               "AdviserTeams",
	AdviserEvaluateeEvaluators: "AdviserEvaluateeEvaluators",
	AdviserEvaluatorEvaluatees: "AdviserEvaluatorEvaluatees",
	AdviserSubmission:          "AdviserSubmission",
	AdviserUserEvaluation:      "AdviserUserEvaluation",
	AdviserTeamEvaluation:      "AdviserTeamEvaluation",
	AdviserM1MakeEvaluation:    "AdviserM1MakeEvaluation",
	AdviserM1ViewEvaluation:    "AdviserM1ViewEvaluation",
	AdviserM2MakeEvaluation:    "AdviserM2MakeEvaluation",
	AdviserM2ViewEvaluation:    "AdviserM2ViewEvaluation",
	AdviserM3MakeEvaluation:    "AdviserM3MakeEvaluation",
	AdviserM3ViewEvaluation:    "AdviserM3ViewEvaluation",

	MentorDashboard: "MentorDashboard",

	AdminDashboard:         "AdminDashboard",
	AdminCreateUser:        "AdminCreateUser",
	AdminCreateUserConfirm: "AdminCreateUserConfirm",
	AdminListCohorts:       "AdminListCohorts",
	AdminListUsers:         "AdminListUsers",
	AdminUser:              "AdminUser",
	AdminListPeriods:       "AdminListPeriods",
	AdminListForms:         "AdminListForms",
	AdminForm:              "AdminForm",
	AdminListTeams:         "AdminListTeams",
	AdminTeam:              "AdminTeam",
	AdminListApplications:  "AdminListApplications",
	AdminApplication:       "AdminApplication",
	AdminListFeedbacks:     "AdminListFeedbacks",
	AdminDumpJson:          "AdminDumpJson",
	AdminTestmail:          "AdminTestmail", // experimental
}

func AddSections(funcs template.FuncMap) template.FuncMap {
	for section, symbol := range sectionSymbols {
		section, symbol := section, symbol
		funcs[symbol] = func() string { return section }
	}
	return funcs
}

func IsValidSection(section string) bool {
	_, ok := sectionSymbols[section]
	return ok
}

func (skylb Skylab) SetRoleSection(w http.ResponseWriter, r *http.Request, role, section string) *http.Request {
	skylb.Log.TraceRequest(r)
	ctx := r.Context()

	// Set ContextCurrentRole
	switch role {
	case RolePreserve:
		cookie, _ := r.Cookie(LastRoleCookieName)
		if cookie != nil {
			role = cookie.Value
			if Contains(Roles(), role) {
				ctx = context.WithValue(ctx, ContextCurrentRole, role)
				skylb.Log.Printf("Preserving Role: Found cookie '%s' with valid role '%s'", LastRoleCookieName, role)
			} else {
				skylb.Log.Printf("Failed Preserving Role: Found cookie '%s' with invalid role '%s'", LastRoleCookieName, role)
			}
		} else {
			skylb.Log.Printf("Failed Preserving Role: Cookie '%s' not found", LastRoleCookieName)
		}
	default:
		ctx = context.WithValue(ctx, ContextCurrentRole, role)
		cookies.SetCookie(w, LastRoleCookieName, role)
		skylb.Log.Printf("Setting Role '%s'", role)
	}

	if !Contains(Roles(), role) {
		skylb.Log.Printf("Unable to set section because role '%s' is invalid", role)
		return r.WithContext(ctx)
	}

	// Set ContextCurrentSection
	switch section {
	case SectionPreserve:
		cookiename := LastSectionCookieName(role)
		cookie, _ := r.Cookie(cookiename)
		if cookiename != "" && cookie != nil {
			section = cookie.Value
			if IsValidSection(section) {
				ctx = context.WithValue(ctx, ContextCurrentSection, section)
				skylb.Log.Printf("Preserving Section: Found cookie '%s' with valid section '%s'", cookiename, section)
			} else {
				skylb.Log.Printf("Failed Preserving Section: Found cookie '%s' with invalid section '%s'", cookiename, section)
			}
		} else {
			skylb.Log.Printf("Failed Preserving Section: Cookie '%s' not found for role '%s'", cookiename, role)
		}
	default:
		ctx = context.WithValue(ctx, ContextCurrentSection, section)
		cookies.SetCookie(w, LastSectionCookieName(role), section)
		skylb.Log.Printf("Setting Section '%s'", section)
	}
	r = r.WithContext(ctx)
	return r
}

func (skylb Skylab) SetRoleSectionHandler(role, section string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = skylb.SetRoleSection(w, r, role, section)
			next.ServeHTTP(w, r)
		})
	}
}

func (skylb Skylab) RedirectToLastSection(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			skylb.Log.TraceRequest(r)
			headers.DoNotCache(w)
			section := cookies.GetCookieValue(r, LastSectionCookieName(role))
			if IsValidSection(section) {
				skylb.Log.Printf("%s's last section was %s, redirecting", role, section)
				http.Redirect(w, r, section, http.StatusMovedPermanently)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

var lastSectionCookieNames = map[string]string{
	RoleStudent: "_student_section",
	RoleAdviser: "_adviser_section",
	RoleMentor:  "_mentor_section",
	RoleAdmin:   "_admin_section",
}

const LastRoleCookieName = "_skylab_role"

func LastSectionCookieName(role string) string {
	return lastSectionCookieNames[role]
}
