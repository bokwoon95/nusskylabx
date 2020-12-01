package skylab

import (
	"context"
	"net/http"

	"github.com/bokwoon95/nusskylabx/helpers/erro"
)

const MultipartMaxSize = 32 << 20

// Contains is general purpose string function that checks if a slice of values
// contains the target string.
func Contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

// ProjectLevel consts correspond to the project levels present inside the
// project_level_enum table in the database
const (
	ProjectLevelVostok  = "vostok"
	ProjectLevelGemini  = "gemini"
	ProjectLevelApollo  = "apollo"
	ProjectLevelArtemis = "artemis"
)

func ProjectLevels() []string {
	return []string{
		ProjectLevelVostok,
		ProjectLevelGemini,
		ProjectLevelApollo,
		ProjectLevelArtemis,
	}
}

// addConstProjectLevel adds ProjectLevel consts to FuncMap
func addConstProjectLevel(funcs map[string]interface{}) map[string]interface{} {
	if funcs == nil {
		funcs = map[string]interface{}{}
	}
	funcs["SkylabProjectLevels"] = func() []string { return ProjectLevels() }
	funcs["ProjectLevelVostok"] = func() string { return ProjectLevelVostok }
	funcs["ProjectLevelGemini"] = func() string { return ProjectLevelGemini }
	funcs["ProjectLevelApollo"] = func() string { return ProjectLevelApollo }
	funcs["ProjectLevelArtemis"] = func() string { return ProjectLevelArtemis }
	return funcs
}

func addProjectLevel(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		data = make(map[string]interface{})
	}
	data["SkylabProjectLevels"] = ProjectLevels()
	data["ProjectLevelVostok"] = ProjectLevelVostok
	data["ProjectLevelGemini"] = ProjectLevelGemini
	data["ProjectLevelApollo"] = ProjectLevelApollo
	data["ProjectLevelArtemis"] = ProjectLevelArtemis
	return data
}

// Role consts correspond to the roles present inside the role_enum table in
// the database
const (
	RoleApplicant = "applicant"
	RoleStudent   = "student"
	RoleAdviser   = "adviser"
	RoleMentor    = "mentor"
	RoleAdmin     = "admin"
	RoleNull      = ""

	// RolePreserve is a special role that does not exist in the database, but
	// is only for facilitating the app
	RolePreserve = "preserve"
)

func Roles() []string {
	return []string{
		RoleApplicant,
		RoleStudent,
		RoleAdviser,
		RoleMentor,
		RoleAdmin,
	}
}

// addConstRole adds Role consts to FuncMap
func addConstRole(funcs map[string]interface{}) map[string]interface{} {
	if funcs == nil {
		funcs = map[string]interface{}{}
	}
	funcs["SkylabRoles"] = func() []string { return Roles() }
	funcs["RoleApplicant"] = func() string { return RoleApplicant }
	funcs["RoleStudent"] = func() string { return RoleStudent }
	funcs["RoleAdviser"] = func() string { return RoleAdviser }
	funcs["RoleMentor"] = func() string { return RoleMentor }
	funcs["RoleAdmin"] = func() string { return RoleAdmin }
	funcs["RoleNull"] = func() string { return RoleNull }
	return funcs
}

func addRole(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		data = map[string]interface{}{}
	}
	data["SkylabRoles"] = Roles()
	data["RoleApplicant"] = RoleApplicant
	data["RoleStudent"] = RoleStudent
	data["RoleAdviser"] = RoleAdviser
	data["RoleMentor"] = RoleMentor
	data["RoleAdmin"] = RoleAdmin
	data["RoleNull"] = RoleNull
	return data
}

// ApplicationStatus consts correspond to the statuses present inside the
// applications_status_enum table in the database
const (
	ApplicationStatusPending  = "pending"
	ApplicationStatusAccepted = "accepted"
	ApplicationStatusDeleted  = "deleted"
)

func ApplicationStatuses() []string {
	return []string{
		ApplicationStatusPending,
		ApplicationStatusAccepted,
		ApplicationStatusDeleted,
	}
}

// addConstApplicationStatus adds ApplicationStatus consts to FuncMap
func addConstApplicationStatus(funcs map[string]interface{}) map[string]interface{} {
	if funcs == nil {
		funcs = map[string]interface{}{}
	}
	funcs["SkylabApplicationStatuses"] = func() []string { return ApplicationStatuses() }
	funcs["ApplicationStatusPending"] = func() string { return ApplicationStatusPending }
	funcs["ApplicationStatusAccepted"] = func() string { return ApplicationStatusAccepted }
	funcs["ApplicationStatusDeleted"] = func() string { return ApplicationStatusDeleted }
	return funcs
}

func addApplicationStatus(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		data = map[string]interface{}{}
	}
	data["SkylabApplicationStatuses"] = ApplicationStatuses()
	data["ApplicationStatusPending"] = ApplicationStatusPending
	data["ApplicationStatusAccepted"] = ApplicationStatusAccepted
	data["ApplicationStatusDeleted"] = ApplicationStatusDeleted
	return data
}

func (skylb Skylab) ValidateCohortStageMilestone(cohort, stage, milestone string) error {
	if !Contains(skylb.Cohorts(), cohort) {
		return erro.Wrap(erro.Errorf(ErrCohortInvalid, cohort))
	}
	if !Contains(Stages(), stage) {
		return erro.Wrap(erro.Errorf(ErrStageInvalid, stage))
	}
	if !Contains(Milestones(), milestone) {
		return erro.Wrap(erro.Errorf(ErrMilestoneInvalid, milestone))
	}
	return nil
}

