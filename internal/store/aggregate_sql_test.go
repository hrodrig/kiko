package store

import (
	"testing"
	"time"
)

func TestAggregateSQLHelpers(t *testing.T) {
	hour := time.Date(2024, 6, 15, 14, 0, 0, 0, time.UTC)
	for _, driver := range []string{"sqlite", "postgres", "mysql"} {
		t.Run(driver+" hitCount", func(t *testing.T) {
			q, args := hitCountUpsertSQL(driver, "h", 1, hour, 3)
			if q == "" || len(args) != 4 {
				t.Fatalf("hitCountUpsertSQL(%s) = %q args=%v", driver, q, args)
			}
		})
		t.Run(driver+" hitUniqueInsert", func(t *testing.T) {
			q, args := hitUniqueInsertSQL(driver, "h", 1, hour, "vh")
			if q == "" || len(args) != 4 {
				t.Fatalf("hitUniqueInsertSQL(%s) = %q args=%v", driver, q, args)
			}
		})
		t.Run(driver+" hitUniqueBump", func(t *testing.T) {
			q, args := hitUniqueBumpSQL(driver, "h", 1, hour)
			if q == "" || len(args) != 3 {
				t.Fatalf("hitUniqueBumpSQL(%s) = %q args=%v", driver, q, args)
			}
		})
		t.Run(driver+" refCount", func(t *testing.T) {
			q, args := refCountUpsertSQL(driver, "h", 2, hour, 5)
			if q == "" || len(args) != 4 {
				t.Fatalf("refCountUpsertSQL(%s) = %q args=%v", driver, q, args)
			}
		})
		t.Run(driver+" hourValue", func(t *testing.T) {
			v := hourValue(driver, hour)
			if v == nil {
				t.Fatal("hourValue nil")
			}
		})
	}
}

func TestAggregateTableSQLAllDrivers(t *testing.T) {
	for _, driver := range []string{"sqlite", "postgres", "mysql"} {
		stmts := aggregateTableSQL(driver)
		if len(stmts) < 5 {
			t.Errorf("driver %s: got %d stmts", driver, len(stmts))
		}
	}
}
