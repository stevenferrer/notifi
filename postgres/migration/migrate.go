package migration

import (
	"database/sql"

	// postgres driver
	_ "github.com/lib/pq"
	"github.com/lopezator/migrator"
)

// defaultOpts is the migration options
var defaultOpts = []migrator.Option{migrator.WithLogger(NopLogger())}

// Migrate migrates the database.
func Migrate(db *sql.DB, opts ...migrator.Option) error {
	if len(opts) == 0 {
		opts = defaultOpts
	}

	opts = append(opts, migrations)

	m, err := migrator.New(opts...)
	if err != nil {
		return err
	}

	return m.Migrate(db)
}

// MustMigrate migrates the database and panics if an error occurs.
func MustMigrate(db *sql.DB, opts ...migrator.Option) {
	err := Migrate(db, opts...)
	if err != nil {
		panic(err)
	}
}

type nopLogger struct{}

func NopLogger() migrator.Logger {
	return &nopLogger{}
}

func (l *nopLogger) Printf(string, ...interface{}) {}
