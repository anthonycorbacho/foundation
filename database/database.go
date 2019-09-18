package database

import (
	"context"
	"database/sql"
	"net/url"
	"time"

	"contrib.go.opencensus.io/integrations/ocsql"
	"github.com/jmoiron/sqlx"
	"go.opencensus.io/trace"

	// MySQL driver
	_ "github.com/go-sql-driver/mysql"
	// Postgres driver
	_ "github.com/lib/pq"
)

// Config is the required properties to use the database.
type Config struct {
	User       string
	Password   string
	Host       string
	Name       string
	DisableTLS bool
}

// Open knows how to open a database connection based on the configuration.
func Open(cfg Config) (*sqlx.DB, func(), error) {

	// Define SSL mode.
	sslMode := "require"
	if cfg.DisableTLS {
		sslMode = "disable"
	}

	// Query parameters.
	q := make(url.Values)
	q.Set("sslmode", sslMode)
	q.Set("timezone", "utc")

	// Construct url.
	u := url.URL{
		Scheme:   cfg.Name,
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     cfg.Host,
		Path:     cfg.Name,
		RawQuery: q.Encode(),
	}

	ocsql.RegisterAllViews()
	// Register our ocsql wrapper for the provided database driver.
	driverName, err := ocsql.Register(cfg.Name, ocsql.WithAllTraceOptions())
	if err != nil {
		return nil, nil, err
	}

	// Connect to a mysql database using the ocsql driver wrapper.
	db, err := sql.Open(driverName, u.String())
	if err != nil {
		return nil, nil, err
	}
	// Record DB stats every 5 seconds until we exit.
	stop := ocsql.RecordStats(db, 5*time.Second)
	dbx := sqlx.NewDb(db, cfg.Name)

	close := func() {
		stop()
		_ = dbx.Close() // explicitly ignore error.
	}

	return dbx, close, nil
}

// StatusCheck returns nil if it can successfully talk to the database. It
// returns a non-nil error otherwise.
func StatusCheck(ctx context.Context, db *sqlx.DB) error {
	ctx, span := trace.StartSpan(ctx, "database.StatusCheck")
	defer span.End()

	// Run a simple query to determine connectivity. The db has a "Ping" method
	// but it can false-positive when it was previously able to talk to the
	// database but the database has since gone away. Running this query forces a
	// round trip to the database.
	const q = `SELECT true`
	var tmp bool
	return db.QueryRowContext(ctx, q).Scan(&tmp)
}
