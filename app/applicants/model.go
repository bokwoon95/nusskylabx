package applicants

import (
	"github.com/bokwoon95/nusskylabx/app/skylab"

	"github.com/bokwoon95/nusskylabx/helpers/erro"
)

func (apt Applicants) JoinApplicationModel(user skylab.User, magicstring string) (err error) {
	query := "SELECT app.join_application($1, $2, $3)"
	_, err = apt.skylb.DB.Exec(query, user.Displayname, user.Email, magicstring)
	return erro.Wrap(err)
}

func (apt Applicants) LeaveApplicationModel(user skylab.User) (err error) {
	query := "SELECT app.leave_application($1)"
	_, err = apt.skylb.DB.Exec(query, user.UserID)
	return erro.Wrap(err)
}
