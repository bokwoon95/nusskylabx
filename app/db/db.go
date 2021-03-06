package db

import (
	"database/sql"
	"errors"

	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/app/skylab"
	"github.com/bokwoon95/nusskylabx/helpers/erro"
	"github.com/bokwoon95/nusskylabx/tables"
)

type DB struct {
	skylb skylab.Skylab
}

func New(skylb skylab.Skylab) DB {
	return DB{
		skylb: skylb,
	}
}

func (d DB) CreateUser(user skylab.User, cohort string) (skylab.User, error) {
	if user.Roles == nil {
		user.Roles = make(map[string]int)
	}
	if user.Email == "" {
		return user, erro.Wrap(skylab.ErrEmailEmpty)
	}
	u := tables.USERS()
	err := sq.WithDefaultLog(sq.Lstats).
		InsertInto(u).
		Columns(u.DISPLAYNAME, u.EMAIL).
		Values(user.Displayname, user.Email).
		OnConflict().DoNothing().
		ReturningRowx(func(row *sq.Row) {
			user.Valid = row.IntValid(u.USER_ID)
			user.UserID = row.Int(u.USER_ID)
		}).
		Fetch(d.skylb.DB)
	if errors.Is(err, sql.ErrNoRows) {
		err = sq.WithDefaultLog(sq.Lstats).
			From(u).
			Where(
				u.EMAIL.EqString(user.Email),
				u.DISPLAYNAME.EqString(user.Displayname),
			).
			SelectRowx(func(row *sq.Row) {
				user.Valid = row.IntValid(u.USER_ID)
				user.UserID = row.Int(u.USER_ID)
			}).
			Fetch(d.skylb.DB)
		if err != nil {
			return user, erro.Wrap(err)
		}
	} else if err != nil {
		return user, erro.Wrap(err)
	}
	ur := tables.USER_ROLES()
	_, err = sq.WithDefaultLog(sq.Lstats).
		InsertInto(ur).
		Valuesx(func(col *sq.Column) {
			for role := range user.Roles {
				col.SetString(ur.COHORT, cohort)
				col.SetInt(ur.USER_ID, user.UserID)
				col.SetString(ur.ROLE, role)
			}
		}).
		OnConflict().DoNothing().
		Exec(d.skylb.DB, sq.ErowsAffected)
	if err != nil {
		return user, erro.Wrap(err)
	}
	var userRoleID int
	var role string
	err = sq.WithDefaultLog(sq.Lstats).
		From(ur).
		Where(
			ur.COHORT.EqString(cohort),
			ur.USER_ID.EqInt(user.UserID),
		).
		Selectx(func(row *sq.Row) {
			userRoleID = row.Int(ur.USER_ROLE_ID)
			role = row.String(ur.ROLE)
		}, func() {
			user.Roles[role] = userRoleID
		}).
		Fetch(d.skylb.DB)
	return user, erro.Wrap(err)
}
