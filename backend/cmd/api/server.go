package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Declare a HTTP server with timeout settings, which listens on the port
// provided in the config struct and uses servemux as the handler
// Create a new Go log.Logger instance, The "" and 0 indicate that the
// log.Logger instance should not use a prefix or any flags
// Any log messages that http.Server writes will be passed to our Logger.Write() method
// because our Logger type satisfies the io.Writer interface (due to Write() method)
func (app *application) server() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		app.logger.PrintInfo("shutting down server", map[string]string{
			"signal": s.String(),
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Call Shutdown() on our server, passing in the context, method returns nil if
		// the graceful shutdown was successful or because the shutdown didn't complete
		// before the 5-second context deadline is hit. This will cause ListenAndServe()
		// to immediately return a http.ErrServerClosed error
		shutdownError <- srv.Shutdown(ctx)
	}()

	app.logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  app.config.env,
	})

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return nil
	}

	app.logger.PrintInfo("stopped server", map[string]string{
		"addr": srv.Addr,
	})

	return nil
}
