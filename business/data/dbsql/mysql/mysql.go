// Package db provides support for access to a mysql database using the mysql driver
package db

import (
	"context"
	"net/url"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// Config is the required properties to use the database.
type Config struct {
	User         string
	Password     string
	Host         string
	Name         string
	Schema       string
	MaxIdleConns int
	MaxOpenConns int
	DisableTLS   bool
}

// Open opens a database specified by a configuration struct 
// and a driver-specific data source name, usually consisting 
// of at least a database name and connection information.
func Open(config Config) (*sqlx.DB, error) {

	q := make(url.Values)
	q.Set("parseTime", "true")
	q.Set("timeout", "10s")
	q.Set("readTimeout", "10s")
	q.Set("writeTimeout", "10s")
	//q.Set("tls", "custom")
	//TO DO: Add TLS encryption
	
	u := url.URL{
		User:     url.UserPassword(config.User, config.Password),
		Host:     config.Host,
		Path:     config.Name,
		RawQuery: q.Encode(),
	}

	decoded, _ := url.QueryUnescape(u.String()[2:])
	db, err := sqlx.Open("mysql", decoded)
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetMaxOpenConns(config.MaxOpenConns)

	return db, nil
}

// StatusCheck returns nil if it can successfully talk to the database. It
// returns a non-nil error otherwise.
func StatusCheck(ctx context.Context, db *sqlx.DB) error {
	// Run a simple query to determine connectivity.
	// Running this query forces a round trip through the database.
	const q = `SELECT true`
	var tmp bool
	return db.QueryRowContext(ctx, q).Scan(&tmp)
}