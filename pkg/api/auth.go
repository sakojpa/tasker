package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sakojpa/tasker/config"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
	"time"
)

type CustomClaims struct {
	jwt.RegisteredClaims
	HashedPassword []byte
}

type Auth struct{}

// Make authenticates user and issues JWT token if password is correct.
func (a Auth) make(request *AuthRequest) (*AuthResp, error, int) {
	var authResp AuthResp
	var err error
	password := os.Getenv("TODO_PASSWORD")
	if password != request.Password {
		return nil, fmt.Errorf("password is incorrect"), http.StatusUnauthorized
	}
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("unable to hash password"), http.StatusInternalServerError
	}
	claims := &CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			Issuer:    "tasker",
		},
		HashedPassword: hashedPwd,
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	authResp.Token, err = jwtToken.SignedString(secretKey)
	if err != nil {
		fmt.Printf("Failed to sign JWT: %s\n", err.Error())
		return nil, fmt.Errorf("something went wrong"), http.StatusInternalServerError
	}
	return &authResp, nil, http.StatusOK
}

// tokenValidate verifies JWT token signature, checks expiration, and compares hashed passwords.
func tokenValidate(jwtToken string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(
		jwtToken,
		&CustomClaims{},
		func(tk *jwt.Token) (interface{}, error) {
			if tk.Method.Alg() != "HS256" {
				return nil, fmt.Errorf("unknown signing method: %v", tk.Header["alg"])
			}
			claims := tk.Claims.(*CustomClaims)
			err := bcrypt.CompareHashAndPassword(claims.HashedPassword, []byte(os.Getenv("TODO_PASSWORD")))
			if err != nil && errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
				return nil, fmt.Errorf("password changes. token is expired")
			}
			if claims.ExpiresAt.Before(time.Now()) {
				return nil, fmt.Errorf("token is expired")
			}
			return secretKey, nil
		},
	)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func AuthConnect(
	handler func(w http.ResponseWriter, r *http.Request, ctx context.Context), c *config.Config,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pass := c.Auth.Password
		if len(pass) > 0 {
			var jwt string
			cookie, err := r.Cookie("token")
			if err == nil {
				jwt = cookie.Value
			}
			_, err = tokenValidate(jwt)
			if err != nil {
				sentErrorJson(w, err.Error(), http.StatusUnauthorized)
				return
			}
		}
		ctx, cancel := context.WithTimeout(context.Background(), c.DB.Timeout)
		defer cancel()
		handler(w, r, ctx)
	}
}