// TeamStatus consts correspond to the statuses present inside the
// teams_status_enum table in the database
const (
	TeamStatusGood          = "good"
	TeamStatusOk            = "ok"
	TeamStatusUncontactable = "uncontactable"
)

func TeamStatuses() []string {
	return []string{
		TeamStatusGood,
		TeamStatusOk,
		TeamStatusUncontactable,
	}
}

// addConstTeamStatus adds TeamStatus consts to FuncMap
func addConstTeamStatus(funcs map[string]interface{}) map[string]interface{} {
	if funcs == nil {
		funcs = map[string]interface{}{}
	}
	funcs["SkylabTeamStatuses"] = func() []string { return TeamStatuses() }
	funcs["TeamStatusGood"] = func() string { return TeamStatusGood }
	funcs["TeamStatusOk"] = func() string { return TeamStatusOk }
	funcs["TeamStatusUncontactable"] = func() string { return TeamStatusUncontactable }
	return funcs
}

func addTeamStatus(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		data = map[string]interface{}{}
	}
	data["SkylabTeamStatuses"] = TeamStatuses()
	data["TeamStatusGood"] = TeamStatusGood
	data["TeamStatusOk"] = TeamStatusOk
	data["TeamStatusUncontactable"] = TeamStatusUncontactable
	return data
}

// Stage consts correspond to the stages present inside the stage_enum table in
// the database
const (
	StageApplication = "application"
	StageSubmission  = "submission"
	StageEvaluation  = "evaluation"
	StageFeedback    = "feedback"
	StageNull        = ""
)

func Stages() []string {
	return []string{
		StageApplication,
		StageSubmission,
		StageEvaluation,
		StageFeedback,
		StageNull,
	}
}

// addConstStage adds Stage consts to FuncMap
func addConstStage(funcs map[string]interface{}) map[string]interface{} {
	if funcs == nil {
		funcs = map[string]interface{}{}
	}
	funcs["SkylabStages"] = func() []string { return Stages() }
	funcs["StageApplication"] = func() string { return StageApplication }
	funcs["StageSubmission"] = func() string { return StageSubmission }
	funcs["StageEvaluation"] = func() string { return StageEvaluation }
	funcs["StageFeedback"] = func() string { return StageFeedback }
	funcs["StageNull"] = func() string { return StageNull }
	return funcs
}

func addStage(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		data = map[string]interface{}{}
	}
	data["SkylabStages"] = Stages()
	data["StageApplication"] = StageApplication
	data["StageSubmission"] = StageSubmission
	data["StageEvaluation"] = StageEvaluation
	data["StageFeedback"] = StageFeedback
	data["StageNull"] = StageNull
	return data
}

// Milestone consts correspond to the milestones present inside the
// milestone_enum table in the database
const (
	Milestone1    = "milestone1"
	Milestone2    = "milestone2"
	Milestone3    = "milestone3"
	MilestoneNull = ""
)

func Milestones() []string {
	return []string{
		Milestone1,
		Milestone2,
		Milestone3,
		MilestoneNull,
	}
}

// addConstMilestone adds Milestone consts to FuncMap
func addConstMilestone(funcs map[string]interface{}) map[string]interface{} {
	if funcs == nil {
		funcs = map[string]interface{}{}
	}
	funcs["SkylabMilestones"] = func() []string { return Milestones() }
	funcs["Milestone1"] = func() string { return Milestone1 }
	funcs["Milestone2"] = func() string { return Milestone2 }
	funcs["Milestone3"] = func() string { return Milestone3 }
	funcs["MilestoneNull"] = func() string { return MilestoneNull }
	return funcs
}

func addMilestone(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		data = map[string]interface{}{}
	}
	data["SkylabMilestones"] = Milestones()
	data["Milestone1"] = Milestone1
	data["Milestone2"] = Milestone2
	data["Milestone3"] = Milestone3
	data["MilestoneNull"] = MilestoneNull
	return data
}

// ApplicationSubsection consts correspond to the subsection columns inside the
// form_schema table in the database, only applicable to application form
// schemas
const (
	ApplicationSubsectionApplication = "application"
	ApplicationSubsectionApplicant   = "applicant"
)

func ApplicationSubsections() []string {
	return []string{
		ApplicationSubsectionApplication,
		ApplicationSubsectionApplicant,
	}
}

type temptype string

const tempNameBruh temptype = "skylab"

func SetVars(namespace string, ctx context.Context) context.Context {
	data := map[string]interface{}{}
	data = addProjectLevel(data)
	data = addRole(data)
	data = addApplicationStatus(data)
	data = addTeamStatus(data)
	data = addStage(data)
	data = addMilestone(data)
	ctx = context.WithValue(ctx, tempNameBruh, map[string]interface{}{
		namespace: data,
	})
	return ctx
}

func GetVars(ctx context.Context) map[string]interface{} {
	data, ok := ctx.Value(tempNameBruh).(map[string]interface{})
	if !ok {
		return map[string]interface{}{}
	}
	return data
}

func SetVarsHandler(namespace string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := SetVars(namespace, r.Context())
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
