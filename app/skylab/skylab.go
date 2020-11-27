// Package app implements the Skylab website for Orbital
package skylab

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"

	"github.com/bokwoon95/nusskylabx/helpers/auth"
	"github.com/bokwoon95/nusskylabx/helpers/cookies"
	"github.com/bokwoon95/nusskylabx/helpers/flash"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"

	"github.com/DATA-DOG/go-txdb"
	"github.com/bokwoon95/nusskylabx/helpers/erro"
	"github.com/bokwoon95/nusskylabx/helpers/logutil"
	"github.com/bokwoon95/nusskylabx/helpers/testutil"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/microcosm-cc/bluemonday"
	"github.com/oxtoacart/bpool"
)

var (
	// sourcefile is the path to this file
	_, sourcefile, _, _ = runtime.Caller(0)
	// ProjectRootDir will point to wherever this project's root directory is located
	// on the user's computer. It is calculated as an offset of two
	// directories up from the skylab.go file
	ProjectRootDir = filepath.Join(filepath.Dir(sourcefile), ".."+string(os.PathSeparator)+"..") + string(os.PathSeparator)
)

func init() {
	txdb.Register("txdb", "postgres", os.Getenv("DATABASE_URL"))
}

// LoadDotenv will read load the environment variables from the .env file in
// the project root directory. The environment variable can then be accessed
// normally with os.Getenv.
//
// If the .env file is not found, it will read from the .env.default file
// instead. Such is the case when carrying out the automated tests in Github
// Actions. It is therefore important to ensure that the .env.default file has
// useful defaults set in it.
func LoadDotenv() {
	if err := godotenv.Load(ProjectRootDir + ".env"); err != nil {
		if err := godotenv.Load(ProjectRootDir + ".env.default"); err != nil {
			panic("unable to source from either .env or .env.default")
		}
	}
}

// Config contains all the parameters needed to start an instance of the Skylab
// application
type Config struct {
	BaseURL      string // optional
	Port         string // optional
	DatabaseURL  string // !REQUIRED
	MigrationDir string // !REQUIRED
	IsProd       string // optional
	DebugMode    string // optional
	SecretKey    string // optional
	DisableCsrf  string // optional

	// Experimental
	MailerEnabled string
	SmtpHost      string
	SmtpPort      string
	SmtpUsername  string
	SmtpPassword  string
}

// Skylab is the server struct.
type Skylab struct {
	BaseURL     string
	port        string
	IsProd      bool
	SecretKey   string
	Policy      *bluemonday.Policy
	DB          *sqlx.DB
	Mux         *chi.Mux
	Bufpool     *bpool.BufferPool
	Log         *logutil.Logger
	DisableCsrf bool
	Templates   *template.Template

	// Mailer
	// NOTE: not used
	MailerEnabled bool
	SmtpHost      string
	SmtpPort      int
	SmtpUsername  string
	SmtpPassword  string

	// These fields below are not safe for concurrent access and access must be
	// synchronized with a mutex.
	RW        *sync.RWMutex
	cohorts   []string
	templates map[string]*template.Template
}

