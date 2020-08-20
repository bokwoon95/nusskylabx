package skylab

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/bokwoon95/nusskylabx/helpers/formx"
)

type ApplicationEdit struct {
	ApplicantUserID int
	Application     Application
}

type ApplicationView struct {
	Application Application
}

type FormEdit struct {
	Title            string
	Form             Form
	QuestionsAnswers []formx.QuestionAnswer
	PreviewURL       string
	UpdateURL        string
}

type FormView struct {
	Title            string
	Form             Form
	QuestionsAnswers []formx.QuestionAnswer
	EditURL          string
	UpdateURL        string
}

// SubmissionViewData is the data struct that targets the
// "app/skylab/submission_view.html" template
type SubmissionViewData struct {
	Submission Submission
	EditURL    string
	SubmitURL  string
}

// SubmissionEditData is the data struct that targets the
// "app/skylab/submission_edit.html" template
type SubmissionEditData struct {
	Submission        Submission
	PeerEvaluations   []TeamEvaluation
	AdviserEvaluation UserEvaluation
	MentorEvaluation  UserEvaluation
	PreviewURL        string
	UpdateURL         string
	SubmitURL         string
}

type TeamView struct {
	Team        Team
	UserBaseURL string
}

type TeamEdit struct {
	Team      Team
	UpdateURL string
}

type TeamEvaluationView struct {
	TeamEvaluation TeamEvaluation
	SubmitURL      string
	SubmissionURL  string
	EditURL        string
}

type TeamEvaluationEdit struct {
	TeamEvaluation TeamEvaluation
	UpdateURL      string
	SubmitURL      string
	SubmissionURL  string
	PreviewURL     string
}

type UserView struct {
	User           User
	Team           Team
	AdvisingTeams  []Team
	MentoringTeams []Team
	PreviewURL     string
	UserBaseURL    string
	TeamBaseURL    string
}

type UserEdit struct {
	User           User
	Team           Team
	AdvisingTeams  []Team
	MentoringTeams []Team
	UserBaseURL    string
	TeamBaseURL    string
}

func UserViewFuncs(funcs template.FuncMap, data UserView) template.FuncMap {
	if funcs == nil {
		funcs = make(template.FuncMap)
	}
	funcs["DisplayTeams"] = displayTeams(data.TeamBaseURL)
	return funcs
}

func displayTeams(teamBaseURL string) func([]Team) template.HTML {
	return func(teams []Team) template.HTML {
		var links []string
		for _, team := range teams {
			links = append(links, fmt.Sprintf(
				`<li>[%s]&nbsp;&nbsp;<a href="%s/%d">%s</a></li>`,
				team.ProjectLevel, teamBaseURL, team.TeamID, team.TeamName,
			))
		}
		return template.HTML(`<ul>` + strings.Join(links, "") + `</ul>`)
	}
}

type UserEvaluationEdit struct {
	Evaluation    UserEvaluation
	SubmissionURL string
	SubmitURL     string
	PreviewURL    string
	UpdateURL     string
}

type UserEvaluationView struct {
	Evaluation    UserEvaluation
	SubmissionURL string
	SubmitURL     string
	EditURL       string
}
