package main

import (
    "context"      // New import
    "database/sql" // New import
    "flag"
    "fmt"
    "log/slog"
    "net/http"
    "os"
    "time"

    // Import the pq driver so that it can register itself with the database/sql 
    // package. Note that we alias this import to the blank identifier, to stop the Go 
    // compiler complaining that the package isn't being used.
    _ "github.com/lib/pq"
)

const version = "1.0.0"

// Add a db struct field to hold the configuration settings for our database connection
// pool. For now this only holds the DSN, which we will read in from a command-line flag.
type config struct {
    port int
    env  string
    db   struct {
        dsn          string
        maxOpenConns int
        maxIdleConns int
        maxIdleTime  time.Duration
    }
}

type application struct {
    config config
    logger *slog.Logger
}

func main() {
    var cfg config

    flag.IntVar(&cfg.port, "port", 4000, "API server port")
    flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

    // Read the DSN value from the db-dsn command-line flag into the config struct. We
    // default to using our development DSN if no flag is provided.
    flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL DSN")
    flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
    flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
    flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "PostgreSQL max connection idle time")



    flag.Parse()

    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

    // Call the openDB() helper function (see below) to create the connection pool,
    // passing in the config struct. If this returns an error, we log it and exit the
    // application immediately.
    db, err := openDB(cfg)
    if err != nil {
        logger.Error(err.Error())
        os.Exit(1)
    }

    // Defer a call to db.Close() so that the connection pool is closed before the
    // main() function exits.
    defer db.Close()

    // Also log a message to say that the connection pool has been successfully 
    // established.
    logger.Info("database connection pool established")

    app := &application{
        config: cfg,
        logger: logger,
    }

    srv := &http.Server{
        Addr:         fmt.Sprintf(":%d", cfg.port),
        Handler:      app.routes(),
        IdleTimeout:  time.Minute,
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 10 * time.Second,
        ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
    }

    logger.Info("starting server", "addr", srv.Addr, "env", cfg.env)
    
    // Because the err variable is now already declared in the code above, we need
    // to use the = operator here, instead of the := operator.
    err = srv.ListenAndServe()
    logger.Error(err.Error())
    os.Exit(1)
}


func openDB(cfg config) (*sql.DB, error) {
    db, err := sql.Open("postgres", cfg.db.dsn)
    if err != nil {
        return nil, err
    }

    // Set the maximum number of open (in-use + idle) connections in the pool. Note that
    // passing a value less than or equal to 0 will mean there is no limit.
    db.SetMaxOpenConns(cfg.db.maxOpenConns)

    // Set the maximum number of idle connections in the pool. Again, passing a value
    // less than or equal to 0 will mean there is no limit.
    db.SetMaxIdleConns(cfg.db.maxIdleConns)

    // Set the maximum idle timeout for connections in the pool. Passing a duration less
    // than or equal to 0 will mean that connections are not closed due to their idle time. 
    db.SetConnMaxIdleTime(cfg.db.maxIdleTime)

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    err = db.PingContext(ctx)
    if err != nil {
        db.Close()
        return nil, err
    }

    return db, nil
}