// NewWithoutDB creates a new instance of Skylab that initializes everything
// except the database connection.
func NewWithoutDB(config Config) Skylab {
	skylb := Skylab{}

	// RWMutex
	skylb.RW = &sync.RWMutex{}

	// templates
	skylb.templates = make(map[string]*template.Template)

	// BaseURL
	skylb.BaseURL = "localhost" //
	if config.BaseURL != "" {
		skylb.BaseURL = strings.TrimSuffix(config.BaseURL, "/")
	}

	// Port
	skylb.port = "0"
	if config.Port != "" {
		skylb.port = strings.TrimPrefix(config.Port, ":")
	}

	// IsProd
	skylb.IsProd = config.IsProd == "true"

	// DebugMode
	skylb.Log = logutil.NewLogger(ioutil.Discard)
	if config.DebugMode == "true" {
		skylb.Log.SetOutput(os.Stdout)
	}

	// SecretKey
	skylb.SecretKey = "this-is-the-default-hmac-key"
	if config.SecretKey != "" {
		skylb.SecretKey = config.SecretKey
	}

	// Policy
	skylb.Policy = bluemonday.UGCPolicy()
	skylb.Policy.AllowStyling() // allow styling with the html class attribute

	// Mux
	skylb.Mux = chi.NewRouter()

	// BufPool
	skylb.Bufpool = bpool.NewBufferPool(64)

	// DisableCsrf
	skylb.DisableCsrf = config.DisableCsrf == "true"

	// Mailer
	var err error
	skylb.MailerEnabled = config.MailerEnabled == "true"
	skylb.SmtpHost = config.SmtpHost
	skylb.SmtpPort, err = strconv.Atoi(config.SmtpPort)
	if err != nil && skylb.MailerEnabled {
		log.Fatalf("Mailer is enabled but config.SmtpPort '%s' is not a valid port (need integer)", config.SmtpPort)
	}
	skylb.SmtpUsername = config.SmtpUsername
	skylb.SmtpPassword = config.SmtpPassword

	return skylb
}

// New creates a new instance of Skylab
func New(config Config) (Skylab, error) {
	skylb := NewWithoutDB(config)
	// Database
	if config.DatabaseURL == "" {
		return skylb, erro.Wrap(fmt.Errorf("DatabaseURL cannot be empty!"))
	}
	var err error
	skylb.DB, err = sqlx.Open("postgres", config.DatabaseURL)
	if err != nil {
		return skylb, erro.Wrap(err)
	}
	err = skylb.DB.Ping() // Make sure db connection works
	if err != nil {
		return skylb, erro.Wrap(fmt.Errorf("Could not ping the database, is the database running? %w", err))
	}
	err = EnsureDatabasePopulated(config.DatabaseURL, config.MigrationDir) // Ensure database is not empty
	if err != nil {
		return skylb, erro.Wrap(err)
	}
	// cohorts
	err = skylb.RefreshCohorts()
	if err != nil {
		return skylb, erro.Wrap(err)
	}
	return skylb, nil
}

// NewWithTestDB creates a new instance of Skylab with a test database. The
// test database reuses the same DATABASE_URL as the actual database, except
// all changes are rolled back once the test database is closed i.e.
// DB.Close(). This makes it easy for running database-related tests, without
// affecting the actual state of the database. You can optionally provide a
// txdbIdentifier string to id the test database being created, or you can
// leave it blank for a random identifier to be used. I assume databases with
// the same txdbIdentifier will share the same state, making it possible to
// share the same database between two tests by using the same txdbIdentifier.
func NewWithTestDB(t *testing.T, config Config, txdbIdentifier string) Skylab {
	skylb := NewWithoutDB(config)
	// Database
	if config.DatabaseURL == "" {
		panic("DatabaseURL cannot be empty!")
	}
	var err error
	if txdbIdentifier == "" {
		txdbIdentifier = uuid.New().String()
	}
	skylb.DB, err = sqlx.Open("txdb", txdbIdentifier)
	if err != nil {
		panic(err)
	}
	t.Cleanup(func() { skylb.DB.Close() })
	err = skylb.DB.Ping() // Make sure db connection works
	if err != nil {
		panic(err)
	}
	err = EnsureDatabasePopulated(config.DatabaseURL, config.MigrationDir) // Ensure database is not empty
	if err != nil {
		panic(err)
	}
	// cohorts
	err = skylb.RefreshCohorts()
	if err != nil {
		panic(err)
	}
	return skylb
}

// NewTestDefault is the same as NewWithTestDB, except it uses a default
// skylab.Config instead of requiring the caller to pass one
func NewTestDefault(t *testing.T) Skylab {
	LoadDotenv()
	return NewWithTestDB(t, Config{
		DatabaseURL:  testutil.Getenv("DATABASE_URL"),
		MigrationDir: testutil.Getenv("MIGRATION_DIR"),
		DisableCsrf:  "true",
	}, "")
}

