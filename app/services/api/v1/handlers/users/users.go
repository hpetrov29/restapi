package users

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hpetrov29/restapi/business/web/v1/auth"
	"github.com/hpetrov29/restapi/internal/web"
)

// Handlers manages the set of user endpoints.
type Handlers struct {
	auth *auth.Auth
}

// New constructs a new handlers struct for route access.
func New(auth *auth.Auth) *Handlers {
	return &Handlers{
		auth: auth,
	}
}

func (h *Handlers) Query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	fmt.Println("acesssed")
	w.WriteHeader(200)
	return nil
}

func (h *Handlers) Token(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	kid := web.Param(r, "kid")
	if kid == "" {
		return errors.New("key id not provided")
	}

	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "userid",
			Issuer:    "service",
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
	}

	token, err := h.auth.GenerateToken(kid, claims)
	if err != nil {
		return fmt.Errorf("generatetoken: %w", err)
	}

	w.Write([]byte(token))
	return nil
}
