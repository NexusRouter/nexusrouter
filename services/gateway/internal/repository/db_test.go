package repository

import (
	"testing"
)

func Test_gatewayGormConfig_PrepareStmt(t *testing.T) {
	c := gatewayGormConfig()
	if c == nil {
		t.Fatal("nil config")
	}
	if !c.PrepareStmt {
		t.Fatal("PrepareStmt must be true for gateway DB")
	}
	if c.Logger == nil {
		t.Fatal("Logger expected")
	}
}

func Test_effectiveSQLiteBusyMS(t *testing.T) {
	if got := effectiveSQLiteBusyMS(0); got != 3000 {
		t.Fatalf("0 -> %d", got)
	}
	if got := effectiveSQLiteBusyMS(-1); got != 3000 {
		t.Fatalf("-1 -> %d", got)
	}
	if got := effectiveSQLiteBusyMS(5000); got != 5000 {
		t.Fatalf("5000 -> %d", got)
	}
	if got := effectiveSQLiteBusyMS(sqliteBusyTimeoutMaxMS + 1); got != sqliteBusyTimeoutMaxMS {
		t.Fatalf("over max -> %d", got)
	}
}

func Test_sqliteOpenDSN(t *testing.T) {
	if got := sqliteOpenDSN("gateway.db", 0); got != "gateway.db?_busy_timeout=3000" {
		t.Fatalf("got %q", got)
	}
	if got := sqliteOpenDSN("gateway.db", 8000); got != "gateway.db?_busy_timeout=8000" {
		t.Fatalf("got %q", got)
	}
	if got := sqliteOpenDSN("file::memory:?cache=shared", 0); got != "file::memory:?cache=shared&_busy_timeout=3000" {
		t.Fatalf("got %q", got)
	}
}
