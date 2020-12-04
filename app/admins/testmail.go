package admins

import (
	"net/http"
	"strings"

	"github.com/bokwoon95/nusskylabx/app/skylab"

	"github.com/bokwoon95/nusskylabx/helpers/mailutil"
)

func (adm Admins) Testmail(w http.ResponseWriter, r *http.Request) {
	adm.skylb.Log.TraceRequest(r)
	r = adm.skylb.SetRoleSection(w, r, skylab.RoleAdmin, skylab.AdminTestmail)
	adm.skylb.Wender(w, r, nil, "app/admins/testmail.html")
}

func (adm Admins) TestmailPost(w http.ResponseWriter, r *http.Request) {
	config := mailutil.Config{
		SmtpHost:     adm.skylb.SmtpHost,
		SmtpPort:     adm.skylb.SmtpPort,
		SmtpUsername: adm.skylb.SmtpUsername,
		SmtpPassword: adm.skylb.SmtpPassword,
		From:         "e0031874@u.nus.edu",
	}
	recipients := strings.Split(r.FormValue("to"), ",")
	subject := r.FormValue("subject")
	message := r.FormValue("message")
	err := mailutil.Send(config, recipients, subject, message)
	if err != nil {
		adm.skylb.InternalServerError(w, r, err)
		return
	}
	r, _ = adm.skylb.SetFlashMsgs(w, r, map[string][]string{"sent": {"Message sent!"}})
	http.Redirect(w, r, skylab.AdminTestmail, http.StatusMovedPermanently)
}
