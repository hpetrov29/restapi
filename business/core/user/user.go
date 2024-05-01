// Package user provides an example of a core business API. Right now these
// calls are just wrapping the data/data layer.
package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/mail"
	"time"

	"github.com/google/uuid"
	"github.com/hpetrov29/restapi/internal/logger"
	"golang.org/x/crypto/bcrypt"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound              = errors.New("user not found")
	ErrUniqueEmail           = errors.New("email is not unique")
	ErrAuthenticationFailure = errors.New("authentication failed")
)

// =============================================================================

// Storer interface declares the core behavior and is required to write and
// retrieve data.
type Storer interface {
	Create(ctx context.Context, user User) (sql.Result, error)
	Delete(ctx context.Context, user User) error
	QueryByEmail(ctx context.Context, email mail.Address) (User, error)
}

// Core manages the set of APIs for user api access
type Core struct {
	storer Storer
	log *logger.Logger
}

// NewCore constructs a core for user api access.
func NewCore(st Storer, log *logger.Logger) *Core {
	return &Core{
		storer: st, 
		log: log,
	}
}

// Create adds a new user to the system.
func (c *Core) Create(ctx context.Context, newUser NewUser) (User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, fmt.Errorf("generatefrompassword: %w", err)
	}

	now := time.Now()

	usr := User{
		ID:           uuid.New(),
		Name:         newUser.Name,
		Email:        newUser.Email,
		PasswordHash: hash,
		Roles:        newUser.Roles,
		Department:   newUser.Department,
		Enabled:      true,
		DateCreated:  now,
		DateUpdated:  now,
	}

	if _, err := c.storer.Create(ctx, usr); err != nil {
		return User{}, fmt.Errorf("create: %w", err)
	}

	return usr, nil
}

// Delete removes a specified user.
func (c *Core) Delete(ctx context.Context, usr User) error {
	if err := c.storer.Delete(ctx, usr); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// QueryByEmail finds the user by a specified user email.
func (c *Core) QueryByEmail(ctx context.Context, email mail.Address) (User, error) {
	user, err := c.storer.QueryByEmail(ctx, email)
	if err != nil {
		return User{}, fmt.Errorf("query: email[%s]: %w", email, err)
	}

	return user, nil
}


// =============================================================================

// Authenticate finds a user by their email and verifies their password. On
// success it returns a Claims User representing this user. The claims can be
// used to generate a token for future authentication.
func (c *Core) Authenticate(ctx context.Context, email mail.Address, password string) (User, error) {
	usr, err := c.QueryByEmail(ctx, email)
	if err != nil {
		return User{}, fmt.Errorf("query: email[%s]: %w", email, err)
	}

	if err := bcrypt.CompareHashAndPassword(usr.PasswordHash, []byte(password)); err != nil {
		return User{}, fmt.Errorf("comparehashandpassword: %w", ErrAuthenticationFailure)
	}

	return usr, nil
}