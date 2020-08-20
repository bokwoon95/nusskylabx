package skylab

import (
	"database/sql"
	"fmt"
	"strings"

	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/helpers/formx"
	"github.com/bokwoon95/nusskylabx/tables"
)

type User struct {
	Valid       bool           `db:"-"`
	UserID      int            `db:"user_id"`
	Displayname string         `db:"displayname"`
	Email       string         `db:"email"`
	Roles       map[string]int `db:"-" json:"Roles"` // map of the user's role to the user_role_id
}

type Period struct {
	Valid     bool         `db:"-"`
	PeriodID  int          `db:"period_id"`
	Cohort    string       `db:"cohort"`
	Stage     string       `db:"stage"`
	Milestone string       `db:"milestone"`
	StartAt   sql.NullTime `db:"start_at"`
	EndAt     sql.NullTime `db:"end_at"`
}

type Team struct {
	Valid        bool
	TeamID       int
	Cohort       string
	TeamName     string
	ProjectLevel string
	Status       string
	Student1     User
	Student2     User
	Adviser      User
	Mentor       User
}

func (t *Team) RowMapper(tbl tables.VIEW_V_TEAMS) func(*sq.Row) {
	return func(row *sq.Row) {
		// Team
		t.Valid = row.IntValid(tbl.TEAM_ID)
		t.TeamID = row.Int(tbl.TEAM_ID)
		t.Cohort = row.String(tbl.COHORT)
		t.TeamName = row.String(tbl.TEAM_NAME)
		t.ProjectLevel = row.String(tbl.PROJECT_LEVEL)
		t.Status = row.String(tbl.STATUS)
		// Student1
		t.Student1.Valid = row.IntValid(tbl.STUDENT1_USER_ID)
		t.Student1.UserID = row.Int(tbl.STUDENT1_USER_ID)
		t.Student1.Displayname = row.String(tbl.STUDENT1_DISPLAYNAME)
		// Student2
		t.Student2.Valid = row.IntValid(tbl.STUDENT2_USER_ID)
		t.Student2.UserID = row.Int(tbl.STUDENT2_USER_ID)
		t.Student2.Displayname = row.String(tbl.STUDENT2_DISPLAYNAME)
		// Adviser
		t.Adviser.Valid = row.IntValid(tbl.ADVISER_USER_ID)
		t.Adviser.UserID = row.Int(tbl.ADVISER_USER_ID)
		t.Adviser.Displayname = row.String(tbl.ADVISER_DISPLAYNAME)
		// Mentor
		t.Mentor.Valid = row.IntValid(tbl.MENTOR_USER_ID)
		t.Mentor.UserID = row.Int(tbl.MENTOR_USER_ID)
		t.Mentor.Displayname = row.String(tbl.MENTOR_DISPLAYNAME)
	}
}

type Form struct {
	Valid      bool
	FormID     int
	Period     Period
	Name       string
	Subsection string
	Questions  formx.Questions
}

func (fs Form) Title() string {
	var title string
	if !fs.Valid {
		return ""
	}
	cohort := fs.Period.Cohort
	stage := strings.Title(fs.Period.Stage)
	milestone := strings.Title(fs.Period.Milestone)
	if cohort != "" {
		title = fmt.Sprintf("[%s] ", cohort)
	}
	switch {
	case stage != "" && milestone != "":
		title += fmt.Sprintf("%s Form for %s", stage, milestone)
	case stage != "" && milestone == "":
		title += fmt.Sprintf("%s Form", stage)
	case stage == "" && milestone != "":
		title += fmt.Sprintf("Form for %s", milestone)
	case stage == "" && milestone == "":
		title += "Ad Hoc Form"
	}
	switch {
	case fs.Name != "" && fs.Subsection != "":
		title += fmt.Sprintf(" (name: %s, subsection: %s)", fs.Name, fs.Subsection)
	case fs.Name != "" && fs.Subsection == "":
		title += fmt.Sprintf(" (name: %s)", fs.Name)
	case fs.Name == "" && fs.Subsection != "":
		title += fmt.Sprintf(" (subsection: %s)", fs.Subsection)
	}
	return title
}

