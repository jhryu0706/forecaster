package log

import (
	"log"
	"os"
)

func SetupLogging() {
	file, err := os.OpenFile("./log/app.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	log.SetOutput(file)
	// log.SetOutput(os.Stdout)
	log.SetFlags(0)
}
