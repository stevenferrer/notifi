package migration_test

import (
	"testing"

	"github.com/stevenferrer/notifi/postgres/migration"
	"github.com/stevenferrer/notifi/postgres/txdb"
)

func TestMigrate(t *testing.T) {
	db := txdb.MustOpen()
	defer db.Close()
	migration.MustMigrate(db)
}
