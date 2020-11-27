package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bokwoon95/nusskylabx/app/admins"
	"github.com/bokwoon95/nusskylabx/app/advisers"
	"github.com/bokwoon95/nusskylabx/app/applicants"
	"github.com/bokwoon95/nusskylabx/app/mentors"
	"github.com/bokwoon95/nusskylabx/app/skylab"
	"github.com/bokwoon95/nusskylabx/app/students"
	"github.com/bokwoon95/nusskylabx/helpers/auth"
	"github.com/bokwoon95/nusskylabx/helpers/auth/openid"
	"github.com/bokwoon95/nusskylabx/helpers/flash"
	"github.com/bokwoon95/nusskylabx/helpers/headers"
	"github.com/go-chi/chi/middleware"
)

func AllRoutes(skylb skylab.Skylab) {
	// Middleware
	skylb.Mux.Use(middleware.RequestID)    // Insert a unique ID in every request
	skylb.Mux.Use(middleware.StripSlashes) // Remove trailing slashes
	skylb.Mux.Use(middleware.Logger)       // Log all paths hit
	skylb.Mux.Use(middleware.Recoverer)    // Pretty print panic traces
	skylb.Mux.Use(middleware.Compress(-1,  // Compress assets before serving
		"text/html",
		"text/css",
		"text/javascript",
		"application/javascript",
		"application/json",
		"image/svg+xml",
	))
	if skylb.IsProd {
		// Hard limit of 1 minute allowed for every request (only in production)
		skylb.Mux.Use(middleware.Timeout(60 * time.Second))
	}
	skylb.Mux.Use(skylb.AddProdContext)           // Add prod or dev context to request
	skylb.Mux.Use(headers.SecurityHeadersHandler) // Add security related headers to every request
	skylb.Mux.Use(skylab.SetVarsHandler("skylab"))

	Routes(skylb)
	SkylabRoutes(skylb)
	ApplicantRoutes(skylb)
	StudentRoutes(skylb)
	AdviserRoutes(skylb)
	MentorRoutes(skylb)
	AdminRoutes(skylb)
}

func Routes(skylb skylab.Skylab) {
	ap := App{skylb: skylb}

	sessionMux := skylb.Mux.With(skylb.GetSession)

	// /
	sessionMux.Get("/", ap.Landing)

	// /login-page
	sessionMux.Get("/login-page", ap.LoginPage)

	// /login
	callbackURL := skylb.BaseURLWithProtocol() + "/login/callback"
	sessionMux.With(skylb.ChooseProvider).Get("/login", func(w http.ResponseWriter, r *http.Request) {
		provider := r.FormValue("provider")
		if provider == "" {
			skylb.BadRequest(w, r, "provider cannot be blank")
			return
		}
		auth.Redirect(w, r, provider, callbackURL, skylb.InternalServerError)
	})

	// /login/callback
	sessionMux.With(
		auth.Authenticate(callbackURL, skylb.InternalServerError),
		skylb.EnsureIsUser,
		skylb.SetSession,
	).Get("/login/callback", skylb.RedirectUserrole)

	// /logout
	sessionMux.With(
		skylb.RevokeSession,
		skylb.GetSession,
	).HandleFunc("/logout", skylb.RedirectAfterLogout)

	// /showcase/{cohort}/{projectlevel}
	sessionMux.Get(`/showcase`, ap.PastYearShowcase)
	sessionMux.Get(`/showcase/{cohort}`, ap.PastYearShowcase)
	sessionMux.Get(`/showcase/{cohort}/{projectlevel}`, ap.PastYearShowcase)

	// /user
	sessionMux.Get("/user", ap.User)

	// /user/update/{userID}
	sessionMux.Post("/user/update/{userID}", ap.UserUpdate)
}

func SkylabRoutes(skylb skylab.Skylab) {
	skylb.Mux.With(skylb.GetSession).NotFound(skylb.NotFound)
	skylb.Mux.With(skylb.GetSession).MethodNotAllowed(skylb.MethodNotAllowed)
}