type Application struct {
	Valid              bool
	ApplicationID      int
	Cohort             string
	ProjectLevel       string
	Status             string
	Submitted          bool
	Magicstring        sql.NullString
	Applicant1         User
	Applicant2         User
	ApplicationForm    Form
	ApplicationAnswers formx.Answers
	ApplicantForm      Form
	Applicant1Answers  formx.Answers
	Applicant2Answers  formx.Answers
}

func (a *Application) RowMapper(tbl tables.VIEW_V_APPLICATIONS) func(*sq.Row) {
	return func(row *sq.Row) {
		*a = Application{
			Valid:         row.IntValid(tbl.APPLICATION_ID),
			ApplicationID: row.Int(tbl.APPLICATION_ID),
			Cohort:        row.String(tbl.COHORT),
			Status:        row.String(tbl.STATUS),
			ProjectLevel:  row.String(tbl.PROJECT_LEVEL),
			Magicstring:   row.NullString(tbl.MAGICSTRING),
			Submitted:     row.Bool(tbl.SUBMITTED),
			Applicant1: User{
				Valid:       row.IntValid(tbl.APPLICANT1_USER_ID),
				UserID:      row.Int(tbl.APPLICANT1_USER_ID),
				Displayname: row.String(tbl.APPLICANT1_DISPLAYNAME),
				Email:       row.String(tbl.APPLICANT1_EMAIL),
				Roles:       map[string]int{RoleApplicant: row.Int(tbl.APPLICANT1_USER_ROLE_ID)},
			},
			Applicant2: User{
				Valid:       row.IntValid(tbl.APPLICANT2_USER_ID),
				UserID:      row.Int(tbl.APPLICANT2_USER_ID),
				Displayname: row.String(tbl.APPLICANT2_DISPLAYNAME),
				Email:       row.String(tbl.APPLICANT2_EMAIL),
				Roles:       map[string]int{RoleApplicant: row.Int(tbl.APPLICANT2_USER_ROLE_ID)},
			},
			ApplicantForm: Form{
				Valid:  row.IntValid(tbl.APPLICANT_FORM_ID),
				FormID: row.Int(tbl.APPLICANT_FORM_ID),
			},
			ApplicationForm: Form{
				Valid:  row.IntValid(tbl.APPLICATION_FORM_ID),
				FormID: row.Int(tbl.APPLICATION_FORM_ID),
			},
		}
		row.ScanInto(&a.ApplicantForm.Questions, tbl.APPLICANT_QUESTIONS)
		row.ScanInto(&a.Applicant1Answers, tbl.APPLICANT1_ANSWERS)
		row.ScanInto(&a.Applicant2Answers, tbl.APPLICANT2_ANSWERS)
		row.ScanInto(&a.ApplicationForm.Questions, tbl.APPLICATION_QUESTIONS)
		row.ScanInto(&a.ApplicationAnswers, tbl.APPLICATION_ANSWERS)
	}
}

type Submission struct {
	Valid             bool
	SubmissionID      int
	Team              Team
	SubmissionForm    Form
	SubmissionAnswers formx.Answers
	OverrideOpen      bool
	Submitted         bool
	UpdatedAt         sql.NullTime
}

func (s *Submission) RowMapper(tbl tables.VIEW_V_SUBMISSIONS) func(*sq.Row) {
	return func(row *sq.Row) {
		*s = Submission{
			Valid:        row.IntValid(tbl.SUBMISSION_ID),
			SubmissionID: row.Int(tbl.SUBMISSION_ID),
			Team: Team{
				Valid:        row.IntValid(tbl.TEAM_ID),
				TeamID:       row.Int(tbl.TEAM_ID),
				TeamName:     row.String(tbl.TEAM_NAME),
				ProjectLevel: row.String(tbl.PROJECT_LEVEL),
			},
			SubmissionForm: Form{
				Valid:  row.IntValid(tbl.SUBMISSION_FORM_ID),
				FormID: row.Int(tbl.SUBMISSION_FORM_ID),
				Period: Period{
					Valid:     row.StringValid(tbl.COHORT),
					Cohort:    row.String(tbl.COHORT),
					Stage:     StageSubmission,
					Milestone: row.String(tbl.MILESTONE),
					StartAt:   row.NullTime(tbl.START_AT),
					EndAt:     row.NullTime(tbl.END_AT),
				},
			},
			Submitted:    row.Bool(tbl.SUBMITTED),
			OverrideOpen: row.Bool(tbl.OVERRIDE_OPEN),
			UpdatedAt:    row.NullTime(tbl.UPDATED_AT),
		}
		row.ScanInto(&s.SubmissionForm.Questions, tbl.QUESTIONS)
		row.ScanInto(&s.SubmissionAnswers, tbl.ANSWERS)
	}
}

