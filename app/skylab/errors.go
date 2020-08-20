package skylab

import "github.com/bokwoon95/nusskylabx/helpers/erro"

// Each error starts with a 5 letter error code, intended to represent the custom error codes that can be returned from postgres stored procedures.
//
// How to name custom error codes https://stackoverflow.com/a/22600394
// • Start with a capital letter but not F (predefined config file errors), H (fdw), P (PL/pgSQL) or X (internal).
// • Do not use 0 (zero) or P in the 3rd column. Predefined error codes use these commonly.
// • Use a capital letter in the 4th position. No predefined error codes have this.
// 'As an example, start with a character for your app: "T". Then a two-char error class: "3G". Then a sequential code "A0"-"A9", "B0"-"B9", etc. Yields T3GA0, T3GA1, etc.'
const (
	// Roles
	ErrNoRoles        erro.BaseError = "OLAJQ User %+v has no roles"
	ErrNotAnApplicant erro.BaseError = "OC8FY User %+v is not an applicant"
	ErrNotAStudent    erro.BaseError = "ONXIU User %+v is not a student"
	ErrNotAMentor     erro.BaseError = "OQVPW User %+v is not a mentor"
	ErrNotAnAdviser   erro.BaseError = "OKDRA User %+v is not an adviser"
	ErrNotAnAdmin     erro.BaseError = "OD6HR User %+v is not an admin"
	ErrNoPeerTeams    erro.BaseError = "OO6BX Team %+v has no peer teams"

	// Skylab Enums
	ErrCohortInvalid       erro.BaseError = "OLALE Cohort '%s' is not a valid Skylab cohort"
	ErrStageInvalid        erro.BaseError = "OLASP Stage '%s' is not a valid Skylab stage"
	ErrMilestoneInvalid    erro.BaseError = "OLADN Milestone '%s' is not a valid Skylab milestone"
	ErrRoleInvalid         erro.BaseError = "OLAZN Role '%s' is not a valid Skylab role"
	ErrProjectLevelInvalid erro.BaseError = "OEHGC Project Level '%s' is not a valid Skylab Project Level"

	// Forms
	ErrSubmissionFormNotExist  erro.BaseError = "OYOE8 Submission form doesn't exist"
	ErrApplicationFormNotExist erro.BaseError = "OC8BK Application Form doesn't exist"

	// APPLICATION C8
	ErrApplicationNotExist                  erro.BaseError = "OC8U9 Application doesn't exist"
	ErrApplicationPeriodNotFound            erro.BaseError = "OC8EX Application period not found"
	ErrApplicationMagicstringNotExist       erro.BaseError = "OC8UM Applicant {application_id:%d} tried joining an application with an invalid magicstring"
	ErrApplicantJoinedOwnApplication        erro.BaseError = "OC8JK Applicant {uid:%d} tried joining an application created by himself"
	ErrApplicationAlreadyFull               erro.BaseError = "OC8FB Application {application_id:%d} is already full"
	ErrApplicantLeaveNonExistentApplication erro.BaseError = "OC8EN Applicant {uid:%d} tried leaving an application when he isn't in one"
	ErrApplicantLeaveAcceptedApplication    erro.BaseError = "OC8A4 Applicant {uid:%d} tried leaving an application [application_id:%d] that was already accepted"
	ErrApplicationDeleted                   erro.BaseError = "OC8W6 Application {application_id:%d} already accepted/deleted"
	ErrApplicationIncomplete                erro.BaseError = "OC8KH Tried accepting an incomplete application"
	ErrApplicationNoTeam                    erro.BaseError = "OC8R1 Tried un-accepting an application that had never been accepted"

	// Misc
	ErrStudentNoTeam           erro.BaseError = "ONXDI Student {uid:%d} does not belong to any team"
	ErrTeamMoreThanTwoStudents erro.BaseError = "OYSGQ Team {tid:%d} has more than two students"
	ErrEmailNotAuthorized      erro.BaseError = "OLALP Email '%s' is not authorized to signup for any role"
	ErrEmailEmpty              erro.BaseError = "OLAR9 Email must be non-empty"

	// Not exist
	ErrUserNotExist           erro.BaseError = "OLAMC User does not exist: %s"
	ErrTeamNotExist           erro.BaseError = "OWHZT Team does not exist: %s"
	ErrSubmissionNotExist     erro.BaseError = "OQ7WT Submission does not exist: %s"
	ErrFormdataNotExist       erro.BaseError = "ON354 Formdata {fdid:%d} does not exist"
	ErrTeamEvaluationNotExist erro.BaseError = "OSF99 Team Evaluation does not exist: %s"
	ErrUserEvaluationNotExist erro.BaseError = "OR2KO User Evaluation does not exist: %s"
	ErrTeamFeedbackNotExist   erro.BaseError = "ONDHA Team Feedback does not exist: %s"
	ErrUserFeedbackNotExist   erro.BaseError = "OXZW4 User Feedback does not exist: %s"
	ErrFormNotExist           erro.BaseError = "OLAJX Form does not exist: %s"
	ErrPeriodNotExist         erro.BaseError = "OLAEE Period does not exist: %s"
)
