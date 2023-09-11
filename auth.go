package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
)

const (
	RoleAdmin = "admin"
)

type User interface {
	IsAdmin() bool
	Username() string
}

type user struct {
	claims jwt.MapClaims
}

func GetUser(c echo.Context) User {
	user := c.Get("user")
	if user, ok := user.(User); ok {
		return user
	}
	return nil
}

func (u *user) getClaim(claimName string) string {
	if s, ok := u.claims[claimName]; ok {
		switch s := s.(type) {
		case string:
			return s
		case []string:
			if len(s) > 0 {
				return s[0]
			}
		case []interface{}:
			if len(s) > 0 {
				return fmt.Sprintf("%v", s[0])
			}
		}
	}
	return ""
}

func (u *user) IsAdmin() bool {
	return u.getClaim("blocks:role") == RoleAdmin
}

func (u *user) Username() string {
	return u.getClaim("name")
}

type anonymous struct{}

func (d *anonymous) IsAdmin() bool {
	return false
}

func (d *anonymous) Username() string {
	return "anonymous"
}

func AnonymousAccess() echo.MiddlewareFunc {
	user := &anonymous{}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("user", user)
			return next(c)
		}
	}
}

func Authorize(jwksUri string) echo.MiddlewareFunc {
	options := keyfunc.Options{
		Ctx: context.Background(),
		RefreshErrorHandler: func(err error) {
			log.Printf("There was an error with the jwt.Keyfunc\nError: %s", err.Error())
		},
		RefreshInterval:   5 * time.Minute,
		RefreshRateLimit:  10 * time.Second,
		RefreshTimeout:    10 * time.Second,
		RefreshUnknownKID: true,
	}

	jwks, err := keyfunc.Get(jwksUri, options)
	if err != nil {
		log.Fatalf("Failed to create JWKS from resource at the given URL.\nError: %s", err.Error())
	}

	const bearerPrefix = "bearer "

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			headerValue := c.Request().Header.Get("authorization")
			if headerValue == "" || !strings.HasPrefix(strings.ToLower(headerValue), bearerPrefix) {
				return echo.ErrUnauthorized
			}

			jwtB64 := headerValue[len(bearerPrefix):]

			claims := make(jwt.MapClaims)
			token, err := jwt.ParseWithClaims(jwtB64, claims, jwks.Keyfunc)
			if err != nil {
				return err
			}

			if !token.Valid {
				return echo.ErrUnauthorized
			}

			c.Set("user", &user{claims})
			return next(c)
		}
	}
}
