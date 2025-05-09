package api

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sakojpa/tasker/utils"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
	"time"
)

type CustomClaims struct {
	jwt.RegisteredClaims
	HashedPassword []byte
}

// MakeAuth authenticates user and issues JWT token if password is correct.
func MakeAuth(request *utils.AuthRequest) (*utils.AuthResp, error, int) {
	var authResp utils.AuthResp
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
	authResp.Token, err = jwtToken.SignedString(utils.SecretKey)
	if err != nil {
		fmt.Printf("Failed to sign JWT: %s\n", err.Error())
		return nil, fmt.Errorf("something went wrong"), http.StatusInternalServerError
	}
	return &authResp, nil, http.StatusOK
}

// TokenValidate verifies JWT token signature, checks expiration, and compares hashed passwords.
func TokenValidate(jwtToken string) (*jwt.Token, error) {
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
			return utils.SecretKey, nil
		},
	)
	if err != nil {
		return nil, err
	}
	return token, nil
}
