package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/lyttonliao/StratCheck/internal/data"

	// Alias this import to blank identifier to stop Go compiler from erroring
	_ "github.com/lib/pq"
)

// Declare a string containing the app version number as hard-coded global constant
// Will automatically generate this at build time later
const version = "1.0.0"

// Define config struct to hold all configuration settings
// For now, config settings will only have the network port that we want the server to listen on
// and the name of the operation environment for the application (dev, staging, prod)
type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
}

// Define application struct to hold dependencies for our HTTP handlers, helpers, middleware
type application struct {
	config config
	logger *log.Logger
	models data.Models
}

func main() {
	var cfg config

	// Read value of the port and env command-lines into config struct
	// Default values: 4000, development
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("STRATCHECK_DB_DSN"), "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")
	flag.Parse()

	// Initialize a new logger which writes message to standard out stream
	// prefixed with current date and time
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// Call the openDB() helper function to create the connection pool, passing in the config struct
	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}

	// Defer a call to db.Close() so that the connection pool is closed before exiting main() function
	defer db.Close()
	logger.Printf("database connection established")

	// Declare an instance of the application struct, containing the config struct and logger
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	// Declare a HTTP server with timeout settings, which listens on the port
	// provided in the config struct and uses servemux as the handler
	srv := http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Start the HTTP server
	logger.Printf("starting %s server on %s", cfg.env, srv.Addr)
	err = srv.ListenAndServe()
	logger.Fatal(err)
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// Values less than or equal to 0 will mean no limit
	// Set max number of open (in use + idle) connections in pool
	db.SetMaxIdleConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	// time.ParseDuration() function convert hte idle timeout duration string to a time.Duration type
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	// Create a context with a 5-second timeout deadline
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use PingContext() to establish a new connection to the database, passing in the
	// context we created above as a param
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
