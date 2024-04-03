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

	"github.com/lyttonliao/StratCheck/internal/jsonlog"

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
	logger *jsonlog.Logger
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

	// Initialize a new jsonlog.Logger which write snay messages *at or above* the INFO severity level to the stdout stream
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	// Call the openDB() helper function to create the connection pool, passing in the config struct
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	// Defer a call to db.Close() so that the connection pool is closed before exiting main() function
	defer db.Close()

	logger.PrintInfo("database connection pool established", nil)

	// Declare an instance of the application struct, containing the config struct and logger
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	// Declare a HTTP server with timeout settings, which listens on the port
	// provided in the config struct and uses servemux as the handler
	// Create a new Go log.Logger instance, The "" and 0 indicate that the
	// log.Logger instance should not use a prefix or any flags
	// Any log messages that http.Server writes will be passed to our Logger.Write() method
	// because our Logger type satisfies the io.Writer interface (due to Write() method)
	srv := http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		ErrorLog:     log.New(logger, "", 0),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.PrintInfo("starting server", map[string]string{
		"addr": cfg.env,
		"env":  srv.Addr,
	})
	err = srv.ListenAndServe()
	logger.PrintFatal(err, nil)
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