// TODO: instead of being a custom function, EnsureDatabasePopulated should
// utilise loadsql's functions with the configuration -clean. So running `go
// run main.go` is enough to bootstrap everything (tables, triggers, views,
// functions, data). loadsql should accept an io.Writer to write to instead of
// taking control of stdout, and EnsureDatabasePopulated simply has to pass in
// os.Stdout to print to stdout as usual.
func EnsureDatabasePopulated(databaseURL, migrationDir string) error {
	// https://github.com/golang-migrate/migrate/issues/292.
	// filepath.ToSlash must be used to convert Windows backslashes to Unix
	// forward slashes as that is the only format that migrate recognizes
	migrationDirAbs := filepath.ToSlash(path.Join(ProjectRootDir, migrationDir))
	m, err := migrate.New("file://"+migrationDirAbs, databaseURL)
	if err != nil {
		return erro.Wrap(err)
	}
	_, _, err = m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		// No migration has been applied to the database yet, perform initial migration
		fmt.Printf("%s has no tables, attempting migration... \n", databaseURL)
		err = m.Up()
		if err != nil {
			return erro.Wrap(err)
		}
		fmt.Println("migration success")
		fmt.Println("-------------------- NOTICE ------------------------")
		fmt.Println("Please run the command")
		fmt.Println("    go run cmd/loadsql/main.go -clean")
		fmt.Println("to load")
		fmt.Println("1) sql functions")
		fmt.Println("2) sql views")
		fmt.Println("3) dummy data")
		fmt.Println("into the database for the website to work")
		fmt.Println("-------------------- END NOTICE --------------------")
	}
	return nil
}

// BaseURL returns the base URL of the website e.g. localhost or
// nusskylabx.comp.nus.edu.sg.
//
// This is purely for display purposes as the actual application is always
// running on localhost, even in production. On the production server, Nginx
// will reverse proxy the localhost application to the outside world. This
// greatly simplifies deploying the app in production as everything is
// exactly the same as in the development environment. The application also
// need not concern itself with a HTTPS certificate as that is handled on the
// Nginx layer.
func (skylb Skylab) BaseURLWithProtocol() string {
	// A bit hacky, we assume anything that is not "localhost" or "127.0.0.1"
	// to be on a production https URL.
	var url string
	switch skylb.BaseURL {
	case "localhost", "127.0.0.1":
		url = fmt.Sprintf("http://%s%s", skylb.BaseURL, skylb.Port())
	default:
		url = fmt.Sprintf("https://%s", skylb.BaseURL)
	}
	return url
}

// Port returns the port that the application is currently running on.
func (skylab Skylab) Port() string {
	return ":" + skylab.port
}

// Cohorts returns all cohorts (excluding the empty cohort).
func (skylb Skylab) Cohorts() []string {
	skylb.RW.RLock()
	defer skylb.RW.RUnlock()
	return skylb.cohorts
}

// CurrentCohort returns the current cohort.
//
// The CurrentCohort() is not the same as the LatestCohort(). CurrentCohort()
// only points to the current calendar year. LatestCohort() points to the
// latest created cohort in the database, which may not be the same as cohorts
// may be created ahead of the current calendar year.
func (skylb Skylab) CurrentCohort() string {
	cohort := strconv.Itoa(time.Now().Year())
	return cohort
}

// LatestCohort returns the latest cohort in the database.
//
// The LatestCohort() is not the same as the CurrentCohort(). LatestCohort()
// points to the latest cohort created in the database. CurrentCohort() points
// only to the current calendar year, which may not be the same as cohorts may
// be created ahead of the current calendar year.
func (skylb Skylab) LatestCohort() string {
	skylb.RW.RLock()
	defer skylb.RW.RUnlock()
	if len(skylb.cohorts) > 0 {
		return skylb.cohorts[0]
	}
	return ""
}

