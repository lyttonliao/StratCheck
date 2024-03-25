package main

import (
	"fmt"
	"flag"
	"log"
	"net/http"
	"os"
	"time"
)

// Declare a string containing the app version number as hard-coded global constant
// Will automatically generate this at build time later
const version = "1.0.0"

// Define config struct to hold all configuration settings
// For now, config settings will only have the network port that we want the server to listen on
// and the name of the operation environment for the application (dev, staging, prod)
type config struct {
	port int
	env string
}

// Define application struct to hold dependencies for our HTTP handlers, helpers, middleware
type application struct {
	config config
	logger *log.Logger
}

func main() {
	var cfg config

	// Read value of the port and env command-lines into config struct
	// Default values: 4000, development
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	// Initialize a new logger which writes message to standard out stream
	// prefixed with current date and time
	logger := log.New(os.Stdout, "", log.Ldate | log.Ltime)

	// Declare an instance of the application struct, containing the config struct and logger
	app := &application{
		config: cfg,
		logger: logger,
	}

	// Declare a HTTP server with timeout settings, which listens on the port
	// provided in the config struct and uses servemux as the handler
	srv := http.Server{
		Addr: 			fmt.Sprintf(":%d", cfg.port),
		Handler: 		app.routes(),
		IdleTimeout:	time.Minute,
		ReadTimeout:	10 * time.Second,
		WriteTimeout: 	30 * time.Second,
	}

	// Start the HTTP server
	logger.Printf("starting %s server on %s", cfg.env, srv.Addr)
	err := srv.ListenAndServe()
	logger.Fatal(err)
}
