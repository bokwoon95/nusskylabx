// Package students implements the student-facing pages on Skylab
package students

import (
	"github.com/bokwoon95/nusskylabx/app/db"
	"github.com/bokwoon95/nusskylabx/app/skylab"
	"github.com/bokwoon95/nusskylabx/helpers/erro"
)

type Students struct {
	skylb skylab.Skylab
	d     db.DB
}

func New(skylb skylab.Skylab) Students {
	return Students{
		skylb: skylb,
		d:     db.New(skylb),
	}
}

func milestoneFromSection(section string) (milestone string) {
	switch section {
	case skylab.StudentM1Submission, skylab.StudentM1Evaluation:
		return skylab.Milestone1
	case skylab.StudentM2Submission, skylab.StudentM2Evaluation:
		return skylab.Milestone2
	case skylab.StudentM3Submission, skylab.StudentM3Evaluation:
		return skylab.Milestone3
	default:
		return skylab.MilestoneNull
	}
}

func (stu Students) getSectionFromTsid(tsid int) (string, error) {
	query := `
SELECT
	p.milestone
FROM
	team_submissions AS ts
	JOIN form_schema AS fs ON fs.fsid = ts.schema
	JOIN periods AS p ON p.pid = fs.period
WHERE
	ts.tsid = $1
`
	var milestone string
	err := stu.skylb.DB.QueryRowx(query, tsid).Scan(&milestone)
	if err != nil {
		return string(""), erro.Wrap(err)
	}
	switch milestone {
	case skylab.Milestone1:
		return skylab.StudentM1Submission, nil
	case skylab.Milestone2:
		return skylab.StudentM2Submission, nil
	case skylab.Milestone3:
		return skylab.StudentM3Submission, nil
	default:
		return "", erro.Wrap(skylab.ErrMilestoneInvalid)
	}
}
