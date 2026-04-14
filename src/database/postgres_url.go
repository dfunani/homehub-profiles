package database

import (
	"fmt"
	"net/url"
)

// PostgresURLFromConfig builds a postgres URL suitable for GORM, Atlas, and other drivers.
// SSLMode defaults to "require" (AWS RDS). Set POSTGRES_SSLMODE=disable for local Postgres without TLS.
func PostgresURLFromConfig(cfg *DatabaseConfig) string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Path:   "/" + cfg.DBName,
	}
	ssl := cfg.SSLMode

	q := url.Values{}
	q.Set("sslmode", ssl)
	u.RawQuery = q.Encode()
	return u.String()
}
