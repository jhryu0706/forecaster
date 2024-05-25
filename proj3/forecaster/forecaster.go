package main

import (
	"encoding/json"
	"fmt"
	"os"
	"proj3/data/db"
	"proj3/log"
	"proj3/server"
	"strconv"
	"time"

	"github.com/coocood/qbs"
)

// We collect requirements on run type (parallel or sequential) and number of threads

var usage string = "Usage: forecaster <run_version> <threadcount>\nArguments:\n\t<run_version>: 'p' for parallel or 's' for sequential\n\t<threadcount>: Only provide this value if running parallel version\n"

func main() {
	var config server.Config
	if len(os.Args) < 2 {
		print(usage)
	}
	if os.Args[1] == "s" {
		config.Threadcount = 1
		config.Mode = "s"
	} else if os.Args[1] == "p" {
		threadcount, _ := strconv.Atoi(os.Args[2])
		config.Threadcount = threadcount
		config.Mode = "p"
	} else {
		fmt.Println("Configuration not recognized")
	}
	if len(os.Args) > 3 && os.Args[3] == "w" {
		config.IsWorkstealing = true
	} else {
		config.IsWorkstealing = false
	}
	//set upt db connection
	qbs.Register("sqlite3", "./forecaster/daily_movement.db", "", qbs.NewSqlite3())
	db.CreateTable()
	//set up logging
	log.SetupLogging()
	//create configuration
	decoder := json.NewDecoder(os.Stdin)
	file, err := os.OpenFile("./forecaster/output.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	config.Decoder = decoder
	config.Encoder = encoder
	start := time.Now()
	server.Run(config)
	end := time.Now()
	duration := end.Sub(start)
	fmt.Printf("%d\n", duration.Microseconds())
}
