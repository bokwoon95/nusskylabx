package main

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/bokwoon95/nusskylabx/app"
	"github.com/bokwoon95/nusskylabx/app/skylab"
	"github.com/bokwoon95/nusskylabx/helpers/logutil"
	"github.com/bokwoon95/nusskylabx/helpers/urlparams"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/google/uuid"
)

var (
	_, sourcefile, _, _ = runtime.Caller(0)                                   // sourcefile is the path to this file
	rootdir             = filepath.Dir(sourcefile) + string(os.PathSeparator) // rootdir is the directory containing this file
)

// Log is a global logger variable
var Log = logutil.NewLogger(os.Stdin)

// DB is a global pooled database connection
var DB *sql.DB

func main() {
	skylab.LoadDotenv()
	PORT := os.Getenv("PORT")
	var err error
	// The txdb (https://github.com/DATA-DOG/go-txdb) driver is using our
	// actual database under the hood. But any changes made with the txdb
	// connection will be rolled back once DB.Close() is called, allowing us to
	// use the database without actually making permanent changes to it.
	DB, err = sql.Open("txdb", os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}
	defer DB.Close()

	mux := chi.NewRouter()
	mux.Use(middleware.RequestID) // Insert a unique ID in every request
	mux.Use(middleware.Recoverer) // pretty print any panicked errors
	mux.Use(middleware.Logger)    // Log all paths hit

	// Serve the static directory because we need the style.css file
	app.ServeDirectory(mux, http.Dir(skylab.ProjectRootDir+"static"), "/static")

	mux.Get("/", Home)
	mux.Post("/upload", Upload)
	mux.Get("/img/{uuid}", ServeImg)
	fmt.Println("Listening on localhost:" + PORT)
	err = http.ListenAndServe(":"+PORT, mux)
	if strings.Contains(err.Error(), "address already in use") {
		log.Fatalf("Application already listening on port :%s, please terminate that application first", PORT)
	}
	log.Fatalln(err)
}

func Home(w http.ResponseWriter, r *http.Request) {
	Log.TraceRequest(r)
	type Image struct {
		UUID     uuid.UUID
		Filename string
		Filetype string
	}
	type Data struct {
		Images []Image
	}
	var data Data
	rows, err := DB.Query(`SELECT uuid, name, type FROM media ORDER BY name ASC, created_at DESC`)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	fmt.Println("Files on the system:")
	var i int
	for rows.Next() {
		var img Image
		err = rows.Scan(&img.UUID, &img.Filename, &img.Filetype)
		if err != nil {
			panic(err)
		}
		i++
		fmt.Printf("%d) uuid:%s, filename:%s, filetype:%s\n", i, img.UUID, img.Filename, img.Filetype)
		data.Images = append(data.Images, img)
	}
	t, err := template.ParseFiles(rootdir + "main.html")
	if err != nil {
		panic(err)
	}
	err = t.Execute(w, data)
	if err != nil {
		panic(err)
	}
}

// Upload https://stackoverflow.com/a/40699578
func Upload(w http.ResponseWriter, r *http.Request) {
	Log.TraceRequest(r)
	r.ParseMultipartForm(32 << 20) // limit uploaded file size
	buf := &bytes.Buffer{}
	file, header, err := r.FormFile("upload_img")
	if err != nil {
		switch {
		case errors.Is(err, http.ErrMissingFile):
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
			return
		default:
			panic(err)
		}
	}
	defer file.Close()
	io.Copy(buf, file)
	data := buf.Bytes()
	var id uuid.UUID
	err = DB.QueryRow(
		`INSERT INTO media (name, type, data) VALUES ($1, $2, $3) RETURNING uuid`,
		header.Filename, http.DetectContentType(data), data,
	).Scan(&id)
	fmt.Printf("Uploaded file %s with uuid %s\n", header.Filename, id)
	if err != nil {
		panic(err)
	}
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func ServeImg(w http.ResponseWriter, r *http.Request) {
	Log.TraceRequest(r)
	id, err := urlparams.String(r, "uuid")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Serving image with uuid %s\n", id)
	var imageb []byte
	err = DB.QueryRow(`SELECT data FROM media WHERE uuid = $1`, id).Scan(&imageb)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Unable to find file with uuid '%s'\n", id)
		default:
			panic(err)
		}
		return
	}
	_, err = w.Write(imageb)
	if err != nil {
		panic(err)
	}
}
