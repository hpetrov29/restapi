// Package db provides support for access to a mysql database using the mysql driver
package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/hpetrov29/restapi/internal/logger"
	"github.com/jmoiron/sqlx"
)

const (
	uniqueViolation = "23505"
	undefinedTable  = "1146"
)

// Set of error variables for CRUD operations.
var (
	ErrDBNotFound        = sql.ErrNoRows
	ErrDBDuplicatedEntry = errors.New("duplicated entry")
	ErrUndefinedTable    = errors.New("undefined table")
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

// ExecContext is a helper function to execute a CUD operation with
// logging and tracing.
func ExecContext(ctx context.Context, log *logger.Logger, db sqlx.ExtContext, query string) (sql.Result, error) {
	return NamedExecContext(ctx, log, db, query, struct{}{})
}

// NamedExecContext is a helper function to execute a CUD operation with
// logging and tracing where field replacement is necessary.
func NamedExecContext(ctx context.Context, log *logger.Logger, db sqlx.ExtContext, query string, data any) (sql.Result, error) {
	q := queryString(query, data)

	if _, ok := data.(struct{}); ok {
		log.Infoc(ctx, 5, "database.NamedExecContext", "query", q)
	} else {
		log.Infoc(ctx, 4, "database.NamedExecContext", "query", q)
	}

	res, err := sqlx.NamedExecContext(ctx, db, query, data); 
	if err != nil {
		return nil, err
	}

	return res, nil
}

// NamedQueryStruct is a helper function for executing queries that return a
// single value to be unmarshalled into a struct type where field replacement is necessary.
func NamedQueryStruct(ctx context.Context, log *logger.Logger, db sqlx.ExtContext, query string, data any, dest any) error {
	return namedQueryStruct(ctx, log, db, query, data, dest, false)
}

func namedQueryStruct(ctx context.Context, log *logger.Logger, db sqlx.ExtContext, query string, data any, dest any, withIn bool) error {
	q := queryString(query, data)

	log.Infoc(ctx, 5, "database.NamedQueryStruct", "query", q)

	var rows *sqlx.Rows
	var err error

	switch withIn {
	case true:
		rows, err = func() (*sqlx.Rows, error) {
			named, args, err := sqlx.Named(query, data)
			if err != nil {
				return nil, err
			}

			query, args, err := sqlx.In(named, args...)
			if err != nil {
				return nil, err
			}

			query = db.Rebind(query)
			return db.QueryxContext(ctx, query, args...)
		}()

	default:
		rows, err = sqlx.NamedQueryContext(ctx, db, query, data)
	}

	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		return ErrDBNotFound
	}

	if err := rows.StructScan(dest); err != nil {
		return err
	}

	return nil
}

// queryString provides a pretty print version of the query and parameters.
func queryString(query string, args any) string {
	query, params, err := sqlx.Named(query, args)
	if err != nil {
		return err.Error()
	}

	for _, param := range params {
		var value string
		switch v := param.(type) {
		case string:
			value = fmt.Sprintf("'%s'", v)
		case []byte:
			value = fmt.Sprintf("'%s'", string(v))
		default:
			value = fmt.Sprintf("%v", v)
		}
		query = strings.Replace(query, "?", value, 1)
	}

	query = strings.ReplaceAll(query, "\t", "")
	query = strings.ReplaceAll(query, "\n", " ")

	return strings.Trim(query, " ")
}