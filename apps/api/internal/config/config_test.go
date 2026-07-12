package config

import "testing"

func TestDatabaseConfigDSN(t *testing.T) {
	config := DatabaseConfig{
		Host:     "db.example",
		Port:     5432,
		User:     "movie user",
		Password: "p@ss word",
		DBName:   "movie tracker",
		SSLMode:  "require",
	}

	got := config.DSN()
	want := "postgres://movie%20user:p%40ss%20word@db.example:5432/movie%20tracker?sslmode=require"
	if got != want {
		t.Fatalf("DSN() = %q, want %q", got, want)
	}
}
