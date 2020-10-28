package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/bokwoon95/nusskylabx/app"
	"github.com/bokwoon95/nusskylabx/app/skylab"

	"github.com/go-chi/docgen"
)

var (
	setup      = flag.Bool("setup", false, "")
	docgenmd   = flag.Bool("docgenmd", false, "")
	docgenjson = flag.Bool("docgenjson", false, "")
)

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Llongfile)
	skylab.LoadDotenv()
	// ENTRYPOINT: All routes are registered here
	skylb, err := app.NewSkylab(skylab.Config{
		BaseURL:       os.Getenv("BASE_URL"),
		Port:          os.Getenv("PORT"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		MigrationDir:  os.Getenv("MIGRATION_DIR"),
		IsProd:        os.Getenv("IS_PROD"),
		DebugMode:     os.Getenv("DEBUG_MODE"),
		SecretKey:     os.Getenv("SECRET_KEY"),
		MailerEnabled: os.Getenv("MAILER_ENABLED"),
		SmtpHost:      os.Getenv("SMTP_HOST"),
		SmtpPort:      os.Getenv("SMTP_PORT"),
		SmtpUsername:  os.Getenv("SMTP_USERNAME"),
		SmtpPassword:  os.Getenv("SMTP_PASSWORD"),
	})
	if err != nil {
		log.Fatalln(err)
		return
	}
	if *docgenmd {
		fmt.Println(docgen.MarkdownRoutesDoc(skylb.Mux, docgen.MarkdownOpts{
			ProjectPath: "github.com/bokwoon95/nusskylabx",
			Intro:       "Lorem Ipsum Dolor Sit Amet",
		}))
		return
	}
	if *docgenjson {
		fmt.Println(docgen.JSONRoutesDoc(skylb.Mux))
		return
	}
	if *setup {
		fmt.Println("setup completed")
		return
	}
	switch skylb.BaseURL {
	case "localhost", "127.0.0.1":
		fmt.Printf("Listening on localhost%s\n", skylb.Port())
	default:
		fmt.Printf("Listening on localhost%s, reverse proxied from %s\n", skylb.Port(), skylb.BaseURLWithProtocol())
	}
	log.Fatal(http.ListenAndServe(skylb.Port(), skylb.Mux))
}
