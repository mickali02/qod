// Filename: cmd/api/main.go
package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
	// Import the internal/data package
	"github.com/mickali02/qod/internal/data"
)


type configuration struct {
	port int
	env  string
	db   struct {
		dsn string
	}
	cors struct {
		trustedOrigins []string
	}
	limiter struct {
        rps float64                      // requests per second
        burst int                        // initial requests possible
        enabled bool                     // enable or disable rate limiter
    }
}

type application struct {
	config       configuration
	logger       *slog.Logger
	db           *sql.DB
	commentModel data.CommentModel
}

func main() {
	// Initialize configuration
	cfg := loadConfig()
	// Initialize logger
	logger := setupLogger(cfg.env)

	// ---- DATABASE CODE ----
	// the call to openDB() sets up our connection pool
	db, err := openDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	// release the database resources before exiting
	defer db.Close()
	logger.Info("database connection pool established")

	// Initialize application with dependencies
	app := &application{
		config:       cfg,
		logger:       logger,
		db:           db,
		commentModel: data.CommentModel{DB: db},
	}
	// Run the application
	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

} // end of main

// loadConfig reads configuration from command line flags
func loadConfig() configuration {
	var cfg configuration

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment(development|staging|production)")

	// Add this line to read the database DSN
	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://comments:fishsticks@localhost/comments", "PostgreSQL DSN")

	// This creates a custom flag that can handle a space-separated list of origins.
	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

    flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate Limiter maximum requests per second")

    flag.IntVar(&cfg.limiter.burst, "limiter-burst", 5, "Rate Limiter maximum burst")

    flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")


	flag.Parse()

	return cfg
}

// setupLogger configures the application logger based on environment
func setupLogger(env string) *slog.Logger {
	var logger *slog.Logger

	logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

	return logger
}

func openDB(cfg configuration) (*sql.DB, error) {
	// open a connection pool
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// set a context to ensure DB operations don't take too long
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// let's test if the connection pool was created
	// we trying pinging it with a 5-second timeout
	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	// return the connection pool (sql.DB)
	return db, nil
}