// A TeamEvaluation is carried out by a Team on a Team
type TeamEvaluation struct {
	Valid             bool
	TeamEvaluationID  int
	Evaluator         Team
	Evaluatee         Submission
	EvaluationForm    Form
	EvaluationAnswers formx.Answers
	OverrideOpen      bool
	Submitted         bool
	UpdatedAt         sql.NullTime
}

func (e *TeamEvaluation) RowMapper(tbl tables.VIEW_V_TEAM_EVALUATIONS) func(*sq.Row) {
	return func(row *sq.Row) {
		*e = TeamEvaluation{
			Valid:            row.IntValid(tbl.TEAM_EVALUATION_ID),
			TeamEvaluationID: row.Int(tbl.TEAM_EVALUATION_ID),
			Evaluator: Team{
				Valid:        row.IntValid(tbl.EVALUATOR_TEAM_ID),
				TeamID:       row.Int(tbl.EVALUATEE_TEAM_ID),
				TeamName:     row.String(tbl.EVALUATOR_TEAM_NAME),
				ProjectLevel: row.String(tbl.EVALUATOR_PROJECT_LEVEL),
			},
			EvaluationForm: Form{
				Valid:  row.IntValid(tbl.EVALUATION_FORM_ID),
				FormID: row.Int(tbl.EVALUATION_FORM_ID),
				Period: Period{
					Valid:     true,
					Cohort:    row.String(tbl.COHORT),
					Stage:     row.String(tbl.STAGE),
					Milestone: row.String(tbl.MILESTONE),
					StartAt:   row.NullTime(tbl.EVALUATION_START_AT),
					EndAt:     row.NullTime(tbl.EVALUATION_END_AT),
				},
			},
			Evaluatee: Submission{
				Valid:        row.IntValid(tbl.SUBMISSION_ID),
				SubmissionID: row.Int(tbl.SUBMISSION_ID),
				Team: Team{
					Valid:        row.IntValid(tbl.EVALUATEE_TEAM_ID),
					TeamID:       row.Int(tbl.EVALUATEE_TEAM_ID),
					TeamName:     row.String(tbl.EVALUATEE_TEAM_NAME),
					ProjectLevel: row.String(tbl.EVALUATEE_PROJECT_LEVEL),
				},
				SubmissionForm: Form{
					Valid:  row.IntValid(tbl.SUBMISSION_FORM_ID),
					FormID: row.Int(tbl.SUBMISSION_FORM_ID),
					Period: Period{
						Valid:   true,
						StartAt: row.NullTime(tbl.SUBMISSION_START_AT),
						EndAt:   row.NullTime(tbl.SUBMISSION_END_AT),
					},
				},
				OverrideOpen: row.Bool(tbl.SUBMISSION_OVERRIDE_OPEN),
				Submitted:    row.Bool(tbl.SUBMISSION_SUBMITTED),
				UpdatedAt:    row.NullTime(tbl.SUBMISSION_UPDATED_AT),
			},
			OverrideOpen: row.Bool(tbl.EVALUATION_OVERRIDE_OPEN),
			Submitted:    row.Bool(tbl.EVALUATION_SUBMITTED),
			UpdatedAt:    row.NullTime(tbl.EVALUATION_UPDATED_AT),
		}
		row.ScanInto(&e.EvaluationForm.Questions, tbl.EVALUATION_QUESTIONS)
		row.ScanInto(&e.EvaluationAnswers, tbl.EVALUATION_ANSWERS)
		row.ScanInto(&e.Evaluatee.SubmissionForm.Questions, tbl.SUBMISSION_QUESTIONS)
		row.ScanInto(&e.Evaluatee.SubmissionAnswers, tbl.SUBMISSION_ANSWERS)
	}
}