func ApplicantRoutes(skylb skylab.Skylab) {
	apt := applicants.New(skylb)

	skylb.Mux.With(skylb.GetSession).Get("/applicant/application/closed", apt.Closed)
	flashMessager := flash.NewEncoder(skylb.SecretKey)

	// Ensures user is an applicant before passing through
	applicantsMux := skylb.Mux.With(skylb.EnsureRole(skylab.RoleApplicant), apt.CheckIfOpen)

	// /applicant/application/join
	skylb.Mux.With(
		skylb.GetSession,
		apt.CheckIfOpen,
		apt.MagicstringVerifier,
		apt.JoinIfLoggedin,
	).HandleFunc("/applicant/application/join", openid.Redirect(
		openid.ProviderNUS,
		skylb.BaseURLWithProtocol()+"/applicant/application/join/callback",
		skylb.InternalServerError,
	))

	// /applicant/application/join/callback
	skylb.Mux.With(
		openid.Authenticate(openid.ProviderNUS, skylb.InternalServerError),
		apt.CheckIfOpen,
		apt.MagicstringVerifier,
		apt.JoinApplication,
		skylb.SetSession,
	).Get("/applicant/application/join/callback", skylb.Redirect("/applicant/application"))

	// /applicant/login
	skylb.Mux.With(
		apt.CheckIfOpen,
	).Get("/applicant/login", openid.Redirect(
		openid.ProviderNUS,
		skylb.BaseURLWithProtocol()+"/applicant/login/callback",
		skylb.InternalServerError,
	))

	// /applicant/login/callback
	skylb.Mux.With(
		openid.Authenticate(openid.ProviderNUS, skylb.InternalServerError),
		apt.CheckIfOpen,
		apt.IdempotentCreateApplicant,
		skylb.SetSession,
	).Get("/applicant/login/callback", skylb.Redirect("/applicant"))

	// /applicant
	applicantsMux.Get("/applicant", apt.Applicant)

	// /applicant/application
	applicantsMux.Get("/applicant/application", apt.Application)

	// /applucant/application/create
	applicantsMux.With(
		apt.IdempotentCreateApplication,
	).Post("/applicant/application/create", skylb.Redirect("/applicant/application"))

	// /applucant/application/leave
	applicantsMux.With(
		apt.LeaveApplication,
	).Post("/applicant/application/leave", skylb.Redirect("/applicant"))

	// /applucant/application/update
	applicantsMux.With(
		apt.UpdateApplication,
	).Post("/applicant/application/update", skylb.Redirect("/applicant/application"))

	// /applucant/application/submit
	applicantsMux.With(
		apt.UpdateApplication,
		flashMessager.UnsetFlashMsgsHandler(flash.Success),
		apt.SubmitApplication,
	).Post("/applicant/application/submit", skylb.Redirect("/applicant/application"))
}

