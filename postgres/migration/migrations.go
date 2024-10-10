package migration

import (
	"database/sql"

	"github.com/lopezator/migrator"
)

var migrations = migrator.Migrations(
	&migrator.Migration{
		Name: "Create tokens table",
		Func: func(tx *sql.Tx) error {
			stmnt := `CREATE TABLE IF NOT EXISTS "tokens" (
				id varchar PRIMARY KEY,
				cb_key varchar NOT NULL,
				created_at timestamp NOT NULL DEFAULT NOW()
			)`
			if _, err := tx.Exec(stmnt); err != nil {
				return err
			}

			return nil
		},
	},
	&migrator.Migration{
		Name: "Create callbacks table",
		Func: func(tx *sql.Tx) error {
			stmnt := `CREATE TABLE IF NOT EXISTS "callbacks" (
				id varchar PRIMARY KEY,
				token_id varchar NOT NULL REFERENCES tokens (id),
				cb_type varchar NOT NULL,
				cb_url varchar NOT NULL,
				updated_at timestamp,
				created_at timestamp NOT NULL DEFAULT NOW()
			)`
			if _, err := tx.Exec(stmnt); err != nil {
				return err
			}

			return nil
		},
	},
	&migrator.Migration{
		Name: "Create notifications table",
		Func: func(tx *sql.Tx) error {
			stmnt := `CREATE TABLE IF NOT EXISTS "notifications" (
				id varchar PRIMARY KEY,
				src_token_id varchar NOT NULL REFERENCES tokens (id),
				dest_token_id varchar NOT NULL REFERENCES tokens (id),
				cb_type varchar NOT NULL,
				status varchar NOT NULL,
				payload jsonb NOT NULL,
				updated_at timestamp,
				created_at timestamp NOT NULL DEFAULT NOW()
			)`
			if _, err := tx.Exec(stmnt); err != nil {
				return err
			}

			return nil
		},
	},
	&migrator.Migration{
		Name: "Create idemp_keys table",
		Func: func(tx *sql.Tx) error {
			stmnt := `CREATE TABLE IF NOT EXISTS "idemp_keys" (
				key varchar PRIMARY KEY,
				created_at timestamp NOT NULL DEFAULT NOW()
			)`
			if _, err := tx.Exec(stmnt); err != nil {
				return err
			}

			return nil
		},
	},
)
