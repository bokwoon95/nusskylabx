package applicants

import (
	"testing"
)

// email, displayname -> applicant with email, displayname
// repeated calls to IdempotentCreateApplicant don't change anything
// displayname || email missing -> 400
// email is empty -> 400
func TestIdempotentCreateApplicant(t *testing.T) {
	// apt := Applicants{Skylb: app.NewTestDefault()}
	// displayname, email := uuid.New().String(), uuid.New().String()
	// del := func(t *testing.T, email string) {
	// 	query := `DELETE FROM `
	// }
	// setup := func(r *http.Request) *http.Request {
	// 	r = r.WithContext(context.WithValue(r.Context(), "displayname", displayname))
	// 	r = r.WithContext(context.WithValue(r.Context(), "email", email))
	// 	return r
	// }
	// test := func(w *httptest.ResponseRecorder, r *http.Request) {
	// 	is := is.New(t)
	// 	var exists bool
	// 	query := `
	// 	SELECT EXISTS(
	// 		SELECT
	// 			1
	// 		FROM
	// 			user_roles AS ur JOIN users AS u USING (uid)
	// 		WHERE
	// 			u.displayname = $1
	// 			AND u.email = $2
	// 			AND ur.role = $3
	// 			AND ur.cohort = $4
	// 	)
	// 	`
	// 	err := apt.Skylb.DB.QueryRowx(query, displayname, email, app.RoleApplicant).Scan(&exists)
	// 	is.NoErr(err)
	// 	is.True(exists)
	// 	if _, err := apt.Skylb.DB.Exec(`DELETE FROM users`)
	// }
	// testutil.TestMiddlewares(t, setup, test, apt.IdempotentCreateApplicant)
}
