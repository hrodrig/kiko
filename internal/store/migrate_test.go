package store

import (
	"database/sql"
	"os"
	"testing"

	"github.com/hrodrig/kiko/internal/config"
	_ "modernc.org/sqlite"
)

func TestSchemaSQLAllDrivers(t *testing.T) {
	for _, driver := range []string{"sqlite", "postgres", "mysql"} {
		if stmts := schemaSQL(driver); len(stmts) == 0 {
			t.Fatalf("schemaSQL(%q) empty", driver)
		}
	}
}

func TestMigrateSQLiteIdempotent(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	for i := 0; i < 2; i++ {
		if err := migrate(db, "sqlite"); err != nil {
			t.Fatalf("migrate pass %d: %v", i, err)
		}
	}
}

func TestOpenSQLiteDefaultPath(t *testing.T) {
	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(wd)

	st, err := Open(config.DatabaseCfg{Driver: "sqlite"})
	if err != nil {
		t.Fatalf("Open() = %v", err)
	}
	st.Close()

	if _, err := os.Stat("./data/kiko.db"); err != nil {
		t.Fatalf("default db file missing: %v", err)
	}
}
