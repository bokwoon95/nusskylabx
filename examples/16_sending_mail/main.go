package main

import (
	"log"
	"os"
	"strconv"

	"github.com/bokwoon95/nusskylabx/app/skylab"
	"github.com/bokwoon95/nusskylabx/helpers/mailutil"
)

func main() {
	skylab.LoadDotenv()
	PORT, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	cfg := mailutil.Config{
		SmtpHost:     os.Getenv("SMTP_HOST"),
		SmtpPort:     PORT,
		SmtpUsername: os.Getenv("SMTP_USERNAME"),
		SmtpPassword: os.Getenv("SMTP_PASSWORD"),
		From:         os.Getenv("nusskylab.2@gmail.com"),
	}
	log.Println(cfg)
	// err := mailutil.Send(cfg, []string{"bokwoon.c@gmail.com"}, "ayyo", "whassup my man")
	// if err != nil {
	// 	log.Fatalln(err)
	// }
}
