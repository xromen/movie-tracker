package config

import "testing"

func TestTMDBRateLimitConfig(t *testing.T) {
	t.Setenv("TMDB_RPM", "")
	t.Setenv("TMDB_BURST", "")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.TMDB.RPM != 240 || cfg.TMDB.Burst != 40 {
		t.Fatalf("unexpected TMDB defaults: RPM = %d, Burst = %d", cfg.TMDB.RPM, cfg.TMDB.Burst)
	}
	for _, key := range []string{"TMDB_RPM", "TMDB_BURST"} {
		for _, value := range []string{"0", "-1", "invalid"} {
			t.Run(key+"="+value, func(t *testing.T) {
				t.Setenv(key, value)
				if _, err := Load(); err == nil {
					t.Fatal("invalid rate limiter setting was accepted")
				}
			})
		}
	}
}

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