func StudentRoutes(skylb skylab.Skylab) {
	stu := students.New(skylb)
	flashMessager := flash.NewEncoder(skylb.SecretKey)

	// studentsMux ensures user is a student before passing through
	studentsMux := skylb.Mux.With(skylb.EnsureRole(skylab.RoleStudent))

	// Redirects to /student/submission/{submissionID}
	redirectSubmissionView := skylb.Redirect(skylab.StudentSubmission + `/{submissionID}`)

	// Redirects to /student/submission/{submissionID}/edit
	redirectSubmissionEdit := skylb.Redirect(skylab.StudentSubmission + `/{submissionID}/edit`)

	// Redirects to /student/team-evaluation/{teamEvaluationID}
	redirectEvaluationView := skylb.Redirect(skylab.StudentTeamEvaluation + "/{teamEvaluationID}")

	// Redirects to /student/team-evaluation/{teamEvaluationID}/edit
	redirectEvaluationEdit := skylb.Redirect(skylab.StudentTeamEvaluation + "/{teamEvaluationID}/edit")

	// /student
	studentsMux.With(
		skylb.RedirectToLastSection(skylab.RoleStudent),
	).Get("/student", stu.Dashboard)

	// /student/dashboard
	studentsMux.Get(skylab.StudentDashboard, stu.Dashboard)

	// /student/teams
	studentsMux.Get(skylab.StudentTeam, stu.Team)

	// /student/submission/{submissionID}
	studentsMux.With(
		stu.CanViewSubmission,
	).Get(skylab.StudentSubmission+`/{submissionID:\d+}`, skylb.SubmissionView(skylab.RoleStudent))

	// /student/submission/{submissionID}/edit
	studentsMux.With(
		stu.CanEditSubmission,
	).Get(skylab.StudentSubmission+`/{submissionID:\d+}/edit`, skylb.SubmissionEdit(skylab.RoleStudent))

	// /student/milestone1/submission
	studentsMux.With(
		addMilestoneCtxHandler(skylab.Milestone1),
		stu.IdempotentSubmissionCreate,
	).Get(skylab.StudentM1Submission, redirectSubmissionEdit)

	// /student/milestone2/submission
	studentsMux.With(
		addMilestoneCtxHandler(skylab.Milestone2),
		stu.IdempotentSubmissionCreate,
	).Get(skylab.StudentM2Submission, redirectSubmissionEdit)

	// /student/milestone3/submission
	studentsMux.With(
		addMilestoneCtxHandler(skylab.Milestone3),
		stu.IdempotentSubmissionCreate,
	).Get(skylab.StudentM3Submission, redirectSubmissionEdit)

	// /student/submission/create
	studentsMux.With(
		stu.IdempotentSubmissionCreate,
	).HandleFunc(skylab.StudentSubmission+"/create", redirectSubmissionEdit)

	// /student/submission/{submissionID}/preview
	studentsMux.With(
		stu.CanEditSubmission,
		stu.SubmissionUpdate,
		flashMessager.UnsetFlashMsgsHandler(flash.Success),
	).Post(skylab.StudentSubmission+`/{submissionID:\d+}/preview`, redirectSubmissionView)

	// /student/submission/{submissionID}/update
	studentsMux.With(
		stu.CanEditSubmission,
		stu.SubmissionUpdate,
	).Post(skylab.StudentSubmission+`/{submissionID:\d+}/update`, redirectSubmissionEdit)

	// /student/submission/{submissionID}/submit
	studentsMux.With(
		stu.CanEditSubmission,
		stu.SubmissionUpdate,
		flashMessager.UnsetFlashMsgsHandler(flash.Success),
		stu.SubmissionSubmit,
	).Post(skylab.StudentSubmission+`/{submissionID:\d+}/submit`, func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("view") == "true" {
			redirectSubmissionView(w, r)
			return
		}
		redirectSubmissionEdit(w, r)
	})

	// /student/milestone1/evaluation
	studentsMux.Get(skylab.StudentM1Evaluation, stu.MilestoneTeamEvaluation(skylab.StudentM1Evaluation))

	// /student/milestone2/evaluation
	studentsMux.Get(skylab.StudentM2Evaluation, stu.MilestoneTeamEvaluation(skylab.StudentM2Evaluation))

	// /student/milestone3/evaluation
	studentsMux.Get(skylab.StudentM3Evaluation, stu.MilestoneTeamEvaluation(skylab.StudentM3Evaluation))

	// /student/team-evaluation/create
	studentsMux.Post(skylab.StudentTeamEvaluation+"/create", stu.TeamEvaluationCreate)

	// /student/user-evaluation/{userEvaluationID}
	studentsMux.With(
		stu.CanViewUserEvaluation,
	).Get(skylab.StudentUserEvaluation+`/{userEvaluationID:\d+}`, skylb.UserEvaluationView(skylab.RoleStudent))

	// /student/team-evaluation/{teamEvaluationID}
	studentsMux.With(
		stu.CanViewTeamEvaluation,
	).Get(skylab.StudentTeamEvaluation+`/{teamEvaluationID:\d+}`, skylb.TeamEvaluationView(skylab.RoleStudent))

	// /student/team-evaluation/{teamEvaluationID}/edit
	studentsMux.With(
		stu.CanEditTeamEvaluation,
	).Get(skylab.StudentTeamEvaluation+`/{teamEvaluationID:\d+}/edit`, stu.TeamEvaluationEdit)

	// /student/team-evaluation/{teamEvaluationID}/preview
	studentsMux.With(
		stu.CanEditTeamEvaluation,
		stu.TeamEvaluationUpdate,
		flashMessager.UnsetFlashMsgsHandler(flash.Success),
	).Post(skylab.StudentTeamEvaluation+`/{teamEvaluationID:\d+}/preview`, redirectEvaluationView)

	// /student/team-evaluation/{teamEvaluationID}/update
	studentsMux.With(
		stu.CanEditTeamEvaluation,
		stu.TeamEvaluationUpdate,
	).Post(skylab.StudentTeamEvaluation+`/{teamEvaluationID:\d+}/update`, redirectEvaluationEdit)

	// /student/team-evaluation/{teamEvaluationID}/submit
	studentsMux.With(
		stu.CanEditTeamEvaluation,
		stu.TeamEvaluationUpdate,
		flashMessager.UnsetFlashMsgsHandler(flash.Success),
		stu.TeamEvaluationSubmit,
	).Post(skylab.StudentTeamEvaluation+`/{teamEvaluationID:\d+}/submit`, redirectEvaluationEdit)

	// /student/feedbacks
	studentsMux.Get(skylab.StudentListFeedbacks, stu.ListFeedbacks)

	// /student/feedback/team/{feedbackIDOnTeam}/edit
	studentsMux.With(
		stu.CanEditTeamFeedback,
	).Get(skylab.StudentTeamFeedback+`/{feedbackIDOnTeam:\d+}/edit`, stu.TeamFeedbackEdit)
}

