// Filename: cmd/api/main.go

package main

import (
	"flag"
	"log/slog"
	"os"
)

// The version number
const version = "1.0.0"

// Configuration settings go in here
// NOTE: For now we are only setting up the
//       setting for the server.  Later we
//       will add setting for the DB, rate limiter, etc.
//       The 'configuration' type is lowercase to
//       signal that it is private (non-exportable) to the 
//       main package
type configuration struct {
	port int
	env  string
}

// Set up Dependency Injection
// NOTE: In Go the variable name comes first
//       and the type comes second
//       The 'application' type is lowercase to
//       signal that it is private (non-exportable) to the 
//       main package
//       'logger' is a pointer because we want to share
//       one instance of it across the application (centralized logging)
type application struct {
	config configuration
	logger *slog.Logger
}

func main() { 
	// Initialize configuration
	cfg := loadConfig()
	// Initialize logger
	// Pass the environment from the config to the logger
	logger := setupLogger(cfg.env) 
	// Initialize application with dependencies
	app := &application{
		config: cfg,
		logger: logger,
	}

	// Run the application
	// Capture the error returned by app.serve()
	err := app.serve() 
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
        
	}

} // end of main

// loadConfig reads configuration from command line flags
func loadConfig() configuration {
	var cfg configuration

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	return cfg
}

// setupLogger configures the application logger based on environment
func setupLogger(env string) *slog.Logger {
	var logger *slog.Logger

	logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

	return logger
}