// RefreshCohorts refreshes the current list of cohorts (kept in memory) by
// refreshing it from the database.
//
// The reason why a list of cohorts is kept in-memory instead of querying it
// from the database on demand is because the cohort list rarely changes (once
// every year). That makes it inefficient to request it from database all the
// time. By keeping it in-memory, cohort retrieval is much faster but we have
// to refresh the cohort list everytime we add or remove a cohort from the
// database.
//
// This may be a premature optimization.
func (skylb *Skylab) RefreshCohorts() error {
	skylb.RW.Lock()
	defer skylb.RW.Unlock()
	skylb.cohorts = skylb.cohorts[:0]
	rows, err := skylb.DB.Queryx("SELECT cohort FROM cohort_enum ORDER BY insertion_order DESC")
	if err != nil {
		return erro.Wrap(err)
	}
	defer rows.Close()
	for rows.Next() {
		var cohort string
		err := rows.Scan(&cohort)
		if err != nil {
			return erro.Wrap(err)
		}
		skylb.cohorts = append(skylb.cohorts, cohort)
	}
	return nil
}

// Hash will hash input byte slice and output a string
func (skylb Skylab) Hash(input []byte) (output string) {
	return auth.Hash(skylb.SecretKey, input)
}

// EncodeVariableInCookie will serialize any go variable into a string and
// store it in a client side cookie, where it can later be retrieved from the
// cookie with DecodeVariableFromCookie.
//
// The cookie will be securely signed with a digital hash signature to prevent
// anyone from tampering with the cookie's contents.
func (skylb Skylab) EncodeVariableInCookie(w http.ResponseWriter, cookiename string, variable interface{}, opts ...cookies.CookieOpt) error {
	err := cookies.NewEncoder(skylb.SecretKey).EncodeVariableInCookie(w, cookiename, variable, opts...)
	return erro.Wrap(err)
}

// DecodeVariableFromCookie will retrieve a go variable from a client side
// cookie that was previously set by EncodeVariableInCookie. If the named
// cookie does not exist, it will fail with a cookies.ErrNoCookie error. Note
// that you need to pass a pointer to the variable you wish to decode the
// cookie value into, not the variable itself.
//
// The cookie's digital hash signature will be checked for any tampering. If
// the signature is wrong, DecodeVariableFromCookie will fail with an
// auth.ErrDeserializeOutputInvalid error.
func (skylb Skylab) DecodeVariableFromCookie(r *http.Request, cookiename string, variablePtr interface{}) error {
	err := cookies.NewEncoder(skylb.SecretKey).DecodeVariableFromCookie(r, cookiename, variablePtr)
	return erro.Wrap(err)
}

// SetFlashMsgs will set a bunch of flash messages into a cookie that can be
// read later by GetFlashMsgs. See also: SetFlashMsgsCtx.
//
// NOTE: This returns a new request with the flash messages in the context
// updated, so make sure to re-assign the request:
//		r, _ = skylb.SetFlashMsgs(w, r, msgs)
func (skylb Skylab) SetFlashMsgs(w http.ResponseWriter, r *http.Request, msgs map[string][]string) (*http.Request, error) {
	var err error
	r, err = flash.NewEncoder(skylb.SecretKey).SetFlashMsgs(w, r, msgs)
	return r, erro.Wrap(err)
}

// GetFlashMsgs will get all flash messages from both the cookie and request
// context, after which it will delete the cookie (hence a 'flash' message).
//
// Usually you won't call this function directly, instead you will include the
// "helpers/flash/flash.html" template in your html file which will
// automatically pick up the flash messages and display them if any exist.
func (skylb Skylab) GetFlashMsgs(w http.ResponseWriter, r *http.Request) (map[string][]flash.FlashMsg, error) {
	flashmsgs, err := flash.NewEncoder(skylb.SecretKey).GetFlashMsgs(w, r)
	return flashmsgs, erro.Wrap(err)
}
