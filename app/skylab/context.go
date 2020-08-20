package skylab

type skylabContext string

// The Context consts are used as keys to access various objects from the
// request context. Every app handler that sets/gets an object from the
// context must use one of these keys, so that all objects that are accessible
// from the context are effectively documented by this list here
const (
	ContextUser             skylabContext = "ContextUser"             // skylab.User
	ContextAdmin            skylabContext = "ContextAdmin"            // skylab.User
	ContextCurrentRole      skylabContext = "ContextCurrentRole"      // string
	ContextCurrentSection   skylabContext = "ContextCurrentSection"   // string
	ContextCurrentMilestone skylabContext = "ContextCurrentMilestone" // string
	ContextDumpJson         skylabContext = "ContextDumpJson"         // bool
	ContextIsProd           skylabContext = "ContextIsProd"           // bool

	// Submission
	ContextCanViewSubmission skylabContext = "ContextCanViewSubmission" // bool
	ContextCanEditSubmission skylabContext = "ContextCanEditSubmission" // bool

	// Evaluation
	ContextCanViewEvaluation skylabContext = "ContextCanViewEvaluation" // bool
	ContextCanEditEvaluation skylabContext = "ContextCanEditEvaluation" // bool

	// User
	ContextCanViewUser skylabContext = "ContextCanViewUser" // bool
	ContextCanEditUser skylabContext = "ContextCanEditUser" // bool

	// Team
	ContextCanViewTeam skylabContext = "ContextCanViewTeam" // bool
	ContextCanEditTeam skylabContext = "ContextCanEditTeam" // bool

	// Form
	ContextCanViewForm skylabContext = "ContextCanViewForm" // bool
	ContextCanEditForm skylabContext = "ContextCanEditForm" // bool

	// Application
	ContextCanViewApplication skylabContext = "ContextCanViewApplication" // bool
	ContextCanEditApplication skylabContext = "ContextCanEditApplication" // bool
)
