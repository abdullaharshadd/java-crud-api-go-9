package store

import (
	"fmt"
	"os"
	"strings"
)

// MySQLDSNFromEnv builds a go-sql-driver/mysql DSN from the environment. It
// honors a full DSN in DATABASE_URL / DB_URL when provided (stripping any
// jdbc:mysql:// or mysql:// scheme prefix), and otherwise assembles the DSN
// from the individual DB_HOST / DB_PORT / DB_USER / DB_PASSWORD / DB_NAME
// variables, falling back to the sandbox defaults shown in the environment
// context.
func MySQLDSNFromEnv() string {
	if raw := strings.TrimSpace(os.Getenv("DATABASE_URL")); raw != "" {
		return normalizeDSN(raw)
	}
	if raw := strings.TrimSpace(os.Getenv("DB_URL")); raw != "" {
		return normalizeDSN(raw)
	}

	host := envOr("DB_HOST", envOr("MYSQL_HOST", "127.0.0.1"))
	port := envOr("DB_PORT", envOr("MYSQL_PORT", "3306"))
	user := envOr("DB_USER", envOr("DB_USERNAME", envOr("MYSQL_USER", envOr("MYSQL_USERNAME", "app"))))
	pass := envOr("DB_PASSWORD", envOr("MYSQL_PASSWORD", "app"))
	name := envOr("DB_NAME", envOr("DB_DATABASE", envOr("MYSQL_DATABASE", "app")))
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, pass, host, port, name)
}

// normalizeDSN strips known scheme prefixes and query markers that other
// ecosystems (JDBC) add but the Go MySQL driver does not understand.
func normalizeDSN(raw string) string {
	for _, prefix := range []string{"jdbc:mysql://", "mysql://"} {
		if strings.HasPrefix(raw, prefix) {
			raw = strings.TrimPrefix(raw, prefix)
			break
		}
	}
	if !strings.Contains(raw, "?") {
		raw += "?parseTime=true"
	}
	return raw
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