// An UserEvaluation is carried out by a User on a Team
type UserEvaluation struct {
	Valid             bool
	UserEvaluationID  int
	Evaluator         User
	Role              string
	Evaluatee         Submission
	EvaluationAnswers formx.Answers
	EvaluationForm    Form
	Submitted         bool
	OverrideOpen      bool
	UpdatedAt         sql.NullTime
}

func (e *UserEvaluation) RowMapper(tbl tables.VIEW_V_USER_EVALUATIONS) func(*sq.Row) {
	return func(row *sq.Row) {
		*e = UserEvaluation{
			Valid:            row.IntValid(tbl.USER_EVALUATION_ID),
			UserEvaluationID: row.Int(tbl.USER_EVALUATION_ID),
			Role:             row.String(tbl.EVALUATOR_ROLE),
			Evaluator: User{
				Valid:       row.IntValid(tbl.EVALUATOR_USER_ID),
				UserID:      row.Int(tbl.EVALUATOR_USER_ID),
				Displayname: row.String(tbl.EVALUATOR_DISPLAYNAME),
				Roles:       map[string]int{row.String(tbl.EVALUATOR_ROLE): row.Int(tbl.EVALUATOR_USER_ROLE_ID)},
			},
			EvaluationForm: Form{
				Valid:  row.IntValid(tbl.EVALUATION_FORM_ID),
				FormID: row.Int(tbl.EVALUATION_FORM_ID),
				Period: Period{
					Valid:     true,
					Cohort:    row.String(tbl.COHORT),
					Stage:     StageEvaluation,
					Milestone: row.String(tbl.MILESTONE),
					StartAt:   row.NullTime(tbl.EVALUATION_START_AT),
					EndAt:     row.NullTime(tbl.EVALUATION_END_AT),
				},
			},
			Evaluatee: Submission{
				Valid:        row.IntValid(tbl.SUBMISSION_ID),
				SubmissionID: row.Int(tbl.SUBMISSION_ID),
				Team: Team{
					Valid:        row.IntValid(tbl.EVALUATEE_TEAM_ID),
					TeamID:       row.Int(tbl.EVALUATEE_TEAM_ID),
					TeamName:     row.String(tbl.EVALUATEE_TEAM_NAME),
					ProjectLevel: row.String(tbl.EVALUATEE_PROJECT_LEVEL),
				},
				SubmissionForm: Form{
					Valid:  row.IntValid(tbl.SUBMISSION_FORM_ID),
					FormID: row.Int(tbl.SUBMISSION_FORM_ID),
					Period: Period{
						Valid:   true,
						StartAt: row.NullTime(tbl.SUBMISSION_START_AT),
						EndAt:   row.NullTime(tbl.SUBMISSION_END_AT),
					},
				},
				OverrideOpen: row.Bool(tbl.SUBMISSION_OVERRIDE_OPEN),
				Submitted:    row.Bool(tbl.SUBMISSION_SUBMITTED),
				UpdatedAt:    row.NullTime(tbl.SUBMISSION_UPDATED_AT),
			},
			OverrideOpen: row.Bool(tbl.EVALUATION_OVERRIDE_OPEN),
			Submitted:    row.Bool(tbl.EVALUATION_SUBMITTED),
			UpdatedAt:    row.NullTime(tbl.EVALUATION_UPDATED_AT),
		}
		row.ScanInto(&e.EvaluationForm.Questions, tbl.EVALUATION_QUESTIONS)
		row.ScanInto(&e.EvaluationAnswers, tbl.EVALUATION_ANSWERS)
		row.ScanInto(&e.Evaluatee.SubmissionForm.Questions, tbl.SUBMISSION_QUESTIONS)
		row.ScanInto(&e.Evaluatee.SubmissionAnswers, tbl.SUBMISSION_ANSWERS)
	}
}

// Feedback given to a Team, by a Team
type TeamFeedback struct {
	Valid            bool
	FeedbackIDOnTeam int
	Evaluator        Team
	Evaluatee        Team
	FeedbackForm     Form
	FeedbackAnswers  formx.Answers
	Submitted        bool
	OverrideOpen     bool
}

// Feedback given to a User, by a Team
type UserFeedback struct {
	Valid            bool
	FeedbackIDOnUser int
	Evaluator        Team
	Evaluatee        User
	Role             string
	FeedbackForm     Form
	FeedbackAnswers  formx.Answers
	Submitted        bool
	OverrideOpen     bool
}