func addMilestoneCtxHandler(milestone string) func(http.Handler) http.Handler {
	if !skylab.Contains(skylab.Milestones(), milestone) {
		panic(fmt.Sprintf("milestone %s is invalid", milestone))
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), skylab.ContextCurrentMilestone, milestone))
			next.ServeHTTP(w, r)
		})
	}
}

func AdviserRoutes(skylb skylab.Skylab) {
	adv := advisers.New(skylb)
	flashMessager := flash.NewEncoder(skylb.SecretKey)

	advisersMux := skylb.Mux.With(skylb.EnsureRole(skylab.RoleAdviser))
	advisersMux.With(skylb.RedirectToLastSection(skylab.RoleAdviser)).Get("/adviser", adv.Dashboard)

	// Redirects to /adviser/user-evaluation/{userEvaluationID}
	redirectEvaluationView := skylb.Redirect(skylab.AdviserUserEvaluation + "/{userEvaluationID}")

	// Redirects to /adviser/user-evaluation/{userEvaluationID}/edit
	redirectEvaluationEdit := skylb.Redirect(skylab.AdviserUserEvaluation + "/{userEvaluationID}/edit")

	// /adviser/dashboard
	advisersMux.Get(skylab.AdviserDashboard, adv.Dashboard)

	// /adviser/teams
	advisersMux.Get(skylab.AdviserTeams, adv.Teams)

	// /adviser/evaluatee-evaluators
	advisersMux.Get(skylab.AdviserEvaluateeEvaluators, adv.EvaluateeEvaluators)

	// /adviser/evaluatee-evaluators/update
	advisersMux.Post(skylab.AdviserEvaluateeEvaluators+"/update", adv.EvaluateeEvaluatorsUpdate)

	// /adviser/evaluator-evaluatees
	advisersMux.Get(skylab.AdviserEvaluatorEvaluatees, adv.EvaluatorEvaluatees)

	// /adviser/evaluator-evaluatees/update
	advisersMux.Post(skylab.AdviserEvaluatorEvaluatees+"/update", adv.EvaluatorEvaluateesUpdate)

	// /adviser/milestone1/evaluation
	advisersMux.Get(skylab.AdviserM1MakeEvaluation, adv.MilestoneUserEvaluation(skylab.AdviserM1MakeEvaluation))

	// /adviser/milestone2/evaluation
	advisersMux.Get(skylab.AdviserM2MakeEvaluation, adv.MilestoneUserEvaluation(skylab.AdviserM2MakeEvaluation))

	// /adviser/milestone3/evaluation
	advisersMux.Get(skylab.AdviserM3MakeEvaluation, adv.MilestoneUserEvaluation(skylab.AdviserM3MakeEvaluation))

	// /adviser/user-evaluation/create
	advisersMux.With(
		adv.UserEvaluationCreate,
	).Post(skylab.AdviserUserEvaluation+"/create", redirectEvaluationEdit)

	// /adviser/submission/{submissionID}
	advisersMux.With(
		adv.CanViewSubmission,
	).Get(skylab.AdviserSubmission+`/{submissionID:\d+}`, skylb.SubmissionView(skylab.RoleAdviser))

	// /adviser/user-evaluation/{userEvaluationID}
	advisersMux.With(
		adv.CanViewUserEvaluation,
	).Get(skylab.AdviserUserEvaluation+`/{userEvaluationID:\d+}`, skylb.UserEvaluationView(skylab.RoleAdviser))

	// /adviser/user-evaluation/{userEvaluationID}/edit
	advisersMux.With(
		adv.CanEditUserEvaluation,
	).Get(skylab.AdviserUserEvaluation+`/{userEvaluationID:\d+}/edit`, skylb.UserEvaluationEdit(skylab.RoleAdviser))

	// /adviser/user-evaluation/{userEvaluationID}/preview
	advisersMux.With(
		adv.CanEditUserEvaluation,
		adv.UserEvaluationUpdate,
		flashMessager.UnsetFlashMsgsHandler(flash.Success),
	).Post(skylab.AdviserUserEvaluation+`/{userEvaluationID:\d+}/preview`, redirectEvaluationView)

	// /adviser/user-evaluation/{userEvaluationID}/update
	advisersMux.With(
		adv.CanEditUserEvaluation,
		adv.UserEvaluationUpdate,
	).Post(skylab.AdviserUserEvaluation+`/{userEvaluationID:\d+}/update`, redirectEvaluationEdit)

	// /adviser/user-evaluation/{userEvaluationID}/submit
	advisersMux.With(
		adv.CanEditUserEvaluation,
		adv.UserEvaluationUpdate,
		flashMessager.UnsetFlashMsgsHandler(flash.Success),
		adv.UserEvaluationSubmit,
	).Post(skylab.AdviserUserEvaluation+`/{userEvaluationID:\d+}/submit`, redirectEvaluationEdit)

	// /adviser/milestone1/evaluations
	advisersMux.Get(skylab.AdviserM1ViewEvaluation, adv.MilestoneTeamEvaluation(skylab.AdviserM1ViewEvaluation))

	// /adviser/milestone2/evaluations
	advisersMux.Get(skylab.AdviserM2ViewEvaluation, adv.MilestoneTeamEvaluation(skylab.AdviserM2ViewEvaluation))

	// /adviser/milestone3/evaluations
	advisersMux.Get(skylab.AdviserM3ViewEvaluation, adv.MilestoneTeamEvaluation(skylab.AdviserM3ViewEvaluation))

	// /adviser/team-evaluation/{teamEvaluationID}
	advisersMux.With(
		adv.CanViewTeamEvaluation,
	).Get(skylab.AdviserTeamEvaluation+`/{teamEvaluationID:\d+}`, skylb.TeamEvaluationView(skylab.RoleAdviser))
}

