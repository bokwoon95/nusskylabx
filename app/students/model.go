package students

import (
	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/app/skylab"
	"github.com/bokwoon95/nusskylabx/tables"

	"github.com/bokwoon95/nusskylabx/helpers/erro"
	"github.com/bokwoon95/nusskylabx/helpers/formx"
)

func answersPresent(answers formx.Answers) bool {
	for _, answer := range answers {
		if len(answer) != 0 {
			return true
		}
	}
	return false
}

func (stu Students) UpdateSubmissionAnswers(submissionID int, form map[string][]string) error {
	var questions formx.Questions
	var answers formx.Answers
	var err error
	s, f := tables.SUBMISSIONS(), tables.FORMS()
	err = sq.WithDefaultLog(sq.Lverbose).
		From(s).
		Join(f, f.FORM_ID.Eq(s.SUBMISSION_FORM_ID)).
		Where(s.SUBMISSION_ID.EqInt(submissionID)).
		SelectRowx(func(row *sq.Row) { row.ScanInto(&questions, f.QUESTIONS) }).
		Fetch(stu.skylb.DB)
	if err != nil {
		return erro.Wrap(err)
	}
	answers = formx.ExtractAnswers(form, questions)
	// If no answers are present at all (which is different from answers having
	// blank values), do not proceed with the data update as that is not what
	// we want under any circumstance. If a user wishes to clear out an answer,
	// they would at least provide an empty string.
	if !answersPresent(answers) {
		return nil
	}
	_, err = sq.WithDefaultLog(sq.Lverbose).
		Update(s).
		Set(s.SUBMISSION_DATA.Set(answers)).
		Where(s.SUBMISSION_ID.EqInt(submissionID)).
		Exec(stu.skylb.DB, sq.ErowsAffected)
	return erro.Wrap(err)
}

func (stu Students) UpdateEvaluationAnswers(teamEvaluationID int, form map[string][]string) error {
	var questions formx.Questions
	var answers formx.Answers
	var err error
	te, f := tables.TEAM_EVALUATIONS(), tables.FORMS()
	err = sq.WithDefaultLog(sq.Lverbose).
		From(te).
		Join(f, f.FORM_ID.Eq(te.EVALUATION_FORM_ID)).
		Where(te.TEAM_EVALUATION_ID.EqInt(teamEvaluationID)).
		SelectRowx(func(row *sq.Row) { row.ScanInto(&questions, f.QUESTIONS) }).
		Fetch(stu.skylb.DB)
	if err != nil {
		return erro.Wrap(err)
	}
	answers = formx.ExtractAnswers(form, questions)
	// If no answers are present at all (which is different from answers having
	// blank values), do not proceed with the data update as that is not what
	// we want under any circumstance. If a user wishes to clear out an answer,
	// they would at least provide an empty string.
	if !answersPresent(answers) {
		return nil
	}
	_, err = sq.WithDefaultLog(sq.Lstats).
		Update(te).
		Set(te.EVALUATION_DATA.Set(answers)).
		Where(te.TEAM_EVALUATION_ID.EqInt(teamEvaluationID)).
		Exec(stu.skylb.DB, sq.ErowsAffected)
	return erro.Wrap(err)
}

func (stu Students) UpsertEvaluationAnswers(user skylab.User, milestone string, evaluateeSubmissionID int, form map[string][]string) error {
	var err error
	urs := tables.USER_ROLES_STUDENTS()
	var evaluatorTeamID int
	err = sq.WithDefaultLog(sq.Lstats).
		From(urs).
		Where(urs.USER_ROLE_ID.EqInt(user.Roles[skylab.RoleStudent])).
		SelectRowx(func(row *sq.Row) { evaluatorTeamID = row.Int(urs.TEAM_ID) }).
		Fetch(stu.skylb.DB)
	if err != nil {
		return erro.Wrap(err)
	}
	if !skylab.Contains(skylab.Milestones(), milestone) {
		return erro.Wrap(skylab.ErrMilestoneInvalid)
	}
	f, p := tables.FORMS(), tables.PERIODS()
	var questions formx.Questions
	err = sq.WithDefaultLog(sq.Lstats).
		From(f).
		Join(p, p.PERIOD_ID.Eq(f.PERIOD_ID)).
		Where(
			p.COHORT.EqString(stu.skylb.CurrentCohort()),
			p.STAGE.EqString(skylab.StageEvaluation),
			p.MILESTONE.EqString(milestone),
			f.NAME.EqString(""),
			f.SUBSECTION.EqString(""),
		).
		SelectRowx(func(row *sq.Row) { row.ScanInto(&questions, f.QUESTIONS) }).
		Fetch(stu.skylb.DB)
	if err != nil {
		return erro.Wrap(err)
	}
	answers := formx.ExtractAnswers(form, questions)
	// If no answers are present at all (which is different from answers having
	// blank values), do not proceed with the data update as that is not what
	// we want under any circumstance. If a user wishes to clear out an answer,
	// they would at least provide an empty string.
	if !answersPresent(answers) {
		return nil
	}
	_, err = sq.WithDefaultLog(sq.Lstats).
		Select(tables.UPSERT_EVALUATION(milestone, evaluatorTeamID, evaluateeSubmissionID, answers)).
		Exec(stu.skylb.DB, 0)
	return erro.Wrap(err)
}
