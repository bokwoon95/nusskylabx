// Package advisers implements the adviser-facing pages on Skylab
package advisers

import (
	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/app/db"
	"github.com/bokwoon95/nusskylabx/app/skylab"
	"github.com/bokwoon95/nusskylabx/helpers/erro"
	"github.com/bokwoon95/nusskylabx/tables"
)

type Advisers struct {
	skylb skylab.Skylab
	d     db.DB
}

func New(skylb skylab.Skylab) Advisers {
	return Advisers{
		skylb: skylb,
		d:     db.New(skylb),
	}
}

func milestoneFromSection(section string) (milestone string) {
	switch section {
	case skylab.AdviserM1MakeEvaluation, skylab.AdviserM1ViewEvaluation:
		return skylab.Milestone1
	case skylab.AdviserM2MakeEvaluation, skylab.AdviserM2ViewEvaluation:
		return skylab.Milestone2
	case skylab.AdviserM3MakeEvaluation, skylab.AdviserM3ViewEvaluation:
		return skylab.Milestone3
	default:
		return skylab.MilestoneNull
	}
}

func (adv Advisers) getTeamIDs(user skylab.User) (teamIDs []int, err error) {
	t := tables.TEAMS()
	var adviserTeamIDs []int64
	err = sq.WithDefaultLog(sq.Lstats).
		From(t).
		Where(t.ADVISER_USER_ROLE_ID.EqInt(user.Roles[skylab.RoleAdviser])).
		GroupBy(t.ADVISER_USER_ROLE_ID).
		SelectRowx(func(row *sq.Row) {
			row.ScanArray(&adviserTeamIDs, sq.Fieldf("array_agg(?)", t.TEAM_ID))
		}).
		Fetch(adv.skylb.DB)
	if err != nil {
		return teamIDs, erro.Wrap(err)
	}
	teamIDs = make([]int, len(adviserTeamIDs))
	for i := range adviserTeamIDs {
		teamIDs[i] = int(adviserTeamIDs[i])
	}
	return teamIDs, nil
}