func MentorRoutes(skylb skylab.Skylab) {
	mnt := mentors.New(skylb)

	// mentorsMux ensures all users who access the routes are Advisers
	mentorsMux := skylb.Mux.With(
		skylb.EnsureRole(skylab.RoleMentor),
	)

	// Mentors //
	mentorsMux.With(
		skylb.RedirectToLastSection(skylab.RoleMentor),
	).Get("/mentor", mnt.Dashboard)

	// Dashboard //
	mentorsMux.Get(skylab.MentorDashboard, mnt.Dashboard)
}

func AdminRoutes(skylb skylab.Skylab) {
	adm := admins.New(skylb)

	adminsMux := skylb.Mux.With(skylb.GetSession, skylb.AllowIfDevelopment)
	adminsMux.With(skylb.RedirectToLastSection(skylab.RoleAdmin)).Get("/admin", adm.Dashboard)
	flashMessager := flash.NewEncoder(skylb.SecretKey)

	// /admin/dashboard
	adminsMux.Get(skylab.AdminDashboard, adm.Dashboard)

	// /admin/create-user
	adminsMux.Get(skylab.AdminCreateUser, adm.CreateUser)

	// /admin/create-user/confirm
	adminsMux.HandleFunc(skylab.AdminCreateUserConfirm, adm.CreateUserConfirm)

	// /admin/create-user/confirm/post
	adminsMux.With(
		adm.CreateUserConfirmPost,
	).Post(skylab.AdminCreateUserConfirm+"/post", skylb.Redirect(skylab.AdminCreateUser))

	// /admin/cohorts
	adminsMux.Get(skylab.AdminListCohorts, adm.ListCohorts)

	// /admin/cohorts/create
	adminsMux.With(
		adm.ListCohortsCreate,
	).HandleFunc(skylab.AdminListCohorts+`/{cohort}/create`, skylb.Redirect(skylab.AdminListCohorts))

	// /admin/cohorts/delete
	adminsMux.With(
		adm.ListCohortsDelete,
	).Post(skylab.AdminListCohorts+`/delete`, skylb.Redirect(skylab.AdminListCohorts))

	// /admin/cohorts/refresh
	adminsMux.With(
		adm.ListCohortsRefresh,
	).Post(skylab.AdminListCohorts+`/refresh`, skylb.Redirect(skylab.AdminListCohorts))

	// /admin/users/{cohort}/{role}
	adminsMux.Get(skylab.AdminListUsers, adm.ListUsers)
	adminsMux.Get(skylab.AdminListUsers+`/{cohort}`, adm.ListUsers)
	adminsMux.Get(skylab.AdminListUsers+`/{cohort}/{role}`, adm.ListUsers)

	// /admin/user/{userID}
	adminsMux.Get(skylab.AdminUser+`/{userID:\d+}`, adm.UserView)

	// /admin/user/{userID}/preview
	adminsMux.With(
		adm.UserPreviewAs,
	).Post(skylab.AdminUser+`/{userID:\d+}/preview`, skylb.Redirect(skylab.AdminUser+`/{userID:\d+}`))

	// /admin/periods/{cohort}
	adminsMux.Get(skylab.AdminListPeriods, adm.ListPeriods)
	adminsMux.Get(skylab.AdminListPeriods+`/{cohort}`, adm.ListPeriods)

	// /admin/periods/create
	adminsMux.With(
		adm.ListPeriodsCreate,
	).Post(skylab.AdminListPeriods+`/create`, skylb.Redirect(skylab.AdminListPeriods+`/{cohort}`))

	// /admin/periods/delete
	adminsMux.With(
		adm.ListPeriodsDelete,
	).Post(skylab.AdminListPeriods+`/delete`, skylb.Redirect(skylab.AdminListPeriods))

	// /admin/periods/duplicate
	adminsMux.With(
		adm.ListPeriodsDuplicate,
	).Post(skylab.AdminListPeriods+`/duplicate`, skylb.Redirect(skylab.AdminListPeriods+`/{cohort}`))

	// /admin/forms/{cohort}
	adminsMux.Get(skylab.AdminListForms, adm.ListForms)
	adminsMux.Get(skylab.AdminListForms+`/{cohort}`, adm.ListForms)

	// /admin/forms/create
	adminsMux.With(
		adm.ListFormsCreate,
	).Post(skylab.AdminListForms+`/create`, skylb.Redirect(skylab.AdminForm+`/{formID}/edit`))

	// /admin/forms/duplicate
	adminsMux.With(
		adm.ListFormsDuplicate,
	).Post(skylab.AdminListForms+"/duplicate", skylb.Redirect(skylab.AdminListForms+`/{cohort}`))

	// /admin/forms/delete
	adminsMux.With(
		adm.ListFormsDelete,
	).Post(skylab.AdminListForms+"/delete", skylb.Redirect(skylab.AdminListForms))

	// /admin/form/{formID}
	adminsMux.Get(skylab.AdminForm+`/{formID:\d+}`, adm.FormView)

	// /admin/form/{formID}/edit
	adminsMux.Get(skylab.AdminForm+`/{formID:\d+}/edit`, adm.FormEdit)

	// /admin/form/{formID}/update
	adminsMux.With(
		adm.FormUpdate,
	).Post(skylab.AdminForm+`/{formID:\d+}/update`, skylb.Redirect(skylab.AdminForm+`/{formID}/edit`))

	// /admin/form/{formID}/preview
	adminsMux.With(
		adm.FormUpdate,
		flashMessager.UnsetFlashMsgsHandler(flash.Success),
	).Post(skylab.AdminForm+`/{formID:\d+}/preview`, skylb.Redirect(skylab.AdminForm+`/{formID}`))

	// /admin/teams/{cohort}
	adminsMux.Get(skylab.AdminListTeams, adm.ListTeams)
	adminsMux.Get(skylab.AdminListTeams+`/{cohort}`, adm.ListTeams)

	// /admin/team/{teamID}
	adminsMux.Get(skylab.AdminTeam+`/{teamID:\d+}`, adm.TeamView)

	// /admin/applications/{cohort}
	adminsMux.Get(skylab.AdminListApplications, adm.ListApplications)
	adminsMux.Get(skylab.AdminListApplications+`/{cohort}`, adm.ListApplications)

	// AdminApplication //
	// TODO?

	// /admin/feedbacks
	adminsMux.Get(skylab.AdminListFeedbacks, adm.ListFeedbacks)

	// /admin/dump-json
	adminsMux.Get(skylab.AdminDumpJson, adm.DumpJson)

	// /admin/dump-json/url
	adminsMux.Get(skylab.AdminDumpJson+"/url", adm.DumpJsonPost)

	// /admin/testmail
	adminsMux.Get(skylab.AdminTestmail, adm.Testmail)
	adminsMux.Post(skylab.AdminTestmail, adm.TestmailPost)
}
